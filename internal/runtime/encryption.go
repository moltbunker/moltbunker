package runtime

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// EncryptionManager manages encrypted volumes for containers
type EncryptionManager struct {
	dataDir string
}

// NewEncryptionManager creates a new encryption manager
func NewEncryptionManager(dataDir string) *EncryptionManager {
	return &EncryptionManager{
		dataDir: dataDir,
	}
}

// CreateEncryptedVolume creates an encrypted LUKS volume for a container
func (em *EncryptionManager) CreateEncryptedVolume(containerID string, sizeGB int, key []byte) (string, error) {
	// Create volume file
	volumePath := filepath.Join(em.dataDir, fmt.Sprintf("%s.vol", containerID))
	volumeSize := int64(sizeGB) * 1024 * 1024 * 1024 // Convert GB to bytes

	// Create sparse file
	file, err := os.Create(volumePath)
	if err != nil {
		return "", fmt.Errorf("failed to create volume file: %w", err)
	}

	if err := file.Truncate(volumeSize); err != nil {
		file.Close()
		return "", fmt.Errorf("failed to truncate volume file: %w", err)
	}
	file.Close()

	// Create LUKS encrypted volume
	// This requires cryptsetup to be installed
	cmd := exec.Command("cryptsetup", "luksFormat", volumePath)
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to create LUKS volume: %w", err)
	}

	// Open encrypted volume
	mapperName := fmt.Sprintf("moltbunker-%s", containerID)
	cmd = exec.Command("cryptsetup", "open", volumePath, mapperName)
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to open LUKS volume: %w", err)
	}

	mapperPath := filepath.Join("/dev/mapper", mapperName)
	return mapperPath, nil
}

// DeleteEncryptedVolume deletes an encrypted volume
func (em *EncryptionManager) DeleteEncryptedVolume(containerID string) error {
	mapperName := fmt.Sprintf("moltbunker-%s", containerID)

	// Close encrypted volume
	cmd := exec.Command("cryptsetup", "close", mapperName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to close LUKS volume: %w", err)
	}

	// Delete volume file
	volumePath := filepath.Join(em.dataDir, fmt.Sprintf("%s.vol", containerID))
	if err := os.Remove(volumePath); err != nil {
		return fmt.Errorf("failed to delete volume file: %w", err)
	}

	return nil
}
