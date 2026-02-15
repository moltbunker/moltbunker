package runtime

import (
	"bytes"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/moltbunker/moltbunker/pkg/types"
)

// ---------------------------------------------------------------------------
// EncryptionManager tests
// ---------------------------------------------------------------------------

func TestNewEncryptionManager(t *testing.T) {
	t.Run("creates required directories", func(t *testing.T) {
		tmpDir := t.TempDir()
		dataDir := filepath.Join(tmpDir, "encryption-data")

		em, err := NewEncryptionManager(dataDir)
		if err != nil {
			t.Fatalf("NewEncryptionManager failed: %v", err)
		}
		if em == nil {
			t.Fatal("expected non-nil EncryptionManager")
		}

		// Verify directories were created
		for _, sub := range []string{"volumes", "mounts", "keys"} {
			dir := filepath.Join(dataDir, sub)
			info, err := os.Stat(dir)
			if err != nil {
				t.Errorf("directory %s should exist: %v", sub, err)
				continue
			}
			if !info.IsDir() {
				t.Errorf("%s should be a directory", sub)
			}
		}
	})

	t.Run("handles read-only parent directory gracefully", func(t *testing.T) {
		// This should fail because the directory can't be created
		_, err := NewEncryptionManager("/proc/nonexistent/encryption")
		if err == nil {
			t.Error("expected error creating encryption manager with invalid path")
		}
	})

	t.Run("fields are initialized correctly", func(t *testing.T) {
		tmpDir := t.TempDir()
		em, err := NewEncryptionManager(tmpDir)
		if err != nil {
			t.Fatalf("NewEncryptionManager failed: %v", err)
		}

		if em.dataDir != tmpDir {
			t.Errorf("dataDir = %q, want %q", em.dataDir, tmpDir)
		}
		if em.volumesDir != filepath.Join(tmpDir, "volumes") {
			t.Errorf("volumesDir = %q, want %q", em.volumesDir, filepath.Join(tmpDir, "volumes"))
		}
		if em.mountsDir != filepath.Join(tmpDir, "mounts") {
			t.Errorf("mountsDir = %q, want %q", em.mountsDir, filepath.Join(tmpDir, "mounts"))
		}
		if em.keysDir != filepath.Join(tmpDir, "keys") {
			t.Errorf("keysDir = %q, want %q", em.keysDir, filepath.Join(tmpDir, "keys"))
		}
		if em.volumes == nil {
			t.Error("volumes map should be initialized")
		}
	})
}

func TestEncryptionManager_GenerateKey(t *testing.T) {
	tmpDir := t.TempDir()
	em, err := NewEncryptionManager(tmpDir)
	if err != nil {
		t.Fatalf("NewEncryptionManager failed: %v", err)
	}

	t.Run("returns 32-byte key", func(t *testing.T) {
		key, err := em.GenerateKey()
		if err != nil {
			t.Fatalf("GenerateKey failed: %v", err)
		}
		if len(key) != 32 {
			t.Errorf("key length = %d, want 32", len(key))
		}
	})

	t.Run("generates unique keys", func(t *testing.T) {
		key1, err := em.GenerateKey()
		if err != nil {
			t.Fatalf("GenerateKey 1 failed: %v", err)
		}
		key2, err := em.GenerateKey()
		if err != nil {
			t.Fatalf("GenerateKey 2 failed: %v", err)
		}
		if bytes.Equal(key1, key2) {
			t.Error("consecutive keys should be different (extremely unlikely to be equal)")
		}
	})

	t.Run("key is valid hex-encodable", func(t *testing.T) {
		key, err := em.GenerateKey()
		if err != nil {
			t.Fatalf("GenerateKey failed: %v", err)
		}
		hexStr := hex.EncodeToString(key)
		if len(hexStr) != 64 {
			t.Errorf("hex-encoded key length = %d, want 64", len(hexStr))
		}
	})
}

func TestEncryptionManager_GetVolume(t *testing.T) {
	tmpDir := t.TempDir()
	em, err := NewEncryptionManager(tmpDir)
	if err != nil {
		t.Fatalf("NewEncryptionManager failed: %v", err)
	}

	t.Run("returns false for nonexistent volume", func(t *testing.T) {
		vol, exists := em.GetVolume("nonexistent")
		if exists {
			t.Error("should not find nonexistent volume")
		}
		if vol != nil {
			t.Error("volume should be nil for nonexistent")
		}
	})

	t.Run("returns volume after manual insertion", func(t *testing.T) {
		// Directly insert a volume into the map for testing
		em.mu.Lock()
		em.volumes["test-container"] = &EncryptedVolume{
			ContainerID: "test-container",
			VolumePath:  "/tmp/test.img",
			MapperName:  "moltbunker-test-container",
			SizeGB:      5,
			Mounted:     false,
		}
		em.mu.Unlock()

		vol, exists := em.GetVolume("test-container")
		if !exists {
			t.Error("should find test-container volume")
		}
		if vol == nil {
			t.Fatal("volume should not be nil")
		}
		if vol.ContainerID != "test-container" {
			t.Errorf("ContainerID = %q, want %q", vol.ContainerID, "test-container")
		}
		if vol.SizeGB != 5 {
			t.Errorf("SizeGB = %d, want 5", vol.SizeGB)
		}
	})
}

