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
)

// EscrowContract provides interface to the BunkerEscrow smart contract.
// It bridges Go-side [32]byte job IDs to on-chain uint256 reservation IDs.
type EscrowContract struct {
	baseClient    *BaseClient
	tokenContract *TokenContract
	contract      *bind.BoundContract
	contractABI   abi.ABI
	contractAddr  common.Address
	mockMode      bool

	// Maps Go-side jobID → on-chain reservationId (real mode only)
	reservationIDs    map[[32]byte]*big.Int
	reservationTimes  map[[32]byte]time.Time // tracks when each reservation was stored
	reservationMu     sync.RWMutex

	// Mock state
	mockEscrows map[[32]byte]*EscrowData
	mockMu      sync.RWMutex
}

// EscrowData represents escrow data
type EscrowData struct {
	JobID         [32]byte
	ReservationID *big.Int // on-chain reservation ID
	Requester     common.Address
	Provider      common.Address
	Providers     [3]common.Address
	Amount        *big.Int
	Released      *big.Int
	Duration      *big.Int // in seconds
	StartTime     time.Time
	State         EscrowState
}

// EscrowCreatedEvent represents a reservation creation event
type EscrowCreatedEvent struct {
	ReservationID *big.Int
	JobID         [32]byte // populated by the wrapper from the mapping
	Requester     common.Address
	Amount        *big.Int
	Duration      *big.Int
	Timestamp     time.Time
}

// PaymentReleasedEvent represents a payment release event
type PaymentReleasedEvent struct {
	ReservationID *big.Int
	JobID         [32]byte
	Amount        *big.Int
	Timestamp     time.Time
}

// NewEscrowContract creates a new escrow contract client
func NewEscrowContract(baseClient *BaseClient, tokenContract *TokenContract, contractAddr common.Address) (*EscrowContract, error) {
	ec := &EscrowContract{
		baseClient:     baseClient,
		tokenContract:  tokenContract,
		contractAddr:   contractAddr,
		reservationIDs:   make(map[[32]byte]*big.Int),
		reservationTimes: make(map[[32]byte]time.Time),
		mockEscrows:      make(map[[32]byte]*EscrowData),
	}

	// Require a connected base client; use NewMockEscrowContract() for testing
	if baseClient == nil {
		return nil, fmt.Errorf("base client is required (use NewMockEscrowContract for testing)")
	}
	if !baseClient.IsConnected() {
		return nil, fmt.Errorf("base client not connected to RPC")
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
		mockMode:       true,
		reservationIDs:   make(map[[32]byte]*big.Int),
		reservationTimes: make(map[[32]byte]time.Time),
		mockEscrows:      make(map[[32]byte]*EscrowData),
	}
}

// IsMockMode returns whether running in mock mode
func (ec *EscrowContract) IsMockMode() bool {
	return ec.mockMode
}

// getReservationID looks up the on-chain reservation ID for a job ID
func (ec *EscrowContract) getReservationID(jobID [32]byte) (*big.Int, error) {
	ec.reservationMu.RLock()
	defer ec.reservationMu.RUnlock()
	resID, exists := ec.reservationIDs[jobID]
	if !exists {
		return nil, fmt.Errorf("no reservation ID found for job %x", jobID[:8])
	}
	return new(big.Int).Set(resID), nil
}

// storeReservationID stores the mapping from job ID to on-chain reservation ID
func (ec *EscrowContract) storeReservationID(jobID [32]byte, resID *big.Int) {
	ec.reservationMu.Lock()
	defer ec.reservationMu.Unlock()
	ec.reservationIDs[jobID] = new(big.Int).Set(resID)
	ec.reservationTimes[jobID] = time.Now()
}

// StoreExternalReservationID stores a user-created on-chain reservation ID
// for a job, allowing SelectProviders and other operations to reference it.
func (ec *EscrowContract) StoreExternalReservationID(jobID [32]byte, resID *big.Int) {
	ec.storeReservationID(jobID, resID)
}

// removeReservationID removes the jobID→reservationID mapping when the escrow
// reaches a terminal state (finalized, refunded) to prevent unbounded growth.
func (ec *EscrowContract) removeReservationID(jobID [32]byte) {
	ec.reservationMu.Lock()
	defer ec.reservationMu.Unlock()
	delete(ec.reservationIDs, jobID)
	delete(ec.reservationTimes, jobID)
}

// CleanupStaleReservations removes reservation mappings older than maxAge.
// This prevents unbounded growth from abandoned jobs that never finalize or refund.
func (ec *EscrowContract) CleanupStaleReservations(maxAge time.Duration) int {
	ec.reservationMu.Lock()
	defer ec.reservationMu.Unlock()
	cutoff := time.Now().Add(-maxAge)
	removed := 0
	for jobID, created := range ec.reservationTimes {
		if created.Before(cutoff) {
			delete(ec.reservationIDs, jobID)
			delete(ec.reservationTimes, jobID)
			removed++
		}
	}
	return removed
}

