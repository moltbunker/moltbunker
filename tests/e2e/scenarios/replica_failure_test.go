//go:build e2e

package scenarios

import (
	"testing"
	"time"

	"github.com/moltbunker/moltbunker/internal/redundancy"
	"github.com/moltbunker/moltbunker/pkg/types"
	"github.com/moltbunker/moltbunker/tests/e2e/testutil"
)

// ---------------------------------------------------------------------------
// TestE2E_ReplicaFailureRecovery
//
// Creates a 3-replica set, marks one replica as unhealthy, and verifies
// that the health monitor detects the failure. Then triggers re-replication
// by replacing the failed replica and verifies the new replica is healthy.
// ---------------------------------------------------------------------------

func TestE2E_ReplicaFailureRecovery(t *testing.T) {
	h := testutil.NewTestHarness(t)
	assert := testutil.NewAssertions(t)

	ctx, cancel := h.WithTimeout(60 * time.Second)
	defer cancel()

	containerID := "e2e-replica-failure-recovery"
	regions := []string{"us-east", "eu-west", "ap-southeast"}

	// Create replicator and health monitor.
	replicator := redundancy.NewReplicator()
	healthMonitor := redundancy.NewHealthMonitor()
	consensus := redundancy.NewConsensusManager()

	// Phase 1: Create the replica set with 3 replicas.
	t.Log("phase 1: creating replica set with 3 replicas")

	replicaSet, err := replicator.CreateReplicaSet(containerID, regions)
	assert.NoError(err, "CreateReplicaSet should succeed")
	assert.Equal(containerID, replicaSet.ContainerID)
	assert.Equal("us-east", replicaSet.Region1)
	assert.Equal("eu-west", replicaSet.Region2)
	assert.Equal("ap-southeast", replicaSet.Region3)

	// Create mock containers for each replica using the mock containerd.
	replicas := make([]*types.Container, 3)
	for i := 0; i < 3; i++ {
		replicaID := containerID + "-replica-" + string(rune('0'+i))
		_, err := h.Containerd.CreateContainer(ctx, replicaID, "nginx:latest", types.ResourceLimits{
			MemoryLimit: 256 * 1024 * 1024,
		})
		assert.NoError(err, "Create replica container")

		err = h.Containerd.StartContainer(ctx, replicaID)
		assert.NoError(err, "Start replica container")

		replicas[i] = &types.Container{
			ID:       replicaID,
			ImageCID: "bafytest123",
			Status:   types.ContainerStatusRunning,
		}

		err = replicator.AddReplica(containerID, i, replicas[i])
		assert.NoError(err, "AddReplica should succeed")

		// Register healthy status in the health monitor.
		healthMonitor.UpdateHealth(containerID, i, types.HealthStatus{
			Healthy:    true,
			LastUpdate: time.Now(),
			CPUUsage:   10.0,
		})
	}

	// Update consensus state.
	rs, found := replicator.GetReplicaSet(containerID)
	assert.True(found, "ReplicaSet should exist")
	consensus.UpdateState(containerID, types.ContainerStatusRunning, rs.Replicas)

	// Verify all replicas are healthy.
	for i := 0; i < 3; i++ {
		assert.True(healthMonitor.IsHealthy(containerID, i), "Replica %d should be healthy initially")
	}
	t.Log("phase 1 complete: all 3 replicas created and healthy")

	// Phase 2: Kill replica 1 (mark unhealthy).
	t.Log("phase 2: failing replica 1")

	healthMonitor.UpdateHealth(containerID, 1, types.HealthStatus{
		Healthy:    false,
		LastUpdate: time.Now(),
		CPUUsage:   0,
	})

	// Verify health monitor detects the failure.
	assert.False(healthMonitor.IsHealthy(containerID, 1), "Replica 1 should be unhealthy")
	assert.True(healthMonitor.IsHealthy(containerID, 0), "Replica 0 should still be healthy")
	assert.True(healthMonitor.IsHealthy(containerID, 2), "Replica 2 should still be healthy")

	unhealthyReplicas := healthMonitor.GetUnhealthyReplicas(containerID)
	assert.True(len(unhealthyReplicas) >= 1, "Should have at least 1 unhealthy replica")
	t.Logf("unhealthy replicas detected: %v", unhealthyReplicas)

	// Phase 3: Re-replicate by replacing the failed replica.
	t.Log("phase 3: replacing failed replica")

	newReplicaID := containerID + "-replica-1-replacement"
	_, err = h.Containerd.CreateContainer(ctx, newReplicaID, "nginx:latest", types.ResourceLimits{
		MemoryLimit: 256 * 1024 * 1024,
	})
	assert.NoError(err, "Create replacement replica")

	err = h.Containerd.StartContainer(ctx, newReplicaID)
	assert.NoError(err, "Start replacement replica")

	newContainer := &types.Container{
		ID:       newReplicaID,
		ImageCID: "bafytest123",
		Status:   types.ContainerStatusRunning,
	}

	err = replicator.ReplaceReplica(ctx, containerID, 1, newContainer)
	assert.NoError(err, "ReplaceReplica should succeed")

	// Update health for the replacement.
	healthMonitor.UpdateHealth(containerID, 1, types.HealthStatus{
		Healthy:    true,
		LastUpdate: time.Now(),
		CPUUsage:   8.0,
	})

	// Phase 4: Verify recovery is complete.
	t.Log("phase 4: verifying recovery")

	for i := 0; i < 3; i++ {
		assert.True(healthMonitor.IsHealthy(containerID, i),
			"All replicas should be healthy after recovery")
	}

	unhealthyAfter := healthMonitor.GetUnhealthyReplicas(containerID)
	assert.Len(unhealthyAfter, 0, "No unhealthy replicas after recovery")

	// Verify the replica set has the replacement container.
	rsAfter, found := replicator.GetReplicaSet(containerID)
	assert.True(found)
	assert.Equal(newReplicaID, rsAfter.Replicas[1].ID,
		"Replica 1 should be the replacement container")

	// Verify consensus can determine status.
	consensus.UpdateState(containerID, types.ContainerStatusRunning, rsAfter.Replicas)
	consensusStatus, err := consensus.GetConsensusStatus(containerID)
	assert.NoError(err)
	assert.Equal(types.ContainerStatusRunning, consensusStatus)

	t.Log("replica failure recovery test passed")

	// Cleanup mock containers.
	for i := 0; i < 3; i++ {
		replicaID := containerID + "-replica-" + string(rune('0'+i))
		_ = h.Containerd.DeleteContainer(ctx, replicaID)
	}
	_ = h.Containerd.DeleteContainer(ctx, newReplicaID)
}

