// Package networking — R13: per-container network policy.
//
// This file defines a per-container default-deny lateral policy model:
// containers on the same host cannot reach each other unless their tenant
// explicitly opts in via AllowedPeers. The policy is platform-agnostic; the
// Linux enforcer (policy_linux.go) translates it into nftables rules.
//
// Threat model addressed:
//
//   Today: every container's veth lands on the same 10.88.0.0/16 bridge with
//   no inter-container filtering. A compromised container can scan and reach
//   every other container on the same provider — lateral movement across
//   tenants is trivial.
//
//   With this policy: each container's netns gets a default-deny FORWARD rule
//   for traffic destined to other 10.88.0.0/16 addresses. AllowedPeers names
//   exceptions (same-tenant multi-container apps).
//
// This file is platform-agnostic. The actual nftables rules live in
// policy_linux.go (and policy_other.go provides a no-op stub elsewhere).
package networking

import (
	"errors"
	"fmt"
	"net"
	"sync"
)

// EgressMode controls outbound traffic from a container to addresses OUTSIDE
// the 10.88.0.0/16 intra-host network — i.e., the public internet and the
// host's own services.
type EgressMode int

const (
	// EgressDefaultAllow lets the container reach any external address. This
	// matches today's behavior. Pair with EgressDeny CIDRs to carve out
	// metadata-service / RFC1918 destinations.
	EgressDefaultAllow EgressMode = iota

	// EgressDefaultDeny blocks all egress unless the destination matches
	// EgressAllow. This is R14's eventual default.
	EgressDefaultDeny
)

// NetworkPolicy describes which other endpoints a container is permitted to
// talk to. The struct is value-typed and safe to copy.
type NetworkPolicy struct {
	// AllowedPeers lists deployment IDs of OTHER containers on the same host
	// that this container is permitted to reach. Empty means full lateral
	// isolation (default-deny intra-host).
	AllowedPeers []string

	// EgressMode controls default behavior for outbound traffic to the
	// outside world.
	EgressMode EgressMode

	// EgressAllow is a list of CIDRs that are allowed regardless of
	// EgressMode. Use for known-good destinations (e.g. an upstream API the
	// container needs).
	EgressAllow []string

	// EgressDeny is a list of CIDRs that are blocked regardless of
	// EgressMode. Use to carve out dangerous destinations even under
	// EgressDefaultAllow (e.g. 169.254.169.254 cloud metadata, 10.0.0.0/8
	// RFC1918 private networks).
	EgressDeny []string
}

// DefaultNetworkPolicy returns the recommended baseline: full lateral isolation
// (no AllowedPeers) and EgressDefaultAllow with no carved exceptions. Callers
// upgrading toward R14 should switch EgressMode to EgressDefaultDeny.
func DefaultNetworkPolicy() NetworkPolicy {
	return NetworkPolicy{
		AllowedPeers: nil,
		EgressMode:   EgressDefaultAllow,
		EgressAllow:  nil,
		EgressDeny:   nil,
	}
}

// AllowsPeer reports whether the policy permits traffic to another
// deployment's container on the same host.
func (p NetworkPolicy) AllowsPeer(deploymentID string) bool {
	for _, id := range p.AllowedPeers {
		if id == deploymentID {
			return true
		}
	}
	return false
}

// PolicyEnforcer applies and removes NetworkPolicy rules at the OS level.
// Each implementation is platform-specific.
type PolicyEnforcer interface {
	// Apply installs policy rules for the given deployment's container.
	// containerIP is the IP allocated in the 10.88.0.0/16 intra-host network.
	Apply(deploymentID, containerIP string, policy NetworkPolicy) error

	// Remove tears down any rules previously installed for deploymentID.
	// Safe to call even if Apply was never called.
	Remove(deploymentID string) error
}

// PolicyStore tracks active policies in memory. It is the source of truth for
// "what policy does container X currently have?" — useful for AllowedPeers
// reciprocity checks (A allows B but B does not allow A: traffic from A to B
// should still be blocked by B's ingress rules).
type PolicyStore struct {
	mu       sync.RWMutex
	policies map[string]NetworkPolicy // deploymentID → policy
	ips      map[string]string        // deploymentID → containerIP
}

// NewPolicyStore returns an empty PolicyStore.
func NewPolicyStore() *PolicyStore {
	return &PolicyStore{
		policies: make(map[string]NetworkPolicy),
		ips:      make(map[string]string),
	}
}

