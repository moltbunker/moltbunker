package payment

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	pkgtypes "github.com/moltbunker/moltbunker/pkg/types"
)

// PaymentService provides a unified interface for all payment operations
type PaymentService struct {
	config          *PaymentServiceConfig
	baseClient      *BaseClient
	tokenContract   *TokenContract
	stakingContract *StakingContract
	escrowContract  *EscrowContract
	slashingContract *SlashingContract

	// Pricing calculator
	pricingCalculator *PricingCalculator

	// Service state
	started bool
	mu      sync.RWMutex
}

// PaymentServiceConfig holds configuration for the payment service
type PaymentServiceConfig struct {
	// Base network settings
	RPCURL             string
	WSEndpoint         string
	ChainID            int64
	BlockConfirmations int

	// Contract addresses
	TokenAddress     common.Address
	RegistryAddress  common.Address
	EscrowAddress    common.Address
	StakingAddress   common.Address
	SlashingAddress  common.Address

	// Wallet (nil for read-only mode)
	PrivateKey *ecdsa.PrivateKey

	// Pricing
	BasePricePerHour *big.Int

	// Mock mode
	MockMode bool
}

// NewPaymentService creates a new payment service
func NewPaymentService(config *PaymentServiceConfig) (*PaymentService, error) {
	ps := &PaymentService{
		config: config,
	}

	if config.MockMode {
		return ps.setupMockMode()
	}

	// Create base client
	baseConfig := &BaseClientConfig{
		RPCURL:             config.RPCURL,
		WSEndpoint:         config.WSEndpoint,
		ChainID:            config.ChainID,
		BlockConfirmations: config.BlockConfirmations,
	}

	var err error
	ps.baseClient, err = NewBaseClient(baseConfig, config.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create base client: %w", err)
	}

	// Create token contract
	ps.tokenContract, err = NewTokenContract(ps.baseClient, config.TokenAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to create token contract: %w", err)
	}

	// Create staking contract
	ps.stakingContract, err = NewStakingContract(ps.baseClient, ps.tokenContract, config.StakingAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to create staking contract: %w", err)
	}

	// Create escrow contract
	ps.escrowContract, err = NewEscrowContract(ps.baseClient, ps.tokenContract, config.EscrowAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to create escrow contract: %w", err)
	}

	// Create slashing contract
	ps.slashingContract, err = NewSlashingContract(ps.baseClient, ps.stakingContract, config.SlashingAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to create slashing contract: %w", err)
	}

	// Create pricing calculator
	basePricePerHour := config.BasePricePerHour
	if basePricePerHour == nil {
		basePricePerHour = parseWei("1") // 1 BUNKER per hour default
	}
	ps.pricingCalculator = NewPricingCalculator(basePricePerHour)

	return ps, nil
}

// setupMockMode sets up the service in mock mode
func (ps *PaymentService) setupMockMode() (*PaymentService, error) {
	ps.tokenContract = NewMockTokenContract()
	ps.stakingContract = NewMockStakingContract()
	ps.escrowContract = NewMockEscrowContract()
	ps.slashingContract = NewMockSlashingContract()

	basePricePerHour := ps.config.BasePricePerHour
	if basePricePerHour == nil {
		basePricePerHour = parseWei("1")
	}
	ps.pricingCalculator = NewPricingCalculator(basePricePerHour)

	return ps, nil
}

// Start starts the payment service
func (ps *PaymentService) Start(ctx context.Context) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if ps.started {
		return nil
	}

	// Connect to Base network (if not mock mode)
	if ps.baseClient != nil {
		if err := ps.baseClient.Connect(ctx); err != nil {
			return fmt.Errorf("failed to connect to Base network: %w", err)
		}
	}

	ps.started = true
	return nil
}

// Stop stops the payment service
func (ps *PaymentService) Stop() {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if !ps.started {
		return
	}

	if ps.baseClient != nil {
		ps.baseClient.Close()
	}

	ps.started = false
}

// IsConnected returns true if connected to Base network
func (ps *PaymentService) IsConnected() bool {
	if ps.config.MockMode {
		return true
	}
	if ps.baseClient == nil {
		return false
	}
	return ps.baseClient.IsConnected()
}

