package redundancy

import (
	"testing"
	"time"

	"github.com/moltbunker/moltbunker/pkg/types"
)

func createTestContainerForConsensus(id string, status types.ContainerStatus) *types.Container {
	var nodeID types.NodeID
	// Safely copy id bytes, padding with zeros if shorter than 32 bytes
	idBytes := []byte(id)
	if len(idBytes) > 32 {
		copy(nodeID[:], idBytes[:32])
	} else {
		copy(nodeID[:], idBytes)
	}
	return &types.Container{
		ID:        id,
		NodeID:    nodeID,
		Status:    status,
		CreatedAt: time.Now(),
	}
}

func TestConsensusManager_UpdateState(t *testing.T) {
	cm := NewConsensusManager()

	containerID := "test-container"
	status := types.ContainerStatusRunning

	var replicas [3]*types.Container
	replicas[0] = createTestContainerForConsensus("replica1", status)
	replicas[1] = createTestContainerForConsensus("replica2", status)
	replicas[2] = createTestContainerForConsensus("replica3", status)

	cm.UpdateState(containerID, status, replicas)

	state, exists := cm.GetState(containerID)
	if !exists {
		t.Fatal("State should exist")
	}

	if state.Status != status {
		t.Errorf("Status mismatch: got %s, want %s", state.Status, status)
	}

	if state.Version != 1 {
		t.Errorf("Version should be 1, got %d", state.Version)
	}
}

func TestConsensusManager_GetConsensusStatus(t *testing.T) {
	cm := NewConsensusManager()

	containerID := "test-container"
	status := types.ContainerStatusRunning

	var replicas [3]*types.Container
	replicas[0] = createTestContainerForConsensus("replica1", status)
	replicas[1] = createTestContainerForConsensus("replica2", status)
	replicas[2] = createTestContainerForConsensus("replica3", status)

	cm.UpdateState(containerID, status, replicas)

	consensusStatus, err := cm.GetConsensusStatus(containerID)
	if err != nil {
		t.Fatalf("Failed to get consensus status: %v", err)
	}

	if consensusStatus != status {
		t.Errorf("Consensus status mismatch: got %s, want %s", consensusStatus, status)
	}
}

func TestConsensusManager_GetConsensusStatus_Majority(t *testing.T) {
	cm := NewConsensusManager()

	containerID := "test-container"

	var replicas [3]*types.Container
	replicas[0] = createTestContainerForConsensus("replica1", types.ContainerStatusRunning)
	replicas[1] = createTestContainerForConsensus("replica2", types.ContainerStatusRunning)
	replicas[2] = createTestContainerForConsensus("replica3", types.ContainerStatusStopped)

	cm.UpdateState(containerID, types.ContainerStatusRunning, replicas)

	consensusStatus, err := cm.GetConsensusStatus(containerID)
	if err != nil {
		t.Fatalf("Failed to get consensus status: %v", err)
	}

	if consensusStatus != types.ContainerStatusRunning {
		t.Errorf("Consensus should be Running (majority), got %s", consensusStatus)
	}
}

func TestConsensusManager_MergeState(t *testing.T) {
	cm := NewConsensusManager()

	containerID := "test-container"

	// Initial state
	var replicas1 [3]*types.Container
	replicas1[0] = createTestContainerForConsensus("replica1", types.ContainerStatusRunning)
	cm.UpdateState(containerID, types.ContainerStatusRunning, replicas1)

	// Merge with newer state
	var replicas2 [3]*types.Container
	replicas2[0] = createTestContainerForConsensus("replica1", types.ContainerStatusStopped)
	otherState := &ContainerState{
		ContainerID: containerID,
		Status:      types.ContainerStatusStopped,
		Replicas:    replicas2,
		Version:     2,
	}

	cm.MergeState(containerID, otherState)

	state, _ := cm.GetState(containerID)
	if state.Status != types.ContainerStatusStopped {
		t.Errorf("State should be updated: got %s, want Stopped", state.Status)
	}

	if state.Version != 2 {
		t.Errorf("Version should be 2, got %d", state.Version)
	}
}

func TestConsensusManager_MergeState_OlderVersion(t *testing.T) {
	cm := NewConsensusManager()

	containerID := "test-container"

	// Initial state with version 2
	var replicas1 [3]*types.Container
	replicas1[0] = createTestContainerForConsensus("replica1", types.ContainerStatusStopped)
	cm.UpdateState(containerID, types.ContainerStatusStopped, replicas1)
	state1, _ := cm.GetState(containerID)
	state1.Version = 2

	// Try to merge with older version
	var replicas2 [3]*types.Container
	replicas2[0] = createTestContainerForConsensus("replica1", types.ContainerStatusRunning)
	otherState := &ContainerState{
		ContainerID: containerID,
		Status:      types.ContainerStatusRunning,
		Replicas:    replicas2,
		Version:     1,
	}

	cm.MergeState(containerID, otherState)

	state, _ := cm.GetState(containerID)
	if state.Status != types.ContainerStatusStopped {
		t.Error("State should not be updated with older version")
	}
}
