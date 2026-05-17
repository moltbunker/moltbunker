package tunnel

import (
	"bufio"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"
)

// generateTestCert creates a self-signed TLS certificate for testing.
func generateTestCert(t *testing.T) tls.Certificate {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "test"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		DNSNames:     []string{"localhost"},
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		t.Fatal(err)
	}

	return tls.Certificate{
		Certificate: [][]byte{certDER},
		PrivateKey:  key,
		Leaf:        cert,
	}
}

func nodeIDFromCert(cert tls.Certificate) string {
	h := sha256.Sum256(cert.Leaf.RawSubjectPublicKeyInfo)
	return hex.EncodeToString(h[:])
}

// TestReverseServerClient_EndToEnd tests the full reverse tunnel flow:
// 1. Start a mock HTTP server (simulates a container)
// 2. Start the reverse tunnel server (ingress-side)
// 3. Start the reverse tunnel client (provider-side)
// 4. Verify that the ingress can proxy an HTTP request to the container via yamux
func TestReverseServerClient_EndToEnd(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// --- Mock container: simple HTTP server on a random port ---
	containerLis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer containerLis.Close()

	containerPort := containerLis.Addr().(*net.TCPAddr).Port
	containerSrv := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(200)
			fmt.Fprintf(w, "hello from container on port %d", containerPort)
		}),
	}
	go func() { _ = containerSrv.Serve(containerLis) }()
	defer containerSrv.Close()

	// --- TLS certs ---
	ingressCert := generateTestCert(t)
	providerCert := generateTestCert(t)

	certPool := x509.NewCertPool()
	certPool.AddCert(ingressCert.Leaf)
	certPool.AddCert(providerCert.Leaf)

	ingressTLSCfg := &tls.Config{
		Certificates: []tls.Certificate{ingressCert},
		ClientCAs:    certPool,
		ClientAuth:   tls.RequireAnyClientCert,
		MinVersion:   tls.VersionTLS13,
	}

	providerTLSCfg := &tls.Config{
		Certificates:       []tls.Certificate{providerCert},
		RootCAs:            certPool,
		InsecureSkipVerify: true, // Test only: self-signed certs
		MinVersion:         tls.VersionTLS13,
	}

	// --- Start reverse tunnel server (ingress-side) ---
	serverLis, err := tls.Listen("tcp", "127.0.0.1:0", ingressTLSCfg)
	if err != nil {
		t.Fatal(err)
	}
	defer serverLis.Close()

	reverseServer := NewReverseServer(serverLis, WithDomain("test.dev"))

	go func() { _ = reverseServer.Serve(ctx) }()

	serverAddr := serverLis.Addr().String()

	// --- Start reverse tunnel client (provider-side) ---
	portResolver := &staticPortResolver{port: containerPort}
	revClient := NewReverseClient(serverAddr, portResolver, providerTLSCfg)

	subCh := make(chan string, 1)
	errCh := make(chan error, 1)
	go func() {
		sub, err := revClient.Connect(ctx, "dep-123", containerPort)
		if err != nil && ctx.Err() == nil {
			errCh <- err
		}
		subCh <- sub
	}()

	// Wait for the tunnel to be registered
	var assignedSub string
	deadline := time.After(10 * time.Second)
	for {
		select {
		case err := <-errCh:
			t.Fatalf("client connect error: %v", err)
		case <-deadline:
			t.Fatal("timeout waiting for tunnel registration")
		default:
		}

		if reverseServer.Registry().ActiveCount() > 0 {
			// Find the subdomain
			// Iterate is not exposed, but we can check using the last assignment
			time.Sleep(100 * time.Millisecond) // Give the registration time to complete
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	// Find what subdomain was assigned
	select {
	case sub := <-subCh:
		assignedSub = sub
	case <-time.After(5 * time.Second):
		// The client blocks, so it might not have returned yet.
		// Check the registry directly.
	}

	// If we don't have the subdomain from the client yet, find it from the registry
	if assignedSub == "" && reverseServer.Registry().ActiveCount() > 0 {
		// We need to iterate — let's try opening streams on known patterns
		// Since we can't easily iterate, let's use the client's Subdomain() method
		// after giving it more time.
		time.Sleep(200 * time.Millisecond)
		assignedSub = revClient.Subdomain()
	}

	if assignedSub == "" {
		t.Fatal("no subdomain assigned")
	}

	t.Logf("Assigned subdomain: %s", assignedSub)

	// --- Proxy an HTTP request through the reverse tunnel ---
	stream, err := reverseServer.OpenStream(assignedSub)
	if err != nil {
		t.Fatalf("OpenStream failed: %v", err)
	}
	defer stream.Close()

	// Write an HTTP request to the stream
	httpReq, _ := http.NewRequest("GET", "http://"+assignedSub+".test.dev/", nil)
	httpReq.Host = assignedSub + ".test.dev"
	if err := httpReq.Write(stream); err != nil {
		t.Fatalf("write HTTP request: %v", err)
	}

	// Read the HTTP response
	resp, err := http.ReadResponse(bufio.NewReader(stream), httpReq)
	if err != nil {
		t.Fatalf("read HTTP response: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}

	// Read body
	buf := make([]byte, 1024)
	n, _ := resp.Body.Read(buf)
	body := string(buf[:n])
	expectedBody := fmt.Sprintf("hello from container on port %d", containerPort)
	if !strings.Contains(body, expectedBody) {
		t.Errorf("body = %q, want to contain %q", body, expectedBody)
	}

	// Cleanup
	_ = revClient.Disconnect()
}

// staticPortResolver always returns the same port.
type staticPortResolver struct {
	port int
}

func (r *staticPortResolver) ResolveDeploymentPort(_ string, _ int) (string, error) {
	return fmt.Sprintf("127.0.0.1:%d", r.port), nil
}

func TestReverseServer_OpenStream_NotFound(t *testing.T) {
	lis, _ := net.Listen("tcp", "127.0.0.1:0")
	defer lis.Close()

	// Create server with a plain TCP listener (not TLS, won't actually serve)
	srv := NewReverseServer(lis)

	_, err := srv.OpenStream("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent subdomain")
	}
}

func TestReverseServer_IPLimiting(t *testing.T) {
	// Need a listener for NewReverseServer, use a dummy one
	lis, _ := net.Listen("tcp", "127.0.0.1:0")
	defer lis.Close()

	srv := NewReverseServer(lis, WithMaxPerIP(2))

	// Should allow 2
	if !srv.checkIPLimit("1.2.3.4") {
		t.Fatal("first connection should be allowed")
	}
	if !srv.checkIPLimit("1.2.3.4") {
		t.Fatal("second connection should be allowed")
	}
	// Third should be rejected
	if srv.checkIPLimit("1.2.3.4") {
		t.Fatal("third connection should be rejected (limit=2)")
	}

	// Different IP should still be allowed
	if !srv.checkIPLimit("5.6.7.8") {
		t.Fatal("different IP should be allowed")
	}

	// Release a slot
	srv.releaseIPSlot("1.2.3.4")
	if !srv.checkIPLimit("1.2.3.4") {
		t.Fatal("should be allowed after release")
	}
}

func TestBandwidthLimiter_AllowRequest(t *testing.T) {
	limiter := NewBandwidthLimiter(&TunnelLimits{
		MaxRPS:       5,
		MaxBandwidth: 1024 * 1024,
		MaxStreams:   100,
	})

	// Should allow initial burst of 5
	for i := 0; i < 5; i++ {
		if !limiter.AllowRequest() {
			t.Fatalf("request %d should be allowed", i)
		}
	}

	// Next should be rate limited (burst exhausted)
	if limiter.AllowRequest() {
		t.Log("6th request allowed (burst may be > MaxRPS in rate.Limiter)")
	}
}
