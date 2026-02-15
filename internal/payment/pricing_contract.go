package payment

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/moltbunker/moltbunker/internal/logging"
)

// OnChainPricingContract provides read-only interface to BunkerPricing contract.
// This reads on-chain oracle pricing data, distinct from the local PricingCalculator.
type OnChainPricingContract struct {
	baseClient   *BaseClient
	contract     *bind.BoundContract
	contractABI  abi.ABI
	contractAddr common.Address
	mockMode     bool

	// Mock state
	mockPrices          *ResourcePricesData
	mockMultipliers     *MultipliersData
	mockTokenPrice      *big.Int
	mockProviderPricing map[common.Address]*ProviderPricingData
	mockMu              sync.RWMutex
}

// NewOnChainPricingContract creates a new on-chain pricing contract client.
func NewOnChainPricingContract(baseClient *BaseClient, contractAddr common.Address) (*OnChainPricingContract, error) {
	pc := &OnChainPricingContract{
		baseClient:          baseClient,
		contractAddr:        contractAddr,
		mockProviderPricing: make(map[common.Address]*ProviderPricingData),
	}

	// Require a connected base client; use NewMockOnChainPricingContract() for testing
	if baseClient == nil {
		return nil, fmt.Errorf("base client is required (use NewMockOnChainPricingContract for testing)")
	}
	if !baseClient.IsConnected() {
		return nil, fmt.Errorf("base client not connected to RPC")
	}

	parsedABI, err := abi.JSON(strings.NewReader(PricingContractABI))
	if err != nil {
		return nil, fmt.Errorf("failed to parse pricing ABI: %w", err)
	}
	pc.contractABI = parsedABI

	client := baseClient.Client()
	pc.contract = bind.NewBoundContract(contractAddr, parsedABI, client, client, client)

	return pc, nil
}

// NewMockOnChainPricingContract creates a mock pricing contract for testing.
func NewMockOnChainPricingContract() *OnChainPricingContract {
	pc := &OnChainPricingContract{
		mockMode:            true,
		mockProviderPricing: make(map[common.Address]*ProviderPricingData),
	}
	pc.setupMockDefaults()
	return pc
}

func (pc *OnChainPricingContract) setupMockDefaults() {
	// Default resource prices (in token wei per unit-hour)
	pc.mockPrices = &ResourcePricesData{
		CPUPerCoreHour:     parseWei("1"),  // 1 BUNKER per core-hour
		MemoryPerGBHour:    parseWei("0"),  // ~0.5 BUNKER per GB-hour (using parseWei for simplicity)
		StoragePerGBHour:   parseWei("0"),  // ~0.1 BUNKER per GB-hour
		BandwidthPerGBHour: parseWei("0"),  // ~0.01 BUNKER per GB-hour
		GPUPerHour:         parseWei("10"), // 10 BUNKER per GPU-hour
	}
	// Set non-zero values with big.Int math for fractional amounts
	pc.mockPrices.MemoryPerGBHour = new(big.Int).Div(parseWei("1"), big.NewInt(2))   // 0.5
	pc.mockPrices.StoragePerGBHour = new(big.Int).Div(parseWei("1"), big.NewInt(10)) // 0.1
	pc.mockPrices.BandwidthPerGBHour = new(big.Int).Div(parseWei("1"), big.NewInt(100))

	pc.mockMultipliers = &MultipliersData{
		DemandMultiplier: big.NewInt(10000), // 1.0x (basis points: 10000 = 1x)
		RegionMultiplier: big.NewInt(10000),
		TierDiscount:     big.NewInt(0), // no discount
	}

	// Token price: 1 BUNKER = $0.001 USD (represented as 8-decimal Chainlink format)
	pc.mockTokenPrice = big.NewInt(100000) // $0.001 * 1e8
}

// IsMockMode returns whether running in mock mode.
func (pc *OnChainPricingContract) IsMockMode() bool {
	return pc.mockMode
}

