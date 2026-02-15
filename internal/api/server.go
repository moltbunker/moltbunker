package api

import (
	"context"
	"crypto/subtle"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"net/http/pprof"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"github.com/moltbunker/moltbunker/internal/client"
	"github.com/moltbunker/moltbunker/internal/cloning"
	"github.com/moltbunker/moltbunker/internal/config"
	"github.com/moltbunker/moltbunker/internal/daemon"
	"github.com/moltbunker/moltbunker/internal/logging"
	"github.com/moltbunker/moltbunker/internal/metrics"
	"github.com/moltbunker/moltbunker/internal/snapshot"
	"github.com/moltbunker/moltbunker/internal/threat"
)

// Server is the external HTTP API server
type Server struct {
	config          *ServerConfig
	fullConfig      *config.Config // Full app config for fallback values
	httpServer      *http.Server
	tlsServer       *http.Server
	mu              sync.RWMutex
	running         bool

	// Core components
	daemonAPI       *daemon.APIServer
	threatDetector  *threat.Detector
	snapshotManager *snapshot.Manager
	cloningManager  *cloning.Manager

	// Daemon bridge for forwarding requests
	daemonBridge    *DaemonBridge

	// API key manager
	apiKeyManager   *APIKeyManager

	// Wallet authentication manager (permissionless)
	walletAuth      *WalletAuthManager

	// WebSocket hub
	wsHub           *WebSocketHub

	// Exec session manager
	execSessions    *ExecSessionManager

	// Metrics collector
	metricsCollector *metrics.Collector

	// Admin stores
	adminStore   *AdminMetadataStore
	policyStore  *PolicyStore
	catalogStore *CatalogStore

	// Per-IP rate limiters
	rateLimiters    sync.Map

	// Rate limiter cleanup control
	rateLimitCtx    context.Context
	rateLimitCancel context.CancelFunc
}

// rateLimiterEntry holds a rate limiter and the last time it was used
type rateLimiterEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// ServerConfig configures the HTTP API server
type ServerConfig struct {
	// HTTP settings
	HTTPAddr        string `yaml:"http_addr"`        // e.g., ":8080"
	HTTPSAddr       string `yaml:"https_addr"`       // e.g., ":8443"
	EnableHTTPS     bool   `yaml:"enable_https"`
	TLSCertFile     string `yaml:"tls_cert_file"`
	TLSKeyFile      string `yaml:"tls_key_file"`

	// Daemon connection
	DaemonSocketPath string `yaml:"daemon_socket_path"`
	DaemonPoolSize   int    `yaml:"daemon_pool_size"`

	// Rate limiting
	RateLimit       int    `yaml:"rate_limit"`        // Requests per minute
	RateLimitBurst  int    `yaml:"rate_limit_burst"`

	// Authentication
	EnableAuth      bool   `yaml:"enable_auth"`
	APIKeyHeader    string `yaml:"api_key_header"`    // Default: X-API-Key
	APIKeyStorePath string `yaml:"api_key_store_path"`

	// Proxy trust (only enable behind a trusted reverse proxy like Cloudflare)
	TrustProxy bool `yaml:"trust_proxy"`

	// CORS
	EnableCORS      bool     `yaml:"enable_cors"`
	AllowedOrigins  []string `yaml:"allowed_origins"`

	// Timeouts
	ReadTimeout       time.Duration `yaml:"read_timeout"`
	ReadHeaderTimeout time.Duration `yaml:"read_header_timeout"`
	WriteTimeout      time.Duration `yaml:"write_timeout"`
	IdleTimeout       time.Duration `yaml:"idle_timeout"`

	// WebSocket
	EnableWebSocket bool `yaml:"enable_websocket"`

	// Profiling
	EnablePprof bool `yaml:"enable_pprof"` // Enable /debug/pprof/ endpoints (disabled by default)
}

// DefaultServerConfig returns the default server configuration
func DefaultServerConfig() *ServerConfig {
	return &ServerConfig{
		HTTPAddr:         ":8080",
		HTTPSAddr:        ":8443",
		EnableHTTPS:      false,
		DaemonSocketPath: client.DefaultSocketPath(),
		DaemonPoolSize:   10,
		RateLimit:        100,
		RateLimitBurst:   20,
		EnableAuth:       true,
		APIKeyHeader:     "X-API-Key",
		APIKeyStorePath:  "",
		EnableCORS:       true,
		AllowedOrigins:   []string{},
		ReadTimeout:      30 * time.Second,
		WriteTimeout:     30 * time.Second,
		IdleTimeout:      120 * time.Second,
		EnableWebSocket:  true,
	}
}

