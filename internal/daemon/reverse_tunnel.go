package daemon

import (
	"context"
	"fmt"
	"sync"

	"github.com/moltbunker/moltbunker/internal/logging"
	"github.com/moltbunker/moltbunker/internal/tunnel"
	"github.com/moltbunker/moltbunker/internal/util"
)

// ReverseTunnelManager manages reverse tunnel connections for exposed deployments.
// Each deployment+port pair gets its own ReverseClient that maintains a persistent
// yamux session to the ingress node.
type ReverseTunnelManager struct {
	factory func() *tunnel.ReverseClient // creates a new client per connection

	mu     sync.Mutex
	active map[string]*reverseTunnelEntry // "deploymentID:port" → entry
	ctx    context.Context
	cancel context.CancelFunc
}

type reverseTunnelEntry struct {
	client    *tunnel.ReverseClient
	cancel    context.CancelFunc
	subdomain string
}

// NewReverseTunnelManager creates a manager that will create ReverseClient instances
// using the provided factory function.
func NewReverseTunnelManager(ctx context.Context, factory func() *tunnel.ReverseClient) *ReverseTunnelManager {
	ctx, cancel := context.WithCancel(ctx)
	return &ReverseTunnelManager{
		factory: factory,
		active:  make(map[string]*reverseTunnelEntry),
		ctx:     ctx,
		cancel:  cancel,
	}
}

// Expose connects a deployment+port to the ingress via a reverse tunnel.
// Runs asynchronously — the tunnel is established in the background.
// Returns immediately.
func (m *ReverseTunnelManager) Expose(deploymentID string, containerPort int) {
	key := fmt.Sprintf("%s:%d", deploymentID, containerPort)

	m.mu.Lock()
	if _, exists := m.active[key]; exists {
		m.mu.Unlock()
		return // Already exposed
	}

	client := m.factory()
	entryCtx, entryCancel := context.WithCancel(m.ctx)
	entry := &reverseTunnelEntry{
		client: client,
		cancel: entryCancel,
	}
	m.active[key] = entry
	m.mu.Unlock()

	util.SafeGoWithName("reverse-tunnel-"+key, func() {
		subdomain, err := client.Connect(entryCtx, deploymentID, containerPort)
		if err != nil && entryCtx.Err() == nil {
			logging.Error("reverse tunnel connection failed",
				"deployment_id", deploymentID,
				"port", containerPort,
				logging.Err(err),
				logging.Component("reverse-tunnel"))
		}

		m.mu.Lock()
		if e, ok := m.active[key]; ok {
			e.subdomain = subdomain
		}
		m.mu.Unlock()

		if subdomain != "" {
			logging.Info("reverse tunnel exposed",
				"deployment_id", deploymentID,
				"port", containerPort,
				"subdomain", subdomain,
				logging.Component("reverse-tunnel"))
		}
	})
}

// Unexpose disconnects all reverse tunnels for a deployment.
func (m *ReverseTunnelManager) Unexpose(deploymentID string) {
	m.mu.Lock()
	var toRemove []string
	for key, entry := range m.active {
		// Key format: "deploymentID:port" — check prefix
		if len(key) > len(deploymentID) && key[:len(deploymentID)] == deploymentID && key[len(deploymentID)] == ':' {
			entry.cancel()
			if err := entry.client.Disconnect(); err != nil {
				logging.Warn("failed to disconnect reverse tunnel client",
					"deployment_id", deploymentID,
					"key", key,
					logging.Err(err),
					logging.Component("reverse-tunnel"))
			}
			toRemove = append(toRemove, key)
		}
	}
	for _, key := range toRemove {
		delete(m.active, key)
	}
	m.mu.Unlock()

	if len(toRemove) > 0 {
		logging.Info("reverse tunnel unexposed",
			"deployment_id", deploymentID,
			"tunnels_closed", len(toRemove),
			logging.Component("reverse-tunnel"))
	}
}

// Subdomain returns the assigned subdomain for a deployment+port, or empty if not connected.
func (m *ReverseTunnelManager) Subdomain(deploymentID string, containerPort int) string {
	key := fmt.Sprintf("%s:%d", deploymentID, containerPort)
	m.mu.Lock()
	defer m.mu.Unlock()
	if entry, ok := m.active[key]; ok {
		return entry.subdomain
	}
	return ""
}

// ActiveCount returns the number of active reverse tunnel connections.
func (m *ReverseTunnelManager) ActiveCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.active)
}

// Stop disconnects all active reverse tunnels.
func (m *ReverseTunnelManager) Stop() {
	m.cancel()

	m.mu.Lock()
	for key, entry := range m.active {
		entry.cancel()
		if err := entry.client.Disconnect(); err != nil {
			logging.Warn("failed to disconnect reverse tunnel client during shutdown",
				"key", key,
				logging.Err(err),
				logging.Component("reverse-tunnel"))
		}
	}
	m.active = make(map[string]*reverseTunnelEntry)
	m.mu.Unlock()
}
