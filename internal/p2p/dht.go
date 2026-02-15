package p2p

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p-kad-dht/dual"
	ma "github.com/multiformats/go-multiaddr"

	"github.com/moltbunker/moltbunker/internal/identity"
	"github.com/moltbunker/moltbunker/internal/logging"
	"github.com/moltbunker/moltbunker/internal/util"
	"github.com/moltbunker/moltbunker/pkg/types"
)

// DHTConfig contains DHT configuration
type DHTConfig struct {
	Port           int
	BootstrapPeers []string
	EnableMDNS     bool
	EnableNAT      bool // Enable NAT traversal (AutoNAT, port mapping, relay, hole punching)
	ExternalIP     string
	AnnounceAddrs  []string
	MaxPeers       int

	// Connection manager watermarks for automatic pruning.
	// When connections exceed HighWatermark, the manager prunes down to LowWatermark.
	ConnManagerLowWatermark  int
	ConnManagerHighWatermark int

	// AddressBookPath is the file path for persisting the peer address book.
	// If empty, the address book is kept in memory only.
	AddressBookPath string

	// DNSBootstrap configures DNS-based bootstrap peer resolution.
	// When set, bootstrap peers are resolved via DNS TXT records at
	// _dnsaddr.<domain> before falling back to static peers.
	// If nil, the default BootstrapConfig is used for DNS resolution.
	DNSBootstrap *BootstrapConfig
}

// DHT implements Kademlia DHT for node discovery
type DHT struct {
	host        host.Host
	dht         *dual.DHT
	mdns        mdns.Service
	peers       map[types.NodeID]*types.Node
	peersMu     sync.RWMutex
	config      *DHTConfig
	keyManager  *identity.KeyManager
	localNode   *types.Node
	addressBook *AddressBook

	// Callbacks
	onPeerConnected    func(*types.Node)
	onPeerDisconnected func(*types.Node)
}

// DefaultBootstrapPeers returns default bootstrap peers.
// It resolves DNS-based bootstrap via /dnsaddr/bootstrap.moltbunker.com.
// If DNS resolution fails, returns an empty list — the node will still
// discover peers via mDNS (LAN) and the persisted address book.
// We intentionally do NOT fall back to public IPFS/libp2p bootstrap nodes
// to avoid untrusted peers polluting our routing table.
func DefaultBootstrapPeers() []string {
	resolved := BootstrapPeerStrings(DefaultBootstrapConfig())
	if len(resolved) > 0 {
		return resolved
	}

	logging.Info("no bootstrap peers resolved via DNS, relying on mDNS and address book",
		logging.Component("dht"))
	return []string{}
}

