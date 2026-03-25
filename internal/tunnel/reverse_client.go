package tunnel

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/hashicorp/yamux"
	"github.com/moltbunker/moltbunker/internal/logging"
	"github.com/moltbunker/moltbunker/pkg/types"
)

const (
	// maxReconnectAttempts before giving up and re-registering fresh.
	maxReconnectAttempts = 10

	// maxReconnectBackoff is the maximum backoff between retries.
	maxReconnectBackoff = 30 * time.Second
)

// ReverseClient connects to an ingress node and maintains a persistent
// reverse tunnel, exposing local container ports via assigned subdomains.
type ReverseClient struct {
	ingressAddr  string
	tlsConfig    *tls.Config
	portResolver PortResolver
	nodeID       types.NodeID

	// State
	mu          sync.Mutex
	session     *yamux.Session
	ctrlStream  net.Conn
	subdomain   string
	reconnToken string
	cancel      context.CancelFunc
	wg          sync.WaitGroup
}

// NewReverseClient creates a new reverse tunnel client.
func NewReverseClient(ingressAddr string, portResolver PortResolver, tlsCfg *tls.Config) *ReverseClient {
	return &ReverseClient{
		ingressAddr:  ingressAddr,
		tlsConfig:    tlsCfg,
		portResolver: portResolver,
	}
}

// Connect establishes a reverse tunnel to the ingress and exposes a deployment's container port.
// Returns the assigned subdomain. Blocks while the tunnel is active.
func (c *ReverseClient) Connect(ctx context.Context, deploymentID string, containerPort int) (string, error) {
	ctx, cancel := context.WithCancel(ctx)
	c.mu.Lock()
	c.cancel = cancel
	c.mu.Unlock()
	defer cancel()

	var lastErr error
	attempts := 0

	for {
		select {
		case <-ctx.Done():
			return c.subdomain, ctx.Err()
		default:
		}

		subdomain, err := c.connectOnce(ctx, deploymentID, containerPort)
		if err != nil {
			lastErr = err
			attempts++
			logging.Warn("reverse tunnel connection failed",
				"attempt", attempts,
				"error", err.Error(),
				"ingress", c.ingressAddr,
				logging.Component("reverse-tunnel"))

			if attempts > maxReconnectAttempts {
				// Clear reconnection token to force fresh registration
				c.mu.Lock()
				c.reconnToken = ""
				c.mu.Unlock()
				attempts = 0
			}

			// Exponential backoff: 1s, 2s, 4s, 8s... max 30s
			backoff := time.Duration(1<<uint(min(attempts, 5))) * time.Second
			if backoff > maxReconnectBackoff {
				backoff = maxReconnectBackoff
			}

			select {
			case <-ctx.Done():
				return c.subdomain, ctx.Err()
			case <-time.After(backoff):
				continue
			}
		}

		return subdomain, lastErr
	}
}

// connectOnce establishes a single connection and runs the stream accept loop.
func (c *ReverseClient) connectOnce(ctx context.Context, deploymentID string, containerPort int) (string, error) {
	// TLS dial to ingress
	dialer := &net.Dialer{Timeout: 15 * time.Second}
	conn, err := tls.DialWithDialer(dialer, "tcp", c.ingressAddr, c.tlsConfig)
	if err != nil {
		return "", fmt.Errorf("dial ingress %s: %w", c.ingressAddr, err)
	}
	defer conn.Close()

	// Derive our NodeID from our TLS certificate
	tlsState := conn.ConnectionState()
	if len(tlsState.PeerCertificates) > 0 {
		// Pin ingress SPKI on first connect (TOFU)
		// The TLS config already handles this
	}

	// Our NodeID from our local cert
	if c.tlsConfig.Certificates != nil && len(c.tlsConfig.Certificates) > 0 {
		localCert := c.tlsConfig.Certificates[0]
		if localCert.Leaf != nil {
			spki := localCert.Leaf.RawSubjectPublicKeyInfo
			nodeIDBytes := sha256.Sum256(spki)
			copy(c.nodeID[:], nodeIDBytes[:])
		}
	}

	// Wrap with yamux — provider is yamux server
	yamuxCfg := yamux.DefaultConfig()
	yamuxCfg.MaxStreamWindowSize = 256 * 1024
	yamuxCfg.StreamOpenTimeout = 10 * time.Second
	yamuxCfg.EnableKeepAlive = true
	yamuxCfg.KeepAliveInterval = 30 * time.Second
	yamuxCfg.ConnectionWriteTimeout = 10 * time.Second
	yamuxCfg.LogOutput = io.Discard

	session, err := yamux.Server(conn, yamuxCfg)
	if err != nil {
		return "", fmt.Errorf("yamux server setup: %w", err)
	}
	defer session.Close()

	c.mu.Lock()
	c.session = session
	c.mu.Unlock()

	// Accept control stream from ingress (the ingress opens stream 0)
	ctrlStream, err := session.Accept()
	if err != nil {
		return "", fmt.Errorf("accept control stream: %w", err)
	}
	defer ctrlStream.Close()

	c.mu.Lock()
	c.ctrlStream = ctrlStream
	c.mu.Unlock()

	// Read the registration request prompt — actually, we need to read the
	// registration response since the ingress reads our request first.
	// Wait, the flow is: ingress opens control stream, reads our registration.
	// So we write our registration to the control stream that the ingress opened.

	// Generate registration request
	var nonce [32]byte
	rand.Read(nonce[:])

	c.mu.Lock()
	reconnToken := c.reconnToken
	wantSubdomain := c.subdomain
	c.mu.Unlock()

	req := TunnelRegisterRequest{
		NodeID:        c.nodeID.String(),
		Subdomain:     wantSubdomain,
		DeploymentID:  deploymentID,
		ContainerPort: containerPort,
		Nonce:         hex.EncodeToString(nonce[:]),
		Timestamp:     time.Now().Unix(),
		TLSBinding:    hex.EncodeToString(tlsState.TLSUnique),
		ReconnToken:   reconnToken,
	}

	reqPayload, _ := json.Marshal(req)
	if err := writeControlMsg(ctrlStream, MsgTunnelRegister, reqPayload); err != nil {
		return "", fmt.Errorf("send registration: %w", err)
	}

	// Read response
	msgType, respPayload, err := readControlMsg(ctrlStream)
	if err != nil {
		return "", fmt.Errorf("read registration response: %w", err)
	}

	switch msgType {
	case MsgTunnelRegistered:
		var resp TunnelRegisterResponse
		if err := json.Unmarshal(respPayload, &resp); err != nil {
			return "", fmt.Errorf("parse registration response: %w", err)
		}

		c.mu.Lock()
		c.subdomain = resp.Subdomain
		c.reconnToken = resp.ReconnToken
		c.mu.Unlock()

		logging.Info("reverse tunnel established",
			"subdomain", resp.Subdomain,
			"full_domain", resp.FullDomain,
			"max_rps", resp.Limits.MaxRPS,
			logging.Component("reverse-tunnel"))

		// Run the stream accept loop + heartbeat responder
		return resp.Subdomain, c.serveStreams(ctx, session, ctrlStream, containerPort)

	case MsgTunnelError:
		return "", fmt.Errorf("registration rejected: %s", string(respPayload))

	default:
		return "", fmt.Errorf("unexpected response type: 0x%02x", msgType)
	}
}

