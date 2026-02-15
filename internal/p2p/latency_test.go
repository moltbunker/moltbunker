package p2p

import (
	"sync"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
)

// testPeerID returns a deterministic peer.ID for testing purposes.
func testPeerID(name string) peer.ID {
	return peer.ID(name)
}

func TestRecordLatency_FirstMeasurement(t *testing.T) {
	lm := NewLatencyMonitor()
	pid := testPeerID("peer-1")

	lm.RecordLatency(pid, "Americas", 50*time.Millisecond)

	m, ok := lm.GetLatency(pid)
	if !ok {
		t.Fatal("expected measurement to exist")
	}
	if m.PeerID != pid {
		t.Errorf("expected peer ID %s, got %s", pid, m.PeerID)
	}
	if m.Region != "Americas" {
		t.Errorf("expected region Americas, got %s", m.Region)
	}
	if m.LastPing != 50*time.Millisecond {
		t.Errorf("expected last ping 50ms, got %v", m.LastPing)
	}
	if m.AvgPing != 50*time.Millisecond {
		t.Errorf("expected avg ping 50ms, got %v", m.AvgPing)
	}
	if m.MinPing != 50*time.Millisecond {
		t.Errorf("expected min ping 50ms, got %v", m.MinPing)
	}
	if m.MaxPing != 50*time.Millisecond {
		t.Errorf("expected max ping 50ms, got %v", m.MaxPing)
	}
	if m.SampleCount != 1 {
		t.Errorf("expected sample count 1, got %d", m.SampleCount)
	}
}

func TestRecordLatency_UpdatesAverage(t *testing.T) {
	lm := NewLatencyMonitor()
	pid := testPeerID("peer-1")

	lm.RecordLatency(pid, "Europe", 100*time.Millisecond)
	lm.RecordLatency(pid, "Europe", 200*time.Millisecond)

	m, ok := lm.GetLatency(pid)
	if !ok {
		t.Fatal("expected measurement to exist")
	}
	if m.SampleCount != 2 {
		t.Errorf("expected sample count 2, got %d", m.SampleCount)
	}
	if m.LastPing != 200*time.Millisecond {
		t.Errorf("expected last ping 200ms, got %v", m.LastPing)
	}
	// Average of 100ms and 200ms = 150ms
	if m.AvgPing != 150*time.Millisecond {
		t.Errorf("expected avg ping 150ms, got %v", m.AvgPing)
	}
}

func TestRecordLatency_MultipleMeasurementsUpdateStats(t *testing.T) {
	lm := NewLatencyMonitor()
	pid := testPeerID("peer-1")

	latencies := []time.Duration{
		100 * time.Millisecond,
		50 * time.Millisecond,
		200 * time.Millisecond,
		75 * time.Millisecond,
	}

	for _, l := range latencies {
		lm.RecordLatency(pid, "Asia-Pacific", l)
	}

	m, ok := lm.GetLatency(pid)
	if !ok {
		t.Fatal("expected measurement to exist")
	}
	if m.SampleCount != 4 {
		t.Errorf("expected sample count 4, got %d", m.SampleCount)
	}
	if m.MinPing != 50*time.Millisecond {
		t.Errorf("expected min ping 50ms, got %v", m.MinPing)
	}
	if m.MaxPing != 200*time.Millisecond {
		t.Errorf("expected max ping 200ms, got %v", m.MaxPing)
	}
	if m.LastPing != 75*time.Millisecond {
		t.Errorf("expected last ping 75ms, got %v", m.LastPing)
	}

	// Verify average is reasonable: (100+50+200+75)/4 = 106.25ms
	// Incremental average may differ slightly due to integer division
	expectedAvg := 106 * time.Millisecond
	tolerance := 2 * time.Millisecond
	diff := m.AvgPing - expectedAvg
	if diff < 0 {
		diff = -diff
	}
	if diff > tolerance {
		t.Errorf("expected avg ping ~106ms, got %v", m.AvgPing)
	}
}

func TestGetLatency_NotFound(t *testing.T) {
	lm := NewLatencyMonitor()

	_, ok := lm.GetLatency(testPeerID("nonexistent"))
	if ok {
		t.Error("expected measurement to not exist")
	}
}

func TestGetRegionLatencies_Aggregates(t *testing.T) {
	lm := NewLatencyMonitor()

	// Two peers in Americas
	lm.RecordLatency(testPeerID("us-1"), "Americas", 30*time.Millisecond)
	lm.RecordLatency(testPeerID("us-2"), "Americas", 50*time.Millisecond)

	// One peer in Europe
	lm.RecordLatency(testPeerID("eu-1"), "Europe", 120*time.Millisecond)

	regions := lm.GetRegionLatencies()

	if len(regions) != 2 {
		t.Fatalf("expected 2 regions, got %d", len(regions))
	}

	americasAvg := regions["Americas"]
	// (30+50)/2 = 40ms
	if americasAvg != 40*time.Millisecond {
		t.Errorf("expected Americas avg 40ms, got %v", americasAvg)
	}

	europeAvg := regions["Europe"]
	if europeAvg != 120*time.Millisecond {
		t.Errorf("expected Europe avg 120ms, got %v", europeAvg)
	}
}

