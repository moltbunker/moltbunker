package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/moltbunker/moltbunker/pkg/types"
)

// Requester API request/response types

// RequesterJob represents a job from the requester's perspective
type RequesterJob struct {
	ID             string    `json:"id"`
	ContainerID    string    `json:"container_id"`
	Image          string    `json:"image"`
	Status         string    `json:"status"`
	ProviderNodeID string    `json:"provider_node_id"`
	Region         string    `json:"region"`
	CreatedAt      time.Time `json:"created_at"`
	StartedAt      time.Time `json:"started_at,omitempty"`
	CompletedAt    time.Time `json:"completed_at,omitempty"`
	Cost           JobCost   `json:"cost"`
	OnionAddress   string    `json:"onion_address,omitempty"`
	Replicas       int       `json:"replicas"`
}

// JobCost contains cost information for a job
type JobCost struct {
	EstimatedTotal string `json:"estimated_total"`
	ActualSpent    string `json:"actual_spent"`
	Rate           string `json:"rate_per_hour"`
}

// RequesterOutput represents output from a completed job
type RequesterOutput struct {
	JobID       string    `json:"job_id"`
	ContainerID string    `json:"container_id"`
	ExitCode    int       `json:"exit_code"`
	Logs        string    `json:"logs,omitempty"`
	Artifacts   []string  `json:"artifacts,omitempty"`
	CompletedAt time.Time `json:"completed_at"`
}

// RequesterBalance contains balance information
type RequesterBalance struct {
	WalletAddress  string `json:"wallet_address"`
	BUNKERBalance  string `json:"bunker_balance"`
	ETHBalance     string `json:"eth_balance"`
	DepositedFunds string `json:"deposited_funds"` // Funds in escrow
	ReservedFunds  string `json:"reserved_funds"`  // Funds reserved for active jobs
	AvailableFunds string `json:"available_funds"` // Can be used for new deployments
}

// RequesterDepositRequest contains deposit parameters
type RequesterDepositRequest struct {
	Amount string `json:"amount"` // Amount in BUNKER tokens
}

// RequesterDepositResponse contains deposit result
type RequesterDepositResponse struct {
	TransactionHash string `json:"transaction_hash"`
	NewBalance      string `json:"new_balance"`
	Status          string `json:"status"`
}

// RequesterWithdrawRequest contains withdrawal parameters
type RequesterWithdrawRequest struct {
	Amount string `json:"amount"` // Amount to withdraw (or "all")
}

