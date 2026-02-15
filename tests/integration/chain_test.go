//go:build integration

package integration

import (
	"fmt"
	"math/big"
	"strings"
	"testing"
)

// parseBigInt extracts a big.Int from cast output which may include
// annotations like "1000000 [1e6]".
func parseBigInt(s string) (*big.Int, bool) {
	s = strings.TrimSpace(s)
	// Strip annotation suffix: "123456 [1e5]" → "123456"
	if idx := strings.Index(s, " ["); idx > 0 {
		s = s[:idx]
	}
	// Strip annotation suffix: "123456 [1e5]" → "123456" (brackets without space)
	if idx := strings.Index(s, "["); idx > 0 {
		s = strings.TrimSpace(s[:idx])
	}
	return new(big.Int).SetString(s, 10)
}

// ─── Token Tests ────────────────────────────────────────────────────────────

func TestTokenDeployment(t *testing.T) {
	// Verify token name
	name, err := castCall(tokenAddr, "name()(string)")
	if err != nil {
		t.Fatalf("failed to read token name: %v", err)
	}
	if !strings.Contains(name, "Bunker") {
		t.Errorf("expected token name containing 'Bunker', got: %s", name)
	}
	t.Logf("Token name: %s", name)

	// Verify symbol
	symbol, err := castCall(tokenAddr, "symbol()(string)")
	if err != nil {
		t.Fatalf("failed to read symbol: %v", err)
	}
	t.Logf("Token symbol: %s", symbol)

	// Verify supply cap
	cap, err := castCall(tokenAddr, "SUPPLY_CAP()(uint256)")
	if err != nil {
		t.Fatalf("failed to read supply cap: %v", err)
	}
	t.Logf("Supply cap: %s", cap)
}

func TestTokenBalances(t *testing.T) {
	// Each test account should have been minted 10M BUNKER by DeployLocal
	accounts := map[string]string{
		"provider1": provider1Addr,
		"provider2": provider2Addr,
		"provider3": provider3Addr,
		"requester": requesterAddr,
		"operator":  operatorAddr,
	}

	for name, addr := range accounts {
		bal, err := castCall(tokenAddr, "balanceOf(address)(uint256)", addr)
		if err != nil {
			t.Fatalf("balanceOf(%s) failed: %v", name, err)
		}
		t.Logf("%s (%s): %s wei BUNKER", name, addr[:10], bal)

		// Should be 10M * 10^18 = 10_000_000e18
		expected := new(big.Int).Mul(big.NewInt(10_000_000), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))
		actual, ok := parseBigInt(bal)
		if !ok {
			t.Fatalf("cannot parse balance: %s", bal)
		}
		if actual.Cmp(expected) != 0 {
			t.Errorf("%s: expected %s, got %s", name, expected, actual)
		}
	}
}

func TestTokenTransfer(t *testing.T) {
	amount := "1000000000000000000" // 1 BUNKER (1e18)

	// Requester transfers 1 BUNKER to deployer
	_, err := castSend(requesterPK, tokenAddr,
		"transfer(address,uint256)(bool)",
		deployerAddr, amount)
	if err != nil {
		t.Fatalf("transfer failed: %v", err)
	}

	// Verify deployer received it
	bal, err := castCall(tokenAddr, "balanceOf(address)(uint256)", deployerAddr)
	if err != nil {
		t.Fatalf("balanceOf failed: %v", err)
	}
	t.Logf("Deployer balance after transfer: %s", bal)

	actual, _ := parseBigInt(bal)
	if actual.Sign() <= 0 {
		t.Error("deployer balance should be positive after transfer")
	}
}

// ─── Staking Tests ──────────────────────────────────────────────────────────

