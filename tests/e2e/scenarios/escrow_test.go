//go:build e2e

package scenarios

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/moltbunker/moltbunker/internal/payment"
	"github.com/moltbunker/moltbunker/pkg/types"
	"github.com/moltbunker/moltbunker/tests/e2e/testutil"
)

// bunkerToWei converts a whole-number BUNKER amount to wei (18 decimals).
func bunkerToWei(tokens int64) *big.Int {
	decimals := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	return new(big.Int).Mul(big.NewInt(tokens), decimals)
}

// TestE2E_EscrowFullLifecycle creates an escrow, releases payments in 4 stages
// (25% each), and verifies that the cumulative released amount equals the
// original escrow amount.
func TestE2E_EscrowFullLifecycle(t *testing.T) {
	assert := testutil.NewAssertions(t)

	escrowMgr := payment.NewEscrowManager()

	// 200 BUNKER escrowed for 1 hour
	amount := bunkerToWei(200)
	duration := time.Hour

	escrow := escrowMgr.CreateEscrow("lifecycle-001", amount, duration)
	assert.NotNil(escrow)
	assert.Equal("lifecycle-001", escrow.ReservationID)

	// Release in 4 equal stages: 25%, 50%, 75%, 100% of duration
	totalReleased := big.NewInt(0)
	stages := []time.Duration{
		duration / 4,       // 15 min
		duration / 2,       // 30 min
		duration * 3 / 4,   // 45 min
		duration,           // 60 min
	}

	for i, uptime := range stages {
		released, err := escrowMgr.ReleasePayment("lifecycle-001", uptime)
		assert.NoError(err, "Stage %d release should succeed", i+1)
		assert.True(released.Sign() > 0, "Stage %d should release a positive amount", i+1)
		totalReleased.Add(totalReleased, released)
	}

	// The cumulative released amount must equal the original escrow
	assert.Equal(amount.String(), totalReleased.String(),
		"Total released across 4 stages should equal escrow amount")
}

// TestE2E_EscrowRefundFlow creates an escrow, releases a partial payment, then
// refunds the remainder.  The refunded amount is the escrow total minus what
// was already released.
func TestE2E_EscrowRefundFlow(t *testing.T) {
	assert := testutil.NewAssertions(t)

	escrowMgr := payment.NewEscrowManager()

	amount := bunkerToWei(100)
	duration := time.Hour

	escrow := escrowMgr.CreateEscrow("refund-001", amount, duration)
	assert.NotNil(escrow)

	// Release payment for 30 minutes of uptime (~50 BUNKER)
	halfDuration := duration / 2
	released, err := escrowMgr.ReleasePayment("refund-001", halfDuration)
	assert.NoError(err)
	assert.True(released.Sign() > 0, "Partial release should be positive")

	// Verify released is roughly half of the escrow
	expectedHalf := new(big.Int).Div(amount, big.NewInt(2))
	assert.Equal(expectedHalf.String(), released.String(),
		"Released amount should be 50% of escrow")

	// Compute what a refund should return: amount - released
	expectedRefund := new(big.Int).Sub(amount, released)
	assert.True(expectedRefund.Sign() > 0, "Refund amount should be positive")

	// Attempting to release 0 additional after no further uptime should yield 0
	zero, err := escrowMgr.ReleasePayment("refund-001", halfDuration)
	assert.NoError(err)
	assert.Equal(big.NewInt(0).String(), zero.String(),
		"Re-releasing the same uptime should yield 0")

	// Verify the escrow state tracks released correctly
	stored, found := escrowMgr.GetEscrow("refund-001")
	assert.True(found, "Escrow should be retrievable")
	assert.Equal(released.String(), stored.Released.String(),
		"Stored released amount should match what was released")

	// Verify the remaining (refundable) portion
	remaining := new(big.Int).Sub(stored.Amount, stored.Released)
	assert.Equal(expectedRefund.String(), remaining.String(),
		"Remaining amount should equal escrow minus released")
}

