package commands

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	"github.com/moltbunker/moltbunker/internal/client"
	"github.com/moltbunker/moltbunker/internal/identity"
	"github.com/spf13/cobra"
)

// WebSocket frame types for exec protocol (matches internal/api/exec_handler.go
// and the exec-agent wire protocol in cmd/exec-agent/protocol.go).
const (
	wsFrameData    byte = 0x01 // terminal I/O (plaintext on the legacy path, ciphertext on the E2E path)
	wsFrameResize  byte = 0x02 // terminal dimensions
	wsFramePing    byte = 0x03 // keepalive
	wsFramePong    byte = 0x04 // keepalive response
	wsFrameClose   byte = 0x05 // graceful close
	wsFrameError   byte = 0x06 // error message
	wsFrameKeyInit byte = 0x07 // session key initialization (carries session_nonce)
	wsFrameKeyAck  byte = 0x08 // session key acknowledgment
)

// keyAckTimeout bounds how long the CLI waits for the exec-agent's KEY_ACK
// before giving up on the E2E handshake.
const keyAckTimeout = 15 * time.Second

// NewExecCmd creates the exec command
func NewExecCmd() *cobra.Command {
	var (
		apiURL      string
		keystoreDir string
		direct      bool
	)

	cmd := &cobra.Command{
		Use:   "exec <container-id> [command]",
		Short: "Execute a command in a running container",
		Long: `Open an interactive terminal session in a running container.

Requires wallet authentication — only the container owner can exec.

By default, connects via the API WebSocket (Path A). Use --direct for
a direct P2P connection to the provider node (lower latency, requires staking).

Examples:
  moltbunker exec dep-abc123              # Interactive shell (default /bin/sh)
  moltbunker exec dep-abc123 bash         # Specific shell
  moltbunker exec dep-abc123 ls -la       # Run a command
  moltbunker exec --direct dep-abc123     # Direct P2P connection`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			containerID := args[0]
			execCmd := args[1:]

			if direct {
				return runExecDirect(cmd.Context(), containerID, execCmd, keystoreDir)
			}
			return runExecWebSocket(cmd.Context(), containerID, execCmd, apiURL, keystoreDir)
		},
	}

	cmd.Flags().StringVar(&apiURL, "api-url", "ws://127.0.0.1:8080", "API server WebSocket URL")
	cmd.Flags().StringVar(&keystoreDir, "keystore", defaultKeystoreDir(), "Path to wallet keystore")
	cmd.Flags().BoolVar(&direct, "direct", false, "Connect directly to provider via P2P (requires staking)")

	return cmd
}

// runExecWebSocket connects to the container via the API WebSocket (Path A).
func runExecWebSocket(ctx context.Context, containerID string, execCmd []string, apiURL string, keystoreDir string) error {
	// Load wallet for authentication
	wm, err := identity.LoadWalletManager(keystoreDir)
	if err != nil {
		return fmt.Errorf("failed to load wallet: %w", err)
	}
	if wm == nil {
		return fmt.Errorf("no wallet found at %s (create one with: moltbunker wallet create)", keystoreDir)
	}

	walletAddr := wm.Address().Hex()

	// Step 1: Fetch container detail (provider info + exec-agent status + deploy_nonce)
	// and build the challenge nonce. The WebSocket endpoint handles the actual
	// challenge-response.
	detail, err := fetchContainerDetail(containerID)
	if err != nil {
		return fmt.Errorf("failed to get container info: %w", err)
	}

	nonce := challengeNonce(containerID, walletAddr)

	// Step 2: Sign the nonce with the wallet
	password, err := getWalletPassword()
	if err != nil {
		return fmt.Errorf("failed to get wallet password: %w", err)
	}

	privKey, err := wm.ExportKey(password)
	if err != nil {
		return fmt.Errorf("failed to unlock wallet: %w", err)
	}

	signature, err := signExecChallenge(privKey, nonce)
	if err != nil {
		return fmt.Errorf("failed to sign challenge: %w", err)
	}

	// Step 3: Connect to WebSocket with signed auth
	cols, rows := terminalSize()
	wsURL, err := url.Parse(apiURL)
	if err != nil {
		return fmt.Errorf("invalid API URL: %w", err)
	}
	wsURL.Path = "/v1/exec/ws"
	q := wsURL.Query()
	q.Set("container_id", containerID)
	q.Set("nonce", nonce)
	q.Set("signature", signature)
	q.Set("wallet", walletAddr)
	q.Set("cols", fmt.Sprintf("%d", cols))
	q.Set("rows", fmt.Sprintf("%d", rows))
	if len(execCmd) > 0 {
		q.Set("command", strings.Join(execCmd, " "))
	}
	wsURL.RawQuery = q.Encode()

	dialer := websocket.DefaultDialer
	dialer.HandshakeTimeout = 10 * time.Second

	conn, _, err := dialer.DialContext(ctx, wsURL.String(), nil)
	if err != nil {
		return fmt.Errorf("WebSocket connection failed: %w", err)
	}
	defer conn.Close()

	// Step 4: If the container has an exec-agent, perform the E2E handshake and
	// run the encrypted bridge. Otherwise fall back to the plaintext bridge so
	// containers deployed without an exec key keep working unchanged.
	if detail.ExecAgentEnabled && detail.DeployNonce != "" {
		session, err := negotiateExecSession(conn, wm, password, detail.DeployNonce, cols, rows)
		if err != nil {
			return fmt.Errorf("E2E exec handshake failed: %w", err)
		}
		return bridgeTerminalEncrypted(ctx, conn, session)
	}

	return bridgeTerminalToWebSocket(ctx, conn, cols, rows)
}

