package molt

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/tetratelabs/wazero/api"

	"github.com/moltbunker/moltbunker/internal/logging"
	"github.com/moltbunker/moltbunker/internal/proxy"
)

const (
	httpRequestTimeout   = 30 * time.Second
	httpMaxResponseBytes = 10 * 1024 * 1024 // 10MB
)

// HostHTTPRequest is the JSON envelope sent from WASM to the host for HTTP requests.
type HostHTTPRequest struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    string            `json:"body,omitempty"` // base64-encoded
}

// HostHTTPResponse is the JSON envelope returned from the host to WASM.
type HostHTTPResponse struct {
	Status  int               `json:"status"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    string            `json:"body,omitempty"` // base64-encoded
}

// hostHTTPRequest executes an HTTP request on behalf of the WASM module.
// Params: [req_ptr i32, req_len i32] → [handle i32]
// Returns a positive handle with HostHTTPResponse JSON, or a negative error handle.
func hostHTTPRequest(ctx context.Context, mod api.Module, stack []uint64) {
	reqPtr := api.DecodeU32(stack[0])
	reqLen := api.DecodeU32(stack[1])

	svc := servicesFromContext(ctx)
	if svc == nil {
		stack[0] = api.EncodeI32(-1)
		return
	}

	if !svc.Config.HTTPEnabled {
		stack[0] = api.EncodeI32(svc.results.StoreError("http: service disabled"))
		return
	}

	mem := mod.Memory()
	if mem == nil {
		stack[0] = api.EncodeI32(svc.results.StoreError("http: no memory"))
		return
	}

	reqBytes, ok := mem.Read(reqPtr, reqLen)
	if !ok {
		stack[0] = api.EncodeI32(svc.results.StoreError("http: invalid memory read"))
		return
	}

	var req HostHTTPRequest
	if err := json.Unmarshal(reqBytes, &req); err != nil {
		stack[0] = api.EncodeI32(svc.results.StoreError(fmt.Sprintf("http: invalid request JSON: %v", err)))
		return
	}

	result, err := executeHTTPRequest(ctx, svc, &req)
	if err != nil {
		stack[0] = api.EncodeI32(svc.results.StoreError(fmt.Sprintf("http: %v", err)))
		return
	}

	respJSON, err := json.Marshal(result)
	if err != nil {
		stack[0] = api.EncodeI32(svc.results.StoreError(fmt.Sprintf("http: marshal response: %v", err)))
		return
	}

	stack[0] = api.EncodeI32(svc.results.Store(respJSON))
}

// ExecuteHTTPRequest performs an HTTP request using HostServices.
// Exported for use by the Deno JS runtime's host_call dispatch.
func ExecuteHTTPRequest(ctx context.Context, svc *HostServices, req *HostHTTPRequest) (*HostHTTPResponse, error) {
	return executeHTTPRequest(ctx, svc, req)
}

func executeHTTPRequest(ctx context.Context, svc *HostServices, req *HostHTTPRequest) (*HostHTTPResponse, error) {
	// Validate URL
	parsed, err := url.Parse(req.URL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	// SSRF guard: only allow http/https
	switch strings.ToLower(parsed.Scheme) {
	case "http", "https":
		// OK
	default:
		return nil, fmt.Errorf("blocked scheme: %s", parsed.Scheme)
	}

	// SSRF guard: block private/reserved IPs
	if err := validateHost(parsed.Hostname()); err != nil {
		return nil, err
	}

	// Allowlist/blocklist enforcement
	if err := checkHostPolicy(parsed.Hostname(), svc.Config.HTTPAllowedHosts, svc.Config.HTTPBlockedHosts); err != nil {
		return nil, err
	}

	// Build HTTP client with optional proxy transport
	var transport http.RoundTripper
	if svc.Proxy != nil {
		transport = &http.Transport{
			DialContext: svc.Proxy.DialContext,
		}
	} else {
		transport = &http.Transport{
			DialContext: (&proxy.DirectDialer{}).DialContext,
		}
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   httpRequestTimeout,
	}

	// Decode body
	var bodyReader io.Reader
	if req.Body != "" {
		decoded, err := base64.StdEncoding.DecodeString(req.Body)
		if err != nil {
			bodyReader = strings.NewReader(req.Body) // treat as raw string fallback
		} else {
			bodyReader = bytes.NewReader(decoded)
		}
	}

	method := req.Method
	if method == "" {
		method = "GET"
	}

	httpReq, err := http.NewRequestWithContext(ctx, method, req.URL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	for k, v := range req.Headers {
		httpReq.Header.Set(k, v)
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body (capped)
	body, err := io.ReadAll(io.LimitReader(resp.Body, httpMaxResponseBytes))
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	// Collect response headers
	headers := make(map[string]string, len(resp.Header))
	for k, vs := range resp.Header {
		headers[k] = strings.Join(vs, ", ")
	}

	logging.Debug("host.http_request completed",
		"method", method, "url", req.URL, "status", resp.StatusCode, "body_bytes", len(body))

	return &HostHTTPResponse{
		Status:  resp.StatusCode,
		Headers: headers,
		Body:    base64.StdEncoding.EncodeToString(body),
	}, nil
}

// ssrfBypass is a test hook to skip SSRF validation. Only set in tests.
var ssrfBypass bool

// validateHost checks if a hostname resolves to private/reserved IPs.
func validateHost(host string) error {
	if ssrfBypass {
		return nil
	}

	// Check if the host itself is an IP
	if ip := net.ParseIP(host); ip != nil {
		if isPrivateIP(ip) {
			return fmt.Errorf("blocked private IP: %s", host)
		}
		return nil
	}

	// Resolve hostname to check for DNS rebinding to private IPs
	addrs, err := net.LookupHost(host)
	if err != nil {
		return fmt.Errorf("DNS lookup failed: %w", err)
	}
	for _, addr := range addrs {
		if ip := net.ParseIP(addr); ip != nil && isPrivateIP(ip) {
			return fmt.Errorf("blocked: %s resolves to private IP %s", host, addr)
		}
	}
	return nil
}

// isPrivateIP returns true for private, loopback, link-local, and metadata IPs.
func isPrivateIP(ip net.IP) bool {
	privateRanges := []struct {
		network *net.IPNet
	}{
		{mustParseCIDR("10.0.0.0/8")},
		{mustParseCIDR("172.16.0.0/12")},
		{mustParseCIDR("192.168.0.0/16")},
		{mustParseCIDR("127.0.0.0/8")},
		{mustParseCIDR("169.254.0.0/16")},   // link-local
		{mustParseCIDR("169.254.169.254/32")}, // AWS/GCP metadata
		{mustParseCIDR("::1/128")},            // IPv6 loopback
		{mustParseCIDR("fc00::/7")},           // IPv6 unique local
		{mustParseCIDR("fe80::/10")},          // IPv6 link-local
	}

	for _, r := range privateRanges {
		if r.network.Contains(ip) {
			return true
		}
	}
	return false
}

func mustParseCIDR(s string) *net.IPNet {
	_, network, err := net.ParseCIDR(s)
	if err != nil {
		panic(fmt.Sprintf("invalid CIDR: %s", s))
	}
	return network
}

// checkHostPolicy enforces allowlist/blocklist for HTTP hosts.
func checkHostPolicy(host string, allowed, blocked []string) error {
	// If allowlist is set, only allow listed hosts
	if len(allowed) > 0 {
		for _, h := range allowed {
			if strings.EqualFold(host, h) {
				return nil
			}
		}
		return fmt.Errorf("host %s not in allowlist", host)
	}

	// Check blocklist
	for _, h := range blocked {
		if strings.EqualFold(host, h) {
			return fmt.Errorf("host %s is blocked", host)
		}
	}

	return nil
}
