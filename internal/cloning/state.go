package cloning

import (
	"bytes"
	"compress/gzip"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"time"
)

// StateTransfer handles state serialization and transfer for cloning
type StateTransfer struct {
	compress  bool
	encrypt   bool
	chunkSize int
}

// StateTransferConfig configures state transfer
type StateTransferConfig struct {
	Compress  bool `yaml:"compress"`
	Encrypt   bool `yaml:"encrypt"`
	ChunkSize int  `yaml:"chunk_size"` // For large transfers
}

// DefaultStateTransferConfig returns default state transfer configuration
func DefaultStateTransferConfig() *StateTransferConfig {
	return &StateTransferConfig{
		Compress:  true,
		Encrypt:   true,
		ChunkSize: 1024 * 1024, // 1MB chunks
	}
}

// NewStateTransfer creates a new state transfer handler
func NewStateTransfer(config *StateTransferConfig) *StateTransfer {
	if config == nil {
		config = DefaultStateTransferConfig()
	}

	return &StateTransfer{
		compress:  config.Compress,
		encrypt:   config.Encrypt,
		chunkSize: config.ChunkSize,
	}
}

// StatePackage represents packaged state for transfer
type StatePackage struct {
	Version     int               `json:"version"`
	ContainerID string            `json:"container_id"`
	Timestamp   time.Time         `json:"timestamp"`
	Checksum    string            `json:"checksum"`
	Compressed  bool              `json:"compressed"`
	Encrypted   bool              `json:"encrypted"`
	Size        int64             `json:"size"`
	Data        []byte            `json:"data"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// PackageState packages container state for transfer
func (s *StateTransfer) PackageState(containerID string, data []byte, metadata map[string]string) (*StatePackage, error) {
	processedData := data

	// Compress if enabled
	if s.compress {
		var buf bytes.Buffer
		gz := gzip.NewWriter(&buf)
		if _, err := gz.Write(data); err != nil {
			return nil, fmt.Errorf("compression failed: %w", err)
		}
		if err := gz.Close(); err != nil {
			return nil, fmt.Errorf("compression close failed: %w", err)
		}
		processedData = buf.Bytes()
	}

	// Calculate checksum
	hash := sha256.Sum256(data)
	checksum := fmt.Sprintf("%x", hash)

	pkg := &StatePackage{
		Version:     1,
		ContainerID: containerID,
		Timestamp:   time.Now(),
		Checksum:    checksum,
		Compressed:  s.compress,
		Encrypted:   false,
		Size:        int64(len(data)),
		Data:        processedData,
		Metadata:    metadata,
	}

	return pkg, nil
}

// PackageStateEncrypted packages and encrypts container state
func (s *StateTransfer) PackageStateEncrypted(containerID string, data []byte, key []byte, metadata map[string]string) (*StatePackage, error) {
	pkg, err := s.PackageState(containerID, data, metadata)
	if err != nil {
		return nil, err
	}

	// Encrypt if key provided
	if len(key) > 0 {
		encrypted, err := encryptAESGCM(pkg.Data, key)
		if err != nil {
			return nil, fmt.Errorf("encryption failed: %w", err)
		}
		pkg.Data = encrypted
		pkg.Encrypted = true
	}

	return pkg, nil
}

// UnpackageState extracts state from a package
func (s *StateTransfer) UnpackageState(pkg *StatePackage, key []byte) ([]byte, error) {
	data := pkg.Data

	// Decrypt if encrypted
	if pkg.Encrypted {
		if len(key) == 0 {
			return nil, fmt.Errorf("encrypted package requires decryption key")
		}
		decrypted, err := decryptAESGCM(data, key)
		if err != nil {
			return nil, fmt.Errorf("decryption failed: %w", err)
		}
		data = decrypted
	}

	// Decompress if compressed
	if pkg.Compressed {
		gz, err := gzip.NewReader(bytes.NewReader(data))
		if err != nil {
			return nil, fmt.Errorf("decompression init failed: %w", err)
		}
		defer gz.Close()

		decompressed, err := io.ReadAll(gz)
		if err != nil {
			return nil, fmt.Errorf("decompression failed: %w", err)
		}
		data = decompressed
	}

	// Verify checksum
	hash := sha256.Sum256(data)
	checksum := fmt.Sprintf("%x", hash)
	if checksum != pkg.Checksum {
		return nil, fmt.Errorf("checksum mismatch: expected %s, got %s", pkg.Checksum, checksum)
	}

	return data, nil
}

// SerializePackage serializes a state package for transfer
func (s *StateTransfer) SerializePackage(pkg *StatePackage) ([]byte, error) {
	return json.Marshal(pkg)
}

// DeserializePackage deserializes a state package
func (s *StateTransfer) DeserializePackage(data []byte) (*StatePackage, error) {
	var pkg StatePackage
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, err
	}
	return &pkg, nil
}

// Encryption helpers

func encryptAESGCM(plaintext, key []byte) ([]byte, error) {
	// Ensure key is 32 bytes (AES-256)
	if len(key) < 32 {
		// Derive key using SHA-256
		hash := sha256.Sum256(key)
		key = hash[:]
	} else if len(key) > 32 {
		key = key[:32]
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

func decryptAESGCM(ciphertext, key []byte) ([]byte, error) {
	// Ensure key is 32 bytes (AES-256)
	if len(key) < 32 {
		hash := sha256.Sum256(key)
		key = hash[:]
	} else if len(key) > 32 {
		key = key[:32]
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// ChunkedTransfer handles large state transfers in chunks
type ChunkedTransfer struct {
	PackageID  string `json:"package_id"`
	TotalSize  int64  `json:"total_size"`
	ChunkSize  int    `json:"chunk_size"`
	ChunkCount int    `json:"chunk_count"`
	Checksum   string `json:"checksum"`
}

// Chunk represents a single chunk of data
type Chunk struct {
	PackageID string `json:"package_id"`
	Index     int    `json:"index"`
	Data      []byte `json:"data"`
	Checksum  string `json:"checksum"`
}

// SplitIntoChunks splits a state package into chunks for transfer
func (s *StateTransfer) SplitIntoChunks(pkg *StatePackage) (*ChunkedTransfer, []Chunk, error) {
	data, err := s.SerializePackage(pkg)
	if err != nil {
		return nil, nil, err
	}

	// Generate package ID
	idBytes := make([]byte, 8)
	rand.Read(idBytes)
	packageID := fmt.Sprintf("%x", idBytes)

	// Calculate overall checksum
	hash := sha256.Sum256(data)
	checksum := fmt.Sprintf("%x", hash)

	// Split into chunks
	chunkSize := s.chunkSize
	chunkCount := (len(data) + chunkSize - 1) / chunkSize
	chunks := make([]Chunk, chunkCount)

	for i := 0; i < chunkCount; i++ {
		start := i * chunkSize
		end := start + chunkSize
		if end > len(data) {
			end = len(data)
		}

		chunkData := data[start:end]
		chunkHash := sha256.Sum256(chunkData)

		chunks[i] = Chunk{
			PackageID: packageID,
			Index:     i,
			Data:      chunkData,
			Checksum:  fmt.Sprintf("%x", chunkHash),
		}
	}

	transfer := &ChunkedTransfer{
		PackageID:  packageID,
		TotalSize:  int64(len(data)),
		ChunkSize:  chunkSize,
		ChunkCount: chunkCount,
		Checksum:   checksum,
	}

	return transfer, chunks, nil
}

// AssembleChunks reassembles chunks into a state package
func (s *StateTransfer) AssembleChunks(transfer *ChunkedTransfer, chunks []Chunk) (*StatePackage, error) {
	// Verify we have all chunks
	if len(chunks) != transfer.ChunkCount {
		return nil, fmt.Errorf("incomplete chunks: expected %d, got %d", transfer.ChunkCount, len(chunks))
	}

	// Sort chunks by index
	sortedChunks := make([]Chunk, transfer.ChunkCount)
	for _, chunk := range chunks {
		if chunk.Index < 0 || chunk.Index >= transfer.ChunkCount {
			return nil, fmt.Errorf("invalid chunk index: %d", chunk.Index)
		}
		sortedChunks[chunk.Index] = chunk
	}

	// Assemble data and verify chunk checksums
	var data []byte
	for i, chunk := range sortedChunks {
		// Verify chunk checksum
		hash := sha256.Sum256(chunk.Data)
		if fmt.Sprintf("%x", hash) != chunk.Checksum {
			return nil, fmt.Errorf("chunk %d checksum mismatch", i)
		}
		data = append(data, chunk.Data...)
	}

	// Verify overall checksum
	hash := sha256.Sum256(data)
	if fmt.Sprintf("%x", hash) != transfer.Checksum {
		return nil, fmt.Errorf("overall checksum mismatch")
	}

	// Deserialize package
	return s.DeserializePackage(data)
}
