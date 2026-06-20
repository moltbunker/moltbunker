//go:build e2e

package golden

import (
	"context"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/moltbunker/moltbunker/internal/ingress"
	"github.com/moltbunker/moltbunker/tests/e2e/testutil"
)

// These tests validate the ingress resolver logic (pure Go, no network calls)
// and the loopback HTTPS origin as standalone legs, independent of the full
// golden-path test. [MOCK: tunnel=loopback HTTP/TLS; REAL: ingress.Resolver]

func TestGoldenPath_LoopbackIngress200(t *testing.T) {
	a := testutil.NewAssertions(t)
	t.Log("[MOCK: loopback origin] resolver-routed request returns HTTPS 200 + golden body")

	li := NewLoopbackIngress(t)
	defer li.Close()

	const deploymentID = "dep-0011223344556677889900aabbccddee"
	li.Seed(deploymentID)

	// Resolve via the 8-char prefix, then issue an HTTPS GET to the resolved
	// origin (the TLS server stands in for the edge's terminating TLS).
	prefix := bareID(deploymentID)[:8]
	entry, err := li.Resolver.Resolve(prefix)
	a.NoError(err, "resolver should map the prefix to a service entry")
	a.NotNil(entry, "resolved entry should be non-nil")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, li.TLSOriginURL(), nil)
	a.NoError(err)
	resp, err := li.TLSClient().Do(req)
	a.NoError(err, "HTTPS GET to the loopback origin should succeed")
	if resp == nil {
		t.Fatal("nil response")
	}
	defer func() { _ = resp.Body.Close() }()

	a.Equal(http.StatusOK, resp.StatusCode, "origin should return 200")
	body, _ := io.ReadAll(resp.Body)
	a.Contains(string(body), goldenPathBody, "body should be the golden-path marker")
}

func TestGoldenPath_ResolverPrefixMatch(t *testing.T) {
	a := testutil.NewAssertions(t)
	t.Log("[REAL: resolver] 8-char hex prefix resolves to the seeded deployment id")

	li := NewLoopbackIngress(t)
	defer li.Close()

	// 32-char hex deployment id (after the dep- prefix).
	const deploymentID = "dep-deadbeefcafe00112233445566778899"
	li.Seed(deploymentID)

	prefix := bareID(deploymentID)[:8] // "deadbeef"
	entry, err := li.Resolver.Resolve(prefix)
	a.NoError(err, "prefix resolution should succeed")
	a.NotNil(entry)
	a.Equal(deploymentID, entry.DeploymentID,
		"resolved entry's DeploymentID should match the seeded deployment")
}

func TestGoldenPath_ResolverExactMatch(t *testing.T) {
	a := testutil.NewAssertions(t)
	t.Log("[REAL: resolver] exact deployment-id match resolves")

	li := NewLoopbackIngress(t)
	defer li.Close()

	const deploymentID = "dep-1234567890abcdef1234567890abcdef"
	li.Resolver.Register(&ingress.ServiceEntry{
		DeploymentID:  deploymentID,
		ProviderAddr:  li.Server.Listener.Addr().String(),
		ContainerPort: 80,
		LastSeen:      time.Now(),
	})

	entry, err := li.Resolver.Resolve(deploymentID)
	a.NoError(err, "exact match should resolve")
	a.NotNil(entry)
	a.Equal(deploymentID, entry.DeploymentID)
}

func TestGoldenPath_ResolverNotFound(t *testing.T) {
	a := testutil.NewAssertions(t)
	t.Log("[REAL: resolver] unknown name returns 'service not found'")

	li := NewLoopbackIngress(t)
	defer li.Close()

	_, err := li.Resolver.Resolve("nonexistentname")
	a.Error(err, "unknown subdomain should not resolve")
	a.ErrorContains(err, "service not found", "error should be the not-found sentinel")
}
