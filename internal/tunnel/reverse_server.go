package tunnel

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/hashicorp/yamux"
	"github.com/moltbunker/moltbunker/internal/logging"
	"github.com/moltbunker/moltbunker/internal/p2p"
	"github.com/moltbunker/moltbunker/pkg/types"
)

const (
	// defaultMaxConns is the global connection cap.
	defaultMaxConns = 10000

	// defaultMaxPerIP is the maximum concurrent connections per remote IP.
	defaultMaxPerIP = 5

	// tlsHandshakeTimeout is the deadline for the TLS handshake.
	tlsHandshakeTimeout = 10 * time.Second

	// controlStreamTimeout is the deadline for reading the registration request.
	controlStreamTimeout = 15 * time.Second

	// heartbeatInterval is how often the ingress pings providers.
	heartbeatInterval = 30 * time.Second

	// heartbeatTimeout is the deadline for a pong response.
	heartbeatTimeout = 10 * time.Second

	// maxMissedHeartbeats is how many consecutive missed pongs trigger teardown.
	maxMissedHeartbeats = 3

	// idleReapTimeout is how long a tunnel can be idle before being reaped.
	idleReapTimeout = 10 * time.Minute
)

// ReverseServerOption configures a ReverseServer.
type ReverseServerOption func(*ReverseServer)

// WithMaxConns sets the global connection cap.
func WithMaxConns(n int) ReverseServerOption {
	return func(s *ReverseServer) { s.maxConns = n }
}

// WithMaxPerIP sets the per-IP connection limit.
func WithMaxPerIP(n int) ReverseServerOption {
	return func(s *ReverseServer) { s.maxPerIP = n }
}

// WithHMACSecret sets the HMAC secret for reconnection tokens.
func WithHMACSecret(secret []byte) ReverseServerOption {
	return func(s *ReverseServer) { s.hmacSecret = secret }
}

// WithDomain sets the base domain for full domain generation.
func WithDomain(domain string) ReverseServerOption {
	return func(s *ReverseServer) { s.domain = domain }
}

// WalletVerifyFunc verifies wallet proof and returns the staking tier.
// Returns the tier string ("free", "starter", "bronze", etc.) or error.
type WalletVerifyFunc func(proof *WalletProof, nodeID string) (tier string, err error)

// WithWalletVerifier sets the wallet verification function for staked tiers.
func WithWalletVerifier(fn WalletVerifyFunc) ReverseServerOption {
	return func(s *ReverseServer) { s.verifyWallet = fn }
}

// ReverseServer accepts outbound connections from providers,
// establishes yamux sessions, and registers subdomains.
type ReverseServer struct {
	listener     net.Listener
	registry     *TunnelRegistry
	nonceTracker *p2p.NonceTracker
	hmacSecret   []byte
	domain       string // e.g., "moltbunker.dev"
	maxConns     int
	maxPerIP     int
	verifyWallet WalletVerifyFunc

	// Connection tracking
	connSemaphore chan struct{}
	ipTracker     sync.Map // string → *int32 (IP → count)

	// Bandwidth limiters per subdomain
	limiters   map[string]*BandwidthLimiter
	limitersMu sync.RWMutex

	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewReverseServer creates a new reverse tunnel server.
func NewReverseServer(listener net.Listener, opts ...ReverseServerOption) *ReverseServer {
	s := &ReverseServer{
		listener:     listener,
		registry:     NewTunnelRegistry(),
		nonceTracker: p2p.NewNonceTrackerWithConfig(maxRegistrationAge, 30*time.Second, 10*time.Minute),
		domain:       "moltbunker.dev",
		maxConns:     defaultMaxConns,
		maxPerIP:     defaultMaxPerIP,
		limiters:     make(map[string]*BandwidthLimiter),
	}

	// Generate HMAC secret if not provided
	s.hmacSecret = make([]byte, 32)
	rand.Read(s.hmacSecret)

	for _, opt := range opts {
		opt(s)
	}

	s.connSemaphore = make(chan struct{}, s.maxConns)
	return s
}

// Registry returns the tunnel registry for external lookups (e.g., by the ingress proxy).
func (s *ReverseServer) Registry() *TunnelRegistry {
	return s.registry
}

// Serve accepts and handles provider connections. Blocks until ctx is cancelled.
func (s *ReverseServer) Serve(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	s.cancel = cancel
	defer cancel()

	// Periodic nonce cleanup
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.nonceTracker.CleanExpired()
			}
		}
	}()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				logging.Error("reverse tunnel accept error",
					logging.Err(err),
					logging.Component("reverse-tunnel"))
				continue
			}
		}

		// Check per-IP limit before TLS handshake
		remoteIP := extractIP(conn.RemoteAddr())
		if !s.checkIPLimit(remoteIP) {
			logging.Debug("reverse tunnel rejected: per-IP limit",
				"ip", remoteIP,
				logging.Component("reverse-tunnel"))
			conn.Close()
			continue
		}

		// Check global connection cap
		select {
		case s.connSemaphore <- struct{}{}:
		default:
			logging.Warn("reverse tunnel rejected: global connection limit reached",
				logging.Component("reverse-tunnel"))
			conn.Close()
			s.releaseIPSlot(remoteIP)
			continue
		}

		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			defer func() { <-s.connSemaphore }()
			defer s.releaseIPSlot(remoteIP)
			s.handleProviderConn(ctx, conn)
		}()
	}
}

