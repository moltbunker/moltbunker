package p2p

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"

	"github.com/moltbunker/moltbunker/internal/logging"
)

const (
	// DefaultDNSResolveTimeout is the timeout for DNS-based bootstrap resolution.
	DefaultDNSResolveTimeout = 10 * time.Second

	// DefaultBootstrapDomain is the primary DNS domain for bootstrap peers.
	DefaultBootstrapDomain = "bootstrap.moltbunker.com"

	// dnsaddrPrefix is the TXT record prefix for dnsaddr multiaddrs.
	dnsaddrPrefix = "dnsaddr="
)

// BootstrapConfig holds configuration for bootstrap peer resolution,
// combining DNS-based discovery with HTTP and static fallback peers.
type BootstrapConfig struct {
	// DNSDomains is the list of DNS domains to resolve for bootstrap peers.
	// Each domain should have _dnsaddr.<domain> TXT records containing
	// multiaddr strings prefixed with "dnsaddr=".
	DNSDomains []string

	// HTTPEndpoints is a list of HTTP(S) URLs to query for bootstrap peers
	// when DNS resolution fails. The endpoint must return JSON matching
	// the BootstrapResponse format ({"peers": [...]}).
	HTTPEndpoints []string

	// StaticPeers is a list of static fallback multiaddr strings used when
	// both DNS and HTTP resolution fail or return no results.
	StaticPeers []string

	// ResolveTimeout is the maximum time allowed for DNS resolution per domain.
	// Defaults to DefaultDNSResolveTimeout if zero.
	ResolveTimeout time.Duration

	// resolver is an injectable DNS resolver for testing. If nil, the
	// default net.Resolver is used.
	resolver dnsResolver

	// httpClient is an injectable HTTP client for testing. If nil, a
	// default client with 10s timeout is used.
	httpClient httpDoer
}

// dnsResolver abstracts DNS TXT lookups for testing.
type dnsResolver interface {
	LookupTXT(ctx context.Context, name string) ([]string, error)
}

// httpDoer abstracts HTTP requests for testing.
type httpDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// DefaultBootstrapConfig returns a BootstrapConfig with sensible defaults.
// StaticPeers is empty — we do NOT use public IPFS nodes as fallbacks to
// prevent untrusted peers from influencing our DHT routing table.
// Once testnet bootstrap nodes are deployed, add them to StaticPeers here.
func DefaultBootstrapConfig() *BootstrapConfig {
	return &BootstrapConfig{
		DNSDomains: []string{
			DefaultBootstrapDomain,
		},
		HTTPEndpoints: []string{
			"https://api.moltbunker.com/v1/bootstrap",
		},
		StaticPeers:    []string{},
		ResolveTimeout: DefaultDNSResolveTimeout,
	}
}

// resolveTimeout returns the configured timeout or the default.
func (bc *BootstrapConfig) resolveTimeout() time.Duration {
	if bc.ResolveTimeout > 0 {
		return bc.ResolveTimeout
	}
	return DefaultDNSResolveTimeout
}

// getResolver returns the configured resolver or the default net.Resolver.
func (bc *BootstrapConfig) getResolver() dnsResolver {
	if bc.resolver != nil {
		return bc.resolver
	}
	return &net.Resolver{}
}

// getHTTPClient returns the configured HTTP client or a default with 10s timeout.
func (bc *BootstrapConfig) getHTTPClient() httpDoer {
	if bc.httpClient != nil {
		return bc.httpClient
	}
	return &http.Client{Timeout: 10 * time.Second}
}

// httpBootstrapPeer matches the JSON format returned by GET /v1/bootstrap.
type httpBootstrapPeer struct {
	NodeID  string   `json:"node_id"`
	Addrs   []string `json:"addrs"`
	Region  string   `json:"region,omitempty"`
	Country string   `json:"country,omitempty"`
}

// httpBootstrapResponse matches the JSON format returned by GET /v1/bootstrap.
type httpBootstrapResponse struct {
	Peers []httpBootstrapPeer `json:"peers"`
}

// ResolveHTTPBootstrap fetches bootstrap peers from an HTTP(S) endpoint.
// The endpoint must return JSON with {"peers": [{node_id, addrs, ...}]}.
// Each addr entry should be a multiaddr string including /p2p/<peerID>.
func ResolveHTTPBootstrap(url string, cfg *BootstrapConfig) ([]peer.AddrInfo, error) {
	if cfg == nil {
		cfg = DefaultBootstrapConfig()
	}

	httpClient := cfg.getHTTPClient()

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create HTTP request for %s: %w", url, err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP bootstrap request to %s failed: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP bootstrap %s returned status %d", url, resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1*1024*1024)) // 1MB limit
	if err != nil {
		return nil, fmt.Errorf("read HTTP bootstrap response from %s: %w", url, err)
	}

	var result httpBootstrapResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse HTTP bootstrap response from %s: %w", url, err)
	}

	var peers []peer.AddrInfo
	seen := make(map[peer.ID]int)

	for _, bp := range result.Peers {
		for _, addrStr := range bp.Addrs {
			maddr, err := ma.NewMultiaddr(addrStr)
			if err != nil {
				logging.Debug("skipping invalid multiaddr from HTTP bootstrap",
					"url", url,
					"addr", addrStr,
					"error", err,
					logging.Component("bootstrap"))
				continue
			}

			addrInfo, err := peer.AddrInfoFromP2pAddr(maddr)
			if err != nil {
				logging.Debug("skipping multiaddr without peer ID from HTTP bootstrap",
					"url", url,
					"addr", addrStr,
					"error", err,
					logging.Component("bootstrap"))
				continue
			}

			if idx, exists := seen[addrInfo.ID]; exists {
				peers[idx].Addrs = append(peers[idx].Addrs, addrInfo.Addrs...)
			} else {
				seen[addrInfo.ID] = len(peers)
				peers = append(peers, *addrInfo)
			}
		}
	}

	return peers, nil
}

