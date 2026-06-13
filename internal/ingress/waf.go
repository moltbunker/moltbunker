package ingress

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"sort"
	"strconv"
	"strings"

	coreruleset "github.com/corazawaf/coraza-coreruleset/v4"
	coraza "github.com/corazawaf/coraza/v3"
	corazatypes "github.com/corazawaf/coraza/v3/types"

	"github.com/moltbunker/moltbunker/internal/logging"
)

// WAFEngine inspects an inbound HTTP request and reports whether the request
// should be blocked according to the configured rule set. The engine is
// pluggable: the real implementation (CorazaWAF) wraps Coraza + the OWASP CRS,
// and NoopWAFEngine is a zero-cost default used when the WAF is disabled so the
// hot path never needs a nil check.
//
// Implementations MUST be safe for concurrent use: a single engine instance is
// shared across all ingress requests (the ~25MB CRS pattern-matcher state is
// paid once). Per-request state lives in a Coraza transaction created and
// closed inside Inspect.
type WAFEngine interface {
	// Inspect runs the request through the WAF for the given tenant
	// (subdomain). The returned WAFResult reports whether enforcement would
	// block the request and which rules matched. Inspect never mutates r; the
	// caller is responsible for buffering/replacing r.Body for the upstream.
	Inspect(r *http.Request, tenantID string) (*WAFResult, error)

	// Mode reports the effective enforcement mode for a tenant: ModeBlocking
	// means a blocked result should produce a 403; ModeDetection means matches
	// are recorded but the request is allowed through.
	Mode(tenantID string) WAFMode
}

// WAFMode is the enforcement mode of the WAF for a given tenant.
type WAFMode string

const (
	// ModeDetection records rule matches but never blocks (default-safe).
	ModeDetection WAFMode = "detection"
	// ModeBlocking returns 403 when a disruptive rule matches.
	ModeBlocking WAFMode = "blocking"
)

// parseWAFMode normalizes an operator-supplied mode string. Anything other
// than "blocking" falls back to detection (safe default).
func parseWAFMode(s string) WAFMode {
	if strings.EqualFold(strings.TrimSpace(s), string(ModeBlocking)) {
		return ModeBlocking
	}
	return ModeDetection
}

// WAFResult is the outcome of inspecting a single request.
type WAFResult struct {
	// Blocked is true when a disruptive rule matched AND the engine is in
	// blocking mode for this tenant. In detection mode this is always false
	// even when rules matched (see MatchedRules).
	Blocked bool
	// MatchedRules holds the IDs (as strings) of every rule that fired,
	// regardless of mode. Useful for detection-mode logging and metrics.
	MatchedRules []string
	// MatchedPhases maps the rule ID -> the CRS processing phase it fired in.
	MatchedPhases map[string]int
	// Phase is the phase of the disruptive interruption (0 if none).
	Phase int
	// Truncated is true when the request body exceeded BodyLimitBytes and only
	// the first BodyLimitBytes were inspected.
	Truncated bool
}

// WAFConfig is the runtime configuration for CorazaWAF. It mirrors the
// operator-facing config.WAFIngressConfig but lives in this package to avoid
// an import cycle and to allow per-tenant overrides.
type WAFConfig struct {
	// Enabled selects the real engine; when false callers should use
	// NoopWAFEngine instead (NewWAFEngine handles this).
	Enabled bool
	// Mode is the default enforcement mode for all tenants.
	Mode string
	// BodyLimitBytes caps how many request body bytes are read into the WAF.
	BodyLimitBytes int
	// ExcludeRuleIDs lists CRS rule IDs disabled globally (SecRuleRemoveById).
	ExcludeRuleIDs []int
	// TenantOverrides allows per-subdomain mode / rule exclusion overrides.
	TenantOverrides map[string]TenantWAFConfig
}

// TenantWAFConfig overrides WAF behavior for a single tenant (subdomain).
type TenantWAFConfig struct {
	Mode           string
	ExcludeRuleIDs []int
}

const defaultBodyLimitBytes = 65536

// CorazaWAF is the production WAFEngine backed by Coraza v3 and the embedded
// OWASP Core Rule Set. A single instance is created per ingress node and is
// safe for concurrent use.
type CorazaWAF struct {
	waf     coraza.WAF
	mode    WAFMode
	tenants map[string]TenantWAFConfig
	bodyCap int
}

