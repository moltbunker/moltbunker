package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/moltbunker/moltbunker/pkg/types"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg == nil {
		t.Fatal("DefaultConfig returned nil")
	}

	// Verify daemon defaults
	if cfg.Daemon.Port != 9000 {
		t.Errorf("expected default port 9000, got %d", cfg.Daemon.Port)
	}
	if cfg.Daemon.LogLevel != "info" {
		t.Errorf("expected default log level 'info', got %s", cfg.Daemon.LogLevel)
	}
	if cfg.Daemon.LogFormat != "text" {
		t.Errorf("expected default log format 'text', got %s", cfg.Daemon.LogFormat)
	}

	// Verify node defaults
	if cfg.Node.Role != types.NodeRoleRequester {
		t.Errorf("expected default role 'requester', got %s", cfg.Node.Role)
	}

	// Verify P2P defaults
	if cfg.P2P.MaxPeers != 50 {
		t.Errorf("expected default max_peers 50, got %d", cfg.P2P.MaxPeers)
	}
	if cfg.P2P.NetworkMode != "hybrid" {
		t.Errorf("expected default network mode 'hybrid', got %s", cfg.P2P.NetworkMode)
	}
	if !cfg.P2P.EnableMDNS {
		t.Error("expected mDNS enabled by default")
	}

	// Verify redundancy defaults
	if cfg.Redundancy.ReplicaCount != 3 {
		t.Errorf("expected default replica count 3, got %d", cfg.Redundancy.ReplicaCount)
	}

	// Verify economics defaults
	if cfg.Economics.TokenDecimals != 18 {
		t.Errorf("expected token decimals 18, got %d", cfg.Economics.TokenDecimals)
	}
	if cfg.Economics.ChainID != 8453 {
		t.Errorf("expected chain ID 8453 (Base mainnet), got %d", cfg.Economics.ChainID)
	}

	// Verify encryption defaults
	if !cfg.Encryption.VolumeEncryptionEnabled {
		t.Error("expected volume encryption enabled by default")
	}
	if cfg.Encryption.DataEncryptionAlgorithm != "AES-256-GCM" {
		t.Errorf("expected AES-256-GCM, got %s", cfg.Encryption.DataEncryptionAlgorithm)
	}

	// Verify security defaults
	if cfg.Security.TLSMinVersion != "1.3" {
		t.Errorf("expected TLS 1.3, got %s", cfg.Security.TLSMinVersion)
	}
	if !cfg.Security.CertPinning {
		t.Error("expected cert pinning enabled by default")
	}
	if !cfg.Security.MutualTLS {
		t.Error("expected mutual TLS enabled by default")
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		modify  func(c *Config)
		wantErr bool
	}{
		{
			name:    "default config is valid",
			modify:  func(c *Config) {},
			wantErr: false,
		},
		{
			name:    "port zero is invalid",
			modify:  func(c *Config) { c.Daemon.Port = 0 },
			wantErr: true,
		},
		{
			name:    "port negative is invalid",
			modify:  func(c *Config) { c.Daemon.Port = -1 },
			wantErr: true,
		},
		{
			name:    "port over 65535 is invalid",
			modify:  func(c *Config) { c.Daemon.Port = 70000 },
			wantErr: true,
		},
		{
			name:    "port 65535 is valid",
			modify:  func(c *Config) { c.Daemon.Port = 65535 },
			wantErr: false,
		},
		{
			name:    "port 1 is valid",
			modify:  func(c *Config) { c.Daemon.Port = 1 },
			wantErr: false,
		},
		{
			name:    "invalid node role",
			modify:  func(c *Config) { c.Node.Role = "invalid" },
			wantErr: true,
		},
		{
			name:    "provider role is valid",
			modify:  func(c *Config) { c.Node.Role = types.NodeRoleProvider },
			wantErr: false,
		},
		{
			name:    "requester role is valid",
			modify:  func(c *Config) { c.Node.Role = types.NodeRoleRequester },
			wantErr: false,
		},
		{
			name:    "max peers zero is invalid",
			modify:  func(c *Config) { c.P2P.MaxPeers = 0 },
			wantErr: true,
		},
		{
			name:    "invalid network mode",
			modify:  func(c *Config) { c.P2P.NetworkMode = "invalid" },
			wantErr: true,
		},
		{
			name:    "clearnet network mode is valid",
			modify:  func(c *Config) { c.P2P.NetworkMode = "clearnet" },
			wantErr: false,
		},
		{
			name:    "tor_only network mode is valid",
			modify:  func(c *Config) { c.P2P.NetworkMode = "tor_only" },
			wantErr: false,
		},
		{
			name:    "replica count zero is invalid",
			modify:  func(c *Config) { c.Redundancy.ReplicaCount = 0 },
			wantErr: true,
		},
		{
			name:    "replica count over 10 is invalid",
			modify:  func(c *Config) { c.Redundancy.ReplicaCount = 11 },
			wantErr: true,
		},
		{
			name:    "replica count 10 is valid",
			modify:  func(c *Config) { c.Redundancy.ReplicaCount = 10 },
			wantErr: false,
		},
		{
			name:    "negative token decimals is invalid",
			modify:  func(c *Config) { c.Economics.TokenDecimals = -1 },
			wantErr: true,
		},
		{
			name:    "token decimals over 18 is invalid",
			modify:  func(c *Config) { c.Economics.TokenDecimals = 19 },
			wantErr: true,
		},
		{
			name:    "invalid TLS version",
			modify:  func(c *Config) { c.Security.TLSMinVersion = "1.0" },
			wantErr: true,
		},
		{
			name:    "TLS 1.2 is valid",
			modify:  func(c *Config) { c.Security.TLSMinVersion = "1.2" },
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			tt.modify(cfg)
			err := cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIsProviderIsRequester(t *testing.T) {
	tests := []struct {
		role         types.NodeRole
		isProvider   bool
		isRequester  bool
	}{
		{types.NodeRoleProvider, true, false},
		{types.NodeRoleRequester, false, true},
		{types.NodeRoleHybrid, true, true},
	}

	for _, tt := range tests {
		t.Run(string(tt.role), func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.Node.Role = tt.role

			if got := cfg.IsProvider(); got != tt.isProvider {
				t.Errorf("IsProvider() = %v, want %v for role %s", got, tt.isProvider, tt.role)
			}
			if got := cfg.IsRequester(); got != tt.isRequester {
				t.Errorf("IsRequester() = %v, want %v for role %s", got, tt.isRequester, tt.role)
			}
		})
	}
}

func TestGetTierConfig(t *testing.T) {
	cfg := DefaultConfig()

	// Starter tier should exist
	tier, exists := cfg.GetTierConfig(types.StakingTierStarter)
	if !exists {
		t.Fatal("expected StakingTierStarter to exist")
	}
	if tier == nil {
		t.Fatal("expected non-nil tier config")
	}

	// Nonexistent tier
	_, exists = cfg.GetTierConfig(types.StakingTier("nonexistent"))
	if exists {
		t.Error("expected nonexistent tier to not exist")
	}
}

func TestSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Create config, modify, and save
	cfg := DefaultConfig()
	cfg.Daemon.Port = 12345
	cfg.P2P.MaxPeers = 100
	cfg.Daemon.DataDir = tmpDir
	cfg.Node.Role = types.NodeRoleProvider

	if err := cfg.Save(configPath); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	// Verify file was created with restrictive permissions
	info, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("Config file not created: %v", err)
	}
	perm := info.Mode().Perm()
	if perm != 0600 {
		t.Errorf("expected file permissions 0600, got %o", perm)
	}

	// Load the config
	loaded, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if loaded.Daemon.Port != 12345 {
		t.Errorf("expected port 12345, got %d", loaded.Daemon.Port)
	}
	if loaded.P2P.MaxPeers != 100 {
		t.Errorf("expected max_peers 100, got %d", loaded.P2P.MaxPeers)
	}
	if loaded.Node.Role != types.NodeRoleProvider {
		t.Errorf("expected role 'provider', got %s", loaded.Node.Role)
	}
}

