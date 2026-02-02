package security

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"

	"golang.org/x/crypto/curve25519"
	"golang.org/x/crypto/hkdf"
	"golang.org/x/crypto/sha3"

	"github.com/moltbunker/moltbunker/pkg/types"
)

const (
	// X25519KeySize is the size of X25519 public/private keys
	X25519KeySize = 32
	// DEKSize is the size of the Data Encryption Key (AES-256)
	DEKSize = 32
	// NonceSize is the size of the nonce for AES-GCM
	NonceSize = 12
)

// DeploymentEncryptionManager manages encryption keys for deployments
type DeploymentEncryptionManager struct {
	deployments map[string]*DeploymentKeys
	mu          sync.RWMutex
	storePath   string
}

// DeploymentKeys holds encryption keys for a deployment
type DeploymentKeys struct {
	DeploymentID    string    `json:"deployment_id"`
	RequesterPubKey []byte    `json:"requester_pub_key"` // Requester's X25519 public key
	ProviderPubKey  []byte    `json:"provider_pub_key"`  // Provider's ephemeral X25519 public key
	EncryptedDEK    []byte    `json:"encrypted_dek"`     // DEK encrypted with shared secret
	DEKNonce        []byte    `json:"dek_nonce"`         // Nonce used for DEK encryption
	CreatedAt       time.Time `json:"created_at"`

	// Not persisted - derived at runtime
	sharedSecret []byte
	dek          []byte
}

// NewDeploymentEncryptionManager creates a new deployment encryption manager
func NewDeploymentEncryptionManager(storePath string) *DeploymentEncryptionManager {
	return &DeploymentEncryptionManager{
		deployments: make(map[string]*DeploymentKeys),
		storePath:   storePath,
	}
}

// GenerateX25519KeyPair generates a new X25519 key pair
func GenerateX25519KeyPair() (publicKey, privateKey []byte, err error) {
	privateKey = make([]byte, X25519KeySize)
	if _, err := rand.Read(privateKey); err != nil {
		return nil, nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	publicKey, err = curve25519.X25519(privateKey, curve25519.Basepoint)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to compute public key: %w", err)
	}

	return publicKey, privateKey, nil
}

// SetupDeploymentEncryption creates encryption keys for a new deployment
// The requesterPubKey is the X25519 public key from the requester
func (dem *DeploymentEncryptionManager) SetupDeploymentEncryption(deploymentID string, requesterPubKey []byte) (*types.DeploymentEncryption, error) {
	if len(requesterPubKey) != X25519KeySize {
		return nil, fmt.Errorf("invalid requester public key size: expected %d, got %d", X25519KeySize, len(requesterPubKey))
	}

	// Generate ephemeral key pair for this deployment
	providerPubKey, providerPrivKey, err := GenerateX25519KeyPair()
	if err != nil {
		return nil, fmt.Errorf("failed to generate provider key pair: %w", err)
	}

	// Compute shared secret using X25519
	sharedSecret, err := curve25519.X25519(providerPrivKey, requesterPubKey)
	if err != nil {
		return nil, fmt.Errorf("failed to compute shared secret: %w", err)
	}

	// Derive DEK using HKDF with SHA3-256
	dek, err := deriveKey(sharedSecret, []byte(deploymentID), DEKSize)
	if err != nil {
		return nil, fmt.Errorf("failed to derive DEK: %w", err)
	}

	// Encrypt the DEK with the shared secret for storage
	// This allows the requester to decrypt using their private key
	encryptedDEK, nonce, err := encryptWithSharedSecret(sharedSecret, dek)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt DEK: %w", err)
	}

	keys := &DeploymentKeys{
		DeploymentID:    deploymentID,
		RequesterPubKey: requesterPubKey,
		ProviderPubKey:  providerPubKey,
		EncryptedDEK:    encryptedDEK,
		DEKNonce:        nonce,
		CreatedAt:       time.Now(),
		sharedSecret:    sharedSecret,
		dek:             dek,
	}

	dem.mu.Lock()
	dem.deployments[deploymentID] = keys
	dem.mu.Unlock()

	return &types.DeploymentEncryption{
		DeploymentID:    deploymentID,
		RequesterPubKey: requesterPubKey,
		EncryptedDEK:    encryptedDEK,
		DEKAlgorithm:    "AES-256-GCM",
		KeyDerivation:   "HKDF-SHA3-256",
		Nonce:           nonce,
		CreatedAt:       keys.CreatedAt,
	}, nil
}

