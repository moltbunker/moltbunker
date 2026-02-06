package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/moltbunker/moltbunker/internal/snapshot"
)

// SnapshotManager is set externally after daemon startup
var snapshotManager *snapshot.Manager

// SetSnapshotManager sets the global snapshot manager
func SetSnapshotManager(m *snapshot.Manager) {
	snapshotManager = m
}

// SnapshotResponse contains snapshot information
type SnapshotResponse struct {
	ID          string            `json:"id"`
	ContainerID string            `json:"container_id"`
	CreatedAt   time.Time         `json:"created_at"`
	Size        int64             `json:"size"`
	Checksum    string            `json:"checksum"`
	Type        string            `json:"type"`
	ParentID    string            `json:"parent_id,omitempty"`
	Compressed  bool              `json:"compressed"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// handleSnapshotCreate handles snapshot creation
func (s *APIServer) handleSnapshotCreate(ctx context.Context, req *APIRequest) *APIResponse {
	if snapshotManager == nil {
		return &APIResponse{
			Error: "snapshot manager not initialized",
			ID:    req.ID,
		}
	}

	var params struct {
		ContainerID string            `json:"container_id"`
		Type        string            `json:"type"`
		Metadata    map[string]string `json:"metadata"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return &APIResponse{
			Error: fmt.Sprintf("invalid params: %v", err),
			ID:    req.ID,
		}
	}

	if params.ContainerID == "" {
		return &APIResponse{
			Error: "container_id is required",
			ID:    req.ID,
		}
	}

	// Determine snapshot type
	snapshotType := snapshot.SnapshotTypeFull
	switch params.Type {
	case "incremental":
		snapshotType = snapshot.SnapshotTypeIncremental
	case "checkpoint":
		snapshotType = snapshot.SnapshotTypeCheckpoint
	}

	// Get container state (in a real implementation)
	// For now, use empty state
	stateData := []byte("{}")

	snap, err := snapshotManager.CreateSnapshot(params.ContainerID, stateData, snapshotType, params.Metadata)
	if err != nil {
		return &APIResponse{
			Error: fmt.Sprintf("snapshot creation failed: %v", err),
			ID:    req.ID,
		}
	}

	response := SnapshotResponse{
		ID:          snap.ID,
		ContainerID: snap.ContainerID,
		CreatedAt:   snap.CreatedAt,
		Size:        snap.Size,
		Checksum:    snap.Checksum,
		Type:        string(snap.Type),
		ParentID:    snap.ParentID,
		Compressed:  snap.Compressed,
		Metadata:    snap.Metadata,
	}

	return &APIResponse{
		Result: response,
		ID:     req.ID,
	}
}

// handleSnapshotGet handles snapshot retrieval
func (s *APIServer) handleSnapshotGet(ctx context.Context, req *APIRequest) *APIResponse {
	if snapshotManager == nil {
		return &APIResponse{
			Error: "snapshot manager not initialized",
			ID:    req.ID,
		}
	}

	var params struct {
		SnapshotID string `json:"snapshot_id"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return &APIResponse{
			Error: fmt.Sprintf("invalid params: %v", err),
			ID:    req.ID,
		}
	}

	if params.SnapshotID == "" {
		return &APIResponse{
			Error: "snapshot_id is required",
			ID:    req.ID,
		}
	}

	snap, err := snapshotManager.GetSnapshot(params.SnapshotID)
	if err != nil {
		return &APIResponse{
			Error: fmt.Sprintf("snapshot not found: %v", err),
			ID:    req.ID,
		}
	}

	response := SnapshotResponse{
		ID:          snap.ID,
		ContainerID: snap.ContainerID,
		CreatedAt:   snap.CreatedAt,
		Size:        snap.Size,
		Checksum:    snap.Checksum,
		Type:        string(snap.Type),
		ParentID:    snap.ParentID,
		Compressed:  snap.Compressed,
		Metadata:    snap.Metadata,
	}

	return &APIResponse{
		Result: response,
		ID:     req.ID,
	}
}

// handleSnapshotList handles listing snapshots
func (s *APIServer) handleSnapshotList(ctx context.Context, req *APIRequest) *APIResponse {
	if snapshotManager == nil {
		return &APIResponse{
			Error: "snapshot manager not initialized",
			ID:    req.ID,
		}
	}

	var params struct {
		ContainerID string `json:"container_id"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return &APIResponse{
			Error: fmt.Sprintf("invalid params: %v", err),
			ID:    req.ID,
		}
	}

	if params.ContainerID == "" {
		return &APIResponse{
			Error: "container_id is required",
			ID:    req.ID,
		}
	}

	snapshots := snapshotManager.ListSnapshots(params.ContainerID)

	responses := make([]SnapshotResponse, 0, len(snapshots))
	for _, snap := range snapshots {
		responses = append(responses, SnapshotResponse{
			ID:          snap.ID,
			ContainerID: snap.ContainerID,
			CreatedAt:   snap.CreatedAt,
			Size:        snap.Size,
			Checksum:    snap.Checksum,
			Type:        string(snap.Type),
			ParentID:    snap.ParentID,
			Compressed:  snap.Compressed,
		})
	}

	return &APIResponse{
		Result: map[string]interface{}{
			"snapshots": responses,
			"total":     len(responses),
		},
		ID: req.ID,
	}
}