// CalculateCost calculates the cost for a resource request at base prices.
func (pc *OnChainPricingContract) CalculateCost(ctx context.Context, req ResourceRequestData) (*big.Int, error) {
	if pc.mockMode {
		return pc.mockCalculateCost(req), nil
	}

	// Build the ABI tuple struct
	abiReq := struct {
		CpuCores      uint16
		MemoryGB      uint16
		StorageGB     uint16
		BandwidthMbps uint16
		DurationHours uint32
		GpuCount      uint8
	}{
		CpuCores:      req.CPUCores,
		MemoryGB:      req.MemoryGB,
		StorageGB:     req.StorageGB,
		BandwidthMbps: req.BandwidthMbps,
		DurationHours: req.DurationHours,
		GpuCount:      req.GPUCount,
	}

	var result []interface{}
	err := pc.contract.Call(&bind.CallOpts{Context: ctx}, &result, "calculateCost", abiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate cost: %w", err)
	}

	if len(result) > 0 {
		if cost, ok := result[0].(*big.Int); ok {
			return cost, nil
		}
	}
	return big.NewInt(0), nil
}

func (pc *OnChainPricingContract) mockCalculateCost(req ResourceRequestData) *big.Int {
	pc.mockMu.RLock()
	defer pc.mockMu.RUnlock()

	hours := big.NewInt(int64(req.DurationHours))
	total := big.NewInt(0)

	// CPU cost
	cpuCost := new(big.Int).Mul(pc.mockPrices.CPUPerCoreHour, big.NewInt(int64(req.CPUCores)))
	cpuCost.Mul(cpuCost, hours)
	total.Add(total, cpuCost)

	// Memory cost
	memCost := new(big.Int).Mul(pc.mockPrices.MemoryPerGBHour, big.NewInt(int64(req.MemoryGB)))
	memCost.Mul(memCost, hours)
	total.Add(total, memCost)

	// Storage cost
	storageCost := new(big.Int).Mul(pc.mockPrices.StoragePerGBHour, big.NewInt(int64(req.StorageGB)))
	storageCost.Mul(storageCost, hours)
	total.Add(total, storageCost)

	// GPU cost
	if req.GPUCount > 0 {
		gpuCost := new(big.Int).Mul(pc.mockPrices.GPUPerHour, big.NewInt(int64(req.GPUCount)))
		gpuCost.Mul(gpuCost, hours)
		total.Add(total, gpuCost)
	}

	return total
}

// CalculateProviderCost calculates cost using provider-specific pricing.
func (pc *OnChainPricingContract) CalculateProviderCost(ctx context.Context, provider common.Address, req ResourceRequestData) (*big.Int, error) {
	if pc.mockMode {
		// Use base cost for mock (providers don't override yet)
		return pc.mockCalculateCost(req), nil
	}

	abiReq := struct {
		CpuCores      uint16
		MemoryGB      uint16
		StorageGB     uint16
		BandwidthMbps uint16
		DurationHours uint32
		GpuCount      uint8
	}{
		CpuCores:      req.CPUCores,
		MemoryGB:      req.MemoryGB,
		StorageGB:     req.StorageGB,
		BandwidthMbps: req.BandwidthMbps,
		DurationHours: req.DurationHours,
		GpuCount:      req.GPUCount,
	}

	var result []interface{}
	err := pc.contract.Call(&bind.CallOpts{Context: ctx}, &result, "calculateProviderCost", provider, abiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate provider cost: %w", err)
	}

	if len(result) > 0 {
		if cost, ok := result[0].(*big.Int); ok {
			return cost, nil
		}
	}
	return big.NewInt(0), nil
}

