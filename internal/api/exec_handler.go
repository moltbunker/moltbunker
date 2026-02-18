package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/moltbunker/moltbunker/internal/daemon"
	"github.com/moltbunker/moltbunker/internal/logging"
	"github.com/moltbunker/moltbunker/pkg/types"
)

// execUpgrader uses larger buffers for terminal data.
// Origin validation is handled by checkExecOrigin (see below).
var execUpgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin:     checkExecOrigin,
}

// execAllowedOrigins is the set of origins permitted to open exec WebSockets.
// Loaded once at startup; extend via MOLTBUNKER_EXEC_ALLOWED_ORIGINS env (comma-separated).
var execAllowedOrigins = initExecAllowedOrigins()

func initExecAllowedOrigins() map[string]bool {
	defaults := map[string]bool{
		"https://moltbunker.com":     true,
		"https://app.moltbunker.com": true,
		"https://www.moltbunker.com": true,
	}
	if extra := os.Getenv("MOLTBUNKER_EXEC_ALLOWED_ORIGINS"); extra != "" {
		for _, o := range strings.Split(extra, ",") {
			o = strings.TrimSpace(o)
			if o != "" {
				defaults[o] = true
			}
		}
	}
	return defaults
}

func checkExecOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		// No Origin header = non-browser client (curl, SDK). Allow.
		return true
	}
	if execAllowedOrigins[origin] {
		return true
	}
	// Allow localhost for development
	if strings.HasPrefix(origin, "http://localhost:") || strings.HasPrefix(origin, "http://127.0.0.1:") {
		return true
	}
	logging.Warn("exec WebSocket origin rejected",
		"origin", origin,
		logging.Component("exec_api"))
	return false
}

// ExecChallengeRequest is the request body for POST /v1/exec/challenge
type ExecChallengeRequest struct {
	ContainerID string `json:"container_id"`
}

// ExecChallengeResponse is the response for POST /v1/exec/challenge
type ExecChallengeResponse struct {
	Nonce   string `json:"nonce"`
	Message string `json:"message"`
}

// WSFrame types for the exec WebSocket protocol
const (
	WSFrameData   byte = 0x01 // Terminal data (stdin/stdout)
	WSFrameResize byte = 0x02 // Terminal resize event
	WSFramePing   byte = 0x03 // Keep-alive ping
	WSFramePong   byte = 0x04 // Keep-alive pong
	WSFrameClose   byte = 0x05 // Close session
	WSFrameError   byte = 0x06 // Error message
	WSFrameKeyInit byte = 0x07 // Session key initialization (carries session_nonce)
	WSFrameKeyAck  byte = 0x08 // Session key acknowledgment
)

// WebSocket timing constants
const (
	// wsReadWait is the max time between received messages before the connection
	// is considered dead. Must be longer than the client ping interval (25s).
	wsReadWait = 60 * time.Second

	// wsWriteWait is the max time to complete a single write operation.
	wsWriteWait = 10 * time.Second
)

// safeWSConn serializes writes to a gorilla/websocket.Conn.
// gorilla/websocket does not support concurrent writers — all calls to
// WriteMessage and SetWriteDeadline must be serialized.
type safeWSConn struct {
	conn *websocket.Conn
	wmu  sync.Mutex
}

func newSafeWSConn(conn *websocket.Conn) *safeWSConn {
	return &safeWSConn{conn: conn}
}

// writeMessage sends a binary message with a per-write deadline.
func (c *safeWSConn) writeMessage(data []byte) error {
	c.wmu.Lock()
	defer c.wmu.Unlock()
	c.conn.SetWriteDeadline(time.Now().Add(wsWriteWait))
	return c.conn.WriteMessage(websocket.BinaryMessage, data)
}