func TestLoadNonExistentReturnsDefaults(t *testing.T) {
	cfg, err := Load("/nonexistent/path/config.yaml")
	if err != nil {
		t.Fatalf("Load() of nonexistent file should not error, got: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected default config, got nil")
	}
	if cfg.Daemon.Port != 9000 {
		t.Errorf("expected default port 9000, got %d", cfg.Daemon.Port)
	}
}

func TestLoadInvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	if err := os.WriteFile(configPath, []byte("{{{{invalid yaml"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(configPath)
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestExpandPath(t *testing.T) {
	homeDir, _ := os.UserHomeDir()

	tests := []struct {
		input    string
		expected string
	}{
		{"~/test", filepath.Join(homeDir, "test")},
		{"~/.moltbunker", filepath.Join(homeDir, ".moltbunker")},
		{"/absolute/path", "/absolute/path"},
		{"relative/path", "relative/path"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := expandPath(tt.input)
			if got != tt.expected {
				t.Errorf("expandPath(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestEnsureDirectories(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := DefaultConfig()
	cfg.Daemon.DataDir = filepath.Join(tmpDir, "data")
	cfg.Daemon.KeyPath = filepath.Join(tmpDir, "keys", "node.key")
	cfg.Daemon.KeystoreDir = filepath.Join(tmpDir, "keystore")
	cfg.Daemon.SocketPath = filepath.Join(tmpDir, "daemon.sock")
	cfg.Tor.DataDir = filepath.Join(tmpDir, "tor")
	cfg.Runtime.LogsDir = filepath.Join(tmpDir, "logs")
	cfg.Runtime.VolumesDir = filepath.Join(tmpDir, "volumes")
	cfg.Encryption.KeyStorePath = filepath.Join(tmpDir, "encryption")

	if err := cfg.EnsureDirectories(); err != nil {
		t.Fatalf("EnsureDirectories() error: %v", err)
	}

	// Verify directories exist
	dirsToCheck := []string{
		cfg.Daemon.DataDir,
		cfg.Daemon.KeystoreDir,
		cfg.Tor.DataDir,
		cfg.Runtime.LogsDir,
		cfg.Runtime.VolumesDir,
		cfg.Encryption.KeyStorePath,
		filepath.Join(cfg.Daemon.DataDir, "containers"),
		filepath.Join(cfg.Daemon.DataDir, "state"),
	}

	for _, dir := range dirsToCheck {
		info, err := os.Stat(dir)
		if err != nil {
			t.Errorf("directory %s was not created: %v", dir, err)
			continue
		}
		if !info.IsDir() {
			t.Errorf("%s is not a directory", dir)
		}
	}
}

func TestDefaultAPIConfig(t *testing.T) {
	api := DefaultAPIConfig()

	if api.RateLimitRequests != 100 {
		t.Errorf("expected rate limit 100, got %d", api.RateLimitRequests)
	}
	if api.RateLimitWindowSecs != 60 {
		t.Errorf("expected rate limit window 60, got %d", api.RateLimitWindowSecs)
	}
	if api.MaxConcurrentConns != 100 {
		t.Errorf("expected max concurrent conns 100, got %d", api.MaxConcurrentConns)
	}
	if api.MaxRequestSize != 10*1024*1024 {
		t.Errorf("expected max request size 10MB, got %d", api.MaxRequestSize)
	}
	if api.ReadTimeoutSecs != 30 {
		t.Errorf("expected read timeout 30, got %d", api.ReadTimeoutSecs)
	}
	if api.WriteTimeoutSecs != 30 {
		t.Errorf("expected write timeout 30, got %d", api.WriteTimeoutSecs)
	}
	if api.IdleTimeoutSecs != 120 {
		t.Errorf("expected idle timeout 120, got %d", api.IdleTimeoutSecs)
	}
}

func TestDefaultCircuitBreakerConfig(t *testing.T) {
	cb := DefaultCircuitBreakerConfig()

	if !cb.Enabled {
		t.Error("expected circuit breaker enabled by default")
	}
	if cb.FailureThreshold != 5 {
		t.Errorf("expected failure threshold 5, got %d", cb.FailureThreshold)
	}
	if cb.SuccessThreshold != 2 {
		t.Errorf("expected success threshold 2, got %d", cb.SuccessThreshold)
	}
	if cb.TimeoutSecs != 30 {
		t.Errorf("expected timeout 30, got %d", cb.TimeoutSecs)
	}
	if cb.HalfOpenMaxRequests != 3 {
		t.Errorf("expected half open max requests 3, got %d", cb.HalfOpenMaxRequests)
	}
}

func TestDefaultConfigPath(t *testing.T) {
	path := DefaultConfigPath()
	if path == "" {
		t.Error("expected non-empty default config path")
	}

	homeDir, _ := os.UserHomeDir()
	expected := filepath.Join(homeDir, ".moltbunker", "config.yaml")
	if path != expected {
		t.Errorf("expected %q, got %q", expected, path)
	}
}

func TestLoadInvalidConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Write a valid YAML with invalid config values
	yamlContent := `
daemon:
  port: 0
`
	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(configPath)
	if err == nil {
		t.Error("expected validation error for port 0")
	}
}