// GetTokenPrice returns the current BUNKER token price from the Chainlink oracle.
func (pc *OnChainPricingContract) GetTokenPrice(ctx context.Context) (*big.Int, error) {
	if pc.mockMode {
		pc.mockMu.RLock()
		defer pc.mockMu.RUnlock()
		return new(big.Int).Set(pc.mockTokenPrice), nil
	}

	var result []interface{}
	err := pc.contract.Call(&bind.CallOpts{Context: ctx}, &result, "getTokenPrice")
	if err != nil {
		return nil, fmt.Errorf("failed to get token price: %w", err)
	}

	if len(result) > 0 {
		if price, ok := result[0].(*big.Int); ok {
			return price, nil
		}
	}
	return big.NewInt(0), nil
}

// GetPrices returns the current base resource prices.
func (pc *OnChainPricingContract) GetPrices(ctx context.Context) (*ResourcePricesData, error) {
	if pc.mockMode {
		pc.mockMu.RLock()
		defer pc.mockMu.RUnlock()
		return &ResourcePricesData{
			CPUPerCoreHour:     new(big.Int).Set(pc.mockPrices.CPUPerCoreHour),
			MemoryPerGBHour:    new(big.Int).Set(pc.mockPrices.MemoryPerGBHour),
			StoragePerGBHour:   new(big.Int).Set(pc.mockPrices.StoragePerGBHour),
			BandwidthPerGBHour: new(big.Int).Set(pc.mockPrices.BandwidthPerGBHour),
			GPUPerHour:         new(big.Int).Set(pc.mockPrices.GPUPerHour),
		}, nil
	}

	var result []interface{}
	err := pc.contract.Call(&bind.CallOpts{Context: ctx}, &result, "getPrices")
	if err != nil {
		return nil, fmt.Errorf("failed to get prices: %w", err)
	}

	data := &ResourcePricesData{
		CPUPerCoreHour:     big.NewInt(0),
		MemoryPerGBHour:    big.NewInt(0),
		StoragePerGBHour:   big.NewInt(0),
		BandwidthPerGBHour: big.NewInt(0),
		GPUPerHour:         big.NewInt(0),
	}
	if len(result) > 0 {
		if res, ok := result[0].(struct {
			CpuPerCoreHour     *big.Int
			MemoryPerGBHour    *big.Int
			StoragePerGBHour   *big.Int
			BandwidthPerGBHour *big.Int
			GpuPerHour         *big.Int
		}); ok {
			data.CPUPerCoreHour = res.CpuPerCoreHour
			data.MemoryPerGBHour = res.MemoryPerGBHour
			data.StoragePerGBHour = res.StoragePerGBHour
			data.BandwidthPerGBHour = res.BandwidthPerGBHour
			data.GPUPerHour = res.GpuPerHour
		}
	}
	return data, nil
}

// GetMultipliers returns the current pricing multipliers.
func (pc *OnChainPricingContract) GetMultipliers(ctx context.Context) (*MultipliersData, error) {
	if pc.mockMode {
		pc.mockMu.RLock()
		defer pc.mockMu.RUnlock()
		return &MultipliersData{
			DemandMultiplier: new(big.Int).Set(pc.mockMultipliers.DemandMultiplier),
			RegionMultiplier: new(big.Int).Set(pc.mockMultipliers.RegionMultiplier),
			TierDiscount:     new(big.Int).Set(pc.mockMultipliers.TierDiscount),
		}, nil
	}

	var result []interface{}
	err := pc.contract.Call(&bind.CallOpts{Context: ctx}, &result, "getMultipliers")
	if err != nil {
		return nil, fmt.Errorf("failed to get multipliers: %w", err)
	}

	data := &MultipliersData{
		DemandMultiplier: big.NewInt(10000),
		RegionMultiplier: big.NewInt(10000),
		TierDiscount:     big.NewInt(0),
	}
	if len(result) > 0 {
		if res, ok := result[0].(struct {
			DemandMultiplier *big.Int
			RegionMultiplier *big.Int
			TierDiscount     *big.Int
		}); ok {
			data.DemandMultiplier = res.DemandMultiplier
			data.RegionMultiplier = res.RegionMultiplier
			data.TierDiscount = res.TierDiscount
		}
	}
	return data, nil
}

