//go:build e2e

package scenarios

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/moltbunker/moltbunker/internal/redundancy"
	"github.com/moltbunker/moltbunker/pkg/types"
	"github.com/moltbunker/moltbunker/tests/e2e/testutil"
)

// gossipNetwork simulates a gossip network of ConsensusManager nodes.
// It provides helpers to propagate state between nodes, simulating what
// the real GossipProtocol does over P2P transport.
type gossipNetwork struct {
	nodes       []*redundancy.ConsensusManager
	partitions  map[int]int // node index -> partition group
}

// newGossipNetwork creates a simulated gossip network with n nodes.
func newGossipNetwork(n int) *gossipNetwork {
	nodes := make([]*redundancy.ConsensusManager, n)
	for i := 0; i < n; i++ {
		nodes[i] = redundancy.NewConsensusManager()
	}
	return &gossipNetwork{
		nodes:      nodes,
		partitions: nil,
	}
}

// setPartitions configures which nodes can communicate. Nodes in the same
// partition group can gossip with each other; nodes in different groups cannot.
func (gn *gossipNetwork) setPartitions(groups map[int]int) {
	gn.partitions = groups
}

// clearPartitions removes all network partitions.
func (gn *gossipNetwork) clearPartitions() {
	gn.partitions = nil
}

// canCommunicate returns true if node i and node j can exchange gossip.
func (gn *gossipNetwork) canCommunicate(i, j int) bool {
	if gn.partitions == nil {
		return true
	}
	return gn.partitions[i] == gn.partitions[j]
}

// propagate simulates one round of gossip for a specific container ID.
// Each node sends its state to all reachable peers. Multiple rounds may be
// needed for full convergence in partitioned networks.
func (gn *gossipNetwork) propagate(containerID string) {
	// Collect current states from all nodes before propagation
	type snapshot struct {
		state  *redundancy.ContainerState
		exists bool
	}
	snapshots := make([]snapshot, len(gn.nodes))
	for i, node := range gn.nodes {
		state, exists := node.GetState(containerID)
		snapshots[i] = snapshot{state: state, exists: exists}
	}

	// Each node merges state from every reachable peer
	for i, node := range gn.nodes {
		for j, snap := range snapshots {
			if i == j || !snap.exists {
				continue
			}
			if !gn.canCommunicate(i, j) {
				continue
			}
			// Deep-copy the state to avoid shared pointer issues
			copied := copyContainerState(snap.state)
			node.MergeState(containerID, copied)
		}
	}
}

// propagateRounds runs multiple rounds of gossip propagation.
func (gn *gossipNetwork) propagateRounds(containerID string, rounds int) {
	for r := 0; r < rounds; r++ {
		gn.propagate(containerID)
	}
}

// copyContainerState creates a deep copy of a ContainerState.
func copyContainerState(cs *redundancy.ContainerState) *redundancy.ContainerState {
	if cs == nil {
		return nil
	}
	copied := &redundancy.ContainerState{
		ContainerID: cs.ContainerID,
		Status:      cs.Status,
		Version:     cs.Version,
	}
	for i, r := range cs.Replicas {
		if r != nil {
			c := *r
			copied.Replicas[i] = &c
		}
	}
	return copied
}

// makeReplicas creates a [3]*types.Container array with the given statuses.
func makeReplicas(statuses ...types.ContainerStatus) [3]*types.Container {
	var replicas [3]*types.Container
	for i := 0; i < 3 && i < len(statuses); i++ {
		replicas[i] = &types.Container{
			ID:        fmt.Sprintf("replica-%d", i),
			Status:    statuses[i],
			CreatedAt: time.Now(),
		}
	}
	return replicas
}

