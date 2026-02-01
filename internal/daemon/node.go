package daemon

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/moltbunker/moltbunker/internal/config"
	"github.com/moltbunker/moltbunker/internal/identity"
	"github.com/moltbunker/moltbunker/internal/p2p"
	"github.com/moltbunker/moltbunker/internal/security"
	"github.com/moltbunker/moltbunker/pkg/types"
)

// Node represents a P2P node in the network
type Node struct {
	keyManager    *identity.KeyManager
	walletManager *identity.WalletManager
	dht           *p2p.DHT
	router        *p2p.Router
	transport     *p2p.Transport
	geolocator    *p2p.GeoLocator
	nodeInfo      *types.Node
	listener      net.Listener
	mu            sync.RWMutex
	running       bool
}

// NewNode creates a new P2P node
func NewNode(ctx context.Context, keyPath string, keystoreDir string, port int) (*Node, error) {
	// Initialize key manager
	keyManager, err := identity.NewKeyManager(keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create key manager: %w", err)
	}

	// Initialize wallet manager
	walletManager, err := identity.NewWalletManager(keystoreDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet manager: %w", err)
	}

	// Initialize DHT with configuration
	dhtConfig := &p2p.DHTConfig{
		Port:           port,
		BootstrapPeers: p2p.DefaultBootstrapPeers(),
		EnableMDNS:     true,
		MaxPeers:       50,
	}
	dht, err := p2p.NewDHT(ctx, dhtConfig, keyManager)
	if err != nil {
		return nil, fmt.Errorf("failed to create DHT: %w", err)
	}

	// Initialize router
	router := p2p.NewRouter(dht)

	// Initialize certificate manager
	certManager, err := identity.NewCertificateManager(keyManager)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate manager: %w", err)
	}

	// Initialize pin store
	pinStore := security.NewCertPinStore()

	// Initialize transport
	transport, err := p2p.NewTransport(certManager, pinStore)
	if err != nil {
		return nil, fmt.Errorf("failed to create transport: %w", err)
	}

	// Initialize geolocator
	geolocator := p2p.NewGeoLocator()

	// Create node info
	nodeInfo := &types.Node{
		ID:        keyManager.NodeID(),
		PublicKey: keyManager.PublicKey(),
		Port:      port,
		Capabilities: types.NodeCapabilities{
			ContainerRuntime: true,
			TorSupport:      true,
		},
	}

	// Set transport on router
	router.SetTransport(transport)

	return &Node{
		keyManager:    keyManager,
		walletManager: walletManager,
		dht:           dht,
		router:        router,
		transport:     transport,
		geolocator:    geolocator,
		nodeInfo:      nodeInfo,
	}, nil
}

// NewNodeWithConfig creates a new P2P node from configuration
func NewNodeWithConfig(ctx context.Context, cfg *config.Config) (*Node, error) {
	// Initialize key manager
	keyManager, err := identity.NewKeyManager(cfg.Daemon.KeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create key manager: %w", err)
	}

	// Initialize wallet manager
	walletManager, err := identity.NewWalletManager(cfg.Daemon.KeystoreDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet manager: %w", err)
	}

	// Build DHT configuration from config
	dhtConfig := &p2p.DHTConfig{
		Port:           cfg.Daemon.Port,
		BootstrapPeers: cfg.P2P.BootstrapNodes,
		EnableMDNS:     cfg.P2P.EnableMDNS,
		ExternalIP:     cfg.P2P.ExternalIP,
		AnnounceAddrs:  cfg.P2P.AnnounceAddrs,
		MaxPeers:       cfg.P2P.MaxPeers,
	}

	// Add default bootstrap peers if none configured
	if len(dhtConfig.BootstrapPeers) == 0 {
		dhtConfig.BootstrapPeers = p2p.DefaultBootstrapPeers()
	}

	// Initialize DHT
	dht, err := p2p.NewDHT(ctx, dhtConfig, keyManager)
	if err != nil {
		return nil, fmt.Errorf("failed to create DHT: %w", err)
	}

	// Initialize router
	router := p2p.NewRouter(dht)

	// Initialize certificate manager
	certManager, err := identity.NewCertificateManager(keyManager)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate manager: %w", err)
	}

	// Initialize pin store
	pinStore := security.NewCertPinStore()

	// Initialize transport
	transport, err := p2p.NewTransport(certManager, pinStore)
	if err != nil {
		return nil, fmt.Errorf("failed to create transport: %w", err)
	}

	// Set dial timeout from config
	if cfg.P2P.DialTimeout > 0 {
		transport.SetDialTimeout(time.Duration(cfg.P2P.DialTimeout) * time.Second)
	}

	// Initialize geolocator
	geolocator := p2p.NewGeoLocator()

	// Create node info
	nodeInfo := &types.Node{
		ID:        keyManager.NodeID(),
		PublicKey: keyManager.PublicKey(),
		Port:      cfg.Daemon.Port,
		Capabilities: types.NodeCapabilities{
			ContainerRuntime: true,
			TorSupport:       cfg.Tor.Enabled,
		},
	}

	// Set transport on router
	router.SetTransport(transport)

	return &Node{
		keyManager:    keyManager,
		walletManager: walletManager,
		dht:           dht,
		router:        router,
		transport:     transport,
		geolocator:    geolocator,
		nodeInfo:      nodeInfo,
	}, nil
}

