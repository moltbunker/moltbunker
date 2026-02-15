//go:build e2e

package scenarios

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/moltbunker/moltbunker/internal/payment"
	"github.com/moltbunker/moltbunker/pkg/types"
	"github.com/moltbunker/moltbunker/tests/e2e/testutil"
)

// bunker converts a human-readable BUNKER token amount to wei (18 decimals).
func bunker(amount string) *big.Int {
	tokens, ok := new(big.Int).SetString(amount, 10)
	if !ok {
		panic("invalid amount: " + amount)
	}
	decimals := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	return new(big.Int).Mul(tokens, decimals)
}

// TestE2E_StakingTierProgression tests incremental staking through all 5 tiers,
// verifying the correct tier assignment at each level.
// Tier thresholds (StakingManager via DefaultStakingTiers()):
//
//	Starter:  500 BUNKER
//	Bronze:   2,000 BUNKER
//	Silver:   10,000 BUNKER
//	Gold:     50,000 BUNKER
//	Platinum: 250,000 BUNKER
func TestE2E_StakingTierProgression(t *testing.T) {
	assert := testutil.NewAssertions(t)

	minStake := bunker("500")
	stakingMgr := payment.NewStakingManager(nil, minStake)

	provider := common.HexToAddress("0xAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")

	// Verify no tier before staking
	tier := stakingMgr.GetTier(provider)
	assert.Equal(types.StakingTier(""), tier, "Should have no tier before staking")

	// Step 1: Stake 500 BUNKER -> Starter tier
	err := stakingMgr.Stake(nil, provider, bunker("500"))
	assert.NoError(err, "Stake 500 should succeed")
	assert.Equal(types.StakingTierStarter, stakingMgr.GetTier(provider), "500 BUNKER should be Starter tier")
	assert.Equal(bunker("500").String(), stakingMgr.GetStake(provider).String(), "Stake should be 500")

	// Step 2: Add 1500 BUNKER (total: 2000) -> Bronze tier
	err = stakingMgr.Stake(nil, provider, bunker("1500"))
	assert.NoError(err, "Additional stake of 1500 should succeed")
	assert.Equal(types.StakingTierBronze, stakingMgr.GetTier(provider), "2000 BUNKER should be Bronze tier")
	assert.Equal(bunker("2000").String(), stakingMgr.GetStake(provider).String(), "Stake should be 2000")

	// Step 3: Add 8000 BUNKER (total: 10000) -> Silver tier
	err = stakingMgr.Stake(nil, provider, bunker("8000"))
	assert.NoError(err, "Additional stake of 8000 should succeed")
	assert.Equal(types.StakingTierSilver, stakingMgr.GetTier(provider), "10000 BUNKER should be Silver tier")
	assert.Equal(bunker("10000").String(), stakingMgr.GetStake(provider).String(), "Stake should be 10000")

	// Step 4: Add 40000 BUNKER (total: 50000) -> Gold tier
	err = stakingMgr.Stake(nil, provider, bunker("40000"))
	assert.NoError(err, "Additional stake of 40000 should succeed")
	assert.Equal(types.StakingTierGold, stakingMgr.GetTier(provider), "50000 BUNKER should be Gold tier")
	assert.Equal(bunker("50000").String(), stakingMgr.GetStake(provider).String(), "Stake should be 50000")

	// Step 5: Add 200000 BUNKER (total: 250000) -> Platinum tier
	err = stakingMgr.Stake(nil, provider, bunker("200000"))
	assert.NoError(err, "Additional stake of 200000 should succeed")
	assert.Equal(types.StakingTierPlatinum, stakingMgr.GetTier(provider), "250000 BUNKER should be Platinum tier")
	assert.Equal(bunker("250000").String(), stakingMgr.GetStake(provider).String(), "Stake should be 250000")

	// Verify tier config properties via DefaultStakingTiers
	tiers := types.DefaultStakingTiers()

	// Starter: no priority queue, no governance
	starterCfg := tiers[types.StakingTierStarter]
	assert.False(starterCfg.PriorityQueue, "Starter should not have priority queue")
	assert.False(starterCfg.Governance, "Starter should not have governance")
	assert.Equal(3, starterCfg.MaxConcurrentJobs, "Starter should have 3 max concurrent jobs")

	// Platinum: has everything
	platinumCfg := tiers[types.StakingTierPlatinum]
	assert.True(platinumCfg.PriorityQueue, "Platinum should have priority queue")
	assert.True(platinumCfg.Governance, "Platinum should have governance")
	assert.True(platinumCfg.RevenueShare, "Platinum should have revenue share")
}

