//go:build linux

package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
	"sync"
	"sync/atomic"

	"golang.org/x/crypto/hkdf"
)

const (
	// sessionKeySize is the AES-256-GCM key size.
	sessionKeySize = 32
	// gcmNonceSize is the standard GCM nonce size.
	gcmNonceSize = 12
	// gcmTagSize is the AES-GCM authentication tag size.
	gcmTagSize = 16
)

// Session holds the per-session encryption state for E2E terminal I/O.
// Both sides (exec-agent inside container + CLI/browser) derive the same
// session_key from the shared exec_key and a random session_nonce.
type Session struct {
	sessionKey []byte
	aead       cipher.AEAD

	// Monotonic nonce counters prevent nonce reuse.
	// Separate counters for encrypt (outgoing) and decrypt (incoming)
	// because each side maintains its own send counter.
	encryptCounter atomic.Uint64
	decryptCounter atomic.Uint64

	mu sync.Mutex // protects Seal calls (concurrent PTY reads)
}

// NewSession derives a session key from the exec_key and session_nonce,
// then initializes an AES-256-GCM AEAD cipher.
//
// Key derivation: session_key = HKDF-SHA256(exec_key, salt=session_nonce, info="session-key")
func NewSession(execKey, sessionNonce []byte) (*Session, error) {
	if len(execKey) != sessionKeySize {
		return nil, fmt.Errorf("exec_key must be %d bytes, got %d", sessionKeySize, len(execKey))
	}
	if len(sessionNonce) == 0 {
		return nil, fmt.Errorf("session_nonce must not be empty")
	}

	// Derive session key using HKDF-SHA256
	hkdfReader := hkdf.New(sha256.New, execKey, sessionNonce, []byte("session-key"))
	sessionKey := make([]byte, sessionKeySize)
	if _, err := io.ReadFull(hkdfReader, sessionKey); err != nil {
		return nil, fmt.Errorf("HKDF derive session key: %w", err)
	}

	block, err := aes.NewCipher(sessionKey)
	if err != nil {
		return nil, fmt.Errorf("create AES cipher: %w", err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create GCM: %w", err)
	}

	return &Session{
		sessionKey: sessionKey,
		aead:       aead,
	}, nil
}

// Encrypt encrypts plaintext using AES-256-GCM with a counter-based nonce.
// The nonce is prepended to the ciphertext: [8-byte counter][4-byte random][ciphertext+tag].
func (s *Session) Encrypt(plaintext []byte) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	nonce := make([]byte, gcmNonceSize)
	// First 8 bytes: monotonic counter (big-endian)
	counter := s.encryptCounter.Add(1)
	binary.BigEndian.PutUint64(nonce[:8], counter)
	// Last 4 bytes: random for additional uniqueness
	if _, err := rand.Read(nonce[8:]); err != nil {
		return nil, fmt.Errorf("generate nonce random: %w", err)
	}

	ciphertext := s.aead.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// Decrypt decrypts ciphertext produced by Encrypt.
// Expects format: [12-byte nonce][ciphertext+tag].
func (s *Session) Decrypt(data []byte) ([]byte, error) {
	if len(data) < gcmNonceSize+gcmTagSize {
		return nil, fmt.Errorf("ciphertext too short: %d bytes", len(data))
	}

	nonce := data[:gcmNonceSize]
	ciphertext := data[gcmNonceSize:]

	plaintext, err := s.aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("GCM decrypt: %w", err)
	}

	return plaintext, nil
}

// DeriveExecKey derives an exec_key from a master KEK and deploy nonce.
// exec_key = HKDF-SHA256(master_kek, salt=deploy_nonce, info="exec-key")
// This is used on the CLI/browser side during deployment.
func DeriveExecKey(masterKEK, deployNonce []byte) ([]byte, error) {
	if len(masterKEK) != sessionKeySize {
		return nil, fmt.Errorf("master_kek must be %d bytes", sessionKeySize)
	}

	hkdfReader := hkdf.New(sha256.New, masterKEK, deployNonce, []byte("exec-key"))
	execKey := make([]byte, sessionKeySize)
	if _, err := io.ReadFull(hkdfReader, execKey); err != nil {
		return nil, fmt.Errorf("HKDF derive exec key: %w", err)
	}
	return execKey, nil
}
