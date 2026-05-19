package networking

import (
	"errors"
	"sort"
	"testing"
)

func TestNetworkPolicy_AllowsPeer(t *testing.T) {
	p := NetworkPolicy{AllowedPeers: []string{"dep-a", "dep-b"}}
	if !p.AllowsPeer("dep-a") {
		t.Fatal("expected dep-a to be allowed")
	}
	if !p.AllowsPeer("dep-b") {
		t.Fatal("expected dep-b to be allowed")
	}
	if p.AllowsPeer("dep-c") {
		t.Fatal("dep-c should not be allowed")
	}
	if (NetworkPolicy{}).AllowsPeer("dep-a") {
		t.Fatal("zero-value policy should allow no peer")
	}
}

func TestNetworkPolicy_Validate(t *testing.T) {
	cases := []struct {
		name    string
		policy  NetworkPolicy
		self    string
		wantErr bool
	}{
		{
			name:   "default policy validates",
			policy: DefaultNetworkPolicy(),
			self:   "dep-x",
		},
		{
			name:    "self-reference rejected",
			policy:  NetworkPolicy{AllowedPeers: []string{"dep-x"}},
			self:    "dep-x",
			wantErr: true,
		},
		{
			name:    "empty peer id rejected",
			policy:  NetworkPolicy{AllowedPeers: []string{""}},
			self:    "dep-x",
			wantErr: true,
		},
		{
			name:    "empty egress allow CIDR rejected",
			policy:  NetworkPolicy{EgressAllow: []string{""}},
			self:    "dep-x",
			wantErr: true,
		},
		{
			name:    "empty egress deny CIDR rejected",
			policy:  NetworkPolicy{EgressDeny: []string{""}},
			self:    "dep-x",
			wantErr: true,
		},
		{
			name: "valid full-feature policy",
			policy: NetworkPolicy{
				AllowedPeers: []string{"dep-a"},
				EgressMode:   EgressDefaultDeny,
				EgressAllow:  []string{"10.0.0.0/8"},
				EgressDeny:   []string{"169.254.169.254/32"},
			},
			self: "dep-x",
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := tc.policy.Validate(tc.self)
			if (err != nil) != tc.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tc.wantErr)
			}
			if tc.wantErr && !errors.Is(err, ErrInvalidPolicy) {
				t.Fatalf("err = %v, expected to wrap ErrInvalidPolicy", err)
			}
		})
	}
}

func TestPolicyStore_BasicCRUD(t *testing.T) {
	s := NewPolicyStore()
	p := NetworkPolicy{AllowedPeers: []string{"dep-b"}}

	if _, ok := s.Get("dep-a"); ok {
		t.Fatal("empty store should not have dep-a")
	}

	s.Set("dep-a", "10.88.0.2", p)

	got, ok := s.Get("dep-a")
	if !ok {
		t.Fatal("dep-a should be present after Set")
	}
	if !got.AllowsPeer("dep-b") {
		t.Fatal("stored policy lost AllowedPeers")
	}
	ip, ok := s.PeerIP("dep-a")
	if !ok || ip != "10.88.0.2" {
		t.Fatalf("PeerIP = %q, %v; want 10.88.0.2, true", ip, ok)
	}

	s.Remove("dep-a")
	if _, ok := s.Get("dep-a"); ok {
		t.Fatal("dep-a should be gone after Remove")
	}
	if _, ok := s.PeerIP("dep-a"); ok {
		t.Fatal("PeerIP should also be gone")
	}
}

func TestPolicyStore_ResolveAllowedPeerIPs_ReciprocityRequired(t *testing.T) {
	s := NewPolicyStore()

	// A allows B; B does NOT allow A.
	s.Set("dep-a", "10.88.0.2", NetworkPolicy{AllowedPeers: []string{"dep-b"}})
	s.Set("dep-b", "10.88.0.3", NetworkPolicy{}) // empty allow-list

	// Strict mode: A→B should be empty because B did not allow A back.
	strict := s.ResolveAllowedPeerIPs("dep-a", NetworkPolicy{AllowedPeers: []string{"dep-b"}}, false)
	if len(strict) != 0 {
		t.Fatalf("strict reciprocity: expected no peers, got %v", strict)
	}

	// Lax mode: A→B is allowed unilaterally.
	lax := s.ResolveAllowedPeerIPs("dep-a", NetworkPolicy{AllowedPeers: []string{"dep-b"}}, true)
	if len(lax) != 1 || lax[0] != "10.88.0.3" {
		t.Fatalf("lax mode: expected [10.88.0.3], got %v", lax)
	}
}

