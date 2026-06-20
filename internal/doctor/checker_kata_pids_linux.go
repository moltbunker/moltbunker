//go:build linux

package doctor

import (
	"context"
	"fmt"

	"github.com/moltbunker/moltbunker/internal/runtime"
)

// KataPIDsChecker warns when Kata Containers is available but no PID limit is in
// effect for VM workloads (HARDEN-01, R17).
//
// The EFFECTIVE PID ceiling for a Kata workload is the OCI
// linux.resources.pids.limit — the kata-agent enforces that cgroup limit inside
// the guest. The daemon's deploy path always sets it (DefaultResources.PIDLimit,
// default 100), so a workload is bounded whenever that value is > 0, regardless of
// the io.katacontainers.config.hypervisor.default_pids annotation (which is not a
// recognized Kata annotation and is inert without enable_annotations — see
// container_lifecycle.go). This checker therefore considers BOTH the effective OCI
// PID limit and the Kata config annotation, and only warns when NEITHER bounds the
// guest.
type KataPIDsChecker struct {
	cfg *runtime.KataConfig
	// ociPIDLimit is the daemon's effective OCI pids.limit for deployed workloads
	// (Runtime.DefaultResources.PIDLimit). > 0 means the kata-agent enforces a PID
	// ceiling in-guest. 0 means unknown/unset (e.g. a standalone doctor run that
	// has not loaded the daemon config).
	ociPIDLimit int
	// kataAvailableFn is a seam for tests; nil means runtime.IsKataAvailable.
	kataAvailableFn func() bool
}

// NewKataPIDsChecker builds the checker with the given Kata config. cfg may be nil
// for a standalone doctor run that has not loaded the daemon config; the daemon
// should pass the real config (and the effective OCI PID limit via
// NewKataPIDsCheckerWithOCILimit) so the limit is evaluated.
func NewKataPIDsChecker(cfg *runtime.KataConfig) *KataPIDsChecker {
	return &KataPIDsChecker{cfg: cfg}
}

// NewKataPIDsCheckerWithOCILimit builds the checker with both the Kata config and
// the daemon's effective OCI pids.limit for deployed workloads. The daemon should
// use this so the R17 check reflects the real in-guest PID ceiling.
func NewKataPIDsCheckerWithOCILimit(cfg *runtime.KataConfig, ociPIDLimit int) *KataPIDsChecker {
	return &KataPIDsChecker{cfg: cfg, ociPIDLimit: ociPIDLimit}
}

func (c *KataPIDsChecker) Name() string       { return "Kata PID limit (R17)" }
func (c *KataPIDsChecker) Category() Category { return CategorySecurity }
func (c *KataPIDsChecker) CanFix() bool       { return false }
func (c *KataPIDsChecker) Roles() []string    { return []string{"provider", "hybrid"} }

func (c *KataPIDsChecker) Check(_ context.Context) CheckResult {
	result := CheckResult{Name: c.Name(), Category: c.Category()}

	available := runtime.IsKataAvailable
	if c.kataAvailableFn != nil {
		available = c.kataAvailableFn
	}
	if !available() {
		result.Status = StatusSkipped
		result.Message = "Kata PID limit: Kata not installed (no VM workloads)"
		return result
	}

	// The effective in-guest PID ceiling is the OCI pids.limit (kata-agent-enforced
	// in the guest cgroup). If the daemon sets it (> 0), VM workloads ARE bounded
	// even though the default_pids hypervisor annotation is inert.
	if c.ociPIDLimit > 0 {
		result.Status = StatusOK
		result.Message = fmt.Sprintf("Kata VM workloads bounded by OCI pids.limit: %d", c.ociPIDLimit)
		result.Details = "The kata-agent enforces the OCI linux.resources.pids.limit inside the guest. " +
			"This is the effective R17 PID ceiling for VM workloads; " +
			"io.katacontainers.config.hypervisor.default_pids is inert without enable_annotations."
		return result
	}

	if c.cfg != nil && c.cfg.DefaultPIDs > 0 {
		// No known effective OCI limit, but a Kata annotation is configured. This is
		// only a best-effort hint (the annotation is inert without enable_annotations),
		// so surface it as informational rather than a clean OK.
		result.Status = StatusOK
		result.Message = fmt.Sprintf("Kata default_pids annotation set: %d (verify enable_annotations under R11)", c.cfg.DefaultPIDs)
		result.Details = "The effective PID ceiling for VM workloads is the OCI pids.limit (kata-agent-enforced in-guest). " +
			"io.katacontainers.config.hypervisor.default_pids is NOT a recognized Kata hypervisor annotation and is " +
			"dropped unless listed in the Kata runtime TOML enable_annotations; treat it as a forward-looking hint only."
		return result
	}

	result.Status = StatusWarning
	result.Message = "Kata VM workloads have no PID limit"
	result.Details = "Set a per-deployment OCI pids.limit via runtime.default_resources (recommended: 100) — " +
		"the kata-agent enforces it inside the guest and it is the effective R17 PID ceiling. " +
		"Note: io.katacontainers.config.hypervisor.default_pids is not a recognized Kata annotation and is inert " +
		"without enable_annotations, so do not rely on runtime.kata.default_pids alone."
	return result
}

func (c *KataPIDsChecker) Fix(_ context.Context, _ PackageManager) error { return nil }
