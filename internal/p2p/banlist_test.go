package p2p

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/moltbunker/moltbunker/pkg/types"
)

// testNodeID creates a deterministic NodeID from a byte value for testing.
func testNodeID(b byte) types.NodeID {
	var id types.NodeID
	for i := range id {
		id[i] = b
	}
	return id
}

func TestBanList_BanAndIsBanned(t *testing.T) {
	bl := NewBanList()

	peer := testNodeID(0x01)

	if bl.IsBanned(peer) {
		t.Error("expected peer to not be banned initially")
	}

	bl.Ban(peer, "misbehavior", 0)

	if !bl.IsBanned(peer) {
		t.Error("expected peer to be banned after Ban()")
	}
}

func TestBanList_Unban(t *testing.T) {
	bl := NewBanList()

	peer := testNodeID(0x02)
	bl.Ban(peer, "test", 0)

	if !bl.IsBanned(peer) {
		t.Fatal("expected peer to be banned")
	}

	bl.Unban(peer)

	if bl.IsBanned(peer) {
		t.Error("expected peer to not be banned after Unban()")
	}
}

func TestBanList_UnbanNonExistent(t *testing.T) {
	bl := NewBanList()
	// Unbanning a peer that was never banned should not panic
	bl.Unban(testNodeID(0xFF))
}

func TestBanList_TemporaryBanExpires(t *testing.T) {
	bl := NewBanList()
	peer := testNodeID(0x03)

	// Ban for a very short duration
	bl.Ban(peer, "short ban", 50*time.Millisecond)

	if !bl.IsBanned(peer) {
		t.Fatal("expected peer to be banned immediately after ban")
	}

	// Wait for ban to expire
	time.Sleep(60 * time.Millisecond)

	if bl.IsBanned(peer) {
		t.Error("expected ban to have expired")
	}
}

func TestBanList_PermanentBanDoesNotExpire(t *testing.T) {
	bl := NewBanList()
	peer := testNodeID(0x04)

	bl.Ban(peer, "permanent", 0)

	// Permanent bans should never expire
	if !bl.IsBanned(peer) {
		t.Error("expected permanent ban to remain active")
	}
}

func TestBanList_CleanExpired(t *testing.T) {
	bl := NewBanList()

	peer1 := testNodeID(0x10)
	peer2 := testNodeID(0x11)
	peer3 := testNodeID(0x12)

	bl.Ban(peer1, "short", 50*time.Millisecond)
	bl.Ban(peer2, "permanent", 0)
	bl.Ban(peer3, "also short", 50*time.Millisecond)

	// Wait for temporary bans to expire
	time.Sleep(60 * time.Millisecond)

	bl.CleanExpired()

	// peer1 and peer3 should be cleaned, peer2 (permanent) should remain
	if bl.IsBanned(peer1) {
		t.Error("expected peer1 ban to be cleaned")
	}
	if !bl.IsBanned(peer2) {
		t.Error("expected peer2 permanent ban to remain")
	}
	if bl.IsBanned(peer3) {
		t.Error("expected peer3 ban to be cleaned")
	}

	if bl.Len() != 1 {
		t.Errorf("expected 1 active ban after cleanup, got %d", bl.Len())
	}
}

func TestBanList_List(t *testing.T) {
	bl := NewBanList()

	peer1 := testNodeID(0x20)
	peer2 := testNodeID(0x21)

	bl.Ban(peer1, "reason1", 0)
	bl.Ban(peer2, "reason2", time.Hour)

	list := bl.List()
	if len(list) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(list))
	}

	// Verify both peers are in the list
	found := map[types.NodeID]bool{}
	for _, entry := range list {
		found[entry.PeerID] = true
	}
	if !found[peer1] || !found[peer2] {
		t.Error("expected both peers in the list")
	}
}

func TestBanList_ListExcludesExpired(t *testing.T) {
	bl := NewBanList()

	peer1 := testNodeID(0x30)
	peer2 := testNodeID(0x31)

	bl.Ban(peer1, "short", 50*time.Millisecond)
	bl.Ban(peer2, "permanent", 0)

	time.Sleep(60 * time.Millisecond)

	list := bl.List()
	if len(list) != 1 {
		t.Fatalf("expected 1 active entry, got %d", len(list))
	}
	if list[0].PeerID != peer2 {
		t.Errorf("expected peer2 in list, got %s", list[0].PeerID.String()[:16])
	}
}