// NewDHT creates a new DHT instance
func NewDHT(ctx context.Context, config *DHTConfig, keyManager *identity.KeyManager) (*DHT, error) {
	if config == nil {
		config = &DHTConfig{
			Port:           9000,
			BootstrapPeers: DefaultBootstrapPeers(),
			EnableMDNS:     true,
			EnableNAT:      true,
			MaxPeers:       50,
		}
	}

	// Resolve bootstrap peers via DNS if DNSBootstrap is configured
	// and no explicit BootstrapPeers are set
	if len(config.BootstrapPeers) == 0 && config.DNSBootstrap != nil {
		config.BootstrapPeers = BootstrapPeerStrings(config.DNSBootstrap)
	}

	// Build listen addresses
	listenAddrs := []string{
		fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", config.Port),
		fmt.Sprintf("/ip6/::/tcp/%d", config.Port),
	}

	// Build libp2p options
	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(listenAddrs...),
	}

	// Configure resource manager to prevent resource exhaustion from connection flooding
	rm, err := NewResourceManager()
	if err != nil {
		logging.Warn("failed to create resource manager, continuing without it", "error", err)
	} else {
		opts = append(opts, libp2p.ResourceManager(rm))
	}

	// Configure connection manager with watermark-based pruning
	{
		low := config.ConnManagerLowWatermark
		if low <= 0 {
			low = DefaultConnManagerLowWatermark
		}
		high := config.ConnManagerHighWatermark
		if high <= 0 {
			high = DefaultConnManagerHighWatermark
		}
		cm, cmErr := NewConnectionManager(low, high, DefaultConnManagerGracePeriod)
		if cmErr != nil {
			logging.Warn("failed to create connection manager, continuing without it", "error", cmErr)
		} else {
			opts = append(opts, libp2p.ConnectionManager(cm))
		}
	}

	// Enable NAT traversal options when configured
	if config.EnableNAT {
		opts = append(opts,
			libp2p.NATPortMap(),        // Automatic port mapping via UPnP/NAT-PMP
			libp2p.EnableNATService(),   // Help other nodes determine their NAT status
			libp2p.EnableRelay(),        // Act as a relay for nodes behind NAT
			libp2p.EnableHolePunching(), // Direct connection upgrade via hole punching
		)
	}

	// Add external addresses if configured
	if config.ExternalIP != "" {
		extAddr := fmt.Sprintf("/ip4/%s/tcp/%d", config.ExternalIP, config.Port)
		opts = append(opts, libp2p.AddrsFactory(func(addrs []ma.Multiaddr) []ma.Multiaddr {
			extMA, err := ma.NewMultiaddr(extAddr)
			if err == nil {
				addrs = append(addrs, extMA)
			}
			return addrs
		}))
	}

	// Add announce addresses
	if len(config.AnnounceAddrs) > 0 {
		var announceAddrs []ma.Multiaddr
		for _, addr := range config.AnnounceAddrs {
			maddr, err := ma.NewMultiaddr(addr)
			if err == nil {
				announceAddrs = append(announceAddrs, maddr)
			}
		}
		if len(announceAddrs) > 0 {
			opts = append(opts, libp2p.AddrsFactory(func(addrs []ma.Multiaddr) []ma.Multiaddr {
				return append(addrs, announceAddrs...)
			}))
		}
	}

	// Create libp2p host
	h, err := libp2p.New(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create libp2p host: %w", err)
	}

	// Create dual DHT
	dhtInstance, err := dual.New(ctx, h,
		dual.DHTOption(
			dht.Mode(dht.ModeAutoServer),
			dht.ProtocolPrefix("/moltbunker"),
			dht.BucketSize(20),
		),
	)
	if err != nil {
		h.Close()
		return nil, fmt.Errorf("failed to create DHT: %w", err)
	}

	d := &DHT{
		host:        h,
		dht:         dhtInstance,
		peers:       make(map[types.NodeID]*types.Node),
		config:      config,
		keyManager:  keyManager,
		addressBook: NewAddressBook(),
	}

	// Create local node info
	d.localNode = d.createLocalNode()

	// Load address book from disk and connect to known peers before DHT bootstrap
	if config.AddressBookPath != "" {
		if err := d.addressBook.Load(config.AddressBookPath); err != nil {
			logging.Warn("failed to load address book", "error", err, logging.Component("dht"))
		} else {
			d.connectFromAddressBook(ctx)
		}
	}

	// Set up connection notifier
	h.Network().Notify(&network.NotifyBundle{
		ConnectedF: func(n network.Network, conn network.Conn) {
			d.handlePeerConnected(conn.RemotePeer())
		},
		DisconnectedF: func(n network.Network, conn network.Conn) {
			d.handlePeerDisconnected(conn.RemotePeer())
		},
	})

	// Start mDNS if enabled
	if config.EnableMDNS {
		notifee := &mdnsNotifee{dht: d}
		mdnsService := mdns.NewMdnsService(h, "moltbunker-discovery", notifee)
		if mdnsService != nil {
			d.mdns = mdnsService
		}
	}

	// Bootstrap
	if err := d.Bootstrap(ctx); err != nil {
		// Bootstrap failure is non-fatal - we might still get peers via mDNS
		// Error is ignored as this is expected in some network configurations
		_ = err
	}

	return d, nil
}

