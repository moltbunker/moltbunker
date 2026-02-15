package main

import (
	"context"
	"crypto/ecdsa"
	"crypto/sha256"
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	goruntime "runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/moltbunker/moltbunker/internal/api"
	"github.com/moltbunker/moltbunker/internal/cloning"
	"github.com/moltbunker/moltbunker/internal/config"
	"github.com/moltbunker/moltbunker/internal/daemon"
	"github.com/moltbunker/moltbunker/internal/identity"
	"github.com/moltbunker/moltbunker/internal/logging"
	"github.com/moltbunker/moltbunker/internal/p2p"
	"github.com/moltbunker/moltbunker/internal/payment"
	"github.com/moltbunker/moltbunker/internal/snapshot"
	"github.com/moltbunker/moltbunker/internal/threat"
	"github.com/moltbunker/moltbunker/internal/util"
)

var (
	configPath  = flag.String("config", "", "Path to config file (default: ~/.moltbunker/config.yaml)")
	port        = flag.Int("port", 0, "P2P port (overrides config)")
	keyPath     = flag.String("key", "", "Path to node key (overrides config)")
	keystoreDir = flag.String("keystore", "", "Path to keystore (overrides config)")
	dataDir     = flag.String("data", "", "Path to data directory (overrides config)")
	socketPath  = flag.String("socket", "", "Unix socket path for API (overrides config)")
	httpAddr    = flag.String("http", "127.0.0.1:8080", "HTTP API listen address (set empty to disable)")
)

