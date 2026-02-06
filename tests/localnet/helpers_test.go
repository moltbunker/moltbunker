//go:build localnet

package localnet

import (
	"context"
	"fmt"
	"math/big"
	"math/rand"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/moltbunker/moltbunker/internal/daemon"
	"github.com/moltbunker/moltbunker/pkg/types"
)

// testNode wraps a daemon.Node with test metadata.
type testNode struct {
	Node    *daemon.Node
	KeyPath string
	Port    int
	Index   int
}

// randomPort returns a random high port in the range [30000, 40000) to avoid
// conflicts with other services and between parallel test runs.
func randomPort() int {
	return 30000 + rand.Intn(10000)
}

// createTestNode creates a single test node using temporary directories for key
// storage and a random high port. The node is not started; call node.Start(ctx)
// to begin listening.
func createTestNode(t *testing.T, ctx context.Context, index int) *testNode {
	t.Helper()

	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "node.key")
	keystoreDir := filepath.Join(tmpDir, "keystore")
	port := randomPort()

	node, err := daemon.NewNode(ctx, keyPath, keystoreDir, port)
	if err != nil {
		t.Fatalf("failed to create test node %d: %v", index, err)
	}

	return &testNode{
		Node:    node,
		KeyPath: keyPath,
		Port:    port,
		Index:   index,
	}
}

// startTestNode creates and starts a test node. The node is registered for
// cleanup so that it is closed when the test finishes.
func startTestNode(t *testing.T, ctx context.Context, index int) *testNode {
	t.Helper()

	tn := createTestNode(t, ctx, index)

	if err := tn.Node.Start(ctx); err != nil {
		t.Fatalf("failed to start test node %d: %v", index, err)
	}

	t.Cleanup(func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := tn.Node.Shutdown(shutdownCtx); err != nil {
			t.Logf("warning: node %d shutdown error: %v", index, err)
		}
	})

	return tn
}

// startTestNodes creates and starts n test nodes. All nodes are registered for
// cleanup.
func startTestNodes(t *testing.T, ctx context.Context, n int) []*testNode {
	t.Helper()

	nodes := make([]*testNode, n)
	for i := 0; i < n; i++ {
		nodes[i] = startTestNode(t, ctx, i)
	}
	return nodes
}

// waitForPeerDiscovery polls each node's Router until every node has discovered
// at least minPeers peers, or the context is canceled. It checks once every
// pollInterval.
func waitForPeerDiscovery(ctx context.Context, nodes []*testNode, minPeers int, pollInterval time.Duration) error {
	for {
		select {
		case <-ctx.Done():
			// Build a diagnostic string showing how many peers each node has.
			var diag string
			for _, tn := range nodes {
				diag += fmt.Sprintf("  node %d: %d peers\n", tn.Index, tn.Node.Router().PeerCount())
			}
			return fmt.Errorf("timed out waiting for peer discovery (min %d):\n%s", minPeers, diag)
		default:
		}

		allReady := true
		for _, tn := range nodes {
			if tn.Node.Router().PeerCount() < minPeers {
				allReady = false
				break
			}
		}
		if allReady {
			return nil
		}

		time.Sleep(pollInterval)
	}
}

// waitForCondition polls until condFn returns true or the context expires.
func waitForCondition(ctx context.Context, pollInterval time.Duration, condFn func() bool) error {
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("condition not met before timeout: %w", ctx.Err())
		default:
		}

		if condFn() {
			return nil
		}

		time.Sleep(pollInterval)
	}
}

// collectNodeIDs returns the NodeID for each test node.
func collectNodeIDs(nodes []*testNode) []types.NodeID {
	ids := make([]types.NodeID, len(nodes))
	for i, tn := range nodes {
		ids[i] = tn.Node.NodeInfo().ID
	}
	return ids
}

// messageCollector is a thread-safe accumulator for received messages.
// Tests can register a handler that appends incoming messages so they can be
// inspected after the fact.
type messageCollector struct {
	mu       sync.Mutex
	messages []*types.Message
}

func newMessageCollector() *messageCollector {
	return &messageCollector{}
}

func (mc *messageCollector) collect(_ context.Context, msg *types.Message, _ *types.Node) error {
	mc.mu.Lock()
	mc.messages = append(mc.messages, msg)
	mc.mu.Unlock()
	return nil
}

func (mc *messageCollector) count() int {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	return len(mc.messages)
}

func (mc *messageCollector) getMessages() []*types.Message {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	out := make([]*types.Message, len(mc.messages))
	copy(out, mc.messages)
	return out
}

// newBigInt parses a decimal string into a *big.Int. It panics on invalid
// input, which is acceptable for test constants.
func newBigInt(s string) *big.Int {
	n, ok := new(big.Int).SetString(s, 10)
	if !ok {
		panic("invalid big.Int string: " + s)
	}
	return n
}
