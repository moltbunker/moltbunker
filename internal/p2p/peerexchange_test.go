package p2p

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"

	"github.com/moltbunker/moltbunker/pkg/types"
)

func TestNewPeerExchangeProtocol_Defaults(t *testing.T) {
	router := NewRouter(nil, nil)
	defer router.Close()

	ab := NewAddressBook()
	pex := NewPeerExchangeProtocol(router, ab, nil)

	if pex == nil {
		t.Fatal("NewPeerExchangeProtocol should not return nil")
	}

	if pex.config.Interval != DefaultPeerExchangeInterval {
		t.Errorf("expected default interval %v, got %v", DefaultPeerExchangeInterval, pex.config.Interval)
	}

	if pex.config.MaxSharedPeers != DefaultMaxSharedPeers {
		t.Errorf("expected default max shared peers %d, got %d", DefaultMaxSharedPeers, pex.config.MaxSharedPeers)
	}
}

func TestNewPeerExchangeProtocol_CustomConfig(t *testing.T) {
	router := NewRouter(nil, nil)
	defer router.Close()

	ab := NewAddressBook()
	cfg := &PeerExchangeConfig{
		Interval:       2 * time.Minute,
		MaxSharedPeers: 10,
	}

	pex := NewPeerExchangeProtocol(router, ab, cfg)

	if pex.config.Interval != 2*time.Minute {
		t.Errorf("expected interval 2m, got %v", pex.config.Interval)
	}
	if pex.config.MaxSharedPeers != 10 {
		t.Errorf("expected max shared peers 10, got %d", pex.config.MaxSharedPeers)
	}
}

func TestNewPeerExchangeProtocol_NilRouter(t *testing.T) {
	ab := NewAddressBook()
	pex := NewPeerExchangeProtocol(nil, ab, nil)

	if pex == nil {
		t.Fatal("NewPeerExchangeProtocol should not return nil with nil router")
	}
}

func TestHandlePeerExchange_EmptyAddressBook(t *testing.T) {
	router := NewRouter(nil, nil)
	defer router.Close()

	ab := NewAddressBook()
	pex := NewPeerExchangeProtocol(router, ab, nil)

	msg := &PeerExchangeMessage{RequestID: "test-1"}
	result := pex.HandlePeerExchange(msg)

	if len(result) != 0 {
		t.Errorf("expected 0 peers from empty address book, got %d", len(result))
	}
}

func TestHandlePeerExchange_ReturnsPeers(t *testing.T) {
	router := NewRouter(nil, nil)
	defer router.Close()

	ab := NewAddressBook()

	// Add some peers to the address book
	for i := 0; i < 5; i++ {
		pid := peer.ID(fmt.Sprintf("test-peer-%d", i))
		addr, _ := ma.NewMultiaddr(fmt.Sprintf("/ip4/10.0.0.%d/tcp/9000", i+1))
		ab.AddPeer(pid, []ma.Multiaddr{addr}, "test")
		ab.RecordConnectionAttempt(pid, true)
	}

	pex := NewPeerExchangeProtocol(router, ab, nil)

	msg := &PeerExchangeMessage{RequestID: "test-2"}
	result := pex.HandlePeerExchange(msg)

	if len(result) != 5 {
		t.Errorf("expected 5 peers, got %d", len(result))
	}

	for _, pa := range result {
		if len(pa.Addrs) == 0 {
			t.Error("peer address should have at least one address")
		}
		if pa.NodeID == "" {
			t.Error("peer address should have a node ID")
		}
	}
}

func TestHandlePeerExchange_LimitsToMaxShared(t *testing.T) {
	router := NewRouter(nil, nil)
	defer router.Close()

	ab := NewAddressBook()

	// Add more peers than MaxSharedPeers
	for i := 0; i < 30; i++ {
		pid := peer.ID(fmt.Sprintf("peer-%d", i))
		addr, _ := ma.NewMultiaddr(fmt.Sprintf("/ip4/10.0.0.%d/tcp/9000", i+1))
		ab.AddPeer(pid, []ma.Multiaddr{addr}, "test")
	}

	cfg := &PeerExchangeConfig{MaxSharedPeers: 10}
	pex := NewPeerExchangeProtocol(router, ab, cfg)

	msg := &PeerExchangeMessage{RequestID: "test-3"}
	result := pex.HandlePeerExchange(msg)

	if len(result) > 10 {
		t.Errorf("expected at most 10 peers, got %d", len(result))
	}
}

func TestHandlePeerExchange_NilAddressBook(t *testing.T) {
	router := NewRouter(nil, nil)
	defer router.Close()

	pex := NewPeerExchangeProtocol(router, nil, nil)

	msg := &PeerExchangeMessage{RequestID: "test-4"}
	result := pex.HandlePeerExchange(msg)

	if result != nil {
		t.Errorf("expected nil from nil address book, got %v", result)
	}
}

