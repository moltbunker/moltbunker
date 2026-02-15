package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/moltbunker/moltbunker/internal/metrics"
)

func TestHandleMetrics_BasicOutput(t *testing.T) {
	cfg := DefaultServerConfig()
	cfg.EnableAuth = false
	s := NewServer(cfg)

	collector := metrics.NewCollector()
	s.SetMetricsCollector(collector)

	req := httptest.NewRequest(http.MethodGet, "/v1/metrics", nil)
	rec := httptest.NewRecorder()

	s.handleMetrics(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	ct := rec.Header().Get("Content-Type")
	if ct != "text/plain; version=0.0.4; charset=utf-8" {
		t.Errorf("unexpected Content-Type: %s", ct)
	}

	body := rec.Body.String()

	// Verify Go runtime metrics are present
	for _, metric := range []string{
		"go_goroutines",
		"go_memstats_alloc_bytes",
		"go_memstats_sys_bytes",
		"go_memstats_heap_alloc_bytes",
		"go_gc_cycles_total",
	} {
		if !strings.Contains(body, metric) {
			t.Errorf("expected body to contain %q", metric)
		}
	}

	// Verify application metrics are present
	for _, metric := range []string{
		"moltbunker_uptime_seconds",
		"moltbunker_active_connections",
		"moltbunker_container_count",
		"moltbunker_peer_count",
	} {
		if !strings.Contains(body, metric) {
			t.Errorf("expected body to contain %q", metric)
		}
	}
}

func TestHandleMetrics_WithCollectorData(t *testing.T) {
	cfg := DefaultServerConfig()
	cfg.EnableAuth = false
	s := NewServer(cfg)

	collector := metrics.NewCollector()
	collector.RecordRequest("deploy")
	collector.RecordRequest("deploy")
	collector.RecordRequest("status")
	collector.RecordLatency("deploy", 15*time.Millisecond)
	collector.RecordLatency("deploy", 25*time.Millisecond)
	collector.SetContainerCount(5)
	collector.SetPeerCount(12)
	collector.IncrementConnections()
	collector.IncrementConnections()
	s.SetMetricsCollector(collector)

	req := httptest.NewRequest(http.MethodGet, "/v1/metrics", nil)
	rec := httptest.NewRecorder()

	s.handleMetrics(rec, req)

	body := rec.Body.String()

	// Check request counts
	if !strings.Contains(body, `moltbunker_requests_total{method="deploy"} 2`) {
		t.Errorf("expected deploy request count of 2 in body:\n%s", body)
	}
	if !strings.Contains(body, `moltbunker_requests_total{method="status"} 1`) {
		t.Errorf("expected status request count of 1 in body:\n%s", body)
	}

	// Check gauges
	if !strings.Contains(body, "moltbunker_container_count 5") {
		t.Errorf("expected container_count 5 in body:\n%s", body)
	}
	if !strings.Contains(body, "moltbunker_peer_count 12") {
		t.Errorf("expected peer_count 12 in body:\n%s", body)
	}
	if !strings.Contains(body, "moltbunker_active_connections 2") {
		t.Errorf("expected active_connections 2 in body:\n%s", body)
	}

	// Check histogram entries exist
	if !strings.Contains(body, "moltbunker_request_duration_seconds_bucket") {
		t.Errorf("expected histogram buckets in body:\n%s", body)
	}
	if !strings.Contains(body, "moltbunker_request_duration_seconds_sum") {
		t.Errorf("expected histogram sum in body:\n%s", body)
	}
	if !strings.Contains(body, "moltbunker_request_duration_seconds_count") {
		t.Errorf("expected histogram count in body:\n%s", body)
	}
}

func TestHandleMetrics_NoCollector(t *testing.T) {
	cfg := DefaultServerConfig()
	cfg.EnableAuth = false
	s := NewServer(cfg)
	// Do not set a metrics collector

	req := httptest.NewRequest(http.MethodGet, "/v1/metrics", nil)
	rec := httptest.NewRecorder()

	s.handleMetrics(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	body := rec.Body.String()

	// Should still have Go runtime metrics
	if !strings.Contains(body, "go_goroutines") {
		t.Error("expected Go runtime metrics even without collector")
	}

	// Should NOT have application metrics
	if strings.Contains(body, "moltbunker_uptime_seconds") {
		t.Error("did not expect application metrics without collector")
	}
}

func TestHandleMetrics_MethodNotAllowed(t *testing.T) {
	cfg := DefaultServerConfig()
	cfg.EnableAuth = false
	s := NewServer(cfg)

	req := httptest.NewRequest(http.MethodPost, "/v1/metrics", nil)
	rec := httptest.NewRecorder()

	s.handleMetrics(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status 405, got %d", rec.Code)
	}
}

func TestHandleMetrics_PrometheusFormat(t *testing.T) {
	cfg := DefaultServerConfig()
	cfg.EnableAuth = false
	s := NewServer(cfg)

	collector := metrics.NewCollector()
	collector.RecordRequest("test")
	s.SetMetricsCollector(collector)

	req := httptest.NewRequest(http.MethodGet, "/v1/metrics", nil)
	rec := httptest.NewRecorder()

	s.handleMetrics(rec, req)

	body := rec.Body.String()

	// Verify Prometheus text format: every metric should have HELP and TYPE
	lines := strings.Split(body, "\n")
	helpCount := 0
	typeCount := 0
	for _, line := range lines {
		if strings.HasPrefix(line, "# HELP ") {
			helpCount++
		}
		if strings.HasPrefix(line, "# TYPE ") {
			typeCount++
		}
	}

	if helpCount == 0 {
		t.Error("expected at least one HELP line")
	}
	if typeCount == 0 {
		t.Error("expected at least one TYPE line")
	}
	if helpCount != typeCount {
		t.Errorf("HELP count (%d) should match TYPE count (%d)", helpCount, typeCount)
	}
}
