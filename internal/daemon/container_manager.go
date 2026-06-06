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
	goruntime "runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/moltbunker/moltbunker/internal/ingress"
	"github.com/moltbunker/moltbunker/internal/logging"
	"github.com/moltbunker/moltbunker/internal/molt"
	"github.com/moltbunker/moltbunker/internal/networking"
	"github.com/moltbunker/moltbunker/internal/p2p"
	"github.com/moltbunker/moltbunker/internal/payment"
	"github.com/moltbunker/moltbunker/internal/redundancy"
	"github.com/moltbunker/moltbunker/internal/runtime"
	"github.com/moltbunker/moltbunker/internal/security"
	"github.com/moltbunker/moltbunker/internal/state"
	"github.com/moltbunker/moltbunker/internal/tor"
	"github.com/moltbunker/moltbunker/internal/util"
	"github.com/moltbunker/moltbunker/pkg/types"
)

// execKeyLen is the required length of a recovered plaintext exec key. The
// exec-agent enforces exactly this size at startup.
const execKeyLen = 32

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

	stateStore  state.StateStore // nil = legacy JSON fallback
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

	networkManager   *networking.NetworkManager

	dataDir          string
	containerdSocket string
	runtimeName      string          // resolved OCI runtime name for reconnection
	kataConfig       *runtime.KataConfig // saved for reconnection path
	acceptServices   bool                // accept long-running services
	acceptJobs       bool                // accept batch jobs
	acceptFunctions  bool                // accept serverless functions (Molt)
	imageGC          *runtime.ImageGC    // image garbage collector (nil if no containerd)
	cleanupMgr       *runtime.CleanupManager // P1-1: orphan container cleanup
	healthChecker    *runtime.HealthChecker  // P1-2: container-level health probes

	// R4: image vulnerability scanner. Never nil after NewContainerManager —
	// a NoopScanner when scanning is disabled / trivy is absent, so the scan
	// gate never blocks a deploy on a host without trivy.
	imageScanner runtime.ImageScanner

	// R13/R14: per-deployment network/egress policy. policyStore is the source
	// of truth; policyEnforcer applies rules at container setup (real nft exec
	// is a Linux-only stub today — off-Linux it only records intent).
	policyStore    *networking.PolicyStore
	policyEnforcer networking.PolicyEnforcer

	// Molt (WASM serverless) manager
	moltManager *MoltManager

	// Reverse tunnel manager (optional, set via SetReverseTunnelManager)
	reverseTunnel *ReverseTunnelManager

	// providerKey is the daemon's stable X25519 keypair used to unwrap
	// E2E-encrypted exec keys (ECIES). nil if it could not be loaded/created.
	providerKey *ProviderKeyManager

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
		// R20: attach the per-tenant security-profile store BEFORE
		// LoadExistingContainers (below) runs, so reattached containers recover
		// their stored profile instead of silently downgrading to the default.
		attachProfileStore(cc, config.DataDir)
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
		if err := encryption.LoadExistingVolumes(); err != nil {
			return nil, fmt.Errorf("failed to load existing encrypted volumes: %w", err)
		}
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

	// R4: construct the image scanner once. Real Trivy only when explicitly
	// enabled AND the binary is on PATH; otherwise a NoopScanner so a host
	// without trivy never fails a deploy. The scanner is always non-nil.
	imageScanner := buildImageScanner(config.EnableImageScan)

	// R13/R14: construct the network-policy store + enforcer once. The enforcer
	// is a no-op recorder off Linux and a (currently stubbed) nft applier on
	// Linux; either way nil/empty policies mean allow-all.
	policyStore := networking.NewPolicyStore()
	policyEnforcer := networking.NewNftPolicyEnforcer(policyStore)

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
		networkManager:     networking.NewNetworkManager(),
		stateStore:         config.StateStore,
		deployments:        make(map[string]*Deployment),
		pendingDeployments: make(map[string]*pendingDeployment),
		execStreams:        NewExecStreamManager(),
		execRelays:         make(map[string]*ExecRelay),
		dataDir:            config.DataDir,
		containerdSocket:   config.ContainerdSocket,
		runtimeName:        rtCaps.RuntimeName,
		kataConfig:         config.KataConfig,
		acceptServices:     config.AcceptServices,
		acceptJobs:         config.AcceptJobs,
		acceptFunctions:    config.AcceptFunctions,
		imageScanner:       imageScanner,
		policyStore:        policyStore,
		policyEnforcer:     policyEnforcer,
	}

	// Load (or create) the stable provider X25519 keypair used to unwrap
	// E2E-encrypted exec keys. A failure here is non-fatal: the daemon still
	// runs, but E2E exec falls back to unavailable (deploys that send a sealed
	// exec key will be rejected at prepareExecAgent).
	if pk, pkErr := LoadOrCreateProviderKey(config.DataDir); pkErr != nil {
		logging.Warn("failed to load provider X25519 keypair; E2E exec disabled",
			logging.Err(pkErr), logging.Component("exec"))
	} else {
		cm.providerKey = pk
	}

	// Initialize Molt (WASM serverless) runtime if enabled
	if config.MoltEnabled {
		moltCfg := molt.DefaultMoltConfig()
		if config.MoltConfig != nil {
			moltCfg = *config.MoltConfig
		}
		moltRuntime, moltErr := molt.NewMoltRuntime(ctx, moltCfg)
		if moltErr != nil {
			logging.Warn("molt runtime not available", logging.Err(moltErr))
		} else {
			cm.moltManager = NewMoltManager(moltRuntime)
			logging.Info("molt runtime initialized",
				"memory_limit_mb", moltCfg.MemoryLimitMB,
				"timeout_ms", moltCfg.TimeoutMs,
				"max_instances", moltCfg.MaxInstances,
			)
		}
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
		if err := crt.LoadExistingContainers(ctx); err != nil {
			return nil, fmt.Errorf("failed to load existing containers: %w", err)
		}
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
		Spot:            req.Spot,
		// R3/R4/R13/R14: carry the per-deployment security policy onto the
		// gossiped Deployment so replica nodes enforce the same gates. All
		// fields are opt-out: empty/nil => legacy behavior.
		RequireSignature:  req.RequireSignature,
		TrustedPublishers: req.TrustedPublishers,
		ImageSignature:    req.ImageSignature,
		IgnoreCVEs:        req.IgnoreCVEs,
		NetworkPolicy:     req.NetworkPolicy,
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

	// Create container with security hardening.
	//
	// R3 (image signature) and R4 (CVE scan) gates are populated from the deploy
	// request. They are opt-out by default: toTrustPolicy/toImageSignature yield
	// no-op values when the request carries nothing, and cm.imageScanner is a
	// NoopScanner unless scanning was enabled with trivy present. So a request
	// with none of the new fields produces identical behavior to before.
	secConfig := runtime.SecureContainerConfig{
		ID:              deploymentID,
		ImageRef:        req.Image,
		Resources:       req.Resources,
		SecurityProfile: types.DeploymentSecurityProfile(),
		ImageSignature:  toImageSignature(req.ImageSignature),
		TrustPolicy:     toTrustPolicy(req.RequireSignature, req.TrustedPublishers),
		Scanner:         cm.imageScanner,
		ScanPolicy:      resolveScanPolicy(req.IgnoreCVEs),
	}

	// Inject exec-agent if an encrypted exec key envelope is provided (E2E
	// encrypted exec). The CLI seals the exec key to this provider's stable
	// X25519 public key; prepareExecAgent unwraps it with the private key.
	if len(req.EncryptedExecKey) > 0 {
		mounts, keyPath, err := cm.prepareExecAgent(deploymentID, req.EncryptedExecKey, req.RequesterEphemeralPubKey)
		if err != nil {
			logging.Warn("failed to prepare exec-agent, continuing without E2E exec",
				logging.ContainerID(deploymentID),
				logging.Err(err))
		} else {
			secConfig.BindMounts = mounts
			deployment.ExecAgentEnabled = true
			deployment.ExecKeyPath = keyPath
			deployment.DeployNonce = req.DeployNonce
			// Persist the envelope on the deployment so it can reach replicas.
			deployment.EncryptedExecKey = req.EncryptedExecKey
			deployment.ExecKeyNonce = req.ExecKeyNonce
			deployment.RequesterEphemeralPubKey = req.RequesterEphemeralPubKey
		}
	}

	container, err := cm.containerd.CreateSecureContainer(ctx, secConfig)
	if err != nil {
		// Cleanup encrypted volume on failure
		if cm.encryption != nil && encryptedVolume != nil {
			if cleanupErr := cm.encryption.DeleteEncryptedVolume(deploymentID); cleanupErr != nil {
				logging.Warn("failed to delete encrypted volume during cleanup",
					"deployment_id", deploymentID,
					logging.Err(cleanupErr),
					logging.Component("container_manager"))
			}
		}
		cm.cleanupExecKey(deployment)
		cm.recordJobFailed(ctx, deploymentID)
		cm.failuresTotal.Add(1) // P1-10
		return fmt.Errorf("failed to create container: %w", err)
	}

	// Start container
	if err := cm.containerd.StartContainer(ctx, deploymentID); err != nil {
		if cleanupErr := cm.containerd.DeleteContainer(ctx, deploymentID); cleanupErr != nil {
			logging.Warn("failed to delete container during cleanup",
				"deployment_id", deploymentID,
				logging.Err(cleanupErr),
				logging.Component("container_manager"))
		}
		if cm.encryption != nil && encryptedVolume != nil {
			if cleanupErr := cm.encryption.DeleteEncryptedVolume(deploymentID); cleanupErr != nil {
				logging.Warn("failed to delete encrypted volume during cleanup",
					"deployment_id", deploymentID,
					logging.Err(cleanupErr),
					logging.Component("container_manager"))
			}
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

	// Set up networking for exposed ports
	if len(req.ExposePorts) > 0 && cm.networkManager != nil {
		netPorts := convertExposedPorts(req.ExposePorts)
		containerNet, err := cm.networkManager.SetupNetwork(deploymentID, netPorts)
		if err != nil {
			logging.Warn("network setup failed for exposed ports",
				logging.ContainerID(deploymentID),
				logging.Err(err))
		} else {
			deployment.ExposedPorts = req.ExposePorts
			// R13/R14: apply per-deployment network/egress policy once the
			// container IP is known. nil/empty policy => allow-all (no-op).
			// Real nft enforcement is a Linux-only stub; off-Linux this only
			// records intent.
			cm.applyNetworkPolicy(deploymentID, containerNet.ContainerIP(), req.NetworkPolicy)
			// no-op: ingress domain currently hardcoded; future versions may
			// derive it from cm.node.nodeInfo when config plumbing is added.
			ingressDomain := "moltbunker.dev"
			subdomain := deploymentID[len("dep-"):]
			if len(subdomain) > 8 {
				subdomain = subdomain[:8]
			}
			for _, p := range netPorts {
				hostPort, _ := containerNet.ResolvePort(p.ContainerPort)
				deployment.PublicURLs = append(deployment.PublicURLs,
					fmt.Sprintf("https://%s.%s", subdomain, ingressDomain))
				cm.publishServiceExposure(deploymentID, p.ContainerPort, hostPort)
			}
		}
	}

	return nil
}

// prepareExecAgent unwraps the ECIES-sealed exec key with the daemon's stable
// X25519 private key, writes the recovered 32-byte plaintext key to a secure
// file, and returns bind-mounts for the exec-agent binary and key file. The
// exec-agent binary path is resolved from the daemon's own binary directory.
//
// The exec key is never logged. The plaintext only exists transiently in memory
// and in the 0600 key file that is bind-mounted read-only into the container.
func (cm *ContainerManager) prepareExecAgent(deploymentID string, encryptedExecKey, ephemeralPubKey []byte) ([]runtime.BindMount, string, error) {
	if cm.providerKey == nil {
		return nil, "", fmt.Errorf("provider X25519 key unavailable; cannot unwrap exec key")
	}

	// Unwrap the ECIES envelope to recover the 32-byte plaintext exec key.
	execKey, err := security.OpenFromX25519(cm.providerKey.privateKey(), &security.X25519Envelope{
		EphemeralPub: ephemeralPubKey,
		Ciphertext:   encryptedExecKey,
	})
	if err != nil {
		return nil, "", fmt.Errorf("unwrap exec key: %w", err)
	}
	if len(execKey) != execKeyLen {
		return nil, "", fmt.Errorf("unwrapped exec key has invalid size %d (expected %d)", len(execKey), execKeyLen)
	}

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
//
// Candidate names are tried in order: the GOARCH-suffixed name for the host
// architecture (e.g. "exec-agent-arm64" or "exec-agent-amd64", matching the
// Makefile/.goreleaser artifact names), then the bare "exec-agent" name (the
// Dockerfile/deploy.sh install name). This ensures arm64 providers resolve the
// correct binary instead of silently falling back to no-E2E exec.
func (cm *ContainerManager) resolveExecAgentPath() string {
	candidates := []string{"exec-agent-" + goruntime.GOARCH, "exec-agent"}

	// Check alongside the daemon binary
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
		for _, name := range candidates {
			candidate := filepath.Join(dir, name)
			if _, err := os.Stat(candidate); err == nil {
				return candidate
			}
		}
	}
	// Fallback: /usr/local/bin
	for _, name := range candidates {
		candidate := filepath.Join("/usr/local/bin", name)
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	return ""
}

// ProviderExecPubKey returns the daemon's stable X25519 public key used to seal
// exec keys (ECIES). Returns nil if the keypair is unavailable.
func (cm *ContainerManager) ProviderExecPubKey() []byte {
	if cm.providerKey == nil {
		return nil
	}
	return cm.providerKey.PublicKey()
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

	// Tear down networking for exposed ports
	if cm.networkManager != nil {
		if err := cm.networkManager.TeardownNetwork(containerID); err != nil {
			logging.Warn("failed to teardown network on stop",
				logging.ContainerID(containerID), logging.Err(err))
		}
	}
	// R13/R14: drop any per-deployment network policy rules.
	cm.removeNetworkPolicy(containerID)
	cm.removeServiceExposure(containerID)

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

		// Finalize escrow on stop so it doesn't leak if Delete is never called.
		jobID := payment.JobIDFromString(containerID)
		if err := cm.payment.FinalizeJob(ctx, jobID); err != nil {
			logging.Warn("failed to finalize escrow on stop",
				logging.ContainerID(containerID),
				logging.Err(err))
		}
	}

	// Release image from GC tracking
	cm.unmarkImageInUse(deployment.Image)

	// Update status with volume retention. Also remove the plaintext exec_key
	// file (and its bind-mount source) so no key material is left on disk after
	// teardown. cleanupExecKey mutates deployment.ExecKeyPath, so it runs under
	// the same lock as the status update.
	cm.mu.Lock()
	deployment.Status = types.ContainerStatusStopped
	deployment.StoppedAt = time.Now()
	deployment.VolumeExpiresAt = deployment.StoppedAt.Add(volumeRetentionDuration)
	cm.cleanupExecKey(deployment)
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

	// C2: Record job failure in reputation contract for failed/error deployments.
	// RecordJobCompleted is already called on Stop() (line 843). Here we track
	// deployments that never ran successfully (error, failed, pending states).
	if cm.payment != nil && isOriginator {
		if deployment.Status == types.ContainerStatusFailed || deployment.Status == types.ContainerStatusPending {
			providerAddr := cm.node.WalletAddress()
			if providerAddr != (common.Address{}) {
				if err := cm.payment.RecordJobFailed(ctx, providerAddr); err != nil {
					logging.Warn("failed to record job failure in reputation",
						logging.ContainerID(containerID),
						logging.Err(err))
				}
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

	// Tear down networking
	if cm.networkManager != nil {
		if err := cm.networkManager.TeardownNetwork(containerID); err != nil {
			logging.Warn("failed to teardown network on delete",
				logging.ContainerID(containerID),
				logging.Err(err))
		}
	}
	// R13/R14: drop any per-deployment network policy rules.
	cm.removeNetworkPolicy(containerID)
	cm.removeServiceExposure(containerID)

	// Delete container
	if cm.containerd != nil {
		if err := cm.containerd.DeleteContainer(ctx, containerID); err != nil {
			logging.Warn("failed to delete container during cleanup",
				logging.ContainerID(containerID),
				logging.Err(err),
				logging.Component("container_manager"))
		}
	}

	// Delete encrypted volume
	if cm.encryption != nil && deployment.EncryptedVolume != "" {
		if err := cm.encryption.DeleteEncryptedVolume(containerID); err != nil {
			logging.Warn("failed to delete encrypted volume during cleanup",
				logging.ContainerID(containerID),
				logging.Err(err),
				logging.Component("container_manager"))
		}
	}

	// Remove the plaintext exec_key file (and its bind-mount source) so no key
	// material is left on disk after teardown. The deployment is already removed
	// from cm.deployments, but cleanupExecKey mutates deployment.ExecKeyPath, so
	// take the lock to avoid racing the async state writers.
	cm.mu.Lock()
	cm.cleanupExecKey(deployment)
	cm.mu.Unlock()

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

	// Persist state: delete just this deployment (more efficient than full re-save)
	cm.deleteDeploymentState(containerID)

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
		// R20: re-attach the profile store so persistence survives a reconnect,
		// not just a process restart.
		attachProfileStore(cc, cm.dataDir)
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

					// Reconcile payment for crashed containers: record failure and refund.
					// Only the originator owns the escrow.
					isOriginator := dep.OriginatorID == cm.node.nodeInfo.ID
					if cm.payment != nil && isOriginator {
						providerAddr := cm.node.WalletAddress()
						if providerAddr != (common.Address{}) {
							jobID := payment.JobIDFromString(id)
							if err := cm.payment.RecordJobFailed(ctx, providerAddr); err != nil {
								logging.Warn("failed to record job failure for crashed container",
									logging.ContainerID(id), logging.Err(err))
							}
							if err := cm.payment.RefundJob(ctx, jobID); err != nil {
								logging.Warn("failed to refund escrow for crashed container",
									logging.ContainerID(id), logging.Err(err))
							}
						}
					}
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

// FinalizeAllEscrows releases proportional payment and finalizes escrows
// for all running containers. Called during graceful shutdown to prevent
// escrows from being stranded if the daemon doesn't restart.
func (cm *ContainerManager) FinalizeAllEscrows(ctx context.Context) {
	if cm.payment == nil {
		return
	}

	cm.mu.RLock()
	var running []struct {
		id           string
		startedAt    time.Time
		originatorID types.NodeID
	}
	for id, dep := range cm.deployments {
		if dep.Status == types.ContainerStatusRunning && !dep.StartedAt.IsZero() {
			running = append(running, struct {
				id           string
				startedAt    time.Time
				originatorID types.NodeID
			}{id, dep.StartedAt, dep.OriginatorID})
		}
	}
	cm.mu.RUnlock()

	for _, r := range running {
		if r.originatorID != cm.node.nodeInfo.ID {
			continue // Only the originator owns the escrow
		}
		jobID := payment.JobIDFromString(r.id)
		uptime := time.Since(r.startedAt)

		if err := cm.payment.ReleaseJobPayment(ctx, jobID, uptime); err != nil {
			logging.Warn("shutdown: failed to release payment",
				logging.ContainerID(r.id), logging.Err(err))
		}
		if err := cm.payment.FinalizeJob(ctx, jobID); err != nil {
			logging.Warn("shutdown: failed to finalize escrow",
				logging.ContainerID(r.id), logging.Err(err))
		} else {
			logging.Info("shutdown: escrow finalized",
				logging.ContainerID(r.id))
		}
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

	// Close Molt (WASM) manager and runtime
	if cm.moltManager != nil {
		if err := cm.moltManager.Close(context.Background()); err != nil {
			logging.Error("failed to close molt manager", logging.Err(err))
		}
	}

	// Close containerd client (does NOT stop containers)
	if cm.containerd != nil {
		cm.containerd.Close()
	}

	// Stop Tor
	if cm.torService != nil {
		if err := cm.torService.Stop(); err != nil {
			logging.Warn("failed to stop tor service during shutdown",
				logging.Err(err),
				logging.Component("container_manager"))
		}
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
	cm.checkMissedAttestations(ctx)

	attestTicker := time.NewTicker(24 * time.Hour)
	checkTicker := time.NewTicker(6 * time.Hour)
	defer attestTicker.Stop()
	defer checkTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-attestTicker.C:
			cm.submitAttestation(ctx)
		case <-checkTicker.C:
			cm.checkMissedAttestations(ctx)
		}
	}
}

// checkMissedAttestations checks if this provider has missed attestations
// and logs a warning. This ensures operators are alerted before penalties.
func (cm *ContainerManager) checkMissedAttestations(ctx context.Context) {
	providerAddr := cm.node.WalletAddress()
	if providerAddr == (common.Address{}) {
		return
	}
	if err := cm.payment.CheckMissedAttestations(ctx, providerAddr); err != nil {
		logging.Warn("missed attestation check failed or attestations missed",
			logging.Err(err),
			logging.Component("attestation"))
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

					// Report the violation on-chain.
					// Each node monitors its own replicas and reports against itself.
					// In multi-node, the originator receives health status via gossip
					// and each provider self-reports their own violations.
					jobID := payment.JobIDFromString(containerID)
					providerAddr := cm.node.WalletAddress()
					_, err := cm.payment.ReportViolation(ctx,
						providerAddr,
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

						// Bridge slashing → reputation: record the slash event
						providerAddr := cm.node.WalletAddress()
						if providerAddr != (common.Address{}) {
							if slashErr := cm.payment.RecordSlashEvent(ctx, providerAddr); slashErr != nil {
								logging.Warn("failed to record slash event in reputation",
									logging.ContainerID(containerID),
									logging.Err(slashErr))
							}
						}
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

// GossipProtocol returns the gossip protocol instance.
func (cm *ContainerManager) GossipProtocol() *p2p.GossipProtocol {
	return cm.gossip
}

// NetworkManager returns the network manager for port resolution.
func (cm *ContainerManager) NetworkManager() *networking.NetworkManager {
	return cm.networkManager
}

// SetReverseTunnelManager sets the reverse tunnel manager for exposing
// deployments via reverse tunnels to NAT'd providers.
func (cm *ContainerManager) SetReverseTunnelManager(rtm *ReverseTunnelManager) {
	cm.reverseTunnel = rtm
}

// MoltManager returns the Molt (WASM serverless) manager.
// Returns nil if the Molt runtime failed to initialize.
func (cm *ContainerManager) MoltManager() *MoltManager {
	return cm.moltManager
}

// convertExposedPorts converts daemon ExposedPort to networking ExposedPort.
func convertExposedPorts(ports []ExposedPort) []networking.ExposedPort {
	result := make([]networking.ExposedPort, len(ports))
	for i, p := range ports {
		proto := p.Protocol
		if proto == "" {
			proto = "tcp"
		}
		result[i] = networking.ExposedPort{
			ContainerPort: p.ContainerPort,
			Protocol:      proto,
		}
	}
	return result
}

// publishServiceExposure writes an exposed service entry to gossip state
// so ingress nodes can discover and route traffic to this container.
func (cm *ContainerManager) publishServiceExposure(deploymentID string, containerPort, hostPort int) {
	if cm.gossip == nil {
		return
	}

	nodeAddr := ""
	if cm.node != nil && cm.node.nodeInfo != nil {
		nodeAddr = fmt.Sprintf("%s:%d", cm.node.nodeInfo.Address, cm.node.nodeInfo.Port+2) // tunnel port
	}

	entry := &ingress.ServiceEntry{
		DeploymentID:   deploymentID,
		ProviderNodeID: cm.node.nodeInfo.ID.String(),
		ProviderAddr:   nodeAddr,
		ContainerPort:  containerPort,
		HostPort:       hostPort,
		RuntimeType:    "container",
	}

	// Validate deploymentID has no colons to prevent gossip key injection
	for i := 0; i < len(deploymentID); i++ {
		if deploymentID[i] == ':' {
			logging.Error("deployment ID contains colon, refusing to publish gossip key",
				logging.ContainerID(deploymentID))
			return
		}
	}

	key := fmt.Sprintf("expose:%s:%d", deploymentID, containerPort)
	cm.gossip.UpdateState(key, entry)

	logging.Info("published service exposure to gossip",
		logging.ContainerID(deploymentID),
		"container_port", containerPort,
		"host_port", hostPort,
		"key", key)

	// Also expose via reverse tunnel if configured (for NAT'd providers)
	if cm.reverseTunnel != nil {
		cm.reverseTunnel.Expose(deploymentID, containerPort)
	}
}

// removeServiceExposure removes all exposed service entries for a deployment from gossip.
// Gossip has no DeleteState, so we set values to nil — the adapter filters these out.
func (cm *ContainerManager) removeServiceExposure(containerID string) {
	if cm.gossip == nil {
		return
	}

	cm.mu.RLock()
	dep, exists := cm.deployments[containerID]
	cm.mu.RUnlock()
	if !exists || len(dep.ExposedPorts) == 0 {
		return
	}

	for _, p := range dep.ExposedPorts {
		key := fmt.Sprintf("expose:%s:%d", containerID, p.ContainerPort)
		cm.gossip.UpdateState(key, nil) // nil = removed
	}

	// Disconnect reverse tunnel if active
	if cm.reverseTunnel != nil {
		cm.reverseTunnel.Unexpose(containerID)
	}
}

// publishMoltServiceExposure writes a Molt deployment's service entry to gossip.
// Molt deployments expose port 80 (HTTP handler) and are marked with RuntimeType "molt".
//
//nolint:unused
func (cm *ContainerManager) publishMoltServiceExposure(deploymentID string) {
	if cm.gossip == nil || cm.node == nil || cm.node.nodeInfo == nil {
		return
	}

	nodeAddr := fmt.Sprintf("%s:%d", cm.node.nodeInfo.Address, cm.node.nodeInfo.Port+2)

	entry := &ingress.ServiceEntry{
		DeploymentID:   deploymentID,
		ProviderNodeID: cm.node.nodeInfo.ID.String(),
		ProviderAddr:   nodeAddr,
		ContainerPort:  80,
		HostPort:       0, // Molt uses HTTP handler, not host port mapping
		RuntimeType:    "molt",
	}

	key := fmt.Sprintf("expose:%s:%d", deploymentID, 80)
	cm.gossip.UpdateState(key, entry)

	logging.Info("published molt service exposure to gossip",
		"deployment_id", deploymentID,
		"key", key)
}
