package runtime

import (
	"context"
	"fmt"
	"io"
	"sync"
	"syscall"
	"time"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/containerd/oci"

	"github.com/moltbunker/moltbunker/pkg/types"
)

// ContainerdClient wraps containerd client with full lifecycle management
type ContainerdClient struct {
	client     *containerd.Client
	namespace  string
	containers map[string]*ManagedContainer
	mu         sync.RWMutex
	logManager *LogManager
	logsDir    string
}

// ManagedContainer represents a container managed by the runtime
type ManagedContainer struct {
	ID          string
	Image       string
	Container   containerd.Container
	Task        containerd.Task
	Status      types.ContainerStatus
	CreatedAt   time.Time
	StartedAt   time.Time
	Resources   types.ResourceLimits
	EncryptedVolume string
	OnionAddress    string
	mu          sync.RWMutex
}

// NewContainerdClient creates a new containerd client
func NewContainerdClient(socketPath string, namespace string, logsDir string) (*ContainerdClient, error) {
	client, err := containerd.New(socketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to containerd: %w", err)
	}

	if namespace == "" {
		namespace = "moltbunker"
	}

	if logsDir == "" {
		logsDir = "/var/log/moltbunker/containers"
	}

	// Create log manager
	logManager, err := NewLogManager(logsDir)
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to create log manager: %w", err)
	}

	return &ContainerdClient{
		client:     client,
		namespace:  namespace,
		containers: make(map[string]*ManagedContainer),
		logManager: logManager,
		logsDir:    logsDir,
	}, nil
}

// Client returns the containerd client
func (cc *ContainerdClient) Client() *containerd.Client {
	return cc.client
}

// WithNamespace returns a context with namespace
func (cc *ContainerdClient) WithNamespace(ctx context.Context) context.Context {
	return namespaces.WithNamespace(ctx, cc.namespace)
}

// PullImage pulls a container image
func (cc *ContainerdClient) PullImage(ctx context.Context, ref string) (containerd.Image, error) {
	ctx = cc.WithNamespace(ctx)

	image, err := cc.client.Pull(ctx, ref, containerd.WithPullUnpack)
	if err != nil {
		return nil, fmt.Errorf("failed to pull image %s: %w", ref, err)
	}

	return image, nil
}

// GetImage gets an existing image
func (cc *ContainerdClient) GetImage(ctx context.Context, ref string) (containerd.Image, error) {
	ctx = cc.WithNamespace(ctx)
	return cc.client.GetImage(ctx, ref)
}

