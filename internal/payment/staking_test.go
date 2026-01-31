package payment

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
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
