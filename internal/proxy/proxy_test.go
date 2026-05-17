package proxy

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

// --- Session Tracker tests ---

func TestSessionTracker_AddRemove(t *testing.T) {
	tracker := NewSessionTracker(2)

	s1 := &Session{ID: "s1", Wallet: "w1"}
	s2 := &Session{ID: "s2", Wallet: "w2"}
	s3 := &Session{ID: "s3", Wallet: "w3"}

	if !tracker.Add(s1) {
		t.Error("should add s1")
	}
	if !tracker.Add(s2) {
		t.Error("should add s2")
	}
	if tracker.Add(s3) {
		t.Error("should not add s3 (limit reached)")
	}
	if tracker.Count() != 2 {
		t.Errorf("count = %d, want 2", tracker.Count())
	}

	tracker.Remove("s1")
	if tracker.Count() != 1 {
		t.Errorf("count after remove = %d, want 1", tracker.Count())
	}

	if !tracker.Add(s3) {
		t.Error("should add s3 after removing s1")
	}
}

func TestSessionTracker_GetList(t *testing.T) {
	tracker := NewSessionTracker(10)

	s1 := &Session{ID: "s1", Wallet: "w1", Protocol: "socks5"}
	s2 := &Session{ID: "s2", Wallet: "w2", Protocol: "http_connect"}
	tracker.Add(s1)
	tracker.Add(s2)

	got, ok := tracker.Get("s1")
	if !ok || got.Protocol != "socks5" {
		t.Errorf("get s1: ok=%v, protocol=%q", ok, got.Protocol)
	}

	_, ok = tracker.Get("nonexistent")
	if ok {
		t.Error("get nonexistent should return false")
	}

	list := tracker.List()
	if len(list) != 2 {
		t.Errorf("list len = %d, want 2", len(list))
	}
}

// --- Bandwidth Meter tests ---

func TestBandwidthMeter(t *testing.T) {
	meter := NewBandwidthMeter()
	if meter.BytesRead() != 0 || meter.BytesWritten() != 0 {
		t.Error("new meter should have zero counts")
	}

	// Create a pipe to test MeteredConn
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	metered := meter.WrapReader(client)

	// Write from server, read through metered client
	go func() {
		_, _ = server.Write([]byte("hello"))
		server.Close()
	}()

	buf := make([]byte, 10)
	n, _ := metered.Read(buf)
	if n != 5 {
		t.Errorf("read %d bytes, want 5", n)
	}
	if meter.BytesRead() != 5 {
		t.Errorf("meter read = %d, want 5", meter.BytesRead())
	}
}

func TestMeteredConn_Write(t *testing.T) {
	meter := NewBandwidthMeter()
	server, client := net.Pipe()
	defer server.Close()

	metered := &MeteredConn{Conn: client, meter: meter}

	// Read from server in background
	go func() {
		_, _ = io.Copy(io.Discard, server)
	}()

	_, _ = metered.Write([]byte("test data"))
	metered.Close()

	if meter.BytesWritten() != 9 {
		t.Errorf("meter written = %d, want 9", meter.BytesWritten())
	}
}

// --- ACL tests ---

func TestACL_AllowAll(t *testing.T) {
	acl := NewACL(nil, nil)
	if !acl.IsAllowed("example.com") {
		t.Error("empty ACL should allow all")
	}
}

func TestACL_Blocklist(t *testing.T) {
	acl := NewACL(nil, []string{"evil.com", "bad.org"})

	if acl.IsAllowed("evil.com") {
		t.Error("evil.com should be blocked")
	}
	if acl.IsAllowed("sub.evil.com") {
		t.Error("sub.evil.com should be blocked (parent match)")
	}
	if !acl.IsAllowed("good.com") {
		t.Error("good.com should be allowed")
	}
}

func TestACL_Allowlist(t *testing.T) {
	acl := NewACL([]string{"allowed.com", "good.org"}, nil)

	if !acl.IsAllowed("allowed.com") {
		t.Error("allowed.com should be allowed")
	}
	if acl.IsAllowed("other.com") {
		t.Error("other.com should not be allowed (not in allowlist)")
	}
}

func TestACL_BlocklistOverridesAllowlist(t *testing.T) {
	acl := NewACL([]string{"example.com"}, []string{"example.com"})

	if acl.IsAllowed("example.com") {
		t.Error("blocklist should override allowlist")
	}
}