// NewServer creates a new HTTP API server
func NewServer(cfg *ServerConfig) *Server {
	if cfg == nil {
		cfg = DefaultServerConfig()
	}

	s := &Server{
		config: cfg,
	}

	// Initialize daemon bridge
	if cfg.DaemonSocketPath != "" {
		s.daemonBridge = NewDaemonBridge(cfg.DaemonSocketPath, cfg.DaemonPoolSize)
	}

	// Initialize API key manager
	if cfg.EnableAuth && cfg.APIKeyStorePath != "" {
		var err error
		s.apiKeyManager, err = NewAPIKeyManager(cfg.APIKeyStorePath)
		if err != nil {
			logging.Warn("failed to initialize API key manager, using in-memory storage",
				"error", err.Error(),
				logging.Component("api"))
			s.apiKeyManager = NewAPIKeyManagerInMemory()
		}
	} else if cfg.EnableAuth {
		s.apiKeyManager = NewAPIKeyManagerInMemory()
	}

	// Initialize wallet authentication manager (permissionless)
	s.walletAuth = NewWalletAuthManager(5 * time.Minute)

	if cfg.EnableWebSocket {
		s.wsHub = NewWebSocketHub()
	}

	// Initialize exec session manager
	s.execSessions = NewExecSessionManager()

	return s
}

// SetFullConfig sets the full app config for fallback values
func (s *Server) SetFullConfig(cfg *config.Config) {
	s.fullConfig = cfg
}

// SetDaemonAPI sets the daemon API server for forwarding requests
func (s *Server) SetDaemonAPI(api *daemon.APIServer) {
	s.daemonAPI = api
}

// SetThreatDetector sets the threat detector
func (s *Server) SetThreatDetector(detector *threat.Detector) {
	s.threatDetector = detector
}

// SetSnapshotManager sets the snapshot manager
func (s *Server) SetSnapshotManager(manager *snapshot.Manager) {
	s.snapshotManager = manager
}

// SetCloningManager sets the cloning manager
func (s *Server) SetCloningManager(manager *cloning.Manager) {
	s.cloningManager = manager
}

// SetAdminStore sets the admin metadata store
func (s *Server) SetAdminStore(store *AdminMetadataStore) {
	s.adminStore = store
}

// SetPolicyStore sets the policy store
func (s *Server) SetPolicyStore(store *PolicyStore) {
	s.policyStore = store
}

// SetCatalogStore sets the deploy catalog store
func (s *Server) SetCatalogStore(store *CatalogStore) {
	s.catalogStore = store
}

// GetAdminStore returns the admin metadata store (for daemon badge merging)
func (s *Server) GetAdminStore() *AdminMetadataStore {
	return s.adminStore
}

