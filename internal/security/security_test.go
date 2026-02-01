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

// --- Certificate Creation Helpers ---

// createTestCertificateWithOptions creates a test certificate with custom options
func createTestCertificateWithOptions(notBefore, notAfter time.Time) (*x509.Certificate, ed25519.PrivateKey, error) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, nil, err
	}

	template := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"MoltBunker Test"},
			CommonName:   "test-node",
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  false,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, publicKey, privateKey)
	if err != nil {
		return nil, nil, err
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return nil, nil, err
	}

	return cert, privateKey, nil
}

// createValidCertificate creates a certificate valid for 1 year
func createValidCertificate() (*x509.Certificate, ed25519.PrivateKey, error) {
	now := time.Now()
	return createTestCertificateWithOptions(now.Add(-time.Hour), now.Add(365*24*time.Hour))
}

// createExpiredCertificate creates an already-expired certificate
func createExpiredCertificate() (*x509.Certificate, ed25519.PrivateKey, error) {
	now := time.Now()
	return createTestCertificateWithOptions(now.Add(-365*24*time.Hour), now.Add(-time.Hour))
}

// createFutureCertificate creates a certificate that's not yet valid
func createFutureCertificate() (*x509.Certificate, ed25519.PrivateKey, error) {
	now := time.Now()
	return createTestCertificateWithOptions(now.Add(time.Hour), now.Add(365*24*time.Hour))
}

// --- Certificate Pinning Tests ---

func TestCertificatePinning_Works(t *testing.T) {
	store := NewCertPinStore()

	// Create first certificate
	cert1, _, err := createValidCertificate()
	if err != nil {
		t.Fatalf("Failed to create certificate: %v", err)
	}

	nodeID := "test-node-1"

	// Pin the certificate
	store.PinCertificate(nodeID, cert1)

	// Verify the pin was stored
	pin, exists := store.GetPin(nodeID)
	if !exists {
		t.Fatal("Pin should exist after pinning")
	}

	if len(pin) != 32 {
		t.Errorf("Pin should be SHA-256 hash (32 bytes), got %d bytes", len(pin))
	}

	// Verify the same certificate passes verification
	err = store.VerifyCertificate(nodeID, cert1)
	if err != nil {
		t.Errorf("Certificate verification should succeed for pinned cert: %v", err)
	}
}

func TestCertificatePinning_MultiplePins(t *testing.T) {
	store := NewCertPinStore()

	// Create certificates for multiple nodes
	numNodes := 10
	certs := make([]*x509.Certificate, numNodes)
	nodeIDs := make([]string, numNodes)

	for i := 0; i < numNodes; i++ {
		cert, _, err := createValidCertificate()
		if err != nil {
			t.Fatalf("Failed to create certificate %d: %v", i, err)
		}
		certs[i] = cert
		nodeIDs[i] = "node-" + string(rune('A'+i))
		store.PinCertificate(nodeIDs[i], cert)
	}

	// Verify all pins exist and validate correctly
	for i := 0; i < numNodes; i++ {
		_, exists := store.GetPin(nodeIDs[i])
		if !exists {
			t.Errorf("Pin should exist for node %s", nodeIDs[i])
		}

		err := store.VerifyCertificate(nodeIDs[i], certs[i])
		if err != nil {
			t.Errorf("Verification failed for node %s: %v", nodeIDs[i], err)
		}
	}
}

// --- Certificate Mismatch Tests ---

func TestCertificateMismatch_IsRejected(t *testing.T) {
	store := NewCertPinStore()

	// Create two different certificates
	cert1, _, err := createValidCertificate()
	if err != nil {
		t.Fatalf("Failed to create first certificate: %v", err)
	}

	cert2, _, err := createValidCertificate()
	if err != nil {
		t.Fatalf("Failed to create second certificate: %v", err)
	}

	nodeID := "test-node"

	// Pin the first certificate
	store.PinCertificate(nodeID, cert1)

	// Try to verify with a different certificate (should fail)
	err = store.VerifyCertificate(nodeID, cert2)
	if err == nil {
		t.Error("Verification should fail with mismatched certificate")
	}

	// The error should indicate a pinning failure
	if err != nil {
		t.Logf("Got expected error: %v", err)
	}
}

