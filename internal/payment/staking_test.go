package payment

import (
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/moltbunker/moltbunker/pkg/types"
)

func createTestStakingManager() *StakingManager {
	// Note: This would normally connect to a test network
	// For unit tests, we'll use nil client and test the logic
	return NewStakingManager(nil, big.NewInt(1000000000000000000)) // 1 BUNKER minimum
}

func TestStakingManager_Stake(t *testing.T) {
	sm := createTestStakingManager()

	provider := common.HexToAddress("0x1234567890123456789012345678901234567890")
	amount := big.NewInt(2000000000000000000) // 2 BUNKER

	err := sm.Stake(nil, provider, amount)
	if err != nil {
		t.Fatalf("Failed to stake: %v", err)
	}

	stake := sm.GetStake(provider)
	if stake.Cmp(amount) != 0 {
		t.Errorf("Stake mismatch: got %s, want %s", stake.String(), amount.String())
	}
}

func TestStakingManager_Stake_BelowMinimum(t *testing.T) {
	sm := createTestStakingManager()

	provider := common.HexToAddress("0x1234567890123456789012345678901234567890")
	amount := big.NewInt(500000000000000000) // 0.5 BUNKER (below minimum)

	err := sm.Stake(nil, provider, amount)
	if err == nil {
		t.Error("Should fail when stake is below minimum")
	}
}

func TestStakingManager_GetStake(t *testing.T) {
	sm := createTestStakingManager()

	provider := common.HexToAddress("0x1234567890123456789012345678901234567890")

	stake := sm.GetStake(provider)
	if stake.Sign() != 0 {
		t.Error("Stake should be zero for new provider")
	}
}

func TestStakingManager_GetStake_AfterStaking(t *testing.T) {
	sm := createTestStakingManager()

	provider := common.HexToAddress("0x1234567890123456789012345678901234567890")
	amount := big.NewInt(2000000000000000000)

	sm.Stake(nil, provider, amount)

	stake := sm.GetStake(provider)
	if stake.Cmp(amount) != 0 {
		t.Errorf("Stake mismatch: got %s, want %s", stake.String(), amount.String())
	}
}

func TestStakingManager_Slash(t *testing.T) {
	sm := createTestStakingManager()

	provider := common.HexToAddress("0x1234567890123456789012345678901234567890")
	stakeAmount := big.NewInt(2000000000000000000)
	slashAmount := big.NewInt(500000000000000000) // 0.5 BUNKER

	sm.Stake(nil, provider, stakeAmount)

	err := sm.Slash(nil, provider, slashAmount)
	if err != nil {
		t.Fatalf("Failed to slash: %v", err)
	}

	remainingStake := sm.GetStake(provider)
	expected := new(big.Int).Sub(stakeAmount, slashAmount)

	if remainingStake.Cmp(expected) != 0 {
		t.Errorf("Remaining stake mismatch: got %s, want %s", remainingStake.String(), expected.String())
	}
}

func TestStakingManager_Slash_ExceedsStake(t *testing.T) {
	sm := createTestStakingManager()

	provider := common.HexToAddress("0x1234567890123456789012345678901234567890")
	stakeAmount := big.NewInt(2000000000000000000)
	slashAmount := big.NewInt(3000000000000000000) // More than stake

	sm.Stake(nil, provider, stakeAmount)

	err := sm.Slash(nil, provider, slashAmount)
	if err != nil {
		t.Fatalf("Failed to slash: %v", err)
	}

	// Should only slash up to stake amount
	remainingStake := sm.GetStake(provider)
	if remainingStake.Sign() != 0 {
		t.Error("Remaining stake should be zero after slashing more than stake")
	}
}

func TestStakingManager_HasMinimumStake(t *testing.T) {
	sm := createTestStakingManager()

	provider1 := common.HexToAddress("0x1111111111111111111111111111111111111111")
	provider2 := common.HexToAddress("0x2222222222222222222222222222222222222222")

	// Provider 1: below minimum
	sm.Stake(nil, provider1, big.NewInt(500000000000000000))

	// Provider 2: above minimum
	sm.Stake(nil, provider2, big.NewInt(2000000000000000000))

	if sm.HasMinimumStake(provider1) {
		t.Error("Provider 1 should not have minimum stake")
	}

	if !sm.HasMinimumStake(provider2) {
		t.Error("Provider 2 should have minimum stake")
	}
}

