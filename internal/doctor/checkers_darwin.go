//go:build darwin

package doctor

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
)

// MinGoVersion is the minimum required Go version
const MinGoVersion = "1.24"

// MinDiskSpaceGB is the minimum required disk space in GB
const MinDiskSpaceGB = 10

// MinMemoryGB is the minimum recommended memory in GB
const MinMemoryGB = 4

// GoVersionChecker checks if Go is installed with the correct version
type GoVersionChecker struct{}

func NewGoVersionChecker() *GoVersionChecker {
	return &GoVersionChecker{}
}

func (c *GoVersionChecker) Name() string     { return "Go version" }
func (c *GoVersionChecker) Category() Category { return CategoryRuntime }
func (c *GoVersionChecker) CanFix() bool     { return true }

func (c *GoVersionChecker) Check(ctx context.Context) CheckResult {
	result := CheckResult{
		Name:       c.Name(),
		Category:   c.Category(),
		Fixable:    true,
		FixCommand: "brew install go",
		FixPackage: "go",
	}

	// Check if go is installed
	goPath, err := exec.LookPath("go")
	if err != nil {
		result.Status = StatusError
		result.Message = "Go: Not installed"
		return result
	}

	// Get version
	out, err := exec.CommandContext(ctx, goPath, "version").Output()
	if err != nil {
		result.Status = StatusError
		result.Message = "Go: Failed to get version"
		result.Details = err.Error()
		return result
	}

	// Parse version (output: "go version go1.24.6 darwin/arm64")
	versionStr := string(out)
	parts := strings.Fields(versionStr)
	if len(parts) < 3 {
		result.Status = StatusError
		result.Message = "Go: Unable to parse version"
		return result
	}

	version := strings.TrimPrefix(parts[2], "go")
	major, minor := parseGoVersion(version)
	minMajor, minMinor := parseGoVersion(MinGoVersion)

	if major > minMajor || (major == minMajor && minor >= minMinor) {
		result.Status = StatusOK
		result.Message = fmt.Sprintf("Go version: %s (>= %s required)", version, MinGoVersion)
		result.Fixable = false
	} else {
		result.Status = StatusError
		result.Message = fmt.Sprintf("Go version: %s (>= %s required)", version, MinGoVersion)
		result.Details = "Your Go version is too old"
	}

	return result
}

func (c *GoVersionChecker) Fix(ctx context.Context, pm PackageManager) error {
	return pm.Install(ctx, "go")
}

// parseGoVersion extracts major.minor from a Go version string
func parseGoVersion(v string) (major, minor int) {
	parts := strings.Split(v, ".")
	if len(parts) >= 1 {
		major, _ = strconv.Atoi(parts[0])
	}
	if len(parts) >= 2 {
		minor, _ = strconv.Atoi(parts[1])
	}
	return
}

// ColimaChecker checks if Colima is installed and running (macOS only)
type ColimaChecker struct{}

func NewColimaChecker() *ColimaChecker {
	return &ColimaChecker{}
}

func (c *ColimaChecker) Name() string       { return "Colima" }
func (c *ColimaChecker) Category() Category { return CategoryRuntime }
func (c *ColimaChecker) CanFix() bool       { return true }

func (c *ColimaChecker) Check(ctx context.Context) CheckResult {
	result := CheckResult{
		Name:       c.Name(),
		Category:   c.Category(),
		Fixable:    true,
		FixCommand: "brew install colima && colima start",
		FixPackage: "colima",
	}

	// Check if colima is installed
	_, err := exec.LookPath("colima")
	if err != nil {
		result.Status = StatusError
		result.Message = "Colima: Not installed"
		result.Details = "Colima provides containerd on macOS"
		return result
	}

	// Check if colima is running
	out, err := exec.CommandContext(ctx, "colima", "status").CombinedOutput()
	outLower := strings.ToLower(string(out))
	if err != nil || !strings.Contains(outLower, "running") {
		result.Status = StatusWarning
		result.Message = "Colima: Installed but not running"
		result.Details = "Run 'colima start' to start"
		result.FixCommand = "colima start"
		return result
	}

	result.Status = StatusOK
	result.Message = "Colima: Running"
	result.Fixable = false
	return result
}

