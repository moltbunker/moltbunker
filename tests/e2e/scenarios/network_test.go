//go:build e2e

package scenarios

import (
	"testing"
	"time"

	"github.com/moltbunker/moltbunker/pkg/types"
	"github.com/moltbunker/moltbunker/tests/e2e/testutil"
)

// TestE2E_TorIntegration tests Tor integration with containers
func TestE2E_TorIntegration(t *testing.T) {
	h := testutil.NewTestHarness(t)
	assert := testutil.NewAssertions(t)

	ctx, cancel := h.WithTimeout(60 * time.Second)
	defer cancel()

	// Start Tor
	err := h.Tor.Start(ctx)
	assert.NoError(err)
	h.AssertTorRunning()

	// Create multiple containers with Tor
	containers := []struct {
		id   string
		port int
	}{
		{"tor-web-1", 8080},
		{"tor-web-2", 8081},
		{"tor-api", 9000},
	}

	for _, c := range containers {
		// Create container
		_, err := h.Containerd.CreateContainer(ctx, c.id, "nginx:latest", types.ResourceLimits{
			MemoryLimit: 128 * 1024 * 1024,
		})
		assert.NoError(err)

		// Create onion service
		onionAddr, err := h.Tor.CreateOnionService(ctx, c.port)
		assert.NoError(err)

		// Associate with container
		err = h.Containerd.SetOnionAddress(c.id, onionAddr)
		assert.NoError(err)

		// Start container
		err = h.Containerd.StartContainer(ctx, c.id)
		assert.NoError(err)

		// Verify onion address is set
		container, exists := h.Containerd.GetContainer(c.id)
		assert.True(exists)
		assert.Equal(onionAddr, container.OnionAddress)
	}

	// Verify all onion services
	services := h.Tor.ListOnionServices()
	assert.Len(services, 3)

	// Test circuit rotation
	err = h.Tor.RotateCircuit(ctx)
	assert.NoError(err)

	// Get circuit info
	circuits, err := h.Tor.GetCircuitInfo(ctx)
	assert.NoError(err)
	assert.NotEmpty(circuits)

	// Cleanup containers
	for _, c := range containers {
		err := h.Containerd.DeleteContainer(ctx, c.id)
		assert.NoError(err)

		err = h.Tor.CloseOnionService(c.port)
		assert.NoError(err)
	}

	// Stop Tor
	err = h.Tor.Stop()
	assert.NoError(err)
	h.AssertTorNotRunning()
}

// TestE2E_TorServicePersistence tests that onion addresses persist
func TestE2E_TorServicePersistence(t *testing.T) {
	h := testutil.NewTestHarness(t)
	assert := testutil.NewAssertions(t)

	ctx, cancel := h.WithTimeout(60 * time.Second)
	defer cancel()

	// Start Tor
	err := h.Tor.Start(ctx)
	assert.NoError(err)

	// Create onion service
	port := 8080
	addr1, err := h.Tor.CreateOnionService(ctx, port)
	assert.NoError(err)

	// Request same port again - should return same address
	addr2, err := h.Tor.CreateOnionService(ctx, port)
	assert.NoError(err)
	assert.Equal(addr1, addr2, "Same port should return same onion address")

	// Cleanup
	err = h.Tor.Stop()
	assert.NoError(err)
}

// TestE2E_TorErrorHandling tests Tor error scenarios
func TestE2E_TorErrorHandling(t *testing.T) {
	h := testutil.NewTestHarness(t)
	assert := testutil.NewAssertions(t)

	ctx, cancel := h.WithTimeout(60 * time.Second)
	defer cancel()

	// Try to create onion service before Tor is started
	_, err := h.Tor.CreateOnionService(ctx, 8080)
	assert.Error(err)
	assert.Contains(err.Error(), "not running")

	// Try to rotate circuit before Tor is started
	err = h.Tor.RotateCircuit(ctx)
	assert.Error(err)
	assert.Contains(err.Error(), "not running")

	// Start Tor
	err = h.Tor.Start(ctx)
	assert.NoError(err)

	// Inject error
	h.Tor.SetCreateServiceError(ErrMockInjected)

	_, err = h.Tor.CreateOnionService(ctx, 8080)
	assert.Error(err)

	// Clear error
	h.Tor.SetCreateServiceError(nil)

	_, err = h.Tor.CreateOnionService(ctx, 8080)
	assert.NoError(err)

	// Cleanup
	err = h.Tor.Stop()
	assert.NoError(err)
}

