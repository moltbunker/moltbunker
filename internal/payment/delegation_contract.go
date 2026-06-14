package payment

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/moltbunker/moltbunker/internal/logging"
	"github.com/moltbunker/moltbunker/internal/payment/bindings"
)

// DelegationContract provides interface to the BunkerDelegation smart contract.
type DelegationContract struct {
	baseClient   *BaseClient
	contract     *bind.BoundContract
	contractABI  abi.ABI
	contractAddr common.Address
	mockMode     bool

	// caller is the abigen-generated, type-safe read binding (ABIGEN-01 Phase 1).
	// Read paths (GetDelegation, GetProviderConfig) decode through this instead of
	// the hand-typed DelegationContractABI + fragile struct assertions. The
	// contract/contractABI fields above are still used by the Transact methods
	// until Phase 2 migrates them to the generated Transactor.
	// TODO(ABIGEN-01-phase2): move Transact methods onto the generated binding and
	// drop the hand-typed DelegationContractABI + the dual ABI load.
	caller *bindings.BunkerDelegationCaller

	// Mock state
	mockDelegations    map[common.Address]*DelegationData
	mockProviderConfig map[common.Address]*ProviderDelegationConfigData
	mockTotalDelegated map[common.Address]*big.Int
	mockUnbonding      map[common.Address][]*UnbondingRequestData
	mockMu             sync.RWMutex
}

// NewDelegationContract creates a new delegation contract client.
func NewDelegationContract(baseClient *BaseClient, contractAddr common.Address) (*DelegationContract, error) {
	dc := &DelegationContract{
		baseClient:         baseClient,
		contractAddr:       contractAddr,
		mockDelegations:    make(map[common.Address]*DelegationData),
		mockProviderConfig: make(map[common.Address]*ProviderDelegationConfigData),
		mockTotalDelegated: make(map[common.Address]*big.Int),
		mockUnbonding:      make(map[common.Address][]*UnbondingRequestData),
	}

	// Require a connected base client; use NewMockDelegationContract() for testing
	if baseClient == nil {
		return nil, fmt.Errorf("base client is required (use NewMockDelegationContract for testing)")
	}
	if !baseClient.IsConnected() {
		return nil, fmt.Errorf("base client not connected to RPC")
	}

	parsedABI, err := abi.JSON(strings.NewReader(DelegationContractABI))
	if err != nil {
		return nil, fmt.Errorf("failed to parse delegation ABI: %w", err)
	}
	dc.contractABI = parsedABI

	client := baseClient.Client()
	dc.contract = bind.NewBoundContract(contractAddr, parsedABI, client, client, client)

	// ABIGEN-01 Phase 1: build the generated, type-safe read binding. Its ABI
	// is sourced from contracts/out/BunkerDelegation.sol (canonical), so the
	// getDelegation/getProviderConfig tuple shapes always match the deployed
	// contract — closing the silent-decode-mismatch class of bugs that PAY-01
	// patched by hand.
	caller, err := bindings.NewBunkerDelegationCaller(contractAddr, client)
	if err != nil {
		return nil, fmt.Errorf("failed to build delegation caller binding: %w", err)
	}
	dc.caller = caller

	return dc, nil
}

// NewMockDelegationContract creates a mock delegation contract for testing.
func NewMockDelegationContract() *DelegationContract {
	return &DelegationContract{
		mockMode:           true,
		mockDelegations:    make(map[common.Address]*DelegationData),
		mockProviderConfig: make(map[common.Address]*ProviderDelegationConfigData),
		mockTotalDelegated: make(map[common.Address]*big.Int),
		mockUnbonding:      make(map[common.Address][]*UnbondingRequestData),
	}
}

// IsMockMode returns whether running in mock mode.
func (dc *DelegationContract) IsMockMode() bool {
	return dc.mockMode
}

// Delegate delegates tokens to a provider.
func (dc *DelegationContract) Delegate(ctx context.Context, provider common.Address, amount *big.Int) (*types.Transaction, error) {
	if dc.mockMode {
		return dc.mockDelegate(ctx, provider, amount)
	}

	auth, err := dc.baseClient.GetTransactOpts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction options: %w", err)
	}

	tx, err := dc.contract.Transact(auth, "delegate", provider, amount)
	if err != nil {
		return nil, fmt.Errorf("failed to delegate: %w", err)
	}
	return tx, nil
}

