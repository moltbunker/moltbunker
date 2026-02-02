package security

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"golang.org/x/crypto/argon2"
)

// RequesterKeyManager manages the requester's encryption keys
type RequesterKeyManager struct {
	keyStorePath string
	privateKey   []byte
	publicKey    []byte
	mu           sync.RWMutex
}

// EncryptedKeyStore represents the encrypted key storage format
type EncryptedKeyStore struct {
	Version        int    `json:"version"`
	EncryptedKey   []byte `json:"encrypted_key"`   // ChaCha20Poly1305 encrypted private key
	PublicKey      []byte `json:"public_key"`      // Public key (not encrypted)
	Salt           []byte `json:"salt"`            // Argon2 salt for key derivation
	Algorithm      string `json:"algorithm"`       // "ChaCha20Poly1305"
	KeyDerivation  string `json:"key_derivation"`  // "Argon2id"
}

const (
	keyStoreVersion   = 1
	argon2Time        = 3
	argon2Memory      = 64 * 1024 // 64 MB
	argon2Threads     = 4
	argon2KeyLen      = 32
	saltLen           = 32
)

// NewRequesterKeyManager creates a new requester key manager
func NewRequesterKeyManager(keyStorePath string) *RequesterKeyManager {
	return &RequesterKeyManager{
		keyStorePath: keyStorePath,
	}
}

// GenerateNewKeys generates a new X25519 key pair for the requester
func (rkm *RequesterKeyManager) GenerateNewKeys() error {
	pubKey, privKey, err := GenerateX25519KeyPair()
	if err != nil {
		return fmt.Errorf("failed to generate key pair: %w", err)
	}

	rkm.mu.Lock()
	rkm.publicKey = pubKey
	rkm.privateKey = privKey
	rkm.mu.Unlock()

	return nil
}

// SaveKeys saves the keys to disk, encrypted with a passphrase
func (rkm *RequesterKeyManager) SaveKeys(passphrase string) error {
	rkm.mu.RLock()
	pubKey := rkm.publicKey
	privKey := rkm.privateKey
	rkm.mu.RUnlock()

	if pubKey == nil || privKey == nil {
		return fmt.Errorf("no keys to save - generate keys first")
	}

	// Generate salt for key derivation
	salt, err := GenerateKey(saltLen)
	if err != nil {
		return fmt.Errorf("failed to generate salt: %w", err)
	}

	// Derive encryption key from passphrase using Argon2id
	encKey := argon2.IDKey([]byte(passphrase), salt, argon2Time, argon2Memory, argon2Threads, argon2KeyLen)

	// Encrypt private key with ChaCha20Poly1305
	encryptedPrivKey, err := EncryptChaCha20Poly1305(encKey, privKey)
	if err != nil {
		return fmt.Errorf("failed to encrypt private key: %w", err)
	}

	// Create key store
	keyStore := EncryptedKeyStore{
		Version:        keyStoreVersion,
		EncryptedKey:   encryptedPrivKey,
		PublicKey:      pubKey,
		Salt:           salt,
		Algorithm:      "ChaCha20Poly1305",
		KeyDerivation:  "Argon2id",
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(keyStore, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal key store: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(rkm.keyStorePath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create key store directory: %w", err)
	}

	// Write to file with restrictive permissions
	if err := os.WriteFile(rkm.keyStorePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write key store: %w", err)
	}

	return nil
}

// LoadKeys loads the keys from disk, decrypting with the passphrase
func (rkm *RequesterKeyManager) LoadKeys(passphrase string) error {
	// Read key store file
	data, err := os.ReadFile(rkm.keyStorePath)
	if err != nil {
		return fmt.Errorf("failed to read key store: %w", err)
	}

	// Parse key store
	var keyStore EncryptedKeyStore
	if err := json.Unmarshal(data, &keyStore); err != nil {
		return fmt.Errorf("failed to parse key store: %w", err)
	}

	// Check version
	if keyStore.Version != keyStoreVersion {
		return fmt.Errorf("unsupported key store version: %d", keyStore.Version)
	}

	// Derive decryption key from passphrase
	decKey := argon2.IDKey([]byte(passphrase), keyStore.Salt, argon2Time, argon2Memory, argon2Threads, argon2KeyLen)

	// Decrypt private key
	privKey, err := DecryptChaCha20Poly1305(decKey, keyStore.EncryptedKey)
	if err != nil {
		return fmt.Errorf("failed to decrypt private key (wrong passphrase?): %w", err)
	}

	rkm.mu.Lock()
	rkm.publicKey = keyStore.PublicKey
	rkm.privateKey = privKey
	rkm.mu.Unlock()

	return nil
}

// GetPublicKey returns the public key (safe to share with providers)
func (rkm *RequesterKeyManager) GetPublicKey() ([]byte, error) {
	rkm.mu.RLock()
	defer rkm.mu.RUnlock()

	if rkm.publicKey == nil {
		return nil, fmt.Errorf("no keys loaded")
	}

	// Return a copy to prevent modification
	result := make([]byte, len(rkm.publicKey))
	copy(result, rkm.publicKey)
	return result, nil
}

// CreateDecryptor creates a decryptor for decrypting deployment outputs
func (rkm *RequesterKeyManager) CreateDecryptor() (*RequesterDecryptor, error) {
	rkm.mu.RLock()
	defer rkm.mu.RUnlock()

	if rkm.privateKey == nil || rkm.publicKey == nil {
		return nil, fmt.Errorf("no keys loaded")
	}

	return NewRequesterDecryptor(rkm.privateKey, rkm.publicKey)
}

// KeysExist checks if a key store file exists
func (rkm *RequesterKeyManager) KeysExist() bool {
	_, err := os.Stat(rkm.keyStorePath)
	return err == nil
}

// DeleteKeys deletes the key store file
func (rkm *RequesterKeyManager) DeleteKeys() error {
	rkm.mu.Lock()
	rkm.privateKey = nil
	rkm.publicKey = nil
	rkm.mu.Unlock()

	if err := os.Remove(rkm.keyStorePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete key store: %w", err)
	}

	return nil
}

// PublicKeyHex returns the public key as a hex string for display
func (rkm *RequesterKeyManager) PublicKeyHex() (string, error) {
	pubKey, err := rkm.GetPublicKey()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", pubKey), nil
}