// Set registers the policy + ip for a deployment.
func (s *PolicyStore) Set(deploymentID, containerIP string, p NetworkPolicy) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.policies[deploymentID] = p
	s.ips[deploymentID] = containerIP
}

// Remove drops the policy for a deployment.
func (s *PolicyStore) Remove(deploymentID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.policies, deploymentID)
	delete(s.ips, deploymentID)
}

// Get returns the policy for a deployment, false if not registered.
func (s *PolicyStore) Get(deploymentID string) (NetworkPolicy, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.policies[deploymentID]
	return p, ok
}

// PeerIP returns the in-network IP of a peer deployment, false if unknown.
func (s *PolicyStore) PeerIP(deploymentID string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ip, ok := s.ips[deploymentID]
	return ip, ok
}

// ResolveAllowedPeerIPs returns the IPs of every peer in policy.AllowedPeers
// whose policy reciprocally lists deploymentID. This is the SAFE set of
// peers — both ends have to opt in for traffic to flow.
//
// If "lax" is true, returns peers based on policy.AllowedPeers regardless of
// reciprocity (one-way allowlist).
func (s *PolicyStore) ResolveAllowedPeerIPs(deploymentID string, policy NetworkPolicy, lax bool) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var ips []string
	for _, peerID := range policy.AllowedPeers {
		ip, ok := s.ips[peerID]
		if !ok {
			continue
		}
		if lax {
			ips = append(ips, ip)
			continue
		}
		// Strict mode: peer must also list us in their AllowedPeers.
		peerPolicy, ok := s.policies[peerID]
		if !ok {
			continue
		}
		if peerPolicy.AllowsPeer(deploymentID) {
			ips = append(ips, ip)
		}
	}
	return ips
}

// Validate checks a NetworkPolicy for malformed CIDRs and self-references.
// Returns the first error encountered.
func (p NetworkPolicy) Validate(selfDeploymentID string) error {
	for _, peer := range p.AllowedPeers {
		if peer == selfDeploymentID {
			return fmt.Errorf("%w: peer list includes self (%s)", ErrInvalidPolicy, selfDeploymentID)
		}
		if peer == "" {
			return fmt.Errorf("%w: empty peer id", ErrInvalidPolicy)
		}
	}
	for _, c := range p.EgressAllow {
		if c == "" {
			return fmt.Errorf("%w: empty egress allow CIDR", ErrInvalidPolicy)
		}
		if _, _, err := net.ParseCIDR(c); err != nil {
			return fmt.Errorf("%w: malformed egress allow CIDR %q: %v", ErrInvalidPolicy, c, err)
		}
	}
	for _, c := range p.EgressDeny {
		if c == "" {
			return fmt.Errorf("%w: empty egress deny CIDR", ErrInvalidPolicy)
		}
		if _, _, err := net.ParseCIDR(c); err != nil {
			return fmt.Errorf("%w: malformed egress deny CIDR %q: %v", ErrInvalidPolicy, c, err)
		}
	}
	return nil
}

// ErrInvalidPolicy is returned by NetworkPolicy.Validate when the policy is
// malformed.
var ErrInvalidPolicy = errors.New("invalid network policy")

// R14 — egress evaluation.
//
// EvaluateEgress applies a policy to a single outbound destination IP and
// returns whether the destination is permitted. This is the pure-function
// counterpart of what the nftables rule set does in-kernel — keeping it as
// Go logic means we can unit-test the precedence semantics in isolation, and
// it lets the daemon perform pre-deploy "would this destination be reachable?"
// checks without actually wiring nft rules.

// EgressDecision is the outcome of evaluating an outbound destination against
// a NetworkPolicy's egress rules.
type EgressDecision int

const (
	// EgressAllowed means the destination is permitted.
	EgressAllowed EgressDecision = iota
	// EgressBlocked means the destination is denied.
	EgressBlocked
)

// String returns "ALLOW" or "BLOCK" for log lines.
func (d EgressDecision) String() string {
	if d == EgressAllowed {
		return "ALLOW"
	}
	return "BLOCK"
}

