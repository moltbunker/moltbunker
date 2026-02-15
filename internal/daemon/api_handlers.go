package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	goruntime "runtime"
	"time"

	"github.com/moltbunker/moltbunker/pkg/types"
)

// handleStatus handles status requests
func (s *APIServer) handleStatus(ctx context.Context, req *APIRequest) *APIResponse {
	peers := s.node.router.GetPeers()
	uptime := time.Since(s.startTime).Round(time.Second)

	torRunning, torAddress := s.containerManager.GetTorStatus()

	deployments := s.containerManager.ListDeployments()
	containerCount := len(deployments)

	// Count encrypted containers
	encryptedCount := 0
	for _, d := range deployments {
		if d.Encrypted {
			encryptedCount++
		}
	}

	peerCount := len(peers)

	// SEV-SNP from auto-detected hardware profile (config layer)
	sevSupported := false
	sevActive := false
	if s.config != nil {
		sevSupported = s.config.Node.Provider.Hardware.SEVSNPSupported
		sevActive = s.config.Node.Provider.Hardware.SEVSNPLevel == "snp"
	}
	sec := &SecurityStatus{
		TLSVersion:          "1.3",
		EncryptionAlgo:      "AES-256-GCM",
		SEVSNPSupported:     sevSupported,
		SEVSNPActive:        sevActive,
		SeccompEnabled:      true,
		TorEnabled:          torRunning,
		CertPinnedPeers:     peerCount,
		EncryptedContainers: encryptedCount,
		TotalContainers:     containerCount,
	}

	// Use profile manager for capacity, tier, reputation, known nodes
	var capacity *AggregatedCapacity
	var knownNodes []NodeProfile
	nodeTier := "starter"
	nodeRole := "hybrid"
	reputation := 0

	if s.profileManager != nil {
		s.profileManager.RefreshSelf()
		s.profileManager.RefreshPeers()
		capacity = s.profileManager.GetAggregatedCapacity()
		knownNodes = s.profileManager.GetAll()
		if self := s.profileManager.GetSelf(); self != nil {
			nodeTier = self.Tier
			nodeRole = self.Role
			reputation = self.ReputationScore
		}
	} else if s.config != nil {
		// Fallback if profile manager not initialized
		nodeTier = string(s.config.Node.Provider.TargetTier)
		nodeRole = string(s.config.Node.Role)
		capacity = &AggregatedCapacity{
			CPUTotal:       s.config.Node.Provider.DeclaredCPU,
			MemoryTotalGB:  s.config.Node.Provider.DeclaredMemoryGB,
			StorageTotalGB: s.config.Node.Provider.DeclaredStorageGB,
			OnlineNodes:    1,
			TotalNodes:     1,
		}
	}

	// Merge admin metadata (badges, blocked) into known nodes
	if s.adminBadgeGetter != nil && knownNodes != nil {
		for i := range knownNodes {
			if meta := s.adminBadgeGetter.Get(knownNodes[i].NodeID); meta != nil {
				knownNodes[i].Badges = meta.Badges
				knownNodes[i].Blocked = meta.Blocked
			}
		}
	}

	loc := s.node.nodeInfo.Location
	status := StatusResponse{
		NodeID:          s.node.nodeInfo.ID.String(),
		Running:         s.node.IsRunning(),
		Port:            s.node.nodeInfo.Port,
		NetworkNodes:    peerCount + 1,
		Uptime:          uptime.String(),
		Version:         "0.1.0",
		TorEnabled:      torRunning,
		TorAddress:      torAddress,
		Containers:      containerCount,
		Region:          s.node.nodeInfo.Region,
		Location:        &loc,
		NetworkCapacity: capacity,
		Security:        sec,
		NodeTier:        nodeTier,
		NodeRole:        nodeRole,
		ReputationScore: reputation,
		KnownNodes:      knownNodes,
	}

	return &APIResponse{
		Result: status,
		ID:     req.ID,
	}
}