func TestBanList_SaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "bans.json")

	bl := NewBanList()

	peer1 := testNodeID(0x40)
	peer2 := testNodeID(0x41)

	bl.Ban(peer1, "reason1", 0)
	bl.Ban(peer2, "reason2", 24*time.Hour)

	// Save
	if err := bl.Save(path); err != nil {
		t.Fatalf("Save error: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected file to exist: %v", err)
	}

	// Verify no .tmp file remains
	tmpPath := path + ".tmp"
	if _, err := os.Stat(tmpPath); !os.IsNotExist(err) {
		t.Error("expected .tmp file to not exist after save")
	}

	// Load into a new BanList
	bl2 := NewBanList()
	if err := bl2.Load(path); err != nil {
		t.Fatalf("Load error: %v", err)
	}

	if !bl2.IsBanned(peer1) {
		t.Error("expected peer1 to be banned after load")
	}
	if !bl2.IsBanned(peer2) {
		t.Error("expected peer2 to be banned after load")
	}

	// Verify entry details
	list := bl2.List()
	if len(list) != 2 {
		t.Fatalf("expected 2 entries after load, got %d", len(list))
	}
}

func TestBanList_SaveCreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "subdir", "nested", "bans.json")

	bl := NewBanList()
	bl.Ban(testNodeID(0x50), "test", 0)

	if err := bl.Save(path); err != nil {
		t.Fatalf("Save error: %v", err)
	}

	// Verify the file was created
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected file to exist: %v", err)
	}
}

func TestBanList_LoadNonExistentFile(t *testing.T) {
	bl := NewBanList()
	err := bl.Load("/tmp/nonexistent-banlist-test-12345.json")
	if err != nil {
		t.Errorf("expected no error for non-existent file, got: %v", err)
	}
}

func TestBanList_LoadInvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "bans.json")

	if err := os.WriteFile(path, []byte("not valid json{"), 0600); err != nil {
		t.Fatalf("write error: %v", err)
	}

	bl := NewBanList()
	err := bl.Load(path)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestBanList_LoadEmptyArray(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "bans.json")

	if err := os.WriteFile(path, []byte("[]"), 0600); err != nil {
		t.Fatalf("write error: %v", err)
	}

	bl := NewBanList()
	if err := bl.Load(path); err != nil {
		t.Fatalf("Load error: %v", err)
	}

	if bl.Len() != 0 {
		t.Errorf("expected 0 entries, got %d", bl.Len())
	}
}

func TestBanList_RebanUpdateEntry(t *testing.T) {
	bl := NewBanList()
	peer := testNodeID(0x60)

	bl.Ban(peer, "first offense", time.Hour)
	bl.Ban(peer, "second offense", 0)

	if !bl.IsBanned(peer) {
		t.Error("expected peer to be banned")
	}

	list := bl.List()
	if len(list) != 1 {
		t.Fatalf("expected 1 entry (re-ban should overwrite), got %d", len(list))
	}
	if list[0].Reason != "second offense" {
		t.Errorf("expected reason to be updated, got %s", list[0].Reason)
	}
	if !list[0].IsPermanent() {
		t.Error("expected re-ban to be permanent")
	}
}

func TestBanEntry_JSON_RoundTrip(t *testing.T) {
	now := time.Now().Truncate(time.Millisecond)
	entry := BanEntry{
		PeerID:    testNodeID(0x70),
		Reason:    "test reason",
		BannedAt:  now,
		ExpiresAt: now.Add(time.Hour),
	}

	data, err := json.Marshal(entry)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var decoded BanEntry
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if decoded.PeerID != entry.PeerID {
		t.Errorf("PeerID mismatch")
	}
	if decoded.Reason != entry.Reason {
		t.Errorf("Reason mismatch: %s != %s", decoded.Reason, entry.Reason)
	}
	if !decoded.BannedAt.Equal(entry.BannedAt) {
		t.Errorf("BannedAt mismatch")
	}
	if !decoded.ExpiresAt.Equal(entry.ExpiresAt) {
		t.Errorf("ExpiresAt mismatch")
	}
}

