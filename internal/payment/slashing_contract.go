package payment

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
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
)

// SlashingContract provides interface to the slashing smart contract
type SlashingContract struct {
	baseClient   *BaseClient
	stakingContract *StakingContract
	contract     *bind.BoundContract
	contractABI  abi.ABI
	contractAddr common.Address
	mockMode     bool

	// Mock state
	mockDisputes map[[32]byte]*DisputeData
	mockHistory  map[common.Address]*SlashingHistory
	mockMu       sync.RWMutex
}

// DisputeData represents a dispute
type DisputeData struct {
	DisputeID [32]byte
	Reporter  common.Address
	Provider  common.Address
	JobID     [32]byte
	Reason    ViolationReason
	State     DisputeState
	Evidence  []byte
	Defense   []byte
	Timestamp time.Time
}

// SlashingHistory represents slashing history for a provider
type SlashingHistory struct {
	TotalSlashed *big.Int
	Violations   int
}

// SlashEvent represents a slashing event
type SlashEvent struct {
	Provider  common.Address
	Amount    *big.Int
	Reason    ViolationReason
	Timestamp time.Time
}

// NewSlashingContract creates a new slashing contract client
func NewSlashingContract(baseClient *BaseClient, stakingContract *StakingContract, contractAddr common.Address) (*SlashingContract, error) {
	sc := &SlashingContract{
		baseClient:     baseClient,
		stakingContract: stakingContract,
		contractAddr:   contractAddr,
		mockDisputes:   make(map[[32]byte]*DisputeData),
		mockHistory:    make(map[common.Address]*SlashingHistory),
	}

	// If no base client, use mock mode
	if baseClient == nil || !baseClient.IsConnected() {
		sc.mockMode = true
		return sc, nil
	}

	parsedABI, err := abi.JSON(strings.NewReader(SlashingContractABI))
	if err != nil {
		return nil, fmt.Errorf("failed to parse slashing ABI: %w", err)
	}
	sc.contractABI = parsedABI

	client := baseClient.Client()
	sc.contract = bind.NewBoundContract(contractAddr, parsedABI, client, client, client)

	return sc, nil
}

// NewMockSlashingContract creates a mock slashing contract for testing
func NewMockSlashingContract() *SlashingContract {
	return &SlashingContract{
		mockMode:     true,
		mockDisputes: make(map[[32]byte]*DisputeData),
		mockHistory:  make(map[common.Address]*SlashingHistory),
	}
}

// IsMockMode returns whether running in mock mode
func (sc *SlashingContract) IsMockMode() bool {
	return sc.mockMode
}

// ReportViolation reports a violation against a provider
func (sc *SlashingContract) ReportViolation(ctx context.Context, provider common.Address, jobID [32]byte, reason ViolationReason, evidence []byte) ([32]byte, *types.Transaction, error) {
	if sc.mockMode {
		return sc.mockReportViolation(ctx, provider, jobID, reason, evidence)
	}

	auth, err := sc.baseClient.GetTransactOpts(ctx)
	if err != nil {
		return [32]byte{}, nil, fmt.Errorf("failed to get transaction options: %w", err)
	}

	tx, err := sc.contract.Transact(auth, "reportViolation", provider, jobID, uint8(reason), evidence)
	if err != nil {
		return [32]byte{}, nil, fmt.Errorf("failed to report violation: %w", err)
	}

	// Generate dispute ID (would come from event in real implementation)
	disputeID := generateDisputeID(provider, jobID)

	return disputeID, tx, nil
}

// mockReportViolation handles violation reporting in mock mode
func (sc *SlashingContract) mockReportViolation(_ context.Context, provider common.Address, jobID [32]byte, reason ViolationReason, evidence []byte) ([32]byte, *types.Transaction, error) {
	sc.mockMu.Lock()
	defer sc.mockMu.Unlock()

	disputeID := generateDisputeID(provider, jobID)
	if _, exists := sc.mockDisputes[disputeID]; exists {
		return [32]byte{}, nil, fmt.Errorf("dispute already exists")
	}

	var reporter common.Address
	if sc.baseClient != nil {
		reporter = sc.baseClient.Address()
	}
	sc.mockDisputes[disputeID] = &DisputeData{
		DisputeID: disputeID,
		Reporter:  reporter,
		Provider:  provider,
		JobID:     jobID,
		Reason:    reason,
		State:     DisputeStatePending,
		Evidence:  evidence,
		Timestamp: time.Now(),
	}

	fmt.Printf("[MOCK] Reported violation against %s: reason=%s, disputeID=%x\n",
		provider.Hex(), reason.String(), disputeID[:8])

	return disputeID, nil, nil
}

// generateDisputeID generates a deterministic dispute ID
func generateDisputeID(provider common.Address, jobID [32]byte) [32]byte {
	h := sha256.New()
	h.Write(provider.Bytes())
	h.Write(jobID[:])
	now := make([]byte, 8)
	binary.BigEndian.PutUint64(now, uint64(time.Now().UnixNano()))
	h.Write(now)
	var id [32]byte
	copy(id[:], h.Sum(nil))
	return id
}

