package daemon

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/moltbunker/moltbunker/internal/config"
	"github.com/moltbunker/moltbunker/internal/identity"
	"github.com/moltbunker/moltbunker/internal/logging"
	"github.com/moltbunker/moltbunker/internal/p2p"
	"github.com/moltbunker/moltbunker/internal/payment"
	"github.com/moltbunker/moltbunker/internal/runtime"
	"github.com/moltbunker/moltbunker/internal/security"
	"github.com/moltbunker/moltbunker/internal/util"
	"github.com/moltbunker/moltbunker/pkg/types"
)

const (
	// defaultMaxConcurrentConnections is the default max connections if not configured
	defaultMaxConcurrentConnections = 100

	// defaultConnectionReadTimeout is the default timeout for reading from a connection
	defaultConnectionReadTimeout = 5 * time.Minute

	// maxMessagePayloadSize is the maximum allowed message payload size (10MB)
	maxMessagePayloadSize = 10 * 1024 * 1024
)

// Node represents a P2P node in the network
type Node struct {
	keyManager     *identity.KeyManager
	walletManager  *identity.WalletManager
	paymentService *payment.PaymentService
	dht            *p2p.DHT
	router         *p2p.Router
	transport      *p2p.Transport
	geolocator     *p2p.GeoLocator
	nodeInfo       *types.Node
	listener       net.Listener
	mu             sync.RWMutex
	running        bool

	// Connection limiting (configurable)
	connSemaphore      chan struct{}  // Semaphore for limiting concurrent connections
	activeConns        int64          // Atomic counter for active connections
	maxConcurrentConns int            // Maximum concurrent connections (from config)
	connReadTimeout    time.Duration  // Connection read timeout (from config)

	// Stake verification and Sybil resistance
	stakeVerifier *p2p.StakeVerifier
	subnetLimiter *p2p.SubnetLimiter
	nonceTracker  *p2p.NonceTracker

	// Certificate pinning store (persisted across restarts)
	certPinStore *security.CertPinStore
}

// NewNode creates a new P2P node
func NewNode(ctx context.Context, keyPath string, keystoreDir string, port int) (*Node, error) {
	// Initialize key manager
	keyManager, err := identity.NewKeyManager(keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create key manager: %w", err)
	}

	// Load wallet if one exists (nil = read-only mode)
	walletManager, err := identity.LoadWalletManager(keystoreDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load wallet: %w", err)
	}

	// Initialize DHT with configuration
	dhtConfig := &p2p.DHTConfig{
		Port:           port,
		BootstrapPeers: p2p.DefaultBootstrapPeers(),
		EnableMDNS:     true,
		EnableNAT:      true,
		MaxPeers:       50,
	}
	dht, err := p2p.NewDHT(ctx, dhtConfig, keyManager)
	if err != nil {
		return nil, fmt.Errorf("failed to create DHT: %w", err)
	}

	// Initialize router
	router := p2p.NewRouter(dht, keyManager)

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
		keyManager:         keyManager,
		walletManager:      walletManager,
		dht:                dht,
		router:             router,
		transport:          transport,
		geolocator:         geolocator,
		nodeInfo:           nodeInfo,
		connSemaphore:      make(chan struct{}, defaultMaxConcurrentConnections),
		maxConcurrentConns: defaultMaxConcurrentConnections,
		connReadTimeout:    defaultConnectionReadTimeout,
		certPinStore:       pinStore,
	}, nil
}