// Start starts the HTTP API server
func (s *Server) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return fmt.Errorf("server already running")
	}
	s.running = true
	s.mu.Unlock()

	// Start rate limiter cleanup goroutine
	if s.config.RateLimit > 0 {
		s.rateLimitCtx, s.rateLimitCancel = context.WithCancel(ctx)
		s.startRateLimiterCleanup()
	}

	// Build router
	router := s.buildRouter()

	// Start WebSocket hub if enabled
	if s.wsHub != nil {
		go s.wsHub.Run(ctx)
	}

	// Start HTTP server.
	// Use ReadHeaderTimeout (not ReadTimeout) so header parsing is bounded
	// but long-lived WebSocket connections aren't killed after the timeout.
	// WriteTimeout is 0 — each handler manages its own write deadlines.
	// The exec WebSocket handler clears deadlines after upgrade and sets
	// per-message deadlines; REST handlers are fast enough to not need one.
	readHdrTimeout := s.config.ReadHeaderTimeout
	if readHdrTimeout == 0 {
		readHdrTimeout = s.config.ReadTimeout // fallback to ReadTimeout if not set
	}
	s.httpServer = &http.Server{
		Addr:              s.config.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: readHdrTimeout,
		IdleTimeout:       s.config.IdleTimeout,
	}

	go func() {
		logging.Info("HTTP API server starting",
			"addr", s.config.HTTPAddr,
			logging.Component("api"))

		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logging.Error("HTTP server error",
				"error", err.Error(),
				logging.Component("api"))
		}
	}()

	// Start HTTPS server if enabled
	if s.config.EnableHTTPS && s.config.TLSCertFile != "" && s.config.TLSKeyFile != "" {
		tlsConfig := &tls.Config{
			MinVersion: tls.VersionTLS13,
			CipherSuites: []uint16{
				tls.TLS_AES_256_GCM_SHA384,
				tls.TLS_CHACHA20_POLY1305_SHA256,
				tls.TLS_AES_128_GCM_SHA256,
			},
		}

		s.tlsServer = &http.Server{
			Addr:              s.config.HTTPSAddr,
			Handler:           router,
			TLSConfig:         tlsConfig,
			ReadHeaderTimeout: readHdrTimeout,
			IdleTimeout:       s.config.IdleTimeout,
		}

		go func() {
			logging.Info("HTTPS API server starting",
				"addr", s.config.HTTPSAddr,
				logging.Component("api"))

			if err := s.tlsServer.ListenAndServeTLS(s.config.TLSCertFile, s.config.TLSKeyFile); err != nil && err != http.ErrServerClosed {
				logging.Error("HTTPS server error",
					"error", err.Error(),
					logging.Component("api"))
			}
		}()
	}

	return nil
}

// Stop stops the HTTP API server
func (s *Server) Stop(ctx context.Context) error {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return nil
	}
	s.running = false
	s.mu.Unlock()

	var errs []error

	// Shutdown HTTP server
	if s.httpServer != nil {
		if err := s.httpServer.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("HTTP server shutdown: %w", err))
		}
	}

	// Shutdown HTTPS server
	if s.tlsServer != nil {
		if err := s.tlsServer.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("HTTPS server shutdown: %w", err))
		}
	}

	// Stop rate limiter cleanup
	if s.rateLimitCancel != nil {
		s.rateLimitCancel()
	}

	// Close daemon bridge
	if s.daemonBridge != nil {
		s.daemonBridge.Close()
	}

	// Close exec session manager
	if s.execSessions != nil {
		s.execSessions.Close()
	}

	logging.Info("API server stopped", logging.Component("api"))

	if len(errs) > 0 {
		return fmt.Errorf("shutdown errors: %v", errs)
	}
	return nil
}

// SetDaemonBridge sets the daemon bridge for forwarding requests
func (s *Server) SetDaemonBridge(bridge *DaemonBridge) {
	s.daemonBridge = bridge
}

// GetDaemonBridge returns the daemon bridge
func (s *Server) GetDaemonBridge() *DaemonBridge {
	return s.daemonBridge
}

// GetAPIKeyManager returns the API key manager
func (s *Server) GetAPIKeyManager() *APIKeyManager {
	return s.apiKeyManager
}