func TestACL_DynamicBlocklist(t *testing.T) {
	acl := NewACL(nil, nil)

	if !acl.IsAllowed("example.com") {
		t.Error("should be allowed initially")
	}

	acl.AddToBlocklist("example.com")
	if acl.IsAllowed("example.com") {
		t.Error("should be blocked after adding to blocklist")
	}

	acl.RemoveFromBlocklist("example.com")
	if !acl.IsAllowed("example.com") {
		t.Error("should be allowed after removing from blocklist")
	}
}

func TestACL_NilSafe(t *testing.T) {
	var acl *ACL
	if !acl.IsAllowed("anything.com") {
		t.Error("nil ACL should allow all")
	}
}

// --- SOCKS5 Protocol tests ---

// mockDialer records dial calls and returns a mock connection.
type mockDialer struct {
	mu      sync.Mutex
	calls   []string
	failFor map[string]error
}

func newMockDialer() *mockDialer {
	return &mockDialer{failFor: make(map[string]error)}
}

func (d *mockDialer) DialContext(_ context.Context, _, address string) (net.Conn, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.calls = append(d.calls, address)
	if err, ok := d.failFor[address]; ok {
		return nil, err
	}
	server, client := net.Pipe()
	go func() {
		_, _ = io.Copy(io.Discard, server)
		server.Close()
	}()
	return client, nil
}

func TestSOCKS5_Greeting_NoAuth(t *testing.T) {
	dialer := newMockDialer()
	auth := &AllowAllAuth{DefaultWallet: "test-wallet"}
	tracker := NewSessionTracker(10)
	srv := NewSOCKS5Server(dialer, auth, tracker, nil)

	// Create client/server pair
	clientConn, serverConn := net.Pipe()

	go func() {
		defer serverConn.Close()
		// Simulate: greeting → connect → close
		wallet, err := srv.handleGreeting(serverConn)
		if err != nil {
			t.Errorf("handleGreeting: %v", err)
			return
		}
		if wallet != "test-wallet" {
			t.Errorf("wallet = %q, want %q", wallet, "test-wallet")
		}
	}()

	// Send SOCKS5 greeting with no-auth method
	_, _ = clientConn.Write([]byte{socks5Version, 1, authNone})

	// Read server's method selection
	resp := make([]byte, 2)
	_, _ = io.ReadFull(clientConn, resp)
	if resp[0] != socks5Version || resp[1] != authNone {
		t.Errorf("method selection = %v, want [5, 0]", resp)
	}

	clientConn.Close()
}

func TestSOCKS5_Greeting_UserPassAuth(t *testing.T) {
	dialer := newMockDialer()
	auth := &AllowAllAuth{DefaultWallet: "auth-wallet"}
	tracker := NewSessionTracker(10)
	srv := NewSOCKS5Server(dialer, auth, tracker, nil)

	clientConn, serverConn := net.Pipe()

	done := make(chan string, 1)
	go func() {
		defer serverConn.Close()
		wallet, err := srv.handleGreeting(serverConn)
		if err != nil {
			t.Errorf("handleGreeting: %v", err)
			return
		}
		done <- wallet
	}()

	// Send greeting with username/password method
	_, _ = clientConn.Write([]byte{socks5Version, 1, authUsernamePasswd})

	// Read method selection
	resp := make([]byte, 2)
	_, _ = io.ReadFull(clientConn, resp)
	if resp[1] != authUsernamePasswd {
		t.Fatalf("expected username/password method, got %d", resp[1])
	}

	// Send username/password auth (RFC 1929)
	user := "testuser"
	pass := "testpass"
	authReq := []byte{0x01, byte(len(user))}
	authReq = append(authReq, []byte(user)...)
	authReq = append(authReq, byte(len(pass)))
	authReq = append(authReq, []byte(pass)...)
	_, _ = clientConn.Write(authReq)

	// Read auth response
	authResp := make([]byte, 2)
	_, _ = io.ReadFull(clientConn, authResp)
	if authResp[1] != 0x00 {
		t.Error("auth should succeed")
	}

	wallet := <-done
	if wallet != "auth-wallet" {
		t.Errorf("wallet = %q, want %q", wallet, "auth-wallet")
	}

	clientConn.Close()
}

func TestSOCKS5_ConnectRequest_Domain(t *testing.T) {
	dialer := newMockDialer()
	auth := &AllowAllAuth{DefaultWallet: "w1"}
	tracker := NewSessionTracker(10)
	srv := NewSOCKS5Server(dialer, auth, tracker, nil)

	clientConn, serverConn := net.Pipe()

	done := make(chan string, 1)
	go func() {
		defer serverConn.Close()
		target, err := srv.handleConnectRequest(serverConn)
		if err != nil {
			t.Errorf("handleConnectRequest: %v", err)
			return
		}
		done <- target
	}()

	// Send CONNECT request with domain address
	domain := "example.com"
	req := []byte{socks5Version, cmdConnect, 0x00, addrDomain, byte(len(domain))}
	req = append(req, []byte(domain)...)
	port := make([]byte, 2)
	binary.BigEndian.PutUint16(port, 443)
	req = append(req, port...)
	_, _ = clientConn.Write(req)

	target := <-done
	if target != "example.com:443" {
		t.Errorf("target = %q, want %q", target, "example.com:443")
	}

	clientConn.Close()
}