// CreateEscrow creates a new escrow (reservation) for a job.
// In real mode, calls createReservation(amount, duration) and stores the returned reservation ID.
func (ec *EscrowContract) CreateEscrow(ctx context.Context, jobID [32]byte, provider common.Address, amount *big.Int, duration *big.Int) (*types.Transaction, error) {
	if ec.mockMode {
		return ec.mockCreateEscrow(ctx, jobID, provider, amount, duration)
	}

	// Approve the escrow contract to spend tokens
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

	tx, err := ec.contract.Transact(auth, "createReservation", amount, duration)
	if err != nil {
		return nil, fmt.Errorf("failed to create reservation: %w", err)
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

	// Generate a mock reservation ID
	resID := big.NewInt(int64(len(ec.mockEscrows) + 1))

	ec.mockEscrows[jobID] = &EscrowData{
		JobID:         jobID,
		ReservationID: resID,
		Requester:     requester,
		Provider:      provider,
		Amount:        new(big.Int).Set(amount),
		Released:      big.NewInt(0),
		Duration:      new(big.Int).Set(duration),
		StartTime:     time.Now(),
		State:         EscrowStateCreated,
	}

	// Store reservation ID mapping
	ec.reservationMu.Lock()
	ec.reservationIDs[jobID] = new(big.Int).Set(resID)
	ec.reservationMu.Unlock()

	logging.Info("created reservation", "job_id", fmt.Sprintf("%x", jobID[:8]), "reservation_id", resID.String(), "amount", amount.String(), "requester", requester.Hex())

	return nil, nil
}

// CreateEscrowAndWait creates escrow and waits for confirmation.
// Parses the ReservationCreated event to extract the reservation ID.
func (ec *EscrowContract) CreateEscrowAndWait(ctx context.Context, jobID [32]byte, provider common.Address, amount *big.Int, duration *big.Int) (*types.Receipt, error) {
	tx, err := ec.CreateEscrow(ctx, jobID, provider, amount, duration)
	if err != nil {
		return nil, err
	}

	if ec.mockMode || tx == nil {
		return nil, nil
	}

	receipt, err := ec.baseClient.WaitForTransaction(ctx, tx)
	if err != nil {
		return nil, err
	}

	// Parse ReservationCreated event from receipt to get the reservation ID
	resID := ec.parseReservationIDFromReceipt(receipt)
	if resID != nil {
		ec.storeReservationID(jobID, resID)
	}

	return receipt, nil
}

// parseReservationIDFromReceipt extracts the reservation ID from ReservationCreated event logs
func (ec *EscrowContract) parseReservationIDFromReceipt(receipt *types.Receipt) *big.Int {
	eventID := ec.contractABI.Events["ReservationCreated"].ID
	for _, log := range receipt.Logs {
		if len(log.Topics) > 0 && log.Topics[0] == eventID {
			// ReservationCreated has reservationId as indexed topic[1]
			if len(log.Topics) > 1 {
				return new(big.Int).SetBytes(log.Topics[1].Bytes())
			}
		}
	}
	return nil
}

// SelectProviders assigns providers to an existing reservation.
// Must be called after CreateEscrow to activate the escrow.
func (ec *EscrowContract) SelectProviders(ctx context.Context, jobID [32]byte, providers [3]common.Address) (*types.Transaction, error) {
	if ec.mockMode {
		return ec.mockSelectProviders(ctx, jobID, providers)
	}

	resID, err := ec.getReservationID(jobID)
	if err != nil {
		return nil, err
	}

	auth, err := ec.baseClient.GetTransactOpts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction options: %w", err)
	}

	tx, err := ec.contract.Transact(auth, "selectProviders", resID, providers)
	if err != nil {
		return nil, fmt.Errorf("failed to select providers: %w", err)
	}

	return tx, nil
}

// mockSelectProviders handles provider selection in mock mode
func (ec *EscrowContract) mockSelectProviders(_ context.Context, jobID [32]byte, providers [3]common.Address) (*types.Transaction, error) {
	ec.mockMu.Lock()
	defer ec.mockMu.Unlock()

	escrow, exists := ec.mockEscrows[jobID]
	if !exists {
		return nil, fmt.Errorf("escrow not found")
	}

	if escrow.State != EscrowStateCreated {
		return nil, fmt.Errorf("escrow not in created state (current: %s)", escrow.State)
	}

	escrow.Providers = providers
	escrow.State = EscrowStateActive
	escrow.StartTime = time.Now()

	logging.Info("providers selected", "job_id", fmt.Sprintf("%x", jobID[:8]),
		"provider0", providers[0].Hex(), "provider1", providers[1].Hex(), "provider2", providers[2].Hex())

	return nil, nil
}

