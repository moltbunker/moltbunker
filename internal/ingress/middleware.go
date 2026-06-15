package ingress

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/moltbunker/moltbunker/internal/config"
	"github.com/moltbunker/moltbunker/internal/logging"
)

// IngressMiddleware chains the three edge components — per-tenant abuse limits,
// the L7 WAF, and edge observability — in front of the proxy dispatch handler.
//
// The chain order is: rate limit -> concurrency gate -> body size cap -> WAF
// inspection -> dispatch. Each layer records its own metrics. When the WAF is
// in detection mode a match is logged and counted but the request proceeds.
//
// IngressMiddleware is safe for concurrent use; per-request state is local.
type IngressMiddleware struct {
	waf     WAFEngine
	rl      *IngressRateLimiter
	metrics *EdgeMetrics
}

// NewIngressMiddleware builds the middleware. Any of waf/rl/metrics may be nil:
// a nil WAF becomes a NoopWAFEngine, a nil limiter disables abuse controls, and
// nil metrics disables observability (the EdgeMetrics methods are nil-safe).
func NewIngressMiddleware(waf WAFEngine, rl *IngressRateLimiter, metrics *EdgeMetrics) *IngressMiddleware {
	if waf == nil {
		waf = NoopWAFEngine{}
	}
	return &IngressMiddleware{waf: waf, rl: rl, metrics: metrics}
}

// NewIngressMiddlewareFromConfig builds the full edge chain from operator
// config: it constructs the WAF engine (real Coraza+CRS when WAF.Enabled, else
// a no-op), the per-tenant rate limiter (whose background cleanup goroutine is
// started), and the edge metrics (registered into reg; a private registry is
// used when reg is nil).
//
// It returns (nil, nil) only if there is nothing to install — currently it
// always installs the chain because rate limiting is always on by default. A
// non-nil error is returned only if the Coraza engine fails to initialize.
func NewIngressMiddlewareFromConfig(cfg config.IngressConfig, reg *prometheus.Registry) (*IngressMiddleware, error) {
	waf, err := NewWAFEngine(WAFConfig{
		Enabled:        cfg.WAF.Enabled,
		Mode:           cfg.WAF.Mode,
		BodyLimitBytes: cfg.WAF.BodyLimitBytes,
		ExcludeRuleIDs: cfg.WAF.ExcludeRuleIDs,
	})
	if err != nil {
		return nil, err
	}

	rl := NewIngressRateLimiter(RateLimitConfig{
		DefaultRPS:     cfg.RateLimit.DefaultRPS,
		DefaultBurst:   cfg.RateLimit.DefaultBurst,
		MaxConcurrency: cfg.RateLimit.MaxConcurrency,
		MaxBodyBytes:   cfg.RateLimit.MaxBodyBytes,
	})
	rl.Start()

	em := NewEdgeMetrics(reg)
	return NewIngressMiddleware(waf, rl, em), nil
}

// Wrap returns an http.Handler that runs the edge chain for tenantID and then
// invokes next on success.
func (m *IngressMiddleware) Wrap(tenantID string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.serve(tenantID, w, r, next)
	})
}