// TestGossipConvergence_MajorityAgreement simulates 5 nodes where 3 out of 5
// update their container status to Running. After gossip propagation, the
// highest-versioned state should win (LWW), and consensus among the 3 replicas
// in that state should achieve 2/3 majority.
func TestGossipConvergence_MajorityAgreement(t *testing.T) {
	assert := testutil.NewAssertions(t)

	net := newGossipNetwork(5)
	containerID := "convergence-majority"

	// Nodes 0, 1, 2 set status to Running (version increments to 1 on each)
	runningReplicas := makeReplicas(
		types.ContainerStatusRunning,
		types.ContainerStatusRunning,
		types.ContainerStatusRunning,
	)
	for i := 0; i < 3; i++ {
		net.nodes[i].UpdateState(containerID, types.ContainerStatusRunning, runningReplicas)
	}

	// Nodes 3, 4 set status to Pending (also version 1)
	pendingReplicas := makeReplicas(
		types.ContainerStatusPending,
		types.ContainerStatusPending,
		types.ContainerStatusPending,
	)
	for i := 3; i < 5; i++ {
		net.nodes[i].UpdateState(containerID, types.ContainerStatusPending, pendingReplicas)
	}

	// Node 0 updates again so its version (2) is strictly higher, making it the
	// authoritative state after LWW merge.
	net.nodes[0].UpdateState(containerID, types.ContainerStatusRunning, runningReplicas)

	// Propagate gossip across all 5 nodes for several rounds
	net.propagateRounds(containerID, 3)

	// Verify all nodes converge to the same state
	for i, node := range net.nodes {
		state, exists := node.GetState(containerID)
		assert.True(exists, fmt.Sprintf("Node %d should have state", i))
		assert.Equal(types.ContainerStatusRunning, state.Status,
			fmt.Sprintf("Node %d should converge to Running", i))

		// Verify 2/3 consensus on replicas
		consensusStatus, err := node.GetConsensusStatus(containerID)
		assert.NoError(err, fmt.Sprintf("Node %d consensus error", i))
		assert.Equal(types.ContainerStatusRunning, consensusStatus,
			fmt.Sprintf("Node %d consensus should be Running (2/3 majority)", i))
	}

	// Verify the winning version is 2 (from node 0's second update)
	for i, node := range net.nodes {
		state, _ := node.GetState(containerID)
		assert.Equal(int64(2), state.Version,
			fmt.Sprintf("Node %d should have version 2", i))
	}
}

// TestGossipConvergence_SplitBrain simulates a network partition where nodes
// are split into two groups (2 vs 3). Each partition evolves independently.
// The larger partition (3 nodes) makes more updates, achieving a higher version.
// After healing the partition, the larger partition's state should win via LWW.
func TestGossipConvergence_SplitBrain(t *testing.T) {
	assert := testutil.NewAssertions(t)

	net := newGossipNetwork(5)
	containerID := "convergence-splitbrain"

	// Initial state: all nodes agree on Running
	runningReplicas := makeReplicas(
		types.ContainerStatusRunning,
		types.ContainerStatusRunning,
		types.ContainerStatusRunning,
	)
	for i := 0; i < 5; i++ {
		net.nodes[i].UpdateState(containerID, types.ContainerStatusRunning, runningReplicas)
	}
	net.propagateRounds(containerID, 2)

	// Create partition: group A = {0, 1}, group B = {2, 3, 4}
	net.setPartitions(map[int]int{0: 0, 1: 0, 2: 1, 3: 1, 4: 1})

	// Group A: nodes 0,1 see container as Stopped
	stoppedReplicas := makeReplicas(
		types.ContainerStatusStopped,
		types.ContainerStatusStopped,
		types.ContainerStatusRunning,
	)
	net.nodes[0].UpdateState(containerID, types.ContainerStatusStopped, stoppedReplicas)
	net.propagateRounds(containerID, 2) // Only within partitions

	// Group B: nodes 2,3,4 update status multiple times (higher version)
	replicatingReplicas := makeReplicas(
		types.ContainerStatusReplicating,
		types.ContainerStatusReplicating,
		types.ContainerStatusReplicating,
	)
	net.nodes[2].UpdateState(containerID, types.ContainerStatusReplicating, replicatingReplicas)
	net.nodes[2].UpdateState(containerID, types.ContainerStatusReplicating, replicatingReplicas)
	net.propagateRounds(containerID, 2) // Only within partition B

	// Verify partition isolation: group A and group B have different states
	stateA, _ := net.nodes[0].GetState(containerID)
	stateB, _ := net.nodes[2].GetState(containerID)
	assert.NotEqual(stateA.Version, stateB.Version, "Partitions should have diverged")

	// Heal partition
	net.clearPartitions()
	net.propagateRounds(containerID, 3)

	// After healing, all nodes should converge to the higher-version state
	// Group B had more updates, so its version should be higher
	for i, node := range net.nodes {
		state, exists := node.GetState(containerID)
		assert.True(exists, fmt.Sprintf("Node %d should have state", i))
		assert.Equal(types.ContainerStatusReplicating, state.Status,
			fmt.Sprintf("Node %d should converge to Replicating (larger partition wins)", i))
	}

	// Verify consensus: all 3 replicas in the winning state are Replicating
	for i, node := range net.nodes {
		consensusStatus, err := node.GetConsensusStatus(containerID)
		assert.NoError(err, fmt.Sprintf("Node %d consensus error", i))
		assert.Equal(types.ContainerStatusReplicating, consensusStatus,
			fmt.Sprintf("Node %d consensus should be Replicating", i))
	}
}

