package p2p

import (
	"fmt"
	"sync"
	"time"

	"github.com/moltbunker/moltbunker/internal/logging"
)

const (
	// DefaultMaxMessageAge is the maximum age of a message before it's rejected
	DefaultMaxMessageAge = 5 * time.Minute

	// DefaultMaxFutureSkew is the maximum time a message can be in the future
	DefaultMaxFutureSkew = 30 * time.Second

	// DefaultNonceWindow is how long seen nonces are retained
	DefaultNonceWindow = 10 * time.Minute

	// DefaultNonceCleanupInterval is how often expired nonces are cleaned up
	DefaultNonceCleanupInterval = 60 * time.Second
)

// NonceTracker deduplicates messages using nonce + timestamp validation.
// It rejects messages that are too old, too far in the future, or have
// already been seen (replay protection).
type NonceTracker struct {
	mu          sync.RWMutex
	seen        map[[24]byte]time.Time
	maxAge      time.Duration // reject messages older than this
	futureSkew  time.Duration // reject messages more than this in the future
	nonceWindow time.Duration // how long to remember seen nonces
	nowFunc     func() time.Time
}

// NewNonceTracker creates a new nonce tracker with default settings.
func NewNonceTracker() *NonceTracker {
	return &NonceTracker{
		seen:        make(map[[24]byte]time.Time),
		maxAge:      DefaultMaxMessageAge,
		futureSkew:  DefaultMaxFutureSkew,
		nonceWindow: DefaultNonceWindow,
		nowFunc:     time.Now,
	}
}

// NewNonceTrackerWithConfig creates a nonce tracker with custom settings.
func NewNonceTrackerWithConfig(maxAge, futureSkew, nonceWindow time.Duration) *NonceTracker {
	return &NonceTracker{
		seen:        make(map[[24]byte]time.Time),
		maxAge:      maxAge,
		futureSkew:  futureSkew,
		nonceWindow: nonceWindow,
		nowFunc:     time.Now,
	}
}

// Check validates a nonce + timestamp pair. Returns nil if valid, error if rejected.
// A zero nonce (all bytes zero) is always rejected — callers must generate a random nonce.
func (nt *NonceTracker) Check(nonce [24]byte, timestamp time.Time) error {
	now := nt.nowFunc()

	// Reject zero nonce
	var zero [24]byte
	if nonce == zero {
		return fmt.Errorf("zero nonce not allowed")
	}

	// Reject messages too far in the past
	if now.Sub(timestamp) > nt.maxAge {
		return fmt.Errorf("message too old: %s ago (max %s)", now.Sub(timestamp).Round(time.Second), nt.maxAge)
	}

	// Reject messages too far in the future
	if timestamp.Sub(now) > nt.futureSkew {
		return fmt.Errorf("message timestamp too far in future: %s ahead (max %s)", timestamp.Sub(now).Round(time.Second), nt.futureSkew)
	}

	// Check for replay
	nt.mu.Lock()
	defer nt.mu.Unlock()

	if _, exists := nt.seen[nonce]; exists {
		return fmt.Errorf("duplicate nonce (replay detected)")
	}

	// Record this nonce
	nt.seen[nonce] = now
	return nil
}

// CleanExpired removes nonces older than the nonce window.
func (nt *NonceTracker) CleanExpired() {
	nt.mu.Lock()
	defer nt.mu.Unlock()

	cutoff := nt.nowFunc().Add(-nt.nonceWindow)
	removed := 0
	for nonce, seenAt := range nt.seen {
		if seenAt.Before(cutoff) {
			delete(nt.seen, nonce)
			removed++
		}
	}

	if removed > 0 {
		logging.Debug("cleaned expired nonces",
			"removed", removed,
			"remaining", len(nt.seen),
			logging.Component("nonce_tracker"))
	}
}

// Len returns the number of tracked nonces.
func (nt *NonceTracker) Len() int {
	nt.mu.RLock()
	defer nt.mu.RUnlock()
	return len(nt.seen)
}
