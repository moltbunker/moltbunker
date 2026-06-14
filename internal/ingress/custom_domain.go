package ingress

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

// custom_domain.go implements BYO ("bring your own") custom-hostname ownership
// verification for the edge layer (EDGE-02).
//
// A customer who wants app.customer.com served by moltbunker proves they
// control the DNS for that host BEFORE the autocert hostPolicy will let Let's
// Encrypt issue a certificate for it. Verification is a second ownership-proof
// layer ON TOP of ACME (not a replacement for the ACME challenge): the customer
// publishes a CNAME or TXT record carrying a per-host token derived from an
// in-memory HMAC secret, the daemon confirms it via stdlib DNS, then persists
// the (host -> deployment) mapping so hostPolicy and the proxy will accept it.
//
// This mirrors how Cloudflare Pages (CNAME) and Vercel (CNAME/TXT) verify
// custom domains. DNS-01 ACME delegation (for wildcard custom certs) is a
// documented follow-up; here we issue an individual LE cert per verified host.

const (
	// verifyTokenLabelPrefix is the DNS label the TXT verifier looks under:
	// _moltbunker.<host>.
	verifyTXTPrefix = "_moltbunker."
	// txtRecordPrefix is the value prefix in the published TXT record.
	txtRecordPrefix = "moltbunker-verify="
	// defaultOwnershipTTL is the verification validity window when the config
	// does not set one.
	defaultOwnershipTTL = 72 * time.Hour
)

// VerifyMethod identifies which DNS proof a verifier expects.
type VerifyMethod string

const (
	// MethodCNAME expects a CNAME from <host> to <token>.<domain>.
	MethodCNAME VerifyMethod = "cname"
	// MethodTXT expects a TXT record moltbunker-verify=<token> on
	// _moltbunker.<host>.
	MethodTXT VerifyMethod = "txt"
)

// VerifiedDomain is a persisted record that a custom host was proven to belong
// to a deployment. All fields are public routing metadata — no secrets.
type VerifiedDomain struct {
	Host         string       `json:"host"`
	DeploymentID string       `json:"deployment_id"`
	OwnerWallet  string       `json:"owner_wallet,omitempty"`
	Method       VerifyMethod `json:"method"`
	VerifiedAt   time.Time    `json:"verified_at"`
	ExpiresAt    time.Time    `json:"expires_at"`
}

// expired reports whether the record's TTL has elapsed. A zero ExpiresAt never
// expires (TTL disabled).
func (v VerifiedDomain) expired(now time.Time) bool {
	return !v.ExpiresAt.IsZero() && now.After(v.ExpiresAt)
}

// DomainOwnershipStore is a thread-safe map of verified custom hosts. In-memory
// for this PR (persistence across daemon restart is a documented follow-up);
// the autocert DirCache still caches issued certs on disk, so a restart only
// requires re-verifying ownership, not re-issuing certs.
type DomainOwnershipStore struct {
	mu      sync.RWMutex
	domains map[string]VerifiedDomain // lowercased host -> record
	ttl     time.Duration
}

// NewDomainOwnershipStore creates an empty store. A non-positive ttl uses the
// 72h default; pass a negative sentinel via WithNoTTL semantics is not needed —
// callers that want no expiry should set OwnershipTTLHours appropriately.
func NewDomainOwnershipStore(ttl time.Duration) *DomainOwnershipStore {
	if ttl <= 0 {
		ttl = defaultOwnershipTTL
	}
	return &DomainOwnershipStore{
		domains: make(map[string]VerifiedDomain),
		ttl:     ttl,
	}
}

// normalizeHost lowercases and strips a trailing dot / port so lookups are
// stable regardless of how the Host header arrives.
func normalizeHost(host string) string {
	host = strings.ToLower(strings.TrimSpace(host))
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	return strings.TrimSuffix(host, ".")
}

