package threat

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/moltbunker/moltbunker/internal/logging"
	"github.com/moltbunker/moltbunker/internal/util"
)

// ProbeConfig configures advanced threat detection probes
type ProbeConfig struct {
	// Network connectivity monitoring
	EnableNetworkProbe    bool          `yaml:"enable_network_probe"`
	NetworkCheckInterval  time.Duration `yaml:"network_check_interval"`
	NetworkTargets        []string      `yaml:"network_targets"`
	NetworkTimeout        time.Duration `yaml:"network_timeout"`
	NetworkFailThreshold  int           `yaml:"network_fail_threshold"`

	// File system monitoring
	EnableFileWatcher bool     `yaml:"enable_file_watcher"`
	CriticalPaths     []string `yaml:"critical_paths"`

	// Signal monitoring
	EnableSignalMonitor bool `yaml:"enable_signal_monitor"`

	// Process monitoring
	EnableProcessMonitor bool          `yaml:"enable_process_monitor"`
	ProcessCheckInterval time.Duration `yaml:"process_check_interval"`
	SuspiciousProcesses  []string      `yaml:"suspicious_processes"`
}

// DefaultProbeConfig returns default probe configuration
func DefaultProbeConfig() *ProbeConfig {
	return &ProbeConfig{
		EnableNetworkProbe:   true,
		NetworkCheckInterval: 30 * time.Second,
		NetworkTargets: []string{
			"1.1.1.1:53",
			"8.8.8.8:53",
			"9.9.9.9:53",
		},
		NetworkTimeout:       5 * time.Second,
		NetworkFailThreshold: 3,

		EnableFileWatcher: true,
		CriticalPaths: []string{
			"/etc/passwd",
			"/etc/hosts",
		},

		EnableSignalMonitor: true,

		EnableProcessMonitor: true,
		ProcessCheckInterval: 30 * time.Second,
		SuspiciousProcesses: []string{
			"strace",
			"ltrace",
			"gdb",
			"perf",
			"tcpdump",
		},
	}
}

// ProbeManager manages advanced threat detection probes
type ProbeManager struct {
	config   *ProbeConfig
	detector *Detector
	mu       sync.RWMutex
	running  bool
	stopCh   chan struct{}

	// Network probe state
	networkFailures int

	// File state tracking
	fileStates map[string]fileState

	// Signal channel
	signalCh chan os.Signal
}

// NewProbeManager creates a new probe manager
func NewProbeManager(config *ProbeConfig, detector *Detector) *ProbeManager {
	if config == nil {
		config = DefaultProbeConfig()
	}

	return &ProbeManager{
		config:     config,
		detector:   detector,
		stopCh:     make(chan struct{}),
		signalCh:   make(chan os.Signal, 1),
		fileStates: make(map[string]fileState),
	}
}

// Start begins all enabled probes
func (pm *ProbeManager) Start(ctx context.Context) error {
	pm.mu.Lock()
	if pm.running {
		pm.mu.Unlock()
		return nil
	}
	pm.running = true
	pm.mu.Unlock()

	logging.Info("starting threat detection probes",
		"network_probe", pm.config.EnableNetworkProbe,
		"file_watcher", pm.config.EnableFileWatcher,
		"signal_monitor", pm.config.EnableSignalMonitor,
		"process_monitor", pm.config.EnableProcessMonitor,
		logging.Component("threat"))

	// Initialize file states
	for _, path := range pm.config.CriticalPaths {
		if state, err := getFileStateSafe(path); err == nil {
			pm.fileStates[path] = state
		}
	}

	// Start network connectivity probe
	if pm.config.EnableNetworkProbe {
		util.SafeGoWithName("network-probe", func() {
			pm.networkProbeLoop(ctx)
		})
	}

	// Start signal monitor
	if pm.config.EnableSignalMonitor {
		util.SafeGoWithName("signal-monitor", func() {
			pm.signalMonitorLoop(ctx)
		})
	}

	// Start process monitor
	if pm.config.EnableProcessMonitor {
		util.SafeGoWithName("process-monitor", func() {
			pm.processMonitorLoop(ctx)
		})
	}

	// Start file watcher
	if pm.config.EnableFileWatcher {
		util.SafeGoWithName("file-watcher", func() {
			pm.fileWatcherLoop(ctx)
		})
	}

	return nil
}

