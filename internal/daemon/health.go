package daemon

import (
	"context"
	"io"
	"time"

	"github.com/moltbunker/moltbunker/pkg/types"
)

// GetLogs returns logs for a container
func (cm *ContainerManager) GetLogs(ctx context.Context, containerID string, follow bool, tail int) (io.ReadCloser, error) {
	if cm.containerd == nil {
		return nil, ErrContainerdNotAvailable
	}

	return cm.containerd.GetContainerLogs(ctx, containerID, follow, tail)
}

// GetHealth returns health status for a deployment
func (cm *ContainerManager) GetHealth(ctx context.Context, containerID string) (*types.HealthStatus, error) {
	cm.mu.RLock()
	_, exists := cm.deployments[containerID]
	cm.mu.RUnlock()

	if !exists {
		return nil, ErrDeploymentNotFound{ContainerID: containerID}
	}

	// Get health from health monitor
	for i := 0; i < 3; i++ {
		if health, ok := cm.healthMonitor.GetHealth(containerID, i); ok {
			return &health.Health, nil
		}
	}

	// Get from containerd if available
	if cm.containerd != nil {
		return cm.containerd.GetHealthStatus(ctx, containerID)
	}

	return &types.HealthStatus{
		Healthy:    false,
		LastUpdate: time.Now(),
	}, nil
}

// GetUnhealthyDeployments returns deployments with unhealthy replicas
func (cm *ContainerManager) GetUnhealthyDeployments() map[string][]int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	result := make(map[string][]int)
	for containerID := range cm.deployments {
		unhealthy := cm.healthMonitor.GetUnhealthyReplicas(containerID)
		if len(unhealthy) > 0 {
			result[containerID] = unhealthy
		}
	}
	return result
}
