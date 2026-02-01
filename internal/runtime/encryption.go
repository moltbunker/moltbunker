package runtime

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

// EncryptionManager manages encrypted volumes for containers
type EncryptionManager struct {
	dataDir     string
	volumesDir  string
	mountsDir   string
	keysDir     string
	volumes     map[string]*EncryptedVolume
	mu          sync.RWMutex
}

// EncryptedVolume represents an encrypted container volume
type EncryptedVolume struct {
	ContainerID string
	VolumePath  string
	MapperName  string
	MapperPath  string
	MountPath   string
	SizeGB      int
	Mounted     bool
	KeyPath     string
}

// NewEncryptionManager creates a new encryption manager
func NewEncryptionManager(dataDir string) (*EncryptionManager, error) {
	volumesDir := filepath.Join(dataDir, "volumes")
	mountsDir := filepath.Join(dataDir, "mounts")
	keysDir := filepath.Join(dataDir, "keys")

	// Create directories
	for _, dir := range []string{volumesDir, mountsDir, keysDir} {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return &EncryptionManager{
		dataDir:    dataDir,
		volumesDir: volumesDir,
		mountsDir:  mountsDir,
		keysDir:    keysDir,
		volumes:    make(map[string]*EncryptedVolume),
	}, nil
}

// GenerateKey generates a random encryption key
func (em *EncryptionManager) GenerateKey() ([]byte, error) {
	key := make([]byte, 32) // 256-bit key
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("failed to generate key: %w", err)
	}
	return key, nil
}

// CreateEncryptedVolume creates an encrypted LUKS volume for a container
func (em *EncryptionManager) CreateEncryptedVolume(containerID string, sizeGB int, key []byte) (*EncryptedVolume, error) {
	em.mu.Lock()
	defer em.mu.Unlock()

	if _, exists := em.volumes[containerID]; exists {
		return nil, fmt.Errorf("volume already exists for container: %s", containerID)
	}

	volumePath := filepath.Join(em.volumesDir, fmt.Sprintf("%s.img", containerID))
	keyPath := filepath.Join(em.keysDir, fmt.Sprintf("%s.key", containerID))
	mapperName := fmt.Sprintf("moltbunker-%s", containerID)
	mapperPath := filepath.Join("/dev/mapper", mapperName)
	mountPath := filepath.Join(em.mountsDir, containerID)

	// Save key to file (with restricted permissions)
	if err := os.WriteFile(keyPath, key, 0400); err != nil {
		return nil, fmt.Errorf("failed to save key: %w", err)
	}

	// Create sparse file for volume
	volumeSize := int64(sizeGB) * 1024 * 1024 * 1024
	file, err := os.Create(volumePath)
	if err != nil {
		os.Remove(keyPath)
		return nil, fmt.Errorf("failed to create volume file: %w", err)
	}

	if err := file.Truncate(volumeSize); err != nil {
		file.Close()
		os.Remove(keyPath)
		os.Remove(volumePath)
		return nil, fmt.Errorf("failed to truncate volume file: %w", err)
	}
	file.Close()

	// Format as LUKS with key file (non-interactive)
	cmd := exec.Command("cryptsetup", "luksFormat",
		"--type", "luks2",
		"--key-file", keyPath,
		"--batch-mode",
		volumePath,
	)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		os.Remove(keyPath)
		os.Remove(volumePath)
		return nil, fmt.Errorf("failed to create LUKS volume: %w (%s)", err, stderr.String())
	}

	volume := &EncryptedVolume{
		ContainerID: containerID,
		VolumePath:  volumePath,
		MapperName:  mapperName,
		MapperPath:  mapperPath,
		MountPath:   mountPath,
		SizeGB:      sizeGB,
		Mounted:     false,
		KeyPath:     keyPath,
	}

	em.volumes[containerID] = volume
	return volume, nil
}

