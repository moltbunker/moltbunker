package redundancy

import (
	"context"
	"sync"
	"time"

	"github.com/moltbunker/moltbunker/pkg/types"
)

// HealthMonitor monitors health of container replicas
type HealthMonitor struct {
	replicas    map[string]map[int]*ReplicaHealth
	mu          sync.RWMutex
	interval    time.Duration
	timeout     time.Duration
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
	}
}

// Start starts health monitoring
func (hm *HealthMonitor) Start(ctx context.Context) {
	ticker := time.NewTicker(hm.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			hm.checkHealth(ctx)
		}
	}
}

// checkHealth checks health of all replicas
func (hm *HealthMonitor) checkHealth(ctx context.Context) {
	hm.mu.RLock()
	replicasCopy := make(map[string]map[int]*ReplicaHealth)
	for containerID, replicas := range hm.replicas {
		replicasCopy[containerID] = make(map[int]*ReplicaHealth)
		for idx, health := range replicas {
			replicasCopy[containerID][idx] = health
		}
	}
	hm.mu.RUnlock()

	for containerID, replicas := range replicasCopy {
		for idx, health := range replicas {
			if time.Since(health.LastHeartbeat) > hm.timeout {
				// Replica is unhealthy
				hm.mu.Lock()
				hm.replicas[containerID][idx].Healthy = false
				hm.mu.Unlock()
			}
		}
	}
}

// UpdateHealth updates health status for a replica
func (hm *HealthMonitor) UpdateHealth(containerID string, replicaIndex int, health types.HealthStatus) {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	if hm.replicas[containerID] == nil {
		hm.replicas[containerID] = make(map[int]*ReplicaHealth)
	}

	hm.replicas[containerID][replicaIndex] = &ReplicaHealth{
		ContainerID:   containerID,
		ReplicaIndex:  replicaIndex,
		LastHeartbeat: time.Now(),
		Health:        health,
		Healthy:      true,
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
