package security

import (
	"bytes"
	"crypto/rand"
	"testing"
)

// FuzzEncryptDecryptChaCha20 fuzzes the ChaCha20-Poly1305 encrypt/decrypt roundtrip.
// It verifies that any plaintext encrypts and decrypts back to the original,
// and that wrong keys fail to decrypt.
func FuzzEncryptDecryptChaCha20(f *testing.F) {
	// Seed corpus with representative inputs
	f.Add([]byte("hello world"))
	f.Add([]byte(""))
	f.Add([]byte("a"))
	f.Add([]byte("The quick brown fox jumps over the lazy dog"))
	f.Add(bytes.Repeat([]byte("X"), 1024))       // 1 KB
	f.Add(bytes.Repeat([]byte("\x00"), 256))      // null bytes
	f.Add(bytes.Repeat([]byte("\xff"), 256))       // all 0xFF
	f.Add([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}) // binary data
	f.Add([]byte(`{"container_id":"ctr-123","status":"running"}`)) // JSON payload

	f.Fuzz(func(t *testing.T, plaintext []byte) {
		// Generate a valid 32-byte key for ChaCha20-Poly1305
		key := make([]byte, 32)
		if _, err := rand.Read(key); err != nil {
			t.Fatalf("failed to generate key: %v", err)
		}

		// Encrypt
		ciphertext, err := EncryptChaCha20Poly1305(key, plaintext)
		if err != nil {
			t.Fatalf("EncryptChaCha20Poly1305 failed: %v", err)
		}

		// Ciphertext must be longer than plaintext (nonce + auth tag overhead)
		if len(ciphertext) <= len(plaintext) {
			t.Errorf("ciphertext (%d bytes) should be longer than plaintext (%d bytes)",
				len(ciphertext), len(plaintext))
		}

		// Decrypt with correct key
		decrypted, err := DecryptChaCha20Poly1305(key, ciphertext)
		if err != nil {
			t.Fatalf("DecryptChaCha20Poly1305 failed with correct key: %v", err)
		}

		// Verify roundtrip
		if !bytes.Equal(decrypted, plaintext) {
			t.Errorf("roundtrip failed: plaintext length %d, decrypted length %d",
				len(plaintext), len(decrypted))
		}

		// Generate a different key and verify decryption fails
		wrongKey := make([]byte, 32)
		if _, err := rand.Read(wrongKey); err != nil {
			t.Fatalf("failed to generate wrong key: %v", err)
		}
		// Ensure the wrong key is actually different
		if bytes.Equal(key, wrongKey) {
			// Astronomically unlikely but handle it
			wrongKey[0] ^= 0xff
		}

		_, err = DecryptChaCha20Poly1305(wrongKey, ciphertext)
		if err == nil {
			t.Error("DecryptChaCha20Poly1305 should fail with wrong key")
		}

		// Verify that tampering with ciphertext causes decryption failure
		if len(ciphertext) > 0 {
			tampered := make([]byte, len(ciphertext))
			copy(tampered, ciphertext)
			tampered[len(tampered)-1] ^= 0xff
			_, err = DecryptChaCha20Poly1305(key, tampered)
			if err == nil {
				t.Error("DecryptChaCha20Poly1305 should fail with tampered ciphertext")
			}
		}
	})
}

// FuzzEncryptDecryptAES256GCM fuzzes the AES-256-GCM encrypt/decrypt roundtrip.
// It verifies that any plaintext encrypts and decrypts back to the original,
// and that wrong keys fail to decrypt.
func FuzzEncryptDecryptAES256GCM(f *testing.F) {
	// Seed corpus with representative inputs
	f.Add([]byte("hello world"))
	f.Add([]byte(""))
	f.Add([]byte("a"))
	f.Add([]byte("The quick brown fox jumps over the lazy dog"))
	f.Add(bytes.Repeat([]byte("X"), 1024))       // 1 KB
	f.Add(bytes.Repeat([]byte("\x00"), 256))      // null bytes
	f.Add(bytes.Repeat([]byte("\xff"), 256))       // all 0xFF
	f.Add([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}) // binary data
	f.Add([]byte(`{"image_cid":"QmABC123","encrypted":true}`)) // JSON payload

	f.Fuzz(func(t *testing.T, plaintext []byte) {
		// Generate a valid 32-byte key for AES-256
		key := make([]byte, 32)
		if _, err := rand.Read(key); err != nil {
			t.Fatalf("failed to generate key: %v", err)
		}

		// Encrypt
		ciphertext, err := EncryptAES256GCM(key, plaintext)
		if err != nil {
			t.Fatalf("EncryptAES256GCM failed: %v", err)
		}

		// Ciphertext must be longer than plaintext (nonce + auth tag overhead)
		if len(ciphertext) <= len(plaintext) {
			t.Errorf("ciphertext (%d bytes) should be longer than plaintext (%d bytes)",
				len(ciphertext), len(plaintext))
		}

		// Decrypt with correct key
		decrypted, err := DecryptAES256GCM(key, ciphertext)
		if err != nil {
			t.Fatalf("DecryptAES256GCM failed with correct key: %v", err)
		}

		// Verify roundtrip
		if !bytes.Equal(decrypted, plaintext) {
			t.Errorf("roundtrip failed: plaintext length %d, decrypted length %d",
				len(plaintext), len(decrypted))
		}

		// Generate a different key and verify decryption fails
		wrongKey := make([]byte, 32)
		if _, err := rand.Read(wrongKey); err != nil {
			t.Fatalf("failed to generate wrong key: %v", err)
		}
		// Ensure the wrong key is actually different
		if bytes.Equal(key, wrongKey) {
			// Astronomically unlikely but handle it
			wrongKey[0] ^= 0xff
		}

		_, err = DecryptAES256GCM(wrongKey, ciphertext)
		if err == nil {
			t.Error("DecryptAES256GCM should fail with wrong key")
		}

		// Verify that tampering with ciphertext causes decryption failure
		if len(ciphertext) > 0 {
			tampered := make([]byte, len(ciphertext))
			copy(tampered, ciphertext)
			tampered[len(tampered)-1] ^= 0xff
			_, err = DecryptAES256GCM(key, tampered)
			if err == nil {
				t.Error("DecryptAES256GCM should fail with tampered ciphertext")
			}
		}
	})
}
