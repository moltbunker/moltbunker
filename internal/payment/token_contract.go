package payment

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// TokenContract provides interface to the BUNKER ERC20 token contract
type TokenContract struct {
	baseClient   *BaseClient
	contract     *bind.BoundContract
	contractABI  abi.ABI
	contractAddr common.Address
	mockMode     bool

	// Mock state
	mockBalances   map[common.Address]*big.Int
	mockAllowances map[common.Address]map[common.Address]*big.Int
	mockMu         sync.RWMutex
}

// NewTokenContract creates a new token contract client
func NewTokenContract(baseClient *BaseClient, contractAddr common.Address) (*TokenContract, error) {
	tc := &TokenContract{
		baseClient:     baseClient,
		contractAddr:   contractAddr,
		mockBalances:   make(map[common.Address]*big.Int),
		mockAllowances: make(map[common.Address]map[common.Address]*big.Int),
	}

	// If no base client, use mock mode
	if baseClient == nil || !baseClient.IsConnected() {
		tc.mockMode = true
		return tc, nil
	}

	parsedABI, err := abi.JSON(strings.NewReader(BunkerTokenABI))
	if err != nil {
		return nil, fmt.Errorf("failed to parse token ABI: %w", err)
	}
	tc.contractABI = parsedABI

	client := baseClient.Client()
	tc.contract = bind.NewBoundContract(contractAddr, parsedABI, client, client, client)

	return tc, nil
}

// NewMockTokenContract creates a mock token contract for testing
func NewMockTokenContract() *TokenContract {
	return &TokenContract{
		mockMode:       true,
		mockBalances:   make(map[common.Address]*big.Int),
		mockAllowances: make(map[common.Address]map[common.Address]*big.Int),
	}
}

// IsMockMode returns whether running in mock mode
func (tc *TokenContract) IsMockMode() bool {
	return tc.mockMode
}

// BalanceOf returns the token balance for an address
func (tc *TokenContract) BalanceOf(ctx context.Context, account common.Address) (*big.Int, error) {
	if tc.mockMode {
		tc.mockMu.RLock()
		defer tc.mockMu.RUnlock()
		balance, exists := tc.mockBalances[account]
		if !exists {
			return big.NewInt(0), nil
		}
		return new(big.Int).Set(balance), nil
	}

	var result []interface{}
	err := tc.contract.Call(&bind.CallOpts{Context: ctx}, &result, "balanceOf", account)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}

	if len(result) == 0 {
		return big.NewInt(0), nil
	}
	if balance, ok := result[0].(*big.Int); ok {
		return balance, nil
	}
	return big.NewInt(0), nil
}

// Allowance returns the allowance for a spender
func (tc *TokenContract) Allowance(ctx context.Context, owner, spender common.Address) (*big.Int, error) {
	if tc.mockMode {
		tc.mockMu.RLock()
		defer tc.mockMu.RUnlock()
		if ownerAllowances, exists := tc.mockAllowances[owner]; exists {
			if allowance, ok := ownerAllowances[spender]; ok {
				return new(big.Int).Set(allowance), nil
			}
		}
		return big.NewInt(0), nil
	}

	var result []interface{}
	err := tc.contract.Call(&bind.CallOpts{Context: ctx}, &result, "allowance", owner, spender)
	if err != nil {
		return nil, fmt.Errorf("failed to get allowance: %w", err)
	}

	if len(result) == 0 {
		return big.NewInt(0), nil
	}
	if allowance, ok := result[0].(*big.Int); ok {
		return allowance, nil
	}
	return big.NewInt(0), nil
}

// Approve approves a spender to spend tokens
func (tc *TokenContract) Approve(ctx context.Context, spender common.Address, amount *big.Int) (*types.Transaction, error) {
	if tc.mockMode {
		return tc.mockApprove(ctx, spender, amount)
	}

	auth, err := tc.baseClient.GetTransactOpts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction options: %w", err)
	}

	tx, err := tc.contract.Transact(auth, "approve", spender, amount)
	if err != nil {
		return nil, fmt.Errorf("failed to approve: %w", err)
	}

	return tx, nil
}

