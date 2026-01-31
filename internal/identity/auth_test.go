package identity

import (
	"crypto/ed25519"
	"crypto/rand"
	"testing"
	"time"

	"github.com/moltbunker/moltbunker/pkg/types"
)

func TestAuthManager_CreateAuthToken(t *testing.T) {
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	km := &KeyManager{
		privateKey: privateKey,
		publicKey:  privateKey.Public().(ed25519.PublicKey),
	}

	am := NewAuthManager(km)

	var nodeID types.NodeID
	rand.Read(nodeID[:])

	timestamp := time.Now()
	token, err := am.CreateAuthToken(nodeID, timestamp)
	if err != nil {
		t.Fatalf("Failed to create auth token: %v", err)
	}

	if len(token) != 104 { // 32 (nodeID) + 8 (timestamp) + 64 (signature)
		t.Errorf("Token length should be 104 bytes, got %d", len(token))
	}
}

func TestAuthManager_VerifyAuthToken(t *testing.T) {
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	km := &KeyManager{
		privateKey: privateKey,
		publicKey:  privateKey.Public().(ed25519.PublicKey),
	}

	am := NewAuthManager(km)

	var nodeID types.NodeID
	rand.Read(nodeID[:])

	timestamp := time.Now()
	token, err := am.CreateAuthToken(nodeID, timestamp)
	if err != nil {
		t.Fatalf("Failed to create auth token: %v", err)
	}

	verifiedNodeID, verifiedTimestamp, err := am.VerifyAuthToken(token, km.PublicKey())
	if err != nil {
		t.Fatalf("Failed to verify auth token: %v", err)
	}

	if verifiedNodeID != nodeID {
		t.Error("Verified node ID doesn't match")
	}

	if verifiedTimestamp.Unix() != timestamp.Unix() {
		t.Error("Verified timestamp doesn't match")
	}
}

func TestAuthManager_VerifyAuthToken_Expired(t *testing.T) {
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	km := &KeyManager{
		privateKey: privateKey,
		publicKey:  privateKey.Public().(ed25519.PublicKey),
	}

	am := NewAuthManager(km)

	var nodeID types.NodeID
	rand.Read(nodeID[:])

	// Create token with old timestamp (6 minutes ago)
	timestamp := time.Now().Add(-6 * time.Minute)
	token, err := am.CreateAuthToken(nodeID, timestamp)
	if err != nil {
		t.Fatalf("Failed to create auth token: %v", err)
	}

	_, _, err = am.VerifyAuthToken(token, km.PublicKey())
	if err == nil {
		t.Error("Should fail verification for expired token")
	}
}

func TestAuthManager_SignVerifyMessage(t *testing.T) {
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	km := &KeyManager{
		privateKey: privateKey,
		publicKey:  privateKey.Public().(ed25519.PublicKey),
	}

	am := NewAuthManager(km)

	message := []byte("test message")
	signature, err := am.SignMessage(message)
	if err != nil {
		t.Fatalf("Failed to sign message: %v", err)
	}

	if !am.VerifyMessage(message, signature, km.PublicKey()) {
		t.Error("Message verification failed")
	}

	// Test with wrong message
	if am.VerifyMessage([]byte("wrong message"), signature, km.PublicKey()) {
		t.Error("Message verification should fail with wrong message")
	}
}

func TestGenerateNonce(t *testing.T) {
	nonce, err := GenerateNonce()
	if err != nil {
		t.Fatalf("Failed to generate nonce: %v", err)
	}

	if len(nonce) != 24 {
		t.Errorf("Nonce should be 24 bytes, got %d", len(nonce))
	}
}
