package p2p

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p-kad-dht/dual"
	"github.com/moltbunker/moltbunker/pkg/types"
)

// DHT implements Kademlia DHT for node discovery
type DHT struct {
	host     host.Host
	dht      *dual.DHT
	mdns     mdns.Service
	peers    map[types.NodeID]*types.Node
	peersMu  sync.RWMutex
	bootstrap []peer.AddrInfo
}

// NewDHT creates a new DHT instance
func NewDHT(ctx context.Context, port int, keyManager interface{}) (*DHT, error) {
	// Note: keyManager parameter is for future use with libp2p identity
	// Create libp2p host
	h, err := libp2p.New(
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port)),
		libp2p.NATPortMap(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create libp2p host: %w", err)
	}

	// Create dual DHT (supports both IPFS and IPNS)
	dhtInstance, err := dual.New(ctx, h, dual.DHTOption(
		dht.Mode(dht.ModeServer),
		dht.ProtocolPrefix("/moltbunker"),
	))
	if err != nil {
		return nil, fmt.Errorf("failed to create DHT: %w", err)
	}

	// Create mDNS service for local discovery
	mdnsService := mdns.NewMdnsService(h, "moltbunker", &mdnsNotifee{
		peers: make(chan peer.AddrInfo, 10),
	})

	d := &DHT{
		host:  h,
		dht:   dhtInstance,
		mdns:  mdnsService,
		peers: make(map[types.NodeID]*types.Node),
	}

	// Bootstrap DHT
	if err := d.Bootstrap(ctx); err != nil {
		return nil, fmt.Errorf("failed to bootstrap DHT: %w", err)
	}

	return d, nil
}

// Bootstrap connects to bootstrap nodes
func (d *DHT) Bootstrap(ctx context.Context) error {
	// Add hardcoded bootstrap nodes (clearnet)
	// In production, these would be actual bootstrap nodes
	bootstrapPeers := []string{
		// Add bootstrap peer addresses here
	}

	for _, addr := range bootstrapPeers {
		peerInfo, err := peer.AddrInfoFromString(addr)
		if err != nil {
			continue
		}
		d.bootstrap = append(d.bootstrap, *peerInfo)
	}

	// Connect to bootstrap peers
	for _, peerInfo := range d.bootstrap {
		if err := d.host.Connect(ctx, peerInfo); err != nil {
			continue // Continue if bootstrap fails
		}
	}

	// Bootstrap DHT
	if err := d.dht.Bootstrap(ctx); err != nil {
		return fmt.Errorf("failed to bootstrap DHT: %w", err)
	}

	return nil
}

// FindNode finds nodes close to a given node ID
func (d *DHT) FindNode(ctx context.Context, targetID types.NodeID) ([]*types.Node, error) {
	// Convert NodeID to peer.ID (using first 20 bytes for libp2p peer ID)
	var peerIDBytes [20]byte
	copy(peerIDBytes[:], targetID[:20])
	peerID := peer.ID(peerIDBytes)

	peers, err := d.dht.FindPeers(ctx, peerID)
	if err != nil {
		return nil, fmt.Errorf("failed to find peers: %w", err)
	}

	var nodes []*types.Node
	for p := range peers {
		// Convert peer to node
		address := ""
		if len(p.Addrs) > 0 {
			address = p.Addrs[0].String()
		}
		
		// Convert peer.ID back to NodeID (pad with zeros)
		var nodeID types.NodeID
		copy(nodeID[:], []byte(p.ID))
		
		node := &types.Node{
			ID:       nodeID,
			Address:  address,
			LastSeen: time.Now(),
		}
		nodes = append(nodes, node)
	}

	return nodes, nil
}

// AddPeer adds a peer to the peer list
func (d *DHT) AddPeer(node *types.Node) {
	d.peersMu.Lock()
	defer d.peersMu.Unlock()
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

// Host returns the libp2p host
func (d *DHT) Host() host.Host {
	return d.host
}

// Close closes the DHT
func (d *DHT) Close() error {
	if err := d.mdns.Close(); err != nil {
		return err
	}
	return d.dht.Close()
}

// mdnsNotifee handles mDNS peer discovery
type mdnsNotifee struct {
	peers chan peer.AddrInfo
}

func (m *mdnsNotifee) HandlePeerFound(pi peer.AddrInfo) {
	select {
	case m.peers <- pi:
	default:
		// Channel full, skip
	}
}