func TestPolicyStore_ResolveAllowedPeerIPs_BidirectionalAllow(t *testing.T) {
	s := NewPolicyStore()

	// Both allow each other.
	s.Set("dep-a", "10.88.0.2", NetworkPolicy{AllowedPeers: []string{"dep-b"}})
	s.Set("dep-b", "10.88.0.3", NetworkPolicy{AllowedPeers: []string{"dep-a"}})

	strict := s.ResolveAllowedPeerIPs("dep-a", NetworkPolicy{AllowedPeers: []string{"dep-b"}}, false)
	if len(strict) != 1 || strict[0] != "10.88.0.3" {
		t.Fatalf("expected [10.88.0.3], got %v", strict)
	}

	// And from B's perspective.
	strictRev := s.ResolveAllowedPeerIPs("dep-b", NetworkPolicy{AllowedPeers: []string{"dep-a"}}, false)
	if len(strictRev) != 1 || strictRev[0] != "10.88.0.2" {
		t.Fatalf("expected [10.88.0.2], got %v", strictRev)
	}
}

func TestPolicyStore_ResolveAllowedPeerIPs_MultiplePeers(t *testing.T) {
	s := NewPolicyStore()

	// A allows B, C, D. B and C allow A back; D does not.
	s.Set("dep-a", "10.88.0.2", NetworkPolicy{AllowedPeers: []string{"dep-b", "dep-c", "dep-d"}})
	s.Set("dep-b", "10.88.0.3", NetworkPolicy{AllowedPeers: []string{"dep-a"}})
	s.Set("dep-c", "10.88.0.4", NetworkPolicy{AllowedPeers: []string{"dep-a"}})
	s.Set("dep-d", "10.88.0.5", NetworkPolicy{}) // doesn't reciprocate

	ips := s.ResolveAllowedPeerIPs("dep-a", NetworkPolicy{AllowedPeers: []string{"dep-b", "dep-c", "dep-d"}}, false)
	sort.Strings(ips)
	want := []string{"10.88.0.3", "10.88.0.4"}
	if len(ips) != len(want) {
		t.Fatalf("got %v, want %v", ips, want)
	}
	for i := range want {
		if ips[i] != want[i] {
			t.Fatalf("got %v, want %v", ips, want)
		}
	}
}

func TestNftPolicyEnforcer_ApplyRemove(t *testing.T) {
	store := NewPolicyStore()
	e := NewNftPolicyEnforcer(store)

	if e.HasRules("dep-x") {
		t.Fatal("fresh enforcer should not have rules")
	}

	policy := NetworkPolicy{AllowedPeers: []string{"dep-y"}}
	if err := e.Apply("dep-x", "10.88.0.2", policy); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if !e.HasRules("dep-x") {
		t.Fatal("expected rules after Apply")
	}

	// Store was updated.
	got, ok := store.Get("dep-x")
	if !ok {
		t.Fatal("store should have dep-x")
	}
	if !got.AllowsPeer("dep-y") {
		t.Fatal("stored policy lost AllowedPeers")
	}

	if err := e.Remove("dep-x"); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	if e.HasRules("dep-x") {
		t.Fatal("expected no rules after Remove")
	}
	if _, ok := store.Get("dep-x"); ok {
		t.Fatal("store should not have dep-x after Remove")
	}
}

func TestNftPolicyEnforcer_RejectsInvalidPolicy(t *testing.T) {
	e := NewNftPolicyEnforcer(nil)
	// Self-reference: dep-x in dep-x's own allow list.
	err := e.Apply("dep-x", "10.88.0.2", NetworkPolicy{AllowedPeers: []string{"dep-x"}})
	if err == nil {
		t.Fatal("expected error for self-referencing policy")
	}
	if !errors.Is(err, ErrInvalidPolicy) {
		t.Fatalf("err = %v, want wraps ErrInvalidPolicy", err)
	}
}

func TestNftPolicyEnforcer_RemoveBeforeApplyIsSafe(t *testing.T) {
	e := NewNftPolicyEnforcer(nil)
	if err := e.Remove("never-applied"); err != nil {
		t.Fatalf("Remove on unknown deployment should be safe, got %v", err)
	}
}

func TestDefaultNetworkPolicy_IsLaterallyIsolated(t *testing.T) {
	p := DefaultNetworkPolicy()
	if len(p.AllowedPeers) != 0 {
		t.Fatalf("default policy should have no AllowedPeers, got %v", p.AllowedPeers)
	}
	if p.EgressMode != EgressDefaultAllow {
		t.Fatalf("default EgressMode should be EgressDefaultAllow, got %v", p.EgressMode)
	}
}
