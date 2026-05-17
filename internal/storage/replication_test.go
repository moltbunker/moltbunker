package storage

import (
	"fmt"
	"testing"
	"time"
)

func TestReplication_TrackAndAddReplica(t *testing.T) {
	rm := NewReplicationManager(DefaultReplicationConfig())
	rm.TrackObject("bucket", "file.dat", "wallet1", 1024)

	err := rm.AddReplica("bucket", "file.dat", ObjectReplica{
		ProviderID: "provider-1",
		Region:     "us-east",
		Status:     ReplicaStatusSyncing,
	})
	if err != nil {
		t.Fatalf("AddReplica: %v", err)
	}

	rs, ok := rm.GetReplicaSet("bucket", "file.dat")
	if !ok {
		t.Fatal("replica set not found")
	}
	if len(rs.Replicas) != 1 {
		t.Fatalf("replicas = %d, want 1", len(rs.Replicas))
	}
	if rs.Replicas[0].ProviderID != "provider-1" {
		t.Errorf("provider = %q, want provider-1", rs.Replicas[0].ProviderID)
	}
}

func TestReplication_DuplicateProvider(t *testing.T) {
	rm := NewReplicationManager(DefaultReplicationConfig())
	rm.TrackObject("b", "k", "w", 100)

	if err := rm.AddReplica("b", "k", ObjectReplica{ProviderID: "p1", Region: "us"}); err != nil {
		t.Fatalf("AddReplica: %v", err)
	}
	err := rm.AddReplica("b", "k", ObjectReplica{ProviderID: "p1", Region: "eu"})
	if err == nil {
		t.Error("expected error for duplicate provider")
	}
}

func TestReplication_MaxReplicas(t *testing.T) {
	rm := NewReplicationManager(DefaultReplicationConfig())
	rm.TrackObject("b", "k", "w", 100)

	for i := 0; i < 3; i++ {
		err := rm.AddReplica("b", "k", ObjectReplica{
			ProviderID: fmt.Sprintf("p%d", i),
			Region:     fmt.Sprintf("region-%d", i),
		})
		if err != nil {
			t.Fatalf("AddReplica %d: %v", i, err)
		}
	}

	// 4th replica should fail
	err := rm.AddReplica("b", "k", ObjectReplica{ProviderID: "p3", Region: "extra"})
	if err == nil {
		t.Error("expected error when exceeding target replicas")
	}
}

func TestReplication_UpdateStatus(t *testing.T) {
	rm := NewReplicationManager(DefaultReplicationConfig())
	rm.TrackObject("b", "k", "w", 100)
	if err := rm.AddReplica("b", "k", ObjectReplica{ProviderID: "p1", Status: ReplicaStatusSyncing}); err != nil {
		t.Fatalf("AddReplica: %v", err)
	}

	err := rm.UpdateReplicaStatus("b", "k", "p1", ReplicaStatusHealthy)
	if err != nil {
		t.Fatalf("UpdateReplicaStatus: %v", err)
	}

	rs, _ := rm.GetReplicaSet("b", "k")
	if rs.Replicas[0].Status != ReplicaStatusHealthy {
		t.Errorf("status = %q, want healthy", rs.Replicas[0].Status)
	}
	if rs.Replicas[0].SyncedAt.IsZero() {
		t.Error("SyncedAt should be set when status becomes healthy")
	}
}

func TestReplication_UpdateStatus_NotFound(t *testing.T) {
	rm := NewReplicationManager(DefaultReplicationConfig())
	rm.TrackObject("b", "k", "w", 100)

	err := rm.UpdateReplicaStatus("b", "k", "nonexistent", ReplicaStatusHealthy)
	if err == nil {
		t.Error("expected error for nonexistent provider")
	}
}

func TestReplication_UntrackedObject(t *testing.T) {
	rm := NewReplicationManager(DefaultReplicationConfig())

	err := rm.AddReplica("b", "k", ObjectReplica{ProviderID: "p1"})
	if err == nil {
		t.Error("expected error for untracked object")
	}

	_, ok := rm.GetReplicaSet("b", "k")
	if ok {
		t.Error("expected not found for untracked object")
	}
}

