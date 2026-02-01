package redundancy

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/moltbunker/moltbunker/internal/logging"
	"github.com/moltbunker/moltbunker/pkg/types"
)

// FailoverManager handles automatic failover for failed replicas
type FailoverManager struct {
	replicator    *Replicator
	healthMonitor *HealthMonitor
	onFailover    func(ctx context.Context, containerID string, replicaIndex int, region string) (*types.Container, error)
	checkInterval time.Duration
	running       bool
	mu            sync.Mutex
	stopCh        chan struct{}
}

// NewFailoverManager creates a new failover manager
func NewFailoverManager(replicator *Replicator, healthMonitor *HealthMonitor) *FailoverManager {
	return &FailoverManager{
		replicator:    replicator,
		healthMonitor: healthMonitor,
		checkInterval: 30 * time.Second, // Check every 30 seconds
		stopCh:        make(chan struct{}),
	}
}

// SetCheckInterval sets the failover check interval
func (fm *FailoverManager) SetCheckInterval(interval time.Duration) {
	fm.checkInterval = interval
}

// SetFailoverCallback sets the callback for creating new replicas
func (fm *FailoverManager) SetFailoverCallback(callback func(ctx context.Context, containerID string, replicaIndex int, region string) (*types.Container, error)) {
	fm.onFailover = callback
}

// CheckAndFailover checks for failed replicas and triggers failover
func (fm *FailoverManager) CheckAndFailover(ctx context.Context, containerID string) error {
	// Get unhealthy replicas
	unhealthy := fm.healthMonitor.GetUnhealthyReplicas(containerID)

	if len(unhealthy) == 0 {
		return nil // All replicas healthy
	}

	// Get replica set
	replicaSet, exists := fm.replicator.GetReplicaSet(containerID)
	if !exists {
		return fmt.Errorf("replica set not found: %s", containerID)
	}

	// Replace unhealthy replicas
	for _, replicaIndex := range unhealthy {
		// Determine region for replacement
		var region string
		switch replicaIndex {
		case 0:
			region = replicaSet.Region1
		case 1:
			region = replicaSet.Region2
		case 2:
			region = replicaSet.Region3
		default:
			return fmt.Errorf("invalid replica index: %d", replicaIndex)
		}

		// Trigger failover
		if fm.onFailover != nil {
			newContainer, err := fm.onFailover(ctx, containerID, replicaIndex, region)
			if err != nil {
				return fmt.Errorf("failed to create replacement replica: %w", err)
			}

			// Add new replica
			if err := fm.replicator.AddReplica(containerID, replicaIndex, newContainer); err != nil {
				return fmt.Errorf("failed to add replacement replica: %w", err)
			}
		}
	}

	return nil
}

// Start starts automatic failover monitoring
func (fm *FailoverManager) Start(ctx context.Context) {
	fm.mu.Lock()
	if fm.running {
		fm.mu.Unlock()
		return
	}
	fm.running = true
	fm.mu.Unlock()

	ticker := time.NewTicker(fm.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-fm.stopCh:
			return
		case <-ticker.C:
			fm.checkAllContainers(ctx)
		}
	}
}

// checkAllContainers checks all containers for failover
func (fm *FailoverManager) checkAllContainers(ctx context.Context) {
	// Get all replica sets
	replicaSets := fm.replicator.GetAllReplicaSets()

	for containerID := range replicaSets {
		if err := fm.CheckAndFailover(ctx, containerID); err != nil {
			logging.Warn("failover check failed",
				logging.ContainerID(containerID),
				logging.Err(err))
		}
	}
}

// Stop stops the failover manager
func (fm *FailoverManager) Stop() {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	if !fm.running {
		return
	}
	fm.running = false

	// Close stopCh to signal goroutines to stop
	select {
	case <-fm.stopCh:
		// Already closed
	default:
		close(fm.stopCh)
	}
}

// Reset resets the failover manager for reuse after Stop
func (fm *FailoverManager) Reset() {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	if fm.running {
		return // Can't reset while running
	}

	// Recreate stopCh for next Start
	fm.stopCh = make(chan struct{})
}