// handleExecChallenge creates a single-use nonce for exec authentication.
// The user signs this nonce with their wallet to prove ownership.
// POST /v1/exec/challenge
func (s *Server) handleExecChallenge(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	// Require authenticated session
	walletAddr := s.extractWalletAddress(r)
	if walletAddr == "" {
		http.Error(w, `{"error":"unauthorized - wallet session required"}`, http.StatusUnauthorized)
		return
	}

	var req ExecChallengeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.ContainerID == "" {
		http.Error(w, `{"error":"container_id is required"}`, http.StatusBadRequest)
		return
	}

	// Verify container exists, is running, and the caller owns it.
	// daemonAPI is required — without it we cannot verify ownership.
	if s.daemonAPI == nil {
		http.Error(w, `{"error":"daemon not available"}`, http.StatusServiceUnavailable)
		return
	}
	cm := s.daemonAPI.GetContainerManager()
	if cm == nil {
		http.Error(w, `{"error":"container manager not available"}`, http.StatusServiceUnavailable)
		return
	}
	deployment, exists := cm.GetDeployment(req.ContainerID)
	if !exists {
		http.Error(w, `{"error":"container not found"}`, http.StatusNotFound)
		return
	}
	if deployment.Status != types.ContainerStatusRunning {
		http.Error(w, `{"error":"container not running"}`, http.StatusConflict)
		return
	}
	// Ownership check: only the deployer's wallet can exec
	if deployment.Owner != "" && !strings.EqualFold(deployment.Owner, walletAddr) {
		logging.Warn("exec challenge denied: wallet does not own container",
			"wallet", walletAddr,
			"owner", deployment.Owner,
			"container_id", req.ContainerID,
			logging.Component("exec_api"))
		logging.Audit(logging.AuditEvent{
			Operation: "exec_challenge_denied",
			Actor:     walletAddr,
			Target:    req.ContainerID,
			Result:    "forbidden",
			Details:   "wallet_not_owner",
		})
		http.Error(w, `{"error":"forbidden - you do not own this container"}`, http.StatusForbidden)
		return
	}

	challenge, err := s.execSessions.CreateChallenge(req.ContainerID, walletAddr)
	if err != nil {
		logging.Error("failed to create exec challenge",
			"error", err.Error(),
			"wallet", walletAddr,
			logging.Component("exec_api"))
		http.Error(w, `{"error":"failed to create challenge"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ExecChallengeResponse{
		Nonce:   challenge.Nonce,
		Message: challenge.Message,
	})
}

// handleExecWebSocket upgrades to WebSocket and bridges to the container's PTY.
// GET /v1/exec/ws?nonce=...&signature=...&cols=80&rows=24
func (s *Server) handleExecWebSocket(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	// Extract and validate params
	nonce := r.URL.Query().Get("nonce")
	signature := r.URL.Query().Get("signature")
	if nonce == "" || signature == "" {
		http.Error(w, `{"error":"nonce and signature are required"}`, http.StatusBadRequest)
		return
	}

	// Validate the challenge nonce (single-use, 30s expiry)
	challenge, err := s.execSessions.ValidateChallenge(nonce)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"invalid challenge: %s"}`, err.Error()), http.StatusUnauthorized)
		return
	}

	// Verify wallet signature over the challenge message
	if s.walletAuth == nil {
		http.Error(w, `{"error":"wallet auth not configured"}`, http.StatusInternalServerError)
		return
	}

	recoveredAddr, err := s.walletAuth.VerifySignature(challenge.Message, signature, challenge.Address)
	if err != nil {
		logging.Warn("exec wallet signature verification failed",
			"address", challenge.Address,
			"error", err.Error(),
			logging.Component("exec_api"))
		http.Error(w, `{"error":"signature verification failed"}`, http.StatusUnauthorized)
		return
	}

	// Get container manager for P2P messaging
	if s.daemonAPI == nil {
		http.Error(w, `{"error":"daemon not available"}`, http.StatusServiceUnavailable)
		return
	}
	cm := s.daemonAPI.GetContainerManager()
	if cm == nil {
		http.Error(w, `{"error":"container manager not available"}`, http.StatusServiceUnavailable)
		return
	}

	// Find the provider node for this container
	providerID, ok := cm.GetContainerProviderNode(challenge.ContainerID)
	if !ok {
		http.Error(w, `{"error":"container provider not found"}`, http.StatusNotFound)
		return
	}

	// Generate session ID
	sessionID, err := GenerateSessionID()
	if err != nil {
		http.Error(w, `{"error":"failed to generate session"}`, http.StatusInternalServerError)
		return
	}

	// Parse terminal dimensions
	cols := parseIntParam(r.URL.Query().Get("cols"), 80)
	rows := parseIntParam(r.URL.Query().Get("rows"), 24)

	// Register session
	now := time.Now()
	session := &ExecSession{
		SessionID:     sessionID,
		ContainerID:   challenge.ContainerID,
		WalletAddress: strings.ToLower(recoveredAddr),
		CreatedAt:     now,
		LastActivity:  now,
	}
	if err := s.execSessions.AddSession(session); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusTooManyRequests)
		return
	}

	// Upgrade to WebSocket
	conn, err := execUpgrader.Upgrade(w, r, nil)
	if err != nil {
		s.execSessions.RemoveSession(sessionID)
		logging.Warn("exec WebSocket upgrade failed",
			"error", err.Error(),
			logging.Component("exec_api"))
		return
	}

	// Clear server-level Read/WriteTimeout deadlines inherited from http.Server.
	// WebSocket is long-lived — the exec handler manages its own per-message deadlines.
	conn.SetReadDeadline(time.Time{})
	conn.SetWriteDeadline(time.Time{})

	logging.Info("exec WebSocket session started",
		"session_id", sessionID,
		"container_id", challenge.ContainerID,
		"wallet", recoveredAddr,
		"provider", providerID.String()[:16],
		"cols", cols,
		"rows", rows,
		logging.Component("exec_api"))

	// Audit log
	logging.Audit(logging.AuditEvent{
		Operation: "exec_session_start",
		Actor:     recoveredAddr,
		Target:    challenge.ContainerID,
		Result:    "success",
		Details:   fmt.Sprintf("session=%s provider=%s", sessionID, providerID.String()[:16]),
	})

	// Bridge WebSocket ↔ P2P
	s.bridgeExecSession(conn, cm, session, providerID, cols, rows)
}

