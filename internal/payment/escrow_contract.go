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
)

// EscrowContract provides interface to the payment escrow smart contract
type EscrowContract struct {
	baseClient    *BaseClient
	tokenContract *TokenContract
	contract      *bind.BoundContract
	contractABI   abi.ABI
	contractAddr  common.Address
	mockMode      bool

	// Mock state
	mockEscrows map[[32]byte]*EscrowData
	mockMu      sync.RWMutex
}

// EscrowData represents escrow data
type EscrowData struct {
	JobID     [32]byte
	Requester common.Address
	Provider  common.Address
	Amount    *big.Int
	Released  *big.Int
	Duration  *big.Int  // in seconds
	StartTime time.Time
	State     EscrowState
}

// EscrowCreatedEvent represents an escrow creation event
type EscrowCreatedEvent struct {
	JobID     [32]byte
	Requester common.Address
	Provider  common.Address
	Amount    *big.Int
	Timestamp time.Time
}

// PaymentReleasedEvent represents a payment release event
type PaymentReleasedEvent struct {
	JobID     [32]byte
	Amount    *big.Int
	Timestamp time.Time
}

// NewEscrowContract creates a new escrow contract client
func NewEscrowContract(baseClient *BaseClient, tokenContract *TokenContract, contractAddr common.Address) (*EscrowContract, error) {
	ec := &EscrowContract{
		baseClient:    baseClient,
		tokenContract: tokenContract,
		contractAddr:  contractAddr,
		mockEscrows:   make(map[[32]byte]*EscrowData),
	}

	// If no base client, use mock mode
	if baseClient == nil || !baseClient.IsConnected() {
		ec.mockMode = true
		return ec, nil
	}

	parsedABI, err := abi.JSON(strings.NewReader(EscrowContractABI))
	if err != nil {
		return nil, fmt.Errorf("failed to parse escrow ABI: %w", err)
	}
	ec.contractABI = parsedABI

	client := baseClient.Client()
	ec.contract = bind.NewBoundContract(contractAddr, parsedABI, client, client, client)

	return ec, nil
}

// NewMockEscrowContract creates a mock escrow contract for testing
func NewMockEscrowContract() *EscrowContract {
	return &EscrowContract{
		mockMode:    true,
		mockEscrows: make(map[[32]byte]*EscrowData),
	}
}

// IsMockMode returns whether running in mock mode
func (ec *EscrowContract) IsMockMode() bool {
	return ec.mockMode
}

// CreateEscrow creates a new escrow for a job
func (ec *EscrowContract) CreateEscrow(ctx context.Context, jobID [32]byte, provider common.Address, amount *big.Int, duration *big.Int) (*types.Transaction, error) {
	if ec.mockMode {
		return ec.mockCreateEscrow(ctx, jobID, provider, amount, duration)
	}

	// First approve the escrow contract to spend tokens
	if ec.tokenContract != nil {
		_, err := ec.tokenContract.ApproveAndWait(ctx, ec.contractAddr, amount)
		if err != nil {
			return nil, fmt.Errorf("failed to approve token spend: %w", err)
		}
	}

	auth, err := ec.baseClient.GetTransactOpts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction options: %w", err)
	}

	tx, err := ec.contract.Transact(auth, "createEscrow", jobID, provider, amount, duration)
	if err != nil {
		return nil, fmt.Errorf("failed to create escrow: %w", err)
	}

	return tx, nil
}

// mockCreateEscrow handles escrow creation in mock mode
func (ec *EscrowContract) mockCreateEscrow(_ context.Context, jobID [32]byte, provider common.Address, amount *big.Int, duration *big.Int) (*types.Transaction, error) {
	ec.mockMu.Lock()
	defer ec.mockMu.Unlock()

	if _, exists := ec.mockEscrows[jobID]; exists {
		return nil, fmt.Errorf("escrow already exists for job")
	}

	var requester common.Address
	if ec.baseClient != nil {
		requester = ec.baseClient.Address()
	}
	ec.mockEscrows[jobID] = &EscrowData{
		JobID:     jobID,
		Requester: requester,
		Provider:  provider,
		Amount:    new(big.Int).Set(amount),
		Released:  big.NewInt(0),
		Duration:  new(big.Int).Set(duration),
		StartTime: time.Now(),
		State:     EscrowStateActive,
	}

	fmt.Printf("[MOCK] Created escrow for job %x: %s BUNKER from %s to %s\n",
		jobID[:8], amount.String(), requester.Hex(), provider.Hex())

	return nil, nil
}

