package client

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/moltbunker/moltbunker/pkg/types"
)

// DaemonClient provides communication with the moltbunker daemon
type DaemonClient struct {
	socketPath string
	conn       net.Conn
	mu         sync.Mutex
	encoder    *json.Encoder
	decoder    *json.Decoder
	reqID      int
}

// APIRequest represents a JSON-RPC style request
type APIRequest struct {
	Method string      `json:"method"`
	Params interface{} `json:"params,omitempty"`
	ID     int         `json:"id"`
}

// APIResponse represents a JSON-RPC style response
type APIResponse struct {
	Result json.RawMessage `json:"result,omitempty"`
	Error  string          `json:"error,omitempty"`
	ID     int             `json:"id"`
}

// StatusResponse contains node status information
type StatusResponse struct {
	NodeID     string `json:"node_id"`
	Running    bool   `json:"running"`
	Port       int    `json:"port"`
	PeerCount  int    `json:"peer_count"`
	Uptime     string `json:"uptime"`
	Version    string `json:"version"`
	TorEnabled bool   `json:"tor_enabled"`
	TorAddress string `json:"tor_address,omitempty"`
	Containers int    `json:"containers"`
	Region     string `json:"region"`
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
	ContainerID     string   `json:"container_id"`
	OnionAddress    string   `json:"onion_address,omitempty"`
	Status          string   `json:"status"`
	EncryptedVolume string   `json:"encrypted_volume,omitempty"`
	Regions         []string `json:"regions"`
}

