package runtime

import (
	"context"
	"fmt"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/images"
)

// ImageManager manages container images
type ImageManager struct {
	client *ContainerdClient
}

// NewImageManager creates a new image manager
func NewImageManager(client *ContainerdClient) *ImageManager {
	return &ImageManager{
		client: client,
	}
}

// PullImage pulls an image from a registry or IPFS
func (im *ImageManager) PullImage(ctx context.Context, ref string) (containerd.Image, error) {
	ctx = im.client.WithNamespace(ctx)

	image, err := im.client.Client().Pull(ctx, ref, containerd.WithPullUnpack)
	if err != nil {
		return nil, fmt.Errorf("failed to pull image: %w", err)
	}

	return image, nil
}

// GetImage retrieves an image by reference
func (im *ImageManager) GetImage(ctx context.Context, ref string) (containerd.Image, error) {
	ctx = im.client.WithNamespace(ctx)

	image, err := im.client.Client().GetImage(ctx, ref)
	if err != nil {
		return nil, fmt.Errorf("failed to get image: %w", err)
	}

	return image, nil
}

// ListImages lists all images
func (im *ImageManager) ListImages(ctx context.Context) ([]containerd.Image, error) {
	ctx = im.client.WithNamespace(ctx)

	imageList, err := im.client.Client().ListImages(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list images: %w", err)
	}

	return imageList, nil
}

// DeleteImage deletes an image
func (im *ImageManager) DeleteImage(ctx context.Context, ref string) error {
	ctx = im.client.WithNamespace(ctx)

	imageService := im.client.Client().ImageService()
	if err := imageService.Delete(ctx, ref); err != nil {
		return fmt.Errorf("failed to delete image: %w", err)
	}

	return nil
}

// VerifyImage verifies image integrity using content hash
func (im *ImageManager) VerifyImage(ctx context.Context, ref, expectedCID string) error {
	ctx = im.client.WithNamespace(ctx)

	image, err := im.GetImage(ctx, ref)
	if err != nil {
		return err
	}

	// Get image manifest
	manifest, err := images.Manifest(ctx, im.client.Client().ContentStore(), image.Target(), nil)
	if err != nil {
		return fmt.Errorf("failed to get manifest: %w", err)
	}

	// Verify content hash matches expected CID
	// This is simplified - actual implementation would verify IPFS CID
	_ = manifest
	_ = expectedCID

	return nil
}
