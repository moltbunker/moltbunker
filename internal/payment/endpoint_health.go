package payment

import (
	"sort"
	"sync"
	"time"
)

const (
	defaultMaxConsecutiveErrors = 3
	defaultRecoveryInterval    = 30 * time.Second
	ewmaAlpha                  = 0.3  // Weight for new latency samples
	defaultInitialLatency      = 100 * time.Millisecond // Sentinel latency for unmeasured endpoints
)

// EndpointHealth tracks the health state of a single RPC endpoint.
type EndpointHealth struct {
	URL              string
	Latency          time.Duration // EWMA
	ConsecutiveErrs  int
	LastSuccess      time.Time
	LastError        time.Time
	Healthy          bool
	latencySamples   int // number of samples seen (0 = no EWMA yet)
}

// EndpointTracker manages health state for a set of RPC endpoints.
type EndpointTracker struct {
	mu        sync.RWMutex
	endpoints []*EndpointHealth
	maxErrors int           // consecutive errors before marking unhealthy
	recovery  time.Duration // wait before retrying an unhealthy endpoint
}

// NewEndpointTracker creates a tracker from a list of URLs.
// All endpoints start healthy.
func NewEndpointTracker(urls []string) *EndpointTracker {
	endpoints := make([]*EndpointHealth, len(urls))
	for i, u := range urls {
		endpoints[i] = &EndpointHealth{
			URL:     u,
			Healthy: true,
			Latency: defaultInitialLatency, // Avoid 0-latency sorting first
		}
	}
	return &EndpointTracker{
		endpoints: endpoints,
		maxErrors: defaultMaxConsecutiveErrors,
		recovery:  defaultRecoveryInterval,
	}
}

// RecordSuccess records a successful call to the endpoint.
func (et *EndpointTracker) RecordSuccess(url string, latency time.Duration) {
	et.mu.Lock()
	defer et.mu.Unlock()

	ep := et.find(url)
	if ep == nil {
		return
	}

	ep.ConsecutiveErrs = 0
	ep.LastSuccess = time.Now()
	ep.Healthy = true

	// Update EWMA latency
	if ep.latencySamples == 0 {
		ep.Latency = latency
	} else {
		ep.Latency = time.Duration(
			ewmaAlpha*float64(latency) + (1-ewmaAlpha)*float64(ep.Latency),
		)
	}
	ep.latencySamples++
}

// RecordError records a failed call to the endpoint.
func (et *EndpointTracker) RecordError(url string) {
	et.mu.Lock()
	defer et.mu.Unlock()

	ep := et.find(url)
	if ep == nil {
		return
	}

	ep.ConsecutiveErrs++
	ep.LastError = time.Now()

	if ep.ConsecutiveErrs >= et.maxErrors {
		ep.Healthy = false
	}
}

// GetHealthy returns healthy endpoint URLs sorted by latency (lowest first).
// Endpoints that have been unhealthy longer than the recovery interval are
// included for retry (at the end of the list).
func (et *EndpointTracker) GetHealthy() []string {
	et.mu.RLock()
	defer et.mu.RUnlock()

	now := time.Now()

	type candidate struct {
		url     string
		latency time.Duration
		recover bool // true if this is a recovery probe candidate
	}

	var candidates []candidate
	for _, ep := range et.endpoints {
		if ep.Healthy {
			candidates = append(candidates, candidate{url: ep.url(), latency: ep.Latency})
		} else if !ep.LastError.IsZero() && now.Sub(ep.LastError) >= et.recovery {
			// Eligible for recovery retry — put at the end with max latency
			candidates = append(candidates, candidate{url: ep.url(), latency: time.Hour, recover: true})
		}
	}

	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].recover != candidates[j].recover {
			return !candidates[i].recover // healthy before recovery
		}
		return candidates[i].latency < candidates[j].latency
	})

	urls := make([]string, len(candidates))
	for i, c := range candidates {
		urls[i] = c.url
	}
	return urls
}

// GetNext returns the next healthy endpoint excluding the given URL.
// Returns ("", false) if no alternatives exist.
func (et *EndpointTracker) GetNext(exclude string) (string, bool) {
	healthy := et.GetHealthy()
	for _, u := range healthy {
		if u != exclude {
			return u, true
		}
	}
	return "", false
}

// Len returns the total number of tracked endpoints.
func (et *EndpointTracker) Len() int {
	et.mu.RLock()
	defer et.mu.RUnlock()
	return len(et.endpoints)
}

// find returns the endpoint with the given URL (must hold lock).
func (et *EndpointTracker) find(url string) *EndpointHealth {
	for _, ep := range et.endpoints {
		if ep.URL == url {
			return ep
		}
	}
	return nil
}

// url is a helper to avoid field/method name collision.
func (ep *EndpointHealth) url() string {
	return ep.URL
}