// TestE2E_StakingUnstakeWithCooldown tests the full unstake lifecycle:
// stake -> request unstake -> verify cooldown enforced -> advance time -> withdraw.
func TestE2E_StakingUnstakeWithCooldown(t *testing.T) {
	assert := testutil.NewAssertions(t)

	minStake := bunker("500")
	stakingMgr := payment.NewStakingManager(nil, minStake)

	provider := common.HexToAddress("0xBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB")
	ctx := context.Background()

	// Stake 10000 BUNKER (Silver tier)
	err := stakingMgr.Stake(ctx, provider, bunker("10000"))
	assert.NoError(err, "Staking 10000 should succeed")
	assert.Equal(types.StakingTierSilver, stakingMgr.GetTier(provider), "Should be Silver tier")

	// Set a controllable time for the staking manager
	now := time.Now()
	stakingMgr.SetNowFunc(func() time.Time { return now })

	// Request unstake of 5000 BUNKER
	err = stakingMgr.Unstake(ctx, provider, bunker("5000"))
	assert.NoError(err, "Unstake request should succeed")

	// Stake should still be 10000 during cooldown (tokens locked)
	assert.Equal(bunker("10000").String(), stakingMgr.GetStake(provider).String(),
		"Stake should remain 10000 during cooldown")

	// Attempting to complete unstake before cooldown should fail
	_, err = stakingMgr.CompleteUnstake(ctx, provider)
	assert.Error(err, "CompleteUnstake before cooldown should fail")
	assert.True(err != nil, "Error should be non-nil")

	// Attempting another unstake request during cooldown should fail
	err = stakingMgr.Unstake(ctx, provider, bunker("1000"))
	assert.Error(err, "Second unstake during cooldown should fail")

	// Advance time by 6 days - still within cooldown (7 days)
	now = now.Add(6 * 24 * time.Hour)
	_, err = stakingMgr.CompleteUnstake(ctx, provider)
	assert.Error(err, "CompleteUnstake after 6 days should still fail")

	// Advance time past cooldown (total 7 days + 1 minute)
	now = now.Add(24*time.Hour + time.Minute)
	withdrawn, err := stakingMgr.CompleteUnstake(ctx, provider)
	assert.NoError(err, "CompleteUnstake after cooldown should succeed")
	assert.Equal(bunker("5000").String(), withdrawn.String(), "Should withdraw 5000 BUNKER")

	// Remaining stake should be 5000
	remaining := stakingMgr.GetStake(provider)
	assert.Equal(bunker("5000").String(), remaining.String(), "Remaining stake should be 5000")

	// Provider should now be at Bronze tier (5000 >= 2000 threshold)
	assert.Equal(types.StakingTierBronze, stakingMgr.GetTier(provider),
		"After partial unstake, should be Bronze tier (5000 BUNKER)")

	// Verify no more pending unstake
	_, err = stakingMgr.CompleteUnstake(ctx, provider)
	assert.Error(err, "No pending unstake should mean error")

	// Request unstake of remaining 5000 (unstake everything)
	err = stakingMgr.Unstake(ctx, provider, bunker("5000"))
	assert.NoError(err, "Unstake of remaining should succeed")

	// Advance past cooldown
	now = now.Add(8 * 24 * time.Hour)
	withdrawn, err = stakingMgr.CompleteUnstake(ctx, provider)
	assert.NoError(err, "CompleteUnstake should succeed")
	assert.Equal(bunker("5000").String(), withdrawn.String(), "Should withdraw remaining 5000")

	// Stake should now be zero
	assert.Equal(big.NewInt(0).String(), stakingMgr.GetStake(provider).String(),
		"Stake should be 0 after full withdrawal")
	assert.False(stakingMgr.HasMinimumStake(provider), "Should not have minimum stake")
	assert.Equal(types.StakingTier(""), stakingMgr.GetTier(provider), "Should have no tier")
}

