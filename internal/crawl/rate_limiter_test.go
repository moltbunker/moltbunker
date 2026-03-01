package crawl

import (
	"testing"
	"time"
)

func TestRateLimiter_Allow(t *testing.T) {
	rl := NewDomainRateLimiter(100 * time.Millisecond)

	if !rl.Allow("example.com") {
		t.Error("first request should be allowed")
	}

	if rl.Allow("example.com") {
		t.Error("immediate second request should be rate limited")
	}

	// Different domain should be allowed
	if !rl.Allow("other.com") {
		t.Error("different domain should be allowed")
	}

	// Wait for interval to pass
	time.Sleep(120 * time.Millisecond)
	if !rl.Allow("example.com") {
		t.Error("request after interval should be allowed")
	}
}

func TestURLDedup_SeenAndMark(t *testing.T) {
	d := NewURLDedup()

	if d.Seen("job1", "https://example.com") {
		t.Error("should not be seen before marking")
	}

	d.Mark("job1", "https://example.com")

	if !d.Seen("job1", "https://example.com") {
		t.Error("should be seen after marking")
	}

	// Different job should not be affected
	if d.Seen("job2", "https://example.com") {
		t.Error("different job should not be seen")
	}
}

func TestURLDedup_Count(t *testing.T) {
	d := NewURLDedup()

	if d.Count("job1") != 0 {
		t.Error("empty job should have count 0")
	}

	d.Mark("job1", "https://a.com")
	d.Mark("job1", "https://b.com")

	if d.Count("job1") != 2 {
		t.Errorf("count = %d, want 2", d.Count("job1"))
	}
}

func TestURLDedup_Clear(t *testing.T) {
	d := NewURLDedup()

	d.Mark("job1", "https://a.com")
	d.Mark("job1", "https://b.com")
	d.Clear("job1")

	if d.Count("job1") != 0 {
		t.Error("count should be 0 after clear")
	}
	if d.Seen("job1", "https://a.com") {
		t.Error("should not be seen after clear")
	}
}

func TestURLDedup_DuplicateMark(t *testing.T) {
	d := NewURLDedup()

	d.Mark("job1", "https://a.com")
	d.Mark("job1", "https://a.com") // duplicate

	if d.Count("job1") != 1 {
		t.Errorf("count = %d, want 1 (dedup should prevent double count)", d.Count("job1"))
	}
}