// handleDeploy handles deployment requests
func (s *APIServer) handleDeploy(ctx context.Context, req *APIRequest) *APIResponse {
	var deployReq DeployRequest
	if err := json.Unmarshal(req.Params, &deployReq); err != nil {
		return &APIResponse{
			Error: fmt.Sprintf("invalid deploy params: %v", err),
			ID:    req.ID,
		}
	}

	// Validate the deployment request
	if err := validateDeployRequest(&deployReq); err != nil {
		return &APIResponse{
			Error: fmt.Sprintf("validation failed: %v", err),
			ID:    req.ID,
		}
	}

	// Deploy via container manager
	result, err := s.containerManager.Deploy(ctx, &deployReq)
	if err != nil {
		return &APIResponse{
			Error: fmt.Sprintf("deployment failed: %v", err),
			ID:    req.ID,
		}
	}

	response := DeployResponse{
		ContainerID:     result.Deployment.ID,
		Status:          string(result.Deployment.Status),
		OnionAddress:    result.Deployment.OnionAddress,
		EncryptedVolume: result.Deployment.EncryptedVolume,
		Regions:         result.Deployment.Regions,
		Locations:       result.Deployment.Locations,
		ReplicaCount:    result.ReplicaCount,
	}

	return &APIResponse{
		Result: response,
		ID:     req.ID,
	}
}

// handleStop handles stop requests
func (s *APIServer) handleStop(ctx context.Context, req *APIRequest) *APIResponse {
	var params struct {
		ContainerID string `json:"container_id"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return &APIResponse{
			Error: fmt.Sprintf("invalid params: %v", err),
			ID:    req.ID,
		}
	}

	// Validate container ID
	if err := validateContainerID(params.ContainerID); err != nil {
		return &APIResponse{
			Error: fmt.Sprintf("validation failed: %v", err),
			ID:    req.ID,
		}
	}

	if err := s.containerManager.Stop(ctx, params.ContainerID); err != nil {
		return &APIResponse{
			Error: fmt.Sprintf("failed to stop container: %v", err),
			ID:    req.ID,
		}
	}

	return &APIResponse{
		Result: map[string]interface{}{
			"status":       "stopped",
			"container_id": params.ContainerID,
		},
		ID: req.ID,
	}
}

// handleStart handles start requests (restart a stopped container)
func (s *APIServer) handleStart(ctx context.Context, req *APIRequest) *APIResponse {
	var params struct {
		ContainerID string `json:"container_id"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return &APIResponse{
			Error: fmt.Sprintf("invalid params: %v", err),
			ID:    req.ID,
		}
	}

	// Validate container ID
	if err := validateContainerID(params.ContainerID); err != nil {
		return &APIResponse{
			Error: fmt.Sprintf("validation failed: %v", err),
			ID:    req.ID,
		}
	}

	if err := s.containerManager.Start(ctx, params.ContainerID); err != nil {
		return &APIResponse{
			Error: fmt.Sprintf("failed to start container: %v", err),
			ID:    req.ID,
		}
	}

	return &APIResponse{
		Result: map[string]interface{}{
			"status":       "started",
			"container_id": params.ContainerID,
		},
		ID: req.ID,
	}
}