// TestE2E_StakingContractIntegration tests the MockStakingContract for the full
// stake -> get info -> unstake -> withdraw flow at the smart contract level.
func TestE2E_StakingContractIntegration(t *testing.T) {
	assert := testutil.NewAssertions(t)

	sc := payment.NewMockStakingContract()
	ctx := context.Background()
	zeroAddr := common.Address{} // Mock contract uses zero address when baseClient is nil

	// Initially no stake
	stake, err := sc.GetStake(ctx, zeroAddr)
	assert.NoError(err, "GetStake should succeed")
	assert.Equal(big.NewInt(0).String(), stake.String(), "Initial stake should be 0")

	// No tier for zero stake
	tier, err := sc.GetTier(ctx, zeroAddr)
	assert.NoError(err, "GetTier should succeed")
	assert.Equal(types.StakingTier(""), tier, "Should have no tier")

	// Stake 500 BUNKER -> Starter
	_, err = sc.Stake(ctx, bunker("500"))
	assert.NoError(err, "Stake 500 should succeed")

	tier, err = sc.GetTier(ctx, zeroAddr)
	assert.NoError(err)
	assert.Equal(types.StakingTierStarter, tier, "500 BUNKER should be Starter tier")

	// Stake additional 1500 BUNKER -> Bronze (total 2000)
	_, err = sc.Stake(ctx, bunker("1500"))
	assert.NoError(err, "Additional stake should succeed")

	tier, err = sc.GetTier(ctx, zeroAddr)
	assert.NoError(err)
	assert.Equal(types.StakingTierBronze, tier, "2000 BUNKER should be Bronze tier")

	// Verify total staked
	total, err := sc.GetTotalStaked(ctx)
	assert.NoError(err)
	assert.Equal(bunker("2000").String(), total.String(), "Total staked should be 2000")

	// Get full stake info
	info, err := sc.GetStakeInfo(ctx, zeroAddr)
	assert.NoError(err, "GetStakeInfo should succeed")
	assert.Equal(bunker("2000").String(), info.StakedAmount.String(), "StakeInfo.StakedAmount should be 2000")
	assert.Equal(big.NewInt(0).String(), info.PendingUnstake.String(), "No pending unstake yet")
	assert.Equal(types.StakingTierBronze, info.Tier, "StakeInfo tier should be Bronze")

	// Request unstake of 1000 BUNKER
	_, err = sc.RequestUnstake(ctx, bunker("1000"))
	assert.NoError(err, "RequestUnstake should succeed")

	// Verify pending unstake via GetStakeInfo
	info, err = sc.GetStakeInfo(ctx, zeroAddr)
	assert.NoError(err)
	assert.Equal(bunker("1000").String(), info.StakedAmount.String(),
		"StakeInfo.StakedAmount should be 1000 after unstake request (mock subtracts immediately)")
	assert.Equal(bunker("1000").String(), info.PendingUnstake.String(),
		"PendingUnstake should be 1000")
	assert.True(info.UnlockTime.After(time.Now()), "Unlock time should be in the future")

	// Attempt withdraw before cooldown (should fail)
	_, err = sc.Withdraw(ctx)
	assert.Error(err, "Withdraw before cooldown should fail")

	// Verify stake amount did not change from failed withdraw
	stake, err = sc.GetStake(ctx, zeroAddr)
	assert.NoError(err)
	assert.Equal(bunker("1000").String(), stake.String(), "Stake should remain 1000 after failed withdraw")

	// Check has minimum stake (1000 BUNKER, contract minimum is 500)
	hasMin, err := sc.HasMinimumStake(ctx, zeroAddr)
	assert.NoError(err)
	assert.True(hasMin, "1000 BUNKER should meet contract minimum of 500")

	// Verify mock mode
	assert.True(sc.IsMockMode(), "Should be in mock mode")
}