func TestRouter_HandleMessage_BannedPeerRejected(t *testing.T) {
	bl := NewBanList()
	r := NewRouter(nil, nil)
	r.SetBanList(bl)

	peer := testNodeID(0x80)
	from := &types.Node{ID: peer}
	bl.Ban(peer, "test ban", 0)

	r.RegisterHandler(types.MessageTypePing, func(ctx context.Context, msg *types.Message, from *types.Node) error {
		t.Error("handler should not be called for banned peer")
		return nil
	})

	msg := &types.Message{
		Type:    types.MessageTypePing,
		Version: types.ProtocolVersion,
	}

	err := r.HandleMessage(context.Background(), msg, from)
	if err == nil {
		t.Error("expected error when handling message from banned peer")
	}
}

func TestRouter_HandleMessage_UnbannedPeerAllowed(t *testing.T) {
	bl := NewBanList()
	r := NewRouter(nil, nil)
	r.SetBanList(bl)

	peer := testNodeID(0x81)
	from := &types.Node{ID: peer}

	called := false
	r.RegisterHandler(types.MessageTypePing, func(ctx context.Context, msg *types.Message, from *types.Node) error {
		called = true
		return nil
	})

	msg := &types.Message{
		Type:    types.MessageTypePing,
		Version: types.ProtocolVersion,
	}

	err := r.HandleMessage(context.Background(), msg, from)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected handler to be called for non-banned peer")
	}
}

func TestRouter_ViolationCounter_AutoBan(t *testing.T) {
	bl := NewBanList()
	r := NewRouter(nil, nil)
	r.SetBanList(bl)

	peer := testNodeID(0x90)

	// Record violations up to the threshold
	for i := 0; i < violationThreshold; i++ {
		r.recordViolation(peer)
	}

	if !bl.IsBanned(peer) {
		t.Error("expected peer to be auto-banned after reaching violation threshold")
	}

	// Verify the ban entry details
	list := bl.List()
	found := false
	for _, entry := range list {
		if entry.PeerID == peer {
			found = true
			if entry.IsPermanent() {
				t.Error("expected auto-ban to be temporary, not permanent")
			}
			if entry.Reason == "" {
				t.Error("expected auto-ban to have a reason")
			}
		}
	}
	if !found {
		t.Error("expected to find auto-banned peer in ban list")
	}
}

func TestRouter_ViolationCounter_BelowThreshold(t *testing.T) {
	bl := NewBanList()
	r := NewRouter(nil, nil)
	r.SetBanList(bl)

	peer := testNodeID(0x91)

	// Record fewer violations than the threshold
	for i := 0; i < violationThreshold-1; i++ {
		r.recordViolation(peer)
	}

	if bl.IsBanned(peer) {
		t.Error("expected peer to not be banned below violation threshold")
	}
}

func TestRouter_ViolationCounter_WindowReset(t *testing.T) {
	bl := NewBanList()
	r := NewRouter(nil, nil)
	r.SetBanList(bl)

	peer := testNodeID(0x92)

	// Record a violation
	r.recordViolation(peer)

	// Manually expire the window by modifying the stored entry
	val, ok := r.violationCounters.Load(peer)
	if !ok {
		t.Fatal("expected violation entry to exist")
	}
	entry := val.(*ViolationEntry)
	entry.FirstViolation = time.Now().Add(-violationWindow - time.Second)
	r.violationCounters.Store(peer, entry)

	// Next violation should reset the counter
	r.recordViolation(peer)

	// Verify counter was reset (count should be 1, not 2)
	val, ok = r.violationCounters.Load(peer)
	if !ok {
		t.Fatal("expected violation entry to exist after reset")
	}
	entry = val.(*ViolationEntry)
	if entry.Count != 1 {
		t.Errorf("expected count to be 1 after window reset, got %d", entry.Count)
	}
}

func TestRouter_ViolationCounter_NoBanListNoPanic(t *testing.T) {
	// Verify that recording violations without a ban list does not panic
	r := NewRouter(nil, nil)
	peer := testNodeID(0x93)

	for i := 0; i < violationThreshold+1; i++ {
		r.recordViolation(peer) // Should not panic even without a ban list
	}
}
