package p2p

import (
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/moltbunker/moltbunker/pkg/types"
)

const (
	// announceMessagePrefix is the prefix for the EIP-191 signed announce message.
	announceMessagePrefix = "moltbunker-announce"

	// AnnounceMaxAge is how old an announce timestamp can be before rejection.
	AnnounceMaxAge = 5 * time.Minute

	// AnnounceMaxFutureSkew is how far in the future an announce timestamp can be.
	AnnounceMaxFutureSkew = 30 * time.Second
)

// CreateAnnouncePayload creates a signed announce payload proving wallet ownership.
// The returned payload binds the NodeID to the wallet address via EIP-191 signature.
func CreateAnnouncePayload(nodeID types.NodeID, wallet common.Address, privKey *ecdsa.PrivateKey) (*types.AnnouncePayload, error) {
	if privKey == nil {
		return nil, fmt.Errorf("private key is nil")
	}

	// Generate random nonce
	nonceBytes := make([]byte, 32)
	if _, err := rand.Read(nonceBytes); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	now := time.Now().Unix()
	nonceHex := hex.EncodeToString(nonceBytes)
	nodeIDHex := nodeID.String()
	walletAddr := wallet.Hex()

	// Construct the message to sign
	message := fmt.Sprintf("%s:%s:%s:%d:%s", announceMessagePrefix, nodeIDHex, walletAddr, now, nonceHex)

	// EIP-191 personal sign: prefix with "\x19Ethereum Signed Message:\n{len}{message}"
	hash := ethPersonalHash(message)

	// Sign the hash
	sig, err := crypto.Sign(hash, privKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign announce message: %w", err)
	}

	// Adjust V value for Ethereum convention (add 27)
	if sig[64] < 27 {
		sig[64] += 27
	}

	return &types.AnnouncePayload{
		WalletAddress: walletAddr,
		NodeID:        nodeIDHex,
		Timestamp:     now,
		Nonce:         nonceHex,
		EthSignature:  hex.EncodeToString(sig),
	}, nil
}

// VerifyAnnouncePayload verifies an announce payload's signature and returns
// the recovered Ethereum address. The caller should check that:
// 1. The recovered address matches the claimed WalletAddress
// 2. The NodeID matches the TLS-authenticated peer
// 3. The timestamp is recent
func VerifyAnnouncePayload(payload *types.AnnouncePayload) (common.Address, error) {
	if payload == nil {
		return common.Address{}, fmt.Errorf("nil announce payload")
	}

	// Validate timestamp
	ts := time.Unix(payload.Timestamp, 0)
	now := time.Now()
	if now.Sub(ts) > AnnounceMaxAge {
		return common.Address{}, fmt.Errorf("announce timestamp too old: %s ago", now.Sub(ts).Round(time.Second))
	}
	if ts.Sub(now) > AnnounceMaxFutureSkew {
		return common.Address{}, fmt.Errorf("announce timestamp too far in future")
	}

	// Validate claimed address format
	if !common.IsHexAddress(payload.WalletAddress) {
		return common.Address{}, fmt.Errorf("invalid wallet address format: %s", payload.WalletAddress)
	}

	// Validate nonce length (32 bytes = 64 hex chars)
	if len(payload.Nonce) != 64 {
		return common.Address{}, fmt.Errorf("invalid nonce length: expected 64 hex chars, got %d", len(payload.Nonce))
	}

	// Decode signature
	sigBytes, err := hex.DecodeString(strings.TrimPrefix(payload.EthSignature, "0x"))
	if err != nil {
		return common.Address{}, fmt.Errorf("invalid signature hex: %w", err)
	}
	if len(sigBytes) != 65 {
		return common.Address{}, fmt.Errorf("invalid signature length: expected 65, got %d", len(sigBytes))
	}

	// Reconstruct the signed message
	message := fmt.Sprintf("%s:%s:%s:%d:%s", announceMessagePrefix, payload.NodeID, payload.WalletAddress, payload.Timestamp, payload.Nonce)

	// Hash with EIP-191 prefix
	hash := ethPersonalHash(message)

	// Adjust V value for recovery
	sigForRecovery := make([]byte, 65)
	copy(sigForRecovery, sigBytes)
	if sigForRecovery[64] >= 27 {
		sigForRecovery[64] -= 27
	}

	// Recover public key
	pubKey, err := crypto.SigToPub(hash, sigForRecovery)
	if err != nil {
		return common.Address{}, fmt.Errorf("failed to recover public key: %w", err)
	}

	recoveredAddr := crypto.PubkeyToAddress(*pubKey)

	// Verify recovered address matches claimed address
	claimedAddr := common.HexToAddress(payload.WalletAddress)
	if recoveredAddr != claimedAddr {
		return common.Address{}, fmt.Errorf("signature address mismatch: claimed %s, recovered %s",
			claimedAddr.Hex(), recoveredAddr.Hex())
	}

	return recoveredAddr, nil
}

// ethPersonalHash computes the EIP-191 personal_sign hash of a message.
func ethPersonalHash(message string) []byte {
	prefixed := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)
	return crypto.Keccak256([]byte(prefixed))
}
