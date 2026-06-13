package ingress

import (
	"bufio"
	"bytes"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

// stubWAF is an in-package WAFEngine stub for middleware wiring tests so they
// never need the real Coraza engine / CRS. (tests/mocks.MockWAFEngine offers
// the same behavior for cross-package callers.)
type stubWAF struct {
	block   bool
	mode    WAFMode
	matched []string
}

func (s stubWAF) Inspect(_ *http.Request, _ string) (*WAFResult, error) {
	res := &WAFResult{MatchedPhases: map[string]int{}, MatchedRules: s.matched}
	for _, id := range s.matched {
		res.MatchedPhases[id] = 2
	}
	if s.block {
		res.Blocked = true
		res.Phase = 2
	}
	return res, nil
}

func (s stubWAF) Mode(string) WAFMode {
	if s.mode != "" {
		return s.mode
	}
	return ModeDetection
}

func permissiveRL() *IngressRateLimiter {
	return NewIngressRateLimiter(RateLimitConfig{
		DefaultRPS: 1000, DefaultBurst: 1000, MaxConcurrency: 1000, MaxBodyBytes: 1 << 20,
	})
}

func okHandler(called *bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if called != nil {
			*called = true
		}
		// Drain body to mimic the upstream proxy consuming it.
		_, _ = io.Copy(io.Discard, r.Body)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
}

func TestMiddleware_CleanRequestPassesThrough(t *testing.T) {
	em := NewEdgeMetrics(prometheus.NewRegistry())
	mw := NewIngressMiddleware(NoopWAFEngine{}, permissiveRL(), em)

	called := false
	h := mw.Wrap("acme", okHandler(&called))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rec, req)

	if !called {
		t.Fatal("handler was not called for a clean request")
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}
}

func TestMiddleware_RateLimitBlocks429(t *testing.T) {
	em := NewEdgeMetrics(prometheus.NewRegistry())
	rl := NewIngressRateLimiter(RateLimitConfig{DefaultRPS: 1, DefaultBurst: 1, MaxConcurrency: 100, MaxBodyBytes: 1 << 20})
	mw := NewIngressMiddleware(NoopWAFEngine{}, rl, em)
	h := mw.Wrap("acme", okHandler(nil))

	rec1 := httptest.NewRecorder()
	h.ServeHTTP(rec1, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec1.Code != http.StatusOK {
		t.Fatalf("first request want 200, got %d", rec1.Code)
	}

	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec2.Code != http.StatusTooManyRequests {
		t.Fatalf("second request want 429, got %d", rec2.Code)
	}
}

func TestMiddleware_ConcurrencyBlocks503(t *testing.T) {
	em := NewEdgeMetrics(prometheus.NewRegistry())
	rl := NewIngressRateLimiter(RateLimitConfig{DefaultRPS: 1000, DefaultBurst: 1000, MaxConcurrency: 1, MaxBodyBytes: 1 << 20})

	release := make(chan struct{})
	entered := make(chan struct{})
	var once sync.Once
	holder := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		once.Do(func() { close(entered) })
		<-release
		w.WriteHeader(http.StatusOK)
	})

	mw := NewIngressMiddleware(NoopWAFEngine{}, rl, em)
	h := mw.Wrap("acme", holder)

	// First request occupies the only concurrency slot.
	go func() {
		h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
	}()
	select {
	case <-entered:
	case <-time.After(2 * time.Second):
		t.Fatal("first handler never entered")
	}

	// Second concurrent request must be rejected with 503.
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("second concurrent request want 503, got %d", rec.Code)
	}
	close(release)
}

