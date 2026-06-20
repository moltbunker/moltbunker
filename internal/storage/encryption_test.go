package storage

import (
	"bytes"
	"testing"

	"github.com/moltbunker/moltbunker/internal/security"
)

// --- New X25519 envelope path ---

func TestOwnerKeyEncryptor_RoundTrip(t *testing.T) {
	pub, priv, err := security.GenerateX25519KeyPair()
	if err != nil {
		t.Fatalf("GenerateX25519KeyPair: %v", err)
	}

	enc := NewOwnerKeyEncryptor()
	plaintext := []byte("hello, X25519-sealed storage!")

	res, err := enc.Encrypt(plaintext, pub)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	if bytes.Equal(res.Ciphertext, plaintext) {
		t.Error("ciphertext should not equal plaintext")
	}
	if res.SealedDEK == nil || len(res.SealedDEK.EphemeralPub) != security.X25519KeySize {
		t.Fatal("sealed DEK envelope is malformed")
	}

	got, err := enc.Decrypt(res.Ciphertext, res.SealedDEK, priv)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}
	if !bytes.Equal(got, plaintext) {
		t.Errorf("decrypted = %q, want %q", got, plaintext)
	}
}

func TestOwnerKeyEncryptor_WrongKey(t *testing.T) {
	pub, _, err := security.GenerateX25519KeyPair()
	if err != nil {
		t.Fatalf("keypair: %v", err)
	}
	_, wrongPriv, err := security.GenerateX25519KeyPair()
	if err != nil {
		t.Fatalf("keypair: %v", err)
	}

	enc := NewOwnerKeyEncryptor()
	res, err := enc.Encrypt([]byte("secret"), pub)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}

	if _, err := enc.Decrypt(res.Ciphertext, res.SealedDEK, wrongPriv); err == nil {
		t.Fatal("decrypting with wrong private key should fail")
	}
}

func TestOwnerKeyEncryptor_DifferentDEKs(t *testing.T) {
	pub, priv, _ := security.GenerateX25519KeyPair()
	enc := NewOwnerKeyEncryptor()
	plaintext := []byte("same content")

	r1, _ := enc.Encrypt(plaintext, pub)
	r2, _ := enc.Encrypt(plaintext, pub)

	if bytes.Equal(r1.Ciphertext, r2.Ciphertext) {
		t.Error("two encryptions of the same plaintext should differ (random DEK)")
	}
	d1, _ := enc.Decrypt(r1.Ciphertext, r1.SealedDEK, priv)
	d2, _ := enc.Decrypt(r2.Ciphertext, r2.SealedDEK, priv)
	if !bytes.Equal(d1, d2) {
		t.Error("both should decrypt to the same plaintext")
	}
}

func TestOwnerKeyEncryptor_InvalidPubKey(t *testing.T) {
	enc := NewOwnerKeyEncryptor()
	if _, err := enc.Encrypt([]byte("x"), []byte("too short")); err == nil {
		t.Fatal("expected error for invalid public key size")
	}
}

func TestMarshalSealedDEK_RoundTrip(t *testing.T) {
	pub, priv, _ := security.GenerateX25519KeyPair()
	enc := NewOwnerKeyEncryptor()
	res, err := enc.Encrypt([]byte("payload"), pub)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}

	blob, err := MarshalSealedDEK(res.SealedDEK)
	if err != nil {
		t.Fatalf("MarshalSealedDEK: %v", err)
	}
	env, ok := looksLikeX25519Envelope(blob)
	if !ok {
		t.Fatal("serialized envelope should be recognized as X25519 envelope")
	}
	got, err := enc.Decrypt(res.Ciphertext, env, priv)
	if err != nil {
		t.Fatalf("Decrypt after round-trip: %v", err)
	}
	if string(got) != "payload" {
		t.Errorf("got %q", got)
	}
}

func TestLooksLikeX25519Envelope_LegacyBlob(t *testing.T) {
	// A legacy raw AES-GCM-wrapped DEK is NOT a JSON envelope.
	legacyKey, _ := legacyDeriveOwnerKey("0xOwner")
	legacy := NewObjectEncryptor()
	res, err := legacy.Encrypt([]byte("data"), legacyKey)
	if err != nil {
		t.Fatalf("legacy Encrypt: %v", err)
	}
	if _, ok := looksLikeX25519Envelope(res.EncryptedDEK); ok {
		t.Error("legacy wrapped DEK should NOT be detected as an X25519 envelope")
	}
}

// --- Legacy AES-GCM-wrapped-DEK path (back-compat) ---

func TestObjectEncryptor_RoundTrip(t *testing.T) {
	enc := NewObjectEncryptor()
	ownerKey, err := legacyDeriveOwnerKey("0xAc1D8d6e25E54c05986E8bFa9b759063D5e69592")
	if err != nil {
		t.Fatalf("legacyDeriveOwnerKey: %v", err)
	}
	plaintext := []byte("hello, encrypted storage!")

	result, err := enc.Encrypt(plaintext, ownerKey)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	if bytes.Equal(result.Ciphertext, plaintext) {
		t.Error("ciphertext should not equal plaintext")
	}
	decrypted, err := enc.Decrypt(result.Ciphertext, result.EncryptedDEK, ownerKey)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}
	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("decrypted = %q, want %q", decrypted, plaintext)
	}
}

func TestObjectEncryptor_WrongKey(t *testing.T) {
	enc := NewObjectEncryptor()
	ownerKey, _ := legacyDeriveOwnerKey("0xCorrectOwner")
	wrongKey, _ := legacyDeriveOwnerKey("0xWrongOwner")

	plaintext := []byte("secret data")
	result, err := enc.Encrypt(plaintext, ownerKey)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	if _, err := enc.Decrypt(result.Ciphertext, result.EncryptedDEK, wrongKey); err == nil {
		t.Fatal("decrypting with wrong key should fail")
	}
}

func TestObjectEncryptor_EmptyContent(t *testing.T) {
	enc := NewObjectEncryptor()
	ownerKey, _ := legacyDeriveOwnerKey("0xOwner")

	result, err := enc.Encrypt([]byte{}, ownerKey)
	if err != nil {
		t.Fatalf("Encrypt empty: %v", err)
	}
	decrypted, err := enc.Decrypt(result.Ciphertext, result.EncryptedDEK, ownerKey)
	if err != nil {
		t.Fatalf("Decrypt empty: %v", err)
	}
	if len(decrypted) != 0 {
		t.Errorf("decrypted length = %d, want 0", len(decrypted))
	}
}

func TestObjectEncryptor_LargeContent(t *testing.T) {
	enc := NewObjectEncryptor()
	ownerKey, _ := legacyDeriveOwnerKey("0xOwner")

	plaintext := make([]byte, 1024*1024)
	for i := range plaintext {
		plaintext[i] = byte(i % 256)
	}
	result, err := enc.Encrypt(plaintext, ownerKey)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	decrypted, err := enc.Decrypt(result.Ciphertext, result.EncryptedDEK, ownerKey)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}
	if !bytes.Equal(decrypted, plaintext) {
		t.Error("large content round-trip failed")
	}
}

func TestLegacyDeriveOwnerKey_Deterministic(t *testing.T) {
	k1, err := legacyDeriveOwnerKey("0xSameOwner")
	if err != nil {
		t.Fatalf("legacyDeriveOwnerKey 1: %v", err)
	}
	k2, err := legacyDeriveOwnerKey("0xSameOwner")
	if err != nil {
		t.Fatalf("legacyDeriveOwnerKey 2: %v", err)
	}
	if !bytes.Equal(k1, k2) {
		t.Error("same owner should produce same key")
	}
}
