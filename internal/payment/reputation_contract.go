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
)

const (
	reputationInitialScore  = 500
	reputationMaxScore      = 1000
	reputationMinScoreJobs  = 250
	reputationMaxDelta      = 200
)

// ReputationContract provides interface to the BunkerReputation smart contract.
type ReputationContract struct {
	baseClient   *BaseClient
	contract     *bind.BoundContract
	contractABI  abi.ABI
	contractAddr common.Address
	mockMode     bool

	// Mock state
	mockReputations map[common.Address]*ReputationDataOnChain
	mockMu          sync.RWMutex
}

// NewReputationContract creates a new reputation contract client.
func NewReputationContract(baseClient *BaseClient, contractAddr common.Address) (*ReputationContract, error) {
	rc := &ReputationContract{
		baseClient:      baseClient,
		contractAddr:    contractAddr,
		mockReputations: make(map[common.Address]*ReputationDataOnChain),
	}

	// Require a connected base client; use NewMockReputationContract() for testing
	if baseClient == nil {
		return nil, fmt.Errorf("base client is required (use NewMockReputationContract for testing)")
	}
	if !baseClient.IsConnected() {
		return nil, fmt.Errorf("base client not connected to RPC")
	}

	parsedABI, err := abi.JSON(strings.NewReader(ReputationContractABI))
	if err != nil {
		return nil, fmt.Errorf("failed to parse reputation ABI: %w", err)
	}
	rc.contractABI = parsedABI

	client := baseClient.Client()
	rc.contract = bind.NewBoundContract(contractAddr, parsedABI, client, client, client)

	return rc, nil
}

// NewMockReputationContract creates a mock reputation contract for testing.
func NewMockReputationContract() *ReputationContract {
	return &ReputationContract{
		mockMode:        true,
		mockReputations: make(map[common.Address]*ReputationDataOnChain),
	}
}

// IsMockMode returns whether running in mock mode.
func (rc *ReputationContract) IsMockMode() bool {
	return rc.mockMode
}

// ensureRegistered initializes a mock reputation entry (must hold write lock).
func (rc *ReputationContract) ensureRegistered(provider common.Address) *ReputationDataOnChain {
	rep, exists := rc.mockReputations[provider]
	if !exists {
		rep = &ReputationDataOnChain{
			Score:         big.NewInt(reputationInitialScore),
			JobsCompleted: big.NewInt(0),
			JobsFailed:    big.NewInt(0),
			SlashEvents:   big.NewInt(0),
			LastDecay:     time.Now(),
			Registered:    true,
		}
		rc.mockReputations[provider] = rep
	}
	return rep
}

// GetScore returns the reputation score for a provider.
func (rc *ReputationContract) GetScore(ctx context.Context, provider common.Address) (*big.Int, error) {
	if rc.mockMode {
		rc.mockMu.RLock()
		defer rc.mockMu.RUnlock()
		rep, exists := rc.mockReputations[provider]
		if !exists {
			return big.NewInt(0), nil
		}
		return new(big.Int).Set(rep.Score), nil
	}

	var result []interface{}
	err := rc.contract.Call(&bind.CallOpts{Context: ctx}, &result, "getScore", provider)
	if err != nil {
		return nil, fmt.Errorf("failed to get score: %w", err)
	}
	if len(result) > 0 {
		if score, ok := result[0].(*big.Int); ok {
			return score, nil
		}
	}
	return big.NewInt(0), nil
}

// GetTier returns the reputation tier for a provider.
func (rc *ReputationContract) GetTier(ctx context.Context, provider common.Address) (ReputationTier, error) {
	if rc.mockMode {
		rc.mockMu.RLock()
		defer rc.mockMu.RUnlock()
		rep, exists := rc.mockReputations[provider]
		if !exists {
			return ReputationTierUntrusted, nil
		}
		return rc.tierFromScore(rep.Score), nil
	}

	var result []interface{}
	err := rc.contract.Call(&bind.CallOpts{Context: ctx}, &result, "getTier", provider)
	if err != nil {
		return ReputationTierUntrusted, fmt.Errorf("failed to get tier: %w", err)
	}
	if len(result) > 0 {
		if tier, ok := result[0].(uint8); ok {
			return ReputationTier(tier), nil
		}
	}
	return ReputationTierUntrusted, nil
}