// OpenVolume opens an encrypted volume
func (em *EncryptionManager) OpenVolume(containerID string) error {
	em.mu.Lock()
	defer em.mu.Unlock()

	volume, exists := em.volumes[containerID]
	if !exists {
		return fmt.Errorf("volume not found: %s", containerID)
	}

	// Check if already open
	if _, err := os.Stat(volume.MapperPath); err == nil {
		return nil // Already open
	}

	// Open LUKS volume
	cmd := exec.Command("cryptsetup", "open",
		"--key-file", volume.KeyPath,
		volume.VolumePath,
		volume.MapperName,
	)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to open LUKS volume: %w (%s)", err, stderr.String())
	}

	return nil
}

// FormatVolume formats an opened volume with ext4
func (em *EncryptionManager) FormatVolume(containerID string) error {
	em.mu.RLock()
	volume, exists := em.volumes[containerID]
	em.mu.RUnlock()

	if !exists {
		return fmt.Errorf("volume not found: %s", containerID)
	}

	// Format with ext4
	cmd := exec.Command("mkfs.ext4", "-F", volume.MapperPath)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to format volume: %w (%s)", err, stderr.String())
	}

	return nil
}

// MountVolume mounts an encrypted volume
func (em *EncryptionManager) MountVolume(containerID string) (string, error) {
	em.mu.Lock()
	defer em.mu.Unlock()

	volume, exists := em.volumes[containerID]
	if !exists {
		return "", fmt.Errorf("volume not found: %s", containerID)
	}

	if volume.Mounted {
		return volume.MountPath, nil
	}

	// Create mount point
	if err := os.MkdirAll(volume.MountPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create mount point: %w", err)
	}

	// Mount the volume
	cmd := exec.Command("mount", volume.MapperPath, volume.MountPath)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to mount volume: %w (%s)", err, stderr.String())
	}

	volume.Mounted = true
	return volume.MountPath, nil
}

// UnmountVolume unmounts an encrypted volume
func (em *EncryptionManager) UnmountVolume(containerID string) error {
	em.mu.Lock()
	defer em.mu.Unlock()

	volume, exists := em.volumes[containerID]
	if !exists {
		return fmt.Errorf("volume not found: %s", containerID)
	}

	if !volume.Mounted {
		return nil
	}

	// Unmount the volume
	cmd := exec.Command("umount", volume.MountPath)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to unmount volume: %w (%s)", err, stderr.String())
	}

	volume.Mounted = false
	return nil
}

// CloseVolume closes an encrypted volume
func (em *EncryptionManager) CloseVolume(containerID string) error {
	// Get volume info under lock
	em.mu.RLock()
	volume, exists := em.volumes[containerID]
	if !exists {
		em.mu.RUnlock()
		return fmt.Errorf("volume not found: %s", containerID)
	}
	isMounted := volume.Mounted
	mapperName := volume.MapperName
	mapperPath := volume.MapperPath
	em.mu.RUnlock()

	// Unmount first if mounted (this function has its own locking)
	if isMounted {
		if err := em.UnmountVolume(containerID); err != nil {
			return err
		}
	}

	// Close LUKS volume (no lock needed for external command)
	cmd := exec.Command("cryptsetup", "close", mapperName)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		// Check if already closed
		if _, statErr := os.Stat(mapperPath); os.IsNotExist(statErr) {
			return nil
		}
		return fmt.Errorf("failed to close LUKS volume: %w (%s)", err, stderr.String())
	}

	return nil
}

// DeleteEncryptedVolume deletes an encrypted volume and all associated files
func (em *EncryptionManager) DeleteEncryptedVolume(containerID string) error {
	em.mu.Lock()
	defer em.mu.Unlock()

	volume, exists := em.volumes[containerID]
	if !exists {
		// Try to clean up files even if not tracked
		volumePath := filepath.Join(em.volumesDir, fmt.Sprintf("%s.img", containerID))
		keyPath := filepath.Join(em.keysDir, fmt.Sprintf("%s.key", containerID))
		mountPath := filepath.Join(em.mountsDir, containerID)

		os.Remove(volumePath)
		os.Remove(keyPath)
		os.RemoveAll(mountPath)
		return nil
	}

	// Close volume first
	em.mu.Unlock()
	em.CloseVolume(containerID)
	em.mu.Lock()

	// Delete files
	if err := os.Remove(volume.VolumePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete volume file: %w", err)
	}

	if err := os.Remove(volume.KeyPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete key file: %w", err)
	}

	if err := os.RemoveAll(volume.MountPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete mount point: %w", err)
	}

	delete(em.volumes, containerID)
	return nil
}

