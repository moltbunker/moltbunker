//go:build linux

// R13/R14 — Linux nftables enforcer for NetworkPolicy.
//
// This enforcer translates a NetworkPolicy into an `nft -f -` script and pipes
// it to the nft binary. The external command is injected (execFn) so the
// enforcement logic is unit-testable on any platform without a real nft binary;
// production uses DefaultNftExec, which resolves `nft` via exec.LookPath.
//
// Table layout (single shared inet table, per-deployment chains):
//
//	table inet moltbunker_policy {
//	  chain forward { type filter hook forward priority filter; policy accept; }
//	  chain mb_<deploymentID>_in  { ... ingress (lateral isolation) ... }
//	  chain mb_<deploymentID>_out { ... egress (R14) ... }
//	}
//
// Apply is idempotent: the table/forward chain are created with `add` (a no-op
// if they already exist), the per-deployment chains are flushed before rules are
// re-added, and the shared forward chain is rebuilt-from-state (flushed, then
// re-populated with jumps for every currently-applied deployment) on every
// Apply and Remove. The rebuild-from-state design is load-bearing for two
// reasons: (1) a re-Apply for the same deployment does not duplicate its forward
// jumps, and (2) Remove can delete a per-deployment chain without nftables
// returning EBUSY, because the rebuild has already removed the only forward jump
// that targeted it (nft does NOT auto-remove jump references to a deleted
// chain — `delete chain` on a still-referenced chain fails).
package networking

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// nftApplyTimeout bounds a single nft invocation so a wedged nft cannot block a
// deploy or stop indefinitely.
const nftApplyTimeout = 10 * time.Second

// NftPolicyEnforcer is the Linux implementation of PolicyEnforcer using
// nftables. It pipes computed rule sets to nft via the injected execFn.
type NftPolicyEnforcer struct {
	store  *PolicyStore
	execFn NftExecFn

	mu        sync.Mutex
	applied   map[string]bool     // deploymentID → has rules installed
	appliedIP map[string]string   // deploymentID → containerIP (for forward-jump rebuild)
	lastRule  map[string][]string // deploymentID → last computed nft rule set (for audit/debug)
}

// NewNftPolicyEnforcer returns a Linux-native enforcer backed by `store` that
// applies rules by running the real `nft` binary (DefaultNftExec).
func NewNftPolicyEnforcer(store *PolicyStore) *NftPolicyEnforcer {
	return NewNftPolicyEnforcerWithExec(store, DefaultNftExec)
}

// NewNftPolicyEnforcerWithExec returns an enforcer that pipes its scripts to the
// supplied execFn. Tests inject a capturing func so they can assert on the
// generated nft script without a real nft binary. A nil execFn falls back to
// DefaultNftExec.
func NewNftPolicyEnforcerWithExec(store *PolicyStore, execFn NftExecFn) *NftPolicyEnforcer {
	if store == nil {
		store = NewPolicyStore()
	}
	if execFn == nil {
		execFn = DefaultNftExec
	}
	return &NftPolicyEnforcer{
		store:     store,
		execFn:    execFn,
		applied:   make(map[string]bool),
		appliedIP: make(map[string]string),
		lastRule:  make(map[string][]string),
	}
}

// Apply installs the policy for one container. It records intent in the store,
// computes the ingress (lateral isolation) and egress (R14) rule sets, and pipes
// a single atomic nft script that (a) idempotently creates the table + chains,
// (b) flushes the per-deployment chains, (c) re-adds the ingress + egress rule
// lines, and (d) rebuilds the shared forward chain's jump wiring from the full
// applied set (so a re-Apply never duplicates the forward jumps).
func (e *NftPolicyEnforcer) Apply(deploymentID, containerIP string, policy NetworkPolicy) error {
	if deploymentID == "" {
		return fmt.Errorf("%w: empty deploymentID", ErrInvalidPolicy)
	}
	if containerIP == "" {
		return fmt.Errorf("%w: empty containerIP", ErrInvalidPolicy)
	}
	if err := policy.Validate(deploymentID); err != nil {
		return err
	}

	e.store.Set(deploymentID, containerIP, policy)

	// Only peers that reciprocally allow this deployment may reach it.
	allowedIPs := e.store.ResolveAllowedPeerIPs(deploymentID, policy, false)
	ingressRules := ComputeIngressRules(deploymentID, containerIP, allowedIPs)
	egressRules := ComputeEgressRules(deploymentID, containerIP, policy)

	// Serialize all nft mutations: concurrent Apply/Remove calls share the
	// global table + forward chain, and nft -f is not safe to interleave. The
	// forward-chain rebuild reads the applied set, so the script must be built
	// under the same lock that guards that set.
	e.mu.Lock()
	defer e.mu.Unlock()

	// Build the forward-jump set from the full applied set PLUS this deployment.
	// We add it to the map before building so the rebuild includes it, and roll
	// it back if the exec fails (so a failed Apply leaves no phantom jump).
	prevIP, hadPrev := e.appliedIP[deploymentID]
	e.appliedIP[deploymentID] = containerIP
	script := e.buildApplyScript(deploymentID, ingressRules, egressRules)

	ctx, cancel := context.WithTimeout(context.Background(), nftApplyTimeout)
	defer cancel()
	if err := e.execFn(ctx, script); err != nil {
		if hadPrev {
			e.appliedIP[deploymentID] = prevIP
		} else {
			delete(e.appliedIP, deploymentID)
		}
		return fmt.Errorf("nft apply for %s: %w", deploymentID, err)
	}

	e.applied[deploymentID] = true
	combined := make([]string, 0, len(ingressRules)+len(egressRules))
	combined = append(combined, ingressRules...)
	combined = append(combined, egressRules...)
	e.lastRule[deploymentID] = combined
	return nil
}

