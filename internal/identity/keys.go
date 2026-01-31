package identity

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"github.com/moltbunker/moltbunker/pkg/types"
)

// KeyManager manages Ed25519 key pairs for P2P identity
type KeyManager struct {
	privateKey ed25519.PrivateKey
	publicKey  ed25519.PublicKey
	nodeID     types.NodeID
	keyPath    string
}

// NewKeyManager creates a new key manager
func NewKeyManager(keyPath string) (*KeyManager, error) {
	km := &KeyManager{
		keyPath: keyPath,
	}

	// Try to load existing keys
	if err := km.LoadKeys(); err != nil {
		// Generate new keys if they don't exist
		if os.IsNotExist(err) {
			if err := km.GenerateKeys(); err != nil {
				return nil, fmt.Errorf("failed to generate keys: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to load keys: %w", err)
		}
	}

	return km, nil
}

// GenerateKeys generates a new Ed25519 key pair
func (km *KeyManager) GenerateKeys() error {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return fmt.Errorf("failed to generate Ed25519 key: %w", err)
	}

	km.privateKey = privateKey
	km.publicKey = publicKey

	// Compute NodeID as SHA-256 of public key
	hash := sha256.Sum256(publicKey)
	km.nodeID = types.NodeID(hash)

	// Save keys to disk
	if err := km.SaveKeys(); err != nil {
		return fmt.Errorf("failed to save keys: %w", err)
	}

	return nil
}

// LoadKeys loads keys from disk
func (km *KeyManager) LoadKeys() error {
	// Load private key
	privateKeyData, err := os.ReadFile(km.keyPath)
	if err != nil {
		return err
	}

	privateKey := ed25519.PrivateKey(privateKeyData)
	publicKey := privateKey.Public().(ed25519.PublicKey)

	km.privateKey = privateKey
	km.publicKey = publicKey

	// Compute NodeID
	hash := sha256.Sum256(publicKey)
	km.nodeID = types.NodeID(hash)

	return nil
}

// SaveKeys saves keys to disk
func (km *KeyManager) SaveKeys() error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(km.keyPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create key directory: %w", err)
	}

	// Save private key with restricted permissions
	if err := os.WriteFile(km.keyPath, km.privateKey, 0600); err != nil {
		return fmt.Errorf("failed to write private key: %w", err)
	}

	return nil
}

// PrivateKey returns the private key
func (km *KeyManager) PrivateKey() ed25519.PrivateKey {
	return km.privateKey
}

// PublicKey returns the public key
func (km *KeyManager) PublicKey() ed25519.PublicKey {
	return km.publicKey
}

// NodeID returns the node ID (SHA-256 of public key)
func (km *KeyManager) NodeID() types.NodeID {
	return km.nodeID
}

// NodeIDString returns the node ID as a hex string
func (km *KeyManager) NodeIDString() string {
	return hex.EncodeToString(km.nodeID[:])
}

// Sign signs a message with the private key
func (km *KeyManager) Sign(message []byte) ([]byte, error) {
	return ed25519.Sign(km.privateKey, message), nil
}

// Verify verifies a signature with the public key
func (km *KeyManager) Verify(message []byte, signature []byte) bool {
	return ed25519.Verify(km.publicKey, message, signature)
}
