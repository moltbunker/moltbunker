package tunnel

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"sync"

	"github.com/moltbunker/moltbunker/internal/logging"
)

// PortResolver resolves a deployment+port to an actual local host:port.
type PortResolver interface {
	// ResolveDeploymentPort returns the local address (host:port) for a deployment's exposed port.
	ResolveDeploymentPort(deploymentID string, containerPort int) (localAddr string, err error)
}

// Server listens for tunnel connections from ingress nodes and proxies traffic
// to local containers. It runs on provider nodes.
type Server struct {
	listener     net.Listener
	resolver     PortResolver
	activeTunnel map[uint32]*tunnel
	mu           sync.RWMutex
	cancel       context.CancelFunc
}

// NewServer creates a tunnel server on the given listener.
func NewServer(listener net.Listener, resolver PortResolver) *Server {
	return &Server{
		listener:     listener,
		resolver:     resolver,
		activeTunnel: make(map[uint32]*tunnel),
	}
}

// Serve accepts and handles tunnel connections. Blocks until ctx is cancelled.
func (s *Server) Serve(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	s.cancel = cancel
	defer cancel()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				logging.Error("tunnel accept error", logging.Err(err), logging.Component("tunnel"))
				continue
			}
		}

		go s.handleConnection(ctx, conn)
	}
}

// handleConnection handles a single ingress connection.
func (s *Server) handleConnection(ctx context.Context, conn net.Conn) {
	defer conn.Close()

	// Read TUNNEL_OPEN request
	msgType, payload, err := readControlMsg(conn)
	if err != nil {
		logging.Error("tunnel read open", logging.Err(err), logging.Component("tunnel"))
		return
	}
	if msgType != MsgTunnelOpen {
		writeControlMsg(conn, MsgTunnelError, []byte("expected TUNNEL_OPEN"))
		return
	}

	var req TunnelOpenRequest
	if err := json.Unmarshal(payload, &req); err != nil {
		writeControlMsg(conn, MsgTunnelError, []byte("invalid TUNNEL_OPEN payload"))
		return
	}

	logging.Info("tunnel open request",
		"deployment_id", req.DeploymentID,
		"port", req.Port,
		"stream_id", req.StreamID,
		logging.Component("tunnel"))

	// Resolve the deployment port to a local address
	localAddr, err := s.resolver.ResolveDeploymentPort(req.DeploymentID, req.Port)
	if err != nil {
		errMsg := fmt.Sprintf("resolve port: %v", err)
		writeControlMsg(conn, MsgTunnelError, []byte(errMsg))
		return
	}

	// Connect to the local container
	containerConn, err := net.Dial("tcp", localAddr)
	if err != nil {
		errMsg := fmt.Sprintf("connect to container: %v", err)
		writeControlMsg(conn, MsgTunnelError, []byte(errMsg))
		return
	}
	defer containerConn.Close()

	// Send TUNNEL_READY
	readyPayload, _ := json.Marshal(TunnelReadyResponse{
		StreamID: req.StreamID,
		HostPort: req.Port,
	})
	if err := writeControlMsg(conn, MsgTunnelReady, readyPayload); err != nil {
		return
	}

	// Track tunnel
	tun := newTunnel(req.StreamID, conn)
	s.mu.Lock()
	s.activeTunnel[req.StreamID] = tun
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		delete(s.activeTunnel, req.StreamID)
		s.mu.Unlock()
		tun.Close()
	}()

	// Proxy bidirectionally
	_ = ProxyBidirectional(ctx, conn, containerConn)
}

// Close stops the tunnel server.
func (s *Server) Close() error {
	if s.cancel != nil {
		s.cancel()
	}
	return s.listener.Close()
}

// ActiveCount returns the number of active tunnels.
func (s *Server) ActiveCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.activeTunnel)
}
