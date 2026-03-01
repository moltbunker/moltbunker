package molt

import (
	"sync"
	"sync/atomic"
	"time"
)

// MoltStats is a point-in-time snapshot of invocation statistics.
type MoltStats struct {
	TotalInvocations   uint64        `json:"total_invocations"`
	SuccessInvocations uint64        `json:"success_invocations"`
	ErrorInvocations   uint64        `json:"error_invocations"`
	TimeoutInvocations uint64        `json:"timeout_invocations"`
	ActiveInvocations  int64         `json:"active_invocations"`
	TotalDuration      time.Duration `json:"total_duration"`
}

// deploymentStats tracks per-deployment counters under a mutex.
type deploymentStats struct {
	total    uint64
	success  uint64
	errors   uint64
	timeouts uint64
	duration uint64 // nanoseconds
}

// MoltMetrics collects invocation metrics using atomic counters (global)
// and a mutex-protected map (per-deployment).
type MoltMetrics struct {
	// Global atomic counters
	totalInvocations   uint64
	successInvocations uint64
	errorInvocations   uint64
	timeoutInvocations uint64
	activeInvocations  int64
	totalDuration      uint64 // nanoseconds

	// Per-deployment stats
	deployments map[string]*deploymentStats
	mu          sync.RWMutex
}

// NewMoltMetrics creates a new metrics collector.
func NewMoltMetrics() *MoltMetrics {
	return &MoltMetrics{
		deployments: make(map[string]*deploymentStats),
	}
}

// RecordInvocation records a completed invocation.
func (m *MoltMetrics) RecordInvocation(deploymentID string, duration time.Duration, success bool, timeout bool) {
	ns := uint64(duration.Nanoseconds())

	atomic.AddUint64(&m.totalInvocations, 1)
	atomic.AddUint64(&m.totalDuration, ns)

	if timeout {
		atomic.AddUint64(&m.timeoutInvocations, 1)
	} else if success {
		atomic.AddUint64(&m.successInvocations, 1)
	} else {
		atomic.AddUint64(&m.errorInvocations, 1)
	}

	// Per-deployment
	m.mu.Lock()
	ds, ok := m.deployments[deploymentID]
	if !ok {
		ds = &deploymentStats{}
		m.deployments[deploymentID] = ds
	}
	ds.total++
	ds.duration += ns
	if timeout {
		ds.timeouts++
	} else if success {
		ds.success++
	} else {
		ds.errors++
	}
	m.mu.Unlock()
}

// IncrementActive increments the active invocation gauge.
func (m *MoltMetrics) IncrementActive() {
	atomic.AddInt64(&m.activeInvocations, 1)
}

// DecrementActive decrements the active invocation gauge.
func (m *MoltMetrics) DecrementActive() {
	atomic.AddInt64(&m.activeInvocations, -1)
}

// GetGlobalStats returns a snapshot of global invocation stats.
func (m *MoltMetrics) GetGlobalStats() MoltStats {
	return MoltStats{
		TotalInvocations:   atomic.LoadUint64(&m.totalInvocations),
		SuccessInvocations: atomic.LoadUint64(&m.successInvocations),
		ErrorInvocations:   atomic.LoadUint64(&m.errorInvocations),
		TimeoutInvocations: atomic.LoadUint64(&m.timeoutInvocations),
		ActiveInvocations:  atomic.LoadInt64(&m.activeInvocations),
		TotalDuration:      time.Duration(atomic.LoadUint64(&m.totalDuration)),
	}
}

// GetStats returns a snapshot for a specific deployment. Returns zero stats if not found.
func (m *MoltMetrics) GetStats(deploymentID string) MoltStats {
	m.mu.RLock()
	ds, ok := m.deployments[deploymentID]
	if !ok {
		m.mu.RUnlock()
		return MoltStats{}
	}
	stats := MoltStats{
		TotalInvocations:   ds.total,
		SuccessInvocations: ds.success,
		ErrorInvocations:   ds.errors,
		TimeoutInvocations: ds.timeouts,
		TotalDuration:      time.Duration(ds.duration),
	}
	m.mu.RUnlock()
	return stats
}
