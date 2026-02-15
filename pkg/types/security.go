package types

import (
	"time"
)

// ContainerSecurityProfile defines hardened container security settings
type ContainerSecurityProfile struct {
	// Interactive access controls
	DisableExec   bool `yaml:"disable_exec" json:"disable_exec"`
	DisableAttach bool `yaml:"disable_attach" json:"disable_attach"`
	DisableShell  bool `yaml:"disable_shell" json:"disable_shell"`

	// Capabilities
	DropAllCapabilities bool     `yaml:"drop_all_capabilities" json:"drop_all_capabilities"`
	AddCapabilities     []string `yaml:"add_capabilities" json:"add_capabilities"`

	// Namespace isolation
	UserNamespace    bool `yaml:"user_namespace" json:"user_namespace"`
	PIDNamespace     bool `yaml:"pid_namespace" json:"pid_namespace"`
	NetworkNamespace bool `yaml:"network_namespace" json:"network_namespace"`
	IPCNamespace     bool `yaml:"ipc_namespace" json:"ipc_namespace"`

	// Filesystem
	ReadOnlyRoot    bool     `yaml:"read_only_root" json:"read_only_root"`
	NoNewPrivileges bool     `yaml:"no_new_privileges" json:"no_new_privileges"`
	MaskPaths       []string `yaml:"mask_paths" json:"mask_paths"`
	ReadOnlyPaths   []string `yaml:"readonly_paths" json:"readonly_paths"`

	// Seccomp
	SeccompProfile   string   `yaml:"seccomp_profile" json:"seccomp_profile"` // "strict", "default", "unconfined"
	BlockedSyscalls  []string `yaml:"blocked_syscalls" json:"blocked_syscalls"`
	AllowedSyscalls  []string `yaml:"allowed_syscalls" json:"allowed_syscalls"` // Whitelist mode

	// AppArmor/SELinux
	AppArmorProfile string `yaml:"apparmor_profile" json:"apparmor_profile"`
	SELinuxLabel    string `yaml:"selinux_label" json:"selinux_label"`

	// Resource limits for DoS prevention
	Ulimits UlimitConfig `yaml:"ulimits" json:"ulimits"`

	// Network restrictions
	AllowedOutboundPorts []int  `yaml:"allowed_outbound_ports" json:"allowed_outbound_ports"` // Empty = all allowed
	AllowedInboundPorts  []int  `yaml:"allowed_inbound_ports" json:"allowed_inbound_ports"`   // Empty = all allowed
	DNSPolicy            string `yaml:"dns_policy" json:"dns_policy"`                         // "default", "none", "custom"
	CustomDNS            []string `yaml:"custom_dns" json:"custom_dns"`
}

// UlimitConfig defines ulimit settings
type UlimitConfig struct {
	NoFile  int64 `yaml:"nofile" json:"nofile"`   // Max open files
	NProc   int64 `yaml:"nproc" json:"nproc"`     // Max processes
	MemLock int64 `yaml:"memlock" json:"memlock"` // Max locked memory (0 = none)
	Core    int64 `yaml:"core" json:"core"`       // Core file size (0 = no core dumps)
	Stack   int64 `yaml:"stack" json:"stack"`     // Stack size in bytes
}

