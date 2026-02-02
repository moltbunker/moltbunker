package payment

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/moltbunker/moltbunker/internal/util"
)

// BaseClientConfig holds configuration for the Base network client
type BaseClientConfig struct {
	RPCURL             string
	WSEndpoint         string
	ChainID            int64
	BlockConfirmations int
	GasLimitMultiplier float64 // Multiplier for estimated gas (default: 1.2)
	MaxGasPrice        *big.Int
	RetryConfig        *util.RetryConfig
}

// DefaultBaseClientConfig returns sensible defaults
func DefaultBaseClientConfig() *BaseClientConfig {
	return &BaseClientConfig{
		RPCURL:             "https://mainnet.base.org",
		WSEndpoint:         "wss://mainnet.base.org",
		ChainID:            8453, // Base mainnet
		BlockConfirmations: 12,
		GasLimitMultiplier: 1.2,
		MaxGasPrice:        big.NewInt(100e9), // 100 gwei max
		RetryConfig:        util.DefaultRetryConfig(),
	}
}

// BaseClient provides access to the Base L2 network
type BaseClient struct {
	config     *BaseClientConfig
	client     *ethclient.Client
	wsClient   *ethclient.Client
	privateKey *ecdsa.PrivateKey
	address    common.Address
	chainID    *big.Int

	// Nonce management
	nonceMu     sync.Mutex
	pendingNonce uint64

	// Connection state
	connected bool
	mu        sync.RWMutex
}

// NewBaseClient creates a new Base network client
func NewBaseClient(config *BaseClientConfig, privateKey *ecdsa.PrivateKey) (*BaseClient, error) {
	if config == nil {
		config = DefaultBaseClientConfig()
	}

	bc := &BaseClient{
		config:     config,
		privateKey: privateKey,
		chainID:    big.NewInt(config.ChainID),
	}

	if privateKey != nil {
		bc.address = crypto.PubkeyToAddress(privateKey.PublicKey)
	}

	return bc, nil
}

// Connect establishes connection to Base network
func (bc *BaseClient) Connect(ctx context.Context) error {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	var err error

	// Connect via HTTP RPC
	_, result := util.RetryWithValue(ctx, bc.config.RetryConfig, func() (*ethclient.Client, error) {
		return ethclient.DialContext(ctx, bc.config.RPCURL)
	})
	if result.LastError != nil {
		return fmt.Errorf("failed to connect to Base RPC: %w", result.LastError)
	}
	bc.client, _ = util.RetryWithValue(ctx, bc.config.RetryConfig, func() (*ethclient.Client, error) {
		return ethclient.DialContext(ctx, bc.config.RPCURL)
	})

	// Connect via WebSocket for subscriptions (optional)
	if bc.config.WSEndpoint != "" {
		bc.wsClient, err = ethclient.DialContext(ctx, bc.config.WSEndpoint)
		if err != nil {
			// WebSocket is optional, log but don't fail
			fmt.Printf("Warning: Failed to connect to WebSocket endpoint: %v\n", err)
		}
	}

	// Verify chain ID
	chainID, err := bc.client.ChainID(ctx)
	if err != nil {
		return fmt.Errorf("failed to get chain ID: %w", err)
	}
	if chainID.Cmp(bc.chainID) != 0 {
		return fmt.Errorf("chain ID mismatch: expected %d, got %d", bc.chainID, chainID)
	}

	// Initialize nonce
	if bc.privateKey != nil {
		nonce, err := bc.client.PendingNonceAt(ctx, bc.address)
		if err != nil {
			return fmt.Errorf("failed to get nonce: %w", err)
		}
		bc.pendingNonce = nonce
	}

	bc.connected = true
	return nil
}

// Close closes the connection
func (bc *BaseClient) Close() {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	if bc.client != nil {
		bc.client.Close()
		bc.client = nil
	}
	if bc.wsClient != nil {
		bc.wsClient.Close()
		bc.wsClient = nil
	}
	bc.connected = false
}

// IsConnected returns true if connected to Base network
func (bc *BaseClient) IsConnected() bool {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return bc.connected
}

// Client returns the underlying ethclient
func (bc *BaseClient) Client() *ethclient.Client {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return bc.client
}

// WSClient returns the WebSocket client for subscriptions
func (bc *BaseClient) WSClient() *ethclient.Client {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return bc.wsClient
}

// Address returns the wallet address
func (bc *BaseClient) Address() common.Address {
	return bc.address
}

// ChainID returns the chain ID
func (bc *BaseClient) ChainID() *big.Int {
	return bc.chainID
}

