package daemon

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"regexp"
	goruntime "runtime"
	"sync"
	"syscall"
	"time"

	"github.com/moltbunker/moltbunker/internal/metrics"
	"github.com/moltbunker/moltbunker/internal/util"
	"github.com/moltbunker/moltbunker/pkg/types"
)

// Security constants
const (
	// MaxRequestSize is the maximum allowed request size (1MB)
	MaxRequestSize = 1 * 1024 * 1024

	// TODO: Make rate limit values configurable via config file
	// Per-connection rate limiting: 100 requests per minute
	rateLimitTokens    = 100
	rateLimitRefillSec = 60

	// TODO: Make global rate limit values configurable via config file
	// Global rate limiting: 1000 requests per minute total
	globalRateLimitTokens    = 1000
	globalRateLimitRefillSec = 60

	// Resource limits bounds
	maxMemoryLimit = 64 * 1024 * 1024 * 1024  // 64GB max
	maxDiskLimit   = 1024 * 1024 * 1024 * 1024 // 1TB max
	maxCPUQuota    = 10000000                  // 10 seconds in microseconds
	maxPIDLimit    = 10000
	maxNetworkBW   = 10 * 1024 * 1024 * 1024   // 10GB/s max

	// MaxLogTailLines is the maximum number of lines that can be requested via Tail parameter
	MaxLogTailLines = 10000
)

// Validation errors
var (
	ErrInvalidImageName     = errors.New("invalid container image name")
	ErrInvalidContainerID   = errors.New("invalid container ID format")
	ErrInvalidResourceLimit = errors.New("invalid resource limit")
	ErrInvalidTailValue     = errors.New("invalid tail value")
	ErrRateLimited          = errors.New("rate limit exceeded")
	ErrRequestTooLarge      = errors.New("request exceeds maximum size")
)

// Regex patterns for validation
var (
	// containerImageRegex allows alphanumeric, colons, slashes, dots, hyphens, and underscores
	// Examples: nginx:latest, docker.io/library/nginx:1.21, my-registry.com/my-image:v1.0.0
	containerImageRegex = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._\-/:]*[a-zA-Z0-9]$|^[a-zA-Z0-9]$`)

	// containerIDRegex allows hex characters (SHA256 prefix) or alphanumeric with hyphens
	containerIDRegex = regexp.MustCompile(`^[a-f0-9]{8,64}$|^[a-zA-Z0-9][a-zA-Z0-9\-]{0,62}[a-zA-Z0-9]$|^[a-zA-Z0-9]$`)
)

// rateLimiter implements a simple token bucket rate limiter
type rateLimiter struct {
	mu           sync.Mutex
	tokens       int
	maxTokens    int
	lastRefill   time.Time
	refillPeriod time.Duration
}

// newRateLimiter creates a new rate limiter
func newRateLimiter(maxTokens int, refillPeriod time.Duration) *rateLimiter {
	return &rateLimiter{
		tokens:       maxTokens,
		maxTokens:    maxTokens,
		lastRefill:   time.Now(),
		refillPeriod: refillPeriod,
	}
}

// allow checks if a request is allowed and consumes a token if so
func (r *rateLimiter) allow() bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Refill tokens based on elapsed time
	now := time.Now()
	elapsed := now.Sub(r.lastRefill)
	if elapsed >= r.refillPeriod {
		// Full refill after the period
		r.tokens = r.maxTokens
		r.lastRefill = now
	}

	if r.tokens > 0 {
		r.tokens--
		return true
	}
	return false
}

// APIServer provides Unix socket API for CLI communication
type APIServer struct {
	node             *Node
	containerManager *ContainerManager
	socketPath       string
	dataDir          string
	listener         net.Listener
	startTime        time.Time
	mu               sync.RWMutex
	running          bool
	metrics          *metrics.Collector
	globalLimiter    *rateLimiter
}

// APIRequest represents a JSON-RPC style request
type APIRequest struct {
	Method string          `json:"method"`
	Params json.RawMessage `json:"params,omitempty"`
	ID     int             `json:"id"`
}

// APIResponse represents a JSON-RPC style response
type APIResponse struct {
	Result interface{} `json:"result,omitempty"`
	Error  string      `json:"error,omitempty"`
	ID     int         `json:"id"`
}

// StatusResponse contains node status information
type StatusResponse struct {
	NodeID       string    `json:"node_id"`
	Running      bool      `json:"running"`
	Port         int       `json:"port"`
	PeerCount    int       `json:"peer_count"`
	Uptime       string    `json:"uptime"`
	Version      string    `json:"version"`
	TorEnabled   bool      `json:"tor_enabled"`
	TorAddress   string    `json:"tor_address,omitempty"`
	Containers   int       `json:"containers"`
	Region       string    `json:"region"`
}

// DeployRequest contains deployment parameters
type DeployRequest struct {
	Image           string               `json:"image"`
	Resources       types.ResourceLimits `json:"resources,omitempty"`
	TorOnly         bool                 `json:"tor_only"`
	OnionService    bool                 `json:"onion_service"`
	OnionPort       int                  `json:"onion_port,omitempty"`       // Port to expose via Tor (default: 80)
	WaitForReplicas bool                 `json:"wait_for_replicas,omitempty"` // If true, wait for at least 1 replica ack before returning
}

// DeployResponse contains deployment result
type DeployResponse struct {
	ContainerID     string   `json:"container_id"`
	OnionAddress    string   `json:"onion_address,omitempty"`
	Status          string   `json:"status"`
	EncryptedVolume string   `json:"encrypted_volume,omitempty"`
	Regions         []string `json:"regions"`
	ReplicaCount    int      `json:"replica_count"` // Number of successful replica acks received
}

// LogsRequest contains log streaming parameters
type LogsRequest struct {
	ContainerID string `json:"container_id"`
	Follow      bool   `json:"follow"`
	Tail        int    `json:"tail"`
}

// TorStatusResponse contains Tor status information
type TorStatusResponse struct {
	Running      bool      `json:"running"`
	OnionAddress string    `json:"onion_address,omitempty"`
	StartedAt    time.Time `json:"started_at,omitempty"`
	CircuitCount int       `json:"circuit_count"`
}

// ContainerInfo contains container information
type ContainerInfo struct {
	ID              string    `json:"id"`
	Image           string    `json:"image"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
	StartedAt       time.Time `json:"started_at,omitempty"`
	Encrypted       bool      `json:"encrypted"`
	OnionAddress    string    `json:"onion_address,omitempty"`
	Regions         []string  `json:"regions"`
}

