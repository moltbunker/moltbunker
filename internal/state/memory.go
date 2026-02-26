package state

import (
	"context"
	"fmt"
	"sync"
)

// MemoryStore is an in-memory implementation of StateStore for testing.
// All data is lost when the process exits.
type MemoryStore struct {
	mu      sync.RWMutex
	buckets map[string]map[string][]byte
	closed  bool
}

// NewMemoryStore creates a new in-memory state store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		buckets: map[string]map[string][]byte{
			BucketMeta:           {},
			BucketDeployments:    {},
			BucketBans:           {},
			BucketPeers:          {},
			BucketCertPins:       {},
			BucketAPIKeys:        {},
			BucketStorageBuckets: {},
			BucketStorageObjects: {},
			BucketProxySessions:  {},
			BucketCrawlJobs:      {},
			BucketAgentState:     {},
		},
	}
}

func (m *MemoryStore) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	return nil
}

func (m *MemoryStore) put(bucket, key string, data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return fmt.Errorf("store is closed")
	}
	b := m.buckets[bucket]
	if b == nil {
		return fmt.Errorf("unknown bucket: %s", bucket)
	}
	cp := make([]byte, len(data))
	copy(cp, data)
	b[key] = cp
	return nil
}

func (m *MemoryStore) get(bucket, key string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.closed {
		return nil, fmt.Errorf("store is closed")
	}
	b := m.buckets[bucket]
	if b == nil {
		return nil, fmt.Errorf("unknown bucket: %s", bucket)
	}
	data, ok := b[key]
	if !ok {
		return nil, nil
	}
	cp := make([]byte, len(data))
	copy(cp, data)
	return cp, nil
}

func (m *MemoryStore) del(bucket, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return fmt.Errorf("store is closed")
	}
	b := m.buckets[bucket]
	if b == nil {
		return fmt.Errorf("unknown bucket: %s", bucket)
	}
	delete(b, key)
	return nil
}

func (m *MemoryStore) list(bucket string) (map[string][]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.closed {
		return nil, fmt.Errorf("store is closed")
	}
	b := m.buckets[bucket]
	if b == nil {
		return nil, fmt.Errorf("unknown bucket: %s", bucket)
	}
	result := make(map[string][]byte, len(b))
	for k, v := range b {
		cp := make([]byte, len(v))
		copy(cp, v)
		result[k] = cp
	}
	return result, nil
}

// --- Deployments ---

func (m *MemoryStore) PutDeployment(_ context.Context, id string, data []byte) error {
	return m.put(BucketDeployments, id, data)
}

func (m *MemoryStore) GetDeployment(_ context.Context, id string) ([]byte, error) {
	return m.get(BucketDeployments, id)
}

func (m *MemoryStore) DeleteDeployment(_ context.Context, id string) error {
	return m.del(BucketDeployments, id)
}

func (m *MemoryStore) ListDeployments(_ context.Context) (map[string][]byte, error) {
	return m.list(BucketDeployments)
}

// --- Bans ---

func (m *MemoryStore) PutBan(_ context.Context, peerID string, data []byte) error {
	return m.put(BucketBans, peerID, data)
}

func (m *MemoryStore) DeleteBan(_ context.Context, peerID string) error {
	return m.del(BucketBans, peerID)
}

func (m *MemoryStore) ListBans(_ context.Context) (map[string][]byte, error) {
	return m.list(BucketBans)
}

// --- Peers ---

func (m *MemoryStore) PutPeer(_ context.Context, peerID string, data []byte) error {
	return m.put(BucketPeers, peerID, data)
}

func (m *MemoryStore) DeletePeer(_ context.Context, peerID string) error {
	return m.del(BucketPeers, peerID)
}

func (m *MemoryStore) ListPeers(_ context.Context) (map[string][]byte, error) {
	return m.list(BucketPeers)
}

// --- Certificate Pins ---

func (m *MemoryStore) PutCertPin(_ context.Context, nodeID string, hash []byte) error {
	return m.put(BucketCertPins, nodeID, hash)
}

func (m *MemoryStore) DeleteCertPin(_ context.Context, nodeID string) error {
	return m.del(BucketCertPins, nodeID)
}

func (m *MemoryStore) ListCertPins(_ context.Context) (map[string][]byte, error) {
	return m.list(BucketCertPins)
}

// --- API Keys ---

