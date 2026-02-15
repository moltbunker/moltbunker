package security

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// FuzzCertPinSaveLoad fuzzes the CertPinStore Save/Load roundtrip.
// It verifies that arbitrary node IDs survive serialization and that
// verification still works after save/load.
func FuzzCertPinSaveLoad(f *testing.F) {
	f.Add("node-1")
	f.Add("")
	f.Add("node with spaces")
	f.Add("node/with/slashes")
	f.Add("特殊字符")
	f.Add("a]b\"c\\d")
	f.Add("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")

	f.Fuzz(func(t *testing.T, nodeID string) {
		if nodeID == "" {
			return // empty node IDs are not useful
		}

		cert := generateFuzzCert(t)

		// Pin and save
		store1 := NewCertPinStore()
		store1.PinCertificate(nodeID, cert)

		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "pins.json")

		if err := store1.Save(path); err != nil {
			t.Fatalf("Save failed: %v", err)
		}

		// Load into new store
		store2 := NewCertPinStore()
		if err := store2.Load(path); err != nil {
			t.Fatalf("Load failed: %v", err)
		}

		// Verify roundtrip
		if store2.PinCount() != 1 {
			t.Fatalf("expected 1 pin after load, got %d", store2.PinCount())
		}

		if err := store2.VerifyCertificate(nodeID, cert); err != nil {
			t.Fatalf("verification failed after save/load: %v", err)
		}
	})
}

// FuzzCertPinLoadCorrupt fuzzes the Load path with arbitrary data.
// It verifies that the load function doesn't panic on arbitrary input.
func FuzzCertPinLoadCorrupt(f *testing.F) {
	f.Add([]byte(`[]`))
	f.Add([]byte(`[{"node_id":"a","hash":"abcd"}]`))
	f.Add([]byte(`not json`))
	f.Add([]byte{0xff, 0xfe, 0x00})
	f.Add([]byte(`[{"node_id":"","hash":""}]`))

	f.Fuzz(func(t *testing.T, data []byte) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "pins.json")

		if err := writeFile(path, data); err != nil {
			t.Skip("cannot write test file")
		}

		store := NewCertPinStore()
		// Must not panic — errors are acceptable
		_ = store.Load(path)
	})
}

func generateFuzzCert(t *testing.T) *x509.Certificate {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("keygen: %v", err)
	}

	tmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{Organization: []string{"Fuzz"}},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}

	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, pub, priv)
	if err != nil {
		t.Fatalf("create cert: %v", err)
	}

	cert, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatalf("parse cert: %v", err)
	}
	return cert
}

func writeFile(path string, data []byte) error {
	return os.WriteFile(path, data, 0644)
}
