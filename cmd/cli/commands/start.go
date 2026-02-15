package commands

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/moltbunker/moltbunker/internal/client"
	"github.com/moltbunker/moltbunker/internal/identity"
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
	homeDir, _ := os.UserHomeDir()
	dataDir := filepath.Join(homeDir, ".moltbunker")
	keyPath := filepath.Join(dataDir, "keys", "node.key")
	keystoreDir := filepath.Join(dataDir, "keystore")
	pidFile := filepath.Join(dataDir, "daemon.pid")

	// Check if daemon is already running (socket check)
	daemonClient := client.NewDaemonClient(SocketPath)
	if daemonClient.IsDaemonRunning() {
		pid := readPIDFile(pidFile)
		if pid > 0 {
			Info(fmt.Sprintf("Daemon is already running (PID: %d)", pid))
		} else {
			Info("Daemon is already running")
		}
		return nil
	}

	// Socket check failed — also check PID file (catches cases where
	// the daemon is alive but the socket hasn't been created yet,
	// or the socket path is wrong)
	if pid := readPIDFile(pidFile); pid > 0 && isProcessAlive(pid) {
		Info(fmt.Sprintf("Daemon is already running (PID: %d)", pid))
		fmt.Println(Hint("Use 'moltbunker stop' to stop it, or 'moltbunker status' to check"))
		return nil
	}

	// Check if installed
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		return fmt.Errorf("moltbunker not installed. Run 'moltbunker init' first")
	}

	// Pre-flight: resolve wallet password NOW (in the CLI foreground) so any
	// macOS Keychain dialogs appear where the user can see them. The resolved
	// password is passed to the daemon via env var so it never touches the keyring.
	walletPassword, err := preflightWalletCheck(keystoreDir)
	if err != nil {
		return err
	}

	if startForeground {
		return startForegroundDaemon(keyPath, keystoreDir, dataDir, walletPassword)
	}

	return startBackgroundDaemon(keyPath, keystoreDir, dataDir, walletPassword)
}

func startForegroundDaemon(keyPath, keystoreDir, dataDir, walletPassword string) error {
	Info("Starting daemon in foreground...")
	fmt.Println(KeyValue("Port", fmt.Sprintf("%d", startPort)))
	fmt.Println(KeyValue("Data", dataDir))
	fmt.Println(Hint("Press Ctrl+C to stop"))

	// Find daemon binary
	daemonBin := findDaemonBinary()
	if daemonBin == "" {
		return fmt.Errorf("daemon binary not found")
	}

	// Build daemon arguments
	args := []string{
		"--port", fmt.Sprintf("%d", startPort),
		"--key", keyPath,
		"--keystore", keystoreDir,
		"--data", dataDir,
	}
	configPath := filepath.Join(dataDir, "config.yaml")
	if _, err := os.Stat(configPath); err == nil {
		args = append(args, "--config", configPath)
	}

	daemonCmd := exec.Command(daemonBin, args...)
	daemonCmd.Stdout = os.Stdout
	daemonCmd.Stderr = os.Stderr
	daemonCmd.Stdin = os.Stdin
	// Pass wallet password via env so daemon doesn't hit Keychain again
	daemonCmd.Env = appendWalletEnv(os.Environ(), walletPassword)

	return daemonCmd.Run()
}

func startBackgroundDaemon(keyPath, keystoreDir, dataDir, walletPassword string) error {
	Info("Starting daemon...")

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
	return startDaemonProcess(keyPath, keystoreDir, dataDir, walletPassword)
}