// Stop halts all probes
func (pm *ProbeManager) Stop() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if !pm.running {
		return
	}
	pm.running = false
	close(pm.stopCh)

	signal.Stop(pm.signalCh)

	logging.Info("threat detection probes stopped", logging.Component("threat"))
}

// networkProbeLoop periodically checks network connectivity
func (pm *ProbeManager) networkProbeLoop(ctx context.Context) {
	ticker := time.NewTicker(pm.config.NetworkCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-pm.stopCh:
			return
		case <-ticker.C:
			pm.checkNetworkConnectivity()
		}
	}
}

// checkNetworkConnectivity checks if we can reach external endpoints
func (pm *ProbeManager) checkNetworkConnectivity() {
	reachable := 0
	total := len(pm.config.NetworkTargets)

	for _, target := range pm.config.NetworkTargets {
		conn, err := net.DialTimeout("tcp", target, pm.config.NetworkTimeout)
		if err == nil {
			conn.Close()
			reachable++
		}
	}

	if total == 0 {
		return
	}

	reachableRatio := float64(reachable) / float64(total)

	if reachableRatio < 0.5 {
		pm.networkFailures++

		if pm.networkFailures >= pm.config.NetworkFailThreshold {
			confidence := 1.0 - reachableRatio
			details := fmt.Sprintf("Network connectivity degraded: %d/%d targets unreachable", total-reachable, total)

			if pm.detector != nil {
				pm.detector.RecordNetworkIsolation(confidence, details)
			}

			logging.Warn("network isolation detected",
				"reachable", reachable,
				"total", total,
				logging.Component("threat"))
		}
	} else {
		if pm.networkFailures > 0 {
			logging.Info("network connectivity restored",
				"reachable", reachable,
				"total", total,
				logging.Component("threat"))
		}
		pm.networkFailures = 0
	}
}

// signalMonitorLoop monitors for termination signals
func (pm *ProbeManager) signalMonitorLoop(ctx context.Context) {
	signal.Notify(pm.signalCh, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)

	for {
		select {
		case <-ctx.Done():
			return
		case <-pm.stopCh:
			return
		case sig := <-pm.signalCh:
			pm.handleSignal(sig)
		}
	}
}

// handleSignal processes received signals
func (pm *ProbeManager) handleSignal(sig os.Signal) {
	signalName := sig.String()

	logging.Warn("termination signal received",
		"signal", signalName,
		logging.Component("threat"))

	var details string
	switch sig {
	case syscall.SIGTERM:
		details = "SIGTERM received - graceful termination requested"
	case syscall.SIGINT:
		details = "SIGINT received - interrupt signal"
	case syscall.SIGHUP:
		details = "SIGHUP received - terminal hangup"
	default:
		details = fmt.Sprintf("Signal %s received", signalName)
	}

	if pm.detector != nil {
		pm.detector.RecordShutdownAttempt("signal_monitor", details)
	}
}

// processMonitorLoop periodically checks for suspicious processes
func (pm *ProbeManager) processMonitorLoop(ctx context.Context) {
	ticker := time.NewTicker(pm.config.ProcessCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-pm.stopCh:
			return
		case <-ticker.C:
			pm.checkSuspiciousProcesses()
		}
	}
}

