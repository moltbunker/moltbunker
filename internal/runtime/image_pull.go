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

// GetImage gets an existing image
func (cc *ContainerdClient) GetImage(ctx context.Context, ref string) (containerd.Image, error) {
	ctx = cc.WithNamespace(ctx)
	ref = NormalizeImageRef(ref)
	return cc.client.GetImage(ctx, ref)
}