func startDaemonProcess(keyPath, keystoreDir, dataDir, walletPassword string) error {
	daemonBin := findDaemonBinary()
	if daemonBin == "" {
		return fmt.Errorf("daemon binary not found. Build with 'make daemon' (creates bin/moltbunkerd) or run 'make install'")
	}

	logFile := filepath.Join(dataDir, "logs", "daemon.log")
	logDir := filepath.Dir(logFile)
	os.MkdirAll(logDir, 0755)

	// Record log position before starting so we can read new output on failure
	logOffset := getFileSize(logFile)

	// Open log file for daemon output
	logFD, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	// Build daemon arguments
	args := []string{
		"--port", fmt.Sprintf("%d", startPort),
		"--key", keyPath,
		"--keystore", keystoreDir,
		"--data", dataDir,
	}
	configPath := filepath.Join(dataDir, "config.yaml")
	if _, err := os.Stat(configPath); err == nil {
		args = append(args, "--config", configPath)
	}

	daemonCmd := exec.Command(daemonBin, args...)
	daemonCmd.Stdout = logFD
	daemonCmd.Stderr = logFD
	// Pass wallet password via env so daemon doesn't hit Keychain again
	daemonCmd.Env = appendWalletEnv(os.Environ(), walletPassword)

	// Set process group for proper cleanup
	daemonCmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	if err := daemonCmd.Start(); err != nil {
		logFD.Close()
		return fmt.Errorf("failed to start daemon: %w", err)
	}

	pid := daemonCmd.Process.Pid

	// Do NOT write PID file here — the daemon manages its own PID file
	// in checkAndWritePIDFile(). Writing it here causes a race condition
	// where the daemon detects its own PID and exits with "already running".

	// Wait for daemon to either establish itself or exit
	time.Sleep(2 * time.Second)
	logFD.Close()

	// Check if daemon is responding via socket
	daemonClient := client.NewDaemonClient(SocketPath)
	if daemonClient.IsDaemonRunning() {
		Success(fmt.Sprintf("Daemon started (PID: %d)", pid))
		fmt.Println(KeyValue("Log file", logFile))
		return nil
	}

	// Daemon didn't respond — check if the process is still alive
	if isProcessAlive(pid) {
		// Process alive but socket not ready yet — might just need more time
		Info(fmt.Sprintf("Daemon starting (PID: %d), waiting for socket...", pid))
		// Give it a few more seconds
		for i := 0; i < 3; i++ {
			time.Sleep(time.Second)
			if daemonClient.IsDaemonRunning() {
				Success("Daemon is running")
				fmt.Println(KeyValue("Log file", logFile))
				return nil
			}
		}
		// Still not ready but process is alive
		Warning("Daemon process is running but not yet responding")
		fmt.Println(KeyValue("PID", fmt.Sprintf("%d", pid)))
		fmt.Println(KeyValue("Log file", logFile))
		fmt.Println(Hint("Check logs for progress: tail -f " + logFile))
		return nil
	}

	// Process died — read the log to find out why
	reason := readDaemonFailureReason(logFile, logOffset)
	if reason != "" {
		return fmt.Errorf("daemon failed to start: %s", reason)
	}

	return fmt.Errorf("daemon exited unexpectedly. Check logs: %s", logFile)
}

func findDaemonBinary() string {
	// Check common locations
	locations := []string{
		"./bin/moltbunkerd",
		"./daemon",
		"/usr/local/bin/moltbunkerd",
		"/usr/bin/moltbunkerd",
	}

	// Also check in same directory as CLI binary
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
		locations = append([]string{filepath.Join(dir, "moltbunkerd")}, locations...)
	}

	for _, loc := range locations {
		if _, err := os.Stat(loc); err == nil {
			return loc
		}
	}

	// Check PATH
	if path, err := exec.LookPath("moltbunkerd"); err == nil {
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
	Info("Starting via systemd...")
	cmd := exec.Command("systemctl", "start", "moltbunker")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("systemctl start failed: %w", err)
	}
	Success("Daemon started via systemd")
	return nil
}

func startWithLaunchd() error {
	Info("Starting via launchd...")
	cmd := exec.Command("launchctl", "start", "com.moltbunker.daemon")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("launchctl start failed: %w", err)
	}
	Success("Daemon started via launchd")
	return nil
}

