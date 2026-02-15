package migration

import (
	"fmt"
	"os"
	"path/filepath"
)

// RegisterDefaultMigrations registers the built-in migration steps on the
// given migrator.
func RegisterDefaultMigrations(m *Migrator) {
	m.Register(Migration{
		Version:     1,
		Description: "Initial schema",
		Up: func(dataDir string) error {
			dirs := []string{"peers", "state", "certs"}
			for _, d := range dirs {
				if err := os.MkdirAll(filepath.Join(dataDir, d), 0o700); err != nil {
					return fmt.Errorf("create %s directory: %w", d, err)
				}
			}
			return nil
		},
	})

	m.Register(Migration{
		Version:     2,
		Description: "Add peer metrics",
		Up: func(dataDir string) error {
			metricsPath := filepath.Join(dataDir, "peers", "metrics.json")
			// Only create the file if it does not already exist
			if _, err := os.Stat(metricsPath); os.IsNotExist(err) {
				if err := os.WriteFile(metricsPath, []byte("{}"), 0o600); err != nil {
					return fmt.Errorf("create peer metrics file: %w", err)
				}
			}
			return nil
		},
	})
}
