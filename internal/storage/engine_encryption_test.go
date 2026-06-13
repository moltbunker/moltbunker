package storage

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/moltbunker/moltbunker/internal/security"
	"github.com/moltbunker/moltbunker/internal/state"
)

// fakeProviderResolver is a self-recipient key resolver backed by an in-memory
// X25519 keypair.
type fakeProviderResolver struct {
	pub  []byte
	priv []byte
}

func newFakeResolver(t *testing.T) *fakeProviderResolver {
	t.Helper()
	pub, priv, err := security.GenerateX25519KeyPair()
	if err != nil {
		t.Fatalf("GenerateX25519KeyPair: %v", err)
	}
	return &fakeProviderResolver{pub: pub, priv: priv}
}

func (f *fakeProviderResolver) PublicKey() []byte  { return append([]byte(nil), f.pub...) }
func (f *fakeProviderResolver) PrivateKey() []byte { return append([]byte(nil), f.priv...) }

func newEncryptedEngine(t *testing.T) (*StorageEngine, *fakeProviderResolver) {
	t.Helper()
	dir := t.TempDir()
	store, err := state.NewBboltStore(filepath.Join(dir, "state.db"), nil)
	if err != nil {
		t.Fatalf("NewBboltStore: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	eng, err := NewStorageEngine(dir, store, DefaultEngineConfig())
	if err != nil {
		t.Fatalf("NewStorageEngine: %v", err)
	}

	resolver := newFakeResolver(t)
	ks, err := NewProviderKeyStore(resolver)
	if err != nil {
		t.Fatalf("NewProviderKeyStore: %v", err)
	}
	eng.WithOwnerKeyStore(ks)
	return eng, resolver
}

func TestStorageEngine_EncryptedPutGet(t *testing.T) {
	eng, _ := newEncryptedEngine(t)
	ctx := context.Background()
	owner := "0xAc1D8d6e25E54c05986E8bFa9b759063D5e69592"

	if _, err := eng.CreateBucket(ctx, "my-bucket", owner); err != nil {
		t.Fatalf("CreateBucket: %v", err)
	}

	plaintext := []byte("top secret object contents at rest")
	obj, err := eng.PutObject(ctx, &PutObjectInput{
		Bucket: "my-bucket",
		Key:    "secret.txt",
		Body:   bytes.NewReader(plaintext),
		Owner:  owner,
	})
	if err != nil {
		t.Fatalf("PutObject: %v", err)
	}
	if len(obj.EncryptedDEK) == 0 {
		t.Fatal("expected object to carry a sealed DEK")
	}
	if obj.Size != int64(len(plaintext)) {
		t.Errorf("Size = %d, want %d (plaintext size)", obj.Size, len(plaintext))
	}

	// On-disk blob must be ciphertext.
	blobPath, _ := eng.blobPath("my-bucket", "secret.txt")
	// #nosec G304 -- test-controlled path
	onDisk, err := os.ReadFile(blobPath)
	if err != nil {
		t.Fatalf("read blob: %v", err)
	}
	if bytes.Contains(onDisk, plaintext) {
		t.Fatal("plaintext leaked into encrypted blob on disk")
	}

	// GetObject returns decrypted plaintext.
	out, err := eng.GetObject(ctx, "my-bucket", "secret.txt", owner)
	if err != nil {
		t.Fatalf("GetObject: %v", err)
	}
	defer out.Body.Close()
	got, err := io.ReadAll(out.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if !bytes.Equal(got, plaintext) {
		t.Errorf("round-trip mismatch: got %q want %q", got, plaintext)
	}
}

// TestStorageEngine_LegacyReadthrough writes an object whose blob is encrypted
// with the legacy HKDF-derived symmetric key and a raw AES-GCM-wrapped DEK, then
// verifies GetObject transparently decrypts it via the back-compat path.
func TestStorageEngine_LegacyReadthrough(t *testing.T) {
	eng, _ := newEncryptedEngine(t)
	ctx := context.Background()
	owner := "0xLegacyOwner"

	if _, err := eng.CreateBucket(ctx, "legacy-bucket", owner); err != nil {
		t.Fatalf("CreateBucket: %v", err)
	}

	plaintext := []byte("legacy object written by the old scheme")

	// Manually produce a legacy-encrypted blob + wrapped DEK.
	legacyKey, err := legacyDeriveOwnerKey(owner)
	if err != nil {
		t.Fatalf("legacyDeriveOwnerKey: %v", err)
	}
	res, err := NewObjectEncryptor().Encrypt(plaintext, legacyKey)
	if err != nil {
		t.Fatalf("legacy Encrypt: %v", err)
	}

	// Write the ciphertext blob directly.
	blobPath, _ := eng.blobPath("legacy-bucket", "old.txt")
	if err := os.MkdirAll(filepath.Dir(blobPath), 0700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(blobPath, res.Ciphertext, 0600); err != nil {
		t.Fatalf("write blob: %v", err)
	}

	// Persist metadata with the legacy (non-JSON) wrapped DEK.
	if err := eng.metadata.PutObject(ctx, &ObjectInfo{
		Bucket:       "legacy-bucket",
		Key:          "old.txt",
		Size:         int64(len(plaintext)),
		Owner:        owner,
		EncryptedDEK: res.EncryptedDEK,
	}); err != nil {
		t.Fatalf("PutObject metadata: %v", err)
	}

	out, err := eng.GetObject(ctx, "legacy-bucket", "old.txt", owner)
	if err != nil {
		t.Fatalf("GetObject (legacy): %v", err)
	}
	defer out.Body.Close()
	got, err := io.ReadAll(out.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if !bytes.Equal(got, plaintext) {
		t.Errorf("legacy read-through mismatch: got %q want %q", got, plaintext)
	}
}

func TestStorageEngine_PlaintextWhenNoKeyStore(t *testing.T) {
	dir := t.TempDir()
	store, err := state.NewBboltStore(filepath.Join(dir, "state.db"), nil)
	if err != nil {
		t.Fatalf("NewBboltStore: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })
	eng, err := NewStorageEngine(dir, store, DefaultEngineConfig())
	if err != nil {
		t.Fatalf("NewStorageEngine: %v", err)
	}

	ctx := context.Background()
	owner := "0xOwner"
	if _, err := eng.CreateBucket(ctx, "plain-bucket", owner); err != nil {
		t.Fatalf("CreateBucket: %v", err)
	}

	plaintext := []byte("not encrypted")
	obj, err := eng.PutObject(ctx, &PutObjectInput{
		Bucket: "plain-bucket",
		Key:    "p.txt",
		Body:   bytes.NewReader(plaintext),
		Owner:  owner,
	})
	if err != nil {
		t.Fatalf("PutObject: %v", err)
	}
	if len(obj.EncryptedDEK) != 0 {
		t.Error("object should not carry a sealed DEK when encryption disabled")
	}

	blobPath, _ := eng.blobPath("plain-bucket", "p.txt")
	// #nosec G304 -- test-controlled path
	onDisk, _ := os.ReadFile(blobPath)
	if !bytes.Equal(onDisk, plaintext) {
		t.Error("blob should be stored as plaintext when encryption disabled")
	}
}

// TestStorageEngine_EncryptedInMemoryCeiling verifies that the encrypted path
// rejects objects larger than EncryptedMaxInMemoryBytes on both Put and Get,
// while the plaintext path is unaffected by the ceiling.
func TestStorageEngine_EncryptedInMemoryCeiling(t *testing.T) {
	dir := t.TempDir()
	store, err := state.NewBboltStore(filepath.Join(dir, "state.db"), nil)
	if err != nil {
		t.Fatalf("NewBboltStore: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	// Small ceiling so the test stays fast and uses no real memory pressure.
	cfg := DefaultEngineConfig()
	cfg.EncryptedMaxInMemoryBytes = 1024 // 1KB
	eng, err := NewStorageEngine(dir, store, cfg)
	if err != nil {
		t.Fatalf("NewStorageEngine: %v", err)
	}
	resolver := newFakeResolver(t)
	ks, err := NewProviderKeyStore(resolver)
	if err != nil {
		t.Fatalf("NewProviderKeyStore: %v", err)
	}
	eng.WithOwnerKeyStore(ks)

	ctx := context.Background()
	owner := "0xAc1D8d6e25E54c05986E8bFa9b759063D5e69592"
	if _, err := eng.CreateBucket(ctx, "ceil-bucket", owner); err != nil {
		t.Fatalf("CreateBucket: %v", err)
	}

	// Over-ceiling Put must be rejected.
	big := bytes.Repeat([]byte("A"), 2048) // 2KB > 1KB ceiling
	if _, err := eng.PutObject(ctx, &PutObjectInput{
		Bucket: "ceil-bucket", Key: "big.bin", Body: bytes.NewReader(big), Owner: owner,
	}); err == nil {
		t.Fatal("expected oversized encrypted PutObject to be rejected")
	}

	// The temp blob must not be left behind as a committed object.
	if _, err := eng.GetObject(ctx, "ceil-bucket", "big.bin", owner); err == nil {
		t.Fatal("oversized object should not have been stored")
	}

	// Under-ceiling Put/Get round-trips fine.
	small := []byte("small enough")
	if _, err := eng.PutObject(ctx, &PutObjectInput{
		Bucket: "ceil-bucket", Key: "small.txt", Body: bytes.NewReader(small), Owner: owner,
	}); err != nil {
		t.Fatalf("small PutObject: %v", err)
	}
	out, err := eng.GetObject(ctx, "ceil-bucket", "small.txt", owner)
	if err != nil {
		t.Fatalf("small GetObject: %v", err)
	}
	defer out.Body.Close()
	got, _ := io.ReadAll(out.Body)
	if !bytes.Equal(got, small) {
		t.Errorf("round-trip mismatch: got %q want %q", got, small)
	}

	// GetObject ceiling: simulate a stored object whose recorded plaintext Size
	// exceeds the ceiling and confirm Get rejects it before buffering.
	if err := eng.metadata.PutObject(ctx, &ObjectInfo{
		Bucket: "ceil-bucket", Key: "oversize-meta", Size: 4096, Owner: owner,
		EncryptedDEK: []byte(`{"ephemeral_pub":"AAAA"}`), // non-empty so the decrypt path is taken
	}); err != nil {
		t.Fatalf("PutObject metadata: %v", err)
	}
	// Write a dummy blob so os.Open succeeds and the Size guard is what trips.
	blobPath, _ := eng.blobPath("ceil-bucket", "oversize-meta")
	if err := os.WriteFile(blobPath, []byte("dummy"), 0600); err != nil {
		t.Fatalf("write blob: %v", err)
	}
	if _, err := eng.GetObject(ctx, "ceil-bucket", "oversize-meta", owner); err == nil {
		t.Fatal("expected oversized encrypted GetObject to be rejected by the Size guard")
	}
}

// TestStorageEngine_PlaintextIgnoresCeiling verifies the in-memory ceiling does
// not constrain the plaintext (streaming) path.
func TestStorageEngine_PlaintextIgnoresCeiling(t *testing.T) {
	dir := t.TempDir()
	store, err := state.NewBboltStore(filepath.Join(dir, "state.db"), nil)
	if err != nil {
		t.Fatalf("NewBboltStore: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })
	cfg := DefaultEngineConfig()
	cfg.EncryptedMaxInMemoryBytes = 1024 // 1KB ceiling, but encryption disabled
	eng, err := NewStorageEngine(dir, store, cfg)
	if err != nil {
		t.Fatalf("NewStorageEngine: %v", err)
	}
	ctx := context.Background()
	owner := "0xOwner"
	if _, err := eng.CreateBucket(ctx, "plain-ceil", owner); err != nil {
		t.Fatalf("CreateBucket: %v", err)
	}
	big := bytes.Repeat([]byte("B"), 4096) // 4KB > ceiling, but plaintext path streams
	if _, err := eng.PutObject(ctx, &PutObjectInput{
		Bucket: "plain-ceil", Key: "big.bin", Body: bytes.NewReader(big), Owner: owner,
	}); err != nil {
		t.Fatalf("plaintext PutObject should ignore encrypted ceiling: %v", err)
	}
}

func TestNewProviderKeyStore_NilResolver(t *testing.T) {
	if _, err := NewProviderKeyStore(nil); err == nil {
		t.Fatal("expected error for nil resolver")
	}
}

func TestWalletKeyStore_Stubbed(t *testing.T) {
	ks := NewWalletKeyStore()
	if _, err := ks.PublicKeyForOwner("0x1"); err != ErrOwnerKeyNotImplemented {
		t.Errorf("PublicKeyForOwner err = %v, want ErrOwnerKeyNotImplemented", err)
	}
	if _, err := ks.PrivateKeyForOwner("0x1"); err != ErrOwnerKeyNotImplemented {
		t.Errorf("PrivateKeyForOwner err = %v, want ErrOwnerKeyNotImplemented", err)
	}
}
