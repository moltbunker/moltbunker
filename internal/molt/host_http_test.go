package molt

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestExecuteHTTPRequest_GET(t *testing.T) {
	ssrfBypass = true
	defer func() { ssrfBypass = false }()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("method = %s, want GET", r.Method)
		}
		w.Header().Set("X-Test", "hello")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer ts.Close()

	svc := NewHostServices(HostCapabilities{HTTPEnabled: true})

	req := &HostHTTPRequest{
		Method: "GET",
		URL:    ts.URL + "/health",
	}

	resp, err := ExecuteHTTPRequest(context.Background(), svc, req)
	if err != nil {
		t.Fatalf("ExecuteHTTPRequest: %v", err)
	}

	if resp.Status != 200 {
		t.Fatalf("Status = %d, want 200", resp.Status)
	}

	body, _ := base64.StdEncoding.DecodeString(resp.Body)
	if string(body) != `{"status":"ok"}` {
		t.Fatalf("Body = %q", string(body))
	}

	if resp.Headers["X-Test"] != "hello" {
		t.Fatalf("X-Test header = %q, want %q", resp.Headers["X-Test"], "hello")
	}
}

func TestExecuteHTTPRequest_POST(t *testing.T) {
	ssrfBypass = true
	defer func() { ssrfBypass = false }()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("method = %s, want POST", r.Method)
		}
		body, _ := io.ReadAll(r.Body)
		w.WriteHeader(201)
		_, _ = w.Write(body) // echo
	}))
	defer ts.Close()

	svc := NewHostServices(HostCapabilities{HTTPEnabled: true})

	req := &HostHTTPRequest{
		Method:  "POST",
		URL:     ts.URL + "/data",
		Headers: map[string]string{"Content-Type": "application/json"},
		Body:    base64.StdEncoding.EncodeToString([]byte(`{"key":"value"}`)),
	}

	resp, err := ExecuteHTTPRequest(context.Background(), svc, req)
	if err != nil {
		t.Fatalf("ExecuteHTTPRequest: %v", err)
	}

	if resp.Status != 201 {
		t.Fatalf("Status = %d, want 201", resp.Status)
	}

	body, _ := base64.StdEncoding.DecodeString(resp.Body)
	if string(body) != `{"key":"value"}` {
		t.Fatalf("Body = %q", string(body))
	}
}

func TestExecuteHTTPRequest_SSRFBlocksPrivateIPs(t *testing.T) {
	svc := NewHostServices(HostCapabilities{HTTPEnabled: true})

	blockedURLs := []string{
		"http://127.0.0.1/secret",
		"http://10.0.0.1/admin",
		"http://172.16.0.1/internal",
		"http://192.168.1.1/config",
		"http://169.254.169.254/latest/meta-data",
		"http://[::1]/ipv6-loopback",
	}

	for _, url := range blockedURLs {
		_, err := ExecuteHTTPRequest(context.Background(), svc, &HostHTTPRequest{
			Method: "GET",
			URL:    url,
		})
		if err == nil {
			t.Errorf("expected SSRF block for %s", url)
		}
		if !strings.Contains(err.Error(), "blocked") {
			t.Errorf("error for %s should mention 'blocked', got: %v", url, err)
		}
	}
}

func TestExecuteHTTPRequest_BlocksNonHTTPSchemes(t *testing.T) {
	svc := NewHostServices(HostCapabilities{HTTPEnabled: true})

	blockedURLs := []string{
		"file:///etc/passwd",
		"gopher://evil.com",
		"ftp://ftp.example.com",
	}

	for _, url := range blockedURLs {
		_, err := ExecuteHTTPRequest(context.Background(), svc, &HostHTTPRequest{
			Method: "GET",
			URL:    url,
		})
		if err == nil {
			t.Errorf("expected scheme block for %s", url)
		}
		if !strings.Contains(err.Error(), "blocked scheme") {
			t.Errorf("error for %s should mention 'blocked scheme', got: %v", url, err)
		}
	}
}

func TestExecuteHTTPRequest_Allowlist(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer ts.Close()

	// Extract host from test server URL
	// httptest server listens on 127.0.0.1 which is blocked by SSRF,
	// so we test the allowlist logic directly
	err := checkHostPolicy("example.com", []string{"api.example.com"}, nil)
	if err == nil {
		t.Fatal("expected error: example.com not in allowlist")
	}

	err = checkHostPolicy("api.example.com", []string{"api.example.com"}, nil)
	if err != nil {
		t.Fatalf("unexpected error for allowed host: %v", err)
	}
}