func TestCertificateMismatch_AfterRenewal(t *testing.T) {
	store := NewCertPinStore()

	// Simulate certificate renewal scenario
	nodeID := "renewal-test-node"

	// Original certificate
	originalCert, _, err := createValidCertificate()
	if err != nil {
		t.Fatalf("Failed to create original certificate: %v", err)
	}

	// Pin original
	store.PinCertificate(nodeID, originalCert)

	// Create "renewed" certificate (different key pair)
	renewedCert, _, err := createValidCertificate()
	if err != nil {
		t.Fatalf("Failed to create renewed certificate: %v", err)
	}

	// Verification should fail (different public key)
	err = store.VerifyCertificate(nodeID, renewedCert)
	if err == nil {
		t.Error("Verification should fail with renewed certificate (different key)")
	}

	// To accept renewed cert, admin must remove old pin first
	store.RemovePin(nodeID)

	// Now verification should succeed (TOFU - Trust On First Use)
	err = store.VerifyCertificate(nodeID, renewedCert)
	if err != nil {
		t.Errorf("After removing pin, TOFU should accept new certificate: %v", err)
	}
}

func TestCertificateMismatch_DifferentKeysSameCertFields(t *testing.T) {
	store := NewCertPinStore()

	// Create two certificates with same metadata but different keys
	now := time.Now()

	pub1, priv1, _ := ed25519.GenerateKey(rand.Reader)
	pub2, priv2, _ := ed25519.GenerateKey(rand.Reader)

	// Use same serial number and subject
	template := &x509.Certificate{
		SerialNumber: big.NewInt(12345),
		Subject: pkix.Name{
			Organization: []string{"Same Org"},
			CommonName:   "same-node",
		},
		NotBefore:             now.Add(-time.Hour),
		NotAfter:              now.Add(365*24*time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}

	cert1DER, _ := x509.CreateCertificate(rand.Reader, template, template, pub1, priv1)
	cert2DER, _ := x509.CreateCertificate(rand.Reader, template, template, pub2, priv2)

	cert1, _ := x509.ParseCertificate(cert1DER)
	cert2, _ := x509.ParseCertificate(cert2DER)

	nodeID := "same-fields-node"
	store.PinCertificate(nodeID, cert1)

	// Even with same cert fields, different public key should fail
	err := store.VerifyCertificate(nodeID, cert2)
	if err == nil {
		t.Error("Certificates with same fields but different keys should fail verification")
	}
}

// --- Expired Certificate Tests ---

func TestExpiredCertificate_IsRejected(t *testing.T) {
	// Note: The CertPinStore doesn't check expiration directly,
	// but we should test that expired certs can still be detected
	// and handled at the application level

	expiredCert, _, err := createExpiredCertificate()
	if err != nil {
		t.Fatalf("Failed to create expired certificate: %v", err)
	}

	// Check certificate validity
	now := time.Now()
	if now.Before(expiredCert.NotBefore) {
		t.Error("Certificate should be past NotBefore")
	}
	if !now.After(expiredCert.NotAfter) {
		t.Error("Certificate should be past NotAfter (expired)")
	}

	// Verify the certificate is expired
	if err := verifyCertificateExpiration(expiredCert); err == nil {
		t.Error("Should detect expired certificate")
	}
}

func TestExpiredCertificate_PinStoreStillWorks(t *testing.T) {
	// CertPinStore doesn't check expiration - that's the TLS layer's job
	// But we can pin and verify based on public key hash
	store := NewCertPinStore()

	expiredCert, _, err := createExpiredCertificate()
	if err != nil {
		t.Fatalf("Failed to create expired certificate: %v", err)
	}

	nodeID := "expired-node"
	store.PinCertificate(nodeID, expiredCert)

	// Pin verification still works (checks public key, not expiration)
	err = store.VerifyCertificate(nodeID, expiredCert)
	if err != nil {
		t.Logf("Pin verification result: %v", err)
	}

	// The store should still have the pin
	_, exists := store.GetPin(nodeID)
	if !exists {
		t.Error("Pin should exist even for expired cert")
	}
}

func TestNotYetValidCertificate(t *testing.T) {
	futureCert, _, err := createFutureCertificate()
	if err != nil {
		t.Fatalf("Failed to create future certificate: %v", err)
	}

	// Verify the certificate is not yet valid
	now := time.Now()
	if !now.Before(futureCert.NotBefore) {
		t.Error("Certificate should be before NotBefore (not yet valid)")
	}

	// Check validity
	if err := verifyCertificateExpiration(futureCert); err == nil {
		t.Error("Should detect not-yet-valid certificate")
	}
}

// verifyCertificateExpiration checks if a certificate is currently valid
func verifyCertificateExpiration(cert *x509.Certificate) error {
	now := time.Now()
	if now.Before(cert.NotBefore) {
		return x509.CertificateInvalidError{
			Cert:   cert,
			Reason: x509.NotAuthorizedToSign, // Using this as "not yet valid"
			Detail: "certificate is not yet valid",
		}
	}
	if now.After(cert.NotAfter) {
		return x509.CertificateInvalidError{
			Cert:   cert,
			Reason: x509.Expired,
			Detail: "certificate has expired",
		}
	}
	return nil
}

// --- Trust On First Use (TOFU) Tests ---

func TestTOFU_FirstConnection(t *testing.T) {
	store := NewCertPinStore()

	cert, _, err := createValidCertificate()
	if err != nil {
		t.Fatalf("Failed to create certificate: %v", err)
	}

	nodeID := "new-node"

	// First verification should succeed and auto-pin
	err = store.VerifyCertificate(nodeID, cert)
	if err != nil {
		t.Errorf("TOFU should accept first certificate: %v", err)
	}

	// Verify it was pinned
	pin, exists := store.GetPin(nodeID)
	if !exists {
		t.Error("Certificate should be auto-pinned after first verification")
	}

	if len(pin) == 0 {
		t.Error("Pinned hash should not be empty")
	}
}

func TestTOFU_SubsequentConnections(t *testing.T) {
	store := NewCertPinStore()

	cert, _, err := createValidCertificate()
	if err != nil {
		t.Fatalf("Failed to create certificate: %v", err)
	}

	nodeID := "persistent-node"

	// First connection (TOFU)
	err = store.VerifyCertificate(nodeID, cert)
	if err != nil {
		t.Fatalf("First verification failed: %v", err)
	}

	// Subsequent connections with same cert should succeed
	for i := 0; i < 10; i++ {
		err = store.VerifyCertificate(nodeID, cert)
		if err != nil {
			t.Errorf("Verification %d failed: %v", i+1, err)
		}
	}
}

// --- Concurrent Access Tests ---

func TestCertPinStore_ConcurrentAccess(t *testing.T) {
	store := NewCertPinStore()

	// Create certificates for concurrent testing
	numGoroutines := 50
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() { done <- true }()

			cert, _, err := createValidCertificate()
			if err != nil {
				t.Errorf("Failed to create certificate in goroutine %d: %v", id, err)
				return
			}

			nodeID := "concurrent-node-" + string(rune('0'+id%10))

			// Mix of operations
			switch id % 3 {
			case 0:
				store.PinCertificate(nodeID, cert)
			case 1:
				store.VerifyCertificate(nodeID, cert)
			case 2:
				store.GetPin(nodeID)
			}
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
}

func TestCertPinStore_ConcurrentPinAndVerify(t *testing.T) {
	store := NewCertPinStore()

	cert, _, err := createValidCertificate()
	if err != nil {
		t.Fatalf("Failed to create certificate: %v", err)
	}

	nodeID := "shared-node"
	store.PinCertificate(nodeID, cert)

	// Concurrent verifications
	numGoroutines := 100
	done := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			done <- store.VerifyCertificate(nodeID, cert)
		}()
	}

	// All verifications should succeed
	for i := 0; i < numGoroutines; i++ {
		if err := <-done; err != nil {
			t.Errorf("Concurrent verification failed: %v", err)
		}
	}
}