func TestReplication_RemoveObject(t *testing.T) {
	rm := NewReplicationManager(DefaultReplicationConfig())
	rm.TrackObject("b", "k", "w", 100)
	if err := rm.AddReplica("b", "k", ObjectReplica{ProviderID: "p1"}); err != nil {
		t.Fatalf("AddReplica: %v", err)
	}

	rm.RemoveObject("b", "k")

	_, ok := rm.GetReplicaSet("b", "k")
	if ok {
		t.Error("replica set should be gone after removal")
	}
}

func TestReplication_GetUnderReplicated(t *testing.T) {
	rm := NewReplicationManager(DefaultReplicationConfig())

	// Object with only 1 healthy replica
	rm.TrackObject("b", "under", "w", 100)
	if err := rm.AddReplica("b", "under", ObjectReplica{
		ProviderID:    "p1",
		Status:        ReplicaStatusHealthy,
		LastHeartbeat: time.Now(),
	}); err != nil {
		t.Fatalf("AddReplica under/p1: %v", err)
	}

	// Object with 3 healthy replicas
	rm.TrackObject("b", "full", "w", 200)
	for i := 0; i < 3; i++ {
		if err := rm.AddReplica("b", "full", ObjectReplica{
			ProviderID:    fmt.Sprintf("p%d", i),
			Region:        fmt.Sprintf("r%d", i),
			Status:        ReplicaStatusHealthy,
			LastHeartbeat: time.Now(),
		}); err != nil {
			t.Fatalf("AddReplica full/p%d: %v", i, err)
		}
	}

	underRep := rm.GetUnderReplicated()
	if len(underRep) != 1 {
		t.Fatalf("under-replicated count = %d, want 1", len(underRep))
	}
	if underRep[0].Key != "under" {
		t.Errorf("under-replicated key = %q, want under", underRep[0].Key)
	}
}

func TestReplication_HealthyReplicaCount(t *testing.T) {
	rm := NewReplicationManager(DefaultReplicationConfig())
	rm.TrackObject("b", "k", "w", 100)

	if err := rm.AddReplica("b", "k", ObjectReplica{
		ProviderID:    "p1",
		Status:        ReplicaStatusHealthy,
		LastHeartbeat: time.Now(),
	}); err != nil {
		t.Fatalf("AddReplica p1: %v", err)
	}
	if err := rm.AddReplica("b", "k", ObjectReplica{
		ProviderID:    "p2",
		Status:        ReplicaStatusDegraded,
		LastHeartbeat: time.Now(),
	}); err != nil {
		t.Fatalf("AddReplica p2: %v", err)
	}
	if err := rm.AddReplica("b", "k", ObjectReplica{
		ProviderID:    "p3",
		Status:        ReplicaStatusHealthy,
		LastHeartbeat: time.Now(),
	}); err != nil {
		t.Fatalf("AddReplica p3: %v", err)
	}

	count := rm.HealthyReplicaCount("b", "k")
	if count != 2 {
		t.Errorf("healthy count = %d, want 2", count)
	}
}

func TestReplication_StaleHeartbeat(t *testing.T) {
	cfg := DefaultReplicationConfig()
	cfg.HeartbeatTimeout = 100 * time.Millisecond
	rm := NewReplicationManager(cfg)

	rm.TrackObject("b", "k", "w", 100)
	if err := rm.AddReplica("b", "k", ObjectReplica{
		ProviderID:    "p1",
		Status:        ReplicaStatusHealthy,
		LastHeartbeat: time.Now().Add(-time.Second), // stale
	}); err != nil {
		t.Fatalf("AddReplica: %v", err)
	}

	count := rm.HealthyReplicaCount("b", "k")
	if count != 0 {
		t.Errorf("stale heartbeat should not count as healthy, got %d", count)
	}
}

