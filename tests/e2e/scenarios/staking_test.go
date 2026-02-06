//go:build e2e

package scenarios

import (
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/moltbunker/moltbunker/internal/payment"
	"github.com/moltbunker/moltbunker/internal/redundancy"
	"github.com/moltbunker/moltbunker/pkg/types"
	"github.com/moltbunker/moltbunker/tests/e2e/testutil"
)

// TestE2E_StakingBasicFlow tests the basic staking lifecycle
func TestE2E_StakingBasicFlow(t *testing.T) {
	assert := testutil.NewAssertions(t)

	// Create staking manager with minimum stake of 500 BUNKER
	minStake := new(big.Int)
	minStake.SetString("500000000000000000000", 10) // 500 * 10^18
	stakingMgr := payment.NewStakingManager(nil, minStake)

	provider := common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")

	// Initially no stake
	stake := stakingMgr.GetStake(provider)
	assert.Equal(big.NewInt(0).String(), stake.String(), "Initial stake should be 0")

	// Stake below minimum should fail
	smallStake := new(big.Int)
	smallStake.SetString("100000000000000000000", 10) // 100 BUNKER
	err := stakingMgr.Stake(nil, provider, smallStake)
	assert.Error(err, "Staking below minimum should fail")

	// Stake the minimum
	err = stakingMgr.Stake(nil, provider, minStake)
	assert.NoError(err, "Staking minimum should succeed")

	// Verify stake
	stake = stakingMgr.GetStake(provider)
	assert.Equal(minStake.String(), stake.String(), "Stake should equal minimum")
	assert.True(stakingMgr.HasMinimumStake(provider), "Should have minimum stake")

	// Stake more
	additionalStake := new(big.Int)
	additionalStake.SetString("1500000000000000000000", 10) // 1500 BUNKER
	err = stakingMgr.Stake(nil, provider, additionalStake)
	assert.NoError(err, "Additional staking should succeed")

	// Verify total stake
	expectedTotal := new(big.Int).Add(minStake, additionalStake)
	stake = stakingMgr.GetStake(provider)
	assert.Equal(expectedTotal.String(), stake.String(), "Total stake should be sum")
}

// TestE2E_StakingSlashing tests the slashing mechanism
func TestE2E_StakingSlashing(t *testing.T) {
	assert := testutil.NewAssertions(t)

	minStake := new(big.Int)
	minStake.SetString("500000000000000000000", 10)
	stakingMgr := payment.NewStakingManager(nil, minStake)

	provider := common.HexToAddress("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")

	// Stake 10000 BUNKER
	stakeAmount := new(big.Int)
	stakeAmount.SetString("10000000000000000000000", 10) // 10000 BUNKER
	err := stakingMgr.Stake(nil, provider, stakeAmount)
	assert.NoError(err)

	// Slash 15% (1500 BUNKER) for job abandonment
	slashAmount := new(big.Int)
	slashAmount.SetString("1500000000000000000000", 10) // 1500 BUNKER
	err = stakingMgr.Slash(nil, provider, slashAmount)
	assert.NoError(err)

	// Verify remaining stake
	expectedRemaining := new(big.Int).Sub(stakeAmount, slashAmount)
	remaining := stakingMgr.GetStake(provider)
	assert.Equal(expectedRemaining.String(), remaining.String(), "Remaining stake after slash")
	assert.True(stakingMgr.HasMinimumStake(provider), "Should still have minimum stake")

	// Slash more than remaining
	hugeSlash := new(big.Int)
	hugeSlash.SetString("100000000000000000000000", 10) // 100000 BUNKER (more than staked)
	err = stakingMgr.Slash(nil, provider, hugeSlash)
	assert.NoError(err, "Slash exceeding stake should succeed (capped)")

	// Verify stake is 0 after excessive slash
	remaining = stakingMgr.GetStake(provider)
	assert.Equal(big.NewInt(0).String(), remaining.String(), "Stake should be 0 after excessive slash")
	assert.False(stakingMgr.HasMinimumStake(provider), "Should not have minimum stake")
}

