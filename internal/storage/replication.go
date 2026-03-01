package storage

import (
	"fmt"
	"sync"
	"time"

	"github.com/moltbunker/moltbunker/internal/logging"
)

// ReplicaStatus tracks the state of a single object replica.
type ReplicaStatus string

const (
	ReplicaStatusPending  ReplicaStatus = "pending"
	ReplicaStatusSyncing  ReplicaStatus = "syncing"
	ReplicaStatusHealthy  ReplicaStatus = "healthy"
	ReplicaStatusDegraded ReplicaStatus = "degraded"
	ReplicaStatusLost     ReplicaStatus = "lost"
)

// ObjectReplica describes one copy of an object on a specific provider.
type ObjectReplica struct {
	ProviderID    string        `json:"provider_id"`
	Region        string        `json:"region"`
	Status        ReplicaStatus `json:"status"`
	CID           string        `json:"cid,omitempty"`
	LastHeartbeat time.Time     `json:"last_heartbeat"`
	SyncedAt      time.Time     `json:"synced_at,omitempty"`
}

// ObjectReplicaSet tracks all replicas for a single object.
type ObjectReplicaSet struct {
	Bucket   string           `json:"bucket"`
	Key      string           `json:"key"`
	Owner    string           `json:"owner"`
	Size     int64            `json:"size"`
	Replicas []ObjectReplica  `json:"replicas"`
	Version  int64            `json:"version"` // LWW version for consensus
}

// ReplicationConfig configures the replication manager.
type ReplicationConfig struct {
	TargetReplicas    int           // Target replica count (default: 3)
	HeartbeatTimeout  time.Duration // Max time before replica is degraded
	RepairInterval    time.Duration // How often to check for under-replicated objects
}

// DefaultReplicationConfig returns sensible defaults.
func DefaultReplicationConfig() ReplicationConfig {
	return ReplicationConfig{
		TargetReplicas:   3,
		HeartbeatTimeout: 2 * time.Minute,
		RepairInterval:   5 * time.Minute,
	}
}

// ReplicationManager tracks object replicas across the P2P network.
type ReplicationManager struct {
	mu       sync.RWMutex
	replicas map[string]*ObjectReplicaSet // key: "bucket/key"
	config   ReplicationConfig
}

// NewReplicationManager creates a new replication manager.
func NewReplicationManager(cfg ReplicationConfig) *ReplicationManager {
	return &ReplicationManager{
		replicas: make(map[string]*ObjectReplicaSet),
		config:   cfg,
	}
}

// replicaKey builds the lookup key for a bucket/object pair.
func replicaKey(bucket, key string) string {
	return bucket + "/" + key
}

// TrackObject registers an object for replication tracking.
func (rm *ReplicationManager) TrackObject(bucket, key, owner string, size int64) {
	rk := replicaKey(bucket, key)

	rm.mu.Lock()
	defer rm.mu.Unlock()

	rm.replicas[rk] = &ObjectReplicaSet{
		Bucket:   bucket,
		Key:      key,
		Owner:    owner,
		Size:     size,
		Replicas: make([]ObjectReplica, 0, rm.config.TargetReplicas),
		Version:  1,
	}
}

// AddReplica records a replica for an object.
func (rm *ReplicationManager) AddReplica(bucket, key string, replica ObjectReplica) error {
	rk := replicaKey(bucket, key)

	rm.mu.Lock()
	defer rm.mu.Unlock()

	rs, ok := rm.replicas[rk]
	if !ok {
		return fmt.Errorf("object %q not tracked for replication", rk)
	}

	if len(rs.Replicas) >= rm.config.TargetReplicas {
		return fmt.Errorf("object %q already has %d replicas", rk, rm.config.TargetReplicas)
	}

	// Check for duplicate provider
	for _, r := range rs.Replicas {
		if r.ProviderID == replica.ProviderID {
			return fmt.Errorf("provider %q already hosts a replica of %q", replica.ProviderID, rk)
		}
	}

	rs.Replicas = append(rs.Replicas, replica)
	rs.Version++

	logging.Debug("replica added",
		"bucket", bucket,
		"key", key,
		"provider", replica.ProviderID,
		"region", replica.Region,
		"replicas", len(rs.Replicas),
		logging.Component("storage-replication"))

	return nil
}

// UpdateReplicaStatus updates the status and heartbeat of a replica.
func (rm *ReplicationManager) UpdateReplicaStatus(bucket, key, providerID string, status ReplicaStatus) error {
	rk := replicaKey(bucket, key)

	rm.mu.Lock()
	defer rm.mu.Unlock()

	rs, ok := rm.replicas[rk]
	if !ok {
		return fmt.Errorf("object %q not tracked", rk)
	}

	for i := range rs.Replicas {
		if rs.Replicas[i].ProviderID == providerID {
			rs.Replicas[i].Status = status
			rs.Replicas[i].LastHeartbeat = time.Now()
			if status == ReplicaStatusHealthy && rs.Replicas[i].SyncedAt.IsZero() {
				rs.Replicas[i].SyncedAt = time.Now()
			}
			rs.Version++
			return nil
		}
	}

	return fmt.Errorf("no replica for provider %q on object %q", providerID, rk)
}

