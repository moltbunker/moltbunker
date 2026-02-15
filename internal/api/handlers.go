package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/moltbunker/moltbunker/internal/client"
	"github.com/moltbunker/moltbunker/internal/cloning"
	"github.com/moltbunker/moltbunker/internal/logging"
	"github.com/moltbunker/moltbunker/internal/snapshot"
	"github.com/moltbunker/moltbunker/pkg/types"
)

// Request/Response types matching the website API spec

// ReserveRequest matches POST /reserve
type ReserveRequest struct {
	Tier         string `json:"tier"`          // minimal, standard, performance, enterprise
	DurationHours int   `json:"duration_hours"`
	Region       string `json:"region,omitempty"`
	TorOnly      bool   `json:"tor_only,omitempty"`
}

// ReserveResponse is the response for POST /reserve
type ReserveResponse struct {
	RuntimeID    string    `json:"runtime_id"`
	Status       string    `json:"status"`
	Tier         string    `json:"tier"`
	Region       string    `json:"region"`
	ExpiresAt    time.Time `json:"expires_at"`
	OnionAddress string    `json:"onion_address,omitempty"`
	Cost         string    `json:"cost"`
}

// DeployRequest matches POST /deploy
type DeployRequest struct {
	RuntimeID       string               `json:"runtime_id,omitempty"`
	Image           string               `json:"image"`
	CodeHash        string               `json:"code_hash,omitempty"`
	Resources       types.ResourceLimits `json:"resources,omitempty"`
	TorOnly         bool                 `json:"tor_only"`
	OnionService    bool                 `json:"onion_service"`
	ReservationID   string               `json:"reservation_id,omitempty"`
	MinProviderTier string               `json:"min_provider_tier,omitempty"`
}

// DeployResponse is the response for POST /deploy
type DeployResponse struct {
	ContainerID  string    `json:"container_id"`
	Status       string    `json:"status"`
	OnionAddress string    `json:"onion_address,omitempty"`
	Regions      []string  `json:"regions"`
	ReplicaCount int       `json:"replica_count"`
	CreatedAt    time.Time `json:"created_at"`
}

// CloneRequest matches POST /clone
type CloneRequest struct {
	CodeHash      string `json:"code_hash,omitempty"`
	StateSnapshot string `json:"state_snapshot,omitempty"` // Snapshot ID
	TargetRegion  string `json:"target_region,omitempty"`
	Reason        string `json:"reason,omitempty"`
}

// CloneResponse is the response for POST /clone
type CloneResponse struct {
	CloneID      string    `json:"clone_id"`
	SourceID     string    `json:"source_id"`
	TargetID     string    `json:"target_id,omitempty"`
	Status       string    `json:"status"`
	TargetRegion string    `json:"target_region"`
	CreatedAt    time.Time `json:"created_at"`
}

// MigrateRequest matches POST /migrate
type MigrateRequest struct {
	ContainerID  string `json:"container_id"`
	TargetRegion string `json:"target_region"`
	KeepOriginal bool   `json:"keep_original,omitempty"`
}

// MigrateResponse is the response for POST /migrate
type MigrateResponse struct {
	MigrationID  string    `json:"migration_id"`
	Status       string    `json:"status"`
	SourceRegion string    `json:"source_region"`
	TargetRegion string    `json:"target_region"`
	StartedAt    time.Time `json:"started_at"`
}

// BalanceResponse matches GET /balance
type BalanceResponse struct {
	WalletAddress string `json:"wallet_address"`
	BUNKERBalance string `json:"bunker_balance"`
	ETHBalance    string `json:"eth_balance"`
	Deposited     string `json:"deposited"`
	Reserved      string `json:"reserved"`
	Available     string `json:"available"`
}

