package commands

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"io"
	"sync/atomic"
	"testing"

	"golang.org/x/crypto/hkdf"
)

// agentDirBit mirrors the exec-agent's direction tag (cmd/exec-agent/session.go
// nonceDirBit): the agent encrypts with the high bit of nonce[0] CLEAR
// (direction 0). The CLI sets it (direction 1, execNonceDirBit). The two MUST
// differ so the directions never collide on a (key, nonce) pair.
const agentDirBit = byte(0x80)

// agentSession is a faithful, in-test reimplementation of the exec-agent's
// Session type (cmd/exec-agent/session.go). The real agent file is
// //go:build linux and cannot be imported from the CLI package (which builds on
// darwin too), so we mirror it here byte-for-byte to prove wire interop:
//
//	session_key = HKDF-SHA256(exec_key, salt=session_nonce, info="session-key")
//	Encrypt: nonce = [8B BE counter starting at 1, high bit CLEARED for direction 0][4B random]
//	         out = Seal(nonce,nonce,pt,nil)
//	Decrypt: data = [12B nonce][ct+tag]; Open(nil,nonce,ct,nil)
//
// If this struct and execSession diverge, the interop round-trip tests fail.
type agentSession struct {
	aead           cipher.AEAD
	encryptCounter atomic.Uint64
}

func newAgentSession(execKey, sessionNonce []byte) (*agentSession, error) {
	r := hkdf.New(sha256.New, execKey, sessionNonce, []byte("session-key"))
	sessionKey := make([]byte, 32)
	if _, err := io.ReadFull(r, sessionKey); err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(sessionKey)
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &agentSession{aead: aead}, nil
}

func (s *agentSession) Encrypt(plaintext []byte) ([]byte, error) {
	nonce := make([]byte, 12)
	counter := s.encryptCounter.Add(1)
	binary.BigEndian.PutUint64(nonce[:8], counter)
	nonce[0] &^= agentDirBit // agent direction = 0
	if _, err := rand.Read(nonce[8:]); err != nil {
		return nil, err
	}
	return s.aead.Seal(nonce, nonce, plaintext, nil), nil
}

func (s *agentSession) Decrypt(data []byte) ([]byte, error) {
	nonce := data[:12]
	ct := data[12:]
	return s.aead.Open(nil, nonce, ct, nil)
}

func testKeyAndNonce(t *testing.T) (execKey, sessionNonce []byte) {
	t.Helper()
	execKey = make([]byte, execSessionKeySize)
	if _, err := rand.Read(execKey); err != nil {
		t.Fatalf("rand execKey: %v", err)
	}
	sessionNonce = make([]byte, execSessionKeySize)
	if _, err := rand.Read(sessionNonce); err != nil {
		t.Fatalf("rand sessionNonce: %v", err)
	}
	return execKey, sessionNonce
}

// TestExecSession_CLIEncrypt_AgentDecrypt proves the CLI's outgoing ciphertext
// is decryptable by the exec-agent's session format.
func TestExecSession_CLIEncrypt_AgentDecrypt(t *testing.T) {
	execKey, sessionNonce := testKeyAndNonce(t)

	cli, err := newExecSession(execKey, sessionNonce)
	if err != nil {
		t.Fatalf("newExecSession: %v", err)
	}
	agent, err := newAgentSession(execKey, sessionNonce)
	if err != nil {
		t.Fatalf("newAgentSession: %v", err)
	}

	for _, pt := range [][]byte{
		[]byte("ls -la\r"),
		[]byte(""),
		bytes.Repeat([]byte{0xAB}, 4096),
		{0x00},
	} {
		ct, err := cli.Encrypt(pt)
		if err != nil {
			t.Fatalf("CLI Encrypt: %v", err)
		}
		got, err := agent.Decrypt(ct)
		if err != nil {
			t.Fatalf("agent Decrypt: %v", err)
		}
		if !bytes.Equal(got, pt) {
			t.Fatalf("agent decrypt mismatch: got %x want %x", got, pt)
		}
	}
}