// GetReplicaSet returns a deep copy of the replica set for an object.
func (rm *ReplicationManager) GetReplicaSet(bucket, key string) (*ObjectReplicaSet, bool) {
	rk := replicaKey(bucket, key)

	rm.mu.RLock()
	defer rm.mu.RUnlock()

	rs, ok := rm.replicas[rk]
	if !ok {
		return nil, false
	}

	// Deep copy
	rsCopy := &ObjectReplicaSet{
		Bucket:   rs.Bucket,
		Key:      rs.Key,
		Owner:    rs.Owner,
		Size:     rs.Size,
		Version:  rs.Version,
		Replicas: make([]ObjectReplica, len(rs.Replicas)),
	}
	copy(rsCopy.Replicas, rs.Replicas)
	return rsCopy, true
}

// RemoveObject stops tracking an object.
func (rm *ReplicationManager) RemoveObject(bucket, key string) {
	rk := replicaKey(bucket, key)

	rm.mu.Lock()
	defer rm.mu.Unlock()

	delete(rm.replicas, rk)
}

// GetUnderReplicated returns objects with fewer healthy replicas than the target.
func (rm *ReplicationManager) GetUnderReplicated() []ObjectReplicaSet {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	now := time.Now()
	var result []ObjectReplicaSet

	for _, rs := range rm.replicas {
		healthy := 0
		for _, r := range rs.Replicas {
			if r.Status == ReplicaStatusHealthy && now.Sub(r.LastHeartbeat) < rm.config.HeartbeatTimeout {
				healthy++
			}
		}

		if healthy < rm.config.TargetReplicas {
			// Deep copy
			rsCopy := ObjectReplicaSet{
				Bucket:   rs.Bucket,
				Key:      rs.Key,
				Owner:    rs.Owner,
				Size:     rs.Size,
				Version:  rs.Version,
				Replicas: make([]ObjectReplica, len(rs.Replicas)),
			}
			copy(rsCopy.Replicas, rs.Replicas)
			result = append(result, rsCopy)
		}
	}

	return result
}

// MergeReplicaSet handles incoming gossip state — LWW conflict resolution.
func (rm *ReplicationManager) MergeReplicaSet(incoming *ObjectReplicaSet) bool {
	rk := replicaKey(incoming.Bucket, incoming.Key)

	rm.mu.Lock()
	defer rm.mu.Unlock()

	existing, ok := rm.replicas[rk]
	if !ok {
		// New object — accept it
		cp := &ObjectReplicaSet{
			Bucket:   incoming.Bucket,
			Key:      incoming.Key,
			Owner:    incoming.Owner,
			Size:     incoming.Size,
			Version:  incoming.Version,
			Replicas: make([]ObjectReplica, len(incoming.Replicas)),
		}
		copy(cp.Replicas, incoming.Replicas)
		rm.replicas[rk] = cp
		return true
	}

	// LWW: accept if incoming version is higher
	if incoming.Version > existing.Version {
		existing.Replicas = make([]ObjectReplica, len(incoming.Replicas))
		copy(existing.Replicas, incoming.Replicas)
		existing.Version = incoming.Version
		return true
	}

	return false
}

// HealthyReplicaCount returns the number of healthy replicas for an object.
func (rm *ReplicationManager) HealthyReplicaCount(bucket, key string) int {
	rk := replicaKey(bucket, key)

	rm.mu.RLock()
	defer rm.mu.RUnlock()

	rs, ok := rm.replicas[rk]
	if !ok {
		return 0
	}

	now := time.Now()
	count := 0
	for _, r := range rs.Replicas {
		if r.Status == ReplicaStatusHealthy && now.Sub(r.LastHeartbeat) < rm.config.HeartbeatTimeout {
			count++
		}
	}
	return count
}

// Stats returns replication statistics.
func (rm *ReplicationManager) Stats() ReplicationStats {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	now := time.Now()
	stats := ReplicationStats{
		TrackedObjects: len(rm.replicas),
	}

	for _, rs := range rm.replicas {
		healthy := 0
		for _, r := range rs.Replicas {
			stats.TotalReplicas++
			if r.Status == ReplicaStatusHealthy && now.Sub(r.LastHeartbeat) < rm.config.HeartbeatTimeout {
				stats.HealthyReplicas++
				healthy++
			}
		}
		if healthy >= rm.config.TargetReplicas {
			stats.FullyReplicated++
		} else {
			stats.UnderReplicated++
		}
	}

	return stats
}

// ReplicationStats aggregates replication health metrics.
type ReplicationStats struct {
	TrackedObjects   int `json:"tracked_objects"`
	TotalReplicas    int `json:"total_replicas"`
	HealthyReplicas  int `json:"healthy_replicas"`
	FullyReplicated  int `json:"fully_replicated"`
	UnderReplicated  int `json:"under_replicated"`
}
