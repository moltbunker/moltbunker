package daemon

import (
	"context"
	"testing"
	"time"

	"github.com/moltbunker/moltbunker/internal/tunnel"
)

// mockPortResolver implements tunnel.PortResolver for testing.
type mockPortResolver struct{}

func (r *mockPortResolver) ResolveDeploymentPort(_ string, port int) (string, error) {
	return "127.0.0.1:9999", nil
}

func TestReverseTunnelManager_ExposeUnexpose(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	callCount := 0
	factory := func() *tunnel.ReverseClient {
		callCount++
		// Return a client pointed at a non-existent address — it will fail to connect
		// but that's fine for testing the manager's tracking logic.
		return tunnel.NewReverseClient("127.0.0.1:1", &mockPortResolver{}, nil)
	}

	rtm := NewReverseTunnelManager(ctx, factory)
	defer rtm.Stop()

	// Expose a deployment
	rtm.Expose("dep-abc123", 8080)
	time.Sleep(50 * time.Millisecond) // Let goroutine start

	if rtm.ActiveCount() != 1 {
		t.Errorf("ActiveCount = %d, want 1", rtm.ActiveCount())
	}

	// Duplicate expose should be idempotent
	rtm.Expose("dep-abc123", 8080)
	time.Sleep(50 * time.Millisecond)

	if rtm.ActiveCount() != 1 {
		t.Errorf("ActiveCount after duplicate = %d, want 1", rtm.ActiveCount())
	}
	if callCount != 1 {
		t.Errorf("factory called %d times, want 1 (duplicate should be no-op)", callCount)
	}

	// Different port should create a new connection
	rtm.Expose("dep-abc123", 3000)
	time.Sleep(50 * time.Millisecond)

	if rtm.ActiveCount() != 2 {
		t.Errorf("ActiveCount after second port = %d, want 2", rtm.ActiveCount())
	}

	// Unexpose all tunnels for the deployment
	rtm.Unexpose("dep-abc123")
	time.Sleep(50 * time.Millisecond)

	if rtm.ActiveCount() != 0 {
		t.Errorf("ActiveCount after unexpose = %d, want 0", rtm.ActiveCount())
	}
}

func TestReverseTunnelManager_Stop(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	factory := func() *tunnel.ReverseClient {
		return tunnel.NewReverseClient("127.0.0.1:1", &mockPortResolver{}, nil)
	}

	rtm := NewReverseTunnelManager(ctx, factory)

	rtm.Expose("dep-1", 80)
	rtm.Expose("dep-2", 80)
	time.Sleep(50 * time.Millisecond)

	if rtm.ActiveCount() != 2 {
		t.Errorf("ActiveCount = %d, want 2", rtm.ActiveCount())
	}

	rtm.Stop()

	if rtm.ActiveCount() != 0 {
		t.Errorf("ActiveCount after Stop = %d, want 0", rtm.ActiveCount())
	}
}

func TestReverseTunnelManager_SubdomainTracking(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	factory := func() *tunnel.ReverseClient {
		return tunnel.NewReverseClient("127.0.0.1:1", &mockPortResolver{}, nil)
	}

	rtm := NewReverseTunnelManager(ctx, factory)
	defer rtm.Stop()

	// Before expose, subdomain should be empty
	if sub := rtm.Subdomain("dep-x", 80); sub != "" {
		t.Errorf("Subdomain before expose = %q, want empty", sub)
	}

	rtm.Expose("dep-x", 80)
	time.Sleep(50 * time.Millisecond)

	// Subdomain will be empty since the client can't actually connect
	// But the entry should exist
	if rtm.ActiveCount() != 1 {
		t.Errorf("ActiveCount = %d, want 1", rtm.ActiveCount())
	}
}
