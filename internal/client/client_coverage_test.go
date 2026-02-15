package client

import (
	"encoding/json"
	"net"
	"os"
	"sync"
	"testing"
	"time"
)

// =============================================================================
// DaemonClient creation and configuration tests
// =============================================================================

func TestNewDaemonClient_CustomPath(t *testing.T) {
	client := NewDaemonClient("/custom/path/daemon.sock")

	if client.SocketPath() != "/custom/path/daemon.sock" {
		t.Errorf("SocketPath mismatch: got %s, want /custom/path/daemon.sock", client.SocketPath())
	}
}

func TestNewDaemonClient_DefaultPath(t *testing.T) {
	client := NewDaemonClient("")

	defaultPath := DefaultSocketPath()
	if client.SocketPath() != defaultPath {
		t.Errorf("SocketPath should use default when empty: got %s, want %s", client.SocketPath(), defaultPath)
	}
}

// =============================================================================
// Connection error handling tests
// =============================================================================

func TestDaemonClient_ConnectToNonexistent(t *testing.T) {
	client := NewDaemonClient("/nonexistent/path/daemon.sock")

	err := client.Connect()
	if err == nil {
		t.Error("Connect to nonexistent socket should fail")
		client.Close()
	}
}

func TestDaemonClient_DoubleClose(t *testing.T) {
	client := NewDaemonClient("/nonexistent/path.sock")

	// First close - not connected
	if err := client.Close(); err != nil {
		t.Errorf("First close should not error: %v", err)
	}

	// Second close - still not connected
	if err := client.Close(); err != nil {
		t.Errorf("Second close should not error: %v", err)
	}
}

func TestDaemonClient_StatusWithoutConnection(t *testing.T) {
	client := NewDaemonClient("/nonexistent/path.sock")

	_, err := client.Status()
	if err == nil {
		t.Error("Status without connection should fail")
	}
}

func TestDaemonClient_DeployWithoutConnection(t *testing.T) {
	client := NewDaemonClient("/nonexistent/path.sock")

	_, err := client.Deploy(&DeployRequest{Image: "nginx:latest"})
	if err == nil {
		t.Error("Deploy without connection should fail")
	}
}

func TestDaemonClient_ListWithoutConnection(t *testing.T) {
	client := NewDaemonClient("/nonexistent/path.sock")

	_, err := client.List()
	if err == nil {
		t.Error("List without connection should fail")
	}
}

func TestDaemonClient_StopWithoutConnection(t *testing.T) {
	client := NewDaemonClient("/nonexistent/path.sock")

	err := client.Stop("container-123")
	if err == nil {
		t.Error("Stop without connection should fail")
	}
}

func TestDaemonClient_DeleteWithoutConnection(t *testing.T) {
	client := NewDaemonClient("/nonexistent/path.sock")

	err := client.Delete("container-123")
	if err == nil {
		t.Error("Delete without connection should fail")
	}
}

func TestDaemonClient_HealthWithoutConnection(t *testing.T) {
	client := NewDaemonClient("/nonexistent/path.sock")

	_, err := client.Health("")
	if err == nil {
		t.Error("Health without connection should fail")
	}
}

func TestDaemonClient_GetLogsWithoutConnection(t *testing.T) {
	client := NewDaemonClient("/nonexistent/path.sock")

	_, err := client.GetLogs("container-123", false, 100)
	if err == nil {
		t.Error("GetLogs without connection should fail")
	}
}

func TestDaemonClient_GetPeersWithoutConnection(t *testing.T) {
	client := NewDaemonClient("/nonexistent/path.sock")

	_, err := client.GetPeers()
	if err == nil {
		t.Error("GetPeers without connection should fail")
	}
}

func TestDaemonClient_TorStartWithoutConnection(t *testing.T) {
	client := NewDaemonClient("/nonexistent/path.sock")

	_, err := client.TorStart()
	if err == nil {
		t.Error("TorStart without connection should fail")
	}
}

func TestDaemonClient_TorStatusWithoutConnection(t *testing.T) {
	client := NewDaemonClient("/nonexistent/path.sock")

	_, err := client.TorStatus()
	if err == nil {
		t.Error("TorStatus without connection should fail")
	}
}

func TestDaemonClient_TorRotateWithoutConnection(t *testing.T) {
	client := NewDaemonClient("/nonexistent/path.sock")

	err := client.TorRotate()
	if err == nil {
		t.Error("TorRotate without connection should fail")
	}
}

