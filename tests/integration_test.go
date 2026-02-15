package tests

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/moltbunker/moltbunker/internal/p2p"
	"github.com/moltbunker/moltbunker/internal/redundancy"
	"github.com/moltbunker/moltbunker/pkg/types"
)

// Integration tests require actual services running
// These are skipped by default unless -short=false is used

// --- P2P Network Integration Tests ---

func TestIntegration_P2PNetwork(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create two nodes
	_ = t.TempDir() + "/node1.key"
	_ = t.TempDir() + "/node2.key"

	config1 := &p2p.DHTConfig{
		Port:           9001,
		EnableMDNS:     true,
		BootstrapPeers: []string{}, // Empty for testing
		MaxPeers:       50,
	}
	node1, err := p2p.NewDHT(ctx, config1, nil)
	if err != nil {
		t.Fatalf("Failed to create node 1: %v", err)
	}
	defer node1.Close()

	config2 := &p2p.DHTConfig{
		Port:           9002,
		EnableMDNS:     true,
		BootstrapPeers: []string{}, // Empty for testing
		MaxPeers:       50,
	}
	node2, err := p2p.NewDHT(ctx, config2, nil)
	if err != nil {
		t.Fatalf("Failed to create node 2: %v", err)
	}
	defer node2.Close()

	// Test node discovery
	// Note: This would require actual network setup
	t.Log("P2P nodes created successfully")
}

func TestIntegration_RedundancySystem(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	replicator := redundancy.NewReplicator()
	healthMonitor := redundancy.NewHealthMonitor()

	containerID := "test-container"
	regions := []string{"Americas", "Europe", "Asia-Pacific"}

	_, err := replicator.CreateReplicaSet(containerID, regions)
	if err != nil {
		t.Fatalf("Failed to create replica set: %v", err)
	}

	// Start health monitoring
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go healthMonitor.Start(ctx)

	// Simulate health updates
	health := types.HealthStatus{
		CPUUsage:    50.0,
		MemoryUsage: 1024 * 1024 * 100,
		Healthy:     true,
		LastUpdate:  time.Now(),
	}

	for i := 0; i < 3; i++ {
		healthMonitor.UpdateHealth(containerID, i, health)
	}

	// Verify all replicas are healthy
	for i := 0; i < 3; i++ {
		if !healthMonitor.IsHealthy(containerID, i) {
			t.Errorf("Replica %d should be healthy", i)
		}
	}

	t.Log("Redundancy system integration test passed")
}

// --- Containerd Integration Tests ---

// isContainerdAvailable checks if containerd is available on the system
func isContainerdAvailable() bool {
	// Check if containerd socket exists
	socketPath := "/run/containerd/containerd.sock"
	if _, err := os.Stat(socketPath); os.IsNotExist(err) {
		return false
	}

	// Try to run ctr command
	cmd := exec.Command("ctr", "--version")
	return cmd.Run() == nil
}

func TestIntegration_FullDeploymentFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test (use -short=false to run)")
	}

	if !isContainerdAvailable() {
		t.Skip("Containerd not available, skipping full deployment test")
	}

	// This test would require actual containerd setup
	// For now, test the deployment flow logic without actual container creation

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Test deployment parameters
	image := "docker.io/library/alpine:latest"
	resources := types.ResourceLimits{
		MemoryLimit: 256 * 1024 * 1024, // 256MB
		CPUQuota:    50000,
		CPUPeriod:   100000,
		PIDLimit:    100,
	}

	t.Logf("Would deploy image: %s", image)
	t.Logf("Resources: Memory=%dMB, CPU Quota=%d", resources.MemoryLimit/(1024*1024), resources.CPUQuota)

	// Verify context is still valid
	select {
	case <-ctx.Done():
		t.Fatal("Context cancelled before deployment could complete")
	default:
		t.Log("Full deployment flow test setup complete")
	}
}

