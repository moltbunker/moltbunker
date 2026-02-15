package runtime

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/moltbunker/moltbunker/internal/logging"
)

// ProbeType defines the type of health probe
type ProbeType string

const (
	// ProbeHTTP makes an HTTP GET request and checks for 2xx status
	ProbeHTTP ProbeType = "http"
	// ProbeTCP attempts a TCP connection to a port
	ProbeTCP ProbeType = "tcp"
	// ProbeExec runs a command inside the container and checks exit code
	ProbeExec ProbeType = "exec"
)

// ProbeState represents the current state of a health probe
type ProbeState string

const (
	ProbeStateUnknown   ProbeState = "unknown"
	ProbeStateHealthy   ProbeState = "healthy"
	ProbeStateUnhealthy ProbeState = "unhealthy"
)

// HealthProbeConfig defines the configuration for a health probe
type HealthProbeConfig struct {
	Type ProbeType

	// HTTP specific
	HTTPPath          string
	HTTPPort          int
	HTTPExpectedCodes []int // defaults to [200]

	// TCP specific
	TCPPort int

	// Exec specific
	ExecCommand []string

	// Common timing parameters
	InitialDelay     time.Duration
	Interval         time.Duration
	Timeout          time.Duration
	SuccessThreshold int // consecutive successes needed to become healthy
	FailureThreshold int // consecutive failures needed to become unhealthy
}

// ProbeStatus represents the current status of a health probe for a container
type ProbeStatus struct {
	ContainerID       string
	State             ProbeState
	ConsecutivePass   int
	ConsecutiveFail   int
	LastCheckTime     time.Time
	LastTransitionTime time.Time
	LastError         string
	TotalChecks       int
	TotalSuccesses    int
	TotalFailures     int
}

// ExecFunc is a function that executes a command inside a container.
// It returns the exit code and any error. When containerd is unavailable
// (mock mode), this can be set to a no-op that returns success.
type ExecFunc func(ctx context.Context, containerID string, cmd []string) (exitCode int, err error)

// ContainerAddressFunc resolves a container ID to its network address (host:port).
// This allows the HealthChecker to work with different networking models.
type ContainerAddressFunc func(containerID string) (host string, err error)

// probeEntry holds per-container probe state and configuration
type probeEntry struct {
	config    HealthProbeConfig
	status    ProbeStatus
	startedAt time.Time
	mu        sync.RWMutex
}

// HealthChecker manages health probes for containers
type HealthChecker struct {
	probes       map[string]*probeEntry
	mu           sync.RWMutex
	execFunc     ExecFunc
	addressFunc  ContainerAddressFunc
	httpClient   *http.Client
	running      bool
	stopCh       chan struct{}
	doneCh       chan struct{}
	nowFunc      func() time.Time // injectable clock for testing
}

// NewHealthChecker creates a new HealthChecker
func NewHealthChecker() *HealthChecker {
	return &HealthChecker{
		probes: make(map[string]*probeEntry),
		httpClient: &http.Client{
			// Per-request timeouts are set in CheckHealth; this is a safety net.
			Timeout: 30 * time.Second,
			// Do not follow redirects for health checks.
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		stopCh:  make(chan struct{}),
		doneCh:  make(chan struct{}),
		nowFunc: time.Now,
	}
}

// SetExecFunc sets the function used to execute commands inside containers.
// When containerd is not available, set this to a mock/no-op function.
func (hc *HealthChecker) SetExecFunc(fn ExecFunc) {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	hc.execFunc = fn
}

// SetAddressFunc sets the function used to resolve container IDs to network addresses.
func (hc *HealthChecker) SetAddressFunc(fn ContainerAddressFunc) {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	hc.addressFunc = fn
}

// RegisterProbe registers a health probe for a container.
// If a probe already exists for the container, it is replaced.
func (hc *HealthChecker) RegisterProbe(containerID string, config HealthProbeConfig) {
	// Apply defaults
	config = applyDefaults(config)

	entry := &probeEntry{
		config:    config,
		startedAt: hc.now(),
		status: ProbeStatus{
			ContainerID:        containerID,
			State:              ProbeStateUnknown,
			LastTransitionTime: hc.now(),
		},
	}

	hc.mu.Lock()
	defer hc.mu.Unlock()
	hc.probes[containerID] = entry

	logging.Debug("health probe registered",
		"container_id", containerID,
		"probe_type", string(config.Type),
	)
}

// UnregisterProbe removes the health probe for a container.
func (hc *HealthChecker) UnregisterProbe(containerID string) {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	delete(hc.probes, containerID)

	logging.Debug("health probe unregistered",
		"container_id", containerID,
	)
}

// CheckHealth runs a single health check for the given container.
// Returns true if the probe succeeded, false otherwise.
func (hc *HealthChecker) CheckHealth(ctx context.Context, containerID string) (bool, error) {
	hc.mu.RLock()
	entry, exists := hc.probes[containerID]
	hc.mu.RUnlock()

	if !exists {
		return false, fmt.Errorf("no probe registered for container: %s", containerID)
	}

	// Create a timeout context for the probe
	entry.mu.RLock()
	timeout := entry.config.Timeout
	entry.mu.RUnlock()

	probeCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Execute the probe
	var success bool
	var probeErr error

	entry.mu.RLock()
	probeType := entry.config.Type
	entry.mu.RUnlock()

	switch probeType {
	case ProbeHTTP:
		success, probeErr = hc.runHTTPProbe(probeCtx, containerID, entry)
	case ProbeTCP:
		success, probeErr = hc.runTCPProbe(probeCtx, containerID, entry)
	case ProbeExec:
		success, probeErr = hc.runExecProbe(probeCtx, containerID, entry)
	default:
		return false, fmt.Errorf("unknown probe type: %s", probeType)
	}

	// Update probe status
	hc.updateProbeStatus(entry, containerID, success, probeErr)

	return success, probeErr
}

// GetProbeStatus returns the current probe status for a container.
// Returns nil if no probe is registered for the container.
func (hc *HealthChecker) GetProbeStatus(containerID string) *ProbeStatus {
	hc.mu.RLock()
	entry, exists := hc.probes[containerID]
	hc.mu.RUnlock()

	if !exists {
		return nil
	}

	entry.mu.RLock()
	defer entry.mu.RUnlock()

	// Return a copy to avoid data races
	status := entry.status
	return &status
}

// Start begins the background health checking loop. It runs health checks
// at each probe's configured interval. Start blocks until ctx is cancelled
// or Stop is called.
func (hc *HealthChecker) Start(ctx context.Context) {
	hc.mu.Lock()
	if hc.running {
		hc.mu.Unlock()
		return
	}
	hc.running = true
	hc.mu.Unlock()

	logging.Info("health checker started")
	defer func() {
		hc.mu.Lock()
		hc.running = false
		hc.mu.Unlock()
		close(hc.doneCh)
		logging.Info("health checker stopped")
	}()

	// Use a short tick interval to check if any probes are due.
	// Each probe tracks its own interval independently.
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-hc.stopCh:
			return
		case <-ticker.C:
			hc.runDueProbes(ctx)
		}
	}
}

