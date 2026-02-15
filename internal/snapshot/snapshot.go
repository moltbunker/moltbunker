package snapshot

import (
	"bytes"
	"compress/gzip"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/moltbunker/moltbunker/internal/logging"
	"github.com/moltbunker/moltbunker/internal/security"
)

// Snapshot represents a point-in-time capture of container state
type Snapshot struct {
	ID           string            `json:"id"`
	ContainerID  string            `json:"container_id"`
	CreatedAt    time.Time         `json:"created_at"`
	Size         int64             `json:"size"`         // Original data size
	StoredSize   int64             `json:"stored_size"`  // Size on disk (after compression/encryption)
	Checksum     string            `json:"checksum"`     // Checksum of original data
	Type         SnapshotType      `json:"type"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	ParentID     string            `json:"parent_id,omitempty"` // For incremental snapshots
	DataPath     string            `json:"data_path"`
	Compressed   bool              `json:"compressed"`
	Encrypted    bool              `json:"encrypted"`
	KeyID        string            `json:"key_id,omitempty"`      // Reference to encryption key
	DeltaBlocks  []DeltaBlock      `json:"delta_blocks,omitempty"` // For incremental snapshots
}

// SnapshotType defines the type of snapshot
type SnapshotType string

const (
	SnapshotTypeFull        SnapshotType = "full"
	SnapshotTypeIncremental SnapshotType = "incremental"
	SnapshotTypeCheckpoint  SnapshotType = "checkpoint"
)

// DeltaBlock represents a changed block in an incremental snapshot
type DeltaBlock struct {
	Offset int64  `json:"offset"` // Offset in original data
	Length int64  `json:"length"` // Length of the block
	Hash   string `json:"hash"`   // Hash of the block for verification
}

// blockSize is the size of blocks for incremental diffing (4KB)
const blockSize = 4096

// SnapshotConfig configures snapshot behavior
type SnapshotConfig struct {
	// Storage settings
	StoragePath      string `yaml:"storage_path"`
	MaxSnapshots     int    `yaml:"max_snapshots"`      // Per container
	MaxTotalSize     int64  `yaml:"max_total_size"`     // Total size in bytes
	CompressionLevel int    `yaml:"compression_level"`  // 0-9, 0 = none

	// Encryption settings
	EncryptionEnabled bool   `yaml:"encryption_enabled"` // Enable encryption at rest
	EncryptionKeyPath string `yaml:"encryption_key_path"` // Path to master encryption key

	// Retention
	RetentionDays int `yaml:"retention_days"` // Days to keep snapshots
}

// DefaultSnapshotConfig returns the default snapshot configuration
func DefaultSnapshotConfig() *SnapshotConfig {
	return &SnapshotConfig{
		StoragePath:       "",  // Will be set to datadir/snapshots
		MaxSnapshots:      10,
		MaxTotalSize:      10 * 1024 * 1024 * 1024, // 10GB
		CompressionLevel:  6,
		EncryptionEnabled: true,
		RetentionDays:     7,
	}
}

// Manager handles snapshot creation and lifecycle
type Manager struct {
	config        *SnapshotConfig
	mu            sync.RWMutex
	snapshots     map[string][]*Snapshot // containerID -> snapshots
	totalSize     int64
	encryptionKey []byte                          // Master encryption key (32 bytes for AES-256)
	blockHashes   map[string]map[int64]string     // containerID -> offset -> hash (for incremental)
}

// NewManager creates a new snapshot manager
func NewManager(config *SnapshotConfig) (*Manager, error) {
	if config == nil {
		config = DefaultSnapshotConfig()
	}

	// Ensure storage directory exists
	if config.StoragePath != "" {
		if err := os.MkdirAll(config.StoragePath, 0755); err != nil {
			return nil, fmt.Errorf("failed to create snapshot storage: %w", err)
		}
	}

	m := &Manager{
		config:      config,
		snapshots:   make(map[string][]*Snapshot),
		blockHashes: make(map[string]map[int64]string),
	}

	// Initialize encryption if enabled
	if config.EncryptionEnabled {
		if err := m.initEncryption(); err != nil {
			return nil, fmt.Errorf("failed to initialize encryption: %w", err)
		}
	}

	// Load existing snapshots
	if err := m.loadSnapshots(); err != nil {
		logging.Warn("failed to load existing snapshots",
			"error", err.Error(),
			logging.Component("snapshot"))
	}

	return m, nil
}

// initEncryption initializes or loads the encryption key
func (m *Manager) initEncryption() error {
	keyPath := m.config.EncryptionKeyPath
	if keyPath == "" && m.config.StoragePath != "" {
		keyPath = filepath.Join(m.config.StoragePath, ".snapshot_key")
	}

	if keyPath == "" {
		// Generate ephemeral key if no path specified
		key, err := security.GenerateKey(32)
		if err != nil {
			return fmt.Errorf("failed to generate encryption key: %w", err)
		}
		m.encryptionKey = key
		logging.Warn("using ephemeral encryption key - snapshots will not be recoverable after restart",
			logging.Component("snapshot"))
		return nil
	}

	// Try to load existing key
	if data, err := os.ReadFile(keyPath); err == nil && len(data) == 32 {
		m.encryptionKey = data
		logging.Debug("loaded existing encryption key",
			logging.Component("snapshot"))
		return nil
	}

	// Generate new key
	key, err := security.GenerateKey(32)
	if err != nil {
		return fmt.Errorf("failed to generate encryption key: %w", err)
	}

	// Save key with restrictive permissions
	if err := os.WriteFile(keyPath, key, 0600); err != nil {
		return fmt.Errorf("failed to save encryption key: %w", err)
	}

	m.encryptionKey = key
	logging.Info("generated new encryption key",
		"path", keyPath,
		logging.Component("snapshot"))

	return nil
}

// compressData compresses data using gzip
func (m *Manager) compressData(data []byte) ([]byte, error) {
	if m.config.CompressionLevel == 0 {
		return data, nil
	}

	var buf bytes.Buffer
	writer, err := gzip.NewWriterLevel(&buf, m.config.CompressionLevel)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip writer: %w", err)
	}

	if _, err := writer.Write(data); err != nil {
		writer.Close()
		return nil, fmt.Errorf("failed to write compressed data: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close gzip writer: %w", err)
	}

	return buf.Bytes(), nil
}

// decompressData decompresses gzip data
func (m *Manager) decompressData(data []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer reader.Close()

	decompressed, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress data: %w", err)
	}

	return decompressed, nil
}

// encryptData encrypts data using AES-256-GCM
func (m *Manager) encryptData(data []byte) ([]byte, error) {
	if !m.config.EncryptionEnabled || m.encryptionKey == nil {
		return data, nil
	}

	return security.EncryptAES256GCM(m.encryptionKey, data)
}

// decryptData decrypts data using AES-256-GCM
func (m *Manager) decryptData(data []byte) ([]byte, error) {
	if !m.config.EncryptionEnabled || m.encryptionKey == nil {
		return data, nil
	}

	return security.DecryptAES256GCM(m.encryptionKey, data)
}

// computeBlockHashes computes hashes for each block of data
func computeBlockHashes(data []byte) map[int64]string {
	hashes := make(map[int64]string)
	for offset := int64(0); offset < int64(len(data)); offset += blockSize {
		end := offset + blockSize
		if end > int64(len(data)) {
			end = int64(len(data))
		}
		block := data[offset:end]
		hash := sha256.Sum256(block)
		hashes[offset] = hex.EncodeToString(hash[:])
	}
	return hashes
}

// computeDelta computes the delta between old and new data
func computeDelta(oldHashes map[int64]string, newData []byte) ([]byte, []DeltaBlock) {
	var deltaData bytes.Buffer
	var deltaBlocks []DeltaBlock

	for offset := int64(0); offset < int64(len(newData)); offset += blockSize {
		end := offset + blockSize
		if end > int64(len(newData)) {
			end = int64(len(newData))
		}
		block := newData[offset:end]
		hash := sha256.Sum256(block)
		hashStr := hex.EncodeToString(hash[:])

		// Check if block has changed
		if oldHash, exists := oldHashes[offset]; !exists || oldHash != hashStr {
			// Write block offset and length to delta
			deltaBlocks = append(deltaBlocks, DeltaBlock{
				Offset: offset,
				Length: int64(len(block)),
				Hash:   hashStr,
			})
			deltaData.Write(block)
		}
	}

	return deltaData.Bytes(), deltaBlocks
}

// applyDelta applies a delta to reconstruct full data
func applyDelta(baseData []byte, deltaData []byte, deltaBlocks []DeltaBlock) ([]byte, error) {
	// Start with a copy of base data
	result := make([]byte, len(baseData))
	copy(result, baseData)

	// Apply each delta block
	deltaOffset := int64(0)
	for _, block := range deltaBlocks {
		// Extend result if needed
		if block.Offset+block.Length > int64(len(result)) {
			extended := make([]byte, block.Offset+block.Length)
			copy(extended, result)
			result = extended
		}

		// Copy delta block data
		if deltaOffset+block.Length > int64(len(deltaData)) {
			return nil, fmt.Errorf("delta data too short for block at offset %d", block.Offset)
		}
		copy(result[block.Offset:], deltaData[deltaOffset:deltaOffset+block.Length])
		deltaOffset += block.Length
	}

	return result, nil
}

// generateKeyID generates a unique key identifier
func generateKeyID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// getStoredSize returns the stored size of a snapshot (for backwards compatibility)
func getStoredSize(snap *Snapshot) int64 {
	if snap.StoredSize > 0 {
		return snap.StoredSize
	}
	return snap.Size
}

// loadSnapshots loads snapshot metadata from disk
func (m *Manager) loadSnapshots() error {
	if m.config.StoragePath == "" {
		return nil
	}

	metaDir := filepath.Join(m.config.StoragePath, "meta")
	if _, err := os.Stat(metaDir); os.IsNotExist(err) {
		return nil
	}

	entries, err := os.ReadDir(metaDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		data, err := os.ReadFile(filepath.Join(metaDir, entry.Name()))
		if err != nil {
			continue
		}

		var snap Snapshot
		if err := json.Unmarshal(data, &snap); err != nil {
			continue
		}

		m.snapshots[snap.ContainerID] = append(m.snapshots[snap.ContainerID], &snap)
		// Use StoredSize if available, otherwise fall back to Size for backwards compatibility
		if snap.StoredSize > 0 {
			m.totalSize += snap.StoredSize
		} else {
			m.totalSize += snap.Size
		}
	}

	return nil
}

// CreateSnapshot creates a new snapshot of container state
func (m *Manager) CreateSnapshot(containerID string, data []byte, snapshotType SnapshotType, metadata map[string]string) (*Snapshot, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if we have room
	if err := m.enforceLimit(containerID); err != nil {
		return nil, fmt.Errorf("failed to enforce limits: %w", err)
	}

	// Generate snapshot ID
	timestamp := time.Now()
	idData := fmt.Sprintf("%s:%d:%d", containerID, timestamp.UnixNano(), len(data))
	hash := sha256.Sum256([]byte(idData))
	snapshotID := hex.EncodeToString(hash[:8])

	// Calculate checksum of original data
	checksum := sha256.Sum256(data)
	checksumStr := hex.EncodeToString(checksum[:])
	originalSize := int64(len(data))

	// Determine parent and handle incremental snapshots
	var parentID string
	var deltaBlocks []DeltaBlock
	var dataToStore []byte

	if snapshotType == SnapshotTypeIncremental {
		// Find latest snapshot for this container
		if existing := m.snapshots[containerID]; len(existing) > 0 {
			parent := existing[len(existing)-1]
			parentID = parent.ID

			// Get parent's block hashes
			if parentHashes, ok := m.blockHashes[containerID]; ok && len(parentHashes) > 0 {
				// Compute delta
				dataToStore, deltaBlocks = computeDelta(parentHashes, data)

				// If delta is larger than 50% of full data, fall back to full snapshot
				if len(dataToStore) > len(data)/2 {
					logging.Debug("incremental snapshot larger than 50%, falling back to full",
						"delta_size", len(dataToStore),
						"full_size", len(data),
						logging.Component("snapshot"))
					snapshotType = SnapshotTypeFull
					parentID = ""
					deltaBlocks = nil
					dataToStore = data
				} else {
					logging.Debug("created incremental snapshot",
						"delta_size", len(dataToStore),
						"full_size", len(data),
						"blocks_changed", len(deltaBlocks),
						logging.Component("snapshot"))
				}
			} else {
				// No parent hashes available, create full snapshot
				snapshotType = SnapshotTypeFull
				parentID = ""
				dataToStore = data
			}
		} else {
			// No parent exists, create full snapshot
			snapshotType = SnapshotTypeFull
			dataToStore = data
		}
	} else {
		dataToStore = data
	}

	// Update block hashes for this container
	m.blockHashes[containerID] = computeBlockHashes(data)

	// Compress data
	compressed := false
	if m.config.CompressionLevel > 0 {
		compressedData, err := m.compressData(dataToStore)
		if err != nil {
			logging.Warn("compression failed, storing uncompressed",
				"error", err.Error(),
				logging.Component("snapshot"))
		} else if len(compressedData) < len(dataToStore) {
			dataToStore = compressedData
			compressed = true
		}
	}

	// Encrypt data
	encrypted := false
	keyID := ""
	if m.config.EncryptionEnabled && m.encryptionKey != nil {
		encryptedData, err := m.encryptData(dataToStore)
		if err != nil {
			return nil, fmt.Errorf("encryption failed: %w", err)
		}
		dataToStore = encryptedData
		encrypted = true
		keyID = generateKeyID()
	}

	// Save snapshot data
	dataPath := ""
	if m.config.StoragePath != "" {
		dataDir := filepath.Join(m.config.StoragePath, "data")
		if err := os.MkdirAll(dataDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create data directory: %w", err)
		}

		// Use different extension for encrypted files
		ext := ".snap"
		if encrypted {
			ext = ".snap.enc"
		}
		dataPath = filepath.Join(dataDir, snapshotID+ext)
		if err := os.WriteFile(dataPath, dataToStore, 0600); err != nil {
			return nil, fmt.Errorf("failed to write snapshot data: %w", err)
		}
	}

	snapshot := &Snapshot{
		ID:          snapshotID,
		ContainerID: containerID,
		CreatedAt:   timestamp,
		Size:        originalSize,
		StoredSize:  int64(len(dataToStore)),
		Checksum:    checksumStr,
		Type:        snapshotType,
		Metadata:    metadata,
		ParentID:    parentID,
		DataPath:    dataPath,
		Compressed:  compressed,
		Encrypted:   encrypted,
		KeyID:       keyID,
		DeltaBlocks: deltaBlocks,
	}

	// Save metadata
	if m.config.StoragePath != "" {
		if err := m.saveSnapshotMeta(snapshot); err != nil {
			return nil, fmt.Errorf("failed to save snapshot metadata: %w", err)
		}
	}

	// Add to in-memory index
	m.snapshots[containerID] = append(m.snapshots[containerID], snapshot)
	m.totalSize += snapshot.StoredSize

	logging.Info("snapshot created",
		"snapshot_id", snapshotID,
		"container_id", containerID,
		"type", snapshotType,
		"original_size", originalSize,
		"stored_size", snapshot.StoredSize,
		"compressed", compressed,
		"encrypted", encrypted,
		logging.Component("snapshot"))

	return snapshot, nil
}

// saveSnapshotMeta saves snapshot metadata to disk
func (m *Manager) saveSnapshotMeta(snapshot *Snapshot) error {
	metaDir := filepath.Join(m.config.StoragePath, "meta")
	if err := os.MkdirAll(metaDir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return err
	}

	metaPath := filepath.Join(metaDir, snapshot.ID+".json")
	return os.WriteFile(metaPath, data, 0600)
}

// GetSnapshot retrieves a snapshot by ID
func (m *Manager) GetSnapshot(snapshotID string) (*Snapshot, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, snapshots := range m.snapshots {
		for _, snap := range snapshots {
			if snap.ID == snapshotID {
				return snap, nil
			}
		}
	}

	return nil, fmt.Errorf("snapshot not found: %s", snapshotID)
}

// GetSnapshotData retrieves the data for a snapshot
func (m *Manager) GetSnapshotData(snapshotID string) ([]byte, error) {
	snapshot, err := m.GetSnapshot(snapshotID)
	if err != nil {
		return nil, err
	}

	return m.getSnapshotDataInternal(snapshot)
}

// getSnapshotDataInternal retrieves snapshot data with decryption, decompression, and delta reconstruction
func (m *Manager) getSnapshotDataInternal(snapshot *Snapshot) ([]byte, error) {
	if snapshot.DataPath == "" {
		return nil, fmt.Errorf("snapshot has no data path")
	}

	data, err := os.ReadFile(snapshot.DataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read snapshot data: %w", err)
	}

	// Decrypt if encrypted
	if snapshot.Encrypted {
		data, err = m.decryptData(data)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt snapshot: %w", err)
		}
	}

	// Decompress if compressed
	if snapshot.Compressed {
		data, err = m.decompressData(data)
		if err != nil {
			return nil, fmt.Errorf("failed to decompress snapshot: %w", err)
		}
	}

	// Handle incremental snapshots - reconstruct from parent + delta
	if snapshot.Type == SnapshotTypeIncremental && snapshot.ParentID != "" {
		parentSnapshot, err := m.GetSnapshot(snapshot.ParentID)
		if err != nil {
			return nil, fmt.Errorf("failed to get parent snapshot: %w", err)
		}

		// Recursively get parent data
		parentData, err := m.getSnapshotDataInternal(parentSnapshot)
		if err != nil {
			return nil, fmt.Errorf("failed to get parent data: %w", err)
		}

		// Apply delta to reconstruct full data
		data, err = applyDelta(parentData, data, snapshot.DeltaBlocks)
		if err != nil {
			return nil, fmt.Errorf("failed to apply delta: %w", err)
		}
	}

	// Verify checksum of reconstructed data
	checksum := sha256.Sum256(data)
	if hex.EncodeToString(checksum[:]) != snapshot.Checksum {
		return nil, fmt.Errorf("snapshot checksum mismatch")
	}

	return data, nil
}

// ListSnapshots returns all snapshots for a container
func (m *Manager) ListSnapshots(containerID string) []*Snapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()

	snapshots := m.snapshots[containerID]
	result := make([]*Snapshot, len(snapshots))
	copy(result, snapshots)

	// Sort by creation time (newest first)
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.After(result[j].CreatedAt)
	})

	return result
}

// GetLatestSnapshot returns the most recent snapshot for a container
func (m *Manager) GetLatestSnapshot(containerID string) (*Snapshot, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	snapshots := m.snapshots[containerID]
	if len(snapshots) == 0 {
		return nil, fmt.Errorf("no snapshots found for container: %s", containerID)
	}

	// Find newest
	var latest *Snapshot
	for _, snap := range snapshots {
		if latest == nil || snap.CreatedAt.After(latest.CreatedAt) {
			latest = snap
		}
	}

	return latest, nil
}

// DeleteSnapshot removes a snapshot
func (m *Manager) DeleteSnapshot(snapshotID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for containerID, snapshots := range m.snapshots {
		for i, snap := range snapshots {
			if snap.ID == snapshotID {
				// Remove from disk
				if snap.DataPath != "" {
					os.Remove(snap.DataPath)
				}
				if m.config.StoragePath != "" {
					metaPath := filepath.Join(m.config.StoragePath, "meta", snapshotID+".json")
					os.Remove(metaPath)
				}

				// Update total size
				m.totalSize -= getStoredSize(snap)

				// Remove from index
				m.snapshots[containerID] = append(snapshots[:i], snapshots[i+1:]...)

				logging.Info("snapshot deleted",
					"snapshot_id", snapshotID,
					"container_id", containerID,
					logging.Component("snapshot"))

				return nil
			}
		}
	}

	return fmt.Errorf("snapshot not found: %s", snapshotID)
}

// DeleteContainerSnapshots removes all snapshots for a container
func (m *Manager) DeleteContainerSnapshots(containerID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	snapshots := m.snapshots[containerID]
	for _, snap := range snapshots {
		if snap.DataPath != "" {
			os.Remove(snap.DataPath)
		}
		if m.config.StoragePath != "" {
			metaPath := filepath.Join(m.config.StoragePath, "meta", snap.ID+".json")
			os.Remove(metaPath)
		}
		m.totalSize -= getStoredSize(snap)
	}

	delete(m.snapshots, containerID)
	delete(m.blockHashes, containerID)

	logging.Info("all container snapshots deleted",
		"container_id", containerID,
		logging.Component("snapshot"))

	return nil
}

// enforceLimit ensures we stay within snapshot limits
func (m *Manager) enforceLimit(containerID string) error {
	snapshots := m.snapshots[containerID]

	// Enforce per-container limit
	for len(snapshots) >= m.config.MaxSnapshots {
		// Remove oldest
		oldest := snapshots[0]
		for _, snap := range snapshots[1:] {
			if snap.CreatedAt.Before(oldest.CreatedAt) {
				oldest = snap
			}
		}

		// Remove from disk
		if oldest.DataPath != "" {
			os.Remove(oldest.DataPath)
		}
		if m.config.StoragePath != "" {
			metaPath := filepath.Join(m.config.StoragePath, "meta", oldest.ID+".json")
			os.Remove(metaPath)
		}

		m.totalSize -= getStoredSize(oldest)

		// Remove from slice
		for i, snap := range snapshots {
			if snap.ID == oldest.ID {
				snapshots = append(snapshots[:i], snapshots[i+1:]...)
				break
			}
		}
		m.snapshots[containerID] = snapshots

		logging.Debug("old snapshot removed due to limit",
			"snapshot_id", oldest.ID,
			"container_id", containerID,
			logging.Component("snapshot"))
	}

	// Enforce total size limit
	for m.totalSize >= m.config.MaxTotalSize {
		// Find oldest snapshot across all containers
		var oldest *Snapshot
		var oldestContainer string

		for cid, snaps := range m.snapshots {
			for _, snap := range snaps {
				if oldest == nil || snap.CreatedAt.Before(oldest.CreatedAt) {
					oldest = snap
					oldestContainer = cid
				}
			}
		}

		if oldest == nil {
			break
		}

		// Remove it
		if oldest.DataPath != "" {
			os.Remove(oldest.DataPath)
		}
		if m.config.StoragePath != "" {
			metaPath := filepath.Join(m.config.StoragePath, "meta", oldest.ID+".json")
			os.Remove(metaPath)
		}

		m.totalSize -= getStoredSize(oldest)

		// Remove from slice
		snaps := m.snapshots[oldestContainer]
		for i, snap := range snaps {
			if snap.ID == oldest.ID {
				m.snapshots[oldestContainer] = append(snaps[:i], snaps[i+1:]...)
				break
			}
		}

		logging.Debug("old snapshot removed due to total size limit",
			"snapshot_id", oldest.ID,
			"container_id", oldestContainer,
			logging.Component("snapshot"))
	}

	return nil
}

// CleanupExpired removes snapshots older than retention period
func (m *Manager) CleanupExpired() int {
	m.mu.Lock()
	defer m.mu.Unlock()

	cutoff := time.Now().AddDate(0, 0, -m.config.RetentionDays)
	removed := 0

	for containerID, snapshots := range m.snapshots {
		var remaining []*Snapshot

		for _, snap := range snapshots {
			if snap.CreatedAt.Before(cutoff) {
				// Remove from disk
				if snap.DataPath != "" {
					os.Remove(snap.DataPath)
				}
				if m.config.StoragePath != "" {
					metaPath := filepath.Join(m.config.StoragePath, "meta", snap.ID+".json")
					os.Remove(metaPath)
				}

				m.totalSize -= getStoredSize(snap)
				removed++

				logging.Debug("expired snapshot removed",
					"snapshot_id", snap.ID,
					"container_id", containerID,
					"age_days", int(time.Since(snap.CreatedAt).Hours()/24),
					logging.Component("snapshot"))
			} else {
				remaining = append(remaining, snap)
			}
		}

		m.snapshots[containerID] = remaining
	}

	if removed > 0 {
		logging.Info("expired snapshots cleaned up",
			"removed", removed,
			logging.Component("snapshot"))
	}

	return removed
}

// GetStats returns snapshot statistics
func (m *Manager) GetStats() SnapshotStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := SnapshotStats{
		TotalSize:       m.totalSize,
		ContainerCount:  len(m.snapshots),
		MaxSize:         m.config.MaxTotalSize,
		MaxPerContainer: m.config.MaxSnapshots,
	}

	for _, snaps := range m.snapshots {
		stats.TotalCount += len(snaps)
	}

	return stats
}

// SnapshotStats contains snapshot statistics
type SnapshotStats struct {
	TotalCount       int     `json:"total_count"`
	TotalSize        int64   `json:"total_size"`         // Total stored size
	TotalOriginalSize int64  `json:"total_original_size"` // Total original size before compression
	ContainerCount   int     `json:"container_count"`
	MaxSize          int64   `json:"max_size"`
	MaxPerContainer  int     `json:"max_per_container"`
	CompressionRatio float64 `json:"compression_ratio"`   // Average compression ratio
	EncryptedCount   int     `json:"encrypted_count"`
}

// GetDetailedStats returns detailed snapshot statistics including compression ratios
func (m *Manager) GetDetailedStats() SnapshotStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := SnapshotStats{
		TotalSize:       m.totalSize,
		ContainerCount:  len(m.snapshots),
		MaxSize:         m.config.MaxTotalSize,
		MaxPerContainer: m.config.MaxSnapshots,
	}

	var totalOriginal int64
	for _, snaps := range m.snapshots {
		stats.TotalCount += len(snaps)
		for _, snap := range snaps {
			totalOriginal += snap.Size
			if snap.Encrypted {
				stats.EncryptedCount++
			}
		}
	}

	stats.TotalOriginalSize = totalOriginal
	if stats.TotalSize > 0 && totalOriginal > 0 {
		stats.CompressionRatio = float64(totalOriginal) / float64(stats.TotalSize)
	}

	return stats
}

// ListAllSnapshots returns all snapshots across all containers
func (m *Manager) ListAllSnapshots() []*Snapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var all []*Snapshot
	for _, snaps := range m.snapshots {
		all = append(all, snaps...)
	}

	// Sort by creation time (newest first)
	sort.Slice(all, func(i, j int) bool {
		return all[i].CreatedAt.After(all[j].CreatedAt)
	})

	return all
}

// ExportSnapshot exports a snapshot to a writer with optional re-encryption
func (m *Manager) ExportSnapshot(snapshotID string, w io.Writer) error {
	snapshot, err := m.GetSnapshot(snapshotID)
	if err != nil {
		return err
	}

	// Read raw data (still encrypted if applicable)
	data, err := os.ReadFile(snapshot.DataPath)
	if err != nil {
		return fmt.Errorf("failed to read snapshot data: %w", err)
	}

	// Write snapshot metadata as JSON header
	meta, err := json.Marshal(snapshot)
	if err != nil {
		return fmt.Errorf("failed to marshal snapshot metadata: %w", err)
	}

	// Write format: [4 bytes meta length][meta JSON][data]
	metaLen := int32(len(meta))
	if err := writeInt32(w, metaLen); err != nil {
		return err
	}
	if _, err := w.Write(meta); err != nil {
		return err
	}
	if _, err := w.Write(data); err != nil {
		return err
	}

	return nil
}

// ImportSnapshot imports a snapshot from a reader
func (m *Manager) ImportSnapshot(r io.Reader) (*Snapshot, error) {
	// Read metadata length
	metaLen, err := readInt32(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata length: %w", err)
	}

	// Read metadata
	meta := make([]byte, metaLen)
	if _, err := io.ReadFull(r, meta); err != nil {
		return nil, fmt.Errorf("failed to read metadata: %w", err)
	}

	var snapshot Snapshot
	if err := json.Unmarshal(meta, &snapshot); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	// Read data
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read data: %w", err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Save data to storage
	if m.config.StoragePath != "" {
		dataDir := filepath.Join(m.config.StoragePath, "data")
		if err := os.MkdirAll(dataDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create data directory: %w", err)
		}

		ext := ".snap"
		if snapshot.Encrypted {
			ext = ".snap.enc"
		}
		snapshot.DataPath = filepath.Join(dataDir, snapshot.ID+ext)
		if err := os.WriteFile(snapshot.DataPath, data, 0600); err != nil {
			return nil, fmt.Errorf("failed to write snapshot data: %w", err)
		}

		if err := m.saveSnapshotMeta(&snapshot); err != nil {
			return nil, fmt.Errorf("failed to save metadata: %w", err)
		}
	}

	// Add to index
	m.snapshots[snapshot.ContainerID] = append(m.snapshots[snapshot.ContainerID], &snapshot)
	m.totalSize += getStoredSize(&snapshot)

	logging.Info("snapshot imported",
		"snapshot_id", snapshot.ID,
		"container_id", snapshot.ContainerID,
		logging.Component("snapshot"))

	return &snapshot, nil
}

// writeInt32 writes a 32-bit integer in big-endian format
func writeInt32(w io.Writer, v int32) error {
	buf := make([]byte, 4)
	buf[0] = byte(v >> 24)
	buf[1] = byte(v >> 16)
	buf[2] = byte(v >> 8)
	buf[3] = byte(v)
	_, err := w.Write(buf)
	return err
}

// readInt32 reads a 32-bit integer in big-endian format
func readInt32(r io.Reader) (int32, error) {
	buf := make([]byte, 4)
	if _, err := io.ReadFull(r, buf); err != nil {
		return 0, err
	}
	return int32(buf[0])<<24 | int32(buf[1])<<16 | int32(buf[2])<<8 | int32(buf[3]), nil
}

// RotateEncryptionKey rotates the encryption key and re-encrypts all snapshots
func (m *Manager) RotateEncryptionKey() error {
	if !m.config.EncryptionEnabled {
		return fmt.Errorf("encryption not enabled")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Generate new key
	newKey, err := security.GenerateKey(32)
	if err != nil {
		return fmt.Errorf("failed to generate new key: %w", err)
	}

	oldKey := m.encryptionKey

	// Re-encrypt all snapshots
	for _, snaps := range m.snapshots {
		for _, snap := range snaps {
			if !snap.Encrypted {
				continue
			}

			// Read with old key
			data, err := os.ReadFile(snap.DataPath)
			if err != nil {
				return fmt.Errorf("failed to read snapshot %s: %w", snap.ID, err)
			}

			// Decrypt with old key
			decrypted, err := security.DecryptAES256GCM(oldKey, data)
			if err != nil {
				return fmt.Errorf("failed to decrypt snapshot %s: %w", snap.ID, err)
			}

			// Re-encrypt with new key
			encrypted, err := security.EncryptAES256GCM(newKey, decrypted)
			if err != nil {
				return fmt.Errorf("failed to re-encrypt snapshot %s: %w", snap.ID, err)
			}

			// Write back
			if err := os.WriteFile(snap.DataPath, encrypted, 0600); err != nil {
				return fmt.Errorf("failed to write snapshot %s: %w", snap.ID, err)
			}

			// Update stored size
			m.totalSize -= getStoredSize(snap)
			snap.StoredSize = int64(len(encrypted))
			snap.KeyID = generateKeyID()
			m.totalSize += getStoredSize(snap)

			// Update metadata
			if err := m.saveSnapshotMeta(snap); err != nil {
				return fmt.Errorf("failed to save metadata for %s: %w", snap.ID, err)
			}
		}
	}

	// Save new key
	m.encryptionKey = newKey
	keyPath := m.config.EncryptionKeyPath
	if keyPath == "" && m.config.StoragePath != "" {
		keyPath = filepath.Join(m.config.StoragePath, ".snapshot_key")
	}
	if keyPath != "" {
		if err := os.WriteFile(keyPath, newKey, 0600); err != nil {
			return fmt.Errorf("failed to save new key: %w", err)
		}
	}

	logging.Info("encryption key rotated",
		logging.Component("snapshot"))

	return nil
}
