//go:build linux

package runtime

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnsureProfile_AlreadyLoaded(t *testing.T) {
	called := false
	l := &AppArmorLoader{
		isLoaded: func(string) bool { return true },
		execRun: func(context.Context, string, ...string) ([]byte, error) {
			called = true
			return nil, nil
		},
	}
	if err := l.EnsureProfile(context.Background(), AppArmorProfileName, ""); err != nil {
		t.Fatalf("EnsureProfile returned error when already loaded: %v", err)
	}
	if called {
		t.Error("apparmor_parser should NOT be invoked when the profile is already loaded")
	}
}

func TestEnsureProfile_ParserMissing(t *testing.T) {
	// Point PATH at an empty dir so exec.LookPath("apparmor_parser") fails.
	t.Setenv("PATH", t.TempDir())
	l := &AppArmorLoader{isLoaded: func(string) bool { return false }}
	err := l.EnsureProfile(context.Background(), AppArmorProfileName, "")
	if !errors.Is(err, ErrAppArmorParserMissing) {
		t.Fatalf("expected ErrAppArmorParserMissing, got %v", err)
	}
}

func TestEnsureProfile_ParseFails(t *testing.T) {
	installFakeParser(t)
	l := &AppArmorLoader{
		isLoaded: func(string) bool { return false },
		execRun: func(_ context.Context, _ string, _ ...string) ([]byte, error) {
			return []byte("AppArmor parser error: syntax error near line 3"), errors.New("exit status 1")
		},
	}
	err := l.EnsureProfile(context.Background(), AppArmorProfileName, "")
	if err == nil {
		t.Fatal("expected error when apparmor_parser fails")
	}
	if !strings.Contains(err.Error(), "syntax error") {
		t.Errorf("error should wrap parser stderr, got %v", err)
	}
}

func TestEnsureProfile_NotPresentAfterLoad(t *testing.T) {
	installFakeParser(t)
	l := &AppArmorLoader{
		isLoaded: func(string) bool { return false }, // never appears
		execRun: func(context.Context, string, ...string) ([]byte, error) {
			return nil, nil // parser "succeeds"
		},
	}
	err := l.EnsureProfile(context.Background(), AppArmorProfileName, "")
	if err == nil || !strings.Contains(err.Error(), "not present in kernel after load") {
		t.Fatalf("expected post-load presence error, got %v", err)
	}
}

func TestEnsureProfile_Success(t *testing.T) {
	installFakeParser(t)

	var gotArgs []string
	loadedNow := false
	l := &AppArmorLoader{
		isLoaded: func(string) bool { return loadedNow },
		execRun: func(_ context.Context, _ string, args ...string) ([]byte, error) {
			gotArgs = args
			loadedNow = true // becomes present after a successful parse
			return nil, nil
		},
	}
	if err := l.EnsureProfile(context.Background(), AppArmorProfileName, ""); err != nil {
		t.Fatalf("EnsureProfile failed: %v", err)
	}
	// args should be: -r -W <tmpfile>
	if len(gotArgs) != 3 || gotArgs[0] != "-r" || gotArgs[1] != "-W" {
		t.Fatalf("unexpected apparmor_parser args: %v", gotArgs)
	}
	if _, err := os.Stat(gotArgs[2]); !os.IsNotExist(err) {
		t.Errorf("temp profile file should be removed after load, stat err = %v", err)
	}
}

func TestEnsureProfile_ReadsProfilePath(t *testing.T) {
	installFakeParser(t)

	dir := t.TempDir()
	path := filepath.Join(dir, "custom.profile")
	const body = "profile custom-test flags=(complain) { }\n"
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}

	loadedNow := false
	l := &AppArmorLoader{
		isLoaded: func(string) bool { return loadedNow },
		execRun: func(_ context.Context, _ string, args ...string) ([]byte, error) {
			// Verify the temp file passed to the parser holds the file's bytes.
			data, _ := os.ReadFile(args[len(args)-1])
			if string(data) != body {
				t.Errorf("parser fed %q, want %q", string(data), body)
			}
			loadedNow = true
			return nil, nil
		},
	}
	if err := l.EnsureProfile(context.Background(), "custom-test", path); err != nil {
		t.Fatalf("EnsureProfile(path) failed: %v", err)
	}
}

func TestEmbeddedAppArmorProfile_Content(t *testing.T) {
	b := EmbeddedAppArmorProfile()
	if len(b) == 0 {
		t.Fatal("embedded AppArmor profile is empty")
	}
	s := string(b)
	for _, want := range []string{"#include <tunables/global>", "profile moltbunker-container"} {
		if !strings.Contains(s, want) {
			t.Errorf("embedded profile missing %q", want)
		}
	}
}

func TestIsProfileLoaded_DelegatesToSeam(t *testing.T) {
	l := &AppArmorLoader{isLoaded: func(name string) bool { return name == "x" }}
	if !l.IsProfileLoaded("x") {
		t.Error("IsProfileLoaded should report true for the seam's match")
	}
	if l.IsProfileLoaded("y") {
		t.Error("IsProfileLoaded should report false for a non-match")
	}
}

// installFakeParser puts a dummy `apparmor_parser` executable on PATH so
// exec.LookPath succeeds; the actual run is intercepted by the execRun seam.
func installFakeParser(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	bin := filepath.Join(dir, "apparmor_parser")
	if err := os.WriteFile(bin, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil { // #nosec G306 -- test fixture must be executable
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)
}
