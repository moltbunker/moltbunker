package runtime

import (
	"context"
	"fmt"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/oci"
	"github.com/moltbunker/moltbunker/pkg/types"
)

// SandboxManager manages container sandboxes
type SandboxManager struct {
	client *ContainerdClient
}

// NewSandboxManager creates a new sandbox manager
func NewSandboxManager(client *ContainerdClient) *SandboxManager {
	return &SandboxManager{
		client: client,
	}
}

// CreateSandbox creates a new container sandbox with resource limits
func (sm *SandboxManager) CreateSandbox(ctx context.Context, namespace string, containerID string, image containerd.Image, resources types.ResourceLimits) (containerd.Container, error) {
	ctx = sm.client.WithNamespace(ctx, namespace)

	// Create container with resource limits
	container, err := sm.client.Client().NewContainer(
		ctx,
		containerID,
		containerd.WithImage(image),
		containerd.WithNewSnapshot(containerID+"-snapshot", image),
		containerd.WithNewSpec(
			oci.WithImageConfig(image),
			oci.WithCPUCFS(resources.CPUQuota, resources.CPUPeriod),
			oci.WithMemoryLimit(uint64(resources.MemoryLimit)),
			oci.WithPidsLimit(int64(resources.PIDLimit)),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	return container, nil
}

// StartSandbox starts a container sandbox
func (sm *SandboxManager) StartSandbox(ctx context.Context, namespace string, container containerd.Container) (containerd.Task, error) {
	ctx = sm.client.WithNamespace(ctx, namespace)

	task, err := container.NewTask(ctx, cio.NewCreator(cio.WithStdio))
	if err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	if err := task.Start(ctx); err != nil {
		return nil, fmt.Errorf("failed to start task: %w", err)
	}

	return task, nil
}

// StopSandbox stops a container sandbox
func (sm *SandboxManager) StopSandbox(ctx context.Context, namespace string, task containerd.Task) error {
	ctx = sm.client.WithNamespace(ctx, namespace)

	if _, err := task.Delete(ctx); err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	return nil
}

// DeleteSandbox deletes a container sandbox
func (sm *SandboxManager) DeleteSandbox(ctx context.Context, namespace string, container containerd.Container) error {
	ctx = sm.client.WithNamespace(ctx, namespace)

	if err := container.Delete(ctx); err != nil {
		return fmt.Errorf("failed to delete container: %w", err)
	}

	return nil
}

// GetSandbox retrieves a container sandbox
func (sm *SandboxManager) GetSandbox(ctx context.Context, namespace, containerID string) (containerd.Container, error) {
	ctx = sm.client.WithNamespace(ctx, namespace)

	container, err := sm.client.Client().LoadContainer(ctx, containerID)
	if err != nil {
		return nil, fmt.Errorf("failed to load container: %w", err)
	}

	return container, nil
}