// ---------------------------------------------------------------------------
// TestE2E_MultipleReplicaFailure
//
// Creates a 3-replica set, then kills 2 replicas simultaneously. Verifies
// that the system detects insufficient replicas (only 1 out of 3 healthy).
// ---------------------------------------------------------------------------

func TestE2E_MultipleReplicaFailure(t *testing.T) {
	h := testutil.NewTestHarness(t)
	assert := testutil.NewAssertions(t)

	ctx, cancel := h.WithTimeout(60 * time.Second)
	defer cancel()

	containerID := "e2e-multi-replica-failure"
	regions := []string{"us-east", "eu-west", "ap-southeast"}

	replicator := redundancy.NewReplicator()
	healthMonitor := redundancy.NewHealthMonitor()
	consensus := redundancy.NewConsensusManager()

	// Create replica set and register all replicas as healthy.
	t.Log("phase 1: creating 3 healthy replicas")

	replicaSet, err := replicator.CreateReplicaSet(containerID, regions)
	assert.NoError(err)
	_ = replicaSet

	replicas := make([]*types.Container, 3)
	for i := 0; i < 3; i++ {
		replicaID := containerID + "-replica-" + string(rune('0'+i))
		_, err := h.Containerd.CreateContainer(ctx, replicaID, "alpine:latest", types.ResourceLimits{
			MemoryLimit: 128 * 1024 * 1024,
		})
		assert.NoError(err)

		err = h.Containerd.StartContainer(ctx, replicaID)
		assert.NoError(err)

		replicas[i] = &types.Container{
			ID:       replicaID,
			ImageCID: "bafytest456",
			Status:   types.ContainerStatusRunning,
		}

		err = replicator.AddReplica(containerID, i, replicas[i])
		assert.NoError(err)

		healthMonitor.UpdateHealth(containerID, i, types.HealthStatus{
			Healthy:    true,
			LastUpdate: time.Now(),
		})
	}

	rs, _ := replicator.GetReplicaSet(containerID)
	consensus.UpdateState(containerID, types.ContainerStatusRunning, rs.Replicas)

	t.Log("phase 1 complete: 3 replicas running and healthy")

	// Phase 2: Kill replicas 0 and 2 simultaneously.
	t.Log("phase 2: killing replicas 0 and 2 simultaneously")

	healthMonitor.UpdateHealth(containerID, 0, types.HealthStatus{
		Healthy:    false,
		LastUpdate: time.Now(),
	})
	healthMonitor.UpdateHealth(containerID, 2, types.HealthStatus{
		Healthy:    false,
		LastUpdate: time.Now(),
	})

	// Verify detection.
	assert.False(healthMonitor.IsHealthy(containerID, 0), "Replica 0 should be unhealthy")
	assert.True(healthMonitor.IsHealthy(containerID, 1), "Replica 1 should still be healthy")
	assert.False(healthMonitor.IsHealthy(containerID, 2), "Replica 2 should be unhealthy")

	unhealthy := healthMonitor.GetUnhealthyReplicas(containerID)
	assert.True(len(unhealthy) >= 2,
		"Should detect at least 2 unhealthy replicas")
	t.Logf("detected %d unhealthy replicas: %v", len(unhealthy), unhealthy)

	// Verify consensus detects degraded state.
	// With 2 failed replicas, only 1 is running - less than 2/3 majority.
	replicas[0].Status = types.ContainerStatusFailed
	replicas[2].Status = types.ContainerStatusFailed
	consensus.UpdateState(containerID, types.ContainerStatusFailed, rs.Replicas)

	// Count healthy replicas.
	healthyCount := 0
	for i := 0; i < 3; i++ {
		if healthMonitor.IsHealthy(containerID, i) {
			healthyCount++
		}
	}
	assert.Equal(1, healthyCount, "Only 1 replica should be healthy")
	t.Logf("healthy replicas remaining: %d/3 (insufficient for redundancy)", healthyCount)

	// Phase 3: Verify recovery is possible by replacing both failed replicas.
	t.Log("phase 3: replacing both failed replicas")

	for _, failedIdx := range []int{0, 2} {
		newID := containerID + "-replacement-" + string(rune('0'+failedIdx))
		_, err := h.Containerd.CreateContainer(ctx, newID, "alpine:latest", types.ResourceLimits{
			MemoryLimit: 128 * 1024 * 1024,
		})
		assert.NoError(err)

		err = h.Containerd.StartContainer(ctx, newID)
		assert.NoError(err)

		newContainer := &types.Container{
			ID:       newID,
			ImageCID: "bafytest456",
			Status:   types.ContainerStatusRunning,
		}

		err = replicator.ReplaceReplica(ctx, containerID, failedIdx, newContainer)
		assert.NoError(err)

		healthMonitor.UpdateHealth(containerID, failedIdx, types.HealthStatus{
			Healthy:    true,
			LastUpdate: time.Now(),
		})
	}

	// Verify full recovery.
	for i := 0; i < 3; i++ {
		assert.True(healthMonitor.IsHealthy(containerID, i),
			"All replicas should be healthy after recovery")
	}

	unhealthyAfter := healthMonitor.GetUnhealthyReplicas(containerID)
	assert.Len(unhealthyAfter, 0, "No unhealthy replicas after recovery")

	t.Log("multiple replica failure test passed")

	// Cleanup.
	for i := 0; i < 3; i++ {
		_ = h.Containerd.DeleteContainer(ctx, containerID+"-replica-"+string(rune('0'+i)))
	}
	for _, idx := range []int{0, 2} {
		_ = h.Containerd.DeleteContainer(ctx, containerID+"-replacement-"+string(rune('0'+idx)))
	}
}

