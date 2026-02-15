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
	"github.com/moltbunker/moltbunker/internal/logging"
	"github.com/moltbunker/moltbunker/internal/util"
	pkgtypes "github.com/moltbunker/moltbunker/pkg/types"
)

// StakingContract provides interface to the BunkerStaking smart contract
type StakingContract struct {
	baseClient    *BaseClient
	tokenContract *TokenContract
	contract      *bind.BoundContract
	contractABI   abi.ABI
	contractAddr  common.Address
	mockMode      bool
	retryConfig   *util.RetryConfig

	// Optional delegation contract for computing effective stake (personal + delegated)
	delegationContract *DelegationContract

	// Mock state
	mockStakes     map[common.Address]*big.Int
	mockPending    map[common.Address][]*PendingUnstake // queue-based
	mockIdentities map[common.Address]*NodeIdentity
	mockNodeIDMap  map[[32]byte]common.Address // nodeID → provider
	mockMu         sync.RWMutex
}

// SetDelegationContract sets the delegation contract for effective stake calculation.
// When set, determineTier includes delegated stake in addition to personal stake.
func (sc *StakingContract) SetDelegationContract(dc *DelegationContract) {
	sc.delegationContract = dc
}

// PendingUnstake represents a pending unstake request
type PendingUnstake struct {
	Amount     *big.Int
	UnlockTime time.Time
	Completed  bool
}

// ProviderInfoData contains provider information from the contract
type ProviderInfoData struct {
	StakedAmount   *big.Int
	TotalUnbonding *big.Int
	Beneficiary    common.Address
	RegisteredAt   time.Time
	Active         bool
	NodeID         [32]byte // On-chain NodeID binding
	Region         [32]byte // Region identifier
	Capabilities   uint64   // Capability flags
	Frozen         bool     // Provider frozen by governance
}

// NodeIdentity holds the identity fields for mock mode.
type NodeIdentity struct {
	NodeID       [32]byte
	Region       [32]byte
	Capabilities uint64
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
		baseClient:     baseClient,
		tokenContract:  tokenContract,
		contractAddr:   contractAddr,
		mockStakes:     make(map[common.Address]*big.Int),
		mockPending:    make(map[common.Address][]*PendingUnstake),
		mockIdentities: make(map[common.Address]*NodeIdentity),
		mockNodeIDMap:  make(map[[32]byte]common.Address),
		retryConfig:    util.DefaultRetryConfig(),
	}

	// Require a connected base client; use NewMockStakingContract() for testing
	if baseClient == nil {
		return nil, fmt.Errorf("base client is required (use NewMockStakingContract for testing)")
	}
	if !baseClient.IsConnected() {
		return nil, fmt.Errorf("base client not connected to RPC")
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
		mockMode:       true,
		mockStakes:     make(map[common.Address]*big.Int),
		mockPending:    make(map[common.Address][]*PendingUnstake),
		mockIdentities: make(map[common.Address]*NodeIdentity),
		mockNodeIDMap:  make(map[[32]byte]common.Address),
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

	logging.Info("staked tokens", "amount", amount.String(), "provider", addr.Hex(), "total_stake", sc.mockStakes[addr].String())

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
	sc.mockPending[addr] = append(sc.mockPending[addr], &PendingUnstake{
		Amount:     new(big.Int).Set(amount),
		UnlockTime: time.Now().Add(7 * 24 * time.Hour),
	})
	sc.mockStakes[addr] = new(big.Int).Sub(current, amount)

	logging.Info("requested unstake", "amount", amount.String(), "provider", addr.Hex(),
		"queue_index", len(sc.mockPending[addr])-1)

	return nil, nil
}

// CompleteUnstake completes a pending unstake request by queue index
func (sc *StakingContract) CompleteUnstake(ctx context.Context, requestIndex *big.Int) (*types.Transaction, error) {
	if sc.mockMode {
		return sc.mockCompleteUnstake(ctx, requestIndex)
	}

	auth, err := sc.baseClient.GetTransactOpts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction options: %w", err)
	}

	tx, err := sc.contract.Transact(auth, "completeUnstake", requestIndex)
	if err != nil {
		return nil, fmt.Errorf("failed to complete unstake: %w", err)
	}

	return tx, nil
}

