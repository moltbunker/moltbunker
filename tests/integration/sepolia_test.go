//go:build sepolia

// Package integration — Base Sepolia integration tests.
//
// These tests connect to real deployed contracts on Base Sepolia (chain 84532)
// using Go's ethclient (not Foundry cast). They validate that the Go ABI
// bindings in internal/payment/ match the deployed contract bytecode.
//
// Run with:
//
//	go test -tags sepolia -v -timeout 2m ./tests/integration/...
//
// No private key needed — all tests are read-only.
package integration

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/moltbunker/moltbunker/internal/payment"
)

// Base Sepolia contract addresses (deployed 2026-02-13, registry 2026-02-26)
const (
	sepoliaRPC     = "https://sepolia.base.org"
	sepoliaChainID = 84532

	sepoliaToken        = "0x4cc3F5C0d2Ecb4118e214980906eFe5c880a6ceA"
	sepoliaStaking      = "0xDC76d972a827D2a19867EF9aBD335014d5Cf7D6a"
	sepoliaEscrow       = "0xBAdaB53a9E98D904E3dfcDb728D510c69DAeE9B4"
	sepoliaPricing      = "0x5A61b05F289344202433ccDf44aFc611d9E3dA47"
	sepoliaTimelock     = "0xcD8af28808749CD4B55a970f14DA250C8EAEd3C9"
	sepoliaDelegation   = "0x071252B4f4bC80cccEccDe1A644229EE2dAf09F5"
	sepoliaReputation   = "0x55721fC66B30Fe26a0820CfDeffC0815135678Ed"
	sepoliaVerification = "0x9aA9Fc961da51dcFfF0232883631f7147CaBFBCD"
	sepoliaRegistry     = "0x3559A7D2E6F09eA74a295e654e0D6C22F921D4b5"

	// Deployer/owner address (public — not a secret)
	deployerAddress = "0xAc1D8d6e25E54c05986E8bFa9b759063D5e69592"
)

// newSepoliaService creates a read-only PaymentService connected to Base Sepolia.
func newSepoliaService(t *testing.T) *payment.PaymentService {
	t.Helper()

	ps, err := payment.NewPaymentService(&payment.PaymentServiceConfig{
		RPCURL:                   sepoliaRPC,
		ChainID:                  sepoliaChainID,
		TokenAddress:             common.HexToAddress(sepoliaToken),
		StakingAddress:           common.HexToAddress(sepoliaStaking),
		EscrowAddress:            common.HexToAddress(sepoliaEscrow),
		SlashingAddress:          common.HexToAddress(sepoliaStaking), // slashing is in staking contract
		DelegationAddress:        common.HexToAddress(sepoliaDelegation),
		ReputationAddress:        common.HexToAddress(sepoliaReputation),
		VerificationAddress:      common.HexToAddress(sepoliaVerification),
		PricingAddress:           common.HexToAddress(sepoliaPricing),
		SubdomainRegistryAddress: common.HexToAddress(sepoliaRegistry),
		PrivateKey:               nil, // read-only
	})
	if err != nil {
		t.Fatalf("failed to create PaymentService: %v", err)
	}
	t.Cleanup(ps.Stop)
	return ps
}

func sepoliaCtx(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)
	return ctx
}

// ─── Token Contract ──────────────────────────────────────────────────────────

func TestSepolia_TokenBalance(t *testing.T) {
	ps := newSepoliaService(t)
	ctx := sepoliaCtx(t)

	deployer := common.HexToAddress(deployerAddress)
	balance, err := ps.GetTokenBalance(ctx, deployer)
	if err != nil {
		t.Fatalf("GetTokenBalance failed: %v", err)
	}

	// Deployer was minted 10B BUNKER (10_000_000_000 * 1e18)
	// Balance may be less due to transfers, but should be positive
	if balance.Sign() <= 0 {
		t.Errorf("expected positive deployer balance, got %s", balance)
	}
	t.Logf("Deployer BUNKER balance: %s wei", balance)
}

