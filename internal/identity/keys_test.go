package identity

import (
	"bytes"
	"errors"
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

func TestSaveAndLoadEncryptedKey(t *testing.T) {
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "test.key")
	encPath := filepath.Join(tmpDir, "test.key.enc")
	passphrase := []byte("strong-test-passphrase-42!")

	// Create a key manager with a fresh key pair
	km, err := NewKeyManager(keyPath)
	if err != nil {
		t.Fatalf("Failed to create key manager: %v", err)
	}

	originalPrivateKey := make([]byte, len(km.PrivateKey()))
	copy(originalPrivateKey, km.PrivateKey())
	originalNodeID := km.NodeID()

	// Save encrypted key
	if err := km.SaveEncryptedKey(encPath, passphrase); err != nil {
		t.Fatalf("Failed to save encrypted key: %v", err)
	}

	// Verify the encrypted file exists
	if _, err := os.Stat(encPath); os.IsNotExist(err) {
		t.Fatal("Encrypted key file was not created")
	}

	// Verify the encrypted file content is NOT the raw private key
	encData, err := os.ReadFile(encPath)
	if err != nil {
		t.Fatalf("Failed to read encrypted file: %v", err)
	}
	if bytes.Contains(encData, originalPrivateKey) {
		t.Error("Encrypted file contains the raw private key bytes")
	}

	// Verify the file starts with magic bytes
	if !bytes.HasPrefix(encData, encryptedKeyMagic) {
		t.Error("Encrypted file does not start with expected magic bytes")
	}

	// Load the encrypted key into a new KeyManager
	km2 := &KeyManager{}
	if err := km2.LoadEncryptedKey(encPath, passphrase); err != nil {
		t.Fatalf("Failed to load encrypted key: %v", err)
	}

	// Verify the loaded key matches the original
	if !bytes.Equal(km2.PrivateKey(), originalPrivateKey) {
		t.Error("Loaded private key does not match original")
	}

	if km2.NodeID() != originalNodeID {
		t.Error("Loaded NodeID does not match original")
	}

	// Verify sign/verify still works with loaded key
	msg := []byte("test message for encrypted key")
	sig, err := km2.Sign(msg)
	if err != nil {
		t.Fatalf("Failed to sign with loaded key: %v", err)
	}
	if !km2.Verify(msg, sig) {
		t.Error("Signature verification failed with loaded encrypted key")
	}

	// Cross-verify: original key manager can verify signature from loaded key
	if !km.Verify(msg, sig) {
		t.Error("Original key manager cannot verify signature from loaded key")
	}
}

func TestLoadEncryptedKeyWrongPassphrase(t *testing.T) {
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "test.key")
	encPath := filepath.Join(tmpDir, "test.key.enc")
	passphrase := []byte("correct-passphrase")
	wrongPassphrase := []byte("wrong-passphrase")

	km, err := NewKeyManager(keyPath)
	if err != nil {
		t.Fatalf("Failed to create key manager: %v", err)
	}

	if err := km.SaveEncryptedKey(encPath, passphrase); err != nil {
		t.Fatalf("Failed to save encrypted key: %v", err)
	}

	// Attempt to load with wrong passphrase
	km2 := &KeyManager{}
	err = km2.LoadEncryptedKey(encPath, wrongPassphrase)
	if err == nil {
		t.Fatal("Expected error when loading with wrong passphrase, got nil")
	}

	// Verify it returns the correct sentinel error
	if !errors.Is(err, ErrWrongPassphrase) {
		t.Errorf("Expected ErrWrongPassphrase, got: %v", err)
	}
}