// handleSnapshotDelete handles snapshot deletion
func (s *APIServer) handleSnapshotDelete(ctx context.Context, req *APIRequest) *APIResponse {
	if snapshotManager == nil {
		return &APIResponse{
			Error: "snapshot manager not initialized",
			ID:    req.ID,
		}
	}

	var params struct {
		SnapshotID  string `json:"snapshot_id,omitempty"`
		ContainerID string `json:"container_id,omitempty"`
		All         bool   `json:"all,omitempty"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return &APIResponse{
			Error: fmt.Sprintf("invalid params: %v", err),
			ID:    req.ID,
		}
	}

	if params.SnapshotID != "" {
		if err := snapshotManager.DeleteSnapshot(params.SnapshotID); err != nil {
			return &APIResponse{
				Error: fmt.Sprintf("delete failed: %v", err),
				ID:    req.ID,
			}
		}
		return &APIResponse{
			Result: map[string]interface{}{
				"status":      "deleted",
				"snapshot_id": params.SnapshotID,
			},
			ID: req.ID,
		}
	}

	if params.ContainerID != "" && params.All {
		if err := snapshotManager.DeleteContainerSnapshots(params.ContainerID); err != nil {
			return &APIResponse{
				Error: fmt.Sprintf("delete failed: %v", err),
				ID:    req.ID,
			}
		}
		return &APIResponse{
			Result: map[string]interface{}{
				"status":       "deleted",
				"container_id": params.ContainerID,
				"message":      "all snapshots deleted",
			},
			ID: req.ID,
		}
	}

	return &APIResponse{
		Error: "specify snapshot_id or container_id with all=true",
		ID:    req.ID,
	}
}

// handleSnapshotRestore handles restoring from a snapshot
func (s *APIServer) handleSnapshotRestore(ctx context.Context, req *APIRequest) *APIResponse {
	if snapshotManager == nil {
		return &APIResponse{
			Error: "snapshot manager not initialized",
			ID:    req.ID,
		}
	}

	var params struct {
		SnapshotID   string `json:"snapshot_id"`
		TargetRegion string `json:"target_region,omitempty"`
		NewContainer bool   `json:"new_container,omitempty"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return &APIResponse{
			Error: fmt.Sprintf("invalid params: %v", err),
			ID:    req.ID,
		}
	}

	if params.SnapshotID == "" {
		return &APIResponse{
			Error: "snapshot_id is required",
			ID:    req.ID,
		}
	}

	// Get snapshot data
	data, err := snapshotManager.GetSnapshotData(params.SnapshotID)
	if err != nil {
		return &APIResponse{
			Error: fmt.Sprintf("failed to get snapshot data: %v", err),
			ID:    req.ID,
		}
	}

	// In a real implementation, this would:
	// 1. Create a new container or update existing
	// 2. Apply the snapshot state
	// For now, return success with the data size

	return &APIResponse{
		Result: map[string]interface{}{
			"status":       "restoring",
			"snapshot_id":  params.SnapshotID,
			"data_size":    len(data),
			"target_region": params.TargetRegion,
			"new_container": params.NewContainer,
		},
		ID: req.ID,
	}
}

// handleSnapshotStats handles snapshot statistics
func (s *APIServer) handleSnapshotStats(ctx context.Context, req *APIRequest) *APIResponse {
	if snapshotManager == nil {
		return &APIResponse{
			Error: "snapshot manager not initialized",
			ID:    req.ID,
		}
	}

	stats := snapshotManager.GetStats()

	return &APIResponse{
		Result: stats,
		ID:     req.ID,
	}
}