// NewNodeWithConfig creates a new P2P node from configuration
func NewNodeWithConfig(ctx context.Context, cfg *config.Config) (*Node, error) {
	// Initialize key manager
	keyManager, err := identity.NewKeyManager(cfg.Daemon.KeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create key manager: %w", err)
	}

	// Load wallet if one exists (nil = read-only mode, no announce/staking)
	walletManager, err := identity.LoadWalletManager(cfg.Daemon.KeystoreDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load wallet: %w", err)
	}

	// Build DHT configuration from config
	bootstrapCfg := p2p.DefaultBootstrapConfig()
	if len(cfg.P2P.BootstrapHTTPEndpoints) > 0 {
		bootstrapCfg.HTTPEndpoints = cfg.P2P.BootstrapHTTPEndpoints
	}

	dhtConfig := &p2p.DHTConfig{
		Port:           cfg.Daemon.Port,
		BootstrapPeers: cfg.P2P.BootstrapNodes,
		EnableMDNS:     cfg.P2P.EnableMDNS,
		EnableNAT:      cfg.P2P.EnableNAT,
		ExternalIP:     cfg.P2P.ExternalIP,
		AnnounceAddrs:  cfg.P2P.AnnounceAddrs,
		MaxPeers:       cfg.P2P.MaxPeers,
		DNSBootstrap:   bootstrapCfg,
	}

	// Add default bootstrap peers if none configured
	if len(dhtConfig.BootstrapPeers) == 0 {
		dhtConfig.BootstrapPeers = p2p.BootstrapPeerStrings(bootstrapCfg)
	}

	// Initialize DHT
	dht, err := p2p.NewDHT(ctx, dhtConfig, keyManager)
	if err != nil {
		return nil, fmt.Errorf("failed to create DHT: %w", err)
	}

	// Initialize router
	router := p2p.NewRouter(dht, keyManager)

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

	// Initialize geolocator and detect region
	geolocator := p2p.NewGeoLocator()
	location := types.NodeLocation{Region: "Unknown"}
	if cfg.Node.Region != "" {
		// Manual region override from config
		location.Region = cfg.Node.Region
		logging.Info("Using configured region override",
			"region", location.Region,
			logging.Component("daemon"))
	} else if loc, err := geolocator.GetLocationFromIP(""); err == nil {
		location = types.NodeLocation{
			Region:      p2p.GetRegionFromCountry(loc.CountryCode),
			Country:     loc.CountryCode,
			CountryName: loc.Country,
			City:        loc.City,
			Lat:         loc.Lat,
			Lon:         loc.Lon,
		}
		logging.Info("Detected node location",
			"country", loc.Country,
			"country_code", loc.CountryCode,
			"city", loc.City,
			"region", location.Region,
			"lat", loc.Lat,
			"lon", loc.Lon,
			"ip", loc.IP,
			logging.Component("daemon"))
	} else {
		logging.Warn("Failed to detect location and no override configured, using Unknown",
			"error", err.Error(),
			logging.Component("daemon"))
	}

	// Detect provider tier from runtime capabilities
	providerTier := runtime.DetectProviderTier()

	// Create node info
	nodeInfo := &types.Node{
		ID:           keyManager.NodeID(),
		PublicKey:    keyManager.PublicKey(),
		Port:         cfg.Daemon.Port,
		Region:       location.Region,
		Country:      location.Country,
		Location:     location,
		ProviderTier: providerTier,
		Capabilities: types.NodeCapabilities{
			ContainerRuntime: true,
			TorSupport:       cfg.Tor.Enabled,
		},
	}

	// Set transport on router
	router.SetTransport(transport)

	// Initialize stake verifier (PaymentService set later via SetPaymentService)
	stakeVerifier := p2p.NewStakeVerifier(nil)
	router.SetStakeVerifier(stakeVerifier)

	// Initialize nonce tracker for replay protection
	nonceTracker := p2p.NewNonceTracker()
	router.SetNonceTracker(nonceTracker)

	// Initialize diversity enforcer for eclipse prevention
	diversityEnforcer := p2p.NewPeerDiversityEnforcer()
	router.SetDiversityEnforcer(diversityEnforcer)

	// Initialize subnet limiter for Sybil resistance
	maxPeersPerSubnet := cfg.P2P.MaxPeersPerSubnet
	if maxPeersPerSubnet <= 0 {
		maxPeersPerSubnet = p2p.DefaultMaxPeersPerSubnet
	}
	subnetLimiter := p2p.NewSubnetLimiter(maxPeersPerSubnet)

	// Initialize ban list and peer scorer, wire into router
	banList := p2p.NewBanList()
	router.SetBanList(banList)

	peerScorer := p2p.NewPeerScorer(nil) // uses default config
	router.SetPeerScorer(peerScorer)

	// Get connection limits from config (use defaults if not set)
	maxConns := cfg.API.MaxConcurrentConns
	if maxConns <= 0 {
		maxConns = defaultMaxConcurrentConnections
	}

	connTimeout := defaultConnectionReadTimeout
	if cfg.API.IdleTimeoutSecs > 0 {
		connTimeout = time.Duration(cfg.API.IdleTimeoutSecs) * time.Second
	}

	return &Node{
		keyManager:         keyManager,
		walletManager:      walletManager,
		dht:                dht,
		router:             router,
		transport:          transport,
		geolocator:         geolocator,
		nodeInfo:           nodeInfo,
		connSemaphore:      make(chan struct{}, maxConns),
		maxConcurrentConns: maxConns,
		connReadTimeout:    connTimeout,
		stakeVerifier:      stakeVerifier,
		subnetLimiter:      subnetLimiter,
		nonceTracker:       nonceTracker,
		certPinStore:       pinStore,
	}, nil
}

