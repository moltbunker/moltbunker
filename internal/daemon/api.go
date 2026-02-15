package daemon

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/moltbunker/moltbunker/internal/config"
	"github.com/moltbunker/moltbunker/internal/logging"
	"github.com/moltbunker/moltbunker/internal/metrics"
	"github.com/moltbunker/moltbunker/internal/runtime"
	"github.com/moltbunker/moltbunker/internal/util"
)

// APIServer provides Unix socket API for CLI communication
type APIServer struct {
	node             *Node
	containerManager *ContainerManager
	profileManager   *NodeProfileManager
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

	// Admin badge getter — set by external API server to merge badges into status
	adminBadgeGetter AdminBadgeGetter
}

// AdminBadgeGetter returns admin-assigned badges/blocked status for a node
type AdminBadgeGetter interface {
	Get(nodeID string) *AdminNodeMeta
}

// AdminNodeMeta is a minimal struct for badge merging (avoids import cycle)
type AdminNodeMeta struct {
	Badges  []string
	Blocked bool
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

// GetContainerManager returns the container manager for direct API access (e.g., exec relay)
func (s *APIServer) GetContainerManager() *ContainerManager {
	return s.containerManager
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

	// Set umask before creating socket to avoid TOCTOU race condition.
	// Use 0007 to allow group access (the API service runs as the moltbunker group
	// and needs to connect when the daemon runs as root).
	oldUmask := syscall.Umask(0007)

	// Create Unix socket listener
	listener, err := net.Listen("unix", s.socketPath)

	// Restore the original umask immediately after socket creation
	syscall.Umask(oldUmask)

	if err != nil {
		s.mu.Unlock()
		return fmt.Errorf("failed to create unix socket: %w", err)
	}

	// If running as root, set socket group to "moltbunker" so the API service can connect
	if os.Getuid() == 0 {
		if grp, err := user.LookupGroup("moltbunker"); err == nil {
			if gid, err := strconv.Atoi(grp.Gid); err == nil {
				os.Chown(s.socketPath, 0, gid)
			}
		}
	}

	// Initialize container manager
	containerdSocket := "/run/containerd/containerd.sock"
	if s.config != nil && s.config.Runtime.ContainerdSocket != "" {
		containerdSocket = s.config.Runtime.ContainerdSocket
	}
	runtimeName := "auto"
	if s.config != nil && s.config.Runtime.RuntimeName != "" {
		runtimeName = s.config.Runtime.RuntimeName
	}
	// Map config.Runtime.Kata → runtime.KataConfig for Kata hypervisor annotations
	var kataConfig *runtime.KataConfig
	if s.config != nil {
		kata := s.config.Runtime.Kata
		if kata.VMMemoryMB > 0 || kata.VMCPUs > 0 || kata.KernelPath != "" || kata.ImagePath != "" {
			kataConfig = &runtime.KataConfig{
				VMMemoryMB: kata.VMMemoryMB,
				VMCPUs:     kata.VMCPUs,
				KernelPath: kata.KernelPath,
				ImagePath:  kata.ImagePath,
			}
		}
	}
	cmConfig := ContainerManagerConfig{
		DataDir:          s.dataDir,
		ContainerdSocket: containerdSocket,
		RuntimeName:      runtimeName,
		KataConfig:       kataConfig,
		TorDataDir:       filepath.Join(s.dataDir, "tor"),
		EnableEncryption: true,
		PaymentService:   s.node.PaymentService(),
	}
	containerManager, err := NewContainerManager(ctx, cmConfig, s.node)
	if err != nil {
		listener.Close()
		s.mu.Unlock()
		return fmt.Errorf("failed to create container manager: %w", err)
	}

	s.listener = listener
	s.containerManager = containerManager

	// Initialize node profile manager for dashboard/status data
	if s.config != nil {
		nodesDir := filepath.Join(s.dataDir, "nodes")
		pm, err := NewNodeProfileManager(nodesDir, s.config, s.node, containerManager)
		if err != nil {
			logging.Warn("failed to create node profile manager", "error", err.Error())
		} else {
			s.profileManager = pm
		}
	}

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

	// Close container manager (stops health monitoring, containers, containerd)
	if s.containerManager != nil {
		logging.Info("closing container manager", logging.Component("api"))
		s.containerManager.Close()
	}

	// Close listener to stop accepting new connections
	if s.listener != nil {
		logging.Info("closing API socket listener", logging.Component("api"))
		s.listener.Close()
	}

	// Clean up socket file
	os.Remove(s.socketPath)

	return nil
}

// SetAdminBadgeGetter sets the admin badge getter for merging into status responses
func (s *APIServer) SetAdminBadgeGetter(getter AdminBadgeGetter) {
	s.adminBadgeGetter = getter
}

// SocketPath returns the socket path
func (s *APIServer) SocketPath() string {
	return s.socketPath
}
