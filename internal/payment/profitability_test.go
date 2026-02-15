package payment

import (
	"math"
	"math/big"
	"testing"
)

// newTestProfitabilityCalculator creates a ProfitabilityCalculator with a
// standard base price suitable for tests.
func newTestProfitabilityCalculator() *ProfitabilityCalculator {
	basePrice := big.NewInt(1000000000000000) // 0.001 BUNKER per hour
	pricing := NewPricingCalculator(basePrice)
	return NewProfitabilityCalculator(pricing)
}

// standardResources returns a typical provider resource spec for tests:
// 4 CPU cores, 8 GB memory, 100 GB storage.
func standardResources() ResourceSpec {
	return ResourceSpec{
		CPUCores:  4,
		MemoryGB:  8,
		StorageGB: 100,
	}
}

// zeroCosts returns zero costs for tests that isolate revenue calculations.
func zeroCosts() ProviderCosts {
	return ProviderCosts{}
}

// tinyCosts returns costs proportional to the pricing calculator's per-job
// revenue (~4e-9 BUNKER). This ensures tests can verify profitability logic
// with realistic relative magnitudes.
func tinyCosts() ProviderCosts {
	return ProviderCosts{
		CPUCostPerCoreHour:    1e-11,
		MemoryCostPerGBHour:   1e-11,
		StorageCostPerGBMonth: 1e-10,
		NetworkCostPerGB:      0,
		FixedMonthlyCost:      1e-8, // Roughly 2-3 jobs worth of revenue
	}
}

func TestProfitabilityCalculator_BasicCalculation(t *testing.T) {
	pc := newTestProfitabilityCalculator()
	resources := standardResources()
	costs := tinyCosts()

	report := pc.Calculate(resources, costs, 1_000_000, 100) // Starter tier, 100 jobs/month

	if report == nil {
		t.Fatal("report should not be nil")
	}

	// Revenue checks
	if report.GrossRevenue <= 0 {
		t.Errorf("gross revenue should be positive for non-zero jobs, got %e", report.GrossRevenue)
	}
	if report.ProtocolFees <= 0 {
		t.Errorf("protocol fees should be positive when there is revenue, got %e", report.ProtocolFees)
	}
	if report.NetRevenue >= report.GrossRevenue {
		t.Error("net revenue should be less than gross revenue after protocol fees")
	}
	if report.NetRevenue <= 0 {
		t.Error("net revenue should be positive")
	}

	// Cost checks
	if report.InfrastructureCost <= 0 {
		t.Errorf("infrastructure cost should be positive with non-zero costs and jobs, got %e", report.InfrastructureCost)
	}
	if report.FixedCost != costs.FixedMonthlyCost {
		t.Errorf("fixed cost should equal input: got %e, want %e", report.FixedCost, costs.FixedMonthlyCost)
	}
	if report.TotalCost <= 0 {
		t.Error("total cost should be positive")
	}

	// Staking checks
	if report.StakeTier != "starter" {
		t.Errorf("expected starter tier for 1,000,000 BUNKER, got %q", report.StakeTier)
	}
	if report.StakeAmount != 1_000_000 {
		t.Errorf("stake amount should be 1000000, got %f", report.StakeAmount)
	}
	if report.RewardMultiplier != 1.0 {
		t.Errorf("starter tier multiplier should be 1.0, got %f", report.RewardMultiplier)
	}
}

func TestProfitabilityCalculator_ProtocolFeeDeduction(t *testing.T) {
	pc := newTestProfitabilityCalculator()
	resources := standardResources()
	costs := zeroCosts()

	report := pc.Calculate(resources, costs, 500, 100)

	expectedFees := report.GrossRevenue * 0.05 // 5% protocol fee
	tolerance := 1e-20

	if math.Abs(report.ProtocolFees-expectedFees) > tolerance {
		t.Errorf("protocol fees mismatch: got %e, want %e", report.ProtocolFees, expectedFees)
	}

	expectedNet := report.GrossRevenue - report.ProtocolFees
	if math.Abs(report.NetRevenue-expectedNet) > tolerance {
		t.Errorf("net revenue mismatch: got %e, want %e", report.NetRevenue, expectedNet)
	}

	// With zero costs, net profit should equal net revenue
	if math.Abs(report.NetProfit-report.NetRevenue) > tolerance {
		t.Errorf("with zero costs, net profit should equal net revenue: got %e, want %e",
			report.NetProfit, report.NetRevenue)
	}
}