// --- Pin Removal Tests ---

func TestCertPinStore_RemovePinAndAcceptNew(t *testing.T) {
	store := NewCertPinStore()

	cert, _, err := createValidCertificate()
	if err != nil {
		t.Fatalf("Failed to create certificate: %v", err)
	}

	nodeID := "removable-node"

	// Pin and verify it exists
	store.PinCertificate(nodeID, cert)
	_, exists := store.GetPin(nodeID)
	if !exists {
		t.Fatal("Pin should exist after pinning")
	}

	// Remove the pin
	store.RemovePin(nodeID)

	// Verify it's gone
	_, exists = store.GetPin(nodeID)
	if exists {
		t.Error("Pin should not exist after removal")
	}

	// New certificate should be accepted (TOFU)
	newCert, _, err := createValidCertificate()
	if err != nil {
		t.Fatalf("Failed to create new certificate: %v", err)
	}

	err = store.VerifyCertificate(nodeID, newCert)
	if err != nil {
		t.Errorf("After pin removal, should accept new certificate: %v", err)
	}
}

func TestCertPinStore_RemoveNonexistentPin(t *testing.T) {
	store := NewCertPinStore()

	// Should not panic when removing non-existent pin
	store.RemovePin("nonexistent-node")

	// Store should still work
	cert, _, err := createValidCertificate()
	if err != nil {
		t.Fatalf("Failed to create certificate: %v", err)
	}

	store.PinCertificate("new-node", cert)
	_, exists := store.GetPin("new-node")
	if !exists {
		t.Error("Pin should exist after pinning")
	}
}

