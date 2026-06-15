//go:build linux

package networking

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
)

// scriptCapture is a test NftExecFn that records every script piped to it.
type scriptCapture struct {
	mu      sync.Mutex
	scripts []string
	err     error // if non-nil, returned from every call
}

func (c *scriptCapture) exec(_ context.Context, script string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.scripts = append(c.scripts, script)
	return c.err
}

func (c *scriptCapture) all() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return strings.Join(c.scripts, "\n---\n")
}

func (c *scriptCapture) count() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.scripts)
}

func TestNftPolicyEnforcer_Apply_EmitsInitAndRules(t *testing.T) {
	cap := &scriptCapture{}
	e := NewNftPolicyEnforcerWithExec(NewPolicyStore(), cap.exec)

	policy := NetworkPolicy{EgressMode: EgressDefaultDeny}
	if err := e.Apply("dep-1", "10.88.0.5", policy); err != nil {
		t.Fatalf("Apply: %v", err)
	}

	if cap.count() == 0 {
		t.Fatal("Apply emitted no nft script")
	}
	got := cap.all()
	// Chain names are sanitized ('-' -> '_') so they are valid unquoted nft
	// identifiers: deploymentID "dep-1" -> chain suffix "dep_1".
	for _, want := range []string{policyTable, "10.88.0.5", "drop", "add table", "mb_dep_1_out"} {
		if !strings.Contains(got, want) {
			t.Errorf("script missing %q\nscript:\n%s", want, got)
		}
	}
	// The unsanitized hyphenated chain name must NOT appear (an unquoted '-' can
	// be rejected by nft and silently no-op enforcement).
	if strings.Contains(got, "mb_dep-1_") {
		t.Errorf("script must not contain unsanitized hyphenated chain name; script:\n%s", got)
	}
	if !e.HasRules("dep-1") {
		t.Error("HasRules(dep-1) should be true after Apply")
	}
}

func TestNftPolicyEnforcer_Apply_IdempotentSecondApply(t *testing.T) {
	cap := &scriptCapture{}
	e := NewNftPolicyEnforcerWithExec(NewPolicyStore(), cap.exec)

	if err := e.Apply("dep-2", "10.88.0.9", NetworkPolicy{EgressMode: EgressDefaultAllow}); err != nil {
		t.Fatalf("first Apply: %v", err)
	}
	// Re-apply with a changed policy: must flush + re-add (idempotent update).
	if err := e.Apply("dep-2", "10.88.0.9", NetworkPolicy{EgressMode: EgressDefaultDeny}); err != nil {
		t.Fatalf("second Apply: %v", err)
	}

	if cap.count() != 2 {
		t.Fatalf("expected 2 script invocations, got %d", cap.count())
	}
	got := cap.all()
	if !strings.Contains(got, "flush chain") {
		t.Errorf("re-Apply must flush the chain before re-adding; script:\n%s", got)
	}

	// The forward chain is rebuilt-from-state: it must be flushed before the
	// jumps are re-added, otherwise a re-Apply appends a duplicate pair of
	// forward jumps for the same deployment.
	second := lastScript(t, cap)
	if !strings.Contains(second, "flush chain inet "+policyTable+" forward") {
		t.Errorf("re-Apply must flush the forward chain before re-adding jumps; script:\n%s", second)
	}
	if n := strings.Count(second, "jump "+inChain("dep-2")); n != 1 {
		t.Errorf("re-Apply must add exactly one ingress forward jump for dep-2, got %d; script:\n%s", n, second)
	}
	if n := strings.Count(second, "jump "+outChain("dep-2")); n != 1 {
		t.Errorf("re-Apply must add exactly one egress forward jump for dep-2, got %d; script:\n%s", n, second)
	}
}

// lastScript returns the most recently captured nft script.
func lastScript(t *testing.T, c *scriptCapture) string {
	t.Helper()
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.scripts) == 0 {
		t.Fatal("no scripts captured")
	}
	return c.scripts[len(c.scripts)-1]
}

