package p2p

import (
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/moltbunker/moltbunker/internal/logging"
	"github.com/moltbunker/moltbunker/pkg/types"
)

const (
	// DefaultMaxPeersPerSubnet is the default max peers from the same /24 subnet.
	DefaultMaxPeersPerSubnet = 3
)

// SubnetLimiter limits the number of peers accepted from the same /24 subnet
// to defend against Sybil attacks from a single network.
type SubnetLimiter struct {
	mu           sync.RWMutex
	subnetCounts map[string]int           // /24 prefix → count
	peerSubnets  map[types.NodeID]string  // nodeID → /24 prefix
	maxPerSubnet int
}

// NewSubnetLimiter creates a subnet limiter with the given max peers per /24.
func NewSubnetLimiter(maxPerSubnet int) *SubnetLimiter {
	if maxPerSubnet <= 0 {
		maxPerSubnet = DefaultMaxPeersPerSubnet
	}
	return &SubnetLimiter{
		subnetCounts: make(map[string]int),
		peerSubnets:  make(map[types.NodeID]string),
		maxPerSubnet: maxPerSubnet,
	}
}

// Allow checks whether a new peer from the given remote address can be accepted.
// Returns true if allowed, false if the subnet limit has been reached.
// Exempt addresses (localhost, private networks, .onion) always return true.
func (sl *SubnetLimiter) Allow(remoteAddr string, nodeID types.NodeID) bool {
	// Parse host from address (may include port)
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		// Maybe address has no port
		host = remoteAddr
	}

	// Exempt .onion addresses
	if strings.HasSuffix(host, ".onion") {
		return true
	}

	ip := net.ParseIP(host)
	if ip == nil {
		// Can't parse — allow (might be hostname or .onion)
		return true
	}

	// Exempt private/localhost addresses
	if isExemptIP(ip) {
		return true
	}

	// Get /24 prefix
	subnet := subnetPrefix(ip)

	sl.mu.Lock()
	defer sl.mu.Unlock()

	// If this peer is already tracked, allow (re-connection)
	if existingSubnet, exists := sl.peerSubnets[nodeID]; exists {
		if existingSubnet == subnet {
			return true
		}
		// Peer changed subnet — remove old tracking
		sl.subnetCounts[existingSubnet]--
		if sl.subnetCounts[existingSubnet] <= 0 {
			delete(sl.subnetCounts, existingSubnet)
		}
	}

	// Check limit
	if sl.subnetCounts[subnet] >= sl.maxPerSubnet {
		logging.Warn("subnet peer limit reached",
			"subnet", subnet,
			"count", sl.subnetCounts[subnet],
			"max", sl.maxPerSubnet,
			logging.NodeID(nodeID.String()[:16]),
			logging.Component("subnet_limiter"))
		return false
	}

	// Allow and track
	sl.subnetCounts[subnet]++
	sl.peerSubnets[nodeID] = subnet
	return true
}

// Remove removes a peer from subnet tracking.
func (sl *SubnetLimiter) Remove(nodeID types.NodeID) {
	sl.mu.Lock()
	defer sl.mu.Unlock()

	subnet, exists := sl.peerSubnets[nodeID]
	if !exists {
		return
	}

	delete(sl.peerSubnets, nodeID)
	sl.subnetCounts[subnet]--
	if sl.subnetCounts[subnet] <= 0 {
		delete(sl.subnetCounts, subnet)
	}
}

// PeerCount returns the number of tracked peers.
func (sl *SubnetLimiter) PeerCount() int {
	sl.mu.RLock()
	defer sl.mu.RUnlock()
	return len(sl.peerSubnets)
}

// subnetPrefix returns the /24 prefix string for an IPv4 address.
// For IPv6, returns /48 prefix.
func subnetPrefix(ip net.IP) string {
	if v4 := ip.To4(); v4 != nil {
		return fmt.Sprintf("%d.%d.%d.0/24", v4[0], v4[1], v4[2])
	}
	// IPv6: use /48
	ip6 := ip.To16()
	prefix := make(net.IP, 16)
	copy(prefix[:6], ip6[:6])
	return prefix.String() + "/48"
}

// isExemptIP returns true for addresses that are exempt from subnet limiting.
func isExemptIP(ip net.IP) bool {
	// Loopback (127.0.0.0/8)
	if ip.IsLoopback() {
		return true
	}

	// Private ranges
	privateRanges := []struct {
		network *net.IPNet
	}{
		{mustParseCIDR("10.0.0.0/8")},
		{mustParseCIDR("172.16.0.0/12")},
		{mustParseCIDR("192.168.0.0/16")},
		{mustParseCIDR("fc00::/7")},   // IPv6 unique local
		{mustParseCIDR("fe80::/10")},  // IPv6 link-local
	}

	for _, pr := range privateRanges {
		if pr.network.Contains(ip) {
			return true
		}
	}

	return false
}

func mustParseCIDR(s string) *net.IPNet {
	_, network, err := net.ParseCIDR(s)
	if err != nil {
		panic("invalid CIDR: " + s)
	}
	return network
}
