package metrics

import (
	"encoding/json"
	"sync"
	"sync/atomic"
	"time"
)

// Collector collects and aggregates metrics for the daemon
type Collector struct {
	// Request counts by method
	requestCounts map[string]*uint64
	requestCountsMu sync.RWMutex

	// Request latencies by method (stored as nanoseconds)
	latencies    map[string]*LatencyHistogram
	latenciesMu  sync.RWMutex

	// Active connections gauge
	activeConnections int64

	// Container count gauge
	containerCount int64

	// Peer count gauge
	peerCount int64

	// Start time for uptime calculation
	startTime time.Time
}

// LatencyHistogram tracks request latencies in buckets
type LatencyHistogram struct {
	// Bucket boundaries in milliseconds
	// Buckets: [0-1ms], [1-5ms], [5-10ms], [10-25ms], [25-50ms], [50-100ms], [100-250ms], [250-500ms], [500-1000ms], [1000ms+]
	buckets [10]uint64
	sum     uint64 // Total latency in nanoseconds
	count   uint64 // Total count
	mu      sync.Mutex
}

// bucket boundaries in milliseconds
var bucketBoundaries = []int64{1, 5, 10, 25, 50, 100, 250, 500, 1000}

// NewCollector creates a new metrics collector
func NewCollector() *Collector {
	return &Collector{
		requestCounts: make(map[string]*uint64),
		latencies:     make(map[string]*LatencyHistogram),
		startTime:     time.Now(),
	}
}

// RecordRequest records a request for the given method
func (c *Collector) RecordRequest(method string) {
	c.requestCountsMu.Lock()
	counter, exists := c.requestCounts[method]
	if !exists {
		var val uint64
		counter = &val
		c.requestCounts[method] = counter
	}
	c.requestCountsMu.Unlock()

	atomic.AddUint64(counter, 1)
}

// RecordLatency records the latency for a request
func (c *Collector) RecordLatency(method string, duration time.Duration) {
	c.latenciesMu.Lock()
	hist, exists := c.latencies[method]
	if !exists {
		hist = &LatencyHistogram{}
		c.latencies[method] = hist
	}
	c.latenciesMu.Unlock()

	hist.Record(duration)
}

// Record records a latency value in the histogram
func (h *LatencyHistogram) Record(d time.Duration) {
	h.mu.Lock()
	defer h.mu.Unlock()

	ms := d.Milliseconds()

	// Find the appropriate bucket
	bucketIdx := len(bucketBoundaries) // Default to last bucket (overflow)
	for i, boundary := range bucketBoundaries {
		if ms < boundary {
			bucketIdx = i
			break
		}
	}

	h.buckets[bucketIdx]++
	h.sum += uint64(d.Nanoseconds())
	h.count++
}

// IncrementConnections increments the active connection count
func (c *Collector) IncrementConnections() {
	atomic.AddInt64(&c.activeConnections, 1)
}

// DecrementConnections decrements the active connection count
func (c *Collector) DecrementConnections() {
	atomic.AddInt64(&c.activeConnections, -1)
}

// SetContainerCount sets the current container count
func (c *Collector) SetContainerCount(count int) {
	atomic.StoreInt64(&c.containerCount, int64(count))
}

// SetPeerCount sets the current peer count
func (c *Collector) SetPeerCount(count int) {
	atomic.StoreInt64(&c.peerCount, int64(count))
}

// Metrics represents the current state of all metrics
type Metrics struct {
	Uptime            string                     `json:"uptime"`
	UptimeSeconds     float64                    `json:"uptime_seconds"`
	RequestCounts     map[string]uint64          `json:"request_counts"`
	RequestLatencies  map[string]LatencyStats    `json:"request_latencies"`
	ActiveConnections int64                      `json:"active_connections"`
	ContainerCount    int64                      `json:"container_count"`
	PeerCount         int64                      `json:"peer_count"`
	CollectedAt       time.Time                  `json:"collected_at"`
}

// LatencyStats contains latency statistics for a method
type LatencyStats struct {
	Count       uint64             `json:"count"`
	SumMs       float64            `json:"sum_ms"`
	AvgMs       float64            `json:"avg_ms"`
	Buckets     map[string]uint64  `json:"buckets"`
}

// GetMetrics returns the current metrics as a Metrics struct
func (c *Collector) GetMetrics() *Metrics {
	uptime := time.Since(c.startTime)

	// Collect request counts
	requestCounts := make(map[string]uint64)
	c.requestCountsMu.RLock()
	for method, counter := range c.requestCounts {
		requestCounts[method] = atomic.LoadUint64(counter)
	}
	c.requestCountsMu.RUnlock()

	// Collect latencies
	latencies := make(map[string]LatencyStats)
	c.latenciesMu.RLock()
	for method, hist := range c.latencies {
		hist.mu.Lock()
		stats := LatencyStats{
			Count:   hist.count,
			SumMs:   float64(hist.sum) / float64(time.Millisecond),
			Buckets: make(map[string]uint64),
		}
		if hist.count > 0 {
			stats.AvgMs = float64(hist.sum) / float64(hist.count) / float64(time.Millisecond)
		}

		// Format bucket labels
		bucketLabels := []string{
			"0-1ms", "1-5ms", "5-10ms", "10-25ms", "25-50ms",
			"50-100ms", "100-250ms", "250-500ms", "500-1000ms", "1000ms+",
		}
		for i, count := range hist.buckets {
			if count > 0 {
				stats.Buckets[bucketLabels[i]] = count
			}
		}
		hist.mu.Unlock()
		latencies[method] = stats
	}
	c.latenciesMu.RUnlock()

	return &Metrics{
		Uptime:            uptime.Round(time.Second).String(),
		UptimeSeconds:     uptime.Seconds(),
		RequestCounts:     requestCounts,
		RequestLatencies:  latencies,
		ActiveConnections: atomic.LoadInt64(&c.activeConnections),
		ContainerCount:    atomic.LoadInt64(&c.containerCount),
		PeerCount:         atomic.LoadInt64(&c.peerCount),
		CollectedAt:       time.Now(),
	}
}

// GetMetricsJSON returns the current metrics as JSON
func (c *Collector) GetMetricsJSON() ([]byte, error) {
	metrics := c.GetMetrics()
	return json.Marshal(metrics)
}

// Reset resets all metrics (useful for testing)
func (c *Collector) Reset() {
	c.requestCountsMu.Lock()
	c.requestCounts = make(map[string]*uint64)
	c.requestCountsMu.Unlock()

	c.latenciesMu.Lock()
	c.latencies = make(map[string]*LatencyHistogram)
	c.latenciesMu.Unlock()

	atomic.StoreInt64(&c.activeConnections, 0)
	atomic.StoreInt64(&c.containerCount, 0)
	atomic.StoreInt64(&c.peerCount, 0)
	c.startTime = time.Now()
}
