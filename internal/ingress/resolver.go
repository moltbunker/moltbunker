package ingress

import (
	"fmt"
	"sync"
	"time"
)

// ServiceEntry describes a resolved exposed service.
type ServiceEntry struct {
	DeploymentID   string    `json:"deployment_id"`
	ProviderNodeID string    `json:"provider_node_id"`
	ProviderAddr   string    `json:"provider_addr"` // host:port for tunnel connection
	ContainerPort  int       `json:"container_port"`
	HostPort       int       `json:"host_port"`
	LastSeen       time.Time `json:"last_seen"`
}

// GossipReader reads exposed service entries from the gossip state.
type GossipReader interface {
	// GetExposedServices returns all exposed service entries from gossip state.
	// Keys are formatted as "expose:<deploymentID>:<port>".
	GetExposedServices() map[string]*ServiceEntry
}

// Resolver maps deployment IDs to provider addresses using gossip state.
// Ingress nodes participate in gossip and automatically receive service
// location updates from providers.
type Resolver struct {
	services map[string]*ServiceEntry // deploymentID → service entry
	gossip   GossipReader
	mu       sync.RWMutex
}

// NewResolver creates a service resolver.
func NewResolver(gossip GossipReader) *Resolver {
	return &Resolver{
		services: make(map[string]*ServiceEntry),
		gossip:   gossip,
	}
}

// Resolve returns the service entry for a deployment ID.
func (r *Resolver) Resolve(deploymentID string) (*ServiceEntry, error) {
	r.mu.RLock()
	entry, ok := r.services[deploymentID]
	r.mu.RUnlock()

	if ok && time.Since(entry.LastSeen) < 5*time.Minute {
		return entry, nil
	}

	// Refresh from gossip
	if r.gossip != nil {
		r.refreshFromGossip()
		r.mu.RLock()
		entry, ok = r.services[deploymentID]
		r.mu.RUnlock()
		if ok {
			return entry, nil
		}
	}

	return nil, fmt.Errorf("service not found: %s", deploymentID)
}

// refreshFromGossip updates the local service cache from gossip state.
func (r *Resolver) refreshFromGossip() {
	if r.gossip == nil {
		return
	}

	entries := r.gossip.GetExposedServices()
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, entry := range entries {
		r.services[entry.DeploymentID] = entry
	}
}

// Register manually adds a service entry (used by providers to register their own services).
func (r *Resolver) Register(entry *ServiceEntry) {
	r.mu.Lock()
	defer r.mu.Unlock()
	entry.LastSeen = time.Now()
	r.services[entry.DeploymentID] = entry
}

// Unregister removes a service entry.
func (r *Resolver) Unregister(deploymentID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.services, deploymentID)
}

// Count returns the number of known services.
func (r *Resolver) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.services)
}
