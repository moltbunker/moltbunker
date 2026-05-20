package runtime

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"sync"
	"testing"

	"github.com/moltbunker/moltbunker/pkg/types"
)

func newTestStore(t *testing.T) (*ProfileStore, string) {
	t.Helper()
	dir := t.TempDir()
	s, err := NewProfileStore(dir)
	if err != nil {
		t.Fatalf("NewProfileStore: %v", err)
	}
	return s, dir
}

func TestProfileStore_NewProfileStore_EmptyDirRejected(t *testing.T) {
	_, err := NewProfileStore("")
	if err == nil {
		t.Fatal("expected error for empty directory")
	}
}

func TestProfileStore_NewProfileStore_CreatesDir(t *testing.T) {
	parent := t.TempDir()
	nested := filepath.Join(parent, "profiles")
	if _, err := os.Stat(nested); !os.IsNotExist(err) {
		t.Fatalf("expected %q not to exist yet", nested)
	}
	_, err := NewProfileStore(nested)
	if err != nil {
		t.Fatalf("NewProfileStore: %v", err)
	}
	info, err := os.Stat(nested)
	if err != nil {
		t.Fatalf("dir not created: %v", err)
	}
	if !info.IsDir() {
		t.Fatal("expected directory")
	}
	if info.Mode().Perm() != 0o700 {
		t.Fatalf("expected 0700 perms, got %v", info.Mode().Perm())
	}
}

func TestProfileStore_WriteRead_Roundtrip(t *testing.T) {
	s, _ := newTestStore(t)
	want := types.DeploymentSecurityProfile()

	if err := s.Write("dep-x", want); err != nil {
		t.Fatalf("Write: %v", err)
	}

	got, err := s.Read("dep-x")
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("roundtrip mismatch:\n got: %+v\nwant: %+v", got, want)
	}
}

func TestProfileStore_Read_NotFoundReturnsSentinel(t *testing.T) {
	s, _ := newTestStore(t)
	_, err := s.Read("never-written")
	if err == nil {
		t.Fatal("expected error for missing profile")
	}
	if !errors.Is(err, ErrProfileNotFound) {
		t.Fatalf("err = %v, want ErrProfileNotFound", err)
	}
}

func TestProfileStore_Write_AtomicViaTempFile(t *testing.T) {
	s, dir := newTestStore(t)
	if err := s.Write("dep-x", types.DeploymentSecurityProfile()); err != nil {
		t.Fatalf("Write: %v", err)
	}

	// No .tmp files should remain.
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".tmp" {
			t.Fatalf("leaked tmp file: %s", e.Name())
		}
	}
	// And the final file exists.
	if _, err := os.Stat(filepath.Join(dir, "dep-x.json")); err != nil {
		t.Fatalf("final file missing: %v", err)
	}
}

func TestProfileStore_Write_FilePermissions(t *testing.T) {
	s, dir := newTestStore(t)
	if err := s.Write("dep-x", types.DeploymentSecurityProfile()); err != nil {
		t.Fatalf("Write: %v", err)
	}
	info, err := os.Stat(filepath.Join(dir, "dep-x.json"))
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("expected 0600, got %v", info.Mode().Perm())
	}
}

func TestProfileStore_Delete_RemovesFile(t *testing.T) {
	s, dir := newTestStore(t)
	if err := s.Write("dep-x", types.DeploymentSecurityProfile()); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if err := s.Delete("dep-x"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "dep-x.json")); !os.IsNotExist(err) {
		t.Fatalf("file not deleted: %v", err)
	}
}

func TestProfileStore_Delete_MissingIsNoError(t *testing.T) {
	s, _ := newTestStore(t)
	if err := s.Delete("never-existed"); err != nil {
		t.Fatalf("Delete on missing should be no-op, got %v", err)
	}
}

func TestProfileStore_List_ReturnsContainerIDs(t *testing.T) {
	s, dir := newTestStore(t)
	for _, id := range []string{"dep-a", "dep-b", "dep-c"} {
		if err := s.Write(id, types.DeploymentSecurityProfile()); err != nil {
			t.Fatalf("Write %s: %v", id, err)
		}
	}
	// Sprinkle in a non-JSON file that should be ignored.
	if err := os.WriteFile(filepath.Join(dir, "README.txt"), []byte("ignore me"), 0o600); err != nil {
		t.Fatalf("write decoy: %v", err)
	}

	ids, err := s.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	sort.Strings(ids)
	want := []string{"dep-a", "dep-b", "dep-c"}
	if !reflect.DeepEqual(ids, want) {
		t.Fatalf("List() = %v, want %v", ids, want)
	}
}

func TestProfileStore_Read_RejectsBadSchemaVersion(t *testing.T) {
	s, dir := newTestStore(t)
	// Hand-write a bad-schema file.
	bad := `{"schema_version": 999, "container_id": "dep-x", "profile": {}, "saved_at": "2026-01-01T00:00:00Z"}`
	if err := os.WriteFile(filepath.Join(dir, "dep-x.json"), []byte(bad), 0o600); err != nil {
		t.Fatalf("seed file: %v", err)
	}
	_, err := s.Read("dep-x")
	if err == nil {
		t.Fatal("expected schema-version mismatch error")
	}
}

func TestProfileStore_Read_RejectsMalformedJSON(t *testing.T) {
	s, dir := newTestStore(t)
	if err := os.WriteFile(filepath.Join(dir, "dep-x.json"), []byte("not json"), 0o600); err != nil {
		t.Fatalf("seed file: %v", err)
	}
	_, err := s.Read("dep-x")
	if err == nil {
		t.Fatal("expected unmarshal error")
	}
}

func TestProfileStore_Write_EmptyContainerIDRejected(t *testing.T) {
	s, _ := newTestStore(t)
	if err := s.Write("", types.DeploymentSecurityProfile()); err == nil {
		t.Fatal("expected error for empty containerID")
	}
}

func TestProfileStore_ConcurrentWriteRead(t *testing.T) {
	s, _ := newTestStore(t)
	profile := types.DeploymentSecurityProfile()

	const N = 50
	var wg sync.WaitGroup
	wg.Add(N * 2)

	// N concurrent writers writing to distinct keys
	for i := 0; i < N; i++ {
		go func(i int) {
			defer wg.Done()
			id := containerIDFor(i)
			if err := s.Write(id, profile); err != nil {
				t.Errorf("Write %s: %v", id, err)
			}
		}(i)
	}
	// N concurrent reads of the same set — some will hit ErrProfileNotFound
	// (write race) and that's fine; we only assert no panics and no errors
	// besides not-found.
	for i := 0; i < N; i++ {
		go func(i int) {
			defer wg.Done()
			_, err := s.Read(containerIDFor(i))
			if err != nil && !errors.Is(err, ErrProfileNotFound) {
				t.Errorf("Read: unexpected error %v", err)
			}
		}(i)
	}
	wg.Wait()

	// After all writes, every Read should succeed.
	for i := 0; i < N; i++ {
		_, err := s.Read(containerIDFor(i))
		if err != nil {
			t.Fatalf("post-Write Read(%s): %v", containerIDFor(i), err)
		}
	}
}

func containerIDFor(i int) string {
	return "dep-" + string(rune('a'+(i%26))) + "-" + string(rune('0'+(i/26)%10))
}

func TestProfileStore_SchemaVersionConstant_IsCurrent(t *testing.T) {
	// Tripwire: bumping the constant requires updating tests + migration plan.
	if profileSchemaVersion != 1 {
		t.Fatalf("profileSchemaVersion = %d; tests assume 1. Update migration plan if you bump this.", profileSchemaVersion)
	}
}
