package p2p

import (
	crand "crypto/rand"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
)

// randTestPeerID generates a random valid libp2p peer.ID for testing
// JSON round-trip serialization (peer.ID requires a valid multihash).
func randTestPeerID(t *testing.T) peer.ID {
	t.Helper()
	_, pub, err := crypto.GenerateEd25519Key(crand.Reader)
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}
	pid, err := peer.IDFromPublicKey(pub)
	if err != nil {
		t.Fatalf("failed to create peer ID: %v", err)
	}
	return pid
}

// testMultiaddrs parses multiaddr strings for testing, skipping invalid ones.
// testPeerID is defined in latency_test.go (same package).
func testMultiaddrs(t *testing.T, strs ...string) []ma.Multiaddr {
	t.Helper()
	var addrs []ma.Multiaddr
	for _, s := range strs {
		maddr, err := ma.NewMultiaddr(s)
		if err != nil {
			t.Fatalf("invalid test multiaddr %q: %v", s, err)
		}
		addrs = append(addrs, maddr)
	}
	return addrs
}

func TestAddressBook_NewAddressBook(t *testing.T) {
	ab := NewAddressBook()
	if ab == nil {
		t.Fatal("NewAddressBook returned nil")
	}
	if ab.Len() != 0 {
		t.Errorf("expected empty address book, got %d entries", ab.Len())
	}
}

func TestAddressBook_AddPeer(t *testing.T) {
	ab := NewAddressBook()

	pid := testPeerID("peer1")
	addrs := testMultiaddrs(t, "/ip4/192.168.1.1/tcp/9000")

	ab.AddPeer(pid, addrs, "dht")

	if ab.Len() != 1 {
		t.Fatalf("expected 1 entry, got %d", ab.Len())
	}

	peers := ab.GetPeers()
	if len(peers) != 1 {
		t.Fatalf("expected 1 peer, got %d", len(peers))
	}

	entry := peers[0]
	if entry.PeerID != pid {
		t.Errorf("expected peer ID %s, got %s", pid, entry.PeerID)
	}
	if len(entry.Addrs) != 1 {
		t.Errorf("expected 1 address, got %d", len(entry.Addrs))
	}
	if entry.Addrs[0] != "/ip4/192.168.1.1/tcp/9000" {
		t.Errorf("unexpected address: %s", entry.Addrs[0])
	}
	if entry.Source != "dht" {
		t.Errorf("expected source 'dht', got %s", entry.Source)
	}
}

func TestAddressBook_AddPeer_Update(t *testing.T) {
	ab := NewAddressBook()

	pid := testPeerID("peer1")
	addrs1 := testMultiaddrs(t, "/ip4/192.168.1.1/tcp/9000")
	addrs2 := testMultiaddrs(t, "/ip4/10.0.0.1/tcp/9000")

	ab.AddPeer(pid, addrs1, "dht")
	ab.AddPeer(pid, addrs2, "mdns")

	if ab.Len() != 1 {
		t.Fatalf("expected 1 entry after update, got %d", ab.Len())
	}

	peers := ab.GetPeers()
	entry := peers[0]

	// Addresses should be merged
	if len(entry.Addrs) != 2 {
		t.Fatalf("expected 2 addresses after merge, got %d", len(entry.Addrs))
	}

	// Source should be updated to the latest
	if entry.Source != "mdns" {
		t.Errorf("expected source to be updated to 'mdns', got %s", entry.Source)
	}
}

func TestAddressBook_AddPeer_DeduplicateAddrs(t *testing.T) {
	ab := NewAddressBook()

	pid := testPeerID("peer1")
	addrs := testMultiaddrs(t, "/ip4/192.168.1.1/tcp/9000")

	ab.AddPeer(pid, addrs, "dht")
	ab.AddPeer(pid, addrs, "dht") // Same address again

	peers := ab.GetPeers()
	if len(peers[0].Addrs) != 1 {
		t.Errorf("expected 1 address after dedup, got %d", len(peers[0].Addrs))
	}
}