// CreateContainer creates a new container
func (cc *ContainerdClient) CreateContainer(ctx context.Context, id string, imageRef string, resources types.ResourceLimits) (*ManagedContainer, error) {
	ctx = cc.WithNamespace(ctx)

	// Get or pull image
	image, err := cc.GetImage(ctx, imageRef)
	if err != nil {
		image, err = cc.PullImage(ctx, imageRef)
		if err != nil {
			return nil, fmt.Errorf("failed to get image: %w", err)
		}
	}

	// Build container spec with resource limits
	// Use the image's default entrypoint/cmd - don't override with sleep
	opts := []oci.SpecOpts{
		oci.WithImageConfig(image),
	}

	// Add resource limits
	if resources.MemoryLimit > 0 {
		opts = append(opts, oci.WithMemoryLimit(uint64(resources.MemoryLimit)))
	}
	if resources.CPUQuota > 0 && resources.CPUPeriod > 0 {
		opts = append(opts, oci.WithCPUCFS(int64(resources.CPUQuota), uint64(resources.CPUPeriod)))
	}
	if resources.PIDLimit > 0 {
		opts = append(opts, oci.WithPidsLimit(int64(resources.PIDLimit)))
	}

	// Create container
	container, err := cc.client.NewContainer(
		ctx,
		id,
		containerd.WithImage(image),
		containerd.WithNewSnapshot(id+"-snapshot", image),
		containerd.WithNewSpec(opts...),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	managed := &ManagedContainer{
		ID:        id,
		Image:     imageRef,
		Container: container,
		Status:    types.ContainerStatusCreated,
		CreatedAt: time.Now(),
		Resources: resources,
	}

	cc.mu.Lock()
	cc.containers[id] = managed
	cc.mu.Unlock()

	return managed, nil
}

// StartContainer starts a container
func (cc *ContainerdClient) StartContainer(ctx context.Context, id string) error {
	ctx = cc.WithNamespace(ctx)

	cc.mu.RLock()
	managed, exists := cc.containers[id]
	cc.mu.RUnlock()

	if !exists {
		return fmt.Errorf("container not found: %s", id)
	}

	managed.mu.Lock()
	defer managed.mu.Unlock()

	// Create log files for the container
	containerLog, err := cc.logManager.CreateLog(id)
	if err != nil {
		return fmt.Errorf("failed to create container logs: %w", err)
	}

	// Create task with log output
	// Use FIFO-based I/O that writes to our log files
	task, err := managed.Container.NewTask(ctx, cio.NewCreator(
		cio.WithStreams(nil, containerLog.StdoutWriter(), containerLog.StderrWriter()),
	))
	if err != nil {
		cc.logManager.CloseLog(id)
		return fmt.Errorf("failed to create task: %w", err)
	}

	// Start task
	if err := task.Start(ctx); err != nil {
		task.Delete(ctx)
		cc.logManager.CloseLog(id)
		return fmt.Errorf("failed to start task: %w", err)
	}

	managed.Task = task
	managed.Status = types.ContainerStatusRunning
	managed.StartedAt = time.Now()

	return nil
}

// StopContainer stops a container
func (cc *ContainerdClient) StopContainer(ctx context.Context, id string, timeout time.Duration) error {
	ctx = cc.WithNamespace(ctx)

	cc.mu.RLock()
	managed, exists := cc.containers[id]
	cc.mu.RUnlock()

	if !exists {
		return fmt.Errorf("container not found: %s", id)
	}

	managed.mu.Lock()
	defer managed.mu.Unlock()

	if managed.Task == nil {
		return nil
	}

	// Send SIGTERM
	if err := managed.Task.Kill(ctx, syscall.SIGTERM); err != nil {
		return fmt.Errorf("failed to send SIGTERM: %w", err)
	}

	// Wait for exit with timeout
	exitCh, err := managed.Task.Wait(ctx)
	if err != nil {
		return fmt.Errorf("failed to wait for task: %w", err)
	}

	select {
	case <-exitCh:
		// Task exited
	case <-time.After(timeout):
		// Force kill
		if err := managed.Task.Kill(ctx, syscall.SIGKILL); err != nil {
			return fmt.Errorf("failed to send SIGKILL: %w", err)
		}
		<-exitCh
	}

	// Delete task
	if _, err := managed.Task.Delete(ctx); err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	// Close log files (but don't delete - keep for history)
	cc.logManager.CloseLog(id)

	managed.Task = nil
	managed.Status = types.ContainerStatusStopped

	return nil
}

// DeleteContainer deletes a container
func (cc *ContainerdClient) DeleteContainer(ctx context.Context, id string) error {
	ctx = cc.WithNamespace(ctx)

	cc.mu.Lock()
	managed, exists := cc.containers[id]
	if !exists {
		cc.mu.Unlock()
		return fmt.Errorf("container not found: %s", id)
	}
	delete(cc.containers, id)
	cc.mu.Unlock()

	managed.mu.Lock()
	defer managed.mu.Unlock()

	// Stop if running
	if managed.Task != nil {
		managed.Task.Kill(ctx, syscall.SIGKILL)
		managed.Task.Delete(ctx)
	}

	// Close and delete logs
	cc.logManager.DeleteLog(id)

	// Delete container
	if err := managed.Container.Delete(ctx, containerd.WithSnapshotCleanup); err != nil {
		return fmt.Errorf("failed to delete container: %w", err)
	}

	return nil
}

// GetContainer returns a managed container
func (cc *ContainerdClient) GetContainer(id string) (*ManagedContainer, bool) {
	cc.mu.RLock()
	defer cc.mu.RUnlock()
	managed, exists := cc.containers[id]
	return managed, exists
}

// ListContainers returns all managed containers (returns copies for thread safety)
func (cc *ContainerdClient) ListContainers() []ManagedContainerInfo {
	cc.mu.RLock()
	defer cc.mu.RUnlock()

	result := make([]ManagedContainerInfo, 0, len(cc.containers))
	for _, c := range cc.containers {
		c.mu.RLock()
		info := ManagedContainerInfo{
			ID:              c.ID,
			Image:           c.Image,
			Status:          c.Status,
			CreatedAt:       c.CreatedAt,
			StartedAt:       c.StartedAt,
			Resources:       c.Resources,
			EncryptedVolume: c.EncryptedVolume,
			OnionAddress:    c.OnionAddress,
		}
		c.mu.RUnlock()
		result = append(result, info)
	}
	return result
}

// ManagedContainerInfo is a snapshot of container info (safe for concurrent access)
type ManagedContainerInfo struct {
	ID              string
	Image           string
	Status          types.ContainerStatus
	CreatedAt       time.Time
	StartedAt       time.Time
	Resources       types.ResourceLimits
	EncryptedVolume string
	OnionAddress    string
}

// GetContainerStatus returns the status of a container
func (cc *ContainerdClient) GetContainerStatus(ctx context.Context, id string) (types.ContainerStatus, error) {
	ctx = cc.WithNamespace(ctx)

	cc.mu.RLock()
	managed, exists := cc.containers[id]
	cc.mu.RUnlock()

	if !exists {
		return types.ContainerStatusFailed, fmt.Errorf("container not found: %s", id)
	}

	managed.mu.RLock()
	defer managed.mu.RUnlock()

	if managed.Task == nil {
		return managed.Status, nil
	}

	// Check task status
	status, err := managed.Task.Status(ctx)
	if err != nil {
		return types.ContainerStatusFailed, err
	}

	switch status.Status {
	case containerd.Running:
		return types.ContainerStatusRunning, nil
	case containerd.Stopped:
		return types.ContainerStatusStopped, nil
	case containerd.Paused:
		return types.ContainerStatusPaused, nil
	default:
		return types.ContainerStatusFailed, nil
	}
}

// GetContainerLogs returns logs from a container
func (cc *ContainerdClient) GetContainerLogs(ctx context.Context, id string, follow bool, tail int) (io.ReadCloser, error) {
	cc.mu.RLock()
	_, exists := cc.containers[id]
	cc.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("container not found: %s", id)
	}

	// Use log manager to read container logs
	return cc.logManager.ReadLogs(ctx, id, follow, tail)
}

