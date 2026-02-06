package client

import (
	"encoding/json"
	"fmt"
	"time"
)

// Threat-related types and methods

// ThreatLevelResponse contains threat level information
type ThreatLevelResponse struct {
	Score          float64            `json:"score"`
	Level          string             `json:"level"`
	ActiveSignals  []ThreatSignalInfo `json:"active_signals"`
	Recommendation string             `json:"recommendation"`
	Timestamp      time.Time          `json:"timestamp"`
}

// ThreatSignalInfo contains signal information
type ThreatSignalInfo struct {
	Type       string    `json:"type"`
	Score      float64   `json:"score"`
	Confidence float64   `json:"confidence"`
	Source     string    `json:"source"`
	Details    string    `json:"details"`
	Timestamp  time.Time `json:"timestamp"`
}

// ThreatLevel retrieves the current threat level
func (c *DaemonClient) ThreatLevel() (*ThreatLevelResponse, error) {
	resp, err := c.call("threat_level", nil)
	if err != nil {
		return nil, err
	}

	var level ThreatLevelResponse
	if err := json.Unmarshal(resp.Result, &level); err != nil {
		return nil, fmt.Errorf("failed to parse threat level: %w", err)
	}

	return &level, nil
}

// ThreatSignal reports an external threat signal
func (c *DaemonClient) ThreatSignal(signalType string, confidence float64, source, details string) error {
	_, err := c.call("threat_signal", map[string]interface{}{
		"type":       signalType,
		"confidence": confidence,
		"source":     source,
		"details":    details,
	})
	return err
}

// ThreatClear clears threat signals
func (c *DaemonClient) ThreatClear(signalType string, containerID string, all bool) error {
	_, err := c.call("threat_clear", map[string]interface{}{
		"type":         signalType,
		"container_id": containerID,
		"all":          all,
	})
	return err
}

// Clone-related types and methods

// CloneRequest contains clone parameters
type CloneRequest struct {
	SourceID     string `json:"source_id,omitempty"`
	TargetRegion string `json:"target_region,omitempty"`
	Priority     int    `json:"priority,omitempty"`
	Reason       string `json:"reason,omitempty"`
	IncludeState bool   `json:"include_state,omitempty"`
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

// CloneStats contains cloning statistics
type CloneStats struct {
	ActiveClones   int `json:"active_clones"`
	TotalCompleted int `json:"total_completed"`
	TotalFailed    int `json:"total_failed"`
	MaxConcurrent  int `json:"max_concurrent"`
}

// Clone initiates a clone operation
func (c *DaemonClient) Clone(req *CloneRequest) (*CloneResponse, error) {
	resp, err := c.call("clone", req)
	if err != nil {
		return nil, err
	}

	var clone CloneResponse
	if err := json.Unmarshal(resp.Result, &clone); err != nil {
		return nil, fmt.Errorf("failed to parse clone response: %w", err)
	}

	return &clone, nil
}

// CloneStatus retrieves clone status
func (c *DaemonClient) CloneStatus(cloneID string) (*CloneResponse, error) {
	resp, err := c.call("clone_status", map[string]string{"clone_id": cloneID})
	if err != nil {
		return nil, err
	}

	var clone CloneResponse
	if err := json.Unmarshal(resp.Result, &clone); err != nil {
		return nil, fmt.Errorf("failed to parse clone status: %w", err)
	}

	return &clone, nil
}

// CloneList lists clone operations
func (c *DaemonClient) CloneList(active bool, limit int) ([]CloneResponse, error) {
	resp, err := c.call("clone_list", map[string]interface{}{
		"active": active,
		"limit":  limit,
	})
	if err != nil {
		return nil, err
	}

	var result struct {
		Clones []CloneResponse `json:"clones"`
		Total  int             `json:"total"`
	}
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to parse clone list: %w", err)
	}

	return result.Clones, nil
}

// CloneCancel cancels a clone operation
func (c *DaemonClient) CloneCancel(cloneID string) error {
	_, err := c.call("clone_cancel", map[string]string{"clone_id": cloneID})
	return err
}

// CloneGetStats retrieves clone statistics
func (c *DaemonClient) CloneGetStats() (*CloneStats, error) {
	resp, err := c.call("clone_stats", nil)
	if err != nil {
		return nil, err
	}

	var stats CloneStats
	if err := json.Unmarshal(resp.Result, &stats); err != nil {
		return nil, fmt.Errorf("failed to parse clone stats: %w", err)
	}

	return &stats, nil
}

// Snapshot-related types and methods

