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

// hostPolicy decides whether to issue a certificate for the given host.
// It rejects bare domains, wrong domain suffixes, and subdomains that
// don't resolve to any active service.
func (a *AutoTLSConfig) hostPolicy(ctx context.Context, host string) error {
	// Must be under our domain
	suffix := "." + a.domain
	if !strings.HasSuffix(host, suffix) {
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
