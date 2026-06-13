package doctor

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// checker_security_tools.go — RUN-01 doctor checks for the host tooling that the
// expose-to-internet security gates depend on. Each gate (R3 image-signature
// verify, R4 CVE scan, R13/R14 nftables enforcement) is built to FAIL OPEN when
// its tool is missing: rather than break a deploy, the gate silently no-ops.
// That is the right runtime default, but it means an operator can enable a gate
// in config and get zero protection with no signal. These checkers turn that
// silent no-op into a visible `moltbunker doctor` Warning.
//
// Severity is Warning (not Error): a missing tool only disables an opt-in gate;
// it never breaks a working node. This matches buildImageScanner's fail-open
// philosophy.

// providerRoles is the role set for tooling that only matters on nodes that run
// containers. Pure requesters never enforce these gates.
func providerRoles() []string { return []string{"provider", "hybrid"} }

// runVersion runs `<bin> <args...>` and returns the trimmed first line of
// combined output, or "" if the command fails. Used to confirm a tool actually
// responds (not just that the binary exists on PATH).
func runVersion(ctx context.Context, bin string, args ...string) string {
	// #nosec G204 -- exec.CommandContext (no shell); bin is a path already
	// resolved via exec.LookPath by the caller and args are package constants.
	out, err := exec.CommandContext(ctx, bin, args...).CombinedOutput()
	if err != nil {
		return ""
	}
	line := strings.SplitN(strings.TrimSpace(string(out)), "\n", 2)
	if len(line) == 0 {
		return ""
	}
	return strings.TrimSpace(line[0])
}

// ---------------------------------------------------------------------------
// TrivyChecker — R4 image vulnerability scanning.
// ---------------------------------------------------------------------------

// TrivyChecker verifies the trivy image scanner is installed. The R4 CVE gate
// (config.Security image scanning) silently no-ops to a NoopScanner when trivy
// is absent, so this surfaces the missing dependency.
type TrivyChecker struct{}

func NewTrivyChecker() *TrivyChecker { return &TrivyChecker{} }

func (c *TrivyChecker) Name() string       { return "trivy (image scanner)" }
func (c *TrivyChecker) Category() Category { return CategoryRuntime }
func (c *TrivyChecker) CanFix() bool       { return true }
func (c *TrivyChecker) Roles() []string    { return providerRoles() }

func (c *TrivyChecker) Check(ctx context.Context) CheckResult {
	result := CheckResult{Name: c.Name(), Category: c.Category()}

	path, err := exec.LookPath("trivy")
	if err != nil {
		result.Status = StatusWarning
		result.Fixable = true
		result.FixPackage = "trivy"
		result.Message = "trivy not on PATH; the R4 image-scan gate will silently no-op when enabled"
		result.Details = "Install: https://aquasecurity.github.io/trivy"
		return result
	}

	version := runVersion(ctx, path, "--version")
	result.Status = StatusOK
	if version != "" {
		result.Message = "trivy: " + version
	} else {
		result.Message = "trivy: installed"
	}
	return result
}

func (c *TrivyChecker) Fix(ctx context.Context, pm PackageManager) error {
	if runtime.GOOS == "linux" {
		return fmt.Errorf("install trivy: https://aquasecurity.github.io/trivy")
	}
	if pm == nil {
		return fmt.Errorf("no package manager available to install trivy")
	}
	return pm.Install(ctx, "trivy")
}

// ---------------------------------------------------------------------------
// NftChecker — R13/R14 nftables network-policy enforcement.
// ---------------------------------------------------------------------------

// NftChecker verifies the `nft` binary is installed and responds. R13/R14
// network-policy enforcement pipes rule sets to `nft -f -`; without it the
// enforcer errors per-deploy (non-fatal) and policy is never installed. On
// non-Linux platforms nftables does not apply, so the check is Skipped.
type NftChecker struct{}

func NewNftChecker() *NftChecker { return &NftChecker{} }

func (c *NftChecker) Name() string       { return "nft (nftables)" }
func (c *NftChecker) Category() Category { return CategoryRuntime }
func (c *NftChecker) CanFix() bool       { return false }
func (c *NftChecker) Roles() []string    { return providerRoles() }

func (c *NftChecker) Check(ctx context.Context) CheckResult {
	result := CheckResult{Name: c.Name(), Category: c.Category()}

	if runtime.GOOS != "linux" {
		result.Status = StatusSkipped
		result.Message = "nftables not applicable on this platform"
		return result
	}

	path, err := exec.LookPath("nft")
	if err != nil {
		result.Status = StatusWarning
		result.Message = "nft not on PATH; R13/R14 network-policy enforcement will fail per-deploy when a policy is set"
		result.Details = "Install nftables, e.g. apt install nftables"
		return result
	}

	version := runVersion(ctx, path, "--version")
	if version == "" {
		result.Status = StatusWarning
		result.Message = "nft found on PATH but did not respond to --version"
		result.Details = "Ensure nft is runnable (may require CAP_NET_ADMIN / root at enforcement time)"
		return result
	}
	result.Status = StatusOK
	result.Message = "nft: " + version
	return result
}

func (c *NftChecker) Fix(ctx context.Context, pm PackageManager) error {
	return fmt.Errorf("nftables installation is platform-specific; install via your distro (e.g. apt install nftables)")
}

// ---------------------------------------------------------------------------
// ImageSignatureToolingChecker — R3 image signature verification.
// ---------------------------------------------------------------------------

// ImageSignatureToolingChecker verifies cosign (sigstore) is installed. R3
// signature verification can be driven by caller-supplied signatures, but the
// surrounding tooling/signing workflow relies on cosign; absent it, operators
// who enable signature requirements get a gate that cannot source signatures.
type ImageSignatureToolingChecker struct{}

func NewImageSignatureToolingChecker() *ImageSignatureToolingChecker {
	return &ImageSignatureToolingChecker{}
}

func (c *ImageSignatureToolingChecker) Name() string       { return "cosign (image signature tooling)" }
func (c *ImageSignatureToolingChecker) Category() Category { return CategoryRuntime }
func (c *ImageSignatureToolingChecker) CanFix() bool       { return false }
func (c *ImageSignatureToolingChecker) Roles() []string    { return providerRoles() }

func (c *ImageSignatureToolingChecker) Check(ctx context.Context) CheckResult {
	result := CheckResult{Name: c.Name(), Category: c.Category()}

	path, err := exec.LookPath("cosign")
	if err != nil {
		result.Status = StatusWarning
		result.Message = "cosign not on PATH; the R3 signature-verification gate will silently no-op without it"
		result.Details = "Install cosign: https://docs.sigstore.dev/cosign/installation"
		return result
	}

	version := runVersion(ctx, path, "version")
	result.Status = StatusOK
	if version != "" {
		result.Message = "cosign: " + version
	} else {
		result.Message = "cosign: installed"
	}
	return result
}

func (c *ImageSignatureToolingChecker) Fix(ctx context.Context, pm PackageManager) error {
	return fmt.Errorf("install cosign: https://docs.sigstore.dev/cosign/installation")
}