func (c *ColimaChecker) Fix(ctx context.Context, pm PackageManager) error {
	// Install if not present
	if _, err := exec.LookPath("colima"); err != nil {
		if err := pm.Install(ctx, "colima"); err != nil {
			return err
		}
	}
	// Start colima
	return exec.CommandContext(ctx, "colima", "start").Run()
}

// ContainerdChecker checks if containerd is installed
type ContainerdChecker struct{}

func NewContainerdChecker() *ContainerdChecker {
	return &ContainerdChecker{}
}

func (c *ContainerdChecker) Name() string     { return "containerd" }
func (c *ContainerdChecker) Category() Category { return CategoryRuntime }
func (c *ContainerdChecker) CanFix() bool     { return true }

func (c *ContainerdChecker) Check(ctx context.Context) CheckResult {
	result := CheckResult{
		Name:       c.Name(),
		Category:   c.Category(),
		Fixable:    true,
		FixCommand: "brew install colima && colima start",
		FixPackage: "colima",
	}

	// On macOS, containerd runs inside Colima
	if runtime.GOOS == "darwin" {
		// Check if we can connect to containerd via Colima
		homeDir, _ := os.UserHomeDir()

		// Check Colima docker socket (most common)
		colimaDockerSocket := filepath.Join(homeDir, ".colima", "default", "docker.sock")
		if _, err := os.Stat(colimaDockerSocket); err == nil {
			result.Status = StatusOK
			result.Message = "Containerd: Available via Colima"
			result.Fixable = false
			return result
		}

		// Check Colima containerd socket
		colimaContainerdSocket := filepath.Join(homeDir, ".colima", "default", "containerd.sock")
		if _, err := os.Stat(colimaContainerdSocket); err == nil {
			result.Status = StatusOK
			result.Message = "Containerd: Available via Colima"
			result.Fixable = false
			return result
		}

		// Check Docker Desktop socket as fallback
		if _, err := os.Stat("/var/run/docker.sock"); err == nil {
			result.Status = StatusOK
			result.Message = "Containerd: Available via Docker Desktop"
			result.Fixable = false
			return result
		}

		result.Status = StatusError
		result.Message = "Containerd: Not available (Colima not running)"
		result.Details = "Run 'colima start' to enable containerd"
		return result
	}

	// Linux: check native containerd
	_, err := exec.LookPath("containerd")
	if err != nil {
		// Also check for ctr as an alternative
		_, err2 := exec.LookPath("ctr")
		if err2 != nil {
			result.Status = StatusError
			result.Message = "Containerd: Not installed"
			result.FixCommand = "apt install containerd"
			result.FixPackage = "containerd"
			return result
		}
	}

	result.Status = StatusOK
	result.Message = "Containerd: Installed"
	result.Fixable = false
	return result
}

func (c *ContainerdChecker) Fix(ctx context.Context, pm PackageManager) error {
	return pm.Install(ctx, "containerd")
}

// TorChecker checks if Tor is installed
type TorChecker struct{}

func NewTorChecker() *TorChecker {
	return &TorChecker{}
}

func (c *TorChecker) Name() string     { return "Tor" }
func (c *TorChecker) Category() Category { return CategoryServices }
func (c *TorChecker) CanFix() bool     { return true }

func (c *TorChecker) Check(ctx context.Context) CheckResult {
	result := CheckResult{
		Name:       c.Name(),
		Category:   c.Category(),
		Fixable:    true,
		FixCommand: "brew install tor",
		FixPackage: "tor",
	}

	_, err := exec.LookPath("tor")
	if err != nil {
		result.Status = StatusError
		result.Message = "Tor: Not installed"
		return result
	}

	result.Status = StatusOK
	result.Message = "Tor: Installed"
	result.Fixable = false
	return result
}

func (c *TorChecker) Fix(ctx context.Context, pm PackageManager) error {
	return pm.Install(ctx, "tor")
}

// IPFSChecker checks if IPFS is installed
type IPFSChecker struct{}

func NewIPFSChecker() *IPFSChecker {
	return &IPFSChecker{}
}

func (c *IPFSChecker) Name() string     { return "IPFS" }
func (c *IPFSChecker) Category() Category { return CategoryServices }
func (c *IPFSChecker) CanFix() bool     { return true }

