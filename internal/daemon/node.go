package daemon

import (
	"context"
	"fmt"
	"net"

	"github.com/moltbunker/moltbunker/internal/identity"
	"github.com/moltbunker/moltbunker/internal/p2p"
	"github.com/moltbunker/moltbunker/internal/security"
	"github.com/moltbunker/moltbunker/pkg/types"
)

// Node represents a P2P node in the network
type Node struct {
	keyManager    *identity.KeyManager
	walletManager *identity.WalletManager
	dht           *p2p.DHT
	router        *p2p.Router
	transport     *p2p.Transport
	geolocator    *p2p.GeoLocator
	nodeInfo      *types.Node
}

// NewNode creates a new P2P node
func NewNode(ctx context.Context, keyPath string, keystoreDir string, port int) (*Node, error) {
	// Initialize key manager
	keyManager, err := identity.NewKeyManager(keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create key manager: %w", err)
	}

	// Initialize wallet manager
	walletManager, err := identity.NewWalletManager(keystoreDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet manager: %w", err)
	}

	// Initialize DHT
	dht, err := p2p.NewDHT(ctx, port, keyManager)
	if err != nil {
		return nil, fmt.Errorf("failed to create DHT: %w", err)
	}

	// Initialize router
	router := p2p.NewRouter(dht)

	// Initialize certificate manager
	certManager, err := identity.NewCertificateManager(keyManager)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate manager: %w", err)
	}

	// Initialize pin store
	pinStore := security.NewCertPinStore()

	// Initialize transport
	transport, err := p2p.NewTransport(certManager, pinStore)
	if err != nil {
		return nil, fmt.Errorf("failed to create transport: %w", err)
	}

	// Initialize geolocator
	geolocator := p2p.NewGeoLocator()

	// Create node info
	nodeInfo := &types.Node{
		ID:        keyManager.NodeID(),
		PublicKey: keyManager.PublicKey(),
		Port:      port,
		Capabilities: types.NodeCapabilities{
			ContainerRuntime: true,
			TorSupport:      true,
		},
	}

	return &Node{
		keyManager:    keyManager,
		walletManager: walletManager,
		dht:           dht,
		router:        router,
		transport:     transport,
		geolocator:    geolocator,
		nodeInfo:      nodeInfo,
	}, nil
}

// Start starts the node
func (n *Node) Start(ctx context.Context) error {
	// Start listening
	listener, err := n.transport.Listen(fmt.Sprintf(":%d", n.nodeInfo.Port))
	if err != nil {
		return fmt.Errorf("failed to start listener: %w", err)
	}
	defer listener.Close()

	// Handle connections
	go n.handleConnections(ctx, listener)

	return nil
}

// handleConnections handles incoming connections
func (n *Node) handleConnections(ctx context.Context, listener net.Listener) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			conn, err := listener.Accept()
			if err != nil {
				continue
			}

			go n.handleConnection(ctx, conn)
		}
	}
}

// handleConnection handles a single connection
func (n *Node) handleConnection(ctx context.Context, conn net.Conn) {
	// TODO: Handle TLS connection and process messages
	defer conn.Close()
}

// NodeInfo returns node information
func (n *Node) NodeInfo() *types.Node {
	return n.nodeInfo
}

// Close closes the node
func (n *Node) Close() error {
	return n.dht.Close()
}