// NewWAFEngine returns a WAFEngine according to cfg. When cfg.Enabled is false
// it returns a NoopWAFEngine (zero cost, no CRS loaded). Otherwise it builds a
// real CorazaWAF with the OWASP CRS embedded — this loads the full rule set and
// costs ~25MB of pattern-matcher state.
func NewWAFEngine(cfg WAFConfig) (WAFEngine, error) {
	if !cfg.Enabled {
		return NoopWAFEngine{}, nil
	}
	return NewCorazaWAF(cfg)
}

// NewCorazaWAF builds the Coraza engine with the OWASP CRS loaded from the
// embedded coreruleset filesystem. No external files or network are required;
// the CRS ships inside the binary.
func NewCorazaWAF(cfg WAFConfig) (*CorazaWAF, error) {
	bodyCap := cfg.BodyLimitBytes
	if bodyCap <= 0 {
		bodyCap = defaultBodyLimitBytes
	}

	directives := buildDirectives(bodyCap, cfg.ExcludeRuleIDs)

	wafCfg := coraza.NewWAFConfig().
		WithRootFS(coreruleset.FS).
		WithRequestBodyAccess().
		WithRequestBodyLimit(bodyCap).
		WithRequestBodyInMemoryLimit(bodyCap).
		WithErrorCallback(func(mr corazatypes.MatchedRule) {
			// Surface WAF matches at debug level only; we never log request
			// body content, only the rule metadata.
			logging.Debug("waf rule matched",
				"rule_id", mr.Rule().ID(),
				"phase", int(mr.Rule().Phase()),
				logging.Component("ingress"))
		}).
		WithDirectives(directives)

	waf, err := coraza.NewWAF(wafCfg)
	if err != nil {
		return nil, fmt.Errorf("initialize coraza waf: %w", err)
	}

	tenants := make(map[string]TenantWAFConfig, len(cfg.TenantOverrides))
	for k, v := range cfg.TenantOverrides {
		tenants[k] = v
	}

	return &CorazaWAF{
		waf:     waf,
		mode:    parseWAFMode(cfg.Mode),
		tenants: tenants,
		bodyCap: bodyCap,
	}, nil
}

// crsSetupVersion is the OWASP CRS setup version variable expected by the rule
// files (corresponds to coraza-coreruleset v4.x). It must be set before the
// rules load or the CRS bootstrap rule (901001) aborts the engine.
const crsSetupVersion = "4140"

// buildDirectives assembles the SecLang directive string.
//
// The engine ALWAYS runs in "On" mode with a blocking SecDefaultAction so that
// Coraza produces a real interruption when the CRS anomaly score crosses the
// inbound threshold. Detection vs blocking is decided by the middleware based
// on the per-tenant Mode (CorazaWAF.Mode) — NOT by SecRuleEngine. This lets us
// always gather full match metadata for metrics/logging while only writing a
// 403 when the tenant is in blocking mode.
//
// We deliberately do NOT Include @crs-setup.conf.example, because that file
// hard-codes SecDefaultAction "...,pass" (detection-only), which would suppress
// the interruption entirely. Instead we supply a minimal CRS setup inline:
// the anomaly thresholds and the crs_setup_version marker the rule files
// require.
func buildDirectives(bodyCap int, excludeRuleIDs []int) string {
	var b strings.Builder
	// Coraza recommended base config (phases, default collections, etc.).
	b.WriteString("Include @coraza.conf-recommended\n")
	b.WriteString("SecRuleEngine On\n")
	b.WriteString("SecRequestBodyAccess On\n")
	b.WriteString("SecRequestBodyLimit " + strconv.Itoa(bodyCap) + "\n")
	b.WriteString("SecRequestBodyInMemoryLimit " + strconv.Itoa(bodyCap) + "\n")
	b.WriteString("SecRequestBodyLimitAction ProcessPartial\n")
	// Audit logging off by default — we never want to write request bodies to
	// disk implicitly (see secret_safety in the blueprint).
	b.WriteString("SecAuditEngine Off\n")
	// Blocking default action so rule 949110 (inbound anomaly evaluation) can
	// deny. The middleware ignores the resulting interruption in detection
	// mode.
	b.WriteString("SecDefaultAction \"phase:1,log,auditlog,deny,status:403\"\n")
	b.WriteString("SecDefaultAction \"phase:2,log,auditlog,deny,status:403\"\n")
	// Minimal CRS setup: anomaly thresholds + setup version marker. These would
	// normally live in crs-setup.conf; we inline them to keep a blocking
	// default action (the example setup forces pass).
	b.WriteString("SecAction \"id:900110,phase:1,nolog,pass,t:none," +
		"setvar:tx.inbound_anomaly_score_threshold=5," +
		"setvar:tx.outbound_anomaly_score_threshold=4\"\n")
	b.WriteString("SecAction \"id:900990,phase:1,nolog,pass,t:none," +
		"setvar:tx.crs_setup_version=" + crsSetupVersion + "\"\n")
	// Load the OWASP CRS rule files from the embedded FS.
	b.WriteString("Include @owasp_crs/*.conf\n")

	if len(excludeRuleIDs) > 0 {
		ids := dedupSortInts(excludeRuleIDs)
		parts := make([]string, len(ids))
		for i, id := range ids {
			parts[i] = strconv.Itoa(id)
		}
		b.WriteString("SecRuleRemoveById " + strings.Join(parts, " ") + "\n")
	}
	return b.String()
}

