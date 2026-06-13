package config

import (
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/moltbunker/moltbunker/pkg/types"
	"gopkg.in/yaml.v3"
)

// Config represents the complete daemon configuration
type Config struct {
	Daemon     DaemonConfig     `yaml:"daemon"`
	API        APIConfig        `yaml:"api"`
	Node       NodeConfig       `yaml:"node"`
	P2P        P2PConfig        `yaml:"p2p"`
	Tor        TorConfig        `yaml:"tor"`
	Runtime    RuntimeConfig    `yaml:"runtime"`
	Security   SecurityConfig   `yaml:"security"`
	Redundancy RedundancyConfig `yaml:"redundancy"`
	Economics  EconomicsConfig  `yaml:"economics"`
	Encryption EncryptionConfig `yaml:"encryption"`

	// P0 services
	Storage StorageConfig `yaml:"storage"`
	Proxy   ProxyConfig   `yaml:"proxy"`
	Agent   AgentConfig   `yaml:"agent"`
	Crawl   CrawlConfig   `yaml:"crawl"`
}

// DaemonConfig contains daemon settings
type DaemonConfig struct {
	Port        int    `yaml:"port"`
	DataDir     string `yaml:"data_dir"`
	KeyPath     string `yaml:"key_path"`
	KeystoreDir string `yaml:"keystore_dir"`
	SocketPath  string `yaml:"socket_path"`
	LogLevel    string `yaml:"log_level"`
	LogFormat   string `yaml:"log_format"`    // "json" or "text"
	StateDBPath string `yaml:"state_db_path"` // Path to bbolt state database (default: <data_dir>/moltbunker.db)
}

// APIConfig contains API server settings
type APIConfig struct {
	// Rate limiting
	RateLimitRequests   int `yaml:"rate_limit_requests"`    // Max requests per window (default: 100)
	RateLimitWindowSecs int `yaml:"rate_limit_window_secs"` // Window duration in seconds (default: 60)

	// Connection limits
	MaxConcurrentConns int `yaml:"max_concurrent_conns"` // Max concurrent connections (default: 100)
	MaxRequestSize     int `yaml:"max_request_size"`     // Max request body size in bytes (default: 10MB)

	// Timeouts
	ReadTimeoutSecs  int `yaml:"read_timeout_secs"`  // Read timeout (default: 30)
	WriteTimeoutSecs int `yaml:"write_timeout_secs"` // Write timeout (default: 30)
	IdleTimeoutSecs  int `yaml:"idle_timeout_secs"`  // Idle connection timeout (default: 120)

	// Admin
	AdminWallets []string `yaml:"admin_wallets"` // Ethereum addresses with admin access
}

// DefaultAPIConfig returns the default API configuration
func DefaultAPIConfig() APIConfig {
	return APIConfig{
		RateLimitRequests:   100,
		RateLimitWindowSecs: 60,
		MaxConcurrentConns:  100,
		MaxRequestSize:      10 * 1024 * 1024, // 10MB
		ReadTimeoutSecs:     30,
		WriteTimeoutSecs:    30,
		IdleTimeoutSecs:     120,
	}
}

// NodeConfig contains node role and identity settings
type NodeConfig struct {
	Role               types.NodeRole `yaml:"role"`                 // provider, requester, hybrid
	Region             string         `yaml:"region"`               // Manual region override (Americas, Europe, Asia-Pacific, Middle-East, Africa)
	WalletAddress      string         `yaml:"wallet_address"`       // Ethereum wallet address
	WalletKeyFile      string         `yaml:"wallet_key_file"`      // Path to encrypted keystore file
	WalletPasswordFile string         `yaml:"wallet_password_file"` // Path to file containing wallet password
	AutoRegister       bool           `yaml:"auto_register"`        // Auto-register on startup (provider)
	APIEndpoint        string         `yaml:"api_endpoint"`         // HTTP API URL for daemonless requester mode

	// Provider-specific settings
	Provider ProviderNodeConfig `yaml:"provider,omitempty"`

	// Requester-specific settings
	Requester RequesterNodeConfig `yaml:"requester,omitempty"`
}

// HardwareProfile contains detailed hardware information for a node.
// Auto-detected at startup; any field can be overridden in config.yaml.
type HardwareProfile struct {
	// CPU
	CPUModel   string `yaml:"cpu_model" json:"cpu_model"`     // "AMD EPYC 8224P"
	CPUArch    string `yaml:"cpu_arch" json:"cpu_arch"`       // "x86_64" / "arm64"
	CPUThreads int    `yaml:"cpu_threads" json:"cpu_threads"` // Logical CPUs (48)
	CPUCores   int    `yaml:"cpu_cores" json:"cpu_cores"`     // Physical cores (24)
	CPUSockets int    `yaml:"cpu_sockets" json:"cpu_sockets"` // Socket count (1)

	// Memory
	MemoryGB   int    `yaml:"memory_gb" json:"memory_gb"`     // Total RAM (96)
	MemoryType string `yaml:"memory_type" json:"memory_type"` // "DDR5" / "DDR4"
	MemoryECC  bool   `yaml:"memory_ecc" json:"memory_ecc"`   // ECC support

	// Storage
	StorageGB    int    `yaml:"storage_gb" json:"storage_gb"`       // Total disk (960)
	StorageType  string `yaml:"storage_type" json:"storage_type"`   // "NVMe" / "SSD" / "HDD"
	StorageModel string `yaml:"storage_model" json:"storage_model"` // "Samsung PM9A3"

	// Network
	BandwidthMbps    int    `yaml:"bandwidth_mbps" json:"bandwidth_mbps"`                           // 1000
	NetworkInterface string `yaml:"network_interface,omitempty" json:"network_interface,omitempty"` // "eth0"

	// Security Hardware
	SEVSNPSupported bool   `yaml:"sev_snp_supported" json:"sev_snp_supported"` // AMD SEV-SNP
	SEVSNPLevel     string `yaml:"sev_snp_level" json:"sev_snp_level"`         // "snp" / "sev" / "none"
	TPMVersion      string `yaml:"tpm_version" json:"tpm_version"`             // "2.0" / "none"

	// OS
	OS        string `yaml:"os" json:"os"`                 // "linux" / "darwin"
	OSVersion string `yaml:"os_version" json:"os_version"` // "Ubuntu 24.04" / "macOS 15.3"
	Kernel    string `yaml:"kernel" json:"kernel"`         // "6.8.0-49-generic"
	Hostname  string `yaml:"hostname" json:"hostname"`     // "moltbunker-main"
}

