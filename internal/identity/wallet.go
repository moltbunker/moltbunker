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
	keystore *keystore.KeyStore
	keyPath  string
	address  common.Address
	privateKey *ecdsa.PrivateKey
}

// NewWalletManager creates a new wallet manager
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
	}

	return wm, nil
}

// Address returns the Ethereum address
func (wm *WalletManager) Address() common.Address {
	return wm.address
}

// AddressString returns the address as a hex string
func (wm *WalletManager) AddressString() string {
	return wm.address.Hex()
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