// EvaluateEgress applies the policy to a single outbound destination IP. The
// precedence rules are:
//
//  1. Explicit deny — if any EgressDeny CIDR contains ip, return EgressBlocked.
//  2. Explicit allow — if any EgressAllow CIDR contains ip, return EgressAllowed.
//  3. Default — fall back to EgressMode.
//
// "Deny beats allow, explicit beats default." A nil or unparseable ip returns
// EgressBlocked (fail-closed for bad input).
//
// The CIDRs in the policy must have been validated via Validate() first;
// EvaluateEgress assumes well-formed CIDRs and silently skips malformed ones.
func (p NetworkPolicy) EvaluateEgress(ip net.IP) EgressDecision {
	if ip == nil {
		return EgressBlocked
	}

	for _, c := range p.EgressDeny {
		if _, network, err := net.ParseCIDR(c); err == nil && network.Contains(ip) {
			return EgressBlocked
		}
	}
	for _, c := range p.EgressAllow {
		if _, network, err := net.ParseCIDR(c); err == nil && network.Contains(ip) {
			return EgressAllowed
		}
	}
	if p.EgressMode == EgressDefaultDeny {
		return EgressBlocked
	}
	return EgressAllowed
}

// EvaluateEgressString is a convenience wrapper that parses ipStr (an IPv4 or
// IPv6 address, not a CIDR) before delegating to EvaluateEgress. Returns
// EgressBlocked if ipStr is unparseable.
func (p NetworkPolicy) EvaluateEgressString(ipStr string) EgressDecision {
	return p.EvaluateEgress(net.ParseIP(ipStr))
}

// ComputeEgressRules translates a NetworkPolicy into the nftables rule lines
// that would be installed for one container. Returned in the order they
// should be applied — earlier rules match first, so explicit deny entries
// come BEFORE allow entries, which come before the default fallback.
//
// The output is the exact text intended for `nft -f -`; tests can assert on
// it without running real nft. When the enforcer is wired for production
// (deferred until R11 — Linux runtime CI is in place), it pipes these lines
// to the nft binary.
//
// Empty deploymentID or containerIP returns nil.
func ComputeEgressRules(deploymentID, containerIP string, policy NetworkPolicy) []string {
	if deploymentID == "" || containerIP == "" {
		return nil
	}

	chain := "mb_" + deploymentID + "_out"
	var rules []string

	// 1. Explicit deny CIDRs (highest precedence)
	for _, c := range policy.EgressDeny {
		rules = append(rules, fmt.Sprintf(
			"add rule inet moltbunker_policy %s ip saddr %s ip daddr %s drop",
			chain, containerIP, c,
		))
	}

	// 2. Explicit allow CIDRs
	for _, c := range policy.EgressAllow {
		rules = append(rules, fmt.Sprintf(
			"add rule inet moltbunker_policy %s ip saddr %s ip daddr %s accept",
			chain, containerIP, c,
		))
	}

	// 3. Default fallback
	switch policy.EgressMode {
	case EgressDefaultDeny:
		rules = append(rules, fmt.Sprintf(
			"add rule inet moltbunker_policy %s ip saddr %s drop",
			chain, containerIP,
		))
	case EgressDefaultAllow:
		rules = append(rules, fmt.Sprintf(
			"add rule inet moltbunker_policy %s ip saddr %s accept",
			chain, containerIP,
		))
	}

	return rules
}

// DefaultRestrictiveEgressPolicy returns a baseline EgressDefaultDeny policy
// with carve-outs for common safe destinations (loopback, DNS via Cloudflare
// 1.1.1.1 and Google 8.8.8.8). Tenants who want full lockdown should start
// here and remove allow entries; tenants who want broad access should NOT
// start here.
//
// This is intentionally a starting-point, not a one-size-fits-all default —
// the actual default in DefaultNetworkPolicy() remains EgressDefaultAllow
// for backwards compatibility.
func DefaultRestrictiveEgressPolicy() NetworkPolicy {
	return NetworkPolicy{
		EgressMode: EgressDefaultDeny,
		EgressAllow: []string{
			"1.1.1.1/32", // Cloudflare DNS
			"8.8.8.8/32", // Google DNS
		},
		// Block cloud-instance metadata + RFC1918 by default to prevent
		// SSRF / lateral-from-cloud-host pivots even if a user later adds a
		// broad allow.
		EgressDeny: []string{
			"169.254.169.254/32", // AWS / GCP / Azure IMDS
			"169.254.0.0/16",     // link-local
			"10.0.0.0/8",         // RFC1918
			"172.16.0.0/12",      // RFC1918
			"192.168.0.0/16",     // RFC1918
		},
	}
}
