package cloning

import (
	"context"
	"crypto/sha256"
	"fmt"
	"testing"
	"time"

	"github.com/moltbunker/moltbunker/pkg/types"
)

// --- Mock implementations for testing ---

type mockNodeSelector struct {
	nodes []*types.Node
	err   error
}

func (m *mockNodeSelector) FindNodes(ctx context.Context, region string, excludeNodes []types.NodeID) ([]*types.Node, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.nodes, nil
}

func (m *mockNodeSelector) GetNodeByID(ctx context.Context, nodeID types.NodeID) (*types.Node, error) {
	if m.err != nil {
		return nil, m.err
	}
	for _, n := range m.nodes {
		if n.ID == nodeID {
			return n, nil
		}
	}
	return nil, fmt.Errorf("node not found")
}

type mockDeployer struct {
	deployErr error
	verifyErr error
}

func (m *mockDeployer) DeployClone(ctx context.Context, clone *Clone, stateData []byte) error {
	if m.deployErr != nil {
		return m.deployErr
	}
	clone.TargetID = "target-" + clone.ID
	return nil
}

func (m *mockDeployer) VerifyDeployment(ctx context.Context, clone *Clone) error {
	return m.verifyErr
}

func makeNodeID(seed byte) types.NodeID {
	var id types.NodeID
	for i := range id {
		id[i] = seed
	}
	return id
}

func makeNodes(count int) []*types.Node {
	nodes := make([]*types.Node, count)
	for i := 0; i < count; i++ {
		nodes[i] = &types.Node{
			ID:       makeNodeID(byte(i + 1)),
			Region:   "americas",
			LastSeen: time.Now(),
			Capabilities: types.NodeCapabilities{
				ContainerRuntime: true,
			},
		}
	}
	return nodes
}

// --- DefaultCloneConfig tests ---

func TestDefaultCloneConfig(t *testing.T) {
	cfg := DefaultCloneConfig()
	if cfg == nil {
		t.Fatal("DefaultCloneConfig returned nil")
	}
	if cfg.MaxConcurrentClones != 3 {
		t.Errorf("expected MaxConcurrentClones 3, got %d", cfg.MaxConcurrentClones)
	}
	if cfg.CloneTimeoutSeconds != 300 {
		t.Errorf("expected CloneTimeoutSeconds 300, got %d", cfg.CloneTimeoutSeconds)
	}
	if cfg.MaxRetries != 3 {
		t.Errorf("expected MaxRetries 3, got %d", cfg.MaxRetries)
	}
	if !cfg.AvoidSameProvider {
		t.Error("expected AvoidSameProvider true by default")
	}
	if !cfg.IncludeState {
		t.Error("expected IncludeState true by default")
	}
	if !cfg.CompressTransfer {
		t.Error("expected CompressTransfer true by default")
	}
}

// --- Manager tests ---

func TestNewManager(t *testing.T) {
	t.Run("nil config uses defaults", func(t *testing.T) {
		m := NewManager(nil, nil, nil, nil)
		if m == nil {
			t.Fatal("NewManager returned nil")
		}
		if m.config.MaxConcurrentClones != 3 {
			t.Errorf("expected default MaxConcurrentClones 3, got %d", m.config.MaxConcurrentClones)
		}
	})

	t.Run("custom config", func(t *testing.T) {
		cfg := &CloneConfig{MaxConcurrentClones: 5, CloneTimeoutSeconds: 60}
		m := NewManager(cfg, nil, nil, nil)
		if m.config.MaxConcurrentClones != 5 {
			t.Errorf("expected MaxConcurrentClones 5, got %d", m.config.MaxConcurrentClones)
		}
	})
}

func TestManagerStartStop(t *testing.T) {
	m := NewManager(nil, nil, nil, nil)

	ctx := context.Background()
	if err := m.Start(ctx); err != nil {
		t.Fatalf("Start() error: %v", err)
	}

	m.mu.RLock()
	running := m.running
	m.mu.RUnlock()
	if !running {
		t.Error("expected running = true after Start()")
	}

	// Double start should be no-op
	if err := m.Start(ctx); err != nil {
		t.Fatalf("second Start() error: %v", err)
	}

	m.Stop()
	m.mu.RLock()
	running = m.running
	m.mu.RUnlock()
	if running {
		t.Error("expected running = false after Stop()")
	}
}

