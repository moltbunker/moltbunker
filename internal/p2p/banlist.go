package p2p

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/moltbunker/moltbunker/internal/logging"
	"github.com/moltbunker/moltbunker/pkg/types"
)

// BanEntry represents a single ban record for a peer.
type BanEntry struct {
	PeerID    types.NodeID `json:"peer_id"`
	Reason    string       `json:"reason"`
	BannedAt  time.Time    `json:"banned_at"`
	ExpiresAt time.Time    `json:"expires_at"` // Zero value means permanent ban
}

// IsPermanent returns true if the ban has no expiry.
func (e BanEntry) IsPermanent() bool {
	return e.ExpiresAt.IsZero()
}

// IsExpired returns true if the ban has an expiry time that has passed.
func (e BanEntry) IsExpired() bool {
	if e.IsPermanent() {
		return false
	}
	return time.Now().After(e.ExpiresAt)
}

// BanList is a thread-safe collection of banned peers.
type BanList struct {
	mu      sync.RWMutex
	entries map[types.NodeID]BanEntry
}

// NewBanList creates a new empty BanList.
func NewBanList() *BanList {
	return &BanList{
		entries: make(map[types.NodeID]BanEntry),
	}
}

// Ban adds a peer to the ban list. A duration of 0 means a permanent ban.
func (bl *BanList) Ban(peerID types.NodeID, reason string, duration time.Duration) {
	bl.mu.Lock()
	defer bl.mu.Unlock()

	entry := BanEntry{
		PeerID:   peerID,
		Reason:   reason,
		BannedAt: time.Now(),
	}
	if duration > 0 {
		entry.ExpiresAt = time.Now().Add(duration)
	}

	bl.entries[peerID] = entry

	if duration > 0 {
		logging.Info("peer banned",
			logging.NodeID(peerID.String()[:16]),
			"reason", reason,
			"duration", duration.String(),
			logging.Component("banlist"))
	} else {
		logging.Info("peer banned permanently",
			logging.NodeID(peerID.String()[:16]),
			"reason", reason,
			logging.Component("banlist"))
	}
}

// Unban removes a peer from the ban list.
func (bl *BanList) Unban(peerID types.NodeID) {
	bl.mu.Lock()
	defer bl.mu.Unlock()

	if _, exists := bl.entries[peerID]; exists {
		delete(bl.entries, peerID)
		logging.Info("peer unbanned",
			logging.NodeID(peerID.String()[:16]),
			logging.Component("banlist"))
	}
}

// IsBanned returns true if the peer is currently banned (and the ban has not expired).
func (bl *BanList) IsBanned(peerID types.NodeID) bool {
	bl.mu.RLock()
	defer bl.mu.RUnlock()

	entry, exists := bl.entries[peerID]
	if !exists {
		return false
	}

	// If the ban has expired, it is no longer active.
	// We don't remove it here (read lock); CleanExpired handles removal.
	if entry.IsExpired() {
		return false
	}

	return true
}

// CleanExpired removes all expired bans from the list.
func (bl *BanList) CleanExpired() {
	bl.mu.Lock()
	defer bl.mu.Unlock()

	var removed int
	for id, entry := range bl.entries {
		if entry.IsExpired() {
			delete(bl.entries, id)
			removed++
		}
	}

	if removed > 0 {
		logging.Debug("cleaned expired bans",
			"removed", removed,
			logging.Component("banlist"))
	}
}

// List returns all currently active (non-expired) ban entries.
func (bl *BanList) List() []BanEntry {
	bl.mu.RLock()
	defer bl.mu.RUnlock()

	var result []BanEntry
	for _, entry := range bl.entries {
		if !entry.IsExpired() {
			result = append(result, entry)
		}
	}
	return result
}

// Len returns the number of active (non-expired) bans.
func (bl *BanList) Len() int {
	bl.mu.RLock()
	defer bl.mu.RUnlock()

	count := 0
	for _, entry := range bl.entries {
		if !entry.IsExpired() {
			count++
		}
	}
	return count
}

// Save persists the ban list to a JSON file using an atomic write
// (write to .tmp then rename).
func (bl *BanList) Save(path string) error {
	bl.mu.RLock()
	entries := make([]BanEntry, 0, len(bl.entries))
	for _, entry := range bl.entries {
		entries = append(entries, entry)
	}
	bl.mu.RUnlock()

	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal ban list: %w", err)
	}

	// Ensure the parent directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("create directory for ban list: %w", err)
	}

	// Atomic write: write to .tmp then rename
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0600); err != nil {
		return fmt.Errorf("write ban list temp file: %w", err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		// Clean up temp file on rename failure
		os.Remove(tmpPath)
		return fmt.Errorf("rename ban list file: %w", err)
	}

	logging.Debug("saved ban list",
		"entries", len(entries),
		"path", path,
		logging.Component("banlist"))

	return nil
}

// Load reads the ban list from a JSON file. If the file does not exist,
// the ban list is left empty (no error).
func (bl *BanList) Load(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			logging.Debug("no ban list file found, starting fresh",
				"path", path,
				logging.Component("banlist"))
			return nil
		}
		return fmt.Errorf("read ban list: %w", err)
	}

	var entries []BanEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return fmt.Errorf("unmarshal ban list: %w", err)
	}

	bl.mu.Lock()
	defer bl.mu.Unlock()

	bl.entries = make(map[types.NodeID]BanEntry, len(entries))
	for _, entry := range entries {
		bl.entries[entry.PeerID] = entry
	}

	logging.Info("loaded ban list",
		"entries", len(entries),
		"path", path,
		logging.Component("banlist"))

	return nil
}
