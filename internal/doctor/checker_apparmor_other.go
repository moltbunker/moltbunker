//go:build !linux

package doctor

import "context"

// AppArmorChecker is a no-op on non-Linux platforms: it always reports Skipped so
// `moltbunker doctor` output stays consistent across platforms without a build-tag
// split at the registration site.
type AppArmorChecker struct{}

func NewAppArmorChecker() *AppArmorChecker { return &AppArmorChecker{} }

func (c *AppArmorChecker) Name() string       { return "AppArmor profile" }
func (c *AppArmorChecker) Category() Category { return CategorySecurity }
func (c *AppArmorChecker) CanFix() bool       { return false }
func (c *AppArmorChecker) Roles() []string    { return []string{"provider", "hybrid"} }

func (c *AppArmorChecker) Check(_ context.Context) CheckResult {
	return CheckResult{
		Name:     c.Name(),
		Category: c.Category(),
		Status:   StatusSkipped,
		Message:  "AppArmor: not applicable on this platform",
	}
}

func (c *AppArmorChecker) Fix(_ context.Context, _ PackageManager) error { return nil }
