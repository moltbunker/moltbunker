package daemon

import (
	"context"
	cryptorand "crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/moltbunker/moltbunker/internal/molt"
	"github.com/moltbunker/moltbunker/pkg/types"
)

// handleMoltDeploy handles deployment of a Molt serverless function.
func (s *APIServer) handleMoltDeploy(ctx context.Context, req *APIRequest) *APIResponse {
	var params MoltDeployRequest
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return &APIResponse{Error: fmt.Sprintf("invalid molt deploy params: %v", err), ID: req.ID}
	}

	if params.ModuleCID == "" && len(params.WasmBytes) == 0 {
		return &APIResponse{Error: "module_cid or wasm_bytes required", ID: req.ID}
	}

	mm := s.containerManager.MoltManager()
	if mm == nil || !mm.Available() {
		return &APIResponse{Error: "molt runtime not available on this node", ID: req.ID}
	}

	// Build MoltSpec from request
	spec := &types.MoltSpec{
		ModuleCID:     params.ModuleCID,
		MemoryLimitMB: params.MemoryLimitMB,
		TimeoutMs:     params.TimeoutMs,
		MaxInstances:  params.MaxInstances,
		Environment:   params.Environment,
	}

	// Use inline WASM bytes or placeholder for IPFS fetch (Phase 4+)
	wasmBytes := params.WasmBytes
	if len(wasmBytes) == 0 {
		return &APIResponse{Error: "wasm_bytes required (IPFS fetch not yet implemented)", ID: req.ID}
	}

	// Generate deployment ID
	deploymentID := generateMoltDeploymentID()

	// If no CID provided, use deployment ID as cache key
	if spec.ModuleCID == "" {
		spec.ModuleCID = deploymentID
	}

	deployment, err := mm.Deploy(ctx, deploymentID, wasmBytes, spec, params.Owner)
	if err != nil {
		return &APIResponse{Error: fmt.Sprintf("molt deploy failed: %v", err), ID: req.ID}
	}

	return &APIResponse{
		Result: MoltDeployResponse{
			DeploymentID: deployment.ID,
			ModuleCID:    deployment.ModuleCID,
			Status:       string(deployment.Status),
		},
		ID: req.ID,
	}
}

// handleMoltList lists all Molt deployments on this node.
func (s *APIServer) handleMoltList(_ context.Context, req *APIRequest) *APIResponse {
	mm := s.containerManager.MoltManager()
	if mm == nil || !mm.Available() {
		return &APIResponse{Result: []MoltInfo{}, ID: req.ID}
	}

	deployments := mm.List()
	infos := make([]MoltInfo, 0, len(deployments))
	for _, d := range deployments {
		info := MoltInfo{
			ID:        d.ID,
			ModuleCID: d.ModuleCID,
			Status:    string(d.Status),
			CreatedAt: d.CreatedAt,
			Owner:     d.Owner,
			Metrics:   mm.GetMetrics(d.ID),
		}
		if d.Spec != nil {
			info.MemoryLimitMB = d.Spec.MemoryLimitMB
			info.TimeoutMs = d.Spec.TimeoutMs
		}
		infos = append(infos, info)
	}

	return &APIResponse{Result: infos, ID: req.ID}
}