// ProviderNodeConfig contains provider-specific configuration
type ProviderNodeConfig struct {
	// Declared resources (populated from Hardware profile if not set)
	DeclaredCPU       int `yaml:"declared_cpu"`            // Number of CPU cores
	DeclaredMemoryGB  int `yaml:"declared_memory_gb"`      // Memory in GB
	DeclaredStorageGB int `yaml:"declared_storage_gb"`     // Storage in GB
	DeclaredBandwidth int `yaml:"declared_bandwidth_mbps"` // Network bandwidth in Mbps

	// Detailed hardware profile (auto-detected, admin-overridable)
	Hardware HardwareProfile `yaml:"hardware" json:"hardware"`

	// GPU (optional)
	GPUEnabled  bool   `yaml:"gpu_enabled"`
	GPUModel    string `yaml:"gpu_model,omitempty"`
	GPUCount    int    `yaml:"gpu_count,omitempty"`
	GPUMemoryGB int    `yaml:"gpu_memory_gb,omitempty"`

	// Staking
	TargetTier types.StakingTier `yaml:"target_tier"` // Desired staking tier
	AutoStake  bool              `yaml:"auto_stake"`  // Auto-stake on startup

	// Job acceptance
	AcceptServices  bool `yaml:"accept_services"`  // Accept long-running services
	AcceptJobs      bool `yaml:"accept_jobs"`      // Accept batch jobs
	AcceptScheduled bool `yaml:"accept_scheduled"` // Accept scheduled jobs
	AcceptFunctions bool `yaml:"accept_functions"` // Accept serverless functions

	// Availability
	MaintenanceMode bool `yaml:"maintenance_mode"` // Don't accept new jobs

	// Ingress (service exposure)
	IngressEnabled bool   `yaml:"ingress_enabled"` // Act as ingress node for exposed services
	IngressPort    int    `yaml:"ingress_port"`    // HTTP port for ingress proxy (default: 9090)
	IngressDomain  string `yaml:"ingress_domain"`  // Domain for public URLs (e.g., "moltbunker.dev")
	TunnelPort     int    `yaml:"tunnel_port"`     // TLS tunnel port for ingress→provider (default: base port + 2)

	// Cloudflare DNS sync for subdomain A records
	CloudflareAPIToken string `yaml:"cloudflare_api_token,omitempty"` // Cloudflare API bearer token
	CloudflareZoneID   string `yaml:"cloudflare_zone_id,omitempty"`   // Cloudflare zone ID for the domain
	IngressIP          string `yaml:"ingress_ip,omitempty"`           // Public IP for DNS A records

	// Auto-TLS (Let's Encrypt)
	IngressAutoTLS   bool   `yaml:"ingress_auto_tls,omitempty"`   // Enable automatic TLS via Let's Encrypt
	IngressCertDir   string `yaml:"ingress_cert_dir,omitempty"`   // Directory for autocert cache
	IngressACMEEmail string `yaml:"ingress_acme_email,omitempty"` // ACME account email for Let's Encrypt

	// Reverse tunnel (expose local containers via *.moltbunker.dev — like ngrok)
	ReverseTunnelEnabled bool   `yaml:"reverse_tunnel_enabled,omitempty"` // Enable reverse tunnel client (provider side)
	ReverseTunnelIngress string `yaml:"reverse_tunnel_ingress,omitempty"` // Ingress address to dial (e.g., "tunnel.moltbunker.dev:9443")

	// Reverse tunnel server (ingress-side — accepts incoming reverse tunnels from providers)
	ReverseTunnelPort     int `yaml:"reverse_tunnel_port,omitempty"`      // Listen port (default: 9443)
	ReverseTunnelMaxConns int `yaml:"reverse_tunnel_max_conns,omitempty"` // Global max connections (default: 10000)
}

// RequesterNodeConfig contains requester-specific configuration
type RequesterNodeConfig struct {
	DefaultNetworkMode   types.NetworkMode `yaml:"default_network_mode"`   // Default network mode for deployments
	DefaultOnionService  bool              `yaml:"default_onion_service"`  // Default to creating onion services
	DefaultEncryption    bool              `yaml:"default_encryption"`     // Default to encrypted volumes
	MaxConcurrentDeploys int               `yaml:"max_concurrent_deploys"` // Max concurrent deployments
	MaxMonthlyBudget     string            `yaml:"max_monthly_budget"`     // Max BUNKER spend per month (0 = unlimited)
	PreferredRegions     []string          `yaml:"preferred_regions"`      // Preferred regions for deployments
	ExcludedRegions      []string          `yaml:"excluded_regions"`       // Excluded regions for compliance
}

