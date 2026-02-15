package snapshot

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"
)

// --- Config tests ---

func TestDefaultSnapshotConfig(t *testing.T) {
	cfg := DefaultSnapshotConfig()
	if cfg == nil {
		t.Fatal("DefaultSnapshotConfig returned nil")
	}
	if cfg.MaxSnapshots != 10 {
		t.Errorf("expected MaxSnapshots 10, got %d", cfg.MaxSnapshots)
	}
	if cfg.MaxTotalSize != 10*1024*1024*1024 {
		t.Errorf("expected MaxTotalSize 10GB, got %d", cfg.MaxTotalSize)
	}
	if cfg.CompressionLevel != 6 {
		t.Errorf("expected CompressionLevel 6, got %d", cfg.CompressionLevel)
	}
	if !cfg.EncryptionEnabled {
		t.Error("expected EncryptionEnabled true by default")
	}
	if cfg.RetentionDays != 7 {
		t.Errorf("expected RetentionDays 7, got %d", cfg.RetentionDays)
	}
}

// --- Manager tests ---

func TestNewManager(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("nil config uses defaults", func(t *testing.T) {
		m, err := NewManager(&SnapshotConfig{
			StoragePath:       tmpDir,
			MaxSnapshots:      10,
			MaxTotalSize:      10 * 1024 * 1024,
			CompressionLevel:  0,
			EncryptionEnabled: false,
			RetentionDays:     7,
		})
		if err != nil {
			t.Fatalf("NewManager() error: %v", err)
		}
		if m == nil {
			t.Fatal("NewManager returned nil")
		}
	})

	t.Run("no encryption", func(t *testing.T) {
		m, err := NewManager(&SnapshotConfig{
			StoragePath:       t.TempDir(),
			MaxSnapshots:      5,
			CompressionLevel:  0,
			EncryptionEnabled: false,
			RetentionDays:     7,
		})
		if err != nil {
			t.Fatalf("NewManager() error: %v", err)
		}
		if m.encryptionKey != nil {
			t.Error("expected no encryption key when encryption disabled")
		}
	})
}

func TestCreateAndGetSnapshot(t *testing.T) {
	tmpDir := t.TempDir()
	m, err := NewManager(&SnapshotConfig{
		StoragePath:       tmpDir,
		MaxSnapshots:      10,
		MaxTotalSize:      10 * 1024 * 1024,
		CompressionLevel:  0,
		EncryptionEnabled: false,
		RetentionDays:     7,
	})
	if err != nil {
		t.Fatalf("NewManager() error: %v", err)
	}

	data := []byte("container state data for testing")
	metadata := map[string]string{"env": "test"}

	snap, err := m.CreateSnapshot("container-1", data, SnapshotTypeFull, metadata)
	if err != nil {
		t.Fatalf("CreateSnapshot() error: %v", err)
	}

	if snap.ContainerID != "container-1" {
		t.Errorf("expected ContainerID 'container-1', got %s", snap.ContainerID)
	}
	if snap.Size != int64(len(data)) {
		t.Errorf("expected Size %d, got %d", len(data), snap.Size)
	}
	if snap.Type != SnapshotTypeFull {
		t.Errorf("expected Type 'full', got %s", snap.Type)
	}
	if snap.Metadata["env"] != "test" {
		t.Error("metadata not preserved")
	}

	// Verify checksum
	hash := sha256.Sum256(data)
	expectedChecksum := hex.EncodeToString(hash[:])
	if snap.Checksum != expectedChecksum {
		t.Error("checksum mismatch")
	}

	// Get by ID
	retrieved, err := m.GetSnapshot(snap.ID)
	if err != nil {
		t.Fatalf("GetSnapshot() error: %v", err)
	}
	if retrieved.ID != snap.ID {
		t.Error("retrieved snapshot has wrong ID")
	}
}

func TestGetSnapshotData(t *testing.T) {
	tmpDir := t.TempDir()
	m, err := NewManager(&SnapshotConfig{
		StoragePath:       tmpDir,
		MaxSnapshots:      10,
		MaxTotalSize:      10 * 1024 * 1024,
		CompressionLevel:  0,
		EncryptionEnabled: false,
		RetentionDays:     7,
	})
	if err != nil {
		t.Fatal(err)
	}

	data := []byte("my important container state")
	snap, err := m.CreateSnapshot("c1", data, SnapshotTypeFull, nil)
	if err != nil {
		t.Fatal(err)
	}

	retrieved, err := m.GetSnapshotData(snap.ID)
	if err != nil {
		t.Fatalf("GetSnapshotData() error: %v", err)
	}
	if !bytes.Equal(retrieved, data) {
		t.Errorf("data mismatch: got %q, want %q", retrieved, data)
	}
}

