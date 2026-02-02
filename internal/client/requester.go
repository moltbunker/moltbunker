package client

import (
	"encoding/json"
	"fmt"
)

// Requester-related request/response types

// RequesterJobsResponse contains the list of requester's jobs
type RequesterJobsResponse struct {
	Jobs []RequesterJobInfo `json:"jobs"`
}

// RequesterJobInfo contains job information for requesters
type RequesterJobInfo struct {
	ID           string `json:"id"`
	DeploymentID string `json:"deployment_id"`
	Image        string `json:"image"`
	Status       string `json:"status"`
	Duration     string `json:"duration"`
	Cost         string `json:"cost"`
	ReplicaCount int    `json:"replica_count"`
	CreatedAt    string `json:"created_at"`
	CompletedAt  string `json:"completed_at,omitempty"`
}

// RequesterOutputsResponse contains encrypted outputs for a job
type RequesterOutputsResponse struct {
	JobID   string            `json:"job_id"`
	Outputs []EncryptedOutput `json:"outputs"`
}

// EncryptedOutput represents an encrypted output from a job
type EncryptedOutput struct {
	DeploymentID   string `json:"deployment_id"`
	Type           string `json:"type"` // "stdout", "stderr", "log", "file"
	EncryptedData  []byte `json:"encrypted_data"`
	ProviderPubKey []byte `json:"provider_pub_key"`
	EncryptedDEK   []byte `json:"encrypted_dek"`
	DEKNonce       []byte `json:"dek_nonce"`
	Timestamp      string `json:"timestamp"`
}

// RequesterBalanceResponse contains balance information
type RequesterBalanceResponse struct {
	Available  string `json:"available"`
	Reserved   string `json:"reserved"`
	TotalSpent string `json:"total_spent"`
}

// RequesterJobs retrieves the list of jobs submitted by the requester
func (c *DaemonClient) RequesterJobs() (*RequesterJobsResponse, error) {
	resp, err := c.call("requester_jobs", nil)
	if err != nil {
		return nil, err
	}

	var result RequesterJobsResponse
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to parse jobs response: %w", err)
	}

	return &result, nil
}

// RequesterOutputs retrieves encrypted outputs for a job
func (c *DaemonClient) RequesterOutputs(jobID string) (*RequesterOutputsResponse, error) {
	resp, err := c.call("requester_outputs", map[string]string{"job_id": jobID})
	if err != nil {
		return nil, err
	}

	var result RequesterOutputsResponse
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to parse outputs response: %w", err)
	}

	return &result, nil
}

// RequesterBalance retrieves the requester's BUNKER balance
func (c *DaemonClient) RequesterBalance() (*RequesterBalanceResponse, error) {
	resp, err := c.call("requester_balance", nil)
	if err != nil {
		return nil, err
	}

	var result RequesterBalanceResponse
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to parse balance response: %w", err)
	}

	return &result, nil
}

// RequesterSubmitJob submits a new job with encryption
func (c *DaemonClient) RequesterSubmitJob(req *DeployRequest, requesterPubKey []byte) (*DeployResponse, error) {
	params := map[string]interface{}{
		"image":              req.Image,
		"resources":         req.Resources,
		"tor_only":          req.TorOnly,
		"onion_service":     req.OnionService,
		"onion_port":        req.OnionPort,
		"requester_pub_key": requesterPubKey,
	}

	resp, err := c.call("requester_submit_job", params)
	if err != nil {
		return nil, err
	}

	var result DeployResponse
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to parse deploy response: %w", err)
	}

	return &result, nil
}
