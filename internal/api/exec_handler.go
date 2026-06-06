package api

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
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
	WSFrameData    byte = 0x01 // Terminal data (stdin/stdout)
	WSFrameResize  byte = 0x02 // Terminal resize event
	WSFramePing    byte = 0x03 // Keep-alive ping
	WSFramePong    byte = 0x04 // Keep-alive pong
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
	if err := c.conn.SetWriteDeadline(time.Now().Add(wsWriteWait)); err != nil {
		logging.Warn("failed to set websocket write deadline",
			logging.Err(err),
			logging.Component("exec_handler"))
	}
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
	// Ownership check: only the deployer's wallet can exec. An empty
	// Owner field means no requester has claimed this deployment; treat
	// that the same as a mismatched wallet rather than letting it through.
	if deployment.Owner == "" || !strings.EqualFold(deployment.Owner, walletAddr) {
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
	if err := json.NewEncoder(w).Encode(ExecChallengeResponse{
		Nonce:   challenge.Nonce,
		Message: challenge.Message,
	}); err != nil {
		logging.Warn("failed to encode exec challenge response",
			logging.Err(err),
			logging.Component("exec_handler"))
	}
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
	if err := conn.SetReadDeadline(time.Time{}); err != nil {
		logging.Warn("failed to clear websocket read deadline",
			logging.Err(err),
			logging.Component("exec_handler"))
	}
	if err := conn.SetWriteDeadline(time.Time{}); err != nil {
		logging.Warn("failed to clear websocket write deadline",
			logging.Err(err),
			logging.Component("exec_handler"))
	}

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

	// Detect exec-agent mode (E2E encrypted exec). When the container has the
	// exec-agent injected, ExecLocal launches "/usr/local/bin/exec-agent --stdio"
	// instead of a plaintext PTY (mirrors handleExecOpen on the remote path).
	// In agent mode the API server is a TRANSPARENT byte-pipe of the exec-agent
	// frame protocol: it never decrypts, and must translate between the
	// browser/CLI WebSocket framing ([type][payload], no length prefix) and the
	// exec-agent stdio wire framing ([4-byte BE len][type][payload]).
	execAgentMode := false
	if deployment, ok := cm.GetDeployment(session.ContainerID); ok {
		execAgentMode = deployment.ExecAgentEnabled
	}

	// Open the session on the local container runtime (with ownership
	// verification). ExecLocal returns either a plaintext PTY (non-agent) or a
	// raw exec-agent session (agent mode), branching on ExecAgentEnabled.
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
		"exec_agent", execAgentMode,
		logging.Component("exec_api"))

	// Goroutine: PTY/agent stdout → WebSocket (all writes go through ws to serialize)
	done := make(chan struct{})
	if execAgentMode {
		go s.relayAgentStdoutToWS(ws, ptySession.Stdout, session, done)
	} else {
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
					if writeErr := ws.writeMessage(closeFrame); writeErr != nil {
						logging.Warn("failed to write close frame on PTY end",
							"session_id", session.SessionID,
							"error", writeErr.Error(),
							logging.Component("exec_api"))
					}
					return
				}
			}
		}()
	}

	// Main loop: WebSocket → PTY stdin
	conn.SetReadLimit(64 * 1024)
	if err := conn.SetReadDeadline(time.Now().Add(wsReadWait)); err != nil {
		logging.Warn("failed to set initial read deadline on exec WebSocket",
			"session_id", session.SessionID,
			"error", err.Error(),
			logging.Component("exec_api"))
	}
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
		if err := conn.SetReadDeadline(time.Now().Add(wsReadWait)); err != nil {
			logging.Warn("failed to extend read deadline on exec WebSocket",
				"session_id", session.SessionID,
				"error", err.Error(),
				logging.Component("exec_api"))
		}

		if len(message) == 0 {
			continue
		}

		frameType := message[0]
		frameData := message[1:]

		// Agent mode: translate every browser/CLI WS frame into an exec-agent
		// stdio wire frame ([4-byte BE len][type][payload]) and write it to the
		// agent's stdin. The agent handles resize/key-exchange/ciphertext itself;
		// the API server never decrypts and never strips the frame type byte.
		if execAgentMode {
			switch frameType {
			case WSFrameData:
				// Plain terminal DATA: the WS type byte is stripped on the wire,
				// so re-stamp it as a DATA (0x01) agent frame.
				if writeErr := writeAgentStdinFrame(ptySession.Stdin, WSFrameData, frameData); writeErr != nil {
					logging.Warn("failed to write DATA frame to agent stdin",
						"session_id", session.SessionID,
						"error", writeErr.Error(),
						logging.Component("exec_api"))
				}
				s.execSessions.AddBytes(session.SessionID, 0, int64(len(frameData)))
			case WSFrameResize, WSFrameKeyInit, WSFrameKeyAck, WSFramePing, WSFramePong, WSFrameClose:
				// Control frames carry their semantic type as message[0]; relay
				// that type verbatim to the agent (resize, key exchange, etc.).
				if writeErr := writeAgentStdinFrame(ptySession.Stdin, frameType, frameData); writeErr != nil {
					logging.Warn("failed to relay control frame to agent stdin",
						"session_id", session.SessionID,
						"frame_type", frameType,
						"error", writeErr.Error(),
						logging.Component("exec_api"))
				}
				if frameType == WSFrameClose {
					goto cleanup
				}
			}
			continue
		}

		switch frameType {
		case WSFrameData:
			if _, writeErr := ptySession.Stdin.Write(frameData); writeErr != nil {
				logging.Warn("failed to write data to PTY stdin",
					"session_id", session.SessionID,
					"error", writeErr.Error(),
					logging.Component("exec_api"))
			}
			s.execSessions.AddBytes(session.SessionID, 0, int64(len(frameData)))
		case WSFrameResize:
			if len(frameData) >= 4 {
				newCols := uint16(frameData[0])<<8 | uint16(frameData[1])
				newRows := uint16(frameData[2])<<8 | uint16(frameData[3])
				if resizeErr := ptySession.Resize(newCols, newRows); resizeErr != nil {
					logging.Warn("failed to resize PTY",
						"session_id", session.SessionID,
						"error", resizeErr.Error(),
						logging.Component("exec_api"))
				}
			}
		case WSFrameKeyInit, WSFrameKeyAck:
			// E2E key exchange frames — relay to container stdin as exec-agent frames
			if _, writeErr := ptySession.Stdin.Write(frameData); writeErr != nil {
				logging.Warn("failed to relay key exchange frame to PTY stdin",
					"session_id", session.SessionID,
					"error", writeErr.Error(),
					logging.Component("exec_api"))
			}
		case WSFramePing:
			if writeErr := ws.writeMessage([]byte{WSFramePong}); writeErr != nil {
				logging.Warn("failed to write pong frame",
					"session_id", session.SessionID,
					"error", writeErr.Error(),
					logging.Component("exec_api"))
			}
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

// maxAgentFrameSize bounds a single exec-agent stdio frame payload (1MB).
// Must match cmd/exec-agent/protocol.go maxFrameSize.
const maxAgentFrameSize = 1 << 20

// writeAgentStdinFrame encodes a single exec-agent stdio wire frame and writes
// it to the agent's stdin. Wire format: [4-byte BE total length][1-byte type]
// [payload], matching cmd/exec-agent/protocol.go writeFrame. The frame type uses
// the shared WS/agent frame numbers (DATA 0x01, RESIZE 0x02, KEY_INIT 0x07, …).
func writeAgentStdinFrame(w io.Writer, frameType byte, payload []byte) error {
	frame := make([]byte, 4+1+len(payload))
	binary.BigEndian.PutUint32(frame[0:4], uint32(1+len(payload)))
	frame[4] = frameType
	copy(frame[5:], payload)
	_, err := w.Write(frame)
	return err
}

// relayAgentStdoutToWS reads length-prefixed exec-agent stdio frames from the
// agent's stdout and forwards each as a top-level WebSocket frame, preserving the
// frame type byte. DATA frames carry opaque ciphertext; control frames (KEY_ACK,
// ERROR, PONG, …) keep their type so the browser/CLI can distinguish them. The
// type byte is stamped as message[0], matching what the local inbound path and
// the CLI client (decodeControlFrame) expect. On agent stdout EOF/CLOSE a
// WSFrameClose is sent. This is the local-path mirror of the daemon ExecStream
// readLoopFramed relay; the API server never decrypts.
func (s *Server) relayAgentStdoutToWS(ws *safeWSConn, r io.Reader, session *ExecSession, done chan struct{}) {
	defer close(done)
	pumpAgentStdoutFrames(r, ws.writeMessage, func(n int) {
		s.execSessions.AddBytes(session.SessionID, int64(n), 0)
	}, session.SessionID)
}

// pumpAgentStdoutFrames is the pure frame-translation loop behind
// relayAgentStdoutToWS, factored out for testability. It reads exec-agent stdio
// wire frames ([4-byte BE len][type][payload]) from r and, for each, calls
// writeFrame with a top-level WS frame [type][payload]. onData(n) is invoked with
// the ciphertext length for DATA frames (byte accounting). On EOF, an invalid
// frame length, or an agent CLOSE frame, it emits a final WSFrameClose and
// returns. If writeFrame errors, the pump stops without emitting a close.
func pumpAgentStdoutFrames(r io.Reader, writeFrame func([]byte) error, onData func(n int), sessionID string) {
	var lenBuf [4]byte
	for {
		if _, err := io.ReadFull(r, lenBuf[:]); err != nil {
			break
		}
		totalLen := binary.BigEndian.Uint32(lenBuf[:])
		if totalLen < 1 || totalLen > maxAgentFrameSize {
			logging.Warn("exec-agent invalid stdout frame length",
				"session_id", sessionID,
				"length", totalLen,
				logging.Component("exec_api"))
			break
		}
		frameBuf := make([]byte, totalLen)
		if _, err := io.ReadFull(r, frameBuf); err != nil {
			break
		}
		agentType := frameBuf[0]
		payload := frameBuf[1:]

		if agentType == WSFrameClose {
			break
		}

		// Forward as a top-level WS frame [type][payload]. For DATA this is
		// [WSFrameData][ciphertext]; for control frames [WSFrameKeyAck][...] etc.
		wsFrame := make([]byte, 1+len(payload))
		wsFrame[0] = agentType
		copy(wsFrame[1:], payload)
		if err := writeFrame(wsFrame); err != nil {
			return
		}
		if agentType == WSFrameData && onData != nil {
			onData(len(payload))
		}
	}

	// Agent exited or closed the stream — notify the client.
	closeFrame := []byte{WSFrameClose}
	closeFrame = append(closeFrame, []byte("session_ended")...)
	if writeErr := writeFrame(closeFrame); writeErr != nil {
		logging.Warn("failed to write close frame on agent stdout end",
			"session_id", sessionID,
			"error", writeErr.Error(),
			logging.Component("exec_api"))
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
				if writeErr := ws.writeMessage(closeFrame); writeErr != nil {
					logging.Warn("failed to write provider-close frame",
						"session_id", session.SessionID,
						"error", writeErr.Error(),
						logging.Component("exec_api"))
				}
				return

			case <-relayCtx.Done():
				return
			}
		}
	}()

	// Main loop: WebSocket → Provider (reads from browser)
	conn.SetReadLimit(64 * 1024) // 64KB max message
	if err := conn.SetReadDeadline(time.Now().Add(wsReadWait)); err != nil {
		logging.Warn("failed to set initial read deadline on remote exec WebSocket",
			"session_id", session.SessionID,
			"error", err.Error(),
			logging.Component("exec_api"))
	}
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
		if err := conn.SetReadDeadline(time.Now().Add(wsReadWait)); err != nil {
			logging.Warn("failed to extend read deadline on remote exec WebSocket",
				"session_id", session.SessionID,
				"error", err.Error(),
				logging.Component("exec_api"))
		}

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
			if sendErr := cm.SendExecMessage(ctx, providerID, &types.Message{
				Type:      types.MessageTypeExecData,
				From:      cm.LocalNodeID(),
				To:        providerID,
				Payload:   pb,
				Timestamp: time.Now(),
			}); sendErr != nil {
				logging.Warn("failed to send exec data to provider",
					"session_id", session.SessionID,
					"provider_id", providerID,
					"error", sendErr.Error(),
					logging.Component("exec_api"))
			}
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
				if sendErr := cm.SendExecMessage(ctx, providerID, &types.Message{
					Type:      types.MessageTypeExecResize,
					From:      cm.LocalNodeID(),
					To:        providerID,
					Payload:   pb,
					Timestamp: time.Now(),
				}); sendErr != nil {
					logging.Warn("failed to send exec resize to provider",
						"session_id", session.SessionID,
						"provider_id", providerID,
						"error", sendErr.Error(),
						logging.Component("exec_api"))
				}
			}

		case WSFrameKeyInit, WSFrameKeyAck:
			// E2E key exchange frames — forward as ExecData to provider (opaque relay).
			// The provider's ExecStream wraps it in exec-agent framing automatically.
			dataPayload := types.ExecDataPayload{
				SessionID: session.SessionID,
				Data:      message, // include frame type byte — provider relays as-is
			}
			pb, _ := json.Marshal(dataPayload)
			if sendErr := cm.SendExecMessage(ctx, providerID, &types.Message{
				Type:      types.MessageTypeExecData,
				From:      cm.LocalNodeID(),
				To:        providerID,
				Payload:   pb,
				Timestamp: time.Now(),
			}); sendErr != nil {
				logging.Warn("failed to relay key exchange frame to provider",
					"session_id", session.SessionID,
					"provider_id", providerID,
					"error", sendErr.Error(),
					logging.Component("exec_api"))
			}

		case WSFramePing:
			if writeErr := ws.writeMessage([]byte{WSFramePong}); writeErr != nil {
				logging.Warn("failed to write pong frame to remote exec WebSocket",
					"session_id", session.SessionID,
					"error", writeErr.Error(),
					logging.Component("exec_api"))
			}

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
	if sendErr := cm.SendExecMessage(ctx, providerID, &types.Message{
		Type:      types.MessageTypeExecClose,
		From:      cm.LocalNodeID(),
		To:        providerID,
		Payload:   pb,
		Timestamp: time.Now(),
	}); sendErr != nil {
		logging.Warn("failed to notify provider of exec session close",
			"session_id", session.SessionID,
			"provider_id", providerID,
			"error", sendErr.Error(),
			logging.Component("exec_api"))
	}

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
	if err := conn.WriteMessage(websocket.BinaryMessage, frame); err != nil {
		logging.Warn("failed to write exec error frame",
			"error", err.Error(),
			logging.Component("exec_api"))
	}
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
