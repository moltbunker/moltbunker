package runtime

import (
	"context"
	"errors"
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

	"github.com/moltbunker/moltbunker/internal/logging"
	"github.com/moltbunker/moltbunker/pkg/types"
)

// applyDiskQuota applies the configured XFS disk quota for a container and
// surfaces failures as visible, structured warnings. The non-XFS case (R16) gets a
// filesystem-specific message; all OTHER SetDiskQuota errors (xfs_quota CLI errors,
// FS_IOC_FSSETXATTR ioctl errno, upperdir parse failures) are also logged at Warn
// rather than swallowed — otherwise a misconfigured XFS host would silently run
// containers with NO disk limit, which is exactly the silent-no-op class R16 set
// out to eliminate. All cases remain non-fatal: container creation is never aborted
// here (the disk_enforcer provides best-effort secondary enforcement).
func (cc *ContainerdClient) applyDiskQuota(ctx context.Context, id string, limitBytes int64) {
	err := cc.SetDiskQuota(ctx, id, limitBytes)
	if err == nil {
		return
	}
	var notXFS *DiskQuotaNotSupportedError
	if errors.As(err, &notXFS) {
		logging.Warn("disk quota unavailable on non-XFS filesystem; container disk usage will not be limited",
			"filesystem_type", notXFS.FS,
			logging.ContainerID(id),
			logging.Component("disk_quota"))
		return
	}
	// Other errors are non-fatal but must be visible (R16): a real XFS-host quota
	// failure means the container runs with no disk limit, so log it at Warn.
	logging.Warn("disk quota not applied; container disk usage will not be limited",
		logging.ContainerID(id),
		logging.Err(err),
		logging.Component("disk_quota"))
}

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
	// R17: the EFFECTIVE PID ceiling for Kata VM workloads is the OCI
	// linux.resources.pids.limit (wired in CreateContainer via oci.WithPidsLimit;
	// defaults to 100 on the deploy path). The kata-agent enforces that cgroup
	// pids.limit INSIDE the guest, so it is the mechanism that actually bounds
	// process count for VM workloads — not this annotation.
	//
	// NOTE: io.katacontainers.config.hypervisor.default_pids is NOT a recognized
	// Kata hypervisor annotation (Kata's hypervisor annotation set is
	// default_memory/default_vcpus/kernel/image/initrd/machine_type/etc.), and Kata
	// additionally drops any annotation not in the runtime TOML enable_annotations
	// allow-list. So this is almost certainly INERT today. It is emitted only as a
	// forward-looking hint and requires (a) an enable_annotations entry in the Kata
	// runtime config and (b) validation against a real Kata shim under R11 before it
	// can be relied on. Do not treat it as the PID ceiling — the OCI pids.limit is.
	if cc.kataConfig.DefaultPIDs > 0 {
		annotations["io.katacontainers.config.hypervisor.default_pids"] = strconv.Itoa(cc.kataConfig.DefaultPIDs)
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

	// R3 — image signature verification.
	// If TrustPolicy.RequireSignature is true, ImageSignature must verify
	// against one of TrustPolicy.TrustedPublishers before the container is
	// created. Leave both zero-valued to opt out (current default).
	ImageSignature *ImageSignature
	TrustPolicy    TrustPolicy

	// R4 — image vulnerability scanning.
	// If Scanner is non-nil it is invoked after the image is fetched and
	// before the container spec is built. Findings are evaluated against
	// ScanPolicy; a policy violation aborts container creation.
	// Use NewNoopScanner() to satisfy the interface when scanning is
	// disabled by daemon config.
	Scanner    ImageScanner
	ScanPolicy ScanPolicy

	// R5 — image content encryption at rest.
	// If ImageCrypter is non-nil and Enabled(), encrypted image layers in the
	// local content store are decrypted in-process (using ImageDecryptKey, this
	// node's stable X25519 private key) BEFORE the rootfs is unpacked. Plaintext
	// images are detected via imgcrypt's HasEncryptedLayer and pass through
	// untouched, so this is a safe no-op for unencrypted public images.
	//
	// ImageEncryptRecipients, when non-empty, are the X25519 public keys (this
	// deployment's providers — originator + replicas) the image should be
	// (re)encrypted to at rest after a successful pull. The daemon zone sources
	// these from the gossiped Deployment so every replica can decrypt the same
	// image. Leave ImageCrypter as a NoopImageCrypter (or nil) to opt out.
	ImageCrypter           ImageCrypter
	ImageDecryptKey        []byte
	ImageEncryptRecipients [][]byte
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

	// Apply XFS disk quota if configured (R16: warns on non-XFS, non-fatal).
	cc.applyDiskQuota(ctx, id, resources.DiskLimit)

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

	// Apply XFS disk quota if configured (R16: warns on non-XFS, non-fatal).
	cc.applyDiskQuota(ctx, id, resources.DiskLimit)

	return managed, nil
}