// buildRouter builds the HTTP router with all handlers
func (s *Server) buildRouter() http.Handler {
	mux := http.NewServeMux()

	// Authentication endpoints (no auth required)
	mux.HandleFunc("/v1/auth/challenge", s.withCORS(s.handleAuthChallenge))
	mux.HandleFunc("/v1/auth/verify", s.withCORS(s.handleAuthVerify))

	// API v1 routes (auth required)
	mux.HandleFunc("/v1/status", s.withMiddleware(s.handleStatus))
	mux.HandleFunc("/v1/deploy", s.withMiddleware(s.handleDeploy))
	mux.HandleFunc("/v1/reserve", s.withMiddleware(s.handleReserve))
	mux.HandleFunc("/v1/clone", s.withMiddleware(s.handleClone))
	mux.HandleFunc("/v1/migrate", s.withMiddleware(s.handleMigrate))
	mux.HandleFunc("/v1/balance", s.withMiddleware(s.handleBalance))
	mux.HandleFunc("/v1/snapshot", s.withMiddleware(s.handleSnapshot))
	mux.HandleFunc("/v1/restore", s.withMiddleware(s.handleRestore))
	mux.HandleFunc("/v1/threat", s.withMiddleware(s.handleThreat))

	// Bot management
	mux.HandleFunc("/v1/bots", s.withMiddleware(s.handleBots))
	mux.HandleFunc("/v1/bots/", s.withMiddleware(s.handleBotByID))

	// Runtime management
	mux.HandleFunc("/v1/runtimes/reserve", s.withMiddleware(s.handleRuntimeReserve))
	mux.HandleFunc("/v1/runtimes/", s.withMiddleware(s.handleRuntimeByID))

	// Deployment management
	mux.HandleFunc("/v1/deployments", s.withMiddleware(s.handleDeployments))
	mux.HandleFunc("/v1/deployments/", s.withMiddleware(s.handleDeploymentByID))

	// Snapshot management
	mux.HandleFunc("/v1/snapshots", s.withMiddleware(s.handleSnapshots))
	mux.HandleFunc("/v1/snapshots/", s.withMiddleware(s.handleSnapshotByID))

	// Clone management
	mux.HandleFunc("/v1/clones", s.withMiddleware(s.handleClones))
	mux.HandleFunc("/v1/clones/", s.withMiddleware(s.handleCloneByID))

	// Container management
	mux.HandleFunc("/v1/containers", s.withMiddleware(s.handleContainers))
	mux.HandleFunc("/v1/containers/", s.withMiddleware(s.handleContainerByID))

	// Exec terminal endpoints
	mux.HandleFunc("/v1/exec/challenge", s.withMiddleware(s.handleExecChallenge))
	mux.HandleFunc("/v1/exec/ws", s.handleExecWebSocket) // Auth via challenge/signature, not middleware

	// Admin endpoints (admin auth required)
	mux.HandleFunc("/v1/admin/me", s.withAdminMiddleware(s.handleAdminMe))
	mux.HandleFunc("/v1/admin/nodes", s.withAdminMiddleware(s.handleAdminNodes))
	mux.HandleFunc("/v1/admin/nodes/", s.withAdminMiddleware(s.handleAdminNodeByID))
	mux.HandleFunc("/v1/admin/policies", s.withAdminMiddleware(s.handleAdminPolicies))

	// API key management (admin) — dual paths for frontend compatibility
	mux.HandleFunc("/v1/api-keys", s.withAdminMiddleware(s.handleGatewayKeys))
	mux.HandleFunc("/v1/api-keys/", s.withAdminMiddleware(s.handleGatewayKeyByID))
	mux.HandleFunc("/v1/admin/gateway/keys", s.withAdminMiddleware(s.handleGatewayKeys))
	mux.HandleFunc("/v1/admin/gateway/keys/", s.withAdminMiddleware(s.handleGatewayKeyByID))
	mux.HandleFunc("/v1/admin/gateway/metrics", s.withAdminMiddleware(s.handleGatewayMetrics))

	// Bootstrap peer discovery (public, no auth — needed before nodes have API keys)
	mux.HandleFunc("/v1/bootstrap", s.withCORS(s.handleBootstrap))

	// Deploy catalog (public, no auth)
	mux.HandleFunc("/v1/catalog", s.withCORS(s.handleCatalog))

	// Deploy catalog management (admin)
	mux.HandleFunc("/v1/admin/catalog", s.withAdminMiddleware(s.handleAdminCatalog))
	mux.HandleFunc("/v1/admin/catalog/presets", s.withAdminMiddleware(s.handleAdminCatalogPresets))
	mux.HandleFunc("/v1/admin/catalog/presets/", s.withAdminMiddleware(s.handleAdminCatalogPresetByID))
	mux.HandleFunc("/v1/admin/catalog/categories", s.withAdminMiddleware(s.handleAdminCatalogCategories))
	mux.HandleFunc("/v1/admin/catalog/categories/", s.withAdminMiddleware(s.handleAdminCatalogCategoryByID))
	mux.HandleFunc("/v1/admin/catalog/tiers/", s.withAdminMiddleware(s.handleAdminCatalogTierByID))

	// Health endpoints (no auth required)
	mux.HandleFunc("/v1/health", s.handleHealth)
	mux.HandleFunc("/v1/healthz", s.handleHealthz)
	mux.HandleFunc("/v1/readyz", s.handleReadyz)

	// Metrics endpoint (admin auth required — use Prometheus bearer_token config)
	mux.HandleFunc("/v1/metrics", s.withAdminMiddleware(s.handleMetrics))

	// Top-level health check for load balancer probes (no auth required)
	mux.HandleFunc("/health", s.handleHealthCheck)

	// WebSocket endpoint (auth required)
	if s.config.EnableWebSocket && s.wsHub != nil {
		mux.HandleFunc("/v1/ws", s.withMiddleware(s.handleWebSocket))
	}

	// pprof debug endpoints (admin-only — exposes heap dumps, goroutine stacks)
	if s.config.EnablePprof {
		mux.HandleFunc("/debug/pprof/", s.withAdminMiddleware(pprof.Index))
		mux.HandleFunc("/debug/pprof/cmdline", s.withAdminMiddleware(pprof.Cmdline))
		mux.HandleFunc("/debug/pprof/profile", s.withAdminMiddleware(pprof.Profile))
		mux.HandleFunc("/debug/pprof/symbol", s.withAdminMiddleware(pprof.Symbol))
		mux.HandleFunc("/debug/pprof/trace", s.withAdminMiddleware(pprof.Trace))
		logging.Info("pprof debug endpoints enabled (admin-only)", logging.Component("api"))
	}

	// Wrap the entire mux with global CORS so even unmatched routes get CORS headers.
	// Without this, preflight OPTIONS to unknown paths returns 404 without CORS,
	// which browsers treat as a CORS error (masking the real 404).
	if s.config.EnableCORS {
		return s.globalCORSMiddleware(mux)
	}
	return mux
}

