package p2p

import (
	"testing"
	"time"
)

func TestCircuitBreaker_InitialState(t *testing.T) {
	cb := NewCircuitBreaker(nil)

	if cb.State() != CircuitClosed {
		t.Errorf("expected initial state to be Closed, got %s", cb.State())
	}

	if !cb.Allow() {
		t.Error("expected Allow() to return true in closed state")
	}
}

func TestCircuitBreaker_OpensAfterFailures(t *testing.T) {
	config := &CircuitBreakerConfig{
		FailureThreshold:    3,
		SuccessThreshold:    2,
		Timeout:             100 * time.Millisecond,
		HalfOpenMaxRequests: 2,
	}
	cb := NewCircuitBreaker(config)

	// Record failures up to threshold
	for i := 0; i < 3; i++ {
		cb.RecordFailure()
	}

	if cb.State() != CircuitOpen {
		t.Errorf("expected state to be Open after %d failures, got %s", 3, cb.State())
	}

	if cb.Allow() {
		t.Error("expected Allow() to return false in open state")
	}
}

func TestCircuitBreaker_TransitionsToHalfOpen(t *testing.T) {
	config := &CircuitBreakerConfig{
		FailureThreshold:    2,
		SuccessThreshold:    1,
		Timeout:             50 * time.Millisecond,
		HalfOpenMaxRequests: 2,
	}
	cb := NewCircuitBreaker(config)

	// Open the circuit
	cb.RecordFailure()
	cb.RecordFailure()

	if cb.State() != CircuitOpen {
		t.Errorf("expected Open state, got %s", cb.State())
	}

	// Wait for timeout
	time.Sleep(60 * time.Millisecond)

	// Should transition to half-open on next Allow()
	if !cb.Allow() {
		t.Error("expected Allow() to return true after timeout (half-open)")
	}

	if cb.State() != CircuitHalfOpen {
		t.Errorf("expected HalfOpen state, got %s", cb.State())
	}
}

func TestCircuitBreaker_ClosesAfterSuccesses(t *testing.T) {
	config := &CircuitBreakerConfig{
		FailureThreshold:    2,
		SuccessThreshold:    2,
		Timeout:             50 * time.Millisecond,
		HalfOpenMaxRequests: 3,
	}
	cb := NewCircuitBreaker(config)

	// Open the circuit
	cb.RecordFailure()
	cb.RecordFailure()

	// Wait for timeout and transition to half-open
	time.Sleep(60 * time.Millisecond)
	cb.Allow()

	// Record successes in half-open state
	cb.RecordSuccess()
	if cb.State() != CircuitHalfOpen {
		t.Error("expected to still be half-open after 1 success")
	}

	cb.RecordSuccess()
	if cb.State() != CircuitClosed {
		t.Errorf("expected Closed state after 2 successes, got %s", cb.State())
	}
}

func TestCircuitBreaker_ReopensOnFailureInHalfOpen(t *testing.T) {
	config := &CircuitBreakerConfig{
		FailureThreshold:    2,
		SuccessThreshold:    2,
		Timeout:             50 * time.Millisecond,
		HalfOpenMaxRequests: 3,
	}
	cb := NewCircuitBreaker(config)

	// Open the circuit
	cb.RecordFailure()
	cb.RecordFailure()

	// Wait for timeout and transition to half-open
	time.Sleep(60 * time.Millisecond)
	cb.Allow()

	// Failure in half-open should reopen
	cb.RecordFailure()

	if cb.State() != CircuitOpen {
		t.Errorf("expected Open state after failure in half-open, got %s", cb.State())
	}
}

func TestCircuitBreaker_HalfOpenLimitsRequests(t *testing.T) {
	config := &CircuitBreakerConfig{
		FailureThreshold:    2,
		SuccessThreshold:    2,
		Timeout:             50 * time.Millisecond,
		HalfOpenMaxRequests: 2,
	}
	cb := NewCircuitBreaker(config)

	// Open the circuit
	cb.RecordFailure()
	cb.RecordFailure()

	// Wait for timeout
	time.Sleep(60 * time.Millisecond)

	// Should allow HalfOpenMaxRequests
	for i := 0; i < 2; i++ {
		if !cb.Allow() {
			t.Errorf("expected Allow() to return true for request %d in half-open", i+1)
		}
	}

	// Should block additional requests
	if cb.Allow() {
		t.Error("expected Allow() to return false when half-open limit reached")
	}
}