func TestDaemonClient_TorStopWithoutConnection(t *testing.T) {
	client := NewDaemonClient("/nonexistent/path.sock")

	err := client.TorStop()
	if err == nil {
		t.Error("TorStop without connection should fail")
	}
}

func TestDaemonClient_ConfigGetWithoutConnection(t *testing.T) {
	client := NewDaemonClient("/nonexistent/path.sock")

	_, err := client.ConfigGet("some_key")
	if err == nil {
		t.Error("ConfigGet without connection should fail")
	}
}

func TestDaemonClient_ConfigSetWithoutConnection(t *testing.T) {
	client := NewDaemonClient("/nonexistent/path.sock")

	err := client.ConfigSet("key", "value")
	if err == nil {
		t.Error("ConfigSet without connection should fail")
	}
}

// =============================================================================
// Extended client methods without connection
// =============================================================================

func TestDaemonClient_ThreatLevelWithoutConnection(t *testing.T) {
	client := NewDaemonClient("/nonexistent/path.sock")

	_, err := client.ThreatLevel()
	if err == nil {
		t.Error("ThreatLevel without connection should fail")
	}
}

func TestDaemonClient_ThreatSignalWithoutConnection(t *testing.T) {
	client := NewDaemonClient("/nonexistent/path.sock")

	err := client.ThreatSignal("network_anomaly", 0.8, "test", "test details")
	if err == nil {
		t.Error("ThreatSignal without connection should fail")
	}
}

func TestDaemonClient_ThreatClearWithoutConnection(t *testing.T) {
	client := NewDaemonClient("/nonexistent/path.sock")

	err := client.ThreatClear("", "", true)
	if err == nil {
		t.Error("ThreatClear without connection should fail")
	}
}

func TestDaemonClient_CloneWithoutConnection(t *testing.T) {
	client := NewDaemonClient("/nonexistent/path.sock")

	_, err := client.Clone(&CloneRequest{SourceID: "src-1"})
	if err == nil {
		t.Error("Clone without connection should fail")
	}
}

func TestDaemonClient_CloneStatusWithoutConnection(t *testing.T) {
	client := NewDaemonClient("/nonexistent/path.sock")

	_, err := client.CloneStatus("clone-1")
	if err == nil {
		t.Error("CloneStatus without connection should fail")
	}
}

func TestDaemonClient_CloneListWithoutConnection(t *testing.T) {
	client := NewDaemonClient("/nonexistent/path.sock")

	_, err := client.CloneList(true, 10)
	if err == nil {
		t.Error("CloneList without connection should fail")
	}
}

func TestDaemonClient_CloneCancelWithoutConnection(t *testing.T) {
	client := NewDaemonClient("/nonexistent/path.sock")

	err := client.CloneCancel("clone-1")
	if err == nil {
		t.Error("CloneCancel without connection should fail")
	}
}

func TestDaemonClient_CloneGetStatsWithoutConnection(t *testing.T) {
	client := NewDaemonClient("/nonexistent/path.sock")

	_, err := client.CloneGetStats()
	if err == nil {
		t.Error("CloneGetStats without connection should fail")
	}
}

func TestDaemonClient_SnapshotCreateWithoutConnection(t *testing.T) {
	client := NewDaemonClient("/nonexistent/path.sock")

	_, err := client.SnapshotCreate(&SnapshotRequest{ContainerID: "cnt-1"})
	if err == nil {
		t.Error("SnapshotCreate without connection should fail")
	}
}

func TestDaemonClient_SnapshotGetWithoutConnection(t *testing.T) {
	client := NewDaemonClient("/nonexistent/path.sock")

	_, err := client.SnapshotGet("snap-1")
	if err == nil {
		t.Error("SnapshotGet without connection should fail")
	}
}

func TestDaemonClient_SnapshotListWithoutConnection(t *testing.T) {
	client := NewDaemonClient("/nonexistent/path.sock")

	_, err := client.SnapshotList("cnt-1")
	if err == nil {
		t.Error("SnapshotList without connection should fail")
	}
}

func TestDaemonClient_SnapshotDeleteWithoutConnection(t *testing.T) {
	client := NewDaemonClient("/nonexistent/path.sock")

	err := client.SnapshotDelete("snap-1")
	if err == nil {
		t.Error("SnapshotDelete without connection should fail")
	}
}

func TestDaemonClient_SnapshotDeleteAllWithoutConnection(t *testing.T) {
	client := NewDaemonClient("/nonexistent/path.sock")

	err := client.SnapshotDeleteAll("cnt-1")
	if err == nil {
		t.Error("SnapshotDeleteAll without connection should fail")
	}
}

