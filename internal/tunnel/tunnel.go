// Package tunnel provides dedicated TCP tunnels between ingress nodes and provider nodes.
// Unlike P2P messages, tunnels carry raw TCP traffic with minimal overhead — no JSON encoding,
// no rate limits, no 10MB payload cap. They use the same TLS 1.3 transport for authentication.
package tunnel

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"
)

// Tunnel message types for the control protocol.
const (
	MsgTunnelOpen  byte = 0x01 // Ingress → Provider: open tunnel to deployment
	MsgTunnelReady byte = 0x02 // Provider → Ingress: tunnel is ready
	MsgTunnelError byte = 0x03 // Error message
	MsgTunnelData  byte = 0x04 // Raw TCP data
	MsgTunnelClose byte = 0x05 // Close tunnel
)

// TunnelOpenRequest is sent by ingress to provider to open a tunnel.
type TunnelOpenRequest struct {
	DeploymentID string `json:"deployment_id"`
	Port         int    `json:"port"`
	StreamID     uint32 `json:"stream_id"` // For multiplexing multiple streams
}

// TunnelReadyResponse is sent by provider when tunnel is established.
type TunnelReadyResponse struct {
	StreamID uint32 `json:"stream_id"`
	HostPort int    `json:"host_port"` // Actual port on provider
}

// Tunnel represents a bidirectional TCP tunnel between ingress and provider.
type Tunnel interface {
	// StreamID returns the unique stream identifier.
	StreamID() uint32
	// Read reads data from the tunnel.
	Read(buf []byte) (int, error)
	// Write writes data to the tunnel.
	Write(data []byte) (int, error)
	// Close closes the tunnel.
	Close() error
	// Done returns a channel closed when the tunnel ends.
	Done() <-chan struct{}
}

// tunnel is the concrete implementation.
type tunnel struct {
	streamID uint32
	conn     net.Conn
	done     chan struct{}
}

func newTunnel(streamID uint32, conn net.Conn) *tunnel {
	return &tunnel{
		streamID: streamID,
		conn:     conn,
		done:     make(chan struct{}),
	}
}

func (t *tunnel) StreamID() uint32 { return t.streamID }
func (t *tunnel) Done() <-chan struct{} { return t.done }

func (t *tunnel) Read(buf []byte) (int, error) {
	return t.conn.Read(buf)
}

func (t *tunnel) Write(data []byte) (int, error) {
	return t.conn.Write(data)
}

func (t *tunnel) Close() error {
	select {
	case <-t.done:
		return nil
	default:
		close(t.done)
	}
	return t.conn.Close()
}

// writeControlMsg writes a length-prefixed control message.
func writeControlMsg(w io.Writer, msgType byte, payload []byte) error {
	totalLen := uint32(1 + len(payload))
	if err := binary.Write(w, binary.BigEndian, totalLen); err != nil {
		return fmt.Errorf("write control msg length: %w", err)
	}
	if _, err := w.Write([]byte{msgType}); err != nil {
		return fmt.Errorf("write control msg type: %w", err)
	}
	if len(payload) > 0 {
		if _, err := w.Write(payload); err != nil {
			return fmt.Errorf("write control msg payload: %w", err)
		}
	}
	return nil
}

// readControlMsg reads a length-prefixed control message.
func readControlMsg(r io.Reader) (byte, []byte, error) {
	var totalLen uint32
	if err := binary.Read(r, binary.BigEndian, &totalLen); err != nil {
		return 0, nil, fmt.Errorf("read control msg length: %w", err)
	}
	if totalLen < 1 || totalLen > 1<<20 {
		return 0, nil, fmt.Errorf("invalid control msg length: %d", totalLen)
	}
	buf := make([]byte, totalLen)
	if _, err := io.ReadFull(r, buf); err != nil {
		return 0, nil, fmt.Errorf("read control msg data: %w", err)
	}
	return buf[0], buf[1:], nil
}

// ProxyBidirectional copies data between two connections until one side closes.
// Returns when either direction encounters an error or EOF.
func ProxyBidirectional(ctx context.Context, a, b net.Conn) error {
	errCh := make(chan error, 2)

	go func() {
		_, err := io.Copy(a, b)
		errCh <- err
	}()
	go func() {
		_, err := io.Copy(b, a)
		errCh <- err
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errCh:
		// One direction ended; set deadline to flush the other
		deadline := time.Now().Add(5 * time.Second)
		a.SetDeadline(deadline)
		b.SetDeadline(deadline)
		return err
	}
}