// SnapshotRequest matches POST /snapshot
type SnapshotRequest struct {
	ContainerID string            `json:"container_id"`
	Type        string            `json:"type,omitempty"` // full, incremental, checkpoint
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// SnapshotResponse is the response for POST /snapshot
type SnapshotResponse struct {
	SnapshotID  string    `json:"snapshot_id"`
	ContainerID string    `json:"container_id"`
	Type        string    `json:"type"`
	Size        int64     `json:"size"`
	Checksum    string    `json:"checksum"`
	CreatedAt   time.Time `json:"created_at"`
}

// RestoreRequest matches POST /restore
type RestoreRequest struct {
	SnapshotID   string `json:"snapshot_id"`
	TargetRegion string `json:"target_region,omitempty"`
	NewContainer bool   `json:"new_container,omitempty"`
}

// RestoreResponse is the response for POST /restore
type RestoreResponse struct {
	ContainerID string    `json:"container_id"`
	SnapshotID  string    `json:"snapshot_id"`
	Status      string    `json:"status"`
	RestoredAt  time.Time `json:"restored_at"`
}

// ThreatResponse matches GET /threat
type ThreatResponse struct {
	Score          float64         `json:"score"`
	Level          string          `json:"level"`
	ActiveSignals  []SignalInfo    `json:"active_signals"`
	Recommendation string          `json:"recommendation"`
	Timestamp      time.Time       `json:"timestamp"`
}

// SignalInfo is a simplified signal for API response
type SignalInfo struct {
	Type       string  `json:"type"`
	Score      float64 `json:"score"`
	Confidence float64 `json:"confidence"`
	Source     string  `json:"source"`
}

// AggregatedCapacity contains aggregated node resource capacity
type AggregatedCapacity struct {
	CPUTotal       int     `json:"cpu_total"`
	MemoryTotalGB  int     `json:"memory_total_gb"`
	StorageTotalGB int     `json:"storage_total_gb"`

	CPUUsed       int     `json:"cpu_used"`
	MemoryUsedGB  float64 `json:"memory_used_gb"`
	StorageUsedGB float64 `json:"storage_used_gb"`

	OnlineNodes int `json:"online_nodes"`
	TotalNodes  int `json:"total_nodes"`
}

// NetworkCapacity is an alias for backward compatibility
type NetworkCapacity = AggregatedCapacity

// SecurityStatus contains security feature status
type SecurityStatus struct {
	TLSVersion          string `json:"tls_version"`
	EncryptionAlgo      string `json:"encryption_algo"`
	SEVSNPSupported     bool   `json:"sev_snp_supported"`
	SEVSNPActive        bool   `json:"sev_snp_active"`
	SeccompEnabled      bool   `json:"seccomp_enabled"`
	TorEnabled          bool   `json:"tor_enabled"`
	CertPinnedPeers     int    `json:"cert_pinned_peers"`
	EncryptedContainers int    `json:"encrypted_containers"`
	TotalContainers     int    `json:"total_containers"`
}

// NodeProfile represents a known node in the network
type NodeProfile struct {
	NodeID           string    `json:"node_id"`
	Address          string    `json:"address,omitempty"`
	WalletAddress    string    `json:"wallet_address,omitempty"`
	Region           string    `json:"region"`
	Country          string    `json:"country,omitempty"`
	Online           bool      `json:"online"`
	LastSeen         time.Time `json:"last_seen"`
	Capacity         CapacityProfile `json:"capacity"`
	Tier             string    `json:"tier"`
	Role             string    `json:"role"`
	ReputationScore  int       `json:"reputation_score"`
	StakingAmount    uint64    `json:"staking_amount"`
	ActiveContainers int       `json:"active_containers"`
	EncryptedCount   int       `json:"encrypted_containers"`
	Version          string    `json:"version,omitempty"`

	// Admin-assigned metadata
	Badges  []string `json:"badges,omitempty"`
	Blocked bool     `json:"blocked,omitempty"`
}

// HardwareProfile contains detailed hardware information for a node
type HardwareProfile struct {
	CPUModel         string `json:"cpu_model"`
	CPUArch          string `json:"cpu_arch"`
	CPUThreads       int    `json:"cpu_threads"`
	CPUCores         int    `json:"cpu_cores"`
	CPUSockets       int    `json:"cpu_sockets"`
	MemoryGB         int    `json:"memory_gb"`
	MemoryType       string `json:"memory_type"`
	MemoryECC        bool   `json:"memory_ecc"`
	StorageGB        int    `json:"storage_gb"`
	StorageType      string `json:"storage_type"`
	StorageModel     string `json:"storage_model"`
	BandwidthMbps    int    `json:"bandwidth_mbps"`
	NetworkInterface string `json:"network_interface,omitempty"`
	SEVSNPSupported  bool   `json:"sev_snp_supported"`
	SEVSNPLevel      string `json:"sev_snp_level"`
	TPMVersion       string `json:"tpm_version"`
	OS               string `json:"os"`
	OSVersion        string `json:"os_version"`
	Kernel           string `json:"kernel"`
	Hostname         string `json:"hostname"`
}

// CapacityProfile describes a node's declared resource capacity
type CapacityProfile struct {
	CPUCores      int              `json:"cpu_cores"`
	MemoryGB      int              `json:"memory_gb"`
	StorageGB     int              `json:"storage_gb"`
	BandwidthMbps int              `json:"bandwidth_mbps"`
	GPUCount      int              `json:"gpu_count,omitempty"`
	GPUModel      string           `json:"gpu_model,omitempty"`
	Hardware      *HardwareProfile `json:"hardware,omitempty"`
}

// StatusResponse matches GET /status
type StatusResponse struct {
	NodeID       string    `json:"node_id"`
	Running      bool      `json:"running"`
	Version      string    `json:"version"`
	Uptime       string    `json:"uptime"`
	NetworkNodes int       `json:"network_nodes"`
	Containers   int       `json:"containers"`
	TorEnabled   bool      `json:"tor_enabled"`
	TorAddress   string    `json:"tor_address,omitempty"`
	Region       string    `json:"region"`
	ThreatLevel  float64   `json:"threat_level"`
	Timestamp    time.Time `json:"timestamp"`

	// Extended fields
	NetworkCapacity *AggregatedCapacity `json:"network_capacity,omitempty"`
	Security        *SecurityStatus     `json:"security,omitempty"`
	NodeTier        string              `json:"node_tier,omitempty"`
	NodeRole        string              `json:"node_role,omitempty"`
	ReputationScore int                 `json:"reputation_score"`
	KnownNodes      []NodeProfile       `json:"known_nodes,omitempty"`
}

// Handler implementations

// handleStatus handles GET /v1/status
func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	// Try to get status from daemon via bridge
	if s.daemonBridge != nil {
		status, err := s.daemonBridge.Status()
		if err != nil {
			logging.Warn("failed to get daemon status via bridge",
				"error", err.Error(),
				logging.Component("api"))
		} else {
			// Map client capacity/security structs to API layer
			var apiCapacity *AggregatedCapacity
			if status.NetworkCapacity != nil {
				apiCapacity = &AggregatedCapacity{
					CPUTotal:       status.NetworkCapacity.CPUTotal,
					MemoryTotalGB:  status.NetworkCapacity.MemoryTotalGB,
					StorageTotalGB: status.NetworkCapacity.StorageTotalGB,
					CPUUsed:        status.NetworkCapacity.CPUUsed,
					MemoryUsedGB:   status.NetworkCapacity.MemoryUsedGB,
					StorageUsedGB:  status.NetworkCapacity.StorageUsedGB,
					OnlineNodes:    status.NetworkCapacity.OnlineNodes,
					TotalNodes:     status.NetworkCapacity.TotalNodes,
				}
			}
			var apiSecurity *SecurityStatus
			if status.Security != nil {
				apiSecurity = &SecurityStatus{
					TLSVersion:          status.Security.TLSVersion,
					EncryptionAlgo:      status.Security.EncryptionAlgo,
					SEVSNPSupported:     status.Security.SEVSNPSupported,
					SEVSNPActive:        status.Security.SEVSNPActive,
					SeccompEnabled:      status.Security.SeccompEnabled,
					TorEnabled:          status.Security.TorEnabled,
					CertPinnedPeers:     status.Security.CertPinnedPeers,
					EncryptedContainers: status.Security.EncryptedContainers,
					TotalContainers:     status.Security.TotalContainers,
				}
			}

			// Map known nodes
			var apiNodes []NodeProfile
			for _, n := range status.KnownNodes {
				cap := CapacityProfile{
					CPUCores:      n.Capacity.CPUCores,
					MemoryGB:      n.Capacity.MemoryGB,
					StorageGB:     n.Capacity.StorageGB,
					BandwidthMbps: n.Capacity.BandwidthMbps,
					GPUCount:      n.Capacity.GPUCount,
					GPUModel:      n.Capacity.GPUModel,
				}
				if n.Capacity.Hardware != nil {
					h := n.Capacity.Hardware
					cap.Hardware = &HardwareProfile{
						CPUModel:         h.CPUModel,
						CPUArch:          h.CPUArch,
						CPUThreads:       h.CPUThreads,
						CPUCores:         h.CPUCores,
						CPUSockets:       h.CPUSockets,
						MemoryGB:         h.MemoryGB,
						MemoryType:       h.MemoryType,
						MemoryECC:        h.MemoryECC,
						StorageGB:        h.StorageGB,
						StorageType:      h.StorageType,
						StorageModel:     h.StorageModel,
						BandwidthMbps:    h.BandwidthMbps,
						NetworkInterface: h.NetworkInterface,
						SEVSNPSupported:  h.SEVSNPSupported,
						SEVSNPLevel:      h.SEVSNPLevel,
						TPMVersion:       h.TPMVersion,
						OS:               h.OS,
						OSVersion:        h.OSVersion,
						Kernel:           h.Kernel,
						Hostname:         h.Hostname,
					}
				}
				apiNodes = append(apiNodes, NodeProfile{
					NodeID:           n.NodeID,
					Address:          n.Address,
					WalletAddress:    n.WalletAddress,
					Region:           n.Region,
					Country:          n.Country,
					Online:           n.Online,
					LastSeen:         n.LastSeen,
					Capacity:         cap,
					Tier:             n.Tier,
					Role:             n.Role,
					ReputationScore:  n.ReputationScore,
					StakingAmount:    n.StakingAmount,
					ActiveContainers: n.ActiveContainers,
					EncryptedCount:   n.EncryptedCount,
					Version:          n.Version,
					Badges:           n.Badges,
					Blocked:          n.Blocked,
				})
			}

			// Merge admin metadata (badges, blocked) into nodes
			if s.adminStore != nil {
				for i := range apiNodes {
					if meta := s.adminStore.Get(apiNodes[i].NodeID); meta != nil {
						apiNodes[i].Badges = meta.Badges
						apiNodes[i].Blocked = meta.Blocked
					}
				}
			}

			response := StatusResponse{
				NodeID:          status.NodeID,
				Running:         status.Running,
				Version:         status.Version,
				Uptime:          status.Uptime,
				NetworkNodes:    status.NetworkNodes,
				Containers:      status.Containers,
				TorEnabled:      status.TorEnabled,
				TorAddress:      status.TorAddress,
				Region:          status.Region,
				ThreatLevel:     status.ThreatLevel,
				Timestamp:       time.Now(),
				NetworkCapacity: apiCapacity,
				Security:        apiSecurity,
				NodeTier:        status.NodeTier,
				NodeRole:        status.NodeRole,
				ReputationScore: status.ReputationScore,
				KnownNodes:      apiNodes,
			}
			s.writeJSON(w, http.StatusOK, response)
			return
		}
	}

	// Fallback to local state — use full config if available for declared values
	response := StatusResponse{
		NodeID:       "unknown",
		Running:      true,
		Version:      "0.1.0",
		Uptime:       "0s",
		NetworkNodes: 1, // at least self
		Containers:   0,
		TorEnabled:   false,
		Region:       "unknown",
		Timestamp:    time.Now(),
	}

	if s.fullConfig != nil {
		cfg := s.fullConfig
		// Region is detected at runtime via geolocation, not in config
		response.TorEnabled = cfg.Tor.Enabled
		response.NodeTier = string(cfg.Node.Provider.TargetTier)
		response.NodeRole = string(cfg.Node.Role)
		response.ReputationScore = 1000 // default initial (matches on-chain max)
		response.NetworkCapacity = &AggregatedCapacity{
			CPUTotal:       cfg.Node.Provider.DeclaredCPU,
			MemoryTotalGB:  cfg.Node.Provider.DeclaredMemoryGB,
			StorageTotalGB: cfg.Node.Provider.DeclaredStorageGB,
			OnlineNodes:    1,
			TotalNodes:     1,
		}
		response.Security = &SecurityStatus{
			TLSVersion:          "1.3",
			EncryptionAlgo:      "AES-256-GCM",
			SeccompEnabled:      true,
			TorEnabled:          cfg.Tor.Enabled,
			SEVSNPSupported:     cfg.Node.Provider.Hardware.SEVSNPSupported,
			SEVSNPActive:        cfg.Node.Provider.Hardware.SEVSNPLevel == "snp",
		}
		// Map config hardware to API hardware profile
		cfgHW := cfg.Node.Provider.Hardware
		hw := &HardwareProfile{
			CPUModel:         cfgHW.CPUModel,
			CPUArch:          cfgHW.CPUArch,
			CPUThreads:       cfgHW.CPUThreads,
			CPUCores:         cfgHW.CPUCores,
			CPUSockets:       cfgHW.CPUSockets,
			MemoryGB:         cfgHW.MemoryGB,
			MemoryType:       cfgHW.MemoryType,
			MemoryECC:        cfgHW.MemoryECC,
			StorageGB:        cfgHW.StorageGB,
			StorageType:      cfgHW.StorageType,
			StorageModel:     cfgHW.StorageModel,
			BandwidthMbps:    cfgHW.BandwidthMbps,
			NetworkInterface: cfgHW.NetworkInterface,
			SEVSNPSupported:  cfgHW.SEVSNPSupported,
			SEVSNPLevel:      cfgHW.SEVSNPLevel,
			TPMVersion:       cfgHW.TPMVersion,
			OS:               cfgHW.OS,
			OSVersion:        cfgHW.OSVersion,
			Kernel:           cfgHW.Kernel,
			Hostname:         cfgHW.Hostname,
		}
		// Include self as a known node
		response.KnownNodes = []NodeProfile{
			{
				NodeID:          "self",
				Region:          "unknown",
				Online:          true,
				LastSeen:        time.Now(),
				Tier:            string(cfg.Node.Provider.TargetTier),
				Role:            string(cfg.Node.Role),
				ReputationScore: 1000,
				Version:         "0.1.0",
				Capacity: CapacityProfile{
					CPUCores:      cfg.Node.Provider.DeclaredCPU,
					MemoryGB:      cfg.Node.Provider.DeclaredMemoryGB,
					StorageGB:     cfg.Node.Provider.DeclaredStorageGB,
					BandwidthMbps: cfg.Node.Provider.DeclaredBandwidth,
					Hardware:      hw,
				},
			},
		}
	}

	if s.threatDetector != nil {
		response.ThreatLevel = s.threatDetector.GetScore()
	}

	s.writeJSON(w, http.StatusOK, response)
}