func TestLoadEncryptedKeyCorruptedFile(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name    string
		content []byte
		wantErr error
	}{
		{
			name:    "empty file",
			content: []byte{},
			wantErr: ErrInvalidKeyFile,
		},
		{
			name:    "wrong magic bytes",
			content: []byte{0x00, 0x00, 0x00, 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1A, 0x1B, 0x1C, 0x1D, 0x1E, 0x1F, 0x20, 0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28, 0x29, 0x2A, 0x2B, 0x2C, 0x2D},
			wantErr: ErrInvalidKeyFile,
		},
		{
			name:    "file too short - only magic",
			content: []byte{0x4D, 0x42, 0x45, 0x4B},
			wantErr: ErrInvalidKeyFile,
		},
		{
			name: "correct magic but garbage ciphertext",
			// magic (4) + salt (16) + nonce (12) + 17 bytes garbage (>= minSize threshold)
			content: append(append(append(
				[]byte{0x4D, 0x42, 0x45, 0x4B},
				make([]byte, 16)...), // salt
				make([]byte, 12)...), // nonce
				make([]byte, 17)...), // garbage ciphertext (1 byte + 16 byte GCM tag size)
			wantErr: ErrWrongPassphrase, // GCM will fail to authenticate
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(tmpDir, tt.name+".key.enc")
			if err := os.WriteFile(path, tt.content, 0600); err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			km := &KeyManager{}
			err := km.LoadEncryptedKey(path, []byte("any-passphrase"))
			if err == nil {
				t.Fatal("Expected error for corrupted file, got nil")
			}

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Expected error %v, got: %v", tt.wantErr, err)
			}
		})
	}
}

func TestEncryptedKeyFilePermissions(t *testing.T) {
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "test.key")
	encPath := filepath.Join(tmpDir, "test.key.enc")
	passphrase := []byte("test-passphrase")

	km, err := NewKeyManager(keyPath)
	if err != nil {
		t.Fatalf("Failed to create key manager: %v", err)
	}

	if err := km.SaveEncryptedKey(encPath, passphrase); err != nil {
		t.Fatalf("Failed to save encrypted key: %v", err)
	}

	info, err := os.Stat(encPath)
	if err != nil {
		t.Fatalf("Failed to stat encrypted key file: %v", err)
	}

	perm := info.Mode().Perm()
	if perm != 0600 {
		t.Errorf("Encrypted key file permissions should be 0600, got %04o", perm)
	}
}

func TestIsEncryptedKeyFile(t *testing.T) {
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "test.key")
	encPath := filepath.Join(tmpDir, "test.key.enc")
	passphrase := []byte("test-passphrase")

	km, err := NewKeyManager(keyPath)
	if err != nil {
		t.Fatalf("Failed to create key manager: %v", err)
	}

	if err := km.SaveEncryptedKey(encPath, passphrase); err != nil {
		t.Fatalf("Failed to save encrypted key: %v", err)
	}

	// Encrypted file should be detected
	isEnc, err := IsEncryptedKeyFile(encPath)
	if err != nil {
		t.Fatalf("Failed to check encrypted key file: %v", err)
	}
	if !isEnc {
		t.Error("Expected encrypted key file to be detected as encrypted")
	}

	// Unencrypted file should not be detected
	isEnc, err = IsEncryptedKeyFile(keyPath)
	if err != nil {
		t.Fatalf("Failed to check unencrypted key file: %v", err)
	}
	if isEnc {
		t.Error("Expected unencrypted key file to NOT be detected as encrypted")
	}
}

func TestSaveEncryptedKeyCreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "test.key")
	encPath := filepath.Join(tmpDir, "subdir", "nested", "test.key.enc")
	passphrase := []byte("test-passphrase")

	km, err := NewKeyManager(keyPath)
	if err != nil {
		t.Fatalf("Failed to create key manager: %v", err)
	}

	if err := km.SaveEncryptedKey(encPath, passphrase); err != nil {
		t.Fatalf("Failed to save encrypted key to nested path: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(encPath); os.IsNotExist(err) {
		t.Error("Encrypted key file was not created in nested directory")
	}
}

func TestSaveEncryptedKeyNoPrivateKey(t *testing.T) {
	km := &KeyManager{}
	err := km.SaveEncryptedKey("/tmp/test.key.enc", []byte("pass"))
	if err == nil {
		t.Fatal("Expected error when saving with no private key, got nil")
	}
}
