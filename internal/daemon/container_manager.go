package daemon

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/moltbunker/moltbunker/internal/logging"
	"github.com/moltbunker/moltbunker/internal/p2p"
	"github.com/moltbunker/moltbunker/internal/redundancy"
	"github.com/moltbunker/moltbunker/internal/runtime"
	"github.com/moltbunker/moltbunker/internal/tor"
	"github.com/moltbunker/moltbunker/internal/util"
	"github.com/moltbunker/moltbunker/pkg/types"
)

// ContainerManager coordinates container lifecycle across all subsystems
type ContainerManager struct {
	containerd    *runtime.ContainerdClient
	encryption    *runtime.EncryptionManager
	replicator    *redundancy.Replicator
	healthMonitor *redundancy.HealthMonitor
	consensus     *redundancy.ConsensusManager
	router        *p2p.Router
	geoRouter     *p2p.GeographicRouter
	torService    *tor.TorService
	node          *Node

	deployments map[string]*Deployment
	mu          sync.RWMutex

	// pendingDeployments tracks deployments waiting for replica acknowledgments
	pendingDeployments map[string]*pendingDeployment
	pendingMu          sync.RWMutex

	dataDir          string
	containerdSocket string
}

// NewContainerManager creates a new container manager
func NewContainerManager(ctx context.Context, config ContainerManagerConfig, node *Node) (*ContainerManager, error) {
	// Default containerd socket
	if config.ContainerdSocket == "" {
		config.ContainerdSocket = "/run/containerd/containerd.sock"
	}

	// Initialize containerd client with logs directory
	logsDir := filepath.Join(config.DataDir, "logs")
	containerd, err := runtime.NewContainerdClient(config.ContainerdSocket, "moltbunker", logsDir)
	if err != nil {
		// If containerd is not available, we can still run in P2P-only mode
		containerd = nil
	}

	// Initialize encryption manager
	var encryption *runtime.EncryptionManager
	if config.EnableEncryption && runtime.IsEncryptionAvailable() {
		encryption, err = runtime.NewEncryptionManager(config.DataDir)
		if err != nil {
			return nil, fmt.Errorf("failed to create encryption manager: %w", err)
		}
		// Load existing volumes
		encryption.LoadExistingVolumes()
	}

	// Initialize redundancy components
	replicator := redundancy.NewReplicator()
	healthMonitor := redundancy.NewHealthMonitor()
	consensus := redundancy.NewConsensusManager()

	// Initialize geographic router
	geolocator := p2p.NewGeoLocator()
	geoRouter := p2p.NewGeographicRouter(geolocator)

	// Initialize Tor service (optional)
	var torService *tor.TorService
	if config.TorDataDir != "" {
		torConfig := tor.DefaultTorConfig(config.TorDataDir)
		torService, err = tor.NewTorService(torConfig)
		if err != nil {
			// Tor not available, continue without it
			torService = nil
		}
	}

	cm := &ContainerManager{
		containerd:         containerd,
		encryption:         encryption,
		replicator:         replicator,
		healthMonitor:      healthMonitor,
		consensus:          consensus,
		router:             node.Router(),
		geoRouter:          geoRouter,
		torService:         torService,
		node:               node,
		deployments:        make(map[string]*Deployment),
		pendingDeployments: make(map[string]*pendingDeployment),
		dataDir:            config.DataDir,
		containerdSocket:   config.ContainerdSocket,
	}

	// Set up health probe function if containerd is available
	if containerd != nil {
		healthMonitor.SetProbeFunc(func(ctx context.Context, containerID string) (bool, error) {
			status, err := containerd.GetContainerStatus(ctx, containerID)
			if err != nil {
				return false, err
			}
			return status == types.ContainerStatusRunning, nil
		})
	}

	// Start health monitoring
	util.SafeGoWithName("health-monitor", func() {
		healthMonitor.Start(ctx)
	})

	// Register P2P message handlers for container operations
	cm.registerMessageHandlers()

	// Load existing containers from containerd
	if containerd != nil {
		containerd.LoadExistingContainers(ctx)
	}

	// Load persisted state from disk
	if err := cm.loadState(); err != nil {
		logging.Warn("failed to load persisted state", logging.Err(err))
	}

	return cm, nil
}

