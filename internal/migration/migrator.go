package migration

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

// Migration represents a single versioned migration step.
type Migration struct {
	Version     int
	Description string
	Up          func(dataDir string) error
}

// Migrator tracks and executes ordered migrations against a data directory.
type Migrator struct {
	dataDir    string
	applied    map[int]time.Time // version -> applied time
	migrations []Migration
	mu         sync.Mutex
}

// appliedRecord is the JSON-serialisable form of a single applied migration.
type appliedRecord struct {
	Version   int       `json:"version"`
	AppliedAt time.Time `json:"applied_at"`
}

// NewMigrator creates a new Migrator for the given data directory.
func NewMigrator(dataDir string) *Migrator {
	return &Migrator{
		dataDir: dataDir,
		applied: make(map[int]time.Time),
	}
}

// Register adds a migration to the migrator. Migrations are sorted by version
// before execution, so registration order does not matter.
func (m *Migrator) Register(migration Migration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.migrations = append(m.migrations, migration)
}

// migrationsFilePath returns the path to the migrations state file.
func (m *Migrator) migrationsFilePath() string {
	return filepath.Join(m.dataDir, "migrations.json")
}

// LoadApplied reads previously applied migrations from disk.
// If the file does not exist the applied set is left empty (no error).
func (m *Migrator) LoadApplied() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := os.ReadFile(m.migrationsFilePath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read migrations file: %w", err)
	}

	var records []appliedRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return fmt.Errorf("parse migrations file: %w", err)
	}

	applied := make(map[int]time.Time, len(records))
	for _, r := range records {
		applied[r.Version] = r.AppliedAt
	}
	m.applied = applied
	return nil
}

// SaveApplied writes the current applied state to disk using an atomic
// write-to-temp-then-rename pattern.
func (m *Migrator) SaveApplied() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.saveAppliedLocked()
}

// saveAppliedLocked performs the actual save; caller must hold m.mu.
func (m *Migrator) saveAppliedLocked() error {
	records := make([]appliedRecord, 0, len(m.applied))
	for v, t := range m.applied {
		records = append(records, appliedRecord{Version: v, AppliedAt: t})
	}
	// Sort for deterministic output
	sort.Slice(records, func(i, j int) bool {
		return records[i].Version < records[j].Version
	})

	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal migrations: %w", err)
	}

	// Ensure the data directory exists
	if err := os.MkdirAll(m.dataDir, 0o700); err != nil {
		return fmt.Errorf("create data dir: %w", err)
	}

	// Atomic write: write to temp file then rename
	tmpPath := m.migrationsFilePath() + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0o600); err != nil {
		return fmt.Errorf("write temp migrations file: %w", err)
	}
	if err := os.Rename(tmpPath, m.migrationsFilePath()); err != nil {
		return fmt.Errorf("rename migrations file: %w", err)
	}
	return nil
}

// Pending returns all registered migrations that have not yet been applied,
// sorted by version ascending.
func (m *Migrator) Pending() []Migration {
	m.mu.Lock()
	defer m.mu.Unlock()

	var pending []Migration
	for _, mig := range m.migrations {
		if _, ok := m.applied[mig.Version]; !ok {
			pending = append(pending, mig)
		}
	}
	sort.Slice(pending, func(i, j int) bool {
		return pending[i].Version < pending[j].Version
	})
	return pending
}

// Run executes all pending migrations in version order.
// After each successful migration the applied state is persisted to disk.
func (m *Migrator) Run() error {
	pending := m.Pending()
	for _, mig := range pending {
		if err := mig.Up(m.dataDir); err != nil {
			return fmt.Errorf("migration v%d (%s): %w", mig.Version, mig.Description, err)
		}

		m.mu.Lock()
		m.applied[mig.Version] = time.Now()
		if err := m.saveAppliedLocked(); err != nil {
			// Roll back in-memory state to stay consistent with disk
			delete(m.applied, mig.Version)
			m.mu.Unlock()
			return fmt.Errorf("save after migration v%d: %w", mig.Version, err)
		}
		m.mu.Unlock()
	}
	return nil
}

// CurrentVersion returns the highest applied migration version, or 0 if no
// migrations have been applied.
func (m *Migrator) CurrentVersion() int {
	m.mu.Lock()
	defer m.mu.Unlock()

	max := 0
	for v := range m.applied {
		if v > max {
			max = v
		}
	}
	return max
}
