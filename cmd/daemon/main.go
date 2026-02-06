package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/moltbunker/moltbunker/internal/cloning"
	"github.com/moltbunker/moltbunker/internal/config"
	"github.com/moltbunker/moltbunker/internal/daemon"
	"github.com/moltbunker/moltbunker/internal/snapshot"
	"github.com/moltbunker/moltbunker/internal/threat"
)

var (
	configPath  = flag.String("config", "", "Path to config file (default: ~/.moltbunker/config.yaml)")
	port        = flag.Int("port", 0, "P2P port (overrides config)")
	keyPath     = flag.String("key", "", "Path to node key (overrides config)")
	keystoreDir = flag.String("keystore", "", "Path to keystore (overrides config)")
	dataDir     = flag.String("data", "", "Path to data directory (overrides config)")
	socketPath  = flag.String("socket", "", "Unix socket path for API (overrides config)")
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

	// Start node (P2P layer)
	if err := node.Start(ctx); err != nil {
		log.Fatalf("Failed to start node: %v", err)
	}

	// Create and start API server for CLI communication
	apiServer := daemon.NewAPIServerWithFullConfig(node, cfg)
	if err := apiServer.Start(ctx); err != nil {
		log.Fatalf("Failed to start API server: %v", err)
	}

	// Initialize threat detection system
	threatDetector := threat.NewDetector(threat.DefaultDetectorConfig())
	if err := threatDetector.Start(ctx); err != nil {
		log.Printf("Warning: Failed to start threat detector: %v", err)
	}

	// Initialize snapshot manager
	snapshotCfg := snapshot.DefaultSnapshotConfig()
	snapshotCfg.StoragePath = filepath.Join(cfg.Daemon.DataDir, "snapshots")
	snapshotMgr, err := snapshot.NewManager(snapshotCfg)
	if err != nil {
		log.Printf("Warning: Failed to initialize snapshot manager: %v", err)
	}

	// Initialize checkpoint system
	var checkpointer *snapshot.Checkpointer
	if snapshotMgr != nil {
		checkpointer = snapshot.NewCheckpointer(snapshotMgr, snapshot.DefaultCheckpointConfig())
		if err := checkpointer.Start(ctx); err != nil {
			log.Printf("Warning: Failed to start checkpointer: %v", err)
		}
	}

	// Initialize cloning manager
	cloningCfg := cloning.DefaultCloneConfig()
	cloningMgr := cloning.NewManager(cloningCfg, snapshotMgr, nil, nil)
	if err := cloningMgr.Start(ctx); err != nil {
		log.Printf("Warning: Failed to start cloning manager: %v", err)
	}

	// Wire up threat-triggered cloning
	threatResponder := threat.NewResponder(threatDetector, threat.DefaultResponseConfig())
	cloningMgr.SetThreatDetector(threatDetector, threatResponder)
	threatResponder.Start(ctx)

	fmt.Printf("Moltbunker daemon started\n")
	fmt.Printf("  P2P Port:    %d\n", cfg.Daemon.Port)
	fmt.Printf("  Node ID:     %s\n", node.NodeInfo().ID.String())
	fmt.Printf("  API Socket:  %s\n", cfg.Daemon.SocketPath)
	fmt.Printf("  Data Dir:    %s\n", cfg.Daemon.DataDir)
	fmt.Printf("  Config:      %s\n", cfgPath)
	fmt.Printf("  Network:     %s\n", cfg.P2P.NetworkMode)
	if cfg.Tor.Enabled {
		fmt.Printf("  Tor:         enabled\n")
	}
	if len(cfg.P2P.BootstrapNodes) > 0 {
		fmt.Printf("  Bootstrap:   %d nodes\n", len(cfg.P2P.BootstrapNodes))
	}

	// Wait for signal
	<-sigChan
	fmt.Println("\nShutting down gracefully...")

	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Stop new systems first
	threatResponder.Stop()
	threatDetector.Stop()
	cloningMgr.Stop()
	if checkpointer != nil {
		checkpointer.Stop()
	}

	// Stop API server
	if err := apiServer.Stop(); err != nil {
		log.Printf("Error stopping API server: %v", err)
	}

	// Gracefully shutdown node with timeout
	if err := node.Shutdown(shutdownCtx); err != nil {
		log.Printf("Error during shutdown: %v", err)
	}

	fmt.Println("Daemon stopped")
}