// Start starts the node
func (n *Node) Start(ctx context.Context) error {
	n.mu.Lock()
	if n.running {
		n.mu.Unlock()
		return fmt.Errorf("node already running")
	}

	// Start TLS listener on port+1 (libp2p DHT uses the base port)
	tlsPort := n.nodeInfo.Port + 1
	listener, err := n.transport.Listen(fmt.Sprintf(":%d", tlsPort))
	if err != nil {
		n.mu.Unlock()
		return fmt.Errorf("failed to start listener: %w", err)
	}

	n.listener = listener
	n.running = true
	n.mu.Unlock()

	// Register default message handlers
	n.registerHandlers()

	// Handle connections using SafeGo for panic recovery
	util.SafeGoWithName("node-connection-handler", func() {
		n.handleConnections(ctx, listener)
	})

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

	// Nodes handler — response to FindNode, contains discovered peers
	n.router.RegisterHandler(types.MessageTypeNodes, func(ctx context.Context, msg *types.Message, from *types.Node) error {
		var nodes []*types.Node
		if err := json.Unmarshal(msg.Payload, &nodes); err != nil {
			logging.Debug("failed to unmarshal nodes response", logging.Err(err))
			return nil // Don't propagate parse errors
		}
		for _, node := range nodes {
			if node != nil {
				n.dht.AddPeer(node)
			}
		}
		return nil
	})

	// Bid handler — placeholder for future auction-based provider selection
	n.router.RegisterHandler(types.MessageTypeBid, func(ctx context.Context, msg *types.Message, from *types.Node) error {
		logging.Debug("received bid message (not yet implemented)",
			"from", from.ID.String()[:16],
			logging.Component("daemon"))
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
				// Check if listener was closed (expected during shutdown)
				select {
				case <-ctx.Done():
					return
				default:
					logging.Warn("failed to accept connection",
						logging.Err(err),
						logging.Component("node"))
					continue
				}
			}

			// Subnet Sybil check before accepting
			if n.subnetLimiter != nil {
				// Use a temporary NodeID for the check — will be replaced after TLS handshake
				var tempID types.NodeID
				if !n.subnetLimiter.Allow(conn.RemoteAddr().String(), tempID) {
					logging.Warn("subnet peer limit reached, rejecting connection",
						"remote_addr", conn.RemoteAddr().String(),
						logging.Component("node"))
					conn.Close()
					continue
				}
			}

			// Try to acquire a connection slot (non-blocking)
			select {
			case n.connSemaphore <- struct{}{}:
				// Got a slot, handle the connection
				atomic.AddInt64(&n.activeConns, 1)
				util.SafeGoWithName("node-handle-connection", func() {
					defer func() {
						<-n.connSemaphore
						atomic.AddInt64(&n.activeConns, -1)
					}()
					n.handleConnection(ctx, conn)
				})
			default:
				// At connection limit, reject the connection
				logging.Warn("connection limit reached, rejecting connection",
					"active_connections", atomic.LoadInt64(&n.activeConns),
					"max_connections", n.maxConcurrentConns,
					"remote_addr", conn.RemoteAddr().String(),
					logging.Component("node"))
				conn.Close()
			}
		}
	}
}

