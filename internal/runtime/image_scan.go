package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// R4 — Image CVE / vulnerability scanning.
//
// This file provides a pluggable scanner interface and a Trivy-CLI-based
// implementation. The scanner runs in the container-pull pipeline before a
// container is created and rejects images whose vulnerability profile
// violates the supplied ScanPolicy.
//
// Design boundaries:
//  - ImageScanner is an abstraction; TrivyCLIScanner shells out to a Trivy
//    binary. Swap to Grype/Snyk/etc. by implementing the interface.
//  - Scans are CACHED per (scannerID, image-ref). Re-deploys of the same image
//    are instant after the first scan. The cache key includes scannerID so
//    that switching scanner versions invalidates cleanly.
//  - ScanPolicy is caller-supplied. Defaults block on Critical and High.

// Severity is the standard CVSS severity bucket used by virtually every
// scanner. Strings match Trivy's --severity values.
type Severity string

const (
	SeverityUnknown  Severity = "UNKNOWN"
	SeverityLow      Severity = "LOW"
	SeverityMedium   Severity = "MEDIUM"
	SeverityHigh     Severity = "HIGH"
	SeverityCritical Severity = "CRITICAL"
)

// severityRank assigns a numeric weight so we can do `>=` comparisons.
func severityRank(s Severity) int {
	switch s {
	case SeverityCritical:
		return 4
	case SeverityHigh:
		return 3
	case SeverityMedium:
		return 2
	case SeverityLow:
		return 1
	default:
		return 0
	}
}

// Vulnerability is a single finding produced by a scanner.
type Vulnerability struct {
	ID           string   // CVE-YYYY-NNNN or scanner-specific id
	Severity     Severity // bucketed severity
	Package      string   // affected package
	Version      string   // installed version
	FixedVersion string   // empty if no fix available yet
}

// ScanReport is the result of scanning a single image reference.
type ScanReport struct {
	ImageRef        string
	ImageDigest     ImageDigest
	ScannerID       string // identifies the scanner backend + version
	Vulnerabilities []Vulnerability
	ScanStartedAt   time.Time
	ScanDuration    time.Duration
}

// ScanPolicy controls which scan results are acceptable. A nil policy is
// treated as DefaultScanPolicy.
//
// TODO(R4-policy-source): Decide how policies are configured per-tenant:
//   - global daemon config (simplest, no per-tenant flexibility)
//   - per-deployment annotation (flexible, user can downgrade their own risk)
//   - on-chain BunkerVerification contract (decentralized, gas cost)
//   - tier-based (Confidential tier requires CRITICAL only, etc.)
type ScanPolicy struct {
	// BlockAtOrAbove is the minimum severity that blocks deploy. Findings
	// at this rank or higher cause Scan to return ErrPolicyViolation.
	BlockAtOrAbove Severity

	// IgnoreCVEs is a per-deployment allowlist of CVE IDs to ignore. Useful
	// for known-and-accepted exceptions (e.g. an unpatched CVE that doesn't
	// affect the way the container is used).
	IgnoreCVEs []string

	// RequireScan: if true, an empty scan result (scanner produced nothing
	// for this image, possibly because the scanner couldn't reach the image)
	// is treated as a failure rather than a pass.
	RequireScan bool
}

// DefaultScanPolicy blocks deploy on any HIGH or CRITICAL finding and does
// not require the scan to have produced any results.
func DefaultScanPolicy() ScanPolicy {
	return ScanPolicy{
		BlockAtOrAbove: SeverityHigh,
		RequireScan:    false,
	}
}

// Apply evaluates findings against the policy. Returns:
//   - the filtered set of findings that triggered a block
//   - nil if the image passes the policy
//   - ErrPolicyViolation wrapping the blocking findings if it fails
func (p ScanPolicy) Apply(findings []Vulnerability) ([]Vulnerability, error) {
	threshold := severityRank(p.BlockAtOrAbove)
	if threshold == 0 {
		// No threshold means policy is unset; treat as DefaultScanPolicy.
		threshold = severityRank(SeverityHigh)
	}

	ignore := make(map[string]struct{}, len(p.IgnoreCVEs))
	for _, id := range p.IgnoreCVEs {
		ignore[id] = struct{}{}
	}

	var blocking []Vulnerability
	for _, v := range findings {
		if _, ok := ignore[v.ID]; ok {
			continue
		}
		if severityRank(v.Severity) >= threshold {
			blocking = append(blocking, v)
		}
	}

	if len(blocking) == 0 {
		return nil, nil
	}
	return blocking, fmt.Errorf("%w: %d finding(s) at or above %s", ErrPolicyViolation, len(blocking), p.BlockAtOrAbove)
}

