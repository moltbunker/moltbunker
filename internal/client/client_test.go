package client

import (
	"testing"
)

func TestDefaultSocketPath(t *testing.T) {
	path := DefaultSocketPath()

	if path == "" {
		t.Error("DefaultSocketPath should not be empty")
	}

	// Should end with daemon.sock
	if len(path) < 11 || path[len(path)-11:] != "daemon.sock" {
		t.Errorf("Socket path should end with daemon.sock: %s", path)
	}
}

func TestNewDaemonClient(t *testing.T) {
	tests := []struct {
		name       string
		socketPath string
		wantPath   string
	}{
		{
			name:       "with custom path",
			socketPath: "/custom/path.sock",
			wantPath:   "/custom/path.sock",
		},
		{
			name:       "with empty path (uses default)",
			socketPath: "",
			wantPath:   DefaultSocketPath(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewDaemonClient(tt.socketPath)

			if client == nil {
				t.Fatal("NewDaemonClient returned nil")
			}

			if client.SocketPath() != tt.wantPath {
				t.Errorf("SocketPath mismatch: got %s, want %s", client.SocketPath(), tt.wantPath)
			}
		})
	}
}

func TestDaemonClient_NotConnected(t *testing.T) {
	client := NewDaemonClient("/nonexistent/socket.sock")

	// Close should not error when not connected
	err := client.Close()
	if err != nil {
		t.Errorf("Close should not error when not connected: %v", err)
	}
}

func TestDaemonClient_IsDaemonRunning_NotRunning(t *testing.T) {
	client := NewDaemonClient("/nonexistent/socket.sock")

	running := client.IsDaemonRunning()
	if running {
		t.Error("IsDaemonRunning should return false for nonexistent socket")
	}
}

func TestStatusResponse_Structure(t *testing.T) {
	status := StatusResponse{
		NodeID:     "test-node-id",
		Running:    true,
		Port:       9000,
		NetworkNodes: 5,
		Version:    "0.1.0",
		TorEnabled: true,
		TorAddress: "test.onion",
	}

	if status.NodeID != "test-node-id" {
		t.Error("NodeID not set correctly")
	}

	if !status.Running {
		t.Error("Running should be true")
	}

	if status.Port != 9000 {
		t.Error("Port not set correctly")
	}
}

func TestDeployRequest_Structure(t *testing.T) {
	req := DeployRequest{
		Image:        "nginx:latest",
		TorOnly:      true,
		OnionService: true,
	}

	if req.Image != "nginx:latest" {
		t.Error("Image not set correctly")
	}

	if !req.TorOnly {
		t.Error("TorOnly should be true")
	}

	if !req.OnionService {
		t.Error("OnionService should be true")
	}
}

func TestDeployResponse_Structure(t *testing.T) {
	resp := DeployResponse{
		ContainerID:  "cnt-123",
		OnionAddress: "abc123.onion",
		Status:       "pending",
	}

	if resp.ContainerID != "cnt-123" {
		t.Error("ContainerID not set correctly")
	}

	if resp.OnionAddress != "abc123.onion" {
		t.Error("OnionAddress not set correctly")
	}

	if resp.Status != "pending" {
		t.Error("Status not set correctly")
	}
}

func TestTorStatusResponse_Structure(t *testing.T) {
	status := TorStatusResponse{
		Running:      true,
		OnionAddress: "test.onion",
		CircuitCount: 3,
	}

	if !status.Running {
		t.Error("Running should be true")
	}

	if status.OnionAddress != "test.onion" {
		t.Error("OnionAddress not set correctly")
	}

	if status.CircuitCount != 3 {
		t.Error("CircuitCount not set correctly")
	}
}

func TestPeerInfo_Structure(t *testing.T) {
	peer := PeerInfo{
		ID:      "peer-123",
		Address: "192.168.1.100:9000",
		Region:  "us-east",
	}

	if peer.ID != "peer-123" {
		t.Error("ID not set correctly")
	}

	if peer.Address != "192.168.1.100:9000" {
		t.Error("Address not set correctly")
	}

	if peer.Region != "us-east" {
		t.Error("Region not set correctly")
	}
}

func TestLogsRequest_Structure(t *testing.T) {
	req := LogsRequest{
		ContainerID: "cnt-123",
		Follow:      true,
		Tail:        100,
	}

	if req.ContainerID != "cnt-123" {
		t.Error("ContainerID not set correctly")
	}

	if !req.Follow {
		t.Error("Follow should be true")
	}

	if req.Tail != 100 {
		t.Error("Tail not set correctly")
	}
}

func TestAPIRequest_Structure(t *testing.T) {
	req := APIRequest{
		Method: "status",
		Params: nil,
		ID:     1,
	}

	if req.Method != "status" {
		t.Error("Method not set correctly")
	}

	if req.ID != 1 {
		t.Error("ID not set correctly")
	}
}

func TestAPIResponse_Structure(t *testing.T) {
	resp := APIResponse{
		Error: "test error",
		ID:    1,
	}

	if resp.Error != "test error" {
		t.Error("Error not set correctly")
	}

	if resp.ID != 1 {
		t.Error("ID not set correctly")
	}
}