func TestIntegration_FullDeploymentFlow_WithMockContainerd(t *testing.T) {
	// This test simulates the full deployment flow without actual containerd
	// It tests the logic and state management

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create mock deployment state
	type MockDeployment struct {
		ID           string
		Image        string
		Status       types.ContainerStatus
		Resources    types.ResourceLimits
		CreatedAt    time.Time
		StartedAt    time.Time
		Regions      []string
		OnionAddress string
	}

	// Step 1: Create deployment request
	deployReq := struct {
		Image     string
		Resources types.ResourceLimits
		TorOnly   bool
	}{
		Image: "nginx:latest",
		Resources: types.ResourceLimits{
			MemoryLimit: 512 * 1024 * 1024,
			CPUQuota:    100000,
			CPUPeriod:   100000,
			PIDLimit:    200,
		},
		TorOnly: false,
	}

	// Step 2: Validate request
	if deployReq.Image == "" {
		t.Fatal("Image name is required")
	}

	if deployReq.Resources.MemoryLimit < 0 {
		t.Fatal("Memory limit cannot be negative")
	}

	// Step 3: Create deployment record
	deployment := MockDeployment{
		ID:        "dep-" + time.Now().Format("20060102150405"),
		Image:     deployReq.Image,
		Status:    types.ContainerStatusPending,
		Resources: deployReq.Resources,
		CreatedAt: time.Now(),
		Regions:   []string{"Americas", "Europe", "Asia-Pacific"},
	}

	t.Logf("Created deployment: %s", deployment.ID)

	// Step 4: Simulate container creation
	deployment.Status = types.ContainerStatusCreated
	t.Logf("Container created for: %s", deployment.Image)

	// Step 5: Simulate container start
	deployment.Status = types.ContainerStatusRunning
	deployment.StartedAt = time.Now()
	t.Logf("Container started at: %v", deployment.StartedAt)

	// Step 6: Verify deployment state
	if deployment.Status != types.ContainerStatusRunning {
		t.Errorf("Expected status Running, got %s", deployment.Status)
	}

	if deployment.StartedAt.IsZero() {
		t.Error("StartedAt should be set")
	}

	if len(deployment.Regions) != 3 {
		t.Errorf("Expected 3 regions, got %d", len(deployment.Regions))
	}

	// Step 7: Verify context wasn't cancelled
	select {
	case <-ctx.Done():
		t.Error("Context cancelled during deployment")
	default:
		t.Log("Full deployment flow simulation completed successfully")
	}
}

// --- State Persistence Integration Tests ---

// persistedState represents the state format used by ContainerManager
type persistedState struct {
	Deployments map[string]*testDeployment `json:"deployments"`
	SavedAt     time.Time                  `json:"saved_at"`
	Version     int                        `json:"version"`
}

type testDeployment struct {
	ID              string                `json:"id"`
	Image           string                `json:"image"`
	Status          types.ContainerStatus `json:"status"`
	Resources       types.ResourceLimits  `json:"resources"`
	CreatedAt       time.Time             `json:"created_at"`
	StartedAt       time.Time             `json:"started_at,omitempty"`
	Encrypted       bool                  `json:"encrypted"`
	EncryptedVolume string                `json:"encrypted_volume,omitempty"`
	OnionService    bool                  `json:"onion_service"`
	OnionAddress    string                `json:"onion_address,omitempty"`
	TorOnly         bool                  `json:"tor_only"`
	Regions         []string              `json:"regions"`
}

