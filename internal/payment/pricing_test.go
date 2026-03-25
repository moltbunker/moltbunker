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
		CPUQuota:    1000000,     // 1 CPU
		CPUPeriod:   100000,      // 100ms period
		MemoryLimit: 1073741824,  // 1GB
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
	stake1 := big.NewInt(1000000000000000000) // 1 BUNKER
	stake2 := big.NewInt(2000000000000000000) // 2 BUNKER

	bid1 := pc.CalculateBid(resources, duration, stake1)
	bid2 := pc.CalculateBid(resources, duration, stake2)

	// Higher stake should result in lower bid (better price)
	if bid2.Cmp(bid1) >= 0 {
		t.Error("Higher stake should result in lower bid")
	}
}

// --- Molt Invocation Pricing Tests ---

func TestCalculateMoltInvocationPrice_MinimumFloor(t *testing.T) {
	pc := NewPricingCalculator(big.NewInt(1))
	cfg := types.DefaultPricingConfig()

	// 10ms invocation should be billed as 100ms (minimum floor)
	cost := pc.CalculateMoltInvocationPrice(10*time.Millisecond, 1*1024*1024, cfg)
	if cost.Sign() <= 0 {
		t.Fatal("cost should be positive")
	}

	// 100ms invocation should produce the same cost (exactly at floor)
	costAtFloor := pc.CalculateMoltInvocationPrice(100*time.Millisecond, 1*1024*1024, cfg)
	if cost.Cmp(costAtFloor) != 0 {
		t.Fatalf("10ms and 100ms should cost the same (floor), got %s vs %s", cost, costAtFloor)
	}
}

func TestCalculateMoltInvocationPrice_LongerCostsMore(t *testing.T) {
	pc := NewPricingCalculator(big.NewInt(1))
	cfg := types.DefaultPricingConfig()

	cost100ms := pc.CalculateMoltInvocationPrice(100*time.Millisecond, 64*1024*1024, cfg)
	cost1s := pc.CalculateMoltInvocationPrice(1*time.Second, 64*1024*1024, cfg)
	cost10s := pc.CalculateMoltInvocationPrice(10*time.Second, 64*1024*1024, cfg)

	if cost1s.Cmp(cost100ms) <= 0 {
		t.Errorf("1s should cost more than 100ms: %s vs %s", cost1s, cost100ms)
	}
	if cost10s.Cmp(cost1s) <= 0 {
		t.Errorf("10s should cost more than 1s: %s vs %s", cost10s, cost1s)
	}
}

func TestCalculateMoltInvocationPrice_MoreMemoryCostsMore(t *testing.T) {
	pc := NewPricingCalculator(big.NewInt(1))
	cfg := types.DefaultPricingConfig()

	cost1MB := pc.CalculateMoltInvocationPrice(1*time.Second, 1*1024*1024, cfg)
	cost256MB := pc.CalculateMoltInvocationPrice(1*time.Second, 256*1024*1024, cfg)

	if cost256MB.Cmp(cost1MB) <= 0 {
		t.Errorf("256MB should cost more than 1MB: %s vs %s", cost256MB, cost1MB)
	}
}

func TestCalculateMoltInvocationPrice_NilConfig(t *testing.T) {
	pc := NewPricingCalculator(big.NewInt(1))

	// Should use defaults, not panic
	cost := pc.CalculateMoltInvocationPrice(500*time.Millisecond, 64*1024*1024, nil)
	if cost.Sign() <= 0 {
		t.Fatal("cost should be positive with nil config")
	}
}

func TestCalculateMoltInvocationPrice_ZeroMemory(t *testing.T) {
	pc := NewPricingCalculator(big.NewInt(1))
	cfg := types.DefaultPricingConfig()

	// 0 bytes memory should be billed as 1MB minimum
	cost := pc.CalculateMoltInvocationPrice(1*time.Second, 0, cfg)
	if cost.Sign() <= 0 {
		t.Fatal("cost should be positive even with 0 memory")
	}
}

func TestCalculateMoltInvocationPrice_ProportionalToTime(t *testing.T) {
	pc := NewPricingCalculator(big.NewInt(1))
	cfg := types.DefaultPricingConfig()

	// 100ms invocation should cost roughly 1/36000th of a 1-hour invocation
	// (100ms / 3,600,000ms = 1/36000). Verify proportionality.
	cost100ms := pc.CalculateMoltInvocationPrice(100*time.Millisecond, 64*1024*1024, cfg)
	cost1hr := pc.CalculateMoltInvocationPrice(1*time.Hour, 64*1024*1024, cfg)

	// 1hr / 100ms = 36000x ratio. Allow 2x margin for rounding.
	ratio := new(big.Int).Div(cost1hr, cost100ms)
	if ratio.Int64() < 18000 || ratio.Int64() > 72000 {
		t.Errorf("1hr/100ms cost ratio = %d, expected ~36000", ratio.Int64())
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