// Store records a verified custom host. VerifiedAt is stamped now and ExpiresAt
// is set from the store TTL.
func (s *DomainOwnershipStore) Store(host, deploymentID, ownerWallet string, method VerifyMethod) {
	host = normalizeHost(host)
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.domains[host] = VerifiedDomain{
		Host:         host,
		DeploymentID: deploymentID,
		OwnerWallet:  ownerWallet,
		Method:       method,
		VerifiedAt:   now,
		ExpiresAt:    now.Add(s.ttl),
	}
}

// StoreIfUnderCap atomically enforces the per-deployment domain cap and stores
// the mapping under a single write lock, removing the check-then-store TOCTOU
// that a separate CountForDeployment + Store would leave open under concurrent
// verifies for the same deployment.
//
// maxPerDep <= 0 means unlimited (always stores). Re-verifying a host that is
// already mapped to the SAME deployment never counts against the cap (it is an
// in-place refresh). It returns the stored record and ok=false (without storing)
// when adding a NEW host would exceed the cap.
func (s *DomainOwnershipStore) StoreIfUnderCap(host, deploymentID, ownerWallet string, method VerifyMethod, maxPerDep int) (VerifiedDomain, bool) {
	host = normalizeHost(host)
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()

	if maxPerDep > 0 {
		// A re-verify of an existing, non-expired mapping to the same deployment
		// is an in-place refresh and must not be blocked by the cap.
		existing, present := s.domains[host]
		isRefresh := present && !existing.expired(now) && existing.DeploymentID == deploymentID
		if !isRefresh {
			n := 0
			for _, rec := range s.domains {
				if rec.DeploymentID == deploymentID && !rec.expired(now) {
					n++
				}
			}
			if n >= maxPerDep {
				return VerifiedDomain{}, false
			}
		}
	}

	rec := VerifiedDomain{
		Host:         host,
		DeploymentID: deploymentID,
		OwnerWallet:  ownerWallet,
		Method:       method,
		VerifiedAt:   now,
		ExpiresAt:    now.Add(s.ttl),
	}
	s.domains[host] = rec
	return rec, true
}

// LookupByHost returns the verified record for host if present and not expired.
func (s *DomainOwnershipStore) LookupByHost(host string) (VerifiedDomain, bool) {
	host = normalizeHost(host)
	s.mu.RLock()
	rec, ok := s.domains[host]
	s.mu.RUnlock()
	if !ok || rec.expired(time.Now()) {
		return VerifiedDomain{}, false
	}
	return rec, true
}

// Remove deletes a custom-host record (operator revocation / customer offboard).
func (s *DomainOwnershipStore) Remove(host string) {
	host = normalizeHost(host)
	s.mu.Lock()
	delete(s.domains, host)
	s.mu.Unlock()
}

// Count returns the number of (possibly expired) stored records.
func (s *DomainOwnershipStore) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.domains)
}

// CountForDeployment returns the number of non-expired hosts mapped to a
// deployment (used to enforce MaxDomainsPerDeployment).
func (s *DomainOwnershipStore) CountForDeployment(deploymentID string) int {
	now := time.Now()
	s.mu.RLock()
	defer s.mu.RUnlock()
	n := 0
	for _, rec := range s.domains {
		if rec.DeploymentID == deploymentID && !rec.expired(now) {
			n++
		}
	}
	return n
}

// GenerateVerificationToken derives a deterministic per-(host,deployment) proof
// token. Binding the token to an HMAC secret prevents an attacker from
// pre-computing a valid token for a host they do not control: only a party that
// can publish the DNS record for the host (which they prove by doing so) AND
// that received the token from this daemon can pass verification. The token is
// hex(HMAC-SHA256(secret, host + "|" + deploymentID)).
func GenerateVerificationToken(secret []byte, host, deploymentID string) string {
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(normalizeHost(host)))
	mac.Write([]byte("|"))
	mac.Write([]byte(deploymentID))
	return hex.EncodeToString(mac.Sum(nil))
}

