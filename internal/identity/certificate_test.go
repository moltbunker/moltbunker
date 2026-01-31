package identity

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"testing"
)

func TestCertificateManager_NewCertificateManager(t *testing.T) {
	// Generate test key
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	km := &KeyManager{
		privateKey: privateKey,
		publicKey:  privateKey.Public().(ed25519.PublicKey),
	}

	cm, err := NewCertificateManager(km)
	if err != nil {
		t.Fatalf("Failed to create certificate manager: %v", err)
	}

	if cm.Certificate() == nil {
		t.Error("Certificate is nil")
	}

	if cm.Certificate().PublicKey == nil {
		t.Error("Certificate public key is nil")
	}
}

func TestCertificateManager_TLSConfig(t *testing.T) {
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	km := &KeyManager{
		privateKey: privateKey,
		publicKey:  privateKey.Public().(ed25519.PublicKey),
	}

	cm, err := NewCertificateManager(km)
	if err != nil {
		t.Fatalf("Failed to create certificate manager: %v", err)
	}

	tlsConfig := cm.TLSConfig()
	if tlsConfig == nil {
		t.Error("TLS config is nil")
	}

	if len(tlsConfig.Certificates) == 0 {
		t.Error("No certificates in TLS config")
	}

	if tlsConfig.MinVersion != 0x0304 { // TLS 1.3
		t.Error("TLS version should be 1.3")
	}
}

func TestCertificateManager_PEMEncodeDecode(t *testing.T) {
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	km := &KeyManager{
		privateKey: privateKey,
		publicKey:  privateKey.Public().(ed25519.PublicKey),
	}

	cm1, err := NewCertificateManager(km)
	if err != nil {
		t.Fatalf("Failed to create certificate manager: %v", err)
	}

	certPEM, keyPEM, err := cm1.PEMEncode()
	if err != nil {
		t.Fatalf("Failed to encode PEM: %v", err)
	}

	if len(certPEM) == 0 {
		t.Error("Certificate PEM is empty")
	}

	if len(keyPEM) == 0 {
		t.Error("Key PEM is empty")
	}

	// Decode
	cm2, err := PEMDecode(certPEM, keyPEM)
	if err != nil {
		t.Fatalf("Failed to decode PEM: %v", err)
	}

	if cm2.Certificate() == nil {
		t.Error("Decoded certificate is nil")
	}
}

func TestCertificateManager_PublicKeyHash(t *testing.T) {
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	km := &KeyManager{
		privateKey: privateKey,
		publicKey:  privateKey.Public().(ed25519.PublicKey),
	}

	cm, err := NewCertificateManager(km)
	if err != nil {
		t.Fatalf("Failed to create certificate manager: %v", err)
	}

	hash := cm.PublicKeyHash()
	if len(hash) == 0 {
		t.Error("Public key hash is empty")
	}
}

func TestCertificateManager_VerifyPeerCertificate(t *testing.T) {
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	km := &KeyManager{
		privateKey: privateKey,
		publicKey:  privateKey.Public().(ed25519.PublicKey),
	}

	cm, err := NewCertificateManager(km)
	if err != nil {
		t.Fatalf("Failed to create certificate manager: %v", err)
	}

	pinnedKey := cm.PublicKeyHash()
	verifyFunc := cm.VerifyPeerCertificate(pinnedKey)

	// Create test certificate
	cert := cm.Certificate()
	rawCerts := [][]byte{cert.Raw}
	verifiedChains := [][]*x509.Certificate{{cert}}

	err = verifyFunc(rawCerts, verifiedChains)
	if err != nil {
		t.Errorf("Certificate verification failed: %v", err)
	}
}
