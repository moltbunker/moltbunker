package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"path/filepath"

	"github.com/moltbunker/moltbunker/internal/api"
	"github.com/moltbunker/moltbunker/internal/cloning"
	"github.com/moltbunker/moltbunker/internal/config"
	"github.com/moltbunker/moltbunker/internal/logging"
	"github.com/moltbunker/moltbunker/internal/snapshot"
	"github.com/moltbunker/moltbunker/internal/threat"
)

func main() {
	// Parse flags
	httpAddr := flag.String("http", ":8080", "HTTP listen address")
	httpsAddr := flag.String("https", ":8443", "HTTPS listen address")
	enableHTTPS := flag.Bool("tls", false, "Enable HTTPS")
	tlsCert := flag.String("tls-cert", "", "TLS certificate file")
	tlsKey := flag.String("tls-key", "", "TLS key file")
	configPath := flag.String("config", "", "Path to config file")
	enableAuth := flag.Bool("auth", true, "Enable API authentication")
	flag.Parse()

	// Load configuration
	var cfg *config.Config
	var err error
	if *configPath != "" {
		cfg, err = config.Load(*configPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			os.Exit(1)
		}
	} else {
		cfg = config.DefaultConfig()
	}

	// Set up logging
	if cfg.Daemon.LogFormat == "json" {
		logging.SetOutput(os.Stdout)
	} else {
		logging.SetTextOutput(os.Stdout)
	}

	// Create server config
	serverCfg := &api.ServerConfig{
		HTTPAddr:         *httpAddr,
		HTTPSAddr:        *httpsAddr,
		EnableHTTPS:      *enableHTTPS,
		TLSCertFile:      *tlsCert,
		TLSKeyFile:       *tlsKey,
		DaemonSocketPath: cfg.Daemon.SocketPath,
		DaemonPoolSize:   4,
		RateLimit:        100,
		RateLimitBurst:   20,
		EnableAuth:       *enableAuth,
		APIKeyHeader:     "X-API-Key",
		EnableCORS:       true,
		AllowedOrigins:   []string{"*"},
		ReadTimeout:      30 * time.Second,
		WriteTimeout:     30 * time.Second,
		IdleTimeout:      120 * time.Second,
		EnableWebSocket:  true,
	}

	// Create server
	server := api.NewServer(serverCfg)

	// Initialize threat detector
	detectorCfg := threat.DefaultDetectorConfig()
	detector := threat.NewDetector(detectorCfg)

	// Initialize snapshot manager
	snapshotCfg := snapshot.DefaultSnapshotConfig()
	snapshotCfg.StoragePath = cfg.Daemon.DataDir + "/snapshots"
	snapshotMgr, err := snapshot.NewManager(snapshotCfg)
	if err != nil {
		logging.Warn("Failed to initialize snapshot manager",
			"error", err.Error(),
			logging.Component("api"))
	}

	// Initialize cloning manager
	cloningCfg := cloning.DefaultCloneConfig()
	cloningMgr := cloning.NewManager(cloningCfg, snapshotMgr, nil, nil)

	// Set up threat responder with cloning integration
	responderCfg := threat.DefaultResponseConfig()
	responder := threat.NewResponder(detector, responderCfg)
	cloningMgr.SetThreatDetector(detector, responder)

	// Initialize admin stores
	adminStore := api.NewAdminMetadataStore(filepath.Join(cfg.Daemon.DataDir, "admin_metadata.json"))
	policyStore := api.NewPolicyStore(filepath.Join(cfg.Daemon.DataDir, "admin_policies.json"))
	catalogStore := api.NewCatalogStore(filepath.Join(cfg.Daemon.DataDir, "catalog.json"))

	// Wire up components
	server.SetFullConfig(cfg)
	server.SetThreatDetector(detector)
	server.SetSnapshotManager(snapshotMgr)
	server.SetCloningManager(cloningMgr)
	server.SetAdminStore(adminStore)
	server.SetPolicyStore(policyStore)
	server.SetCatalogStore(catalogStore)

	// Set up signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Start components
	if err := detector.Start(ctx); err != nil {
		logging.Error("Failed to start threat detector",
			"error", err.Error(),
			logging.Component("api"))
	}
	responder.Start(ctx)

	if err := cloningMgr.Start(ctx); err != nil {
		logging.Error("Failed to start cloning manager",
			"error", err.Error(),
			logging.Component("api"))
	}

	// Start API server
	if err := server.Start(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting server: %v\n", err)
		os.Exit(1)
	}

	logging.Info("MoltBunker API server started",
		"http_addr", *httpAddr,
		"https_enabled", *enableHTTPS,
		"auth_enabled", *enableAuth,
		logging.Component("api"))

	// Wait for shutdown signal
	<-sigCh
	logging.Info("Shutting down...", logging.Component("api"))

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := server.Stop(shutdownCtx); err != nil {
		logging.Error("Error during shutdown",
			"error", err.Error(),
			logging.Component("api"))
	}

	detector.Stop()
	responder.Stop()
	cloningMgr.Stop()

	logging.Info("Shutdown complete", logging.Component("api"))
}