func TestAddressBook_UpdateLastSeen(t *testing.T) {
	ab := NewAddressBook()

	pid := testPeerID("peer1")
	addrs := testMultiaddrs(t, "/ip4/192.168.1.1/tcp/9000")

	ab.AddPeer(pid, addrs, "dht")

	peers := ab.GetPeers()
	originalLastSeen := peers[0].LastSeen

	time.Sleep(10 * time.Millisecond)
	ab.UpdateLastSeen(pid)

	peers = ab.GetPeers()
	if !peers[0].LastSeen.After(originalLastSeen) {
		t.Error("expected LastSeen to be updated")
	}
}

func TestAddressBook_UpdateLastSeen_NonExistent(t *testing.T) {
	ab := NewAddressBook()

	// Should not panic for non-existent peer
	ab.UpdateLastSeen(testPeerID("ghost"))
}

func TestAddressBook_RecordConnectionAttempt(t *testing.T) {
	ab := NewAddressBook()

	pid := testPeerID("peer1")
	addrs := testMultiaddrs(t, "/ip4/192.168.1.1/tcp/9000")

	ab.AddPeer(pid, addrs, "dht")

	ab.RecordConnectionAttempt(pid, true)
	ab.RecordConnectionAttempt(pid, true)
	ab.RecordConnectionAttempt(pid, false)

	peers := ab.GetPeers()
	entry := peers[0]

	if entry.ConnAttempts != 3 {
		t.Errorf("expected 3 connection attempts, got %d", entry.ConnAttempts)
	}
	if entry.ConnSuccess != 2 {
		t.Errorf("expected 2 successful connections, got %d", entry.ConnSuccess)
	}
}

func TestAddressBook_RecordConnectionAttempt_NonExistent(t *testing.T) {
	ab := NewAddressBook()

	// Should not panic for non-existent peer
	ab.RecordConnectionAttempt(testPeerID("ghost"), true)
}

func TestAddressBook_RemovePeer(t *testing.T) {
	ab := NewAddressBook()

	pid := testPeerID("peer1")
	addrs := testMultiaddrs(t, "/ip4/192.168.1.1/tcp/9000")

	ab.AddPeer(pid, addrs, "dht")

	if ab.Len() != 1 {
		t.Fatalf("expected 1 entry, got %d", ab.Len())
	}

	ab.RemovePeer(pid)

	if ab.Len() != 0 {
		t.Errorf("expected 0 entries after remove, got %d", ab.Len())
	}
}

func TestAddressBook_RemovePeer_NonExistent(t *testing.T) {
	ab := NewAddressBook()

	// Should not panic for non-existent peer
	ab.RemovePeer(testPeerID("ghost"))
}

func TestAddressBook_GetBestPeers_SortBySuccessRate(t *testing.T) {
	ab := NewAddressBook()

	// peer1: 50% success rate (1/2)
	pid1 := testPeerID("peer1")
	ab.AddPeer(pid1, testMultiaddrs(t, "/ip4/1.1.1.1/tcp/9000"), "dht")
	ab.RecordConnectionAttempt(pid1, true)
	ab.RecordConnectionAttempt(pid1, false)

	// peer2: 100% success rate (3/3)
	pid2 := testPeerID("peer2")
	ab.AddPeer(pid2, testMultiaddrs(t, "/ip4/2.2.2.2/tcp/9000"), "dht")
	ab.RecordConnectionAttempt(pid2, true)
	ab.RecordConnectionAttempt(pid2, true)
	ab.RecordConnectionAttempt(pid2, true)

	// peer3: 0% success rate (0/2)
	pid3 := testPeerID("peer3")
	ab.AddPeer(pid3, testMultiaddrs(t, "/ip4/3.3.3.3/tcp/9000"), "dht")
	ab.RecordConnectionAttempt(pid3, false)
	ab.RecordConnectionAttempt(pid3, false)

	best := ab.GetBestPeers(3)

	if len(best) != 3 {
		t.Fatalf("expected 3 peers, got %d", len(best))
	}

	// peer2 (100%) should be first
	if best[0].PeerID != pid2 {
		t.Errorf("expected peer2 first (100%% success), got %s", best[0].PeerID)
	}
	// peer1 (50%) should be second
	if best[1].PeerID != pid1 {
		t.Errorf("expected peer1 second (50%% success), got %s", best[1].PeerID)
	}
	// peer3 (0%) should be last
	if best[2].PeerID != pid3 {
		t.Errorf("expected peer3 last (0%% success), got %s", best[2].PeerID)
	}
}

