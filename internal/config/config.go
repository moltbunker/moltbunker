package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents the complete daemon configuration
type Config struct {
	Daemon     DaemonConfig     `yaml:"daemon"`
	P2P        P2PConfig        `yaml:"p2p"`
	Tor        TorConfig        `yaml:"tor"`
	Runtime    RuntimeConfig    `yaml:"runtime"`
	Redundancy RedundancyConfig `yaml:"redundancy"`
	Payment    PaymentConfig    `yaml:"payment"`
}

// DaemonConfig contains daemon settings
type DaemonConfig struct {
	Port        int    `yaml:"port"`
	DataDir     string `yaml:"data_dir"`
	KeyPath     string `yaml:"key_path"`
	KeystoreDir string `yaml:"keystore_dir"`
	SocketPath  string `yaml:"socket_path"`
	LogLevel    string `yaml:"log_level"`
}

// P2PConfig contains P2P network settings
type P2PConfig struct {
	BootstrapNodes []string `yaml:"bootstrap_nodes"`
	NetworkMode    string   `yaml:"network_mode"` // clearnet, tor_only, hybrid
	MaxPeers       int      `yaml:"max_peers"`
	DialTimeout    int      `yaml:"dial_timeout_seconds"`
	EnableMDNS     bool     `yaml:"enable_mdns"` // Local network discovery
	ExternalIP     string   `yaml:"external_ip"` // Public IP if behind NAT
	AnnounceAddrs  []string `yaml:"announce_addrs"`
}

// TorConfig contains Tor settings
type TorConfig struct {
	Enabled         bool   `yaml:"enabled"`
	DataDir         string `yaml:"data_dir"`
	SOCKS5Port      int    `yaml:"socks5_port"`
	ControlPort     int    `yaml:"control_port"`
	ExitNodeCountry string `yaml:"exit_node_country"`
	StrictNodes     bool   `yaml:"strict_nodes"`
}

// RuntimeConfig contains container runtime settings
type RuntimeConfig struct {
	ContainerdSocket string          `yaml:"containerd_socket"`
	Namespace        string          `yaml:"namespace"`
	DefaultResources ResourceLimits  `yaml:"default_resources"`
	EnableEncryption bool            `yaml:"enable_encryption"`
}

// ResourceLimits defines default resource limits
type ResourceLimits struct {
	CPUQuota    int64 `yaml:"cpu_quota"`
	CPUPeriod   int64 `yaml:"cpu_period"`
	MemoryLimit int64 `yaml:"memory_limit"`
	DiskLimit   int64 `yaml:"disk_limit"`
	NetworkBW   int64 `yaml:"network_bw"`
	PIDLimit    int   `yaml:"pid_limit"`
}

// RedundancyConfig contains redundancy settings
type RedundancyConfig struct {
	ReplicaCount        int `yaml:"replica_count"`
	HealthCheckInterval int `yaml:"health_check_interval_seconds"`
	HealthTimeout       int `yaml:"health_timeout_seconds"`
	FailoverDelay       int `yaml:"failover_delay_seconds"`
}

// PaymentConfig contains payment settings (mocked for now)
type PaymentConfig struct {
	Enabled         bool   `yaml:"enabled"`
	ContractAddress string `yaml:"contract_address"`
	MinStake        string `yaml:"min_stake"`
	BasePricePerHour string `yaml:"base_price_per_hour"`
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
		},
		P2P: P2PConfig{
			BootstrapNodes: []string{
				// Default bootstrap nodes - these would be operated by the project
				// For now, empty - users must configure their own or use mDNS
			},
			NetworkMode:   "hybrid",
			MaxPeers:      50,
			DialTimeout:   30,
			EnableMDNS:    true, // Enable local network discovery by default
			ExternalIP:    "",
			AnnounceAddrs: []string{},
		},
		Tor: TorConfig{
			Enabled:         false,
			DataDir:         filepath.Join(dataDir, "tor"),
			SOCKS5Port:      9050,
			ControlPort:     9051,
			ExitNodeCountry: "",
			StrictNodes:     false,
		},
		Runtime: RuntimeConfig{
			ContainerdSocket: "/run/containerd/containerd.sock",
			Namespace:        "moltbunker",
			DefaultResources: ResourceLimits{
				CPUQuota:    100000,
				CPUPeriod:   100000,
				MemoryLimit: 1024 * 1024 * 1024, // 1GB
				DiskLimit:   10 * 1024 * 1024 * 1024, // 10GB
				NetworkBW:   10 * 1024 * 1024, // 10MB/s
				PIDLimit:    100,
			},
			EnableEncryption: true,
		},
		Redundancy: RedundancyConfig{
			ReplicaCount:        3,
			HealthCheckInterval: 10,
			HealthTimeout:       30,
			FailoverDelay:       60,
		},
		Payment: PaymentConfig{
			Enabled:          false, // Disabled by default - smart contracts not implemented
			ContractAddress:  "",
			MinStake:         "1000000000000000000", // 1 token
			BasePricePerHour: "1000000000000000",    // 0.001 token
		},
	}
}

// Load loads configuration from file
func Load(path string) (*Config, error) {
	// Start with defaults
	cfg := DefaultConfig()

	// Expand path
	path = expandPath(path)

	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Return defaults if file doesn't exist
			return cfg, nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Expand paths in config
	cfg.expandPaths()

	// Validate
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// Save saves configuration to file
func (c *Config) Save(path string) error {
	path = expandPath(path)

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal to YAML
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write file with secure permissions (owner read/write only)
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Daemon.Port < 1 || c.Daemon.Port > 65535 {
		return fmt.Errorf("invalid port: %d", c.Daemon.Port)
	}

	if c.P2P.MaxPeers < 1 {
		return fmt.Errorf("max_peers must be at least 1")
	}

	if c.P2P.NetworkMode != "clearnet" && c.P2P.NetworkMode != "tor_only" && c.P2P.NetworkMode != "hybrid" {
		return fmt.Errorf("invalid network_mode: %s", c.P2P.NetworkMode)
	}

	if c.Redundancy.ReplicaCount < 1 || c.Redundancy.ReplicaCount > 10 {
		return fmt.Errorf("replica_count must be between 1 and 10")
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

// EnsureDirectories creates all necessary directories
func (c *Config) EnsureDirectories() error {
	dirs := []string{
		c.Daemon.DataDir,
		filepath.Dir(c.Daemon.KeyPath),
		c.Daemon.KeystoreDir,
		filepath.Dir(c.Daemon.SocketPath),
		c.Tor.DataDir,
		filepath.Join(c.Daemon.DataDir, "containers"),
		filepath.Join(c.Daemon.DataDir, "volumes"),
		filepath.Join(c.Daemon.DataDir, "logs"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}
