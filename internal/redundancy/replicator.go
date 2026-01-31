package redundancy

import (
	"context"
	"fmt"
	"sync"

	"github.com/moltbunker/moltbunker/pkg/types"
)

// Replicator manages 3-copy redundancy
type Replicator struct {
	replicas map[string]*types.ReplicaSet
	mu       sync.RWMutex
}

// NewReplicator creates a new replicator
func NewReplicator() *Replicator {
	return &Replicator{
		replicas: make(map[string]*types.ReplicaSet),
	}
}

// CreateReplicaSet creates a new replica set with 3 copies
func (r *Replicator) CreateReplicaSet(containerID string, regions []string) (*types.ReplicaSet, error) {
	if len(regions) < 3 {
		return nil, fmt.Errorf("need at least 3 regions for redundancy")
	}

	replicaSet := &types.ReplicaSet{
		ContainerID: containerID,
		Replicas:    [3]*types.Container{},
		Region1:     regions[0],
		Region2:     regions[1],
		Region3:     regions[2],
	}

	r.mu.Lock()
	r.replicas[containerID] = replicaSet
	r.mu.Unlock()

	return replicaSet, nil
}

// AddReplica adds a replica to a replica set
func (r *Replicator) AddReplica(containerID string, replicaIndex int, container *types.Container) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	replicaSet, exists := r.replicas[containerID]
	if !exists {
		return fmt.Errorf("replica set not found: %s", containerID)
	}

	if replicaIndex < 0 || replicaIndex >= 3 {
		return fmt.Errorf("invalid replica index: %d", replicaIndex)
	}

	replicaSet.Replicas[replicaIndex] = container
	return nil
}

// GetReplicaSet retrieves a replica set
func (r *Replicator) GetReplicaSet(containerID string) (*types.ReplicaSet, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	replicaSet, exists := r.replicas[containerID]
	return replicaSet, exists
}

// RemoveReplica removes a replica from a replica set
func (r *Replicator) RemoveReplica(containerID string, replicaIndex int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	replicaSet, exists := r.replicas[containerID]
	if !exists {
		return fmt.Errorf("replica set not found: %s", containerID)
	}

	if replicaIndex < 0 || replicaIndex >= 3 {
		return fmt.Errorf("invalid replica index: %d", replicaIndex)
	}

	replicaSet.Replicas[replicaIndex] = nil
	return nil
}

// ReplaceReplica replaces a failed replica
func (r *Replicator) ReplaceReplica(ctx context.Context, containerID string, replicaIndex int, newContainer *types.Container) error {
	return r.AddReplica(containerID, replicaIndex, newContainer)
}