func TestNftPolicyEnforcer_ForwardJumpsRebuiltFromState(t *testing.T) {
	cap := &scriptCapture{}
	e := NewNftPolicyEnforcerWithExec(NewPolicyStore(), cap.exec)

	if err := e.Apply("dep-a", "10.88.0.2", NetworkPolicy{EgressMode: EgressDefaultDeny}); err != nil {
		t.Fatalf("Apply dep-a: %v", err)
	}
	// Applying a second deployment must re-emit dep-a's jumps too (rebuild from
	// the full applied set), so a single flush+rebuild keeps all live jumps.
	if err := e.Apply("dep-b", "10.88.0.3", NetworkPolicy{EgressMode: EgressDefaultDeny}); err != nil {
		t.Fatalf("Apply dep-b: %v", err)
	}
	second := lastScript(t, cap)
	for _, want := range []string{
		"jump " + inChain("dep-a"), "jump " + outChain("dep-a"),
		"jump " + inChain("dep-b"), "jump " + outChain("dep-b"),
	} {
		if !strings.Contains(second, want) {
			t.Errorf("second Apply must rebuild jumps for ALL applied deployments; missing %q\nscript:\n%s", want, second)
		}
	}

	// Removing dep-a must rebuild the forward chain WITHOUT dep-a's jumps (so
	// the subsequent delete chain does not hit EBUSY) but KEEP dep-b's jumps.
	cap.mu.Lock()
	cap.scripts = nil
	cap.mu.Unlock()
	if err := e.Remove("dep-a"); err != nil {
		t.Fatalf("Remove dep-a: %v", err)
	}
	rm := lastScript(t, cap)
	if !strings.Contains(rm, "flush chain inet "+policyTable+" forward") {
		t.Errorf("Remove must flush the forward chain before deleting the per-deployment chains; script:\n%s", rm)
	}
	// dep-a's jumps must be GONE from the rebuilt forward chain...
	if strings.Contains(rm, "jump "+inChain("dep-a")) || strings.Contains(rm, "jump "+outChain("dep-a")) {
		t.Errorf("Remove must NOT re-add the removed deployment's forward jumps; script:\n%s", rm)
	}
	// ...while dep-b's jumps must still be present (it is still applied).
	if !strings.Contains(rm, "jump "+inChain("dep-b")) || !strings.Contains(rm, "jump "+outChain("dep-b")) {
		t.Errorf("Remove must preserve the still-applied deployment's forward jumps; script:\n%s", rm)
	}
	// The forward flush+rebuild must come BEFORE the delete chain for dep-a,
	// otherwise nft returns EBUSY (chain still targeted by a jump).
	flushIdx := strings.Index(rm, "flush chain inet "+policyTable+" forward")
	delIdx := strings.Index(rm, "delete chain inet "+policyTable+" "+inChain("dep-a"))
	if flushIdx < 0 || delIdx < 0 || flushIdx > delIdx {
		t.Errorf("forward flush/rebuild must precede the delete chain for dep-a (EBUSY otherwise); script:\n%s", rm)
	}
}

func TestNftPolicyEnforcer_RemoveLastDeployment_EmptyForwardRebuild(t *testing.T) {
	cap := &scriptCapture{}
	e := NewNftPolicyEnforcerWithExec(NewPolicyStore(), cap.exec)

	if err := e.Apply("dep-solo", "10.88.0.42", NetworkPolicy{EgressMode: EgressDefaultDeny}); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	cap.mu.Lock()
	cap.scripts = nil
	cap.mu.Unlock()
	if err := e.Remove("dep-solo"); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	rm := lastScript(t, cap)
	// With no deployments left, the forward chain is flushed and left empty (no
	// stray jumps), then the per-deployment chains are deleted.
	if !strings.Contains(rm, "flush chain inet "+policyTable+" forward") {
		t.Errorf("Remove of last deployment must still flush the forward chain; script:\n%s", rm)
	}
	if strings.Contains(rm, "jump ") {
		t.Errorf("Remove of last deployment must leave no forward jumps; script:\n%s", rm)
	}
	if !strings.Contains(rm, "delete chain inet "+policyTable+" "+inChain("dep-solo")) {
		t.Errorf("Remove must delete the per-deployment chain; script:\n%s", rm)
	}
}

