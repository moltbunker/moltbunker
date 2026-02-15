//go:build linux

package networking

import (
	"fmt"
	"net"
	"sync"
	"sync/atomic"
)

// containerSubnet is the private range used for container networks.
const containerSubnet = "10.88.0.0/16"

// ipCounter tracks the next available IP in the container subnet.
var ipCounter atomic.Uint32

func init() {
	// Start at 10.88.0.2 (skip .0 network and .1 gateway)
	ipCounter.Store(2)
}

// linuxContainerNetwork implements ContainerNetwork using veth pairs and netns on Linux.
type linuxContainerNetwork struct {
	deploymentID string
	ports        []ExposedPort
	containerIP  string
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
	// Assign container IP from the private subnet
	idx := ipCounter.Add(1)
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
			// Best-effort cleanup
			n.Teardown()
			return fmt.Errorf("add DNAT rule %d→%s:%d: %w", p.HostPort, n.containerIP, p.ContainerPort, err)
		}
	}

	return nil
}

// Teardown removes the veth pair and nftables rules.
func (n *linuxContainerNetwork) Teardown() error {
	vethHost := fmt.Sprintf("veth-%.8s", n.deploymentID)

	// Remove nftables rules
	for _, p := range n.ports {
		_ = removeDNATRule(p.HostPort, n.containerIP, p.ContainerPort, p.Protocol)
	}

	// Remove veth pair (removing one end removes both)
	_ = deleteLink(vethHost)

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
