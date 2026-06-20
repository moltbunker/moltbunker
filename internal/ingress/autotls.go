package ingress

import (
	"context"
	"crypto/tls"
	"fmt"
	"strings"

	"golang.org/x/crypto/acme/autocert"
)

// AutoTLSConfig manages automatic TLS certificates via Let's Encrypt (ACME).
// It uses TLS-ALPN-01 challenge on port 443 — no port 80 needed.
// The hostPolicy callback ensures certificates are only issued for subdomains
// that resolve to active services, preventing rate limit abuse.
type AutoTLSConfig struct {
	Manager  *autocert.Manager
	domain   string
	resolver *Resolver

	// customDomains, when set, lets hostPolicy accept BYO custom hostnames that
	// have proven DNS ownership (EDGE-02). Nil = the original behavior (only
	// *.<domain> hosts are issued certs). Wired post-construction via
	// SetCustomDomains so main.go can build the store after the manager.
	customDomains *DomainOwnershipStore
}

// NewAutoTLSConfig creates an auto-TLS configuration with Let's Encrypt.
// certDir is the directory for cached certificates, domain is the base domain
// (e.g., "moltbunker.dev"), and acmeEmail is the registration email.
func NewAutoTLSConfig(certDir, domain, acmeEmail string, resolver *Resolver) *AutoTLSConfig {
	a := &AutoTLSConfig{
		domain:   domain,
		resolver: resolver,
	}

	a.Manager = &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		Cache:      autocert.DirCache(certDir),
		HostPolicy: a.hostPolicy,
		Email:      acmeEmail,
	}

	return a
}

// TLSConfig returns a *tls.Config that uses the autocert manager.
// The returned config handles TLS-ALPN-01 challenges automatically.
func (a *AutoTLSConfig) TLSConfig() *tls.Config {
	tlsCfg := a.Manager.TLSConfig()
	tlsCfg.MinVersion = tls.VersionTLS12
	return tlsCfg
}

// SetCustomDomains wires the verified-custom-domain store so hostPolicy will
// also issue certs for BYO hostnames that proved ownership. Safe to call once
// post-construction (EDGE-02). A nil store leaves the original policy intact.
func (a *AutoTLSConfig) SetCustomDomains(store *DomainOwnershipStore) {
	a.customDomains = store
}

// hostPolicy decides whether to issue a certificate for the given host.
// It rejects bare domains, wrong domain suffixes, and subdomains that
// don't resolve to any active service. A verified BYO custom hostname
// (EDGE-02) is accepted via a secondary lookup once the *.<domain> fast path
// has been ruled out.
func (a *AutoTLSConfig) hostPolicy(ctx context.Context, host string) error {
	// Must be under our domain — fast path, unchanged for *.<domain> hosts.
	suffix := "." + a.domain
	if !strings.HasSuffix(host, suffix) {
		// Secondary path: a BYO custom hostname that proved DNS ownership.
		// Only consulted when a custom-domain store is wired AND the host is
		// NOT under our base domain, so existing behavior is preserved both for
		// *.<domain> hosts and when the feature is disabled.
		if a.customDomains != nil {
			if _, ok := a.customDomains.LookupByHost(host); ok {
				return nil
			}
			return fmt.Errorf("custom domain not verified: %s", host)
		}
		return fmt.Errorf("host %q is not under %s", host, a.domain)
	}

	// Extract subdomain (strip ".<domain>" suffix)
	subdomain := strings.TrimSuffix(host, suffix)
	if subdomain == "" {
		return fmt.Errorf("bare domain %q not allowed", host)
	}

	// Only issue certs for subdomains that actually resolve to a service
	if a.resolver != nil {
		if _, err := a.resolver.Resolve(subdomain); err != nil {
			return fmt.Errorf("subdomain %q does not resolve: %w", subdomain, err)
		}
	}

	return nil
}
