package runtime

import (
	"context"
	"errors"
	"fmt"

	"github.com/containerd/containerd/content"
	imgencryption "github.com/containerd/imgcrypt/images/encryption"
	encconfig "github.com/containers/ocicrypt/config"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/moltbunker/moltbunker/internal/security"
)

// R5 — Image content encryption at rest (content-store encrypt/decrypt).
//
// ImageCrypter encrypts every layer of an OCI image in a containerd content
// store at rest, and decrypts it back in-process just before unpack. It builds
// on imgcrypt's EncryptImage/DecryptImage (which operate directly on the content
// store via converters — no external ctd-decoder stream processor required) plus
// the moltbunker X25519 keywrapper registered in image_keywrap.go.
//
// Threat model (state honestly): this protects the encrypted layer BLOBS of the
// durable image record at rest, for a STOPPED / cold provider. Two surfaces stay
// plaintext WHILE a container runs and are explicitly out of scope here: (a) the
// unpacked overlayfs rootfs (the kernel must read it — covered by the separate
// LUKS volume layer, internal/runtime/encryption.go); and (b) the original
// plaintext layer blobs, which remain pinned by the running container's
// containerd GC reference until it stops (see image_encrypt_store.go limitation
// #1). So R5 v1 defends a stolen-disk / cold-backup of a NON-running provider's
// image store, not a live running container and not a live host-root attacker.
//
// Recipient model (v1 = SELF-RECIPIENT): each provider encrypts the images it
// pulls to its OWN stable X25519 key. There is NO cross-node key delivery: in the
// current data plane every provider (originator + replicas) independently pulls
// the same image from the registry, so each protects its own at-rest copy with
// zero coordination. Sealing one content key to multiple replicas (so a single
// provider can encrypt-and-distribute) is a documented follow-up gated on
// per-node X25519 advertisement; the keywrapper already supports N recipients.

// ImageCrypter is the at-rest image encryption boundary. Implementations either
// perform real ocicrypt encryption (OcicryptImageCrypter) or pass images
// through untouched (NoopImageCrypter) when encryption is disabled.
type ImageCrypter interface {
	// Encrypt re-encrypts every layer of the image identified by desc, sealing
	// each layer key to every recipient X25519 public key. It returns the
	// descriptor of the encrypted image (a new manifest written to cs).
	Encrypt(ctx context.Context, cs content.Store, desc ocispec.Descriptor, recipients [][]byte) (ocispec.Descriptor, error)

	// Decrypt decrypts every encrypted layer of desc using this node's stable
	// X25519 private key, returning the descriptor of the decrypted image
	// (ready to unpack). If desc has no encrypted layers it is returned as-is.
	Decrypt(ctx context.Context, cs content.Store, desc ocispec.Descriptor, privKey []byte) (ocispec.Descriptor, error)

	// Enabled reports whether this crypter performs real encryption. A daemon
	// with image encryption disabled wires a NoopImageCrypter (Enabled=false)
	// so unencrypted public images deploy unchanged.
	Enabled() bool
}

// allLayers is the imgcrypt LayerFilter that selects every layer for
// encryption/decryption.
func allLayers(_ ocispec.Descriptor) bool { return true }

// OcicryptImageCrypter is the default ocicrypt/imgcrypt-backed ImageCrypter.
type OcicryptImageCrypter struct{}

// NewOcicryptImageCrypter returns the real crypter and ensures the moltbunker
// X25519 keywrapper is registered with ocicrypt.
func NewOcicryptImageCrypter() *OcicryptImageCrypter {
	RegisterX25519KeyWrapper()
	return &OcicryptImageCrypter{}
}

// Enabled implements ImageCrypter.
func (c *OcicryptImageCrypter) Enabled() bool { return true }

// Encrypt implements ImageCrypter.
func (c *OcicryptImageCrypter) Encrypt(ctx context.Context, cs content.Store, desc ocispec.Descriptor, recipients [][]byte) (ocispec.Descriptor, error) {
	if cs == nil {
		return ocispec.Descriptor{}, errors.New("image encrypt: nil content store")
	}
	if len(recipients) == 0 {
		return ocispec.Descriptor{}, errors.New("image encrypt: no recipient keys")
	}
	for i, pub := range recipients {
		if len(pub) != security.X25519KeySize {
			return ocispec.Descriptor{}, fmt.Errorf("image encrypt: recipient %d has invalid X25519 key length %d", i, len(pub))
		}
	}

	cc := &encconfig.CryptoConfig{
		EncryptConfig: &encconfig.EncryptConfig{
			Parameters: map[string][][]byte{
				X25519RecipientsParam: recipients,
			},
			DecryptConfig: encconfig.DecryptConfig{Parameters: map[string][][]byte{}},
		},
	}

	newDesc, _, err := imgencryption.EncryptImage(ctx, cs, desc, cc, allLayers)
	if err != nil {
		return ocispec.Descriptor{}, fmt.Errorf("image encrypt: %w", err)
	}
	return newDesc, nil
}

// Decrypt implements ImageCrypter.
func (c *OcicryptImageCrypter) Decrypt(ctx context.Context, cs content.Store, desc ocispec.Descriptor, privKey []byte) (ocispec.Descriptor, error) {
	if cs == nil {
		return ocispec.Descriptor{}, errors.New("image decrypt: nil content store")
	}
	if len(privKey) != security.X25519KeySize {
		return ocispec.Descriptor{}, fmt.Errorf("image decrypt: invalid X25519 private key length %d", len(privKey))
	}

	cc := &encconfig.CryptoConfig{
		DecryptConfig: &encconfig.DecryptConfig{
			Parameters: map[string][][]byte{
				X25519PrivKeyParam: {privKey},
			},
		},
	}

	newDesc, _, err := imgencryption.DecryptImage(ctx, cs, desc, cc, allLayers)
	if err != nil {
		return ocispec.Descriptor{}, fmt.Errorf("image decrypt: %w", err)
	}
	return newDesc, nil
}

// Compile-time assertion.
var _ ImageCrypter = (*OcicryptImageCrypter)(nil)

// NoopImageCrypter is the opt-out crypter used when image encryption is
// disabled: images pass through untouched. Encrypt with a NoopImageCrypter is a
// programming error (the daemon must not request encryption when disabled) and
// returns an error; Decrypt returns the descriptor unchanged.
type NoopImageCrypter struct{}

// NewNoopImageCrypter returns the pass-through crypter.
func NewNoopImageCrypter() *NoopImageCrypter { return &NoopImageCrypter{} }

// Enabled implements ImageCrypter.
func (c *NoopImageCrypter) Enabled() bool { return false }

// Encrypt implements ImageCrypter.
func (c *NoopImageCrypter) Encrypt(_ context.Context, _ content.Store, desc ocispec.Descriptor, _ [][]byte) (ocispec.Descriptor, error) {
	return desc, errors.New("image encrypt: encryption is disabled (noop crypter)")
}

// Decrypt implements ImageCrypter.
func (c *NoopImageCrypter) Decrypt(_ context.Context, _ content.Store, desc ocispec.Descriptor, _ []byte) (ocispec.Descriptor, error) {
	return desc, nil
}

// Compile-time assertion.
var _ ImageCrypter = (*NoopImageCrypter)(nil)