func TestSOCKS5_ConnectRequest_IPv4(t *testing.T) {
	dialer := newMockDialer()
	auth := &AllowAllAuth{DefaultWallet: "w1"}
	tracker := NewSessionTracker(10)
	srv := NewSOCKS5Server(dialer, auth, tracker, nil)

	clientConn, serverConn := net.Pipe()

	done := make(chan string, 1)
	go func() {
		defer serverConn.Close()
		target, err := srv.handleConnectRequest(serverConn)
		if err != nil {
			t.Errorf("handleConnectRequest: %v", err)
			return
		}
		done <- target
	}()

	// Send CONNECT request with IPv4 address 93.184.216.34:80
	req := []byte{socks5Version, cmdConnect, 0x00, addrIPv4, 93, 184, 216, 34}
	port := make([]byte, 2)
	binary.BigEndian.PutUint16(port, 80)
	req = append(req, port...)
	_, _ = clientConn.Write(req)

	target := <-done
	if target != "93.184.216.34:80" {
		t.Errorf("target = %q, want %q", target, "93.184.216.34:80")
	}

	clientConn.Close()
}

// --- SOCKS5 full integration test (loopback) ---

func TestSOCKS5_FullConnection(t *testing.T) {
	// Start a mock target server
	targetLn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen target: %v", err)
	}
	defer targetLn.Close()

	targetAddr := targetLn.Addr().String()
	go func() {
		conn, err := targetLn.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		buf := make([]byte, 100)
		n, _ := conn.Read(buf)
		_, _ = conn.Write([]byte("echo:" + string(buf[:n])))
	}()

	// Start SOCKS5 server
	auth := &AllowAllAuth{DefaultWallet: "test-wallet"}
	tracker := NewSessionTracker(10)
	srv := NewSOCKS5Server(&DirectDialer{}, auth, tracker, nil)

	socks5Ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen socks5: %v", err)
	}
	srv.listener = socks5Ln
	socks5Addr := socks5Ln.Addr().String()

	go func() {
		for {
			conn, err := socks5Ln.Accept()
			if err != nil {
				return
			}
			go srv.handleConnection(conn)
		}
	}()
	defer socks5Ln.Close()

	// Connect through SOCKS5
	conn, err := net.DialTimeout("tcp", socks5Addr, 5*time.Second)
	if err != nil {
		t.Fatalf("dial socks5: %v", err)
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(5 * time.Second))

	// Send greeting
	_, _ = conn.Write([]byte{socks5Version, 1, authNone})
	resp := make([]byte, 2)
	_, _ = io.ReadFull(conn, resp)
	if resp[0] != socks5Version || resp[1] != authNone {
		t.Fatalf("greeting response = %v", resp)
	}

	// Send CONNECT to target (IPv4)
	host, portStr, _ := net.SplitHostPort(targetAddr)
	ip := net.ParseIP(host).To4()
	var portNum uint16
	_, _ = fmt.Sscanf(portStr, "%d", &portNum)

	req := []byte{socks5Version, cmdConnect, 0x00, addrIPv4}
	req = append(req, ip...)
	port := make([]byte, 2)
	binary.BigEndian.PutUint16(port, portNum)
	req = append(req, port...)
	_, _ = conn.Write(req)

	// Read connect reply (at least 10 bytes for IPv4)
	reply := make([]byte, 10)
	_, _ = io.ReadFull(conn, reply)
	if reply[1] != repSuccess {
		t.Fatalf("connect reply = %d, want success", reply[1])
	}

	// Send data through the tunnel
	_, _ = conn.Write([]byte("hello"))

	// Read echoed response
	buf := make([]byte, 100)
	n, err := conn.Read(buf)
	if err != nil {
		t.Fatalf("read response: %v", err)
	}
	if string(buf[:n]) != "echo:hello" {
		t.Errorf("response = %q, want %q", string(buf[:n]), "echo:hello")
	}
}

// --- HTTP Proxy tests ---