func TestReplication_MergeReplicaSet(t *testing.T) {
	rm := NewReplicationManager(DefaultReplicationConfig())

	// Merge new object
	accepted := rm.MergeReplicaSet(&ObjectReplicaSet{
		Bucket:  "b",
		Key:     "k",
		Owner:   "w",
		Size:    100,
		Version: 1,
		Replicas: []ObjectReplica{
			{ProviderID: "p1", Status: ReplicaStatusHealthy},
		},
	})
	if !accepted {
		t.Error("initial merge should be accepted")
	}

	// Merge higher version — should accept
	accepted = rm.MergeReplicaSet(&ObjectReplicaSet{
		Bucket:  "b",
		Key:     "k",
		Owner:   "w",
		Size:    100,
		Version: 5,
		Replicas: []ObjectReplica{
			{ProviderID: "p1", Status: ReplicaStatusHealthy},
			{ProviderID: "p2", Status: ReplicaStatusSyncing},
		},
	})
	if !accepted {
		t.Error("higher version merge should be accepted")
	}

	rs, _ := rm.GetReplicaSet("b", "k")
	if len(rs.Replicas) != 2 {
		t.Errorf("replicas = %d, want 2 after merge", len(rs.Replicas))
	}

	// Merge lower version — should reject
	accepted = rm.MergeReplicaSet(&ObjectReplicaSet{
		Bucket:  "b",
		Key:     "k",
		Owner:   "w",
		Size:    100,
		Version: 3,
		Replicas: []ObjectReplica{
			{ProviderID: "p1", Status: ReplicaStatusLost},
		},
	})
	if accepted {
		t.Error("lower version merge should be rejected")
	}

	// Verify original state preserved
	rs, _ = rm.GetReplicaSet("b", "k")
	if len(rs.Replicas) != 2 {
		t.Errorf("replicas should still be 2 after rejected merge, got %d", len(rs.Replicas))
	}
}

func TestReplication_Stats(t *testing.T) {
	rm := NewReplicationManager(DefaultReplicationConfig())

	rm.TrackObject("b", "full", "w", 100)
	for i := 0; i < 3; i++ {
		if err := rm.AddReplica("b", "full", ObjectReplica{
			ProviderID:    fmt.Sprintf("p%d", i),
			Status:        ReplicaStatusHealthy,
			LastHeartbeat: time.Now(),
		}); err != nil {
			t.Fatalf("AddReplica full/p%d: %v", i, err)
		}
	}

	rm.TrackObject("b", "partial", "w", 200)
	if err := rm.AddReplica("b", "partial", ObjectReplica{
		ProviderID:    "p1",
		Status:        ReplicaStatusHealthy,
		LastHeartbeat: time.Now(),
	}); err != nil {
		t.Fatalf("AddReplica partial/p1: %v", err)
	}

	stats := rm.Stats()
	if stats.TrackedObjects != 2 {
		t.Errorf("tracked = %d, want 2", stats.TrackedObjects)
	}
	if stats.FullyReplicated != 1 {
		t.Errorf("fully replicated = %d, want 1", stats.FullyReplicated)
	}
	if stats.UnderReplicated != 1 {
		t.Errorf("under replicated = %d, want 1", stats.UnderReplicated)
	}
	if stats.TotalReplicas != 4 {
		t.Errorf("total replicas = %d, want 4", stats.TotalReplicas)
	}
	if stats.HealthyReplicas != 4 {
		t.Errorf("healthy replicas = %d, want 4", stats.HealthyReplicas)
	}
}

func TestReplication_DeepCopy(t *testing.T) {
	rm := NewReplicationManager(DefaultReplicationConfig())
	rm.TrackObject("b", "k", "w", 100)
	if err := rm.AddReplica("b", "k", ObjectReplica{ProviderID: "p1", Status: ReplicaStatusHealthy}); err != nil {
		t.Fatalf("AddReplica: %v", err)
	}

	rs, _ := rm.GetReplicaSet("b", "k")
	rs.Replicas[0].Status = ReplicaStatusLost // mutate copy

	// Original should be unchanged
	rsOrig, _ := rm.GetReplicaSet("b", "k")
	if rsOrig.Replicas[0].Status != ReplicaStatusHealthy {
		t.Error("mutation of copy should not affect original")
	}
}
