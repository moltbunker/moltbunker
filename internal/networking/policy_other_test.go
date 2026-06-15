//go:build !linux

package networking

import (
	"errors"
	"testing"
)

// TestStubApply_InvalidInputs asserts the non-Linux stub enforcer rejects the
// same malformed inputs the Linux enforcer rejects (empty deploymentID / empty
// containerIP). Without this parity, a caller bug that passes an empty ID would
// be silently accepted on dev machines and only surface in production on Linux.
func TestStubApply_InvalidInputs(t *testing.T) {
	e := NewNftPolicyEnforcer(NewPolicyStore())

	if err := e.Apply("", "10.88.0.1", NetworkPolicy{}); err == nil {
		t.Error("stub Apply: empty deploymentID should error")
	} else if !errors.Is(err, ErrInvalidPolicy) {
		t.Errorf("stub Apply: empty deploymentID error = %v, want wrap of ErrInvalidPolicy", err)
	}

	if err := e.Apply("dep-x", "", NetworkPolicy{}); err == nil {
		t.Error("stub Apply: empty containerIP should error")
	} else if !errors.Is(err, ErrInvalidPolicy) {
		t.Errorf("stub Apply: empty containerIP error = %v, want wrap of ErrInvalidPolicy", err)
	}

	// A self-referencing peer list must still be rejected by policy.Validate.
	if err := e.Apply("dep-x", "10.88.0.1", NetworkPolicy{AllowedPeers: []string{"dep-x"}}); err == nil {
		t.Error("stub Apply: self-referencing peer should error")
	}

	// A valid Apply must succeed and record intent.
	if err := e.Apply("dep-ok", "10.88.0.5", NetworkPolicy{EgressMode: EgressDefaultDeny}); err != nil {
		t.Fatalf("stub Apply (valid): %v", err)
	}
	if !e.HasRules("dep-ok") {
		t.Error("stub HasRules(dep-ok) should be true after a valid Apply")
	}
}

// TestStubEgressRules_ChainNameSanitized asserts that the platform-agnostic
// ComputeEgressRules produces sanitized (hyphen-free) chain names, so the names
// match what the Linux enforcer creates/flushes/deletes.
func TestStubEgressRules_ChainNameSanitized(t *testing.T) {
	rules := ComputeEgressRules("dep-7f3a", "10.88.0.9", NetworkPolicy{EgressMode: EgressDefaultDeny})
	if len(rules) == 0 {
		t.Fatal("expected egress rules")
	}
	for _, r := range rules {
		if !contains(r, "mb_dep_7f3a_out") {
			t.Errorf("egress rule must reference the sanitized chain mb_dep_7f3a_out; got %q", r)
		}
		if contains(r, "mb_dep-7f3a_out") {
			t.Errorf("egress rule must NOT contain the unsanitized hyphenated chain name; got %q", r)
		}
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
