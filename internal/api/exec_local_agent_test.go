package api

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"testing"
)

// decodeAgentWireFrame parses one exec-agent stdio wire frame
// ([4-byte BE len][type][payload]) from r. This mirrors readFrame in
// cmd/exec-agent/protocol.go so the tests assert against the real wire contract
// the in-container agent parses.
func decodeAgentWireFrame(r io.Reader) (frameType byte, payload []byte, err error) {
	var lenBuf [4]byte
	if _, err := io.ReadFull(r, lenBuf[:]); err != nil {
		return 0, nil, err
	}
	total := binary.BigEndian.Uint32(lenBuf[:])
	if total < 1 {
		return 0, nil, errors.New("invalid frame length")
	}
	buf := make([]byte, total)
	if _, err := io.ReadFull(r, buf); err != nil {
		return 0, nil, err
	}
	return buf[0], buf[1:], nil
}

// TestWriteAgentStdinFrame_DataWireFormat verifies plain terminal DATA is framed
// into the exec-agent stdio wire format with the DATA (0x01) type re-stamped, and
// that the agent's readFrame-equivalent parses it back losslessly. The WS layer
// strips the DATA type byte, so the local relay must re-add it.
func TestWriteAgentStdinFrame_DataWireFormat(t *testing.T) {
	var buf bytes.Buffer
	cipher := []byte("opaque-ciphertext-bytes")
	if err := writeAgentStdinFrame(&buf, WSFrameData, cipher); err != nil {
		t.Fatalf("writeAgentStdinFrame: %v", err)
	}

	// Raw wire layout: [4-byte BE len=1+len(payload)][0x01][payload]
	if got := buf.Len(); got != 4+1+len(cipher) {
		t.Fatalf("wire length = %d, want %d", got, 4+1+len(cipher))
	}
	gotLen := binary.BigEndian.Uint32(buf.Bytes()[:4])
	if gotLen != uint32(1+len(cipher)) {
		t.Fatalf("encoded total length = %d, want %d", gotLen, 1+len(cipher))
	}

	ft, payload, err := decodeAgentWireFrame(&buf)
	if err != nil {
		t.Fatalf("decodeAgentWireFrame: %v", err)
	}
	if ft != WSFrameData {
		t.Fatalf("frame type = 0x%02x, want DATA 0x%02x", ft, WSFrameData)
	}
	if !bytes.Equal(payload, cipher) {
		t.Fatalf("payload = %q, want %q", payload, cipher)
	}
}

// TestWriteAgentStdinFrame_KeyInitTypePreserved is the core regression for the
// local exec-agent path: a KEY_INIT (0x07) control frame must reach the agent's
// stdin AS a 0x07 wire frame — not silently rewrapped as DATA (0x01), which would
// abort the E2E key handshake on single-node ("expose laptop") containers.
func TestWriteAgentStdinFrame_KeyInitTypePreserved(t *testing.T) {
	var buf bytes.Buffer
	sessionNonce := bytes.Repeat([]byte{0xAB}, 32)
	if err := writeAgentStdinFrame(&buf, WSFrameKeyInit, sessionNonce); err != nil {
		t.Fatalf("writeAgentStdinFrame: %v", err)
	}

	ft, payload, err := decodeAgentWireFrame(&buf)
	if err != nil {
		t.Fatalf("decodeAgentWireFrame: %v", err)
	}
	if ft != WSFrameKeyInit {
		t.Fatalf("frame type = 0x%02x, want KEY_INIT 0x%02x (must NOT be rewrapped as DATA)", ft, WSFrameKeyInit)
	}
	if ft == WSFrameData {
		t.Fatal("KEY_INIT was rewrapped as DATA — handshake would break")
	}
	if !bytes.Equal(payload, sessionNonce) {
		t.Fatalf("payload corrupted: got %x, want %x", payload, sessionNonce)
	}
}

// TestWriteAgentStdinFrame_ControlTypesPreserved checks all relayed control
// frame types keep their type byte on the wire.
func TestWriteAgentStdinFrame_ControlTypesPreserved(t *testing.T) {
	for _, ft := range []byte{WSFrameKeyInit, WSFrameKeyAck, WSFrameResize, WSFramePing, WSFramePong, WSFrameClose} {
		var buf bytes.Buffer
		payload := []byte{0x01, 0x02, 0x03, 0x04}
		if err := writeAgentStdinFrame(&buf, ft, payload); err != nil {
			t.Fatalf("writeAgentStdinFrame(0x%02x): %v", ft, err)
		}
		gotType, gotPayload, err := decodeAgentWireFrame(&buf)
		if err != nil {
			t.Fatalf("decode 0x%02x: %v", ft, err)
		}
		if gotType != ft {
			t.Fatalf("type = 0x%02x, want 0x%02x", gotType, ft)
		}
		if !bytes.Equal(gotPayload, payload) {
			t.Fatalf("payload mismatch for 0x%02x", ft)
		}
	}
}

// encodeAgentWireFrame builds an exec-agent stdio wire frame (used to simulate
// the agent's stdout in tests).
func encodeAgentWireFrame(frameType byte, payload []byte) []byte {
	frame := make([]byte, 4+1+len(payload))
	binary.BigEndian.PutUint32(frame[0:4], uint32(1+len(payload)))
	frame[4] = frameType
	copy(frame[5:], payload)
	return frame
}

