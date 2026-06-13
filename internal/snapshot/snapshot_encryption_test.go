package snapshot

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// failingKeyProvider always errors, used to assert fail-closed behavior.
type failingKeyProvider struct{}

func (failingKeyProvider) MasterKey() ([]byte, error) {
	return nil, fmt.Errorf("simulated key provider failure")
}

func TestManager_KeyProviderError(t *testing.T) {
	// EncryptionEnabled with a failing provider must error rather than silently
	// going ephemeral.
	_, err := NewManager(&SnapshotConfig{
		StoragePath:       t.TempDir(),
		MaxSnapshots:      5,
		EncryptionEnabled: true,
		RetentionDays:     7,
	}, failingKeyProvider{})
	if err == nil {
		t.Fatal("expected NewManager to fail when key provider errors")
	}
}

func TestManager_EncryptionRequiresProvider(t *testing.T) {
	// EncryptionEnabled with NO provider must be a hard error (no ephemeral key).
	_, err := NewManager(&SnapshotConfig{
		StoragePath:       t.TempDir(),
		MaxSnapshots:      5,
		EncryptionEnabled: true,
		RetentionDays:     7,
	})
	if err == nil {
		t.Fatal("expected NewManager to fail when encryption enabled but no provider supplied")
	}
}

func TestManager_EncryptedRoundTrip(t *testing.T) {
	dir := t.TempDir()
	keyPath := filepath.Join(dir, ".snapshot_key")

	provider, err := NewFileKeyProvider(keyPath)
	if err != nil {
		t.Fatalf("NewFileKeyProvider: %v", err)
	}

	m, err := NewManager(&SnapshotConfig{
		StoragePath:       dir,
		MaxSnapshots:      10,
		MaxTotalSize:      10 * 1024 * 1024,
		CompressionLevel:  0,
		EncryptionEnabled: true,
		RetentionDays:     7,
	}, provider)
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}

	plaintext := []byte("sensitive container state that must be encrypted at rest")
	snap, err := m.CreateSnapshot("c1", plaintext, SnapshotTypeFull, nil)
	if err != nil {
		t.Fatalf("CreateSnapshot: %v", err)
	}
	if !snap.Encrypted {
		t.Fatal("snapshot should be marked encrypted")
	}

	// On-disk blob must NOT contain the plaintext.
	onDisk, err := os.ReadFile(snap.DataPath)
	if err != nil {
		t.Fatalf("read blob: %v", err)
	}
	if bytes.Contains(onDisk, plaintext) {
		t.Fatal("plaintext leaked into encrypted snapshot blob")
	}

	// Round-trip decrypts correctly.
	got, err := m.GetSnapshotData(snap.ID)
	if err != nil {
		t.Fatalf("GetSnapshotData: %v", err)
	}
	if !bytes.Equal(got, plaintext) {
		t.Errorf("round-trip mismatch: got %q want %q", got, plaintext)
	}
}

func TestManager_RotateEncryptionKey(t *testing.T) {
	dir := t.TempDir()
	provider, err := NewFileKeyProvider(filepath.Join(dir, ".snapshot_key"))
	if err != nil {
		t.Fatalf("NewFileKeyProvider: %v", err)
	}

	m, err := NewManager(&SnapshotConfig{
		StoragePath:       dir,
		MaxSnapshots:      10,
		MaxTotalSize:      10 * 1024 * 1024,
		CompressionLevel:  0,
		EncryptionEnabled: true,
		RetentionDays:     7,
	}, provider)
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}

	plaintext := []byte("rotate me")
	snap, err := m.CreateSnapshot("c1", plaintext, SnapshotTypeFull, nil)
	if err != nil {
		t.Fatalf("CreateSnapshot: %v", err)
	}

	// Rotate to a new file-backed key (different path => different key).
	newProvider, err := NewFileKeyProvider(filepath.Join(dir, ".snapshot_key.new"))
	if err != nil {
		t.Fatalf("NewFileKeyProvider (new): %v", err)
	}
	if err := m.RotateEncryptionKey(newProvider); err != nil {
		t.Fatalf("RotateEncryptionKey: %v", err)
	}

	// Data is still readable under the new key.
	got, err := m.GetSnapshotData(snap.ID)
	if err != nil {
		t.Fatalf("GetSnapshotData after rotate: %v", err)
	}
	if !bytes.Equal(got, plaintext) {
		t.Errorf("post-rotation mismatch: got %q want %q", got, plaintext)
	}

	// Nil provider is rejected.
	if err := m.RotateEncryptionKey(nil); err == nil {
		t.Fatal("expected RotateEncryptionKey(nil) to error")
	}
}