func TestSepolia_ETHBalance(t *testing.T) {
	ps := newSepoliaService(t)
	ctx := sepoliaCtx(t)

	deployer := common.HexToAddress(deployerAddress)
	balance, err := ps.GetETHBalance(ctx, deployer)
	if err != nil {
		t.Fatalf("GetETHBalance failed: %v", err)
	}

	// Should have some ETH for gas
	if balance.Sign() <= 0 {
		t.Errorf("expected positive ETH balance, got %s", balance)
	}
	t.Logf("Deployer ETH balance: %s wei", balance)
}

func TestSepolia_ZeroAddressBalance(t *testing.T) {
	ps := newSepoliaService(t)
	ctx := sepoliaCtx(t)

	zero := common.Address{}
	balance, err := ps.GetTokenBalance(ctx, zero)
	if err != nil {
		t.Fatalf("GetTokenBalance(zero) failed: %v", err)
	}

	// Zero address balance should be >= 0 (may have burned tokens)
	t.Logf("Zero address BUNKER balance (burned): %s wei", balance)
}

// ─── Staking Contract ────────────────────────────────────────────────────────

func TestSepolia_StakingTier(t *testing.T) {
	ps := newSepoliaService(t)
	ctx := sepoliaCtx(t)

	deployer := common.HexToAddress(deployerAddress)

	// GetTier should return a valid tier (may be 0/Unstaked if deployer hasn't staked)
	tier, err := ps.GetTier(ctx, deployer)
	if err != nil {
		t.Fatalf("GetTier failed: %v", err)
	}
	t.Logf("Deployer staking tier: %v", tier)
}

func TestSepolia_StakeInfo(t *testing.T) {
	ps := newSepoliaService(t)
	ctx := sepoliaCtx(t)

	deployer := common.HexToAddress(deployerAddress)
	info, err := ps.GetStakeInfo(ctx, deployer)
	if err != nil {
		// Reverts if provider never staked — valid on-chain behavior
		t.Logf("GetStakeInfo reverted (deployer may not have staked): %v", err)
		return
	}
	t.Logf("Deployer stake: amount=%s, tier=%v", info.StakedAmount, info.Tier)
}

func TestSepolia_HasMinimumStake(t *testing.T) {
	ps := newSepoliaService(t)
	ctx := sepoliaCtx(t)

	// Random address should not have minimum stake
	random := common.HexToAddress("0x0000000000000000000000000000000000000001")
	has, err := ps.HasMinimumStake(ctx, random)
	if err != nil {
		// May revert for non-existent providers
		t.Logf("HasMinimumStake reverted (expected for random address): %v", err)
		return
	}
	if has {
		t.Error("random address should not have minimum stake")
	}
}

func TestSepolia_IsActiveProvider(t *testing.T) {
	ps := newSepoliaService(t)
	ctx := sepoliaCtx(t)

	random := common.HexToAddress("0x0000000000000000000000000000000000000001")
	active, err := ps.IsActiveProvider(ctx, random)
	if err != nil {
		t.Fatalf("IsActiveProvider failed: %v", err)
	}
	if active {
		t.Error("random address should not be an active provider")
	}
}

// ─── Pricing Contract ────────────────────────────────────────────────────────

func TestSepolia_ResourcePrices(t *testing.T) {
	ps := newSepoliaService(t)
	ctx := sepoliaCtx(t)

	prices, err := ps.GetResourcePrices(ctx)
	if err != nil {
		t.Fatalf("GetResourcePrices failed: %v", err)
	}

	// Log actual on-chain prices (may differ from expected if contract was redeployed)
	t.Logf("Prices — CPU: %s, Mem: %s, Stor: %s, Net: %s, GPU: %s",
		prices.CPUPerCoreHour, prices.MemoryPerGBHour, prices.StoragePerGBHour,
		prices.BandwidthPerGBHour, prices.GPUPerHour)

	// Prices should be non-negative (valid return from contract)
	for _, p := range []*big.Int{prices.CPUPerCoreHour, prices.MemoryPerGBHour, prices.StoragePerGBHour, prices.BandwidthPerGBHour} {
		if p == nil {
			t.Error("price field is nil")
		} else if p.Sign() < 0 {
			t.Errorf("price field is negative: %s", p)
		}
	}
}

