package payment

import (
	"context"
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// StakingManager manages provider staking
type StakingManager struct {
	client     *ethclient.Client
	stakes     map[common.Address]*big.Int
	mu         sync.RWMutex
	minStake   *big.Int
}

// NewStakingManager creates a new staking manager
func NewStakingManager(client *ethclient.Client, minStake *big.Int) *StakingManager {
	return &StakingManager{
		client:   client,
		stakes:   make(map[common.Address]*big.Int),
		minStake: minStake,
	}
}

// Stake stakes BUNKER tokens for a provider
func (sm *StakingManager) Stake(ctx context.Context, provider common.Address, amount *big.Int) error {
	if amount.Cmp(sm.minStake) < 0 {
		return fmt.Errorf("stake amount below minimum: %s", sm.minStake.String())
	}

	sm.mu.Lock()
	currentStake, exists := sm.stakes[provider]
	if !exists {
		sm.stakes[provider] = new(big.Int).Set(amount)
	} else {
		sm.stakes[provider] = new(big.Int).Add(currentStake, amount)
	}
	sm.mu.Unlock()

	// Mock implementation: Log the staking action
	// TODO: Replace with actual Base network contract call
	// Contract call would be:
	//   tx, err := stakingContract.Transact(auth, "stake", amount)
	//   if err != nil { return err }
	//   _, err = bind.WaitMined(ctx, sm.client, tx)
	fmt.Printf("[MOCK] Staked %s tokens for provider %s\n", amount.String(), provider.Hex())

	return nil
}

// GetStake returns the stake amount for a provider
func (sm *StakingManager) GetStake(provider common.Address) *big.Int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	stake, exists := sm.stakes[provider]
	if !exists {
		return big.NewInt(0)
	}

	return new(big.Int).Set(stake)
}

// Slash slashes stake for misbehavior
func (sm *StakingManager) Slash(ctx context.Context, provider common.Address, amount *big.Int) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	currentStake, exists := sm.stakes[provider]
	if !exists {
		return fmt.Errorf("no stake found for provider")
	}

	if amount.Cmp(currentStake) > 0 {
		amount = currentStake
	}

	sm.stakes[provider] = new(big.Int).Sub(currentStake, amount)

	// Mock implementation: Log the slashing action
	// TODO: Replace with actual Base network contract call
	// Contract call would be:
	//   tx, err := stakingContract.Transact(auth, "slash", provider, amount)
	//   if err != nil { return err }
	//   _, err = bind.WaitMined(ctx, sm.client, tx)
	fmt.Printf("[MOCK] Slashed %s tokens from provider %s\n", amount.String(), provider.Hex())

	return nil
}

// HasMinimumStake checks if provider has minimum stake
func (sm *StakingManager) HasMinimumStake(provider common.Address) bool {
	stake := sm.GetStake(provider)
	return stake.Cmp(sm.minStake) >= 0
}