// globalCORSMiddleware wraps an entire handler tree with CORS headers.
// This ensures preflight OPTIONS and error responses always include CORS.
func (s *Server) globalCORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.setCORSHeaders(w, r)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// withCORS wraps a handler with CORS support only (no auth)
func (s *Server) withCORS(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.config.EnableCORS {
			s.setCORSHeaders(w, r)
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}
		}
		handler(w, r)
	}
}

// withMiddleware wraps a handler with common middleware
func (s *Server) withMiddleware(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// CORS
		if s.config.EnableCORS {
			s.setCORSHeaders(w, r)
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}
		}

		// Authentication
		if s.config.EnableAuth {
			if !s.authenticate(r) {
				http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
				return
			}
		}

		// Rate limiting
		if s.config.RateLimit > 0 {
			ip := s.extractClientIP(r)
			limiter := s.getRateLimiter(ip)
			if !limiter.Allow() {
				logging.Warn("rate limit exceeded",
					"ip", ip,
					"path", r.URL.Path,
					"method", r.Method,
					logging.Component("api"))
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Retry-After", "60")
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"error": "rate limit exceeded", "retry_after": 60}`))
				return
			}
		}

		// Call the actual handler
		handler(w, r)
	}
}

// getRateLimiter returns the rate limiter for the given IP address.
// It creates a new limiter if one does not already exist.
func (s *Server) getRateLimiter(ip string) *rate.Limiter {
	now := time.Now()

	if val, ok := s.rateLimiters.Load(ip); ok {
		entry := val.(*rateLimiterEntry)
		entry.lastSeen = now
		return entry.limiter
	}

	// Convert requests per minute to requests per second
	rps := rate.Limit(float64(s.config.RateLimit) / 60.0)
	limiter := rate.NewLimiter(rps, s.config.RateLimitBurst)

	entry := &rateLimiterEntry{
		limiter:  limiter,
		lastSeen: now,
	}
	actual, _ := s.rateLimiters.LoadOrStore(ip, entry)
	return actual.(*rateLimiterEntry).limiter
}

// extractClientIP extracts the client IP address from the request.
// Only trusts proxy headers (X-Forwarded-For, CF-Connecting-IP, X-Real-IP)
// when TrustProxy is enabled in the server config.
func (s *Server) extractClientIP(r *http.Request) string {
	if s.config.TrustProxy {
		// Cloudflare sets CF-Connecting-IP (most reliable behind CF Tunnel)
		if cfIP := r.Header.Get("CF-Connecting-IP"); cfIP != "" {
			return strings.TrimSpace(cfIP)
		}

		// X-Forwarded-For: use the first (leftmost) IP
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			if idx := strings.IndexByte(xff, ','); idx != -1 {
				return strings.TrimSpace(xff[:idx])
			}
			return strings.TrimSpace(xff)
		}

		// X-Real-IP
		if xri := r.Header.Get("X-Real-IP"); xri != "" {
			return strings.TrimSpace(xri)
		}
	}

	// Default: use TCP remote address (not spoofable)
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

// startRateLimiterCleanup starts a goroutine that periodically removes stale rate limiters
func (s *Server) startRateLimiterCleanup() {
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-s.rateLimitCtx.Done():
				return
			case <-ticker.C:
				s.cleanupRateLimiters()
			}
		}
	}()
}

// cleanupRateLimiters removes rate limiter entries that have not been seen recently
func (s *Server) cleanupRateLimiters() {
	staleThreshold := time.Now().Add(-10 * time.Minute)
	var cleaned int

	s.rateLimiters.Range(func(key, value any) bool {
		entry := value.(*rateLimiterEntry)
		if entry.lastSeen.Before(staleThreshold) {
			s.rateLimiters.Delete(key)
			cleaned++
		}
		return true
	})

	if cleaned > 0 {
		logging.Debug("cleaned up stale rate limiters",
			"count", cleaned,
			logging.Component("api"))
	}
}

// setCORSHeaders sets CORS headers on the response
func (s *Server) setCORSHeaders(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	if origin == "" {
		origin = "*"
	}

	// Check if origin is allowed
	allowed := false
	for _, o := range s.config.AllowedOrigins {
		if o == "*" || o == origin {
			allowed = true
			break
		}
	}

	if allowed {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key")
		w.Header().Set("Access-Control-Max-Age", "86400")
	}
}

// ContextKey type for context values
type ContextKey string

const (
	// CtxAPIKeyKey is the context key for API key
	CtxAPIKeyKey ContextKey = "api_key"
	// CtxWalletKey is the context key for wallet address
	CtxWalletKey ContextKey = "wallet_address"
	// CtxAuthTypeKey is the context key for auth type
	CtxAuthTypeKey ContextKey = "auth_type"
)

// authenticate checks if the request is authenticated and returns the auth context
func (s *Server) authenticate(r *http.Request) bool {
	// Check API key header
	apiKey := r.Header.Get(s.config.APIKeyHeader)
	if apiKey != "" {
		if s.apiKeyManager != nil {
			return s.apiKeyManager.ValidateKey(apiKey)
		}
		return false
	}

	// Check Authorization header for Bearer token
	auth := r.Header.Get("Authorization")
	if auth != "" && len(auth) > 7 && auth[:7] == "Bearer " {
		token := auth[7:]

		// Check if it's an API key (mb_ prefix)
		if len(token) > 3 && token[:3] == "mb_" {
			if s.apiKeyManager != nil {
				return s.apiKeyManager.ValidateKey(token)
			}
			return false
		}

		// Check if it's a wallet session token (wt_ prefix)
		if len(token) > 3 && token[:3] == "wt_" {
			if s.walletAuth != nil {
				_, valid := s.walletAuth.ValidateSession(token)
				return valid
			}
		}
	}

	// Check inline wallet signature headers (permissionless)
	walletAddr := r.Header.Get("X-Wallet-Address")
	walletSig := r.Header.Get("X-Wallet-Signature")
	walletMsg := r.Header.Get("X-Wallet-Message")

	if walletAddr != "" && walletSig != "" && walletMsg != "" {
		if s.walletAuth != nil {
			_, err := s.walletAuth.VerifyInlineAuth(walletAddr, walletSig, walletMsg)
			if err != nil {
				logging.Debug("wallet inline auth failed",
					"address", walletAddr,
					"error", err.Error(),
					logging.Component("api"))
				return false
			}
			return true
		}
	}

	return false
}

// withAdminMiddleware wraps a handler with CORS, rate limiting, auth, and admin verification.
// Rate limiting is applied BEFORE authentication to prevent auth-based DoS attacks.
func (s *Server) withAdminMiddleware(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// CORS
		if s.config.EnableCORS {
			s.setCORSHeaders(w, r)
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}
		}

		// Rate limiting (before auth to prevent auth DoS)
		if s.config.RateLimit > 0 {
			ip := s.extractClientIP(r)
			limiter := s.getRateLimiter(ip)
			if !limiter.Allow() {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Retry-After", "60")
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"error": "rate limit exceeded", "retry_after": 60}`))
				return
			}
		}

		// Authentication (standard auth first)
		if s.config.EnableAuth {
			if !s.authenticate(r) {
				http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
				return
			}
		}

		// Admin authorization
		if _, ok := s.authenticateAdmin(r); !ok {
			http.Error(w, `{"error": "forbidden: admin access required"}`, http.StatusForbidden)
			return
		}

		handler(w, r)
	}
}