// serveStreams accepts yamux streams from ingress and proxies to local containers.
// Also handles heartbeat pong responses on the control stream.
func (c *ReverseClient) serveStreams(ctx context.Context, session *yamux.Session,
	ctrl net.Conn, containerPort int) error {

	// Heartbeat responder goroutine
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		c.heartbeatResponder(ctx, ctrl)
	}()

	// Accept streams from ingress (HTTP request proxying)
	for {
		stream, err := session.Accept()
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return fmt.Errorf("accept stream: %w", err)
		}

		c.wg.Add(1)
		go func() {
			defer c.wg.Done()
			defer stream.Close()
			c.handleStream(ctx, stream, containerPort)
		}()
	}
}

// handleStream proxies a single HTTP request from ingress to the local container.
func (c *ReverseClient) handleStream(ctx context.Context, stream net.Conn, containerPort int) {
	// Connect to the local container
	localAddr := fmt.Sprintf("127.0.0.1:%d", containerPort)
	containerConn, err := net.DialTimeout("tcp", localAddr, 5*time.Second)
	if err != nil {
		logging.Debug("reverse tunnel: failed to connect to local container",
			"addr", localAddr,
			logging.Err(err),
			logging.Component("reverse-tunnel"))
		return
	}
	defer containerConn.Close()

	// Proxy bidirectionally
	ProxyBidirectional(ctx, stream, containerConn)
}

// heartbeatResponder reads pings from the control stream and responds with pongs.
func (c *ReverseClient) heartbeatResponder(ctx context.Context, ctrl net.Conn) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		// Read next message (blocking)
		ctrl.SetReadDeadline(time.Now().Add(2 * heartbeatInterval))
		msgType, payload, err := readControlMsg(ctrl)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue // Read timeout is normal between pings
			}
			logging.Debug("reverse tunnel: control stream read error",
				logging.Err(err),
				logging.Component("reverse-tunnel"))
			return
		}

		switch msgType {
		case MsgTunnelPing:
			var ping TunnelPing
			if err := json.Unmarshal(payload, &ping); err != nil {
				continue
			}
			// Echo challenge back
			pong := TunnelPong{Challenge: ping.Challenge}
			pongPayload, _ := json.Marshal(pong)
			ctrl.SetWriteDeadline(time.Now().Add(heartbeatTimeout))
			if err := writeControlMsg(ctrl, MsgTunnelPong, pongPayload); err != nil {
				return
			}

		case MsgTunnelDeregister:
			logging.Info("reverse tunnel: ingress requested deregistration",
				logging.Component("reverse-tunnel"))
			return

		default:
			// Unknown message type on control stream — ignore
			logging.Debug("reverse tunnel: unknown control message",
				"type", fmt.Sprintf("0x%02x", msgType),
				logging.Component("reverse-tunnel"))
		}
	}
}

// Disconnect gracefully closes the reverse tunnel.
func (c *ReverseClient) Disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cancel != nil {
		c.cancel()
	}

	// Send deregister on control stream if available
	if c.ctrlStream != nil {
		writeControlMsg(c.ctrlStream, MsgTunnelDeregister, nil)
		c.ctrlStream.Close()
		c.ctrlStream = nil
	}

	if c.session != nil {
		c.session.Close()
		c.session = nil
	}

	c.wg.Wait()
	return nil
}

// Subdomain returns the currently assigned subdomain, or empty if not connected.
func (c *ReverseClient) Subdomain() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.subdomain
}
