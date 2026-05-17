package p2p

import (
	"crypto/rand"
	"testing"
	"time"
)

func randomNonce(t *testing.T) [24]byte {
	t.Helper()
	var nonce [24]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		t.Fatalf("rand.Read: %v", err)
	}
	return nonce
}

func TestNonceTracker_ValidMessage(t *testing.T) {
	nt := NewNonceTracker()

	nonce := randomNonce(t)
	if err := nt.Check(nonce, time.Now()); err != nil {
		t.Fatalf("expected valid message to pass: %v", err)
	}
}

func TestNonceTracker_RejectZeroNonce(t *testing.T) {
	nt := NewNonceTracker()

	var zero [24]byte
	if err := nt.Check(zero, time.Now()); err == nil {
		t.Fatal("expected zero nonce to be rejected")
	}
}

func TestNonceTracker_RejectReplay(t *testing.T) {
	nt := NewNonceTracker()

	nonce := randomNonce(t)
	if err := nt.Check(nonce, time.Now()); err != nil {
		t.Fatalf("first check should pass: %v", err)
	}

	if err := nt.Check(nonce, time.Now()); err == nil {
		t.Fatal("expected replay to be rejected")
	}
}

func TestNonceTracker_RejectOldMessage(t *testing.T) {
	nt := NewNonceTracker()

	nonce := randomNonce(t)
	oldTimestamp := time.Now().Add(-10 * time.Minute)
	if err := nt.Check(nonce, oldTimestamp); err == nil {
		t.Fatal("expected old message to be rejected")
	}
}

func TestNonceTracker_RejectFutureMessage(t *testing.T) {
	nt := NewNonceTracker()

	nonce := randomNonce(t)
	futureTimestamp := time.Now().Add(2 * time.Minute)
	if err := nt.Check(nonce, futureTimestamp); err == nil {
		t.Fatal("expected future message to be rejected")
	}
}

func TestNonceTracker_AllowSlightFuture(t *testing.T) {
	nt := NewNonceTracker()

	nonce := randomNonce(t)
	// 10 seconds in the future should be OK (within 30s skew)
	if err := nt.Check(nonce, time.Now().Add(10*time.Second)); err != nil {
		t.Fatalf("expected slight future to pass: %v", err)
	}
}

func TestNonceTracker_CleanExpired(t *testing.T) {
	now := time.Now()
	nt := NewNonceTrackerWithConfig(5*time.Minute, 30*time.Second, 1*time.Minute)
	nt.nowFunc = func() time.Time { return now }

	// Add a nonce
	nonce := randomNonce(t)
	if err := nt.Check(nonce, now); err != nil {
		t.Fatalf("check failed: %v", err)
	}

	if nt.Len() != 1 {
		t.Fatalf("expected 1 tracked nonce, got %d", nt.Len())
	}

	// Advance time past the window
	nt.nowFunc = func() time.Time { return now.Add(2 * time.Minute) }
	nt.CleanExpired()

	if nt.Len() != 0 {
		t.Fatalf("expected 0 tracked nonces after cleanup, got %d", nt.Len())
	}
}

func TestNonceTracker_DifferentNoncesSameTimestamp(t *testing.T) {
	nt := NewNonceTracker()
	ts := time.Now()

	for i := 0; i < 100; i++ {
		nonce := randomNonce(t)
		if err := nt.Check(nonce, ts); err != nil {
			t.Fatalf("nonce %d should pass: %v", i, err)
		}
	}

	if nt.Len() != 100 {
		t.Fatalf("expected 100 tracked nonces, got %d", nt.Len())
	}
}