// TestExecSession_AgentEncrypt_CLIDecrypt proves the exec-agent's outgoing
// ciphertext is decryptable by the CLI session format.
func TestExecSession_AgentEncrypt_CLIDecrypt(t *testing.T) {
	execKey, sessionNonce := testKeyAndNonce(t)

	cli, err := newExecSession(execKey, sessionNonce)
	if err != nil {
		t.Fatalf("newExecSession: %v", err)
	}
	agent, err := newAgentSession(execKey, sessionNonce)
	if err != nil {
		t.Fatalf("newAgentSession: %v", err)
	}

	for _, pt := range [][]byte{
		[]byte("total 0\r\ndrwxr-xr-x\r\n"),
		bytes.Repeat([]byte("x"), 1000),
		{},
	} {
		ct, err := agent.Encrypt(pt)
		if err != nil {
			t.Fatalf("agent Encrypt: %v", err)
		}
		got, err := cli.Decrypt(ct)
		if err != nil {
			t.Fatalf("CLI Decrypt: %v", err)
		}
		if !bytes.Equal(got, pt) {
			t.Fatalf("CLI decrypt mismatch: got %x want %x", got, pt)
		}
	}
}

// TestExecSession_WireFormat verifies the exact on-wire layout the agent
// expects: 12-byte nonce prefix ([8B BE counter starting at 1][4B random])
// followed by ciphertext+tag, and a monotonically increasing counter.
func TestExecSession_WireFormat(t *testing.T) {
	execKey, sessionNonce := testKeyAndNonce(t)
	cli, err := newExecSession(execKey, sessionNonce)
	if err != nil {
		t.Fatalf("newExecSession: %v", err)
	}

	plaintext := []byte("hello")
	first, err := cli.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	// 12B nonce + len(plaintext) + 16B tag.
	if want := execGCMNonceSize + len(plaintext) + execGCMTagSize; len(first) != want {
		t.Fatalf("ciphertext length = %d, want %d", len(first), want)
	}
	// The CLI stamps the direction bit (high bit of nonce[0]) into the counter
	// region, so it is set in the wire counter and must be masked off to read
	// the logical counter value.
	if first[0]&execNonceDirBit == 0 {
		t.Fatalf("CLI nonce[0] direction bit not set: 0x%02x", first[0])
	}
	if c := binary.BigEndian.Uint64(first[:8]) &^ (uint64(execNonceDirBit) << 56); c != 1 {
		t.Fatalf("first counter = %d, want 1", c)
	}

	second, err := cli.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt 2: %v", err)
	}
	if c := binary.BigEndian.Uint64(second[:8]) &^ (uint64(execNonceDirBit) << 56); c != 2 {
		t.Fatalf("second counter = %d, want 2", c)
	}
}

// TestExecSession_DirectionBitDiffers proves the two directions of the same
// session can never produce the same 12-byte nonce for equal counters: the CLI
// (client) stamps the high bit of nonce[0], the agent clears it. This is the
// cross-direction (key, nonce) collision guarantee — it must hold regardless of
// the random nonce tail. Round-trip tests pass even with a polarity mismatch
// (Decrypt reads the nonce off the wire), so this distinctness check is the only
// test that actually guards the domain separation.
func TestExecSession_DirectionBitDiffers(t *testing.T) {
	execKey, sessionNonce := testKeyAndNonce(t)
	cli, err := newExecSession(execKey, sessionNonce)
	if err != nil {
		t.Fatalf("newExecSession: %v", err)
	}
	agent, err := newAgentSession(execKey, sessionNonce)
	if err != nil {
		t.Fatalf("newAgentSession: %v", err)
	}

	pt := []byte("collision-check")
	for i := 1; i <= 1000; i++ {
		cliCT, err := cli.Encrypt(pt)
		if err != nil {
			t.Fatalf("CLI Encrypt: %v", err)
		}
		agentCT, err := agent.Encrypt(pt)
		if err != nil {
			t.Fatalf("agent Encrypt: %v", err)
		}
		// Counters advance in lockstep (both at i), so the only nonce difference
		// we are guaranteed is the direction bit.
		cliNonce := cliCT[:execGCMNonceSize]
		agentNonce := agentCT[:execGCMNonceSize]

		if cliNonce[0]&execNonceDirBit == 0 {
			t.Fatalf("iter %d: CLI direction bit not set: nonce[0]=0x%02x", i, cliNonce[0])
		}
		if agentNonce[0]&execNonceDirBit != 0 {
			t.Fatalf("iter %d: agent direction bit set: nonce[0]=0x%02x", i, agentNonce[0])
		}
		// The high bit alone guarantees the nonces differ even if the random
		// tails happened to collide.
		if bytes.Equal(cliNonce, agentNonce) {
			t.Fatalf("iter %d: CLI and agent produced identical nonce %x", i, cliNonce)
		}
	}
}

