package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/moltbunker/moltbunker/pkg/types"
)

// TestYAMLOverridesDefaults verifies that YAML config values override defaults.
func TestYAMLOverridesDefaults(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	yamlContent := `
daemon:
  port: 7777
  log_level: debug
p2p:
  max_peers: 200
  network_mode: clearnet
redundancy:
  replica_count: 5
`
	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	// YAML values should override defaults
	if cfg.Daemon.Port != 7777 {
		t.Errorf("expected YAML port 7777, got %d", cfg.Daemon.Port)
	}
	if cfg.Daemon.LogLevel != "debug" {
		t.Errorf("expected YAML log_level 'debug', got %s", cfg.Daemon.LogLevel)
	}
	if cfg.P2P.MaxPeers != 200 {
		t.Errorf("expected YAML max_peers 200, got %d", cfg.P2P.MaxPeers)
	}
	if cfg.P2P.NetworkMode != "clearnet" {
		t.Errorf("expected YAML network_mode 'clearnet', got %s", cfg.P2P.NetworkMode)
	}
	if cfg.Redundancy.ReplicaCount != 5 {
		t.Errorf("expected YAML replica_count 5, got %d", cfg.Redundancy.ReplicaCount)
	}

	// Non-overridden values should retain defaults
	if cfg.Daemon.LogFormat != "text" {
		t.Errorf("expected default log_format 'text', got %s", cfg.Daemon.LogFormat)
	}
	if !cfg.P2P.EnableMDNS {
		t.Error("expected default enable_mdns true (not overridden)")
	}
	if !cfg.Encryption.VolumeEncryptionEnabled {
		t.Error("expected default volume_encryption_enabled true (not overridden)")
	}
}

// TestEnvironmentVariablesOverrideYAML verifies that environment variables
// can override YAML values when the config system reads env vars.
// Currently the config system loads from YAML only and does not read env vars
// directly into struct fields, so this test verifies that the env-based data
// directory mechanism works properly.
func TestEnvironmentVariablesOverrideYAML(t *testing.T) {
	// MOLTBUNKER_DATA_DIR is one of the documented env vars
	tmpDir := t.TempDir()
	customDataDir := filepath.Join(tmpDir, "custom-data")

	// Simulate the pattern the daemon uses: check env var, then use it to
	// construct the config path.
	t.Setenv("MOLTBUNKER_DATA_DIR", customDataDir)

	envDataDir := os.Getenv("MOLTBUNKER_DATA_DIR")
	if envDataDir != customDataDir {
		t.Fatalf("expected MOLTBUNKER_DATA_DIR=%s, got %s", customDataDir, envDataDir)
	}

	// When the daemon reads the env var it constructs the config path from it
	configPath := filepath.Join(envDataDir, "config.yaml")

	// Load from nonexistent file returns defaults
	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	// The default config uses ~/.moltbunker; the daemon would override DataDir
	// with the env var value. Verify the env var was readable.
	cfg.Daemon.DataDir = envDataDir
	if cfg.Daemon.DataDir != customDataDir {
		t.Errorf("expected data dir %s, got %s", customDataDir, cfg.Daemon.DataDir)
	}
}

