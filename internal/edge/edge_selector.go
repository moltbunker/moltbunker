package edge

import (
	"context"
	"fmt"
	"sync"

	"github.com/moltbunker/moltbunker/pkg/types"
)

// EdgeSelector picks an edge node for a new tunnel and is the foundation for
// multi-edge failover: a provider that loses its edge can call SelectEdge again
// to obtain the next-best healthy candidate. Selection policy for this PR is
// least-connections with round-robin among ties, which spreads load across a
// small operator-curated edge set. Session-affinity (consistent hashing) is a
// documented follow-up.
//
// EdgeSelector owns no goroutines; it is backed by an EdgeRegistry and tracks
// an in-flight connection count per node. Safe for concurrent use.
type EdgeSelector struct {
	registry *EdgeRegistry

	mu       sync.Mutex
	conns    map[types.NodeID]int // in-flight tunnel count per node
	rrCursor int                  // round-robin cursor for tie-breaking
}

// NewEdgeSelector creates a selector backed by registry.
func NewEdgeSelector(registry *EdgeRegistry) *EdgeSelector {
	return &EdgeSelector{
		registry: registry,
		conns:    make(map[types.NodeID]int),
	}
}

// AddEdge registers an edge node (delegates to the backing registry).
func (s *EdgeSelector) AddEdge(info EdgeNodeInfo) {
	s.registry.Register(info)
}

// RemoveEdge unregisters an edge node and drops its connection count.
func (s *EdgeSelector) RemoveEdge(nodeID types.NodeID) {
	s.registry.Unregister(nodeID)
	s.mu.Lock()
	delete(s.conns, nodeID)
	s.mu.Unlock()
}

// MarkHealthy flips a node's health in the backing registry.
func (s *EdgeSelector) MarkHealthy(nodeID types.NodeID, healthy bool) {
	s.registry.UpdateHealth(nodeID, healthy)
}

// SelectEdge returns the best healthy edge node: the one with the fewest
// in-flight tunnels, breaking ties round-robin. Returns an error when no healthy
// edge is available (the caller's failover signal). The chosen node's
// connection count is incremented; callers MUST call Release(nodeID) when the
// tunnel closes.
func (s *EdgeSelector) SelectEdge(_ context.Context) (EdgeNodeInfo, error) {
	healthy := s.registry.ListHealthy()
	if len(healthy) == 0 {
		return EdgeNodeInfo{}, fmt.Errorf("no healthy edge nodes available")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Find the minimum in-flight count.
	minConns := -1
	for _, n := range healthy {
		c := s.conns[n.NodeID]
		if minConns < 0 || c < minConns {
			minConns = c
		}
	}

	// Collect the tied candidates (deterministic membership; order within the
	// slice follows the registry snapshot, which the round-robin cursor walks).
	candidates := make([]EdgeNodeInfo, 0, len(healthy))
	for _, n := range healthy {
		if s.conns[n.NodeID] == minConns {
			candidates = append(candidates, n)
		}
	}

	chosen := candidates[s.rrCursor%len(candidates)]
	s.rrCursor++
	s.conns[chosen.NodeID]++
	return chosen, nil
}

// Release decrements the in-flight tunnel count for a node (call when a tunnel
// established via SelectEdge closes). It never goes below zero.
func (s *EdgeSelector) Release(nodeID types.NodeID) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.conns[nodeID] > 0 {
		s.conns[nodeID]--
	}
}

// InFlight returns the current in-flight tunnel count for a node (test/observability).
func (s *EdgeSelector) InFlight(nodeID types.NodeID) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.conns[nodeID]
}
