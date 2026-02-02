//go:build e2e

package smoke

import (
	"context"
	"testing"
	"time"

	"github.com/moltbunker/moltbunker/pkg/types"
	"github.com/moltbunker/moltbunker/tests/e2e/testutil"
)

// TestSmoke_MockContainerdBasics tests basic mock containerd functionality
func TestSmoke_MockContainerdBasics(t *testing.T) {
	h := testutil.NewTestHarness(t)
	assert := testutil.NewAssertions(t)

	ctx, cancel := h.WithTimeout(30 * time.Second)
	defer cancel()

	// Test container creation
	containerID := "smoke-test-container"
	imageRef := "docker.io/library/nginx:latest"

	container, err := h.Containerd.CreateContainer(ctx, containerID, imageRef, types.ResourceLimits{
		CPUQuota:    100000,
		CPUPeriod:   100000,
		MemoryLimit: 256 * 1024 * 1024,
		PIDLimit:    100,
	})
	assert.NoError(err, "CreateContainer should succeed")
	assert.NotNil(container, "Container should not be nil")
	assert.Equal(types.ContainerStatusCreated, container.Status, "Container should be created")

	// Test container start
	err = h.Containerd.StartContainer(ctx, containerID)
	assert.NoError(err, "StartContainer should succeed")

	status, err := h.Containerd.GetContainerStatus(ctx, containerID)
	assert.NoError(err, "GetContainerStatus should succeed")
	assert.Equal(types.ContainerStatusRunning, status, "Container should be running")

	// Test container stop
	err = h.Containerd.StopContainer(ctx, containerID, 10*time.Second)
	assert.NoError(err, "StopContainer should succeed")

	status, err = h.Containerd.GetContainerStatus(ctx, containerID)
	assert.NoError(err, "GetContainerStatus should succeed")
	assert.Equal(types.ContainerStatusStopped, status, "Container should be stopped")

	// Test container delete
	err = h.Containerd.DeleteContainer(ctx, containerID)
	assert.NoError(err, "DeleteContainer should succeed")

	_, exists := h.Containerd.GetContainer(containerID)
	assert.False(exists, "Container should not exist after deletion")
}

// TestSmoke_MockTorBasics tests basic mock Tor functionality
func TestSmoke_MockTorBasics(t *testing.T) {
	h := testutil.NewTestHarness(t)
	assert := testutil.NewAssertions(t)

	ctx, cancel := h.WithTimeout(30 * time.Second)
	defer cancel()

	// Test Tor start
	err := h.Tor.Start(ctx)
	assert.NoError(err, "Tor start should succeed")
	assert.True(h.Tor.IsRunning(), "Tor should be running")

	// Test onion service creation
	addr, err := h.Tor.CreateOnionService(ctx, 8080)
	assert.NoError(err, "CreateOnionService should succeed")
	assert.NotEmpty(addr, "Onion address should not be empty")
	assert.Contains(addr, ".onion", "Address should be an onion address")

	// Test circuit rotation
	err = h.Tor.RotateCircuit(ctx)
	assert.NoError(err, "RotateCircuit should succeed")

	// Test circuit info
	circuits, err := h.Tor.GetCircuitInfo(ctx)
	assert.NoError(err, "GetCircuitInfo should succeed")
	assert.NotEmpty(circuits, "Should have circuit info")

	// Test Tor stop
	err = h.Tor.Stop()
	assert.NoError(err, "Tor stop should succeed")
	assert.False(h.Tor.IsRunning(), "Tor should not be running")
}

// TestSmoke_MockIPFSBasics tests basic mock IPFS functionality
func TestSmoke_MockIPFSBasics(t *testing.T) {
	h := testutil.NewTestHarness(t)
	assert := testutil.NewAssertions(t)

	ctx, cancel := h.WithTimeout(30 * time.Second)
	defer cancel()

	// Test adding content
	content := []byte("Hello, IPFS!")
	cid, err := h.IPFS.AddBytes(ctx, content)
	assert.NoError(err, "Add should succeed")
	assert.NotEmpty(cid, "CID should not be empty")

	// Test getting content
	retrieved, err := h.IPFS.Cat(ctx, cid)
	assert.NoError(err, "Cat should succeed")
	assert.Equal(string(content), string(retrieved), "Content should match")

	// Test pinning
	err = h.IPFS.Pin(ctx, cid)
	assert.NoError(err, "Pin should succeed")
	assert.True(h.IPFS.IsPinned(cid), "Content should be pinned")

	// Test unpinning
	err = h.IPFS.Unpin(ctx, cid)
	assert.NoError(err, "Unpin should succeed")
	assert.False(h.IPFS.IsPinned(cid), "Content should not be pinned")

	// Test stats
	stats := h.IPFS.Stats()
	assert.Equal(1, stats.NumObjects, "Should have 1 object")
}

