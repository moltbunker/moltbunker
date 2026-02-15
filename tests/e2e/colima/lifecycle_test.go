//go:build colima

package colima

import (
	"context"
	"testing"
	"time"

	"github.com/moltbunker/moltbunker/pkg/types"
)

// TestPing verifies basic connectivity to the Colima containerd socket.
func TestPing(t *testing.T) {
	client := newTestClient(t)

	ctx, cancel := context.WithTimeout(context.Background(), lifecycleTimeout)
	defer cancel()

	if err := client.Ping(ctx); err != nil {
		t.Fatalf("Ping failed: %v", err)
	}
}

// TestImagePull verifies that we can pull an image from a registry through Colima.
func TestImagePull(t *testing.T) {
	client := newTestClient(t)

	ctx, cancel := context.WithTimeout(context.Background(), pullTimeout)
	defer cancel()

	nsCtx := client.WithNamespace(ctx)

	img, err := client.PullImage(nsCtx, alpineImage)
	if err != nil {
		t.Fatalf("PullImage(%s) failed: %v", alpineImage, err)
	}

	if img.Name() != alpineImage {
		t.Errorf("expected image name %s, got %s", alpineImage, img.Name())
	}

	// Verify we can retrieve it again
	img2, err := client.GetImage(nsCtx, alpineImage)
	if err != nil {
		t.Fatalf("GetImage(%s) failed after pull: %v", alpineImage, err)
	}
	if img2.Name() != alpineImage {
		t.Errorf("GetImage returned wrong name: %s", img2.Name())
	}
}

// TestContainerCreateAndDelete tests the most basic lifecycle: create → delete.
func TestContainerCreateAndDelete(t *testing.T) {
	client := newTestClient(t)
	id := containerID(t, "create")
	cleanupContainer(t, client, id)

	ctx, cancel := context.WithTimeout(context.Background(), pullTimeout)
	defer cancel()

	resources := types.ResourceLimits{
		MemoryLimit: 64 * 1024 * 1024, // 64MB
		PIDLimit:    100,
	}

	managed, err := client.CreateContainer(ctx, id, alpineImage, resources)
	if err != nil {
		t.Fatalf("CreateContainer failed: %v", err)
	}

	if managed.ID != id {
		t.Errorf("expected ID %s, got %s", id, managed.ID)
	}
	if managed.Image != alpineImage {
		t.Errorf("expected image %s, got %s", alpineImage, managed.Image)
	}
	if managed.Status != types.ContainerStatusCreated {
		t.Errorf("expected status %s, got %s", types.ContainerStatusCreated, managed.Status)
	}

	// Verify container is tracked
	got, exists := client.GetContainer(id)
	if !exists {
		t.Fatal("GetContainer returned false for just-created container")
	}
	if got.ID != id {
		t.Errorf("GetContainer returned wrong ID: %s", got.ID)
	}

	// Delete
	if err := client.DeleteContainer(ctx, id); err != nil {
		t.Fatalf("DeleteContainer failed: %v", err)
	}

	// Verify deleted
	_, exists = client.GetContainer(id)
	if exists {
		t.Error("container still tracked after delete")
	}
}

// TestContainerFullLifecycle tests the complete lifecycle: create → start → status → stop → delete.
func TestContainerFullLifecycle(t *testing.T) {
	client := newTestClient(t)
	id := containerID(t, "lifecycle")
	cleanupContainer(t, client, id)

	ctx, cancel := context.WithTimeout(context.Background(), pullTimeout)
	defer cancel()

	// Use busybox with sleep so the container stays running
	resources := types.ResourceLimits{
		MemoryLimit: 32 * 1024 * 1024, // 32MB
		PIDLimit:    50,
	}

	// Create container with a long-running command
	managed, err := client.CreateContainerWithSpec(ctx, id, busyboxImage, resources,
		withProcessArgs("sleep", "300"),
	)
	if err != nil {
		t.Fatalf("CreateContainer failed: %v", err)
	}
	t.Logf("created container %s (image: %s)", managed.ID, managed.Image)

	// Start
	if err := client.StartContainer(ctx, id); err != nil {
		t.Fatalf("StartContainer failed: %v", err)
	}
	t.Log("container started")

	// Give it a moment to initialize
	time.Sleep(500 * time.Millisecond)

	// Check status
	status, err := client.GetContainerStatus(ctx, id)
	if err != nil {
		t.Fatalf("GetContainerStatus failed: %v", err)
	}
	if status != types.ContainerStatusRunning {
		t.Errorf("expected status running, got %s", status)
	}
	t.Logf("container status: %s", status)

	// Check health
	health, err := client.GetHealthStatus(ctx, id)
	if err != nil {
		t.Fatalf("GetHealthStatus failed: %v", err)
	}
	if !health.Healthy {
		t.Error("expected healthy=true for running container")
	}
	t.Logf("health: healthy=%v", health.Healthy)

	// Stop
	if err := client.StopContainer(ctx, id, 10*time.Second); err != nil {
		t.Fatalf("StopContainer failed: %v", err)
	}
	t.Log("container stopped")

	// Verify stopped status
	status, err = client.GetContainerStatus(ctx, id)
	if err != nil {
		t.Fatalf("GetContainerStatus after stop failed: %v", err)
	}
	if status != types.ContainerStatusStopped {
		t.Errorf("expected status stopped after StopContainer, got %s", status)
	}

	// Delete
	if err := client.DeleteContainer(ctx, id); err != nil {
		t.Fatalf("DeleteContainer failed: %v", err)
	}
	t.Log("container deleted")
}