func TestDaemonClient_SnapshotRestoreWithoutConnection(t *testing.T) {
	client := NewDaemonClient("/nonexistent/path.sock")

	_, err := client.SnapshotRestore("snap-1", "Americas", true)
	if err == nil {
		t.Error("SnapshotRestore without connection should fail")
	}
}

func TestDaemonClient_SnapshotGetStatsWithoutConnection(t *testing.T) {
	client := NewDaemonClient("/nonexistent/path.sock")

	_, err := client.SnapshotGetStats()
	if err == nil {
		t.Error("SnapshotGetStats without connection should fail")
	}
}

func TestDaemonClient_SnapshotRotateKeyWithoutConnection(t *testing.T) {
	client := NewDaemonClient("/nonexistent/path.sock")

	err := client.SnapshotRotateKey()
	if err == nil {
		t.Error("SnapshotRotateKey without connection should fail")
	}
}

func TestDaemonClient_SnapshotExportWithoutConnection(t *testing.T) {
	client := NewDaemonClient("/nonexistent/path.sock")

	err := client.SnapshotExport("snap-1", "/tmp/snap.tar")
	if err == nil {
		t.Error("SnapshotExport without connection should fail")
	}
}

func TestDaemonClient_SnapshotImportWithoutConnection(t *testing.T) {
	client := NewDaemonClient("/nonexistent/path.sock")

	_, err := client.SnapshotImport("/tmp/snap.tar")
	if err == nil {
		t.Error("SnapshotImport without connection should fail")
	}
}

func TestDaemonClient_RequesterBalanceMapWithoutConnection(t *testing.T) {
	client := NewDaemonClient("/nonexistent/path.sock")

	_, err := client.RequesterBalanceMap()
	if err == nil {
		t.Error("RequesterBalanceMap without connection should fail")
	}
}

// =============================================================================
// Provider methods without connection
// =============================================================================

func TestDaemonClient_ProviderRegisterWithoutConnection(t *testing.T) {
	client := NewDaemonClient("/nonexistent/path.sock")

	_, err := client.ProviderRegister(&ProviderRegisterRequest{
		DeclaredCPU: 8,
	})
	if err == nil {
		t.Error("ProviderRegister without connection should fail")
	}
}

func TestDaemonClient_ProviderStatusWithoutConnection(t *testing.T) {
	client := NewDaemonClient("/nonexistent/path.sock")

	_, err := client.ProviderStatus()
	if err == nil {
		t.Error("ProviderStatus without connection should fail")
	}
}

func TestDaemonClient_ProviderStakeAddWithoutConnection(t *testing.T) {
	client := NewDaemonClient("/nonexistent/path.sock")

	_, err := client.ProviderStakeAdd("1000")
	if err == nil {
		t.Error("ProviderStakeAdd without connection should fail")
	}
}

func TestDaemonClient_ProviderStakeWithdrawWithoutConnection(t *testing.T) {
	client := NewDaemonClient("/nonexistent/path.sock")

	_, err := client.ProviderStakeWithdraw()
	if err == nil {
		t.Error("ProviderStakeWithdraw without connection should fail")
	}
}

func TestDaemonClient_ProviderEarningsWithoutConnection(t *testing.T) {
	client := NewDaemonClient("/nonexistent/path.sock")

	_, err := client.ProviderEarnings()
	if err == nil {
		t.Error("ProviderEarnings without connection should fail")
	}
}

func TestDaemonClient_ProviderJobsWithoutConnection(t *testing.T) {
	client := NewDaemonClient("/nonexistent/path.sock")

	_, err := client.ProviderJobs()
	if err == nil {
		t.Error("ProviderJobs without connection should fail")
	}
}

func TestDaemonClient_ProviderMaintenanceModeWithoutConnection(t *testing.T) {
	client := NewDaemonClient("/nonexistent/path.sock")

	err := client.ProviderMaintenanceMode(true)
	if err == nil {
		t.Error("ProviderMaintenanceMode without connection should fail")
	}
}

// =============================================================================
// Extended struct tests
// =============================================================================

func TestThreatLevelResponse_Structure(t *testing.T) {
	resp := ThreatLevelResponse{
		Score:          0.75,
		Level:          "high",
		Recommendation: "initiate clone",
		Timestamp:      time.Now(),
		ActiveSignals: []ThreatSignalInfo{
			{
				Type:       "network_anomaly",
				Score:      0.8,
				Confidence: 0.9,
				Source:     "internal",
				Details:    "unusual traffic pattern",
			},
		},
	}

	if resp.Score != 0.75 {
		t.Errorf("Score mismatch: got %f", resp.Score)
	}
	if resp.Level != "high" {
		t.Errorf("Level mismatch: got %s", resp.Level)
	}
	if len(resp.ActiveSignals) != 1 {
		t.Errorf("Expected 1 signal, got %d", len(resp.ActiveSignals))
	}
	if resp.ActiveSignals[0].Type != "network_anomaly" {
		t.Errorf("Signal type mismatch: got %s", resp.ActiveSignals[0].Type)
	}
}