// P2PConfig contains P2P network settings
type P2PConfig struct {
	BootstrapNodes         []string `yaml:"bootstrap_nodes"`
	BootstrapHTTPEndpoints []string `yaml:"bootstrap_http_endpoints"` // HTTP(S) URLs for bootstrap peer discovery
	NetworkMode            string   `yaml:"network_mode"`             // clearnet, tor_only, hybrid
	MaxPeers               int      `yaml:"max_peers"`
	DialTimeoutSecs        int      `yaml:"dial_timeout_seconds"`
	DialTimeout            int      `yaml:"dial_timeout"` // Alias for DialTimeoutSecs
	EnableMDNS             bool     `yaml:"enable_mdns"`
	EnableNAT              bool     `yaml:"enable_nat"` // Enable NAT traversal (default: true)
	ExternalIP             string   `yaml:"external_ip"`
	AnnounceAddrs          []string `yaml:"announce_addrs"`
	ConnectionTimeout      int      `yaml:"connection_timeout_seconds"`
	IdleTimeout            int      `yaml:"idle_timeout_seconds"`

	// Circuit breaker settings
	CircuitBreaker CircuitBreakerConfig `yaml:"circuit_breaker"`

	// Stake verification settings
	StakeVerification StakeVerificationConfig `yaml:"stake_verification"`

	// Sybil resistance
	MaxPeersPerSubnet int `yaml:"max_peers_per_subnet"` // Max peers from same /24 (default: 3)

	// Replay protection
	MaxMessageAgeSecs int `yaml:"max_message_age_secs"` // Max message age in seconds (default: 300)
	NonceWindowSecs   int `yaml:"nonce_window_secs"`    // Nonce retention window (default: 600)

	// Peer scoring
	PeerScoring PeerScoringConfig `yaml:"peer_scoring"`
}

// StakeVerificationConfig controls on-chain stake verification for P2P peers.
type StakeVerificationConfig struct {
	Enabled         bool `yaml:"enabled"`           // Enable stake verification (default: true)
	CacheTTLSecs    int  `yaml:"cache_ttl_secs"`    // Positive stake cache TTL (default: 300)
	NegativeTTLSecs int  `yaml:"negative_ttl_secs"` // No-stake cache TTL (default: 120)
	GracePeriodSecs int  `yaml:"grace_period_secs"` // Grace after announce before stake check (default: 30)
}

// DefaultStakeVerificationConfig returns the default stake verification settings.
func DefaultStakeVerificationConfig() StakeVerificationConfig {
	return StakeVerificationConfig{
		Enabled:         true,
		CacheTTLSecs:    300,
		NegativeTTLSecs: 120,
		GracePeriodSecs: 30,
	}
}

// PeerScoringConfig controls behavioral scoring thresholds.
type PeerScoringConfig struct {
	Enabled           bool    `yaml:"enabled"`             // Enable peer scoring (default: true)
	WarnThreshold     float64 `yaml:"warn_threshold"`      // Score below this triggers warnings (default: 0.3)
	ThrottleThreshold float64 `yaml:"throttle_threshold"`  // Score below this reduces rate limits (default: 0.2)
	BanThreshold      float64 `yaml:"ban_threshold"`       // Score below this triggers ban (default: 0.1)
	DecayInterval     int     `yaml:"decay_interval_secs"` // Score decay interval (default: 3600)
}

// DefaultPeerScoringConfig returns the default peer scoring settings.
func DefaultPeerScoringConfig() PeerScoringConfig {
	return PeerScoringConfig{
		Enabled:           true,
		WarnThreshold:     0.3,
		ThrottleThreshold: 0.2,
		BanThreshold:      0.1,
		DecayInterval:     3600,
	}
}

// CircuitBreakerConfig contains circuit breaker settings for P2P connections
type CircuitBreakerConfig struct {
	Enabled             bool `yaml:"enabled"`                // Enable circuit breaker (default: true)
	FailureThreshold    int  `yaml:"failure_threshold"`      // Failures before opening (default: 5)
	SuccessThreshold    int  `yaml:"success_threshold"`      // Successes to close (default: 2)
	TimeoutSecs         int  `yaml:"timeout_secs"`           // Open duration before half-open (default: 30)
	HalfOpenMaxRequests int  `yaml:"half_open_max_requests"` // Max requests in half-open (default: 3)
}

// DefaultCircuitBreakerConfig returns the default circuit breaker configuration
func DefaultCircuitBreakerConfig() CircuitBreakerConfig {
	return CircuitBreakerConfig{
		Enabled:             true,
		FailureThreshold:    5,
		SuccessThreshold:    2,
		TimeoutSecs:         30,
		HalfOpenMaxRequests: 3,
	}
}

// TorConfig contains Tor settings
type TorConfig struct {
	Enabled         bool   `yaml:"enabled"`
	DataDir         string `yaml:"data_dir"`
	SOCKS5Port      int    `yaml:"socks5_port"`
	ControlPort     int    `yaml:"control_port"`
	ExitNodeCountry string `yaml:"exit_node_country"`
	StrictNodes     bool   `yaml:"strict_nodes"`
	CircuitTimeout  int    `yaml:"circuit_timeout_seconds"`
	RotateInterval  int    `yaml:"rotate_interval_seconds"`
}

// RuntimeConfig contains container runtime settings
type RuntimeConfig struct {
	ContainerdSocket string               `yaml:"containerd_socket"`
	Namespace        string               `yaml:"namespace"`
	RuntimeName      string               `yaml:"runtime_name"` // "auto", "io.containerd.runc.v2", "io.containerd.kata.v2", etc.
	Kata             KataConfig           `yaml:"kata"`
	Molt             MoltRuntimeConfig    `yaml:"molt"`
	DefaultResources types.ResourceLimits `yaml:"default_resources"`
	MaxResources     types.ResourceLimits `yaml:"max_resources"` // Maximum allocatable
	LogsDir          string               `yaml:"logs_dir"`
	VolumesDir       string               `yaml:"volumes_dir"`
}

