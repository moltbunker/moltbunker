package p2p

import (
	"testing"

	"github.com/moltbunker/moltbunker/pkg/types"
)

func TestRequiresStake_FreeMessages(t *testing.T) {
	freeTypes := []types.MessageType{
		types.MessageTypePing,
		types.MessageTypePong,
		types.MessageTypeFindNode,
		types.MessageTypeNodes,
		types.MessageTypeHealth,
		types.MessageTypeAnnounce,
	}

	for _, mt := range freeTypes {
		if RequiresStake(mt) {
			t.Errorf("expected %s to be free (no stake required)", mt)
		}
	}
}

func TestRequiresStake_StakedMessages(t *testing.T) {
	stakedTypes := []types.MessageType{
		types.MessageTypeDeploy,
		types.MessageTypeDeployAck,
		types.MessageTypeBid,
		types.MessageTypeGossip,
		types.MessageTypeContainerStatus,
		types.MessageTypeReplicaSync,
		types.MessageTypeLogs,
		types.MessageTypeStop,
		types.MessageTypeDelete,
		types.MessageTypeExecOpen,
		types.MessageTypeExecData,
		types.MessageTypeExecResize,
		types.MessageTypeExecClose,
	}

	for _, mt := range stakedTypes {
		if !RequiresStake(mt) {
			t.Errorf("expected %s to require stake", mt)
		}
	}
}

func TestRequiresStake_UnknownTypeFailsClosed(t *testing.T) {
	if !RequiresStake("unknown_type") {
		t.Error("expected unknown message type to require stake (fail-closed)")
	}
}
