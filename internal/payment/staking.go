package payment

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/moltbunker/moltbunker/internal/logging"
	"github.com/moltbunker/moltbunker/pkg/types"
)

// beneficiaryChange tracks a pending beneficiary change with timelock
type beneficiaryChange struct {
	NewBeneficiary common.Address
	RequestedAt    time.Time
}

// unstakeRequest tracks a pending unstake with cooldown
type unstakeRequest struct {
	Amount      *big.Int
	RequestedAt time.Time
}

// BeneficiaryTimelockDuration is the required wait time before a beneficiary change takes effect
const BeneficiaryTimelockDuration = 24 * time.Hour

// UnstakeCooldownDuration is the required wait time before unstaked tokens can be withdrawn
const UnstakeCooldownDuration = 7 * 24 * time.Hour

// StakingManager manages provider staking
type StakingManager struct {
	client              *ethclient.Client
	stakes              map[common.Address]*big.Int
	beneficiaries       map[common.Address]common.Address
	pendingBeneficiary  map[common.Address]*beneficiaryChange
	rewards             map[common.Address]*big.Int
	unstakeRequests     map[common.Address]*unstakeRequest
	mu                  sync.RWMutex
	minStake            *big.Int
	nowFunc             func() time.Time // for testing
}

// SetNowFunc sets the time function used by the staking manager (for testing)
func (sm *StakingManager) SetNowFunc(fn func() time.Time) {
	sm.nowFunc = fn
}

