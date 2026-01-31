package distribution

import (
	"context"
	"fmt"
	"io"

	shell "github.com/ipfs/go-ipfs-api"
)

// IPFSClient wraps IPFS API client
type IPFSClient struct {
	client *shell.Shell
}

// NewIPFSClient creates a new IPFS client
func NewIPFSClient(apiAddr string) *IPFSClient {
	return &IPFSClient{
		client: shell.NewShell(apiAddr),
	}
}

// AddImage adds a container image to IPFS
func (ic *IPFSClient) AddImage(ctx context.Context, imageReader io.Reader) (string, error) {
	cid, err := ic.client.Add(imageReader)
	if err != nil {
		return "", fmt.Errorf("failed to add image to IPFS: %w", err)
	}

	return cid, nil
}

// GetImage retrieves a container image from IPFS
func (ic *IPFSClient) GetImage(ctx context.Context, cid string) (io.ReadCloser, error) {
	reader, err := ic.client.Cat(cid)
	if err != nil {
		return nil, fmt.Errorf("failed to get image from IPFS: %w", err)
	}

	return reader, nil
}

// PinImage pins an image in IPFS
func (ic *IPFSClient) PinImage(ctx context.Context, cid string) error {
	if err := ic.client.Pin(cid); err != nil {
		return fmt.Errorf("failed to pin image: %w", err)
	}

	return nil
}

// UnpinImage unpins an image from IPFS
func (ic *IPFSClient) UnpinImage(ctx context.Context, cid string) error {
	if err := ic.client.Unpin(cid); err != nil {
		return fmt.Errorf("failed to unpin image: %w", err)
	}

	return nil
}
