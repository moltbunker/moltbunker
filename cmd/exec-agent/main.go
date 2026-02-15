//go:build linux

package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
)

const (
	// defaultSecretPath is where the exec_key is mounted inside the container.
	defaultSecretPath = "/run/secrets/exec_key"
	// defaultSocketPath is the Unix socket the agent listens on.
	defaultSocketPath = "/run/exec-agent.sock"
	// defaultCols and defaultRows are the default terminal dimensions.
	defaultCols = 80
	defaultRows = 24
)

func main() {
	log.SetFlags(log.Ltime | log.Lmicroseconds)
	log.SetPrefix("[exec-agent] ")

	secretPath := envOr("EXEC_KEY_PATH", defaultSecretPath)
	socketPath := envOr("EXEC_SOCKET_PATH", defaultSocketPath)

	// Load exec_key from mounted secret
	execKey, err := os.ReadFile(secretPath)
	if err != nil {
		log.Fatalf("failed to read exec_key from %s: %v", secretPath, err)
	}
	if len(execKey) != sessionKeySize {
		log.Fatalf("exec_key must be %d bytes, got %d", sessionKeySize, len(execKey))
	}
	log.Printf("loaded exec_key (%d bytes) from %s", len(execKey), secretPath)

	// Remove stale socket
	os.Remove(socketPath)

	// Listen for connections from the provider daemon
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		log.Fatalf("listen on %s: %v", socketPath, err)
	}
	defer listener.Close()
	log.Printf("listening on %s", socketPath)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	var wg sync.WaitGroup

	// Accept connections (one per exec session)
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				select {
				case <-ctx.Done():
					return
				default:
					log.Printf("accept error: %v", err)
					continue
				}
			}

			wg.Add(1)
			go func() {
				defer wg.Done()
				handleSession(ctx, conn, execKey)
			}()
		}
	}()

	<-ctx.Done()
	log.Println("shutting down...")
	listener.Close()
	wg.Wait()
	log.Println("exit")
}

