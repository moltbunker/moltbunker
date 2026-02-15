package metrics

import (
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// PrometheusCollector wraps the existing Collector and mirrors its metrics
// into Prometheus format. Both the original JSON output and the Prometheus
// exposition format are supported simultaneously.
type PrometheusCollector struct {
	collector *Collector
	registry  *prometheus.Registry

	// Prometheus metrics
	requestCount    *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec

	activeConnections prometheus.Gauge
	containerCount    prometheus.Gauge
	peerCount         prometheus.Gauge
	goroutineCount    prometheus.Gauge
	uptimeSeconds     prometheus.Gauge

	startTime time.Time

	// Track per-method counters so we can compute deltas from the
	// Collector's cumulative counts.
	lastCounts   map[string]uint64
	lastCountsMu sync.Mutex
}

// NewPrometheusCollector creates a PrometheusCollector that wraps an existing
// Collector. Prometheus metrics are registered in a dedicated registry so they
// do not interfere with the default global registry.
func NewPrometheusCollector(c *Collector) *PrometheusCollector {
	reg := prometheus.NewRegistry()

	requestCount := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "moltbunker",
		Name:      "request_count",
		Help:      "Total number of requests by method.",
	}, []string{"method"})

	requestDuration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "moltbunker",
		Name:      "request_duration_seconds",
		Help:      "Request latency histogram by method.",
		Buckets:   []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0},
	}, []string{"method"})

	activeConns := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "moltbunker",
		Name:      "active_connections",
		Help:      "Number of active P2P connections.",
	})

	containerCnt := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "moltbunker",
		Name:      "container_count",
		Help:      "Number of running containers.",
	})

	peerCnt := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "moltbunker",
		Name:      "peer_count",
		Help:      "Number of known peers.",
	})

	goroutineCnt := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "moltbunker",
		Name:      "goroutine_count",
		Help:      "Number of goroutines.",
	})

	uptimeSec := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "moltbunker",
		Name:      "uptime_seconds",
		Help:      "Time since the daemon started in seconds.",
	})

	reg.MustRegister(requestCount)
	reg.MustRegister(requestDuration)
	reg.MustRegister(activeConns)
	reg.MustRegister(containerCnt)
	reg.MustRegister(peerCnt)
	reg.MustRegister(goroutineCnt)
	reg.MustRegister(uptimeSec)

	return &PrometheusCollector{
		collector:         c,
		registry:          reg,
		requestCount:      requestCount,
		requestDuration:   requestDuration,
		activeConnections: activeConns,
		containerCount:    containerCnt,
		peerCount:         peerCnt,
		goroutineCount:    goroutineCnt,
		uptimeSeconds:     uptimeSec,
		startTime:         time.Now(),
		lastCounts:        make(map[string]uint64),
	}
}

// Registry returns the Prometheus registry used by this collector.
// This is useful for registering additional metrics (e.g. libp2p rcmgr).
func (p *PrometheusCollector) Registry() *prometheus.Registry {
	return p.registry
}

// RecordRequest records a request in both the custom Collector and
// the Prometheus counter.
func (p *PrometheusCollector) RecordRequest(method string) {
	p.collector.RecordRequest(method)
	p.requestCount.WithLabelValues(method).Inc()
}

// RecordLatency records latency in both the custom Collector and
// the Prometheus histogram.
func (p *PrometheusCollector) RecordLatency(method string, duration time.Duration) {
	p.collector.RecordLatency(method, duration)
	p.requestDuration.WithLabelValues(method).Observe(duration.Seconds())
}

// IncrementConnections increments connections in both collectors.
func (p *PrometheusCollector) IncrementConnections() {
	p.collector.IncrementConnections()
	p.activeConnections.Inc()
}

// DecrementConnections decrements connections in both collectors.
func (p *PrometheusCollector) DecrementConnections() {
	p.collector.DecrementConnections()
	p.activeConnections.Dec()
}

// SetContainerCount sets the container count in both collectors.
func (p *PrometheusCollector) SetContainerCount(count int) {
	p.collector.SetContainerCount(count)
	p.containerCount.Set(float64(count))
}

// SetPeerCount sets the peer count in both collectors.
func (p *PrometheusCollector) SetPeerCount(count int) {
	p.collector.SetPeerCount(count)
	p.peerCount.Set(float64(count))
}

// UpdateGoroutineCount updates the goroutine count in both collectors.
func (p *PrometheusCollector) UpdateGoroutineCount() {
	p.collector.UpdateGoroutineCount()
	p.goroutineCount.Set(float64(runtime.NumGoroutine()))
}

// Sync synchronizes the Prometheus gauges with the current state of the
// underlying Collector. Call this periodically or before serving metrics
// to ensure gauges reflect the latest values.
func (p *PrometheusCollector) Sync() {
	m := p.collector.GetMetrics()

	p.activeConnections.Set(float64(m.ActiveConnections))
	p.containerCount.Set(float64(m.ContainerCount))
	p.peerCount.Set(float64(m.PeerCount))
	p.goroutineCount.Set(float64(m.GoroutineCount))
	p.uptimeSeconds.Set(m.UptimeSeconds)

	// Sync request counts: compute deltas and add them to counters
	p.lastCountsMu.Lock()
	for method, total := range m.RequestCounts {
		prev := p.lastCounts[method]
		if total > prev {
			p.requestCount.WithLabelValues(method).Add(float64(total - prev))
		}
		p.lastCounts[method] = total
	}
	p.lastCountsMu.Unlock()
}

// GetMetrics returns the JSON metrics from the underlying Collector.
func (p *PrometheusCollector) GetMetrics() *Metrics {
	return p.collector.GetMetrics()
}

// GetMetricsJSON returns JSON-encoded metrics from the underlying Collector.
func (p *PrometheusCollector) GetMetricsJSON() ([]byte, error) {
	return p.collector.GetMetricsJSON()
}

// Collector returns the underlying custom Collector.
func (p *PrometheusCollector) Collector() *Collector {
	return p.collector
}

// PrometheusHandler returns an http.Handler that serves metrics in the
// Prometheus text exposition format. The handler synchronizes gauge values
// from the underlying Collector before each scrape.
func (p *PrometheusCollector) PrometheusHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Sync gauges before serving
		p.Sync()
		promhttp.HandlerFor(p.registry, promhttp.HandlerOpts{}).ServeHTTP(w, r)
	})
}
