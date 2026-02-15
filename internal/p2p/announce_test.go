package p2p

import (
	"crypto/ecdsa"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/moltbunker/moltbunker/pkg/types"
)

func generateTestKey(t *testing.T) (*ecdsa.PrivateKey, common.Address) {
	t.Helper()
	key, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}
	addr := crypto.PubkeyToAddress(key.PublicKey)
	return key, addr
}

func TestCreateAndVerifyAnnounce(t *testing.T) {
	key, addr := generateTestKey(t)
	nodeID := randomNodeID()

	payload, err := CreateAnnouncePayload(nodeID, addr, key)
	if err != nil {
		t.Fatalf("failed to create announce: %v", err)
	}

	recovered, err := VerifyAnnouncePayload(payload)
	if err != nil {
		t.Fatalf("failed to verify announce: %v", err)
	}

	if recovered != addr {
		t.Fatalf("recovered address %s != expected %s", recovered.Hex(), addr.Hex())
	}
}

func TestVerifyAnnounce_WrongKey(t *testing.T) {
	key1, _ := generateTestKey(t)
	_, addr2 := generateTestKey(t)
	nodeID := randomNodeID()

	// Sign with key1 but claim addr2
	payload, err := CreateAnnouncePayload(nodeID, addr2, key1)
	if err != nil {
		t.Fatalf("failed to create announce: %v", err)
	}

	// Verification should fail because key1's address != addr2
	_, err = VerifyAnnouncePayload(payload)
	if err == nil {
		t.Fatal("expected verification to fail with wrong key")
	}
}

func TestVerifyAnnounce_ExpiredTimestamp(t *testing.T) {
	key, addr := generateTestKey(t)
	nodeID := randomNodeID()

	payload, err := CreateAnnouncePayload(nodeID, addr, key)
	if err != nil {
		t.Fatalf("failed to create announce: %v", err)
	}

	// Make timestamp old
	payload.Timestamp = time.Now().Add(-10 * time.Minute).Unix()

	_, err = VerifyAnnouncePayload(payload)
	if err == nil {
		t.Fatal("expected expired timestamp to be rejected")
	}
}

func TestVerifyAnnounce_InvalidSignature(t *testing.T) {
	key, addr := generateTestKey(t)
	nodeID := randomNodeID()

	payload, err := CreateAnnouncePayload(nodeID, addr, key)
	if err != nil {
		t.Fatalf("failed to create announce: %v", err)
	}

	// Corrupt signature
	payload.EthSignature = "00" + payload.EthSignature[2:]

	_, err = VerifyAnnouncePayload(payload)
	if err == nil {
		t.Fatal("expected corrupted signature to be rejected")
	}
}

func TestVerifyAnnounce_BadNonceLength(t *testing.T) {
	payload := &types.AnnouncePayload{
		WalletAddress: "0x0000000000000000000000000000000000000001",
		NodeID:        "deadbeef",
		Timestamp:     time.Now().Unix(),
		Nonce:         "short",
		EthSignature:  "0000",
	}

	_, err := VerifyAnnouncePayload(payload)
	if err == nil {
		t.Fatal("expected bad nonce length to be rejected")
	}
}

func TestVerifyAnnounce_NilPayload(t *testing.T) {
	_, err := VerifyAnnouncePayload(nil)
	if err == nil {
		t.Fatal("expected nil payload to be rejected")
	}
}

func TestCreateAnnounce_NilKey(t *testing.T) {
	nodeID := randomNodeID()
	_, err := CreateAnnouncePayload(nodeID, common.Address{}, nil)
	if err == nil {
		t.Fatal("expected nil key to be rejected")
	}
}

func TestVerifyAnnounce_NodeIDBinding(t *testing.T) {
	key, addr := generateTestKey(t)
	nodeID1 := randomNodeID()

	payload, err := CreateAnnouncePayload(nodeID1, addr, key)
	if err != nil {
		t.Fatalf("failed to create announce: %v", err)
	}

	// Signature is valid for nodeID1. If we tamper with the NodeID field,
	// the message content changes but the signature won't match.
	nodeID2 := randomNodeID()
	payload.NodeID = nodeID2.String()

	_, err = VerifyAnnouncePayload(payload)
	if err == nil {
		t.Fatal("expected tampered nodeID to invalidate signature")
	}
}
