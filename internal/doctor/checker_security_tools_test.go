package doctor

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// emptyPATH points PATH at an empty temp dir so exec.LookPath finds nothing.
func emptyPATH(t *testing.T) {
	t.Helper()
	t.Setenv("PATH", t.TempDir())
}

// fakeBinOnPATH writes an executable script named `name` that prints `output`
// and puts its dir on PATH. Returns nothing; the binary is discoverable by
// exec.LookPath for the duration of the test.
func fakeBinOnPATH(t *testing.T, name, output string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("fake-binary shim relies on a POSIX shell")
	}
	dir := t.TempDir()
	script := "#!/bin/sh\necho \"" + output + "\"\n"
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil { // #nosec G306 -- test fixture must be executable
		t.Fatalf("write fake %s: %v", name, err)
	}
	t.Setenv("PATH", dir)
}

func TestTrivyChecker_AbsentReturnsWarning(t *testing.T) {
	emptyPATH(t)
	res := NewTrivyChecker().Check(context.Background())
	if res.Status != StatusWarning {
		t.Errorf("status = %q, want warning when trivy absent", res.Status)
	}
	if !res.Fixable {
		t.Error("absent trivy should be marked fixable")
	}
}

func TestTrivyChecker_PresentReturnsOK(t *testing.T) {
	fakeBinOnPATH(t, "trivy", "Version: 0.50.0")
	res := NewTrivyChecker().Check(context.Background())
	if res.Status != StatusOK {
		t.Fatalf("status = %q (%s), want ok when trivy present", res.Status, res.Message)
	}
	if res.Message == "" {
		t.Error("present trivy should report a message")
	}
}

func TestNftChecker_NonLinuxReturnsSkipped(t *testing.T) {
	if runtime.GOOS == "linux" {
		t.Skip("nft check is active on linux; skipped-behavior test is for non-linux")
	}
	res := NewNftChecker().Check(context.Background())
	if res.Status != StatusSkipped {
		t.Errorf("status = %q, want skipped on non-linux", res.Status)
	}
}

func TestImageSignatureToolingChecker_AbsentReturnsWarning(t *testing.T) {
	emptyPATH(t)
	res := NewImageSignatureToolingChecker().Check(context.Background())
	if res.Status != StatusWarning {
		t.Errorf("status = %q, want warning when cosign absent", res.Status)
	}
}

func TestImageSignatureToolingChecker_PresentReturnsOK(t *testing.T) {
	fakeBinOnPATH(t, "cosign", "GitVersion: v2.2.0")
	res := NewImageSignatureToolingChecker().Check(context.Background())
	if res.Status != StatusOK {
		t.Fatalf("status = %q (%s), want ok when cosign present", res.Status, res.Message)
	}
}

// TestSecurityToolCheckers_RoleAware confirms all three checkers are
// provider/hybrid scoped so pure requesters do not see them.
func TestSecurityToolCheckers_RoleAware(t *testing.T) {
	checkers := []RoleAware{
		NewTrivyChecker(),
		NewNftChecker(),
		NewImageSignatureToolingChecker(),
	}
	for _, c := range checkers {
		roles := c.Roles()
		if len(roles) != 2 {
			t.Errorf("%T roles = %v, want provider+hybrid", c, roles)
			continue
		}
		want := map[string]bool{"provider": true, "hybrid": true}
		for _, r := range roles {
			if !want[r] {
				t.Errorf("%T unexpected role %q", c, r)
			}
		}
	}
}

// TestSecurityToolCheckers_Interface confirms basic Checker conformance.
func TestSecurityToolCheckers_Interface(t *testing.T) {
	var _ Checker = NewTrivyChecker()
	var _ Checker = NewNftChecker()
	var _ Checker = NewImageSignatureToolingChecker()

	if NewTrivyChecker().Category() != CategoryRuntime {
		t.Error("trivy checker should be runtime category")
	}
	// nft/cosign are not auto-fixable.
	if NewNftChecker().CanFix() {
		t.Error("nft checker should not be fixable")
	}
	if NewImageSignatureToolingChecker().CanFix() {
		t.Error("cosign checker should not be fixable")
	}
}
