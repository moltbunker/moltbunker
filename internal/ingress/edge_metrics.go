package ingress

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// EdgeMetrics holds the edge-specific Prometheus series for the HTTP ingress.
// All series are registered into the *prometheus.Registry handed in (typically
// the one returned by metrics.PrometheusCollector.Registry()), so they are
// served by the existing /v1/metrics scrape endpoint with no new HTTP surface.
//
// Series live under namespace "moltbunker", subsystem "edge". Labels are kept
// low-cardinality: subdomain is bounded by the active tunnel registry, status
// is bucketed into a status class (2xx/3xx/4xx/5xx), and the WAF rule_id is the
// CRS integer rule ID as a string.
type EdgeMetrics struct {
	requestTotal      *prometheus.CounterVec
	requestDuration   *prometheus.HistogramVec
	requestBodyBytes  *prometheus.HistogramVec
	responseBodyBytes *prometheus.HistogramVec

	wafInspectTotal   *prometheus.CounterVec
	wafRuleMatchTotal *prometheus.CounterVec
	rateLimitTotal    *prometheus.CounterVec

	tunnelSessionsActive *prometheus.GaugeVec
}

// WAF verdict label values.
const (
	wafVerdictPass      = "pass"
	wafVerdictBlock     = "block"
	wafVerdictDetect    = "detect"
	wafVerdictTruncated = "truncated"
)

// Rate-limit reason label values.
const (
	rateLimitReasonRPS         = "rps"
	rateLimitReasonConcurrency = "concurrency"
	rateLimitReasonBodySize    = "body_size"
)

// byteBuckets is a coarse exponential bucket set spanning empty bodies up to a
// few MB, suitable for both request and response sizes.
var byteBuckets = []float64{
	0, 256, 1024, 4096, 16384, 65536, 262144, 1048576, 4194304, 16777216,
}

// NewEdgeMetrics builds and registers all edge series into reg. If reg is nil a
// fresh private registry is created (useful for tests / standalone use).
func NewEdgeMetrics(reg *prometheus.Registry) *EdgeMetrics {
	if reg == nil {
		reg = prometheus.NewRegistry()
	}

	m := &EdgeMetrics{
		requestTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "moltbunker", Subsystem: "edge",
			Name: "ingress_request_total",
			Help: "Total HTTP ingress requests by subdomain, method and status class.",
		}, []string{"subdomain", "method", "status_class"}),

		requestDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "moltbunker", Subsystem: "edge",
			Name:    "ingress_request_duration_seconds",
			Help:    "HTTP ingress request latency by subdomain (matched_waf=1 when a WAF rule fired).",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0},
		}, []string{"subdomain", "matched_waf"}),

		requestBodyBytes: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "moltbunker", Subsystem: "edge",
			Name:    "ingress_request_body_bytes",
			Help:    "HTTP ingress request body size in bytes by subdomain.",
			Buckets: byteBuckets,
		}, []string{"subdomain"}),

		responseBodyBytes: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "moltbunker", Subsystem: "edge",
			Name:    "ingress_response_body_bytes",
			Help:    "HTTP ingress response body size in bytes by subdomain.",
			Buckets: byteBuckets,
		}, []string{"subdomain"}),

		wafInspectTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "moltbunker", Subsystem: "edge",
			Name: "waf_inspect_total",
			Help: "WAF inspections by subdomain and verdict (pass|block|detect|truncated).",
		}, []string{"subdomain", "verdict"}),

		wafRuleMatchTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "moltbunker", Subsystem: "edge",
			Name: "waf_rule_match_total",
			Help: "WAF rule matches by subdomain, CRS rule_id and phase.",
		}, []string{"subdomain", "rule_id", "phase"}),

		rateLimitTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "moltbunker", Subsystem: "edge",
			Name: "ingress_rate_limit_total",
			Help: "Ingress requests rejected by abuse controls, by subdomain and reason (rps|concurrency|body_size).",
		}, []string{"subdomain", "reason"}),

		tunnelSessionsActive: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "moltbunker", Subsystem: "edge",
			Name: "tunnel_sessions_active",
			Help: "Active tunnel sessions by subdomain, tunnel_type (forward|reverse) and tier.",
		}, []string{"subdomain", "tunnel_type", "tier"}),
	}

	reg.MustRegister(
		m.requestTotal,
		m.requestDuration,
		m.requestBodyBytes,
		m.responseBodyBytes,
		m.wafInspectTotal,
		m.wafRuleMatchTotal,
		m.rateLimitTotal,
		m.tunnelSessionsActive,
	)
	return m
}

// statusClass buckets an HTTP status code into a low-cardinality class label.
func statusClass(code int) string {
	switch {
	case code >= 500:
		return "5xx"
	case code >= 400:
		return "4xx"
	case code >= 300:
		return "3xx"
	case code >= 200:
		return "2xx"
	default:
		return "1xx"
	}
}

// RecordRequest records a completed ingress request's count, latency, and
// request/response body sizes. matchedWAF indicates whether any WAF rule fired.
func (m *EdgeMetrics) RecordRequest(subdomain, method string, statusCode int, dur time.Duration, reqBytes, respBytes int64, matchedWAF bool) {
	if m == nil {
		return
	}
	m.requestTotal.WithLabelValues(subdomain, method, statusClass(statusCode)).Inc()
	m.requestDuration.WithLabelValues(subdomain, boolLabel(matchedWAF)).Observe(dur.Seconds())
	m.requestBodyBytes.WithLabelValues(subdomain).Observe(float64(reqBytes))
	m.responseBodyBytes.WithLabelValues(subdomain).Observe(float64(respBytes))
}

// RecordWAFVerdict records the overall verdict of a WAF inspection.
func (m *EdgeMetrics) RecordWAFVerdict(subdomain, verdict string) {
	if m == nil {
		return
	}
	m.wafInspectTotal.WithLabelValues(subdomain, verdict).Inc()
}

// RecordWAFRuleMatch records a single matched CRS rule.
func (m *EdgeMetrics) RecordWAFRuleMatch(subdomain, ruleID string, phase int) {
	if m == nil {
		return
	}
	m.wafRuleMatchTotal.WithLabelValues(subdomain, ruleID, strconv.Itoa(phase)).Inc()
}

// RecordRateLimit records an abuse-control rejection.
func (m *EdgeMetrics) RecordRateLimit(subdomain, reason string) {
	if m == nil {
		return
	}
	m.rateLimitTotal.WithLabelValues(subdomain, reason).Inc()
}

// SetTunnelSession adjusts the active-tunnel gauge for a subdomain.
func (m *EdgeMetrics) SetTunnelSession(subdomain, tunnelType, tier string, active bool) {
	if m == nil {
		return
	}
	g := m.tunnelSessionsActive.WithLabelValues(subdomain, tunnelType, tier)
	if active {
		g.Inc()
	} else {
		g.Dec()
	}
}

func boolLabel(b bool) string {
	if b {
		return "1"
	}
	return "0"
}
