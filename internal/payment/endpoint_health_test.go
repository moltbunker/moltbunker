package payment

import (
	"testing"
	"time"
)

func TestEndpointTracker_NewTracker(t *testing.T) {
	urls := []string{"https://rpc1.example.com", "https://rpc2.example.com"}
	et := NewEndpointTracker(urls)

	if et.Len() != 2 {
		t.Fatalf("expected 2 endpoints, got %d", et.Len())
	}

	// All should start healthy
	healthy := et.GetHealthy()
	if len(healthy) != 2 {
		t.Fatalf("expected 2 healthy endpoints, got %d", len(healthy))
	}
}

func TestEndpointTracker_RecordSuccessUpdatesLatency(t *testing.T) {
	et := NewEndpointTracker([]string{"https://rpc1.example.com"})

	et.RecordSuccess("https://rpc1.example.com", 100*time.Millisecond)
	et.RecordSuccess("https://rpc1.example.com", 200*time.Millisecond)

	healthy := et.GetHealthy()
	if len(healthy) != 1 {
		t.Fatalf("expected 1 healthy endpoint, got %d", len(healthy))
	}

	// EWMA after two samples: first=100ms, second=0.3*200+0.7*100=130ms
	et.mu.RLock()
	ep := et.endpoints[0]
	et.mu.RUnlock()

	if ep.Latency < 120*time.Millisecond || ep.Latency > 140*time.Millisecond {
		t.Errorf("expected EWMA latency ~130ms, got %v", ep.Latency)
	}
}

func TestEndpointTracker_ErrorMarksUnhealthy(t *testing.T) {
	et := NewEndpointTracker([]string{"https://rpc1.example.com", "https://rpc2.example.com"})

	// Record 3 consecutive errors (default threshold)
	for i := 0; i < 3; i++ {
		et.RecordError("https://rpc1.example.com")
	}

	healthy := et.GetHealthy()
	if len(healthy) != 1 {
		t.Fatalf("expected 1 healthy endpoint after errors, got %d", len(healthy))
	}
	if healthy[0] != "https://rpc2.example.com" {
		t.Errorf("expected rpc2 to be healthy, got %s", healthy[0])
	}
}

func TestEndpointTracker_SuccessResetsErrors(t *testing.T) {
	et := NewEndpointTracker([]string{"https://rpc1.example.com"})

	// 2 errors (below threshold)
	et.RecordError("https://rpc1.example.com")
	et.RecordError("https://rpc1.example.com")

	// Success resets counter
	et.RecordSuccess("https://rpc1.example.com", 50*time.Millisecond)

	// 2 more errors — still below threshold since counter was reset
	et.RecordError("https://rpc1.example.com")
	et.RecordError("https://rpc1.example.com")

	healthy := et.GetHealthy()
	if len(healthy) != 1 {
		t.Fatalf("expected endpoint still healthy after reset, got %d healthy", len(healthy))
	}
}

func TestEndpointTracker_GetNextExcludesCurrent(t *testing.T) {
	et := NewEndpointTracker([]string{"https://rpc1.example.com", "https://rpc2.example.com"})
	et.RecordSuccess("https://rpc1.example.com", 100*time.Millisecond)
	et.RecordSuccess("https://rpc2.example.com", 50*time.Millisecond)

	next, ok := et.GetNext("https://rpc2.example.com")
	if !ok {
		t.Fatal("expected to find alternative endpoint")
	}
	if next != "https://rpc1.example.com" {
		t.Errorf("expected rpc1, got %s", next)
	}
}

func TestEndpointTracker_GetNextNoAlternatives(t *testing.T) {
	et := NewEndpointTracker([]string{"https://rpc1.example.com"})

	_, ok := et.GetNext("https://rpc1.example.com")
	if ok {
		t.Error("expected no alternative when only one endpoint exists")
	}
}

func TestEndpointTracker_SortsByLatency(t *testing.T) {
	et := NewEndpointTracker([]string{"https://slow.example.com", "https://fast.example.com"})

	et.RecordSuccess("https://slow.example.com", 200*time.Millisecond)
	et.RecordSuccess("https://fast.example.com", 10*time.Millisecond)

	healthy := et.GetHealthy()
	if len(healthy) != 2 {
		t.Fatalf("expected 2 healthy, got %d", len(healthy))
	}
	if healthy[0] != "https://fast.example.com" {
		t.Errorf("expected fast endpoint first, got %s", healthy[0])
	}
}

func TestEndpointTracker_RecoveryAfterInterval(t *testing.T) {
	et := NewEndpointTracker([]string{"https://rpc1.example.com"})
	et.recovery = 10 * time.Millisecond // short for testing

	// Mark unhealthy
	for i := 0; i < 3; i++ {
		et.RecordError("https://rpc1.example.com")
	}

	healthy := et.GetHealthy()
	if len(healthy) != 0 {
		t.Fatalf("expected 0 healthy immediately after errors, got %d", len(healthy))
	}

	// Wait for recovery interval
	time.Sleep(15 * time.Millisecond)

	healthy = et.GetHealthy()
	if len(healthy) != 1 {
		t.Fatalf("expected 1 endpoint eligible for recovery, got %d", len(healthy))
	}
}

func TestEndpointTracker_EmptyTracker(t *testing.T) {
	et := NewEndpointTracker(nil)

	if et.Len() != 0 {
		t.Errorf("expected 0 endpoints, got %d", et.Len())
	}

	healthy := et.GetHealthy()
	if len(healthy) != 0 {
		t.Errorf("expected 0 healthy, got %d", len(healthy))
	}

	_, ok := et.GetNext("")
	if ok {
		t.Error("expected no next endpoint from empty tracker")
	}
}

func TestEndpointTracker_UnknownURLIgnored(t *testing.T) {
	et := NewEndpointTracker([]string{"https://rpc1.example.com"})

	// These should not panic or affect tracked endpoints
	et.RecordSuccess("https://unknown.example.com", 50*time.Millisecond)
	et.RecordError("https://unknown.example.com")

	if et.Len() != 1 {
		t.Errorf("expected 1 endpoint, got %d", et.Len())
	}
}