// ResolveDNSBootstrap resolves _dnsaddr TXT records for the given domain
// and returns the parsed peer addresses. The TXT record format follows the
// dnsaddr multiaddr convention: each TXT record value is prefixed with
// "dnsaddr=" followed by a full multiaddr string including /p2p/<peerID>.
//
// For example, a TXT record at _dnsaddr.bootstrap.moltbunker.com might contain:
//
//	dnsaddr=/ip4/1.2.3.4/tcp/9000/p2p/12D3KooW...
func ResolveDNSBootstrap(domain string, cfg *BootstrapConfig) ([]peer.AddrInfo, error) {
	if cfg == nil {
		cfg = DefaultBootstrapConfig()
	}

	resolver := cfg.getResolver()
	timeout := cfg.resolveTimeout()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	lookupName := "_dnsaddr." + domain
	txtRecords, err := resolver.LookupTXT(ctx, lookupName)
	if err != nil {
		return nil, fmt.Errorf("DNS TXT lookup for %s failed: %w", lookupName, err)
	}

	var peers []peer.AddrInfo
	seen := make(map[peer.ID]int) // peerID -> index in peers slice

	for _, txt := range txtRecords {
		if !strings.HasPrefix(txt, dnsaddrPrefix) {
			continue
		}

		addrStr := strings.TrimPrefix(txt, dnsaddrPrefix)
		addrStr = strings.TrimSpace(addrStr)
		if addrStr == "" {
			continue
		}

		maddr, err := ma.NewMultiaddr(addrStr)
		if err != nil {
			logging.Debug("skipping invalid multiaddr from DNS",
				"domain", domain,
				"addr", addrStr,
				"error", err,
				logging.Component("bootstrap"))
			continue
		}

		addrInfo, err := peer.AddrInfoFromP2pAddr(maddr)
		if err != nil {
			logging.Debug("skipping multiaddr without peer ID",
				"domain", domain,
				"addr", addrStr,
				"error", err,
				logging.Component("bootstrap"))
			continue
		}

		// Merge addresses for the same peer
		if idx, exists := seen[addrInfo.ID]; exists {
			peers[idx].Addrs = append(peers[idx].Addrs, addrInfo.Addrs...)
		} else {
			seen[addrInfo.ID] = len(peers)
			peers = append(peers, *addrInfo)
		}
	}

	return peers, nil
}

// ResolveBootstrapPeers resolves all configured DNS domains and combines
// results with static fallback peers. If DNS resolution fails or returns
// no peers, static fallback peers are used instead.
func ResolveBootstrapPeers(cfg *BootstrapConfig) []peer.AddrInfo {
	if cfg == nil {
		cfg = DefaultBootstrapConfig()
	}

	var allPeers []peer.AddrInfo
	seen := make(map[peer.ID]bool)

	// Resolve each DNS domain
	for _, domain := range cfg.DNSDomains {
		resolved, err := ResolveDNSBootstrap(domain, cfg)
		if err != nil {
			logging.Warn("DNS bootstrap resolution failed, will use fallback",
				"domain", domain,
				"error", err,
				logging.Component("bootstrap"))
			continue
		}

		for _, pi := range resolved {
			if !seen[pi.ID] {
				seen[pi.ID] = true
				allPeers = append(allPeers, pi)
			}
		}
	}

	// If DNS returned results, use them
	if len(allPeers) > 0 {
		logging.Info("resolved bootstrap peers via DNS",
			"count", len(allPeers),
			logging.Component("bootstrap"))
		return allPeers
	}

	// Try HTTP endpoints
	for _, url := range cfg.HTTPEndpoints {
		resolved, err := ResolveHTTPBootstrap(url, cfg)
		if err != nil {
			logging.Warn("HTTP bootstrap resolution failed, will try next",
				"url", url,
				"error", err,
				logging.Component("bootstrap"))
			continue
		}

		for _, pi := range resolved {
			if !seen[pi.ID] {
				seen[pi.ID] = true
				allPeers = append(allPeers, pi)
			}
		}
	}

	if len(allPeers) > 0 {
		logging.Info("resolved bootstrap peers via HTTP",
			"count", len(allPeers),
			logging.Component("bootstrap"))
		return allPeers
	}

	// Fallback to static peers
	logging.Info("using static fallback bootstrap peers",
		"count", len(cfg.StaticPeers),
		logging.Component("bootstrap"))

	for _, addrStr := range cfg.StaticPeers {
		maddr, err := ma.NewMultiaddr(addrStr)
		if err != nil {
			continue
		}
		addrInfo, err := peer.AddrInfoFromP2pAddr(maddr)
		if err != nil {
			// Try parsing as a full string (e.g. /dnsaddr/...)
			pi, parseErr := peer.AddrInfoFromString(addrStr)
			if parseErr != nil {
				continue
			}
			addrInfo = pi
		}
		if !seen[addrInfo.ID] {
			seen[addrInfo.ID] = true
			allPeers = append(allPeers, *addrInfo)
		}
	}

	return allPeers
}

// BootstrapPeerStrings converts a BootstrapConfig into a list of multiaddr
// strings suitable for DHTConfig.BootstrapPeers. It resolves DNS domains
// and falls back to static peers, returning the results as strings.
func BootstrapPeerStrings(cfg *BootstrapConfig) []string {
	peers := ResolveBootstrapPeers(cfg)
	var result []string
	for _, pi := range peers {
		for _, addr := range pi.Addrs {
			full := fmt.Sprintf("%s/p2p/%s", addr.String(), pi.ID.String())
			result = append(result, full)
		}
	}
	return result
}
