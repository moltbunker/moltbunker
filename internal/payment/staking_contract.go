package payment

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/moltbunker/moltbunker/internal/util"
	pkgtypes "github.com/moltbunker/moltbunker/pkg/types"
)

// StakingContract provides interface to the staking smart contract
type StakingContract struct {
	baseClient     *BaseClient
	tokenContract  *TokenContract
	contract       *bind.BoundContract
	contractABI    abi.ABI
	contractAddr   common.Address
	mockMode       bool
	retryConfig    *util.RetryConfig

	// Mock state
	mockStakes   map[common.Address]*big.Int
	mockPending  map[common.Address]*PendingUnstake
	mockMu       sync.RWMutex
}

// PendingUnstake represents a pending unstake request
type PendingUnstake struct {
	Amount     *big.Int
	UnlockTime time.Time
}

// StakeInfo contains staking information for a provider
type StakeInfo struct {
	StakedAmount   *big.Int
	PendingUnstake *big.Int
	UnlockTime     time.Time
	Tier           pkgtypes.StakingTier
}

// StakeEvent represents a staking event
type StakeEvent struct {
	Provider   common.Address
	Amount     *big.Int
	TotalStake *big.Int
	Timestamp  time.Time
}

// NewStakingContract creates a new staking contract client
func NewStakingContract(baseClient *BaseClient, tokenContract *TokenContract, contractAddr common.Address) (*StakingContract, error) {
	sc := &StakingContract{
		baseClient:    baseClient,
		tokenContract: tokenContract,
		contractAddr:  contractAddr,
		mockStakes:    make(map[common.Address]*big.Int),
		mockPending:   make(map[common.Address]*PendingUnstake),
		retryConfig:   util.DefaultRetryConfig(),
	}

	// If no base client, use mock mode
	if baseClient == nil || !baseClient.IsConnected() {
		sc.mockMode = true
		return sc, nil
	}

	parsedABI, err := abi.JSON(strings.NewReader(StakingContractABI))
	if err != nil {
		return nil, fmt.Errorf("failed to parse staking ABI: %w", err)
	}
	sc.contractABI = parsedABI

	client := baseClient.Client()
	sc.contract = bind.NewBoundContract(contractAddr, parsedABI, client, client, client)

	return sc, nil
}

// NewMockStakingContract creates a mock staking contract for testing
func NewMockStakingContract() *StakingContract {
	return &StakingContract{
		mockMode:    true,
		mockStakes:  make(map[common.Address]*big.Int),
		mockPending: make(map[common.Address]*PendingUnstake),
	}
}

// IsMockMode returns whether running in mock mode
func (sc *StakingContract) IsMockMode() bool {
	return sc.mockMode
}

// Stake stakes BUNKER tokens
func (sc *StakingContract) Stake(ctx context.Context, amount *big.Int) (*types.Transaction, error) {
	if sc.mockMode {
		return sc.mockStake(ctx, amount)
	}

	// First approve the staking contract to spend tokens
	if sc.tokenContract != nil {
		_, err := sc.tokenContract.Approve(ctx, sc.contractAddr, amount)
		if err != nil {
			return nil, fmt.Errorf("failed to approve token spend: %w", err)
		}
	}

	auth, err := sc.baseClient.GetTransactOpts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction options: %w", err)
	}

	tx, err := sc.contract.Transact(auth, "stake", amount)
	if err != nil {
		return nil, fmt.Errorf("failed to stake: %w", err)
	}

	return tx, nil
}

// StakeAndWait stakes tokens and waits for confirmation
func (sc *StakingContract) StakeAndWait(ctx context.Context, amount *big.Int) (*types.Receipt, error) {
	tx, err := sc.Stake(ctx, amount)
	if err != nil {
		return nil, err
	}

	if sc.mockMode || tx == nil {
		return nil, nil
	}

	return sc.baseClient.WaitForTransaction(ctx, tx)
}

// mockStake handles staking in mock mode
func (sc *StakingContract) mockStake(_ context.Context, amount *big.Int) (*types.Transaction, error) {
	sc.mockMu.Lock()
	defer sc.mockMu.Unlock()

	var addr common.Address
	if sc.baseClient != nil {
		addr = sc.baseClient.Address()
	}
	current, exists := sc.mockStakes[addr]
	if !exists {
		sc.mockStakes[addr] = new(big.Int).Set(amount)
	} else {
		sc.mockStakes[addr] = new(big.Int).Add(current, amount)
	}

	fmt.Printf("[MOCK] Staked %s BUNKER tokens for %s (total: %s)\n",
		amount.String(), addr.Hex(), sc.mockStakes[addr].String())

	return nil, nil
}

// RequestUnstake requests to unstake tokens (starts cooldown period)
func (sc *StakingContract) RequestUnstake(ctx context.Context, amount *big.Int) (*types.Transaction, error) {
	if sc.mockMode {
		return sc.mockRequestUnstake(ctx, amount)
	}

	auth, err := sc.baseClient.GetTransactOpts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction options: %w", err)
	}

	tx, err := sc.contract.Transact(auth, "requestUnstake", amount)
	if err != nil {
		return nil, fmt.Errorf("failed to request unstake: %w", err)
	}

	return tx, nil
}

