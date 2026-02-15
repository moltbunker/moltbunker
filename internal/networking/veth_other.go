//go:build !linux

package networking

import "sync"

// fallbackContainerNetwork is a no-op network implementation for non-Linux platforms.
// On macOS (Colima), containers use host networking within the VM, so veth pairs
// and nftables are not needed. Port forwarding relies on Colima's built-in port mapping.
type fallbackContainerNetwork struct {
	deploymentID string
	ports        []ExposedPort
	portMap      map[int]int // containerPort → hostPort
	mu           sync.RWMutex
}

// newContainerNetwork creates a platform-specific container network.
func newContainerNetwork(deploymentID string, ports []ExposedPort) ContainerNetwork {
	portMap := make(map[int]int, len(ports))
	for _, p := range ports {
		portMap[p.ContainerPort] = p.HostPort
	}

	return &fallbackContainerNetwork{
		deploymentID: deploymentID,
		ports:        ports,
		portMap:      portMap,
	}
}

func (n *fallbackContainerNetwork) Setup() error {
	// On macOS, port forwarding is handled by Colima's VM networking.
	// The host ports in portMap are directly reachable on localhost.
	return nil
}

func (n *fallbackContainerNetwork) Teardown() error {
	return nil
}

func (n *fallbackContainerNetwork) ResolvePort(containerPort int) (int, bool) {
	n.mu.RLock()
	defer n.mu.RUnlock()
	hp, ok := n.portMap[containerPort]
	return hp, ok
}

func (n *fallbackContainerNetwork) ContainerIP() string {
	return "127.0.0.1" // Colima host networking
}
