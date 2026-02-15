package api

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/moltbunker/moltbunker/internal/logging"
)

// WalletAuthManager manages wallet-based authentication for permissionless access
type WalletAuthManager struct {
	challenges map[string]*Challenge
	sessions   map[string]*WalletSession
	mu         sync.RWMutex
	timeout    time.Duration
}

// Challenge represents an authentication challenge for wallet signing
type Challenge struct {
	Nonce     string
	Address   string
	Message   string
	CreatedAt time.Time
	ExpiresAt time.Time
}

// WalletSession represents an authenticated wallet session
type WalletSession struct {
	Address   string
	Token     string
	CreatedAt time.Time
	ExpiresAt time.Time
}

// NewWalletAuthManager creates a new wallet authentication manager
func NewWalletAuthManager(timeout time.Duration) *WalletAuthManager {
	if timeout == 0 {
		timeout = 5 * time.Minute
	}

	wam := &WalletAuthManager{
		challenges: make(map[string]*Challenge),
		sessions:   make(map[string]*WalletSession),
		timeout:    timeout,
	}

	// Start cleanup goroutine
	go wam.cleanupLoop()

	return wam
}

// CreateChallenge creates a new auth challenge for a wallet address
func (m *WalletAuthManager) CreateChallenge(address string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Normalize address
	address = strings.ToLower(address)
	if !common.IsHexAddress(address) {
		return "", fmt.Errorf("invalid wallet address format")
	}

	// If a valid unexpired challenge already exists for this address, return it
	// (prevents DoS via challenge overwrite)
	if existing, ok := m.challenges[address]; ok && time.Now().Before(existing.ExpiresAt) {
		return existing.Message, nil
	}

	// Generate random nonce
	nonce := make([]byte, 32)
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}
	nonceHex := hex.EncodeToString(nonce)

	// Create challenge message
	timestamp := time.Now().Unix()
	message := fmt.Sprintf("Sign this message to authenticate with MoltBunker.\n\nWallet: %s\nNonce: %s\nTimestamp: %d",
		address, nonceHex, timestamp)

	m.challenges[address] = &Challenge{
		Nonce:     nonceHex,
		Address:   address,
		Message:   message,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(m.timeout),
	}

	logging.Debug("wallet auth challenge created",
		"address", address,
		"expires_in", m.timeout.String(),
		logging.Component("api"))

	return message, nil
}

// VerifySignature verifies a wallet signature and returns the verified address
func (m *WalletAuthManager) VerifySignature(message, signature, claimedAddress string) (string, error) {
	// Normalize claimed address
	claimedAddress = strings.ToLower(claimedAddress)
	if !common.IsHexAddress(claimedAddress) {
		return "", fmt.Errorf("invalid claimed address format")
	}

	// Decode signature
	sigBytes, err := hex.DecodeString(strings.TrimPrefix(signature, "0x"))
	if err != nil {
		return "", fmt.Errorf("invalid signature format: %w", err)
	}

	if len(sigBytes) != 65 {
		return "", fmt.Errorf("invalid signature length: expected 65, got %d", len(sigBytes))
	}

	// Ethereum signed message prefix
	prefixedMessage := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)
	hash := crypto.Keccak256Hash([]byte(prefixedMessage))

	// Adjust V value for recovery
	if sigBytes[64] >= 27 {
		sigBytes[64] -= 27
	}

	// Recover public key from signature
	pubKey, err := crypto.SigToPub(hash.Bytes(), sigBytes)
	if err != nil {
		return "", fmt.Errorf("failed to recover public key: %w", err)
	}

	recoveredAddr := crypto.PubkeyToAddress(*pubKey)
	recoveredAddrLower := strings.ToLower(recoveredAddr.Hex())

	// Verify address matches
	if recoveredAddrLower != claimedAddress {
		logging.Warn("wallet signature verification failed - address mismatch",
			"claimed", claimedAddress,
			"recovered", recoveredAddrLower,
			logging.Component("api"))
		return "", fmt.Errorf("signature does not match claimed address")
	}

	logging.Debug("wallet signature verified",
		"address", recoveredAddrLower,
		logging.Component("api"))

	return recoveredAddr.Hex(), nil
}

// VerifyInlineAuth verifies inline wallet authentication headers
// This is for stateless requests where the client signs each request
func (m *WalletAuthManager) VerifyInlineAuth(walletAddr, signature, message string) (string, error) {
	// For inline auth, verify the message timestamp is recent
	// The message format is: "moltbunker-auth:{timestamp}"
	if strings.HasPrefix(message, "moltbunker-auth:") {
		parts := strings.Split(message, ":")
		if len(parts) >= 2 {
			var timestamp int64
			fmt.Sscanf(parts[1], "%d", &timestamp)

			// Check if timestamp is within 5 minutes
			now := time.Now().Unix()
			if now-timestamp > 300 || timestamp-now > 60 {
				return "", fmt.Errorf("auth message timestamp expired or invalid")
			}
		}
	}

	return m.VerifySignature(message, signature, walletAddr)
}

// CreateSession creates a new authenticated session for a wallet
func (m *WalletAuthManager) CreateSession(address string) (string, time.Time, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Generate session token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", time.Time{}, fmt.Errorf("failed to generate session token: %w", err)
	}
	token := "wt_" + hex.EncodeToString(tokenBytes)

	expiresAt := time.Now().Add(1 * time.Hour)

	m.sessions[token] = &WalletSession{
		Address:   strings.ToLower(address),
		Token:     token,
		CreatedAt: time.Now(),
		ExpiresAt: expiresAt,
	}

	logging.Debug("wallet session created",
		"address", address,
		"expires_at", expiresAt.Format(time.RFC3339),
		logging.Component("api"))

	return token, expiresAt, nil
}

// ValidateSession validates a session token and returns the wallet address
func (m *WalletAuthManager) ValidateSession(token string) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	session, exists := m.sessions[token]
	if !exists {
		return "", false
	}

	if time.Now().After(session.ExpiresAt) {
		return "", false
	}

	return session.Address, true
}

// RevokeSession revokes a session token
func (m *WalletAuthManager) RevokeSession(token string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sessions, token)
}

// CleanupExpired removes expired challenges and sessions
func (m *WalletAuthManager) CleanupExpired() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()

	// Cleanup challenges
	for addr, challenge := range m.challenges {
		if now.After(challenge.ExpiresAt) {
			delete(m.challenges, addr)
		}
	}

	// Cleanup sessions
	for token, session := range m.sessions {
		if now.After(session.ExpiresAt) {
			delete(m.sessions, token)
		}
	}
}

// cleanupLoop periodically cleans up expired entries
func (m *WalletAuthManager) cleanupLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		m.CleanupExpired()
	}
}

// GetChallengeForAddress returns the current challenge for an address (for verification)
func (m *WalletAuthManager) GetChallengeForAddress(address string) (*Challenge, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	address = strings.ToLower(address)
	challenge, exists := m.challenges[address]
	if !exists || time.Now().After(challenge.ExpiresAt) {
		return nil, false
	}
	return challenge, true
}

// Stats returns authentication statistics
func (m *WalletAuthManager) Stats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]interface{}{
		"active_challenges": len(m.challenges),
		"active_sessions":   len(m.sessions),
	}
}