func dedupSortInts(in []int) []int {
	seen := make(map[int]struct{}, len(in))
	out := make([]int, 0, len(in))
	for _, v := range in {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	sort.Ints(out)
	return out
}

// Mode returns the effective enforcement mode for a tenant, honoring any
// per-tenant override.
func (c *CorazaWAF) Mode(tenantID string) WAFMode {
	if ov, ok := c.tenants[tenantID]; ok && ov.Mode != "" {
		return parseWAFMode(ov.Mode)
	}
	return c.mode
}

// Inspect runs the request through Coraza. It feeds the connection metadata,
// URI, headers, and a size-capped copy of the body into a fresh transaction,
// processes phases 1 and 2, and reports the result. The transaction is always
// closed before returning.
//
// Inspect does not consume r.Body destructively: callers that have already
// buffered the body should pass a request whose Body is a re-readable reader.
// To keep the engine self-contained, Inspect reads up to bodyCap bytes from
// r.Body and then leaves r.Body positioned at whatever remains; the middleware
// is responsible for restoring r.Body for the upstream.
func (c *CorazaWAF) Inspect(r *http.Request, tenantID string) (*WAFResult, error) {
	tx := c.waf.NewTransaction()
	// Coraza requires the transaction to be closed to release buffers.
	// The ProcessLogging + Close pair is the documented teardown.
	defer func() {
		tx.ProcessLogging()
		// #nosec G104 -- Close only releases per-transaction buffers; an error
		// here cannot affect the already-computed verdict and there is no
		// meaningful recovery action in a deferred teardown.
		_ = tx.Close()
	}()

	result := &WAFResult{MatchedPhases: make(map[string]int)}

	// Phase 1: connection + URI + headers.
	clientIP, clientPort := splitHostPortBestEffort(r.RemoteAddr)
	tx.ProcessConnection(clientIP, clientPort, "", 0)

	uri := r.URL.RequestURI()
	tx.ProcessURI(uri, r.Method, protoString(r))

	if host := r.Host; host != "" {
		tx.SetServerName(host)
		tx.AddRequestHeader("Host", host)
	}
	for name, values := range r.Header {
		for _, v := range values {
			tx.AddRequestHeader(name, v)
		}
	}

	// Process phase 1. We do NOT return early on an interruption here: we still
	// feed and process the body so request-body inspection and truncation
	// accounting always run. The final verdict is read from the transaction's
	// interruption state after all phases, which is the canonical Coraza usage.
	_ = tx.ProcessRequestHeaders()

	// Phase 2: request body (size-capped).
	if r.Body != nil {
		truncated, err := c.feedBody(tx, r.Body)
		if err != nil {
			c.collectMatches(result, tx)
			return result, fmt.Errorf("waf body inspection: %w", err)
		}
		result.Truncated = truncated
	}

	if _, err := tx.ProcessRequestBody(); err != nil {
		c.collectMatches(result, tx)
		return result, fmt.Errorf("waf process request body: %w", err)
	}

	if it := tx.Interruption(); it != nil {
		c.applyInterruption(result, it, tx, tenantID)
	}

	c.collectMatches(result, tx)
	return result, nil
}

// feedBody streams up to bodyCap bytes of the request body into the
// transaction. It returns truncated=true when the body exceeded bodyCap. The
// remaining (unread) bytes are left on r.Body untouched — the middleware
// already holds the full buffered body for the upstream.
func (c *CorazaWAF) feedBody(tx corazatypes.Transaction, body io.Reader) (bool, error) {
	limited := io.LimitReader(body, int64(c.bodyCap))
	buf := make([]byte, 0, c.bodyCap)
	tmp := bytes.NewBuffer(buf)
	n, err := io.Copy(tmp, limited)
	if err != nil {
		return false, err
	}
	if n > 0 {
		if _, _, werr := tx.WriteRequestBody(tmp.Bytes()); werr != nil {
			return false, werr
		}
	}
	// Detect truncation: if there is at least one more byte beyond the cap.
	one := make([]byte, 1)
	m, _ := body.Read(one)
	return m > 0, nil
}

// applyInterruption records a disruptive interruption. Blocked is only set in
// blocking mode; in detection mode the match is recorded but the request is
// allowed through. The interruption's rule phase is preferred; we fall back to
// the highest matched-rule phase.
func (c *CorazaWAF) applyInterruption(result *WAFResult, it *corazatypes.Interruption, tx corazatypes.Transaction, tenantID string) {
	result.Phase = interruptionPhase(it, tx)
	if c.Mode(tenantID) == ModeBlocking {
		result.Blocked = true
	}
}

// interruptionPhase resolves the phase that produced the interruption by
// matching its RuleID against the matched rules; falls back to the highest
// matched phase.
func interruptionPhase(it *corazatypes.Interruption, tx corazatypes.Transaction) int {
	if it != nil {
		for _, mr := range tx.MatchedRules() {
			if mr.Rule().ID() == it.RuleID {
				return int(mr.Rule().Phase())
			}
		}
	}
	return highestMatchedPhase(tx)
}

// collectMatches records the matched rules that represent real attack
// detections, independent of enforcement mode. It deliberately skips the CRS
// housekeeping / initialization SecActions (the 900000-901999 setup/bootstrap
// block and the per-category NNN013/NNN014 "initialization" markers) so that
// the detection signal, block decisions, and metrics reflect actual attacks
// rather than baseline CRS bootstrap noise. Without this filter a completely
// clean request reports ~60 "matched" rules — none of which are detections —
// which makes matched=len(MatchedRules)>0 always true and blows up the
// per-rule metric cardinality. See isDetectionMatch.
func (c *CorazaWAF) collectMatches(result *WAFResult, tx corazatypes.Transaction) {
	for _, mr := range tx.MatchedRules() {
		if !isDetectionMatch(mr) {
			continue
		}
		id := strconv.Itoa(mr.Rule().ID())
		result.MatchedRules = append(result.MatchedRules, id)
		result.MatchedPhases[id] = int(mr.Rule().Phase())
	}
}

// isDetectionMatch reports whether a matched rule is a real attack/anomaly
// detection rather than a CRS configuration, setup, or per-phase
// initialization marker.
//
// Two gates, both required:
//  1. The rule ID must be outside the CRS config/init range 900000-901999
//     (crs-setup, the 901xxx bootstrap block, and our inline setup SecActions
//     900110/900990 all live here and carry pass actions, e.g. 901340
//     "Enabling body inspection").
//  2. The rule must carry a non-empty Message(). CRS setup/init SecActions and
//     the per-category NNN013/NNN014 "initialization" markers emit no message;
//     only detection rules (e.g. 942100 "SQL Injection Attack Detected",
//     930100 "Path Traversal Attack") and the anomaly-scoring evaluators do.
//
// Verified empirically: a clean request yields zero detection matches under
// this filter, while a SQLi or path-traversal probe yields the expected
// detection rule(s).
func isDetectionMatch(mr corazatypes.MatchedRule) bool {
	id := mr.Rule().ID()
	if id >= 900000 && id <= 901999 {
		return false
	}
	return mr.Message() != ""
}

func highestMatchedPhase(tx corazatypes.Transaction) int {
	phase := 0
	for _, mr := range tx.MatchedRules() {
		if p := int(mr.Rule().Phase()); p > phase {
			phase = p
		}
	}
	return phase
}

func splitHostPortBestEffort(addr string) (string, int) {
	if addr == "" {
		return "", 0
	}
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return addr, 0
	}
	port, _ := strconv.Atoi(portStr)
	return host, port
}

func protoString(r *http.Request) string {
	if r.Proto != "" {
		// Coraza expects "1.1" / "2.0" style, not "HTTP/1.1".
		return strings.TrimPrefix(r.Proto, "HTTP/")
	}
	return "1.1"
}
