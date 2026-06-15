package daemon

import (
	"bytes"
	"testing"

	"github.com/moltbunker/moltbunker/internal/networking"
	"github.com/moltbunker/moltbunker/internal/runtime"
)

// TestToImageSignature_NilIsOptOut verifies that a request carrying no signature
// produces a nil runtime.ImageSignature, keeping the R3 gate dormant.
func TestToImageSignature_NilIsOptOut(t *testing.T) {
	if got := toImageSignature(nil); got != nil {
		t.Fatalf("toImageSignature(nil) = %+v, want nil", got)
	}
}

// TestToImageSignature_Mapping verifies a full field-for-field mapping.
func TestToImageSignature_Mapping(t *testing.T) {
	spec := &ImageSignatureSpec{
		Digest:      "sha256:abc",
		PublisherID: "deadbeef",
		Signature:   []byte{1, 2, 3},
	}
	got := toImageSignature(spec)
	if got == nil {
		t.Fatal("toImageSignature returned nil for a non-nil spec")
	}
	if got.Digest != runtime.ImageDigest("sha256:abc") {
		t.Errorf("Digest = %q, want sha256:abc", got.Digest)
	}
	if got.PublisherID != "deadbeef" {
		t.Errorf("PublisherID = %q, want deadbeef", got.PublisherID)
	}
	if !bytes.Equal(got.Signature, []byte{1, 2, 3}) {
		t.Errorf("Signature = %v, want [1 2 3]", got.Signature)
	}
}

// TestToTrustPolicy_OptOutByDefault verifies the zero case is opt-out: no
// signature requirement, so the R3 gate never fires for a vanilla deploy.
func TestToTrustPolicy_OptOutByDefault(t *testing.T) {
	p := toTrustPolicy(false, nil)
	if p.RequireSignature {
		t.Error("RequireSignature should be false by default")
	}
	if len(p.TrustedPublishers) != 0 {
		t.Errorf("TrustedPublishers = %v, want empty", p.TrustedPublishers)
	}
}

// TestToTrustPolicy_RequireWithoutPublishersDowngrades guards against the
// deny-all sentinel: RequireSignature=true with no trusted publishers would
// reject every image, so it must be downgraded to off.
func TestToTrustPolicy_RequireWithoutPublishersDowngrades(t *testing.T) {
	p := toTrustPolicy(true, nil)
	if p.RequireSignature {
		t.Error("RequireSignature must be downgraded to false when no publishers are supplied (would deny-all)")
	}
}

// TestToTrustPolicy_RequireWithPublishers verifies enforcement is honored when
// the caller supplies both the flag and a trust list.
func TestToTrustPolicy_RequireWithPublishers(t *testing.T) {
	pubs := []string{"aa", "bb"}
	p := toTrustPolicy(true, pubs)
	if !p.RequireSignature {
		t.Error("RequireSignature should be true when publishers are supplied")
	}
	if len(p.TrustedPublishers) != 2 {
		t.Errorf("TrustedPublishers = %v, want 2 entries", p.TrustedPublishers)
	}
}

// TestResolveScanPolicy verifies the default policy is used and the CVE
// allowlist threads through.
func TestResolveScanPolicy(t *testing.T) {
	def := runtime.DefaultScanPolicy()
	p := resolveScanPolicy(nil)
	if p.BlockAtOrAbove != def.BlockAtOrAbove {
		t.Errorf("BlockAtOrAbove = %v, want %v", p.BlockAtOrAbove, def.BlockAtOrAbove)
	}
	if p.RequireScan {
		t.Error("RequireScan must stay false so a noop scanner never blocks a deploy")
	}
	p2 := resolveScanPolicy([]string{"CVE-2024-1"})
	if len(p2.IgnoreCVEs) != 1 || p2.IgnoreCVEs[0] != "CVE-2024-1" {
		t.Errorf("IgnoreCVEs = %v, want [CVE-2024-1]", p2.IgnoreCVEs)
	}
}

// TestToNetworkPolicy_NilIsAllowAll verifies a nil spec yields the allow-all
// default and reports "not present" so callers skip enforcement entirely.
func TestToNetworkPolicy_NilIsAllowAll(t *testing.T) {
	policy, present := toNetworkPolicy(nil)
	if present {
		t.Error("present should be false for a nil spec")
	}
	if policy.EgressMode != networking.EgressDefaultAllow {
		t.Errorf("EgressMode = %v, want EgressDefaultAllow", policy.EgressMode)
	}
	if len(policy.EgressDeny) != 0 || len(policy.EgressAllow) != 0 || len(policy.AllowedPeers) != 0 {
		t.Error("nil spec must produce an empty allow-all policy")
	}
}

// TestToNetworkPolicy_Mapping verifies the wire spec maps onto the networking
// type, including the EgressDeny bool -> EgressDefaultDeny mode translation and
// EgressBlock -> EgressDeny field rename.
func TestToNetworkPolicy_Mapping(t *testing.T) {
	spec := &NetworkPolicySpec{
		AllowedPeers: []string{"dep-1"},
		EgressDeny:   true,
		EgressAllow:  []string{"1.1.1.1/32"},
		EgressBlock:  []string{"169.254.169.254/32"},
	}
	policy, present := toNetworkPolicy(spec)
	if !present {
		t.Error("present should be true for a supplied spec")
	}
	if policy.EgressMode != networking.EgressDefaultDeny {
		t.Errorf("EgressMode = %v, want EgressDefaultDeny", policy.EgressMode)
	}
	if len(policy.AllowedPeers) != 1 || policy.AllowedPeers[0] != "dep-1" {
		t.Errorf("AllowedPeers = %v", policy.AllowedPeers)
	}
	if len(policy.EgressAllow) != 1 || policy.EgressAllow[0] != "1.1.1.1/32" {
		t.Errorf("EgressAllow = %v", policy.EgressAllow)
	}
	if len(policy.EgressDeny) != 1 || policy.EgressDeny[0] != "169.254.169.254/32" {
		t.Errorf("EgressDeny = %v (from EgressBlock)", policy.EgressDeny)
	}
}