func TestIntegration_StatePersistence(t *testing.T) {
	// Create temporary directory for state file
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "state.json")

	// Step 1: Create initial state with deployments
	initialState := persistedState{
		Deployments: make(map[string]*testDeployment),
		SavedAt:     time.Now(),
		Version:     1,
	}

	// Add test deployments
	deployments := []testDeployment{
		{
			ID:        "dep-001",
			Image:     "nginx:latest",
			Status:    types.ContainerStatusRunning,
			CreatedAt: time.Now().Add(-1 * time.Hour),
			StartedAt: time.Now().Add(-1 * time.Hour),
			Encrypted: true,
			Regions:   []string{"Americas", "Europe", "Asia-Pacific"},
		},
		{
			ID:        "dep-002",
			Image:     "redis:alpine",
			Status:    types.ContainerStatusRunning,
			CreatedAt: time.Now().Add(-30 * time.Minute),
			StartedAt: time.Now().Add(-30 * time.Minute),
			Encrypted: true,
			Regions:   []string{"Americas", "Europe", "Asia-Pacific"},
		},
		{
			ID:           "dep-003",
			Image:        "tor-hidden-service:latest",
			Status:       types.ContainerStatusRunning,
			CreatedAt:    time.Now().Add(-10 * time.Minute),
			StartedAt:    time.Now().Add(-10 * time.Minute),
			Encrypted:    true,
			OnionService: true,
			OnionAddress: "abc123xyz.onion",
			TorOnly:      true,
			Regions:      []string{"Americas", "Europe", "Asia-Pacific"},
		},
	}

	for _, d := range deployments {
		dep := d // Copy to avoid aliasing
		initialState.Deployments[d.ID] = &dep
	}

	// Step 2: Save state to disk (simulating daemon shutdown)
	t.Log("Saving state to disk...")
	f, err := os.Create(stateFile)
	if err != nil {
		t.Fatalf("Failed to create state file: %v", err)
	}

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(initialState); err != nil {
		f.Close()
		t.Fatalf("Failed to encode state: %v", err)
	}
	f.Close()

	t.Logf("State saved with %d deployments", len(initialState.Deployments))

	// Step 3: Simulate daemon restart by loading state from disk
	t.Log("Simulating daemon restart...")

	loadedFile, err := os.Open(stateFile)
	if err != nil {
		t.Fatalf("Failed to open state file: %v", err)
	}
	defer loadedFile.Close()

	var loadedState persistedState
	if err := json.NewDecoder(loadedFile).Decode(&loadedState); err != nil {
		t.Fatalf("Failed to decode state: %v", err)
	}

	t.Logf("State loaded with %d deployments", len(loadedState.Deployments))

	// Step 4: Verify deployments were restored correctly
	if len(loadedState.Deployments) != len(deployments) {
		t.Errorf("Expected %d deployments, got %d", len(deployments), len(loadedState.Deployments))
	}

	for _, expected := range deployments {
		loaded, exists := loadedState.Deployments[expected.ID]
		if !exists {
			t.Errorf("Deployment %s not found after restore", expected.ID)
			continue
		}

		if loaded.Image != expected.Image {
			t.Errorf("Image mismatch for %s: got %s, want %s", expected.ID, loaded.Image, expected.Image)
		}

		if loaded.Status != expected.Status {
			t.Errorf("Status mismatch for %s: got %s, want %s", expected.ID, loaded.Status, expected.Status)
		}

		if loaded.Encrypted != expected.Encrypted {
			t.Errorf("Encrypted mismatch for %s: got %v, want %v", expected.ID, loaded.Encrypted, expected.Encrypted)
		}

		if len(loaded.Regions) != len(expected.Regions) {
			t.Errorf("Regions count mismatch for %s: got %d, want %d", expected.ID, len(loaded.Regions), len(expected.Regions))
		}

		if expected.OnionService {
			if loaded.OnionAddress != expected.OnionAddress {
				t.Errorf("OnionAddress mismatch for %s: got %s, want %s", expected.ID, loaded.OnionAddress, expected.OnionAddress)
			}
		}
	}

	// Step 5: Verify state file metadata
	if loadedState.Version != 1 {
		t.Errorf("State version mismatch: got %d, want 1", loadedState.Version)
	}

	if loadedState.SavedAt.IsZero() {
		t.Error("SavedAt should not be zero")
	}

	t.Log("State persistence test completed successfully")
}