// mockCompleteUnstake handles unstake completion in mock mode
func (sc *StakingContract) mockCompleteUnstake(_ context.Context, requestIndex *big.Int) (*types.Transaction, error) {
	sc.mockMu.Lock()
	defer sc.mockMu.Unlock()

	var addr common.Address
	if sc.baseClient != nil {
		addr = sc.baseClient.Address()
	}

	queue, exists := sc.mockPending[addr]
	if !exists {
		return nil, fmt.Errorf("no pending unstake requests")
	}

	idx := int(requestIndex.Int64())
	if idx >= len(queue) {
		return nil, fmt.Errorf("invalid request index: %d", idx)
	}

	req := queue[idx]
	if req.Completed {
		return nil, fmt.Errorf("unstake request already completed")
	}
	if time.Now().Before(req.UnlockTime) {
		return nil, fmt.Errorf("cooldown not complete")
	}

	req.Completed = true
	logging.Info("completed unstake", "amount", req.Amount.String(), "provider", addr.Hex(), "index", idx)

	return nil, nil
}

// Withdraw is kept for backwards compatibility, completes the first pending unstake
// Deprecated: Use CompleteUnstake with explicit request index
func (sc *StakingContract) Withdraw(ctx context.Context) (*types.Transaction, error) {
	return sc.CompleteUnstake(ctx, big.NewInt(0))
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

// IsActiveProvider checks if a provider is active (staked and registered)
func (sc *StakingContract) IsActiveProvider(ctx context.Context, provider common.Address) (bool, error) {
	if sc.mockMode {
		sc.mockMu.RLock()
		defer sc.mockMu.RUnlock()
		stake, exists := sc.mockStakes[provider]
		if !exists {
			return false, nil
		}
		minStake := parseWei("1000000") // Starter tier minimum (1,000,000 BUNKER)
		return stake.Cmp(minStake) >= 0, nil
	}

	var result []interface{}
	err := sc.contract.Call(&bind.CallOpts{Context: ctx}, &result, "isActiveProvider", provider)
	if err != nil {
		return false, fmt.Errorf("failed to check active provider: %w", err)
	}

	if len(result) == 0 {
		return false, nil
	}
	if active, ok := result[0].(bool); ok {
		return active, nil
	}
	return false, nil
}

// GetProviderInfo returns provider information from the contract
func (sc *StakingContract) GetProviderInfo(ctx context.Context, provider common.Address) (*ProviderInfoData, error) {
	if sc.mockMode {
		sc.mockMu.RLock()
		defer sc.mockMu.RUnlock()

		stake, _ := sc.mockStakes[provider]
		if stake == nil {
			stake = big.NewInt(0)
		}

		totalUnbonding := big.NewInt(0)
		if queue, exists := sc.mockPending[provider]; exists {
			for _, req := range queue {
				if !req.Completed {
					totalUnbonding.Add(totalUnbonding, req.Amount)
				}
			}
		}

		minStake := parseWei("1000000") // Starter tier minimum (1,000,000 BUNKER)
		info := &ProviderInfoData{
			StakedAmount:   new(big.Int).Set(stake),
			TotalUnbonding: totalUnbonding,
			Active:         stake.Cmp(minStake) >= 0,
		}

		if identity, exists := sc.mockIdentities[provider]; exists {
			info.NodeID = identity.NodeID
			info.Region = identity.Region
			info.Capabilities = identity.Capabilities
		}

		return info, nil
	}

	var result []interface{}
	err := sc.contract.Call(&bind.CallOpts{Context: ctx}, &result, "getProviderInfo", provider)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider info: %w", err)
	}

	if len(result) == 0 {
		return &ProviderInfoData{
			StakedAmount:   big.NewInt(0),
			TotalUnbonding: big.NewInt(0),
		}, nil
	}

	info := &ProviderInfoData{
		StakedAmount:   big.NewInt(0),
		TotalUnbonding: big.NewInt(0),
	}

	// go-ethereum unpacks the full struct including identity fields
	if res, ok := result[0].(struct {
		StakedAmount   *big.Int
		TotalUnbonding *big.Int
		Beneficiary    common.Address
		RegisteredAt   *big.Int
		Active         bool
		NodeId         [32]byte
		Region         [32]byte
		Capabilities   uint64
		Frozen         bool
	}); ok {
		info.StakedAmount = res.StakedAmount
		info.TotalUnbonding = res.TotalUnbonding
		info.Beneficiary = res.Beneficiary
		if res.RegisteredAt != nil {
			info.RegisteredAt = time.Unix(res.RegisteredAt.Int64(), 0)
		}
		info.Active = res.Active
		info.NodeID = res.NodeId
		info.Region = res.Region
		info.Capabilities = res.Capabilities
		info.Frozen = res.Frozen
	}

	return info, nil
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
		if queue, exists := sc.mockPending[provider]; exists {
			for _, req := range queue {
				if !req.Completed {
					info.PendingUnstake.Add(info.PendingUnstake, req.Amount)
					// Use the earliest unlock time
					if info.UnlockTime.IsZero() || req.UnlockTime.Before(info.UnlockTime) {
						info.UnlockTime = req.UnlockTime
					}
				}
			}
		}
		sc.mockMu.RUnlock()
	} else {
		// Get total unbonding from provider info
		providerInfo, err := sc.GetProviderInfo(ctx, provider)
		if err == nil && providerInfo != nil {
			info.PendingUnstake = providerInfo.TotalUnbonding
		}
	}

	// Determine tier based on effective stake (personal + delegated)
	effectiveStake := new(big.Int).Set(stake)
	if sc.delegationContract != nil {
		if delegated, err := sc.delegationContract.GetTotalDelegatedTo(ctx, provider); err == nil && delegated != nil {
			effectiveStake.Add(effectiveStake, delegated)
		}
	}
	info.Tier = sc.determineTier(effectiveStake)

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
		"",                           // 0 = none
		pkgtypes.StakingTierStarter,  // 1
		pkgtypes.StakingTierBronze,   // 2
		pkgtypes.StakingTierSilver,   // 3
		pkgtypes.StakingTierGold,     // 4
		pkgtypes.StakingTierPlatinum, // 5
	}
	if int(index) < len(tiers) {
		return tiers[index]
	}
	return ""
}

