package state

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestMigrateFromJSON_StateFile(t *testing.T) {
	dataDir := t.TempDir()
	ctx := context.Background()

	// Write a state.json in the legacy format
	stateJSON := `{
  "deployments": {
    "dep-1": {"id": "dep-1", "image": "nginx:latest", "status": "running"},
    "dep-2": {"id": "dep-2", "image": "redis:7", "status": "stopped"}
  },
  "saved_at": "2026-02-26T12:00:00Z",
  "version": 1
}`
	if err := os.WriteFile(filepath.Join(dataDir, "state.json"), []byte(stateJSON), 0600); err != nil {
		t.Fatal(err)
	}

	store := NewMemoryStore()
	defer store.Close()

	if err := MigrateFromJSON(ctx, store, dataDir); err != nil {
		t.Fatalf("MigrateFromJSON: %v", err)
	}

	// Verify deployments migrated
	deps, _ := store.ListDeployments(ctx)
	if len(deps) != 2 {
		t.Fatalf("expected 2 deployments, got %d", len(deps))
	}

	d1, _ := store.GetDeployment(ctx, "dep-1")
	if d1 == nil {
		t.Fatal("dep-1 should exist")
	}

	// Verify schema version set
	v, _ := store.SchemaVersion(ctx)
	if v != CurrentSchemaVersion {
		t.Errorf("schema version: got %d, want %d", v, CurrentSchemaVersion)
	}

	// Verify migrated_from metadata
	mf, _ := store.GetMeta(ctx, MetaMigratedFrom)
	if string(mf) != "json_v1" {
		t.Errorf("migrated_from: got %s, want json_v1", mf)
	}

	// Verify original file renamed to .bak
	if _, err := os.Stat(filepath.Join(dataDir, "state.json")); !os.IsNotExist(err) {
		t.Error("state.json should have been renamed to .bak")
	}
	if _, err := os.Stat(filepath.Join(dataDir, "state.json.bak")); os.IsNotExist(err) {
		t.Error("state.json.bak should exist")
	}
}

func TestMigrateFromJSON_BanList(t *testing.T) {
	dataDir := t.TempDir()
	stateDir := filepath.Join(dataDir, "state")
	if err := os.MkdirAll(stateDir, 0700); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	ctx := context.Background()

	banPath := filepath.Join(stateDir, "banlist.json")
	banJSON := `[
  {"peer_id": "abc123", "reason": "spam", "banned_at": "2026-02-26T00:00:00Z", "expires_at": "2026-03-26T00:00:00Z"},
  {"peer_id": "def456", "reason": "abuse", "banned_at": "2026-02-26T00:00:00Z", "expires_at": "0001-01-01T00:00:00Z"}
]`
	if err := os.WriteFile(banPath, []byte(banJSON), 0600); err != nil {
		t.Fatal(err)
	}

	store := NewMemoryStore()
	defer store.Close()

	if err := MigrateFromJSON(ctx, store, dataDir); err != nil {
		t.Fatalf("MigrateFromJSON: %v", err)
	}

	bans, _ := store.ListBans(ctx)
	if len(bans) != 2 {
		t.Fatalf("expected 2 bans, got %d", len(bans))
	}

	// Auxiliary files should NOT be renamed (subsystems still read from them)
	if _, err := os.Stat(banPath); os.IsNotExist(err) {
		t.Error("banlist.json should NOT have been renamed")
	}
}

func TestMigrateFromJSON_AddressBook(t *testing.T) {
	dataDir := t.TempDir()
	stateDir := filepath.Join(dataDir, "state")
	if err := os.MkdirAll(stateDir, 0700); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	ctx := context.Background()

	abJSON := `[
  {"peer_id": "peer1", "addrs": ["/ip4/1.2.3.4/tcp/9000"], "last_seen": "2026-02-26T00:00:00Z"},
  {"peer_id": "peer2", "addrs": ["/ip4/5.6.7.8/tcp/9000"], "last_seen": "2026-02-25T00:00:00Z"}
]`
	if err := os.WriteFile(filepath.Join(stateDir, "addressbook.json"), []byte(abJSON), 0600); err != nil {
		t.Fatal(err)
	}

	store := NewMemoryStore()
	defer store.Close()

	if err := MigrateFromJSON(ctx, store, dataDir); err != nil {
		t.Fatalf("MigrateFromJSON: %v", err)
	}

	peers, _ := store.ListPeers(ctx)
	if len(peers) != 2 {
		t.Fatalf("expected 2 peers, got %d", len(peers))
	}
}

func TestMigrateFromJSON_CertPins(t *testing.T) {
	dataDir := t.TempDir()
	stateDir := filepath.Join(dataDir, "state")
	if err := os.MkdirAll(stateDir, 0700); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	ctx := context.Background()

	cpJSON := `[
  {"node_id": "aabbccdd", "hash": "11223344"}
]`
	if err := os.WriteFile(filepath.Join(stateDir, "certpins.json"), []byte(cpJSON), 0600); err != nil {
		t.Fatal(err)
	}

	store := NewMemoryStore()
	defer store.Close()

	if err := MigrateFromJSON(ctx, store, dataDir); err != nil {
		t.Fatalf("MigrateFromJSON: %v", err)
	}

	pins, _ := store.ListCertPins(ctx)
	if len(pins) != 1 {
		t.Fatalf("expected 1 cert pin, got %d", len(pins))
	}
}

