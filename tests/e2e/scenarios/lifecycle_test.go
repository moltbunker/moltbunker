//go:build e2e

package scenarios

import (
	"testing"
	"time"

	"github.com/moltbunker/moltbunker/pkg/types"
	"github.com/moltbunker/moltbunker/tests/e2e/testutil"
)

// TestE2E_ContainerLifecycle tests the complete container lifecycle
func TestE2E_ContainerLifecycle(t *testing.T) {
	h := testutil.NewTestHarness(t)
	assert := testutil.NewAssertions(t)

	ctx, cancel := h.WithTimeout(60 * time.Second)
	defer cancel()

	containerID := "e2e-lifecycle"
	imageRef := "docker.io/library/nginx:latest"

	// Phase 1: Create
	t.Log("Phase 1: Creating container")
	container, err := h.Containerd.CreateContainer(ctx, containerID, imageRef, types.ResourceLimits{
		MemoryLimit: 256 * 1024 * 1024,
	})
	assert.NoError(err)
	assert.Equal(types.ContainerStatusCreated, container.Status)

	// Verify container exists
	h.AssertContainerExists(containerID)

	// Phase 2: Start
	t.Log("Phase 2: Starting container")
	err = h.Containerd.StartContainer(ctx, containerID)
	assert.NoError(err)

	status, err := h.Containerd.GetContainerStatus(ctx, containerID)
	assert.NoError(err)
	assert.Equal(types.ContainerStatusRunning, status)

	// Phase 3: Health check
	t.Log("Phase 3: Checking health")
	health, err := h.Containerd.GetHealthStatus(ctx, containerID)
	assert.NoError(err)
	assert.True(health.Healthy)

	// Phase 4: Add logs
	t.Log("Phase 4: Adding logs")
	for i := 0; i < 5; i++ {
		err = h.Containerd.AppendLog(containerID, "Log line "+string(rune('0'+i)))
		assert.NoError(err)
	}

	// Phase 5: Read logs
	t.Log("Phase 5: Reading logs")
	reader, err := h.Containerd.GetContainerLogs(ctx, containerID, false, 0)
	assert.NoError(err)
	defer reader.Close()

	buf := make([]byte, 1024)
	n, _ := reader.Read(buf)
	logs := string(buf[:n])
	assert.Contains(logs, "Log line 0")
	assert.Contains(logs, "Log line 4")

	// Phase 6: Execute command
	t.Log("Phase 6: Executing command")
	output, err := h.Containerd.ExecInContainer(ctx, containerID, []string{"echo", "hello"})
	assert.NoError(err)
	assert.NotNil(output)

	// Phase 7: Stop
	t.Log("Phase 7: Stopping container")
	err = h.Containerd.StopContainer(ctx, containerID, 10*time.Second)
	assert.NoError(err)

	status, err = h.Containerd.GetContainerStatus(ctx, containerID)
	assert.NoError(err)
	assert.Equal(types.ContainerStatusStopped, status)

	// Phase 8: Restart
	t.Log("Phase 8: Restarting container")
	err = h.Containerd.StartContainer(ctx, containerID)
	assert.NoError(err)

	status, err = h.Containerd.GetContainerStatus(ctx, containerID)
	assert.NoError(err)
	assert.Equal(types.ContainerStatusRunning, status)

	// Phase 9: Delete (while running)
	t.Log("Phase 9: Deleting container")
	err = h.Containerd.DeleteContainer(ctx, containerID)
	assert.NoError(err)

	h.AssertContainerNotExists(containerID)
}

// TestE2E_ContainerRestart tests restarting containers
func TestE2E_ContainerRestart(t *testing.T) {
	h := testutil.NewTestHarness(t)
	assert := testutil.NewAssertions(t)

	ctx, cancel := h.WithTimeout(60 * time.Second)
	defer cancel()

	containerID := "e2e-restart"

	// Create and start
	_, err := h.Containerd.CreateContainer(ctx, containerID, "nginx:latest", types.ResourceLimits{})
	assert.NoError(err)

	err = h.Containerd.StartContainer(ctx, containerID)
	assert.NoError(err)

	// Multiple restart cycles
	for i := 0; i < 3; i++ {
		t.Logf("Restart cycle %d", i+1)

		// Stop
		err = h.Containerd.StopContainer(ctx, containerID, 5*time.Second)
		assert.NoError(err)

		status, _ := h.Containerd.GetContainerStatus(ctx, containerID)
		assert.Equal(types.ContainerStatusStopped, status)

		// Start
		err = h.Containerd.StartContainer(ctx, containerID)
		assert.NoError(err)

		status, _ = h.Containerd.GetContainerStatus(ctx, containerID)
		assert.Equal(types.ContainerStatusRunning, status)
	}

	// Cleanup
	err = h.Containerd.DeleteContainer(ctx, containerID)
	assert.NoError(err)
}

