package security

import (
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// maxPinnedCerts limits the number of pinned certificates to prevent
// unbounded memory growth from TOFU with high peer churn.
const maxPinnedCerts = 10000

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

	// If updating an existing pin, no capacity check needed
	if _, exists := cps.pins[nodeID]; !exists && len(cps.pins) >= maxPinnedCerts {
		// Evict one arbitrary entry to make room
		for id := range cps.pins {
			delete(cps.pins, id)
			break
		}
	}

	// Hash the public key
	hash := sha256.Sum256(cert.RawSubjectPublicKeyInfo)
	cps.pins[nodeID] = hash[:]
}

// VerifyCertificate verifies a certificate against pinned public key
func (cps *CertPinStore) VerifyCertificate(nodeID string, cert *x509.Certificate) error {
	// First check if we have a pin
	cps.mu.RLock()
	pinnedHash, exists := cps.pins[nodeID]
	cps.mu.RUnlock()

	if !exists {
		// First time seeing this node - use TOFU (Trust On First Use)
		cps.PinCertificate(nodeID, cert)
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

// PinCount returns the number of pinned certificates
func (cps *CertPinStore) PinCount() int {
	cps.mu.RLock()
	defer cps.mu.RUnlock()
	return len(cps.pins)
}

// pinRecord is the serializable representation of a pinned certificate hash.
type pinRecord struct {
	NodeID string `json:"node_id"`
	Hash   string `json:"hash"` // hex-encoded SHA-256 of public key
}

// Save persists the pin store to a JSON file using an atomic write
// (write to .tmp then rename).
func (cps *CertPinStore) Save(path string) error {
	cps.mu.RLock()
	records := make([]pinRecord, 0, len(cps.pins))
	for nodeID, hash := range cps.pins {
		records = append(records, pinRecord{
			NodeID: nodeID,
			Hash:   hex.EncodeToString(hash),
		})
	}
	cps.mu.RUnlock()

	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal cert pin store: %w", err)
	}

	// Ensure the parent directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("create directory for cert pin store: %w", err)
	}

	// Atomic write: write to .tmp then rename
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0600); err != nil {
		return fmt.Errorf("write cert pin store temp file: %w", err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		// Clean up temp file on rename failure
		os.Remove(tmpPath)
		return fmt.Errorf("rename cert pin store file: %w", err)
	}

	return nil
}

// Load reads pinned certificates from a JSON file. If the file does not
// exist, the pin store is left empty (no error).
func (cps *CertPinStore) Load(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read cert pin store: %w", err)
	}

	var records []pinRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return fmt.Errorf("unmarshal cert pin store: %w", err)
	}

	cps.mu.Lock()
	defer cps.mu.Unlock()

	cps.pins = make(map[string][]byte, len(records))
	for _, rec := range records {
		hash, err := hex.DecodeString(rec.Hash)
		if err != nil {
			return fmt.Errorf("decode pin hash for node %s: %w", rec.NodeID, err)
		}
		cps.pins[rec.NodeID] = hash
	}

	return nil
}
