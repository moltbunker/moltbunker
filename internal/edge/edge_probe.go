package edge

import (
	"context"
	"crypto/tls"
	"net"
	"sync"
	"time"

	"github.com/moltbunker/moltbunker/internal/logging"
	"github.com/moltbunker/moltbunker/internal/util"
	"github.com/moltbunker/moltbunker/pkg/types"
)

const (
	// defaultProbeInterval is how often each edge node is probed.
	defaultProbeInterval = 30 * time.Second
	// defaultProbeTimeout bounds a single dial+handshake.
	defaultProbeTimeout = 5 * time.Second
	// probeFailThreshold is how many consecutive failed probes flip a node
	// unhealthy. A single success resets the counter and marks it healthy.
	probeFailThreshold = 3
)

// EdgeProbe is a background goroutine that periodically TLS-dials each
// registered edge node's tunnel port to verify liveness, then updates the
// registry's health flag. It is a pure connectivity check — no yamux, no
// authentication — mirroring the ingress-side intent of the reverse server's
// heartbeat but pointed outward at the edge set.
type EdgeProbe struct {
	registry *EdgeRegistry
	tlsCfg   *tls.Config
	interval time.Duration
	timeout  time.Duration

	mu     sync.Mutex
	misses map[types.NodeID]int
	cancel context.CancelFunc
	done   chan struct{}
}

// NewEdgeProbe creates a probe. Non-positive interval/timeout fall back to the
// package defaults. A nil tlsCfg uses InsecureSkipVerify=false defaults (a real
// edge cert is expected in production); tests inject a config trusting the
// self-signed test cert.
func NewEdgeProbe(registry *EdgeRegistry, tlsCfg *tls.Config, interval, timeout time.Duration) *EdgeProbe {
	if interval <= 0 {
		interval = defaultProbeInterval
	}
	if timeout <= 0 {
		timeout = defaultProbeTimeout
	}
	return &EdgeProbe{
		registry: registry,
		tlsCfg:   tlsCfg,
		interval: interval,
		timeout:  timeout,
		misses:   make(map[types.NodeID]int),
	}
}

// Start launches the probe loop until ctx is cancelled or Stop is called.
func (p *EdgeProbe) Start(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	p.mu.Lock()
	p.cancel = cancel
	p.done = make(chan struct{})
	p.mu.Unlock()

	util.SafeGoWithName("edge-probe", func() {
		defer close(p.done)
		ticker := time.NewTicker(p.interval)
		defer ticker.Stop()
		// Probe once immediately so health is fresh without waiting a full tick.
		p.probeAll(ctx)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				p.probeAll(ctx)
			}
		}
	})
}

// Stop cancels the probe loop and waits for it to exit.
func (p *EdgeProbe) Stop() {
	p.mu.Lock()
	cancel := p.cancel
	done := p.done
	p.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	if done != nil {
		<-done
	}
}

// probeAll probes every registered node once and updates health.
func (p *EdgeProbe) probeAll(ctx context.Context) {
	for _, n := range p.registry.ListAll() {
		if ctx.Err() != nil {
			return
		}
		p.probeOne(ctx, n)
	}
}

// probeOne dials a single node. On success the miss counter resets and the node
// is marked healthy; on failure the counter increments and the node is marked
// unhealthy once it reaches probeFailThreshold consecutive misses.
func (p *EdgeProbe) probeOne(ctx context.Context, n EdgeNodeInfo) {
	if p.dial(ctx, n.FullAddr()) {
		p.mu.Lock()
		p.misses[n.NodeID] = 0
		p.mu.Unlock()
		p.registry.UpdateHealth(n.NodeID, true)
		return
	}

	p.mu.Lock()
	p.misses[n.NodeID]++
	missed := p.misses[n.NodeID]
	p.mu.Unlock()

	if missed >= probeFailThreshold {
		logging.Warn("edge node marked unhealthy",
			"node_id", n.NodeID.String()[:16],
			"addr", n.FullAddr(),
			"missed", missed,
			logging.Component("edge"))
		p.registry.UpdateHealth(n.NodeID, false)
	}
}

// dial performs a single TCP+TLS connectivity check, returning true on a
// completed handshake. It never returns an error to the caller — failure is
// just "not reachable right now".
func (p *EdgeProbe) dial(ctx context.Context, addr string) bool {
	dialCtx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	dialer := &tls.Dialer{
		NetDialer: &net.Dialer{Timeout: p.timeout},
		Config:    p.tlsCfg,
	}
	conn, err := dialer.DialContext(dialCtx, "tcp", addr)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}
