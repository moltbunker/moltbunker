package runtime

import (
	"context"
	"fmt"
	"io"
	goruntime "runtime"
	"strconv"
	"syscall"
	"time"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/oci"
	specs "github.com/opencontainers/runtime-spec/specs-go"

	"github.com/moltbunker/moltbunker/pkg/types"
)

// Note: runtime name is stored in ContainerdClient.runtimeName (set via NewContainerdClient).
// Constants for runtime names are in runtime_detect.go.

// baseSpecOpts returns the foundational OCI spec options for container creation.
// On macOS (Colima), it generates a linux/arm64 default spec so the Linux section
// exists. On Linux, it uses the host default. Image config is always overlaid.
func (cc *ContainerdClient) baseSpecOpts(image containerd.Image) []oci.SpecOpts {
	opts := []oci.SpecOpts{}
	if goruntime.GOOS != "linux" {
		opts = append(opts, oci.WithDefaultSpecForPlatform("linux/arm64"))
	}
	opts = append(opts, oci.WithImageConfig(image))
	return opts
}

// baseContainerOpts returns the foundational container options (image, snapshot, runtime).
// When the runtime is a Kata variant and KataConfig is set, OCI annotations for the
// Kata hypervisor are injected into the spec.
func (cc *ContainerdClient) baseContainerOpts(id string, image containerd.Image, specOpts []oci.SpecOpts) []containerd.NewContainerOpts {
	// Inject Kata hypervisor annotations when running under a Kata runtime
	if isKataRuntime(cc.runtimeName) && cc.kataConfig != nil {
		if ann := cc.kataAnnotations(); ann != nil {
			specOpts = append(specOpts, ann)
		}
	}

	return []containerd.NewContainerOpts{
		containerd.WithImage(image),
		containerd.WithNewSnapshot(id+"-snapshot", image),
		containerd.WithRuntime(cc.runtimeName, nil),
		containerd.WithNewSpec(specOpts...),
	}
}

// kataAnnotations returns an OCI spec option that sets Kata hypervisor annotations.
// Only non-zero KataConfig fields are emitted; Kata's internal defaults apply otherwise.
func (cc *ContainerdClient) kataAnnotations() oci.SpecOpts {
	if cc.kataConfig == nil {
		return nil
	}

	annotations := make(map[string]string)

	if cc.kataConfig.VMMemoryMB > 0 {
		annotations["io.katacontainers.config.hypervisor.default_memory"] = strconv.Itoa(cc.kataConfig.VMMemoryMB)
	}
	if cc.kataConfig.VMCPUs > 0 {
		annotations["io.katacontainers.config.hypervisor.default_vcpus"] = strconv.Itoa(cc.kataConfig.VMCPUs)
	}
	if cc.kataConfig.KernelPath != "" {
		annotations["io.katacontainers.config.hypervisor.kernel"] = cc.kataConfig.KernelPath
	}
	if cc.kataConfig.ImagePath != "" {
		annotations["io.katacontainers.config.hypervisor.image"] = cc.kataConfig.ImagePath
	}

	if len(annotations) == 0 {
		return nil
	}

	return oci.WithAnnotations(annotations)
}

// BindMount describes a host-to-container bind mount.
type BindMount struct {
	HostPath      string // Absolute path on host
	ContainerPath string // Path inside container
	ReadOnly      bool
}

// SecureContainerConfig holds configuration for creating a secure container
type SecureContainerConfig struct {
	ID              string
	ImageRef        string
	Resources       types.ResourceLimits
	SecurityProfile *types.ContainerSecurityProfile
	DeploymentID    string
	RequesterPubKey []byte
	Environment     map[string]string
	Command         []string
	Args            []string
	BindMounts      []BindMount // Host paths bind-mounted into the container
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

	// Build container spec with resource limits.
	// Start with a Linux default spec to ensure the Linux section exists
	// (required when the client runs on macOS with Colima).
	opts := cc.baseSpecOpts(image)

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
	container, err := cc.client.NewContainer(ctx, id, cc.baseContainerOpts(id, image, opts)...)
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

	// Apply XFS disk quota if configured
	if err := cc.SetDiskQuota(ctx, id, resources.DiskLimit); err != nil {
		// Non-fatal: disk_enforcer provides secondary enforcement
		_ = err
	}

	return managed, nil
}

// CreateContainerWithSpec creates a container with custom OCI spec modifications
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

	// Build container spec with Linux defaults + resource limits
	opts := cc.baseSpecOpts(image)

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
	container, err := cc.client.NewContainer(ctx, id, cc.baseContainerOpts(id, image, opts)...)
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

	// Apply XFS disk quota if configured
	if err := cc.SetDiskQuota(ctx, id, resources.DiskLimit); err != nil {
		_ = err
	}

	return managed, nil
}

