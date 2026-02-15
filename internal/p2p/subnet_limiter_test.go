package p2p

import (
	"crypto/rand"
	"testing"

	"github.com/moltbunker/moltbunker/pkg/types"
)

func randomNodeID() types.NodeID {
	var id types.NodeID
	rand.Read(id[:])
	return id
}

func TestSubnetLimiter_AllowUpToMax(t *testing.T) {
	sl := NewSubnetLimiter(3)

	// 3 peers from same /24 should be allowed
	for i := 0; i < 3; i++ {
		nodeID := randomNodeID()
		if !sl.Allow("1.2.3.4:9000", nodeID) {
			t.Fatalf("peer %d should be allowed", i)
		}
	}

	// 4th peer from same /24 should be rejected
	nodeID := randomNodeID()
	if sl.Allow("1.2.3.5:9000", nodeID) {
		t.Fatal("4th peer from same /24 should be rejected")
	}
}

func TestSubnetLimiter_DifferentSubnets(t *testing.T) {
	sl := NewSubnetLimiter(2)

	// 2 from subnet A
	sl.Allow("10.0.1.1:9000", randomNodeID())
	sl.Allow("10.0.1.2:9000", randomNodeID())

	// 2 from subnet B — should be fine
	nodeID := randomNodeID()
	if !sl.Allow("10.0.2.1:9000", nodeID) {
		t.Fatal("different subnet should be allowed")
	}
}

func TestSubnetLimiter_ExemptLocalhost(t *testing.T) {
	sl := NewSubnetLimiter(1)

	// Fill the /24
	sl.Allow("127.0.0.1:9000", randomNodeID())

	// Localhost should always be exempt
	if !sl.Allow("127.0.0.2:9000", randomNodeID()) {
		t.Fatal("localhost should be exempt from subnet limits")
	}
}

func TestSubnetLimiter_ExemptPrivateRanges(t *testing.T) {
	sl := NewSubnetLimiter(1)

	// These should all be exempt
	exemptAddrs := []string{
		"10.0.0.1:9000",
		"172.16.0.1:9000",
		"192.168.1.1:9000",
		"127.0.0.1:9000",
	}

	for i, addr := range exemptAddrs {
		for j := 0; j < 5; j++ {
			if !sl.Allow(addr, randomNodeID()) {
				t.Fatalf("exempt address %d (%s) attempt %d should always be allowed", i, addr, j)
			}
		}
	}
}

func TestSubnetLimiter_ExemptOnion(t *testing.T) {
	sl := NewSubnetLimiter(1)

	for i := 0; i < 5; i++ {
		if !sl.Allow("abc123.onion:9000", randomNodeID()) {
			t.Fatalf(".onion attempt %d should be exempt", i)
		}
	}
}

func TestSubnetLimiter_Remove(t *testing.T) {
	sl := NewSubnetLimiter(2)

	id1 := randomNodeID()
	id2 := randomNodeID()
	sl.Allow("8.8.8.1:9000", id1)
	sl.Allow("8.8.8.2:9000", id2)

	// At limit
	id3 := randomNodeID()
	if sl.Allow("8.8.8.3:9000", id3) {
		t.Fatal("should be at limit")
	}

	// Remove one peer
	sl.Remove(id1)

	// Now should be allowed
	if !sl.Allow("8.8.8.3:9000", id3) {
		t.Fatal("should be allowed after removing a peer")
	}
}

func TestSubnetLimiter_ReconnectionAllowed(t *testing.T) {
	sl := NewSubnetLimiter(1)

	nodeID := randomNodeID()
	if !sl.Allow("8.8.8.1:9000", nodeID) {
		t.Fatal("first connection should be allowed")
	}

	// Same nodeID reconnecting — should be allowed even at limit
	if !sl.Allow("8.8.8.1:9001", nodeID) {
		t.Fatal("reconnection with same nodeID should be allowed")
	}
}

func TestSubnetLimiter_PeerCount(t *testing.T) {
	sl := NewSubnetLimiter(10)

	for i := 0; i < 5; i++ {
		sl.Allow("1.2.3.4:9000", randomNodeID())
	}

	if sl.PeerCount() != 5 {
		t.Fatalf("expected 5 peers, got %d", sl.PeerCount())
	}
}