func (c *IPFSChecker) Check(ctx context.Context) CheckResult {
	result := CheckResult{
		Name:       c.Name(),
		Category:   c.Category(),
		Fixable:    true,
		FixCommand: "brew install ipfs",
		FixPackage: "ipfs",
	}

	_, err := exec.LookPath("ipfs")
	if err != nil {
		result.Status = StatusError
		result.Message = "IPFS: Not installed"
		return result
	}

	result.Status = StatusOK
	result.Message = "IPFS: Installed"
	result.Fixable = false
	return result
}

func (c *IPFSChecker) Fix(ctx context.Context, pm PackageManager) error {
	return pm.Install(ctx, "ipfs")
}

// DiskSpaceChecker checks available disk space
type DiskSpaceChecker struct{}

func NewDiskSpaceChecker() *DiskSpaceChecker {
	return &DiskSpaceChecker{}
}

func (c *DiskSpaceChecker) Name() string     { return "Disk space" }
func (c *DiskSpaceChecker) Category() Category { return CategorySystem }
func (c *DiskSpaceChecker) CanFix() bool     { return false }

func (c *DiskSpaceChecker) Check(ctx context.Context) CheckResult {
	result := CheckResult{
		Name:     c.Name(),
		Category: c.Category(),
		Fixable:  false,
	}

	// Get home directory for checking space
	homeDir, err := os.UserHomeDir()
	if err != nil {
		result.Status = StatusError
		result.Message = "Disk space: Unable to check"
		result.Details = err.Error()
		return result
	}

	var stat syscall.Statfs_t
	if err := syscall.Statfs(homeDir, &stat); err != nil {
		result.Status = StatusError
		result.Message = "Disk space: Unable to check"
		result.Details = err.Error()
		return result
	}

	// Available space in bytes
	availableBytes := stat.Bavail * uint64(stat.Bsize)
	availableGB := float64(availableBytes) / (1024 * 1024 * 1024)

	if availableGB >= MinDiskSpaceGB {
		result.Status = StatusOK
		result.Message = fmt.Sprintf("Disk space: %.1f GB available (>= %d GB required)", availableGB, MinDiskSpaceGB)
	} else if availableGB >= MinDiskSpaceGB/2 {
		result.Status = StatusWarning
		result.Message = fmt.Sprintf("Disk space: %.1f GB available (>= %d GB recommended)", availableGB, MinDiskSpaceGB)
		result.Details = "Consider freeing up some disk space"
	} else {
		result.Status = StatusError
		result.Message = fmt.Sprintf("Disk space: %.1f GB available (>= %d GB required)", availableGB, MinDiskSpaceGB)
		result.Details = "Insufficient disk space for container storage"
	}

	return result
}

func (c *DiskSpaceChecker) Fix(ctx context.Context, pm PackageManager) error {
	return fmt.Errorf("disk space issues cannot be auto-fixed")
}

// MemoryChecker checks available system memory
type MemoryChecker struct{}

func NewMemoryChecker() *MemoryChecker {
	return &MemoryChecker{}
}

func (c *MemoryChecker) Name() string     { return "Memory" }
func (c *MemoryChecker) Category() Category { return CategorySystem }
func (c *MemoryChecker) CanFix() bool     { return false }

func (c *MemoryChecker) Check(ctx context.Context) CheckResult {
	result := CheckResult{
		Name:     c.Name(),
		Category: c.Category(),
		Fixable:  false,
	}

	// On macOS, use sysctl to get memory info
	out, err := exec.CommandContext(ctx, "sysctl", "-n", "hw.memsize").Output()
	if err != nil {
		result.Status = StatusWarning
		result.Message = "Memory: Unable to check"
		result.Details = err.Error()
		return result
	}

	memBytes, err := strconv.ParseUint(strings.TrimSpace(string(out)), 10, 64)
	if err != nil {
		result.Status = StatusWarning
		result.Message = "Memory: Unable to parse"
		return result
	}

	memGB := float64(memBytes) / (1024 * 1024 * 1024)

	if memGB >= MinMemoryGB {
		result.Status = StatusOK
		result.Message = fmt.Sprintf("Memory: %.0f GB total (>= %d GB recommended)", memGB, MinMemoryGB)
	} else {
		result.Status = StatusWarning
		result.Message = fmt.Sprintf("Memory: %.0f GB total (>= %d GB recommended)", memGB, MinMemoryGB)
		result.Details = "Low memory may affect container performance"
	}

	return result
}

