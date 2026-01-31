package redundancy

import (
	"context"
	"fmt"

	"github.com/moltbunker/moltbunker/pkg/types"
)

// FailoverManager handles automatic failover for failed replicas
type FailoverManager struct {
	replicator *Replicator
	healthMonitor *HealthMonitor
	onFailover func(ctx context.Context, containerID string, replicaIndex int, region string) (*types.Container, error)
}

// NewFailoverManager creates a new failover manager
func NewFailoverManager(replicator *Replicator, healthMonitor *HealthMonitor) *FailoverManager {
	return &FailoverManager{
		replicator:    replicator,
		healthMonitor: healthMonitor,
	}
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
	// Failover is triggered by health monitor
	// This would typically be called periodically
}