// ---------------------------------------------------------------------------
// TestE2E_CascadingFailure
//
// Simulates a cascading failure scenario: replica 1 fails and replacement
// begins, then replica 2 fails during the replacement process. Verifies
// the system handles overlapping failures gracefully.
// ---------------------------------------------------------------------------

func TestE2E_CascadingFailure(t *testing.T) {
	h := testutil.NewTestHarness(t)
	assert := testutil.NewAssertions(t)

	ctx, cancel := h.WithTimeout(60 * time.Second)
	defer cancel()

	containerID := "e2e-cascading-failure"
	regions := []string{"us-east", "eu-west", "ap-southeast"}

	replicator := redundancy.NewReplicator()
	healthMonitor := redundancy.NewHealthMonitor()
	consensus := redundancy.NewConsensusManager()

	// Create initial 3 healthy replicas.
	t.Log("phase 1: creating 3 healthy replicas")

	_, err := replicator.CreateReplicaSet(containerID, regions)
	assert.NoError(err)

	replicas := make([]*types.Container, 3)
	for i := 0; i < 3; i++ {
		replicaID := containerID + "-replica-" + string(rune('0'+i))
		_, err := h.Containerd.CreateContainer(ctx, replicaID, "redis:latest", types.ResourceLimits{
			MemoryLimit: 256 * 1024 * 1024,
		})
		assert.NoError(err)

		err = h.Containerd.StartContainer(ctx, replicaID)
		assert.NoError(err)

		replicas[i] = &types.Container{
			ID:       replicaID,
			ImageCID: "bafytest789",
			Status:   types.ContainerStatusRunning,
		}

		err = replicator.AddReplica(containerID, i, replicas[i])
		assert.NoError(err)

		healthMonitor.UpdateHealth(containerID, i, types.HealthStatus{
			Healthy:    true,
			LastUpdate: time.Now(),
		})
	}

	rs, _ := replicator.GetReplicaSet(containerID)
	consensus.UpdateState(containerID, types.ContainerStatusRunning, rs.Replicas)
	t.Log("phase 1 complete: all 3 replicas running")

	// Phase 2: Fail replica 1.
	t.Log("phase 2: failing replica 1")

	healthMonitor.UpdateHealth(containerID, 1, types.HealthStatus{
		Healthy:    false,
		LastUpdate: time.Now(),
	})

	assert.False(healthMonitor.IsHealthy(containerID, 1))

	unhealthy1 := healthMonitor.GetUnhealthyReplicas(containerID)
	assert.True(len(unhealthy1) >= 1, "Should detect replica 1 as unhealthy")
	t.Logf("first failure detected: unhealthy=%v", unhealthy1)

	// Phase 3: Begin replacement of replica 1 (simulate in-progress replacement).
	t.Log("phase 3: starting replacement of replica 1 (in progress)")

	replacement1ID := containerID + "-replacement-1"
	_, err = h.Containerd.CreateContainer(ctx, replacement1ID, "redis:latest", types.ResourceLimits{
		MemoryLimit: 256 * 1024 * 1024,
	})
	assert.NoError(err)
	// Note: NOT yet started -- replacement is "in progress".

	// Phase 4: While replacement is in progress, fail replica 2 as well.
	t.Log("phase 4: failing replica 2 during replacement")

	healthMonitor.UpdateHealth(containerID, 2, types.HealthStatus{
		Healthy:    false,
		LastUpdate: time.Now(),
	})

	assert.False(healthMonitor.IsHealthy(containerID, 2))

	// Now we have 2 unhealthy replicas with one replacement in flight.
	unhealthy2 := healthMonitor.GetUnhealthyReplicas(containerID)
	assert.True(len(unhealthy2) >= 2,
		"Should detect 2 unhealthy replicas during cascading failure")
	t.Logf("cascading failure detected: unhealthy=%v", unhealthy2)

	// Only replica 0 should be healthy.
	assert.True(healthMonitor.IsHealthy(containerID, 0), "Replica 0 should still be healthy")

	// Phase 5: Complete both replacements.
	t.Log("phase 5: completing both replacements")

	// Finish replica 1 replacement.
	err = h.Containerd.StartContainer(ctx, replacement1ID)
	assert.NoError(err)

	replacementContainer1 := &types.Container{
		ID:       replacement1ID,
		ImageCID: "bafytest789",
		Status:   types.ContainerStatusRunning,
	}
	err = replicator.ReplaceReplica(ctx, containerID, 1, replacementContainer1)
	assert.NoError(err)

	healthMonitor.UpdateHealth(containerID, 1, types.HealthStatus{
		Healthy:    true,
		LastUpdate: time.Now(),
	})

	// Now replace replica 2.
	replacement2ID := containerID + "-replacement-2"
	_, err = h.Containerd.CreateContainer(ctx, replacement2ID, "redis:latest", types.ResourceLimits{
		MemoryLimit: 256 * 1024 * 1024,
	})
	assert.NoError(err)

	err = h.Containerd.StartContainer(ctx, replacement2ID)
	assert.NoError(err)

	replacementContainer2 := &types.Container{
		ID:       replacement2ID,
		ImageCID: "bafytest789",
		Status:   types.ContainerStatusRunning,
	}
	err = replicator.ReplaceReplica(ctx, containerID, 2, replacementContainer2)
	assert.NoError(err)

	healthMonitor.UpdateHealth(containerID, 2, types.HealthStatus{
		Healthy:    true,
		LastUpdate: time.Now(),
	})

	// Phase 6: Verify full recovery from cascading failure.
	t.Log("phase 6: verifying full recovery from cascading failure")

	for i := 0; i < 3; i++ {
		assert.True(healthMonitor.IsHealthy(containerID, i),
			"All replicas should be healthy after cascading recovery")
	}

	unhealthyFinal := healthMonitor.GetUnhealthyReplicas(containerID)
	assert.Len(unhealthyFinal, 0, "No unhealthy replicas after cascading recovery")

	// Verify the replaced replicas have correct IDs.
	rsFinal, found := replicator.GetReplicaSet(containerID)
	assert.True(found)
	assert.Equal(replacement1ID, rsFinal.Replicas[1].ID)
	assert.Equal(replacement2ID, rsFinal.Replicas[2].ID)

	// Update consensus and verify.
	consensus.UpdateState(containerID, types.ContainerStatusRunning, rsFinal.Replicas)
	status, err := consensus.GetConsensusStatus(containerID)
	assert.NoError(err)
	assert.Equal(types.ContainerStatusRunning, status)

	t.Log("cascading failure recovery test passed")

	// Cleanup.
	for i := 0; i < 3; i++ {
		_ = h.Containerd.DeleteContainer(ctx, containerID+"-replica-"+string(rune('0'+i)))
	}
	_ = h.Containerd.DeleteContainer(ctx, replacement1ID)
	_ = h.Containerd.DeleteContainer(ctx, replacement2ID)
}

