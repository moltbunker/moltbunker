package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/moltbunker/moltbunker/internal/config"
	"github.com/moltbunker/moltbunker/internal/identity"
	"github.com/spf13/cobra"
)

var (
	installDataDir string
)

func NewInstallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install and initialize the moltbunker daemon",
		Long: `Install and initialize the moltbunker daemon.

This command will:
- Create necessary directories
- Generate Ed25519 node keys
- Generate TLS certificates
- Create default configuration
- Optionally install systemd/launchd service`,
		RunE: runInstall,
	}

	homeDir, _ := os.UserHomeDir()
	defaultDataDir := filepath.Join(homeDir, ".moltbunker")

	cmd.Flags().StringVar(&installDataDir, "data-dir", defaultDataDir, "Data directory for moltbunker")
	cmd.Flags().Bool("systemd", false, "Install systemd service (Linux)")
	cmd.Flags().Bool("launchd", false, "Install launchd service (macOS)")

	return cmd
}

func runInstall(cmd *cobra.Command, args []string) error {
	fmt.Println("Installing moltbunker daemon...")

	// Create directory structure
	dirs := []string{
		installDataDir,
		filepath.Join(installDataDir, "keys"),
		filepath.Join(installDataDir, "keystore"),
		filepath.Join(installDataDir, "data"),
		filepath.Join(installDataDir, "tor"),
		filepath.Join(installDataDir, "containers"),
		filepath.Join(installDataDir, "logs"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
		fmt.Printf("  Created: %s\n", dir)
	}

	// Generate node keys
	keyPath := filepath.Join(installDataDir, "keys", "node.key")
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		fmt.Println("Generating Ed25519 node keys...")
		keyManager, err := identity.NewKeyManager(keyPath)
		if err != nil {
			return fmt.Errorf("failed to generate keys: %w", err)
		}
		fmt.Printf("  Node ID: %s\n", keyManager.NodeID().String())
	} else {
		fmt.Println("  Node keys already exist, skipping generation")
	}

	// Write default configuration
	configPath := filepath.Join(installDataDir, "config.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		config := generateDefaultConfig(installDataDir)
		if err := os.WriteFile(configPath, []byte(config), 0644); err != nil {
			return fmt.Errorf("failed to write config: %w", err)
		}
		fmt.Printf("  Created config: %s\n", configPath)
	} else {
		fmt.Println("  Config already exists, skipping")
	}

	// Install service (optional)
	fmt.Println("\nTo install as a system service:")
	switch runtime.GOOS {
	case "linux":
		fmt.Println("  sudo moltbunker install --systemd")
		installSystemdService(cmd)
	case "darwin":
		fmt.Println("  moltbunker install --launchd")
		installLaunchdService(cmd)
	default:
		fmt.Println("  Service installation not supported on this platform")
	}

	fmt.Println("\nInstallation complete!")
	fmt.Println("\nTo start the daemon:")
	fmt.Println("  moltbunker start")

	return nil
}

func generateDefaultConfig(dataDir string) string {
	// Detect containerd socket — only include runtime section if found
	runtimeSection := ""
	if socket := config.DetectContainerdSocket(); socket != "" {
		runtimeSection = fmt.Sprintf(`
# Runtime settings (only needed for provider nodes)
runtime:
  containerd_socket: %s
  namespace: moltbunker
  runtime_name: auto
`, socket)
	}

	return fmt.Sprintf(`# Moltbunker Configuration

# Daemon settings
daemon:
  port: 9000
  data_dir: %s
  log_level: info

# Node role: requester (deploy jobs) or provider (host containers) or hybrid (both)
# To become a provider, change to "provider" or "hybrid" and ensure containerd is running
node:
  role: requester

# P2P settings
p2p:
  bootstrap_nodes: []
  network_mode: hybrid  # clearnet, tor_only, hybrid

# Tor settings
tor:
  enabled: false
  socks_port: 9050
  control_port: 9051
  data_dir: %s/tor
%s
# Container settings
container:
  default_cpu_quota: 100000
  default_memory_limit: 536870912  # 512MB
  default_disk_limit: 10737418240  # 10GB

# Redundancy settings
redundancy:
  replica_count: 3
  health_check_interval: 30s

# Economics settings
economics:
  mock_payments: true  # Set to false for real on-chain payments

# Payment settings
payment:
  contract_address: ""
  rpc_url: ""
`, dataDir, dataDir, runtimeSection)
}

func installSystemdService(cmd *cobra.Command) {
	// Check if --systemd flag is set
	systemdFlag, _ := cmd.Flags().GetBool("systemd")
	if !systemdFlag {
		return
	}

	serviceContent := fmt.Sprintf(`[Unit]
Description=Moltbunker P2P Container Runtime
After=network.target

[Service]
Type=simple
User=%s
ExecStart=/usr/local/bin/moltbunkerd --data=%s
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
`, os.Getenv("USER"), installDataDir)

	servicePath := "/etc/systemd/system/moltbunker.service"
	if err := os.WriteFile(servicePath, []byte(serviceContent), 0644); err != nil {
		fmt.Printf("  Failed to write systemd service: %v\n", err)
		return
	}

	// Reload systemd
	exec.Command("systemctl", "daemon-reload").Run()
	exec.Command("systemctl", "enable", "moltbunker").Run()

	fmt.Println("  Systemd service installed")
	fmt.Println("  Start with: sudo systemctl start moltbunker")
}

func installLaunchdService(cmd *cobra.Command) {
	// Check if --launchd flag is set
	launchdFlag, _ := cmd.Flags().GetBool("launchd")
	if !launchdFlag {
		return
	}

	homeDir, _ := os.UserHomeDir()
	plistContent := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.moltbunker.daemon</string>
    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/moltbunkerd</string>
        <string>--data=%s</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>%s/logs/daemon.log</string>
    <key>StandardErrorPath</key>
    <string>%s/logs/daemon.error.log</string>
</dict>
</plist>
`, installDataDir, installDataDir, installDataDir)

	plistPath := filepath.Join(homeDir, "Library", "LaunchAgents", "com.moltbunker.daemon.plist")
	if err := os.WriteFile(plistPath, []byte(plistContent), 0644); err != nil {
		fmt.Printf("  Failed to write launchd plist: %v\n", err)
		return
	}

	exec.Command("launchctl", "load", plistPath).Run()

	fmt.Println("  Launchd service installed")
	fmt.Println("  Start with: launchctl start com.moltbunker.daemon")
}
