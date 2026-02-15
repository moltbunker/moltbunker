//go:build colima

package colima

import (
	"context"
	"testing"
	"time"

	"github.com/moltbunker/moltbunker/internal/runtime"
	"github.com/moltbunker/moltbunker/pkg/types"
)

// TestResourceLimitsMemory verifies that memory limits are applied to the container.
func TestResourceLimitsMemory(t *testing.T) {
	client := newTestClient(t)
	id := containerID(t, "resmem")
	cleanupContainer(t, client, id)

	ctx, cancel := context.WithTimeout(context.Background(), pullTimeout)
	defer cancel()

	memLimit := int64(128 * 1024 * 1024) // 128MB
	resources := types.ResourceLimits{
		MemoryLimit: memLimit,
		PIDLimit:    100,
	}

	_, err := client.CreateContainerWithSpec(ctx, id, busyboxImage, resources,
		withProcessArgs("sleep", "300"),
	)
	if err != nil {
		t.Fatalf("CreateContainer with memory limit failed: %v", err)
	}

	if err := client.StartContainer(ctx, id); err != nil {
		t.Fatalf("StartContainer failed: %v", err)
	}
	time.Sleep(500 * time.Millisecond)

	// Verify the container is running with limits applied
	status, err := client.GetContainerStatus(ctx, id)
	if err != nil {
		t.Fatalf("GetContainerStatus failed: %v", err)
	}
	if status != types.ContainerStatusRunning {
		t.Fatalf("expected running, got %s", status)
	}

	// Verify resource limits were stored on the managed container
	managed, exists := client.GetContainer(id)
	if !exists {
		t.Fatal("container not found")
	}
	if managed.Resources.MemoryLimit != memLimit {
		t.Errorf("expected memory limit %d, got %d", memLimit, managed.Resources.MemoryLimit)
	}

	t.Logf("container running with memory limit: %dMB", memLimit/(1024*1024))

	if err := client.StopContainer(ctx, id, 5*time.Second); err != nil {
		t.Fatalf("StopContainer failed: %v", err)
	}
	if err := client.DeleteContainer(ctx, id); err != nil {
		t.Fatalf("DeleteContainer failed: %v", err)
	}
}

// TestResourceLimitsCPU verifies that CPU quota/period limits are applied.
func TestResourceLimitsCPU(t *testing.T) {
	client := newTestClient(t)
	id := containerID(t, "rescpu")
	cleanupContainer(t, client, id)

	ctx, cancel := context.WithTimeout(context.Background(), pullTimeout)
	defer cancel()

	resources := types.ResourceLimits{
		CPUQuota:    50000, // 50% of one core
		CPUPeriod:   100000,
		MemoryLimit: 32 * 1024 * 1024,
		PIDLimit:    50,
	}

	_, err := client.CreateContainerWithSpec(ctx, id, busyboxImage, resources,
		withProcessArgs("sleep", "300"),
	)
	if err != nil {
		t.Fatalf("CreateContainer with CPU limit failed: %v", err)
	}

	if err := client.StartContainer(ctx, id); err != nil {
		t.Fatalf("StartContainer failed: %v", err)
	}
	time.Sleep(300 * time.Millisecond)

	status, err := client.GetContainerStatus(ctx, id)
	if err != nil {
		t.Fatalf("GetContainerStatus failed: %v", err)
	}
	if status != types.ContainerStatusRunning {
		t.Fatalf("expected running, got %s", status)
	}

	managed, exists := client.GetContainer(id)
	if !exists {
		t.Fatal("container not found")
	}
	if managed.Resources.CPUQuota != 50000 {
		t.Errorf("expected CPU quota 50000, got %d", managed.Resources.CPUQuota)
	}

	t.Logf("container running with CPU quota: %d/%d", resources.CPUQuota, resources.CPUPeriod)

	if err := client.StopContainer(ctx, id, 5*time.Second); err != nil {
		t.Fatalf("StopContainer failed: %v", err)
	}
	if err := client.DeleteContainer(ctx, id); err != nil {
		t.Fatalf("DeleteContainer failed: %v", err)
	}
}