// CreateSecureContainer creates a container with a full security profile for provider nodes
// This is the primary method for creating containers that enforce container opacity
func (cc *ContainerdClient) CreateSecureContainer(ctx context.Context, config SecureContainerConfig) (*ManagedContainer, error) {
	ctx = cc.WithNamespace(ctx)

	// Get or pull image
	image, err := cc.GetImage(ctx, config.ImageRef)
	if err != nil {
		image, err = cc.PullImage(ctx, config.ImageRef)
		if err != nil {
			return nil, fmt.Errorf("failed to get image: %w", err)
		}
	}

	// Create security enforcer
	securityEnforcer := NewSecurityEnforcer(config.SecurityProfile)

	// Build OCI spec options starting with Linux defaults + image config
	opts := cc.baseSpecOpts(image)

	// Add resource limits
	if config.Resources.MemoryLimit > 0 {
		opts = append(opts, oci.WithMemoryLimit(uint64(config.Resources.MemoryLimit)))
	}
	if config.Resources.CPUQuota > 0 && config.Resources.CPUPeriod > 0 {
		opts = append(opts, oci.WithCPUCFS(int64(config.Resources.CPUQuota), uint64(config.Resources.CPUPeriod)))
	}
	if config.Resources.PIDLimit > 0 {
		opts = append(opts, oci.WithPidsLimit(int64(config.Resources.PIDLimit)))
	}

	// Add security profile options
	opts = append(opts, securityEnforcer.BuildOCISpecOpts()...)

	// Add environment variables if provided
	if len(config.Environment) > 0 {
		envList := make([]string, 0, len(config.Environment))
		for k, v := range config.Environment {
			envList = append(envList, k+"="+v)
		}
		opts = append(opts, oci.WithEnv(envList))
	}

	// Add bind mounts (e.g., exec-agent binary + exec_key secret)
	if len(config.BindMounts) > 0 {
		var mounts []specs.Mount
		for _, bm := range config.BindMounts {
			mountOpts := []string{"rbind"}
			if bm.ReadOnly {
				mountOpts = append(mountOpts, "ro")
			}
			mounts = append(mounts, specs.Mount{
				Destination: bm.ContainerPath,
				Source:      bm.HostPath,
				Type:        "bind",
				Options:     mountOpts,
			})
		}
		opts = append(opts, oci.WithMounts(mounts))
	}

	// Add custom command if provided
	if len(config.Command) > 0 {
		opts = append(opts, oci.WithProcessArgs(append(config.Command, config.Args...)...))
	}

	// Create container
	container, err := cc.client.NewContainer(ctx, config.ID, cc.baseContainerOpts(config.ID, image, opts)...)
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	managed := &ManagedContainer{
		ID:               config.ID,
		Image:            config.ImageRef,
		Container:        container,
		Status:           types.ContainerStatusCreated,
		CreatedAt:        time.Now(),
		Resources:        config.Resources,
		SecurityEnforcer: securityEnforcer,
		DeploymentID:     config.DeploymentID,
		RequesterPubKey:  config.RequesterPubKey,
	}

	cc.mu.Lock()
	cc.containers[config.ID] = managed
	cc.mu.Unlock()

	// Apply XFS disk quota if configured
	if err := cc.SetDiskQuota(ctx, config.ID, config.Resources.DiskLimit); err != nil {
		_ = err
	}

	return managed, nil
}

// SetContainerSecurityProfile sets or updates the security profile for an existing container
// NOTE: This only affects runtime behavior (exec/attach/shell checks), not the container spec
func (cc *ContainerdClient) SetContainerSecurityProfile(id string, profile *types.ContainerSecurityProfile) error {
	cc.mu.RLock()
	managed, exists := cc.containers[id]
	cc.mu.RUnlock()

	if !exists {
		return fmt.Errorf("container not found: %s", id)
	}

	managed.mu.Lock()
	managed.SecurityEnforcer = NewSecurityEnforcer(profile)
	managed.mu.Unlock()

	return nil
}

// GetContainerSecurityProfile returns the security profile for a container
func (cc *ContainerdClient) GetContainerSecurityProfile(id string) (*types.ContainerSecurityProfile, error) {
	cc.mu.RLock()
	managed, exists := cc.containers[id]
	cc.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("container not found: %s", id)
	}

	managed.mu.RLock()
	defer managed.mu.RUnlock()

	if managed.SecurityEnforcer == nil {
		return nil, nil
	}
	return managed.SecurityEnforcer.GetProfile(), nil
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

	// Create task with log output.
	// On Linux, use FIFO-based I/O that pipes to our log writers.
	// On macOS (Colima), FIFOs can't span the VM boundary, so use
	// containerd's built-in file logger which writes inside the VM.
	// With virtiofs, the log files are accessible from macOS too.
	var taskCreator cio.Creator
	if goruntime.GOOS == "linux" {
		taskCreator = cio.NewCreator(
			cio.WithStreams(nil, containerLog.StdoutWriter(), containerLog.StderrWriter()),
		)
	} else {
		taskCreator = cio.LogFile(containerLog.StdoutPath)
	}

	task, err := managed.Container.NewTask(ctx, taskCreator)
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
		// Wait for exit with a bounded timeout to avoid hanging on unkillable processes
		select {
		case <-exitCh:
		case <-time.After(5 * time.Second):
			return fmt.Errorf("task %s did not exit after SIGKILL", id)
		}
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

	// Remove disk quota before snapshot cleanup (best-effort)
	cc.RemoveDiskQuota(ctx, id)

	// Delete container
	if err := managed.Container.Delete(ctx, containerd.WithSnapshotCleanup); err != nil {
		return fmt.Errorf("failed to delete container: %w", err)
	}

	return nil
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