// handleSession manages a single exec session on the accepted Unix socket connection.
// Protocol:
//  1. Read KEY_INIT frame containing session_nonce
//  2. Derive session_key = HKDF(exec_key, session_nonce)
//  3. Send KEY_ACK frame
//  4. Bridge encrypted frames <-> PTY
func handleSession(ctx context.Context, conn net.Conn, execKey []byte) {
	defer conn.Close()
	log.Println("new session connection")

	// Step 1: Read KEY_INIT frame with session_nonce
	initFrame, err := readFrame(conn)
	if err != nil {
		log.Printf("read KEY_INIT: %v", err)
		return
	}
	if initFrame.Type != FrameKeyInit {
		log.Printf("expected KEY_INIT (0x%02x), got 0x%02x", FrameKeyInit, initFrame.Type)
		return
	}

	sessionNonce := initFrame.Payload
	if len(sessionNonce) == 0 {
		log.Println("empty session_nonce in KEY_INIT")
		return
	}
	log.Printf("received KEY_INIT with %d-byte nonce", len(sessionNonce))

	// Step 2: Derive session key
	session, err := NewSession(execKey, sessionNonce)
	if err != nil {
		log.Printf("derive session key: %v", err)
		sendError(conn, fmt.Sprintf("key derivation failed: %v", err))
		return
	}
	log.Println("session key derived")

	// Step 3: Send KEY_ACK
	if err := writeFrame(conn, &Frame{Type: FrameKeyAck}); err != nil {
		log.Printf("send KEY_ACK: %v", err)
		return
	}
	log.Println("sent KEY_ACK")

	// Step 4: Spawn PTY shell
	cols, rows := parseInitialSize(initFrame.Payload)
	pty, err := SpawnShell(cols, rows)
	if err != nil {
		log.Printf("spawn shell: %v", err)
		sendError(conn, fmt.Sprintf("spawn shell failed: %v", err))
		return
	}
	defer pty.Close()
	log.Printf("shell spawned (%dx%d)", cols, rows)

	sessionCtx, sessionCancel := context.WithCancel(ctx)
	defer sessionCancel()

	var wg sync.WaitGroup

	// PTY stdout -> encrypt -> send DATA frames
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer sessionCancel()
		buf := make([]byte, 4096)
		for {
			select {
			case <-sessionCtx.Done():
				return
			default:
			}

			n, err := pty.Read(buf)
			if n > 0 {
				encrypted, encErr := session.Encrypt(buf[:n])
				if encErr != nil {
					log.Printf("encrypt stdout: %v", encErr)
					return
				}
				if wErr := writeFrame(conn, &Frame{Type: FrameData, Payload: encrypted}); wErr != nil {
					log.Printf("send DATA frame: %v", wErr)
					return
				}
			}
			if err != nil {
				if err != io.EOF {
					log.Printf("read PTY: %v", err)
				}
				return
			}
		}
	}()

	// Read frames from conn -> decrypt -> write to PTY stdin / handle control
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer sessionCancel()
		for {
			select {
			case <-sessionCtx.Done():
				return
			default:
			}

			frame, err := readFrame(conn)
			if err != nil {
				if err != io.EOF && !strings.Contains(err.Error(), "use of closed") {
					log.Printf("read frame: %v", err)
				}
				return
			}

			switch frame.Type {
			case FrameData:
				plaintext, err := session.Decrypt(frame.Payload)
				if err != nil {
					log.Printf("decrypt stdin: %v", err)
					return
				}
				if _, err := pty.Write(plaintext); err != nil {
					log.Printf("write PTY stdin: %v", err)
					return
				}

			case FrameResize:
				c, r := parseResizePayload(frame.Payload)
				if c > 0 && r > 0 {
					if err := pty.Resize(c, r); err != nil {
						log.Printf("resize PTY: %v", err)
					}
				}

			case FramePing:
				_ = writeFrame(conn, &Frame{Type: FramePong})

			case FrameClose:
				log.Println("received CLOSE frame")
				return

			default:
				log.Printf("unknown frame type: 0x%02x", frame.Type)
			}
		}
	}()

	// Wait for shell to exit
	go func() {
		_ = pty.Wait()
		sessionCancel()
	}()

	wg.Wait()

	// Send CLOSE frame
	_ = writeFrame(conn, &Frame{Type: FrameClose})
	log.Println("session ended")
}

// parseInitialSize extracts terminal dimensions from the KEY_INIT payload.
// The nonce is the first 32 bytes; optional cols:rows follow after.
// If no size info is present, returns defaults.
func parseInitialSize(payload []byte) (cols, rows uint16) {
	cols, rows = defaultCols, defaultRows

	// If payload is longer than 32 bytes (nonce), the rest may contain "cols:rows"
	if len(payload) > 32 {
		sizeStr := string(payload[32:])
		parts := strings.SplitN(sizeStr, ":", 2)
		if len(parts) == 2 {
			if c, err := strconv.ParseUint(parts[0], 10, 16); err == nil && c > 0 {
				cols = uint16(c)
			}
			if r, err := strconv.ParseUint(parts[1], 10, 16); err == nil && r > 0 {
				rows = uint16(r)
			}
		}
	}
	return cols, rows
}

// parseResizePayload parses a RESIZE frame payload.
// Format: 4 bytes big-endian cols + 4 bytes big-endian rows.
func parseResizePayload(payload []byte) (cols, rows uint16) {
	if len(payload) >= 4 {
		cols = binary.BigEndian.Uint16(payload[0:2])
		rows = binary.BigEndian.Uint16(payload[2:4])
	}
	return cols, rows
}

// sendError sends an ERROR frame to the connection.
func sendError(conn net.Conn, msg string) {
	_ = writeFrame(conn, &Frame{Type: FrameError, Payload: []byte(msg)})
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