func TestCloneRequest_Structure(t *testing.T) {
	req := CloneRequest{
		SourceID:     "dep-1",
		TargetRegion: "Europe",
		Priority:     5,
		Reason:       "threat response",
		IncludeState: true,
	}

	if req.SourceID != "dep-1" {
		t.Errorf("SourceID mismatch: got %s", req.SourceID)
	}
	if req.TargetRegion != "Europe" {
		t.Errorf("TargetRegion mismatch: got %s", req.TargetRegion)
	}
	if req.Priority != 5 {
		t.Errorf("Priority mismatch: got %d", req.Priority)
	}
	if !req.IncludeState {
		t.Error("IncludeState should be true")
	}
}

func TestCloneResponse_Structure(t *testing.T) {
	resp := CloneResponse{
		CloneID:      "clone-1",
		SourceID:     "dep-1",
		TargetID:     "dep-2",
		TargetRegion: "Europe",
		Status:       "completed",
		Priority:     5,
		Reason:       "threat response",
		CreatedAt:    time.Now(),
	}

	if resp.CloneID != "clone-1" {
		t.Errorf("CloneID mismatch: got %s", resp.CloneID)
	}
	if resp.Status != "completed" {
		t.Errorf("Status mismatch: got %s", resp.Status)
	}
}

func TestSnapshotRequest_Structure(t *testing.T) {
	req := SnapshotRequest{
		ContainerID: "cnt-1",
		Type:        "full",
		Metadata:    map[string]string{"version": "1.0"},
	}

	if req.ContainerID != "cnt-1" {
		t.Errorf("ContainerID mismatch: got %s", req.ContainerID)
	}
	if req.Type != "full" {
		t.Errorf("Type mismatch: got %s", req.Type)
	}
}

func TestSnapshotResponse_Structure(t *testing.T) {
	resp := SnapshotResponse{
		ID:          "snap-1",
		ContainerID: "cnt-1",
		Size:        1024,
		StoredSize:  512,
		Compressed:  true,
		Encrypted:   true,
		Checksum:    "abc123",
	}

	if resp.ID != "snap-1" {
		t.Errorf("ID mismatch: got %s", resp.ID)
	}
	if resp.StoredSize != 512 {
		t.Errorf("StoredSize mismatch: got %d", resp.StoredSize)
	}
	if !resp.Compressed {
		t.Error("Compressed should be true")
	}
	if !resp.Encrypted {
		t.Error("Encrypted should be true")
	}
}

func TestSnapshotStats_Structure(t *testing.T) {
	stats := SnapshotStats{
		TotalCount:        10,
		TotalSize:         5120,
		TotalOriginalSize: 10240,
		ContainerCount:    3,
		CompressionRatio:  0.5,
		EncryptedCount:    8,
	}

	if stats.TotalCount != 10 {
		t.Errorf("TotalCount mismatch: got %d", stats.TotalCount)
	}
	if stats.CompressionRatio != 0.5 {
		t.Errorf("CompressionRatio mismatch: got %f", stats.CompressionRatio)
	}
}

func TestCloneStats_Structure(t *testing.T) {
	stats := CloneStats{
		ActiveClones:   2,
		TotalCompleted: 10,
		TotalFailed:    1,
		MaxConcurrent:  5,
	}

	if stats.ActiveClones != 2 {
		t.Errorf("ActiveClones mismatch: got %d", stats.ActiveClones)
	}
	if stats.TotalFailed != 1 {
		t.Errorf("TotalFailed mismatch: got %d", stats.TotalFailed)
	}
}

func TestResourceLimits_ClientStruct(t *testing.T) {
	limits := ResourceLimits{
		CPUShares:   1024,
		MemoryMB:    512,
		StorageMB:   10240,
		NetworkMbps: 100,
	}

	if limits.CPUShares != 1024 {
		t.Errorf("CPUShares mismatch: got %d", limits.CPUShares)
	}
	if limits.MemoryMB != 512 {
		t.Errorf("MemoryMB mismatch: got %d", limits.MemoryMB)
	}
}

