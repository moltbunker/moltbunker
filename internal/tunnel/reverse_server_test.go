package tunnel

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"io"
	"net"
	"testing"
	"time"

	"github.com/hashicorp/yamux"
	"github.com/moltbunker/moltbunker/pkg/types"
)

// edgeProbeProvider dials the reverse server, establishes the provider-side
// yamux session, sends a register request, and returns the first control
// message type the server writes back. It is a minimal hand-rolled provider so
// the test can inject EdgeCapabilities (which the production ReverseClient does
// not yet send).
func edgeProbeProvider(t *testing.T, serverAddr string, clientCfg *tls.Config, req TunnelRegisterRequest) (byte, []byte) {
	t.Helper()
	conn, err := tls.Dial("tcp", serverAddr, clientCfg)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	yc := yamux.DefaultConfig()
	yc.LogOutput = io.Discard
	session, err := yamux.Server(conn, yc)
	if err != nil {
		t.Fatalf("yamux server: %v", err)
	}
	defer session.Close()

	// The ingress opens the control stream; accept it.
	ctrl, err := session.Accept()
	if err != nil {
		t.Fatalf("accept control stream: %v", err)
	}
	defer ctrl.Close()
	_ = ctrl.SetDeadline(time.Now().Add(5 * time.Second))

	payload, _ := json.Marshal(req)
	if err := writeControlMsg(ctrl, MsgTunnelRegister, payload); err != nil {
		t.Fatalf("write register: %v", err)
	}

	msgType, respPayload, err := readControlMsg(ctrl)
	if err != nil {
		t.Fatalf("read response: %v", err)
	}
	return msgType, respPayload
}

// startEdgeTestServer spins up a ReverseServer with the given options and
// returns its address and a client TLS config trusting it.
func startEdgeTestServer(t *testing.T, opts ...ReverseServerOption) string {
	t.Helper()
	ingressCert := generateTestCert(t)
	providerCert := generateTestCert(t)

	pool := x509.NewCertPool()
	pool.AddCert(ingressCert.Leaf)
	pool.AddCert(providerCert.Leaf)

	ingressTLSCfg := &tls.Config{
		Certificates: []tls.Certificate{ingressCert},
		ClientCAs:    pool,
		ClientAuth:   tls.RequireAnyClientCert,
		MinVersion:   tls.VersionTLS13,
	}
	lis, err := tls.Listen("tcp", "127.0.0.1:0", ingressTLSCfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = lis.Close() })

	srv := NewReverseServer(lis, opts...)
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	go func() { _ = srv.Serve(ctx) }()

	// Stash the provider cert on the test via a closure-captured client config.
	t.Cleanup(func() {})
	clientCfgHolder[t.Name()] = &tls.Config{
		Certificates:       []tls.Certificate{providerCert},
		RootCAs:            pool,
		InsecureSkipVerify: true, // test only: self-signed
		MinVersion:         tls.VersionTLS13,
	}
	return lis.Addr().String()
}

// clientCfgHolder passes the per-test provider TLS config from the server
// helper to the test body without changing the helper's return signature.
var clientCfgHolder = map[string]*tls.Config{}

func baseEdgeRegisterReq() TunnelRegisterRequest {
	nonce := make([]byte, 32)
	for i := range nonce {
		nonce[i] = byte(i)
	}
	return TunnelRegisterRequest{
		NodeID:        "edge-node",
		DeploymentID:  "dep-edge",
		ContainerPort: 8080,
		Nonce:         hex.EncodeToString(nonce),
		Timestamp:     time.Now().Unix(),
		EdgeCapabilities: &EdgeCapabilities{
			IngressAddr: "edge.example.com",
			TunnelPort:  9443,
			MaxStreams:  1000,
		},
		WalletProof: &WalletProof{Address: "0x1111111111111111111111111111111111111111"},
	}
}

func TestReverseServer_EdgeGate_Rejected(t *testing.T) {
	checker := func(_, _ string) (bool, error) { return false, nil }
	addr := startEdgeTestServer(t, WithEdgeTierChecker(checker))
	clientCfg := clientCfgHolder[t.Name()]

	msgType, payload := edgeProbeProvider(t, addr, clientCfg, baseEdgeRegisterReq())
	if msgType != MsgEdgeRoleRejected {
		t.Fatalf("msgType = 0x%02x, want MsgEdgeRoleRejected (0x%02x); payload=%q", msgType, MsgEdgeRoleRejected, payload)
	}
}

func TestReverseServer_EdgeGate_Authorized(t *testing.T) {
	// Authorized edge passes the gate, then proceeds to registration validation.
	// The hand-rolled request does not carry a valid TLS-bound nonce, so it is
	// rejected at validation with MsgTunnelError — NOT MsgEdgeRoleRejected. That
	// distinction proves the gate let it through.
	checker := func(_, _ string) (bool, error) { return true, nil }
	reg := newEdgeRegistrarSpy()
	addr := startEdgeTestServer(t, WithEdgeTierChecker(checker), WithEdgeRegistry(reg))
	clientCfg := clientCfgHolder[t.Name()]

	msgType, _ := edgeProbeProvider(t, addr, clientCfg, baseEdgeRegisterReq())
	if msgType == MsgEdgeRoleRejected {
		t.Fatal("authorized edge must not be rejected by the edge gate")
	}
	if msgType != MsgTunnelError && msgType != MsgTunnelRegistered {
		t.Fatalf("unexpected msgType 0x%02x", msgType)
	}
	// When registration completes, the edge node is recorded in the registry.
	if msgType == MsgTunnelRegistered && reg.calls == 0 {
		t.Fatal("edge node was not recorded in the registry on successful registration")
	}
}

func TestReverseServer_Blocklist_OpenStreamRejected(t *testing.T) {
	bl := NewBlocklist()
	bl.Block("evil", "takedown")
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer lis.Close()
	srv := NewReverseServer(lis, WithBlocklist(bl))

	if _, err := srv.OpenStream("evil"); err == nil {
		t.Fatal("expected blocked subdomain to be refused at OpenStream")
	}
	// A non-blocked subdomain falls through to the normal not-registered error.
	if _, err := srv.OpenStream("clean"); err == nil {
		t.Fatal("expected not-found error for unregistered subdomain")
	}
}

// edgeRegistrarSpy records RegisterEdge calls.
type edgeRegistrarSpy struct {
	calls int
}

func newEdgeRegistrarSpy() *edgeRegistrarSpy { return &edgeRegistrarSpy{} }

func (s *edgeRegistrarSpy) RegisterEdge(_ types.NodeID, _, _ string, _, _ int) {
	s.calls++
}
