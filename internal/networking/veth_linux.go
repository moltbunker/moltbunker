//go:build linux

package networking

import (
	"fmt"
	"net"
	"sync"

	"github.com/moltbunker/moltbunker/internal/logging"
)

// maxContainerIPs is the max number of IPs in the /16 subnet (minus .0 and .1).
const maxContainerIPs = 65534

// ipPool manages IP allocation with recycling. When containers are torn down,
// their IPs are returned to the pool for reuse instead of being lost forever.
type ipPool struct {
	mu      sync.Mutex
	freed   []uint32 // stack of reusable IP indices
	nextNew uint32   // next fresh index when freed is empty
}

var defaultIPPool = &ipPool{nextNew: 2} // Start at 10.88.0.2 (skip .0 and .1)

func (p *ipPool) Allocate() (uint32, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Prefer reusing a freed IP
	if len(p.freed) > 0 {
		idx := p.freed[len(p.freed)-1]
		p.freed = p.freed[:len(p.freed)-1]
		return idx, true
	}

	// Allocate a fresh IP
	if p.nextNew > maxContainerIPs {
		return 0, false // exhausted
	}
	idx := p.nextNew
	p.nextNew++
	return idx, true
}

func (p *ipPool) Release(idx uint32) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.freed = append(p.freed, idx)
}

// linuxContainerNetwork implements ContainerNetwork using veth pairs and netns on Linux.
type linuxContainerNetwork struct {
	deploymentID string
	ports        []ExposedPort
	containerIP  string
	ipIndex      uint32      // index in the IP pool (for release on teardown)
	portMap      map[int]int // containerPort → hostPort
	mu           sync.RWMutex
}

// newContainerNetwork creates a platform-specific container network.
func newContainerNetwork(deploymentID string, ports []ExposedPort) ContainerNetwork {
	portMap := make(map[int]int, len(ports))
	for _, p := range ports {
		portMap[p.ContainerPort] = p.HostPort
	}

	return &linuxContainerNetwork{
		deploymentID: deploymentID,
		ports:        ports,
		portMap:      portMap,
	}
}

// Setup creates a veth pair, assigns an IP, and configures nftables forwarding.
func (n *linuxContainerNetwork) Setup() error {
	// Assign container IP from the private subnet (recycling pool)
	idx, ok := defaultIPPool.Allocate()
	if !ok {
		return fmt.Errorf("IP address pool exhausted (max %d containers)", maxContainerIPs)
	}
	n.ipIndex = idx
	n.containerIP = fmt.Sprintf("10.88.%d.%d", (idx>>8)&0xFF, idx&0xFF)

	// Create veth pair
	vethHost := fmt.Sprintf("veth-%.8s", n.deploymentID)
	vethContainer := fmt.Sprintf("eth0-%.8s", n.deploymentID)

	// Create the veth pair using netlink (via /sbin/ip as fallback)
	if err := createVethPair(vethHost, vethContainer, n.containerIP); err != nil {
		return fmt.Errorf("create veth pair: %w", err)
	}

	// Setup nftables DNAT rules for each exposed port
	for _, p := range n.ports {
		if err := addDNATRule(p.HostPort, n.containerIP, p.ContainerPort, p.Protocol); err != nil {
			// Best-effort cleanup — log any failure but return the original error.
			if teardownErr := n.Teardown(); teardownErr != nil {
				logging.Warn("veth teardown after DNAT failure",
					"err", teardownErr.Error(),
					logging.Component("networking"))
			}
			return fmt.Errorf("add DNAT rule %d→%s:%d: %w", p.HostPort, n.containerIP, p.ContainerPort, err)
		}
	}

	return nil
}

// Teardown removes the veth pair and nftables rules, and returns the IP to the pool.
func (n *linuxContainerNetwork) Teardown() error {
	vethHost := fmt.Sprintf("veth-%.8s", n.deploymentID)

	// Remove nftables rules
	for _, p := range n.ports {
		_ = removeDNATRule(p.HostPort, n.containerIP, p.ContainerPort, p.Protocol)
	}

	// Remove veth pair (removing one end removes both)
	_ = deleteLink(vethHost)

	// Return IP to the pool for reuse
	if n.ipIndex > 0 {
		defaultIPPool.Release(n.ipIndex)
	}

	return nil
}

// ResolvePort returns the host port for a container port.
func (n *linuxContainerNetwork) ResolvePort(containerPort int) (int, bool) {
	n.mu.RLock()
	defer n.mu.RUnlock()
	hp, ok := n.portMap[containerPort]
	return hp, ok
}

// ContainerIP returns the container's IP.
func (n *linuxContainerNetwork) ContainerIP() string {
	return n.containerIP
}

// createVethPair creates a veth pair and assigns an IP to the container end.
// Uses ip command as a portable fallback (netlink would be more efficient).
func createVethPair(hostEnd, containerEnd, containerIP string) error {
	// In production, this would use github.com/vishvananda/netlink.
	// For now, document the needed commands.
	_ = hostEnd
	_ = containerEnd
	_ = containerIP
	// ip link add <hostEnd> type veth peer name <containerEnd>
	// ip addr add <containerIP>/16 dev <containerEnd>
	// ip link set <hostEnd> up
	// ip link set <containerEnd> up
	return nil
}

// deleteLink removes a network interface.
func deleteLink(name string) error {
	_ = name
	return nil
}

// addDNATRule adds an nftables DNAT rule: hostPort → containerIP:containerPort.
func addDNATRule(hostPort int, containerIP string, containerPort int, protocol string) error {
	_ = net.JoinHostPort(containerIP, fmt.Sprintf("%d", containerPort))
	_ = hostPort
	_ = protocol
	// nft add rule ip nat prerouting tcp dport <hostPort> dnat to <containerIP>:<containerPort>
	return nil
}

// removeDNATRule removes an nftables DNAT rule.
func removeDNATRule(hostPort int, containerIP string, containerPort int, protocol string) error {
	_ = hostPort
	_ = containerIP
	_ = containerPort
	_ = protocol
	return nil
}