// MoltRuntimeConfig contains Molt (WASM serverless) runtime settings.
type MoltRuntimeConfig struct {
	// Enabled controls whether the Molt WASM runtime is initialized at startup.
	// When false, all Molt API calls return "molt runtime not available".
	Enabled bool `yaml:"enabled"`

	// MemoryLimitMB is the max linear memory each WASM instance can use (default: 256).
	MemoryLimitMB uint32 `yaml:"memory_limit_mb"`

	// TimeoutMs is the max execution time per invocation in milliseconds (default: 30000).
	TimeoutMs int `yaml:"timeout_ms"`

	// MaxInstances is the max concurrent WASM instances across all deployments (default: 100).
	MaxInstances int `yaml:"max_instances"`

	// CacheDir is the directory for wazero's on-disk compilation cache.
	// Empty string defaults to ~/.moltbunker/molt-cache/.
	CacheDir string `yaml:"cache_dir"`

	// MaxCacheEntries is the max number of compiled modules kept in the in-memory LRU cache (default: 256).
	MaxCacheEntries int `yaml:"max_cache_entries"`

	// Host function capability flags (all default false — must be explicitly enabled per deployment)
	HTTPEnabled      bool     `yaml:"http_enabled"`                 // Allow WASM host.http_request
	StorageEnabled   bool     `yaml:"storage_enabled"`              // Allow WASM host.storage_*
	CrawlEnabled     bool     `yaml:"crawl_enabled"`                // Allow WASM host.crawl_page
	HTTPAllowedHosts []string `yaml:"http_allowed_hosts,omitempty"` // Restrict HTTP to these hosts only
	HTTPBlockedHosts []string `yaml:"http_blocked_hosts,omitempty"` // Block HTTP to these hosts

	// JS/TS runtime (Deno worker pool)
	JSRuntime JSRuntimeConfig `yaml:"js_runtime"`
}

// JSRuntimeConfig contains Deno-based JavaScript/TypeScript runtime settings.
type JSRuntimeConfig struct {
	// Enabled controls whether the JS runtime (Deno worker pool) is available.
	Enabled bool `yaml:"enabled"`

	// DenoPath is the path to the Deno binary. Empty defaults to "deno" (found in PATH).
	DenoPath string `yaml:"deno_path"`

	// PoolSize is the number of warm Deno worker processes (default: 10).
	PoolSize int `yaml:"pool_size"`

	// TimeoutMs is the max execution time per JS invocation in milliseconds (default: 30000).
	TimeoutMs int `yaml:"timeout_ms"`

	// MaxMemoryMB is the V8 heap size limit in MB per worker (default: 128).
	MaxMemoryMB int `yaml:"max_memory_mb"`
}

// DefaultJSRuntimeConfig returns sensible defaults for the JS runtime.
func DefaultJSRuntimeConfig() JSRuntimeConfig {
	return JSRuntimeConfig{
		Enabled:     false,
		DenoPath:    "deno",
		PoolSize:    10,
		TimeoutMs:   30000,
		MaxMemoryMB: 128,
	}
}

// KataConfig contains Kata Containers-specific settings
type KataConfig struct {
	// VMMemoryMB is the VM memory overhead in MB (default: 256)
	VMMemoryMB int `yaml:"vm_memory_mb"`
	// VMCPUs is the number of vCPUs for the Kata VM (default: 1)
	VMCPUs int `yaml:"vm_cpus"`
	// KernelPath overrides the Kata kernel image path (default: auto-detect)
	KernelPath string `yaml:"kernel_path,omitempty"`
	// ImagePath overrides the Kata rootfs/initrd path (default: auto-detect)
	ImagePath string `yaml:"image_path,omitempty"`
}

// SecurityConfig contains security and container opacity settings
type SecurityConfig struct {
	// Container security profile
	ContainerProfile types.ContainerSecurityProfile `yaml:"container_profile"`

	// Verification
	Verification types.VerificationConfig `yaml:"verification"`

	// TLS settings
	TLSMinVersion   string   `yaml:"tls_min_version"`   // "1.2" or "1.3"
	TLSCipherSuites []string `yaml:"tls_cipher_suites"` // Allowed cipher suites
	CertPinning     bool     `yaml:"cert_pinning"`      // Enable certificate pinning
	MutualTLS       bool     `yaml:"mutual_tls"`        // Require client certificates

	// R4 — image vulnerability scanning. When true AND the trivy binary is on
	// PATH, deploys are scanned (block HIGH/CRITICAL). Default false: a no-op
	// scanner is used so a host without trivy never fails a deploy.
	ImageScanEnabled bool `yaml:"image_scan_enabled"`

	// R8 — state-at-rest encryption. The zero value (false) means ENABLED, so
	// existing configs get bbolt value encryption by default. Set to true only to
	// opt out (e.g. for debugging the raw database). The key lives at
	// DataDir/state.key with 0600 perms; see internal/state/statekey.go for the
	// threat model.
	StateEncryptionDisabled bool `yaml:"state_encryption_disabled"`

	// R5 — image content encryption at rest (ocicrypt/X25519). The zero value
	// (false) means DISABLED: this is opt-in, not opt-out like R8, because it
	// rewrites the at-rest image representation in containerd's content store and
	// needs real-containerd validation (Linux/Colima runtime CI, R11) plus, for
	// the multi-replica recipient model, per-node X25519 key advertisement.
	// When true, each provider encrypts the images it pulls to its own stable
	// X25519 key (DataDir/provider_x25519.key) so the at-rest layer blobs are
	// ciphertext; decryption happens in-process just before unpack. See
	// internal/runtime/image_encrypt.go for the threat model and boundaries.
	ImageEncryptionEnabled bool `yaml:"image_encryption_enabled"`

	// KEY-01 — at-rest state key rotation sweep. When true, the daemon runs an
	// idempotent RotateKey pass over the bbolt store on startup, re-encrypting
	// every value with the current state key and bumping the on-disk magic to
	// MBENC2. Default false (lazy migration on next write). Use after a key
	// change or to eagerly retag a legacy (MBENC1) database.
	StateKeyRotationSweep bool `yaml:"state_key_rotation_sweep"`
}

// RedundancyConfig contains redundancy settings
type RedundancyConfig struct {
	ReplicaCount        int                `yaml:"replica_count"` // Always 3 for production
	HealthCheckInterval int                `yaml:"health_check_interval_seconds"`
	HealthTimeout       int                `yaml:"health_timeout_seconds"`
	FailoverDelay       int                `yaml:"failover_delay_seconds"`
	RegionConfig        types.RegionConfig `yaml:"region_config"`
}

