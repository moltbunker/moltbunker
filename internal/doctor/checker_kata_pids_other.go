//go:build !linux

package doctor

import (
	"context"

	"github.com/moltbunker/moltbunker/internal/runtime"
)

// KataPIDsChecker is a no-op on non-Linux platforms (Kata VM isolation is Linux
// only); it always reports Skipped. The constructor keeps the same signature as
// the Linux build so the registration site compiles without a build-tag split.
type KataPIDsChecker struct {
	cfg         *runtime.KataConfig
	ociPIDLimit int
}

func NewKataPIDsChecker(cfg *runtime.KataConfig) *KataPIDsChecker {
	return &KataPIDsChecker{cfg: cfg}
}

// NewKataPIDsCheckerWithOCILimit mirrors the Linux constructor so the daemon's
// registration site compiles without a build-tag split. The field is unused on
// non-Linux platforms (Kata VM isolation is Linux only).
func NewKataPIDsCheckerWithOCILimit(cfg *runtime.KataConfig, ociPIDLimit int) *KataPIDsChecker {
	return &KataPIDsChecker{cfg: cfg, ociPIDLimit: ociPIDLimit}
}

func (c *KataPIDsChecker) Name() string       { return "Kata PID limit (R17)" }
func (c *KataPIDsChecker) Category() Category { return CategorySecurity }
func (c *KataPIDsChecker) CanFix() bool       { return false }
func (c *KataPIDsChecker) Roles() []string    { return []string{"provider", "hybrid"} }

func (c *KataPIDsChecker) Check(_ context.Context) CheckResult {
	return CheckResult{
		Name:     c.Name(),
		Category: c.Category(),
		Status:   StatusSkipped,
		Message:  "Kata PID limit: not applicable on this platform",
	}
}

func (c *KataPIDsChecker) Fix(_ context.Context, _ PackageManager) error { return nil }
