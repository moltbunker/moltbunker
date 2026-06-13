package api

import (
	"bytes"
	"testing"

	"github.com/moltbunker/moltbunker/internal/client"
)

// TestBuildDaemonDeployRequest_ExecEnvelopeFlows verifies the E2E exec envelope
// fields and exposed ports are copied from the REST DeployRequest into the
// daemon client DeployRequest. Before RUN-01 the browser path dropped these,
// forcing every REST-deployed container into plaintext exec.
func TestBuildDaemonDeployRequest_ExecEnvelopeFlows(t *testing.T) {
	req := &DeployRequest{
		Image:                    "nginx:latest",
		EncryptedExecKey:         []byte{0xaa, 0xbb, 0xcc},
		ExecKeyNonce:             []byte{0x01, 0x02},
		RequesterEphemeralPubKey: bytes.Repeat([]byte{0x07}, 32),
		DeployNonce:              "deadbeefcafef00d",
		ExposePorts:              []client.ExposedPort{{ContainerPort: 8080, Protocol: "tcp"}},
	}

	got := buildDaemonDeployRequest(req, "0xWallet")

	if got.Owner != "0xWallet" {
		t.Errorf("Owner = %q, want 0xWallet", got.Owner)
	}
	if !bytes.Equal(got.EncryptedExecKey, req.EncryptedExecKey) {
		t.Errorf("EncryptedExecKey = %v, want %v", got.EncryptedExecKey, req.EncryptedExecKey)
	}
	if !bytes.Equal(got.ExecKeyNonce, req.ExecKeyNonce) {
		t.Errorf("ExecKeyNonce not forwarded")
	}
	if !bytes.Equal(got.RequesterEphemeralPubKey, req.RequesterEphemeralPubKey) {
		t.Errorf("RequesterEphemeralPubKey not forwarded")
	}
	if got.DeployNonce != "deadbeefcafef00d" {
		t.Errorf("DeployNonce = %q, want deadbeefcafef00d", got.DeployNonce)
	}
	if len(got.ExposePorts) != 1 || got.ExposePorts[0].ContainerPort != 8080 {
		t.Errorf("ExposePorts = %+v, want [{8080 tcp}]", got.ExposePorts)
	}
}

// TestBuildDaemonDeployRequest_OmittedEnvelopeIsLegacy verifies that a request
// with no exec envelope produces empty fields, so the daemon injects no
// exec-agent — identical to pre-SEC-10 behavior. No regression for older
// browsers / API clients.
func TestBuildDaemonDeployRequest_OmittedEnvelopeIsLegacy(t *testing.T) {
	got := buildDaemonDeployRequest(&DeployRequest{Image: "alpine"}, "")
	if len(got.EncryptedExecKey) != 0 || len(got.ExecKeyNonce) != 0 ||
		len(got.RequesterEphemeralPubKey) != 0 || got.DeployNonce != "" {
		t.Errorf("omitted exec envelope must stay empty, got %+v", got)
	}
	if len(got.ExposePorts) != 0 {
		t.Errorf("omitted ExposePorts must stay empty, got %+v", got.ExposePorts)
	}
}

// TestBuildDeployResponse_EchoesExecFields verifies the response surfaces
// ExecAgentEnabled + DeployNonce so the browser knows whether to run the exec
// handshake.
func TestBuildDeployResponse_EchoesExecFields(t *testing.T) {
	result := &client.DeployResponse{
		ContainerID:      "dep-123",
		Status:           "running",
		ExecAgentEnabled: true,
		DeployNonce:      "abc123",
	}
	got := buildDeployResponse(result)
	if got.ContainerID != "dep-123" {
		t.Errorf("ContainerID = %q, want dep-123", got.ContainerID)
	}
	if !got.ExecAgentEnabled {
		t.Error("ExecAgentEnabled should be echoed as true")
	}
	if got.DeployNonce != "abc123" {
		t.Errorf("DeployNonce = %q, want abc123", got.DeployNonce)
	}
}
