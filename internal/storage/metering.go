package storage

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/moltbunker/moltbunker/internal/logging"
)

// StorageMeter tracks per-wallet storage usage and calculates billing.
type StorageMeter struct {
	mu      sync.RWMutex
	wallets map[string]*WalletUsage

	// Pricing (in wei per GB-month)
	ratePerGBMonth    *big.Int
	redundancyFactor  int // number of replicas (default: 3)
}

// WalletUsage tracks storage usage for a single wallet.
type WalletUsage struct {
	Address      string    `json:"address"`
	TotalBytes   int64     `json:"total_bytes"`
	ObjectCount  int64     `json:"object_count"`
	BucketCount  int       `json:"bucket_count"`
	LastUpdated  time.Time `json:"last_updated"`
}

// MeteringConfig configures the storage meter.
type MeteringConfig struct {
	RatePerGBMonth   *big.Int // Price per GB-month in wei (default: from PricingConfig)
	RedundancyFactor int      // Number of replicas to charge for (default: 3)
}

// DefaultMeteringConfig returns sensible defaults.
// Uses StoragePerGBMonth from PricingConfig: 50000000000000000 (0.05 BUNKER).
func DefaultMeteringConfig() MeteringConfig {
	rate := new(big.Int)
	rate.SetString("50000000000000000", 10) // 0.05 BUNKER in wei
	return MeteringConfig{
		RatePerGBMonth:   rate,
		RedundancyFactor: 3,
	}
}

// NewStorageMeter creates a new storage meter.
func NewStorageMeter(cfg MeteringConfig) *StorageMeter {
	if cfg.RatePerGBMonth == nil {
		cfg = DefaultMeteringConfig()
	}
	if cfg.RedundancyFactor <= 0 {
		cfg.RedundancyFactor = 3
	}
	return &StorageMeter{
		wallets:          make(map[string]*WalletUsage),
		ratePerGBMonth:   new(big.Int).Set(cfg.RatePerGBMonth),
		redundancyFactor: cfg.RedundancyFactor,
	}
}

// RecordUpload records a new object upload for billing.
func (m *StorageMeter) RecordUpload(wallet string, sizeBytes int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	usage := m.getOrCreateLocked(wallet)
	usage.TotalBytes += sizeBytes
	usage.ObjectCount++
	usage.LastUpdated = time.Now()
}

// RecordDelete records an object deletion for billing.
func (m *StorageMeter) RecordDelete(wallet string, sizeBytes int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	usage := m.getOrCreateLocked(wallet)
	usage.TotalBytes -= sizeBytes
	if usage.TotalBytes < 0 {
		usage.TotalBytes = 0
	}
	usage.ObjectCount--
	if usage.ObjectCount < 0 {
		usage.ObjectCount = 0
	}
	usage.LastUpdated = time.Now()
}

// RecordBucketCreate records a bucket creation.
func (m *StorageMeter) RecordBucketCreate(wallet string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	usage := m.getOrCreateLocked(wallet)
	usage.BucketCount++
	usage.LastUpdated = time.Now()
}

// RecordBucketDelete records a bucket deletion.
func (m *StorageMeter) RecordBucketDelete(wallet string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	usage := m.getOrCreateLocked(wallet)
	usage.BucketCount--
	if usage.BucketCount < 0 {
		usage.BucketCount = 0
	}
	usage.LastUpdated = time.Now()
}

// GetUsage returns current usage for a wallet.
func (m *StorageMeter) GetUsage(wallet string) *WalletUsage {
	m.mu.RLock()
	defer m.mu.RUnlock()

	usage, ok := m.wallets[wallet]
	if !ok {
		return &WalletUsage{Address: wallet}
	}

	// Return copy
	cp := *usage
	return &cp
}

// CalculateMonthlyBill calculates the monthly storage bill for a wallet.
// Formula: RatePerGBMonth × (TotalBytes / 1GB) × RedundancyFactor
func (m *StorageMeter) CalculateMonthlyBill(wallet string) *big.Int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	usage, ok := m.wallets[wallet]
	if !ok || usage.TotalBytes == 0 {
		return big.NewInt(0)
	}

	return m.calculateCost(usage.TotalBytes)
}

// CalculateCostForSize calculates the monthly cost for a given size.
func (m *StorageMeter) CalculateCostForSize(sizeBytes int64) *big.Int {
	return m.calculateCost(sizeBytes)
}

// calculateCost computes storage cost using big.Int arithmetic.
// Uses ceiling division to ensure at least 1 wei for any storage > 0.
func (m *StorageMeter) calculateCost(sizeBytes int64) *big.Int {
	if sizeBytes <= 0 {
		return big.NewInt(0)
	}

	// cost = rate × sizeBytes × redundancy / bytesPerGB
	// Use big.Int to avoid overflow
	const bytesPerGB = 1024 * 1024 * 1024

	cost := new(big.Int).Set(m.ratePerGBMonth)
	cost.Mul(cost, big.NewInt(sizeBytes))
	cost.Mul(cost, big.NewInt(int64(m.redundancyFactor)))

	// Ceiling division: (a + b - 1) / b
	divisor := big.NewInt(bytesPerGB)
	cost.Add(cost, new(big.Int).Sub(divisor, big.NewInt(1)))
	cost.Div(cost, divisor)

	return cost
}

// SweepAndBill calculates bills for all wallets and returns billing records.
// Called periodically (e.g., hourly) by the daemon.
func (m *StorageMeter) SweepAndBill(_ context.Context) []BillingRecord {
	m.mu.RLock()
	defer m.mu.RUnlock()

	now := time.Now()
	var records []BillingRecord

	for wallet, usage := range m.wallets {
		if usage.TotalBytes <= 0 {
			continue
		}

		cost := m.calculateCost(usage.TotalBytes)
		if cost.Sign() <= 0 {
			continue
		}

		records = append(records, BillingRecord{
			Wallet:     wallet,
			TotalBytes: usage.TotalBytes,
			Cost:       cost,
			Period:     now,
		})
	}

	if len(records) > 0 {
		logging.Info("storage billing sweep",
			"wallets", len(records),
			logging.Component("storage-metering"))
	}

	return records
}

// BillingRecord represents a monthly storage bill for a wallet.
type BillingRecord struct {
	Wallet     string    `json:"wallet"`
	TotalBytes int64     `json:"total_bytes"`
	Cost       *big.Int  `json:"cost"`       // Cost in wei
	Period     time.Time `json:"period"`
}

// FormatCost returns a human-readable cost string in BUNKER.
func (br *BillingRecord) FormatCost() string {
	if br.Cost == nil || br.Cost.Sign() == 0 {
		return "0 BUNKER"
	}

	// 1 BUNKER = 10^18 wei
	bunker := new(big.Float).SetInt(br.Cost)
	divisor := new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))
	bunker.Quo(bunker, divisor)

	return fmt.Sprintf("%.6f BUNKER", bunker)
}

// AllUsage returns usage for all tracked wallets.
func (m *StorageMeter) AllUsage() []WalletUsage {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]WalletUsage, 0, len(m.wallets))
	for _, u := range m.wallets {
		cp := *u
		result = append(result, cp)
	}
	return result
}

func (m *StorageMeter) getOrCreateLocked(wallet string) *WalletUsage {
	usage, ok := m.wallets[wallet]
	if !ok {
		usage = &WalletUsage{Address: wallet}
		m.wallets[wallet] = usage
	}
	return usage
}
