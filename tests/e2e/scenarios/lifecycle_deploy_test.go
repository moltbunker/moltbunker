//go:build e2e

package scenarios

import (
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/moltbunker/moltbunker/internal/payment"
	"github.com/moltbunker/moltbunker/internal/redundancy"
	"github.com/moltbunker/moltbunker/pkg/types"
	"github.com/moltbunker/moltbunker/tests/e2e/testutil"
)

// TestE2E_FullDeploymentLifecycle tests the complete deployment lifecycle:
// create deployment with 3 replicas across regions, verify all running,
// simulate one failure, verify failover and re-replication, then cleanup.
func TestE2E_FullDeploymentLifecycle(t *testing.T) {
	h := testutil.NewTestHarness(t)
	assert := testutil.NewAssertions(t)

	ctx, cancel := h.WithTimeout(120 * time.Second)
	defer cancel()

	containerID := "e2e-full-lifecycle"
	imageRef := "docker.io/library/nginx:latest"
	regions := []string{"Americas", "Europe", "Asia-Pacific"}
	resources := types.ResourceLimits{
		CPUQuota:    100000,
		CPUPeriod:   100000,
		MemoryLimit: 256 * 1024 * 1024, // 256MB
		DiskLimit:   1 * 1024 * 1024 * 1024,
		PIDLimit:    100,
	}

	replicator := redundancy.NewReplicator()
	healthMonitor := redundancy.NewHealthMonitor()
	consensus := redundancy.NewConsensusManager()

	// ---------------------------------------------------------------
	// Phase 1: Create ReplicaSet across 3 geographic regions
	// ---------------------------------------------------------------
	t.Log("Phase 1: Creating 3-region replica set")
	replicaSet, err := replicator.CreateReplicaSet(containerID, regions)
	assert.NoError(err, "CreateReplicaSet should succeed")
	assert.Equal("Americas", replicaSet.Region1)
	assert.Equal("Europe", replicaSet.Region2)
	assert.Equal("Asia-Pacific", replicaSet.Region3)

	// ---------------------------------------------------------------
	// Phase 2: Deploy container replicas on mock containerd (3 replicas)
	// ---------------------------------------------------------------
	t.Log("Phase 2: Deploying container replicas")
	replicaContainerIDs := make([]string, 3)
	replicaContainers := [3]*types.Container{}

	for i := 0; i < 3; i++ {
		replicaID := fmt.Sprintf("%s-replica-%d", containerID, i)
		replicaContainerIDs[i] = replicaID

		_, err := h.Containerd.CreateContainer(ctx, replicaID, imageRef, resources)
		assert.NoError(err, "Create replica %d should succeed", i)

		err = h.Containerd.StartContainer(ctx, replicaID)
		assert.NoError(err, "Start replica %d should succeed", i)

		// Wait for container to be running
		err = h.WaitForContainer(replicaID, string(types.ContainerStatusRunning), 10*time.Second)
		assert.NoError(err, "Replica %d should reach running state", i)

		// Register replica in the replicator
		container := &types.Container{
			ID:        replicaID,
			ImageCID:  imageRef,
			Status:    types.ContainerStatusRunning,
			Resources: resources,
			CreatedAt: time.Now(),
			Health: types.HealthStatus{
				Healthy:    true,
				LastUpdate: time.Now(),
			},
		}
		replicaContainers[i] = container

		err = replicator.AddReplica(containerID, i, container)
		assert.NoError(err, "AddReplica %d should succeed", i)
	}

	// ---------------------------------------------------------------
	// Phase 3: Verify all replicas are running via health monitor
	// ---------------------------------------------------------------
	t.Log("Phase 3: Verifying all replicas healthy")
	for i := 0; i < 3; i++ {
		healthMonitor.UpdateHealth(containerID, i, types.HealthStatus{
			Healthy:     true,
			CPUUsage:    20.0 + float64(i)*5.0,
			MemoryUsage: 100 * 1024 * 1024,
			LastUpdate:  time.Now(),
		})
	}

	for i := 0; i < 3; i++ {
		assert.True(healthMonitor.IsHealthy(containerID, i),
			"Replica %d should be healthy", i)
	}

	unhealthy := healthMonitor.GetUnhealthyReplicas(containerID)
	assert.Len(unhealthy, 0, "No replicas should be unhealthy initially")

	// Verify consensus: all running
	consensus.UpdateState(containerID, types.ContainerStatusRunning, replicaContainers)
	status, err := consensus.GetConsensusStatus(containerID)
	assert.NoError(err)
	assert.Equal(types.ContainerStatusRunning, status, "Consensus should be running")

	// ---------------------------------------------------------------
	// Phase 4: Verify health check via containerd mock
	// ---------------------------------------------------------------
	t.Log("Phase 4: Running health checks on mock containers")
	for i := 0; i < 3; i++ {
		health, err := h.Containerd.GetHealthStatus(ctx, replicaContainerIDs[i])
		assert.NoError(err, "Health check for replica %d should succeed", i)
		assert.True(health.Healthy, "Replica %d mock health should be healthy", i)
	}

	// ---------------------------------------------------------------
	// Phase 5: Simulate failure of replica 1 (Europe)
	// ---------------------------------------------------------------
	t.Log("Phase 5: Simulating failure of replica 1 (Europe)")

	// Stop the container to simulate failure
	err = h.Containerd.StopContainer(ctx, replicaContainerIDs[1], 5*time.Second)
	assert.NoError(err, "Stopping replica 1 should succeed")

	// Update health monitor to reflect failure
	healthMonitor.UpdateHealth(containerID, 1, types.HealthStatus{
		Healthy:    false,
		LastUpdate: time.Now(),
	})

	// Update consensus with failure
	replicaContainers[1] = &types.Container{
		ID:     replicaContainerIDs[1],
		Status: types.ContainerStatusFailed,
	}
	consensus.UpdateState(containerID, types.ContainerStatusRunning, replicaContainers)

	// Verify failure detected
	assert.False(healthMonitor.IsHealthy(containerID, 1), "Replica 1 should be unhealthy")
	unhealthy = healthMonitor.GetUnhealthyReplicas(containerID)
	assert.Len(unhealthy, 1, "Should have 1 unhealthy replica")
	assert.Equal(1, unhealthy[0], "Replica 1 should be the unhealthy one")

	// Consensus should still report running (2/3 majority)
	status, err = consensus.GetConsensusStatus(containerID)
	assert.NoError(err)
	assert.Equal(types.ContainerStatusRunning, status, "2/3 consensus should still be running")

	// ---------------------------------------------------------------
	// Phase 6: Failover - replace failed replica
	// ---------------------------------------------------------------
	t.Log("Phase 6: Performing failover - replacing failed replica")

	// Delete the failed container
	err = h.Containerd.DeleteContainer(ctx, replicaContainerIDs[1])
	assert.NoError(err, "Delete failed replica should succeed")

	// Deploy replacement
	replacementID := fmt.Sprintf("%s-replica-1-replacement", containerID)
	replicaContainerIDs[1] = replacementID

	_, err = h.Containerd.CreateContainer(ctx, replacementID, imageRef, resources)
	assert.NoError(err, "Create replacement replica should succeed")

	err = h.Containerd.StartContainer(ctx, replacementID)
	assert.NoError(err, "Start replacement replica should succeed")

	err = h.WaitForContainer(replacementID, string(types.ContainerStatusRunning), 10*time.Second)
	assert.NoError(err, "Replacement replica should reach running state")

	// Update replicator with replacement
	replacementContainer := &types.Container{
		ID:        replacementID,
		ImageCID:  imageRef,
		Status:    types.ContainerStatusRunning,
		Resources: resources,
		CreatedAt: time.Now(),
		Health: types.HealthStatus{
			Healthy:    true,
			LastUpdate: time.Now(),
		},
	}
	err = replicator.ReplaceReplica(nil, containerID, 1, replacementContainer)
	assert.NoError(err, "ReplaceReplica should succeed")

	// Update health for replacement
	healthMonitor.UpdateHealth(containerID, 1, types.HealthStatus{
		Healthy:     true,
		CPUUsage:    22.0,
		MemoryUsage: 105 * 1024 * 1024,
		LastUpdate:  time.Now(),
	})

	// Update consensus with the replacement
	replicaContainers[1] = replacementContainer
	consensus.UpdateState(containerID, types.ContainerStatusRunning, replicaContainers)

	// ---------------------------------------------------------------
	// Phase 7: Verify re-replication is complete
	// ---------------------------------------------------------------
	t.Log("Phase 7: Verifying re-replication complete")

	for i := 0; i < 3; i++ {
		assert.True(healthMonitor.IsHealthy(containerID, i),
			"Replica %d should be healthy after failover", i)
	}

	unhealthy = healthMonitor.GetUnhealthyReplicas(containerID)
	assert.Len(unhealthy, 0, "No replicas should be unhealthy after failover")

	// Verify full consensus
	status, err = consensus.GetConsensusStatus(containerID)
	assert.NoError(err)
	assert.Equal(types.ContainerStatusRunning, status, "Full consensus should be running")

	// Verify replica set integrity
	rs, exists := replicator.GetReplicaSet(containerID)
	assert.True(exists, "ReplicaSet should still exist")
	for i := 0; i < 3; i++ {
		assert.NotNil(rs.Replicas[i], "Replica %d should not be nil", i)
		assert.Equal(types.ContainerStatusRunning, rs.Replicas[i].Status,
			"Replica %d should be running", i)
	}

	// ---------------------------------------------------------------
	// Phase 8: Cleanup all resources
	// ---------------------------------------------------------------
	t.Log("Phase 8: Cleaning up all resources")

	for _, cID := range replicaContainerIDs {
		err := h.Containerd.StopContainer(ctx, cID, 5*time.Second)
		assert.NoError(err, "Stop %s should succeed", cID)

		err = h.Containerd.DeleteContainer(ctx, cID)
		assert.NoError(err, "Delete %s should succeed", cID)
	}

	replicator.DeleteReplicaSet(containerID)
	_, exists = replicator.GetReplicaSet(containerID)
	assert.False(exists, "ReplicaSet should be deleted")

	allContainers := h.Containerd.ListContainers()
	assert.Empty(allContainers, "All containers should be cleaned up")
}