// GetDeploymentKeys returns the encryption keys for a deployment
func (dem *DeploymentEncryptionManager) GetDeploymentKeys(deploymentID string) (*DeploymentKeys, bool) {
	dem.mu.RLock()
	defer dem.mu.RUnlock()
	keys, exists := dem.deployments[deploymentID]
	return keys, exists
}

// EncryptData encrypts data for a deployment using its DEK
func (dem *DeploymentEncryptionManager) EncryptData(deploymentID string, plaintext []byte) ([]byte, error) {
	dem.mu.RLock()
	keys, exists := dem.deployments[deploymentID]
	dem.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("deployment not found: %s", deploymentID)
	}

	if keys.dek == nil {
		return nil, fmt.Errorf("DEK not available for deployment: %s", deploymentID)
	}

	return EncryptAES256GCM(keys.dek, plaintext)
}

// DecryptData decrypts data for a deployment using its DEK (only works on provider side)
func (dem *DeploymentEncryptionManager) DecryptData(deploymentID string, ciphertext []byte) ([]byte, error) {
	dem.mu.RLock()
	keys, exists := dem.deployments[deploymentID]
	dem.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("deployment not found: %s", deploymentID)
	}

	if keys.dek == nil {
		return nil, fmt.Errorf("DEK not available for deployment: %s", deploymentID)
	}

	return DecryptAES256GCM(keys.dek, ciphertext)
}

// RemoveDeployment removes encryption keys for a deployment
func (dem *DeploymentEncryptionManager) RemoveDeployment(deploymentID string) {
	dem.mu.Lock()
	defer dem.mu.Unlock()
	delete(dem.deployments, deploymentID)
}

// GetEncryptionMetadata returns the encryption metadata for a deployment
// This is what gets sent to the requester so they can decrypt outputs
func (dem *DeploymentEncryptionManager) GetEncryptionMetadata(deploymentID string) (*EncryptionMetadata, error) {
	dem.mu.RLock()
	keys, exists := dem.deployments[deploymentID]
	dem.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("deployment not found: %s", deploymentID)
	}

	return &EncryptionMetadata{
		DeploymentID:    deploymentID,
		ProviderPubKey:  keys.ProviderPubKey,
		EncryptedDEK:    keys.EncryptedDEK,
		DEKNonce:        keys.DEKNonce,
		Algorithm:       "AES-256-GCM",
		KeyDerivation:   "HKDF-SHA3-256",
		KeyExchange:     "X25519",
	}, nil
}

// EncryptionMetadata contains all information needed by requester to decrypt
type EncryptionMetadata struct {
	DeploymentID   string `json:"deployment_id"`
	ProviderPubKey []byte `json:"provider_pub_key"` // Provider's ephemeral public key
	EncryptedDEK   []byte `json:"encrypted_dek"`    // DEK encrypted with shared secret
	DEKNonce       []byte `json:"dek_nonce"`        // Nonce for DEK encryption
	Algorithm      string `json:"algorithm"`        // "AES-256-GCM"
	KeyDerivation  string `json:"key_derivation"`   // "HKDF-SHA3-256"
	KeyExchange    string `json:"key_exchange"`     // "X25519"
}

// MarshalJSON marshals encryption metadata to JSON
func (em *EncryptionMetadata) MarshalJSON() ([]byte, error) {
	type Alias EncryptionMetadata
	return json.Marshal((*Alias)(em))
}

// RequesterDecryptor allows requesters to decrypt deployment outputs
type RequesterDecryptor struct {
	privateKey []byte
	publicKey  []byte
}

