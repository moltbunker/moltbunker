package state

import (
	"context"
	"path/filepath"
	"testing"
)

// storeFactory creates a StateStore for testing. The returned store must be
// closed by the caller.
type storeFactory func(t *testing.T) StateStore

// conformanceTests runs the full test suite against any StateStore implementation.
func conformanceTests(t *testing.T, name string, factory storeFactory) {
	t.Run(name+"/PutGetDeployment", func(t *testing.T) {
		s := factory(t)
		defer s.Close()
		ctx := context.Background()

		data := []byte(`{"id":"dep-1","image":"nginx:latest"}`)
		if err := s.PutDeployment(ctx, "dep-1", data); err != nil {
			t.Fatalf("PutDeployment: %v", err)
		}

		got, err := s.GetDeployment(ctx, "dep-1")
		if err != nil {
			t.Fatalf("GetDeployment: %v", err)
		}
		if string(got) != string(data) {
			t.Errorf("got %s, want %s", got, data)
		}
	})

	t.Run(name+"/GetDeploymentMissing", func(t *testing.T) {
		s := factory(t)
		defer s.Close()
		ctx := context.Background()

		got, err := s.GetDeployment(ctx, "nonexistent")
		if err != nil {
			t.Fatalf("GetDeployment: %v", err)
		}
		if got != nil {
			t.Errorf("expected nil for missing key, got %s", got)
		}
	})

	t.Run(name+"/DeleteDeployment", func(t *testing.T) {
		s := factory(t)
		defer s.Close()
		ctx := context.Background()

		if err := s.PutDeployment(ctx, "dep-del", []byte("data")); err != nil {
			t.Fatalf("PutDeployment: %v", err)
		}
		if err := s.DeleteDeployment(ctx, "dep-del"); err != nil {
			t.Fatalf("DeleteDeployment: %v", err)
		}

		got, _ := s.GetDeployment(ctx, "dep-del")
		if got != nil {
			t.Errorf("expected nil after delete, got %s", got)
		}
	})

	t.Run(name+"/DeleteNonexistent", func(t *testing.T) {
		s := factory(t)
		defer s.Close()
		ctx := context.Background()

		// Deleting a key that doesn't exist should not error
		if err := s.DeleteDeployment(ctx, "nonexistent"); err != nil {
			t.Fatalf("DeleteDeployment nonexistent: %v", err)
		}
	})

	t.Run(name+"/ListDeployments", func(t *testing.T) {
		s := factory(t)
		defer s.Close()
		ctx := context.Background()

		if err := s.PutDeployment(ctx, "d1", []byte("one")); err != nil {
			t.Fatalf("PutDeployment d1: %v", err)
		}
		if err := s.PutDeployment(ctx, "d2", []byte("two")); err != nil {
			t.Fatalf("PutDeployment d2: %v", err)
		}
		if err := s.PutDeployment(ctx, "d3", []byte("three")); err != nil {
			t.Fatalf("PutDeployment d3: %v", err)
		}

		all, err := s.ListDeployments(ctx)
		if err != nil {
			t.Fatalf("ListDeployments: %v", err)
		}
		if len(all) != 3 {
			t.Fatalf("expected 3, got %d", len(all))
		}
		if string(all["d1"]) != "one" {
			t.Errorf("d1: got %s", all["d1"])
		}
		if string(all["d2"]) != "two" {
			t.Errorf("d2: got %s", all["d2"])
		}
	})

	t.Run(name+"/ListEmpty", func(t *testing.T) {
		s := factory(t)
		defer s.Close()
		ctx := context.Background()

		all, err := s.ListDeployments(ctx)
		if err != nil {
			t.Fatalf("ListDeployments: %v", err)
		}
		if len(all) != 0 {
			t.Errorf("expected empty map, got %d entries", len(all))
		}
	})

	t.Run(name+"/Overwrite", func(t *testing.T) {
		s := factory(t)
		defer s.Close()
		ctx := context.Background()

		if err := s.PutDeployment(ctx, "dep-1", []byte("v1")); err != nil {
			t.Fatalf("PutDeployment v1: %v", err)
		}
		if err := s.PutDeployment(ctx, "dep-1", []byte("v2")); err != nil {
			t.Fatalf("PutDeployment v2: %v", err)
		}

		got, _ := s.GetDeployment(ctx, "dep-1")
		if string(got) != "v2" {
			t.Errorf("expected v2 after overwrite, got %s", got)
		}
	})

	t.Run(name+"/BansCRUD", func(t *testing.T) {
		s := factory(t)
		defer s.Close()
		ctx := context.Background()

		if err := s.PutBan(ctx, "peer-1", []byte(`{"reason":"spam"}`)); err != nil {
			t.Fatalf("PutBan peer-1: %v", err)
		}
		if err := s.PutBan(ctx, "peer-2", []byte(`{"reason":"abuse"}`)); err != nil {
			t.Fatalf("PutBan peer-2: %v", err)
		}

		all, _ := s.ListBans(ctx)
		if len(all) != 2 {
			t.Fatalf("expected 2 bans, got %d", len(all))
		}

		if err := s.DeleteBan(ctx, "peer-1"); err != nil {
			t.Fatalf("DeleteBan: %v", err)
		}
		all, _ = s.ListBans(ctx)
		if len(all) != 1 {
			t.Fatalf("expected 1 ban after delete, got %d", len(all))
		}
	})

	t.Run(name+"/PeersCRUD", func(t *testing.T) {
		s := factory(t)
		defer s.Close()
		ctx := context.Background()

		if err := s.PutPeer(ctx, "p1", []byte("addr1")); err != nil {
			t.Fatalf("PutPeer p1: %v", err)
		}
		if err := s.PutPeer(ctx, "p2", []byte("addr2")); err != nil {
			t.Fatalf("PutPeer p2: %v", err)
		}

		all, _ := s.ListPeers(ctx)
		if len(all) != 2 {
			t.Fatalf("expected 2 peers, got %d", len(all))
		}

		if err := s.DeletePeer(ctx, "p1"); err != nil {
			t.Fatalf("DeletePeer: %v", err)
		}
		all, _ = s.ListPeers(ctx)
		if len(all) != 1 {
			t.Fatalf("expected 1 peer after delete, got %d", len(all))
		}
	})

	t.Run(name+"/CertPinsCRUD", func(t *testing.T) {
		s := factory(t)
		defer s.Close()
		ctx := context.Background()

		hash := make([]byte, 32)
		hash[0] = 0xAB
		if err := s.PutCertPin(ctx, "node-1", hash); err != nil {
			t.Fatalf("PutCertPin: %v", err)
		}

		all, _ := s.ListCertPins(ctx)
		if len(all) != 1 {
			t.Fatalf("expected 1 cert pin, got %d", len(all))
		}
		if all["node-1"][0] != 0xAB {
			t.Errorf("cert pin data mismatch")
		}

		if err := s.DeleteCertPin(ctx, "node-1"); err != nil {
			t.Fatalf("DeleteCertPin: %v", err)
		}
		all, _ = s.ListCertPins(ctx)
		if len(all) != 0 {
			t.Fatalf("expected 0 cert pins after delete, got %d", len(all))
		}
	})

	t.Run(name+"/APIKeysCRUD", func(t *testing.T) {
		s := factory(t)
		defer s.Close()
		ctx := context.Background()

		if err := s.PutAPIKey(ctx, "key-1", []byte(`{"hash":"abc"}`)); err != nil {
			t.Fatalf("PutAPIKey: %v", err)
		}
		all, _ := s.ListAPIKeys(ctx)
		if len(all) != 1 {
			t.Fatalf("expected 1 api key, got %d", len(all))
		}

		if err := s.DeleteAPIKey(ctx, "key-1"); err != nil {
			t.Fatalf("DeleteAPIKey: %v", err)
		}
		all, _ = s.ListAPIKeys(ctx)
		if len(all) != 0 {
			t.Fatalf("expected 0 after delete, got %d", len(all))
		}
	})

	t.Run(name+"/SchemaVersion", func(t *testing.T) {
		s := factory(t)
		defer s.Close()
		ctx := context.Background()

		v, err := s.SchemaVersion(ctx)
		if err != nil {
			t.Fatalf("SchemaVersion: %v", err)
		}
		if v != 0 {
			t.Errorf("initial schema version should be 0, got %d", v)
		}

		if err := s.SetSchemaVersion(ctx, 1); err != nil {
			t.Fatalf("SetSchemaVersion: %v", err)
		}

		v, _ = s.SchemaVersion(ctx)
		if v != 1 {
			t.Errorf("schema version should be 1, got %d", v)
		}
	})

	t.Run(name+"/Metadata", func(t *testing.T) {
		s := factory(t)
		defer s.Close()
		ctx := context.Background()

		if err := s.PutMeta(ctx, "created_at", []byte("2026-02-26")); err != nil {
			t.Fatalf("PutMeta: %v", err)
		}

		got, err := s.GetMeta(ctx, "created_at")
		if err != nil {
			t.Fatalf("GetMeta: %v", err)
		}
		if string(got) != "2026-02-26" {
			t.Errorf("got %s, want 2026-02-26", got)
		}

		// Missing key returns nil
		got, _ = s.GetMeta(ctx, "nonexistent")
		if got != nil {
			t.Errorf("expected nil for missing meta, got %s", got)
		}
	})

	t.Run(name+"/DataIsolation", func(t *testing.T) {
		s := factory(t)
		defer s.Close()
		ctx := context.Background()

		// Same key name in different buckets should not collide
		if err := s.PutDeployment(ctx, "shared-key", []byte("deployment")); err != nil {
			t.Fatalf("PutDeployment: %v", err)
		}
		if err := s.PutBan(ctx, "shared-key", []byte("ban")); err != nil {
			t.Fatalf("PutBan: %v", err)
		}
		if err := s.PutPeer(ctx, "shared-key", []byte("peer")); err != nil {
			t.Fatalf("PutPeer: %v", err)
		}

		d, _ := s.GetDeployment(ctx, "shared-key")
		if string(d) != "deployment" {
			t.Errorf("deployment bucket: got %s", d)
		}

		bans, _ := s.ListBans(ctx)
		if string(bans["shared-key"]) != "ban" {
			t.Errorf("bans bucket: got %s", bans["shared-key"])
		}

		peers, _ := s.ListPeers(ctx)
		if string(peers["shared-key"]) != "peer" {
			t.Errorf("peers bucket: got %s", peers["shared-key"])
		}
	})
}

