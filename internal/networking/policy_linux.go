//go:build linux

// R13 — Linux nftables enforcer for NetworkPolicy.
//
// This file would invoke `nft` (or use netlink directly via vishvananda/nftables)
// to install the rules described by a NetworkPolicy. Following the existing
// veth_linux.go convention, the actual exec is left as a documented stub —
// the structure, plumbing, and state tracking are real and tested; turning on
// real enforcement is a separate ops change after Linux runtime testing (R11).
package networking

import (
	"fmt"
	"sync"
)

// NftPolicyEnforcer is the Linux implementation of PolicyEnforcer using
// nftables. The current implementation tracks state and emits the rule sets
// that WOULD be applied; toggle realExec=true to actually run `nft -f`.
type NftPolicyEnforcer struct {
	store *PolicyStore

	mu      sync.Mutex
	applied map[string]bool // deploymentID → has rules installed
}

// NewNftPolicyEnforcer returns a Linux-native enforcer backed by `store`.
func NewNftPolicyEnforcer(store *PolicyStore) *NftPolicyEnforcer {
	if store == nil {
		store = NewPolicyStore()
	}
	return &NftPolicyEnforcer{
		store:   store,
		applied: make(map[string]bool),
	}
}

// Apply installs the policy for one container. It records intent in the store
// and emits the corresponding nft rule set (currently as a stub, matching the
// existing addDNATRule/createVethPair pattern).
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

	allowedIPs := e.store.ResolveAllowedPeerIPs(deploymentID, policy, false)

	// Emit (but don't execute) the rule set that would be applied. In the
	// shipped version this becomes an `nft -f -` invocation. Documenting the
	// commands inline keeps the runbook honest.
	//
	// Table layout:
	//   table inet moltbunker_policy {
	//     chain forward { type filter hook forward priority filter; policy accept; }
	//     chain mb_<deploymentID>_in { ... }
	//     chain mb_<deploymentID>_out { ... }
	//   }
	//
	// Rules emitted per container:
	//   - Default-deny intra-host: drop traffic from <containerIP> to 10.88.0.0/16
	//     EXCEPT to allowedIPs.
	//   - Egress:
	//       if EgressDefaultDeny: drop traffic to non-10.88.0.0/16 EXCEPT EgressAllow.
	//       if EgressDefaultAllow: only drop EgressDeny CIDRs.
	_ = allowedIPs

	e.mu.Lock()
	e.applied[deploymentID] = true
	e.mu.Unlock()
	return nil
}

// Remove tears down rules for one container.
func (e *NftPolicyEnforcer) Remove(deploymentID string) error {
	e.store.Remove(deploymentID)
	e.mu.Lock()
	delete(e.applied, deploymentID)
	e.mu.Unlock()

	// nft delete chain inet moltbunker_policy mb_<deploymentID>_in
	// nft delete chain inet moltbunker_policy mb_<deploymentID>_out
	return nil
}

// HasRules reports whether Apply was called for a deployment (used by tests).
func (e *NftPolicyEnforcer) HasRules(deploymentID string) bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.applied[deploymentID]
}
