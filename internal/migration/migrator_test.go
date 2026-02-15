package migration

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewMigrator(t *testing.T) {
	m := NewMigrator("/tmp/test-data")
	if m == nil {
		t.Fatal("NewMigrator returned nil")
	}
	if m.dataDir != "/tmp/test-data" {
		t.Errorf("expected dataDir /tmp/test-data, got %s", m.dataDir)
	}
	if m.applied == nil {
		t.Error("expected initialized applied map")
	}
	if len(m.migrations) != 0 {
		t.Errorf("expected 0 migrations, got %d", len(m.migrations))
	}
}

func TestRegisterAndPending(t *testing.T) {
	m := NewMigrator(t.TempDir())

	m.Register(Migration{Version: 2, Description: "second"})
	m.Register(Migration{Version: 1, Description: "first"})

	pending := m.Pending()
	if len(pending) != 2 {
		t.Fatalf("expected 2 pending, got %d", len(pending))
	}
	// Pending should be sorted by version
	if pending[0].Version != 1 {
		t.Errorf("expected first pending version 1, got %d", pending[0].Version)
	}
	if pending[1].Version != 2 {
		t.Errorf("expected second pending version 2, got %d", pending[1].Version)
	}
}

func TestRunMigrations(t *testing.T) {
	dir := t.TempDir()
	m := NewMigrator(dir)
	RegisterDefaultMigrations(m)

	if err := m.Run(); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	// Verify migration 1 created subdirectories
	for _, sub := range []string{"peers", "state", "certs"} {
		info, err := os.Stat(filepath.Join(dir, sub))
		if err != nil {
			t.Errorf("expected %s directory to exist: %v", sub, err)
		} else if !info.IsDir() {
			t.Errorf("expected %s to be a directory", sub)
		}
	}

	// Verify migration 2 created peer metrics file
	metricsPath := filepath.Join(dir, "peers", "metrics.json")
	if _, err := os.Stat(metricsPath); err != nil {
		t.Errorf("expected peers/metrics.json to exist: %v", err)
	}

	// All should be applied now
	pending := m.Pending()
	if len(pending) != 0 {
		t.Errorf("expected 0 pending after Run, got %d", len(pending))
	}
	if m.CurrentVersion() != 2 {
		t.Errorf("expected current version 2, got %d", m.CurrentVersion())
	}
}

func TestRunMigrations_Idempotent(t *testing.T) {
	dir := t.TempDir()
	m := NewMigrator(dir)

	callCount := 0
	m.Register(Migration{
		Version:     1,
		Description: "counting migration",
		Up: func(dataDir string) error {
			callCount++
			return nil
		},
	})

	// Run twice
	if err := m.Run(); err != nil {
		t.Fatalf("first Run() error: %v", err)
	}
	if err := m.Run(); err != nil {
		t.Fatalf("second Run() error: %v", err)
	}

	if callCount != 1 {
		t.Errorf("expected migration to run exactly once, ran %d times", callCount)
	}
}

func TestSaveLoadApplied(t *testing.T) {
	dir := t.TempDir()

	// Create and run migrations
	m1 := NewMigrator(dir)
	RegisterDefaultMigrations(m1)

	if err := m1.Run(); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	// Create a new migrator and load applied state
	m2 := NewMigrator(dir)
	RegisterDefaultMigrations(m2)

	if err := m2.LoadApplied(); err != nil {
		t.Fatalf("LoadApplied() error: %v", err)
	}

	// Should have no pending migrations
	pending := m2.Pending()
	if len(pending) != 0 {
		t.Errorf("expected 0 pending after load, got %d", len(pending))
	}

	if m2.CurrentVersion() != 2 {
		t.Errorf("expected current version 2, got %d", m2.CurrentVersion())
	}
}

func TestCurrentVersion(t *testing.T) {
	m := NewMigrator(t.TempDir())

	// No migrations applied
	if v := m.CurrentVersion(); v != 0 {
		t.Errorf("expected current version 0, got %d", v)
	}

	m.Register(Migration{
		Version:     1,
		Description: "first",
		Up:          func(string) error { return nil },
	})
	m.Register(Migration{
		Version:     3,
		Description: "third",
		Up:          func(string) error { return nil },
	})

	if err := m.Run(); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	if v := m.CurrentVersion(); v != 3 {
		t.Errorf("expected current version 3, got %d", v)
	}
}

func TestLoadApplied_NoFile(t *testing.T) {
	m := NewMigrator(t.TempDir())

	// LoadApplied should succeed even if the file doesn't exist
	if err := m.LoadApplied(); err != nil {
		t.Fatalf("LoadApplied() with no file should not error: %v", err)
	}

	if m.CurrentVersion() != 0 {
		t.Errorf("expected version 0 when no file, got %d", m.CurrentVersion())
	}
}