// SubmitDefense submits a defense for a dispute
func (sc *SlashingContract) SubmitDefense(ctx context.Context, disputeID [32]byte, defense []byte) (*types.Transaction, error) {
	if sc.mockMode {
		return sc.mockSubmitDefense(ctx, disputeID, defense)
	}

	auth, err := sc.baseClient.GetTransactOpts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction options: %w", err)
	}

	tx, err := sc.contract.Transact(auth, "submitDefense", disputeID, defense)
	if err != nil {
		return nil, fmt.Errorf("failed to submit defense: %w", err)
	}

	return tx, nil
}

// mockSubmitDefense handles defense submission in mock mode
func (sc *SlashingContract) mockSubmitDefense(_ context.Context, disputeID [32]byte, defense []byte) (*types.Transaction, error) {
	sc.mockMu.Lock()
	defer sc.mockMu.Unlock()

	dispute, exists := sc.mockDisputes[disputeID]
	if !exists {
		return nil, fmt.Errorf("dispute not found")
	}

	if dispute.State != DisputeStatePending {
		return nil, fmt.Errorf("dispute not pending")
	}

	dispute.Defense = defense
	dispute.State = DisputeStateDefenseSubmitted

	fmt.Printf("[MOCK] Defense submitted for dispute %x\n", disputeID[:8])

	return nil, nil
}

// ResolveDispute resolves a dispute (governance/arbitration)
func (sc *SlashingContract) ResolveDispute(ctx context.Context, disputeID [32]byte, slashAmount *big.Int) (*types.Transaction, error) {
	if sc.mockMode {
		return sc.mockResolveDispute(ctx, disputeID, slashAmount)
	}

	auth, err := sc.baseClient.GetTransactOpts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction options: %w", err)
	}

	tx, err := sc.contract.Transact(auth, "resolveDispute", disputeID, slashAmount)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve dispute: %w", err)
	}

	return tx, nil
}

// mockResolveDispute handles dispute resolution in mock mode
func (sc *SlashingContract) mockResolveDispute(_ context.Context, disputeID [32]byte, slashAmount *big.Int) (*types.Transaction, error) {
	sc.mockMu.Lock()
	defer sc.mockMu.Unlock()

	dispute, exists := sc.mockDisputes[disputeID]
	if !exists {
		return nil, fmt.Errorf("dispute not found")
	}

	if dispute.State == DisputeStateResolved {
		return nil, fmt.Errorf("dispute already resolved")
	}

	dispute.State = DisputeStateResolved

	// Update slashing history
	history, exists := sc.mockHistory[dispute.Provider]
	if !exists {
		history = &SlashingHistory{TotalSlashed: big.NewInt(0)}
		sc.mockHistory[dispute.Provider] = history
	}
	history.TotalSlashed.Add(history.TotalSlashed, slashAmount)
	history.Violations++

	fmt.Printf("[MOCK] Resolved dispute %x: slashed %s from %s\n",
		disputeID[:8], slashAmount.String(), dispute.Provider.Hex())

	return nil, nil
}

// Slash directly slashes a provider (for automated penalties)
func (sc *SlashingContract) Slash(ctx context.Context, provider common.Address, amount *big.Int, reason ViolationReason) (*types.Transaction, error) {
	if sc.mockMode {
		return sc.mockSlash(ctx, provider, amount, reason)
	}

	auth, err := sc.baseClient.GetTransactOpts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction options: %w", err)
	}

	tx, err := sc.contract.Transact(auth, "slash", provider, amount, uint8(reason))
	if err != nil {
		return nil, fmt.Errorf("failed to slash: %w", err)
	}

	return tx, nil
}

// mockSlash handles direct slashing in mock mode
func (sc *SlashingContract) mockSlash(_ context.Context, provider common.Address, amount *big.Int, reason ViolationReason) (*types.Transaction, error) {
	sc.mockMu.Lock()
	defer sc.mockMu.Unlock()

	// Update slashing history
	history, exists := sc.mockHistory[provider]
	if !exists {
		history = &SlashingHistory{TotalSlashed: big.NewInt(0)}
		sc.mockHistory[provider] = history
	}
	history.TotalSlashed.Add(history.TotalSlashed, amount)
	history.Violations++

	fmt.Printf("[MOCK] Slashed %s from %s: reason=%s\n",
		amount.String(), provider.Hex(), reason.String())

	return nil, nil
}

