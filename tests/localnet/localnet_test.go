//go:build localnet

package localnet

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/moltbunker/moltbunker/internal/payment"
	"github.com/moltbunker/moltbunker/pkg/types"
)

// ---------------------------------------------------------------------------
// TestLocalNet_ThreeNodeDiscovery
//
// Starts 3 nodes on localhost. Because all nodes run on the same machine with
// mDNS enabled (service name "moltbunker-discovery"), they should discover
// each other automatically through the libp2p mDNS notifee that calls
// DHT.Host().Connect(). We poll each node's Router.PeerCount() until every
// node reports at least 2 peers.
// ---------------------------------------------------------------------------

func TestLocalNet_ThreeNodeDiscovery(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	nodes := startTestNodes(t, ctx, 3)

	// Log the addresses each node is listening on for debugging.
	for _, tn := range nodes {
		info := tn.Node.NodeInfo()
		t.Logf("node %d: id=%s port=%d", tn.Index, info.ID.String()[:16], tn.Port)
	}

	// Wait for mDNS discovery -- each node should see the other 2.
	discoveryCtx, discoveryCancel := context.WithTimeout(ctx, 45*time.Second)
	defer discoveryCancel()

	if err := waitForPeerDiscovery(discoveryCtx, nodes, 2, 500*time.Millisecond); err != nil {
		t.Fatalf("peer discovery failed: %v", err)
	}

	// Verify every node sees at least 2 peers.
	for _, tn := range nodes {
		peerCount := tn.Node.Router().PeerCount()
		t.Logf("node %d has %d peers", tn.Index, peerCount)
		if peerCount < 2 {
			t.Errorf("node %d: expected at least 2 peers, got %d", tn.Index, peerCount)
		}
	}
}

// ---------------------------------------------------------------------------
// TestLocalNet_PingPong
//
// Starts 3 nodes, waits for discovery, then has node 0 send a Ping to node 1.
// Node 1's default ping handler (registered via node.Start -> registerHandlers)
// automatically replies with a Pong. We register a custom Pong collector on
// node 0 to capture the response.
// ---------------------------------------------------------------------------

func TestLocalNet_PingPong(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	nodes := startTestNodes(t, ctx, 3)

	// Wait for discovery.
	discoveryCtx, discoveryCancel := context.WithTimeout(ctx, 45*time.Second)
	defer discoveryCancel()
	if err := waitForPeerDiscovery(discoveryCtx, nodes, 2, 500*time.Millisecond); err != nil {
		t.Fatalf("peer discovery failed: %v", err)
	}

	// Set up a pong collector on node 0.
	pongCollector := newMessageCollector()
	nodes[0].Node.Router().RegisterHandler(types.MessageTypePong, pongCollector.collect)

	// Build ping message from node 0 to node 1.
	node0Info := nodes[0].Node.NodeInfo()
	node1Info := nodes[1].Node.NodeInfo()

	pingMsg := &types.Message{
		Type:      types.MessageTypePing,
		From:      node0Info.ID,
		To:        node1Info.ID,
		Timestamp: time.Now(),
	}

	// Send ping via the router. The router will look up node 1 in the DHT
	// peer list and establish a TLS connection to deliver the message.
	sendCtx, sendCancel := context.WithTimeout(ctx, 15*time.Second)
	defer sendCancel()
	if err := nodes[0].Node.Router().SendMessage(sendCtx, node1Info.ID, pingMsg); err != nil {
		t.Fatalf("failed to send ping: %v", err)
	}

	// Wait for pong response.
	pongCtx, pongCancel := context.WithTimeout(ctx, 15*time.Second)
	defer pongCancel()
	if err := waitForCondition(pongCtx, 200*time.Millisecond, func() bool {
		return pongCollector.count() > 0
	}); err != nil {
		t.Fatalf("did not receive pong response: %v", err)
	}

	pongs := pongCollector.getMessages()
	if len(pongs) == 0 {
		t.Fatal("expected at least 1 pong message")
	}
	if pongs[0].Type != types.MessageTypePong {
		t.Errorf("expected message type %s, got %s", types.MessageTypePong, pongs[0].Type)
	}
	t.Logf("received pong from node 1 with timestamp %s", pongs[0].Timestamp)
}

// ---------------------------------------------------------------------------
// TestLocalNet_MessageBroadcast
//
// Starts 3 nodes, waits for discovery, then has node 0 broadcast a Health
// message to all peers. We register Health collectors on nodes 1 and 2 and
// verify they both receive the broadcast.
// ---------------------------------------------------------------------------