// TestSmoke_TestHarnessSetup tests that the test harness sets up correctly
func TestSmoke_TestHarnessSetup(t *testing.T) {
	h := testutil.NewTestHarness(t)
	assert := testutil.NewAssertions(t)

	// Check directories were created
	assert.NotEmpty(h.TempDir, "TempDir should not be empty")
	assert.NotEmpty(h.DataDir, "DataDir should not be empty")
	assert.NotEmpty(h.LogsDir, "LogsDir should not be empty")
	assert.NotEmpty(h.SocketPath, "SocketPath should not be empty")

	// Check mocks were initialized
	assert.NotNil(h.Containerd, "Containerd mock should be initialized")
	assert.NotNil(h.Tor, "Tor mock should be initialized")
	assert.NotNil(h.IPFS, "IPFS mock should be initialized")

	// Test config creation
	err := h.CreateTestConfig()
	assert.NoError(err, "CreateTestConfig should succeed")
}

// TestSmoke_ContainerLogs tests container log functionality
func TestSmoke_ContainerLogs(t *testing.T) {
	h := testutil.NewTestHarness(t)
	assert := testutil.NewAssertions(t)

	ctx, cancel := h.WithTimeout(30 * time.Second)
	defer cancel()

	containerID := "log-test-container"

	// Create and start container
	_, err := h.Containerd.CreateContainer(ctx, containerID, "nginx:latest", types.ResourceLimits{})
	assert.NoError(err)

	err = h.Containerd.StartContainer(ctx, containerID)
	assert.NoError(err)

	// Add some log lines
	logLines := []string{
		"Container started",
		"Processing request",
		"Request completed",
	}
	for _, line := range logLines {
		err := h.Containerd.AppendLog(containerID, line)
		assert.NoError(err)
	}

	// Get logs
	reader, err := h.Containerd.GetContainerLogs(ctx, containerID, false, 0)
	assert.NoError(err)
	defer reader.Close()

	buf := make([]byte, 1024)
	n, _ := reader.Read(buf)
	logs := string(buf[:n])

	for _, line := range logLines {
		assert.Contains(logs, line, "Logs should contain: "+line)
	}
}

// TestSmoke_ErrorInjection tests that error injection works
func TestSmoke_ErrorInjection(t *testing.T) {
	h := testutil.NewTestHarness(t)
	assert := testutil.NewAssertions(t)

	ctx, cancel := h.WithTimeout(30 * time.Second)
	defer cancel()

	// Inject create error
	expectedErr := context.DeadlineExceeded
	h.Containerd.SetCreateError(expectedErr)

	_, err := h.Containerd.CreateContainer(ctx, "test", "nginx:latest", types.ResourceLimits{})
	assert.Error(err, "Should return error")

	// Clear error
	h.Containerd.SetCreateError(nil)

	_, err = h.Containerd.CreateContainer(ctx, "test", "nginx:latest", types.ResourceLimits{})
	assert.NoError(err, "Should succeed after clearing error")
}

// TestSmoke_WaitFor tests the WaitFor helper
func TestSmoke_WaitFor(t *testing.T) {
	h := testutil.NewTestHarness(t)
	assert := testutil.NewAssertions(t)

	// Test successful wait
	counter := 0
	err := h.WaitFor(func() bool {
		counter++
		return counter >= 3
	}, 1*time.Second, "counter to reach 3")
	assert.NoError(err, "WaitFor should succeed")

	// Test timeout
	err = h.WaitFor(func() bool {
		return false
	}, 100*time.Millisecond, "impossible condition")
	assert.Error(err, "WaitFor should timeout")
	assert.Contains(err.Error(), "timeout", "Error should mention timeout")
}

// TestSmoke_MethodCallTracking tests that method calls are tracked
func TestSmoke_MethodCallTracking(t *testing.T) {
	h := testutil.NewTestHarness(t)
	assert := testutil.NewAssertions(t)

	ctx, cancel := h.WithTimeout(30 * time.Second)
	defer cancel()

	// Perform some operations
	_, _ = h.Containerd.CreateContainer(ctx, "test1", "nginx", types.ResourceLimits{})
	_ = h.Containerd.StartContainer(ctx, "test1")
	_, _ = h.Containerd.GetContainerStatus(ctx, "test1")

	// Check call history
	calls := h.Containerd.GetCalls()
	assert.GreaterOrEqual(len(calls), 3, "Should have at least 3 calls")

	// Check call methods
	methods := make([]string, len(calls))
	for i, call := range calls {
		methods[i] = call.Method
	}
	assert.Contains(methods[0], "CreateContainer", "First call should be CreateContainer")
	assert.Contains(methods[1], "StartContainer", "Second call should be StartContainer")

	// Clear calls
	h.Containerd.ClearCalls()
	calls = h.Containerd.GetCalls()
	assert.Empty(calls, "Calls should be empty after clearing")
}