func TestAddressBook_GetBestPeers_TiebreakByLastSeen(t *testing.T) {
	ab := NewAddressBook()

	// Both peers have 0 attempts (0% rate), tiebreak by LastSeen
	pid1 := testPeerID("peer1")
	ab.AddPeer(pid1, testMultiaddrs(t, "/ip4/1.1.1.1/tcp/9000"), "dht")

	time.Sleep(10 * time.Millisecond)

	pid2 := testPeerID("peer2")
	ab.AddPeer(pid2, testMultiaddrs(t, "/ip4/2.2.2.2/tcp/9000"), "dht")

	best := ab.GetBestPeers(2)

	if len(best) != 2 {
		t.Fatalf("expected 2 peers, got %d", len(best))
	}

	// peer2 was added later (more recent LastSeen), should come first
	if best[0].PeerID != pid2 {
		t.Errorf("expected peer2 first (more recent), got %s", best[0].PeerID)
	}
	if best[1].PeerID != pid1 {
		t.Errorf("expected peer1 second (less recent), got %s", best[1].PeerID)
	}
}

func TestAddressBook_GetBestPeers_LimitN(t *testing.T) {
	ab := NewAddressBook()

	for i := 0; i < 5; i++ {
		pid := testPeerID(string(rune('a' + i)))
		ab.AddPeer(pid, testMultiaddrs(t, "/ip4/1.1.1.1/tcp/9000"), "dht")
	}

	best := ab.GetBestPeers(3)
	if len(best) != 3 {
		t.Errorf("expected 3 peers, got %d", len(best))
	}
}

func TestAddressBook_GetBestPeers_NGreaterThanTotal(t *testing.T) {
	ab := NewAddressBook()

	pid := testPeerID("peer1")
	ab.AddPeer(pid, testMultiaddrs(t, "/ip4/1.1.1.1/tcp/9000"), "dht")

	best := ab.GetBestPeers(10)
	if len(best) != 1 {
		t.Errorf("expected 1 peer (total available), got %d", len(best))
	}
}

func TestAddressBook_CleanStale(t *testing.T) {
	ab := NewAddressBook()

	// Add two peers
	pid1 := testPeerID("peer1")
	pid2 := testPeerID("peer2")

	ab.AddPeer(pid1, testMultiaddrs(t, "/ip4/1.1.1.1/tcp/9000"), "dht")
	ab.AddPeer(pid2, testMultiaddrs(t, "/ip4/2.2.2.2/tcp/9000"), "dht")

	// Manually backdate peer1's LastSeen
	ab.mu.Lock()
	ab.entries[pid1].LastSeen = time.Now().Add(-2 * time.Hour)
	ab.mu.Unlock()

	// Clean entries older than 1 hour
	ab.CleanStale(1 * time.Hour)

	if ab.Len() != 1 {
		t.Fatalf("expected 1 entry after cleanup, got %d", ab.Len())
	}

	peers := ab.GetPeers()
	if peers[0].PeerID != pid2 {
		t.Errorf("expected peer2 to remain, got %s", peers[0].PeerID)
	}
}

func TestAddressBook_CleanStale_AllFresh(t *testing.T) {
	ab := NewAddressBook()

	ab.AddPeer(testPeerID("peer1"), testMultiaddrs(t, "/ip4/1.1.1.1/tcp/9000"), "dht")
	ab.AddPeer(testPeerID("peer2"), testMultiaddrs(t, "/ip4/2.2.2.2/tcp/9000"), "dht")

	ab.CleanStale(1 * time.Hour)

	if ab.Len() != 2 {
		t.Errorf("expected 2 entries (all fresh), got %d", ab.Len())
	}
}