// Shutdown gracefully stops the server.
func (s *ReverseServer) Shutdown(ctx context.Context) error {
	if s.cancel != nil {
		s.cancel()
	}
	err := s.listener.Close()

	// Wait for all handlers to finish (with timeout from ctx)
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
	}
	return err
}

// OpenStream opens a yamux stream to the provider for a given subdomain.
// Called by the ingress proxy when routing an HTTP request.
func (s *ReverseServer) OpenStream(subdomain string) (net.Conn, error) {
	sess, ok := s.registry.Lookup(subdomain)
	if !ok {
		return nil, fmt.Errorf("no reverse tunnel for subdomain %q", subdomain)
	}

	if sess.YamuxSess == nil || sess.YamuxSess.IsClosed() {
		s.registry.Unregister(subdomain)
		return nil, fmt.Errorf("reverse tunnel session closed for %q", subdomain)
	}

	// Check RPS limit
	s.limitersMu.RLock()
	limiter := s.limiters[subdomain]
	s.limitersMu.RUnlock()
	if limiter != nil && !limiter.AllowRequest() {
		return nil, fmt.Errorf("rate limit exceeded for subdomain %q", subdomain)
	}

	stream, err := sess.YamuxSess.Open()
	if err != nil {
		return nil, fmt.Errorf("open yamux stream: %w", err)
	}

	// Wrap with bandwidth limiter if available
	if limiter != nil {
		return limiter.WrapConn(stream), nil
	}
	return stream, nil
}

