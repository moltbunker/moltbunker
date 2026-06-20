//go:build keyring_integration

// Package snapshot keyring integration test.
//
// This test talks to the real OS keyring (darwin Keychain, linux
// SecretService/kwallet) and is therefore excluded from the default
// `go test ./...` run. Run it explicitly with:
//
//	go test -tags keyring_integration ./internal/snapshot/...
//
// It may prompt for keychain access on first run.
package snapshot

import (
	"bytes"
	"testing"
)

func TestKeyringKeyProvider_RoundTrip(t *testing.T) {
	p := NewKeyringKeyProvider("moltbunker-snapshot-test", "snapshot-master-key-test", t.TempDir())

	k1, err := p.MasterKey()
	if err != nil {
		t.Fatalf("MasterKey (generate): %v", err)
	}
	if len(k1) != masterKeySize {
		t.Fatalf("key size = %d, want %d", len(k1), masterKeySize)
	}

	p2 := NewKeyringKeyProvider("moltbunker-snapshot-test", "snapshot-master-key-test", t.TempDir())
	k2, err := p2.MasterKey()
	if err != nil {
		t.Fatalf("MasterKey (reload): %v", err)
	}
	if !bytes.Equal(k1, k2) {
		t.Error("reloaded keyring key differs from stored key")
	}
}
