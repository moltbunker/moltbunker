package daemon

import (
	"errors"
	"sync"
	"time"
)

// Default rate limiting constants (can be overridden via config)
const (
	// Per-connection rate limiting defaults: 100 requests per minute
	rateLimitTokens    = 100
	rateLimitRefillSec = 60

	// Global rate limiting defaults: 1000 requests per minute total
	globalRateLimitTokens    = 1000
	globalRateLimitRefillSec = 60
)

// ErrRateLimited is returned when rate limit is exceeded
var ErrRateLimited = errors.New("rate limit exceeded")

// rateLimiter implements a simple token bucket rate limiter
type rateLimiter struct {
	mu           sync.Mutex
	tokens       int
	maxTokens    int
	lastRefill   time.Time
	refillPeriod time.Duration
}

// newRateLimiter creates a new rate limiter
func newRateLimiter(maxTokens int, refillPeriod time.Duration) *rateLimiter {
	return &rateLimiter{
		tokens:       maxTokens,
		maxTokens:    maxTokens,
		lastRefill:   time.Now(),
		refillPeriod: refillPeriod,
	}
}

// allow checks if a request is allowed and consumes a token if so
func (r *rateLimiter) allow() bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Refill tokens based on elapsed time
	now := time.Now()
	elapsed := now.Sub(r.lastRefill)
	if elapsed >= r.refillPeriod {
		// Full refill after the period
		r.tokens = r.maxTokens
		r.lastRefill = now
	}

	if r.tokens > 0 {
		r.tokens--
		return true
	}
	return false
}