func TestEncryptionManager_GetKeyAsHex(t *testing.T) {
	tmpDir := t.TempDir()
	em, err := NewEncryptionManager(tmpDir)
	if err != nil {
		t.Fatalf("NewEncryptionManager failed: %v", err)
	}

	t.Run("error for nonexistent volume", func(t *testing.T) {
		_, err := em.GetKeyAsHex("nonexistent")
		if err == nil {
			t.Error("expected error for nonexistent volume")
		}
		if !strings.Contains(err.Error(), "volume not found") {
			t.Errorf("expected 'volume not found' error, got: %v", err)
		}
	})

	t.Run("returns hex key for volume with key file", func(t *testing.T) {
		// Create a key file
		keyData := []byte("abcdefghijklmnopqrstuvwxyz012345") // 32 bytes
		keyPath := filepath.Join(tmpDir, "keys", "hextest.key")
		if err := os.WriteFile(keyPath, keyData, 0400); err != nil {
			t.Fatalf("failed to write key file: %v", err)
		}

		em.mu.Lock()
		em.volumes["hextest"] = &EncryptedVolume{
			ContainerID: "hextest",
			KeyPath:     keyPath,
		}
		em.mu.Unlock()

		hexStr, err := em.GetKeyAsHex("hextest")
		if err != nil {
			t.Fatalf("GetKeyAsHex failed: %v", err)
		}
		if hexStr != hex.EncodeToString(keyData) {
			t.Errorf("hex key mismatch: got %s, want %s", hexStr, hex.EncodeToString(keyData))
		}
	})

	t.Run("error when key file missing", func(t *testing.T) {
		em.mu.Lock()
		em.volumes["nokey"] = &EncryptedVolume{
			ContainerID: "nokey",
			KeyPath:     filepath.Join(tmpDir, "keys", "nonexistent.key"),
		}
		em.mu.Unlock()

		_, err := em.GetKeyAsHex("nokey")
		if err == nil {
			t.Error("expected error when key file is missing")
		}
	})
}

func TestEncryptionManager_LoadExistingVolumes(t *testing.T) {
	tmpDir := t.TempDir()
	em, err := NewEncryptionManager(tmpDir)
	if err != nil {
		t.Fatalf("NewEncryptionManager failed: %v", err)
	}

	t.Run("handles empty volumes directory", func(t *testing.T) {
		err := em.LoadExistingVolumes()
		if err != nil {
			t.Errorf("LoadExistingVolumes should not fail on empty dir: %v", err)
		}
	})

	t.Run("loads volume with matching key file", func(t *testing.T) {
		// Create a .img file and corresponding .key file
		imgPath := filepath.Join(tmpDir, "volumes", "loaded-container.img")
		keyPath := filepath.Join(tmpDir, "keys", "loaded-container.key")

		if err := os.WriteFile(imgPath, []byte("dummy"), 0644); err != nil {
			t.Fatalf("failed to write img file: %v", err)
		}
		if err := os.WriteFile(keyPath, []byte("dummykey12345678dummykey12345678"), 0400); err != nil {
			t.Fatalf("failed to write key file: %v", err)
		}

		// Reset volumes map
		em.mu.Lock()
		em.volumes = make(map[string]*EncryptedVolume)
		em.mu.Unlock()

		err := em.LoadExistingVolumes()
		if err != nil {
			t.Fatalf("LoadExistingVolumes failed: %v", err)
		}

		vol, exists := em.GetVolume("loaded-container")
		if !exists {
			t.Fatal("expected to find loaded-container volume")
		}
		if vol.ContainerID != "loaded-container" {
			t.Errorf("ContainerID = %q, want %q", vol.ContainerID, "loaded-container")
		}
		if vol.MapperName != "moltbunker-loaded-container" {
			t.Errorf("MapperName = %q, want %q", vol.MapperName, "moltbunker-loaded-container")
		}
	})

	t.Run("skips img files without key", func(t *testing.T) {
		imgPath := filepath.Join(tmpDir, "volumes", "nokey-container.img")
		if err := os.WriteFile(imgPath, []byte("dummy"), 0644); err != nil {
			t.Fatalf("failed to write img file: %v", err)
		}

		em.mu.Lock()
		em.volumes = make(map[string]*EncryptedVolume)
		em.mu.Unlock()

		err := em.LoadExistingVolumes()
		if err != nil {
			t.Fatalf("LoadExistingVolumes failed: %v", err)
		}

		_, exists := em.GetVolume("nokey-container")
		if exists {
			t.Error("should not load volume without matching key file")
		}
	})

	t.Run("skips non-img files", func(t *testing.T) {
		txtPath := filepath.Join(tmpDir, "volumes", "notanimage.txt")
		if err := os.WriteFile(txtPath, []byte("dummy"), 0644); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}

		em.mu.Lock()
		em.volumes = make(map[string]*EncryptedVolume)
		em.mu.Unlock()

		err := em.LoadExistingVolumes()
		if err != nil {
			t.Fatalf("LoadExistingVolumes failed: %v", err)
		}

		_, exists := em.GetVolume("notanimage")
		if exists {
			t.Error("should not load non-img file as volume")
		}
	})
}

