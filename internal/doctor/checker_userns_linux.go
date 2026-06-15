//go:build linux

package doctor

import (
	"context"

	"github.com/moltbunker/moltbunker/internal/runtime"
)

// UserNSChecker reports whether unprivileged user namespaces are usable for tenant
// deployment workloads (HARDEN-01, R12). A warning here means containers will run
// in the host user namespace (container root == host root) instead of being mapped
// to an unprivileged subordinate UID range.
type UserNSChecker struct {
	// compatFn is a seam for tests; nil means use runtime.CheckUserNSCompat.
	compatFn func() runtime.UserNSCompatResult
}

func NewUserNSChecker() *UserNSChecker { return &UserNSChecker{} }

func (c *UserNSChecker) Name() string       { return "User namespace support" }
func (c *UserNSChecker) Category() Category { return CategorySecurity }
func (c *UserNSChecker) CanFix() bool       { return false }
func (c *UserNSChecker) Roles() []string    { return []string{"provider", "hybrid"} }

func (c *UserNSChecker) Check(_ context.Context) CheckResult {
	result := CheckResult{Name: c.Name(), Category: c.Category()}

	compat := runtime.CheckUserNSCompat
	if c.compatFn != nil {
		compat = c.compatFn
	}
	res := compat()

	if res.Supported {
		result.Status = StatusOK
		result.Message = "User namespaces enabled: tenant containers map to an unprivileged host UID range"
		return result
	}

	result.Status = StatusWarning
	result.Message = "User namespaces disabled: containers will run as host root"
	result.Details = res.Reason + "\n" +
		"Enable with: sysctl -w kernel.unprivileged_userns_clone=1 (Debian/Ubuntu), " +
		"and ensure the daemon user has a subordinate ID range (e.g. add a line to /etc/subuid and /etc/subgid). " +
		"Disk quota note (R16): with userns active, XFS project quotas track the host subordinate UID, not container UID 0."
	return result
}

func (c *UserNSChecker) Fix(_ context.Context, _ PackageManager) error { return nil }
