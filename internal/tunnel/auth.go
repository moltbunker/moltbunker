package tunnel

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"crypto/tls"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/moltbunker/moltbunker/internal/p2p"
	"github.com/moltbunker/moltbunker/pkg/types"
)

const (
	// maxRegistrationAge is the maximum age of a registration request.
	maxRegistrationAge = 60 * time.Second
)

// ValidateRegistration validates a TunnelRegisterRequest.
// It checks the nonce (replay protection), timestamp freshness, and TLS channel binding.
func ValidateRegistration(req *TunnelRegisterRequest, tlsState tls.ConnectionState,
	nonceTracker *p2p.NonceTracker) error {

	// Validate nonce format (must be 32-byte hex = 64 chars)
	if len(req.Nonce) != 64 {
		return fmt.Errorf("invalid nonce length: expected 64 hex chars, got %d", len(req.Nonce))
	}
	nonceBytes, err := hex.DecodeString(req.Nonce)
	if err != nil {
		return fmt.Errorf("invalid nonce hex: %w", err)
	}

	// Check nonce + timestamp via NonceTracker (replay protection)
	var nonce [24]byte
	copy(nonce[:], nonceBytes[:24])
	ts := time.Unix(req.Timestamp, 0)
	if err := nonceTracker.Check(nonce, ts); err != nil {
		return fmt.Errorf("nonce validation failed: %w", err)
	}

	// Validate timestamp freshness
	age := time.Since(ts)
	if age > maxRegistrationAge {
		return fmt.Errorf("registration too old: %s", age.Round(time.Second))
	}
	if age < -maxRegistrationAge {
		return fmt.Errorf("registration timestamp in future: %s", (-age).Round(time.Second))
	}

	// Validate TLS channel binding (prevents token relay attacks)
	if req.TLSBinding != "" {
		// TLSUnique may not be available with TLS 1.3 session tickets.
		// When available, verify it matches.
		if len(tlsState.TLSUnique) > 0 {
			expectedBinding := hex.EncodeToString(tlsState.TLSUnique)
			if subtle.ConstantTimeCompare([]byte(req.TLSBinding), []byte(expectedBinding)) != 1 {
				return fmt.Errorf("TLS channel binding mismatch")
			}
		}
	}

	return nil
}

// IssueReconnToken creates an HMAC-SHA256 reconnection token bound to a NodeID and subdomain.
// Format: hex(HMAC-SHA256(secret, nodeID || subdomain || issuedAt))
func IssueReconnToken(secret []byte, nodeID types.NodeID, subdomain string) string {
	now := time.Now().Unix()
	mac := hmac.New(sha256.New, secret)
	mac.Write(nodeID[:])
	mac.Write([]byte(subdomain))
	var ts [8]byte
	binary.BigEndian.PutUint64(ts[:], uint64(now))
	mac.Write(ts[:])
	return hex.EncodeToString(mac.Sum(nil)) + fmt.Sprintf(":%d", now)
}

// ValidateReconnToken verifies a reconnection token.
// Returns true if the token is valid, not expired, and matches the nodeID + subdomain.
func ValidateReconnToken(secret []byte, token string, nodeID types.NodeID,
	subdomain string, maxAge time.Duration) bool {

	// Parse token: hex_hmac:timestamp
	var hmacHex string
	var issuedAt int64
	n, err := fmt.Sscanf(token, "%64s:%d", &hmacHex, &issuedAt)
	if err != nil || n != 2 {
		// Try alternate parsing (the %64s may grab the colon)
		for i := len(token) - 1; i >= 0; i-- {
			if token[i] == ':' {
				hmacHex = token[:i]
				// Best-effort parse; the issuedAt == 0 guard below rejects failures.
				_, _ = fmt.Sscanf(token[i+1:], "%d", &issuedAt)
				break
			}
		}
		if hmacHex == "" || issuedAt == 0 {
			return false
		}
	}

	// Check age
	age := time.Since(time.Unix(issuedAt, 0))
	if age > maxAge || age < -maxRegistrationAge {
		return false
	}

	// Recompute HMAC
	mac := hmac.New(sha256.New, secret)
	mac.Write(nodeID[:])
	mac.Write([]byte(subdomain))
	var ts [8]byte
	binary.BigEndian.PutUint64(ts[:], uint64(issuedAt))
	mac.Write(ts[:])
	expected := hex.EncodeToString(mac.Sum(nil))

	return subtle.ConstantTimeCompare([]byte(hmacHex), []byte(expected)) == 1
}