func TestEncryptionManager_OpenVolume_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	em, err := NewEncryptionManager(tmpDir)
	if err != nil {
		t.Fatalf("NewEncryptionManager failed: %v", err)
	}

	err = em.OpenVolume("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent volume")
	}
	if !strings.Contains(err.Error(), "volume not found") {
		t.Errorf("expected 'volume not found' error, got: %v", err)
	}
}

func TestEncryptionManager_FormatVolume_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	em, err := NewEncryptionManager(tmpDir)
	if err != nil {
		t.Fatalf("NewEncryptionManager failed: %v", err)
	}

	err = em.FormatVolume("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent volume")
	}
	if !strings.Contains(err.Error(), "volume not found") {
		t.Errorf("expected 'volume not found' error, got: %v", err)
	}
}

func TestEncryptionManager_MountVolume_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	em, err := NewEncryptionManager(tmpDir)
	if err != nil {
		t.Fatalf("NewEncryptionManager failed: %v", err)
	}

	_, err = em.MountVolume("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent volume")
	}
	if !strings.Contains(err.Error(), "volume not found") {
		t.Errorf("expected 'volume not found' error, got: %v", err)
	}
}

func TestEncryptionManager_UnmountVolume_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	em, err := NewEncryptionManager(tmpDir)
	if err != nil {
		t.Fatalf("NewEncryptionManager failed: %v", err)
	}

	err = em.UnmountVolume("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent volume")
	}
	if !strings.Contains(err.Error(), "volume not found") {
		t.Errorf("expected 'volume not found' error, got: %v", err)
	}
}

func TestEncryptionManager_UnmountVolume_NotMounted(t *testing.T) {
	tmpDir := t.TempDir()
	em, err := NewEncryptionManager(tmpDir)
	if err != nil {
		t.Fatalf("NewEncryptionManager failed: %v", err)
	}

	em.mu.Lock()
	em.volumes["unmounted"] = &EncryptedVolume{
		ContainerID: "unmounted",
		Mounted:     false,
	}
	em.mu.Unlock()

	err = em.UnmountVolume("unmounted")
	if err != nil {
		t.Errorf("UnmountVolume should succeed for unmounted volume: %v", err)
	}
}

func TestEncryptionManager_MountVolume_AlreadyMounted(t *testing.T) {
	tmpDir := t.TempDir()
	em, err := NewEncryptionManager(tmpDir)
	if err != nil {
		t.Fatalf("NewEncryptionManager failed: %v", err)
	}

	mountPath := filepath.Join(tmpDir, "mounts", "already-mounted")
	em.mu.Lock()
	em.volumes["already-mounted"] = &EncryptedVolume{
		ContainerID: "already-mounted",
		Mounted:     true,
		MountPath:   mountPath,
	}
	em.mu.Unlock()

	path, err := em.MountVolume("already-mounted")
	if err != nil {
		t.Errorf("MountVolume should succeed for already-mounted: %v", err)
	}
	if path != mountPath {
		t.Errorf("mount path = %q, want %q", path, mountPath)
	}
}

func TestEncryptionManager_CloseVolume_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	em, err := NewEncryptionManager(tmpDir)
	if err != nil {
		t.Fatalf("NewEncryptionManager failed: %v", err)
	}

	err = em.CloseVolume("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent volume")
	}
	if !strings.Contains(err.Error(), "volume not found") {
		t.Errorf("expected 'volume not found' error, got: %v", err)
	}
}

func TestEncryptionManager_DeleteEncryptedVolume_NotTracked(t *testing.T) {
	tmpDir := t.TempDir()
	em, err := NewEncryptionManager(tmpDir)
	if err != nil {
		t.Fatalf("NewEncryptionManager failed: %v", err)
	}

	// DeleteEncryptedVolume should not error for non-tracked containers
	err = em.DeleteEncryptedVolume("non-tracked")
	if err != nil {
		t.Errorf("DeleteEncryptedVolume should succeed for non-tracked: %v", err)
	}
}

