package jsruntime

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
)

//go:embed bindings.ts
var bindingsTS []byte

// WriteBindingsFile writes the embedded bindings.ts to a temporary file
// and returns its path. The caller should clean up the file when done.
func WriteBindingsFile(dir string) (string, error) {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", fmt.Errorf("create bindings dir: %w", err)
	}
	path := filepath.Join(dir, "bindings.ts")
	if err := os.WriteFile(path, bindingsTS, 0o600); err != nil {
		return "", fmt.Errorf("write bindings.ts: %w", err)
	}
	return path, nil
}
