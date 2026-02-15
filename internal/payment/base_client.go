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
	"github.com/moltbunker/moltbunker/internal/logging"
	"github.com/moltbunker/moltbunker/internal/util"
)

// BaseClientConfig holds configuration for the Base network client
type BaseClientConfig struct {
	RPCURL             string
	WSEndpoint         string
	RPCURLs            []string // Additional RPC endpoints for failover
	WSEndpoints        []string // Additional WS endpoints for failover
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

	// RPC failover
	rpcTracker *EndpointTracker
	wsTracker  *EndpointTracker
	activeRPC  string
	activeWS   string

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

	// Build RPC endpoint list (primary + extras)
	rpcURLs := mergeEndpoints(config.RPCURL, config.RPCURLs)
	wsURLs := mergeEndpoints(config.WSEndpoint, config.WSEndpoints)

	if len(rpcURLs) > 0 {
		bc.rpcTracker = NewEndpointTracker(rpcURLs)
	}
	if len(wsURLs) > 0 {
		bc.wsTracker = NewEndpointTracker(wsURLs)
	}

	return bc, nil
}

// mergeEndpoints combines a primary endpoint with a list, deduplicating.
func mergeEndpoints(primary string, extras []string) []string {
	seen := make(map[string]bool)
	var result []string
	if primary != "" {
		result = append(result, primary)
		seen[primary] = true
	}
	for _, u := range extras {
		if u != "" && !seen[u] {
			result = append(result, u)
			seen[u] = true
		}
	}
	return result
}