func TestManagerGetClone(t *testing.T) {
	m := NewManager(nil, nil, nil, nil)

	// Add a clone to active
	clone := &Clone{
		ID:       "test-clone-1",
		SourceID: "container-1",
		Status:   CloneStatusPending,
	}
	m.activeClones["test-clone-1"] = clone

	// Should find in active
	found, ok := m.GetClone("test-clone-1")
	if !ok {
		t.Fatal("expected to find clone in active clones")
	}
	if found.ID != "test-clone-1" {
		t.Errorf("expected ID 'test-clone-1', got %s", found.ID)
	}

	// Move to history
	delete(m.activeClones, "test-clone-1")
	m.cloneHistory = append(m.cloneHistory, clone)

	// Should find in history
	found, ok = m.GetClone("test-clone-1")
	if !ok {
		t.Fatal("expected to find clone in history")
	}

	// Should not find nonexistent
	_, ok = m.GetClone("nonexistent")
	if ok {
		t.Error("should not find nonexistent clone")
	}
}

func TestManagerGetActiveClones(t *testing.T) {
	m := NewManager(nil, nil, nil, nil)
	m.activeClones["a"] = &Clone{ID: "a"}
	m.activeClones["b"] = &Clone{ID: "b"}

	active := m.GetActiveClones()
	if len(active) != 2 {
		t.Errorf("expected 2 active clones, got %d", len(active))
	}
}

func TestManagerGetCloneHistory(t *testing.T) {
	m := NewManager(nil, nil, nil, nil)

	// Add 5 items to history
	for i := 0; i < 5; i++ {
		m.cloneHistory = append(m.cloneHistory, &Clone{
			ID:     fmt.Sprintf("clone-%d", i),
			Status: CloneStatusComplete,
		})
	}

	// Get all
	history := m.GetCloneHistory(0)
	if len(history) != 5 {
		t.Errorf("expected 5 history items, got %d", len(history))
	}

	// Get limited
	history = m.GetCloneHistory(3)
	if len(history) != 3 {
		t.Errorf("expected 3 history items, got %d", len(history))
	}
	// Should be most recent first
	if history[0].ID != "clone-4" {
		t.Errorf("expected most recent clone-4 first, got %s", history[0].ID)
	}

	// Limit exceeding total returns all
	history = m.GetCloneHistory(100)
	if len(history) != 5 {
		t.Errorf("expected 5 history items for large limit, got %d", len(history))
	}
}

func TestManagerGetLineage(t *testing.T) {
	m := NewManager(nil, nil, nil, nil)

	m.lineage["source-1"] = []string{"clone-a", "clone-b"}

	lineage := m.GetLineage("source-1")
	if len(lineage) != 2 {
		t.Errorf("expected 2 lineage entries, got %d", len(lineage))
	}

	// Modifying returned slice should not affect internal state
	lineage[0] = "modified"
	internal := m.GetLineage("source-1")
	if internal[0] == "modified" {
		t.Error("GetLineage should return a copy")
	}

	// Nonexistent source
	empty := m.GetLineage("nonexistent")
	if empty != nil {
		t.Errorf("expected nil for nonexistent source, got %v", empty)
	}
}

func TestManagerCancelClone(t *testing.T) {
	m := NewManager(nil, nil, nil, nil)

	// Add an active clone
	clone := &Clone{
		ID:       "cancel-me",
		SourceID: "src-1",
		Status:   CloneStatusPreparing,
	}
	m.activeClones["cancel-me"] = clone

	// Cancel it
	if err := m.CancelClone("cancel-me"); err != nil {
		t.Fatalf("CancelClone() error: %v", err)
	}

	// Should be in history, not active
	if _, ok := m.activeClones["cancel-me"]; ok {
		t.Error("clone should be removed from active clones after cancel")
	}
	if len(m.cloneHistory) != 1 {
		t.Fatalf("expected 1 history entry, got %d", len(m.cloneHistory))
	}
	if m.cloneHistory[0].Status != CloneStatusFailed {
		t.Errorf("expected status Failed, got %s", m.cloneHistory[0].Status)
	}
	if m.cloneHistory[0].Error != "cancelled" {
		t.Errorf("expected error 'cancelled', got %s", m.cloneHistory[0].Error)
	}

	// Cancel nonexistent should error
	if err := m.CancelClone("nonexistent"); err == nil {
		t.Error("expected error when cancelling nonexistent clone")
	}
}