func main() {
	flag.Parse()

	// Load configuration
	cfgPath := *configPath
	if cfgPath == "" {
		cfgPath = config.DefaultConfigPath()
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Override with command-line flags
	if *port != 0 {
		cfg.Daemon.Port = *port
	}
	if *keyPath != "" {
		cfg.Daemon.KeyPath = *keyPath
	}
	if *keystoreDir != "" {
		cfg.Daemon.KeystoreDir = *keystoreDir
	}
	if *dataDir != "" {
		cfg.Daemon.DataDir = *dataDir
	}
	if *socketPath != "" {
		cfg.Daemon.SocketPath = *socketPath
	}

	// Ensure directories exist
	if err := cfg.EnsureDirectories(); err != nil {
		log.Fatalf("Failed to create directories: %v", err)
	}

	// M1: PID file handling
	pidPath := filepath.Join(cfg.Daemon.DataDir, "daemon.pid")
	if err := checkAndWritePIDFile(pidPath); err != nil {
		log.Fatalf("PID file check failed: %v", err)
	}
	defer os.Remove(pidPath)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Create node with full configuration
	node, err := daemon.NewNodeWithConfig(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to create node: %v", err)
	}

	// Load wallet and resolve password for on-chain transactions
	walletPrivKey := loadWallet(cfg, node)

	// C1: Load persisted state from data dir
	stateDir := filepath.Join(cfg.Daemon.DataDir, "state")
	banListPath := filepath.Join(stateDir, "banlist.json")
	addressBookPath := filepath.Join(stateDir, "addressbook.json")
	certPinsPath := filepath.Join(stateDir, "certpins.json")

	// Create and load ban list
	banList := p2p.NewBanList()
	if err := banList.Load(banListPath); err != nil {
		logging.Warn("failed to load ban list, starting fresh",
			logging.Err(err), logging.Component("daemon"))
	}
	node.Router().SetBanList(banList)

	// Create and load address book
	addressBook := p2p.NewAddressBook()
	if err := addressBook.Load(addressBookPath); err != nil {
		logging.Warn("failed to load address book, starting fresh",
			logging.Err(err), logging.Component("daemon"))
	}

	// Load cert pin store (already created inside node)
	if certPinStore := node.CertPinStore(); certPinStore != nil {
		if err := certPinStore.Load(certPinsPath); err != nil {
			logging.Warn("failed to load cert pin store, starting fresh",
				logging.Err(err), logging.Component("daemon"))
		}
	}

	// Start node (P2P layer)
	if err := node.Start(ctx); err != nil {
		log.Fatalf("Failed to start node: %v", err)
	}

	// C5: Start periodic cleanup goroutines
	if nonceTracker := node.NonceTracker(); nonceTracker != nil {
		util.SafeGoWithName("nonce-tracker-cleanup", func() {
			ticker := time.NewTicker(60 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					nonceTracker.CleanExpired()
				}
			}
		})
	}

	util.SafeGoWithName("banlist-cleanup", func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				banList.CleanExpired()
			}
		}
	})

	// Initialize payment service
	paymentCfg := &payment.PaymentServiceConfig{
		RPCURL:              cfg.Economics.RPCURL,
		WSEndpoint:          cfg.Economics.WSEndpoint,
		RPCURLs:             cfg.Economics.RPCURLs,
		WSEndpoints:         cfg.Economics.WSEndpoints,
		ChainID:             cfg.Economics.ChainID,
		BlockConfirmations:  cfg.Economics.BlockConfirmations,
		TokenAddress:        common.HexToAddress(cfg.Economics.TokenAddress),
		StakingAddress:      common.HexToAddress(cfg.Economics.StakingAddress),
		EscrowAddress:       common.HexToAddress(cfg.Economics.EscrowAddress),
		SlashingAddress:     common.HexToAddress(cfg.Economics.SlashingAddress),
		DelegationAddress:   common.HexToAddress(cfg.Economics.DelegationAddress),
		ReputationAddress:   common.HexToAddress(cfg.Economics.ReputationAddress),
		VerificationAddress: common.HexToAddress(cfg.Economics.VerificationAddress),
		PricingAddress:      common.HexToAddress(cfg.Economics.PricingAddress),
		PrivateKey:          walletPrivKey,
		MockMode:            cfg.Economics.MockPayments,
	}
	paymentSvc, err := payment.NewPaymentService(paymentCfg)
	if err != nil {
		log.Fatalf("Failed to create payment service: %v", err)
	}
	// H4: If production mode (MockPayments=false), payment failure is fatal.
	// If dev mode (MockPayments=true), silently fall back to mock.
	if err := paymentSvc.Start(ctx); err != nil {
		if !cfg.Economics.MockPayments {
			log.Fatalf("Failed to start payment service (production mode): %v", err)
		}
		logging.Warn("failed to start payment service, continuing in mock mode",
			logging.Err(err), logging.Component("daemon"))
		paymentCfg.MockMode = true
		paymentSvc, _ = payment.NewPaymentService(paymentCfg)
	}
	logging.Info("payment service initialized",
		"mock_mode", paymentCfg.MockMode,
		logging.Component("daemon"))

	// P1-11: Verify Go-side tier thresholds match deployed contract
	if !paymentCfg.MockMode {
		util.SafeGoWithName("tier-threshold-verify", func() {
			// Delay to let the node fully start
			select {
			case <-ctx.Done():
				return
			case <-time.After(10 * time.Second):
			}
			walletAddr := node.WalletAddress()
			if walletAddr == (common.Address{}) {
				return
			}
			verifyCtx, verifyCancel := context.WithTimeout(ctx, 15*time.Second)
			defer verifyCancel()
			stakeInfo, err := paymentSvc.GetStakeInfo(verifyCtx, walletAddr)
			if err != nil {
				logging.Warn("tier verification: failed to get stake info",
					logging.Err(err), logging.Component("daemon"))
				return
			}
			contractTier, err := paymentSvc.GetTier(verifyCtx, walletAddr)
			if err != nil {
				logging.Warn("tier verification: failed to get contract tier",
					logging.Err(err), logging.Component("daemon"))
				return
			}
			if stakeInfo.Tier != contractTier {
				logging.Error("TIER THRESHOLD MISMATCH: Go-side tier differs from contract",
					"go_tier", string(stakeInfo.Tier),
					"contract_tier", string(contractTier),
					"stake", stakeInfo.StakedAmount.String(),
					logging.Component("daemon"))
			} else {
				logging.Info("tier thresholds verified",
					"tier", string(stakeInfo.Tier),
					logging.Component("daemon"))
			}
		})
	}

	// Start event watcher for on-chain event-driven cache invalidation
	var eventWatcher *payment.EventWatcher
	if !paymentCfg.MockMode {
		eventWatcher = paymentSvc.NewEventWatcherFromService()
		if eventWatcher != nil {
			if err := eventWatcher.Start(ctx); err != nil {
				logging.Warn("failed to start event watcher, continuing without events",
					logging.Err(err), logging.Component("daemon"))
			} else {
				// Wire stake/slash events into the StakeVerifier for cache invalidation
				if sv := node.Router().StakeVerifier(); sv != nil {
					sv.StartEventListener(ctx, eventWatcher.StakeEvents(), eventWatcher.SlashEvents())
					logging.Info("event-driven stake cache invalidation enabled",
						logging.Component("daemon"))
				}
			}
		}
	}

	// Pass payment service to node for container manager
	node.SetPaymentService(paymentSvc)

	// Create and start API server for CLI communication
	apiServer := daemon.NewAPIServerWithFullConfig(node, cfg)
	if err := apiServer.Start(ctx); err != nil {
		log.Fatalf("Failed to start API server: %v", err)
	}

	// Start disk usage enforcer (monitors container writable layer, stops if over limit)
	if cm := apiServer.GetContainerManager(); cm != nil {
		cm.StartDiskEnforcer(ctx, 60*time.Second)
	}

	// P1-5: Start certificate rotator for automatic TLS cert renewal
	certDir := filepath.Join(cfg.Daemon.DataDir, "certs")
	os.MkdirAll(certDir, 0700)
	certRotator := identity.NewCertRotator(
		node.KeyManager(),
		filepath.Join(certDir, "node.crt"),
		filepath.Join(certDir, "node.key"),
	)
	// P1-12: Wire UpdateIdentity callback — when the TLS cert rotates, the NodeID
	// changes (NodeID = SHA256(SPKI)), so we update the on-chain identity to match.
	if !paymentCfg.MockMode {
		certRotator.SetOnRotation(func(cert *tls.Certificate) {
			if cert.Leaf == nil {
				return
			}
			newNodeID := sha256.Sum256(cert.Leaf.RawSubjectPublicKeyInfo)
			// Use the node's configured region; capabilities = 0 (unchanged)
			var region [32]byte
			copy(region[:], []byte(cfg.Node.Region))
			updateCtx, updateCancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer updateCancel()
			if err := paymentSvc.UpdateIdentity(updateCtx, newNodeID, region, 0); err != nil {
				logging.Warn("failed to update on-chain identity after cert rotation",
					logging.Err(err), logging.Component("cert-rotation"))
			} else {
				logging.Info("on-chain identity updated after cert rotation",
					"new_node_id", fmt.Sprintf("%x", newNodeID[:8]),
					logging.Component("cert-rotation"))
			}
		})
	}
	certRotator.Start(ctx)
	logging.Info("certificate rotator started", logging.Component("daemon"))

	// P1-9: Start periodic goroutine count tracking
	util.SafeGoWithName("goroutine-tracker", func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				count := goruntime.NumGoroutine()
				if count > 1000 {
					logging.Warn("high goroutine count",
						"count", count,
						logging.Component("daemon"))
				}
			}
		}
	})

	// Initialize threat detection system
	threatDetector := threat.NewDetector(threat.DefaultDetectorConfig())
	if err := threatDetector.Start(ctx); err != nil {
		logging.Warn("failed to start threat detector", logging.Err(err), logging.Component("daemon"))
	}

	// Initialize snapshot manager
	snapshotCfg := snapshot.DefaultSnapshotConfig()
	snapshotCfg.StoragePath = filepath.Join(cfg.Daemon.DataDir, "snapshots")
	snapshotMgr, err := snapshot.NewManager(snapshotCfg)
	if err != nil {
		logging.Warn("failed to initialize snapshot manager", logging.Err(err), logging.Component("daemon"))
	}

	// Initialize checkpoint system
	var checkpointer *snapshot.Checkpointer
	if snapshotMgr != nil {
		checkpointer = snapshot.NewCheckpointer(snapshotMgr, snapshot.DefaultCheckpointConfig())
		if err := checkpointer.Start(ctx); err != nil {
			logging.Warn("failed to start checkpointer", logging.Err(err), logging.Component("daemon"))
		}
	}

	// Initialize cloning manager
	cloningCfg := cloning.DefaultCloneConfig()
	cloningMgr := cloning.NewManager(cloningCfg, snapshotMgr, nil, nil)
	if err := cloningMgr.Start(ctx); err != nil {
		logging.Warn("failed to start cloning manager", logging.Err(err), logging.Component("daemon"))
	}

	// Wire up threat-triggered cloning
	threatResponder := threat.NewResponder(threatDetector, threat.DefaultResponseConfig())
	cloningMgr.SetThreatDetector(threatDetector, threatResponder)
	threatResponder.Start(ctx)

	// Start embedded HTTP API server (serves web UI, exec terminal, admin API)
	// This runs in the same process as the daemon so exec has direct ContainerManager access.
	var httpAPIServer *api.Server
	if *httpAddr != "" {
		httpServerCfg := &api.ServerConfig{
			HTTPAddr:         *httpAddr,
			DaemonSocketPath: cfg.Daemon.SocketPath,
			DaemonPoolSize:   4,
			RateLimit:        cfg.API.RateLimitRequests,
			RateLimitBurst:   20,
			EnableAuth:       true,
			APIKeyHeader:     "X-API-Key",
			EnableCORS:       true,
			AllowedOrigins:   []string{"*"},
			ReadTimeout:      time.Duration(cfg.API.ReadTimeoutSecs) * time.Second,
			WriteTimeout:     time.Duration(cfg.API.WriteTimeoutSecs) * time.Second,
			IdleTimeout:      time.Duration(cfg.API.IdleTimeoutSecs) * time.Second,
			EnableWebSocket:  true,
			APIKeyStorePath:  filepath.Join(cfg.Daemon.DataDir, "api_keys.json"),
		}

		httpAPIServer = api.NewServer(httpServerCfg)
		httpAPIServer.SetFullConfig(cfg)
		httpAPIServer.SetDaemonAPI(apiServer) // Direct ContainerManager access for exec
		httpAPIServer.SetThreatDetector(threatDetector)
		httpAPIServer.SetSnapshotManager(snapshotMgr)
		httpAPIServer.SetCloningManager(cloningMgr)
		httpAPIServer.SetAdminStore(api.NewAdminMetadataStore(filepath.Join(cfg.Daemon.DataDir, "admin_metadata.json")))
		httpAPIServer.SetPolicyStore(api.NewPolicyStore(filepath.Join(cfg.Daemon.DataDir, "admin_policies.json")))
		httpAPIServer.SetCatalogStore(api.NewCatalogStore(filepath.Join(cfg.Daemon.DataDir, "catalog.json")))

		if err := httpAPIServer.Start(ctx); err != nil {
			log.Fatalf("Failed to start HTTP API server: %v", err)
		}
		logging.Info("HTTP API server started (exec enabled)",
			"addr", *httpAddr,
			logging.Component("daemon"))
	}

	logging.Info("daemon started",
		"p2p_port", cfg.Daemon.Port,
		"node_id", node.NodeInfo().ID.String(),
		"api_socket", cfg.Daemon.SocketPath,
		"http_api", *httpAddr,
		"data_dir", cfg.Daemon.DataDir,
		"config", cfgPath,
		"network", cfg.P2P.NetworkMode,
		logging.Component("daemon"))
	if cfg.Tor.Enabled {
		logging.Info("tor enabled", logging.Component("daemon"))
	}
	if len(cfg.P2P.BootstrapNodes) > 0 {
		logging.Info("bootstrap nodes configured", "count", len(cfg.P2P.BootstrapNodes), logging.Component("daemon"))
	}

	// Wait for signal
	sig := <-sigChan
	logging.Info("received signal, shutting down gracefully...", "signal", sig.String(), logging.Component("daemon"))

	// Cancel the main context to propagate shutdown to all goroutines
	cancel()

	// Create shutdown context with timeout for cleanup operations
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Run shutdown in a goroutine so we can enforce the timeout with a force exit
	shutdownDone := make(chan struct{})
	go func() {
		defer close(shutdownDone)

		// Stop HTTP API server first (stop accepting web requests)
		if httpAPIServer != nil {
			logging.Info("stopping HTTP API server...", logging.Component("daemon"))
			if err := httpAPIServer.Stop(shutdownCtx); err != nil {
				logging.Error("error stopping HTTP API server", logging.Err(err), logging.Component("daemon"))
			}
		}

		// C8: Stop Unix socket API server (stop accepting CLI requests)
		logging.Info("stopping API server...", logging.Component("daemon"))
		if err := apiServer.Stop(); err != nil {
			logging.Error("error stopping API server", logging.Err(err), logging.Component("daemon"))
		}

		// Stop event watcher (before payment service so WS subscriptions close cleanly)
		if eventWatcher != nil {
			logging.Info("stopping event watcher...", logging.Component("daemon"))
			eventWatcher.Stop()
		}

		// Stop certificate rotator
		logging.Info("stopping certificate rotator...", logging.Component("daemon"))
		certRotator.Stop()

		// Stop payment service
		logging.Info("stopping payment service...", logging.Component("daemon"))
		paymentSvc.Stop()

		// Stop new systems
		logging.Info("stopping threat responder...", logging.Component("daemon"))
		threatResponder.Stop()

		logging.Info("stopping threat detector...", logging.Component("daemon"))
		threatDetector.Stop()

		logging.Info("stopping cloning manager...", logging.Component("daemon"))
		cloningMgr.Stop()

		if checkpointer != nil {
			logging.Info("stopping checkpointer...", logging.Component("daemon"))
			checkpointer.Stop()
		}

		// C1: Save persisted state before stopping node
		logging.Info("saving persisted state...", logging.Component("daemon"))
		if err := banList.Save(banListPath); err != nil {
			logging.Error("failed to save ban list", logging.Err(err), logging.Component("daemon"))
		}
		if err := addressBook.Save(addressBookPath); err != nil {
			logging.Error("failed to save address book", logging.Err(err), logging.Component("daemon"))
		}
		if certPinStore := node.CertPinStore(); certPinStore != nil {
			if err := certPinStore.Save(certPinsPath); err != nil {
				logging.Error("failed to save cert pin store", logging.Err(err), logging.Component("daemon"))
			}
		}

		// Gracefully shutdown node (closes listener, router, DHT)
		logging.Info("stopping P2P node...", logging.Component("daemon"))
		if err := node.Shutdown(shutdownCtx); err != nil {
			logging.Error("error during node shutdown", logging.Err(err), logging.Component("daemon"))
		}

		// M1: Remove PID file
		os.Remove(pidPath)
	}()

	// Wait for shutdown to complete or timeout to expire
	select {
	case <-shutdownDone:
		logging.Info("daemon stopped", logging.Component("daemon"))
	case <-shutdownCtx.Done():
		logging.Error("shutdown timed out after 30 seconds, forcing exit", logging.Component("daemon"))
		os.Exit(1)
	}
}