func TestCircuitBreaker_Reset(t *testing.T) {
	config := &CircuitBreakerConfig{
		FailureThreshold: 2,
		Timeout:          time.Hour, // Long timeout
	}
	cb := NewCircuitBreaker(config)

	// Open the circuit
	cb.RecordFailure()
	cb.RecordFailure()

	if cb.State() != CircuitOpen {
		t.Error("expected Open state")
	}

	// Reset
	cb.Reset()

	if cb.State() != CircuitClosed {
		t.Errorf("expected Closed state after reset, got %s", cb.State())
	}

	if !cb.Allow() {
		t.Error("expected Allow() to return true after reset")
	}
}

func TestCircuitBreaker_SuccessResetsFailureCount(t *testing.T) {
	config := &CircuitBreakerConfig{
		FailureThreshold: 3,
	}
	cb := NewCircuitBreaker(config)

	// Record some failures
	cb.RecordFailure()
	cb.RecordFailure()

	// Record a success - should reset failure count
	cb.RecordSuccess()

	// Record another failure - should not open circuit
	cb.RecordFailure()

	if cb.State() != CircuitClosed {
		t.Errorf("expected Closed state (failure count reset), got %s", cb.State())
	}
}

func TestPeerCircuitBreakers_Basic(t *testing.T) {
	pcb := NewPeerCircuitBreakers(&CircuitBreakerConfig{
		FailureThreshold: 2,
		Timeout:          time.Hour, // Long timeout to keep circuit open
	})

	peer1 := "peer-1"
	peer2 := "peer-2"

	// Both peers should allow initially
	if !pcb.AllowRequest(peer1) {
		t.Error("expected AllowRequest to return true for peer1")
	}
	if !pcb.AllowRequest(peer2) {
		t.Error("expected AllowRequest to return true for peer2")
	}

	// Open circuit for peer1
	pcb.RecordFailure(peer1)
	pcb.RecordFailure(peer1)

	// peer1 should be blocked, peer2 should still be allowed
	if pcb.AllowRequest(peer1) {
		t.Error("expected AllowRequest to return false for peer1 (circuit open)")
	}
	if !pcb.AllowRequest(peer2) {
		t.Error("expected AllowRequest to return true for peer2 (independent)")
	}
}

func TestPeerCircuitBreakers_GetOpenCircuits(t *testing.T) {
	pcb := NewPeerCircuitBreakers(&CircuitBreakerConfig{
		FailureThreshold: 1,
	})

	// Open circuits for peer1 and peer3
	pcb.RecordFailure("peer-1")
	pcb.RecordSuccess("peer-2") // Just touch peer-2
	pcb.RecordFailure("peer-3")

	open := pcb.GetOpenCircuits()

	if len(open) != 2 {
		t.Errorf("expected 2 open circuits, got %d", len(open))
	}

	// Check that peer-2 is not in the list
	for _, id := range open {
		if id == "peer-2" {
			t.Error("peer-2 should not have an open circuit")
		}
	}
}

func TestPeerCircuitBreakers_Remove(t *testing.T) {
	pcb := NewPeerCircuitBreakers(nil)

	peer := "peer-1"
	pcb.RecordFailure(peer)

	stats := pcb.GetStats()
	if _, exists := stats[peer]; !exists {
		t.Error("expected peer to exist in stats")
	}

	pcb.Remove(peer)

	stats = pcb.GetStats()
	if _, exists := stats[peer]; exists {
		t.Error("expected peer to be removed from stats")
	}
}

func TestCircuitState_String(t *testing.T) {
	tests := []struct {
		state    CircuitState
		expected string
	}{
		{CircuitClosed, "closed"},
		{CircuitOpen, "open"},
		{CircuitHalfOpen, "half-open"},
		{CircuitState(99), "unknown"},
	}

	for _, tt := range tests {
		if got := tt.state.String(); got != tt.expected {
			t.Errorf("CircuitState(%d).String() = %q, want %q", tt.state, got, tt.expected)
		}
	}
}