func TestEncryptionManager_CreateEncryptedVolume_Duplicate(t *testing.T) {
	tmpDir := t.TempDir()
	em, err := NewEncryptionManager(tmpDir)
	if err != nil {
		t.Fatalf("NewEncryptionManager failed: %v", err)
	}

	em.mu.Lock()
	em.volumes["dup-container"] = &EncryptedVolume{
		ContainerID: "dup-container",
	}
	em.mu.Unlock()

	key := make([]byte, 32)
	_, err = em.CreateEncryptedVolume("dup-container", 1, key)
	if err == nil {
		t.Error("expected error for duplicate volume")
	}
	if !strings.Contains(err.Error(), "volume already exists") {
		t.Errorf("expected 'volume already exists' error, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// EncryptedVolume struct tests
// ---------------------------------------------------------------------------

func TestEncryptedVolumeStruct(t *testing.T) {
	vol := EncryptedVolume{
		ContainerID: "test-1",
		VolumePath:  "/data/volumes/test-1.img",
		MapperName:  "moltbunker-test-1",
		MapperPath:  "/dev/mapper/moltbunker-test-1",
		MountPath:   "/data/mounts/test-1",
		SizeGB:      10,
		Mounted:     true,
		KeyPath:     "/data/keys/test-1.key",
	}

	if vol.ContainerID != "test-1" {
		t.Errorf("ContainerID = %q, want %q", vol.ContainerID, "test-1")
	}
	if vol.SizeGB != 10 {
		t.Errorf("SizeGB = %d, want 10", vol.SizeGB)
	}
	if !vol.Mounted {
		t.Error("Mounted should be true")
	}
}

// ---------------------------------------------------------------------------
// CgroupManager tests
// ---------------------------------------------------------------------------

func TestNewCgroupManager(t *testing.T) {
	t.Run("uses default root when empty", func(t *testing.T) {
		cm := NewCgroupManager("")
		if cm == nil {
			t.Fatal("expected non-nil CgroupManager")
		}
		if cm.cgroupRoot != "/sys/fs/cgroup" {
			t.Errorf("cgroupRoot = %q, want %q", cm.cgroupRoot, "/sys/fs/cgroup")
		}
	})

	t.Run("uses custom root", func(t *testing.T) {
		cm := NewCgroupManager("/custom/cgroup")
		if cm.cgroupRoot != "/custom/cgroup" {
			t.Errorf("cgroupRoot = %q, want %q", cm.cgroupRoot, "/custom/cgroup")
		}
	})
}

func TestCgroupManager_CreateAndDeleteCgroup(t *testing.T) {
	tmpDir := t.TempDir()
	cm := NewCgroupManager(tmpDir)

	containerID := "cgroup-test"
	err := cm.CreateCgroup(containerID)
	if err != nil {
		t.Fatalf("CreateCgroup failed: %v", err)
	}

	cgroupPath := filepath.Join(tmpDir, "moltbunker", containerID)
	info, err := os.Stat(cgroupPath)
	if err != nil {
		t.Fatalf("cgroup directory should exist: %v", err)
	}
	if !info.IsDir() {
		t.Error("cgroup should be a directory")
	}

	// Delete
	err = cm.DeleteCgroup(containerID)
	if err != nil {
		t.Fatalf("DeleteCgroup failed: %v", err)
	}

	_, err = os.Stat(cgroupPath)
	if !os.IsNotExist(err) {
		t.Error("cgroup directory should not exist after deletion")
	}
}

func TestCgroupManager_SetCPULimit(t *testing.T) {
	tmpDir := t.TempDir()
	cm := NewCgroupManager(tmpDir)

	containerID := "cpu-test"
	_ = cm.CreateCgroup(containerID)

	err := cm.SetCPULimit(containerID, 100000, 100000)
	if err != nil {
		t.Fatalf("SetCPULimit failed: %v", err)
	}

	// Verify content
	quotaPath := filepath.Join(tmpDir, "moltbunker", containerID, "cpu.max")
	data, err := os.ReadFile(quotaPath)
	if err != nil {
		t.Fatalf("failed to read cpu.max: %v", err)
	}
	if string(data) != "100000 100000" {
		t.Errorf("cpu.max = %q, want %q", string(data), "100000 100000")
	}
}

func TestCgroupManager_SetMemoryLimit(t *testing.T) {
	tmpDir := t.TempDir()
	cm := NewCgroupManager(tmpDir)

	containerID := "mem-test"
	_ = cm.CreateCgroup(containerID)

	err := cm.SetMemoryLimit(containerID, 1073741824)
	if err != nil {
		t.Fatalf("SetMemoryLimit failed: %v", err)
	}

	memPath := filepath.Join(tmpDir, "moltbunker", containerID, "memory.max")
	data, err := os.ReadFile(memPath)
	if err != nil {
		t.Fatalf("failed to read memory.max: %v", err)
	}
	if string(data) != "1073741824" {
		t.Errorf("memory.max = %q, want %q", string(data), "1073741824")
	}
}

func TestCgroupManager_SetPIDLimit(t *testing.T) {
	tmpDir := t.TempDir()
	cm := NewCgroupManager(tmpDir)

	containerID := "pid-test"
	_ = cm.CreateCgroup(containerID)

	err := cm.SetPIDLimit(containerID, 512)
	if err != nil {
		t.Fatalf("SetPIDLimit failed: %v", err)
	}

	pidPath := filepath.Join(tmpDir, "moltbunker", containerID, "pids.max")
	data, err := os.ReadFile(pidPath)
	if err != nil {
		t.Fatalf("failed to read pids.max: %v", err)
	}
	if string(data) != "512" {
		t.Errorf("pids.max = %q, want %q", string(data), "512")
	}
}

func TestCgroupManager_ApplyResourceLimits_AllFields(t *testing.T) {
	tmpDir := t.TempDir()
	cm := NewCgroupManager(tmpDir)

	resources := types.ResourceLimits{
		CPUQuota:    200000,
		CPUPeriod:   100000,
		MemoryLimit: 536870912,
		PIDLimit:    256,
	}

	err := cm.ApplyResourceLimits("apply-test", resources)
	if err != nil {
		t.Fatalf("ApplyResourceLimits failed: %v", err)
	}

	// Verify all files were written
	cgroupPath := filepath.Join(tmpDir, "moltbunker", "apply-test")

	data, err := os.ReadFile(filepath.Join(cgroupPath, "cpu.max"))
	if err != nil {
		t.Fatalf("failed to read cpu.max: %v", err)
	}
	if string(data) != "200000 100000" {
		t.Errorf("cpu.max = %q, want %q", string(data), "200000 100000")
	}

	data, err = os.ReadFile(filepath.Join(cgroupPath, "memory.max"))
	if err != nil {
		t.Fatalf("failed to read memory.max: %v", err)
	}
	if string(data) != "536870912" {
		t.Errorf("memory.max = %q", string(data))
	}

	data, err = os.ReadFile(filepath.Join(cgroupPath, "pids.max"))
	if err != nil {
		t.Fatalf("failed to read pids.max: %v", err)
	}
	if string(data) != "256" {
		t.Errorf("pids.max = %q, want %q", string(data), "256")
	}
}

func TestCgroupManager_DeleteCgroup_Nonexistent(t *testing.T) {
	tmpDir := t.TempDir()
	cm := NewCgroupManager(tmpDir)

	// Deleting nonexistent should not error (RemoveAll is tolerant)
	err := cm.DeleteCgroup("nonexistent")
	if err != nil {
		t.Errorf("DeleteCgroup should not fail for nonexistent: %v", err)
	}
}

// ---------------------------------------------------------------------------
// LogManager tests
// ---------------------------------------------------------------------------

func TestNewLogManager(t *testing.T) {
	t.Run("creates log directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		logsDir := filepath.Join(tmpDir, "logs")

		lm, err := NewLogManager(logsDir)
		if err != nil {
			t.Fatalf("NewLogManager failed: %v", err)
		}
		if lm == nil {
			t.Fatal("expected non-nil LogManager")
		}

		info, err := os.Stat(logsDir)
		if err != nil {
			t.Fatalf("logs directory should exist: %v", err)
		}
		if !info.IsDir() {
			t.Error("should be a directory")
		}
	})

	t.Run("initializes fields", func(t *testing.T) {
		tmpDir := t.TempDir()
		lm, err := NewLogManager(tmpDir)
		if err != nil {
			t.Fatalf("NewLogManager failed: %v", err)
		}

		if lm.logsDir != tmpDir {
			t.Errorf("logsDir = %q, want %q", lm.logsDir, tmpDir)
		}
		if lm.maxFileSize != 100*1024*1024 {
			t.Errorf("maxFileSize = %d, want %d", lm.maxFileSize, 100*1024*1024)
		}
		if lm.maxFiles != 5 {
			t.Errorf("maxFiles = %d, want 5", lm.maxFiles)
		}
		if lm.logFiles == nil {
			t.Error("logFiles map should be initialized")
		}
	})
}

func TestLogManager_CreateAndGetLog(t *testing.T) {
	tmpDir := t.TempDir()
	lm, err := NewLogManager(tmpDir)
	if err != nil {
		t.Fatalf("NewLogManager failed: %v", err)
	}

	containerID := "log-test"
	cl, err := lm.CreateLog(containerID)
	if err != nil {
		t.Fatalf("CreateLog failed: %v", err)
	}
	if cl == nil {
		t.Fatal("expected non-nil ContainerLog")
	}
	if cl.ContainerID != containerID {
		t.Errorf("ContainerID = %q, want %q", cl.ContainerID, containerID)
	}
	if cl.StdoutFile == nil {
		t.Error("StdoutFile should be set")
	}
	if cl.StderrFile == nil {
		t.Error("StderrFile should be set")
	}

	// Verify GetLog
	fetched, exists := lm.GetLog(containerID)
	if !exists {
		t.Error("GetLog should find the log")
	}
	if fetched != cl {
		t.Error("GetLog should return the same log")
	}

	// Nonexistent
	_, exists = lm.GetLog("nonexistent")
	if exists {
		t.Error("GetLog should not find nonexistent log")
	}
}

func TestLogManager_CloseLog(t *testing.T) {
	tmpDir := t.TempDir()
	lm, err := NewLogManager(tmpDir)
	if err != nil {
		t.Fatalf("NewLogManager failed: %v", err)
	}

	containerID := "close-test"
	_, err = lm.CreateLog(containerID)
	if err != nil {
		t.Fatalf("CreateLog failed: %v", err)
	}

	err = lm.CloseLog(containerID)
	if err != nil {
		t.Fatalf("CloseLog failed: %v", err)
	}

	// Should be removed from map
	_, exists := lm.GetLog(containerID)
	if exists {
		t.Error("log should not exist after close")
	}

	// Closing again should be safe
	err = lm.CloseLog(containerID)
	if err != nil {
		t.Errorf("CloseLog should be idempotent: %v", err)
	}
}

func TestLogManager_DeleteLog(t *testing.T) {
	tmpDir := t.TempDir()
	lm, err := NewLogManager(tmpDir)
	if err != nil {
		t.Fatalf("NewLogManager failed: %v", err)
	}

	containerID := "delete-test"
	_, err = lm.CreateLog(containerID)
	if err != nil {
		t.Fatalf("CreateLog failed: %v", err)
	}

	err = lm.DeleteLog(containerID)
	if err != nil {
		t.Fatalf("DeleteLog failed: %v", err)
	}

	// Directory should be gone
	containerDir := filepath.Join(tmpDir, containerID)
	_, err = os.Stat(containerDir)
	if !os.IsNotExist(err) {
		t.Error("container log directory should be removed")
	}
}

func TestContainerLog_WriteStdout(t *testing.T) {
	tmpDir := t.TempDir()
	lm, err := NewLogManager(tmpDir)
	if err != nil {
		t.Fatalf("NewLogManager failed: %v", err)
	}

	cl, err := lm.CreateLog("write-test")
	if err != nil {
		t.Fatalf("CreateLog failed: %v", err)
	}

	n, err := cl.WriteStdout([]byte("hello world"))
	if err != nil {
		t.Fatalf("WriteStdout failed: %v", err)
	}
	if n == 0 {
		t.Error("expected bytes written > 0")
	}

	// Verify file content
	data, err := os.ReadFile(cl.StdoutPath)
	if err != nil {
		t.Fatalf("failed to read stdout log: %v", err)
	}
	if !strings.Contains(string(data), "hello world") {
		t.Error("stdout log should contain 'hello world'")
	}
}

func TestContainerLog_WriteStderr(t *testing.T) {
	tmpDir := t.TempDir()
	lm, err := NewLogManager(tmpDir)
	if err != nil {
		t.Fatalf("NewLogManager failed: %v", err)
	}

	cl, err := lm.CreateLog("write-err-test")
	if err != nil {
		t.Fatalf("CreateLog failed: %v", err)
	}

	n, err := cl.WriteStderr([]byte("error message"))
	if err != nil {
		t.Fatalf("WriteStderr failed: %v", err)
	}
	if n == 0 {
		t.Error("expected bytes written > 0")
	}

	data, err := os.ReadFile(cl.StderrPath)
	if err != nil {
		t.Fatalf("failed to read stderr log: %v", err)
	}
	if !strings.Contains(string(data), "error message") {
		t.Error("stderr log should contain 'error message'")
	}
}

func TestContainerLog_WriteStdout_NilFile(t *testing.T) {
	cl := &ContainerLog{
		ContainerID: "nil-file",
		StdoutFile:  nil,
	}
	_, err := cl.WriteStdout([]byte("test"))
	if err == nil {
		t.Error("expected error when stdout file is nil")
	}
}

func TestContainerLog_WriteStderr_NilFile(t *testing.T) {
	cl := &ContainerLog{
		ContainerID: "nil-file",
		StderrFile:  nil,
	}
	_, err := cl.WriteStderr([]byte("test"))
	if err == nil {
		t.Error("expected error when stderr file is nil")
	}
}

func TestContainerLog_Writers(t *testing.T) {
	tmpDir := t.TempDir()
	lm, err := NewLogManager(tmpDir)
	if err != nil {
		t.Fatalf("NewLogManager failed: %v", err)
	}

	cl, err := lm.CreateLog("writer-test")
	if err != nil {
		t.Fatalf("CreateLog failed: %v", err)
	}

	// Test StdoutWriter
	stdoutW := cl.StdoutWriter()
	if stdoutW == nil {
		t.Fatal("StdoutWriter should not be nil")
	}
	n, err := stdoutW.Write([]byte("stdout via writer"))
	if err != nil {
		t.Fatalf("StdoutWriter.Write failed: %v", err)
	}
	if n == 0 {
		t.Error("expected bytes written > 0")
	}

	// Test StderrWriter
	stderrW := cl.StderrWriter()
	if stderrW == nil {
		t.Fatal("StderrWriter should not be nil")
	}
	n, err = stderrW.Write([]byte("stderr via writer"))
	if err != nil {
		t.Fatalf("StderrWriter.Write failed: %v", err)
	}
	if n == 0 {
		t.Error("expected bytes written > 0")
	}
}

func TestLogManager_GetLogPaths(t *testing.T) {
	tmpDir := t.TempDir()
	lm, err := NewLogManager(tmpDir)
	if err != nil {
		t.Fatalf("NewLogManager failed: %v", err)
	}

	stdout, stderr := lm.GetLogPaths("test-container")
	expectedStdout := filepath.Join(tmpDir, "test-container", "stdout.log")
	expectedStderr := filepath.Join(tmpDir, "test-container", "stderr.log")

	if stdout != expectedStdout {
		t.Errorf("stdout path = %q, want %q", stdout, expectedStdout)
	}
	if stderr != expectedStderr {
		t.Errorf("stderr path = %q, want %q", stderr, expectedStderr)
	}
}

func TestEmptyReader(t *testing.T) {
	r := emptyReader{}
	buf := make([]byte, 10)
	n, err := r.Read(buf)
	if n != 0 {
		t.Errorf("expected 0 bytes, got %d", n)
	}
	if err != io.EOF {
		t.Errorf("expected io.EOF, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// ManagedContainerInfo struct tests
// ---------------------------------------------------------------------------

func TestManagedContainerInfoStruct(t *testing.T) {
	now := time.Now()
	info := ManagedContainerInfo{
		ID:              "container-1",
		Image:           "nginx:latest",
		Status:          types.ContainerStatusRunning,
		CreatedAt:       now.Add(-1 * time.Hour),
		StartedAt:       now,
		Resources:       types.ResourceLimits{CPUQuota: 100000, MemoryLimit: 512 * 1024 * 1024},
		EncryptedVolume: "/dev/mapper/moltbunker-1",
		OnionAddress:    "abc123.onion",
		DeploymentID:    "deploy-1",
		ExecDisabled:    true,
		AttachDisabled:  true,
		ShellDisabled:   false,
	}

	if info.ID != "container-1" {
		t.Errorf("ID = %q, want %q", info.ID, "container-1")
	}
	if info.Status != types.ContainerStatusRunning {
		t.Errorf("Status = %v, want Running", info.Status)
	}
	if !info.ExecDisabled {
		t.Error("ExecDisabled should be true")
	}
	if !info.AttachDisabled {
		t.Error("AttachDisabled should be true")
	}
	if info.ShellDisabled {
		t.Error("ShellDisabled should be false")
	}
	if info.DeploymentID != "deploy-1" {
		t.Errorf("DeploymentID = %q, want %q", info.DeploymentID, "deploy-1")
	}
}

// ---------------------------------------------------------------------------
// SecureContainerConfig struct tests
// ---------------------------------------------------------------------------

func TestSecureContainerConfigStruct(t *testing.T) {
	config := SecureContainerConfig{
		ID:       "secure-1",
		ImageRef: "docker.io/library/nginx:latest",
		Resources: types.ResourceLimits{
			CPUQuota:    200000,
			CPUPeriod:   100000,
			MemoryLimit: 1073741824,
			PIDLimit:    512,
		},
		SecurityProfile: types.DefaultContainerSecurityProfile(),
		DeploymentID:    "deploy-secure-1",
		RequesterPubKey: []byte("pubkey-data"),
		Environment:     map[string]string{"ENV_VAR": "value"},
		Command:         []string{"/bin/app"},
		Args:            []string{"--port", "8080"},
	}

	if config.ID != "secure-1" {
		t.Errorf("ID = %q, want %q", config.ID, "secure-1")
	}
	if config.ImageRef != "docker.io/library/nginx:latest" {
		t.Errorf("ImageRef = %q", config.ImageRef)
	}
	if config.SecurityProfile == nil {
		t.Error("SecurityProfile should not be nil")
	}
	if len(config.RequesterPubKey) == 0 {
		t.Error("RequesterPubKey should not be empty")
	}
	if config.Environment["ENV_VAR"] != "value" {
		t.Error("Environment should contain ENV_VAR")
	}
	if len(config.Command) != 1 || config.Command[0] != "/bin/app" {
		t.Error("Command mismatch")
	}
	if len(config.Args) != 2 {
		t.Error("Args should have 2 elements")
	}
}

// ---------------------------------------------------------------------------
// SecurityEnforcer OCI spec opts additional tests
// ---------------------------------------------------------------------------

func TestBuildOCISpecOpts_AllFeatures(t *testing.T) {
	profile := &types.ContainerSecurityProfile{
		DropAllCapabilities: true,
		AddCapabilities:     []string{"NET_BIND_SERVICE"},
		ReadOnlyRoot:        true,
		NoNewPrivileges:     true,
		MaskPaths:           []string{"/proc/kcore"},
		ReadOnlyPaths:       []string{"/proc/sys"},
		SeccompProfile:      "strict",
		BlockedSyscalls:     []string{"mount", "ptrace"},
		AllowedSyscalls:     []string{"read", "write"},
		AppArmorProfile:     "moltbunker-default",
		SELinuxLabel:        "system_u:system_r:container_t:s0",
		UserNamespace:       true,
		Ulimits:             types.UlimitConfig{NoFile: 1024, NProc: 512},
	}
	se := NewSecurityEnforcer(profile)
	opts := se.BuildOCISpecOpts()

	// All features enabled should produce multiple options
	// DropAllCapabilities -> 1
	// ReadOnlyRoot -> 1
	// NoNewPrivileges -> 1
	// MaskPaths -> 1
	// ReadOnlyPaths -> 1
	// SeccompProfile -> 1
	// AppArmorProfile -> 1 (only if loaded on Linux, skipped otherwise)
	// SELinuxLabel -> 1
	// Ulimits -> 1
	// UserNamespace -> 1
	// = at least 9 options (10 if AppArmor profile is loaded on Linux)
	if len(opts) < 9 {
		t.Errorf("expected at least 9 OCI spec opts, got %d", len(opts))
	}
}

func TestBuildOCISpecOpts_EmptyProfile(t *testing.T) {
	profile := &types.ContainerSecurityProfile{}
	se := NewSecurityEnforcer(profile)
	opts := se.BuildOCISpecOpts()

	// Empty profile should still produce ulimits option
	if len(opts) < 1 {
		t.Errorf("expected at least 1 OCI spec opt (ulimits), got %d", len(opts))
	}
}

func TestSecurityEnforcer_GetProfile(t *testing.T) {
	profile := &types.ContainerSecurityProfile{
		DisableExec: true,
	}
	se := NewSecurityEnforcer(profile)
	got := se.GetProfile()
	if got != profile {
		t.Error("GetProfile should return the same profile")
	}
}

// ---------------------------------------------------------------------------
// readLastLines pure function test
// ---------------------------------------------------------------------------

func TestReadLastLines(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.log")

	t.Run("reads last N lines from file", func(t *testing.T) {
		content := "line1\nline2\nline3\nline4\nline5\n"
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}

		f, err := os.Open(filePath)
		if err != nil {
			t.Fatalf("failed to open file: %v", err)
		}
		defer f.Close()

		lines, err := readLastLines(f, 3)
		if err != nil {
			t.Fatalf("readLastLines failed: %v", err)
		}
		if len(lines) != 3 {
			t.Errorf("expected 3 lines, got %d: %v", len(lines), lines)
		}
	})

	t.Run("handles empty file", func(t *testing.T) {
		emptyPath := filepath.Join(tmpDir, "empty.log")
		if err := os.WriteFile(emptyPath, []byte(""), 0644); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}

		f, err := os.Open(emptyPath)
		if err != nil {
			t.Fatalf("failed to open file: %v", err)
		}
		defer f.Close()

		lines, err := readLastLines(f, 5)
		if err != nil {
			t.Fatalf("readLastLines failed: %v", err)
		}
		if lines != nil {
			t.Errorf("expected nil for empty file, got %v", lines)
		}
	})

	t.Run("handles fewer lines than requested", func(t *testing.T) {
		smallPath := filepath.Join(tmpDir, "small.log")
		if err := os.WriteFile(smallPath, []byte("only one line\n"), 0644); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}

		f, err := os.Open(smallPath)
		if err != nil {
			t.Fatalf("failed to open file: %v", err)
		}
		defer f.Close()

		lines, err := readLastLines(f, 10)
		if err != nil {
			t.Fatalf("readLastLines failed: %v", err)
		}
		if len(lines) > 10 {
			t.Errorf("should not return more than requested: got %d", len(lines))
		}
	})
}

// ---------------------------------------------------------------------------
// LogManager ReadLogs test
// ---------------------------------------------------------------------------

func TestLogManager_ReadLogs_NoLogs(t *testing.T) {
	tmpDir := t.TempDir()
	lm, err := NewLogManager(tmpDir)
	if err != nil {
		t.Fatalf("NewLogManager failed: %v", err)
	}

	reader, err := lm.ReadLogs(nil, "nonexistent", false, 0)
	if err != nil {
		t.Fatalf("ReadLogs failed: %v", err)
	}
	defer reader.Close()

	// Should return empty reader
	buf := make([]byte, 100)
	n, readErr := reader.Read(buf)
	if n != 0 && readErr != io.EOF {
		t.Errorf("expected empty read, got %d bytes, err=%v", n, readErr)
	}
}

// ---------------------------------------------------------------------------
// SecurityProfileError tests (additional)
// ---------------------------------------------------------------------------

func TestSecurityProfileError_Fields(t *testing.T) {
	tests := []struct {
		name      string
		err       *SecurityProfileError
		operation string
		reason    string
	}{
		{
			name:      "exec disabled",
			err:       ErrExecDisabled,
			operation: "exec",
			reason:    "container exec is disabled by security policy",
		},
		{
			name:      "attach disabled",
			err:       ErrAttachDisabled,
			operation: "attach",
			reason:    "container attach is disabled by security policy",
		},
		{
			name:      "shell disabled",
			err:       ErrShellDisabled,
			operation: "shell",
			reason:    "shell access is disabled by security policy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Operation != tt.operation {
				t.Errorf("Operation = %q, want %q", tt.err.Operation, tt.operation)
			}
			if tt.err.Reason != tt.reason {
				t.Errorf("Reason = %q, want %q", tt.err.Reason, tt.reason)
			}
			// Verify Error() format
			expected := "security policy violation: " + tt.operation + " - " + tt.reason
			if tt.err.Error() != expected {
				t.Errorf("Error() = %q, want %q", tt.err.Error(), expected)
			}
		})
	}
}

