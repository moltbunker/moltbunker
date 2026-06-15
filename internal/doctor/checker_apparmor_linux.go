//go:build linux

package doctor

import (
	"context"
	"os"
	"os/exec"

	"github.com/moltbunker/moltbunker/internal/runtime"
)

// AppArmorChecker verifies the moltbunker-container AppArmor profile is loaded so
// the runtime's AppArmor confinement gate actually fires (HARDEN-01, R9). It can
// auto-fix by loading the embedded profile via the AppArmorLoader.
type AppArmorChecker struct {
	// statFn / lookPathFn are seams for tests; nil means use the real os/exec funcs.
	statFn     func(string) (os.FileInfo, error)
	lookPathFn func(string) (string, error)
	loader     appArmorLoader
}

// appArmorLoader is the subset of runtime.AppArmorLoader the checker depends on,
// so tests can inject a fake without a real kernel.
type appArmorLoader interface {
	IsProfileLoaded(name string) bool
	EnsureProfile(ctx context.Context, profileName, profilePath string) error
}

func NewAppArmorChecker() *AppArmorChecker {
	return &AppArmorChecker{
		statFn:     os.Stat,
		lookPathFn: exec.LookPath,
		loader:     &runtime.AppArmorLoader{},
	}
}

func (c *AppArmorChecker) Name() string       { return "AppArmor profile" }
func (c *AppArmorChecker) Category() Category { return CategorySecurity }
func (c *AppArmorChecker) CanFix() bool       { return true }
func (c *AppArmorChecker) Roles() []string    { return []string{"provider", "hybrid"} }

func (c *AppArmorChecker) stat(p string) (os.FileInfo, error) {
	if c.statFn != nil {
		return c.statFn(p)
	}
	return os.Stat(p)
}

func (c *AppArmorChecker) lookPath(name string) (string, error) {
	if c.lookPathFn != nil {
		return c.lookPathFn(name)
	}
	return exec.LookPath(name)
}

func (c *AppArmorChecker) Check(_ context.Context) CheckResult {
	result := CheckResult{Name: c.Name(), Category: c.Category()}

	// (1) Is AppArmor active on this kernel at all?
	if _, err := c.stat("/sys/kernel/security/apparmor"); err != nil {
		result.Status = StatusSkipped
		result.Message = "AppArmor: not available on this kernel/distro"
		result.Details = "AppArmor LSM not present (/sys/kernel/security/apparmor missing). " +
			"On SELinux hosts, container confinement is provided by container-selinux instead."
		return result
	}

	// (2) Profile already loaded → all good.
	if c.loader.IsProfileLoaded(runtime.AppArmorProfileName) {
		result.Status = StatusOK
		result.Message = "AppArmor: moltbunker-container profile loaded"
		return result
	}

	// (3) Not loaded — is the parser available to load it?
	if _, err := c.lookPath("apparmor_parser"); err != nil {
		result.Status = StatusError
		result.Message = "AppArmor: profile not loaded and apparmor_parser is missing"
		result.Details = "Install the apparmor userspace tools (apt install apparmor) so the " +
			"moltbunker-container profile can be loaded; otherwise containers run without AppArmor confinement."
		result.Fixable = false
		result.FixPackage = "apparmor"
		return result
	}

	// (4) Parser present, profile not loaded → fixable warning.
	result.Status = StatusWarning
	result.Message = "AppArmor: moltbunker-container profile not loaded"
	result.Details = "The runtime's AppArmor confinement gate will no-op until the profile is loaded. " +
		"Run `moltbunker node doctor --fix` to load the embedded profile."
	result.Fixable = true
	result.FixCommand = "moltbunker node doctor --fix"
	return result
}

func (c *AppArmorChecker) Fix(ctx context.Context, _ PackageManager) error {
	return c.loader.EnsureProfile(ctx, runtime.AppArmorProfileName, "")
}
