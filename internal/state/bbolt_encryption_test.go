package state

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	bolt "go.etcd.io/bbolt"

	"github.com/moltbunker/moltbunker/internal/security"
)

func newKey(t *testing.T) []byte {
	t.Helper()
	k, err := security.GenerateKey(32)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}
	return k
}

func openEncrypted(t *testing.T, key []byte) (*BboltStore, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "enc.db")
	s, err := NewBboltStore(path, key)
	if err != nil {
		t.Fatalf("NewBboltStore: %v", err)
	}
	return s, path
}

// TestEncryptedRoundTrip verifies a value written and read back under encryption
// matches the original plaintext.
func TestEncryptedRoundTrip(t *testing.T) {
	ctx := context.Background()
	key := newKey(t)
	s, _ := openEncrypted(t, key)
	defer s.Close()

	plain := []byte(`{"id":"dep-1","image":"nginx:latest","secret":"hunter2"}`)
	if err := s.PutDeployment(ctx, "dep-1", plain); err != nil {
		t.Fatalf("PutDeployment: %v", err)
	}
	got, err := s.GetDeployment(ctx, "dep-1")
	if err != nil {
		t.Fatalf("GetDeployment: %v", err)
	}
	if !bytes.Equal(got, plain) {
		t.Fatalf("round-trip mismatch: got %q want %q", got, plain)
	}
}

// TestRawValueIsEncrypted opens the underlying bolt file directly and asserts the
// stored bytes are NOT the plaintext and DO carry the magic prefix.
func TestRawValueIsEncrypted(t *testing.T) {
	ctx := context.Background()
	key := newKey(t)
	s, path := openEncrypted(t, key)

	plain := []byte(`{"id":"dep-secret","token":"do-not-leak"}`)
	if err := s.PutDeployment(ctx, "dep-secret", plain); err != nil {
		t.Fatalf("PutDeployment: %v", err)
	}
	// Close so we can re-open the file read-only at the bolt level.
	if err := s.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	db, err := bolt.Open(path, 0600, &bolt.Options{ReadOnly: true})
	if err != nil {
		t.Fatalf("bolt.Open: %v", err)
	}
	defer db.Close()

	var raw []byte
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BucketDeployments))
		v := b.Get([]byte("dep-secret"))
		raw = append([]byte(nil), v...)
		return nil
	})
	if err != nil {
		t.Fatalf("View: %v", err)
	}

	if bytes.Equal(raw, plain) {
		t.Fatal("raw on-disk value equals plaintext — not encrypted")
	}
	if bytes.Contains(raw, []byte("do-not-leak")) {
		t.Fatal("plaintext token leaked into raw on-disk value")
	}
	if !bytes.HasPrefix(raw, encMagic) {
		t.Fatalf("raw value missing magic prefix: %x", raw)
	}
}

// TestBackCompatPlaintextReadByEncryptedStore verifies a value written by a
// nil-key (plaintext) store is still readable by a key-enabled store — the
// lazy-migration path.
func TestBackCompatPlaintextReadByEncryptedStore(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "compat.db")

	// Write plaintext with a nil-key store.
	plainStore, err := NewBboltStore(path, nil)
	if err != nil {
		t.Fatalf("NewBboltStore plaintext: %v", err)
	}
	legacy := []byte(`{"legacy":true}`)
	if err := plainStore.PutDeployment(ctx, "legacy-1", legacy); err != nil {
		t.Fatalf("PutDeployment plaintext: %v", err)
	}
	if err := plainStore.Close(); err != nil {
		t.Fatalf("Close plaintext: %v", err)
	}

	// Reopen with a key and confirm the legacy plaintext value still reads.
	key := newKey(t)
	encStore, err := NewBboltStore(path, key)
	if err != nil {
		t.Fatalf("NewBboltStore encrypted: %v", err)
	}
	defer encStore.Close()

	got, err := encStore.GetDeployment(ctx, "legacy-1")
	if err != nil {
		t.Fatalf("GetDeployment legacy: %v", err)
	}
	if !bytes.Equal(got, legacy) {
		t.Fatalf("legacy read mismatch: got %q want %q", got, legacy)
	}

	// Rewriting the value should now encrypt it (lazy migration).
	updated := []byte(`{"legacy":false}`)
	if err := encStore.PutDeployment(ctx, "legacy-1", updated); err != nil {
		t.Fatalf("rewrite: %v", err)
	}
	got, err = encStore.GetDeployment(ctx, "legacy-1")
	if err != nil {
		t.Fatalf("GetDeployment after rewrite: %v", err)
	}
	if !bytes.Equal(got, updated) {
		t.Fatalf("after rewrite mismatch: got %q want %q", got, updated)
	}
}