func (dc *DelegationContract) mockDelegate(_ context.Context, provider common.Address, amount *big.Int) (*types.Transaction, error) {
	dc.mockMu.Lock()
	defer dc.mockMu.Unlock()

	var delegator common.Address
	if dc.baseClient != nil {
		delegator = dc.baseClient.Address()
	}

	dc.mockDelegations[delegator] = &DelegationData{
		Provider:    provider,
		Amount:      new(big.Int).Set(amount),
		DelegatedAt: time.Now(),
		Active:      true,
	}

	total, exists := dc.mockTotalDelegated[provider]
	if !exists {
		total = big.NewInt(0)
	}
	dc.mockTotalDelegated[provider] = new(big.Int).Add(total, amount)

	logging.Info("delegated tokens", "provider", provider.Hex(), "amount", amount.String())
	return nil, nil
}

// RequestUndelegate requests to undelegate tokens (starts unbonding).
func (dc *DelegationContract) RequestUndelegate(ctx context.Context, amount *big.Int) (*types.Transaction, error) {
	if dc.mockMode {
		return dc.mockRequestUndelegate(ctx, amount)
	}

	auth, err := dc.baseClient.GetTransactOpts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction options: %w", err)
	}

	tx, err := dc.contract.Transact(auth, "requestUndelegate", amount)
	if err != nil {
		return nil, fmt.Errorf("failed to request undelegate: %w", err)
	}
	return tx, nil
}

func (dc *DelegationContract) mockRequestUndelegate(_ context.Context, amount *big.Int) (*types.Transaction, error) {
	dc.mockMu.Lock()
	defer dc.mockMu.Unlock()

	var delegator common.Address
	if dc.baseClient != nil {
		delegator = dc.baseClient.Address()
	}

	del, exists := dc.mockDelegations[delegator]
	if !exists || del.Amount.Cmp(amount) < 0 {
		return nil, fmt.Errorf("insufficient delegation")
	}

	del.Amount.Sub(del.Amount, amount)

	dc.mockUnbonding[delegator] = append(dc.mockUnbonding[delegator], &UnbondingRequestData{
		Amount:     new(big.Int).Set(amount),
		UnlockTime: time.Now().Add(7 * 24 * time.Hour),
	})

	logging.Info("requested undelegate", "amount", amount.String())
	return nil, nil
}

// CompleteUndelegate completes a pending undelegation.
func (dc *DelegationContract) CompleteUndelegate(ctx context.Context, index *big.Int) (*types.Transaction, error) {
	if dc.mockMode {
		return dc.mockCompleteUndelegate(ctx, index)
	}

	auth, err := dc.baseClient.GetTransactOpts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction options: %w", err)
	}

	tx, err := dc.contract.Transact(auth, "completeUndelegate", index)
	if err != nil {
		return nil, fmt.Errorf("failed to complete undelegate: %w", err)
	}
	return tx, nil
}

func (dc *DelegationContract) mockCompleteUndelegate(_ context.Context, index *big.Int) (*types.Transaction, error) {
	dc.mockMu.Lock()
	defer dc.mockMu.Unlock()

	var delegator common.Address
	if dc.baseClient != nil {
		delegator = dc.baseClient.Address()
	}

	queue := dc.mockUnbonding[delegator]
	idx := int(index.Int64())
	if idx >= len(queue) {
		return nil, fmt.Errorf("invalid unbonding index: %d", idx)
	}

	req := queue[idx]
	if req.Completed {
		return nil, fmt.Errorf("already completed")
	}
	if time.Now().Before(req.UnlockTime) {
		return nil, fmt.Errorf("unbonding period not complete")
	}
	req.Completed = true

	logging.Info("completed undelegate", "amount", req.Amount.String(), "index", idx)
	return nil, nil
}

