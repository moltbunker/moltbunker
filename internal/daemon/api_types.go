package daemon

import (
	"encoding/json"
	"time"

	"github.com/moltbunker/moltbunker/pkg/types"
)

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
	Image           string               `json:"image"`
	Resources       types.ResourceLimits `json:"resources,omitempty"`
	TorOnly         bool                 `json:"tor_only"`
	OnionService    bool                 `json:"onion_service"`
	OnionPort       int                  `json:"onion_port,omitempty"`        // Port to expose via Tor (default: 80)
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
	ID           string    `json:"id"`
	Image        string    `json:"image"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	StartedAt    time.Time `json:"started_at,omitempty"`
	Encrypted    bool      `json:"encrypted"`
	OnionAddress string    `json:"onion_address,omitempty"`
	Regions      []string  `json:"regions"`
}

// HealthzResponse contains detailed health check information
type HealthzResponse struct {
	Status              string    `json:"status"` // "healthy", "degraded", or "unhealthy"
	NodeRunning         bool      `json:"node_running"`
	ContainerdConnected bool      `json:"containerd_connected"`
	PeerCount           int       `json:"peer_count"`
	GoroutineCount      int       `json:"goroutine_count"`
	MemoryUsageMB       float64   `json:"memory_usage_mb"`
	MemoryAllocMB       float64   `json:"memory_alloc_mb"`
	Timestamp           time.Time `json:"timestamp"`
}

// ReadyzResponse contains readiness probe information
type ReadyzResponse struct {
	Ready     bool      `json:"ready"`
	Message   string    `json:"message,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}
