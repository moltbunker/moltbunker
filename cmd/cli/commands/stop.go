package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"syscall"

	"github.com/spf13/cobra"
)

func NewStopCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop the moltbunker daemon",
		Long:  "Stop the running moltbunker daemon.",
		RunE:  runStop,
	}
}

func runStop(cmd *cobra.Command, args []string) error {
	Info("Stopping moltbunker daemon...")

	// Check for systemd/launchd first
	switch runtime.GOOS {
	case "linux":
		if isSystemdServiceRunning() {
			return stopWithSystemd()
		}
	case "darwin":
		if isLaunchdServiceRunning() {
			return stopWithLaunchd()
		}
	}

	// Fall back to PID file
	return stopDaemonProcess()
}

func stopDaemonProcess() error {
	homeDir, _ := os.UserHomeDir()
	pidFile := filepath.Join(homeDir, ".moltbunker", "daemon.pid")

	data, err := os.ReadFile(pidFile)
	if err != nil {
		return fmt.Errorf("daemon not running (no PID file)")
	}

	pid, err := strconv.Atoi(string(data))
	if err != nil {
		return fmt.Errorf("invalid PID file")
	}

	// Check if process exists
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("process not found: %w", err)
	}

	// Send SIGTERM
	if err := process.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("failed to stop daemon: %w", err)
	}

	// Remove PID file
	os.Remove(pidFile)

	Success(fmt.Sprintf("Daemon stopped (PID: %d)", pid))
	return nil
}

func isSystemdServiceRunning() bool {
	cmd := exec.Command("systemctl", "is-active", "moltbunker")
	output, _ := cmd.Output()
	return string(output) == "active\n"
}

func isLaunchdServiceRunning() bool {
	cmd := exec.Command("launchctl", "list", "com.moltbunker.daemon")
	return cmd.Run() == nil
}

func stopWithSystemd() error {
	Info("Stopping via systemd...")
	cmd := exec.Command("systemctl", "stop", "moltbunker")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("systemctl stop failed: %w", err)
	}
	Success("Daemon stopped via systemd")
	return nil
}

func stopWithLaunchd() error {
	Info("Stopping via launchd...")
	cmd := exec.Command("launchctl", "stop", "com.moltbunker.daemon")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("launchctl stop failed: %w", err)
	}
	Success("Daemon stopped via launchd")
	return nil
}