// handleDelete handles delete requests
func (s *APIServer) handleDelete(ctx context.Context, req *APIRequest) *APIResponse {
	var params struct {
		ContainerID string `json:"container_id"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return &APIResponse{
			Error: fmt.Sprintf("invalid params: %v", err),
			ID:    req.ID,
		}
	}

	// Validate container ID
	if err := validateContainerID(params.ContainerID); err != nil {
		return &APIResponse{
			Error: fmt.Sprintf("validation failed: %v", err),
			ID:    req.ID,
		}
	}

	if err := s.containerManager.Delete(ctx, params.ContainerID); err != nil {
		return &APIResponse{
			Error: fmt.Sprintf("failed to delete container: %v", err),
			ID:    req.ID,
		}
	}

	return &APIResponse{
		Result: map[string]interface{}{
			"status":       "deleted",
			"container_id": params.ContainerID,
		},
		ID: req.ID,
	}
}

// handleLogs handles log streaming requests
func (s *APIServer) handleLogs(ctx context.Context, req *APIRequest) *APIResponse {
	var logsReq LogsRequest
	if err := json.Unmarshal(req.Params, &logsReq); err != nil {
		return &APIResponse{
			Error: fmt.Sprintf("invalid logs params: %v", err),
			ID:    req.ID,
		}
	}

	// Validate container ID
	if err := validateContainerID(logsReq.ContainerID); err != nil {
		return &APIResponse{
			Error: fmt.Sprintf("validation failed: %v", err),
			ID:    req.ID,
		}
	}

	// Validate Tail parameter: must be >= 0 and <= MaxLogTailLines
	if logsReq.Tail < 0 {
		return &APIResponse{
			Error: fmt.Sprintf("%v: tail cannot be negative", ErrInvalidTailValue),
			ID:    req.ID,
		}
	}
	if logsReq.Tail > MaxLogTailLines {
		return &APIResponse{
			Error: fmt.Sprintf("%v: tail exceeds maximum of %d lines", ErrInvalidTailValue, MaxLogTailLines),
			ID:    req.ID,
		}
	}

	// Get logs from container manager
	reader, err := s.containerManager.GetLogs(ctx, logsReq.ContainerID, logsReq.Follow, logsReq.Tail)
	if err != nil {
		return &APIResponse{
			Error: fmt.Sprintf("failed to get logs: %v", err),
			ID:    req.ID,
		}
	}
	defer reader.Close()

	// Read logs
	logs, err := io.ReadAll(reader)
	if err != nil {
		return &APIResponse{
			Error: fmt.Sprintf("failed to read logs: %v", err),
			ID:    req.ID,
		}
	}

	return &APIResponse{
		Result: map[string]interface{}{
			"container_id": logsReq.ContainerID,
			"logs":         string(logs),
		},
		ID: req.ID,
	}
}

// handleList handles list deployments requests
func (s *APIServer) handleList(ctx context.Context, req *APIRequest) *APIResponse {
	deployments := s.containerManager.ListDeployments()

	containers := make([]ContainerInfo, 0, len(deployments))
	for _, d := range deployments {
		hasVolume := d.EncryptedVolume != "" &&
			(d.Status != types.ContainerStatusStopped ||
				d.VolumeExpiresAt.IsZero() ||
				time.Now().Before(d.VolumeExpiresAt))
		containers = append(containers, ContainerInfo{
			ID:              d.ID,
			Image:           d.Image,
			Status:          string(d.Status),
			CreatedAt:       d.CreatedAt,
			StartedAt:       d.StartedAt,
			Encrypted:       d.Encrypted,
			OnionAddress:    d.OnionAddress,
			Regions:         d.Regions,
			Locations:       d.Locations,
			Owner:           d.Owner,
			StoppedAt:       d.StoppedAt,
			VolumeExpiresAt: d.VolumeExpiresAt,
			HasVolume:       hasVolume,
		})
	}

	return &APIResponse{
		Result: containers,
		ID:     req.ID,
	}
}

// handleTorStart handles Tor start requests
func (s *APIServer) handleTorStart(ctx context.Context, req *APIRequest) *APIResponse {
	if err := s.containerManager.StartTor(ctx); err != nil {
		return &APIResponse{
			Error: fmt.Sprintf("failed to start Tor: %v", err),
			ID:    req.ID,
		}
	}

	_, address := s.containerManager.GetTorStatus()

	return &APIResponse{
		Result: map[string]interface{}{
			"status":  "started",
			"address": address,
		},
		ID: req.ID,
	}
}

// handleTorStop handles Tor stop requests
func (s *APIServer) handleTorStop(ctx context.Context, req *APIRequest) *APIResponse {
	if err := s.containerManager.StopTor(); err != nil {
		return &APIResponse{
			Error: fmt.Sprintf("failed to stop Tor: %v", err),
			ID:    req.ID,
		}
	}

	return &APIResponse{
		Result: map[string]interface{}{
			"status": "stopped",
		},
		ID: req.ID,
	}
}

// handleTorStatus handles Tor status requests
func (s *APIServer) handleTorStatus(ctx context.Context, req *APIRequest) *APIResponse {
	running, address := s.containerManager.GetTorStatus()

	status := TorStatusResponse{
		Running:      running,
		OnionAddress: address,
		CircuitCount: -1, // -1 indicates circuit count not available
	}

	if running {
		status.StartedAt = time.Now() // Would need to track actual start time
	}

	return &APIResponse{
		Result: status,
		ID:     req.ID,
	}
}

// handleTorRotate handles Tor circuit rotation requests
func (s *APIServer) handleTorRotate(ctx context.Context, req *APIRequest) *APIResponse {
	if err := s.containerManager.RotateTorCircuit(ctx); err != nil {
		return &APIResponse{
			Error: fmt.Sprintf("failed to rotate circuit: %v", err),
			ID:    req.ID,
		}
	}

	return &APIResponse{
		Result: map[string]interface{}{
			"status": "rotated",
		},
		ID: req.ID,
	}
}

// handlePeers handles peer list requests
func (s *APIServer) handlePeers(ctx context.Context, req *APIRequest) *APIResponse {
	peers := s.node.router.GetPeers()

	peerList := make([]map[string]interface{}, 0, len(peers))
	for _, peer := range peers {
		peerList = append(peerList, map[string]interface{}{
			"id":        peer.ID.String(),
			"address":   peer.Address,
			"region":    peer.Region,
			"country":   peer.Country,
			"location":  peer.Location,
			"last_seen": peer.LastSeen,
		})
	}

	return &APIResponse{
		Result: peerList,
		ID:     req.ID,
	}
}

// handleHealth handles health check requests
func (s *APIServer) handleHealth(ctx context.Context, req *APIRequest) *APIResponse {
	var params struct {
		ContainerID string `json:"container_id"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		// Return overall health if no container specified
		unhealthy := s.containerManager.GetUnhealthyDeployments()
		return &APIResponse{
			Result: map[string]interface{}{
				"healthy":              len(unhealthy) == 0,
				"unhealthy_containers": unhealthy,
			},
			ID: req.ID,
		}
	}

	// Validate container ID if specified
	if params.ContainerID != "" {
		if err := validateContainerID(params.ContainerID); err != nil {
			return &APIResponse{
				Error: fmt.Sprintf("validation failed: %v", err),
				ID:    req.ID,
			}
		}
	}

	health, err := s.containerManager.GetHealth(ctx, params.ContainerID)
	if err != nil {
		return &APIResponse{
			Error: fmt.Sprintf("failed to get health: %v", err),
			ID:    req.ID,
		}
	}

	return &APIResponse{
		Result: health,
		ID:     req.ID,
	}
}

