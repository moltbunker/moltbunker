package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/moltbunker/moltbunker/internal/client"
	"github.com/spf13/cobra"
)

var (
	startForeground bool
	startPort       int
)

func NewStartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the moltbunker daemon",
		Long: `Start the moltbunker daemon.

By default, the daemon runs in the background. Use --foreground to run in the foreground.`,
		RunE: runStart,
	}

	cmd.Flags().BoolVarP(&startForeground, "foreground", "f", false, "Run in foreground")
	cmd.Flags().IntVarP(&startPort, "port", "p", 9000, "P2P port")

	return cmd
}

func runStart(cmd *cobra.Command, args []string) error {
	// Check if daemon is already running
	daemonClient := client.NewDaemonClient("")
	if daemonClient.IsDaemonRunning() {
		fmt.Println("Daemon is already running")
		return nil
	}

	homeDir, _ := os.UserHomeDir()
	dataDir := filepath.Join(homeDir, ".moltbunker")
	keyPath := filepath.Join(dataDir, "keys", "node.key")
	keystoreDir := filepath.Join(dataDir, "keystore")

	// Check if installed
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		return fmt.Errorf("moltbunker not installed. Run 'moltbunker install' first")
	}

	if startForeground {
		return startForegroundDaemon(keyPath, keystoreDir, dataDir)
	}

	return startBackgroundDaemon(keyPath, keystoreDir, dataDir)
}

func startForegroundDaemon(keyPath, keystoreDir, dataDir string) error {
	fmt.Println("Starting moltbunker daemon in foreground...")
	fmt.Printf("  Port: %d\n", startPort)
	fmt.Printf("  Data: %s\n", dataDir)
	fmt.Println("Press Ctrl+C to stop")

	// Find daemon binary
	daemonBin := findDaemonBinary()
	if daemonBin == "" {
		return fmt.Errorf("daemon binary not found")
	}

	// Execute daemon
	daemonCmd := exec.Command(daemonBin,
		"--port", fmt.Sprintf("%d", startPort),
		"--key", keyPath,
		"--keystore", keystoreDir,
		"--data", dataDir,
	)
	daemonCmd.Stdout = os.Stdout
	daemonCmd.Stderr = os.Stderr
	daemonCmd.Stdin = os.Stdin

	return daemonCmd.Run()
}

func startBackgroundDaemon(keyPath, keystoreDir, dataDir string) error {
	fmt.Println("Starting moltbunker daemon...")

	// Check for systemd/launchd first
	switch runtime.GOOS {
	case "linux":
		if isSystemdAvailable() {
			return startWithSystemd()
		}
	case "darwin":
		if isLaunchdAvailable() {
			return startWithLaunchd()
		}
	}

	// Fall back to direct process start
	return startDaemonProcess(keyPath, keystoreDir, dataDir)
}

func startDaemonProcess(keyPath, keystoreDir, dataDir string) error {
	daemonBin := findDaemonBinary()
	if daemonBin == "" {
		return fmt.Errorf("daemon binary not found. Build with 'make daemon' or install it to /usr/local/bin")
	}

	logFile := filepath.Join(dataDir, "logs", "daemon.log")
	logDir := filepath.Dir(logFile)
	os.MkdirAll(logDir, 0755)

	// Open log file
	log, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	// Start daemon process
	daemonCmd := exec.Command(daemonBin,
		"--port", fmt.Sprintf("%d", startPort),
		"--key", keyPath,
		"--keystore", keystoreDir,
		"--data", dataDir,
	)
	daemonCmd.Stdout = log
	daemonCmd.Stderr = log

	// Set process group for proper cleanup
	daemonCmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	if err := daemonCmd.Start(); err != nil {
		log.Close()
		return fmt.Errorf("failed to start daemon: %w", err)
	}

	// Write PID file
	pidFile := filepath.Join(dataDir, "daemon.pid")
	if err := os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", daemonCmd.Process.Pid)), 0644); err != nil {
		fmt.Printf("Warning: failed to write PID file: %v\n", err)
	}

	fmt.Printf("Daemon started (PID: %d)\n", daemonCmd.Process.Pid)
	fmt.Printf("  Log file: %s\n", logFile)

	// Wait a moment and check if daemon is running
	time.Sleep(2 * time.Second)
	daemonClient := client.NewDaemonClient("")
	if daemonClient.IsDaemonRunning() {
		fmt.Println("Daemon is running successfully")
	} else {
		fmt.Println("Warning: Daemon may have failed to start. Check logs.")
	}

	return nil
}

func findDaemonBinary() string {
	// Check common locations
	locations := []string{
		"./bin/moltbunker-daemon",
		"/usr/local/bin/moltbunker-daemon",
		"/usr/bin/moltbunker-daemon",
	}

	// Also check in same directory as CLI binary
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
		locations = append([]string{filepath.Join(dir, "moltbunker-daemon")}, locations...)
	}

	for _, loc := range locations {
		if _, err := os.Stat(loc); err == nil {
			return loc
		}
	}

	// Check PATH
	if path, err := exec.LookPath("moltbunker-daemon"); err == nil {
		return path
	}

	return ""
}

func isSystemdAvailable() bool {
	_, err := exec.LookPath("systemctl")
	if err != nil {
		return false
	}
	// Check if service is installed
	cmd := exec.Command("systemctl", "list-unit-files", "moltbunker.service")
	output, _ := cmd.Output()
	return len(output) > 0 && string(output) != ""
}

func isLaunchdAvailable() bool {
	homeDir, _ := os.UserHomeDir()
	plistPath := filepath.Join(homeDir, "Library", "LaunchAgents", "com.moltbunker.daemon.plist")
	_, err := os.Stat(plistPath)
	return err == nil
}

func startWithSystemd() error {
	fmt.Println("Starting via systemd...")
	cmd := exec.Command("systemctl", "start", "moltbunker")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("systemctl start failed: %w", err)
	}
	fmt.Println("Daemon started via systemd")
	return nil
}

func startWithLaunchd() error {
	fmt.Println("Starting via launchd...")
	cmd := exec.Command("launchctl", "start", "com.moltbunker.daemon")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("launchctl start failed: %w", err)
	}
	fmt.Println("Daemon started via launchd")
	return nil
}
