package ingress

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// newTestCoraza builds a real CorazaWAF with the embedded OWASP CRS. The first
// build loads the full rule set (~25MB) so this is reused across subtests where
// possible.
func newTestCoraza(t *testing.T, mode string) *CorazaWAF {
	t.Helper()
	waf, err := NewCorazaWAF(WAFConfig{
		Enabled:        true,
		Mode:           mode,
		BodyLimitBytes: 65536,
	})
	if err != nil {
		t.Fatalf("NewCorazaWAF: %v", err)
	}
	return waf
}

func TestCorazaWAF_SQLInjectionBlocked(t *testing.T) {
	waf := newTestCoraza(t, "blocking")

	// Classic SQLi probe in the query string.
	r := httptest.NewRequest(http.MethodGet, "/?id=1'%20OR%20'1'='1", nil)
	r.Host = "tenant.moltbunker.dev"

	res, err := waf.Inspect(r, "tenant")
	if err != nil {
		t.Fatalf("Inspect: %v", err)
	}
	if !res.Blocked {
		t.Fatalf("expected SQLi to be blocked in blocking mode, got Blocked=false; matched=%v", res.MatchedRules)
	}
	if len(res.MatchedRules) == 0 {
		t.Fatalf("expected at least one matched rule for SQLi")
	}
}

func TestCorazaWAF_CleanRequestPasses(t *testing.T) {
	waf := newTestCoraza(t, "blocking")

	r := httptest.NewRequest(http.MethodGet, "/api/v1/status", nil)
	r.Host = "tenant.moltbunker.dev"
	r.Header.Set("User-Agent", "curl/8.0.0")
	r.Header.Set("Accept", "application/json")

	res, err := waf.Inspect(r, "tenant")
	if err != nil {
		t.Fatalf("Inspect: %v", err)
	}
	if res.Blocked {
		t.Fatalf("expected clean request to pass, got Blocked=true; matched=%v", res.MatchedRules)
	}
	// A benign request must record ZERO matched rules: the CRS housekeeping /
	// initialization SecActions (900xxx/901xxx setup, NNN013/NNN014 markers)
	// must be filtered out by collectMatches so the detection signal and
	// metrics are not polluted by baseline bootstrap noise. (Regression guard
	// for the EDGE-01 over-counting bug: a clean request previously reported
	// ~62 "matched" rules.)
	if len(res.MatchedRules) != 0 {
		t.Fatalf("clean request must record 0 detection matches, got %d: %v",
			len(res.MatchedRules), res.MatchedRules)
	}
}

// TestCorazaWAF_AttackDetectedInDetectionMode confirms that real attacks are
// still recorded (non-zero matches) even after the housekeeping-rule filter,
// in both detection and blocking modes. Pairs with
// TestCorazaWAF_CleanRequestPasses (zero detections on benign traffic).
func TestCorazaWAF_AttackDetectedInDetectionMode(t *testing.T) {
	waf := newTestCoraza(t, "detection")

	cases := []struct {
		name string
		uri  string
	}{
		{"path_traversal", "/?x=../../etc/passwd"},
		{"sqli_probe", "/?id=1'%20OR%20'1'='1"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, tc.uri, nil)
			r.Host = "tenant.moltbunker.dev"

			res, err := waf.Inspect(r, "tenant")
			if err != nil {
				t.Fatalf("Inspect: %v", err)
			}
			if res.Blocked {
				t.Fatalf("detection mode must not block")
			}
			if len(res.MatchedRules) == 0 {
				t.Fatalf("expected the %s attack to be recorded as a detection", tc.name)
			}
		})
	}
}

func TestCorazaWAF_DetectionModeDoesNotBlock(t *testing.T) {
	waf := newTestCoraza(t, "detection")

	r := httptest.NewRequest(http.MethodGet, "/?id=1'%20OR%20'1'='1", nil)
	r.Host = "tenant.moltbunker.dev"

	res, err := waf.Inspect(r, "tenant")
	if err != nil {
		t.Fatalf("Inspect: %v", err)
	}
	if res.Blocked {
		t.Fatalf("detection mode must not block, got Blocked=true")
	}
	if len(res.MatchedRules) == 0 {
		t.Fatalf("detection mode should still record matched rules for SQLi")
	}
}

func TestCorazaWAF_TenantOverrideBlocking(t *testing.T) {
	// Global default detection, but a specific tenant set to blocking.
	waf, err := NewCorazaWAF(WAFConfig{
		Enabled:        true,
		Mode:           "detection",
		BodyLimitBytes: 65536,
		TenantOverrides: map[string]TenantWAFConfig{
			"strict": {Mode: "blocking"},
		},
	})
	if err != nil {
		t.Fatalf("NewCorazaWAF: %v", err)
	}

	if got := waf.Mode("strict"); got != ModeBlocking {
		t.Fatalf("tenant override: want blocking, got %s", got)
	}
	if got := waf.Mode("relaxed"); got != ModeDetection {
		t.Fatalf("default mode: want detection, got %s", got)
	}

	r := httptest.NewRequest(http.MethodGet, "/?id=1'%20OR%20'1'='1", nil)
	res, err := waf.Inspect(r, "strict")
	if err != nil {
		t.Fatalf("Inspect: %v", err)
	}
	if !res.Blocked {
		t.Fatalf("expected strict tenant to block SQLi")
	}
}

func TestCorazaWAF_BodyLimitTruncates(t *testing.T) {
	waf, err := NewCorazaWAF(WAFConfig{
		Enabled:        true,
		Mode:           "detection",
		BodyLimitBytes: 4096,
	})
	if err != nil {
		t.Fatalf("NewCorazaWAF: %v", err)
	}

	body := bytes.Repeat([]byte("A"), 1<<20) // 1 MiB
	r := httptest.NewRequest(http.MethodPost, "/upload", bytes.NewReader(body))
	r.Host = "tenant.moltbunker.dev"
	r.Header.Set("Content-Type", "text/plain")

	res, err := waf.Inspect(r, "tenant")
	if err != nil {
		t.Fatalf("Inspect: %v", err)
	}
	if !res.Truncated {
		t.Fatalf("expected Truncated=true for 1MiB body with 4KiB cap")
	}
}

func TestNoopWAFEngine_NeverBlocks(t *testing.T) {
	var waf WAFEngine = NoopWAFEngine{}

	cases := []*http.Request{
		httptest.NewRequest(http.MethodGet, "/?id=1'%20OR%20'1'='1", nil),
		httptest.NewRequest(http.MethodPost, "/x", strings.NewReader("<script>alert(1)</script>")),
		httptest.NewRequest(http.MethodGet, "/etc/passwd", nil),
	}
	for i, r := range cases {
		res, err := waf.Inspect(r, "tenant")
		if err != nil {
			t.Fatalf("case %d: Inspect: %v", i, err)
		}
		if res.Blocked {
			t.Fatalf("case %d: NoopWAFEngine must never block", i)
		}
	}
	if waf.Mode("tenant") != ModeDetection {
		t.Fatalf("NoopWAFEngine mode should be detection")
	}
}