// EconomicsConfig contains all economic parameters
type EconomicsConfig struct {
	// Token settings
	TokenAddress  string `yaml:"token_address"`  // BUNKER token contract address
	TokenDecimals int    `yaml:"token_decimals"` // Token decimals (18)

	// Contract addresses
	RegistryAddress          string `yaml:"registry_address"`           // Provider registry contract
	StakingAddress           string `yaml:"staking_address"`            // BunkerStaking contract
	EscrowAddress            string `yaml:"escrow_address"`             // Payment escrow contract
	SlashingAddress          string `yaml:"slashing_address"`           // Slashing contract
	PricingAddress           string `yaml:"pricing_address"`            // Pricing oracle contract
	GovernanceAddress        string `yaml:"governance_address"`         // Governance contract
	DelegationAddress        string `yaml:"delegation_address"`         // BunkerDelegation contract
	ReputationAddress        string `yaml:"reputation_address"`         // BunkerReputation contract
	VerificationAddress      string `yaml:"verification_address"`       // BunkerVerification contract
	SubdomainRegistryAddress string `yaml:"subdomain_registry_address"` // BunkerRegistry contract

	// Payment mode
	MockPayments bool `yaml:"mock_payments"` // Use mock payment layer (default: true for dev)

	// Configuration (loaded from contracts or overridden locally)
	StakingTiers map[types.StakingTier]*types.StakingTierConfig `yaml:"staking_tiers"`
	Pricing      *types.PricingConfig                           `yaml:"pricing"`
	Slashing     *types.SlashingConfig                          `yaml:"slashing"`
	Reputation   *types.ReputationConfig                        `yaml:"reputation"`
	Fairness     *types.FairnessConfig                          `yaml:"fairness"`
	ProtocolFees *types.ProtocolFees                            `yaml:"protocol_fees"`
	Unstaking    *types.UnstakingConfig                         `yaml:"unstaking"`

	// Chain settings
	ChainID            int64    `yaml:"chain_id"`            // Base chain ID
	RPCURL             string   `yaml:"rpc_url"`             // Primary RPC endpoint
	RPCURLs            []string `yaml:"rpc_urls"`            // Additional RPC endpoints for failover
	WSEndpoint         string   `yaml:"ws_endpoint"`         // Primary WebSocket endpoint
	WSEndpoints        []string `yaml:"ws_endpoints"`        // Additional WS endpoints for failover
	BlockConfirmations int      `yaml:"block_confirmations"` // Required confirmations
}

// ResolvedRPCURLs merges the single RPCURL with the RPCURLs list, deduplicating.
// The single URL is placed first as the primary.
func (ec *EconomicsConfig) ResolvedRPCURLs() []string {
	return mergeURLs(ec.RPCURL, ec.RPCURLs)
}

// ResolvedWSEndpoints merges the single WSEndpoint with the WSEndpoints list, deduplicating.
func (ec *EconomicsConfig) ResolvedWSEndpoints() []string {
	return mergeURLs(ec.WSEndpoint, ec.WSEndpoints)
}

// mergeURLs combines a primary URL with a list, deduplicating and preserving order.
func mergeURLs(primary string, extras []string) []string {
	seen := make(map[string]bool)
	var result []string

	if primary != "" {
		result = append(result, primary)
		seen[primary] = true
	}
	for _, u := range extras {
		if u != "" && !seen[u] {
			result = append(result, u)
			seen[u] = true
		}
	}
	return result
}

// EncryptionConfig contains encryption settings
type EncryptionConfig struct {
	// Volume encryption
	VolumeEncryptionEnabled bool   `yaml:"volume_encryption_enabled"`
	VolumeEncryptionCipher  string `yaml:"volume_encryption_cipher"`
	VolumeKeySize           int    `yaml:"volume_key_size"`

	// Data encryption
	DataEncryptionAlgorithm string `yaml:"data_encryption_algorithm"`
	KeyDerivationFunction   string `yaml:"key_derivation_function"`

	// Key exchange
	KeyExchangeAlgorithm string `yaml:"key_exchange_algorithm"`

	// Key storage
	KeyStorePath    string `yaml:"key_store_path"`
	KeyStoreEncrypt bool   `yaml:"key_store_encrypt"`
}

// StorageConfig contains object storage service settings.
type StorageConfig struct {
	Enabled       bool   `yaml:"enabled"`         // Enable object storage service
	DataDir       string `yaml:"data_dir"`        // Directory for object blobs (default: <data_dir>/storage)
	MaxBuckets    int    `yaml:"max_buckets"`     // Max buckets per wallet (default: 100)
	MaxObjectSize int64  `yaml:"max_object_size"` // Max single object size in bytes (default: 5GB)
	S3Port        int    `yaml:"s3_port"`         // S3-compatible API port (default: 9300)
	EnableS3      bool   `yaml:"enable_s3"`       // Enable S3-compatible endpoint

	// KEY-01 — at-rest object encryption. When true (and the provider X25519
	// key is available), object blobs are encrypted with a per-object DEK sealed
	// to the provider's own X25519 key (self-recipient model, matching R5 image
	// encryption). Default false: blobs are stored as plaintext. Objects written
	// before this was enabled remain readable (back-compat read path).
	EncryptionEnabled bool `yaml:"encryption_enabled"`

	// KEY-01 — in-memory ceiling for the (non-streaming) encrypted object path.
	// When EncryptionEnabled is true, encrypted PutObject/GetObject buffer the
	// whole object in memory to seal/open it, so an object larger than this is
	// rejected to avoid OOM. Only consulted when encryption is enabled. Zero
	// falls back to the engine default (256MB). A streaming/chunked AEAD that
	// removes this ceiling is a follow-up.
	EncryptedMaxInMemoryBytes int64 `yaml:"encrypted_max_in_memory_bytes"`
}

// DefaultStorageConfig returns the default storage configuration.
func DefaultStorageConfig() StorageConfig {
	return StorageConfig{
		Enabled:       false,
		DataDir:       "", // resolved at runtime
		MaxBuckets:    100,
		MaxObjectSize: 5 * 1024 * 1024 * 1024, // 5GB
		S3Port:        9300,
		EnableS3:      false,
	}
}

