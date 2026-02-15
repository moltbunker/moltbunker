package commands

import (
	"fmt"
	"os"

	"github.com/moltbunker/moltbunker/internal/client"
	"github.com/moltbunker/moltbunker/internal/identity"
)

// GetClient builds a SmartClient using the current CLI globals.
// It tries the daemon socket first, then falls back to HTTP API with wallet signing.
func GetClient() (*client.SmartClient, error) {
	daemon := client.NewDaemonClient(SocketPath)

	var apiClient *client.APIClient
	if endpoint := GetAPIEndpoint(); endpoint != "" {
		signer, _ := loadWalletSigner() // nil signer is OK — unauthenticated requests
		apiClient = client.NewAPIClient(endpoint, signer)
	}

	sc := client.NewSmartClient(daemon, apiClient)
	if err := sc.Connect(); err != nil {
		return nil, err
	}

	return sc, nil
}

// GetClientOrDie builds a SmartClient or exits with a styled error.
func GetClientOrDie() *client.SmartClient {
	c, err := GetClient()
	if err != nil {
		Error(err.Error())
		os.Exit(1)
	}
	return c
}

// loadWalletSigner attempts to load the wallet and resolve the password
// for HTTP API signing. Returns (nil, err) if no wallet or password available.
func loadWalletSigner() (*client.WalletSigner, error) {
	keystoreDir := GetKeystoreDir()
	wm, err := identity.LoadWalletManager(keystoreDir)
	if err != nil || wm == nil {
		return nil, fmt.Errorf("no wallet available")
	}

	password := resolveWalletPassword()
	if password == "" {
		return nil, fmt.Errorf("wallet password not available")
	}

	return client.NewWalletSigner(wm, password), nil
}

// resolveWalletPassword tries to get the wallet password from:
// 1. Platform keyring (macOS Keychain / Linux Secret Service)
// 2. Kernel keyring
// 3. MOLTBUNKER_WALLET_PASSWORD env var
// 4. Password file from config
func resolveWalletPassword() string {
	// Try platform keyring
	if pw, err := identity.RetrieveWalletPassword(); err == nil && pw != "" {
		return pw
	}

	// Try kernel keyring
	if pw, err := identity.RetrieveKernelKeyring(); err == nil && pw != "" {
		return pw
	}

	// Try env var
	if pw := os.Getenv("MOLTBUNKER_WALLET_PASSWORD"); pw != "" {
		return pw
	}

	// Try password file from config
	cfg := loadConfigQuiet()
	if cfg != nil && cfg.Node.WalletPasswordFile != "" {
		data, err := os.ReadFile(cfg.Node.WalletPasswordFile)
		if err == nil && len(data) > 0 {
			return string(data)
		}
	}

	return ""
}
