package security

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
)

// MITMPreventor prevents MITM attacks via certificate pinning
type MITMPreventor struct {
	pinStore *CertPinStore
}

// NewMITMPreventor creates a new MITM preventor
func NewMITMPreventor(pinStore *CertPinStore) *MITMPreventor {
	return &MITMPreventor{
		pinStore: pinStore,
	}
}

// VerifyTLSConnection verifies a TLS connection against pinned certificates
func (mp *MITMPreventor) VerifyTLSConnection(nodeID string, conn *tls.Conn) error {
	state := conn.ConnectionState()
	if len(state.PeerCertificates) == 0 {
		return fmt.Errorf("no peer certificates")
	}

	cert := state.PeerCertificates[0]
	return mp.pinStore.VerifyCertificate(nodeID, cert)
}

// TLSVerifyCallback returns a TLS verify callback for certificate pinning
func (mp *MITMPreventor) TLSVerifyCallback(nodeID string) func([][]byte, [][]*x509.Certificate) error {
	return func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
		if len(rawCerts) == 0 {
			return fmt.Errorf("no certificates provided")
		}

		cert, err := x509.ParseCertificate(rawCerts[0])
		if err != nil {
			return fmt.Errorf("failed to parse certificate: %w", err)
		}

		return mp.pinStore.VerifyCertificate(nodeID, cert)
	}
}

// RotateCertificate rotates a pinned certificate (for key rotation)
func (mp *MITMPreventor) RotateCertificate(nodeID string, newCert *x509.Certificate) error {
	mp.pinStore.PinCertificate(nodeID, newCert)
	return nil
}