// CreateSecureContainer creates a container with a full security profile for provider nodes
// This is the primary method for creating containers that enforce container opacity
func (cc *ContainerdClient) CreateSecureContainer(ctx context.Context, config SecureContainerConfig) (*ManagedContainer, error) {
	ctx = cc.WithNamespace(ctx)

	// Get or pull image — with R3 signature verification when a policy is set.
	image, err := cc.GetImage(ctx, config.ImageRef)
	if err != nil {
		if config.TrustPolicy.RequireSignature || config.ImageSignature != nil {
			image, err = cc.PullImageVerified(ctx, config.ImageRef, config.ImageSignature, config.TrustPolicy, nil)
		} else {
			image, err = cc.PullImage(ctx, config.ImageRef)
		}
		if err != nil {
			return nil, fmt.Errorf("failed to get image: %w", err)
		}
	} else if config.TrustPolicy.RequireSignature || config.ImageSignature != nil {
		// Image already in local content store; still enforce the policy on
		// its digest. We do NOT delete on failure here because other callers
		// may legitimately have this image without a signature.
		digest := ImageDigest(image.Target().Digest.String())
		verifier := NewEdImageVerifier()
		if verifyErr := verifier.Verify(digest, config.ImageSignature, config.TrustPolicy); verifyErr != nil {
			return nil, fmt.Errorf("image %s (digest %s) failed signature verification: %w", config.ImageRef, digest, verifyErr)
		}
	}

	// R5 — decrypt encrypted image layers at rest before scanning/unpack. The
	// image must be plaintext for the scanner to read it and for WithNewSnapshot
	// to build the rootfs. decryptImageForRun is a safe no-op for plaintext
	// images (it detects encrypted layers by digest), so this never affects
	// unencrypted public images. Build-verified; runtime-validated under R11.
	if config.ImageCrypter != nil && config.ImageCrypter.Enabled() && len(config.ImageDecryptKey) > 0 {
		image, err = cc.decryptImageForRun(ctx, image, config.ImageRef, config.ImageCrypter, config.ImageDecryptKey)
		if err != nil {
			return nil, fmt.Errorf("image %s: %w", config.ImageRef, err)
		}
	}

	// R4 — vulnerability scan gate. Runs after the image is local but before
	// any container resource is allocated.
	if config.Scanner != nil {
		report, scanErr := config.Scanner.Scan(ctx, config.ImageRef)
		if scanErr != nil {
			return nil, fmt.Errorf("image %s scan failed: %w", config.ImageRef, scanErr)
		}
		if config.ScanPolicy.RequireScan && (report == nil || len(report.Vulnerabilities) == 0 && config.Scanner.ID() == "noop") {
			return nil, fmt.Errorf("image %s: %w", config.ImageRef, ErrScanRequired)
		}
		if report != nil {
			if _, policyErr := config.ScanPolicy.Apply(report.Vulnerabilities); policyErr != nil {
				return nil, fmt.Errorf("image %s: %w", config.ImageRef, policyErr)
			}
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

	// R5 — encrypt the durable image record's layer blobs at rest now that the
	// container's rootfs snapshot has been unpacked. Self-recipient: sealed to
	// this node's own X25519 key (config.ImageEncryptRecipients). Best-effort and
	// FAIL-OPEN by design: a failure does NOT tear down an otherwise-healthy
	// container (the running rootfs already lives in the snapshot); it emits a
	// Warn and continues. Caveat (honest): even on success the ORIGINAL plaintext
	// blobs stay pinned by this container's GC ref until it stops (see
	// image_encrypt_store.go limitation #1), and the fail-open downgrade is
	// surfaced only by this log line — no metric/state flag yet (R11 follow-up).
	// No-op when encryption is disabled. Build-verified; runtime-validated under R11.
	if config.ImageCrypter != nil && config.ImageCrypter.Enabled() && len(config.ImageEncryptRecipients) > 0 {
		if encErr := cc.EncryptImageAtRest(ctx, config.ImageRef, config.ImageCrypter, config.ImageEncryptRecipients); encErr != nil {
			logging.Warn("R5: failed to encrypt image at rest; image content remains plaintext on disk",
				logging.Err(encErr))
		}
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

	// Apply XFS disk quota if configured (R16: warns on non-XFS, non-fatal).
	cc.applyDiskQuota(ctx, config.ID, config.Resources.DiskLimit)

	// R20 — persist the security profile so daemon restart can restore it
	// instead of silently downgrading to the default. Non-fatal: a write
	// failure logs a warning but does not abort container creation.
	if cc.profileStore != nil && config.SecurityProfile != nil {
		if writeErr := cc.profileStore.Write(config.ID, config.SecurityProfile); writeErr != nil {
			logging.Warn("profile store: failed to persist profile",
				logging.ContainerID(config.ID),
				logging.Err(writeErr))
		}
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
		if closeErr := cc.logManager.CloseLog(id); closeErr != nil {
			logging.Warn("failed to close container log after task creation failure",
				logging.ContainerID(id),
				logging.Err(closeErr))
		}
		return fmt.Errorf("failed to create task: %w", err)
	}

	// Start task
	if err := task.Start(ctx); err != nil {
		if _, delErr := task.Delete(ctx); delErr != nil {
			logging.Warn("failed to delete task after start failure",
				logging.ContainerID(id),
				logging.Err(delErr))
		}
		if closeErr := cc.logManager.CloseLog(id); closeErr != nil {
			logging.Warn("failed to close container log after task start failure",
				logging.ContainerID(id),
				logging.Err(closeErr))
		}
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
	if closeErr := cc.logManager.CloseLog(id); closeErr != nil {
		logging.Warn("failed to close container log after stop",
			logging.ContainerID(id),
			logging.Err(closeErr))
	}

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
		if killErr := managed.Task.Kill(ctx, syscall.SIGKILL); killErr != nil {
			logging.Warn("failed to send SIGKILL during container delete",
				logging.ContainerID(id),
				logging.Err(killErr))
		}
		if _, delErr := managed.Task.Delete(ctx); delErr != nil {
			logging.Warn("failed to delete task during container delete",
				logging.ContainerID(id),
				logging.Err(delErr))
		}
	}

	// Close and delete logs
	if delErr := cc.logManager.DeleteLog(id); delErr != nil {
		logging.Warn("failed to delete container log",
			logging.ContainerID(id),
			logging.Err(delErr))
	}

	// Remove disk quota before snapshot cleanup (best-effort)
	cc.RemoveDiskQuota(ctx, id)

	// Delete container
	if err := managed.Container.Delete(ctx, containerd.WithSnapshotCleanup); err != nil {
		return fmt.Errorf("failed to delete container: %w", err)
	}

	// R20 — drop the persisted profile sidecar so it doesn't leak on a
	// future container reusing the same ID. Best-effort.
	if cc.profileStore != nil {
		if delErr := cc.profileStore.Delete(id); delErr != nil {
			logging.Warn("profile store: failed to delete profile",
				logging.ContainerID(id),
				logging.Err(delErr))
		}
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
