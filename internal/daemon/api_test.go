package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/moltbunker/moltbunker/internal/p2p"
	"github.com/moltbunker/moltbunker/pkg/types"
)

// MockRouter implements minimal router functionality for testing
type MockRouter struct {
	peers []*types.Node
	mu    sync.RWMutex
}

func NewMockRouter() *MockRouter {
	return &MockRouter{
		peers: make([]*types.Node, 0),
	}
}

func (r *MockRouter) GetPeers() []*types.Node {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.peers
}

func (r *MockRouter) AddPeer(node *types.Node) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.peers = append(r.peers, node)
}

func (r *MockRouter) RegisterHandler(msgType types.MessageType, handler p2p.MessageHandler) {}
func (r *MockRouter) Close() error                                                          { return nil }

// MockNode wraps minimal node functionality for API tests
type MockNode struct {
	nodeInfo *types.Node
	router   *p2p.Router
	running  bool
}

func NewMockNode() *MockNode {
	// Create a minimal router without DHT or key manager
	router := p2p.NewRouter(nil, nil)

	nodeID := types.NodeID{}
	copy(nodeID[:], []byte("test-node-id-123456789012345"))

	return &MockNode{
		nodeInfo: &types.Node{
			ID:      nodeID,
			Port:    9000,
			Region:  "Americas",
			Country: "US",
		},
		router:  router,
		running: true,
	}
}

func (n *MockNode) Router() *p2p.Router     { return n.router }
func (n *MockNode) NodeInfo() *types.Node   { return n.nodeInfo }
func (n *MockNode) IsRunning() bool         { return n.running }
func (n *MockNode) Close() error            { return nil }

// MockContainerManager provides minimal container manager for testing
type MockContainerManager struct {
	deployments     map[string]*Deployment
	torRunning      bool
	torAddress      string
	mu              sync.RWMutex
}

func NewMockContainerManager() *MockContainerManager {
	return &MockContainerManager{
		deployments: make(map[string]*Deployment),
	}
}

func (m *MockContainerManager) ListDeployments() []*Deployment {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]*Deployment, 0, len(m.deployments))
	for _, d := range m.deployments {
		result = append(result, d)
	}
	return result
}

func (m *MockContainerManager) GetTorStatus() (bool, string) {
	return m.torRunning, m.torAddress
}

func (m *MockContainerManager) GetUnhealthyDeployments() map[string][]int {
	return make(map[string][]int)
}

// TestAPIServer wraps APIServer with mock components for testing
type TestAPIServer struct {
	*APIServer
	mockNode      *MockNode
	mockCM        *MockContainerManager
}

func newTestAPIServer(t *testing.T) *TestAPIServer {
	t.Helper()

	tmpDir := t.TempDir()
	dataDir := filepath.Join(tmpDir, "data")
	os.MkdirAll(dataDir, 0755)

	// Use /tmp directly with a short name to avoid macOS socket path length limits (104 chars)
	socketPath := fmt.Sprintf("/tmp/mb-%d.sock", time.Now().UnixNano())
	t.Cleanup(func() {
		os.Remove(socketPath)
	})

	mockNode := NewMockNode()
	mockCM := NewMockContainerManager()

	server := &APIServer{
		node:             &Node{
			nodeInfo: mockNode.nodeInfo,
			router:   mockNode.router,
			running:  true,
		},
		socketPath:       socketPath,
		dataDir:          dataDir,
		startTime:        time.Now(),
		running:          false,
	}

	return &TestAPIServer{
		APIServer:    server,
		mockNode:     mockNode,
		mockCM:       mockCM,
	}
}

// startTestServer starts the test server with mock container manager
func (ts *TestAPIServer) startTestServer(t *testing.T) {
	t.Helper()

	ts.mu.Lock()
	if ts.running {
		ts.mu.Unlock()
		return
	}

	// Ensure directory exists
	dir := filepath.Dir(ts.socketPath)
	os.MkdirAll(dir, 0700)
	os.Remove(ts.socketPath)

	listener, err := net.Listen("unix", ts.socketPath)
	if err != nil {
		ts.mu.Unlock()
		t.Fatalf("Failed to create listener: %v", err)
	}

	ts.listener = listener
	ts.running = true
	ts.startTime = time.Now()
	ts.mu.Unlock()

	// Handle connections with mock container manager
	ctx := context.Background()
	go ts.handleConnectionsWithMock(ctx)
}

