package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/moltbunker/moltbunker/internal/ingress"
	"github.com/moltbunker/moltbunker/internal/logging"
	"github.com/moltbunker/moltbunker/internal/p2p"
	"github.com/moltbunker/moltbunker/internal/payment"
	"github.com/moltbunker/moltbunker/pkg/types"
)

// cachedSubdomainResult stores a cached on-chain subdomain resolution result.
type cachedSubdomainResult struct {
	DeploymentID string
	Found        bool
	FetchedAt    time.Time
}

// Default cache TTLs for on-chain subdomain resolution.
const (
	defaultSubdomainCacheTTL         = 5 * time.Minute // positive results
	defaultSubdomainNegativeCacheTTL = 2 * time.Minute // negative (not found) results
)

// GossipServiceAdapter bridges p2p.GossipProtocol to ingress.GossipReader.
// It reads "expose:<deploymentID>:<port>" entries from gossip state and
// converts them to ingress.ServiceEntry structs.
//
// It also implements SubdomainResolver by reading "subdomain:<name>" entries
// from gossip state to resolve vanity names to deployment IDs.
type GossipServiceAdapter struct {
	gossip         *p2p.GossipProtocol
	paymentService *payment.PaymentService

	// On-chain subdomain resolution cache
	subdomainCache map[string]*cachedSubdomainResult
	cacheMu        sync.RWMutex
	cacheTTL       time.Duration // TTL for positive results
	negativeTTL    time.Duration // TTL for negative (not found) results
	nowFunc        func() time.Time
}

// NewGossipServiceAdapter creates a new gossip service adapter.
func NewGossipServiceAdapter(gossip *p2p.GossipProtocol) *GossipServiceAdapter {
	return &GossipServiceAdapter{
		gossip:         gossip,
		subdomainCache: make(map[string]*cachedSubdomainResult),
		cacheTTL:       defaultSubdomainCacheTTL,
		negativeTTL:    defaultSubdomainNegativeCacheTTL,
		nowFunc:        time.Now,
	}
}

// SetPaymentService sets the payment service for on-chain subdomain resolution.
func (a *GossipServiceAdapter) SetPaymentService(ps *payment.PaymentService) {
	a.paymentService = ps
}

// GetExposedServices implements ingress.GossipReader.
// It reads all "expose:*" entries from gossip state. Values may be either
// *ingress.ServiceEntry (local updates) or map[string]interface{} (after
// JSON round-trip through gossip sync). Both cases are handled.
func (a *GossipServiceAdapter) GetExposedServices() map[string]*ingress.ServiceEntry {
	raw := a.gossip.GetStateByPrefix("expose:")
	result := make(map[string]*ingress.ServiceEntry, len(raw))

	for key, val := range raw {
		if val == nil {
			continue // Deleted entry (nil = removed)
		}

		entry := a.toServiceEntry(val)
		if entry == nil {
			continue
		}
		result[key] = entry
	}
	return result
}

// ResolveVanityName implements ingress.SubdomainResolver.
// It reads "subdomain:<name>" from gossip state and returns the deployment ID.
func (a *GossipServiceAdapter) ResolveVanityName(name string) (string, bool) {
	val, ok := a.gossip.GetState("subdomain:" + name)
	if !ok || val == nil {
		return "", false
	}
	if depID, ok := val.(string); ok {
		return depID, true
	}
	return "", false
}

