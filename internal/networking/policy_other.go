//go:build !linux

package networking

// R13/R14 — non-Linux stub. nftables is Linux-only; on macOS/Colima we keep a
// no-op enforcer that satisfies the interface so the rest of the daemon can
// be developed cross-platform.

import (
	"context"
	"fmt"
	"sync"
)

// NftExecFn mirrors the Linux signature so callers (and tests) can reference it
// cross-platform. On non-Linux it is never invoked.
type NftExecFn func(ctx context.Context, script string) error

// NftPolicyEnforcer is the cross-platform stub for the Linux nftables
// enforcer. On non-Linux platforms it just records intent.
type NftPolicyEnforcer struct {
	store *PolicyStore

	mu       sync.Mutex
	applied  map[string]bool
	lastRule map[string][]string
}

// NewNftPolicyEnforcer returns a no-op enforcer outside of Linux.
func NewNftPolicyEnforcer(store *PolicyStore) *NftPolicyEnforcer {
	return NewNftPolicyEnforcerWithExec(store, nil)
}

// NewNftPolicyEnforcerWithExec returns a no-op enforcer outside of Linux. The
// execFn is accepted for API parity with the Linux build but is never called.
func NewNftPolicyEnforcerWithExec(store *PolicyStore, _ NftExecFn) *NftPolicyEnforcer {
	if store == nil {
		store = NewPolicyStore()
	}
	return &NftPolicyEnforcer{
		store:    store,
		applied:  make(map[string]bool),
		lastRule: make(map[string][]string),
	}
}

// Apply records the policy in the store; no OS-level enforcement happens.
//
// The empty-deploymentID/empty-containerIP guards mirror the Linux Apply
// (policy_linux.go) so the stub rejects the same malformed inputs on dev
// machines — without this parity a caller bug would surface only in production
// on Linux.
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
	e.mu.Lock()
	e.applied[deploymentID] = true
	e.lastRule[deploymentID] = ComputeEgressRules(deploymentID, containerIP, policy)
	e.mu.Unlock()
	return nil
}

// Remove drops the policy from the store.
func (e *NftPolicyEnforcer) Remove(deploymentID string) error {
	e.store.Remove(deploymentID)
	e.mu.Lock()
	delete(e.applied, deploymentID)
	delete(e.lastRule, deploymentID)
	e.mu.Unlock()
	return nil
}

// HasRules reports whether Apply was called.
func (e *NftPolicyEnforcer) HasRules(deploymentID string) bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.applied[deploymentID]
}

// LastRules returns the nft rule lines computed at the most recent Apply.
// Returns nil if Apply was never called for this deployment.
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