// GetDispute returns dispute data
func (sc *SlashingContract) GetDispute(ctx context.Context, disputeID [32]byte) (*DisputeData, error) {
	if sc.mockMode {
		sc.mockMu.RLock()
		defer sc.mockMu.RUnlock()
		dispute, exists := sc.mockDisputes[disputeID]
		if !exists {
			return nil, fmt.Errorf("dispute not found")
		}
		// Return a copy
		return &DisputeData{
			DisputeID: dispute.DisputeID,
			Reporter:  dispute.Reporter,
			Provider:  dispute.Provider,
			JobID:     dispute.JobID,
			Reason:    dispute.Reason,
			State:     dispute.State,
			Evidence:  dispute.Evidence,
			Defense:   dispute.Defense,
			Timestamp: dispute.Timestamp,
		}, nil
	}

	var result []interface{}
	err := sc.contract.Call(&bind.CallOpts{Context: ctx}, &result, "getDispute", disputeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get dispute: %w", err)
	}

	if len(result) < 6 {
		return nil, fmt.Errorf("unexpected result format")
	}

	dispute := &DisputeData{
		DisputeID: disputeID,
	}

	if addr, ok := result[0].(common.Address); ok {
		dispute.Reporter = addr
	}
	if addr, ok := result[1].(common.Address); ok {
		dispute.Provider = addr
	}
	if jobID, ok := result[2].([32]byte); ok {
		dispute.JobID = jobID
	}
	if reason, ok := result[3].(uint8); ok {
		dispute.Reason = ViolationReason(reason)
	}
	if state, ok := result[4].(uint8); ok {
		dispute.State = DisputeState(state)
	}
	if timestamp, ok := result[5].(*big.Int); ok {
		dispute.Timestamp = time.Unix(timestamp.Int64(), 0)
	}

	return dispute, nil
}

// GetSlashingHistory returns slashing history for a provider
func (sc *SlashingContract) GetSlashingHistory(ctx context.Context, provider common.Address) (*SlashingHistory, error) {
	if sc.mockMode {
		sc.mockMu.RLock()
		defer sc.mockMu.RUnlock()
		history, exists := sc.mockHistory[provider]
		if !exists {
			return &SlashingHistory{TotalSlashed: big.NewInt(0)}, nil
		}
		return &SlashingHistory{
			TotalSlashed: new(big.Int).Set(history.TotalSlashed),
			Violations:   history.Violations,
		}, nil
	}

	var result []interface{}
	err := sc.contract.Call(&bind.CallOpts{Context: ctx}, &result, "getSlashingHistory", provider)
	if err != nil {
		return nil, fmt.Errorf("failed to get slashing history: %w", err)
	}

	history := &SlashingHistory{TotalSlashed: big.NewInt(0)}
	if len(result) >= 2 {
		if total, ok := result[0].(*big.Int); ok {
			history.TotalSlashed = total
		}
		if violations, ok := result[1].(*big.Int); ok {
			history.Violations = int(violations.Int64())
		}
	}

	return history, nil
}

// SubscribeSlashEvents subscribes to slashing events
func (sc *SlashingContract) SubscribeSlashEvents(ctx context.Context, ch chan<- *SlashEvent) error {
	if sc.mockMode {
		return nil // No events in mock mode
	}

	wsClient := sc.baseClient.WSClient()
	if wsClient == nil {
		return fmt.Errorf("WebSocket client not available")
	}

	query := ethereum.FilterQuery{
		Addresses: []common.Address{sc.contractAddr},
		Topics:    [][]common.Hash{{sc.contractABI.Events["Slashed"].ID}},
	}

	logs := make(chan types.Log)
	sub, err := wsClient.SubscribeFilterLogs(ctx, query, logs)
	if err != nil {
		return fmt.Errorf("failed to subscribe to events: %w", err)
	}

	util.SafeGoWithName("slash-event-listener", func() {
		defer sub.Unsubscribe()
		for {
			select {
			case <-ctx.Done():
				return
			case log := <-logs:
				event, err := sc.parseSlashEvent(log)
				if err != nil {
					continue
				}
				ch <- event
			case err := <-sub.Err():
				if err != nil {
					fmt.Printf("Slash event subscription error: %v\n", err)
				}
				return
			}
		}
	})

	return nil
}

// parseSlashEvent parses a slash event from log
func (sc *SlashingContract) parseSlashEvent(log types.Log) (*SlashEvent, error) {
	event := &SlashEvent{
		Timestamp: time.Now(),
	}

	if len(log.Topics) > 1 {
		event.Provider = common.HexToAddress(log.Topics[1].Hex())
	}

	// Parse data (amount, reason)
	if len(log.Data) >= 64 {
		event.Amount = new(big.Int).SetBytes(log.Data[:32])
		event.Reason = ViolationReason(new(big.Int).SetBytes(log.Data[32:64]).Uint64())
	}

	return event, nil
}

// CalculateSlashAmount calculates the slash amount based on violation severity
func (sc *SlashingContract) CalculateSlashAmount(stake *big.Int, reason ViolationReason) *big.Int {
	// Slashing percentages based on violation severity
	var percentage int64
	switch reason {
	case ViolationDowntime:
		percentage = 5 // 5% for downtime
	case ViolationSLAViolation:
		percentage = 10 // 10% for SLA violations
	case ViolationDataLoss:
		percentage = 25 // 25% for data loss
	case ViolationSecurityBreach:
		percentage = 50 // 50% for security breach
	case ViolationMaliciousBehavior:
		percentage = 100 // 100% for malicious behavior
	default:
		percentage = 0
	}

	if percentage == 0 {
		return big.NewInt(0)
	}

	// Calculate: stake * percentage / 100
	slashAmount := new(big.Int).Mul(stake, big.NewInt(percentage))
	slashAmount.Div(slashAmount, big.NewInt(100))

	return slashAmount
}