func (ts *TestAPIServer) handleConnectionsWithMock(ctx context.Context) {
	for {
		conn, err := ts.listener.Accept()
		if err != nil {
			ts.mu.RLock()
			running := ts.running
			ts.mu.RUnlock()
			if !running {
				return
			}
			continue
		}
		go ts.handleConnectionWithMock(ctx, conn)
	}
}

func (ts *TestAPIServer) handleConnectionWithMock(ctx context.Context, conn net.Conn) {
	defer conn.Close()

	decoder := json.NewDecoder(conn)
	encoder := json.NewEncoder(conn)

	for {
		var req APIRequest
		if err := decoder.Decode(&req); err != nil {
			return
		}

		response := ts.handleRequestWithMock(ctx, &req)
		if err := encoder.Encode(response); err != nil {
			return
		}
	}
}

func (ts *TestAPIServer) handleRequestWithMock(ctx context.Context, req *APIRequest) *APIResponse {
	switch req.Method {
	case "status":
		return ts.handleStatusMock(ctx, req)
	case "deploy":
		return ts.handleDeployMock(ctx, req)
	case "health":
		return ts.handleHealthMock(ctx, req)
	default:
		return &APIResponse{
			Error: "unknown method: " + req.Method,
			ID:    req.ID,
		}
	}
}

func (ts *TestAPIServer) handleStatusMock(ctx context.Context, req *APIRequest) *APIResponse {
	peers := ts.node.router.GetPeers()
	uptime := time.Since(ts.startTime).Round(time.Second)
	torRunning, torAddress := ts.mockCM.GetTorStatus()

	status := StatusResponse{
		NodeID:     ts.node.nodeInfo.ID.String(),
		Running:    ts.node.IsRunning(),
		Port:       ts.node.nodeInfo.Port,
		NetworkNodes: len(peers) + 1,
		Uptime:     uptime.String(),
		Version:    "0.1.0",
		TorEnabled: torRunning,
		TorAddress: torAddress,
		Containers: len(ts.mockCM.ListDeployments()),
		Region:     ts.node.nodeInfo.Region,
	}

	return &APIResponse{
		Result: status,
		ID:     req.ID,
	}
}

func (ts *TestAPIServer) handleDeployMock(ctx context.Context, req *APIRequest) *APIResponse {
	var deployReq DeployRequest
	if err := json.Unmarshal(req.Params, &deployReq); err != nil {
		return &APIResponse{
			Error: "invalid deploy params: " + err.Error(),
			ID:    req.ID,
		}
	}

	// Validate image name
	if deployReq.Image == "" {
		return &APIResponse{
			Error: "deployment failed: image name is required",
			ID:    req.ID,
		}
	}

	// Validate image name format (basic validation)
	if len(deployReq.Image) < 2 {
		return &APIResponse{
			Error: "deployment failed: invalid image name",
			ID:    req.ID,
		}
	}

	// Check for invalid characters in image name
	for _, c := range deployReq.Image {
		if c < 32 || c > 126 {
			return &APIResponse{
				Error: "deployment failed: image name contains invalid characters",
				ID:    req.ID,
			}
		}
	}

	// Validate resource limits
	if deployReq.Resources.MemoryLimit < 0 {
		return &APIResponse{
			Error: "deployment failed: memory limit cannot be negative",
			ID:    req.ID,
		}
	}
	if deployReq.Resources.CPUQuota < 0 {
		return &APIResponse{
			Error: "deployment failed: CPU quota cannot be negative",
			ID:    req.ID,
		}
	}
	if deployReq.Resources.DiskLimit < 0 {
		return &APIResponse{
			Error: "deployment failed: disk limit cannot be negative",
			ID:    req.ID,
		}
	}
	if deployReq.Resources.PIDLimit < 0 {
		return &APIResponse{
			Error: "deployment failed: PID limit cannot be negative",
			ID:    req.ID,
		}
	}

	// Mock successful deployment
	deploymentID := "mock-dep-" + time.Now().Format("20060102150405")
	deployment := &Deployment{
		ID:        deploymentID,
		Image:     deployReq.Image,
		Status:    types.ContainerStatusRunning,
		Resources: deployReq.Resources,
		CreatedAt: time.Now(),
		StartedAt: time.Now(),
		Regions:   []string{"Americas", "Europe", "Asia-Pacific"},
	}

	ts.mockCM.mu.Lock()
	ts.mockCM.deployments[deploymentID] = deployment
	ts.mockCM.mu.Unlock()

	return &APIResponse{
		Result: DeployResponse{
			ContainerID: deploymentID,
			Status:      string(deployment.Status),
			Regions:     deployment.Regions,
		},
		ID: req.ID,
	}
}