func TestHandleRequest_RegisteredOnRouter(t *testing.T) {
	router := NewRouter(nil, nil)
	defer router.Close()

	ab := NewAddressBook()
	pid := peer.ID("test-peer-registered")
	addr, _ := ma.NewMultiaddr("/ip4/10.0.0.1/tcp/9000")
	ab.AddPeer(pid, []ma.Multiaddr{addr}, "test")

	_ = NewPeerExchangeProtocol(router, ab, nil)

	// Verify request handler is registered by creating a message
	reqMsg := PeerExchangeMessage{RequestID: "req-1"}
	payload, _ := json.Marshal(reqMsg)

	msg := &types.Message{
		Type:    MessageTypePeerExchangeRequest,
		Payload: payload,
		Version: types.ProtocolVersion,
	}

	// HandleMessage should not return "no handler" error
	err := router.HandleMessage(context.Background(), msg, nil)
	// The handler sends a response to 'from'; with nil from, it should succeed
	// (just not send the response)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestHandleResponse_MatchesPendingRequest(t *testing.T) {
	router := NewRouter(nil, nil)
	defer router.Close()

	ab := NewAddressBook()
	pex := NewPeerExchangeProtocol(router, ab, nil)

	// Simulate a pending request
	requestID := "test-req-match"
	respCh := make(chan []PeerExchangeAddr, 1)
	pex.pendingRequestsMu.Lock()
	pex.pendingRequests[requestID] = respCh
	pex.pendingRequestsMu.Unlock()

	// Build a response
	respMsg := PeerExchangeMessage{
		RequestID: requestID,
		Peers: []PeerExchangeAddr{
			{NodeID: "peer-abc", Addrs: []string{"/ip4/1.2.3.4/tcp/9000"}},
		},
	}
	payload, _ := json.Marshal(respMsg)

	msg := &types.Message{
		Type:    MessageTypePeerExchangeResponse,
		Payload: payload,
		Version: types.ProtocolVersion,
	}

	err := router.HandleMessage(context.Background(), msg, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check the channel received the peers
	select {
	case peers := <-respCh:
		if len(peers) != 1 {
			t.Errorf("expected 1 peer, got %d", len(peers))
		}
		if peers[0].NodeID != "peer-abc" {
			t.Errorf("expected peer-abc, got %s", peers[0].NodeID)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for response on channel")
	}
}

func TestHandleResponse_UnknownRequestID(t *testing.T) {
	router := NewRouter(nil, nil)
	defer router.Close()

	ab := NewAddressBook()
	_ = NewPeerExchangeProtocol(router, ab, nil)

	// Response with no matching request - should not error
	respMsg := PeerExchangeMessage{
		RequestID: "nonexistent-req",
		Peers:     []PeerExchangeAddr{{NodeID: "peer-xyz"}},
	}
	payload, _ := json.Marshal(respMsg)

	msg := &types.Message{
		Type:    MessageTypePeerExchangeResponse,
		Payload: payload,
		Version: types.ProtocolVersion,
	}

	err := router.HandleMessage(context.Background(), msg, nil)
	if err != nil {
		t.Errorf("unexpected error for unknown request ID: %v", err)
	}
}

func TestIntegrateDiscoveredPeers(t *testing.T) {
	router := NewRouterWithConfig(nil, nil, 50, DefaultConnectionTimeout)
	defer router.Close()

	ab := NewAddressBook()
	pex := NewPeerExchangeProtocol(router, ab, nil)

	peers := []PeerExchangeAddr{
		{
			NodeID:  "0101010101010101010101010101010101010101010101010101010101010101",
			Addrs:   []string{"/ip4/10.0.0.1/tcp/9000"},
			Region:  "Americas",
			Country: "US",
		},
		{
			NodeID:  "0202020202020202020202020202020202020202020202020202020202020202",
			Addrs:   []string{"/ip4/10.0.0.2/tcp/9000"},
			Region:  "Europe",
			Country: "DE",
		},
	}

	pex.integrateDiscoveredPeers(peers)

	// Peers should have been added to the router
	routerPeers := router.GetPeers()
	if len(routerPeers) != 2 {
		t.Errorf("expected 2 peers in router, got %d", len(routerPeers))
	}
}

func TestIntegrateDiscoveredPeers_SkipsEmptyAddrs(t *testing.T) {
	router := NewRouterWithConfig(nil, nil, 50, DefaultConnectionTimeout)
	defer router.Close()

	ab := NewAddressBook()
	pex := NewPeerExchangeProtocol(router, ab, nil)

	peers := []PeerExchangeAddr{
		{NodeID: "0303030303030303030303030303030303030303030303030303030303030303", Addrs: nil},
		{NodeID: "0404040404040404040404040404040404040404040404040404040404040404", Addrs: []string{}},
		{NodeID: "0505050505050505050505050505050505050505050505050505050505050505", Addrs: []string{"/ip4/10.0.0.1/tcp/9000"}},
	}

	pex.integrateDiscoveredPeers(peers)

	routerPeers := router.GetPeers()
	if len(routerPeers) != 1 {
		t.Errorf("expected 1 peer (only one with addrs), got %d", len(routerPeers))
	}
}

func TestIntegrateDiscoveredPeers_Empty(t *testing.T) {
	router := NewRouter(nil, nil)
	defer router.Close()

	ab := NewAddressBook()
	pex := NewPeerExchangeProtocol(router, ab, nil)

	// Should not panic
	pex.integrateDiscoveredPeers(nil)
	pex.integrateDiscoveredPeers([]PeerExchangeAddr{})
}

func TestSortPeersByQuality(t *testing.T) {
	now := time.Now()

	entries := []*AddressEntry{
		{
			PeerID:       peer.ID("low-quality"),
			ConnAttempts: 10,
			ConnSuccess:  1,
			LastSeen:     now.Add(-48 * time.Hour),
		},
		{
			PeerID:       peer.ID("high-quality"),
			ConnAttempts: 10,
			ConnSuccess:  10,
			LastSeen:     now.Add(-30 * time.Minute),
		},
		{
			PeerID:       peer.ID("medium-quality"),
			ConnAttempts: 10,
			ConnSuccess:  5,
			LastSeen:     now.Add(-3 * time.Hour),
		},
	}

	sorted := SortPeersByQuality(entries)

	if len(sorted) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(sorted))
	}

	// high-quality should be first (100% success + recent)
	if sorted[0].PeerID != peer.ID("high-quality") {
		t.Errorf("expected high-quality first, got %s", sorted[0].PeerID)
	}

	// medium-quality second
	if sorted[1].PeerID != peer.ID("medium-quality") {
		t.Errorf("expected medium-quality second, got %s", sorted[1].PeerID)
	}

	// low-quality last
	if sorted[2].PeerID != peer.ID("low-quality") {
		t.Errorf("expected low-quality last, got %s", sorted[2].PeerID)
	}
}