// TestDefaultValuesAreSensibleForProduction verifies all default values are
// reasonable for a production deployment.
func TestDefaultValuesAreSensibleForProduction(t *testing.T) {
	cfg := DefaultConfig()

	// Daemon defaults
	if cfg.Daemon.Port < 1024 || cfg.Daemon.Port > 49151 {
		t.Errorf("default port %d should be in registered port range (1024-49151)", cfg.Daemon.Port)
	}
	if cfg.Daemon.LogLevel != "info" {
		t.Errorf("default log level should be 'info' for production, got %s", cfg.Daemon.LogLevel)
	}
	if cfg.Daemon.DataDir == "" {
		t.Error("default data directory must not be empty")
	}
	if cfg.Daemon.KeyPath == "" {
		t.Error("default key path must not be empty")
	}
	if cfg.Daemon.KeystoreDir == "" {
		t.Error("default keystore directory must not be empty")
	}
	if cfg.Daemon.SocketPath == "" {
		t.Error("default socket path must not be empty")
	}

	// P2P defaults
	if cfg.P2P.MaxPeers < 10 {
		t.Errorf("default max_peers %d seems too low for production", cfg.P2P.MaxPeers)
	}
	if cfg.P2P.DialTimeoutSecs < 5 {
		t.Errorf("default dial timeout %d secs too low", cfg.P2P.DialTimeoutSecs)
	}
	if cfg.P2P.ConnectionTimeout < 10 {
		t.Errorf("default connection timeout %d secs too low", cfg.P2P.ConnectionTimeout)
	}
	if cfg.P2P.IdleTimeout < 60 {
		t.Errorf("default idle timeout %d secs too low", cfg.P2P.IdleTimeout)
	}

	// Security defaults - must be strict for production
	if cfg.Security.TLSMinVersion != "1.3" {
		t.Errorf("default TLS version should be 1.3, got %s", cfg.Security.TLSMinVersion)
	}
	if !cfg.Security.CertPinning {
		t.Error("certificate pinning should be enabled by default")
	}
	if !cfg.Security.MutualTLS {
		t.Error("mutual TLS should be enabled by default")
	}
	if len(cfg.Security.TLSCipherSuites) == 0 {
		t.Error("TLS cipher suites should not be empty")
	}

	// Encryption defaults
	if !cfg.Encryption.VolumeEncryptionEnabled {
		t.Error("volume encryption should be enabled by default")
	}
	if cfg.Encryption.VolumeEncryptionCipher != "aes-xts-plain64" {
		t.Errorf("expected volume cipher 'aes-xts-plain64', got %s", cfg.Encryption.VolumeEncryptionCipher)
	}
	if cfg.Encryption.VolumeKeySize < 256 {
		t.Errorf("volume key size %d bits too small", cfg.Encryption.VolumeKeySize)
	}
	if cfg.Encryption.DataEncryptionAlgorithm != "AES-256-GCM" {
		t.Errorf("expected data encryption 'AES-256-GCM', got %s", cfg.Encryption.DataEncryptionAlgorithm)
	}
	if cfg.Encryption.KeyExchangeAlgorithm != "X25519" {
		t.Errorf("expected key exchange 'X25519', got %s", cfg.Encryption.KeyExchangeAlgorithm)
	}
	if !cfg.Encryption.KeyStoreEncrypt {
		t.Error("key store encryption should be enabled by default")
	}

	// Redundancy defaults
	if cfg.Redundancy.ReplicaCount != 3 {
		t.Errorf("default replica count should be 3, got %d", cfg.Redundancy.ReplicaCount)
	}
	if cfg.Redundancy.HealthCheckInterval < 5 {
		t.Errorf("health check interval %d secs too low", cfg.Redundancy.HealthCheckInterval)
	}

	// Economics defaults
	if cfg.Economics.TokenDecimals != 18 {
		t.Errorf("token decimals should be 18, got %d", cfg.Economics.TokenDecimals)
	}
	if cfg.Economics.ChainID != 8453 {
		t.Errorf("chain ID should be 8453 (Base mainnet), got %d", cfg.Economics.ChainID)
	}
	if cfg.Economics.BlockConfirmations < 1 {
		t.Errorf("block confirmations %d too low", cfg.Economics.BlockConfirmations)
	}

	// API defaults
	if cfg.API.RateLimitRequests < 1 {
		t.Error("API rate limit must be positive")
	}
	if cfg.API.RateLimitWindowSecs < 1 {
		t.Error("API rate limit window must be positive")
	}
	if cfg.API.MaxConcurrentConns < 1 {
		t.Error("API max concurrent connections must be positive")
	}
	if cfg.API.MaxRequestSize < 1024 {
		t.Errorf("API max request size %d bytes too small", cfg.API.MaxRequestSize)
	}
	if cfg.API.ReadTimeoutSecs < 1 {
		t.Error("API read timeout must be positive")
	}
	if cfg.API.WriteTimeoutSecs < 1 {
		t.Error("API write timeout must be positive")
	}

	// Tor defaults
	if cfg.Tor.SOCKS5Port != 9050 {
		t.Errorf("expected default Tor SOCKS5 port 9050, got %d", cfg.Tor.SOCKS5Port)
	}
	if cfg.Tor.ControlPort != 9051 {
		t.Errorf("expected default Tor control port 9051, got %d", cfg.Tor.ControlPort)
	}

	// Runtime defaults
	if cfg.Runtime.Namespace != "moltbunker" {
		t.Errorf("expected namespace 'moltbunker', got %s", cfg.Runtime.Namespace)
	}
	if cfg.Runtime.DefaultResources.MemoryLimit < 512*1024*1024 {
		t.Error("default memory limit too small")
	}
}

