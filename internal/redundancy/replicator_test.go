package redundancy

import (
	"testing"
	"time"

	"github.com/moltbunker/moltbunker/pkg/types"
)

func createTestContainer(id string, nodeID types.NodeID) *types.Container {
	return &types.Container{
		ID:        id,
		NodeID:    nodeID,
		Status:    types.ContainerStatusRunning,
		CreatedAt: time.Now(),
		Health: types.HealthStatus{
			Healthy: true,
		},
	}
}

func TestReplicator_CreateReplicaSet(t *testing.T) {
	r := NewReplicator()

	containerID := "test-container"
	regions := []string{"Americas", "Europe", "Asia-Pacific"}

	replicaSet, err := r.CreateReplicaSet(containerID, regions)
	if err != nil {
		t.Fatalf("Failed to create replica set: %v", err)
	}

	if replicaSet.ContainerID != containerID {
		t.Errorf("ContainerID mismatch: got %s, want %s", replicaSet.ContainerID, containerID)
	}

	if replicaSet.Region1 != regions[0] {
		t.Errorf("Region1 mismatch: got %s, want %s", replicaSet.Region1, regions[0])
	}

	if replicaSet.Region2 != regions[1] {
		t.Errorf("Region2 mismatch: got %s, want %s", replicaSet.Region2, regions[1])
	}

	if replicaSet.Region3 != regions[2] {
		t.Errorf("Region3 mismatch: got %s, want %s", replicaSet.Region3, regions[2])
	}
}

func TestReplicator_CreateReplicaSet_InsufficientRegions(t *testing.T) {
	r := NewReplicator()

	regions := []string{"Americas", "Europe"}

	_, err := r.CreateReplicaSet("test-container", regions)
	if err == nil {
		t.Error("Should fail with insufficient regions")
	}
}

func TestReplicator_AddReplica(t *testing.T) {
	r := NewReplicator()

	containerID := "test-container"
	regions := []string{"Americas", "Europe", "Asia-Pacific"}
	r.CreateReplicaSet(containerID, regions)

	var nodeID types.NodeID
	copy(nodeID[:], []byte("node1"))
	container := createTestContainer("replica1", nodeID)

	err := r.AddReplica(containerID, 0, container)
	if err != nil {
		t.Fatalf("Failed to add replica: %v", err)
	}

	replicaSet, exists := r.GetReplicaSet(containerID)
	if !exists {
		t.Fatal("Replica set should exist")
	}

	if replicaSet.Replicas[0] == nil {
		t.Error("Replica should be added")
	}

	if replicaSet.Replicas[0].ID != "replica1" {
		t.Errorf("Replica ID mismatch: got %s, want replica1", replicaSet.Replicas[0].ID)
	}
}

func TestReplicator_AddReplica_InvalidIndex(t *testing.T) {
	r := NewReplicator()

	containerID := "test-container"
	regions := []string{"Americas", "Europe", "Asia-Pacific"}
	r.CreateReplicaSet(containerID, regions)

	var nodeID types.NodeID
	container := createTestContainer("replica1", nodeID)

	err := r.AddReplica(containerID, 5, container)
	if err == nil {
		t.Error("Should fail with invalid replica index")
	}
}

func TestReplicator_GetReplicaSet(t *testing.T) {
	r := NewReplicator()

	containerID := "test-container"
	regions := []string{"Americas", "Europe", "Asia-Pacific"}
	r.CreateReplicaSet(containerID, regions)

	replicaSet, exists := r.GetReplicaSet(containerID)
	if !exists {
		t.Error("Replica set should exist")
	}

	if replicaSet == nil {
		t.Error("Replica set should not be nil")
	}
}

func TestReplicator_GetReplicaSet_NotExists(t *testing.T) {
	r := NewReplicator()

	_, exists := r.GetReplicaSet("nonexistent")
	if exists {
		t.Error("Replica set should not exist")
	}
}

func TestReplicator_RemoveReplica(t *testing.T) {
	r := NewReplicator()

	containerID := "test-container"
	regions := []string{"Americas", "Europe", "Asia-Pacific"}
	r.CreateReplicaSet(containerID, regions)

	var nodeID types.NodeID
	container := createTestContainer("replica1", nodeID)
	r.AddReplica(containerID, 0, container)

	err := r.RemoveReplica(containerID, 0)
	if err != nil {
		t.Fatalf("Failed to remove replica: %v", err)
	}

	replicaSet, _ := r.GetReplicaSet(containerID)
	if replicaSet.Replicas[0] != nil {
		t.Error("Replica should be removed")
	}
}