// createLocalNode creates the local node info
func (d *DHT) createLocalNode() *types.Node {
	var nodeID types.NodeID

	if d.keyManager != nil {
		nodeID = d.keyManager.NodeID()
	} else {
		// Derive from peer ID
		peerIDBytes := []byte(d.host.ID())
		hash := sha256.Sum256(peerIDBytes)
		nodeID = types.NodeID(hash)
	}

	// Get addresses
	addrs := d.host.Addrs()
	address := ""
	port := d.config.Port

	for _, addr := range addrs {
		addrStr := addr.String()
		// Prefer non-localhost addresses
		if !strings.Contains(addrStr, "127.0.0.1") && !strings.Contains(addrStr, "::1") {
			address = addrStr
			break
		}
	}

	if address == "" && len(addrs) > 0 {
		address = addrs[0].String()
	}

	return &types.Node{
		ID:       nodeID,
		Address:  address,
		Port:     port,
		LastSeen: time.Now(),
		Capabilities: types.NodeCapabilities{
			ContainerRuntime: true,
			TorSupport:       true,
		},
	}
}

// Bootstrap connects to bootstrap nodes
func (d *DHT) Bootstrap(ctx context.Context) error {
	bootstrapPeers := d.config.BootstrapPeers
	if len(bootstrapPeers) == 0 {
		bootstrapPeers = DefaultBootstrapPeers()
	}

	var wg sync.WaitGroup
	var connectedCount int
	var mu sync.Mutex

	for _, addrStr := range bootstrapPeers {
		peerInfo, err := peer.AddrInfoFromString(addrStr)
		if err != nil {
			continue
		}

		// Add bootstrap peers to address book
		d.addressBook.AddPeer(peerInfo.ID, peerInfo.Addrs, "bootstrap")

		wg.Add(1)
		pi := *peerInfo
		util.SafeGoWithName("dht-bootstrap-connect", func() {
			defer wg.Done()

			connectCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
			defer cancel()

			d.host.Peerstore().AddAddrs(pi.ID, pi.Addrs, peerstore.PermanentAddrTTL)

			if err := d.host.Connect(connectCtx, pi); err != nil {
				d.addressBook.RecordConnectionAttempt(pi.ID, false)
				return
			}

			d.addressBook.RecordConnectionAttempt(pi.ID, true)
			mu.Lock()
			connectedCount++
			mu.Unlock()
		})
	}

	wg.Wait()

	// Bootstrap the DHT
	if err := d.dht.Bootstrap(ctx); err != nil {
		return fmt.Errorf("DHT bootstrap failed: %w", err)
	}

	return nil
}

// handlePeerConnected handles new peer connections
func (d *DHT) handlePeerConnected(peerID peer.ID) {
	node := d.peerIDToNode(peerID)

	d.peersMu.Lock()
	d.peers[node.ID] = node
	d.peersMu.Unlock()

	// Track in address book
	addrs := d.host.Peerstore().Addrs(peerID)
	if len(addrs) > 0 {
		d.addressBook.AddPeer(peerID, addrs, "dht")
	}

	if d.onPeerConnected != nil {
		d.onPeerConnected(node)
	}
}

// handlePeerDisconnected handles peer disconnections
func (d *DHT) handlePeerDisconnected(peerID peer.ID) {
	node := d.peerIDToNode(peerID)

	d.peersMu.Lock()
	delete(d.peers, node.ID)
	d.peersMu.Unlock()

	// Update last seen in address book (peer was known until now)
	d.addressBook.UpdateLastSeen(peerID)

	if d.onPeerDisconnected != nil {
		d.onPeerDisconnected(node)
	}
}