// negotiateExecSession performs the CLI side of the E2E exec handshake:
//
//  1. master_kek = HKDF(wallet signature, ...)
//  2. exec_key   = HKDF(master_kek, salt=deploy_nonce, info="exec-key")
//  3. session_nonce = 32 random bytes
//  4. session_key = HKDF(exec_key, salt=session_nonce, info="session-key")
//  5. send KEY_INIT(0x07) with payload = the exact 32-byte session_nonce
//  6. await KEY_ACK(0x08) (handle ERROR 0x06)
//  7. send the real terminal size as a RESIZE(0x02) frame
//
// Per the design rule, KEY_INIT carries ONLY the 32-byte session_nonce (no
// cols:rows suffix) so the HKDF salt is exactly the nonce on both sides; the
// terminal size is delivered separately via RESIZE right after KEY_ACK.
func negotiateExecSession(conn *websocket.Conn, wm *identity.WalletManager, password, deployNonceHex string, cols, rows uint16) (*execSession, error) {
	deployNonce, err := hex.DecodeString(deployNonceHex)
	if err != nil {
		return nil, fmt.Errorf("decode deploy_nonce: %w", err)
	}

	masterKEK, err := deriveMasterKEK(wm, password)
	if err != nil {
		return nil, fmt.Errorf("derive master KEK: %w", err)
	}

	execKey, err := deriveExecKey(masterKEK, deployNonce)
	if err != nil {
		return nil, fmt.Errorf("derive exec key: %w", err)
	}

	sessionNonce := make([]byte, execSessionKeySize)
	if _, err := rand.Read(sessionNonce); err != nil {
		return nil, fmt.Errorf("generate session nonce: %w", err)
	}

	session, err := newExecSession(execKey, sessionNonce)
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	// Send KEY_INIT as a top-level WS frame: [0x07][32-byte session_nonce].
	// The API switch keys on message[0], so the type byte must be top-level.
	if err := writeWSFrame(conn, wsFrameKeyInit, sessionNonce); err != nil {
		return nil, fmt.Errorf("send KEY_INIT: %w", err)
	}

	// Await KEY_ACK. The provider->CLI direction re-wraps the agent's frame in
	// an outer WSFrameData byte, so KEY_ACK arrives as [0x01][0x08]; we also
	// accept a defensive top-level [0x08].
	if err := awaitKeyAck(conn); err != nil {
		return nil, err
	}

	// Send the real terminal size now (4-byte big-endian cols/rows), matching
	// the daemon's RESIZE decoder and the exec-agent's parseResizePayload.
	if err := writeWSFrame(conn, wsFrameResize, resizePayloadBE(cols, rows)); err != nil {
		return nil, fmt.Errorf("send initial RESIZE: %w", err)
	}

	return session, nil
}

// awaitKeyAck blocks until a KEY_ACK frame is received or the timeout elapses.
// It surfaces ERROR frames from the agent.
func awaitKeyAck(conn *websocket.Conn) error {
	deadline := time.Now().Add(keyAckTimeout)
	if err := conn.SetReadDeadline(deadline); err != nil {
		return fmt.Errorf("set read deadline: %w", err)
	}
	// Clear the read deadline before returning so the bridge read loop blocks
	// normally afterwards.
	defer func() { _ = conn.SetReadDeadline(time.Time{}) }()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			return fmt.Errorf("read KEY_ACK: %w", err)
		}
		ackType, payload, ok := decodeControlFrame(message)
		if !ok {
			continue
		}
		switch ackType {
		case wsFrameKeyAck:
			return nil
		case wsFrameError:
			return fmt.Errorf("agent error during handshake: %s", string(payload))
		default:
			// Ignore any stray frames (e.g. PONG) before KEY_ACK.
		}
	}
}