// GetProviderPrices returns a provider's custom pricing multipliers.
func (pc *OnChainPricingContract) GetProviderPrices(ctx context.Context, provider common.Address) (*ProviderPricingData, error) {
	if pc.mockMode {
		pc.mockMu.RLock()
		defer pc.mockMu.RUnlock()
		pp, exists := pc.mockProviderPricing[provider]
		if !exists {
			return &ProviderPricingData{
				CPUMultiplier:     big.NewInt(10000), // 1x
				MemoryMultiplier:  big.NewInt(10000),
				StorageMultiplier: big.NewInt(10000),
				MinPrice:          big.NewInt(0),
				MaxPrice:          big.NewInt(0),
			}, nil
		}
		return &ProviderPricingData{
			CPUMultiplier:     new(big.Int).Set(pp.CPUMultiplier),
			MemoryMultiplier:  new(big.Int).Set(pp.MemoryMultiplier),
			StorageMultiplier: new(big.Int).Set(pp.StorageMultiplier),
			MinPrice:          new(big.Int).Set(pp.MinPrice),
			MaxPrice:          new(big.Int).Set(pp.MaxPrice),
		}, nil
	}

	var result []interface{}
	err := pc.contract.Call(&bind.CallOpts{Context: ctx}, &result, "getProviderPrices", provider)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider prices: %w", err)
	}

	data := &ProviderPricingData{
		CPUMultiplier:     big.NewInt(10000),
		MemoryMultiplier:  big.NewInt(10000),
		StorageMultiplier: big.NewInt(10000),
		MinPrice:          big.NewInt(0),
		MaxPrice:          big.NewInt(0),
	}
	if len(result) > 0 {
		if res, ok := result[0].(struct {
			CpuMultiplier     *big.Int
			MemoryMultiplier  *big.Int
			StorageMultiplier *big.Int
			MinPrice          *big.Int
			MaxPrice          *big.Int
		}); ok {
			data.CPUMultiplier = res.CpuMultiplier
			data.MemoryMultiplier = res.MemoryMultiplier
			data.StorageMultiplier = res.StorageMultiplier
			data.MinPrice = res.MinPrice
			data.MaxPrice = res.MaxPrice
		}
	}
	return data, nil
}

// ─── Admin Setters ───────────────────────────────────────────────────────────

// SetMaxMultiplierBps adjusts the maximum multiplier cap for pricing.
func (pc *OnChainPricingContract) SetMaxMultiplierBps(ctx context.Context, newMax *big.Int) (*types.Transaction, error) {
	if pc.mockMode {
		logging.Info("mock: setMaxMultiplierBps max=%s", newMax)
		return nil, nil
	}
	auth, err := pc.baseClient.GetTransactOpts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction options: %w", err)
	}
	tx, err := pc.contract.Transact(auth, "setMaxMultiplierBps", newMax)
	if err != nil {
		return nil, fmt.Errorf("failed to set max multiplier: %w", err)
	}
	return tx, nil
}

// SetStaleThresholdBounds adjusts the min/max bounds for the stale price threshold.
func (pc *OnChainPricingContract) SetStaleThresholdBounds(ctx context.Context, newMin, newMax *big.Int) (*types.Transaction, error) {
	if pc.mockMode {
		logging.Info("mock: setStaleThresholdBounds min=%s max=%s", newMin, newMax)
		return nil, nil
	}
	auth, err := pc.baseClient.GetTransactOpts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction options: %w", err)
	}
	tx, err := pc.contract.Transact(auth, "setStaleThresholdBounds", newMin, newMax)
	if err != nil {
		return nil, fmt.Errorf("failed to set stale threshold bounds: %w", err)
	}
	return tx, nil
}