// handleConnection handles a single connection
func (n *Node) handleConnection(ctx context.Context, conn net.Conn) {
	defer conn.Close()

	// Track which NodeID is registered in the subnet limiter so we can clean it
	// up when the connection closes. Starts as the zero-value temp ID (registered
	// in handleConnections), and is updated to the real peerNodeID after TLS.
	var subnetNodeID types.NodeID
	if n.subnetLimiter != nil {
		defer func() {
			n.subnetLimiter.Remove(subnetNodeID)
		}()
	}

	// Create connection-scoped context for cleanup goroutines
	connCtx, connCancel := context.WithCancel(ctx)
	defer connCancel() // Ensures the grace timer goroutine exits when connection closes

	// Assert TLS connection
	tlsConn, ok := conn.(*tls.Conn)
	if !ok {
		logging.Warn("received non-TLS connection",
			"remote_addr", conn.RemoteAddr().String(),
			logging.Component("node"))
		return
	}

	// Perform TLS handshake
	if err := tlsConn.HandshakeContext(ctx); err != nil {
		logging.Debug("TLS handshake failed",
			"remote_addr", conn.RemoteAddr().String(),
			logging.Err(err),
			logging.Component("node"))
		return
	}

	// Get peer certificate info
	state := tlsConn.ConnectionState()
	var peerNodeID types.NodeID
	if len(state.PeerCertificates) > 0 {
		// Extract node ID from certificate using SHA256 hash of the entire
		// subject public key info. This avoids the zero-value collision issue
		// where short keys would result in all-zero node IDs, making different
		// peers appear to be the same node.
		pkInfo := state.PeerCertificates[0].RawSubjectPublicKeyInfo
		hash := sha256.Sum256(pkInfo)
		copy(peerNodeID[:], hash[:])
	}

	// Reject zero NodeID — indicates empty PeerCertificates or cert parsing failure.
	// A zero NodeID would bypass per-peer rate limiting and subnet enforcement.
	var zeroNodeID types.NodeID
	if peerNodeID == zeroNodeID {
		logging.Warn("zero NodeID from TLS certificate, disconnecting",
			"remote_addr", conn.RemoteAddr().String(),
			logging.Component("node"))
		return
	}

	// Re-check subnet limiter with real NodeID (replacing temp).
	// Remove the temp registration first, then re-check with real ID.
	if n.subnetLimiter != nil {
		// Remove the zero-value temp entry
		n.subnetLimiter.Remove(subnetNodeID)
		if !n.subnetLimiter.Allow(conn.RemoteAddr().String(), peerNodeID) {
			// Allow failed — nothing new was registered, clear subnetNodeID
			// so the deferred Remove is a harmless no-op.
			subnetNodeID = types.NodeID{}
			logging.Warn("subnet peer limit reached after TLS, disconnecting",
				logging.NodeID(peerNodeID.String()[:16]),
				"remote_addr", conn.RemoteAddr().String(),
				logging.Component("node"))
			return
		}
		// Allow succeeded — peerNodeID is now registered; update for deferred cleanup
		subnetNodeID = peerNodeID
	}

	// Create peer node info
	peerNode := &types.Node{
		ID:      peerNodeID,
		Address: conn.RemoteAddr().String(),
	}

	// Send our announce message to prove wallet ownership
	if n.walletManager != nil && n.stakeVerifier != nil {
		n.sendAnnounce(tlsConn)
	}

	// Add peer to DHT
	n.dht.AddPeer(peerNode)

	// Start announce grace period timer — peer must send announce within 30s
	// or we disconnect. Using a channel so the message loop can cancel it.
	announceReceived := make(chan struct{}, 1)
	const announceGracePeriod = 30 * time.Second
	if n.stakeVerifier != nil {
		util.SafeGoWithName("announce-grace-timer", func() {
			timer := time.NewTimer(announceGracePeriod)
			defer timer.Stop()
			select {
			case <-announceReceived:
				// Peer announced in time
			case <-timer.C:
				// Grace period expired — if peer still hasn't announced, close conn
				if !n.stakeVerifier.HasAnnounced(peerNodeID) {
					logging.Warn("peer did not announce within grace period, disconnecting",
						logging.NodeID(peerNodeID.String()[:16]),
						logging.Component("node"))
					conn.Close()
				}
			case <-connCtx.Done():
				// Connection closed or parent context cancelled
			}
		})
	}

	// Message reading loop
	for {
		select {
		case <-ctx.Done():
			return
		default:
			// Set read deadline before reading
			if err := conn.SetReadDeadline(time.Now().Add(n.connReadTimeout)); err != nil {
				logging.Warn("failed to set read deadline",
					"remote_addr", conn.RemoteAddr().String(),
					logging.Err(err),
					logging.Component("node"))
				return
			}

			// Read length-prefixed message
			msgData, err := p2p.ReadLengthPrefixed(tlsConn)
			if err != nil {
				if err == io.EOF {
					logging.Debug("connection closed by peer",
						"remote_addr", conn.RemoteAddr().String(),
						logging.Component("node"))
					return
				}
				// Check if it's a timeout error
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					logging.Debug("connection read timeout",
						"remote_addr", conn.RemoteAddr().String(),
						logging.Component("node"))
					return
				}
				logging.Warn("failed to read message from peer",
					"remote_addr", conn.RemoteAddr().String(),
					logging.Err(err),
					logging.Component("node"))
				return
			}

			// Deserialize message
			var msg types.Message
			if err := json.Unmarshal(msgData, &msg); err != nil {
				logging.Warn("failed to unmarshal message",
					"remote_addr", conn.RemoteAddr().String(),
					logging.Err(err),
					logging.Component("node"))
				continue
			}

			// C2: Validate msg.From matches TLS-authenticated identity
			if msg.From != peerNodeID {
				logging.Warn("message From mismatch with TLS peer identity, dropping",
					"claimed_from", msg.From.String()[:16],
					logging.NodeID(peerNodeID.String()[:16]),
					"message_type", string(msg.Type),
					logging.Component("node"))
				continue
			}

			// H2: Reject oversized message payloads
			if len(msg.Payload) > maxMessagePayloadSize {
				logging.Warn("dropping oversized message",
					"payload_size", len(msg.Payload),
					"max_size", maxMessagePayloadSize,
					logging.Component("node"))
				continue
			}

			// Handle announce messages directly (identity exchange)
			if msg.Type == types.MessageTypeAnnounce {
				// M6: Rate limit announce messages — reject duplicates
				if n.stakeVerifier != nil && n.stakeVerifier.HasAnnounced(peerNodeID) {
					logging.Debug("duplicate announce from peer, ignoring",
						logging.NodeID(peerNodeID.String()[:16]),
						logging.Component("node"))
					continue
				}
				n.handleAnnounceMessage(&msg, peerNode)
				// Signal grace period timer that announce was received
				select {
				case announceReceived <- struct{}{}:
				default:
				}
				continue
			}

			// Route message to handler
			if err := n.router.HandleMessage(ctx, &msg, peerNode); err != nil {
				logging.Warn("message handler error",
					"message_type", msg.Type,
					"remote_addr", conn.RemoteAddr().String(),
					logging.Err(err),
					logging.Component("node"))
				continue
			}
		}
	}
}

