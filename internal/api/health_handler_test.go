package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/moltbunker/moltbunker/internal/metrics"
)

func TestHandleHealthCheck_Healthy(t *testing.T) {
	cfg := DefaultServerConfig()
	cfg.EnableAuth = false
	cfg.DaemonSocketPath = "" // No daemon bridge for this test
	s := NewServer(cfg)

	// Mark server as running
	s.mu.Lock()
	s.running = true
	s.mu.Unlock()

	collector := metrics.NewCollector()
	collector.SetPeerCount(5)
	s.SetMetricsCollector(collector)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	s.handleHealthCheck(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var resp HealthResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Status != "healthy" {
		t.Errorf("expected status 'healthy', got %q", resp.Status)
	}
	if resp.Version != "1.0.0" {
		t.Errorf("expected version '1.0.0', got %q", resp.Version)
	}
	if resp.PeerCount != 5 {
		t.Errorf("expected peer_count 5, got %d", resp.PeerCount)
	}
	if resp.Uptime == "" {
		t.Error("expected non-empty uptime")
	}
	if resp.Reason != "" {
		t.Errorf("expected no reason when healthy, got %q", resp.Reason)
	}
}

func TestHandleHealthCheck_Unhealthy_NotRunning(t *testing.T) {
	cfg := DefaultServerConfig()
	cfg.EnableAuth = false
	s := NewServer(cfg)

	// Server is not running (default running = false)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	s.handleHealthCheck(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status 503, got %d", rec.Code)
	}

	var resp HealthResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Status != "unhealthy" {
		t.Errorf("expected status 'unhealthy', got %q", resp.Status)
	}
	if resp.Reason != "server not running" {
		t.Errorf("expected reason 'server not running', got %q", resp.Reason)
	}
}

func TestHandleHealthCheck_NoMetricsCollector(t *testing.T) {
	cfg := DefaultServerConfig()
	cfg.EnableAuth = false
	cfg.DaemonSocketPath = "" // No daemon bridge for this test
	s := NewServer(cfg)

	// Mark server as running but no metrics collector
	s.mu.Lock()
	s.running = true
	s.mu.Unlock()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	s.handleHealthCheck(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var resp HealthResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Status != "healthy" {
		t.Errorf("expected status 'healthy', got %q", resp.Status)
	}
	if resp.Uptime == "" {
		t.Error("expected non-empty uptime even without metrics collector")
	}
}

func TestHandleHealthCheck_MethodNotAllowed(t *testing.T) {
	cfg := DefaultServerConfig()
	cfg.EnableAuth = false
	s := NewServer(cfg)

	req := httptest.NewRequest(http.MethodPost, "/health", nil)
	rec := httptest.NewRecorder()

	s.handleHealthCheck(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status 405, got %d", rec.Code)
	}
}

func TestHandleHealthCheck_ContentType(t *testing.T) {
	cfg := DefaultServerConfig()
	cfg.EnableAuth = false
	cfg.DaemonSocketPath = "" // No daemon bridge for this test
	s := NewServer(cfg)

	s.mu.Lock()
	s.running = true
	s.mu.Unlock()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	s.handleHealthCheck(rec, req)

	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got %q", ct)
	}
}

func TestHandleHealthCheck_JSONFields(t *testing.T) {
	cfg := DefaultServerConfig()
	cfg.EnableAuth = false
	cfg.DaemonSocketPath = "" // No daemon bridge for this test
	s := NewServer(cfg)

	s.mu.Lock()
	s.running = true
	s.mu.Unlock()

	collector := metrics.NewCollector()
	collector.SetPeerCount(10)
	s.SetMetricsCollector(collector)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	s.handleHealthCheck(rec, req)

	// Decode as raw map to verify JSON field names
	var raw map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&raw); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	requiredFields := []string{"status", "uptime", "peer_count", "version"}
	for _, field := range requiredFields {
		if _, ok := raw[field]; !ok {
			t.Errorf("expected field %q in response, got: %v", field, raw)
		}
	}

	// "reason" should not be present when healthy (omitempty)
	if _, ok := raw["reason"]; ok {
		t.Errorf("expected 'reason' to be omitted when healthy, got: %v", raw["reason"])
	}
}