func TestProfitabilityCalculator_ZeroJobs(t *testing.T) {
	pc := newTestProfitabilityCalculator()
	resources := standardResources()
	costs := tinyCosts()

	report := pc.Calculate(resources, costs, 500, 0)

	if report.GrossRevenue != 0 {
		t.Errorf("gross revenue should be 0 with 0 jobs, got %e", report.GrossRevenue)
	}
	if report.ProtocolFees != 0 {
		t.Errorf("protocol fees should be 0 with 0 jobs, got %e", report.ProtocolFees)
	}
	if report.NetRevenue != 0 {
		t.Errorf("net revenue should be 0 with 0 jobs, got %e", report.NetRevenue)
	}
	// With zero revenue but fixed costs, profit should be negative
	if report.NetProfit >= 0 {
		t.Error("net profit should be negative with zero jobs and fixed costs")
	}
	if !math.IsInf(report.PaybackPeriod, 1) {
		t.Error("payback period should be infinite when not profitable")
	}
}

func TestProfitabilityCalculator_ZeroCosts(t *testing.T) {
	pc := newTestProfitabilityCalculator()
	resources := standardResources()
	costs := zeroCosts()

	report := pc.Calculate(resources, costs, 500, 100)

	if report.TotalCost != 0 {
		t.Errorf("total cost should be 0 with zero costs, got %e", report.TotalCost)
	}
	if report.InfrastructureCost != 0 {
		t.Errorf("infrastructure cost should be 0, got %e", report.InfrastructureCost)
	}
	if report.NetProfit <= 0 {
		t.Error("should be profitable with zero costs and non-zero jobs")
	}
	if report.ProfitMargin != 100 {
		t.Errorf("profit margin should be 100%% with zero costs, got %f", report.ProfitMargin)
	}
}

func TestProfitabilityCalculator_TierResolution(t *testing.T) {
	pc := newTestProfitabilityCalculator()
	resources := standardResources()
	costs := zeroCosts()

	tests := []struct {
		stake    float64
		wantTier string
		wantMult float64
	}{
		{100_000, "", 1.0},                 // Below minimum
		{1_000_000, "starter", 1.0},        // Starter minimum
		{5_000_000, "bronze", 1.0},         // Bronze minimum
		{10_000_000, "silver", 1.05},       // Silver minimum
		{100_000_000, "gold", 1.1},         // Gold minimum
		{1_000_000_000, "platinum", 1.2},   // Platinum minimum
		{10_000_000_000, "platinum", 1.2},  // Well above platinum
	}

	for _, tt := range tests {
		report := pc.Calculate(resources, costs, tt.stake, 100)
		if report.StakeTier != tt.wantTier {
			t.Errorf("stake=%f: got tier %q, want %q", tt.stake, report.StakeTier, tt.wantTier)
		}
		if report.RewardMultiplier != tt.wantMult {
			t.Errorf("stake=%f: got multiplier %f, want %f", tt.stake, report.RewardMultiplier, tt.wantMult)
		}
	}
}

func TestProfitabilityCalculator_BreakEvenAnalysis(t *testing.T) {
	pc := newTestProfitabilityCalculator()
	resources := standardResources()
	costs := tinyCosts()

	breakEven := pc.BreakEvenAnalysis(resources, costs, 100_000)

	if breakEven <= 0 {
		t.Error("break-even should require at least 1 job with fixed costs")
	}
	if breakEven == math.MaxInt {
		t.Fatal("break-even should be achievable with tiny costs")
	}

	// Verify: at break-even jobs, profit should be near zero or positive
	report := pc.Calculate(resources, costs, 100_000, breakEven)
	if report.NetProfit < -1e-10 { // Allow small floating point error
		t.Errorf("at break-even (%d jobs), profit should be near zero, got %e",
			breakEven, report.NetProfit)
	}

	// One less job should be less profitable
	if breakEven > 1 {
		reportBelow := pc.Calculate(resources, costs, 100_000, breakEven-1)
		if reportBelow.NetProfit > report.NetProfit {
			t.Errorf("fewer jobs should not be more profitable: %d jobs=%e, %d jobs=%e",
				breakEven-1, reportBelow.NetProfit, breakEven, report.NetProfit)
		}
	}
}

