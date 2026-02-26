package storage

import (
	"bytes"
	"testing"
)

func TestObjectEncryptor_RoundTrip(t *testing.T) {
	enc := NewObjectEncryptor()

	ownerKey, err := DeriveOwnerKey("0xAc1D8d6e25E54c05986E8bFa9b759063D5e69592")
	if err != nil {
		t.Fatalf("DeriveOwnerKey: %v", err)
	}

	plaintext := []byte("hello, encrypted storage!")

	result, err := enc.Encrypt(plaintext, ownerKey)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}

	// Ciphertext should be different from plaintext
	if bytes.Equal(result.Ciphertext, plaintext) {
		t.Error("ciphertext should not equal plaintext")
	}

	// Decrypt
	decrypted, err := enc.Decrypt(result.Ciphertext, result.EncryptedDEK, ownerKey)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("decrypted = %q, want %q", string(decrypted), string(plaintext))
	}
}

func TestObjectEncryptor_DifferentDEKs(t *testing.T) {
	enc := NewObjectEncryptor()

	ownerKey, err := DeriveOwnerKey("0xAc1D8d6e25E54c05986E8bFa9b759063D5e69592")
	if err != nil {
		t.Fatalf("DeriveOwnerKey: %v", err)
	}

	plaintext := []byte("same content")

	r1, err := enc.Encrypt(plaintext, ownerKey)
	if err != nil {
		t.Fatalf("Encrypt 1: %v", err)
	}

	r2, err := enc.Encrypt(plaintext, ownerKey)
	if err != nil {
		t.Fatalf("Encrypt 2: %v", err)
	}

	// Each encryption should use a different random DEK, producing different ciphertext
	if bytes.Equal(r1.Ciphertext, r2.Ciphertext) {
		t.Error("two encryptions of the same plaintext should produce different ciphertext")
	}

	// But both should decrypt to the same plaintext
	d1, _ := enc.Decrypt(r1.Ciphertext, r1.EncryptedDEK, ownerKey)
	d2, _ := enc.Decrypt(r2.Ciphertext, r2.EncryptedDEK, ownerKey)
	if !bytes.Equal(d1, d2) {
		t.Error("both should decrypt to the same plaintext")
	}
}

func TestObjectEncryptor_WrongKey(t *testing.T) {
	enc := NewObjectEncryptor()

	ownerKey, _ := DeriveOwnerKey("0xCorrectOwner")
	wrongKey, _ := DeriveOwnerKey("0xWrongOwner")

	plaintext := []byte("secret data")
	result, err := enc.Encrypt(plaintext, ownerKey)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}

	_, err = enc.Decrypt(result.Ciphertext, result.EncryptedDEK, wrongKey)
	if err == nil {
		t.Fatal("decrypting with wrong key should fail")
	}
}

func TestObjectEncryptor_EmptyContent(t *testing.T) {
	enc := NewObjectEncryptor()
	ownerKey, _ := DeriveOwnerKey("0xOwner")

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
	ownerKey, _ := DeriveOwnerKey("0xOwner")

	// 1MB of data
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

func TestDeriveOwnerKey_Deterministic(t *testing.T) {
	k1, err := DeriveOwnerKey("0xSameOwner")
	if err != nil {
		t.Fatalf("DeriveOwnerKey 1: %v", err)
	}

	k2, err := DeriveOwnerKey("0xSameOwner")
	if err != nil {
		t.Fatalf("DeriveOwnerKey 2: %v", err)
	}

	if !bytes.Equal(k1, k2) {
		t.Error("same owner should produce same key")
	}
}

func TestDeriveOwnerKey_DifferentOwners(t *testing.T) {
	k1, _ := DeriveOwnerKey("0xOwnerA")
	k2, _ := DeriveOwnerKey("0xOwnerB")

	if bytes.Equal(k1, k2) {
		t.Error("different owners should produce different keys")
	}
}