func TestHTTPProxy_CONNECT(t *testing.T) {
	// Start a target HTTPS server (we just echo, no real TLS needed for tunnel test)
	targetLn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer targetLn.Close()

	go func() {
		conn, err := targetLn.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		buf := make([]byte, 100)
		n, _ := conn.Read(buf)
		_, _ = conn.Write([]byte("tunnel:" + string(buf[:n])))
	}()

	// Create HTTP proxy
	auth := &AllowAllAuth{DefaultWallet: "w1"}
	tracker := NewSessionTracker(10)
	proxy := NewHTTPProxyServer(&DirectDialer{}, auth, tracker, nil)

	proxyServer := httptest.NewServer(proxy)
	defer proxyServer.Close()

	// Connect to proxy and issue CONNECT
	proxyConn, err := net.Dial("tcp", strings.TrimPrefix(proxyServer.URL, "http://"))
	if err != nil {
		t.Fatalf("dial proxy: %v", err)
	}
	defer proxyConn.Close()
	_ = proxyConn.SetDeadline(time.Now().Add(5 * time.Second))

	// Send CONNECT request
	connectReq := fmt.Sprintf("CONNECT %s HTTP/1.1\r\nHost: %s\r\n\r\n",
		targetLn.Addr().String(), targetLn.Addr().String())
	_, _ = proxyConn.Write([]byte(connectReq))

	// Read response
	buf := make([]byte, 1024)
	n, err := proxyConn.Read(buf)
	if err != nil {
		t.Fatalf("read connect response: %v", err)
	}
	resp := string(buf[:n])
	if !strings.Contains(resp, "200") {
		t.Fatalf("CONNECT response = %q, want 200", resp)
	}

	// Send data through tunnel
	_, _ = proxyConn.Write([]byte("data"))
	n, err = proxyConn.Read(buf)
	if err != nil {
		t.Fatalf("read tunnel data: %v", err)
	}
	if string(buf[:n]) != "tunnel:data" {
		t.Errorf("tunnel response = %q, want %q", string(buf[:n]), "tunnel:data")
	}
}

func TestHTTPProxy_Forward(t *testing.T) {
	// Start a target HTTP server
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Test", "proxied")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("forwarded response"))
	}))
	defer target.Close()

	// Create HTTP proxy
	auth := &AllowAllAuth{DefaultWallet: "w1"}
	tracker := NewSessionTracker(10)
	proxy := NewHTTPProxyServer(&DirectDialer{}, auth, tracker, nil)

	proxyServer := httptest.NewServer(proxy)
	defer proxyServer.Close()

	// Manual forward request: send absolute URL to proxy
	proxyConn, err := net.Dial("tcp", strings.TrimPrefix(proxyServer.URL, "http://"))
	if err != nil {
		t.Fatalf("dial proxy: %v", err)
	}
	defer proxyConn.Close()
	_ = proxyConn.SetDeadline(time.Now().Add(5 * time.Second))

	forwardReq := fmt.Sprintf("GET %s/test HTTP/1.1\r\nHost: %s\r\nConnection: close\r\n\r\n",
		target.URL, strings.TrimPrefix(target.URL, "http://"))
	_, _ = proxyConn.Write([]byte(forwardReq))

	buf := make([]byte, 4096)
	n, _ := proxyConn.Read(buf)
	resp := string(buf[:n])

	if !strings.Contains(resp, "200") {
		t.Errorf("forward response status not 200: %s", resp)
	}
	if !strings.Contains(resp, "forwarded response") {
		t.Errorf("forward response body missing: %s", resp)
	}
}

// --- REST Handler tests ---

func TestRESTHandler_Status(t *testing.T) {
	cfg := DefaultConfig()
	srv := NewServer(cfg, &DirectDialer{}, &AllowAllAuth{DefaultWallet: "w1"})
	handler := NewRESTHandler(srv)

	mux := http.NewServeMux()
	pass := func(h http.HandlerFunc) http.HandlerFunc { return h }
	handler.RegisterRoutes(mux, pass, pass)

	req := httptest.NewRequest("GET", "/v1/proxy/status", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status: %d", w.Code)
	}

	var status map[string]any
	if err := json.NewDecoder(w.Body).Decode(&status); err != nil {
		t.Fatalf("decode status: %v", err)
	}
	if status["max_sessions"] != float64(1000) {
		t.Errorf("max_sessions = %v, want 1000", status["max_sessions"])
	}
}

