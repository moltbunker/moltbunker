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

// CreateReplicaSet creates a new replica set with 1-3 copies
func (r *Replicator) CreateReplicaSet(containerID string, regions []string) (*types.ReplicaSet, error) {
	if len(regions) == 0 {
		return nil, fmt.Errorf("need at least 1 region for deployment")
	}
	if len(regions) > 3 {
		regions = regions[:3]
	}

	replicaSet := &types.ReplicaSet{
		ContainerID: containerID,
		Replicas:    make([]*types.Container, len(regions)),
		Regions:     make([]string, len(regions)),
		CreatedAt:   time.Now(),
	}
	copy(replicaSet.Regions, regions)

	r.mu.Lock()
	r.replicas[containerID] = replicaSet
	r.mu.Unlock()

	// Return a copy to prevent race conditions on the internal data
	regionsCopy := make([]string, len(regions))
	copy(regionsCopy, regions)
	returnCopy := &types.ReplicaSet{
		ContainerID: replicaSet.ContainerID,
		Replicas:    make([]*types.Container, len(regions)),
		Regions:     regionsCopy,
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

	if replicaIndex < 0 || replicaIndex >= len(replicaSet.Replicas) {
		return fmt.Errorf("invalid replica index: %d (max: %d)", replicaIndex, len(replicaSet.Replicas)-1)
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
	replicasCopy := make([]*types.Container, len(replicaSet.Replicas))
	copy(replicasCopy, replicaSet.Replicas)
	regionsCopy := make([]string, len(replicaSet.Regions))
	copy(regionsCopy, replicaSet.Regions)
	rsCopy := &types.ReplicaSet{
		ContainerID: replicaSet.ContainerID,
		Replicas:    replicasCopy,
		Regions:     regionsCopy,
		CreatedAt:   replicaSet.CreatedAt,
	}
	return rsCopy, true
}

// RemoveReplica removes a replica from a replica set
func (r *Replicator) RemoveReplica(containerID string, replicaIndex int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	replicaSet, exists := r.replicas[containerID]
	if !exists {
		return fmt.Errorf("replica set not found: %s", containerID)
	}

	if replicaIndex < 0 || replicaIndex >= len(replicaSet.Replicas) {
		return fmt.Errorf("invalid replica index: %d (max: %d)", replicaIndex, len(replicaSet.Replicas)-1)
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
		replicasCopy := make([]*types.Container, len(v.Replicas))
		copy(replicasCopy, v.Replicas)
		regionsCopy := make([]string, len(v.Regions))
		copy(regionsCopy, v.Regions)
		result[k] = &types.ReplicaSet{
			ContainerID: v.ContainerID,
			Replicas:    replicasCopy,
			Regions:     regionsCopy,
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
