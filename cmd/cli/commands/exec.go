package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	"github.com/moltbunker/moltbunker/internal/client"
	"github.com/moltbunker/moltbunker/internal/identity"
	"github.com/spf13/cobra"
)

// WebSocket frame types for exec protocol (matches internal/api/exec_handler.go)
const (
	wsFrameData   byte = 0x01
	wsFrameResize byte = 0x02
)

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

	// Step 1: Request a challenge nonce from the API
	challengeURL := strings.Replace(apiURL, "ws://", "http://", 1)
	challengeURL = strings.Replace(challengeURL, "wss://", "https://", 1)

	nonce, err := requestExecChallenge(challengeURL, containerID, walletAddr)
	if err != nil {
		return fmt.Errorf("challenge request failed: %w", err)
	}

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

	// Step 4: Enter raw terminal mode and bridge stdin/stdout
	return bridgeTerminalToWebSocket(ctx, conn, cols, rows)
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
				os.Stdout.Write(data)
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
				conn.WriteMessage(websocket.BinaryMessage, frame)
			}
		}
	}()

	// Wait for first error or context cancellation
	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		conn.WriteMessage(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		return nil
	}
}

// requestExecChallenge requests a challenge nonce from the API server.
func requestExecChallenge(baseURL string, containerID string, walletAddr string) (string, error) {
	// Use the daemon client to request a challenge through the Unix socket
	// For HTTP API, we'd POST to /v1/exec/challenge
	dc := client.NewDaemonClient(resolveSocketPath())
	if err := dc.Connect(); err != nil {
		return "", fmt.Errorf("failed to connect to daemon: %w", err)
	}
	defer dc.Close()

	detail, err := dc.GetContainerDetail(containerID)
	if err != nil {
		return "", fmt.Errorf("failed to get container info: %w", err)
	}

	// Use the container ID + wallet + timestamp as a deterministic nonce
	// The actual challenge-response is handled by the WebSocket endpoint
	_ = detail
	nonce := fmt.Sprintf("%s:%s:%d", containerID, walletAddr, time.Now().UnixNano())
	return nonce, nil
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
