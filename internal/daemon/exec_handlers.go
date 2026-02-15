package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/moltbunker/moltbunker/internal/logging"
	"github.com/moltbunker/moltbunker/pkg/types"
)

// registerExecHandlers registers P2P message handlers for exec operations.
// These handlers run on provider nodes that host the containers.
func (cm *ContainerManager) registerExecHandlers() {
	cm.router.RegisterHandler(types.MessageTypeExecOpen, cm.handleExecOpen)
	cm.router.RegisterHandler(types.MessageTypeExecData, cm.handleExecData)
	cm.router.RegisterHandler(types.MessageTypeExecResize, cm.handleExecResize)
	cm.router.RegisterHandler(types.MessageTypeExecClose, cm.handleExecClose)
}

// handleExecOpen handles a request to open an interactive exec session.
// It verifies the container exists and is running, opens a PTY, and starts streaming.
func (cm *ContainerManager) handleExecOpen(ctx context.Context, msg *types.Message, from *types.Node) error {
	var payload types.ExecOpenPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return fmt.Errorf("invalid exec open payload: %w", err)
	}

	logging.Info("exec open request",
		"session_id", payload.SessionID,
		logging.ContainerID(payload.ContainerID),
		"wallet", payload.WalletAddress,
		logging.Component("exec"))

	// Verify container exists on this node
	cm.mu.RLock()
	deployment, exists := cm.deployments[payload.ContainerID]
	cm.mu.RUnlock()

	if !exists {
		return cm.sendExecClose(ctx, msg.From, payload.SessionID, "container_not_found")
	}

	if deployment.Status != types.ContainerStatusRunning {
		return cm.sendExecClose(ctx, msg.From, payload.SessionID, "container_not_running")
	}

	// Only the deployment originator is authorized to exec into the container.
	// This prevents arbitrary peers from gaining shell access to containers they don't own.
	if deployment.OriginatorID != msg.From {
		logging.Warn("rejecting exec: sender is not originator",
			"session_id", payload.SessionID,
			logging.ContainerID(payload.ContainerID),
			logging.NodeID(msg.From.String()[:16]),
			logging.Component("exec"))
		return cm.sendExecClose(ctx, msg.From, payload.SessionID, "not_authorized")
	}

	// Check if exec is allowed by security policy
	if cm.containerd != nil {
		if !cm.containerd.CanExec(payload.ContainerID) {
			return cm.sendExecClose(ctx, msg.From, payload.SessionID, "exec_disabled")
		}
	}

	// Set default terminal size
	cols := uint16(payload.Cols)
	rows := uint16(payload.Rows)
	if cols == 0 {
		cols = 80
	}
	if rows == 0 {
		rows = 24
	}

	// Open interactive PTY session
	session, err := cm.containerd.ExecInteractive(ctx, payload.ContainerID, cols, rows)
	if err != nil {
		logging.Error("exec open failed",
			"session_id", payload.SessionID,
			logging.ContainerID(payload.ContainerID),
			logging.Err(err),
			logging.Component("exec"))
		return cm.sendExecClose(ctx, msg.From, payload.SessionID, fmt.Sprintf("exec_failed: %v", err))
	}

	// Create exec stream to bridge PTY ↔ P2P
	stream := newExecStream(ctx, payload.SessionID, payload.ContainerID, msg.From, cm.node.nodeInfo.ID, session, cm.router)

	// Register with stream manager
	if err := cm.execStreams.Add(stream); err != nil {
		stream.Close()
		return cm.sendExecClose(ctx, msg.From, payload.SessionID, fmt.Sprintf("limit: %v", err))
	}

	logging.Info("exec session started",
		"session_id", payload.SessionID,
		logging.ContainerID(payload.ContainerID),
		"cols", cols,
		"rows", rows,
		logging.Component("exec"))

	// Log audit event
	logging.Audit(logging.AuditEvent{
		Operation: "exec_session_start",
		Actor:     payload.WalletAddress,
		Target:    payload.ContainerID,
		Result:    "success",
		Details:   fmt.Sprintf("session=%s cols=%d rows=%d", payload.SessionID, cols, rows),
	})

	return nil
}

// handleExecData handles incoming terminal data.
// On provider side: keystrokes from user → write to container stdin.
// On requester side: container output from provider → forward to WebSocket relay.
func (cm *ContainerManager) handleExecData(_ context.Context, msg *types.Message, _ *types.Node) error {
	var payload types.ExecDataPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return fmt.Errorf("invalid exec data payload: %w", err)
	}

	// Provider side: session exists locally, write to container stdin
	stream, ok := cm.execStreams.Get(payload.SessionID)
	if ok {
		return stream.WriteData(payload.Data)
	}

	// Requester side: relay container output to WebSocket
	relay, ok := cm.getExecRelay(payload.SessionID)
	if ok && relay.OnData != nil {
		relay.OnData(payload.Data)
		return nil
	}

	return nil // Session may have ended; silently ignore
}

// handleExecResize handles terminal resize events.
func (cm *ContainerManager) handleExecResize(_ context.Context, msg *types.Message, _ *types.Node) error {
	var payload types.ExecResizePayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return fmt.Errorf("invalid exec resize payload: %w", err)
	}

	stream, ok := cm.execStreams.Get(payload.SessionID)
	if !ok {
		return nil
	}

	return stream.Resize(uint16(payload.Cols), uint16(payload.Rows))
}

// handleExecClose handles a request to close an exec session.
// On provider side: requester asked to close → stop PTY.
// On requester side: provider reports session ended → notify WebSocket relay.
func (cm *ContainerManager) handleExecClose(_ context.Context, msg *types.Message, _ *types.Node) error {
	var payload types.ExecClosePayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return fmt.Errorf("invalid exec close payload: %w", err)
	}

	// Provider side: close the PTY stream
	stream, ok := cm.execStreams.Get(payload.SessionID)
	if ok {
		logging.Info("exec session closed by requester",
			"session_id", payload.SessionID,
			"reason", payload.Reason,
			logging.Component("exec"))
		stream.Close()
		return nil
	}

	// Requester side: provider reports session ended → notify WebSocket
	relay, ok := cm.getExecRelay(payload.SessionID)
	if ok && relay.OnClose != nil {
		logging.Info("exec session closed by provider",
			"session_id", payload.SessionID,
			"reason", payload.Reason,
			logging.Component("exec"))
		relay.OnClose(payload.Reason)
		cm.RemoveExecRelay(payload.SessionID)
		return nil
	}

	return nil
}

// sendExecClose sends an exec close message with a reason
func (cm *ContainerManager) sendExecClose(ctx context.Context, to types.NodeID, sessionID, reason string) error {
	payload := types.ExecClosePayload{
		SessionID: sessionID,
		Reason:    reason,
	}
	payloadBytes, _ := json.Marshal(payload)

	return cm.router.SendMessage(ctx, to, &types.Message{
		Type:      types.MessageTypeExecClose,
		From:      cm.node.nodeInfo.ID,
		To:        to,
		Payload:   payloadBytes,
		Timestamp: time.Now(),
	})
}
