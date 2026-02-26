package daemon

import (
	"testing"

	"github.com/moltbunker/moltbunker/pkg/types"
)

func TestGossipStateValidator_RejectsRemoteSubdomainEntries(t *testing.T) {
	localID := types.NodeID{0x01}
	validator := NewGossipStateValidator(localID)

	remoteID := types.NodeID{0x02}

	// Remote peer trying to inject a subdomain entry — must be rejected
	if validator(remoteID, "subdomain:my-app", "dep-abc123") {
		t.Error("validator should reject remote subdomain entries")
	}

	// Remote peer trying to inject a different subdomain
	if validator(remoteID, "subdomain:victim-app", "dep-evil") {
		t.Error("validator should reject all remote subdomain entries")
	}
}

func TestGossipStateValidator_AcceptsLocalSubdomainEntries(t *testing.T) {
	localID := types.NodeID{0x01}
	validator := NewGossipStateValidator(localID)

	// Local updates (zero sender ID) should always pass
	zeroID := types.NodeID{}
	if !validator(zeroID, "subdomain:my-app", "dep-abc123") {
		t.Error("validator should accept local subdomain entries")
	}
}

func TestGossipStateValidator_AcceptsValidExposeEntries(t *testing.T) {
	localID := types.NodeID{0x01}
	validator := NewGossipStateValidator(localID)

	senderID := types.NodeID{0x02}

	// expose: entry with matching ProviderNodeID should be accepted
	entry := map[string]interface{}{
		"provider_node_id": senderID.String(),
		"deployment_id":    "dep-abc123",
		"provider_addr":    "1.2.3.4:9002",
	}
	if !validator(senderID, "expose:dep-abc123:8080", entry) {
		t.Error("validator should accept expose entry with matching provider ID")
	}
}

func TestGossipStateValidator_RejectsExposeWithWrongProvider(t *testing.T) {
	localID := types.NodeID{0x01}
	validator := NewGossipStateValidator(localID)

	senderID := types.NodeID{0x02}
	differentID := types.NodeID{0x03}

	// expose: entry where ProviderNodeID doesn't match sender — reject
	entry := map[string]interface{}{
		"provider_node_id": differentID.String(),
		"deployment_id":    "dep-abc123",
		"provider_addr":    "1.2.3.4:9002",
	}
	if validator(senderID, "expose:dep-abc123:8080", entry) {
		t.Error("validator should reject expose entry with mismatched provider ID")
	}
}

func TestGossipStateValidator_AcceptsOtherKeys(t *testing.T) {
	localID := types.NodeID{0x01}
	validator := NewGossipStateValidator(localID)

	remoteID := types.NodeID{0x02}

	// Non-expose, non-subdomain keys should be accepted
	if !validator(remoteID, "status:dep-abc123", "running") {
		t.Error("validator should accept non-sensitive keys")
	}
	if !validator(remoteID, "health:node-xyz", "healthy") {
		t.Error("validator should accept non-sensitive keys")
	}
}