// loadWallet loads and unlocks the wallet. The daemon requires an unlocked wallet
// to operate — it's needed for signing P2P messages, announce protocol, staking
// verification, and payment transactions regardless of role.
// Fatals if the wallet is missing, password unavailable, or password wrong.
func loadWallet(cfg *config.Config, node *daemon.Node) *ecdsa.PrivateKey {
	wm := node.WalletManager()
	if wm == nil {
		log.Fatalf("no wallet found — the daemon requires a wallet to operate. " +
			"Create one with: moltbunker wallet create")
	}

	logging.Info("wallet loaded",
		"address", wm.Address().Hex(),
		"keystore", cfg.Daemon.KeystoreDir,
		logging.Component("daemon"))

	// Resolve wallet password
	password, found := resolveWalletPassword(cfg)
	if !found {
		log.Fatalf("wallet password not available — the daemon cannot unlock the wallet. " +
			"Store password with: moltbunker wallet create (saves to keyring), " +
			"or set MOLTBUNKER_WALLET_PASSWORD env var, " +
			"or set node.wallet_password_file in config.yaml")
	}

	// Unlock the wallet
	privKey, err := wm.PrivateKey(password)
	if err != nil {
		log.Fatalf("failed to unlock wallet (wrong password?): %v — "+
			"check your keyring or MOLTBUNKER_WALLET_PASSWORD", err)
	}

	logging.Info("wallet unlocked",
		"address", wm.Address().Hex(),
		logging.Component("daemon"))

	return privKey
}