// decodeControlFrame interprets a WS message and returns the semantic agent
// frame type and its payload. Because the provider->CLI direction wraps agent
// frames in an outer WSFrameData byte, control frames arrive as
// [WSFrameData][agentType][agentPayload]. DATA frames arrive as
// [WSFrameData][ciphertext] and are reported with type wsFrameData and the
// ciphertext as payload. A defensive top-level interpretation is also applied
// for non-DATA frames.
func decodeControlFrame(message []byte) (frameType byte, payload []byte, ok bool) {
	if len(message) < 1 {
		return 0, nil, false
	}
	if message[0] == wsFrameData {
		inner := message[1:]
		if len(inner) == 0 {
			return 0, nil, false
		}
		switch inner[0] {
		case wsFrameKeyAck, wsFrameError, wsFramePing, wsFramePong, wsFrameClose:
			// Control frame re-wrapped by the provider relay.
			return inner[0], inner[1:], true
		default:
			// Genuine DATA payload (ciphertext) — first byte is part of it.
			return wsFrameData, inner, true
		}
	}
	// Defensive: a top-level control frame (no outer WSFrameData wrapper).
	return message[0], message[1:], true
}

// bridgeTerminalToWebSocket enters raw mode and relays between terminal and WebSocket.
func bridgeTerminalToWebSocket(ctx context.Context, conn *websocket.Conn, cols, rows uint16) error {
	// Enter raw terminal mode
	restore, err := makeTerminalRaw()
	if err != nil {
		return fmt.Errorf("failed to set terminal to raw mode: %w", err)
	}
	defer restore()

	// Context for cancellation
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Handle SIGWINCH for terminal resize
	sigWinch := make(chan os.Signal, 1)
	signal.Notify(sigWinch, syscall.SIGWINCH)
	defer signal.Stop(sigWinch)

	errCh := make(chan error, 2)

	// Read from WebSocket → stdout
	go func() {
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
					errCh <- nil
				} else {
					errCh <- err
				}
				return
			}
			if len(message) < 1 {
				continue
			}

			frameType := message[0]
			data := message[1:]

			switch frameType {
			case wsFrameData:
				_, _ = os.Stdout.Write(data)
			}
		}
	}()

	// Read from stdin → WebSocket
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := os.Stdin.Read(buf)
			if n > 0 {
				frame := make([]byte, 1+n)
				frame[0] = wsFrameData
				copy(frame[1:], buf[:n])
				if wErr := conn.WriteMessage(websocket.BinaryMessage, frame); wErr != nil {
					errCh <- wErr
					return
				}
			}
			if err != nil {
				errCh <- nil
				return
			}
		}
	}()

	// Handle SIGWINCH → resize frame
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-sigWinch:
				newCols, newRows := terminalSize()
				resizePayload, _ := json.Marshal(map[string]uint16{
					"cols": newCols,
					"rows": newRows,
				})
				frame := make([]byte, 1+len(resizePayload))
				frame[0] = wsFrameResize
				copy(frame[1:], resizePayload)
				// Best-effort resize notification; failure surfaces via the read loop.
				_ = conn.WriteMessage(websocket.BinaryMessage, frame)
			}
		}
	}()

	// Wait for first error or context cancellation
	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		// Best-effort close; we are shutting down regardless.
		_ = conn.WriteMessage(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		return nil
	}
}

