package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/moltbunker/moltbunker/pkg/types"
)

// Provider API request/response types

// ProviderRegisterRequest contains registration parameters
type ProviderRegisterRequest struct {
	WalletAddress string               `json:"wallet_address"`
	Resources     ProviderResources    `json:"resources"`
	Region        string               `json:"region"`
	Capabilities  ProviderCapabilities `json:"capabilities"`
}

// ProviderResources describes declared provider resources
type ProviderResources struct {
	CPUCores      int   `json:"cpu_cores"`
	MemoryGB      int   `json:"memory_gb"`
	StorageGB     int   `json:"storage_gb"`
	BandwidthMbps int   `json:"bandwidth_mbps"`
	GPUCount      int   `json:"gpu_count,omitempty"`
	GPUModel      string `json:"gpu_model,omitempty"`
}

// ProviderCapabilities describes what the provider can do
type ProviderCapabilities struct {
	TorSupport      bool `json:"tor_support"`
	EncryptedVolumes bool `json:"encrypted_volumes"`
	GPUCompute      bool `json:"gpu_compute"`
}

// ProviderStatusResponse contains provider status information
type ProviderStatusResponse struct {
	Registered     bool                `json:"registered"`
	WalletAddress  string              `json:"wallet_address,omitempty"`
	NodeID         string              `json:"node_id"`
	Status         string              `json:"status"` // active, maintenance, suspended
	Tier           types.StakingTier   `json:"tier"`
	StakedAmount   string              `json:"staked_amount"`
	ActiveJobs     int                 `json:"active_jobs"`
	TotalJobsRun   int                 `json:"total_jobs_run"`
	Reputation     int                 `json:"reputation"`     // 0-1000
	Uptime         float64             `json:"uptime_percent"` // 0-100
	Region         string              `json:"region"`
	Resources      ProviderResources   `json:"resources"`
	Earnings       ProviderEarnings    `json:"earnings"`
	RegisteredAt   time.Time           `json:"registered_at,omitempty"`
}

// ProviderEarnings contains earnings information
type ProviderEarnings struct {
	TotalEarned    string `json:"total_earned"`
	PendingPayout  string `json:"pending_payout"`
	LastPayoutAt   time.Time `json:"last_payout_at,omitempty"`
	ThisMonth      string `json:"this_month"`
}

// ProviderStakeRequest contains staking parameters
type ProviderStakeRequest struct {
	Amount     string            `json:"amount"`      // Amount in BUNKER tokens
	TargetTier types.StakingTier `json:"target_tier"` // Target staking tier
}

// ProviderStakeResponse contains staking result
type ProviderStakeResponse struct {
	TransactionHash string            `json:"transaction_hash"`
	NewStake        string            `json:"new_stake"`
	NewTier         types.StakingTier `json:"new_tier"`
	Status          string            `json:"status"`
}

// ProviderWithdrawRequest contains withdrawal parameters
type ProviderWithdrawRequest struct {
	Amount string `json:"amount"` // Amount to withdraw (or "all")
}

// ProviderWithdrawResponse contains withdrawal result
type ProviderWithdrawResponse struct {
	TransactionHash  string    `json:"transaction_hash"`
	WithdrawnAmount  string    `json:"withdrawn_amount"`
	RemainingStake   string    `json:"remaining_stake"`
	NewTier          types.StakingTier `json:"new_tier"`
	UnlockTime       time.Time `json:"unlock_time"` // When funds become available
	Status           string    `json:"status"`
}

// ProviderJob represents a job assigned to the provider
type ProviderJob struct {
	ID            string    `json:"id"`
	RequesterID   string    `json:"requester_id"`
	ContainerID   string    `json:"container_id"`
	Status        string    `json:"status"`
	StartedAt     time.Time `json:"started_at"`
	Duration      string    `json:"duration"`
	EarnedAmount  string    `json:"earned_amount"`
	ResourceUsage struct {
		CPUPercent    float64 `json:"cpu_percent"`
		MemoryMB      int64   `json:"memory_mb"`
		NetworkInMB   int64   `json:"network_in_mb"`
		NetworkOutMB  int64   `json:"network_out_mb"`
	} `json:"resource_usage"`
}

