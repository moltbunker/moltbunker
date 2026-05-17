package proxy

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/moltbunker/moltbunker/internal/logging"
)

// HTTPProxyServer handles HTTP CONNECT tunneling and HTTP forward proxying.
type HTTPProxyServer struct {
	dialer  Dialer
	auth    Authenticator
	tracker *SessionTracker
	acl     *ACL
	server  *http.Server

	ctx    context.Context
	cancel context.CancelFunc
}

// NewHTTPProxyServer creates a new HTTP proxy server.
func NewHTTPProxyServer(dialer Dialer, auth Authenticator, tracker *SessionTracker, acl *ACL) *HTTPProxyServer {
	ctx, cancel := context.WithCancel(context.Background())
	p := &HTTPProxyServer{
		dialer:  dialer,
		auth:    auth,
		tracker: tracker,
		acl:     acl,
		ctx:     ctx,
		cancel:  cancel,
	}
	p.server = &http.Server{
		Handler:           p,
		ReadHeaderTimeout: 30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
	return p
}

// ListenAndServe starts the HTTP proxy server.
func (p *HTTPProxyServer) ListenAndServe(addr string) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("http proxy listen: %w", err)
	}
	logging.Info("HTTP proxy listening", "addr", addr, logging.Component("proxy"))
	return p.server.Serve(ln)
}

// Close shuts down the HTTP proxy server.
func (p *HTTPProxyServer) Close() error {
	p.cancel()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return p.server.Shutdown(ctx)
}

// ServeHTTP dispatches between CONNECT tunneling and forward proxying.
func (p *HTTPProxyServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Authenticate
	wallet := p.authenticate(r)
	if wallet == "" {
		http.Error(w, "Proxy Authentication Required", http.StatusProxyAuthRequired)
		return
	}

	if r.Method == http.MethodConnect {
		p.handleConnect(w, r, wallet)
	} else {
		p.handleForward(w, r, wallet)
	}
}

// authenticate extracts and validates proxy credentials from the request.
func (p *HTTPProxyServer) authenticate(r *http.Request) string {
	// Check Proxy-Authorization header
	username, password, ok := r.BasicAuth()
	if !ok {
		// Try Proxy-Authorization header directly
		proxyAuth := r.Header.Get("Proxy-Authorization")
		if proxyAuth == "" {
			// Fall back to no-auth mode
			return p.auth.Authenticate("", "")
		}
	}
	return p.auth.Authenticate(username, password)
}

// handleConnect implements HTTP CONNECT tunneling for HTTPS traffic.
func (p *HTTPProxyServer) handleConnect(w http.ResponseWriter, r *http.Request, wallet string) {
	target := r.Host
	if !strings.Contains(target, ":") {
		target = target + ":443"
	}

	// ACL check
	host, _, _ := net.SplitHostPort(target)
	if p.acl != nil && !p.acl.IsAllowed(host) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Create session
	sessionID, err := generateSessionID()
	if err != nil {
		logging.Warn("failed to generate http_connect session ID",
			"error", err.Error(),
			logging.Component("proxy"))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	session := &Session{
		ID:        sessionID,
		Wallet:    wallet,
		Protocol:  "http_connect",
		Target:    target,
		StartedAt: time.Now(),
	}
	if !p.tracker.Add(session) {
		http.Error(w, "Too Many Connections", http.StatusTooManyRequests)
		return
	}
	defer p.tracker.Remove(sessionID)

	// Connect to target
	targetConn, err := p.dialer.DialContext(p.ctx, "tcp", target)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to connect: %v", err), http.StatusBadGateway)
		return
	}
	defer targetConn.Close()

	// Hijack the HTTP connection to get the raw TCP connection
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}

	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer clientConn.Close()

	// Send 200 Connection Established
	if _, err := clientConn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n")); err != nil {
		logging.Debug("failed to write connect response",
			"error", err.Error(),
			logging.Component("proxy"))
		return
	}

	// Relay data with bandwidth metering
	meter := NewBandwidthMeter()
	relay(p.ctx, clientConn, targetConn, meter)

	session.BytesIn = meter.BytesRead()
	session.BytesOut = meter.BytesWritten()
}

// Hop-by-hop headers that should not be forwarded.
var hopByHopHeaders = []string{
	"Connection",
	"Keep-Alive",
	"Proxy-Authenticate",
	"Proxy-Authorization",
	"TE",
	"Trailers",
	"Transfer-Encoding",
	"Upgrade",
}

// handleForward implements HTTP forward proxying for plain HTTP requests.
func (p *HTTPProxyServer) handleForward(w http.ResponseWriter, r *http.Request, wallet string) {
	if !r.URL.IsAbs() {
		http.Error(w, "Absolute URL required for proxy", http.StatusBadRequest)
		return
	}

	target := r.URL.Host
	if !strings.Contains(target, ":") {
		target = target + ":80"
	}

	// ACL check
	host, _, _ := net.SplitHostPort(target)
	if p.acl != nil && !p.acl.IsAllowed(host) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Create session
	sessionID, err := generateSessionID()
	if err != nil {
		logging.Warn("failed to generate http_forward session ID",
			"error", err.Error(),
			logging.Component("proxy"))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	session := &Session{
		ID:        sessionID,
		Wallet:    wallet,
		Protocol:  "http_forward",
		Target:    target,
		StartedAt: time.Now(),
	}
	if !p.tracker.Add(session) {
		http.Error(w, "Too Many Connections", http.StatusTooManyRequests)
		return
	}
	defer p.tracker.Remove(sessionID)

	// Build outbound request
	outReq := r.Clone(p.ctx)
	outReq.RequestURI = ""

	// Remove hop-by-hop headers
	for _, h := range hopByHopHeaders {
		outReq.Header.Del(h)
	}

	// Use a transport that goes through our dialer
	transport := &http.Transport{
		DialContext: p.dialer.DialContext,
	}
	resp, err := transport.RoundTrip(outReq)
	if err != nil {
		logging.Debug("http forward error", "target", target, "error", err.Error(), logging.Component("proxy"))
		http.Error(w, fmt.Sprintf("Proxy error: %v", err), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Copy response headers (except hop-by-hop)
	for key, vals := range resp.Header {
		isHopByHop := false
		for _, h := range hopByHopHeaders {
			if strings.EqualFold(key, h) {
				isHopByHop = true
				break
			}
		}
		if !isHopByHop {
			for _, v := range vals {
				w.Header().Add(key, v)
			}
		}
	}

	w.WriteHeader(resp.StatusCode)
	bytesWritten, _ := io.Copy(w, resp.Body)
	session.BytesOut = bytesWritten
}
