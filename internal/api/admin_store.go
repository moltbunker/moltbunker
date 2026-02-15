package api

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"

	"github.com/moltbunker/moltbunker/internal/daemon"
	"github.com/moltbunker/moltbunker/internal/logging"
)

// adminNodeIDPattern validates hex node IDs (8-64 chars)
var adminNodeIDPattern = regexp.MustCompile(`^[a-fA-F0-9]{8,64}$`)

// ValidateNodeID checks that a node ID is a valid hex string
func ValidateNodeID(id string) error {
	if !adminNodeIDPattern.MatchString(id) {
		return fmt.Errorf("invalid node ID: must be 8-64 hex characters")
	}
	return nil
}

// Allowed badge values (whitelist)
var allowedBadges = map[string]bool{
	"trusted":  true,
	"verified": true,
	"partner":  true,
}

// ValidateBadges checks that all badges are in the whitelist
func ValidateBadges(badges []string) error {
	if len(badges) > 20 {
		return fmt.Errorf("too many badges (max 20)")
	}
	for _, b := range badges {
		if !allowedBadges[b] {
			return fmt.Errorf("invalid badge: %q", b)
		}
	}
	return nil
}

const (
	maxBlockReasonLen = 1024
	maxNotesLen       = 4096
)

// AdminNodeMetadata contains admin-assigned metadata for a node
type AdminNodeMetadata struct {
	NodeID      string    `json:"node_id"`
	Badges      []string  `json:"badges,omitempty"`       // e.g. ["trusted"]
	Blocked     bool      `json:"blocked"`
	BlockReason string    `json:"block_reason,omitempty"`
	Notes       string    `json:"notes,omitempty"`
	UpdatedAt   time.Time `json:"updated_at"`
	UpdatedBy   string    `json:"updated_by"`             // admin wallet address
}

// AdminMetadataStore persists per-node admin metadata to a JSON file
type AdminMetadataStore struct {
	mu       sync.RWMutex
	nodes    map[string]*AdminNodeMetadata // nodeID → metadata (exact match only)
	filePath string
}

// NewAdminMetadataStore creates a store, loading existing data from disk
func NewAdminMetadataStore(filePath string) *AdminMetadataStore {
	// Ensure parent directory exists with secure permissions
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		logging.Warn("failed to create admin metadata directory",
			"dir", dir,
			"error", err.Error(),
			logging.Component("api"))
	}

	s := &AdminMetadataStore{
		nodes:    make(map[string]*AdminNodeMetadata),
		filePath: filePath,
	}
	if err := s.load(); err != nil && !os.IsNotExist(err) {
		logging.Warn("failed to load admin metadata",
			"error", err.Error(),
			logging.Component("api"))
	}
	return s
}

// load reads metadata from disk
func (s *AdminMetadataStore) load() error {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return err
	}

	var nodes []*AdminNodeMetadata
	if err := json.Unmarshal(data, &nodes); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	for _, n := range nodes {
		s.nodes[n.NodeID] = n
	}
	return nil
}

// saveLocked serializes and writes metadata to disk.
// MUST be called while s.mu is held (Lock or RLock).
func (s *AdminMetadataStore) saveLocked() error {
	// Snapshot data under existing lock
	nodes := make([]*AdminNodeMetadata, 0, len(s.nodes))
	for _, n := range s.nodes {
		nodes = append(nodes, n)
	}

	data, err := json.MarshalIndent(nodes, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.filePath, data, 0600)
}

// Get returns metadata for a node by exact ID match only.
func (s *AdminMetadataStore) Get(nodeID string) *AdminNodeMetadata {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if m, ok := s.nodes[nodeID]; ok {
		return m
	}
	return nil
}

// Set stores metadata for a node and persists to disk
func (s *AdminMetadataStore) Set(meta *AdminNodeMetadata) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	meta.UpdatedAt = time.Now()
	s.nodes[meta.NodeID] = meta
	return s.saveLocked()
}

// Delete removes metadata for a node
func (s *AdminMetadataStore) Delete(nodeID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.nodes, nodeID)
	return s.saveLocked()
}

// GetAll returns all stored metadata
func (s *AdminMetadataStore) GetAll() []*AdminNodeMetadata {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*AdminNodeMetadata, 0, len(s.nodes))
	for _, n := range s.nodes {
		result = append(result, n)
	}
	return result
}

// AsBadgeGetter returns an adapter implementing daemon.AdminBadgeGetter
func (s *AdminMetadataStore) AsBadgeGetter() daemon.AdminBadgeGetter {
	return &adminBadgeAdapter{store: s}
}

// adminBadgeAdapter bridges AdminMetadataStore to daemon.AdminBadgeGetter
type adminBadgeAdapter struct {
	store *AdminMetadataStore
}

func (a *adminBadgeAdapter) Get(nodeID string) *daemon.AdminNodeMeta {
	meta := a.store.Get(nodeID)
	if meta == nil {
		return nil
	}
	return &daemon.AdminNodeMeta{
		Badges:  meta.Badges,
		Blocked: meta.Blocked,
	}
}
