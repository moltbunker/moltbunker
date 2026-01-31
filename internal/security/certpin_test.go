package security

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
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
