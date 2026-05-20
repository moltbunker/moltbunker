package runtime

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/moltbunker/moltbunker/internal/logging"
	"github.com/moltbunker/moltbunker/internal/util"
)

// R15 — image GC: configurable schedule + size-bounded LRU eviction.
//
// The scheduler itself already existed (Start/Stop loop), but with two real
// gaps that this revision closes:
//
//   1. Interval was hardcoded to 1 hour. Now configurable via WithInterval.
//   2. The maxSize field was never consulted by CollectGarbage — so the GC
//      could never bound total image bytes. Now there's a Phase 2 LRU pass
//      that evicts oldest-by-lastUsed entries until total is under maxSize,
//      skipping images marked "active" within the last activeThreshold.

// defaultGCInterval is the fallback schedule when WithInterval is not called.
const defaultGCInterval = 1 * time.Hour

// defaultActiveThreshold is how recently an image must have been marked
// in-use to be considered "active" (and exempt from size-driven LRU eviction).
const defaultActiveThreshold = 5 * time.Minute

// ImageGC performs garbage collection of unused container images.
// It tracks which images are in use by running containers and periodically
// removes images that have not been referenced for longer than maxAge.
type ImageGC struct {
	imgMgr  *ImageManager
	maxAge  time.Duration // images unused longer than this are eligible for GC
	maxSize int64         // target max total image size in bytes (0 = unlimited)

	interval        time.Duration // schedule interval; 0 = defaultGCInterval
	activeThreshold time.Duration // images marked within this window are exempt from LRU eviction

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
		imgMgr:          imgMgr,
		maxAge:          maxAge,
		maxSize:         maxSize,
		interval:        defaultGCInterval,
		activeThreshold: defaultActiveThreshold,
		inUse:           make(map[string]time.Time),
		stopCh:          make(chan struct{}),
		nowFunc:         time.Now,
	}
}

// WithInterval sets the GC scheduler tick interval. Returns the receiver for
// chaining: `NewImageGC(...).WithInterval(15*time.Minute)`. Must be called
// before Start; setting after Start has no effect on the running ticker.
//
// A zero or negative interval is normalized to defaultGCInterval.
func (gc *ImageGC) WithInterval(d time.Duration) *ImageGC {
	gc.mu.Lock()
	defer gc.mu.Unlock()
	if d <= 0 {
		d = defaultGCInterval
	}
	gc.interval = d
	return gc
}

