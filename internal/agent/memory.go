package agent

import (
	"fmt"
	"sync"
	"time"
)

// MemoryStore provides persistent key-value memory for agents.
// In production, this delegates to Object Storage. This implementation
// provides an in-memory version for testing and local development.
type MemoryStore struct {
	mu      sync.RWMutex
	entries map[string]map[string]*MemoryEntry // agentID → key → entry
}

// NewMemoryStore creates a new in-memory agent memory store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		entries: make(map[string]map[string]*MemoryEntry),
	}
}

// Put stores a memory entry for an agent.
func (m *MemoryStore) Put(agentID, key, value string) error {
	if agentID == "" || key == "" {
		return fmt.Errorf("agent ID and key are required")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	bucket, ok := m.entries[agentID]
	if !ok {
		bucket = make(map[string]*MemoryEntry)
		m.entries[agentID] = bucket
	}

	bucket[key] = &MemoryEntry{
		Key:       key,
		Value:     value,
		UpdatedAt: time.Now(),
	}

	return nil
}

// Get retrieves a memory entry for an agent.
func (m *MemoryStore) Get(agentID, key string) (*MemoryEntry, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	bucket, ok := m.entries[agentID]
	if !ok {
		return nil, fmt.Errorf("no memory for agent %q", agentID)
	}

	entry, ok := bucket[key]
	if !ok {
		return nil, fmt.Errorf("key %q not found in agent %q memory", key, agentID)
	}

	cp := *entry
	return &cp, nil
}

// Delete removes a memory entry.
func (m *MemoryStore) Delete(agentID, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	bucket, ok := m.entries[agentID]
	if !ok {
		return nil
	}

	delete(bucket, key)
	return nil
}

// List returns all memory entries for an agent.
func (m *MemoryStore) List(agentID string) []MemoryEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()

	bucket, ok := m.entries[agentID]
	if !ok {
		return nil
	}

	result := make([]MemoryEntry, 0, len(bucket))
	for _, entry := range bucket {
		cp := *entry
		result = append(result, cp)
	}
	return result
}

// Clear removes all memory for an agent.
func (m *MemoryStore) Clear(agentID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.entries, agentID)
}

// Size returns the number of entries for an agent.
func (m *MemoryStore) Size(agentID string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.entries[agentID])
}