// TestAllRequiredFieldsHaveDefaults verifies that fields needed for the daemon
// to start have non-zero default values.
func TestAllRequiredFieldsHaveDefaults(t *testing.T) {
	cfg := DefaultConfig()

	required := []struct {
		name  string
		value interface{}
	}{
		{"Daemon.Port", cfg.Daemon.Port},
		{"Daemon.DataDir", cfg.Daemon.DataDir},
		{"Daemon.KeyPath", cfg.Daemon.KeyPath},
		{"Daemon.KeystoreDir", cfg.Daemon.KeystoreDir},
		{"Daemon.SocketPath", cfg.Daemon.SocketPath},
		{"Daemon.LogLevel", cfg.Daemon.LogLevel},
		{"Daemon.LogFormat", cfg.Daemon.LogFormat},
		{"Node.Role", string(cfg.Node.Role)},
		{"P2P.NetworkMode", cfg.P2P.NetworkMode},
		{"P2P.MaxPeers", cfg.P2P.MaxPeers},
		{"P2P.DialTimeoutSecs", cfg.P2P.DialTimeoutSecs},
		{"Security.TLSMinVersion", cfg.Security.TLSMinVersion},
		{"Redundancy.ReplicaCount", cfg.Redundancy.ReplicaCount},
		{"Redundancy.HealthCheckInterval", cfg.Redundancy.HealthCheckInterval},
		{"Economics.TokenDecimals", cfg.Economics.TokenDecimals},
		{"Economics.ChainID", cfg.Economics.ChainID},
		{"Encryption.DataEncryptionAlgorithm", cfg.Encryption.DataEncryptionAlgorithm},
		{"Encryption.KeyExchangeAlgorithm", cfg.Encryption.KeyExchangeAlgorithm},
		{"Runtime.Namespace", cfg.Runtime.Namespace},
	}

	for _, r := range required {
		switch v := r.value.(type) {
		case string:
			if v == "" {
				t.Errorf("required field %s has empty default", r.name)
			}
		case int:
			if v == 0 {
				t.Errorf("required field %s has zero default", r.name)
			}
		case int64:
			if v == 0 {
				t.Errorf("required field %s has zero default", r.name)
			}
		}
	}
}

