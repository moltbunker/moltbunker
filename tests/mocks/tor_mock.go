package mocks

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// MockTorService simulates the Tor service for testing
type MockTorService struct {
	running       bool
	onionAddr     string
	onionServices map[int]string
	mu            sync.RWMutex

	// Track method calls for verification
	calls   []MethodCall
	callsMu sync.Mutex

	// Error injection
	startErr         error
	stopErr          error
	createServiceErr error
	rotateErr        error

	// Behavior customization
	StartDelay time.Duration
}

// NewMockTorService creates a new mock Tor service
func NewMockTorService() *MockTorService {
	return &MockTorService{
		onionServices: make(map[int]string),
		calls:         make([]MethodCall, 0),
	}
}

// recordCall records a method call for later verification
func (m *MockTorService) recordCall(method string, args ...interface{}) {
	m.callsMu.Lock()
	defer m.callsMu.Unlock()
	m.calls = append(m.calls, MethodCall{
		Method:    method,
		Args:      args,
		Timestamp: time.Now(),
	})
}

// GetCalls returns all recorded method calls
func (m *MockTorService) GetCalls() []MethodCall {
	m.callsMu.Lock()
	defer m.callsMu.Unlock()
	result := make([]MethodCall, len(m.calls))
	copy(result, m.calls)
	return result
}

// ClearCalls clears all recorded method calls
func (m *MockTorService) ClearCalls() {
	m.callsMu.Lock()
	defer m.callsMu.Unlock()
	m.calls = make([]MethodCall, 0)
}

// SetStartError sets an error to be returned on Start
func (m *MockTorService) SetStartError(err error) {
	m.startErr = err
}

// SetStopError sets an error to be returned on Stop
func (m *MockTorService) SetStopError(err error) {
	m.stopErr = err
}

// SetCreateServiceError sets an error to be returned on CreateOnionService
func (m *MockTorService) SetCreateServiceError(err error) {
	m.createServiceErr = err
}

// SetRotateError sets an error to be returned on RotateCircuit
func (m *MockTorService) SetRotateError(err error) {
	m.rotateErr = err
}

// Start starts the mock Tor service
func (m *MockTorService) Start(ctx context.Context) error {
	m.recordCall("Start")

	if m.startErr != nil {
		return m.startErr
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running {
		return nil
	}

	if m.StartDelay > 0 {
		select {
		case <-time.After(m.StartDelay):
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	m.running = true
	return nil
}

// Stop stops the mock Tor service
func (m *MockTorService) Stop() error {
	m.recordCall("Stop")

	if m.stopErr != nil {
		return m.stopErr
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return nil
	}

	m.running = false
	m.onionAddr = ""
	m.onionServices = make(map[int]string)

	return nil
}

// IsRunning returns whether the mock Tor is running
func (m *MockTorService) IsRunning() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.running
}

// GetOnionAddress returns the main .onion address
func (m *MockTorService) GetOnionAddress() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.onionAddr
}

// OnionAddress is an alias for GetOnionAddress
func (m *MockTorService) OnionAddress() string {
	return m.GetOnionAddress()
}

// CreateOnionService creates a mock onion service for a specific port
func (m *MockTorService) CreateOnionService(ctx context.Context, port int) (string, error) {
	m.recordCall("CreateOnionService", port)

	if m.createServiceErr != nil {
		return "", m.createServiceErr
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return "", fmt.Errorf("Tor not running")
	}

	// Check if we already have a service for this port
	if addr, exists := m.onionServices[port]; exists {
		return addr, nil
	}

	// Generate a random mock onion address
	addr := generateMockOnionAddress()
	m.onionServices[port] = addr

	// Set as main address if this is the first one
	if m.onionAddr == "" {
		m.onionAddr = addr
	}

	return addr, nil
}

// RotateCircuit simulates rotating the Tor circuit
func (m *MockTorService) RotateCircuit(ctx context.Context) error {
	m.recordCall("RotateCircuit")

	if m.rotateErr != nil {
		return m.rotateErr
	}

	m.mu.RLock()
	running := m.running
	m.mu.RUnlock()

	if !running {
		return fmt.Errorf("Tor not running")
	}

	return nil
}

// GetCircuitInfo returns mock circuit information
func (m *MockTorService) GetCircuitInfo(ctx context.Context) ([]MockCircuitInfo, error) {
	m.recordCall("GetCircuitInfo")

	m.mu.RLock()
	running := m.running
	m.mu.RUnlock()

	if !running {
		return nil, fmt.Errorf("Tor not running")
	}

	// Return mock circuit info
	return []MockCircuitInfo{
		{
			ID:     "1",
			Status: "BUILT",
			Path:   []string{"guard", "middle", "exit"},
		},
		{
			ID:     "2",
			Status: "BUILT",
			Path:   []string{"guard2", "middle2", "exit2"},
		},
	}, nil
}

// MockCircuitInfo contains mock information about a Tor circuit
type MockCircuitInfo struct {
	ID     string
	Status string
	Path   []string
}

// CloseOnionService closes a specific onion service
func (m *MockTorService) CloseOnionService(port int) error {
	m.recordCall("CloseOnionService", port)

	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.onionServices, port)
	return nil
}

// ListOnionServices returns all active onion service addresses
func (m *MockTorService) ListOnionServices() map[int]string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[int]string)
	for port, addr := range m.onionServices {
		result[port] = addr
	}
	return result
}

// generateMockOnionAddress generates a realistic-looking mock .onion address
func generateMockOnionAddress() string {
	// v3 onion addresses are 56 characters + .onion
	bytes := make([]byte, 28)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to a fixed address if random fails
		return "mockonionaddressxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx.onion"
	}
	return hex.EncodeToString(bytes) + ".onion"
}

// SetRunning directly sets the running state (useful for tests)
func (m *MockTorService) SetRunning(running bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.running = running
}

// SetOnionAddress directly sets the onion address (useful for tests)
func (m *MockTorService) SetOnionAddress(addr string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onionAddr = addr
}