// resolveWalletPassword resolves the wallet password from the best available source.
// Priority: env var → password file → platform keyring → kernel keyring.
// Env var is checked first because the CLI pre-resolves the password from the
// keyring and passes it via MOLTBUNKER_WALLET_PASSWORD to avoid duplicate
// macOS Keychain dialogs from the daemon process.
// Returns (password, true) if a source was found, or ("", false) if no source configured.
func resolveWalletPassword(cfg *config.Config) (string, bool) {
	// 1. Environment variable (set by CLI start, or by user directly)
	if pw := os.Getenv("MOLTBUNKER_WALLET_PASSWORD"); pw != "" {
		// Clear from environment to reduce exposure in /proc/PID/environ
		os.Unsetenv("MOLTBUNKER_WALLET_PASSWORD")
		logging.Info("wallet password from environment",
			logging.Component("daemon"))
		return pw, true
	}

	// 2. Password file from config
	if cfg.Node.WalletPasswordFile != "" {
		data, err := os.ReadFile(cfg.Node.WalletPasswordFile)
		if err != nil {
			logging.Warn("failed to read wallet password file",
				"path", cfg.Node.WalletPasswordFile,
				logging.Err(err),
				logging.Component("daemon"))
			return "", false
		}
		logging.Info("wallet password from file",
			"path", cfg.Node.WalletPasswordFile,
			logging.Component("daemon"))
		return strings.TrimSpace(string(data)), true
	}

	// 3. Platform keyring (macOS Keychain, Linux Secret Service)
	if pw, err := identity.RetrieveWalletPassword(); err == nil && pw != "" {
		logging.Info("wallet password from platform keyring",
			logging.Component("daemon"))
		return pw, true
	}

	// 4. Linux kernel keyring (headless servers)
	if pw, err := identity.RetrieveKernelKeyring(); err == nil && pw != "" {
		logging.Info("wallet password from kernel keyring",
			logging.Component("daemon"))
		return pw, true
	}

	return "", false
}

