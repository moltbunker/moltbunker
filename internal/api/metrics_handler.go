package api

import (
	"fmt"
	"net/http"
	"runtime"
	"sort"
	"strings"

	"github.com/moltbunker/moltbunker/internal/metrics"
)

// SetMetricsCollector sets the metrics collector for the /v1/metrics endpoint
func (s *Server) SetMetricsCollector(collector *metrics.Collector) {
	s.metricsCollector = collector
}

// handleMetrics serves metrics in Prometheus text exposition format.
// No authentication required for metrics scraping.
func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")

	var b strings.Builder

	// Go runtime metrics
	writeRuntimeMetrics(&b)

	// Custom application metrics from the Collector
	if s.metricsCollector != nil {
		writeApplicationMetrics(&b, s.metricsCollector)
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, b.String())
}

// writeRuntimeMetrics writes Go runtime metrics in Prometheus text format.
func writeRuntimeMetrics(b *strings.Builder) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	writeGauge(b, "go_goroutines", "Number of goroutines that currently exist", float64(runtime.NumGoroutine()))
	writeGauge(b, "go_threads", "Number of OS threads created", float64(runtime.GOMAXPROCS(0)))
	writeGauge(b, "go_memstats_alloc_bytes", "Number of bytes allocated and still in use", float64(memStats.Alloc))
	writeGauge(b, "go_memstats_alloc_bytes_total", "Total number of bytes allocated, even if freed", float64(memStats.TotalAlloc))
	writeGauge(b, "go_memstats_sys_bytes", "Number of bytes obtained from system", float64(memStats.Sys))
	writeGauge(b, "go_memstats_heap_alloc_bytes", "Number of heap bytes allocated and still in use", float64(memStats.HeapAlloc))
	writeGauge(b, "go_memstats_heap_sys_bytes", "Number of heap bytes obtained from system", float64(memStats.HeapSys))
	writeGauge(b, "go_memstats_heap_idle_bytes", "Number of heap bytes waiting to be used", float64(memStats.HeapIdle))
	writeGauge(b, "go_memstats_heap_inuse_bytes", "Number of heap bytes that are in use", float64(memStats.HeapInuse))
	writeGauge(b, "go_memstats_heap_released_bytes", "Number of heap bytes released to OS", float64(memStats.HeapReleased))
	writeGauge(b, "go_memstats_heap_objects", "Number of allocated objects", float64(memStats.HeapObjects))
	writeGauge(b, "go_memstats_stack_inuse_bytes", "Number of bytes in use by the stack allocator", float64(memStats.StackInuse))
	writeGauge(b, "go_memstats_stack_sys_bytes", "Number of bytes obtained from system for stack allocator", float64(memStats.StackSys))
	writeCounter(b, "go_memstats_mallocs_total", "Total number of mallocs", float64(memStats.Mallocs))
	writeCounter(b, "go_memstats_frees_total", "Total number of frees", float64(memStats.Frees))
	writeCounter(b, "go_gc_cycles_total", "Total number of completed GC cycles", float64(memStats.NumGC))
	writeGauge(b, "go_gc_pause_seconds_last", "Duration of the last GC pause in seconds", float64(memStats.PauseNs[(memStats.NumGC+255)%256])/1e9)
}

// writeApplicationMetrics writes custom application metrics from the Collector.
func writeApplicationMetrics(b *strings.Builder, c *metrics.Collector) {
	m := c.GetMetrics()

	// Uptime
	writeGauge(b, "moltbunker_uptime_seconds", "Time since the daemon started in seconds", m.UptimeSeconds)

	// Gauges
	writeGauge(b, "moltbunker_active_connections", "Number of active P2P connections", float64(m.ActiveConnections))
	writeGauge(b, "moltbunker_container_count", "Number of running containers", float64(m.ContainerCount))
	writeGauge(b, "moltbunker_peer_count", "Number of known peers", float64(m.PeerCount))

	// Request counts (counter per method)
	if len(m.RequestCounts) > 0 {
		b.WriteString("# HELP moltbunker_requests_total Total number of requests by method\n")
		b.WriteString("# TYPE moltbunker_requests_total counter\n")
		methods := sortedKeys(m.RequestCounts)
		for _, method := range methods {
			count := m.RequestCounts[method]
			fmt.Fprintf(b, "moltbunker_requests_total{method=%q} %d\n", method, count)
		}
	}

	// Request latencies (histogram per method)
	if len(m.RequestLatencies) > 0 {
		// Bucket boundaries in seconds for Prometheus histograms
		bucketBoundariesSec := []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0}
		bucketLabelsMs := []string{
			"0-1ms", "1-5ms", "5-10ms", "10-25ms", "25-50ms",
			"50-100ms", "100-250ms", "250-500ms", "500-1000ms", "1000ms+",
		}

		b.WriteString("# HELP moltbunker_request_duration_seconds Request latency histogram by method\n")
		b.WriteString("# TYPE moltbunker_request_duration_seconds histogram\n")

		latencyMethods := make([]string, 0, len(m.RequestLatencies))
		for method := range m.RequestLatencies {
			latencyMethods = append(latencyMethods, method)
		}
		sort.Strings(latencyMethods)

		for _, method := range latencyMethods {
			stats := m.RequestLatencies[method]

			// Emit cumulative bucket counts
			var cumulative uint64
			for i, le := range bucketBoundariesSec {
				label := bucketLabelsMs[i]
				if count, ok := stats.Buckets[label]; ok {
					cumulative += count
				}
				fmt.Fprintf(b, "moltbunker_request_duration_seconds_bucket{method=%q,le=\"%.3f\"} %d\n", method, le, cumulative)
			}
			// +Inf bucket (includes overflow)
			overflowLabel := bucketLabelsMs[len(bucketLabelsMs)-1]
			if count, ok := stats.Buckets[overflowLabel]; ok {
				cumulative += count
			}
			fmt.Fprintf(b, "moltbunker_request_duration_seconds_bucket{method=%q,le=\"+Inf\"} %d\n", method, cumulative)

			// Sum and count
			fmt.Fprintf(b, "moltbunker_request_duration_seconds_sum{method=%q} %.6f\n", method, stats.SumMs/1000.0)
			fmt.Fprintf(b, "moltbunker_request_duration_seconds_count{method=%q} %d\n", method, stats.Count)
		}
	}
}

// writeGauge writes a single gauge metric in Prometheus text format.
func writeGauge(b *strings.Builder, name, help string, value float64) {
	fmt.Fprintf(b, "# HELP %s %s\n", name, help)
	fmt.Fprintf(b, "# TYPE %s gauge\n", name)
	fmt.Fprintf(b, "%s %g\n", name, value)
}

// writeCounter writes a single counter metric in Prometheus text format.
func writeCounter(b *strings.Builder, name, help string, value float64) {
	fmt.Fprintf(b, "# HELP %s %s\n", name, help)
	fmt.Fprintf(b, "# TYPE %s counter\n", name)
	fmt.Fprintf(b, "%s %g\n", name, value)
}

// sortedKeys returns the keys of a map[string]uint64 in sorted order.
func sortedKeys(m map[string]uint64) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