// buildApplyScript assembles the full idempotent `nft -f -` script for one
// deployment: init block, flush of the per-deployment chains, the ingress +
// egress rule lines, and a rebuild-from-state of the shared forward chain's
// jump wiring. Callers must hold e.mu (it reads e.appliedIP).
//
// The forward chain is flushed and re-populated with jumps for every currently
// applied deployment on every Apply. This is what makes re-Apply idempotent:
// the prior pair of forward jumps for this deployment is cleared by the flush
// and re-added exactly once, so a policy update never doubles the jumps.
func (e *NftPolicyEnforcer) buildApplyScript(deploymentID string, ingressRules, egressRules []string) string {
	var b strings.Builder
	b.WriteString(buildInitScript(deploymentID))
	// Flush the per-deployment chains so a re-Apply (policy update) replaces the
	// prior rule set rather than appending to it.
	fmt.Fprintf(&b, "flush chain inet %s %s\n", policyTable, inChain(deploymentID))
	fmt.Fprintf(&b, "flush chain inet %s %s\n", policyTable, outChain(deploymentID))
	for _, r := range ingressRules {
		b.WriteString(r)
		b.WriteByte('\n')
	}
	for _, r := range egressRules {
		b.WriteString(r)
		b.WriteByte('\n')
	}
	b.WriteString(e.buildForwardRebuild())
	return b.String()
}

// buildForwardRebuild emits a flush of the shared forward chain followed by the
// jump wiring for every currently-applied deployment (read from e.appliedIP).
// Callers must hold e.mu.
//
// Rebuilding the forward chain from state on every Apply/Remove is the
// mechanism that (a) keeps re-Apply from duplicating jumps and (b) lets Remove
// delete a per-deployment chain without hitting EBUSY: by the time the chain is
// deleted, the flush has already removed the only forward jump that targeted it
// (contrary to the earlier belief that nft auto-removes references to a deleted
// chain — it does not; `delete chain` on a chain that is still a jump target
// returns "Device or resource busy").
//
// Deployments are emitted in sorted order so the generated script is
// deterministic (stable for tests and audit diffs).
func (e *NftPolicyEnforcer) buildForwardRebuild() string {
	var b strings.Builder
	fmt.Fprintf(&b, "flush chain inet %s forward\n", policyTable)
	ids := make([]string, 0, len(e.appliedIP))
	for id := range e.appliedIP {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		for _, r := range jumpRules(id, e.appliedIP[id]) {
			b.WriteString(r)
			b.WriteByte('\n')
		}
	}
	return b.String()
}

// Remove tears down rules for one container. It first drops the deployment from
// the applied set and rebuilds the shared forward chain from the remaining
// applied deployments — this deletes the two forward jumps that targeted this
// deployment's chains. Only THEN are the per-deployment chains flushed and
// deleted, because nftables refuses to `delete chain` while any jump/goto still
// targets it (it returns "Device or resource busy" / EBUSY). The shared table +
// forward chain are left in place (other deployments may still use them).
// Safe to call even if Apply was never called.
func (e *NftPolicyEnforcer) Remove(deploymentID string) error {
	e.store.Remove(deploymentID)

	e.mu.Lock()
	defer e.mu.Unlock()

	wasApplied := e.applied[deploymentID]
	delete(e.applied, deploymentID)
	delete(e.lastRule, deploymentID)
	prevIP, hadIP := e.appliedIP[deploymentID]
	delete(e.appliedIP, deploymentID)

	if !wasApplied {
		// Nothing was installed for this deployment; deleting a non-existent
		// chain would make nft error, so skip the exec entirely.
		return nil
	}

	var b strings.Builder
	// Rebuild the forward chain WITHOUT this deployment's jumps first, so the
	// subsequent delete chain does not hit EBUSY (no forward jump references the
	// chains being deleted anymore).
	b.WriteString(e.buildForwardRebuild())
	fmt.Fprintf(&b, "flush chain inet %s %s\n", policyTable, inChain(deploymentID))
	fmt.Fprintf(&b, "flush chain inet %s %s\n", policyTable, outChain(deploymentID))
	fmt.Fprintf(&b, "delete chain inet %s %s\n", policyTable, inChain(deploymentID))
	fmt.Fprintf(&b, "delete chain inet %s %s\n", policyTable, outChain(deploymentID))

	ctx, cancel := context.WithTimeout(context.Background(), nftApplyTimeout)
	defer cancel()
	if err := e.execFn(ctx, b.String()); err != nil {
		// Restore applied-set tracking so a later retry can rebuild correctly.
		e.applied[deploymentID] = true
		if hadIP {
			e.appliedIP[deploymentID] = prevIP
		}
		return fmt.Errorf("nft remove for %s: %w", deploymentID, err)
	}
	return nil
}

// LastRules returns the nft rule lines computed for a deployment at its most
// recent Apply call. Returns nil if Apply was never called for this ID.
// Useful for audit logging and ops debugging.
func (e *NftPolicyEnforcer) LastRules(deploymentID string) []string {
	e.mu.Lock()
	defer e.mu.Unlock()
	if rules, ok := e.lastRule[deploymentID]; ok {
		out := make([]string, len(rules))
		copy(out, rules)
		return out
	}
	return nil
}

// HasRules reports whether Apply was called for a deployment (used by tests).
func (e *NftPolicyEnforcer) HasRules(deploymentID string) bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.applied[deploymentID]
}
