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

	// Governance contracts
	delegationContract   *DelegationContract
	reputationContract   *ReputationContract
	verificationContract *VerificationContract
	onChainPricing       *OnChainPricingContract

	// Subdomain registry
	registryContract *RegistryContract

	// Pricing calculator
	pricingCalculator *PricingCalculator

	// Molt credits (in-memory prepaid balances for serverless invocations)
	moltCredits *MoltCreditManager

	// P0 service metering (storage, proxy, crawl, agent)
	serviceMeter *ServiceMeter

	// Service state
	started bool
	mu      sync.RWMutex
}

// PaymentServiceConfig holds configuration for the payment service
type PaymentServiceConfig struct {
	// Base network settings
	RPCURL             string
	WSEndpoint         string
	RPCURLs            []string // Additional RPC endpoints for failover
	WSEndpoints        []string // Additional WS endpoints for failover
	ChainID            int64
	BlockConfirmations int

	// Contract addresses
	TokenAddress              common.Address
	RegistryAddress           common.Address
	EscrowAddress             common.Address
	StakingAddress            common.Address
	SlashingAddress           common.Address
	DelegationAddress         common.Address
	ReputationAddress         common.Address
	VerificationAddress       common.Address
	PricingAddress            common.Address
	SubdomainRegistryAddress  common.Address

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
		RPCURLs:            config.RPCURLs,
		WSEndpoints:        config.WSEndpoints,
		ChainID:            config.ChainID,
		BlockConfirmations: config.BlockConfirmations,
	}

	var err error
	ps.baseClient, err = NewBaseClient(baseConfig, config.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create base client: %w", err)
	}

	// Connect to RPC before creating contracts (they require a connected client)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := ps.baseClient.Connect(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect to RPC: %w", err)
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

	// Create governance contracts
	ps.delegationContract, err = NewDelegationContract(ps.baseClient, config.DelegationAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to create delegation contract: %w", err)
	}

	// Wire delegation into staking for effective stake tier calculation
	ps.stakingContract.SetDelegationContract(ps.delegationContract)

	ps.reputationContract, err = NewReputationContract(ps.baseClient, config.ReputationAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to create reputation contract: %w", err)
	}

	ps.verificationContract, err = NewVerificationContract(ps.baseClient, config.VerificationAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to create verification contract: %w", err)
	}

	ps.onChainPricing, err = NewOnChainPricingContract(ps.baseClient, config.PricingAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to create on-chain pricing contract: %w", err)
	}

	// Create subdomain registry contract (optional — zero address means disabled)
	if config.SubdomainRegistryAddress != (common.Address{}) {
		ps.registryContract, err = NewRegistryContract(ps.baseClient, ps.tokenContract, config.SubdomainRegistryAddress)
		if err != nil {
			return nil, fmt.Errorf("failed to create registry contract: %w", err)
		}
	}

	// Create pricing calculator
	basePricePerHour := config.BasePricePerHour
	if basePricePerHour == nil {
		basePricePerHour = parseWei("1") // 1 BUNKER per hour default
	}
	ps.pricingCalculator = NewPricingCalculator(basePricePerHour)
	ps.moltCredits = NewMoltCreditManager()
	ps.serviceMeter = NewServiceMeter(DefaultServicePricing())

	return ps, nil
}