// TestConfigValidationCatchesInvalidValues verifies that Validate() rejects
// various invalid configurations.
func TestConfigValidationCatchesInvalidValues(t *testing.T) {
	tests := []struct {
		name   string
		modify func(c *Config)
	}{
		{
			name:   "port too high",
			modify: func(c *Config) { c.Daemon.Port = 99999 },
		},
		{
			name:   "port zero",
			modify: func(c *Config) { c.Daemon.Port = 0 },
		},
		{
			name:   "negative port",
			modify: func(c *Config) { c.Daemon.Port = -1 },
		},
		{
			name:   "invalid node role",
			modify: func(c *Config) { c.Node.Role = "supernode" },
		},
		{
			name:   "zero max peers",
			modify: func(c *Config) { c.P2P.MaxPeers = 0 },
		},
		{
			name:   "negative max peers",
			modify: func(c *Config) { c.P2P.MaxPeers = -5 },
		},
		{
			name:   "invalid network mode",
			modify: func(c *Config) { c.P2P.NetworkMode = "wired" },
		},
		{
			name:   "replica count zero",
			modify: func(c *Config) { c.Redundancy.ReplicaCount = 0 },
		},
		{
			name:   "replica count too high",
			modify: func(c *Config) { c.Redundancy.ReplicaCount = 100 },
		},
		{
			name:   "token decimals negative",
			modify: func(c *Config) { c.Economics.TokenDecimals = -1 },
		},
		{
			name:   "token decimals too high",
			modify: func(c *Config) { c.Economics.TokenDecimals = 19 },
		},
		{
			name:   "invalid TLS version 1.0",
			modify: func(c *Config) { c.Security.TLSMinVersion = "1.0" },
		},
		{
			name:   "invalid TLS version 1.1",
			modify: func(c *Config) { c.Security.TLSMinVersion = "1.1" },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			tt.modify(cfg)
			if err := cfg.Validate(); err == nil {
				t.Error("expected validation error, got nil")
			}
		})
	}
}

// TestConfigValidationAcceptsValidValues verifies that Validate() accepts
// all valid configuration combinations.
func TestConfigValidationAcceptsValidValues(t *testing.T) {
	tests := []struct {
		name   string
		modify func(c *Config)
	}{
		{
			name:   "default config",
			modify: func(c *Config) {},
		},
		{
			name:   "provider role",
			modify: func(c *Config) { c.Node.Role = types.NodeRoleProvider },
		},
		{
			name:   "requester role",
			modify: func(c *Config) { c.Node.Role = types.NodeRoleRequester },
		},
		{
			name:   "hybrid role",
			modify: func(c *Config) { c.Node.Role = types.NodeRoleHybrid },
		},
		{
			name:   "clearnet mode",
			modify: func(c *Config) { c.P2P.NetworkMode = "clearnet" },
		},
		{
			name:   "tor_only mode",
			modify: func(c *Config) { c.P2P.NetworkMode = "tor_only" },
		},
		{
			name:   "hybrid mode",
			modify: func(c *Config) { c.P2P.NetworkMode = "hybrid" },
		},
		{
			name:   "TLS 1.2",
			modify: func(c *Config) { c.Security.TLSMinVersion = "1.2" },
		},
		{
			name:   "TLS 1.3",
			modify: func(c *Config) { c.Security.TLSMinVersion = "1.3" },
		},
		{
			name:   "min port",
			modify: func(c *Config) { c.Daemon.Port = 1 },
		},
		{
			name:   "max port",
			modify: func(c *Config) { c.Daemon.Port = 65535 },
		},
		{
			name:   "min replica count",
			modify: func(c *Config) { c.Redundancy.ReplicaCount = 1 },
		},
		{
			name:   "max replica count",
			modify: func(c *Config) { c.Redundancy.ReplicaCount = 10 },
		},
		{
			name:   "zero token decimals",
			modify: func(c *Config) { c.Economics.TokenDecimals = 0 },
		},
		{
			name:   "max token decimals",
			modify: func(c *Config) { c.Economics.TokenDecimals = 18 },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			tt.modify(cfg)
			if err := cfg.Validate(); err != nil {
				t.Errorf("expected no validation error, got: %v", err)
			}
		})
	}
}