// TestPumpAgentStdoutFrames_DataAndControl verifies the agent->client direction:
// agent stdout wire frames are translated into top-level WS frames [type][payload]
// with the type byte preserved. DATA -> [0x01][ciphertext], KEY_ACK -> [0x08][...].
func TestPumpAgentStdoutFrames_DataAndControl(t *testing.T) {
	cipher := []byte("encrypted-stdout")
	ack := []byte("ack-payload")

	var stdout bytes.Buffer
	stdout.Write(encodeAgentWireFrame(WSFrameKeyAck, ack))
	stdout.Write(encodeAgentWireFrame(WSFrameData, cipher))
	// EOF after these frames triggers the synthetic CLOSE.

	var frames [][]byte
	var dataBytes int
	pumpAgentStdoutFrames(&stdout, func(b []byte) error {
		cp := make([]byte, len(b))
		copy(cp, b)
		frames = append(frames, cp)
		return nil
	}, func(n int) { dataBytes += n }, "test-session")

	// Expect: KEY_ACK frame, DATA frame, then synthetic CLOSE on EOF.
	if len(frames) != 3 {
		t.Fatalf("got %d WS frames, want 3 (KEY_ACK, DATA, CLOSE)", len(frames))
	}

	// KEY_ACK: top-level type byte preserved.
	if frames[0][0] != WSFrameKeyAck {
		t.Fatalf("frame[0] type = 0x%02x, want KEY_ACK 0x%02x", frames[0][0], WSFrameKeyAck)
	}
	if !bytes.Equal(frames[0][1:], ack) {
		t.Fatalf("KEY_ACK payload mismatch")
	}

	// DATA: [WSFrameData][ciphertext]
	if frames[1][0] != WSFrameData {
		t.Fatalf("frame[1] type = 0x%02x, want DATA 0x%02x", frames[1][0], WSFrameData)
	}
	if !bytes.Equal(frames[1][1:], cipher) {
		t.Fatalf("DATA payload mismatch")
	}

	// CLOSE on EOF.
	if frames[2][0] != WSFrameClose {
		t.Fatalf("frame[2] type = 0x%02x, want CLOSE 0x%02x", frames[2][0], WSFrameClose)
	}

	if dataBytes != len(cipher) {
		t.Fatalf("data byte accounting = %d, want %d", dataBytes, len(cipher))
	}
}

// TestPumpAgentStdoutFrames_AgentCloseStopsPump verifies an explicit agent CLOSE
// frame ends the pump and produces exactly one WS CLOSE (no double-close).
func TestPumpAgentStdoutFrames_AgentCloseStopsPump(t *testing.T) {
	var stdout bytes.Buffer
	stdout.Write(encodeAgentWireFrame(WSFrameData, []byte("x")))
	stdout.Write(encodeAgentWireFrame(WSFrameClose, nil))
	stdout.Write(encodeAgentWireFrame(WSFrameData, []byte("should-not-appear")))

	var frames [][]byte
	pumpAgentStdoutFrames(&stdout, func(b []byte) error {
		cp := make([]byte, len(b))
		copy(cp, b)
		frames = append(frames, cp)
		return nil
	}, nil, "test-session")

	if len(frames) != 2 {
		t.Fatalf("got %d frames, want 2 (DATA, CLOSE)", len(frames))
	}
	if frames[0][0] != WSFrameData {
		t.Fatalf("frame[0] = 0x%02x, want DATA", frames[0][0])
	}
	if frames[1][0] != WSFrameClose {
		t.Fatalf("frame[1] = 0x%02x, want CLOSE", frames[1][0])
	}
}

// TestRoundTrip_StdinThenStdout exercises the full local-path framing contract:
// a WS DATA frame is written to the agent (stdin direction), read back as an
// agent wire frame (what the in-container agent would parse), then a simulated
// agent stdout frame is pumped back out to a WS frame.
func TestRoundTrip_StdinThenStdout(t *testing.T) {
	// Inbound: WS DATA -> agent stdin wire frame -> agent parses DATA + payload.
	var stdin bytes.Buffer
	payload := []byte("keystrokes")
	if err := writeAgentStdinFrame(&stdin, WSFrameData, payload); err != nil {
		t.Fatalf("write stdin: %v", err)
	}
	ft, got, err := decodeAgentWireFrame(&stdin)
	if err != nil {
		t.Fatalf("agent parse stdin: %v", err)
	}
	if ft != WSFrameData || !bytes.Equal(got, payload) {
		t.Fatalf("stdin round trip mismatch: type=0x%02x payload=%q", ft, got)
	}

	// Outbound: agent stdout DATA frame -> WS DATA frame.
	var stdout bytes.Buffer
	stdout.Write(encodeAgentWireFrame(WSFrameData, payload))
	var frames [][]byte
	pumpAgentStdoutFrames(&stdout, func(b []byte) error {
		cp := make([]byte, len(b))
		copy(cp, b)
		frames = append(frames, cp)
		return nil
	}, nil, "rt")
	if len(frames) < 1 || frames[0][0] != WSFrameData || !bytes.Equal(frames[0][1:], payload) {
		t.Fatalf("stdout round trip mismatch: %v", frames)
	}
}