// handleDeploy handles POST /v1/deploy
func (s *Server) handleDeploy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req DeployRequest
	if err := s.readJSON(r, &req); err != nil {
		s.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if req.Image == "" {
		s.writeError(w, http.StatusBadRequest, "image is required")
		return
	}

	// Extract deployer wallet for ownership tracking
	wallet := s.extractWalletAddress(r)

	// Forward to daemon via bridge
	if s.daemonBridge != nil {
		daemonReq := &client.DeployRequest{
			Image:           req.Image,
			CodeHash:        req.CodeHash,
			TorOnly:         req.TorOnly,
			OnionService:    req.OnionService,
			ReservationID:   req.ReservationID,
			Owner:           wallet,
			MinProviderTier: req.MinProviderTier,
		}
		if req.Resources.CPUQuota > 0 || req.Resources.MemoryLimit > 0 {
			daemonReq.Resources = &client.ResourceLimits{
				CPUShares:   req.Resources.CPUQuota,
				MemoryMB:    req.Resources.MemoryLimit / (1024 * 1024), // Convert bytes to MB
				StorageMB:   req.Resources.DiskLimit / (1024 * 1024),   // Convert bytes to MB
				NetworkMbps: int(req.Resources.NetworkBW / 125000),     // Convert bytes/sec to Mbps
			}
		}

		result, err := s.daemonBridge.Deploy(daemonReq)
		if err != nil {
			logging.Warn("deploy via daemon bridge failed",
				"error", err.Error(),
				logging.Component("api"))
			s.writeError(w, http.StatusInternalServerError, "deployment failed: "+err.Error())
			return
		}

		response := DeployResponse{
			ContainerID:  result.ContainerID,
			Status:       result.Status,
			OnionAddress: result.OnionAddress,
			Regions:      result.Regions,
			ReplicaCount: result.ReplicaCount,
			CreatedAt:    result.CreatedAt,
		}
		s.writeJSON(w, http.StatusCreated, response)
		return
	}

	// Daemon bridge not available — cannot deploy
	s.writeError(w, http.StatusServiceUnavailable, "daemon not connected: cannot process deployment")
}

