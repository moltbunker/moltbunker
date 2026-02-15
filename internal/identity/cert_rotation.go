package identity

import (
	"context"
	"crypto/ed25519"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/moltbunker/moltbunker/internal/logging"
	"github.com/moltbunker/moltbunker/internal/util"
)

const (
	// defaultRenewBefore is the default duration before expiry to trigger renewal
	defaultRenewBefore = 30 * 24 * time.Hour // 30 days

	// defaultCheckInterval is the default interval between expiry checks
	defaultCheckInterval = 24 * time.Hour
)

// CertRotator manages automatic TLS certificate rotation.
// It periodically checks if the current certificate is approaching expiry
// and generates a new one when the renewBefore window is reached.
type CertRotator struct {
	keyManager    *KeyManager
	certPath      string
	keyPath       string
	checkInterval time.Duration
	renewBefore   time.Duration
	mu            sync.RWMutex
	currentCert   *tls.Certificate
	onRotation    func(cert *tls.Certificate)
	cancel        context.CancelFunc
	stopped       chan struct{}
	running       bool

	// nowFunc allows injecting a custom clock for testing
	nowFunc func() time.Time
}

// NewCertRotator creates a new CertRotator that manages certificate lifecycle.
// It attempts to load an existing certificate from disk. If none is found or
// loading fails, it generates a fresh certificate immediately.
func NewCertRotator(keyManager *KeyManager, certPath, keyPath string) *CertRotator {
	cr := &CertRotator{
		keyManager:    keyManager,
		certPath:      certPath,
		keyPath:       keyPath,
		checkInterval: defaultCheckInterval,
		renewBefore:   defaultRenewBefore,
		nowFunc:       time.Now,
	}

	// Attempt to load existing certificate from disk
	if err := cr.loadFromDisk(); err != nil {
		logging.Info("no existing certificate found, will generate on first use",
			logging.Err(err),
			logging.Component("cert-rotation"))
	}

	return cr
}

// SetRenewBefore sets how long before expiry the certificate should be renewed.
// Default is 30 days.
func (cr *CertRotator) SetRenewBefore(d time.Duration) {
	cr.mu.Lock()
	defer cr.mu.Unlock()
	cr.renewBefore = d
}

// SetCheckInterval sets how frequently the rotator checks for expiry.
// Default is 24 hours.
func (cr *CertRotator) SetCheckInterval(d time.Duration) {
	cr.mu.Lock()
	defer cr.mu.Unlock()
	cr.checkInterval = d
}

// SetOnRotation sets a callback that is invoked when a certificate rotation
// occurs. This can be used to notify peers of the new certificate.
func (cr *CertRotator) SetOnRotation(fn func(cert *tls.Certificate)) {
	cr.mu.Lock()
	defer cr.mu.Unlock()
	cr.onRotation = fn
}

// GetCertificate returns the current TLS certificate in a thread-safe manner.
// This is suitable for use as a tls.Config.GetCertificate callback.
func (cr *CertRotator) GetCertificate() *tls.Certificate {
	cr.mu.RLock()
	defer cr.mu.RUnlock()
	return cr.currentCert
}

// NeedsRenewal checks whether the current certificate expires within the
// renewBefore window. Returns true if no certificate is loaded.
func (cr *CertRotator) NeedsRenewal() bool {
	cr.mu.RLock()
	defer cr.mu.RUnlock()

	if cr.currentCert == nil {
		return true
	}

	// Parse the leaf certificate to check expiry
	leaf, err := cr.leafCertificate()
	if err != nil {
		return true
	}

	now := cr.nowFunc()
	return now.Add(cr.renewBefore).After(leaf.NotAfter)
}

// Rotate generates a new certificate, saves it to disk, and updates the
// current certificate atomically. If an onRotation callback is set, it is
// invoked with the new certificate after successful rotation.
func (cr *CertRotator) Rotate(ctx context.Context) (*tls.Certificate, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Generate a new certificate using the existing key manager
	certManager, err := NewCertificateManager(cr.keyManager)
	if err != nil {
		return nil, fmt.Errorf("failed to generate certificate: %w", err)
	}

	// Encode to PEM for disk persistence
	certPEM, keyPEM, err := certManager.PEMEncode()
	if err != nil {
		return nil, fmt.Errorf("failed to encode certificate PEM: %w", err)
	}

	// Save to disk
	if err := cr.saveToDisk(certPEM, keyPEM); err != nil {
		return nil, fmt.Errorf("failed to save certificate to disk: %w", err)
	}

	// Build tls.Certificate directly from CertificateManager data.
	// We cannot use tls.X509KeyPair because PEMEncode writes raw Ed25519
	// key bytes rather than PKCS8-encoded DER, which X509KeyPair expects.
	leaf, err := x509.ParseCertificate(certManager.certDER)
	if err != nil {
		return nil, fmt.Errorf("failed to parse generated certificate: %w", err)
	}

	tlsCert := &tls.Certificate{
		Certificate: [][]byte{certManager.certDER},
		PrivateKey:  certManager.key,
		Leaf:        leaf,
	}

	// Update current cert atomically
	cr.mu.Lock()
	cr.currentCert = tlsCert
	onRotation := cr.onRotation
	cr.mu.Unlock()

	logging.Info("certificate rotated successfully",
		"cert_path", cr.certPath,
		logging.Component("cert-rotation"))

	// Invoke rotation callback outside the lock
	if onRotation != nil {
		onRotation(tlsCert)
	}

	return tlsCert, nil
}

