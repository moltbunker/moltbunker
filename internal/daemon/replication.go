package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/moltbunker/moltbunker/internal/logging"
	"github.com/moltbunker/moltbunker/internal/p2p"
	"github.com/moltbunker/moltbunker/internal/util"
	"github.com/moltbunker/moltbunker/pkg/types"
)

// broadcastDeployment broadcasts deployment to network for redundancy
func (cm *ContainerManager) broadcastDeployment(ctx context.Context, deployment *Deployment) error {
	// Get peers in target regions
	peers := cm.router.GetPeers()

	if len(peers) == 0 {
		return fmt.Errorf("no peers available for replication")
	}

	// Find nodes in different regions for replication
	selectedNodes, err := cm.geoRouter.SelectNodesForReplication(peers)
	if err != nil {
		// Log warning but continue with available nodes
		logging.Warn("geographic selection failed, using available peers",
			logging.Err(err),
			"peer_count", len(peers))
		selectedNodes = peers
		if len(selectedNodes) > 3 {
			selectedNodes = selectedNodes[:3]
		}
	}

	// Send deployment request to selected nodes
	deployData, err := json.Marshal(deployment)
	if err != nil {
		return fmt.Errorf("failed to marshal deployment: %w", err)
	}

	msg := &types.Message{
		Type:      types.MessageTypeDeploy,
		From:      cm.node.nodeInfo.ID,
		Payload:   deployData,
		Timestamp: time.Now(),
	}

	var successCount int
	var lastErr error
	for _, node := range selectedNodes {
		if node.ID == cm.node.nodeInfo.ID {
			continue // Skip self
		}
		if err := cm.router.SendMessage(ctx, node.ID, msg); err != nil {
			lastErr = err
			logging.Warn("failed to send deployment to peer",
				logging.ContainerID(deployment.ID),
				logging.NodeID(node.ID.String()[:16]),
				logging.Err(err))
			continue
		}
		successCount++
	}

	if successCount == 0 && lastErr != nil {
		return fmt.Errorf("failed to replicate to any peer: %w", lastErr)
	}

	logging.Info("deployment replicated",
		logging.ContainerID(deployment.ID),
		"peer_count", successCount)
	return nil
}

// handleDeployRequest handles incoming deployment requests from other nodes
func (cm *ContainerManager) handleDeployRequest(ctx context.Context, msg *types.Message, from *types.Node) error {
	var deployment Deployment
	if err := json.Unmarshal(msg.Payload, &deployment); err != nil {
		return err
	}

	// Check if we should accept this deployment
	// (based on resources, region requirements, etc.)

	cm.mu.Lock()
	if _, exists := cm.deployments[deployment.ID]; exists {
		cm.mu.Unlock()
		return nil // Already have this deployment
	}

	// Make a copy of the deployment to store (avoid race conditions)
	storedDeployment := deployment
	cm.deployments[deployment.ID] = &storedDeployment

	// Copy regions for use after unlock (avoid race on slice access)
	regions := make([]string, len(deployment.Regions))
	copy(regions, deployment.Regions)

	// Make a copy of the deployment BEFORE releasing the lock for use in goroutine
	deploymentCopy := deployment
	originatorID := msg.From
	cm.mu.Unlock()

	// If we have containerd and this is for our region, deploy locally
	if cm.containerd != nil {
		myRegion := p2p.GetRegionFromCountry(cm.node.nodeInfo.Country)
		for _, region := range regions {
			if region == myRegion {
				// We're a target for this deployment - run async but log errors
				util.SafeGoWithName("deploy-replica", func() {
					if err := cm.deployReplica(ctx, &deploymentCopy, originatorID); err != nil {
						logging.Warn("failed to deploy replica",
							logging.ContainerID(deploymentCopy.ID),
							logging.Err(err))
					}
				})
				break
			}
		}
	}

	return nil
}

