package security

import (
	"bytes"
	"testing"
)

func TestGenerateX25519KeyPair(t *testing.T) {
	pubKey, privKey, err := GenerateX25519KeyPair()
	if err != nil {
		t.Fatalf("failed to generate key pair: %v", err)
	}

	if len(pubKey) != X25519KeySize {
		t.Errorf("expected public key size %d, got %d", X25519KeySize, len(pubKey))
	}

	if len(privKey) != X25519KeySize {
		t.Errorf("expected private key size %d, got %d", X25519KeySize, len(privKey))
	}

	// Keys should be different
	if bytes.Equal(pubKey, privKey) {
		t.Error("public and private keys should be different")
	}
}

func TestDeploymentEncryptionManager_SetupAndEncrypt(t *testing.T) {
	dem := NewDeploymentEncryptionManager("")

	// Generate requester's key pair
	requesterPubKey, requesterPrivKey, err := GenerateX25519KeyPair()
	if err != nil {
		t.Fatalf("failed to generate requester keys: %v", err)
	}

	// Setup deployment encryption
	deploymentID := "test-deployment-123"
	encInfo, err := dem.SetupDeploymentEncryption(deploymentID, requesterPubKey)
	if err != nil {
		t.Fatalf("failed to setup deployment encryption: %v", err)
	}

	if encInfo.DeploymentID != deploymentID {
		t.Errorf("expected deployment ID %s, got %s", deploymentID, encInfo.DeploymentID)
	}

	if encInfo.DEKAlgorithm != "AES-256-GCM" {
		t.Errorf("expected algorithm AES-256-GCM, got %s", encInfo.DEKAlgorithm)
	}

	// Encrypt some data
	plaintext := []byte("This is secret data that only the requester should see")
	ciphertext, err := dem.EncryptData(deploymentID, plaintext)
	if err != nil {
		t.Fatalf("failed to encrypt data: %v", err)
	}

	if bytes.Equal(plaintext, ciphertext) {
		t.Error("ciphertext should not equal plaintext")
	}

	// Verify that provider can decrypt (for verification purposes during processing)
	decrypted, err := dem.DecryptData(deploymentID, ciphertext)
	if err != nil {
		t.Fatalf("failed to decrypt data: %v", err)
	}

	if !bytes.Equal(plaintext, decrypted) {
		t.Error("decrypted data does not match original")
	}

	// Now test requester decryption
	metadata, err := dem.GetEncryptionMetadata(deploymentID)
	if err != nil {
		t.Fatalf("failed to get encryption metadata: %v", err)
	}

	requesterDecryptor, err := NewRequesterDecryptor(requesterPrivKey, requesterPubKey)
	if err != nil {
		t.Fatalf("failed to create requester decryptor: %v", err)
	}

	requesterDecrypted, err := requesterDecryptor.DecryptOutput(metadata, ciphertext)
	if err != nil {
		t.Fatalf("requester failed to decrypt: %v", err)
	}

	if !bytes.Equal(plaintext, requesterDecrypted) {
		t.Error("requester decrypted data does not match original")
	}
}

func TestDeploymentEncryptionManager_MultipleDeployments(t *testing.T) {
	dem := NewDeploymentEncryptionManager("")

	// Create two different deployments with different requesters
	requesterPubKey1, requesterPrivKey1, _ := GenerateX25519KeyPair()
	requesterPubKey2, requesterPrivKey2, _ := GenerateX25519KeyPair()

	_, err := dem.SetupDeploymentEncryption("deployment-1", requesterPubKey1)
	if err != nil {
		t.Fatalf("failed to setup deployment 1: %v", err)
	}

	_, err = dem.SetupDeploymentEncryption("deployment-2", requesterPubKey2)
	if err != nil {
		t.Fatalf("failed to setup deployment 2: %v", err)
	}

	// Encrypt data for each deployment
	data1 := []byte("Secret data for deployment 1")
	data2 := []byte("Secret data for deployment 2")

	cipher1, err := dem.EncryptData("deployment-1", data1)
	if err != nil {
		t.Fatalf("failed to encrypt for deployment 1: %v", err)
	}

	cipher2, err := dem.EncryptData("deployment-2", data2)
	if err != nil {
		t.Fatalf("failed to encrypt for deployment 2: %v", err)
	}

	// Verify each requester can only decrypt their own data
	metadata1, _ := dem.GetEncryptionMetadata("deployment-1")
	metadata2, _ := dem.GetEncryptionMetadata("deployment-2")

	decryptor1, _ := NewRequesterDecryptor(requesterPrivKey1, requesterPubKey1)
	decryptor2, _ := NewRequesterDecryptor(requesterPrivKey2, requesterPubKey2)

	// Requester 1 can decrypt deployment 1's data
	decrypted1, err := decryptor1.DecryptOutput(metadata1, cipher1)
	if err != nil {
		t.Fatalf("requester 1 failed to decrypt deployment 1: %v", err)
	}
	if !bytes.Equal(data1, decrypted1) {
		t.Error("requester 1 decrypted wrong data")
	}

	// Requester 2 can decrypt deployment 2's data
	decrypted2, err := decryptor2.DecryptOutput(metadata2, cipher2)
	if err != nil {
		t.Fatalf("requester 2 failed to decrypt deployment 2: %v", err)
	}
	if !bytes.Equal(data2, decrypted2) {
		t.Error("requester 2 decrypted wrong data")
	}

	// Requester 1 should NOT be able to decrypt deployment 2's data
	_, err = decryptor1.DecryptOutput(metadata2, cipher2)
	if err == nil {
		t.Error("requester 1 should NOT be able to decrypt deployment 2's data")
	}

	// Requester 2 should NOT be able to decrypt deployment 1's data
	_, err = decryptor2.DecryptOutput(metadata1, cipher1)
	if err == nil {
		t.Error("requester 2 should NOT be able to decrypt deployment 1's data")
	}
}

