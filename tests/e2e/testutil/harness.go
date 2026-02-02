package testutil

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/moltbunker/moltbunker/tests/mocks"
)

// TestHarness provides infrastructure for E2E testing
type TestHarness struct {
	t          *testing.T
	TempDir    string
	SocketPath string
	DataDir    string
	LogsDir    string
	ConfigPath string

	// Mock services
	Containerd *mocks.MockContainerdClient
	Tor        *mocks.MockTorService
	IPFS       *mocks.MockIPFSClient

	// Cleanup functions
	cleanupFns []func()
	mu         sync.Mutex

	// Context management
	ctx    context.Context
	cancel context.CancelFunc
}

// NewTestHarness creates a new test harness
func NewTestHarness(t *testing.T) *TestHarness {
	t.Helper()

	// Create temp directory for test
	tempDir, err := os.MkdirTemp("", "moltbunker-e2e-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	// Create subdirectories
	dataDir := filepath.Join(tempDir, "data")
	logsDir := filepath.Join(tempDir, "logs")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to create data directory: %v", err)
	}
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to create logs directory: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	h := &TestHarness{
		t:          t,
		TempDir:    tempDir,
		SocketPath: filepath.Join(tempDir, "test.sock"),
		DataDir:    dataDir,
		LogsDir:    logsDir,
		ConfigPath: filepath.Join(tempDir, "config.yaml"),
		Containerd: mocks.NewMockContainerdClient(),
		Tor:        mocks.NewMockTorService(),
		IPFS:       mocks.NewMockIPFSClient(),
		cleanupFns: make([]func(), 0),
		ctx:        ctx,
		cancel:     cancel,
	}

	// Register cleanup
	t.Cleanup(h.Cleanup)

	return h
}

// Context returns the harness context
func (h *TestHarness) Context() context.Context {
	return h.ctx
}

// WithTimeout returns a context with timeout
func (h *TestHarness) WithTimeout(d time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(h.ctx, d)
}

// AddCleanup adds a cleanup function to be called on harness cleanup
func (h *TestHarness) AddCleanup(fn func()) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.cleanupFns = append(h.cleanupFns, fn)
}

// Cleanup cleans up all test resources
func (h *TestHarness) Cleanup() {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Cancel context first
	if h.cancel != nil {
		h.cancel()
	}

	// Run cleanup functions in reverse order
	for i := len(h.cleanupFns) - 1; i >= 0; i-- {
		h.cleanupFns[i]()
	}

	// Remove temp directory
	if h.TempDir != "" {
		os.RemoveAll(h.TempDir)
	}
}

// CreateTestConfig creates a minimal test configuration file
func (h *TestHarness) CreateTestConfig() error {
	config := fmt.Sprintf(`daemon:
  port: 0
  data_dir: %s
  socket_path: %s
  log_level: debug

p2p:
  network_mode: clearnet
  max_peers: 10
  enable_mdns: false

tor:
  enabled: false
  data_dir: %s/tor

runtime:
  containerd_socket: /run/containerd/containerd.sock
  namespace: moltbunker-test
  enable_encryption: false

redundancy:
  replica_count: 1
  health_check_interval_seconds: 10
  health_timeout_seconds: 5

payment:
  enabled: false
`, h.DataDir, h.SocketPath, h.TempDir)

	return os.WriteFile(h.ConfigPath, []byte(config), 0644)
}

// WaitFor waits for a condition to be true, with timeout
func (h *TestHarness) WaitFor(condition func() bool, timeout time.Duration, message string) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return nil
		}
		time.Sleep(50 * time.Millisecond)
	}
	return fmt.Errorf("timeout waiting for: %s", message)
}

// WaitForContainer waits for a container to reach a specific status
func (h *TestHarness) WaitForContainer(id string, expectedStatus string, timeout time.Duration) error {
	return h.WaitFor(func() bool {
		container, exists := h.Containerd.GetContainer(id)
		if !exists {
			return false
		}
		return string(container.Status) == expectedStatus
	}, timeout, fmt.Sprintf("container %s to reach status %s", id, expectedStatus))
}

// AssertContainerExists checks that a container exists
func (h *TestHarness) AssertContainerExists(id string) {
	h.t.Helper()
	if _, exists := h.Containerd.GetContainer(id); !exists {
		h.t.Errorf("Expected container %s to exist, but it doesn't", id)
	}
}

// AssertContainerNotExists checks that a container does not exist
func (h *TestHarness) AssertContainerNotExists(id string) {
	h.t.Helper()
	if _, exists := h.Containerd.GetContainer(id); exists {
		h.t.Errorf("Expected container %s not to exist, but it does", id)
	}
}

// AssertTorRunning checks that Tor is running
func (h *TestHarness) AssertTorRunning() {
	h.t.Helper()
	if !h.Tor.IsRunning() {
		h.t.Error("Expected Tor to be running, but it isn't")
	}
}

// AssertTorNotRunning checks that Tor is not running
func (h *TestHarness) AssertTorNotRunning() {
	h.t.Helper()
	if h.Tor.IsRunning() {
		h.t.Error("Expected Tor not to be running, but it is")
	}
}

// MockPaymentEnvironment sets up the mock payment environment
func (h *TestHarness) MockPaymentEnvironment() {
	os.Setenv("MOLTBUNKER_MOCK_PAYMENTS", "true")
	h.AddCleanup(func() {
		os.Unsetenv("MOLTBUNKER_MOCK_PAYMENTS")
	})
}

// SetupMockContainer creates a mock container for testing
func (h *TestHarness) SetupMockContainer(id, image string) (*mocks.MockContainer, error) {
	ctx := h.Context()

	container, err := h.Containerd.CreateContainer(ctx, id, image, defaultTestResources())
	if err != nil {
		return nil, err
	}

	if err := h.Containerd.StartContainer(ctx, id); err != nil {
		return nil, err
	}

	return container, nil
}

// TempFile creates a temporary file in the test directory
func (h *TestHarness) TempFile(name string, content []byte) string {
	h.t.Helper()
	path := filepath.Join(h.TempDir, name)
	if err := os.WriteFile(path, content, 0644); err != nil {
		h.t.Fatalf("Failed to create temp file: %v", err)
	}
	return path
}

// TempSubDir creates a temporary subdirectory
func (h *TestHarness) TempSubDir(name string) string {
	h.t.Helper()
	path := filepath.Join(h.TempDir, name)
	if err := os.MkdirAll(path, 0755); err != nil {
		h.t.Fatalf("Failed to create temp subdirectory: %v", err)
	}
	return path
}

// defaultTestResources returns default resource limits for test containers
func defaultTestResources() mocks.MockResourceLimits {
	return mocks.MockResourceLimits{
		CPUQuota:    100000,
		CPUPeriod:   100000,
		MemoryLimit: 256 * 1024 * 1024, // 256MB
		DiskLimit:   1024 * 1024 * 1024, // 1GB
		PIDLimit:    100,
	}
}