// validateContainerImage validates a container image name
func validateContainerImage(image string) error {
	if image == "" {
		return fmt.Errorf("%w: image name cannot be empty", ErrInvalidImageName)
	}
	if len(image) > 256 {
		return fmt.Errorf("%w: image name exceeds maximum length of 256 characters", ErrInvalidImageName)
	}
	if !containerImageRegex.MatchString(image) {
		return fmt.Errorf("%w: image name contains invalid characters (allowed: alphanumeric, colons, slashes, dots, hyphens, underscores)", ErrInvalidImageName)
	}
	return nil
}

// validateContainerID validates a container ID format
func validateContainerID(id string) error {
	if id == "" {
		return fmt.Errorf("%w: container ID cannot be empty", ErrInvalidContainerID)
	}
	if len(id) > 64 {
		return fmt.Errorf("%w: container ID exceeds maximum length of 64 characters", ErrInvalidContainerID)
	}
	if !containerIDRegex.MatchString(id) {
		return fmt.Errorf("%w: container ID must be hex characters or alphanumeric with hyphens", ErrInvalidContainerID)
	}
	return nil
}

// validateResourceLimits validates resource limits are positive and within reasonable bounds
func validateResourceLimits(r types.ResourceLimits) error {
	// CPU quota validation (if set, must be positive and within bounds)
	if r.CPUQuota < 0 {
		return fmt.Errorf("%w: CPU quota cannot be negative", ErrInvalidResourceLimit)
	}
	if r.CPUQuota > maxCPUQuota {
		return fmt.Errorf("%w: CPU quota exceeds maximum of %d", ErrInvalidResourceLimit, maxCPUQuota)
	}

	// Memory limit validation
	if r.MemoryLimit < 0 {
		return fmt.Errorf("%w: memory limit cannot be negative", ErrInvalidResourceLimit)
	}
	if r.MemoryLimit > maxMemoryLimit {
		return fmt.Errorf("%w: memory limit exceeds maximum of %d bytes", ErrInvalidResourceLimit, maxMemoryLimit)
	}

	// Disk limit validation
	if r.DiskLimit < 0 {
		return fmt.Errorf("%w: disk limit cannot be negative", ErrInvalidResourceLimit)
	}
	if r.DiskLimit > maxDiskLimit {
		return fmt.Errorf("%w: disk limit exceeds maximum of %d bytes", ErrInvalidResourceLimit, maxDiskLimit)
	}

	// Network bandwidth validation
	if r.NetworkBW < 0 {
		return fmt.Errorf("%w: network bandwidth cannot be negative", ErrInvalidResourceLimit)
	}
	if r.NetworkBW > maxNetworkBW {
		return fmt.Errorf("%w: network bandwidth exceeds maximum of %d bytes/sec", ErrInvalidResourceLimit, maxNetworkBW)
	}

	// PID limit validation
	if r.PIDLimit < 0 {
		return fmt.Errorf("%w: PID limit cannot be negative", ErrInvalidResourceLimit)
	}
	if r.PIDLimit > maxPIDLimit {
		return fmt.Errorf("%w: PID limit exceeds maximum of %d", ErrInvalidResourceLimit, maxPIDLimit)
	}

	return nil
}