// AllowWebSocket applies the abuse-control gates that are compatible with the
// hijack-and-stream WebSocket path: the per-tenant request rate limit and the
// concurrency cap. The body-buffering / WAF steps are intentionally skipped
// (they are incompatible with a hijacked, full-duplex connection), but a
// WebSocket upgrade still consumes a rate token and a concurrency slot so a
// tenant subdomain cannot be used to open unbounded concurrent long-lived
// connections (the exact abuse case the concurrency cap exists for).
//
// On rejection it writes the appropriate 429/503 response, records the metric,
// and returns ok=false with a no-op release. On success it returns ok=true and
// a release closure that the caller MUST defer until the hijacked connection
// closes; release frees the concurrency slot and decrements the active
// tunnel-session gauge. release is safe to call exactly once.
//
// tunnelType is "forward" or "reverse"; tier is the edge-provider tier label
// (use "default" when unknown). These feed the tunnel_sessions_active gauge.
func (m *IngressMiddleware) AllowWebSocket(tenantID, tunnelType, tier string, w http.ResponseWriter) (release func(), ok bool) {
	if m.rl != nil {
		if !m.rl.Allow(tenantID) {
			m.metrics.RecordRateLimit(tenantID, rateLimitReasonRPS)
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return func() {}, false
		}
		concRelease, acquired := m.rl.AcquireConcurrency(tenantID)
		if !acquired {
			m.metrics.RecordRateLimit(tenantID, rateLimitReasonConcurrency)
			http.Error(w, "too many concurrent requests", http.StatusServiceUnavailable)
			return func() {}, false
		}
		m.metrics.SetTunnelSession(tenantID, tunnelType, tier, true)
		var once sync.Once
		return func() {
			once.Do(func() {
				concRelease()
				m.metrics.SetTunnelSession(tenantID, tunnelType, tier, false)
			})
		}, true
	}
	// No limiter: still track the active session for observability.
	m.metrics.SetTunnelSession(tenantID, tunnelType, tier, true)
	var once sync.Once
	return func() {
		once.Do(func() { m.metrics.SetTunnelSession(tenantID, tunnelType, tier, false) })
	}, true
}

func (m *IngressMiddleware) serve(tenantID string, w http.ResponseWriter, r *http.Request, next http.Handler) {
	// 1. Per-tenant request rate (token bucket).
	if m.rl != nil && !m.rl.Allow(tenantID) {
		m.metrics.RecordRateLimit(tenantID, rateLimitReasonRPS)
		http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
		return
	}

	// 2. Per-tenant concurrency cap.
	if m.rl != nil {
		release, ok := m.rl.AcquireConcurrency(tenantID)
		if !ok {
			m.metrics.RecordRateLimit(tenantID, rateLimitReasonConcurrency)
			http.Error(w, "too many concurrent requests", http.StatusServiceUnavailable)
			return
		}
		defer release()
	}

	// 3. Body size cap + WAF body buffering.
	//
	// We must read the body once for the WAF and still hand the full body to
	// the upstream proxy (proxyHTTP does r.Write(tun)). So we buffer the body
	// here under the hard MaxBodyBytes cap and replace r.Body with a reader
	// over the buffered bytes. A body that exceeds the cap yields 413.
	var reqBodyLen int64
	var matchedWAF bool
	if r.Body != nil {
		maxBytes := int64(0)
		if m.rl != nil {
			maxBytes = m.rl.MaxBodyBytes()
		}
		buffered, tooBig, err := readCappedBody(r.Body, maxBytes)
		_ = r.Body.Close()
		if tooBig {
			m.metrics.RecordRateLimit(tenantID, rateLimitReasonBodySize)
			http.Error(w, "request entity too large", http.StatusRequestEntityTooLarge)
			return
		}
		if err != nil {
			http.Error(w, "error reading request body", http.StatusBadRequest)
			return
		}
		reqBodyLen = int64(len(buffered))
		// Restore a re-readable body for the WAF + upstream.
		r.Body = io.NopCloser(bytes.NewReader(buffered))
		// Run the WAF over a separate reader so r.Body stays at offset 0 for
		// the upstream write.
		r.ContentLength = reqBodyLen
		wafReq := r.Clone(r.Context())
		wafReq.Body = io.NopCloser(bytes.NewReader(buffered))
		blocked, matched := m.inspect(tenantID, wafReq, w)
		if blocked {
			return
		}
		matchedWAF = matched
	} else {
		blocked, matched := m.inspect(tenantID, r, w)
		if blocked {
			return
		}
		matchedWAF = matched
	}

	// 4. Dispatch with a response wrapper to capture status + bytes for metrics.
	rw := newResponseWriterWrapper(w)
	next.ServeHTTP(rw, r)

	// 5. Observability.
	m.metrics.RecordRequest(tenantID, r.Method, rw.status, rw.elapsed(), reqBodyLen, rw.bytesWritten, matchedWAF)
}

