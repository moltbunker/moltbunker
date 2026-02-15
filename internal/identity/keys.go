package identity

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/moltbunker/moltbunker/internal/logging"
	"github.com/moltbunker/moltbunker/pkg/types"
	"golang.org/x/crypto/argon2"
)

// Encrypted key file format:
// [4 bytes magic] [16 bytes salt] [12 bytes nonce] [variable ciphertext]
//
// Magic bytes identify the file as an encrypted Moltbunker key file.
// Salt is used for argon2id key derivation.
// Nonce is used for AES-256-GCM encryption.
// Ciphertext is the encrypted Ed25519 private key (64 bytes + 16 bytes GCM tag).

var (
	// encryptedKeyMagic identifies an encrypted key file ("MBEK" = MoltBunker Encrypted Key)
	encryptedKeyMagic = []byte{0x4D, 0x42, 0x45, 0x4B}

	// argon2id parameters
	argon2Time    uint32 = 1
	argon2Memory  uint32 = 64 * 1024 // 64 MB
	argon2Threads uint8  = 4
	argon2KeyLen  uint32 = 32 // AES-256

	// ErrWrongPassphrase is returned when decryption fails due to wrong passphrase
	ErrWrongPassphrase = errors.New("wrong passphrase or corrupted key file")

	// ErrInvalidKeyFile is returned when the file format is invalid
	ErrInvalidKeyFile = errors.New("invalid encrypted key file")
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
			logging.Audit(logging.AuditEvent{
				Operation: "key_generated",
				Actor:     km.NodeIDString(),
				Target:    keyPath,
				Result:    "success",
				Details:   "new Ed25519 key pair generated",
			})
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

	// Validate key size before using. ed25519.PrivateKey is a []byte alias —
	// Go will cast any slice, but Sign() panics if len != PrivateKeySize.
	if len(privateKeyData) != ed25519.PrivateKeySize {
		return fmt.Errorf("invalid private key file: expected %d bytes, got %d",
			ed25519.PrivateKeySize, len(privateKeyData))
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

// SaveEncryptedKey encrypts the private key with a passphrase and saves it to disk.
// The file format is: magic (4B) + salt (16B) + nonce (12B) + ciphertext (variable).
// The encryption key is derived from the passphrase using argon2id.
func (km *KeyManager) SaveEncryptedKey(path string, passphrase []byte) error {
	if km.privateKey == nil {
		return fmt.Errorf("save encrypted key: no private key loaded")
	}

	// Generate random salt for argon2id
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return fmt.Errorf("save encrypted key: failed to generate salt: %w", err)
	}

	// Derive encryption key from passphrase using argon2id
	derivedKey := argon2.IDKey(passphrase, salt, argon2Time, argon2Memory, argon2Threads, argon2KeyLen)

	// Create AES-256-GCM cipher
	block, err := aes.NewCipher(derivedKey)
	if err != nil {
		return fmt.Errorf("save encrypted key: failed to create cipher: %w", err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("save encrypted key: failed to create GCM: %w", err)
	}

	// Generate random nonce
	nonce := make([]byte, aead.NonceSize()) // 12 bytes for GCM
	if _, err := rand.Read(nonce); err != nil {
		return fmt.Errorf("save encrypted key: failed to generate nonce: %w", err)
	}

	// Encrypt the private key
	ciphertext := aead.Seal(nil, nonce, []byte(km.privateKey), nil)

	// Assemble file: magic + salt + nonce + ciphertext
	fileData := make([]byte, 0, len(encryptedKeyMagic)+len(salt)+len(nonce)+len(ciphertext))
	fileData = append(fileData, encryptedKeyMagic...)
	fileData = append(fileData, salt...)
	fileData = append(fileData, nonce...)
	fileData = append(fileData, ciphertext...)

	// Create directory if needed
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("save encrypted key: failed to create directory: %w", err)
	}

	// Write file with restricted permissions
	if err := os.WriteFile(path, fileData, 0600); err != nil {
		return fmt.Errorf("save encrypted key: failed to write file: %w", err)
	}

	return nil
}

// LoadEncryptedKey reads an encrypted key file, decrypts it with the passphrase,
// and restores the private key, public key, and node ID.
func (km *KeyManager) LoadEncryptedKey(path string, passphrase []byte) error {
	fileData, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("load encrypted key: %w", err)
	}

	// Minimum file size: magic (4) + salt (16) + nonce (12) + at least 1 byte ciphertext + 16 byte tag
	const minSize = 4 + 16 + 12 + 1 + 16
	if len(fileData) < minSize {
		return fmt.Errorf("load encrypted key: %w: file too short", ErrInvalidKeyFile)
	}

	// Verify magic bytes
	for i := 0; i < len(encryptedKeyMagic); i++ {
		if fileData[i] != encryptedKeyMagic[i] {
			return fmt.Errorf("load encrypted key: %w: invalid magic bytes", ErrInvalidKeyFile)
		}
	}

	// Extract salt, nonce, ciphertext
	offset := len(encryptedKeyMagic)
	salt := fileData[offset : offset+16]
	offset += 16
	nonce := fileData[offset : offset+12]
	offset += 12
	ciphertext := fileData[offset:]

	// Derive decryption key from passphrase using same argon2id params
	derivedKey := argon2.IDKey(passphrase, salt, argon2Time, argon2Memory, argon2Threads, argon2KeyLen)

	// Create AES-256-GCM cipher
	block, err := aes.NewCipher(derivedKey)
	if err != nil {
		return fmt.Errorf("load encrypted key: failed to create cipher: %w", err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("load encrypted key: failed to create GCM: %w", err)
	}

	// Decrypt
	plaintext, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return fmt.Errorf("load encrypted key: %w", ErrWrongPassphrase)
	}

	// Validate that decrypted data is a valid Ed25519 private key (64 bytes)
	if len(plaintext) != ed25519.PrivateKeySize {
		return fmt.Errorf("load encrypted key: %w: unexpected key size %d", ErrInvalidKeyFile, len(plaintext))
	}

	// Restore key material
	privateKey := ed25519.PrivateKey(plaintext)
	publicKey := privateKey.Public().(ed25519.PublicKey)

	km.privateKey = privateKey
	km.publicKey = publicKey

	// Compute NodeID
	hash := sha256.Sum256(publicKey)
	km.nodeID = types.NodeID(hash)

	return nil
}

// IsEncryptedKeyFile checks whether the file at the given path is an encrypted key file
// by reading the first 4 bytes and comparing to the magic header.
func IsEncryptedKeyFile(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	magic := make([]byte, len(encryptedKeyMagic))
	n, err := f.Read(magic)
	if err != nil {
		return false, err
	}
	if n < len(encryptedKeyMagic) {
		return false, nil
	}

	for i := 0; i < len(encryptedKeyMagic); i++ {
		if magic[i] != encryptedKeyMagic[i] {
			return false, nil
		}
	}
	return true, nil
}