// validateDeployRequest validates a deployment request
func validateDeployRequest(req *DeployRequest) error {
	// Validate image name
	if err := validateContainerImage(req.Image); err != nil {
		return err
	}

	// Validate resource limits
	if err := validateResourceLimits(req.Resources); err != nil {
		return err
	}

	// Validate onion port if specified
	if req.OnionPort < 0 || req.OnionPort > 65535 {
		return fmt.Errorf("invalid onion port: must be between 0 and 65535")
	}

	return nil
}

// HealthzResponse contains detailed health check information
type HealthzResponse struct {
	Status              string  `json:"status"` // "healthy", "degraded", or "unhealthy"
	NodeRunning         bool    `json:"node_running"`
	ContainerdConnected bool    `json:"containerd_connected"`
	PeerCount           int     `json:"peer_count"`
	GoroutineCount      int     `json:"goroutine_count"`
	MemoryUsageMB       float64 `json:"memory_usage_mb"`
	MemoryAllocMB       float64 `json:"memory_alloc_mb"`
	Timestamp           time.Time `json:"timestamp"`
}

// ReadyzResponse contains readiness probe information
type ReadyzResponse struct {
	Ready      bool   `json:"ready"`
	Message    string `json:"message,omitempty"`
	Timestamp  time.Time `json:"timestamp"`
}

// NewAPIServer creates a new API server
func NewAPIServer(node *Node, socketPath string, dataDir string) *APIServer {
	return &APIServer{
		node:          node,
		socketPath:    socketPath,
		dataDir:       dataDir,
		metrics:       metrics.NewCollector(),
		globalLimiter: newRateLimiter(globalRateLimitTokens, time.Duration(globalRateLimitRefillSec)*time.Second),
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
	cmConfig := ContainerManagerConfig{
		DataDir:          s.dataDir,
		ContainerdSocket: "/run/containerd/containerd.sock",
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

// handleConnections accepts and handles incoming connections
func (s *APIServer) handleConnections(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			conn, err := s.listener.Accept()
			if err != nil {
				s.mu.RLock()
				running := s.running
				s.mu.RUnlock()
				if !running {
					return
				}
				continue
			}

			util.SafeGoWithName("api-handle-connection", func() {
				s.handleConnection(ctx, conn)
			})
		}
	}
}

// handleConnection handles a single API connection
func (s *APIServer) handleConnection(ctx context.Context, conn net.Conn) {
	defer conn.Close()

	// Track active connections
	s.metrics.IncrementConnections()
	defer s.metrics.DecrementConnections()

	// Create a rate limiter for this connection (100 requests per minute)
	limiter := newRateLimiter(rateLimitTokens, time.Duration(rateLimitRefillSec)*time.Second)

	// Create a limited reader to enforce max request size
	limitedReader := io.LimitReader(conn, MaxRequestSize)
	decoder := json.NewDecoder(limitedReader)
	encoder := json.NewEncoder(conn)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			// Check global rate limit first (shared across all connections)
			if !s.globalLimiter.allow() {
				s.sendError(encoder, 0, ErrRateLimited.Error())
				// Close connection when rate limited to prevent infinite loop
				return
			}

			// Check per-connection rate limit as additional protection
			if !limiter.allow() {
				s.sendError(encoder, 0, ErrRateLimited.Error())
				// Close connection when rate limited to prevent infinite loop
				return
			}

			var req APIRequest
			if err := decoder.Decode(&req); err != nil {
				if err == io.EOF {
					return
				}
				// Check if it's a size limit error
				if err.Error() == "unexpected EOF" {
					s.sendError(encoder, 0, ErrRequestTooLarge.Error())
					return
				}
				s.sendError(encoder, 0, "invalid request")
				return
			}

			response := s.handleRequest(ctx, &req)
			if err := encoder.Encode(response); err != nil {
				return
			}

			// Reset the limited reader for the next request
			// We need to recreate it since LimitReader doesn't reset
			limitedReader = io.LimitReader(conn, MaxRequestSize)
			decoder = json.NewDecoder(limitedReader)
		}
	}
}