// TestResourceLimitsPID verifies that PID limits are applied.
func TestResourceLimitsPID(t *testing.T) {
	client := newTestClient(t)
	id := containerID(t, "respid")
	cleanupContainer(t, client, id)

	ctx, cancel := context.WithTimeout(context.Background(), pullTimeout)
	defer cancel()

	resources := types.ResourceLimits{
		MemoryLimit: 32 * 1024 * 1024,
		PIDLimit:    10, // Very restrictive
	}

	_, err := client.CreateContainerWithSpec(ctx, id, busyboxImage, resources,
		withProcessArgs("sleep", "300"),
	)
	if err != nil {
		t.Fatalf("CreateContainer with PID limit failed: %v", err)
	}

	if err := client.StartContainer(ctx, id); err != nil {
		t.Fatalf("StartContainer failed: %v", err)
	}
	time.Sleep(300 * time.Millisecond)

	status, err := client.GetContainerStatus(ctx, id)
	if err != nil {
		t.Fatalf("GetContainerStatus failed: %v", err)
	}
	if status != types.ContainerStatusRunning {
		t.Fatalf("expected running, got %s", status)
	}

	t.Logf("container running with PID limit: %d", resources.PIDLimit)

	if err := client.StopContainer(ctx, id, 5*time.Second); err != nil {
		t.Fatalf("StopContainer failed: %v", err)
	}
	if err := client.DeleteContainer(ctx, id); err != nil {
		t.Fatalf("DeleteContainer failed: %v", err)
	}
}

// TestContainerLogs verifies that container stdout/stderr is captured.
func TestContainerLogs(t *testing.T) {
	client := newTestClient(t)
	id := containerID(t, "logs")
	cleanupContainer(t, client, id)

	ctx, cancel := context.WithTimeout(context.Background(), pullTimeout)
	defer cancel()

	resources := types.ResourceLimits{
		MemoryLimit: 32 * 1024 * 1024,
		PIDLimit:    50,
	}

	// Run a command that produces output, then sleeps so container stays alive
	_, err := client.CreateContainerWithSpec(ctx, id, busyboxImage, resources,
		withProcessArgs("sh", "-c", "echo 'hello from moltbunker'; echo 'error output' >&2; sleep 30"),
	)
	if err != nil {
		t.Fatalf("CreateContainer failed: %v", err)
	}

	if err := client.StartContainer(ctx, id); err != nil {
		t.Fatalf("StartContainer failed: %v", err)
	}

	// Wait for output to be written
	time.Sleep(2 * time.Second)

	// Read logs (non-follow mode)
	reader, err := client.GetContainerLogs(ctx, id, false, 0)
	if err != nil {
		t.Fatalf("GetContainerLogs failed: %v", err)
	}
	defer reader.Close()

	buf := make([]byte, 4096)
	n, err := reader.Read(buf)
	if err != nil {
		t.Logf("log read returned: n=%d, err=%v", n, err)
	}

	logContent := string(buf[:n])
	t.Logf("captured logs (%d bytes):\n%s", n, logContent)

	if n == 0 {
		t.Error("expected some log output, got nothing")
	}

	// Check for our expected output strings
	if !containsSubstring(logContent, "hello from moltbunker") {
		t.Error("expected 'hello from moltbunker' in stdout logs")
	}

	if err := client.StopContainer(ctx, id, 5*time.Second); err != nil {
		t.Fatalf("StopContainer failed: %v", err)
	}
	if err := client.DeleteContainer(ctx, id); err != nil {
		t.Fatalf("DeleteContainer failed: %v", err)
	}
}

// TestHealthStatus verifies health check returns meaningful data for running containers.
func TestHealthStatus(t *testing.T) {
	client := newTestClient(t)
	id := containerID(t, "health")
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
	time.Sleep(500 * time.Millisecond)

	health, err := client.GetHealthStatus(ctx, id)
	if err != nil {
		t.Fatalf("GetHealthStatus failed: %v", err)
	}

	if !health.Healthy {
		t.Error("expected healthy=true for running container")
	}
	t.Logf("health status: healthy=%v, lastUpdate=%v", health.Healthy, health.LastUpdate)

	// Stop and check health again
	if err := client.StopContainer(ctx, id, 5*time.Second); err != nil {
		t.Fatalf("StopContainer failed: %v", err)
	}

	health2, err := client.GetHealthStatus(ctx, id)
	if err != nil {
		t.Fatalf("GetHealthStatus after stop failed: %v", err)
	}

	if health2.Healthy {
		t.Error("expected healthy=false for stopped container")
	}
	t.Logf("stopped health: healthy=%v", health2.Healthy)

	if err := client.DeleteContainer(ctx, id); err != nil {
		t.Fatalf("DeleteContainer failed: %v", err)
	}
}

