package daemon

import (
	"encoding/binary"
	"io"
	"testing"
	"time"

	"github.com/moltbunker/moltbunker/internal/runtime"
)

// readAgentFrame reads a single length-prefixed exec-agent frame from r.
// Wire format: [4-byte big-endian total length][1-byte type][payload].
// Mirrors cmd/exec-agent/protocol.go readFrame so tests assert the exact bytes
// the in-container agent would observe.
func readAgentFrame(t *testing.T, r io.Reader) (frameType byte, payload []byte) {
	t.Helper()
	var totalLen uint32
	if err := binary.Read(r, binary.BigEndian, &totalLen); err != nil {
		t.Fatalf("read frame length: %v", err)
	}
	if totalLen < 1 {
		t.Fatalf("invalid frame length: %d", totalLen)
	}
	buf := make([]byte, totalLen)
	if _, err := io.ReadFull(r, buf); err != nil {
		t.Fatalf("read frame body: %v", err)
	}
	return buf[0], buf[1:]
}

// newAgentWriteStream builds an ExecStream in exec-agent mode whose Stdin is the
// write end of an in-memory pipe. The returned reader receives whatever bytes
// WriteData/Resize emit toward the (simulated) in-container exec-agent.
func newAgentWriteStream(t *testing.T) (*ExecStream, *io.PipeReader) {
	t.Helper()
	pr, pw := io.Pipe()
	es := &ExecStream{
		sessionID:     "test-session",
		containerID:   "test-container",
		session:       &runtime.InteractiveSession{Stdin: pw},
		execAgentMode: true,
	}
	return es, pr
}

// frameResult carries a frame read off the pipe by a background goroutine.
type frameResult struct {
	frameType byte
	payload   []byte
}

// readFrameAsync reads one frame in a goroutine so a blocking pipe write in the
// test body unblocks. Returns a channel that yields the parsed frame.
func readFrameAsync(t *testing.T, r io.Reader) <-chan frameResult {
	t.Helper()
	ch := make(chan frameResult, 1)
	go func() {
		ft, payload := readAgentFrame(t, r)
		ch <- frameResult{frameType: ft, payload: payload}
	}()
	return ch
}

func awaitFrame(t *testing.T, ch <-chan frameResult) frameResult {
	t.Helper()
	select {
	case fr := <-ch:
		return fr
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for relayed frame")
		return frameResult{}
	}
}

// TestWriteData_KeyInitPreservesType is the core regression test for BLOCKER 1:
// a KEY_INIT (0x07) control frame sent by the requester (type byte included)
// must reach the exec-agent as a real 0x07 frame — NOT silently rewrapped as a
// DATA (0x01) frame, which previously broke the E2E key handshake.
func TestWriteData_KeyInitPreservesType(t *testing.T) {
	es, pr := newAgentWriteStream(t)

	sessionNonce := []byte("0123456789abcdef0123456789abcdef") // 32-byte nonce
	// Requester sends the FULL frame: [0x07][session_nonce...] (type byte present).
	input := append([]byte{execAgentFrameKeyInit}, sessionNonce...)

	frameCh := readFrameAsync(t, pr)
	if err := es.WriteData(input); err != nil {
		t.Fatalf("WriteData: %v", err)
	}
	fr := awaitFrame(t, frameCh)

	if fr.frameType != execAgentFrameKeyInit {
		t.Fatalf("KEY_INIT type not preserved: got 0x%02x, want 0x%02x (must NOT be rewrapped as DATA 0x01)",
			fr.frameType, execAgentFrameKeyInit)
	}
	if string(fr.payload) != string(sessionNonce) {
		t.Fatalf("KEY_INIT payload corrupted: got %q, want %q", fr.payload, sessionNonce)
	}
}

