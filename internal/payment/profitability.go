package payment

import (
	"math"
	"math/big"
	"sort"
	"time"

	"github.com/moltbunker/moltbunker/pkg/types"
)

// ProfitabilityCalculator estimates provider profitability across staking tiers
type ProfitabilityCalculator struct {
	pricing *PricingCalculator
}

// ResourceSpec describes the resources a provider offers for a single job
type ResourceSpec struct {
	CPUCores  float64 // Number of CPU cores
	MemoryGB  float64 // Memory in GB
	StorageGB float64 // Storage in GB
}

// ProviderCosts describes a provider's real-world infrastructure costs
type ProviderCosts struct {
	CPUCostPerCoreHour    float64 // Provider's actual CPU cost ($/core-hour)
	MemoryCostPerGBHour   float64 // Provider's actual memory cost ($/GB-hour)
	StorageCostPerGBMonth float64 // Provider's actual storage cost ($/GB-month)
	NetworkCostPerGB      float64 // Provider's actual bandwidth cost ($/GB)
	FixedMonthlyCost      float64 // Fixed infrastructure costs (server rental, etc.)
}

// ProfitabilityReport contains the full profitability analysis for a provider
type ProfitabilityReport struct {
	// Revenue
	GrossRevenue float64 // Total revenue from jobs (BUNKER)
	ProtocolFees float64 // 5% fee deducted (BUNKER)
	NetRevenue   float64 // After fees (BUNKER)

	// Costs
	InfrastructureCost float64 // Variable compute costs
	FixedCost          float64 // Fixed monthly costs
	StakingCost        float64 // Opportunity cost of staked tokens (assumed 0 for simplicity)
	TotalCost          float64

	// Profit
	NetProfit      float64 // Net revenue minus total costs
	ProfitMargin   float64 // Percentage (0-100)
	MonthlyROI     float64 // Return on staked investment percentage
	BreakEvenJobs  int     // Jobs needed per month to break even
	PaybackPeriod  float64 // Months to recover stake from profit

	// Staking
	StakeTier        string  // Tier name
	StakeAmount      float64 // BUNKER tokens staked
	RewardMultiplier float64 // Tier-based multiplier
	BidDiscount      float64 // Price discount percentage from staking
}

// NewProfitabilityCalculator creates a new profitability calculator
func NewProfitabilityCalculator(pricing *PricingCalculator) *ProfitabilityCalculator {
	return &ProfitabilityCalculator{
		pricing: pricing,
	}
}

// Calculate computes a full profitability report for the given parameters.
// resources describes the per-job resource allocation, costs are the provider's
// real-world infrastructure costs, stakeAmount is BUNKER tokens staked, and
// jobsPerMonth is the expected number of 1-hour jobs per month.
func (pc *ProfitabilityCalculator) Calculate(resources ResourceSpec, costs ProviderCosts, stakeAmount float64, jobsPerMonth int) *ProfitabilityReport {
	report := &ProfitabilityReport{}

	// Determine staking tier and multiplier
	tier, tierCfg := pc.resolveTier(stakeAmount)
	report.StakeTier = string(tier)
	report.StakeAmount = stakeAmount
	report.RewardMultiplier = tierCfg.RewardMultiplier

	// Calculate per-job revenue using the existing pricing calculator
	perJobRevenue := pc.perJobRevenue(resources)

	// Apply reward multiplier from staking tier
	perJobRevenue *= tierCfg.RewardMultiplier

	// Calculate bid discount: formula from pricing.go
	// adjustedPrice = basePrice * normalizer / (normalizer + stake)
	// discount = 1 - adjustedPrice/basePrice = stake / (normalizer + stake)
	stakeWei := bunkerToWei(stakeAmount)
	normalizerFloat := 1e18 // 1 BUNKER in wei
	bidDiscount := 0.0
	if stakeWei > 0 {
		bidDiscount = stakeWei / (normalizerFloat + stakeWei) * 100
	}
	report.BidDiscount = bidDiscount

	// Revenue calculation
	protocolFees := types.DefaultProtocolFees()
	feeRate := protocolFees.TotalFeePercent / 100.0

	report.GrossRevenue = perJobRevenue * float64(jobsPerMonth)
	report.ProtocolFees = report.GrossRevenue * feeRate
	report.NetRevenue = report.GrossRevenue - report.ProtocolFees

	// Cost calculation
	report.InfrastructureCost = pc.calculateInfraCost(resources, costs, jobsPerMonth)
	report.FixedCost = costs.FixedMonthlyCost
	report.StakingCost = 0 // Opportunity cost is hard to estimate; kept at 0
	report.TotalCost = report.InfrastructureCost + report.FixedCost + report.StakingCost

	// Profit
	report.NetProfit = report.NetRevenue - report.TotalCost
	if report.NetRevenue > 0 {
		report.ProfitMargin = (report.NetProfit / report.NetRevenue) * 100
	}
	if stakeAmount > 0 {
		report.MonthlyROI = (report.NetProfit / stakeAmount) * 100
	}

	// Break-even analysis
	report.BreakEvenJobs = pc.breakEvenJobs(resources, costs, stakeAmount)

	// Payback period: months to recover the staked amount from profit
	if report.NetProfit > 0 {
		report.PaybackPeriod = stakeAmount / report.NetProfit
	} else {
		report.PaybackPeriod = math.Inf(1) // Never recoverable if not profitable
	}

	return report
}

