package distribution

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// ImageCache manages local image cache
type ImageCache struct {
	cacheDir string
	cache    map[string]string // CID -> local path
	mu       sync.RWMutex
}

// NewImageCache creates a new image cache
func NewImageCache(cacheDir string) (*ImageCache, error) {
	if err := os.MkdirAll(cacheDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	return &ImageCache{
		cacheDir: cacheDir,
		cache:    make(map[string]string),
	}, nil
}

// GetCachedImage returns cached image path if exists
func (ic *ImageCache) GetCachedImage(cid string) (string, bool) {
	ic.mu.RLock()
	defer ic.mu.RUnlock()

	path, exists := ic.cache[cid]
	return path, exists
}

// CacheImage caches an image locally
func (ic *ImageCache) CacheImage(cid string, imagePath string) error {
	ic.mu.Lock()
	defer ic.mu.Unlock()

	// Copy image to cache directory. Sanitize the CID to a bare filename so it
	// cannot escape the cache directory via path separators or "..".
	cachePath := filepath.Join(ic.cacheDir, filepath.Base(cid))

	// Read source file
	// #nosec G304 -- imagePath is an internally managed image file path, not request input
	data, err := os.ReadFile(imagePath)
	if err != nil {
		return fmt.Errorf("failed to read image: %w", err)
	}

	// Write to cache
	// #nosec G304 G703 -- cachePath is cacheDir joined with filepath.Base(cid); traversal not possible
	if err := os.WriteFile(cachePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write cache: %w", err)
	}

	ic.cache[cid] = cachePath
	return nil
}

// RemoveCachedImage removes a cached image
func (ic *ImageCache) RemoveCachedImage(cid string) error {
	ic.mu.Lock()
	defer ic.mu.Unlock()

	path, exists := ic.cache[cid]
	if !exists {
		return nil
	}

	if err := os.Remove(path); err != nil {
		return fmt.Errorf("failed to remove cached image: %w", err)
	}

	delete(ic.cache, cid)
	return nil
}