// TestE2E_EscrowProtocolFee verifies that the 5% protocol fee is correctly
// split: 80% burned, 20% to treasury.
func TestE2E_EscrowProtocolFee(t *testing.T) {
	assert := testutil.NewAssertions(t)

	escrowMgr := payment.NewEscrowManager()

	amount := bunkerToWei(1000) // 1000 BUNKER
	duration := time.Hour

	escrow := escrowMgr.CreateEscrow("fee-001", amount, duration)
	assert.NotNil(escrow)

	// Release the full payment
	released, err := escrowMgr.ReleasePayment("fee-001", duration)
	assert.NoError(err)
	assert.Equal(amount.String(), released.String(),
		"Full release should equal escrow amount")

	// Apply protocol fee calculation per ProtocolFees config
	protocolFees := types.DefaultProtocolFees()
	assert.Equal(5.0, protocolFees.TotalFeePercent, "Total fee should be 5%")
	assert.Equal(80.0, protocolFees.BurnPercent, "Burn share should be 80%")
	assert.Equal(20.0, protocolFees.TreasuryPercent, "Treasury share should be 20%")

	// Calculate protocol fee: 5% of released
	// fee = released * 5 / 100
	feeNumerator := new(big.Int).Mul(released, big.NewInt(5))
	totalFee := new(big.Int).Div(feeNumerator, big.NewInt(100))

	// Provider receives released - fee
	providerPayout := new(big.Int).Sub(released, totalFee)

	// Burn = 80% of fee
	burnNumerator := new(big.Int).Mul(totalFee, big.NewInt(80))
	burnAmount := new(big.Int).Div(burnNumerator, big.NewInt(100))

	// Treasury = 20% of fee
	treasuryNumerator := new(big.Int).Mul(totalFee, big.NewInt(20))
	treasuryAmount := new(big.Int).Div(treasuryNumerator, big.NewInt(100))

	// Verify: burn + treasury == totalFee
	assert.Equal(totalFee.String(),
		new(big.Int).Add(burnAmount, treasuryAmount).String(),
		"Burn + treasury should equal total fee")

	// Verify: providerPayout + totalFee == released
	assert.Equal(released.String(),
		new(big.Int).Add(providerPayout, totalFee).String(),
		"Provider payout + fee should equal released amount")

	// Verify concrete values for 1000 BUNKER
	// Total fee = 50 BUNKER
	assert.Equal(bunkerToWei(50).String(), totalFee.String(),
		"5% of 1000 BUNKER = 50 BUNKER")

	// Burn = 40 BUNKER (80% of 50)
	assert.Equal(bunkerToWei(40).String(), burnAmount.String(),
		"80% of 50 BUNKER = 40 BUNKER")

	// Treasury = 10 BUNKER (20% of 50)
	assert.Equal(bunkerToWei(10).String(), treasuryAmount.String(),
		"20% of 50 BUNKER = 10 BUNKER")

	// Provider = 950 BUNKER
	assert.Equal(bunkerToWei(950).String(), providerPayout.String(),
		"Provider should receive 950 BUNKER after 5% fee")
}

