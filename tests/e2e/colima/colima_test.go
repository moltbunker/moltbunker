//go:build colima

// Package colima contains E2E tests that run against a real containerd instance
// inside Colima on macOS. These tests verify the ContainerdClient against actual
// container lifecycle operations.
//
// Prerequisites:
//   - Colima running: `colima start --runtime containerd`
//   - containerd socket at ~/.colima/default/containerd.sock
//
// Run with:
//
//	go test -tags colima -v -timeout 5m ./tests/e2e/colima/...
package colima

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/moltbunker/moltbunker/internal/runtime"
)

const (
	testNamespace = "moltbunker-e2e"

	// Small images for fast pulls
	alpineImage  = "docker.io/library/alpine:latest"
	busyboxImage = "docker.io/library/busybox:latest"

	// Timeouts
	pullTimeout      = 2 * time.Minute
	lifecycleTimeout = 30 * time.Second
)

// colimaSocket returns the containerd socket path, checking env override first.
func colimaSocket() string {
	if s := os.Getenv("COLIMA_CONTAINERD_SOCKET"); s != "" {
		return s
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".colima", "default", "containerd.sock")
}

// skipIfNoColima skips the test if Colima's containerd socket is not reachable.
func skipIfNoColima(t *testing.T) {
	t.Helper()

	sock := colimaSocket()
	if sock == "" {
		t.Skip("cannot determine Colima socket path")
	}

	if _, err := os.Stat(sock); os.IsNotExist(err) {
		t.Skipf("Colima containerd socket not found at %s (is Colima running?)", sock)
	}

	// Verify actual connectivity - socket file can exist but be stale
	conn, err := net.DialTimeout("unix", sock, 2*time.Second)
	if err != nil {
		t.Skipf("Colima containerd socket not responding at %s: %v", sock, err)
	}
	conn.Close()
}

// newTestClient creates a ContainerdClient connected to Colima with a temp logs dir.
// It registers cleanup to close the client when the test ends.
//
// The logs directory is created under the user's home (~/.moltbunker/e2e-logs/)
// rather than /var/folders/ because Colima's virtiofs only mounts /Users/ss.
// containerd inside the VM needs write access to the log directory.
func newTestClient(t *testing.T) *runtime.ContainerdClient {
	t.Helper()
	skipIfNoColima(t)

	// Use a directory under $HOME so Colima's virtiofs mount gives the VM access
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("cannot determine home dir: %v", err)
	}
	logsDir := filepath.Join(home, ".moltbunker", "e2e-logs", fmt.Sprintf("run-%d", time.Now().UnixNano()))
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		t.Fatalf("failed to create logs dir: %v", err)
	}
	t.Cleanup(func() {
		os.RemoveAll(logsDir)
	})

	client, err := runtime.NewContainerdClient(colimaSocket(), testNamespace, logsDir, "", nil)
	if err != nil {
		t.Fatalf("failed to create containerd client: %v", err)
	}

	t.Cleanup(func() {
		client.Close()
	})

	return client
}

// containerID generates a unique container ID from the test name.
func containerID(t *testing.T, suffix string) string {
	t.Helper()

	// Sanitize test name: containerd IDs must be [a-zA-Z0-9][a-zA-Z0-9_.-]
	name := strings.ReplaceAll(t.Name(), "/", "-")
	name = strings.ReplaceAll(name, " ", "-")
	name = strings.ToLower(name)

	id := fmt.Sprintf("e2e-%s-%s-%d", name, suffix, time.Now().UnixNano()%100000)
	if len(id) > 76 { // containerd max is 76 chars
		id = id[:76]
	}
	return id
}

// cleanupContainer ensures a container is deleted after the test, ignoring errors
// (the container might already be deleted by the test itself).
func cleanupContainer(t *testing.T, client *runtime.ContainerdClient, id string) {
	t.Helper()
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		// Best effort - ignore errors
		_ = client.StopContainer(ctx, id, 5*time.Second)
		_ = client.DeleteContainer(ctx, id)
	})
}