// TestE2E_IPFSIntegration tests IPFS integration
func TestE2E_IPFSIntegration(t *testing.T) {
	h := testutil.NewTestHarness(t)
	assert := testutil.NewAssertions(t)

	ctx, cancel := h.WithTimeout(60 * time.Second)
	defer cancel()

	// Store container image in IPFS
	imageData := []byte("mock container image layer data")
	imageCID, err := h.IPFS.AddBytes(ctx, imageData)
	assert.NoError(err)

	// Pin for persistence
	err = h.IPFS.Pin(ctx, imageCID)
	assert.NoError(err)

	// Deploy container referencing IPFS image
	containerID := "ipfs-container"
	imageRef := "ipfs://" + imageCID

	_, err = h.Containerd.CreateContainer(ctx, containerID, imageRef, types.ResourceLimits{})
	assert.NoError(err)

	err = h.Containerd.StartContainer(ctx, containerID)
	assert.NoError(err)

	// Store container data in IPFS
	containerData := []byte("container output data")
	dataCID, err := h.IPFS.AddBytes(ctx, containerData)
	assert.NoError(err)

	// Verify both CIDs exist
	assert.True(h.IPFS.Exists(imageCID))
	assert.True(h.IPFS.Exists(dataCID))

	// Verify pinned status
	assert.True(h.IPFS.IsPinned(imageCID))
	assert.False(h.IPFS.IsPinned(dataCID))

	// Cleanup
	err = h.Containerd.DeleteContainer(ctx, containerID)
	assert.NoError(err)
}

// TestE2E_ConcurrentContainers tests concurrent container operations
func TestE2E_ConcurrentContainers(t *testing.T) {
	h := testutil.NewTestHarness(t)
	assert := testutil.NewAssertions(t)

	ctx, cancel := h.WithTimeout(120 * time.Second)
	defer cancel()

	numContainers := 10
	containerIDs := make([]string, numContainers)

	// Create all containers
	for i := 0; i < numContainers; i++ {
		containerIDs[i] = "concurrent-" + string(rune('a'+i))
		_, err := h.Containerd.CreateContainer(ctx, containerIDs[i], "nginx:latest", types.ResourceLimits{
			MemoryLimit: 64 * 1024 * 1024,
		})
		assert.NoError(err)
	}

	// Start all containers
	for _, id := range containerIDs {
		err := h.Containerd.StartContainer(ctx, id)
		assert.NoError(err)
	}

	// Verify all running
	containers := h.Containerd.ListContainers()
	assert.Len(containers, numContainers)

	for _, id := range containerIDs {
		status, err := h.Containerd.GetContainerStatus(ctx, id)
		assert.NoError(err)
		assert.Equal(types.ContainerStatusRunning, status)
	}

	// Stop all containers
	for _, id := range containerIDs {
		err := h.Containerd.StopContainer(ctx, id, 5*time.Second)
		assert.NoError(err)
	}

	// Delete all containers
	for _, id := range containerIDs {
		err := h.Containerd.DeleteContainer(ctx, id)
		assert.NoError(err)
	}

	// Verify all deleted
	containers = h.Containerd.ListContainers()
	assert.Empty(containers)
}

// TestE2E_NetworkModes tests different network configurations
func TestE2E_NetworkModes(t *testing.T) {
	h := testutil.NewTestHarness(t)
	assert := testutil.NewAssertions(t)

	ctx, cancel := h.WithTimeout(60 * time.Second)
	defer cancel()

	// Test cases for different network modes
	testCases := []struct {
		name    string
		useTor  bool
		mode    string
	}{
		{"clearnet", false, "clearnet"},
		{"tor-only", true, "tor_only"},
		{"hybrid", true, "hybrid"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			containerID := "network-" + tc.mode

			if tc.useTor {
				err := h.Tor.Start(ctx)
				assert.NoError(err)
			}

			// Create container
			_, err := h.Containerd.CreateContainer(ctx, containerID, "nginx:latest", types.ResourceLimits{})
			assert.NoError(err)

			if tc.useTor {
				// Create onion service
				addr, err := h.Tor.CreateOnionService(ctx, 8080)
				assert.NoError(err)

				err = h.Containerd.SetOnionAddress(containerID, addr)
				assert.NoError(err)
			}

			err = h.Containerd.StartContainer(ctx, containerID)
			assert.NoError(err)

			// Verify configuration
			container, exists := h.Containerd.GetContainer(containerID)
			assert.True(exists)

			if tc.useTor {
				assert.NotEmpty(container.OnionAddress)
			} else {
				assert.Empty(container.OnionAddress)
			}

			// Cleanup
			err = h.Containerd.DeleteContainer(ctx, containerID)
			assert.NoError(err)

			if tc.useTor {
				err = h.Tor.Stop()
				assert.NoError(err)
			}
		})
	}
}

// ErrMockInjected is used for testing error injection
var ErrMockInjected = &mockError{msg: "injected error for testing"}

type mockError struct {
	msg string
}

func (e *mockError) Error() string {
	return e.msg
}
