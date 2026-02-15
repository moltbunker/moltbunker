package daemon

import (
	"context"
	"time"

	"github.com/moltbunker/moltbunker/internal/logging"
	"github.com/moltbunker/moltbunker/internal/runtime"
	"github.com/moltbunker/moltbunker/internal/util"
	"github.com/moltbunker/moltbunker/pkg/types"
)

// StartDiskEnforcer launches a goroutine that periodically checks container disk usage
// against their configured DiskLimit. Containers exceeding the limit are paused/stopped.
// This is a soft enforcement mechanism — containers may briefly exceed the limit between checks.
func (cm *ContainerManager) StartDiskEnforcer(ctx context.Context, interval time.Duration) {
	if interval == 0 {
		interval = 60 * time.Second
	}

	util.SafeGoWithName("disk-enforcer", func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				cm.checkDiskUsage(ctx)
			}
		}
	})

	logging.Info("disk enforcer started",
		"interval", interval.String(),
		logging.Component("disk_enforcer"))
}

func (cm *ContainerManager) checkDiskUsage(ctx context.Context) {
	// Need the concrete ContainerdClient to call GetContainerDiskUsage
	cc, ok := cm.containerd.(*runtime.ContainerdClient)
	if !ok || cc == nil {
		return
	}

	cm.mu.RLock()
	deployments := make([]*Deployment, 0, len(cm.deployments))
	for _, d := range cm.deployments {
		deployments = append(deployments, d)
	}
	cm.mu.RUnlock()

	for _, d := range deployments {
		if d.Status != types.ContainerStatusRunning {
			continue
		}
		if d.Resources.DiskLimit <= 0 {
			continue
		}

		usage, err := cc.GetContainerDiskUsage(ctx, d.ID)
		if err != nil {
			// Snapshot may not exist yet or container may be transitioning
			continue
		}

		pct := float64(usage) / float64(d.Resources.DiskLimit) * 100

		if usage > d.Resources.DiskLimit {
			logging.Warn("container exceeds disk limit — stopping",
				"container_id", d.ID,
				"usage_mb", usage/(1024*1024),
				"limit_mb", d.Resources.DiskLimit/(1024*1024),
				"pct", int(pct),
				logging.Component("disk_enforcer"))

			logging.Audit(logging.AuditEvent{
				Operation: "disk_limit_exceeded",
				Actor:     "system",
				Target:    d.ID,
				Result:    "stopped",
				Details:   "usage exceeded configured disk_limit",
			})

			if err := cm.containerd.StopContainer(ctx, d.ID, 30*time.Second); err != nil {
				logging.Error("failed to stop container exceeding disk limit",
					"container_id", d.ID,
					logging.Err(err),
					logging.Component("disk_enforcer"))
			}
		} else if pct > 90 {
			logging.Warn("container approaching disk limit (>90%)",
				"container_id", d.ID,
				"usage_mb", usage/(1024*1024),
				"limit_mb", d.Resources.DiskLimit/(1024*1024),
				"pct", int(pct),
				logging.Component("disk_enforcer"))
		}
	}
}
