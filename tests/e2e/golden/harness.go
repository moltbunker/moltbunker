//go:build e2e

// Package golden contains the single golden-path acceptance test that exercises
// every major moltbunker product promise in sequence across one in-process
// harness: wallet keygen -> on-chain escrow reserve -> image verify+scan gate ->
// container start -> subdomain resolve -> tunnel -> public HTTPS 200 -> stop ->
// escrow finalize.
//
// MOCK-VS-REAL POLICY (read this before changing anything):
//
// Several legs of the real product cannot run on a darwin CI runner (no Linux
// containerd, no nftables, no live chain, no public DNS/ACME). For those legs
// this harness substitutes an in-process mock and MARKS IT CLEARLY both in a
// per-phase comment and in a t.Log("[MOCK: ...]") line emitted at runtime. Legs
// that ARE real (pure-Go crypto, policy rule computation, resolver pipeline) are
// marked "[REAL: ...]". The real-component legs (live containerd, live anvil)
// are gated behind environment variables / separate build tags so the default
// `go test -tags e2e ./tests/e2e/golden/...` stays green on darwin.
package golden

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/moltbunker/moltbunker/internal/ingress"
	"github.com/moltbunker/moltbunker/internal/payment"
	"github.com/moltbunker/moltbunker/internal/runtime"
	"github.com/moltbunker/moltbunker/tests/e2e/testutil"
)

// GoldenHarness bundles every mock/real component the golden-path test drives.
// It composes the existing testutil.TestHarness (which owns the mock containerd,
// temp dirs, and context lifecycle) rather than reinventing that plumbing.
type GoldenHarness struct {
	t   *testing.T
	ctx context.Context

	// Base harness — owns the mock containerd client, temp dirs, context.
	Base *testutil.TestHarness

	// Payment mocks — in-memory contract simulators (no chain).
	Staking *payment.StakingContract // payment.NewMockStakingContract()
	Escrow  *payment.EscrowContract  // payment.NewMockEscrowContract()

	// Runtime gate components.
	Verifier *runtime.EdImageVerifier // REAL Ed25519 verifier
	Scanner  *MockImageScanner        // MOCK scanner (no trivy binary)

	// Secure-deploy recorder — confirms the sig/scan gate was invoked.
	Deployer *MockSecureDeployer

	// Ingress — loopback HTTPS endpoint + resolver pipeline.
	Ingress *LoopbackIngress
}

// NewGoldenHarness wires every component together and registers teardown.
func NewGoldenHarness(t *testing.T) *GoldenHarness {
	t.Helper()

	base := testutil.NewTestHarness(t)

	h := &GoldenHarness{
		t:        t,
		ctx:      base.Context(),
		Base:     base,
		Staking:  payment.NewMockStakingContract(), // [MOCK: staking contract]
		Escrow:   payment.NewMockEscrowContract(),  // [MOCK: escrow contract]
		Verifier: runtime.NewEdImageVerifier(),     // [REAL: sig verification]
		Scanner:  NewMockImageScanner(),            // [MOCK: scanner, no trivy]
		Deployer: NewMockSecureDeployer(),
		Ingress:  NewLoopbackIngress(t),
	}

	t.Cleanup(h.Teardown)
	return h
}

// Context returns the harness context (cancelled on teardown).
func (h *GoldenHarness) Context() context.Context { return h.ctx }

// Teardown stops the loopback ingress server. The base harness cleanup is
// registered separately by testutil.NewTestHarness.
func (h *GoldenHarness) Teardown() {
	if h.Ingress != nil {
		h.Ingress.Close()
	}
}

// MakeJobID derives the on-chain [32]byte job id from a deployment id string.
func MakeJobID(deploymentID string) [32]byte {
	return payment.JobIDFromString(deploymentID)
}

// -----------------------------------------------------------------------------
// MockImageScanner — implements runtime.ImageScanner without a trivy binary.
// -----------------------------------------------------------------------------

// MockImageScanner returns a configurable ScanReport. It records how many times
// Scan was called so callers can assert the gate ran.
type MockImageScanner struct {
	// FindingsToReturn is the vulnerability set the next Scan returns. Empty
	// means a clean image.
	FindingsToReturn []runtime.Vulnerability
	// ScanCallCount counts Scan invocations.
	ScanCallCount int
}

// NewMockImageScanner returns a scanner that reports zero findings by default.
func NewMockImageScanner() *MockImageScanner {
	return &MockImageScanner{FindingsToReturn: []runtime.Vulnerability{}}
}

// ID identifies the mock scanner backend.
func (s *MockImageScanner) ID() string { return "mock-scanner:v0" }

// Scan returns a ScanReport carrying FindingsToReturn and increments the counter.
func (s *MockImageScanner) Scan(_ context.Context, ref string) (*runtime.ScanReport, error) {
	s.ScanCallCount++
	return &runtime.ScanReport{
		ImageRef:        ref,
		ScannerID:       s.ID(),
		Vulnerabilities: s.FindingsToReturn,
		ScanStartedAt:   time.Now(),
	}, nil
}

// -----------------------------------------------------------------------------
// MockSecureDeployer — records that the sig/scan gate was evaluated for a deploy.
// -----------------------------------------------------------------------------

// GateRecord captures the security-gate inputs for one deploy decision.
type GateRecord struct {
	ImageRef         string
	Digest           runtime.ImageDigest
	SignatureChecked bool
	ScanChecked      bool
}