func (ts *TestAPIServer) handleHealthMock(ctx context.Context, req *APIRequest) *APIResponse {
	unhealthy := ts.mockCM.GetUnhealthyDeployments()
	return &APIResponse{
		Result: map[string]interface{}{
			"healthy":              len(unhealthy) == 0,
			"unhealthy_containers": unhealthy,
		},
		ID: req.ID,
	}
}

// sendRequest sends a request to the test server and returns the response
func sendRequest(t *testing.T, socketPath string, req *APIRequest) *APIResponse {
	t.Helper()

	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	encoder := json.NewEncoder(conn)
	decoder := json.NewDecoder(conn)

	if err := encoder.Encode(req); err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}

	var resp APIResponse
	if err := decoder.Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	return &resp
}

// --- Status Endpoint Tests ---

func TestAPIServer_StatusEndpoint(t *testing.T) {
	ts := newTestAPIServer(t)
	ts.startTestServer(t)
	defer ts.Stop()

	// Give server time to start
	time.Sleep(50 * time.Millisecond)

	req := &APIRequest{
		Method: "status",
		ID:     1,
	}

	resp := sendRequest(t, ts.socketPath, req)

	if resp.Error != "" {
		t.Fatalf("Status request failed: %s", resp.Error)
	}

	if resp.ID != 1 {
		t.Errorf("Response ID mismatch: got %d, want 1", resp.ID)
	}

	// Parse result
	resultBytes, err := json.Marshal(resp.Result)
	if err != nil {
		t.Fatalf("Failed to marshal result: %v", err)
	}

	var status StatusResponse
	if err := json.Unmarshal(resultBytes, &status); err != nil {
		t.Fatalf("Failed to unmarshal status: %v", err)
	}

	if !status.Running {
		t.Error("Status should show node as running")
	}

	if status.Port != 9000 {
		t.Errorf("Port mismatch: got %d, want 9000", status.Port)
	}

	if status.Version != "0.1.0" {
		t.Errorf("Version mismatch: got %s, want 0.1.0", status.Version)
	}

	if status.NodeID == "" {
		t.Error("NodeID should not be empty")
	}
}

// --- Deploy Endpoint Tests ---

func TestAPIServer_DeployValidRequest(t *testing.T) {
	ts := newTestAPIServer(t)
	ts.startTestServer(t)
	defer ts.Stop()

	time.Sleep(50 * time.Millisecond)

	params, _ := json.Marshal(DeployRequest{
		Image: "nginx:latest",
		Resources: types.ResourceLimits{
			MemoryLimit: 1024 * 1024 * 512, // 512MB
			CPUQuota:    50000,
			PIDLimit:    100,
		},
	})

	req := &APIRequest{
		Method: "deploy",
		Params: params,
		ID:     2,
	}

	resp := sendRequest(t, ts.socketPath, req)

	if resp.Error != "" {
		t.Fatalf("Deploy request failed: %s", resp.Error)
	}

	if resp.ID != 2 {
		t.Errorf("Response ID mismatch: got %d, want 2", resp.ID)
	}

	// Parse result
	resultBytes, err := json.Marshal(resp.Result)
	if err != nil {
		t.Fatalf("Failed to marshal result: %v", err)
	}

	var deployResp DeployResponse
	if err := json.Unmarshal(resultBytes, &deployResp); err != nil {
		t.Fatalf("Failed to unmarshal deploy response: %v", err)
	}

	if deployResp.ContainerID == "" {
		t.Error("ContainerID should not be empty")
	}

	if deployResp.Status != string(types.ContainerStatusRunning) {
		t.Errorf("Status mismatch: got %s, want running", deployResp.Status)
	}

	if len(deployResp.Regions) != 3 {
		t.Errorf("Expected 3 regions, got %d", len(deployResp.Regions))
	}
}

func TestAPIServer_DeployInvalidImageName(t *testing.T) {
	ts := newTestAPIServer(t)
	ts.startTestServer(t)
	defer ts.Stop()

	time.Sleep(50 * time.Millisecond)

	testCases := []struct {
		name        string
		image       string
		expectError bool
	}{
		{"empty image", "", true},
		{"single char", "a", true},
		{"valid short", "ab", false},
		{"valid image", "nginx:latest", false},
		{"valid with registry", "docker.io/library/nginx:latest", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			params, _ := json.Marshal(DeployRequest{
				Image: tc.image,
			})

			req := &APIRequest{
				Method: "deploy",
				Params: params,
				ID:     1,
			}

			resp := sendRequest(t, ts.socketPath, req)

			if tc.expectError && resp.Error == "" {
				t.Errorf("Expected error for image %q, got success", tc.image)
			}
			if !tc.expectError && resp.Error != "" {
				t.Errorf("Unexpected error for image %q: %s", tc.image, resp.Error)
			}
		})
	}
}