// handleRequest routes and handles an API request
func (s *APIServer) handleRequest(ctx context.Context, req *APIRequest) *APIResponse {
	// Record request metrics
	start := time.Now()
	s.metrics.RecordRequest(req.Method)

	// Update gauges
	if s.containerManager != nil {
		s.metrics.SetContainerCount(len(s.containerManager.ListDeployments()))
	}
	if s.node != nil && s.node.router != nil {
		s.metrics.SetPeerCount(len(s.node.router.GetPeers()))
	}

	var response *APIResponse

	switch req.Method {
	case "status":
		response = s.handleStatus(ctx, req)
	case "deploy":
		response = s.handleDeploy(ctx, req)
	case "stop":
		response = s.handleStop(ctx, req)
	case "delete":
		response = s.handleDelete(ctx, req)
	case "logs":
		response = s.handleLogs(ctx, req)
	case "list":
		response = s.handleList(ctx, req)
	case "tor_start":
		response = s.handleTorStart(ctx, req)
	case "tor_stop":
		response = s.handleTorStop(ctx, req)
	case "tor_status":
		response = s.handleTorStatus(ctx, req)
	case "tor_rotate":
		response = s.handleTorRotate(ctx, req)
	case "peers":
		response = s.handlePeers(ctx, req)
	case "health":
		response = s.handleHealth(ctx, req)
	case "healthz":
		response = s.handleHealthz(ctx, req)
	case "readyz":
		response = s.handleReadyz(ctx, req)
	case "metrics":
		response = s.handleMetrics(ctx, req)
	case "config_get":
		response = s.handleConfigGet(ctx, req)
	case "config_set":
		response = s.handleConfigSet(ctx, req)
	default:
		response = &APIResponse{
			Error: fmt.Sprintf("unknown method: %s", req.Method),
			ID:    req.ID,
		}
	}

	// Record latency
	s.metrics.RecordLatency(req.Method, time.Since(start))

	return response
}

// handleStatus handles status requests
func (s *APIServer) handleStatus(ctx context.Context, req *APIRequest) *APIResponse {
	peers := s.node.router.GetPeers()
	uptime := time.Since(s.startTime).Round(time.Second)

	torRunning, torAddress := s.containerManager.GetTorStatus()

	containerCount := len(s.containerManager.ListDeployments())

	status := StatusResponse{
		NodeID:     s.node.nodeInfo.ID.String(),
		Running:    s.node.IsRunning(),
		Port:       s.node.nodeInfo.Port,
		PeerCount:  len(peers),
		Uptime:     uptime.String(),
		Version:    "0.1.0",
		TorEnabled: torRunning,
		TorAddress: torAddress,
		Containers: containerCount,
		Region:     s.node.nodeInfo.Region,
	}

	return &APIResponse{
		Result: status,
		ID:     req.ID,
	}
}