// handleReserve handles POST /v1/reserve
func (s *Server) handleReserve(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req ReserveRequest
	if err := s.readJSON(r, &req); err != nil {
		s.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if req.Tier == "" {
		req.Tier = "standard"
	}
	if req.DurationHours <= 0 {
		req.DurationHours = 24
	}
	if req.Region == "" {
		req.Region = "random"
	}

	// Calculate cost based on tier
	cost := "0.00037" // Default standard tier
	switch req.Tier {
	case "minimal":
		cost = "0.000185"
	case "performance":
		cost = "0.00148"
	case "enterprise":
		cost = "0.00592"
	}

	response := ReserveResponse{
		RuntimeID: fmt.Sprintf("rt-%d", time.Now().UnixNano()),
		Status:    "reserved",
		Tier:      req.Tier,
		Region:    req.Region,
		ExpiresAt: time.Now().Add(time.Duration(req.DurationHours) * time.Hour),
		Cost:      cost,
	}

	s.writeJSON(w, http.StatusCreated, response)
}

// handleClone handles POST /v1/clone
func (s *Server) handleClone(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req CloneRequest
	if err := s.readJSON(r, &req); err != nil {
		s.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if req.TargetRegion == "" {
		req.TargetRegion = "random"
	}
	if req.Reason == "" {
		req.Reason = "manual_clone"
	}

	// Try daemon bridge first
	if s.daemonBridge != nil {
		daemonReq := &client.CloneRequest{
			TargetRegion: req.TargetRegion,
			Priority:     2,
			Reason:       req.Reason,
			IncludeState: req.StateSnapshot != "",
		}

		result, err := s.daemonBridge.Clone(daemonReq)
		if err != nil {
			logging.Warn("clone via daemon bridge failed",
				"error", err.Error(),
				logging.Component("api"))
			s.writeError(w, http.StatusInternalServerError, "clone failed: "+err.Error())
			return
		}

		response := CloneResponse{
			CloneID:      result.CloneID,
			SourceID:     result.SourceID,
			TargetID:     result.TargetID,
			Status:       result.Status,
			TargetRegion: result.TargetRegion,
			CreatedAt:    result.CreatedAt,
		}
		s.writeJSON(w, http.StatusAccepted, response)
		return
	}

	// If cloning manager is available locally, use it
	if s.cloningManager != nil {
		cloneReq := &cloning.CloneRequest{
			TargetRegion: req.TargetRegion,
			Priority:     2,
			Reason:       req.Reason,
			IncludeState: req.StateSnapshot != "",
		}

		clone, err := s.cloningManager.RequestClone(context.Background(), cloneReq)
		if err != nil {
			s.writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		response := CloneResponse{
			CloneID:      clone.ID,
			SourceID:     clone.SourceID,
			Status:       string(clone.Status),
			TargetRegion: clone.TargetRegion,
			CreatedAt:    clone.CreatedAt,
		}

		s.writeJSON(w, http.StatusAccepted, response)
		return
	}

	// Daemon bridge not available — cannot clone
	s.writeError(w, http.StatusServiceUnavailable, "daemon not connected: cannot process clone")
}

// handleMigrate handles POST /v1/migrate
func (s *Server) handleMigrate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req MigrateRequest
	if err := s.readJSON(r, &req); err != nil {
		s.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if req.ContainerID == "" {
		s.writeError(w, http.StatusBadRequest, "container_id is required")
		return
	}
	if req.TargetRegion == "" {
		s.writeError(w, http.StatusBadRequest, "target_region is required")
		return
	}

	response := MigrateResponse{
		MigrationID:  fmt.Sprintf("mig-%d", time.Now().UnixNano()),
		Status:       "pending",
		SourceRegion: "unknown",
		TargetRegion: req.TargetRegion,
		StartedAt:    time.Now(),
	}

	s.writeJSON(w, http.StatusAccepted, response)
}

// handleBalance handles GET /v1/balance
func (s *Server) handleBalance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	// Try daemon bridge first
	if s.daemonBridge != nil {
		result, err := s.daemonBridge.RequesterBalance()
		if err != nil {
			logging.Warn("balance via daemon bridge failed",
				"error", err.Error(),
				logging.Component("api"))
		} else {
			response := BalanceResponse{
				WalletAddress: getStringFromMap(result, "wallet_address"),
				BUNKERBalance: getStringFromMap(result, "bunker_balance"),
				ETHBalance:    getStringFromMap(result, "eth_balance"),
				Deposited:     getStringFromMap(result, "deposited"),
				Reserved:      getStringFromMap(result, "reserved"),
				Available:     getStringFromMap(result, "available"),
			}
			s.writeJSON(w, http.StatusOK, response)
			return
		}
	}

	// Fallback response
	response := BalanceResponse{
		WalletAddress: "0x...",
		BUNKERBalance: "0",
		ETHBalance:    "0",
		Deposited:     "0",
		Reserved:      "0",
		Available:     "0",
	}

	s.writeJSON(w, http.StatusOK, response)
}

// getStringFromMap safely extracts a string value from a map
func getStringFromMap(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// firstOrDefault returns the first element of a slice or a default value
func firstOrDefault(slice []string, defaultVal string) string {
	if len(slice) > 0 {
		return slice[0]
	}
	return defaultVal
}

// handleSnapshot handles POST /v1/snapshot
func (s *Server) handleSnapshot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req SnapshotRequest
	if err := s.readJSON(r, &req); err != nil {
		s.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if req.ContainerID == "" {
		s.writeError(w, http.StatusBadRequest, "container_id is required")
		return
	}

	// Try daemon bridge first
	if s.daemonBridge != nil {
		daemonReq := &client.SnapshotRequest{
			ContainerID: req.ContainerID,
			Type:        req.Type,
			Metadata:    req.Metadata,
		}

		result, err := s.daemonBridge.SnapshotCreate(daemonReq)
		if err != nil {
			logging.Warn("snapshot create via daemon bridge failed",
				"error", err.Error(),
				logging.Component("api"))
			s.writeError(w, http.StatusInternalServerError, "snapshot creation failed: "+err.Error())
			return
		}

		response := SnapshotResponse{
			SnapshotID:  result.ID,
			ContainerID: result.ContainerID,
			Type:        result.Type,
			Size:        result.Size,
			Checksum:    result.Checksum,
			CreatedAt:   result.CreatedAt,
		}
		s.writeJSON(w, http.StatusCreated, response)
		return
	}

	snapshotType := snapshot.SnapshotTypeFull
	if req.Type == "incremental" {
		snapshotType = snapshot.SnapshotTypeIncremental
	} else if req.Type == "checkpoint" {
		snapshotType = snapshot.SnapshotTypeCheckpoint
	}

	// If snapshot manager is available locally, use it
	if s.snapshotManager != nil {
		// In a real implementation, get container state data
		stateData := []byte("{}")

		snap, err := s.snapshotManager.CreateSnapshot(req.ContainerID, stateData, snapshotType, req.Metadata)
		if err != nil {
			s.writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		response := SnapshotResponse{
			SnapshotID:  snap.ID,
			ContainerID: snap.ContainerID,
			Type:        string(snap.Type),
			Size:        snap.Size,
			Checksum:    snap.Checksum,
			CreatedAt:   snap.CreatedAt,
		}

		s.writeJSON(w, http.StatusCreated, response)
		return
	}

	// Daemon bridge not available — cannot snapshot
	s.writeError(w, http.StatusServiceUnavailable, "daemon not connected: cannot process snapshot")
}

// handleRestore handles POST /v1/restore
func (s *Server) handleRestore(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req RestoreRequest
	if err := s.readJSON(r, &req); err != nil {
		s.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if req.SnapshotID == "" {
		s.writeError(w, http.StatusBadRequest, "snapshot_id is required")
		return
	}

	// Try daemon bridge first
	if s.daemonBridge != nil {
		result, err := s.daemonBridge.SnapshotRestore(req.SnapshotID, req.TargetRegion, req.NewContainer)
		if err != nil {
			logging.Warn("snapshot restore via daemon bridge failed",
				"error", err.Error(),
				logging.Component("api"))
			s.writeError(w, http.StatusInternalServerError, "restore failed: "+err.Error())
			return
		}

		response := RestoreResponse{
			ContainerID: getStringFromMap(result, "container_id"),
			SnapshotID:  req.SnapshotID,
			Status:      getStringFromMap(result, "status"),
			RestoredAt:  time.Now(),
		}
		s.writeJSON(w, http.StatusAccepted, response)
		return
	}

	// Daemon bridge not available — cannot restore
	s.writeError(w, http.StatusServiceUnavailable, "daemon not connected: cannot process restore")
}

// handleThreat handles GET /v1/threat
func (s *Server) handleThreat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	// Try daemon bridge first
	if s.daemonBridge != nil {
		result, err := s.daemonBridge.ThreatLevel()
		if err != nil {
			logging.Warn("threat level via daemon bridge failed",
				"error", err.Error(),
				logging.Component("api"))
		} else {
			response := ThreatResponse{
				Score:          result.Score,
				Level:          result.Level,
				ActiveSignals:  make([]SignalInfo, 0, len(result.ActiveSignals)),
				Recommendation: result.Recommendation,
				Timestamp:      result.Timestamp,
			}
			for _, sig := range result.ActiveSignals {
				response.ActiveSignals = append(response.ActiveSignals, SignalInfo{
					Type:       sig.Type,
					Score:      sig.Score,
					Confidence: sig.Confidence,
					Source:     sig.Source,
				})
			}
			s.writeJSON(w, http.StatusOK, response)
			return
		}
	}

	// Fallback to local threat detector
	response := ThreatResponse{
		Score:          0.0,
		Level:          "low",
		ActiveSignals:  []SignalInfo{},
		Recommendation: "continue_normal",
		Timestamp:      time.Now(),
	}

	if s.threatDetector != nil {
		level := s.threatDetector.GetThreatLevel()
		response.Score = level.Score
		response.Level = level.Level
		response.Recommendation = level.Recommendation
		response.Timestamp = level.Timestamp

		for _, signal := range level.ActiveSignals {
			response.ActiveSignals = append(response.ActiveSignals, SignalInfo{
				Type:       string(signal.Type),
				Score:      signal.Score,
				Confidence: signal.Confidence,
				Source:     signal.Source,
			})
		}
	}

	s.writeJSON(w, http.StatusOK, response)
}

// Health check handlers (no auth required)

// handleHealth handles GET /v1/health
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "ok"}`))
}

// handleHealthz handles GET /v1/healthz
func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now(),
	})
}

// handleReadyz handles GET /v1/readyz
func (s *Server) handleReadyz(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	running := s.running
	s.mu.RUnlock()

	if !running {
		s.writeJSON(w, http.StatusServiceUnavailable, map[string]interface{}{
			"ready":   false,
			"message": "server not running",
		})
		return
	}

	// Check daemon connectivity
	daemonConnected := false
	if s.daemonBridge != nil {
		daemonConnected = s.daemonBridge.IsConnected()
	}

	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"ready":            true,
		"daemon_connected": daemonConnected,
		"timestamp":        time.Now(),
	})
}

// Helper methods

// readJSON reads JSON from request body
func (s *Server) readJSON(r *http.Request, v interface{}) error {
	body, err := io.ReadAll(io.LimitReader(r.Body, 10*1024*1024)) // 10MB limit
	if err != nil {
		return err
	}
	return json.Unmarshal(body, v)
}

// writeJSON writes JSON response
func (s *Server) writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// writeError writes an error response
func (s *Server) writeError(w http.ResponseWriter, status int, message string) {
	s.writeJSON(w, status, map[string]string{"error": message})
}

// ============================================================================
// Wallet Authentication Handlers (Permissionless)
// ============================================================================

// AuthChallengeRequest is the request for POST /auth/challenge
type AuthChallengeRequest struct {
	Address string `json:"address"`
}

// AuthChallengeResponse is the response for POST /auth/challenge
type AuthChallengeResponse struct {
	Message   string `json:"message"`
	ExpiresIn int    `json:"expires_in"` // seconds
}

// AuthVerifyRequest is the request for POST /auth/verify
type AuthVerifyRequest struct {
	Address   string `json:"address"`
	Message   string `json:"message"`
	Signature string `json:"signature"`
}

// AuthVerifyResponse is the response for POST /auth/verify
type AuthVerifyResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"` // seconds
	Wallet      string `json:"wallet"`
	AuthType    string `json:"auth_type"`
}

// handleAuthChallenge handles POST /v1/auth/challenge
func (s *Server) handleAuthChallenge(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req AuthChallengeRequest
	if err := s.readJSON(r, &req); err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Address == "" {
		s.writeError(w, http.StatusBadRequest, "address is required")
		return
	}

	if s.walletAuth == nil {
		s.writeError(w, http.StatusServiceUnavailable, "wallet authentication not available")
		return
	}

	message, err := s.walletAuth.CreateChallenge(req.Address)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	s.writeJSON(w, http.StatusOK, AuthChallengeResponse{
		Message:   message,
		ExpiresIn: 300, // 5 minutes
	})
}

// handleAuthVerify handles POST /v1/auth/verify
func (s *Server) handleAuthVerify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req AuthVerifyRequest
	if err := s.readJSON(r, &req); err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Address == "" || req.Message == "" || req.Signature == "" {
		s.writeError(w, http.StatusBadRequest, "address, message, and signature are required")
		return
	}

	if s.walletAuth == nil {
		s.writeError(w, http.StatusServiceUnavailable, "wallet authentication not available")
		return
	}

	// Verify signature
	verifiedAddr, err := s.walletAuth.VerifySignature(req.Message, req.Signature, req.Address)
	if err != nil {
		s.writeError(w, http.StatusUnauthorized, err.Error())
		return
	}

	// Create session token
	token, expiresAt, err := s.walletAuth.CreateSession(verifiedAddr)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "failed to create session")
		return
	}

	expiresIn := int(time.Until(expiresAt).Seconds())

	s.writeJSON(w, http.StatusOK, AuthVerifyResponse{
		AccessToken: token,
		ExpiresIn:   expiresIn,
		Wallet:      verifiedAddr,
		AuthType:    "wallet",
	})
}

