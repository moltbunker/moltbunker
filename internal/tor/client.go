package tor

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/moltbunker/moltbunker/internal/util"
	"golang.org/x/net/proxy"
)

// TorClient provides SOCKS5 proxy client for Tor
type TorClient struct {
	socksAddr string
	dialer    proxy.Dialer
}

// NewTorClient creates a new Tor SOCKS5 client
func NewTorClient(socksAddr string) (*TorClient, error) {
	// Create SOCKS5 dialer
	dialer, err := proxy.SOCKS5("tcp", socksAddr, nil, proxy.Direct)
	if err != nil {
		return nil, fmt.Errorf("failed to create SOCKS5 dialer: %w", err)
	}

	return &TorClient{
		socksAddr: socksAddr,
		dialer:    dialer,
	}, nil
}

// Dial connects through Tor SOCKS5 proxy
func (tc *TorClient) Dial(network, address string) (net.Conn, error) {
	return tc.dialer.Dial(network, address)
}

// DialContext connects through Tor SOCKS5 proxy with context
func (tc *TorClient) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	// Create a channel for the connection
	type result struct {
		conn net.Conn
		err  error
	}
	resultChan := make(chan result, 1)

	util.SafeGoWithName("tor-dial", func() {
		conn, err := tc.Dial(network, address)
		resultChan <- result{conn: conn, err: err}
	})

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case res := <-resultChan:
		return res.conn, res.err
	}
}

// DialTimeout connects through Tor with timeout
func (tc *TorClient) DialTimeout(network, address string, timeout time.Duration) (net.Conn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return tc.DialContext(ctx, network, address)
}

// SOCKSAddress returns the SOCKS5 proxy address
func (tc *TorClient) SOCKSAddress() string {
	return tc.socksAddr
}