func TestManagerGetStats(t *testing.T) {
	cfg := &CloneConfig{MaxConcurrentClones: 5, CloneTimeoutSeconds: 60}
	m := NewManager(cfg, nil, nil, nil)

	m.activeClones["a1"] = &Clone{ID: "a1", Status: CloneStatusPreparing}
	m.cloneHistory = append(m.cloneHistory,
		&Clone{ID: "h1", Status: CloneStatusComplete},
		&Clone{ID: "h2", Status: CloneStatusComplete},
		&Clone{ID: "h3", Status: CloneStatusFailed},
	)

	stats := m.GetStats()
	if stats.ActiveClones != 1 {
		t.Errorf("expected 1 active clone, got %d", stats.ActiveClones)
	}
	if stats.TotalCompleted != 2 {
		t.Errorf("expected 2 completed, got %d", stats.TotalCompleted)
	}
	if stats.TotalFailed != 1 {
		t.Errorf("expected 1 failed, got %d", stats.TotalFailed)
	}
	if stats.MaxConcurrent != 5 {
		t.Errorf("expected max concurrent 5, got %d", stats.MaxConcurrent)
	}
}

// --- StateTransfer tests ---

func TestNewStateTransfer(t *testing.T) {
	t.Run("nil config uses defaults", func(t *testing.T) {
		st := NewStateTransfer(nil)
		if st == nil {
			t.Fatal("NewStateTransfer returned nil")
		}
		if !st.compress {
			t.Error("expected compress true by default")
		}
		if !st.encrypt {
			t.Error("expected encrypt true by default")
		}
		if st.chunkSize != 1024*1024 {
			t.Errorf("expected chunk size 1MB, got %d", st.chunkSize)
		}
	})
}

func TestPackageAndUnpackageState(t *testing.T) {
	st := NewStateTransfer(&StateTransferConfig{
		Compress:  true,
		Encrypt:   false,
		ChunkSize: 1024,
	})

	data := []byte("hello world container state data for testing purposes")
	metadata := map[string]string{"key": "value"}

	pkg, err := st.PackageState("container-1", data, metadata)
	if err != nil {
		t.Fatalf("PackageState() error: %v", err)
	}

	if pkg.ContainerID != "container-1" {
		t.Errorf("expected ContainerID 'container-1', got %s", pkg.ContainerID)
	}
	if pkg.Size != int64(len(data)) {
		t.Errorf("expected Size %d, got %d", len(data), pkg.Size)
	}
	if !pkg.Compressed {
		t.Error("expected Compressed true")
	}
	if pkg.Encrypted {
		t.Error("expected Encrypted false")
	}
	if pkg.Version != 1 {
		t.Errorf("expected Version 1, got %d", pkg.Version)
	}

	// Verify checksum
	hash := sha256.Sum256(data)
	expectedChecksum := fmt.Sprintf("%x", hash)
	if pkg.Checksum != expectedChecksum {
		t.Error("checksum mismatch in package")
	}

	// Unpackage
	unpacked, err := st.UnpackageState(pkg, nil)
	if err != nil {
		t.Fatalf("UnpackageState() error: %v", err)
	}
	if string(unpacked) != string(data) {
		t.Errorf("unpacked data mismatch: got %q, want %q", string(unpacked), string(data))
	}
}

func TestPackageStateEncryptedRoundTrip(t *testing.T) {
	st := NewStateTransfer(&StateTransferConfig{
		Compress:  true,
		Encrypt:   true,
		ChunkSize: 1024,
	})

	data := []byte("secret container state that must be encrypted")
	key := []byte("my-secret-key-that-is-short")

	pkg, err := st.PackageStateEncrypted("container-2", data, key, nil)
	if err != nil {
		t.Fatalf("PackageStateEncrypted() error: %v", err)
	}
	if !pkg.Encrypted {
		t.Error("expected Encrypted true")
	}

	// Unpackage with correct key
	unpacked, err := st.UnpackageState(pkg, key)
	if err != nil {
		t.Fatalf("UnpackageState() error: %v", err)
	}
	if string(unpacked) != string(data) {
		t.Errorf("unpacked data mismatch")
	}
}

