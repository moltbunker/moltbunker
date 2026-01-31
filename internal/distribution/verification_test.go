package distribution

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestVerifier_VerifyImage(t *testing.T) {
	v := NewVerifier()

	tmpDir := t.TempDir()
	imagePath := filepath.Join(tmpDir, "test-image.tar")

	testData := []byte("test image data")
	if err := os.WriteFile(imagePath, testData, 0644); err != nil {
		t.Fatalf("Failed to create test image: %v", err)
	}

	// Calculate expected CID
	hash := sha256.Sum256(testData)
	expectedCID := fmt.Sprintf("%x", hash)

	err := v.VerifyImage(imagePath, expectedCID)
	if err != nil {
		t.Fatalf("Image verification failed: %v", err)
	}
}

func TestVerifier_VerifyImage_WrongCID(t *testing.T) {
	v := NewVerifier()

	tmpDir := t.TempDir()
	imagePath := filepath.Join(tmpDir, "test-image.tar")

	testData := []byte("test image data")
	if err := os.WriteFile(imagePath, testData, 0644); err != nil {
		t.Fatalf("Failed to create test image: %v", err)
	}

	wrongCID := "wrong-cid-123"

	err := v.VerifyImage(imagePath, wrongCID)
	if err == nil {
		t.Error("Should fail verification with wrong CID")
	}
}

func TestVerifier_CalculateCID(t *testing.T) {
	v := NewVerifier()

	tmpDir := t.TempDir()
	imagePath := filepath.Join(tmpDir, "test-image.tar")

	testData := []byte("test image data")
	if err := os.WriteFile(imagePath, testData, 0644); err != nil {
		t.Fatalf("Failed to create test image: %v", err)
	}

	cid, err := v.CalculateCID(imagePath)
	if err != nil {
		t.Fatalf("Failed to calculate CID: %v", err)
	}

	if len(cid) == 0 {
		t.Error("CID should not be empty")
	}

	// Verify CID matches expected hash
	hash := sha256.Sum256(testData)
	expectedCID := fmt.Sprintf("%x", hash)

	if cid != expectedCID {
		t.Errorf("CID mismatch: got %s, want %s", cid, expectedCID)
	}
}

func TestVerifier_CalculateCID_NotExists(t *testing.T) {
	v := NewVerifier()

	_, err := v.CalculateCID("/nonexistent/path/image.tar")
	if err == nil {
		t.Error("Should fail for nonexistent file")
	}
}