// IsEligibleForJobs checks if a provider is eligible for job assignment.
func (rc *ReputationContract) IsEligibleForJobs(ctx context.Context, provider common.Address) (bool, error) {
	if rc.mockMode {
		rc.mockMu.RLock()
		defer rc.mockMu.RUnlock()
		rep, exists := rc.mockReputations[provider]
		if !exists {
			return false, nil
		}
		return rep.Score.Cmp(big.NewInt(reputationMinScoreJobs)) >= 0, nil
	}

	var result []interface{}
	err := rc.contract.Call(&bind.CallOpts{Context: ctx}, &result, "isEligibleForJobs", provider)
	if err != nil {
		return false, fmt.Errorf("failed to check eligibility: %w", err)
	}
	if len(result) > 0 {
		if eligible, ok := result[0].(bool); ok {
			return eligible, nil
		}
	}
	return false, nil
}

// GetReputation returns full reputation data for a provider.
func (rc *ReputationContract) GetReputation(ctx context.Context, provider common.Address) (*ReputationDataOnChain, error) {
	if rc.mockMode {
		rc.mockMu.RLock()
		defer rc.mockMu.RUnlock()
		rep, exists := rc.mockReputations[provider]
		if !exists {
			return &ReputationDataOnChain{
				Score:         big.NewInt(0),
				JobsCompleted: big.NewInt(0),
				JobsFailed:    big.NewInt(0),
				SlashEvents:   big.NewInt(0),
			}, nil
		}
		return &ReputationDataOnChain{
			Score:         new(big.Int).Set(rep.Score),
			JobsCompleted: new(big.Int).Set(rep.JobsCompleted),
			JobsFailed:    new(big.Int).Set(rep.JobsFailed),
			SlashEvents:   new(big.Int).Set(rep.SlashEvents),
			LastDecay:     rep.LastDecay,
			Registered:    rep.Registered,
		}, nil
	}

	var result []interface{}
	err := rc.contract.Call(&bind.CallOpts{Context: ctx}, &result, "getReputation", provider)
	if err != nil {
		return nil, fmt.Errorf("failed to get reputation: %w", err)
	}

	data := &ReputationDataOnChain{
		Score:         big.NewInt(0),
		JobsCompleted: big.NewInt(0),
		JobsFailed:    big.NewInt(0),
		SlashEvents:   big.NewInt(0),
	}
	if len(result) > 0 {
		if res, ok := result[0].(struct {
			Score         *big.Int
			JobsCompleted *big.Int
			JobsFailed    *big.Int
			SlashEvents   *big.Int
			LastDecay     *big.Int
			Registered    bool
		}); ok {
			data.Score = res.Score
			data.JobsCompleted = res.JobsCompleted
			data.JobsFailed = res.JobsFailed
			data.SlashEvents = res.SlashEvents
			if res.LastDecay != nil {
				data.LastDecay = time.Unix(res.LastDecay.Int64(), 0)
			}
			data.Registered = res.Registered
		}
	}
	return data, nil
}

// RegisterProvider registers a provider in the reputation system.
func (rc *ReputationContract) RegisterProvider(ctx context.Context, provider common.Address) (*types.Transaction, error) {
	if rc.mockMode {
		rc.mockMu.Lock()
		defer rc.mockMu.Unlock()
		rc.ensureRegistered(provider)
		logging.Info("registered provider in reputation system", "provider", provider.Hex())
		return nil, nil
	}

	auth, err := rc.baseClient.GetTransactOpts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction options: %w", err)
	}

	tx, err := rc.contract.Transact(auth, "registerProvider", provider)
	if err != nil {
		return nil, fmt.Errorf("failed to register provider: %w", err)
	}
	return tx, nil
}