// CreateEscrowAndWait creates escrow and waits for confirmation
func (ec *EscrowContract) CreateEscrowAndWait(ctx context.Context, jobID [32]byte, provider common.Address, amount *big.Int, duration *big.Int) (*types.Receipt, error) {
	tx, err := ec.CreateEscrow(ctx, jobID, provider, amount, duration)
	if err != nil {
		return nil, err
	}

	if ec.mockMode || tx == nil {
		return nil, nil
	}

	return ec.baseClient.WaitForTransaction(ctx, tx)
}

// ReleasePayment releases payment to provider based on uptime
func (ec *EscrowContract) ReleasePayment(ctx context.Context, jobID [32]byte, uptime *big.Int) (*types.Transaction, error) {
	if ec.mockMode {
		return ec.mockReleasePayment(ctx, jobID, uptime)
	}

	auth, err := ec.baseClient.GetTransactOpts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction options: %w", err)
	}

	tx, err := ec.contract.Transact(auth, "releasePayment", jobID, uptime)
	if err != nil {
		return nil, fmt.Errorf("failed to release payment: %w", err)
	}

	return tx, nil
}

// mockReleasePayment handles payment release in mock mode
func (ec *EscrowContract) mockReleasePayment(_ context.Context, jobID [32]byte, uptime *big.Int) (*types.Transaction, error) {
	ec.mockMu.Lock()
	defer ec.mockMu.Unlock()

	escrow, exists := ec.mockEscrows[jobID]
	if !exists {
		return nil, fmt.Errorf("escrow not found")
	}

	if escrow.State != EscrowStateActive {
		return nil, fmt.Errorf("escrow not active")
	}

	// Calculate release amount based on uptime/duration ratio
	// uptime and duration are both in seconds
	proportion := new(big.Float).Quo(
		new(big.Float).SetInt(uptime),
		new(big.Float).SetInt(escrow.Duration),
	)
	if proportion.Cmp(big.NewFloat(1.0)) > 0 {
		proportion = big.NewFloat(1.0)
	}

	totalReleasable, _ := new(big.Float).Mul(
		new(big.Float).SetInt(escrow.Amount),
		proportion,
	).Int(nil)

	toRelease := new(big.Int).Sub(totalReleasable, escrow.Released)
	if toRelease.Sign() <= 0 {
		return nil, nil // Nothing to release
	}

	escrow.Released.Add(escrow.Released, toRelease)

	fmt.Printf("[MOCK] Released %s BUNKER for job %x (total released: %s/%s)\n",
		toRelease.String(), jobID[:8], escrow.Released.String(), escrow.Amount.String())

	return nil, nil
}

// Refund refunds the remaining escrow to requester
func (ec *EscrowContract) Refund(ctx context.Context, jobID [32]byte) (*types.Transaction, error) {
	if ec.mockMode {
		return ec.mockRefund(ctx, jobID)
	}

	auth, err := ec.baseClient.GetTransactOpts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction options: %w", err)
	}

	tx, err := ec.contract.Transact(auth, "refund", jobID)
	if err != nil {
		return nil, fmt.Errorf("failed to refund: %w", err)
	}

	return tx, nil
}

// mockRefund handles refund in mock mode
func (ec *EscrowContract) mockRefund(_ context.Context, jobID [32]byte) (*types.Transaction, error) {
	ec.mockMu.Lock()
	defer ec.mockMu.Unlock()

	escrow, exists := ec.mockEscrows[jobID]
	if !exists {
		return nil, fmt.Errorf("escrow not found")
	}

	if escrow.State != EscrowStateActive {
		return nil, fmt.Errorf("escrow not active")
	}

	refundAmount := new(big.Int).Sub(escrow.Amount, escrow.Released)
	escrow.State = EscrowStateRefunded

	fmt.Printf("[MOCK] Refunded %s BUNKER for job %x\n", refundAmount.String(), jobID[:8])

	return nil, nil
}

