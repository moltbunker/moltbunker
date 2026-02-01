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
