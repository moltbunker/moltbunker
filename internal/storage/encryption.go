package storage

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"

	"golang.org/x/crypto/hkdf"

	"github.com/moltbunker/moltbunker/internal/security"
)

// Object encryption uses envelope encryption: each object gets a fresh random
// 32-byte Data Encryption Key (DEK) that encrypts the content with AES-256-GCM.
// The DEK itself is then sealed to the owner's X25519 public key using the
// project's ECIES primitive (security.SealToX25519) — the SAME key ladder used
// for exec-key delivery (SEC-10) and R5 image encryption. This replaces the old
// DeriveOwnerKey scheme, which derived an AES key deterministically from the
// (public) owner address string — meaning anyone who knew the address could
// recompute the DEK-wrapping key. The sealed DEK is stored as a JSON-encoded
// security.X25519Envelope in ObjectInfo.EncryptedDEK.
//
// Back-compat: objects written by the old scheme carry a raw AES-GCM-wrapped DEK
// (not a JSON envelope). Those are detected on read (json.Unmarshal into an
// X25519Envelope fails, or the EphemeralPub field is absent) and decrypted via
// legacyDeriveOwnerKey. New writes always use the X25519 envelope.

// OwnerKeyEncryptor performs envelope encryption sealing the DEK to an owner's
// X25519 public key.
type OwnerKeyEncryptor struct{}

// NewOwnerKeyEncryptor creates a new X25519 envelope encryptor.
func NewOwnerKeyEncryptor() *OwnerKeyEncryptor { return &OwnerKeyEncryptor{} }

// X25519EncryptionResult is the result of sealing an object to an owner key.
type X25519EncryptionResult struct {
	// Ciphertext is the AES-256-GCM-encrypted content (nonce prepended).
	Ciphertext []byte
	// SealedDEK is the per-object DEK sealed to the owner's X25519 public key.
	SealedDEK *security.X25519Envelope
}

// Encrypt generates a random DEK, encrypts the plaintext with it, then seals the
// DEK to the owner's 32-byte X25519 public key.
func (e *OwnerKeyEncryptor) Encrypt(plaintext, ownerPub []byte) (*X25519EncryptionResult, error) {
	if len(ownerPub) != security.X25519KeySize {
		return nil, fmt.Errorf("owner public key must be %d bytes, got %d", security.X25519KeySize, len(ownerPub))
	}

	dek := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, dek); err != nil {
		return nil, fmt.Errorf("generate DEK: %w", err)
	}

	ciphertext, err := security.EncryptAES256GCM(dek, plaintext)
	if err != nil {
		return nil, fmt.Errorf("encrypt content: %w", err)
	}

	sealed, err := security.SealToX25519(ownerPub, dek)
	if err != nil {
		return nil, fmt.Errorf("seal DEK: %w", err)
	}

	return &X25519EncryptionResult{Ciphertext: ciphertext, SealedDEK: sealed}, nil
}

// Decrypt opens the sealed DEK with the owner's X25519 private key, then decrypts
// the content.
func (e *OwnerKeyEncryptor) Decrypt(ciphertext []byte, sealedDEK *security.X25519Envelope, ownerPriv []byte) ([]byte, error) {
	if sealedDEK == nil {
		return nil, fmt.Errorf("nil sealed DEK")
	}
	dek, err := security.OpenFromX25519(ownerPriv, sealedDEK)
	if err != nil {
		return nil, fmt.Errorf("open DEK: %w", err)
	}
	plaintext, err := security.DecryptAES256GCM(dek, ciphertext)
	if err != nil {
		return nil, fmt.Errorf("decrypt content: %w", err)
	}
	return plaintext, nil
}

// MarshalSealedDEK serializes an X25519 envelope for storage in
// ObjectInfo.EncryptedDEK.
func MarshalSealedDEK(env *security.X25519Envelope) ([]byte, error) {
	if env == nil {
		return nil, fmt.Errorf("nil envelope")
	}
	return json.Marshal(env)
}

// looksLikeX25519Envelope reports whether the stored EncryptedDEK is a
// JSON-encoded X25519 envelope (new scheme) rather than a raw AES-GCM-wrapped
// DEK (legacy scheme). It checks both that the bytes parse as JSON and that the
// envelope's required EphemeralPub field is present.
func looksLikeX25519Envelope(encryptedDEK []byte) (*security.X25519Envelope, bool) {
	if len(encryptedDEK) == 0 {
		return nil, false
	}
	var env security.X25519Envelope
	if err := json.Unmarshal(encryptedDEK, &env); err != nil {
		return nil, false
	}
	if len(env.EphemeralPub) != security.X25519KeySize {
		return nil, false
	}
	return &env, true
}

// --- Back-compat (legacy) path ---

// ObjectEncryptor is the legacy envelope encryptor that wraps the DEK with a
// raw 32-byte symmetric key (the AES-GCM-wrapped-DEK scheme). It is retained for
// reading objects written before the X25519 migration. New code should use
// OwnerKeyEncryptor.
type ObjectEncryptor struct{}

// NewObjectEncryptor creates a new legacy object encryptor.
func NewObjectEncryptor() *ObjectEncryptor { return &ObjectEncryptor{} }

// EncryptionResult contains the legacy encrypted blob and key material.
type EncryptionResult struct {
	Ciphertext   []byte // AES-256-GCM encrypted content (nonce prepended)
	EncryptedDEK []byte // DEK encrypted with the (symmetric) owner key
}

// Encrypt encrypts content with a random DEK, then wraps the DEK with the
// provided symmetric owner key (legacy scheme).
func (e *ObjectEncryptor) Encrypt(plaintext, ownerKey []byte) (*EncryptionResult, error) {
	dek := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, dek); err != nil {
		return nil, fmt.Errorf("generate DEK: %w", err)
	}
	ciphertext, err := security.EncryptAES256GCM(dek, plaintext)
	if err != nil {
		return nil, fmt.Errorf("encrypt content: %w", err)
	}
	encryptedDEK, err := security.EncryptAES256GCM(ownerKey, dek)
	if err != nil {
		return nil, fmt.Errorf("encrypt DEK: %w", err)
	}
	return &EncryptionResult{Ciphertext: ciphertext, EncryptedDEK: encryptedDEK}, nil
}

// Decrypt decrypts the DEK with the symmetric owner key, then decrypts content
// (legacy scheme).
func (e *ObjectEncryptor) Decrypt(ciphertext, encryptedDEK, ownerKey []byte) ([]byte, error) {
	dek, err := security.DecryptAES256GCM(ownerKey, encryptedDEK)
	if err != nil {
		return nil, fmt.Errorf("decrypt DEK: %w", err)
	}
	plaintext, err := security.DecryptAES256GCM(dek, ciphertext)
	if err != nil {
		return nil, fmt.Errorf("decrypt content: %w", err)
	}
	return plaintext, nil
}

// legacyDeriveOwnerKey derives a 32-byte symmetric key from an owner identifier
// using HKDF-SHA256. This is the OLD scheme: the owner address is public, so the
// derived key is computable by any observer — it is NOT a secret. It is retained
// only to read objects written before the X25519 migration. Do not use for new
// writes.
func legacyDeriveOwnerKey(owner string) ([]byte, error) {
	hkdfReader := hkdf.New(sha256.New, []byte(owner), []byte("moltbunker-storage-v1"), []byte("object-encryption"))
	key := make([]byte, 32)
	if _, err := io.ReadFull(hkdfReader, key); err != nil {
		return nil, fmt.Errorf("derive owner key: %w", err)
	}
	return key, nil
}
