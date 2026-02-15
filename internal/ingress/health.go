package ingress

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/moltbunker/moltbunker/internal/logging"
	"github.com/moltbunker/moltbunker/internal/tunnel"
)

// ServiceHealth tracks the health status of an exposed service.
type ServiceHealth struct {
	DeploymentID string    `json:"deployment_id"`
	Healthy      bool      `json:"healthy"`
	LastCheck    time.Time `json:"last_check"`
	LastError    string    `json:"last_error,omitempty"`
	Latency      time.Duration `json:"latency"`
	CheckCount   int       `json:"check_count"`
	FailCount    int       `json:"fail_count"`
}

// HealthChecker periodically checks the health of exposed services
// by opening tunnel connections and verifying connectivity.
type HealthChecker struct {
	resolver     *Resolver
	tunnelClient *tunnel.Client
	interval     time.Duration
	timeout      time.Duration

	statuses map[string]*ServiceHealth
	mu       sync.RWMutex
	cancel   context.CancelFunc
}

// NewHealthChecker creates a service health checker.
func NewHealthChecker(resolver *Resolver, tunnelClient *tunnel.Client) *HealthChecker {
	return &HealthChecker{
		resolver:     resolver,
		tunnelClient: tunnelClient,
		interval:     30 * time.Second,
		timeout:      10 * time.Second,
		statuses:     make(map[string]*ServiceHealth),
	}
}

// Start begins periodic health checking. Call Stop() to terminate.
func (hc *HealthChecker) Start(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	hc.cancel = cancel

	go func() {
		ticker := time.NewTicker(hc.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				hc.checkAll(ctx)
			}
		}
	}()
}

// Stop terminates the health checker.
func (hc *HealthChecker) Stop() {
	if hc.cancel != nil {
		hc.cancel()
	}
}

// GetStatus returns the health status for a deployment.
func (hc *HealthChecker) GetStatus(deploymentID string) (*ServiceHealth, bool) {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	s, ok := hc.statuses[deploymentID]
	return s, ok
}

// IsHealthy returns true if the service is healthy.
func (hc *HealthChecker) IsHealthy(deploymentID string) bool {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	s, ok := hc.statuses[deploymentID]
	return ok && s.Healthy
}

// checkAll performs health checks on all known services.
func (hc *HealthChecker) checkAll(ctx context.Context) {
	hc.mu.RLock()
	services := make(map[string]*ServiceEntry)
	for id, entry := range hc.resolver.services {
		services[id] = entry
	}
	hc.mu.RUnlock()

	for _, entry := range services {
		hc.checkService(ctx, entry)
	}
}

// checkService verifies connectivity to a single service.
func (hc *HealthChecker) checkService(ctx context.Context, entry *ServiceEntry) {
	ctx, cancel := context.WithTimeout(ctx, hc.timeout)
	defer cancel()

	start := time.Now()
	_ = ctx // used for timeout deadline

	tun, err := hc.tunnelClient.OpenTunnel(entry.ProviderAddr, entry.DeploymentID, entry.ContainerPort)

	hc.mu.Lock()
	defer hc.mu.Unlock()

	status, ok := hc.statuses[entry.DeploymentID]
	if !ok {
		status = &ServiceHealth{DeploymentID: entry.DeploymentID}
		hc.statuses[entry.DeploymentID] = status
	}
	status.CheckCount++
	status.LastCheck = time.Now()
	status.Latency = time.Since(start)

	if err != nil {
		status.Healthy = false
		status.FailCount++
		status.LastError = fmt.Sprintf("tunnel open: %v", err)
		logging.Debug("service health check failed",
			"deployment_id", entry.DeploymentID,
			"error", err.Error(),
			logging.Component("ingress"))
		return
	}
	tun.Close()

	status.Healthy = true
	status.LastError = ""
}