// --- Edge Cases ---

func TestCertPinStore_EmptyNodeID(t *testing.T) {
	store := NewCertPinStore()

	cert, _, err := createValidCertificate()
	if err != nil {
		t.Fatalf("Failed to create certificate: %v", err)
	}

	// Empty node ID should still work (it's just a map key)
	store.PinCertificate("", cert)

	_, exists := store.GetPin("")
	if !exists {
		t.Error("Should be able to pin with empty node ID")
	}

	err = store.VerifyCertificate("", cert)
	if err != nil {
		t.Errorf("Verification with empty node ID should work: %v", err)
	}
}

func TestCertPinStore_UpdatePin(t *testing.T) {
	store := NewCertPinStore()

	cert1, _, err := createValidCertificate()
	if err != nil {
		t.Fatalf("Failed to create first certificate: %v", err)
	}

	cert2, _, err := createValidCertificate()
	if err != nil {
		t.Fatalf("Failed to create second certificate: %v", err)
	}

	nodeID := "update-node"

	// Pin first certificate
	store.PinCertificate(nodeID, cert1)
	pin1, _ := store.GetPin(nodeID)

	// Update to second certificate
	store.PinCertificate(nodeID, cert2)
	pin2, _ := store.GetPin(nodeID)

	// Pins should be different
	if string(pin1) == string(pin2) {
		t.Error("Updated pin should be different from original")
	}

	// Verification with old cert should fail
	err = store.VerifyCertificate(nodeID, cert1)
	if err == nil {
		t.Error("Verification with old certificate should fail after update")
	}

	// Verification with new cert should succeed
	err = store.VerifyCertificate(nodeID, cert2)
	if err != nil {
		t.Errorf("Verification with new certificate should succeed: %v", err)
	}
}
