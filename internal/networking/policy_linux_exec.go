//go:build linux

// R13/R14 — Linux nftables exec helpers.
//
// This file contains the pieces that turn a NetworkPolicy into a concrete
// `nft -f -` script and the default function that actually runs `nft`. The exec
// function is injected into NftPolicyEnforcer at construction time (see
// NewNftPolicyEnforcerWithExec) so the enforcer can be unit-tested on any
// platform without a real `nft` binary on PATH.
package networking

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
)

// NftExecFn pipes an `nft -f -` script to the nft binary. Implementations must
// return a non-nil error (including any stderr output) when nft exits non-zero.
type NftExecFn func(ctx context.Context, script string) error

// DefaultNftExec resolves `nft` via exec.LookPath at call time and pipes the
// supplied script to its stdin (`nft -f -`). It returns the combined stderr on
// a non-zero exit so the daemon can log exactly which rule set failed.
//
// nft is resolved lazily (per call) rather than cached so that a provider that
// installs nftables after the daemon starts does not need a restart, and so the
// enforcer can be constructed on hosts that do not yet have nft installed
// (Apply then surfaces a clear error that the doctor NftChecker also flags).
func DefaultNftExec(ctx context.Context, script string) error {
	nftPath, err := exec.LookPath("nft")
	if err != nil {
		return fmt.Errorf("nft binary not found on PATH: %w", err)
	}

	// #nosec G204 -- exec.CommandContext (no shell); nftPath is resolved via
	// exec.LookPath and the only arguments are the constants "-f" and "-". The
	// rule set is supplied on stdin, never as a shell-interpolated argument, and
	// is built solely from internal container IPs and operator-validated CIDRs.
	cmd := exec.CommandContext(ctx, nftPath, "-f", "-")
	cmd.Stdin = bytes.NewReader([]byte(script))
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if stderr.Len() > 0 {
			return fmt.Errorf("nft -f - failed: %w: %s", err, stderr.String())
		}
		return fmt.Errorf("nft -f - failed: %w", err)
	}
	return nil
}

// buildInitScript returns an idempotent table/chain creation block. `add table`
// and `add chain` silently succeed if the object already exists, so this is safe
// to run on every Apply and safe to run concurrently for different deployments
// (the shared table + forward chain creation is idempotent; per-deployment
// chains are namespaced by deploymentID).
//
// The forward chain is hooked into netfilter's forward path; the per-deployment
// chains are populated by Apply and jumped into from the forward chain by the
// jump rules added alongside the policy rules.
func buildInitScript(deploymentID string) string {
	var b bytes.Buffer
	fmt.Fprintf(&b, "add table inet %s\n", policyTable)
	fmt.Fprintf(&b, "add chain inet %s forward { type filter hook forward priority filter; policy accept; }\n", policyTable)
	fmt.Fprintf(&b, "add chain inet %s %s\n", policyTable, inChain(deploymentID))
	fmt.Fprintf(&b, "add chain inet %s %s\n", policyTable, outChain(deploymentID))
	return b.String()
}

// ComputeIngressRules translates the lateral (intra-host) isolation model into
// nftables rule lines for one container's ingress chain. The container may be
// reached from each allowed peer IP; every other source inside the intra-host
// range is dropped. Traffic to/from outside the intra-host range is not the
// ingress chain's concern (that is egress, handled by ComputeEgressRules).
//
// Rule ordering matters: accept rules for allowed peers come first, then a
// terminal drop for the rest of the intra-host range. Returns nil for empty
// deploymentID or containerIP.
func ComputeIngressRules(deploymentID, containerIP string, allowedIPs []string) []string {
	if deploymentID == "" || containerIP == "" {
		return nil
	}
	chain := inChain(deploymentID)
	var rules []string

	// Allow each reciprocally-approved peer to reach this container.
	for _, peer := range allowedIPs {
		if peer == "" {
			continue
		}
		rules = append(rules, fmt.Sprintf(
			"add rule inet %s %s ip saddr %s ip daddr %s accept",
			policyTable, chain, peer, containerIP,
		))
	}

	// Drop all other intra-host traffic destined for this container. This is the
	// default-deny lateral isolation rule: a compromised neighbour cannot reach
	// this container unless its tenant reciprocally opted in above.
	rules = append(rules, fmt.Sprintf(
		"add rule inet %s %s ip saddr %s ip daddr %s drop",
		policyTable, chain, intraHostCIDR, containerIP,
	))

	return rules
}

// jumpRules wire the shared forward chain into the per-deployment chains so the
// kernel actually evaluates the policy. Ingress is selected by destination IP
// (traffic TO the container), egress by source IP (traffic FROM the container).
//
// The forward chain is rebuilt-from-state on every Apply/Remove (it is flushed,
// then these jumps are re-emitted for every currently-applied deployment), so
// these are always `add` (append to a freshly-flushed chain) and never need
// handle-based deletion: a Remove simply omits the removed deployment from the
// rebuild. This keeps re-Apply from duplicating jumps and lets Remove delete the
// per-deployment chains without hitting EBUSY (no forward jump still targets
// them once the chain has been flushed).
func jumpRules(deploymentID, containerIP string) []string {
	return []string{
		fmt.Sprintf("add rule inet %s forward ip daddr %s jump %s",
			policyTable, containerIP, inChain(deploymentID)),
		fmt.Sprintf("add rule inet %s forward ip saddr %s jump %s",
			policyTable, containerIP, outChain(deploymentID)),
	}
}
