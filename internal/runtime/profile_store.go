package runtime

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/moltbunker/moltbunker/pkg/types"
)

// R20 — per-tenant container security profile persistence across daemon restart.
//
// Today, when the daemon restarts and LoadExistingContainers reattaches to
// existing containerd containers, every container gets the default
// DeploymentSecurityProfile because there is no record of the per-tenant
// profile that was originally applied. That's a silent security downgrade
// for tenants who deployed with stricter-than-default policy (exec-disabled,
// custom syscall sets, etc.). The kernel-level enforcement (OCI spec the
// container was created with) is intact — but the daemon-side checks that
// gate the exec terminal endpoint revert to "default allow."
//
// This file provides a JSON-sidecar persistence layer: one
// data_dir/profiles/<containerID>.json file per container. It is read on
// LoadExistingContainers and written on CreateSecureContainer.
//
// Schema: see StoredProfile. Always include schema_version so future field
// changes can migrate without ambiguity.

// profileSchemaVersion is incremented when the on-disk shape changes.
const profileSchemaVersion = 1

// StoredProfile is the JSON shape persisted to disk.
type StoredProfile struct {
	SchemaVersion int                              `json:"schema_version"`
	ContainerID   string                           `json:"container_id"`
	Profile       *types.ContainerSecurityProfile  `json:"profile"`
	SavedAt       time.Time                        `json:"saved_at"`
}

// ProfileStore persists per-container security profiles as JSON sidecars in
// a directory. Methods are safe for concurrent use.
type ProfileStore struct {
	dir string
	mu  sync.RWMutex
}

// NewProfileStore creates a ProfileStore rooted at dir. The directory is
// created with 0700 permissions if it doesn't exist.
func NewProfileStore(dir string) (*ProfileStore, error) {
	if dir == "" {
		return nil, errors.New("profile store: empty directory")
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("profile store: create dir %q: %w", dir, err)
	}
	return &ProfileStore{dir: dir}, nil
}

// path returns the on-disk path for a containerID. It does NOT verify
// existence.
func (s *ProfileStore) path(containerID string) string {
	return filepath.Join(s.dir, containerID+".json")
}

// Write persists the profile for containerID. Writes atomically via
// temp-file-and-rename so a crash mid-write never leaves a corrupt file.
func (s *ProfileStore) Write(containerID string, profile *types.ContainerSecurityProfile) error {
	if containerID == "" {
		return errors.New("profile store: empty containerID")
	}

	stored := StoredProfile{
		SchemaVersion: profileSchemaVersion,
		ContainerID:   containerID,
		Profile:       profile,
		SavedAt:       time.Now().UTC(),
	}
	data, err := json.MarshalIndent(stored, "", "  ")
	if err != nil {
		return fmt.Errorf("profile store: marshal: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	finalPath := s.path(containerID)
	tmpPath := finalPath + ".tmp"

	if err := os.WriteFile(tmpPath, data, 0o600); err != nil {
		return fmt.Errorf("profile store: write tmp: %w", err)
	}
	if err := os.Rename(tmpPath, finalPath); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("profile store: rename tmp: %w", err)
	}
	return nil
}

// ErrProfileNotFound is returned by Read when no profile exists for a
// containerID. Callers should fall back to the default profile.
var ErrProfileNotFound = errors.New("profile store: profile not found")

// Read loads the profile for containerID. Returns ErrProfileNotFound if no
// sidecar exists — callers should fall back to a default profile when they
// see this error.
//
// On malformed JSON, the caller gets the parse error; the file is NOT
// auto-deleted. Operators decide whether to delete it manually.
func (s *ProfileStore) Read(containerID string) (*types.ContainerSecurityProfile, error) {
	if containerID == "" {
		return nil, errors.New("profile store: empty containerID")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := os.ReadFile(s.path(containerID))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrProfileNotFound
		}
		return nil, fmt.Errorf("profile store: read: %w", err)
	}

	var stored StoredProfile
	if err := json.Unmarshal(data, &stored); err != nil {
		return nil, fmt.Errorf("profile store: unmarshal %q: %w", containerID, err)
	}
	if stored.SchemaVersion != profileSchemaVersion {
		return nil, fmt.Errorf("profile store: unsupported schema_version %d for %q (want %d)",
			stored.SchemaVersion, containerID, profileSchemaVersion)
	}
	return stored.Profile, nil
}

// Delete removes the sidecar for containerID. It is safe to call when no
// profile exists; only filesystem errors other than "not found" are
// returned.
func (s *ProfileStore) Delete(containerID string) error {
	if containerID == "" {
		return errors.New("profile store: empty containerID")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	err := os.Remove(s.path(containerID))
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("profile store: delete: %w", err)
	}
	return nil
}

// List returns the container IDs that currently have stored profiles. Useful
// on daemon startup to detect orphans (profiles whose containers no longer
// exist).
func (s *ProfileStore) List() ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return nil, fmt.Errorf("profile store: readdir: %w", err)
	}

	var ids []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if filepath.Ext(name) != ".json" {
			continue
		}
		ids = append(ids, name[:len(name)-len(".json")])
	}
	return ids, nil
}
