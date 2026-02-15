package p2p

import "github.com/moltbunker/moltbunker/pkg/types"

// RequiresStake returns true if the given message type requires the sender
// to have a verified on-chain stake. Free messages (discovery, health) can
// be sent by any peer; staked messages (deploy, gossip, exec) require proof
// of stake to prevent spam and Sybil attacks.
func RequiresStake(msgType types.MessageType) bool {
	switch msgType {
	// Free — open for discovery and liveness
	case types.MessageTypePing,
		types.MessageTypePong,
		types.MessageTypeFindNode,
		types.MessageTypeNodes,
		types.MessageTypeHealth,
		types.MessageTypeAnnounce:
		return false

	// Staked — privileged actions requiring economic commitment
	case types.MessageTypeDeploy,
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
		types.MessageTypeExecClose:
		return true

	default:
		// Unknown message types require stake (fail-closed)
		return true
	}
}
