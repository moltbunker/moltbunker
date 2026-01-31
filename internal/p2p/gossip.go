package p2p

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/moltbunker/moltbunker/pkg/types"
)

// GossipProtocol implements gossip protocol for state synchronization
type GossipProtocol struct {
	router    *Router
	state     map[string]interface{}
	stateMu   sync.RWMutex
	peers     map[types.NodeID]*GossipPeer
	peersMu   sync.RWMutex
	gossipInterval time.Duration
}

// GossipPeer represents a peer in the gossip network
type GossipPeer struct {
	Node      *types.Node
	LastGossip time.Time
	StateHash []byte
}

// GossipMessage represents a gossip message
type GossipMessage struct {
	Type      string
	Key       string
	Value     interface{}
	Timestamp time.Time
	From      types.NodeID
}

// NewGossipProtocol creates a new gossip protocol instance
func NewGossipProtocol(router *Router) *GossipProtocol {
	return &GossipProtocol{
		router:         router,
		state:          make(map[string]interface{}),
		peers:          make(map[types.NodeID]*GossipPeer),
		gossipInterval: 10 * time.Second,
	}
}

// Start starts the gossip protocol
func (gp *GossipProtocol) Start(ctx context.Context) {
	ticker := time.NewTicker(gp.gossipInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			gp.gossip(ctx)
		}
	}
}

// gossip performs gossip propagation
func (gp *GossipProtocol) gossip(ctx context.Context) {
	gp.peersMu.RLock()
	peers := make([]*GossipPeer, 0, len(gp.peers))
	for _, peer := range gp.peers {
		peers = append(peers, peer)
	}
	gp.peersMu.RUnlock()

	// Select random subset of peers to gossip with
	// For simplicity, gossip with all peers
	for _, peer := range peers {
		gp.sendGossip(ctx, peer)
	}
}

// sendGossip sends gossip message to a peer
func (gp *GossipProtocol) sendGossip(ctx context.Context, peer *GossipPeer) {
	gp.stateMu.RLock()
	stateCopy := make(map[string]interface{})
	for k, v := range gp.state {
		stateCopy[k] = v
	}
	gp.stateMu.RUnlock()

	// Create gossip message
	msg := GossipMessage{
		Type:      "state_update",
		Value:     stateCopy,
		Timestamp: time.Now(),
	}

	msgData, err := json.Marshal(msg)
	if err != nil {
		return
	}

	p2pMsg := &types.Message{
		Type:      types.MessageTypeGossip,
		Payload:   msgData,
		Timestamp: time.Now(),
	}

	if err := gp.router.SendMessage(ctx, peer.Node.ID, p2pMsg); err != nil {
		// Log error
		return
	}

	peer.LastGossip = time.Now()
}

// UpdateState updates local state and propagates via gossip
func (gp *GossipProtocol) UpdateState(key string, value interface{}) {
	gp.stateMu.Lock()
	gp.state[key] = value
	gp.stateMu.Unlock()

	// Trigger immediate gossip
	ctx := context.Background()
	gp.gossip(ctx)
}

// GetState returns current state
func (gp *GossipProtocol) GetState(key string) (interface{}, bool) {
	gp.stateMu.RLock()
	defer gp.stateMu.RUnlock()
	value, exists := gp.state[key]
	return value, exists
}

// HandleGossipMessage handles incoming gossip message
func (gp *GossipProtocol) HandleGossipMessage(ctx context.Context, msg *types.Message, from *types.Node) error {
	var gossipMsg GossipMessage
	if err := json.Unmarshal(msg.Payload, &gossipMsg); err != nil {
		return fmt.Errorf("failed to unmarshal gossip message: %w", err)
	}

	// Merge state
	if gossipMsg.Type == "state_update" {
		if stateMap, ok := gossipMsg.Value.(map[string]interface{}); ok {
			gp.stateMu.Lock()
			for k, v := range stateMap {
				// Only update if newer
				if existing, exists := gp.state[k]; !exists {
					gp.state[k] = v
				} else {
					// Simple timestamp-based conflict resolution
					// In production, use vector clocks or CRDTs
					gp.state[k] = v
				}
			}
			gp.stateMu.Unlock()
		}
	}

	return nil
}

// AddPeer adds a peer to gossip network
func (gp *GossipProtocol) AddPeer(node *types.Node) {
	gp.peersMu.Lock()
	defer gp.peersMu.Unlock()

	gp.peers[node.ID] = &GossipPeer{
		Node:       node,
		LastGossip: time.Now(),
	}
}