// preflightWalletCheck verifies wallet and password availability before starting
// the daemon. The daemon requires an unlocked wallet for all roles (P2P signing,
// announce, payments). This runs in the CLI (foreground) so any macOS Keychain
// dialogs appear where the user can see and respond to them.
// Returns the resolved password so it can be passed to the daemon via env var,
// avoiding a second Keychain access from the daemon process.
func preflightWalletCheck(keystoreDir string) (string, error) {
	// Check if wallet exists
	wm, err := identity.LoadWalletManager(keystoreDir)
	if err != nil {
		return "", fmt.Errorf("failed to check wallet: %w\n  Create one with: moltbunker wallet create", err)
	}
	if wm == nil {
		return "", fmt.Errorf("no wallet found — the daemon requires a wallet to operate.\n  Create one with: moltbunker wallet create")
	}

	// Check if password is accessible (this triggers macOS Keychain dialog if needed)
	pw, _ := identity.RetrieveWalletPassword()
	if pw == "" {
		pw, _ = identity.RetrieveKernelKeyring()
	}
	if pw == "" {
		pw = os.Getenv("MOLTBUNKER_WALLET_PASSWORD")
	}
	if pw == "" {
		return "", fmt.Errorf("wallet password not available — the daemon needs it to unlock the wallet.\n" +
			"  Options:\n" +
			"    1. Store in keyring:  moltbunker wallet create (saves password automatically)\n" +
			"    2. Environment var:   export MOLTBUNKER_WALLET_PASSWORD=<password>\n" +
			"    3. Password file:     set node.wallet_password_file in config.yaml")
	}

	return pw, nil
}

// appendWalletEnv returns a copy of env with MOLTBUNKER_WALLET_PASSWORD set.
// This passes the already-resolved password to the daemon so it doesn't need
// to access the Keychain again (avoids duplicate macOS password dialogs).
// The daemon clears this env var from its own process after reading it.
func appendWalletEnv(env []string, password string) []string {
	// Remove any existing MOLTBUNKER_WALLET_PASSWORD to avoid duplicates
	filtered := make([]string, 0, len(env)+1)
	for _, e := range env {
		if !strings.HasPrefix(e, "MOLTBUNKER_WALLET_PASSWORD=") {
			filtered = append(filtered, e)
		}
	}
	return append(filtered, "MOLTBUNKER_WALLET_PASSWORD="+password)
}

// readPIDFile reads a PID from the given file path. Returns 0 on any error.
func readPIDFile(path string) int {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil || pid <= 0 {
		return 0
	}
	return pid
}

// isProcessAlive checks if a process with the given PID exists and is running.
func isProcessAlive(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	// On Unix, FindProcess always succeeds. Send signal 0 to check.
	return process.Signal(syscall.Signal(0)) == nil
}

// getFileSize returns the current size of a file, or 0 if it doesn't exist.
func getFileSize(path string) int64 {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return info.Size()
}

// readDaemonFailureReason reads log output written after startOffset and
// extracts a human-readable failure reason from common error patterns.
func readDaemonFailureReason(logFile string, startOffset int64) string {
	f, err := os.Open(logFile)
	if err != nil {
		return ""
	}
	defer f.Close()

	// Seek to where the daemon started writing
	if startOffset > 0 {
		f.Seek(startOffset, 0)
	}

	// Known error patterns → user-friendly messages
	patterns := map[string]string{
		"already running":        "another daemon is already running",
		"address already in use": "port is already in use",
		"permission denied":      "permission denied (try running with sudo or check file permissions)",
		"no such file":           "required file not found",
		"bind: address":          "port binding failed",
		"no wallet found":        "no wallet found (create one with: moltbunker wallet create)",
		"no wallet password":     "wallet password not available (store with: moltbunker wallet create, or set MOLTBUNKER_WALLET_PASSWORD)",
		"failed to unlock wallet": "wrong wallet password (check keyring or MOLTBUNKER_WALLET_PASSWORD)",
	}

	scanner := bufio.NewScanner(f)
	var lastLine string
	for scanner.Scan() {
		line := scanner.Text()
		lastLine = line
		lower := strings.ToLower(line)
		for pattern, msg := range patterns {
			if strings.Contains(lower, pattern) {
				return msg
			}
		}
	}

	// No known pattern matched — return the last log line (trimmed)
	if lastLine != "" {
		// Strip common log prefixes (timestamp, level)
		if idx := strings.Index(lastLine, "msg="); idx >= 0 {
			return strings.Trim(lastLine[idx+4:], "\"")
		}
		// Limit length for readability
		if len(lastLine) > 120 {
			lastLine = lastLine[:120] + "..."
		}
		return lastLine
	}

	return ""
}
