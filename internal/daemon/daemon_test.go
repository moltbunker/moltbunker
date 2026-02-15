package daemon

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/moltbunker/moltbunker/pkg/types"
)

// =============================================================================
// Deployment type tests
// =============================================================================

func TestDeployment_Struct(t *testing.T) {
	now := time.Now()
	d := &Deployment{
		ID:           "dep-1234567890",
		Image:        "nginx:latest",
		Status:       types.ContainerStatusRunning,
		CreatedAt:    now,
		StartedAt:    now.Add(time.Second),
		Encrypted:    true,
		OnionService: true,
		OnionAddress: "abc123.onion",
		OnionPort:    8080,
		TorOnly:      true,
		LocalReplica: 0,
		Regions:      []string{"Americas", "Europe", "Asia-Pacific"},
		Resources: types.ResourceLimits{
			CPUQuota:    100000,
			CPUPeriod:   100000,
			MemoryLimit: 1024 * 1024 * 512,
			PIDLimit:    100,
		},
	}

	if d.ID != "dep-1234567890" {
		t.Errorf("ID mismatch: got %s", d.ID)
	}
	if d.Image != "nginx:latest" {
		t.Errorf("Image mismatch: got %s", d.Image)
	}
	if d.Status != types.ContainerStatusRunning {
		t.Errorf("Status mismatch: got %s", d.Status)
	}
	if !d.Encrypted {
		t.Error("Encrypted should be true")
	}
	if !d.OnionService {
		t.Error("OnionService should be true")
	}
	if d.OnionPort != 8080 {
		t.Errorf("OnionPort mismatch: got %d", d.OnionPort)
	}
	if !d.TorOnly {
		t.Error("TorOnly should be true")
	}
	if len(d.Regions) != 3 {
		t.Errorf("Expected 3 regions, got %d", len(d.Regions))
	}
}

func TestDeployment_JSONRoundTrip(t *testing.T) {
	now := time.Now().Truncate(time.Millisecond) // JSON loses sub-ms precision
	d := &Deployment{
		ID:        "dep-json-test",
		Image:     "redis:7",
		Status:    types.ContainerStatusPending,
		CreatedAt: now,
		Encrypted: true,
		Regions:   []string{"Americas"},
		Resources: types.ResourceLimits{
			CPUQuota:    50000,
			MemoryLimit: 256 * 1024 * 1024,
		},
	}

	data, err := json.Marshal(d)
	if err != nil {
		t.Fatalf("Failed to marshal deployment: %v", err)
	}

	var decoded Deployment
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal deployment: %v", err)
	}

	if decoded.ID != d.ID {
		t.Errorf("ID mismatch after roundtrip: got %s, want %s", decoded.ID, d.ID)
	}
	if decoded.Image != d.Image {
		t.Errorf("Image mismatch after roundtrip: got %s, want %s", decoded.Image, d.Image)
	}
	if decoded.Status != d.Status {
		t.Errorf("Status mismatch after roundtrip: got %s, want %s", decoded.Status, d.Status)
	}
	if decoded.Encrypted != d.Encrypted {
		t.Errorf("Encrypted mismatch after roundtrip: got %v, want %v", decoded.Encrypted, d.Encrypted)
	}
	if decoded.Resources.CPUQuota != d.Resources.CPUQuota {
		t.Errorf("CPUQuota mismatch after roundtrip: got %d, want %d", decoded.Resources.CPUQuota, d.Resources.CPUQuota)
	}
}

// =============================================================================
// pendingDeployment tests
// =============================================================================

func TestPendingDeployment_Close(t *testing.T) {
	pd := &pendingDeployment{
		containerID: "dep-close-test",
		ackChan:     make(chan replicaAck, 10),
		created:     time.Now(),
		acks:        make([]replicaAck, 0),
	}

	if pd.isClosed() {
		t.Error("pendingDeployment should not be closed initially")
	}

	pd.close()

	if !pd.isClosed() {
		t.Error("pendingDeployment should be closed after close()")
	}

	// Close again -- should not panic (sync.Once)
	pd.close()

	if !pd.isClosed() {
		t.Error("pendingDeployment should still be closed after double close")
	}
}

