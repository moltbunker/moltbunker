package security

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRequesterKeyManager_GenerateAndSaveKeys(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "moltbunker-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	keyStorePath := filepath.Join(tmpDir, "keys.json")
	rkm := NewRequesterKeyManager(keyStorePath)

	// Generate new keys
	if err := rkm.GenerateNewKeys(); err != nil {
		t.Fatalf("failed to generate keys: %v", err)
	}

	// Get public key
	pubKey, err := rkm.GetPublicKey()
	if err != nil {
		t.Fatalf("failed to get public key: %v", err)
	}

	if len(pubKey) != X25519KeySize {
		t.Errorf("expected public key size %d, got %d", X25519KeySize, len(pubKey))
	}

	// Save keys with passphrase
	passphrase := "test-passphrase-12345"
	if err := rkm.SaveKeys(passphrase); err != nil {
		t.Fatalf("failed to save keys: %v", err)
	}

	// Verify file exists
	if !rkm.KeysExist() {
		t.Error("key store file should exist")
	}

	// Check file permissions (Unix only)
	info, err := os.Stat(keyStorePath)
	if err != nil {
		t.Fatalf("failed to stat key store: %v", err)
	}
	mode := info.Mode().Perm()
	if mode != 0600 {
		t.Errorf("expected file permissions 0600, got %o", mode)
	}
}

func TestRequesterKeyManager_LoadKeys(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "moltbunker-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	keyStorePath := filepath.Join(tmpDir, "keys.json")
	passphrase := "my-secret-passphrase"

	// First manager: generate and save
	rkm1 := NewRequesterKeyManager(keyStorePath)
	if err := rkm1.GenerateNewKeys(); err != nil {
		t.Fatalf("failed to generate keys: %v", err)
	}
	originalPubKey, _ := rkm1.GetPublicKey()
	if err := rkm1.SaveKeys(passphrase); err != nil {
		t.Fatalf("failed to save keys: %v", err)
	}

	// Second manager: load keys
	rkm2 := NewRequesterKeyManager(keyStorePath)
	if err := rkm2.LoadKeys(passphrase); err != nil {
		t.Fatalf("failed to load keys: %v", err)
	}

	loadedPubKey, err := rkm2.GetPublicKey()
	if err != nil {
		t.Fatalf("failed to get loaded public key: %v", err)
	}

	// Verify keys match
	if len(loadedPubKey) != len(originalPubKey) {
		t.Error("loaded public key has different length")
	}
	for i := range loadedPubKey {
		if loadedPubKey[i] != originalPubKey[i] {
			t.Error("loaded public key does not match original")
			break
		}
	}
}

func TestRequesterKeyManager_WrongPassphrase(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "moltbunker-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	keyStorePath := filepath.Join(tmpDir, "keys.json")

	// Create and save with one passphrase
	rkm1 := NewRequesterKeyManager(keyStorePath)
	rkm1.GenerateNewKeys()
	rkm1.SaveKeys("correct-passphrase")

	// Try to load with wrong passphrase
	rkm2 := NewRequesterKeyManager(keyStorePath)
	err = rkm2.LoadKeys("wrong-passphrase")
	if err == nil {
		t.Error("should fail with wrong passphrase")
	}
}

func TestRequesterKeyManager_CreateDecryptor(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "moltbunker-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	keyStorePath := filepath.Join(tmpDir, "keys.json")

	rkm := NewRequesterKeyManager(keyStorePath)

	// Should fail without keys
	_, err = rkm.CreateDecryptor()
	if err == nil {
		t.Error("should fail without keys")
	}

	// Generate keys
	if err := rkm.GenerateNewKeys(); err != nil {
		t.Fatalf("failed to generate keys: %v", err)
	}

	// Now should succeed
	decryptor, err := rkm.CreateDecryptor()
	if err != nil {
		t.Fatalf("failed to create decryptor: %v", err)
	}

	if decryptor == nil {
		t.Error("decryptor should not be nil")
	}
}

func TestRequesterKeyManager_EndToEndEncryption(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "moltbunker-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	keyStorePath := filepath.Join(tmpDir, "keys.json")
	passphrase := "end-to-end-test"

	// Setup requester
	rkm := NewRequesterKeyManager(keyStorePath)
	if err := rkm.GenerateNewKeys(); err != nil {
		t.Fatalf("failed to generate keys: %v", err)
	}
	if err := rkm.SaveKeys(passphrase); err != nil {
		t.Fatalf("failed to save keys: %v", err)
	}

	requesterPubKey, _ := rkm.GetPublicKey()

	// Setup provider's deployment encryption
	dem := NewDeploymentEncryptionManager("")
	deploymentID := "e2e-test-deployment"
	_, err = dem.SetupDeploymentEncryption(deploymentID, requesterPubKey)
	if err != nil {
		t.Fatalf("failed to setup deployment encryption: %v", err)
	}

	// Provider encrypts some output
	secretData := []byte("This is the job output that only the requester should see")
	output, err := dem.CreateEncryptedOutput(deploymentID, secretData, "stdout")
	if err != nil {
		t.Fatalf("failed to create encrypted output: %v", err)
	}

	// Simulate requester loading keys from storage
	rkm2 := NewRequesterKeyManager(keyStorePath)
	if err := rkm2.LoadKeys(passphrase); err != nil {
		t.Fatalf("failed to load keys: %v", err)
	}

	// Requester decrypts the output
	decryptor, err := rkm2.CreateDecryptor()
	if err != nil {
		t.Fatalf("failed to create decryptor: %v", err)
	}

	decrypted, err := decryptor.DecryptOutput(output.Metadata, output.EncryptedData)
	if err != nil {
		t.Fatalf("failed to decrypt output: %v", err)
	}

	// Verify data matches
	if string(decrypted) != string(secretData) {
		t.Errorf("decrypted data does not match: got %q, want %q", decrypted, secretData)
	}
}

func TestRequesterKeyManager_DeleteKeys(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "moltbunker-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	keyStorePath := filepath.Join(tmpDir, "keys.json")

	rkm := NewRequesterKeyManager(keyStorePath)
	rkm.GenerateNewKeys()
	rkm.SaveKeys("passphrase")

	if !rkm.KeysExist() {
		t.Error("key store should exist before deletion")
	}

	if err := rkm.DeleteKeys(); err != nil {
		t.Fatalf("failed to delete keys: %v", err)
	}

	if rkm.KeysExist() {
		t.Error("key store should not exist after deletion")
	}

	// Getting public key should fail
	_, err = rkm.GetPublicKey()
	if err == nil {
		t.Error("should fail to get public key after deletion")
	}
}

func TestRequesterKeyManager_PublicKeyHex(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "moltbunker-test-*")
	defer os.RemoveAll(tmpDir)

	rkm := NewRequesterKeyManager(filepath.Join(tmpDir, "keys.json"))
	rkm.GenerateNewKeys()

	hex, err := rkm.PublicKeyHex()
	if err != nil {
		t.Fatalf("failed to get public key hex: %v", err)
	}

	// X25519 key is 32 bytes, hex is 64 chars
	if len(hex) != 64 {
		t.Errorf("expected hex length 64, got %d", len(hex))
	}
}