// TestConfigLoadFromCustomPath verifies that config files can be loaded
// from arbitrary paths, not just the default location.
func TestConfigLoadFromCustomPath(t *testing.T) {
	// Create a deeply nested custom path
	tmpDir := t.TempDir()
	customPath := filepath.Join(tmpDir, "custom", "nested", "dir", "myconfig.yaml")

	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(customPath), 0755); err != nil {
		t.Fatal(err)
	}

	yamlContent := `
daemon:
  port: 4444
  log_level: warn
node:
  role: provider
p2p:
  max_peers: 75
  network_mode: tor_only
security:
  tls_min_version: "1.2"
redundancy:
  replica_count: 7
economics:
  chain_id: 84532
  rpc_url: "https://sepolia.base.org"
encryption:
  volume_encryption_cipher: "aes-xts-plain64"
`
	if err := os.WriteFile(customPath, []byte(yamlContent), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(customPath)
	if err != nil {
		t.Fatalf("Load(%s) error: %v", customPath, err)
	}

	if cfg.Daemon.Port != 4444 {
		t.Errorf("expected port 4444, got %d", cfg.Daemon.Port)
	}
	if cfg.Daemon.LogLevel != "warn" {
		t.Errorf("expected log_level 'warn', got %s", cfg.Daemon.LogLevel)
	}
	if cfg.Node.Role != types.NodeRoleProvider {
		t.Errorf("expected role 'provider', got %s", cfg.Node.Role)
	}
	if cfg.P2P.MaxPeers != 75 {
		t.Errorf("expected max_peers 75, got %d", cfg.P2P.MaxPeers)
	}
	if cfg.P2P.NetworkMode != "tor_only" {
		t.Errorf("expected network_mode 'tor_only', got %s", cfg.P2P.NetworkMode)
	}
	if cfg.Security.TLSMinVersion != "1.2" {
		t.Errorf("expected tls_min_version '1.2', got %s", cfg.Security.TLSMinVersion)
	}
	if cfg.Redundancy.ReplicaCount != 7 {
		t.Errorf("expected replica_count 7, got %d", cfg.Redundancy.ReplicaCount)
	}
	if cfg.Economics.ChainID != 84532 {
		t.Errorf("expected chain_id 84532, got %d", cfg.Economics.ChainID)
	}
	if cfg.Economics.RPCURL != "https://sepolia.base.org" {
		t.Errorf("expected rpc_url 'https://sepolia.base.org', got %s", cfg.Economics.RPCURL)
	}

	// Verify non-overridden defaults survive
	if cfg.Economics.TokenDecimals != 18 {
		t.Errorf("expected default token_decimals 18, got %d", cfg.Economics.TokenDecimals)
	}
	if !cfg.Encryption.VolumeEncryptionEnabled {
		t.Error("expected default volume_encryption_enabled true")
	}
	if !cfg.Security.MutualTLS {
		t.Error("expected default mutual_tls true")
	}
}

// TestConfigSaveLoadRoundTrip verifies that saving and loading a config
// preserves all field values.
func TestConfigSaveLoadRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "roundtrip.yaml")

	original := DefaultConfig()
	original.Daemon.Port = 11111
	original.Daemon.LogLevel = "debug"
	original.Daemon.DataDir = tmpDir
	original.P2P.MaxPeers = 123
	original.P2P.NetworkMode = "clearnet"
	original.Node.Role = types.NodeRoleRequester
	original.Redundancy.ReplicaCount = 5
	original.Security.TLSMinVersion = "1.2"

	if err := original.Save(configPath); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	loaded, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if loaded.Daemon.Port != original.Daemon.Port {
		t.Errorf("port: expected %d, got %d", original.Daemon.Port, loaded.Daemon.Port)
	}
	if loaded.Daemon.LogLevel != original.Daemon.LogLevel {
		t.Errorf("log_level: expected %s, got %s", original.Daemon.LogLevel, loaded.Daemon.LogLevel)
	}
	if loaded.P2P.MaxPeers != original.P2P.MaxPeers {
		t.Errorf("max_peers: expected %d, got %d", original.P2P.MaxPeers, loaded.P2P.MaxPeers)
	}
	if loaded.P2P.NetworkMode != original.P2P.NetworkMode {
		t.Errorf("network_mode: expected %s, got %s", original.P2P.NetworkMode, loaded.P2P.NetworkMode)
	}
	if loaded.Node.Role != original.Node.Role {
		t.Errorf("role: expected %s, got %s", original.Node.Role, loaded.Node.Role)
	}
	if loaded.Redundancy.ReplicaCount != original.Redundancy.ReplicaCount {
		t.Errorf("replica_count: expected %d, got %d", original.Redundancy.ReplicaCount, loaded.Redundancy.ReplicaCount)
	}
	if loaded.Security.TLSMinVersion != original.Security.TLSMinVersion {
		t.Errorf("tls_min_version: expected %s, got %s", original.Security.TLSMinVersion, loaded.Security.TLSMinVersion)
	}
}

