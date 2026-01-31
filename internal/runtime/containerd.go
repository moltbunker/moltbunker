package runtime

import (
	"context"
	"fmt"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/namespaces"
)

// ContainerdClient wraps containerd client
type ContainerdClient struct {
	client *containerd.Client
}

// NewContainerdClient creates a new containerd client
func NewContainerdClient(socketPath string) (*ContainerdClient, error) {
	client, err := containerd.New(socketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to containerd: %w", err)
	}

	return &ContainerdClient{
		client: client,
	}, nil
}

// Client returns the containerd client
func (cc *ContainerdClient) Client() *containerd.Client {
	return cc.client
}

// WithNamespace returns a context with namespace
func (cc *ContainerdClient) WithNamespace(ctx context.Context, namespace string) context.Context {
	return namespaces.WithNamespace(ctx, namespace)
}

// Close closes the containerd client
func (cc *ContainerdClient) Close() error {
	return cc.client.Close()
}
