package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
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
	containerd      *runtime.ContainerdClient
	encryption      *runtime.EncryptionManager
	replicator      *redundancy.Replicator
	healthMonitor   *redundancy.HealthMonitor
	consensus       *redundancy.ConsensusManager
	router          *p2p.Router
	geoRouter       *p2p.GeographicRouter
	torService      *tor.TorService
	node            *Node

	deployments     map[string]*Deployment
	mu              sync.RWMutex

	// pendingDeployments tracks deployments waiting for replica acknowledgments
	pendingDeployments map[string]*pendingDeployment
	pendingMu          sync.RWMutex

	dataDir         string
	containerdSocket string
}

// pendingDeployment tracks replica acknowledgments for a deployment
type pendingDeployment struct {
	containerID   string
	ackChan       chan replicaAck
	ackCount      int
	successCount  int
	acks          []replicaAck
	mu            sync.Mutex
	created       time.Time
	closed        bool // tracks if ackChan has been closed
}

// replicaAck represents an acknowledgment from a replica node
type replicaAck struct {
	NodeID  string
	Region  string
	Success bool
	Error   string
}

// Deployment represents a deployed container with its metadata
type Deployment struct {
	ID              string                 `json:"id"`
	Image           string                 `json:"image"`
	Status          types.ContainerStatus  `json:"status"`
	Resources       types.ResourceLimits   `json:"resources"`
	CreatedAt       time.Time              `json:"created_at"`
	StartedAt       time.Time              `json:"started_at,omitempty"`
	Encrypted       bool                   `json:"encrypted"`
	EncryptedVolume string                 `json:"encrypted_volume,omitempty"`
	OnionService    bool                   `json:"onion_service"`
	OnionAddress    string                 `json:"onion_address,omitempty"`
	OnionPort       int                    `json:"onion_port,omitempty"` // Port exposed via Tor
	TorOnly         bool                   `json:"tor_only"`
	ReplicaSet      *types.ReplicaSet `json:"replica_set,omitempty"`
	LocalReplica    int                    `json:"local_replica"`
	Regions         []string               `json:"regions"`
}

// ContainerManagerConfig contains configuration for the container manager
type ContainerManagerConfig struct {
	DataDir          string
	ContainerdSocket string
	TorDataDir       string
	EnableEncryption bool
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

// registerMessageHandlers registers handlers for container-related P2P messages
func (cm *ContainerManager) registerMessageHandlers() {
	// Handler for deployment requests from other nodes
	cm.router.RegisterHandler(types.MessageTypeDeploy, cm.handleDeployRequest)

	// Handler for deployment acknowledgments
	cm.router.RegisterHandler(types.MessageTypeDeployAck, cm.handleDeployAck)

	// Handler for container status queries
	cm.router.RegisterHandler(types.MessageTypeContainerStatus, cm.handleStatusRequest)

	// Handler for replica sync
	cm.router.RegisterHandler(types.MessageTypeReplicaSync, cm.handleReplicaSync)
}

// DeployResult contains the result of a deployment including replica count
type DeployResult struct {
	Deployment   *Deployment
	ReplicaCount int
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
			cm.mu.Unlock()
			return nil, fmt.Errorf("failed to create container: %w", err)
		}

		// Start container
		if err := cm.containerd.StartContainer(ctx, deploymentID); err != nil {
			cm.containerd.DeleteContainer(ctx, deploymentID)
			if cm.encryption != nil && encryptedVolume != nil {
				cm.encryption.DeleteEncryptedVolume(deploymentID)
			}
			cm.mu.Unlock()
			return nil, fmt.Errorf("failed to start container: %w", err)
		}

		// Update container with encrypted volume info
		if encryptedVolume != nil {
			container.EncryptedVolume = encryptedVolume.MountPath
		}

		deployment.Status = types.ContainerStatusRunning
		deployment.StartedAt = time.Now()
		deployment.LocalReplica = 0 // This node is replica 0
	}

	// Setup onion service if requested
	if req.OnionService && cm.torService != nil {
		// Create onion service for container on the specified port
		onionAddr, err := cm.torService.CreateOnionService(ctx, deployment.OnionPort)
		if err != nil {
			logging.Warn("failed to create onion service",
				logging.ContainerID(deploymentID),
				logging.Err(err))
		} else {
			deployment.OnionAddress = onionAddr
			logging.Info("created onion service",
				logging.ContainerID(deploymentID),
				"onion_address", onionAddr,
				"port", deployment.OnionPort)
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
		result.ReplicaCount = replicaCount

		// Cleanup pending deployment tracker - close channel first to prevent races
		cm.pendingMu.Lock()
		if pending, exists := cm.pendingDeployments[deploymentID]; exists {
			pending.mu.Lock()
			if !pending.closed {
				close(pending.ackChan)
				pending.closed = true
			}
			pending.mu.Unlock()
			delete(cm.pendingDeployments, deploymentID)
		}
		cm.pendingMu.Unlock()
	}

	return result, nil
}