// mockApprove handles approval in mock mode
func (tc *TokenContract) mockApprove(_ context.Context, spender common.Address, amount *big.Int) (*types.Transaction, error) {
	tc.mockMu.Lock()
	defer tc.mockMu.Unlock()

	var owner common.Address
	if tc.baseClient != nil {
		owner = tc.baseClient.Address()
	}
	if _, exists := tc.mockAllowances[owner]; !exists {
		tc.mockAllowances[owner] = make(map[common.Address]*big.Int)
	}
	tc.mockAllowances[owner][spender] = new(big.Int).Set(amount)

	fmt.Printf("[MOCK] Approved %s to spend %s BUNKER tokens for %s\n",
		spender.Hex(), amount.String(), owner.Hex())

	return nil, nil
}

// ApproveAndWait approves and waits for confirmation
func (tc *TokenContract) ApproveAndWait(ctx context.Context, spender common.Address, amount *big.Int) (*types.Receipt, error) {
	tx, err := tc.Approve(ctx, spender, amount)
	if err != nil {
		return nil, err
	}

	if tc.mockMode || tx == nil {
		return nil, nil
	}

	return tc.baseClient.WaitForTransaction(ctx, tx)
}

// Transfer transfers tokens to an address
func (tc *TokenContract) Transfer(ctx context.Context, to common.Address, amount *big.Int) (*types.Transaction, error) {
	if tc.mockMode {
		return tc.mockTransfer(ctx, to, amount)
	}

	auth, err := tc.baseClient.GetTransactOpts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction options: %w", err)
	}

	tx, err := tc.contract.Transact(auth, "transfer", to, amount)
	if err != nil {
		return nil, fmt.Errorf("failed to transfer: %w", err)
	}

	return tx, nil
}

// mockTransfer handles transfer in mock mode
func (tc *TokenContract) mockTransfer(_ context.Context, to common.Address, amount *big.Int) (*types.Transaction, error) {
	tc.mockMu.Lock()
	defer tc.mockMu.Unlock()

	var from common.Address
	if tc.baseClient != nil {
		from = tc.baseClient.Address()
	}
	balance, exists := tc.mockBalances[from]
	if !exists || balance.Cmp(amount) < 0 {
		return nil, fmt.Errorf("insufficient balance")
	}

	tc.mockBalances[from] = new(big.Int).Sub(balance, amount)

	toBalance, exists := tc.mockBalances[to]
	if !exists {
		tc.mockBalances[to] = new(big.Int).Set(amount)
	} else {
		tc.mockBalances[to] = new(big.Int).Add(toBalance, amount)
	}

	fmt.Printf("[MOCK] Transferred %s BUNKER from %s to %s\n",
		amount.String(), from.Hex(), to.Hex())

	return nil, nil
}

// TransferAndWait transfers and waits for confirmation
func (tc *TokenContract) TransferAndWait(ctx context.Context, to common.Address, amount *big.Int) (*types.Receipt, error) {
	tx, err := tc.Transfer(ctx, to, amount)
	if err != nil {
		return nil, err
	}

	if tc.mockMode || tx == nil {
		return nil, nil
	}

	return tc.baseClient.WaitForTransaction(ctx, tx)
}

// TransferFrom transfers tokens from one address to another (requires allowance)
func (tc *TokenContract) TransferFrom(ctx context.Context, from, to common.Address, amount *big.Int) (*types.Transaction, error) {
	if tc.mockMode {
		return tc.mockTransferFrom(ctx, from, to, amount)
	}

	auth, err := tc.baseClient.GetTransactOpts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction options: %w", err)
	}

	tx, err := tc.contract.Transact(auth, "transferFrom", from, to, amount)
	if err != nil {
		return nil, fmt.Errorf("failed to transferFrom: %w", err)
	}

	return tx, nil
}