// TestExecSession_NonceDistinctSameCounterAndRandom isolates the direction bit
// from the random tail: it forces both directions to use the same counter and
// the same 4 random bytes, then asserts the full 12-byte nonces still differ.
// This is the worst case the random tail cannot save us from, and the one the
// direction bit is designed to cover.
func TestExecSession_NonceDistinctSameCounterAndRandom(t *testing.T) {
	const counter uint64 = 1
	randTail := []byte{0xDE, 0xAD, 0xBE, 0xEF}

	cliNonce := make([]byte, execGCMNonceSize)
	binary.BigEndian.PutUint64(cliNonce[:8], counter)
	cliNonce[0] |= execNonceDirBit // CLI/client direction = 1
	copy(cliNonce[8:], randTail)

	agentNonce := make([]byte, execGCMNonceSize)
	binary.BigEndian.PutUint64(agentNonce[:8], counter)
	agentNonce[0] &^= agentDirBit // agent direction = 0
	copy(agentNonce[8:], randTail)

	if bytes.Equal(cliNonce, agentNonce) {
		t.Fatalf("identical nonces for same counter+random: %x", cliNonce)
	}
	if cliNonce[0] != 0x80 {
		t.Fatalf("CLI nonce[0] = 0x%02x, want 0x80 (counter high byte 0 + dir bit)", cliNonce[0])
	}
	if agentNonce[0] != 0x00 {
		t.Fatalf("agent nonce[0] = 0x%02x, want 0x00", agentNonce[0])
	}
}

// TestExecSession_WrongKeyFails ensures a different exec_key (or session_nonce)
// yields a key that cannot decrypt the peer's ciphertext (authenticated).
func TestExecSession_WrongKeyFails(t *testing.T) {
	execKey, sessionNonce := testKeyAndNonce(t)
	cli, err := newExecSession(execKey, sessionNonce)
	if err != nil {
		t.Fatalf("newExecSession: %v", err)
	}

	otherKey := make([]byte, execSessionKeySize)
	if _, err := rand.Read(otherKey); err != nil {
		t.Fatalf("rand: %v", err)
	}
	wrong, err := newExecSession(otherKey, sessionNonce)
	if err != nil {
		t.Fatalf("newExecSession wrong: %v", err)
	}

	ct, err := cli.Encrypt([]byte("secret"))
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	if _, err := wrong.Decrypt(ct); err == nil {
		t.Fatal("expected decrypt with wrong key to fail")
	}
}

// TestExecSession_TamperFails ensures GCM authentication rejects modified
// ciphertext.
func TestExecSession_TamperFails(t *testing.T) {
	execKey, sessionNonce := testKeyAndNonce(t)
	cli, err := newExecSession(execKey, sessionNonce)
	if err != nil {
		t.Fatalf("newExecSession: %v", err)
	}
	agent, err := newAgentSession(execKey, sessionNonce)
	if err != nil {
		t.Fatalf("newAgentSession: %v", err)
	}

	ct, err := cli.Encrypt([]byte("authenticated"))
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	ct[len(ct)-1] ^= 0xFF // flip a tag byte
	if _, err := agent.Decrypt(ct); err == nil {
		t.Fatal("expected tampered ciphertext to fail authentication")
	}
}

// TestExecSession_ShortCiphertext ensures Decrypt rejects truncated input.
func TestExecSession_ShortCiphertext(t *testing.T) {
	execKey, sessionNonce := testKeyAndNonce(t)
	cli, err := newExecSession(execKey, sessionNonce)
	if err != nil {
		t.Fatalf("newExecSession: %v", err)
	}
	if _, err := cli.Decrypt(make([]byte, execGCMNonceSize+execGCMTagSize-1)); err == nil {
		t.Fatal("expected short ciphertext to be rejected")
	}
}