// TestGossipConvergence_LateJoiner simulates 3 nodes converging on a state,
// then a 4th node joins the network and receives the correct state via
// MergeState (simulating gossip propagation to the newcomer).
func TestGossipConvergence_LateJoiner(t *testing.T) {
	assert := testutil.NewAssertions(t)

	// Start with 3 nodes
	net := newGossipNetwork(3)
	containerID := "convergence-latejoin"

	// All 3 nodes agree: container is Running
	runningReplicas := makeReplicas(
		types.ContainerStatusRunning,
		types.ContainerStatusRunning,
		types.ContainerStatusRunning,
	)
	for i := 0; i < 3; i++ {
		net.nodes[i].UpdateState(containerID, types.ContainerStatusRunning, runningReplicas)
	}

	// Evolve state: node 0 updates multiple times
	net.nodes[0].UpdateState(containerID, types.ContainerStatusRunning, runningReplicas) // v2
	net.nodes[0].UpdateState(containerID, types.ContainerStatusRunning, runningReplicas) // v3

	// Propagate among existing 3 nodes
	net.propagateRounds(containerID, 3)

	// Verify existing nodes are converged
	for i := 0; i < 3; i++ {
		state, _ := net.nodes[i].GetState(containerID)
		assert.Equal(int64(3), state.Version,
			fmt.Sprintf("Node %d should be at version 3", i))
	}

	// Late joiner: add a 4th node
	lateJoiner := redundancy.NewConsensusManager()
	net.nodes = append(net.nodes, lateJoiner)

	// The late joiner has no state at all
	_, exists := lateJoiner.GetState(containerID)
	assert.False(exists, "Late joiner should have no initial state")

	// Simulate gossip from existing node (node 0) to the late joiner
	existingState, _ := net.nodes[0].GetState(containerID)
	lateJoiner.MergeState(containerID, copyContainerState(existingState))

	// Verify the late joiner received the correct state
	joinerState, exists := lateJoiner.GetState(containerID)
	assert.True(exists, "Late joiner should now have state")
	assert.Equal(types.ContainerStatusRunning, joinerState.Status,
		"Late joiner should have Running status")
	assert.Equal(int64(3), joinerState.Version,
		"Late joiner should have version 3")

	// Verify consensus on the late joiner
	consensusStatus, err := lateJoiner.GetConsensusStatus(containerID)
	assert.NoError(err, "Late joiner consensus should not error")
	assert.Equal(types.ContainerStatusRunning, consensusStatus,
		"Late joiner consensus should be Running (3/3 replicas agree)")

	// Verify the late joiner's replicas match the source
	for i := 0; i < 3; i++ {
		if existingState.Replicas[i] != nil {
			assert.NotNil(joinerState.Replicas[i],
				fmt.Sprintf("Late joiner replica %d should exist", i))
			assert.Equal(existingState.Replicas[i].Status, joinerState.Replicas[i].Status,
				fmt.Sprintf("Late joiner replica %d status should match", i))
		}
	}
}

