package p2p

import (
	"testing"
)

func TestDiversity_AllowDuringBootstrap(t *testing.T) {
	pde := NewPeerDiversityEnforcer()

	// With < 4 peers, everything should be allowed
	for i := 0; i < 3; i++ {
		if !pde.AllowPeer("Europe", "1.2.0.0/16") {
			t.Fatal("should allow during bootstrap")
		}
		pde.RecordPeer("Europe", "1.2.0.0/16")
	}
}

func TestDiversity_RejectRegionDomination(t *testing.T) {
	pde := NewPeerDiversityEnforcerWithConfig(0.5, 0.3, 2)

	// Add 5 Europe peers
	for i := 0; i < 5; i++ {
		pde.RecordPeer("Europe", "1.2.0.0/16")
	}
	// Add 4 Americas peers
	for i := 0; i < 4; i++ {
		pde.RecordPeer("Americas", "3.4.0.0/16")
	}

	// Now at 9 total. Europe has 5/9 = 0.555 which exceeds 0.5 max.
	// Adding another Europe peer would make it 6/10 = 0.6 — should be rejected.
	if pde.AllowPeer("Europe", "5.6.0.0/16") {
		t.Fatal("should reject Europe peer when region share > 50%")
	}

	// Americas has 4/9 = 0.444. Adding one would make 5/10 = 0.5 — allowed.
	if !pde.AllowPeer("Americas", "7.8.0.0/16") {
		t.Fatal("should allow Americas peer when region share <= 50%")
	}
}

func TestDiversity_RejectSubnetDomination(t *testing.T) {
	pde := NewPeerDiversityEnforcerWithConfig(0.8, 0.3, 1)

	// Add 4 peers from same /16, 6 from others
	for i := 0; i < 4; i++ {
		pde.RecordPeer("Europe", "1.2.0.0/16")
	}
	for i := 0; i < 6; i++ {
		pde.RecordPeer("Europe", "3.4.0.0/16")
	}

	// At 10 total. Subnet 1.2 has 4/10 = 0.4 > 0.3 max.
	// Adding another would be 5/11 = 0.45 — rejected.
	if pde.AllowPeer("Europe", "1.2.0.0/16") {
		t.Fatal("should reject peer from dominant /16 subnet")
	}
}

func TestDiversity_RemovePeer(t *testing.T) {
	pde := NewPeerDiversityEnforcer()

	pde.RecordPeer("Europe", "1.2.0.0/16")
	pde.RecordPeer("Americas", "3.4.0.0/16")

	if pde.TotalPeers() != 2 {
		t.Fatalf("expected 2 peers, got %d", pde.TotalPeers())
	}

	pde.RemovePeer("Europe", "1.2.0.0/16")

	if pde.TotalPeers() != 1 {
		t.Fatalf("expected 1 peer after removal, got %d", pde.TotalPeers())
	}
}

func TestDiversity_RegionCount(t *testing.T) {
	pde := NewPeerDiversityEnforcer()

	pde.RecordPeer("Europe", "1.2.0.0/16")
	pde.RecordPeer("Americas", "3.4.0.0/16")
	pde.RecordPeer("Asia-Pacific", "5.6.0.0/16")

	if pde.RegionCount() != 3 {
		t.Fatalf("expected 3 regions, got %d", pde.RegionCount())
	}
}

func TestDiversity_NeedsDiversity(t *testing.T) {
	pde := NewPeerDiversityEnforcerWithConfig(0.5, 0.3, 2)

	pde.RecordPeer("Europe", "1.2.0.0/16")

	if !pde.NeedsDiversity() {
		t.Fatal("should need diversity with only 1 region")
	}

	pde.RecordPeer("Americas", "3.4.0.0/16")

	if pde.NeedsDiversity() {
		t.Fatal("should not need diversity with 2 regions")
	}
}

func TestDiversity_AllowUnknownRegion(t *testing.T) {
	pde := NewPeerDiversityEnforcer()

	// Unknown region should be treated leniently
	for i := 0; i < 10; i++ {
		pde.RecordPeer("Unknown", "1.2.0.0/16")
	}

	// "Unknown" is skipped in region check
	if !pde.AllowPeer("Unknown", "3.4.0.0/16") {
		t.Fatal("Unknown region should always pass region check")
	}
}

func TestDiversity_EmptySubnet(t *testing.T) {
	pde := NewPeerDiversityEnforcer()

	for i := 0; i < 10; i++ {
		pde.RecordPeer("Europe", "")
	}

	// Empty subnet should always pass subnet check
	if !pde.AllowPeer("Americas", "") {
		t.Fatal("empty subnet should always pass")
	}
}
