package p2p

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"

	"github.com/moltbunker/moltbunker/internal/logging"
)

// AddressEntry stores peer address information with connection statistics.
type AddressEntry struct {
	PeerID       peer.ID  `json:"peer_id"`
	Addrs        []string `json:"addrs"`          // multiaddr strings
	LastSeen     time.Time `json:"last_seen"`
	ConnAttempts int      `json:"conn_attempts"`
	ConnSuccess  int      `json:"conn_success"`
	Source       string   `json:"source"`          // "dht", "mdns", "manual", "bootstrap"
}

// successRate returns the connection success rate as a float64 in [0, 1].
// Returns 0 if no connection attempts have been made.
func (e *AddressEntry) successRate() float64 {
	if e.ConnAttempts == 0 {
		return 0
	}
	return float64(e.ConnSuccess) / float64(e.ConnAttempts)
}

// AddressBook is a thread-safe persistent address book for libp2p peers.
// It tracks peer addresses, connection statistics, and discovery sources.
type AddressBook struct {
	entries map[peer.ID]*AddressEntry
	mu      sync.RWMutex
}

// NewAddressBook creates a new empty AddressBook.
func NewAddressBook() *AddressBook {
	return &AddressBook{
		entries: make(map[peer.ID]*AddressEntry),
	}
}

// AddPeer adds a new peer or updates an existing peer's addresses and source.
// Addresses are merged (deduplicated) with any existing addresses.
func (ab *AddressBook) AddPeer(peerID peer.ID, addrs []ma.Multiaddr, source string) {
	ab.mu.Lock()
	defer ab.mu.Unlock()

	addrStrs := make([]string, 0, len(addrs))
	for _, addr := range addrs {
		addrStrs = append(addrStrs, addr.String())
	}

	entry, exists := ab.entries[peerID]
	if !exists {
		ab.entries[peerID] = &AddressEntry{
			PeerID:   peerID,
			Addrs:    addrStrs,
			LastSeen: time.Now(),
			Source:   source,
		}
		return
	}

	// Merge addresses (deduplicate)
	existing := make(map[string]bool, len(entry.Addrs))
	for _, a := range entry.Addrs {
		existing[a] = true
	}
	for _, a := range addrStrs {
		if !existing[a] {
			entry.Addrs = append(entry.Addrs, a)
		}
	}

	entry.LastSeen = time.Now()
	entry.Source = source
}

// UpdateLastSeen updates the last seen timestamp for a peer.
// If the peer does not exist, this is a no-op.
func (ab *AddressBook) UpdateLastSeen(peerID peer.ID) {
	ab.mu.Lock()
	defer ab.mu.Unlock()

	if entry, exists := ab.entries[peerID]; exists {
		entry.LastSeen = time.Now()
	}
}

// RecordConnectionAttempt records a connection attempt and its result.
// If the peer does not exist, this is a no-op.
func (ab *AddressBook) RecordConnectionAttempt(peerID peer.ID, success bool) {
	ab.mu.Lock()
	defer ab.mu.Unlock()

	entry, exists := ab.entries[peerID]
	if !exists {
		return
	}

	entry.ConnAttempts++
	if success {
		entry.ConnSuccess++
	}
}

// GetPeers returns all entries in the address book.
func (ab *AddressBook) GetPeers() []*AddressEntry {
	ab.mu.RLock()
	defer ab.mu.RUnlock()

	entries := make([]*AddressEntry, 0, len(ab.entries))
	for _, entry := range ab.entries {
		cp := *entry
		cp.Addrs = make([]string, len(entry.Addrs))
		copy(cp.Addrs, entry.Addrs)
		entries = append(entries, &cp)
	}
	return entries
}

// GetBestPeers returns the top n peers sorted by success rate (descending),
// then by last seen time (most recent first) as a tiebreaker.
func (ab *AddressBook) GetBestPeers(n int) []*AddressEntry {
	entries := ab.GetPeers()

	sort.Slice(entries, func(i, j int) bool {
		ri := entries[i].successRate()
		rj := entries[j].successRate()
		if ri != rj {
			return ri > rj
		}
		return entries[i].LastSeen.After(entries[j].LastSeen)
	})

	if n > len(entries) {
		n = len(entries)
	}
	return entries[:n]
}

// RemovePeer removes a peer from the address book.
func (ab *AddressBook) RemovePeer(peerID peer.ID) {
	ab.mu.Lock()
	defer ab.mu.Unlock()

	delete(ab.entries, peerID)
}

// CleanStale removes entries that have not been seen within maxAge.
func (ab *AddressBook) CleanStale(maxAge time.Duration) {
	ab.mu.Lock()
	defer ab.mu.Unlock()

	cutoff := time.Now().Add(-maxAge)
	var removed int
	for id, entry := range ab.entries {
		if entry.LastSeen.Before(cutoff) {
			delete(ab.entries, id)
			removed++
		}
	}

	if removed > 0 {
		logging.Debug("cleaned stale address book entries",
			"removed", removed,
			logging.Component("addressbook"))
	}
}

// Save persists the address book to a JSON file using an atomic write
// (write to .tmp then rename).
func (ab *AddressBook) Save(path string) error {
	ab.mu.RLock()
	entries := make([]*AddressEntry, 0, len(ab.entries))
	for _, entry := range ab.entries {
		cp := *entry
		cp.Addrs = make([]string, len(entry.Addrs))
		copy(cp.Addrs, entry.Addrs)
		entries = append(entries, &cp)
	}
	ab.mu.RUnlock()

	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal address book: %w", err)
	}

	// Ensure the parent directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("create directory for address book: %w", err)
	}

	// Atomic write: write to .tmp then rename
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0600); err != nil {
		return fmt.Errorf("write address book temp file: %w", err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		// Clean up temp file on rename failure
		os.Remove(tmpPath)
		return fmt.Errorf("rename address book file: %w", err)
	}

	logging.Debug("saved address book",
		"entries", len(entries),
		"path", path,
		logging.Component("addressbook"))

	return nil
}

// Load reads the address book from a JSON file. If the file does not exist,
// the address book is left empty (no error returned).
func (ab *AddressBook) Load(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			logging.Debug("no address book file found, starting fresh",
				"path", path,
				logging.Component("addressbook"))
			return nil
		}
		return fmt.Errorf("read address book: %w", err)
	}

	var entries []*AddressEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return fmt.Errorf("unmarshal address book: %w", err)
	}

	ab.mu.Lock()
	defer ab.mu.Unlock()

	ab.entries = make(map[peer.ID]*AddressEntry, len(entries))
	for _, entry := range entries {
		ab.entries[entry.PeerID] = entry
	}

	logging.Info("loaded address book",
		"entries", len(entries),
		"path", path,
		logging.Component("addressbook"))

	return nil
}

// Len returns the number of entries in the address book.
func (ab *AddressBook) Len() int {
	ab.mu.RLock()
	defer ab.mu.RUnlock()

	return len(ab.entries)
}