func TestPendingDeployment_CloseChannelDrains(t *testing.T) {
	pd := &pendingDeployment{
		containerID: "dep-drain-test",
		ackChan:     make(chan replicaAck, 5),
		created:     time.Now(),
		acks:        make([]replicaAck, 0),
	}

	// Send some acks before closing
	pd.ackChan <- replicaAck{NodeID: "node1", Success: true}
	pd.ackChan <- replicaAck{NodeID: "node2", Success: false, Error: "timeout"}

	pd.close()

	// Read remaining acks from closed channel
	count := 0
	for range pd.ackChan {
		count++
	}

	if count != 2 {
		t.Errorf("Expected 2 remaining acks, got %d", count)
	}
}

func TestPendingDeployment_ConcurrentClose(t *testing.T) {
	pd := &pendingDeployment{
		containerID: "dep-concurrent-close",
		ackChan:     make(chan replicaAck, 10),
		created:     time.Now(),
		acks:        make([]replicaAck, 0),
	}

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			pd.close()
		}()
	}

	wg.Wait()

	if !pd.isClosed() {
		t.Error("pendingDeployment should be closed after concurrent closes")
	}
}

// =============================================================================
// ContainerManagerConfig tests
// =============================================================================

func TestContainerManagerConfig_Defaults(t *testing.T) {
	config := ContainerManagerConfig{}

	if config.DataDir != "" {
		t.Error("DataDir should be empty by default")
	}
	if config.ContainerdSocket != "" {
		t.Error("ContainerdSocket should be empty by default")
	}
	if config.TorDataDir != "" {
		t.Error("TorDataDir should be empty by default")
	}
	if config.EnableEncryption {
		t.Error("EnableEncryption should be false by default")
	}
}

func TestContainerManagerConfig_WithValues(t *testing.T) {
	config := ContainerManagerConfig{
		DataDir:          "/var/lib/moltbunker",
		ContainerdSocket: "/run/containerd/containerd.sock",
		TorDataDir:       "/var/lib/moltbunker/tor",
		EnableEncryption: true,
	}

	if config.DataDir != "/var/lib/moltbunker" {
		t.Errorf("DataDir mismatch: got %s", config.DataDir)
	}
	if config.ContainerdSocket != "/run/containerd/containerd.sock" {
		t.Errorf("ContainerdSocket mismatch: got %s", config.ContainerdSocket)
	}
	if !config.EnableEncryption {
		t.Error("EnableEncryption should be true")
	}
}

// =============================================================================
// DeployResult tests
// =============================================================================

func TestDeployResult_Struct(t *testing.T) {
	result := &DeployResult{
		Deployment: &Deployment{
			ID:    "dep-result-test",
			Image: "alpine:latest",
		},
		ReplicaCount: 2,
	}

	if result.Deployment.ID != "dep-result-test" {
		t.Errorf("Deployment ID mismatch: got %s", result.Deployment.ID)
	}
	if result.ReplicaCount != 2 {
		t.Errorf("ReplicaCount mismatch: got %d", result.ReplicaCount)
	}
}

// =============================================================================
// Error types tests
// =============================================================================

func TestErrDeploymentNotFound(t *testing.T) {
	err := ErrDeploymentNotFound{ContainerID: "dep-404"}

	expected := "deployment not found: dep-404"
	if err.Error() != expected {
		t.Errorf("Error message mismatch: got %q, want %q", err.Error(), expected)
	}
}