func TestMiddleware_WAFBlocksInBlockingMode(t *testing.T) {
	em := NewEdgeMetrics(prometheus.NewRegistry())
	waf := stubWAF{block: true, mode: ModeBlocking, matched: []string{"942100"}}
	mw := NewIngressMiddleware(waf, permissiveRL(), em)

	called := false
	h := mw.Wrap("acme", okHandler(&called))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/?id=1'+OR+'1'='1", nil))

	if called {
		t.Fatal("handler should not be called when WAF blocks")
	}
	if rec.Code != http.StatusForbidden {
		t.Fatalf("want 403, got %d", rec.Code)
	}
}

func TestMiddleware_WAFDetectOnlyAllowsThrough(t *testing.T) {
	em := NewEdgeMetrics(prometheus.NewRegistry())
	// block=false simulates detection mode: matches recorded, not blocked.
	waf := stubWAF{block: false, mode: ModeDetection, matched: []string{"942100"}}
	mw := NewIngressMiddleware(waf, permissiveRL(), em)

	called := false
	h := mw.Wrap("acme", okHandler(&called))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/?id=1'+OR+'1'='1", nil))

	if !called {
		t.Fatal("handler should be called in detection mode despite a rule match")
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("detection mode want 200, got %d", rec.Code)
	}
}

func TestMiddleware_BodySizeCapEnforced413(t *testing.T) {
	em := NewEdgeMetrics(prometheus.NewRegistry())
	rl := NewIngressRateLimiter(RateLimitConfig{DefaultRPS: 1000, DefaultBurst: 1000, MaxConcurrency: 100, MaxBodyBytes: 65536})
	mw := NewIngressMiddleware(NoopWAFEngine{}, rl, em)

	called := false
	h := mw.Wrap("acme", okHandler(&called))

	big := bytes.Repeat([]byte("A"), 2<<20) // 2 MiB > 64 KiB cap
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/upload", bytes.NewReader(big))
	h.ServeHTTP(rec, req)

	if called {
		t.Fatal("handler should not be called when body exceeds cap")
	}
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("want 413, got %d", rec.Code)
	}
}

func TestMiddleware_BodyPreservedForUpstream(t *testing.T) {
	em := NewEdgeMetrics(prometheus.NewRegistry())
	mw := NewIngressMiddleware(stubWAF{}, permissiveRL(), em)

	var seen []byte
	h := mw.Wrap("acme", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))

	payload := "hello upstream"
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/", strings.NewReader(payload)))

	if string(seen) != payload {
		t.Fatalf("upstream body: want %q, got %q", payload, string(seen))
	}
}

// hijackRecorder is an httptest.ResponseRecorder that also implements
// http.Hijacker, so we can assert the middleware wrapper preserves the
// interface for the WebSocket upgrade path.
type hijackRecorder struct {
	*httptest.ResponseRecorder
	hijacked bool
}

func (h *hijackRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h.hijacked = true
	c1, c2 := net.Pipe()
	_ = c2.Close()
	return c1, bufio.NewReadWriter(bufio.NewReader(c1), bufio.NewWriter(c1)), nil
}

func TestMiddleware_HijackPreservedForWebSocket(t *testing.T) {
	em := NewEdgeMetrics(prometheus.NewRegistry())
	mw := NewIngressMiddleware(NoopWAFEngine{}, permissiveRL(), em)

	var sawHijacker bool
	h := mw.Wrap("acme", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hj, ok := w.(http.Hijacker)
		if !ok {
			return
		}
		sawHijacker = true
		conn, _, err := hj.Hijack()
		if err == nil && conn != nil {
			_ = conn.Close()
		}
	}))

	rec := &hijackRecorder{ResponseRecorder: httptest.NewRecorder()}
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if !sawHijacker {
		t.Fatal("response writer wrapper must still satisfy http.Hijacker")
	}
	if !rec.hijacked {
		t.Fatal("Hijack should delegate to the underlying ResponseWriter")
	}
}

