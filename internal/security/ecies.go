package security

import (
	"crypto/sha256"
	"fmt"
	"io"

	"golang.org/x/crypto/curve25519"
	"golang.org/x/crypto/hkdf"
)

// ECIES (Elliptic Curve Integrated Encryption Scheme) for sealing a small blob
// (e.g. a 32-byte exec key) to a recipient's STABLE X25519 public key.
//
// Scheme: ephemeral-static X25519 ECDH -> HKDF-SHA256 -> AES-256-GCM.
//
//	eph_pub, eph_priv = X25519 keypair (fresh per Seal)
//	shared            = X25519(eph_priv, recipient_pub)
//	key               = HKDF-SHA256(shared, salt=eph_pub||recipient_pub, info="moltbunker-ecies-v1")
//	ciphertext        = AES-256-GCM(key, plaintext)   // nonce is prepended by EncryptAES256GCM
//
// The envelope carries the ephemeral public key so the recipient can recompute
// the shared secret with its stable private key. Domain separation ("v1" info)
// keeps this key ladder distinct from the deployment-DEK ladder in
// deployment_encryption.go (which uses HKDF-SHA3-256, info "moltbunker-deployment-key").

const (
	// eciesInfo is the HKDF info string. Distinct from the deployment-key ladder.
	eciesInfo = "moltbunker-ecies-v1"
)

// X25519Envelope is a self-contained ECIES ciphertext sealed to a recipient
// X25519 public key. All fields are required to decrypt.
type X25519Envelope struct {
	// EphemeralPub is the sender's ephemeral X25519 public key (32 bytes).
	EphemeralPub []byte `json:"ephemeral_pub"`
	// Nonce is retained for wire compatibility. The AES-GCM nonce is already
	// prepended to Ciphertext by EncryptAES256GCM, so this field is the raw
	// 12-byte GCM nonce extracted for callers that carry it in a separate
	// transport field. OpenFromX25519 does not require it to be set (it reads
	// the nonce from the Ciphertext prefix).
	Nonce []byte `json:"nonce,omitempty"`
	// Ciphertext is nonce(12) || AES-256-GCM(ciphertext) || tag(16).
	Ciphertext []byte `json:"ciphertext"`
}

// X25519PublicFromPrivate recomputes the X25519 public key for a given private
// key. Useful for callers that persist only the private key.
func X25519PublicFromPrivate(privateKey []byte) ([]byte, error) {
	if len(privateKey) != X25519KeySize {
		return nil, fmt.Errorf("ecies: invalid private key size: expected %d, got %d", X25519KeySize, len(privateKey))
	}
	pub, err := curve25519.X25519(privateKey, curve25519.Basepoint)
	if err != nil {
		return nil, fmt.Errorf("ecies: derive public key: %w", err)
	}
	return pub, nil
}

// eciesDeriveKey derives the AES-256 key from the ECDH shared secret, binding
// both the ephemeral and recipient public keys into the HKDF salt.
func eciesDeriveKey(shared, ephemeralPub, recipientPub []byte) ([]byte, error) {
	salt := make([]byte, 0, len(ephemeralPub)+len(recipientPub))
	salt = append(salt, ephemeralPub...)
	salt = append(salt, recipientPub...)

	r := hkdf.New(sha256.New, shared, salt, []byte(eciesInfo))
	key := make([]byte, DEKSize)
	if _, err := io.ReadFull(r, key); err != nil {
		return nil, fmt.Errorf("ecies: derive key: %w", err)
	}
	return key, nil
}

// SealToX25519 encrypts plaintext to a recipient's stable X25519 public key,
// returning a self-contained envelope.
func SealToX25519(recipientPub, plaintext []byte) (*X25519Envelope, error) {
	if len(recipientPub) != X25519KeySize {
		return nil, fmt.Errorf("ecies: invalid recipient public key size: expected %d, got %d", X25519KeySize, len(recipientPub))
	}

	ephemeralPub, ephemeralPriv, err := GenerateX25519KeyPair()
	if err != nil {
		return nil, fmt.Errorf("ecies: generate ephemeral key: %w", err)
	}

	shared, err := curve25519.X25519(ephemeralPriv, recipientPub)
	if err != nil {
		// X25519 errors on low-order / all-zero outputs.
		return nil, fmt.Errorf("ecies: compute shared secret: %w", err)
	}

	key, err := eciesDeriveKey(shared, ephemeralPub, recipientPub)
	if err != nil {
		return nil, err
	}

	ciphertext, err := EncryptAES256GCM(key, plaintext)
	if err != nil {
		return nil, fmt.Errorf("ecies: encrypt: %w", err)
	}

	// Extract the GCM nonce prefix for callers that carry it separately.
	var nonce []byte
	if len(ciphertext) >= NonceSize {
		nonce = make([]byte, NonceSize)
		copy(nonce, ciphertext[:NonceSize])
	}

	return &X25519Envelope{
		EphemeralPub: ephemeralPub,
		Nonce:        nonce,
		Ciphertext:   ciphertext,
	}, nil
}

// OpenFromX25519 decrypts an envelope using the recipient's stable X25519
// private key. The recipient public key is recomputed from the private key so
// the HKDF salt matches what SealToX25519 used.
func OpenFromX25519(recipientPriv []byte, env *X25519Envelope) ([]byte, error) {
	if env == nil {
		return nil, fmt.Errorf("ecies: nil envelope")
	}
	if len(recipientPriv) != X25519KeySize {
		return nil, fmt.Errorf("ecies: invalid recipient private key size: expected %d, got %d", X25519KeySize, len(recipientPriv))
	}
	if len(env.EphemeralPub) != X25519KeySize {
		return nil, fmt.Errorf("ecies: invalid ephemeral public key size: expected %d, got %d", X25519KeySize, len(env.EphemeralPub))
	}

	recipientPub, err := curve25519.X25519(recipientPriv, curve25519.Basepoint)
	if err != nil {
		return nil, fmt.Errorf("ecies: derive recipient public key: %w", err)
	}

	shared, err := curve25519.X25519(recipientPriv, env.EphemeralPub)
	if err != nil {
		return nil, fmt.Errorf("ecies: compute shared secret: %w", err)
	}

	key, err := eciesDeriveKey(shared, env.EphemeralPub, recipientPub)
	if err != nil {
		return nil, err
	}

	plaintext, err := DecryptAES256GCM(key, env.Ciphertext)
	if err != nil {
		return nil, fmt.Errorf("ecies: decrypt: %w", err)
	}
	return plaintext, nil
}