// ExecInContainer executes a command in a running container
func (cc *ContainerdClient) ExecInContainer(ctx context.Context, id string, cmd []string) ([]byte, error) {
	ctx = cc.WithNamespace(ctx)

	cc.mu.RLock()
	managed, exists := cc.containers[id]
	cc.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("container not found: %s", id)
	}

	managed.mu.RLock()
	task := managed.Task
	managed.mu.RUnlock()

	if task == nil {
		return nil, fmt.Errorf("container not running: %s", id)
	}

	// Create exec process
	spec, err := managed.Container.Spec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get container spec: %w", err)
	}

	pspec := spec.Process
	pspec.Args = cmd

	execID := fmt.Sprintf("exec-%d", time.Now().UnixNano())

	process, err := task.Exec(ctx, execID, pspec, cio.NewCreator(cio.WithStdio))
	if err != nil {
		return nil, fmt.Errorf("failed to create exec: %w", err)
	}
	defer process.Delete(ctx)

	if err := process.Start(ctx); err != nil {
		return nil, fmt.Errorf("failed to start exec: %w", err)
	}

	exitCh, err := process.Wait(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to wait for exec: %w", err)
	}

	<-exitCh

	return nil, nil
}

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

// SetContainerWithSpec creates a container with custom OCI spec modifications
func (cc *ContainerdClient) CreateContainerWithSpec(ctx context.Context, id string, imageRef string, resources types.ResourceLimits, specOpts ...oci.SpecOpts) (*ManagedContainer, error) {
	ctx = cc.WithNamespace(ctx)

	// Get or pull image
	image, err := cc.GetImage(ctx, imageRef)
	if err != nil {
		image, err = cc.PullImage(ctx, imageRef)
		if err != nil {
			return nil, fmt.Errorf("failed to get image: %w", err)
		}
	}

	// Build container spec with resource limits
	opts := []oci.SpecOpts{
		oci.WithImageConfig(image),
	}

	// Add resource limits
	if resources.MemoryLimit > 0 {
		opts = append(opts, oci.WithMemoryLimit(uint64(resources.MemoryLimit)))
	}
	if resources.CPUQuota > 0 && resources.CPUPeriod > 0 {
		opts = append(opts, oci.WithCPUCFS(int64(resources.CPUQuota), uint64(resources.CPUPeriod)))
	}
	if resources.PIDLimit > 0 {
		opts = append(opts, oci.WithPidsLimit(int64(resources.PIDLimit)))
	}

	// Add custom spec options
	opts = append(opts, specOpts...)

	// Create container
	container, err := cc.client.NewContainer(
		ctx,
		id,
		containerd.WithImage(image),
		containerd.WithNewSnapshot(id+"-snapshot", image),
		containerd.WithNewSpec(opts...),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	managed := &ManagedContainer{
		ID:        id,
		Image:     imageRef,
		Container: container,
		Status:    types.ContainerStatusCreated,
		CreatedAt: time.Now(),
		Resources: resources,
	}

	cc.mu.Lock()
	cc.containers[id] = managed
	cc.mu.Unlock()

	return managed, nil
}

// LoadExistingContainers loads existing containers from containerd
func (cc *ContainerdClient) LoadExistingContainers(ctx context.Context) error {
	ctx = cc.WithNamespace(ctx)

	containers, err := cc.client.Containers(ctx)
	if err != nil {
		return fmt.Errorf("failed to list containers: %w", err)
	}

	cc.mu.Lock()
	defer cc.mu.Unlock()

	for _, container := range containers {
		info, err := container.Info(ctx)
		if err != nil {
			continue
		}

		managed := &ManagedContainer{
			ID:        container.ID(),
			Image:     info.Image,
			Container: container,
			Status:    types.ContainerStatusStopped,
			CreatedAt: info.CreatedAt,
		}

		// Check if task is running
		task, err := container.Task(ctx, nil)
		if err == nil {
			managed.Task = task
			status, err := task.Status(ctx)
			if err == nil && status.Status == containerd.Running {
				managed.Status = types.ContainerStatusRunning
			}
		}

		cc.containers[container.ID()] = managed
	}

	return nil
}

// Close closes the containerd client
func (cc *ContainerdClient) Close() error {
	return cc.client.Close()
}