func TestAddressBook_CleanStale_AllStale(t *testing.T) {
	ab := NewAddressBook()

	pid1 := testPeerID("peer1")
	pid2 := testPeerID("peer2")

	ab.AddPeer(pid1, testMultiaddrs(t, "/ip4/1.1.1.1/tcp/9000"), "dht")
	ab.AddPeer(pid2, testMultiaddrs(t, "/ip4/2.2.2.2/tcp/9000"), "dht")

	// Backdate both
	ab.mu.Lock()
	ab.entries[pid1].LastSeen = time.Now().Add(-2 * time.Hour)
	ab.entries[pid2].LastSeen = time.Now().Add(-3 * time.Hour)
	ab.mu.Unlock()

	ab.CleanStale(1 * time.Hour)

	if ab.Len() != 0 {
		t.Errorf("expected 0 entries (all stale), got %d", ab.Len())
	}
}

func TestAddressBook_SaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "addressbook.json")

	ab := NewAddressBook()

	// Use valid libp2p peer IDs for JSON round-trip (peer.ID has custom JSON marshaling)
	pid1 := randTestPeerID(t)
	pid2 := randTestPeerID(t)

	ab.AddPeer(pid1, testMultiaddrs(t, "/ip4/192.168.1.1/tcp/9000", "/ip4/10.0.0.1/tcp/9000"), "dht")
	ab.AddPeer(pid2, testMultiaddrs(t, "/ip4/172.16.0.1/tcp/9001"), "mdns")

	// Record some connection stats
	ab.RecordConnectionAttempt(pid1, true)
	ab.RecordConnectionAttempt(pid1, false)
	ab.RecordConnectionAttempt(pid2, true)

	// Save
	if err := ab.Save(path); err != nil {
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

	// Load into a new AddressBook
	ab2 := NewAddressBook()
	if err := ab2.Load(path); err != nil {
		t.Fatalf("Load error: %v", err)
	}

	if ab2.Len() != 2 {
		t.Fatalf("expected 2 entries after load, got %d", ab2.Len())
	}

	// Verify entries were preserved
	peers := ab2.GetPeers()
	peerMap := make(map[peer.ID]*AddressEntry)
	for _, entry := range peers {
		peerMap[entry.PeerID] = entry
	}

	entry1, ok := peerMap[pid1]
	if !ok {
		t.Fatal("expected peer1 to be present after load")
	}
	if len(entry1.Addrs) != 2 {
		t.Errorf("expected 2 addresses for peer1, got %d", len(entry1.Addrs))
	}
	if entry1.ConnAttempts != 2 {
		t.Errorf("expected 2 conn attempts for peer1, got %d", entry1.ConnAttempts)
	}
	if entry1.ConnSuccess != 1 {
		t.Errorf("expected 1 conn success for peer1, got %d", entry1.ConnSuccess)
	}
	if entry1.Source != "dht" {
		t.Errorf("expected source 'dht' for peer1, got %s", entry1.Source)
	}

	entry2, ok := peerMap[pid2]
	if !ok {
		t.Fatal("expected peer2 to be present after load")
	}
	if entry2.Source != "mdns" {
		t.Errorf("expected source 'mdns' for peer2, got %s", entry2.Source)
	}
}

func TestAddressBook_SaveCreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "subdir", "nested", "addressbook.json")

	ab := NewAddressBook()
	ab.AddPeer(randTestPeerID(t), testMultiaddrs(t, "/ip4/1.1.1.1/tcp/9000"), "manual")

	if err := ab.Save(path); err != nil {
		t.Fatalf("Save error: %v", err)
	}

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected file to exist: %v", err)
	}
}

func TestAddressBook_LoadNonExistentFile(t *testing.T) {
	ab := NewAddressBook()
	err := ab.Load("/tmp/nonexistent-addressbook-test-12345.json")
	if err != nil {
		t.Errorf("expected no error for non-existent file, got: %v", err)
	}
	if ab.Len() != 0 {
		t.Errorf("expected 0 entries after loading non-existent file, got %d", ab.Len())
	}
}

