package networking

import (
	"errors"
	"net"
	"strings"
	"testing"
)

// R14 — egress evaluation + rule-set computation tests.

func TestNetworkPolicy_EvaluateEgress_Precedence(t *testing.T) {
	policy := NetworkPolicy{
		EgressMode:  EgressDefaultAllow,
		EgressAllow: []string{"10.0.0.0/8"},
		// /24 inside the allowed /8 — explicit deny must beat explicit allow.
		EgressDeny: []string{"10.0.5.0/24"},
	}

	cases := []struct {
		ip   string
		want EgressDecision
	}{
		{"10.0.5.10", EgressBlocked},  // hit deny first
		{"10.0.6.10", EgressAllowed},  // hit allow after deny miss
		{"172.16.0.1", EgressAllowed}, // falls to EgressDefaultAllow
	}
	for _, c := range cases {
		got := policy.EvaluateEgressString(c.ip)
		if got != c.want {
			t.Errorf("EvaluateEgress(%q) = %v, want %v", c.ip, got, c.want)
		}
	}
}

func TestNetworkPolicy_EvaluateEgress_DefaultDeny(t *testing.T) {
	policy := NetworkPolicy{
		EgressMode:  EgressDefaultDeny,
		EgressAllow: []string{"1.1.1.1/32"},
	}
	if policy.EvaluateEgressString("1.1.1.1") != EgressAllowed {
		t.Fatal("1.1.1.1 should be allowed under default-deny + explicit allow")
	}
	if policy.EvaluateEgressString("8.8.8.8") != EgressBlocked {
		t.Fatal("8.8.8.8 should be blocked under default-deny + no allow")
	}
}

func TestNetworkPolicy_EvaluateEgress_DefaultAllow(t *testing.T) {
	policy := NetworkPolicy{EgressMode: EgressDefaultAllow}
	if policy.EvaluateEgressString("8.8.8.8") != EgressAllowed {
		t.Fatal("8.8.8.8 should be allowed under default-allow + no deny")
	}
}

func TestNetworkPolicy_EvaluateEgress_FailsClosedOnBadInput(t *testing.T) {
	policy := NetworkPolicy{EgressMode: EgressDefaultAllow}
	// nil IP → block
	if policy.EvaluateEgress(nil) != EgressBlocked {
		t.Fatal("nil IP should fail closed")
	}
	// unparseable string → block
	if policy.EvaluateEgressString("not-an-ip") != EgressBlocked {
		t.Fatal("unparseable IP string should fail closed")
	}
}

func TestNetworkPolicy_EvaluateEgress_IPv6(t *testing.T) {
	policy := NetworkPolicy{
		EgressMode: EgressDefaultDeny,
		EgressAllow: []string{
			"2001:4860:4860::/48", // Google's IPv6 DNS prefix area
		},
	}
	if policy.EvaluateEgressString("2001:4860:4860::8888") != EgressAllowed {
		t.Fatal("IPv6 in allow range should be allowed")
	}
	if policy.EvaluateEgressString("2001:db8::1") != EgressBlocked {
		t.Fatal("IPv6 outside allow range should be blocked under default-deny")
	}
}

func TestNetworkPolicy_EvaluateEgress_BoundaryAddresses(t *testing.T) {
	policy := NetworkPolicy{
		EgressMode:  EgressDefaultDeny,
		EgressAllow: []string{"10.1.1.0/24"},
	}
	// .0 is network address; should still match CIDR.
	if policy.EvaluateEgressString("10.1.1.0") != EgressAllowed {
		t.Fatal("network address should be inside the CIDR")
	}
	// .255 is broadcast; should still match CIDR.
	if policy.EvaluateEgressString("10.1.1.255") != EgressAllowed {
		t.Fatal("broadcast address should be inside the CIDR")
	}
	// .1 in adjacent /24 must NOT match.
	if policy.EvaluateEgressString("10.1.2.1") != EgressBlocked {
		t.Fatal("adjacent /24 should not match")
	}
}