// DefaultContainerSecurityProfile returns the default hardened security profile
func DefaultContainerSecurityProfile() *ContainerSecurityProfile {
	return &ContainerSecurityProfile{
		// Disable all interactive access
		DisableExec:   true,
		DisableAttach: true,
		DisableShell:  true,

		// Drop all capabilities
		DropAllCapabilities: true,
		AddCapabilities:     []string{}, // None added

		// Full namespace isolation
		UserNamespace:    true,
		PIDNamespace:     true,
		NetworkNamespace: true,
		IPCNamespace:     true,

		// Filesystem hardening
		ReadOnlyRoot:    true,
		NoNewPrivileges: true,
		MaskPaths: []string{
			"/proc/kcore",
			"/proc/kmem",
			"/proc/mem",
			"/proc/acpi",
			"/proc/scsi",
			"/proc/keys",
			"/proc/latency_stats",
			"/proc/timer_list",
			"/proc/timer_stats",
			"/proc/sched_debug",
			"/sys/firmware",
			"/sys/fs/selinux",
		},
		ReadOnlyPaths: []string{
			"/proc/bus",
			"/proc/fs",
			"/proc/irq",
			"/proc/sys",
			"/proc/sysrq-trigger",
		},

		// Strict seccomp
		SeccompProfile: "strict",
		BlockedSyscalls: []string{
			"ptrace",
			"process_vm_readv",
			"process_vm_writev",
			"personality",
			"mount",
			"umount2",
			"pivot_root",
			"keyctl",
			"request_key",
			"add_key",
			"kexec_load",
			"kexec_file_load",
			"init_module",
			"finit_module",
			"delete_module",
			"acct",
			"quotactl",
			"reboot",
			"swapon",
			"swapoff",
			"sethostname",
			"setdomainname",
			"iopl",
			"ioperm",
			"create_module",
			"get_kernel_syms",
			"query_module",
			"uselib",
			"nfsservctl",
			"getpmsg",
			"putpmsg",
			"afs_syscall",
			"tuxcall",
			"security",
			"lookup_dcookie",
			"perf_event_open",
			"bpf",
		},
		AllowedSyscalls: []string{}, // Empty = use default allow list

		// AppArmor
		AppArmorProfile: "moltbunker-container",
		SELinuxLabel:    "",

		// Ulimits
		Ulimits: UlimitConfig{
			NoFile:  1024,
			NProc:   100,
			MemLock: 0,     // No locked memory
			Core:    0,     // No core dumps
			Stack:   8388608, // 8MB stack
		},

		// Network (all allowed by default, can be restricted)
		AllowedOutboundPorts: []int{},
		AllowedInboundPorts:  []int{},
		DNSPolicy:            "default",
		CustomDNS:            []string{},
	}
}

// DeploymentSecurityProfile returns a production-practical security profile
// for containers deployed through the network. It follows Docker's default
// security model: drop all capabilities then add back the safe default set,
// use a blocklist seccomp profile (block dangerous syscalls, allow everything
// else), mask sensitive /proc and /sys paths, and set practical ulimits.
//
// This profile is designed to work with real-world container images while
// still providing meaningful security hardening over a completely unconfined
// container.
func DeploymentSecurityProfile() *ContainerSecurityProfile {
	return &ContainerSecurityProfile{
		// Allow interactive access — wallet auth + deployment ownership gates access
		DisableExec:   false,
		DisableAttach: false,
		DisableShell:  false,

		// Drop all capabilities, add back Docker's default safe set
		DropAllCapabilities: true,
		AddCapabilities: []string{
			"CAP_CHOWN",
			"CAP_DAC_OVERRIDE",
			"CAP_FSETID",
			"CAP_FOWNER",
			"CAP_MKNOD",
			"CAP_NET_RAW",
			"CAP_SETGID",
			"CAP_SETUID",
			"CAP_SETFCAP",
			"CAP_SETPCAP",
			"CAP_NET_BIND_SERVICE",
			"CAP_SYS_CHROOT",
			"CAP_KILL",
			"CAP_AUDIT_WRITE",
		},

		// No user namespace — UID 0→65534 mapping breaks most images
		UserNamespace:    false,
		PIDNamespace:     true,
		NetworkNamespace: true,
		IPCNamespace:     true,

		// Filesystem — no read-only root (many images need writable root)
		ReadOnlyRoot:    false,
		NoNewPrivileges: true,
		MaskPaths: []string{
			"/proc/kcore",
			"/proc/kmem",
			"/proc/mem",
			"/proc/acpi",
			"/proc/scsi",
			"/proc/keys",
			"/proc/latency_stats",
			"/proc/timer_list",
			"/proc/timer_stats",
			"/proc/sched_debug",
			"/sys/firmware",
			"/sys/fs/selinux",
		},
		ReadOnlyPaths: []string{
			"/proc/bus",
			"/proc/fs",
			"/proc/irq",
			"/proc/sys",
			"/proc/sysrq-trigger",
		},

		// Blocklist seccomp — block dangerous syscalls, allow everything else
		SeccompProfile: "default",
		BlockedSyscalls: []string{
			"ptrace",
			"process_vm_readv",
			"process_vm_writev",
			"personality",
			"mount",
			"umount2",
			"pivot_root",
			"keyctl",
			"request_key",
			"add_key",
			"kexec_load",
			"kexec_file_load",
			"init_module",
			"finit_module",
			"delete_module",
			"acct",
			"quotactl",
			"reboot",
			"swapon",
			"swapoff",
			"sethostname",
			"setdomainname",
			"iopl",
			"ioperm",
			"create_module",
			"get_kernel_syms",
			"query_module",
			"uselib",
			"nfsservctl",
			"getpmsg",
			"putpmsg",
			"afs_syscall",
			"tuxcall",
			"security",
			"lookup_dcookie",
			"perf_event_open",
			"bpf",
			"userfaultfd",
			"unshare",
			"setns",
			"clock_settime",
			"clock_adjtime",
			"settimeofday",
			"syslog",
			"vhangup",
			"open_by_handle_at",
			"move_pages",
			"kcmp",
			"clone3",
			"move_mount",
			"open_tree",
			"fsopen",
			"fspick",
			"fsconfig",
			"fsmount",
		},
		AllowedSyscalls: []string{}, // Unused in "default" mode

		// No AppArmor by default — applied conditionally at runtime
		AppArmorProfile: "",
		SELinuxLabel:    "",

		// Practical ulimits
		Ulimits: UlimitConfig{
			NoFile:  65536,
			NProc:   4096,
			MemLock: 0,       // No locked memory
			Core:    0,       // No core dumps
			Stack:   8388608, // 8MB stack
		},

		// Network (all allowed by default)
		AllowedOutboundPorts: []int{},
		AllowedInboundPorts:  []int{},
		DNSPolicy:            "default",
		CustomDNS:            []string{},
	}
}

