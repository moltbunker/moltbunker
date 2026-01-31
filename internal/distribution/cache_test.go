package distribution

import (
	"os"
	"path/filepath"
	"testing"
)

func TestImageCache_GetCachedImage_NotExists(t *testing.T) {
	tmpDir := t.TempDir()
	ic, err := NewImageCache(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create image cache: %v", err)
	}

	_, exists := ic.GetCachedImage("test-cid")
	if exists {
		t.Error("Image should not exist in cache")
	}
}

func TestImageCache_CacheImage(t *testing.T) {
	tmpDir := t.TempDir()
	ic, err := NewImageCache(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create image cache: %v", err)
	}

	cid := "test-cid-123"
	imagePath := filepath.Join(tmpDir, "test-image.tar")

	// Create test image file
	testData := []byte("test image data")
	if err := os.WriteFile(imagePath, testData, 0644); err != nil {
		t.Fatalf("Failed to create test image: %v", err)
	}

	err = ic.CacheImage(cid, imagePath)
	if err != nil {
		t.Fatalf("Failed to cache image: %v", err)
	}

	cachedPath, exists := ic.GetCachedImage(cid)
	if !exists {
		t.Fatal("Image should exist in cache")
	}

	// Verify cached file exists and has correct content
	cachedData, err := os.ReadFile(cachedPath)
	if err != nil {
		t.Fatalf("Failed to read cached image: %v", err)
	}

	if string(cachedData) != string(testData) {
		t.Error("Cached image data doesn't match original")
	}
}

func TestImageCache_RemoveCachedImage(t *testing.T) {
	tmpDir := t.TempDir()
	ic, err := NewImageCache(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create image cache: %v", err)
	}

	cid := "test-cid-123"
	imagePath := filepath.Join(tmpDir, "test-image.tar")

	testData := []byte("test image data")
	if err := os.WriteFile(imagePath, testData, 0644); err != nil {
		t.Fatalf("Failed to create test image: %v", err)
	}

	ic.CacheImage(cid, imagePath)

	err = ic.RemoveCachedImage(cid)
	if err != nil {
		t.Fatalf("Failed to remove cached image: %v", err)
	}

	_, exists := ic.GetCachedImage(cid)
	if exists {
		t.Error("Image should not exist in cache after removal")
	}
}

func TestImageCache_RemoveCachedImage_NotExists(t *testing.T) {
	tmpDir := t.TempDir()
	ic, err := NewImageCache(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create image cache: %v", err)
	}

	err = ic.RemoveCachedImage("nonexistent-cid")
	if err != nil {
		t.Fatalf("Should not fail when removing nonexistent image: %v", err)
	}
}