// RecordJobCompleted records a successful job completion.
func (rc *ReputationContract) RecordJobCompleted(ctx context.Context, provider common.Address) (*types.Transaction, error) {
	if rc.mockMode {
		rc.mockMu.Lock()
		defer rc.mockMu.Unlock()
		rep := rc.ensureRegistered(provider)
		rep.JobsCompleted.Add(rep.JobsCompleted, big.NewInt(1))
		// Increase score by 10, capped at max
		rep.Score.Add(rep.Score, big.NewInt(10))
		if rep.Score.Cmp(big.NewInt(reputationMaxScore)) > 0 {
			rep.Score.SetInt64(reputationMaxScore)
		}
		return nil, nil
	}

	auth, err := rc.baseClient.GetTransactOpts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction options: %w", err)
	}

	tx, err := rc.contract.Transact(auth, "recordJobCompleted", provider)
	if err != nil {
		return nil, fmt.Errorf("failed to record job completed: %w", err)
	}
	return tx, nil
}

// RecordJobFailed records a failed job.
func (rc *ReputationContract) RecordJobFailed(ctx context.Context, provider common.Address) (*types.Transaction, error) {
	if rc.mockMode {
		rc.mockMu.Lock()
		defer rc.mockMu.Unlock()
		rep := rc.ensureRegistered(provider)
		rep.JobsFailed.Add(rep.JobsFailed, big.NewInt(1))
		// Decrease score by 50, floored at 0
		rep.Score.Sub(rep.Score, big.NewInt(50))
		if rep.Score.Sign() < 0 {
			rep.Score.SetInt64(0)
		}
		return nil, nil
	}

	auth, err := rc.baseClient.GetTransactOpts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction options: %w", err)
	}

	tx, err := rc.contract.Transact(auth, "recordJobFailed", provider)
	if err != nil {
		return nil, fmt.Errorf("failed to record job failed: %w", err)
	}
	return tx, nil
}

// RecordSlashEvent records a slash event against a provider.
func (rc *ReputationContract) RecordSlashEvent(ctx context.Context, provider common.Address) (*types.Transaction, error) {
	if rc.mockMode {
		rc.mockMu.Lock()
		defer rc.mockMu.Unlock()
		rep := rc.ensureRegistered(provider)
		rep.SlashEvents.Add(rep.SlashEvents, big.NewInt(1))
		// Decrease score by 100
		rep.Score.Sub(rep.Score, big.NewInt(100))
		if rep.Score.Sign() < 0 {
			rep.Score.SetInt64(0)
		}
		return nil, nil
	}

	auth, err := rc.baseClient.GetTransactOpts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction options: %w", err)
	}

	tx, err := rc.contract.Transact(auth, "recordSlashEvent", provider)
	if err != nil {
		return nil, fmt.Errorf("failed to record slash event: %w", err)
	}
	return tx, nil
}

// RecordEvent records an arbitrary score delta with a reason.
func (rc *ReputationContract) RecordEvent(ctx context.Context, provider common.Address, delta *big.Int, reason string) (*types.Transaction, error) {
	if rc.mockMode {
		rc.mockMu.Lock()
		defer rc.mockMu.Unlock()
		rep := rc.ensureRegistered(provider)

		// Clamp delta to ±reputationMaxDelta
		clampedDelta := new(big.Int).Set(delta)
		maxDelta := big.NewInt(reputationMaxDelta)
		minDelta := new(big.Int).Neg(maxDelta)
		if clampedDelta.Cmp(maxDelta) > 0 {
			clampedDelta.Set(maxDelta)
		} else if clampedDelta.Cmp(minDelta) < 0 {
			clampedDelta.Set(minDelta)
		}

		rep.Score.Add(rep.Score, clampedDelta)
		if rep.Score.Sign() < 0 {
			rep.Score.SetInt64(0)
		}
		if rep.Score.Cmp(big.NewInt(reputationMaxScore)) > 0 {
			rep.Score.SetInt64(reputationMaxScore)
		}

		logging.Info("recorded reputation event",
			"provider", provider.Hex(), "delta", delta.String(), "reason", reason,
			"new_score", rep.Score.String())
		return nil, nil
	}

	auth, err := rc.baseClient.GetTransactOpts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction options: %w", err)
	}

	tx, err := rc.contract.Transact(auth, "recordEvent", provider, delta, reason)
	if err != nil {
		return nil, fmt.Errorf("failed to record event: %w", err)
	}
	return tx, nil
}

