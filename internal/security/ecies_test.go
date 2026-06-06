package security

import (
	"bytes"
	"crypto/rand"
	"testing"
)

func TestSealOpenX25519_RoundTrip(t *testing.T) {
	recipientPub, recipientPriv, err := GenerateX25519KeyPair()
	if err != nil {
		t.Fatalf("GenerateX25519KeyPair: %v", err)
	}

	plaintext := make([]byte, 32) // exec-key sized
	if _, err := rand.Read(plaintext); err != nil {
		t.Fatalf("rand: %v", err)
	}

	env, err := SealToX25519(recipientPub, plaintext)
	if err != nil {
		t.Fatalf("SealToX25519: %v", err)
	}
	if len(env.EphemeralPub) != X25519KeySize {
		t.Fatalf("ephemeral pub size = %d, want %d", len(env.EphemeralPub), X25519KeySize)
	}
	if len(env.Nonce) != NonceSize {
		t.Fatalf("nonce size = %d, want %d", len(env.Nonce), NonceSize)
	}
	if bytes.Equal(env.Ciphertext, plaintext) {
		t.Fatal("ciphertext equals plaintext (not encrypted)")
	}

	got, err := OpenFromX25519(recipientPriv, env)
	if err != nil {
		t.Fatalf("OpenFromX25519: %v", err)
	}
	if !bytes.Equal(got, plaintext) {
		t.Fatalf("round-trip mismatch: got %x want %x", got, plaintext)
	}
}

func TestSealOpenX25519_VariousSizes(t *testing.T) {
	recipientPub, recipientPriv, err := GenerateX25519KeyPair()
	if err != nil {
		t.Fatalf("GenerateX25519KeyPair: %v", err)
	}

	for _, size := range []int{0, 1, 16, 32, 64, 1024} {
		pt := make([]byte, size)
		if _, err := rand.Read(pt); err != nil {
			t.Fatalf("rand: %v", err)
		}
		env, err := SealToX25519(recipientPub, pt)
		if err != nil {
			t.Fatalf("seal size %d: %v", size, err)
		}
		got, err := OpenFromX25519(recipientPriv, env)
		if err != nil {
			t.Fatalf("open size %d: %v", size, err)
		}
		if !bytes.Equal(got, pt) {
			t.Fatalf("size %d round-trip mismatch", size)
		}
	}
}

func TestSealOpenX25519_WrongKeyFails(t *testing.T) {
	recipientPub, _, err := GenerateX25519KeyPair()
	if err != nil {
		t.Fatalf("GenerateX25519KeyPair: %v", err)
	}
	_, wrongPriv, err := GenerateX25519KeyPair()
	if err != nil {
		t.Fatalf("GenerateX25519KeyPair (wrong): %v", err)
	}

	plaintext := []byte("super secret exec key bytes here")
	env, err := SealToX25519(recipientPub, plaintext)
	if err != nil {
		t.Fatalf("SealToX25519: %v", err)
	}

	if _, err := OpenFromX25519(wrongPriv, env); err == nil {
		t.Fatal("expected decryption with wrong private key to fail, got nil error")
	}
}

func TestSealOpenX25519_TamperFails(t *testing.T) {
	recipientPub, recipientPriv, err := GenerateX25519KeyPair()
	if err != nil {
		t.Fatalf("GenerateX25519KeyPair: %v", err)
	}

	plaintext := []byte("0123456789abcdef0123456789abcdef")
	env, err := SealToX25519(recipientPub, plaintext)
	if err != nil {
		t.Fatalf("SealToX25519: %v", err)
	}

	// Tamper the ciphertext body (after the nonce prefix).
	t.Run("ciphertext", func(t *testing.T) {
		tampered := &X25519Envelope{
			EphemeralPub: append([]byte(nil), env.EphemeralPub...),
			Nonce:        append([]byte(nil), env.Nonce...),
			Ciphertext:   append([]byte(nil), env.Ciphertext...),
		}
		// Flip a bit in the GCM payload region (last byte = part of tag).
		tampered.Ciphertext[len(tampered.Ciphertext)-1] ^= 0x01
		if _, err := OpenFromX25519(recipientPriv, tampered); err == nil {
			t.Fatal("expected tampered ciphertext to fail authentication")
		}
	})

	// Tamper the ephemeral public key -> wrong shared secret -> auth failure.
	t.Run("ephemeral_pub", func(t *testing.T) {
		tampered := &X25519Envelope{
			EphemeralPub: append([]byte(nil), env.EphemeralPub...),
			Nonce:        append([]byte(nil), env.Nonce...),
			Ciphertext:   append([]byte(nil), env.Ciphertext...),
		}
		tampered.EphemeralPub[0] ^= 0x01
		if _, err := OpenFromX25519(recipientPriv, tampered); err == nil {
			t.Fatal("expected tampered ephemeral pubkey to fail")
		}
	})
}

func TestSealToX25519_BadKeySizes(t *testing.T) {
	if _, err := SealToX25519([]byte{1, 2, 3}, []byte("x")); err == nil {
		t.Fatal("expected error for short recipient public key")
	}
	if _, err := OpenFromX25519([]byte{1, 2, 3}, &X25519Envelope{}); err == nil {
		t.Fatal("expected error for short recipient private key")
	}
	if _, err := OpenFromX25519(make([]byte, X25519KeySize), nil); err == nil {
		t.Fatal("expected error for nil envelope")
	}
}

func TestSealToX25519_FreshEphemeralPerCall(t *testing.T) {
	recipientPub, _, err := GenerateX25519KeyPair()
	if err != nil {
		t.Fatalf("GenerateX25519KeyPair: %v", err)
	}
	pt := []byte("same plaintext both times--------")
	e1, err := SealToX25519(recipientPub, pt)
	if err != nil {
		t.Fatalf("seal 1: %v", err)
	}
	e2, err := SealToX25519(recipientPub, pt)
	if err != nil {
		t.Fatalf("seal 2: %v", err)
	}
	if bytes.Equal(e1.EphemeralPub, e2.EphemeralPub) {
		t.Fatal("ephemeral public keys should differ per Seal")
	}
	if bytes.Equal(e1.Ciphertext, e2.Ciphertext) {
		t.Fatal("ciphertexts should differ per Seal (fresh ephemeral + nonce)")
	}
}