// NewStakingManager creates a new staking manager
func NewStakingManager(client *ethclient.Client, minStake *big.Int) *StakingManager {
	return &StakingManager{
		client:             client,
		stakes:             make(map[common.Address]*big.Int),
		beneficiaries:      make(map[common.Address]common.Address),
		pendingBeneficiary: make(map[common.Address]*beneficiaryChange),
		rewards:            make(map[common.Address]*big.Int),
		unstakeRequests:    make(map[common.Address]*unstakeRequest),
		minStake:           minStake,
		nowFunc:            time.Now,
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

	logging.Info("staked tokens",
		"provider", provider.Hex(),
		"amount", amount.String(),
		"total_stake", sm.stakes[provider].String())

	logging.Audit(logging.AuditEvent{
		Operation: "stake_created",
		Actor:     provider.Hex(),
		Target:    provider.Hex(),
		Result:    "success",
		Details:   fmt.Sprintf("amount=%s total_stake=%s", amount.String(), sm.stakes[provider].String()),
	})

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

	logging.Info("slashed tokens",
		"provider", provider.Hex(),
		"amount", amount.String(),
		"remaining_stake", sm.stakes[provider].String())

	logging.Audit(logging.AuditEvent{
		Operation: "stake_slashed",
		Actor:     "system",
		Target:    provider.Hex(),
		Result:    "success",
		Details:   fmt.Sprintf("amount=%s remaining_stake=%s", amount.String(), sm.stakes[provider].String()),
	})

	return nil
}

// HasMinimumStake checks if provider has minimum stake
func (sm *StakingManager) HasMinimumStake(provider common.Address) bool {
	stake := sm.GetStake(provider)
	return stake.Cmp(sm.minStake) >= 0
}

// SetBeneficiary initiates a beneficiary address change for a provider.
// The change is subject to a 24-hour timelock before it takes effect.
// Calling SetBeneficiary again before the timelock expires replaces the pending change.
func (sm *StakingManager) SetBeneficiary(provider, beneficiary common.Address) error {
	if beneficiary == (common.Address{}) {
		return fmt.Errorf("beneficiary address cannot be zero address")
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Provider must have a stake
	stake, exists := sm.stakes[provider]
	if !exists || stake.Sign() == 0 {
		return fmt.Errorf("provider has no stake")
	}

	sm.pendingBeneficiary[provider] = &beneficiaryChange{
		NewBeneficiary: beneficiary,
		RequestedAt:    sm.nowFunc(),
	}

	logging.Info("beneficiary change requested",
		"provider", provider.Hex(),
		"new_beneficiary", beneficiary.Hex(),
		"timelock", BeneficiaryTimelockDuration.String())

	return nil
}

// GetBeneficiary returns the active beneficiary address for a provider.
// If no beneficiary is set, the provider's own address is returned.
// A pending beneficiary change is applied automatically if the timelock has elapsed.
func (sm *StakingManager) GetBeneficiary(provider common.Address) common.Address {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Check if there is a pending beneficiary change that has matured
	if pending, exists := sm.pendingBeneficiary[provider]; exists {
		if sm.nowFunc().Sub(pending.RequestedAt) >= BeneficiaryTimelockDuration {
			sm.beneficiaries[provider] = pending.NewBeneficiary
			delete(sm.pendingBeneficiary, provider)
		}
	}

	if beneficiary, exists := sm.beneficiaries[provider]; exists {
		return beneficiary
	}

	return provider
}

// ClaimRewards claims accumulated rewards for a provider and sends them to the beneficiary.
func (sm *StakingManager) ClaimRewards(ctx context.Context, provider common.Address) (*big.Int, error) {
	// Resolve the beneficiary (this also applies matured pending changes)
	beneficiary := sm.GetBeneficiary(provider)

	sm.mu.Lock()
	defer sm.mu.Unlock()

	reward, exists := sm.rewards[provider]
	if !exists || reward.Sign() == 0 {
		return big.NewInt(0), fmt.Errorf("no rewards to claim")
	}

	claimed := new(big.Int).Set(reward)
	sm.rewards[provider] = big.NewInt(0)

	logging.Info("claimed rewards",
		"provider", provider.Hex(),
		"beneficiary", beneficiary.Hex(),
		"amount", claimed.String())

	return claimed, nil
}

// AddRewards adds rewards to a provider's balance (used internally by the system)
func (sm *StakingManager) AddRewards(provider common.Address, amount *big.Int) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	current, exists := sm.rewards[provider]
	if !exists {
		sm.rewards[provider] = new(big.Int).Set(amount)
	} else {
		sm.rewards[provider] = new(big.Int).Add(current, amount)
	}
}

// GetTier returns the staking tier for a provider based on their staked amount.
// Tiers are evaluated from highest to lowest; the highest qualifying tier is returned.
// Returns empty string if the provider does not meet any tier minimum.
func (sm *StakingManager) GetTier(provider common.Address) types.StakingTier {
	stake := sm.GetStake(provider)

	tiers := types.DefaultStakingTiers()

	// Parse all tier minimums
	for _, cfg := range tiers {
		if err := cfg.ParseMinStake(); err != nil {
			continue
		}
	}

	// Evaluate tiers from highest to lowest
	tierOrder := []types.StakingTier{
		types.StakingTierPlatinum,
		types.StakingTierGold,
		types.StakingTierSilver,
		types.StakingTierBronze,
		types.StakingTierStarter,
	}

	for _, tier := range tierOrder {
		cfg := tiers[tier]
		if cfg.MinStake != nil && stake.Cmp(cfg.MinStake) >= 0 {
			return tier
		}
	}

	return ""
}

// Unstake initiates an unstaking request with a 7-day cooldown.
// The tokens remain staked during the cooldown period.
func (sm *StakingManager) Unstake(ctx context.Context, provider common.Address, amount *big.Int) error {
	if amount.Sign() <= 0 {
		return fmt.Errorf("unstake amount must be positive")
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	currentStake, exists := sm.stakes[provider]
	if !exists || currentStake.Sign() == 0 {
		return fmt.Errorf("no stake found for provider")
	}

	if amount.Cmp(currentStake) > 0 {
		return fmt.Errorf("unstake amount exceeds current stake: have %s, want %s",
			currentStake.String(), amount.String())
	}

	// Check if there's already a pending unstake request
	if existing, exists := sm.unstakeRequests[provider]; exists {
		elapsed := sm.nowFunc().Sub(existing.RequestedAt)
		if elapsed < UnstakeCooldownDuration {
			return fmt.Errorf("existing unstake request pending, cooldown expires in %s",
				(UnstakeCooldownDuration - elapsed).Round(time.Second))
		}

		// Previous cooldown has expired; execute it before creating a new request
		sm.stakes[provider] = new(big.Int).Sub(sm.stakes[provider], existing.Amount)
		delete(sm.unstakeRequests, provider)

		// Re-check that the new unstake doesn't exceed updated stake
		if amount.Cmp(sm.stakes[provider]) > 0 {
			return fmt.Errorf("unstake amount exceeds current stake after pending withdrawal: have %s, want %s",
				sm.stakes[provider].String(), amount.String())
		}
	}

	sm.unstakeRequests[provider] = &unstakeRequest{
		Amount:      new(big.Int).Set(amount),
		RequestedAt: sm.nowFunc(),
	}

	logging.Info("unstake request created",
		"provider", provider.Hex(),
		"amount", amount.String(),
		"cooldown", UnstakeCooldownDuration.String())

	return nil
}

// CompleteUnstake completes an unstaking request after the cooldown period has elapsed.
func (sm *StakingManager) CompleteUnstake(ctx context.Context, provider common.Address) (*big.Int, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	request, exists := sm.unstakeRequests[provider]
	if !exists {
		return nil, fmt.Errorf("no pending unstake request")
	}

	elapsed := sm.nowFunc().Sub(request.RequestedAt)
	if elapsed < UnstakeCooldownDuration {
		return nil, fmt.Errorf("cooldown period not elapsed, %s remaining",
			(UnstakeCooldownDuration - elapsed).Round(time.Second))
	}

	currentStake := sm.stakes[provider]
	unstakeAmount := request.Amount

	// Cap at current stake in case slashing occurred during cooldown
	if unstakeAmount.Cmp(currentStake) > 0 {
		unstakeAmount = new(big.Int).Set(currentStake)
	}

	sm.stakes[provider] = new(big.Int).Sub(currentStake, unstakeAmount)
	delete(sm.unstakeRequests, provider)

	logging.Info("unstake completed",
		"provider", provider.Hex(),
		"amount", unstakeAmount.String(),
		"remaining_stake", sm.stakes[provider].String())

	return unstakeAmount, nil
}

// GetProviderState returns the current state of a provider
func (sm *StakingManager) GetProviderState(provider common.Address) *types.ProviderState {
	stake := sm.GetStake(provider)
	tier := sm.GetTier(provider)
	beneficiary := sm.GetBeneficiary(provider)

	sm.mu.RLock()
	reward, _ := sm.rewards[provider]
	if reward == nil {
		reward = big.NewInt(0)
	}

	var unstakeInitiated *time.Time
	if req, exists := sm.unstakeRequests[provider]; exists {
		t := req.RequestedAt
		unstakeInitiated = &t
	}
	sm.mu.RUnlock()

	return &types.ProviderState{
		WalletAddress:    beneficiary.Hex(),
		StakedAmount:     stake,
		Tier:             tier,
		TotalEarnings:    new(big.Int).Set(reward),
		UnstakeInitiated: unstakeInitiated,
	}
}
