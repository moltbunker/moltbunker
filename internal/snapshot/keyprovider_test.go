package snapshot

import (
	"bytes"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"

	"github.com/99designs/keyring"
)

func TestFileKeyProvider_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".snapshot_key")

	p, err := NewFileKeyProvider(path)
	if err != nil {
		t.Fatalf("NewFileKeyProvider: %v", err)
	}

	k1, err := p.MasterKey()
	if err != nil {
		t.Fatalf("MasterKey (generate): %v", err)
	}
	if len(k1) != masterKeySize {
		t.Fatalf("key size = %d, want %d", len(k1), masterKeySize)
	}

	// A fresh provider at the same path must load the SAME key (no orphaning).
	p2, err := NewFileKeyProvider(path)
	if err != nil {
		t.Fatalf("NewFileKeyProvider 2: %v", err)
	}
	k2, err := p2.MasterKey()
	if err != nil {
		t.Fatalf("MasterKey (reload): %v", err)
	}
	if !bytes.Equal(k1, k2) {
		t.Error("reloaded key differs from persisted key")
	}

	// Cached call returns the same key.
	k3, err := p.MasterKey()
	if err != nil {
		t.Fatalf("MasterKey (cached): %v", err)
	}
	if !bytes.Equal(k1, k3) {
		t.Error("cached key differs")
	}

	// Returned slice must be a copy, not the cached backing array.
	k1[0] ^= 0xFF
	k4, _ := p.MasterKey()
	if bytes.Equal(k1, k4) {
		t.Error("MasterKey returned an aliased slice; mutation leaked into cache")
	}
}

func TestFileKeyProvider_InvalidSize(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".snapshot_key")
	if err := os.WriteFile(path, []byte("too short"), 0600); err != nil {
		t.Fatalf("write key: %v", err)
	}

	p, err := NewFileKeyProvider(path)
	if err != nil {
		t.Fatalf("NewFileKeyProvider: %v", err)
	}
	if _, err := p.MasterKey(); err == nil {
		t.Fatal("expected error for wrong-size key file, got nil")
	}
}

func TestFileKeyProvider_EmptyPath(t *testing.T) {
	if _, err := NewFileKeyProvider(""); err == nil {
		t.Fatal("expected error for empty path")
	}
}

func TestEnvKeyProvider_Missing(t *testing.T) {
	t.Setenv(snapshotKeyEnvVar, "")
	p := NewEnvKeyProvider()
	if _, err := p.MasterKey(); err == nil {
		t.Fatal("expected error when env var is unset")
	}
}

func TestEnvKeyProvider_WrongLength(t *testing.T) {
	t.Setenv(snapshotKeyEnvVar, "abcd") // valid hex, but only 2 bytes
	p := NewEnvKeyProvider()
	if _, err := p.MasterKey(); err == nil {
		t.Fatal("expected error for wrong-length key")
	}
}

func TestEnvKeyProvider_InvalidHex(t *testing.T) {
	t.Setenv(snapshotKeyEnvVar, "not-hex-zzzz")
	p := NewEnvKeyProvider()
	if _, err := p.MasterKey(); err == nil {
		t.Fatal("expected error for invalid hex")
	}
}

func TestEnvKeyProvider_Valid(t *testing.T) {
	want := make([]byte, masterKeySize)
	for i := range want {
		want[i] = byte(i)
	}
	t.Setenv(snapshotKeyEnvVar, hex.EncodeToString(want))

	p := NewEnvKeyProvider()
	got, err := p.MasterKey()
	if err != nil {
		t.Fatalf("MasterKey: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("key mismatch: got %x want %x", got, want)
	}
}

func TestNewKeyProviderFromConfig(t *testing.T) {
	dir := t.TempDir()

	t.Run("default is file", func(t *testing.T) {
		kp, err := NewKeyProviderFromConfig(&SnapshotConfig{StoragePath: dir}, dir)
		if err != nil {
			t.Fatalf("factory: %v", err)
		}
		if _, ok := kp.(*FileKeyProvider); !ok {
			t.Errorf("default backend = %T, want *FileKeyProvider", kp)
		}
	})

	t.Run("explicit env", func(t *testing.T) {
		kp, err := NewKeyProviderFromConfig(&SnapshotConfig{KeyProviderBackend: "env"}, dir)
		if err != nil {
			t.Fatalf("factory: %v", err)
		}
		if _, ok := kp.(*EnvKeyProvider); !ok {
			t.Errorf("backend = %T, want *EnvKeyProvider", kp)
		}
	})

	t.Run("explicit keyring", func(t *testing.T) {
		kp, err := NewKeyProviderFromConfig(&SnapshotConfig{KeyProviderBackend: "keyring", StoragePath: dir}, dir)
		if err != nil {
			t.Fatalf("factory: %v", err)
		}
		if _, ok := kp.(*KeyringKeyProvider); !ok {
			t.Errorf("backend = %T, want *KeyringKeyProvider", kp)
		}
	})

	t.Run("unknown backend errors", func(t *testing.T) {
		if _, err := NewKeyProviderFromConfig(&SnapshotConfig{KeyProviderBackend: "bogus"}, dir); err == nil {
			t.Fatal("expected error for unknown backend")
		}
	})

	t.Run("nil config errors", func(t *testing.T) {
		if _, err := NewKeyProviderFromConfig(nil, dir); err == nil {
			t.Fatal("expected error for nil config")
		}
	})
}

// TestKeyringKeyProvider_MockBackend exercises the KeyringKeyProvider contract
// against an in-memory keyring (no OS keychain prompt) by overriding the opener.
func TestKeyringKeyProvider_MockBackend(t *testing.T) {
	orig := keyringOpener
	t.Cleanup(func() { keyringOpener = orig })

	ring := newMemKeyring()
	keyringOpener = func(_ keyring.Config) (keyring.Keyring, error) {
		return ring, nil
	}

	p := NewKeyringKeyProvider("", "", t.TempDir())
	k1, err := p.MasterKey()
	if err != nil {
		t.Fatalf("MasterKey (generate): %v", err)
	}
	if len(k1) != masterKeySize {
		t.Fatalf("key size = %d, want %d", len(k1), masterKeySize)
	}

	// A second provider must reload the same key from the (shared) keyring.
	p2 := NewKeyringKeyProvider("", "", t.TempDir())
	k2, err := p2.MasterKey()
	if err != nil {
		t.Fatalf("MasterKey (reload): %v", err)
	}
	if !bytes.Equal(k1, k2) {
		t.Error("reloaded keyring key differs from stored key")
	}
}

// memKeyring is a minimal in-memory keyring.Keyring for tests.
type memKeyring struct {
	items map[string]keyring.Item
}

func newMemKeyring() *memKeyring {
	return &memKeyring{items: make(map[string]keyring.Item)}
}

func (m *memKeyring) Get(key string) (keyring.Item, error) {
	it, ok := m.items[key]
	if !ok {
		return keyring.Item{}, keyring.ErrKeyNotFound
	}
	return it, nil
}

func (m *memKeyring) GetMetadata(string) (keyring.Metadata, error) {
	return keyring.Metadata{}, keyring.ErrMetadataNotSupported
}

func (m *memKeyring) Set(item keyring.Item) error {
	m.items[item.Key] = item
	return nil
}

func (m *memKeyring) Remove(key string) error {
	delete(m.items, key)
	return nil
}

func (m *memKeyring) Keys() ([]string, error) {
	out := make([]string, 0, len(m.items))
	for k := range m.items {
		out = append(out, k)
	}
	return out, nil
}