// inspect runs the WAF and returns blocked=true if the request was blocked (and
// a 403 has been written), plus matched=true if any WAF rule fired (regardless
// of mode). Detection-mode matches are recorded but not blocked.
func (m *IngressMiddleware) inspect(tenantID string, r *http.Request, w http.ResponseWriter) (blocked, matched bool) {
	res, err := m.waf.Inspect(r, tenantID)
	if err != nil {
		// Fail-open: a WAF error must not take down ingress. Log and proceed.
		logging.Warn("waf inspection error (failing open)",
			"subdomain", tenantID,
			logging.Err(err),
			logging.Component("ingress"))
		m.metrics.RecordWAFVerdict(tenantID, wafVerdictPass)
		return false, false
	}

	for _, id := range res.MatchedRules {
		m.metrics.RecordWAFRuleMatch(tenantID, id, res.MatchedPhases[id])
	}
	if res.Truncated {
		m.metrics.RecordWAFVerdict(tenantID, wafVerdictTruncated)
	}
	matched = len(res.MatchedRules) > 0

	if res.Blocked {
		m.metrics.RecordWAFVerdict(tenantID, wafVerdictBlock)
		logging.Info("waf blocked request",
			"subdomain", tenantID,
			"phase", res.Phase,
			"matched_rules", len(res.MatchedRules),
			logging.Component("ingress"))
		http.Error(w, "forbidden", http.StatusForbidden)
		return true, matched
	}

	if matched {
		// Detection mode (or blocking-mode non-disruptive match): record but
		// allow through.
		m.metrics.RecordWAFVerdict(tenantID, wafVerdictDetect)
	} else {
		m.metrics.RecordWAFVerdict(tenantID, wafVerdictPass)
	}
	return false, matched
}

// readCappedBody reads up to maxBytes from body. If maxBytes <= 0 there is no
// cap. It returns tooBig=true (and discards the buffer) when the body exceeds
// maxBytes.
func readCappedBody(body io.Reader, maxBytes int64) (buf []byte, tooBig bool, err error) {
	if maxBytes <= 0 {
		b, rerr := io.ReadAll(body)
		return b, false, rerr
	}
	// Read one extra byte to detect overflow.
	limited := io.LimitReader(body, maxBytes+1)
	b, rerr := io.ReadAll(limited)
	if rerr != nil {
		return nil, false, rerr
	}
	if int64(len(b)) > maxBytes {
		return nil, true, nil
	}
	return b, false, nil
}

// responseWriterWrapper captures the status code and bytes written so the
// middleware can emit per-request metrics. It preserves http.Flusher and
// http.Hijacker (the proxy hijacks for WebSocket) by delegating to the
// underlying ResponseWriter.
type responseWriterWrapper struct {
	http.ResponseWriter
	status       int
	bytesWritten int64
	wroteHeader  bool
	start        time.Time
}

func newResponseWriterWrapper(w http.ResponseWriter) *responseWriterWrapper {
	return &responseWriterWrapper{
		ResponseWriter: w,
		status:         http.StatusOK,
		start:          time.Now(),
	}
}

func (w *responseWriterWrapper) WriteHeader(code int) {
	if !w.wroteHeader {
		w.status = code
		w.wroteHeader = true
	}
	w.ResponseWriter.WriteHeader(code)
}

func (w *responseWriterWrapper) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	n, err := w.ResponseWriter.Write(b)
	w.bytesWritten += int64(n)
	return n, err
}

// Flush implements http.Flusher, delegating when the underlying writer
// supports it (the proxy streams response bodies and flushes).
func (w *responseWriterWrapper) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// Hijack implements http.Hijacker so the WebSocket upgrade path in proxy.go
// (handleWebSocket / handleWebSocketViaStream) keeps working through the
// middleware.
func (w *responseWriterWrapper) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hj, ok := w.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("ingress: underlying ResponseWriter does not support hijacking")
	}
	return hj.Hijack()
}

func (w *responseWriterWrapper) elapsed() time.Duration {
	return time.Since(w.start)
}
