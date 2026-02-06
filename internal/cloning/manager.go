package cloning

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/moltbunker/moltbunker/internal/logging"
	"github.com/moltbunker/moltbunker/internal/snapshot"
	"github.com/moltbunker/moltbunker/internal/threat"
	"github.com/moltbunker/moltbunker/internal/util"
	"github.com/moltbunker/moltbunker/pkg/types"
)

// CloneStatus represents the status of a clone operation
type CloneStatus string

const (
	CloneStatusPending    CloneStatus = "pending"
	CloneStatusPreparing  CloneStatus = "preparing"
	CloneStatusTransfer   CloneStatus = "transferring"
	CloneStatusDeploying  CloneStatus = "deploying"
	CloneStatusVerifying  CloneStatus = "verifying"
	CloneStatusComplete   CloneStatus = "complete"
	CloneStatusFailed     CloneStatus = "failed"
)

// Clone represents a clone operation
type Clone struct {
	ID            string            `json:"id"`
	SourceID      string            `json:"source_id"`       // Original container ID
	TargetID      string            `json:"target_id"`       // New container ID
	TargetNodeID  types.NodeID      `json:"target_node_id"`
	TargetRegion  string            `json:"target_region"`
	Status        CloneStatus       `json:"status"`
	SnapshotID    string            `json:"snapshot_id,omitempty"`
	CodeHash      string            `json:"code_hash"`
	Priority      int               `json:"priority"`
	Reason        string            `json:"reason"`
	CreatedAt     time.Time         `json:"created_at"`
	CompletedAt   time.Time         `json:"completed_at,omitempty"`
	Error         string            `json:"error,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
}

// CloneConfig configures the cloning system
type CloneConfig struct {
	// Maximum concurrent clone operations
	MaxConcurrentClones int `yaml:"max_concurrent_clones"`

	// Timeout for clone operations
	CloneTimeoutSeconds int `yaml:"clone_timeout_seconds"`

	// Retry configuration
	MaxRetries   int `yaml:"max_retries"`
	RetryDelaySecs int `yaml:"retry_delay_seconds"`

	// Target selection
	PreferredRegions  []string `yaml:"preferred_regions"`
	AvoidSameProvider bool     `yaml:"avoid_same_provider"`
	MinRegionDistance int      `yaml:"min_region_distance"` // Prefer nodes in different regions

	// State transfer
	IncludeState     bool `yaml:"include_state"`
	CompressTransfer bool `yaml:"compress_transfer"`
}

// DefaultCloneConfig returns the default clone configuration
func DefaultCloneConfig() *CloneConfig {
	return &CloneConfig{
		MaxConcurrentClones: 3,
		CloneTimeoutSeconds: 300,
		MaxRetries:          3,
		RetryDelaySecs:      10,
		PreferredRegions:    []string{},
		AvoidSameProvider:   true,
		MinRegionDistance:   1,
		IncludeState:        true,
		CompressTransfer:    true,
	}
}

// NodeSelector provides methods to find suitable target nodes
type NodeSelector interface {
	// FindNodes finds nodes matching criteria
	FindNodes(ctx context.Context, region string, excludeNodes []types.NodeID) ([]*types.Node, error)
	// GetNodeByID returns a specific node
	GetNodeByID(ctx context.Context, nodeID types.NodeID) (*types.Node, error)
}

// ContainerDeployer handles container deployment
type ContainerDeployer interface {
	// DeployClone deploys a cloned container to a target node
	DeployClone(ctx context.Context, clone *Clone, stateData []byte) error
	// VerifyDeployment verifies the clone is running correctly
	VerifyDeployment(ctx context.Context, clone *Clone) error
}

// Manager handles self-cloning operations
type Manager struct {
	config     *CloneConfig
	snapshots  *snapshot.Manager
	selector   NodeSelector
	deployer   ContainerDeployer
	detector   *threat.Detector
	responder  *threat.Responder
	mu         sync.RWMutex
	running    bool
	stopCh     chan struct{}

	// Clone tracking
	activeClones   map[string]*Clone
	cloneHistory   []*Clone
	cloneSemaphore chan struct{}

	// Lineage tracking (original -> clones)
	lineage map[string][]string
}

// NewManager creates a new cloning manager
func NewManager(
	config *CloneConfig,
	snapshots *snapshot.Manager,
	selector NodeSelector,
	deployer ContainerDeployer,
) *Manager {
	if config == nil {
		config = DefaultCloneConfig()
	}

	m := &Manager{
		config:         config,
		snapshots:      snapshots,
		selector:       selector,
		deployer:       deployer,
		activeClones:   make(map[string]*Clone),
		cloneHistory:   make([]*Clone, 0),
		cloneSemaphore: make(chan struct{}, config.MaxConcurrentClones),
		lineage:        make(map[string][]string),
		stopCh:         make(chan struct{}),
	}

	return m
}

// SetThreatDetector sets the threat detector for automatic cloning
func (m *Manager) SetThreatDetector(detector *threat.Detector, responder *threat.Responder) {
	m.detector = detector
	m.responder = responder

	// Register clone callback with responder
	if responder != nil {
		responder.OnClone(func(ctx context.Context, req *threat.CloneRequest) error {
			// Convert threat.CloneRequest to our clone request
			_, err := m.RequestClone(ctx, &CloneRequest{
				SourceID:     req.ContainerID,
				TargetRegion: req.TargetRegion,
				Priority:     req.Priority,
				Reason:       req.Reason,
				IncludeState: true,
				Metadata:     req.Metadata,
			})
			return err
		})
	}
}

// Start begins the cloning service
func (m *Manager) Start(ctx context.Context) error {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return nil
	}
	m.running = true
	m.mu.Unlock()

	logging.Info("cloning manager started",
		"max_concurrent", m.config.MaxConcurrentClones,
		logging.Component("cloning"))

	return nil
}

// Stop halts the cloning service
func (m *Manager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return
	}
	m.running = false
	close(m.stopCh)

	logging.Info("cloning manager stopped", logging.Component("cloning"))
}

// CloneRequest represents a request to clone a container
type CloneRequest struct {
	SourceID     string
	TargetRegion string // "random", "americas", "europe", "asia_pacific", or specific node
	TargetNodeID *types.NodeID
	Priority     int
	Reason       string
	IncludeState bool
	Metadata     map[string]string
}

// RequestClone initiates a clone operation
func (m *Manager) RequestClone(ctx context.Context, req *CloneRequest) (*Clone, error) {
	// Generate clone ID
	idBytes := make([]byte, 8)
	if _, err := rand.Read(idBytes); err != nil {
		return nil, fmt.Errorf("failed to generate clone ID: %w", err)
	}
	cloneID := hex.EncodeToString(idBytes)

	clone := &Clone{
		ID:           cloneID,
		SourceID:     req.SourceID,
		TargetRegion: req.TargetRegion,
		Status:       CloneStatusPending,
		Priority:     req.Priority,
		Reason:       req.Reason,
		CreatedAt:    time.Now(),
		Metadata:     req.Metadata,
	}

	// Add to active clones
	m.mu.Lock()
	m.activeClones[cloneID] = clone
	m.mu.Unlock()

	logging.Info("clone requested",
		"clone_id", cloneID,
		"source_id", req.SourceID,
		"target_region", req.TargetRegion,
		"priority", req.Priority,
		"reason", req.Reason,
		logging.Component("cloning"))

	// Execute clone asynchronously
	util.SafeGoWithName("clone-"+cloneID, func() {
		m.executeClone(ctx, clone, req)
	})

	return clone, nil
}

// executeClone performs the actual clone operation
func (m *Manager) executeClone(ctx context.Context, clone *Clone, req *CloneRequest) {
	// Acquire semaphore
	select {
	case m.cloneSemaphore <- struct{}{}:
		defer func() { <-m.cloneSemaphore }()
	case <-ctx.Done():
		m.failClone(clone, "context cancelled while waiting for slot")
		return
	case <-m.stopCh:
		m.failClone(clone, "cloning manager stopped")
		return
	}

	// Set timeout
	timeout := time.Duration(m.config.CloneTimeoutSeconds) * time.Second
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Step 1: Prepare state snapshot
	m.updateCloneStatus(clone, CloneStatusPreparing)

	var stateData []byte
	if req.IncludeState && m.snapshots != nil {
		snapshot, err := m.snapshots.GetLatestSnapshot(req.SourceID)
		if err == nil {
			clone.SnapshotID = snapshot.ID
			data, err := m.snapshots.GetSnapshotData(snapshot.ID)
			if err == nil {
				stateData = data
			}
		}
		// Not having a snapshot is not fatal
	}

	// Step 2: Select target node
	targetNode, err := m.selectTargetNode(ctx, clone, req)
	if err != nil {
		m.failClone(clone, fmt.Sprintf("failed to select target node: %v", err))
		return
	}
	clone.TargetNodeID = targetNode.ID

	// Step 3: Transfer state to target
	m.updateCloneStatus(clone, CloneStatusTransfer)

	// Step 4: Deploy on target node
	m.updateCloneStatus(clone, CloneStatusDeploying)

	if m.deployer != nil {
		if err := m.deployer.DeployClone(ctx, clone, stateData); err != nil {
			m.failClone(clone, fmt.Sprintf("deployment failed: %v", err))
			return
		}
	}

	// Step 5: Verify deployment
	m.updateCloneStatus(clone, CloneStatusVerifying)

	if m.deployer != nil {
		if err := m.deployer.VerifyDeployment(ctx, clone); err != nil {
			m.failClone(clone, fmt.Sprintf("verification failed: %v", err))
			return
		}
	}

	// Success
	m.completeClone(clone)
}

// selectTargetNode selects an appropriate node for the clone
func (m *Manager) selectTargetNode(ctx context.Context, clone *Clone, req *CloneRequest) (*types.Node, error) {
	if m.selector == nil {
		return nil, fmt.Errorf("no node selector available")
	}

	// If specific node requested
	if req.TargetNodeID != nil {
		return m.selector.GetNodeByID(ctx, *req.TargetNodeID)
	}

	// Determine region
	region := req.TargetRegion
	if region == "random" || region == "" {
		// Pick a different region than source if possible
		regions := []string{"americas", "europe", "asia_pacific"}
		for _, r := range regions {
			if len(m.config.PreferredRegions) > 0 {
				// Use preferred regions
				for _, pr := range m.config.PreferredRegions {
					if pr != region {
						region = pr
						break
					}
				}
			} else {
				region = r
				break
			}
		}
	}

	// Build exclusion list
	var excludeNodes []types.NodeID
	// Could exclude the source node's provider here

	// Find candidates
	nodes, err := m.selector.FindNodes(ctx, region, excludeNodes)
	if err != nil {
		return nil, err
	}

	if len(nodes) == 0 {
		return nil, fmt.Errorf("no suitable nodes found in region %s", region)
	}

	// Select best candidate (for now, just pick first)
	return nodes[0], nil
}

// updateCloneStatus updates the status of a clone operation
func (m *Manager) updateCloneStatus(clone *Clone, status CloneStatus) {
	m.mu.Lock()
	clone.Status = status
	m.mu.Unlock()

	logging.Debug("clone status updated",
		"clone_id", clone.ID,
		"status", status,
		logging.Component("cloning"))
}

// completeClone marks a clone as successfully completed
func (m *Manager) completeClone(clone *Clone) {
	m.mu.Lock()
	clone.Status = CloneStatusComplete
	clone.CompletedAt = time.Now()
	delete(m.activeClones, clone.ID)
	m.cloneHistory = append(m.cloneHistory, clone)

	// Update lineage
	m.lineage[clone.SourceID] = append(m.lineage[clone.SourceID], clone.TargetID)
	m.mu.Unlock()

	logging.Info("clone completed",
		"clone_id", clone.ID,
		"source_id", clone.SourceID,
		"target_id", clone.TargetID,
		"target_node", clone.TargetNodeID.String(),
		"duration", time.Since(clone.CreatedAt).String(),
		logging.Component("cloning"))
}

// failClone marks a clone as failed
func (m *Manager) failClone(clone *Clone, reason string) {
	m.mu.Lock()
	clone.Status = CloneStatusFailed
	clone.Error = reason
	clone.CompletedAt = time.Now()
	delete(m.activeClones, clone.ID)
	m.cloneHistory = append(m.cloneHistory, clone)
	m.mu.Unlock()

	logging.Error("clone failed",
		"clone_id", clone.ID,
		"source_id", clone.SourceID,
		"reason", reason,
		logging.Component("cloning"))
}

// GetClone returns a clone by ID
func (m *Manager) GetClone(cloneID string) (*Clone, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if clone, ok := m.activeClones[cloneID]; ok {
		return clone, true
	}

	for _, clone := range m.cloneHistory {
		if clone.ID == cloneID {
			return clone, true
		}
	}

	return nil, false
}

// GetActiveClones returns all active clone operations
func (m *Manager) GetActiveClones() []*Clone {
	m.mu.RLock()
	defer m.mu.RUnlock()

	clones := make([]*Clone, 0, len(m.activeClones))
	for _, clone := range m.activeClones {
		clones = append(clones, clone)
	}
	return clones
}

// GetCloneHistory returns completed clone operations
func (m *Manager) GetCloneHistory(limit int) []*Clone {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if limit <= 0 || limit > len(m.cloneHistory) {
		limit = len(m.cloneHistory)
	}

	// Return most recent first
	result := make([]*Clone, limit)
	for i := 0; i < limit; i++ {
		result[i] = m.cloneHistory[len(m.cloneHistory)-1-i]
	}
	return result
}

// GetLineage returns all clone IDs derived from a source container
func (m *Manager) GetLineage(sourceID string) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if lineage, ok := m.lineage[sourceID]; ok {
		result := make([]string, len(lineage))
		copy(result, lineage)
		return result
	}
	return nil
}

// CancelClone cancels an active clone operation
func (m *Manager) CancelClone(cloneID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	clone, ok := m.activeClones[cloneID]
	if !ok {
		return fmt.Errorf("clone not found or already completed: %s", cloneID)
	}

	clone.Status = CloneStatusFailed
	clone.Error = "cancelled"
	clone.CompletedAt = time.Now()
	delete(m.activeClones, cloneID)
	m.cloneHistory = append(m.cloneHistory, clone)

	logging.Info("clone cancelled",
		"clone_id", cloneID,
		logging.Component("cloning"))

	return nil
}

// GetStats returns cloning statistics
func (m *Manager) GetStats() CloneStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := CloneStats{
		ActiveClones:    len(m.activeClones),
		TotalCompleted:  0,
		TotalFailed:     0,
		MaxConcurrent:   m.config.MaxConcurrentClones,
	}

	for _, clone := range m.cloneHistory {
		switch clone.Status {
		case CloneStatusComplete:
			stats.TotalCompleted++
		case CloneStatusFailed:
			stats.TotalFailed++
		}
	}

	return stats
}

// CloneStats contains cloning statistics
type CloneStats struct {
	ActiveClones   int `json:"active_clones"`
	TotalCompleted int `json:"total_completed"`
	TotalFailed    int `json:"total_failed"`
	MaxConcurrent  int `json:"max_concurrent"`
}
