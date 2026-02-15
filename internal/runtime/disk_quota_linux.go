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
// On non-XFS filesystems or if xfs_quota is missing, it logs a warning and returns nil
// (graceful degradation — the disk_enforcer provides secondary enforcement).
func (cc *ContainerdClient) SetDiskQuota(ctx context.Context, containerID string, limitBytes int64) error {
	if limitBytes <= 0 {
		return nil
	}

	upperDir, projectID, err := cc.snapshotUpperDir(ctx, containerID)
	if err != nil {
		return err
	}

	// Set XFS project ID + inheritance flag via ioctl
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
	f, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer f.Close()

	var attr fsxattr
	if _, _, errno := unix.Syscall(unix.SYS_IOCTL, f.Fd(), fsIOCGetXAttr, uintptr(unsafe.Pointer(&attr))); errno != 0 {
		return fmt.Errorf("FS_IOC_FSGETXATTR: %w", errno)
	}

	attr.Projid = projectID
	attr.Xflags |= fsXFlagProjInherit

	if _, _, errno := unix.Syscall(unix.SYS_IOCTL, f.Fd(), fsIOCSetXAttr, uintptr(unsafe.Pointer(&attr))); errno != 0 {
		return fmt.Errorf("FS_IOC_FSSETXATTR: %w", errno)
	}

	return nil
}
