package runtime

import (
	"sync"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// ImageGC unit tests
// ---------------------------------------------------------------------------

func TestNewImageGC(t *testing.T) {
	t.Run("creates GC with correct fields", func(t *testing.T) {
		gc := NewImageGC(nil, 24*time.Hour, 10*1024*1024*1024)
		if gc == nil {
			t.Fatal("expected non-nil ImageGC")
		}
		if gc.maxAge != 24*time.Hour {
			t.Errorf("maxAge = %v, want 24h", gc.maxAge)
		}
		if gc.maxSize != 10*1024*1024*1024 {
			t.Errorf("maxSize = %d, want 10GB", gc.maxSize)
		}
		if gc.inUse == nil {
			t.Error("inUse map should be initialized")
		}
		if gc.stopCh == nil {
			t.Error("stopCh should be initialized")
		}
	})

	t.Run("zero maxSize means unlimited", func(t *testing.T) {
		gc := NewImageGC(nil, time.Hour, 0)
		if gc.maxSize != 0 {
			t.Errorf("maxSize = %d, want 0", gc.maxSize)
		}
	})
}

func TestImageGC_MarkAndUnmark(t *testing.T) {
	gc := NewImageGC(nil, time.Hour, 0)

	t.Run("mark puts entry in map", func(t *testing.T) {
		gc.MarkInUse("nginx:latest")

		gc.mu.RLock()
		_, exists := gc.inUse["nginx:latest"]
		gc.mu.RUnlock()

		if !exists {
			t.Error("image should be in inUse map after MarkInUse")
		}
	})

	t.Run("isInUse returns true for recently marked image", func(t *testing.T) {
		gc.mu.RLock()
		inUse := gc.isInUse("nginx:latest")
		gc.mu.RUnlock()

		if !inUse {
			t.Error("recently marked image should be in use")
		}
	})

	t.Run("isInUse returns false for unknown image", func(t *testing.T) {
		gc.mu.RLock()
		inUse := gc.isInUse("unknown:image")
		gc.mu.RUnlock()

		if inUse {
			t.Error("unknown image should not be in use")
		}
	})

	t.Run("unmark resets timestamp to now", func(t *testing.T) {
		gc.UnmarkInUse("nginx:latest")

		gc.mu.RLock()
		lastUsed, exists := gc.inUse["nginx:latest"]
		gc.mu.RUnlock()

		if !exists {
			t.Error("image should still be in map after UnmarkInUse")
		}
		// Should have been updated to approximately now.
		if time.Since(lastUsed) > time.Second {
			t.Error("lastUsed should be recent after UnmarkInUse")
		}
	})
}

func TestImageGC_InUseExpiry(t *testing.T) {
	gc := NewImageGC(nil, 100*time.Millisecond, 0)

	// Use a controllable time function.
	now := time.Now()
	gc.nowFunc = func() time.Time { return now }

	gc.MarkInUse("alpine:3.18")

	// Should be in use right now.
	gc.mu.RLock()
	if !gc.isInUse("alpine:3.18") {
		t.Error("image should be in use immediately after marking")
	}
	gc.mu.RUnlock()

	// Advance time past maxAge.
	now = now.Add(200 * time.Millisecond)

	gc.mu.RLock()
	if gc.isInUse("alpine:3.18") {
		t.Error("image should no longer be in use after maxAge")
	}
	gc.mu.RUnlock()
}

func TestImageGC_InUseCount(t *testing.T) {
	gc := NewImageGC(nil, time.Hour, 0)

	if gc.InUseCount() != 0 {
		t.Errorf("expected 0 in-use images, got %d", gc.InUseCount())
	}

	gc.MarkInUse("image-a")
	gc.MarkInUse("image-b")
	gc.MarkInUse("image-c")

	if gc.InUseCount() != 3 {
		t.Errorf("expected 3 in-use images, got %d", gc.InUseCount())
	}

	// Expire one by manipulating the map directly.
	gc.mu.Lock()
	gc.inUse["image-b"] = time.Now().Add(-2 * time.Hour)
	gc.mu.Unlock()

	if gc.InUseCount() != 2 {
		t.Errorf("expected 2 in-use images after expiry, got %d", gc.InUseCount())
	}
}

func TestImageGC_InUseCountWithControlledTime(t *testing.T) {
	gc := NewImageGC(nil, 50*time.Millisecond, 0)

	now := time.Now()
	gc.nowFunc = func() time.Time { return now }

	gc.MarkInUse("img-1")
	gc.MarkInUse("img-2")

	if gc.InUseCount() != 2 {
		t.Errorf("expected 2, got %d", gc.InUseCount())
	}

	// Advance past maxAge.
	now = now.Add(100 * time.Millisecond)

	if gc.InUseCount() != 0 {
		t.Errorf("expected 0 after expiry, got %d", gc.InUseCount())
	}
}

func TestImageGC_StopIdempotent(t *testing.T) {
	gc := NewImageGC(nil, time.Hour, 0)

	gc.Stop()
	gc.Stop() // second stop should not panic
}

func TestImageGC_StopBeforeStart(t *testing.T) {
	gc := NewImageGC(nil, time.Hour, 0)
	// Stopping before starting should be safe.
	gc.Stop()
}

func TestImageGC_ConcurrentMarkUnmark(t *testing.T) {
	gc := NewImageGC(nil, time.Hour, 0)

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(2)
		ref := "concurrent-" + string(rune('a'+i%26))
		go func(r string) {
			defer wg.Done()
			gc.MarkInUse(r)
		}(ref)
		go func(r string) {
			defer wg.Done()
			gc.UnmarkInUse(r)
		}(ref)
	}
	wg.Wait()

	// Just verify no panics or data races occurred.
	count := gc.InUseCount()
	if count < 0 {
		t.Errorf("InUseCount should be non-negative, got %d", count)
	}
}

