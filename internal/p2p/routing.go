package p2p

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/moltbunker/moltbunker/pkg/types"
)

// Router handles message routing and peer discovery
type Router struct {
	dht      *DHT
	peers    map[types.NodeID]*PeerConnection
	peersMu  sync.RWMutex
	handlers map[types.MessageType]MessageHandler
}

// PeerConnection represents a connection to a peer
type PeerConnection struct {
	Node      *types.Node
	Connected bool
	LastSeen  time.Time
	mu        sync.RWMutex
}

// MessageHandler handles incoming messages
type MessageHandler func(ctx context.Context, msg *types.Message, from *types.Node) error

// NewRouter creates a new router
func NewRouter(dht *DHT) *Router {
	return &Router{
		dht:      dht,
		peers:    make(map[types.NodeID]*PeerConnection),
		handlers: make(map[types.MessageType]MessageHandler),
	}
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
			return fmt.Errorf("peer not found")
		}

		peerConn = &PeerConnection{
			Node:      nodes[0],
			Connected: false,
			LastSeen:  time.Now(),
		}

		r.peersMu.Lock()
		r.peers[to] = peerConn
		r.peersMu.Unlock()
	}

	// TODO: Actually send message via transport
	// This is a placeholder - would use Transport.Dial and send over TLS

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

	for _, peer := range peers {
		msgCopy := *msg
		msgCopy.To = peer.Node.ID
		if err := r.SendMessage(ctx, peer.Node.ID, &msgCopy); err != nil {
			// Log error but continue
			continue
		}
	}

	return nil
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
	r.peersMu.RLock()
	defer r.peersMu.RUnlock()

	peers := make([]*types.Node, 0, len(r.peers))
	for _, peer := range r.peers {
		peers = append(peers, peer.Node)
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