// VerificationConfig defines multi-party verification settings
type VerificationConfig struct {
	// Heartbeat settings
	HeartbeatIntervalSeconds int `yaml:"heartbeat_interval_seconds" json:"heartbeat_interval_seconds"`
	HeartbeatTimeoutSeconds  int `yaml:"heartbeat_timeout_seconds" json:"heartbeat_timeout_seconds"`
	MissedHeartbeatsBeforeAlert int `yaml:"missed_heartbeats_before_alert" json:"missed_heartbeats_before_alert"`

	// Cross-replica verification
	ConsensusRequired        int  `yaml:"consensus_required" json:"consensus_required"`         // Out of 3 (e.g., 2)
	OutputHashVerification   bool `yaml:"output_hash_verification" json:"output_hash_verification"`
	DeterministicJobsOnly    bool `yaml:"deterministic_jobs_only" json:"deterministic_jobs_only"` // Only verify deterministic jobs

	// Canary probes
	EnableCanaryProbes       bool `yaml:"enable_canary_probes" json:"enable_canary_probes"`
	CanaryProbeIntervalMins  int  `yaml:"canary_probe_interval_mins" json:"canary_probe_interval_mins"`
	CanaryFailureThreshold   int  `yaml:"canary_failure_threshold" json:"canary_failure_threshold"`

	// Resource attestation
	ResourceAttestationEnabled   bool `yaml:"resource_attestation_enabled" json:"resource_attestation_enabled"`
	ResourceMismatchThreshold    float64 `yaml:"resource_mismatch_threshold" json:"resource_mismatch_threshold"` // Percentage difference to flag
}

// DefaultVerificationConfig returns the default verification configuration
func DefaultVerificationConfig() *VerificationConfig {
	return &VerificationConfig{
		// Heartbeat
		HeartbeatIntervalSeconds:    30,
		HeartbeatTimeoutSeconds:     10,
		MissedHeartbeatsBeforeAlert: 3,

		// Consensus
		ConsensusRequired:      2,    // 2 out of 3
		OutputHashVerification: true,
		DeterministicJobsOnly:  true,

		// Canary
		EnableCanaryProbes:      true,
		CanaryProbeIntervalMins: 15,
		CanaryFailureThreshold:  2,

		// Resource attestation
		ResourceAttestationEnabled: true,
		ResourceMismatchThreshold:  20.0, // 20% mismatch threshold
	}
}

