package doctor

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/moltbunker/moltbunker/internal/identity"
)

// WalletChecker checks if an Ethereum wallet is configured.
// Required for ALL roles — requesters need it to sign API requests,
// providers need it for staking and identity binding.
type WalletChecker struct{}

func NewWalletChecker() *WalletChecker {
	return &WalletChecker{}
}

func (c *WalletChecker) Name() string       { return "Wallet" }
func (c *WalletChecker) Category() Category { return CategoryConfig }
func (c *WalletChecker) CanFix() bool       { return true }

func (c *WalletChecker) Check(ctx context.Context) CheckResult {
	result := CheckResult{
		Name:       c.Name(),
		Category:   c.Category(),
		Fixable:    true,
		FixCommand: "moltbunker wallet create",
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		result.Status = StatusError
		result.Message = "Wallet: Unable to check"
		result.Details = err.Error()
		return result
	}

	keystoreDir := filepath.Join(homeDir, ".moltbunker", "keystore")

	wm, err := identity.LoadWalletManager(keystoreDir)
	if err != nil || wm == nil {
		result.Status = StatusError
		result.Message = "Wallet: Not configured"
		result.Details = "A wallet is required for signing transactions and API requests"
		return result
	}

	result.Status = StatusOK
	addr := wm.Address().Hex()
	if len(addr) > 10 {
		addr = addr[:6] + "..." + addr[len(addr)-4:]
	}
	result.Message = fmt.Sprintf("Wallet: %s", addr)
	result.Fixable = false
	return result
}

func (c *WalletChecker) Fix(ctx context.Context, pm PackageManager) error {
	return fmt.Errorf("run 'moltbunker wallet create' to create a wallet")
}

// NodeKeysChecker checks if Ed25519 node identity keys exist.
// Required for ALL roles — keys are used for TLS certificates and P2P identity.
type NodeKeysChecker struct{}

func NewNodeKeysChecker() *NodeKeysChecker {
	return &NodeKeysChecker{}
}

func (c *NodeKeysChecker) Name() string       { return "Node keys" }
func (c *NodeKeysChecker) Category() Category { return CategoryConfig }
func (c *NodeKeysChecker) CanFix() bool       { return true }

func (c *NodeKeysChecker) Check(ctx context.Context) CheckResult {
	result := CheckResult{
		Name:       c.Name(),
		Category:   c.Category(),
		Fixable:    true,
		FixCommand: "moltbunker init",
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		result.Status = StatusError
		result.Message = "Node keys: Unable to check"
		result.Details = err.Error()
		return result
	}

	keyPath := filepath.Join(homeDir, ".moltbunker", "keys", "node.key")

	info, err := os.Stat(keyPath)
	if os.IsNotExist(err) {
		result.Status = StatusError
		result.Message = "Node keys: Not found"
		result.Details = "Run 'moltbunker init' to generate node keys"
		return result
	}
	if err != nil {
		result.Status = StatusError
		result.Message = "Node keys: Unable to check"
		result.Details = err.Error()
		return result
	}

	if info.Size() == 0 {
		result.Status = StatusError
		result.Message = "Node keys: Key file is empty"
		result.Details = "Delete the empty file and run 'moltbunker init' to regenerate"
		return result
	}

	// Verify key is loadable
	_, err = identity.NewKeyManager(keyPath)
	if err != nil {
		result.Status = StatusError
		result.Message = "Node keys: Invalid key file"
		result.Details = err.Error()
		return result
	}

	result.Status = StatusOK
	result.Message = "Node keys: Ed25519 keypair found"
	result.Fixable = false
	return result
}

func (c *NodeKeysChecker) Fix(ctx context.Context, pm PackageManager) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	keyPath := filepath.Join(homeDir, ".moltbunker", "keys", "node.key")
	if err := os.MkdirAll(filepath.Dir(keyPath), 0700); err != nil {
		return err
	}

	// Remove empty file if it exists (NewKeyManager panics on empty files)
	if info, err := os.Stat(keyPath); err == nil && info.Size() == 0 {
		os.Remove(keyPath)
	}

	_, err = identity.NewKeyManager(keyPath)
	return err
}