func (m *MemoryStore) PutAPIKey(_ context.Context, id string, data []byte) error {
	return m.put(BucketAPIKeys, id, data)
}

func (m *MemoryStore) DeleteAPIKey(_ context.Context, id string) error {
	return m.del(BucketAPIKeys, id)
}

func (m *MemoryStore) ListAPIKeys(_ context.Context) (map[string][]byte, error) {
	return m.list(BucketAPIKeys)
}

// --- Storage Buckets ---

func (m *MemoryStore) PutStorageBucket(_ context.Context, name string, data []byte) error {
	return m.put(BucketStorageBuckets, name, data)
}

func (m *MemoryStore) GetStorageBucket(_ context.Context, name string) ([]byte, error) {
	return m.get(BucketStorageBuckets, name)
}

func (m *MemoryStore) DeleteStorageBucket(_ context.Context, name string) error {
	return m.del(BucketStorageBuckets, name)
}

func (m *MemoryStore) ListStorageBuckets(_ context.Context) (map[string][]byte, error) {
	return m.list(BucketStorageBuckets)
}

// --- Storage Objects ---

func (m *MemoryStore) PutStorageObject(_ context.Context, key string, data []byte) error {
	return m.put(BucketStorageObjects, key, data)
}

func (m *MemoryStore) GetStorageObject(_ context.Context, key string) ([]byte, error) {
	return m.get(BucketStorageObjects, key)
}

func (m *MemoryStore) DeleteStorageObject(_ context.Context, key string) error {
	return m.del(BucketStorageObjects, key)
}

func (m *MemoryStore) ListStorageObjects(_ context.Context) (map[string][]byte, error) {
	return m.list(BucketStorageObjects)
}

// --- Proxy Sessions ---

func (m *MemoryStore) PutProxySession(_ context.Context, id string, data []byte) error {
	return m.put(BucketProxySessions, id, data)
}

func (m *MemoryStore) GetProxySession(_ context.Context, id string) ([]byte, error) {
	return m.get(BucketProxySessions, id)
}

func (m *MemoryStore) DeleteProxySession(_ context.Context, id string) error {
	return m.del(BucketProxySessions, id)
}

func (m *MemoryStore) ListProxySessions(_ context.Context) (map[string][]byte, error) {
	return m.list(BucketProxySessions)
}

// --- Crawl Jobs ---

func (m *MemoryStore) PutCrawlJob(_ context.Context, id string, data []byte) error {
	return m.put(BucketCrawlJobs, id, data)
}

func (m *MemoryStore) GetCrawlJob(_ context.Context, id string) ([]byte, error) {
	return m.get(BucketCrawlJobs, id)
}

func (m *MemoryStore) DeleteCrawlJob(_ context.Context, id string) error {
	return m.del(BucketCrawlJobs, id)
}

func (m *MemoryStore) ListCrawlJobs(_ context.Context) (map[string][]byte, error) {
	return m.list(BucketCrawlJobs)
}

// --- Agent State ---

func (m *MemoryStore) PutAgentState(_ context.Context, id string, data []byte) error {
	return m.put(BucketAgentState, id, data)
}

func (m *MemoryStore) GetAgentState(_ context.Context, id string) ([]byte, error) {
	return m.get(BucketAgentState, id)
}

func (m *MemoryStore) DeleteAgentState(_ context.Context, id string) error {
	return m.del(BucketAgentState, id)
}

func (m *MemoryStore) ListAgentState(_ context.Context) (map[string][]byte, error) {
	return m.list(BucketAgentState)
}

// --- Schema ---

func (m *MemoryStore) SchemaVersion(_ context.Context) (int, error) {
	data, err := m.get(BucketMeta, MetaSchemaVersion)
	if err != nil {
		return 0, err
	}
	if data == nil {
		return 0, nil
	}
	var v int
	_, err = fmt.Sscanf(string(data), "%d", &v)
	return v, err
}

func (m *MemoryStore) SetSchemaVersion(_ context.Context, version int) error {
	return m.put(BucketMeta, MetaSchemaVersion, []byte(fmt.Sprintf("%d", version)))
}

// --- Metadata ---

func (m *MemoryStore) PutMeta(_ context.Context, key string, data []byte) error {
	return m.put(BucketMeta, key, data)
}

func (m *MemoryStore) GetMeta(_ context.Context, key string) ([]byte, error) {
	return m.get(BucketMeta, key)
}