// ReleasePayment releases payment to providers based on settled duration
func (ec *EscrowContract) ReleasePayment(ctx context.Context, jobID [32]byte, settledDuration *big.Int) (*types.Transaction, error) {
	if ec.mockMode {
		return ec.mockReleasePayment(ctx, jobID, settledDuration)
	}

	resID, err := ec.getReservationID(jobID)
	if err != nil {
		return nil, err
	}

	auth, err := ec.baseClient.GetTransactOpts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction options: %w", err)
	}

	tx, err := ec.contract.Transact(auth, "releasePayment", resID, settledDuration)
	if err != nil {
		return nil, fmt.Errorf("failed to release payment: %w", err)
	}

	return tx, nil
}

// mockReleasePayment handles payment release in mock mode
func (ec *EscrowContract) mockReleasePayment(_ context.Context, jobID [32]byte, settledDuration *big.Int) (*types.Transaction, error) {
	ec.mockMu.Lock()
	defer ec.mockMu.Unlock()

	escrow, exists := ec.mockEscrows[jobID]
	if !exists {
		return nil, fmt.Errorf("escrow not found")
	}

	if escrow.State != EscrowStateActive {
		return nil, fmt.Errorf("escrow not active")
	}

	// Calculate release amount based on settledDuration/duration ratio
	proportion := new(big.Float).Quo(
		new(big.Float).SetInt(settledDuration),
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

	logging.Info("released payment", "amount", toRelease.String(), "job_id", fmt.Sprintf("%x", jobID[:8]), "total_released", escrow.Released.String(), "total_amount", escrow.Amount.String())

	return nil, nil
}

// Refund refunds the remaining escrow to requester
func (ec *EscrowContract) Refund(ctx context.Context, jobID [32]byte) (*types.Transaction, error) {
	if ec.mockMode {
		return ec.mockRefund(ctx, jobID)
	}

	resID, err := ec.getReservationID(jobID)
	if err != nil {
		return nil, err
	}

	auth, err := ec.baseClient.GetTransactOpts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction options: %w", err)
	}

	tx, err := ec.contract.Transact(auth, "refund", resID)
	if err != nil {
		return nil, fmt.Errorf("failed to refund: %w", err)
	}

	// Clean up the local mapping — escrow is in a terminal state
	ec.removeReservationID(jobID)

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

	if escrow.State != EscrowStateActive && escrow.State != EscrowStateCreated {
		return nil, fmt.Errorf("escrow cannot be refunded (current: %s)", escrow.State)
	}

	refundAmount := new(big.Int).Sub(escrow.Amount, escrow.Released)
	escrow.State = EscrowStateRefunded

	// Clean up the local mapping — escrow is in a terminal state
	ec.reservationMu.Lock()
	delete(ec.reservationIDs, jobID)
	ec.reservationMu.Unlock()

	logging.Info("refunded escrow", "amount", refundAmount.String(), "job_id", fmt.Sprintf("%x", jobID[:8]))

	return nil, nil
}