// Start begins the background goroutine that periodically checks certificate
// expiry and rotates when necessary. If no certificate is currently loaded,
// it performs an immediate rotation.
func (cr *CertRotator) Start(ctx context.Context) {
	cr.mu.Lock()
	if cr.running {
		cr.mu.Unlock()
		return
	}
	cr.running = true
	cr.stopped = make(chan struct{})

	innerCtx, cancel := context.WithCancel(ctx)
	cr.cancel = cancel
	checkInterval := cr.checkInterval
	cr.mu.Unlock()

	// If no certificate is loaded, rotate immediately
	if cr.GetCertificate() == nil {
		if _, err := cr.Rotate(innerCtx); err != nil {
			logging.Error("failed initial certificate rotation",
				logging.Err(err),
				logging.Component("cert-rotation"))
		}
	}

	util.SafeGoWithName("cert-rotation-checker", func() {
		defer func() {
			cr.mu.Lock()
			cr.running = false
			cr.mu.Unlock()
			close(cr.stopped)
		}()

		ticker := time.NewTicker(checkInterval)
		defer ticker.Stop()

		for {
			select {
			case <-innerCtx.Done():
				return
			case <-ticker.C:
				if cr.NeedsRenewal() {
					logging.Info("certificate approaching expiry, rotating",
						logging.Component("cert-rotation"))
					if _, err := cr.Rotate(innerCtx); err != nil {
						logging.Error("failed to rotate certificate",
							logging.Err(err),
							logging.Component("cert-rotation"))
					}
				}
			}
		}
	})
}

// Stop stops the background certificate rotation checker and waits for the
// goroutine to exit.
func (cr *CertRotator) Stop() {
	cr.mu.RLock()
	cancel := cr.cancel
	running := cr.running
	stopped := cr.stopped
	cr.mu.RUnlock()

	if !running || cancel == nil {
		return
	}

	cancel()

	// Wait for the goroutine to finish
	if stopped != nil {
		<-stopped
	}
}

// loadFromDisk loads an existing certificate and key from their PEM files.
// The key PEM may contain raw Ed25519 bytes (as produced by CertificateManager.PEMEncode)
// rather than PKCS8-encoded DER, so we handle both formats.
func (cr *CertRotator) loadFromDisk() error {
	certPEM, err := os.ReadFile(cr.certPath)
	if err != nil {
		return fmt.Errorf("failed to read cert file: %w", err)
	}

	keyPEM, err := os.ReadFile(cr.keyPath)
	if err != nil {
		return fmt.Errorf("failed to read key file: %w", err)
	}

	// Parse certificate
	certBlock, _ := pem.Decode(certPEM)
	if certBlock == nil {
		return fmt.Errorf("failed to decode certificate PEM")
	}

	leaf, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse certificate: %w", err)
	}

	// Parse private key - handle both raw Ed25519 and PKCS8 formats
	keyBlock, _ := pem.Decode(keyPEM)
	if keyBlock == nil {
		return fmt.Errorf("failed to decode key PEM")
	}

	var privateKey ed25519.PrivateKey
	parsed, pkcs8Err := x509.ParsePKCS8PrivateKey(keyBlock.Bytes)
	if pkcs8Err != nil {
		// Fall back to raw Ed25519 key (64 bytes)
		if len(keyBlock.Bytes) == ed25519.PrivateKeySize {
			privateKey = ed25519.PrivateKey(keyBlock.Bytes)
		} else {
			return fmt.Errorf("failed to parse private key: %w", pkcs8Err)
		}
	} else {
		var ok bool
		privateKey, ok = parsed.(ed25519.PrivateKey)
		if !ok {
			return fmt.Errorf("private key is not Ed25519")
		}
	}

	tlsCert := &tls.Certificate{
		Certificate: [][]byte{certBlock.Bytes},
		PrivateKey:  privateKey,
		Leaf:        leaf,
	}

	cr.mu.Lock()
	cr.currentCert = tlsCert
	cr.mu.Unlock()

	return nil
}

// saveToDisk writes PEM-encoded certificate and key files with appropriate permissions.
func (cr *CertRotator) saveToDisk(certPEM, keyPEM []byte) error {
	// Create directories if they don't exist
	if err := os.MkdirAll(filepath.Dir(cr.certPath), 0700); err != nil {
		return fmt.Errorf("failed to create cert directory: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(cr.keyPath), 0700); err != nil {
		return fmt.Errorf("failed to create key directory: %w", err)
	}

	if err := os.WriteFile(cr.certPath, certPEM, 0600); err != nil {
		return fmt.Errorf("failed to write cert file: %w", err)
	}

	if err := os.WriteFile(cr.keyPath, keyPEM, 0600); err != nil {
		return fmt.Errorf("failed to write key file: %w", err)
	}

	return nil
}

// leafCertificate parses and returns the leaf certificate from the current
// TLS certificate. Must be called with at least a read lock held.
func (cr *CertRotator) leafCertificate() (*x509.Certificate, error) {
	if cr.currentCert == nil || len(cr.currentCert.Certificate) == 0 {
		return nil, fmt.Errorf("no certificate loaded")
	}

	// Use the cached leaf if available
	if cr.currentCert.Leaf != nil {
		return cr.currentCert.Leaf, nil
	}

	// Parse the DER-encoded leaf certificate
	leaf, err := x509.ParseCertificate(cr.currentCert.Certificate[0])
	if err != nil {
		return nil, fmt.Errorf("failed to parse leaf certificate: %w", err)
	}

	return leaf, nil
}