// sendAnnounce creates and sends an announce message over a TLS connection.
// The wallet's private key must already be unlocked (cached in WalletManager)
// by the daemon startup code before the node starts accepting connections.
func (n *Node) sendAnnounce(conn *tls.Conn) {
	// Snapshot walletManager under lock to prevent data race with SetWalletManager.
	n.mu.RLock()
	wm := n.walletManager
	n.mu.RUnlock()

	if wm == nil {
		return
	}

	// PrivateKey("") works if the key was already unlocked and cached.
	// If wallet password is required but not yet provided, this will fail gracefully.
	privKey, err := wm.PrivateKey("")
	if err != nil {
		logging.Debug("cannot get wallet key for announce, skipping (wallet may not be unlocked)",
			logging.Err(err),
			logging.Component("node"))
		return
	}

	payload, err := p2p.CreateAnnouncePayload(n.nodeInfo.ID, wm.Address(), privKey)
	if err != nil {
		logging.Warn("failed to create announce payload",
			logging.Err(err),
			logging.Component("node"))
		return
	}

	// Include local provider tier in announce
	payload.ProviderTier = n.nodeInfo.ProviderTier

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		logging.Warn("failed to marshal announce payload",
			logging.Err(err),
			logging.Component("node"))
		return
	}

	msg := &types.Message{
		Type:      types.MessageTypeAnnounce,
		From:      n.nodeInfo.ID,
		Payload:   payloadBytes,
		Timestamp: time.Now(),
		Version:   types.ProtocolVersion,
	}

	// Write directly to the connection (can't use SendMessage for self)
	msgData, err := json.Marshal(msg)
	if err != nil {
		logging.Warn("failed to marshal announce message",
			logging.Err(err),
			logging.Component("node"))
		return
	}

	if err := p2p.WriteLengthPrefixed(conn, msgData); err != nil {
		logging.Warn("failed to send announce",
			logging.Err(err),
			logging.Component("node"))
	}
}

