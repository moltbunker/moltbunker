package redundancy

import (
	"context"
	"fmt"
	"sync"
	"time"

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
		CreatedAt:   time.Now(),
	}

	r.mu.Lock()
	r.replicas[containerID] = replicaSet
	r.mu.Unlock()

	// Return a copy to prevent race conditions on the internal data
	returnCopy := &types.ReplicaSet{
		ContainerID: replicaSet.ContainerID,
		Replicas:    replicaSet.Replicas,
		Region1:     replicaSet.Region1,
		Region2:     replicaSet.Region2,
		Region3:     replicaSet.Region3,
		CreatedAt:   replicaSet.CreatedAt,
	}

	return returnCopy, nil
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

// GetReplicaSet retrieves a copy of a replica set
func (r *Replicator) GetReplicaSet(containerID string) (*types.ReplicaSet, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	replicaSet, exists := r.replicas[containerID]
	if !exists {
		return nil, false
	}

	// Return a copy to prevent race conditions
	copy := &types.ReplicaSet{
		ContainerID: replicaSet.ContainerID,
		Replicas:    replicaSet.Replicas, // Array copy is safe
		Region1:     replicaSet.Region1,
		Region2:     replicaSet.Region2,
		Region3:     replicaSet.Region3,
		CreatedAt:   replicaSet.CreatedAt,
	}
	return copy, true
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

// GetAllReplicaSets returns copies of all replica sets
func (r *Replicator) GetAllReplicaSets() map[string]*types.ReplicaSet {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Return deep copies to avoid race conditions
	result := make(map[string]*types.ReplicaSet, len(r.replicas))
	for k, v := range r.replicas {
		result[k] = &types.ReplicaSet{
			ContainerID: v.ContainerID,
			Replicas:    v.Replicas,
			Region1:     v.Region1,
			Region2:     v.Region2,
			Region3:     v.Region3,
			CreatedAt:   v.CreatedAt,
		}
	}
	return result
}

// DeleteReplicaSet removes a replica set
func (r *Replicator) DeleteReplicaSet(containerID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.replicas, containerID)
}