// TestE2E_DeploymentWithEscrowPayment tests the deployment lifecycle with
// escrow-based payments: create escrow, run for a simulated duration,
// release payments progressively, and verify amounts are correct.
func TestE2E_DeploymentWithEscrowPayment(t *testing.T) {
	h := testutil.NewTestHarness(t)
	assert := testutil.NewAssertions(t)

	ctx, cancel := h.WithTimeout(60 * time.Second)
	defer cancel()

	containerID := "e2e-escrow-deploy"
	imageRef := "docker.io/library/nginx:latest"
	regions := []string{"Americas", "Europe", "Asia-Pacific"}

	// Set up payment infrastructure
	minStake := new(big.Int)
	minStake.SetString("500000000000000000000", 10) // 500 BUNKER
	stakingMgr := payment.NewStakingManager(nil, minStake)
	escrowMgr := payment.NewEscrowManager()
	replicator := redundancy.NewReplicator()
	healthMonitor := redundancy.NewHealthMonitor()

	// ---------------------------------------------------------------
	// Phase 1: Provider staking (3 providers)
	// ---------------------------------------------------------------
	t.Log("Phase 1: Setting up provider stakes")
	providers := [3]common.Address{
		common.HexToAddress("0x1111111111111111111111111111111111111111"),
		common.HexToAddress("0x2222222222222222222222222222222222222222"),
		common.HexToAddress("0x3333333333333333333333333333333333333333"),
	}

	providerStake := new(big.Int)
	providerStake.SetString("2000000000000000000000", 10) // 2000 BUNKER each

	for _, provider := range providers {
		err := stakingMgr.Stake(nil, provider, new(big.Int).Set(providerStake))
		assert.NoError(err, "Provider %s staking should succeed", provider.Hex())
		assert.True(stakingMgr.HasMinimumStake(provider),
			"Provider %s should have minimum stake", provider.Hex())
	}

	// ---------------------------------------------------------------
	// Phase 2: Create escrow for deployment
	// ---------------------------------------------------------------
	t.Log("Phase 2: Creating escrow for deployment")

	escrowAmount := new(big.Int)
	escrowAmount.SetString("300000000000000000000", 10) // 300 BUNKER for 3 replicas over 1 hour
	escrowDuration := time.Hour

	escrow := escrowMgr.CreateEscrow(containerID, escrowAmount, escrowDuration)
	assert.NotNil(escrow, "Escrow should be created")
	assert.Equal(containerID, escrow.ReservationID)

	// Verify escrow is retrievable
	retrieved, exists := escrowMgr.GetEscrow(containerID)
	assert.True(exists, "Escrow should be retrievable")
	assert.Equal(escrowAmount.String(), retrieved.Amount.String())

	// ---------------------------------------------------------------
	// Phase 3: Deploy replicas
	// ---------------------------------------------------------------
	t.Log("Phase 3: Deploying container replicas")

	replicaSet, err := replicator.CreateReplicaSet(containerID, regions)
	assert.NoError(err)
	assert.NotNil(replicaSet)

	for i := 0; i < 3; i++ {
		replicaID := fmt.Sprintf("%s-pay-replica-%d", containerID, i)

		_, err := h.Containerd.CreateContainer(ctx, replicaID, imageRef, types.ResourceLimits{
			MemoryLimit: 256 * 1024 * 1024,
		})
		assert.NoError(err, "Create replica %d should succeed", i)

		err = h.Containerd.StartContainer(ctx, replicaID)
		assert.NoError(err, "Start replica %d should succeed", i)

		container := &types.Container{
			ID:     replicaID,
			Status: types.ContainerStatusRunning,
		}
		err = replicator.AddReplica(containerID, i, container)
		assert.NoError(err)

		healthMonitor.UpdateHealth(containerID, i, types.HealthStatus{
			Healthy:    true,
			LastUpdate: time.Now(),
		})
	}

	// ---------------------------------------------------------------
	// Phase 4: Progressive payment releases (simulating uptime)
	// ---------------------------------------------------------------
	t.Log("Phase 4: Progressive escrow payment releases")

	totalReleased := big.NewInt(0)

	// Release at 25% uptime (15 minutes)
	released1, err := escrowMgr.ReleasePayment(containerID, 15*time.Minute)
	assert.NoError(err, "First release should succeed")
	assert.True(released1.Sign() > 0, "First release should be positive")
	totalReleased.Add(totalReleased, released1)
	t.Logf("  Released at 25%%: %s", released1.String())

	// Release at 50% uptime (30 minutes)
	released2, err := escrowMgr.ReleasePayment(containerID, 30*time.Minute)
	assert.NoError(err, "Second release should succeed")
	assert.True(released2.Sign() > 0, "Second release should be positive")
	totalReleased.Add(totalReleased, released2)
	t.Logf("  Released at 50%%: %s", released2.String())

	// Release at 75% uptime (45 minutes)
	released3, err := escrowMgr.ReleasePayment(containerID, 45*time.Minute)
	assert.NoError(err, "Third release should succeed")
	assert.True(released3.Sign() > 0, "Third release should be positive")
	totalReleased.Add(totalReleased, released3)
	t.Logf("  Released at 75%%: %s", released3.String())

	// Release at 100% uptime (60 minutes = full duration)
	released4, err := escrowMgr.ReleasePayment(containerID, escrowDuration)
	assert.NoError(err, "Final release should succeed")
	assert.True(released4.Sign() > 0, "Final release should be positive")
	totalReleased.Add(totalReleased, released4)
	t.Logf("  Released at 100%%: %s", released4.String())

	// Total released should equal the escrow amount
	assert.Equal(escrowAmount.String(), totalReleased.String(),
		"Total released should equal escrow amount")

	// ---------------------------------------------------------------
	// Phase 5: Verify no more payments after full release
	// ---------------------------------------------------------------
	t.Log("Phase 5: Verifying no further payments after full release")

	extraRelease, err := escrowMgr.ReleasePayment(containerID, 2*escrowDuration)
	assert.NoError(err, "Extra release call should not error")
	assert.Equal(big.NewInt(0).String(), extraRelease.String(),
		"No additional funds should be released")

	// ---------------------------------------------------------------
	// Phase 6: Cleanup
	// ---------------------------------------------------------------
	t.Log("Phase 6: Cleaning up")
	for i := 0; i < 3; i++ {
		replicaID := fmt.Sprintf("%s-pay-replica-%d", containerID, i)
		_ = h.Containerd.DeleteContainer(ctx, replicaID)
	}
	replicator.DeleteReplicaSet(containerID)
}