// TestGossipConvergence_ConcurrentUpdates tests that concurrent state updates
// from multiple goroutines produce a consistent final consensus. This exercises
// the mutex-protected MergeState under contention.
func TestGossipConvergence_ConcurrentUpdates(t *testing.T) {
	assert := testutil.NewAssertions(t)

	net := newGossipNetwork(5)
	containerID := "convergence-concurrent"

	// Each node will concurrently update its own state
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(nodeIdx int) {
			defer wg.Done()
			// Each node performs multiple updates
			for j := 0; j < 10; j++ {
				status := types.ContainerStatusRunning
				if j%3 == 0 {
					status = types.ContainerStatusPaused
				}
				replicas := makeReplicas(status, status, status)
				net.nodes[nodeIdx].UpdateState(containerID, status, replicas)
			}
		}(i)
	}
	wg.Wait()

	// Now propagate gossip concurrently as well
	var wg2 sync.WaitGroup
	for round := 0; round < 5; round++ {
		wg2.Add(1)
		go func() {
			defer wg2.Done()
			net.propagate(containerID)
		}()
	}
	wg2.Wait()

	// Do a final serial propagation to ensure full convergence
	net.propagateRounds(containerID, 5)

	// All nodes should have converged to the same state
	var referenceVersion int64
	var referenceStatus types.ContainerStatus
	refState, exists := net.nodes[0].GetState(containerID)
	assert.True(exists, "Node 0 should have state")
	referenceVersion = refState.Version
	referenceStatus = refState.Status

	for i := 1; i < 5; i++ {
		state, exists := net.nodes[i].GetState(containerID)
		assert.True(exists, fmt.Sprintf("Node %d should have state", i))
		assert.Equal(referenceVersion, state.Version,
			fmt.Sprintf("Node %d version should match reference (got %d, want %d)",
				i, state.Version, referenceVersion))
		assert.Equal(referenceStatus, state.Status,
			fmt.Sprintf("Node %d status should match reference", i))
	}

	// Verify consensus is consistent across all nodes
	refConsensus, err := net.nodes[0].GetConsensusStatus(containerID)
	assert.NoError(err, "Reference consensus should not error")

	for i := 1; i < 5; i++ {
		consensus, err := net.nodes[i].GetConsensusStatus(containerID)
		assert.NoError(err, fmt.Sprintf("Node %d consensus error", i))
		assert.Equal(refConsensus, consensus,
			fmt.Sprintf("Node %d consensus should match reference", i))
	}
}

