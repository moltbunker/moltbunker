//go:build e2e

package scenarios

import (
	"context"
	"testing"
	"time"

	"github.com/moltbunker/moltbunker/pkg/types"
	"github.com/moltbunker/moltbunker/tests/e2e/testutil"
)

// TestE2E_DeploySimpleContainer tests deploying a simple container
func TestE2E_DeploySimpleContainer(t *testing.T) {
	h := testutil.NewTestHarness(t)
	assert := testutil.NewAssertions(t)

	ctx, cancel := h.WithTimeout(60 * time.Second)
	defer cancel()

	// Simulate deploying a container
	containerID := "e2e-test-deploy-simple"
	imageRef := "docker.io/library/nginx:latest"
	resources := types.ResourceLimits{
		CPUQuota:    100000,
		CPUPeriod:   100000,
		MemoryLimit: 256 * 1024 * 1024, // 256MB
		DiskLimit:   1 * 1024 * 1024 * 1024, // 1GB
		PIDLimit:    100,
	}

	// Create container
	container, err := h.Containerd.CreateContainer(ctx, containerID, imageRef, resources)
	assert.NoError(err, "Container creation should succeed")
	assert.Equal(containerID, container.ID)
	assert.Equal(imageRef, container.Image)
	assert.Equal(types.ContainerStatusCreated, container.Status)

	// Start container
	err = h.Containerd.StartContainer(ctx, containerID)
	assert.NoError(err, "Container start should succeed")

	// Verify container is running
	err = h.WaitForContainer(containerID, string(types.ContainerStatusRunning), 10*time.Second)
	assert.NoError(err, "Container should be running")

	// Verify through GetContainerStatus
	status, err := h.Containerd.GetContainerStatus(ctx, containerID)
	assert.NoError(err)
	assert.Equal(types.ContainerStatusRunning, status)

	// Cleanup
	err = h.Containerd.StopContainer(ctx, containerID, 10*time.Second)
	assert.NoError(err)

	err = h.Containerd.DeleteContainer(ctx, containerID)
	assert.NoError(err)
}

// TestE2E_DeployWithEncryption tests deploying an encrypted container
func TestE2E_DeployWithEncryption(t *testing.T) {
	h := testutil.NewTestHarness(t)
	assert := testutil.NewAssertions(t)

	ctx, cancel := h.WithTimeout(60 * time.Second)
	defer cancel()

	containerID := "e2e-test-deploy-encrypted"
	imageRef := "docker.io/library/alpine:latest"

	// Create container
	_, err := h.Containerd.CreateContainer(ctx, containerID, imageRef, types.ResourceLimits{
		MemoryLimit: 128 * 1024 * 1024,
	})
	assert.NoError(err)

	// Simulate setting up encrypted volume
	err = h.Containerd.SetEncryptedVolume(containerID, "/dev/mapper/moltbunker-"+containerID)
	assert.NoError(err)

	container, exists := h.Containerd.GetContainer(containerID)
	assert.True(exists)
	assert.NotEmpty(container.EncryptedVolume, "Encrypted volume should be set")

	// Start container
	err = h.Containerd.StartContainer(ctx, containerID)
	assert.NoError(err)

	// Verify container is running
	status, err := h.Containerd.GetContainerStatus(ctx, containerID)
	assert.NoError(err)
	assert.Equal(types.ContainerStatusRunning, status)

	// Cleanup
	err = h.Containerd.DeleteContainer(ctx, containerID)
	assert.NoError(err)
}

// TestE2E_DeployWithTor tests deploying a container with Tor onion service
func TestE2E_DeployWithTor(t *testing.T) {
	h := testutil.NewTestHarness(t)
	assert := testutil.NewAssertions(t)

	ctx, cancel := h.WithTimeout(60 * time.Second)
	defer cancel()

	containerID := "e2e-test-deploy-tor"
	imageRef := "docker.io/library/nginx:latest"
	containerPort := 8080

	// Start Tor
	err := h.Tor.Start(ctx)
	assert.NoError(err)
	h.AssertTorRunning()

	// Create container
	_, err = h.Containerd.CreateContainer(ctx, containerID, imageRef, types.ResourceLimits{
		MemoryLimit: 256 * 1024 * 1024,
	})
	assert.NoError(err)

	// Create onion service for container
	onionAddr, err := h.Tor.CreateOnionService(ctx, containerPort)
	assert.NoError(err)
	assert.Contains(onionAddr, ".onion")

	// Associate onion address with container
	err = h.Containerd.SetOnionAddress(containerID, onionAddr)
	assert.NoError(err)

	// Start container
	err = h.Containerd.StartContainer(ctx, containerID)
	assert.NoError(err)

	// Verify container has onion address
	container, exists := h.Containerd.GetContainer(containerID)
	assert.True(exists)
	assert.Equal(onionAddr, container.OnionAddress)

	// Verify Tor services
	services := h.Tor.ListOnionServices()
	assert.Len(services, 1, "Should have one onion service")

	// Cleanup
	err = h.Containerd.DeleteContainer(ctx, containerID)
	assert.NoError(err)

	err = h.Tor.CloseOnionService(containerPort)
	assert.NoError(err)

	err = h.Tor.Stop()
	assert.NoError(err)
}