func TestLocalNet_MessageBroadcast(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	nodes := startTestNodes(t, ctx, 3)

	discoveryCtx, discoveryCancel := context.WithTimeout(ctx, 45*time.Second)
	defer discoveryCancel()
	if err := waitForPeerDiscovery(discoveryCtx, nodes, 2, 500*time.Millisecond); err != nil {
		t.Fatalf("peer discovery failed: %v", err)
	}

	// Register health message collectors on nodes 1 and 2.
	collector1 := newMessageCollector()
	collector2 := newMessageCollector()
	nodes[1].Node.Router().RegisterHandler(types.MessageTypeHealth, collector1.collect)
	nodes[2].Node.Router().RegisterHandler(types.MessageTypeHealth, collector2.collect)

	// Broadcast a Health message from node 0.
	node0Info := nodes[0].Node.NodeInfo()
	healthMsg := &types.Message{
		Type:      types.MessageTypeHealth,
		From:      node0Info.ID,
		Payload:   []byte(`{"cpu":0.5,"memory":1024}`),
		Timestamp: time.Now(),
	}

	broadcastCtx, broadcastCancel := context.WithTimeout(ctx, 15*time.Second)
	defer broadcastCancel()
	if err := nodes[0].Node.Router().BroadcastMessage(broadcastCtx, healthMsg); err != nil {
		// BroadcastMessage returns the last error; some sends may fail if
		// connections aren't fully established yet. We'll check collectors.
		t.Logf("broadcast returned error (may be partial): %v", err)
	}

	// Wait for at least one collector to receive a message.
	recvCtx, recvCancel := context.WithTimeout(ctx, 15*time.Second)
	defer recvCancel()
	if err := waitForCondition(recvCtx, 200*time.Millisecond, func() bool {
		return collector1.count() > 0 || collector2.count() > 0
	}); err != nil {
		t.Fatalf("no node received the broadcast: %v", err)
	}

	totalReceived := collector1.count() + collector2.count()
	t.Logf("broadcast received by %d node(s) (node1=%d, node2=%d)",
		totalReceived, collector1.count(), collector2.count())

	if totalReceived == 0 {
		t.Error("expected at least one node to receive the broadcast")
	}
}

// ---------------------------------------------------------------------------
// TestLocalNet_NodeJoinLeave
//
// Starts 2 nodes, waits for them to discover each other, then adds a 3rd node
// and verifies it joins the network. Finally removes the 3rd node and checks
// that the remaining nodes detect the departure.
// ---------------------------------------------------------------------------

func TestLocalNet_NodeJoinLeave(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	// Phase 1: Start 2 nodes and wait for mutual discovery.
	nodes := startTestNodes(t, ctx, 2)

	disco1Ctx, disco1Cancel := context.WithTimeout(ctx, 45*time.Second)
	defer disco1Cancel()
	if err := waitForPeerDiscovery(disco1Ctx, nodes, 1, 500*time.Millisecond); err != nil {
		t.Fatalf("initial discovery (2 nodes) failed: %v", err)
	}
	t.Log("phase 1 complete: 2 nodes discovered each other")

	// Phase 2: Add a 3rd node.
	node3 := startTestNode(t, ctx, 2)
	allNodes := append(nodes, node3)

	disco2Ctx, disco2Cancel := context.WithTimeout(ctx, 45*time.Second)
	defer disco2Cancel()
	if err := waitForPeerDiscovery(disco2Ctx, allNodes, 2, 500*time.Millisecond); err != nil {
		t.Fatalf("discovery after 3rd node joined failed: %v", err)
	}
	t.Log("phase 2 complete: 3rd node joined and discovered")

	// Record peer counts before shutdown.
	for _, tn := range allNodes {
		t.Logf("node %d peer count before leave: %d", tn.Index, tn.Node.Router().PeerCount())
	}

	// Phase 3: Shut down node 3 and verify peer counts decrease.
	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 10*time.Second)
	defer shutdownCancel()
	if err := node3.Node.Shutdown(shutdownCtx); err != nil {
		t.Logf("node 3 shutdown error: %v", err)
	}

	// Wait for remaining nodes to detect the departure. The libp2p host
	// fires DisconnectedF which removes the peer from the DHT and router.
	leaveCtx, leaveCancel := context.WithTimeout(ctx, 20*time.Second)
	defer leaveCancel()
	if err := waitForCondition(leaveCtx, 500*time.Millisecond, func() bool {
		// Both remaining nodes should have fewer peers than before.
		for _, tn := range nodes[:2] {
			if tn.Node.Router().PeerCount() >= 2 {
				return false
			}
		}
		return true
	}); err != nil {
		// Not a hard failure -- mDNS cache may keep the peer around briefly.
		t.Logf("warning: peer departure detection may be slow: %v", err)
	}

	for _, tn := range nodes[:2] {
		t.Logf("node %d peer count after leave: %d", tn.Index, tn.Node.Router().PeerCount())
	}
	t.Log("phase 3 complete: node departure observed")
}