// broadcastDeployment broadcasts deployment to network for redundancy
func (cm *ContainerManager) broadcastDeployment(ctx context.Context, deployment *Deployment) error {
	// Get peers in target regions
	peers := cm.router.GetPeers()

	if len(peers) == 0 {
		return fmt.Errorf("no peers available for replication")
	}

	// Find nodes in different regions for replication
	selectedNodes, err := cm.geoRouter.SelectNodesForReplication(peers)
	if err != nil {
		// Log warning but continue with available nodes
		logging.Warn("geographic selection failed, using available peers",
			logging.Err(err),
			"peer_count", len(peers))
		selectedNodes = peers
		if len(selectedNodes) > 3 {
			selectedNodes = selectedNodes[:3]
		}
	}

	// Send deployment request to selected nodes
	deployData, err := json.Marshal(deployment)
	if err != nil {
		return fmt.Errorf("failed to marshal deployment: %w", err)
	}

	msg := &types.Message{
		Type:      types.MessageTypeDeploy,
		From:      cm.node.nodeInfo.ID,
		Payload:   deployData,
		Timestamp: time.Now(),
	}

	var successCount int
	var lastErr error
	for _, node := range selectedNodes {
		if node.ID == cm.node.nodeInfo.ID {
			continue // Skip self
		}
		if err := cm.router.SendMessage(ctx, node.ID, msg); err != nil {
			lastErr = err
			logging.Warn("failed to send deployment to peer",
				logging.ContainerID(deployment.ID),
				logging.NodeID(node.ID.String()[:16]),
				logging.Err(err))
			continue
		}
		successCount++
	}

	if successCount == 0 && lastErr != nil {
		return fmt.Errorf("failed to replicate to any peer: %w", lastErr)
	}

	logging.Info("deployment replicated",
		logging.ContainerID(deployment.ID),
		"peer_count", successCount)
	return nil
}

// handleDeployRequest handles incoming deployment requests from other nodes
func (cm *ContainerManager) handleDeployRequest(ctx context.Context, msg *types.Message, from *types.Node) error {
	var deployment Deployment
	if err := json.Unmarshal(msg.Payload, &deployment); err != nil {
		return err
	}

	// Check if we should accept this deployment
	// (based on resources, region requirements, etc.)

	cm.mu.Lock()
	if _, exists := cm.deployments[deployment.ID]; exists {
		cm.mu.Unlock()
		return nil // Already have this deployment
	}

	// Make a copy of the deployment to store (avoid race conditions)
	storedDeployment := deployment
	cm.deployments[deployment.ID] = &storedDeployment

	// Copy regions for use after unlock (avoid race on slice access)
	regions := make([]string, len(deployment.Regions))
	copy(regions, deployment.Regions)

	// Make a copy of the deployment BEFORE releasing the lock for use in goroutine
	deploymentCopy := deployment
	originatorID := msg.From
	cm.mu.Unlock()

	// If we have containerd and this is for our region, deploy locally
	if cm.containerd != nil {
		myRegion := p2p.GetRegionFromCountry(cm.node.nodeInfo.Country)
		for _, region := range regions {
			if region == myRegion {
				// We're a target for this deployment - run async but log errors
				util.SafeGoWithName("deploy-replica", func() {
					if err := cm.deployReplica(ctx, &deploymentCopy, originatorID); err != nil {
						logging.Warn("failed to deploy replica",
							logging.ContainerID(deploymentCopy.ID),
							logging.Err(err))
					}
				})
				break
			}
		}
	}

	return nil
}