// checkAndWritePIDFile checks for stale PID files and writes the current PID.
// Returns an error if another daemon process is already running.
func checkAndWritePIDFile(pidPath string) error {
	// Check for existing PID file
	data, err := os.ReadFile(pidPath)
	if err == nil {
		// PID file exists, check if process is still running
		pid, parseErr := strconv.Atoi(string(data))
		if parseErr == nil && pid > 0 {
			process, findErr := os.FindProcess(pid)
			if findErr == nil {
				// On Unix, FindProcess always succeeds. Send signal 0 to check if alive.
				if err := process.Signal(syscall.Signal(0)); err == nil {
					return fmt.Errorf("another daemon is already running (pid %d)", pid)
				}
			}
		}
		// Stale PID file, remove it
		logging.Warn("removing stale PID file",
			"path", pidPath,
			"stale_pid", string(data),
			logging.Component("daemon"))
		os.Remove(pidPath)
	}

	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(pidPath), 0700); err != nil {
		return fmt.Errorf("create PID file directory: %w", err)
	}

	// Write current PID
	pidData := []byte(strconv.Itoa(os.Getpid()))
	if err := os.WriteFile(pidPath, pidData, 0600); err != nil {
		return fmt.Errorf("write PID file: %w", err)
	}

	return nil
}
