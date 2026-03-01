package p2p

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/moltbunker/moltbunker/internal/identity"
	"github.com/moltbunker/moltbunker/internal/security"
	"github.com/moltbunker/moltbunker/pkg/types"
)

// createTestTransportWithNodeID builds a Transport and returns its TLS-derived NodeID
// (SHA256 of cert's RawSubjectPublicKeyInfo), which matches what transport.DialContext verifies.
func createTestTransportWithNodeID(t testing.TB, name string) (*Transport, types.NodeID) {
	t.Helper()
	dir := t.TempDir()
	km, err := identity.NewKeyManager(filepath.Join(dir, name+".key"))
	if err != nil {
		t.Fatalf("key manager %s: %v", name, err)
	}
	cm, err := identity.NewCertificateManager(km)
	if err != nil {
		t.Fatalf("cert manager %s: %v", name, err)
	}
	tr, err := NewTransport(cm, security.NewCertPinStore())
	if err != nil {
		t.Fatalf("transport %s: %v", name, err)
	}
	tr.SetDialTimeout(5 * time.Second)

	// NodeID = SHA256(RawSubjectPublicKeyInfo) — same as transport.go:215
	hash := sha256.Sum256(cm.Certificate().RawSubjectPublicKeyInfo)
	var nodeID types.NodeID
	copy(nodeID[:], hash[:])
	return tr, nodeID
}

// acceptAndHandshake runs an accept loop that completes the TLS handshake
// before closing connections. Stops when listener is closed.
func acceptAndHandshake(listener net.Listener, counter *atomic.Int64) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		go func(c net.Conn) {
			defer c.Close()
			if tlsConn, ok := c.(*tls.Conn); ok {
				_ = tlsConn.Handshake()
			}
			if counter != nil {
				counter.Add(1)
			}
		}(conn)
	}
}

// TestTLS_MutualAuthConcurrent verifies TLS 1.3 mutual authentication works
// correctly under concurrent load. 50 unique clients connect simultaneously
// to one server, each performing a full TLS handshake with NodeID verification.
func TestTLS_MutualAuthConcurrent(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("skip TLS load test in CI")
	}

	const numClients = 50

	serverTransport, serverNodeID := createTestTransportWithNodeID(t, "server")

	listener, err := serverTransport.Listen("127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer listener.Close()
	addr := listener.Addr().String()

	var serverAccepted atomic.Int64
	go acceptAndHandshake(listener, &serverAccepted)

	var wg sync.WaitGroup
	var clientSuccess atomic.Int64
	var clientErrors atomic.Int64
	latencies := make([]time.Duration, numClients)

	for i := 0; i < numClients; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			clientTransport, _ := createTestTransportWithNodeID(t, fmt.Sprintf("client-%d", idx))

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			start := time.Now()
			conn, err := clientTransport.DialContext(ctx, serverNodeID, addr)
			latencies[idx] = time.Since(start)

			if err != nil {
				clientErrors.Add(1)
				return
			}
			conn.Close()
			clientSuccess.Add(1)
		}(i)
	}

	wg.Wait()

	successCount := clientSuccess.Load()
	t.Logf("Results: %d/%d success, %d errors, server accepted %d",
		successCount, numClients, clientErrors.Load(), serverAccepted.Load())

	var totalLatency time.Duration
	var minLatency, maxLatency time.Duration
	first := true
	for _, lat := range latencies {
		if lat == 0 {
			continue
		}
		totalLatency += lat
		if first || lat < minLatency {
			minLatency = lat
		}
		if lat > maxLatency {
			maxLatency = lat
		}
		first = false
	}
	if successCount > 0 {
		t.Logf("Latency — min: %v, avg: %v, max: %v",
			minLatency, totalLatency/time.Duration(successCount), maxLatency)
	}

	minRequired := int64(numClients * 90 / 100)
	if successCount < minRequired {
		t.Errorf("too many failures: %d/%d succeeded (need %d)", successCount, numClients, minRequired)
	}
}

// TestTLS_NodeIDMismatchRejection verifies that connecting with the wrong
// expected NodeID is correctly rejected under load.
func TestTLS_NodeIDMismatchRejection(t *testing.T) {
	serverTransport, _ := createTestTransportWithNodeID(t, "server")
	_, wrongNodeID := createTestTransportWithNodeID(t, "wrong-server")

	listener, err := serverTransport.Listen("127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer listener.Close()
	addr := listener.Addr().String()

	go acceptAndHandshake(listener, nil)

	const numAttempts = 20
	var wg sync.WaitGroup
	var rejected atomic.Int64

	for i := 0; i < numAttempts; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			clientTransport, _ := createTestTransportWithNodeID(t, fmt.Sprintf("mismatch-client-%d", idx))
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			conn, err := clientTransport.DialContext(ctx, wrongNodeID, addr)
			if err != nil {
				rejected.Add(1)
				return
			}
			conn.Close()
		}(i)
	}

	wg.Wait()

	rejectedCount := rejected.Load()
	t.Logf("NodeID mismatch: %d/%d correctly rejected", rejectedCount, numAttempts)

	if rejectedCount != int64(numAttempts) {
		t.Errorf("expected all %d rejected for wrong NodeID, only %d rejected",
			numAttempts, rejectedCount)
	}
}

// TestTLS_CertPinningUnderLoad verifies TOFU cert pinning is consistent
// when multiple sequential connections are made to the same server.
func TestTLS_CertPinningUnderLoad(t *testing.T) {
	serverTransport, serverNodeID := createTestTransportWithNodeID(t, "pinning-server")

	listener, err := serverTransport.Listen("127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer listener.Close()
	addr := listener.Addr().String()

	go acceptAndHandshake(listener, nil)

	clientTransport, _ := createTestTransportWithNodeID(t, "pinning-client")

	const numConnections = 30
	successCount := 0

	for i := 0; i < numConnections; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		conn, err := clientTransport.DialContext(ctx, serverNodeID, addr)
		cancel()
		if err != nil {
			t.Logf("connection %d failed: %v", i, err)
			continue
		}
		conn.Close()
		successCount++
	}

	t.Logf("Cert pinning: %d/%d connections succeeded", successCount, numConnections)

	if successCount < numConnections {
		t.Errorf("pinning failures: only %d/%d succeeded", successCount, numConnections)
	}
}

// BenchmarkTLS_Handshake measures TLS 1.3 mutual auth handshake latency.
func BenchmarkTLS_Handshake(b *testing.B) {
	serverTransport, serverNodeID := createTestTransportWithNodeID(b, "bench-server")

	listener, err := serverTransport.Listen("127.0.0.1:0")
	if err != nil {
		b.Fatal(err)
	}
	defer listener.Close()
	addr := listener.Addr().String()

	go acceptAndHandshake(listener, nil)

	clientTransport, _ := createTestTransportWithNodeID(b, "bench-client")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		conn, err := clientTransport.DialContext(ctx, serverNodeID, addr)
		cancel()
		if err != nil {
			b.Fatalf("dial failed at iteration %d: %v", i, err)
		}
		conn.Close()
	}
}
