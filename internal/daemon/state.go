package daemon

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/moltbunker/moltbunker/internal/logging"
	"github.com/moltbunker/moltbunker/internal/util"
)

// persistedState represents the state to save to disk
type persistedState struct {
	Deployments map[string]*Deployment `json:"deployments"`
	SavedAt     time.Time              `json:"saved_at"`
	Version     int                    `json:"version"`
}

// stateFilePath returns the path to the state file
func (cm *ContainerManager) stateFilePath() string {
	return filepath.Join(cm.dataDir, "state.json")
}

// saveState saves the current state to disk
func (cm *ContainerManager) saveState() error {
	cm.mu.RLock()
	state := persistedState{
		Deployments: make(map[string]*Deployment, len(cm.deployments)),
		SavedAt:     time.Now(),
		Version:     1,
	}
	for k, v := range cm.deployments {
		// Deep copy to avoid race
		depCopy := *v
		state.Deployments[k] = &depCopy
	}
	cm.mu.RUnlock()

	// Create temp file for atomic write
	tmpPath := cm.stateFilePath() + ".tmp"
	f, err := os.OpenFile(tmpPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to create temp state file: %w", err)
	}

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(state); err != nil {
		f.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("failed to encode state: %w", err)
	}

	if err := f.Sync(); err != nil {
		f.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("failed to sync state file: %w", err)
	}

	if err := f.Close(); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to close state file: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tmpPath, cm.stateFilePath()); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to rename state file: %w", err)
	}

	logging.Debug("state saved",
		"deployments", len(state.Deployments),
		"path", cm.stateFilePath())

	return nil
}

// loadState loads state from disk
func (cm *ContainerManager) loadState() error {
	f, err := os.Open(cm.stateFilePath())
	if err != nil {
		if os.IsNotExist(err) {
			logging.Debug("no state file found, starting fresh")
			return nil
		}
		return fmt.Errorf("failed to open state file: %w", err)
	}
	defer f.Close()

	var state persistedState
	if err := json.NewDecoder(f).Decode(&state); err != nil {
		return fmt.Errorf("failed to decode state: %w", err)
	}

	cm.mu.Lock()
	for id, deployment := range state.Deployments {
		cm.deployments[id] = deployment
		logging.Info("restored deployment from state",
			logging.ContainerID(id),
			"status", string(deployment.Status))
	}
	cm.mu.Unlock()

	logging.Info("state loaded",
		"deployments", len(state.Deployments),
		"saved_at", state.SavedAt)

	return nil
}

// saveStateAsync saves state asynchronously (debounced)
func (cm *ContainerManager) saveStateAsync() {
	util.SafeGoWithName("save-state", func() {
		if err := cm.saveState(); err != nil {
			logging.Error("failed to save state", logging.Err(err))
		}
	})
}
