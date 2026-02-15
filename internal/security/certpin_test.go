package security

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"path/filepath"
	"testing"
	"time"
)

func createTestCertificate() (*x509.Certificate, error) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Test"},
		},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA: false,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, publicKey, privateKey)
	if err != nil {
		return nil, err
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return nil, err
	}

	return cert, nil
}

func TestCertPinStore_PinCertificate(t *testing.T) {
	cps := NewCertPinStore()

	cert, err := createTestCertificate()
	if err != nil {
		t.Fatalf("Failed to create test certificate: %v", err)
	}

	cps.PinCertificate("node1", cert)

	pinnedHash, exists := cps.GetPin("node1")
	if !exists {
		t.Error("Certificate should be pinned")
	}

	if len(pinnedHash) == 0 {
		t.Error("Pinned hash should not be empty")
	}
}

func TestCertPinStore_VerifyCertificate(t *testing.T) {
	cps := NewCertPinStore()

	cert, err := createTestCertificate()
	if err != nil {
		t.Fatalf("Failed to create test certificate: %v", err)
	}

	// Pin certificate
	cps.PinCertificate("node1", cert)

	// Verify same certificate
	err = cps.VerifyCertificate("node1", cert)
	if err != nil {
		t.Errorf("Certificate verification failed: %v", err)
	}
}

func TestCertPinStore_VerifyCertificate_DifferentCert(t *testing.T) {
	cps := NewCertPinStore()

	cert1, err := createTestCertificate()
	if err != nil {
		t.Fatalf("Failed to create test certificate: %v", err)
	}

	cert2, err := createTestCertificate()
	if err != nil {
		t.Fatalf("Failed to create test certificate: %v", err)
	}

	cps.PinCertificate("node1", cert1)

	// Try to verify with different certificate
	err = cps.VerifyCertificate("node1", cert2)
	if err == nil {
		t.Error("Should fail verification with different certificate")
	}
}

func TestCertPinStore_VerifyCertificate_FirstTime(t *testing.T) {
	cps := NewCertPinStore()

	cert, err := createTestCertificate()
	if err != nil {
		t.Fatalf("Failed to create test certificate: %v", err)
	}

	// First time seeing this node - should auto-pin
	err = cps.VerifyCertificate("node1", cert)
	if err != nil {
		t.Errorf("First-time verification should succeed and auto-pin: %v", err)
	}

	// Verify it was pinned
	_, exists := cps.GetPin("node1")
	if !exists {
		t.Error("Certificate should have been auto-pinned")
	}
}

func TestCertPinStore_RemovePin(t *testing.T) {
	cps := NewCertPinStore()

	cert, err := createTestCertificate()
	if err != nil {
		t.Fatalf("Failed to create test certificate: %v", err)
	}

	cps.PinCertificate("node1", cert)

	cps.RemovePin("node1")

	_, exists := cps.GetPin("node1")
	if exists {
		t.Error("Pin should have been removed")
	}
}

func TestCertPinStore_GetPin_NotExists(t *testing.T) {
	cps := NewCertPinStore()

	_, exists := cps.GetPin("nonexistent")
	if exists {
		t.Error("Pin should not exist for nonexistent node")
	}
}

func TestCertPinStore_PinCount(t *testing.T) {
	cps := NewCertPinStore()

	if cps.PinCount() != 0 {
		t.Errorf("Expected 0 pins, got %d", cps.PinCount())
	}

	cert1, err := createTestCertificate()
	if err != nil {
		t.Fatalf("Failed to create test certificate: %v", err)
	}
	cert2, err := createTestCertificate()
	if err != nil {
		t.Fatalf("Failed to create test certificate: %v", err)
	}

	cps.PinCertificate("node1", cert1)
	if cps.PinCount() != 1 {
		t.Errorf("Expected 1 pin, got %d", cps.PinCount())
	}

	cps.PinCertificate("node2", cert2)
	if cps.PinCount() != 2 {
		t.Errorf("Expected 2 pins, got %d", cps.PinCount())
	}

	cps.RemovePin("node1")
	if cps.PinCount() != 1 {
		t.Errorf("Expected 1 pin after removal, got %d", cps.PinCount())
	}
}

func TestCertPinStore_SaveLoad(t *testing.T) {
	tmpDir := t.TempDir()
	pinPath := filepath.Join(tmpDir, "certpins.json")

	// Create store and pin certificates
	cps := NewCertPinStore()

	cert1, err := createTestCertificate()
	if err != nil {
		t.Fatalf("Failed to create test certificate: %v", err)
	}
	cert2, err := createTestCertificate()
	if err != nil {
		t.Fatalf("Failed to create test certificate: %v", err)
	}

	cps.PinCertificate("node1", cert1)
	cps.PinCertificate("node2", cert2)

	// Save to disk
	if err := cps.Save(pinPath); err != nil {
		t.Fatalf("Failed to save cert pin store: %v", err)
	}

	// Load into a new store
	cps2 := NewCertPinStore()
	if err := cps2.Load(pinPath); err != nil {
		t.Fatalf("Failed to load cert pin store: %v", err)
	}

	// Verify pin count matches
	if cps2.PinCount() != 2 {
		t.Errorf("Expected 2 pins after load, got %d", cps2.PinCount())
	}

	// Verify that the loaded store can verify the same certificates
	if err := cps2.VerifyCertificate("node1", cert1); err != nil {
		t.Errorf("Verification of node1 cert should succeed after load: %v", err)
	}
	if err := cps2.VerifyCertificate("node2", cert2); err != nil {
		t.Errorf("Verification of node2 cert should succeed after load: %v", err)
	}

	// Verify that the loaded store rejects a different certificate for node1
	if err := cps2.VerifyCertificate("node1", cert2); err == nil {
		t.Error("Verification should fail with wrong certificate after load")
	}
}