// deployReplica deploys a replica of a container locally and sends acknowledgment
func (cm *ContainerManager) deployReplica(ctx context.Context, deployment *Deployment, originatorID types.NodeID) error {
	// Create container (this will pull the image if not present)
	logging.Info("pulling image and creating replica container",
		logging.ContainerID(deployment.ID),
		"image", deployment.Image)
	managed, err := cm.containerd.CreateContainer(ctx, deployment.ID, deployment.Image, deployment.Resources)
	if err != nil {
		logging.Error("failed to create replica container",
			logging.ContainerID(deployment.ID),
			logging.Err(err))
		cm.sendDeployAck(ctx, originatorID, deployment.ID, false, err.Error())
		return fmt.Errorf("failed to create container: %w", err)
	}

	logging.Info("created replica container",
		logging.ContainerID(deployment.ID),
		"image", deployment.Image)
	_ = managed // Use the managed container

	// Start container
	if err := cm.containerd.StartContainer(ctx, deployment.ID); err != nil {
		logging.Error("failed to start replica container",
			logging.ContainerID(deployment.ID),
			logging.Err(err))
		cm.containerd.DeleteContainer(ctx, deployment.ID)
		cm.sendDeployAck(ctx, originatorID, deployment.ID, false, err.Error())
		return fmt.Errorf("failed to start container: %w", err)
	}

	cm.mu.Lock()
	deployment.Status = types.ContainerStatusRunning
	deployment.StartedAt = time.Now()
	cm.mu.Unlock()

	logging.Info("replica container started successfully",
		logging.ContainerID(deployment.ID))

	// Send acknowledgment back to originator
	cm.sendDeployAck(ctx, originatorID, deployment.ID, true, "")

	return nil
}

// sendDeployAck sends a deployment acknowledgment message
func (cm *ContainerManager) sendDeployAck(ctx context.Context, to types.NodeID, containerID string, success bool, errMsg string) {
	ack := map[string]interface{}{
		"container_id": containerID,
		"success":      success,
		"error":        errMsg,
		"node_id":      cm.node.nodeInfo.ID.String(),
		"region":       cm.node.nodeInfo.Region,
	}

	ackData, err := json.Marshal(ack)
	if err != nil {
		return
	}

	msg := &types.Message{
		Type:      types.MessageTypeDeployAck,
		From:      cm.node.nodeInfo.ID,
		To:        to,
		Payload:   ackData,
		Timestamp: time.Now(),
	}

	if err := cm.router.SendMessage(ctx, to, msg); err != nil {
		logging.Warn("failed to send deploy ack",
			logging.NodeID(to.String()[:16]),
			logging.Err(err))
	}
}

// handleStatusRequest handles container status queries
func (cm *ContainerManager) handleStatusRequest(ctx context.Context, msg *types.Message, from *types.Node) error {
	containerID := string(msg.Payload)

	cm.mu.RLock()
	deployment, exists := cm.deployments[containerID]
	cm.mu.RUnlock()

	if !exists {
		return nil
	}

	// Get current status
	status := deployment.Status
	if cm.containerd != nil {
		if s, err := cm.containerd.GetContainerStatus(ctx, containerID); err == nil {
			status = s
		}
	}

	// Send status response
	response := map[string]interface{}{
		"container_id": containerID,
		"status":       status,
		"started_at":   deployment.StartedAt,
	}
	responseData, _ := json.Marshal(response)

	return cm.router.SendMessage(ctx, from.ID, &types.Message{
		Type:      types.MessageTypeContainerStatus,
		From:      cm.node.nodeInfo.ID,
		To:        from.ID,
		Payload:   responseData,
		Timestamp: time.Now(),
	})
}

// handleDeployAck handles deployment acknowledgment from replica nodes
func (cm *ContainerManager) handleDeployAck(ctx context.Context, msg *types.Message, from *types.Node) error {
	var ack struct {
		ContainerID string `json:"container_id"`
		Success     bool   `json:"success"`
		Error       string `json:"error"`
		NodeID      string `json:"node_id"`
		Region      string `json:"region"`
	}

	if err := json.Unmarshal(msg.Payload, &ack); err != nil {
		return err
	}

	// Truncate NodeID for logging (handle short IDs gracefully)
	nodeIDForLog := ack.NodeID
	if len(nodeIDForLog) > 16 {
		nodeIDForLog = nodeIDForLog[:16]
	}

	if ack.Success {
		logging.Info("replica confirmed",
			logging.ContainerID(ack.ContainerID),
			logging.NodeID(nodeIDForLog),
			logging.Region(ack.Region))

		// Update health status for this replica
		cm.healthMonitor.UpdateHealth(ack.ContainerID, 1, types.HealthStatus{
			Healthy:    true,
			LastUpdate: time.Now(),
		})
	} else {
		logging.Error("replica deployment failed",
			logging.ContainerID(ack.ContainerID),
			logging.NodeID(nodeIDForLog),
			"reason", ack.Error)

		// Mark replica as unhealthy
		cm.healthMonitor.UpdateHealth(ack.ContainerID, 1, types.HealthStatus{
			Healthy:    false,
			LastUpdate: time.Now(),
		})
	}

	// Notify pending deployment tracker if exists
	cm.pendingMu.RLock()
	pending, exists := cm.pendingDeployments[ack.ContainerID]
	cm.pendingMu.RUnlock()

	if exists && pending != nil {
		replicaAckData := replicaAck{
			NodeID:  ack.NodeID,
			Region:  ack.Region,
			Success: ack.Success,
			Error:   ack.Error,
		}

		// Update pending deployment state and send to channel under lock
		pending.mu.Lock()
		pending.ackCount++
		if ack.Success {
			pending.successCount++
		}
		pending.acks = append(pending.acks, replicaAckData)

		// Non-blocking send to ack channel only if not closed
		if !pending.closed {
			select {
			case pending.ackChan <- replicaAckData:
			default:
				// Channel full, ack is still recorded in the slice
			}
		}
		pending.mu.Unlock()
	}

	return nil
}

