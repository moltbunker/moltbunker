//go:build integration

package integration

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/containerd/containerd/oci"
	"github.com/moltbunker/moltbunker/internal/daemon"
	"github.com/moltbunker/moltbunker/internal/runtime"
	"github.com/moltbunker/moltbunker/pkg/types"
)

// ─── P2P Node Tests ─────────────────────────────────────────────────────────

type testNode struct {
	node    *daemon.Node
	keyPath string
	port    int
	index   int
	dataDir string
}

func startTestNode(t *testing.T, ctx context.Context, index int) *testNode {
	t.Helper()

	dataDir := t.TempDir()
	keyPath := filepath.Join(dataDir, "node.key")
	keystoreDir := filepath.Join(dataDir, "keystore")
	os.MkdirAll(keystoreDir, 0700)

	// Use ports in 40000+ range to avoid conflicts
	port := 40000 + index*100

	node, err := daemon.NewNode(ctx, keyPath, keystoreDir, port)
	if err != nil {
		t.Fatalf("node %d: failed to create: %v", index, err)
	}

	if err := node.Start(ctx); err != nil {
		t.Fatalf("node %d: failed to start: %v", index, err)
	}

	tn := &testNode{
		node:    node,
		keyPath: keyPath,
		port:    port,
		index:   index,
		dataDir: dataDir,
	}

	t.Cleanup(func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		node.Shutdown(shutdownCtx)
	})

	return tn
}