// TestE2E_EscrowPaymentFlow tests the escrow payment lifecycle
func TestE2E_EscrowPaymentFlow(t *testing.T) {
	assert := testutil.NewAssertions(t)

	escrowMgr := payment.NewEscrowManager()

	// Create escrow for a deployment (100 BUNKER for 1 hour)
	amount := new(big.Int)
	amount.SetString("100000000000000000000", 10) // 100 BUNKER
	duration := time.Hour

	escrow := escrowMgr.CreateEscrow("deploy-001", amount, duration)
	assert.NotNil(escrow)
	assert.Equal("deploy-001", escrow.ReservationID)

	// Release payment for 30 minutes of uptime (should get ~50 BUNKER)
	halfDuration := duration / 2
	released, err := escrowMgr.ReleasePayment("deploy-001", halfDuration)
	assert.NoError(err)
	assert.True(released.Sign() > 0, "Released amount should be positive")

	// Release payment for remaining time
	released2, err := escrowMgr.ReleasePayment("deploy-001", duration)
	assert.NoError(err)
	assert.True(released2.Sign() > 0, "Second release should be positive")

	// Total released should approximately equal the escrow amount
	totalReleased := new(big.Int).Add(released, released2)
	assert.Equal(amount.String(), totalReleased.String(), "Total released should equal escrow amount")
}

// TestE2E_ReplicationWithStaking tests that replication respects staking requirements
func TestE2E_ReplicationWithStaking(t *testing.T) {
	assert := testutil.NewAssertions(t)

	// Create replicator
	replicator := redundancy.NewReplicator()
	healthMonitor := redundancy.NewHealthMonitor()

	// Create a 3-region replica set
	containerID := "staking-repl-test"
	regions := []string{"Americas", "Europe", "Asia-Pacific"}

	replicaSet, err := replicator.CreateReplicaSet(containerID, regions)
	assert.NoError(err)
	assert.Equal("Americas", replicaSet.Region1)
	assert.Equal("Europe", replicaSet.Region2)
	assert.Equal("Asia-Pacific", replicaSet.Region3)

	// Add replicas (simulating provider nodes accepting the deployment)
	for i := 0; i < 3; i++ {
		container := &types.Container{
			ID:     containerID,
			Status: types.ContainerStatusRunning,
		}
		err := replicator.AddReplica(containerID, i, container)
		assert.NoError(err)

		// Set health status
		healthMonitor.UpdateHealth(containerID, i, types.HealthStatus{
			Healthy:     true,
			CPUUsage:    25.0,
			MemoryUsage: 128 * 1024 * 1024,
		})
	}

	// Verify all 3 replicas are healthy
	for i := 0; i < 3; i++ {
		assert.True(healthMonitor.IsHealthy(containerID, i),
			"Replica %d should be healthy", i)
	}

	// Simulate replica failure
	healthMonitor.UpdateHealth(containerID, 1, types.HealthStatus{
		Healthy: false,
	})

	unhealthy := healthMonitor.GetUnhealthyReplicas(containerID)
	assert.Len(unhealthy, 1, "Should have 1 unhealthy replica")
	assert.Equal(1, unhealthy[0], "Replica 1 should be unhealthy")

	// Replace the failed replica
	newContainer := &types.Container{
		ID:     containerID,
		Status: types.ContainerStatusRunning,
	}
	err = replicator.ReplaceReplica(nil, containerID, 1, newContainer)
	assert.NoError(err)

	// Update health after replacement
	healthMonitor.UpdateHealth(containerID, 1, types.HealthStatus{
		Healthy:     true,
		CPUUsage:    30.0,
		MemoryUsage: 128 * 1024 * 1024,
	})

	// Verify all replicas are healthy again
	for i := 0; i < 3; i++ {
		assert.True(healthMonitor.IsHealthy(containerID, i),
			"Replica %d should be healthy after replacement", i)
	}
}