// handleDeploy handles deployment requests
func (s *APIServer) handleDeploy(ctx context.Context, req *APIRequest) *APIResponse {
	var deployReq DeployRequest
	if err := json.Unmarshal(req.Params, &deployReq); err != nil {
		return &APIResponse{
			Error: fmt.Sprintf("invalid deploy params: %v", err),
			ID:    req.ID,
		}
	}

	// Validate the deployment request
	if err := validateDeployRequest(&deployReq); err != nil {
		return &APIResponse{
			Error: fmt.Sprintf("validation failed: %v", err),
			ID:    req.ID,
		}
	}

	// Deploy via container manager
	result, err := s.containerManager.Deploy(ctx, &deployReq)
	if err != nil {
		return &APIResponse{
			Error: fmt.Sprintf("deployment failed: %v", err),
			ID:    req.ID,
		}
	}

	response := DeployResponse{
		ContainerID:     result.Deployment.ID,
		Status:          string(result.Deployment.Status),
		OnionAddress:    result.Deployment.OnionAddress,
		EncryptedVolume: result.Deployment.EncryptedVolume,
		Regions:         result.Deployment.Regions,
		ReplicaCount:    result.ReplicaCount,
	}

	return &APIResponse{
		Result: response,
		ID:     req.ID,
	}
}

// handleStop handles stop requests
func (s *APIServer) handleStop(ctx context.Context, req *APIRequest) *APIResponse {
	var params struct {
		ContainerID string `json:"container_id"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return &APIResponse{
			Error: fmt.Sprintf("invalid params: %v", err),
			ID:    req.ID,
		}
	}

	// Validate container ID
	if err := validateContainerID(params.ContainerID); err != nil {
		return &APIResponse{
			Error: fmt.Sprintf("validation failed: %v", err),
			ID:    req.ID,
		}
	}

	if err := s.containerManager.Stop(ctx, params.ContainerID); err != nil {
		return &APIResponse{
			Error: fmt.Sprintf("failed to stop container: %v", err),
			ID:    req.ID,
		}
	}

	return &APIResponse{
		Result: map[string]interface{}{
			"status":       "stopped",
			"container_id": params.ContainerID,
		},
		ID: req.ID,
	}
}

// handleDelete handles delete requests
func (s *APIServer) handleDelete(ctx context.Context, req *APIRequest) *APIResponse {
	var params struct {
		ContainerID string `json:"container_id"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return &APIResponse{
			Error: fmt.Sprintf("invalid params: %v", err),
			ID:    req.ID,
		}
	}

	// Validate container ID
	if err := validateContainerID(params.ContainerID); err != nil {
		return &APIResponse{
			Error: fmt.Sprintf("validation failed: %v", err),
			ID:    req.ID,
		}
	}

	if err := s.containerManager.Delete(ctx, params.ContainerID); err != nil {
		return &APIResponse{
			Error: fmt.Sprintf("failed to delete container: %v", err),
			ID:    req.ID,
		}
	}

	return &APIResponse{
		Result: map[string]interface{}{
			"status":       "deleted",
			"container_id": params.ContainerID,
		},
		ID: req.ID,
	}
}

// handleLogs handles log streaming requests
func (s *APIServer) handleLogs(ctx context.Context, req *APIRequest) *APIResponse {
	var logsReq LogsRequest
	if err := json.Unmarshal(req.Params, &logsReq); err != nil {
		return &APIResponse{
			Error: fmt.Sprintf("invalid logs params: %v", err),
			ID:    req.ID,
		}
	}

	// Validate container ID
	if err := validateContainerID(logsReq.ContainerID); err != nil {
		return &APIResponse{
			Error: fmt.Sprintf("validation failed: %v", err),
			ID:    req.ID,
		}
	}

	// Validate Tail parameter: must be >= 0 and <= MaxLogTailLines
	if logsReq.Tail < 0 {
		return &APIResponse{
			Error: fmt.Sprintf("%v: tail cannot be negative", ErrInvalidTailValue),
			ID:    req.ID,
		}
	}
	if logsReq.Tail > MaxLogTailLines {
		return &APIResponse{
			Error: fmt.Sprintf("%v: tail exceeds maximum of %d lines", ErrInvalidTailValue, MaxLogTailLines),
			ID:    req.ID,
		}
	}

	// Get logs from container manager
	reader, err := s.containerManager.GetLogs(ctx, logsReq.ContainerID, logsReq.Follow, logsReq.Tail)
	if err != nil {
		return &APIResponse{
			Error: fmt.Sprintf("failed to get logs: %v", err),
			ID:    req.ID,
		}
	}
	defer reader.Close()

	// Read logs
	logs, err := io.ReadAll(reader)
	if err != nil {
		return &APIResponse{
			Error: fmt.Sprintf("failed to read logs: %v", err),
			ID:    req.ID,
		}
	}

	return &APIResponse{
		Result: map[string]interface{}{
			"container_id": logsReq.ContainerID,
			"logs":         string(logs),
		},
		ID: req.ID,
	}
}

