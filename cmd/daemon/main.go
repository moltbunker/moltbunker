package main

import (
	"context"
	"crypto/ecdsa"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
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
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/moltbunker/moltbunker/internal/agent"
	"github.com/moltbunker/moltbunker/internal/api"
	"github.com/moltbunker/moltbunker/internal/cloning"
	"github.com/moltbunker/moltbunker/internal/config"
	"github.com/moltbunker/moltbunker/internal/crawl"
	"github.com/moltbunker/moltbunker/internal/daemon"
	"github.com/moltbunker/moltbunker/internal/identity"
	"github.com/moltbunker/moltbunker/internal/ingress"
	"github.com/moltbunker/moltbunker/internal/logging"
	"github.com/moltbunker/moltbunker/internal/p2p"
	"github.com/moltbunker/moltbunker/internal/payment"
	"github.com/moltbunker/moltbunker/internal/proxy"
	"github.com/moltbunker/moltbunker/internal/snapshot"
	"github.com/moltbunker/moltbunker/internal/state"
	"github.com/moltbunker/moltbunker/internal/storage"
	"github.com/moltbunker/moltbunker/internal/threat"
	"github.com/moltbunker/moltbunker/internal/tunnel"
	"github.com/moltbunker/moltbunker/internal/util"
)

// Build-time version information (set via -ldflags)
var (
	version   = "dev"
	commit    = "unknown"
	buildDate = "unknown"
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

	log.Printf("moltbunkerd version=%s commit=%s built=%s", version, commit, buildDate)

	// Load configuration
	cfgPath := *configPath
	if cfgPath == "" {
		cfgPath = config.DefaultConfigPath()
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Wrap the structured logger with the redacting handler before any
	// subsystem starts emitting attributes. Anything that looks like a
	// private key, wallet keystore JSON, API key, or session token is
	// scrubbed before reaching the underlying handler.
	logging.EnableRedaction()

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

	// C1: Open persistent state database (bbolt)
	stateDBPath := cfg.Daemon.StateDBPath
	if stateDBPath == "" {
		stateDBPath = filepath.Join(cfg.Daemon.DataDir, "moltbunker.db")
	}

	// R8: state-at-rest encryption. Enabled by default (zero value of
	// StateEncryptionDisabled). Mitigates stolen-disk / leaked-backup / casual
	// filesystem access to moltbunker.db; it does NOT defend against a live
	// host-root attacker who can also read DataDir/state.key (that needs
	// SEV-SNP / TPM). Key-load failure is FATAL: falling back to a nil key would
	// silently mis-read existing encrypted values as garbage and write new
	// values in plaintext, so we fail closed and let the operator fix the key
	// (or set security.state_encryption_disabled to opt out deliberately).
	var stateEncKey []byte
	if !cfg.Security.StateEncryptionDisabled {
		k, kerr := state.LoadOrCreateStateKey(cfg.Daemon.DataDir)
		if kerr != nil {
			log.Fatalf("Failed to load state encryption key (set security.state_encryption_disabled to run without it): %v", kerr)
		}
		stateEncKey = k
		logging.Info("state-at-rest encryption enabled", logging.Component("daemon"))
	} else {
		logging.Info("state-at-rest encryption disabled", logging.Component("daemon"))
	}

	stateStore, err := state.NewBboltStore(stateDBPath, stateEncKey)
	if err != nil {
		log.Fatalf("Failed to open state database: %v", err)
	}
	defer stateStore.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Run JSON → bbolt migration (no-op if already migrated or no JSON files)
	if err := state.MigrateFromJSON(ctx, stateStore, cfg.Daemon.DataDir); err != nil {
		logging.Warn("state migration failed, continuing",
			logging.Err(err), logging.Component("daemon"))
	}

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
		RPCURL:                   cfg.Economics.RPCURL,
		WSEndpoint:               cfg.Economics.WSEndpoint,
		RPCURLs:                  cfg.Economics.RPCURLs,
		WSEndpoints:              cfg.Economics.WSEndpoints,
		ChainID:                  cfg.Economics.ChainID,
		BlockConfirmations:       cfg.Economics.BlockConfirmations,
		TokenAddress:             common.HexToAddress(cfg.Economics.TokenAddress),
		StakingAddress:           common.HexToAddress(cfg.Economics.StakingAddress),
		EscrowAddress:            common.HexToAddress(cfg.Economics.EscrowAddress),
		SlashingAddress:          common.HexToAddress(cfg.Economics.SlashingAddress),
		DelegationAddress:        common.HexToAddress(cfg.Economics.DelegationAddress),
		ReputationAddress:        common.HexToAddress(cfg.Economics.ReputationAddress),
		VerificationAddress:      common.HexToAddress(cfg.Economics.VerificationAddress),
		PricingAddress:           common.HexToAddress(cfg.Economics.PricingAddress),
		SubdomainRegistryAddress: common.HexToAddress(cfg.Economics.SubdomainRegistryAddress),
		PrivateKey:               walletPrivKey,
		MockMode:                 cfg.Economics.MockPayments,
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

	// Register provider in reputation system if not yet registered.
	// Must happen before event watcher so reputation calls succeed.
	if !paymentCfg.MockMode {
		util.SafeGoWithName("reputation-register", func() {
			select {
			case <-ctx.Done():
				return
			case <-time.After(15 * time.Second): // wait for staking info to be available
			}
			walletAddr := node.WalletAddress()
			if walletAddr == (common.Address{}) {
				return
			}
			regCtx, regCancel := context.WithTimeout(ctx, 15*time.Second)
			defer regCancel()
			if err := paymentSvc.RegisterProviderReputation(regCtx, walletAddr); err != nil {
				logging.Debug("provider reputation registration skipped or failed",
					logging.Err(err), logging.Component("daemon"))
			} else {
				logging.Info("provider registered in reputation system",
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
	apiServer.SetStateStore(stateStore)
	if err := apiServer.Start(ctx); err != nil {
		log.Fatalf("Failed to start API server: %v", err)
	}

	// Start disk usage enforcer (monitors container writable layer, stops if over limit)
	if cm := apiServer.GetContainerManager(); cm != nil {
		cm.StartDiskEnforcer(ctx, 60*time.Second)
	}

	// Escrow event consumer: drive local reservation-ID cache invalidation off
	// the on-chain escrow lifecycle. On Refunded/Finalized, the reservation has
	// reached a terminal state, so the jobID→reservationID mapping must be
	// dropped — otherwise a redeploy with the same jobID would reuse a stale
	// on-chain reservation. The events only carry the reservationID, so we
	// reverse-resolve the jobID from the payment service's cache.
	if eventWatcher != nil {
		util.SafeGoWithName("escrow-event-consumer", func() {
			for ev := range eventWatcher.EscrowEvents() {
				switch ev.Kind {
				case payment.EscrowEventRefunded, payment.EscrowEventFinalized:
					if jobID, ok := paymentSvc.JobIDForReservationID(ev.ReservationID); ok {
						paymentSvc.InvalidateEscrowReservation(jobID)
						logging.Debug("escrow event: invalidated reservation cache",
							"kind", ev.Kind.String(),
							logging.Component("daemon"))
					}
				}
			}
		})
	}

	// Start subdomain expiry cleanup (removes expired gossip entries hourly)
	if cm := apiServer.GetContainerManager(); cm != nil && cm.GossipProtocol() != nil {
		daemon.StartSubdomainCleanup(ctx, cm.GossipProtocol(), paymentSvc)
	}

	// ── Ingress + Tunnel wiring ──

	// Provider nodes: start tunnel server so ingress nodes can proxy traffic to containers
	var tunnelSrv *tunnel.Server
	if cfg.IsProvider() {
		cm := apiServer.GetContainerManager()
		if cm != nil && cm.NetworkManager() != nil {
			tunnelPort := cfg.Node.Provider.TunnelPort
			if tunnelPort == 0 {
				tunnelPort = cfg.Daemon.Port + 2 // Convention: base+1=TLS P2P, base+2=tunnel
			}
			portResolver := daemon.NewDeploymentPortResolver(cm.NetworkManager())
			tunnelListener, listenErr := tls.Listen("tcp", fmt.Sprintf(":%d", tunnelPort), node.TLSServerConfig())
			if listenErr != nil {
				logging.Warn("failed to start tunnel server",
					"port", tunnelPort,
					logging.Err(listenErr),
					logging.Component("tunnel"))
			} else {
				tunnelSrv = tunnel.NewServer(tunnelListener, portResolver)
				util.SafeGoWithName("tunnel-server", func() {
					if srvErr := tunnelSrv.Serve(ctx); srvErr != nil && ctx.Err() == nil {
						logging.Error("tunnel server error",
							logging.Err(srvErr),
							logging.Component("tunnel"))
					}
				})
				logging.Info("tunnel server started",
					"port", tunnelPort,
					logging.Component("tunnel"))
			}
		}
	}

	// Ingress nodes: start HTTP reverse proxy for subdomain routing
	var ingressProxy *ingress.Proxy
	var ingressHealthChecker *ingress.HealthChecker
	if cfg.Node.Provider.IngressEnabled {
		cm := apiServer.GetContainerManager()
		if cm != nil && cm.GossipProtocol() != nil {
			ingressPort := cfg.Node.Provider.IngressPort
			if ingressPort == 0 {
				ingressPort = 9090
			}
			ingressDomain := cfg.Node.Provider.IngressDomain
			if ingressDomain == "" {
				ingressDomain = "moltbunker.dev"
			}

			// Set gossip state validator to prevent expose: key poisoning
			cm.GossipProtocol().SetStateValidator(daemon.NewGossipStateValidator(node.NodeInfo().ID))
			gossipAdapter := daemon.NewGossipServiceAdapter(cm.GossipProtocol())
			gossipAdapter.SetPaymentService(paymentSvc) // Enable on-chain subdomain resolution fallback
			tunnelDialer := daemon.NewTLSTunnelDialer(node.TLSClientConfig())
			tunnelClient := tunnel.NewClient(tunnelDialer)
			resolver := ingress.NewResolver(gossipAdapter, gossipAdapter) // implements both GossipReader and SubdomainResolver
			ingressProxy = ingress.NewProxy(resolver, tunnelClient, ingressDomain)

			// Wire Cloudflare DNS sync if configured
			if cfg.Node.Provider.CloudflareAPIToken != "" && cfg.Node.Provider.CloudflareZoneID != "" && cfg.Node.Provider.IngressIP != "" {
				dnsSync := ingress.NewDNSSync(
					cfg.Node.Provider.CloudflareAPIToken,
					cfg.Node.Provider.CloudflareZoneID,
					cfg.Node.Provider.IngressIP,
					ingressDomain,
				)
				apiServer.SetDNSSync(dnsSync)
				logging.Info("cloudflare DNS sync enabled",
					"domain", ingressDomain,
					logging.Component("ingress"))
			}

			// TLS configuration: use Let's Encrypt autocert if enabled, else node self-signed cert
			var ingressTLSCfg *tls.Config
			if cfg.Node.Provider.IngressAutoTLS {
				certDir := cfg.Node.Provider.IngressCertDir
				if certDir == "" {
					certDir = filepath.Join(cfg.Daemon.DataDir, "ingress-certs")
				}
				email := cfg.Node.Provider.IngressACMEEmail
				autoTLS := ingress.NewAutoTLSConfig(certDir, ingressDomain, email, resolver)
				ingressTLSCfg = autoTLS.TLSConfig()
				logging.Info("ingress auto-TLS enabled (Let's Encrypt)",
					"cert_dir", certDir,
					"domain", ingressDomain,
					logging.Component("ingress"))
			} else {
				// Existing: node self-signed cert
				ingressTLSCfg = node.TLSServerConfig()
				ingressTLSCfg.ClientAuth = tls.NoClientCert // Public clients don't present certs
			}

			ingressListener, listenErr := tls.Listen("tcp", fmt.Sprintf(":%d", ingressPort), ingressTLSCfg)
			if listenErr != nil {
				logging.Warn("failed to start ingress proxy",
					"port", ingressPort,
					logging.Err(listenErr),
					logging.Component("ingress"))
			} else {
				util.SafeGoWithName("ingress-proxy", func() {
					if srvErr := ingressProxy.Serve(ingressListener); srvErr != nil && ctx.Err() == nil {
						logging.Error("ingress proxy error",
							logging.Err(srvErr),
							logging.Component("ingress"))
					}
				})

				// Start health checker for exposed services
				ingressHealthChecker = ingress.NewHealthChecker(resolver, tunnelClient)
				ingressHealthChecker.Start(ctx)

				logging.Info("ingress proxy started",
					"port", ingressPort,
					"domain", ingressDomain,
					logging.Component("ingress"))

				// Reverse tunnel server (ingress-side): accept connections from NAT'd providers
				if cfg.Node.Provider.ReverseTunnelPort > 0 {
					revTLSCfg := node.TLSServerConfig()
					revTLSCfg.ClientAuth = tls.RequireAnyClientCert // Providers must present a cert
					revPort := cfg.Node.Provider.ReverseTunnelPort
					revListener, revErr := tls.Listen("tcp", fmt.Sprintf(":%d", revPort), revTLSCfg)
					if revErr != nil {
						logging.Warn("failed to start reverse tunnel server",
							"port", revPort,
							logging.Err(revErr),
							logging.Component("reverse-tunnel"))
					} else {
						revOpts := []tunnel.ReverseServerOption{
							tunnel.WithDomain(ingressDomain),
							tunnel.WithWalletVerifier(func(proof *tunnel.WalletProof, nodeID string) (string, error) {
								if proof == nil || proof.Address == "" || proof.Signature == "" {
									return "", fmt.Errorf("incomplete wallet proof")
								}
								// Verify EIP-191 signature: message must bind nodeID to wallet
								walletAddr := common.HexToAddress(proof.Address)
								msgHash := p2p.EthPersonalHash(proof.Message)
								sigBytes, err := hex.DecodeString(strings.TrimPrefix(proof.Signature, "0x"))
								if err != nil || len(sigBytes) != 65 {
									return "", fmt.Errorf("invalid signature format")
								}
								sigForRecovery := make([]byte, 65)
								copy(sigForRecovery, sigBytes)
								if sigForRecovery[64] >= 27 {
									sigForRecovery[64] -= 27
								}
								pubKey, err := crypto.SigToPub(msgHash, sigForRecovery)
								if err != nil {
									return "", fmt.Errorf("signature recovery failed: %w", err)
								}
								recovered := crypto.PubkeyToAddress(*pubKey)
								if recovered != walletAddr {
									return "", fmt.Errorf("wallet address mismatch")
								}
								// Check on-chain stake tier
								checkCtx, checkCancel := context.WithTimeout(ctx, 5*time.Second)
								defer checkCancel()
								tierVal, err := paymentSvc.GetTier(checkCtx, walletAddr)
								if err != nil {
									return "", fmt.Errorf("stake check failed: %w", err)
								}
								return string(tierVal), nil
							}),
						}
						if cfg.Node.Provider.ReverseTunnelMaxConns > 0 {
							revOpts = append(revOpts, tunnel.WithMaxConns(cfg.Node.Provider.ReverseTunnelMaxConns))
						}
						reverseServer := tunnel.NewReverseServer(revListener, revOpts...)
						ingressProxy.SetReverseStreamOpener(reverseServer)
						util.SafeGoWithName("reverse-tunnel-server", func() {
							if srvErr := reverseServer.Serve(ctx); srvErr != nil && ctx.Err() == nil {
								logging.Error("reverse tunnel server error",
									logging.Err(srvErr),
									logging.Component("reverse-tunnel"))
							}
						})
						logging.Info("reverse tunnel server started",
							"port", revPort,
							"domain", ingressDomain,
							logging.Component("reverse-tunnel"))
					}
				}
			}
		}
	}

	// Provider-side reverse tunnel: wire into ContainerManager so deployments
	// with exposed ports automatically get a reverse tunnel to the ingress.
	if cfg.Node.Provider.ReverseTunnelEnabled && cfg.Node.Provider.ReverseTunnelIngress != "" {
		cm := apiServer.GetContainerManager()
		if cm != nil && cm.NetworkManager() != nil {
			portResolver := daemon.NewDeploymentPortResolver(cm.NetworkManager())
			revTLSCfg := node.TLSClientConfig()
			ingressAddr := cfg.Node.Provider.ReverseTunnelIngress

			// Factory creates a new ReverseClient per deployment+port
			clientFactory := func() *tunnel.ReverseClient {
				return tunnel.NewReverseClient(ingressAddr, portResolver, revTLSCfg)
			}

			rtm := daemon.NewReverseTunnelManager(ctx, clientFactory)
			cm.SetReverseTunnelManager(rtm)

			logging.Info("reverse tunnel manager enabled",
				"ingress", ingressAddr,
				logging.Component("reverse-tunnel"))
		}
	}

	// P1-5: Start certificate rotator for automatic TLS cert renewal
	certDir := filepath.Join(cfg.Daemon.DataDir, "certs")
	if err := os.MkdirAll(certDir, 0700); err != nil {
		log.Fatalf("failed to create cert directory %s: %v", certDir, err)
	}
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
		// #nosec G101 -- not a credential: config-struct literal (addr/header-name/timeout fields), no secrets
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

		// ── P0 Services ──

		// Object Storage
		if cfg.Storage.Enabled {
			storageDataDir := cfg.Storage.DataDir
			if storageDataDir == "" {
				storageDataDir = filepath.Join(cfg.Daemon.DataDir, "storage")
			}
			storageEngine, storageErr := storage.NewStorageEngine(storageDataDir, stateStore, storage.EngineConfig{
				MaxBuckets:    cfg.Storage.MaxBuckets,
				MaxObjectSize: cfg.Storage.MaxObjectSize,
			})
			if storageErr != nil {
				log.Fatalf("Failed to create storage engine: %v", storageErr)
			}
			// Wire storage usage metering for billing (PaymentService satisfies
			// storage.MeteringHook structurally).
			storageEngine.SetMeteringHook(paymentSvc)
			httpAPIServer.SetStorageHandler(storage.NewRESTHandler(storageEngine))
			logging.Info("object storage service enabled", logging.Component("daemon"))
		}

		// Decentralized Proxy
		if cfg.Proxy.Enabled {
			proxyServer := proxy.NewServer(proxy.Config{
				SOCKS5Addr:  cfg.Proxy.SOCKS5Addr,
				HTTPAddr:    cfg.Proxy.HTTPAddr,
				UseTor:      cfg.Proxy.UseTor,
				MaxSessions: cfg.Proxy.MaxSessions,
			}, &proxy.DirectDialer{}, &proxy.AllowAllAuth{DefaultWallet: "system"})
			// Wire proxy session metering before Start so it propagates to the
			// SOCKS5/HTTP sub-servers (PaymentService satisfies proxy.ProxyMeteringHook).
			proxyServer.SetMeteringHook(paymentSvc)
			if proxyErr := proxyServer.Start(ctx); proxyErr != nil {
				log.Fatalf("Failed to start proxy service: %v", proxyErr)
			}
			httpAPIServer.SetProxyHandler(proxy.NewRESTHandler(proxyServer))
			// Store reference for shutdown (captured in shutdown goroutine closure)
			defer func() {
				if err := proxyServer.Stop(); err != nil {
					logging.Warn("failed to stop proxy server",
						logging.Err(err), logging.Component("daemon"))
				}
			}()
			logging.Info("proxy service enabled",
				"socks5", cfg.Proxy.SOCKS5Addr,
				"http", cfg.Proxy.HTTPAddr,
				logging.Component("daemon"))
		}

		// Web Crawling
		if cfg.Crawl.Enabled {
			scheduler := crawl.NewScheduler(crawl.SchedulerConfig{
				MaxConcurrentJobs: cfg.Crawl.MaxConcurrent,
				MaxPagesPerJob:    cfg.Crawl.MaxPages,
			})
			// Wire crawl job metering (PaymentService satisfies crawl.CrawlMeteringHook).
			scheduler.SetMeteringHook(paymentSvc)
			httpAPIServer.SetCrawlHandler(crawl.NewRESTHandler(scheduler, crawl.NewRobotsChecker()))
			logging.Info("web crawling service enabled", logging.Component("daemon"))
		}

		// AI Agent Runtime
		if cfg.Agent.Enabled {
			agentRuntime := agent.NewAgentRuntime(agent.RuntimeConfig{
				MaxAgentsPerWallet: cfg.Agent.MaxAgentsPerWallet,
			})
			agentHandler := agent.NewRESTHandler(agentRuntime, agent.NewMemoryStore())
			// Wire agent invocation metering (PaymentService satisfies agent.AgentMeteringHook).
			agentHandler.SetMeteringHook(paymentSvc)
			httpAPIServer.SetAgentHandler(agentHandler)
			logging.Info("agent runtime service enabled", logging.Component("daemon"))
		}

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

		// Stop ingress proxy and health checker first (edge traffic)
		if ingressHealthChecker != nil {
			logging.Info("stopping ingress health checker...", logging.Component("daemon"))
			ingressHealthChecker.Stop()
		}
		if ingressProxy != nil {
			logging.Info("stopping ingress proxy...", logging.Component("daemon"))
			if err := ingressProxy.Shutdown(shutdownCtx); err != nil {
				logging.Error("error stopping ingress proxy", logging.Err(err), logging.Component("daemon"))
			}
		}

		// Stop tunnel server (provider-side)
		if tunnelSrv != nil {
			logging.Info("stopping tunnel server...", logging.Component("daemon"))
			if err := tunnelSrv.Close(); err != nil {
				logging.Error("error stopping tunnel server", logging.Err(err), logging.Component("daemon"))
			}
		}

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

		// Finalize escrows for running containers before payment service shuts down.
		// This prevents escrows from being stranded if the daemon never restarts.
		logging.Info("finalizing escrows for running containers...", logging.Component("daemon"))
		if cm := apiServer.GetContainerManager(); cm != nil {
			cm.FinalizeAllEscrows(shutdownCtx)
		}

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

		// Broadcast gossip leave so peers stop routing traffic to us
		if cm := apiServer.GetContainerManager(); cm != nil {
			if gp := cm.GossipProtocol(); gp != nil {
				gp.RemoveLocalState()
				logging.Info("gossip leave broadcast sent", logging.Component("daemon"))
			}
		}

		// Gracefully shutdown node (closes listener, router, DHT)
		logging.Info("stopping P2P node...", logging.Component("daemon"))
		if err := node.Shutdown(shutdownCtx); err != nil {
			logging.Error("error during node shutdown", logging.Err(err), logging.Component("daemon"))
		}

		// M1: Remove PID file
		_ = os.Remove(pidPath)
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
	// #nosec G304 -- pidPath is the daemon's configured PID file path (DataDir-derived), not request input
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
		_ = os.Remove(pidPath)
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
