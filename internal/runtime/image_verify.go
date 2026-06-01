package runtime

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

// R3 — Image signature verification.
//
// This file provides Ed25519-based image-signature verification that runs in the
// container-pull pipeline before a container is created. It is intentionally
// minimal and dependency-free; a cosign/sigstore-compatible verifier can wrap
// on top later by implementing the ImageVerifier interface.
//
// Design boundaries:
//  - The Verifier is pure crypto: given (digest, signature, policy) it returns
//    nil iff the image is acceptable. It does NOT know where the signature
//    came from (OCI label, sidecar, on-chain). Source plumbing is a separate
//    concern handled by callers.
//  - Trust policies are caller-supplied (per-tenant or per-deployment).
//  - Unsigned images are allowed by default; flip RequireSignature to deny.

// ImageDigest is an opaque image identifier — typically "sha256:<hex>" as
// produced by containerd.Image.Target().Digest.String().
type ImageDigest string

// ImageSignature is an Ed25519 signature over the bytes returned by
// digestBytes(Digest). PublisherID is the hex-encoded 32-byte Ed25519 public
// key that produced the signature.
type ImageSignature struct {
	Digest      ImageDigest
	PublisherID string
	Signature   []byte
}

// TrustPolicy controls what an ImageVerifier accepts.
type TrustPolicy struct {
	// RequireSignature: if true, unsigned images are rejected with
	// ErrSignatureRequired.
	RequireSignature bool

	// TrustedPublishers is the set of hex-encoded Ed25519 public keys that are
	// permitted to sign images. An empty list combined with RequireSignature=true
	// rejects every image — useful as a "deny all" sentinel.
	TrustedPublishers []string
}

// ImageVerifier verifies an image signature against a TrustPolicy.
type ImageVerifier interface {
	// Verify returns nil iff the (digest, sig, policy) tuple is acceptable.
	// A nil sig with RequireSignature=false returns nil (allowed unsigned).
	Verify(digest ImageDigest, sig *ImageSignature, policy TrustPolicy) error
}

// EdImageVerifier is the default Ed25519-based ImageVerifier.
type EdImageVerifier struct{}

// NewEdImageVerifier returns the default verifier.
func NewEdImageVerifier() *EdImageVerifier { return &EdImageVerifier{} }

// Sentinel errors. Test for these with errors.Is.
var (
	// ErrSignatureRequired is returned when policy requires a signature but
	// none is provided.
	ErrSignatureRequired = errors.New("image signature required by trust policy")

	// ErrUntrustedPublisher is returned when the signing key is not in the
	// trust list.
	ErrUntrustedPublisher = errors.New("image signed by untrusted publisher")

	// ErrInvalidSignature is returned when the Ed25519 signature does not
	// verify against the digest.
	ErrInvalidSignature = errors.New("image signature does not verify")

	// ErrDigestMismatch is returned when the signature's claimed digest field
	// disagrees with the image's actual digest.
	ErrDigestMismatch = errors.New("image signature digest does not match image")

	// ErrMalformedPublisher is returned when the publisher id is not a valid
	// 32-byte hex string.
	ErrMalformedPublisher = errors.New("malformed publisher id")
)

// Verify implements ImageVerifier.
func (v *EdImageVerifier) Verify(digest ImageDigest, sig *ImageSignature, policy TrustPolicy) error {
	if sig == nil {
		if policy.RequireSignature {
			return ErrSignatureRequired
		}
		return nil
	}

	if sig.Digest != digest {
		return ErrDigestMismatch
	}

	if !publisherInList(sig.PublisherID, policy.TrustedPublishers) {
		return ErrUntrustedPublisher
	}

	pubBytes, err := hex.DecodeString(sig.PublisherID)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrMalformedPublisher, err)
	}
	if len(pubBytes) != ed25519.PublicKeySize {
		return fmt.Errorf("%w: length %d", ErrMalformedPublisher, len(pubBytes))
	}

	if !ed25519.Verify(ed25519.PublicKey(pubBytes), digestBytes(digest), sig.Signature) {
		return ErrInvalidSignature
	}
	return nil
}

// publisherInList reports whether id appears in list (constant-time within a
// single comparison; the linear scan does not leak which entry matched).
func publisherInList(id string, list []string) bool {
	for _, p := range list {
		if p == id {
			return true
		}
	}
	return false
}

// digestBytes returns the canonical signed payload for an ImageDigest.
//
// For "sha256:<hex>" forms (the common OCI case) it returns the raw 32-byte
// hash. For any other form it returns SHA-256 of the entire digest string —
// so other digest algorithms can coexist without breaking signature
// determinism.
func digestBytes(d ImageDigest) []byte {
	s := string(d)
	if strings.HasPrefix(s, "sha256:") {
		if b, err := hex.DecodeString(s[len("sha256:"):]); err == nil && len(b) == sha256.Size {
			return b
		}
	}
	h := sha256.Sum256([]byte(s))
	return h[:]
}

// SignImageDigest is a test/dev helper that produces a valid ImageSignature
// over the given digest using priv. It is exported so callers (and tests in
// other packages) can mint signatures without re-deriving the canonical
// payload encoding.
func SignImageDigest(digest ImageDigest, priv ed25519.PrivateKey) *ImageSignature {
	pub := priv.Public().(ed25519.PublicKey)
	return &ImageSignature{
		Digest:      digest,
		PublisherID: hex.EncodeToString(pub),
		Signature:   ed25519.Sign(priv, digestBytes(digest)),
	}
}

// R3-sourcing: As of SEC-09 the daemon zone sources ImageSignature + the trust
// list directly from the deploy request (DeployRequest.ImageSignature /
// .TrustedPublishers / .RequireSignature, plumbed through SecureContainerConfig
// in internal/daemon/container_manager.go and replication.go). The
// caller-supplied path is the simplest source and is now wired end-to-end.
//
// TODO(R3-sourcing): registry-/chain-side resolution is still a follow-up so the
// daemon can RESOLVE a signature when the request omits one. Candidates:
//
//   1. OCI annotation on the image manifest (org.moltbunker.signature).
//      Pros: travels with the image, no extra distribution. Cons: producer
//      must annotate at build time; registries that strip annotations break it.
//   2. Sidecar object pulled from IPFS via image CID. Pros: independent of
//      registry. Cons: needs CID-first deploys (R6).
//   3. On-chain BunkerImageTrust contract mapping digest → signature.
//      Pros: tamper-proof, decentralized. Cons: gas cost per image.
//   4. Image label/env baked into image at build. Pros: simplest. Cons: the
//      thing being signed can sign itself, weakens guarantees.
//
// Any such resolver belongs in a separate `image_sig_source.go` (or in the
// daemon zone if it needs deployment context). This file stays source-agnostic.
