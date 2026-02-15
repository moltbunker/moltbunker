package daemon

import (
	"context"
	cryptorand "crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/moltbunker/moltbunker/internal/logging"
	"github.com/moltbunker/moltbunker/internal/p2p"
	"github.com/moltbunker/moltbunker/internal/payment"
	"github.com/moltbunker/moltbunker/internal/redundancy"
	"github.com/moltbunker/moltbunker/internal/runtime"
	"github.com/moltbunker/moltbunker/internal/tor"
	"github.com/moltbunker/moltbunker/internal/util"
	"github.com/moltbunker/moltbunker/pkg/types"
)

// ContainerManager coordinates container lifecycle across all subsystems
type ContainerManager struct {
	containerd    runtime.ContainerRuntime
	encryption    *runtime.EncryptionManager
	replicator    *redundancy.Replicator
	healthMonitor *redundancy.HealthMonitor
	consensus     *redundancy.ConsensusManager
	router        *p2p.Router
	geoRouter     *p2p.GeographicRouter
	gossip        *p2p.GossipProtocol
	torService    *tor.TorService
	node          *Node
	payment       *payment.PaymentService

	deployments map[string]*Deployment
	mu          sync.RWMutex

	// pendingDeployments tracks deployments waiting for replica acknowledgments
	pendingDeployments map[string]*pendingDeployment
	pendingMu          sync.RWMutex

	// execStreams tracks active exec sessions on this provider node
	execStreams *ExecStreamManager

	// execRelays tracks requester-side exec relays (P2P → WebSocket forwarding)
	execRelays   map[string]*ExecRelay
	execRelaysMu sync.RWMutex

	dataDir          string
	containerdSocket string
	runtimeName      string          // resolved OCI runtime name for reconnection
	kataConfig       *runtime.KataConfig // saved for reconnection path
	imageGC          *runtime.ImageGC    // image garbage collector (nil if no containerd)
	cleanupMgr       *runtime.CleanupManager // P1-1: orphan container cleanup
	healthChecker    *runtime.HealthChecker  // P1-2: container-level health probes

	// P1-10: Container lifecycle event counters (atomic, lock-free)
	deploysTotal  atomic.Int64
	stopsTotal    atomic.Int64
	deletesTotal  atomic.Int64
	failuresTotal atomic.Int64
}