func TestProfitabilityCalculator_BreakEvenZeroFixedCosts(t *testing.T) {
	pc := newTestProfitabilityCalculator()
	resources := standardResources()
	costs := ProviderCosts{
		CPUCostPerCoreHour:    1e-12, // Very low variable costs
		MemoryCostPerGBHour:   1e-12,
		StorageCostPerGBMonth: 1e-11,
		FixedMonthlyCost:      0, // No fixed costs
	}

	breakEven := pc.BreakEvenAnalysis(resources, costs, 100_000)

	if breakEven != 0 {
		t.Errorf("with zero fixed costs and profitable per-job, break-even should be 0, got %d", breakEven)
	}
}

func TestProfitabilityCalculator_BreakEvenUnprofitablePerJob(t *testing.T) {
	pc := newTestProfitabilityCalculator()
	resources := standardResources()
	// Per-job costs exceed per-job revenue (~4e-9 BUNKER)
	costs := ProviderCosts{
		CPUCostPerCoreHour:    1.0, // Way above per-job revenue
		MemoryCostPerGBHour:   1.0,
		StorageCostPerGBMonth: 1.0,
		FixedMonthlyCost:      100.0,
	}

	breakEven := pc.BreakEvenAnalysis(resources, costs, 100_000)

	if breakEven != math.MaxInt {
		t.Errorf("should never break even when per-job costs exceed revenue, got %d", breakEven)
	}
}

func TestProfitabilityCalculator_TierComparison(t *testing.T) {
	pc := newTestProfitabilityCalculator()
	resources := standardResources()
	costs := zeroCosts() // Zero costs to isolate tier effects

	reports := pc.TierComparison(resources, costs, 100)

	if len(reports) != 5 {
		t.Fatalf("expected 5 tier reports, got %d", len(reports))
	}

	// Reports should be sorted by stake amount ascending
	for i := 1; i < len(reports); i++ {
		if reports[i].StakeAmount < reports[i-1].StakeAmount {
			t.Error("reports should be sorted by stake amount ascending")
			break
		}
	}

	// Higher tiers should have better gross revenue (due to reward multiplier)
	// Silver (1.05x), Gold (1.1x), Platinum (1.2x) > Starter/Bronze (1.0x)
	starterRevenue := reports[0].GrossRevenue
	platinumRevenue := reports[4].GrossRevenue

	if platinumRevenue <= starterRevenue {
		t.Errorf("platinum revenue (%e) should exceed starter revenue (%e) due to multiplier",
			platinumRevenue, starterRevenue)
	}

	// Verify tier names are present and in expected order
	expectedTiers := []string{"starter", "bronze", "silver", "gold", "platinum"}
	for i, expected := range expectedTiers {
		if reports[i].StakeTier != expected {
			t.Errorf("report[%d] tier: got %q, want %q", i, reports[i].StakeTier, expected)
		}
	}
}

func TestProfitabilityCalculator_HigherTierBetterProfitAtScale(t *testing.T) {
	pc := newTestProfitabilityCalculator()
	resources := standardResources()
	costs := zeroCosts()

	// At high volume, higher tiers should show more gross revenue per job
	// (due to reward multiplier), which translates to higher total profit
	// when costs are the same.
	reportStarter := pc.Calculate(resources, costs, 1_000_000, 1000)       // Starter (1.0x)
	reportPlatinum := pc.Calculate(resources, costs, 1_000_000_000, 1000) // Platinum (1.2x)

	// Platinum should have higher gross revenue (1.2x multiplier vs 1.0x)
	if reportPlatinum.GrossRevenue <= reportStarter.GrossRevenue {
		t.Errorf("platinum gross revenue (%e) should exceed starter (%e)",
			reportPlatinum.GrossRevenue, reportStarter.GrossRevenue)
	}

	// Platinum net revenue should be proportionally higher
	if reportPlatinum.NetRevenue <= reportStarter.NetRevenue {
		t.Errorf("platinum net revenue (%e) should exceed starter (%e)",
			reportPlatinum.NetRevenue, reportStarter.NetRevenue)
	}
}