// authenticateAdmin extracts the wallet address and checks if it's an admin.
// Returns (wallet, true) if the caller is an admin.
func (s *Server) authenticateAdmin(r *http.Request) (string, bool) {
	// Check wallet session token
	auth := r.Header.Get("Authorization")
	if auth != "" && len(auth) > 7 && auth[:7] == "Bearer " {
		token := auth[7:]

		// Wallet session token
		if len(token) > 3 && token[:3] == "wt_" {
			if s.walletAuth != nil {
				addr, valid := s.walletAuth.ValidateSession(token)
				if valid && s.isAdminWallet(addr) {
					return addr, true
				}
			}
		}

		// API key with admin permission
		if len(token) > 3 && token[:3] == "mb_" {
			if s.apiKeyManager != nil && s.apiKeyManager.ValidateKeyWithPermission(token, "admin") {
				return "api_key", true
			}
		}
	}

	// Check API key header
	apiKey := r.Header.Get(s.config.APIKeyHeader)
	if apiKey != "" && s.apiKeyManager != nil {
		if s.apiKeyManager.ValidateKeyWithPermission(apiKey, "admin") {
			return "api_key", true
		}
	}

	// Check inline wallet headers
	walletAddr := r.Header.Get("X-Wallet-Address")
	walletSig := r.Header.Get("X-Wallet-Signature")
	walletMsg := r.Header.Get("X-Wallet-Message")
	if walletAddr != "" && walletSig != "" && walletMsg != "" {
		if s.walletAuth != nil {
			addr, err := s.walletAuth.VerifyInlineAuth(walletAddr, walletSig, walletMsg)
			if err == nil && s.isAdminWallet(addr) {
				return addr, true
			}
		}
	}

	return "", false
}