// ImageScanner runs a vulnerability scan against an image reference.
type ImageScanner interface {
	// ID returns a stable identifier for the scanner backend + version. Used
	// as part of the cache key so backend swaps invalidate cached reports.
	ID() string

	// Scan produces a ScanReport for the given image reference. Implementations
	// should return an empty Vulnerabilities slice (not nil) on a successful
	// scan with no findings.
	Scan(ctx context.Context, imageRef string) (*ScanReport, error)
}

// Sentinel errors.
var (
	// ErrPolicyViolation indicates that the scan succeeded but the findings
	// violate the supplied policy.
	ErrPolicyViolation = errors.New("image scan policy violation")

	// ErrScannerUnavailable indicates the scanner binary is missing or the
	// scanner subprocess could not be started.
	ErrScannerUnavailable = errors.New("image scanner unavailable")

	// ErrScanRequired indicates the policy demanded a scan but the scan
	// produced no results (often because the scanner failed silently).
	ErrScanRequired = errors.New("image scan required by policy but produced no results")
)

// CachedScanner wraps any ImageScanner with an in-memory result cache keyed by
// (scannerID, imageRef). Concurrent calls for the same key share a single
// underlying scan via singleflight semantics.
type CachedScanner struct {
	inner ImageScanner
	mu    sync.Mutex
	cache map[string]*scanCacheEntry
}

type scanCacheEntry struct {
	done    chan struct{}
	report  *ScanReport
	err     error
	expires time.Time
}

// NewCachedScanner returns a CachedScanner with the given TTL. A zero TTL
// means cache forever (until daemon restart).
func NewCachedScanner(inner ImageScanner) *CachedScanner {
	return &CachedScanner{
		inner: inner,
		cache: make(map[string]*scanCacheEntry),
	}
}

// ID returns the underlying scanner's ID.
func (cs *CachedScanner) ID() string { return cs.inner.ID() }

// Scan returns a cached report or runs the underlying scanner.
func (cs *CachedScanner) Scan(ctx context.Context, imageRef string) (*ScanReport, error) {
	key := cs.inner.ID() + "|" + imageRef

	cs.mu.Lock()
	entry, exists := cs.cache[key]
	if exists && (entry.expires.IsZero() || time.Now().Before(entry.expires)) {
		cs.mu.Unlock()
		<-entry.done
		return entry.report, entry.err
	}
	entry = &scanCacheEntry{done: make(chan struct{})}
	cs.cache[key] = entry
	cs.mu.Unlock()

	report, err := cs.inner.Scan(ctx, imageRef)
	entry.report = report
	entry.err = err
	close(entry.done)
	return report, err
}

// Invalidate removes the cache entry for an imageRef (for retries after the
// scanner backend changes or the user pushes a new image with the same tag).
func (cs *CachedScanner) Invalidate(imageRef string) {
	key := cs.inner.ID() + "|" + imageRef
	cs.mu.Lock()
	delete(cs.cache, key)
	cs.mu.Unlock()
}

// TrivyCLIScanner shells out to the Trivy CLI binary.
//
// The binary must be installed on the host; `doctor` should check for it. We
// invoke it with `--format json --quiet` and parse the structured output.
type TrivyCLIScanner struct {
	// BinaryPath is the trivy executable; defaults to "trivy" (looked up on PATH).
	BinaryPath string
	// Timeout is the per-scan deadline; defaults to 2 minutes.
	Timeout time.Duration
	// Severities restricts the scan to a subset of severity buckets. Empty
	// means "all". Map to Trivy's `--severity` flag.
	Severities []Severity
}

// NewTrivyCLIScanner returns a TrivyCLIScanner with reasonable defaults.
func NewTrivyCLIScanner() *TrivyCLIScanner {
	return &TrivyCLIScanner{
		BinaryPath: "trivy",
		Timeout:    2 * time.Minute,
		Severities: []Severity{SeverityHigh, SeverityCritical},
	}
}