// BreakEvenAnalysis returns the minimum number of 1-hour jobs per month
// needed to cover all costs (infrastructure + fixed + protocol fees).
func (pc *ProfitabilityCalculator) BreakEvenAnalysis(resources ResourceSpec, costs ProviderCosts, stakeAmount float64) int {
	return pc.breakEvenJobs(resources, costs, stakeAmount)
}

// TierComparison returns profitability reports for all staking tiers at the
// given job volume, sorted from lowest to highest tier.
func (pc *ProfitabilityCalculator) TierComparison(resources ResourceSpec, costs ProviderCosts, jobsPerMonth int) []ProfitabilityReport {
	tierOrder := []types.StakingTier{
		types.StakingTierStarter,
		types.StakingTierBronze,
		types.StakingTierSilver,
		types.StakingTierGold,
		types.StakingTierPlatinum,
	}

	tiers := types.DefaultStakingTiers()
	for _, cfg := range tiers {
		_ = cfg.ParseMinStake()
	}

	reports := make([]ProfitabilityReport, 0, len(tierOrder))
	for _, tier := range tierOrder {
		cfg := tiers[tier]
		if cfg == nil || cfg.MinStake == nil {
			continue
		}

		// Use the minimum stake for each tier
		stakeFloat := weiToBunker(cfg.MinStake)
		report := pc.Calculate(resources, costs, stakeFloat, jobsPerMonth)
		reports = append(reports, *report)
	}

	// Sort by tier stake amount ascending (already ordered, but be explicit)
	sort.Slice(reports, func(i, j int) bool {
		return reports[i].StakeAmount < reports[j].StakeAmount
	})

	return reports
}

// perJobRevenue calculates the BUNKER revenue for a single 1-hour job using
// the existing PricingCalculator. It converts the ResourceSpec to ResourceLimits
// and calls CalculatePrice.
func (pc *ProfitabilityCalculator) perJobRevenue(resources ResourceSpec) float64 {
	rl := specToResourceLimits(resources)
	priceWei := pc.pricing.CalculatePrice(rl, 1*time.Hour)
	return weiBigIntToBunker(priceWei)
}

// calculateInfraCost calculates the provider's variable infrastructure cost for
// the given number of 1-hour jobs per month.
func (pc *ProfitabilityCalculator) calculateInfraCost(resources ResourceSpec, costs ProviderCosts, jobsPerMonth int) float64 {
	perJobCost := (resources.CPUCores * costs.CPUCostPerCoreHour) +
		(resources.MemoryGB * costs.MemoryCostPerGBHour) +
		(resources.StorageGB * costs.StorageCostPerGBMonth / 730) // ~730 hours/month, prorate to per-hour
	return perJobCost * float64(jobsPerMonth)
}