// tierFromScore maps a score to a reputation tier.
func (rc *ReputationContract) tierFromScore(score *big.Int) ReputationTier {
	s := score.Int64()
	switch {
	case s >= 900:
		return ReputationTierElite
	case s >= 700:
		return ReputationTierEstablished
	case s >= 500:
		return ReputationTierReliable
	case s >= 250:
		return ReputationTierNewcomer
	default:
		return ReputationTierUntrusted
	}
}

// ─── Admin Setters ───────────────────────────────────────────────────────────

// SetTierThresholds adjusts the reputation tier threshold scores.
func (rc *ReputationContract) SetTierThresholds(ctx context.Context, probation, standard, trusted, elite uint16) (*types.Transaction, error) {
	if rc.mockMode {
		logging.Info("mock: setTierThresholds p=%d s=%d t=%d e=%d", probation, standard, trusted, elite)
		return nil, nil
	}
	auth, err := rc.baseClient.GetTransactOpts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction options: %w", err)
	}
	tx, err := rc.contract.Transact(auth, "setTierThresholds", probation, standard, trusted, elite)
	if err != nil {
		return nil, fmt.Errorf("failed to set tier thresholds: %w", err)
	}
	return tx, nil
}

// SetDecayParams adjusts the reputation decay rate and floor.
func (rc *ReputationContract) SetDecayParams(ctx context.Context, rate, floor uint16) (*types.Transaction, error) {
	if rc.mockMode {
		logging.Info("mock: setDecayParams rate=%d floor=%d", rate, floor)
		return nil, nil
	}
	auth, err := rc.baseClient.GetTransactOpts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction options: %w", err)
	}
	tx, err := rc.contract.Transact(auth, "setDecayParams", rate, floor)
	if err != nil {
		return nil, fmt.Errorf("failed to set decay params: %w", err)
	}
	return tx, nil
}

// SetMinScoreForJobs adjusts the minimum reputation score for job eligibility.
func (rc *ReputationContract) SetMinScoreForJobs(ctx context.Context, minScore uint16) (*types.Transaction, error) {
	if rc.mockMode {
		logging.Info("mock: setMinScoreForJobs min=%d", minScore)
		return nil, nil
	}
	auth, err := rc.baseClient.GetTransactOpts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction options: %w", err)
	}
	tx, err := rc.contract.Transact(auth, "setMinScoreForJobs", minScore)
	if err != nil {
		return nil, fmt.Errorf("failed to set min score for jobs: %w", err)
	}
	return tx, nil
}

// SetMaxCustomDelta adjusts the max custom delta range for reputation events.
func (rc *ReputationContract) SetMaxCustomDelta(ctx context.Context, maxDelta int16) (*types.Transaction, error) {
	if rc.mockMode {
		logging.Info("mock: setMaxCustomDelta delta=%d", maxDelta)
		return nil, nil
	}
	auth, err := rc.baseClient.GetTransactOpts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction options: %w", err)
	}
	tx, err := rc.contract.Transact(auth, "setMaxCustomDelta", maxDelta)
	if err != nil {
		return nil, fmt.Errorf("failed to set max custom delta: %w", err)
	}
	return tx, nil
}
