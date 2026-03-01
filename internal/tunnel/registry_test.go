package tunnel

import (
	"sync"
	"testing"
	"time"

	"github.com/hashicorp/yamux"
	"github.com/moltbunker/moltbunker/pkg/types"
)

func testNodeID(b byte) types.NodeID {
	var id types.NodeID
	id[0] = b
	return id
}

func TestTunnelRegistry_RegisterLookup(t *testing.T) {
	reg := NewTunnelRegistry()

	sess := &TunnelSession{
		NodeID:       testNodeID(1),
		Subdomain:    "abc12345",
		RegisteredAt: time.Now(),
		Tier:         "free",
		Limits:       FreeTierLimits,
	}

	if err := reg.Register("abc12345", sess); err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	got, ok := reg.Lookup("abc12345")
	if !ok {
		t.Fatal("Lookup returned false for registered subdomain")
	}
	if got.NodeID != testNodeID(1) {
		t.Errorf("NodeID = %v, want %v", got.NodeID, testNodeID(1))
	}
}

func TestTunnelRegistry_DuplicateSubdomain(t *testing.T) {
	reg := NewTunnelRegistry()

	sess := &TunnelSession{NodeID: testNodeID(1), Subdomain: "dup"}
	if err := reg.Register("dup", sess); err != nil {
		t.Fatal(err)
	}

	sess2 := &TunnelSession{NodeID: testNodeID(2), Subdomain: "dup"}
	if err := reg.Register("dup", sess2); err == nil {
		t.Fatal("expected error for duplicate subdomain")
	}
}

func TestTunnelRegistry_Unregister(t *testing.T) {
	reg := NewTunnelRegistry()

	sess := &TunnelSession{NodeID: testNodeID(1), Subdomain: "gone"}
	reg.Register("gone", sess)
	reg.Unregister("gone")

	_, ok := reg.Lookup("gone")
	if ok {
		t.Fatal("Lookup should return false after Unregister")
	}

	if count := reg.CountForNodeID(testNodeID(1)); count != 0 {
		t.Errorf("CountForNodeID = %d after unregister, want 0", count)
	}
}

func TestTunnelRegistry_CountForNodeID(t *testing.T) {
	reg := NewTunnelRegistry()
	nid := testNodeID(5)

	for i := 0; i < 3; i++ {
		sub := string(rune('a'+i)) + "1234567"
		reg.Register(sub, &TunnelSession{NodeID: nid, Subdomain: sub})
	}

	if got := reg.CountForNodeID(nid); got != 3 {
		t.Errorf("CountForNodeID = %d, want 3", got)
	}
}

func TestTunnelRegistry_UnregisterAll(t *testing.T) {
	reg := NewTunnelRegistry()
	nid := testNodeID(7)

	reg.Register("sub1", &TunnelSession{NodeID: nid, Subdomain: "sub1"})
	reg.Register("sub2", &TunnelSession{NodeID: nid, Subdomain: "sub2"})

	removed := reg.UnregisterAll(nid)
	if len(removed) != 2 {
		t.Errorf("UnregisterAll returned %d items, want 2", len(removed))
	}

	if reg.ActiveCount() != 0 {
		t.Errorf("ActiveCount = %d after UnregisterAll, want 0", reg.ActiveCount())
	}
}

func TestTunnelRegistry_AssignRandomSubdomain(t *testing.T) {
	reg := NewTunnelRegistry()

	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		sub, err := reg.AssignRandomSubdomain()
		if err != nil {
			t.Fatal(err)
		}
		if len(sub) != 8 {
			t.Errorf("subdomain %q has length %d, want 8", sub, len(sub))
		}
		if seen[sub] {
			t.Errorf("duplicate subdomain generated: %s", sub)
		}
		seen[sub] = true
	}
}

func TestTunnelRegistry_ConcurrentAccess(t *testing.T) {
	reg := NewTunnelRegistry()
	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			nid := testNodeID(byte(idx))
			sub, _ := reg.AssignRandomSubdomain()
			reg.Register(sub, &TunnelSession{NodeID: nid, Subdomain: sub})
			reg.Lookup(sub)
			reg.CountForNodeID(nid)
			reg.Unregister(sub)
		}(i)
	}

	wg.Wait()
}

// Suppress unused yamux import for tests that may use it.
var _ = (*yamux.Session)(nil)
