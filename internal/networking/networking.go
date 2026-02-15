// Package networking manages container network setup for service exposure.
// It handles veth pair creation (Linux), port allocation, and nftables forwarding.
package networking

import (
	"fmt"
	"sync"
)

// ContainerNetwork manages the network namespace and port forwarding for a single container.
type ContainerNetwork interface {
	// Setup creates the network namespace, veth pair, and port forwarding rules.
	Setup() error
	// Teardown removes network resources (veth, nftables rules, netns).
	Teardown() error
	// ResolvePort returns the host port mapped to a given container port.
	ResolvePort(containerPort int) (hostPort int, ok bool)
	// ContainerIP returns the container's IP address in the private network.
	ContainerIP() string
}

// ExposedPort describes a port that should be publicly reachable.
type ExposedPort struct {
	ContainerPort int    `json:"container_port"`
	HostPort      int    `json:"host_port,omitempty"` // 0 = auto-assign
	Protocol      string `json:"protocol,omitempty"`  // "tcp" (default) or "udp"
}

// ExposedService is gossiped by providers to announce publicly exposed container services.
// Ingress nodes use this to route incoming HTTP requests to the correct provider.
type ExposedService struct {
	DeploymentID   string `json:"deployment_id"`
	ProviderNodeID string `json:"provider_node_id"`
	ProviderAddr   string `json:"provider_addr"` // host:port for tunnel connections
	ContainerPort  int    `json:"container_port"`
	HostPort       int    `json:"host_port"`
	Domain         string `json:"domain,omitempty"` // e.g., "<id>.moltbunker.dev"
}

// NetworkManager coordinates container networks and port allocations across all deployments.
type NetworkManager struct {
	networks  map[string]ContainerNetwork // deploymentID → network
	portAlloc *PortAllocator
	mu        sync.RWMutex
}

// NewNetworkManager creates a new network manager.
func NewNetworkManager() *NetworkManager {
	return &NetworkManager{
		networks:  make(map[string]ContainerNetwork),
		portAlloc: NewPortAllocator(49152, 65535),
	}
}

// SetupNetwork creates a container network for the given deployment and exposed ports.
func (nm *NetworkManager) SetupNetwork(deploymentID string, ports []ExposedPort) (ContainerNetwork, error) {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	if _, exists := nm.networks[deploymentID]; exists {
		return nil, fmt.Errorf("network already exists for deployment %s", deploymentID)
	}

	// Allocate host ports
	allocated := make([]ExposedPort, len(ports))
	for i, p := range ports {
		allocated[i] = p
		if p.HostPort == 0 {
			hostPort, err := nm.portAlloc.Allocate()
			if err != nil {
				// Release any previously allocated ports
				for j := 0; j < i; j++ {
					if ports[j].HostPort == 0 {
						nm.portAlloc.Release(allocated[j].HostPort)
					}
				}
				return nil, fmt.Errorf("allocate host port for container port %d: %w", p.ContainerPort, err)
			}
			allocated[i].HostPort = hostPort
		}
		if allocated[i].Protocol == "" {
			allocated[i].Protocol = "tcp"
		}
	}

	net := newContainerNetwork(deploymentID, allocated)
	if err := net.Setup(); err != nil {
		for _, p := range allocated {
			if p.HostPort != 0 {
				nm.portAlloc.Release(p.HostPort)
			}
		}
		return nil, fmt.Errorf("setup network: %w", err)
	}

	nm.networks[deploymentID] = net
	return net, nil
}

// TeardownNetwork removes the container network for the given deployment.
func (nm *NetworkManager) TeardownNetwork(deploymentID string) error {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	net, exists := nm.networks[deploymentID]
	if !exists {
		return nil
	}

	if err := net.Teardown(); err != nil {
		return fmt.Errorf("teardown network: %w", err)
	}

	delete(nm.networks, deploymentID)
	return nil
}

// GetNetwork returns the container network for a deployment.
func (nm *NetworkManager) GetNetwork(deploymentID string) (ContainerNetwork, bool) {
	nm.mu.RLock()
	defer nm.mu.RUnlock()
	net, ok := nm.networks[deploymentID]
	return net, ok
}