// Start starts the node
func (n *Node) Start(ctx context.Context) error {
	n.mu.Lock()
	if n.running {
		n.mu.Unlock()
		return fmt.Errorf("node already running")
	}

	// Start listening
	listener, err := n.transport.Listen(fmt.Sprintf(":%d", n.nodeInfo.Port))
	if err != nil {
		n.mu.Unlock()
		return fmt.Errorf("failed to start listener: %w", err)
	}

	n.listener = listener
	n.running = true
	n.mu.Unlock()

	// Register default message handlers
	n.registerHandlers()

	// Handle connections
	go n.handleConnections(ctx, listener)

	return nil
}

// registerHandlers registers default message handlers
func (n *Node) registerHandlers() {
	// Ping handler
	n.router.RegisterHandler(types.MessageTypePing, func(ctx context.Context, msg *types.Message, from *types.Node) error {
		// Respond with pong
		pong := &types.Message{
			Type:      types.MessageTypePong,
			From:      n.nodeInfo.ID,
			To:        from.ID,
			Timestamp: msg.Timestamp,
		}
		return n.router.SendMessage(ctx, from.ID, pong)
	})

	// Pong handler
	n.router.RegisterHandler(types.MessageTypePong, func(ctx context.Context, msg *types.Message, from *types.Node) error {
		// Update peer last seen
		n.dht.AddPeer(from)
		return nil
	})

	// Find node handler
	n.router.RegisterHandler(types.MessageTypeFindNode, func(ctx context.Context, msg *types.Message, from *types.Node) error {
		// Parse target from payload
		var targetID types.NodeID
		copy(targetID[:], msg.Payload)

		// Find closest nodes
		nodes, err := n.dht.FindNode(ctx, targetID)
		if err != nil {
			return err
		}

		// Serialize and send response
		nodesData, err := json.Marshal(nodes)
		if err != nil {
			return err
		}

		response := &types.Message{
			Type:      types.MessageTypeNodes,
			From:      n.nodeInfo.ID,
			To:        from.ID,
			Payload:   nodesData,
			Timestamp: msg.Timestamp,
		}
		return n.router.SendMessage(ctx, from.ID, response)
	})

	// Health handler
	n.router.RegisterHandler(types.MessageTypeHealth, func(ctx context.Context, msg *types.Message, from *types.Node) error {
		// Update peer health info
		n.dht.AddPeer(from)
		return nil
	})
}

// handleConnections handles incoming connections
func (n *Node) handleConnections(ctx context.Context, listener net.Listener) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			conn, err := listener.Accept()
			if err != nil {
				continue
			}

			go n.handleConnection(ctx, conn)
		}
	}
}

// handleConnection handles a single connection
func (n *Node) handleConnection(ctx context.Context, conn net.Conn) {
	defer conn.Close()

	// Assert TLS connection
	tlsConn, ok := conn.(*tls.Conn)
	if !ok {
		return
	}

	// Perform TLS handshake
	if err := tlsConn.HandshakeContext(ctx); err != nil {
		return
	}

	// Get peer certificate info
	state := tlsConn.ConnectionState()
	var peerNodeID types.NodeID
	if len(state.PeerCertificates) > 0 {
		// Extract node ID from certificate (use subject public key info hash)
		pkInfo := state.PeerCertificates[0].RawSubjectPublicKeyInfo
		if len(pkInfo) >= 32 {
			copy(peerNodeID[:], pkInfo[:32])
		}
	}

	// Create peer node info
	peerNode := &types.Node{
		ID:      peerNodeID,
		Address: conn.RemoteAddr().String(),
	}

	// Add peer to DHT
	n.dht.AddPeer(peerNode)

	// Message reading loop
	for {
		select {
		case <-ctx.Done():
			return
		default:
			// Read length-prefixed message
			msgData, err := p2p.ReadLengthPrefixed(tlsConn)
			if err != nil {
				if err == io.EOF {
					return
				}
				return
			}

			// Deserialize message
			var msg types.Message
			if err := json.Unmarshal(msgData, &msg); err != nil {
				continue
			}

			// Route message to handler
			if err := n.router.HandleMessage(ctx, &msg, peerNode); err != nil {
				// Log error but continue processing
				continue
			}
		}
	}
}

// NodeInfo returns node information
func (n *Node) NodeInfo() *types.Node {
	return n.nodeInfo
}

// Close closes the node gracefully
func (n *Node) Close() error {
	n.mu.Lock()
	if !n.running {
		n.mu.Unlock()
		return nil
	}
	n.running = false
	n.mu.Unlock()

	var errs []error

	// Close listener first to stop accepting new connections
	if n.listener != nil {
		if err := n.listener.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close listener: %w", err))
		}
	}

	// Close router connections
	if n.router != nil {
		if err := n.router.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close router: %w", err))
		}
	}

	// Close DHT
	if n.dht != nil {
		if err := n.dht.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close DHT: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("shutdown errors: %v", errs)
	}
	return nil
}

// Shutdown performs a graceful shutdown with timeout
func (n *Node) Shutdown(ctx context.Context) error {
	done := make(chan error, 1)

	go func() {
		done <- n.Close()
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return fmt.Errorf("shutdown timed out: %w", ctx.Err())
	}
}

// Router returns the node's router
func (n *Node) Router() *p2p.Router {
	return n.router
}

// IsRunning returns whether the node is running
func (n *Node) IsRunning() bool {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.running
}
