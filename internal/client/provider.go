package client

import (
	"encoding/json"
	"fmt"
)

// Provider-related request/response types

// ProviderRegisterRequest contains provider registration parameters
type ProviderRegisterRequest struct {
	DeclaredCPU    int    `json:"declared_cpu"`
	DeclaredMemory int    `json:"declared_memory"`
	DeclaredDisk   int    `json:"declared_disk"`
	TargetTier     string `json:"target_tier"`
	AutoStake      bool   `json:"auto_stake"`
	GPUEnabled     bool   `json:"gpu_enabled"`
	GPUModel       string `json:"gpu_model,omitempty"`
	GPUCount       int    `json:"gpu_count,omitempty"`
}

// ProviderRegisterResponse contains registration result
type ProviderRegisterResponse struct {
	NodeID        string `json:"node_id"`
	CurrentTier   string `json:"current_tier"`
	StakedAmount  string `json:"staked_amount"`
	RequiredStake string `json:"required_stake,omitempty"`
}

// ProviderStatusResponse contains provider status information
type ProviderStatusResponse struct {
	NodeID              string  `json:"node_id"`
	WalletAddress       string  `json:"wallet_address"`
	Status              string  `json:"status"`
	Tier                string  `json:"tier"`
	StakedAmount        string  `json:"staked_amount"`
	NextTier            string  `json:"next_tier"`
	NextTierRequired    string  `json:"next_tier_required"`
	ReputationScore     int     `json:"reputation_score"`
	ReputationTier      string  `json:"reputation_tier"`
	Uptime30Days        float64 `json:"uptime_30_days"`
	ActiveJobs          int     `json:"active_jobs"`
	MaxConcurrentJobs   int     `json:"max_concurrent_jobs"`
	JobsCompleted7Days  int     `json:"jobs_completed_7_days"`
	TotalJobsCompleted  int     `json:"total_jobs_completed"`
	DeclaredCPU         int     `json:"declared_cpu"`
	DeclaredMemory      int     `json:"declared_memory"`
	DeclaredDisk        int     `json:"declared_disk"`
	CPUUsagePercent     float64 `json:"cpu_usage_percent"`
	MemoryUsagePercent  float64 `json:"memory_usage_percent"`
	DiskUsagePercent    float64 `json:"disk_usage_percent"`
	GPUEnabled          bool    `json:"gpu_enabled"`
	GPUModel            string  `json:"gpu_model,omitempty"`
	GPUCount            int     `json:"gpu_count,omitempty"`
}

// ProviderStakeAddResponse contains stake add result
type ProviderStakeAddResponse struct {
	TxHash          string `json:"tx_hash"`
	NewStakedAmount string `json:"new_staked_amount"`
	NewTier         string `json:"new_tier"`
}

// ProviderStakeWithdrawResponse contains unstake initiation result
type ProviderStakeWithdrawResponse struct {
	InitiatedAt string `json:"initiated_at"`
	AvailableAt string `json:"available_at"`
	Amount      string `json:"amount"`
}

// ProviderEarningsResponse contains earnings information
type ProviderEarningsResponse struct {
	TotalEarnings   string           `json:"total_earnings"`
	Earnings30Days  string           `json:"earnings_30_days"`
	Earnings7Days   string           `json:"earnings_7_days"`
	PendingPayouts  string           `json:"pending_payouts"`
	RecentPayments  []PaymentInfo    `json:"recent_payments"`
}

// PaymentInfo contains payment details
type PaymentInfo struct {
	Date   string `json:"date"`
	JobID  string `json:"job_id"`
	Amount string `json:"amount"`
	TxHash string `json:"tx_hash"`
}

// ProviderJobsResponse contains job information
type ProviderJobsResponse struct {
	ActiveJobs []ProviderJobInfo `json:"active_jobs"`
	RecentJobs []ProviderJobInfo `json:"recent_jobs"`
}

// ProviderJobInfo contains job details
type ProviderJobInfo struct {
	ID          string  `json:"id"`
	Image       string  `json:"image"`
	Status      string  `json:"status"`
	CPUUsage    float64 `json:"cpu_usage"`
	MemoryUsage float64 `json:"memory_usage"`
	StartedAt   string  `json:"started_at"`
	Duration    string  `json:"duration,omitempty"`
	Earned      string  `json:"earned,omitempty"`
}

// ProviderRegister registers this node as a provider
func (c *DaemonClient) ProviderRegister(req *ProviderRegisterRequest) (*ProviderRegisterResponse, error) {
	resp, err := c.call("provider_register", req)
	if err != nil {
		return nil, err
	}

	var result ProviderRegisterResponse
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to parse register response: %w", err)
	}

	return &result, nil
}

// ProviderStatus retrieves provider status
func (c *DaemonClient) ProviderStatus() (*ProviderStatusResponse, error) {
	resp, err := c.call("provider_status", nil)
	if err != nil {
		return nil, err
	}

	var result ProviderStatusResponse
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to parse status response: %w", err)
	}

	return &result, nil
}

// ProviderStakeAdd stakes additional BUNKER tokens
func (c *DaemonClient) ProviderStakeAdd(amount string) (*ProviderStakeAddResponse, error) {
	resp, err := c.call("provider_stake_add", map[string]string{"amount": amount})
	if err != nil {
		return nil, err
	}

	var result ProviderStakeAddResponse
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to parse stake response: %w", err)
	}

	return &result, nil
}

// ProviderStakeWithdraw initiates unstaking
func (c *DaemonClient) ProviderStakeWithdraw() (*ProviderStakeWithdrawResponse, error) {
	resp, err := c.call("provider_stake_withdraw", nil)
	if err != nil {
		return nil, err
	}

	var result ProviderStakeWithdrawResponse
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to parse withdraw response: %w", err)
	}

	return &result, nil
}

// ProviderEarnings retrieves earnings information
func (c *DaemonClient) ProviderEarnings() (*ProviderEarningsResponse, error) {
	resp, err := c.call("provider_earnings", nil)
	if err != nil {
		return nil, err
	}

	var result ProviderEarningsResponse
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to parse earnings response: %w", err)
	}

	return &result, nil
}

// ProviderJobs retrieves active and recent jobs
func (c *DaemonClient) ProviderJobs() (*ProviderJobsResponse, error) {
	resp, err := c.call("provider_jobs", nil)
	if err != nil {
		return nil, err
	}

	var result ProviderJobsResponse
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to parse jobs response: %w", err)
	}

	return &result, nil
}

// ProviderMaintenanceMode sets maintenance mode
func (c *DaemonClient) ProviderMaintenanceMode(enabled bool) error {
	_, err := c.call("provider_maintenance", map[string]bool{"enabled": enabled})
	return err
}
