package commands

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/moltbunker/moltbunker/internal/client"
	"github.com/moltbunker/moltbunker/internal/identity"
)

// runExecDirect connects to the container via direct P2P (Path B).
// Requires the CLI user to have staked tokens (stake gate on exec messages).
func runExecDirect(ctx context.Context, containerID string, execCmd []string, keystoreDir string) error {
	// Load wallet
	wm, err := identity.LoadWalletManager(keystoreDir)
	if err != nil {
		return fmt.Errorf("failed to load wallet: %w", err)
	}
	if wm == nil {
		return fmt.Errorf("no wallet found (create one with: moltbunker wallet create)")
	}

	// Query daemon for container detail (provider NodeID + address)
	dc := client.NewDaemonClient(resolveSocketPath())
	if err := dc.Connect(); err != nil {
		return fmt.Errorf("failed to connect to daemon: %w", err)
	}
	defer dc.Close()

	detail, err := dc.GetContainerDetail(containerID)
	if err != nil {
		return fmt.Errorf("failed to get container detail: %w", err)
	}

	if detail.ProviderAddress == "" {
		return fmt.Errorf("provider address not available for container %s", containerID)
	}

	_, _ = execCmd, detail

	// TODO: Phase 2B implementation
	// 1. Generate ephemeral TLS cert via identity.NewCertificateManager
	// 2. p2p.Transport.DialContext(providerNodeID, providerAddr) with TLS 1.3
	// 3. Announce protocol (EIP-191 wallet proof within 30s)
	// 4. Send ExecOpen P2P message
	// 5. Bridge stdin/stdout ↔ ExecData P2P messages

	return fmt.Errorf("direct P2P exec is not yet implemented (use Path A without --direct)")
}

// signExecChallenge signs a challenge nonce using EIP-191 personal_sign.
// Returns the hex-encoded signature.
func signExecChallenge(privKey *ecdsa.PrivateKey, nonce string) (string, error) {
	// EIP-191 personal_sign: prefix + message
	msg := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(nonce), nonce)
	hash := crypto.Keccak256([]byte(msg))

	sig, err := crypto.Sign(hash, privKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign: %w", err)
	}

	// Adjust V value for Ethereum compatibility (27/28)
	if sig[64] < 27 {
		sig[64] += 27
	}

	return "0x" + hex.EncodeToString(sig), nil
}