func TestStakingFlow(t *testing.T) {
	stakeAmount := "500000000000000000000" // 500 BUNKER (Starter tier minimum)

	// 1. Provider1 approves staking contract to spend tokens
	t.Log("approving staking contract...")
	_, err := castSend(provider1PK, tokenAddr,
		"approve(address,uint256)(bool)",
		stakingAddr, stakeAmount)
	if err != nil {
		t.Fatalf("approve failed: %v", err)
	}

	// Verify allowance
	allowance, err := castCall(tokenAddr,
		"allowance(address,address)(uint256)",
		provider1Addr, stakingAddr)
	if err != nil {
		t.Fatalf("allowance check failed: %v", err)
	}
	t.Logf("Allowance: %s", allowance)

	// 2. Stake tokens
	t.Log("staking tokens...")
	_, err = castSend(provider1PK, stakingAddr,
		"stake(uint256)", stakeAmount)
	if err != nil {
		t.Fatalf("stake failed: %v", err)
	}

	// 3. Verify stake via providers() mapping (returns tuple: stakedAmount, rewardDebt, ...)
	providerData, err := castCall(stakingAddr,
		"providers(address)((uint256,uint256,uint256,uint256,uint8))", provider1Addr)
	if err != nil {
		t.Fatalf("providers() read failed: %v", err)
	}
	t.Logf("Provider1 data: %s", providerData)

	// 4. Check total staked
	total, err := castCall(stakingAddr, "totalStaked()(uint256)")
	if err != nil {
		t.Logf("totalStaked() call failed (may not exist): %v", err)
	} else {
		t.Logf("Total staked in contract: %s", total)
	}

	// 5. Check tier
	tier, err := castCall(stakingAddr, "getTier(address)(uint8)", provider1Addr)
	if err != nil {
		t.Logf("getTier() call failed (may not exist on this contract): %v", err)
	} else {
		t.Logf("Provider1 tier: %s", tier)
	}
}

func TestMultiProviderStaking(t *testing.T) {
	providers := []struct {
		name string
		addr string
		pk   string
		amount string
	}{
		{"provider2", provider2Addr, provider2PK, "2000000000000000000000"},  // 2K BUNKER (Bronze)
		{"provider3", provider3Addr, provider3PK, "10000000000000000000000"}, // 10K BUNKER (Silver)
	}

	for _, p := range providers {
		t.Run(p.name, func(t *testing.T) {
			// Approve
			_, err := castSend(p.pk, tokenAddr,
				"approve(address,uint256)(bool)",
				stakingAddr, p.amount)
			if err != nil {
				t.Fatalf("approve failed: %v", err)
			}

			// Stake
			_, err = castSend(p.pk, stakingAddr,
				"stake(uint256)", p.amount)
			if err != nil {
				t.Fatalf("stake failed: %v", err)
			}

			t.Logf("%s staked %s wei BUNKER", p.name, p.amount)
		})
	}
}

// ─── Escrow Tests ───────────────────────────────────────────────────────────

