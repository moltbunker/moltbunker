package metrics

import (
	"encoding/json"
	"sync"
	"testing"
	"time"
)

func TestNewCollector(t *testing.T) {
	c := NewCollector()
	if c == nil {
		t.Fatal("NewCollector returned nil")
	}
	if c.requestCounts == nil {
		t.Error("expected initialized requestCounts map")
	}
	if c.latencies == nil {
		t.Error("expected initialized latencies map")
	}
	if c.startTime.IsZero() {
		t.Error("expected non-zero start time")
	}
}

func TestRecordRequest(t *testing.T) {
	c := NewCollector()

	c.RecordRequest("GET")
	c.RecordRequest("GET")
	c.RecordRequest("POST")

	metrics := c.GetMetrics()

	if metrics.RequestCounts["GET"] != 2 {
		t.Errorf("expected GET count 2, got %d", metrics.RequestCounts["GET"])
	}
	if metrics.RequestCounts["POST"] != 1 {
		t.Errorf("expected POST count 1, got %d", metrics.RequestCounts["POST"])
	}
}

func TestRecordRequestConcurrent(t *testing.T) {
	c := NewCollector()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.RecordRequest("GET")
		}()
	}
	wg.Wait()

	metrics := c.GetMetrics()
	if metrics.RequestCounts["GET"] != 100 {
		t.Errorf("expected GET count 100, got %d", metrics.RequestCounts["GET"])
	}
}

func TestRecordLatency(t *testing.T) {
	c := NewCollector()

	c.RecordLatency("GET", 500*time.Microsecond) // 0.5ms -> bucket [0-1ms]
	c.RecordLatency("GET", 3*time.Millisecond)   // 3ms -> bucket [1-5ms]
	c.RecordLatency("GET", 50*time.Millisecond)  // 50ms -> bucket [25-50ms]

	metrics := c.GetMetrics()

	stats, ok := metrics.RequestLatencies["GET"]
	if !ok {
		t.Fatal("expected GET latency stats")
	}
	if stats.Count != 3 {
		t.Errorf("expected count 3, got %d", stats.Count)
	}
	if stats.AvgMs <= 0 {
		t.Error("expected positive average latency")
	}
	if stats.SumMs <= 0 {
		t.Error("expected positive sum")
	}
}

func TestLatencyHistogramBuckets(t *testing.T) {
	h := &LatencyHistogram{}

	tests := []struct {
		duration       time.Duration
		expectedBucket int
	}{
		{500 * time.Microsecond, 0},  // 0-1ms
		{2 * time.Millisecond, 1},    // 1-5ms
		{7 * time.Millisecond, 2},    // 5-10ms
		{15 * time.Millisecond, 3},   // 10-25ms
		{30 * time.Millisecond, 4},   // 25-50ms
		{75 * time.Millisecond, 5},   // 50-100ms
		{200 * time.Millisecond, 6},  // 100-250ms
		{400 * time.Millisecond, 7},  // 250-500ms
		{800 * time.Millisecond, 8},  // 500-1000ms
		{2000 * time.Millisecond, 9}, // 1000ms+
	}

	for _, tt := range tests {
		h.Record(tt.duration)
	}

	for i, tc := range tests {
		if h.buckets[tc.expectedBucket] == 0 {
			t.Errorf("test %d: expected non-zero count in bucket %d for duration %v",
				i, tc.expectedBucket, tc.duration)
		}
	}

	if h.count != uint64(len(tests)) {
		t.Errorf("expected count %d, got %d", len(tests), h.count)
	}
}

func TestLatencyHistogramConcurrent(t *testing.T) {
	h := &LatencyHistogram{}

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			h.Record(time.Duration(i) * time.Millisecond)
		}(i)
	}
	wg.Wait()

	if h.count != 100 {
		t.Errorf("expected count 100, got %d", h.count)
	}
}

func TestIncrementDecrementConnections(t *testing.T) {
	c := NewCollector()

	c.IncrementConnections()
	c.IncrementConnections()
	c.IncrementConnections()

	metrics := c.GetMetrics()
	if metrics.ActiveConnections != 3 {
		t.Errorf("expected 3 connections, got %d", metrics.ActiveConnections)
	}

	c.DecrementConnections()
	metrics = c.GetMetrics()
	if metrics.ActiveConnections != 2 {
		t.Errorf("expected 2 connections, got %d", metrics.ActiveConnections)
	}
}