func TestErrDeploymentNotFound_Interface(t *testing.T) {
	var err error = ErrDeploymentNotFound{ContainerID: "dep-iface"}

	if err == nil {
		t.Error("Error should not be nil")
	}

	// Should be able to check with type assertion
	var notFound ErrDeploymentNotFound
	if !errors.As(err, &notFound) {
		t.Error("Should be able to use errors.As with ErrDeploymentNotFound")
	}

	if notFound.ContainerID != "dep-iface" {
		t.Errorf("ContainerID mismatch: got %s", notFound.ContainerID)
	}
}

func TestErrContainerdNotAvailable(t *testing.T) {
	if ErrContainerdNotAvailable == nil {
		t.Fatal("ErrContainerdNotAvailable should not be nil")
	}
	if ErrContainerdNotAvailable.Error() != "containerd not available" {
		t.Errorf("Error message mismatch: got %q", ErrContainerdNotAvailable.Error())
	}
}

// =============================================================================
// State persistence tests
// =============================================================================

func TestPersistedState_SaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a ContainerManager with just the fields needed for state persistence
	cm := &ContainerManager{
		deployments: make(map[string]*Deployment),
		dataDir:     tmpDir,
	}

	// Add some deployments
	now := time.Now().Truncate(time.Millisecond) // Truncate for JSON precision
	cm.deployments["dep-1"] = &Deployment{
		ID:        "dep-1",
		Image:     "nginx:latest",
		Status:    types.ContainerStatusRunning,
		CreatedAt: now,
		StartedAt: now.Add(time.Second),
		Encrypted: true,
		Regions:   []string{"Americas", "Europe", "Asia-Pacific"},
	}
	cm.deployments["dep-2"] = &Deployment{
		ID:        "dep-2",
		Image:     "redis:7",
		Status:    types.ContainerStatusStopped,
		CreatedAt: now.Add(-time.Hour),
		Regions:   []string{"Americas"},
	}

	// Save state
	if err := cm.saveState(); err != nil {
		t.Fatalf("Failed to save state: %v", err)
	}

	// Verify file exists
	statePath := filepath.Join(tmpDir, "state.json")
	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		t.Fatal("State file should exist after save")
	}

	// Create new ContainerManager and load state
	cm2 := &ContainerManager{
		deployments: make(map[string]*Deployment),
		dataDir:     tmpDir,
	}

	if err := cm2.loadState(); err != nil {
		t.Fatalf("Failed to load state: %v", err)
	}

	// Verify deployments loaded correctly
	if len(cm2.deployments) != 2 {
		t.Fatalf("Expected 2 deployments, got %d", len(cm2.deployments))
	}

	d1, exists := cm2.deployments["dep-1"]
	if !exists {
		t.Fatal("dep-1 should exist in loaded state")
	}
	if d1.Image != "nginx:latest" {
		t.Errorf("dep-1 image mismatch: got %s", d1.Image)
	}
	if d1.Status != types.ContainerStatusRunning {
		t.Errorf("dep-1 status mismatch: got %s", d1.Status)
	}
	if !d1.Encrypted {
		t.Error("dep-1 should be encrypted")
	}
	if len(d1.Regions) != 3 {
		t.Errorf("dep-1 expected 3 regions, got %d", len(d1.Regions))
	}

	d2, exists := cm2.deployments["dep-2"]
	if !exists {
		t.Fatal("dep-2 should exist in loaded state")
	}
	if d2.Image != "redis:7" {
		t.Errorf("dep-2 image mismatch: got %s", d2.Image)
	}
	if d2.Status != types.ContainerStatusStopped {
		t.Errorf("dep-2 status mismatch: got %s", d2.Status)
	}
}

func TestPersistedState_LoadNonexistent(t *testing.T) {
	tmpDir := t.TempDir()
	cm := &ContainerManager{
		deployments: make(map[string]*Deployment),
		dataDir:     tmpDir,
	}

	// Loading from a nonexistent state file should succeed (start fresh)
	if err := cm.loadState(); err != nil {
		t.Fatalf("Loading nonexistent state should not error: %v", err)
	}

	if len(cm.deployments) != 0 {
		t.Errorf("Expected 0 deployments, got %d", len(cm.deployments))
	}
}

