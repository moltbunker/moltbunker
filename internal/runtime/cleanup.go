package runtime

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/moltbunker/moltbunker/internal/logging"
	"github.com/moltbunker/moltbunker/pkg/types"
)

// CleanupManager handles container cleanup on deployment termination.
// It ensures containers are fully stopped, cgroups removed, mounts cleaned,
// and PID files removed. It tracks cleanup state to prevent double-cleanup.
type CleanupManager struct {
	client    *ContainerdClient
	cgroups   *CgroupManager
	dataDir   string
	cleanedUp map[string]bool // tracks which containers have been cleaned up
	mu        sync.Mutex
}

// NewCleanupManager creates a new CleanupManager.
func NewCleanupManager(client *ContainerdClient, cgroups *CgroupManager, dataDir string) *CleanupManager {
	return &CleanupManager{
		client:    client,
		cgroups:   cgroups,
		dataDir:   dataDir,
		cleanedUp: make(map[string]bool),
	}
}

// CleanupContainer ensures a container is fully stopped and all resources are released.
// It stops the container (if running), removes its cgroup, cleans mount points,
// removes the PID file, and deletes the container from the runtime.
// It is safe to call multiple times; subsequent calls are no-ops.
func (cm *CleanupManager) CleanupContainer(ctx context.Context, containerID string) error {
	cm.mu.Lock()
	if cm.cleanedUp[containerID] {
		cm.mu.Unlock()
		return nil // already cleaned up
	}
	// Mark as cleaned up early to prevent concurrent double-cleanup.
	cm.cleanedUp[containerID] = true
	cm.mu.Unlock()

	var firstErr error
	record := func(step string, err error) {
		if err != nil {
			logging.Warn("cleanup step failed",
				logging.ContainerID(containerID),
				"step", step,
				logging.Err(err))
			if firstErr == nil {
				firstErr = fmt.Errorf("cleanup %s: %w", step, err)
			}
		}
	}

	// 1. Stop the container if it is still running.
	if cm.client != nil {
		managed, exists := cm.client.GetContainer(containerID)
		if exists {
			managed.mu.RLock()
			status := managed.Status
			managed.mu.RUnlock()

			if status == types.ContainerStatusRunning {
				record("stop", cm.client.StopContainer(ctx, containerID, 10*1e9)) // 10s
			}

			// 2. Delete the container (this also kills a lingering task and removes the snapshot).
			record("delete", cm.client.DeleteContainer(ctx, containerID))
		}
	}

	// 3. Remove the cgroup for this container.
	if cm.cgroups != nil {
		record("cgroup", cm.cgroups.DeleteCgroup(containerID))
	}

	// 4. Clean mount points.
	mountDir := filepath.Join(cm.dataDir, "mounts", containerID)
	if _, err := os.Stat(mountDir); err == nil {
		record("mount", os.RemoveAll(mountDir))
	}

	// 5. Remove PID file.
	pidFile := filepath.Join(cm.dataDir, "pids", containerID+".pid")
	if _, err := os.Stat(pidFile); err == nil {
		record("pid", os.Remove(pidFile))
	}

	if firstErr == nil {
		logging.Info("container cleaned up", logging.ContainerID(containerID))
	}

	return firstErr
}

// CleanupOrphaned finds and cleans up containers that exist in the containerd
// runtime but are not tracked by the ContainerdClient's in-memory state.
// This handles containers left behind by a previous crash or unclean shutdown.
func (cm *CleanupManager) CleanupOrphaned(ctx context.Context) ([]string, error) {
	// Load containers that containerd knows about.
	nsCtx := cm.client.WithNamespace(ctx)
	runtimeContainers, err := cm.client.Client().Containers(nsCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to list runtime containers: %w", err)
	}

	// Build a set of tracked container IDs.
	tracked := make(map[string]bool)
	for _, info := range cm.client.ListContainers() {
		tracked[info.ID] = true
	}

	var cleaned []string
	for _, rc := range runtimeContainers {
		id := rc.ID()
		if tracked[id] {
			continue
		}
		// Orphaned container: not in our tracked state.
		logging.Warn("found orphaned container", logging.ContainerID(id))

		// Try to kill any running task first.
		if task, taskErr := rc.Task(nsCtx, nil); taskErr == nil && task != nil {
			task.Kill(nsCtx, 9) // SIGKILL
			task.Delete(nsCtx)
		}

		// Delete the container and its snapshot.
		if delErr := rc.Delete(nsCtx); delErr != nil {
			logging.Warn("failed to delete orphaned container",
				logging.ContainerID(id),
				logging.Err(delErr))
			continue
		}

		// Remove cgroup if present.
		if cm.cgroups != nil {
			cm.cgroups.DeleteCgroup(id)
		}

		cleaned = append(cleaned, id)
		logging.Info("cleaned orphaned container", logging.ContainerID(id))
	}

	return cleaned, nil
}

// CleanupAll cleans up all tracked containers. This is intended for graceful
// shutdown scenarios where the node needs to release every resource.
func (cm *CleanupManager) CleanupAll(ctx context.Context) error {
	containers := cm.client.ListContainers()
	var firstErr error
	for _, c := range containers {
		if err := cm.CleanupContainer(ctx, c.ID); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

// IsCleanedUp returns true if the given container has already been cleaned up.
func (cm *CleanupManager) IsCleanedUp(containerID string) bool {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	return cm.cleanedUp[containerID]
}

// ResetCleanupState removes the cleaned-up marker for a container, allowing
// it to be cleaned up again. This is useful after a container is re-created
// with the same ID.
func (cm *CleanupManager) ResetCleanupState(containerID string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	delete(cm.cleanedUp, containerID)
}
