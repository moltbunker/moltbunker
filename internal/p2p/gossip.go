package p2p

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/moltbunker/moltbunker/pkg/types"
)

const (
	// maxStateEntries is the maximum number of entries in the gossip state map
	maxStateEntries = 10000
	// maxPeers is the maximum number of peers in the gossip network
	maxPeers = 1000
	// maxStateValueSize limits the serialized size of a single gossip state value
	// to prevent memory abuse from oversized payloads (64 KB per value)
	maxStateValueSize = 64 * 1024
)

// GossipProtocol implements gossip protocol for state synchronization
type GossipProtocol struct {
	router         *Router
	stakeVerifier  *StakeVerifier
	state          map[string]*stateEntry
	stateMu        sync.RWMutex
	peers          map[types.NodeID]*GossipPeer
	peersMu        sync.RWMutex
	gossipInterval time.Duration
}

// stateEntry wraps a state value with a timestamp for staleness tracking
type stateEntry struct {
	Value       interface{}
	LastUpdated time.Time
}

// GossipPeer represents a peer in the gossip network
type GossipPeer struct {
	Node      *types.Node
	LastGossip time.Time
	LastSeen   time.Time
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
		state:          make(map[string]*stateEntry),
		peers:          make(map[types.NodeID]*GossipPeer),
		gossipInterval: 10 * time.Second,
	}
}

// Start starts the gossip protocol
func (gp *GossipProtocol) Start(ctx context.Context) {
	ticker := time.NewTicker(gp.gossipInterval)
	defer ticker.Stop()

	// Clean stale gossip state every 10 minutes
	cleanTicker := time.NewTicker(10 * time.Minute)
	defer cleanTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			gp.gossip(ctx)
		case <-cleanTicker.C:
			gp.CleanStaleState(30 * time.Minute)
		}
	}
}

// SetStakeVerifier sets the stake verifier for gossip gating.
func (gp *GossipProtocol) SetStakeVerifier(sv *StakeVerifier) {
	gp.peersMu.Lock()
	gp.stakeVerifier = sv
	gp.peersMu.Unlock()
}

// gossip performs gossip propagation, skipping unstaked peers.
func (gp *GossipProtocol) gossip(ctx context.Context) {
	gp.peersMu.RLock()
	sv := gp.stakeVerifier
	peers := make([]*GossipPeer, 0, len(gp.peers))
	for _, peer := range gp.peers {
		peers = append(peers, peer)
	}
	gp.peersMu.RUnlock()

	// Select random subset of peers to gossip with
	// For simplicity, gossip with all peers (that are staked)
	for _, peer := range peers {
		// Skip unstaked peers if stake verifier is set
		if sv != nil && !sv.IsStaked(ctx, peer.Node.ID) {
			continue
		}
		gp.sendGossip(ctx, peer)
	}
}