func TestRESTHandler_Sessions(t *testing.T) {
	cfg := DefaultConfig()
	srv := NewServer(cfg, &DirectDialer{}, &AllowAllAuth{DefaultWallet: "w1"})
	handler := NewRESTHandler(srv)

	// Add sessions for two wallets
	srv.Tracker().Add(&Session{
		ID:       "test-session",
		Wallet:   "w1",
		Protocol: "socks5",
		Target:   "example.com:443",
	})
	srv.Tracker().Add(&Session{
		ID:       "other-session",
		Wallet:   "w2",
		Protocol: "http_connect",
		Target:   "other.com:443",
	})

	mux := http.NewServeMux()
	pass := func(h http.HandlerFunc) http.HandlerFunc { return h }
	handler.RegisterRoutes(mux, pass, pass)

	// List sessions — should only see own wallet's sessions
	req := httptest.NewRequest("GET", "/v1/proxy/sessions", nil)
	req.Header.Set("X-Moltbunker-Verified-Wallet", "w1")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("list sessions: status %d", w.Code)
	}

	var listResp struct {
		Sessions []Session `json:"sessions"`
		Count    int       `json:"count"`
	}
	if err := json.NewDecoder(w.Body).Decode(&listResp); err != nil {
		t.Fatalf("decode list response: %v", err)
	}
	if listResp.Count != 1 {
		t.Errorf("count = %d, want 1 (only own sessions)", listResp.Count)
	}

	// Get session by ID (own)
	req = httptest.NewRequest("GET", "/v1/proxy/sessions/test-session", nil)
	req.Header.Set("X-Moltbunker-Verified-Wallet", "w1")
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("get own session: status %d", w.Code)
	}

	// Get session by ID (other wallet's) — should return 404
	req = httptest.NewRequest("GET", "/v1/proxy/sessions/other-session", nil)
	req.Header.Set("X-Moltbunker-Verified-Wallet", "w1")
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("get other session: status %d, want 404", w.Code)
	}

	// Delete session (own)
	req = httptest.NewRequest("DELETE", "/v1/proxy/sessions/test-session", nil)
	req.Header.Set("X-Moltbunker-Verified-Wallet", "w1")
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("delete session: status %d", w.Code)
	}

	// other-session should still exist
	if srv.Tracker().Count() != 1 {
		t.Errorf("tracker count = %d, want 1 (other session remains)", srv.Tracker().Count())
	}
}

func TestRESTHandler_Usage(t *testing.T) {
	cfg := DefaultConfig()
	srv := NewServer(cfg, &DirectDialer{}, &AllowAllAuth{DefaultWallet: "w1"})
	handler := NewRESTHandler(srv)

	srv.Tracker().Add(&Session{
		ID:       "s1",
		Wallet:   "w1",
		Protocol: "socks5",
		BytesIn:  1000,
		BytesOut: 2000,
	})

	mux := http.NewServeMux()
	pass := func(h http.HandlerFunc) http.HandlerFunc { return h }
	handler.RegisterRoutes(mux, pass, pass)

	req := httptest.NewRequest("GET", "/v1/proxy/usage", nil)
	req.Header.Set("X-Moltbunker-Verified-Wallet", "w1")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("usage: status %d", w.Code)
	}

	var report UsageReport
	if err := json.NewDecoder(w.Body).Decode(&report); err != nil {
		t.Fatalf("decode usage report: %v", err)
	}
	if report.TotalIn != 1000 {
		t.Errorf("total_in = %d, want 1000", report.TotalIn)
	}
	if report.TotalOut != 2000 {
		t.Errorf("total_out = %d, want 2000", report.TotalOut)
	}
	if report.SessionCount != 1 {
		t.Errorf("session_count = %d, want 1", report.SessionCount)
	}
}

// --- Server tests ---

func TestServer_Usage(t *testing.T) {
	srv := NewServer(DefaultConfig(), &DirectDialer{}, &AllowAllAuth{DefaultWallet: "w1"})

	srv.Tracker().Add(&Session{ID: "a", Wallet: "w1", BytesIn: 100, BytesOut: 200})
	srv.Tracker().Add(&Session{ID: "b", Wallet: "w2", BytesIn: 300, BytesOut: 400})
	srv.Tracker().Add(&Session{ID: "c", Wallet: "w1", BytesIn: 500, BytesOut: 600})

	report := srv.Usage("w1")
	if report.SessionCount != 2 {
		t.Errorf("w1 session count = %d, want 2", report.SessionCount)
	}
	if report.TotalIn != 600 {
		t.Errorf("w1 total_in = %d, want 600", report.TotalIn)
	}
	if report.TotalOut != 800 {
		t.Errorf("w1 total_out = %d, want 800", report.TotalOut)
	}
}

func TestServer_IsRunning(t *testing.T) {
	srv := NewServer(DefaultConfig(), &DirectDialer{}, &AllowAllAuth{DefaultWallet: "w1"})
	if srv.IsRunning() {
		t.Error("should not be running initially")
	}
}
