package commands

import (
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"

	"github.com/moltbunker/moltbunker/internal/config"
	"github.com/moltbunker/moltbunker/pkg/types"
)

// Global CLI flags
var (
	// SocketPath is the path to the daemon socket
	SocketPath string

	// APIEndpoint is the HTTP API base URL for daemonless requester mode
	APIEndpoint string

	// OutputFormat controls output format: "" (auto), "json", "plain"
	OutputFormat string
)

// GetNodeRole loads the node role from config without requiring a daemon.
func GetNodeRole() types.NodeRole {
	cfg := loadConfigQuiet()
	if cfg == nil {
		return types.NodeRoleRequester // default
	}
	if cfg.Node.Role == "" {
		return types.NodeRoleRequester
	}
	return cfg.Node.Role
}

// IsProvider returns true if the configured role is provider or hybrid.
func IsProvider() bool {
	role := GetNodeRole()
	return role == types.NodeRoleProvider || role == types.NodeRoleHybrid
}

// GetAPIEndpoint returns the API endpoint from flag, config, or default.
func GetAPIEndpoint() string {
	if APIEndpoint != "" {
		return APIEndpoint
	}
	cfg := loadConfigQuiet()
	if cfg != nil && cfg.Node.APIEndpoint != "" {
		return cfg.Node.APIEndpoint
	}
	return "" // no API configured
}

// GetKeystoreDir returns the keystore directory from config or default.
func GetKeystoreDir() string {
	cfg := loadConfigQuiet()
	if cfg != nil && cfg.Daemon.KeystoreDir != "" {
		return cfg.Daemon.KeystoreDir
	}
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".moltbunker", "keystore")
}

// loadConfigQuiet loads config from the default path, returning nil on error.
func loadConfigQuiet() *config.Config {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	configPath := filepath.Join(homeDir, ".moltbunker", "config.yaml")
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil
	}
	return cfg
}

// Version information (set at build time)
var (
	Version   = "dev"
	Commit    = "unknown"
	BuildDate = "unknown"
)

// GetVersion returns the version string
func GetVersion() string {
	if Version != "dev" {
		return Version
	}
	// Try to get version from build info
	if info, ok := debug.ReadBuildInfo(); ok {
		if info.Main.Version != "" && info.Main.Version != "(devel)" {
			return info.Main.Version
		}
	}
	return "dev"
}

// GetCommit returns the git commit
func GetCommit() string {
	if Commit != "unknown" {
		return Commit
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				if len(setting.Value) > 8 {
					return setting.Value[:8]
				}
				return setting.Value
			}
		}
	}
	return "unknown"
}

// GetGoVersion returns the Go version
func GetGoVersion() string {
	return runtime.Version()
}