func TestGetRegionLatencies_Empty(t *testing.T) {
	lm := NewLatencyMonitor()

	regions := lm.GetRegionLatencies()
	if len(regions) != 0 {
		t.Errorf("expected empty map, got %d entries", len(regions))
	}
}

func TestGetBestPeersByRegion_SortsByLatency(t *testing.T) {
	lm := NewLatencyMonitor()

	// Record peers with different latencies in the same region
	p1 := testPeerID("slow")
	p2 := testPeerID("fast")
	p3 := testPeerID("medium")

	lm.RecordLatency(p1, "Europe", 200*time.Millisecond)
	lm.RecordLatency(p2, "Europe", 20*time.Millisecond)
	lm.RecordLatency(p3, "Europe", 100*time.Millisecond)

	best := lm.GetBestPeersByRegion("Europe", 3)
	if len(best) != 3 {
		t.Fatalf("expected 3 peers, got %d", len(best))
	}
	if best[0] != p2 {
		t.Errorf("expected fastest peer first, got %s", best[0])
	}
	if best[1] != p3 {
		t.Errorf("expected medium peer second, got %s", best[1])
	}
	if best[2] != p1 {
		t.Errorf("expected slowest peer third, got %s", best[2])
	}
}

func TestGetBestPeersByRegion_LimitsResults(t *testing.T) {
	lm := NewLatencyMonitor()

	lm.RecordLatency(testPeerID("p1"), "Americas", 10*time.Millisecond)
	lm.RecordLatency(testPeerID("p2"), "Americas", 20*time.Millisecond)
	lm.RecordLatency(testPeerID("p3"), "Americas", 30*time.Millisecond)

	best := lm.GetBestPeersByRegion("Americas", 2)
	if len(best) != 2 {
		t.Fatalf("expected 2 peers, got %d", len(best))
	}
}

func TestGetBestPeersByRegion_RequestMoreThanAvailable(t *testing.T) {
	lm := NewLatencyMonitor()

	lm.RecordLatency(testPeerID("p1"), "Africa", 50*time.Millisecond)

	best := lm.GetBestPeersByRegion("Africa", 5)
	if len(best) != 1 {
		t.Fatalf("expected 1 peer, got %d", len(best))
	}
}

func TestGetBestPeersByRegion_NoMatchingRegion(t *testing.T) {
	lm := NewLatencyMonitor()

	lm.RecordLatency(testPeerID("p1"), "Americas", 50*time.Millisecond)

	best := lm.GetBestPeersByRegion("Europe", 3)
	if len(best) != 0 {
		t.Errorf("expected 0 peers for non-matching region, got %d", len(best))
	}
}

func TestCleanStale_RemovesOldEntries(t *testing.T) {
	lm := NewLatencyMonitor()

	pid := testPeerID("stale-peer")
	lm.RecordLatency(pid, "Americas", 50*time.Millisecond)

	// Manually set LastUpdated to the past
	lm.mu.Lock()
	lm.measurements[pid].LastUpdated = time.Now().Add(-2 * time.Hour)
	lm.mu.Unlock()

	// Add a fresh peer
	freshPID := testPeerID("fresh-peer")
	lm.RecordLatency(freshPID, "Europe", 100*time.Millisecond)

	lm.CleanStale(1 * time.Hour)

	_, staleExists := lm.GetLatency(pid)
	if staleExists {
		t.Error("expected stale peer to be removed")
	}

	_, freshExists := lm.GetLatency(freshPID)
	if !freshExists {
		t.Error("expected fresh peer to still exist")
	}
}

func TestCleanStale_KeepsFreshEntries(t *testing.T) {
	lm := NewLatencyMonitor()

	pid := testPeerID("fresh-peer")
	lm.RecordLatency(pid, "Americas", 50*time.Millisecond)

	lm.CleanStale(1 * time.Hour)

	_, ok := lm.GetLatency(pid)
	if !ok {
		t.Error("expected fresh peer to still exist after cleanup")
	}
}

func TestConcurrentAccess(t *testing.T) {
	lm := NewLatencyMonitor()
	var wg sync.WaitGroup

	// Concurrent writers
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			pid := testPeerID("peer-concurrent")
			for j := 0; j < 100; j++ {
				lm.RecordLatency(pid, "Americas", time.Duration(idx+j)*time.Millisecond)
			}
		}(i)
	}

	// Concurrent readers
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				lm.GetLatency(testPeerID("peer-concurrent"))
				lm.GetRegionLatencies()
				lm.GetBestPeersByRegion("Americas", 5)
			}
		}()
	}

	// Concurrent cleanup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for j := 0; j < 50; j++ {
			lm.CleanStale(1 * time.Hour)
		}
	}()

	wg.Wait()

	// If we get here without a data race or panic, the test passes.
	// Verify we have at least one measurement
	m, ok := lm.GetLatency(testPeerID("peer-concurrent"))
	if !ok {
		t.Fatal("expected measurement to exist after concurrent writes")
	}
	if m.SampleCount < 1 {
		t.Error("expected at least one sample after concurrent writes")
	}
}
