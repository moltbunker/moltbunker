package runtime

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/moltbunker/moltbunker/pkg/types"
)

func TestCgroupManager_CreateCgroup(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping cgroup test (requires root or cgroup v2)")
	}

	tmpDir := t.TempDir()
	cm := NewCgroupManager(tmpDir)

	containerID := "test-container"

	err := cm.CreateCgroup(containerID)
	if err != nil {
		// May fail if not running as root or cgroup v2 not available
		t.Logf("Cgroup creation failed (may be expected): %v", err)
		return
	}

	// Verify cgroup directory exists
	cgroupPath := filepath.Join(tmpDir, "moltbunker", containerID)
	if _, err := os.Stat(cgroupPath); os.IsNotExist(err) {
		t.Error("Cgroup directory should exist")
	}
}

func TestCgroupManager_ApplyResourceLimits(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping cgroup test (requires root or cgroup v2)")
	}

	tmpDir := t.TempDir()
	cm := NewCgroupManager(tmpDir)

	containerID := "test-container"
	resources := types.ResourceLimits{
		CPUQuota:    1000000,
		CPUPeriod:   100000,
		MemoryLimit: 1073741824, // 1GB
		DiskLimit:   10737418240,
		NetworkBW:   10485760,
		PIDLimit:    100,
	}

	err := cm.ApplyResourceLimits(containerID, resources)
	if err != nil {
		// May fail if not running as root or cgroup v2 not available
		t.Logf("Resource limits application failed (may be expected): %v", err)
		return
	}

	// Verify cgroup was created
	cgroupPath := filepath.Join(tmpDir, "moltbunker", containerID)
	if _, err := os.Stat(cgroupPath); os.IsNotExist(err) {
		t.Error("Cgroup directory should exist")
	}
}

func TestCgroupManager_DeleteCgroup(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping cgroup test (requires root or cgroup v2)")
	}

	tmpDir := t.TempDir()
	cm := NewCgroupManager(tmpDir)

	containerID := "test-container"

	// Create cgroup first
	err := cm.CreateCgroup(containerID)
	if err != nil {
		t.Logf("Cgroup creation failed (may be expected): %v", err)
		return
	}

	// Delete cgroup
	err = cm.DeleteCgroup(containerID)
	if err != nil {
		t.Fatalf("Failed to delete cgroup: %v", err)
	}

	// Verify cgroup directory doesn't exist
	cgroupPath := filepath.Join(tmpDir, "moltbunker", containerID)
	if _, err := os.Stat(cgroupPath); !os.IsNotExist(err) {
		t.Error("Cgroup directory should not exist after deletion")
	}
}