func TestDeploymentEncryptionManager_RemoveDeployment(t *testing.T) {
	dem := NewDeploymentEncryptionManager("")
	requesterPubKey, _, _ := GenerateX25519KeyPair()

	deploymentID := "test-deployment"
	_, err := dem.SetupDeploymentEncryption(deploymentID, requesterPubKey)
	if err != nil {
		t.Fatalf("failed to setup deployment: %v", err)
	}

	// Verify deployment exists
	_, exists := dem.GetDeploymentKeys(deploymentID)
	if !exists {
		t.Error("deployment should exist")
	}

	// Remove deployment
	dem.RemoveDeployment(deploymentID)

	// Verify deployment is removed
	_, exists = dem.GetDeploymentKeys(deploymentID)
	if exists {
		t.Error("deployment should not exist after removal")
	}

	// Encrypting should fail now
	_, err = dem.EncryptData(deploymentID, []byte("test"))
	if err == nil {
		t.Error("encryption should fail for removed deployment")
	}
}

func TestDeploymentEncryptionManager_InvalidRequesterKey(t *testing.T) {
	dem := NewDeploymentEncryptionManager("")

	// Try with invalid key size
	_, err := dem.SetupDeploymentEncryption("test", []byte("too-short"))
	if err == nil {
		t.Error("should fail with invalid key size")
	}

	// Try with empty key
	_, err = dem.SetupDeploymentEncryption("test", nil)
	if err == nil {
		t.Error("should fail with nil key")
	}
}

func TestCreateEncryptedOutput(t *testing.T) {
	dem := NewDeploymentEncryptionManager("")
	requesterPubKey, requesterPrivKey, _ := GenerateX25519KeyPair()

	deploymentID := "test-deployment"
	_, err := dem.SetupDeploymentEncryption(deploymentID, requesterPubKey)
	if err != nil {
		t.Fatalf("failed to setup deployment: %v", err)
	}

	// Create encrypted output
	logData := []byte("2024-01-15T10:30:00Z INFO Application started successfully")
	output, err := dem.CreateEncryptedOutput(deploymentID, logData, "log")
	if err != nil {
		t.Fatalf("failed to create encrypted output: %v", err)
	}

	if output.OutputType != "log" {
		t.Errorf("expected output type 'log', got %s", output.OutputType)
	}

	if output.Metadata == nil {
		t.Error("metadata should not be nil")
	}

	// Requester should be able to decrypt
	decryptor, _ := NewRequesterDecryptor(requesterPrivKey, requesterPubKey)
	decrypted, err := decryptor.DecryptOutput(output.Metadata, output.EncryptedData)
	if err != nil {
		t.Fatalf("failed to decrypt output: %v", err)
	}

	if !bytes.Equal(logData, decrypted) {
		t.Error("decrypted output does not match original")
	}
}

func TestRequesterDecryptor_InvalidKeys(t *testing.T) {
	_, err := NewRequesterDecryptor([]byte("short"), []byte("key"))
	if err == nil {
		t.Error("should fail with invalid key sizes")
	}

	validKey := make([]byte, X25519KeySize)
	_, err = NewRequesterDecryptor(validKey, []byte("short"))
	if err == nil {
		t.Error("should fail with invalid public key size")
	}
}

func BenchmarkEncryptData(b *testing.B) {
	dem := NewDeploymentEncryptionManager("")
	requesterPubKey, _, _ := GenerateX25519KeyPair()

	dem.SetupDeploymentEncryption("bench-deployment", requesterPubKey)
	data := make([]byte, 1024) // 1KB of data

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dem.EncryptData("bench-deployment", data)
	}
}

func BenchmarkDecryptOutput(b *testing.B) {
	dem := NewDeploymentEncryptionManager("")
	requesterPubKey, requesterPrivKey, _ := GenerateX25519KeyPair()

	dem.SetupDeploymentEncryption("bench-deployment", requesterPubKey)
	data := make([]byte, 1024) // 1KB of data
	ciphertext, _ := dem.EncryptData("bench-deployment", data)
	metadata, _ := dem.GetEncryptionMetadata("bench-deployment")
	decryptor, _ := NewRequesterDecryptor(requesterPrivKey, requesterPubKey)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		decryptor.DecryptOutput(metadata, ciphertext)
	}
}
