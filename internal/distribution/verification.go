package distribution

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
)

// Verifier verifies image integrity using content hash
type Verifier struct{}

// NewVerifier creates a new verifier
func NewVerifier() *Verifier {
	return &Verifier{}
}

// VerifyImage verifies image integrity against expected CID
func (v *Verifier) VerifyImage(imagePath string, expectedCID string) error {
	file, err := os.Open(imagePath)
	if err != nil {
		return fmt.Errorf("failed to open image: %w", err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return fmt.Errorf("failed to hash image: %w", err)
	}

	actualHash := fmt.Sprintf("%x", hash.Sum(nil))

	// Compare with expected CID (simplified - actual IPFS CID is more complex)
	if actualHash != expectedCID {
		return fmt.Errorf("image integrity check failed: expected %s, got %s", expectedCID, actualHash)
	}

	return nil
}

// CalculateCID calculates content hash for an image
func (v *Verifier) CalculateCID(imagePath string) (string, error) {
	file, err := os.Open(imagePath)
	if err != nil {
		return "", fmt.Errorf("failed to open image: %w", err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("failed to hash image: %w", err)
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}