// peerIDToNode converts a peer.ID to a types.Node
func (d *DHT) peerIDToNode(peerID peer.ID) *types.Node {
	// Get peer addresses
	addrs := d.host.Peerstore().Addrs(peerID)
	address := ""
	port := 0

	for _, addr := range addrs {
		addrStr := addr.String()
		// Parse address to extract IP and port
		if strings.Contains(addrStr, "/tcp/") {
			address = addrStr
			// Extract port from multiaddr
			parts := strings.Split(addrStr, "/tcp/")
			if len(parts) > 1 {
				portParts := strings.Split(parts[1], "/")
				if len(portParts) > 0 {
					fmt.Sscanf(portParts[0], "%d", &port)
				}
			}
			break
		}
	}

	// Create NodeID from peer ID
	var nodeID types.NodeID
	peerIDBytes := []byte(peerID)
	hash := sha256.Sum256(peerIDBytes)
	nodeID = types.NodeID(hash)

	// Try to get geolocation
	geolocator := NewGeoLocator()
	var country, region string
	if ip := extractIPFromMultiaddr(address); ip != "" {
		if loc, err := geolocator.GetLocationFromIP(ip); err == nil && loc != nil {
			country = loc.Country
			region = GetRegionFromCountry(country)
		}
	}

	return &types.Node{
		ID:       nodeID,
		Address:  address,
		Port:     port,
		Country:  country,
		Region:   region,
		LastSeen: time.Now(),
		Capabilities: types.NodeCapabilities{
			ContainerRuntime: true,
			TorSupport:       true,
		},
	}
}

// extractIPFromMultiaddr extracts IP address from multiaddr string
func extractIPFromMultiaddr(addr string) string {
	// /ip4/1.2.3.4/tcp/9000 -> 1.2.3.4
	if strings.HasPrefix(addr, "/ip4/") {
		parts := strings.Split(addr, "/")
		if len(parts) > 2 {
			return parts[2]
		}
	}
	return ""
}

// FindNode finds nodes close to a given node ID
func (d *DHT) FindNode(ctx context.Context, targetID types.NodeID) ([]*types.Node, error) {
	// First check our local peer cache
	d.peersMu.RLock()
	if node, exists := d.peers[targetID]; exists {
		d.peersMu.RUnlock()
		return []*types.Node{node}, nil
	}
	d.peersMu.RUnlock()

	// Search via DHT
	targetIDStr := hex.EncodeToString(targetID[:])
	peerID, err := peer.Decode(targetIDStr)
	if err == nil {
		// Try to find the specific peer
		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		peerInfo, err := d.dht.FindPeer(ctx, peerID)
		if err == nil {
			node := d.peerIDToNode(peerInfo.ID)
			return []*types.Node{node}, nil
		}
	}

	// Return all connected peers
	return d.GetConnectedPeers(), nil
}

// GetConnectedPeers returns all currently connected peers
func (d *DHT) GetConnectedPeers() []*types.Node {
	connectedPeers := d.host.Network().Peers()
	nodes := make([]*types.Node, 0, len(connectedPeers))

	for _, pID := range connectedPeers {
		node := d.peerIDToNode(pID)
		nodes = append(nodes, node)
	}

	return nodes
}

// AddPeer adds a peer to the peer list
func (d *DHT) AddPeer(node *types.Node) {
	d.peersMu.Lock()
	defer d.peersMu.Unlock()
	node.LastSeen = time.Now()
	d.peers[node.ID] = node
}

// GetPeer retrieves a peer by node ID
func (d *DHT) GetPeer(nodeID types.NodeID) (*types.Node, bool) {
	d.peersMu.RLock()
	defer d.peersMu.RUnlock()
	node, exists := d.peers[nodeID]
	return node, exists
}

// GetPeers returns all known peers
func (d *DHT) GetPeers() []*types.Node {
	d.peersMu.RLock()
	defer d.peersMu.RUnlock()

	peers := make([]*types.Node, 0, len(d.peers))
	for _, node := range d.peers {
		peers = append(peers, node)
	}
	return peers
}

// LocalNode returns the local node info
func (d *DHT) LocalNode() *types.Node {
	return d.localNode
}

// Host returns the libp2p host
func (d *DHT) Host() host.Host {
	return d.host
}

// PeerCount returns the number of connected peers
func (d *DHT) PeerCount() int {
	return len(d.host.Network().Peers())
}

// SetPeerConnectedCallback sets the callback for peer connections
func (d *DHT) SetPeerConnectedCallback(cb func(*types.Node)) {
	d.onPeerConnected = cb
}

// SetPeerDisconnectedCallback sets the callback for peer disconnections
func (d *DHT) SetPeerDisconnectedCallback(cb func(*types.Node)) {
	d.onPeerDisconnected = cb
}

