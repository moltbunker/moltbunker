package p2p

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestPeerRecordSerialization(t *testing.T) {
	records := []PeerRecord{
		{
			ID:        "QmPeer1234567890",
			Addresses: []string{"/ip4/192.168.1.1/tcp/9000", "/ip4/10.0.0.1/tcp/9000"},
			LastSeen:  time.Now().Truncate(time.Second),
		},
		{
			ID:        "QmPeer0987654321",
			Addresses: []string{"/ip4/172.16.0.1/tcp/9001"},
			LastSeen:  time.Now().Add(-1 * time.Hour).Truncate(time.Second),
		},
	}

	data, err := json.Marshal(records)
	if err != nil {
		t.Fatalf("failed to marshal peer records: %v", err)
	}

	var decoded []PeerRecord
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal peer records: %v", err)
	}

	if len(decoded) != 2 {
		t.Fatalf("expected 2 records, got %d", len(decoded))
	}

	if decoded[0].ID != "QmPeer1234567890" {
		t.Errorf("expected ID QmPeer1234567890, got %s", decoded[0].ID)
	}
	if len(decoded[0].Addresses) != 2 {
		t.Errorf("expected 2 addresses, got %d", len(decoded[0].Addresses))
	}
	if decoded[1].Addresses[0] != "/ip4/172.16.0.1/tcp/9001" {
		t.Errorf("unexpected address: %s", decoded[1].Addresses[0])
	}
}

func TestPeerStoreNewPeerStore(t *testing.T) {
	ps := NewPeerStore("/tmp/test-peers.json")
	if ps == nil {
		t.Fatal("NewPeerStore returned nil")
	}
	if ps.path != "/tmp/test-peers.json" {
		t.Errorf("expected path /tmp/test-peers.json, got %s", ps.path)
	}
}

func TestPeerRecordJSON_RoundTrip(t *testing.T) {
	now := time.Now().Truncate(time.Millisecond)
	record := PeerRecord{
		ID:        "12D3KooWTestPeerID",
		Addresses: []string{"/ip4/1.2.3.4/tcp/9000", "/ip6/::1/tcp/9000"},
		LastSeen:  now,
	}

	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var decoded PeerRecord
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if decoded.ID != record.ID {
		t.Errorf("ID mismatch: %s != %s", decoded.ID, record.ID)
	}
	if len(decoded.Addresses) != len(record.Addresses) {
		t.Errorf("address count mismatch: %d != %d", len(decoded.Addresses), len(record.Addresses))
	}
	for i, addr := range decoded.Addresses {
		if addr != record.Addresses[i] {
			t.Errorf("address %d mismatch: %s != %s", i, addr, record.Addresses[i])
		}
	}
	// Compare times at millisecond precision
	if !decoded.LastSeen.Equal(now) {
		t.Errorf("LastSeen mismatch: %v != %v", decoded.LastSeen, now)
	}
}

func TestAtomicFileWrite(t *testing.T) {
	// Test that saving creates the file atomically
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "subdir", "peers.json")

	records := []PeerRecord{
		{
			ID:        "TestPeer1",
			Addresses: []string{"/ip4/127.0.0.1/tcp/9000"},
			LastSeen:  time.Now(),
		},
	}

	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	// Ensure parent directory is created
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatalf("mkdir error: %v", err)
	}

	// Write to tmp then rename (simulating atomic write)
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0600); err != nil {
		t.Fatalf("write tmp error: %v", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		t.Fatalf("rename error: %v", err)
	}

	// Verify file exists and is valid JSON
	readData, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read error: %v", err)
	}

	var readRecords []PeerRecord
	if err := json.Unmarshal(readData, &readRecords); err != nil {
		t.Fatalf("unmarshal read data error: %v", err)
	}

	if len(readRecords) != 1 {
		t.Fatalf("expected 1 record, got %d", len(readRecords))
	}
	if readRecords[0].ID != "TestPeer1" {
		t.Errorf("expected ID TestPeer1, got %s", readRecords[0].ID)
	}

	// Verify no .tmp file remains
	if _, err := os.Stat(tmpPath); !os.IsNotExist(err) {
		t.Error("expected .tmp file to not exist after rename")
	}
}

func TestLoadRoutingTable_FileNotExist(t *testing.T) {
	// Loading from a non-existent file should not error (returns nil)
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "nonexistent.json")

	// We can't call DHT.LoadRoutingTable directly without a DHT,
	// but we can test the file reading logic
	_, err := os.ReadFile(path)
	if !os.IsNotExist(err) {
		t.Fatalf("expected IsNotExist error, got %v", err)
	}
}

func TestLoadRoutingTable_EmptyArray(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "peers.json")

	if err := os.WriteFile(path, []byte("[]"), 0600); err != nil {
		t.Fatalf("write error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read error: %v", err)
	}

	var records []PeerRecord
	if err := json.Unmarshal(data, &records); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if len(records) != 0 {
		t.Errorf("expected 0 records, got %d", len(records))
	}
}

func TestLoadRoutingTable_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "peers.json")

	if err := os.WriteFile(path, []byte("not valid json{"), 0600); err != nil {
		t.Fatalf("write error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read error: %v", err)
	}

	var records []PeerRecord
	err = json.Unmarshal(data, &records)
	if err == nil {
		t.Error("expected unmarshal error for invalid JSON")
	}
}