func TestAddressBook_LoadInvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "addressbook.json")

	if err := os.WriteFile(path, []byte("not valid json{"), 0600); err != nil {
		t.Fatalf("write error: %v", err)
	}

	ab := NewAddressBook()
	err := ab.Load(path)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestAddressBook_LoadEmptyArray(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "addressbook.json")

	if err := os.WriteFile(path, []byte("[]"), 0600); err != nil {
		t.Fatalf("write error: %v", err)
	}

	ab := NewAddressBook()
	if err := ab.Load(path); err != nil {
		t.Fatalf("Load error: %v", err)
	}

	if ab.Len() != 0 {
		t.Errorf("expected 0 entries, got %d", ab.Len())
	}
}

func TestAddressBook_SaveEmptyBook(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "addressbook.json")

	ab := NewAddressBook()
	if err := ab.Save(path); err != nil {
		t.Fatalf("Save error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read error: %v", err)
	}

	var entries []*AddressEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected empty array in file, got %d entries", len(entries))
	}
}

func TestAddressBook_ConcurrentAccess(t *testing.T) {
	ab := NewAddressBook()

	var wg sync.WaitGroup
	const goroutines = 20

	// Concurrent adds
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			pid := testPeerID(string(rune('A' + n)))
			addrs := testMultiaddrs(t, "/ip4/1.1.1.1/tcp/9000")
			ab.AddPeer(pid, addrs, "dht")
		}(i)
	}
	wg.Wait()

	if ab.Len() != goroutines {
		t.Errorf("expected %d entries after concurrent adds, got %d", goroutines, ab.Len())
	}

	// Concurrent reads + writes
	for i := 0; i < goroutines; i++ {
		wg.Add(3)
		go func(n int) {
			defer wg.Done()
			ab.GetPeers()
		}(i)
		go func(n int) {
			defer wg.Done()
			ab.GetBestPeers(5)
		}(i)
		go func(n int) {
			defer wg.Done()
			pid := testPeerID(string(rune('A' + n)))
			ab.UpdateLastSeen(pid)
			ab.RecordConnectionAttempt(pid, n%2 == 0)
		}(i)
	}
	wg.Wait()

	// Should still have all entries (no removes)
	if ab.Len() != goroutines {
		t.Errorf("expected %d entries after concurrent ops, got %d", goroutines, ab.Len())
	}
}

func TestAddressBook_GetPeers_ReturnsCopies(t *testing.T) {
	ab := NewAddressBook()

	pid := testPeerID("peer1")
	ab.AddPeer(pid, testMultiaddrs(t, "/ip4/1.1.1.1/tcp/9000"), "dht")

	peers := ab.GetPeers()
	// Modify the returned entry
	peers[0].ConnAttempts = 999
	peers[0].Addrs = append(peers[0].Addrs, "/ip4/2.2.2.2/tcp/9000")

	// Original should be unchanged
	peers2 := ab.GetPeers()
	if peers2[0].ConnAttempts != 0 {
		t.Errorf("expected ConnAttempts to be 0 (original unchanged), got %d", peers2[0].ConnAttempts)
	}
	if len(peers2[0].Addrs) != 1 {
		t.Errorf("expected 1 address (original unchanged), got %d", len(peers2[0].Addrs))
	}
}

func TestAddressEntry_SuccessRate(t *testing.T) {
	tests := []struct {
		name     string
		attempts int
		success  int
		expected float64
	}{
		{"no attempts", 0, 0, 0},
		{"all success", 4, 4, 1.0},
		{"all failure", 4, 0, 0},
		{"half", 4, 2, 0.5},
		{"one third", 3, 1, 1.0 / 3.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := &AddressEntry{
				ConnAttempts: tt.attempts,
				ConnSuccess:  tt.success,
			}
			rate := entry.successRate()
			if rate != tt.expected {
				t.Errorf("expected success rate %f, got %f", tt.expected, rate)
			}
		})
	}
}
