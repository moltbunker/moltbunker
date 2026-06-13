package ingress

import "net/http"

// NoopWAFEngine is a zero-cost WAFEngine used when the WAF is disabled
// (config.Ingress.WAF.Enabled=false). It never blocks and never loads the
// OWASP CRS, so the ~25MB Coraza pattern-matcher state is not allocated.
//
// Holding a non-nil NoopWAFEngine instead of a nil WAFEngine keeps the request
// hot path free of nil checks: the caller always invokes Inspect.
type NoopWAFEngine struct{}

// Inspect always reports the request as allowed.
func (NoopWAFEngine) Inspect(_ *http.Request, _ string) (*WAFResult, error) {
	return &WAFResult{}, nil
}

// Mode always reports detection (the no-op never enforces).
func (NoopWAFEngine) Mode(_ string) WAFMode { return ModeDetection }
