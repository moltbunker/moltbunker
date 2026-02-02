package p2p

import (
	"time"

	"github.com/moltbunker/moltbunker/internal/logging"
	"github.com/moltbunker/moltbunker/pkg/types"
)

// GetPeers returns all known peers
func (r *Router) GetPeers() []*types.Node {
	// First get peers from DHT (includes connected libp2p peers)
	if r.dht != nil {
		return r.dht.GetConnectedPeers()
	}

	// Fallback to local peer list
	r.peersMu.RLock()
	defer r.peersMu.RUnlock()

	peers := make([]*types.Node, 0, len(r.peers))
	for _, peer := range r.peers {
		peers = append(peers, peer.Node)
	}
	return peers
}

// GetPeersByRegion returns peers in a specific region
func (r *Router) GetPeersByRegion(region string) []*types.Node {
	r.peersMu.RLock()
	defer r.peersMu.RUnlock()

	var peers []*types.Node
	for _, peer := range r.peers {
		if peer.Node.Region == region {
			peers = append(peers, peer.Node)
		}
	}
	return peers
}

// AddPeer adds a peer to the routing table
// Returns true if the peer was added, false if at capacity
func (r *Router) AddPeer(node *types.Node) bool {
	r.peersMu.Lock()
	defer r.peersMu.Unlock()

	// Check if peer already exists
	if existing, exists := r.peers[node.ID]; exists {
		// Update existing peer's last seen time
		existing.mu.Lock()
		existing.LastSeen = time.Now()
		existing.mu.Unlock()
		return true
	}

	// Check if we're at capacity
	if len(r.peers) >= r.maxPeers {
		// Try to evict the oldest idle peer
		if !r.evictOldestIdlePeerLocked() {
			logging.Warn("peer limit reached, rejecting new peer",
				logging.NodeID(node.ID.String()[:16]),
				"max_peers", r.maxPeers,
				"current_peers", len(r.peers),
				logging.Component("router"))
			return false
		}
	}

	r.peers[node.ID] = &PeerConnection{
		Node:      node,
		Connected: false,
		LastSeen:  time.Now(),
	}
	return true
}

// evictOldestIdlePeerLocked evicts the oldest idle (disconnected) peer
// Must be called with peersMu held
func (r *Router) evictOldestIdlePeerLocked() bool {
	var oldestID types.NodeID
	var oldestTime time.Time
	found := false

	for nodeID, peer := range r.peers {
		peer.mu.RLock()
		connected := peer.Connected
		lastSeen := peer.LastSeen
		peer.mu.RUnlock()

		// Only consider disconnected peers for eviction
		if !connected {
			if !found || lastSeen.Before(oldestTime) {
				oldestID = nodeID
				oldestTime = lastSeen
				found = true
			}
		}
	}

	if found {
		if peer, exists := r.peers[oldestID]; exists {
			peer.mu.Lock()
			if peer.Conn != nil {
				peer.Conn.Close()
			}
			peer.mu.Unlock()
		}
		delete(r.peers, oldestID)
		logging.Debug("evicted oldest idle peer",
			logging.NodeID(oldestID.String()[:16]),
			logging.Component("router"))
		return true
	}

	return false
}

// RemovePeer removes a peer from the routing table
func (r *Router) RemovePeer(nodeID types.NodeID) {
	r.peersMu.Lock()
	defer r.peersMu.Unlock()

	if peerConn, exists := r.peers[nodeID]; exists {
		if peerConn.Conn != nil {
			peerConn.Conn.Close()
		}
		delete(r.peers, nodeID)
	}
}

// PeerCount returns the number of known peers
func (r *Router) PeerCount() int {
	if r.dht != nil {
		return r.dht.PeerCount()
	}

	r.peersMu.RLock()
	defer r.peersMu.RUnlock()
	return len(r.peers)
}

// CleanupStaleConnections removes stale peer connections
func (r *Router) CleanupStaleConnections(maxAge time.Duration) {
	r.peersMu.Lock()
	defer r.peersMu.Unlock()

	now := time.Now()
	for nodeID, peer := range r.peers {
		if now.Sub(peer.LastSeen) > maxAge {
			if peer.Conn != nil {
				peer.Conn.Close()
			}
			delete(r.peers, nodeID)
		}
	}
}
