package proxy

import (
	"context"
	"net"

	"github.com/moltbunker/moltbunker/internal/tor"
)

// TorDialer routes connections through the Tor network.
type TorDialer struct {
	client *tor.TorClient
}

// NewTorDialer creates a dialer that routes traffic through Tor.
func NewTorDialer(socksAddr string) (*TorDialer, error) {
	client, err := tor.NewTorClient(socksAddr)
	if err != nil {
		return nil, err
	}
	return &TorDialer{client: client}, nil
}

// DialContext connects to the target through the Tor network.
func (d *TorDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	return d.client.DialContext(ctx, network, address)
}

// FallbackDialer tries Tor first, then falls back to direct connection.
type FallbackDialer struct {
	primary  Dialer
	fallback Dialer
}

// NewFallbackDialer creates a dialer that tries primary first, then fallback.
func NewFallbackDialer(primary, fallback Dialer) *FallbackDialer {
	return &FallbackDialer{primary: primary, fallback: fallback}
}

// DialContext tries the primary dialer first, falling back on error.
func (d *FallbackDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	conn, err := d.primary.DialContext(ctx, network, address)
	if err != nil && d.fallback != nil {
		return d.fallback.DialContext(ctx, network, address)
	}
	return conn, err
}