// TestE2E_DeploymentFailoverConsensus tests that when 1 of 3 replicas fails,
// the consensus mechanism correctly reports 2/3 running, and when the
// failed replica is replaced, consensus returns to 3/3.
func TestE2E_DeploymentFailoverConsensus(t *testing.T) {
	h := testutil.NewTestHarness(t)
	assert := testutil.NewAssertions(t)

	ctx, cancel := h.WithTimeout(60 * time.Second)
	defer cancel()

	containerID := "e2e-failover-consensus"
	imageRef := "docker.io/library/alpine:latest"
	regions := []string{"Americas", "Europe", "Asia-Pacific"}

	replicator := redundancy.NewReplicator()
	healthMonitor := redundancy.NewHealthMonitor()
	consensus := redundancy.NewConsensusManager()

	// ---------------------------------------------------------------
	// Phase 1: Set up 3 healthy replicas
	// ---------------------------------------------------------------
	t.Log("Phase 1: Setting up 3 healthy replicas")

	replicaSet, err := replicator.CreateReplicaSet(containerID, regions)
	assert.NoError(err)
	assert.NotNil(replicaSet)

	replicaContainerIDs := make([]string, 3)
	replicaContainers := [3]*types.Container{}

	for i := 0; i < 3; i++ {
		replicaID := fmt.Sprintf("%s-fc-replica-%d", containerID, i)
		replicaContainerIDs[i] = replicaID

		_, err := h.Containerd.CreateContainer(ctx, replicaID, imageRef, types.ResourceLimits{
			MemoryLimit: 128 * 1024 * 1024,
		})
		assert.NoError(err)

		err = h.Containerd.StartContainer(ctx, replicaID)
		assert.NoError(err)

		container := &types.Container{
			ID:     replicaID,
			Status: types.ContainerStatusRunning,
		}
		replicaContainers[i] = container
		err = replicator.AddReplica(containerID, i, container)
		assert.NoError(err)

		healthMonitor.UpdateHealth(containerID, i, types.HealthStatus{
			Healthy:     true,
			CPUUsage:    15.0,
			MemoryUsage: 64 * 1024 * 1024,
			LastUpdate:  time.Now(),
		})
	}

	// Verify initial consensus: 3/3 running
	consensus.UpdateState(containerID, types.ContainerStatusRunning, replicaContainers)
	status, err := consensus.GetConsensusStatus(containerID)
	assert.NoError(err)
	assert.Equal(types.ContainerStatusRunning, status, "Initial consensus: 3/3 running")

	// ---------------------------------------------------------------
	// Phase 2: Fail replica 2 (Asia-Pacific), verify 2/3 consensus
	// ---------------------------------------------------------------
	t.Log("Phase 2: Failing replica 2 (Asia-Pacific)")

	err = h.Containerd.StopContainer(ctx, replicaContainerIDs[2], 5*time.Second)
	assert.NoError(err)

	healthMonitor.UpdateHealth(containerID, 2, types.HealthStatus{
		Healthy:    false,
		LastUpdate: time.Now(),
	})

	replicaContainers[2] = &types.Container{
		ID:     replicaContainerIDs[2],
		Status: types.ContainerStatusFailed,
	}
	consensus.UpdateState(containerID, types.ContainerStatusRunning, replicaContainers)

	// 2/3 replicas running => consensus is still "running"
	status, err = consensus.GetConsensusStatus(containerID)
	assert.NoError(err)
	assert.Equal(types.ContainerStatusRunning, status,
		"2/3 consensus should still be running after 1 failure")

	// Confirm health monitor sees 1 unhealthy
	unhealthy := healthMonitor.GetUnhealthyReplicas(containerID)
	assert.Len(unhealthy, 1, "Should have 1 unhealthy replica")

	// ---------------------------------------------------------------
	// Phase 3: Fail another replica (0 - Americas), verify split
	// ---------------------------------------------------------------
	t.Log("Phase 3: Failing replica 0 (Americas) - now 1/3 running")

	err = h.Containerd.StopContainer(ctx, replicaContainerIDs[0], 5*time.Second)
	assert.NoError(err)

	healthMonitor.UpdateHealth(containerID, 0, types.HealthStatus{
		Healthy:    false,
		LastUpdate: time.Now(),
	})

	replicaContainers[0] = &types.Container{
		ID:     replicaContainerIDs[0],
		Status: types.ContainerStatusFailed,
	}
	consensus.UpdateState(containerID, types.ContainerStatusFailed, replicaContainers)

	// 2/3 failed => consensus is "failed"
	status, err = consensus.GetConsensusStatus(containerID)
	assert.NoError(err)
	assert.Equal(types.ContainerStatusFailed, status,
		"2/3 consensus should report failed when 2 replicas are down")

	unhealthy = healthMonitor.GetUnhealthyReplicas(containerID)
	assert.Len(unhealthy, 2, "Should have 2 unhealthy replicas")

	// ---------------------------------------------------------------
	// Phase 4: Replace both failed replicas, restore 3/3
	// ---------------------------------------------------------------
	t.Log("Phase 4: Replacing failed replicas to restore 3/3")

	for _, failedIdx := range []int{0, 2} {
		err := h.Containerd.DeleteContainer(ctx, replicaContainerIDs[failedIdx])
		assert.NoError(err)

		replacementID := fmt.Sprintf("%s-fc-replacement-%d", containerID, failedIdx)
		replicaContainerIDs[failedIdx] = replacementID

		_, err = h.Containerd.CreateContainer(ctx, replacementID, imageRef, types.ResourceLimits{
			MemoryLimit: 128 * 1024 * 1024,
		})
		assert.NoError(err)

		err = h.Containerd.StartContainer(ctx, replacementID)
		assert.NoError(err)

		replacement := &types.Container{
			ID:     replacementID,
			Status: types.ContainerStatusRunning,
		}
		replicaContainers[failedIdx] = replacement

		err = replicator.ReplaceReplica(nil, containerID, failedIdx, replacement)
		assert.NoError(err)

		healthMonitor.UpdateHealth(containerID, failedIdx, types.HealthStatus{
			Healthy:     true,
			CPUUsage:    18.0,
			MemoryUsage: 70 * 1024 * 1024,
			LastUpdate:  time.Now(),
		})
	}

	// Update consensus with all running
	consensus.UpdateState(containerID, types.ContainerStatusRunning, replicaContainers)

	status, err = consensus.GetConsensusStatus(containerID)
	assert.NoError(err)
	assert.Equal(types.ContainerStatusRunning, status,
		"Consensus should be running after replacing all failed replicas")

	unhealthy = healthMonitor.GetUnhealthyReplicas(containerID)
	assert.Len(unhealthy, 0, "No replicas should be unhealthy after full recovery")

	// ---------------------------------------------------------------
	// Phase 5: Verify gossip merge with stale data does not revert status
	// ---------------------------------------------------------------
	t.Log("Phase 5: Verifying gossip merge with stale version")

	state, exists := consensus.GetState(containerID)
	assert.True(exists, "State should exist")
	currentVersion := state.Version

	// Simulate receiving stale gossip (lower version) with failed status
	staleState := &redundancy.ContainerState{
		ContainerID: containerID,
		Status:      types.ContainerStatusFailed,
		Replicas:    [3]*types.Container{},
		Version:     currentVersion - 1, // stale
	}
	consensus.MergeState(containerID, staleState)

	// Status should remain running because stale version is ignored
	status, err = consensus.GetConsensusStatus(containerID)
	assert.NoError(err)
	assert.Equal(types.ContainerStatusRunning, status,
		"Stale gossip should not override current consensus")

	// ---------------------------------------------------------------
	// Phase 6: Cleanup
	// ---------------------------------------------------------------
	t.Log("Phase 6: Cleanup")
	for _, cID := range replicaContainerIDs {
		_ = h.Containerd.StopContainer(ctx, cID, 5*time.Second)
		_ = h.Containerd.DeleteContainer(ctx, cID)
	}
	replicator.DeleteReplicaSet(containerID)
}

