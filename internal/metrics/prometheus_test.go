package metrics

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

func TestNewPrometheusCollector(t *testing.T) {
	c := NewCollector()
	pc := NewPrometheusCollector(c)

	if pc == nil {
		t.Fatal("NewPrometheusCollector returned nil")
	}
	if pc.collector != c {
		t.Error("expected PrometheusCollector to wrap the given Collector")
	}
	if pc.registry == nil {
		t.Error("expected non-nil Prometheus registry")
	}
}

func TestPrometheusRecordRequest(t *testing.T) {
	c := NewCollector()
	pc := NewPrometheusCollector(c)

	pc.RecordRequest("GET")
	pc.RecordRequest("GET")
	pc.RecordRequest("POST")

	// Verify custom collector still works
	m := c.GetMetrics()
	if m.RequestCounts["GET"] != 2 {
		t.Errorf("expected custom collector GET count 2, got %d", m.RequestCounts["GET"])
	}
	if m.RequestCounts["POST"] != 1 {
		t.Errorf("expected custom collector POST count 1, got %d", m.RequestCounts["POST"])
	}

	// Verify Prometheus counter
	getCounter := getCounterValue(t, pc.requestCount, "GET")
	if getCounter != 2 {
		t.Errorf("expected Prometheus GET counter 2, got %f", getCounter)
	}

	postCounter := getCounterValue(t, pc.requestCount, "POST")
	if postCounter != 1 {
		t.Errorf("expected Prometheus POST counter 1, got %f", postCounter)
	}
}

func TestPrometheusRecordLatency(t *testing.T) {
	c := NewCollector()
	pc := NewPrometheusCollector(c)

	pc.RecordLatency("deploy", 10*time.Millisecond)
	pc.RecordLatency("deploy", 50*time.Millisecond)

	// Verify custom collector
	m := c.GetMetrics()
	stats, ok := m.RequestLatencies["deploy"]
	if !ok {
		t.Fatal("expected deploy latency stats in custom collector")
	}
	if stats.Count != 2 {
		t.Errorf("expected count 2, got %d", stats.Count)
	}

	// Verify Prometheus histogram
	observer := pc.requestDuration.WithLabelValues("deploy")
	metric := &dto.Metric{}
	if err := observer.(prometheus.Metric).Write(metric); err != nil {
		t.Fatalf("failed to read prometheus metric: %v", err)
	}
	hist := metric.GetHistogram()
	if hist == nil {
		t.Fatal("expected histogram metric")
	}
	if hist.GetSampleCount() != 2 {
		t.Errorf("expected histogram sample count 2, got %d", hist.GetSampleCount())
	}
	// Sum should be approximately 0.060 seconds (10ms + 50ms)
	if hist.GetSampleSum() < 0.05 || hist.GetSampleSum() > 0.07 {
		t.Errorf("expected histogram sum ~0.060, got %f", hist.GetSampleSum())
	}
}

func TestPrometheusConnections(t *testing.T) {
	c := NewCollector()
	pc := NewPrometheusCollector(c)

	pc.IncrementConnections()
	pc.IncrementConnections()
	pc.IncrementConnections()
	pc.DecrementConnections()

	// Verify custom collector
	m := c.GetMetrics()
	if m.ActiveConnections != 2 {
		t.Errorf("expected 2 active connections in custom collector, got %d", m.ActiveConnections)
	}

	// Verify Prometheus gauge
	gaugeVal := getGaugeValue(t, pc.activeConnections)
	if gaugeVal != 2 {
		t.Errorf("expected Prometheus active connections gauge 2, got %f", gaugeVal)
	}
}

func TestPrometheusSetContainerCount(t *testing.T) {
	c := NewCollector()
	pc := NewPrometheusCollector(c)

	pc.SetContainerCount(5)

	m := c.GetMetrics()
	if m.ContainerCount != 5 {
		t.Errorf("expected custom collector container count 5, got %d", m.ContainerCount)
	}

	gaugeVal := getGaugeValue(t, pc.containerCount)
	if gaugeVal != 5 {
		t.Errorf("expected Prometheus container count gauge 5, got %f", gaugeVal)
	}
}

func TestPrometheusSetPeerCount(t *testing.T) {
	c := NewCollector()
	pc := NewPrometheusCollector(c)

	pc.SetPeerCount(12)

	m := c.GetMetrics()
	if m.PeerCount != 12 {
		t.Errorf("expected custom collector peer count 12, got %d", m.PeerCount)
	}

	gaugeVal := getGaugeValue(t, pc.peerCount)
	if gaugeVal != 12 {
		t.Errorf("expected Prometheus peer count gauge 12, got %f", gaugeVal)
	}
}

func TestPrometheusUpdateGoroutineCount(t *testing.T) {
	c := NewCollector()
	pc := NewPrometheusCollector(c)

	pc.UpdateGoroutineCount()

	m := c.GetMetrics()
	if m.GoroutineCount <= 0 {
		t.Errorf("expected positive goroutine count, got %d", m.GoroutineCount)
	}

	gaugeVal := getGaugeValue(t, pc.goroutineCount)
	if gaugeVal <= 0 {
		t.Errorf("expected positive Prometheus goroutine gauge, got %f", gaugeVal)
	}
}

func TestPrometheusSync(t *testing.T) {
	c := NewCollector()
	pc := NewPrometheusCollector(c)

	// Modify the underlying collector directly (simulating external changes)
	c.SetContainerCount(3)
	c.SetPeerCount(7)
	c.IncrementConnections()

	// Sync should pull values from the collector into Prometheus gauges
	pc.Sync()

	if v := getGaugeValue(t, pc.containerCount); v != 3 {
		t.Errorf("expected synced container count 3, got %f", v)
	}
	if v := getGaugeValue(t, pc.peerCount); v != 7 {
		t.Errorf("expected synced peer count 7, got %f", v)
	}
	if v := getGaugeValue(t, pc.activeConnections); v != 1 {
		t.Errorf("expected synced active connections 1, got %f", v)
	}
}