// TestE2E_DeployMultipleContainers tests deploying multiple containers
func TestE2E_DeployMultipleContainers(t *testing.T) {
	h := testutil.NewTestHarness(t)
	assert := testutil.NewAssertions(t)

	ctx, cancel := h.WithTimeout(60 * time.Second)
	defer cancel()

	containers := []struct {
		id    string
		image string
	}{
		{"e2e-multi-1", "nginx:latest"},
		{"e2e-multi-2", "alpine:latest"},
		{"e2e-multi-3", "redis:latest"},
	}

	// Create and start all containers
	for _, c := range containers {
		_, err := h.Containerd.CreateContainer(ctx, c.id, c.image, types.ResourceLimits{
			MemoryLimit: 128 * 1024 * 1024,
		})
		assert.NoError(err, "Create "+c.id)

		err = h.Containerd.StartContainer(ctx, c.id)
		assert.NoError(err, "Start "+c.id)
	}

	// Verify all containers are running
	allContainers := h.Containerd.ListContainers()
	assert.Len(allContainers, 3, "Should have 3 containers")

	for _, c := range containers {
		status, err := h.Containerd.GetContainerStatus(ctx, c.id)
		assert.NoError(err)
		assert.Equal(types.ContainerStatusRunning, status, c.id+" should be running")
	}

	// Stop and delete all containers
	for _, c := range containers {
		err := h.Containerd.StopContainer(ctx, c.id, 5*time.Second)
		assert.NoError(err)

		err = h.Containerd.DeleteContainer(ctx, c.id)
		assert.NoError(err)
	}

	// Verify all containers are gone
	allContainers = h.Containerd.ListContainers()
	assert.Empty(allContainers, "All containers should be deleted")
}

// TestE2E_DeployWithResourceLimits tests container resource limits
func TestE2E_DeployWithResourceLimits(t *testing.T) {
	h := testutil.NewTestHarness(t)
	assert := testutil.NewAssertions(t)

	ctx, cancel := h.WithTimeout(60 * time.Second)
	defer cancel()

	containerID := "e2e-resource-limits"
	resources := types.ResourceLimits{
		CPUQuota:    50000,  // 50% CPU
		CPUPeriod:   100000,
		MemoryLimit: 512 * 1024 * 1024, // 512MB
		DiskLimit:   2 * 1024 * 1024 * 1024, // 2GB
		NetworkBW:   10 * 1024 * 1024, // 10MB/s
		PIDLimit:    50,
	}

	// Create container with limits
	container, err := h.Containerd.CreateContainer(ctx, containerID, "nginx:latest", resources)
	assert.NoError(err)

	// Verify resources were set
	assert.Equal(resources.CPUQuota, container.Resources.CPUQuota)
	assert.Equal(resources.MemoryLimit, container.Resources.MemoryLimit)
	assert.Equal(resources.PIDLimit, container.Resources.PIDLimit)

	// Start and verify
	err = h.Containerd.StartContainer(ctx, containerID)
	assert.NoError(err)

	// Check health status
	health, err := h.Containerd.GetHealthStatus(ctx, containerID)
	assert.NoError(err)
	assert.True(health.Healthy)

	// Cleanup
	err = h.Containerd.DeleteContainer(ctx, containerID)
	assert.NoError(err)
}

// TestE2E_DeployFailure tests handling of deployment failures
func TestE2E_DeployFailure(t *testing.T) {
	h := testutil.NewTestHarness(t)
	assert := testutil.NewAssertions(t)

	ctx, cancel := h.WithTimeout(60 * time.Second)
	defer cancel()

	containerID := "e2e-deploy-failure"

	// Inject a pull error to simulate image not found
	h.Containerd.SetPullError(context.DeadlineExceeded)
	defer h.Containerd.SetPullError(nil)

	// Attempt to create container (should still work with mock)
	_, err := h.Containerd.CreateContainer(ctx, containerID, "nonexistent:latest", types.ResourceLimits{})
	// Mock doesn't use pull, so this should succeed
	assert.NoError(err)

	// Inject a start error
	h.Containerd.SetStartError(context.DeadlineExceeded)
	defer h.Containerd.SetStartError(nil)

	// Start should fail
	err = h.Containerd.StartContainer(ctx, containerID)
	assert.Error(err, "Start should fail with injected error")

	// Clear error and retry
	h.Containerd.SetStartError(nil)
	err = h.Containerd.StartContainer(ctx, containerID)
	assert.NoError(err, "Start should succeed after clearing error")

	// Cleanup
	err = h.Containerd.DeleteContainer(ctx, containerID)
	assert.NoError(err)
}

// TestE2E_DeployWithIPFSImage tests deploying using IPFS-cached image
func TestE2E_DeployWithIPFSImage(t *testing.T) {
	h := testutil.NewTestHarness(t)
	assert := testutil.NewAssertions(t)

	ctx, cancel := h.WithTimeout(60 * time.Second)
	defer cancel()

	// Simulate storing image in IPFS
	imageContent := []byte("mock container image content")
	imageCID, err := h.IPFS.AddBytes(ctx, imageContent)
	assert.NoError(err)

	// Pin the image
	err = h.IPFS.Pin(ctx, imageCID)
	assert.NoError(err)
	assert.True(h.IPFS.IsPinned(imageCID))

	// Deploy container referencing IPFS CID
	containerID := "e2e-ipfs-deploy"
	imageRef := "ipfs://" + imageCID

	_, err = h.Containerd.CreateContainer(ctx, containerID, imageRef, types.ResourceLimits{})
	assert.NoError(err)

	err = h.Containerd.StartContainer(ctx, containerID)
	assert.NoError(err)

	// Verify IPFS content is still accessible
	assert.True(h.IPFS.Exists(imageCID))

	// Cleanup
	err = h.Containerd.DeleteContainer(ctx, containerID)
	assert.NoError(err)
}