// handleConfigGet handles config get requests
func (s *APIServer) handleConfigGet(ctx context.Context, req *APIRequest) *APIResponse {
	return &APIResponse{
		Result: map[string]interface{}{
			"port":        s.node.nodeInfo.Port,
			"node_id":     s.node.nodeInfo.ID.String(),
			"data_dir":    s.dataDir,
			"socket_path": s.socketPath,
			"region":      s.node.nodeInfo.Region,
			"country":     s.node.nodeInfo.Country,
			"location":    s.node.nodeInfo.Location,
		},
		ID: req.ID,
	}
}

// handleConfigSet handles config set requests
func (s *APIServer) handleConfigSet(ctx context.Context, req *APIRequest) *APIResponse {
	// Runtime config changes require editing config file and restarting daemon
	// This is by design - configuration should be persistent
	return &APIResponse{
		Result: map[string]interface{}{
			"status":  "requires_restart",
			"message": "Edit ~/.moltbunker/config.yaml and restart the daemon to apply changes",
		},
		ID: req.ID,
	}
}

// handleHealthz handles detailed health check requests for liveness probes
func (s *APIServer) handleHealthz(ctx context.Context, req *APIRequest) *APIResponse {
	// Get memory stats
	var memStats goruntime.MemStats
	goruntime.ReadMemStats(&memStats)

	// Check node running status
	nodeRunning := s.node != nil && s.node.IsRunning()

	// Check containerd connection status
	containerdConnected := s.containerManager != nil && s.containerManager.IsContainerdConnected()

	// Get peer count
	peerCount := 0
	if s.node != nil && s.node.router != nil {
		peerCount = len(s.node.router.GetPeers())
	}

	// Determine overall health status
	status := "healthy"
	if !nodeRunning {
		status = "unhealthy"
	} else if !containerdConnected {
		status = "degraded"
	}

	healthz := HealthzResponse{
		Status:              status,
		NodeRunning:         nodeRunning,
		ContainerdConnected: containerdConnected,
		PeerCount:           peerCount,
		GoroutineCount:      goruntime.NumGoroutine(),
		MemoryUsageMB:       float64(memStats.Sys) / (1024 * 1024),
		MemoryAllocMB:       float64(memStats.Alloc) / (1024 * 1024),
		Timestamp:           time.Now(),
	}

	return &APIResponse{
		Result: healthz,
		ID:     req.ID,
	}
}

