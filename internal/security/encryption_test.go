package security

import (
	"crypto/rand"
	"testing"
)

func TestEncryptDecryptChaCha20Poly1305(t *testing.T) {
	key := make([]byte, 32) // ChaCha20Poly1305 requires 32-byte key
	rand.Read(key)

	plaintext := []byte("test message for encryption")

	ciphertext, err := EncryptChaCha20Poly1305(key, plaintext)
	if err != nil {
		t.Fatalf("Failed to encrypt: %v", err)
	}

	if len(ciphertext) <= len(plaintext) {
		t.Error("Ciphertext should be longer than plaintext (includes nonce)")
	}

	decrypted, err := DecryptChaCha20Poly1305(key, ciphertext)
	if err != nil {
		t.Fatalf("Failed to decrypt: %v", err)
	}

	if string(decrypted) != string(plaintext) {
		t.Errorf("Decrypted text doesn't match: got %s, want %s", string(decrypted), string(plaintext))
	}
}

func TestEncryptDecryptAES256GCM(t *testing.T) {
	key := make([]byte, 32) // AES-256 requires 32-byte key
	rand.Read(key)

	plaintext := []byte("test message for AES encryption")

	ciphertext, err := EncryptAES256GCM(key, plaintext)
	if err != nil {
		t.Fatalf("Failed to encrypt: %v", err)
	}

	if len(ciphertext) <= len(plaintext) {
		t.Error("Ciphertext should be longer than plaintext (includes nonce)")
	}

	decrypted, err := DecryptAES256GCM(key, ciphertext)
	if err != nil {
		t.Fatalf("Failed to decrypt: %v", err)
	}

	if string(decrypted) != string(plaintext) {
		t.Errorf("Decrypted text doesn't match: got %s, want %s", string(decrypted), string(plaintext))
	}
}

func TestEncryptDecryptChaCha20Poly1305_WrongKey(t *testing.T) {
	key1 := make([]byte, 32)
	rand.Read(key1)

	key2 := make([]byte, 32)
	rand.Read(key2)

	plaintext := []byte("test message")

	ciphertext, err := EncryptChaCha20Poly1305(key1, plaintext)
	if err != nil {
		t.Fatalf("Failed to encrypt: %v", err)
	}

	_, err = DecryptChaCha20Poly1305(key2, ciphertext)
	if err == nil {
		t.Error("Should fail to decrypt with wrong key")
	}
}

func TestEncryptDecryptAES256GCM_WrongKey(t *testing.T) {
	key1 := make([]byte, 32)
	rand.Read(key1)

	key2 := make([]byte, 32)
	rand.Read(key2)

	plaintext := []byte("test message")

	ciphertext, err := EncryptAES256GCM(key1, plaintext)
	if err != nil {
		t.Fatalf("Failed to encrypt: %v", err)
	}

	_, err = DecryptAES256GCM(key2, ciphertext)
	if err == nil {
		t.Error("Should fail to decrypt with wrong key")
	}
}

func TestGenerateKey(t *testing.T) {
	key, err := GenerateKey(32)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	if len(key) != 32 {
		t.Errorf("Key length should be 32 bytes, got %d", len(key))
	}

	// Generate another key and verify they're different
	key2, err := GenerateKey(32)
	if err != nil {
		t.Fatalf("Failed to generate second key: %v", err)
	}

	if string(key) == string(key2) {
		t.Error("Generated keys should be different")
	}
}

func TestEncryptDecryptChaCha20Poly1305_EmptyMessage(t *testing.T) {
	key := make([]byte, 32)
	rand.Read(key)

	plaintext := []byte{}

	ciphertext, err := EncryptChaCha20Poly1305(key, plaintext)
	if err != nil {
		t.Fatalf("Failed to encrypt empty message: %v", err)
	}

	decrypted, err := DecryptChaCha20Poly1305(key, ciphertext)
	if err != nil {
		t.Fatalf("Failed to decrypt empty message: %v", err)
	}

	if len(decrypted) != 0 {
		t.Error("Decrypted empty message should be empty")
	}
}
