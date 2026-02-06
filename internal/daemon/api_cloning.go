package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/moltbunker/moltbunker/internal/cloning"
)

// CloningManager is set externally after daemon startup
var cloningManager *cloning.Manager

// SetCloningManager sets the global cloning manager
func SetCloningManager(m *cloning.Manager) {
	cloningManager = m
}

// CloneResponse contains clone operation information
type CloneResponse struct {
	CloneID      string            `json:"clone_id"`
	SourceID     string            `json:"source_id"`
	TargetID     string            `json:"target_id,omitempty"`
	TargetNodeID string            `json:"target_node_id,omitempty"`
	TargetRegion string            `json:"target_region"`
	Status       string            `json:"status"`
	SnapshotID   string            `json:"snapshot_id,omitempty"`
	Priority     int               `json:"priority"`
	Reason       string            `json:"reason"`
	CreatedAt    time.Time         `json:"created_at"`
	CompletedAt  time.Time         `json:"completed_at,omitempty"`
	Error        string            `json:"error,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// handleClone handles clone requests
func (s *APIServer) handleClone(ctx context.Context, req *APIRequest) *APIResponse {
	if cloningManager == nil {
		return &APIResponse{
			Error: "cloning manager not initialized",
			ID:    req.ID,
		}
	}

	var params struct {
		SourceID     string `json:"source_id"`
		TargetRegion string `json:"target_region"`
		Priority     int    `json:"priority"`
		Reason       string `json:"reason"`
		IncludeState bool   `json:"include_state"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return &APIResponse{
			Error: fmt.Sprintf("invalid params: %v", err),
			ID:    req.ID,
		}
	}

	if params.TargetRegion == "" {
		params.TargetRegion = "random"
	}
	if params.Reason == "" {
		params.Reason = "manual"
	}

	cloneReq := &cloning.CloneRequest{
		SourceID:     params.SourceID,
		TargetRegion: params.TargetRegion,
		Priority:     params.Priority,
		Reason:       params.Reason,
		IncludeState: params.IncludeState,
	}

	clone, err := cloningManager.RequestClone(ctx, cloneReq)
	if err != nil {
		return &APIResponse{
			Error: fmt.Sprintf("clone request failed: %v", err),
			ID:    req.ID,
		}
	}

	response := CloneResponse{
		CloneID:      clone.ID,
		SourceID:     clone.SourceID,
		TargetID:     clone.TargetID,
		TargetNodeID: clone.TargetNodeID.String(),
		TargetRegion: clone.TargetRegion,
		Status:       string(clone.Status),
		SnapshotID:   clone.SnapshotID,
		Priority:     clone.Priority,
		Reason:       clone.Reason,
		CreatedAt:    clone.CreatedAt,
		Metadata:     clone.Metadata,
	}

	return &APIResponse{
		Result: response,
		ID:     req.ID,
	}
}

// handleCloneStatus handles clone status queries
func (s *APIServer) handleCloneStatus(ctx context.Context, req *APIRequest) *APIResponse {
	if cloningManager == nil {
		return &APIResponse{
			Error: "cloning manager not initialized",
			ID:    req.ID,
		}
	}

	var params struct {
		CloneID string `json:"clone_id"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return &APIResponse{
			Error: fmt.Sprintf("invalid params: %v", err),
			ID:    req.ID,
		}
	}

	if params.CloneID == "" {
		return &APIResponse{
			Error: "clone_id is required",
			ID:    req.ID,
		}
	}

	clone, found := cloningManager.GetClone(params.CloneID)
	if !found {
		return &APIResponse{
			Error: "clone not found",
			ID:    req.ID,
		}
	}

	response := CloneResponse{
		CloneID:      clone.ID,
		SourceID:     clone.SourceID,
		TargetID:     clone.TargetID,
		TargetNodeID: clone.TargetNodeID.String(),
		TargetRegion: clone.TargetRegion,
		Status:       string(clone.Status),
		SnapshotID:   clone.SnapshotID,
		Priority:     clone.Priority,
		Reason:       clone.Reason,
		CreatedAt:    clone.CreatedAt,
		CompletedAt:  clone.CompletedAt,
		Error:        clone.Error,
		Metadata:     clone.Metadata,
	}

	return &APIResponse{
		Result: response,
		ID:     req.ID,
	}
}

// handleCloneList handles listing clones
func (s *APIServer) handleCloneList(ctx context.Context, req *APIRequest) *APIResponse {
	if cloningManager == nil {
		return &APIResponse{
			Error: "cloning manager not initialized",
			ID:    req.ID,
		}
	}

	var params struct {
		Active  bool `json:"active"`
		Limit   int  `json:"limit"`
	}
	if req.Params != nil {
		json.Unmarshal(req.Params, &params)
	}

	if params.Limit <= 0 {
		params.Limit = 50
	}

	var clones []*cloning.Clone
	if params.Active {
		clones = cloningManager.GetActiveClones()
	} else {
		clones = cloningManager.GetCloneHistory(params.Limit)
	}

	responses := make([]CloneResponse, 0, len(clones))
	for _, clone := range clones {
		responses = append(responses, CloneResponse{
			CloneID:      clone.ID,
			SourceID:     clone.SourceID,
			TargetID:     clone.TargetID,
			TargetNodeID: clone.TargetNodeID.String(),
			TargetRegion: clone.TargetRegion,
			Status:       string(clone.Status),
			SnapshotID:   clone.SnapshotID,
			Priority:     clone.Priority,
			Reason:       clone.Reason,
			CreatedAt:    clone.CreatedAt,
			CompletedAt:  clone.CompletedAt,
			Error:        clone.Error,
		})
	}

	return &APIResponse{
		Result: map[string]interface{}{
			"clones": responses,
			"total":  len(responses),
		},
		ID: req.ID,
	}
}

// handleCloneCancel handles cancelling a clone operation
func (s *APIServer) handleCloneCancel(ctx context.Context, req *APIRequest) *APIResponse {
	if cloningManager == nil {
		return &APIResponse{
			Error: "cloning manager not initialized",
			ID:    req.ID,
		}
	}

	var params struct {
		CloneID string `json:"clone_id"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return &APIResponse{
			Error: fmt.Sprintf("invalid params: %v", err),
			ID:    req.ID,
		}
	}

	if params.CloneID == "" {
		return &APIResponse{
			Error: "clone_id is required",
			ID:    req.ID,
		}
	}

	if err := cloningManager.CancelClone(params.CloneID); err != nil {
		return &APIResponse{
			Error: fmt.Sprintf("cancel failed: %v", err),
			ID:    req.ID,
		}
	}

	return &APIResponse{
		Result: map[string]interface{}{
			"status":   "cancelled",
			"clone_id": params.CloneID,
		},
		ID: req.ID,
	}
}

// handleCloneStats handles clone statistics queries
func (s *APIServer) handleCloneStats(ctx context.Context, req *APIRequest) *APIResponse {
	if cloningManager == nil {
		return &APIResponse{
			Error: "cloning manager not initialized",
			ID:    req.ID,
		}
	}

	stats := cloningManager.GetStats()

	return &APIResponse{
		Result: stats,
		ID:     req.ID,
	}
}