// TestE2E_DeploymentCleanup verifies that all resources are properly cleaned up
// after a deployment is torn down: containers, escrow, health monitors,
// replica sets, and consensus state.
func TestE2E_DeploymentCleanup(t *testing.T) {
	h := testutil.NewTestHarness(t)
	assert := testutil.NewAssertions(t)

	ctx, cancel := h.WithTimeout(60 * time.Second)
	defer cancel()

	containerID := "e2e-cleanup-lifecycle"
	imageRef := "docker.io/library/redis:latest"
	regions := []string{"Americas", "Europe", "Asia-Pacific"}

	replicator := redundancy.NewReplicator()
	healthMonitor := redundancy.NewHealthMonitor()
	consensus := redundancy.NewConsensusManager()
	escrowMgr := payment.NewEscrowManager()

	minStake := new(big.Int)
	minStake.SetString("500000000000000000000", 10)
	stakingMgr := payment.NewStakingManager(nil, minStake)

	// ---------------------------------------------------------------
	// Phase 1: Create full deployment with all subsystems
	// ---------------------------------------------------------------
	t.Log("Phase 1: Setting up full deployment with all subsystems")

	// Stake providers
	provider := common.HexToAddress("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	stakeAmount := new(big.Int)
	stakeAmount.SetString("5000000000000000000000", 10) // 5000 BUNKER
	err := stakingMgr.Stake(nil, provider, stakeAmount)
	assert.NoError(err)

	// Create escrow
	escrowAmount := new(big.Int)
	escrowAmount.SetString("100000000000000000000", 10) // 100 BUNKER
	escrow := escrowMgr.CreateEscrow(containerID, escrowAmount, time.Hour)
	assert.NotNil(escrow)

	// Create replica set
	replicaSet, err := replicator.CreateReplicaSet(containerID, regions)
	assert.NoError(err)
	assert.NotNil(replicaSet)

	// Deploy and register replicas
	replicaContainerIDs := make([]string, 3)
	replicaContainers := [3]*types.Container{}

	for i := 0; i < 3; i++ {
		replicaID := fmt.Sprintf("%s-cleanup-replica-%d", containerID, i)
		replicaContainerIDs[i] = replicaID

		_, err := h.Containerd.CreateContainer(ctx, replicaID, imageRef, types.ResourceLimits{
			MemoryLimit: 128 * 1024 * 1024,
		})
		assert.NoError(err)

		err = h.Containerd.StartContainer(ctx, replicaID)
		assert.NoError(err)

		container := &types.Container{
			ID:     replicaID,
			Status: types.ContainerStatusRunning,
		}
		replicaContainers[i] = container
		err = replicator.AddReplica(containerID, i, container)
		assert.NoError(err)

		healthMonitor.UpdateHealth(containerID, i, types.HealthStatus{
			Healthy:     true,
			CPUUsage:    10.0,
			MemoryUsage: 50 * 1024 * 1024,
			LastUpdate:  time.Now(),
		})
	}

	consensus.UpdateState(containerID, types.ContainerStatusRunning, replicaContainers)

	// ---------------------------------------------------------------
	// Phase 2: Verify everything is set up
	// ---------------------------------------------------------------
	t.Log("Phase 2: Verifying all subsystems are active")

	assert.Len(h.Containerd.ListContainers(), 3, "Should have 3 containers")

	_, exists := replicator.GetReplicaSet(containerID)
	assert.True(exists, "ReplicaSet should exist")

	for i := 0; i < 3; i++ {
		assert.True(healthMonitor.IsHealthy(containerID, i),
			"Replica %d should be healthy", i)
	}

	status, err := consensus.GetConsensusStatus(containerID)
	assert.NoError(err)
	assert.Equal(types.ContainerStatusRunning, status)

	_, escrowExists := escrowMgr.GetEscrow(containerID)
	assert.True(escrowExists, "Escrow should exist")

	assert.True(stakingMgr.HasMinimumStake(provider), "Provider should have stake")

	// ---------------------------------------------------------------
	// Phase 3: Release partial escrow before cleanup
	// ---------------------------------------------------------------
	t.Log("Phase 3: Releasing partial escrow payment")

	released, err := escrowMgr.ReleasePayment(containerID, 30*time.Minute)
	assert.NoError(err)
	assert.True(released.Sign() > 0, "Partial release should be positive")

	// ---------------------------------------------------------------
	// Phase 4: Tear down - stop and delete all containers
	// ---------------------------------------------------------------
	t.Log("Phase 4: Tearing down deployment")

	for _, cID := range replicaContainerIDs {
		err := h.Containerd.StopContainer(ctx, cID, 5*time.Second)
		assert.NoError(err, "Stop %s should succeed", cID)

		status, err := h.Containerd.GetContainerStatus(ctx, cID)
		assert.NoError(err)
		assert.Equal(types.ContainerStatusStopped, status)

		err = h.Containerd.DeleteContainer(ctx, cID)
		assert.NoError(err, "Delete %s should succeed", cID)
	}

	// ---------------------------------------------------------------
	// Phase 5: Clean up replica set and verify
	// ---------------------------------------------------------------
	t.Log("Phase 5: Cleaning up replica set")

	replicator.DeleteReplicaSet(containerID)
	_, exists = replicator.GetReplicaSet(containerID)
	assert.False(exists, "ReplicaSet should be deleted")

	// ---------------------------------------------------------------
	// Phase 6: Verify all mock containers are gone
	// ---------------------------------------------------------------
	t.Log("Phase 6: Verifying all containers cleaned up")

	allContainers := h.Containerd.ListContainers()
	assert.Empty(allContainers, "All containers should be deleted")

	for _, cID := range replicaContainerIDs {
		_, containerExists := h.Containerd.GetContainer(cID)
		assert.False(containerExists, "Container %s should not exist", cID)

		_, err := h.Containerd.GetContainerStatus(ctx, cID)
		assert.Error(err, "GetContainerStatus for %s should error", cID)
	}

	// ---------------------------------------------------------------
	// Phase 7: Verify health reports reflect cleanup
	// ---------------------------------------------------------------
	t.Log("Phase 7: Updating health monitors to reflect teardown")

	// Mark all replicas as unhealthy (simulating cleanup notification)
	for i := 0; i < 3; i++ {
		healthMonitor.UpdateHealth(containerID, i, types.HealthStatus{
			Healthy:    false,
			LastUpdate: time.Now(),
		})
	}

	unhealthy := healthMonitor.GetUnhealthyReplicas(containerID)
	assert.Len(unhealthy, 3, "All 3 replicas should be reported unhealthy after teardown")

	// ---------------------------------------------------------------
	// Phase 8: Verify escrow final state
	// ---------------------------------------------------------------
	t.Log("Phase 8: Verifying escrow final state")

	// Release remaining escrow (for accounting)
	finalRelease, err := escrowMgr.ReleasePayment(containerID, time.Hour)
	assert.NoError(err)
	// Some amount should remain since we only released 30 minutes
	assert.True(finalRelease.Sign() > 0,
		"Remaining escrow should be released on full duration")

	// After full release, no more funds
	zeroRelease, err := escrowMgr.ReleasePayment(containerID, 2*time.Hour)
	assert.NoError(err)
	assert.Equal(big.NewInt(0).String(), zeroRelease.String(),
		"No funds should remain after full release")

	// Verify all replica sets are gone
	allSets := replicator.GetAllReplicaSets()
	assert.Len(allSets, 0, "All replica sets should be cleaned up")

	t.Log("Deployment cleanup verification complete")
}
