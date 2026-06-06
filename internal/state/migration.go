package state

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/moltbunker/moltbunker/internal/logging"
)

// MigrateFromJSON migrates legacy JSON state files to the StateStore.
// It is safe to call even if no JSON files exist (no-op).
//
// Only state.json is renamed to .bak since deployments are now fully served
// by the StateStore. Other JSON files (banlist, addressbook, certpins, apikeys)
// are imported but NOT renamed — their subsystems still read from JSON until
// those packages are updated to use StateStore directly.
func MigrateFromJSON(ctx context.Context, store StateStore, dataDir string) error {
	// Check if migration is needed: if schema version is already set, skip
	v, err := store.SchemaVersion(ctx)
	if err != nil {
		return fmt.Errorf("check schema version: %w", err)
	}
	if v >= CurrentSchemaVersion {
		return nil // Already migrated
	}

	stateDir := filepath.Join(dataDir, "state")
	migrated := false

	// 1. Migrate state.json (deployments)
	statePath := filepath.Join(dataDir, "state.json")
	if n, err := migrateStateJSON(ctx, store, statePath); err != nil {
		logging.Warn("failed to migrate state.json, skipping",
			logging.Err(err), logging.Component("migration"))
	} else if n > 0 {
		migrated = true
		logging.Info("migrated deployments from state.json",
			"count", n, logging.Component("migration"))
	}

	// 2-5: Import auxiliary JSON files into bbolt for forward-compatibility.
	// These files are NOT renamed — their subsystems still read from JSON.
	auxFiles := []struct {
		path     string
		keyField string
		put      putFunc
		label    string
	}{
		{filepath.Join(stateDir, "banlist.json"), "peer_id", store.PutBan, "bans"},
		{filepath.Join(stateDir, "addressbook.json"), "peer_id", store.PutPeer, "peers"},
		{filepath.Join(stateDir, "certpins.json"), "node_id", store.PutCertPin, "cert pins"},
		{filepath.Join(dataDir, "api_keys.json"), "id", store.PutAPIKey, "API keys"},
	}
	for _, f := range auxFiles {
		if n, err := migrateArrayJSON(ctx, store, f.path, f.keyField, f.put); err != nil {
			logging.Warn("failed to import "+f.label+", skipping",
				logging.Err(err), logging.Component("migration"))
		} else if n > 0 {
			migrated = true
			logging.Info("imported "+f.label+" from JSON",
				"count", n, logging.Component("migration"))
		}
	}

	// Set schema version and metadata
	if err := store.SetSchemaVersion(ctx, CurrentSchemaVersion); err != nil {
		return fmt.Errorf("set schema version: %w", err)
	}
	if err := store.PutMeta(ctx, MetaCreatedAt, []byte(time.Now().UTC().Format(time.RFC3339))); err != nil {
		return fmt.Errorf("set created_at metadata: %w", err)
	}
	if migrated {
		if err := store.PutMeta(ctx, MetaMigratedFrom, []byte("json_v1")); err != nil {
			return fmt.Errorf("set migrated_from metadata: %w", err)
		}
	}

	return nil
}

// migrateStateJSON reads the legacy state.json (object with "deployments" map)
// and writes each deployment to the store. Returns the number migrated.
func migrateStateJSON(ctx context.Context, store StateStore, path string) (int, error) {
	// #nosec G304 -- path is a daemon-configured legacy state file (DataDir-derived), not request input
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("read %s: %w", path, err)
	}

	// Parse just enough to extract deployment map
	var wrapper struct {
		Deployments map[string]json.RawMessage `json:"deployments"`
	}
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return 0, fmt.Errorf("parse %s: %w", path, err)
	}

	for id, raw := range wrapper.Deployments {
		if err := store.PutDeployment(ctx, id, raw); err != nil {
			return 0, fmt.Errorf("put deployment %s: %w", id, err)
		}
	}

	// Rename original to .bak
	backupPath := path + ".bak"
	if err := os.Rename(path, backupPath); err != nil {
		logging.Warn("failed to rename state.json to .bak",
			logging.Err(err), logging.Component("migration"))
	}

	return len(wrapper.Deployments), nil
}

// putFunc is the signature for store Put methods.
type putFunc func(ctx context.Context, key string, data []byte) error

// migrateArrayJSON reads a JSON file containing an array of objects,
// extracts the specified key field from each object, and writes each
// entry to the store. The original file is NOT renamed — callers that
// still read from JSON need the file to remain.
func migrateArrayJSON(ctx context.Context, _ StateStore, path, keyField string, put putFunc) (int, error) {
	// #nosec G304 -- path is a daemon-configured legacy state file (DataDir-derived), not request input
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("read %s: %w", path, err)
	}

	var entries []json.RawMessage
	if err := json.Unmarshal(data, &entries); err != nil {
		return 0, fmt.Errorf("parse %s: %w", path, err)
	}

	for i, raw := range entries {
		key, err := extractKey(raw, keyField)
		if err != nil {
			return 0, fmt.Errorf("extract key from entry %d in %s: %w", i, path, err)
		}
		if err := put(ctx, key, raw); err != nil {
			return 0, fmt.Errorf("put entry %d from %s: %w", i, path, err)
		}
	}

	return len(entries), nil
}

// extractKey extracts a string representation of the key field from a JSON object.
// The key field can be a string, number, or JSON array (for types like NodeID [32]byte).
func extractKey(raw json.RawMessage, field string) (string, error) {
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(raw, &obj); err != nil {
		return "", fmt.Errorf("parse object: %w", err)
	}

	keyRaw, ok := obj[field]
	if !ok {
		return "", fmt.Errorf("field %q not found", field)
	}

	// Try string first (most common: id, node_id as hex string)
	var strVal string
	if err := json.Unmarshal(keyRaw, &strVal); err == nil {
		return strVal, nil
	}

	// Fall back to raw JSON as key (handles NodeID [32]byte → JSON int array)
	return string(keyRaw), nil
}
