package identity

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/tls"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// newTestKeyManager creates a KeyManager with a fresh key pair for testing.
func newTestKeyManager(t *testing.T) *KeyManager {
	t.Helper()
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}
	return &KeyManager{
		privateKey: privateKey,
		publicKey:  privateKey.Public().(ed25519.PublicKey),
	}
}

func TestCertRotator_NeedsRenewal_NoCert(t *testing.T) {
	km := newTestKeyManager(t)
	tmpDir := t.TempDir()

	cr := NewCertRotator(km,
		filepath.Join(tmpDir, "cert.pem"),
		filepath.Join(tmpDir, "key.pem"),
	)

	// With no certificate loaded, NeedsRenewal should return true
	if !cr.NeedsRenewal() {
		t.Error("NeedsRenewal should return true when no certificate is loaded")
	}
}

func TestCertRotator_NeedsRenewal_FreshCert(t *testing.T) {
	km := newTestKeyManager(t)
	tmpDir := t.TempDir()
	certPath := filepath.Join(tmpDir, "cert.pem")
	keyPath := filepath.Join(tmpDir, "key.pem")

	cr := NewCertRotator(km, certPath, keyPath)
	cr.SetRenewBefore(30 * 24 * time.Hour) // 30 days

	// Generate a fresh certificate (valid for 1 year)
	ctx := context.Background()
	_, err := cr.Rotate(ctx)
	if err != nil {
		t.Fatalf("Failed to rotate: %v", err)
	}

	// A fresh certificate (1 year validity) should NOT need renewal with 30-day window
	if cr.NeedsRenewal() {
		t.Error("NeedsRenewal should return false for a fresh certificate")
	}
}

func TestCertRotator_NeedsRenewal_NearExpiry(t *testing.T) {
	km := newTestKeyManager(t)
	tmpDir := t.TempDir()
	certPath := filepath.Join(tmpDir, "cert.pem")
	keyPath := filepath.Join(tmpDir, "key.pem")

	cr := NewCertRotator(km, certPath, keyPath)

	// Generate a certificate first
	ctx := context.Background()
	_, err := cr.Rotate(ctx)
	if err != nil {
		t.Fatalf("Failed to rotate: %v", err)
	}

	// Set renewBefore to a very large window (2 years) so the 1-year cert is "near expiry"
	cr.SetRenewBefore(2 * 365 * 24 * time.Hour)

	if !cr.NeedsRenewal() {
		t.Error("NeedsRenewal should return true when renewBefore exceeds remaining validity")
	}
}

func TestCertRotator_NeedsRenewal_WithMockTime(t *testing.T) {
	km := newTestKeyManager(t)
	tmpDir := t.TempDir()
	certPath := filepath.Join(tmpDir, "cert.pem")
	keyPath := filepath.Join(tmpDir, "key.pem")

	cr := NewCertRotator(km, certPath, keyPath)
	cr.SetRenewBefore(30 * 24 * time.Hour)

	// Generate a certificate
	ctx := context.Background()
	_, err := cr.Rotate(ctx)
	if err != nil {
		t.Fatalf("Failed to rotate: %v", err)
	}

	// Advance mock time to 340 days from now (within 30-day renewBefore of 365-day cert)
	cr.nowFunc = func() time.Time {
		return time.Now().Add(340 * 24 * time.Hour)
	}

	if !cr.NeedsRenewal() {
		t.Error("NeedsRenewal should return true when mock time is near expiry")
	}
}

func TestCertRotator_Rotate_GeneratesValidCert(t *testing.T) {
	km := newTestKeyManager(t)
	tmpDir := t.TempDir()
	certPath := filepath.Join(tmpDir, "cert.pem")
	keyPath := filepath.Join(tmpDir, "key.pem")

	cr := NewCertRotator(km, certPath, keyPath)

	ctx := context.Background()
	cert, err := cr.Rotate(ctx)
	if err != nil {
		t.Fatalf("Rotate failed: %v", err)
	}

	if cert == nil {
		t.Fatal("Rotate returned nil certificate")
	}

	if len(cert.Certificate) == 0 {
		t.Error("Rotated certificate has no DER data")
	}

	// Verify the certificate is stored
	stored := cr.GetCertificate()
	if stored == nil {
		t.Error("GetCertificate returned nil after rotation")
	}
}

