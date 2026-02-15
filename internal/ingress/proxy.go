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

// Proxy is an HTTP reverse proxy that routes subdomain requests to container services.
// Incoming requests like "a1b2c3d4.moltbunker.dev" are parsed to extract the deployment ID,
// resolved via the gossip-based service resolver, and proxied through a TLS tunnel.
type Proxy struct {
	resolver     *Resolver
	tunnelClient *tunnel.Client
	domain       string // e.g., "moltbunker.dev"
	server       *http.Server
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
// It extracts the deployment ID from the Host header subdomain,
// looks up the provider via gossip, opens a tunnel, and proxies the request.
func (p *Proxy) handleRequest(w http.ResponseWriter, r *http.Request) {
	// Extract deployment ID from Host header
	deploymentID := p.extractDeploymentID(r.Host)
	if deploymentID == "" {
		http.Error(w, "invalid host", http.StatusBadRequest)
		return
	}

	// Resolve service location via gossip
	service, err := p.resolver.Resolve(deploymentID)
	if err != nil {
		logging.Debug("service not found",
			"deployment_id", deploymentID,
			"error", err.Error(),
			logging.Component("ingress"))
		http.Error(w, "service not found", http.StatusNotFound)
		return
	}

	// Open tunnel to the provider
	tun, err := p.tunnelClient.OpenTunnel(service.ProviderAddr, deploymentID, service.ContainerPort)
	if err != nil {
		logging.Error("tunnel open failed",
			"deployment_id", deploymentID,
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

// extractDeploymentID parses the deployment ID from the Host header.
// Expected format: "<deployment-id>.moltbunker.dev" or "<deployment-id>.moltbunker.dev:port"
func (p *Proxy) extractDeploymentID(host string) string {
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

	// Copy response headers
	for k, vv := range resp.Header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}

	// Add proxy headers
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