// ContainerInfo contains container information
type ContainerInfo struct {
	ID           string    `json:"id"`
	Image        string    `json:"image"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	StartedAt    time.Time `json:"started_at,omitempty"`
	Encrypted    bool      `json:"encrypted"`
	OnionAddress string    `json:"onion_address,omitempty"`
	Regions      []string  `json:"regions"`
}

// HealthResponse contains health information
type HealthResponse struct {
	Healthy             bool           `json:"healthy"`
	UnhealthyContainers map[string][]int `json:"unhealthy_containers,omitempty"`
}

// TorStatusResponse contains Tor status information
type TorStatusResponse struct {
	Running      bool      `json:"running"`
	OnionAddress string    `json:"onion_address,omitempty"`
	StartedAt    time.Time `json:"started_at,omitempty"`
	CircuitCount int       `json:"circuit_count"`
}

// PeerInfo contains peer information
type PeerInfo struct {
	ID       string    `json:"id"`
	Address  string    `json:"address"`
	Region   string    `json:"region"`
	LastSeen time.Time `json:"last_seen"`
}

// LogsRequest contains log streaming parameters
type LogsRequest struct {
	ContainerID string `json:"container_id"`
	Follow      bool   `json:"follow"`
	Tail        int    `json:"tail"`
}

// NewDaemonClient creates a new daemon client
func NewDaemonClient(socketPath string) *DaemonClient {
	if socketPath == "" {
		socketPath = DefaultSocketPath()
	}
	return &DaemonClient{
		socketPath: socketPath,
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

// Connect establishes connection to the daemon
func (c *DaemonClient) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		return nil // Already connected
	}

	conn, err := net.Dial("unix", c.socketPath)
	if err != nil {
		return fmt.Errorf("failed to connect to daemon: %w", err)
	}

	c.conn = conn
	c.encoder = json.NewEncoder(conn)
	c.decoder = json.NewDecoder(conn)

	return nil
}

// Close closes the connection to the daemon
func (c *DaemonClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil {
		return nil
	}

	err := c.conn.Close()
	c.conn = nil
	c.encoder = nil
	c.decoder = nil

	return err
}

// call makes an API call to the daemon
func (c *DaemonClient) call(method string, params interface{}) (*APIResponse, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil {
		return nil, fmt.Errorf("not connected to daemon")
	}

	c.reqID++
	req := APIRequest{
		Method: method,
		Params: params,
		ID:     c.reqID,
	}

	if err := c.encoder.Encode(req); err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	var resp APIResponse
	if err := c.decoder.Decode(&resp); err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.Error != "" {
		return nil, fmt.Errorf("daemon error: %s", resp.Error)
	}

	return &resp, nil
}

// Status retrieves the daemon status
func (c *DaemonClient) Status() (*StatusResponse, error) {
	resp, err := c.call("status", nil)
	if err != nil {
		return nil, err
	}

	var status StatusResponse
	if err := json.Unmarshal(resp.Result, &status); err != nil {
		return nil, fmt.Errorf("failed to parse status: %w", err)
	}

	return &status, nil
}

// Deploy deploys a container
func (c *DaemonClient) Deploy(req *DeployRequest) (*DeployResponse, error) {
	resp, err := c.call("deploy", req)
	if err != nil {
		return nil, err
	}

	var deployResp DeployResponse
	if err := json.Unmarshal(resp.Result, &deployResp); err != nil {
		return nil, fmt.Errorf("failed to parse deploy response: %w", err)
	}

	return &deployResp, nil
}

// GetLogs retrieves container logs
func (c *DaemonClient) GetLogs(containerID string, follow bool, tail int) (string, error) {
	req := LogsRequest{
		ContainerID: containerID,
		Follow:      follow,
		Tail:        tail,
	}

	resp, err := c.call("logs", req)
	if err != nil {
		return "", err
	}

	var result struct {
		Logs string `json:"logs"`
	}
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return "", fmt.Errorf("failed to parse logs: %w", err)
	}

	return result.Logs, nil
}

// TorStart starts the Tor service
func (c *DaemonClient) TorStart() (map[string]interface{}, error) {
	resp, err := c.call("tor_start", nil)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return result, nil
}

// TorStatus retrieves Tor service status
func (c *DaemonClient) TorStatus() (*TorStatusResponse, error) {
	resp, err := c.call("tor_status", nil)
	if err != nil {
		return nil, err
	}

	var status TorStatusResponse
	if err := json.Unmarshal(resp.Result, &status); err != nil {
		return nil, fmt.Errorf("failed to parse tor status: %w", err)
	}

	return &status, nil
}

// TorRotate rotates the Tor circuit
func (c *DaemonClient) TorRotate() error {
	_, err := c.call("tor_rotate", nil)
	return err
}

// GetPeers retrieves the list of connected peers
func (c *DaemonClient) GetPeers() ([]PeerInfo, error) {
	resp, err := c.call("peers", nil)
	if err != nil {
		return nil, err
	}

	var peers []PeerInfo
	if err := json.Unmarshal(resp.Result, &peers); err != nil {
		return nil, fmt.Errorf("failed to parse peers: %w", err)
	}

	return peers, nil
}

// ConfigGet retrieves configuration values
func (c *DaemonClient) ConfigGet(key string) (interface{}, error) {
	resp, err := c.call("config_get", map[string]string{"key": key})
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return result, nil
}

// ConfigSet sets a configuration value
func (c *DaemonClient) ConfigSet(key string, value interface{}) error {
	_, err := c.call("config_set", map[string]interface{}{
		"key":   key,
		"value": value,
	})
	return err
}

// IsDaemonRunning checks if the daemon is running
func (c *DaemonClient) IsDaemonRunning() bool {
	if err := c.Connect(); err != nil {
		return false
	}

	_, err := c.Status()
	return err == nil
}

// SocketPath returns the socket path
func (c *DaemonClient) SocketPath() string {
	return c.socketPath
}

// List lists all deployed containers
func (c *DaemonClient) List() ([]ContainerInfo, error) {
	resp, err := c.call("list", nil)
	if err != nil {
		return nil, err
	}

	var containers []ContainerInfo
	if err := json.Unmarshal(resp.Result, &containers); err != nil {
		return nil, fmt.Errorf("failed to parse containers: %w", err)
	}

	return containers, nil
}

// Stop stops a container
func (c *DaemonClient) Stop(containerID string) error {
	_, err := c.call("stop", map[string]string{"container_id": containerID})
	return err
}

// Delete deletes a container
func (c *DaemonClient) Delete(containerID string) error {
	_, err := c.call("delete", map[string]string{"container_id": containerID})
	return err
}

// Health retrieves health information
func (c *DaemonClient) Health(containerID string) (*HealthResponse, error) {
	var params interface{}
	if containerID != "" {
		params = map[string]string{"container_id": containerID}
	}

	resp, err := c.call("health", params)
	if err != nil {
		return nil, err
	}

	var health HealthResponse
	if err := json.Unmarshal(resp.Result, &health); err != nil {
		return nil, fmt.Errorf("failed to parse health: %w", err)
	}

	return &health, nil
}

// TorStop stops the Tor service
func (c *DaemonClient) TorStop() error {
	_, err := c.call("tor_stop", nil)
	return err
}