// TestBuildImageScanner_DisabledIsNoop verifies scanning-off yields a noop
// scanner so the R4 gate can never block a deploy.
func TestBuildImageScanner_DisabledIsNoop(t *testing.T) {
	s := buildImageScanner(false)
	if s == nil {
		t.Fatal("buildImageScanner returned nil; must never be nil")
	}
	if s.ID() != "noop" {
		t.Errorf("scanner ID = %q, want noop when scanning is disabled", s.ID())
	}
}

// TestApplyNetworkPolicy_NilSkips verifies that a deploy with no network policy
// records nothing in the enforcer — i.e. allow-all / legacy behavior.
func TestApplyNetworkPolicy_NilSkips(t *testing.T) {
	store := networking.NewPolicyStore()
	enf := networking.NewNftPolicyEnforcer(store)
	cm := &ContainerManager{policyStore: store, policyEnforcer: enf}

	cm.applyNetworkPolicy("dep-x", "10.88.0.2", nil)

	if _, ok := store.Get("dep-x"); ok {
		t.Error("nil policy should not be recorded; expected allow-all (no enforcement)")
	}
}

// TestApplyNetworkPolicy_RecordsSuppliedPolicy verifies a supplied policy
// reaches the enforcer (recorded off-Linux; real nft on Linux).
func TestApplyNetworkPolicy_RecordsSuppliedPolicy(t *testing.T) {
	store := networking.NewPolicyStore()
	enf := networking.NewNftPolicyEnforcer(store)
	cm := &ContainerManager{policyStore: store, policyEnforcer: enf}

	spec := &NetworkPolicySpec{EgressDeny: true, EgressAllow: []string{"1.1.1.1/32"}}
	cm.applyNetworkPolicy("dep-y", "10.88.0.3", spec)

	got, ok := store.Get("dep-y")
	if !ok {
		t.Fatal("supplied policy should be recorded in the store")
	}
	if got.EgressMode != networking.EgressDefaultDeny {
		t.Errorf("recorded EgressMode = %v, want EgressDefaultDeny", got.EgressMode)
	}

	cm.removeNetworkPolicy("dep-y")
	if _, ok := store.Get("dep-y"); ok {
		t.Error("removeNetworkPolicy should drop the recorded policy")
	}
}

// TestEnforceDeployNetworkPolicy_NoPortDeploy verifies the no-port / replica
// path: enforceDeployNetworkPolicy allocates a port-less network to obtain a
// container IP and then applies the policy, so a container with no exposed
// ports still gets R13/R14 enforcement.
func TestEnforceDeployNetworkPolicy_NoPortDeploy(t *testing.T) {
	store := networking.NewPolicyStore()
	enf := networking.NewNftPolicyEnforcer(store)
	cm := &ContainerManager{
		policyStore:    store,
		policyEnforcer: enf,
		networkManager: networking.NewNetworkManager(),
	}

	spec := &NetworkPolicySpec{EgressDeny: true}
	cm.enforceDeployNetworkPolicy("dep-noport", spec)

	got, ok := store.Get("dep-noport")
	if !ok {
		t.Fatal("no-port deploy should still have its policy recorded in the store")
	}
	if got.EgressMode != networking.EgressDefaultDeny {
		t.Errorf("recorded EgressMode = %v, want EgressDefaultDeny", got.EgressMode)
	}
	// A port-less network must have been provisioned so the policy has an IP.
	if _, ok := cm.networkManager.GetNetwork("dep-noport"); !ok {
		t.Error("enforceDeployNetworkPolicy should provision a network for a no-port deploy")
	}
}

// TestEnforceDeployNetworkPolicy_NilPolicyNoNetwork verifies that a deploy with
// no policy does NOT provision a port-less network (no extra IP consumption)
// and records nothing — identical to legacy allow-all behavior.
func TestEnforceDeployNetworkPolicy_NilPolicyNoNetwork(t *testing.T) {
	store := networking.NewPolicyStore()
	enf := networking.NewNftPolicyEnforcer(store)
	cm := &ContainerManager{
		policyStore:    store,
		policyEnforcer: enf,
		networkManager: networking.NewNetworkManager(),
	}

	cm.enforceDeployNetworkPolicy("dep-nopolicy", nil)

	if _, ok := cm.networkManager.GetNetwork("dep-nopolicy"); ok {
		t.Error("a nil policy must not provision a network (no extra IP allocation)")
	}
	if _, ok := store.Get("dep-nopolicy"); ok {
		t.Error("a nil policy must not be recorded")
	}
}

// TestResolveContainerIP_ReusesExistingNetwork verifies that when a network
// already exists (port-exposing deploy) resolveContainerIP returns its IP
// rather than provisioning a second one.
func TestResolveContainerIP_ReusesExistingNetwork(t *testing.T) {
	cm := &ContainerManager{networkManager: networking.NewNetworkManager()}

	first, err := cm.networkManager.SetupNetwork("dep-existing", []networking.ExposedPort{{ContainerPort: 80}})
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	got := cm.resolveContainerIP("dep-existing")
	if got != first.ContainerIP() {
		t.Errorf("resolveContainerIP = %q, want existing network IP %q", got, first.ContainerIP())
	}
}