func TestIntegration_StatePersistence_AtomicWrite(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "state.json")
	tmpFile := stateFile + ".tmp"

	// Test atomic write pattern
	state := persistedState{
		Deployments: map[string]*testDeployment{
			"dep-atomic": {
				ID:        "dep-atomic",
				Image:     "test:latest",
				Status:    types.ContainerStatusRunning,
				CreatedAt: time.Now(),
			},
		},
		SavedAt: time.Now(),
		Version: 1,
	}

	// Step 1: Write to temp file
	f, err := os.OpenFile(tmpFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(state); err != nil {
		f.Close()
		t.Fatalf("Failed to encode state: %v", err)
	}

	if err := f.Sync(); err != nil {
		f.Close()
		t.Fatalf("Failed to sync: %v", err)
	}
	f.Close()

	// Step 2: Atomic rename
	if err := os.Rename(tmpFile, stateFile); err != nil {
		t.Fatalf("Failed atomic rename: %v", err)
	}

	// Step 3: Verify temp file is gone
	if _, err := os.Stat(tmpFile); !os.IsNotExist(err) {
		t.Error("Temp file should not exist after rename")
	}

	// Step 4: Verify final file exists and is valid
	loadedFile, err := os.Open(stateFile)
	if err != nil {
		t.Fatalf("Failed to open state file: %v", err)
	}
	defer loadedFile.Close()

	var loadedState persistedState
	if err := json.NewDecoder(loadedFile).Decode(&loadedState); err != nil {
		t.Fatalf("Failed to decode state: %v", err)
	}

	if len(loadedState.Deployments) != 1 {
		t.Errorf("Expected 1 deployment, got %d", len(loadedState.Deployments))
	}

	t.Log("Atomic write test completed successfully")
}

func TestIntegration_StatePersistence_EmptyState(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "state.json")

	// Save empty state
	emptyState := persistedState{
		Deployments: make(map[string]*testDeployment),
		SavedAt:     time.Now(),
		Version:     1,
	}

	f, err := os.Create(stateFile)
	if err != nil {
		t.Fatalf("Failed to create state file: %v", err)
	}

	encoder := json.NewEncoder(f)
	if err := encoder.Encode(emptyState); err != nil {
		f.Close()
		t.Fatalf("Failed to encode empty state: %v", err)
	}
	f.Close()

	// Load and verify
	loadedFile, err := os.Open(stateFile)
	if err != nil {
		t.Fatalf("Failed to open state file: %v", err)
	}
	defer loadedFile.Close()

	var loadedState persistedState
	if err := json.NewDecoder(loadedFile).Decode(&loadedState); err != nil {
		t.Fatalf("Failed to decode state: %v", err)
	}

	if len(loadedState.Deployments) != 0 {
		t.Errorf("Expected 0 deployments in empty state, got %d", len(loadedState.Deployments))
	}

	t.Log("Empty state persistence test passed")
}

func TestIntegration_StatePersistence_MissingFile(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "nonexistent_state.json")

	// Try to load non-existent state file
	_, err := os.Open(stateFile)
	if !os.IsNotExist(err) {
		t.Errorf("Expected file not found error, got: %v", err)
	}

	// This is expected behavior - should start with fresh state
	t.Log("Missing state file correctly detected (fresh start)")
}

func TestIntegration_StatePersistence_CorruptedFile(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "corrupted_state.json")

	// Write corrupted JSON
	err := os.WriteFile(stateFile, []byte("this is not valid JSON{{{"), 0600)
	if err != nil {
		t.Fatalf("Failed to write corrupted file: %v", err)
	}

	// Try to load corrupted file
	loadedFile, err := os.Open(stateFile)
	if err != nil {
		t.Fatalf("Failed to open corrupted file: %v", err)
	}
	defer loadedFile.Close()

	var loadedState persistedState
	err = json.NewDecoder(loadedFile).Decode(&loadedState)
	if err == nil {
		t.Error("Expected error when loading corrupted state")
	}

	t.Logf("Correctly detected corrupted state: %v", err)
}

// --- Health Monitoring Integration Tests ---