// IsMockMode returns true if running in mock mode
func (ps *PaymentService) IsMockMode() bool {
	return ps.config.MockMode
}

// GetWalletAddress returns the wallet address
func (ps *PaymentService) GetWalletAddress() common.Address {
	if ps.baseClient != nil {
		return ps.baseClient.Address()
	}
	if ps.config.PrivateKey != nil {
		return crypto.PubkeyToAddress(ps.config.PrivateKey.PublicKey)
	}
	return common.Address{}
}

// ===== Token Operations =====

// GetTokenBalance returns the BUNKER token balance
func (ps *PaymentService) GetTokenBalance(ctx context.Context, address common.Address) (*big.Int, error) {
	return ps.tokenContract.BalanceOf(ctx, address)
}

// GetETHBalance returns the ETH balance (for gas)
func (ps *PaymentService) GetETHBalance(ctx context.Context, address common.Address) (*big.Int, error) {
	if ps.baseClient == nil {
		return big.NewInt(0), nil
	}
	return ps.baseClient.GetBalance(ctx, address)
}

// TransferTokens transfers BUNKER tokens
func (ps *PaymentService) TransferTokens(ctx context.Context, to common.Address, amount *big.Int) error {
	_, err := ps.tokenContract.TransferAndWait(ctx, to, amount)
	return err
}

// ===== Staking Operations =====

// Stake stakes BUNKER tokens
func (ps *PaymentService) Stake(ctx context.Context, amount *big.Int) error {
	_, err := ps.stakingContract.StakeAndWait(ctx, amount)
	return err
}

// RequestUnstake requests to unstake tokens
func (ps *PaymentService) RequestUnstake(ctx context.Context, amount *big.Int) error {
	_, err := ps.stakingContract.RequestUnstake(ctx, amount)
	return err
}

// Withdraw withdraws unstaked tokens
func (ps *PaymentService) Withdraw(ctx context.Context) error {
	_, err := ps.stakingContract.Withdraw(ctx)
	return err
}

// GetStakeInfo returns staking information for a provider
func (ps *PaymentService) GetStakeInfo(ctx context.Context, provider common.Address) (*StakeInfo, error) {
	return ps.stakingContract.GetStakeInfo(ctx, provider)
}

// GetTier returns the staking tier for a provider
func (ps *PaymentService) GetTier(ctx context.Context, provider common.Address) (pkgtypes.StakingTier, error) {
	return ps.stakingContract.GetTier(ctx, provider)
}

// HasMinimumStake checks if provider has minimum stake
func (ps *PaymentService) HasMinimumStake(ctx context.Context, provider common.Address) (bool, error) {
	return ps.stakingContract.HasMinimumStake(ctx, provider)
}

// ===== Escrow Operations =====

// CreateJobEscrow creates an escrow for a job
func (ps *PaymentService) CreateJobEscrow(ctx context.Context, jobID [32]byte, provider common.Address, amount *big.Int, duration time.Duration) error {
	durationSecs := big.NewInt(int64(duration.Seconds()))
	_, err := ps.escrowContract.CreateEscrowAndWait(ctx, jobID, provider, amount, durationSecs)
	return err
}

// ReleaseJobPayment releases payment for a job based on uptime
func (ps *PaymentService) ReleaseJobPayment(ctx context.Context, jobID [32]byte, uptime time.Duration) error {
	uptimeSecs := big.NewInt(int64(uptime.Seconds()))
	_, err := ps.escrowContract.ReleasePayment(ctx, jobID, uptimeSecs)
	return err
}

// RefundJob refunds remaining escrow to requester
func (ps *PaymentService) RefundJob(ctx context.Context, jobID [32]byte) error {
	_, err := ps.escrowContract.Refund(ctx, jobID)
	return err
}

// FinalizeJob finalizes a job escrow
func (ps *PaymentService) FinalizeJob(ctx context.Context, jobID [32]byte) error {
	_, err := ps.escrowContract.FinalizeEscrow(ctx, jobID)
	return err
}

