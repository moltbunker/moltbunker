// Package ingress implements an HTTP reverse proxy that routes requests
// from public URLs (e.g., <id>.moltbunker.dev) to the correct provider
// via TLS 1.3 tunnels. Ingress nodes participate in the gossip protocol
// to discover which provider hosts each deployment.
package ingress

import (
	"bufio"
	"context"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/moltbunker/moltbunker/internal/logging"
	"github.com/moltbunker/moltbunker/internal/tunnel"
)

// ReverseStreamOpener opens yamux streams to providers via reverse tunnels.
// Implemented by tunnel.ReverseServer.
type ReverseStreamOpener interface {
	OpenStream(subdomain string) (net.Conn, error)
}

// Proxy is an HTTP reverse proxy that routes subdomain requests to container services.
// Incoming requests like "a1b2c3d4.moltbunker.dev" are parsed to extract the deployment ID,
// resolved via the gossip-based service resolver, and proxied through a TLS tunnel.
// It supports both forward tunnels (ingress dials provider) and reverse tunnels
// (provider dials ingress, traffic multiplexed via yamux).
type Proxy struct {
	resolver       *Resolver
	tunnelClient   *tunnel.Client
	reverseOpener  ReverseStreamOpener // reverse tunnel (optional)
	domain         string              // e.g., "moltbunker.dev"
	server         *http.Server
}

// NewProxy creates a new ingress proxy.
func NewProxy(resolver *Resolver, tunnelClient *tunnel.Client, domain string) *Proxy {
	p := &Proxy{
		resolver:     resolver,
		tunnelClient: tunnelClient,
		domain:       domain,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", p.handleRequest)

	p.server = &http.Server{
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	return p
}

// SetReverseStreamOpener sets the reverse tunnel stream opener.
// When set, the proxy will try reverse tunnels as a fallback if forward tunnel
// resolution fails (i.e., the subdomain is not in gossip state).
func (p *Proxy) SetReverseStreamOpener(opener ReverseStreamOpener) {
	p.reverseOpener = opener
}

// Serve starts the proxy on the given listener. Blocks until stopped.
func (p *Proxy) Serve(listener net.Listener) error {
	logging.Info("ingress proxy started",
		"domain", p.domain,
		logging.Component("ingress"))
	return p.server.Serve(listener)
}

// Shutdown gracefully stops the proxy.
func (p *Proxy) Shutdown(ctx context.Context) error {
	return p.server.Shutdown(ctx)
}

// handleRequest processes an incoming HTTP request.
// It extracts the subdomain from the Host header, resolves the service
// (by deployment ID prefix or vanity name), opens a tunnel, and proxies.
func (p *Proxy) handleRequest(w http.ResponseWriter, r *http.Request) {
	// Extract subdomain from Host header
	subdomain := p.extractSubdomain(r.Host)
	if subdomain == "" {
		http.Error(w, "invalid host", http.StatusBadRequest)
		return
	}

	// Resolve service location (prefix match or vanity name)
	service, err := p.resolver.Resolve(subdomain)
	if err != nil {
		// Fallback: try reverse tunnel registry
		if p.reverseOpener != nil {
			if stream, streamErr := p.reverseOpener.OpenStream(subdomain); streamErr == nil {
				defer stream.Close()
				if isWebSocketUpgrade(r) {
					p.handleWebSocketViaStream(w, r, stream)
					return
				}
				p.proxyHTTPViaStream(w, r, stream, subdomain)
				return
			}
		}

		logging.Debug("service not found",
			"subdomain", subdomain,
			"error", err.Error(),
			logging.Component("ingress"))
		http.Error(w, "service not found", http.StatusNotFound)
		return
	}

	// Open tunnel to the provider
	tun, err := p.tunnelClient.OpenTunnel(service.ProviderAddr, service.DeploymentID, service.ContainerPort, service.ProviderNodeID)
	if err != nil {
		logging.Error("tunnel open failed",
			"deployment_id", service.DeploymentID,
			"provider", service.ProviderAddr,
			logging.Err(err),
			logging.Component("ingress"))
		http.Error(w, "service unavailable", http.StatusBadGateway)
		return
	}
	defer tun.Close()

	// For HTTP/1.1: write the original request through the tunnel and relay response.
	// For WebSocket upgrades: hijack and proxy bidirectionally.
	if isWebSocketUpgrade(r) {
		p.handleWebSocket(w, r, tun)
		return
	}

	p.proxyHTTP(w, r, tun, service)
}

// extractSubdomain parses the deployment ID from the Host header.
// Expected format: "<deployment-id>.moltbunker.dev" or "<deployment-id>.moltbunker.dev:port"
func (p *Proxy) extractSubdomain(host string) string {
	// Strip port if present
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}

	// Remove the domain suffix
	suffix := "." + p.domain
	if !strings.HasSuffix(host, suffix) {
		return ""
	}

	deploymentID := strings.TrimSuffix(host, suffix)
	if deploymentID == "" {
		return ""
	}

	return deploymentID
}

// proxyHTTP forwards an HTTP request through the tunnel and relays the response.
func (p *Proxy) proxyHTTP(w http.ResponseWriter, r *http.Request, tun tunnel.Tunnel, service *ServiceEntry) {
	// Write the HTTP request through the tunnel connection
	if err := r.Write(tun); err != nil {
		http.Error(w, "proxy write failed", http.StatusBadGateway)
		return
	}

	// Read the response from the tunnel
	resp, err := http.ReadResponse(bufio.NewReader(tun), r)
	if err != nil {
		http.Error(w, "proxy read failed", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Copy ONLY safe response headers from the backend container.
	// Containers share the *.moltbunker.dev domain, so a malicious container
	// could inject CORS/CSP/security headers to attack other tenants.
	// Use an allowlist — not a blocklist — to prevent cross-tenant attacks.
	for k, vv := range resp.Header {
		canonical := http.CanonicalHeaderKey(k)
		if !allowedResponseHeaders[canonical] {
			continue
		}
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}

	// Set security headers server-side — containers must NOT control these
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

	// Add proxy identification headers
	w.Header().Set("X-Moltbunker-Provider", service.ProviderNodeID)
	w.Header().Set("X-Moltbunker-Deployment", service.DeploymentID)

	w.WriteHeader(resp.StatusCode)

	// Stream response body
	buf := make([]byte, 32*1024)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			if _, wErr := w.Write(buf[:n]); wErr != nil {
				return
			}
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}
		if err != nil {
			return
		}
	}
}

// proxyHTTPViaStream forwards an HTTP request through a reverse tunnel yamux stream.
func (p *Proxy) proxyHTTPViaStream(w http.ResponseWriter, r *http.Request, stream net.Conn, subdomain string) {
	// Write the HTTP request through the stream
	if err := r.Write(stream); err != nil {
		http.Error(w, "proxy write failed", http.StatusBadGateway)
		return
	}

	// Read the response from the stream
	resp, err := http.ReadResponse(bufio.NewReader(stream), r)
	if err != nil {
		http.Error(w, "proxy read failed", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Apply same response header allowlist as forward tunnels
	for k, vv := range resp.Header {
		canonical := http.CanonicalHeaderKey(k)
		if !allowedResponseHeaders[canonical] {
			continue
		}
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}

	// Set security headers — containers must NOT control these
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
	w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
	w.Header().Set("X-Moltbunker-Tunnel", "reverse")
	w.Header().Set("X-Moltbunker-Subdomain", subdomain)

	w.WriteHeader(resp.StatusCode)

	// Stream response body
	buf := make([]byte, 32*1024)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			if _, wErr := w.Write(buf[:n]); wErr != nil {
				return
			}
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}
		if err != nil {
			return
		}
	}
}

// handleWebSocketViaStream hijacks and proxies WebSocket through a reverse tunnel stream.
func (p *Proxy) handleWebSocketViaStream(w http.ResponseWriter, r *http.Request, stream net.Conn) {
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "websocket not supported", http.StatusInternalServerError)
		return
	}

	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, "hijack failed", http.StatusInternalServerError)
		return
	}
	defer clientConn.Close()

	// Forward the original request to the stream
	if err := r.Write(stream); err != nil {
		return
	}

	// Proxy bidirectionally
	_ = tunnel.ProxyBidirectional(r.Context(), clientConn, stream)
}

