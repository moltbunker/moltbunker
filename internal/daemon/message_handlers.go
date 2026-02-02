package daemon

import (
	"context"
	"encoding/json"
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

		// Non-blocking send to ack channel only if not closed
		// Use a local copy of closed status under the lock
		isClosed := pending.closed
		pending.mu.Unlock()

		if !isClosed {
			select {
			case pending.ackChan <- replicaAckData:
			default:
				// Channel full, ack is still recorded in the slice
			}
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

	// Update consensus state
	cm.consensus.UpdateState(syncData.ContainerID, syncData.Status, [3]*types.Container{})

	return nil
}