func TestCertRotator_Rotate_SavesToDisk(t *testing.T) {
	km := newTestKeyManager(t)
	tmpDir := t.TempDir()
	certPath := filepath.Join(tmpDir, "cert.pem")
	keyPath := filepath.Join(tmpDir, "key.pem")

	cr := NewCertRotator(km, certPath, keyPath)

	ctx := context.Background()
	_, err := cr.Rotate(ctx)
	if err != nil {
		t.Fatalf("Rotate failed: %v", err)
	}

	// Verify files were written by creating a new rotator that loads from disk
	cr2 := NewCertRotator(km, certPath, keyPath)
	loaded := cr2.GetCertificate()
	if loaded == nil {
		t.Fatal("Failed to load rotated certificate from disk via new CertRotator")
	}
	if len(loaded.Certificate) == 0 {
		t.Error("Loaded certificate has no DER data")
	}
}

func TestCertRotator_Rotate_NewCertEachTime(t *testing.T) {
	km := newTestKeyManager(t)
	tmpDir := t.TempDir()
	certPath := filepath.Join(tmpDir, "cert.pem")
	keyPath := filepath.Join(tmpDir, "key.pem")

	cr := NewCertRotator(km, certPath, keyPath)

	ctx := context.Background()
	cert1, err := cr.Rotate(ctx)
	if err != nil {
		t.Fatalf("First rotate failed: %v", err)
	}

	cert2, err := cr.Rotate(ctx)
	if err != nil {
		t.Fatalf("Second rotate failed: %v", err)
	}

	// The two certificates should be different (different serial numbers at minimum)
	if &cert1.Certificate[0] == &cert2.Certificate[0] {
		t.Error("Expected different certificate objects after two rotations")
	}
}

func TestCertRotator_GetCertificate_ThreadSafe(t *testing.T) {
	km := newTestKeyManager(t)
	tmpDir := t.TempDir()
	certPath := filepath.Join(tmpDir, "cert.pem")
	keyPath := filepath.Join(tmpDir, "key.pem")

	cr := NewCertRotator(km, certPath, keyPath)

	ctx := context.Background()
	// Seed initial certificate
	if _, err := cr.Rotate(ctx); err != nil {
		t.Fatalf("Initial rotate failed: %v", err)
	}

	// Concurrent readers and one writer
	var wg sync.WaitGroup
	const readers = 50
	const rotations = 5

	// Start readers
	for i := 0; i < readers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				cert := cr.GetCertificate()
				if cert == nil {
					t.Error("GetCertificate returned nil during concurrent access")
					return
				}
			}
		}()
	}

	// Start writer
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < rotations; i++ {
			if _, err := cr.Rotate(ctx); err != nil {
				t.Errorf("Rotate failed during concurrent access: %v", err)
				return
			}
		}
	}()

	wg.Wait()
}

func TestCertRotator_OnRotationCallback(t *testing.T) {
	km := newTestKeyManager(t)
	tmpDir := t.TempDir()
	certPath := filepath.Join(tmpDir, "cert.pem")
	keyPath := filepath.Join(tmpDir, "key.pem")

	cr := NewCertRotator(km, certPath, keyPath)

	var callbackCalled atomic.Int32
	var callbackCert *tls.Certificate
	cr.SetOnRotation(func(cert *tls.Certificate) {
		callbackCalled.Add(1)
		callbackCert = cert
	})

	ctx := context.Background()
	rotatedCert, err := cr.Rotate(ctx)
	if err != nil {
		t.Fatalf("Rotate failed: %v", err)
	}

	if callbackCalled.Load() != 1 {
		t.Errorf("Expected callback to be called once, got %d", callbackCalled.Load())
	}

	if callbackCert == nil {
		t.Error("Callback received nil certificate")
	}

	if callbackCert != rotatedCert {
		t.Error("Callback certificate does not match rotated certificate")
	}
}