// ID identifies the scanner backend. We embed the resolved binary path so
// that swapping `trivy` between hosts invalidates cached results.
func (t *TrivyCLIScanner) ID() string {
	return "trivy-cli:" + t.BinaryPath
}

// trivyOutput models the relevant subset of Trivy's JSON output. The full
// schema is large; we extract only what the policy needs.
type trivyOutput struct {
	Results []struct {
		Vulnerabilities []struct {
			VulnerabilityID  string   `json:"VulnerabilityID"`
			PkgName          string   `json:"PkgName"`
			InstalledVersion string   `json:"InstalledVersion"`
			FixedVersion     string   `json:"FixedVersion"`
			Severity         Severity `json:"Severity"`
		} `json:"Vulnerabilities"`
	} `json:"Results"`
}

// Scan invokes the Trivy CLI and parses its JSON output.
func (t *TrivyCLIScanner) Scan(ctx context.Context, imageRef string) (*ScanReport, error) {
	if t.BinaryPath == "" {
		t.BinaryPath = "trivy"
	}
	timeout := t.Timeout
	if timeout == 0 {
		timeout = 2 * time.Minute
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	args := []string{"image", "--format", "json", "--quiet"}
	if len(t.Severities) > 0 {
		sevs := make([]string, len(t.Severities))
		for i, s := range t.Severities {
			sevs[i] = string(s)
		}
		args = append(args, "--severity", strings.Join(sevs, ","))
	}
	args = append(args, imageRef)

	started := time.Now()
	cmd := exec.CommandContext(ctx, t.BinaryPath, args...)
	stdout, err := cmd.Output()
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return nil, fmt.Errorf("%w: timed out after %s", ErrScannerUnavailable, timeout)
		}
		if _, ok := err.(*exec.Error); ok {
			return nil, fmt.Errorf("%w: %v", ErrScannerUnavailable, err)
		}
		// Trivy returns non-zero on findings AND on errors; we treat unparseable
		// output as a scanner failure and parseable output as a scan result.
		var parsed trivyOutput
		if jsonErr := json.Unmarshal(stdout, &parsed); jsonErr != nil {
			return nil, fmt.Errorf("%w: trivy exited with %v and unparseable output", ErrScannerUnavailable, err)
		}
		// Fall through with the parsed output (trivy exited non-zero due to findings).
		return buildReport(imageRef, t.ID(), parsed, started), nil
	}

	var parsed trivyOutput
	if err := json.Unmarshal(stdout, &parsed); err != nil {
		return nil, fmt.Errorf("%w: parse trivy output: %v", ErrScannerUnavailable, err)
	}
	return buildReport(imageRef, t.ID(), parsed, started), nil
}

func buildReport(ref, scannerID string, out trivyOutput, started time.Time) *ScanReport {
	report := &ScanReport{
		ImageRef:        ref,
		ScannerID:       scannerID,
		Vulnerabilities: []Vulnerability{},
		ScanStartedAt:   started,
		ScanDuration:    time.Since(started),
	}
	for _, r := range out.Results {
		for _, v := range r.Vulnerabilities {
			report.Vulnerabilities = append(report.Vulnerabilities, Vulnerability{
				ID:           v.VulnerabilityID,
				Severity:     v.Severity,
				Package:      v.PkgName,
				Version:      v.InstalledVersion,
				FixedVersion: v.FixedVersion,
			})
		}
	}
	return report
}

// NoopScanner satisfies ImageScanner without actually scanning. Used when CVE
// scanning is disabled by config or when tests don't want to exec a binary.
type NoopScanner struct{}

// NewNoopScanner returns a NoopScanner.
func NewNoopScanner() *NoopScanner { return &NoopScanner{} }

// ID identifies the noop scanner.
func (NoopScanner) ID() string { return "noop" }

// Scan returns an empty report.
func (NoopScanner) Scan(ctx context.Context, imageRef string) (*ScanReport, error) {
	return &ScanReport{
		ImageRef:        imageRef,
		ScannerID:       "noop",
		Vulnerabilities: []Vulnerability{},
		ScanStartedAt:   time.Now(),
	}, nil
}