// SnapshotRequest contains snapshot creation parameters
type SnapshotRequest struct {
	ContainerID string            `json:"container_id"`
	Type        string            `json:"type,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// SnapshotResponse contains snapshot information
type SnapshotResponse struct {
	ID          string            `json:"id"`
	ContainerID string            `json:"container_id"`
	CreatedAt   time.Time         `json:"created_at"`
	Size        int64             `json:"size"`        // Original size
	StoredSize  int64             `json:"stored_size"` // Size on disk
	Checksum    string            `json:"checksum"`
	Type        string            `json:"type"`
	ParentID    string            `json:"parent_id,omitempty"`
	Compressed  bool              `json:"compressed"`
	Encrypted   bool              `json:"encrypted"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// SnapshotStats contains snapshot statistics
type SnapshotStats struct {
	TotalCount        int     `json:"total_count"`
	TotalSize         int64   `json:"total_size"`          // Stored size
	TotalOriginalSize int64   `json:"total_original_size"` // Before compression
	ContainerCount    int     `json:"container_count"`
	MaxSize           int64   `json:"max_size"`
	MaxPerContainer   int     `json:"max_per_container"`
	CompressionRatio  float64 `json:"compression_ratio"`
	EncryptedCount    int     `json:"encrypted_count"`
}

// SnapshotCreate creates a snapshot
func (c *DaemonClient) SnapshotCreate(req *SnapshotRequest) (*SnapshotResponse, error) {
	resp, err := c.call("snapshot_create", req)
	if err != nil {
		return nil, err
	}

	var snapshot SnapshotResponse
	if err := json.Unmarshal(resp.Result, &snapshot); err != nil {
		return nil, fmt.Errorf("failed to parse snapshot response: %w", err)
	}

	return &snapshot, nil
}

// SnapshotGet retrieves a snapshot
func (c *DaemonClient) SnapshotGet(snapshotID string) (*SnapshotResponse, error) {
	resp, err := c.call("snapshot_get", map[string]string{"snapshot_id": snapshotID})
	if err != nil {
		return nil, err
	}

	var snapshot SnapshotResponse
	if err := json.Unmarshal(resp.Result, &snapshot); err != nil {
		return nil, fmt.Errorf("failed to parse snapshot: %w", err)
	}

	return &snapshot, nil
}

// SnapshotList lists snapshots for a container
func (c *DaemonClient) SnapshotList(containerID string) ([]SnapshotResponse, error) {
	resp, err := c.call("snapshot_list", map[string]string{"container_id": containerID})
	if err != nil {
		return nil, err
	}

	var result struct {
		Snapshots []SnapshotResponse `json:"snapshots"`
		Total     int                `json:"total"`
	}
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to parse snapshot list: %w", err)
	}

	return result.Snapshots, nil
}

// SnapshotDelete deletes a snapshot
func (c *DaemonClient) SnapshotDelete(snapshotID string) error {
	_, err := c.call("snapshot_delete", map[string]string{"snapshot_id": snapshotID})
	return err
}

// SnapshotDeleteAll deletes all snapshots for a container
func (c *DaemonClient) SnapshotDeleteAll(containerID string) error {
	_, err := c.call("snapshot_delete", map[string]interface{}{
		"container_id": containerID,
		"all":          true,
	})
	return err
}

// SnapshotRestore restores from a snapshot
func (c *DaemonClient) SnapshotRestore(snapshotID, targetRegion string, newContainer bool) (map[string]interface{}, error) {
	resp, err := c.call("snapshot_restore", map[string]interface{}{
		"snapshot_id":   snapshotID,
		"target_region": targetRegion,
		"new_container": newContainer,
	})
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to parse restore response: %w", err)
	}

	return result, nil
}

// SnapshotGetStats retrieves snapshot statistics
func (c *DaemonClient) SnapshotGetStats() (*SnapshotStats, error) {
	resp, err := c.call("snapshot_stats", nil)
	if err != nil {
		return nil, err
	}

	var stats SnapshotStats
	if err := json.Unmarshal(resp.Result, &stats); err != nil {
		return nil, fmt.Errorf("failed to parse snapshot stats: %w", err)
	}

	return &stats, nil
}

// RequesterBalanceMap retrieves balance as a generic map (for HTTP API)
func (c *DaemonClient) RequesterBalanceMap() (map[string]interface{}, error) {
	resp, err := c.call("requester_balance", nil)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to parse balance: %w", err)
	}

	return result, nil
}

// SnapshotRotateKey rotates the snapshot encryption key
func (c *DaemonClient) SnapshotRotateKey() error {
	_, err := c.call("snapshot_rotate_key", nil)
	return err
}

// SnapshotExport exports a snapshot to a file
func (c *DaemonClient) SnapshotExport(snapshotID, outputPath string) error {
	_, err := c.call("snapshot_export", map[string]string{
		"snapshot_id": snapshotID,
		"output_path": outputPath,
	})
	return err
}

// SnapshotImport imports a snapshot from a file
func (c *DaemonClient) SnapshotImport(inputPath string) (*SnapshotResponse, error) {
	resp, err := c.call("snapshot_import", map[string]string{
		"input_path": inputPath,
	})
	if err != nil {
		return nil, err
	}

	var snapshot SnapshotResponse
	if err := json.Unmarshal(resp.Result, &snapshot); err != nil {
		return nil, fmt.Errorf("failed to parse snapshot: %w", err)
	}

	return &snapshot, nil
}