func TestProfitabilityCalculator_BidDiscount(t *testing.T) {
	pc := newTestProfitabilityCalculator()
	resources := standardResources()
	costs := zeroCosts()

	// Zero stake = zero discount
	report0 := pc.Calculate(resources, costs, 0, 100)
	if report0.BidDiscount != 0 {
		t.Errorf("bid discount should be 0 with zero stake, got %f", report0.BidDiscount)
	}

	// Some stake = some discount
	report500 := pc.Calculate(resources, costs, 500, 100)
	if report500.BidDiscount <= 0 {
		t.Error("bid discount should be positive with non-zero stake")
	}

	// Higher stake = higher discount
	report10k := pc.Calculate(resources, costs, 10000, 100)
	if report10k.BidDiscount <= report500.BidDiscount {
		t.Errorf("higher stake should give higher discount: 10k=%f, 500=%f",
			report10k.BidDiscount, report500.BidDiscount)
	}

	// Discount should never reach 100%
	reportHuge := pc.Calculate(resources, costs, 1000000, 100)
	if reportHuge.BidDiscount >= 100 {
		t.Errorf("bid discount should never reach 100%%, got %f", reportHuge.BidDiscount)
	}
}

func TestProfitabilityCalculator_PaybackPeriod(t *testing.T) {
	pc := newTestProfitabilityCalculator()
	resources := standardResources()
	costs := zeroCosts() // Zero costs ensures profit is positive

	report := pc.Calculate(resources, costs, 500, 100)

	if report.NetProfit <= 0 {
		t.Fatal("test expects positive profit for payback calculation")
	}

	expectedPayback := 500 / report.NetProfit
	tolerance := 0.001

	if math.Abs(report.PaybackPeriod-expectedPayback) > tolerance {
		t.Errorf("payback period: got %f, want %f", report.PaybackPeriod, expectedPayback)
	}
}

func TestProfitabilityCalculator_PaybackPeriodUnprofitable(t *testing.T) {
	pc := newTestProfitabilityCalculator()
	resources := standardResources()
	// Costs that exceed revenue
	costs := ProviderCosts{
		CPUCostPerCoreHour:  1.0,
		MemoryCostPerGBHour: 1.0,
		FixedMonthlyCost:    100.0,
	}

	report := pc.Calculate(resources, costs, 500, 100)

	if !math.IsInf(report.PaybackPeriod, 1) {
		t.Errorf("payback period should be infinite when not profitable, got %f", report.PaybackPeriod)
	}
}

func TestProfitabilityCalculator_MonthlyROI(t *testing.T) {
	pc := newTestProfitabilityCalculator()
	resources := standardResources()
	costs := zeroCosts()

	report := pc.Calculate(resources, costs, 500, 100)

	expectedROI := (report.NetProfit / 500) * 100
	tolerance := 1e-10

	if math.Abs(report.MonthlyROI-expectedROI) > tolerance {
		t.Errorf("monthly ROI: got %e, want %e", report.MonthlyROI, expectedROI)
	}
}

func TestProfitabilityCalculator_ZeroStake(t *testing.T) {
	pc := newTestProfitabilityCalculator()
	resources := standardResources()
	costs := zeroCosts()

	report := pc.Calculate(resources, costs, 0, 100)

	// Below minimum tier
	if report.StakeTier != "" {
		t.Errorf("expected empty tier for zero stake, got %q", report.StakeTier)
	}
	if report.MonthlyROI != 0 {
		t.Errorf("monthly ROI should be 0 with zero stake (avoid div by zero), got %f", report.MonthlyROI)
	}
	if report.BidDiscount != 0 {
		t.Errorf("bid discount should be 0 with zero stake, got %f", report.BidDiscount)
	}
}

func TestProfitabilityCalculator_ProfitMarginCalculation(t *testing.T) {
	pc := newTestProfitabilityCalculator()
	resources := standardResources()

	// With zero costs, profit margin should be 100%
	report := pc.Calculate(resources, zeroCosts(), 500, 100)
	if report.ProfitMargin != 100 {
		t.Errorf("profit margin with zero costs should be 100%%, got %f", report.ProfitMargin)
	}

	// With some costs, margin should be less than 100
	reportWithCosts := pc.Calculate(resources, tinyCosts(), 500, 100)
	if reportWithCosts.ProfitMargin >= 100 {
		t.Errorf("profit margin with some costs should be less than 100, got %f",
			reportWithCosts.ProfitMargin)
	}
}
