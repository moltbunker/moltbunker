package payment

import (
	"math/big"
	"time"

	"github.com/moltbunker/moltbunker/pkg/types"
)

// PricingCalculator calculates dynamic pricing
type PricingCalculator struct {
	basePricePerHour *big.Int
}

// NewPricingCalculator creates a new pricing calculator
func NewPricingCalculator(basePricePerHour *big.Int) *PricingCalculator {
	return &PricingCalculator{
		basePricePerHour: basePricePerHour,
	}
}

// CalculatePrice calculates price based on resources and duration
func (pc *PricingCalculator) CalculatePrice(resources types.ResourceLimits, duration time.Duration) *big.Int {
	// Base calculation: CPU + Memory + Disk + Network
	cpuPrice := new(big.Int).Mul(
		big.NewInt(resources.CPUQuota),
		big.NewInt(1000), // Price per CPU unit
	)

	memoryPrice := new(big.Int).Div(
		big.NewInt(resources.MemoryLimit),
		big.NewInt(1024*1024*1024), // Price per GB
	)

	diskPrice := new(big.Int).Div(
		big.NewInt(resources.DiskLimit),
		big.NewInt(1024*1024*1024), // Price per GB
	)

	networkPrice := new(big.Int).Div(
		big.NewInt(resources.NetworkBW),
		big.NewInt(1024*1024), // Price per MB/s
	)

	// Sum all components
	totalPrice := new(big.Int).Add(cpuPrice, memoryPrice)
	totalPrice.Add(totalPrice, diskPrice)
	totalPrice.Add(totalPrice, networkPrice)

	// Multiply by duration (in hours)
	hours := duration.Hours()
	hoursInt := big.NewInt(int64(hours))
	totalPrice.Mul(totalPrice, hoursInt)

	return totalPrice
}

// CalculateMoltInvocationPrice calculates the cost of a single Molt invocation.
// Pricing is per-invocation based on wall-clock execution time and memory used.
// A minimum billing floor (MinimumFunctionMillis, default 100ms) prevents
// micro-invocation abuse. Memory is billed per MB-second.
//
// Formula:
//   cost = (cpuRate * billedMs / 3_600_000) + (memRate * memMB * billedMs / 3_600_000)
//
// where billedMs = max(actualMs, minimumMs) and rates are per-hour from PricingConfig.
func (pc *PricingCalculator) CalculateMoltInvocationPrice(duration time.Duration, memoryUsedBytes int64, pricingCfg *types.PricingConfig) *big.Int {
	// Apply minimum billing floor
	minimumMs := int64(100)
	if pricingCfg != nil && pricingCfg.MinimumFunctionMillis > 0 {
		minimumMs = int64(pricingCfg.MinimumFunctionMillis)
	}

	billedMs := duration.Milliseconds()
	if billedMs < minimumMs {
		billedMs = minimumMs
	}

	// Parse per-hour rates from pricing config (BUNKER wei strings)
	cpuRate := parsePricingRate(pricingCfg, "cpu")
	memRate := parsePricingRate(pricingCfg, "memory")

	// CPU component: (cpuRate * billedMs) / 3_600_000
	// This gives the fraction of one CPU core-hour used by the invocation.
	// Molt functions use a single core, so no core-count multiplier.
	msPerHour := big.NewInt(3_600_000)
	cpuCost := new(big.Int).Mul(cpuRate, big.NewInt(billedMs))
	cpuCost.Div(cpuCost, msPerHour)

	// Memory component: (memRate * memMB * billedMs) / 3_600_000
	// Convert bytes to MB (round up to nearest MB for fairness).
	memMB := (memoryUsedBytes + 1024*1024 - 1) / (1024 * 1024)
	if memMB < 1 {
		memMB = 1 // Minimum 1 MB billed
	}
	memCost := new(big.Int).Mul(memRate, big.NewInt(memMB))
	memCost.Mul(memCost, big.NewInt(billedMs))
	memCost.Div(memCost, msPerHour)

	// Total = CPU + Memory
	total := new(big.Int).Add(cpuCost, memCost)

	// Ensure at least 1 wei for any valid invocation
	if total.Sign() <= 0 {
		return big.NewInt(1)
	}

	return total
}

// parsePricingRate extracts a per-hour rate from PricingConfig.
func parsePricingRate(cfg *types.PricingConfig, resource string) *big.Int {
	if cfg == nil {
		return defaultRate(resource)
	}
	var rateStr string
	switch resource {
	case "cpu":
		rateStr = cfg.CPUPerCoreHour
	case "memory":
		rateStr = cfg.MemoryPerGBHour
	default:
		return big.NewInt(0)
	}
	rate, ok := new(big.Int).SetString(rateStr, 10)
	if !ok || rate.Sign() <= 0 {
		return defaultRate(resource)
	}
	return rate
}

// defaultRate returns fallback per-hour rates (from DefaultPricingConfig).
func defaultRate(resource string) *big.Int {
	switch resource {
	case "cpu":
		rate, _ := new(big.Int).SetString("500000000000000000", 10) // 0.5 BUNKER
		return rate
	case "memory":
		rate, _ := new(big.Int).SetString("100000000000000000", 10) // 0.1 BUNKER
		return rate
	default:
		return big.NewInt(0)
	}
}

// CalculateBid calculates a bid price for hosting
func (pc *PricingCalculator) CalculateBid(resources types.ResourceLimits, duration time.Duration, stake *big.Int) *big.Int {
	basePrice := pc.CalculatePrice(resources, duration)

	// Normalize stake to BUNKER units (1 BUNKER = 10^18 wei)
	// Higher stake = lower bid (discount for more committed providers)
	// Formula: bid = basePrice * normalizer / (normalizer + stake)
	// This gives a discount that approaches 50% as stake approaches normalizer
	normalizer := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil) // 1 BUNKER in wei

	numerator := new(big.Int).Mul(basePrice, normalizer)
	denominator := new(big.Int).Add(normalizer, stake)

	adjustedPrice := new(big.Int).Div(numerator, denominator)

	// Ensure bid is at least 1 (never zero for valid resources)
	if adjustedPrice.Sign() <= 0 && basePrice.Sign() > 0 {
		return big.NewInt(1)
	}

	return adjustedPrice
}
