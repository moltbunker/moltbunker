package security

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultSeccompProfile(t *testing.T) {
	profile := DefaultSeccompProfile()

	if profile.DefaultAction != "SCMP_ACT_ERRNO" {
		t.Errorf("Default action should be SCMP_ACT_ERRNO, got %s", profile.DefaultAction)
	}

	if len(profile.Syscalls) == 0 {
		t.Error("Should have syscall rules")
	}

	if len(profile.Architectures) == 0 {
		t.Error("Should have architectures")
	}
}

func TestSaveLoadSeccompProfile(t *testing.T) {
	tmpDir := t.TempDir()
	profilePath := filepath.Join(tmpDir, "seccomp.json")

	profile := DefaultSeccompProfile()

	err := SaveSeccompProfile(profile, profilePath)
	if err != nil {
		t.Fatalf("Failed to save seccomp profile: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(profilePath); os.IsNotExist(err) {
		t.Error("Seccomp profile file should exist")
	}

	// Load profile
	loadedProfile, err := LoadSeccompProfile(profilePath)
	if err != nil {
		t.Fatalf("Failed to load seccomp profile: %v", err)
	}

	if loadedProfile.DefaultAction != profile.DefaultAction {
		t.Errorf("Default action mismatch: got %s, want %s", loadedProfile.DefaultAction, profile.DefaultAction)
	}

	if len(loadedProfile.Syscalls) != len(profile.Syscalls) {
		t.Errorf("Syscall count mismatch: got %d, want %d", len(loadedProfile.Syscalls), len(profile.Syscalls))
	}
}

func TestLoadSeccompProfile_NotExists(t *testing.T) {
	_, err := LoadSeccompProfile("/nonexistent/seccomp.json")
	if err == nil {
		t.Error("Should fail for nonexistent file")
	}
}
