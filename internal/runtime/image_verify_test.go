package runtime

import (
	"crypto/ed25519"
	"encoding/hex"
	"errors"
	"testing"
)

func mustGenKey(t *testing.T) (pubHex string, priv ed25519.PrivateKey) {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	return hex.EncodeToString(pub), priv
}

func digestForBytes(b []byte) ImageDigest {
	return ImageDigest("sha256:" + hex.EncodeToString(b))
}

func TestEdImageVerifier_Verify(t *testing.T) {
	pubHex, priv := mustGenKey(t)
	_, otherPriv := mustGenKey(t)

	digest := digestForBytes(make([]byte, 32))
	otherBytes := make([]byte, 32)
	otherBytes[0] = 0xff
	otherDigest := digestForBytes(otherBytes)

	cases := []struct {
		name   string
		sig    *ImageSignature
		policy TrustPolicy
		want   error
	}{
		{
			name:   "unsigned allowed when not required",
			sig:    nil,
			policy: TrustPolicy{RequireSignature: false},
			want:   nil,
		},
		{
			name:   "unsigned rejected when required",
			sig:    nil,
			policy: TrustPolicy{RequireSignature: true},
			want:   ErrSignatureRequired,
		},
		{
			name:   "valid signature from trusted publisher",
			sig:    SignImageDigest(digest, priv),
			policy: TrustPolicy{RequireSignature: true, TrustedPublishers: []string{pubHex}},
			want:   nil,
		},
		{
			name:   "valid signature but publisher not in trust list",
			sig:    SignImageDigest(digest, priv),
			policy: TrustPolicy{RequireSignature: true, TrustedPublishers: nil},
			want:   ErrUntrustedPublisher,
		},
		{
			name: "claimed digest does not match image digest",
			sig:  SignImageDigest(otherDigest, priv),
			policy: TrustPolicy{
				RequireSignature:  true,
				TrustedPublishers: []string{pubHex},
			},
			want: ErrDigestMismatch,
		},
		{
			name: "signature minted by different key claiming trusted publisher",
			sig: &ImageSignature{
				Digest:      digest,
				PublisherID: pubHex, // claims pubHex
				Signature:   ed25519.Sign(otherPriv, digestBytes(digest)),
			},
			policy: TrustPolicy{RequireSignature: true, TrustedPublishers: []string{pubHex}},
			want:   ErrInvalidSignature,
		},
		{
			name: "malformed publisher id (non-hex)",
			sig: &ImageSignature{
				Digest:      digest,
				PublisherID: "zzznot-hex",
				Signature:   make([]byte, ed25519.SignatureSize),
			},
			policy: TrustPolicy{RequireSignature: true, TrustedPublishers: []string{"zzznot-hex"}},
			want:   ErrMalformedPublisher,
		},
		{
			name: "publisher id wrong length",
			sig: &ImageSignature{
				Digest:      digest,
				PublisherID: "deadbeef",
				Signature:   make([]byte, ed25519.SignatureSize),
			},
			policy: TrustPolicy{RequireSignature: true, TrustedPublishers: []string{"deadbeef"}},
			want:   ErrMalformedPublisher,
		},
	}

	v := NewEdImageVerifier()
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := v.Verify(digest, tc.sig, tc.policy)
			if !errors.Is(err, tc.want) {
				t.Fatalf("Verify() = %v, want %v", err, tc.want)
			}
		})
	}
}

// TestDigestBytes_Canonicalization verifies that the signed payload is stable
// across calls (signature determinism) and that non-sha256 forms fall back to
// hashing the digest string.
func TestDigestBytes_Canonicalization(t *testing.T) {
	d := digestForBytes(make([]byte, 32))
	a := digestBytes(d)
	b := digestBytes(d)
	if string(a) != string(b) {
		t.Fatalf("digestBytes not deterministic")
	}
	if len(a) != 32 {
		t.Fatalf("sha256 digest should produce 32 bytes, got %d", len(a))
	}

	// Unknown algo falls back to SHA-256 of the string form.
	weird := digestBytes(ImageDigest("blake3:abcd"))
	if len(weird) != 32 {
		t.Fatalf("fallback should produce 32-byte sha256, got %d", len(weird))
	}
}

// TestEdImageVerifier_UnsignedAllowed_WithEmptyTrustList ensures that when
// RequireSignature is false, the TrustedPublishers list is not consulted at
// all (an unsigned image is allowed even if no publishers are trusted).
func TestEdImageVerifier_UnsignedAllowed_WithEmptyTrustList(t *testing.T) {
	v := NewEdImageVerifier()
	err := v.Verify(digestForBytes(make([]byte, 32)), nil, TrustPolicy{
		RequireSignature:  false,
		TrustedPublishers: nil,
	})
	if err != nil {
		t.Fatalf("unsigned image should be allowed, got %v", err)
	}
}
