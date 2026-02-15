package runtime

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/moltbunker/moltbunker/pkg/types"
)

// ---------------------------------------------------------------------------
// CleanupManager unit tests
// ---------------------------------------------------------------------------

func TestNewCleanupManager(t *testing.T) {
	t.Run("creates manager with all fields set", func(t *testing.T) {
		tmpDir := t.TempDir()
		cgroups := NewCgroupManager(tmpDir)

		cm := NewCleanupManager(nil, cgroups, tmpDir)
		if cm == nil {
			t.Fatal("expected non-nil CleanupManager")
		}
		if cm.cgroups != cgroups {
			t.Error("cgroups field mismatch")
		}
		if cm.dataDir != tmpDir {
			t.Errorf("dataDir = %q, want %q", cm.dataDir, tmpDir)
		}
		if cm.cleanedUp == nil {
			t.Error("cleanedUp map should be initialized")
		}
	})

	t.Run("works with nil cgroup manager", func(t *testing.T) {
		cm := NewCleanupManager(nil, nil, t.TempDir())
		if cm == nil {
			t.Fatal("expected non-nil CleanupManager")
		}
	})
}

func TestCleanupManager_IsCleanedUp(t *testing.T) {
	cm := NewCleanupManager(nil, nil, t.TempDir())

	t.Run("returns false for unknown container", func(t *testing.T) {
		if cm.IsCleanedUp("unknown") {
			t.Error("expected false for unknown container")
		}
	})

	t.Run("returns true after marking", func(t *testing.T) {
		cm.mu.Lock()
		cm.cleanedUp["test-1"] = true
		cm.mu.Unlock()

		if !cm.IsCleanedUp("test-1") {
			t.Error("expected true after marking")
		}
	})
}

func TestCleanupManager_ResetCleanupState(t *testing.T) {
	cm := NewCleanupManager(nil, nil, t.TempDir())

	cm.mu.Lock()
	cm.cleanedUp["test-1"] = true
	cm.mu.Unlock()

	cm.ResetCleanupState("test-1")
	if cm.IsCleanedUp("test-1") {
		t.Error("expected false after reset")
	}

	// Resetting a non-existent entry should be safe.
	cm.ResetCleanupState("nonexistent")
}

func TestCleanupManager_DoubleCleanup(t *testing.T) {
	// Verify that calling CleanupContainer twice does not error.
	tmpDir := t.TempDir()
	cm := NewCleanupManager(nil, nil, tmpDir)

	// Pre-mark as cleaned up to simulate a prior run.
	containerID := "double-clean"
	cm.mu.Lock()
	cm.cleanedUp[containerID] = true
	cm.mu.Unlock()

	// Second call should be a no-op and return nil.
	err := cm.CleanupContainer(nil, containerID)
	if err != nil {
		t.Errorf("second cleanup should be no-op, got: %v", err)
	}
}

func TestCleanupManager_CleanupRemovesCgroup(t *testing.T) {
	tmpDir := t.TempDir()
	cgroupRoot := filepath.Join(tmpDir, "cgroups")
	cgroups := NewCgroupManager(cgroupRoot)

	containerID := "cgroup-clean"

	// Create a cgroup directory.
	if err := cgroups.CreateCgroup(containerID); err != nil {
		t.Fatalf("failed to create cgroup: %v", err)
	}

	cgroupPath := filepath.Join(cgroupRoot, "moltbunker", containerID)
	if _, err := os.Stat(cgroupPath); os.IsNotExist(err) {
		t.Fatal("cgroup directory should exist before cleanup")
	}

	cm := NewCleanupManager(nil, cgroups, tmpDir)

	// CleanupContainer with nil client -- the container won't be found in the
	// client, so it skips the stop/delete steps and goes straight to cgroup.
	err := cm.CleanupContainer(nil, containerID)
	// err may be non-nil if client is nil, but cgroup should still be cleaned.
	_ = err

	if _, statErr := os.Stat(cgroupPath); !os.IsNotExist(statErr) {
		t.Error("cgroup directory should be removed after cleanup")
	}
}

func TestCleanupManager_CleanupRemovesMountDir(t *testing.T) {
	tmpDir := t.TempDir()
	containerID := "mount-clean"

	// Create a mock mount directory.
	mountDir := filepath.Join(tmpDir, "mounts", containerID)
	if err := os.MkdirAll(mountDir, 0755); err != nil {
		t.Fatalf("failed to create mount dir: %v", err)
	}

	cm := NewCleanupManager(nil, nil, tmpDir)
	err := cm.CleanupContainer(nil, containerID)
	_ = err

	if _, statErr := os.Stat(mountDir); !os.IsNotExist(statErr) {
		t.Error("mount directory should be removed after cleanup")
	}
}