// GetVolume returns volume info for a container
func (em *EncryptionManager) GetVolume(containerID string) (*EncryptedVolume, bool) {
	em.mu.RLock()
	defer em.mu.RUnlock()
	vol, exists := em.volumes[containerID]
	return vol, exists
}

// SetupEncryptedVolume creates, opens, formats, and mounts an encrypted volume
func (em *EncryptionManager) SetupEncryptedVolume(containerID string, sizeGB int) (*EncryptedVolume, error) {
	// Generate key
	key, err := em.GenerateKey()
	if err != nil {
		return nil, err
	}

	// Create volume
	volume, err := em.CreateEncryptedVolume(containerID, sizeGB, key)
	if err != nil {
		return nil, err
	}

	// Open volume
	if err := em.OpenVolume(containerID); err != nil {
		em.DeleteEncryptedVolume(containerID)
		return nil, err
	}

	// Format volume
	if err := em.FormatVolume(containerID); err != nil {
		em.CloseVolume(containerID)
		em.DeleteEncryptedVolume(containerID)
		return nil, err
	}

	// Mount volume
	if _, err := em.MountVolume(containerID); err != nil {
		em.CloseVolume(containerID)
		em.DeleteEncryptedVolume(containerID)
		return nil, err
	}

	return volume, nil
}

// LoadExistingVolumes loads existing encrypted volumes
func (em *EncryptionManager) LoadExistingVolumes() error {
	em.mu.Lock()
	defer em.mu.Unlock()

	entries, err := os.ReadDir(em.volumesDir)
	if err != nil {
		return nil // Directory might not exist yet
	}

	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".img") {
			continue
		}

		containerID := strings.TrimSuffix(entry.Name(), ".img")
		volumePath := filepath.Join(em.volumesDir, entry.Name())
		keyPath := filepath.Join(em.keysDir, fmt.Sprintf("%s.key", containerID))
		mapperName := fmt.Sprintf("moltbunker-%s", containerID)
		mapperPath := filepath.Join("/dev/mapper", mapperName)
		mountPath := filepath.Join(em.mountsDir, containerID)

		// Check if key exists
		if _, err := os.Stat(keyPath); os.IsNotExist(err) {
			continue
		}

		// Check if mounted
		mounted := false
		if _, err := os.Stat(mountPath); err == nil {
			// Check if actually mounted
			cmd := exec.Command("mountpoint", "-q", mountPath)
			if cmd.Run() == nil {
				mounted = true
			}
		}

		volume := &EncryptedVolume{
			ContainerID: containerID,
			VolumePath:  volumePath,
			MapperName:  mapperName,
			MapperPath:  mapperPath,
			MountPath:   mountPath,
			Mounted:     mounted,
			KeyPath:     keyPath,
		}

		em.volumes[containerID] = volume
	}

	return nil
}

// IsEncryptionAvailable checks if encryption tools are available
func IsEncryptionAvailable() bool {
	// Check for cryptsetup
	cmd := exec.Command("cryptsetup", "--version")
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

// GetKeyAsHex returns the encryption key as hex string
func (em *EncryptionManager) GetKeyAsHex(containerID string) (string, error) {
	em.mu.RLock()
	volume, exists := em.volumes[containerID]
	em.mu.RUnlock()

	if !exists {
		return "", fmt.Errorf("volume not found: %s", containerID)
	}

	key, err := os.ReadFile(volume.KeyPath)
	if err != nil {
		return "", fmt.Errorf("failed to read key: %w", err)
	}

	return hex.EncodeToString(key), nil
}
