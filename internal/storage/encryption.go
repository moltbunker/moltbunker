package storage

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"

	"golang.org/x/crypto/hkdf"

	"github.com/moltbunker/moltbunker/internal/security"
)

// ObjectEncryptor handles per-object encryption using AES-256-GCM.
//
// Each object gets a random 32-byte Data Encryption Key (DEK).
// The DEK encrypts the plaintext content. The DEK itself is then
// encrypted with the owner's key (in Phase 2, via X25519 ECDH —
// for now we use HKDF key derivation from the owner address).
type ObjectEncryptor struct{}

// NewObjectEncryptor creates a new object encryptor.
func NewObjectEncryptor() *ObjectEncryptor {
	return &ObjectEncryptor{}
}

// EncryptionResult contains the encrypted blob and key material.
type EncryptionResult struct {
	Ciphertext   []byte // AES-256-GCM encrypted content (nonce prepended)
	EncryptedDEK []byte // DEK encrypted with owner key
}

// Encrypt encrypts content with a random DEK, then wraps the DEK
// with the provided owner key.
func (e *ObjectEncryptor) Encrypt(plaintext []byte, ownerKey []byte) (*EncryptionResult, error) {
	// Generate random DEK
	dek := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, dek); err != nil {
		return nil, fmt.Errorf("generate DEK: %w", err)
	}

	// Encrypt content with DEK
	ciphertext, err := security.EncryptAES256GCM(dek, plaintext)
	if err != nil {
		return nil, fmt.Errorf("encrypt content: %w", err)
	}

	// Encrypt DEK with owner key
	encryptedDEK, err := security.EncryptAES256GCM(ownerKey, dek)
	if err != nil {
		return nil, fmt.Errorf("encrypt DEK: %w", err)
	}

	return &EncryptionResult{
		Ciphertext:   ciphertext,
		EncryptedDEK: encryptedDEK,
	}, nil
}

// Decrypt decrypts the DEK with the owner key, then decrypts the content.
func (e *ObjectEncryptor) Decrypt(ciphertext, encryptedDEK, ownerKey []byte) ([]byte, error) {
	// Decrypt DEK
	dek, err := security.DecryptAES256GCM(ownerKey, encryptedDEK)
	if err != nil {
		return nil, fmt.Errorf("decrypt DEK: %w", err)
	}

	// Decrypt content
	plaintext, err := security.DecryptAES256GCM(dek, ciphertext)
	if err != nil {
		return nil, fmt.Errorf("decrypt content: %w", err)
	}

	return plaintext, nil
}

// DeriveOwnerKey derives a 32-byte encryption key from an owner identifier
// using HKDF with SHA-256. This is a placeholder — in production, X25519
// ECDH with the owner's public key provides forward secrecy.
func DeriveOwnerKey(owner string) ([]byte, error) {
	hkdfReader := hkdf.New(sha256.New, []byte(owner), []byte("moltbunker-storage-v1"), []byte("object-encryption"))
	key := make([]byte, 32)
	if _, err := io.ReadFull(hkdfReader, key); err != nil {
		return nil, fmt.Errorf("derive owner key: %w", err)
	}
	return key, nil
}
