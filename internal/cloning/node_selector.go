package cloning

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"time"

	"github.com/moltbunker/moltbunker/internal/p2p"
	"github.com/moltbunker/moltbunker/pkg/types"
)

// DHTNodeSelector implements NodeSelector using the P2P DHT
type DHTNodeSelector struct {
	dht *p2p.DHT
}

// NewDHTNodeSelector creates a new DHT-based node selector
func NewDHTNodeSelector(dht *p2p.DHT) *DHTNodeSelector {
	return &DHTNodeSelector{
		dht: dht,
	}
}

// FindNodes finds nodes matching criteria
func (s *DHTNodeSelector) FindNodes(ctx context.Context, region string, excludeNodes []types.NodeID) ([]*types.Node, error) {
	if s.dht == nil {
		return nil, fmt.Errorf("DHT not initialized")
	}

	// Get all connected peers
	peers := s.dht.GetConnectedPeers()
	if len(peers) == 0 {
		return nil, fmt.Errorf("no connected peers available")
	}

	// Build exclusion set
	excluded := make(map[types.NodeID]bool)
	for _, id := range excludeNodes {
		excluded[id] = true
	}

	// Also exclude local node
	localNode := s.dht.LocalNode()
	if localNode != nil {
		excluded[localNode.ID] = true
	}

	// Filter and score nodes
	candidates := make([]*nodeCandidate, 0, len(peers))
	for _, node := range peers {
		if excluded[node.ID] {
			continue
		}

		// Check if node has container runtime capability
		if !node.Capabilities.ContainerRuntime {
			continue
		}

		// Calculate score based on region match and freshness
		score := s.scoreNode(node, region)
		candidates = append(candidates, &nodeCandidate{
			node:  node,
			score: score,
		})
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("no suitable nodes found")
	}

	// Sort by score (highest first)
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].score > candidates[j].score
	})

	// Return top candidates (up to 5)
	maxNodes := 5
	if len(candidates) < maxNodes {
		maxNodes = len(candidates)
	}

	result := make([]*types.Node, maxNodes)
	for i := 0; i < maxNodes; i++ {
		result[i] = candidates[i].node
	}

	return result, nil
}

// GetNodeByID returns a specific node
func (s *DHTNodeSelector) GetNodeByID(ctx context.Context, nodeID types.NodeID) (*types.Node, error) {
	if s.dht == nil {
		return nil, fmt.Errorf("DHT not initialized")
	}

	// Check local peer cache first
	node, exists := s.dht.GetPeer(nodeID)
	if exists {
		return node, nil
	}

	// Try to find via DHT
	nodes, err := s.dht.FindNode(ctx, nodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to find node: %w", err)
	}

	if len(nodes) == 0 {
		return nil, fmt.Errorf("node not found: %s", nodeID.String())
	}

	// Return the first match (should be exact match)
	for _, n := range nodes {
		if n.ID == nodeID {
			return n, nil
		}
	}

	return nodes[0], nil
}

// nodeCandidate holds a node with its selection score
type nodeCandidate struct {
	node  *types.Node
	score float64
}

// scoreNode calculates a score for node selection
func (s *DHTNodeSelector) scoreNode(node *types.Node, targetRegion string) float64 {
	score := 1.0

	// Region matching (highest priority)
	if targetRegion != "" && targetRegion != "random" {
		if node.Region == targetRegion {
			score += 10.0 // Strong preference for matching region
		} else if s.isSameContinent(node.Region, targetRegion) {
			score += 5.0 // Moderate preference for same continent
		}
	}

	// Freshness bonus (prefer recently seen nodes)
	since := time.Since(node.LastSeen)
	if since < 1*time.Minute {
		score += 3.0
	} else if since < 5*time.Minute {
		score += 2.0
	} else if since < 15*time.Minute {
		score += 1.0
	}

	// Capability bonus
	if node.Capabilities.TorSupport {
		score += 1.0
	}

	// Add some randomness to prevent always selecting the same nodes
	score += rand.Float64() * 0.5

	return score
}

// isSameContinent checks if two regions are on the same continent
func (s *DHTNodeSelector) isSameContinent(region1, region2 string) bool {
	continentMap := map[string]string{
		"americas":          "americas",
		"us-east":           "americas",
		"us-west":           "americas",
		"sa-east":           "americas",
		"europe":            "europe",
		"eu-west":           "europe",
		"eu-central":        "europe",
		"asia_pacific":      "asia",
		"ap-southeast":      "asia",
		"ap-northeast":      "asia",
		"ap-south":          "asia",
		"middle_east":       "middle_east",
		"africa":            "africa",
	}

	c1, ok1 := continentMap[region1]
	c2, ok2 := continentMap[region2]

	if !ok1 || !ok2 {
		return false
	}

	return c1 == c2
}

// RandomNodeSelector is a simple random node selector for testing
type RandomNodeSelector struct {
	nodes []*types.Node
}

// NewRandomNodeSelector creates a node selector from a static list
func NewRandomNodeSelector(nodes []*types.Node) *RandomNodeSelector {
	return &RandomNodeSelector{
		nodes: nodes,
	}
}

// FindNodes returns random nodes from the list
func (s *RandomNodeSelector) FindNodes(ctx context.Context, region string, excludeNodes []types.NodeID) ([]*types.Node, error) {
	if len(s.nodes) == 0 {
		return nil, fmt.Errorf("no nodes available")
	}

	// Build exclusion set
	excluded := make(map[types.NodeID]bool)
	for _, id := range excludeNodes {
		excluded[id] = true
	}

	// Filter nodes
	candidates := make([]*types.Node, 0, len(s.nodes))
	for _, node := range s.nodes {
		if excluded[node.ID] {
			continue
		}
		if region != "" && region != "random" && node.Region != region {
			continue
		}
		candidates = append(candidates, node)
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("no suitable nodes found")
	}

	// Shuffle and return up to 5
	rand.Shuffle(len(candidates), func(i, j int) {
		candidates[i], candidates[j] = candidates[j], candidates[i]
	})

	maxNodes := 5
	if len(candidates) < maxNodes {
		maxNodes = len(candidates)
	}

	return candidates[:maxNodes], nil
}

// GetNodeByID returns a specific node
func (s *RandomNodeSelector) GetNodeByID(ctx context.Context, nodeID types.NodeID) (*types.Node, error) {
	for _, node := range s.nodes {
		if node.ID == nodeID {
			return node, nil
		}
	}
	return nil, fmt.Errorf("node not found: %s", nodeID.String())
}
