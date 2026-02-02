package mocks

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"sync"
	"time"
)

// MockIPFSClient simulates an IPFS client for testing
type MockIPFSClient struct {
	objects map[string][]byte // CID -> content
	pins    map[string]bool   // CID -> pinned
	mu      sync.RWMutex

	// Track method calls for verification
	calls   []MethodCall
	callsMu sync.Mutex

	// Error injection
	addErr  error
	getErr  error
	pinErr  error
	catErr  error

	// Behavior customization
	AddDelay time.Duration
	GetDelay time.Duration
}

// NewMockIPFSClient creates a new mock IPFS client
func NewMockIPFSClient() *MockIPFSClient {
	return &MockIPFSClient{
		objects: make(map[string][]byte),
		pins:    make(map[string]bool),
		calls:   make([]MethodCall, 0),
	}
}

// recordCall records a method call for later verification
func (m *MockIPFSClient) recordCall(method string, args ...interface{}) {
	m.callsMu.Lock()
	defer m.callsMu.Unlock()
	m.calls = append(m.calls, MethodCall{
		Method:    method,
		Args:      args,
		Timestamp: time.Now(),
	})
}

// GetCalls returns all recorded method calls
func (m *MockIPFSClient) GetCalls() []MethodCall {
	m.callsMu.Lock()
	defer m.callsMu.Unlock()
	result := make([]MethodCall, len(m.calls))
	copy(result, m.calls)
	return result
}

// ClearCalls clears all recorded method calls
func (m *MockIPFSClient) ClearCalls() {
	m.callsMu.Lock()
	defer m.callsMu.Unlock()
	m.calls = make([]MethodCall, 0)
}

// SetAddError sets an error to be returned on Add
func (m *MockIPFSClient) SetAddError(err error) {
	m.addErr = err
}

// SetGetError sets an error to be returned on Get
func (m *MockIPFSClient) SetGetError(err error) {
	m.getErr = err
}

// SetPinError sets an error to be returned on Pin
func (m *MockIPFSClient) SetPinError(err error) {
	m.pinErr = err
}

// SetCatError sets an error to be returned on Cat
func (m *MockIPFSClient) SetCatError(err error) {
	m.catErr = err
}

// Add adds content to the mock IPFS and returns a CID
func (m *MockIPFSClient) Add(ctx context.Context, r io.Reader) (string, error) {
	m.recordCall("Add")

	if m.addErr != nil {
		return "", m.addErr
	}

	if m.AddDelay > 0 {
		select {
		case <-time.After(m.AddDelay):
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}

	// Read all content
	content, err := io.ReadAll(r)
	if err != nil {
		return "", fmt.Errorf("failed to read content: %w", err)
	}

	// Generate a mock CID based on content hash
	cid := generateMockCID(content)

	m.mu.Lock()
	m.objects[cid] = content
	m.mu.Unlock()

	return cid, nil
}

// AddBytes is a convenience method for adding byte content
func (m *MockIPFSClient) AddBytes(ctx context.Context, content []byte) (string, error) {
	return m.Add(ctx, bytes.NewReader(content))
}

// Get retrieves content from the mock IPFS
func (m *MockIPFSClient) Get(ctx context.Context, cid string) (io.ReadCloser, error) {
	m.recordCall("Get", cid)

	if m.getErr != nil {
		return nil, m.getErr
	}

	if m.GetDelay > 0 {
		select {
		case <-time.After(m.GetDelay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	m.mu.RLock()
	content, exists := m.objects[cid]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("object not found: %s", cid)
	}

	return io.NopCloser(bytes.NewReader(content)), nil
}

// Cat retrieves content as bytes
func (m *MockIPFSClient) Cat(ctx context.Context, cid string) ([]byte, error) {
	m.recordCall("Cat", cid)

	if m.catErr != nil {
		return nil, m.catErr
	}

	reader, err := m.Get(ctx, cid)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	return io.ReadAll(reader)
}

// Pin pins content in the mock IPFS
func (m *MockIPFSClient) Pin(ctx context.Context, cid string) error {
	m.recordCall("Pin", cid)

	if m.pinErr != nil {
		return m.pinErr
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.objects[cid]; !exists {
		return fmt.Errorf("object not found: %s", cid)
	}

	m.pins[cid] = true
	return nil
}

// Unpin unpins content from the mock IPFS
func (m *MockIPFSClient) Unpin(ctx context.Context, cid string) error {
	m.recordCall("Unpin", cid)

	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.pins, cid)
	return nil
}

// IsPinned checks if content is pinned
func (m *MockIPFSClient) IsPinned(cid string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.pins[cid]
}

// ListPins returns all pinned CIDs
func (m *MockIPFSClient) ListPins() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]string, 0, len(m.pins))
	for cid := range m.pins {
		result = append(result, cid)
	}
	return result
}

// Remove removes content from the mock IPFS (garbage collection simulation)
func (m *MockIPFSClient) Remove(ctx context.Context, cid string) error {
	m.recordCall("Remove", cid)

	m.mu.Lock()
	defer m.mu.Unlock()

	// Don't remove pinned content
	if m.pins[cid] {
		return fmt.Errorf("cannot remove pinned object: %s", cid)
	}

	delete(m.objects, cid)
	return nil
}

// Exists checks if content exists in the mock IPFS
func (m *MockIPFSClient) Exists(cid string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, exists := m.objects[cid]
	return exists
}

// GetSize returns the size of content
func (m *MockIPFSClient) GetSize(cid string) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	content, exists := m.objects[cid]
	if !exists {
		return 0, fmt.Errorf("object not found: %s", cid)
	}

	return int64(len(content)), nil
}

// Stats returns mock IPFS stats
func (m *MockIPFSClient) Stats() MockIPFSStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var totalSize int64
	for _, content := range m.objects {
		totalSize += int64(len(content))
	}

	return MockIPFSStats{
		NumObjects:   len(m.objects),
		NumPinned:    len(m.pins),
		TotalSize:    totalSize,
		RepoSize:     totalSize,
	}
}

// MockIPFSStats contains mock IPFS statistics
type MockIPFSStats struct {
	NumObjects int
	NumPinned  int
	TotalSize  int64
	RepoSize   int64
}

// PreloadContent adds content directly to the mock (useful for test setup)
func (m *MockIPFSClient) PreloadContent(cid string, content []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.objects[cid] = content
}

// Clear removes all content from the mock IPFS
func (m *MockIPFSClient) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.objects = make(map[string][]byte)
	m.pins = make(map[string]bool)
}

// generateMockCID generates a mock CID from content
func generateMockCID(content []byte) string {
	hash := sha256.Sum256(content)
	// Simulate a CIDv1 format: Qm prefix + base58 encoded hash
	return "Qm" + hex.EncodeToString(hash[:])[:44]
}