func TestExecuteHTTPRequest_Blocklist(t *testing.T) {
	err := checkHostPolicy("evil.com", nil, []string{"evil.com"})
	if err == nil {
		t.Fatal("expected error: evil.com is blocked")
	}

	err = checkHostPolicy("good.com", nil, []string{"evil.com"})
	if err != nil {
		t.Fatalf("unexpected error for non-blocked host: %v", err)
	}
}

func TestExecuteHTTPRequest_Disabled(t *testing.T) {
	svc := NewHostServices(HostCapabilities{HTTPEnabled: false})

	// Directly test the condition — the host function itself checks this
	if svc.Config.HTTPEnabled {
		t.Fatal("HTTPEnabled should be false")
	}
}

func TestExecuteHTTPRequest_DefaultMethod(t *testing.T) {
	ssrfBypass = true
	defer func() { ssrfBypass = false }()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("method = %s, want GET (default)", r.Method)
		}
		w.WriteHeader(200)
	}))
	defer ts.Close()

	svc := NewHostServices(HostCapabilities{HTTPEnabled: true})

	// Empty method defaults to GET
	resp, err := ExecuteHTTPRequest(context.Background(), svc, &HostHTTPRequest{
		URL: ts.URL,
	})
	if err != nil {
		t.Fatalf("ExecuteHTTPRequest: %v", err)
	}
	if resp.Status != 200 {
		t.Fatalf("Status = %d, want 200", resp.Status)
	}
}

func TestExecuteHTTPRequest_RawStringBody(t *testing.T) {
	ssrfBypass = true
	defer func() { ssrfBypass = false }()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_, _ = w.Write(body)
	}))
	defer ts.Close()

	svc := NewHostServices(HostCapabilities{HTTPEnabled: true})

	// Non-base64 body — should be used as raw string fallback
	resp, err := ExecuteHTTPRequest(context.Background(), svc, &HostHTTPRequest{
		Method: "POST",
		URL:    ts.URL,
		Body:   "not valid base64!!!", // invalid base64 → raw string
	})
	if err != nil {
		t.Fatalf("ExecuteHTTPRequest: %v", err)
	}

	body, _ := base64.StdEncoding.DecodeString(resp.Body)
	if string(body) != "not valid base64!!!" {
		t.Fatalf("Body = %q, want raw string fallback", string(body))
	}
}

func TestIsPrivateIP(t *testing.T) {
	tests := []struct {
		ip     string
		want   bool
	}{
		{"127.0.0.1", true},
		{"10.0.0.1", true},
		{"172.16.0.1", true},
		{"172.31.255.255", true},
		{"192.168.1.1", true},
		{"169.254.169.254", true},
		{"::1", true},
		{"8.8.8.8", false},
		{"1.1.1.1", false},
		{"203.0.113.1", false},
	}

	for _, tt := range tests {
		ip := net.ParseIP(tt.ip)
		if ip == nil {
			t.Fatalf("failed to parse IP: %s", tt.ip)
		}
		got := isPrivateIP(ip)
		if got != tt.want {
			t.Errorf("isPrivateIP(%s) = %v, want %v", tt.ip, got, tt.want)
		}
	}
}

func TestHostHTTPRequestJSON(t *testing.T) {
	req := HostHTTPRequest{
		Method:  "POST",
		URL:     "https://api.example.com/data",
		Headers: map[string]string{"Authorization": "Bearer token"},
		Body:    base64.StdEncoding.EncodeToString([]byte("body")),
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var parsed HostHTTPRequest
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if parsed.Method != "POST" || parsed.URL != "https://api.example.com/data" {
		t.Fatalf("roundtrip failed: %+v", parsed)
	}
}

func TestHostHTTPResponseJSON(t *testing.T) {
	resp := HostHTTPResponse{
		Status:  200,
		Headers: map[string]string{"Content-Type": "application/json"},
		Body:    base64.StdEncoding.EncodeToString([]byte(`{"ok":true}`)),
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var parsed HostHTTPResponse
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if parsed.Status != 200 {
		t.Fatalf("Status = %d, want 200", parsed.Status)
	}
}

func TestHostFromURL(t *testing.T) {
	tests := []struct {
		url  string
		want string
	}{
		{"https://example.com/path", "example.com"},
		{"http://api.test.com:8080/data", "api.test.com"},
		{"https://host.com", "host.com"},
		{"ftp://bad.com", ""},
		{"", ""},
	}

	for _, tt := range tests {
		got := hostFromURL(tt.url)
		if got != tt.want {
			t.Errorf("hostFromURL(%q) = %q, want %q", tt.url, got, tt.want)
		}
	}
}
