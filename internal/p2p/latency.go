package p2p

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
)

// LatencyMonitor tracks cross-region latency measurements for peers.
type LatencyMonitor struct {
	measurements map[peer.ID]*LatencyMeasurement
	mu           sync.RWMutex
}

// LatencyMeasurement stores latency statistics for a single peer.
type LatencyMeasurement struct {
	PeerID      peer.ID       `json:"peer_id"`
	Region      string        `json:"region"`
	LastPing    time.Duration `json:"last_ping_ns"`
	AvgPing     time.Duration `json:"avg_ping_ns"`
	MinPing     time.Duration `json:"min_ping_ns"`
	MaxPing     time.Duration `json:"max_ping_ns"`
	SampleCount int           `json:"sample_count"`
	LastUpdated time.Time     `json:"last_updated"`
}

// NewLatencyMonitor creates a new LatencyMonitor.
func NewLatencyMonitor() *LatencyMonitor {
	return &LatencyMonitor{
		measurements: make(map[peer.ID]*LatencyMeasurement),
	}
}

// RecordLatency records a latency measurement for a peer, updating the running
// average and min/max statistics.
func (lm *LatencyMonitor) RecordLatency(peerID peer.ID, region string, latency time.Duration) {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	m, exists := lm.measurements[peerID]
	if !exists {
		lm.measurements[peerID] = &LatencyMeasurement{
			PeerID:      peerID,
			Region:      region,
			LastPing:    latency,
			AvgPing:     latency,
			MinPing:     latency,
			MaxPing:     latency,
			SampleCount: 1,
			LastUpdated: time.Now(),
		}
		return
	}

	m.LastPing = latency
	m.SampleCount++
	m.Region = region
	m.LastUpdated = time.Now()

	// Update running average: newAvg = oldAvg + (sample - oldAvg) / count
	m.AvgPing = m.AvgPing + (latency-m.AvgPing)/time.Duration(m.SampleCount)

	if latency < m.MinPing {
		m.MinPing = latency
	}
	if latency > m.MaxPing {
		m.MaxPing = latency
	}
}

// GetLatency returns the latency measurement for a specific peer.
func (lm *LatencyMonitor) GetLatency(peerID peer.ID) (*LatencyMeasurement, bool) {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	m, exists := lm.measurements[peerID]
	if !exists {
		return nil, false
	}

	// Return a copy to avoid data races on the caller side
	cp := *m
	return &cp, true
}

// GetRegionLatencies returns the average latency for each region, computed
// across all peers in that region.
func (lm *LatencyMonitor) GetRegionLatencies() map[string]time.Duration {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	type regionAccum struct {
		total time.Duration
		count int
	}

	accum := make(map[string]*regionAccum)

	for _, m := range lm.measurements {
		ra, exists := accum[m.Region]
		if !exists {
			ra = &regionAccum{}
			accum[m.Region] = ra
		}
		ra.total += m.AvgPing
		ra.count++
	}

	result := make(map[string]time.Duration, len(accum))
	for region, ra := range accum {
		result[region] = ra.total / time.Duration(ra.count)
	}
	return result
}

// GetBestPeersByRegion returns up to n peer IDs in the given region, sorted by
// lowest average latency first.
func (lm *LatencyMonitor) GetBestPeersByRegion(region string, n int) []peer.ID {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	// Collect peers in the requested region
	var peers []*LatencyMeasurement
	for _, m := range lm.measurements {
		if m.Region == region {
			peers = append(peers, m)
		}
	}

	// Sort by average latency ascending
	sort.Slice(peers, func(i, j int) bool {
		return peers[i].AvgPing < peers[j].AvgPing
	})

	if n > len(peers) {
		n = len(peers)
	}

	result := make([]peer.ID, n)
	for i := 0; i < n; i++ {
		result[i] = peers[i].PeerID
	}
	return result
}

// StartCleanup runs a background loop that periodically removes stale
// measurements. It blocks until ctx is cancelled.
func (lm *LatencyMonitor) StartCleanup(ctx context.Context, interval, maxAge time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			lm.CleanStale(maxAge)
		}
	}
}

// CleanStale removes measurements that have not been updated within maxAge.
func (lm *LatencyMonitor) CleanStale(maxAge time.Duration) {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	now := time.Now()
	for id, m := range lm.measurements {
		if now.Sub(m.LastUpdated) > maxAge {
			delete(lm.measurements, id)
		}
	}
}