// GetJobEscrow returns escrow data for a job
func (ps *PaymentService) GetJobEscrow(ctx context.Context, jobID [32]byte) (*EscrowData, error) {
	return ps.escrowContract.GetEscrow(ctx, jobID)
}

// ===== Slashing Operations =====

// ReportViolation reports a violation against a provider
func (ps *PaymentService) ReportViolation(ctx context.Context, provider common.Address, jobID [32]byte, reason ViolationReason, evidence []byte) ([32]byte, error) {
	disputeID, _, err := ps.slashingContract.ReportViolation(ctx, provider, jobID, reason, evidence)
	return disputeID, err
}

// SubmitDefense submits a defense for a dispute
func (ps *PaymentService) SubmitDefense(ctx context.Context, disputeID [32]byte, defense []byte) error {
	_, err := ps.slashingContract.SubmitDefense(ctx, disputeID, defense)
	return err
}

// GetDispute returns dispute data
func (ps *PaymentService) GetDispute(ctx context.Context, disputeID [32]byte) (*DisputeData, error) {
	return ps.slashingContract.GetDispute(ctx, disputeID)
}

// GetSlashingHistory returns slashing history for a provider
func (ps *PaymentService) GetSlashingHistory(ctx context.Context, provider common.Address) (*SlashingHistory, error) {
	return ps.slashingContract.GetSlashingHistory(ctx, provider)
}

// ===== Pricing Operations =====

// CalculateJobPrice calculates the price for a job
func (ps *PaymentService) CalculateJobPrice(resources pkgtypes.ResourceLimits, duration time.Duration) *big.Int {
	return ps.pricingCalculator.CalculatePrice(resources, duration)
}

// CalculateProviderBid calculates a bid price for a provider
func (ps *PaymentService) CalculateProviderBid(resources pkgtypes.ResourceLimits, duration time.Duration, stake *big.Int) *big.Int {
	return ps.pricingCalculator.CalculateBid(resources, duration, stake)
}

// ===== Event Subscriptions =====

// SubscribeStakeEvents subscribes to staking events
func (ps *PaymentService) SubscribeStakeEvents(ctx context.Context, ch chan<- *StakeEvent) error {
	return ps.stakingContract.SubscribeStakeEvents(ctx, ch)
}

// SubscribeEscrowEvents subscribes to escrow events
func (ps *PaymentService) SubscribeEscrowEvents(ctx context.Context, ch chan<- *EscrowCreatedEvent) error {
	return ps.escrowContract.SubscribeEscrowEvents(ctx, ch)
}

// SubscribeSlashEvents subscribes to slashing events
func (ps *PaymentService) SubscribeSlashEvents(ctx context.Context, ch chan<- *SlashEvent) error {
	return ps.slashingContract.SubscribeSlashEvents(ctx, ch)
}

// ===== Utility Functions =====

// JobIDFromString converts a string to a job ID
func JobIDFromString(s string) [32]byte {
	var id [32]byte
	copy(id[:], []byte(s))
	return id
}

// JobIDToHex converts a job ID to hex string
func JobIDToHex(id [32]byte) string {
	return fmt.Sprintf("%x", id)
}

// NewPaymentServiceFromConfig creates a payment service from daemon config
func NewPaymentServiceFromConfig(economics interface {
	GetRPCURL() string
	GetWSEndpoint() string
	GetChainID() int64
	GetBlockConfirmations() int
	GetTokenAddress() string
	GetEscrowAddress() string
	GetSlashingAddress() string
}, privateKey *ecdsa.PrivateKey, mockMode bool) (*PaymentService, error) {
	config := &PaymentServiceConfig{
		RPCURL:             economics.GetRPCURL(),
		WSEndpoint:         economics.GetWSEndpoint(),
		ChainID:            economics.GetChainID(),
		BlockConfirmations: economics.GetBlockConfirmations(),
		TokenAddress:       common.HexToAddress(economics.GetTokenAddress()),
		EscrowAddress:      common.HexToAddress(economics.GetEscrowAddress()),
		SlashingAddress:    common.HexToAddress(economics.GetSlashingAddress()),
		PrivateKey:         privateKey,
		MockMode:           mockMode,
	}

	return NewPaymentService(config)
}