// TestDeriveSessionKey_MatchesHKDF proves deriveSessionKey equals
// HKDF-SHA256(exec_key, salt=session_nonce, info="session-key"), i.e. the exact
// derivation the exec-agent's NewSession performs.
func TestDeriveSessionKey_MatchesHKDF(t *testing.T) {
	execKey, sessionNonce := testKeyAndNonce(t)

	got, err := deriveSessionKey(execKey, sessionNonce)
	if err != nil {
		t.Fatalf("deriveSessionKey: %v", err)
	}
	if len(got) != execSessionKeySize {
		t.Fatalf("session key length = %d, want %d", len(got), execSessionKeySize)
	}

	r := hkdf.New(sha256.New, execKey, sessionNonce, []byte("session-key"))
	want := make([]byte, execSessionKeySize)
	if _, err := io.ReadFull(r, want); err != nil {
		t.Fatalf("reference HKDF: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("deriveSessionKey mismatch:\n got  %x\n want %x", got, want)
	}
}

// TestDeriveSessionKey_RejectsBadExecKey ensures a non-32-byte exec_key errors.
func TestDeriveSessionKey_RejectsBadExecKey(t *testing.T) {
	if _, err := deriveSessionKey(make([]byte, 16), make([]byte, 32)); err == nil {
		t.Fatal("expected error for 16-byte exec_key")
	}
}

// TestDecodeControlFrame covers the relay asymmetry: provider->CLI re-wraps
// agent frames in an outer WSFrameData byte, so control frames arrive as
// [0x01][type][payload] and DATA arrives as [0x01][ciphertext].
func TestDecodeControlFrame(t *testing.T) {
	// KEY_ACK re-wrapped: [WSFrameData][KEY_ACK]
	ft, _, ok := decodeControlFrame([]byte{wsFrameData, wsFrameKeyAck})
	if !ok || ft != wsFrameKeyAck {
		t.Fatalf("rewrapped KEY_ACK: ok=%v ft=0x%02x", ok, ft)
	}

	// ERROR re-wrapped carries a message payload.
	ft, payload, ok := decodeControlFrame(append([]byte{wsFrameData, wsFrameError}, []byte("boom")...))
	if !ok || ft != wsFrameError || string(payload) != "boom" {
		t.Fatalf("rewrapped ERROR: ok=%v ft=0x%02x payload=%q", ok, ft, payload)
	}

	// DATA ciphertext whose first byte happens to be a non-control value.
	cipher := []byte{0x42, 0x99, 0x01}
	ft, payload, ok = decodeControlFrame(append([]byte{wsFrameData}, cipher...))
	if !ok || ft != wsFrameData || !bytes.Equal(payload, cipher) {
		t.Fatalf("DATA passthrough: ok=%v ft=0x%02x payload=%x", ok, ft, payload)
	}

	// Defensive top-level KEY_ACK (no outer wrapper).
	ft, _, ok = decodeControlFrame([]byte{wsFrameKeyAck})
	if !ok || ft != wsFrameKeyAck {
		t.Fatalf("top-level KEY_ACK: ok=%v ft=0x%02x", ok, ft)
	}

	// Empty message is rejected.
	if _, _, ok := decodeControlFrame(nil); ok {
		t.Fatal("empty message should not decode")
	}
}

// TestResizePayloadBE verifies the 4-byte big-endian resize encoding matches
// what the daemon relay and exec-agent decode.
func TestResizePayloadBE(t *testing.T) {
	p := resizePayloadBE(120, 40)
	if len(p) != 4 {
		t.Fatalf("resize payload length = %d, want 4", len(p))
	}
	if c := binary.BigEndian.Uint16(p[0:2]); c != 120 {
		t.Fatalf("cols = %d, want 120", c)
	}
	if r := binary.BigEndian.Uint16(p[2:4]); r != 40 {
		t.Fatalf("rows = %d, want 40", r)
	}
}