// DeploymentEncryption represents encryption keys for a deployment
type DeploymentEncryption struct {
	DeploymentID     string    `json:"deployment_id"`
	RequesterPubKey  []byte    `json:"requester_pub_key"`  // Ed25519 or X25519 public key
	EncryptedDEK     []byte    `json:"encrypted_dek"`      // DEK encrypted with requester's public key
	DEKAlgorithm     string    `json:"dek_algorithm"`      // "AES-256-GCM"
	KeyDerivation    string    `json:"key_derivation"`     // "HKDF-SHA256"
	Nonce            []byte    `json:"nonce"`              // For DEK encryption
	CreatedAt        time.Time `json:"created_at"`
}

// EncryptionConfig defines encryption settings
type EncryptionConfig struct {
	// Volume encryption
	VolumeEncryptionEnabled bool   `yaml:"volume_encryption_enabled" json:"volume_encryption_enabled"`
	VolumeEncryptionCipher  string `yaml:"volume_encryption_cipher" json:"volume_encryption_cipher"` // "aes-xts-plain64"
	VolumeKeySize           int    `yaml:"volume_key_size" json:"volume_key_size"`                   // 256 or 512

	// Data encryption (logs, outputs)
	DataEncryptionAlgorithm string `yaml:"data_encryption_algorithm" json:"data_encryption_algorithm"` // "AES-256-GCM"
	KeyDerivationFunction   string `yaml:"key_derivation_function" json:"key_derivation_function"`     // "HKDF-SHA256"

	// Key exchange
	KeyExchangeAlgorithm string `yaml:"key_exchange_algorithm" json:"key_exchange_algorithm"` // "X25519"

	// Requester key storage
	KeyStorePath    string `yaml:"key_store_path" json:"key_store_path"`
	KeyStoreEncrypt bool   `yaml:"key_store_encrypt" json:"key_store_encrypt"` // Encrypt key store with passphrase
}

// DefaultEncryptionConfig returns the default encryption configuration
func DefaultEncryptionConfig() *EncryptionConfig {
	return &EncryptionConfig{
		// Volume encryption
		VolumeEncryptionEnabled: true,
		VolumeEncryptionCipher:  "aes-xts-plain64",
		VolumeKeySize:           256,

		// Data encryption
		DataEncryptionAlgorithm: "AES-256-GCM",
		KeyDerivationFunction:   "HKDF-SHA256",

		// Key exchange
		KeyExchangeAlgorithm: "X25519",

		// Key storage
		KeyStorePath:    "", // Set from config
		KeyStoreEncrypt: true,
	}
}

// WorkloadType represents the type of workload
type WorkloadType string

const (
	WorkloadTypeService   WorkloadType = "service"   // Long-running
	WorkloadTypeJob       WorkloadType = "job"       // Batch
	WorkloadTypeScheduled WorkloadType = "scheduled" // Cron
	WorkloadTypeFunction  WorkloadType = "function"  // Serverless
)

// WorkloadSpec defines a workload specification
type WorkloadSpec struct {
	Type         WorkloadType   `yaml:"type" json:"type"`
	Name         string         `yaml:"name" json:"name"`
	Image        string         `yaml:"image" json:"image"`
	Resources    ResourceLimits `yaml:"resources" json:"resources"`
	Command      []string       `yaml:"command,omitempty" json:"command,omitempty"`
	Args         []string       `yaml:"args,omitempty" json:"args,omitempty"`
	Environment  map[string]string `yaml:"environment,omitempty" json:"environment,omitempty"`
	Ports        []PortMapping  `yaml:"ports,omitempty" json:"ports,omitempty"`
	HealthCheck  *HealthCheck   `yaml:"health_check,omitempty" json:"health_check,omitempty"`
	Network      NetworkConfig  `yaml:"network" json:"network"`
	Timeout      time.Duration  `yaml:"timeout,omitempty" json:"timeout,omitempty"`
	Retries      int            `yaml:"retries,omitempty" json:"retries,omitempty"`

	// For scheduled jobs
	Schedule     string `yaml:"schedule,omitempty" json:"schedule,omitempty"` // Cron expression
	Timezone     string `yaml:"timezone,omitempty" json:"timezone,omitempty"`

	// For functions
	Trigger      []TriggerConfig `yaml:"trigger,omitempty" json:"trigger,omitempty"`
	ScaleToZero  bool            `yaml:"scale_to_zero,omitempty" json:"scale_to_zero,omitempty"`

	// Input/Output for jobs
	Inputs       []DataSource `yaml:"inputs,omitempty" json:"inputs,omitempty"`
	Outputs      []DataSource `yaml:"outputs,omitempty" json:"outputs,omitempty"`
}