// handleList handles list deployments requests
func (s *APIServer) handleList(ctx context.Context, req *APIRequest) *APIResponse {
	deployments := s.containerManager.ListDeployments()

	containers := make([]ContainerInfo, 0, len(deployments))
	for _, d := range deployments {
		containers = append(containers, ContainerInfo{
			ID:           d.ID,
			Image:        d.Image,
			Status:       string(d.Status),
			CreatedAt:    d.CreatedAt,
			StartedAt:    d.StartedAt,
			Encrypted:    d.Encrypted,
			OnionAddress: d.OnionAddress,
			Regions:      d.Regions,
		})
	}

	return &APIResponse{
		Result: containers,
		ID:     req.ID,
	}
}

// handleTorStart handles Tor start requests
func (s *APIServer) handleTorStart(ctx context.Context, req *APIRequest) *APIResponse {
	if err := s.containerManager.StartTor(ctx); err != nil {
		return &APIResponse{
			Error: fmt.Sprintf("failed to start Tor: %v", err),
			ID:    req.ID,
		}
	}

	_, address := s.containerManager.GetTorStatus()

	return &APIResponse{
		Result: map[string]interface{}{
			"status":  "started",
			"address": address,
		},
		ID: req.ID,
	}
}

// handleTorStop handles Tor stop requests
func (s *APIServer) handleTorStop(ctx context.Context, req *APIRequest) *APIResponse {
	if err := s.containerManager.StopTor(); err != nil {
		return &APIResponse{
			Error: fmt.Sprintf("failed to stop Tor: %v", err),
			ID:    req.ID,
		}
	}

	return &APIResponse{
		Result: map[string]interface{}{
			"status": "stopped",
		},
		ID: req.ID,
	}
}

// handleTorStatus handles Tor status requests
func (s *APIServer) handleTorStatus(ctx context.Context, req *APIRequest) *APIResponse {
	running, address := s.containerManager.GetTorStatus()

	status := TorStatusResponse{
		Running:      running,
		OnionAddress: address,
		CircuitCount: -1, // -1 indicates circuit count not available
	}

	if running {
		status.StartedAt = time.Now() // Would need to track actual start time
	}

	return &APIResponse{
		Result: status,
		ID:     req.ID,
	}
}

// handleTorRotate handles Tor circuit rotation requests
func (s *APIServer) handleTorRotate(ctx context.Context, req *APIRequest) *APIResponse {
	if err := s.containerManager.RotateTorCircuit(ctx); err != nil {
		return &APIResponse{
			Error: fmt.Sprintf("failed to rotate circuit: %v", err),
			ID:    req.ID,
		}
	}

	return &APIResponse{
		Result: map[string]interface{}{
			"status": "rotated",
		},
		ID: req.ID,
	}
}

// handlePeers handles peer list requests
func (s *APIServer) handlePeers(ctx context.Context, req *APIRequest) *APIResponse {
	peers := s.node.router.GetPeers()

	peerList := make([]map[string]interface{}, 0, len(peers))
	for _, peer := range peers {
		peerList = append(peerList, map[string]interface{}{
			"id":        peer.ID.String(),
			"address":   peer.Address,
			"region":    peer.Region,
			"country":   peer.Country,
			"last_seen": peer.LastSeen,
		})
	}

	return &APIResponse{
		Result: peerList,
		ID:     req.ID,
	}
}

