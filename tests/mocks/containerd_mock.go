package mocks

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/moltbunker/moltbunker/pkg/types"
)

// MockResourceLimits defines resource limits for mock containers
type MockResourceLimits = types.ResourceLimits

// MockContainer represents a mock container
type MockContainer struct {
	ID              string
	Image           string
	Status          types.ContainerStatus
	CreatedAt       time.Time
	StartedAt       time.Time
	Resources       types.ResourceLimits
	EncryptedVolume string
	OnionAddress    string
	Logs            []string
	mu              sync.RWMutex
}

// MockContainerdClient is a mock implementation of the containerd client for testing
type MockContainerdClient struct {
	containers map[string]*MockContainer
	mu         sync.RWMutex
	namespace  string

	// Track method calls for verification
	calls     []MethodCall
	callsMu   sync.Mutex

	// Error injection
	createErr error
	startErr  error
	stopErr   error
	deleteErr error
	pullErr   error

	// Behavior customization
	StartDelay time.Duration
	StopDelay  time.Duration
}

// MethodCall records a method invocation
type MethodCall struct {
	Method    string
	Args      []interface{}
	Timestamp time.Time
}

// NewMockContainerdClient creates a new mock containerd client
func NewMockContainerdClient() *MockContainerdClient {
	return &MockContainerdClient{
		containers: make(map[string]*MockContainer),
		namespace:  "moltbunker-test",
		calls:      make([]MethodCall, 0),
	}
}

// recordCall records a method call for later verification
func (m *MockContainerdClient) recordCall(method string, args ...interface{}) {
	m.callsMu.Lock()
	defer m.callsMu.Unlock()
	m.calls = append(m.calls, MethodCall{
		Method:    method,
		Args:      args,
		Timestamp: time.Now(),
	})
}

// GetCalls returns all recorded method calls
func (m *MockContainerdClient) GetCalls() []MethodCall {
	m.callsMu.Lock()
	defer m.callsMu.Unlock()
	result := make([]MethodCall, len(m.calls))
	copy(result, m.calls)
	return result
}

// ClearCalls clears all recorded method calls
func (m *MockContainerdClient) ClearCalls() {
	m.callsMu.Lock()
	defer m.callsMu.Unlock()
	m.calls = make([]MethodCall, 0)
}

// SetCreateError sets an error to be returned on CreateContainer
func (m *MockContainerdClient) SetCreateError(err error) {
	m.createErr = err
}

// SetStartError sets an error to be returned on StartContainer
func (m *MockContainerdClient) SetStartError(err error) {
	m.startErr = err
}

// SetStopError sets an error to be returned on StopContainer
func (m *MockContainerdClient) SetStopError(err error) {
	m.stopErr = err
}

// SetDeleteError sets an error to be returned on DeleteContainer
func (m *MockContainerdClient) SetDeleteError(err error) {
	m.deleteErr = err
}

// SetPullError sets an error to be returned on PullImage
func (m *MockContainerdClient) SetPullError(err error) {
	m.pullErr = err
}

// PullImage simulates pulling an image
func (m *MockContainerdClient) PullImage(ctx context.Context, ref string) error {
	m.recordCall("PullImage", ref)
	if m.pullErr != nil {
		return m.pullErr
	}
	// Simulate some pull time
	time.Sleep(10 * time.Millisecond)
	return nil
}

// CreateContainer creates a mock container
func (m *MockContainerdClient) CreateContainer(ctx context.Context, id string, imageRef string, resources types.ResourceLimits) (*MockContainer, error) {
	m.recordCall("CreateContainer", id, imageRef, resources)

	if m.createErr != nil {
		return nil, m.createErr
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.containers[id]; exists {
		return nil, fmt.Errorf("container already exists: %s", id)
	}

	container := &MockContainer{
		ID:        id,
		Image:     imageRef,
		Status:    types.ContainerStatusCreated,
		CreatedAt: time.Now(),
		Resources: resources,
		Logs:      make([]string, 0),
	}

	m.containers[id] = container
	return container, nil
}

// StartContainer starts a mock container
func (m *MockContainerdClient) StartContainer(ctx context.Context, id string) error {
	m.recordCall("StartContainer", id)

	if m.startErr != nil {
		return m.startErr
	}

	m.mu.Lock()
	container, exists := m.containers[id]
	m.mu.Unlock()

	if !exists {
		return fmt.Errorf("container not found: %s", id)
	}

	if m.StartDelay > 0 {
		time.Sleep(m.StartDelay)
	}

	container.mu.Lock()
	container.Status = types.ContainerStatusRunning
	container.StartedAt = time.Now()
	container.mu.Unlock()

	return nil
}

// StopContainer stops a mock container
func (m *MockContainerdClient) StopContainer(ctx context.Context, id string, timeout time.Duration) error {
	m.recordCall("StopContainer", id, timeout)

	if m.stopErr != nil {
		return m.stopErr
	}

	m.mu.RLock()
	container, exists := m.containers[id]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("container not found: %s", id)
	}

	if m.StopDelay > 0 {
		time.Sleep(m.StopDelay)
	}

	container.mu.Lock()
	container.Status = types.ContainerStatusStopped
	container.mu.Unlock()

	return nil
}

