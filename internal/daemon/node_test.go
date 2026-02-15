package daemon

import (
	"testing"
)

// Note: Most daemon tests require integration testing with actual
// libp2p networking. These unit tests focus on configuration and state.

func TestAPIServer_DefaultSocketPath(t *testing.T) {
	path := DefaultSocketPath()

	if path == "" {
		t.Error("DefaultSocketPath should not be empty")
	}

	// Should end with daemon.sock
	if len(path) < 11 || path[len(path)-11:] != "daemon.sock" {
		t.Errorf("Socket path should end with daemon.sock: %s", path)
	}
}

func TestAPIServer_NewAPIServer(t *testing.T) {
	// Note: We can't easily create a Node without real networking
	// This test validates the API server creation logic

	socketPath := "/tmp/test-moltbunker.sock"
	dataDir := "/tmp/test-moltbunker-data"
	server := NewAPIServer(nil, socketPath, dataDir)

	if server == nil {
		t.Fatal("NewAPIServer returned nil")
	}

	if server.SocketPath() != socketPath {
		t.Errorf("Socket path mismatch: got %s, want %s", server.SocketPath(), socketPath)
	}
}

func TestAPIRequest_Structure(t *testing.T) {
	req := APIRequest{
		Method: "status",
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
		ID:    1,
		Error: "",
	}

	if resp.ID != 1 {
		t.Error("ID not set correctly")
	}

	if resp.Error != "" {
		t.Error("Error should be empty")
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

	if status.NetworkNodes != 5 {
		t.Error("NetworkNodes not set correctly")
	}

	if status.Version != "0.1.0" {
		t.Error("Version not set correctly")
	}

	if !status.TorEnabled {
		t.Error("TorEnabled should be true")
	}

	if status.TorAddress != "test.onion" {
		t.Error("TorAddress not set correctly")
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
