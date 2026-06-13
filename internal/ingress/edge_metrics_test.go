package ingress

import (
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestEdgeMetrics_RecordRequest_CounterIncrements(t *testing.T) {
	reg := prometheus.NewRegistry()
	em := NewEdgeMetrics(reg)

	em.RecordRequest("acme", "GET", 200, 50*time.Millisecond, 128, 4096, false)
	em.RecordRequest("acme", "GET", 200, 10*time.Millisecond, 0, 100, false)
	em.RecordRequest("acme", "POST", 500, 5*time.Millisecond, 64, 0, true)

	if got := testutil.ToFloat64(em.requestTotal.WithLabelValues("acme", "GET", "2xx")); got != 2 {
		t.Fatalf("GET 2xx counter: want 2, got %v", got)
	}
	if got := testutil.ToFloat64(em.requestTotal.WithLabelValues("acme", "POST", "5xx")); got != 1 {
		t.Fatalf("POST 5xx counter: want 1, got %v", got)
	}
}

func TestEdgeMetrics_RecordWAF_BlockCounter(t *testing.T) {
	reg := prometheus.NewRegistry()
	em := NewEdgeMetrics(reg)

	em.RecordWAFVerdict("acme", wafVerdictBlock)
	em.RecordWAFVerdict("acme", wafVerdictBlock)
	em.RecordWAFVerdict("acme", wafVerdictPass)
	em.RecordWAFRuleMatch("acme", "942100", 2)

	if got := testutil.ToFloat64(em.wafInspectTotal.WithLabelValues("acme", wafVerdictBlock)); got != 2 {
		t.Fatalf("waf block counter: want 2, got %v", got)
	}
	if got := testutil.ToFloat64(em.wafInspectTotal.WithLabelValues("acme", wafVerdictPass)); got != 1 {
		t.Fatalf("waf pass counter: want 1, got %v", got)
	}
	if got := testutil.ToFloat64(em.wafRuleMatchTotal.WithLabelValues("acme", "942100", "2")); got != 1 {
		t.Fatalf("waf rule match counter: want 1, got %v", got)
	}
}

func TestEdgeMetrics_RecordRateLimit(t *testing.T) {
	reg := prometheus.NewRegistry()
	em := NewEdgeMetrics(reg)

	em.RecordRateLimit("acme", rateLimitReasonRPS)
	em.RecordRateLimit("acme", rateLimitReasonConcurrency)
	em.RecordRateLimit("acme", rateLimitReasonRPS)

	if got := testutil.ToFloat64(em.rateLimitTotal.WithLabelValues("acme", rateLimitReasonRPS)); got != 2 {
		t.Fatalf("rate limit rps counter: want 2, got %v", got)
	}
	if got := testutil.ToFloat64(em.rateLimitTotal.WithLabelValues("acme", rateLimitReasonConcurrency)); got != 1 {
		t.Fatalf("rate limit concurrency counter: want 1, got %v", got)
	}
}

func TestEdgeMetrics_SetTunnelSession_GaugeUpDown(t *testing.T) {
	reg := prometheus.NewRegistry()
	em := NewEdgeMetrics(reg)

	em.SetTunnelSession("acme", "reverse", "starter", true)
	em.SetTunnelSession("acme", "reverse", "starter", true)
	if got := testutil.ToFloat64(em.tunnelSessionsActive.WithLabelValues("acme", "reverse", "starter")); got != 2 {
		t.Fatalf("tunnel gauge after 2 up: want 2, got %v", got)
	}
	em.SetTunnelSession("acme", "reverse", "starter", false)
	if got := testutil.ToFloat64(em.tunnelSessionsActive.WithLabelValues("acme", "reverse", "starter")); got != 1 {
		t.Fatalf("tunnel gauge after down: want 1, got %v", got)
	}
}

func TestEdgeMetrics_NilSafe(t *testing.T) {
	var em *EdgeMetrics // nil
	// None of these should panic.
	em.RecordRequest("x", "GET", 200, time.Millisecond, 1, 1, false)
	em.RecordWAFVerdict("x", wafVerdictPass)
	em.RecordWAFRuleMatch("x", "1", 1)
	em.RecordRateLimit("x", rateLimitReasonRPS)
	em.SetTunnelSession("x", "forward", "pro", true)
}

func TestEdgeMetrics_RegistersIntoSharedRegistry(t *testing.T) {
	reg := prometheus.NewRegistry()
	em := NewEdgeMetrics(reg)
	// CounterVec/HistogramVec only emit a metric family once a labeled child
	// has been observed, so touch one of each before gathering.
	em.RecordRequest("acme", "GET", 200, time.Millisecond, 1, 1, false)
	em.RecordWAFVerdict("acme", wafVerdictPass)
	em.RecordRateLimit("acme", rateLimitReasonRPS)
	em.SetTunnelSession("acme", "forward", "starter", true)

	mfs, err := reg.Gather()
	if err != nil {
		t.Fatalf("Gather: %v", err)
	}
	if len(mfs) == 0 {
		t.Fatal("expected edge metric families registered in the shared registry")
	}
	// Confirm the series live under the expected namespace/subsystem.
	var found bool
	for _, mf := range mfs {
		if strings.HasPrefix(mf.GetName(), "moltbunker_edge_") {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected metric families under moltbunker_edge_ prefix")
	}
}
