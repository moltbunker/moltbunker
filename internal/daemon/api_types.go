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
	Duration        string               `json:"duration,omitempty"` // Job duration (e.g. "24h", "720h"); default: 720h (30 days)
	TorOnly         bool                 `json:"tor_only"`
	OnionService    bool                 `json:"onion_service"`
	OnionPort       int                  `json:"onion_port,omitempty"`        // Port to expose via Tor (default: 80)
	WaitForReplicas bool                 `json:"wait_for_replicas,omitempty"` // If true, wait for at least 1 replica ack before returning
	ReservationID   string               `json:"reservation_id,omitempty"`    // On-chain escrow reservation ID (user-created)
	Owner           string               `json:"owner,omitempty"`             // Wallet address of the deployer

	// Minimum provider tier requirement
	MinProviderTier string `json:"min_provider_tier,omitempty"` // "confidential", "standard", "dev", or empty for any

	// E2E exec encryption (optional). The CLI seals the 32-byte exec_key to the
	// provider's stable X25519 public key using ECIES (ephemeral-static X25519 ->
	// HKDF-SHA256 -> AES-256-GCM). The envelope below carries everything the
	// daemon needs to unwrap it with its stable private key.
	EncryptedExecKey         []byte `json:"encrypted_exec_key,omitempty"`          // ECIES envelope ciphertext: gcm_nonce(12) || ciphertext || tag(16)
	ExecKeyNonce             []byte `json:"exec_key_nonce,omitempty"`              // GCM nonce (also prefixed in EncryptedExecKey; carried for transport parity)
	RequesterEphemeralPubKey []byte `json:"requester_ephemeral_pub_key,omitempty"` // Sender's ephemeral X25519 public key (32 bytes)
	DeployNonce              string `json:"deploy_nonce,omitempty"`                // Deploy nonce used to derive exec_key

	// Spot pricing (optional — lower cost, preemptible)
	Spot bool `json:"spot,omitempty"`

	// Service exposure (optional)
	ExposePorts []ExposedPort `json:"expose_ports,omitempty"` // Ports to expose publicly

	// R3 — image signature verification (optional, opt-out by default).
	// All fields zero/empty => no verification (identical to legacy behavior).
	// RequireSignature only takes effect when at least one TrustedPublisher is
	// provided, otherwise it would deny-all (see DeployRequest.toTrustPolicy).
	RequireSignature  bool                `json:"require_signature,omitempty"`  // Enforce a valid signature before create
	TrustedPublishers []string            `json:"trusted_publishers,omitempty"` // Hex-encoded Ed25519 pubkeys allowed to sign
	ImageSignature    *ImageSignatureSpec `json:"image_signature,omitempty"`    // Caller-supplied signature for the image digest

	// R4 — image vulnerability scan policy (optional). When unset, the daemon's
	// default scan policy applies (block HIGH/CRITICAL, never RequireScan).
	IgnoreCVEs []string `json:"ignore_cves,omitempty"` // Per-deployment CVE allowlist (e.g. "CVE-2024-1234")

	// R13/R14 — per-deployment network / egress policy (optional). A nil
	// NetworkPolicy means allow-all (current behavior). The real nft
	// enforcement is Linux-only and stubbed today; the policy is recorded and
	// flows toward the enforcer regardless of platform.
	NetworkPolicy *NetworkPolicySpec `json:"network_policy,omitempty"`
}

// ImageSignatureSpec is the wire form of an Ed25519 image signature carried on
// a deploy request (R3). It maps 1:1 to runtime.ImageSignature.
type ImageSignatureSpec struct {
	Digest      string `json:"digest"`       // Image digest, typically "sha256:<hex>"
	PublisherID string `json:"publisher_id"` // Hex-encoded 32-byte Ed25519 public key
	Signature   []byte `json:"signature"`    // Ed25519 signature over the digest
}

