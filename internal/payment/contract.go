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
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/moltbunker/moltbunker/internal/util"
)

// PaymentContract interacts with Base network payment contract
type PaymentContract struct {
	client     *ethclient.Client
	contract   *bind.BoundContract
	address    common.Address

	// Mock state for testing
	mockMode      bool
	mockReservations map[string]*MockReservation
	mockMu        sync.RWMutex
}

// MockReservation represents a mock reservation for testing
type MockReservation struct {
	ID        string
	Amount    *big.Int
	Duration  *big.Int
	StartTime time.Time
	Provider  common.Address
}

// NewPaymentContract creates a new payment contract client
func NewPaymentContract(client *ethclient.Client, contractAddress common.Address, abiJSON string) (*PaymentContract, error) {
	pc := &PaymentContract{
		client:           client,
		address:          contractAddress,
		mockReservations: make(map[string]*MockReservation),
	}

	// If no client provided, use mock mode
	if client == nil {
		pc.mockMode = true
		return pc, nil
	}

	parsedABI, err := abi.JSON(strings.NewReader(abiJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to parse ABI: %w", err)
	}

	pc.contract = bind.NewBoundContract(contractAddress, parsedABI, client, client, client)

	return pc, nil
}

// NewMockPaymentContract creates a mock payment contract for testing
func NewMockPaymentContract() *PaymentContract {
	return &PaymentContract{
		mockMode:         true,
		mockReservations: make(map[string]*MockReservation),
	}
}

// ReserveRuntime reserves runtime and pays BUNKER tokens
func (pc *PaymentContract) ReserveRuntime(ctx context.Context, auth *bind.TransactOpts, amount *big.Int, duration *big.Int) (*types.Transaction, error) {
	// Mock mode: track reservation in memory
	if pc.mockMode {
		pc.mockMu.Lock()
		defer pc.mockMu.Unlock()

		reservationID := fmt.Sprintf("res-%d", time.Now().UnixNano())
		pc.mockReservations[reservationID] = &MockReservation{
			ID:        reservationID,
			Amount:    new(big.Int).Set(amount),
			Duration:  new(big.Int).Set(duration),
			StartTime: time.Now(),
		}

		fmt.Printf("[MOCK] Reserved runtime: ID=%s, Amount=%s, Duration=%s\n",
			reservationID, amount.String(), duration.String())

		// Return nil transaction in mock mode
		return nil, nil
	}

	// Production mode: call actual contract
	tx, err := pc.contract.Transact(auth, "reserveRuntime", amount, duration)
	if err != nil {
		return nil, fmt.Errorf("failed to reserve runtime: %w", err)
	}

	return tx, nil
}

// GetMockReservation returns a mock reservation by ID (for testing)
func (pc *PaymentContract) GetMockReservation(id string) (*MockReservation, bool) {
	pc.mockMu.RLock()
	defer pc.mockMu.RUnlock()
	res, exists := pc.mockReservations[id]
	return res, exists
}

// IsMockMode returns whether the contract is in mock mode
func (pc *PaymentContract) IsMockMode() bool {
	return pc.mockMode
}

// ListenReservationEvents listens for reservation events
func (pc *PaymentContract) ListenReservationEvents(ctx context.Context, ch chan<- *ReservationEvent) error {
	query := ethereum.FilterQuery{
		Addresses: []common.Address{pc.address},
	}

	logs := make(chan types.Log)
	sub, err := pc.client.SubscribeFilterLogs(ctx, query, logs)
	if err != nil {
		return fmt.Errorf("failed to subscribe to logs: %w", err)
	}

	util.SafeGoWithName("reservation-event-listener", func() {
		for {
			select {
			case <-ctx.Done():
				return
			case log := <-logs:
				// Parse reservation event
				event := &ReservationEvent{
					ReservationID: log.Topics[1].Hex(),
					Amount:        big.NewInt(0), // Parse from log data
					Duration:      big.NewInt(0), // Parse from log data
				}
				ch <- event
			case err := <-sub.Err():
				// Handle error
				_ = err
			}
		}
	})

	return nil
}

// ReservationEvent represents a reservation event
type ReservationEvent struct {
	ReservationID string
	Amount        *big.Int
	Duration      *big.Int
}
