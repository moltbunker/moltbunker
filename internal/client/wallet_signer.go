package client

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/moltbunker/moltbunker/internal/identity"
)

// WalletSigner signs HTTP API requests using the node's Ethereum wallet.
// It produces EIP-191 personal_sign signatures for inline auth headers.
type WalletSigner struct {
	wallet   *identity.WalletManager
	password string
}

// NewWalletSigner creates a signer from an existing wallet and password.
func NewWalletSigner(wallet *identity.WalletManager, password string) *WalletSigner {
	return &WalletSigner{wallet: wallet, password: password}
}

// Address returns the checksummed wallet address.
func (s *WalletSigner) Address() string {
	return s.wallet.Address().Hex()
}

// SignAuth produces the three inline-auth header values:
// address, signature (hex), and message.
//
// Server-side format (wallet_auth.go:161): "moltbunker-auth:{unix_timestamp}"
func (s *WalletSigner) SignAuth() (address, signature, message string, err error) {
	address = s.wallet.Address().Hex()
	message = fmt.Sprintf("moltbunker-auth:%d", time.Now().Unix())

	// EIP-191 personal_sign: keccak256("\x19Ethereum Signed Message:\n" + len(msg) + msg)
	prefixed := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)
	hash := crypto.Keccak256([]byte(prefixed))

	sig, err := s.wallet.SignHash(hash, s.password)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to sign auth message: %w", err)
	}

	// Adjust V value to Ethereum convention (27/28)
	if sig[64] < 27 {
		sig[64] += 27
	}

	signature = "0x" + hex.EncodeToString(sig)
	return address, signature, message, nil
}