// mockTransferFrom handles transferFrom in mock mode
func (tc *TokenContract) mockTransferFrom(_ context.Context, from, to common.Address, amount *big.Int) (*types.Transaction, error) {
	tc.mockMu.Lock()
	defer tc.mockMu.Unlock()

	var spender common.Address
	if tc.baseClient != nil {
		spender = tc.baseClient.Address()
	}

	// Check allowance
	allowance := big.NewInt(0)
	if ownerAllowances, exists := tc.mockAllowances[from]; exists {
		if a, ok := ownerAllowances[spender]; ok {
			allowance = a
		}
	}
	if allowance.Cmp(amount) < 0 {
		return nil, fmt.Errorf("insufficient allowance")
	}

	// Check balance
	balance, exists := tc.mockBalances[from]
	if !exists || balance.Cmp(amount) < 0 {
		return nil, fmt.Errorf("insufficient balance")
	}

	// Perform transfer
	tc.mockBalances[from] = new(big.Int).Sub(balance, amount)
	tc.mockAllowances[from][spender] = new(big.Int).Sub(allowance, amount)

	toBalance, exists := tc.mockBalances[to]
	if !exists {
		tc.mockBalances[to] = new(big.Int).Set(amount)
	} else {
		tc.mockBalances[to] = new(big.Int).Add(toBalance, amount)
	}

	fmt.Printf("[MOCK] TransferFrom %s BUNKER from %s to %s (spender: %s)\n",
		amount.String(), from.Hex(), to.Hex(), spender.Hex())

	return nil, nil
}

// SetMockBalance sets a mock balance for testing
func (tc *TokenContract) SetMockBalance(account common.Address, amount *big.Int) {
	if !tc.mockMode {
		return
	}
	tc.mockMu.Lock()
	defer tc.mockMu.Unlock()
	tc.mockBalances[account] = new(big.Int).Set(amount)
}

// FormatTokenAmount formats a token amount for display (18 decimals)
func FormatTokenAmount(amount *big.Int) string {
	if amount == nil {
		return "0"
	}

	decimals := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	whole := new(big.Int).Div(amount, decimals)
	frac := new(big.Int).Mod(amount, decimals)

	if frac.Sign() == 0 {
		return whole.String()
	}

	// Format with up to 4 decimal places
	fracStr := frac.String()
	for len(fracStr) < 18 {
		fracStr = "0" + fracStr
	}
	// Trim trailing zeros
	for len(fracStr) > 0 && fracStr[len(fracStr)-1] == '0' {
		fracStr = fracStr[:len(fracStr)-1]
	}
	if len(fracStr) > 4 {
		fracStr = fracStr[:4]
	}

	if len(fracStr) == 0 {
		return whole.String()
	}
	return whole.String() + "." + fracStr
}

// ParseTokenAmount parses a token amount string to wei
func ParseTokenAmount(amount string) (*big.Int, error) {
	// Simple parsing - supports whole numbers and decimals
	parts := strings.Split(amount, ".")

	whole, ok := new(big.Int).SetString(parts[0], 10)
	if !ok {
		return nil, fmt.Errorf("invalid amount: %s", amount)
	}

	decimals := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	result := new(big.Int).Mul(whole, decimals)

	if len(parts) == 2 {
		fracStr := parts[1]
		// Pad or truncate to 18 decimals
		for len(fracStr) < 18 {
			fracStr += "0"
		}
		if len(fracStr) > 18 {
			fracStr = fracStr[:18]
		}
		frac, ok := new(big.Int).SetString(fracStr, 10)
		if !ok {
			return nil, fmt.Errorf("invalid decimal: %s", parts[1])
		}
		result.Add(result, frac)
	}

	return result, nil
}
