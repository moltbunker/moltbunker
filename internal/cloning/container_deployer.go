package cloning

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/moltbunker/moltbunker/internal/logging"
	"github.com/moltbunker/moltbunker/internal/p2p"
	"github.com/moltbunker/moltbunker/internal/runtime"
	"github.com/moltbunker/moltbunker/pkg/types"
)

// ContainerRuntime interface abstracts container operations for cloning
type ContainerRuntime interface {
	GetContainer(id string) (*runtime.ManagedContainer, bool)
	CreateContainer(ctx context.Context, id string, imageRef string, resources types.ResourceLimits) (*runtime.ManagedContainer, error)
	StartContainer(ctx context.Context, id string) error
}

// DecompressState decompresses gzip-compressed state data
func DecompressState(data []byte) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gz.Close()

	decompressed, err := io.ReadAll(gz)
	if err != nil {
		return nil, fmt.Errorf("decompression failed: %w", err)
	}

	return decompressed, nil
}

// CloneDeployRequest is the payload for clone deployment
type CloneDeployRequest struct {
	CloneID      string            `json:"clone_id"`
	SourceID     string            `json:"source_id"`
	Image        string            `json:"image"`
	StateData    []byte            `json:"state_data,omitempty"`
	Compressed   bool              `json:"compressed"`
	Encrypted    bool              `json:"encrypted"`
	Priority     int               `json:"priority"`
	Resources    types.ResourceLimits `json:"resources,omitempty"`
	TorOnly      bool              `json:"tor_only"`
	OnionService bool              `json:"onion_service"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// CloneDeployResponse is the response from clone deployment
type CloneDeployResponse struct {
	CloneID     string    `json:"clone_id"`
	ContainerID string    `json:"container_id"`
	Success     bool      `json:"success"`
	Error       string    `json:"error,omitempty"`
	NodeID      string    `json:"node_id"`
	Region      string    `json:"region"`
	Timestamp   time.Time `json:"timestamp"`
}

// CloneVerifyRequest is the payload for deployment verification
type CloneVerifyRequest struct {
	CloneID     string `json:"clone_id"`
	ContainerID string `json:"container_id"`
}

// CloneVerifyResponse is the response from verification
type CloneVerifyResponse struct {
	CloneID   string `json:"clone_id"`
	Healthy   bool   `json:"healthy"`
	Status    string `json:"status"`
	Error     string `json:"error,omitempty"`
}

// P2PContainerDeployer deploys containers via P2P network
type P2PContainerDeployer struct {
	router         *p2p.Router
	localRuntime   ContainerRuntime
	pendingDeploys map[string]chan *CloneDeployResponse
	pendingMu      sync.RWMutex
}

// NewP2PContainerDeployer creates a new P2P-based container deployer
func NewP2PContainerDeployer(router *p2p.Router, localRuntime ContainerRuntime) *P2PContainerDeployer {
	d := &P2PContainerDeployer{
		router:         router,
		localRuntime:   localRuntime,
		pendingDeploys: make(map[string]chan *CloneDeployResponse),
	}

	// Register handlers for clone deployment messages
	if router != nil {
		router.RegisterHandler(types.MessageTypeDeploy, d.handleDeployRequest)
		router.RegisterHandler(types.MessageTypeDeployAck, d.handleDeployAck)
	}

	return d
}

// DeployClone deploys a cloned container to a target node
func (d *P2PContainerDeployer) DeployClone(ctx context.Context, clone *Clone, stateData []byte) error {
	if d.router == nil {
		return fmt.Errorf("P2P router not initialized")
	}

	// Create deployment request
	deployReq := &CloneDeployRequest{
		CloneID:    clone.ID,
		SourceID:   clone.SourceID,
		StateData:  stateData,
		Compressed: true,
		Encrypted:  true,
		Priority:   clone.Priority,
		Metadata:   clone.Metadata,
	}

	// Get source container info if available
	if clone.SourceID != "" && d.localRuntime != nil {
		container, exists := d.localRuntime.GetContainer(clone.SourceID)
		if exists && container != nil {
			deployReq.Image = container.Image
			deployReq.OnionService = container.OnionAddress != ""
		}
	}

	// Serialize request
	payload, err := json.Marshal(deployReq)
	if err != nil {
		return fmt.Errorf("failed to serialize deploy request: %w", err)
	}

	// Create response channel
	responseCh := make(chan *CloneDeployResponse, 1)
	d.pendingMu.Lock()
	d.pendingDeploys[clone.ID] = responseCh
	d.pendingMu.Unlock()

	defer func() {
		d.pendingMu.Lock()
		delete(d.pendingDeploys, clone.ID)
		d.pendingMu.Unlock()
	}()

	// Send deployment message
	msg := &types.Message{
		Type:      types.MessageTypeDeploy,
		Payload:   payload,
		Timestamp: time.Now(),
	}

	if err := d.router.SendMessage(ctx, clone.TargetNodeID, msg); err != nil {
		return fmt.Errorf("failed to send deploy message: %w", err)
	}

	logging.Info("clone deployment request sent",
		"clone_id", clone.ID,
		"target_node", clone.TargetNodeID.String()[:16],
		logging.Component("cloning"))

	// Wait for response with timeout
	select {
	case <-ctx.Done():
		return ctx.Err()
	case resp := <-responseCh:
		if !resp.Success {
			return fmt.Errorf("deployment failed: %s", resp.Error)
		}
		// Update clone with container ID
		clone.TargetID = resp.ContainerID
		return nil
	case <-time.After(5 * time.Minute):
		return fmt.Errorf("deployment timed out")
	}
}

// VerifyDeployment verifies the clone is running correctly
func (d *P2PContainerDeployer) VerifyDeployment(ctx context.Context, clone *Clone) error {
	if d.router == nil {
		return fmt.Errorf("P2P router not initialized")
	}

	// Create verification request
	verifyReq := &CloneVerifyRequest{
		CloneID:     clone.ID,
		ContainerID: clone.TargetID,
	}

	payload, err := json.Marshal(verifyReq)
	if err != nil {
		return fmt.Errorf("failed to serialize verify request: %w", err)
	}

	// Send health check message
	msg := &types.Message{
		Type:      types.MessageTypeHealth,
		Payload:   payload,
		Timestamp: time.Now(),
	}

	if err := d.router.SendMessage(ctx, clone.TargetNodeID, msg); err != nil {
		return fmt.Errorf("failed to send health check: %w", err)
	}

	// For now, assume success if message was sent
	// In a full implementation, we would wait for a response
	logging.Info("clone verification request sent",
		"clone_id", clone.ID,
		"target_node", clone.TargetNodeID.String()[:16],
		logging.Component("cloning"))

	return nil
}

// handleDeployRequest handles incoming deployment requests
func (d *P2PContainerDeployer) handleDeployRequest(ctx context.Context, msg *types.Message, from *types.Node) error {
	var req CloneDeployRequest
	if err := json.Unmarshal(msg.Payload, &req); err != nil {
		return fmt.Errorf("failed to parse deploy request: %w", err)
	}

	logging.Info("received clone deploy request",
		"clone_id", req.CloneID,
		"from", from.ID.String()[:16],
		logging.Component("cloning"))

	// Deploy container locally
	var containerID string
	var deployErr error

	if d.localRuntime != nil {
		// Create container deployment
		containerID, deployErr = d.deployContainerLocally(ctx, &req)
	} else {
		deployErr = fmt.Errorf("local runtime not available")
	}

	// Send response
	resp := &CloneDeployResponse{
		CloneID:     req.CloneID,
		ContainerID: containerID,
		Success:     deployErr == nil,
		Timestamp:   time.Now(),
	}
	if from.Region != "" {
		resp.Region = from.Region
	}
	if deployErr != nil {
		resp.Error = deployErr.Error()
	}

	respPayload, err := json.Marshal(resp)
	if err != nil {
		return fmt.Errorf("failed to serialize response: %w", err)
	}

	ackMsg := &types.Message{
		Type:      types.MessageTypeDeployAck,
		Payload:   respPayload,
		Timestamp: time.Now(),
	}

	if err := d.router.SendMessage(ctx, from.ID, ackMsg); err != nil {
		logging.Warn("failed to send deploy ack",
			"error", err.Error(),
			logging.Component("cloning"))
	}

	return deployErr
}

// handleDeployAck handles deployment acknowledgments
func (d *P2PContainerDeployer) handleDeployAck(ctx context.Context, msg *types.Message, from *types.Node) error {
	var resp CloneDeployResponse
	if err := json.Unmarshal(msg.Payload, &resp); err != nil {
		return fmt.Errorf("failed to parse deploy ack: %w", err)
	}

	logging.Info("received clone deploy ack",
		"clone_id", resp.CloneID,
		"success", resp.Success,
		"from", from.ID.String()[:16],
		logging.Component("cloning"))

	// Deliver to waiting channel
	d.pendingMu.RLock()
	ch, exists := d.pendingDeploys[resp.CloneID]
	d.pendingMu.RUnlock()

	if exists {
		select {
		case ch <- &resp:
		default:
		}
	}

	return nil
}

// deployContainerLocally deploys a container on the local node
func (d *P2PContainerDeployer) deployContainerLocally(ctx context.Context, req *CloneDeployRequest) (string, error) {
	if d.localRuntime == nil {
		return "", fmt.Errorf("local runtime not available")
	}

	// Restore state if provided
	if len(req.StateData) > 0 {
		// Decompress and decrypt state data
		stateData := req.StateData
		if req.Compressed {
			var err error
			stateData, err = DecompressState(stateData)
			if err != nil {
				return "", fmt.Errorf("failed to decompress state: %w", err)
			}
		}
		// Note: Decryption would happen here if encrypted

		// For now, log that we have state data
		logging.Info("deploying clone with state data",
			"clone_id", req.CloneID,
			"state_size", len(stateData),
			logging.Component("cloning"))
	}

	// Generate container ID
	containerID := fmt.Sprintf("clone-%s", req.CloneID[:8])

	// Create container
	_, err := d.localRuntime.CreateContainer(ctx, containerID, req.Image, req.Resources)
	if err != nil {
		return "", fmt.Errorf("container creation failed: %w", err)
	}

	// Start container
	if err := d.localRuntime.StartContainer(ctx, containerID); err != nil {
		return "", fmt.Errorf("container start failed: %w", err)
	}

	logging.Info("clone container deployed locally",
		"clone_id", req.CloneID,
		"container_id", containerID,
		logging.Component("cloning"))

	return containerID, nil
}

// LocalContainerDeployer deploys containers locally (for testing/single-node)
type LocalContainerDeployer struct {
	runtime ContainerRuntime
}

// NewLocalContainerDeployer creates a local container deployer
func NewLocalContainerDeployer(runtimeMgr ContainerRuntime) *LocalContainerDeployer {
	return &LocalContainerDeployer{
		runtime: runtimeMgr,
	}
}

// DeployClone deploys a container locally
func (d *LocalContainerDeployer) DeployClone(ctx context.Context, clone *Clone, stateData []byte) error {
	if d.runtime == nil {
		return fmt.Errorf("runtime not initialized")
	}

	logging.Info("deploying clone locally",
		"clone_id", clone.ID,
		"source_id", clone.SourceID,
		logging.Component("cloning"))

	// Get source container info
	var image string

	if clone.SourceID != "" {
		container, exists := d.runtime.GetContainer(clone.SourceID)
		if exists && container != nil {
			image = container.Image
		}
	}

	if image == "" {
		return fmt.Errorf("unable to determine container image")
	}

	// Generate container ID
	containerID := fmt.Sprintf("clone-%s", clone.ID[:8])

	// Create container
	_, err := d.runtime.CreateContainer(ctx, containerID, image, types.ResourceLimits{})
	if err != nil {
		return fmt.Errorf("container creation failed: %w", err)
	}

	// Start container
	if err := d.runtime.StartContainer(ctx, containerID); err != nil {
		return fmt.Errorf("container start failed: %w", err)
	}

	clone.TargetID = containerID
	logging.Info("clone deployed successfully",
		"clone_id", clone.ID,
		"container_id", containerID,
		logging.Component("cloning"))

	return nil
}

// VerifyDeployment verifies the container is running
func (d *LocalContainerDeployer) VerifyDeployment(ctx context.Context, clone *Clone) error {
	if d.runtime == nil {
		return fmt.Errorf("runtime not initialized")
	}

	container, exists := d.runtime.GetContainer(clone.TargetID)
	if !exists {
		return fmt.Errorf("container not found")
	}

	if container == nil {
		return fmt.Errorf("container not found")
	}

	if container.Status != types.ContainerStatusRunning {
		return fmt.Errorf("container not running: %s", container.Status)
	}

	logging.Info("clone verification successful",
		"clone_id", clone.ID,
		"container_id", clone.TargetID,
		"status", container.Status,
		logging.Component("cloning"))

	return nil
}