func TestCertPinStore_LoadNonExistent(t *testing.T) {
	cps := NewCertPinStore()

	// Loading from a non-existent file should not be an error
	err := cps.Load("/nonexistent/path/certpins.json")
	if err != nil {
		t.Errorf("Load from non-existent file should not error: %v", err)
	}

	if cps.PinCount() != 0 {
		t.Error("Pin count should be 0 after loading non-existent file")
	}
}

func TestCertPinStore_SaveLoadEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	pinPath := filepath.Join(tmpDir, "certpins.json")

	cps := NewCertPinStore()

	// Save empty store
	if err := cps.Save(pinPath); err != nil {
		t.Fatalf("Failed to save empty cert pin store: %v", err)
	}

	// Load into a new store
	cps2 := NewCertPinStore()
	if err := cps2.Load(pinPath); err != nil {
		t.Fatalf("Failed to load empty cert pin store: %v", err)
	}

	if cps2.PinCount() != 0 {
		t.Errorf("Expected 0 pins after loading empty store, got %d", cps2.PinCount())
	}
}

func TestCertPinStore_PinVerifyRoundtrip(t *testing.T) {
	cps := NewCertPinStore()

	cert, err := createTestCertificate()
	if err != nil {
		t.Fatalf("Failed to create test certificate: %v", err)
	}

	// Pin the certificate
	cps.PinCertificate("node1", cert)

	// Verify same certificate succeeds
	if err := cps.VerifyCertificate("node1", cert); err != nil {
		t.Errorf("Pin/verify roundtrip failed: %v", err)
	}

	// Verify the hash is a SHA-256 (32 bytes)
	hash, exists := cps.GetPin("node1")
	if !exists {
		t.Fatal("Pin should exist")
	}
	if len(hash) != 32 {
		t.Errorf("Expected 32-byte SHA-256 hash, got %d bytes", len(hash))
	}
}

func TestCertPinStore_TOFUAutoPin(t *testing.T) {
	cps := NewCertPinStore()

	cert, err := createTestCertificate()
	if err != nil {
		t.Fatalf("Failed to create test certificate: %v", err)
	}

	// First verification for unknown node should auto-pin (TOFU)
	if err := cps.VerifyCertificate("new-node", cert); err != nil {
		t.Errorf("TOFU verification should succeed: %v", err)
	}

	// Verify it was actually pinned
	if cps.PinCount() != 1 {
		t.Errorf("Expected 1 pin after TOFU, got %d", cps.PinCount())
	}

	// Subsequent verification with same cert should succeed
	if err := cps.VerifyCertificate("new-node", cert); err != nil {
		t.Errorf("Second verification should succeed: %v", err)
	}

	// Different cert for same node should fail
	cert2, err := createTestCertificate()
	if err != nil {
		t.Fatalf("Failed to create test certificate: %v", err)
	}
	if err := cps.VerifyCertificate("new-node", cert2); err == nil {
		t.Error("Verification with different cert should fail after TOFU pin")
	}
}

func TestCertPinStore_RemovePinAndRepin(t *testing.T) {
	cps := NewCertPinStore()

	cert1, err := createTestCertificate()
	if err != nil {
		t.Fatalf("Failed to create test certificate: %v", err)
	}
	cert2, err := createTestCertificate()
	if err != nil {
		t.Fatalf("Failed to create test certificate: %v", err)
	}

	// Pin cert1
	cps.PinCertificate("node1", cert1)

	// Remove pin
	cps.RemovePin("node1")

	// After removal, TOFU should allow pinning cert2
	if err := cps.VerifyCertificate("node1", cert2); err != nil {
		t.Errorf("After removal, TOFU should succeed with new cert: %v", err)
	}

	// Now cert1 should fail
	if err := cps.VerifyCertificate("node1", cert1); err == nil {
		t.Error("Original cert should fail after re-pin with different cert")
	}
}

func TestCertPinStore_SaveCreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	pinPath := filepath.Join(tmpDir, "subdir", "nested", "certpins.json")

	cps := NewCertPinStore()
	cert, err := createTestCertificate()
	if err != nil {
		t.Fatalf("Failed to create test certificate: %v", err)
	}
	cps.PinCertificate("node1", cert)

	// Save should create intermediate directories
	if err := cps.Save(pinPath); err != nil {
		t.Fatalf("Save should create directories: %v", err)
	}

	// Load should work from the created path
	cps2 := NewCertPinStore()
	if err := cps2.Load(pinPath); err != nil {
		t.Fatalf("Load should work: %v", err)
	}
	if cps2.PinCount() != 1 {
		t.Errorf("Expected 1 pin, got %d", cps2.PinCount())
	}
}