// PortMapping defines port exposure
type PortMapping struct {
	Name          string `yaml:"name" json:"name"`
	ContainerPort int    `yaml:"container_port" json:"container_port"`
	HostPort      int    `yaml:"host_port,omitempty" json:"host_port,omitempty"`
	Protocol      string `yaml:"protocol" json:"protocol"` // TCP, UDP
}

// HealthCheck defines health check configuration
type HealthCheck struct {
	Type             string `yaml:"type" json:"type"` // http, tcp, exec
	Path             string `yaml:"path,omitempty" json:"path,omitempty"`
	Port             int    `yaml:"port,omitempty" json:"port,omitempty"`
	Command          []string `yaml:"command,omitempty" json:"command,omitempty"`
	IntervalSeconds  int    `yaml:"interval_seconds" json:"interval_seconds"`
	TimeoutSeconds   int    `yaml:"timeout_seconds" json:"timeout_seconds"`
	FailureThreshold int    `yaml:"failure_threshold" json:"failure_threshold"`
	SuccessThreshold int    `yaml:"success_threshold" json:"success_threshold"`
}

// NetworkConfig defines network settings for a workload
type NetworkConfig struct {
	Mode         NetworkMode `yaml:"mode" json:"mode"` // tor, direct, hybrid
	OnionService bool        `yaml:"onion_service" json:"onion_service"`
	TorOnly      bool        `yaml:"tor_only" json:"tor_only"`
}

// TriggerConfig defines function triggers
type TriggerConfig struct {
	Type    string   `yaml:"type" json:"type"`       // http, event, schedule
	Path    string   `yaml:"path,omitempty" json:"path,omitempty"`
	Methods []string `yaml:"methods,omitempty" json:"methods,omitempty"`
	Source  string   `yaml:"source,omitempty" json:"source,omitempty"`
	Event   string   `yaml:"event,omitempty" json:"event,omitempty"`
}

// DataSource defines input/output data location
type DataSource struct {
	Source      string `yaml:"source" json:"source"`           // ipfs://, file://, etc.
	Destination string `yaml:"destination" json:"destination"` // Container path
}

// GeographicRegion represents a geographic region
type GeographicRegion string

const (
	RegionAmericas    GeographicRegion = "americas"
	RegionEurope      GeographicRegion = "europe"
	RegionAsiaPacific GeographicRegion = "asia_pacific"
	RegionAfrica      GeographicRegion = "africa"
	RegionMiddleEast  GeographicRegion = "middle_east"
)

// RegionConfig defines region-specific settings
type RegionConfig struct {
	PrimaryRegions   []GeographicRegion `yaml:"primary_regions" json:"primary_regions"`
	SecondaryRegions []GeographicRegion `yaml:"secondary_regions" json:"secondary_regions"`
	MinProvidersForSecondary int        `yaml:"min_providers_for_secondary" json:"min_providers_for_secondary"`
}

// DefaultRegionConfig returns the default region configuration
func DefaultRegionConfig() *RegionConfig {
	return &RegionConfig{
		PrimaryRegions: []GeographicRegion{
			RegionAmericas,
			RegionEurope,
			RegionAsiaPacific,
		},
		SecondaryRegions: []GeographicRegion{
			RegionAfrica,
			RegionMiddleEast,
		},
		MinProvidersForSecondary: 50,
	}
}