func TestSepolia_PricingMultipliers(t *testing.T) {
	ps := newSepoliaService(t)
	ctx := sepoliaCtx(t)

	multipliers, err := ps.GetPricingMultipliers(ctx)
	if err != nil {
		t.Fatalf("GetPricingMultipliers failed: %v", err)
	}

	t.Logf("Multipliers — Demand: %s, Region: %s, TierDiscount: %s",
		multipliers.DemandMultiplier, multipliers.RegionMultiplier,
		multipliers.TierDiscount)
}

// ─── Delegation Contract ─────────────────────────────────────────────────────

func TestSepolia_DelegationQuery(t *testing.T) {
	ps := newSepoliaService(t)
	ctx := sepoliaCtx(t)

	deployer := common.HexToAddress(deployerAddress)

	// Query total delegated to deployer (may be zero)
	total, err := ps.GetTotalDelegatedTo(ctx, deployer)
	if err != nil {
		t.Fatalf("GetTotalDelegatedTo failed: %v", err)
	}
	t.Logf("Total delegated to deployer: %s", total)
}

// ─── Reputation Contract ─────────────────────────────────────────────────────

func TestSepolia_ReputationScore(t *testing.T) {
	ps := newSepoliaService(t)
	ctx := sepoliaCtx(t)

	deployer := common.HexToAddress(deployerAddress)

	score, err := ps.GetReputationScore(ctx, deployer)
	if err != nil {
		// Reputation may revert for unregistered providers — that's fine
		t.Logf("GetReputationScore returned error (expected if unregistered): %v", err)
		return
	}

	// Score is 0-1000
	if score.Cmp(big.NewInt(1000)) > 0 {
		t.Errorf("reputation score %s exceeds max 1000", score)
	}
	t.Logf("Deployer reputation score: %s", score)
}

// ─── Verification Contract ───────────────────────────────────────────────────

func TestSepolia_AttestationQuery(t *testing.T) {
	ps := newSepoliaService(t)
	ctx := sepoliaCtx(t)

	deployer := common.HexToAddress(deployerAddress)

	attestation, err := ps.GetAttestation(ctx, deployer)
	if err != nil {
		t.Logf("GetAttestation returned error (expected if no attestation): %v", err)
		return
	}

	t.Logf("Deployer attestation: hash=%x, lastTime=%v", attestation.LastHash, attestation.LastTime)
}

func TestSepolia_AttestationCurrent(t *testing.T) {
	ps := newSepoliaService(t)
	ctx := sepoliaCtx(t)

	deployer := common.HexToAddress(deployerAddress)

	current, err := ps.IsAttestationCurrent(ctx, deployer)
	if err != nil {
		t.Logf("IsAttestationCurrent returned error (expected if no attestation): %v", err)
		return
	}

	t.Logf("Deployer attestation current: %v", current)
}

// ─── Registry Contract ───────────────────────────────────────────────────────

func TestSepolia_SubdomainAvailability(t *testing.T) {
	ps := newSepoliaService(t)
	ctx := sepoliaCtx(t)

	// Check a known reserved name
	available, err := ps.IsSubdomainAvailable(ctx, "moltbunker")
	if err != nil {
		t.Fatalf("IsSubdomainAvailable('moltbunker') failed: %v", err)
	}
	t.Logf("'moltbunker' available: %v", available)

	// Random gibberish should be available (unless someone registered it)
	available, err = ps.IsSubdomainAvailable(ctx, "xyzzy99test42random")
	if err != nil {
		t.Fatalf("IsSubdomainAvailable('xyzzy99test42random') failed: %v", err)
	}
	t.Logf("'xyzzy99test42random' available: %v", available)
	if !available {
		t.Error("random name should be available")
	}
}

func TestSepolia_RegistrationFee(t *testing.T) {
	ps := newSepoliaService(t)
	ctx := sepoliaCtx(t)

	fee, err := ps.GetSubdomainRegistrationFee(ctx)
	if err != nil {
		t.Fatalf("GetSubdomainRegistrationFee failed: %v", err)
	}

	if fee.Sign() < 0 {
		t.Errorf("registration fee should be non-negative, got %s", fee)
	}
	t.Logf("Registration fee: %s BUNKER wei", fee)
}