// resolverFunc is the minimal DNS surface the verifiers need. The stdlib
// *net.Resolver satisfies it; tests inject a fake without a DNS library.
type resolverFunc interface {
	LookupCNAME(ctx context.Context, host string) (string, error)
	LookupTXT(ctx context.Context, host string) ([]string, error)
}

// DomainVerifier proves DNS ownership of a custom host for a deployment.
type DomainVerifier interface {
	// Verify checks the DNS proof for host bound to deploymentID. It returns nil
	// when ownership is proven, or an error describing the missing/incorrect
	// record. Method reports which proof this verifier checks.
	Verify(ctx context.Context, host, deploymentID string) error
	Method() VerifyMethod
}

// baseVerifier holds the shared resolver + secret + expected CNAME target
// domain (e.g. "moltbunker.dev").
type baseVerifier struct {
	resolver  resolverFunc
	secret    []byte
	verifyDom string
}

// CNAMEVerifier proves ownership by checking that <host> CNAMEs to
// <token>.<verifyDom>.
type CNAMEVerifier struct{ baseVerifier }

// Method implements DomainVerifier.
func (CNAMEVerifier) Method() VerifyMethod { return MethodCNAME }

// Verify implements DomainVerifier for the CNAME proof.
func (v CNAMEVerifier) Verify(ctx context.Context, host, deploymentID string) error {
	host = normalizeHost(host)
	token := GenerateVerificationToken(v.secret, host, deploymentID)
	want := normalizeHost(token + "." + v.verifyDom)

	cname, err := v.resolver.LookupCNAME(ctx, host)
	if err != nil {
		return fmt.Errorf("CNAME lookup for %q failed: %w", host, err)
	}
	if normalizeHost(cname) != want {
		return fmt.Errorf("CNAME for %q is %q, expected %q", host, normalizeHost(cname), want)
	}
	return nil
}

// TXTVerifier proves ownership by checking for a TXT record
// moltbunker-verify=<token> on _moltbunker.<host>.
type TXTVerifier struct{ baseVerifier }

// Method implements DomainVerifier.
func (TXTVerifier) Method() VerifyMethod { return MethodTXT }

// Verify implements DomainVerifier for the TXT proof.
func (v TXTVerifier) Verify(ctx context.Context, host, deploymentID string) error {
	host = normalizeHost(host)
	token := GenerateVerificationToken(v.secret, host, deploymentID)
	want := txtRecordPrefix + token

	lookupName := verifyTXTPrefix + host
	records, err := v.resolver.LookupTXT(ctx, lookupName)
	if err != nil {
		return fmt.Errorf("TXT lookup for %q failed: %w", lookupName, err)
	}
	for _, rec := range records {
		if strings.TrimSpace(rec) == want {
			return nil
		}
	}
	return fmt.Errorf("no matching TXT record on %q (expected %q)", lookupName, want)
}

// NewDomainVerifier builds the verifier for the given method. An unknown method
// defaults to CNAME. verifyDom is the base domain custom hosts CNAME to (e.g.
// "moltbunker.dev"). resolver may be nil to use the stdlib default resolver.
func NewDomainVerifier(method VerifyMethod, verifyDom string, secret []byte, resolver resolverFunc) DomainVerifier {
	if resolver == nil {
		resolver = net.DefaultResolver
	}
	base := baseVerifier{resolver: resolver, secret: secret, verifyDom: verifyDom}
	if method == MethodTXT {
		return TXTVerifier{base}
	}
	return CNAMEVerifier{base}
}

// CNAMETarget returns the value a customer must set their CNAME record to for
// the given host+deployment (shown by the challenge API).
func CNAMETarget(secret []byte, verifyDom, host, deploymentID string) string {
	return GenerateVerificationToken(secret, host, deploymentID) + "." + verifyDom
}

// TXTRecord returns the (name, value) pair a customer must publish for the TXT
// proof (shown by the challenge API).
func TXTRecord(secret []byte, host, deploymentID string) (name, value string) {
	host = normalizeHost(host)
	return verifyTXTPrefix + host, txtRecordPrefix + GenerateVerificationToken(secret, host, deploymentID)
}
