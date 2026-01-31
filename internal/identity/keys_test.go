package identity

import (
	"os"
	"path/filepath"
	"testing"
)

func TestKeyManager_GenerateKeys(t *testing.T) {
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "test.key")

	km, err := NewKeyManager(keyPath)
	if err != nil {
		t.Fatalf("Failed to create key manager: %v", err)
	}

	// Verify keys were generated
	if km.PrivateKey() == nil {
		t.Error("Private key is nil")
	}

	if km.PublicKey() == nil {
		t.Error("Public key is nil")
	}

	if len(km.NodeID()) != 32 {
		t.Errorf("NodeID should be 32 bytes, got %d", len(km.NodeID()))
	}

	// Verify key file exists
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		t.Error("Key file was not created")
	}
}

func TestKeyManager_LoadKeys(t *testing.T) {
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "test.key")

	// Create first key manager
	km1, err := NewKeyManager(keyPath)
	if err != nil {
		t.Fatalf("Failed to create key manager: %v", err)
	}

	nodeID1 := km1.NodeID()

	// Create second key manager loading from same path
	km2, err := NewKeyManager(keyPath)
	if err != nil {
		t.Fatalf("Failed to load key manager: %v", err)
	}

	nodeID2 := km2.NodeID()

	// Verify same keys loaded
	if nodeID1 != nodeID2 {
		t.Error("Loaded keys don't match original keys")
	}
}

func TestKeyManager_SignVerify(t *testing.T) {
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "test.key")

	km, err := NewKeyManager(keyPath)
	if err != nil {
		t.Fatalf("Failed to create key manager: %v", err)
	}

	message := []byte("test message")
	signature, err := km.Sign(message)
	if err != nil {
		t.Fatalf("Failed to sign message: %v", err)
	}

	if !km.Verify(message, signature) {
		t.Error("Signature verification failed")
	}

	// Test with wrong message
	if km.Verify([]byte("wrong message"), signature) {
		t.Error("Signature verification should fail with wrong message")
	}
}

func TestKeyManager_NodeIDString(t *testing.T) {
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "test.key")

	km, err := NewKeyManager(keyPath)
	if err != nil {
		t.Fatalf("Failed to create key manager: %v", err)
	}

	nodeIDStr := km.NodeIDString()
	if len(nodeIDStr) != 64 { // 32 bytes = 64 hex chars
		t.Errorf("NodeID string should be 64 chars, got %d", len(nodeIDStr))
	}
}
