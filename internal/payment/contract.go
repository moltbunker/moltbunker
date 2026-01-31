package payment

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// PaymentContract interacts with Base network payment contract
type PaymentContract struct {
	client     *ethclient.Client
	contract   *bind.BoundContract
	address    common.Address
}

// NewPaymentContract creates a new payment contract client
func NewPaymentContract(client *ethclient.Client, contractAddress common.Address, abiJSON string) (*PaymentContract, error) {
	parsedABI, err := abi.JSON(strings.NewReader(abiJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to parse ABI: %w", err)
	}

	contract := bind.NewBoundContract(contractAddress, parsedABI, client, client, client)

	return &PaymentContract{
		client:   client,
		contract: contract,
		address:  contractAddress,
	}, nil
}

// ReserveRuntime reserves runtime and pays BUNKER tokens
func (pc *PaymentContract) ReserveRuntime(ctx context.Context, auth *bind.TransactOpts, amount *big.Int, duration *big.Int) (*types.Transaction, error) {
	// Call contract method to reserve runtime
	// This is a placeholder - actual implementation would call the contract method
	tx, err := pc.contract.Transact(auth, "reserveRuntime", amount, duration)
	if err != nil {
		return nil, fmt.Errorf("failed to reserve runtime: %w", err)
	}

	return tx, nil
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

	go func() {
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
	}()

	return nil
}

// ReservationEvent represents a reservation event
type ReservationEvent struct {
	ReservationID string
	Amount        *big.Int
	Duration      *big.Int
}