// handleWebSocket hijacks the HTTP connection and proxies WebSocket bidirectionally.
func (p *Proxy) handleWebSocket(w http.ResponseWriter, r *http.Request, tun tunnel.Tunnel) {
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "websocket not supported", http.StatusInternalServerError)
		return
	}

	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, "hijack failed", http.StatusInternalServerError)
		return
	}
	defer clientConn.Close()

	// Forward the original request to the tunnel
	if err := r.Write(tun); err != nil {
		return
	}

	// Proxy bidirectionally
	_ = tunnel.ProxyBidirectional(r.Context(), clientConn, tunnelToNetConn(tun))
}

// isWebSocketUpgrade checks if the request is a WebSocket upgrade.
func isWebSocketUpgrade(r *http.Request) bool {
	return strings.EqualFold(r.Header.Get("Upgrade"), "websocket") &&
		strings.Contains(strings.ToLower(r.Header.Get("Connection")), "upgrade")
}

// tunnelToNetConn wraps a tunnel.Tunnel to satisfy net.Conn for ProxyBidirectional.
type tunnelNetConn struct {
	tunnel.Tunnel
}

func tunnelToNetConn(t tunnel.Tunnel) net.Conn {
	if conn, ok := t.(net.Conn); ok {
		return conn
	}
	return &tunnelNetConn{Tunnel: t}
}

func (t *tunnelNetConn) LocalAddr() net.Addr                { return nil }
func (t *tunnelNetConn) RemoteAddr() net.Addr               { return nil }
func (t *tunnelNetConn) SetDeadline(_ time.Time) error      { return nil }
func (t *tunnelNetConn) SetReadDeadline(_ time.Time) error  { return nil }
func (t *tunnelNetConn) SetWriteDeadline(_ time.Time) error { return nil }

// allowedResponseHeaders is the set of response headers that the ingress proxy
// forwards from backend containers. All containers share the *.moltbunker.dev
// domain, so we MUST NOT forward security-sensitive headers (CORS, CSP, HSTS,
// Set-Cookie, X-Frame-Options, etc.) — those are set server-side by the proxy.
var allowedResponseHeaders = map[string]bool{
	// Content description
	"Content-Type":     true,
	"Content-Length":   true,
	"Content-Encoding": true,
	"Content-Language": true,
	"Content-Range":    true,

	// Caching
	"Cache-Control": true,
	"Etag":          true,
	"Last-Modified": true,
	"Expires":       true,
	"Vary":          true,
	"Age":           true,

	// Range requests
	"Accept-Ranges": true,

	// Misc safe headers
	"Date":          true,
	"Retry-After":   true,
	"X-Request-Id":  true,
}

