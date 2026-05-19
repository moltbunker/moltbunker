package runtime

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// fakeScanner is a test double for ImageScanner. It returns a pre-configured
// report and counts invocations so we can verify caching behavior.
type fakeScanner struct {
	id       string
	report   *ScanReport
	err      error
	calls    int64
	delay    time.Duration
	perRef   map[string]*ScanReport // optional per-ref override
	perRefMu sync.Mutex
}

func (f *fakeScanner) ID() string { return f.id }

func (f *fakeScanner) Scan(ctx context.Context, imageRef string) (*ScanReport, error) {
	atomic.AddInt64(&f.calls, 1)
	if f.delay > 0 {
		select {
		case <-time.After(f.delay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	f.perRefMu.Lock()
	if r, ok := f.perRef[imageRef]; ok {
		f.perRefMu.Unlock()
		return r, nil
	}
	f.perRefMu.Unlock()
	return f.report, f.err
}

func TestScanPolicy_Apply(t *testing.T) {
	findings := []Vulnerability{
		{ID: "CVE-2024-0001", Severity: SeverityLow},
		{ID: "CVE-2024-0002", Severity: SeverityMedium},
		{ID: "CVE-2024-0003", Severity: SeverityHigh},
		{ID: "CVE-2024-0004", Severity: SeverityCritical},
	}

	cases := []struct {
		name        string
		policy      ScanPolicy
		findings    []Vulnerability
		wantBlocked int
		wantErr     bool
	}{
		{
			name:        "block on high and critical (default)",
			policy:      DefaultScanPolicy(),
			findings:    findings,
			wantBlocked: 2,
			wantErr:     true,
		},
		{
			name:        "block on critical only",
			policy:      ScanPolicy{BlockAtOrAbove: SeverityCritical},
			findings:    findings,
			wantBlocked: 1,
			wantErr:     true,
		},
		{
			name:        "ignore one high CVE removes it from blocking set",
			policy:      ScanPolicy{BlockAtOrAbove: SeverityHigh, IgnoreCVEs: []string{"CVE-2024-0003"}},
			findings:    findings,
			wantBlocked: 1, // only critical remains
			wantErr:     true,
		},
		{
			name:        "ignore both high and critical → pass",
			policy:      ScanPolicy{BlockAtOrAbove: SeverityHigh, IgnoreCVEs: []string{"CVE-2024-0003", "CVE-2024-0004"}},
			findings:    findings,
			wantBlocked: 0,
			wantErr:     false,
		},
		{
			name:        "no findings → pass",
			policy:      DefaultScanPolicy(),
			findings:    []Vulnerability{},
			wantBlocked: 0,
			wantErr:     false,
		},
		{
			name:        "low only with default policy → pass",
			policy:      DefaultScanPolicy(),
			findings:    findings[:1],
			wantBlocked: 0,
			wantErr:     false,
		},
		{
			name:        "zero-value policy treats as default",
			policy:      ScanPolicy{}, // BlockAtOrAbove unset
			findings:    findings,
			wantBlocked: 2, // high + critical
			wantErr:     true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			blocked, err := tc.policy.Apply(tc.findings)
			if (err != nil) != tc.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tc.wantErr)
			}
			if len(blocked) != tc.wantBlocked {
				t.Fatalf("blocked = %d findings, want %d (set: %v)", len(blocked), tc.wantBlocked, blocked)
			}
			if tc.wantErr && !errors.Is(err, ErrPolicyViolation) {
				t.Fatalf("err = %v, expected to wrap ErrPolicyViolation", err)
			}
		})
	}
}

func TestCachedScanner_DedupesConcurrentCalls(t *testing.T) {
	inner := &fakeScanner{
		id:     "fake",
		report: &ScanReport{ImageRef: "alpine:3", ScannerID: "fake"},
		delay:  20 * time.Millisecond,
	}
	cached := NewCachedScanner(inner)

	var wg sync.WaitGroup
	const N = 10
	for i := 0; i < N; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = cached.Scan(context.Background(), "alpine:3")
		}()
	}
	wg.Wait()

	if got := atomic.LoadInt64(&inner.calls); got != 1 {
		t.Fatalf("inner scanner called %d times, want exactly 1 (singleflight)", got)
	}
}

