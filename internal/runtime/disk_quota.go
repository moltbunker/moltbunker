package runtime

import "fmt"

// DiskQuotaNotSupportedError reports that an XFS project quota could not be set
// because the snapshotter's backing filesystem is not XFS.
//
// R16: SetDiskQuota previously logged a Warn and returned nil on non-XFS hosts,
// which silently discarded the fact that the container's disk usage is NOT being
// limited. Returning this typed error lets the caller surface a visible,
// structured warning (with the detected filesystem) while keeping the
// non-fatal semantics — container creation is not aborted, because the
// disk_enforcer provides best-effort secondary enforcement.
//
// Defined in a build-neutral file so callers compiled on all platforms (e.g.
// container_lifecycle.go) can match it via errors.As without a build-tag split.
type DiskQuotaNotSupportedError struct {
	FS string // detected filesystem type, e.g. "ext4", "overlay", "tmpfs"
}

func (e *DiskQuotaNotSupportedError) Error() string {
	fs := e.FS
	if fs == "" {
		fs = "unknown"
	}
	return fmt.Sprintf("disk quota unavailable: snapshotter filesystem is %s, not XFS (XFS project quotas required)", fs)
}
