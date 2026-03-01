package daemon

import (
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"net"
)

// TLSTunnelDialer implements tunnel.Dialer using TLS connections
// with SPKI-based NodeID verification to prevent MITM attacks.
type TLSTunnelDialer struct {
	tlsConfig *tls.Config
}

// NewTLSTunnelDialer creates a new TLS tunnel dialer.
func NewTLSTunnelDialer(tlsConfig *tls.Config) *TLSTunnelDialer {
	return &TLSTunnelDialer{tlsConfig: tlsConfig}
}

// DialProvider implements tunnel.Dialer with SPKI verification.
// expectedNodeID is the hex-encoded SHA256 of the provider's SPKI.
// If non-empty, the peer certificate's SPKI hash must match.
func (d *TLSTunnelDialer) DialProvider(addr string, expectedNodeID string) (net.Conn, error) {
	// Clone the config to set a per-connection VerifyConnection callback
	cfg := d.tlsConfig.Clone()

	if expectedNodeID != "" {
		cfg.InsecureSkipVerify = true // We do manual SPKI verification below
		cfg.VerifyConnection = func(cs tls.ConnectionState) error {
			if len(cs.PeerCertificates) == 0 {
				return fmt.Errorf("provider presented no TLS certificate")
			}
			spki := cs.PeerCertificates[0].RawSubjectPublicKeyInfo
			fingerprint := sha256.Sum256(spki)
			peerNodeID := hex.EncodeToString(fingerprint[:])
			if peerNodeID != expectedNodeID {
				return fmt.Errorf("provider NodeID mismatch: got %s, expected %s", peerNodeID[:16], expectedNodeID[:16])
			}
			return nil
		}
	}

	conn, err := tls.Dial("tcp", addr, cfg)
	if err != nil {
		return nil, fmt.Errorf("TLS dial to %s: %w", addr, err)
	}
	return conn, nil
}
