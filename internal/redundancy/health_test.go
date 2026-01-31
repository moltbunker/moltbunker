package redundancy

import (
	"context"
	"testing"
	"time"

	"github.com/moltbunker/moltbunker/pkg/types"
)

func TestHealthMonitor_UpdateHealth(t *testing.T) {
	hm := NewHealthMonitor()

	containerID := "test-container"
	replicaIndex := 0

	health := types.HealthStatus{
		CPUUsage:    50.0,
		MemoryUsage: 1024 * 1024 * 100, // 100MB
		DiskUsage:   1024 * 1024 * 500, // 500MB
		NetworkIn:   1024 * 100,
		NetworkOut:  1024 * 200,
		LastUpdate:  time.Now(),
		Healthy:     true,
	}

	hm.UpdateHealth(containerID, replicaIndex, health)

	retrievedHealth, exists := hm.GetHealth(containerID, replicaIndex)
	if !exists {
		t.Fatal("Health should exist")
	}

	if retrievedHealth.Health.CPUUsage != health.CPUUsage {
		t.Errorf("CPU usage mismatch: got %f, want %f", retrievedHealth.Health.CPUUsage, health.CPUUsage)
	}

	if !retrievedHealth.Healthy {
		t.Error("Health should be marked as healthy")
	}
}

func TestHealthMonitor_IsHealthy(t *testing.T) {
	hm := NewHealthMonitor()

	containerID := "test-container"
	replicaIndex := 0

	health := types.HealthStatus{
		Healthy:    true,
		LastUpdate: time.Now(),
	}

	hm.UpdateHealth(containerID, replicaIndex, health)

	if !hm.IsHealthy(containerID, replicaIndex) {
		t.Error("Replica should be healthy")
	}
}

func TestHealthMonitor_IsHealthy_Timeout(t *testing.T) {
	hm := NewHealthMonitor()
	hm.timeout = 1 * time.Second

	containerID := "test-container"
	replicaIndex := 0

	health := types.HealthStatus{
		Healthy:    true,
		LastUpdate: time.Now().Add(-2 * time.Second), // 2 seconds ago
	}

	hm.UpdateHealth(containerID, replicaIndex, health)

	if hm.IsHealthy(containerID, replicaIndex) {
		t.Error("Replica should be unhealthy due to timeout")
	}
}

func TestHealthMonitor_GetUnhealthyReplicas(t *testing.T) {
	hm := NewHealthMonitor()
	hm.timeout = 1 * time.Second

	containerID := "test-container"

	// Add healthy replica
	hm.UpdateHealth(containerID, 0, types.HealthStatus{
		Healthy:    true,
		LastUpdate: time.Now(),
	})

	// Add unhealthy replica (old timestamp)
	hm.UpdateHealth(containerID, 1, types.HealthStatus{
		Healthy:    true,
		LastUpdate: time.Now().Add(-2 * time.Second),
	})

	// Add unhealthy replica (marked as unhealthy)
	hm.UpdateHealth(containerID, 2, types.HealthStatus{
		Healthy:    false,
		LastUpdate: time.Now(),
	})

	unhealthy := hm.GetUnhealthyReplicas(containerID)
	if len(unhealthy) != 2 {
		t.Errorf("Should have 2 unhealthy replicas, got %d", len(unhealthy))
	}
}

func TestHealthMonitor_Start(t *testing.T) {
	hm := NewHealthMonitor()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start health monitor
	go hm.Start(ctx)

	// Give it a moment to start
	time.Sleep(100 * time.Millisecond)

	// Cancel context to stop
	cancel()

	// Give it a moment to stop
	time.Sleep(100 * time.Millisecond)
}