func TestEgressDecision_String(t *testing.T) {
	if EgressAllowed.String() != "ALLOW" {
		t.Fatal("EgressAllowed.String() should be ALLOW")
	}
	if EgressBlocked.String() != "BLOCK" {
		t.Fatal("EgressBlocked.String() should be BLOCK")
	}
}

func TestNetworkPolicy_Validate_RejectsMalformedCIDR(t *testing.T) {
	cases := []struct {
		name   string
		policy NetworkPolicy
	}{
		{
			name:   "malformed allow CIDR",
			policy: NetworkPolicy{EgressAllow: []string{"10.0.0.0/33"}},
		},
		{
			name:   "non-CIDR allow",
			policy: NetworkPolicy{EgressAllow: []string{"not-a-cidr"}},
		},
		{
			name:   "malformed deny CIDR",
			policy: NetworkPolicy{EgressDeny: []string{"10.0.0.0/33"}},
		},
		{
			name:   "bare IP in deny (missing /32)",
			policy: NetworkPolicy{EgressDeny: []string{"169.254.169.254"}},
		},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			err := c.policy.Validate("dep-x")
			if err == nil {
				t.Fatal("expected error for malformed CIDR")
			}
			if !errors.Is(err, ErrInvalidPolicy) {
				t.Fatalf("err = %v, want wraps ErrInvalidPolicy", err)
			}
		})
	}
}

func TestComputeEgressRules_DefaultAllow_NoCIDRs(t *testing.T) {
	rules := ComputeEgressRules("dep-x", "10.88.0.2", NetworkPolicy{EgressMode: EgressDefaultAllow})
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule (default accept), got %d: %v", len(rules), rules)
	}
	if !strings.Contains(rules[0], "accept") {
		t.Fatalf("expected default accept rule, got %q", rules[0])
	}
	if !strings.Contains(rules[0], "10.88.0.2") {
		t.Fatalf("expected rule to reference container IP, got %q", rules[0])
	}
}

func TestComputeEgressRules_DefaultDeny_NoCIDRs(t *testing.T) {
	rules := ComputeEgressRules("dep-x", "10.88.0.2", NetworkPolicy{EgressMode: EgressDefaultDeny})
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule (default drop), got %d: %v", len(rules), rules)
	}
	if !strings.Contains(rules[0], "drop") {
		t.Fatalf("expected default drop rule, got %q", rules[0])
	}
}

func TestComputeEgressRules_DenyBeforeAllowBeforeDefault(t *testing.T) {
	policy := NetworkPolicy{
		EgressMode:  EgressDefaultAllow,
		EgressAllow: []string{"10.0.0.0/8", "172.16.0.0/12"},
		EgressDeny:  []string{"169.254.169.254/32", "10.0.5.0/24"},
	}
	rules := ComputeEgressRules("dep-x", "10.88.0.2", policy)

	// 2 denies + 2 allows + 1 default = 5 rules.
	if len(rules) != 5 {
		t.Fatalf("expected 5 rules, got %d: %v", len(rules), rules)
	}

	// First two must be deny rules.
	for i := 0; i < 2; i++ {
		if !strings.Contains(rules[i], "drop") {
			t.Fatalf("rule %d should be a deny, got %q", i, rules[i])
		}
	}
	// Next two must be allow rules.
	for i := 2; i < 4; i++ {
		if !strings.Contains(rules[i], "accept") {
			t.Fatalf("rule %d should be an allow, got %q", i, rules[i])
		}
	}
	// Last is the default fallback.
	if !strings.Contains(rules[4], "accept") {
		t.Fatalf("rule 4 should be default-allow, got %q", rules[4])
	}
}

func TestComputeEgressRules_EmptyInputReturnsNil(t *testing.T) {
	if ComputeEgressRules("", "10.88.0.2", DefaultNetworkPolicy()) != nil {
		t.Fatal("empty deploymentID should return nil")
	}
	if ComputeEgressRules("dep-x", "", DefaultNetworkPolicy()) != nil {
		t.Fatal("empty containerIP should return nil")
	}
}