func TestConnectionsConcurrent(t *testing.T) {
	c := NewCollector()

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.IncrementConnections()
		}()
	}
	wg.Wait()

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.DecrementConnections()
		}()
	}
	wg.Wait()

	metrics := c.GetMetrics()
	if metrics.ActiveConnections != 30 {
		t.Errorf("expected 30 connections, got %d", metrics.ActiveConnections)
	}
}

func TestSetContainerCount(t *testing.T) {
	c := NewCollector()

	c.SetContainerCount(5)
	metrics := c.GetMetrics()
	if metrics.ContainerCount != 5 {
		t.Errorf("expected container count 5, got %d", metrics.ContainerCount)
	}

	c.SetContainerCount(0)
	metrics = c.GetMetrics()
	if metrics.ContainerCount != 0 {
		t.Errorf("expected container count 0, got %d", metrics.ContainerCount)
	}
}

func TestSetPeerCount(t *testing.T) {
	c := NewCollector()

	c.SetPeerCount(10)
	metrics := c.GetMetrics()
	if metrics.PeerCount != 10 {
		t.Errorf("expected peer count 10, got %d", metrics.PeerCount)
	}
}

func TestGetMetrics(t *testing.T) {
	c := NewCollector()

	c.RecordRequest("deploy")
	c.RecordRequest("status")
	c.RecordLatency("deploy", 10*time.Millisecond)
	c.SetContainerCount(3)
	c.SetPeerCount(7)
	c.IncrementConnections()

	metrics := c.GetMetrics()

	if metrics.UptimeSeconds <= 0 {
		t.Error("expected positive uptime")
	}
	if metrics.Uptime == "" {
		t.Error("expected non-empty uptime string")
	}
	if len(metrics.RequestCounts) != 2 {
		t.Errorf("expected 2 request count entries, got %d", len(metrics.RequestCounts))
	}
	if len(metrics.RequestLatencies) != 1 {
		t.Errorf("expected 1 latency entry, got %d", len(metrics.RequestLatencies))
	}
	if metrics.ActiveConnections != 1 {
		t.Errorf("expected 1 connection, got %d", metrics.ActiveConnections)
	}
	if metrics.ContainerCount != 3 {
		t.Errorf("expected 3 containers, got %d", metrics.ContainerCount)
	}
	if metrics.PeerCount != 7 {
		t.Errorf("expected 7 peers, got %d", metrics.PeerCount)
	}
	if metrics.CollectedAt.IsZero() {
		t.Error("expected non-zero CollectedAt")
	}
}

func TestGetMetricsJSON(t *testing.T) {
	c := NewCollector()

	c.RecordRequest("test")
	c.SetContainerCount(2)

	jsonData, err := c.GetMetricsJSON()
	if err != nil {
		t.Fatalf("GetMetricsJSON() error: %v", err)
	}
	if len(jsonData) == 0 {
		t.Fatal("expected non-empty JSON")
	}

	// Verify it's valid JSON
	var metrics Metrics
	if err := json.Unmarshal(jsonData, &metrics); err != nil {
		t.Fatalf("failed to parse metrics JSON: %v", err)
	}

	if metrics.ContainerCount != 2 {
		t.Errorf("expected container count 2 from JSON, got %d", metrics.ContainerCount)
	}
	if metrics.RequestCounts["test"] != 1 {
		t.Errorf("expected test count 1 from JSON, got %d", metrics.RequestCounts["test"])
	}
}

func TestReset(t *testing.T) {
	c := NewCollector()

	// Populate some data
	c.RecordRequest("GET")
	c.RecordLatency("GET", 10*time.Millisecond)
	c.IncrementConnections()
	c.SetContainerCount(5)
	c.SetPeerCount(3)

	// Reset
	c.Reset()

	metrics := c.GetMetrics()
	if len(metrics.RequestCounts) != 0 {
		t.Errorf("expected 0 request counts after reset, got %d", len(metrics.RequestCounts))
	}
	if len(metrics.RequestLatencies) != 0 {
		t.Errorf("expected 0 latencies after reset, got %d", len(metrics.RequestLatencies))
	}
	if metrics.ActiveConnections != 0 {
		t.Errorf("expected 0 connections after reset, got %d", metrics.ActiveConnections)
	}
	if metrics.ContainerCount != 0 {
		t.Errorf("expected 0 containers after reset, got %d", metrics.ContainerCount)
	}
	if metrics.PeerCount != 0 {
		t.Errorf("expected 0 peers after reset, got %d", metrics.PeerCount)
	}
}

