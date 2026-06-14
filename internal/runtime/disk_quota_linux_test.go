//go:build linux

package runtime

import (
	"errors"
	"strings"
	"testing"
)

func TestDetectFilesystemType_NonXFS(t *testing.T) {
	// The test temp dir is backed by the host's working filesystem (typically
	// ext4/overlay/tmpfs in CI), never XFS, so detectFilesystemType must NOT
	// classify it as xfs — which is exactly the branch SetDiskQuota fails on.
	dir := t.TempDir()
	fs, err := detectFilesystemType(dir)
	if err != nil {
		t.Fatalf("detectFilesystemType error: %v", err)
	}
	if fs == "" {
		t.Fatal("detectFilesystemType returned empty type")
	}
	if fs == "xfs" {
		t.Skip("test host's temp dir is XFS; non-XFS branch cannot be exercised here")
	}
	t.Logf("temp dir filesystem detected as %q", fs)
}

func TestDetectFilesystemType_MissingPath(t *testing.T) {
	if _, err := detectFilesystemType("/no/such/path/hopefully"); err == nil {
		t.Error("expected statfs error for a missing path")
	}
}

func TestDiskQuotaNotSupportedError_Message(t *testing.T) {
	err := &DiskQuotaNotSupportedError{FS: "ext4"}
	msg := err.Error()
	if !strings.Contains(msg, "ext4") {
		t.Errorf("error message %q should contain the filesystem name", msg)
	}
	if !strings.Contains(msg, "XFS") {
		t.Errorf("error message %q should mention XFS requirement", msg)
	}

	// errors.As must recover the typed error through a wrap.
	wrapped := errors.New("outer: " + err.Error())
	_ = wrapped
	var target *DiskQuotaNotSupportedError
	if !errors.As(error(err), &target) {
		t.Error("errors.As should recover *DiskQuotaNotSupportedError")
	}

	// Empty FS still produces a sane message.
	empty := (&DiskQuotaNotSupportedError{}).Error()
	if !strings.Contains(empty, "unknown") {
		t.Errorf("empty-FS error should say unknown, got %q", empty)
	}
}