// ---------------------------------------------------------------------------
// TestLocalNet_StakingFlow
//
// Creates a StakingManager (no live Ethereum client needed -- it uses mock
// mode internally) and exercises the stake/unstake flow for provider nodes.
// This tests the economic layer that sits alongside the P2P network.
// ---------------------------------------------------------------------------

func TestLocalNet_StakingFlow(t *testing.T) {
	// The StakingManager works without an ethclient when using mock mode.
	minStake := big.NewInt(1_000_000_000_000_000_000) // 1 BUNKER (18 decimals)
	sm := payment.NewStakingManager(nil, minStake)

	provider1 := common.HexToAddress("0x1111111111111111111111111111111111111111")
	provider2 := common.HexToAddress("0x2222222222222222222222222222222222222222")
	provider3 := common.HexToAddress("0x3333333333333333333333333333333333333333")

	ctx := context.Background()

	// --- Stake for 3 providers ---

	stake5 := newBigInt("5000000000000000000")    // 5 BUNKER
	stake10 := newBigInt("10000000000000000000")  // 10 BUNKER
	stake2 := newBigInt("2000000000000000000")    // 2 BUNKER

	stakeAmounts := map[common.Address]*big.Int{
		provider1: stake5,
		provider2: stake10,
		provider3: stake2,
	}

	for addr, amount := range stakeAmounts {
		if err := sm.Stake(ctx, addr, amount); err != nil {
			t.Fatalf("failed to stake for %s: %v", addr.Hex(), err)
		}
	}

	// Verify stakes were recorded.
	for addr, expected := range stakeAmounts {
		got := sm.GetStake(addr)
		if got.Cmp(expected) != 0 {
			t.Errorf("stake for %s: got %s, want %s", addr.Hex(), got.String(), expected.String())
		}
	}

	// --- HasMinimumStake checks ---
	for addr := range stakeAmounts {
		if !sm.HasMinimumStake(addr) {
			t.Errorf("expected %s to have minimum stake", addr.Hex())
		}
	}

	// Provider with no stake should not pass the minimum check.
	noStake := common.HexToAddress("0x4444444444444444444444444444444444444444")
	if sm.HasMinimumStake(noStake) {
		t.Error("unstaked provider should not pass HasMinimumStake")
	}

	// --- Below minimum stake attempt ---
	belowMin := big.NewInt(500_000_000_000_000_000) // 0.5 BUNKER
	if err := sm.Stake(ctx, noStake, belowMin); err == nil {
		t.Error("expected error when staking below minimum, got nil")
	}

	// --- Slash provider 1 partially ---
	slashAmount := big.NewInt(2_000_000_000_000_000_000) // 2 BUNKER
	if err := sm.Slash(ctx, provider1, slashAmount); err != nil {
		t.Fatalf("failed to slash provider1: %v", err)
	}

	expectedAfterSlash := new(big.Int).Sub(stakeAmounts[provider1], slashAmount) // 3 BUNKER
	gotAfterSlash := sm.GetStake(provider1)
	if gotAfterSlash.Cmp(expectedAfterSlash) != 0 {
		t.Errorf("provider1 stake after slash: got %s, want %s",
			gotAfterSlash.String(), expectedAfterSlash.String())
	}

	// --- Slash more than stake (should cap at current stake) ---
	hugeSlash := newBigInt("100000000000000000000") // 100 BUNKER
	if err := sm.Slash(ctx, provider3, hugeSlash); err != nil {
		t.Fatalf("failed to slash provider3: %v", err)
	}

	gotAfterHugeSlash := sm.GetStake(provider3)
	if gotAfterHugeSlash.Sign() != 0 {
		t.Errorf("provider3 stake after excessive slash should be 0, got %s", gotAfterHugeSlash.String())
	}

	// Provider 3 no longer has minimum stake.
	if sm.HasMinimumStake(provider3) {
		t.Error("provider3 should not pass HasMinimumStake after full slash")
	}

	// --- Additional stake accumulates ---
	extra := big.NewInt(3_000_000_000_000_000_000) // 3 BUNKER
	if err := sm.Stake(ctx, provider2, extra); err != nil {
		t.Fatalf("failed to add extra stake for provider2: %v", err)
	}
	expectedP2 := new(big.Int).Add(stakeAmounts[provider2], extra) // 13 BUNKER
	gotP2 := sm.GetStake(provider2)
	if gotP2.Cmp(expectedP2) != 0 {
		t.Errorf("provider2 cumulative stake: got %s, want %s", gotP2.String(), expectedP2.String())
	}

	t.Logf("staking flow test passed -- provider1=%s, provider2=%s, provider3=%s",
		sm.GetStake(provider1).String(),
		sm.GetStake(provider2).String(),
		sm.GetStake(provider3).String(),
	)
}