// bridgeTerminalEncrypted relays between the terminal and the WebSocket with
// AES-256-GCM E2E encryption against the in-container exec-agent.
//
//   - stdin keystrokes are encrypted into DATA(0x01) frames.
//   - incoming DATA(0x01) frames are decrypted and written to stdout.
//   - PING/PONG/CLOSE/ERROR are handled out-of-band.
//
// The provider never sees plaintext; it only relays opaque ciphertext.
func bridgeTerminalEncrypted(ctx context.Context, conn *websocket.Conn, session *execSession) error {
	restore, err := makeTerminalRaw()
	if err != nil {
		return fmt.Errorf("failed to set terminal to raw mode: %w", err)
	}
	defer restore()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	sigWinch := make(chan os.Signal, 1)
	signal.Notify(sigWinch, syscall.SIGWINCH)
	defer signal.Stop(sigWinch)

	errCh := make(chan error, 2)
	var writeMu sync.Mutex // serialize concurrent WS writes (stdin loop + resize loop)

	// Read from WebSocket → decrypt → stdout
	go func() {
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
					errCh <- nil
				} else {
					errCh <- err
				}
				return
			}
			frameType, payload, ok := decodeControlFrame(message)
			if !ok {
				continue
			}
			switch frameType {
			case wsFrameData:
				plaintext, derr := session.Decrypt(payload)
				if derr != nil {
					errCh <- fmt.Errorf("decrypt terminal data: %w", derr)
					return
				}
				if _, werr := os.Stdout.Write(plaintext); werr != nil {
					errCh <- fmt.Errorf("write terminal output: %w", werr)
					return
				}
			case wsFramePing:
				writeMu.Lock()
				_ = writeWSFrame(conn, wsFramePong, nil)
				writeMu.Unlock()
			case wsFramePong:
				// keepalive ack — nothing to do
			case wsFrameClose:
				errCh <- nil
				return
			case wsFrameError:
				errCh <- fmt.Errorf("agent error: %s", string(payload))
				return
			}
		}
	}()

	// Read from stdin → encrypt → WebSocket
	go func() {
		buf := make([]byte, 4096)
		for {
			n, rerr := os.Stdin.Read(buf)
			if n > 0 {
				ciphertext, eerr := session.Encrypt(buf[:n])
				if eerr != nil {
					errCh <- fmt.Errorf("encrypt keystroke: %w", eerr)
					return
				}
				writeMu.Lock()
				werr := writeWSFrame(conn, wsFrameData, ciphertext)
				writeMu.Unlock()
				if werr != nil {
					errCh <- werr
					return
				}
			}
			if rerr != nil {
				errCh <- nil
				return
			}
		}
	}()

	// Handle SIGWINCH → RESIZE frame (4-byte big-endian, unencrypted control)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-sigWinch:
				newCols, newRows := terminalSize()
				writeMu.Lock()
				_ = writeWSFrame(conn, wsFrameResize, resizePayloadBE(newCols, newRows))
				writeMu.Unlock()
			}
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		writeMu.Lock()
		_ = conn.WriteMessage(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		writeMu.Unlock()
		return nil
	}
}

// writeWSFrame writes a single-byte-type-prefixed WebSocket binary frame.
func writeWSFrame(conn *websocket.Conn, frameType byte, payload []byte) error {
	frame := make([]byte, 1+len(payload))
	frame[0] = frameType
	copy(frame[1:], payload)
	return conn.WriteMessage(websocket.BinaryMessage, frame)
}

// resizePayloadBE encodes a terminal size as the 4-byte big-endian format
// expected by the daemon relay (internal/api/exec_handler.go) and the
// exec-agent (parseResizePayload): [cols hi][cols lo][rows hi][rows lo].
func resizePayloadBE(cols, rows uint16) []byte {
	payload := make([]byte, 4)
	binary.BigEndian.PutUint16(payload[0:2], cols)
	binary.BigEndian.PutUint16(payload[2:4], rows)
	return payload
}

// fetchContainerDetail retrieves container detail (provider info, exec-agent
// status, deploy_nonce) from the local daemon over the Unix socket.
func fetchContainerDetail(containerID string) (*client.ContainerDetail, error) {
	dc := client.NewDaemonClient(resolveSocketPath())
	if err := dc.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect to daemon: %w", err)
	}
	defer dc.Close()

	detail, err := dc.GetContainerDetail(containerID)
	if err != nil {
		return nil, err
	}
	return detail, nil
}

// challengeNonce builds the deterministic challenge nonce.
// The actual challenge-response is handled by the WebSocket endpoint.
func challengeNonce(containerID, walletAddr string) string {
	return fmt.Sprintf("%s:%s:%d", containerID, walletAddr, time.Now().UnixNano())
}

// resolveSocketPath returns the daemon socket path from flag or default.
func resolveSocketPath() string {
	if SocketPath != "" {
		return SocketPath
	}
	return client.DefaultSocketPath()
}

// getWalletPassword retrieves the wallet password from keyring or prompts.
func getWalletPassword() (string, error) {
	// Try platform keyring first
	if pw, err := identity.RetrieveWalletPassword(); err == nil && pw != "" {
		return pw, nil
	}

	// Try kernel keyring
	if pw, err := identity.RetrieveKernelKeyring(); err == nil && pw != "" {
		return pw, nil
	}

	// Try environment variable
	if pw := os.Getenv("MOLTBUNKER_WALLET_PASSWORD"); pw != "" {
		return pw, nil
	}

	// Prompt the user
	fmt.Fprint(os.Stderr, "Enter wallet password: ")
	pw, err := readPasswordNoEcho()
	if err != nil {
		return "", err
	}
	fmt.Fprintln(os.Stderr)
	return pw, nil
}