// bridgeExecSession bridges a WebSocket connection to a container exec session.
// If the container runs locally (same node), it directly opens a PTY.
// If remote, it uses P2P message relay to the provider node.
func (s *Server) bridgeExecSession(
	conn *websocket.Conn,
	cm *daemon.ContainerManager,
	session *ExecSession,
	providerID types.NodeID,
	cols, rows int,
) {
	// If the container runs on the local node, bridge directly to PTY
	if providerID == cm.LocalNodeID() {
		s.bridgeLocalExec(conn, cm, session, cols, rows)
		return
	}
	s.bridgeRemoteExec(conn, cm, session, providerID, cols, rows)
}

// bridgeLocalExec bridges a WebSocket directly to a local container PTY.
// No P2P messaging needed — the API server IS the provider.
func (s *Server) bridgeLocalExec(
	conn *websocket.Conn,
	cm *daemon.ContainerManager,
	session *ExecSession,
	cols, rows int,
) {
	ws := newSafeWSConn(conn)
	defer func() {
		ws.conn.Close()
		s.execSessions.RemoveSession(session.SessionID)

		logging.Info("exec local session ended",
			"session_id", session.SessionID,
			"container_id", session.ContainerID,
			"bytes_sent", session.BytesSent,
			"bytes_received", session.BytesReceived,
			logging.Component("exec_api"))

		logging.Audit(logging.AuditEvent{
			Operation: "exec_session_end",
			Actor:     session.WalletAddress,
			Target:    session.ContainerID,
			Result:    "success",
			Details:   fmt.Sprintf("session=%s sent=%d recv=%d mode=local", session.SessionID, session.BytesSent, session.BytesReceived),
		})
	}()

	// Open PTY directly on the local container runtime (with ownership verification)
	ctx := context.Background()
	ptySession, err := cm.ExecLocal(ctx, session.ContainerID, session.WalletAddress, uint16(cols), uint16(rows))
	if err != nil {
		logging.Error("local exec failed",
			"session_id", session.SessionID,
			"container_id", session.ContainerID,
			"error", err.Error(),
			logging.Component("exec_api"))
		sendWSError(conn, fmt.Sprintf("exec failed: %v", err))
		return
	}
	defer ptySession.Close()

	logging.Info("exec local session started",
		"session_id", session.SessionID,
		"container_id", session.ContainerID,
		"cols", cols,
		"rows", rows,
		logging.Component("exec_api"))

	// Goroutine: PTY stdout → WebSocket (all writes go through ws to serialize)
	done := make(chan struct{})
	go func() {
		defer close(done)
		buf := make([]byte, 4096)
		for {
			n, readErr := ptySession.Stdout.Read(buf)
			if n > 0 {
				frame := make([]byte, 1+n)
				frame[0] = WSFrameData
				copy(frame[1:], buf[:n])
				if writeErr := ws.writeMessage(frame); writeErr != nil {
					return
				}
				s.execSessions.AddBytes(session.SessionID, int64(n), 0)
			}
			if readErr != nil {
				// PTY closed — send close frame
				closeFrame := []byte{WSFrameClose}
				closeFrame = append(closeFrame, []byte("session_ended")...)
				ws.writeMessage(closeFrame)
				return
			}
		}
	}()

	// Main loop: WebSocket → PTY stdin
	conn.SetReadLimit(64 * 1024)
	conn.SetReadDeadline(time.Now().Add(wsReadWait))
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				logging.Debug("exec WebSocket read error",
					"session_id", session.SessionID,
					"error", err.Error(),
					logging.Component("exec_api"))
			}
			break
		}
		// Extend read deadline on every received message (client pings every 25s)
		conn.SetReadDeadline(time.Now().Add(wsReadWait))

		if len(message) == 0 {
			continue
		}

		frameType := message[0]
		frameData := message[1:]

		switch frameType {
		case WSFrameData:
			ptySession.Stdin.Write(frameData)
			s.execSessions.AddBytes(session.SessionID, 0, int64(len(frameData)))
		case WSFrameResize:
			if len(frameData) >= 4 {
				newCols := uint16(frameData[0])<<8 | uint16(frameData[1])
				newRows := uint16(frameData[2])<<8 | uint16(frameData[3])
				ptySession.Resize(newCols, newRows)
			}
		case WSFrameKeyInit, WSFrameKeyAck:
			// E2E key exchange frames — relay to container stdin as exec-agent frames
			ptySession.Stdin.Write(frameData)
		case WSFramePing:
			ws.writeMessage([]byte{WSFramePong})
		case WSFrameClose:
			goto cleanup
		}
	}

