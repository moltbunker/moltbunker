package commands

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"sync"
	"sync/atomic"
)

// AES-256-GCM session parameters. These MUST match cmd/exec-agent/session.go
// byte-for-byte so the CLI and the in-container exec-agent interoperate.
const (
	// execSessionKeySize is the AES-256-GCM key size.
	execSessionKeySize = 32
	// execGCMNonceSize is the standard GCM nonce size.
	execGCMNonceSize = 12
	// execGCMTagSize is the AES-GCM authentication tag size.
	execGCMTagSize = 16
	// execNonceDirBit is the high bit of nonce[0], used to domain-separate the
	// two directions of a session. The CLI/client encrypts with this bit SET
	// (direction 1); the in-container exec-agent encrypts with it CLEAR
	// (direction 0). This MUST match cmd/exec-agent/session.go (nonceDirBit)
	// byte-for-byte so the agent and the CLI interoperate, and the browser
	// (web/src) MUST adopt the same convention (client direction = bit set)
	// when its exec path is wired up.
	execNonceDirBit = byte(0x80)
)

// execSession holds the per-session AES-256-GCM state for E2E terminal I/O on
// the CLI side. It is the wire-compatible counterpart of the exec-agent's
// Session type (cmd/exec-agent/session.go): same nonce layout, same Seal/Open
// semantics, nil AAD.
//
// Wire format produced by Encrypt and consumed by Decrypt:
//
//	[8-byte BE monotonic counter][4-byte random] | ciphertext | 16-byte tag
//
// The 12-byte nonce is prepended to the AEAD output (aead.Seal(nonce, ...)).
//
// Both directions share a single session_key. Each side keeps its own send
// counter; cross-direction (key, nonce) uniqueness is guaranteed by a direction
// bit stamped into the high bit of nonce[0] — the CLI/client sets it, the
// exec-agent clears it (see execNonceDirBit and cmd/exec-agent/session.go). The
// 4 random nonce bytes only add same-direction collision margin; they are no
// longer the cross-direction safeguard.
type execSession struct {
	aead cipher.AEAD

	// encryptCounter is the monotonic send counter. It starts at 0 and is
	// pre-incremented before each Seal (first nonce counter == 1), exactly
	// like the exec-agent.
	encryptCounter atomic.Uint64

	mu sync.Mutex // protects Seal calls (concurrent stdin reads)
}

// newExecSession derives the session key from exec_key + session_nonce and
// initializes an AES-256-GCM AEAD cipher.
//
// session_key = HKDF-SHA256(exec_key, salt=session_nonce, info="session-key")
//
// The derivation is delegated to deriveSessionKey (exec_crypto.go) which is the
// CLI mirror of the agent's NewSession derivation.
func newExecSession(execKey, sessionNonce []byte) (*execSession, error) {
	sessionKey, err := deriveSessionKey(execKey, sessionNonce)
	if err != nil {
		return nil, fmt.Errorf("derive session key: %w", err)
	}

	block, err := aes.NewCipher(sessionKey)
	if err != nil {
		return nil, fmt.Errorf("create AES cipher: %w", err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create GCM: %w", err)
	}

	return &execSession{aead: aead}, nil
}

// Encrypt encrypts plaintext using AES-256-GCM with a counter-based nonce.
// Output: [8-byte BE counter][4-byte random][ciphertext+tag].
func (s *execSession) Encrypt(plaintext []byte) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	nonce := make([]byte, execGCMNonceSize)
	// First 8 bytes: monotonic counter (big-endian), starting at 1.
	counter := s.encryptCounter.Add(1)
	binary.BigEndian.PutUint64(nonce[:8], counter)
	// Direction tag: the CLI/client encrypts with the high bit SET (direction 1)
	// so it can never collide with the exec-agent (direction 0) on a
	// (key, nonce) pair. The counter starts at 1 and never approaches 2^63, so
	// the top bit of nonce[0] is always free for this flag.
	nonce[0] |= execNonceDirBit
	// Last 4 bytes: random for additional uniqueness.
	if _, err := rand.Read(nonce[8:]); err != nil {
		return nil, fmt.Errorf("generate nonce random: %w", err)
	}

	// Seal prepends the nonce to the ciphertext (dst == nonce).
	return s.aead.Seal(nonce, nonce, plaintext, nil), nil
}

// Decrypt decrypts ciphertext produced by the peer's Encrypt.
// Expects format: [12-byte nonce][ciphertext+tag], nil AAD.
func (s *execSession) Decrypt(data []byte) ([]byte, error) {
	if len(data) < execGCMNonceSize+execGCMTagSize {
		return nil, fmt.Errorf("ciphertext too short: %d bytes", len(data))
	}

	nonce := data[:execGCMNonceSize]
	ciphertext := data[execGCMNonceSize:]

	plaintext, err := s.aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("GCM decrypt: %w", err)
	}
	return plaintext, nil
}