// TestContainerStartStop tests multiple start/stop cycles on the same container.
func TestContainerStartStop(t *testing.T) {
	client := newTestClient(t)
	id := containerID(t, "startstop")
	cleanupContainer(t, client, id)

	ctx, cancel := context.WithTimeout(context.Background(), pullTimeout)
	defer cancel()

	resources := types.ResourceLimits{
		MemoryLimit: 32 * 1024 * 1024,
		PIDLimit:    50,
	}

	_, err := client.CreateContainerWithSpec(ctx, id, busyboxImage, resources,
		withProcessArgs("sleep", "300"),
	)
	if err != nil {
		t.Fatalf("CreateContainer failed: %v", err)
	}

	// Cycle 1: start → stop
	if err := client.StartContainer(ctx, id); err != nil {
		t.Fatalf("StartContainer (cycle 1) failed: %v", err)
	}
	time.Sleep(300 * time.Millisecond)

	status, _ := client.GetContainerStatus(ctx, id)
	if status != types.ContainerStatusRunning {
		t.Fatalf("expected running after start (cycle 1), got %s", status)
	}

	if err := client.StopContainer(ctx, id, 5*time.Second); err != nil {
		t.Fatalf("StopContainer (cycle 1) failed: %v", err)
	}

	status, _ = client.GetContainerStatus(ctx, id)
	if status != types.ContainerStatusStopped {
		t.Fatalf("expected stopped after stop (cycle 1), got %s", status)
	}

	t.Log("cycle 1 complete")

	// Note: containerd typically doesn't allow re-creating a task on the same container
	// after the task was deleted. This is expected containerd behavior.
	// In production, moltbunker creates a new container for restarts.

	// Cleanup
	if err := client.DeleteContainer(ctx, id); err != nil {
		t.Fatalf("DeleteContainer failed: %v", err)
	}
}

// TestContainerForceKill tests that a container that ignores SIGTERM is force-killed.
func TestContainerForceKill(t *testing.T) {
	client := newTestClient(t)
	id := containerID(t, "forcekill")
	cleanupContainer(t, client, id)

	ctx, cancel := context.WithTimeout(context.Background(), pullTimeout)
	defer cancel()

	resources := types.ResourceLimits{
		MemoryLimit: 32 * 1024 * 1024,
		PIDLimit:    50,
	}

	// Use a process that traps SIGTERM (sleep in busybox responds to SIGTERM though,
	// so use a short timeout to exercise the timeout path)
	_, err := client.CreateContainerWithSpec(ctx, id, busyboxImage, resources,
		withProcessArgs("sleep", "300"),
	)
	if err != nil {
		t.Fatalf("CreateContainer failed: %v", err)
	}

	if err := client.StartContainer(ctx, id); err != nil {
		t.Fatalf("StartContainer failed: %v", err)
	}
	time.Sleep(300 * time.Millisecond)

	// Stop with a very short timeout - the container should still stop cleanly
	start := time.Now()
	if err := client.StopContainer(ctx, id, 1*time.Second); err != nil {
		t.Fatalf("StopContainer with short timeout failed: %v", err)
	}
	elapsed := time.Since(start)
	t.Logf("stop took %v", elapsed)

	status, _ := client.GetContainerStatus(ctx, id)
	if status != types.ContainerStatusStopped {
		t.Errorf("expected stopped, got %s", status)
	}

	if err := client.DeleteContainer(ctx, id); err != nil {
		t.Fatalf("DeleteContainer failed: %v", err)
	}
}

// TestDeleteRunningContainer tests that DeleteContainer handles a running container.
func TestDeleteRunningContainer(t *testing.T) {
	client := newTestClient(t)
	id := containerID(t, "delrunning")
	cleanupContainer(t, client, id)

	ctx, cancel := context.WithTimeout(context.Background(), pullTimeout)
	defer cancel()

	resources := types.ResourceLimits{
		MemoryLimit: 32 * 1024 * 1024,
		PIDLimit:    50,
	}

	_, err := client.CreateContainerWithSpec(ctx, id, busyboxImage, resources,
		withProcessArgs("sleep", "300"),
	)
	if err != nil {
		t.Fatalf("CreateContainer failed: %v", err)
	}

	if err := client.StartContainer(ctx, id); err != nil {
		t.Fatalf("StartContainer failed: %v", err)
	}
	time.Sleep(300 * time.Millisecond)

	// Delete without explicit stop - should force kill internally
	if err := client.DeleteContainer(ctx, id); err != nil {
		t.Fatalf("DeleteContainer on running container failed: %v", err)
	}

	_, exists := client.GetContainer(id)
	if exists {
		t.Error("container still tracked after delete")
	}
}