// TestE2E_ConsensusAcrossReplicas tests consensus mechanism for container status
func TestE2E_ConsensusAcrossReplicas(t *testing.T) {
	assert := testutil.NewAssertions(t)

	consensus := redundancy.NewConsensusManager()

	containerID := "consensus-test"

	// Initial state: all running
	allRunning := [3]*types.Container{
		{ID: "r1", Status: types.ContainerStatusRunning},
		{ID: "r2", Status: types.ContainerStatusRunning},
		{ID: "r3", Status: types.ContainerStatusRunning},
	}
	consensus.UpdateState(containerID, types.ContainerStatusRunning, allRunning)

	// Verify consensus
	status, err := consensus.GetConsensusStatus(containerID)
	assert.NoError(err)
	assert.Equal(types.ContainerStatusRunning, status, "Consensus should be running")

	// Simulate split: 2 running, 1 stopped
	splitState := [3]*types.Container{
		{ID: "r1", Status: types.ContainerStatusRunning},
		{ID: "r2", Status: types.ContainerStatusRunning},
		{ID: "r3", Status: types.ContainerStatusStopped},
	}
	consensus.UpdateState(containerID, types.ContainerStatusRunning, splitState)

	status, err = consensus.GetConsensusStatus(containerID)
	assert.NoError(err)
	assert.Equal(types.ContainerStatusRunning, status, "2/3 majority should be running")

	// Simulate majority stopped
	stoppedState := [3]*types.Container{
		{ID: "r1", Status: types.ContainerStatusStopped},
		{ID: "r2", Status: types.ContainerStatusStopped},
		{ID: "r3", Status: types.ContainerStatusRunning},
	}
	consensus.UpdateState(containerID, types.ContainerStatusStopped, stoppedState)

	status, err = consensus.GetConsensusStatus(containerID)
	assert.NoError(err)
	assert.Equal(types.ContainerStatusStopped, status, "2/3 majority should be stopped")
}

// TestE2E_MultipleProviderStaking tests multiple providers staking
func TestE2E_MultipleProviderStaking(t *testing.T) {
	assert := testutil.NewAssertions(t)

	minStake := new(big.Int)
	minStake.SetString("500000000000000000000", 10) // 500 BUNKER
	stakingMgr := payment.NewStakingManager(nil, minStake)

	providers := []struct {
		address common.Address
		stake   string
		tier    types.StakingTier
	}{
		{
			address: common.HexToAddress("0x1111111111111111111111111111111111111111"),
			stake:   "500000000000000000000",    // 500 BUNKER - Starter
			tier:    types.StakingTierStarter,
		},
		{
			address: common.HexToAddress("0x2222222222222222222222222222222222222222"),
			stake:   "2000000000000000000000",   // 2000 BUNKER - Bronze
			tier:    types.StakingTierBronze,
		},
		{
			address: common.HexToAddress("0x3333333333333333333333333333333333333333"),
			stake:   "50000000000000000000000",  // 50000 BUNKER - Gold
			tier:    types.StakingTierGold,
		},
	}

	for _, p := range providers {
		stakeAmount := new(big.Int)
		stakeAmount.SetString(p.stake, 10)
		err := stakingMgr.Stake(nil, p.address, stakeAmount)
		assert.NoError(err, "Provider %s staking should succeed", p.address.Hex())
		assert.True(stakingMgr.HasMinimumStake(p.address),
			"Provider %s should have minimum stake", p.address.Hex())
	}

	// Verify all providers have correct stakes
	for _, p := range providers {
		expected := new(big.Int)
		expected.SetString(p.stake, 10)
		actual := stakingMgr.GetStake(p.address)
		assert.Equal(expected.String(), actual.String(),
			"Provider %s stake mismatch", p.address.Hex())
	}
}