// WaitForReplicas waits for replica acknowledgments with the given timeout.
// Returns the number of successful replica acks received.
func (cm *ContainerManager) WaitForReplicas(containerID string, timeout time.Duration) (int, error) {
	cm.pendingMu.RLock()
	pending, exists := cm.pendingDeployments[containerID]
	cm.pendingMu.RUnlock()

	if !exists || pending == nil {
		return 0, fmt.Errorf("no pending deployment found for container: %s", containerID)
	}

	// Check if we already have successful acks
	pending.mu.Lock()
	if pending.successCount > 0 {
		count := pending.successCount
		pending.mu.Unlock()
		return count, nil
	}
	pending.mu.Unlock()

	// Wait for acks with timeout
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	for {
		select {
		case ack := <-pending.ackChan:
			if ack.Success {
				// Got at least one successful ack
				pending.mu.Lock()
				count := pending.successCount
				pending.mu.Unlock()

				logging.Info("replica verification completed",
					logging.ContainerID(containerID),
					"successful_replicas", count)
				return count, nil
			}
			// Continue waiting for more acks
		case <-timer.C:
			// Timeout reached, return current count
			pending.mu.Lock()
			count := pending.successCount
			totalAcks := pending.ackCount
			pending.mu.Unlock()

			if count == 0 {
				logging.Warn("no replica acknowledgments received within timeout",
					logging.ContainerID(containerID),
					"timeout", timeout.String(),
					"total_acks", totalAcks)
				return 0, fmt.Errorf("timeout waiting for replica acks: received %d acks, %d successful", totalAcks, count)
			}

			return count, nil
		}
	}
}

// GetReplicaStatus returns the current replica status for a deployment
func (cm *ContainerManager) GetReplicaStatus(containerID string) (ackCount int, successCount int, exists bool) {
	cm.pendingMu.RLock()
	pending, exists := cm.pendingDeployments[containerID]
	cm.pendingMu.RUnlock()

	if !exists || pending == nil {
		return 0, 0, false
	}

	pending.mu.Lock()
	defer pending.mu.Unlock()
	return pending.ackCount, pending.successCount, true
}

// handleReplicaSync handles replica synchronization messages
func (cm *ContainerManager) handleReplicaSync(ctx context.Context, msg *types.Message, from *types.Node) error {
	// Sync replica state between nodes
	var syncData struct {
		ContainerID string               `json:"container_id"`
		Status      types.ContainerStatus `json:"status"`
		ReplicaIdx  int                  `json:"replica_idx"`
	}

	if err := json.Unmarshal(msg.Payload, &syncData); err != nil {
		return err
	}

	// Update consensus state
	cm.consensus.UpdateState(syncData.ContainerID, syncData.Status, [3]*types.Container{})

	return nil
}

