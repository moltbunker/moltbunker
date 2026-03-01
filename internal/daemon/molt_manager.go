package daemon

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/moltbunker/moltbunker/internal/logging"
	"github.com/moltbunker/moltbunker/internal/molt"
	"github.com/moltbunker/moltbunker/internal/security"
	"github.com/moltbunker/moltbunker/pkg/types"
)

// MoltDeployment tracks an active Molt deployment on this provider node.
type MoltDeployment struct {
	ID           string                       `json:"id"`
	ModuleCID    string                       `json:"module_cid"`
	Status       types.MoltDeploymentStatus   `json:"status"`
	CreatedAt    time.Time                    `json:"created_at"`
	Owner        string                       `json:"owner,omitempty"`
	Spec         *types.MoltSpec              `json:"spec"`
	Handler      http.Handler                 `json:"-"`
	compiled     *molt.CompiledMolt
}

// MoltManager manages the lifecycle of Molt (WASM) deployments.
// It owns the MoltRuntime from internal/molt and tracks active deployments.
type MoltManager struct {
	runtime       *molt.MoltRuntime
	encryptionMgr *security.DeploymentEncryptionManager // optional — enables E2E encrypted I/O
	deployments   map[string]*MoltDeployment
	mu            sync.RWMutex
	closed        bool
}

// NewMoltManager creates a new MoltManager with the given runtime.
// If runtime is nil, Molt support is disabled and all operations return errors.
func NewMoltManager(runtime *molt.MoltRuntime) *MoltManager {
	return &MoltManager{
		runtime:     runtime,
		deployments: make(map[string]*MoltDeployment),
	}
}

// SetEncryptionManager configures E2E encryption for Molt deployments.
// When set, HTTP handlers will decrypt incoming requests and encrypt responses
// for deployments that have encryption keys registered.
func (m *MoltManager) SetEncryptionManager(em *security.DeploymentEncryptionManager) {
	m.encryptionMgr = em
}

// Available returns true if the Molt runtime is initialized.
func (m *MoltManager) Available() bool {
	return m.runtime != nil
}

// Deploy compiles a WASM module and registers it as a running Molt deployment.
func (m *MoltManager) Deploy(ctx context.Context, deploymentID string, wasmBytes []byte, spec *types.MoltSpec, owner string) (*MoltDeployment, error) {
	if m.runtime == nil {
		return nil, fmt.Errorf("molt runtime not available")
	}

	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return nil, fmt.Errorf("molt manager is closed")
	}
	// Check for duplicate
	if _, exists := m.deployments[deploymentID]; exists {
		m.mu.Unlock()
		return nil, fmt.Errorf("molt deployment %s already exists", deploymentID)
	}
	m.mu.Unlock()

	// Compile the WASM module
	compiled, err := m.runtime.Compile(ctx, wasmBytes, spec.ModuleCID)
	if err != nil {
		return nil, fmt.Errorf("compiling molt %s: %w", deploymentID, err)
	}

	// Create the HTTP handler for this deployment
	handler := molt.NewMoltHTTPHandler(m.runtime, compiled, deploymentID)
	if m.encryptionMgr != nil {
		handler.SetEncryptionManager(m.encryptionMgr)
	}

	deployment := &MoltDeployment{
		ID:        deploymentID,
		ModuleCID: spec.ModuleCID,
		Status:    types.MoltStatusRunning,
		CreatedAt: time.Now(),
		Owner:     owner,
		Spec:      spec,
		Handler:   handler,
		compiled:  compiled,
	}

	m.mu.Lock()
	m.deployments[deploymentID] = deployment
	m.mu.Unlock()

	logging.Info("molt deployed",
		"deployment_id", deploymentID,
		"module_cid", spec.ModuleCID,
		"owner", owner,
	)

	return deployment, nil
}

// Invoke runs a single invocation against a deployed Molt.
func (m *MoltManager) Invoke(ctx context.Context, deploymentID string, invocation molt.MoltInvocation) (*molt.MoltResult, error) {
	if m.runtime == nil {
		return nil, fmt.Errorf("molt runtime not available")
	}

	m.mu.RLock()
	deployment, ok := m.deployments[deploymentID]
	m.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("molt deployment %s not found", deploymentID)
	}

	if deployment.Status != types.MoltStatusRunning {
		return nil, fmt.Errorf("molt deployment %s is %s, not running", deploymentID, deployment.Status)
	}

	invocation.DeploymentID = deploymentID
	return m.runtime.Invoke(ctx, deployment.compiled, invocation)
}

// Stop suspends a Molt deployment (keeps compiled cache).
func (m *MoltManager) Stop(deploymentID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	deployment, ok := m.deployments[deploymentID]
	if !ok {
		return fmt.Errorf("molt deployment %s not found", deploymentID)
	}

	deployment.Status = types.MoltStatusStopped
	logging.Info("molt stopped", "deployment_id", deploymentID)
	return nil
}

// Delete removes a Molt deployment entirely.
func (m *MoltManager) Delete(deploymentID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	deployment, ok := m.deployments[deploymentID]
	if !ok {
		return fmt.Errorf("molt deployment %s not found", deploymentID)
	}

	deployment.Status = types.MoltStatusStopped
	delete(m.deployments, deploymentID)
	logging.Info("molt deleted", "deployment_id", deploymentID)
	return nil
}

// Get returns a Molt deployment by ID.
func (m *MoltManager) Get(deploymentID string) (*MoltDeployment, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	d, ok := m.deployments[deploymentID]
	return d, ok
}

// List returns all active Molt deployments.
func (m *MoltManager) List() []*MoltDeployment {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*MoltDeployment, 0, len(m.deployments))
	for _, d := range m.deployments {
		result = append(result, d)
	}
	return result
}

// GetMetrics returns metrics for a specific Molt deployment.
func (m *MoltManager) GetMetrics(deploymentID string) *types.MoltDeploymentMetrics {
	if m.runtime == nil {
		return nil
	}

	stats := m.runtime.Metrics().GetStats(deploymentID)
	var avgLatency time.Duration
	if stats.TotalInvocations > 0 {
		avgLatency = stats.TotalDuration / time.Duration(stats.TotalInvocations)
	}

	return &types.MoltDeploymentMetrics{
		TotalInvocations:   stats.TotalInvocations,
		SuccessInvocations: stats.SuccessInvocations,
		ErrorInvocations:   stats.ErrorInvocations,
		TimeoutInvocations: stats.TimeoutInvocations,
		AvgLatency:         avgLatency,
	}
}

// Close shuts down the MoltManager and its runtime.
func (m *MoltManager) Close(ctx context.Context) error {
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return nil
	}
	m.closed = true

	// Mark all deployments as stopped
	for _, d := range m.deployments {
		d.Status = types.MoltStatusStopped
	}
	m.deployments = make(map[string]*MoltDeployment)
	m.mu.Unlock()

	if m.runtime != nil {
		if err := m.runtime.Close(ctx); err != nil {
			return fmt.Errorf("closing molt runtime: %w", err)
		}
	}

	logging.Info("molt manager closed")
	return nil
}
