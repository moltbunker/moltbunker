//go:build !linux

package runtime

import "context"

// SetDiskQuota is a no-op on non-Linux platforms (XFS project quotas are Linux-only).
func (cc *ContainerdClient) SetDiskQuota(_ context.Context, _ string, _ int64) error {
	return nil
}

// RemoveDiskQuota is a no-op on non-Linux platforms.
func (cc *ContainerdClient) RemoveDiskQuota(_ context.Context, _ string) {}