// ---------------------------------------------------------------------------
// TestE2E_ReplicaRecoveryTiming
//
// Verifies that a failed replica is detected within an expected time window
// by using the health monitor's heartbeat timeout mechanism. Measures the
// time from failure injection to detection.
// ---------------------------------------------------------------------------

func TestE2E_ReplicaRecoveryTiming(t *testing.T) {
	h := testutil.NewTestHarness(t)
	assert := testutil.NewAssertions(t)

	_, cancel := h.WithTimeout(60 * time.Second)
	defer cancel()

	containerID := "e2e-replica-timing"
	regions := []string{"us-east", "eu-west", "ap-southeast"}

	replicator := redundancy.NewReplicator()
	healthMonitor := redundancy.NewHealthMonitor()

	// Use a short timeout for faster testing.
	healthMonitor.SetInterval(100 * time.Millisecond)

	_, err := replicator.CreateReplicaSet(containerID, regions)
	assert.NoError(err)

	// Create 3 healthy replicas with recent heartbeats.
	t.Log("phase 1: creating 3 healthy replicas")

	for i := 0; i < 3; i++ {
		container := &types.Container{
			ID:       containerID + "-timing-" + string(rune('0'+i)),
			ImageCID: "bafytiming",
			Status:   types.ContainerStatusRunning,
		}
		err = replicator.AddReplica(containerID, i, container)
		assert.NoError(err)

		healthMonitor.UpdateHealth(containerID, i, types.HealthStatus{
			Healthy:    true,
			LastUpdate: time.Now(),
		})
	}

	// Verify initial health.
	for i := 0; i < 3; i++ {
		assert.True(healthMonitor.IsHealthy(containerID, i))
	}
	t.Log("phase 1 complete: all replicas healthy")

	// Phase 2: Simulate failure by marking unhealthy and record detection time.
	t.Log("phase 2: injecting failure and measuring detection time")

	failureStart := time.Now()

	healthMonitor.UpdateHealth(containerID, 1, types.HealthStatus{
		Healthy:    false,
		LastUpdate: time.Now(),
	})

	// Immediately verify the health monitor reports unhealthy.
	detectedUnhealthy := !healthMonitor.IsHealthy(containerID, 1)
	detectionTime := time.Since(failureStart)

	assert.True(detectedUnhealthy, "Failure should be detected immediately after UpdateHealth")
	t.Logf("failure detection time: %v", detectionTime)

	// Detection via UpdateHealth should be essentially instantaneous.
	assert.True(detectionTime < 100*time.Millisecond,
		"Direct failure detection should be under 100ms")

	// Phase 3: Measure the time to complete a full recovery cycle.
	t.Log("phase 3: measuring recovery cycle time")

	recoveryStart := time.Now()

	// Simulate creating and starting a replacement.
	healthMonitor.UpdateHealth(containerID, 1, types.HealthStatus{
		Healthy:    true,
		LastUpdate: time.Now(),
	})

	recoveryTime := time.Since(recoveryStart)
	totalTime := time.Since(failureStart)

	assert.True(healthMonitor.IsHealthy(containerID, 1),
		"Replica should be healthy after recovery")
	t.Logf("recovery update time: %v", recoveryTime)
	t.Logf("total failure-to-recovery time: %v", totalTime)

	// Recovery should also be fast (just a state update).
	assert.True(recoveryTime < 100*time.Millisecond,
		"Recovery state update should be under 100ms")

	// Phase 4: Test heartbeat-based timeout detection.
	// Set a very old heartbeat so the timeout check flags it.
	t.Log("phase 4: testing heartbeat timeout detection")

	healthMonitor.UpdateHealth(containerID, 2, types.HealthStatus{
		Healthy:    true,
		LastUpdate: time.Now().Add(-5 * time.Minute), // Simulate stale heartbeat
	})

	// IsHealthy checks both the Healthy flag and heartbeat freshness.
	// With a 30s default timeout, a 5-minute-old heartbeat should fail.
	isHealthy := healthMonitor.IsHealthy(containerID, 2)
	assert.False(isHealthy,
		"Replica with stale heartbeat should be detected as unhealthy")

	// Verify GetUnhealthyReplicas catches the stale replica.
	unhealthy := healthMonitor.GetUnhealthyReplicas(containerID)
	foundStale := false
	for _, idx := range unhealthy {
		if idx == 2 {
			foundStale = true
			break
		}
	}
	assert.True(foundStale, "Stale heartbeat replica should appear in unhealthy list")

	t.Log("replica recovery timing test passed")
}