// Connect establishes connection to Base network.
// Tries endpoints from the RPC tracker in health-sorted order.
func (bc *BaseClient) Connect(ctx context.Context) error {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	// Build ordered list of RPC URLs to try
	var rpcURLs []string
	if bc.rpcTracker != nil {
		rpcURLs = bc.rpcTracker.GetHealthy()
	}
	if len(rpcURLs) == 0 && bc.config.RPCURL != "" {
		rpcURLs = []string{bc.config.RPCURL}
	}

	// Try each RPC endpoint until one works
	var lastErr error
	for _, rpcURL := range rpcURLs {
		start := time.Now()
		client, result := util.RetryWithValue(ctx, bc.config.RetryConfig, func() (*ethclient.Client, error) {
			return ethclient.DialContext(ctx, rpcURL)
		})
		if result.LastError != nil {
			lastErr = result.LastError
			if bc.rpcTracker != nil {
				bc.rpcTracker.RecordError(rpcURL)
			}
			logging.Warn("RPC endpoint failed, trying next",
				"url", rpcURL, "error", result.LastError,
				logging.Component("payment"))
			continue
		}

		// Verify chain ID before accepting this endpoint
		chainID, err := client.ChainID(ctx)
		if err != nil {
			client.Close()
			lastErr = err
			if bc.rpcTracker != nil {
				bc.rpcTracker.RecordError(rpcURL)
			}
			continue
		}
		if chainID.Cmp(bc.chainID) != 0 {
			client.Close()
			lastErr = fmt.Errorf("chain ID mismatch: expected %d, got %d", bc.chainID, chainID)
			if bc.rpcTracker != nil {
				bc.rpcTracker.RecordError(rpcURL)
			}
			continue
		}

		bc.client = client
		bc.activeRPC = rpcURL
		if bc.rpcTracker != nil {
			bc.rpcTracker.RecordSuccess(rpcURL, time.Since(start))
		}
		logging.Info("connected to RPC endpoint", "url", rpcURL, logging.Component("payment"))
		break
	}

	if bc.client == nil {
		if lastErr != nil {
			return fmt.Errorf("failed to connect to any RPC endpoint: %w", lastErr)
		}
		return fmt.Errorf("no RPC endpoints configured")
	}

	// Connect via WebSocket for subscriptions (optional, try all WS endpoints)
	bc.connectWS(ctx)

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

// connectWS attempts to connect to a WebSocket endpoint.
// Tries endpoints from the WS tracker in health-sorted order.
// Must be called with bc.mu held.
func (bc *BaseClient) connectWS(ctx context.Context) {
	var wsURLs []string
	if bc.wsTracker != nil {
		wsURLs = bc.wsTracker.GetHealthy()
	}
	if len(wsURLs) == 0 && bc.config.WSEndpoint != "" {
		wsURLs = []string{bc.config.WSEndpoint}
	}

	for _, wsURL := range wsURLs {
		start := time.Now()
		wsClient, err := ethclient.DialContext(ctx, wsURL)
		if err != nil {
			if bc.wsTracker != nil {
				bc.wsTracker.RecordError(wsURL)
			}
			logging.Warn("WebSocket endpoint failed, trying next",
				"url", wsURL, "error", err,
				logging.Component("payment"))
			continue
		}
		bc.wsClient = wsClient
		bc.activeWS = wsURL
		if bc.wsTracker != nil {
			bc.wsTracker.RecordSuccess(wsURL, time.Since(start))
		}
		logging.Info("connected to WebSocket endpoint", "url", wsURL, logging.Component("payment"))
		return
	}

	if len(wsURLs) > 0 {
		logging.Warn("failed to connect to any WebSocket endpoint", logging.Component("payment"))
	}
}

// switchRPC connects to the next healthy RPC endpoint.
// Returns an error if no alternative is available.
func (bc *BaseClient) switchRPC(ctx context.Context) error {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	if bc.rpcTracker == nil {
		return fmt.Errorf("no RPC failover configured")
	}

	nextURL, ok := bc.rpcTracker.GetNext(bc.activeRPC)
	if !ok {
		return fmt.Errorf("no alternative RPC endpoint available")
	}

	// Close old client
	if bc.client != nil {
		bc.client.Close()
		bc.client = nil
	}

	start := time.Now()
	client, err := ethclient.DialContext(ctx, nextURL)
	if err != nil {
		bc.rpcTracker.RecordError(nextURL)
		bc.connected = false
		return fmt.Errorf("failed to connect to alternative RPC %s: %w", nextURL, err)
	}

	bc.client = client
	bc.activeRPC = nextURL
	bc.rpcTracker.RecordSuccess(nextURL, time.Since(start))
	bc.connected = true

	// Re-sync nonce with new endpoint
	if bc.privateKey != nil {
		nonce, err := bc.client.PendingNonceAt(ctx, bc.address)
		if err != nil {
			logging.Warn("failed to sync nonce after RPC switch",
				"error", err, logging.Component("payment"))
		} else {
			bc.nonceMu.Lock()
			bc.pendingNonce = nonce
			bc.nonceMu.Unlock()
		}
	}

	logging.Info("switched to alternative RPC endpoint",
		"url", nextURL, logging.Component("payment"))
	return nil
}

// ReconnectWS reconnects to the next healthy WebSocket endpoint.
func (bc *BaseClient) ReconnectWS(ctx context.Context) error {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	// Close old WS client
	if bc.wsClient != nil {
		bc.wsClient.Close()
		bc.wsClient = nil
		bc.activeWS = ""
	}

	if bc.wsTracker != nil {
		// Mark current as errored
		if bc.activeWS != "" {
			bc.wsTracker.RecordError(bc.activeWS)
		}
	}

	bc.connectWS(ctx)

	if bc.wsClient == nil {
		return fmt.Errorf("failed to reconnect to any WebSocket endpoint")
	}
	return nil
}

// callWithFailover executes fn. On connection-level error, switches RPC and retries once.
func (bc *BaseClient) callWithFailover(ctx context.Context, fn func(client *ethclient.Client) error) error {
	bc.mu.RLock()
	client := bc.client
	activeRPC := bc.activeRPC
	bc.mu.RUnlock()

	if client == nil {
		return fmt.Errorf("not connected")
	}

	err := fn(client)
	if err == nil {
		if bc.rpcTracker != nil && activeRPC != "" {
			bc.rpcTracker.RecordSuccess(activeRPC, 0) // latency not tracked per-call
		}
		return nil
	}

	// Check if this is a connection-level error worth failing over
	if !isConnectionError(err) {
		return err
	}

	if bc.rpcTracker != nil && activeRPC != "" {
		bc.rpcTracker.RecordError(activeRPC)
	}

	// Try to switch to another endpoint
	if switchErr := bc.switchRPC(ctx); switchErr != nil {
		return fmt.Errorf("%w (failover also failed: %v)", err, switchErr)
	}

	// Retry with new client
	bc.mu.RLock()
	client = bc.client
	bc.mu.RUnlock()

	return fn(client)
}

// isConnectionError returns true if the error indicates a connection problem
// (as opposed to a contract revert or validation error).
func isConnectionError(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	for _, substr := range []string{
		"connection refused",
		"connection reset",
		"EOF",
		"broken pipe",
		"no such host",
		"i/o timeout",
		"TLS handshake timeout",
		"context deadline exceeded",
		"server returned HTTP status 502",
		"server returned HTTP status 503",
		"server returned HTTP status 429",
	} {
		if contains(s, substr) {
			return true
		}
	}
	return false
}

// contains is a simple substring check to avoid importing strings.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchSubstring(s, substr)
}

func searchSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// HasWSConfig returns true if any WebSocket endpoint is configured.
func (bc *BaseClient) HasWSConfig() bool {
	return bc.wsTracker != nil
}

// ActiveRPC returns the currently active RPC URL.
func (bc *BaseClient) ActiveRPC() string {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return bc.activeRPC
}

// ActiveWS returns the currently active WebSocket URL.
func (bc *BaseClient) ActiveWS() string {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return bc.activeWS
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

// ChainID returns a copy of the chain ID
func (bc *BaseClient) ChainID() *big.Int {
	return new(big.Int).Set(bc.chainID)
}

// GetBalance returns the ETH balance
func (bc *BaseClient) GetBalance(ctx context.Context, address common.Address) (*big.Int, error) {
	var balance *big.Int
	err := bc.callWithFailover(ctx, func(client *ethclient.Client) error {
		bal, result := util.RetryWithValue(ctx, bc.config.RetryConfig, func() (*big.Int, error) {
			return client.BalanceAt(ctx, address, nil)
		})
		if result.LastError != nil {
			return fmt.Errorf("failed to get balance: %w", result.LastError)
		}
		balance = bal
		return nil
	})
	return balance, err
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

	// Cap gas price (copy to avoid aliasing the config value)
	if bc.config.MaxGasPrice != nil && gasPrice.Cmp(bc.config.MaxGasPrice) > 0 {
		gasPrice = new(big.Int).Set(bc.config.MaxGasPrice)
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
		// TX may not have been submitted — resync nonce to prevent gap
		if syncErr := bc.SyncNonce(ctx); syncErr != nil {
			logging.Warn("failed to sync nonce after wait failure",
				"error", syncErr,
				logging.Component("payment"))
		}
		return nil, fmt.Errorf("failed waiting for transaction: %w", err)
	}

	if receipt.Status == types.ReceiptStatusFailed {
		// Re-sync nonce with network after TX failure to prevent nonce desync
		if syncErr := bc.SyncNonce(ctx); syncErr != nil {
			logging.Warn("failed to sync nonce after TX failure",
				"tx_hash", tx.Hash().Hex(),
				"error", syncErr,
				logging.Component("payment"))
		}
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
	var adjustedGas uint64
	err := bc.callWithFailover(ctx, func(client *ethclient.Client) error {
		gas, err := client.EstimateGas(ctx, msg)
		if err != nil {
			return fmt.Errorf("failed to estimate gas: %w", err)
		}
		adjustedGas = uint64(float64(gas) * bc.config.GasLimitMultiplier)
		return nil
	})
	return adjustedGas, err
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
	var blockNum uint64
	err := bc.callWithFailover(ctx, func(client *ethclient.Client) error {
		n, err := client.BlockNumber(ctx)
		if err != nil {
			return err
		}
		blockNum = n
		return nil
	})
	return blockNum, err
}