// Stop stops a deployed container
func (cm *ContainerManager) Stop(ctx context.Context, containerID string) error {
	cm.mu.Lock()
	deployment, exists := cm.deployments[containerID]
	if !exists {
		cm.mu.Unlock()
		return fmt.Errorf("deployment not found: %s", containerID)
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
		return fmt.Errorf("deployment not found: %s", containerID)
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

// GetLogs returns logs for a container
func (cm *ContainerManager) GetLogs(ctx context.Context, containerID string, follow bool, tail int) (io.ReadCloser, error) {
	if cm.containerd == nil {
		return nil, fmt.Errorf("containerd not available")
	}

	return cm.containerd.GetContainerLogs(ctx, containerID, follow, tail)
}

// GetHealth returns health status for a deployment
func (cm *ContainerManager) GetHealth(ctx context.Context, containerID string) (*types.HealthStatus, error) {
	cm.mu.RLock()
	_, exists := cm.deployments[containerID]
	cm.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("deployment not found: %s", containerID)
	}

	// Get health from health monitor
	for i := 0; i < 3; i++ {
		if health, ok := cm.healthMonitor.GetHealth(containerID, i); ok {
			return &health.Health, nil
		}
	}

	// Get from containerd if available
	if cm.containerd != nil {
		return cm.containerd.GetHealthStatus(ctx, containerID)
	}

	return &types.HealthStatus{
		Healthy:    false,
		LastUpdate: time.Now(),
	}, nil
}

// GetUnhealthyDeployments returns deployments with unhealthy replicas
func (cm *ContainerManager) GetUnhealthyDeployments() map[string][]int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	result := make(map[string][]int)
	for containerID := range cm.deployments {
		unhealthy := cm.healthMonitor.GetUnhealthyReplicas(containerID)
		if len(unhealthy) > 0 {
			result[containerID] = unhealthy
		}
	}
	return result
}

// StartTor starts the Tor service
func (cm *ContainerManager) StartTor(ctx context.Context) error {
	if cm.torService == nil {
		return fmt.Errorf("Tor service not configured")
	}
	return cm.torService.Start(ctx)
}

// StopTor stops the Tor service
func (cm *ContainerManager) StopTor() error {
	if cm.torService == nil {
		return nil
	}
	return cm.torService.Stop()
}

// GetTorStatus returns Tor service status
func (cm *ContainerManager) GetTorStatus() (bool, string) {
	if cm.torService == nil {
		return false, ""
	}
	return cm.torService.IsRunning(), cm.torService.GetOnionAddress()
}

// RotateTorCircuit rotates the Tor circuit
func (cm *ContainerManager) RotateTorCircuit(ctx context.Context) error {
	if cm.torService == nil {
		return fmt.Errorf("Tor service not configured")
	}
	return cm.torService.RotateCircuit(ctx)
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

// State persistence

// persistedState represents the state to save to disk
type persistedState struct {
	Deployments map[string]*Deployment `json:"deployments"`
	SavedAt     time.Time              `json:"saved_at"`
	Version     int                    `json:"version"`
}

// stateFilePath returns the path to the state file
func (cm *ContainerManager) stateFilePath() string {
	return filepath.Join(cm.dataDir, "state.json")
}

// saveState saves the current state to disk
func (cm *ContainerManager) saveState() error {
	cm.mu.RLock()
	state := persistedState{
		Deployments: make(map[string]*Deployment, len(cm.deployments)),
		SavedAt:     time.Now(),
		Version:     1,
	}
	for k, v := range cm.deployments {
		// Deep copy to avoid race
		depCopy := *v
		state.Deployments[k] = &depCopy
	}
	cm.mu.RUnlock()

	// Create temp file for atomic write
	tmpPath := cm.stateFilePath() + ".tmp"
	f, err := os.OpenFile(tmpPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to create temp state file: %w", err)
	}

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(state); err != nil {
		f.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("failed to encode state: %w", err)
	}

	if err := f.Sync(); err != nil {
		f.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("failed to sync state file: %w", err)
	}

	if err := f.Close(); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to close state file: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tmpPath, cm.stateFilePath()); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to rename state file: %w", err)
	}

	logging.Debug("state saved",
		"deployments", len(state.Deployments),
		"path", cm.stateFilePath())

	return nil
}

// loadState loads state from disk
func (cm *ContainerManager) loadState() error {
	f, err := os.Open(cm.stateFilePath())
	if err != nil {
		if os.IsNotExist(err) {
			logging.Debug("no state file found, starting fresh")
			return nil
		}
		return fmt.Errorf("failed to open state file: %w", err)
	}
	defer f.Close()

	var state persistedState
	if err := json.NewDecoder(f).Decode(&state); err != nil {
		return fmt.Errorf("failed to decode state: %w", err)
	}

	cm.mu.Lock()
	for id, deployment := range state.Deployments {
		cm.deployments[id] = deployment
		logging.Info("restored deployment from state",
			logging.ContainerID(id),
			"status", string(deployment.Status))
	}
	cm.mu.Unlock()

	logging.Info("state loaded",
		"deployments", len(state.Deployments),
		"saved_at", state.SavedAt)

	return nil
}

// saveStateAsync saves state asynchronously (debounced)
func (cm *ContainerManager) saveStateAsync() {
	util.SafeGoWithName("save-state", func() {
		if err := cm.saveState(); err != nil {
			logging.Error("failed to save state", logging.Err(err))
		}
	})
}