// ============================================================================
// Bot Handlers
// ============================================================================

// BotRequest is the request for POST /bots
type BotRequest struct {
	Name        string            `json:"name"`
	Image       string            `json:"image"`
	Description string            `json:"description,omitempty"`
	Resources   *ResourceRequest  `json:"resources,omitempty"`
	Region      string            `json:"region,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// ResourceRequest is resource configuration in requests
type ResourceRequest struct {
	CPUShares  int `json:"cpu_shares,omitempty"`
	MemoryMB   int `json:"memory_mb,omitempty"`
	StorageMB  int `json:"storage_mb,omitempty"`
	NetworkMbps int `json:"network_mbps,omitempty"`
}

// BotResponse is the response for bot operations
type BotResponse struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Image       string            `json:"image"`
	Description string            `json:"description,omitempty"`
	Resources   *ResourceRequest  `json:"resources,omitempty"`
	Region      string            `json:"region"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
}

// handleBots handles /v1/bots
func (s *Server) handleBots(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listBots(w, r)
	case http.MethodPost:
		s.createBot(w, r)
	default:
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

func (s *Server) listBots(w http.ResponseWriter, r *http.Request) {
	if s.daemonBridge != nil && s.daemonBridge.IsConnected() {
		containers, err := s.daemonBridge.List()
		if err != nil {
			logging.Warn("list bots via daemon bridge failed",
				"error", err.Error(),
				logging.Component("api"))
		} else {
			bots := make([]BotResponse, 0, len(containers))
			for _, c := range containers {
				bots = append(bots, BotResponse{
					ID:        c.ID,
					Name:      c.ID, // Use container ID as name fallback
					Image:     c.Image,
					Region:    firstOrDefault(c.Regions, "unknown"),
					CreatedAt: c.CreatedAt,
				})
			}
			s.writeJSON(w, http.StatusOK, map[string]interface{}{
				"bots": bots,
			})
			return
		}
	}

	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"bots": []BotResponse{},
	})
}

