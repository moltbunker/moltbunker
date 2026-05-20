package runtime

import (
	"reflect"
	"sort"
	"testing"
	"time"
)

// R15 — fluent setters + size-bounded LRU eviction tests.

func TestImageGC_WithInterval_SetsField(t *testing.T) {
	gc := NewImageGC(nil, time.Hour, 0)
	if gc.interval != defaultGCInterval {
		t.Fatalf("default interval not applied; got %v want %v", gc.interval, defaultGCInterval)
	}
	gc.WithInterval(15 * time.Minute)
	if gc.interval != 15*time.Minute {
		t.Fatalf("WithInterval not applied; got %v", gc.interval)
	}
}

func TestImageGC_WithInterval_NormalizesZeroAndNegative(t *testing.T) {
	gc := NewImageGC(nil, time.Hour, 0)
	gc.WithInterval(0)
	if gc.interval != defaultGCInterval {
		t.Fatalf("zero should normalize to default; got %v", gc.interval)
	}
	gc.WithInterval(-1 * time.Minute)
	if gc.interval != defaultGCInterval {
		t.Fatalf("negative should normalize to default; got %v", gc.interval)
	}
}

func TestImageGC_WithInterval_IsChainable(t *testing.T) {
	gc := NewImageGC(nil, time.Hour, 0).WithInterval(5 * time.Minute).WithActiveThreshold(30 * time.Second)
	if gc.interval != 5*time.Minute {
		t.Fatalf("interval mis-set after chain")
	}
	if gc.activeThreshold != 30*time.Second {
		t.Fatalf("activeThreshold mis-set after chain")
	}
}

func TestImageGC_WithActiveThreshold_NormalizesZero(t *testing.T) {
	gc := NewImageGC(nil, time.Hour, 0)
	gc.WithActiveThreshold(0)
	if gc.activeThreshold != defaultActiveThreshold {
		t.Fatalf("zero should normalize to default; got %v", gc.activeThreshold)
	}
}

func TestSelectForEviction_NoOpUnderCap(t *testing.T) {
	now := time.Now()
	survivors := []imageInfo{
		{ref: "img-a", lastUsed: now.Add(-time.Hour)},
		{ref: "img-b", lastUsed: now.Add(-2 * time.Hour)},
	}
	sizes := map[string]int64{"img-a": 100, "img-b": 200}

	got := selectForEviction(survivors, sizes, 300, 500, 5*time.Minute, now)
	if got != nil {
		t.Fatalf("under cap should return nil, got %v", got)
	}
}

func TestSelectForEviction_EvictsOldestFirst(t *testing.T) {
	now := time.Now()
	// Three images, total 600 bytes, cap 350. Need to free at least 250.
	// Oldest is img-c (3h ago), then img-b (2h), then img-a (1h).
	survivors := []imageInfo{
		{ref: "img-a", lastUsed: now.Add(-1 * time.Hour)},
		{ref: "img-b", lastUsed: now.Add(-2 * time.Hour)},
		{ref: "img-c", lastUsed: now.Add(-3 * time.Hour)},
	}
	sizes := map[string]int64{"img-a": 200, "img-b": 200, "img-c": 200}

	got := selectForEviction(survivors, sizes, 600, 350, 5*time.Minute, now)
	// Evicting img-c brings us to 400 (still over); img-b brings 200 (under). Two evictions.
	want := []string{"img-c", "img-b"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v (oldest first)", got, want)
	}
}

func TestSelectForEviction_SkipsActiveImages(t *testing.T) {
	now := time.Now()
	// img-c is the oldest BUT is also "active" (marked within the last
	// 5 minutes). The cap is breached; selector must NOT pick img-c.
	survivors := []imageInfo{
		{ref: "img-a", lastUsed: now.Add(-1 * time.Hour)},
		{ref: "img-b", lastUsed: now.Add(-2 * time.Hour)},
		{ref: "img-c", lastUsed: now.Add(-30 * time.Second)}, // active
	}
	sizes := map[string]int64{"img-a": 200, "img-b": 200, "img-c": 200}

	got := selectForEviction(survivors, sizes, 600, 350, 5*time.Minute, now)
	// Active img-c is skipped. Oldest non-active is img-b. Then img-a.
	want := []string{"img-b", "img-a"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v (skip active)", got, want)
	}
}

func TestSelectForEviction_RunsOverBudgetWhenOnlyActiveRemain(t *testing.T) {
	now := time.Now()
	// All survivors are active. Selector must NOT evict any — running over
	// budget is preferable to killing a running container's image.
	survivors := []imageInfo{
		{ref: "img-a", lastUsed: now.Add(-1 * time.Minute)},
		{ref: "img-b", lastUsed: now.Add(-2 * time.Minute)},
	}
	sizes := map[string]int64{"img-a": 500, "img-b": 500}

	got := selectForEviction(survivors, sizes, 1000, 100, 5*time.Minute, now)
	if len(got) != 0 {
		t.Fatalf("expected empty (all active), got %v", got)
	}
}

func TestSelectForEviction_NeverMarkedImagesEvictFirst(t *testing.T) {
	now := time.Now()
	// img-z has no lastUsed (zero time — never marked). Should be evicted
	// before any image with a real timestamp.
	survivors := []imageInfo{
		{ref: "img-a", lastUsed: now.Add(-1 * time.Hour)},
		{ref: "img-z", lastUsed: time.Time{}}, // never marked
	}
	sizes := map[string]int64{"img-a": 100, "img-z": 100}

	got := selectForEviction(survivors, sizes, 200, 100, 5*time.Minute, now)
	want := []string{"img-z"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v (zero-time evicts first)", got, want)
	}
}

func TestSelectForEviction_StopsAtFirstUnderBudget(t *testing.T) {
	now := time.Now()
	// Three images at 100 each, total 300, cap 150 → need to free 150.
	// Evicting oldest (img-c, 100) brings us to 200; still over.
	// Evicting img-b (100) brings us to 100; under.
	// Selector must stop — img-a survives.
	survivors := []imageInfo{
		{ref: "img-a", lastUsed: now.Add(-1 * time.Hour)},
		{ref: "img-b", lastUsed: now.Add(-2 * time.Hour)},
		{ref: "img-c", lastUsed: now.Add(-3 * time.Hour)},
	}
	sizes := map[string]int64{"img-a": 100, "img-b": 100, "img-c": 100}

	got := selectForEviction(survivors, sizes, 300, 150, 5*time.Minute, now)
	want := []string{"img-c", "img-b"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestSelectForEviction_HandlesEmpty(t *testing.T) {
	got := selectForEviction(nil, nil, 0, 100, time.Minute, time.Now())
	if got != nil {
		t.Fatalf("empty input should return nil, got %v", got)
	}
}

// TestSelectForEviction_StableSortByLastUsed checks that when two images
// share a lastUsed timestamp, the eviction order is deterministic enough
// for tests (we don't promise a specific tiebreak rule but sort must not
// crash, and the union of evictions must be correct).
func TestSelectForEviction_HandlesEqualTimestamps(t *testing.T) {
	t0 := time.Now().Add(-1 * time.Hour)
	survivors := []imageInfo{
		{ref: "img-a", lastUsed: t0},
		{ref: "img-b", lastUsed: t0},
	}
	sizes := map[string]int64{"img-a": 100, "img-b": 100}

	got := selectForEviction(survivors, sizes, 200, 50, 5*time.Minute, time.Now())
	sort.Strings(got)
	want := []string{"img-a", "img-b"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v (both should be evicted)", got, want)
	}
}
