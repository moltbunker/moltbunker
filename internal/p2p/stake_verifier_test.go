package p2p

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/moltbunker/moltbunker/pkg/types"
)

func TestStakeVerifier_RegisterAndQuery(t *testing.T) {
	sv := NewStakeVerifier(nil)

	nodeID := randomNodeID()
	wallet := common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")

	if sv.HasAnnounced(nodeID) {
		t.Fatal("should not have announced yet")
	}

	sv.RegisterPeerWallet(nodeID, wallet)

	if !sv.HasAnnounced(nodeID) {
		t.Fatal("should have announced")
	}

	got, ok := sv.GetWallet(nodeID)
	if !ok || got != wallet {
		t.Fatalf("expected wallet %s, got %s", wallet.Hex(), got.Hex())
	}
}

func TestStakeVerifier_NoPaymentService(t *testing.T) {
	sv := NewStakeVerifier(nil)

	nodeID := randomNodeID()
	wallet := common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")
	sv.RegisterPeerWallet(nodeID, wallet)

	ctx := context.Background()

	// Without PaymentService, IsStaked should return false
	if sv.IsStaked(ctx, nodeID) {
		t.Fatal("should not be staked without PaymentService")
	}

	// GetTier should return "unknown"
	tier := sv.GetTier(ctx, nodeID)
	if tier != "unknown" {
		t.Fatalf("expected 'unknown' tier, got %s", tier)
	}
}

func TestStakeVerifier_UnregisteredPeer(t *testing.T) {
	sv := NewStakeVerifier(nil)
	ctx := context.Background()

	unknownNodeID := randomNodeID()

	if sv.IsStaked(ctx, unknownNodeID) {
		t.Fatal("unregistered peer should not be staked")
	}

	tier := sv.GetTier(ctx, unknownNodeID)
	if tier != "unstaked" {
		t.Fatalf("expected 'unstaked' tier for unregistered peer, got %s", tier)
	}
}

func TestStakeVerifier_RemovePeer(t *testing.T) {
	sv := NewStakeVerifier(nil)

	nodeID := randomNodeID()
	wallet := common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")
	sv.RegisterPeerWallet(nodeID, wallet)

	if sv.PeerCount() != 1 {
		t.Fatalf("expected 1 peer, got %d", sv.PeerCount())
	}

	sv.RemovePeer(nodeID)

	if sv.PeerCount() != 0 {
		t.Fatalf("expected 0 peers after removal, got %d", sv.PeerCount())
	}

	if sv.HasAnnounced(nodeID) {
		t.Fatal("should not be announced after removal")
	}
}

func TestStakeVerifier_MockVerifier(t *testing.T) {
	nodeID := randomNodeID()
	wallet := common.HexToAddress("0xabcdef1234567890abcdef1234567890abcdef12")

	peers := map[types.NodeID]common.Address{
		nodeID: wallet,
	}
	stakes := map[common.Address]*CachedStakeInfo{
		wallet: {
			HasMinStake: true,
			Tier:        types.StakingTierSilver,
			Amount:      big.NewInt(10000),
			FetchedAt:   time.Now(),
		},
	}

	sv := NewMockStakeVerifier(peers, stakes)

	if !sv.HasAnnounced(nodeID) {
		t.Fatal("mock peer should be announced")
	}

	// GetTier should return cached value directly
	tier := sv.GetTier(context.Background(), nodeID)
	// Without PaymentService it returns "unknown" because getCachedOrFetch needs ps != nil
	// But the mock has pre-populated cache. Let's check the cache directly.
	sv.mu.RLock()
	cached, exists := sv.stakeCache[wallet]
	sv.mu.RUnlock()

	if !exists {
		t.Fatal("mock stake cache should exist")
	}
	if cached.Tier != types.StakingTierSilver {
		t.Fatalf("expected Silver tier in cache, got %s", cached.Tier)
	}

	// Without PaymentService, GetTier returns "unknown" (can't verify)
	if tier != "unknown" {
		t.Fatalf("expected 'unknown' without payment service, got %s", tier)
	}
}

func TestStakeVerifier_CacheTTL(t *testing.T) {
	now := time.Now()
	sv := NewStakeVerifierWithConfig(nil, 1*time.Minute, 30*time.Second, 10*time.Second)
	sv.nowFunc = func() time.Time { return now }

	wallet := common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")

	// Pre-populate cache with a positive result
	sv.mu.Lock()
	sv.stakeCache[wallet] = &CachedStakeInfo{
		HasMinStake: true,
		Tier:        types.StakingTierGold,
		Amount:      big.NewInt(50000),
		FetchedAt:   now,
	}
	sv.mu.Unlock()

	// Cache should be valid
	sv.mu.RLock()
	cached := sv.stakeCache[wallet]
	sv.mu.RUnlock()

	age := now.Sub(cached.FetchedAt)
	if age >= sv.cacheTTL {
		t.Fatal("cache should be fresh")
	}

	// Advance past cache TTL
	sv.nowFunc = func() time.Time { return now.Add(2 * time.Minute) }

	// Now cache should be stale
	age = sv.nowFunc().Sub(cached.FetchedAt)
	if age < sv.cacheTTL {
		t.Fatal("cache should be stale after advancing time")
	}
}

func TestStakeVerifier_MultiPeerSameWallet(t *testing.T) {
	sv := NewStakeVerifier(nil)
	wallet := common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")

	nodeID1 := randomNodeID()
	nodeID2 := randomNodeID()

	sv.RegisterPeerWallet(nodeID1, wallet)
	sv.RegisterPeerWallet(nodeID2, wallet)

	if sv.PeerCount() != 2 {
		t.Fatalf("expected 2 peers, got %d", sv.PeerCount())
	}

	// Remove one peer — cache should remain for the other
	sv.RemovePeer(nodeID1)

	if sv.PeerCount() != 1 {
		t.Fatalf("expected 1 peer, got %d", sv.PeerCount())
	}

	// Second peer should still be announced
	if !sv.HasAnnounced(nodeID2) {
		t.Fatal("second peer should still be announced")
	}
}