// handleProviderConn handles a single provider connection.
func (s *ReverseServer) handleProviderConn(ctx context.Context, conn net.Conn) {
	defer conn.Close()

	// Derive NodeID from TLS peer certificate
	tlsConn, ok := conn.(*tls.Conn)
	if !ok {
		logging.Error("reverse tunnel: non-TLS connection",
			logging.Component("reverse-tunnel"))
		return
	}

	// Complete TLS handshake with timeout
	if err := tlsConn.SetDeadline(time.Now().Add(tlsHandshakeTimeout)); err != nil {
		return
	}
	if err := tlsConn.Handshake(); err != nil {
		logging.Debug("reverse tunnel TLS handshake failed",
			logging.Err(err),
			logging.Component("reverse-tunnel"))
		return
	}
	tlsConn.SetDeadline(time.Time{}) // Clear deadline

	tlsState := tlsConn.ConnectionState()
	if len(tlsState.PeerCertificates) == 0 {
		logging.Debug("reverse tunnel: no peer certificate",
			logging.Component("reverse-tunnel"))
		return
	}

	// NodeID = SHA256(SPKI)
	spki := tlsState.PeerCertificates[0].RawSubjectPublicKeyInfo
	nodeIDBytes := sha256.Sum256(spki)
	var nodeID types.NodeID
	copy(nodeID[:], nodeIDBytes[:])

	logging.Info("reverse tunnel connection",
		"node_id", nodeID.String()[:16],
		"remote", conn.RemoteAddr().String(),
		logging.Component("reverse-tunnel"))

	// Wrap with yamux — provider is yamux server, ingress is yamux client.
	// We create a yamux Client session because the ingress initiates streams.
	yamuxCfg := yamux.DefaultConfig()
	yamuxCfg.MaxStreamWindowSize = 256 * 1024 // 256KB
	yamuxCfg.StreamOpenTimeout = 10 * time.Second
	yamuxCfg.EnableKeepAlive = true
	yamuxCfg.KeepAliveInterval = 30 * time.Second
	yamuxCfg.ConnectionWriteTimeout = 10 * time.Second
	yamuxCfg.LogOutput = io.Discard // Suppress yamux internal logging

	session, err := yamux.Client(conn, yamuxCfg)
	if err != nil {
		logging.Error("reverse tunnel yamux setup failed",
			logging.Err(err),
			logging.Component("reverse-tunnel"))
		return
	}
	defer session.Close()

	// Open control stream (stream 0) to read registration request
	ctrlStream, err := session.Open()
	if err != nil {
		logging.Error("reverse tunnel: failed to open control stream",
			logging.Err(err),
			logging.Component("reverse-tunnel"))
		return
	}
	defer ctrlStream.Close()

	ctrlStream.SetDeadline(time.Now().Add(controlStreamTimeout))

	// Read registration request
	msgType, payload, err := readControlMsg(ctrlStream)
	if err != nil {
		logging.Debug("reverse tunnel: read registration failed",
			logging.Err(err),
			logging.Component("reverse-tunnel"))
		return
	}
	if msgType != MsgTunnelRegister {
		writeControlMsg(ctrlStream, MsgTunnelError, []byte("expected TUNNEL_REGISTER"))
		return
	}

	var req TunnelRegisterRequest
	if err := json.Unmarshal(payload, &req); err != nil {
		writeControlMsg(ctrlStream, MsgTunnelError, []byte("invalid registration payload"))
		return
	}

	// Determine tier: verify wallet proof if provided, otherwise free tier
	tier := "free"
	if req.WalletProof != nil && s.verifyWallet != nil {
		verified, verifyErr := s.verifyWallet(req.WalletProof, nodeID.String())
		if verifyErr != nil {
			logging.Warn("reverse tunnel: wallet proof verification failed",
				"node_id", nodeID.String()[:16],
				logging.Err(verifyErr),
				logging.Component("reverse-tunnel"))
			// Fall through to free tier — don't reject, just don't upgrade
		} else {
			tier = verified
			logging.Debug("reverse tunnel: wallet verified",
				"node_id", nodeID.String()[:16],
				"tier", tier,
				logging.Component("reverse-tunnel"))
		}
	}

	// Validate registration (nonce, timestamp, TLS binding)
	// Skip validation for reconnection tokens
	if req.ReconnToken != "" {
		subdomain := req.Subdomain
		if subdomain == "" {
			writeControlMsg(ctrlStream, MsgTunnelError, []byte("reconnection requires subdomain"))
			return
		}
		if !ValidateReconnToken(s.hmacSecret, req.ReconnToken, nodeID, subdomain,
			time.Duration(ReconnGraceForTier(tier))*time.Second) {
			writeControlMsg(ctrlStream, MsgTunnelError, []byte("invalid reconnection token"))
			return
		}
	} else {
		if err := ValidateRegistration(&req, tlsState, s.nonceTracker); err != nil {
			writeControlMsg(ctrlStream, MsgTunnelError, []byte(fmt.Sprintf("registration rejected: %v", err)))
			return
		}
	}

	// Check subdomain cap
	maxSubs := MaxSubdomainsForTier(tier)
	if s.registry.CountForNodeID(nodeID) >= maxSubs {
		writeControlMsg(ctrlStream, MsgTunnelError,
			[]byte(fmt.Sprintf("subdomain limit reached (%d/%d)", s.registry.CountForNodeID(nodeID), maxSubs)))
		return
	}

	// Assign or validate subdomain
	subdomain := req.Subdomain
	if subdomain == "" {
		// Free tier: auto-assign random subdomain
		sub, err := s.registry.AssignRandomSubdomain()
		if err != nil {
			writeControlMsg(ctrlStream, MsgTunnelError, []byte("subdomain assignment failed"))
			return
		}
		subdomain = sub
	} else if tier == "free" {
		// Free tier cannot request vanity subdomains
		writeControlMsg(ctrlStream, MsgTunnelError, []byte("free tier cannot request vanity subdomains"))
		return
	}

	// Create and register session
	limits := LimitsForTier(tier)
	reconnToken := IssueReconnToken(s.hmacSecret, nodeID, subdomain)
	tunnelSess := &TunnelSession{
		NodeID:       nodeID,
		Subdomain:    subdomain,
		DeploymentID: req.DeploymentID,
		YamuxSess:    session,
		RegisteredAt: time.Now(),
		Tier:         tier,
		Limits:       limits,
		ReconnToken:  reconnToken,
	}

	if err := s.registry.Register(subdomain, tunnelSess); err != nil {
		writeControlMsg(ctrlStream, MsgTunnelError, []byte(err.Error()))
		return
	}
	defer s.registry.Unregister(subdomain)

	// Create bandwidth limiter
	limiter := NewBandwidthLimiter(limits)
	s.limitersMu.Lock()
	s.limiters[subdomain] = limiter
	s.limitersMu.Unlock()
	defer func() {
		s.limitersMu.Lock()
		delete(s.limiters, subdomain)
		s.limitersMu.Unlock()
	}()

	// Send registration response
	resp := TunnelRegisterResponse{
		Subdomain:   subdomain,
		ReconnToken: reconnToken,
		Limits:      limits,
		FullDomain:  subdomain + "." + s.domain,
	}
	respPayload, _ := json.Marshal(resp)
	if err := writeControlMsg(ctrlStream, MsgTunnelRegistered, respPayload); err != nil {
		return
	}

	ctrlStream.SetDeadline(time.Time{}) // Clear deadline

	logging.Info("reverse tunnel registered",
		"subdomain", subdomain,
		"node_id", nodeID.String()[:16],
		"tier", tier,
		"full_domain", resp.FullDomain,
		logging.Component("reverse-tunnel"))

	// Run heartbeat loop until session dies or context cancelled
	s.heartbeatLoop(ctx, ctrlStream, session, subdomain, nodeID)
}