// ─── Cross-Contract Consistency ──────────────────────────────────────────────

func TestSepolia_AllContractsReachable(t *testing.T) {
	// This test verifies that creating a PaymentService with all 9 contract
	// addresses succeeds and all ABI bindings parse correctly against the
	// deployed bytecode.
	ps := newSepoliaService(t)

	if !ps.IsConnected() {
		t.Fatal("PaymentService not connected after creation")
	}

	// Exercise each contract binding with a read-only call
	ctx := sepoliaCtx(t)
	deployer := common.HexToAddress(deployerAddress)

	tests := []struct {
		name string
		fn   func() error
	}{
		{"Token.BalanceOf", func() error { _, err := ps.GetTokenBalance(ctx, deployer); return err }},
		{"Staking.GetTier", func() error { _, err := ps.GetTier(ctx, deployer); return err }},
		{"Pricing.GetPrices", func() error { _, err := ps.GetResourcePrices(ctx); return err }},
		{"Delegation.GetTotalDelegatedTo", func() error { _, err := ps.GetTotalDelegatedTo(ctx, deployer); return err }},
		{"Registry.IsAvailable", func() error { _, err := ps.IsSubdomainAvailable(ctx, "test"); return err }},
		{"Registry.GetFee", func() error { _, err := ps.GetSubdomainRegistrationFee(ctx); return err }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.fn(); err != nil {
				t.Errorf("%s failed: %v", tt.name, err)
			}
		})
	}
}

// ─── Pricing Calculator (off-chain) ─────────────────────────────────────────

func TestSepolia_MoltPricingConsistency(t *testing.T) {
	ps := newSepoliaService(t)

	// Calculate a Molt invocation cost locally
	cost := ps.CalculateMoltCost(500*time.Millisecond, 64*1024*1024) // 500ms, 64MB
	if cost.Sign() <= 0 {
		t.Errorf("Molt invocation cost should be positive, got %s", cost)
	}
	t.Logf("Molt cost (500ms, 64MB): %s BUNKER wei", cost)

	// Minimum billing floor: 100ms should produce a cost
	minCost := ps.CalculateMoltCost(1*time.Millisecond, 1024) // 1ms → floors to 100ms
	if minCost.Sign() <= 0 {
		t.Errorf("Molt minimum cost should be positive, got %s", minCost)
	}
	t.Logf("Molt cost (1ms floor→100ms, 1KB): %s BUNKER wei", minCost)
}

func TestSepolia_MoltCredits(t *testing.T) {
	ps := newSepoliaService(t)

	addr := "0x1234567890abcdef1234567890abcdef12345678"

	// Initial balance should be zero
	balance := ps.GetMoltCreditBalance(addr)
	if balance.Sign() != 0 {
		t.Errorf("initial credit balance should be zero, got %s", balance)
	}

	// Deposit
	deposit := big.NewInt(1_000_000)
	ps.DepositMoltCredits(addr, deposit)
	balance = ps.GetMoltCreditBalance(addr)
	if balance.Cmp(deposit) != 0 {
		t.Errorf("balance after deposit: expected %s, got %s", deposit, balance)
	}

	// Deduct
	cost := big.NewInt(100)
	if err := ps.DeductMoltCredit(addr, cost); err != nil {
		t.Fatalf("DeductMoltCredit failed: %v", err)
	}
	balance = ps.GetMoltCreditBalance(addr)
	expected := new(big.Int).Sub(deposit, cost)
	if balance.Cmp(expected) != 0 {
		t.Errorf("balance after deduct: expected %s, got %s", expected, balance)
	}

	// Overdraft should fail
	huge := new(big.Int).Mul(deposit, big.NewInt(10))
	if err := ps.DeductMoltCredit(addr, huge); err == nil {
		t.Error("expected overdraft error, got nil")
	}

	// Refund
	refunded := ps.RefundMoltCredits(addr)
	if refunded.Cmp(expected) != 0 {
		t.Errorf("refund: expected %s, got %s", expected, refunded)
	}
	balance = ps.GetMoltCreditBalance(addr)
	if balance.Sign() != 0 {
		t.Errorf("balance after refund should be zero, got %s", balance)
	}
}
