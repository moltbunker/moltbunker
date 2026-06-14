//go:build linux

package runtime

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"unsafe"

	"golang.org/x/sys/unix"

	"github.com/moltbunker/moltbunker/internal/logging"
)

// XFS ioctl constants for project quota management (x86_64 Linux).
const (
	fsIOCGetXAttr       = 0x801C581F // FS_IOC_FSGETXATTR — _IOR('X', 31, struct fsxattr)
	fsIOCSetXAttr       = 0x401C5820 // FS_IOC_FSSETXATTR — _IOW('X', 32, struct fsxattr)
	fsXFlagProjInherit  = 0x00000200 // FS_XFLAG_PROJINHERIT
	snapshotterRootPath = "/var/lib/containerd/io.containerd.snapshotter.v1.overlayfs"
)

// Filesystem magic numbers from <linux/magic.h>, used by detectFilesystemType to
// map statfs f_type into a human-readable name (R16).
const (
	magicXFS     = 0x58465342 // XFS_SUPER_MAGIC
	magicEXT     = 0xEF53     // EXT2/3/4_SUPER_MAGIC
	magicOverlay = 0x794C7630 // OVERLAYFS_SUPER_MAGIC
	magicTmpfs   = 0x01021994 // TMPFS_MAGIC
	magicBtrfs   = 0x9123683E // BTRFS_SUPER_MAGIC
	magicZFS     = 0x2FC12FC1 // ZFS_SUPER_MAGIC
)

// detectFilesystemType returns a human-readable filesystem name for the
// filesystem backing path, by inspecting statfs f_type. Unknown types are
// reported as a hex magic string so the value is still actionable in logs.
func detectFilesystemType(path string) (string, error) {
	var st unix.Statfs_t
	if err := unix.Statfs(path, &st); err != nil {
		return "", fmt.Errorf("statfs %s: %w", path, err)
	}
	// #nosec G115 -- st.Type is a kernel-supplied filesystem magic; comparison only.
	switch int64(st.Type) {
	case magicXFS:
		return "xfs", nil
	case magicEXT:
		return "ext4", nil
	case magicOverlay:
		return "overlay", nil
	case magicTmpfs:
		return "tmpfs", nil
	case magicBtrfs:
		return "btrfs", nil
	case magicZFS:
		return "zfs", nil
	default:
		return fmt.Sprintf("unknown(0x%x)", uint64(st.Type)), nil // #nosec G115 -- magic for display only
	}
}

// fsxattr mirrors the Linux struct fsxattr used by FS_IOC_FS{GET,SET}XATTR.
type fsxattr struct {
	Xflags     uint32
	Extsize    uint32
	Nextents   uint32
	Projid     uint32
	Cowextsize uint32
	Pad        [8]byte
}

// SetDiskQuota sets an XFS project quota on a container's writable snapshot layer.
// It assigns a project ID (derived from the containerd snapshot number) and sets
// a hard block limit. New files in the upper dir inherit the project ID automatically.
//
// R16: on a non-XFS snapshotter filesystem it returns a typed
// DiskQuotaNotSupportedError (instead of the old silent Warn+nil) so the caller can
// surface a visible, structured warning that disk usage is NOT being limited. The
// error is non-fatal by contract — the caller does not abort container creation;
// the disk_enforcer provides best-effort secondary enforcement.
func (cc *ContainerdClient) SetDiskQuota(ctx context.Context, containerID string, limitBytes int64) error {
	if limitBytes <= 0 {
		return nil
	}

	upperDir, projectID, err := cc.snapshotUpperDir(ctx, containerID)
	if err != nil {
		return err
	}
	if projectID < 0 {
		return fmt.Errorf("invalid snapshot project ID %d", projectID)
	}

	// R16: fail-fast on a non-XFS snapshotter filesystem. XFS project quotas only
	// work on XFS; on ext4/overlay/tmpfs the ioctl below would either error or
	// silently no-op. Surface it as a typed error so the caller can warn visibly.
	fsType, fsErr := detectFilesystemType(snapshotterRootPath)
	if fsErr != nil {
		// Could not stat the snapshotter root — fall through and let the ioctl
		// report the concrete failure rather than guessing.
		logging.Warn("disk quota: could not detect snapshotter filesystem type",
			"container_id", containerID,
			"path", snapshotterRootPath,
			"error", fsErr.Error(),
			logging.Component("disk_quota"))
	} else if fsType != "xfs" {
		return &DiskQuotaNotSupportedError{FS: fsType}
	}

	// Set XFS project ID + inheritance flag via ioctl
	// #nosec G115 -- projectID is a non-negative containerd snapshot number (guarded above), fits in uint32
	if err := setXFSProjectID(upperDir, uint32(projectID)); err != nil {
		logging.Warn("failed to set XFS project ID (non-XFS filesystem?)",
			"container_id", containerID,
			"path", upperDir,
			"error", err.Error(),
			logging.Component("disk_quota"))
		return nil // graceful degradation
	}

	// Set hard block limit via xfs_quota CLI
	limitMB := limitBytes / (1024 * 1024)
	if limitMB < 1 {
		limitMB = 1
	}

	// #nosec G204 -- exec.CommandContext (no shell); command name is the constant "xfs_quota", args are internally formatted from a numeric limit and snapshot project ID
	cmd := exec.CommandContext(ctx, "xfs_quota", "-xc",
		fmt.Sprintf("limit -p bhard=%dm %d", limitMB, projectID),
		snapshotterRootPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("xfs_quota limit: %s: %w", strings.TrimSpace(string(out)), err)
	}

	logging.Info("disk quota set",
		"container_id", containerID,
		"project_id", projectID,
		"limit_mb", limitMB,
		logging.Component("disk_quota"))

	return nil
}