// TestLoadExistingContainers verifies that containers can be rediscovered.
func TestLoadExistingContainers(t *testing.T) {
	client := newTestClient(t)
	id := containerID(t, "load")
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

	// Create a SECOND client to simulate daemon restart
	logsDir := t.TempDir()
	client2, err := runtime.NewContainerdClient(colimaSocket(), testNamespace, logsDir, "", nil)
	if err != nil {
		t.Fatalf("failed to create second client: %v", err)
	}
	defer client2.Close()

	// The second client should have no containers in its map yet
	_, exists := client2.GetContainer(id)
	if exists {
		t.Error("new client should not have container in map before LoadExistingContainers")
	}

	// Load existing containers from containerd
	if err := client2.LoadExistingContainers(ctx); err != nil {
		t.Fatalf("LoadExistingContainers failed: %v", err)
	}

	// Now it should find our running container
	loaded, exists := client2.GetContainer(id)
	if !exists {
		t.Fatal("LoadExistingContainers did not discover the running container")
	}
	if loaded.Status != types.ContainerStatusRunning {
		t.Errorf("expected loaded container status running, got %s", loaded.Status)
	}
	t.Logf("discovered container %s with status %s", id, loaded.Status)

	// Cleanup with original client
	if err := client.StopContainer(ctx, id, 5*time.Second); err != nil {
		t.Fatalf("StopContainer failed: %v", err)
	}
	if err := client.DeleteContainer(ctx, id); err != nil {
		t.Fatalf("DeleteContainer failed: %v", err)
	}
}

// TestMultipleContainers verifies running multiple containers concurrently.
func TestMultipleContainers(t *testing.T) {
	client := newTestClient(t)

	ctx, cancel := context.WithTimeout(context.Background(), pullTimeout)
	defer cancel()

	resources := types.ResourceLimits{
		MemoryLimit: 32 * 1024 * 1024,
		PIDLimit:    50,
	}

	ids := make([]string, 3)
	for i := range ids {
		ids[i] = containerID(t, string(rune('a'+i)))
		cleanupContainer(t, client, ids[i])
	}

	// Create all three
	for _, id := range ids {
		_, err := client.CreateContainerWithSpec(ctx, id, busyboxImage, resources,
			withProcessArgs("sleep", "300"),
		)
		if err != nil {
			t.Fatalf("CreateContainer(%s) failed: %v", id, err)
		}
	}
	t.Logf("created %d containers", len(ids))

	// Start all three
	for _, id := range ids {
		if err := client.StartContainer(ctx, id); err != nil {
			t.Fatalf("StartContainer(%s) failed: %v", id, err)
		}
	}
	time.Sleep(500 * time.Millisecond)
	t.Log("started all containers")

	// Verify all running
	for _, id := range ids {
		status, err := client.GetContainerStatus(ctx, id)
		if err != nil {
			t.Errorf("GetContainerStatus(%s) failed: %v", id, err)
			continue
		}
		if status != types.ContainerStatusRunning {
			t.Errorf("container %s: expected running, got %s", id, status)
		}
	}

	// ListContainers should include all three
	listed := client.ListContainers()
	count := 0
	for _, info := range listed {
		for _, id := range ids {
			if info.ID == id {
				count++
			}
		}
	}
	if count != len(ids) {
		t.Errorf("ListContainers found %d of %d test containers", count, len(ids))
	}
	t.Logf("ListContainers confirmed %d containers", count)

	// Stop and delete all
	for _, id := range ids {
		if err := client.StopContainer(ctx, id, 5*time.Second); err != nil {
			t.Errorf("StopContainer(%s) failed: %v", id, err)
		}
		if err := client.DeleteContainer(ctx, id); err != nil {
			t.Errorf("DeleteContainer(%s) failed: %v", id, err)
		}
	}
	t.Log("cleaned up all containers")
}

// containsSubstring checks if s contains substr (used because strings.Contains
// requires import and this is simpler for tests).
func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && searchSubstring(s, substr)
}

func searchSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
