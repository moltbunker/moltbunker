package payment

import (
	"math/big"
	"testing"
	"time"
)

func TestEscrowManager_CreateEscrow(t *testing.T) {
	em := NewEscrowManager()

	reservationID := "test-reservation"
	amount := big.NewInt(1000000000000000000) // 1 BUNKER
	duration := 1 * time.Hour

	escrow := em.CreateEscrow(reservationID, amount, duration)

	if escrow.ReservationID != reservationID {
		t.Errorf("ReservationID mismatch: got %s, want %s", escrow.ReservationID, reservationID)
	}

	if escrow.Amount.Cmp(amount) != 0 {
		t.Error("Amount mismatch")
	}

	if escrow.Duration != duration {
		t.Error("Duration mismatch")
	}
}

func TestEscrowManager_GetEscrow(t *testing.T) {
	em := NewEscrowManager()

	reservationID := "test-reservation"
	amount := big.NewInt(1000000000000000000)
	duration := 1 * time.Hour

	em.CreateEscrow(reservationID, amount, duration)

	escrow, exists := em.GetEscrow(reservationID)
	if !exists {
		t.Fatal("Escrow should exist")
	}

	if escrow.ReservationID != reservationID {
		t.Error("Escrow ID mismatch")
	}
}

func TestEscrowManager_GetEscrow_NotExists(t *testing.T) {
	em := NewEscrowManager()

	_, exists := em.GetEscrow("nonexistent")
	if exists {
		t.Error("Escrow should not exist")
	}
}

func TestEscrowManager_ReleasePayment(t *testing.T) {
	em := NewEscrowManager()

	reservationID := "test-reservation"
	amount := big.NewInt(1000000000000000000) // 1 BUNKER
	duration := 1 * time.Hour

	em.CreateEscrow(reservationID, amount, duration)

	// Release payment after 30 minutes
	uptime := 30 * time.Minute
	released, err := em.ReleasePayment(reservationID, uptime)
	if err != nil {
		t.Fatalf("Failed to release payment: %v", err)
	}

	if released.Sign() <= 0 {
		t.Error("Released amount should be positive")
	}

	// Should release approximately half (30 min / 60 min)
	expectedMin := new(big.Int).Div(amount, big.NewInt(2))
	expectedMin.Sub(expectedMin, big.NewInt(100000000000000)) // Allow some margin

	if released.Cmp(expectedMin) < 0 {
		t.Error("Released amount seems too low")
	}
}

func TestEscrowManager_ReleasePayment_FullDuration(t *testing.T) {
	em := NewEscrowManager()

	reservationID := "test-reservation"
	amount := big.NewInt(1000000000000000000)
	duration := 1 * time.Hour

	em.CreateEscrow(reservationID, amount, duration)

	// Release payment after full duration
	released, err := em.ReleasePayment(reservationID, duration)
	if err != nil {
		t.Fatalf("Failed to release payment: %v", err)
	}

	// Should release close to full amount
	expectedMin := new(big.Int).Mul(amount, big.NewInt(95))
	expectedMin.Div(expectedMin, big.NewInt(100)) // 95% of amount

	if released.Cmp(expectedMin) < 0 {
		t.Error("Released amount should be close to full amount")
	}
}

func TestEscrowManager_ReleasePayment_NotExists(t *testing.T) {
	em := NewEscrowManager()

	_, err := em.ReleasePayment("nonexistent", 1*time.Hour)
	if err == nil {
		t.Error("Should fail for nonexistent escrow")
	}
}

// --- Molt Credit Manager Tests ---

func TestMoltCreditManager_Deposit(t *testing.T) {
	m := NewMoltCreditManager()

	m.Deposit("0xabc", big.NewInt(1000))

	bal := m.GetBalance("0xabc")
	if bal.Cmp(big.NewInt(1000)) != 0 {
		t.Fatalf("balance = %s, want 1000", bal)
	}
}

func TestMoltCreditManager_DepositAdditive(t *testing.T) {
	m := NewMoltCreditManager()

	m.Deposit("0xabc", big.NewInt(500))
	m.Deposit("0xabc", big.NewInt(300))

	bal := m.GetBalance("0xabc")
	if bal.Cmp(big.NewInt(800)) != 0 {
		t.Fatalf("balance = %s, want 800", bal)
	}
}

func TestMoltCreditManager_Deduct(t *testing.T) {
	m := NewMoltCreditManager()
	m.Deposit("0xabc", big.NewInt(1000))

	err := m.Deduct("0xabc", big.NewInt(400))
	if err != nil {
		t.Fatalf("Deduct: %v", err)
	}

	bal := m.GetBalance("0xabc")
	if bal.Cmp(big.NewInt(600)) != 0 {
		t.Fatalf("balance = %s, want 600", bal)
	}
}

func TestMoltCreditManager_DeductInsufficientFunds(t *testing.T) {
	m := NewMoltCreditManager()
	m.Deposit("0xabc", big.NewInt(100))

	err := m.Deduct("0xabc", big.NewInt(200))
	if err == nil {
		t.Fatal("expected error for insufficient funds")
	}

	// Balance should be unchanged
	bal := m.GetBalance("0xabc")
	if bal.Cmp(big.NewInt(100)) != 0 {
		t.Fatalf("balance should be unchanged: got %s, want 100", bal)
	}
}