// TestGossipConvergence_VersionConflict tests LWW (Last Writer Wins) conflict
// resolution. Two conflicting updates with different versions are made, and
// the higher-versioned update should always win regardless of merge order.
func TestGossipConvergence_VersionConflict(t *testing.T) {
	assert := testutil.NewAssertions(t)

	containerID := "convergence-versionconflict"

	// Create two independent nodes that make conflicting updates
	nodeA := redundancy.NewConsensusManager()
	nodeB := redundancy.NewConsensusManager()

	// Node A: updates container to Running, version becomes 1
	runningReplicas := makeReplicas(
		types.ContainerStatusRunning,
		types.ContainerStatusRunning,
		types.ContainerStatusRunning,
	)
	nodeA.UpdateState(containerID, types.ContainerStatusRunning, runningReplicas)

	// Node B: updates container to Stopped, then updates again -> version 2
	stoppedReplicas := makeReplicas(
		types.ContainerStatusStopped,
		types.ContainerStatusStopped,
		types.ContainerStatusStopped,
	)
	nodeB.UpdateState(containerID, types.ContainerStatusStopped, stoppedReplicas)
	nodeB.UpdateState(containerID, types.ContainerStatusStopped, stoppedReplicas) // v2

	// Verify versions before merge
	stateA, _ := nodeA.GetState(containerID)
	stateB, _ := nodeB.GetState(containerID)
	assert.Equal(int64(1), stateA.Version, "Node A should be at version 1")
	assert.Equal(int64(2), stateB.Version, "Node B should be at version 2")

	// --- Test 1: Merge B into A (higher version wins) ---
	nodeA.MergeState(containerID, copyContainerState(stateB))
	mergedA, _ := nodeA.GetState(containerID)
	assert.Equal(types.ContainerStatusStopped, mergedA.Status,
		"After merging higher-version state, A should adopt Stopped")
	assert.Equal(int64(2), mergedA.Version,
		"After merge, A should be at version 2")

	// --- Test 2: Merge A's original state into B (lower version loses) ---
	// Reset nodeA to its original state for this sub-test
	nodeC := redundancy.NewConsensusManager()
	nodeC.UpdateState(containerID, types.ContainerStatusRunning, runningReplicas) // v1

	nodeD := redundancy.NewConsensusManager()
	nodeD.UpdateState(containerID, types.ContainerStatusStopped, stoppedReplicas)
	nodeD.UpdateState(containerID, types.ContainerStatusStopped, stoppedReplicas) // v2

	// Merge lower-version (C, v1) into higher-version (D, v2)
	stateC, _ := nodeC.GetState(containerID)
	nodeD.MergeState(containerID, copyContainerState(stateC))
	mergedD, _ := nodeD.GetState(containerID)
	assert.Equal(types.ContainerStatusStopped, mergedD.Status,
		"D should keep Stopped (higher version wins, lower version ignored)")
	assert.Equal(int64(2), mergedD.Version,
		"D should remain at version 2")

	// --- Test 3: Equal versions, first write preserved ---
	nodeE := redundancy.NewConsensusManager()
	nodeF := redundancy.NewConsensusManager()

	pausedReplicas := makeReplicas(
		types.ContainerStatusPaused,
		types.ContainerStatusPaused,
		types.ContainerStatusPaused,
	)

	nodeE.UpdateState(containerID, types.ContainerStatusRunning, runningReplicas)   // v1
	nodeF.UpdateState(containerID, types.ContainerStatusPaused, pausedReplicas)     // v1

	stateE, _ := nodeE.GetState(containerID)
	stateF, _ := nodeF.GetState(containerID)
	assert.Equal(int64(1), stateE.Version, "E should be v1")
	assert.Equal(int64(1), stateF.Version, "F should be v1")

	// When versions are equal, MergeState does NOT overwrite (not strictly greater)
	nodeE.MergeState(containerID, copyContainerState(stateF))
	mergedE, _ := nodeE.GetState(containerID)
	assert.Equal(types.ContainerStatusRunning, mergedE.Status,
		"Equal version merge should preserve existing state (LWW tie: first in place wins)")

	// --- Test 4: Verify consensus after conflict resolution ---
	// Create a 5-node network where 2 nodes have v1 (Running) and 3 have v2 (Stopped)
	net := newGossipNetwork(5)
	for i := 0; i < 2; i++ {
		net.nodes[i].UpdateState(containerID, types.ContainerStatusRunning, runningReplicas) // v1
	}
	for i := 2; i < 5; i++ {
		net.nodes[i].UpdateState(containerID, types.ContainerStatusStopped, stoppedReplicas)
		net.nodes[i].UpdateState(containerID, types.ContainerStatusStopped, stoppedReplicas) // v2
	}

	// Propagate
	net.propagateRounds(containerID, 3)

	// All nodes should converge to Stopped (version 2 > version 1)
	for i, node := range net.nodes {
		state, _ := node.GetState(containerID)
		assert.Equal(types.ContainerStatusStopped, state.Status,
			fmt.Sprintf("Node %d should converge to Stopped (v2 wins)", i))
		assert.Equal(int64(2), state.Version,
			fmt.Sprintf("Node %d should be at version 2", i))

		consensus, err := node.GetConsensusStatus(containerID)
		assert.NoError(err)
		assert.Equal(types.ContainerStatusStopped, consensus,
			fmt.Sprintf("Node %d consensus should be Stopped", i))
	}
}
