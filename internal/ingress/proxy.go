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
	resolver      *Resolver
	tunnelClient  *tunnel.Client
	reverseOpener ReverseStreamOpener // reverse tunnel (optional)
	domain        string              // e.g., "moltbunker.dev"
	server        *http.Server

	// middleware is the optional L7 edge chain (WAF + abuse limits + metrics).
	// When nil the request path is unchanged (zero overhead). EDGE-01.
	middleware *IngressMiddleware

	// customDomains, when set, routes verified BYO custom hostnames (e.g.
	// app.customer.com) to their deployment alongside *.<domain> traffic.
	// Nil = the original behavior (only *.<domain> Host headers are routed).
	// EDGE-02.
	customDomains *DomainOwnershipStore

	// blocklist, when set, is the operator takedown kill-switch consulted on
	// EVERY request before routing. It is checked at the ingress proxy itself
	// (not only inside the reverse tunnel) so a takedown reliably severs a live
	// deployment regardless of whether it is served via the forward or reverse
	// path. EDGE-02.
	blocklist tunnel.BlocklistChecker
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

// SetMiddleware installs the optional L7 edge middleware (WAF + per-tenant
// abuse limits + edge metrics). When set, every non-WebSocket request is run
// through the full chain before being dispatched to the upstream.
//
// WebSocket upgrade requests skip only the WAF body-buffering / ResponseWriter
// wrapping steps (incompatible with the hijack-and-stream model), but they are
// still subject to the per-tenant rate limit and concurrency cap via
// acquireWebSocket — so a tenant subdomain cannot be used to open unbounded
// concurrent long-lived connections. The concurrency slot is held for the
// connection's lifetime and drives the active-tunnel-session gauge. L4
// byte-stream limits in internal/tunnel also still apply to WebSocket traffic.
//
// When nil the request path is unchanged (zero overhead). EDGE-01.
func (p *Proxy) SetMiddleware(m *IngressMiddleware) {
	p.middleware = m
}

// SetCustomDomains wires the verified-custom-domain store so the proxy will
// route BYO hostnames (e.g. app.customer.com) to their mapped deployment.
// When nil the routing path is unchanged (only *.<domain> Host headers are
// served). EDGE-02.
func (p *Proxy) SetCustomDomains(store *DomainOwnershipStore) {
	p.customDomains = store
}

// SetBlocklist wires the operator takedown kill-switch. When set, every
// incoming request is checked against the blocklist BEFORE routing — by both
// the original Host header (the natural custom-domain takedown target) and the
// resolved subdomain / deployment ID — and a blocked request gets a 403.
//
// This enforces the takedown at the ingress proxy itself, so the forward-tunnel
// serving path (resolver -> tunnelClient.OpenTunnel) and verified custom hosts
// are covered, not just the reverse-tunnel registration/stream-open path.
// When nil, no takedown enforcement happens at the proxy layer. EDGE-02.
func (p *Proxy) SetBlocklist(bl tunnel.BlocklistChecker) {
	p.blocklist = bl
}

// isBlocked reports whether any of the given identifiers is on the takedown
// blocklist, returning the first match's operator reason. Empty identifiers are
// skipped. When no blocklist is wired it always returns false.
func (p *Proxy) isBlocked(ids ...string) (bool, string) {
	if p.blocklist == nil {
		return false, ""
	}
	for _, id := range ids {
		if id == "" {
			continue
		}
		if blocked, reason := p.blocklist.IsBlocked(id); blocked {
			return true, reason
		}
	}
	return false, ""
}

// resolveCustomDomain looks up a verified BYO host and returns its deployment
// ID for routing. Returns "" when no custom-domain store is wired or the host
// is not a verified custom domain.
func (p *Proxy) resolveCustomDomain(host string) string {
	if p.customDomains == nil {
		return ""
	}
	rec, ok := p.customDomains.LookupByHost(host)
	if !ok {
		return ""
	}
	return rec.DeploymentID
}

