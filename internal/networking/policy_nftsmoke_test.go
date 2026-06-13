//go:build linux && nftsmoke

// Real-nft smoke test for the nftables enforcer (R13/R14).
//
// Unit tests inject a fake exec (scriptCapture) and never model real nft
// semantics, so a generated script that nft rejects — an invalid chain
// identifier (e.g. an unquoted '-'), or a `delete chain` that hits EBUSY
// because a forward jump still targets it — passes the unit tests yet fails
// the moment R13/R14 runs on a provider. This test pipes the REAL generated
// scripts to the REAL `nft` binary so those failures are caught in CI, not in
// production where applyNetworkPolicy only logs a Warn.
//
// Gated behind the `nftsmoke` build tag because it requires a real nft binary
// and CAP_NET_ADMIN (root). The OPS-04 runtime-e2e CI job runs it as root after
// `apt-get install nftables`. It is skipped from the default `go test ./...`.
package networking

import (
	"context"
	"os/exec"
	"testing"
)

func requireNft(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("nft"); err != nil {
		t.Skip("nft binary not on PATH; install nftables to run the smoke test")
	}
}

// TestSmoke_ApplyRemoveAgainstRealNft exercises the full Apply -> re-Apply ->
// Remove lifecycle against the real nft binary. It is the regression guard for:
//   - the HIGH EBUSY teardown bug (delete chain while a forward jump targets it),
//   - the MEDIUM duplicate-jump-on-reapply,
//   - the LOW hyphenated chain-name rejection (deploymentID "dep-<hex>").
func TestSmoke_ApplyRemoveAgainstRealNft(t *testing.T) {
	requireNft(t)

	e := NewNftPolicyEnforcer(NewPolicyStore())
	// Best-effort cleanup of the shared table regardless of outcome.
	t.Cleanup(func() {
		_ = DefaultNftExec(context.Background(), "delete table inet "+policyTable+"\n")
	})

	// A hyphenated deploymentID is the production shape (generateDeploymentID).
	// If chain-name sanitization were wrong, the real nft would reject this.
	const depA = "dep-aaaa1111"
	const depB = "dep-bbbb2222"

	if err := e.Apply(depA, "10.88.0.5", NetworkPolicy{EgressMode: EgressDefaultDeny}); err != nil {
		t.Fatalf("Apply %s against real nft: %v", depA, err)
	}
	// Re-Apply (policy update) must not duplicate forward jumps or error.
	if err := e.Apply(depA, "10.88.0.5", NetworkPolicy{EgressMode: EgressDefaultAllow}); err != nil {
		t.Fatalf("re-Apply %s against real nft: %v", depA, err)
	}
	if err := e.Apply(depB, "10.88.0.6", NetworkPolicy{EgressMode: EgressDefaultDeny}); err != nil {
		t.Fatalf("Apply %s against real nft: %v", depB, err)
	}

	// The critical regression: Remove must NOT hit EBUSY. With the old code
	// (delete chain while the forward jump still targets it) this errored.
	if err := e.Remove(depA); err != nil {
		t.Fatalf("Remove %s against real nft (EBUSY regression?): %v", depA, err)
	}
	if err := e.Remove(depB); err != nil {
		t.Fatalf("Remove %s against real nft: %v", depB, err)
	}
}
