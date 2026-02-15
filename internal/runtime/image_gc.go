package runtime

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/moltbunker/moltbunker/internal/logging"
	"github.com/moltbunker/moltbunker/internal/util"
)

// ImageGC performs garbage collection of unused container images.
// It tracks which images are in use by running containers and periodically
// removes images that have not been referenced for longer than maxAge.
type ImageGC struct {
	imgMgr  *ImageManager
	maxAge  time.Duration // images unused longer than this are eligible for GC
	maxSize int64         // target max total image size in bytes (0 = unlimited)

	inUse   map[string]time.Time // imageRef -> last time marked in-use
	mu      sync.RWMutex
	stopCh  chan struct{}
	stopped bool
	nowFunc func() time.Time // for testable time
}

// NewImageGC creates a new image garbage collector.
// maxAge controls how long an unused image survives before being eligible for
// collection. maxSize is the target maximum total size in bytes (0 means no
// size-based eviction).
func NewImageGC(imgMgr *ImageManager, maxAge time.Duration, maxSize int64) *ImageGC {
	return &ImageGC{
		imgMgr:  imgMgr,
		maxAge:  maxAge,
		maxSize: maxSize,
		inUse:   make(map[string]time.Time),
		stopCh:  make(chan struct{}),
		nowFunc: time.Now,
	}
}

// MarkInUse marks an image reference as actively in use by a container.
// As long as an image is marked in use, it will not be collected.
func (gc *ImageGC) MarkInUse(imageRef string) {
	gc.mu.Lock()
	defer gc.mu.Unlock()
	gc.inUse[imageRef] = gc.nowFunc()
}

// UnmarkInUse removes the in-use marker for an image reference.
// The image becomes eligible for collection after maxAge elapses.
func (gc *ImageGC) UnmarkInUse(imageRef string) {
	gc.mu.Lock()
	defer gc.mu.Unlock()
	// Keep the timestamp so we know when it was last used,
	// but set it to now so the maxAge clock starts ticking.
	gc.inUse[imageRef] = gc.nowFunc()
}

// isInUse returns true if the image is currently marked as in-use and has been
// referenced more recently than maxAge. It must be called with at least a read
// lock held.
func (gc *ImageGC) isInUse(imageRef string) bool {
	lastUsed, exists := gc.inUse[imageRef]
	if !exists {
		return false
	}
	return gc.nowFunc().Sub(lastUsed) < gc.maxAge
}

// CollectGarbage removes images that are not referenced by any running container
// and have been unused for longer than maxAge.
func (gc *ImageGC) CollectGarbage(ctx context.Context) ([]string, error) {
	if gc.imgMgr == nil {
		return nil, fmt.Errorf("image manager is not available")
	}
	images, err := gc.imgMgr.ListImages(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list images: %w", err)
	}

	gc.mu.RLock()
	defer gc.mu.RUnlock()

	var collected []string
	for _, img := range images {
		ref := img.Name()
		if gc.isInUse(ref) {
			continue
		}

		// Image is not in use or has expired -- delete it.
		if err := gc.imgMgr.DeleteImage(ctx, ref); err != nil {
			logging.Warn("image GC: failed to delete image",
				"image", ref,
				logging.Err(err))
			continue
		}
		collected = append(collected, ref)
		logging.Info("image GC: collected image", "image", ref)
	}

	return collected, nil
}

// PruneExpired removes stale entries from the inUse map that have expired
// beyond maxAge. This prevents unbounded growth on long-running nodes.
func (gc *ImageGC) PruneExpired() int {
	gc.mu.Lock()
	defer gc.mu.Unlock()

	now := gc.nowFunc()
	pruned := 0
	for ref, lastUsed := range gc.inUse {
		if now.Sub(lastUsed) >= gc.maxAge {
			delete(gc.inUse, ref)
			pruned++
		}
	}
	return pruned
}

// GetImageUsage returns the total size of all stored images in bytes.
func (gc *ImageGC) GetImageUsage(ctx context.Context) (int64, error) {
	if gc.imgMgr == nil {
		return 0, fmt.Errorf("image manager is not available")
	}
	images, err := gc.imgMgr.ListImages(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to list images: %w", err)
	}

	var total int64
	for _, img := range images {
		size, err := img.Size(ctx)
		if err != nil {
			// Skip images whose size cannot be determined.
			continue
		}
		total += size
	}
	return total, nil
}

// InUseCount returns the number of images currently marked as in use.
func (gc *ImageGC) InUseCount() int {
	gc.mu.RLock()
	defer gc.mu.RUnlock()

	count := 0
	for _, lastUsed := range gc.inUse {
		if gc.nowFunc().Sub(lastUsed) < gc.maxAge {
			count++
		}
	}
	return count
}

// Start begins a background goroutine that runs garbage collection every
// interval (default 1 hour). It returns immediately; the goroutine runs
// until Stop is called or ctx is cancelled.
func (gc *ImageGC) Start(ctx context.Context) {
	gc.mu.Lock()
	if gc.stopped {
		// Already stopped, recreate stop channel.
		gc.stopCh = make(chan struct{})
	}
	gc.stopped = false
	gc.mu.Unlock()

	util.SafeGoWithName("image-gc", func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-gc.stopCh:
				return
			case <-ticker.C:
				collected, err := gc.CollectGarbage(ctx)
				if err != nil {
					logging.Warn("image GC cycle failed", logging.Err(err))
				} else if len(collected) > 0 {
					logging.Info("image GC cycle completed",
						"collected", len(collected))
				}
				// Prune expired tracking entries to prevent unbounded map growth
				gc.PruneExpired()
			}
		}
	})
}

// Stop stops the background GC goroutine. It is safe to call multiple times.
func (gc *ImageGC) Stop() {
	gc.mu.Lock()
	defer gc.mu.Unlock()
	if !gc.stopped {
		gc.stopped = true
		close(gc.stopCh)
	}
}
