package runtime

import (
	"context"
	"fmt"
	"time"

	"github.com/moltbunker/moltbunker/pkg/types"
)

// GetHealthStatus returns health status for a container
func (cc *ContainerdClient) GetHealthStatus(ctx context.Context, id string) (*types.HealthStatus, error) {
	ctx = cc.WithNamespace(ctx)

	cc.mu.RLock()
	managed, exists := cc.containers[id]
	cc.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("container not found: %s", id)
	}

	managed.mu.RLock()
	defer managed.mu.RUnlock()

	healthy := managed.Task != nil && managed.Status == types.ContainerStatusRunning

	// Get metrics if available
	var cpuUsage float64
	var memoryUsage int64

	if managed.Task != nil {
		metrics, err := managed.Task.Metrics(ctx)
		if err == nil && metrics != nil {
			// Parse metrics (simplified)
			_ = metrics
		}
	}

	return &types.HealthStatus{
		CPUUsage:    cpuUsage,
		MemoryUsage: memoryUsage,
		Healthy:     healthy,
		LastUpdate:  time.Now(),
	}, nil
}
