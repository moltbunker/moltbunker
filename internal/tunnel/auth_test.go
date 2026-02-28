package tunnel

import (
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"testing"
	"time"

	"github.com/moltbunker/moltbunker/internal/p2p"
	"github.com/moltbunker/moltbunker/pkg/types"
)

func TestValidateRegistration_ValidRequest(t *testing.T) {
	nt := p2p.NewNonceTrackerWithConfig(60*time.Second, 30*time.Second, 10*time.Minute)

	var nonce [32]byte
	rand.Read(nonce[:])

	req := &TunnelRegisterRequest{
		NodeID:    "abcdef1234567890",
		Nonce:     hex.EncodeToString(nonce[:]),
		Timestamp: time.Now().Unix(),
	}

	tlsState := tls.ConnectionState{}
	if err := ValidateRegistration(req, tlsState, nt); err != nil {
		t.Fatalf("valid request rejected: %v", err)
	}
}

func TestValidateRegistration_ReplayDetection(t *testing.T) {
	nt := p2p.NewNonceTrackerWithConfig(60*time.Second, 30*time.Second, 10*time.Minute)

	var nonce [32]byte
	rand.Read(nonce[:])
	nonceHex := hex.EncodeToString(nonce[:])

	req := &TunnelRegisterRequest{
		NodeID:    "test",
		Nonce:     nonceHex,
		Timestamp: time.Now().Unix(),
	}

	tlsState := tls.ConnectionState{}

	// First request should succeed
	if err := ValidateRegistration(req, tlsState, nt); err != nil {
		t.Fatalf("first request rejected: %v", err)
	}

	// Replay should be detected
	if err := ValidateRegistration(req, tlsState, nt); err == nil {
		t.Fatal("replay not detected")
	}
}

func TestValidateRegistration_OldTimestamp(t *testing.T) {
	nt := p2p.NewNonceTrackerWithConfig(60*time.Second, 30*time.Second, 10*time.Minute)

	var nonce [32]byte
	rand.Read(nonce[:])

	req := &TunnelRegisterRequest{
		NodeID:    "test",
		Nonce:     hex.EncodeToString(nonce[:]),
		Timestamp: time.Now().Add(-5 * time.Minute).Unix(), // 5 min old
	}

	tlsState := tls.ConnectionState{}
	if err := ValidateRegistration(req, tlsState, nt); err == nil {
		t.Fatal("old timestamp not rejected")
	}
}

func TestValidateRegistration_InvalidNonce(t *testing.T) {
	nt := p2p.NewNonceTrackerWithConfig(60*time.Second, 30*time.Second, 10*time.Minute)

	req := &TunnelRegisterRequest{
		NodeID:    "test",
		Nonce:     "tooshort",
		Timestamp: time.Now().Unix(),
	}

	tlsState := tls.ConnectionState{}
	if err := ValidateRegistration(req, tlsState, nt); err == nil {
		t.Fatal("short nonce not rejected")
	}
}

func TestIssueAndValidateReconnToken(t *testing.T) {
	secret := make([]byte, 32)
	rand.Read(secret)

	nodeID := testNodeID(42)
	subdomain := "abc12345"

	token := IssueReconnToken(secret, nodeID, subdomain)
	if token == "" {
		t.Fatal("empty token")
	}

	// Valid token
	if !ValidateReconnToken(secret, token, nodeID, subdomain, 1*time.Hour) {
		t.Fatal("valid token rejected")
	}

	// Wrong nodeID
	wrongNodeID := testNodeID(99)
	if ValidateReconnToken(secret, token, wrongNodeID, subdomain, 1*time.Hour) {
		t.Fatal("wrong nodeID accepted")
	}

	// Wrong subdomain
	if ValidateReconnToken(secret, token, nodeID, "different", 1*time.Hour) {
		t.Fatal("wrong subdomain accepted")
	}

	// Wrong secret
	wrongSecret := make([]byte, 32)
	rand.Read(wrongSecret)
	if ValidateReconnToken(wrongSecret, token, nodeID, subdomain, 1*time.Hour) {
		t.Fatal("wrong secret accepted")
	}
}

func TestReconnToken_Expiry(t *testing.T) {
	secret := make([]byte, 32)
	rand.Read(secret)

	nodeID := testNodeID(1)
	token := IssueReconnToken(secret, nodeID, "sub1")

	// With very short maxAge — token was just issued so should be valid
	if !ValidateReconnToken(secret, token, nodeID, "sub1", 5*time.Second) {
		t.Fatal("fresh token should be valid with 5s maxAge")
	}

	// Tampered token: corrupt the HMAC portion by replacing first char
	tampered := "f" + token[1:]
	if token[0] == 'f' {
		tampered = "a" + token[1:]
	}
	if ValidateReconnToken(secret, tampered, nodeID, "sub1", 1*time.Hour) {
		t.Fatal("tampered token accepted")
	}
}

func TestMaxSubdomainsForTier(t *testing.T) {
	tests := []struct {
		tier string
		want int
	}{
		{"free", 3},
		{"starter", 10},
		{"bronze", 25},
		{"silver", 100},
		{"gold", 100},
		{"platinum", 100},
		{"unknown", 3},
	}

	for _, tt := range tests {
		if got := MaxSubdomainsForTier(tt.tier); got != tt.want {
			t.Errorf("MaxSubdomainsForTier(%q) = %d, want %d", tt.tier, got, tt.want)
		}
	}
}

func TestLimitsForTier(t *testing.T) {
	free := LimitsForTier("free")
	if free.MaxRPS != 10 {
		t.Errorf("free MaxRPS = %d, want 10", free.MaxRPS)
	}

	starter := LimitsForTier("starter")
	if starter.MaxRPS != 50 {
		t.Errorf("starter MaxRPS = %d, want 50", starter.MaxRPS)
	}

	platinum := LimitsForTier("platinum")
	if platinum.MaxRPS != 2000 {
		t.Errorf("platinum MaxRPS = %d, want 2000", platinum.MaxRPS)
	}
}

// Suppress unused types import
var _ types.NodeID
