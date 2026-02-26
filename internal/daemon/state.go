package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/moltbunker/moltbunker/internal/logging"
	"github.com/moltbunker/moltbunker/internal/util"
)

// persistedState represents the legacy JSON state format (kept for backward compat)
type persistedState struct {
	Deployments map[string]*Deployment `json:"deployments"`
	SavedAt     time.Time              `json:"saved_at"`
	Version     int                    `json:"version"`
}

// stateFilePath returns the path to the legacy state file
func (cm *ContainerManager) stateFilePath() string {
	return filepath.Join(cm.dataDir, "state.json")
}

// saveState saves the current state to the StateStore (bbolt) or falls back to JSON.
func (cm *ContainerManager) saveState() error {
	if cm.stateStore != nil {
		return cm.saveStateBbolt()
	}
	return cm.saveStateJSON()
}

// loadState loads state from the StateStore (bbolt) or falls back to JSON.
func (cm *ContainerManager) loadState() error {
	if cm.stateStore != nil {
		return cm.loadStateBbolt()
	}
	return cm.loadStateJSON()
}

// saveStateAsync saves state asynchronously (debounced)
func (cm *ContainerManager) saveStateAsync() {
	util.SafeGoWithName("save-state", func() {
		if err := cm.saveState(); err != nil {
			logging.Error("failed to save state", logging.Err(err))
		}
	})
}

// deleteDeploymentState removes a single deployment from persistent storage.
// More efficient than re-saving all deployments.
func (cm *ContainerManager) deleteDeploymentState(id string) {
	if cm.stateStore != nil {
		if err := cm.stateStore.DeleteDeployment(context.Background(), id); err != nil {
			logging.Error("failed to delete deployment from state store",
				logging.ContainerID(id), logging.Err(err))
		}
		return
	}
	// Fallback: full re-save
	cm.saveStateAsync()
}

// --- bbolt implementation ---

func (cm *ContainerManager) saveStateBbolt() error {
	cm.mu.RLock()
	deployments := make(map[string]*Deployment, len(cm.deployments))
	for k, v := range cm.deployments {
		depCopy := *v
		if len(v.Regions) > 0 {
			depCopy.Regions = make([]string, len(v.Regions))
			copy(depCopy.Regions, v.Regions)
		}
		if len(v.Locations) > 0 {
			depCopy.Locations = make([]ReplicaLocation, len(v.Locations))
			copy(depCopy.Locations, v.Locations)
		}
		deployments[k] = &depCopy
	}
	cm.mu.RUnlock()

	ctx := context.Background()
	for id, dep := range deployments {
		data, err := json.Marshal(dep)
		if err != nil {
			return fmt.Errorf("marshal deployment %s: %w", id, err)
		}
		if err := cm.stateStore.PutDeployment(ctx, id, data); err != nil {
			return fmt.Errorf("put deployment %s: %w", id, err)
		}
	}

	logging.Debug("state saved to bbolt", "deployments", len(deployments))
	return nil
}

func (cm *ContainerManager) loadStateBbolt() error {
	ctx := context.Background()
	all, err := cm.stateStore.ListDeployments(ctx)
	if err != nil {
		return fmt.Errorf("list deployments: %w", err)
	}

	cm.mu.Lock()
	for id, data := range all {
		var dep Deployment
		if err := json.Unmarshal(data, &dep); err != nil {
			logging.Warn("skipping corrupt deployment in state store",
				logging.ContainerID(id), logging.Err(err))
			continue
		}
		cm.deployments[id] = &dep
		logging.Info("restored deployment from state",
			logging.ContainerID(id),
			"status", string(dep.Status))
	}
	cm.mu.Unlock()

	logging.Info("state loaded from bbolt", "deployments", len(all))
	return nil
}

// --- legacy JSON implementation (fallback) ---

func (cm *ContainerManager) saveStateJSON() error {
	cm.mu.RLock()
	state := persistedState{
		Deployments: make(map[string]*Deployment, len(cm.deployments)),
		SavedAt:     time.Now(),
		Version:     1,
	}
	for k, v := range cm.deployments {
		depCopy := *v
		if len(v.Regions) > 0 {
			depCopy.Regions = make([]string, len(v.Regions))
			copy(depCopy.Regions, v.Regions)
		}
		if len(v.Locations) > 0 {
			depCopy.Locations = make([]ReplicaLocation, len(v.Locations))
			copy(depCopy.Locations, v.Locations)
		}
		state.Deployments[k] = &depCopy
	}
	cm.mu.RUnlock()

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

	if err := os.Rename(tmpPath, cm.stateFilePath()); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to rename state file: %w", err)
	}

	logging.Debug("state saved",
		"deployments", len(state.Deployments),
		"path", cm.stateFilePath())

	return nil
}

func (cm *ContainerManager) loadStateJSON() error {
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
