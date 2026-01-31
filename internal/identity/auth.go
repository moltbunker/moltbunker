package identity

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/moltbunker/moltbunker/pkg/types"
)

// AuthManager handles authentication and authorization
type AuthManager struct {
	keyManager *KeyManager
}

// NewAuthManager creates a new auth manager
func NewAuthManager(keyManager *KeyManager) *AuthManager {
	return &AuthManager{
		keyManager: keyManager,
	}
}

// CreateAuthToken creates an authentication token
func (am *AuthManager) CreateAuthToken(nodeID types.NodeID, timestamp time.Time) ([]byte, error) {
	// Create token: nodeID + timestamp + signature
	token := make([]byte, 32+8+64) // nodeID (32) + timestamp (8) + signature (64)

	copy(token[0:32], nodeID[:])
	binary.BigEndian.PutUint64(token[32:40], uint64(timestamp.Unix()))

	// Sign the token
	signature, err := am.keyManager.Sign(token[0:40])
	if err != nil {
		return nil, fmt.Errorf("failed to sign token: %w", err)
	}

	copy(token[40:], signature)
	return token, nil
}

// VerifyAuthToken verifies an authentication token
func (am *AuthManager) VerifyAuthToken(token []byte, publicKey ed25519.PublicKey) (types.NodeID, time.Time, error) {
	if len(token) != 104 { // 32 + 8 + 64
		return types.NodeID{}, time.Time{}, fmt.Errorf("invalid token length")
	}

	var nodeID types.NodeID
	copy(nodeID[:], token[0:32])

	timestamp := time.Unix(int64(binary.BigEndian.Uint64(token[32:40])), 0)
	signature := token[40:]

	// Verify signature
	if !ed25519.Verify(publicKey, token[0:40], signature) {
		return types.NodeID{}, time.Time{}, fmt.Errorf("invalid signature")
	}

	// Check token expiration (5 minutes)
	if time.Since(timestamp) > 5*time.Minute {
		return types.NodeID{}, time.Time{}, fmt.Errorf("token expired")
	}

	return nodeID, timestamp, nil
}

// SignMessage signs a message with the node's private key
func (am *AuthManager) SignMessage(message []byte) ([]byte, error) {
	return am.keyManager.Sign(message)
}

// VerifyMessage verifies a message signature
func (am *AuthManager) VerifyMessage(message []byte, signature []byte, publicKey ed25519.PublicKey) bool {
	return ed25519.Verify(publicKey, message, signature)
}

// GenerateNonce generates a random nonce for ChaCha20Poly1305
func GenerateNonce() ([24]byte, error) {
	var nonce [24]byte
	_, err := rand.Read(nonce[:])
	if err != nil {
		return nonce, fmt.Errorf("failed to generate nonce: %w", err)
	}
	return nonce, nil
}
