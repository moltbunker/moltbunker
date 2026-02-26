package ingress

import (
	"fmt"
	"strings"
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
	RuntimeType    string    `json:"runtime_type,omitempty"` // "container" (default) or "molt"
}

// GossipReader reads exposed service entries from the gossip state.
type GossipReader interface {
	// GetExposedServices returns all exposed service entries from gossip state.
	// Keys are formatted as "expose:<deploymentID>:<port>".
	GetExposedServices() map[string]*ServiceEntry
}

// SubdomainResolver resolves vanity subdomain names to deployment IDs.
type SubdomainResolver interface {
	// ResolveVanityName resolves a vanity name to a deployment ID via gossip state.
	ResolveVanityName(name string) (deploymentID string, ok bool)

	// ResolveOnChain resolves a vanity name to a deployment ID via on-chain lookup.
	// Used as a fallback when gossip doesn't have the mapping (cross-node routing).
	ResolveOnChain(name string) (deploymentID string, ok bool)
}

// Resolver maps subdomains to provider addresses using gossip state.
// Supports both deployment ID prefix matching (auto-assigned subdomains)
// and vanity name resolution via SubdomainResolver.
type Resolver struct {
	services          map[string]*ServiceEntry // deploymentID → service entry
	gossip            GossipReader
	subdomainResolver SubdomainResolver
	mu                sync.RWMutex
}

// NewResolver creates a service resolver.
// subdomainResolver may be nil to disable vanity name resolution.
func NewResolver(gossip GossipReader, subdomainResolver SubdomainResolver) *Resolver {
	return &Resolver{
		services:          make(map[string]*ServiceEntry),
		gossip:            gossip,
		subdomainResolver: subdomainResolver,
	}
}

// Resolve returns the service entry for a subdomain string.
// Resolution order:
//  1. Exact deployment ID match
//  2. Prefix match (8-char hex prefix from auto-assigned subdomains)
//  3. Vanity name resolution via SubdomainResolver
//  4. Refresh from gossip and retry steps 1-3
func (r *Resolver) Resolve(subdomain string) (*ServiceEntry, error) {
	// Step 1: Exact match
	r.mu.RLock()
	entry, ok := r.services[subdomain]
	r.mu.RUnlock()
	if ok && time.Since(entry.LastSeen) < 5*time.Minute {
		return entry, nil
	}

	// Step 2: Prefix match (subdomain is first 8 chars of deployment ID)
	if entry := r.resolveByPrefix(subdomain); entry != nil {
		return entry, nil
	}

	// Step 3: Vanity name resolution
	if entry := r.resolveVanity(subdomain); entry != nil {
		return entry, nil
	}

	// Step 4: Refresh from gossip and retry
	if r.gossip != nil {
		r.refreshFromGossip()

		// Retry prefix match
		if entry := r.resolveByPrefix(subdomain); entry != nil {
			return entry, nil
		}
		// Retry vanity
		if entry := r.resolveVanity(subdomain); entry != nil {
			return entry, nil
		}
	}

	// Step 5: On-chain fallback for cross-node vanity routing.
	// Gossip subdomain entries are local-only (rejected from remote peers to
	// prevent spoofing). If this ingress node didn't register the name, the
	// gossip lookup fails. Fall back to the BunkerRegistry contract as the
	// trusted source of truth.
	if entry := r.resolveOnChain(subdomain); entry != nil {
		return entry, nil
	}

	return nil, fmt.Errorf("service not found: %s", subdomain)
}

// minPrefixLen is the minimum subdomain length required for deployment ID
// prefix matching. Auto-assigned subdomains use the first 8 hex chars of
// the deployment ID. Without this floor, a 1-char subdomain like "a" would
// match any deployment starting with "a", enabling cross-tenant routing.
const minPrefixLen = 8

// resolveByPrefix finds a service whose deployment ID starts with the given prefix.
func (r *Resolver) resolveByPrefix(prefix string) *ServiceEntry {
	if len(prefix) < minPrefixLen {
		return nil
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	for depID, entry := range r.services {
		// Strip "dep-" prefix from deployment ID for matching
		bare := strings.TrimPrefix(depID, "dep-")
		if strings.HasPrefix(bare, prefix) && time.Since(entry.LastSeen) < 5*time.Minute {
			return entry
		}
	}
	return nil
}

// resolveVanity resolves a vanity name to a deployment ID, then looks up the service.
func (r *Resolver) resolveVanity(name string) *ServiceEntry {
	if r.subdomainResolver == nil {
		return nil
	}
	depID, ok := r.subdomainResolver.ResolveVanityName(name)
	if !ok || depID == "" {
		return nil
	}

	r.mu.RLock()
	entry, ok := r.services[depID]
	r.mu.RUnlock()
	if ok && time.Since(entry.LastSeen) < 5*time.Minute {
		return entry
	}
	return nil
}

// resolveOnChain resolves a vanity name via on-chain lookup, then finds the service.
func (r *Resolver) resolveOnChain(name string) *ServiceEntry {
	if r.subdomainResolver == nil {
		return nil
	}
	depID, ok := r.subdomainResolver.ResolveOnChain(name)
	if !ok || depID == "" {
		return nil
	}
	return r.findServiceByDeploymentID(depID)
}

// findServiceByDeploymentID searches for a service entry by deployment ID (exact or prefix).
func (r *Resolver) findServiceByDeploymentID(depID string) *ServiceEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Exact match first
	if entry, ok := r.services[depID]; ok && time.Since(entry.LastSeen) < 5*time.Minute {
		return entry
	}

	// Prefix match (deployment IDs may have "dep-" prefix in gossip)
	for key, entry := range r.services {
		bare := strings.TrimPrefix(key, "dep-")
		if bare == depID && time.Since(entry.LastSeen) < 5*time.Minute {
			return entry
		}
	}
	return nil
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