func TestGetSnapshotDataWithCompression(t *testing.T) {
	tmpDir := t.TempDir()
	m, err := NewManager(&SnapshotConfig{
		StoragePath:       tmpDir,
		MaxSnapshots:      10,
		MaxTotalSize:      10 * 1024 * 1024,
		CompressionLevel:  6,
		EncryptionEnabled: false,
		RetentionDays:     7,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Create data that compresses well
	data := bytes.Repeat([]byte("abcdefghijklmnop"), 1000)
	snap, err := m.CreateSnapshot("c1", data, SnapshotTypeFull, nil)
	if err != nil {
		t.Fatal(err)
	}

	if !snap.Compressed {
		t.Error("expected snapshot to be compressed")
	}
	if snap.StoredSize >= snap.Size {
		t.Errorf("expected compressed stored size < original size, got stored=%d, original=%d",
			snap.StoredSize, snap.Size)
	}

	retrieved, err := m.GetSnapshotData(snap.ID)
	if err != nil {
		t.Fatalf("GetSnapshotData() error: %v", err)
	}
	if !bytes.Equal(retrieved, data) {
		t.Error("data mismatch after decompression")
	}
}

func TestGetSnapshotNotFound(t *testing.T) {
	m, err := NewManager(&SnapshotConfig{
		CompressionLevel:  0,
		EncryptionEnabled: false,
		RetentionDays:     7,
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = m.GetSnapshot("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent snapshot")
	}
}

func TestListSnapshots(t *testing.T) {
	tmpDir := t.TempDir()
	m, err := NewManager(&SnapshotConfig{
		StoragePath:       tmpDir,
		MaxSnapshots:      10,
		MaxTotalSize:      10 * 1024 * 1024,
		CompressionLevel:  0,
		EncryptionEnabled: false,
		RetentionDays:     7,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Create 3 snapshots for container-1
	for i := 0; i < 3; i++ {
		_, err := m.CreateSnapshot("container-1", []byte("data"), SnapshotTypeFull, nil)
		if err != nil {
			t.Fatal(err)
		}
		// Small sleep to ensure different timestamps
		time.Sleep(time.Millisecond)
	}

	// Create 2 for container-2
	for i := 0; i < 2; i++ {
		_, err := m.CreateSnapshot("container-2", []byte("data"), SnapshotTypeFull, nil)
		if err != nil {
			t.Fatal(err)
		}
	}

	list1 := m.ListSnapshots("container-1")
	if len(list1) != 3 {
		t.Errorf("expected 3 snapshots for container-1, got %d", len(list1))
	}
	// Verify sorted newest first
	for i := 0; i < len(list1)-1; i++ {
		if list1[i].CreatedAt.Before(list1[i+1].CreatedAt) {
			t.Error("snapshots should be sorted newest first")
		}
	}

	list2 := m.ListSnapshots("container-2")
	if len(list2) != 2 {
		t.Errorf("expected 2 snapshots for container-2, got %d", len(list2))
	}

	// Nonexistent container
	list3 := m.ListSnapshots("nonexistent")
	if len(list3) != 0 {
		t.Errorf("expected 0 snapshots for nonexistent container, got %d", len(list3))
	}
}

func TestGetLatestSnapshot(t *testing.T) {
	tmpDir := t.TempDir()
	m, err := NewManager(&SnapshotConfig{
		StoragePath:       tmpDir,
		MaxSnapshots:      10,
		MaxTotalSize:      10 * 1024 * 1024,
		CompressionLevel:  0,
		EncryptionEnabled: false,
		RetentionDays:     7,
	})
	if err != nil {
		t.Fatal(err)
	}

	// No snapshots
	_, err = m.GetLatestSnapshot("container-1")
	if err == nil {
		t.Error("expected error when no snapshots exist")
	}

	// Create snapshots
	m.CreateSnapshot("container-1", []byte("first"), SnapshotTypeFull, nil)
	time.Sleep(time.Millisecond)
	last, err := m.CreateSnapshot("container-1", []byte("second"), SnapshotTypeFull, nil)
	if err != nil {
		t.Fatal(err)
	}

	latest, err := m.GetLatestSnapshot("container-1")
	if err != nil {
		t.Fatalf("GetLatestSnapshot() error: %v", err)
	}
	if latest.ID != last.ID {
		t.Errorf("expected latest snapshot to be %s, got %s", last.ID, latest.ID)
	}
}

func TestDeleteSnapshot(t *testing.T) {
	tmpDir := t.TempDir()
	m, err := NewManager(&SnapshotConfig{
		StoragePath:       tmpDir,
		MaxSnapshots:      10,
		MaxTotalSize:      10 * 1024 * 1024,
		CompressionLevel:  0,
		EncryptionEnabled: false,
		RetentionDays:     7,
	})
	if err != nil {
		t.Fatal(err)
	}

	snap, err := m.CreateSnapshot("c1", []byte("data to delete"), SnapshotTypeFull, nil)
	if err != nil {
		t.Fatal(err)
	}

	initialSize := m.totalSize

	err = m.DeleteSnapshot(snap.ID)
	if err != nil {
		t.Fatalf("DeleteSnapshot() error: %v", err)
	}

	// Verify deleted
	_, err = m.GetSnapshot(snap.ID)
	if err == nil {
		t.Error("expected error getting deleted snapshot")
	}

	// Verify total size decreased
	if m.totalSize >= initialSize {
		t.Error("total size should decrease after deletion")
	}

	// Delete nonexistent
	err = m.DeleteSnapshot("nonexistent")
	if err == nil {
		t.Error("expected error deleting nonexistent snapshot")
	}
}

func TestDeleteContainerSnapshots(t *testing.T) {
	tmpDir := t.TempDir()
	m, err := NewManager(&SnapshotConfig{
		StoragePath:       tmpDir,
		MaxSnapshots:      10,
		MaxTotalSize:      10 * 1024 * 1024,
		CompressionLevel:  0,
		EncryptionEnabled: false,
		RetentionDays:     7,
	})
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 3; i++ {
		m.CreateSnapshot("c1", []byte("data"), SnapshotTypeFull, nil)
	}
	m.CreateSnapshot("c2", []byte("other data"), SnapshotTypeFull, nil)

	err = m.DeleteContainerSnapshots("c1")
	if err != nil {
		t.Fatalf("DeleteContainerSnapshots() error: %v", err)
	}

	list := m.ListSnapshots("c1")
	if len(list) != 0 {
		t.Errorf("expected 0 snapshots for c1, got %d", len(list))
	}

	// c2 should still have its snapshot
	list2 := m.ListSnapshots("c2")
	if len(list2) != 1 {
		t.Errorf("expected 1 snapshot for c2, got %d", len(list2))
	}
}

func TestEnforcePerContainerLimit(t *testing.T) {
	tmpDir := t.TempDir()
	m, err := NewManager(&SnapshotConfig{
		StoragePath:       tmpDir,
		MaxSnapshots:      3, // Only allow 3
		MaxTotalSize:      10 * 1024 * 1024,
		CompressionLevel:  0,
		EncryptionEnabled: false,
		RetentionDays:     7,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Create 5 snapshots - should only keep 3
	for i := 0; i < 5; i++ {
		_, err := m.CreateSnapshot("c1", []byte("data"), SnapshotTypeFull, nil)
		if err != nil {
			t.Fatal(err)
		}
		time.Sleep(time.Millisecond)
	}

	list := m.ListSnapshots("c1")
	if len(list) != 3 {
		t.Errorf("expected 3 snapshots (limit), got %d", len(list))
	}
}

func TestCleanupExpired(t *testing.T) {
	tmpDir := t.TempDir()
	m, err := NewManager(&SnapshotConfig{
		StoragePath:       tmpDir,
		MaxSnapshots:      10,
		MaxTotalSize:      10 * 1024 * 1024,
		CompressionLevel:  0,
		EncryptionEnabled: false,
		RetentionDays:     1, // 1 day retention
	})
	if err != nil {
		t.Fatal(err)
	}

	// Create a snapshot and backdate it
	snap, err := m.CreateSnapshot("c1", []byte("old data"), SnapshotTypeFull, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Backdate the snapshot to 2 days ago
	m.mu.Lock()
	snap.CreatedAt = time.Now().AddDate(0, 0, -2)
	m.mu.Unlock()

	// Also create a recent snapshot
	m.CreateSnapshot("c1", []byte("new data"), SnapshotTypeFull, nil)

	removed := m.CleanupExpired()
	if removed != 1 {
		t.Errorf("expected 1 removed, got %d", removed)
	}

	list := m.ListSnapshots("c1")
	if len(list) != 1 {
		t.Errorf("expected 1 remaining snapshot, got %d", len(list))
	}
}

func TestGetStats(t *testing.T) {
	tmpDir := t.TempDir()
	m, err := NewManager(&SnapshotConfig{
		StoragePath:       tmpDir,
		MaxSnapshots:      10,
		MaxTotalSize:      10 * 1024 * 1024,
		CompressionLevel:  0,
		EncryptionEnabled: false,
		RetentionDays:     7,
	})
	if err != nil {
		t.Fatal(err)
	}

	m.CreateSnapshot("c1", []byte("data1"), SnapshotTypeFull, nil)
	m.CreateSnapshot("c2", []byte("data2"), SnapshotTypeFull, nil)

	stats := m.GetStats()
	if stats.TotalCount != 2 {
		t.Errorf("expected TotalCount 2, got %d", stats.TotalCount)
	}
	if stats.ContainerCount != 2 {
		t.Errorf("expected ContainerCount 2, got %d", stats.ContainerCount)
	}
	if stats.TotalSize <= 0 {
		t.Error("expected positive TotalSize")
	}
}

func TestGetDetailedStats(t *testing.T) {
	tmpDir := t.TempDir()
	m, err := NewManager(&SnapshotConfig{
		StoragePath:       tmpDir,
		MaxSnapshots:      10,
		MaxTotalSize:      10 * 1024 * 1024,
		CompressionLevel:  6,
		EncryptionEnabled: false,
		RetentionDays:     7,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Create data that compresses well
	data := bytes.Repeat([]byte("x"), 10000)
	m.CreateSnapshot("c1", data, SnapshotTypeFull, nil)

	stats := m.GetDetailedStats()
	if stats.TotalCount != 1 {
		t.Errorf("expected TotalCount 1, got %d", stats.TotalCount)
	}
	if stats.TotalOriginalSize != int64(len(data)) {
		t.Errorf("expected TotalOriginalSize %d, got %d", len(data), stats.TotalOriginalSize)
	}
	if stats.CompressionRatio <= 1.0 {
		t.Errorf("expected compression ratio > 1 for compressible data, got %f", stats.CompressionRatio)
	}
}

func TestListAllSnapshots(t *testing.T) {
	tmpDir := t.TempDir()
	m, err := NewManager(&SnapshotConfig{
		StoragePath:       tmpDir,
		MaxSnapshots:      10,
		MaxTotalSize:      10 * 1024 * 1024,
		CompressionLevel:  0,
		EncryptionEnabled: false,
		RetentionDays:     7,
	})
	if err != nil {
		t.Fatal(err)
	}

	m.CreateSnapshot("c1", []byte("data1"), SnapshotTypeFull, nil)
	time.Sleep(time.Millisecond)
	m.CreateSnapshot("c2", []byte("data2"), SnapshotTypeFull, nil)
	time.Sleep(time.Millisecond)
	m.CreateSnapshot("c1", []byte("data3"), SnapshotTypeFull, nil)

	all := m.ListAllSnapshots()
	if len(all) != 3 {
		t.Errorf("expected 3 total snapshots, got %d", len(all))
	}

	// Should be sorted newest first
	for i := 0; i < len(all)-1; i++ {
		if all[i].CreatedAt.Before(all[i+1].CreatedAt) {
			t.Error("should be sorted newest first")
		}
	}
}

// --- Block hashing and delta tests ---

func TestComputeBlockHashes(t *testing.T) {
	data := bytes.Repeat([]byte("a"), blockSize*3+100)
	hashes := computeBlockHashes(data)

	// Should have 4 blocks (3 full + 1 partial)
	if len(hashes) != 4 {
		t.Errorf("expected 4 blocks, got %d", len(hashes))
	}

	// Verify offsets
	for _, offset := range []int64{0, blockSize, blockSize * 2, blockSize * 3} {
		if _, ok := hashes[offset]; !ok {
			t.Errorf("expected hash at offset %d", offset)
		}
	}
}

func TestComputeDelta(t *testing.T) {
	// Create initial data (3 full blocks)
	oldData := bytes.Repeat([]byte("a"), blockSize*3)
	oldHashes := computeBlockHashes(oldData)

	// Change only the second block
	newData := make([]byte, len(oldData))
	copy(newData, oldData)
	for i := blockSize; i < blockSize*2; i++ {
		newData[i] = 'b'
	}

	deltaData, deltaBlocks := computeDelta(oldHashes, newData)

	// Only 1 block should be changed
	if len(deltaBlocks) != 1 {
		t.Errorf("expected 1 delta block, got %d", len(deltaBlocks))
	}
	if deltaBlocks[0].Offset != int64(blockSize) {
		t.Errorf("expected delta at offset %d, got %d", blockSize, deltaBlocks[0].Offset)
	}
	if len(deltaData) != blockSize {
		t.Errorf("expected delta data of size %d, got %d", blockSize, len(deltaData))
	}
}

func TestApplyDelta(t *testing.T) {
	baseData := []byte("AAAA BBBB CCCC DDDD")
	deltaData := []byte("XXXX")
	deltaBlocks := []DeltaBlock{
		{Offset: 5, Length: 4, Hash: ""},
	}

	result, err := applyDelta(baseData, deltaData, deltaBlocks)
	if err != nil {
		t.Fatalf("applyDelta() error: %v", err)
	}

	expected := "AAAA XXXX CCCC DDDD"
	if string(result) != expected {
		t.Errorf("applyDelta result = %q, want %q", string(result), expected)
	}
}

func TestApplyDeltaExtendData(t *testing.T) {
	baseData := []byte("short")
	deltaData := []byte("EXTRA")
	deltaBlocks := []DeltaBlock{
		{Offset: 10, Length: 5, Hash: ""},
	}

	result, err := applyDelta(baseData, deltaData, deltaBlocks)
	if err != nil {
		t.Fatalf("applyDelta() error: %v", err)
	}

	if len(result) != 15 {
		t.Errorf("expected result length 15, got %d", len(result))
	}
	if string(result[10:15]) != "EXTRA" {
		t.Errorf("expected 'EXTRA' at offset 10, got %q", string(result[10:15]))
	}
}

func TestApplyDeltaTooShort(t *testing.T) {
	baseData := []byte("base")
	deltaData := []byte("X") // Too short
	deltaBlocks := []DeltaBlock{
		{Offset: 0, Length: 10, Hash: ""}, // Needs 10 bytes but only 1 available
	}

	_, err := applyDelta(baseData, deltaData, deltaBlocks)
	if err == nil {
		t.Error("expected error for delta data too short")
	}
}

// --- Export/Import helper tests ---

func TestWriteReadInt32(t *testing.T) {
	tests := []int32{0, 1, 255, 65535, 1000000, -1}

	for _, val := range tests {
		var buf bytes.Buffer
		if err := writeInt32(&buf, val); err != nil {
			t.Fatalf("writeInt32(%d) error: %v", val, err)
		}

		read, err := readInt32(&buf)
		if err != nil {
			t.Fatalf("readInt32() error: %v", err)
		}
		if read != val {
			t.Errorf("readInt32() = %d, want %d", read, val)
		}
	}
}

func TestGetStoredSize(t *testing.T) {
	// With StoredSize set
	snap := &Snapshot{Size: 100, StoredSize: 50}
	if got := getStoredSize(snap); got != 50 {
		t.Errorf("expected 50, got %d", got)
	}

	// Without StoredSize (backwards compat)
	snap = &Snapshot{Size: 100, StoredSize: 0}
	if got := getStoredSize(snap); got != 100 {
		t.Errorf("expected 100 (fallback to Size), got %d", got)
	}
}

// --- Checkpoint config tests ---

func TestDefaultCheckpointConfig(t *testing.T) {
	cfg := DefaultCheckpointConfig()
	if cfg == nil {
		t.Fatal("DefaultCheckpointConfig returned nil")
	}
	if !cfg.Enabled {
		t.Error("expected checkpointing enabled by default")
	}
	if cfg.IntervalSeconds != 300 {
		t.Errorf("expected interval 300, got %d", cfg.IntervalSeconds)
	}
	if cfg.MaxCheckpoints != 10 {
		t.Errorf("expected max checkpoints 10, got %d", cfg.MaxCheckpoints)
	}
	if !cfg.RetainOnShutdown {
		t.Error("expected retain on shutdown true by default")
	}
}

// --- Checkpointer tests ---

type mockStateProvider struct {
	containerID string
	stateData   []byte
	stateErr    error
}

func (m *mockStateProvider) GetState() ([]byte, error) {
	return m.stateData, m.stateErr
}

func (m *mockStateProvider) GetContainerID() string {
	return m.containerID
}

func TestNewCheckpointer(t *testing.T) {
	tmpDir := t.TempDir()
	mgr, err := NewManager(&SnapshotConfig{
		StoragePath:       tmpDir,
		MaxSnapshots:      10,
		MaxTotalSize:      10 * 1024 * 1024,
		CompressionLevel:  0,
		EncryptionEnabled: false,
		RetentionDays:     7,
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Run("nil config uses defaults", func(t *testing.T) {
		cp := NewCheckpointer(mgr, nil)
		if cp == nil {
			t.Fatal("NewCheckpointer returned nil")
		}
		if !cp.IsEnabled() {
			t.Error("expected checkpointing enabled by default")
		}
	})

	t.Run("disabled config", func(t *testing.T) {
		cp := NewCheckpointer(mgr, &CheckpointConfig{Enabled: false})
		if cp.IsEnabled() {
			t.Error("expected checkpointing disabled")
		}
	})
}

func TestCheckpointerRegisterUnregister(t *testing.T) {
	tmpDir := t.TempDir()
	mgr, err := NewManager(&SnapshotConfig{
		StoragePath:       tmpDir,
		MaxSnapshots:      10,
		MaxTotalSize:      10 * 1024 * 1024,
		CompressionLevel:  0,
		EncryptionEnabled: false,
		RetentionDays:     7,
	})
	if err != nil {
		t.Fatal(err)
	}

	cp := NewCheckpointer(mgr, &CheckpointConfig{
		Enabled:          true,
		IntervalSeconds:  3600, // Long interval to prevent automatic triggering
		MaxCheckpoints:   10,
		RetainOnShutdown: true,
	})

	provider := &mockStateProvider{
		containerID: "c1",
		stateData:   []byte("state"),
	}

	if err := cp.RegisterContainer(provider); err != nil {
		t.Fatalf("RegisterContainer() error: %v", err)
	}

	containers := cp.GetRegisteredContainers()
	if len(containers) != 1 {
		t.Errorf("expected 1 registered container, got %d", len(containers))
	}

	// Register again should be no-op
	if err := cp.RegisterContainer(provider); err != nil {
		t.Fatalf("second RegisterContainer() error: %v", err)
	}
	containers = cp.GetRegisteredContainers()
	if len(containers) != 1 {
		t.Errorf("expected 1 registered container after duplicate, got %d", len(containers))
	}

	// Unregister
	cp.UnregisterContainer("c1")
	containers = cp.GetRegisteredContainers()
	if len(containers) != 0 {
		t.Errorf("expected 0 registered containers after unregister, got %d", len(containers))
	}
}

func TestCheckpointerDisabledRegister(t *testing.T) {
	tmpDir := t.TempDir()
	mgr, err := NewManager(&SnapshotConfig{
		StoragePath:       tmpDir,
		MaxSnapshots:      10,
		CompressionLevel:  0,
		EncryptionEnabled: false,
		RetentionDays:     7,
	})
	if err != nil {
		t.Fatal(err)
	}

	cp := NewCheckpointer(mgr, &CheckpointConfig{Enabled: false})
	provider := &mockStateProvider{containerID: "c1", stateData: []byte("data")}

	// Should not error even if disabled
	if err := cp.RegisterContainer(provider); err != nil {
		t.Fatalf("RegisterContainer() error: %v", err)
	}

	// Should not actually register
	containers := cp.GetRegisteredContainers()
	if len(containers) != 0 {
		t.Errorf("expected 0 containers when disabled, got %d", len(containers))
	}
}

func TestCheckpointerTriggerCheckpoint(t *testing.T) {
	tmpDir := t.TempDir()
	mgr, err := NewManager(&SnapshotConfig{
		StoragePath:       tmpDir,
		MaxSnapshots:      10,
		MaxTotalSize:      10 * 1024 * 1024,
		CompressionLevel:  0,
		EncryptionEnabled: false,
		RetentionDays:     7,
	})
	if err != nil {
		t.Fatal(err)
	}

	cp := NewCheckpointer(mgr, &CheckpointConfig{
		Enabled:         true,
		IntervalSeconds: 3600,
		MaxCheckpoints:  10,
	})

	provider := &mockStateProvider{
		containerID: "c1",
		stateData:   []byte("checkpoint state data"),
	}
	cp.RegisterContainer(provider)

	snap, err := cp.TriggerCheckpoint("c1")
	if err != nil {
		t.Fatalf("TriggerCheckpoint() error: %v", err)
	}
	if snap == nil {
		t.Fatal("expected non-nil snapshot")
	}
	if snap.Type != SnapshotTypeCheckpoint {
		t.Errorf("expected type checkpoint, got %s", snap.Type)
	}
	if snap.Metadata["manual"] != "true" {
		t.Error("expected manual=true in metadata")
	}

	// Trigger for unregistered container
	snap, err = cp.TriggerCheckpoint("unknown")
	if err != nil {
		t.Fatal(err)
	}
	if snap != nil {
		t.Error("expected nil for unregistered container")
	}
}

func TestCheckpointerGetLastNextCheckpoint(t *testing.T) {
	tmpDir := t.TempDir()
	mgr, err := NewManager(&SnapshotConfig{
		StoragePath:       tmpDir,
		MaxSnapshots:      10,
		MaxTotalSize:      10 * 1024 * 1024,
		CompressionLevel:  0,
		EncryptionEnabled: false,
		RetentionDays:     7,
	})
	if err != nil {
		t.Fatal(err)
	}

	cp := NewCheckpointer(mgr, &CheckpointConfig{
		Enabled:         true,
		IntervalSeconds: 3600,
		MaxCheckpoints:  10,
	})

	provider := &mockStateProvider{containerID: "c1", stateData: []byte("data")}
	cp.RegisterContainer(provider)

	// Last checkpoint should be zero initially
	_, hasLast := cp.GetLastCheckpoint("c1")
	if hasLast {
		t.Error("expected no last checkpoint initially")
	}

	// Next checkpoint should be set
	next, hasNext := cp.GetNextCheckpoint("c1")
	if !hasNext {
		t.Error("expected next checkpoint to be set")
	}
	if next.IsZero() {
		t.Error("expected non-zero next checkpoint time")
	}

	// Unknown container
	_, hasLast = cp.GetLastCheckpoint("unknown")
	if hasLast {
		t.Error("expected no last checkpoint for unknown container")
	}
	_, hasNext = cp.GetNextCheckpoint("unknown")
	if hasNext {
		t.Error("expected no next checkpoint for unknown container")
	}
}

func TestCheckpointerSetInterval(t *testing.T) {
	tmpDir := t.TempDir()
	mgr, err := NewManager(&SnapshotConfig{
		StoragePath:       tmpDir,
		MaxSnapshots:      10,
		MaxTotalSize:      10 * 1024 * 1024,
		CompressionLevel:  0,
		EncryptionEnabled: false,
		RetentionDays:     7,
	})
	if err != nil {
		t.Fatal(err)
	}

	cp := NewCheckpointer(mgr, &CheckpointConfig{
		Enabled:         true,
		IntervalSeconds: 3600,
		MaxCheckpoints:  10,
	})

	provider := &mockStateProvider{containerID: "c1", stateData: []byte("data")}
	cp.RegisterContainer(provider)

	// Change interval
	cp.SetInterval("c1", 60)

	next, ok := cp.GetNextCheckpoint("c1")
	if !ok {
		t.Fatal("expected next checkpoint after interval change")
	}
	// Next checkpoint should be roughly 60 seconds from now
	expectedNext := time.Now().Add(60 * time.Second)
	diff := next.Sub(expectedNext)
	if diff > 2*time.Second || diff < -2*time.Second {
		t.Errorf("next checkpoint not updated properly, diff: %v", diff)
	}

	// Setting for unknown container should be no-op (no panic)
	cp.SetInterval("unknown", 120)
}
