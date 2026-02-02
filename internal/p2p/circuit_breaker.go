package p2p

import (
	"sync"
	"time"
)

// CircuitState represents the state of a circuit breaker
type CircuitState int

const (
	// CircuitClosed - circuit is healthy, requests flow through
	CircuitClosed CircuitState = iota
	// CircuitOpen - circuit has tripped, requests are blocked
	CircuitOpen
	// CircuitHalfOpen - circuit is testing if the target has recovered
	CircuitHalfOpen
)

func (s CircuitState) String() string {
	switch s {
	case CircuitClosed:
		return "closed"
	case CircuitOpen:
		return "open"
	case CircuitHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// CircuitBreakerConfig holds configuration for a circuit breaker
type CircuitBreakerConfig struct {
	// FailureThreshold is the number of failures before the circuit opens
	FailureThreshold int
	// SuccessThreshold is the number of successes in half-open state before closing
	SuccessThreshold int
	// Timeout is how long the circuit stays open before moving to half-open
	Timeout time.Duration
	// HalfOpenMaxRequests is max requests allowed in half-open state
	HalfOpenMaxRequests int
}

// DefaultCircuitBreakerConfig returns sensible defaults
func DefaultCircuitBreakerConfig() *CircuitBreakerConfig {
	return &CircuitBreakerConfig{
		FailureThreshold:    5,
		SuccessThreshold:    2,
		Timeout:             30 * time.Second,
		HalfOpenMaxRequests: 3,
	}
}

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	config       *CircuitBreakerConfig
	state        CircuitState
	failures     int
	successes    int
	lastFailure  time.Time
	halfOpenReqs int
	mu           sync.RWMutex
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(config *CircuitBreakerConfig) *CircuitBreaker {
	if config == nil {
		config = DefaultCircuitBreakerConfig()
	}
	return &CircuitBreaker{
		config: config,
		state:  CircuitClosed,
	}
}

// Allow checks if a request should be allowed through
func (cb *CircuitBreaker) Allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case CircuitClosed:
		return true

	case CircuitOpen:
		// Check if timeout has elapsed
		if time.Since(cb.lastFailure) >= cb.config.Timeout {
			// Move to half-open state
			cb.state = CircuitHalfOpen
			cb.halfOpenReqs = 1 // This request counts as the first half-open request
			cb.successes = 0
			return true
		}
		return false

	case CircuitHalfOpen:
		// Allow limited requests in half-open state
		if cb.halfOpenReqs < cb.config.HalfOpenMaxRequests {
			cb.halfOpenReqs++
			return true
		}
		return false

	default:
		return false
	}
}

// RecordSuccess records a successful operation
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case CircuitClosed:
		// Reset failure count on success
		cb.failures = 0

	case CircuitHalfOpen:
		cb.successes++
		// If we've had enough successes, close the circuit
		if cb.successes >= cb.config.SuccessThreshold {
			cb.state = CircuitClosed
			cb.failures = 0
			cb.successes = 0
		}
	}
}

// RecordFailure records a failed operation
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.lastFailure = time.Now()

	switch cb.state {
	case CircuitClosed:
		cb.failures++
		// If we've hit the threshold, open the circuit
		if cb.failures >= cb.config.FailureThreshold {
			cb.state = CircuitOpen
		}

	case CircuitHalfOpen:
		// Any failure in half-open state opens the circuit again
		cb.state = CircuitOpen
		cb.successes = 0
	}
}

// State returns the current state of the circuit breaker
func (cb *CircuitBreaker) State() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// Stats returns circuit breaker statistics
func (cb *CircuitBreaker) Stats() CircuitBreakerStats {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return CircuitBreakerStats{
		State:        cb.state,
		Failures:     cb.failures,
		Successes:    cb.successes,
		LastFailure:  cb.lastFailure,
		HalfOpenReqs: cb.halfOpenReqs,
	}
}

