//go:build linux

package runtime

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/moltbunker/moltbunker/internal/logging"
)

// AppArmorProfileName is the canonical name of the profile moltbunker ships and
// loads. It matches the `profile <name>` line inside the embedded policy and the
// ContainerSecurityProfile.AppArmorProfile value set by DeploymentSecurityProfile.
const AppArmorProfileName = "moltbunker-container"

// maxProfileBytes caps how much of a profile file the loader will read, guarding
// against a pathological on-disk file being piped to apparmor_parser.
const maxProfileBytes = 64 * 1024

// ErrAppArmorParserMissing is returned when apparmor_parser is not on PATH, so a
// caller (e.g. the doctor checker) can distinguish "no tooling" from "parse failed".
var ErrAppArmorParserMissing = errors.New("apparmor_parser binary not found on PATH")

// execRunner runs an external command and returns its combined output. It is a
// seam so tests can inject a fake parser without touching the real kernel.
type execRunner func(ctx context.Context, name string, args ...string) ([]byte, error)

// defaultExecRunner pipes stdin through to apparmor_parser via exec.CommandContext.
// The profile bytes are supplied to EnsureProfile and written to a temp file that
// is passed as the final argument, so this runner only forwards name+args.
func defaultExecRunner(ctx context.Context, name string, args ...string) ([]byte, error) {
	// #nosec G204 -- name is the constant "apparmor_parser" supplied by EnsureProfile;
	// args are fixed flags plus a daemon-controlled temp file path, never request input.
	return exec.CommandContext(ctx, name, args...).CombinedOutput()
}

// AppArmorLoader loads the moltbunker container AppArmor profile into the kernel.
//
// It is build-tagged Linux-only; the !linux stub is a no-op so darwin builds stay
// green. The embedded profile bytes (embedded_profiles/moltbunker-container.aaprofile)
// make the daemon binary self-contained: an operator does not have to pre-install
// the policy under /etc/apparmor.d, which was the status quo that left the AppArmor
// gate silently disabled on fresh installs (R9).
type AppArmorLoader struct {
	// execRun runs apparmor_parser; nil means use defaultExecRunner.
	execRun execRunner
	// isLoaded reports whether a named profile is already in the kernel; nil means
	// use the package-level isAppArmorProfileLoaded. Tests inject this seam.
	isLoaded func(name string) bool
}

func (l *AppArmorLoader) runner() execRunner {
	if l.execRun != nil {
		return l.execRun
	}
	return defaultExecRunner
}

func (l *AppArmorLoader) loadedCheck() func(string) bool {
	if l.isLoaded != nil {
		return l.isLoaded
	}
	return isAppArmorProfileLoaded
}

// IsProfileLoaded reports whether the named profile is currently loaded in the
// kernel. Delegates to isAppArmorProfileLoaded (reads
// /sys/kernel/security/apparmor/profiles).
func (l *AppArmorLoader) IsProfileLoaded(name string) bool {
	return l.loadedCheck()(name)
}

// EnsureProfile makes sure the named AppArmor profile is loaded into the kernel.
//
//   - If it is already loaded, returns nil immediately (idempotent).
//   - profilePath selects the policy source: when empty, the embedded profile bytes
//     are used (written to a temp file). When non-empty, that file is read (capped
//     at 64KB).
//   - apparmor_parser must be on PATH; otherwise ErrAppArmorParserMissing is returned.
//   - The bytes are loaded with `apparmor_parser -r -W <file>` (replace + wait).
//   - After loading, the kernel is re-checked; a profile that does not appear is a
//     structured error.
func (l *AppArmorLoader) EnsureProfile(ctx context.Context, profileName, profilePath string) error {
	if profileName == "" {
		profileName = AppArmorProfileName
	}

	if l.IsProfileLoaded(profileName) {
		return nil
	}

	parser, err := exec.LookPath("apparmor_parser")
	if err != nil {
		return ErrAppArmorParserMissing
	}

	profileBytes, err := loadProfileBytes(profilePath)
	if err != nil {
		return fmt.Errorf("apparmor: load profile source for %q: %w", profileName, err)
	}

	tmp, err := os.CreateTemp("", "moltbunker-aa-*.profile")
	if err != nil {
		return fmt.Errorf("apparmor: create temp profile file: %w", err)
	}
	tmpPath := tmp.Name()
	defer func() { _ = os.Remove(tmpPath) }()

	if _, werr := tmp.Write(profileBytes); werr != nil {
		_ = tmp.Close()
		return fmt.Errorf("apparmor: write temp profile file: %w", werr)
	}
	if cerr := tmp.Close(); cerr != nil {
		return fmt.Errorf("apparmor: close temp profile file: %w", cerr)
	}

	if out, perr := l.runner()(ctx, parser, "-r", "-W", tmpPath); perr != nil {
		return fmt.Errorf("apparmor: apparmor_parser failed for %q: %s: %w", profileName, trimOutput(out), perr)
	}

	if !l.IsProfileLoaded(profileName) {
		return fmt.Errorf("apparmor: profile %q not present in kernel after load", profileName)
	}

	logging.Info("AppArmor profile loaded",
		"profile", profileName,
		logging.Component("apparmor"))
	return nil
}

// loadProfileBytes returns the profile source bytes. An empty path means use the
// embedded profile; a non-empty path reads from disk with a 64KB cap.
func loadProfileBytes(profilePath string) ([]byte, error) {
	if profilePath == "" {
		return embeddedAppArmorProfile, nil
	}
	// #nosec G304 -- profilePath is a daemon-controlled config/asset path, not request input.
	f, err := os.Open(profilePath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	buf := make([]byte, maxProfileBytes)
	n, err := f.Read(buf)
	if err != nil && n == 0 {
		return nil, err
	}
	return buf[:n], nil
}

func trimOutput(b []byte) string {
	s := string(b)
	if len(s) > 512 {
		s = s[:512]
	}
	return s
}
