//go:build linux

package main

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

// cliDirBit mirrors the CLI/client direction tag (cmd/cli/commands/exec_session.go
// execNonceDirBit): the client encrypts with the high bit of nonce[0] SET
// (direction 1). The agent's Session clears it (direction 0, nonceDirBit). The
// two MUST differ so the directions never collide on a (key, nonce) pair.
const cliDirBit = byte(0x80)

// cliFormatSession is a faithful, in-test reimplementation of the CLI's
// execSession (cmd/cli/commands/exec_session.go). The real CLI type lives in a
// different (non-linux-tagged) package and cannot be imported here, so we mirror
// it byte-for-byte to prove wire interop:
//
//	session_key = HKDF-SHA256(exec_key, salt=session_nonce, info="session-key")
//	Encrypt: nonce = [8B BE counter starting at 1, high bit SET for direction 1][4B random]
//	         out = Seal(nonce,nonce,pt,nil)
//	Decrypt: data = [12B nonce][ct+tag]; Open(nil,nonce,ct,nil)
type cliFormatSession struct {
	aead           cipher.AEAD
	encryptCounter atomic.Uint64
}

func newCLIFormatSession(execKey, sessionNonce []byte) (*cliFormatSession, error) {
	r := hkdf.New(sha256.New, execKey, sessionNonce, []byte("session-key"))
	sessionKey := make([]byte, sessionKeySize)
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
	return &cliFormatSession{aead: aead}, nil
}

func (s *cliFormatSession) Encrypt(plaintext []byte) ([]byte, error) {
	nonce := make([]byte, gcmNonceSize)
	counter := s.encryptCounter.Add(1)
	binary.BigEndian.PutUint64(nonce[:8], counter)
	nonce[0] |= cliDirBit // CLI/client direction = 1
	if _, err := rand.Read(nonce[8:]); err != nil {
		return nil, err
	}
	return s.aead.Seal(nonce, nonce, plaintext, nil), nil
}

func (s *cliFormatSession) Decrypt(data []byte) ([]byte, error) {
	nonce := data[:gcmNonceSize]
	ct := data[gcmNonceSize:]
	return s.aead.Open(nil, nonce, ct, nil)
}

func testAgentKeyAndNonce(t *testing.T) (execKey, sessionNonce []byte) {
	t.Helper()
	execKey = make([]byte, sessionKeySize)
	if _, err := rand.Read(execKey); err != nil {
		t.Fatalf("rand execKey: %v", err)
	}
	sessionNonce = make([]byte, sessionKeySize)
	if _, err := rand.Read(sessionNonce); err != nil {
		t.Fatalf("rand sessionNonce: %v", err)
	}
	return execKey, sessionNonce
}

// TestSession_AgentEncrypt_CLIDecrypt proves the agent's outgoing ciphertext is
// decryptable by the CLI session format (agent -> client direction).
func TestSession_AgentEncrypt_CLIDecrypt(t *testing.T) {
	execKey, sessionNonce := testAgentKeyAndNonce(t)

	agent, err := NewSession(execKey, sessionNonce)
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	cli, err := newCLIFormatSession(execKey, sessionNonce)
	if err != nil {
		t.Fatalf("newCLIFormatSession: %v", err)
	}

	for _, pt := range [][]byte{
		[]byte("total 0\r\ndrwxr-xr-x\r\n"),
		bytes.Repeat([]byte("y"), 4096),
		{},
		{0x00},
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

// TestSession_CLIEncrypt_AgentDecrypt proves the CLI's outgoing ciphertext is
// decryptable by the agent session (client -> agent direction).
func TestSession_CLIEncrypt_AgentDecrypt(t *testing.T) {
	execKey, sessionNonce := testAgentKeyAndNonce(t)

	agent, err := NewSession(execKey, sessionNonce)
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	cli, err := newCLIFormatSession(execKey, sessionNonce)
	if err != nil {
		t.Fatalf("newCLIFormatSession: %v", err)
	}

	for _, pt := range [][]byte{
		[]byte("ls -la\r"),
		bytes.Repeat([]byte{0xAB}, 1000),
		{},
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

// TestSession_DirectionBitCleared proves the agent encrypts with the direction
// bit CLEAR (direction 0) and that, for equal counters, its nonce differs from
// the CLI's (direction 1). This is the cross-direction collision guarantee.
func TestSession_DirectionBitCleared(t *testing.T) {
	execKey, sessionNonce := testAgentKeyAndNonce(t)
	agent, err := NewSession(execKey, sessionNonce)
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	cli, err := newCLIFormatSession(execKey, sessionNonce)
	if err != nil {
		t.Fatalf("newCLIFormatSession: %v", err)
	}

	pt := []byte("collision-check")
	for i := 1; i <= 1000; i++ {
		agentCT, err := agent.Encrypt(pt)
		if err != nil {
			t.Fatalf("agent Encrypt: %v", err)
		}
		cliCT, err := cli.Encrypt(pt)
		if err != nil {
			t.Fatalf("CLI Encrypt: %v", err)
		}
		agentNonce := agentCT[:gcmNonceSize]
		cliNonce := cliCT[:gcmNonceSize]

		if agentNonce[0]&nonceDirBit != 0 {
			t.Fatalf("iter %d: agent direction bit set: nonce[0]=0x%02x", i, agentNonce[0])
		}
		if cliNonce[0]&nonceDirBit == 0 {
			t.Fatalf("iter %d: CLI direction bit not set: nonce[0]=0x%02x", i, cliNonce[0])
		}
		if bytes.Equal(agentNonce, cliNonce) {
			t.Fatalf("iter %d: agent and CLI produced identical nonce %x", i, agentNonce)
		}
	}
}

// TestSession_WireFormat verifies the agent's on-wire nonce: 12-byte prefix with
// a monotonic counter (direction bit cleared) and a 4-byte random tail.
func TestSession_WireFormat(t *testing.T) {
	execKey, sessionNonce := testAgentKeyAndNonce(t)
	agent, err := NewSession(execKey, sessionNonce)
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}

	pt := []byte("hello")
	first, err := agent.Encrypt(pt)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	if want := gcmNonceSize + len(pt) + gcmTagSize; len(first) != want {
		t.Fatalf("ciphertext length = %d, want %d", len(first), want)
	}
	if first[0]&nonceDirBit != 0 {
		t.Fatalf("agent nonce[0] direction bit set: 0x%02x", first[0])
	}
	if c := binary.BigEndian.Uint64(first[:8]); c != 1 {
		t.Fatalf("first counter = %d, want 1", c)
	}

	second, err := agent.Encrypt(pt)
	if err != nil {
		t.Fatalf("Encrypt 2: %v", err)
	}
	if c := binary.BigEndian.Uint64(second[:8]); c != 2 {
		t.Fatalf("second counter = %d, want 2", c)
	}
}

// TestSession_TamperFails ensures GCM authentication rejects modified ciphertext
// across the direction boundary (CLI-encrypted, tampered, agent-decrypted).
func TestSession_TamperFails(t *testing.T) {
	execKey, sessionNonce := testAgentKeyAndNonce(t)
	agent, err := NewSession(execKey, sessionNonce)
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	cli, err := newCLIFormatSession(execKey, sessionNonce)
	if err != nil {
		t.Fatalf("newCLIFormatSession: %v", err)
	}

	ct, err := cli.Encrypt([]byte("authenticated"))
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	ct[len(ct)-1] ^= 0xFF
	if _, err := agent.Decrypt(ct); err == nil {
		t.Fatal("expected tampered ciphertext to fail authentication")
	}
}