// NetworkPolicySpec is the wire form of a per-deployment network/egress policy
// (R13/R14). It maps to networking.NetworkPolicy. A nil spec or fully-empty
// spec means allow-all (EgressDefaultAllow, no carve-outs) — identical to the
// legacy behavior.
type NetworkPolicySpec struct {
	AllowedPeers []string `json:"allowed_peers,omitempty"` // Other deployment IDs reachable intra-host
	EgressDeny   bool     `json:"egress_deny,omitempty"`   // true => default-deny egress (EgressDefaultDeny)
	EgressAllow  []string `json:"egress_allow,omitempty"`  // CIDRs always allowed
	EgressBlock  []string `json:"egress_block,omitempty"`  // CIDRs always blocked (deny beats allow)
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
	ReplicaCount    int               `json:"replica_count"`         // Number of successful replica acks received
	PublicURLs      []string          `json:"public_urls,omitempty"` // Public URLs if ports are exposed
	// ExecAgentEnabled reports whether the deploy carried a valid E2E exec
	// envelope and the exec-agent was injected. The caller (CLI or browser) uses
	// this to decide whether the exec WebSocket must perform the KEY_INIT/KEY_ACK
	// handshake and encrypt terminal I/O.
	ExecAgentEnabled bool `json:"exec_agent_enabled,omitempty"`
	// DeployNonce is the hex-encoded nonce used as the HKDF salt when deriving
	// the container's exec_key; required to re-derive the key for the exec
	// handshake. Non-secret.
	DeployNonce string `json:"deploy_nonce,omitempty"`
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

// SubdomainRegisterRequest contains subdomain registration parameters.
type SubdomainRegisterRequest struct {
	Name         string `json:"name"`
	DeploymentID string `json:"deployment_id"`
}

// SubdomainRegisterResponse contains the result of a subdomain registration.
type SubdomainRegisterResponse struct {
	Name         string `json:"name"`
	DeploymentID string `json:"deployment_id"`
	URL          string `json:"url"`
	TxHash       string `json:"tx_hash,omitempty"`
}

// SubdomainInfo contains subdomain details.
type SubdomainInfo struct {
	Name         string    `json:"name"`
	DeploymentID string    `json:"deployment_id"`
	Owner        string    `json:"owner"`
	URL          string    `json:"url"`
	RegisteredAt time.Time `json:"registered_at"`
}

// SubdomainReleaseRequest contains subdomain release parameters.
type SubdomainReleaseRequest struct {
	Name string `json:"name"`
}

// SubdomainResolveRequest contains subdomain resolve parameters.
type SubdomainResolveRequest struct {
	Name string `json:"name"`
}

// SubdomainTransferRequest contains subdomain transfer parameters.
type SubdomainTransferRequest struct {
	Name     string `json:"name"`
	NewOwner string `json:"new_owner"`
}

// SubdomainUpdateRequest contains subdomain deployment update parameters.
type SubdomainUpdateRequest struct {
	Name         string `json:"name"`
	DeploymentID string `json:"deployment_id"`
}

// SubdomainRenewRequest contains subdomain renewal parameters.
type SubdomainRenewRequest struct {
	Name string `json:"name"`
}

// SubdomainReserveRequest contains subdomain reservation parameters.
type SubdomainReserveRequest struct {
	Name string `json:"name"`
}

// SubdomainClaimRequest contains subdomain claim parameters.
type SubdomainClaimRequest struct {
	Name         string `json:"name"`
	DeploymentID string `json:"deployment_id"`
}

// SubdomainCancelRequest contains subdomain reservation cancellation parameters.
type SubdomainCancelRequest struct {
	Name string `json:"name"`
}

// SubdomainMetadataRequest contains subdomain metadata update parameters.
type SubdomainMetadataRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	AvatarURL   string `json:"avatar_url"`
}

// SubdomainPrimaryRequest contains subdomain primary name parameters.
type SubdomainPrimaryRequest struct {
	Name string `json:"name"`
}

// SubdomainReclaimRequest contains subdomain reclaim parameters.
type SubdomainReclaimRequest struct {
	Name string `json:"name"`
}

// SubdomainMetadataResponse contains subdomain metadata.
type SubdomainMetadataResponse struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	AvatarURL   string `json:"avatar_url"`
}

// --- Molt (WASM serverless) API types ---

// MoltDeployRequest is the API request to deploy a Molt serverless function.
type MoltDeployRequest struct {
	ModuleCID     string            `json:"module_cid"`                // IPFS CID of the .wasm binary
	MemoryLimitMB uint32            `json:"memory_limit_mb,omitempty"` // Max WASM memory (default: 256MB)
	TimeoutMs     int               `json:"timeout_ms,omitempty"`      // Max execution time (default: 30s)
	MaxInstances  int               `json:"max_instances,omitempty"`   // Max concurrent instances (default: 100)
	Environment   map[string]string `json:"environment,omitempty"`     // Env vars passed to WASM
	Owner         string            `json:"owner,omitempty"`           // Deployer wallet address
	WasmBytes     []byte            `json:"wasm_bytes,omitempty"`      // Inline WASM binary (if not using IPFS)
}

// MoltDeployResponse is the API response after deploying a Molt.
type MoltDeployResponse struct {
	DeploymentID string `json:"deployment_id"`
	ModuleCID    string `json:"module_cid"`
	Status       string `json:"status"`
}

// MoltInfo describes a deployed Molt for list/get responses.
type MoltInfo struct {
	ID            string                       `json:"id"`
	ModuleCID     string                       `json:"module_cid"`
	Status        string                       `json:"status"`
	CreatedAt     time.Time                    `json:"created_at"`
	Owner         string                       `json:"owner,omitempty"`
	MemoryLimitMB uint32                       `json:"memory_limit_mb,omitempty"`
	TimeoutMs     int                          `json:"timeout_ms,omitempty"`
	Metrics       *types.MoltDeploymentMetrics `json:"metrics,omitempty"`
}

// MoltInvokeRequest is the API request to invoke a Molt directly.
type MoltInvokeRequest struct {
	DeploymentID string            `json:"deployment_id"`
	Method       string            `json:"method"` // HTTP method (default: GET)
	Path         string            `json:"path"`   // HTTP path (default: /)
	Headers      map[string]string `json:"headers,omitempty"`
	Body         []byte            `json:"body,omitempty"` // Request body
}

// MoltInvokeResponse is the API response from a Molt invocation.
type MoltInvokeResponse struct {
	StatusCode int               `json:"status_code"`
	Headers    map[string]string `json:"headers,omitempty"`
	Body       []byte            `json:"body,omitempty"`
	DurationMs int64             `json:"duration_ms"`
	Error      string            `json:"error,omitempty"`
}