// Deploy deploys a new container with 3-copy redundancy
func (cm *ContainerManager) Deploy(ctx context.Context, req *DeployRequest) (*DeployResult, error) {
	cm.mu.Lock()

	// Generate deployment ID
	deploymentID := fmt.Sprintf("dep-%d", time.Now().UnixNano())

	// Set default onion port if not specified
	onionPort := req.OnionPort
	if onionPort == 0 {
		onionPort = 80 // Default to port 80 for HTTP
	}

	// Create deployment record
	deployment := &Deployment{
		ID:           deploymentID,
		Image:        req.Image,
		Status:       types.ContainerStatusPending,
		Resources:    req.Resources,
		CreatedAt:    time.Now(),
		Encrypted:    true,
		OnionService: req.OnionService,
		OnionPort:    onionPort,
		TorOnly:      req.TorOnly,
	}

	// Determine regions for replication
	regions := []string{"Americas", "Europe", "Asia-Pacific"}
	deployment.Regions = regions

	// Create replica set
	replicaSet, err := cm.replicator.CreateReplicaSet(deploymentID, regions)
	if err != nil {
		cm.mu.Unlock()
		return nil, fmt.Errorf("failed to create replica set: %w", err)
	}
	deployment.ReplicaSet = replicaSet

	// If we have containerd available, deploy locally
	if cm.containerd != nil {
		if err := cm.deployLocally(ctx, deploymentID, req, deployment); err != nil {
			cm.mu.Unlock()
			return nil, err
		}
	}

	// Setup onion service if requested
	if req.OnionService && cm.torService != nil {
		if err := cm.setupOnionService(ctx, deploymentID, deployment); err != nil {
			// Log but continue - onion service is optional
			logging.Warn("failed to create onion service",
				logging.ContainerID(deploymentID),
				logging.Err(err))
		}
	}

	// Store deployment
	cm.deployments[deploymentID] = deployment

	// Initialize health monitoring for all replicas
	for i := 0; i < 3; i++ {
		cm.healthMonitor.UpdateHealth(deploymentID, i, types.HealthStatus{
			Healthy:    true,
			LastUpdate: time.Now(),
		})
	}

	// Persist state to disk
	cm.saveStateAsync()

	// Create pending deployment tracker if waiting for replicas
	var pending *pendingDeployment
	if req.WaitForReplicas {
		pending = &pendingDeployment{
			containerID: deploymentID,
			ackChan:     make(chan replicaAck, 10), // Buffer for multiple acks
			created:     time.Now(),
			acks:        make([]replicaAck, 0),
		}
		cm.pendingMu.Lock()
		cm.pendingDeployments[deploymentID] = pending
		cm.pendingMu.Unlock()
	}

	cm.mu.Unlock()

	// Broadcast deployment to network for redundancy
	var broadcastErr error
	broadcastDone := make(chan struct{})
	util.SafeGoWithName("broadcast-deployment", func() {
		defer close(broadcastDone)
		if err := cm.broadcastDeployment(ctx, deployment); err != nil {
			broadcastErr = err
			logging.Warn("replication failed",
				logging.ContainerID(deployment.ID),
				logging.Err(err))
		}
	})

	result := &DeployResult{
		Deployment:   deployment,
		ReplicaCount: 0,
	}

	// If waiting for replicas, wait for acknowledgments
	if req.WaitForReplicas && pending != nil {
		result.ReplicaCount = cm.waitForReplicaAcks(ctx, deploymentID, pending, broadcastDone, broadcastErr)
	}

	return result, nil
}