// Stop signals the background health checking loop to stop.
func (hc *HealthChecker) Stop() {
	hc.mu.Lock()
	if !hc.running {
		hc.mu.Unlock()
		return
	}
	hc.running = false
	// Close channel under lock to prevent double-close race
	close(hc.stopCh)
	hc.mu.Unlock()

	// Wait for the background loop to finish (outside lock to avoid deadlock)
	<-hc.doneCh
}

// Reset resets the health checker for reuse after Stop.
func (hc *HealthChecker) Reset() {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	if hc.running {
		return
	}

	hc.stopCh = make(chan struct{})
	hc.doneCh = make(chan struct{})
}

// IsRunning returns whether the background loop is active.
func (hc *HealthChecker) IsRunning() bool {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	return hc.running
}

// ProbeCount returns the number of registered probes.
func (hc *HealthChecker) ProbeCount() int {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	return len(hc.probes)
}

// --- Internal methods ---

func (hc *HealthChecker) now() time.Time {
	if hc.nowFunc != nil {
		return hc.nowFunc()
	}
	return time.Now()
}

// runDueProbes checks all registered probes and runs any that are due.
func (hc *HealthChecker) runDueProbes(ctx context.Context) {
	hc.mu.RLock()
	// Build a snapshot of container IDs and entries to avoid holding lock during probes
	type probeSnapshot struct {
		containerID string
		entry       *probeEntry
	}
	var due []probeSnapshot
	now := hc.now()

	for containerID, entry := range hc.probes {
		entry.mu.RLock()
		sinceStart := now.Sub(entry.startedAt)
		initialDelay := entry.config.InitialDelay
		interval := entry.config.Interval
		lastCheck := entry.status.LastCheckTime
		entry.mu.RUnlock()

		// Skip if still in initial delay period
		if sinceStart < initialDelay {
			continue
		}

		// Check if it's time for the next probe
		if lastCheck.IsZero() || now.Sub(lastCheck) >= interval {
			due = append(due, probeSnapshot{containerID: containerID, entry: entry})
		}
	}
	hc.mu.RUnlock()

	// Run due probes outside the main lock
	for _, snap := range due {
		select {
		case <-ctx.Done():
			return
		default:
		}

		_, _ = hc.CheckHealth(ctx, snap.containerID)
	}
}

// runHTTPProbe performs an HTTP GET health probe.
func (hc *HealthChecker) runHTTPProbe(ctx context.Context, containerID string, entry *probeEntry) (bool, error) {
	entry.mu.RLock()
	path := entry.config.HTTPPath
	port := entry.config.HTTPPort
	expectedCodes := entry.config.HTTPExpectedCodes
	entry.mu.RUnlock()

	host, err := hc.resolveAddress(containerID)
	if err != nil {
		return false, fmt.Errorf("failed to resolve container address: %w", err)
	}

	url := fmt.Sprintf("http://%s:%d%s", host, port, path)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return false, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	resp, err := hc.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("HTTP probe failed: %w", err)
	}
	defer resp.Body.Close()

	// Check if status code matches expected codes
	for _, code := range expectedCodes {
		if resp.StatusCode == code {
			return true, nil
		}
	}

	// If no specific codes were set, accept any 2xx
	if len(expectedCodes) == 0 {
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return true, nil
		}
	}

	return false, fmt.Errorf("HTTP probe returned status %d", resp.StatusCode)
}

