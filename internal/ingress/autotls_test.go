package ingress

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestHostPolicy_ValidResolving(t *testing.T) {
	mock := &mockSubdomainResolver{
		vanity: map[string]string{"myapp": "dep-aabbccdd"},
	}
	resolver := NewResolver(nil, mock)
	resolver.Register(&ServiceEntry{
		DeploymentID:   "dep-aabbccdd",
		ProviderNodeID: "node1",
		ProviderAddr:   "10.0.0.1:9002",
		ContainerPort:  8080,
		HostPort:       32000,
	})

	a := NewAutoTLSConfig(t.TempDir(), "moltbunker.dev", "test@example.com", resolver)

	err := a.hostPolicy(context.Background(), "myapp.moltbunker.dev")
	if err != nil {
		t.Errorf("expected valid subdomain to pass, got: %v", err)
	}
}

func TestHostPolicy_NonExistentSubdomain(t *testing.T) {
	resolver := NewResolver(nil, nil)

	a := NewAutoTLSConfig(t.TempDir(), "moltbunker.dev", "test@example.com", resolver)

	err := a.hostPolicy(context.Background(), "nonexistent.moltbunker.dev")
	if err == nil {
		t.Error("expected non-existent subdomain to be rejected")
	}
	if !strings.Contains(err.Error(), "does not resolve") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestHostPolicy_BareDomain(t *testing.T) {
	a := NewAutoTLSConfig(t.TempDir(), "moltbunker.dev", "test@example.com", nil)

	err := a.hostPolicy(context.Background(), "moltbunker.dev")
	if err == nil {
		t.Error("expected bare domain to be rejected")
	}
	if !strings.Contains(err.Error(), "not under") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestHostPolicy_WrongDomainSuffix(t *testing.T) {
	a := NewAutoTLSConfig(t.TempDir(), "moltbunker.dev", "test@example.com", nil)

	err := a.hostPolicy(context.Background(), "myapp.evil.com")
	if err == nil {
		t.Error("expected wrong domain suffix to be rejected")
	}
	// No custom-domain store wired: original "not under" semantics preserved.
	if !strings.Contains(err.Error(), "not under") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestHostPolicy_CustomDomainVerified(t *testing.T) {
	a := NewAutoTLSConfig(t.TempDir(), "moltbunker.dev", "test@example.com", nil)
	store := NewDomainOwnershipStore(time.Hour)
	store.Store("app.customer.com", "dep-123", "", MethodCNAME)
	a.SetCustomDomains(store)

	if err := a.hostPolicy(context.Background(), "app.customer.com"); err != nil {
		t.Errorf("expected verified custom domain to be accepted, got %v", err)
	}
}

func TestHostPolicy_CustomDomainUnverified(t *testing.T) {
	a := NewAutoTLSConfig(t.TempDir(), "moltbunker.dev", "test@example.com", nil)
	a.SetCustomDomains(NewDomainOwnershipStore(time.Hour)) // empty store

	if err := a.hostPolicy(context.Background(), "app.customer.com"); err == nil {
		t.Error("expected unverified custom domain to be rejected")
	}
}