// ProxyConfig contains decentralized proxy service settings.
type ProxyConfig struct {
	Enabled     bool   `yaml:"enabled"`      // Enable proxy service
	SOCKS5Addr  string `yaml:"socks5_addr"`  // SOCKS5 listen address (default: :1080)
	HTTPAddr    string `yaml:"http_addr"`    // HTTP proxy listen address (default: :8118)
	UseTor      bool   `yaml:"use_tor"`      // Route through Tor by default
	MaxSessions int    `yaml:"max_sessions"` // Max concurrent proxy sessions (default: 1000)
}

// DefaultProxyConfig returns the default proxy configuration.
func DefaultProxyConfig() ProxyConfig {
	return ProxyConfig{
		Enabled:     false,
		SOCKS5Addr:  ":1080",
		HTTPAddr:    ":8118",
		UseTor:      false,
		MaxSessions: 1000,
	}
}

// AgentConfig contains AI agent runtime settings.
type AgentConfig struct {
	Enabled            bool     `yaml:"enabled"`               // Enable agent runtime
	Frameworks         []string `yaml:"frameworks"`            // Enabled frameworks: langgraph, crewai, autogen, custom
	DefaultMemoryMB    int      `yaml:"default_memory_mb"`     // Default agent container memory (default: 2048)
	MaxAgentsPerWallet int      `yaml:"max_agents_per_wallet"` // Max concurrent agents per wallet (default: 10)
	SyncIntervalSecs   int      `yaml:"sync_interval_secs"`    // Memory sync interval (default: 60)
}

// DefaultAgentConfig returns the default agent configuration.
func DefaultAgentConfig() AgentConfig {
	return AgentConfig{
		Enabled:            false,
		Frameworks:         []string{"langgraph", "crewai", "autogen", "custom"},
		DefaultMemoryMB:    2048,
		MaxAgentsPerWallet: 10,
		SyncIntervalSecs:   60,
	}
}

// CrawlConfig contains web crawling service settings.
type CrawlConfig struct {
	Enabled       bool   `yaml:"enabled"`          // Enable crawl service
	MaxDepth      int    `yaml:"max_depth"`        // Default max crawl depth (default: 3)
	MaxPages      int    `yaml:"max_pages"`        // Max pages per job (default: 1000)
	MaxConcurrent int    `yaml:"max_concurrent"`   // Max concurrent crawl workers (default: 10)
	BrowserImage  string `yaml:"browser_image"`    // Chromium container image CID
	RespectRobots bool   `yaml:"respect_robots"`   // Respect robots.txt (default: true)
	DefaultDelay  int    `yaml:"default_delay_ms"` // Default per-domain delay in ms (default: 1000)
}

