//go:build e2e

package golden

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net"
	"testing"

	"github.com/moltbunker/moltbunker/internal/networking"
	"github.com/moltbunker/moltbunker/internal/runtime"
	"github.com/moltbunker/moltbunker/tests/e2e/testutil"
)

// fixedDigest is a stable sha256 digest used across the gate sub-tests.
const fixedDigest = runtime.ImageDigest("sha256:" +
	"abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789")

// All sub-tests below are pure-Go and run on every platform with only the
// `e2e` build tag — no trivy binary, no nft, no chain, no network. They isolate
// the R3 (signature) and R4 (scan) gates and the R13/R14 (egress) rule logic
// so a regression in any single gate fails a focused, named test.

// -----------------------------------------------------------------------------
// R3 — image signature gate. [REAL: runtime.EdImageVerifier]
// -----------------------------------------------------------------------------

func TestGoldenPath_SigGatePass(t *testing.T) {
	a := testutil.NewAssertions(t)
	t.Log("[REAL: Ed25519 sig verify] valid sig over a fixed digest verifies cleanly")

	_, priv, err := ed25519.GenerateKey(rand.Reader)
	a.NoError(err)

	sig := runtime.SignImageDigest(fixedDigest, priv)
	policy := runtime.TrustPolicy{
		RequireSignature:  true,
		TrustedPublishers: []string{sig.PublisherID},
	}
	a.NoError(runtime.NewEdImageVerifier().Verify(fixedDigest, sig, policy),
		"valid signature from trusted publisher should verify")
}

func TestGoldenPath_SigGateFail_Unsigned(t *testing.T) {
	a := testutil.NewAssertions(t)
	t.Log("[REAL: Ed25519 sig verify] RequireSignature + nil sig -> ErrSignatureRequired")

	policy := runtime.TrustPolicy{RequireSignature: true}
	err := runtime.NewEdImageVerifier().Verify(fixedDigest, nil, policy)
	a.True(errors.Is(err, runtime.ErrSignatureRequired),
		"unsigned image under RequireSignature must be ErrSignatureRequired")
}

func TestGoldenPath_SigGateFail_UntrustedPublisher(t *testing.T) {
	a := testutil.NewAssertions(t)
	t.Log("[REAL: Ed25519 sig verify] sig from a key not in trust list -> ErrUntrustedPublisher")

	_, priv, err := ed25519.GenerateKey(rand.Reader)
	a.NoError(err)
	sig := runtime.SignImageDigest(fixedDigest, priv)

	// A different (trusted) publisher key — valid 32-byte hex, but not the signer.
	otherPub, _, err := ed25519.GenerateKey(rand.Reader)
	a.NoError(err)
	policy := runtime.TrustPolicy{
		RequireSignature:  true,
		TrustedPublishers: []string{hex.EncodeToString(otherPub)},
	}
	err = runtime.NewEdImageVerifier().Verify(fixedDigest, sig, policy)
	a.True(errors.Is(err, runtime.ErrUntrustedPublisher),
		"sig from an untrusted publisher must be ErrUntrustedPublisher")
}

// -----------------------------------------------------------------------------
// R4 — CVE scan gate. [MOCK: MockImageScanner -> REAL: ScanPolicy.Apply]
// -----------------------------------------------------------------------------

func TestGoldenPath_ScanGatePass(t *testing.T) {
	a := testutil.NewAssertions(t)
	t.Log("[MOCK: scanner / REAL: policy] zero-finding report passes DefaultScanPolicy")

	report := []runtime.Vulnerability{}
	_, err := runtime.DefaultScanPolicy().Apply(report)
	a.NoError(err, "zero findings should pass")
}

func TestGoldenPath_ScanGateFail_Critical(t *testing.T) {
	a := testutil.NewAssertions(t)
	t.Log("[MOCK: scanner / REAL: policy] single CRITICAL finding -> ErrPolicyViolation")

	findings := []runtime.Vulnerability{{
		ID:       "CVE-2026-9999",
		Severity: runtime.SeverityCritical,
		Package:  "libfoo",
		Version:  "1.0.0",
	}}
	_, err := runtime.DefaultScanPolicy().Apply(findings)
	a.True(errors.Is(err, runtime.ErrPolicyViolation),
		"a CRITICAL finding must block with ErrPolicyViolation")
}

func TestGoldenPath_ScanGateFail_IgnoreCVE(t *testing.T) {
	a := testutil.NewAssertions(t)
	t.Log("[MOCK: scanner / REAL: policy] same CRITICAL finding in IgnoreCVEs allowlist passes")

	findings := []runtime.Vulnerability{{
		ID:       "CVE-2026-9999",
		Severity: runtime.SeverityCritical,
		Package:  "libfoo",
		Version:  "1.0.0",
	}}
	policy := runtime.ScanPolicy{
		BlockAtOrAbove: runtime.SeverityHigh,
		IgnoreCVEs:     []string{"CVE-2026-9999"},
	}
	_, err := policy.Apply(findings)
	a.NoError(err, "a CRITICAL finding on the ignore list must not block")
}

// -----------------------------------------------------------------------------
// R13/R14 — egress policy. [REAL: networking.EvaluateEgress + ComputeEgressRules]
// -----------------------------------------------------------------------------

func TestGoldenPath_NetPolicyEgress(t *testing.T) {
	a := testutil.NewAssertions(t)
	t.Log("[REAL: egress eval + rule gen] default-deny blocks IMDS, allows DNS resolver")

	policy := networking.DefaultRestrictiveEgressPolicy()
	a.NoError(policy.Validate("dep-egress-test"), "restrictive policy must be valid")

	// Pure-function evaluation: deny beats allow, explicit beats default.
	a.Equal(networking.EgressBlocked, policy.EvaluateEgress(net.ParseIP("169.254.169.254")),
		"IMDS must be blocked")
	a.Equal(networking.EgressAllowed, policy.EvaluateEgress(net.ParseIP("8.8.8.8")),
		"Google DNS resolver must be allowed")
	a.Equal(networking.EgressBlocked, policy.EvaluateEgress(net.ParseIP("10.0.0.5")),
		"RFC1918 must be blocked")
	a.Equal(networking.EgressBlocked, policy.EvaluateEgress(nil),
		"nil ip must fail closed")

	// Rule-generation: emitted nft lines must drop IMDS and accept the resolver,
	// with the drop preceding the accept (deny is highest precedence).
	rules := networking.ComputeEgressRules("dep-egress-test", "10.88.0.7", policy)
	a.NotEmpty(rules, "rule set should be non-empty")

	imdsIdx := indexOfRule(rules, "169.254.169.254/32", "drop")
	dnsIdx := indexOfRule(rules, "8.8.8.8/32", "accept")
	a.True(imdsIdx >= 0, "IMDS drop rule must be present")
	a.True(dnsIdx >= 0, "DNS accept rule must be present")
	a.True(imdsIdx < dnsIdx, "deny rules must be emitted before allow rules")
}

// indexOfRule returns the index of the first rule line containing both cidr and
// verb, or -1 if none.
func indexOfRule(rules []string, cidr, verb string) int {
	for i, r := range rules {
		if contains(r, cidr) && contains(r, verb) {
			return i
		}
	}
	return -1
}
