package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/moltbunker/moltbunker/internal/logging"
	"github.com/moltbunker/moltbunker/pkg/types"
)

// registerMessageHandlers registers handlers for container-related P2P messages
func (cm *ContainerManager) registerMessageHandlers() {
	// Handler for deployment requests from other nodes
	cm.router.RegisterHandler(types.MessageTypeDeploy, cm.handleDeployRequest)

	// Handler for deployment acknowledgments
	cm.router.RegisterHandler(types.MessageTypeDeployAck, cm.handleDeployAck)

	// Handler for container status queries
	cm.router.RegisterHandler(types.MessageTypeContainerStatus, cm.handleStatusRequest)

	// Handler for replica sync
	cm.router.RegisterHandler(types.MessageTypeReplicaSync, cm.handleReplicaSync)

	// Handler for remote container stop
	cm.router.RegisterHandler(types.MessageTypeStop, cm.handleRemoteStop)

	// Handler for remote container delete
	cm.router.RegisterHandler(types.MessageTypeDelete, cm.handleRemoteDelete)

	// Handler for remote log requests
	cm.router.RegisterHandler(types.MessageTypeLogs, cm.handleRemoteLogs)

	// Handler for gossip messages
	if cm.gossip != nil {
		cm.router.RegisterHandler(types.MessageTypeGossip, cm.gossip.HandleGossipMessage)
	}
}

// handleStatusRequest handles container status queries
func (cm *ContainerManager) handleStatusRequest(ctx context.Context, msg *types.Message, from *types.Node) error {
	containerID := string(msg.Payload)

	cm.mu.RLock()
	deployment, exists := cm.deployments[containerID]
	cm.mu.RUnlock()

	if !exists {
		return nil
	}

	// Get current status
	status := deployment.Status
	if cm.containerd != nil {
		if s, err := cm.containerd.GetContainerStatus(ctx, containerID); err == nil {
			status = s
		}
	}

	// Send status response
	response := map[string]interface{}{
		"container_id": containerID,
		"status":       status,
		"started_at":   deployment.StartedAt,
	}
	responseData, _ := json.Marshal(response)

	return cm.router.SendMessage(ctx, from.ID, &types.Message{
		Type:      types.MessageTypeContainerStatus,
		From:      cm.node.nodeInfo.ID,
		To:        from.ID,
		Payload:   responseData,
		Timestamp: time.Now(),
	})
}

// handleDeployAck handles deployment acknowledgment from replica nodes
func (cm *ContainerManager) handleDeployAck(ctx context.Context, msg *types.Message, from *types.Node) error {
	var ack struct {
		ContainerID string `json:"container_id"`
		Success     bool   `json:"success"`
		Error       string `json:"error"`
		NodeID      string `json:"node_id"`
		Region      string `json:"region"`
	}

	if err := json.Unmarshal(msg.Payload, &ack); err != nil {
		return err
	}

	// Truncate NodeID for logging (handle short IDs gracefully)
	nodeIDForLog := ack.NodeID
	if len(nodeIDForLog) > 16 {
		nodeIDForLog = nodeIDForLog[:16]
	}

	if ack.Success {
		logging.Info("replica confirmed",
			logging.ContainerID(ack.ContainerID),
			logging.NodeID(nodeIDForLog),
			logging.Region(ack.Region))

		// Update health status for this replica
		cm.healthMonitor.UpdateHealth(ack.ContainerID, 1, types.HealthStatus{
			Healthy:    true,
			LastUpdate: time.Now(),
		})
	} else {
		logging.Error("replica deployment failed",
			logging.ContainerID(ack.ContainerID),
			logging.NodeID(nodeIDForLog),
			"reason", ack.Error)

		// Mark replica as unhealthy
		cm.healthMonitor.UpdateHealth(ack.ContainerID, 1, types.HealthStatus{
			Healthy:    false,
			LastUpdate: time.Now(),
		})
	}

	// Notify pending deployment tracker if exists
	cm.pendingMu.RLock()
	pending, exists := cm.pendingDeployments[ack.ContainerID]
	cm.pendingMu.RUnlock()

	if exists && pending != nil {
		replicaAckData := replicaAck{
			NodeID:  ack.NodeID,
			Region:  ack.Region,
			Success: ack.Success,
			Error:   ack.Error,
		}

		// Update pending deployment state and send to channel under lock
		pending.mu.Lock()
		pending.ackCount++
		if ack.Success {
			pending.successCount++
		}
		pending.acks = append(pending.acks, replicaAckData)

		// Activate escrow once on first successful ack (transition Created → Active)
		shouldActivate := ack.Success && !pending.escrowActivated
		var acksSnapshot []replicaAck
		if shouldActivate {
			pending.escrowActivated = true
			acksSnapshot = make([]replicaAck, len(pending.acks))
			copy(acksSnapshot, pending.acks)
		}

		// Non-blocking send to ack channel under the same lock that
		// protects close(). This prevents a TOCTOU race where close()
		// could close the channel between our check and send.
		if !pending.closed {
			select {
			case pending.ackChan <- replicaAckData:
			default:
				// Channel full, ack is still recorded in the slice
			}
		}
		pending.mu.Unlock()

		// Call SelectProviders outside the lock to avoid blocking ack processing
		if shouldActivate {
			cm.activateEscrow(ctx, ack.ContainerID, acksSnapshot)
		}
	}

	return nil
}

