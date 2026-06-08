package runtime

import (
	"context"
	"fmt"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/content"
	"github.com/containerd/containerd/errdefs"
	"github.com/containerd/containerd/images"
	"github.com/containerd/containerd/platforms"
	imgencryption "github.com/containerd/imgcrypt/images/encryption"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// R5 — Image content encryption at rest (content-store orchestration).
//
// These methods turn the at-rest content-store representation of an image into
// ocicrypt-encrypted layers (EncryptImageAtRest) and decrypt it back in-process
// just before unpack (decryptImageForRun). They follow imgcrypt's canonical
// ctr-enc pattern: run EncryptImage/DecryptImage over the image's target
// descriptor, then point an image-service record at the new descriptor.
//
// VALIDATION STATUS: build-verified only. Every call below touches a real
// containerd content store / image service; the unit-test MockContainerdClient
// is a no-op, so these paths cannot be exercised by `go test`. They are gated on
// the Linux/Colima runtime CI (R11) and the daemon wires them behind an opt-in
// flag that defaults OFF.
//
// KNOWN v1 LIMITATIONS (documented honestly; tracked as R11 follow-ups — do NOT
// claim full at-rest protection while a container runs):
//
//  1. Plaintext blobs are PINNED while the container runs. The deploy path pulls
//     a plaintext image and creates the container from it, so containerd stamps a
//     `gc.ref.content` label on the container record pointing at the PLAINTEXT
//     manifest. EncryptImageAtRest re-points only the IMAGE-SERVICE record (`ref`)
//     at the encrypted manifest; it does NOT re-point the container's GC ref and
//     does NOT delete the plaintext blobs. So while the container runs, the
//     plaintext layer blobs remain readable on disk. The at-rest ciphertext
//     benefit is only realized for the durable image record once the container is
//     stopped and containerd GC reclaims the (now-unreferenced) plaintext blobs.
//     Fully closing the running-state plaintext window requires either the
//     ctd-decoder stream-processor decrypt-on-unpack path or a lease + explicit
//     content GC after the rootfs snapshot is built — both R11 follow-ups.
//
//  2. The transient decrypted `.moltdec` record (decryptImageForRun) is reaped by
//     the daemon's 24h ImageGC, not immediately after unpack, so a plaintext copy
//     of an already-encrypted image lingers for up to the GC interval. Prompt
//     lease-scoped cleanup is an R11 follow-up.
//
//  3. Recipient model is SELF-RECIPIENT (each provider encrypts to its own key);
//     multi-recipient sealing is a separate follow-up. See security_policy.go.

const (
	// decryptedImageSuffix is appended to an image ref to name the transient,
	// decrypted image record produced for unpack. The encrypted image remains
	// canonical under the original ref.
	decryptedImageSuffix = ".moltdec"

	// decryptedOfLabel records which encrypted image a decrypted record was
	// derived from, so a future R11 cleanup sweep can find and reap transient
	// decrypted records by label.
	decryptedOfLabel = "moltbunker.io/decrypted-of"
)

// imageAlreadyEncrypted reports whether the image's local-platform manifest
// already has ocicrypt-encrypted layers. EncryptImageAtRest uses it to skip a
// redundant re-encrypt of an already-encrypted ref: without this guard, every
// redeploy would re-invoke ocicrypt over the encrypted layers, which (per
// ocicrypt's "already encrypted" path) leaves the layer ciphertext untouched but
// APPENDS a fresh wrapped-key envelope to each layer annotation and rewrites the
// manifest — unbounded annotation growth + manifest churn across redeploys.
// Best-effort: on any read error it returns false (encrypt runs).
func imageAlreadyEncrypted(ctx context.Context, cs content.Store, target ocispec.Descriptor) bool {
	manifest, err := images.Manifest(ctx, cs, target, platforms.Default())
	if err != nil {
		return false
	}
	return imgencryption.HasEncryptedLayer(ctx, manifest.Layers)
}

// EncryptImageAtRest re-encrypts the local image identified by ref so its layer
// blobs are stored encrypted at rest, sealing each layer key to every recipient
// X25519 public key. The image record under ref is updated to point at the
// encrypted descriptor.
//
// IMPORTANT (see file header limitation #1): this re-points only the image
// record, not the running container's GC reference, so plaintext blobs pinned by
// a live container are NOT reclaimed until the container stops. It is a safe
// no-op when the crypter is disabled, no recipients are given, the image is
// already encrypted, or the image has no encryptable layers.
func (cc *ContainerdClient) EncryptImageAtRest(ctx context.Context, ref string, crypter ImageCrypter, recipients [][]byte) error {
	if crypter == nil || !crypter.Enabled() || len(recipients) == 0 {
		return nil
	}
	ctx = cc.WithNamespace(ctx)
	ref = NormalizeImageRef(ref)

	is := cc.client.ImageService()
	img, err := is.Get(ctx, ref)
	if err != nil {
		return fmt.Errorf("encrypt-at-rest: get image %s: %w", ref, err)
	}

	cs := cc.client.ContentStore()

	// Idempotency: never re-encrypt an already-encrypted image (avoids manifest
	// churn + unbounded wrapped-key annotation growth on every redeploy).
	if imageAlreadyEncrypted(ctx, cs, img.Target) {
		return nil
	}

	encDesc, err := crypter.Encrypt(ctx, cs, img.Target, recipients)
	if err != nil {
		return fmt.Errorf("encrypt-at-rest: encrypt %s: %w", ref, err)
	}
	if encDesc.Digest == img.Target.Digest {
		// Nothing was encrypted (e.g. no layers matched). Leave the record as-is.
		return nil
	}

	img.Target = encDesc
	if _, err := is.Update(ctx, img, "target"); err != nil {
		return fmt.Errorf("encrypt-at-rest: update %s to encrypted target: %w", ref, err)
	}
	return nil
}

// decryptImageForRun returns a runnable, unpacked image for the (possibly
// encrypted) image. If the image's layers are not encrypted it returns the
// input image unchanged. Otherwise it decrypts the layers in-process, records a
// transient decrypted image under ref+decryptedImageSuffix, unpacks it, and
// returns that image so the caller can create the container snapshot from it.
//
// privKey is this node's stable X25519 private key. Fail-closed: any decrypt or
// content-store error aborts with an error and never returns a partial image.
// See file header limitation #2: the transient decrypted record is reaped by the
// 24h ImageGC, not immediately.
func (cc *ContainerdClient) decryptImageForRun(ctx context.Context, image containerd.Image, ref string, crypter ImageCrypter, privKey []byte) (containerd.Image, error) {
	if crypter == nil || !crypter.Enabled() || len(privKey) == 0 {
		return image, nil
	}
	// Defense-in-depth: namespace the context like EncryptImageAtRest does. The
	// sole caller (CreateSecureContainer) already namespaces, and WithNamespace
	// is idempotent, so this is symmetry + future-proofing against a refactor.
	ctx = cc.WithNamespace(ctx)

	orig := image.Target()
	cs := cc.client.ContentStore()
	decDesc, err := crypter.Decrypt(ctx, cs, orig, privKey)
	if err != nil {
		return nil, fmt.Errorf("decrypt-at-rest: decrypt %s: %w", ref, err)
	}
	if decDesc.Digest == orig.Digest {
		// No encrypted layers — image is already plaintext, pass through.
		return image, nil
	}

	decName := NormalizeImageRef(ref) + decryptedImageSuffix
	is := cc.client.ImageService()
	rec := images.Image{
		Name:   decName,
		Target: decDesc,
		Labels: map[string]string{decryptedOfLabel: NormalizeImageRef(ref)},
	}
	if _, err := is.Create(ctx, rec); err != nil {
		if !errdefs.IsAlreadyExists(err) {
			return nil, fmt.Errorf("decrypt-at-rest: create decrypted record %s: %w", decName, err)
		}
		if _, err := is.Update(ctx, rec, "target"); err != nil {
			return nil, fmt.Errorf("decrypt-at-rest: update decrypted record %s: %w", decName, err)
		}
	}

	decImage, err := cc.client.GetImage(ctx, decName)
	if err != nil {
		return nil, fmt.Errorf("decrypt-at-rest: get decrypted image %s: %w", decName, err)
	}
	// Unpack with the default snapshotter so the caller's WithNewSnapshot can
	// build the container rootfs from the decrypted layers.
	if err := decImage.Unpack(ctx, ""); err != nil {
		return nil, fmt.Errorf("decrypt-at-rest: unpack decrypted image %s: %w", decName, err)
	}
	return decImage, nil
}