// acquireWebSocket applies the WebSocket-compatible abuse gates (per-tenant
// rate + concurrency) via the installed middleware and returns a release
// closure to hold for the connection's lifetime. When no middleware is
// installed it is a no-op that always allows. On rejection the middleware has
// already written the 429/503 response, so the caller must simply return.
//
// tunnelType is "forward" or "reverse"; the tier is not known at this layer
// yet, so "default" is used for the active-session gauge label.
func (p *Proxy) acquireWebSocket(subdomain, tunnelType string, w http.ResponseWriter) (release func(), ok bool) {
	if p.middleware == nil {
		return func() {}, true
	}
	return p.middleware.AllowWebSocket(subdomain, tunnelType, "default", w)
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
	// Strip any client-supplied X-Moltbunker-* control headers before routing.
	// These are proxy-internal signals; a client must never be able to forge
	// them (e.g. spoof the "verified custom domain" marker) or leak them to a
	// tenant backend. The trusted markers are derived below and passed down as
	// typed arguments, never via the request header. EDGE-02.
	stripInternalHeaders(r)

	// Extract subdomain from Host header
	customDomain := false
	originalHost := r.Host
	subdomain := p.extractSubdomain(r.Host)
	if subdomain == "" {
		// Not a *.<domain> host. It may be a verified BYO custom domain
		// (EDGE-02) whose Host header is the customer's own hostname; route it
		// via its mapped deployment ID. The deployment ID resolves through the
		// same forward-tunnel / reverse-tunnel path as a normal subdomain.
		if depID := p.resolveCustomDomain(r.Host); depID != "" {
			customDomain = true
			subdomain = depID
		} else {
			http.Error(w, "invalid host", http.StatusBadRequest)
			return
		}
	}

	// Operator takedown kill-switch, enforced at the ingress proxy on EVERY
	// request (forward AND reverse paths) before any routing. Check both the
	// original Host header (the natural custom-domain takedown target,
	// app.customer.com) and the resolved subdomain / deployment ID, so a
	// takedown reliably severs the deployment no matter how it was addressed.
	// EDGE-02.
	if blocked, _ := p.isBlocked(originalHost, subdomain); blocked {
		http.Error(w, "service unavailable", http.StatusForbidden)
		return
	}

	// Run the L7 edge chain (WAF + per-tenant abuse limits + metrics) when
	// installed. WebSocket upgrades bypass the WAF body buffering (frames are
	// streamed, not request/response shaped) but still flow through dispatch —
	// where they are independently rate-limited + concurrency-capped, see
	// dispatchRequest.
	if p.middleware != nil && !isWebSocketUpgrade(r) {
		p.middleware.Wrap(subdomain, http.HandlerFunc(func(mw http.ResponseWriter, mr *http.Request) {
			p.dispatchRequest(mw, mr, subdomain, customDomain)
		})).ServeHTTP(w, r)
		return
	}

	p.dispatchRequest(w, r, subdomain, customDomain)
}

// stripInternalHeaders removes any client-supplied X-Moltbunker-* control
// headers from the inbound request so they cannot be forged or leaked to tenant
// backends. The proxy sets its own response markers from trusted state. EDGE-02.
func stripInternalHeaders(r *http.Request) {
	for k := range r.Header {
		if strings.HasPrefix(http.CanonicalHeaderKey(k), "X-Moltbunker-") {
			r.Header.Del(k)
		}
	}
}

// dispatchRequest resolves the service for subdomain and proxies the request to
// the provider via a forward tunnel, falling back to the reverse tunnel
// registry. This is the original handleRequest body, extracted so the edge
// middleware can wrap it. Behavior is unchanged when middleware is nil.
func (p *Proxy) dispatchRequest(w http.ResponseWriter, r *http.Request, subdomain string, customDomain bool) {
	// Resolve service location (prefix match or vanity name)
	service, err := p.resolver.Resolve(subdomain)
	if err != nil {
		// Fallback: try reverse tunnel registry
		if p.reverseOpener != nil {
			if stream, streamErr := p.reverseOpener.OpenStream(subdomain); streamErr == nil {
				defer stream.Close()
				if isWebSocketUpgrade(r) {
					// Apply per-tenant abuse gates to the WebSocket upgrade
					// (rate + concurrency; WAF/body-buffering is N/A for a
					// hijacked stream). release is held for the connection's
					// lifetime and also drives the active-session gauge.
					if release, ok := p.acquireWebSocket(subdomain, "reverse", w); ok {
						defer release()
						p.handleWebSocketViaStream(w, r, stream)
					}
					return
				}
				p.proxyHTTPViaStream(w, r, stream, subdomain, customDomain)
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
		// Apply per-tenant abuse gates to the WebSocket upgrade (rate +
		// concurrency; WAF/body-buffering is N/A for a hijacked stream).
		// release is held for the connection's lifetime and also drives the
		// active-session gauge.
		if release, ok := p.acquireWebSocket(subdomain, "forward", w); ok {
			defer release()
			p.handleWebSocket(w, r, tun)
		}
		return
	}

	p.proxyHTTP(w, r, tun, service, customDomain)
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
// customDomain is the trusted "request arrived on a verified BYO host" signal,
// derived from resolveCustomDomain (never from a client header). EDGE-02.
func (p *Proxy) proxyHTTP(w http.ResponseWriter, r *http.Request, tun tunnel.Tunnel, service *ServiceEntry, customDomain bool) {
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
	if customDomain {
		w.Header().Set("X-Moltbunker-CustomDomain", "true")
	}

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
// customDomain is the trusted "request arrived on a verified BYO host" signal,
// derived from resolveCustomDomain (never from a client header). EDGE-02.
func (p *Proxy) proxyHTTPViaStream(w http.ResponseWriter, r *http.Request, stream net.Conn, subdomain string, customDomain bool) {
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
	if customDomain {
		w.Header().Set("X-Moltbunker-CustomDomain", "true")
	}

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
	"Date":         true,
	"Retry-After":  true,
	"X-Request-Id": true,
}