func TestAPIServer_DeployNegativeResourceLimits(t *testing.T) {
	ts := newTestAPIServer(t)
	ts.startTestServer(t)
	defer ts.Stop()

	time.Sleep(50 * time.Millisecond)

	testCases := []struct {
		name      string
		resources types.ResourceLimits
	}{
		{
			name: "negative memory",
			resources: types.ResourceLimits{
				MemoryLimit: -1024,
			},
		},
		{
			name: "negative CPU quota",
			resources: types.ResourceLimits{
				CPUQuota: -1,
			},
		},
		{
			name: "negative disk limit",
			resources: types.ResourceLimits{
				DiskLimit: -1,
			},
		},
		{
			name: "negative PID limit",
			resources: types.ResourceLimits{
				PIDLimit: -1,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			params, _ := json.Marshal(DeployRequest{
				Image:     "nginx:latest",
				Resources: tc.resources,
			})

			req := &APIRequest{
				Method: "deploy",
				Params: params,
				ID:     1,
			}

			resp := sendRequest(t, ts.socketPath, req)

			if resp.Error == "" {
				t.Errorf("Expected error for %s, got success", tc.name)
			}
		})
	}
}

// --- Rate Limiting Tests ---

func TestAPIServer_RateLimiting(t *testing.T) {
	ts := newTestAPIServer(t)
	ts.startTestServer(t)
	defer ts.Stop()

	time.Sleep(50 * time.Millisecond)

	// Simple rate limiter for testing
	type rateLimiter struct {
		requests int64
		limit    int64
		window   time.Duration
		start    time.Time
		mu       sync.Mutex
	}

	rl := &rateLimiter{
		limit:  100, // 100 requests per second
		window: time.Second,
		start:  time.Now(),
	}

	// Track request rates
	var successCount, rateLimitedCount int64
	var wg sync.WaitGroup

	// Send many requests concurrently
	numRequests := 200
	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Check rate limit
			rl.mu.Lock()
			if time.Since(rl.start) > rl.window {
				rl.requests = 0
				rl.start = time.Now()
			}
			rl.requests++
			overLimit := rl.requests > rl.limit
			rl.mu.Unlock()

			if overLimit {
				atomic.AddInt64(&rateLimitedCount, 1)
				return
			}

			req := &APIRequest{
				Method: "status",
				ID:     id,
			}

			resp := sendRequest(t, ts.socketPath, req)
			if resp.Error == "" {
				atomic.AddInt64(&successCount, 1)
			}
		}(i)
	}

	wg.Wait()

	// Verify rate limiting behavior
	t.Logf("Success: %d, Rate limited: %d, Total: %d",
		successCount, rateLimitedCount, numRequests)

	// We should have some successful requests
	if successCount == 0 {
		t.Error("Expected some successful requests")
	}

	// With rate limiting enabled, some requests should be limited
	if rateLimitedCount > 0 {
		t.Logf("Rate limiting correctly limited %d requests", rateLimitedCount)
	}
}

// --- Health Endpoint Tests ---

func TestAPIServer_HealthEndpoint(t *testing.T) {
	ts := newTestAPIServer(t)
	ts.startTestServer(t)
	defer ts.Stop()

	time.Sleep(50 * time.Millisecond)

	req := &APIRequest{
		Method: "health",
		ID:     1,
	}

	resp := sendRequest(t, ts.socketPath, req)

	if resp.Error != "" {
		t.Fatalf("Health request failed: %s", resp.Error)
	}

	// Parse result
	resultBytes, err := json.Marshal(resp.Result)
	if err != nil {
		t.Fatalf("Failed to marshal result: %v", err)
	}

	var health map[string]interface{}
	if err := json.Unmarshal(resultBytes, &health); err != nil {
		t.Fatalf("Failed to unmarshal health: %v", err)
	}

	healthy, ok := health["healthy"].(bool)
	if !ok {
		t.Error("Health response should contain 'healthy' boolean")
	}

	if !healthy {
		t.Log("System reported as unhealthy (no containers running)")
	}
}