// ConnectToPeer connects to a specific peer address
func (d *DHT) ConnectToPeer(ctx context.Context, addr string) error {
	peerInfo, err := peer.AddrInfoFromString(addr)
	if err != nil {
		return fmt.Errorf("invalid peer address: %w", err)
	}

	d.host.Peerstore().AddAddrs(peerInfo.ID, peerInfo.Addrs, peerstore.PermanentAddrTTL)

	if err := d.host.Connect(ctx, *peerInfo); err != nil {
		return fmt.Errorf("failed to connect to peer: %w", err)
	}

	return nil
}

// Addresses returns the multiaddresses this node is listening on
func (d *DHT) Addresses() []string {
	addrs := d.host.Addrs()
	peerID := d.host.ID()

	fullAddrs := make([]string, 0, len(addrs))
	for _, addr := range addrs {
		fullAddrs = append(fullAddrs, fmt.Sprintf("%s/p2p/%s", addr.String(), peerID.String()))
	}

	return fullAddrs
}

// AddressBook returns the address book for this DHT instance.
func (d *DHT) AddressBook() *AddressBook {
	return d.addressBook
}

// connectFromAddressBook connects to the best known peers from the persisted
// address book. This is called on startup before DHT bootstrap to quickly
// reconnect to previously known peers.
func (d *DHT) connectFromAddressBook(ctx context.Context) {
	entries := d.addressBook.GetBestPeers(d.config.MaxPeers)
	if len(entries) == 0 {
		return
	}

	logging.Info("connecting to peers from address book",
		"peers", len(entries),
		logging.Component("dht"))

	var connected, failed int
	for _, entry := range entries {
		// Skip ourselves
		if entry.PeerID == d.host.ID() {
			continue
		}

		// Parse multiaddresses
		var addrs []ma.Multiaddr
		for _, addrStr := range entry.Addrs {
			maddr, err := ma.NewMultiaddr(addrStr)
			if err != nil {
				continue
			}
			addrs = append(addrs, maddr)
		}
		if len(addrs) == 0 {
			continue
		}

		// Add addresses to the peerstore
		d.host.Peerstore().AddAddrs(entry.PeerID, addrs, peerstore.PermanentAddrTTL)

		// Attempt connection with a timeout
		connectCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
		err := d.host.Connect(connectCtx, peer.AddrInfo{
			ID:    entry.PeerID,
			Addrs: addrs,
		})
		cancel()

		if err != nil {
			d.addressBook.RecordConnectionAttempt(entry.PeerID, false)
			failed++
		} else {
			d.addressBook.RecordConnectionAttempt(entry.PeerID, true)
			connected++
		}
	}

	logging.Info("address book reconnection complete",
		"connected", connected,
		"failed", failed,
		logging.Component("dht"))
}

// Close closes the DHT and its underlying libp2p host.
// Both resources are closed regardless of individual errors.
func (d *DHT) Close() error {
	if d.mdns != nil {
		d.mdns.Close()
	}
	dhtErr := d.dht.Close()
	hostErr := d.host.Close()
	return errors.Join(dhtErr, hostErr)
}

// mdnsNotifee handles mDNS peer discovery
type mdnsNotifee struct {
	dht *DHT
}

func (m *mdnsNotifee) HandlePeerFound(pi peer.AddrInfo) {
	// Don't connect to ourselves
	if pi.ID == m.dht.host.ID() {
		return
	}

	// Add addresses to peerstore
	m.dht.host.Peerstore().AddAddrs(pi.ID, pi.Addrs, peerstore.TempAddrTTL)

	// Track in address book
	m.dht.addressBook.AddPeer(pi.ID, pi.Addrs, "mdns")

	// Try to connect
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := m.dht.host.Connect(ctx, pi); err != nil {
		m.dht.addressBook.RecordConnectionAttempt(pi.ID, false)
		// Connection failed, but that's okay for mDNS discovery
		return
	}
	m.dht.addressBook.RecordConnectionAttempt(pi.ID, true)
}