func TestCachedScanner_DifferentRefs_BothScanned(t *testing.T) {
	inner := &fakeScanner{
		id:     "fake",
		report: &ScanReport{ScannerID: "fake"},
	}
	cached := NewCachedScanner(inner)

	if _, err := cached.Scan(context.Background(), "alpine:3"); err != nil {
		t.Fatal(err)
	}
	if _, err := cached.Scan(context.Background(), "nginx:latest"); err != nil {
		t.Fatal(err)
	}

	if got := atomic.LoadInt64(&inner.calls); got != 2 {
		t.Fatalf("inner scanner called %d times, want 2 (one per distinct ref)", got)
	}
}

func TestCachedScanner_InvalidateForcesRescan(t *testing.T) {
	inner := &fakeScanner{
		id:     "fake",
		report: &ScanReport{ScannerID: "fake"},
	}
	cached := NewCachedScanner(inner)

	if _, err := cached.Scan(context.Background(), "alpine:3"); err != nil {
		t.Fatal(err)
	}
	cached.Invalidate("alpine:3")
	if _, err := cached.Scan(context.Background(), "alpine:3"); err != nil {
		t.Fatal(err)
	}

	if got := atomic.LoadInt64(&inner.calls); got != 2 {
		t.Fatalf("inner scanner called %d times, want 2 after invalidate", got)
	}
}

func TestNoopScanner_AlwaysClean(t *testing.T) {
	s := NewNoopScanner()
	report, err := s.Scan(context.Background(), "anything")
	if err != nil {
		t.Fatalf("noop scanner errored: %v", err)
	}
	if len(report.Vulnerabilities) != 0 {
		t.Fatalf("noop scanner returned findings: %v", report.Vulnerabilities)
	}
	if _, err := DefaultScanPolicy().Apply(report.Vulnerabilities); err != nil {
		t.Fatalf("default policy should pass empty report, got %v", err)
	}
}

func TestSeverityRanking(t *testing.T) {
	cases := []struct {
		a, b Severity
		want bool // a >= b
	}{
		{SeverityCritical, SeverityHigh, true},
		{SeverityHigh, SeverityCritical, false},
		{SeverityMedium, SeverityMedium, true},
		{SeverityUnknown, SeverityLow, false},
		{SeverityLow, SeverityUnknown, true},
	}
	for _, tc := range cases {
		got := severityRank(tc.a) >= severityRank(tc.b)
		if got != tc.want {
			t.Fatalf("severity %s >= %s = %v, want %v", tc.a, tc.b, got, tc.want)
		}
	}
}

// Smoke test for the TrivyCLIScanner when the trivy binary is missing — we
// expect ErrScannerUnavailable, not a panic.
func TestTrivyCLIScanner_MissingBinary(t *testing.T) {
	s := &TrivyCLIScanner{
		BinaryPath: "/nonexistent/path/trivy-not-here",
		Timeout:    1 * time.Second,
	}
	_, err := s.Scan(context.Background(), "alpine:3")
	if err == nil {
		t.Fatal("expected error for missing binary")
	}
	if !errors.Is(err, ErrScannerUnavailable) {
		t.Fatalf("err = %v, want ErrScannerUnavailable", err)
	}
}

// TestScanReport_ToError exercises the full ScanPolicy → blocking-findings
// → human-readable error path, so deploy-time log lines stay coherent.
func TestScanReport_PolicyErrorMessage(t *testing.T) {
	findings := []Vulnerability{
		{ID: "CVE-2024-9999", Severity: SeverityCritical, Package: "openssl", Version: "1.0.0"},
	}
	_, err := DefaultScanPolicy().Apply(findings)
	if err == nil {
		t.Fatal("expected error")
	}
	if msg := fmt.Sprint(err); msg == "" {
		t.Fatal("error message empty")
	}
}
