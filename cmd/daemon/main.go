package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/moltbunker/moltbunker/internal/daemon"
)

var (
	port       = flag.Int("port", 9000, "P2P port")
	keyPath    = flag.String("key", "~/.moltbunker/keys/node.key", "Path to node key")
	keystoreDir = flag.String("keystore", "~/.moltbunker/keystore", "Path to keystore")
	dataDir    = flag.String("data", "~/.moltbunker/data", "Path to data directory")
)

func main() {
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Create node
	node, err := daemon.NewNode(ctx, *keyPath, *keystoreDir, *port)
	if err != nil {
		log.Fatalf("Failed to create node: %v", err)
	}

	// Start node
	if err := node.Start(ctx); err != nil {
		log.Fatalf("Failed to start node: %v", err)
	}

	fmt.Printf("Moltbunker daemon started on port %d\n", *port)
	fmt.Printf("Node ID: %s\n", node.NodeInfo().ID.String())

	// Wait for signal
	<-sigChan
	fmt.Println("\nShutting down...")

	// Close node
	if err := node.Close(); err != nil {
		log.Printf("Error closing node: %v", err)
	}
}