func TestSortPeersByQuality_Empty(t *testing.T) {
	sorted := SortPeersByQuality(nil)
	if len(sorted) != 0 {
		t.Errorf("expected empty slice, got %d", len(sorted))
	}
}

func TestPeerQualityScore(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		entry    *AddressEntry
		minScore float64
		maxScore float64
	}{
		{
			name: "new peer with no history",
			entry: &AddressEntry{
				ConnAttempts: 0,
				ConnSuccess:  0,
				LastSeen:     now,
			},
			minScore: 0.3, // recency bonus
			maxScore: 0.4,
		},
		{
			name: "perfect recent peer",
			entry: &AddressEntry{
				ConnAttempts: 10,
				ConnSuccess:  10,
				LastSeen:     now,
			},
			minScore: 1.3, // 1.0 rate + 0.3 recency + 0.1 success bonus
			maxScore: 1.5,
		},
		{
			name: "old unreliable peer",
			entry: &AddressEntry{
				ConnAttempts: 10,
				ConnSuccess:  1,
				LastSeen:     now.Add(-48 * time.Hour),
			},
			minScore: 0.1, // 0.1 rate + 0.1 success bonus
			maxScore: 0.3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := peerQualityScore(tt.entry)
			if score < tt.minScore || score > tt.maxScore {
				t.Errorf("score %f outside expected range [%f, %f]", score, tt.minScore, tt.maxScore)
			}
		})
	}
}

func TestPeerExchangeMessage_Serialization(t *testing.T) {
	msg := PeerExchangeMessage{
		RequestID: "req-123",
		Peers: []PeerExchangeAddr{
			{
				NodeID:  "node-abc",
				Addrs:   []string{"/ip4/1.2.3.4/tcp/9000"},
				Region:  "Americas",
				Country: "US",
			},
		},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var decoded PeerExchangeMessage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if decoded.RequestID != msg.RequestID {
		t.Errorf("request ID mismatch: %s vs %s", decoded.RequestID, msg.RequestID)
	}

	if len(decoded.Peers) != 1 {
		t.Fatalf("expected 1 peer, got %d", len(decoded.Peers))
	}

	if decoded.Peers[0].NodeID != "node-abc" {
		t.Errorf("peer NodeID mismatch: %s", decoded.Peers[0].NodeID)
	}

	if decoded.Peers[0].Region != "Americas" {
		t.Errorf("peer Region mismatch: %s", decoded.Peers[0].Region)
	}
}

func TestPeerExchangeProtocol_StartBackground_Cancellation(t *testing.T) {
	router := NewRouter(nil, nil)
	defer router.Close()

	ab := NewAddressBook()
	cfg := &PeerExchangeConfig{
		Interval: 50 * time.Millisecond, // Short interval for testing
	}
	pex := NewPeerExchangeProtocol(router, ab, cfg)

	ctx, cancel := context.WithCancel(context.Background())

	// StartBackground should not block
	pex.StartBackground(ctx)

	// Let it run for a bit
	time.Sleep(100 * time.Millisecond)

	// Cancel should stop the goroutine
	cancel()

	// Give it time to stop
	time.Sleep(100 * time.Millisecond)
	// No panic or deadlock means success
}