// handleMoltGet returns details for a single Molt deployment.
func (s *APIServer) handleMoltGet(_ context.Context, req *APIRequest) *APIResponse {
	var params struct {
		DeploymentID string `json:"deployment_id"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return &APIResponse{Error: fmt.Sprintf("invalid params: %v", err), ID: req.ID}
	}

	mm := s.containerManager.MoltManager()
	if mm == nil || !mm.Available() {
		return &APIResponse{Error: "molt runtime not available", ID: req.ID}
	}

	d, ok := mm.Get(params.DeploymentID)
	if !ok {
		return &APIResponse{Error: fmt.Sprintf("molt %s not found", params.DeploymentID), ID: req.ID}
	}

	info := MoltInfo{
		ID:        d.ID,
		ModuleCID: d.ModuleCID,
		Status:    string(d.Status),
		CreatedAt: d.CreatedAt,
		Owner:     d.Owner,
		Metrics:   mm.GetMetrics(d.ID),
	}
	if d.Spec != nil {
		info.MemoryLimitMB = d.Spec.MemoryLimitMB
		info.TimeoutMs = d.Spec.TimeoutMs
	}

	return &APIResponse{Result: info, ID: req.ID}
}

// handleMoltStop stops a Molt deployment (keeps compiled cache).
func (s *APIServer) handleMoltStop(_ context.Context, req *APIRequest) *APIResponse {
	var params struct {
		DeploymentID string `json:"deployment_id"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return &APIResponse{Error: fmt.Sprintf("invalid params: %v", err), ID: req.ID}
	}

	mm := s.containerManager.MoltManager()
	if mm == nil || !mm.Available() {
		return &APIResponse{Error: "molt runtime not available", ID: req.ID}
	}

	if err := mm.Stop(params.DeploymentID); err != nil {
		return &APIResponse{Error: fmt.Sprintf("failed to stop molt: %v", err), ID: req.ID}
	}

	return &APIResponse{
		Result: map[string]interface{}{
			"status":        "stopped",
			"deployment_id": params.DeploymentID,
		},
		ID: req.ID,
	}
}

// handleMoltDelete removes a Molt deployment entirely.
func (s *APIServer) handleMoltDelete(_ context.Context, req *APIRequest) *APIResponse {
	var params struct {
		DeploymentID string `json:"deployment_id"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return &APIResponse{Error: fmt.Sprintf("invalid params: %v", err), ID: req.ID}
	}

	mm := s.containerManager.MoltManager()
	if mm == nil || !mm.Available() {
		return &APIResponse{Error: "molt runtime not available", ID: req.ID}
	}

	if err := mm.Delete(params.DeploymentID); err != nil {
		return &APIResponse{Error: fmt.Sprintf("failed to delete molt: %v", err), ID: req.ID}
	}

	return &APIResponse{
		Result: map[string]interface{}{
			"status":        "deleted",
			"deployment_id": params.DeploymentID,
		},
		ID: req.ID,
	}
}

// handleMoltInvoke directly invokes a Molt (for testing/debugging).
func (s *APIServer) handleMoltInvoke(ctx context.Context, req *APIRequest) *APIResponse {
	var params MoltInvokeRequest
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return &APIResponse{Error: fmt.Sprintf("invalid invoke params: %v", err), ID: req.ID}
	}

	if params.DeploymentID == "" {
		return &APIResponse{Error: "deployment_id required", ID: req.ID}
	}

	mm := s.containerManager.MoltManager()
	if mm == nil || !mm.Available() {
		return &APIResponse{Error: "molt runtime not available", ID: req.ID}
	}

	method := params.Method
	if method == "" {
		method = "GET"
	}
	path := params.Path
	if path == "" {
		path = "/"
	}

	result, err := mm.Invoke(ctx, params.DeploymentID, molt.MoltInvocation{
		Method:  method,
		Path:    path,
		Headers: params.Headers,
		Body:    params.Body,
	})
	if err != nil {
		return &APIResponse{Error: fmt.Sprintf("invocation failed: %v", err), ID: req.ID}
	}

	return &APIResponse{
		Result: MoltInvokeResponse{
			StatusCode: result.StatusCode,
			Headers:    result.Headers,
			Body:       result.Body,
			DurationMs: result.Duration.Milliseconds(),
			Error:      result.Error,
		},
		ID: req.ID,
	}
}

// generateMoltDeploymentID generates a unique Molt deployment ID using crypto/rand.
func generateMoltDeploymentID() string {
	var b [16]byte
	if _, err := cryptorand.Read(b[:]); err != nil {
		return fmt.Sprintf("molt-%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("molt-%s", hex.EncodeToString(b[:]))
}
