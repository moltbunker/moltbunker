package payment

import (
	"math/big"
	"testing"
	"time"

	"github.com/moltbunker/moltbunker/pkg/types"
)

func TestPricingCalculator_CalculatePrice(t *testing.T) {
	basePrice := big.NewInt(1000000000000000) // 0.001 BUNKER per hour
	pc := NewPricingCalculator(basePrice)

	resources := types.ResourceLimits{
		CPUQuota:    1000000,  // 1 CPU
		CPUPeriod:   100000,   // 100ms period
		MemoryLimit: 1073741824, // 1GB
		DiskLimit:   10737418240, // 10GB
		NetworkBW:   10485760,    // 10MB/s
		PIDLimit:    100,
	}

	duration := 1 * time.Hour

	price := pc.CalculatePrice(resources, duration)
	if price == nil {
		t.Fatal("Price should not be nil")
	}

	if price.Sign() <= 0 {
		t.Error("Price should be positive")
	}
}

func TestPricingCalculator_CalculateBid(t *testing.T) {
	basePrice := big.NewInt(1000000000000000)
	pc := NewPricingCalculator(basePrice)

	resources := types.ResourceLimits{
		CPUQuota:    1000000,
		CPUPeriod:   100000,
		MemoryLimit: 1073741824,
		DiskLimit:   10737418240,
		NetworkBW:   10485760,
		PIDLimit:    100,
	}

	duration := 1 * time.Hour
	stake := big.NewInt(1000000000000000000) // 1 BUNKER token

	bid := pc.CalculateBid(resources, duration, stake)
	if bid == nil {
		t.Fatal("Bid should not be nil")
	}

	if bid.Sign() <= 0 {
		t.Error("Bid should be positive")
	}
}

func TestPricingCalculator_CalculateBid_HigherStake(t *testing.T) {
	basePrice := big.NewInt(1000000000000000)
	pc := NewPricingCalculator(basePrice)

	resources := types.ResourceLimits{
		CPUQuota:    1000000,
		CPUPeriod:   100000,
		MemoryLimit: 1073741824,
		DiskLimit:   10737418240,
		NetworkBW:   10485760,
		PIDLimit:    100,
	}

	duration := 1 * time.Hour
	stake1 := big.NewInt(1000000000000000000)  // 1 BUNKER
	stake2 := big.NewInt(2000000000000000000)  // 2 BUNKER

	bid1 := pc.CalculateBid(resources, duration, stake1)
	bid2 := pc.CalculateBid(resources, duration, stake2)

	// Higher stake should result in lower bid (better price)
	if bid2.Cmp(bid1) >= 0 {
		t.Error("Higher stake should result in lower bid")
	}
}

func TestPricingCalculator_CalculatePrice_DifferentDurations(t *testing.T) {
	basePrice := big.NewInt(1000000000000000)
	pc := NewPricingCalculator(basePrice)

	resources := types.ResourceLimits{
		CPUQuota:    1000000,
		CPUPeriod:   100000,
		MemoryLimit: 1073741824,
		DiskLimit:   10737418240,
		NetworkBW:   10485760,
		PIDLimit:    100,
	}

	price1Hour := pc.CalculatePrice(resources, 1*time.Hour)
	price2Hours := pc.CalculatePrice(resources, 2*time.Hour)

	if price2Hours.Cmp(price1Hour) <= 0 {
		t.Error("2 hours should cost more than 1 hour")
	}
}