// FinalizeEscrow finalizes an escrow (marks as complete)
func (ec *EscrowContract) FinalizeEscrow(ctx context.Context, jobID [32]byte) (*types.Transaction, error) {
	if ec.mockMode {
		return ec.mockFinalizeEscrow(ctx, jobID)
	}

	auth, err := ec.baseClient.GetTransactOpts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction options: %w", err)
	}

	tx, err := ec.contract.Transact(auth, "finalizeEscrow", jobID)
	if err != nil {
		return nil, fmt.Errorf("failed to finalize escrow: %w", err)
	}

	return tx, nil
}

// mockFinalizeEscrow handles finalization in mock mode
func (ec *EscrowContract) mockFinalizeEscrow(_ context.Context, jobID [32]byte) (*types.Transaction, error) {
	ec.mockMu.Lock()
	defer ec.mockMu.Unlock()

	escrow, exists := ec.mockEscrows[jobID]
	if !exists {
		return nil, fmt.Errorf("escrow not found")
	}

	escrow.State = EscrowStateCompleted
	fmt.Printf("[MOCK] Finalized escrow for job %x\n", jobID[:8])

	return nil, nil
}

// GetEscrow returns escrow data for a job
func (ec *EscrowContract) GetEscrow(ctx context.Context, jobID [32]byte) (*EscrowData, error) {
	if ec.mockMode {
		ec.mockMu.RLock()
		defer ec.mockMu.RUnlock()
		escrow, exists := ec.mockEscrows[jobID]
		if !exists {
			return nil, fmt.Errorf("escrow not found")
		}
		// Return a copy
		return &EscrowData{
			JobID:     escrow.JobID,
			Requester: escrow.Requester,
			Provider:  escrow.Provider,
			Amount:    new(big.Int).Set(escrow.Amount),
			Released:  new(big.Int).Set(escrow.Released),
			Duration:  new(big.Int).Set(escrow.Duration),
			StartTime: escrow.StartTime,
			State:     escrow.State,
		}, nil
	}

	var result []interface{}
	err := ec.contract.Call(&bind.CallOpts{Context: ctx}, &result, "getEscrow", jobID)
	if err != nil {
		return nil, fmt.Errorf("failed to get escrow: %w", err)
	}

	if len(result) < 7 {
		return nil, fmt.Errorf("unexpected result format")
	}

	escrow := &EscrowData{
		JobID: jobID,
	}

	if addr, ok := result[0].(common.Address); ok {
		escrow.Requester = addr
	}
	if addr, ok := result[1].(common.Address); ok {
		escrow.Provider = addr
	}
	if amount, ok := result[2].(*big.Int); ok {
		escrow.Amount = amount
	}
	if released, ok := result[3].(*big.Int); ok {
		escrow.Released = released
	}
	if duration, ok := result[4].(*big.Int); ok {
		escrow.Duration = duration
	}
	if startTime, ok := result[5].(*big.Int); ok {
		escrow.StartTime = time.Unix(startTime.Int64(), 0)
	}
	if state, ok := result[6].(uint8); ok {
		escrow.State = EscrowState(state)
	}

	return escrow, nil
}

// GetRequesterEscrows returns all escrow job IDs for a requester
func (ec *EscrowContract) GetRequesterEscrows(ctx context.Context, requester common.Address) ([][32]byte, error) {
	if ec.mockMode {
		ec.mockMu.RLock()
		defer ec.mockMu.RUnlock()
		var jobIDs [][32]byte
		for _, escrow := range ec.mockEscrows {
			if escrow.Requester == requester {
				jobIDs = append(jobIDs, escrow.JobID)
			}
		}
		return jobIDs, nil
	}

	var result []interface{}
	err := ec.contract.Call(&bind.CallOpts{Context: ctx}, &result, "getRequesterEscrows", requester)
	if err != nil {
		return nil, fmt.Errorf("failed to get requester escrows: %w", err)
	}

	if len(result) == 0 {
		return nil, nil
	}
	// Parse result as [][32]byte
	if jobIDs, ok := result[0].([][32]byte); ok {
		return jobIDs, nil
	}
	return nil, nil
}