// breakEvenJobs calculates the minimum number of jobs per month to break even.
func (pc *ProfitabilityCalculator) breakEvenJobs(resources ResourceSpec, costs ProviderCosts, stakeAmount float64) int {
	_, tierCfg := pc.resolveTier(stakeAmount)

	perJobRevenue := pc.perJobRevenue(resources) * tierCfg.RewardMultiplier
	protocolFees := types.DefaultProtocolFees()
	feeRate := protocolFees.TotalFeePercent / 100.0
	netPerJob := perJobRevenue * (1 - feeRate)

	perJobCost := (resources.CPUCores * costs.CPUCostPerCoreHour) +
		(resources.MemoryGB * costs.MemoryCostPerGBHour) +
		(resources.StorageGB * costs.StorageCostPerGBMonth / 730)

	profitPerJob := netPerJob - perJobCost
	if profitPerJob <= 0 {
		// Each job is a loss; can never break even on variable costs alone
		return math.MaxInt
	}

	// Need: profitPerJob * N >= fixedCost
	if costs.FixedMonthlyCost <= 0 {
		if profitPerJob > 0 {
			return 0 // No fixed costs and each job is profitable
		}
		return math.MaxInt
	}

	n := int(math.Ceil(costs.FixedMonthlyCost / profitPerJob))
	return n
}

// resolveTier determines the staking tier and its config for a given BUNKER amount.
// It uses big.Int arithmetic to avoid float64 precision loss when converting
// whole-number BUNKER amounts to wei.
func (pc *ProfitabilityCalculator) resolveTier(stakeAmount float64) (types.StakingTier, *types.StakingTierConfig) {
	// Convert BUNKER to wei using big.Int to avoid float64 precision loss.
	// Split into integer and fractional parts for exact conversion.
	wholeBunker := int64(stakeAmount)
	fracBunker := stakeAmount - float64(wholeBunker)

	weiPerBunker := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	stakeWei := new(big.Int).Mul(big.NewInt(wholeBunker), weiPerBunker)

	// Add fractional part (less precise but only matters for sub-token amounts)
	if fracBunker > 0 {
		fracWei := new(big.Int).SetInt64(int64(fracBunker * 1e18))
		stakeWei.Add(stakeWei, fracWei)
	}

	tiers := types.DefaultStakingTiers()
	for _, cfg := range tiers {
		_ = cfg.ParseMinStake()
	}

	tierOrder := []types.StakingTier{
		types.StakingTierPlatinum,
		types.StakingTierGold,
		types.StakingTierSilver,
		types.StakingTierBronze,
		types.StakingTierStarter,
	}

	for _, tier := range tierOrder {
		cfg := tiers[tier]
		if cfg.MinStake != nil && stakeWei.Cmp(cfg.MinStake) >= 0 {
			return tier, cfg
		}
	}

	// Below minimum stake; return starter config with 1.0 multiplier as fallback
	return "", &types.StakingTierConfig{
		RewardMultiplier: 1.0,
	}
}

// specToResourceLimits converts a ResourceSpec to types.ResourceLimits for use
// with the existing PricingCalculator.
func specToResourceLimits(spec ResourceSpec) types.ResourceLimits {
	return types.ResourceLimits{
		CPUQuota:    int64(spec.CPUCores * 1000000), // 1 core = 1,000,000 microseconds
		CPUPeriod:   100000,                         // Standard 100ms period
		MemoryLimit: int64(spec.MemoryGB * 1024 * 1024 * 1024),
		DiskLimit:   int64(spec.StorageGB * 1024 * 1024 * 1024),
		NetworkBW:   10485760, // Default 10MB/s
		PIDLimit:    100,
	}
}

// bunkerToWei converts BUNKER token amount to wei (float64 approximation).
func bunkerToWei(bunker float64) float64 {
	return bunker * 1e18
}

// weiToBunker converts a big.Int wei amount to BUNKER float64.
func weiToBunker(wei *big.Int) float64 {
	f := new(big.Float).SetInt(wei)
	divisor := new(big.Float).SetFloat64(1e18)
	result, _ := new(big.Float).Quo(f, divisor).Float64()
	return result
}

// weiBigIntToBunker converts a big.Int price in internal pricing units to
// a BUNKER float64. The PricingCalculator uses raw integer units, so we
// treat the result as-is (not wei-scaled) for estimation purposes.
func weiBigIntToBunker(wei *big.Int) float64 {
	f := new(big.Float).SetInt(wei)
	divisor := new(big.Float).SetFloat64(1e18)
	result, _ := new(big.Float).Quo(f, divisor).Float64()
	return result
}