// determineTier determines the staking tier based on stake amount
func (sc *StakingContract) determineTier(stake *big.Int) pkgtypes.StakingTier {
	// Tier thresholds match BunkerStaking.sol
	tiers := []struct {
		tier   pkgtypes.StakingTier
		minWei *big.Int
	}{
		{pkgtypes.StakingTierPlatinum, parseWei("1000000000")}, // 1,000,000,000 BUNKER
		{pkgtypes.StakingTierGold, parseWei("100000000")},     // 100,000,000 BUNKER
		{pkgtypes.StakingTierSilver, parseWei("10000000")},    // 10,000,000 BUNKER
		{pkgtypes.StakingTierBronze, parseWei("5000000")},     // 5,000,000 BUNKER
		{pkgtypes.StakingTierStarter, parseWei("1000000")},    // 1,000,000 BUNKER
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
					logging.Error("stake event subscription error", "error", err)
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

	// Parse data (amount, totalStake, tier)
	if len(log.Data) >= 64 {
		event.Amount = new(big.Int).SetBytes(log.Data[:32])
		event.TotalStake = new(big.Int).SetBytes(log.Data[32:64])
	}

	return event, nil
}

// StakeWithIdentity stakes tokens with an on-chain NodeID binding.
func (sc *StakingContract) StakeWithIdentity(ctx context.Context, amount *big.Int, nodeID [32]byte, region [32]byte, capabilities uint64) (*types.Transaction, error) {
	if sc.mockMode {
		// First, do the normal stake
		tx, err := sc.mockStake(ctx, amount)
		if err != nil {
			return nil, err
		}
		// Then record identity
		sc.mockMu.Lock()
		var addr common.Address
		if sc.baseClient != nil {
			addr = sc.baseClient.Address()
		}
		sc.mockIdentities[addr] = &NodeIdentity{
			NodeID:       nodeID,
			Region:       region,
			Capabilities: capabilities,
		}
		sc.mockNodeIDMap[nodeID] = addr
		sc.mockMu.Unlock()

		logging.Info("staked with identity",
			"provider", addr.Hex(),
			"node_id", fmt.Sprintf("%x", nodeID[:8]),
			"amount", amount.String())
		return tx, nil
	}

	// Approve token spend
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

	tx, err := sc.contract.Transact(auth, "stakeWithIdentity", amount, nodeID, region, capabilities)
	if err != nil {
		return nil, fmt.Errorf("failed to stake with identity: %w", err)
	}

	return tx, nil
}

// UpdateIdentity updates a provider's on-chain NodeID, region, and capabilities.
func (sc *StakingContract) UpdateIdentity(ctx context.Context, nodeID [32]byte, region [32]byte, capabilities uint64) (*types.Transaction, error) {
	if sc.mockMode {
		sc.mockMu.Lock()
		defer sc.mockMu.Unlock()

		var addr common.Address
		if sc.baseClient != nil {
			addr = sc.baseClient.Address()
		}

		// Remove old nodeID mapping if exists
		if old, exists := sc.mockIdentities[addr]; exists {
			delete(sc.mockNodeIDMap, old.NodeID)
		}

		sc.mockIdentities[addr] = &NodeIdentity{
			NodeID:       nodeID,
			Region:       region,
			Capabilities: capabilities,
		}
		sc.mockNodeIDMap[nodeID] = addr

		logging.Info("updated identity",
			"provider", addr.Hex(),
			"node_id", fmt.Sprintf("%x", nodeID[:8]))
		return nil, nil
	}

	auth, err := sc.baseClient.GetTransactOpts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction options: %w", err)
	}

	tx, err := sc.contract.Transact(auth, "updateIdentity", nodeID, region, capabilities)
	if err != nil {
		return nil, fmt.Errorf("failed to update identity: %w", err)
	}

	return tx, nil
}

