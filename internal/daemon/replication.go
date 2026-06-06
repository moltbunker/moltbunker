package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/moltbunker/moltbunker/internal/logging"
	"github.com/moltbunker/moltbunker/internal/p2p"
	"github.com/moltbunker/moltbunker/internal/runtime"
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

	// W11: For Molt deployments, filter to Molt-capable nodes
	var (
		selectedNodes []*types.Node
		err           error
	)
	if deployment.RuntimeType == types.RuntimeTypeMolt {
		selectedNodes, err = cm.geoRouter.SelectMoltNodes(peers)
	} else {
		selectedNodes, err = cm.geoRouter.SelectNodesForReplication(peers, deployment.MinProviderTier)
	}
	if err != nil {
		logging.Warn("geographic selection failed, using available peers",
			logging.Err(err),
			"peer_count", len(peers))
		selectedNodes = peers
		if len(selectedNodes) > 3 {
			selectedNodes = selectedNodes[:3]
		}
	}

	// Filter out providers ineligible for jobs (slashed, low reputation, unstaked)
	if cm.payment != nil {
		eligible := make([]*types.Node, 0, len(selectedNodes))
		for _, n := range selectedNodes {
			// Look up provider wallet from NodeID via stake verifier
			sv := cm.router.StakeVerifier()
			if sv == nil {
				eligible = append(eligible, n) // No stake verifier — accept all
				continue
			}
			providerAddr, known := sv.GetWallet(n.ID)
			if !known {
				eligible = append(eligible, n) // No wallet mapping yet — accept
				continue
			}
			ok, err := cm.payment.IsEligibleForJobs(ctx, providerAddr)
			if err != nil {
				logging.Debug("eligibility check failed, including peer",
					logging.NodeID(n.ID.String()[:16]), logging.Err(err))
				eligible = append(eligible, n) // On error, include (fail-open for availability)
				continue
			}
			if ok {
				eligible = append(eligible, n)
			} else {
				logging.Debug("peer ineligible for jobs, skipping",
					logging.NodeID(n.ID.String()[:16]))
			}
		}
		if len(eligible) > 0 {
			selectedNodes = eligible
		}
		// If all were ineligible, keep original selection to avoid zero-replica deploys
	}

	// Build region list from actual selected nodes for the broadcast payload.
	// Do NOT mutate deployment.Regions — it's shared with cm.deployments and
	// concurrent access would be a data race. Instead, create a copy for marshaling.
	actualRegions := make([]string, 0, len(selectedNodes)+1)
	seen := make(map[string]bool)
	localRegion := cm.node.nodeInfo.Region
	if localRegion != "" {
		actualRegions = append(actualRegions, localRegion)
		seen[localRegion] = true
	}
	for _, node := range selectedNodes {
		region := p2p.GetRegionFromCountry(node.Country)
		if !seen[region] && region != "" && region != "Unknown" {
			actualRegions = append(actualRegions, region)
			seen[region] = true
		}
	}

	// Create a shallow copy with updated regions for the broadcast payload
	broadcastDeployment := *deployment
	if len(actualRegions) > 0 {
		broadcastDeployment.Regions = actualRegions
	}

	// Send deployment request to selected nodes
	deployData, err := json.Marshal(&broadcastDeployment)
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
		sendCtx, sendCancel := context.WithTimeout(ctx, 10*time.Second)
		if err := cm.router.SendMessage(sendCtx, node.ID, msg); err != nil {
			lastErr = err
			logging.Warn("failed to send deployment to peer",
				logging.ContainerID(deployment.ID),
				logging.NodeID(node.ID.String()[:16]),
				logging.Err(err))
			sendCancel()
			continue
		}
		sendCancel()
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

	// Verify the requester has announced their wallet (proves identity & payment ability)
	sv := cm.router.StakeVerifier()
	if sv != nil {
		requesterWallet, hasAnnounced := sv.GetWallet(msg.From)
		if !hasAnnounced {
			logging.Warn("rejecting deployment: requester has not announced wallet",
				logging.ContainerID(deployment.ID),
				logging.NodeID(msg.From.String()[:16]))
			cm.sendDeployAck(ctx, msg.From, deployment.ID, false, "requester identity not verified")
			return nil
		}
		// Verify requester has token balance (basic payment ability check)
		if cm.payment != nil && requesterWallet != (common.Address{}) {
			balance, err := cm.payment.GetTokenBalance(ctx, requesterWallet)
			if err != nil {
				logging.Warn("failed to check requester balance, accepting deployment",
					logging.ContainerID(deployment.ID),
					logging.Err(err))
			} else if balance != nil && balance.Sign() == 0 {
				logging.Warn("rejecting deployment: requester has zero token balance",
					logging.ContainerID(deployment.ID),
					"requester_wallet", requesterWallet.Hex()[:10])
				cm.sendDeployAck(ctx, msg.From, deployment.ID, false, "insufficient token balance")
				return nil
			}
		}
	}

	// Check if this node meets the deployment's minimum provider tier
	if deployment.MinProviderTier != "" {
		localTier := runtime.DetectProviderTier()
		if !localTier.MeetsTierRequirement(deployment.MinProviderTier) {
			logging.Info("rejecting deployment: local tier insufficient",
				logging.ContainerID(deployment.ID),
				"local_tier", string(localTier),
				"required_tier", string(deployment.MinProviderTier))
			cm.sendDeployAck(ctx, msg.From, deployment.ID, false, "provider tier insufficient")
			return nil
		}
	}

	// Check job acceptance flags
	if deployment.RuntimeType == types.RuntimeTypeMolt && !cm.acceptFunctions {
		logging.Info("rejecting deployment: serverless functions not accepted",
			logging.ContainerID(deployment.ID))
		cm.sendDeployAck(ctx, msg.From, deployment.ID, false, "serverless functions not accepted")
		return nil
	}
	if deployment.RuntimeType != types.RuntimeTypeMolt && !cm.acceptServices {
		logging.Info("rejecting deployment: container services not accepted",
			logging.ContainerID(deployment.ID))
		cm.sendDeployAck(ctx, msg.From, deployment.ID, false, "container services not accepted")
		return nil
	}

	// Check if this node has minimum stake to act as provider
	if cm.payment != nil {
		nodeAddr := cm.node.WalletAddress()
		if nodeAddr != (common.Address{}) {
			hasStake, err := cm.payment.HasMinimumStake(ctx, nodeAddr)
			if err != nil {
				logging.Warn("failed to check staking status, accepting deployment",
					logging.ContainerID(deployment.ID),
					logging.Err(err))
			} else if !hasStake {
				logging.Warn("rejecting deployment: insufficient stake",
					logging.ContainerID(deployment.ID))
				cm.sendDeployAck(ctx, msg.From, deployment.ID, false, "insufficient stake")
				return nil
			}
		}
	}

	cm.mu.Lock()
	if _, exists := cm.deployments[deployment.ID]; exists {
		cm.mu.Unlock()
		return nil // Already have this deployment
	}

	// Make a copy of the deployment to store (avoid race conditions)
	storedDeployment := deployment
	storedDeployment.OriginatorID = msg.From // Track who originated this deployment
	cm.deployments[deployment.ID] = &storedDeployment

	// Copy regions for use after unlock (avoid race on slice access)
	regions := make([]string, len(deployment.Regions))
	copy(regions, deployment.Regions)

	// Make a copy of the deployment BEFORE releasing the lock for use in goroutine
	deploymentCopy := deployment
	originatorID := msg.From
	cm.mu.Unlock()

	// W9: Route Molt (WASM) deployments to MoltManager instead of containerd
	if deploymentCopy.RuntimeType == types.RuntimeTypeMolt {
		if cm.moltManager == nil || !cm.moltManager.Available() {
			logging.Warn("rejecting molt deployment: runtime not available",
				logging.ContainerID(deploymentCopy.ID))
			cm.sendDeployAck(ctx, msg.From, deploymentCopy.ID, false, "molt runtime not available")
			return nil
		}
		util.SafeGoWithName("deploy-molt-replica", func() {
			deployCtx, deployCancel := context.WithTimeout(context.Background(), 2*time.Minute)
			defer deployCancel()
			if err := cm.deployMoltReplica(deployCtx, &deploymentCopy, originatorID); err != nil {
				logging.Warn("failed to deploy molt replica",
					logging.ContainerID(deploymentCopy.ID),
					logging.Err(err))
			}
		})
		return nil
	}

	// If we have containerd and this is for our region, deploy locally
	if cm.containerd != nil {
		myRegion := p2p.GetRegionFromCountry(cm.node.nodeInfo.Country)
		for _, region := range regions {
			if region == myRegion {
				// We're a target for this deployment - run async but log errors.
				// Use a detached context: the parent ctx is scoped to the P2P handler
				// and may be cancelled when the handler returns, which would kill a
				// multi-minute image pull mid-stream.
				util.SafeGoWithName("deploy-replica", func() {
					deployCtx, deployCancel := context.WithTimeout(context.Background(), 10*time.Minute)
					defer deployCancel()
					if err := cm.deployReplica(deployCtx, &deploymentCopy, originatorID); err != nil {
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
	// Create container with security hardening (this will pull the image if not present)
	logging.Info("pulling image and creating replica container",
		logging.ContainerID(deployment.ID),
		"image", deployment.Image)
	secConfig := runtime.SecureContainerConfig{
		ID:              deployment.ID,
		ImageRef:        deployment.Image,
		Resources:       deployment.Resources,
		SecurityProfile: types.DeploymentSecurityProfile(),
	}

	// Best-effort exec-agent seeding on the replica. The gossiped Deployment
	// carries the ECIES exec-key envelope + deploy nonce. The envelope is sealed
	// to the ORIGINATOR provider's stable X25519 public key, so this replica can
	// only unwrap it if the envelope was sealed to its own key. Today peers do
	// not advertise per-node X25519 keys, so prepareExecAgent's unwrap will fail
	// here for foreign replicas — in that case we SKIP exec seeding and log it,
	// without breaking replication. Once per-replica re-encryption (sealing the
	// exec_key to each target's advertised X25519 pubkey) lands, this path will
	// succeed transparently because the envelope will already be addressed to
	// this provider's key.
	if len(deployment.EncryptedExecKey) > 0 {
		mounts, keyPath, execErr := cm.prepareExecAgent(deployment.ID, deployment.EncryptedExecKey, deployment.RequesterEphemeralPubKey)
		if execErr != nil {
			logging.Info("skipping exec-agent seeding on replica (envelope not sealed to this provider's key); replica continues without E2E exec",
				logging.ContainerID(deployment.ID),
				logging.Err(execErr),
				logging.Component("replication"))
		} else {
			secConfig.BindMounts = mounts
			cm.mu.Lock()
			deployment.ExecAgentEnabled = true
			deployment.ExecKeyPath = keyPath
			cm.mu.Unlock()
		}
	}

	managed, err := cm.containerd.CreateSecureContainer(ctx, secConfig)
	if err != nil {
		logging.Error("failed to create replica container",
			logging.ContainerID(deployment.ID),
			logging.Err(err))
		cm.mu.Lock()
		cm.cleanupExecKey(deployment)
		cm.mu.Unlock()
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
		if cleanupErr := cm.containerd.DeleteContainer(ctx, deployment.ID); cleanupErr != nil {
			logging.Warn("failed to delete replica container during cleanup",
				logging.ContainerID(deployment.ID),
				logging.Err(cleanupErr),
				logging.Component("replication"))
		}
		cm.mu.Lock()
		cm.cleanupExecKey(deployment)
		cm.mu.Unlock()
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

// deployMoltReplica deploys a Molt (WASM) replica and sends acknowledgment
func (cm *ContainerManager) deployMoltReplica(ctx context.Context, deployment *Deployment, originatorID types.NodeID) error {
	if deployment.MoltSpec == nil {
		cm.sendDeployAck(ctx, originatorID, deployment.ID, false, "molt spec missing")
		return fmt.Errorf("molt spec missing for deployment %s", deployment.ID)
	}

	// Fetch WASM bytes from IPFS (would need distribution layer — for now, require inline bytes)
	// TODO: integrate with IPFS distribution to pull WASM module by CID
	logging.Info("deploying molt replica",
		logging.ContainerID(deployment.ID),
		"module_cid", deployment.MoltSpec.ModuleCID)

	// Deploy via MoltManager (it handles compilation and handler creation)
	_, err := cm.moltManager.Deploy(ctx, deployment.ID, nil, deployment.MoltSpec, deployment.Owner)
	if err != nil {
		cm.sendDeployAck(ctx, originatorID, deployment.ID, false, err.Error())
		return fmt.Errorf("failed to deploy molt replica: %w", err)
	}

	cm.mu.Lock()
	deployment.Status = types.ContainerStatusRunning
	deployment.StartedAt = time.Now()
	cm.mu.Unlock()

	logging.Info("molt replica deployed successfully",
		logging.ContainerID(deployment.ID))

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
		case ack, ok := <-pending.ackChan:
			if !ok {
				// Channel closed — return current count
				pending.mu.Lock()
				count := pending.successCount
				pending.mu.Unlock()
				return count, nil
			}
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