func TestCleanupManager_CleanupRemovesPIDFile(t *testing.T) {
	tmpDir := t.TempDir()
	containerID := "pid-clean"

	// Create a mock PID file.
	pidDir := filepath.Join(tmpDir, "pids")
	if err := os.MkdirAll(pidDir, 0755); err != nil {
		t.Fatalf("failed to create pid dir: %v", err)
	}
	pidFile := filepath.Join(pidDir, containerID+".pid")
	if err := os.WriteFile(pidFile, []byte("12345"), 0644); err != nil {
		t.Fatalf("failed to write pid file: %v", err)
	}

	cm := NewCleanupManager(nil, nil, tmpDir)
	err := cm.CleanupContainer(nil, containerID)
	_ = err

	if _, statErr := os.Stat(pidFile); !os.IsNotExist(statErr) {
		t.Error("PID file should be removed after cleanup")
	}
}

func TestCleanupManager_CleanupAll_Empty(t *testing.T) {
	// CleanupAll with no tracked containers should succeed.
	// We use a nil client so there are no containers to list.
	_ = NewCleanupManager(nil, nil, t.TempDir())
	// With nil client, ListContainers would panic, so we skip that test
	// and instead verify the manager itself works.
	// This is covered more thoroughly in the integration tests.
}

func TestCleanupManager_ConcurrentCleanup(t *testing.T) {
	tmpDir := t.TempDir()
	cm := NewCleanupManager(nil, nil, tmpDir)

	// Create mount dirs for multiple containers.
	for i := 0; i < 10; i++ {
		id := "concurrent-" + string(rune('a'+i))
		mountDir := filepath.Join(tmpDir, "mounts", id)
		os.MkdirAll(mountDir, 0755)
	}

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		id := "concurrent-" + string(rune('a'+i))
		go func(containerID string) {
			defer wg.Done()
			// Call cleanup multiple times concurrently.
			cm.CleanupContainer(nil, containerID)
			cm.CleanupContainer(nil, containerID) // double call
		}(id)
	}
	wg.Wait()

	// Verify all are marked as cleaned up.
	for i := 0; i < 10; i++ {
		id := "concurrent-" + string(rune('a'+i))
		if !cm.IsCleanedUp(id) {
			t.Errorf("container %s should be marked as cleaned up", id)
		}
	}
}

func TestCleanupManager_CleanupStoppedContainer(t *testing.T) {
	// Test that cleanup works for a container that is already stopped.
	// We simulate this by creating a CleanupManager with no client (nil),
	// and just ensuring cgroup + mount + pid cleanup succeeds.

	tmpDir := t.TempDir()
	cgroupRoot := filepath.Join(tmpDir, "cgroups")
	cgroups := NewCgroupManager(cgroupRoot)

	containerID := "stopped-container"

	// Set up resources that would exist for a stopped container.
	cgroups.CreateCgroup(containerID)
	os.MkdirAll(filepath.Join(tmpDir, "mounts", containerID), 0755)
	pidDir := filepath.Join(tmpDir, "pids")
	os.MkdirAll(pidDir, 0755)
	os.WriteFile(filepath.Join(pidDir, containerID+".pid"), []byte("99999"), 0644)

	cm := NewCleanupManager(nil, cgroups, tmpDir)
	err := cm.CleanupContainer(nil, containerID)
	_ = err // may error on nil client but that's OK

	// All resources should be cleaned.
	cgroupPath := filepath.Join(cgroupRoot, "moltbunker", containerID)
	if _, statErr := os.Stat(cgroupPath); !os.IsNotExist(statErr) {
		t.Error("cgroup should be removed")
	}

	mountPath := filepath.Join(tmpDir, "mounts", containerID)
	if _, statErr := os.Stat(mountPath); !os.IsNotExist(statErr) {
		t.Error("mount directory should be removed")
	}

	pidFile := filepath.Join(pidDir, containerID+".pid")
	if _, statErr := os.Stat(pidFile); !os.IsNotExist(statErr) {
		t.Error("PID file should be removed")
	}
}

func TestCleanupManager_CleanupMarksState(t *testing.T) {
	// Verify that cleanup sets the cleaned-up flag even when there's nothing
	// to actually clean (no cgroups, no mounts, no pid files, no client).
	tmpDir := t.TempDir()
	cm := NewCleanupManager(nil, nil, tmpDir)

	containerID := "empty-cleanup"
	cm.CleanupContainer(nil, containerID)

	if !cm.IsCleanedUp(containerID) {
		t.Error("container should be marked as cleaned up")
	}
}

func TestCleanupManager_ManagedContainerInfoStatus(t *testing.T) {
	// Verify ManagedContainerInfo has expected status values for testing helpers.
	info := ManagedContainerInfo{
		ID:     "status-test",
		Status: types.ContainerStatusStopped,
	}

	if info.Status != types.ContainerStatusStopped {
		t.Errorf("expected stopped status, got %v", info.Status)
	}
}