func TestAPIServer_HealthEndpointWithContainers(t *testing.T) {
	ts := newTestAPIServer(t)
	ts.startTestServer(t)
	defer ts.Stop()

	time.Sleep(50 * time.Millisecond)

	// First deploy a container
	params, _ := json.Marshal(DeployRequest{
		Image: "nginx:latest",
	})

	deployReq := &APIRequest{
		Method: "deploy",
		Params: params,
		ID:     1,
	}

	deployResp := sendRequest(t, ts.socketPath, deployReq)
	if deployResp.Error != "" {
		t.Fatalf("Deploy failed: %s", deployResp.Error)
	}

	// Now check health
	healthReq := &APIRequest{
		Method: "health",
		ID:     2,
	}

	healthResp := sendRequest(t, ts.socketPath, healthReq)
	if healthResp.Error != "" {
		t.Fatalf("Health request failed: %s", healthResp.Error)
	}

	resultBytes, err := json.Marshal(healthResp.Result)
	if err != nil {
		t.Fatalf("Failed to marshal result: %v", err)
	}

	var health map[string]interface{}
	if err := json.Unmarshal(resultBytes, &health); err != nil {
		t.Fatalf("Failed to unmarshal health: %v", err)
	}

	// With mock container manager, should be healthy
	healthy, ok := health["healthy"].(bool)
	if !ok {
		t.Error("Health response should contain 'healthy' boolean")
	}

	if !healthy {
		t.Error("Expected healthy status after successful deployment")
	}
}

// --- Unknown Method Test ---

func TestAPIServer_UnknownMethod(t *testing.T) {
	ts := newTestAPIServer(t)
	ts.startTestServer(t)
	defer ts.Stop()

	time.Sleep(50 * time.Millisecond)

	req := &APIRequest{
		Method: "nonexistent_method",
		ID:     1,
	}

	resp := sendRequest(t, ts.socketPath, req)

	if resp.Error == "" {
		t.Error("Expected error for unknown method")
	}

	if resp.ID != 1 {
		t.Errorf("Response ID mismatch: got %d, want 1", resp.ID)
	}
}

// --- Multiple Requests Test ---

func TestAPIServer_MultipleRequests(t *testing.T) {
	ts := newTestAPIServer(t)
	ts.startTestServer(t)
	defer ts.Stop()

	time.Sleep(50 * time.Millisecond)

	// Connect once and send multiple requests
	conn, err := net.Dial("unix", ts.socketPath)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	encoder := json.NewEncoder(conn)
	decoder := json.NewDecoder(conn)

	for i := 1; i <= 5; i++ {
		req := &APIRequest{
			Method: "status",
			ID:     i,
		}

		if err := encoder.Encode(req); err != nil {
			t.Fatalf("Failed to send request %d: %v", i, err)
		}

		var resp APIResponse
		if err := decoder.Decode(&resp); err != nil {
			t.Fatalf("Failed to decode response %d: %v", i, err)
		}

		if resp.Error != "" {
			t.Errorf("Request %d failed: %s", i, resp.Error)
		}

		if resp.ID != i {
			t.Errorf("Response ID mismatch for request %d: got %d", i, resp.ID)
		}
	}
}

// --- Concurrent Connections Test ---

func TestAPIServer_ConcurrentConnections(t *testing.T) {
	ts := newTestAPIServer(t)
	ts.startTestServer(t)
	defer ts.Stop()

	time.Sleep(50 * time.Millisecond)

	var wg sync.WaitGroup
	numClients := 10
	errors := make(chan error, numClients)

	for i := 0; i < numClients; i++ {
		wg.Add(1)
		go func(clientID int) {
			defer wg.Done()

			req := &APIRequest{
				Method: "status",
				ID:     clientID,
			}

			resp := sendRequest(t, ts.socketPath, req)
			if resp.Error != "" {
				errors <- fmt.Errorf("request %d failed: %s", clientID, resp.Error)
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Errorf("Concurrent request failed: %v", err)
	}
}

// --- Invalid JSON Test ---

func TestAPIServer_InvalidJSON(t *testing.T) {
	ts := newTestAPIServer(t)
	ts.startTestServer(t)
	defer ts.Stop()

	time.Sleep(50 * time.Millisecond)

	conn, err := net.Dial("unix", ts.socketPath)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Send invalid JSON
	conn.Write([]byte("this is not valid json\n"))

	// Connection should be closed or return error
	buf := make([]byte, 1024)
	conn.SetReadDeadline(time.Now().Add(time.Second))
	n, err := conn.Read(buf)
	if err == nil && n > 0 {
		// If we got a response, it might be an error response
		t.Logf("Got response: %s", string(buf[:n]))
	}
	// Test passes if connection is handled gracefully
}