// mockRequestUnstake handles unstake request in mock mode
func (sc *StakingContract) mockRequestUnstake(_ context.Context, amount *big.Int) (*types.Transaction, error) {
	sc.mockMu.Lock()
	defer sc.mockMu.Unlock()

	var addr common.Address
	if sc.baseClient != nil {
		addr = sc.baseClient.Address()
	}
	current, exists := sc.mockStakes[addr]
	if !exists || current.Cmp(amount) < 0 {
		return nil, fmt.Errorf("insufficient stake")
	}

	// 7 day cooldown (mock)
	sc.mockPending[addr] = &PendingUnstake{
		Amount:     new(big.Int).Set(amount),
		UnlockTime: time.Now().Add(7 * 24 * time.Hour),
	}
	sc.mockStakes[addr] = new(big.Int).Sub(current, amount)

	fmt.Printf("[MOCK] Requested unstake of %s BUNKER for %s (unlock: %s)\n",
		amount.String(), addr.Hex(), sc.mockPending[addr].UnlockTime.Format(time.RFC3339))

	return nil, nil
}

// Withdraw withdraws unstaked tokens after cooldown
func (sc *StakingContract) Withdraw(ctx context.Context) (*types.Transaction, error) {
	if sc.mockMode {
		return sc.mockWithdraw(ctx)
	}

	auth, err := sc.baseClient.GetTransactOpts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction options: %w", err)
	}

	tx, err := sc.contract.Transact(auth, "withdraw")
	if err != nil {
		return nil, fmt.Errorf("failed to withdraw: %w", err)
	}

	return tx, nil
}

// mockWithdraw handles withdrawal in mock mode
func (sc *StakingContract) mockWithdraw(_ context.Context) (*types.Transaction, error) {
	sc.mockMu.Lock()
	defer sc.mockMu.Unlock()

	var addr common.Address
	if sc.baseClient != nil {
		addr = sc.baseClient.Address()
	}
	pending, exists := sc.mockPending[addr]
	if !exists {
		return nil, fmt.Errorf("no pending unstake")
	}

	if time.Now().Before(pending.UnlockTime) {
		return nil, fmt.Errorf("cooldown not complete")
	}

	fmt.Printf("[MOCK] Withdrew %s BUNKER tokens for %s\n",
		pending.Amount.String(), addr.Hex())

	delete(sc.mockPending, addr)
	return nil, nil
}

// GetStake returns the staked amount for an address
func (sc *StakingContract) GetStake(ctx context.Context, provider common.Address) (*big.Int, error) {
	if sc.mockMode {
		sc.mockMu.RLock()
		defer sc.mockMu.RUnlock()
		stake, exists := sc.mockStakes[provider]
		if !exists {
			return big.NewInt(0), nil
		}
		return new(big.Int).Set(stake), nil
	}

	var result []interface{}
	err := sc.contract.Call(&bind.CallOpts{Context: ctx}, &result, "getStake", provider)
	if err != nil {
		return nil, fmt.Errorf("failed to get stake: %w", err)
	}

	if len(result) == 0 {
		return big.NewInt(0), nil
	}
	if stake, ok := result[0].(*big.Int); ok {
		return stake, nil
	}
	return big.NewInt(0), nil
}

// GetStakeInfo returns full staking information
func (sc *StakingContract) GetStakeInfo(ctx context.Context, provider common.Address) (*StakeInfo, error) {
	stake, err := sc.GetStake(ctx, provider)
	if err != nil {
		return nil, err
	}

	info := &StakeInfo{
		StakedAmount:   stake,
		PendingUnstake: big.NewInt(0),
	}

	if sc.mockMode {
		sc.mockMu.RLock()
		if pending, exists := sc.mockPending[provider]; exists {
			info.PendingUnstake = new(big.Int).Set(pending.Amount)
			info.UnlockTime = pending.UnlockTime
		}
		sc.mockMu.RUnlock()
	} else {
		// Call getPendingUnstake
		var result []interface{}
		err := sc.contract.Call(&bind.CallOpts{Context: ctx}, &result, "getPendingUnstake", provider)
		if err == nil && len(result) >= 2 {
			if amount, ok := result[0].(*big.Int); ok {
				info.PendingUnstake = amount
			}
			if unlockTime, ok := result[1].(*big.Int); ok {
				info.UnlockTime = time.Unix(unlockTime.Int64(), 0)
			}
		}
	}

	// Determine tier based on stake
	info.Tier = sc.determineTier(stake)

	return info, nil
}

// GetTier returns the staking tier for a provider
func (sc *StakingContract) GetTier(ctx context.Context, provider common.Address) (pkgtypes.StakingTier, error) {
	if sc.mockMode {
		stake, err := sc.GetStake(ctx, provider)
		if err != nil {
			return "", err
		}
		return sc.determineTier(stake), nil
	}

	var result []interface{}
	err := sc.contract.Call(&bind.CallOpts{Context: ctx}, &result, "getTier", provider)
	if err != nil {
		return "", fmt.Errorf("failed to get tier: %w", err)
	}

	if len(result) == 0 {
		return "", nil
	}
	// Contract returns uint8 tier index, map to tier name
	if tierIndex, ok := result[0].(uint8); ok {
		return sc.tierFromIndex(tierIndex), nil
	}
	return "", nil
}

