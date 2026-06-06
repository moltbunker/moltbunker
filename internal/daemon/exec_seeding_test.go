package daemon

import (
	"os"
	"path/filepath"
	goruntime "runtime"
	"testing"

	"github.com/moltbunker/moltbunker/internal/security"
)

// TestResolveExecAgentPath_GOARCHSuffix verifies that the GOARCH-suffixed
// binary name (e.g. exec-agent-arm64 / exec-agent-amd64) is resolved, not only
// the amd64 name. The candidate is placed next to the test binary.
func TestResolveExecAgentPath_GOARCHSuffix(t *testing.T) {
	exe, err := os.Executable()
	if err != nil {
		t.Fatalf("os.Executable: %v", err)
	}
	dir := filepath.Dir(exe)
	archName := "exec-agent-" + goruntime.GOARCH
	archPath := filepath.Join(dir, archName)

	// Ensure neither the arch-suffixed nor the bare binary pre-exist.
	barePath := filepath.Join(dir, "exec-agent")
	if _, err := os.Stat(barePath); err == nil {
		t.Skipf("bare exec-agent already exists next to test binary; cannot isolate arch resolution")
	}

	if err := os.WriteFile(archPath, []byte("#!/bin/true\n"), 0755); err != nil {
		t.Skipf("cannot write fake %s next to test binary: %v", archName, err)
	}
	defer os.Remove(archPath)

	cm := &ContainerManager{}
	got := cm.resolveExecAgentPath()
	if got != archPath {
		t.Fatalf("resolveExecAgentPath = %q, want %q (GOARCH=%s)", got, archPath, goruntime.GOARCH)
	}
}

// TestResolveExecAgentPath_BareFallback verifies the bare "exec-agent" name is
// resolved when no arch-suffixed binary is present.
func TestResolveExecAgentPath_BareFallback(t *testing.T) {
	exe, err := os.Executable()
	if err != nil {
		t.Fatalf("os.Executable: %v", err)
	}
	dir := filepath.Dir(exe)
	archPath := filepath.Join(dir, "exec-agent-"+goruntime.GOARCH)
	if _, err := os.Stat(archPath); err == nil {
		t.Skipf("arch-suffixed exec-agent already exists; cannot isolate bare fallback")
	}
	barePath := filepath.Join(dir, "exec-agent")
	if _, err := os.Stat(barePath); err == nil {
		t.Skipf("bare exec-agent already exists next to test binary")
	}
	if err := os.WriteFile(barePath, []byte("#!/bin/true\n"), 0755); err != nil {
		t.Skipf("cannot write fake exec-agent: %v", err)
	}
	defer os.Remove(barePath)

	cm := &ContainerManager{}
	if got := cm.resolveExecAgentPath(); got != barePath {
		t.Fatalf("resolveExecAgentPath = %q, want %q", got, barePath)
	}
}

// TestCleanupExecKey_RemovesFile verifies the exec_key file is removed and the
// path is cleared (so teardown leaves no plaintext key material on disk).
func TestCleanupExecKey_RemovesFile(t *testing.T) {
	dir := t.TempDir()
	keyPath := filepath.Join(dir, "dep-test.key")
	if err := os.WriteFile(keyPath, make([]byte, execKeyLen), 0600); err != nil {
		t.Fatalf("write key: %v", err)
	}

	cm := &ContainerManager{}
	d := &Deployment{ExecKeyPath: keyPath}
	cm.cleanupExecKey(d)

	if _, err := os.Stat(keyPath); !os.IsNotExist(err) {
		t.Fatalf("exec key file still present after cleanup (err=%v)", err)
	}
	if d.ExecKeyPath != "" {
		t.Fatalf("ExecKeyPath not cleared: %q", d.ExecKeyPath)
	}
}

// TestCleanupExecKey_Idempotent verifies cleanup is a no-op when there is no
// key file (non-exec deployments, or repeated teardown calls).
func TestCleanupExecKey_Idempotent(t *testing.T) {
	cm := &ContainerManager{}
	d := &Deployment{} // no ExecKeyPath
	cm.cleanupExecKey(d)
	cm.cleanupExecKey(d) // second call must not panic or error
}

// TestPrepareExecAgent_ForeignEnvelopeRejected models a replica receiving an
// envelope sealed to the ORIGINATOR's X25519 key. The replica (with a different
// keypair) must fail to unwrap, which is the signal deployReplica uses to skip
// exec seeding gracefully.
func TestPrepareExecAgent_ForeignEnvelopeRejected(t *testing.T) {
	originatorDir := t.TempDir()
	replicaDir := t.TempDir()

	originator, err := LoadOrCreateProviderKey(originatorDir)
	if err != nil {
		t.Fatalf("originator key: %v", err)
	}
	replica, err := LoadOrCreateProviderKey(replicaDir)
	if err != nil {
		t.Fatalf("replica key: %v", err)
	}

	// Seal to the ORIGINATOR's public key.
	env, err := security.SealToX25519(originator.PublicKey(), make([]byte, execKeyLen))
	if err != nil {
		t.Fatalf("seal: %v", err)
	}

	// Replica tries to unwrap with ITS OWN private key — must fail.
	cm := &ContainerManager{dataDir: replicaDir, providerKey: replica}
	if _, _, err := cm.prepareExecAgent("dep-foreign", env.Ciphertext, env.EphemeralPub); err == nil {
		t.Fatal("expected replica to reject envelope sealed to originator's key")
	}
}
