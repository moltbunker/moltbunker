package redundancy

import (
	"context"
	"sync"
	"time"

	"github.com/moltbunker/moltbunker/pkg/types"
)

const (
	// maxTrackedContainers is the maximum number of containers tracked by the health monitor
	maxTrackedContainers = 10000
)

// HealthProbeFunc is a function that performs actual health probe on a container
type HealthProbeFunc func(ctx context.Context, containerID string) (bool, error)

// HealthMonitor monitors health of container replicas
type HealthMonitor struct {
	replicas    map[string]map[int]*ReplicaHealth
	mu          sync.RWMutex
	interval    time.Duration
	timeout     time.Duration
	probeFunc   HealthProbeFunc // Optional probe function for local containers
	running     bool
	stopCh      chan struct{}
}

// ReplicaHealth tracks health of a single replica
type ReplicaHealth struct {
	ContainerID  string
	ReplicaIndex int
	LastHeartbeat time.Time
	Health       types.HealthStatus
	Healthy      bool
}

// NewHealthMonitor creates a new health monitor
func NewHealthMonitor() *HealthMonitor {
	return &HealthMonitor{
		replicas: make(map[string]map[int]*ReplicaHealth),
		interval: 10 * time.Second,
		timeout:  30 * time.Second,
		stopCh:   make(chan struct{}),
	}
}

// SetProbeFunc sets the health probe function for local container checks
func (hm *HealthMonitor) SetProbeFunc(fn HealthProbeFunc) {
	hm.probeFunc = fn
}

// SetInterval sets the health check interval
func (hm *HealthMonitor) SetInterval(interval time.Duration) {
	hm.interval = interval
}

// Start starts health monitoring
func (hm *HealthMonitor) Start(ctx context.Context) {
	hm.mu.Lock()
	if hm.running {
		hm.mu.Unlock()
		return
	}
	hm.running = true
	hm.mu.Unlock()

	ticker := time.NewTicker(hm.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			hm.mu.Lock()
			hm.running = false
			hm.mu.Unlock()
			return
		case <-hm.stopCh:
			hm.mu.Lock()
			hm.running = false
			hm.mu.Unlock()
			return
		case <-ticker.C:
			hm.checkHealth(ctx)
		}
	}
}

// Stop stops the health monitor
func (hm *HealthMonitor) Stop() {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	if !hm.running {
		return
	}

	hm.running = false

	// Signal stop
	select {
	case <-hm.stopCh:
		// Already closed
	default:
		close(hm.stopCh)
	}
}

// Reset resets the health monitor for reuse after Stop
func (hm *HealthMonitor) Reset() {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	if hm.running {
		return
	}

	hm.stopCh = make(chan struct{})
}

// checkHealth checks health of all replicas
func (hm *HealthMonitor) checkHealth(ctx context.Context) {
	// Snapshot replica data under lock
	type replicaSnapshot struct {
		containerID   string
		idx           int
		healthy       bool
		lastHeartbeat time.Time
	}
	hm.mu.RLock()
	var snapshots []replicaSnapshot
	for containerID, replicas := range hm.replicas {
		for idx, health := range replicas {
			snapshots = append(snapshots, replicaSnapshot{
				containerID:   containerID,
				idx:           idx,
				healthy:       health.Healthy,
				lastHeartbeat: health.LastHeartbeat,
			})
		}
	}
	hm.mu.RUnlock()

	for _, snap := range snapshots {
		healthy := snap.healthy

		// Use probe function if available for local container (replica 0)
		if hm.probeFunc != nil && snap.idx == 0 {
			probeHealthy, err := hm.probeFunc(ctx, snap.containerID)
			if err != nil {
				// Probe failed - mark unhealthy
				healthy = false
			} else {
				healthy = probeHealthy
			}
		} else {
			// For remote replicas, use heartbeat timeout
			if time.Since(snap.lastHeartbeat) > hm.timeout {
				healthy = false
			}
		}

		// Update health status if changed
		if healthy != snap.healthy {
			hm.mu.Lock()
			if hm.replicas[snap.containerID] != nil && hm.replicas[snap.containerID][snap.idx] != nil {
				hm.replicas[snap.containerID][snap.idx].Healthy = healthy
				hm.replicas[snap.containerID][snap.idx].LastHeartbeat = time.Now()
			}
			hm.mu.Unlock()
		}
	}
}

// UpdateHealth updates health status for a replica
func (hm *HealthMonitor) UpdateHealth(containerID string, replicaIndex int, health types.HealthStatus) {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	// If this is a new container and we are at capacity, clean stale entries first
	if hm.replicas[containerID] == nil {
		if len(hm.replicas) >= maxTrackedContainers {
			hm.cleanStaleLocked(hm.timeout)
		}
		// If still at capacity after cleaning, reject the update
		if len(hm.replicas) >= maxTrackedContainers {
			return
		}
		hm.replicas[containerID] = make(map[int]*ReplicaHealth)
	}

	// Use LastUpdate from health if provided, otherwise use current time
	lastHeartbeat := health.LastUpdate
	if lastHeartbeat.IsZero() {
		lastHeartbeat = time.Now()
	}

	hm.replicas[containerID][replicaIndex] = &ReplicaHealth{
		ContainerID:   containerID,
		ReplicaIndex:  replicaIndex,
		LastHeartbeat: lastHeartbeat,
		Health:        health,
		Healthy:       health.Healthy,
	}
}

// GetHealth returns health status for a replica
func (hm *HealthMonitor) GetHealth(containerID string, replicaIndex int) (*ReplicaHealth, bool) {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	if replicas, exists := hm.replicas[containerID]; exists {
		if health, exists := replicas[replicaIndex]; exists {
			return health, true
		}
	}

	return nil, false
}

// IsHealthy checks if a replica is healthy
func (hm *HealthMonitor) IsHealthy(containerID string, replicaIndex int) bool {
	health, exists := hm.GetHealth(containerID, replicaIndex)
	if !exists {
		return false
	}

	return health.Healthy && time.Since(health.LastHeartbeat) < hm.timeout
}

// GetUnhealthyReplicas returns list of unhealthy replicas
func (hm *HealthMonitor) GetUnhealthyReplicas(containerID string) []int {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	var unhealthy []int
	if replicas, exists := hm.replicas[containerID]; exists {
		for idx, health := range replicas {
			if !health.Healthy || time.Since(health.LastHeartbeat) > hm.timeout {
				unhealthy = append(unhealthy, idx)
			}
		}
	}

	return unhealthy
}

// RemoveContainer removes all tracked replicas for a container
func (hm *HealthMonitor) RemoveContainer(containerID string) {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	delete(hm.replicas, containerID)
}

// CleanStale removes all containers where every replica's LastHeartbeat is older than maxAge
func (hm *HealthMonitor) CleanStale(maxAge time.Duration) {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	hm.cleanStaleLocked(maxAge)
}

// cleanStaleLocked removes stale containers. Must be called with mu held for writing.
func (hm *HealthMonitor) cleanStaleLocked(maxAge time.Duration) {
	cutoff := time.Now().Add(-maxAge)
	for containerID, replicas := range hm.replicas {
		allStale := true
		for _, health := range replicas {
			if health.LastHeartbeat.After(cutoff) {
				allStale = false
				break
			}
		}
		if allStale {
			delete(hm.replicas, containerID)
		}
	}
}