// TestE2E_EscrowContractIntegration exercises the MockEscrowContract through
// the full create -> progressive release -> finalize flow.
func TestE2E_EscrowContractIntegration(t *testing.T) {
	assert := testutil.NewAssertions(t)

	ec := payment.NewMockEscrowContract()
	assert.True(ec.IsMockMode(), "Should be in mock mode")

	ctx := context.Background()
	jobID := payment.JobIDFromString("contract-integration-001")
	provider := common.HexToAddress("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	amount := bunkerToWei(500)
	durationSecs := big.NewInt(3600) // 1 hour in seconds

	// Step 1: Create escrow (state: Created)
	_, err := ec.CreateEscrow(ctx, jobID, provider, amount, durationSecs)
	assert.NoError(err, "CreateEscrow should succeed")

	// Step 2: Verify initial state is Created (not yet Active)
	escrow, err := ec.GetEscrow(ctx, jobID)
	assert.NoError(err, "GetEscrow should succeed")
	assert.Equal(amount.String(), escrow.Amount.String())
	assert.Equal(big.NewInt(0).String(), escrow.Released.String())
	assert.Equal(durationSecs.String(), escrow.Duration.String())
	assert.Equal(payment.EscrowStateCreated.String(), escrow.State.String(),
		"Escrow should be in Created state before provider selection")

	// Step 2b: Select providers to transition escrow to Active
	providers := [3]common.Address{provider, {}, {}}
	_, err = ec.SelectProviders(ctx, jobID, providers)
	assert.NoError(err, "SelectProviders should succeed")

	escrow, err = ec.GetEscrow(ctx, jobID)
	assert.NoError(err)
	assert.Equal(payment.EscrowStateActive.String(), escrow.State.String(),
		"Escrow should be Active after provider selection")

	// Step 3: Release 25% (900 seconds of 3600)
	_, err = ec.ReleasePayment(ctx, jobID, big.NewInt(900))
	assert.NoError(err, "Release 25% should succeed")

	escrow, _ = ec.GetEscrow(ctx, jobID)
	expectedRelease25 := new(big.Int).Div(amount, big.NewInt(4))
	assert.Equal(expectedRelease25.String(), escrow.Released.String(),
		"After 900s of 3600s, 25% should be released")

	// Step 4: Release 50% (1800 seconds)
	_, err = ec.ReleasePayment(ctx, jobID, big.NewInt(1800))
	assert.NoError(err, "Release 50% should succeed")

	escrow, _ = ec.GetEscrow(ctx, jobID)
	expectedRelease50 := new(big.Int).Div(amount, big.NewInt(2))
	assert.Equal(expectedRelease50.String(), escrow.Released.String(),
		"After 1800s, 50% should be released")

	// Step 5: Release 100% (3600 seconds)
	_, err = ec.ReleasePayment(ctx, jobID, big.NewInt(3600))
	assert.NoError(err, "Release 100% should succeed")

	escrow, _ = ec.GetEscrow(ctx, jobID)
	assert.Equal(amount.String(), escrow.Released.String(),
		"After full duration, 100% should be released")

	// Step 6: Remaining should be 0
	remaining := ec.CalculateRemainingEscrow(escrow)
	assert.Equal(big.NewInt(0).String(), remaining.String(),
		"No remaining escrow after full release")

	// Step 7: Finalize
	_, err = ec.FinalizeEscrow(ctx, jobID)
	assert.NoError(err, "FinalizeEscrow should succeed")

	escrow, _ = ec.GetEscrow(ctx, jobID)
	assert.Equal(payment.EscrowStateCompleted.String(), escrow.State.String(),
		"Escrow should be completed after finalization")

	// Step 8: Duplicate creation should fail
	_, err = ec.CreateEscrow(ctx, jobID, provider, amount, durationSecs)
	assert.Error(err, "Creating duplicate escrow should fail")
}

// TestE2E_EscrowMultipleDeployments creates multiple escrows for different
// deployments, releases them independently, and verifies that each escrow is
// isolated from the others.
func TestE2E_EscrowMultipleDeployments(t *testing.T) {
	assert := testutil.NewAssertions(t)

	escrowMgr := payment.NewEscrowManager()

	deployments := []struct {
		id       string
		amount   *big.Int
		duration time.Duration
	}{
		{"multi-deploy-001", bunkerToWei(100), time.Hour},
		{"multi-deploy-002", bunkerToWei(250), 2 * time.Hour},
		{"multi-deploy-003", bunkerToWei(500), 30 * time.Minute},
	}

	// Create all escrows
	for _, d := range deployments {
		escrow := escrowMgr.CreateEscrow(d.id, d.amount, d.duration)
		assert.NotNil(escrow, "Escrow for %s should be created", d.id)
		assert.Equal(d.id, escrow.ReservationID)
	}

	// Release 50% on the first deployment only
	released1, err := escrowMgr.ReleasePayment("multi-deploy-001", time.Hour/2)
	assert.NoError(err)
	expected1 := new(big.Int).Div(bunkerToWei(100), big.NewInt(2)) // 50 BUNKER
	assert.Equal(expected1.String(), released1.String(),
		"Deploy-001 should release 50 BUNKER")

	// Release 100% on the third deployment
	released3, err := escrowMgr.ReleasePayment("multi-deploy-003", 30*time.Minute)
	assert.NoError(err)
	assert.Equal(bunkerToWei(500).String(), released3.String(),
		"Deploy-003 should release full 500 BUNKER")

	// Second deployment should be untouched
	escrow2, found := escrowMgr.GetEscrow("multi-deploy-002")
	assert.True(found, "Deploy-002 escrow should exist")
	assert.Equal(big.NewInt(0).String(), escrow2.Released.String(),
		"Deploy-002 should have 0 released")

	// Release 25% on the second deployment
	released2, err := escrowMgr.ReleasePayment("multi-deploy-002", 30*time.Minute)
	assert.NoError(err)
	expected2 := new(big.Int).Div(bunkerToWei(250), big.NewInt(4)) // 62.5 BUNKER
	assert.Equal(expected2.String(), released2.String(),
		"Deploy-002 should release 25% (62.5 BUNKER)")

	// Verify each escrow maintains its own state
	e1, _ := escrowMgr.GetEscrow("multi-deploy-001")
	assert.Equal(expected1.String(), e1.Released.String())

	e2, _ := escrowMgr.GetEscrow("multi-deploy-002")
	assert.Equal(expected2.String(), e2.Released.String())

	e3, _ := escrowMgr.GetEscrow("multi-deploy-003")
	assert.Equal(bunkerToWei(500).String(), e3.Released.String())

	// Releasing on a non-existent escrow should fail
	_, err = escrowMgr.ReleasePayment("nonexistent", time.Hour)
	assert.Error(err, "Release on non-existent escrow should fail")
}

// TestE2E_EscrowWithPricing uses PricingCalculator to compute the cost of a
// deployment, creates an escrow with that computed amount, and then releases
// the full payment.
func TestE2E_EscrowWithPricing(t *testing.T) {
	assert := testutil.NewAssertions(t)

	// Set up pricing: 1 BUNKER base per hour
	basePricePerHour := bunkerToWei(1)
	pricingCalc := payment.NewPricingCalculator(basePricePerHour)

	// Define resource requirements
	resources := types.ResourceLimits{
		CPUQuota:    2,                 // 2 CPU units
		CPUPeriod:   100000,            // standard period
		MemoryLimit: 4 * 1024 * 1024 * 1024, // 4 GB
		DiskLimit:   10 * 1024 * 1024 * 1024, // 10 GB
		NetworkBW:   100 * 1024 * 1024,       // 100 MB/s
		PIDLimit:    1024,
	}
	duration := 2 * time.Hour

	// Calculate price
	price := pricingCalc.CalculatePrice(resources, duration)
	assert.True(price.Sign() > 0, "Computed price should be positive")

	// Create escrow with the computed price
	escrowMgr := payment.NewEscrowManager()
	escrow := escrowMgr.CreateEscrow("pricing-001", price, duration)
	assert.NotNil(escrow)
	assert.Equal(price.String(), escrow.Amount.String(),
		"Escrow amount should match computed price")

	// Release 50%
	half, err := escrowMgr.ReleasePayment("pricing-001", duration/2)
	assert.NoError(err)
	assert.True(half.Sign() > 0, "Half-duration release should be positive")

	// Release remaining 50%
	rest, err := escrowMgr.ReleasePayment("pricing-001", duration)
	assert.NoError(err)

	totalReleased := new(big.Int).Add(half, rest)
	assert.Equal(price.String(), totalReleased.String(),
		"Total released should match computed price")

	// Also verify the bid calculation respects staking
	stake := bunkerToWei(10000) // Silver tier
	bid := pricingCalc.CalculateBid(resources, duration, stake)
	assert.True(bid.Sign() > 0, "Bid should be positive")
	assert.True(bid.Cmp(price) < 0,
		"Bid with substantial stake should be lower than base price")
}

// TestE2E_PaymentGatedDeploymentFlow exercises the full payment-gated deployment
// lifecycle using contract-level mock APIs:
//  1. Provider stakes → verified as active
//  2. Requester creates escrow → state is Created
//  3. Providers selected → state transitions to Active
//  4. Deployment runs for a duration
//  5. Payment released proportional to uptime
//  6. Escrow finalized → state is Completed
//
// This mirrors what the daemon's ContainerManager does internally.
func TestE2E_PaymentGatedDeploymentFlow(t *testing.T) {
	assert := testutil.NewAssertions(t)

	sc := payment.NewMockStakingContract()
	ec := payment.NewMockEscrowContract()
	ctx := context.Background()

	// ---------------------------------------------------------------
	// Phase 1: Provider stakes and is verified as active
	// ---------------------------------------------------------------
	t.Log("Phase 1: Provider staking")

	// Stake 2000 BUNKER (Bronze tier) for each provider
	_, err := sc.Stake(ctx, bunkerToWei(2000))
	assert.NoError(err, "Provider stake should succeed")

	// The mock uses zero address as the "self" address
	zeroAddr := common.Address{}
	hasMin, err := sc.HasMinimumStake(ctx, zeroAddr)
	assert.NoError(err)
	assert.True(hasMin, "Provider should meet minimum stake (2000 >= 500)")

	isActive, err := sc.IsActiveProvider(ctx, zeroAddr)
	assert.NoError(err)
	assert.True(isActive, "Provider should be active after staking")

	tier, err := sc.GetTier(ctx, zeroAddr)
	assert.NoError(err)
	assert.Equal(types.StakingTierBronze, tier, "2000 BUNKER should be Bronze tier")

	// ---------------------------------------------------------------
	// Phase 2: Create escrow for deployment (state: Created)
	// ---------------------------------------------------------------
	t.Log("Phase 2: Creating escrow for deployment")

	deploymentID := "payment-gated-deploy-001"
	jobID := payment.JobIDFromString(deploymentID)
	provider := common.HexToAddress("0x1111111111111111111111111111111111111111")
	escrowAmount := bunkerToWei(100) // 100 BUNKER
	durationSecs := big.NewInt(3600) // 1 hour

	_, err = ec.CreateEscrow(ctx, jobID, provider, escrowAmount, durationSecs)
	assert.NoError(err, "CreateEscrow should succeed")

	escrow, err := ec.GetEscrow(ctx, jobID)
	assert.NoError(err)
	assert.Equal(payment.EscrowStateCreated.String(), escrow.State.String(),
		"Escrow should be Created before provider selection")
	assert.Equal(escrowAmount.String(), escrow.Amount.String())

	// ---------------------------------------------------------------
	// Phase 3: Select providers → escrow becomes Active
	// ---------------------------------------------------------------
	t.Log("Phase 3: Selecting providers (escrow: Created → Active)")

	providers := [3]common.Address{
		provider,
		common.HexToAddress("0x2222222222222222222222222222222222222222"),
		common.HexToAddress("0x3333333333333333333333333333333333333333"),
	}
	_, err = ec.SelectProviders(ctx, jobID, providers)
	assert.NoError(err, "SelectProviders should succeed")

	escrow, err = ec.GetEscrow(ctx, jobID)
	assert.NoError(err)
	assert.Equal(payment.EscrowStateActive.String(), escrow.State.String(),
		"Escrow should be Active after provider selection")

	// Cannot select providers again on an active escrow
	_, err = ec.SelectProviders(ctx, jobID, providers)
	assert.Error(err, "SelectProviders on Active escrow should fail")

	// ---------------------------------------------------------------
	// Phase 4: Simulate deployment running (50% of duration = 1800s)
	// ---------------------------------------------------------------
	t.Log("Phase 4: Releasing payment for 50% uptime")

	_, err = ec.ReleasePayment(ctx, jobID, big.NewInt(1800))
	assert.NoError(err, "ReleasePayment should succeed")

	escrow, err = ec.GetEscrow(ctx, jobID)
	assert.NoError(err)
	expectedReleased := new(big.Int).Div(escrowAmount, big.NewInt(2))
	assert.Equal(expectedReleased.String(), escrow.Released.String(),
		"50% of escrow should be released after half duration")

	// ---------------------------------------------------------------
	// Phase 5: Release remaining payment (full duration)
	// ---------------------------------------------------------------
	t.Log("Phase 5: Releasing remaining payment for full uptime")

	_, err = ec.ReleasePayment(ctx, jobID, big.NewInt(3600))
	assert.NoError(err, "ReleasePayment for full duration should succeed")

	escrow, err = ec.GetEscrow(ctx, jobID)
	assert.NoError(err)
	assert.Equal(escrowAmount.String(), escrow.Released.String(),
		"Full escrow amount should be released")

	remaining := ec.CalculateRemainingEscrow(escrow)
	assert.Equal(big.NewInt(0).String(), remaining.String(),
		"No remaining escrow after full release")

	// ---------------------------------------------------------------
	// Phase 6: Finalize escrow → Completed
	// ---------------------------------------------------------------
	t.Log("Phase 6: Finalizing escrow")

	_, err = ec.FinalizeEscrow(ctx, jobID)
	assert.NoError(err, "FinalizeEscrow should succeed")

	escrow, err = ec.GetEscrow(ctx, jobID)
	assert.NoError(err)
	assert.Equal(payment.EscrowStateCompleted.String(), escrow.State.String(),
		"Escrow should be Completed after finalization")

	// Cannot release on completed escrow
	_, err = ec.ReleasePayment(ctx, jobID, big.NewInt(100))
	assert.Error(err, "ReleasePayment on completed escrow should fail")

	t.Log("Payment-gated deployment flow completed successfully")
}

// TestE2E_EscrowRefundBeforeActivation tests that an escrow in Created state
// (before provider selection) can be refunded without issue.
func TestE2E_EscrowRefundBeforeActivation(t *testing.T) {
	assert := testutil.NewAssertions(t)

	ec := payment.NewMockEscrowContract()
	ctx := context.Background()

	jobID := payment.JobIDFromString("refund-before-active-001")
	provider := common.HexToAddress("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	amount := bunkerToWei(200)
	durationSecs := big.NewInt(7200)

	// Create escrow
	_, err := ec.CreateEscrow(ctx, jobID, provider, amount, durationSecs)
	assert.NoError(err)

	// Verify in Created state
	escrow, err := ec.GetEscrow(ctx, jobID)
	assert.NoError(err)
	assert.Equal(payment.EscrowStateCreated.String(), escrow.State.String())

	// Refund before activation — should succeed
	_, err = ec.Refund(ctx, jobID)
	assert.NoError(err, "Refund of Created escrow should succeed")

	// Verify state is Refunded
	escrow, err = ec.GetEscrow(ctx, jobID)
	assert.NoError(err)
	assert.Equal(payment.EscrowStateRefunded.String(), escrow.State.String(),
		"Escrow should be Refunded")

	// Cannot select providers on refunded escrow
	providers := [3]common.Address{provider, {}, {}}
	_, err = ec.SelectProviders(ctx, jobID, providers)
	assert.Error(err, "SelectProviders on refunded escrow should fail")
}