cleanup:
	ptySession.Close()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
	}
}

// bridgeRemoteExec bridges a WebSocket to a remote container via P2P relay.
func (s *Server) bridgeRemoteExec(
	conn *websocket.Conn,
	cm *daemon.ContainerManager,
	session *ExecSession,
	providerID types.NodeID,
	cols, rows int,
) {
	ws := newSafeWSConn(conn)
	defer func() {
		ws.conn.Close()
		cm.RemoveExecRelay(session.SessionID)
		s.execSessions.RemoveSession(session.SessionID)

		logging.Info("exec remote session ended",
			"session_id", session.SessionID,
			"container_id", session.ContainerID,
			"bytes_sent", session.BytesSent,
			"bytes_received", session.BytesReceived,
			logging.Component("exec_api"))

		logging.Audit(logging.AuditEvent{
			Operation: "exec_session_end",
			Actor:     session.WalletAddress,
			Target:    session.ContainerID,
			Result:    "success",
			Details:   fmt.Sprintf("session=%s sent=%d recv=%d mode=remote", session.SessionID, session.BytesSent, session.BytesReceived),
		})
	}()

	// Channel for provider → WebSocket data
	dataCh := make(chan []byte, 64)
	closeCh := make(chan string, 1)

	// Register relay: P2P ExecData from provider → dataCh → WebSocket
	relay := &daemon.ExecRelay{
		SessionID:   session.SessionID,
		ContainerID: session.ContainerID,
		ProviderID:  providerID,
		OnData: func(data []byte) {
			select {
			case dataCh <- data:
			default:
				// Buffer full — drop data to prevent blocking P2P handler
			}
		},
		OnClose: func(reason string) {
			select {
			case closeCh <- reason:
			default:
			}
		},
	}
	cm.RegisterExecRelay(relay)

	// Send ExecOpen to provider via P2P
	openPayload := types.ExecOpenPayload{
		ContainerID:   session.ContainerID,
		SessionID:     session.SessionID,
		WalletAddress: session.WalletAddress,
		Cols:          cols,
		Rows:          rows,
	}
	payloadBytes, _ := json.Marshal(openPayload)

	ctx := context.Background()
	err := cm.SendExecMessage(ctx, providerID, &types.Message{
		Type:      types.MessageTypeExecOpen,
		From:      cm.LocalNodeID(),
		To:        providerID,
		Payload:   payloadBytes,
		Timestamp: time.Now(),
	})
	if err != nil {
		logging.Error("failed to send exec open",
			"session_id", session.SessionID,
			"error", err.Error(),
			logging.Component("exec_api"))
		sendWSError(conn, "failed to connect to container")
		return
	}

	// Start bidirectional relay with cancellation for cleanup
	relayCtx, relayCancel := context.WithCancel(context.Background())
	defer relayCancel()
	done := make(chan struct{})

	// Goroutine: Provider data → WebSocket (all writes go through ws to serialize)
	go func() {
		defer close(done)
		for {
			select {
			case data := <-dataCh:
				// Prepend frame type byte
				frame := make([]byte, 1+len(data))
				frame[0] = WSFrameData
				copy(frame[1:], data)

				if err := ws.writeMessage(frame); err != nil {
					return
				}
				s.execSessions.AddBytes(session.SessionID, int64(len(data)), 0)

			case reason := <-closeCh:
				// Provider closed the session
				closeFrame := []byte{WSFrameClose}
				closeFrame = append(closeFrame, []byte(reason)...)
				ws.writeMessage(closeFrame)
				return

			case <-relayCtx.Done():
				return
			}
		}
	}()

	// Main loop: WebSocket → Provider (reads from browser)
	conn.SetReadLimit(64 * 1024) // 64KB max message
	conn.SetReadDeadline(time.Now().Add(wsReadWait))
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				logging.Debug("exec WebSocket read error",
					"session_id", session.SessionID,
					"error", err.Error(),
					logging.Component("exec_api"))
			}
			break
		}
		// Extend read deadline on every received message (client pings every 25s)
		conn.SetReadDeadline(time.Now().Add(wsReadWait))

		if len(message) == 0 {
			continue
		}

		frameType := message[0]
		frameData := message[1:]

		switch frameType {
		case WSFrameData:
			// Forward keyboard input to provider
			dataPayload := types.ExecDataPayload{
				SessionID: session.SessionID,
				Data:      frameData,
			}
			pb, _ := json.Marshal(dataPayload)
			cm.SendExecMessage(ctx, providerID, &types.Message{
				Type:      types.MessageTypeExecData,
				From:      cm.LocalNodeID(),
				To:        providerID,
				Payload:   pb,
				Timestamp: time.Now(),
			})
			s.execSessions.AddBytes(session.SessionID, 0, int64(len(frameData)))

		case WSFrameResize:
			if len(frameData) >= 4 {
				newCols := int(frameData[0])<<8 | int(frameData[1])
				newRows := int(frameData[2])<<8 | int(frameData[3])
				resizePayload := types.ExecResizePayload{
					SessionID: session.SessionID,
					Cols:      newCols,
					Rows:      newRows,
				}
				pb, _ := json.Marshal(resizePayload)
				cm.SendExecMessage(ctx, providerID, &types.Message{
					Type:      types.MessageTypeExecResize,
					From:      cm.LocalNodeID(),
					To:        providerID,
					Payload:   pb,
					Timestamp: time.Now(),
				})
			}

		case WSFrameKeyInit, WSFrameKeyAck:
			// E2E key exchange frames — forward as ExecData to provider (opaque relay).
			// The provider's ExecStream wraps it in exec-agent framing automatically.
			dataPayload := types.ExecDataPayload{
				SessionID: session.SessionID,
				Data:      message, // include frame type byte — provider relays as-is
			}
			pb, _ := json.Marshal(dataPayload)
			cm.SendExecMessage(ctx, providerID, &types.Message{
				Type:      types.MessageTypeExecData,
				From:      cm.LocalNodeID(),
				To:        providerID,
				Payload:   pb,
				Timestamp: time.Now(),
			})

		case WSFramePing:
			ws.writeMessage([]byte{WSFramePong})

		case WSFrameClose:
			// User requested close
			goto cleanup
		}
	}