// TestE2E_StakingSlashingPenalties tests slashing for different violation types
// and verifies the correct penalty percentages are applied.
func TestE2E_StakingSlashingPenalties(t *testing.T) {
	_ = testutil.NewAssertions(t)

	minStake := bunker("500")
	ctx := context.Background()

	// Violation types and their expected percentages from SlashingContract.CalculateSlashAmount:
	//   Downtime:           5%
	//   SLAViolation:       10%
	//   DataLoss:           25%
	//   SecurityBreach:     50%
	//   MaliciousBehavior:  100%
	violations := []struct {
		name            string
		reason          payment.ViolationReason
		expectedPercent int64
	}{
		{"Downtime", payment.ViolationDowntime, 5},
		{"SLAViolation", payment.ViolationSLAViolation, 10},
		{"JobAbandonment", payment.ViolationJobAbandonment, 15},
		{"SecurityViolation", payment.ViolationSecurityViolation, 50},
		{"Fraud", payment.ViolationFraud, 100},
		{"DataLoss", payment.ViolationDataLoss, 25},
		{"None", payment.ViolationNone, 0},
	}

	slashingContract := payment.NewMockSlashingContract()

	for _, v := range violations {
		t.Run(v.name, func(t *testing.T) {
			assert := testutil.NewAssertions(t)

			// Each provider starts with 10000 BUNKER
			stakeAmount := bunker("10000")

			// Calculate expected slash amount
			expectedSlash := new(big.Int).Mul(stakeAmount, big.NewInt(v.expectedPercent))
			expectedSlash.Div(expectedSlash, big.NewInt(100))

			// Verify CalculateSlashAmount returns correct amount
			actualSlash := slashingContract.CalculateSlashAmount(stakeAmount, v.reason)
			assert.Equal(expectedSlash.String(), actualSlash.String(),
				"CalculateSlashAmount for %s should be %d%%", v.name, v.expectedPercent)

			// Also test with StakingManager: stake then slash the calculated amount
			provider := common.HexToAddress("0xCC" + v.name + "CCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCC")
			// Use a unique address per test (pad/truncate to 20 bytes)
			addrBytes := make([]byte, 20)
			copy(addrBytes, []byte(v.name))
			provider = common.BytesToAddress(addrBytes)

			stakingMgr := payment.NewStakingManager(nil, minStake)
			err := stakingMgr.Stake(ctx, provider, stakeAmount)
			assert.NoError(err, "Staking should succeed for %s test", v.name)

			if actualSlash.Sign() > 0 {
				err = stakingMgr.Slash(ctx, provider, actualSlash)
				assert.NoError(err, "Slashing should succeed for %s", v.name)

				expectedRemaining := new(big.Int).Sub(stakeAmount, actualSlash)
				remaining := stakingMgr.GetStake(provider)
				assert.Equal(expectedRemaining.String(), remaining.String(),
					"Remaining stake after %s slash should be correct", v.name)
			}
		})
	}

	// Test cumulative slashing: a provider gets hit by multiple violations
	t.Run("CumulativeSlashing", func(t *testing.T) {
		assert := testutil.NewAssertions(t)

		provider := common.HexToAddress("0xDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDD")
		stakingMgr := payment.NewStakingManager(nil, minStake)

		initialStake := bunker("10000")
		err := stakingMgr.Stake(ctx, provider, initialStake)
		assert.NoError(err)

		// Slash 1: Downtime (5% of 10000 = 500)
		downtimeSlash := slashingContract.CalculateSlashAmount(initialStake, payment.ViolationDowntime)
		err = stakingMgr.Slash(ctx, provider, downtimeSlash)
		assert.NoError(err)
		afterFirst := stakingMgr.GetStake(provider)
		assert.Equal(bunker("9500").String(), afterFirst.String(), "After downtime slash: 9500 BUNKER")

		// Slash 2: SLA violation (10% of original 10000 = 1000)
		slaSlash := slashingContract.CalculateSlashAmount(initialStake, payment.ViolationSLAViolation)
		err = stakingMgr.Slash(ctx, provider, slaSlash)
		assert.NoError(err)
		afterSecond := stakingMgr.GetStake(provider)
		assert.Equal(bunker("8500").String(), afterSecond.String(), "After SLA slash: 8500 BUNKER")

		// Provider should still be at Silver tier (8500 >= 2000, < 10000) -> actually Bronze
		// 8500 is >= 2000 (Bronze) but < 10000 (Silver)
		assert.Equal(types.StakingTierBronze, stakingMgr.GetTier(provider),
			"After cumulative slashing, should drop to Bronze tier")
	})

	// Test slashing during unstake cooldown: tokens can be slashed while pending
	t.Run("SlashDuringCooldown", func(t *testing.T) {
		assert := testutil.NewAssertions(t)

		provider := common.HexToAddress("0xEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEE")
		stakingMgr := payment.NewStakingManager(nil, minStake)

		now := time.Now()
		stakingMgr.SetNowFunc(func() time.Time { return now })

		err := stakingMgr.Stake(ctx, provider, bunker("10000"))
		assert.NoError(err)

		// Request unstake of 5000
		err = stakingMgr.Unstake(ctx, provider, bunker("5000"))
		assert.NoError(err)

		// Slash 3000 during cooldown (stake is still 10000 in StakingManager)
		err = stakingMgr.Slash(ctx, provider, bunker("3000"))
		assert.NoError(err)
		assert.Equal(bunker("7000").String(), stakingMgr.GetStake(provider).String(),
			"Slash reduces stake even during cooldown")

		// Complete unstake after cooldown - should be capped at remaining stake
		now = now.Add(8 * 24 * time.Hour)
		withdrawn, err := stakingMgr.CompleteUnstake(ctx, provider)
		assert.NoError(err, "CompleteUnstake after cooldown should succeed")
		assert.Equal(bunker("5000").String(), withdrawn.String(),
			"Withdrawal should be the requested 5000 (still <= 7000 remaining)")

		// Remaining stake after withdrawal
		remaining := stakingMgr.GetStake(provider)
		assert.Equal(bunker("2000").String(), remaining.String(),
			"Remaining should be 7000 - 5000 = 2000")
	})
}

