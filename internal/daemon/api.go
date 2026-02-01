package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/moltbunker/moltbunker/pkg/types"
)

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
	Image        string               `json:"image"`
	Resources    types.ResourceLimits `json:"resources,omitempty"`
	TorOnly      bool                 `json:"tor_only"`
	OnionService bool                 `json:"onion_service"`
	OnionPort    int                  `json:"onion_port,omitempty"` // Port to expose via Tor (default: 80)
}

// DeployResponse contains deployment result
type DeployResponse struct {
	ContainerID     string `json:"container_id"`
	OnionAddress    string `json:"onion_address,omitempty"`
	Status          string `json:"status"`
	EncryptedVolume string `json:"encrypted_volume,omitempty"`
	Regions         []string `json:"regions"`
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

// NewAPIServer creates a new API server
func NewAPIServer(node *Node, socketPath string, dataDir string) *APIServer {
	return &APIServer{
		node:       node,
		socketPath: socketPath,
		dataDir:    dataDir,
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

	// Create Unix socket listener
	listener, err := net.Listen("unix", s.socketPath)
	if err != nil {
		s.mu.Unlock()
		return fmt.Errorf("failed to create unix socket: %w", err)
	}

	// Set permissions
	if err := os.Chmod(s.socketPath, 0600); err != nil {
		listener.Close()
		s.mu.Unlock()
		return fmt.Errorf("failed to set socket permissions: %w", err)
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

	// Handle connections
	go s.handleConnections(ctx)

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

			go s.handleConnection(ctx, conn)
		}
	}
}

// handleConnection handles a single API connection
func (s *APIServer) handleConnection(ctx context.Context, conn net.Conn) {
	defer conn.Close()

	decoder := json.NewDecoder(conn)
	encoder := json.NewEncoder(conn)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			var req APIRequest
			if err := decoder.Decode(&req); err != nil {
				if err == io.EOF {
					return
				}
				s.sendError(encoder, 0, "invalid request")
				return
			}

			response := s.handleRequest(ctx, &req)
			if err := encoder.Encode(response); err != nil {
				return
			}
		}
	}
}

// handleRequest routes and handles an API request
func (s *APIServer) handleRequest(ctx context.Context, req *APIRequest) *APIResponse {
	switch req.Method {
	case "status":
		return s.handleStatus(ctx, req)
	case "deploy":
		return s.handleDeploy(ctx, req)
	case "stop":
		return s.handleStop(ctx, req)
	case "delete":
		return s.handleDelete(ctx, req)
	case "logs":
		return s.handleLogs(ctx, req)
	case "list":
		return s.handleList(ctx, req)
	case "tor_start":
		return s.handleTorStart(ctx, req)
	case "tor_stop":
		return s.handleTorStop(ctx, req)
	case "tor_status":
		return s.handleTorStatus(ctx, req)
	case "tor_rotate":
		return s.handleTorRotate(ctx, req)
	case "peers":
		return s.handlePeers(ctx, req)
	case "health":
		return s.handleHealth(ctx, req)
	case "config_get":
		return s.handleConfigGet(ctx, req)
	case "config_set":
		return s.handleConfigSet(ctx, req)
	default:
		return &APIResponse{
			Error: fmt.Sprintf("unknown method: %s", req.Method),
			ID:    req.ID,
		}
	}
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

	// Deploy via container manager
	deployment, err := s.containerManager.Deploy(ctx, &deployReq)
	if err != nil {
		return &APIResponse{
			Error: fmt.Sprintf("deployment failed: %v", err),
			ID:    req.ID,
		}
	}

	response := DeployResponse{
		ContainerID:     deployment.ID,
		Status:          string(deployment.Status),
		OnionAddress:    deployment.OnionAddress,
		EncryptedVolume: deployment.EncryptedVolume,
		Regions:         deployment.Regions,
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
				"healthy":             len(unhealthy) == 0,
				"unhealthy_containers": unhealthy,
			},
			ID: req.ID,
		}
	}

	health, err := s.containerManager.GetHealth(params.ContainerID)
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

// sendError sends an error response
func (s *APIServer) sendError(encoder *json.Encoder, id int, message string) {
	encoder.Encode(&APIResponse{
		Error: message,
		ID:    id,
	})
}

// SocketPath returns the socket path
func (s *APIServer) SocketPath() string {
	return s.socketPath
}