// NodeIDToProvider returns the provider address registered for a given NodeID.
// Returns the zero address if not found.
func (sc *StakingContract) NodeIDToProvider(ctx context.Context, nodeID [32]byte) (common.Address, error) {
	if sc.mockMode {
		sc.mockMu.RLock()
		defer sc.mockMu.RUnlock()
		if addr, exists := sc.mockNodeIDMap[nodeID]; exists {
			return addr, nil
		}
		return common.Address{}, nil
	}

	var result []interface{}
	err := sc.contract.Call(&bind.CallOpts{Context: ctx}, &result, "nodeIdToProvider", nodeID)
	if err != nil {
		return common.Address{}, fmt.Errorf("failed to lookup nodeID: %w", err)
	}

	if len(result) == 0 {
		return common.Address{}, nil
	}
	if addr, ok := result[0].(common.Address); ok {
		return addr, nil
	}
	return common.Address{}, nil
}

// HasMinimumStake checks if provider has minimum stake (1,000,000 BUNKER = Starter tier)
func (sc *StakingContract) HasMinimumStake(ctx context.Context, provider common.Address) (bool, error) {
	stake, err := sc.GetStake(ctx, provider)
	if err != nil {
		return false, err
	}

	minStake := parseWei("1000000") // 1,000,000 BUNKER minimum (Starter tier)
	return stake.Cmp(minStake) >= 0, nil
}

// ─── Admin Setters ───────────────────────────────────────────────────────────

// SetTierRewardMultiplier adjusts the reward multiplier for a staking tier.
func (sc *StakingContract) SetTierRewardMultiplier(ctx context.Context, tier uint8, multiplierBps uint16) (*types.Transaction, error) {
	if sc.mockMode {
		logging.Info("mock: setTierRewardMultiplier tier=%d bps=%d", tier, multiplierBps)
		return nil, nil
	}
	auth, err := sc.baseClient.GetTransactOpts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction options: %w", err)
	}
	tx, err := sc.contract.Transact(auth, "setTierRewardMultiplier", tier, multiplierBps)
	if err != nil {
		return nil, fmt.Errorf("failed to set tier reward multiplier: %w", err)
	}
	return tx, nil
}

