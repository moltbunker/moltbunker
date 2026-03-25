package client

import (
	"encoding/json"
	"fmt"
	"time"
)

// MoltDeployRequest is the client request to deploy a Molt function.
type MoltDeployRequest struct {
	ModuleCID     string            `json:"module_cid,omitempty"`
	MemoryLimitMB uint32            `json:"memory_limit_mb,omitempty"`
	TimeoutMs     int               `json:"timeout_ms,omitempty"`
	MaxInstances  int               `json:"max_instances,omitempty"`
	Environment   map[string]string `json:"environment,omitempty"`
	Owner         string            `json:"owner,omitempty"`
	WasmBytes     []byte            `json:"wasm_bytes,omitempty"`
}

// MoltDeployResponse is the result of a Molt deployment.
type MoltDeployResponse struct {
	DeploymentID string `json:"deployment_id"`
	ModuleCID    string `json:"module_cid"`
	Status       string `json:"status"`
}

// MoltInfo describes a deployed Molt.
type MoltInfo struct {
	ID            string       `json:"id"`
	ModuleCID     string       `json:"module_cid"`
	Status        string       `json:"status"`
	CreatedAt     time.Time    `json:"created_at"`
	Owner         string       `json:"owner,omitempty"`
	MemoryLimitMB uint32       `json:"memory_limit_mb,omitempty"`
	TimeoutMs     int          `json:"timeout_ms,omitempty"`
	Metrics       *MoltMetrics `json:"metrics,omitempty"`
}

// MoltMetrics contains invocation metrics for a Molt deployment.
type MoltMetrics struct {
	TotalInvocations   uint64        `json:"total_invocations"`
	SuccessInvocations uint64        `json:"success_invocations"`
	ErrorInvocations   uint64        `json:"error_invocations"`
	TimeoutInvocations uint64        `json:"timeout_invocations"`
	AvgLatency         time.Duration `json:"avg_latency"`
}

// MoltInvokeRequest is the client request to invoke a Molt.
type MoltInvokeRequest struct {
	DeploymentID string            `json:"deployment_id"`
	Method       string            `json:"method,omitempty"`
	Path         string            `json:"path,omitempty"`
	Headers      map[string]string `json:"headers,omitempty"`
	Body         []byte            `json:"body,omitempty"`
}

// MoltInvokeResponse is the result of a Molt invocation.
type MoltInvokeResponse struct {
	StatusCode int               `json:"status_code"`
	Headers    map[string]string `json:"headers,omitempty"`
	Body       []byte            `json:"body,omitempty"`
	DurationMs int64             `json:"duration_ms"`
	Error      string            `json:"error,omitempty"`
}

// MoltDeploy deploys a WASM function as a Molt.
func (c *DaemonClient) MoltDeploy(req *MoltDeployRequest) (*MoltDeployResponse, error) {
	resp, err := c.call("molt_deploy", req)
	if err != nil {
		return nil, err
	}
	var result MoltDeployResponse
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to parse molt deploy response: %w", err)
	}
	return &result, nil
}

// MoltList lists all deployed Molts.
func (c *DaemonClient) MoltList() ([]MoltInfo, error) {
	resp, err := c.call("molt_list", nil)
	if err != nil {
		return nil, err
	}
	var result []MoltInfo
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to parse molt list: %w", err)
	}
	return result, nil
}

// MoltGet returns details for a single Molt deployment.
func (c *DaemonClient) MoltGet(deploymentID string) (*MoltInfo, error) {
	resp, err := c.call("molt_get", map[string]string{"deployment_id": deploymentID})
	if err != nil {
		return nil, err
	}
	var result MoltInfo
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to parse molt info: %w", err)
	}
	return &result, nil
}

// MoltStop stops a Molt deployment (keeps compiled cache).
func (c *DaemonClient) MoltStop(deploymentID string) error {
	_, err := c.call("molt_stop", map[string]string{"deployment_id": deploymentID})
	return err
}

// MoltDelete removes a Molt deployment entirely.
func (c *DaemonClient) MoltDelete(deploymentID string) error {
	_, err := c.call("molt_delete", map[string]string{"deployment_id": deploymentID})
	return err
}

// MoltInvoke directly invokes a Molt function.
func (c *DaemonClient) MoltInvoke(req *MoltInvokeRequest) (*MoltInvokeResponse, error) {
	resp, err := c.call("molt_invoke", req)
	if err != nil {
		return nil, err
	}
	var result MoltInvokeResponse
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to parse molt invoke response: %w", err)
	}
	return &result, nil
}