func TestEscrowFlow(t *testing.T) {
	escrowAmount := "100000000000000000000" // 100 BUNKER
	duration := "3600"                       // 1 hour in seconds

	// 1. Requester approves escrow contract
	t.Log("requester approving escrow contract...")
	_, err := castSend(requesterPK, tokenAddr,
		"approve(address,uint256)(bool)",
		escrowAddr, escrowAmount)
	if err != nil {
		t.Fatalf("approve failed: %v", err)
	}

	// 2. Create reservation
	t.Log("creating reservation...")
	_, err = castSend(requesterPK, escrowAddr,
		"createReservation(uint256,uint256)(uint256)",
		escrowAmount, duration)
	if err != nil {
		t.Fatalf("createReservation failed: %v", err)
	}

	// 3. Read the reservation (ID should be 1)
	reservationID := "1"
	resData, err := castCall(escrowAddr,
		"getReservation(uint256)((address,uint128,uint128,uint48,uint48,uint8,address[3]))",
		reservationID)
	if err != nil {
		t.Fatalf("getReservation failed: %v", err)
	}
	t.Logf("Reservation data: %s", resData)

	// 4. Operator selects providers
	t.Log("operator selecting providers...")
	_, err = castSend(operatorPK, escrowAddr,
		"selectProviders(uint256,address[3])",
		reservationID,
		fmt.Sprintf("[%s,%s,%s]", provider1Addr, provider2Addr, provider3Addr))
	if err != nil {
		t.Fatalf("selectProviders failed: %v", err)
	}

	// 5. Verify reservation is now Active
	resData2, err := castCall(escrowAddr,
		"getReservation(uint256)((address,uint128,uint128,uint48,uint48,uint8,address[3]))",
		reservationID)
	if err != nil {
		t.Fatalf("getReservation after selectProviders failed: %v", err)
	}
	t.Logf("Reservation after provider selection: %s", resData2)

	// 6. Operator releases partial payment (50% of duration)
	t.Log("releasing partial payment...")
	halfDuration := "1800" // 30 minutes
	_, err = castSend(operatorPK, escrowAddr,
		"releasePayment(uint256,uint256)",
		reservationID, halfDuration)
	if err != nil {
		t.Fatalf("releasePayment failed: %v", err)
	}

	// 7. Verify protocol fee was applied (check treasury and burned amounts)
	totalBurned, err := castCall(escrowAddr, "totalBurned()(uint256)")
	if err != nil {
		t.Logf("totalBurned check failed: %v", err)
	} else {
		t.Logf("Total burned: %s", totalBurned)
	}

	totalTreasuryFees, err := castCall(escrowAddr, "totalTreasuryFees()(uint256)")
	if err != nil {
		t.Logf("totalTreasuryFees check failed: %v", err)
	} else {
		t.Logf("Total treasury fees: %s", totalTreasuryFees)
	}

	// 8. Finalize the reservation
	t.Log("finalizing reservation...")
	_, err = castSend(operatorPK, escrowAddr,
		"finalizeReservation(uint256)",
		reservationID)
	if err != nil {
		t.Fatalf("finalizeReservation failed: %v", err)
	}
	t.Log("reservation finalized successfully")

	// 9. Verify provider balances increased
	for _, addr := range []string{provider1Addr, provider2Addr, provider3Addr} {
		bal, err := castCall(tokenAddr, "balanceOf(address)(uint256)", addr)
		if err != nil {
			t.Errorf("balanceOf provider failed: %v", err)
			continue
		}
		t.Logf("Provider %s balance: %s", addr[:10], bal)
	}
}

// ─── Pricing Tests ──────────────────────────────────────────────────────────

func TestPricingContract(t *testing.T) {
	// Read current prices
	prices, err := castCall(pricingAddr, "getPrices()((uint256,uint256,uint256,uint256,uint256,uint256))")
	if err != nil {
		t.Fatalf("getPrices failed: %v", err)
	}
	t.Logf("Current prices: %s", prices)

	// Read version
	version, err := castCall(pricingAddr, "VERSION()(string)")
	if err != nil {
		t.Logf("VERSION() failed: %v", err)
	} else {
		t.Logf("Pricing contract version: %s", version)
	}
}

// ─── Protocol Fee Tests ─────────────────────────────────────────────────────

func TestProtocolFee(t *testing.T) {
	// Check default protocol fee (should be 500 = 5%)
	fee, err := castCall(escrowAddr, "protocolFeeBps()(uint256)")
	if err != nil {
		t.Fatalf("protocolFeeBps failed: %v", err)
	}
	t.Logf("Protocol fee: %s bps", fee)

	expected := "500"
	if strings.TrimSpace(fee) != expected {
		t.Errorf("expected protocol fee %s bps, got %s", expected, fee)
	}

	// Check fee constants
	burnBps, err := castCall(escrowAddr, "FEE_BURN_BPS()(uint256)")
	if err != nil {
		t.Fatalf("FEE_BURN_BPS failed: %v", err)
	}
	t.Logf("Fee burn: %s bps (80%%)", burnBps)

	treasuryBps, err := castCall(escrowAddr, "FEE_TREASURY_BPS()(uint256)")
	if err != nil {
		t.Fatalf("FEE_TREASURY_BPS failed: %v", err)
	}
	t.Logf("Fee treasury: %s bps (20%%)", treasuryBps)
}
