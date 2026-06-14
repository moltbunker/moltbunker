package edge

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"net"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/moltbunker/moltbunker/pkg/types"
)

// nodeIDFromByte builds a deterministic NodeID for tests.
func nodeIDFromByte(b byte) types.NodeID {
	var id types.NodeID
	for i := range id {
		id[i] = b
	}
	return id
}

func TestConfigEdgeTierChecker_Authorized(t *testing.T) {
	id := nodeIDFromByte(0x01)
	c := NewConfigEdgeTierChecker([]string{id.String()})
	ok, err := c.IsEdgeAuthorized(context.Background(), id, common.Address{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected node in allowlist to be authorized")
	}
}

func TestConfigEdgeTierChecker_NotInList(t *testing.T) {
	c := NewConfigEdgeTierChecker([]string{nodeIDFromByte(0x01).String()})
	ok, err := c.IsEdgeAuthorized(context.Background(), nodeIDFromByte(0x02), common.Address{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatal("expected node not in allowlist to be rejected")
	}
}

func TestConfigEdgeTierChecker_EmptyAllowlistRejects(t *testing.T) {
	c := NewConfigEdgeTierChecker(nil)
	ok, _ := c.IsEdgeAuthorized(context.Background(), nodeIDFromByte(0x03), common.Address{})
	if ok {
		t.Fatal("empty allowlist must authorize nothing")
	}
}

// fakeRegistryReader implements EdgeRegistryReader for the on-chain checker test.
type fakeRegistryReader struct {
	active map[common.Address]bool
	err    error
}

func (f fakeRegistryReader) IsActiveEdgeProvider(_ context.Context, addr common.Address) (bool, error) {
	if f.err != nil {
		return false, f.err
	}
	return f.active[addr], nil
}

func TestOnChainEdgeTierChecker_Active(t *testing.T) {
	addr := common.HexToAddress("0x1111111111111111111111111111111111111111")
	c := NewOnChainEdgeTierChecker(fakeRegistryReader{active: map[common.Address]bool{addr: true}})
	ok, err := c.IsEdgeAuthorized(context.Background(), types.NodeID{}, addr)
	if err != nil || !ok {
		t.Fatalf("expected active provider authorized, got ok=%v err=%v", ok, err)
	}
}

func TestOnChainEdgeTierChecker_NilReaderFailsClosed(t *testing.T) {
	c := NewOnChainEdgeTierChecker(nil)
	ok, err := c.IsEdgeAuthorized(context.Background(), types.NodeID{}, common.Address{})
	if err != nil {
		t.Fatalf("nil reader should not error, got %v", err)
	}
	if ok {
		t.Fatal("nil reader must fail closed (not authorized)")
	}
}

func TestNewEdgeTierChecker_DefaultsToConfig(t *testing.T) {
	id := nodeIDFromByte(0x09)
	chk := NewEdgeTierChecker(Config{AllowedNodeIDs: []string{id.String()}}, nil)
	if _, ok := chk.(*ConfigEdgeTierChecker); !ok {
		t.Fatalf("expected ConfigEdgeTierChecker, got %T", chk)
	}
	ok, _ := chk.IsEdgeAuthorized(context.Background(), id, common.Address{})
	if !ok {
		t.Fatal("config checker should authorize allowlisted node")
	}
}

func TestNewEdgeTierChecker_OnChainSelected(t *testing.T) {
	chk := NewEdgeTierChecker(Config{Mode: ModeOnChain}, fakeRegistryReader{})
	if _, ok := chk.(*OnChainEdgeTierChecker); !ok {
		t.Fatalf("expected OnChainEdgeTierChecker, got %T", chk)
	}
}

func TestNewEdgeTierChecker_OnChainNilReaderFallsBackToConfig(t *testing.T) {
	chk := NewEdgeTierChecker(Config{Mode: ModeOnChain}, nil)
	if _, ok := chk.(*ConfigEdgeTierChecker); !ok {
		t.Fatalf("expected fallback to ConfigEdgeTierChecker, got %T", chk)
	}
}

func TestEdgeRegistry_RegisterAndList(t *testing.T) {
	r := NewEdgeRegistry()
	r.Register(EdgeNodeInfo{NodeID: nodeIDFromByte(0x01), IngressAddr: "10.0.0.1", TunnelPort: 9443})
	r.Register(EdgeNodeInfo{NodeID: nodeIDFromByte(0x02), IngressAddr: "10.0.0.2", TunnelPort: 9443})
	if got := len(r.ListHealthy()); got != 2 {
		t.Fatalf("ListHealthy = %d, want 2", got)
	}
	info, ok := r.ByNodeID(nodeIDFromByte(0x01))
	if !ok {
		t.Fatal("expected node 0x01 present")
	}
	if info.FullAddr() != "10.0.0.1:9443" {
		t.Fatalf("FullAddr = %q, want 10.0.0.1:9443", info.FullAddr())
	}
}

func TestEdgeRegistry_UpdateHealth(t *testing.T) {
	r := NewEdgeRegistry()
	id := nodeIDFromByte(0x05)
	r.Register(EdgeNodeInfo{NodeID: id, IngressAddr: "10.0.0.5", TunnelPort: 9443})
	r.UpdateHealth(id, false)
	if got := len(r.ListHealthy()); got != 0 {
		t.Fatalf("ListHealthy after unhealthy = %d, want 0", got)
	}
	if got := len(r.ListAll()); got != 1 {
		t.Fatalf("ListAll = %d, want 1", got)
	}
}

func TestEdgeRegistry_Unregister(t *testing.T) {
	r := NewEdgeRegistry()
	id := nodeIDFromByte(0x07)
	r.Register(EdgeNodeInfo{NodeID: id, IngressAddr: "10.0.0.7", TunnelPort: 9443})
	r.Unregister(id)
	if _, ok := r.ByNodeID(id); ok {
		t.Fatal("expected node absent after Unregister")
	}
}

func TestEdgeRegistry_RegisterEdgeSeam(t *testing.T) {
	r := NewEdgeRegistry()
	id := nodeIDFromByte(0x08)
	r.RegisterEdge(id, "0xabc", "edge.example.com", 9443, 1000)
	info, ok := r.ByNodeID(id)
	if !ok {
		t.Fatal("expected node registered via RegisterEdge")
	}
	if info.IngressAddr != "edge.example.com" || info.TunnelPort != 9443 {
		t.Fatalf("unexpected info: %+v", info)
	}
}

func TestEdgeSelector_SelectsHealthy(t *testing.T) {
	r := NewEdgeRegistry()
	healthy := nodeIDFromByte(0x01)
	unhealthy := nodeIDFromByte(0x02)
	r.Register(EdgeNodeInfo{NodeID: healthy, IngressAddr: "10.0.0.1", TunnelPort: 9443})
	r.Register(EdgeNodeInfo{NodeID: unhealthy, IngressAddr: "10.0.0.2", TunnelPort: 9443})
	r.UpdateHealth(unhealthy, false)

	s := NewEdgeSelector(r)
	got, err := s.SelectEdge(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.NodeID != healthy {
		t.Fatalf("selected %v, want healthy node", got.NodeID.String()[:8])
	}
}

func TestEdgeSelector_LeastConnRoundRobin(t *testing.T) {
	r := NewEdgeRegistry()
	a := nodeIDFromByte(0x01)
	b := nodeIDFromByte(0x02)
	r.Register(EdgeNodeInfo{NodeID: a, IngressAddr: "10.0.0.1", TunnelPort: 9443})
	r.Register(EdgeNodeInfo{NodeID: b, IngressAddr: "10.0.0.2", TunnelPort: 9443})

	s := NewEdgeSelector(r)
	// Two selections with equal (zero) connection counts must hit both nodes.
	seen := map[types.NodeID]int{}
	for i := 0; i < 2; i++ {
		got, err := s.SelectEdge(context.Background())
		if err != nil {
			t.Fatalf("select %d: %v", i, err)
		}
		seen[got.NodeID]++
	}
	if seen[a] != 1 || seen[b] != 1 {
		t.Fatalf("expected round-robin across both nodes, got %v", seen)
	}
	// Both now have 1 in-flight; releasing a makes it least-loaded → next pick.
	s.Release(a)
	got, err := s.SelectEdge(context.Background())
	if err != nil {
		t.Fatalf("post-release select: %v", err)
	}
	if got.NodeID != a {
		t.Fatalf("expected least-loaded node a after release, got %v", got.NodeID.String()[:8])
	}
}

func TestEdgeSelector_NoHealthyNodes(t *testing.T) {
	r := NewEdgeRegistry()
	s := NewEdgeSelector(r)
	if _, err := s.SelectEdge(context.Background()); err == nil {
		t.Fatal("expected error when no healthy edge nodes")
	}
}

// --- EdgeProbe tests ------------------------------------------------------

func selfSignedTLSConfig(t *testing.T) (*tls.Config, *tls.Config) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "edge-test"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
		DNSNames:     []string{"localhost"},
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	leaf, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatal(err)
	}
	serverCfg := &tls.Config{
		Certificates: []tls.Certificate{{Certificate: [][]byte{der}, PrivateKey: key, Leaf: leaf}},
		MinVersion:   tls.VersionTLS12,
	}
	// Client trusts the self-signed cert (avoids InsecureSkipVerify in tests).
	pool := x509.NewCertPool()
	pool.AddCert(leaf)
	clientCfg := &tls.Config{RootCAs: pool, MinVersion: tls.VersionTLS12}
	return serverCfg, clientCfg
}

func TestEdgeProbe_MarksHealthy(t *testing.T) {
	serverCfg, clientCfg := selfSignedTLSConfig(t)
	lis, err := tls.Listen("tcp", "127.0.0.1:0", serverCfg)
	if err != nil {
		t.Fatal(err)
	}
	defer lis.Close()
	go func() {
		for {
			c, aerr := lis.Accept()
			if aerr != nil {
				return
			}
			_ = c.(*tls.Conn).Handshake()
			_ = c.Close()
		}
	}()

	port := lis.Addr().(*net.TCPAddr).Port
	r := NewEdgeRegistry()
	id := nodeIDFromByte(0x01)
	r.Register(EdgeNodeInfo{NodeID: id, IngressAddr: "127.0.0.1", TunnelPort: port})
	// Force unhealthy first so the probe must flip it back.
	r.UpdateHealth(id, false)

	p := NewEdgeProbe(r, clientCfg, 20*time.Millisecond, time.Second)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	p.Start(ctx)
	defer p.Stop()

	if !waitFor(2*time.Second, func() bool {
		info, ok := r.ByNodeID(id)
		return ok && info.Healthy
	}) {
		t.Fatal("probe did not mark node healthy")
	}
}

func TestEdgeProbe_MarksUnhealthyAfterFailures(t *testing.T) {
	_, clientCfg := selfSignedTLSConfig(t)
	// Bind+close a listener to obtain a port that is then unreachable.
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := lis.Addr().(*net.TCPAddr).Port
	_ = lis.Close()

	r := NewEdgeRegistry()
	id := nodeIDFromByte(0x01)
	r.Register(EdgeNodeInfo{NodeID: id, IngressAddr: "127.0.0.1", TunnelPort: port}) // starts healthy

	p := NewEdgeProbe(r, clientCfg, 15*time.Millisecond, 200*time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	p.Start(ctx)
	defer p.Stop()

	if !waitFor(3*time.Second, func() bool {
		info, ok := r.ByNodeID(id)
		return ok && !info.Healthy
	}) {
		t.Fatal("probe did not mark unreachable node unhealthy after failures")
	}
}

func TestMockEdgeTierChecker(t *testing.T) {
	m := &MockEdgeTierChecker{Authorized: true}
	if ok, _ := m.IsEdgeAuthorized(context.Background(), types.NodeID{}, common.Address{}); !ok {
		t.Fatal("mock should return configured answer")
	}
}

// waitFor polls cond until true or timeout. Returns whether cond became true.
func waitFor(timeout time.Duration, cond func() bool) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return cond()
}
