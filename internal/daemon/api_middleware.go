package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/moltbunker/moltbunker/internal/util"
)

// handleConnections accepts and handles incoming connections
func (s *APIServer) handleConnections(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			conn, err := s.listener.Accept()
			if err != nil {
				s.mu.RLock()
				running := s.running
				s.mu.RUnlock()
				if !running {
					return
				}
				continue
			}

			util.SafeGoWithName("api-handle-connection", func() {
				s.handleConnection(ctx, conn)
			})
		}
	}
}

// handleConnection handles a single API connection with rate limiting
func (s *APIServer) handleConnection(ctx context.Context, conn net.Conn) {
	defer conn.Close()

	// Track active connections
	s.metrics.IncrementConnections()
	defer s.metrics.DecrementConnections()

	// Create a rate limiter for this connection (configurable requests per window)
	limiter := newRateLimiter(s.rateLimitRequests, s.rateLimitWindow)

	// Create a limited reader to enforce max request size
	limitedReader := io.LimitReader(conn, s.maxRequestSize)
	decoder := json.NewDecoder(limitedReader)
	encoder := json.NewEncoder(conn)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			// Check global rate limit first (shared across all connections)
			if !s.globalLimiter.allow() {
				s.sendError(encoder, 0, ErrRateLimited.Error())
				// Close connection when rate limited to prevent infinite loop
				return
			}

			// Check per-connection rate limit as additional protection
			if !limiter.allow() {
				s.sendError(encoder, 0, ErrRateLimited.Error())
				// Close connection when rate limited to prevent infinite loop
				return
			}

			var req APIRequest
			if err := decoder.Decode(&req); err != nil {
				if err == io.EOF {
					return
				}
				// Check if it's a size limit error
				if err.Error() == "unexpected EOF" {
					s.sendError(encoder, 0, ErrRequestTooLarge.Error())
					return
				}
				s.sendError(encoder, 0, "invalid request")
				return
			}

			response := s.handleRequest(ctx, &req)
			if err := encoder.Encode(response); err != nil {
				return
			}

			// Reset the limited reader for the next request
			// We need to recreate it since LimitReader doesn't reset
			limitedReader = io.LimitReader(conn, s.maxRequestSize)
			decoder = json.NewDecoder(limitedReader)
		}
	}
}

// handleRequest routes and handles an API request
func (s *APIServer) handleRequest(ctx context.Context, req *APIRequest) *APIResponse {
	// Record request metrics
	start := time.Now()
	s.metrics.RecordRequest(req.Method)

	// Update gauges
	if s.containerManager != nil {
		s.metrics.SetContainerCount(len(s.containerManager.ListDeployments()))
	}
	if s.node != nil && s.node.router != nil {
		s.metrics.SetPeerCount(len(s.node.router.GetPeers()))
	}

	var response *APIResponse

	switch req.Method {
	case "status":
		response = s.handleStatus(ctx, req)
	case "deploy":
		response = s.handleDeploy(ctx, req)
	case "stop":
		response = s.handleStop(ctx, req)
	case "delete":
		response = s.handleDelete(ctx, req)
	case "logs":
		response = s.handleLogs(ctx, req)
	case "list":
		response = s.handleList(ctx, req)
	case "tor_start":
		response = s.handleTorStart(ctx, req)
	case "tor_stop":
		response = s.handleTorStop(ctx, req)
	case "tor_status":
		response = s.handleTorStatus(ctx, req)
	case "tor_rotate":
		response = s.handleTorRotate(ctx, req)
	case "peers":
		response = s.handlePeers(ctx, req)
	case "health":
		response = s.handleHealth(ctx, req)
	case "healthz":
		response = s.handleHealthz(ctx, req)
	case "readyz":
		response = s.handleReadyz(ctx, req)
	case "metrics":
		response = s.handleMetrics(ctx, req)
	case "config_get":
		response = s.handleConfigGet(ctx, req)
	case "config_set":
		response = s.handleConfigSet(ctx, req)
	default:
		response = &APIResponse{
			Error: fmt.Sprintf("unknown method: %s", req.Method),
			ID:    req.ID,
		}
	}

	// Record latency
	s.metrics.RecordLatency(req.Method, time.Since(start))

	return response
}
