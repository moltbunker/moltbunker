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