// handleReplicaSync handles replica synchronization messages
func (cm *ContainerManager) handleReplicaSync(ctx context.Context, msg *types.Message, from *types.Node) error {
	// Sync replica state between nodes
	var syncData struct {
		ContainerID string                `json:"container_id"`
		Status      types.ContainerStatus `json:"status"`
		ReplicaIdx  int                   `json:"replica_idx"`
	}

	if err := json.Unmarshal(msg.Payload, &syncData); err != nil {
		return err
	}

	// Update consensus state with the specific replica's status
	cm.consensus.UpdateReplicaStatus(syncData.ContainerID, syncData.ReplicaIdx, syncData.Status)

	return nil
}

// handleRemoteStop handles remote container stop requests from peer nodes.
// Only the originator of the deployment is authorized to stop it.
func (cm *ContainerManager) handleRemoteStop(ctx context.Context, msg *types.Message, from *types.Node) error {
	containerID := string(msg.Payload)

	// Verify the sender is the deployment originator
	cm.mu.RLock()
	deployment, exists := cm.deployments[containerID]
	cm.mu.RUnlock()
	if !exists {
		return nil // Unknown container, ignore
	}
	if deployment.OriginatorID != msg.From {
		logging.Warn("rejecting remote stop: sender is not originator",
			logging.ContainerID(containerID),
			logging.NodeID(msg.From.String()[:16]),
			logging.Component("message_handlers"))
		return nil
	}

	logging.Info("received remote stop request",
		logging.ContainerID(containerID),
		logging.NodeID(from.ID.String()[:16]),
		logging.Component("message_handlers"))

	if err := cm.Stop(ctx, containerID); err != nil {
		logging.Warn("remote stop failed",
			logging.ContainerID(containerID),
			logging.Err(err),
			logging.Component("message_handlers"))
		return err
	}

	return nil
}

// handleRemoteDelete handles remote container delete requests from peer nodes.
// Only the originator of the deployment is authorized to delete it.
func (cm *ContainerManager) handleRemoteDelete(ctx context.Context, msg *types.Message, from *types.Node) error {
	containerID := string(msg.Payload)

	// Verify the sender is the deployment originator
	cm.mu.RLock()
	deployment, exists := cm.deployments[containerID]
	cm.mu.RUnlock()
	if !exists {
		return nil // Unknown container, ignore
	}
	if deployment.OriginatorID != msg.From {
		logging.Warn("rejecting remote delete: sender is not originator",
			logging.ContainerID(containerID),
			logging.NodeID(msg.From.String()[:16]),
			logging.Component("message_handlers"))
		return nil
	}

	logging.Info("received remote delete request",
		logging.ContainerID(containerID),
		logging.NodeID(from.ID.String()[:16]),
		logging.Component("message_handlers"))

	if err := cm.Delete(ctx, containerID); err != nil {
		logging.Warn("remote delete failed",
			logging.ContainerID(containerID),
			logging.Err(err),
			logging.Component("message_handlers"))
		return err
	}

	return nil
}

// handleRemoteLogs handles remote log requests from peer nodes
func (cm *ContainerManager) handleRemoteLogs(ctx context.Context, msg *types.Message, from *types.Node) error {
	var req struct {
		ContainerID string `json:"container_id"`
		Lines       int    `json:"lines"`
	}

	if err := json.Unmarshal(msg.Payload, &req); err != nil {
		return err
	}

	if req.Lines == 0 {
		req.Lines = 100
	}

	// Get logs from containerd
	var logs string
	if cm.containerd != nil {
		reader, err := cm.containerd.GetContainerLogs(ctx, req.ContainerID, false, req.Lines)
		if err != nil {
			logs = fmt.Sprintf("error retrieving logs: %v", err)
		} else {
			defer reader.Close()
			data, err := io.ReadAll(reader)
			if err != nil {
				logs = fmt.Sprintf("error reading logs: %v", err)
			} else {
				logs = string(data)
			}
		}
	} else {
		logs = "containerd not available"
	}

	// Send logs response
	response := map[string]string{
		"container_id": req.ContainerID,
		"logs":         logs,
	}
	responseData, _ := json.Marshal(response)

	return cm.router.SendMessage(ctx, from.ID, &types.Message{
		Type:      types.MessageTypeLogs,
		From:      cm.node.nodeInfo.ID,
		To:        from.ID,
		Payload:   responseData,
		Timestamp: time.Now(),
	})
}