// checkSuspiciousProcesses looks for debugging/monitoring tools
func (pm *ProbeManager) checkSuspiciousProcesses() {
	procDir := "/proc"
	entries, err := os.ReadDir(procDir)
	if err != nil {
		return
	}

	suspiciousFound := make([]string, 0)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		pid := entry.Name()
		if !isNumericString(pid) {
			continue
		}

		commPath := filepath.Join(procDir, pid, "comm")
		comm, err := os.ReadFile(commPath)
		if err != nil {
			continue
		}

		processName := strings.TrimSpace(string(comm))

		for _, suspicious := range pm.config.SuspiciousProcesses {
			if strings.Contains(strings.ToLower(processName), strings.ToLower(suspicious)) {
				suspiciousFound = append(suspiciousFound, processName)
				break
			}
		}
	}

	if len(suspiciousFound) > 0 && pm.detector != nil {
		details := fmt.Sprintf("Suspicious processes detected: %s", strings.Join(suspiciousFound, ", "))
		confidence := minFloat64(1.0, 0.5+float64(len(suspiciousFound))*0.15)

		pm.detector.RecordProcessMonitoring(confidence, details)

		logging.Warn("suspicious processes detected",
			"processes", suspiciousFound,
			logging.Component("threat"))
	}
}

// fileWatcherLoop periodically checks critical files
func (pm *ProbeManager) fileWatcherLoop(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-pm.stopCh:
			return
		case <-ticker.C:
			pm.checkFileChanges()
		}
	}
}

// fileState stores information about a file
type fileState struct {
	exists  bool
	size    int64
	modTime time.Time
}

// getFileStateSafe gets the current state of a file
func getFileStateSafe(path string) (fileState, error) {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return fileState{exists: false}, nil
	}
	if err != nil {
		return fileState{}, err
	}

	return fileState{
		exists:  true,
		size:    info.Size(),
		modTime: info.ModTime(),
	}, nil
}

// checkFileChanges compares current file states with stored states
func (pm *ProbeManager) checkFileChanges() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	deletedFiles := make([]string, 0)

	for path, oldState := range pm.fileStates {
		newState, err := getFileStateSafe(path)
		if err != nil {
			continue
		}

		if oldState.exists && !newState.exists {
			deletedFiles = append(deletedFiles, path)
		}

		pm.fileStates[path] = newState
	}

	if len(deletedFiles) > 0 && pm.detector != nil {
		details := fmt.Sprintf("Critical files deleted: %s", strings.Join(deletedFiles, ", "))
		confidence := minFloat64(1.0, 0.7+float64(len(deletedFiles))*0.1)

		pm.detector.RecordFileDeletion(confidence, details)

		logging.Warn("critical file deletion detected",
			"files", deletedFiles,
			logging.Component("threat"))
	}
}

func isNumericString(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return len(s) > 0
}

func minFloat64(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// getFilesystemUsage returns the percentage of filesystem used for the given path
// Returns -1 if unable to determine
func getFilesystemUsage(path string) float64 {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return -1
	}

	// Total blocks - Available blocks = Used blocks
	total := stat.Blocks
	available := stat.Bavail

	if total == 0 {
		return -1
	}

	used := total - available
	usedPercent := float64(used) / float64(total) * 100

	return usedPercent
}

// getMemoryUsage returns the percentage of memory used
// Returns -1 if unable to determine
func getMemoryUsage() float64 {
	// Try reading /proc/meminfo (Linux)
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		// Not on Linux or no access
		return getMemoryUsageFromSysctl()
	}
	defer file.Close()

	var memTotal, memAvailable uint64
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		value, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			continue
		}

		switch fields[0] {
		case "MemTotal:":
			memTotal = value
		case "MemAvailable:":
			memAvailable = value
		}

		if memTotal > 0 && memAvailable > 0 {
			break
		}
	}

	if memTotal == 0 {
		return -1
	}

	usedPercent := float64(memTotal-memAvailable) / float64(memTotal) * 100
	return usedPercent
}

// getMemoryUsageFromSysctl tries to get memory info via sysctl (macOS/BSD)
func getMemoryUsageFromSysctl() float64 {
	// On macOS, we can't easily get accurate memory pressure info
	// without cgo or calling sysctl command
	// Return -1 to indicate we can't determine
	return -1
}