func TestImageGC_MarkInUseUpdatesTimestamp(t *testing.T) {
	gc := NewImageGC(nil, time.Hour, 0)

	now := time.Now()
	gc.nowFunc = func() time.Time { return now }

	gc.MarkInUse("test-image")

	gc.mu.RLock()
	ts1 := gc.inUse["test-image"]
	gc.mu.RUnlock()

	// Advance time and mark again.
	now = now.Add(5 * time.Minute)
	gc.MarkInUse("test-image")

	gc.mu.RLock()
	ts2 := gc.inUse["test-image"]
	gc.mu.RUnlock()

	if !ts2.After(ts1) {
		t.Error("MarkInUse should update the timestamp")
	}
}

func TestImageGC_CollectGarbageNilManager(t *testing.T) {
	// CollectGarbage with nil ImageManager should return an error.
	gc := NewImageGC(nil, time.Hour, 0)
	_, err := gc.CollectGarbage(nil)
	if err == nil {
		t.Error("expected error when ImageManager is nil")
	}
}

func TestImageGC_GetImageUsageNilManager(t *testing.T) {
	// GetImageUsage with nil ImageManager should return an error.
	gc := NewImageGC(nil, time.Hour, 0)
	_, err := gc.GetImageUsage(nil)
	if err == nil {
		t.Error("expected error when ImageManager is nil")
	}
}

func TestImageGC_PreservesInUseImages(t *testing.T) {
	// This is a logic test without a real ImageManager.
	// We verify that isInUse correctly protects images.
	gc := NewImageGC(nil, time.Hour, 0)

	gc.MarkInUse("keep-me:latest")
	gc.MarkInUse("keep-me-too:v1")

	gc.mu.RLock()
	defer gc.mu.RUnlock()

	if !gc.isInUse("keep-me:latest") {
		t.Error("marked image should be in use")
	}
	if !gc.isInUse("keep-me-too:v1") {
		t.Error("marked image should be in use")
	}
	if gc.isInUse("delete-me:old") {
		t.Error("unmarked image should not be in use")
	}
}

func TestImageGC_UnmarkDoesNotRemoveFromMap(t *testing.T) {
	gc := NewImageGC(nil, time.Hour, 0)

	gc.MarkInUse("img-x")
	gc.UnmarkInUse("img-x")

	gc.mu.RLock()
	_, exists := gc.inUse["img-x"]
	gc.mu.RUnlock()

	if !exists {
		t.Error("UnmarkInUse should keep the entry in the map (with updated timestamp)")
	}
}

func TestImageGC_UnmarkUnknownImageIsSafe(t *testing.T) {
	gc := NewImageGC(nil, time.Hour, 0)

	// Should not panic.
	gc.UnmarkInUse("never-marked")

	gc.mu.RLock()
	_, exists := gc.inUse["never-marked"]
	gc.mu.RUnlock()

	if !exists {
		t.Error("UnmarkInUse on unknown should create an entry with current time")
	}
}