cleanup:
	// Send ExecClose to provider
	closePayload := types.ExecClosePayload{
		SessionID: session.SessionID,
		Reason:    "client_disconnect",
	}
	pb, _ := json.Marshal(closePayload)
	cm.SendExecMessage(ctx, providerID, &types.Message{
		Type:      types.MessageTypeExecClose,
		From:      cm.LocalNodeID(),
		To:        providerID,
		Payload:   pb,
		Timestamp: time.Now(),
	})

	// Wait for the writer goroutine to finish
	select {
	case <-done:
	case <-time.After(5 * time.Second):
	}
}

// extractWalletAddress extracts the wallet address from the authenticated request
func (s *Server) extractWalletAddress(r *http.Request) string {
	// Check Bearer token for wallet session
	auth := r.Header.Get("Authorization")
	if auth != "" && len(auth) > 7 && auth[:7] == "Bearer " {
		token := auth[7:]
		if len(token) > 3 && token[:3] == "wt_" {
			if s.walletAuth != nil {
				addr, valid := s.walletAuth.ValidateSession(token)
				if valid {
					return addr
				}
			}
		}
	}

	// Check inline wallet headers
	walletAddr := r.Header.Get("X-Wallet-Address")
	walletSig := r.Header.Get("X-Wallet-Signature")
	walletMsg := r.Header.Get("X-Wallet-Message")
	if walletAddr != "" && walletSig != "" && walletMsg != "" {
		if s.walletAuth != nil {
			addr, err := s.walletAuth.VerifyInlineAuth(walletAddr, walletSig, walletMsg)
			if err == nil {
				return addr
			}
		}
	}

	return ""
}

// sendWSError sends an error frame over WebSocket
func sendWSError(conn *websocket.Conn, msg string) {
	frame := []byte{WSFrameError}
	frame = append(frame, []byte(msg)...)
	conn.WriteMessage(websocket.BinaryMessage, frame)
}

// parseIntParam parses an integer query parameter with a default value
func parseIntParam(s string, defaultVal int) int {
	if s == "" {
		return defaultVal
	}
	var v int
	if _, err := fmt.Sscanf(s, "%d", &v); err != nil || v <= 0 {
		return defaultVal
	}
	return v
}