// handleProviderRegister handles provider registration requests
func (s *APIServer) handleProviderRegister(ctx context.Context, req *APIRequest) *APIResponse {
	var regReq ProviderRegisterRequest
	if err := json.Unmarshal(req.Params, &regReq); err != nil {
		return &APIResponse{
			Error: fmt.Sprintf("invalid register params: %v", err),
			ID:    req.ID,
		}
	}

	// Validate wallet address
	if regReq.WalletAddress == "" {
		return &APIResponse{
			Error: "wallet_address is required",
			ID:    req.ID,
		}
	}

	// Validate resources
	if regReq.Resources.CPUCores <= 0 {
		return &APIResponse{
			Error: "cpu_cores must be positive",
			ID:    req.ID,
		}
	}

	// In a real implementation, this would:
	// 1. Sign a registration message with the node's key
	// 2. Submit to the provider registry contract
	// 3. Wait for confirmation

	// For now, return a success response with mock data
	response := ProviderStatusResponse{
		Registered:    true,
		WalletAddress: regReq.WalletAddress,
		NodeID:        s.node.nodeInfo.ID.String(),
		Status:        "active",
		Tier:          types.StakingTierStarter,
		StakedAmount:  "0",
		ActiveJobs:    0,
		TotalJobsRun:  0,
		Reputation:    500, // Starting reputation
		Uptime:        100.0,
		Region:        regReq.Region,
		Resources:     regReq.Resources,
		Earnings: ProviderEarnings{
			TotalEarned:   "0",
			PendingPayout: "0",
			ThisMonth:     "0",
		},
		RegisteredAt: time.Now(),
	}

	return &APIResponse{
		Result: response,
		ID:     req.ID,
	}
}

// handleProviderStatus handles provider status requests
func (s *APIServer) handleProviderStatus(ctx context.Context, req *APIRequest) *APIResponse {
	// Get status from provider state
	// In a real implementation, this would query the on-chain provider registry

	// Check if this node is registered as a provider
	if s.config == nil || !s.config.IsProvider() {
		return &APIResponse{
			Result: ProviderStatusResponse{
				Registered: false,
				NodeID:     s.node.nodeInfo.ID.String(),
				Status:     "not_registered",
			},
			ID: req.ID,
		}
	}

	// Return current provider status
	containers := s.containerManager.ListDeployments()

	response := ProviderStatusResponse{
		Registered:    true,
		WalletAddress: s.config.Node.WalletAddress,
		NodeID:        s.node.nodeInfo.ID.String(),
		Status:        "active",
		Tier:          s.config.Node.Provider.TargetTier,
		StakedAmount:  "0", // Would query from contract
		ActiveJobs:    len(containers),
		TotalJobsRun:  0,   // Would track in state
		Reputation:    750, // Would query from contract
		Uptime:        99.9,
		Region:        s.node.nodeInfo.Region,
		Resources: ProviderResources{
			CPUCores:      s.config.Node.Provider.DeclaredCPU,
			MemoryGB:      s.config.Node.Provider.DeclaredMemoryGB,
			StorageGB:     s.config.Node.Provider.DeclaredStorageGB,
			BandwidthMbps: s.config.Node.Provider.DeclaredBandwidth,
			GPUCount:      s.config.Node.Provider.GPUCount,
			GPUModel:      s.config.Node.Provider.GPUModel,
		},
		Earnings: ProviderEarnings{
			TotalEarned:   "0",
			PendingPayout: "0",
			ThisMonth:     "0",
		},
	}

	if s.config.Node.Provider.MaintenanceMode {
		response.Status = "maintenance"
	}

	return &APIResponse{
		Result: response,
		ID:     req.ID,
	}
}

// handleProviderStakeAdd handles adding stake
func (s *APIServer) handleProviderStakeAdd(ctx context.Context, req *APIRequest) *APIResponse {
	var stakeReq ProviderStakeRequest
	if err := json.Unmarshal(req.Params, &stakeReq); err != nil {
		return &APIResponse{
			Error: fmt.Sprintf("invalid stake params: %v", err),
			ID:    req.ID,
		}
	}

	// Validate amount
	if stakeReq.Amount == "" || stakeReq.Amount == "0" {
		return &APIResponse{
			Error: "amount is required and must be greater than 0",
			ID:    req.ID,
		}
	}

	// In a real implementation, this would:
	// 1. Approve token transfer to staking contract
	// 2. Call stake() on the staking contract
	// 3. Wait for confirmation

	response := ProviderStakeResponse{
		TransactionHash: "0x" + generateMockTxHash(),
		NewStake:        stakeReq.Amount,
		NewTier:         stakeReq.TargetTier,
		Status:          "pending",
	}

	return &APIResponse{
		Result: response,
		ID:     req.ID,
	}
}