// RemoveDiskQuota resets the XFS project quota for a container (best-effort).
func (cc *ContainerdClient) RemoveDiskQuota(ctx context.Context, containerID string) {
	_, projectID, err := cc.snapshotUpperDir(ctx, containerID)
	if err != nil {
		return // snapshot already gone
	}

	// #nosec G204 -- exec.CommandContext (no shell); command name is the constant "xfs_quota", args are internally formatted from the snapshot project ID
	cmd := exec.CommandContext(ctx, "xfs_quota", "-xc",
		fmt.Sprintf("limit -p bhard=0 %d", projectID),
		snapshotterRootPath)
	_ = cmd.Run()
}

// snapshotUpperDir returns the overlay upper directory path and the containerd
// snapshot numeric ID for a container. The numeric ID is used as the XFS project ID.
func (cc *ContainerdClient) snapshotUpperDir(ctx context.Context, containerID string) (string, int, error) {
	ctx = cc.WithNamespace(ctx)
	snapshotter := cc.client.SnapshotService("")
	mounts, err := snapshotter.Mounts(ctx, containerID+"-snapshot")
	if err != nil {
		return "", 0, fmt.Errorf("get snapshot mounts: %w", err)
	}

	var upperDir string
	for _, m := range mounts {
		for _, opt := range m.Options {
			if strings.HasPrefix(opt, "upperdir=") {
				upperDir = strings.TrimPrefix(opt, "upperdir=")
			}
		}
	}
	if upperDir == "" {
		return "", 0, fmt.Errorf("no upperdir in snapshot mounts for %s", containerID)
	}

	// Parse snapshot numeric ID: /path/snapshots/42/fs → 42
	snapshotDir := filepath.Dir(upperDir)
	projectID, err := strconv.Atoi(filepath.Base(snapshotDir))
	if err != nil {
		return "", 0, fmt.Errorf("parse snapshot ID from %s: %w", snapshotDir, err)
	}

	return upperDir, projectID, nil
}

// setXFSProjectID sets the project ID on a directory with PROJINHERIT via ioctl.
// All new files/subdirectories created under this directory will inherit the project ID.
func setXFSProjectID(dir string, projectID uint32) error {
	// #nosec G304 -- dir is the overlay upperdir reported by the containerd snapshotter, not external/request input
	f, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	var attr fsxattr
	// #nosec G103 -- unsafe.Pointer required to pass the fsxattr struct to the kernel (FS_IOC_FSGETXATTR)
	if _, _, errno := unix.Syscall(unix.SYS_IOCTL, f.Fd(), fsIOCGetXAttr, uintptr(unsafe.Pointer(&attr))); errno != 0 {
		return fmt.Errorf("FS_IOC_FSGETXATTR: %w", errno)
	}

	attr.Projid = projectID
	attr.Xflags |= fsXFlagProjInherit

	// #nosec G103 -- unsafe.Pointer required to pass the fsxattr struct to the kernel (FS_IOC_FSSETXATTR)
	if _, _, errno := unix.Syscall(unix.SYS_IOCTL, f.Fd(), fsIOCSetXAttr, uintptr(unsafe.Pointer(&attr))); errno != 0 {
		return fmt.Errorf("FS_IOC_FSSETXATTR: %w", errno)
	}

	return nil
}