// DefaultCrawlConfig returns the default crawl configuration.
func DefaultCrawlConfig() CrawlConfig {
	return CrawlConfig{
		Enabled:       false,
		MaxDepth:      3,
		MaxPages:      1000,
		MaxConcurrent: 10,
		BrowserImage:  "", // set after image is published to IPFS
		RespectRobots: true,
		DefaultDelay:  1000,
	}
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()
	dataDir := filepath.Join(homeDir, ".moltbunker")

	return &Config{
		Daemon: DaemonConfig{
			Port:        9000,
			DataDir:     dataDir,
			KeyPath:     filepath.Join(dataDir, "keys", "node.key"),
			KeystoreDir: filepath.Join(dataDir, "keystore"),
			SocketPath:  filepath.Join(dataDir, "daemon.sock"),
			LogLevel:    "info",
			LogFormat:   "text",
		},
		API: DefaultAPIConfig(),
		Node: NodeConfig{
			Role:          types.NodeRoleRequester, // Default to requester; opt-in to provider
			WalletAddress: "",
			WalletKeyFile: filepath.Join(dataDir, "keystore", "wallet.json"),
			AutoRegister:  false,
			Provider: func() ProviderNodeConfig {
				hw := detectHardwareProfile()
				return ProviderNodeConfig{
					DeclaredCPU:       hw.CPUCores,
					DeclaredMemoryGB:  hw.MemoryGB,
					DeclaredStorageGB: hw.StorageGB,
					DeclaredBandwidth: hw.BandwidthMbps,
					Hardware:          hw,
					GPUEnabled:        false,
					TargetTier:        types.StakingTierStarter,
					AutoStake:         false,
					AcceptServices:    true,
					AcceptJobs:        true,
					AcceptScheduled:   true,
					AcceptFunctions:   false, // Disabled by default
					MaintenanceMode:   false,
				}
			}(),
			Requester: RequesterNodeConfig{
				DefaultNetworkMode:   types.NetworkModeHybrid,
				DefaultOnionService:  false,
				DefaultEncryption:    true,
				MaxConcurrentDeploys: 10,
				MaxMonthlyBudget:     "0", // Unlimited
				PreferredRegions:     []string{},
				ExcludedRegions:      []string{},
			},
		},
		P2P: P2PConfig{
			BootstrapNodes:         []string{},
			BootstrapHTTPEndpoints: []string{"https://api.moltbunker.com/v1/bootstrap"},
			NetworkMode:            "hybrid",
			MaxPeers:               50,
			DialTimeoutSecs:        30,
			DialTimeout:            30,
			EnableMDNS:             true,
			EnableNAT:              true,
			ExternalIP:             "",
			AnnounceAddrs:          []string{},
			ConnectionTimeout:      60,
			IdleTimeout:            300,
			CircuitBreaker:         DefaultCircuitBreakerConfig(),
			StakeVerification:      DefaultStakeVerificationConfig(),
			MaxPeersPerSubnet:      3,
			MaxMessageAgeSecs:      300,
			NonceWindowSecs:        600,
			PeerScoring:            DefaultPeerScoringConfig(),
		},
		Tor: TorConfig{
			Enabled:         false,
			DataDir:         filepath.Join(dataDir, "tor"),
			SOCKS5Port:      9050,
			ControlPort:     9051,
			ExitNodeCountry: "",
			StrictNodes:     false,
			CircuitTimeout:  30,
			RotateInterval:  600,
		},
		Runtime: RuntimeConfig{
			ContainerdSocket: DetectContainerdSocket(),
			Namespace:        "moltbunker",
			RuntimeName:      "auto",
			Kata: KataConfig{
				VMMemoryMB: 256,
				VMCPUs:     1,
			},
			Molt: MoltRuntimeConfig{
				Enabled:         false, // Opt-in: provider must explicitly enable
				MemoryLimitMB:   256,
				TimeoutMs:       30000,
				MaxInstances:    100,
				CacheDir:        "", // resolved at runtime to ~/.moltbunker/molt-cache/
				MaxCacheEntries: 256,
				JSRuntime:       DefaultJSRuntimeConfig(),
			},
			DefaultResources: types.ResourceLimits{
				CPUQuota:    100000,
				CPUPeriod:   100000,
				MemoryLimit: 1024 * 1024 * 1024,      // 1GB
				DiskLimit:   10 * 1024 * 1024 * 1024, // 10GB
				NetworkBW:   10 * 1024 * 1024,        // 10MB/s
				PIDLimit:    100,
			},
			MaxResources: types.ResourceLimits{
				CPUQuota:    1600000, // 16 cores max
				CPUPeriod:   100000,
				MemoryLimit: 64 * 1024 * 1024 * 1024,   // 64GB
				DiskLimit:   1000 * 1024 * 1024 * 1024, // 1TB
				NetworkBW:   1000 * 1024 * 1024,        // 1GB/s
				PIDLimit:    10000,
			},
			LogsDir:    filepath.Join(dataDir, "logs"),
			VolumesDir: filepath.Join(dataDir, "volumes"),
		},
		Security: SecurityConfig{
			ContainerProfile: *types.DefaultContainerSecurityProfile(),
			Verification:     *types.DefaultVerificationConfig(),
			TLSMinVersion:    "1.3",
			TLSCipherSuites: []string{
				"TLS_AES_256_GCM_SHA384",
				"TLS_CHACHA20_POLY1305_SHA256",
				"TLS_AES_128_GCM_SHA256",
			},
			CertPinning: true,
			MutualTLS:   true,
		},
		Redundancy: RedundancyConfig{
			ReplicaCount:        3,
			HealthCheckInterval: 10,
			HealthTimeout:       30,
			FailoverDelay:       60,
			RegionConfig:        *types.DefaultRegionConfig(),
		},
		Economics: EconomicsConfig{
			TokenAddress:       "", // Set after deployment
			TokenDecimals:      18,
			RegistryAddress:    "",
			StakingAddress:     "",
			EscrowAddress:      "",
			SlashingAddress:    "",
			PricingAddress:     "",
			GovernanceAddress:  "",
			MockPayments:       true, // Mock mode by default for development
			StakingTiers:       types.DefaultStakingTiers(),
			Pricing:            types.DefaultPricingConfig(),
			Slashing:           types.DefaultSlashingConfig(),
			Reputation:         types.DefaultReputationConfig(),
			Fairness:           types.DefaultFairnessConfig(),
			ProtocolFees:       types.DefaultProtocolFees(),
			Unstaking:          types.DefaultUnstakingConfig(),
			ChainID:            8453, // Base mainnet
			RPCURL:             "https://mainnet.base.org",
			BlockConfirmations: 12,
		},
		Encryption: EncryptionConfig{
			VolumeEncryptionEnabled: true,
			VolumeEncryptionCipher:  "aes-xts-plain64",
			VolumeKeySize:           256,
			DataEncryptionAlgorithm: "AES-256-GCM",
			KeyDerivationFunction:   "HKDF-SHA256",
			KeyExchangeAlgorithm:    "X25519",
			KeyStorePath:            filepath.Join(dataDir, "keys", "deployments"),
			KeyStoreEncrypt:         true,
		},

		// P0 services (all disabled by default)
		Storage: func() StorageConfig {
			cfg := DefaultStorageConfig()
			cfg.DataDir = filepath.Join(dataDir, "storage")
			return cfg
		}(),
		Proxy: DefaultProxyConfig(),
		Agent: DefaultAgentConfig(),
		Crawl: DefaultCrawlConfig(),
	}
}

