package crawl

import (
	"sync"
	"time"
)

// DomainRateLimiter enforces per-domain request rate limits.
type DomainRateLimiter struct {
	mu       sync.Mutex
	domains  map[string]time.Time // domain → last request time
	interval time.Duration        // minimum interval between requests
}

// NewDomainRateLimiter creates a new per-domain rate limiter.
func NewDomainRateLimiter(interval time.Duration) *DomainRateLimiter {
	return &DomainRateLimiter{
		domains:  make(map[string]time.Time),
		interval: interval,
	}
}

// Allow returns true if a request to this domain is allowed now.
// If allowed, records the current time as the last request time.
func (rl *DomainRateLimiter) Allow(domain string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	last, ok := rl.domains[domain]
	if ok && now.Sub(last) < rl.interval {
		return false
	}

	rl.domains[domain] = now
	return true
}

// SetDelay overrides the rate limit for a specific domain.
// Used to respect robots.txt crawl-delay directives.
func (rl *DomainRateLimiter) SetDelay(domain string, delay time.Duration) {
	// The delay is applied via the interval check.
	// For per-domain custom delays, we'd need a more complex structure.
	// For now, the global interval applies to all domains uniformly.
	_ = domain
	_ = delay
}

// URLDedup tracks which URLs have been seen per job to avoid re-crawling.
type URLDedup struct {
	mu   sync.RWMutex
	seen map[string]map[string]bool // jobID → set of URLs
}

// NewURLDedup creates a new URL deduplicator.
func NewURLDedup() *URLDedup {
	return &URLDedup{
		seen: make(map[string]map[string]bool),
	}
}

// Seen returns true if this URL has already been crawled for this job.
func (d *URLDedup) Seen(jobID, url string) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()

	urls, ok := d.seen[jobID]
	if !ok {
		return false
	}
	return urls[url]
}

// Mark records a URL as seen for a job.
func (d *URLDedup) Mark(jobID, url string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	urls, ok := d.seen[jobID]
	if !ok {
		urls = make(map[string]bool)
		d.seen[jobID] = urls
	}
	urls[url] = true
}

// Clear removes all tracked URLs for a job.
func (d *URLDedup) Clear(jobID string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	delete(d.seen, jobID)
}

// Count returns the number of unique URLs seen for a job.
func (d *URLDedup) Count(jobID string) int {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return len(d.seen[jobID])
}