// TestWrongKeyGetReturnsError verifies that opening an encrypted DB with the
// wrong key causes Get to error rather than return ciphertext.
func TestWrongKeyGetReturnsError(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "wrongkey.db")

	key := newKey(t)
	s1, err := NewBboltStore(path, key)
	if err != nil {
		t.Fatalf("NewBboltStore: %v", err)
	}
	plain := []byte(`{"id":"dep-1"}`)
	if err := s1.PutDeployment(ctx, "dep-1", plain); err != nil {
		t.Fatalf("PutDeployment: %v", err)
	}
	if err := s1.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	wrong := newKey(t)
	s2, err := NewBboltStore(path, wrong)
	if err != nil {
		t.Fatalf("NewBboltStore wrong key: %v", err)
	}
	defer s2.Close()

	got, err := s2.GetDeployment(ctx, "dep-1")
	if err == nil {
		t.Fatalf("expected error decrypting with wrong key, got value %q", got)
	}
	if got != nil {
		t.Fatalf("expected nil value on decrypt failure, got %q", got)
	}
}

// TestWrongKeyListReturnsError verifies List errors (not ciphertext) on bad key.
func TestWrongKeyListReturnsError(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "wrongkeylist.db")

	key := newKey(t)
	s1, err := NewBboltStore(path, key)
	if err != nil {
		t.Fatalf("NewBboltStore: %v", err)
	}
	if err := s1.PutDeployment(ctx, "dep-1", []byte("secret")); err != nil {
		t.Fatalf("PutDeployment: %v", err)
	}
	if err := s1.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	wrong := newKey(t)
	s2, err := NewBboltStore(path, wrong)
	if err != nil {
		t.Fatalf("NewBboltStore wrong key: %v", err)
	}
	defer s2.Close()

	if _, err := s2.ListDeployments(ctx); err == nil {
		t.Fatal("expected error listing with wrong key")
	}
}