// handleAnnounceMessage processes an announce message from a peer.
func (n *Node) handleAnnounceMessage(msg *types.Message, peerNode *types.Node) {
	if n.stakeVerifier == nil {
		return
	}

	var payload types.AnnouncePayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		logging.Warn("failed to unmarshal announce payload",
			logging.Err(err),
			"remote", peerNode.Address,
			logging.Component("node"))
		return
	}

	// Verify the signature and recover wallet address
	recoveredAddr, err := p2p.VerifyAnnouncePayload(&payload)
	if err != nil {
		logging.Warn("announce verification failed",
			logging.Err(err),
			logging.NodeID(peerNode.ID.String()[:16]),
			logging.Component("node"))
		return
	}

	// Verify the claimed NodeID matches the TLS-authenticated peer
	if payload.NodeID != peerNode.ID.String() {
		logging.Warn("announce NodeID mismatch with TLS peer",
			"claimed", payload.NodeID[:16],
			logging.NodeID(peerNode.ID.String()[:16]),
			logging.Component("node"))
		return
	}

	// Register the NodeID → wallet mapping
	n.stakeVerifier.RegisterPeerWallet(peerNode.ID, recoveredAddr)
	peerNode.WalletAddress = recoveredAddr

	// Set provider tier from announce (self-reported; safe — see design docs)
	if payload.ProviderTier != "" {
		peerNode.ProviderTier = payload.ProviderTier
	}

	logging.Info("peer announced wallet",
		logging.NodeID(peerNode.ID.String()[:16]),
		"wallet", recoveredAddr.Hex()[:10],
		"tier", string(payload.ProviderTier),
		logging.Component("node"))
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
		logging.Info("closing TLS listener", logging.Component("node"))
		if err := n.listener.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close listener: %w", err))
		}
	}

	// Close router connections (disconnects all peers)
	if n.router != nil {
		logging.Info("closing router and disconnecting peers", logging.Component("node"))
		if err := n.router.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close router: %w", err))
		}
	}

	// Close DHT
	if n.dht != nil {
		logging.Info("closing DHT", logging.Component("node"))
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

	util.SafeGoWithName("node-shutdown", func() {
		done <- n.Close()
	})

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return fmt.Errorf("shutdown timed out: %w", ctx.Err())
	}
}

// ActiveConnections returns the number of active connections
func (n *Node) ActiveConnections() int64 {
	return atomic.LoadInt64(&n.activeConns)
}

// Router returns the node's router
func (n *Node) Router() *p2p.Router {
	return n.router
}

// SetPaymentService sets the payment service for use by the container manager
// and forwards it to the stake verifier for on-chain stake lookups.
func (n *Node) SetPaymentService(ps *payment.PaymentService) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.paymentService = ps

	// Forward to stake verifier so it can query on-chain state
	if n.stakeVerifier != nil {
		n.stakeVerifier.SetPaymentService(ps)
	}
}

// SetWalletManager sets the wallet manager on the node.
// Called from daemon main after loading/creating the wallet.
func (n *Node) SetWalletManager(wm *identity.WalletManager) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.walletManager = wm
}

// WalletManager returns the node's wallet manager (may be nil in read-only mode).
func (n *Node) WalletManager() *identity.WalletManager {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.walletManager
}

// PaymentService returns the node's payment service
func (n *Node) PaymentService() *payment.PaymentService {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.paymentService
}

// WalletAddress returns the node's wallet address
func (n *Node) WalletAddress() common.Address {
	n.mu.RLock()
	wm := n.walletManager
	n.mu.RUnlock()

	if wm == nil {
		return common.Address{}
	}
	return wm.Address()
}

// IsRunning returns whether the node is running
func (n *Node) IsRunning() bool {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.running
}

// NonceTracker returns the node's nonce tracker for replay protection cleanup.
func (n *Node) NonceTracker() *p2p.NonceTracker {
	return n.nonceTracker
}

// CertPinStore returns the node's certificate pinning store.
func (n *Node) CertPinStore() *security.CertPinStore {
	return n.certPinStore
}

// KeyManager returns the node's key manager.
func (n *Node) KeyManager() *identity.KeyManager {
	return n.keyManager
}

// BanList returns the ban list from the router, if set.
func (n *Node) BanList() *p2p.BanList {
	if n.router != nil {
		return n.router.BanList()
	}
	return nil
}
