package identity

import (
	"crypto/ecdsa"
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// WalletManager manages Ethereum wallet for Base network payments
type WalletManager struct {
	keystore   *keystore.KeyStore
	keyPath    string
	address    common.Address
	privateKey *ecdsa.PrivateKey
	loaded     bool // true if wallet was loaded from an existing keystore file
}

// NewWalletManager creates a new wallet manager.
// Deprecated: Use LoadWalletManager, CreateWalletManager, or ImportWalletManager instead.
// This auto-creates a wallet with an empty password if none exists, which is insecure.
func NewWalletManager(keystoreDir string) (*WalletManager, error) {
	// Create keystore directory if it doesn't exist
	if err := os.MkdirAll(keystoreDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create keystore directory: %w", err)
	}

	ks := keystore.NewKeyStore(keystoreDir, keystore.StandardScryptN, keystore.StandardScryptP)

	wm := &WalletManager{
		keystore: ks,
		keyPath:  keystoreDir,
	}

	// Try to load existing account or create new one
	accounts := ks.Accounts()
	if len(accounts) == 0 {
		// Create new account
		account, err := ks.NewAccount("")
		if err != nil {
			return nil, fmt.Errorf("failed to create new account: %w", err)
		}
		wm.address = account.Address
	} else {
		wm.address = accounts[0].Address
		wm.loaded = true
	}

	return wm, nil
}

// LoadWalletManager loads an existing wallet from the keystore directory.
// Returns (nil, nil) if no wallet file is found — this signals read-only mode.
func LoadWalletManager(keystoreDir string) (*WalletManager, error) {
	if err := os.MkdirAll(keystoreDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create keystore directory: %w", err)
	}

	ks := keystore.NewKeyStore(keystoreDir, keystore.StandardScryptN, keystore.StandardScryptP)
	accounts := ks.Accounts()
	if len(accounts) == 0 {
		return nil, nil // No wallet found — read-only mode
	}

	return &WalletManager{
		keystore: ks,
		keyPath:  keystoreDir,
		address:  accounts[0].Address,
		loaded:   true,
	}, nil
}

// CreateWalletManager creates a new wallet in the keystore directory.
// Returns an error if a wallet already exists (use LoadWalletManager to load it).
func CreateWalletManager(keystoreDir string, password string) (*WalletManager, error) {
	if err := os.MkdirAll(keystoreDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create keystore directory: %w", err)
	}

	ks := keystore.NewKeyStore(keystoreDir, keystore.StandardScryptN, keystore.StandardScryptP)
	if len(ks.Accounts()) > 0 {
		return nil, fmt.Errorf("wallet already exists in %s (use LoadWalletManager to load it)", keystoreDir)
	}

	account, err := ks.NewAccount(password)
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet: %w", err)
	}

	return &WalletManager{
		keystore: ks,
		keyPath:  keystoreDir,
		address:  account.Address,
		loaded:   true,
	}, nil
}

// ImportWalletManager imports a private key into a new wallet in the keystore directory.
// Returns an error if a wallet already exists.
func ImportWalletManager(keystoreDir string, privKeyHex string, password string) (*WalletManager, error) {
	if err := os.MkdirAll(keystoreDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create keystore directory: %w", err)
	}

	ks := keystore.NewKeyStore(keystoreDir, keystore.StandardScryptN, keystore.StandardScryptP)
	if len(ks.Accounts()) > 0 {
		return nil, fmt.Errorf("wallet already exists in %s (use LoadWalletManager to load it)", keystoreDir)
	}

	privateKey, err := crypto.HexToECDSA(privKeyHex)
	if err != nil {
		return nil, fmt.Errorf("invalid private key hex: %w", err)
	}

	account, err := ks.ImportECDSA(privateKey, password)
	if err != nil {
		return nil, fmt.Errorf("failed to import key: %w", err)
	}

	return &WalletManager{
		keystore: ks,
		keyPath:  keystoreDir,
		address:  account.Address,
		loaded:   true,
	}, nil
}

// IsLoaded returns true if the wallet manager has a loaded wallet.
func (wm *WalletManager) IsLoaded() bool {
	return wm != nil && wm.loaded
}

// Address returns the Ethereum address
func (wm *WalletManager) Address() common.Address {
	return wm.address
}

// AddressString returns the address as a hex string
func (wm *WalletManager) AddressString() string {
	return wm.address.Hex()
}

// KeystoreDir returns the path to the keystore directory
func (wm *WalletManager) KeystoreDir() string {
	return wm.keyPath
}

// PrivateKey returns the private key (for signing transactions)
func (wm *WalletManager) PrivateKey(password string) (*ecdsa.PrivateKey, error) {
	if wm.privateKey != nil {
		return wm.privateKey, nil
	}

	accounts := wm.keystore.Accounts()
	if len(accounts) == 0 {
		return nil, fmt.Errorf("no accounts found")
	}

	account := accounts[0]
	keyJSON, err := os.ReadFile(account.URL.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to read key file: %w", err)
	}

	key, err := keystore.DecryptKey(keyJSON, password)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt key: %w", err)
	}

	wm.privateKey = key.PrivateKey
	return key.PrivateKey, nil
}

// ClearCachedKey zeros and removes the cached private key from memory.
// The key will be re-derived from the keystore on next use.
func (wm *WalletManager) ClearCachedKey() {
	if wm.privateKey != nil {
		// Zero the private key bytes before releasing
		wm.privateKey.D.SetUint64(0)
		wm.privateKey = nil
	}
}

// SignHash signs a hash with the wallet's private key
func (wm *WalletManager) SignHash(hash []byte, password string) ([]byte, error) {
	privateKey, err := wm.PrivateKey(password)
	if err != nil {
		return nil, err
	}

	signature, err := crypto.Sign(hash, privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign hash: %w", err)
	}

	return signature, nil
}

// CreateAccount creates a new Ethereum account
func (wm *WalletManager) CreateAccount(password string) (common.Address, error) {
	account, err := wm.keystore.NewAccount(password)
	if err != nil {
		return common.Address{}, fmt.Errorf("failed to create account: %w", err)
	}

	wm.address = account.Address
	return account.Address, nil
}

// ImportKey imports a private key
func (wm *WalletManager) ImportKey(privateKey *ecdsa.PrivateKey, password string) (common.Address, error) {
	account, err := wm.keystore.ImportECDSA(privateKey, password)
	if err != nil {
		return common.Address{}, fmt.Errorf("failed to import key: %w", err)
	}

	wm.address = account.Address
	return account.Address, nil
}

// ExportKey exports the private key
func (wm *WalletManager) ExportKey(password string) (*ecdsa.PrivateKey, error) {
	return wm.PrivateKey(password)
}
