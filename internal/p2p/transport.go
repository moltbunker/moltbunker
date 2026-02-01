package p2p

import (
	"context"
	"crypto/sha256"
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
	certManager   *identity.CertificateManager
	mitmPreventor *security.MITMPreventor
	pinStore      *security.CertPinStore
	dialTimeout   time.Duration
}

// NewTransport creates a new transport instance
func NewTransport(certManager *identity.CertificateManager, pinStore *security.CertPinStore) (*Transport, error) {
	mitmPreventor := security.NewMITMPreventor(pinStore)

	return &Transport{
		certManager:   certManager,
		mitmPreventor: mitmPreventor,
		pinStore:      pinStore,
		dialTimeout:   30 * time.Second,
	}, nil
}

// SetDialTimeout sets the dial timeout
func (t *Transport) SetDialTimeout(timeout time.Duration) {
	t.dialTimeout = timeout
}

// Listen creates a TLS listener
func (t *Transport) Listen(address string) (net.Listener, error) {
	tlsConfig := t.serverTLSConfig()

	listener, err := tls.Listen("tcp", address, tlsConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create TLS listener: %w", err)
	}

	return listener, nil
}

// serverTLSConfig returns TLS config for server mode
func (t *Transport) serverTLSConfig() *tls.Config {
	baseCfg := t.certManager.TLSConfig()

	return &tls.Config{
		Certificates: baseCfg.Certificates,
		MinVersion:   tls.VersionTLS13,
		MaxVersion:   tls.VersionTLS13,
		CipherSuites: []uint16{
			tls.TLS_AES_256_GCM_SHA384,
			tls.TLS_CHACHA20_POLY1305_SHA256,
			tls.TLS_AES_128_GCM_SHA256,
		},
		// Require client certificates for mutual TLS
		ClientAuth: tls.RequireAnyClientCert,
		// Custom verification for self-signed certificates
		VerifyPeerCertificate: func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
			// Reject connections with no client certificate
			if len(rawCerts) == 0 {
				return fmt.Errorf("client certificate required")
			}

			// Parse and validate the client certificate
			cert, err := x509.ParseCertificate(rawCerts[0])
			if err != nil {
				return fmt.Errorf("invalid client certificate: %w", err)
			}

			// Validate certificate is not expired
			now := time.Now()
			if now.Before(cert.NotBefore) || now.After(cert.NotAfter) {
				return fmt.Errorf("client certificate expired or not yet valid")
			}

			// Validate certificate has proper key usage
			if cert.KeyUsage&x509.KeyUsageDigitalSignature == 0 {
				return fmt.Errorf("client certificate missing digital signature key usage")
			}

			// Compute and store certificate hash for pinning
			certHash := sha256.Sum256(cert.RawSubjectPublicKeyInfo)

			// Check if we have a pin for this certificate
			// If pinned, verify it matches; if not, this is TOFU (trust on first use)
			if existingPin, exists := t.pinStore.GetPin(fmt.Sprintf("%x", certHash)); exists {
				if len(existingPin) != len(certHash) {
					return fmt.Errorf("certificate pin length mismatch")
				}
				for i := range certHash {
					if certHash[i] != existingPin[i] {
						return fmt.Errorf("client certificate does not match pinned certificate")
					}
				}
			}

			return nil
		},
	}
}

// Dial creates a TLS connection with certificate pinning
func (t *Transport) Dial(nodeID types.NodeID, address string) (*tls.Conn, error) {
	return t.DialContext(context.Background(), nodeID, address)
}

// DialContext creates a TLS connection with context and certificate pinning
func (t *Transport) DialContext(ctx context.Context, nodeID types.NodeID, address string) (*tls.Conn, error) {
	// Get pinned certificate hash
	pinnedHash, hasPinnedCert := t.pinStore.GetPin(nodeID.String())

	// Create base TLS config
	tlsConfig := &tls.Config{
		Certificates: t.certManager.TLSConfig().Certificates,
		MinVersion:   tls.VersionTLS13,
		MaxVersion:   tls.VersionTLS13,
		CipherSuites: []uint16{
			tls.TLS_AES_256_GCM_SHA384,
			tls.TLS_CHACHA20_POLY1305_SHA256,
			tls.TLS_AES_128_GCM_SHA256,
		},
		// Skip standard CA verification for self-signed certs
		// We use certificate pinning instead
		InsecureSkipVerify: true,
		// ALWAYS verify peer certificate - either against pin or validate format
		VerifyPeerCertificate: func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
			if len(rawCerts) == 0 {
				return fmt.Errorf("server provided no certificates")
			}

			cert, err := x509.ParseCertificate(rawCerts[0])
			if err != nil {
				return fmt.Errorf("failed to parse server certificate: %w", err)
			}

			// Always validate certificate is not expired
			now := time.Now()
			if now.Before(cert.NotBefore) || now.After(cert.NotAfter) {
				return fmt.Errorf("server certificate expired or not yet valid")
			}

			// Always validate certificate has proper key usage
			if cert.KeyUsage&x509.KeyUsageDigitalSignature == 0 {
				return fmt.Errorf("server certificate missing digital signature key usage")
			}

			// Compute SHA256 hash of certificate public key (standard HPKP approach)
			certHash := sha256.Sum256(cert.RawSubjectPublicKeyInfo)

			// If we have a pinned certificate, verify it matches
			if hasPinnedCert && pinnedHash != nil {
				if len(pinnedHash) != len(certHash) {
					return fmt.Errorf("certificate pin length mismatch")
				}
				for i := range certHash {
					if certHash[i] != pinnedHash[i] {
						return fmt.Errorf("SECURITY: certificate does not match pinned certificate - possible MITM attack")
					}
				}
			}

			return nil
		},
	}

	// Create dialer with timeout
	dialer := &net.Dialer{
		Timeout: t.dialTimeout,
	}

	// Handle context deadline
	if deadline, ok := ctx.Deadline(); ok {
		if timeout := time.Until(deadline); timeout < dialer.Timeout {
			dialer.Timeout = timeout
		}
	}

	// Dial TCP connection first
	tcpConn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return nil, fmt.Errorf("failed to dial TCP connection: %w", err)
	}

	// Wrap with TLS
	tlsConn := tls.Client(tcpConn, tlsConfig)

	// Perform handshake with context
	handshakeCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := tlsConn.HandshakeContext(handshakeCtx); err != nil {
		tcpConn.Close()
		return nil, fmt.Errorf("TLS handshake failed: %w", err)
	}

	// Pin certificate if this is first connection (TOFU - Trust On First Use)
	if !hasPinnedCert || pinnedHash == nil {
		state := tlsConn.ConnectionState()
		if len(state.PeerCertificates) > 0 {
			t.pinStore.PinCertificate(nodeID.String(), state.PeerCertificates[0])
		}
	}

	return tlsConn, nil
}

// EncryptMessage encrypts a message using ChaCha20Poly1305
func (t *Transport) EncryptMessage(key []byte, message []byte) ([]byte, error) {
	return security.EncryptChaCha20Poly1305(key, message)
}

// DecryptMessage decrypts a message using ChaCha20Poly1305
func (t *Transport) DecryptMessage(key []byte, ciphertext []byte) ([]byte, error) {
	return security.DecryptChaCha20Poly1305(key, ciphertext)
}

// GetCertificate returns the transport's certificate
func (t *Transport) GetCertificate() *x509.Certificate {
	return t.certManager.Certificate()
}