// heartbeatLoop sends periodic pings to the provider and tears down on failure.
func (s *ReverseServer) heartbeatLoop(ctx context.Context, ctrl net.Conn,
	session *yamux.Session, subdomain string, nodeID types.NodeID) {

	ticker := time.NewTicker(heartbeatInterval)
	defer ticker.Stop()

	missed := 0
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if session.IsClosed() {
				logging.Info("reverse tunnel session closed",
					"subdomain", subdomain,
					logging.Component("reverse-tunnel"))
				return
			}

			// Send ping with challenge nonce
			var challenge [16]byte
			rand.Read(challenge[:])
			ping := TunnelPing{Challenge: challenge}
			pingPayload, _ := json.Marshal(ping)

			ctrl.SetWriteDeadline(time.Now().Add(heartbeatTimeout))
			if err := writeControlMsg(ctrl, MsgTunnelPing, pingPayload); err != nil {
				missed++
				logging.Debug("reverse tunnel ping failed",
					"subdomain", subdomain,
					"missed", missed,
					logging.Err(err),
					logging.Component("reverse-tunnel"))
				if missed >= maxMissedHeartbeats {
					logging.Warn("reverse tunnel teardown: missed heartbeats",
						"subdomain", subdomain,
						"missed", missed,
						logging.Component("reverse-tunnel"))
					return
				}
				continue
			}

			// Read pong
			ctrl.SetReadDeadline(time.Now().Add(heartbeatTimeout))
			msgType, pongPayload, err := readControlMsg(ctrl)
			if err != nil || msgType != MsgTunnelPong {
				missed++
				if missed >= maxMissedHeartbeats {
					logging.Warn("reverse tunnel teardown: missed heartbeats",
						"subdomain", subdomain,
						"missed", missed,
						logging.Component("reverse-tunnel"))
					return
				}
				continue
			}

			// Verify challenge nonce
			var pong TunnelPong
			if err := json.Unmarshal(pongPayload, &pong); err != nil || pong.Challenge != challenge {
				missed++
				if missed >= maxMissedHeartbeats {
					return
				}
				continue
			}

			missed = 0 // Reset on successful pong
			ctrl.SetDeadline(time.Time{})
		}
	}
}

// checkIPLimit checks if a new connection from this IP is allowed.
func (s *ReverseServer) checkIPLimit(ip string) bool {
	val, _ := s.ipTracker.LoadOrStore(ip, new(int32))
	counter := val.(*int32)
	for {
		old := atomic.LoadInt32(counter)
		if int(old) >= s.maxPerIP {
			return false
		}
		if atomic.CompareAndSwapInt32(counter, old, old+1) {
			return true
		}
	}
}

// releaseIPSlot decrements the IP connection counter.
func (s *ReverseServer) releaseIPSlot(ip string) {
	if val, ok := s.ipTracker.Load(ip); ok {
		counter := val.(*int32)
		if atomic.AddInt32(counter, -1) <= 0 {
			s.ipTracker.Delete(ip)
		}
	}
}

// extractIP returns the IP portion of a net.Addr.
func extractIP(addr net.Addr) string {
	if addr == nil {
		return ""
	}
	host, _, err := net.SplitHostPort(addr.String())
	if err != nil {
		return addr.String()
	}
	return host
}