func TestCertRotator_StartStop(t *testing.T) {
	km := newTestKeyManager(t)
	tmpDir := t.TempDir()
	certPath := filepath.Join(tmpDir, "cert.pem")
	keyPath := filepath.Join(tmpDir, "key.pem")

	cr := NewCertRotator(km, certPath, keyPath)
	cr.SetCheckInterval(50 * time.Millisecond)

	ctx := context.Background()
	cr.Start(ctx)

	// Wait briefly for the initial rotation to happen
	time.Sleep(100 * time.Millisecond)

	// A certificate should be available after start (initial rotation)
	cert := cr.GetCertificate()
	if cert == nil {
		t.Error("Expected certificate to be available after Start")
	}

	cr.Stop()

	// Verify Stop is idempotent
	cr.Stop()
}

func TestCertRotator_StartTriggersRenewal(t *testing.T) {
	km := newTestKeyManager(t)
	tmpDir := t.TempDir()
	certPath := filepath.Join(tmpDir, "cert.pem")
	keyPath := filepath.Join(tmpDir, "key.pem")

	cr := NewCertRotator(km, certPath, keyPath)
	cr.SetCheckInterval(50 * time.Millisecond)
	// Make the renewBefore window absurdly large so any cert triggers renewal
	cr.SetRenewBefore(100 * 365 * 24 * time.Hour)

	var rotationCount atomic.Int32
	cr.SetOnRotation(func(cert *tls.Certificate) {
		rotationCount.Add(1)
	})

	ctx := context.Background()
	cr.Start(ctx)

	// Wait for initial rotation + at least one periodic check
	time.Sleep(200 * time.Millisecond)

	cr.Stop()

	count := rotationCount.Load()
	if count < 2 {
		t.Errorf("Expected at least 2 rotations (initial + periodic), got %d", count)
	}
}

func TestCertRotator_ContextCancellation(t *testing.T) {
	km := newTestKeyManager(t)
	tmpDir := t.TempDir()
	certPath := filepath.Join(tmpDir, "cert.pem")
	keyPath := filepath.Join(tmpDir, "key.pem")

	cr := NewCertRotator(km, certPath, keyPath)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := cr.Rotate(ctx)
	if err == nil {
		t.Error("Expected error when context is already cancelled")
	}
}

func TestCertRotator_LoadFromDisk(t *testing.T) {
	km := newTestKeyManager(t)
	tmpDir := t.TempDir()
	certPath := filepath.Join(tmpDir, "cert.pem")
	keyPath := filepath.Join(tmpDir, "key.pem")

	// First rotator generates and saves a certificate
	cr1 := NewCertRotator(km, certPath, keyPath)
	ctx := context.Background()
	_, err := cr1.Rotate(ctx)
	if err != nil {
		t.Fatalf("Failed to rotate: %v", err)
	}

	// Second rotator should load the certificate from disk
	cr2 := NewCertRotator(km, certPath, keyPath)
	cert := cr2.GetCertificate()
	if cert == nil {
		t.Error("Second CertRotator should have loaded certificate from disk")
	}

	if len(cert.Certificate) == 0 {
		t.Error("Loaded certificate has no DER data")
	}
}

func TestCertRotator_StartIdempotent(t *testing.T) {
	km := newTestKeyManager(t)
	tmpDir := t.TempDir()
	certPath := filepath.Join(tmpDir, "cert.pem")
	keyPath := filepath.Join(tmpDir, "key.pem")

	cr := NewCertRotator(km, certPath, keyPath)
	cr.SetCheckInterval(50 * time.Millisecond)

	ctx := context.Background()

	// Start twice should not panic or start duplicate goroutines
	cr.Start(ctx)
	cr.Start(ctx)

	time.Sleep(100 * time.Millisecond)
	cr.Stop()
}