func TestNftPolicyEnforcer_Remove_DeletesChains(t *testing.T) {
	cap := &scriptCapture{}
	e := NewNftPolicyEnforcerWithExec(NewPolicyStore(), cap.exec)

	if err := e.Apply("dep-3", "10.88.0.12", NetworkPolicy{}); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if err := e.Remove("dep-3"); err != nil {
		t.Fatalf("Remove: %v", err)
	}

	got := cap.all()
	if !strings.Contains(got, "delete chain") {
		t.Errorf("Remove must emit delete chain; script:\n%s", got)
	}
	if !strings.Contains(got, "mb_dep_3_in") || !strings.Contains(got, "mb_dep_3_out") {
		t.Errorf("Remove must delete both _in and _out chains; script:\n%s", got)
	}
	if e.HasRules("dep-3") {
		t.Error("HasRules(dep-3) should be false after Remove")
	}
}

func TestNftPolicyEnforcer_Remove_NeverApplied_NoExec(t *testing.T) {
	cap := &scriptCapture{}
	e := NewNftPolicyEnforcerWithExec(NewPolicyStore(), cap.exec)

	// Removing a deployment that was never applied must not invoke nft (deleting
	// a non-existent chain would error).
	if err := e.Remove("never-applied"); err != nil {
		t.Fatalf("Remove of never-applied deployment: %v", err)
	}
	if cap.count() != 0 {
		t.Errorf("Remove of never-applied deployment should not exec nft; got %d invocations", cap.count())
	}
}

func TestNftPolicyEnforcer_Apply_ExecError_ReturnsError(t *testing.T) {
	sentinel := errors.New("nft boom")
	cap := &scriptCapture{err: sentinel}
	e := NewNftPolicyEnforcerWithExec(NewPolicyStore(), cap.exec)

	err := e.Apply("dep-4", "10.88.0.20", NetworkPolicy{EgressMode: EgressDefaultDeny})
	if err == nil {
		t.Fatal("Apply should return the execFn error")
	}
	if !errors.Is(err, sentinel) {
		t.Errorf("Apply error = %v, want wrap of %v", err, sentinel)
	}
	// On exec failure the deployment must NOT be marked applied (so a later
	// Remove won't try to delete chains that were never created).
	if e.HasRules("dep-4") {
		t.Error("HasRules(dep-4) should be false when Apply exec failed")
	}
}

func TestNftPolicyEnforcer_Apply_NoPortContainer(t *testing.T) {
	// The replica / no-port path: a container with no exposed ports still gets a
	// policy applied. This must succeed and still emit egress rules.
	cap := &scriptCapture{}
	e := NewNftPolicyEnforcerWithExec(NewPolicyStore(), cap.exec)

	policy := DefaultRestrictiveEgressPolicy()
	if err := e.Apply("dep-replica", "10.88.1.7", policy); err != nil {
		t.Fatalf("Apply (no-port replica): %v", err)
	}
	got := cap.all()
	if !strings.Contains(got, "10.88.1.7") {
		t.Errorf("no-port Apply must reference the container IP; script:\n%s", got)
	}
	if !strings.Contains(got, "169.254.169.254") {
		t.Errorf("restrictive policy egress-deny CIDR missing from script:\n%s", got)
	}
}

func TestNftPolicyEnforcer_Apply_InvalidInputs(t *testing.T) {
	cap := &scriptCapture{}
	e := NewNftPolicyEnforcerWithExec(NewPolicyStore(), cap.exec)

	if err := e.Apply("", "10.88.0.1", NetworkPolicy{}); err == nil {
		t.Error("empty deploymentID should error")
	}
	if err := e.Apply("dep-x", "", NetworkPolicy{}); err == nil {
		t.Error("empty containerIP should error")
	}
	if cap.count() != 0 {
		t.Errorf("invalid Apply must not invoke nft; got %d invocations", cap.count())
	}
}

func TestComputeIngressRules(t *testing.T) {
	rules := ComputeIngressRules("dep-a", "10.88.0.5", []string{"10.88.0.6", "10.88.0.7"})
	if len(rules) != 3 { // 2 accepts + 1 terminal drop
		t.Fatalf("expected 3 ingress rules, got %d: %v", len(rules), rules)
	}
	last := rules[len(rules)-1]
	if !strings.Contains(last, "drop") || !strings.Contains(last, intraHostCIDR) {
		t.Errorf("terminal rule should drop intra-host range; got %q", last)
	}
	if ComputeIngressRules("", "10.88.0.5", nil) != nil {
		t.Error("empty deploymentID should yield nil ingress rules")
	}
}