// TestE2E_StakingMultiProviderTierVerification tests multiple providers at different
// tier levels simultaneously and verifies correct tier assignments and provider state.
func TestE2E_StakingMultiProviderTierVerification(t *testing.T) {
	_ = testutil.NewAssertions(t)

	minStake := bunker("500")
	stakingMgr := payment.NewStakingManager(nil, minStake)
	ctx := context.Background()

	providers := []struct {
		name     string
		address  common.Address
		stake    *big.Int
		tier     types.StakingTier
		hasMin   bool
	}{
		{
			name:    "below-minimum",
			address: common.HexToAddress("0x0000000000000000000000000000000000000001"),
			stake:   bunker("100"),
			tier:    types.StakingTier(""),
			hasMin:  false,
		},
		{
			name:    "starter",
			address: common.HexToAddress("0x0000000000000000000000000000000000000002"),
			stake:   bunker("500"),
			tier:    types.StakingTierStarter,
			hasMin:  true,
		},
		{
			name:    "bronze-exact",
			address: common.HexToAddress("0x0000000000000000000000000000000000000003"),
			stake:   bunker("2000"),
			tier:    types.StakingTierBronze,
			hasMin:  true,
		},
		{
			name:    "bronze-mid",
			address: common.HexToAddress("0x0000000000000000000000000000000000000004"),
			stake:   bunker("5000"),
			tier:    types.StakingTierBronze,
			hasMin:  true,
		},
		{
			name:    "silver",
			address: common.HexToAddress("0x0000000000000000000000000000000000000005"),
			stake:   bunker("10000"),
			tier:    types.StakingTierSilver,
			hasMin:  true,
		},
		{
			name:    "gold",
			address: common.HexToAddress("0x0000000000000000000000000000000000000006"),
			stake:   bunker("50000"),
			tier:    types.StakingTierGold,
			hasMin:  true,
		},
		{
			name:    "platinum",
			address: common.HexToAddress("0x0000000000000000000000000000000000000007"),
			stake:   bunker("250000"),
			tier:    types.StakingTierPlatinum,
			hasMin:  true,
		},
		{
			name:    "whale",
			address: common.HexToAddress("0x0000000000000000000000000000000000000008"),
			stake:   bunker("1000000"),
			tier:    types.StakingTierPlatinum,
			hasMin:  true,
		},
	}

	// Stake for each provider (skip below-minimum which will fail)
	for _, p := range providers {
		t.Run("stake-"+p.name, func(t *testing.T) {
			assert := testutil.NewAssertions(t)
			err := stakingMgr.Stake(ctx, p.address, p.stake)
			if p.stake.Cmp(minStake) < 0 {
				assert.Error(err, "Staking below minimum should fail for %s", p.name)
			} else {
				assert.NoError(err, "Staking should succeed for %s", p.name)
			}
		})
	}

	// Verify tiers and state for each provider
	for _, p := range providers {
		t.Run("verify-"+p.name, func(t *testing.T) {
			assert := testutil.NewAssertions(t)

			actualTier := stakingMgr.GetTier(p.address)
			assert.Equal(p.tier, actualTier, "Provider %s should be tier %s", p.name, p.tier)

			hasMin := stakingMgr.HasMinimumStake(p.address)
			assert.Equal(p.hasMin, hasMin, "Provider %s HasMinimumStake mismatch", p.name)

			// Verify GetProviderState
			state := stakingMgr.GetProviderState(p.address)
			assert.NotNil(state, "GetProviderState should return non-nil for %s", p.name)
			assert.Equal(p.tier, state.Tier, "ProviderState.Tier should match for %s", p.name)

			if p.hasMin {
				assert.Equal(p.stake.String(), state.StakedAmount.String(),
					"ProviderState.StakedAmount should match for %s", p.name)
			}
		})
	}

	// Verify provider interactions don't affect each other
	t.Run("provider-isolation", func(t *testing.T) {
		assert := testutil.NewAssertions(t)

		// Slash the silver provider
		silverAddr := common.HexToAddress("0x0000000000000000000000000000000000000005")
		err := stakingMgr.Slash(ctx, silverAddr, bunker("5000"))
		assert.NoError(err, "Slashing Silver provider should succeed")

		// Silver provider should drop to Bronze (5000 remaining)
		assert.Equal(types.StakingTierBronze, stakingMgr.GetTier(silverAddr),
			"Slashed Silver should become Bronze")

		// Gold provider should be unaffected
		goldAddr := common.HexToAddress("0x0000000000000000000000000000000000000006")
		assert.Equal(types.StakingTierGold, stakingMgr.GetTier(goldAddr),
			"Gold provider should be unaffected by Silver's slash")
		assert.Equal(bunker("50000").String(), stakingMgr.GetStake(goldAddr).String(),
			"Gold provider stake should be unchanged")

		// Starter provider should be unaffected
		starterAddr := common.HexToAddress("0x0000000000000000000000000000000000000002")
		assert.Equal(types.StakingTierStarter, stakingMgr.GetTier(starterAddr),
			"Starter provider should be unaffected")
	})

	// Test beneficiary delegation
	t.Run("beneficiary-delegation", func(t *testing.T) {
		assert := testutil.NewAssertions(t)

		provider := common.HexToAddress("0x0000000000000000000000000000000000000006") // Gold
		beneficiary := common.HexToAddress("0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF")

		now := time.Now()
		stakingMgr.SetNowFunc(func() time.Time { return now })

		// Set beneficiary (starts 24h timelock)
		err := stakingMgr.SetBeneficiary(provider, beneficiary)
		assert.NoError(err, "SetBeneficiary should succeed")

		// Before timelock, beneficiary should still be self
		currentBeneficiary := stakingMgr.GetBeneficiary(provider)
		assert.Equal(provider, currentBeneficiary,
			"Beneficiary should still be self before timelock")

		// Advance time past timelock (24h + 1 minute)
		now = now.Add(24*time.Hour + time.Minute)

		// Now beneficiary should be the new address
		currentBeneficiary = stakingMgr.GetBeneficiary(provider)
		assert.Equal(beneficiary, currentBeneficiary,
			"Beneficiary should change after timelock")

		// Add and claim rewards
		stakingMgr.AddRewards(provider, bunker("100"))
		claimed, err := stakingMgr.ClaimRewards(ctx, provider)
		assert.NoError(err, "ClaimRewards should succeed")
		assert.Equal(bunker("100").String(), claimed.String(), "Should claim 100 BUNKER")

		// Second claim should fail (no rewards)
		_, err = stakingMgr.ClaimRewards(ctx, provider)
		assert.Error(err, "Second ClaimRewards should fail")
	})
}