func (c *MemoryChecker) Fix(ctx context.Context, pm PackageManager) error {
	return fmt.Errorf("memory issues cannot be auto-fixed")
}

// ConfigFileChecker checks if the config file exists
type ConfigFileChecker struct{}

func NewConfigFileChecker() *ConfigFileChecker {
	return &ConfigFileChecker{}
}

func (c *ConfigFileChecker) Name() string     { return "Config file" }
func (c *ConfigFileChecker) Category() Category { return CategoryConfig }
func (c *ConfigFileChecker) CanFix() bool     { return true }

func (c *ConfigFileChecker) Check(ctx context.Context) CheckResult {
	result := CheckResult{
		Name:     c.Name(),
		Category: c.Category(),
		Fixable:  true,
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		result.Status = StatusError
		result.Message = "Config file: Unable to check"
		result.Details = err.Error()
		return result
	}

	configPath := filepath.Join(homeDir, ".moltbunker", "config.yaml")
	result.FixCommand = fmt.Sprintf("moltbunker config init")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		result.Status = StatusWarning
		result.Message = "Config file: Not found (will use defaults)"
		result.Details = fmt.Sprintf("Expected at: %s", configPath)
		return result
	}

	result.Status = StatusOK
	result.Message = fmt.Sprintf("Config file: %s", configPath)
	result.Fixable = false
	return result
}

func (c *ConfigFileChecker) Fix(ctx context.Context, pm PackageManager) error {
	// Create default config directory and file
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configDir := filepath.Join(homeDir, ".moltbunker")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	// The actual config initialization should be handled by the config package
	// This is a minimal fix
	return nil
}

// SocketPermissionChecker checks socket permissions
type SocketPermissionChecker struct{}

func NewSocketPermissionChecker() *SocketPermissionChecker {
	return &SocketPermissionChecker{}
}

func (c *SocketPermissionChecker) Name() string     { return "Socket permissions" }
func (c *SocketPermissionChecker) Category() Category { return CategoryPermissions }
func (c *SocketPermissionChecker) CanFix() bool     { return false }

func (c *SocketPermissionChecker) Check(ctx context.Context) CheckResult {
	result := CheckResult{
		Name:     c.Name(),
		Category: c.Category(),
		Fixable:  false,
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		result.Status = StatusError
		result.Message = "Socket permissions: Unable to check"
		result.Details = err.Error()
		return result
	}

	socketDir := filepath.Join(homeDir, ".moltbunker")

	// Check if directory exists
	info, err := os.Stat(socketDir)
	if os.IsNotExist(err) {
		result.Status = StatusOK
		result.Message = "Socket permissions: Directory will be created on first run"
		return result
	}

	if err != nil {
		result.Status = StatusError
		result.Message = "Socket permissions: Unable to check"
		result.Details = err.Error()
		return result
	}

	// Check if we can write to the directory
	if !info.IsDir() {
		result.Status = StatusError
		result.Message = "Socket permissions: Path is not a directory"
		result.Details = socketDir
		return result
	}

	// Try to check if we have write permissions
	testFile := filepath.Join(socketDir, ".write_test")
	f, err := os.Create(testFile)
	if err != nil {
		result.Status = StatusError
		result.Message = "Socket permissions: Cannot write to socket directory"
		result.Details = err.Error()
		return result
	}
	f.Close()
	os.Remove(testFile)

	result.Status = StatusOK
	result.Message = "Socket permissions: OK"

	// Check containerd socket on macOS (usually via colima or lima)
	containerdSocket := "/run/containerd/containerd.sock"
	if runtime.GOOS == "darwin" {
		// On macOS, containerd often runs in a VM (via colima/lima)
		// Check for colima socket
		colimaSocket := filepath.Join(homeDir, ".colima", "default", "containerd.sock")
		if _, err := os.Stat(colimaSocket); err == nil {
			containerdSocket = colimaSocket
		}
	}

	if _, err := os.Stat(containerdSocket); err != nil {
		result.Details = fmt.Sprintf("Note: containerd socket not found at %s (may be normal on macOS)", containerdSocket)
	}

	return result
}

func (c *SocketPermissionChecker) Fix(ctx context.Context, pm PackageManager) error {
	return fmt.Errorf("socket permission issues cannot be auto-fixed")
}
