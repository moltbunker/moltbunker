package p2p

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/moltbunker/moltbunker/internal/logging"
	"github.com/moltbunker/moltbunker/internal/util"
	"github.com/moltbunker/moltbunker/pkg/types"
)

// startCleanupRoutine starts the background goroutine for cleaning up idle connections
func (r *Router) startCleanupRoutine() {
	util.SafeGoWithName("router-cleanup", func() {
		ticker := time.NewTicker(r.cleanupInterval)
		defer ticker.Stop()

		for {
			select {
			case <-r.cleanupCtx.Done():
				return
			case <-ticker.C:
				r.cleanupIdleConnections()
			}
		}
	})
}

// cleanupIdleConnections removes connections that have been idle too long
func (r *Router) cleanupIdleConnections() {
	r.peersMu.Lock()
	defer r.peersMu.Unlock()

	now := time.Now()
	for nodeID, peer := range r.peers {
		peer.mu.RLock()
		lastSeen := peer.LastSeen
		connected := peer.Connected
		peer.mu.RUnlock()

		// Only cleanup idle connected peers
		if connected && now.Sub(lastSeen) > r.connectionTimeout {
			peer.mu.Lock()
			if peer.Conn != nil {
				logging.Debug("closing idle connection",
					logging.NodeID(nodeID.String()[:16]),
					"idle_time", now.Sub(lastSeen).String(),
					logging.Component("router"))
				peer.Conn.Close()
				peer.Conn = nil
				peer.Connected = false
			}
			peer.mu.Unlock()
		}
	}
}

// ensureConnection ensures we have an active connection to the peer
func (r *Router) ensureConnection(ctx context.Context, peerConn *PeerConnection) error {
	peerConn.mu.Lock()
	defer peerConn.mu.Unlock()

	if peerConn.Connected && peerConn.Conn != nil {
		return nil
	}

	if r.transport == nil {
		return fmt.Errorf("transport not configured")
	}

	// Build address from node info
	address := r.resolveAddress(peerConn.Node)
	if address == "" {
		return fmt.Errorf("cannot resolve peer address")
	}

	// Dial the peer with timeout
	dialCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	conn, err := r.transport.DialContext(dialCtx, peerConn.Node.ID, address)
	if err != nil {
		return fmt.Errorf("failed to dial peer %s: %w", address, err)
	}

	peerConn.Conn = conn
	peerConn.Connected = true
	peerConn.LastSeen = time.Now()

	return nil
}

// resolveAddress resolves the address for a peer
func (r *Router) resolveAddress(node *types.Node) string {
	// If address already contains port, use it directly
	if node.Address != "" {
		// Check if it's a multiaddr format
		if strings.HasPrefix(node.Address, "/ip4/") || strings.HasPrefix(node.Address, "/ip6/") {
			// Extract IP and port from multiaddr
			return extractTCPAddressFromMultiaddr(node.Address)
		}

		// Check if it already has a port
		if strings.Contains(node.Address, ":") {
			return node.Address
		}

		// Append port if we have one
		if node.Port > 0 {
			return fmt.Sprintf("%s:%d", node.Address, node.Port)
		}

		// Use default port
		return fmt.Sprintf("%s:9000", node.Address)
	}

	// No address, can't connect
	return ""
}

// extractTCPAddressFromMultiaddr extracts IP:port from multiaddr string
func extractTCPAddressFromMultiaddr(addr string) string {
	// /ip4/1.2.3.4/tcp/9000 -> 1.2.3.4:9000
	parts := strings.Split(addr, "/")

	var ip, port string
	for i := 0; i < len(parts)-1; i++ {
		switch parts[i] {
		case "ip4", "ip6":
			if i+1 < len(parts) {
				ip = parts[i+1]
			}
		case "tcp":
			if i+1 < len(parts) {
				port = parts[i+1]
			}
		}
	}

	if ip != "" && port != "" {
		return net.JoinHostPort(ip, port)
	}
	if ip != "" {
		return net.JoinHostPort(ip, "9000")
	}

	return ""
}
