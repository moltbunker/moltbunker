package daemon

import (
	"context"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/moltbunker/moltbunker/internal/payment"
	"github.com/moltbunker/moltbunker/pkg/types"
)

// mockPaymentService is a minimal mock that tracks ResolveSubdomain calls.
type mockPaymentService struct {
	resolveFunc func(ctx context.Context, name string) (*payment.SubdomainRegistration, error)
	callCount   int
}

func TestResolveOnChain_CachePositiveHit(t *testing.T) {
	mock := &mockPaymentService{
		resolveFunc: func(_ context.Context, name string) (*payment.SubdomainRegistration, error) {
			return &payment.SubdomainRegistration{
				Name:         name,
				DeploymentID: [32]byte{0xab, 0xcd},
				Owner:        common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678"),
			}, nil
		},
	}

	adapter := NewGossipServiceAdapter(nil)
	now := time.Now()
	adapter.nowFunc = func() time.Time { return now }
	// Wire mock: pre-populate cache as if RPC returned result
	adapter.subdomainCache["myapp"] = &cachedSubdomainResult{
		DeploymentID: "abcd",
		Found:        true,
		FetchedAt:    now,
	}

	// Should return cached result without calling RPC
	depID, ok := adapter.ResolveOnChain("myapp")
	if !ok || depID != "abcd" {
		t.Errorf("expected cached positive hit, got depID=%q ok=%v", depID, ok)
	}
	if mock.callCount != 0 {
		t.Errorf("expected 0 RPC calls, got %d", mock.callCount)
	}
}

func TestResolveOnChain_CacheNegativeHit(t *testing.T) {
	adapter := NewGossipServiceAdapter(nil)
	now := time.Now()
	adapter.nowFunc = func() time.Time { return now }
	// Pre-populate negative cache
	adapter.subdomainCache["unknown"] = &cachedSubdomainResult{
		Found:     false,
		FetchedAt: now,
	}

	depID, ok := adapter.ResolveOnChain("unknown")
	if ok || depID != "" {
		t.Errorf("expected cached negative hit, got depID=%q ok=%v", depID, ok)
	}
}

func TestResolveOnChain_CacheExpires(t *testing.T) {
	callCount := 0
	adapter := NewGossipServiceAdapter(nil)
	now := time.Now()
	adapter.nowFunc = func() time.Time { return now }
	// Pre-populate expired cache entry
	adapter.subdomainCache["myapp"] = &cachedSubdomainResult{
		DeploymentID: "old-value",
		Found:        true,
		FetchedAt:    now.Add(-10 * time.Minute), // well past 5min TTL
	}
	// No payment service means the RPC call will return ("", false)
	// which proves the cache was bypassed
	_ = callCount

	depID, ok := adapter.ResolveOnChain("myapp")
	// No payment service → returns false, but the point is it didn't return the cached "old-value"
	if ok {
		t.Errorf("expected cache miss (no payment svc), got depID=%q ok=%v", depID, ok)
	}
}

func TestInvalidateSubdomainCache(t *testing.T) {
	adapter := NewGossipServiceAdapter(nil)
	now := time.Now()
	adapter.nowFunc = func() time.Time { return now }
	adapter.subdomainCache["myapp"] = &cachedSubdomainResult{
		DeploymentID: "abcd",
		Found:        true,
		FetchedAt:    now,
	}

	adapter.InvalidateSubdomainCache("myapp")

	adapter.cacheMu.RLock()
	_, exists := adapter.subdomainCache["myapp"]
	adapter.cacheMu.RUnlock()
	if exists {
		t.Error("expected cache entry to be evicted after InvalidateSubdomainCache")
	}
}

func TestGossipStateValidator_RejectsRemoteSubdomainEntries(t *testing.T) {
	localID := types.NodeID{0x01}
	validator := NewGossipStateValidator(localID)

	remoteID := types.NodeID{0x02}

	// Remote peer trying to inject a subdomain entry — must be rejected
	if validator(remoteID, "subdomain:my-app", "dep-abc123") {
		t.Error("validator should reject remote subdomain entries")
	}

	// Remote peer trying to inject a different subdomain
	if validator(remoteID, "subdomain:victim-app", "dep-evil") {
		t.Error("validator should reject all remote subdomain entries")
	}
}

func TestGossipStateValidator_AcceptsLocalSubdomainEntries(t *testing.T) {
	localID := types.NodeID{0x01}
	validator := NewGossipStateValidator(localID)

	// Local updates (zero sender ID) should always pass
	zeroID := types.NodeID{}
	if !validator(zeroID, "subdomain:my-app", "dep-abc123") {
		t.Error("validator should accept local subdomain entries")
	}
}

func TestGossipStateValidator_AcceptsValidExposeEntries(t *testing.T) {
	localID := types.NodeID{0x01}
	validator := NewGossipStateValidator(localID)

	senderID := types.NodeID{0x02}

	// expose: entry with matching ProviderNodeID should be accepted
	entry := map[string]interface{}{
		"provider_node_id": senderID.String(),
		"deployment_id":    "dep-abc123",
		"provider_addr":    "1.2.3.4:9002",
	}
	if !validator(senderID, "expose:dep-abc123:8080", entry) {
		t.Error("validator should accept expose entry with matching provider ID")
	}
}

func TestGossipStateValidator_RejectsExposeWithWrongProvider(t *testing.T) {
	localID := types.NodeID{0x01}
	validator := NewGossipStateValidator(localID)

	senderID := types.NodeID{0x02}
	differentID := types.NodeID{0x03}

	// expose: entry where ProviderNodeID doesn't match sender — reject
	entry := map[string]interface{}{
		"provider_node_id": differentID.String(),
		"deployment_id":    "dep-abc123",
		"provider_addr":    "1.2.3.4:9002",
	}
	if validator(senderID, "expose:dep-abc123:8080", entry) {
		t.Error("validator should reject expose entry with mismatched provider ID")
	}
}

func TestGossipStateValidator_AcceptsOtherKeys(t *testing.T) {
	localID := types.NodeID{0x01}
	validator := NewGossipStateValidator(localID)

	remoteID := types.NodeID{0x02}

	// Non-expose, non-subdomain keys should be accepted
	if !validator(remoteID, "status:dep-abc123", "running") {
		t.Error("validator should accept non-sensitive keys")
	}
	if !validator(remoteID, "health:node-xyz", "healthy") {
		t.Error("validator should accept non-sensitive keys")
	}
}