func TestValidateExecCommand_ShellVariants(t *testing.T) {
	se := NewSecurityEnforcer(&types.ContainerSecurityProfile{
		DisableExec:  false,
		DisableShell: true,
	})

	// All these should be blocked when shell is disabled
	blockedCommands := [][]string{
		{"ash"},
		{"dash"},
		{"csh"},
		{"tcsh"},
		{"ksh"},
		{"fish"},
		{"/bin/ash"},
		{"/bin/dash"},
		{"/bin/csh"},
		{"/bin/tcsh"},
		{"/bin/ksh"},
		{"/bin/fish"},
		{"/usr/bin/sh"},
		{"/usr/bin/bash"},
		{"/usr/bin/zsh"},
		{"/usr/bin/fish"},
	}

	for _, cmd := range blockedCommands {
		err := se.ValidateExecCommand(cmd)
		if err == nil {
			t.Errorf("expected error for shell command %v", cmd)
		}
		if err != ErrShellDisabled {
			t.Errorf("expected ErrShellDisabled for %v, got %v", cmd, err)
		}
	}

	// Non-shell commands should be allowed
	allowedCommands := [][]string{
		{"ls", "-la"},
		{"cat", "/etc/hostname"},
		{"python3", "script.py"},
		{"node", "app.js"},
	}

	for _, cmd := range allowedCommands {
		err := se.ValidateExecCommand(cmd)
		if err != nil {
			t.Errorf("unexpected error for non-shell command %v: %v", cmd, err)
		}
	}
}