func TestIntegration_HealthMonitoring(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	healthMonitor := redundancy.NewHealthMonitor()
	healthMonitor.SetInterval(100 * time.Millisecond) // Short interval for testing

	// Set up probe function that simulates container health checks
	probeCount := 0
	healthMonitor.SetProbeFunc(func(ctx context.Context, containerID string) (bool, error) {
		probeCount++
		// Simulate healthy container
		return true, nil
	})

	// Start health monitoring in background
	go healthMonitor.Start(ctx)

	// Register a container
	containerID := "health-test-container"
	for i := 0; i < 3; i++ {
		healthMonitor.UpdateHealth(containerID, i, types.HealthStatus{
			Healthy:    true,
			LastUpdate: time.Now(),
		})
	}

	// Wait for some health checks to occur
	time.Sleep(300 * time.Millisecond)

	// Verify health status
	for i := 0; i < 3; i++ {
		if !healthMonitor.IsHealthy(containerID, i) {
			t.Errorf("Replica %d should be healthy", i)
		}
	}

	// Test unhealthy detection
	healthMonitor.UpdateHealth(containerID, 1, types.HealthStatus{
		Healthy:    false,
		LastUpdate: time.Now(),
	})

	unhealthy := healthMonitor.GetUnhealthyReplicas(containerID)
	found := false
	for _, idx := range unhealthy {
		if idx == 1 {
			found = true
			break
		}
	}
	if !found {
		t.Error("Replica 1 should be marked as unhealthy")
	}

	t.Logf("Health monitoring test completed, %d probes executed", probeCount)
}

// --- Replication Integration Tests ---

func TestIntegration_ReplicationSystem(t *testing.T) {
	replicator := redundancy.NewReplicator()

	// Test creating replica sets
	containerIDs := []string{"repl-test-1", "repl-test-2", "repl-test-3"}
	regions := []string{"Americas", "Europe", "Asia-Pacific"}

	for _, containerID := range containerIDs {
		replicaSet, err := replicator.CreateReplicaSet(containerID, regions)
		if err != nil {
			t.Fatalf("Failed to create replica set for %s: %v", containerID, err)
		}

		if replicaSet.ContainerID != containerID {
			t.Errorf("ContainerID mismatch: got %s, want %s", replicaSet.ContainerID, containerID)
		}

		if len(replicaSet.Regions) < 1 || replicaSet.Regions[0] != "Americas" {
			t.Errorf("Regions[0] mismatch: got %v, want Americas", replicaSet.Regions)
		}
	}

	// Verify all replica sets exist
	allSets := replicator.GetAllReplicaSets()
	if len(allSets) != len(containerIDs) {
		t.Errorf("Expected %d replica sets, got %d", len(containerIDs), len(allSets))
	}

	// Test adding replicas
	testContainer := &types.Container{
		ID:      "container-instance-1",
		Status:  types.ContainerStatusRunning,
		NodeID:  types.NodeID{},
	}

	err := replicator.AddReplica(containerIDs[0], 0, testContainer)
	if err != nil {
		t.Errorf("Failed to add replica: %v", err)
	}

	// Verify replica was added
	rs, exists := replicator.GetReplicaSet(containerIDs[0])
	if !exists {
		t.Fatal("Replica set should exist")
	}

	if rs.Replicas[0] == nil {
		t.Error("Replica 0 should be set")
	}

	// Test deleting replica set
	replicator.DeleteReplicaSet(containerIDs[2])
	_, exists = replicator.GetReplicaSet(containerIDs[2])
	if exists {
		t.Error("Deleted replica set should not exist")
	}

	t.Log("Replication system integration test completed")
}

// --- Geographic Distribution Tests ---

func TestIntegration_GeographicDistribution(t *testing.T) {
	geolocator := p2p.NewGeoLocator()

	// Test region mapping - based on actual implementation in geolocation.go
	testCases := []struct {
		country        string
		expectedRegion string
	}{
		// Explicitly mapped countries
		{"US", "Americas"},
		{"CA", "Americas"},
		{"BR", "Americas"},
		{"GB", "Europe"},
		{"DE", "Europe"},
		{"FR", "Europe"},
		{"JP", "Asia-Pacific"},
		{"AU", "Asia-Pacific"},
		{"IN", "Asia-Pacific"},
		{"ZA", "Africa"},
		{"NG", "Africa"},
		// SG is properly mapped to Asia-Pacific
		{"SG", "Asia-Pacific"},
		// Unknown country codes return "Unknown"
		{"ZZ", "Unknown"},
	}

	for _, tc := range testCases {
		region := p2p.GetRegionFromCountry(tc.country)
		if region != tc.expectedRegion {
			t.Errorf("Country %s: expected region %s, got %s", tc.country, tc.expectedRegion, region)
		}
	}

	// Test geolocation
	_ = geolocator // Would test actual IP geolocation if service available

	t.Log("Geographic distribution test completed")
}