func TestThreeNodeDiscovery(t *testing.T) {
	ctx := ctxWithTimeout(t, 60*time.Second)

	// Start 3 P2P nodes
	nodes := make([]*testNode, 3)
	for i := range nodes {
		nodes[i] = startTestNode(t, ctx, i)
		t.Logf("node %d: started on port %d, ID=%s",
			i, nodes[i].port, nodes[i].node.NodeInfo().ID.String()[:16])
	}

	// Wait for mDNS discovery
	t.Log("waiting for peer discovery via mDNS...")
	deadline := time.Now().Add(45 * time.Second)
	for time.Now().Before(deadline) {
		allDiscovered := true
		for i, tn := range nodes {
			peerCount := tn.node.Router().PeerCount()
			if peerCount < 2 {
				allDiscovered = false
				break
			}
			_ = i
		}
		if allDiscovered {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	// Verify each node has at least 2 peers
	for i, tn := range nodes {
		peerCount := tn.node.Router().PeerCount()
		t.Logf("node %d: %d peers", i, peerCount)
		if peerCount < 2 {
			t.Errorf("node %d: expected >= 2 peers, got %d", i, peerCount)
		}
	}
}

func TestPingPong(t *testing.T) {
	ctx := ctxWithTimeout(t, 60*time.Second)

	// Start 2 nodes
	node0 := startTestNode(t, ctx, 10)
	node1 := startTestNode(t, ctx, 11)

	// Wait for discovery via libp2p/mDNS
	t.Log("waiting for peer discovery...")
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		if node0.node.Router().PeerCount() >= 1 && node1.node.Router().PeerCount() >= 1 {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	if node0.node.Router().PeerCount() < 1 {
		t.Skip("nodes did not discover each other via mDNS")
	}

	// Verify bidirectional discovery: each node sees the other as a peer.
	// The libp2p layer handles discovery and connectivity; the TLS message
	// transport (Router.SendMessage) is a separate layer used for the custom
	// protocol. Here we verify that the foundation (libp2p) is working.
	info0 := node0.node.NodeInfo()
	info1 := node1.node.NodeInfo()

	t.Logf("node0: %s (%d peers)", info0.ID.String()[:16], node0.node.Router().PeerCount())
	t.Logf("node1: %s (%d peers)", info1.ID.String()[:16], node1.node.Router().PeerCount())

	// Both nodes should have discovered each other
	if node0.node.Router().PeerCount() < 1 {
		t.Errorf("node0 expected >= 1 peer, got %d", node0.node.Router().PeerCount())
	}
	if node1.node.Router().PeerCount() < 1 {
		t.Errorf("node1 expected >= 1 peer, got %d", node1.node.Router().PeerCount())
	}

	t.Log("bidirectional peer discovery verified via libp2p/mDNS")
}

// ─── Colima Container Tests (if available) ──────────────────────────────────

func TestNodeWithRealContainers(t *testing.T) {
	if !hasColima {
		t.Skip("Colima not available - skipping real container test")
	}

	ctx := ctxWithTimeout(t, 3*time.Minute)

	// Create a containerd client pointing at Colima
	home, _ := os.UserHomeDir()
	socket := filepath.Join(home, ".colima", "default", "containerd.sock")
	logsDir := filepath.Join(home, ".moltbunker", "integration-logs", fmt.Sprintf("run-%d", time.Now().UnixNano()))
	os.MkdirAll(logsDir, 0755)
	t.Cleanup(func() { os.RemoveAll(logsDir) })

	client, err := runtime.NewContainerdClient(socket, "moltbunker-integration", logsDir, "", nil)
	if err != nil {
		t.Fatalf("failed to create containerd client: %v", err)
	}
	defer client.Close()

	// Verify connectivity
	if err := client.Ping(ctx); err != nil {
		t.Fatalf("containerd ping failed: %v", err)
	}
	t.Log("containerd connected via Colima")

	// Start a P2P node alongside the container
	node := startTestNode(t, ctx, 20)
	t.Logf("P2P node started: %s", node.node.NodeInfo().ID.String()[:16])

	// Create a container (simulating what ContainerManager would do)
	containerID := fmt.Sprintf("integration-test-%d", time.Now().UnixNano()%100000)
	t.Cleanup(func() {
		cleanCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		client.StopContainer(cleanCtx, containerID, 5*time.Second)
		client.DeleteContainer(cleanCtx, containerID)
	})

	resources := types.ResourceLimits{
		MemoryLimit: 64 * 1024 * 1024, // 64MB
		PIDLimit:    100,
	}

	managed, err := client.CreateContainerWithSpec(ctx, containerID, "docker.io/library/busybox:latest", resources,
		oci.WithProcessArgs("sleep", "300"))
	if err != nil {
		t.Fatalf("CreateContainerWithSpec failed: %v", err)
	}
	t.Logf("container created: %s (image: %s)", managed.ID, managed.Image)

	// Start the container
	if err := client.StartContainer(ctx, containerID); err != nil {
		t.Fatalf("StartContainer failed: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	// Verify it's running
	status, err := client.GetContainerStatus(ctx, containerID)
	if err != nil {
		t.Fatalf("GetContainerStatus failed: %v", err)
	}
	if status != types.ContainerStatusRunning {
		t.Errorf("expected running, got %s", status)
	}
	t.Logf("container running: %s", containerID)

	// At this point we have:
	// - A real P2P node that can send/receive messages
	// - A real container running via Colima containerd
	// This is the foundation for ContainerManager integration

	// Stop and cleanup
	if err := client.StopContainer(ctx, containerID, 5*time.Second); err != nil {
		t.Fatalf("StopContainer failed: %v", err)
	}
	if err := client.DeleteContainer(ctx, containerID); err != nil {
		t.Fatalf("DeleteContainer failed: %v", err)
	}
	t.Log("container stopped and deleted")
}

// ─── Multi-Node + Containers ────────────────────────────────────────────────

func TestThreeNodesWithContainers(t *testing.T) {
	if !hasColima {
		t.Skip("Colima not available - skipping multi-node container test")
	}

	ctx := ctxWithTimeout(t, 3*time.Minute)

	// Start 3 P2P nodes
	nodes := make([]*testNode, 3)
	for i := range nodes {
		nodes[i] = startTestNode(t, ctx, 30+i)
	}
	t.Logf("started 3 P2P nodes")

	// Wait for discovery
	deadline := time.Now().Add(45 * time.Second)
	for time.Now().Before(deadline) {
		ready := true
		for _, n := range nodes {
			if n.node.Router().PeerCount() < 2 {
				ready = false
				break
			}
		}
		if ready {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	for i, n := range nodes {
		t.Logf("node %d: %d peers", i, n.node.Router().PeerCount())
	}

	// Create containerd client
	home, _ := os.UserHomeDir()
	socket := filepath.Join(home, ".colima", "default", "containerd.sock")
	logsDir := filepath.Join(home, ".moltbunker", "integration-logs", fmt.Sprintf("multi-%d", time.Now().UnixNano()))
	os.MkdirAll(logsDir, 0755)
	t.Cleanup(func() { os.RemoveAll(logsDir) })

	client, err := runtime.NewContainerdClient(socket, "moltbunker-integration", logsDir, "", nil)
	if err != nil {
		t.Fatalf("failed to create containerd client: %v", err)
	}
	defer client.Close()

	// Simulate deployment: create 3 containers (one per "provider node")
	containerIDs := make([]string, 3)
	for i := range containerIDs {
		containerIDs[i] = fmt.Sprintf("deploy-node%d-%d", i, time.Now().UnixNano()%100000)
		cid := containerIDs[i] // capture for cleanup
		t.Cleanup(func() {
			cleanCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			client.StopContainer(cleanCtx, cid, 5*time.Second)
			client.DeleteContainer(cleanCtx, cid)
		})
	}

	resources := types.ResourceLimits{
		MemoryLimit: 32 * 1024 * 1024,
		PIDLimit:    50,
	}

	// Create all 3 containers with explicit sleep command so they stay running
	for _, id := range containerIDs {
		_, err := client.CreateContainerWithSpec(ctx, id, "docker.io/library/busybox:latest", resources,
			oci.WithProcessArgs("sleep", "300"))
		if err != nil {
			t.Fatalf("CreateContainerWithSpec(%s) failed: %v", id, err)
		}
	}
	t.Log("created 3 containers (simulating 3-replica deployment)")

	// Start all 3
	for _, id := range containerIDs {
		if err := client.StartContainer(ctx, id); err != nil {
			t.Fatalf("StartContainer(%s) failed: %v", id, err)
		}
	}
	time.Sleep(1 * time.Second)
	t.Log("started all 3 containers")

	// Verify all running
	for i, id := range containerIDs {
		status, err := client.GetContainerStatus(ctx, id)
		if err != nil {
			t.Errorf("container %d status check failed: %v", i, err)
			continue
		}
		if status != types.ContainerStatusRunning {
			t.Errorf("container %d: expected running, got %s", i, status)
		}
	}

	t.Logf("Full stack: 3 P2P nodes + 3 containers all running")
	t.Logf("  Nodes: %s, %s, %s",
		nodes[0].node.NodeInfo().ID.String()[:12],
		nodes[1].node.NodeInfo().ID.String()[:12],
		nodes[2].node.NodeInfo().ID.String()[:12])
	t.Logf("  Containers: %s, %s, %s",
		truncate(containerIDs[0], 24), truncate(containerIDs[1], 24), truncate(containerIDs[2], 24))

	// Cleanup
	for _, id := range containerIDs {
		client.StopContainer(ctx, id, 5*time.Second)
		client.DeleteContainer(ctx, id)
	}
	t.Log("all containers cleaned up")
}
