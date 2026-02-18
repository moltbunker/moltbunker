package daemon

import (
	"fmt"

	"github.com/moltbunker/moltbunker/internal/networking"
)

// DeploymentPortResolver implements tunnel.PortResolver using the NetworkManager.
// It maps (deploymentID, containerPort) → "127.0.0.1:<hostPort>" for the tunnel
// server to connect to local containers.
type DeploymentPortResolver struct {
	networkManager *networking.NetworkManager
}

// NewDeploymentPortResolver creates a new port resolver.
func NewDeploymentPortResolver(nm *networking.NetworkManager) *DeploymentPortResolver {
	return &DeploymentPortResolver{networkManager: nm}
}

// ResolveDeploymentPort implements tunnel.PortResolver.
func (r *DeploymentPortResolver) ResolveDeploymentPort(deploymentID string, containerPort int) (string, error) {
	net, ok := r.networkManager.GetNetwork(deploymentID)
	if !ok {
		return "", fmt.Errorf("network not found for deployment: %s", deploymentID)
	}

	hostPort, ok := net.ResolvePort(containerPort)
	if !ok {
		return "", fmt.Errorf("port %d not exposed for deployment: %s", containerPort, deploymentID)
	}

	return fmt.Sprintf("127.0.0.1:%d", hostPort), nil
}