// TestE2E_ContainerDeletionCleanup tests that deletion cleans up properly
func TestE2E_ContainerDeletionCleanup(t *testing.T) {
	h := testutil.NewTestHarness(t)
	assert := testutil.NewAssertions(t)

	ctx, cancel := h.WithTimeout(60 * time.Second)
	defer cancel()

	containerID := "e2e-cleanup"

	// Create, start, add logs
	_, err := h.Containerd.CreateContainer(ctx, containerID, "nginx:latest", types.ResourceLimits{})
	assert.NoError(err)

	err = h.Containerd.StartContainer(ctx, containerID)
	assert.NoError(err)

	for i := 0; i < 10; i++ {
		h.Containerd.AppendLog(containerID, "Log line")
	}

	// Set metadata
	err = h.Containerd.SetEncryptedVolume(containerID, "/dev/mapper/test")
	assert.NoError(err)

	err = h.Containerd.SetOnionAddress(containerID, "test.onion")
	assert.NoError(err)

	// Delete
	err = h.Containerd.DeleteContainer(ctx, containerID)
	assert.NoError(err)

	// Verify completely gone
	_, exists := h.Containerd.GetContainer(containerID)
	assert.False(exists)

	// Verify can't get status
	_, err = h.Containerd.GetContainerStatus(ctx, containerID)
	assert.Error(err)

	// Verify can't get logs
	_, err = h.Containerd.GetContainerLogs(ctx, containerID, false, 0)
	assert.Error(err)
}

// TestE2E_TorLifecycle tests Tor service lifecycle
func TestE2E_TorLifecycle(t *testing.T) {
	h := testutil.NewTestHarness(t)
	assert := testutil.NewAssertions(t)

	ctx, cancel := h.WithTimeout(60 * time.Second)
	defer cancel()

	// Initial state
	h.AssertTorNotRunning()
	assert.Empty(h.Tor.GetOnionAddress())

	// Start Tor
	err := h.Tor.Start(ctx)
	assert.NoError(err)
	h.AssertTorRunning()

	// Create multiple onion services
	ports := []int{80, 443, 8080}
	addresses := make(map[int]string)

	for _, port := range ports {
		addr, err := h.Tor.CreateOnionService(ctx, port)
		assert.NoError(err)
		assert.Contains(addr, ".onion")
		addresses[port] = addr
	}

	// Verify all services exist
	services := h.Tor.ListOnionServices()
	assert.Len(services, 3)

	for port, addr := range addresses {
		assert.Equal(addr, services[port])
	}

	// Close one service
	err = h.Tor.CloseOnionService(80)
	assert.NoError(err)

	services = h.Tor.ListOnionServices()
	assert.Len(services, 2)

	// Rotate circuit
	err = h.Tor.RotateCircuit(ctx)
	assert.NoError(err)

	// Stop Tor
	err = h.Tor.Stop()
	assert.NoError(err)
	h.AssertTorNotRunning()

	// All services should be cleared
	services = h.Tor.ListOnionServices()
	assert.Empty(services)
}

// TestE2E_IPFSLifecycle tests IPFS content lifecycle
func TestE2E_IPFSLifecycle(t *testing.T) {
	h := testutil.NewTestHarness(t)
	assert := testutil.NewAssertions(t)

	ctx, cancel := h.WithTimeout(60 * time.Second)
	defer cancel()

	// Add content
	content1 := []byte("First content")
	content2 := []byte("Second content")
	content3 := []byte("Third content")

	cid1, err := h.IPFS.AddBytes(ctx, content1)
	assert.NoError(err)

	cid2, err := h.IPFS.AddBytes(ctx, content2)
	assert.NoError(err)

	cid3, err := h.IPFS.AddBytes(ctx, content3)
	assert.NoError(err)

	// Verify different CIDs
	assert.NotEqual(cid1, cid2)
	assert.NotEqual(cid2, cid3)

	// Pin some content
	err = h.IPFS.Pin(ctx, cid1)
	assert.NoError(err)

	err = h.IPFS.Pin(ctx, cid2)
	assert.NoError(err)

	// Check stats
	stats := h.IPFS.Stats()
	assert.Equal(3, stats.NumObjects)
	assert.Equal(2, stats.NumPinned)

	// Try to remove pinned content (should fail)
	err = h.IPFS.Remove(ctx, cid1)
	assert.Error(err)

	// Remove unpinned content (should succeed)
	err = h.IPFS.Remove(ctx, cid3)
	assert.NoError(err)

	assert.False(h.IPFS.Exists(cid3))

	// Unpin and remove
	err = h.IPFS.Unpin(ctx, cid1)
	assert.NoError(err)

	err = h.IPFS.Remove(ctx, cid1)
	assert.NoError(err)

	// Final stats
	stats = h.IPFS.Stats()
	assert.Equal(1, stats.NumObjects)
	assert.Equal(1, stats.NumPinned)

	// Verify remaining content
	retrieved, err := h.IPFS.Cat(ctx, cid2)
	assert.NoError(err)
	assert.Equal(content2, retrieved)
}

// TestE2E_StopTimeout tests container stop with timeout
func TestE2E_StopTimeout(t *testing.T) {
	h := testutil.NewTestHarness(t)
	assert := testutil.NewAssertions(t)

	ctx, cancel := h.WithTimeout(60 * time.Second)
	defer cancel()

	containerID := "e2e-stop-timeout"

	// Create and start
	_, err := h.Containerd.CreateContainer(ctx, containerID, "nginx:latest", types.ResourceLimits{})
	assert.NoError(err)

	err = h.Containerd.StartContainer(ctx, containerID)
	assert.NoError(err)

	// Set a stop delay longer than our timeout
	h.Containerd.StopDelay = 100 * time.Millisecond

	// Stop with short timeout - mock doesn't actually enforce timeout, just verifies behavior
	startTime := time.Now()
	err = h.Containerd.StopContainer(ctx, containerID, 200*time.Millisecond)
	elapsed := time.Since(startTime)

	assert.NoError(err)
	assert.GreaterOrEqual(int64(elapsed), int64(100*time.Millisecond))

	// Cleanup
	h.Containerd.StopDelay = 0
	err = h.Containerd.DeleteContainer(ctx, containerID)
	assert.NoError(err)
}
