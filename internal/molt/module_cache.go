package molt

import (
	"context"
	"sync"
	"time"

	"github.com/moltbunker/moltbunker/internal/logging"
)

// cacheEntry wraps a CompiledMolt with access tracking for LRU eviction.
type cacheEntry struct {
	molt       *CompiledMolt
	lastAccess time.Time
}

// ModuleCache is an in-memory LRU cache of compiled WASM modules.
// Thread-safe via RWMutex. Evicted entries have their CompiledModule closed.
type ModuleCache struct {
	entries    map[string]*cacheEntry
	maxEntries int
	mu         sync.RWMutex
}

// NewModuleCache creates a cache with the given capacity.
func NewModuleCache(maxEntries int) *ModuleCache {
	if maxEntries <= 0 {
		maxEntries = 256
	}
	return &ModuleCache{
		entries:    make(map[string]*cacheEntry, maxEntries),
		maxEntries: maxEntries,
	}
}

// Get retrieves a compiled module by CID and updates its access time.
// Returns nil if not found.
func (c *ModuleCache) Get(cid string) *CompiledMolt {
	c.mu.RLock()
	entry, ok := c.entries[cid]
	c.mu.RUnlock()
	if !ok {
		return nil
	}

	// Update access time under write lock
	c.mu.Lock()
	entry.lastAccess = time.Now()
	c.mu.Unlock()

	return entry.molt
}

// Put adds a compiled module to the cache, evicting the oldest if full.
func (c *ModuleCache) Put(cid string, molt *CompiledMolt) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Already cached — update
	if existing, ok := c.entries[cid]; ok {
		existing.molt = molt
		existing.lastAccess = time.Now()
		return
	}

	// Evict oldest if at capacity
	if len(c.entries) >= c.maxEntries {
		c.evictOldest()
	}

	c.entries[cid] = &cacheEntry{
		molt:       molt,
		lastAccess: time.Now(),
	}
}

// Evict removes a specific entry by CID and closes its compiled module.
func (c *ModuleCache) Evict(cid string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if entry, ok := c.entries[cid]; ok {
		c.closeEntry(entry)
		delete(c.entries, cid)
	}
}

// Size returns the number of cached entries.
func (c *ModuleCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.entries)
}

// Close closes all cached compiled modules.
func (c *ModuleCache) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for cid, entry := range c.entries {
		c.closeEntry(entry)
		delete(c.entries, cid)
	}
}

// evictOldest removes the least recently accessed entry. Caller must hold write lock.
func (c *ModuleCache) evictOldest() {
	var oldestCID string
	var oldestTime time.Time

	for cid, entry := range c.entries {
		if oldestCID == "" || entry.lastAccess.Before(oldestTime) {
			oldestCID = cid
			oldestTime = entry.lastAccess
		}
	}

	if oldestCID != "" {
		c.closeEntry(c.entries[oldestCID])
		delete(c.entries, oldestCID)
		logging.Debug("molt cache evicted oldest entry", "cid", oldestCID)
	}
}

// closeEntry closes the compiled module, logging errors. Caller must hold lock.
func (c *ModuleCache) closeEntry(entry *cacheEntry) {
	if entry.molt != nil && entry.molt.Module != nil {
		if err := entry.molt.Module.Close(context.Background()); err != nil {
			logging.Warn("failed to close evicted molt module", "cid", entry.molt.CID, "error", err)
		}
	}
}