func TestMigrateFromJSON_APIKeys(t *testing.T) {
	dataDir := t.TempDir()
	ctx := context.Background()

	akJSON := `[
  {"id": "key-1", "name": "test-key", "key_hash": "$2a$10$abc", "enabled": true}
]`
	if err := os.WriteFile(filepath.Join(dataDir, "api_keys.json"), []byte(akJSON), 0600); err != nil {
		t.Fatal(err)
	}

	store := NewMemoryStore()
	defer store.Close()

	if err := MigrateFromJSON(ctx, store, dataDir); err != nil {
		t.Fatalf("MigrateFromJSON: %v", err)
	}

	keys, _ := store.ListAPIKeys(ctx)
	if len(keys) != 1 {
		t.Fatalf("expected 1 api key, got %d", len(keys))
	}
}

func TestMigrateFromJSON_NoFiles(t *testing.T) {
	dataDir := t.TempDir()
	ctx := context.Background()

	store := NewMemoryStore()
	defer store.Close()

	// Should succeed with no files to migrate
	if err := MigrateFromJSON(ctx, store, dataDir); err != nil {
		t.Fatalf("MigrateFromJSON: %v", err)
	}

	// Schema version should still be set
	v, _ := store.SchemaVersion(ctx)
	if v != CurrentSchemaVersion {
		t.Errorf("schema version: got %d, want %d", v, CurrentSchemaVersion)
	}

	// No migrated_from since nothing was migrated
	mf, _ := store.GetMeta(ctx, MetaMigratedFrom)
	if mf != nil {
		t.Errorf("migrated_from should be nil when no files exist, got %s", mf)
	}
}

func TestMigrateFromJSON_AlreadyMigrated(t *testing.T) {
	dataDir := t.TempDir()
	ctx := context.Background()

	// Write a state.json that SHOULD be migrated
	stateJSON := `{"deployments": {"dep-1": {"id": "dep-1"}}, "version": 1}`
	if err := os.WriteFile(filepath.Join(dataDir, "state.json"), []byte(stateJSON), 0600); err != nil {
		t.Fatal(err)
	}

	store := NewMemoryStore()
	defer store.Close()

	// Pre-set schema version to simulate already-migrated
	if err := store.SetSchemaVersion(ctx, CurrentSchemaVersion); err != nil {
		t.Fatalf("SetSchemaVersion: %v", err)
	}

	if err := MigrateFromJSON(ctx, store, dataDir); err != nil {
		t.Fatalf("MigrateFromJSON: %v", err)
	}

	// state.json should NOT have been processed (file still exists, no .bak)
	if _, err := os.Stat(filepath.Join(dataDir, "state.json")); os.IsNotExist(err) {
		t.Error("state.json should still exist (migration skipped)")
	}

	// No deployments should have been imported
	deps, _ := store.ListDeployments(ctx)
	if len(deps) != 0 {
		t.Errorf("expected 0 deployments (already migrated), got %d", len(deps))
	}
}

func TestMigrateFromJSON_BboltStore(t *testing.T) {
	dataDir := t.TempDir()
	stateDir := filepath.Join(dataDir, "state")
	if err := os.MkdirAll(stateDir, 0700); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	ctx := context.Background()

	// Create all legacy files
	if err := os.WriteFile(filepath.Join(dataDir, "state.json"),
		[]byte(`{"deployments":{"d1":{"id":"d1","image":"nginx"}},"version":1}`), 0600); err != nil {
		t.Fatalf("WriteFile state.json: %v", err)
	}
	if err := os.WriteFile(filepath.Join(stateDir, "banlist.json"),
		[]byte(`[{"peer_id":"ban1","reason":"spam"}]`), 0600); err != nil {
		t.Fatalf("WriteFile banlist.json: %v", err)
	}
	if err := os.WriteFile(filepath.Join(stateDir, "addressbook.json"),
		[]byte(`[{"peer_id":"peer1","addrs":["/ip4/1.2.3.4/tcp/9000"]}]`), 0600); err != nil {
		t.Fatalf("WriteFile addressbook.json: %v", err)
	}
	if err := os.WriteFile(filepath.Join(stateDir, "certpins.json"),
		[]byte(`[{"node_id":"node1","hash":"abcd"}]`), 0600); err != nil {
		t.Fatalf("WriteFile certpins.json: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dataDir, "api_keys.json"),
		[]byte(`[{"id":"k1","name":"test"}]`), 0600); err != nil {
		t.Fatalf("WriteFile api_keys.json: %v", err)
	}

	// Use bbolt store
	dbPath := filepath.Join(dataDir, "moltbunker.db")
	store, err := NewBboltStore(dbPath, nil)
	if err != nil {
		t.Fatalf("NewBboltStore: %v", err)
	}
	defer store.Close()

	if err := MigrateFromJSON(ctx, store, dataDir); err != nil {
		t.Fatalf("MigrateFromJSON: %v", err)
	}

	// Verify all data
	deps, _ := store.ListDeployments(ctx)
	if len(deps) != 1 {
		t.Errorf("deployments: got %d, want 1", len(deps))
	}

	bans, _ := store.ListBans(ctx)
	if len(bans) != 1 {
		t.Errorf("bans: got %d, want 1", len(bans))
	}

	peers, _ := store.ListPeers(ctx)
	if len(peers) != 1 {
		t.Errorf("peers: got %d, want 1", len(peers))
	}

	pins, _ := store.ListCertPins(ctx)
	if len(pins) != 1 {
		t.Errorf("cert pins: got %d, want 1", len(pins))
	}

	keys, _ := store.ListAPIKeys(ctx)
	if len(keys) != 1 {
		t.Errorf("api keys: got %d, want 1", len(keys))
	}

	v, _ := store.SchemaVersion(ctx)
	if v != CurrentSchemaVersion {
		t.Errorf("schema version: got %d, want %d", v, CurrentSchemaVersion)
	}
}