// MockSecureDeployer is a pure test-owned recorder: the SignatureChecked /
// ScanChecked flags it stores are set by the test itself in the gate-logic
// phases, not by any production code. Asserting on it proves the test exercised
// the gate logic in order; it does NOT prove the real deploy path wires the
// gates. The "built-but-dormant" failure mode the daemon-todo warns about lives
// in deployLocally/deployReplica, which this harness never invokes — catching
// that regression needs a real SecureContainerConfig deploy assertion (tracked
// in plan/06-roadmap/daemon-todo.md as an OPS-04 / R11 follow-up).
type MockSecureDeployer struct {
	Last      GateRecord
	DeployCnt int
}

// NewMockSecureDeployer returns an empty recorder.
func NewMockSecureDeployer() *MockSecureDeployer { return &MockSecureDeployer{} }

// Deploy records the gate inputs and "succeeds".
func (d *MockSecureDeployer) Deploy(rec GateRecord) error {
	d.Last = rec
	d.DeployCnt++
	return nil
}

// AssertGatesInvoked fails the test if either gate was not recorded as checked.
// "Checked" here means the test's own gate-logic phases ran and recorded their
// result — see the MockSecureDeployer doc for why this does not assert that the
// production deploy path invokes the gates.
func (d *MockSecureDeployer) AssertGatesInvoked(t *testing.T) {
	t.Helper()
	if d.DeployCnt == 0 {
		t.Fatalf("MockSecureDeployer.Deploy was never called: security gate is dormant")
	}
	if !d.Last.SignatureChecked {
		t.Errorf("signature gate not invoked on deploy of %q", d.Last.ImageRef)
	}
	if !d.Last.ScanChecked {
		t.Errorf("scan gate not invoked on deploy of %q", d.Last.ImageRef)
	}
}

// -----------------------------------------------------------------------------
// LoopbackIngress — httptest.Server + ingress.Resolver.
//
// [MOCK: tunnel=loopback HTTP, TLS=httptest self-signed (TLS leg covered via
// httptest.NewTLSServer in TestGoldenPath_LoopbackIngress200), yamux=none].
//
// This exercises the only ingress code that can run without the full P2P stack:
// the 5-step resolver pipeline in internal/ingress/resolver.go. The "tunnel" is
// a direct in-process HTTP call to the local httptest.Server — no real reverse
// tunnel, no yamux session, no remote provider. Real tunnel coverage lives in
// internal/tunnel unit tests and the OPS-04 Linux containerd job.
// -----------------------------------------------------------------------------

// goldenPathBody is the fixed body the loopback origin returns; the test asserts
// on it to prove the request reached the backing httptest.Server.
const goldenPathBody = "golden-path-ok"

// LoopbackIngress is a local HTTP origin plus a resolver seeded with one entry.
type LoopbackIngress struct {
	Server    *httptest.Server
	TLSServer *httptest.Server
	Resolver  *ingress.Resolver
	Gossip    *loopbackGossipState
}

// loopbackGossipState implements ingress.GossipReader, returning a single
// pre-seeded ServiceEntry. LastSeen is refreshed on every read so the resolver's
// 5-minute freshness window never expires mid-test.
type loopbackGossipState struct {
	entry *ingress.ServiceEntry
}

// GetExposedServices implements ingress.GossipReader.
func (g *loopbackGossipState) GetExposedServices() map[string]*ingress.ServiceEntry {
	if g.entry == nil {
		return map[string]*ingress.ServiceEntry{}
	}
	// Refresh freshness so resolveByPrefix's `time.Since(LastSeen) < 5m` holds.
	g.entry.LastSeen = time.Now()
	key := "expose:" + g.entry.DeploymentID + ":" + itoa(g.entry.ContainerPort)
	return map[string]*ingress.ServiceEntry{key: g.entry}
}

// NewLoopbackIngress starts a plain + TLS httptest origin and a resolver.
func NewLoopbackIngress(t *testing.T) *LoopbackIngress {
	t.Helper()

	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(goldenPathBody))
	})

	srv := httptest.NewServer(handler)
	tlsSrv := httptest.NewTLSServer(handler)

	gossip := &loopbackGossipState{}
	resolver := ingress.NewResolver(gossip, nil)

	return &LoopbackIngress{
		Server:    srv,
		TLSServer: tlsSrv,
		Resolver:  resolver,
		Gossip:    gossip,
	}
}

// Seed registers a deployment as an exposed service pointing at the loopback
// origin. It seeds both the gossip state (for refresh-driven prefix resolution)
// and the resolver's local cache (via Register) so exact-match also works.
func (li *LoopbackIngress) Seed(deploymentID string) *ingress.ServiceEntry {
	entry := &ingress.ServiceEntry{
		DeploymentID:   deploymentID,
		ProviderNodeID: "golden-node",
		ProviderAddr:   li.Server.Listener.Addr().String(),
		ContainerPort:  80,
		HostPort:       8080,
		LastSeen:       time.Now(),
		RuntimeType:    "container",
	}
	li.Gossip.entry = entry
	li.Resolver.Register(entry) // sets LastSeen = now
	return entry
}

// OriginURL returns the plain-HTTP origin URL (the loopback "tunnel" target).
func (li *LoopbackIngress) OriginURL() string { return li.Server.URL }

// TLSOriginURL returns the HTTPS origin URL (httptest self-signed cert).
func (li *LoopbackIngress) TLSOriginURL() string { return li.TLSServer.URL }

// TLSClient returns an http.Client that trusts the httptest TLS server cert.
func (li *LoopbackIngress) TLSClient() *http.Client { return li.TLSServer.Client() }

// Close shuts down both httptest servers. Safe to call multiple times.
func (li *LoopbackIngress) Close() {
	if li.Server != nil {
		li.Server.Close()
	}
	if li.TLSServer != nil {
		li.TLSServer.Close()
	}
}

// -----------------------------------------------------------------------------
// small helpers
// -----------------------------------------------------------------------------

// itoa is a tiny int->string helper to avoid importing strconv in one spot.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
