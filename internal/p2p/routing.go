package p2p

import (
	"context"
	"crypto/tls"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/moltbunker/moltbunker/pkg/types"
)

// Router handles message routing and peer discovery
type Router struct {
	dht       *DHT
	transport *Transport
	peers     map[types.NodeID]*PeerConnection
	peersMu   sync.RWMutex
	handlers  map[types.MessageType]MessageHandler
	localNode *types.Node
}

// PeerConnection represents a connection to a peer
type PeerConnection struct {
	Node      *types.Node
	Conn      *tls.Conn
	Connected bool
	LastSeen  time.Time
	mu        sync.RWMutex
}

// MessageHandler handles incoming messages
type MessageHandler func(ctx context.Context, msg *types.Message, from *types.Node) error

// NewRouter creates a new router
func NewRouter(dht *DHT) *Router {
	r := &Router{
		dht:      dht,
		peers:    make(map[types.NodeID]*PeerConnection),
		handlers: make(map[types.MessageType]MessageHandler),
	}

	if dht != nil {
		r.localNode = dht.LocalNode()

		// Set up callbacks for peer events
		dht.SetPeerConnectedCallback(func(node *types.Node) {
			r.AddPeer(node)
		})
		dht.SetPeerDisconnectedCallback(func(node *types.Node) {
			r.RemovePeer(node.ID)
		})
	}

	return r
}

// SetTransport sets the transport layer for the router
func (r *Router) SetTransport(transport *Transport) {
	r.transport = transport
}

// SetLocalNode sets the local node info
func (r *Router) SetLocalNode(node *types.Node) {
	r.localNode = node
}

// RegisterHandler registers a message handler
func (r *Router) RegisterHandler(msgType types.MessageType, handler MessageHandler) {
	r.handlers[msgType] = handler
}

// SendMessage sends a message to a peer
func (r *Router) SendMessage(ctx context.Context, to types.NodeID, msg *types.Message) error {
	r.peersMu.RLock()
	peerConn, exists := r.peers[to]
	r.peersMu.RUnlock()

	if !exists {
		// Find peer via DHT
		nodes, err := r.dht.FindNode(ctx, to)
		if err != nil {
			return fmt.Errorf("failed to find peer: %w", err)
		}

		if len(nodes) == 0 {
			return fmt.Errorf("peer not found: %s", to.String()[:16])
		}

		// Use the first node found
		node := nodes[0]

		// Validate address
		if node.Address == "" && node.Port == 0 {
			return fmt.Errorf("peer has no valid address")
		}

		peerConn = &PeerConnection{
			Node:      node,
			Connected: false,
			LastSeen:  time.Now(),
		}

		r.peersMu.Lock()
		r.peers[to] = peerConn
		r.peersMu.Unlock()
	}

	// Ensure we have a connection
	if err := r.ensureConnection(ctx, peerConn); err != nil {
		return fmt.Errorf("failed to establish connection: %w", err)
	}

	// Set message metadata
	if r.localNode != nil {
		msg.From = r.localNode.ID
	}
	msg.To = to
	if msg.Timestamp.IsZero() {
		msg.Timestamp = time.Now()
	}

	// Serialize the message
	msgData, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to serialize message: %w", err)
	}

	// Send with length prefix (4 bytes, big endian)
	peerConn.mu.Lock()
	err = r.sendLengthPrefixed(peerConn.Conn, msgData)
	if err != nil {
		// Mark connection as disconnected on error
		peerConn.Connected = false
		peerConn.Conn = nil
		peerConn.mu.Unlock()
		return fmt.Errorf("failed to send message: %w", err)
	}
	peerConn.LastSeen = time.Now()
	peerConn.mu.Unlock()

	return nil
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

// sendLengthPrefixed sends data with a 4-byte length prefix
func (r *Router) sendLengthPrefixed(conn *tls.Conn, data []byte) error {
	if conn == nil {
		return fmt.Errorf("connection is nil")
	}

	// Set write deadline
	conn.SetWriteDeadline(time.Now().Add(30 * time.Second))

	// Write length prefix (4 bytes, big endian)
	lengthBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lengthBuf, uint32(len(data)))

	if _, err := conn.Write(lengthBuf); err != nil {
		return err
	}

	// Write data
	if _, err := conn.Write(data); err != nil {
		return err
	}

	return nil
}

// ReadLengthPrefixed reads data with a 4-byte length prefix
func ReadLengthPrefixed(conn io.Reader) ([]byte, error) {
	// Read length prefix
	lengthBuf := make([]byte, 4)
	if _, err := io.ReadFull(conn, lengthBuf); err != nil {
		return nil, err
	}

	length := binary.BigEndian.Uint32(lengthBuf)

	// Sanity check - max message size 16MB
	if length > 16*1024*1024 {
		return nil, fmt.Errorf("message too large: %d bytes", length)
	}

	// Read data
	data := make([]byte, length)
	if _, err := io.ReadFull(conn, data); err != nil {
		return nil, err
	}

	return data, nil
}

// BroadcastMessage broadcasts a message to all peers
func (r *Router) BroadcastMessage(ctx context.Context, msg *types.Message) error {
	r.peersMu.RLock()
	peers := make([]*PeerConnection, 0, len(r.peers))
	for _, peer := range r.peers {
		peers = append(peers, peer)
	}
	r.peersMu.RUnlock()

	var lastErr error
	for _, peer := range peers {
		msgCopy := *msg
		msgCopy.To = peer.Node.ID
		if err := r.SendMessage(ctx, peer.Node.ID, &msgCopy); err != nil {
			lastErr = err
			continue
		}
	}

	return lastErr
}

// BroadcastToRegions broadcasts a message to peers in specific regions
func (r *Router) BroadcastToRegions(ctx context.Context, msg *types.Message, regions []string) error {
	r.peersMu.RLock()
	var targetPeers []*PeerConnection
	for _, peer := range r.peers {
		for _, region := range regions {
			if peer.Node.Region == region {
				targetPeers = append(targetPeers, peer)
				break
			}
		}
	}
	r.peersMu.RUnlock()

	var lastErr error
	for _, peer := range targetPeers {
		msgCopy := *msg
		msgCopy.To = peer.Node.ID
		if err := r.SendMessage(ctx, peer.Node.ID, &msgCopy); err != nil {
			lastErr = err
			continue
		}
	}

	return lastErr
}

// HandleMessage handles an incoming message
func (r *Router) HandleMessage(ctx context.Context, msg *types.Message, from *types.Node) error {
	handler, exists := r.handlers[msg.Type]
	if !exists {
		return fmt.Errorf("no handler for message type: %s", msg.Type)
	}

	return handler(ctx, msg, from)
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

// AddPeer adds a peer to the routing table
func (r *Router) AddPeer(node *types.Node) {
	r.peersMu.Lock()
	defer r.peersMu.Unlock()

	r.peers[node.ID] = &PeerConnection{
		Node:      node,
		Connected: false,
		LastSeen:  time.Now(),
	}
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

// Close closes all peer connections
func (r *Router) Close() error {
	r.peersMu.Lock()
	defer r.peersMu.Unlock()

	for _, peer := range r.peers {
		if peer.Conn != nil {
			peer.Conn.Close()
		}
	}
	r.peers = make(map[types.NodeID]*PeerConnection)

	return nil
}
