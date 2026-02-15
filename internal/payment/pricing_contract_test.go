package payment

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func TestPricingContract_MockCalculateCost(t *testing.T) {
	pc := NewMockOnChainPricingContract()
	ctx := context.Background()

	req := ResourceRequestData{
		CPUCores:      4,
		MemoryGB:      8,
		StorageGB:     100,
		BandwidthMbps: 10,
		DurationHours: 24,
		GPUCount:      0,
	}

	cost, err := pc.CalculateCost(ctx, req)
	if err != nil {
		t.Fatalf("failed to calculate cost: %v", err)
	}
	if cost.Sign() <= 0 {
		t.Error("expected positive cost for non-zero resources")
	}

	// CPU: 4 cores × 1 BUNKER × 24h = 96e18
	// Memory: 8 GB × 0.5 BUNKER × 24h = 96e18
	// Storage: 100 GB × 0.1 BUNKER × 24h = 240e18
	// Total = 432e18
	expected := new(big.Int).Mul(big.NewInt(432), parseWei("1"))
	if cost.Cmp(expected) != 0 {
		t.Errorf("expected cost %s, got %s", expected.String(), cost.String())
	}
}

func TestPricingContract_MockCalculateCostWithGPU(t *testing.T) {
	pc := NewMockOnChainPricingContract()
	ctx := context.Background()

	req := ResourceRequestData{
		CPUCores:      2,
		MemoryGB:      4,
		StorageGB:     0,
		BandwidthMbps: 0,
		DurationHours: 1,
		GPUCount:      2,
	}

	cost, err := pc.CalculateCost(ctx, req)
	if err != nil {
		t.Fatalf("failed to calculate cost: %v", err)
	}

	// CPU: 2 × 1 × 1 = 2e18
	// Memory: 4 × 0.5 × 1 = 2e18
	// GPU: 2 × 10 × 1 = 20e18
	// Total = 24e18
	expected := new(big.Int).Mul(big.NewInt(24), parseWei("1"))
	if cost.Cmp(expected) != 0 {
		t.Errorf("expected cost %s, got %s", expected.String(), cost.String())
	}
}

func TestPricingContract_MockGetTokenPrice(t *testing.T) {
	pc := NewMockOnChainPricingContract()
	ctx := context.Background()

	price, err := pc.GetTokenPrice(ctx)
	if err != nil {
		t.Fatalf("failed to get token price: %v", err)
	}

	// $0.001 × 1e8 = 100000
	expected := big.NewInt(100000)
	if price.Cmp(expected) != 0 {
		t.Errorf("expected token price %s, got %s", expected.String(), price.String())
	}
}

func TestPricingContract_MockGetPrices(t *testing.T) {
	pc := NewMockOnChainPricingContract()
	ctx := context.Background()

	prices, err := pc.GetPrices(ctx)
	if err != nil {
		t.Fatalf("failed to get prices: %v", err)
	}

	if prices.CPUPerCoreHour.Sign() <= 0 {
		t.Error("expected positive CPU price")
	}
	if prices.MemoryPerGBHour.Sign() <= 0 {
		t.Error("expected positive memory price")
	}
	if prices.GPUPerHour.Sign() <= 0 {
		t.Error("expected positive GPU price")
	}
}

func TestPricingContract_MockGetMultipliers(t *testing.T) {
	pc := NewMockOnChainPricingContract()
	ctx := context.Background()

	mult, err := pc.GetMultipliers(ctx)
	if err != nil {
		t.Fatalf("failed to get multipliers: %v", err)
	}

	// Default 1.0x = 10000 basis points
	expected := big.NewInt(10000)
	if mult.DemandMultiplier.Cmp(expected) != 0 {
		t.Errorf("expected demand multiplier %s, got %s", expected.String(), mult.DemandMultiplier.String())
	}
	if mult.RegionMultiplier.Cmp(expected) != 0 {
		t.Errorf("expected region multiplier %s, got %s", expected.String(), mult.RegionMultiplier.String())
	}
}

func TestPricingContract_MockGetProviderPrices(t *testing.T) {
	pc := NewMockOnChainPricingContract()
	ctx := context.Background()

	provider := common.HexToAddress("0x1111111111111111111111111111111111111111")

	pp, err := pc.GetProviderPrices(ctx, provider)
	if err != nil {
		t.Fatalf("failed to get provider prices: %v", err)
	}

	// Default 1x multipliers for unknown provider
	expected := big.NewInt(10000)
	if pp.CPUMultiplier.Cmp(expected) != 0 {
		t.Errorf("expected CPU multiplier %s, got %s", expected.String(), pp.CPUMultiplier.String())
	}
}

func TestPricingContract_MockProviderCost(t *testing.T) {
	pc := NewMockOnChainPricingContract()
	ctx := context.Background()

	provider := common.HexToAddress("0x1111111111111111111111111111111111111111")
	req := ResourceRequestData{
		CPUCores:      1,
		MemoryGB:      1,
		DurationHours: 1,
	}

	cost, err := pc.CalculateProviderCost(ctx, provider, req)
	if err != nil {
		t.Fatalf("failed to calculate provider cost: %v", err)
	}

	// Same as base cost in mock mode
	baseCost, _ := pc.CalculateCost(ctx, req)
	if cost.Cmp(baseCost) != 0 {
		t.Errorf("expected provider cost %s to equal base cost %s in mock", cost.String(), baseCost.String())
	}
}

func TestPricingContract_MockMode(t *testing.T) {
	pc := NewMockOnChainPricingContract()
	if !pc.IsMockMode() {
		t.Error("expected mock mode")
	}
}

func TestPricingContract_ZeroResources(t *testing.T) {
	pc := NewMockOnChainPricingContract()
	ctx := context.Background()

	req := ResourceRequestData{
		DurationHours: 24,
	}

	cost, err := pc.CalculateCost(ctx, req)
	if err != nil {
		t.Fatalf("failed to calculate cost: %v", err)
	}
	if cost.Sign() != 0 {
		t.Errorf("expected zero cost for zero resources, got %s", cost.String())
	}
}
