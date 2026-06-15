package ingress

import (
	"io"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"github.com/moltbunker/moltbunker/internal/logging"
)

// RateLimitConfig configures the per-tenant ingress abuse limiter. These are
// L7 HTTP-request limits (request count, in-flight concurrency, body size),
// distinct from the L4 byte-stream throttling in internal/tunnel/bandwidth.go.
type RateLimitConfig struct {
	// DefaultRPS is the sustained allowed requests-per-second per tenant.
	DefaultRPS int
	// DefaultBurst is the token-bucket burst capacity per tenant.
	DefaultBurst int
	// MaxConcurrency is the max simultaneous in-flight requests per tenant.
	MaxConcurrency int
	// MaxBodyBytes is the hard request body cap applied via LimitReader.
	MaxBodyBytes int64
	// TenantOverrides lets specific subdomains use different limits.
	TenantOverrides map[string]TenantRateConfig
}

// TenantRateConfig overrides rate/concurrency limits for one tenant.
type TenantRateConfig struct {
	RPS            int
	Burst          int
	MaxConcurrency int
}

// tenantBucket holds the per-tenant limiter state.
type tenantBucket struct {
	limiter  *rate.Limiter
	sem      chan struct{}
	lastSeen time.Time
	mu       sync.Mutex
}

// IngressRateLimiter enforces per-tenant token-bucket rate limits plus a
// concurrency cap. Buckets are created lazily per subdomain and evicted by a
// background cleanup goroutine once stale, mirroring api.Server's pattern.
//
// KNOWN LIMITATION (next hardening step for the edge layer): limits are keyed
// strictly by subdomain (tenant), not by client IP. A single abusive client
// can therefore consume a victim tenant's entire per-tenant RPS/concurrency
// budget (a cross-client DoS within one tenant). The correct fix is a second,
// per-client-IP dimension (a global per-IP limiter alongside this per-tenant
// one), keyed on the *true* client IP resolved from a trusted X-Forwarded-For /
// PROXY-protocol header — which requires a configured trusted-proxy boundary so
// the header cannot be spoofed. That trust configuration does not yet exist, so
// adding naive XFF parsing here would create a spoofable bypass; it is
// deliberately deferred to the edge-provider work (EDGE-02) rather than shipped
// half-built. The same true-client-IP resolution must then feed Coraza's
// ProcessConnection in waf.go (today it uses r.RemoteAddr, which is the
// upstream LB/TLS-terminator address when ingress sits behind one).
type IngressRateLimiter struct {
	cfg     RateLimitConfig
	buckets sync.Map // subdomain -> *tenantBucket

	stopCh  chan struct{}
	stopMu  sync.Mutex
	stopped bool
}

// NewIngressRateLimiter builds a limiter from cfg. Defaults are applied for any
// non-positive field so a zero-value config still yields a usable limiter.
func NewIngressRateLimiter(cfg RateLimitConfig) *IngressRateLimiter {
	if cfg.DefaultRPS <= 0 {
		cfg.DefaultRPS = 100
	}
	if cfg.DefaultBurst <= 0 {
		cfg.DefaultBurst = cfg.DefaultRPS * 2
	}
	if cfg.MaxConcurrency <= 0 {
		cfg.MaxConcurrency = 50
	}
	if cfg.MaxBodyBytes <= 0 {
		cfg.MaxBodyBytes = 10 * 1024 * 1024
	}
	return &IngressRateLimiter{
		cfg:    cfg,
		stopCh: make(chan struct{}),
	}
}

// limitsFor resolves the effective (rps, burst, maxConcurrency) for a tenant.
func (l *IngressRateLimiter) limitsFor(tenantID string) (rps, burst, maxConc int) {
	rps, burst, maxConc = l.cfg.DefaultRPS, l.cfg.DefaultBurst, l.cfg.MaxConcurrency
	if ov, ok := l.cfg.TenantOverrides[tenantID]; ok {
		if ov.RPS > 0 {
			rps = ov.RPS
		}
		if ov.Burst > 0 {
			burst = ov.Burst
		}
		if ov.MaxConcurrency > 0 {
			maxConc = ov.MaxConcurrency
		}
	}
	return rps, burst, maxConc
}

// bucketFor returns the tenant bucket, creating it on first use.
func (l *IngressRateLimiter) bucketFor(tenantID string) *tenantBucket {
	if existing, ok := l.buckets.Load(tenantID); ok {
		b := existing.(*tenantBucket)
		b.mu.Lock()
		b.lastSeen = time.Now()
		b.mu.Unlock()
		return b
	}

	rps, burst, maxConc := l.limitsFor(tenantID)
	nb := &tenantBucket{
		limiter:  rate.NewLimiter(rate.Limit(rps), burst),
		sem:      make(chan struct{}, maxConc),
		lastSeen: time.Now(),
	}
	actual, _ := l.buckets.LoadOrStore(tenantID, nb)
	return actual.(*tenantBucket)
}

// Allow reports whether a request for tenantID is within the token-bucket rate.
// It consumes one token when allowed.
func (l *IngressRateLimiter) Allow(tenantID string) bool {
	return l.bucketFor(tenantID).limiter.Allow()
}

// AcquireConcurrency attempts to reserve a concurrency slot for the tenant. It
// returns a release closure (always non-nil; safe to call once) and ok=false if
// the tenant is already at its concurrency cap.
func (l *IngressRateLimiter) AcquireConcurrency(tenantID string) (release func(), ok bool) {
	b := l.bucketFor(tenantID)
	select {
	case b.sem <- struct{}{}:
		var once sync.Once
		return func() {
			once.Do(func() { <-b.sem })
		}, true
	default:
		return func() {}, false
	}
}

// MaxBodyBytes returns the configured hard request body cap.
func (l *IngressRateLimiter) MaxBodyBytes() int64 {
	return l.cfg.MaxBodyBytes
}

// LimitReader wraps r so that at most MaxBodyBytes are readable. tenantID is
// accepted for symmetry / future per-tenant body caps.
func (l *IngressRateLimiter) LimitReader(r io.Reader, _ string) io.Reader {
	return io.LimitReader(r, l.cfg.MaxBodyBytes)
}

// Start launches the background cleanup goroutine (5-minute ticker; evicts
// buckets idle for >10 minutes). It is safe to call once.
func (l *IngressRateLimiter) Start() {
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-l.stopCh:
				return
			case <-ticker.C:
				l.cleanup()
			}
		}
	}()
}

// Stop terminates the cleanup goroutine. Idempotent.
func (l *IngressRateLimiter) Stop() {
	l.stopMu.Lock()
	defer l.stopMu.Unlock()
	if l.stopped {
		return
	}
	l.stopped = true
	close(l.stopCh)
}

// cleanup evicts tenant buckets not seen in the last 10 minutes.
func (l *IngressRateLimiter) cleanup() {
	stale := time.Now().Add(-10 * time.Minute)
	var cleaned int
	l.buckets.Range(func(key, value any) bool {
		b := value.(*tenantBucket)
		b.mu.Lock()
		last := b.lastSeen
		inFlight := len(b.sem)
		b.mu.Unlock()
		// Never evict a bucket with in-flight requests (its release closure
		// holds a reference to b.sem).
		if inFlight == 0 && last.Before(stale) {
			l.buckets.Delete(key)
			cleaned++
		}
		return true
	})
	if cleaned > 0 {
		logging.Debug("cleaned up stale ingress rate limiters",
			"count", cleaned,
			logging.Component("ingress"))
	}
}