// WithActiveThreshold sets how recently an image must have been marked in-use
// to be considered "active" (and skipped during size-driven LRU eviction).
// Returns the receiver for chaining. Zero or negative is normalized to
// defaultActiveThreshold.
func (gc *ImageGC) WithActiveThreshold(d time.Duration) *ImageGC {
	gc.mu.Lock()
	defer gc.mu.Unlock()
	if d <= 0 {
		d = defaultActiveThreshold
	}
	gc.activeThreshold = d
	return gc
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
// and have been unused for longer than maxAge. When maxSize > 0 and total image
// bytes still exceed it after the expiry pass, a second LRU pass evicts the
// oldest-by-lastUsed entries until total is under the cap (skipping images
// marked "active" within activeThreshold).
func (gc *ImageGC) CollectGarbage(ctx context.Context) ([]string, error) {
	if gc.imgMgr == nil {
		return nil, fmt.Errorf("image manager is not available")
	}
	images, err := gc.imgMgr.ListImages(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list images: %w", err)
	}

	gc.mu.RLock()
	// Snapshot to avoid holding the lock during deletes (which may be slow).
	maxSize := gc.maxSize
	now := gc.nowFunc()
	activeThreshold := gc.activeThreshold
	if activeThreshold <= 0 {
		activeThreshold = defaultActiveThreshold
	}
	infos := make([]imageInfo, 0, len(images))
	for _, img := range images {
		ref := img.Name()
		lastUsed, exists := gc.inUse[ref]
		infos = append(infos, imageInfo{
			ref:      ref,
			lastUsed: lastUsed,
			inUse:    exists && now.Sub(lastUsed) < gc.maxAge,
		})
	}
	gc.mu.RUnlock()

	// Phase 1: expiry-based eviction.
	collected := make([]string, 0)
	survivors := make([]imageInfo, 0, len(infos))
	for _, info := range infos {
		if info.inUse {
			survivors = append(survivors, info)
			continue
		}
		if err := gc.imgMgr.DeleteImage(ctx, info.ref); err != nil {
			logging.Warn("image GC: failed to delete image",
				"image", info.ref,
				logging.Err(err))
			survivors = append(survivors, info)
			continue
		}
		collected = append(collected, info.ref)
		logging.Info("image GC: collected expired image", "image", info.ref)
	}

	// Phase 2: size-bounded LRU eviction.
	if maxSize > 0 && len(survivors) > 0 {
		// Recompute total size of survivors.
		var totalSize int64
		sizes := make(map[string]int64, len(survivors))
		for _, img := range images {
			ref := img.Name()
			// Skip those that were already evicted in phase 1.
			if !containsRef(survivors, ref) {
				continue
			}
			sz, sErr := img.Size(ctx)
			if sErr != nil {
				continue
			}
			sizes[ref] = sz
			totalSize += sz
		}

		extraEvictions := selectForEviction(survivors, sizes, totalSize, maxSize, activeThreshold, now)
		for _, ref := range extraEvictions {
			if err := gc.imgMgr.DeleteImage(ctx, ref); err != nil {
				logging.Warn("image GC: failed to evict for size cap",
					"image", ref,
					logging.Err(err))
				continue
			}
			collected = append(collected, ref)
			logging.Info("image GC: evicted for size cap (LRU)", "image", ref)
		}
	}

	return collected, nil
}

// imageInfo is the per-image snapshot used by CollectGarbage and the LRU
// helper. Promoted to a package-level type so selectForEviction can be
// independently testable.
type imageInfo struct {
	ref      string
	lastUsed time.Time
	inUse    bool
}

// selectForEviction is the pure helper that decides which images to evict on
// a size-bounded LRU pass. Pulled out as a free function so the eviction
// policy can be unit-tested without mocking containerd.Image.
//
// Returns the image refs to delete, in eviction order (oldest first). Images
// with lastUsed within activeThreshold of now are considered "active" and
// will NOT be selected even if eviction can't bring the total under maxSize.
// In that case we evict as many non-active candidates as we can and accept
// running over budget — better than killing a running container's image.
func selectForEviction(
	survivors []imageInfo,
	sizes map[string]int64,
	totalSize int64,
	maxSize int64,
	activeThreshold time.Duration,
	now time.Time,
) []string {
	if totalSize <= maxSize {
		return nil
	}

	// Candidates = survivors that are NOT in the active window.
	type cand struct {
		ref      string
		lastUsed time.Time
		size     int64
	}
	candidates := make([]cand, 0, len(survivors))
	for _, s := range survivors {
		// Active = marked AND within activeThreshold. Skip.
		if !s.lastUsed.IsZero() && now.Sub(s.lastUsed) < activeThreshold {
			continue
		}
		candidates = append(candidates, cand{
			ref:      s.ref,
			lastUsed: s.lastUsed,
			size:     sizes[s.ref],
		})
	}

	// Sort oldest first. Zero-time (never marked) sorts before any real time.
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].lastUsed.Before(candidates[j].lastUsed)
	})

	toEvict := make([]string, 0)
	current := totalSize
	for _, c := range candidates {
		if current <= maxSize {
			break
		}
		toEvict = append(toEvict, c.ref)
		current -= c.size
	}
	return toEvict
}

// containsRef reports whether the survivors list contains an entry with the
// given ref.
func containsRef(survivors []imageInfo, ref string) bool {
	for _, s := range survivors {
		if s.ref == ref {
			return true
		}
	}
	return false
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
// configured interval (default 1 hour; set with WithInterval). It returns
// immediately; the goroutine runs until Stop is called or ctx is cancelled.
func (gc *ImageGC) Start(ctx context.Context) {
	gc.mu.Lock()
	if gc.stopped {
		// Already stopped, recreate stop channel.
		gc.stopCh = make(chan struct{})
	}
	gc.stopped = false
	interval := gc.interval
	if interval <= 0 {
		interval = defaultGCInterval
	}
	gc.mu.Unlock()

	util.SafeGoWithName("image-gc", func() {
		ticker := time.NewTicker(interval)
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