// ResolveOnChain implements ingress.SubdomainResolver.
// It queries the BunkerRegistry smart contract as a fallback for cross-node
// vanity routing when gossip state doesn't have the mapping.
// Results are cached with separate TTLs for positive (5min) and negative (2min) hits
// to avoid redundant RPC calls for popular subdomains.
func (a *GossipServiceAdapter) ResolveOnChain(name string) (string, bool) {
	now := a.nowFunc()

	// Check cache first — return cached result even if paymentService is nil
	a.cacheMu.RLock()
	cached, ok := a.subdomainCache[name]
	a.cacheMu.RUnlock()

	if ok {
		ttl := a.cacheTTL
		if !cached.Found {
			ttl = a.negativeTTL
		}
		if now.Sub(cached.FetchedAt) < ttl {
			return cached.DeploymentID, cached.Found
		}
	}

	// Cache miss or expired — need payment service for RPC call
	if a.paymentService == nil {
		return "", false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	reg, err := a.paymentService.ResolveSubdomain(ctx, name)
	if err != nil || reg == nil {
		logging.Debug("on-chain subdomain resolution failed",
			"name", name,
			logging.Err(err),
			logging.Component("ingress"))
		// Cache negative result
		a.cacheMu.Lock()
		a.subdomainCache[name] = &cachedSubdomainResult{Found: false, FetchedAt: now}
		a.cacheMu.Unlock()
		return "", false
	}
	depID := bytes32ToDeploymentID(reg.DeploymentID)
	if depID == "" || depID == fmt.Sprintf("%x", [32]byte{}) {
		a.cacheMu.Lock()
		a.subdomainCache[name] = &cachedSubdomainResult{Found: false, FetchedAt: now}
		a.cacheMu.Unlock()
		return "", false
	}

	// Cache positive result
	a.cacheMu.Lock()
	a.subdomainCache[name] = &cachedSubdomainResult{
		DeploymentID: depID,
		Found:        true,
		FetchedAt:    now,
	}
	a.cacheMu.Unlock()
	return depID, true
}

// InvalidateSubdomainCache evicts a cached on-chain resolution result,
// forcing the next ResolveOnChain call to re-query the contract.
func (a *GossipServiceAdapter) InvalidateSubdomainCache(name string) {
	a.cacheMu.Lock()
	delete(a.subdomainCache, name)
	a.cacheMu.Unlock()
}

// toServiceEntry converts a gossip value to *ingress.ServiceEntry.
// After gossip JSON round-trip, struct values become map[string]interface{}.
// This re-marshals through JSON to handle both original structs and maps.
func (a *GossipServiceAdapter) toServiceEntry(val interface{}) *ingress.ServiceEntry {
	// Fast path: already a ServiceEntry (local update, no round-trip)
	if entry, ok := val.(*ingress.ServiceEntry); ok {
		return entry
	}

	// Slow path: re-marshal through JSON (gossip sync round-trip)
	data, err := json.Marshal(val)
	if err != nil {
		return nil
	}

	var entry ingress.ServiceEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil
	}

	// Validate minimum required fields
	if entry.DeploymentID == "" || entry.ProviderAddr == "" {
		return nil
	}
	return &entry
}

// ExposeKeyParts splits an "expose:<deploymentID>:<port>" key.
func ExposeKeyParts(key string) (deploymentID string, port string, ok bool) {
	parts := strings.SplitN(key, ":", 3)
	if len(parts) != 3 || parts[0] != "expose" {
		return "", "", false
	}
	return parts[1], parts[2], true
}

// NewGossipStateValidator returns a StateValidator that enforces authorization
// on security-sensitive gossip keys:
//   - "expose:*" entries: the ProviderNodeID in the ServiceEntry must match
//     the sender's NodeID (only the actual provider can advertise its services)
//   - "subdomain:*" entries: rejected from all remote peers. Subdomain mappings
//     are only set locally via handleSubdomainRegister, which verifies on-chain
//     ownership before writing to gossip state. This prevents subdomain hijacking
//     where a malicious peer injects a gossip entry to redirect traffic.
//   - All other keys: accepted (default gossip behavior)
func NewGossipStateValidator(localNodeID types.NodeID) p2p.StateValidator {
	return func(senderID types.NodeID, key string, value interface{}) bool {
		// Local updates (senderID is zero) always pass
		if senderID == (types.NodeID{}) {
			return true
		}

		if strings.HasPrefix(key, "expose:") {
			return validateExposeEntry(senderID, value)
		}

		// subdomain: keys must only come from the local node (via
		// handleSubdomainRegister, which verifies on-chain ownership).
		// Reject all remote subdomain entries to prevent gossip-based
		// subdomain hijacking — a malicious peer could inject
		// "subdomain:victim-app" pointing to their own deployment.
		if strings.HasPrefix(key, "subdomain:") {
			return false
		}

		return true
	}
}

// validateExposeEntry checks that the ProviderNodeID in a ServiceEntry matches the sender.
func validateExposeEntry(senderID types.NodeID, value interface{}) bool {
	if value == nil {
		return true // Deletion (nil) is allowed
	}

	// Extract ProviderNodeID from the value (may be struct or map after JSON round-trip)
	var providerNodeID string

	switch v := value.(type) {
	case *ingress.ServiceEntry:
		providerNodeID = v.ProviderNodeID
	case map[string]interface{}:
		if nid, ok := v["provider_node_id"].(string); ok {
			providerNodeID = nid
		}
	default:
		return false // Unknown type — reject
	}

	if providerNodeID == "" {
		return false // Missing provider identity — reject
	}

	// The ProviderNodeID in the entry must match the sender's NodeID
	return providerNodeID == senderID.String()
}
