package tor

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cretz/bine/torutil/ed25519"
)

// OnionManager manages .onion addresses
type OnionManager struct {
	dataDir string
	keys    map[string]ed25519.PrivateKey
}

// NewOnionManager creates a new onion manager
func NewOnionManager(dataDir string) *OnionManager {
	return &OnionManager{
		dataDir: dataDir,
		keys:    make(map[string]ed25519.PrivateKey),
	}
}

// GenerateOnionAddress generates a new .onion address
func (om *OnionManager) GenerateOnionAddress(serviceID string) (string, error) {
	// Generate Ed25519 key
	keyPair, err := ed25519.GenerateKey(nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate onion key: %w", err)
	}

	key, ok := keyPair.(ed25519.PrivateKey)
	if !ok {
		return "", fmt.Errorf("failed to convert key pair to PrivateKey")
	}

	// Save key (simplified - actual implementation would serialize properly)
	keyPath := filepath.Join(om.dataDir, fmt.Sprintf("%s_onion_key", serviceID))
	keyBytes := []byte(key)
	if err := os.WriteFile(keyPath, keyBytes, 0600); err != nil {
		return "", fmt.Errorf("failed to save onion key: %w", err)
	}

	// Store key
	om.keys[serviceID] = key

	// Generate .onion address from public key
	onionAddr := om.generateOnionFromKey(key)
	return onionAddr, nil
}

// LoadOnionAddress loads an existing .onion address
func (om *OnionManager) LoadOnionAddress(serviceID string) (string, error) {
	keyPath := filepath.Join(om.dataDir, fmt.Sprintf("%s_onion_key", serviceID))

	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		return "", fmt.Errorf("failed to read onion key: %w", err)
	}

	key := ed25519.PrivateKey(keyData)
	om.keys[serviceID] = key

	onionAddr := om.generateOnionFromKey(key)
	return onionAddr, nil
}

// generateOnionFromKey generates .onion address from Ed25519 key
func (om *OnionManager) generateOnionFromKey(key ed25519.PrivateKey) string {
	// Extract public key
	publicKey := key.Public().(ed25519.PublicKey)

	// Generate .onion address (v3)
	// This is a simplified version - actual implementation uses specific encoding
	onionAddr := fmt.Sprintf("%x.onion", publicKey[:16])
	return onionAddr
}

// GetKey returns the key for a service
func (om *OnionManager) GetKey(serviceID string) (ed25519.PrivateKey, bool) {
	key, exists := om.keys[serviceID]
	return key, exists
}