func TestMoltCreditManager_DeductUnknownRequester(t *testing.T) {
	m := NewMoltCreditManager()

	err := m.Deduct("0xunknown", big.NewInt(1))
	if err == nil {
		t.Fatal("expected error for unknown requester")
	}
}

func TestMoltCreditManager_GetBalanceUnknown(t *testing.T) {
	m := NewMoltCreditManager()

	bal := m.GetBalance("0xunknown")
	if bal.Sign() != 0 {
		t.Fatalf("balance should be 0 for unknown, got %s", bal)
	}
}

func TestMoltCreditManager_RefundAll(t *testing.T) {
	m := NewMoltCreditManager()
	m.Deposit("0xabc", big.NewInt(1000))
	if err := m.Deduct("0xabc", big.NewInt(300)); err != nil {
		t.Fatalf("Deduct: %v", err)
	}

	refund := m.RefundAll("0xabc")
	if refund.Cmp(big.NewInt(700)) != 0 {
		t.Fatalf("refund = %s, want 700", refund)
	}

	// After refund, balance should be 0 and requester removed
	bal := m.GetBalance("0xabc")
	if bal.Sign() != 0 {
		t.Fatalf("balance after refund should be 0, got %s", bal)
	}
}

func TestMoltCreditManager_RefundUnknown(t *testing.T) {
	m := NewMoltCreditManager()

	refund := m.RefundAll("0xunknown")
	if refund.Sign() != 0 {
		t.Fatalf("refund for unknown should be 0, got %s", refund)
	}
}

func TestMoltCreditManager_GetCredit(t *testing.T) {
	m := NewMoltCreditManager()
	m.Deposit("0xabc", big.NewInt(500))
	if err := m.Deduct("0xabc", big.NewInt(100)); err != nil {
		t.Fatalf("Deduct: %v", err)
	}

	credit := m.GetCredit("0xabc")
	if credit == nil {
		t.Fatal("expected non-nil credit")
	}
	if credit.RequesterAddress != "0xabc" {
		t.Fatalf("RequesterAddress = %s, want 0xabc", credit.RequesterAddress)
	}
	if credit.TotalDeposited.Cmp(big.NewInt(500)) != 0 {
		t.Fatalf("TotalDeposited = %s, want 500", credit.TotalDeposited)
	}
	if credit.TotalSpent.Cmp(big.NewInt(100)) != 0 {
		t.Fatalf("TotalSpent = %s, want 100", credit.TotalSpent)
	}
}

func TestMoltCreditManager_GetCreditUnknown(t *testing.T) {
	m := NewMoltCreditManager()

	credit := m.GetCredit("0xunknown")
	if credit != nil {
		t.Fatal("expected nil for unknown requester")
	}
}

func TestMoltCreditManager_DepletionFlow(t *testing.T) {
	m := NewMoltCreditManager()
	m.Deposit("0xabc", big.NewInt(10))

	// Deduct 10 times (1 each) — should all succeed
	for i := 0; i < 10; i++ {
		if err := m.Deduct("0xabc", big.NewInt(1)); err != nil {
			t.Fatalf("deduction %d failed: %v", i, err)
		}
	}

	// 11th deduction should fail
	if err := m.Deduct("0xabc", big.NewInt(1)); err == nil {
		t.Fatal("expected error on depleted credits")
	}

	bal := m.GetBalance("0xabc")
	if bal.Sign() != 0 {
		t.Fatalf("balance should be 0 after depletion, got %s", bal)
	}
}

func TestEscrowManager_ReleasePayment_Incremental(t *testing.T) {
	em := NewEscrowManager()

	reservationID := "test-reservation"
	amount := big.NewInt(1000000000000000000)
	duration := 1 * time.Hour

	em.CreateEscrow(reservationID, amount, duration)

	// First release after 20 minutes
	released1, err := em.ReleasePayment(reservationID, 20*time.Minute)
	if err != nil {
		t.Fatalf("Failed to release payment: %v", err)
	}

	// Second release after additional 20 minutes (total 40 minutes)
	released2, err := em.ReleasePayment(reservationID, 40*time.Minute)
	if err != nil {
		t.Fatalf("Failed to release payment: %v", err)
	}

	// Both releases should be positive (incremental releases happening)
	if released1.Sign() <= 0 {
		t.Error("First release should be positive")
	}
	if released2.Sign() <= 0 {
		t.Error("Second release should be positive (accounting for already released)")
	}

	// Total released should be approximately 40/60 = 2/3 of the amount
	totalReleased := new(big.Int).Add(released1, released2)
	expectedMin := new(big.Int).Mul(amount, big.NewInt(60))
	expectedMin.Div(expectedMin, big.NewInt(100)) // 60% of amount (with margin for 66.67%)

	if totalReleased.Cmp(expectedMin) < 0 {
		t.Error("Total released should be approximately 2/3 of escrow amount")
	}
}