// TestWriteData_KeyAckPreservesType verifies KEY_ACK (0x08) likewise relays
// with its type byte intact.
func TestWriteData_KeyAckPreservesType(t *testing.T) {
	es, pr := newAgentWriteStream(t)
	input := []byte{execAgentFrameKeyAck}

	frameCh := readFrameAsync(t, pr)
	if err := es.WriteData(input); err != nil {
		t.Fatalf("WriteData: %v", err)
	}
	fr := awaitFrame(t, frameCh)

	if fr.frameType != execAgentFrameKeyAck {
		t.Fatalf("KEY_ACK type not preserved: got 0x%02x, want 0x%02x", fr.frameType, execAgentFrameKeyAck)
	}
	if len(fr.payload) != 0 {
		t.Fatalf("KEY_ACK payload should be empty, got %d bytes", len(fr.payload))
	}
}

// TestWriteData_PlainTerminalWrapsAsData verifies that plain terminal input
// (type byte already stripped by the API layer) is wrapped as a DATA frame.
func TestWriteData_PlainTerminalWrapsAsData(t *testing.T) {
	es, pr := newAgentWriteStream(t)
	// Plain ciphertext keystroke bytes whose first byte is NOT a control type.
	input := []byte{0x00, 0xAB, 0xCD, 0xEF}

	frameCh := readFrameAsync(t, pr)
	if err := es.WriteData(input); err != nil {
		t.Fatalf("WriteData: %v", err)
	}
	fr := awaitFrame(t, frameCh)

	if fr.frameType != execAgentFrameData {
		t.Fatalf("plain terminal input should wrap as DATA 0x01, got 0x%02x", fr.frameType)
	}
	if string(fr.payload) != string(input) {
		t.Fatalf("DATA payload corrupted: got %x, want %x", fr.payload, input)
	}
}

// TestWriteData_ResizeAndCloseControlFrames verifies the remaining requester
// control frame types relay with their type byte preserved.
func TestWriteData_ResizeAndCloseControlFrames(t *testing.T) {
	cases := []struct {
		name      string
		frameType byte
		payload   []byte
	}{
		{"resize", execAgentFrameResize, []byte{0x00, 0x50, 0x00, 0x18}},
		{"ping", execAgentFramePing, nil},
		{"pong", execAgentFramePong, nil},
		{"close", execAgentFrameClose, nil},
		{"error", execAgentFrameError, []byte("boom")},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			es, pr := newAgentWriteStream(t)
			input := append([]byte{tc.frameType}, tc.payload...)

			frameCh := readFrameAsync(t, pr)
			if err := es.WriteData(input); err != nil {
				t.Fatalf("WriteData: %v", err)
			}
			fr := awaitFrame(t, frameCh)

			if fr.frameType != tc.frameType {
				t.Fatalf("%s type not preserved: got 0x%02x, want 0x%02x", tc.name, fr.frameType, tc.frameType)
			}
			if string(fr.payload) != string(tc.payload) {
				t.Fatalf("%s payload corrupted: got %x, want %x", tc.name, fr.payload, tc.payload)
			}
		})
	}
}

// TestWriteData_NonAgentModeIsRaw verifies the plaintext (non-agent) path is
// unchanged: bytes are written to stdin verbatim with no framing, and the
// control-frame heuristic does NOT apply.
func TestWriteData_NonAgentModeIsRaw(t *testing.T) {
	pr, pw := io.Pipe()
	es := &ExecStream{
		sessionID:     "test-session",
		session:       &runtime.InteractiveSession{Stdin: pw},
		execAgentMode: false,
	}
	// First byte equals a control type, but in non-agent mode it must be raw.
	input := append([]byte{execAgentFrameKeyInit}, []byte("hello")...)

	type rawResult struct {
		buf []byte
		err error
	}
	rawCh := make(chan rawResult, 1)
	go func() {
		buf := make([]byte, len(input))
		_, err := io.ReadFull(pr, buf)
		rawCh <- rawResult{buf: buf, err: err}
	}()

	if err := es.WriteData(input); err != nil {
		t.Fatalf("WriteData: %v", err)
	}

	select {
	case res := <-rawCh:
		if res.err != nil {
			t.Fatalf("read raw stdin: %v", res.err)
		}
		if string(res.buf) != string(input) {
			t.Fatalf("non-agent mode must write bytes verbatim: got %x, want %x", res.buf, input)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out reading raw stdin")
	}
}