func TestPersistedState_LoadCorrupted(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "state.json")

	// Write corrupted JSON
	if err := os.WriteFile(statePath, []byte("not valid json{{{"), 0600); err != nil {
		t.Fatalf("Failed to write corrupted state: %v", err)
	}

	cm := &ContainerManager{
		deployments: make(map[string]*Deployment),
		dataDir:     tmpDir,
	}

	err := cm.loadState()
	if err == nil {
		t.Error("Loading corrupted state should return error")
	}
}

func TestPersistedState_EmptyDeployments(t *testing.T) {
	tmpDir := t.TempDir()
	cm := &ContainerManager{
		deployments: make(map[string]*Deployment),
		dataDir:     tmpDir,
	}

	// Save empty state
	if err := cm.saveState(); err != nil {
		t.Fatalf("Failed to save empty state: %v", err)
	}

	// Load it back
	cm2 := &ContainerManager{
		deployments: make(map[string]*Deployment),
		dataDir:     tmpDir,
	}
	if err := cm2.loadState(); err != nil {
		t.Fatalf("Failed to load empty state: %v", err)
	}
	if len(cm2.deployments) != 0 {
		t.Errorf("Expected 0 deployments, got %d", len(cm2.deployments))
	}
}

func TestStateFilePath(t *testing.T) {
	cm := &ContainerManager{
		dataDir: "/tmp/moltbunker-data",
	}

	expected := "/tmp/moltbunker-data/state.json"
	if cm.stateFilePath() != expected {
		t.Errorf("State file path mismatch: got %s, want %s", cm.stateFilePath(), expected)
	}
}

// =============================================================================
// HealthzResponse / ReadyzResponse / ContainerInfo struct tests
// =============================================================================

func TestHealthzResponse_Structure(t *testing.T) {
	h := HealthzResponse{
		Status:              "healthy",
		NodeRunning:         true,
		ContainerdConnected: false,
		PeerCount:           5,
		GoroutineCount:      42,
		MemoryUsageMB:       100.5,
		MemoryAllocMB:       80.2,
		Timestamp:           time.Now(),
	}

	if h.Status != "healthy" {
		t.Error("Status mismatch")
	}
	if !h.NodeRunning {
		t.Error("NodeRunning should be true")
	}
	if h.PeerCount != 5 {
		t.Errorf("PeerCount mismatch: got %d", h.PeerCount)
	}
}

func TestReadyzResponse_Structure(t *testing.T) {
	r := ReadyzResponse{
		Ready:     true,
		Message:   "all systems operational",
		Timestamp: time.Now(),
	}

	if !r.Ready {
		t.Error("Ready should be true")
	}
	if r.Message != "all systems operational" {
		t.Errorf("Message mismatch: got %s", r.Message)
	}
}

func TestContainerInfo_Structure(t *testing.T) {
	ci := ContainerInfo{
		ID:           "cnt-123",
		Image:        "nginx:latest",
		Status:       "running",
		CreatedAt:    time.Now(),
		Encrypted:    true,
		OnionAddress: "abc.onion",
		Regions:      []string{"Americas", "Europe"},
	}

	if ci.ID != "cnt-123" {
		t.Errorf("ID mismatch: got %s", ci.ID)
	}
	if ci.Status != "running" {
		t.Errorf("Status mismatch: got %s", ci.Status)
	}
	if !ci.Encrypted {
		t.Error("Encrypted should be true")
	}
	if len(ci.Regions) != 2 {
		t.Errorf("Expected 2 regions, got %d", len(ci.Regions))
	}
}

// =============================================================================
// replicaAck tests
// =============================================================================