func TestMiddleware_AllowWebSocket_GatesAndGauge(t *testing.T) {
	reg := prometheus.NewRegistry()
	em := NewEdgeMetrics(reg)
	// MaxConcurrency 1 so the second concurrent upgrade is rejected.
	rl := NewIngressRateLimiter(RateLimitConfig{
		DefaultRPS: 1000, DefaultBurst: 1000, MaxConcurrency: 1, MaxBodyBytes: 1 << 20,
	})
	mw := NewIngressMiddleware(NoopWAFEngine{}, rl, em)

	gauge := func() float64 {
		return testutil.ToFloat64(em.tunnelSessionsActive.WithLabelValues("acme", "forward", "default"))
	}

	// First upgrade is admitted: gauge goes to 1.
	rec1 := httptest.NewRecorder()
	release1, ok1 := mw.AllowWebSocket("acme", "forward", "default", rec1)
	if !ok1 {
		t.Fatalf("first websocket upgrade should be admitted, got %d", rec1.Code)
	}
	if g := gauge(); g != 1 {
		t.Fatalf("active gauge after first upgrade: want 1, got %v", g)
	}

	// Second concurrent upgrade exceeds MaxConcurrency=1 -> 503, gauge unchanged.
	rec2 := httptest.NewRecorder()
	_, ok2 := mw.AllowWebSocket("acme", "forward", "default", rec2)
	if ok2 {
		t.Fatal("second concurrent websocket upgrade must be rejected")
	}
	if rec2.Code != http.StatusServiceUnavailable {
		t.Fatalf("rejected upgrade want 503, got %d", rec2.Code)
	}
	if g := gauge(); g != 1 {
		t.Fatalf("active gauge should stay 1 while one session holds the slot, got %v", g)
	}

	// Releasing the first frees the slot and decrements the gauge.
	release1()
	if g := gauge(); g != 0 {
		t.Fatalf("active gauge after release: want 0, got %v", g)
	}

	// A subsequent upgrade now succeeds (slot freed).
	rec3 := httptest.NewRecorder()
	release3, ok3 := mw.AllowWebSocket("acme", "forward", "default", rec3)
	if !ok3 {
		t.Fatalf("upgrade after release should be admitted, got %d", rec3.Code)
	}
	release3()
	// Double-release must be a safe no-op (gauge does not go negative).
	release3()
	if g := gauge(); g != 0 {
		t.Fatalf("active gauge after double release: want 0, got %v", g)
	}
}

func TestMiddleware_AllowWebSocket_RateLimit429(t *testing.T) {
	reg := prometheus.NewRegistry()
	em := NewEdgeMetrics(reg)
	rl := NewIngressRateLimiter(RateLimitConfig{
		DefaultRPS: 1, DefaultBurst: 1, MaxConcurrency: 100, MaxBodyBytes: 1 << 20,
	})
	mw := NewIngressMiddleware(NoopWAFEngine{}, rl, em)

	// First consumes the only token.
	rec1 := httptest.NewRecorder()
	release1, ok1 := mw.AllowWebSocket("acme", "reverse", "default", rec1)
	if !ok1 {
		t.Fatalf("first upgrade should be admitted, got %d", rec1.Code)
	}
	defer release1()

	// Second is rate-limited (429).
	rec2 := httptest.NewRecorder()
	_, ok2 := mw.AllowWebSocket("acme", "reverse", "default", rec2)
	if ok2 {
		t.Fatal("second upgrade must be rate-limited")
	}
	if rec2.Code != http.StatusTooManyRequests {
		t.Fatalf("rate-limited upgrade want 429, got %d", rec2.Code)
	}
}

func TestMiddleware_FlushPreserved(t *testing.T) {
	w := newResponseWriterWrapper(httptest.NewRecorder())
	if _, ok := interface{}(w).(http.Flusher); !ok {
		t.Fatal("wrapper must implement http.Flusher")
	}
	if _, ok := interface{}(w).(http.Hijacker); !ok {
		t.Fatal("wrapper must implement http.Hijacker")
	}
	// Flush on a recorder (non-Flusher underlying) must be a safe no-op.
	w.Flush()
}