// CircuitBreakerStats holds circuit breaker statistics
type CircuitBreakerStats struct {
	State        CircuitState
	Failures     int
	Successes    int
	LastFailure  time.Time
	HalfOpenReqs int
}

// Reset resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.state = CircuitClosed
	cb.failures = 0
	cb.successes = 0
	cb.halfOpenReqs = 0
}

// PeerCircuitBreakers manages circuit breakers for multiple peers
type PeerCircuitBreakers struct {
	breakers map[string]*CircuitBreaker
	config   *CircuitBreakerConfig
	mu       sync.RWMutex
}

// NewPeerCircuitBreakers creates a new peer circuit breaker manager
func NewPeerCircuitBreakers(config *CircuitBreakerConfig) *PeerCircuitBreakers {
	if config == nil {
		config = DefaultCircuitBreakerConfig()
	}
	return &PeerCircuitBreakers{
		breakers: make(map[string]*CircuitBreaker),
		config:   config,
	}
}

// Get returns the circuit breaker for a peer, creating one if needed
func (pcb *PeerCircuitBreakers) Get(peerID string) *CircuitBreaker {
	pcb.mu.RLock()
	cb, exists := pcb.breakers[peerID]
	pcb.mu.RUnlock()

	if exists {
		return cb
	}

	pcb.mu.Lock()
	defer pcb.mu.Unlock()

	// Double-check after acquiring write lock
	if cb, exists = pcb.breakers[peerID]; exists {
		return cb
	}

	cb = NewCircuitBreaker(pcb.config)
	pcb.breakers[peerID] = cb
	return cb
}

// AllowRequest checks if a request to a peer should be allowed
func (pcb *PeerCircuitBreakers) AllowRequest(peerID string) bool {
	return pcb.Get(peerID).Allow()
}

// RecordSuccess records a successful request to a peer
func (pcb *PeerCircuitBreakers) RecordSuccess(peerID string) {
	pcb.Get(peerID).RecordSuccess()
}

// RecordFailure records a failed request to a peer
func (pcb *PeerCircuitBreakers) RecordFailure(peerID string) {
	pcb.Get(peerID).RecordFailure()
}

// GetOpenCircuits returns a list of peer IDs with open circuits
func (pcb *PeerCircuitBreakers) GetOpenCircuits() []string {
	pcb.mu.RLock()
	defer pcb.mu.RUnlock()

	var open []string
	for peerID, cb := range pcb.breakers {
		if cb.State() == CircuitOpen {
			open = append(open, peerID)
		}
	}
	return open
}

// GetStats returns statistics for all circuit breakers
func (pcb *PeerCircuitBreakers) GetStats() map[string]CircuitBreakerStats {
	pcb.mu.RLock()
	defer pcb.mu.RUnlock()

	stats := make(map[string]CircuitBreakerStats)
	for peerID, cb := range pcb.breakers {
		stats[peerID] = cb.Stats()
	}
	return stats
}

// Remove removes a circuit breaker for a peer
func (pcb *PeerCircuitBreakers) Remove(peerID string) {
	pcb.mu.Lock()
	defer pcb.mu.Unlock()
	delete(pcb.breakers, peerID)
}

// Reset resets all circuit breakers
func (pcb *PeerCircuitBreakers) Reset() {
	pcb.mu.Lock()
	defer pcb.mu.Unlock()
	for _, cb := range pcb.breakers {
		cb.Reset()
	}
}

// Cleanup removes circuit breakers for peers that have been healthy for a while
func (pcb *PeerCircuitBreakers) Cleanup(maxAge time.Duration) {
	pcb.mu.Lock()
	defer pcb.mu.Unlock()

	now := time.Now()
	for peerID, cb := range pcb.breakers {
		stats := cb.Stats()
		// Remove if closed and no recent failures
		if stats.State == CircuitClosed && stats.Failures == 0 {
			if stats.LastFailure.IsZero() || now.Sub(stats.LastFailure) > maxAge {
				delete(pcb.breakers, peerID)
			}
		}
	}
}