// isAdminWallet checks if a wallet address is in the admin list.
// Uses constant-time comparison to prevent timing side-channel attacks
// that could reveal which addresses are admin wallets.
func (s *Server) isAdminWallet(address string) bool {
	if s.fullConfig == nil {
		return false
	}
	// Normalize to lowercase hex bytes for constant-time comparison
	addrBytes, err := normalizeWalletForCompare(address)
	if err != nil {
		return false
	}
	found := false
	for _, admin := range s.fullConfig.API.AdminWallets {
		adminBytes, err := normalizeWalletForCompare(admin)
		if err != nil {
			continue
		}
		if subtle.ConstantTimeCompare(addrBytes, adminBytes) == 1 {
			found = true
		}
		// Always iterate all entries to avoid leaking admin list length
	}
	return found
}

// normalizeWalletForCompare converts a hex wallet address to canonical lowercase bytes.
// Returns the raw bytes of the lowercase hex representation for constant-time comparison.
func normalizeWalletForCompare(addr string) ([]byte, error) {
	// Strip 0x prefix if present
	cleaned := strings.TrimPrefix(strings.ToLower(addr), "0x")
	// Validate hex
	if _, err := hex.DecodeString(cleaned); err != nil {
		return nil, err
	}
	// Pad to 40 chars (20 bytes = Ethereum address) for consistent length
	for len(cleaned) < 40 {
		cleaned = "0" + cleaned
	}
	return []byte(cleaned), nil
}

// GetWalletAuthManager returns the wallet auth manager
func (s *Server) GetWalletAuthManager() *WalletAuthManager {
	return s.walletAuth
}

// GetWebSocketHub returns the WebSocket hub for external use
func (s *Server) GetWebSocketHub() *WebSocketHub {
	return s.wsHub
}

// BroadcastEvent broadcasts an event to all WebSocket clients
func (s *Server) BroadcastEvent(eventType string, data interface{}) {
	if s.wsHub != nil {
		s.wsHub.Broadcast(eventType, data)
	}
}

// LoadConfigFromDaemon loads configuration from daemon config
func LoadConfigFromDaemon(cfg *config.Config) *ServerConfig {
	serverCfg := DefaultServerConfig()

	// Set daemon socket path from config
	serverCfg.DaemonSocketPath = cfg.Daemon.SocketPath

	// Set API key store path
	serverCfg.APIKeyStorePath = filepath.Join(cfg.Daemon.DataDir, "api_keys.json")

	return serverCfg
}