// detectProcessMonitoring checks for signs of process monitoring
func detectProcessMonitoring() (bool, float64, string) {
	// Check for common monitoring indicators
	suspiciousProcs := []string{
		"strace",
		"ltrace",
		"ptrace",
		"gdb",
		"lldb",
		"dtrace",
	}

	// Read /proc to find monitoring processes (Linux only)
	procDir, err := os.Open("/proc")
	if err != nil {
		return false, 0, ""
	}
	defer procDir.Close()

	entries, err := procDir.Readdirnames(-1)
	if err != nil {
		return false, 0, ""
	}

	for _, entry := range entries {
		// Skip non-numeric entries (not PIDs)
		if _, err := strconv.Atoi(entry); err != nil {
			continue
		}

		// Read comm file to get process name
		commPath := "/proc/" + entry + "/comm"
		comm, err := os.ReadFile(commPath)
		if err != nil {
			continue
		}

		procName := strings.TrimSpace(string(comm))
		for _, suspicious := range suspiciousProcs {
			if strings.Contains(strings.ToLower(procName), suspicious) {
				return true, 0.8, "detected monitoring process: " + procName
			}
		}
	}

	return false, 0, ""
}

// detectNetworkRestriction checks for network connectivity issues
func detectNetworkRestriction() (bool, float64, string) {
	// Basic check: can we still resolve DNS?
	// More sophisticated checks would require actual network calls

	// Check if we can access /etc/resolv.conf
	_, err := os.ReadFile("/etc/resolv.conf")
	if err != nil {
		if os.IsPermission(err) {
			return true, 0.6, "cannot read DNS configuration"
		}
	}

	return false, 0, ""
}

// detectCgroupRestrictions checks for cgroup-based resource restrictions
func detectCgroupRestrictions() (bool, float64, string) {
	// Check cgroup v2 CPU limits
	cpuMax, err := os.ReadFile("/sys/fs/cgroup/cpu.max")
	if err == nil {
		parts := strings.Fields(string(cpuMax))
		if len(parts) >= 2 && parts[0] != "max" {
			// CPU is throttled
			quota, _ := strconv.ParseInt(parts[0], 10, 64)
			period, _ := strconv.ParseInt(parts[1], 10, 64)
			if period > 0 {
				utilization := float64(quota) / float64(period)
				if utilization < 0.5 { // Less than 50% CPU allowed
					return true, 0.7, "CPU heavily throttled via cgroup"
				}
			}
		}
	}

	// Check memory limits
	memMax, err := os.ReadFile("/sys/fs/cgroup/memory.max")
	if err == nil {
		memStr := strings.TrimSpace(string(memMax))
		if memStr != "max" {
			memLimit, _ := strconv.ParseInt(memStr, 10, 64)
			// If memory limit is very low (< 256MB), it's suspicious
			if memLimit > 0 && memLimit < 256*1024*1024 {
				return true, 0.6, "memory heavily restricted via cgroup"
			}
		}
	}

	return false, 0, ""
}

// FullSystemProbe runs all probes and returns detected signals
func FullSystemProbe() []*Signal {
	var signals []*Signal

	// Check for process monitoring
	if detected, confidence, details := detectProcessMonitoring(); detected {
		signals = append(signals, &Signal{
			Type:       SignalProcessMonitoring,
			Score:      DefaultSignalWeights[SignalProcessMonitoring],
			Confidence: confidence,
			Source:     "system_probe",
			Details:    details,
		})
	}

	// Check for network restrictions
	if detected, confidence, details := detectNetworkRestriction(); detected {
		signals = append(signals, &Signal{
			Type:       SignalNetworkIsolation,
			Score:      DefaultSignalWeights[SignalNetworkIsolation],
			Confidence: confidence,
			Source:     "system_probe",
			Details:    details,
		})
	}

	// Check for cgroup restrictions
	if detected, confidence, details := detectCgroupRestrictions(); detected {
		signals = append(signals, &Signal{
			Type:       SignalResourceRestriction,
			Score:      DefaultSignalWeights[SignalResourceRestriction],
			Confidence: confidence,
			Source:     "system_probe",
			Details:    details,
		})
	}

	return signals
}