// handleHealth handles health check requests
func (s *APIServer) handleHealth(ctx context.Context, req *APIRequest) *APIResponse {
	var params struct {
		ContainerID string `json:"container_id"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		// Return overall health if no container specified
		unhealthy := s.containerManager.GetUnhealthyDeployments()
		return &APIResponse{
			Result: map[string]interface{}{
				"healthy":              len(unhealthy) == 0,
				"unhealthy_containers": unhealthy,
			},
			ID: req.ID,
		}
	}

	// Validate container ID if specified
	if params.ContainerID != "" {
		if err := validateContainerID(params.ContainerID); err != nil {
			return &APIResponse{
				Error: fmt.Sprintf("validation failed: %v", err),
				ID:    req.ID,
			}
		}
	}

	health, err := s.containerManager.GetHealth(ctx, params.ContainerID)
	if err != nil {
		return &APIResponse{
			Error: fmt.Sprintf("failed to get health: %v", err),
			ID:    req.ID,
		}
	}

	return &APIResponse{
		Result: health,
		ID:     req.ID,
	}
}

// handleConfigGet handles config get requests
func (s *APIServer) handleConfigGet(ctx context.Context, req *APIRequest) *APIResponse {
	return &APIResponse{
		Result: map[string]interface{}{
			"port":        s.node.nodeInfo.Port,
			"node_id":     s.node.nodeInfo.ID.String(),
			"data_dir":    s.dataDir,
			"socket_path": s.socketPath,
			"region":      s.node.nodeInfo.Region,
			"country":     s.node.nodeInfo.Country,
		},
		ID: req.ID,
	}
}

// handleConfigSet handles config set requests
func (s *APIServer) handleConfigSet(ctx context.Context, req *APIRequest) *APIResponse {
	// Runtime config changes require editing config file and restarting daemon
	// This is by design - configuration should be persistent
	return &APIResponse{
		Result: map[string]interface{}{
			"status":  "requires_restart",
			"message": "Edit ~/.moltbunker/config.yaml and restart the daemon to apply changes",
		},
		ID: req.ID,
	}
}

// handleHealthz handles detailed health check requests for liveness probes
func (s *APIServer) handleHealthz(ctx context.Context, req *APIRequest) *APIResponse {
	// Get memory stats
	var memStats goruntime.MemStats
	goruntime.ReadMemStats(&memStats)

	// Check node running status
	nodeRunning := s.node != nil && s.node.IsRunning()

	// Check containerd connection status
	containerdConnected := s.containerManager != nil && s.containerManager.IsContainerdConnected()

	// Get peer count
	peerCount := 0
	if s.node != nil && s.node.router != nil {
		peerCount = len(s.node.router.GetPeers())
	}

	// Determine overall health status
	status := "healthy"
	if !nodeRunning {
		status = "unhealthy"
	} else if !containerdConnected {
		status = "degraded"
	}

	healthz := HealthzResponse{
		Status:              status,
		NodeRunning:         nodeRunning,
		ContainerdConnected: containerdConnected,
		PeerCount:           peerCount,
		GoroutineCount:      goruntime.NumGoroutine(),
		MemoryUsageMB:       float64(memStats.Sys) / (1024 * 1024),
		MemoryAllocMB:       float64(memStats.Alloc) / (1024 * 1024),
		Timestamp:           time.Now(),
	}

	return &APIResponse{
		Result: healthz,
		ID:     req.ID,
	}
}

// handleReadyz handles readiness probe requests
func (s *APIServer) handleReadyz(ctx context.Context, req *APIRequest) *APIResponse {
	s.mu.RLock()
	running := s.running
	s.mu.RUnlock()

	// Check if the server is running and ready to accept requests
	ready := running && s.node != nil && s.node.IsRunning()

	var message string
	if !running {
		message = "API server not running"
	} else if s.node == nil {
		message = "Node not initialized"
	} else if !s.node.IsRunning() {
		message = "Node not running"
	}

	readyz := ReadyzResponse{
		Ready:     ready,
		Message:   message,
		Timestamp: time.Now(),
	}

	return &APIResponse{
		Result: readyz,
		ID:     req.ID,
	}
}

// handleMetrics handles metrics endpoint requests
func (s *APIServer) handleMetrics(ctx context.Context, req *APIRequest) *APIResponse {
	metricsData := s.metrics.GetMetrics()

	return &APIResponse{
		Result: metricsData,
		ID:     req.ID,
	}
}

// sendError sends an error response and returns any encoding error
func (s *APIServer) sendError(encoder *json.Encoder, id int, message string) error {
	if err := encoder.Encode(&APIResponse{
		Error: message,
		ID:    id,
	}); err != nil {
		return fmt.Errorf("failed to send error response: %w", err)
	}
	return nil
}

// SocketPath returns the socket path
func (s *APIServer) SocketPath() string {
	return s.socketPath
}
