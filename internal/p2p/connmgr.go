package p2p

import (
	"time"

	"github.com/libp2p/go-libp2p/core/connmgr"
	libp2pconnmgr "github.com/libp2p/go-libp2p/p2p/net/connmgr"

	"github.com/moltbunker/moltbunker/internal/logging"
)

const (
	// DefaultConnManagerLowWatermark is the number of connections below which
	// the connection manager will not prune connections.
	DefaultConnManagerLowWatermark = 100

	// DefaultConnManagerHighWatermark is the number of connections above which
	// the connection manager will start pruning least-useful connections.
	DefaultConnManagerHighWatermark = 400

	// DefaultConnManagerGracePeriod is the duration a new connection is immune
	// from pruning after being opened.
	DefaultConnManagerGracePeriod = 20 * time.Second
)

// NewConnectionManager creates a libp2p connection manager with watermark-based
// pruning. When the number of connections exceeds highWatermark, the manager
// prunes least-useful connections down to lowWatermark. New connections are
// protected from pruning for the duration of gracePeriod.
func NewConnectionManager(lowWatermark, highWatermark int, gracePeriod time.Duration) (connmgr.ConnManager, error) {
	cm, err := libp2pconnmgr.NewConnManager(
		lowWatermark,
		highWatermark,
		libp2pconnmgr.WithGracePeriod(gracePeriod),
	)
	if err != nil {
		return nil, err
	}

	logging.Info("connection manager initialized",
		"low_watermark", lowWatermark,
		"high_watermark", highWatermark,
		"grace_period", gracePeriod.String(),
	)

	return cm, nil
}
