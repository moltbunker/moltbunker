package tunnel

import (
	"encoding/json"
	"fmt"
	"net"
	"sync/atomic"

	"github.com/moltbunker/moltbunker/internal/logging"
)

// streamCounter generates unique stream IDs.
var streamCounter atomic.Uint32

// Dialer opens tunnel connections to provider nodes.
// In production, it uses the TLS 1.3 transport (p2p.Transport.DialContext)
// for authenticated connections.
type Dialer interface {
	// DialProvider connects to a provider's tunnel port.
	DialProvider(providerAddr string) (net.Conn, error)
}

// Client opens tunnels to providers on behalf of ingress nodes.
type Client struct {
	dialer Dialer
}

// NewClient creates a tunnel client with the given dialer.
func NewClient(dialer Dialer) *Client {
	return &Client{dialer: dialer}
}

// OpenTunnel connects to a provider and opens a tunnel to the specified deployment port.
// Returns a Tunnel that can be used for bidirectional TCP proxying.
func (c *Client) OpenTunnel(providerAddr string, deploymentID string, port int) (Tunnel, error) {
	conn, err := c.dialer.DialProvider(providerAddr)
	if err != nil {
		return nil, fmt.Errorf("dial provider %s: %w", providerAddr, err)
	}

	streamID := streamCounter.Add(1)

	// Send TUNNEL_OPEN
	req := TunnelOpenRequest{
		DeploymentID: deploymentID,
		Port:         port,
		StreamID:     streamID,
	}
	payload, _ := json.Marshal(req)
	if err := writeControlMsg(conn, MsgTunnelOpen, payload); err != nil {
		conn.Close()
		return nil, fmt.Errorf("send TUNNEL_OPEN: %w", err)
	}

	// Wait for TUNNEL_READY
	msgType, respPayload, err := readControlMsg(conn)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("read TUNNEL_READY: %w", err)
	}

	switch msgType {
	case MsgTunnelReady:
		var resp TunnelReadyResponse
		if err := json.Unmarshal(respPayload, &resp); err != nil {
			conn.Close()
			return nil, fmt.Errorf("parse TUNNEL_READY: %w", err)
		}
		logging.Info("tunnel opened",
			"deployment_id", deploymentID,
			"port", port,
			"stream_id", resp.StreamID,
			logging.Component("tunnel"))

		return newTunnel(resp.StreamID, conn), nil

	case MsgTunnelError:
		conn.Close()
		return nil, fmt.Errorf("provider rejected tunnel: %s", string(respPayload))

	default:
		conn.Close()
		return nil, fmt.Errorf("unexpected response type: 0x%02x", msgType)
	}
}