// handleProviderStakeWithdraw handles withdrawing stake
func (s *APIServer) handleProviderStakeWithdraw(ctx context.Context, req *APIRequest) *APIResponse {
	var withdrawReq ProviderWithdrawRequest
	if err := json.Unmarshal(req.Params, &withdrawReq); err != nil {
		return &APIResponse{
			Error: fmt.Sprintf("invalid withdraw params: %v", err),
			ID:    req.ID,
		}
	}

	// Validate amount
	if withdrawReq.Amount == "" {
		return &APIResponse{
			Error: "amount is required",
			ID:    req.ID,
		}
	}

	// In a real implementation, this would:
	// 1. Check if provider has any active jobs (must complete first)
	// 2. Initiate unstaking (may have lock-up period)
	// 3. Return unlock time

	// Default unstaking period is 7 days
	unlockTime := time.Now().Add(7 * 24 * time.Hour)

	response := ProviderWithdrawResponse{
		TransactionHash:  "0x" + generateMockTxHash(),
		WithdrawnAmount:  withdrawReq.Amount,
		RemainingStake:   "0",
		NewTier:          types.StakingTierStarter,
		UnlockTime:       unlockTime,
		Status:           "pending_unlock",
	}

	return &APIResponse{
		Result: response,
		ID:     req.ID,
	}
}

// handleProviderEarnings handles earnings query
func (s *APIServer) handleProviderEarnings(ctx context.Context, req *APIRequest) *APIResponse {
	// In a real implementation, this would query earnings from the payment contract

	earnings := ProviderEarnings{
		TotalEarned:   "0",
		PendingPayout: "0",
		ThisMonth:     "0",
	}

	return &APIResponse{
		Result: earnings,
		ID:     req.ID,
	}
}

// handleProviderJobs handles listing provider's jobs
func (s *APIServer) handleProviderJobs(ctx context.Context, req *APIRequest) *APIResponse {
	var params struct {
		Status string `json:"status"` // "active", "completed", "all"
		Limit  int    `json:"limit"`
		Offset int    `json:"offset"`
	}

	// Parse optional params
	if req.Params != nil {
		json.Unmarshal(req.Params, &params)
	}

	if params.Status == "" {
		params.Status = "active"
	}
	if params.Limit <= 0 {
		params.Limit = 50
	}

	// Get jobs from container manager
	deployments := s.containerManager.ListDeployments()

	jobs := make([]ProviderJob, 0, len(deployments))
	for _, d := range deployments {
		if params.Status == "active" && d.Status != types.ContainerStatusRunning {
			continue
		}

		job := ProviderJob{
			ID:          d.ID,
			RequesterID: "unknown", // Would need to track this
			ContainerID: d.ID,
			Status:      string(d.Status),
			StartedAt:   d.StartedAt,
		}

		if !d.StartedAt.IsZero() {
			job.Duration = time.Since(d.StartedAt).Round(time.Second).String()
		}

		jobs = append(jobs, job)
	}

	return &APIResponse{
		Result: map[string]interface{}{
			"jobs":   jobs,
			"total":  len(jobs),
			"limit":  params.Limit,
			"offset": params.Offset,
		},
		ID: req.ID,
	}
}

// handleProviderMaintenance handles maintenance mode toggle
func (s *APIServer) handleProviderMaintenance(ctx context.Context, req *APIRequest) *APIResponse {
	var params struct {
		Enable bool   `json:"enable"`
		Reason string `json:"reason,omitempty"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return &APIResponse{
			Error: fmt.Sprintf("invalid params: %v", err),
			ID:    req.ID,
		}
	}

	// In a real implementation, this would:
	// 1. Stop accepting new jobs
	// 2. Update on-chain status
	// 3. Optionally drain existing jobs

	status := "active"
	if params.Enable {
		status = "maintenance"
	}

	return &APIResponse{
		Result: map[string]interface{}{
			"status":  status,
			"reason":  params.Reason,
			"enabled": params.Enable,
		},
		ID: req.ID,
	}
}

// Helper function to generate mock transaction hash
func generateMockTxHash() string {
	// This would be a real tx hash in production
	return "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
}