func TestUnpackageStateEncryptedNoKey(t *testing.T) {
	st := NewStateTransfer(&StateTransferConfig{
		Compress:  false,
		Encrypt:   true,
		ChunkSize: 1024,
	})

	data := []byte("secret data")
	key := []byte("my-key-for-encryption")

	pkg, err := st.PackageStateEncrypted("c1", data, key, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Attempt to unpackage without key
	_, err = st.UnpackageState(pkg, nil)
	if err == nil {
		t.Error("expected error when unpackaging encrypted data without key")
	}
}

func TestUnpackageStateChecksumMismatch(t *testing.T) {
	st := NewStateTransfer(&StateTransferConfig{
		Compress: false,
		Encrypt:  false,
	})

	data := []byte("original data")
	pkg, err := st.PackageState("c1", data, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Tamper with checksum
	pkg.Checksum = "deadbeef"

	_, err = st.UnpackageState(pkg, nil)
	if err == nil {
		t.Error("expected checksum mismatch error")
	}
}

func TestSerializeDeserializePackage(t *testing.T) {
	st := NewStateTransfer(&StateTransferConfig{
		Compress: false,
		Encrypt:  false,
	})

	pkg := &StatePackage{
		Version:     1,
		ContainerID: "c1",
		Timestamp:   time.Now(),
		Data:        []byte("data"),
		Metadata:    map[string]string{"k": "v"},
	}

	serialized, err := st.SerializePackage(pkg)
	if err != nil {
		t.Fatalf("SerializePackage() error: %v", err)
	}

	deserialized, err := st.DeserializePackage(serialized)
	if err != nil {
		t.Fatalf("DeserializePackage() error: %v", err)
	}

	if deserialized.ContainerID != "c1" {
		t.Errorf("expected container ID 'c1', got %s", deserialized.ContainerID)
	}
	if string(deserialized.Data) != "data" {
		t.Errorf("expected data 'data', got %q", string(deserialized.Data))
	}
}

func TestSplitAndAssembleChunks(t *testing.T) {
	st := NewStateTransfer(&StateTransferConfig{
		Compress:  false,
		Encrypt:   false,
		ChunkSize: 50, // Small chunk size for testing
	})

	data := []byte("this is test data for chunking that should be split into multiple chunks")
	pkg, err := st.PackageState("c1", data, nil)
	if err != nil {
		t.Fatal(err)
	}

	transfer, chunks, err := st.SplitIntoChunks(pkg)
	if err != nil {
		t.Fatalf("SplitIntoChunks() error: %v", err)
	}

	if transfer.ChunkCount != len(chunks) {
		t.Errorf("chunk count mismatch: transfer=%d, chunks=%d", transfer.ChunkCount, len(chunks))
	}
	if transfer.ChunkCount < 2 {
		t.Error("expected at least 2 chunks with small chunk size")
	}

	// Reassemble
	reassembled, err := st.AssembleChunks(transfer, chunks)
	if err != nil {
		t.Fatalf("AssembleChunks() error: %v", err)
	}

	if reassembled.ContainerID != "c1" {
		t.Errorf("expected container ID 'c1', got %s", reassembled.ContainerID)
	}
}

func TestAssembleChunksIncompleteChunks(t *testing.T) {
	st := NewStateTransfer(&StateTransferConfig{
		Compress:  false,
		Encrypt:   false,
		ChunkSize: 50,
	})

	transfer := &ChunkedTransfer{
		ChunkCount: 5,
	}

	// Only provide 3 chunks
	chunks := []Chunk{
		{Index: 0, Data: []byte("a")},
		{Index: 1, Data: []byte("b")},
		{Index: 2, Data: []byte("c")},
	}

	_, err := st.AssembleChunks(transfer, chunks)
	if err == nil {
		t.Error("expected error for incomplete chunks")
	}
}

func TestAssembleChunksInvalidIndex(t *testing.T) {
	st := NewStateTransfer(nil)

	transfer := &ChunkedTransfer{
		ChunkCount: 2,
	}

	chunks := []Chunk{
		{Index: 0, Data: []byte("a")},
		{Index: 99, Data: []byte("b")}, // Invalid index
	}

	_, err := st.AssembleChunks(transfer, chunks)
	if err == nil {
		t.Error("expected error for invalid chunk index")
	}
}

// --- AES-GCM encryption tests ---

func TestEncryptDecryptAESGCM(t *testing.T) {
	plaintext := []byte("sensitive container data")

	tests := []struct {
		name string
		key  []byte
	}{
		{"short key", []byte("short")},
		{"32 byte key", make([]byte, 32)},
		{"long key", make([]byte, 64)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encrypted, err := encryptAESGCM(plaintext, tt.key)
			if err != nil {
				t.Fatalf("encryptAESGCM() error: %v", err)
			}
			if string(encrypted) == string(plaintext) {
				t.Error("encrypted data should differ from plaintext")
			}

			decrypted, err := decryptAESGCM(encrypted, tt.key)
			if err != nil {
				t.Fatalf("decryptAESGCM() error: %v", err)
			}
			if string(decrypted) != string(plaintext) {
				t.Errorf("decrypted data mismatch: got %q, want %q", decrypted, plaintext)
			}
		})
	}
}

func TestDecryptAESGCMCiphertextTooShort(t *testing.T) {
	_, err := decryptAESGCM([]byte("short"), []byte("key"))
	if err == nil {
		t.Error("expected error for ciphertext too short")
	}
}

// --- RandomNodeSelector tests ---

func TestRandomNodeSelectorFindNodes(t *testing.T) {
	nodes := makeNodes(10)
	selector := NewRandomNodeSelector(nodes)

	ctx := context.Background()

	// Basic find
	found, err := selector.FindNodes(ctx, "", nil)
	if err != nil {
		t.Fatalf("FindNodes() error: %v", err)
	}
	if len(found) > 5 {
		t.Errorf("expected at most 5 nodes, got %d", len(found))
	}

	// Find with exclusion
	exclude := []types.NodeID{nodes[0].ID, nodes[1].ID}
	found, err = selector.FindNodes(ctx, "", exclude)
	if err != nil {
		t.Fatalf("FindNodes() error: %v", err)
	}
	for _, f := range found {
		if f.ID == nodes[0].ID || f.ID == nodes[1].ID {
			t.Error("excluded node should not be in results")
		}
	}

	// Find with region filter
	nodes[3].Region = "europe"
	found, err = selector.FindNodes(ctx, "europe", nil)
	if err != nil {
		t.Fatalf("FindNodes() error: %v", err)
	}
	if len(found) != 1 {
		t.Errorf("expected 1 node in europe, got %d", len(found))
	}
}

func TestRandomNodeSelectorEmpty(t *testing.T) {
	selector := NewRandomNodeSelector(nil)
	_, err := selector.FindNodes(context.Background(), "", nil)
	if err == nil {
		t.Error("expected error for empty node list")
	}
}

func TestRandomNodeSelectorGetNodeByID(t *testing.T) {
	nodes := makeNodes(3)
	selector := NewRandomNodeSelector(nodes)

	// Find existing node
	found, err := selector.GetNodeByID(context.Background(), nodes[1].ID)
	if err != nil {
		t.Fatalf("GetNodeByID() error: %v", err)
	}
	if found.ID != nodes[1].ID {
		t.Error("returned wrong node")
	}

	// Find nonexistent node
	_, err = selector.GetNodeByID(context.Background(), makeNodeID(99))
	if err == nil {
		t.Error("expected error for nonexistent node")
	}
}

// --- DHTNodeSelector score tests (pure logic) ---

func TestDHTNodeSelectorIsSameContinent(t *testing.T) {
	s := &DHTNodeSelector{}

	tests := []struct {
		r1, r2   string
		expected bool
	}{
		{"americas", "us-east", true},
		{"us-west", "sa-east", true},
		{"europe", "eu-west", true},
		{"eu-central", "eu-west", true},
		{"asia_pacific", "ap-southeast", true},
		{"americas", "europe", false},
		{"americas", "asia_pacific", false},
		{"unknown", "americas", false},
		{"americas", "unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.r1+"_"+tt.r2, func(t *testing.T) {
			result := s.isSameContinent(tt.r1, tt.r2)
			if result != tt.expected {
				t.Errorf("isSameContinent(%q, %q) = %v, want %v", tt.r1, tt.r2, result, tt.expected)
			}
		})
	}
}