// TestEncryptedListDecryptsAll verifies List decrypts every entry.
func TestEncryptedListDecryptsAll(t *testing.T) {
	ctx := context.Background()
	key := newKey(t)
	s, _ := openEncrypted(t, key)
	defer s.Close()

	want := map[string][]byte{
		"d1": []byte("one"),
		"d2": []byte("two"),
		"d3": []byte(`{"three":3}`),
	}
	for k, v := range want {
		if err := s.PutDeployment(ctx, k, v); err != nil {
			t.Fatalf("PutDeployment %s: %v", k, err)
		}
	}

	got, err := s.ListDeployments(ctx)
	if err != nil {
		t.Fatalf("ListDeployments: %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("expected %d entries, got %d", len(want), len(got))
	}
	for k, v := range want {
		if !bytes.Equal(got[k], v) {
			t.Errorf("%s: got %q want %q", k, got[k], v)
		}
	}
}

// TestEncryptedTypedMethods exercises several distinct typed method pairs under
// encryption: Deployment, APIKey, StorageObject, and Meta.
func TestEncryptedTypedMethods(t *testing.T) {
	ctx := context.Background()
	key := newKey(t)
	s, _ := openEncrypted(t, key)
	defer s.Close()

	// Deployment
	if err := s.PutDeployment(ctx, "dep", []byte("dep-val")); err != nil {
		t.Fatalf("PutDeployment: %v", err)
	}
	if got, _ := s.GetDeployment(ctx, "dep"); !bytes.Equal(got, []byte("dep-val")) {
		t.Errorf("Deployment: got %q", got)
	}

	// APIKey (via List, since there is no GetAPIKey)
	if err := s.PutAPIKey(ctx, "key-1", []byte(`{"hash":"abc"}`)); err != nil {
		t.Fatalf("PutAPIKey: %v", err)
	}
	keys, err := s.ListAPIKeys(ctx)
	if err != nil {
		t.Fatalf("ListAPIKeys: %v", err)
	}
	if !bytes.Equal(keys["key-1"], []byte(`{"hash":"abc"}`)) {
		t.Errorf("APIKey: got %q", keys["key-1"])
	}

	// StorageObject
	if err := s.PutStorageObject(ctx, "obj", []byte("blob")); err != nil {
		t.Fatalf("PutStorageObject: %v", err)
	}
	if got, _ := s.GetStorageObject(ctx, "obj"); !bytes.Equal(got, []byte("blob")) {
		t.Errorf("StorageObject: got %q", got)
	}

	// Meta
	if err := s.PutMeta(ctx, "created_at", []byte("2026-06-07")); err != nil {
		t.Fatalf("PutMeta: %v", err)
	}
	if got, _ := s.GetMeta(ctx, "created_at"); !bytes.Equal(got, []byte("2026-06-07")) {
		t.Errorf("Meta: got %q", got)
	}
}

// TestLoadOrCreateStateKey verifies perms, length, and stability across calls.
func TestLoadOrCreateStateKey(t *testing.T) {
	dir := t.TempDir()

	k1, err := LoadOrCreateStateKey(dir)
	if err != nil {
		t.Fatalf("LoadOrCreateStateKey (create): %v", err)
	}
	if len(k1) != stateKeySize {
		t.Fatalf("key length = %d, want %d", len(k1), stateKeySize)
	}

	path := filepath.Join(dir, stateKeyFile)
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	// Permission check is meaningful only on Unix.
	if runtime.GOOS != "windows" {
		if perm := info.Mode().Perm(); perm != 0600 {
			t.Fatalf("key file perms = %o, want 0600", perm)
		}
	}

	k2, err := LoadOrCreateStateKey(dir)
	if err != nil {
		t.Fatalf("LoadOrCreateStateKey (load): %v", err)
	}
	if !bytes.Equal(k1, k2) {
		t.Fatal("key not stable across calls")
	}
}

// TestLoadStateKeyInvalidSize verifies a corrupt-length key file is rejected.
func TestLoadStateKeyInvalidSize(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, stateKeyFile)
	if err := os.WriteFile(path, []byte("too-short"), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if _, err := LoadOrCreateStateKey(dir); err == nil {
		t.Fatal("expected error for invalid key size")
	}
}

// TestEncryptionDisabledIsPlaintext confirms a nil-key store stores raw plaintext
// (no magic prefix) — preserving legacy behavior / the opt-out path.
func TestEncryptionDisabledIsPlaintext(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "plain.db")
	s, err := NewBboltStore(path, nil)
	if err != nil {
		t.Fatalf("NewBboltStore: %v", err)
	}
	plain := []byte(`{"plain":true}`)
	if err := s.PutDeployment(ctx, "dep-1", plain); err != nil {
		t.Fatalf("PutDeployment: %v", err)
	}
	if err := s.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	db, err := bolt.Open(path, 0600, &bolt.Options{ReadOnly: true})
	if err != nil {
		t.Fatalf("bolt.Open: %v", err)
	}
	defer db.Close()
	var raw []byte
	_ = db.View(func(tx *bolt.Tx) error {
		raw = append([]byte(nil), tx.Bucket([]byte(BucketDeployments)).Get([]byte("dep-1"))...)
		return nil
	})
	if !bytes.Equal(raw, plain) {
		t.Fatalf("plaintext store altered value on disk: got %q", raw)
	}
	if bytes.HasPrefix(raw, encMagic) {
		t.Fatal("plaintext store wrote magic prefix")
	}
}
