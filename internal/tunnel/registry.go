package tunnel

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/hashicorp/yamux"
	"github.com/moltbunker/moltbunker/pkg/types"
)

// TunnelSession represents a single reverse tunnel session between a provider and the ingress.
type TunnelSession struct {
	NodeID       types.NodeID
	Subdomain    string
	DeploymentID string
	YamuxSess    *yamux.Session
	RegisteredAt time.Time
	Tier         string // "free", "starter", "bronze", "silver", "gold", "platinum"
	Limits       *TunnelLimits
	ReconnToken  string
}

// TunnelRegistry is a thread-safe subdomain → yamux session mapping.
type TunnelRegistry struct {
	mu       sync.RWMutex
	sessions map[string]*TunnelSession  // subdomain → session
	byNodeID map[types.NodeID][]string  // nodeID → subdomains
}

// NewTunnelRegistry creates an empty tunnel registry.
func NewTunnelRegistry() *TunnelRegistry {
	return &TunnelRegistry{
		sessions: make(map[string]*TunnelSession),
		byNodeID: make(map[types.NodeID][]string),
	}
}

// Register adds a session to the registry.
// Returns an error if the subdomain is already taken.
func (r *TunnelRegistry) Register(subdomain string, session *TunnelSession) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.sessions[subdomain]; exists {
		return fmt.Errorf("subdomain %q already registered", subdomain)
	}

	r.sessions[subdomain] = session
	r.byNodeID[session.NodeID] = append(r.byNodeID[session.NodeID], subdomain)
	return nil
}

// Unregister removes a subdomain from the registry.
func (r *TunnelRegistry) Unregister(subdomain string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	sess, exists := r.sessions[subdomain]
	if !exists {
		return
	}

	delete(r.sessions, subdomain)

	// Remove from byNodeID index
	subs := r.byNodeID[sess.NodeID]
	for i, s := range subs {
		if s == subdomain {
			r.byNodeID[sess.NodeID] = append(subs[:i], subs[i+1:]...)
			break
		}
	}
	if len(r.byNodeID[sess.NodeID]) == 0 {
		delete(r.byNodeID, sess.NodeID)
	}
}

// Lookup returns the session for a subdomain, or nil if not found.
func (r *TunnelRegistry) Lookup(subdomain string) (*TunnelSession, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	sess, ok := r.sessions[subdomain]
	return sess, ok
}

// CountForNodeID returns the number of active subdomains for a NodeID.
func (r *TunnelRegistry) CountForNodeID(nodeID types.NodeID) int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.byNodeID[nodeID])
}

// SubdomainsForNodeID returns all subdomains registered by a NodeID.
func (r *TunnelRegistry) SubdomainsForNodeID(nodeID types.NodeID) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	subs := r.byNodeID[nodeID]
	out := make([]string, len(subs))
	copy(out, subs)
	return out
}

// UnregisterAll removes all subdomains for a NodeID and returns them.
func (r *TunnelRegistry) UnregisterAll(nodeID types.NodeID) []string {
	r.mu.Lock()
	defer r.mu.Unlock()

	subs := r.byNodeID[nodeID]
	for _, s := range subs {
		delete(r.sessions, s)
	}
	delete(r.byNodeID, nodeID)
	return subs
}

// ActiveCount returns the total number of active sessions.
func (r *TunnelRegistry) ActiveCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.sessions)
}

// AssignRandomSubdomain generates an 8-char hex subdomain via crypto/rand
// and ensures it doesn't collide with existing entries.
func (r *TunnelRegistry) AssignRandomSubdomain() (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for attempts := 0; attempts < 10; attempts++ {
		var buf [4]byte
		if _, err := rand.Read(buf[:]); err != nil {
			return "", fmt.Errorf("generate random subdomain: %w", err)
		}
		sub := hex.EncodeToString(buf[:])
		if _, exists := r.sessions[sub]; !exists {
			return sub, nil
		}
	}
	return "", fmt.Errorf("failed to generate unique subdomain after 10 attempts")
}
