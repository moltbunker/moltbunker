//go:build !linux

package doctor

import "context"

// UserNSChecker is a no-op on non-Linux platforms (user namespaces are a Linux
// kernel feature); it always reports Skipped.
type UserNSChecker struct{}

func NewUserNSChecker() *UserNSChecker { return &UserNSChecker{} }

func (c *UserNSChecker) Name() string       { return "User namespace support" }
func (c *UserNSChecker) Category() Category { return CategorySecurity }
func (c *UserNSChecker) CanFix() bool       { return false }
func (c *UserNSChecker) Roles() []string    { return []string{"provider", "hybrid"} }

func (c *UserNSChecker) Check(_ context.Context) CheckResult {
	return CheckResult{
		Name:     c.Name(),
		Category: c.Category(),
		Status:   StatusSkipped,
		Message:  "User namespaces: not applicable on this platform",
	}
}

func (c *UserNSChecker) Fix(_ context.Context, _ PackageManager) error { return nil }