// handleRequesterJobs handles listing requester's jobs
func (s *APIServer) handleRequesterJobs(ctx context.Context, req *APIRequest) *APIResponse {
	var params struct {
		Status string `json:"status"` // "active", "completed", "failed", "all"
		Limit  int    `json:"limit"`
		Offset int    `json:"offset"`
	}

	// Parse optional params
	if req.Params != nil {
		json.Unmarshal(req.Params, &params)
	}

	if params.Status == "" {
		params.Status = "all"
	}
	if params.Limit <= 0 {
		params.Limit = 50
	}

	// Get deployments (these are jobs from requester perspective)
	deployments := s.containerManager.ListDeployments()

	jobs := make([]RequesterJob, 0, len(deployments))
	for _, d := range deployments {
		// Filter by status if specified
		if params.Status != "all" {
			if params.Status == "active" && d.Status != types.ContainerStatusRunning {
				continue
			}
			if params.Status == "completed" && d.Status != types.ContainerStatusStopped {
				continue
			}
			if params.Status == "failed" && d.Status != types.ContainerStatusFailed {
				continue
			}
		}

		job := RequesterJob{
			ID:             d.ID,
			ContainerID:    d.ID,
			Image:          d.Image,
			Status:         string(d.Status),
			ProviderNodeID: s.node.nodeInfo.ID.String(), // Local node for now
			Region:         s.node.nodeInfo.Region,
			CreatedAt:      d.CreatedAt,
			StartedAt:      d.StartedAt,
			OnionAddress:   d.OnionAddress,
			Replicas:       len(d.Regions),
			Cost: JobCost{
				EstimatedTotal: "0",
				ActualSpent:    "0",
				Rate:           "0",
			},
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

// handleRequesterOutputs handles retrieving job outputs
func (s *APIServer) handleRequesterOutputs(ctx context.Context, req *APIRequest) *APIResponse {
	var params struct {
		JobID string `json:"job_id"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return &APIResponse{
			Error: fmt.Sprintf("invalid params: %v", err),
			ID:    req.ID,
		}
	}

	if params.JobID == "" {
		return &APIResponse{
			Error: "job_id is required",
			ID:    req.ID,
		}
	}

	// Validate job/container ID
	if err := validateContainerID(params.JobID); err != nil {
		return &APIResponse{
			Error: fmt.Sprintf("validation failed: %v", err),
			ID:    req.ID,
		}
	}

	// Get logs for the container
	reader, err := s.containerManager.GetLogs(ctx, params.JobID, false, 1000)
	if err != nil {
		return &APIResponse{
			Error: fmt.Sprintf("failed to get job output: %v", err),
			ID:    req.ID,
		}
	}
	defer reader.Close()

	// Read logs
	logData := make([]byte, 0)
	buf := make([]byte, 4096)
	for {
		n, err := reader.Read(buf)
		if n > 0 {
			logData = append(logData, buf[:n]...)
		}
		if err != nil {
			break
		}
		// Limit log size
		if len(logData) > 1024*1024 { // 1MB max
			break
		}
	}

	output := RequesterOutput{
		JobID:       params.JobID,
		ContainerID: params.JobID,
		ExitCode:    0, // Would need to track this
		Logs:        string(logData),
		Artifacts:   []string{}, // Would need to implement artifact storage
		CompletedAt: time.Now(),
	}

	return &APIResponse{
		Result: output,
		ID:     req.ID,
	}
}

// handleRequesterBalance handles balance queries
func (s *APIServer) handleRequesterBalance(ctx context.Context, req *APIRequest) *APIResponse {
	// In a real implementation, this would:
	// 1. Query BUNKER token balance from contract
	// 2. Query ETH balance
	// 3. Query deposited/reserved funds from escrow contract

	walletAddress := ""
	if s.config != nil {
		walletAddress = s.config.Node.WalletAddress
	}

	balance := RequesterBalance{
		WalletAddress:  walletAddress,
		BUNKERBalance:  "0",
		ETHBalance:     "0",
		DepositedFunds: "0",
		ReservedFunds:  "0",
		AvailableFunds: "0",
	}

	return &APIResponse{
		Result: balance,
		ID:     req.ID,
	}
}

// handleRequesterDeposit handles depositing funds for deployments
func (s *APIServer) handleRequesterDeposit(ctx context.Context, req *APIRequest) *APIResponse {
	var depositReq RequesterDepositRequest
	if err := json.Unmarshal(req.Params, &depositReq); err != nil {
		return &APIResponse{
			Error: fmt.Sprintf("invalid deposit params: %v", err),
			ID:    req.ID,
		}
	}

	if depositReq.Amount == "" || depositReq.Amount == "0" {
		return &APIResponse{
			Error: "amount is required and must be greater than 0",
			ID:    req.ID,
		}
	}

	// In a real implementation, this would:
	// 1. Approve token transfer to escrow contract
	// 2. Call deposit() on escrow contract
	// 3. Wait for confirmation

	response := RequesterDepositResponse{
		TransactionHash: "0x" + generateMockTxHash(),
		NewBalance:      depositReq.Amount,
		Status:          "pending",
	}

	return &APIResponse{
		Result: response,
		ID:     req.ID,
	}
}

// handleRequesterWithdraw handles withdrawing unused funds
func (s *APIServer) handleRequesterWithdraw(ctx context.Context, req *APIRequest) *APIResponse {
	var withdrawReq RequesterWithdrawRequest
	if err := json.Unmarshal(req.Params, &withdrawReq); err != nil {
		return &APIResponse{
			Error: fmt.Sprintf("invalid withdraw params: %v", err),
			ID:    req.ID,
		}
	}

	if withdrawReq.Amount == "" {
		return &APIResponse{
			Error: "amount is required",
			ID:    req.ID,
		}
	}

	// In a real implementation, this would:
	// 1. Check available (non-reserved) balance
	// 2. Call withdraw() on escrow contract
	// 3. Wait for confirmation

	return &APIResponse{
		Result: map[string]interface{}{
			"transaction_hash": "0x" + generateMockTxHash(),
			"withdrawn_amount": withdrawReq.Amount,
			"remaining_balance": "0",
			"status":           "pending",
		},
		ID: req.ID,
	}
}

// handleRequesterEstimate handles cost estimation for deployments
func (s *APIServer) handleRequesterEstimate(ctx context.Context, req *APIRequest) *APIResponse {
	var params struct {
		Image        string `json:"image"`
		CPUCores     int    `json:"cpu_cores"`
		MemoryGB     int    `json:"memory_gb"`
		StorageGB    int    `json:"storage_gb"`
		DurationHours int   `json:"duration_hours"`
		Replicas     int    `json:"replicas"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return &APIResponse{
			Error: fmt.Sprintf("invalid params: %v", err),
			ID:    req.ID,
		}
	}

	// Set defaults
	if params.CPUCores <= 0 {
		params.CPUCores = 1
	}
	if params.MemoryGB <= 0 {
		params.MemoryGB = 1
	}
	if params.StorageGB <= 0 {
		params.StorageGB = 10
	}
	if params.DurationHours <= 0 {
		params.DurationHours = 24
	}
	if params.Replicas <= 0 {
		params.Replicas = 3
	}

	// In a real implementation, this would query the pricing oracle
	// For now, return a mock estimate

	// Mock pricing (BUNKER tokens per resource per hour)
	cpuRate := 0.01    // per core per hour
	memRate := 0.005   // per GB per hour
	storageRate := 0.001 // per GB per hour

	hourlyRate := float64(params.CPUCores)*cpuRate +
		float64(params.MemoryGB)*memRate +
		float64(params.StorageGB)*storageRate

	totalCost := hourlyRate * float64(params.DurationHours) * float64(params.Replicas)

	return &APIResponse{
		Result: map[string]interface{}{
			"estimate": map[string]interface{}{
				"hourly_rate":   fmt.Sprintf("%.6f", hourlyRate),
				"total_cost":    fmt.Sprintf("%.6f", totalCost),
				"duration_hours": params.DurationHours,
				"replicas":      params.Replicas,
			},
			"breakdown": map[string]interface{}{
				"cpu_cost":     fmt.Sprintf("%.6f", float64(params.CPUCores)*cpuRate*float64(params.DurationHours)*float64(params.Replicas)),
				"memory_cost":  fmt.Sprintf("%.6f", float64(params.MemoryGB)*memRate*float64(params.DurationHours)*float64(params.Replicas)),
				"storage_cost": fmt.Sprintf("%.6f", float64(params.StorageGB)*storageRate*float64(params.DurationHours)*float64(params.Replicas)),
			},
			"currency": "BUNKER",
		},
		ID: req.ID,
	}
}
