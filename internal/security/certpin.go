package security

import (
	"crypto/sha256"
	"crypto/x509"
	"fmt"
	"sync"
)

// CertPinStore stores pinned certificate public keys
type CertPinStore struct {
	pins map[string][]byte // nodeID -> public key hash
	mu   sync.RWMutex
}

// NewCertPinStore creates a new certificate pinning store
func NewCertPinStore() *CertPinStore {
	return &CertPinStore{
		pins: make(map[string][]byte),
	}
}

// PinCertificate pins a certificate's public key for a node
func (cps *CertPinStore) PinCertificate(nodeID string, cert *x509.Certificate) {
	cps.mu.Lock()
	defer cps.mu.Unlock()

	// Hash the public key
	hash := sha256.Sum256(cert.RawSubjectPublicKeyInfo)
	cps.pins[nodeID] = hash[:]
}

// VerifyCertificate verifies a certificate against pinned public key
func (cps *CertPinStore) VerifyCertificate(nodeID string, cert *x509.Certificate) error {
	cps.mu.RLock()
	defer cps.mu.RUnlock()

	pinnedHash, exists := cps.pins[nodeID]
	if !exists {
		// First time seeing this node, pin it
		hash := sha256.Sum256(cert.RawSubjectPublicKeyInfo)
		cps.mu.RUnlock()
		cps.mu.Lock()
		cps.pins[nodeID] = hash[:]
		cps.mu.Unlock()
		cps.mu.RLock()
		return nil
	}

	// Verify against pinned hash
	hash := sha256.Sum256(cert.RawSubjectPublicKeyInfo)
	if len(hash) != len(pinnedHash) {
		return fmt.Errorf("certificate public key hash length mismatch")
	}

	for i := range hash {
		if hash[i] != pinnedHash[i] {
			return fmt.Errorf("certificate pinning verification failed: public key mismatch")
		}
	}

	return nil
}

// RemovePin removes a pinned certificate
func (cps *CertPinStore) RemovePin(nodeID string) {
	cps.mu.Lock()
	defer cps.mu.Unlock()
	delete(cps.pins, nodeID)
}

// GetPin returns the pinned hash for a node
func (cps *CertPinStore) GetPin(nodeID string) ([]byte, bool) {
	cps.mu.RLock()
	defer cps.mu.RUnlock()
	hash, exists := cps.pins[nodeID]
	return hash, exists
}
