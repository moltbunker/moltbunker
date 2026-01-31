package p2p

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"time"

	"github.com/moltbunker/moltbunker/internal/identity"
	"github.com/moltbunker/moltbunker/internal/security"
	"github.com/moltbunker/moltbunker/pkg/types"
)

// Transport provides encrypted transport layer (TLS 1.3)
type Transport struct {
	certManager *identity.CertificateManager
	mitmPreventor *security.MITMPreventor
	pinStore    *security.CertPinStore
}

// NewTransport creates a new transport instance
func NewTransport(certManager *identity.CertificateManager, pinStore *security.CertPinStore) (*Transport, error) {
	mitmPreventor := security.NewMITMPreventor(pinStore)

	return &Transport{
		certManager:   certManager,
		mitmPreventor: mitmPreventor,
		pinStore:      pinStore,
	}, nil
}

// Listen creates a TLS listener
func (t *Transport) Listen(address string) (net.Listener, error) {
	tlsConfig := t.certManager.TLSConfig()
	tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert

	listener, err := tls.Listen("tcp", address, tlsConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create TLS listener: %w", err)
	}

	return listener, nil
}

// Dial creates a TLS connection with certificate pinning
func (t *Transport) Dial(nodeID types.NodeID, address string) (*tls.Conn, error) {
	// Get pinned certificate hash
	pinnedHash, exists := t.pinStore.GetPin(nodeID.String())
	if !exists {
		// First connection - will pin on first successful connection
		pinnedHash = nil
	}

	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS13,
		MaxVersion: tls.VersionTLS13,
		InsecureSkipVerify: pinnedHash == nil, // Skip verify on first connection
	}

	if pinnedHash != nil {
		// Set verify callback for certificate pinning
		tlsConfig.VerifyPeerCertificate = t.mitmPreventor.TLSVerifyCallback(nodeID.String())
	}

	conn, err := tls.DialWithDialer(
		&net.Dialer{Timeout: 10 * time.Second},
		"tcp",
		address,
		tlsConfig,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to dial TLS connection: %w", err)
	}

	// Pin certificate if this is first connection
	if pinnedHash == nil {
		state := conn.ConnectionState()
		if len(state.PeerCertificates) > 0 {
			t.pinStore.PinCertificate(nodeID.String(), state.PeerCertificates[0])
		}
	}

	return conn, nil
}

// EncryptMessage encrypts a message using ChaCha20Poly1305
func (t *Transport) EncryptMessage(key []byte, message []byte) ([]byte, error) {
	return security.EncryptChaCha20Poly1305(key, message)
}

// DecryptMessage decrypts a message using ChaCha20Poly1305
func (t *Transport) DecryptMessage(key []byte, ciphertext []byte) ([]byte, error) {
	return security.DecryptChaCha20Poly1305(key, ciphertext)
}
