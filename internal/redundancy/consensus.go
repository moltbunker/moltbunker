package redundancy

import (
	"fmt"
	"sync"

	"github.com/moltbunker/moltbunker/pkg/types"
)

// ConsensusManager manages consensus on container state via gossip
type ConsensusManager struct {
	states map[string]*ContainerState
	mu     sync.RWMutex
}

// ContainerState represents state of a container
type ContainerState struct {
	ContainerID string
	Status      types.ContainerStatus
	Replicas    [3]*types.Container
	Version     int64 // Version for conflict resolution
}

// NewConsensusManager creates a new consensus manager
func NewConsensusManager() *ConsensusManager {
	return &ConsensusManager{
		states: make(map[string]*ContainerState),
	}
}

// UpdateState updates container state
func (cm *ConsensusManager) UpdateState(containerID string, status types.ContainerStatus, replicas [3]*types.Container) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	state, exists := cm.states[containerID]
	if !exists {
		state = &ContainerState{
			ContainerID: containerID,
			Version:     0, // Start at 0 so first update becomes 1
		}
		cm.states[containerID] = state
	}

	state.Status = status
	state.Replicas = replicas
	state.Version++
}

// UpdateReplicaStatus updates a single replica's status within an existing container state.
// Used when receiving sync messages that only contain a replica index and status.
func (cm *ConsensusManager) UpdateReplicaStatus(containerID string, replicaIdx int, status types.ContainerStatus) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	state, exists := cm.states[containerID]
	if !exists {
		state = &ContainerState{
			ContainerID: containerID,
		}
		cm.states[containerID] = state
	}

	if replicaIdx >= 0 && replicaIdx < 3 {
		if state.Replicas[replicaIdx] == nil {
			state.Replicas[replicaIdx] = &types.Container{
				ID:     containerID,
				Status: status,
			}
		} else {
			state.Replicas[replicaIdx].Status = status
		}
	}
	state.Status = status
	state.Version++
}

// GetState returns container state
func (cm *ConsensusManager) GetState(containerID string) (*ContainerState, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	state, exists := cm.states[containerID]
	return state, exists
}

// MergeState merges state from another node (gossip)
func (cm *ConsensusManager) MergeState(containerID string, otherState *ContainerState) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	currentState, exists := cm.states[containerID]
	if !exists {
		cm.states[containerID] = otherState
		return
	}

	// Use version for conflict resolution (simple last-write-wins)
	if otherState.Version > currentState.Version {
		cm.states[containerID] = otherState
	}
}

// RemoveState removes all tracked state for a container.
func (cm *ConsensusManager) RemoveState(containerID string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	delete(cm.states, containerID)
}

// GetConsensusStatus determines consensus status across replicas
func (cm *ConsensusManager) GetConsensusStatus(containerID string) (types.ContainerStatus, error) {
	state, exists := cm.GetState(containerID)
	if !exists {
		return types.ContainerStatusFailed, fmt.Errorf("container state not found")
	}

	// Simple consensus: if 2/3 replicas agree, that's the status
	statusCounts := make(map[types.ContainerStatus]int)
	for _, replica := range state.Replicas {
		if replica != nil {
			statusCounts[replica.Status]++
		}
	}

	// Find status with majority
	maxCount := 0
	var consensusStatus types.ContainerStatus
	for status, count := range statusCounts {
		if count > maxCount {
			maxCount = count
			consensusStatus = status
		}
	}

	if maxCount >= 2 {
		return consensusStatus, nil
	}

	return state.Status, nil
}
