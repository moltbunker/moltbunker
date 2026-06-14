//go:build e2e

package golden

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"errors"
	"io"
	"math/big"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"

	"github.com/moltbunker/moltbunker/internal/networking"
	"github.com/moltbunker/moltbunker/internal/payment"
	"github.com/moltbunker/moltbunker/internal/runtime"
	"github.com/moltbunker/moltbunker/pkg/types"
	"github.com/moltbunker/moltbunker/tests/e2e/testutil"
)

// TestGoldenPath_FullProductPromise walks the entire product promise in one
// in-process flow, one labeled t.Run phase at a time. Each phase logs whether
// its leg is MOCK or REAL. State that crosses phase boundaries (jobID, digest,
// signing key, deploymentID) is declared once at the top of the test and shared
// — phases are intentionally ordered and dependent.
//
// Mock-vs-real ledger (also annotated per-phase):
//
//	1 WalletGen          [REAL: crypto/rand key generation]
//	2 Stake              [MOCK: in-memory MockStakingContract]
//	3 EscrowReserve      [MOCK: in-memory MockEscrowContract]
//	4 ImageSigGate       [REAL: runtime.EdImageVerifier + SignImageDigest]
//	5 ScanGate           [MOCK: MockImageScanner, no trivy binary]
//	6 ContainerDeploy    [MOCK: MockContainerdClient]
//	7 NetworkPolicyAssert[REAL: networking.ComputeEgressRules, no nft exec]
//	8 IngressTunnel      [MOCK: httptest origin + loopback HTTP; no yamux/tunnel]
//	9 SubdomainDNS       [REAL: ingress.Resolver 5-step pipeline; no DNS/ACME]
//	10 StopContainer     [MOCK: MockContainerdClient]
//	11 EscrowFinalize    [MOCK: in-memory MockEscrowContract]
func TestGoldenPath_FullProductPromise(t *testing.T) {
	assert := testutil.NewAssertions(t)

	rootCtx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	h := NewGoldenHarness(t)

	// ---- Shared cross-phase state ----------------------------------------
	const deploymentID = "dep-a1b2c3d4e5f60718293a4b5c6d7e8f90"
	const imageRef = "registry.moltbunker.dev/golden/app:v1"
	jobID := MakeJobID(deploymentID)
	provider := common.HexToAddress("0x1111111111111111111111111111111111111111")
	escrowAmount := BunkerToWei(100)
	durationSecs := big.NewInt(3600) // 1 hour

	// A fixed image digest signed in Phase 4 and "deployed" in Phase 6.
	digest := runtime.ImageDigest("sha256:" +
		"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855")

	// Image-signing keypair, generated in Phase 1, used in Phase 4.
	var imgPub ed25519.PublicKey
	var imgPriv ed25519.PrivateKey

	// ----------------------------------------------------------------------
	// Phase 1: WalletGen — [REAL: crypto/rand key generation]
	// ----------------------------------------------------------------------
	t.Run("01_WalletGen", func(t *testing.T) {
		t.Log("[REAL: crypto/rand] generating ephemeral node + wallet keys in-memory")
		a := testutil.NewAssertions(t)

		var err error
		imgPub, imgPriv, err = ed25519.GenerateKey(rand.Reader)
		a.NoError(err, "ed25519 node/signing key generation should succeed")
		a.Equal(ed25519.PublicKeySize, len(imgPub), "ed25519 public key should be 32 bytes")

		// Ethereum wallet — in-memory only, never written to a keystore.
		ethKey, err := ethcrypto.GenerateKey()
		a.NoError(err, "ethereum wallet generation should succeed")
		addr := ethcrypto.PubkeyToAddress(ethKey.PublicKey)
		a.NotEqual(common.Address{}, addr, "wallet address should be non-zero")
	})

	// ----------------------------------------------------------------------
	// Phase 2: Stake — [MOCK: in-memory MockStakingContract]
	// ----------------------------------------------------------------------
	t.Run("02_Stake", func(t *testing.T) {
		// Tier thresholds come from the mock StakingContract (BunkerStaking.sol):
		// Starter 1M, Bronze 5M, Silver 10M, Gold 100M, Platinum 1B BUNKER.
		// IsActiveProvider requires >= Starter (1M). We stake 5M -> active + Bronze.
		t.Log("[MOCK: staking contract] staking 5,000,000 BUNKER -> active, Bronze tier")
		a := testutil.NewAssertions(t)
		ctx, c := context.WithTimeout(rootCtx, 15*time.Second)
		defer c()

		_, err := h.Staking.Stake(ctx, BunkerToWei(5_000_000))
		a.NoError(err, "provider stake should succeed")

		self := common.Address{} // mock uses the zero address as "self"
		active, err := h.Staking.IsActiveProvider(ctx, self)
		a.NoError(err)
		a.True(active, "provider should be active after staking 5,000,000 BUNKER")

		tier, err := h.Staking.GetTier(ctx, self)
		a.NoError(err)
		a.Equal(types.StakingTierBronze, tier, "5,000,000 BUNKER should be Bronze tier")
	})

	// ----------------------------------------------------------------------
	// Phase 3: EscrowReserve — [MOCK: in-memory MockEscrowContract]
	// ----------------------------------------------------------------------
	t.Run("03_EscrowReserve", func(t *testing.T) {
		t.Log("[MOCK: escrow contract] CreateEscrow -> SelectProviders -> Active")
		a := testutil.NewAssertions(t)
		ctx, c := context.WithTimeout(rootCtx, 15*time.Second)
		defer c()

		_, err := h.Escrow.CreateEscrow(ctx, jobID, provider, escrowAmount, durationSecs)
		a.NoError(err, "CreateEscrow should succeed")

		esc, err := h.Escrow.GetEscrow(ctx, jobID)
		a.NoError(err)
		a.Equal(payment.EscrowStateCreated.String(), esc.State.String(),
			"escrow should be Created before provider selection")

		providers := [3]common.Address{
			provider,
			common.HexToAddress("0x2222222222222222222222222222222222222222"),
			common.HexToAddress("0x3333333333333333333333333333333333333333"),
		}
		_, err = h.Escrow.SelectProviders(ctx, jobID, providers)
		a.NoError(err, "SelectProviders should succeed")

		esc, err = h.Escrow.GetEscrow(ctx, jobID)
		a.NoError(err)
		a.Equal(payment.EscrowStateActive.String(), esc.State.String(),
			"escrow should be Active after provider selection")
	})

	// ----------------------------------------------------------------------
	// Phase 4: ImageSigGate — [REAL: EdImageVerifier + SignImageDigest]
	// ----------------------------------------------------------------------
	t.Run("04_ImageSigGate", func(t *testing.T) {
		t.Log("[REAL: Ed25519 sig verify] valid sig passes; unsigned w/ RequireSignature blocks")
		a := testutil.NewAssertions(t)

		sig := runtime.SignImageDigest(digest, imgPriv)
		policy := runtime.TrustPolicy{
			RequireSignature:  true,
			TrustedPublishers: []string{sig.PublisherID},
		}
		a.NoError(h.Verifier.Verify(digest, sig, policy),
			"valid signature from trusted publisher should verify")

		// Unsigned image under RequireSignature must be rejected.
		err := h.Verifier.Verify(digest, nil, policy)
		a.True(errors.Is(err, runtime.ErrSignatureRequired),
			"unsigned image with RequireSignature=true must return ErrSignatureRequired")

		// Record that the signature gate was evaluated for this deploy.
		h.Deployer.Last.ImageRef = imageRef
		h.Deployer.Last.Digest = digest
		h.Deployer.Last.SignatureChecked = true
	})

	// ----------------------------------------------------------------------
	// Phase 5: ScanGate — [MOCK: MockImageScanner, no trivy binary]
	// ----------------------------------------------------------------------
	t.Run("05_ScanGate", func(t *testing.T) {
		t.Log("[MOCK: scanner] clean report passes DefaultScanPolicy; CRITICAL blocks")
		a := testutil.NewAssertions(t)
		ctx, c := context.WithTimeout(rootCtx, 15*time.Second)
		defer c()

		policy := runtime.DefaultScanPolicy()

		// Clean image — zero findings.
		h.Scanner.FindingsToReturn = []runtime.Vulnerability{}
		report, err := h.Scanner.Scan(ctx, imageRef)
		a.NoError(err, "scan should succeed")
		_, err = policy.Apply(report.Vulnerabilities)
		a.NoError(err, "clean image should pass DefaultScanPolicy")

		// Inject a CRITICAL finding — must block.
		h.Scanner.FindingsToReturn = []runtime.Vulnerability{{
			ID:       "CVE-2026-0001",
			Severity: runtime.SeverityCritical,
			Package:  "libgolden",
			Version:  "1.0.0",
		}}
		report, err = h.Scanner.Scan(ctx, imageRef)
		a.NoError(err)
		_, err = policy.Apply(report.Vulnerabilities)
		a.True(errors.Is(err, runtime.ErrPolicyViolation),
			"CRITICAL finding must block deploy with ErrPolicyViolation")

		// Reset to clean for the deploy phase and record the gate ran.
		h.Scanner.FindingsToReturn = []runtime.Vulnerability{}
		h.Deployer.Last.ScanChecked = true
	})

	// ----------------------------------------------------------------------
	// Phase 6: ContainerDeploy — [MOCK: MockContainerdClient]
	// ----------------------------------------------------------------------
	t.Run("06_ContainerDeploy", func(t *testing.T) {
		t.Log("[MOCK: containerd] create + start container; assert security gate was invoked")
		a := testutil.NewAssertions(t)
		ctx, c := context.WithTimeout(rootCtx, 15*time.Second)
		defer c()

		resources := types.ResourceLimits{
			CPUQuota:    100000,
			CPUPeriod:   100000,
			MemoryLimit: 256 * 1024 * 1024,
			DiskLimit:   1024 * 1024 * 1024,
			PIDLimit:    100,
		}
		_, err := h.Base.Containerd.CreateContainer(ctx, deploymentID, imageRef, resources)
		a.NoError(err, "CreateContainer should succeed")

		err = h.Base.Containerd.StartContainer(ctx, deploymentID)
		a.NoError(err, "StartContainer should succeed")

		err = h.Base.WaitForContainer(deploymentID, string(types.ContainerStatusRunning), 5*time.Second)
		a.NoError(err, "container should reach Running")

		// Record the deploy decision and assert that the gate-LOGIC phases (4 and
		// 5) ran in order before this deploy: phase 4 sets SignatureChecked after a
		// real EdImageVerifier.Verify, phase 5 sets ScanChecked after a real
		// ScanPolicy.Apply. This proves the ordered flow exercised the gates — it
		// does NOT prove the production deploy path wires them. The real
		// deployLocally (internal/daemon/container_manager.go) and deployReplica
		// (internal/daemon/replication.go) are never invoked here; the recorder is
		// a test-owned stand-in, so a regression that leaves those paths dormant
		// would NOT be caught by this assertion.
		// TODO(E2E-01 follow-up): to actually catch the built-but-dormant gate
		// regression, drive a real SecureContainerConfig deploy (or assert
		// deployLocally/deployReplica populate ImageVerify/ScanPolicy/egress config).
		// Tracked in plan/06-roadmap/daemon-todo.md under the OPS-04 / R11 follow-ups.
		a.NoError(h.Deployer.Deploy(h.Deployer.Last), "secure deploy should record")
		h.Deployer.AssertGatesInvoked(t)
	})

	// ----------------------------------------------------------------------
	// Phase 7: NetworkPolicyAssert — [REAL: ComputeEgressRules, no nft exec]
	// ----------------------------------------------------------------------
	t.Run("07_NetworkPolicyAssert", func(t *testing.T) {
		t.Log("[REAL: rule generator] default-deny egress blocks IMDS, allows DNS resolvers")
		a := testutil.NewAssertions(t)

		containerIP := "10.88.0.42"
		policy := networking.DefaultRestrictiveEgressPolicy() // EgressDefaultDeny
		a.NoError(policy.Validate(deploymentID), "restrictive policy should be valid")

		rules := networking.ComputeEgressRules(deploymentID, containerIP, policy)
		a.NotEmpty(rules, "rule set should be non-empty")

		assertRuleSetContains(t, rules, "169.254.169.254/32", "drop",
			"IMDS (169.254.169.254) must be dropped")
		assertRuleSetContains(t, rules, "1.1.1.1/32", "accept",
			"Cloudflare DNS resolver must be allowed")

		// Pure-function egress evaluation (no kernel): deny beats default, allow wins.
		a.Equal(networking.EgressBlocked, policy.EvaluateEgressString("169.254.169.254"),
			"IMDS address must be blocked by EvaluateEgress")
		a.Equal(networking.EgressAllowed, policy.EvaluateEgressString("1.1.1.1"),
			"DNS resolver must be allowed by EvaluateEgress")
		a.Equal(networking.EgressBlocked, policy.EvaluateEgressString("93.184.216.34"),
			"arbitrary public IP must be blocked under default-deny")
	})

	// ----------------------------------------------------------------------
	// Phase 8: IngressTunnel — [MOCK: httptest origin + loopback HTTP]
	// ----------------------------------------------------------------------
	t.Run("08_IngressTunnel", func(t *testing.T) {
		t.Log("[MOCK: tunnel=loopback HTTP, yamux=none] public request -> origin -> HTTPS 200")
		a := testutil.NewAssertions(t)
		ctx, c := context.WithTimeout(rootCtx, 15*time.Second)
		defer c()

		h.Ingress.Seed(deploymentID)

		// Resolve the subdomain to the provider address (the resolver pipeline),
		// then "tunnel" the request straight to the loopback origin. No yamux,
		// no real reverse tunnel — clearly MOCK.
		subdomain := bareID(deploymentID)[:8]
		entry, err := h.Ingress.Resolver.Resolve(subdomain)
		a.NoError(err, "resolver should map the subdomain to a service entry")
		a.NotNil(entry, "resolved entry should be non-nil")

		// Plain-HTTP leg via the resolved provider address (loopback dial).
		dialAddr := entry.ProviderAddr
		conn, err := net.DialTimeout("tcp", dialAddr, 5*time.Second)
		a.NoError(err, "loopback dial to resolved provider address should succeed")
		if conn != nil {
			_ = conn.Close()
		}

		resp, err := httpGet(ctx, h.Ingress.OriginURL())
		a.NoError(err, "GET to the loopback origin should succeed")
		if resp != nil {
			a.Equal(http.StatusOK, resp.StatusCode, "origin should return 200")
		}

		// HTTPS leg: the public promise is HTTPS-200. httptest TLS server stands
		// in for the edge's terminating TLS (mock cert). [MOCK: TLS=self-signed]
		tlsResp, err := h.Ingress.TLSClient().Get(h.Ingress.TLSOriginURL())
		a.NoError(err, "HTTPS GET to the TLS origin should succeed")
		if tlsResp != nil {
			defer func() { _ = tlsResp.Body.Close() }()
			a.Equal(http.StatusOK, tlsResp.StatusCode, "HTTPS origin should return 200")
			body, _ := io.ReadAll(tlsResp.Body)
			a.Contains(string(body), goldenPathBody, "HTTPS body should be the golden-path marker")
		}
	})

	// ----------------------------------------------------------------------
	// Phase 9: SubdomainDNS — [REAL: ingress.Resolver pipeline; no DNS/ACME]
	// ----------------------------------------------------------------------
	t.Run("09_SubdomainDNS", func(t *testing.T) {
		t.Log("[REAL: resolver] 8-char prefix resolves to the seeded deployment")
		a := testutil.NewAssertions(t)

		prefix := bareID(deploymentID)[:8]
		entry, err := h.Ingress.Resolver.Resolve(prefix)
		a.NoError(err, "prefix resolution should succeed")
		a.Equal(deploymentID, entry.DeploymentID, "resolved entry should match the deployment")

		// Unknown name must fail with "service not found".
		_, err = h.Ingress.Resolver.Resolve("ffffffffdeadbeef")
		a.Error(err, "unknown subdomain should not resolve")
		a.ErrorContains(err, "service not found", "error should be the not-found sentinel")
	})

	// ----------------------------------------------------------------------
	// Phase 10: StopContainer — [MOCK: MockContainerdClient]
	// ----------------------------------------------------------------------
	t.Run("10_StopContainer", func(t *testing.T) {
		t.Log("[MOCK: containerd] stop container; wait for Stopped")
		a := testutil.NewAssertions(t)
		ctx, c := context.WithTimeout(rootCtx, 15*time.Second)
		defer c()

		err := h.Base.Containerd.StopContainer(ctx, deploymentID, 10*time.Second)
		a.NoError(err, "StopContainer should succeed")

		err = h.Base.WaitForContainer(deploymentID, string(types.ContainerStatusStopped), 5*time.Second)
		a.NoError(err, "container should reach Stopped")

		// Service is no longer exposed once stopped.
		h.Ingress.Resolver.Unregister(deploymentID)
	})

	// ----------------------------------------------------------------------
	// Phase 11: EscrowFinalize — [MOCK: in-memory MockEscrowContract]
	// ----------------------------------------------------------------------
	t.Run("11_EscrowFinalize", func(t *testing.T) {
		t.Log("[MOCK: escrow contract] release full duration -> finalize -> Completed")
		a := testutil.NewAssertions(t)
		ctx, c := context.WithTimeout(rootCtx, 15*time.Second)
		defer c()

		_, err := h.Escrow.ReleasePayment(ctx, jobID, durationSecs) // full duration
		a.NoError(err, "ReleasePayment for full duration should succeed")

		esc, err := h.Escrow.GetEscrow(ctx, jobID)
		a.NoError(err)
		a.Equal(escrowAmount.String(), esc.Released.String(),
			"full escrow amount should be released after full duration")

		_, err = h.Escrow.FinalizeEscrow(ctx, jobID)
		a.NoError(err, "FinalizeEscrow should succeed")

		esc, err = h.Escrow.GetEscrow(ctx, jobID)
		a.NoError(err)
		a.Equal(payment.EscrowStateCompleted.String(), esc.State.String(),
			"escrow should be Completed after finalization")
		a.Equal(esc.Amount.String(), esc.Released.String(),
			"released should equal amount at settlement")
	})

	// End-state invariants for the whole flow (not a tautology): exactly one
	// secure-deploy decision was recorded in phase 6, and the escrow settled to
	// Completed with the full amount released in phase 11.
	assert.Equal(1, h.Deployer.DeployCnt,
		"exactly one secure deploy should have been recorded across the golden path")
	finalEsc, err := h.Escrow.GetEscrow(rootCtx, jobID)
	assert.NoError(err, "final escrow read should succeed")
	assert.Equal(payment.EscrowStateCompleted.String(), finalEsc.State.String(),
		"escrow must be Completed at end of the golden path")
	assert.Equal(finalEsc.Amount.String(), finalEsc.Released.String(),
		"released must equal amount once the golden path has settled")
}

// httpGet issues a context-bound GET, drains and closes the body, and returns
// the response with its body already closed (caller only inspects StatusCode).
func httpGet(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()
	return resp, nil
}

// assertRuleSetContains fails unless some rule line contains both the cidr and
// the verb (drop/accept) substrings. Keeps the egress-rule assertions readable.
func assertRuleSetContains(t *testing.T, rules []string, cidr, verb, msg string) {
	t.Helper()
	for _, r := range rules {
		if contains(r, cidr) && contains(r, verb) {
			return
		}
	}
	t.Errorf("%s: no rule contains %q and %q\n  rules: %v", msg, cidr, verb, rules)
}
