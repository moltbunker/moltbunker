package p2p

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"github.com/moltbunker/moltbunker/internal/logging"
	"github.com/moltbunker/moltbunker/pkg/types"
)

const (
	// DefaultMaxPeers is the default maximum number of peers allowed
	DefaultMaxPeers = 50

	// DefaultConnectionTimeout is the default timeout for idle connections
	DefaultConnectionTimeout = 5 * time.Minute

	// DefaultCleanupInterval is how often to run the cleanup routine
	DefaultCleanupInterval = 1 * time.Minute

	// DefaultPeerMessageRate is the default messages per second per peer
	DefaultPeerMessageRate = 100

	// DefaultPeerMessageBurst is the default burst size per peer
	DefaultPeerMessageBurst = 200
)

// MessageSigner provides signing and verification for P2P messages
type MessageSigner interface {
	Sign(data []byte) ([]byte, error)
	Verify(data []byte, signature []byte) bool
}

// ViolationEntry tracks rate limit violations for a single peer.
type ViolationEntry struct {
	mu             sync.Mutex
	Count          int
	FirstViolation time.Time
}

// Router handles message routing and peer discovery
type Router struct {
	dht        *DHT
	transport  *Transport
	keyManager MessageSigner
	peers      map[types.NodeID]*PeerConnection
	peersMu    sync.RWMutex
	handlers   map[types.MessageType]MessageHandler
	localNode  *types.Node

	// Connection pool limits
	maxPeers          int           // Maximum number of peers allowed
	connectionTimeout time.Duration // Timeout for idle connections
	cleanupInterval   time.Duration // How often to run cleanup

	// Cleanup goroutine control
	cleanupCtx    context.Context
	cleanupCancel context.CancelFunc

	// Per-peer message rate limiters
	peerLimiters sync.Map

	// Peer banning
	banList *BanList

	// Rate limit violation tracking for auto-ban
	violationCounters sync.Map // types.NodeID -> *ViolationEntry

	// Stake verification and message gating
	stakeVerifier *StakeVerifier

	// Replay protection
	nonceTracker *NonceTracker

	// Behavioral scoring
	peerScorer *PeerScorer

	// Eclipse prevention
	diversityEnforcer *PeerDiversityEnforcer
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
func NewRouter(dht *DHT, keyManager MessageSigner) *Router {
	return NewRouterWithConfig(dht, keyManager, DefaultMaxPeers, DefaultConnectionTimeout)
}

// NewRouterWithConfig creates a new router with custom configuration
func NewRouterWithConfig(dht *DHT, keyManager MessageSigner, maxPeers int, connectionTimeout time.Duration) *Router {
	ctx, cancel := context.WithCancel(context.Background())

	r := &Router{
		dht:               dht,
		keyManager:        keyManager,
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

// SetBanList sets the ban list used by the router to reject messages from banned peers.
func (r *Router) SetBanList(bl *BanList) {
	r.banList = bl
}

// SetStakeVerifier sets the stake verifier for gating privileged messages.
func (r *Router) SetStakeVerifier(sv *StakeVerifier) {
	r.stakeVerifier = sv
}

// SetNonceTracker sets the nonce tracker for replay protection.
func (r *Router) SetNonceTracker(nt *NonceTracker) {
	r.nonceTracker = nt
}

// SetPeerScorer sets the peer scorer for behavioral tracking.
func (r *Router) SetPeerScorer(ps *PeerScorer) {
	r.peerScorer = ps
}

// SetDiversityEnforcer sets the diversity enforcer for eclipse prevention.
func (r *Router) SetDiversityEnforcer(de *PeerDiversityEnforcer) {
	r.diversityEnforcer = de
}

// BanList returns the router's ban list.
func (r *Router) BanList() *BanList {
	return r.banList
}

// DiversityEnforcer returns the router's diversity enforcer.
func (r *Router) DiversityEnforcer() *PeerDiversityEnforcer {
	return r.diversityEnforcer
}

// StakeVerifier returns the router's stake verifier.
func (r *Router) StakeVerifier() *StakeVerifier {
	return r.stakeVerifier
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
	msg.Version = types.ProtocolVersion
	if msg.Timestamp.IsZero() {
		msg.Timestamp = time.Now()
	}

	// Auto-generate nonce if zero (replay protection)
	var zeroNonce [24]byte
	if msg.Nonce == zeroNonce {
		rand.Read(msg.Nonce[:])
	}

	// Sign the message before sending
	if r.keyManager != nil {
		signableBytes := msg.SignableBytes()
		if signableBytes == nil {
			return fmt.Errorf("failed to compute signable bytes")
		}
		sig, err := r.keyManager.Sign(signableBytes)
		if err != nil {
			return fmt.Errorf("failed to sign message: %w", err)
		}
		msg.Signature = sig
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

// broadcastToPeers sends a message concurrently to each peer with a per-peer timeout.
// Returns the last error encountered, or nil if all sends succeeded.
func (r *Router) broadcastToPeers(ctx context.Context, msg *types.Message, peers []*PeerConnection) error {
	var (
		mu      sync.Mutex
		lastErr error
		wg      sync.WaitGroup
	)

	for _, peer := range peers {
		wg.Add(1)
		go func(p *PeerConnection) {
			defer wg.Done()
			sendCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()
			msgCopy := *msg
			msgCopy.To = p.Node.ID
			if err := r.SendMessage(sendCtx, p.Node.ID, &msgCopy); err != nil {
				mu.Lock()
				lastErr = err
				mu.Unlock()
			}
		}(peer)
	}

	wg.Wait()
	return lastErr
}

// BroadcastMessage broadcasts a message to all peers concurrently
func (r *Router) BroadcastMessage(ctx context.Context, msg *types.Message) error {
	r.peersMu.RLock()
	peers := make([]*PeerConnection, 0, len(r.peers))
	for _, peer := range r.peers {
		peers = append(peers, peer)
	}
	r.peersMu.RUnlock()

	return r.broadcastToPeers(ctx, msg, peers)
}

// BroadcastToRegions broadcasts a message to peers in specific regions concurrently
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

	return r.broadcastToPeers(ctx, msg, targetPeers)
}

// violationWindow is the time window within which rate limit violations are counted.
const violationWindow = 5 * time.Minute

// violationThreshold is the number of violations within the window that triggers an auto-ban.
const violationThreshold = 3

// autoBanDuration is the duration of an automatic ban triggered by repeated violations.
const autoBanDuration = 1 * time.Hour

// HandleMessage handles an incoming message through the validation pipeline:
// 1. Tiered rate limit
// 2. Ban check
// 3. Nonce + timestamp validation (replay protection)
// 4. Protocol version check
// 5. Signature verification
// 6. Stake gate (privileged messages require verified stake)
// 7. Handler dispatch
// 8. Peer scoring
func (r *Router) HandleMessage(ctx context.Context, msg *types.Message, from *types.Node) error {
	// 1. Per-peer tiered rate limiting
	if from != nil {
		limiter := r.getPeerLimiter(from.ID)
		if !limiter.Allow() {
			logging.Warn("dropping rate-limited message from peer",
				logging.NodeID(from.ID.String()[:16]),
				"message_type", string(msg.Type),
				logging.Component("router"))

			// Track violation and auto-ban if threshold exceeded
			r.recordViolation(from.ID)

			// Record in peer scorer
			if r.peerScorer != nil {
				r.peerScorer.RecordEvent(from.ID, PeerEventRateLimitHit)
			}

			return fmt.Errorf("peer rate limit exceeded: %s", from.ID.String()[:16])
		}
	}

	// 2. Check if sender is banned
	if from != nil && r.banList != nil && r.banList.IsBanned(from.ID) {
		logging.Warn("rejecting message from banned peer",
			logging.NodeID(from.ID.String()[:16]),
			"message_type", string(msg.Type),
			logging.Component("router"))
		return fmt.Errorf("peer is banned: %s", from.ID.String()[:16])
	}

	// 3. Nonce + timestamp validation (replay protection)
	if r.nonceTracker != nil {
		if err := r.nonceTracker.Check(msg.Nonce, msg.Timestamp); err != nil {
			logging.Warn("rejecting message: nonce/timestamp check failed",
				logging.NodeID(from.ID.String()[:16]),
				"message_type", string(msg.Type),
				"error", err.Error(),
				logging.Component("router"))
			if r.peerScorer != nil && from != nil {
				r.peerScorer.RecordEvent(from.ID, PeerEventInvalidMessage)
			}
			return fmt.Errorf("nonce check failed: %w", err)
		}
	}

	// 4. Check protocol version compatibility
	if msg.Version > types.ProtocolVersion {
		logging.Warn("rejecting message with unsupported protocol version",
			"message_version", msg.Version,
			"local_version", types.ProtocolVersion,
			"message_type", string(msg.Type),
			logging.Component("router"))
		return fmt.Errorf("unsupported protocol version: %d (max supported: %d)", msg.Version, types.ProtocolVersion)
	}

	// Reject messages from peers running too old a protocol version
	if msg.Version < types.MinSupportedVersion {
		logging.Warn("rejecting message with outdated protocol version",
			"message_version", msg.Version,
			"min_supported", types.MinSupportedVersion,
			"message_type", string(msg.Type),
			logging.Component("router"))
		return fmt.Errorf("outdated protocol version: %d (minimum supported: %d)", msg.Version, types.MinSupportedVersion)
	}

	// 5. Verify message signature if we have a key manager and the message has a signature
	if r.keyManager != nil && len(msg.Signature) > 0 {
		signableBytes := msg.SignableBytes()
		if signableBytes == nil {
			return fmt.Errorf("failed to compute signable bytes for verification")
		}
		if !r.keyManager.Verify(signableBytes, msg.Signature) {
			logging.Warn("rejecting message with invalid signature",
				"message_type", string(msg.Type),
				logging.Component("router"))
			if r.peerScorer != nil && from != nil {
				r.peerScorer.RecordEvent(from.ID, PeerEventInvalidMessage)
			}
			return fmt.Errorf("invalid message signature")
		}
	}

	// 6. Stake gate: privileged messages require verified on-chain stake
	if from != nil && r.stakeVerifier != nil && RequiresStake(msg.Type) {
		if !r.stakeVerifier.IsStaked(ctx, from.ID) {
			logging.Warn("rejecting staked message from unstaked peer",
				logging.NodeID(from.ID.String()[:16]),
				"message_type", string(msg.Type),
				logging.Component("router"))
			if r.peerScorer != nil {
				r.peerScorer.RecordEvent(from.ID, PeerEventInvalidMessage)
			}
			return fmt.Errorf("message type %s requires stake: peer %s", msg.Type, from.ID.String()[:16])
		}
	}

	// 7. Handler dispatch
	handler, exists := r.handlers[msg.Type]
	if !exists {
		return fmt.Errorf("no handler for message type: %s", msg.Type)
	}

	err := handler(ctx, msg, from)

	// 8. Record result in peer scorer
	if r.peerScorer != nil && from != nil {
		if err != nil {
			r.peerScorer.RecordEvent(from.ID, PeerEventMalformedPayload)
		} else {
			r.peerScorer.RecordEvent(from.ID, PeerEventValidMessage)
		}
	}

	return err
}

// getPeerLimiter returns the rate limiter for a given peer, creating one if needed.
// When a StakeVerifier is set, the limiter is sized according to the peer's staking tier.
func (r *Router) getPeerLimiter(peerID types.NodeID) *rate.Limiter {
	if val, ok := r.peerLimiters.Load(peerID); ok {
		return val.(*rate.Limiter)
	}

	var limiter *rate.Limiter
	if r.stakeVerifier != nil {
		tier := r.stakeVerifier.GetTier(context.Background(), peerID)
		limiter = NewTieredLimiter(tier)
	} else {
		limiter = rate.NewLimiter(DefaultPeerMessageRate, DefaultPeerMessageBurst)
	}
	actual, _ := r.peerLimiters.LoadOrStore(peerID, limiter)
	return actual.(*rate.Limiter)
}

// RefreshPeerLimiter recalculates the rate limiter for a peer based on their
// current staking tier. Call this after a peer's stake status changes.
func (r *Router) RefreshPeerLimiter(peerID types.NodeID) {
	if r.stakeVerifier == nil {
		return
	}
	tier := r.stakeVerifier.GetTier(context.Background(), peerID)
	limiter := NewTieredLimiter(tier)
	r.peerLimiters.Store(peerID, limiter)
}

// recordViolation increments the violation counter for a peer and auto-bans
// if the threshold is reached within the violation window.
func (r *Router) recordViolation(peerID types.NodeID) {
	now := time.Now()

	val, loaded := r.violationCounters.Load(peerID)
	if !loaded {
		entry := &ViolationEntry{
			Count:          1,
			FirstViolation: now,
		}
		r.violationCounters.Store(peerID, entry)
		return
	}

	entry := val.(*ViolationEntry)

	entry.mu.Lock()

	// If the window has elapsed, reset the counter
	if now.Sub(entry.FirstViolation) > violationWindow {
		entry.Count = 1
		entry.FirstViolation = now
		entry.mu.Unlock()
		return
	}

	entry.Count++
	shouldBan := entry.Count >= violationThreshold
	entry.mu.Unlock()

	if shouldBan {
		if r.banList != nil {
			banDur := autoBanDuration
			if r.stakeVerifier != nil {
				tier := r.stakeVerifier.GetTier(context.Background(), peerID)
				banDur = GetTierBanDuration(tier)
			}
			r.banList.Ban(peerID, "auto-ban: repeated rate limit violations", banDur)
		}
		// Reset counter after banning
		r.violationCounters.Delete(peerID)
	}
}

// CleanupStaleLimiters removes rate limiters and violation counters for peers
// that are no longer in the routing table.
func (r *Router) CleanupStaleLimiters() {
	r.peersMu.RLock()
	activePeers := make(map[types.NodeID]bool, len(r.peers))
	for id := range r.peers {
		activePeers[id] = true
	}
	r.peersMu.RUnlock()

	r.peerLimiters.Range(func(key, _ any) bool {
		if nodeID, ok := key.(types.NodeID); ok {
			if !activePeers[nodeID] {
				r.peerLimiters.Delete(key)
			}
		}
		return true
	})

	r.violationCounters.Range(func(key, _ any) bool {
		if nodeID, ok := key.(types.NodeID); ok {
			if !activePeers[nodeID] {
				r.violationCounters.Delete(key)
			}
		}
		return true
	})
}

// Close closes all peer connections and stops the cleanup routine
func (r *Router) Close() error {
	// Stop the cleanup routine
	if r.cleanupCancel != nil {
		r.cleanupCancel()
	}

	r.peersMu.Lock()
	defer r.peersMu.Unlock()

	for nodeID, peer := range r.peers {
		// Clean subsystem state before closing connections
		if r.diversityEnforcer != nil && peer.Node != nil {
			subnet := subnetFromAddress(peer.Node.Address)
			r.diversityEnforcer.RemovePeer(peer.Node.Region, subnet)
		}
		if r.peerScorer != nil {
			r.peerScorer.RemovePeer(nodeID)
		}
		if r.stakeVerifier != nil {
			r.stakeVerifier.RemovePeer(nodeID)
		}
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
