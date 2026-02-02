package runtime

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/namespaces"

	"github.com/moltbunker/moltbunker/pkg/types"
)

// ContainerdClient wraps containerd client with full lifecycle management
type ContainerdClient struct {
	client     *containerd.Client
	namespace  string
	containers map[string]*ManagedContainer
	mu         sync.RWMutex
	logManager *LogManager
	logsDir    string
}

// ManagedContainer represents a container managed by the runtime
type ManagedContainer struct {
	ID               string
	Image            string
	Container        containerd.Container
	Task             containerd.Task
	Status           types.ContainerStatus
	CreatedAt        time.Time
	StartedAt        time.Time
	Resources        types.ResourceLimits
	EncryptedVolume  string
	OnionAddress     string
	SecurityEnforcer *SecurityEnforcer // Security policy enforcer for container opacity
	DeploymentID     string            // ID of the deployment this container belongs to
	RequesterPubKey  []byte            // Public key of the requester (for encryption)
	mu               sync.RWMutex
}

// ManagedContainerInfo is a snapshot of container info (safe for concurrent access)
type ManagedContainerInfo struct {
	ID              string
	Image           string
	Status          types.ContainerStatus
	CreatedAt       time.Time
	StartedAt       time.Time
	Resources       types.ResourceLimits
	EncryptedVolume string
	OnionAddress    string
	DeploymentID    string
	ExecDisabled    bool // Security: exec is disabled
	AttachDisabled  bool // Security: attach is disabled
	ShellDisabled   bool // Security: shell is disabled
}

// NewContainerdClient creates a new containerd client
func NewContainerdClient(socketPath string, namespace string, logsDir string) (*ContainerdClient, error) {
	client, err := containerd.New(socketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to containerd: %w", err)
	}

	if namespace == "" {
		namespace = "moltbunker"
	}

	if logsDir == "" {
		logsDir = "/var/log/moltbunker/containers"
	}

	// Create log manager
	logManager, err := NewLogManager(logsDir)
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to create log manager: %w", err)
	}

	return &ContainerdClient{
		client:     client,
		namespace:  namespace,
		containers: make(map[string]*ManagedContainer),
		logManager: logManager,
		logsDir:    logsDir,
	}, nil
}

// Client returns the containerd client
func (cc *ContainerdClient) Client() *containerd.Client {
	return cc.client
}

// WithNamespace returns a context with namespace
func (cc *ContainerdClient) WithNamespace(ctx context.Context) context.Context {
	return namespaces.WithNamespace(ctx, cc.namespace)
}

// GetContainer returns a managed container
func (cc *ContainerdClient) GetContainer(id string) (*ManagedContainer, bool) {
	cc.mu.RLock()
	defer cc.mu.RUnlock()
	managed, exists := cc.containers[id]
	return managed, exists
}

// ListContainers returns all managed containers (returns copies for thread safety)
func (cc *ContainerdClient) ListContainers() []ManagedContainerInfo {
	cc.mu.RLock()
	defer cc.mu.RUnlock()

	result := make([]ManagedContainerInfo, 0, len(cc.containers))
	for _, c := range cc.containers {
		c.mu.RLock()
		info := ManagedContainerInfo{
			ID:              c.ID,
			Image:           c.Image,
			Status:          c.Status,
			CreatedAt:       c.CreatedAt,
			StartedAt:       c.StartedAt,
			Resources:       c.Resources,
			EncryptedVolume: c.EncryptedVolume,
			OnionAddress:    c.OnionAddress,
			DeploymentID:    c.DeploymentID,
		}
		// Add security flags if security enforcer exists
		if c.SecurityEnforcer != nil {
			info.ExecDisabled = !c.SecurityEnforcer.CanExec()
			info.AttachDisabled = !c.SecurityEnforcer.CanAttach()
			info.ShellDisabled = !c.SecurityEnforcer.CanShell()
		}
		c.mu.RUnlock()
		result = append(result, info)
	}
	return result
}

// LoadExistingContainers loads existing containers from containerd
func (cc *ContainerdClient) LoadExistingContainers(ctx context.Context) error {
	ctx = cc.WithNamespace(ctx)

	containers, err := cc.client.Containers(ctx)
	if err != nil {
		return fmt.Errorf("failed to list containers: %w", err)
	}

	cc.mu.Lock()
	defer cc.mu.Unlock()

	for _, container := range containers {
		info, err := container.Info(ctx)
		if err != nil {
			continue
		}

		managed := &ManagedContainer{
			ID:        container.ID(),
			Image:     info.Image,
			Container: container,
			Status:    types.ContainerStatusStopped,
			CreatedAt: info.CreatedAt,
		}

		// Check if task is running
		task, err := container.Task(ctx, nil)
		if err == nil {
			managed.Task = task
			status, err := task.Status(ctx)
			if err == nil && status.Status == containerd.Running {
				managed.Status = types.ContainerStatusRunning
			}
		}

		cc.containers[container.ID()] = managed
	}

	return nil
}

// Close closes the containerd client
func (cc *ContainerdClient) Close() error {
	return cc.client.Close()
}