// GetBalance returns the ETH balance
func (bc *BaseClient) GetBalance(ctx context.Context, address common.Address) (*big.Int, error) {
	bc.mu.RLock()
	client := bc.client
	bc.mu.RUnlock()

	if client == nil {
		return nil, fmt.Errorf("not connected")
	}

	balance, result := util.RetryWithValue(ctx, bc.config.RetryConfig, func() (*big.Int, error) {
		return client.BalanceAt(ctx, address, nil)
	})
	if result.LastError != nil {
		return nil, fmt.Errorf("failed to get balance: %w", result.LastError)
	}

	return balance, nil
}

// GetTransactOpts creates transaction options for signing
func (bc *BaseClient) GetTransactOpts(ctx context.Context) (*bind.TransactOpts, error) {
	if bc.privateKey == nil {
		return nil, fmt.Errorf("no private key configured")
	}

	bc.mu.RLock()
	client := bc.client
	bc.mu.RUnlock()

	if client == nil {
		return nil, fmt.Errorf("not connected")
	}

	// Get suggested gas price
	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get gas price: %w", err)
	}

	// Cap gas price
	if bc.config.MaxGasPrice != nil && gasPrice.Cmp(bc.config.MaxGasPrice) > 0 {
		gasPrice = bc.config.MaxGasPrice
	}

	// Create transaction opts
	auth, err := bind.NewKeyedTransactorWithChainID(bc.privateKey, bc.chainID)
	if err != nil {
		return nil, fmt.Errorf("failed to create transactor: %w", err)
	}

	auth.Context = ctx
	auth.GasPrice = gasPrice

	// Set nonce
	bc.nonceMu.Lock()
	auth.Nonce = big.NewInt(int64(bc.pendingNonce))
	bc.pendingNonce++
	bc.nonceMu.Unlock()

	return auth, nil
}

// WaitForTransaction waits for a transaction to be mined and confirmed
func (bc *BaseClient) WaitForTransaction(ctx context.Context, tx *types.Transaction) (*types.Receipt, error) {
	bc.mu.RLock()
	client := bc.client
	bc.mu.RUnlock()

	if client == nil {
		return nil, fmt.Errorf("not connected")
	}

	// Wait for mining
	receipt, err := bind.WaitMined(ctx, client, tx)
	if err != nil {
		return nil, fmt.Errorf("failed waiting for transaction: %w", err)
	}

	if receipt.Status == types.ReceiptStatusFailed {
		return receipt, fmt.Errorf("transaction failed: %s", tx.Hash().Hex())
	}

	// Wait for confirmations
	if bc.config.BlockConfirmations > 0 {
		targetBlock := receipt.BlockNumber.Uint64() + uint64(bc.config.BlockConfirmations)

		for {
			select {
			case <-ctx.Done():
				return receipt, ctx.Err()
			case <-time.After(2 * time.Second):
				currentBlock, err := client.BlockNumber(ctx)
				if err != nil {
					continue // Retry
				}
				if currentBlock >= targetBlock {
					return receipt, nil
				}
			}
		}
	}

	return receipt, nil
}

// EstimateGas estimates gas for a transaction
func (bc *BaseClient) EstimateGas(ctx context.Context, msg ethereum.CallMsg) (uint64, error) {
	bc.mu.RLock()
	client := bc.client
	bc.mu.RUnlock()

	if client == nil {
		return 0, fmt.Errorf("not connected")
	}

	gas, err := client.EstimateGas(ctx, msg)
	if err != nil {
		return 0, fmt.Errorf("failed to estimate gas: %w", err)
	}

	// Apply multiplier for safety margin
	adjustedGas := uint64(float64(gas) * bc.config.GasLimitMultiplier)
	return adjustedGas, nil
}

// SyncNonce synchronizes the nonce with the network
func (bc *BaseClient) SyncNonce(ctx context.Context) error {
	bc.mu.RLock()
	client := bc.client
	bc.mu.RUnlock()

	if client == nil {
		return fmt.Errorf("not connected")
	}

	nonce, err := client.PendingNonceAt(ctx, bc.address)
	if err != nil {
		return fmt.Errorf("failed to get nonce: %w", err)
	}

	bc.nonceMu.Lock()
	bc.pendingNonce = nonce
	bc.nonceMu.Unlock()

	return nil
}

// GetBlockNumber returns the current block number
func (bc *BaseClient) GetBlockNumber(ctx context.Context) (uint64, error) {
	bc.mu.RLock()
	client := bc.client
	bc.mu.RUnlock()

	if client == nil {
		return 0, fmt.Errorf("not connected")
	}

	return client.BlockNumber(ctx)
}
