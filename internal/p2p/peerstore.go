package p2p

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	ma "github.com/multiformats/go-multiaddr"

	"github.com/moltbunker/moltbunker/internal/logging"
)

// PeerRecord stores serializable peer information for routing table persistence.
type PeerRecord struct {
	ID        string    `json:"id"`
	Addresses []string  `json:"addresses"`
	LastSeen  time.Time `json:"last_seen"`
}

// PeerStore handles persisting and loading DHT routing table state.
type PeerStore struct {
	path string
}

// NewPeerStore creates a new PeerStore that reads/writes to the given file path.
func NewPeerStore(path string) *PeerStore {
	return &PeerStore{path: path}
}

// SaveRoutingTable extracts known peers from the DHT and persists them to disk.
// The write is atomic: data is written to a temporary file and then renamed.
func (d *DHT) SaveRoutingTable(path string) error {
	records := d.collectPeerRecords()

	if len(records) == 0 {
		logging.Debug("no peers to save in routing table",
			logging.Component("dht"))
		// Still write an empty array so the file reflects reality
	}

	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal routing table: %w", err)
	}

	// Ensure the parent directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("create directory for routing table: %w", err)
	}

	// Atomic write: write to .tmp then rename
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0600); err != nil {
		return fmt.Errorf("write routing table temp file: %w", err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		// Clean up temp file on rename failure
		os.Remove(tmpPath)
		return fmt.Errorf("rename routing table file: %w", err)
	}

	logging.Info("saved routing table",
		"peers", len(records),
		"path", path,
		logging.Component("dht"))

	return nil
}

// LoadRoutingTable reads persisted peer records from disk and attempts to
// connect to each saved peer. Connection failures are logged but not fatal.
func (d *DHT) LoadRoutingTable(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			logging.Debug("no routing table file found, starting fresh",
				"path", path,
				logging.Component("dht"))
			return nil
		}
		return fmt.Errorf("read routing table: %w", err)
	}

	var records []PeerRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return fmt.Errorf("unmarshal routing table: %w", err)
	}

	if len(records) == 0 {
		logging.Debug("routing table file is empty",
			"path", path,
			logging.Component("dht"))
		return nil
	}

	logging.Info("loading routing table",
		"peers", len(records),
		"path", path,
		logging.Component("dht"))

	var connected, failed int
	for _, record := range records {
		if err := d.connectToPeerRecord(record); err != nil {
			logging.Debug("failed to connect to saved peer",
				"peer_id", record.ID,
				"error", err.Error(),
				logging.Component("dht"))
			failed++
		} else {
			connected++
		}
	}

	logging.Info("routing table loaded",
		"connected", connected,
		"failed", failed,
		"total", len(records),
		logging.Component("dht"))

	return nil
}

// collectPeerRecords gathers peer information from the libp2p host's
// connected peers and from the internal peer map.
func (d *DHT) collectPeerRecords() []PeerRecord {
	seen := make(map[string]bool)
	var records []PeerRecord

	// Collect from connected libp2p peers
	for _, pID := range d.host.Network().Peers() {
		idStr := pID.String()
		if seen[idStr] {
			continue
		}
		seen[idStr] = true

		addrs := d.host.Peerstore().Addrs(pID)
		if len(addrs) == 0 {
			continue
		}

		addrStrs := make([]string, 0, len(addrs))
		for _, addr := range addrs {
			addrStrs = append(addrStrs, addr.String())
		}

		records = append(records, PeerRecord{
			ID:        idStr,
			Addresses: addrStrs,
			LastSeen:  time.Now(),
		})
	}

	// Also collect from internal peer map (may have peers not currently connected)
	d.peersMu.RLock()
	for _, node := range d.peers {
		nodeIDStr := node.ID.String()
		if seen[nodeIDStr] {
			continue
		}
		seen[nodeIDStr] = true

		if node.Address == "" {
			continue
		}

		records = append(records, PeerRecord{
			ID:        nodeIDStr,
			Addresses: []string{node.Address},
			LastSeen:  node.LastSeen,
		})
	}
	d.peersMu.RUnlock()

	return records
}

// connectToPeerRecord connects to a peer from a saved PeerRecord.
func (d *DHT) connectToPeerRecord(record PeerRecord) error {
	peerID, err := peer.Decode(record.ID)
	if err != nil {
		return fmt.Errorf("decode peer ID %q: %w", record.ID, err)
	}

	// Skip ourselves
	if peerID == d.host.ID() {
		return nil
	}

	// Parse multiaddresses
	var addrs []ma.Multiaddr
	for _, addrStr := range record.Addresses {
		maddr, err := ma.NewMultiaddr(addrStr)
		if err != nil {
			logging.Debug("skipping invalid multiaddr from routing table",
				"addr", addrStr,
				"error", err.Error(),
				logging.Component("dht"))
			continue
		}
		addrs = append(addrs, maddr)
	}

	if len(addrs) == 0 {
		return fmt.Errorf("no valid addresses for peer %s", record.ID)
	}

	// Add addresses to the peerstore with a reasonable TTL
	d.host.Peerstore().AddAddrs(peerID, addrs, peerstore.PermanentAddrTTL)

	// Attempt connection with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := d.host.Connect(ctx, peer.AddrInfo{
		ID:    peerID,
		Addrs: addrs,
	}); err != nil {
		return fmt.Errorf("connect to peer %s: %w", record.ID, err)
	}

	return nil
}

// SavePeers is a convenience method on PeerStore that saves the DHT routing table.
func (ps *PeerStore) SavePeers(d *DHT) error {
	return d.SaveRoutingTable(ps.path)
}

// LoadPeers is a convenience method on PeerStore that loads the DHT routing table.
func (ps *PeerStore) LoadPeers(d *DHT) error {
	return d.LoadRoutingTable(ps.path)
}