// GetDelegation returns the delegation data for a delegator.
func (dc *DelegationContract) GetDelegation(ctx context.Context, delegator common.Address) (*DelegationData, error) {
	if dc.mockMode {
		dc.mockMu.RLock()
		defer dc.mockMu.RUnlock()
		del, exists := dc.mockDelegations[delegator]
		if !exists {
			return &DelegationData{Amount: big.NewInt(0)}, nil
		}
		return &DelegationData{
			Provider:    del.Provider,
			Amount:      new(big.Int).Set(del.Amount),
			DelegatedAt: del.DelegatedAt,
			Active:      del.Active,
		}, nil
	}

	// ABIGEN-01 Phase 1: decode through the generated, type-safe caller. The
	// returned BunkerDelegationDelegationInfo mirrors the on-chain DelegationInfo
	// struct exactly (provider address, amount uint128->*big.Int,
	// delegatedAt uint48->*big.Int, active bool), so there is no positional
	// struct-assertion to silently mismatch.
	info, err := dc.caller.GetDelegation(&bind.CallOpts{Context: ctx}, delegator)
	if err != nil {
		return nil, fmt.Errorf("failed to get delegation: %w", err)
	}

	data := &DelegationData{
		Provider: info.Provider,
		Amount:   big.NewInt(0),
		Active:   info.Active,
	}
	if info.Amount != nil {
		data.Amount = info.Amount
	}
	if info.DelegatedAt != nil {
		data.DelegatedAt = time.Unix(info.DelegatedAt.Int64(), 0)
	}
	return data, nil
}

// GetProviderConfig returns a provider's delegation configuration.
func (dc *DelegationContract) GetProviderConfig(ctx context.Context, provider common.Address) (*ProviderDelegationConfigData, error) {
	if dc.mockMode {
		dc.mockMu.RLock()
		defer dc.mockMu.RUnlock()
		cfg, exists := dc.mockProviderConfig[provider]
		if !exists {
			return &ProviderDelegationConfigData{
				RewardCutBps:      1000, // 10% default
				AcceptDelegations: true,
			}, nil
		}
		return &ProviderDelegationConfigData{
			RewardCutBps:         cfg.RewardCutBps,
			FeeShareBps:          cfg.FeeShareBps,
			AcceptDelegations:    cfg.AcceptDelegations,
			PendingRewardCutBps:  cfg.PendingRewardCutBps,
			RewardCutEffectiveAt: cfg.RewardCutEffectiveAt,
		}, nil
	}

	// ABIGEN-01 Phase 1: decode through the generated caller. The on-chain
	// ProviderDelegationConfig tuple is
	//   {rewardCutBps, pendingRewardCutBps, rewardCutEffectiveAt, feeShareBps,
	//    totalDelegated, acceptingDelegations}
	// — which the hand-typed DelegationContractABI got wrong (only five fields,
	// in the wrong order, with acceptDelegations instead of acceptingDelegations
	// and no totalDelegated). The generated binding always matches the deployed
	// contract.
	//
	// NOTE: cfg.TotalDelegated is decoded by the binding but not carried on the
	// domain ProviderDelegationConfigData yet; GetTotalDelegatedTo already serves
	// that value. Surfacing it here (and feeding determineTier in
	// staking_contract.go) is the C4 follow-up.
	cfg, err := dc.caller.GetProviderConfig(&bind.CallOpts{Context: ctx}, provider)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider config: %w", err)
	}

	data := &ProviderDelegationConfigData{
		RewardCutBps:        cfg.RewardCutBps,
		FeeShareBps:         cfg.FeeShareBps,
		AcceptDelegations:   cfg.AcceptingDelegations,
		PendingRewardCutBps: cfg.PendingRewardCutBps,
	}
	if cfg.RewardCutEffectiveAt != nil {
		data.RewardCutEffectiveAt = time.Unix(cfg.RewardCutEffectiveAt.Int64(), 0)
	}
	return data, nil
}

