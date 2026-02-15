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

// NetworkCapacity is kept as an alias for AggregatedCapacity for backward compatibility
type NetworkCapacity = AggregatedCapacity

// HardwareProfile contains detailed hardware information for a node
type HardwareProfile struct {
	CPUModel         string `json:"cpu_model"`
	CPUArch          string `json:"cpu_arch"`
	CPUThreads       int    `json:"cpu_threads"`
	CPUCores         int    `json:"cpu_cores"`
	CPUSockets       int    `json:"cpu_sockets"`
	MemoryGB         int    `json:"memory_gb"`
	MemoryType       string `json:"memory_type"`
	MemoryECC        bool   `json:"memory_ecc"`
	StorageGB        int    `json:"storage_gb"`
	StorageType      string `json:"storage_type"`
	StorageModel     string `json:"storage_model"`
	BandwidthMbps    int    `json:"bandwidth_mbps"`
	NetworkInterface string `json:"network_interface,omitempty"`
	SEVSNPSupported  bool   `json:"sev_snp_supported"`
	SEVSNPLevel      string `json:"sev_snp_level"`
	TPMVersion       string `json:"tpm_version"`
	OS               string `json:"os"`
	OSVersion        string `json:"os_version"`
	Kernel           string `json:"kernel"`
	Hostname         string `json:"hostname"`
}

// SecurityStatus contains security feature status
type SecurityStatus struct {
	TLSVersion          string `json:"tls_version"`
	EncryptionAlgo      string `json:"encryption_algo"`
	SEVSNPSupported     bool   `json:"sev_snp_supported"`
	SEVSNPActive        bool   `json:"sev_snp_active"`
	SeccompEnabled      bool   `json:"seccomp_enabled"`
	TorEnabled          bool   `json:"tor_enabled"`
	CertPinnedPeers     int    `json:"cert_pinned_peers"`
	EncryptedContainers int    `json:"encrypted_containers"`
	TotalContainers     int    `json:"total_containers"`
}

// StatusResponse contains node status information
type StatusResponse struct {
	NodeID       string              `json:"node_id"`
	Running      bool                `json:"running"`
	Port         int                 `json:"port"`
	NetworkNodes int                 `json:"network_nodes"`
	Uptime       string              `json:"uptime"`
	Version      string              `json:"version"`
	TorEnabled   bool                `json:"tor_enabled"`
	TorAddress   string              `json:"tor_address,omitempty"`
	Containers   int                 `json:"containers"`
	Region       string              `json:"region"`
	Location     *types.NodeLocation `json:"location,omitempty"`

	// Extended fields
	NetworkCapacity *AggregatedCapacity `json:"network_capacity,omitempty"`
	Security        *SecurityStatus     `json:"security,omitempty"`
	NodeTier        string              `json:"node_tier,omitempty"`
	NodeRole        string              `json:"node_role,omitempty"`
	ReputationScore int                 `json:"reputation_score"`
	KnownNodes      []NodeProfile       `json:"known_nodes,omitempty"`
}

// DeployRequest contains deployment parameters
type DeployRequest struct {
	Image           string               `json:"image"`
	Resources       types.ResourceLimits `json:"resources,omitempty"`
	Duration        string               `json:"duration,omitempty"`          // Job duration (e.g. "24h", "720h"); default: 720h (30 days)
	TorOnly         bool                 `json:"tor_only"`
	OnionService    bool                 `json:"onion_service"`
	OnionPort       int                  `json:"onion_port,omitempty"`        // Port to expose via Tor (default: 80)
	WaitForReplicas bool                 `json:"wait_for_replicas,omitempty"` // If true, wait for at least 1 replica ack before returning
	ReservationID   string               `json:"reservation_id,omitempty"`    // On-chain escrow reservation ID (user-created)
	Owner           string               `json:"owner,omitempty"`             // Wallet address of the deployer

	// Minimum provider tier requirement
	MinProviderTier string `json:"min_provider_tier,omitempty"` // "confidential", "standard", "dev", or empty for any

	// E2E exec encryption (optional)
	EncryptedExecKey []byte `json:"encrypted_exec_key,omitempty"` // Exec key encrypted with provider's X25519 pubkey
	ExecKeyNonce     []byte `json:"exec_key_nonce,omitempty"`     // Nonce for exec key decryption
	DeployNonce      string `json:"deploy_nonce,omitempty"`       // Deploy nonce used to derive exec_key

	// Service exposure (optional)
	ExposePorts []ExposedPort `json:"expose_ports,omitempty"` // Ports to expose publicly
}

// ExposedPort describes a port to expose publicly via ingress.
type ExposedPort struct {
	ContainerPort int    `json:"container_port"`
	Protocol      string `json:"protocol,omitempty"` // "tcp" (default) or "udp"
}

// ReplicaLocation describes where a single replica is deployed
type ReplicaLocation struct {
	Region      string `json:"region"`
	Country     string `json:"country,omitempty"`
	CountryName string `json:"country_name,omitempty"`
	City        string `json:"city,omitempty"`
}

// DeployResponse contains deployment result
type DeployResponse struct {
	ContainerID     string            `json:"container_id"`
	OnionAddress    string            `json:"onion_address,omitempty"`
	Status          string            `json:"status"`
	EncryptedVolume string            `json:"encrypted_volume,omitempty"`
	Regions         []string          `json:"regions"`
	Locations       []ReplicaLocation `json:"locations,omitempty"`
	ReplicaCount    int               `json:"replica_count"`             // Number of successful replica acks received
	PublicURLs      []string          `json:"public_urls,omitempty"`     // Public URLs if ports are exposed
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
	ID              string            `json:"id"`
	Image           string            `json:"image"`
	Status          string            `json:"status"`
	CreatedAt       time.Time         `json:"created_at"`
	StartedAt       time.Time         `json:"started_at,omitempty"`
	Encrypted       bool              `json:"encrypted"`
	OnionAddress    string            `json:"onion_address,omitempty"`
	Regions         []string          `json:"regions"`
	Locations       []ReplicaLocation `json:"locations,omitempty"`
	Owner           string            `json:"owner,omitempty"`
	StoppedAt       time.Time         `json:"stopped_at,omitempty"`
	VolumeExpiresAt time.Time         `json:"volume_expires_at,omitempty"`
	HasVolume       bool              `json:"has_volume"`
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