func TestLatencyStatsAverage(t *testing.T) {
	c := NewCollector()

	// Record known latencies
	c.RecordLatency("test", 10*time.Millisecond)
	c.RecordLatency("test", 20*time.Millisecond)
	c.RecordLatency("test", 30*time.Millisecond)

	metrics := c.GetMetrics()
	stats := metrics.RequestLatencies["test"]

	// Average should be ~20ms
	if stats.AvgMs < 19 || stats.AvgMs > 21 {
		t.Errorf("expected average ~20ms, got %fms", stats.AvgMs)
	}

	// Sum should be ~60ms
	if stats.SumMs < 59 || stats.SumMs > 61 {
		t.Errorf("expected sum ~60ms, got %fms", stats.SumMs)
	}
}

func TestBucketBoundaries(t *testing.T) {
	// Verify bucket boundaries array
	if len(bucketBoundaries) != 9 {
		t.Errorf("expected 9 bucket boundaries, got %d", len(bucketBoundaries))
	}

	// Verify they're sorted ascending
	for i := 0; i < len(bucketBoundaries)-1; i++ {
		if bucketBoundaries[i] >= bucketBoundaries[i+1] {
			t.Errorf("bucket boundaries not sorted at index %d", i)
		}
	}
}

func TestMetricsUptime(t *testing.T) {
	c := NewCollector()

	// Wait a bit to get measurable uptime
	time.Sleep(10 * time.Millisecond)

	metrics := c.GetMetrics()
	if metrics.UptimeSeconds < 0.01 {
		t.Error("expected measurable uptime")
	}
}

func TestGoroutineCount(t *testing.T) {
	c := NewCollector()

	// UpdateGoroutineCount should store a positive value
	c.UpdateGoroutineCount()
	count := c.GoroutineCount()
	if count <= 0 {
		t.Errorf("expected positive goroutine count, got %d", count)
	}

	// GetMetrics should also populate the goroutine count
	metrics := c.GetMetrics()
	if metrics.GoroutineCount <= 0 {
		t.Errorf("expected positive goroutine count in metrics, got %d", metrics.GoroutineCount)
	}
}

func TestCheckGoroutineHealth(t *testing.T) {
	c := NewCollector()

	count, healthy := c.CheckGoroutineHealth()
	if count <= 0 {
		t.Errorf("expected positive goroutine count, got %d", count)
	}
	// A normal test process should be well below the threshold
	if !healthy {
		t.Errorf("expected healthy (count %d < threshold %d)", count, GoroutineAlertThreshold)
	}

	// The stored count should match what was returned
	stored := c.GoroutineCount()
	if stored != int64(count) {
		t.Errorf("expected stored count %d to match returned count %d", stored, count)
	}
}

func TestGoroutineAlertThreshold(t *testing.T) {
	if GoroutineAlertThreshold != 10000 {
		t.Errorf("expected GoroutineAlertThreshold to be 10000, got %d", GoroutineAlertThreshold)
	}
}

func TestGoroutineCountResetAndCollect(t *testing.T) {
	c := NewCollector()

	// Collect goroutine count
	c.UpdateGoroutineCount()
	if c.GoroutineCount() <= 0 {
		t.Fatal("expected positive goroutine count after update")
	}

	// Reset should zero the goroutine count
	c.Reset()

	// After reset, the stored count should be 0 (until next update)
	if c.GoroutineCount() != 0 {
		t.Errorf("expected 0 goroutine count after reset, got %d", c.GoroutineCount())
	}

	// GetMetrics triggers an update, so it should be positive again
	metrics := c.GetMetrics()
	if metrics.GoroutineCount <= 0 {
		t.Errorf("expected positive goroutine count after GetMetrics, got %d", metrics.GoroutineCount)
	}
}