// GetTotalDelegatedTo returns the total amount delegated to a provider.
func (dc *DelegationContract) GetTotalDelegatedTo(ctx context.Context, provider common.Address) (*big.Int, error) {
	if dc.mockMode {
		dc.mockMu.RLock()
		defer dc.mockMu.RUnlock()
		total, exists := dc.mockTotalDelegated[provider]
		if !exists {
			return big.NewInt(0), nil
		}
		return new(big.Int).Set(total), nil
	}

	var result []interface{}
	err := dc.contract.Call(&bind.CallOpts{Context: ctx}, &result, "getTotalDelegatedTo", provider)
	if err != nil {
		return nil, fmt.Errorf("failed to get total delegated: %w", err)
	}

	if len(result) > 0 {
		if total, ok := result[0].(*big.Int); ok {
			return total, nil
		}
	}
	return big.NewInt(0), nil
}

// SetDelegationConfig updates the provider's delegation configuration.
func (dc *DelegationContract) SetDelegationConfig(ctx context.Context, rewardCutBps, feeShareBps uint16) (*types.Transaction, error) {
	if dc.mockMode {
		dc.mockMu.Lock()
		defer dc.mockMu.Unlock()

		var addr common.Address
		if dc.baseClient != nil {
			addr = dc.baseClient.Address()
		}
		dc.mockProviderConfig[addr] = &ProviderDelegationConfigData{
			RewardCutBps:      rewardCutBps,
			FeeShareBps:       feeShareBps,
			AcceptDelegations: true,
		}
		logging.Info("set delegation config", "rewardCutBps", rewardCutBps, "feeShareBps", feeShareBps)
		return nil, nil
	}

	auth, err := dc.baseClient.GetTransactOpts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction options: %w", err)
	}

	tx, err := dc.contract.Transact(auth, "setDelegationConfig", rewardCutBps, feeShareBps)
	if err != nil {
		return nil, fmt.Errorf("failed to set delegation config: %w", err)
	}
	return tx, nil
}

// ToggleAcceptDelegations toggles whether the provider accepts delegations.
func (dc *DelegationContract) ToggleAcceptDelegations(ctx context.Context, accept bool) (*types.Transaction, error) {
	if dc.mockMode {
		dc.mockMu.Lock()
		defer dc.mockMu.Unlock()

		var addr common.Address
		if dc.baseClient != nil {
			addr = dc.baseClient.Address()
		}
		cfg, exists := dc.mockProviderConfig[addr]
		if !exists {
			cfg = &ProviderDelegationConfigData{RewardCutBps: 1000}
			dc.mockProviderConfig[addr] = cfg
		}
		cfg.AcceptDelegations = accept
		logging.Info("toggled accept delegations", "accept", accept)
		return nil, nil
	}

	auth, err := dc.baseClient.GetTransactOpts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction options: %w", err)
	}

	tx, err := dc.contract.Transact(auth, "toggleAcceptDelegations", accept)
	if err != nil {
		return nil, fmt.Errorf("failed to toggle delegations: %w", err)
	}
	return tx, nil
}

// ─── Admin Setters ───────────────────────────────────────────────────────────

// SetUnbondingPeriod adjusts the delegation unbonding period.
func (dc *DelegationContract) SetUnbondingPeriod(ctx context.Context, newPeriod *big.Int) (*types.Transaction, error) {
	if dc.mockMode {
		logging.Info("mock: delegation setUnbondingPeriod period=%s", newPeriod)
		return nil, nil
	}
	auth, err := dc.baseClient.GetTransactOpts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction options: %w", err)
	}
	tx, err := dc.contract.Transact(auth, "setUnbondingPeriod", newPeriod)
	if err != nil {
		return nil, fmt.Errorf("failed to set delegation unbonding period: %w", err)
	}
	return tx, nil
}

// SetMaxRewardCutBps adjusts the max reward cut providers can charge delegators.
func (dc *DelegationContract) SetMaxRewardCutBps(ctx context.Context, newMax uint16) (*types.Transaction, error) {
	if dc.mockMode {
		logging.Info("mock: delegation setMaxRewardCutBps max=%d", newMax)
		return nil, nil
	}
	auth, err := dc.baseClient.GetTransactOpts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction options: %w", err)
	}
	tx, err := dc.contract.Transact(auth, "setMaxRewardCutBps", newMax)
	if err != nil {
		return nil, fmt.Errorf("failed to set max reward cut: %w", err)
	}
	return tx, nil
}
