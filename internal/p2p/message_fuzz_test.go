package p2p

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/moltbunker/moltbunker/pkg/types"
)

// FuzzMessageParsing fuzzes JSON message deserialization to ensure no panics
// on arbitrary input and that valid messages roundtrip correctly.
func FuzzMessageParsing(f *testing.F) {
	// Seed corpus: valid messages of different types
	seeds := []types.Message{
		{
			Type:      types.MessageTypePing,
			From:      types.NodeID{0x01, 0x02, 0x03},
			To:        types.NodeID{0x04, 0x05, 0x06},
			Payload:   []byte(`{"status":"alive"}`),
			Timestamp: time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC),
			Nonce:     [24]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24},
			Version:   types.ProtocolVersion,
		},
		{
			Type:      types.MessageTypePong,
			From:      types.NodeID{0xaa, 0xbb},
			To:        types.NodeID{0xcc, 0xdd},
			Payload:   nil,
			Timestamp: time.Date(2026, 2, 1, 8, 30, 0, 0, time.UTC),
			Version:   types.ProtocolVersion,
		},
		{
			Type:      types.MessageTypeFindNode,
			From:      types.NodeID{0xff},
			To:        types.NodeID{},
			Payload:   []byte(`{"target":"abcdef0123456789"}`),
			Timestamp: time.Now().UTC(),
			Version:   types.ProtocolVersion,
		},
		{
			Type:      types.MessageTypeDeploy,
			From:      types.NodeID{0x10, 0x20, 0x30, 0x40},
			To:        types.NodeID{0x50, 0x60, 0x70, 0x80},
			Payload:   []byte(`{"image_cid":"QmTest123","resources":{"cpu":1000}}`),
			Timestamp: time.Date(2026, 1, 20, 0, 0, 0, 0, time.UTC),
			Signature: []byte{0xde, 0xad, 0xbe, 0xef},
			Version:   types.ProtocolVersion,
		},
		{
			Type:      types.MessageTypeDeployAck,
			Payload:   []byte(`{"container_id":"ctr-123","status":"created"}`),
			Timestamp: time.Date(2026, 1, 21, 0, 0, 0, 0, time.UTC),
			Version:   types.ProtocolVersion,
		},
		{
			Type:      types.MessageTypeGossip,
			Payload:   []byte(`{"key":"container:abc","value":"running","version":5}`),
			Timestamp: time.Date(2026, 1, 22, 0, 0, 0, 0, time.UTC),
			Version:   types.ProtocolVersion,
		},
		{
			Type:      types.MessageTypeContainerStatus,
			Payload:   []byte(`{"container_id":"ctr-456","status":"running","health":{"healthy":true}}`),
			Timestamp: time.Date(2026, 1, 23, 0, 0, 0, 0, time.UTC),
			Version:   types.ProtocolVersion,
		},
		{
			Type:      types.MessageTypeReplicaSync,
			Payload:   []byte(`{"replica_set_id":"rs-789","replicas":[]}`),
			Timestamp: time.Date(2026, 1, 24, 0, 0, 0, 0, time.UTC),
			Version:   types.ProtocolVersion,
		},
		{
			Type:      types.MessageTypeHealth,
			Payload:   []byte(`{"cpu_usage":0.5,"memory_usage":1024,"healthy":true}`),
			Timestamp: time.Date(2026, 1, 25, 0, 0, 0, 0, time.UTC),
			Version:   types.ProtocolVersion,
		},
		{
			Type:      types.MessageTypeStop,
			Payload:   []byte(`{"container_id":"ctr-stop-1"}`),
			Timestamp: time.Date(2026, 1, 26, 0, 0, 0, 0, time.UTC),
			Version:   types.ProtocolVersion,
		},
		{
			Type:      types.MessageTypeDelete,
			Payload:   []byte(`{"container_id":"ctr-del-1"}`),
			Timestamp: time.Date(2026, 1, 27, 0, 0, 0, 0, time.UTC),
			Version:   types.ProtocolVersion,
		},
		{
			Type:      types.MessageTypeLogs,
			Payload:   []byte(`{"container_id":"ctr-logs-1","lines":100}`),
			Timestamp: time.Date(2026, 1, 28, 0, 0, 0, 0, time.UTC),
			Version:   types.ProtocolVersion,
		},
		{
			Type:      types.MessageTypeBid,
			Payload:   []byte(`{"price":1000,"region":"Americas","stake":5000}`),
			Timestamp: time.Date(2026, 1, 29, 0, 0, 0, 0, time.UTC),
			Version:   types.ProtocolVersion,
		},
	}

	for _, seed := range seeds {
		data, err := json.Marshal(seed)
		if err != nil {
			f.Fatalf("failed to marshal seed message: %v", err)
		}
		f.Add(data)
	}

	// Add some edge-case seeds
	f.Add([]byte(`{}`))
	f.Add([]byte(`{"type":"unknown_type","version":999}`))
	f.Add([]byte(`{"type":"ping","payload":null,"version":1}`))
	f.Add([]byte(`null`))
	f.Add([]byte(`[]`))
	f.Add([]byte(`"string"`))
	f.Add([]byte(`0`))
	f.Add([]byte(``))
	f.Add([]byte(`{invalid json`))
	f.Add([]byte(`{"type":"ping","from":"not-a-node-id"}`))
	f.Add([]byte(`{"type":"ping","version":-1}`))
	f.Add([]byte(`{"type":"ping","payload":"base64data","nonce":"short"}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		// Phase 1: Verify no panics on arbitrary JSON input
		var msg types.Message
		err := json.Unmarshal(data, &msg)

		if err != nil {
			// Invalid JSON or incompatible schema -- that is fine, no panic is the goal
			return
		}

		// Phase 2: Verify that successfully parsed messages can be re-serialized
		reEncoded, err := json.Marshal(msg)
		if err != nil {
			t.Fatalf("failed to re-marshal successfully parsed message: %v", err)
		}

		// Phase 3: Verify roundtrip -- re-deserialize and compare
		var msg2 types.Message
		if err := json.Unmarshal(reEncoded, &msg2); err != nil {
			t.Fatalf("failed to unmarshal re-encoded message: %v", err)
		}

		// Verify structural equality of the roundtripped message
		if msg.Type != msg2.Type {
			t.Errorf("Type mismatch after roundtrip: %q vs %q", msg.Type, msg2.Type)
		}
		if msg.From != msg2.From {
			t.Errorf("From mismatch after roundtrip")
		}
		if msg.To != msg2.To {
			t.Errorf("To mismatch after roundtrip")
		}
		if !bytes.Equal(msg.Payload, msg2.Payload) {
			t.Errorf("Payload mismatch after roundtrip")
		}
		if msg.Nonce != msg2.Nonce {
			t.Errorf("Nonce mismatch after roundtrip")
		}
		if !bytes.Equal(msg.Signature, msg2.Signature) {
			t.Errorf("Signature mismatch after roundtrip")
		}
		if msg.Version != msg2.Version {
			t.Errorf("Version mismatch after roundtrip: %d vs %d", msg.Version, msg2.Version)
		}

		// Phase 4: Verify SignableBytes does not panic
		signable := msg.SignableBytes()
		if signable != nil {
			// If SignableBytes succeeded, verify it is deterministic
			signable2 := msg.SignableBytes()
			if !bytes.Equal(signable, signable2) {
				t.Errorf("SignableBytes is not deterministic")
			}
		}
	})
}
