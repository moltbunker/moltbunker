package p2p

import (
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/network"
	rcmgr "github.com/libp2p/go-libp2p/p2p/host/resource-manager"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/moltbunker/moltbunker/internal/logging"
)

// NewResourceManager creates a libp2p resource manager with limits configured
// to prevent resource exhaustion from connection flooding. The limits protect
// the node from DoS attacks by capping connections, streams, and memory usage
// at the system, transient, per-peer, and per-protocol levels.
//
// Prometheus metrics for the resource manager are enabled by default and
// registered with the default Prometheus registry.
func NewResourceManager() (network.ResourceManager, error) {
	return NewResourceManagerWithRegistry(nil)
}

// NewResourceManagerWithRegistry creates a resource manager and registers
// libp2p resource manager Prometheus metrics with the given registry. If reg
// is nil, the default Prometheus registry is used.
func NewResourceManagerWithRegistry(reg prometheus.Registerer) (network.ResourceManager, error) {
	// Start with default scaling limits and add libp2p protocol limits
	scalingLimits := rcmgr.DefaultLimits
	libp2p.SetDefaultServiceLimits(&scalingLimits)

	// Override with moltbunker-specific base limits:
	//   System:       256 connections, 512 streams, 128MB memory
	//   Transient:    64 connections, 128 streams, 32MB memory
	//   Per-peer:     16 connections, 64 streams, 8MB memory
	//   Per-protocol: 32 streams
	scalingLimits.SystemBaseLimit = rcmgr.BaseLimit{
		Conns:           256,
		ConnsInbound:    128,
		ConnsOutbound:   256,
		Streams:         512,
		StreamsInbound:  256,
		StreamsOutbound: 512,
		Memory:          128 << 20, // 128 MB
		FD:              256,
	}
	scalingLimits.TransientBaseLimit = rcmgr.BaseLimit{
		Conns:           64,
		ConnsInbound:    32,
		ConnsOutbound:   64,
		Streams:         128,
		StreamsInbound:  64,
		StreamsOutbound: 128,
		Memory:          32 << 20, // 32 MB
		FD:              64,
	}
	scalingLimits.PeerBaseLimit = rcmgr.BaseLimit{
		Conns:           16,
		ConnsInbound:    8,
		ConnsOutbound:   16,
		Streams:         64,
		StreamsInbound:  32,
		StreamsOutbound: 64,
		Memory:          8 << 20, // 8 MB
		FD:              16,
	}
	scalingLimits.ProtocolBaseLimit = rcmgr.BaseLimit{
		Streams:         32,
		StreamsInbound:  16,
		StreamsOutbound: 32,
	}

	// Auto-scale the limits based on available system memory and file descriptors
	scaledLimits := scalingLimits.AutoScale()

	// Build a fixed limiter from the scaled limits
	limiter := rcmgr.NewFixedLimiter(scaledLimits)

	// Register libp2p resource manager metrics with Prometheus.
	// MustRegisterWith registers connection, stream, memory, and FD
	// gauges/histograms under the "libp2p_rcmgr" namespace.
	if reg == nil {
		reg = prometheus.DefaultRegisterer
	}
	rcmgr.MustRegisterWith(reg)

	// Create the resource manager with metrics enabled (default behaviour
	// when WithMetricsDisabled is not passed). The rcmgr automatically sets
	// up a StatsTraceReporter that feeds the registered Prometheus metrics.
	rm, err := rcmgr.NewResourceManager(limiter)
	if err != nil {
		return nil, err
	}

	logging.Info("resource manager initialized with Prometheus metrics",
		"system_conns", 256,
		"system_streams", 512,
		"system_memory_mb", 128,
		"transient_conns", 64,
		"peer_conns", 16,
		"protocol_streams", 32,
	)

	return rm, nil
}