// NewRequesterDecryptor creates a new requester decryptor with an X25519 key pair
func NewRequesterDecryptor(privateKey, publicKey []byte) (*RequesterDecryptor, error) {
	if len(privateKey) != X25519KeySize {
		return nil, fmt.Errorf("invalid private key size")
	}
	if len(publicKey) != X25519KeySize {
		return nil, fmt.Errorf("invalid public key size")
	}
	return &RequesterDecryptor{
		privateKey: privateKey,
		publicKey:  publicKey,
	}, nil
}

// GetPublicKey returns the requester's public key (to send to providers)
func (rd *RequesterDecryptor) GetPublicKey() []byte {
	return rd.publicKey
}

// DecryptOutput decrypts an encrypted output using the encryption metadata from the provider
func (rd *RequesterDecryptor) DecryptOutput(metadata *EncryptionMetadata, encryptedData []byte) ([]byte, error) {
	// Compute shared secret using our private key and provider's public key
	sharedSecret, err := curve25519.X25519(rd.privateKey, metadata.ProviderPubKey)
	if err != nil {
		return nil, fmt.Errorf("failed to compute shared secret: %w", err)
	}

	// Derive DEK using HKDF with the same parameters as the provider
	dek, err := deriveKey(sharedSecret, []byte(metadata.DeploymentID), DEKSize)
	if err != nil {
		return nil, fmt.Errorf("failed to derive DEK: %w", err)
	}

	// Decrypt the data
	return DecryptAES256GCM(dek, encryptedData)
}

// deriveKey derives a key using HKDF with SHA3-256
func deriveKey(secret, salt []byte, keyLen int) ([]byte, error) {
	hkdfReader := hkdf.New(sha3.New256, secret, salt, []byte("moltbunker-deployment-key"))
	key := make([]byte, keyLen)
	if _, err := io.ReadFull(hkdfReader, key); err != nil {
		return nil, fmt.Errorf("failed to derive key: %w", err)
	}
	return key, nil
}

// encryptWithSharedSecret encrypts data with a shared secret
func encryptWithSharedSecret(sharedSecret, plaintext []byte) (ciphertext, nonce []byte, err error) {
	// Derive an encryption key from the shared secret
	encKey, err := deriveKey(sharedSecret, []byte("dek-encryption"), DEKSize)
	if err != nil {
		return nil, nil, err
	}

	// Generate nonce
	nonce = make([]byte, NonceSize)
	if _, err := rand.Read(nonce); err != nil {
		return nil, nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt using AES-GCM
	ciphertext, err = EncryptAES256GCM(encKey, plaintext)
	if err != nil {
		return nil, nil, err
	}

	return ciphertext, nonce, nil
}

// EncryptedOutput represents an encrypted output with its metadata
type EncryptedOutput struct {
	Metadata      *EncryptionMetadata `json:"metadata"`
	EncryptedData []byte              `json:"encrypted_data"`
	OutputType    string              `json:"output_type"` // "log", "stdout", "stderr", "file"
	Timestamp     time.Time           `json:"timestamp"`
}

// EncryptedLogEntry represents an encrypted log entry
type EncryptedLogEntry struct {
	DeploymentID  string    `json:"deployment_id"`
	EncryptedData []byte    `json:"encrypted_data"`
	Timestamp     time.Time `json:"timestamp"`
}

// CreateEncryptedOutput creates an encrypted output for delivery to the requester
func (dem *DeploymentEncryptionManager) CreateEncryptedOutput(deploymentID string, data []byte, outputType string) (*EncryptedOutput, error) {
	encryptedData, err := dem.EncryptData(deploymentID, data)
	if err != nil {
		return nil, err
	}

	metadata, err := dem.GetEncryptionMetadata(deploymentID)
	if err != nil {
		return nil, err
	}

	return &EncryptedOutput{
		Metadata:      metadata,
		EncryptedData: encryptedData,
		OutputType:    outputType,
		Timestamp:     time.Now(),
	}, nil
}
