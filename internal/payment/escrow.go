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

// ── Molt Prepaid Credits ────────────────────────────────────────────────────
// Unlike containers (per-job escrow), Molt functions use a prepaid credit model.
// The requester deposits tokens once and each invocation deducts from the balance.
// This avoids per-invocation on-chain transactions.

// MoltCreditManager manages prepaid invocation credits for Molt functions.
type MoltCreditManager struct {
	credits map[string]*MoltCredit // keyed by requester address (lowercase hex)
	mu      sync.RWMutex
}

// MoltCredit tracks a requester's prepaid balance for Molt invocations.
type MoltCredit struct {
	RequesterAddress string
	Balance          *big.Int  // Remaining credit (BUNKER wei)
	TotalDeposited   *big.Int  // Lifetime deposits
	TotalSpent       *big.Int  // Lifetime spend
	CreatedAt        time.Time
	LastInvocation   time.Time
	mu               sync.Mutex
}

// NewMoltCreditManager creates a new Molt credit manager.
func NewMoltCreditManager() *MoltCreditManager {
	return &MoltCreditManager{
		credits: make(map[string]*MoltCredit),
	}
}

// Deposit adds credits for a requester.
func (m *MoltCreditManager) Deposit(requesterAddress string, amount *big.Int) {
	m.mu.Lock()
	credit, exists := m.credits[requesterAddress]
	if !exists {
		credit = &MoltCredit{
			RequesterAddress: requesterAddress,
			Balance:          new(big.Int),
			TotalDeposited:   new(big.Int),
			TotalSpent:       new(big.Int),
			CreatedAt:        time.Now(),
		}
		m.credits[requesterAddress] = credit
	}
	m.mu.Unlock()

	credit.mu.Lock()
	credit.Balance.Add(credit.Balance, amount)
	credit.TotalDeposited.Add(credit.TotalDeposited, amount)
	credit.mu.Unlock()
}

// Deduct subtracts a cost from the requester's credit balance.
// Returns an error if insufficient funds.
func (m *MoltCreditManager) Deduct(requesterAddress string, cost *big.Int) error {
	m.mu.RLock()
	credit, exists := m.credits[requesterAddress]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("no Molt credits for %s", requesterAddress)
	}

	credit.mu.Lock()
	defer credit.mu.Unlock()

	if credit.Balance.Cmp(cost) < 0 {
		return fmt.Errorf("insufficient Molt credits: have %s, need %s", credit.Balance.String(), cost.String())
	}

	credit.Balance.Sub(credit.Balance, cost)
	credit.TotalSpent.Add(credit.TotalSpent, cost)
	credit.LastInvocation = time.Now()

	return nil
}

// GetBalance returns the current credit balance for a requester.
func (m *MoltCreditManager) GetBalance(requesterAddress string) *big.Int {
	m.mu.RLock()
	credit, exists := m.credits[requesterAddress]
	m.mu.RUnlock()

	if !exists {
		return big.NewInt(0)
	}

	credit.mu.Lock()
	defer credit.mu.Unlock()

	return new(big.Int).Set(credit.Balance)
}

// GetCredit returns credit details for a requester, or nil if none exists.
func (m *MoltCreditManager) GetCredit(requesterAddress string) *MoltCredit {
	m.mu.RLock()
	credit, exists := m.credits[requesterAddress]
	m.mu.RUnlock()

	if !exists {
		return nil
	}
	return credit
}

// RefundAll refunds the remaining balance and returns the amount refunded.
func (m *MoltCreditManager) RefundAll(requesterAddress string) *big.Int {
	m.mu.Lock()
	credit, exists := m.credits[requesterAddress]
	if !exists {
		m.mu.Unlock()
		return big.NewInt(0)
	}
	delete(m.credits, requesterAddress)
	m.mu.Unlock()

	credit.mu.Lock()
	defer credit.mu.Unlock()

	refund := new(big.Int).Set(credit.Balance)
	credit.Balance.SetInt64(0)
	return refund
}