// NewContainerManager creates a new container manager
func NewContainerManager(ctx context.Context, config ContainerManagerConfig, node *Node) (*ContainerManager, error) {
	// Default containerd socket
	if config.ContainerdSocket == "" {
		config.ContainerdSocket = "/run/containerd/containerd.sock"
	}

	// Detect the best OCI runtime for this node
	rtCaps := runtime.DetectRuntime(config.RuntimeName)

	// Initialize containerd client with logs directory and detected runtime
	logsDir := filepath.Join(config.DataDir, "logs")
	var crt runtime.ContainerRuntime
	if cc, err := runtime.NewContainerdClient(config.ContainerdSocket, "moltbunker", logsDir, rtCaps.RuntimeName, config.KataConfig); err == nil {
		crt = cc
	}
	// If containerd is not available, crt stays nil and we run in P2P-only mode

	// Initialize encryption manager
	var encryption *runtime.EncryptionManager
	if config.EnableEncryption && runtime.IsEncryptionAvailable() {
		var encErr error
		encryption, encErr = runtime.NewEncryptionManager(config.DataDir)
		if encErr != nil {
			return nil, fmt.Errorf("failed to create encryption manager: %w", encErr)
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
		var torErr error
		torService, torErr = tor.NewTorService(torConfig)
		if torErr != nil {
			// Tor not available, continue without it
			torService = nil
		}
	}

	// Initialize gossip protocol for state synchronization
	gossipProto := p2p.NewGossipProtocol(node.Router())

	cm := &ContainerManager{
		containerd:         crt,
		encryption:         encryption,
		replicator:         replicator,
		healthMonitor:      healthMonitor,
		consensus:          consensus,
		router:             node.Router(),
		geoRouter:          geoRouter,
		gossip:             gossipProto,
		torService:         torService,
		node:               node,
		payment:            config.PaymentService,
		deployments:        make(map[string]*Deployment),
		pendingDeployments: make(map[string]*pendingDeployment),
		execStreams:        NewExecStreamManager(),
		execRelays:         make(map[string]*ExecRelay),
		dataDir:            config.DataDir,
		containerdSocket:   config.ContainerdSocket,
		runtimeName:        rtCaps.RuntimeName,
		kataConfig:         config.KataConfig,
	}

	// Set up health probe function if containerd is available
	if crt != nil {
		healthMonitor.SetProbeFunc(func(ctx context.Context, containerID string) (bool, error) {
			status, err := crt.GetContainerStatus(ctx, containerID)
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

	// Start gossip protocol for state synchronization
	util.SafeGoWithName("gossip-protocol", func() {
		gossipProto.Start(ctx)
	})

	// Register P2P message handlers for container operations
	cm.registerMessageHandlers()
	cm.registerExecHandlers()

	// Load existing containers from containerd
	if crt != nil {
		crt.LoadExistingContainers(ctx)
	}

	// Load persisted state from disk
	if err := cm.loadState(); err != nil {
		logging.Warn("failed to load persisted state", logging.Err(err))
	}

	// Immediate reconciliation: sync persisted state with actual containerd status.
	// Containers survive daemon restarts since containerd is a separate service.
	if crt != nil {
		cm.reconcileOnStartup(ctx)
	}

	// Periodic cleanup of stale pending deployments (async deploys)
	util.SafeGoWithName("pending-deployment-cleanup", func() {
		cm.cleanStalePendingDeployments(ctx)
	})

	// Reconcile deployment status with containerd reality (detect crashes, OOM kills)
	if crt != nil {
		util.SafeGoWithName("status-reconciliation", func() {
			cm.reconcileContainerStatus(ctx)
		})
	}

	// Periodically clean up stopped containers with expired volume retention
	util.SafeGoWithName("volume-retention-cleanup", func() {
		cm.cleanupExpiredVolumes(ctx)
	})

	// P0-2: Start periodic attestation goroutine (providers submit hardware attestation every 24h)
	if config.PaymentService != nil {
		util.SafeGoWithName("attestation-submitter", func() {
			cm.runAttestationLoop(ctx)
		})
	}

	// P0-3: Wire health failure callback to slashing — when a replica goes unhealthy,
	// report the violation on-chain via the slashing contract
	if config.PaymentService != nil {
		util.SafeGoWithName("health-failure-reporter", func() {
			cm.monitorHealthFailures(ctx)
		})
	}

	// P0-11: Start image garbage collection (removes unused images every hour)
	if cc, ok := crt.(*runtime.ContainerdClient); ok && cc != nil {
		imgMgr := runtime.NewImageManager(cc)
		imageGC := runtime.NewImageGC(imgMgr, 24*time.Hour, 0) // 24h max age, no size limit
		imageGC.Start(ctx)
		cm.imageGC = imageGC

		// P1-1: Initialize cleanup manager for orphan container cleanup
		cleanupMgr := runtime.NewCleanupManager(cc, nil, config.DataDir)
		cm.cleanupMgr = cleanupMgr

		// Run orphan cleanup on startup (handles containers left behind by crash)
		if cleaned, err := cleanupMgr.CleanupOrphaned(ctx); err != nil {
			logging.Warn("orphan cleanup failed on startup", logging.Err(err))
		} else if len(cleaned) > 0 {
			logging.Info("cleaned orphaned containers on startup",
				"count", len(cleaned))
		}
	}

	// P2-4: Periodic cleanup of stale escrow reservation mappings (abandoned jobs)
	if cm.payment != nil {
		util.SafeGoWithName("reservation-cleanup", func() {
			ticker := time.NewTicker(1 * time.Hour)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					if n := cm.payment.CleanupStaleReservations(24 * time.Hour); n > 0 {
						logging.Info("cleaned stale reservation mappings", "count", n)
					}
				}
			}
		})
	}

	// P1-2: Initialize container-level health checker
	healthChecker := runtime.NewHealthChecker()
	if crt != nil {
		healthChecker.SetExecFunc(func(execCtx context.Context, containerID string, cmd []string) (int, error) {
			// Adapt ContainerRuntime.ExecInContainer ([]byte, error) to ExecFunc (int, error)
			_, err := crt.ExecInContainer(execCtx, containerID, cmd)
			if err != nil {
				return 1, err
			}
			return 0, nil
		})
	}
	cm.healthChecker = healthChecker
	util.SafeGoWithName("container-health-checker", func() {
		healthChecker.Start(ctx)
	})

	return cm, nil
}

// generateDeploymentID creates a unique deployment ID using crypto/rand.
func generateDeploymentID() string {
	var b [8]byte
	if _, err := cryptorand.Read(b[:]); err != nil {
		// Fallback to UnixNano if crypto/rand fails (shouldn't happen)
		return fmt.Sprintf("dep-%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("dep-%s", hex.EncodeToString(b[:]))
}

// parseDuration parses the duration from a deploy request, defaulting to 720h (30 days)
func parseDuration(d string) time.Duration {
	if d == "" {
		return 720 * time.Hour // 30 days default
	}
	parsed, err := time.ParseDuration(d)
	if err != nil || parsed <= 0 {
		return 720 * time.Hour
	}
	return parsed
}

// Deploy deploys a new container with 3-copy redundancy
func (cm *ContainerManager) Deploy(ctx context.Context, req *DeployRequest) (*DeployResult, error) {
	// Generate unique deployment ID using crypto/rand (collision-safe)
	deploymentID := generateDeploymentID()

	// Validate resource limits before creating escrow
	if err := validateResourceLimits(req.Resources); err != nil {
		return nil, fmt.Errorf("resource validation failed: %w", err)
	}

	// Gate on escrow: either use user-provided reservation or create one
	duration := parseDuration(req.Duration)
	if cm.payment != nil {
		jobID := payment.JobIDFromString(deploymentID)

		if req.ReservationID != "" {
			// User already created the escrow on-chain from their wallet.
			// Just register the mapping so SelectProviders can reference it.
			resID := new(big.Int)
			if _, ok := resID.SetString(req.ReservationID, 10); !ok {
				return nil, fmt.Errorf("invalid reservation_id: %s", req.ReservationID)
			}
			cm.payment.RegisterExternalReservation(jobID, resID)
			logging.Info("registered user-created escrow for deployment",
				logging.ContainerID(deploymentID),
				"reservation_id", req.ReservationID)
		} else {
			// Daemon creates escrow (legacy/CLI flow)
			price := cm.payment.CalculateJobPrice(req.Resources, duration)
			requesterAddr := cm.node.WalletAddress()

			// Pre-flight: verify requester has sufficient token balance
			if requesterAddr != (common.Address{}) {
				balance, err := cm.payment.GetTokenBalance(ctx, requesterAddr)
				if err != nil {
					logging.Warn("failed to check token balance, proceeding with escrow",
						logging.ContainerID(deploymentID),
						logging.Err(err))
				} else if balance != nil && balance.Cmp(price) < 0 {
					return nil, fmt.Errorf("insufficient BUNKER balance: have %s, need %s",
						payment.FormatTokenAmount(balance), payment.FormatTokenAmount(price))
				}
			}

			if err := cm.payment.CreateJobEscrow(ctx, jobID, requesterAddr, price, duration); err != nil {
				return nil, fmt.Errorf("failed to create escrow: %w", err)
			}

			logging.Info("escrow created for deployment",
				logging.ContainerID(deploymentID),
				"price", payment.FormatTokenAmount(price),
				"duration", duration.String())
		}
	}

	cm.mu.Lock()

	// Set default onion port if not specified
	onionPort := req.OnionPort
	if onionPort == 0 {
		onionPort = 80 // Default to port 80 for HTTP
	}

	// Create deployment record
	deployment := &Deployment{
		ID:              deploymentID,
		Image:           req.Image,
		Status:          types.ContainerStatusPending,
		Resources:       req.Resources,
		CreatedAt:       time.Now(),
		Encrypted:       true,
		OnionService:    req.OnionService,
		OnionPort:       onionPort,
		TorOnly:         req.TorOnly,
		OriginatorID:    cm.node.nodeInfo.ID, // Local node is the originator
		Owner:           req.Owner,
		MinProviderTier: types.ProviderTier(req.MinProviderTier),
	}

	// Determine regions from actual network topology
	localRegion := cm.node.nodeInfo.Region
	if localRegion == "" {
		localRegion = "Unknown"
	}
	regions := cm.determineDeploymentRegions(localRegion)
	deployment.Regions = regions

	// Set detailed location for the local replica
	loc := cm.node.nodeInfo.Location
	deployment.Locations = []ReplicaLocation{{
		Region:      loc.Region,
		Country:     loc.Country,
		CountryName: loc.CountryName,
		City:        loc.City,
	}}

	// Create replica set
	replicaSet, err := cm.replicator.CreateReplicaSet(deploymentID, regions)
	if err != nil {
		cm.mu.Unlock()
		cm.refundEscrow(ctx, deploymentID)
		return nil, fmt.Errorf("failed to create replica set: %w", err)
	}
	deployment.ReplicaSet = replicaSet

	// If we have containerd available, deploy locally
	if cm.containerd != nil {
		if err := cm.deployLocally(ctx, deploymentID, req, deployment); err != nil {
			cm.mu.Unlock()
			cm.refundEscrow(ctx, deploymentID)
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
	for i := 0; i < len(regions); i++ {
		cm.healthMonitor.UpdateHealth(deploymentID, i, types.HealthStatus{
			Healthy:    true,
			LastUpdate: time.Now(),
		})
	}

	// Persist state to disk
	cm.saveStateAsync()

	// Always create pending deployment tracker for escrow activation on acks.
	// WaitForReplicas only controls whether Deploy blocks waiting for acks.
	pending := &pendingDeployment{
		containerID: deploymentID,
		ackChan:     make(chan replicaAck, 10), // Buffer for multiple acks
		created:     time.Now(),
		acks:        make([]replicaAck, 0),
	}
	cm.pendingMu.Lock()
	cm.pendingDeployments[deploymentID] = pending
	cm.pendingMu.Unlock()

	cm.mu.Unlock()

	// Broadcast deployment to network for redundancy
	var broadcastErr error
	broadcastDone := make(chan struct{})
	util.SafeGoWithName("broadcast-deployment", func() {
		defer close(broadcastDone)
		if err := cm.broadcastDeployment(ctx, deployment); err != nil {
			broadcastErr = err
			logging.Warn("replication failed, container running locally only",
				logging.ContainerID(deployment.ID),
				logging.Err(err))
			// Don't refund — the local deployment succeeded.
			// On a single-node network this is expected behavior.
			// Escrow activation (selectProviders) will be skipped since
			// no acks will arrive, but the container is live locally.
		}
	})

	result := &DeployResult{
		Deployment:   deployment,
		ReplicaCount: 1, // Local node always counts as 1 replica
	}

	// If waiting for replicas, wait for acknowledgments synchronously
	if req.WaitForReplicas {
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

	// Create container with security hardening
	secConfig := runtime.SecureContainerConfig{
		ID:              deploymentID,
		ImageRef:        req.Image,
		Resources:       req.Resources,
		SecurityProfile: types.DeploymentSecurityProfile(),
	}

	// Inject exec-agent if encrypted exec key is provided (E2E encrypted exec)
	if len(req.EncryptedExecKey) > 0 {
		mounts, keyPath, err := cm.prepareExecAgent(deploymentID, req.EncryptedExecKey)
		if err != nil {
			logging.Warn("failed to prepare exec-agent, continuing without E2E exec",
				logging.ContainerID(deploymentID),
				logging.Err(err))
		} else {
			secConfig.BindMounts = mounts
			deployment.ExecAgentEnabled = true
			deployment.ExecKeyPath = keyPath
		}
	}

	container, err := cm.containerd.CreateSecureContainer(ctx, secConfig)
	if err != nil {
		// Cleanup encrypted volume on failure
		if cm.encryption != nil && encryptedVolume != nil {
			cm.encryption.DeleteEncryptedVolume(deploymentID)
		}
		cm.cleanupExecKey(deployment)
		cm.recordJobFailed(ctx, deploymentID)
		cm.failuresTotal.Add(1) // P1-10
		return fmt.Errorf("failed to create container: %w", err)
	}

	// Start container
	if err := cm.containerd.StartContainer(ctx, deploymentID); err != nil {
		cm.containerd.DeleteContainer(ctx, deploymentID)
		if cm.encryption != nil && encryptedVolume != nil {
			cm.encryption.DeleteEncryptedVolume(deploymentID)
		}
		cm.cleanupExecKey(deployment)
		cm.recordJobFailed(ctx, deploymentID)
		cm.failuresTotal.Add(1) // P1-10
		return fmt.Errorf("failed to start container: %w", err)
	}

	// Update container with encrypted volume info
	if encryptedVolume != nil {
		container.EncryptedVolume = encryptedVolume.MountPath
	}

	deployment.Status = types.ContainerStatusRunning
	deployment.StartedAt = time.Now()
	deployment.LocalReplica = 0 // This node is replica 0

	// Track image as in-use for GC
	cm.markImageInUse(req.Image)

	// P1-2: Register a default exec health probe for this container
	if cm.healthChecker != nil {
		cm.healthChecker.RegisterProbe(deploymentID, runtime.HealthProbeConfig{
			Type:        runtime.ProbeExec,
			ExecCommand: []string{"/bin/true"},
			Interval:    30 * time.Second,
			Timeout:     5 * time.Second,
		})
	}

	// P1-10: Track deploy success
	cm.deploysTotal.Add(1)

	return nil
}

// prepareExecAgent writes the exec key to a temp file and returns bind-mounts
// for the exec-agent binary and key file. The exec-agent binary path is
// resolved from the daemon's own binary directory.
func (cm *ContainerManager) prepareExecAgent(deploymentID string, execKey []byte) ([]runtime.BindMount, string, error) {
	// Write exec_key to a secure temp file
	dataDir := cm.dataDir
	if dataDir == "" {
		dataDir = os.TempDir()
	}
	execKeyDir := filepath.Join(dataDir, "exec-keys")
	if err := os.MkdirAll(execKeyDir, 0700); err != nil {
		return nil, "", fmt.Errorf("create exec-key dir: %w", err)
	}

	keyPath := filepath.Join(execKeyDir, deploymentID+".key")
	if err := os.WriteFile(keyPath, execKey, 0600); err != nil {
		return nil, "", fmt.Errorf("write exec_key: %w", err)
	}

	// Resolve exec-agent binary path (same dir as the daemon binary)
	execAgentPath := cm.resolveExecAgentPath()
	if execAgentPath == "" {
		os.Remove(keyPath)
		return nil, "", fmt.Errorf("exec-agent binary not found")
	}

	mounts := []runtime.BindMount{
		{
			HostPath:      execAgentPath,
			ContainerPath: "/usr/local/bin/exec-agent",
			ReadOnly:      true,
		},
		{
			HostPath:      keyPath,
			ContainerPath: "/run/secrets/exec_key",
			ReadOnly:      true,
		},
	}

	logging.Info("prepared exec-agent bind-mounts",
		logging.ContainerID(deploymentID),
		"agent_path", execAgentPath,
		logging.Component("exec"))

	return mounts, keyPath, nil
}

// resolveExecAgentPath finds the exec-agent binary.
// Checks: same directory as daemon binary, then /usr/local/bin.
func (cm *ContainerManager) resolveExecAgentPath() string {
	// Check alongside the daemon binary
	if exe, err := os.Executable(); err == nil {
		candidate := filepath.Join(filepath.Dir(exe), "exec-agent-amd64")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
		candidate = filepath.Join(filepath.Dir(exe), "exec-agent")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	// Fallback: /usr/local/bin
	for _, name := range []string{"exec-agent-amd64", "exec-agent"} {
		candidate := filepath.Join("/usr/local/bin", name)
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	return ""
}

// cleanupExecKey removes the exec key file from disk.
func (cm *ContainerManager) cleanupExecKey(deployment *Deployment) {
	if deployment.ExecKeyPath != "" {
		os.Remove(deployment.ExecKeyPath)
		deployment.ExecKeyPath = ""
	}
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

	// Release proportional payment based on actual uptime.
	// Only the originator node owns the escrow — replicas must not call payment ops.
	isOriginator := deployment.OriginatorID == cm.node.nodeInfo.ID
	if cm.payment != nil && isOriginator && !deployment.StartedAt.IsZero() {
		jobID := payment.JobIDFromString(containerID)
		uptime := time.Since(deployment.StartedAt)
		if err := cm.payment.ReleaseJobPayment(ctx, jobID, uptime); err != nil {
			logging.Warn("failed to release payment on stop",
				logging.ContainerID(containerID),
				logging.Err(err))
		}
	}

	// Close any active exec sessions for this container
	if cm.execStreams != nil {
		cm.execStreams.CloseAllForContainer(containerID)
	}

	// P1-2: Unregister health probe
	if cm.healthChecker != nil {
		cm.healthChecker.UnregisterProbe(containerID)
	}

	// Clean up pending deployment tracker (no more acks needed)
	cm.pendingMu.Lock()
	if pending, exists := cm.pendingDeployments[containerID]; exists {
		pending.close()
		delete(cm.pendingDeployments, containerID)
	}
	cm.pendingMu.Unlock()

	// Record successful job completion in reputation contract.
	// Only the originator tracks reputation to avoid duplicate reports.
	if cm.payment != nil && isOriginator {
		providerAddr := cm.node.WalletAddress()
		if providerAddr != (common.Address{}) {
			if err := cm.payment.RecordJobCompleted(ctx, providerAddr); err != nil {
				logging.Warn("failed to record job completion in reputation",
					logging.ContainerID(containerID),
					logging.Err(err))
			}
		}
	}

	// Release image from GC tracking
	cm.unmarkImageInUse(deployment.Image)

	// Update status with volume retention
	cm.mu.Lock()
	deployment.Status = types.ContainerStatusStopped
	deployment.StoppedAt = time.Now()
	deployment.VolumeExpiresAt = deployment.StoppedAt.Add(volumeRetentionDuration)
	cm.mu.Unlock()

	// P1-10: Track stop event
	cm.stopsTotal.Add(1)

	// Persist state to disk
	cm.saveStateAsync()

	return nil
}

// volumeRetentionDuration is how long encrypted volumes are retained after stop.
const volumeRetentionDuration = 72 * time.Hour // 3 days

// Start restarts a stopped container whose volume is still retained.
func (cm *ContainerManager) Start(ctx context.Context, containerID string) error {
	cm.mu.Lock()
	deployment, exists := cm.deployments[containerID]
	if !exists {
		cm.mu.Unlock()
		return ErrDeploymentNotFound{ContainerID: containerID}
	}

	if deployment.Status != types.ContainerStatusStopped {
		cm.mu.Unlock()
		return fmt.Errorf("container %s is not stopped (status: %s)", containerID, deployment.Status)
	}

	if !deployment.VolumeExpiresAt.IsZero() && time.Now().After(deployment.VolumeExpiresAt) {
		cm.mu.Unlock()
		return fmt.Errorf("volume for container %s has expired", containerID)
	}
	cm.mu.Unlock()

	// Restart the container via containerd
	if cm.containerd != nil {
		if err := cm.containerd.StartContainer(ctx, containerID); err != nil {
			return fmt.Errorf("failed to start container: %w", err)
		}
	}

	// Update deployment state
	cm.mu.Lock()
	deployment.Status = types.ContainerStatusRunning
	deployment.StartedAt = time.Now()
	deployment.StoppedAt = time.Time{}
	deployment.VolumeExpiresAt = time.Time{}
	cm.mu.Unlock()

	cm.saveStateAsync()

	logging.Info("container restarted",
		logging.ContainerID(containerID))

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

	// Finalize escrow: release remaining payment or refund.
	// Only the originator node owns the escrow — replicas must not call payment ops.
	isOriginator := deployment.OriginatorID == cm.node.nodeInfo.ID
	if cm.payment != nil && isOriginator {
		jobID := payment.JobIDFromString(containerID)
		if deployment.Status == types.ContainerStatusStopped || deployment.Status == types.ContainerStatusRunning {
			// Release final payment for actual uptime before finalizing.
			// If the container was running (not already stopped), release
			// proportional payment so providers get paid for their work.
			if deployment.Status == types.ContainerStatusRunning && !deployment.StartedAt.IsZero() {
				uptime := time.Since(deployment.StartedAt)
				if err := cm.payment.ReleaseJobPayment(ctx, jobID, uptime); err != nil {
					logging.Warn("failed to release final payment on delete",
						logging.ContainerID(containerID),
						logging.Err(err))
				}
			}
			// Finalize the escrow
			if err := cm.payment.FinalizeJob(ctx, jobID); err != nil {
				logging.Warn("failed to finalize escrow on delete",
					logging.ContainerID(containerID),
					logging.Err(err))
			}
		} else {
			// Early termination or never started: refund
			if err := cm.payment.RefundJob(ctx, jobID); err != nil {
				logging.Warn("failed to refund escrow on delete",
					logging.ContainerID(containerID),
					logging.Err(err))
			}
		}
	}

	// Clean up pending deployment tracker
	cm.pendingMu.Lock()
	if pending, exists := cm.pendingDeployments[containerID]; exists {
		pending.close()
		delete(cm.pendingDeployments, containerID)
	}
	cm.pendingMu.Unlock()

	// Delete container
	if cm.containerd != nil {
		cm.containerd.DeleteContainer(ctx, containerID)
	}

	// Delete encrypted volume
	if cm.encryption != nil && deployment.EncryptedVolume != "" {
		cm.encryption.DeleteEncryptedVolume(containerID)
	}

	// Clean up subsystem state to prevent memory leaks
	if cm.healthMonitor != nil {
		cm.healthMonitor.RemoveContainer(containerID)
	}
	if cm.consensus != nil {
		cm.consensus.RemoveState(containerID)
	}
	// P1-2: Unregister health probe
	if cm.healthChecker != nil {
		cm.healthChecker.UnregisterProbe(containerID)
	}

	// P1-10: Track delete event
	cm.deletesTotal.Add(1)

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

// IsContainerdConnected checks if containerd is available and connected.
// If the ping fails, it attempts to reconnect once before returning false.
func (cm *ContainerManager) IsContainerdConnected() bool {
	if cm.containerd == nil {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if cm.containerd.Ping(ctx) == nil {
		return true
	}

	// Attempt reconnection
	logging.Warn("containerd connection lost, attempting reconnect",
		logging.Component("container_manager"))
	logsDir := filepath.Join(cm.dataDir, "logs")
	if cc, err := runtime.NewContainerdClient(cm.containerdSocket, "moltbunker", logsDir, cm.runtimeName, cm.kataConfig); err == nil {
		cm.containerd = cc
		logging.Info("containerd reconnected successfully",
			logging.Component("container_manager"))
		return true
	}
	return false
}

// activateEscrow calls SelectProviders on the escrow contract to transition
// the reservation from Created → Active. Must be called after replica acks
// identify the actual provider nodes. The local node is always provider[0].
func (cm *ContainerManager) activateEscrow(ctx context.Context, deploymentID string, acks []replicaAck) {
	if cm.payment == nil {
		return
	}

	jobID := payment.JobIDFromString(deploymentID)

	// Build provider address array: [local wallet, ack1 wallet, ack2 wallet]
	// Track seen wallets to prevent duplicates (same node acking twice, etc.)
	var providers [3]common.Address
	providers[0] = cm.node.WalletAddress()
	seen := map[common.Address]bool{providers[0]: true}

	sv := cm.router.StakeVerifier()
	idx := 1
	for _, ack := range acks {
		if !ack.Success || idx >= 3 {
			continue
		}
		// Resolve wallet from NodeID via StakeVerifier
		nodeIDBytes, err := hex.DecodeString(ack.NodeID)
		if err != nil || len(nodeIDBytes) != 32 {
			logging.Warn("failed to parse ack NodeID for escrow activation",
				logging.ContainerID(deploymentID),
				"node_id", ack.NodeID)
			continue
		}
		var nodeID types.NodeID
		copy(nodeID[:], nodeIDBytes)

		var wallet common.Address
		if sv != nil {
			if w, ok := sv.GetWallet(nodeID); ok && w != (common.Address{}) {
				wallet = w
			}
		}
		// Fallback: look up wallet via router peers
		if wallet == (common.Address{}) {
			for _, peer := range cm.router.GetPeers() {
				if peer.ID == nodeID && peer.WalletAddress != (common.Address{}) {
					wallet = peer.WalletAddress
					break
				}
			}
		}

		if wallet == (common.Address{}) || seen[wallet] {
			continue // Skip zero-address or duplicate
		}
		seen[wallet] = true
		providers[idx] = wallet
		idx++
	}

	if err := cm.payment.SelectProviders(ctx, jobID, providers); err != nil {
		logging.Warn("failed to activate escrow (SelectProviders)",
			logging.ContainerID(deploymentID),
			logging.Err(err))
		return
	}

	logging.Info("escrow activated with providers",
		logging.ContainerID(deploymentID),
		"provider0", providers[0].Hex()[:10],
		"provider1", providers[1].Hex()[:10],
		"provider2", providers[2].Hex()[:10])
}

// refundEscrow attempts to refund the escrow for a failed deployment.
// Errors are logged but not returned since the deployment already failed.
func (cm *ContainerManager) refundEscrow(ctx context.Context, deploymentID string) {
	if cm.payment == nil {
		return
	}
	jobID := payment.JobIDFromString(deploymentID)
	if err := cm.payment.RefundJob(ctx, jobID); err != nil {
		logging.Warn("failed to refund escrow after deployment failure",
			logging.ContainerID(deploymentID),
			logging.Err(err))
	} else {
		logging.Info("escrow refunded after deployment failure",
			logging.ContainerID(deploymentID))
	}
}

// determineDeploymentRegions builds the actual region list from the local node
// and connected peers. Always includes the local region, then adds distinct
// peer regions up to 3 total.
func (cm *ContainerManager) determineDeploymentRegions(localRegion string) []string {
	regions := []string{localRegion}
	seen := map[string]bool{localRegion: true}

	peers := cm.router.GetPeers()
	for _, peer := range peers {
		region := p2p.GetRegionFromCountry(peer.Country)
		if region == "" || region == "Unknown" {
			region = localRegion // Same-region fallback for unknown peers
		}
		if !seen[region] && len(regions) < 3 {
			regions = append(regions, region)
			seen[region] = true
		}
	}

	return regions
}

// pendingDeploymentTTL is how long a pending deployment tracker is kept before cleanup.
// Async deploys that don't call WaitForReplicas still need time for acks to arrive
// and escrow activation. 5 minutes is generous for image pull + container start.
const pendingDeploymentTTL = 5 * time.Minute

// cleanStalePendingDeployments periodically removes expired pending deployment trackers.
func (cm *ContainerManager) cleanStalePendingDeployments(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			cm.pendingMu.Lock()
			for id, pending := range cm.pendingDeployments {
				if time.Since(pending.created) > pendingDeploymentTTL {
					pending.close()
					delete(cm.pendingDeployments, id)
				}
			}
			cm.pendingMu.Unlock()
		}
	}
}

// cleanupExpiredVolumes periodically deletes stopped containers whose volume retention has expired.
func (cm *ContainerManager) cleanupExpiredVolumes(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			cm.mu.RLock()
			var expired []string
			for id, dep := range cm.deployments {
				if dep.Status == types.ContainerStatusStopped &&
					!dep.VolumeExpiresAt.IsZero() &&
					time.Now().After(dep.VolumeExpiresAt) {
					expired = append(expired, id)
				}
			}
			cm.mu.RUnlock()

			for _, id := range expired {
				logging.Info("auto-deleting container with expired volume",
					logging.ContainerID(id))
				if err := cm.Delete(ctx, id); err != nil {
					logging.Warn("failed to auto-delete expired container",
						logging.ContainerID(id),
						logging.Err(err))
				}
			}
		}
	}
}

// reconcileOnStartup syncs persisted deployment state with actual containerd status.
// Containers persist in containerd across daemon restarts; this updates our records
// to match reality (running containers stay running, crashed ones get marked stopped).
func (cm *ContainerManager) reconcileOnStartup(ctx context.Context) {
	// Collect deployment IDs and their persisted status
	cm.mu.RLock()
	type entry struct {
		id     string
		status types.ContainerStatus
	}
	entries := make([]entry, 0, len(cm.deployments))
	for id, dep := range cm.deployments {
		entries = append(entries, entry{id: id, status: dep.Status})
	}
	cm.mu.RUnlock()

	changed := false
	for _, e := range entries {
		actual, err := cm.containerd.GetContainerStatus(ctx, e.id)
		if err != nil {
			// Container doesn't exist in containerd
			if e.status == types.ContainerStatusRunning || e.status == types.ContainerStatusPending {
				cm.mu.Lock()
				if dep, ok := cm.deployments[e.id]; ok {
					logging.Warn("container missing from containerd on startup, marking stopped",
						logging.ContainerID(e.id),
						"previous_status", string(dep.Status))
					dep.Status = types.ContainerStatusStopped
					changed = true
				}
				cm.mu.Unlock()
			}
			continue
		}

		if e.status != actual {
			cm.mu.Lock()
			if dep, ok := cm.deployments[e.id]; ok {
				logging.Info("reconciled container status on startup",
					logging.ContainerID(e.id),
					"persisted", string(dep.Status),
					"actual", string(actual))
				dep.Status = actual
				changed = true
			}
			cm.mu.Unlock()
		}
	}

	if changed {
		if err := cm.saveState(); err != nil {
			logging.Error("failed to save state after startup reconciliation", logging.Err(err))
		}
	}

	logging.Info("startup reconciliation complete",
		"deployments", len(entries),
		"updated", changed)
}

// reconcileContainerStatus periodically checks containerd for actual container status
// and updates deployment records when reality diverges (e.g., OOM kills, crashes).
func (cm *ContainerManager) reconcileContainerStatus(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			cm.mu.RLock()
			ids := make([]string, 0, len(cm.deployments))
			for id, dep := range cm.deployments {
				if dep.Status == types.ContainerStatusRunning {
					ids = append(ids, id)
				}
			}
			cm.mu.RUnlock()

			for _, id := range ids {
				actual, err := cm.containerd.GetContainerStatus(ctx, id)
				if err != nil {
					continue // Container may not exist in containerd (remote replica)
				}
				cm.mu.Lock()
				dep, exists := cm.deployments[id]
				if exists && dep.Status == types.ContainerStatusRunning && actual != types.ContainerStatusRunning {
					logging.Warn("container status drift detected",
						logging.ContainerID(id),
						"recorded", string(dep.Status),
						"actual", string(actual))
					dep.Status = actual
					cm.saveStateAsync()
				}
				cm.mu.Unlock()
			}
		}
	}
}

// ContainerLifecycleMetrics holds lifecycle event counters for containers.
type ContainerLifecycleMetrics struct {
	Deploys  int64 `json:"deploys"`
	Stops    int64 `json:"stops"`
	Deletes  int64 `json:"deletes"`
	Failures int64 `json:"failures"`
}

// LifecycleMetrics returns a snapshot of container lifecycle event counters.
func (cm *ContainerManager) LifecycleMetrics() ContainerLifecycleMetrics {
	return ContainerLifecycleMetrics{
		Deploys:  cm.deploysTotal.Load(),
		Stops:    cm.stopsTotal.Load(),
		Deletes:  cm.deletesTotal.Load(),
		Failures: cm.failuresTotal.Load(),
	}
}

// Close closes the container manager and all managed resources.
// Containers are NOT stopped — they persist in containerd across daemon restarts.
// Users must explicitly stop/delete containers.
func (cm *ContainerManager) Close() error {
	// Save state before closing
	if err := cm.saveState(); err != nil {
		logging.Error("failed to save state on close", logging.Err(err))
	}

	// Stop health monitor
	if cm.healthMonitor != nil {
		cm.healthMonitor.Stop()
	}

	// Close exec sessions (daemon-side only, containers keep running)
	if cm.execStreams != nil {
		cm.mu.RLock()
		for containerID := range cm.deployments {
			cm.execStreams.CloseAllForContainer(containerID)
		}
		cm.mu.RUnlock()
	}

	// P1-2: Stop container-level health checker
	if cm.healthChecker != nil {
		cm.healthChecker.Stop()
	}

	// Stop image garbage collector
	if cm.imageGC != nil {
		cm.imageGC.Stop()
	}

	// Close containerd client (does NOT stop containers)
	if cm.containerd != nil {
		cm.containerd.Close()
	}

	// Stop Tor
	if cm.torService != nil {
		cm.torService.Stop()
	}

	return nil
}

// recordJobFailed records a failed job in the reputation contract.
// Errors are logged but not returned — reputation is best-effort.
func (cm *ContainerManager) recordJobFailed(ctx context.Context, containerID string) {
	if cm.payment == nil {
		return
	}
	providerAddr := cm.node.WalletAddress()
	if providerAddr == (common.Address{}) {
		return
	}
	if err := cm.payment.RecordJobFailed(ctx, providerAddr); err != nil {
		logging.Warn("failed to record job failure in reputation",
			logging.ContainerID(containerID),
			logging.Err(err))
	}
}

// runAttestationLoop submits a hardware/software attestation hash on startup
// and then every 24 hours. The hash covers the node's runtime capabilities,
// provider tier, and software version — allowing the network to verify that
// providers are running legitimate, up-to-date infrastructure.
func (cm *ContainerManager) runAttestationLoop(ctx context.Context) {
	// Submit initial attestation after a short startup delay
	select {
	case <-ctx.Done():
		return
	case <-time.After(30 * time.Second):
	}

	cm.submitAttestation(ctx)

	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			cm.submitAttestation(ctx)
		}
	}
}

// submitAttestation generates and submits an attestation hash from the current
// node's runtime state.
func (cm *ContainerManager) submitAttestation(ctx context.Context) {
	// Build attestation data from node state
	providerTier := runtime.DetectProviderTier()
	rtCaps := runtime.DetectRuntime("")

	h := sha256Sum(
		[]byte(string(providerTier)),
		[]byte(rtCaps.RuntimeName),
		[]byte(cm.node.nodeInfo.ID.String()),
		[]byte(time.Now().UTC().Format("2006-01-02")),
	)

	if err := cm.payment.SubmitAttestation(ctx, h); err != nil {
		logging.Warn("failed to submit attestation",
			logging.Err(err))
	} else {
		logging.Info("attestation submitted",
			"tier", string(providerTier),
			"runtime", rtCaps.RuntimeName,
			"hash", fmt.Sprintf("%x", h[:8]))
	}
}

// sha256Sum computes a SHA-256 hash over concatenated byte slices.
func sha256Sum(parts ...[]byte) [32]byte {
	h := sha256.New()
	for _, p := range parts {
		h.Write(p)
	}
	var out [32]byte
	copy(out[:], h.Sum(nil))
	return out
}

// monitorHealthFailures periodically checks for unhealthy replicas and reports
// violations on-chain via the slashing contract. This bridges the gap between
// health monitoring (redundancy package) and on-chain accountability (payment package).
func (cm *ContainerManager) monitorHealthFailures(ctx context.Context) {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	// Track which containers we've already reported to avoid duplicate reports
	reported := make(map[string]time.Time)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			cm.mu.RLock()
			var running []string
			for id, dep := range cm.deployments {
				if dep.Status == types.ContainerStatusRunning {
					running = append(running, id)
				}
			}
			cm.mu.RUnlock()

			for _, containerID := range running {
				unhealthy := cm.healthMonitor.GetUnhealthyReplicas(containerID)
				for _, replicaIdx := range unhealthy {
					reportKey := fmt.Sprintf("%s-%d", containerID, replicaIdx)

					// Rate-limit: don't re-report within 10 minutes
					if lastReport, exists := reported[reportKey]; exists && time.Since(lastReport) < 10*time.Minute {
						continue
					}

					// Report the violation on-chain
					jobID := payment.JobIDFromString(containerID)
					_, err := cm.payment.ReportViolation(ctx,
						cm.node.WalletAddress(), // provider to report against
						jobID,
						payment.ViolationDowntime,
						[]byte(fmt.Sprintf("replica %d unhealthy", replicaIdx)),
					)
					if err != nil {
						logging.Warn("failed to report health violation",
							logging.ContainerID(containerID),
							"replica_index", replicaIdx,
							logging.Err(err))
					} else {
						logging.Info("reported health violation on-chain",
							logging.ContainerID(containerID),
							"replica_index", replicaIdx)
					}
					reported[reportKey] = time.Now()
				}
			}

			// Clean old entries from reported map (older than 1 hour)
			for key, t := range reported {
				if time.Since(t) > time.Hour {
					delete(reported, key)
				}
			}
		}
	}
}

// markImageInUse tracks an image as actively used by a container.
func (cm *ContainerManager) markImageInUse(imageRef string) {
	if cm.imageGC != nil {
		cm.imageGC.MarkInUse(imageRef)
	}
}

// unmarkImageInUse marks an image as no longer used by a container.
func (cm *ContainerManager) unmarkImageInUse(imageRef string) {
	if cm.imageGC != nil {
		cm.imageGC.UnmarkInUse(imageRef)
	}
}