// setupMockMode sets up the service in mock mode
func (ps *PaymentService) setupMockMode() (*PaymentService, error) {
	ps.tokenContract = NewMockTokenContract()
	ps.stakingContract = NewMockStakingContract()
	ps.escrowContract = NewMockEscrowContract()
	ps.slashingContract = NewMockSlashingContract()
	ps.delegationContract = NewMockDelegationContract()
	ps.stakingContract.SetDelegationContract(ps.delegationContract)
	ps.reputationContract = NewMockReputationContract()
	ps.verificationContract = NewMockVerificationContract()
	ps.onChainPricing = NewMockOnChainPricingContract()
	ps.registryContract = NewMockRegistryContract()

	basePricePerHour := ps.config.BasePricePerHour
	if basePricePerHour == nil {
		basePricePerHour = parseWei("1")
	}
	ps.pricingCalculator = NewPricingCalculator(basePricePerHour)
	ps.moltCredits = NewMoltCreditManager()
	ps.serviceMeter = NewServiceMeter(DefaultServicePricing())

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

// IsActiveProvider checks if a provider is active on-chain
func (ps *PaymentService) IsActiveProvider(ctx context.Context, provider common.Address) (bool, error) {
	return ps.stakingContract.IsActiveProvider(ctx, provider)
}

// GetProviderInfo returns provider information
func (ps *PaymentService) GetProviderInfo(ctx context.Context, provider common.Address) (*ProviderInfoData, error) {
	return ps.stakingContract.GetProviderInfo(ctx, provider)
}

// ===== Identity Operations =====

// StakeWithIdentity stakes tokens with an on-chain NodeID binding
func (ps *PaymentService) StakeWithIdentity(ctx context.Context, amount *big.Int, nodeID [32]byte, region [32]byte, capabilities uint64) error {
	_, err := ps.stakingContract.StakeWithIdentity(ctx, amount, nodeID, region, capabilities)
	return err
}

// UpdateIdentity updates the on-chain NodeID, region, and capabilities
func (ps *PaymentService) UpdateIdentity(ctx context.Context, nodeID [32]byte, region [32]byte, capabilities uint64) error {
	_, err := ps.stakingContract.UpdateIdentity(ctx, nodeID, region, capabilities)
	return err
}

// NodeIDToProvider returns the provider address for a given on-chain NodeID
func (ps *PaymentService) NodeIDToProvider(ctx context.Context, nodeID [32]byte) (common.Address, error) {
	return ps.stakingContract.NodeIDToProvider(ctx, nodeID)
}

// ===== Escrow Operations =====

// CreateJobEscrow creates an escrow for a job
func (ps *PaymentService) CreateJobEscrow(ctx context.Context, jobID [32]byte, provider common.Address, amount *big.Int, duration time.Duration) error {
	durationSecs := big.NewInt(int64(duration.Seconds()))
	_, err := ps.escrowContract.CreateEscrowAndWait(ctx, jobID, provider, amount, durationSecs)
	return err
}

// RegisterExternalReservation stores a user-created on-chain reservation ID
// so the daemon can later call SelectProviders and other operations on it.
func (ps *PaymentService) RegisterExternalReservation(jobID [32]byte, reservationID *big.Int) {
	ps.escrowContract.StoreExternalReservationID(jobID, reservationID)
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

// SelectProviders assigns providers to a job's escrow
func (ps *PaymentService) SelectProviders(ctx context.Context, jobID [32]byte, providers [3]common.Address) error {
	_, err := ps.escrowContract.SelectProviders(ctx, jobID, providers)
	return err
}

// GetJobEscrow returns escrow data for a job
func (ps *PaymentService) GetJobEscrow(ctx context.Context, jobID [32]byte) (*EscrowData, error) {
	return ps.escrowContract.GetEscrow(ctx, jobID)
}

// CleanupStaleReservations removes reservation ID mappings older than maxAge.
func (ps *PaymentService) CleanupStaleReservations(maxAge time.Duration) int {
	return ps.escrowContract.CleanupStaleReservations(maxAge)
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

// ===== Molt Pricing & Credits =====

// CalculateMoltCost calculates the cost of a single Molt invocation based on
// actual execution duration and memory used. Uses the minimum billing floor
// from PricingConfig (default 100ms).
func (ps *PaymentService) CalculateMoltCost(duration time.Duration, memoryUsedBytes int64) *big.Int {
	pricingCfg := pkgtypes.DefaultPricingConfig()
	return ps.pricingCalculator.CalculateMoltInvocationPrice(duration, memoryUsedBytes, pricingCfg)
}

// DepositMoltCredits adds prepaid credits for a requester's Molt invocations.
func (ps *PaymentService) DepositMoltCredits(requesterAddress string, amount *big.Int) {
	ps.moltCredits.Deposit(requesterAddress, amount)
}

// DeductMoltCredit subtracts an invocation cost from the requester's credit balance.
func (ps *PaymentService) DeductMoltCredit(requesterAddress string, cost *big.Int) error {
	return ps.moltCredits.Deduct(requesterAddress, cost)
}

// GetMoltCreditBalance returns the current prepaid credit balance.
func (ps *PaymentService) GetMoltCreditBalance(requesterAddress string) *big.Int {
	return ps.moltCredits.GetBalance(requesterAddress)
}

// RefundMoltCredits refunds all remaining credits and returns the amount.
func (ps *PaymentService) RefundMoltCredits(requesterAddress string) *big.Int {
	return ps.moltCredits.RefundAll(requesterAddress)
}

// ===== P0 Service Metering =====

// RecordStorageUpload records a storage upload for billing.
func (ps *PaymentService) RecordStorageUpload(wallet string, sizeBytes int64) {
	ps.serviceMeter.RecordStorageUpload(wallet, sizeBytes)
}

// RecordStorageDelete records a storage deletion for billing.
func (ps *PaymentService) RecordStorageDelete(wallet string, sizeBytes int64) {
	ps.serviceMeter.RecordStorageDelete(wallet, sizeBytes)
}

// RecordProxySession records proxy bandwidth usage for billing.
func (ps *PaymentService) RecordProxySession(wallet string, bytesIn, bytesOut int64) {
	ps.serviceMeter.RecordProxySession(wallet, bytesIn, bytesOut)
}

// RecordCrawlJob records a crawl job for billing.
func (ps *PaymentService) RecordCrawlJob(wallet string, pagesCrawled, resultBytes int64) {
	ps.serviceMeter.RecordCrawlJob(wallet, pagesCrawled, resultBytes)
}

// RecordAgentInvocation records an agent invocation for billing.
func (ps *PaymentService) RecordAgentInvocation(wallet string, tokensUsed int64) {
	ps.serviceMeter.RecordAgentInvocation(wallet, tokensUsed)
}

// CalculateProxyCost calculates the cost for proxy bandwidth usage.
func (ps *PaymentService) CalculateProxyCost(bytesIn, bytesOut int64) *big.Int {
	return ps.serviceMeter.CalculateProxyCost(bytesIn, bytesOut)
}

// CalculateCrawlCost calculates the cost for a crawl job.
func (ps *PaymentService) CalculateCrawlCost(pagesCrawled, resultBytes int64) *big.Int {
	return ps.serviceMeter.CalculateCrawlCost(pagesCrawled, resultBytes)
}

// CalculateAgentCost calculates the cost for an agent invocation.
func (ps *PaymentService) CalculateAgentCost(tokensUsed int64) *big.Int {
	return ps.serviceMeter.CalculateAgentCost(tokensUsed)
}

// CalculateStorageBill calculates the monthly storage bill for a wallet.
func (ps *PaymentService) CalculateStorageBill(wallet string) *big.Int {
	return ps.serviceMeter.CalculateStorageBill(wallet)
}

// GetServiceUsage returns usage data for a wallet across all P0 services.
func (ps *PaymentService) GetServiceUsage(wallet string) *WalletServiceUsage {
	return ps.serviceMeter.GetUsage(wallet)
}

// RecordServiceCharge records that a charge was applied to a wallet.
func (ps *PaymentService) RecordServiceCharge(wallet string, amount *big.Int) {
	ps.serviceMeter.RecordCharge(wallet, amount)
}

// ActiveServiceWallets returns the number of wallets with recorded P0 service usage.
func (ps *PaymentService) ActiveServiceWallets() int {
	return ps.serviceMeter.ActiveWallets()
}

// ===== Delegation Operations =====

// Delegate delegates tokens to a provider
func (ps *PaymentService) Delegate(ctx context.Context, provider common.Address, amount *big.Int) error {
	_, err := ps.delegationContract.Delegate(ctx, provider, amount)
	return err
}

// RequestUndelegation requests to undelegate tokens
func (ps *PaymentService) RequestUndelegation(ctx context.Context, amount *big.Int) error {
	_, err := ps.delegationContract.RequestUndelegate(ctx, amount)
	return err
}

// CompleteUndelegation completes a pending undelegation
func (ps *PaymentService) CompleteUndelegation(ctx context.Context, index uint64) error {
	_, err := ps.delegationContract.CompleteUndelegate(ctx, new(big.Int).SetUint64(index))
	return err
}

// GetDelegation returns delegation info for a delegator
func (ps *PaymentService) GetDelegation(ctx context.Context, delegator common.Address) (*DelegationData, error) {
	return ps.delegationContract.GetDelegation(ctx, delegator)
}

// GetDelegationProviderConfig returns a provider's delegation configuration
func (ps *PaymentService) GetDelegationProviderConfig(ctx context.Context, provider common.Address) (*ProviderDelegationConfigData, error) {
	return ps.delegationContract.GetProviderConfig(ctx, provider)
}

// GetTotalDelegatedTo returns total tokens delegated to a provider
func (ps *PaymentService) GetTotalDelegatedTo(ctx context.Context, provider common.Address) (*big.Int, error) {
	return ps.delegationContract.GetTotalDelegatedTo(ctx, provider)
}

// SetDelegationConfig sets a provider's delegation parameters
func (ps *PaymentService) SetDelegationConfig(ctx context.Context, rewardCutBps, feeShareBps uint16) error {
	_, err := ps.delegationContract.SetDelegationConfig(ctx, rewardCutBps, feeShareBps)
	return err
}

// ToggleAcceptDelegations enables or disables delegation acceptance
func (ps *PaymentService) ToggleAcceptDelegations(ctx context.Context, accept bool) error {
	_, err := ps.delegationContract.ToggleAcceptDelegations(ctx, accept)
	return err
}

// ===== Reputation Operations =====

// GetReputationScore returns the reputation score for a provider
func (ps *PaymentService) GetReputationScore(ctx context.Context, provider common.Address) (*big.Int, error) {
	return ps.reputationContract.GetScore(ctx, provider)
}

// GetReputationTier returns the reputation tier for a provider
func (ps *PaymentService) GetReputationTier(ctx context.Context, provider common.Address) (ReputationTier, error) {
	return ps.reputationContract.GetTier(ctx, provider)
}

// IsEligibleForJobs checks if a provider is eligible for job assignment
func (ps *PaymentService) IsEligibleForJobs(ctx context.Context, provider common.Address) (bool, error) {
	return ps.reputationContract.IsEligibleForJobs(ctx, provider)
}

// GetReputation returns full reputation data for a provider
func (ps *PaymentService) GetReputation(ctx context.Context, provider common.Address) (*ReputationDataOnChain, error) {
	return ps.reputationContract.GetReputation(ctx, provider)
}

// RegisterProviderReputation registers a provider in the reputation system
func (ps *PaymentService) RegisterProviderReputation(ctx context.Context, provider common.Address) error {
	_, err := ps.reputationContract.RegisterProvider(ctx, provider)
	return err
}

// RecordJobCompleted records a successful job completion
func (ps *PaymentService) RecordJobCompleted(ctx context.Context, provider common.Address) error {
	_, err := ps.reputationContract.RecordJobCompleted(ctx, provider)
	return err
}

// RecordJobFailed records a failed job
func (ps *PaymentService) RecordJobFailed(ctx context.Context, provider common.Address) error {
	_, err := ps.reputationContract.RecordJobFailed(ctx, provider)
	return err
}

// RecordSlashEvent records a slash event against a provider
func (ps *PaymentService) RecordSlashEvent(ctx context.Context, provider common.Address) error {
	_, err := ps.reputationContract.RecordSlashEvent(ctx, provider)
	return err
}

// ===== Verification Operations =====

// SubmitAttestation submits a hardware/software attestation
func (ps *PaymentService) SubmitAttestation(ctx context.Context, hash [32]byte) error {
	_, err := ps.verificationContract.SubmitAttestation(ctx, hash)
	return err
}

// CheckMissedAttestations checks for missed attestations
func (ps *PaymentService) CheckMissedAttestations(ctx context.Context, provider common.Address) error {
	_, err := ps.verificationContract.CheckMissedAttestations(ctx, provider)
	return err
}

// ChallengeAttestation challenges a provider's attestation
func (ps *PaymentService) ChallengeAttestation(ctx context.Context, provider common.Address, proof []byte) error {
	_, err := ps.verificationContract.ChallengeAttestation(ctx, provider, proof)
	return err
}

// GetAttestation returns attestation data for a provider
func (ps *PaymentService) GetAttestation(ctx context.Context, provider common.Address) (*AttestationData, error) {
	return ps.verificationContract.GetAttestation(ctx, provider)
}

// IsAttestationCurrent checks if a provider's attestation is up to date
func (ps *PaymentService) IsAttestationCurrent(ctx context.Context, provider common.Address) (bool, error) {
	return ps.verificationContract.IsAttestationCurrent(ctx, provider)
}

// ===== On-Chain Pricing Operations =====

// GetOnChainCost calculates cost using on-chain oracle pricing
func (ps *PaymentService) GetOnChainCost(ctx context.Context, req ResourceRequestData) (*big.Int, error) {
	return ps.onChainPricing.CalculateCost(ctx, req)
}

// GetOnChainProviderCost calculates cost using a provider's on-chain pricing
func (ps *PaymentService) GetOnChainProviderCost(ctx context.Context, provider common.Address, req ResourceRequestData) (*big.Int, error) {
	return ps.onChainPricing.CalculateProviderCost(ctx, provider, req)
}

// GetTokenPrice returns the current BUNKER token price from the oracle
func (ps *PaymentService) GetTokenPrice(ctx context.Context) (*big.Int, error) {
	return ps.onChainPricing.GetTokenPrice(ctx)
}

// GetResourcePrices returns current base resource prices from the oracle
func (ps *PaymentService) GetResourcePrices(ctx context.Context) (*ResourcePricesData, error) {
	return ps.onChainPricing.GetPrices(ctx)
}

// GetPricingMultipliers returns current pricing multipliers from the oracle
func (ps *PaymentService) GetPricingMultipliers(ctx context.Context) (*MultipliersData, error) {
	return ps.onChainPricing.GetMultipliers(ctx)
}

// GetProviderPrices returns a provider's custom pricing multipliers
func (ps *PaymentService) GetProviderPrices(ctx context.Context, provider common.Address) (*ProviderPricingData, error) {
	return ps.onChainPricing.GetProviderPrices(ctx, provider)
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

// ===== Event Watcher =====

// NewEventWatcherFromService creates an EventWatcher with access to this service's
// contracts and base client. Returns nil if running in mock mode.
func (ps *PaymentService) NewEventWatcherFromService() *EventWatcher {
	if ps.config.MockMode || ps.baseClient == nil {
		return nil
	}
	return NewEventWatcher(ps.baseClient, ps.stakingContract, ps.escrowContract, ps.slashingContract)
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

// ===== Subdomain Registry Operations =====

// RegisterSubdomain registers a vanity subdomain name on-chain.
func (ps *PaymentService) RegisterSubdomain(ctx context.Context, name string, deploymentID [32]byte) error {
	if ps.registryContract == nil {
		return fmt.Errorf("subdomain registry not configured")
	}
	_, err := ps.registryContract.Register(ctx, name, deploymentID)
	return err
}

// ReleaseSubdomain releases a subdomain name.
func (ps *PaymentService) ReleaseSubdomain(ctx context.Context, name string) error {
	if ps.registryContract == nil {
		return fmt.Errorf("subdomain registry not configured")
	}
	_, err := ps.registryContract.Release(ctx, name)
	return err
}

// TransferSubdomain transfers a subdomain to a new owner.
func (ps *PaymentService) TransferSubdomain(ctx context.Context, name string, newOwner common.Address) error {
	if ps.registryContract == nil {
		return fmt.Errorf("subdomain registry not configured")
	}
	_, err := ps.registryContract.Transfer(ctx, name, newOwner)
	return err
}

// UpdateSubdomainDeployment updates the deployment ID a subdomain points to.
func (ps *PaymentService) UpdateSubdomainDeployment(ctx context.Context, name string, newDeploymentID [32]byte) error {
	if ps.registryContract == nil {
		return fmt.Errorf("subdomain registry not configured")
	}
	_, err := ps.registryContract.UpdateDeployment(ctx, name, newDeploymentID)
	return err
}

// ResolveSubdomain resolves a subdomain name to its registration data.
func (ps *PaymentService) ResolveSubdomain(ctx context.Context, name string) (*SubdomainRegistration, error) {
	if ps.registryContract == nil {
		return nil, fmt.Errorf("subdomain registry not configured")
	}
	return ps.registryContract.Resolve(ctx, name)
}

// IsSubdomainAvailable checks if a subdomain name is available.
func (ps *PaymentService) IsSubdomainAvailable(ctx context.Context, name string) (bool, error) {
	if ps.registryContract == nil {
		return false, fmt.Errorf("subdomain registry not configured")
	}
	return ps.registryContract.IsAvailable(ctx, name)
}

// GetSubdomainRegistrationFee returns the current registration fee.
func (ps *PaymentService) GetSubdomainRegistrationFee(ctx context.Context) (*big.Int, error) {
	if ps.registryContract == nil {
		return nil, fmt.Errorf("subdomain registry not configured")
	}
	return ps.registryContract.GetRegistrationFee(ctx)
}

// ListOwnedSubdomains returns all subdomains owned by an address.
func (ps *PaymentService) ListOwnedSubdomains(ctx context.Context, owner common.Address) ([]SubdomainRegistration, error) {
	if ps.registryContract == nil {
		return nil, fmt.Errorf("subdomain registry not configured")
	}
	return ps.registryContract.ListOwnedNames(ctx, owner)
}

// RenewSubdomain extends a subdomain's expiration by 365 days.
func (ps *PaymentService) RenewSubdomain(ctx context.Context, name string) error {
	if ps.registryContract == nil {
		return fmt.Errorf("subdomain registry not configured")
	}
	_, err := ps.registryContract.Renew(ctx, name)
	return err
}

// ReserveSubdomain reserves a subdomain name for 48 hours.
func (ps *PaymentService) ReserveSubdomain(ctx context.Context, name string) error {
	if ps.registryContract == nil {
		return fmt.Errorf("subdomain registry not configured")
	}
	_, err := ps.registryContract.Reserve(ctx, name)
	return err
}

// ClaimSubdomainReservation finalizes a reserved subdomain with a deployment ID.
func (ps *PaymentService) ClaimSubdomainReservation(ctx context.Context, name string, deploymentID [32]byte) error {
	if ps.registryContract == nil {
		return fmt.Errorf("subdomain registry not configured")
	}
	_, err := ps.registryContract.ClaimReservation(ctx, name, deploymentID)
	return err
}

// CancelSubdomainReservation cancels a pending subdomain reservation.
func (ps *PaymentService) CancelSubdomainReservation(ctx context.Context, name string) error {
	if ps.registryContract == nil {
		return fmt.Errorf("subdomain registry not configured")
	}
	_, err := ps.registryContract.CancelReservation(ctx, name)
	return err
}

// SetSubdomainMetadata sets description and avatar URL for a subdomain.
func (ps *PaymentService) SetSubdomainMetadata(ctx context.Context, name string, description, avatarURL string) error {
	if ps.registryContract == nil {
		return fmt.Errorf("subdomain registry not configured")
	}
	_, err := ps.registryContract.SetMetadata(ctx, name, description, avatarURL)
	return err
}

// SetSubdomainPrimaryName sets a subdomain as the primary name for reverse resolution.
func (ps *PaymentService) SetSubdomainPrimaryName(ctx context.Context, name string) error {
	if ps.registryContract == nil {
		return fmt.Errorf("subdomain registry not configured")
	}
	_, err := ps.registryContract.SetPrimaryName(ctx, name)
	return err
}

// ReclaimSubdomain reclaims a squatted subdomain name.
func (ps *PaymentService) ReclaimSubdomain(ctx context.Context, name string) error {
	if ps.registryContract == nil {
		return fmt.Errorf("subdomain registry not configured")
	}
	_, err := ps.registryContract.ReclaimSquatted(ctx, name)
	return err
}

// IsSubdomainExpired checks if a subdomain name has expired.
func (ps *PaymentService) IsSubdomainExpired(ctx context.Context, name string) (bool, error) {
	if ps.registryContract == nil {
		return false, fmt.Errorf("subdomain registry not configured")
	}
	return ps.registryContract.IsExpired(ctx, name)
}

// ReverseResolveSubdomain looks up the primary name for a deployment ID.
func (ps *PaymentService) ReverseResolveSubdomain(ctx context.Context, deploymentID [32]byte) (string, error) {
	if ps.registryContract == nil {
		return "", fmt.Errorf("subdomain registry not configured")
	}
	return ps.registryContract.ReverseResolve(ctx, deploymentID)
}

// GetSubdomainMetadata returns the metadata for a subdomain.
func (ps *PaymentService) GetSubdomainMetadata(ctx context.Context, name string) (*SubdomainMetadata, error) {
	if ps.registryContract == nil {
		return nil, fmt.Errorf("subdomain registry not configured")
	}
	return ps.registryContract.GetMetadata(ctx, name)
}

// NewPaymentServiceFromConfig creates a payment service from daemon config
func NewPaymentServiceFromConfig(economics interface {
	GetRPCURL() string
	GetWSEndpoint() string
	GetChainID() int64
	GetBlockConfirmations() int
	GetTokenAddress() string
	GetStakingAddress() string
	GetEscrowAddress() string
	GetSlashingAddress() string
}, privateKey *ecdsa.PrivateKey, mockMode bool) (*PaymentService, error) {
	config := &PaymentServiceConfig{
		RPCURL:             economics.GetRPCURL(),
		WSEndpoint:         economics.GetWSEndpoint(),
		ChainID:            economics.GetChainID(),
		BlockConfirmations: economics.GetBlockConfirmations(),
		TokenAddress:       common.HexToAddress(economics.GetTokenAddress()),
		StakingAddress:     common.HexToAddress(economics.GetStakingAddress()),
		EscrowAddress:      common.HexToAddress(economics.GetEscrowAddress()),
		SlashingAddress:    common.HexToAddress(economics.GetSlashingAddress()),
		PrivateKey:         privateKey,
		MockMode:           mockMode,
	}

	return NewPaymentService(config)
}