func TestReplicaAck_Struct(t *testing.T) {
	ack := replicaAck{
		NodeID:  "node-123",
		Region:  "Americas",
		Success: true,
		Error:   "",
	}

	if ack.NodeID != "node-123" {
		t.Errorf("NodeID mismatch: got %s", ack.NodeID)
	}
	if ack.Region != "Americas" {
		t.Errorf("Region mismatch: got %s", ack.Region)
	}
	if !ack.Success {
		t.Error("Success should be true")
	}

	failAck := replicaAck{
		NodeID:  "node-456",
		Region:  "Europe",
		Success: false,
		Error:   "disk full",
	}

	if failAck.Success {
		t.Error("Success should be false for failed ack")
	}
	if failAck.Error != "disk full" {
		t.Errorf("Error mismatch: got %s", failAck.Error)
	}
}

// =============================================================================
// Deployment map operations tests (via exported methods on ContainerManager)
// =============================================================================

func TestContainerManager_DeploymentMapOperations(t *testing.T) {
	cm := &ContainerManager{
		deployments:        make(map[string]*Deployment),
		pendingDeployments: make(map[string]*pendingDeployment),
	}

	// Initially empty
	if list := cm.ListDeployments(); len(list) != 0 {
		t.Errorf("Expected 0 deployments initially, got %d", len(list))
	}

	// Add deployments manually (simulating internal state)
	cm.mu.Lock()
	cm.deployments["dep-1"] = &Deployment{
		ID:     "dep-1",
		Image:  "nginx:latest",
		Status: types.ContainerStatusRunning,
	}
	cm.deployments["dep-2"] = &Deployment{
		ID:     "dep-2",
		Image:  "redis:7",
		Status: types.ContainerStatusStopped,
	}
	cm.mu.Unlock()

	// List should return both
	list := cm.ListDeployments()
	if len(list) != 2 {
		t.Fatalf("Expected 2 deployments, got %d", len(list))
	}

	// Get specific deployment
	d, exists := cm.GetDeployment("dep-1")
	if !exists {
		t.Fatal("dep-1 should exist")
	}
	if d.Image != "nginx:latest" {
		t.Errorf("Image mismatch: got %s", d.Image)
	}

	// Get nonexistent deployment
	_, exists = cm.GetDeployment("dep-nonexistent")
	if exists {
		t.Error("Nonexistent deployment should not be found")
	}
}

// =============================================================================
// GetReplicaStatus tests
// =============================================================================

func TestContainerManager_GetReplicaStatus_NotFound(t *testing.T) {
	cm := &ContainerManager{
		deployments:        make(map[string]*Deployment),
		pendingDeployments: make(map[string]*pendingDeployment),
	}

	ackCount, successCount, exists := cm.GetReplicaStatus("nonexistent")
	if exists {
		t.Error("Should not find replica status for nonexistent deployment")
	}
	if ackCount != 0 || successCount != 0 {
		t.Error("Counts should be zero for nonexistent deployment")
	}
}

func TestContainerManager_GetReplicaStatus_WithPending(t *testing.T) {
	cm := &ContainerManager{
		deployments:        make(map[string]*Deployment),
		pendingDeployments: make(map[string]*pendingDeployment),
	}

	pd := &pendingDeployment{
		containerID:  "dep-pending",
		ackChan:      make(chan replicaAck, 10),
		ackCount:     3,
		successCount: 2,
		created:      time.Now(),
		acks:         make([]replicaAck, 0),
	}

	cm.pendingMu.Lock()
	cm.pendingDeployments["dep-pending"] = pd
	cm.pendingMu.Unlock()

	ackCount, successCount, exists := cm.GetReplicaStatus("dep-pending")
	if !exists {
		t.Fatal("Should find replica status for pending deployment")
	}
	if ackCount != 3 {
		t.Errorf("ackCount mismatch: got %d, want 3", ackCount)
	}
	if successCount != 2 {
		t.Errorf("successCount mismatch: got %d, want 2", successCount)
	}
}