// SetMaxTierMultiplierBps adjusts the maximum tier multiplier cap.
func (sc *StakingContract) SetMaxTierMultiplierBps(ctx context.Context, newMax uint16) (*types.Transaction, error) {
	if sc.mockMode {
		logging.Info("mock: setMaxTierMultiplierBps newMax=%d", newMax)
		return nil, nil
	}
	auth, err := sc.baseClient.GetTransactOpts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction options: %w", err)
	}
	tx, err := sc.contract.Transact(auth, "setMaxTierMultiplierBps", newMax)
	if err != nil {
		return nil, fmt.Errorf("failed to set max tier multiplier: %w", err)
	}
	return tx, nil
}

// SetUnbondingPeriod adjusts the unstaking unbonding period.
func (sc *StakingContract) SetUnbondingPeriod(ctx context.Context, newPeriod *big.Int) (*types.Transaction, error) {
	if sc.mockMode {
		logging.Info("mock: setUnbondingPeriod period=%s", newPeriod)
		return nil, nil
	}
	auth, err := sc.baseClient.GetTransactOpts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction options: %w", err)
	}
	tx, err := sc.contract.Transact(auth, "setUnbondingPeriod", newPeriod)
	if err != nil {
		return nil, fmt.Errorf("failed to set unbonding period: %w", err)
	}
	return tx, nil
}

// SetSlashFeeSplit adjusts the burn/treasury split for slashed tokens.
func (sc *StakingContract) SetSlashFeeSplit(ctx context.Context, burnBps, treasuryBps uint16) (*types.Transaction, error) {
	if sc.mockMode {
		logging.Info("mock: setSlashFeeSplit burn=%d treasury=%d", burnBps, treasuryBps)
		return nil, nil
	}
	auth, err := sc.baseClient.GetTransactOpts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction options: %w", err)
	}
	tx, err := sc.contract.Transact(auth, "setSlashFeeSplit", burnBps, treasuryBps)
	if err != nil {
		return nil, fmt.Errorf("failed to set slash fee split: %w", err)
	}
	return tx, nil
}

// SetAppealWindow adjusts the slash appeal window duration.
func (sc *StakingContract) SetAppealWindow(ctx context.Context, newWindow *big.Int) (*types.Transaction, error) {
	if sc.mockMode {
		logging.Info("mock: setAppealWindow window=%s", newWindow)
		return nil, nil
	}
	auth, err := sc.baseClient.GetTransactOpts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction options: %w", err)
	}
	tx, err := sc.contract.Transact(auth, "setAppealWindow", newWindow)
	if err != nil {
		return nil, fmt.Errorf("failed to set appeal window: %w", err)
	}
	return tx, nil
}

// SetSlashingEnabled enables or disables slashing execution.
// When disabled, proposals can still be created (monitor mode) but execution reverts.
func (sc *StakingContract) SetSlashingEnabled(ctx context.Context, enabled bool) (*types.Transaction, error) {
	if sc.mockMode {
		logging.Info("mock: setSlashingEnabled enabled=%v", enabled)
		return nil, nil
	}
	auth, err := sc.baseClient.GetTransactOpts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction options: %w", err)
	}
	tx, err := sc.contract.Transact(auth, "setSlashingEnabled", enabled)
	if err != nil {
		return nil, fmt.Errorf("failed to set slashing enabled: %w", err)
	}
	return tx, nil
}

// SlashingEnabled returns whether slashing execution is currently enabled.
func (sc *StakingContract) SlashingEnabled(ctx context.Context) (bool, error) {
	if sc.mockMode {
		return false, nil
	}
	var result []interface{}
	err := sc.contract.Call(&bind.CallOpts{Context: ctx}, &result, "slashingEnabled")
	if err != nil {
		return false, fmt.Errorf("failed to query slashing enabled: %w", err)
	}
	return result[0].(bool), nil
}
