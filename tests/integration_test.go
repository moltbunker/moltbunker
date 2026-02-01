package tests

import (
	"context"
	"testing"
	"time"

	"github.com/moltbunker/moltbunker/internal/p2p"
	"github.com/moltbunker/moltbunker/internal/redundancy"
	"github.com/moltbunker/moltbunker/pkg/types"
)

// Integration tests require actual services running
// These are skipped by default unless -short=false is used

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