func TestDefaultRestrictiveEgressPolicy_Validates(t *testing.T) {
	if err := DefaultRestrictiveEgressPolicy().Validate("dep-x"); err != nil {
		t.Fatalf("default restrictive policy should validate, got %v", err)
	}
}

func TestDefaultRestrictiveEgressPolicy_BlocksMetadataServer(t *testing.T) {
	p := DefaultRestrictiveEgressPolicy()
	if p.EvaluateEgressString("169.254.169.254") != EgressBlocked {
		t.Fatal("default restrictive policy should block cloud metadata server")
	}
	// And it should allow 1.1.1.1 which we explicitly carved out.
	if p.EvaluateEgressString("1.1.1.1") != EgressAllowed {
		t.Fatal("default restrictive policy should allow 1.1.1.1 DNS")
	}
	// Random public IP should be blocked under default-deny.
	if p.EvaluateEgressString("93.184.216.34") != EgressBlocked {
		t.Fatal("random public IP should be blocked under default-deny")
	}
	// RFC1918 should be blocked even though we never reached the default.
	if p.EvaluateEgressString("10.10.10.10") != EgressBlocked {
		t.Fatal("RFC1918 should be blocked by explicit deny")
	}
}

func TestNftPolicyEnforcer_LastRules_ReturnsRuleSet(t *testing.T) {
	e := NewNftPolicyEnforcer(nil)
	policy := NetworkPolicy{
		EgressMode:  EgressDefaultDeny,
		EgressAllow: []string{"1.1.1.1/32"},
	}
	if err := e.Apply("dep-x", "10.88.0.2", policy); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	rules := e.LastRules("dep-x")
	if len(rules) == 0 {
		t.Fatal("expected LastRules to return the computed rule set")
	}
	// Should mention the allow CIDR and the default drop.
	joined := strings.Join(rules, "\n")
	if !strings.Contains(joined, "1.1.1.1/32") {
		t.Fatal("LastRules missing allow CIDR")
	}
	if !strings.Contains(joined, "drop") {
		t.Fatal("LastRules missing drop default")
	}

	// After Remove, LastRules should be empty.
	_ = e.Remove("dep-x")
	if e.LastRules("dep-x") != nil {
		t.Fatal("LastRules should be nil after Remove")
	}
}

func TestNftPolicyEnforcer_LastRules_UnknownDeploymentReturnsNil(t *testing.T) {
	e := NewNftPolicyEnforcer(nil)
	if e.LastRules("never-applied") != nil {
		t.Fatal("LastRules on unknown deployment should return nil")
	}
}

// Sanity check: a deploy-time check using EvaluateEgress matches the rule
// that would actually fire in nftables. We don't run nft here, but the
// invariant matters: the Go evaluator and the nft ruleset must agree on
// each IP's outcome.
func TestNetworkPolicy_EvaluateEgress_AgreesWithRuleOrder(t *testing.T) {
	policy := NetworkPolicy{
		EgressMode:  EgressDefaultAllow,
		EgressAllow: []string{"10.0.0.0/8"},
		EgressDeny:  []string{"10.0.5.0/24"},
	}
	// Walk a few representative IPs.
	cases := []struct {
		ip   string
		want EgressDecision
	}{
		{"10.0.5.50", EgressBlocked},  // deny wins
		{"10.1.5.50", EgressAllowed},  // allow CIDR
		{"192.0.2.1", EgressAllowed},  // default allow
		{"10.0.5.255", EgressBlocked}, // deny boundary
	}
	for _, c := range cases {
		c := c
		t.Run(c.ip, func(t *testing.T) {
			got := policy.EvaluateEgress(net.ParseIP(c.ip))
			if got != c.want {
				t.Fatalf("EvaluateEgress(%s) = %v, want %v", c.ip, got, c.want)
			}
		})
	}
}