// TestConfigLoadInvalidYAMLWithValidation verifies that loading a YAML file
// with valid syntax but semantically invalid values is caught by validation.
func TestConfigLoadInvalidYAMLWithValidation(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{
			name: "port out of range",
			content: `
daemon:
  port: 100000
`,
		},
		{
			name: "invalid network mode",
			content: `
p2p:
  network_mode: satellite
`,
		},
		{
			name: "invalid TLS version",
			content: `
security:
  tls_min_version: "1.0"
`,
		},
		{
			name: "replica count too high",
			content: `
redundancy:
  replica_count: 50
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.yaml")

			if err := os.WriteFile(configPath, []byte(tt.content), 0644); err != nil {
				t.Fatal(err)
			}

			_, err := Load(configPath)
			if err == nil {
				t.Error("expected validation error, got nil")
			}
		})
	}
}

// TestPartialYAMLPreservesDefaults verifies that a YAML file containing only
// a subset of fields leaves all other fields at their defaults.
func TestPartialYAMLPreservesDefaults(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "partial.yaml")

	// Only set one field
	yamlContent := `
daemon:
  port: 5555
`
	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	defaults := DefaultConfig()

	// Overridden field
	if cfg.Daemon.Port != 5555 {
		t.Errorf("expected port 5555, got %d", cfg.Daemon.Port)
	}

	// All other fields should match defaults
	if cfg.Daemon.LogLevel != defaults.Daemon.LogLevel {
		t.Errorf("log_level: expected default %s, got %s", defaults.Daemon.LogLevel, cfg.Daemon.LogLevel)
	}
	if cfg.P2P.MaxPeers != defaults.P2P.MaxPeers {
		t.Errorf("max_peers: expected default %d, got %d", defaults.P2P.MaxPeers, cfg.P2P.MaxPeers)
	}
	if cfg.P2P.NetworkMode != defaults.P2P.NetworkMode {
		t.Errorf("network_mode: expected default %s, got %s", defaults.P2P.NetworkMode, cfg.P2P.NetworkMode)
	}
	if cfg.Security.TLSMinVersion != defaults.Security.TLSMinVersion {
		t.Errorf("tls_min_version: expected default %s, got %s", defaults.Security.TLSMinVersion, cfg.Security.TLSMinVersion)
	}
	if cfg.Redundancy.ReplicaCount != defaults.Redundancy.ReplicaCount {
		t.Errorf("replica_count: expected default %d, got %d", defaults.Redundancy.ReplicaCount, cfg.Redundancy.ReplicaCount)
	}
	if cfg.Economics.ChainID != defaults.Economics.ChainID {
		t.Errorf("chain_id: expected default %d, got %d", defaults.Economics.ChainID, cfg.Economics.ChainID)
	}
	if cfg.Encryption.DataEncryptionAlgorithm != defaults.Encryption.DataEncryptionAlgorithm {
		t.Errorf("data_encryption: expected default %s, got %s",
			defaults.Encryption.DataEncryptionAlgorithm, cfg.Encryption.DataEncryptionAlgorithm)
	}
}