func (s *Server) createBot(w http.ResponseWriter, r *http.Request) {
	var req BotRequest
	if err := s.readJSON(r, &req); err != nil {
		s.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if req.Name == "" {
		s.writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	if req.Image == "" {
		s.writeError(w, http.StatusBadRequest, "image is required")
		return
	}

	// Create bot (forward to daemon or create locally)
	bot := BotResponse{
		ID:          fmt.Sprintf("bot-%d", time.Now().UnixNano()),
		Name:        req.Name,
		Image:       req.Image,
		Description: req.Description,
		Resources:   req.Resources,
		Region:      req.Region,
		Metadata:    req.Metadata,
		CreatedAt:   time.Now(),
	}

	s.writeJSON(w, http.StatusCreated, bot)
}

// handleBotByID handles /v1/bots/{id}
func (s *Server) handleBotByID(w http.ResponseWriter, r *http.Request) {
	// Extract bot ID from path
	path := r.URL.Path
	parts := strings.Split(strings.TrimPrefix(path, "/v1/bots/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		s.writeError(w, http.StatusBadRequest, "bot_id required")
		return
	}
	botID := parts[0]

	// Check for sub-resources
	if len(parts) > 1 {
		switch parts[1] {
		case "cloning":
			s.handleBotCloning(w, r, botID)
			return
		case "status":
			s.handleBotStatus(w, r, botID)
			return
		case "clones":
			if len(parts) > 2 && parts[2] == "sync" {
				s.handleBotSyncClones(w, r, botID)
				return
			}
			s.handleBotClones(w, r, botID)
			return
		}
	}

	switch r.Method {
	case http.MethodGet:
		s.getBot(w, r, botID)
	case http.MethodPatch:
		s.updateBot(w, r, botID)
	case http.MethodDelete:
		s.deleteBot(w, r, botID)
	default:
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

func (s *Server) getBot(w http.ResponseWriter, r *http.Request, botID string) {
	if s.daemonBridge != nil && s.daemonBridge.IsConnected() {
		containers, err := s.daemonBridge.List()
		if err != nil {
			logging.Warn("get bot via daemon bridge failed",
				"error", err.Error(),
				logging.Component("api"))
			s.writeError(w, http.StatusServiceUnavailable, "daemon unavailable")
			return
		}

		for _, c := range containers {
			if c.ID == botID {
				bot := BotResponse{
					ID:        c.ID,
					Name:      c.ID,
					Image:     c.Image,
					Region:    firstOrDefault(c.Regions, "unknown"),
					CreatedAt: c.CreatedAt,
				}
				s.writeJSON(w, http.StatusOK, bot)
				return
			}
		}
	}

	s.writeError(w, http.StatusNotFound, "bot not found")
}

func (s *Server) updateBot(w http.ResponseWriter, r *http.Request, botID string) {
	var req map[string]interface{}
	if err := s.readJSON(r, &req); err != nil {
		s.writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (s *Server) deleteBot(w http.ResponseWriter, r *http.Request, botID string) {
	if s.daemonBridge != nil {
		if err := s.daemonBridge.Delete(botID); err != nil {
			logging.Warn("delete bot via daemon bridge failed",
				"error", err.Error(),
				"bot_id", botID,
				logging.Component("api"))
			s.writeError(w, http.StatusInternalServerError, "delete failed: "+err.Error())
			return
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleBotCloning(w http.ResponseWriter, r *http.Request, botID string) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req map[string]interface{}
	if err := s.readJSON(r, &req); err != nil {
		s.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]string{"status": "cloning configured"})
}

func (s *Server) handleBotStatus(w http.ResponseWriter, r *http.Request, botID string) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	// Return bot status
	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":             "running",
		"uptime":             "1h30m",
		"clones":             0,
		"active_deployments": 1,
		"threat_level":       0.0,
	})
}

func (s *Server) handleBotClones(w http.ResponseWriter, r *http.Request, botID string) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"clones": []interface{}{},
	})
}

func (s *Server) handleBotSyncClones(w http.ResponseWriter, r *http.Request, botID string) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]string{"status": "sync initiated"})
}

// ============================================================================
// Runtime Handlers
// ============================================================================

// RuntimeReserveRequest is the request for POST /runtimes/reserve
type RuntimeReserveRequest struct {
	BotID         string `json:"bot_id"`
	MinMemoryMB   int    `json:"min_memory_mb,omitempty"`
	MinCPUShares  int    `json:"min_cpu_shares,omitempty"`
	DurationHours int    `json:"duration_hours,omitempty"`
	Region        string `json:"region,omitempty"`
}

// RuntimeResponse is the response for runtime operations
type RuntimeResponse struct {
	ID        string           `json:"id"`
	BotID     string           `json:"bot_id"`
	NodeID    string           `json:"node_id"`
	Region    string           `json:"region"`
	Resources *ResourceRequest `json:"resources,omitempty"`
	ExpiresAt time.Time        `json:"expires_at"`
}