// sendGossip sends gossip message to a peer
func (gp *GossipProtocol) sendGossip(ctx context.Context, peer *GossipPeer) {
	gp.stateMu.RLock()
	stateCopy := make(map[string]interface{})
	for k, entry := range gp.state {
		stateCopy[k] = entry.Value
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
	// If key already exists, update in place
	if _, exists := gp.state[key]; exists {
		gp.state[key] = &stateEntry{Value: value, LastUpdated: time.Now()}
	} else if len(gp.state) >= maxStateEntries {
		// At capacity - evict the oldest entry before adding
		gp.evictOldestStateLocked()
		gp.state[key] = &stateEntry{Value: value, LastUpdated: time.Now()}
	} else {
		gp.state[key] = &stateEntry{Value: value, LastUpdated: time.Now()}
	}
	gp.stateMu.Unlock()

	// Trigger immediate gossip with a timeout to prevent hanging
	// on unresponsive peers
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	gp.gossip(ctx)
}

// evictOldestStateLocked removes the oldest state entry. Must be called with stateMu held.
func (gp *GossipProtocol) evictOldestStateLocked() {
	var oldestKey string
	var oldestTime time.Time
	first := true
	for k, entry := range gp.state {
		if first || entry.LastUpdated.Before(oldestTime) {
			oldestKey = k
			oldestTime = entry.LastUpdated
			first = false
		}
	}
	if !first {
		delete(gp.state, oldestKey)
	}
}

// CleanStaleState removes state entries older than maxAge
func (gp *GossipProtocol) CleanStaleState(maxAge time.Duration) {
	gp.stateMu.Lock()
	defer gp.stateMu.Unlock()

	cutoff := time.Now().Add(-maxAge)
	for k, entry := range gp.state {
		if entry.LastUpdated.Before(cutoff) {
			delete(gp.state, k)
		}
	}
}

// GetState returns current state
func (gp *GossipProtocol) GetState(key string) (interface{}, bool) {
	gp.stateMu.RLock()
	defer gp.stateMu.RUnlock()
	entry, exists := gp.state[key]
	if !exists {
		return nil, false
	}
	return entry.Value, true
}

// HandleGossipMessage handles incoming gossip message
func (gp *GossipProtocol) HandleGossipMessage(ctx context.Context, msg *types.Message, from *types.Node) error {
	// Reject gossip from unstaked senders
	gp.peersMu.RLock()
	sv := gp.stakeVerifier
	gp.peersMu.RUnlock()
	if from != nil && sv != nil && !sv.IsStaked(ctx, from.ID) {
		return fmt.Errorf("rejecting gossip from unstaked peer: %s", from.ID.String()[:16])
	}

	var gossipMsg GossipMessage
	if err := json.Unmarshal(msg.Payload, &gossipMsg); err != nil {
		return fmt.Errorf("failed to unmarshal gossip message: %w", err)
	}

	// Merge state
	if gossipMsg.Type == "state_update" {
		if stateMap, ok := gossipMsg.Value.(map[string]interface{}); ok {
			gp.stateMu.Lock()
			now := time.Now()
			for k, v := range stateMap {
				// Check serialized size of the value to prevent memory abuse
				valBytes, err := json.Marshal(v)
				if err != nil || len(valBytes) > maxStateValueSize {
					continue // Skip oversized or non-serializable values
				}
				if _, exists := gp.state[k]; exists {
					// Update existing entry
					gp.state[k] = &stateEntry{Value: v, LastUpdated: now}
				} else if len(gp.state) < maxStateEntries {
					// Only add new entries if under the limit
					gp.state[k] = &stateEntry{Value: v, LastUpdated: now}
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

	// If peer already exists, update it
	if _, exists := gp.peers[node.ID]; exists {
		gp.peers[node.ID].Node = node
		gp.peers[node.ID].LastSeen = time.Now()
		return
	}

	// If at capacity, evict the least recently seen peer
	if len(gp.peers) >= maxPeers {
		gp.evictLeastRecentPeerLocked()
	}

	now := time.Now()
	gp.peers[node.ID] = &GossipPeer{
		Node:       node,
		LastGossip: now,
		LastSeen:   now,
	}
}

// evictLeastRecentPeerLocked removes the peer with the oldest LastSeen time.
// Must be called with peersMu held.
func (gp *GossipProtocol) evictLeastRecentPeerLocked() {
	var oldestID types.NodeID
	var oldestTime time.Time
	first := true
	for id, peer := range gp.peers {
		if first || peer.LastSeen.Before(oldestTime) {
			oldestID = id
			oldestTime = peer.LastSeen
			first = false
		}
	}
	if !first {
		delete(gp.peers, oldestID)
	}
}

// RemovePeer removes a peer from the gossip network
func (gp *GossipProtocol) RemovePeer(nodeID types.NodeID) {
	gp.peersMu.Lock()
	defer gp.peersMu.Unlock()
	delete(gp.peers, nodeID)
}