func TestMemoryStore(t *testing.T) {
	conformanceTests(t, "MemoryStore", func(t *testing.T) StateStore {
		return NewMemoryStore()
	})
}

func TestBboltStore(t *testing.T) {
	conformanceTests(t, "BboltStore", func(t *testing.T) StateStore {
		path := filepath.Join(t.TempDir(), "test.db")
		s, err := NewBboltStore(path)
		if err != nil {
			t.Fatalf("NewBboltStore: %v", err)
		}
		return s
	})
}

func TestBboltStore_Persistence(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "persist.db")
	ctx := context.Background()

	// Write data, close, reopen, verify
	s1, err := NewBboltStore(path)
	if err != nil {
		t.Fatalf("open 1: %v", err)
	}
	if err := s1.PutDeployment(ctx, "dep-1", []byte("data-1")); err != nil {
		t.Fatalf("PutDeployment: %v", err)
	}
	if err := s1.SetSchemaVersion(ctx, 1); err != nil {
		t.Fatalf("SetSchemaVersion: %v", err)
	}
	s1.Close()

	s2, err := NewBboltStore(path)
	if err != nil {
		t.Fatalf("open 2: %v", err)
	}
	defer s2.Close()

	got, _ := s2.GetDeployment(ctx, "dep-1")
	if string(got) != "data-1" {
		t.Errorf("after reopen: got %s, want data-1", got)
	}

	v, _ := s2.SchemaVersion(ctx)
	if v != 1 {
		t.Errorf("schema version after reopen: got %d, want 1", v)
	}
}
