package edge

import (
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/moltbunker/moltbunker/pkg/types"
)

// EdgeNodeInfo describes a known edge provider node: where it accepts tunnels
// and its current liveness. Edge nodes are a small, operator-curated, stake-
// gated set (not the DHT/gossip container-provider population), so an in-memory
// registry is the correct fit. Persistence across daemon restart is a follow-up.
type EdgeNodeInfo struct {
	NodeID         types.NodeID
	WalletAddr     string // hex wallet address (public), informational
	IngressAddr    string // public host the edge node terminates TLS on
	TunnelPort     int    // port providers dial to open a reverse tunnel
	TLSFingerprint string // hex SHA256 of the edge node's TLS SPKI (optional)
	Tier           string // staking tier label (informational)
	LastSeen       time.Time
	Healthy        bool
}

// FullAddr returns the dial target "host:port" for the edge node's tunnel
// endpoint. It joins IngressAddr with TunnelPort using net.JoinHostPort so IPv6
// literals are bracketed correctly.
func (e EdgeNodeInfo) FullAddr() string {
	return net.JoinHostPort(e.IngressAddr, strconv.Itoa(e.TunnelPort))
}

// EdgeRegistry is a thread-safe in-memory map of known edge nodes. It is
// populated by the ingress side when an edge provider announces its
// EdgeCapabilities during reverse-tunnel registration, and read by the
// EdgeSelector when choosing an edge for a new tunnel.
type EdgeRegistry struct {
	mu    sync.RWMutex
	nodes map[types.NodeID]*EdgeNodeInfo
}

// NewEdgeRegistry creates an empty registry.
func NewEdgeRegistry() *EdgeRegistry {
	return &EdgeRegistry{nodes: make(map[types.NodeID]*EdgeNodeInfo)}
}

// Register adds or replaces an edge node. LastSeen is stamped to now and a node
// is considered Healthy on (re-)registration; the EdgeProbe flips it unhealthy
// if it stops responding. A copy is stored so callers cannot mutate the record.
func (r *EdgeRegistry) Register(info EdgeNodeInfo) {
	r.mu.Lock()
	defer r.mu.Unlock()
	info.LastSeen = time.Now()
	info.Healthy = true
	cp := info
	r.nodes[info.NodeID] = &cp
}

// RegisterEdge is the convenience seam consumed by the tunnel server's
// EdgeRegistrar interface: it builds an EdgeNodeInfo from the announced
// capabilities and registers it. Keeping the signature primitive-only lets
// internal/tunnel declare the seam without importing this package. The
// maxStreams hint is currently advisory (recorded via TLSFingerprint-adjacent
// metadata is out of scope here); selection uses live connection counts.
func (r *EdgeRegistry) RegisterEdge(nodeID types.NodeID, walletAddr, ingressAddr string, tunnelPort, _ int) {
	r.Register(EdgeNodeInfo{
		NodeID:      nodeID,
		WalletAddr:  walletAddr,
		IngressAddr: ingressAddr,
		TunnelPort:  tunnelPort,
	})
}

// Unregister removes an edge node.
func (r *EdgeRegistry) Unregister(nodeID types.NodeID) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.nodes, nodeID)
}

// ListHealthy returns a snapshot of the currently-healthy edge nodes.
func (r *EdgeRegistry) ListHealthy() []EdgeNodeInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]EdgeNodeInfo, 0, len(r.nodes))
	for _, n := range r.nodes {
		if n.Healthy {
			out = append(out, *n)
		}
	}
	return out
}

// ListAll returns a snapshot of every registered edge node regardless of health.
func (r *EdgeRegistry) ListAll() []EdgeNodeInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]EdgeNodeInfo, 0, len(r.nodes))
	for _, n := range r.nodes {
		out = append(out, *n)
	}
	return out
}

// ByNodeID returns the edge node info for nodeID, if registered.
func (r *EdgeRegistry) ByNodeID(nodeID types.NodeID) (EdgeNodeInfo, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	n, ok := r.nodes[nodeID]
	if !ok {
		return EdgeNodeInfo{}, false
	}
	return *n, true
}

// UpdateHealth sets the health flag for a node and refreshes LastSeen when
// marking healthy. A missing node is a no-op.
func (r *EdgeRegistry) UpdateHealth(nodeID types.NodeID, healthy bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	n, ok := r.nodes[nodeID]
	if !ok {
		return
	}
	n.Healthy = healthy
	if healthy {
		n.LastSeen = time.Now()
	}
}

// Len returns the number of registered edge nodes.
func (r *EdgeRegistry) Len() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.nodes)
}