func TestHealthResponse_Structure(t *testing.T) {
	resp := HealthResponse{
		Healthy: true,
		UnhealthyContainers: map[string][]int{
			"cnt-1": {1, 2},
		},
	}

	if !resp.Healthy {
		t.Error("Healthy should be true")
	}
	if len(resp.UnhealthyContainers) != 1 {
		t.Errorf("Expected 1 unhealthy container entry, got %d", len(resp.UnhealthyContainers))
	}
}

// =============================================================================
// Provider struct tests
// =============================================================================

func TestProviderRegisterRequest_Structure(t *testing.T) {
	req := ProviderRegisterRequest{
		DeclaredCPU:    8,
		DeclaredMemory: 32768,
		DeclaredDisk:   1024000,
		TargetTier:     "Silver",
		AutoStake:      true,
		GPUEnabled:     true,
		GPUModel:       "RTX 4090",
		GPUCount:       2,
	}

	if req.DeclaredCPU != 8 {
		t.Errorf("DeclaredCPU mismatch: got %d", req.DeclaredCPU)
	}
	if !req.GPUEnabled {
		t.Error("GPUEnabled should be true")
	}
	if req.GPUModel != "RTX 4090" {
		t.Errorf("GPUModel mismatch: got %s", req.GPUModel)
	}
}

func TestProviderStatusResponse_Structure(t *testing.T) {
	resp := ProviderStatusResponse{
		NodeID:            "node-123",
		Status:            "active",
		Tier:              "Silver",
		StakedAmount:      "10000",
		ReputationScore:   850,
		ReputationTier:    "excellent",
		Uptime30Days:      99.9,
		ActiveJobs:        3,
		MaxConcurrentJobs: 10,
	}

	if resp.Tier != "Silver" {
		t.Errorf("Tier mismatch: got %s", resp.Tier)
	}
	if resp.ReputationScore != 850 {
		t.Errorf("ReputationScore mismatch: got %d", resp.ReputationScore)
	}
}

// =============================================================================
// Mock server integration test
// =============================================================================

func TestDaemonClient_ConnectAndCall(t *testing.T) {
	// Use /tmp directly to avoid macOS long temp path exceeding Unix socket limit (108 chars)
	socketPath := "/tmp/mb-test-call.sock"
	os.Remove(socketPath)
	defer os.Remove(socketPath)

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}
	defer listener.Close()

	// Start mock server
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		decoder := json.NewDecoder(conn)
		encoder := json.NewEncoder(conn)

		var req APIRequest
		if err := decoder.Decode(&req); err != nil {
			return
		}

		// Respond based on method
		resp := APIResponse{ID: req.ID}
		switch req.Method {
		case "status":
			result, _ := json.Marshal(StatusResponse{
				NodeID:  "test-node",
				Running: true,
				Port:    9000,
				Version: "0.1.0",
			})
			resp.Result = result
		default:
			resp.Error = "unknown method"
		}

		encoder.Encode(resp)
	}()

	// Create client and connect
	client := NewDaemonClient(socketPath)
	if err := client.Connect(); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	// Already connected - connect again should be no-op
	if err := client.Connect(); err != nil {
		t.Fatalf("Second connect should succeed: %v", err)
	}

	// Call status
	status, err := client.Status()
	if err != nil {
		t.Fatalf("Status call failed: %v", err)
	}

	if status.NodeID != "test-node" {
		t.Errorf("NodeID mismatch: got %s, want test-node", status.NodeID)
	}
	if !status.Running {
		t.Error("Running should be true")
	}
	if status.Port != 9000 {
		t.Errorf("Port mismatch: got %d, want 9000", status.Port)
	}

	wg.Wait()
}

func TestDaemonClient_CallWithError(t *testing.T) {
	// Use /tmp directly to avoid macOS long temp path exceeding Unix socket limit (108 chars)
	socketPath := "/tmp/mb-test-err.sock"
	os.Remove(socketPath)
	defer os.Remove(socketPath)

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}
	defer listener.Close()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		decoder := json.NewDecoder(conn)
		encoder := json.NewEncoder(conn)

		var req APIRequest
		if err := decoder.Decode(&req); err != nil {
			return
		}

		// Always return error
		resp := APIResponse{
			ID:    req.ID,
			Error: "deployment failed: image not found",
		}
		encoder.Encode(resp)
	}()

	client := NewDaemonClient(socketPath)
	if err := client.Connect(); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	_, err = client.Deploy(&DeployRequest{Image: "nonexistent:latest"})
	if err == nil {
		t.Error("Deploy should fail when server returns error")
	}

	wg.Wait()
}