// DeleteContainer deletes a mock container
func (m *MockContainerdClient) DeleteContainer(ctx context.Context, id string) error {
	m.recordCall("DeleteContainer", id)

	if m.deleteErr != nil {
		return m.deleteErr
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.containers[id]; !exists {
		return fmt.Errorf("container not found: %s", id)
	}

	delete(m.containers, id)
	return nil
}

// GetContainer returns a mock container
func (m *MockContainerdClient) GetContainer(id string) (*MockContainer, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	container, exists := m.containers[id]
	return container, exists
}

// ListContainers returns all mock containers
func (m *MockContainerdClient) ListContainers() []*MockContainer {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*MockContainer, 0, len(m.containers))
	for _, c := range m.containers {
		result = append(result, c)
	}
	return result
}

// GetContainerStatus returns the status of a mock container
func (m *MockContainerdClient) GetContainerStatus(ctx context.Context, id string) (types.ContainerStatus, error) {
	m.recordCall("GetContainerStatus", id)

	m.mu.RLock()
	container, exists := m.containers[id]
	m.mu.RUnlock()

	if !exists {
		return types.ContainerStatusFailed, fmt.Errorf("container not found: %s", id)
	}

	container.mu.RLock()
	defer container.mu.RUnlock()
	return container.Status, nil
}

// GetContainerLogs returns logs from a mock container
func (m *MockContainerdClient) GetContainerLogs(ctx context.Context, id string, follow bool, tail int) (io.ReadCloser, error) {
	m.recordCall("GetContainerLogs", id, follow, tail)

	m.mu.RLock()
	container, exists := m.containers[id]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("container not found: %s", id)
	}

	container.mu.RLock()
	logs := make([]string, len(container.Logs))
	copy(logs, container.Logs)
	container.mu.RUnlock()

	// Apply tail
	if tail > 0 && tail < len(logs) {
		logs = logs[len(logs)-tail:]
	}

	var buf bytes.Buffer
	for _, line := range logs {
		buf.WriteString(line)
		buf.WriteString("\n")
	}

	return io.NopCloser(&buf), nil
}

// AppendLog adds a log line to a container
func (m *MockContainerdClient) AppendLog(id string, line string) error {
	m.mu.RLock()
	container, exists := m.containers[id]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("container not found: %s", id)
	}

	container.mu.Lock()
	container.Logs = append(container.Logs, line)
	container.mu.Unlock()

	return nil
}

// ExecInContainer simulates executing a command in a container
func (m *MockContainerdClient) ExecInContainer(ctx context.Context, id string, cmd []string) ([]byte, error) {
	m.recordCall("ExecInContainer", id, cmd)

	m.mu.RLock()
	container, exists := m.containers[id]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("container not found: %s", id)
	}

	container.mu.RLock()
	status := container.Status
	container.mu.RUnlock()

	if status != types.ContainerStatusRunning {
		return nil, fmt.Errorf("container not running: %s", id)
	}

	// Return mock output
	return []byte("mock exec output"), nil
}

// GetHealthStatus returns mock health status
func (m *MockContainerdClient) GetHealthStatus(ctx context.Context, id string) (*types.HealthStatus, error) {
	m.recordCall("GetHealthStatus", id)

	m.mu.RLock()
	container, exists := m.containers[id]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("container not found: %s", id)
	}

	container.mu.RLock()
	status := container.Status
	container.mu.RUnlock()

	return &types.HealthStatus{
		CPUUsage:    10.5,
		MemoryUsage: 1024 * 1024 * 100, // 100MB
		Healthy:     status == types.ContainerStatusRunning,
		LastUpdate:  time.Now(),
	}, nil
}

// Close is a no-op for the mock client
func (m *MockContainerdClient) Close() error {
	m.recordCall("Close")
	return nil
}

// SetEncryptedVolume sets the encrypted volume for a container
func (m *MockContainerdClient) SetEncryptedVolume(id string, volume string) error {
	m.mu.RLock()
	container, exists := m.containers[id]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("container not found: %s", id)
	}

	container.mu.Lock()
	container.EncryptedVolume = volume
	container.mu.Unlock()

	return nil
}

// SetOnionAddress sets the onion address for a container
func (m *MockContainerdClient) SetOnionAddress(id string, addr string) error {
	m.mu.RLock()
	container, exists := m.containers[id]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("container not found: %s", id)
	}

	container.mu.Lock()
	container.OnionAddress = addr
	container.mu.Unlock()

	return nil
}
