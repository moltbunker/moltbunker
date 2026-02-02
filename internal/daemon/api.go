package daemon

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/moltbunker/moltbunker/internal/config"
	"github.com/moltbunker/moltbunker/internal/metrics"
	"github.com/moltbunker/moltbunker/internal/util"
)

// APIServer provides Unix socket API for CLI communication
type APIServer struct {
	node             *Node
	containerManager *ContainerManager
	socketPath       string
	dataDir          string
	config           *config.Config // Full configuration
	listener         net.Listener
	startTime        time.Time
	mu               sync.RWMutex
	running          bool
	metrics          *metrics.Collector
	globalLimiter    *rateLimiter

	// Rate limiting configuration (from config)
	rateLimitRequests   int           // Max requests per connection per window
	rateLimitWindow     time.Duration // Rate limit window duration
	maxRequestSize      int64         // Maximum request body size
}

// NewAPIServer creates a new API server with default rate limiting
func NewAPIServer(node *Node, socketPath string, dataDir string) *APIServer {
	return &APIServer{
		node:                node,
		socketPath:          socketPath,
		dataDir:             dataDir,
		metrics:             metrics.NewCollector(),
		globalLimiter:       newRateLimiter(globalRateLimitTokens, time.Duration(globalRateLimitRefillSec)*time.Second),
		rateLimitRequests:   rateLimitTokens,
		rateLimitWindow:     time.Duration(rateLimitRefillSec) * time.Second,
		maxRequestSize:      MaxRequestSize,
	}
}

// NewAPIServerWithFullConfig creates a new API server with full configuration
func NewAPIServerWithFullConfig(node *Node, cfg *config.Config) *APIServer {
	rateLimitReqs := cfg.API.RateLimitRequests
	rateLimitWindowSecs := cfg.API.RateLimitWindowSecs
	maxReqSize := cfg.API.MaxRequestSize

	// Use defaults if values are not set
	if rateLimitReqs <= 0 {
		rateLimitReqs = rateLimitTokens
	}
	if rateLimitWindowSecs <= 0 {
		rateLimitWindowSecs = rateLimitRefillSec
	}
	if maxReqSize <= 0 {
		maxReqSize = int(MaxRequestSize)
	}

	rateLimitWindow := time.Duration(rateLimitWindowSecs) * time.Second

	// Global rate limit is 10x the per-connection rate limit
	globalLimit := rateLimitReqs * 10

	return &APIServer{
		node:                node,
		socketPath:          cfg.Daemon.SocketPath,
		dataDir:             cfg.Daemon.DataDir,
		config:              cfg,
		metrics:             metrics.NewCollector(),
		globalLimiter:       newRateLimiter(globalLimit, rateLimitWindow),
		rateLimitRequests:   rateLimitReqs,
		rateLimitWindow:     rateLimitWindow,
		maxRequestSize:      int64(maxReqSize),
	}
}

// NewAPIServerWithConfig creates a new API server with configuration
func NewAPIServerWithConfig(node *Node, socketPath string, dataDir string, rateLimitReqs int, rateLimitWindowSecs int, maxReqSize int) *APIServer {
	// Use defaults if values are not set
	if rateLimitReqs <= 0 {
		rateLimitReqs = rateLimitTokens
	}
	if rateLimitWindowSecs <= 0 {
		rateLimitWindowSecs = rateLimitRefillSec
	}
	if maxReqSize <= 0 {
		maxReqSize = int(MaxRequestSize)
	}

	rateLimitWindow := time.Duration(rateLimitWindowSecs) * time.Second

	// Global rate limit is 10x the per-connection rate limit
	globalLimit := rateLimitReqs * 10

	return &APIServer{
		node:                node,
		socketPath:          socketPath,
		dataDir:             dataDir,
		metrics:             metrics.NewCollector(),
		globalLimiter:       newRateLimiter(globalLimit, rateLimitWindow),
		rateLimitRequests:   rateLimitReqs,
		rateLimitWindow:     rateLimitWindow,
		maxRequestSize:      int64(maxReqSize),
	}
}

// DefaultSocketPath returns the default socket path
func DefaultSocketPath() string {
	// Try XDG_RUNTIME_DIR first (secure, user-specific)
	if runtimeDir := os.Getenv("XDG_RUNTIME_DIR"); runtimeDir != "" {
		return filepath.Join(runtimeDir, "moltbunker", "daemon.sock")
	}

	// Fall back to home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Last resort: use /var/run with user ID for uniqueness
		return fmt.Sprintf("/var/run/user/%d/moltbunker/daemon.sock", os.Getuid())
	}
	return filepath.Join(homeDir, ".moltbunker", "daemon.sock")
}

// Start starts the API server
func (s *APIServer) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return fmt.Errorf("API server already running")
	}

	// Ensure directory exists with secure permissions (user-only access)
	dir := filepath.Dir(s.socketPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		s.mu.Unlock()
		return fmt.Errorf("failed to create socket directory: %w", err)
	}

	// Remove existing socket
	os.Remove(s.socketPath)

	// Set restrictive umask before creating socket to avoid TOCTOU race condition.
	// This ensures the socket is created with secure permissions from the start,
	// rather than creating it and then chmod'ing (which leaves a window of vulnerability).
	oldUmask := syscall.Umask(0077)

	// Create Unix socket listener
	listener, err := net.Listen("unix", s.socketPath)

	// Restore the original umask immediately after socket creation
	syscall.Umask(oldUmask)

	if err != nil {
		s.mu.Unlock()
		return fmt.Errorf("failed to create unix socket: %w", err)
	}

	// Initialize container manager
	containerdSocket := "/run/containerd/containerd.sock"
	if s.config != nil && s.config.Runtime.ContainerdSocket != "" {
		containerdSocket = s.config.Runtime.ContainerdSocket
	}
	cmConfig := ContainerManagerConfig{
		DataDir:          s.dataDir,
		ContainerdSocket: containerdSocket,
		TorDataDir:       filepath.Join(s.dataDir, "tor"),
		EnableEncryption: true,
	}
	containerManager, err := NewContainerManager(ctx, cmConfig, s.node)
	if err != nil {
		listener.Close()
		s.mu.Unlock()
		return fmt.Errorf("failed to create container manager: %w", err)
	}

	s.listener = listener
	s.containerManager = containerManager
	s.running = true
	s.startTime = time.Now()
	s.mu.Unlock()

	// Handle connections using SafeGo for panic recovery
	util.SafeGoWithName("api-connection-handler", func() {
		s.handleConnections(ctx)
	})

	return nil
}

// Stop stops the API server
func (s *APIServer) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	s.running = false

	// Close container manager
	if s.containerManager != nil {
		s.containerManager.Close()
	}

	if s.listener != nil {
		s.listener.Close()
	}

	// Clean up socket file
	os.Remove(s.socketPath)

	return nil
}

// SocketPath returns the socket path
func (s *APIServer) SocketPath() string {
	return s.socketPath
}
