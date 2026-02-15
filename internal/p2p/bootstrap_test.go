package p2p

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
)

// mockDNSResolver is a test double for dnsResolver that returns preconfigured
// TXT records or an error.
type mockDNSResolver struct {
	records map[string][]string
	err     error
}

func (m *mockDNSResolver) LookupTXT(_ context.Context, name string) ([]string, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.records[name], nil
}

// mockHTTPClient is a test double for httpDoer.
type mockHTTPClient struct {
	statusCode int
	body       string
	err        error
}

func (m *mockHTTPClient) Do(_ *http.Request) (*http.Response, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &http.Response{
		StatusCode: m.statusCode,
		Body:       io.NopCloser(strings.NewReader(m.body)),
	}, nil
}

func TestResolveDNSBootstrap_Success(t *testing.T) {
	// Use a well-known peer ID for testing
	peerID1 := "QmNnooDu7bfjPFoTZYxMNLWUQJyrVwtbZg5gBMjTezGAJN"
	peerID2 := "QmQCU2EcMqAqQPR2i9bChDtGNJchTbq5TbXJJ16u19uLTa"

	resolver := &mockDNSResolver{
		records: map[string][]string{
			"_dnsaddr.bootstrap.moltbunker.com": {
				fmt.Sprintf("dnsaddr=/ip4/1.2.3.4/tcp/9000/p2p/%s", peerID1),
				fmt.Sprintf("dnsaddr=/ip4/5.6.7.8/tcp/9000/p2p/%s", peerID2),
			},
		},
	}

	cfg := &BootstrapConfig{
		ResolveTimeout: 5 * time.Second,
		resolver:       resolver,
	}

	peers, err := ResolveDNSBootstrap("bootstrap.moltbunker.com", cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(peers) != 2 {
		t.Fatalf("expected 2 peers, got %d", len(peers))
	}

	// Verify peer IDs
	pid1, _ := peer.Decode(peerID1)
	pid2, _ := peer.Decode(peerID2)

	foundPeer1, foundPeer2 := false, false
	for _, pi := range peers {
		if pi.ID == pid1 {
			foundPeer1 = true
		}
		if pi.ID == pid2 {
			foundPeer2 = true
		}
	}

	if !foundPeer1 {
		t.Error("peer 1 not found in resolved peers")
	}
	if !foundPeer2 {
		t.Error("peer 2 not found in resolved peers")
	}
}

func TestResolveDNSBootstrap_MergesAddrsForSamePeer(t *testing.T) {
	peerID := "QmNnooDu7bfjPFoTZYxMNLWUQJyrVwtbZg5gBMjTezGAJN"

	resolver := &mockDNSResolver{
		records: map[string][]string{
			"_dnsaddr.example.io": {
				fmt.Sprintf("dnsaddr=/ip4/1.2.3.4/tcp/9000/p2p/%s", peerID),
				fmt.Sprintf("dnsaddr=/ip4/5.6.7.8/tcp/9001/p2p/%s", peerID),
			},
		},
	}

	cfg := &BootstrapConfig{
		ResolveTimeout: 5 * time.Second,
		resolver:       resolver,
	}

	peers, err := ResolveDNSBootstrap("example.io", cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(peers) != 1 {
		t.Fatalf("expected 1 peer (merged), got %d", len(peers))
	}

	if len(peers[0].Addrs) != 2 {
		t.Errorf("expected 2 addresses for merged peer, got %d", len(peers[0].Addrs))
	}
}

func TestResolveDNSBootstrap_DNSError(t *testing.T) {
	resolver := &mockDNSResolver{
		err: &net.DNSError{
			Err:    "no such host",
			Name:   "_dnsaddr.bootstrap.moltbunker.com",
			Server: "8.8.8.8",
		},
	}

	cfg := &BootstrapConfig{
		ResolveTimeout: 5 * time.Second,
		resolver:       resolver,
	}

	_, err := ResolveDNSBootstrap("bootstrap.moltbunker.com", cfg)
	if err == nil {
		t.Fatal("expected error on DNS failure")
	}
}

func TestResolveDNSBootstrap_SkipsInvalidRecords(t *testing.T) {
	peerID := "QmNnooDu7bfjPFoTZYxMNLWUQJyrVwtbZg5gBMjTezGAJN"

	resolver := &mockDNSResolver{
		records: map[string][]string{
			"_dnsaddr.example.io": {
				"not-a-dnsaddr-record",
				"dnsaddr=invalid-multiaddr",
				fmt.Sprintf("dnsaddr=/ip4/1.2.3.4/tcp/9000/p2p/%s", peerID),
				"dnsaddr=",
				"dnsaddr=/ip4/1.2.3.4/tcp/9000", // no peer ID
			},
		},
	}

	cfg := &BootstrapConfig{
		ResolveTimeout: 5 * time.Second,
		resolver:       resolver,
	}

	peers, err := ResolveDNSBootstrap("example.io", cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(peers) != 1 {
		t.Fatalf("expected 1 valid peer, got %d", len(peers))
	}
}

func TestResolveDNSBootstrap_EmptyRecords(t *testing.T) {
	resolver := &mockDNSResolver{
		records: map[string][]string{
			"_dnsaddr.example.io": {},
		},
	}

	cfg := &BootstrapConfig{
		ResolveTimeout: 5 * time.Second,
		resolver:       resolver,
	}

	peers, err := ResolveDNSBootstrap("example.io", cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(peers) != 0 {
		t.Errorf("expected 0 peers from empty records, got %d", len(peers))
	}
}

func TestResolveBootstrapPeers_FallbackOnDNSFailure(t *testing.T) {
	resolver := &mockDNSResolver{
		err: &net.DNSError{Err: "no such host", Name: "_dnsaddr.bootstrap.moltbunker.com"},
	}

	peerID := "QmNnooDu7bfjPFoTZYxMNLWUQJyrVwtbZg5gBMjTezGAJN"

	cfg := &BootstrapConfig{
		DNSDomains: []string{"bootstrap.moltbunker.com"},
		StaticPeers: []string{
			fmt.Sprintf("/ip4/10.0.0.1/tcp/9000/p2p/%s", peerID),
		},
		ResolveTimeout: 5 * time.Second,
		resolver:       resolver,
	}

	peers := ResolveBootstrapPeers(cfg)
	if len(peers) != 1 {
		t.Fatalf("expected 1 fallback peer, got %d", len(peers))
	}

	pid, _ := peer.Decode(peerID)
	if peers[0].ID != pid {
		t.Errorf("expected peer ID %s, got %s", pid, peers[0].ID)
	}
}

func TestResolveBootstrapPeers_PrefersDNSOverStatic(t *testing.T) {
	dnsPeerID := "QmNnooDu7bfjPFoTZYxMNLWUQJyrVwtbZg5gBMjTezGAJN"
	staticPeerID := "QmQCU2EcMqAqQPR2i9bChDtGNJchTbq5TbXJJ16u19uLTa"

	resolver := &mockDNSResolver{
		records: map[string][]string{
			"_dnsaddr.bootstrap.moltbunker.com": {
				fmt.Sprintf("dnsaddr=/ip4/1.2.3.4/tcp/9000/p2p/%s", dnsPeerID),
			},
		},
	}

	cfg := &BootstrapConfig{
		DNSDomains: []string{"bootstrap.moltbunker.com"},
		StaticPeers: []string{
			fmt.Sprintf("/ip4/10.0.0.1/tcp/9000/p2p/%s", staticPeerID),
		},
		ResolveTimeout: 5 * time.Second,
		resolver:       resolver,
	}

	peers := ResolveBootstrapPeers(cfg)
	if len(peers) != 1 {
		t.Fatalf("expected 1 peer (DNS only), got %d", len(peers))
	}

	pid, _ := peer.Decode(dnsPeerID)
	if peers[0].ID != pid {
		t.Errorf("expected DNS peer %s, got %s", pid, peers[0].ID)
	}
}

func TestResolveBootstrapPeers_MultipleDomains(t *testing.T) {
	peerID1 := "QmNnooDu7bfjPFoTZYxMNLWUQJyrVwtbZg5gBMjTezGAJN"
	peerID2 := "QmQCU2EcMqAqQPR2i9bChDtGNJchTbq5TbXJJ16u19uLTa"

	resolver := &mockDNSResolver{
		records: map[string][]string{
			"_dnsaddr.primary.moltbunker.com": {
				fmt.Sprintf("dnsaddr=/ip4/1.2.3.4/tcp/9000/p2p/%s", peerID1),
			},
			"_dnsaddr.secondary.moltbunker.com": {
				fmt.Sprintf("dnsaddr=/ip4/5.6.7.8/tcp/9000/p2p/%s", peerID2),
			},
		},
	}

	cfg := &BootstrapConfig{
		DNSDomains: []string{
			"primary.moltbunker.com",
			"secondary.moltbunker.com",
		},
		ResolveTimeout: 5 * time.Second,
		resolver:       resolver,
	}

	peers := ResolveBootstrapPeers(cfg)
	if len(peers) != 2 {
		t.Fatalf("expected 2 peers from 2 domains, got %d", len(peers))
	}
}

func TestResolveBootstrapPeers_DeduplicateAcrossDomains(t *testing.T) {
	peerID := "QmNnooDu7bfjPFoTZYxMNLWUQJyrVwtbZg5gBMjTezGAJN"

	resolver := &mockDNSResolver{
		records: map[string][]string{
			"_dnsaddr.primary.moltbunker.com": {
				fmt.Sprintf("dnsaddr=/ip4/1.2.3.4/tcp/9000/p2p/%s", peerID),
			},
			"_dnsaddr.secondary.moltbunker.com": {
				fmt.Sprintf("dnsaddr=/ip4/1.2.3.4/tcp/9000/p2p/%s", peerID),
			},
		},
	}

	cfg := &BootstrapConfig{
		DNSDomains: []string{
			"primary.moltbunker.com",
			"secondary.moltbunker.com",
		},
		ResolveTimeout: 5 * time.Second,
		resolver:       resolver,
	}

	peers := ResolveBootstrapPeers(cfg)
	if len(peers) != 1 {
		t.Fatalf("expected 1 deduplicated peer, got %d", len(peers))
	}
}

func TestResolveBootstrapPeers_NilConfig(t *testing.T) {
	// With nil config, uses defaults. DNS will fail (no real records),
	// so it should fall back to static peers.
	peers := ResolveBootstrapPeers(nil)
	// We can't assert on count since it depends on real DNS resolution,
	// but it should not panic.
	_ = peers
}

func TestDefaultBootstrapConfig(t *testing.T) {
	cfg := DefaultBootstrapConfig()

	if len(cfg.DNSDomains) == 0 {
		t.Error("DefaultBootstrapConfig should have DNS domains")
	}

	if cfg.DNSDomains[0] != DefaultBootstrapDomain {
		t.Errorf("expected domain %s, got %s", DefaultBootstrapDomain, cfg.DNSDomains[0])
	}

	// HTTP endpoints should have the API bootstrap URL
	if len(cfg.HTTPEndpoints) == 0 {
		t.Error("DefaultBootstrapConfig should have HTTP endpoints")
	}
	if cfg.HTTPEndpoints[0] != "https://api.moltbunker.com/v1/bootstrap" {
		t.Errorf("expected HTTP endpoint https://api.moltbunker.com/v1/bootstrap, got %s", cfg.HTTPEndpoints[0])
	}

	// Static peers are intentionally empty — we don't fall back to public IPFS nodes
	if len(cfg.StaticPeers) != 0 {
		t.Error("DefaultBootstrapConfig should NOT have public fallback peers")
	}

	if cfg.ResolveTimeout != DefaultDNSResolveTimeout {
		t.Errorf("expected timeout %v, got %v", DefaultDNSResolveTimeout, cfg.ResolveTimeout)
	}
}

func TestBootstrapConfig_ResolveTimeout(t *testing.T) {
	// Zero timeout should use default
	cfg := &BootstrapConfig{}
	if cfg.resolveTimeout() != DefaultDNSResolveTimeout {
		t.Errorf("expected default timeout, got %v", cfg.resolveTimeout())
	}

	// Custom timeout
	cfg.ResolveTimeout = 30 * time.Second
	if cfg.resolveTimeout() != 30*time.Second {
		t.Errorf("expected 30s timeout, got %v", cfg.resolveTimeout())
	}
}

func TestBootstrapPeerStrings(t *testing.T) {
	peerID := "QmNnooDu7bfjPFoTZYxMNLWUQJyrVwtbZg5gBMjTezGAJN"

	resolver := &mockDNSResolver{
		records: map[string][]string{
			"_dnsaddr.bootstrap.moltbunker.com": {
				fmt.Sprintf("dnsaddr=/ip4/1.2.3.4/tcp/9000/p2p/%s", peerID),
			},
		},
	}

	cfg := &BootstrapConfig{
		DNSDomains:     []string{"bootstrap.moltbunker.com"},
		ResolveTimeout: 5 * time.Second,
		resolver:       resolver,
	}

	strs := BootstrapPeerStrings(cfg)
	if len(strs) != 1 {
		t.Fatalf("expected 1 string, got %d", len(strs))
	}

	expected := fmt.Sprintf("/ip4/1.2.3.4/tcp/9000/p2p/%s", peerID)
	if strs[0] != expected {
		t.Errorf("expected %s, got %s", expected, strs[0])
	}
}

func TestResolveHTTPBootstrap_Success(t *testing.T) {
	peerID := "QmNnooDu7bfjPFoTZYxMNLWUQJyrVwtbZg5gBMjTezGAJN"
	body := fmt.Sprintf(`{"peers":[{"node_id":"abc123","addrs":["/ip4/1.2.3.4/tcp/9000/p2p/%s"],"region":"europe"}]}`, peerID)

	client := &mockHTTPClient{
		statusCode: 200,
		body:       body,
	}

	cfg := &BootstrapConfig{
		httpClient: client,
	}

	peers, err := ResolveHTTPBootstrap("https://api.example.com/v1/bootstrap", cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(peers) != 1 {
		t.Fatalf("expected 1 peer, got %d", len(peers))
	}

	pid, _ := peer.Decode(peerID)
	if peers[0].ID != pid {
		t.Errorf("expected peer ID %s, got %s", pid, peers[0].ID)
	}
}

func TestResolveHTTPBootstrap_ServerError(t *testing.T) {
	client := &mockHTTPClient{
		statusCode: 500,
		body:       `{"error":"internal"}`,
	}

	cfg := &BootstrapConfig{
		httpClient: client,
	}

	_, err := ResolveHTTPBootstrap("https://api.example.com/v1/bootstrap", cfg)
	if err == nil {
		t.Fatal("expected error on 500 response")
	}
}

func TestResolveHTTPBootstrap_InvalidJSON(t *testing.T) {
	client := &mockHTTPClient{
		statusCode: 200,
		body:       `not json`,
	}

	cfg := &BootstrapConfig{
		httpClient: client,
	}

	_, err := ResolveHTTPBootstrap("https://api.example.com/v1/bootstrap", cfg)
	if err == nil {
		t.Fatal("expected error on invalid JSON")
	}
}

func TestResolveHTTPBootstrap_EmptyPeers(t *testing.T) {
	client := &mockHTTPClient{
		statusCode: 200,
		body:       `{"peers":[]}`,
	}

	cfg := &BootstrapConfig{
		httpClient: client,
	}

	peers, err := ResolveHTTPBootstrap("https://api.example.com/v1/bootstrap", cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(peers) != 0 {
		t.Errorf("expected 0 peers, got %d", len(peers))
	}
}

func TestResolveBootstrapPeers_HTTPFallbackOnDNSFailure(t *testing.T) {
	peerID := "QmNnooDu7bfjPFoTZYxMNLWUQJyrVwtbZg5gBMjTezGAJN"

	// DNS fails
	resolver := &mockDNSResolver{
		err: &net.DNSError{Err: "no such host", Name: "_dnsaddr.bootstrap.moltbunker.com"},
	}

	// HTTP succeeds
	body := fmt.Sprintf(`{"peers":[{"node_id":"abc","addrs":["/ip4/2.3.4.5/tcp/9000/p2p/%s"]}]}`, peerID)
	httpClient := &mockHTTPClient{
		statusCode: 200,
		body:       body,
	}

	cfg := &BootstrapConfig{
		DNSDomains: []string{"bootstrap.moltbunker.com"},
		HTTPEndpoints: []string{
			"https://api.moltbunker.com/v1/bootstrap",
		},
		StaticPeers: []string{
			fmt.Sprintf("/ip4/10.0.0.1/tcp/9000/p2p/%s", peerID),
		},
		ResolveTimeout: 5 * time.Second,
		resolver:       resolver,
		httpClient:     httpClient,
	}

	peers := ResolveBootstrapPeers(cfg)
	if len(peers) != 1 {
		t.Fatalf("expected 1 HTTP fallback peer, got %d", len(peers))
	}

	pid, _ := peer.Decode(peerID)
	if peers[0].ID != pid {
		t.Errorf("expected peer ID %s, got %s", pid, peers[0].ID)
	}
}

func TestResolveBootstrapPeers_PartialDomainFailure(t *testing.T) {
	// First domain fails, second succeeds
	peerID := "QmNnooDu7bfjPFoTZYxMNLWUQJyrVwtbZg5gBMjTezGAJN"

	resolver := &partialFailResolver{
		failDomains: map[string]bool{
			"_dnsaddr.failing.moltbunker.com": true,
		},
		records: map[string][]string{
			"_dnsaddr.working.moltbunker.com": {
				fmt.Sprintf("dnsaddr=/ip4/1.2.3.4/tcp/9000/p2p/%s", peerID),
			},
		},
	}

	cfg := &BootstrapConfig{
		DNSDomains: []string{
			"failing.moltbunker.com",
			"working.moltbunker.com",
		},
		ResolveTimeout: 5 * time.Second,
		resolver:       resolver,
	}

	peers := ResolveBootstrapPeers(cfg)
	if len(peers) != 1 {
		t.Fatalf("expected 1 peer from working domain, got %d", len(peers))
	}
}

// partialFailResolver fails on specific domains, succeeds on others.
type partialFailResolver struct {
	failDomains map[string]bool
	records     map[string][]string
}

func (r *partialFailResolver) LookupTXT(_ context.Context, name string) ([]string, error) {
	if r.failDomains[name] {
		return nil, &net.DNSError{Err: "no such host", Name: name}
	}
	return r.records[name], nil
}
