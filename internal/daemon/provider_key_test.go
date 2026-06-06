package daemon

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/moltbunker/moltbunker/internal/security"
)

func TestLoadOrCreateProviderKey_GeneratesAndPersists(t *testing.T) {
	dir := t.TempDir()

	pm, err := LoadOrCreateProviderKey(dir)
	if err != nil {
		t.Fatalf("LoadOrCreateProviderKey: %v", err)
	}
	if len(pm.PublicKey()) != security.X25519KeySize {
		t.Fatalf("public key size = %d, want %d", len(pm.PublicKey()), security.X25519KeySize)
	}

	// Key file must exist with 0600 perms.
	path := filepath.Join(dir, providerKeyFile)
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat key file: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Fatalf("key file perm = %o, want 0600", perm)
	}
	if info.Size() != int64(security.X25519KeySize) {
		t.Fatalf("key file size = %d, want %d", info.Size(), security.X25519KeySize)
	}
}

func TestLoadOrCreateProviderKey_StableAcrossReload(t *testing.T) {
	dir := t.TempDir()

	pm1, err := LoadOrCreateProviderKey(dir)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	pm2, err := LoadOrCreateProviderKey(dir)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}

	if !bytes.Equal(pm1.PublicKey(), pm2.PublicKey()) {
		t.Fatal("provider public key changed across reload (not stable)")
	}
}

// TestProviderKey_ECIESRoundTrip simulates the CLI sealing an exec key to the
// daemon's stable public key and the daemon unwrapping it (the prepareExecAgent
// crypto path), proving interop end to end.
func TestProviderKey_ECIESRoundTrip(t *testing.T) {
	dir := t.TempDir()
	pm, err := LoadOrCreateProviderKey(dir)
	if err != nil {
		t.Fatalf("LoadOrCreateProviderKey: %v", err)
	}

	// CLI side: seal a 32-byte exec key to the provider public key.
	execKey := bytes.Repeat([]byte{0xAB}, execKeyLen)
	env, err := security.SealToX25519(pm.PublicKey(), execKey)
	if err != nil {
		t.Fatalf("SealToX25519: %v", err)
	}

	// Daemon side: unwrap with the stable private key.
	recovered, err := security.OpenFromX25519(pm.privateKey(), &security.X25519Envelope{
		EphemeralPub: env.EphemeralPub,
		Ciphertext:   env.Ciphertext,
	})
	if err != nil {
		t.Fatalf("OpenFromX25519: %v", err)
	}
	if !bytes.Equal(recovered, execKey) {
		t.Fatalf("recovered exec key mismatch: got %x want %x", recovered, execKey)
	}
}

// TestPrepareExecAgent_DecryptsEnvelope exercises prepareExecAgent end to end:
// it must unwrap the ECIES envelope and write the 32-byte plaintext to the
// mounted key file.
func TestPrepareExecAgent_DecryptsEnvelope(t *testing.T) {
	dir := t.TempDir()
	pm, err := LoadOrCreateProviderKey(dir)
	if err != nil {
		t.Fatalf("provider key: %v", err)
	}

	// Place a fake exec-agent binary so resolveExecAgentPath succeeds.
	exe, err := os.Executable()
	if err != nil {
		t.Fatalf("os.Executable: %v", err)
	}
	agentPath := filepath.Join(filepath.Dir(exe), "exec-agent")
	if err := os.WriteFile(agentPath, []byte("#!/bin/true\n"), 0755); err != nil {
		t.Skipf("cannot write fake exec-agent next to test binary: %v", err)
	}
	defer os.Remove(agentPath)

	cm := &ContainerManager{dataDir: dir, providerKey: pm}

	execKey := bytes.Repeat([]byte{0x42}, execKeyLen)
	env, err := security.SealToX25519(pm.PublicKey(), execKey)
	if err != nil {
		t.Fatalf("seal: %v", err)
	}

	mounts, keyPath, err := cm.prepareExecAgent("dep-test", env.Ciphertext, env.EphemeralPub)
	if err != nil {
		t.Fatalf("prepareExecAgent: %v", err)
	}
	if len(mounts) != 2 {
		t.Fatalf("expected 2 bind mounts, got %d", len(mounts))
	}

	written, err := os.ReadFile(keyPath)
	if err != nil {
		t.Fatalf("read written key: %v", err)
	}
	if !bytes.Equal(written, execKey) {
		t.Fatal("written exec key does not match the original plaintext")
	}
}

func TestPrepareExecAgent_RejectsTamperedEnvelope(t *testing.T) {
	dir := t.TempDir()
	pm, err := LoadOrCreateProviderKey(dir)
	if err != nil {
		t.Fatalf("provider key: %v", err)
	}
	cm := &ContainerManager{dataDir: dir, providerKey: pm}

	env, err := security.SealToX25519(pm.PublicKey(), bytes.Repeat([]byte{0x7}, execKeyLen))
	if err != nil {
		t.Fatalf("seal: %v", err)
	}
	bad := append([]byte(nil), env.Ciphertext...)
	bad[len(bad)-1] ^= 0xFF

	if _, _, err := cm.prepareExecAgent("dep-bad", bad, env.EphemeralPub); err == nil {
		t.Fatal("expected prepareExecAgent to reject tampered envelope")
	}
}

func TestPrepareExecAgent_NoProviderKey(t *testing.T) {
	cm := &ContainerManager{dataDir: t.TempDir()} // providerKey == nil
	if _, _, err := cm.prepareExecAgent("dep-x", []byte("ct"), make([]byte, security.X25519KeySize)); err == nil {
		t.Fatal("expected error when provider key is unavailable")
	}
}
