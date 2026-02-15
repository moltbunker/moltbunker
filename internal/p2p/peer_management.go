package p2p

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/moltbunker/moltbunker/internal/logging"
	"github.com/moltbunker/moltbunker/pkg/types"
)

// subnetFromAddress extracts a /16 subnet prefix from a peer address string.
// Handles multiaddr, host:port, and bare IP formats.
func subnetFromAddress(address string) string {
	if address == "" {
		return ""
	}

	// Extract IP from multiaddr (e.g. /ip4/1.2.3.4/tcp/9000)
	ip := address
	if strings.HasPrefix(address, "/ip4/") || strings.HasPrefix(address, "/ip6/") {
		parts := strings.Split(address, "/")
		if len(parts) > 2 {
			ip = parts[2]
		}
	} else {
		// host:port format
		host, _, err := net.SplitHostPort(address)
		if err == nil {
			ip = host
		}
	}

	parsed := net.ParseIP(ip)
	if parsed == nil {
		return ""
	}

	v4 := parsed.To4()
	if v4 != nil {
		return fmt.Sprintf("%d.%d.0.0/16", v4[0], v4[1])
	}

	// IPv6: /48 prefix
	return fmt.Sprintf("%x:%x:%x::/48", parsed[0:2], parsed[2:4], parsed[4:6])
}

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

// AddPeer adds a peer to the routing table.
// Returns true if the peer was added, false if at capacity or rejected by diversity check.
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

	// Eclipse prevention: check region/subnet diversity before adding
	if r.diversityEnforcer != nil {
		subnet := subnetFromAddress(node.Address)
		region := node.Region
		if !r.diversityEnforcer.AllowPeer(region, subnet) {
			logging.Debug("peer rejected by diversity enforcer",
				logging.NodeID(node.ID.String()[:16]),
				"region", region,
				"subnet", subnet,
				logging.Component("router"))
			return false
		}
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

	// Track in diversity enforcer
	if r.diversityEnforcer != nil {
		subnet := subnetFromAddress(node.Address)
		r.diversityEnforcer.RecordPeer(node.Region, subnet)
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
			// Clean diversity, scorer, and stake verifier before eviction
			if r.diversityEnforcer != nil {
				subnet := subnetFromAddress(peer.Node.Address)
				r.diversityEnforcer.RemovePeer(peer.Node.Region, subnet)
			}
			if r.peerScorer != nil {
				r.peerScorer.RemovePeer(oldestID)
			}
			if r.stakeVerifier != nil {
				r.stakeVerifier.RemovePeer(oldestID)
			}
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
		// Remove from diversity tracking
		if r.diversityEnforcer != nil {
			subnet := subnetFromAddress(peerConn.Node.Address)
			r.diversityEnforcer.RemovePeer(peerConn.Node.Region, subnet)
		}
		// Remove from peer scorer to prevent stale entry accumulation
		if r.peerScorer != nil {
			r.peerScorer.RemovePeer(nodeID)
		}
		// Remove from stake verifier to prevent peerWallets leak
		if r.stakeVerifier != nil {
			r.stakeVerifier.RemovePeer(nodeID)
		}
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
			// Clean diversity, scorer, and stake verifier before removing
			if r.diversityEnforcer != nil {
				subnet := subnetFromAddress(peer.Node.Address)
				r.diversityEnforcer.RemovePeer(peer.Node.Region, subnet)
			}
			if r.peerScorer != nil {
				r.peerScorer.RemovePeer(nodeID)
			}
			if r.stakeVerifier != nil {
				r.stakeVerifier.RemovePeer(nodeID)
			}
			if peer.Conn != nil {
				peer.Conn.Close()
			}
			delete(r.peers, nodeID)
		}
	}
}
