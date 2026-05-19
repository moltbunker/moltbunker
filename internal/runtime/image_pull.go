package runtime

import (
	"context"
	"fmt"
	"strings"

	"github.com/containerd/containerd"
)

// NormalizeImageRef converts short Docker Hub references to fully-qualified
// containerd references. Unlike Docker, containerd does not auto-normalize
// short names, so "nginx:alpine" must become "docker.io/library/nginx:alpine".
func NormalizeImageRef(ref string) string {
	// Already contains a registry host (has a dot before the first slash)
	if i := strings.IndexByte(ref, '/'); i > 0 {
		host := ref[:i]
		if strings.ContainsAny(host, ".:") {
			return ref
		}
		// User image on Docker Hub: "ollama/ollama:latest" → "docker.io/ollama/ollama:latest"
		return "docker.io/" + ref
	}
	// Official image: "nginx:alpine" or "nginx" → "docker.io/library/nginx:alpine"
	return "docker.io/library/" + ref
}

// PullImage pulls a container image
func (cc *ContainerdClient) PullImage(ctx context.Context, ref string) (containerd.Image, error) {
	ctx = cc.WithNamespace(ctx)
	ref = NormalizeImageRef(ref)

	image, err := cc.client.Pull(ctx, ref, containerd.WithPullUnpack)
	if err != nil {
		return nil, fmt.Errorf("failed to pull image %s: %w", ref, err)
	}

	return image, nil
}

// PullImageVerified pulls a container image and verifies its signature against
// the supplied TrustPolicy before returning. If verification fails the image is
// deleted from the local content store and the verifier's error is returned —
// callers must treat the image as untrusted and avoid creating containers from
// it.
//
// A nil signature is allowed iff policy.RequireSignature is false. The caller
// (typically the daemon-zone deployment orchestrator) is responsible for
// sourcing the ImageSignature; see image_verify.go's TODO(R3-sourcing) for the
// shortlist of approaches.
func (cc *ContainerdClient) PullImageVerified(
	ctx context.Context,
	ref string,
	sig *ImageSignature,
	policy TrustPolicy,
	verifier ImageVerifier,
) (containerd.Image, error) {
	if verifier == nil {
		verifier = NewEdImageVerifier()
	}

	image, err := cc.PullImage(ctx, ref)
	if err != nil {
		return nil, err
	}

	digest := ImageDigest(image.Target().Digest.String())
	if verifyErr := verifier.Verify(digest, sig, policy); verifyErr != nil {
		// Delete the now-untrusted image from the content store so it cannot
		// be reused by a later non-verifying caller. Best-effort: if the
		// delete fails we still return the verification error.
		nsCtx := cc.WithNamespace(ctx)
		_ = cc.client.ImageService().Delete(nsCtx, ref)
		return nil, fmt.Errorf("image %s (digest %s) failed signature verification: %w", ref, digest, verifyErr)
	}

	return image, nil
}

// GetImage gets an existing image
func (cc *ContainerdClient) GetImage(ctx context.Context, ref string) (containerd.Image, error) {
	ctx = cc.WithNamespace(ctx)
	ref = NormalizeImageRef(ref)
	return cc.client.GetImage(ctx, ref)
}