func TestPrometheusSyncRequestCounts(t *testing.T) {
	c := NewCollector()
	pc := NewPrometheusCollector(c)

	// Record requests via the underlying collector only
	c.RecordRequest("status")
	c.RecordRequest("status")
	c.RecordRequest("deploy")

	// Sync should detect these and increment counters
	pc.Sync()

	statusVal := getCounterValue(t, pc.requestCount, "status")
	if statusVal != 2 {
		t.Errorf("expected synced status counter 2, got %f", statusVal)
	}

	deployVal := getCounterValue(t, pc.requestCount, "deploy")
	if deployVal != 1 {
		t.Errorf("expected synced deploy counter 1, got %f", deployVal)
	}

	// Sync again with more records - should only add the delta
	c.RecordRequest("status")
	pc.Sync()

	statusVal = getCounterValue(t, pc.requestCount, "status")
	if statusVal != 3 {
		t.Errorf("expected synced status counter 3 after second sync, got %f", statusVal)
	}
}

func TestPrometheusHandler(t *testing.T) {
	c := NewCollector()
	pc := NewPrometheusCollector(c)

	pc.RecordRequest("deploy")
	pc.RecordLatency("deploy", 25*time.Millisecond)
	pc.SetContainerCount(2)
	pc.SetPeerCount(5)
	pc.IncrementConnections()

	handler := pc.PrometheusHandler()

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	body, err := io.ReadAll(rr.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}
	bodyStr := string(body)

	// Verify key metrics are present in the output
	expectedMetrics := []string{
		"moltbunker_request_count",
		"moltbunker_request_duration_seconds",
		"moltbunker_active_connections",
		"moltbunker_container_count",
		"moltbunker_peer_count",
		"moltbunker_goroutine_count",
		"moltbunker_uptime_seconds",
	}

	for _, name := range expectedMetrics {
		if !strings.Contains(bodyStr, name) {
			t.Errorf("expected metric %q in Prometheus output, not found", name)
		}
	}

	// Verify method label is present
	if !strings.Contains(bodyStr, `method="deploy"`) {
		t.Error("expected method label 'deploy' in Prometheus output")
	}
}

func TestPrometheusHandlerContentType(t *testing.T) {
	c := NewCollector()
	pc := NewPrometheusCollector(c)
	handler := pc.PrometheusHandler()

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	ct := rr.Header().Get("Content-Type")
	if !strings.Contains(ct, "text/plain") {
		t.Errorf("expected Content-Type containing text/plain, got %q", ct)
	}
}

func TestPrometheusGetMetrics(t *testing.T) {
	c := NewCollector()
	pc := NewPrometheusCollector(c)

	pc.RecordRequest("test")
	pc.SetContainerCount(4)

	m := pc.GetMetrics()
	if m.ContainerCount != 4 {
		t.Errorf("expected container count 4, got %d", m.ContainerCount)
	}
	if m.RequestCounts["test"] != 1 {
		t.Errorf("expected test count 1, got %d", m.RequestCounts["test"])
	}
}

func TestPrometheusGetMetricsJSON(t *testing.T) {
	c := NewCollector()
	pc := NewPrometheusCollector(c)

	pc.SetPeerCount(3)

	data, err := pc.GetMetricsJSON()
	if err != nil {
		t.Fatalf("GetMetricsJSON error: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty JSON")
	}
}

func TestPrometheusCollectorReturnsUnderlying(t *testing.T) {
	c := NewCollector()
	pc := NewPrometheusCollector(c)

	if pc.Collector() != c {
		t.Error("Collector() should return the underlying collector")
	}
}

func TestPrometheusRegistry(t *testing.T) {
	c := NewCollector()
	pc := NewPrometheusCollector(c)

	reg := pc.Registry()
	if reg == nil {
		t.Fatal("Registry() returned nil")
	}

	// Verify we can gather metrics from the registry
	families, err := reg.Gather()
	if err != nil {
		t.Fatalf("failed to gather metrics: %v", err)
	}

	// Should have at least the registered metrics
	if len(families) == 0 {
		t.Error("expected at least one metric family from registry")
	}
}

func TestPrometheusUptimeIncreases(t *testing.T) {
	c := NewCollector()
	pc := NewPrometheusCollector(c)

	time.Sleep(10 * time.Millisecond)
	pc.Sync()

	uptime := getGaugeValue(t, pc.uptimeSeconds)
	if uptime < 0.01 {
		t.Errorf("expected measurable uptime, got %f", uptime)
	}
}

// getCounterValue extracts the current counter value for a given label from a CounterVec.
func getCounterValue(t *testing.T, cv *prometheus.CounterVec, label string) float64 {
	t.Helper()
	counter := cv.WithLabelValues(label)
	metric := &dto.Metric{}
	if err := counter.(prometheus.Metric).Write(metric); err != nil {
		t.Fatalf("failed to read counter metric: %v", err)
	}
	return metric.GetCounter().GetValue()
}

// getGaugeValue extracts the current value from a Prometheus Gauge.
func getGaugeValue(t *testing.T, g prometheus.Gauge) float64 {
	t.Helper()
	metric := &dto.Metric{}
	if err := g.Write(metric); err != nil {
		t.Fatalf("failed to read gauge metric: %v", err)
	}
	return metric.GetGauge().GetValue()
}