// runTCPProbe attempts a TCP connection to the container.
func (hc *HealthChecker) runTCPProbe(ctx context.Context, containerID string, entry *probeEntry) (bool, error) {
	entry.mu.RLock()
	port := entry.config.TCPPort
	timeout := entry.config.Timeout
	entry.mu.RUnlock()

	host, err := hc.resolveAddress(containerID)
	if err != nil {
		return false, fmt.Errorf("failed to resolve container address: %w", err)
	}

	addr := fmt.Sprintf("%s:%d", host, port)

	dialer := &net.Dialer{
		Timeout: timeout,
	}

	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return false, fmt.Errorf("TCP probe failed: %w", err)
	}
	conn.Close()

	return true, nil
}

// runExecProbe runs a command inside the container and checks exit code.
func (hc *HealthChecker) runExecProbe(ctx context.Context, containerID string, entry *probeEntry) (bool, error) {
	hc.mu.RLock()
	execFunc := hc.execFunc
	hc.mu.RUnlock()

	if execFunc == nil {
		// No exec function available (mock mode / no containerd).
		// Return success as a no-op.
		logging.Debug("exec probe skipped (no exec function)",
			"container_id", containerID,
		)
		return true, nil
	}

	entry.mu.RLock()
	cmd := entry.config.ExecCommand
	entry.mu.RUnlock()

	if len(cmd) == 0 {
		return false, fmt.Errorf("exec probe has no command configured")
	}

	exitCode, err := execFunc(ctx, containerID, cmd)
	if err != nil {
		return false, fmt.Errorf("exec probe failed: %w", err)
	}

	if exitCode != 0 {
		return false, fmt.Errorf("exec probe returned exit code %d", exitCode)
	}

	return true, nil
}

// resolveAddress resolves a container ID to its host address.
func (hc *HealthChecker) resolveAddress(containerID string) (string, error) {
	hc.mu.RLock()
	addressFunc := hc.addressFunc
	hc.mu.RUnlock()

	if addressFunc == nil {
		// Default to localhost when no address resolver is configured
		return "127.0.0.1", nil
	}

	return addressFunc(containerID)
}

// updateProbeStatus updates the probe status after a health check.
func (hc *HealthChecker) updateProbeStatus(entry *probeEntry, containerID string, success bool, probeErr error) {
	entry.mu.Lock()
	defer entry.mu.Unlock()

	now := hc.now()
	entry.status.LastCheckTime = now
	entry.status.TotalChecks++

	if success {
		entry.status.TotalSuccesses++
		entry.status.ConsecutivePass++
		entry.status.ConsecutiveFail = 0
		entry.status.LastError = ""

		// Transition to healthy if we've hit the success threshold
		if entry.status.ConsecutivePass >= entry.config.SuccessThreshold &&
			entry.status.State != ProbeStateHealthy {
			entry.status.State = ProbeStateHealthy
			entry.status.LastTransitionTime = now
			logging.Info("container health probe: healthy",
				"container_id", containerID,
				"consecutive_pass", entry.status.ConsecutivePass,
			)
		}
	} else {
		entry.status.TotalFailures++
		entry.status.ConsecutiveFail++
		entry.status.ConsecutivePass = 0
		if probeErr != nil {
			entry.status.LastError = probeErr.Error()
		}

		// Transition to unhealthy if we've hit the failure threshold
		if entry.status.ConsecutiveFail >= entry.config.FailureThreshold &&
			entry.status.State != ProbeStateUnhealthy {
			entry.status.State = ProbeStateUnhealthy
			entry.status.LastTransitionTime = now
			logging.Warn("container health probe: unhealthy",
				"container_id", containerID,
				"consecutive_fail", entry.status.ConsecutiveFail,
				"last_error", entry.status.LastError,
			)
		}
	}
}

// applyDefaults fills in zero-value fields with sensible defaults.
func applyDefaults(config HealthProbeConfig) HealthProbeConfig {
	if config.InitialDelay == 0 {
		config.InitialDelay = 0 // No delay by default
	}
	if config.Interval == 0 {
		config.Interval = 10 * time.Second
	}
	if config.Timeout == 0 {
		config.Timeout = 5 * time.Second
	}
	if config.SuccessThreshold <= 0 {
		config.SuccessThreshold = 1
	}
	if config.FailureThreshold <= 0 {
		config.FailureThreshold = 3
	}
	if config.Type == ProbeHTTP {
		if config.HTTPPath == "" {
			config.HTTPPath = "/"
		}
		if config.HTTPPort == 0 {
			config.HTTPPort = 80
		}
		if len(config.HTTPExpectedCodes) == 0 {
			config.HTTPExpectedCodes = []int{200}
		}
	}
	return config
}