// Load loads configuration from file
func Load(path string) (*Config, error) {
	cfg := DefaultConfig()
	path = expandPath(path)

	// #nosec G304 -- path is the operator-supplied config file path (CLI flag / default), not request input
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	cfg.expandPaths()

	// Parse staking tier min stakes
	for _, tier := range cfg.Economics.StakingTiers {
		if err := tier.ParseMinStake(); err != nil {
			return nil, fmt.Errorf("invalid staking tier config: %w", err)
		}
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// Save saves configuration to file
func (c *Config) Save(path string) error {
	path = expandPath(path)

	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Daemon validation
	if c.Daemon.Port < 1 || c.Daemon.Port > 65535 {
		return fmt.Errorf("invalid port: %d", c.Daemon.Port)
	}

	// Node role validation
	if !c.Node.Role.IsValid() {
		return fmt.Errorf("invalid node role: %s", c.Node.Role)
	}

	// P2P validation
	if c.P2P.MaxPeers < 1 {
		return fmt.Errorf("max_peers must be at least 1")
	}
	if c.P2P.NetworkMode != "clearnet" && c.P2P.NetworkMode != "tor_only" && c.P2P.NetworkMode != "hybrid" {
		return fmt.Errorf("invalid network_mode: %s", c.P2P.NetworkMode)
	}

	// P2P subnet limiter validation
	if c.P2P.MaxPeersPerSubnet < 0 || c.P2P.MaxPeersPerSubnet > 100 {
		return fmt.Errorf("max_peers_per_subnet must be between 0 and 100, got %d", c.P2P.MaxPeersPerSubnet)
	}

	// Redundancy validation
	if c.Redundancy.ReplicaCount < 1 || c.Redundancy.ReplicaCount > 10 {
		return fmt.Errorf("replica_count must be between 1 and 10")
	}

	// Economics validation
	if c.Economics.TokenDecimals < 0 || c.Economics.TokenDecimals > 18 {
		return fmt.Errorf("invalid token_decimals: %d", c.Economics.TokenDecimals)
	}

	// Security validation
	if c.Security.TLSMinVersion != "1.2" && c.Security.TLSMinVersion != "1.3" {
		return fmt.Errorf("invalid tls_min_version: %s", c.Security.TLSMinVersion)
	}

	// Contract address validation (only when using real payments)
	if !c.Economics.MockPayments {
		addrs := map[string]string{
			"token_address":        c.Economics.TokenAddress,
			"staking_address":      c.Economics.StakingAddress,
			"escrow_address":       c.Economics.EscrowAddress,
			"pricing_address":      c.Economics.PricingAddress,
			"delegation_address":   c.Economics.DelegationAddress,
			"reputation_address":   c.Economics.ReputationAddress,
			"verification_address": c.Economics.VerificationAddress,
		}
		for name, addr := range addrs {
			if err := validateEthAddress(name, addr); err != nil {
				return err
			}
		}
	}

	return nil
}

// validateEthAddress checks that an Ethereum address is 0x-prefixed, 40 hex chars, and non-zero.
func validateEthAddress(name, addr string) error {
	if addr == "" {
		return fmt.Errorf("%s is required when mock_payments is false", name)
	}
	if !strings.HasPrefix(addr, "0x") && !strings.HasPrefix(addr, "0X") {
		return fmt.Errorf("%s must start with 0x, got %q", name, addr)
	}
	hexPart := addr[2:]
	if len(hexPart) != 40 {
		return fmt.Errorf("%s must be 42 characters (0x + 40 hex), got %d", name, len(addr))
	}
	if _, err := hex.DecodeString(hexPart); err != nil {
		return fmt.Errorf("%s contains invalid hex characters: %w", name, err)
	}
	// Check for zero address
	allZero := true
	for _, c := range hexPart {
		if c != '0' {
			allZero = false
			break
		}
	}
	if allZero {
		return fmt.Errorf("%s must not be the zero address", name)
	}
	return nil
}

// expandPaths expands ~ in all path fields
func (c *Config) expandPaths() {
	c.Daemon.DataDir = expandPath(c.Daemon.DataDir)
	c.Daemon.KeyPath = expandPath(c.Daemon.KeyPath)
	c.Daemon.KeystoreDir = expandPath(c.Daemon.KeystoreDir)
	c.Daemon.SocketPath = expandPath(c.Daemon.SocketPath)
	c.Tor.DataDir = expandPath(c.Tor.DataDir)
	c.Runtime.LogsDir = expandPath(c.Runtime.LogsDir)
	c.Runtime.VolumesDir = expandPath(c.Runtime.VolumesDir)
	c.Runtime.Molt.CacheDir = expandPath(c.Runtime.Molt.CacheDir)
	c.Encryption.KeyStorePath = expandPath(c.Encryption.KeyStorePath)
	c.Node.WalletKeyFile = expandPath(c.Node.WalletKeyFile)
	c.Node.WalletPasswordFile = expandPath(c.Node.WalletPasswordFile)
	c.Storage.DataDir = expandPath(c.Storage.DataDir)
}

// expandPath expands ~ to home directory
func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(homeDir, path[2:])
	}
	return path
}

// DefaultConfigPath returns the default config file path
func DefaultConfigPath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".moltbunker", "config.yaml")
}

// DetectContainerdSocket returns the best containerd socket path for the current platform
func DetectContainerdSocket() string {
	if runtime.GOOS == "darwin" {
		// Check for Colima socket (preferred on macOS)
		homeDir, _ := os.UserHomeDir()

		// Check for Colima docker socket (most common)
		colimaDockerSocket := filepath.Join(homeDir, ".colima", "default", "docker.sock")
		if _, err := os.Stat(colimaDockerSocket); err == nil {
			return colimaDockerSocket
		}

		// Alternative: Colima containerd socket path
		colimaContainerdSocket := filepath.Join(homeDir, ".colima", "default", "containerd.sock")
		if _, err := os.Stat(colimaContainerdSocket); err == nil {
			return colimaContainerdSocket
		}

		// Fallback: Docker Desktop socket on macOS
		if _, err := os.Stat("/var/run/docker.sock"); err == nil {
			return "/var/run/docker.sock"
		}
	}

	// Default Linux socket
	return "/run/containerd/containerd.sock"
}

// EnsureDirectories creates all necessary directories
func (c *Config) EnsureDirectories() error {
	dirs := []string{
		c.Daemon.DataDir,
		filepath.Dir(c.Daemon.KeyPath),
		c.Daemon.KeystoreDir,
		filepath.Dir(c.Daemon.SocketPath),
		c.Tor.DataDir,
		c.Runtime.LogsDir,
		c.Runtime.VolumesDir,
		c.Encryption.KeyStorePath,
		filepath.Join(c.Daemon.DataDir, "containers"),
		filepath.Join(c.Daemon.DataDir, "state"),
	}

	for _, dir := range dirs {
		if dir == "" {
			continue
		}
		// 0700: these dirs hold keys, keystore, state, and volumes (private)
		if err := os.MkdirAll(dir, 0700); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// IsProvider returns true if the node can act as a provider
func (c *Config) IsProvider() bool {
	return c.Node.Role == types.NodeRoleProvider || c.Node.Role == types.NodeRoleHybrid
}

// IsRequester returns true if the node can act as a requester
func (c *Config) IsRequester() bool {
	return c.Node.Role == types.NodeRoleRequester || c.Node.Role == types.NodeRoleHybrid
}

// GetTierConfig returns the staking tier configuration for a given tier
func (c *Config) GetTierConfig(tier types.StakingTier) (*types.StakingTierConfig, bool) {
	config, exists := c.Economics.StakingTiers[tier]
	return config, exists
}

// ── System resource detection ────────────────────────────────────────────────

// detectMemoryGB and detectStorageGB are in platform-specific files:
//   config_sysinfo_darwin.go  (macOS)
//   config_sysinfo_linux.go   (Linux)