func (s *Server) handleRuntimeReserve(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req RuntimeReserveRequest
	if err := s.readJSON(r, &req); err != nil {
		s.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if req.BotID == "" {
		s.writeError(w, http.StatusBadRequest, "bot_id is required")
		return
	}

	durationHours := req.DurationHours
	if durationHours <= 0 {
		durationHours = 24
	}

	runtime := RuntimeResponse{
		ID:        fmt.Sprintf("rt-%d", time.Now().UnixNano()),
		BotID:     req.BotID,
		NodeID:    fmt.Sprintf("node-%d", time.Now().UnixNano()%1000),
		Region:    req.Region,
		ExpiresAt: time.Now().Add(time.Duration(durationHours) * time.Hour),
		Resources: &ResourceRequest{
			CPUShares:  req.MinCPUShares,
			MemoryMB:   req.MinMemoryMB,
		},
	}

	s.writeJSON(w, http.StatusCreated, runtime)
}

func (s *Server) handleRuntimeByID(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	parts := strings.Split(strings.TrimPrefix(path, "/v1/runtimes/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		s.writeError(w, http.StatusBadRequest, "runtime_id required")
		return
	}
	runtimeID := parts[0]

	// Check for sub-resources
	if len(parts) > 1 {
		switch parts[1] {
		case "extend":
			s.handleRuntimeExtend(w, r, runtimeID)
			return
		case "status":
			s.handleRuntimeStatus(w, r, runtimeID)
			return
		}
	}

	switch r.Method {
	case http.MethodGet:
		s.writeError(w, http.StatusNotFound, "runtime not found")
	case http.MethodDelete:
		w.WriteHeader(http.StatusNoContent)
	default:
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleRuntimeExtend(w http.ResponseWriter, r *http.Request, runtimeID string) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		DurationHours int `json:"duration_hours"`
	}
	if err := s.readJSON(r, &req); err != nil {
		s.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"expires_at": time.Now().Add(time.Duration(req.DurationHours) * time.Hour),
	})
}

func (s *Server) handleRuntimeStatus(w http.ResponseWriter, r *http.Request, runtimeID string) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":          "active",
		"remaining_hours": 23.5,
	})
}

// ============================================================================
// Deployment Handlers
// ============================================================================

// DeploymentRequest is the request for POST /deployments
type DeploymentRequest struct {
	RuntimeID  string            `json:"runtime_id"`
	Env        map[string]string `json:"env,omitempty"`
	Cmd        []string          `json:"cmd,omitempty"`
	Entrypoint []string          `json:"entrypoint,omitempty"`
}