// handleReadyz handles readiness probe requests
func (s *APIServer) handleReadyz(ctx context.Context, req *APIRequest) *APIResponse {
	s.mu.RLock()
	running := s.running
	s.mu.RUnlock()

	// Check if the server is running and ready to accept requests
	ready := running && s.node != nil && s.node.IsRunning()

	var message string
	if !running {
		message = "API server not running"
	} else if s.node == nil {
		message = "Node not initialized"
	} else if !s.node.IsRunning() {
		message = "Node not running"
	}

	readyz := ReadyzResponse{
		Ready:     ready,
		Message:   message,
		Timestamp: time.Now(),
	}

	return &APIResponse{
		Result: readyz,
		ID:     req.ID,
	}
}

// handleMetrics handles metrics endpoint requests
func (s *APIServer) handleMetrics(ctx context.Context, req *APIRequest) *APIResponse {
	metricsData := s.metrics.GetMetrics()

	return &APIResponse{
		Result: metricsData,
		ID:     req.ID,
	}
}

// handleContainerDetail returns detailed container info including provider node location.
func (s *APIServer) handleContainerDetail(ctx context.Context, req *APIRequest) *APIResponse {
	var params struct {
		ContainerID string `json:"container_id"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return &APIResponse{
			Error: fmt.Sprintf("invalid params: %v", err),
			ID:    req.ID,
		}
	}

	if err := validateContainerID(params.ContainerID); err != nil {
		return &APIResponse{
			Error: fmt.Sprintf("validation failed: %v", err),
			ID:    req.ID,
		}
	}

	deployment, exists := s.containerManager.GetDeployment(params.ContainerID)
	if !exists {
		return &APIResponse{
			Error: fmt.Sprintf("container not found: %s", params.ContainerID),
			ID:    req.ID,
		}
	}

	// Resolve provider node address
	providerNodeID := ""
	providerAddr := ""
	if peerID, ok := s.containerManager.GetContainerProviderNode(params.ContainerID); ok {
		providerNodeID = peerID.String()
		// Look up address from peer list
		for _, peer := range s.node.router.GetPeers() {
			if peer.ID == peerID {
				providerAddr = peer.Address
				break
			}
		}
		// If provider is us, use our address
		if peerID == s.node.nodeInfo.ID {
			providerAddr = fmt.Sprintf("127.0.0.1:%d", s.node.nodeInfo.Port)
		}
	}

	detail := map[string]interface{}{
		"id":               deployment.ID,
		"image":            deployment.Image,
		"status":           string(deployment.Status),
		"provider_node_id": providerNodeID,
		"provider_address": providerAddr,
		"owner":            deployment.Owner,
	}

	return &APIResponse{
		Result: detail,
		ID:     req.ID,
	}
}

// sendError sends an error response and returns any encoding error
func (s *APIServer) sendError(encoder *json.Encoder, id int, message string) error {
	if err := encoder.Encode(&APIResponse{
		Error: message,
		ID:    id,
	}); err != nil {
		return fmt.Errorf("failed to send error response: %w", err)
	}
	return nil
}
