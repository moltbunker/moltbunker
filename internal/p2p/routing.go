package p2p

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/moltbunker/moltbunker/pkg/types"
)

const (
	// DefaultMaxPeers is the default maximum number of peers allowed
	DefaultMaxPeers = 50

	// DefaultConnectionTimeout is the default timeout for idle connections
	DefaultConnectionTimeout = 5 * time.Minute

	// DefaultCleanupInterval is how often to run the cleanup routine
	DefaultCleanupInterval = 1 * time.Minute
)

// Router handles message routing and peer discovery
type Router struct {
	dht       *DHT
	transport *Transport
	peers     map[types.NodeID]*PeerConnection
	peersMu   sync.RWMutex
	handlers  map[types.MessageType]MessageHandler
	localNode *types.Node

	// Connection pool limits
	maxPeers          int           // Maximum number of peers allowed
	connectionTimeout time.Duration // Timeout for idle connections
	cleanupInterval   time.Duration // How often to run cleanup

	// Cleanup goroutine control
	cleanupCtx    context.Context
	cleanupCancel context.CancelFunc
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
	return NewRouterWithConfig(dht, DefaultMaxPeers, DefaultConnectionTimeout)
}

// NewRouterWithConfig creates a new router with custom configuration
func NewRouterWithConfig(dht *DHT, maxPeers int, connectionTimeout time.Duration) *Router {
	ctx, cancel := context.WithCancel(context.Background())

	r := &Router{
		dht:               dht,
		peers:             make(map[types.NodeID]*PeerConnection),
		handlers:          make(map[types.MessageType]MessageHandler),
		maxPeers:          maxPeers,
		connectionTimeout: connectionTimeout,
		cleanupInterval:   DefaultCleanupInterval,
		cleanupCtx:        ctx,
		cleanupCancel:     cancel,
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

	// Start the idle connection cleanup routine
	r.startCleanupRoutine()

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

// Close closes all peer connections and stops the cleanup routine
func (r *Router) Close() error {
	// Stop the cleanup routine
	if r.cleanupCancel != nil {
		r.cleanupCancel()
	}

	r.peersMu.Lock()
	defer r.peersMu.Unlock()

	for _, peer := range r.peers {
		peer.mu.Lock()
		if peer.Conn != nil {
			peer.Conn.Close()
		}
		peer.mu.Unlock()
	}
	r.peers = make(map[types.NodeID]*PeerConnection)

	return nil
}

// MaxPeers returns the maximum number of peers allowed
func (r *Router) MaxPeers() int {
	return r.maxPeers
}

// SetMaxPeers sets the maximum number of peers allowed
func (r *Router) SetMaxPeers(maxPeers int) {
	r.maxPeers = maxPeers
}

// ConnectionTimeout returns the idle connection timeout
func (r *Router) ConnectionTimeout() time.Duration {
	return r.connectionTimeout
}

// SetConnectionTimeout sets the idle connection timeout
func (r *Router) SetConnectionTimeout(timeout time.Duration) {
	r.connectionTimeout = timeout
}