// FinalizeEscrow finalizes an escrow (marks as complete)
func (ec *EscrowContract) FinalizeEscrow(ctx context.Context, jobID [32]byte) (*types.Transaction, error) {
	if ec.mockMode {
		return ec.mockFinalizeEscrow(ctx, jobID)
	}

	resID, err := ec.getReservationID(jobID)
	if err != nil {
		return nil, err
	}

	auth, err := ec.baseClient.GetTransactOpts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction options: %w", err)
	}

	tx, err := ec.contract.Transact(auth, "finalizeReservation", resID)
	if err != nil {
		return nil, fmt.Errorf("failed to finalize reservation: %w", err)
	}

	// Clean up the local mapping — escrow is in a terminal state
	ec.removeReservationID(jobID)

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

	// Clean up the local mapping — escrow is in a terminal state
	ec.reservationMu.Lock()
	delete(ec.reservationIDs, jobID)
	ec.reservationMu.Unlock()

	logging.Info("finalized escrow", "job_id", fmt.Sprintf("%x", jobID[:8]))

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
			JobID:         escrow.JobID,
			ReservationID: escrow.ReservationID,
			Requester:     escrow.Requester,
			Provider:      escrow.Provider,
			Providers:     escrow.Providers,
			Amount:        new(big.Int).Set(escrow.Amount),
			Released:      new(big.Int).Set(escrow.Released),
			Duration:      new(big.Int).Set(escrow.Duration),
			StartTime:     escrow.StartTime,
			State:         escrow.State,
		}, nil
	}

	resID, err := ec.getReservationID(jobID)
	if err != nil {
		return nil, err
	}

	var result []interface{}
	err = ec.contract.Call(&bind.CallOpts{Context: ctx}, &result, "getReservation", resID)
	if err != nil {
		return nil, fmt.Errorf("failed to get reservation: %w", err)
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("unexpected result format")
	}

	// The result is a struct; go-ethereum returns it as a struct value
	// We'll parse common fields from the returned tuple
	escrow := &EscrowData{
		JobID:         jobID,
		ReservationID: resID,
	}

	// go-ethereum's abi binding unpacks tuple return values
	// Try to parse the struct fields from the result
	type reservationResult struct {
		Requester      common.Address
		TotalAmount    *big.Int
		ReleasedAmount *big.Int
		Duration       *big.Int
		StartTime      *big.Int
		Status         uint8
		Providers      [3]common.Address
	}

	if res, ok := result[0].(struct {
		Requester      common.Address
		TotalAmount    *big.Int
		ReleasedAmount *big.Int
		Duration       *big.Int
		StartTime      *big.Int
		Status         uint8
		Providers      [3]common.Address
	}); ok {
		escrow.Requester = res.Requester
		escrow.Amount = res.TotalAmount
		escrow.Released = res.ReleasedAmount
		escrow.Duration = res.Duration
		if res.StartTime != nil {
			escrow.StartTime = time.Unix(res.StartTime.Int64(), 0)
		}
		escrow.State = EscrowState(res.Status)
		escrow.Providers = res.Providers
		// Use first non-zero provider as the primary provider for backwards compat
		for _, p := range res.Providers {
			if p != (common.Address{}) {
				escrow.Provider = p
				break
			}
		}
	}

	return escrow, nil
}

// GetRequesterEscrows returns all escrow job IDs for a requester
func (ec *EscrowContract) GetRequesterEscrows(_ context.Context, requester common.Address) ([][32]byte, error) {
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

	// The on-chain contract doesn't have getRequesterEscrows — return from local mapping
	ec.reservationMu.RLock()
	defer ec.reservationMu.RUnlock()
	var jobIDs [][32]byte
	for jobID := range ec.reservationIDs {
		jobIDs = append(jobIDs, jobID)
	}
	return jobIDs, nil
}

// GetProviderEscrows returns all escrow job IDs for a provider
func (ec *EscrowContract) GetProviderEscrows(_ context.Context, provider common.Address) ([][32]byte, error) {
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

	// Not directly supported by on-chain contract; return empty
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
		Topics:    [][]common.Hash{{ec.contractABI.Events["ReservationCreated"].ID}},
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
				event, err := ec.parseReservationCreatedEvent(log)
				if err != nil {
					continue
				}
				ch <- event
			case err := <-sub.Err():
				if err != nil {
					logging.Error("escrow event subscription error", "error", err)
				}
				return
			}
		}
	})

	return nil
}

// parseReservationCreatedEvent parses a ReservationCreated event from log
func (ec *EscrowContract) parseReservationCreatedEvent(log types.Log) (*EscrowCreatedEvent, error) {
	event := &EscrowCreatedEvent{
		Timestamp: time.Now(),
	}

	// ReservationCreated(uint256 indexed reservationId, address indexed requester, uint256 amount, uint256 duration)
	if len(log.Topics) > 1 {
		event.ReservationID = new(big.Int).SetBytes(log.Topics[1].Bytes())
	}
	if len(log.Topics) > 2 {
		event.Requester = common.HexToAddress(log.Topics[2].Hex())
	}

	// Parse non-indexed data (amount, duration)
	if len(log.Data) >= 64 {
		event.Amount = new(big.Int).SetBytes(log.Data[:32])
		event.Duration = new(big.Int).SetBytes(log.Data[32:64])
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

// ─── Admin Setters ───────────────────────────────────────────────────────────

// SetFeeSplit adjusts the protocol fee burn/treasury split.
func (ec *EscrowContract) SetFeeSplit(ctx context.Context, burnBps, treasuryBps uint16) (*types.Transaction, error) {
	if ec.mockMode {
		logging.Info("mock: setFeeSplit burn=%d treasury=%d", burnBps, treasuryBps)
		return nil, nil
	}
	auth, err := ec.baseClient.GetTransactOpts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction options: %w", err)
	}
	tx, err := ec.contract.Transact(auth, "setFeeSplit", burnBps, treasuryBps)
	if err != nil {
		return nil, fmt.Errorf("failed to set fee split: %w", err)
	}
	return tx, nil
}