// deployReplica deploys a replica of a container locally and sends acknowledgment
func (cm *ContainerManager) deployReplica(ctx context.Context, deployment *Deployment, originatorID types.NodeID) error {
	// Create container (this will pull the image if not present)
	logging.Info("pulling image and creating replica container",
		logging.ContainerID(deployment.ID),
		"image", deployment.Image)
	managed, err := cm.containerd.CreateContainer(ctx, deployment.ID, deployment.Image, deployment.Resources)
	if err != nil {
		logging.Error("failed to create replica container",
			logging.ContainerID(deployment.ID),
			logging.Err(err))
		cm.sendDeployAck(ctx, originatorID, deployment.ID, false, err.Error())
		return fmt.Errorf("failed to create container: %w", err)
	}

	logging.Info("created replica container",
		logging.ContainerID(deployment.ID),
		"image", deployment.Image)
	_ = managed // Use the managed container

	// Start container
	if err := cm.containerd.StartContainer(ctx, deployment.ID); err != nil {
		logging.Error("failed to start replica container",
			logging.ContainerID(deployment.ID),
			logging.Err(err))
		cm.containerd.DeleteContainer(ctx, deployment.ID)
		cm.sendDeployAck(ctx, originatorID, deployment.ID, false, err.Error())
		return fmt.Errorf("failed to start container: %w", err)
	}

	cm.mu.Lock()
	deployment.Status = types.ContainerStatusRunning
	deployment.StartedAt = time.Now()
	cm.mu.Unlock()

	logging.Info("replica container started successfully",
		logging.ContainerID(deployment.ID))

	// Send acknowledgment back to originator
	cm.sendDeployAck(ctx, originatorID, deployment.ID, true, "")

	return nil
}

// sendDeployAck sends a deployment acknowledgment message
func (cm *ContainerManager) sendDeployAck(ctx context.Context, to types.NodeID, containerID string, success bool, errMsg string) {
	ack := map[string]interface{}{
		"container_id": containerID,
		"success":      success,
		"error":        errMsg,
		"node_id":      cm.node.nodeInfo.ID.String(),
		"region":       cm.node.nodeInfo.Region,
	}

	ackData, err := json.Marshal(ack)
	if err != nil {
		return
	}

	msg := &types.Message{
		Type:      types.MessageTypeDeployAck,
		From:      cm.node.nodeInfo.ID,
		To:        to,
		Payload:   ackData,
		Timestamp: time.Now(),
	}

	if err := cm.router.SendMessage(ctx, to, msg); err != nil {
		logging.Warn("failed to send deploy ack",
			logging.NodeID(to.String()[:16]),
			logging.Err(err))
	}
}

// WaitForReplicas waits for replica acknowledgments with the given timeout.
// Returns the number of successful replica acks received.
func (cm *ContainerManager) WaitForReplicas(containerID string, timeout time.Duration) (int, error) {
	cm.pendingMu.RLock()
	pending, exists := cm.pendingDeployments[containerID]
	cm.pendingMu.RUnlock()

	if !exists || pending == nil {
		return 0, fmt.Errorf("no pending deployment found for container: %s", containerID)
	}

	// Check if we already have successful acks
	pending.mu.Lock()
	if pending.successCount > 0 {
		count := pending.successCount
		pending.mu.Unlock()
		return count, nil
	}
	pending.mu.Unlock()

	// Wait for acks with timeout
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	for {
		select {
		case ack := <-pending.ackChan:
			if ack.Success {
				// Got at least one successful ack
				pending.mu.Lock()
				count := pending.successCount
				pending.mu.Unlock()

				logging.Info("replica verification completed",
					logging.ContainerID(containerID),
					"successful_replicas", count)
				return count, nil
			}
			// Continue waiting for more acks
		case <-timer.C:
			// Timeout reached, return current count
			pending.mu.Lock()
			count := pending.successCount
			totalAcks := pending.ackCount
			pending.mu.Unlock()

			if count == 0 {
				logging.Warn("no replica acknowledgments received within timeout",
					logging.ContainerID(containerID),
					"timeout", timeout.String(),
					"total_acks", totalAcks)
				return 0, fmt.Errorf("timeout waiting for replica acks: received %d acks, %d successful", totalAcks, count)
			}

			return count, nil
		}
	}
}

// GetReplicaStatus returns the current replica status for a deployment
func (cm *ContainerManager) GetReplicaStatus(containerID string) (ackCount int, successCount int, exists bool) {
	cm.pendingMu.RLock()
	pending, exists := cm.pendingDeployments[containerID]
	cm.pendingMu.RUnlock()

	if !exists || pending == nil {
		return 0, 0, false
	}

	pending.mu.Lock()
	defer pending.mu.Unlock()
	return pending.ackCount, pending.successCount, true
}
