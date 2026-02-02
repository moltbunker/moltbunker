package runtime

import (
	"context"
	"fmt"

	"github.com/containerd/containerd"
)

// PullImage pulls a container image
func (cc *ContainerdClient) PullImage(ctx context.Context, ref string) (containerd.Image, error) {
	ctx = cc.WithNamespace(ctx)

	image, err := cc.client.Pull(ctx, ref, containerd.WithPullUnpack)
	if err != nil {
		return nil, fmt.Errorf("failed to pull image %s: %w", ref, err)
	}

	return image, nil
}

// GetImage gets an existing image
func (cc *ContainerdClient) GetImage(ctx context.Context, ref string) (containerd.Image, error) {
	ctx = cc.WithNamespace(ctx)
	return cc.client.GetImage(ctx, ref)
}
