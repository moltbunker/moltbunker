package runtime

import (
	"context"
	"io"
	"time"

	"github.com/moltbunker/moltbunker/pkg/types"
)

// ContainerRuntime defines the interface for container lifecycle management.
// It abstracts the containerd client so that consumers (ContainerManager, cloning, etc.)
// can be tested with mocks and pointed at different containerd sockets (e.g. Colima).
type ContainerRuntime interface {
	// Container lifecycle
	CreateContainer(ctx context.Context, id string, imageRef string, resources types.ResourceLimits) (*ManagedContainer, error)
	CreateSecureContainer(ctx context.Context, config SecureContainerConfig) (*ManagedContainer, error)
	StartContainer(ctx context.Context, id string) error
	StopContainer(ctx context.Context, id string, timeout time.Duration) error
	DeleteContainer(ctx context.Context, id string) error
	GetContainer(id string) (*ManagedContainer, bool)
	GetContainerStatus(ctx context.Context, id string) (types.ContainerStatus, error)
	LoadExistingContainers(ctx context.Context) error

	// Logs
	GetContainerLogs(ctx context.Context, id string, follow bool, tail int) (io.ReadCloser, error)

	// Exec
	ExecInContainer(ctx context.Context, id string, cmd []string) ([]byte, error)
	ExecInteractive(ctx context.Context, id string, cols, rows uint16) (*InteractiveSession, error)
	CanExec(id string) bool

	// Health
	GetHealthStatus(ctx context.Context, id string) (*types.HealthStatus, error)

	// Connectivity check - returns nil if the runtime is reachable
	Ping(ctx context.Context) error

	// Close releases all resources held by the runtime
	Close() error
}

// Compile-time check that ContainerdClient implements ContainerRuntime.
var _ ContainerRuntime = (*ContainerdClient)(nil)