// deployLocally deploys the container on the local node
func (cm *ContainerManager) deployLocally(ctx context.Context, deploymentID string, req *DeployRequest, deployment *Deployment) error {
	// Set default resources if not specified
	if req.Resources.MemoryLimit == 0 {
		req.Resources.MemoryLimit = 1024 * 1024 * 1024 // 1GB
	}
	if req.Resources.CPUQuota == 0 {
		req.Resources.CPUQuota = 100000
		req.Resources.CPUPeriod = 100000
	}
	if req.Resources.PIDLimit == 0 {
		req.Resources.PIDLimit = 100
	}

	// Setup encrypted volume if encryption is available
	var encryptedVolume *runtime.EncryptedVolume
	var err error
	if cm.encryption != nil {
		diskGB := int(req.Resources.DiskLimit / (1024 * 1024 * 1024))
		if diskGB < 1 {
			diskGB = 10 // Default 10GB
		}
		encryptedVolume, err = cm.encryption.SetupEncryptedVolume(deploymentID, diskGB)
		if err != nil {
			// Log warning and continue without encryption
			logging.Warn("failed to create encrypted volume, continuing without encryption",
				logging.ContainerID(deploymentID),
				"disk_gb", diskGB,
				logging.Err(err))
			encryptedVolume = nil
		} else {
			deployment.EncryptedVolume = encryptedVolume.MountPath
		}
	}

	// Create container
	container, err := cm.containerd.CreateContainer(ctx, deploymentID, req.Image, req.Resources)
	if err != nil {
		// Cleanup encrypted volume on failure
		if cm.encryption != nil && encryptedVolume != nil {
			cm.encryption.DeleteEncryptedVolume(deploymentID)
		}
		return fmt.Errorf("failed to create container: %w", err)
	}

	// Start container
	if err := cm.containerd.StartContainer(ctx, deploymentID); err != nil {
		cm.containerd.DeleteContainer(ctx, deploymentID)
		if cm.encryption != nil && encryptedVolume != nil {
			cm.encryption.DeleteEncryptedVolume(deploymentID)
		}
		return fmt.Errorf("failed to start container: %w", err)
	}

	// Update container with encrypted volume info
	if encryptedVolume != nil {
		container.EncryptedVolume = encryptedVolume.MountPath
	}

	deployment.Status = types.ContainerStatusRunning
	deployment.StartedAt = time.Now()
	deployment.LocalReplica = 0 // This node is replica 0

	return nil
}

// setupOnionService creates an onion service for the deployment
func (cm *ContainerManager) setupOnionService(ctx context.Context, deploymentID string, deployment *Deployment) error {
	onionAddr, err := cm.torService.CreateOnionService(ctx, deployment.OnionPort)
	if err != nil {
		return err
	}
	deployment.OnionAddress = onionAddr
	logging.Info("created onion service",
		logging.ContainerID(deploymentID),
		"onion_address", onionAddr,
		"port", deployment.OnionPort)
	return nil
}

// waitForReplicaAcks waits for replica acknowledgments
func (cm *ContainerManager) waitForReplicaAcks(ctx context.Context, deploymentID string, pending *pendingDeployment, broadcastDone <-chan struct{}, broadcastErr error) int {
	// Wait for broadcast to complete first with timeout
	broadcastTimeout := time.NewTimer(60 * time.Second)
	select {
	case <-broadcastDone:
		broadcastTimeout.Stop()
	case <-broadcastTimeout.C:
		logging.Warn("broadcast timed out after 60 seconds",
			logging.ContainerID(deploymentID))
	case <-ctx.Done():
		broadcastTimeout.Stop()
		logging.Warn("context cancelled while waiting for broadcast",
			logging.ContainerID(deploymentID))
	}

	// If broadcast failed completely, still try to wait for any acks that might come
	if broadcastErr != nil {
		logging.Warn("broadcast had errors, waiting for any replica acks",
			logging.ContainerID(deploymentID),
			logging.Err(broadcastErr))
	}

	// Wait for at least 1 replica ack with timeout
	replicaCount, err := cm.WaitForReplicas(deploymentID, 30*time.Second)
	if err != nil {
		logging.Warn("failed to verify replicas",
			logging.ContainerID(deploymentID),
			logging.Err(err))
	}

	// Cleanup pending deployment tracker - use sync.Once to close channel safely
	cm.pendingMu.Lock()
	if pending, exists := cm.pendingDeployments[deploymentID]; exists {
		pending.close()
		delete(cm.pendingDeployments, deploymentID)
	}
	cm.pendingMu.Unlock()

	return replicaCount
}