// GetProviderEscrows returns all escrow job IDs for a provider
func (ec *EscrowContract) GetProviderEscrows(ctx context.Context, provider common.Address) ([][32]byte, error) {
	if ec.mockMode {
		ec.mockMu.RLock()
		defer ec.mockMu.RUnlock()
		var jobIDs [][32]byte
		for _, escrow := range ec.mockEscrows {
			if escrow.Provider == provider {
				jobIDs = append(jobIDs, escrow.JobID)
			}
		}
		return jobIDs, nil
	}

	var result []interface{}
	err := ec.contract.Call(&bind.CallOpts{Context: ctx}, &result, "getProviderEscrows", provider)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider escrows: %w", err)
	}

	if len(result) == 0 {
		return nil, nil
	}
	// Parse result as [][32]byte
	if jobIDs, ok := result[0].([][32]byte); ok {
		return jobIDs, nil
	}
	return nil, nil
}

// SubscribeEscrowEvents subscribes to escrow events
func (ec *EscrowContract) SubscribeEscrowEvents(ctx context.Context, ch chan<- *EscrowCreatedEvent) error {
	if ec.mockMode {
		return nil // No events in mock mode
	}

	wsClient := ec.baseClient.WSClient()
	if wsClient == nil {
		return fmt.Errorf("WebSocket client not available")
	}

	query := ethereum.FilterQuery{
		Addresses: []common.Address{ec.contractAddr},
		Topics:    [][]common.Hash{{ec.contractABI.Events["EscrowCreated"].ID}},
	}

	logs := make(chan types.Log)
	sub, err := wsClient.SubscribeFilterLogs(ctx, query, logs)
	if err != nil {
		return fmt.Errorf("failed to subscribe to events: %w", err)
	}

	util.SafeGoWithName("escrow-event-listener", func() {
		defer sub.Unsubscribe()
		for {
			select {
			case <-ctx.Done():
				return
			case log := <-logs:
				event, err := ec.parseEscrowCreatedEvent(log)
				if err != nil {
					continue
				}
				ch <- event
			case err := <-sub.Err():
				if err != nil {
					fmt.Printf("Escrow event subscription error: %v\n", err)
				}
				return
			}
		}
	})

	return nil
}

// parseEscrowCreatedEvent parses an escrow created event from log
func (ec *EscrowContract) parseEscrowCreatedEvent(log types.Log) (*EscrowCreatedEvent, error) {
	event := &EscrowCreatedEvent{
		Timestamp: time.Now(),
	}

	if len(log.Topics) > 1 {
		copy(event.JobID[:], log.Topics[1].Bytes())
	}
	if len(log.Topics) > 2 {
		event.Requester = common.HexToAddress(log.Topics[2].Hex())
	}
	if len(log.Topics) > 3 {
		event.Provider = common.HexToAddress(log.Topics[3].Hex())
	}

	// Parse data (amount)
	if len(log.Data) >= 32 {
		event.Amount = new(big.Int).SetBytes(log.Data[:32])
	}

	return event, nil
}

// CalculateRemainingEscrow calculates remaining escrow amount
func (ec *EscrowContract) CalculateRemainingEscrow(escrow *EscrowData) *big.Int {
	return new(big.Int).Sub(escrow.Amount, escrow.Released)
}

// CalculateEarnedAmount calculates earned amount based on elapsed time
func (ec *EscrowContract) CalculateEarnedAmount(escrow *EscrowData) *big.Int {
	elapsed := time.Since(escrow.StartTime)
	duration := time.Duration(escrow.Duration.Int64()) * time.Second

	if elapsed >= duration {
		return new(big.Int).Set(escrow.Amount)
	}

	proportion := new(big.Float).Quo(
		new(big.Float).SetFloat64(elapsed.Seconds()),
		new(big.Float).SetFloat64(duration.Seconds()),
	)

	earned, _ := new(big.Float).Mul(
		new(big.Float).SetInt(escrow.Amount),
		proportion,
	).Int(nil)

	return earned
}