func TestStakingManager_SetBeneficiary(t *testing.T) {
	sm := createTestStakingManager()

	provider := common.HexToAddress("0x1234567890123456789012345678901234567890")
	beneficiary := common.HexToAddress("0xAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	stakeAmount := big.NewInt(2000000000000000000) // 2 BUNKER

	// Should fail without a stake
	err := sm.SetBeneficiary(provider, beneficiary)
	if err == nil {
		t.Error("Should fail when provider has no stake")
	}

	// Stake first
	sm.Stake(nil, provider, stakeAmount)

	// Should fail with zero address
	err = sm.SetBeneficiary(provider, common.Address{})
	if err == nil {
		t.Error("Should fail with zero beneficiary address")
	}

	// Should succeed with valid beneficiary
	err = sm.SetBeneficiary(provider, beneficiary)
	if err != nil {
		t.Fatalf("Failed to set beneficiary: %v", err)
	}

	// Before timelock, beneficiary should still be the provider
	got := sm.GetBeneficiary(provider)
	if got != provider {
		t.Errorf("Beneficiary before timelock should be provider, got %s", got.Hex())
	}
}

func TestStakingManager_BeneficiaryTimelock(t *testing.T) {
	sm := createTestStakingManager()

	provider := common.HexToAddress("0x1234567890123456789012345678901234567890")
	beneficiary := common.HexToAddress("0xAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	stakeAmount := big.NewInt(2000000000000000000)

	sm.Stake(nil, provider, stakeAmount)

	// Set a fixed time for deterministic testing
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	sm.nowFunc = func() time.Time { return now }

	err := sm.SetBeneficiary(provider, beneficiary)
	if err != nil {
		t.Fatalf("Failed to set beneficiary: %v", err)
	}

	// Advance time by 12 hours (before timelock)
	sm.nowFunc = func() time.Time { return now.Add(12 * time.Hour) }
	got := sm.GetBeneficiary(provider)
	if got != provider {
		t.Errorf("Beneficiary should still be provider before 24h timelock, got %s", got.Hex())
	}

	// Advance time by 24 hours (exactly at timelock)
	sm.nowFunc = func() time.Time { return now.Add(24 * time.Hour) }
	got = sm.GetBeneficiary(provider)
	if got != beneficiary {
		t.Errorf("Beneficiary should be updated after 24h timelock, got %s, want %s",
			got.Hex(), beneficiary.Hex())
	}

	// Subsequent calls should return the same beneficiary (now committed)
	got = sm.GetBeneficiary(provider)
	if got != beneficiary {
		t.Errorf("Beneficiary should remain set, got %s, want %s",
			got.Hex(), beneficiary.Hex())
	}
}

func TestStakingManager_GetTier(t *testing.T) {
	sm := createTestStakingManager()

	provider := common.HexToAddress("0x1234567890123456789012345678901234567890")

	tests := []struct {
		name       string
		stakeWei   string
		expectTier types.StakingTier
	}{
		{
			name:       "no stake - no tier",
			stakeWei:   "0",
			expectTier: "",
		},
		{
			name:       "below starter - 10000 BUNKER",
			stakeWei:   "10000000000000000000000",
			expectTier: "",
		},
		{
			name:       "starter tier - 1000000 BUNKER",
			stakeWei:   "1000000000000000000000000",
			expectTier: types.StakingTierStarter,
		},
		{
			name:       "bronze tier - 5000000 BUNKER",
			stakeWei:   "5000000000000000000000000",
			expectTier: types.StakingTierBronze,
		},
		{
			name:       "silver tier - 10000000 BUNKER",
			stakeWei:   "10000000000000000000000000",
			expectTier: types.StakingTierSilver,
		},
		{
			name:       "gold tier - 100000000 BUNKER",
			stakeWei:   "100000000000000000000000000",
			expectTier: types.StakingTierGold,
		},
		{
			name:       "platinum tier - 1000000000 BUNKER",
			stakeWei:   "1000000000000000000000000000",
			expectTier: types.StakingTierPlatinum,
		},
		{
			name:       "above platinum - 10000000000 BUNKER",
			stakeWei:   "10000000000000000000000000000",
			expectTier: types.StakingTierPlatinum,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset stakes
			sm.mu.Lock()
			sm.stakes = make(map[common.Address]*big.Int)
			sm.mu.Unlock()

			amount, ok := new(big.Int).SetString(tt.stakeWei, 10)
			if !ok {
				t.Fatalf("Failed to parse stake amount: %s", tt.stakeWei)
			}

			if amount.Sign() > 0 {
				// Use low minStake so we can stake any amount
				sm.minStake = big.NewInt(1)
				sm.Stake(nil, provider, amount)
			}

			tier := sm.GetTier(provider)
			if tier != tt.expectTier {
				t.Errorf("GetTier() = %q, want %q", tier, tt.expectTier)
			}
		})
	}
}

func TestStakingManager_ClaimRewards(t *testing.T) {
	sm := createTestStakingManager()

	provider := common.HexToAddress("0x1234567890123456789012345678901234567890")
	beneficiary := common.HexToAddress("0xAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	stakeAmount := big.NewInt(2000000000000000000)
	rewardAmount := big.NewInt(500000000000000000) // 0.5 BUNKER

	sm.Stake(nil, provider, stakeAmount)

	// Claiming with no rewards should fail
	_, err := sm.ClaimRewards(nil, provider)
	if err == nil {
		t.Error("Should fail when no rewards to claim")
	}

	// Add rewards
	sm.AddRewards(provider, rewardAmount)

	// Claim without beneficiary set (should use provider address)
	claimed, err := sm.ClaimRewards(nil, provider)
	if err != nil {
		t.Fatalf("Failed to claim rewards: %v", err)
	}
	if claimed.Cmp(rewardAmount) != 0 {
		t.Errorf("Claimed amount mismatch: got %s, want %s", claimed.String(), rewardAmount.String())
	}

	// Rewards should be zero after claim
	_, err = sm.ClaimRewards(nil, provider)
	if err == nil {
		t.Error("Should fail when rewards already claimed")
	}

	// Set beneficiary and advance timelock
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	sm.nowFunc = func() time.Time { return now }
	sm.SetBeneficiary(provider, beneficiary)

	// Advance past timelock
	sm.nowFunc = func() time.Time { return now.Add(25 * time.Hour) }

	// Add more rewards and claim - should go to beneficiary
	sm.AddRewards(provider, rewardAmount)
	claimed, err = sm.ClaimRewards(nil, provider)
	if err != nil {
		t.Fatalf("Failed to claim rewards with beneficiary: %v", err)
	}
	if claimed.Cmp(rewardAmount) != 0 {
		t.Errorf("Claimed amount mismatch: got %s, want %s", claimed.String(), rewardAmount.String())
	}

	// Verify beneficiary was resolved
	got := sm.GetBeneficiary(provider)
	if got != beneficiary {
		t.Errorf("Beneficiary should be %s, got %s", beneficiary.Hex(), got.Hex())
	}
}

func TestStakingManager_Unstake(t *testing.T) {
	sm := createTestStakingManager()

	provider := common.HexToAddress("0x1234567890123456789012345678901234567890")
	stakeAmount := big.NewInt(5000000000000000000) // 5 BUNKER
	unstakeAmount := big.NewInt(2000000000000000000) // 2 BUNKER

	// Should fail without a stake
	err := sm.Unstake(nil, provider, unstakeAmount)
	if err == nil {
		t.Error("Should fail when provider has no stake")
	}

	sm.Stake(nil, provider, stakeAmount)

	// Should fail with zero or negative amount
	err = sm.Unstake(nil, provider, big.NewInt(0))
	if err == nil {
		t.Error("Should fail with zero unstake amount")
	}

	// Should fail if amount exceeds stake
	err = sm.Unstake(nil, provider, big.NewInt(6000000000000000000))
	if err == nil {
		t.Error("Should fail when unstake exceeds stake")
	}

	// Set a fixed time for deterministic testing
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	sm.nowFunc = func() time.Time { return now }

	// Should succeed with valid amount
	err = sm.Unstake(nil, provider, unstakeAmount)
	if err != nil {
		t.Fatalf("Failed to unstake: %v", err)
	}

	// Stake should not change yet (still in cooldown)
	currentStake := sm.GetStake(provider)
	if currentStake.Cmp(stakeAmount) != 0 {
		t.Errorf("Stake should not change during cooldown: got %s, want %s",
			currentStake.String(), stakeAmount.String())
	}

	// Should fail to create another unstake during cooldown
	sm.nowFunc = func() time.Time { return now.Add(3 * 24 * time.Hour) } // 3 days later
	err = sm.Unstake(nil, provider, big.NewInt(1000000000000000000))
	if err == nil {
		t.Error("Should fail when existing unstake request is pending")
	}

	// Complete unstake after cooldown
	sm.nowFunc = func() time.Time { return now.Add(8 * 24 * time.Hour) } // 8 days later
	completed, err := sm.CompleteUnstake(nil, provider)
	if err != nil {
		t.Fatalf("Failed to complete unstake: %v", err)
	}
	if completed.Cmp(unstakeAmount) != 0 {
		t.Errorf("Completed unstake amount mismatch: got %s, want %s",
			completed.String(), unstakeAmount.String())
	}

	// Stake should be reduced
	remaining := sm.GetStake(provider)
	expected := new(big.Int).Sub(stakeAmount, unstakeAmount)
	if remaining.Cmp(expected) != 0 {
		t.Errorf("Remaining stake mismatch: got %s, want %s",
			remaining.String(), expected.String())
	}

	// CompleteUnstake should fail when no pending request
	_, err = sm.CompleteUnstake(nil, provider)
	if err == nil {
		t.Error("Should fail when no pending unstake request")
	}
}

func TestStakingManager_Unstake_CooldownNotElapsed(t *testing.T) {
	sm := createTestStakingManager()

	provider := common.HexToAddress("0x1234567890123456789012345678901234567890")
	stakeAmount := big.NewInt(5000000000000000000)
	unstakeAmount := big.NewInt(2000000000000000000)

	sm.Stake(nil, provider, stakeAmount)

	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	sm.nowFunc = func() time.Time { return now }

	sm.Unstake(nil, provider, unstakeAmount)

	// Try to complete before cooldown
	sm.nowFunc = func() time.Time { return now.Add(6 * 24 * time.Hour) } // 6 days, not enough
	_, err := sm.CompleteUnstake(nil, provider)
	if err == nil {
		t.Error("Should fail when cooldown period has not elapsed")
	}
}

func TestStakingManager_GetProviderState(t *testing.T) {
	sm := createTestStakingManager()
	sm.minStake = big.NewInt(1) // low minimum for testing

	provider := common.HexToAddress("0x1234567890123456789012345678901234567890")

	// Stake enough for starter tier (1,000,000 BUNKER)
	stakeAmount, _ := new(big.Int).SetString("1000000000000000000000000", 10) // 1,000,000 BUNKER
	sm.Stake(nil, provider, stakeAmount)

	// Add rewards
	rewardAmount := big.NewInt(500000000000000000)
	sm.AddRewards(provider, rewardAmount)

	state := sm.GetProviderState(provider)

	if state.StakedAmount.Cmp(stakeAmount) != 0 {
		t.Errorf("StakedAmount mismatch: got %s, want %s",
			state.StakedAmount.String(), stakeAmount.String())
	}

	if state.Tier != types.StakingTierStarter {
		t.Errorf("Tier mismatch: got %q, want %q", state.Tier, types.StakingTierStarter)
	}

	if state.WalletAddress != provider.Hex() {
		t.Errorf("WalletAddress should be provider when no beneficiary set, got %s", state.WalletAddress)
	}

	if state.TotalEarnings.Cmp(rewardAmount) != 0 {
		t.Errorf("TotalEarnings mismatch: got %s, want %s",
			state.TotalEarnings.String(), rewardAmount.String())
	}

	if state.UnstakeInitiated != nil {
		t.Error("UnstakeInitiated should be nil when no unstake request")
	}

	// Initiate unstake and check state
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	sm.nowFunc = func() time.Time { return now }
	sm.Unstake(nil, provider, big.NewInt(1000000000000000000))

	state = sm.GetProviderState(provider)
	if state.UnstakeInitiated == nil {
		t.Error("UnstakeInitiated should be set after unstake request")
	}
	if !state.UnstakeInitiated.Equal(now) {
		t.Errorf("UnstakeInitiated time mismatch: got %v, want %v", *state.UnstakeInitiated, now)
	}
}
