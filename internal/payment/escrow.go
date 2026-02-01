package payment

import (
	"fmt"
	"math/big"
	"sync"
	"time"
)

// EscrowManager manages payment escrow
type EscrowManager struct {
	escrows map[string]*Escrow
	mu      sync.RWMutex
}

// Escrow represents an escrow account
type Escrow struct {
	ReservationID string
	Amount        *big.Int
	Duration      time.Duration
	StartTime     time.Time
	Released      *big.Int
	mu            sync.RWMutex
}

// NewEscrowManager creates a new escrow manager
func NewEscrowManager() *EscrowManager {
	return &EscrowManager{
		escrows: make(map[string]*Escrow),
	}
}

// CreateEscrow creates a new escrow
func (em *EscrowManager) CreateEscrow(reservationID string, amount *big.Int, duration time.Duration) *Escrow {
	escrow := &Escrow{
		ReservationID: reservationID,
		Amount:        amount,
		Duration:      duration,
		StartTime:     time.Now(),
		Released:      big.NewInt(0),
	}

	em.mu.Lock()
	em.escrows[reservationID] = escrow
	em.mu.Unlock()

	return escrow
}

// GetEscrow retrieves an escrow
func (em *EscrowManager) GetEscrow(reservationID string) (*Escrow, bool) {
	em.mu.RLock()
	defer em.mu.RUnlock()

	escrow, exists := em.escrows[reservationID]
	return escrow, exists
}

// ReleasePayment releases incremental payment based on uptime
func (em *EscrowManager) ReleasePayment(reservationID string, uptime time.Duration) (*big.Int, error) {
	em.mu.RLock()
	escrow, exists := em.escrows[reservationID]
	em.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("escrow not found: %s", reservationID)
	}

	escrow.mu.Lock()
	defer escrow.mu.Unlock()

	// Use the provided uptime parameter for calculation
	// Cap at the escrow duration
	elapsed := uptime
	if elapsed > escrow.Duration {
		elapsed = escrow.Duration
	}

	// Calculate proportional payment based on uptime / duration
	proportion := new(big.Float).Quo(
		new(big.Float).SetInt64(int64(elapsed)),
		new(big.Float).SetInt64(int64(escrow.Duration)),
	)

	amountFloat := new(big.Float).SetInt(escrow.Amount)
	releaseFloat := new(big.Float).Mul(amountFloat, proportion)
	releaseAmount, _ := releaseFloat.Int(nil)

	// Subtract already released amount
	toRelease := new(big.Int).Sub(releaseAmount, escrow.Released)
	if toRelease.Sign() <= 0 {
		return big.NewInt(0), nil
	}

	escrow.Released.Add(escrow.Released, toRelease)
	return toRelease, nil
}
