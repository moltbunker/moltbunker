package types

import (
	"encoding/hex"
	"testing"
)

func TestNodeID_String(t *testing.T) {
	var nodeID NodeID
	for i := range nodeID {
		nodeID[i] = byte(i)
	}

	str := nodeID.String()
	if len(str) != 64 { // 32 bytes = 64 hex chars
		t.Errorf("NodeID string length should be 64, got %d", len(str))
	}

	// Verify it's valid hex
	decoded, err := hex.DecodeString(str)
	if err != nil {
		t.Fatalf("NodeID string should be valid hex: %v", err)
	}

	if len(decoded) != 32 {
		t.Errorf("Decoded NodeID should be 32 bytes, got %d", len(decoded))
	}
}

func TestContainerStatus(t *testing.T) {
	statuses := []ContainerStatus{
		ContainerStatusPending,
		ContainerStatusRunning,
		ContainerStatusStopped,
		ContainerStatusFailed,
		ContainerStatusReplicating,
	}

	for _, status := range statuses {
		if len(status) == 0 {
			t.Errorf("Status %s should not be empty", status)
		}
	}
}

func TestNetworkMode(t *testing.T) {
	modes := []NetworkMode{
		NetworkModeClearnet,
		NetworkModeTorOnly,
		NetworkModeHybrid,
		NetworkModeTorExit,
	}

	for _, mode := range modes {
		if len(mode) == 0 {
			t.Errorf("Network mode %s should not be empty", mode)
		}
	}
}

func TestMessageType(t *testing.T) {
	types := []MessageType{
		MessageTypePing,
		MessageTypePong,
		MessageTypeFindNode,
		MessageTypeNodes,
		MessageTypeHealth,
		MessageTypeDeploy,
		MessageTypeBid,
		MessageTypeGossip,
	}

	for _, msgType := range types {
		if len(msgType) == 0 {
			t.Errorf("Message type %s should not be empty", msgType)
		}
	}
}