// Stop stops a deployed container
func (cm *ContainerManager) Stop(ctx context.Context, containerID string) error {
	cm.mu.Lock()
	deployment, exists := cm.deployments[containerID]
	if !exists {
		cm.mu.Unlock()
		return ErrDeploymentNotFound{ContainerID: containerID}
	}
	cm.mu.Unlock()

	// Stop container if running locally
	if cm.containerd != nil {
		if err := cm.containerd.StopContainer(ctx, containerID, 30*time.Second); err != nil {
			return err
		}
	}

	// Update status
	cm.mu.Lock()
	deployment.Status = types.ContainerStatusStopped
	cm.mu.Unlock()

	// Persist state to disk
	cm.saveStateAsync()

	return nil
}

// Delete deletes a deployed container
func (cm *ContainerManager) Delete(ctx context.Context, containerID string) error {
	cm.mu.Lock()
	deployment, exists := cm.deployments[containerID]
	if !exists {
		cm.mu.Unlock()
		return ErrDeploymentNotFound{ContainerID: containerID}
	}
	delete(cm.deployments, containerID)
	cm.mu.Unlock()

	// Delete container
	if cm.containerd != nil {
		cm.containerd.DeleteContainer(ctx, containerID)
	}

	// Delete encrypted volume
	if cm.encryption != nil && deployment.EncryptedVolume != "" {
		cm.encryption.DeleteEncryptedVolume(containerID)
	}

	// Persist state to disk
	cm.saveStateAsync()

	return nil
}

// GetDeployment returns deployment info
func (cm *ContainerManager) GetDeployment(containerID string) (*Deployment, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	d, exists := cm.deployments[containerID]
	return d, exists
}

// ListDeployments returns all deployments
func (cm *ContainerManager) ListDeployments() []*Deployment {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	result := make([]*Deployment, 0, len(cm.deployments))
	for _, d := range cm.deployments {
		result = append(result, d)
	}
	return result
}

// IsContainerdConnected checks if containerd is available and connected
func (cm *ContainerManager) IsContainerdConnected() bool {
	if cm.containerd == nil {
		return false
	}
	// Try to ping containerd by checking if we can list containers
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := cm.containerd.Client().Containers(cm.containerd.WithNamespace(ctx))
	return err == nil
}

// Close closes the container manager and all managed resources
func (cm *ContainerManager) Close() error {
	// Save state before closing
	if err := cm.saveState(); err != nil {
		logging.Error("failed to save state on close", logging.Err(err))
	}

	// Stop health monitor
	if cm.healthMonitor != nil {
		cm.healthMonitor.Stop()
	}

	// Stop all containers with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	cm.mu.RLock()
	containerIDs := make([]string, 0, len(cm.deployments))
	for containerID := range cm.deployments {
		containerIDs = append(containerIDs, containerID)
	}
	cm.mu.RUnlock()

	for _, containerID := range containerIDs {
		cm.Stop(ctx, containerID)
	}

	// Close containerd
	if cm.containerd != nil {
		cm.containerd.Close()
	}

	// Stop Tor
	if cm.torService != nil {
		cm.torService.Stop()
	}

	return nil
}