// GetTotalStaked returns total staked tokens
func (sc *StakingContract) GetTotalStaked(ctx context.Context) (*big.Int, error) {
	if sc.mockMode {
		sc.mockMu.RLock()
		defer sc.mockMu.RUnlock()
		total := big.NewInt(0)
		for _, stake := range sc.mockStakes {
			total.Add(total, stake)
		}
		return total, nil
	}

	var result []interface{}
	err := sc.contract.Call(&bind.CallOpts{Context: ctx}, &result, "totalStaked")
	if err != nil {
		return nil, fmt.Errorf("failed to get total staked: %w", err)
	}

	if len(result) == 0 {
		return big.NewInt(0), nil
	}
	if total, ok := result[0].(*big.Int); ok {
		return total, nil
	}
	return big.NewInt(0), nil
}

// tierFromIndex maps a contract tier index to StakingTier
func (sc *StakingContract) tierFromIndex(index uint8) pkgtypes.StakingTier {
	tiers := []pkgtypes.StakingTier{
		"",                          // 0 = none
		pkgtypes.StakingTierStarter, // 1
		pkgtypes.StakingTierBronze,  // 2
		pkgtypes.StakingTierSilver,  // 3
		pkgtypes.StakingTierGold,    // 4
		pkgtypes.StakingTierPlatinum, // 5
	}
	if int(index) < len(tiers) {
		return tiers[index]
	}
	return ""
}

// determineTier determines the staking tier based on stake amount
func (sc *StakingContract) determineTier(stake *big.Int) pkgtypes.StakingTier {
	// Tier thresholds (in wei, assuming 18 decimals)
	// These should match the contract's tier thresholds
	tiers := []struct {
		tier   pkgtypes.StakingTier
		minWei *big.Int
	}{
		{pkgtypes.StakingTierPlatinum, parseWei("100000")}, // 100,000 BUNKER
		{pkgtypes.StakingTierGold, parseWei("50000")},      // 50,000 BUNKER
		{pkgtypes.StakingTierSilver, parseWei("10000")},    // 10,000 BUNKER
		{pkgtypes.StakingTierBronze, parseWei("2000")},     // 2,000 BUNKER
		{pkgtypes.StakingTierStarter, parseWei("500")},     // 500 BUNKER
	}

	for _, t := range tiers {
		if stake.Cmp(t.minWei) >= 0 {
			return t.tier
		}
	}
	return "" // No tier
}

// parseWei converts a token amount string to wei (18 decimals)
func parseWei(amount string) *big.Int {
	tokens, _ := new(big.Int).SetString(amount, 10)
	decimals := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	return new(big.Int).Mul(tokens, decimals)
}

// SubscribeStakeEvents subscribes to staking events
func (sc *StakingContract) SubscribeStakeEvents(ctx context.Context, ch chan<- *StakeEvent) error {
	if sc.mockMode {
		return nil // No events in mock mode
	}

	wsClient := sc.baseClient.WSClient()
	if wsClient == nil {
		return fmt.Errorf("WebSocket client not available")
	}

	query := ethereum.FilterQuery{
		Addresses: []common.Address{sc.contractAddr},
		Topics:    [][]common.Hash{{sc.contractABI.Events["Staked"].ID}},
	}

	logs := make(chan types.Log)
	sub, err := wsClient.SubscribeFilterLogs(ctx, query, logs)
	if err != nil {
		return fmt.Errorf("failed to subscribe to events: %w", err)
	}

	util.SafeGoWithName("stake-event-listener", func() {
		defer sub.Unsubscribe()
		for {
			select {
			case <-ctx.Done():
				return
			case log := <-logs:
				event, err := sc.parseStakeEvent(log)
				if err != nil {
					continue
				}
				ch <- event
			case err := <-sub.Err():
				if err != nil {
					fmt.Printf("Stake event subscription error: %v\n", err)
				}
				return
			}
		}
	})

	return nil
}

// parseStakeEvent parses a stake event from log
func (sc *StakingContract) parseStakeEvent(log types.Log) (*StakeEvent, error) {
	event := &StakeEvent{
		Timestamp: time.Now(),
	}

	if len(log.Topics) > 1 {
		event.Provider = common.HexToAddress(log.Topics[1].Hex())
	}

	// Parse data (amount, totalStake)
	if len(log.Data) >= 64 {
		event.Amount = new(big.Int).SetBytes(log.Data[:32])
		event.TotalStake = new(big.Int).SetBytes(log.Data[32:64])
	}

	return event, nil
}

// HasMinimumStake checks if provider has minimum stake
func (sc *StakingContract) HasMinimumStake(ctx context.Context, provider common.Address) (bool, error) {
	stake, err := sc.GetStake(ctx, provider)
	if err != nil {
		return false, err
	}

	minStake := parseWei("1000") // 1000 BUNKER minimum
	return stake.Cmp(minStake) >= 0, nil
}
