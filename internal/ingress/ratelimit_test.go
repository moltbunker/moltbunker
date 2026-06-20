package ingress

import (
	"io"
	"strings"
	"testing"
)

func TestIngressRateLimiter_AllowsUnderLimit(t *testing.T) {
	rl := NewIngressRateLimiter(RateLimitConfig{DefaultRPS: 10, DefaultBurst: 10, MaxConcurrency: 10})
	for i := 0; i < 5; i++ {
		if !rl.Allow("tenant") {
			t.Fatalf("request %d unexpectedly blocked under burst=10", i)
		}
	}
}

func TestIngressRateLimiter_BlocksOverLimit(t *testing.T) {
	// Burst of 2; the third immediate call must be denied (RPS=1 refills slowly).
	rl := NewIngressRateLimiter(RateLimitConfig{DefaultRPS: 1, DefaultBurst: 2, MaxConcurrency: 10})
	if !rl.Allow("tenant") {
		t.Fatal("first call should be allowed")
	}
	if !rl.Allow("tenant") {
		t.Fatal("second call (burst) should be allowed")
	}
	if rl.Allow("tenant") {
		t.Fatal("third call should be blocked after burst exhausted")
	}
}

func TestIngressRateLimiter_ConcurrencyCap(t *testing.T) {
	rl := NewIngressRateLimiter(RateLimitConfig{DefaultRPS: 1000, DefaultBurst: 1000, MaxConcurrency: 2})

	r1, ok1 := rl.AcquireConcurrency("tenant")
	r2, ok2 := rl.AcquireConcurrency("tenant")
	_, ok3 := rl.AcquireConcurrency("tenant")
	if !ok1 || !ok2 {
		t.Fatalf("first two acquisitions should succeed (ok1=%v ok2=%v)", ok1, ok2)
	}
	if ok3 {
		t.Fatal("third acquisition should fail at MaxConcurrency=2")
	}
	r1()
	r2()
}

func TestIngressRateLimiter_TenantIsolation(t *testing.T) {
	rl := NewIngressRateLimiter(RateLimitConfig{DefaultRPS: 1, DefaultBurst: 1, MaxConcurrency: 10})

	if !rl.Allow("tenant-A") {
		t.Fatal("tenant-A first call should be allowed")
	}
	if rl.Allow("tenant-A") {
		t.Fatal("tenant-A second call should be blocked")
	}
	// tenant-B has its own independent bucket.
	if !rl.Allow("tenant-B") {
		t.Fatal("tenant-B must not be affected by tenant-A exhaustion")
	}
}

func TestIngressRateLimiter_ReleaseRestoresConcurrency(t *testing.T) {
	rl := NewIngressRateLimiter(RateLimitConfig{DefaultRPS: 1000, DefaultBurst: 1000, MaxConcurrency: 1})

	release, ok := rl.AcquireConcurrency("tenant")
	if !ok {
		t.Fatal("first acquire should succeed")
	}
	if _, ok2 := rl.AcquireConcurrency("tenant"); ok2 {
		t.Fatal("second acquire should fail at cap=1")
	}
	release()
	r2, ok3 := rl.AcquireConcurrency("tenant")
	if !ok3 {
		t.Fatal("acquire after release should succeed")
	}
	r2()
}

func TestIngressRateLimiter_ReleaseIdempotent(t *testing.T) {
	rl := NewIngressRateLimiter(RateLimitConfig{MaxConcurrency: 1})
	release, ok := rl.AcquireConcurrency("tenant")
	if !ok {
		t.Fatal("acquire should succeed")
	}
	release()
	release() // must not panic or over-release
	if _, ok := rl.AcquireConcurrency("tenant"); !ok {
		t.Fatal("slot should be available after idempotent release")
	}
}

func TestIngressRateLimiter_TenantOverride(t *testing.T) {
	rl := NewIngressRateLimiter(RateLimitConfig{
		DefaultRPS: 1, DefaultBurst: 1, MaxConcurrency: 1,
		TenantOverrides: map[string]TenantRateConfig{
			"vip": {RPS: 1000, Burst: 5, MaxConcurrency: 5},
		},
	})
	for i := 0; i < 5; i++ {
		if !rl.Allow("vip") {
			t.Fatalf("vip override burst=5: request %d should be allowed", i)
		}
	}
}

func TestIngressRateLimiter_LimitReader(t *testing.T) {
	rl := NewIngressRateLimiter(RateLimitConfig{MaxBodyBytes: 8})
	src := strings.NewReader("0123456789ABCDEF")
	r := rl.LimitReader(src, "tenant")
	b, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if len(b) != 8 {
		t.Fatalf("LimitReader should cap at 8 bytes, got %d", len(b))
	}
}

func TestIngressRateLimiter_StartStop(t *testing.T) {
	rl := NewIngressRateLimiter(RateLimitConfig{})
	rl.Start()
	rl.Stop()
	rl.Stop() // idempotent, must not panic
}
