package identity

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"
)

// CertificateManager manages TLS certificates with pinning
type CertificateManager struct {
	cert    *x509.Certificate
	key     ed25519.PrivateKey
	certDER []byte
	keyDER  []byte
}

// NewCertificateManager creates a new certificate manager
func NewCertificateManager(keyManager *KeyManager) (*CertificateManager, error) {
	cm := &CertificateManager{}

	// Use Ed25519 key from key manager
	privateKey := keyManager.PrivateKey()
	publicKey := keyManager.PublicKey()

	// Create certificate template
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization:  []string{"Moltbunker"},
			Country:       []string{"XX"},
			Province:      []string{},
			Locality:      []string{},
			StreetAddress: []string{},
			PostalCode:    []string{},
		},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(365 * 24 * time.Hour), // 1 year
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA: false,
	}

	// Create certificate
	certDER, err := x509.CreateCertificate(rand.Reader, template, template, publicKey, privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate: %w", err)
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	cm.cert = cert
	cm.key = privateKey
	cm.certDER = certDER
	cm.keyDER = privateKey

	return cm, nil
}

// Certificate returns the X.509 certificate
func (cm *CertificateManager) Certificate() *x509.Certificate {
	return cm.cert
}

// TLSConfig returns a TLS config with certificate pinning
func (cm *CertificateManager) TLSConfig() *tls.Config {
	return &tls.Config{
		Certificates: []tls.Certificate{
			{
				Certificate: [][]byte{cm.certDER},
				PrivateKey:  cm.key,
			},
		},
		MinVersion: tls.VersionTLS13,
		MaxVersion: tls.VersionTLS13,
		CipherSuites: []uint16{
			tls.TLS_AES_256_GCM_SHA384,
			tls.TLS_CHACHA20_POLY1305_SHA256,
			tls.TLS_AES_128_GCM_SHA256,
		},
	}
}

// PublicKeyHash returns SHA-256 hash of public key for pinning
func (cm *CertificateManager) PublicKeyHash() []byte {
	return cm.cert.RawSubjectPublicKeyInfo
}

// VerifyPeerCertificate verifies a peer's certificate against pinned public key
func (cm *CertificateManager) VerifyPeerCertificate(pinnedKey []byte) func([][]byte, [][]*x509.Certificate) error {
	return func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
		if len(rawCerts) == 0 {
			return fmt.Errorf("no certificates provided")
		}

		cert, err := x509.ParseCertificate(rawCerts[0])
		if err != nil {
			return fmt.Errorf("failed to parse certificate: %w", err)
		}

		// Compare public key hash
		certKeyHash := cert.RawSubjectPublicKeyInfo
		if len(certKeyHash) != len(pinnedKey) {
			return fmt.Errorf("certificate public key mismatch")
		}

		for i := range certKeyHash {
			if certKeyHash[i] != pinnedKey[i] {
				return fmt.Errorf("certificate pinning verification failed")
			}
		}

		return nil
	}
}

// PEMEncode encodes certificate and key to PEM format
func (cm *CertificateManager) PEMEncode() (certPEM []byte, keyPEM []byte, error error) {
	certPEM = pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cm.certDER,
	})

	keyPEM = pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: cm.keyDER,
	})

	return certPEM, keyPEM, nil
}

// PEMDecode decodes certificate and key from PEM format
func PEMDecode(certPEM, keyPEM []byte) (*CertificateManager, error) {
	cm := &CertificateManager{}

	// Decode certificate
	certBlock, _ := pem.Decode(certPEM)
	if certBlock == nil {
		return nil, fmt.Errorf("failed to decode certificate PEM")
	}

	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	// Decode private key
	keyBlock, _ := pem.Decode(keyPEM)
	if keyBlock == nil {
		return nil, fmt.Errorf("failed to decode key PEM")
	}

	// Try to parse as PKCS8 first, then try raw Ed25519
	var ed25519Key ed25519.PrivateKey
	privateKey, err := x509.ParsePKCS8PrivateKey(keyBlock.Bytes)
	if err != nil {
		// Try parsing as raw Ed25519 key (64 bytes)
		if len(keyBlock.Bytes) == ed25519.PrivateKeySize {
			ed25519Key = ed25519.PrivateKey(keyBlock.Bytes)
		} else {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
	} else {
		var ok bool
		ed25519Key, ok = privateKey.(ed25519.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("private key is not Ed25519")
		}
	}

	cm.cert = cert
	cm.key = ed25519Key
	cm.certDER = certBlock.Bytes
	cm.keyDER = ed25519Key

	return cm, nil
}