// DeploymentResponse is the response for deployment operations
type DeploymentResponse struct {
	ID           string    `json:"id"`
	BotID        string    `json:"bot_id"`
	RuntimeID    string    `json:"runtime_id"`
	ContainerID  string    `json:"container_id"`
	Status       string    `json:"status"`
	Region       string    `json:"region"`
	NodeID       string    `json:"node_id"`
	OnionAddress string    `json:"onion_address,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	StartedAt    *time.Time `json:"started_at,omitempty"`
}

func (s *Server) handleDeployments(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listDeployments(w, r)
	case http.MethodPost:
		s.createDeployment(w, r)
	default:
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

func (s *Server) listDeployments(w http.ResponseWriter, r *http.Request) {
	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"deployments": []DeploymentResponse{},
	})
}

func (s *Server) createDeployment(w http.ResponseWriter, r *http.Request) {
	var req DeploymentRequest
	if err := s.readJSON(r, &req); err != nil {
		s.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if req.RuntimeID == "" {
		s.writeError(w, http.StatusBadRequest, "runtime_id is required")
		return
	}

	now := time.Now()
	deployment := DeploymentResponse{
		ID:          fmt.Sprintf("dep-%d", now.UnixNano()),
		BotID:       fmt.Sprintf("bot-%d", now.UnixNano()%10000),
		RuntimeID:   req.RuntimeID,
		ContainerID: fmt.Sprintf("mb-%d", now.UnixNano()),
		Status:      "starting",
		Region:      "americas",
		NodeID:      fmt.Sprintf("node-%d", now.UnixNano()%1000),
		CreatedAt:   now,
		StartedAt:   &now,
	}

	s.writeJSON(w, http.StatusCreated, deployment)
}

func (s *Server) handleDeploymentByID(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	parts := strings.Split(strings.TrimPrefix(path, "/v1/deployments/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		s.writeError(w, http.StatusBadRequest, "deployment_id required")
		return
	}
	deploymentID := parts[0]

	// Check for sub-resources
	if len(parts) > 1 && parts[1] == "stop" {
		s.handleDeploymentStop(w, r, deploymentID)
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.writeError(w, http.StatusNotFound, "deployment not found")
	default:
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleDeploymentStop(w http.ResponseWriter, r *http.Request, deploymentID string) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ============================================================================
// Snapshot Handlers
// ============================================================================

func (s *Server) handleSnapshots(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.writeJSON(w, http.StatusOK, map[string]interface{}{
			"snapshots": []interface{}{},
		})
	case http.MethodPost:
		s.handleSnapshot(w, r) // Use existing handler
	default:
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleSnapshotByID(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	parts := strings.Split(strings.TrimPrefix(path, "/v1/snapshots/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		s.writeError(w, http.StatusBadRequest, "snapshot_id required")
		return
	}
	snapshotID := parts[0]

	// Check for sub-resources
	if len(parts) > 1 {
		switch parts[1] {
		case "restore":
			s.handleRestore(w, r) // Use existing handler
			return
		case "data":
			s.handleSnapshotData(w, r, snapshotID)
			return
		}
	}

	switch r.Method {
	case http.MethodGet:
		s.writeError(w, http.StatusNotFound, "snapshot not found")
	case http.MethodDelete:
		w.WriteHeader(http.StatusNoContent)
	default:
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleSnapshotData(w http.ResponseWriter, r *http.Request, snapshotID string) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	if s.daemonBridge != nil {
		snap, err := s.daemonBridge.SnapshotGet(snapshotID)
		if err != nil {
			logging.Warn("snapshot get via daemon bridge failed",
				"error", err.Error(),
				"snapshot_id", snapshotID,
				logging.Component("api"))
			s.writeError(w, http.StatusNotFound, "snapshot not found")
			return
		}

		response := SnapshotResponse{
			SnapshotID:  snap.ID,
			ContainerID: snap.ContainerID,
			Type:        snap.Type,
			Size:        snap.Size,
			Checksum:    snap.Checksum,
			CreatedAt:   snap.CreatedAt,
		}
		s.writeJSON(w, http.StatusOK, response)
		return
	}

	s.writeError(w, http.StatusNotFound, "snapshot not found")
}

// ============================================================================
// Clone Handlers
// ============================================================================

func (s *Server) handleClones(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.writeJSON(w, http.StatusOK, map[string]interface{}{
			"clones": []interface{}{},
		})
	case http.MethodPost:
		s.handleClone(w, r) // Use existing handler
	default:
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleCloneByID(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	parts := strings.Split(strings.TrimPrefix(path, "/v1/clones/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		s.writeError(w, http.StatusBadRequest, "clone_id required")
		return
	}
	cloneID := parts[0]

	// Check for sub-resources
	if len(parts) > 1 && parts[1] == "cancel" {
		s.handleCloneCancel(w, r, cloneID)
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.writeError(w, http.StatusNotFound, "clone not found")
	default:
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleCloneCancel(w http.ResponseWriter, r *http.Request, cloneID string) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ============================================================================
// Container Handlers
// ============================================================================

// handleContainers handles GET /v1/containers (list all containers)
func (s *Server) handleContainers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	if s.daemonBridge == nil {
		s.writeJSON(w, http.StatusOK, []interface{}{})
		return
	}

	containers, err := s.daemonBridge.List()
	if err != nil {
		logging.Warn("list containers via daemon bridge failed",
			"error", err.Error(),
			logging.Component("api"))
		s.writeJSON(w, http.StatusOK, []interface{}{})
		return
	}

	// Filter containers by authenticated wallet
	wallet := s.extractWalletAddress(r)
	if wallet != "" {
		filtered := make([]client.ContainerInfo, 0, len(containers))
		for _, c := range containers {
			if strings.EqualFold(c.Owner, wallet) {
				filtered = append(filtered, c)
			}
		}
		containers = filtered
	}

	s.writeJSON(w, http.StatusOK, containers)
}

func (s *Server) handleContainerByID(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	parts := strings.Split(strings.TrimPrefix(path, "/v1/containers/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		// No container ID — treat as list request
		s.handleContainers(w, r)
		return
	}
	containerID := parts[0]

	// Check for sub-resources
	if len(parts) > 1 {
		switch parts[1] {
		case "logs":
			s.handleContainerLogs(w, r, containerID)
			return
		case "stop":
			if r.Method != http.MethodPost {
				http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
				return
			}
			if s.daemonBridge == nil {
				s.writeError(w, http.StatusServiceUnavailable, "daemon not available")
				return
			}
			// Verify ownership before stop
			stopWallet := s.extractWalletAddress(r)
			if stopWallet != "" {
				containers, err := s.daemonBridge.List()
				if err == nil {
					for _, c := range containers {
						if c.ID == containerID && !strings.EqualFold(c.Owner, stopWallet) {
							s.writeError(w, http.StatusForbidden, "not your container")
							return
						}
					}
				}
			}
			if err := s.daemonBridge.Stop(containerID); err != nil {
				s.writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
			s.writeJSON(w, http.StatusOK, map[string]string{"status": "stopped"})
			return
		case "start":
			if r.Method != http.MethodPost {
				http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
				return
			}
			if s.daemonBridge == nil {
				s.writeError(w, http.StatusServiceUnavailable, "daemon not available")
				return
			}
			// Verify ownership before start
			startWallet := s.extractWalletAddress(r)
			if startWallet != "" {
				containers, err := s.daemonBridge.List()
				if err == nil {
					for _, c := range containers {
						if c.ID == containerID && !strings.EqualFold(c.Owner, startWallet) {
							s.writeError(w, http.StatusForbidden, "not your container")
							return
						}
					}
				}
			}
			if err := s.daemonBridge.Start(containerID); err != nil {
				s.writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
			s.writeJSON(w, http.StatusOK, map[string]string{"status": "started"})
			return
		case "cloning":
			s.handleContainerCloning(w, r, containerID)
			return
		case "checkpoints":
			s.handleContainerCheckpoints(w, r, containerID)
			return
		}
	}

	wallet := s.extractWalletAddress(r)

	switch r.Method {
	case http.MethodGet:
		if s.daemonBridge == nil {
			s.writeError(w, http.StatusNotFound, "container not found")
			return
		}
		containers, err := s.daemonBridge.List()
		if err != nil {
			s.writeError(w, http.StatusInternalServerError, "failed to list containers")
			return
		}
		for _, c := range containers {
			if c.ID == containerID {
				if wallet != "" && !strings.EqualFold(c.Owner, wallet) {
					s.writeError(w, http.StatusNotFound, "container not found")
					return
				}
				s.writeJSON(w, http.StatusOK, c)
				return
			}
		}
		s.writeError(w, http.StatusNotFound, "container not found")
	case http.MethodDelete:
		if s.daemonBridge == nil {
			s.writeError(w, http.StatusServiceUnavailable, "daemon not available")
			return
		}
		// Verify ownership before delete
		if wallet != "" {
			containers, err := s.daemonBridge.List()
			if err != nil {
				s.writeError(w, http.StatusInternalServerError, "failed to verify ownership")
				return
			}
			found := false
			for _, c := range containers {
				if c.ID == containerID {
					if !strings.EqualFold(c.Owner, wallet) {
						s.writeError(w, http.StatusForbidden, "not your container")
						return
					}
					found = true
					break
				}
			}
			if !found {
				s.writeError(w, http.StatusNotFound, "container not found")
				return
			}
		}
		if err := s.daemonBridge.Delete(containerID); err != nil {
			s.writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		s.writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
	default:
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleContainerLogs(w http.ResponseWriter, r *http.Request, containerID string) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	// Parse optional tail parameter (default 100 lines)
	tail := 100
	if tailStr := r.URL.Query().Get("tail"); tailStr != "" {
		if n, err := fmt.Sscanf(tailStr, "%d", &tail); n != 1 || err != nil {
			tail = 100
		}
	}

	// Try to get logs from daemon bridge
	if s.daemonBridge != nil {
		logs, err := s.daemonBridge.GetLogs(containerID, false, tail)
		if err != nil {
			logging.Warn("get container logs via daemon bridge failed",
				"error", err.Error(),
				"container_id", containerID,
				logging.Component("api"))
			s.writeError(w, http.StatusServiceUnavailable, "logs unavailable: daemon connection failed")
			return
		}
		s.writeJSON(w, http.StatusOK, map[string]string{
			"logs": logs,
		})
		return
	}

	s.writeError(w, http.StatusServiceUnavailable, "logs unavailable: daemon bridge not configured")
}

func (s *Server) handleContainerCloning(w http.ResponseWriter, r *http.Request, containerID string) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req map[string]interface{}
	if err := s.readJSON(r, &req); err != nil {
		s.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]string{"status": "cloning configured"})
}

func (s *Server) handleContainerCheckpoints(w http.ResponseWriter, r *http.Request, containerID string) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req map[string]interface{}
	if err := s.readJSON(r, &req); err != nil {
		s.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]string{"status": "checkpoints configured"})
}

// ============================================================================
// Bootstrap Handlers
// ============================================================================

// BootstrapPeer is a peer entry returned by the bootstrap endpoint.
type BootstrapPeer struct {
	NodeID  string   `json:"node_id"`
	Addrs   []string `json:"addrs"`
	Region  string   `json:"region,omitempty"`
	Country string   `json:"country,omitempty"`
}

// BootstrapResponse is the response for GET /v1/bootstrap.
type BootstrapResponse struct {
	Peers []BootstrapPeer `json:"peers"`
}

// handleBootstrap handles GET /v1/bootstrap (public, no auth).
// Returns known online peers so new nodes can discover the network
// without DNS TXT records being set up.
func (s *Server) handleBootstrap(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	const maxBootstrapPeers = 20

	peers := make([]BootstrapPeer, 0)

	if s.daemonBridge != nil {
		status, err := s.daemonBridge.Status()
		if err != nil {
			logging.Debug("bootstrap: daemon bridge unavailable",
				"error", err.Error(),
				logging.Component("api"))
		} else if status != nil {
			// Collect online peers, most recently seen first
			// (KnownNodes are already sorted by the daemon)
			for _, n := range status.KnownNodes {
				if !n.Online || n.Address == "" {
					continue
				}
				if n.Blocked {
					continue
				}

				bp := BootstrapPeer{
					NodeID:  n.NodeID,
					Addrs:   []string{n.Address},
					Region:  n.Region,
					Country: n.Country,
				}
				peers = append(peers, bp)

				if len(peers) >= maxBootstrapPeers {
					break
				}
			}
		}
	}

	s.writeJSON(w, http.StatusOK, BootstrapResponse{Peers: peers})
}
