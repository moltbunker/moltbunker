package storage

import (
	"context"
	"math/big"
	"testing"
)

func TestMeter_RecordUploadAndUsage(t *testing.T) {
	m := NewStorageMeter(DefaultMeteringConfig())

	m.RecordUpload("wallet1", 1024*1024)   // 1MB
	m.RecordUpload("wallet1", 2*1024*1024) // 2MB

	usage := m.GetUsage("wallet1")
	if usage.TotalBytes != 3*1024*1024 {
		t.Errorf("total bytes = %d, want %d", usage.TotalBytes, 3*1024*1024)
	}
	if usage.ObjectCount != 2 {
		t.Errorf("object count = %d, want 2", usage.ObjectCount)
	}
}

func TestMeter_RecordDelete(t *testing.T) {
	m := NewStorageMeter(DefaultMeteringConfig())

	m.RecordUpload("wallet1", 1000)
	m.RecordUpload("wallet1", 2000)
	m.RecordDelete("wallet1", 1000)

	usage := m.GetUsage("wallet1")
	if usage.TotalBytes != 2000 {
		t.Errorf("total bytes after delete = %d, want 2000", usage.TotalBytes)
	}
	if usage.ObjectCount != 1 {
		t.Errorf("object count after delete = %d, want 1", usage.ObjectCount)
	}
}

func TestMeter_DeleteBelowZero(t *testing.T) {
	m := NewStorageMeter(DefaultMeteringConfig())

	m.RecordUpload("w", 100)
	m.RecordDelete("w", 500) // delete more than exists

	usage := m.GetUsage("w")
	if usage.TotalBytes != 0 {
		t.Errorf("bytes should floor at 0, got %d", usage.TotalBytes)
	}
	if usage.ObjectCount != 0 {
		t.Errorf("object count should floor at 0, got %d", usage.ObjectCount)
	}
}

func TestMeter_BucketTracking(t *testing.T) {
	m := NewStorageMeter(DefaultMeteringConfig())

	m.RecordBucketCreate("w")
	m.RecordBucketCreate("w")
	m.RecordBucketDelete("w")

	usage := m.GetUsage("w")
	if usage.BucketCount != 1 {
		t.Errorf("bucket count = %d, want 1", usage.BucketCount)
	}
}

func TestMeter_CalculateMonthlyBill_Zero(t *testing.T) {
	m := NewStorageMeter(DefaultMeteringConfig())

	bill := m.CalculateMonthlyBill("nobody")
	if bill.Sign() != 0 {
		t.Errorf("bill for unknown wallet should be 0, got %s", bill.String())
	}
}

func TestMeter_CalculateMonthlyBill_1GB(t *testing.T) {
	m := NewStorageMeter(DefaultMeteringConfig())

	m.RecordUpload("w", 1024*1024*1024) // 1 GB

	bill := m.CalculateMonthlyBill("w")
	// Expected: 0.05 BUNKER × 1 GB × 3 replicas = 0.15 BUNKER
	// In wei: 50000000000000000 × 1 × 3 = 150000000000000000
	expected := new(big.Int)
	expected.SetString("150000000000000000", 10)

	if bill.Cmp(expected) != 0 {
		t.Errorf("1GB monthly bill = %s, want %s", bill.String(), expected.String())
	}
}

func TestMeter_CalculateMonthlyBill_SmallFile(t *testing.T) {
	m := NewStorageMeter(DefaultMeteringConfig())

	m.RecordUpload("w", 1) // 1 byte

	bill := m.CalculateMonthlyBill("w")
	// Should be > 0 (ceiling division ensures at least 1 wei)
	if bill.Sign() <= 0 {
		t.Error("bill for 1 byte should be > 0")
	}
}

func TestMeter_CalculateCostForSize(t *testing.T) {
	m := NewStorageMeter(DefaultMeteringConfig())

	// 10 GB
	cost := m.CalculateCostForSize(10 * 1024 * 1024 * 1024)
	// 0.05 × 10 × 3 = 1.5 BUNKER = 1500000000000000000 wei
	expected := new(big.Int)
	expected.SetString("1500000000000000000", 10)

	if cost.Cmp(expected) != 0 {
		t.Errorf("10GB cost = %s, want %s", cost.String(), expected.String())
	}
}

func TestMeter_CalculateCostForSize_Zero(t *testing.T) {
	m := NewStorageMeter(DefaultMeteringConfig())

	cost := m.CalculateCostForSize(0)
	if cost.Sign() != 0 {
		t.Error("cost for 0 bytes should be 0")
	}

	cost = m.CalculateCostForSize(-100)
	if cost.Sign() != 0 {
		t.Error("cost for negative bytes should be 0")
	}
}

func TestMeter_SweepAndBill(t *testing.T) {
	m := NewStorageMeter(DefaultMeteringConfig())

	m.RecordUpload("w1", 1024*1024*1024)   // 1GB
	m.RecordUpload("w2", 2*1024*1024*1024) // 2GB

	records := m.SweepAndBill(context.Background())
	if len(records) != 2 {
		t.Fatalf("billing records = %d, want 2", len(records))
	}

	// Both should have positive costs
	for _, r := range records {
		if r.Cost.Sign() <= 0 {
			t.Errorf("wallet %s cost should be > 0", r.Wallet)
		}
		if r.TotalBytes <= 0 {
			t.Errorf("wallet %s should have bytes > 0", r.Wallet)
		}
	}
}

func TestMeter_SweepSkipsZeroUsage(t *testing.T) {
	m := NewStorageMeter(DefaultMeteringConfig())

	m.RecordUpload("w1", 100)
	m.RecordDelete("w1", 100) // back to zero

	records := m.SweepAndBill(context.Background())
	if len(records) != 0 {
		t.Errorf("sweep should skip zero usage, got %d records", len(records))
	}
}

func TestMeter_AllUsage(t *testing.T) {
	m := NewStorageMeter(DefaultMeteringConfig())

	m.RecordUpload("w1", 100)
	m.RecordUpload("w2", 200)

	all := m.AllUsage()
	if len(all) != 2 {
		t.Fatalf("all usage = %d, want 2", len(all))
	}
}

func TestMeter_GetUsage_Unknown(t *testing.T) {
	m := NewStorageMeter(DefaultMeteringConfig())

	usage := m.GetUsage("unknown")
	if usage.TotalBytes != 0 {
		t.Errorf("unknown wallet bytes = %d, want 0", usage.TotalBytes)
	}
	if usage.Address != "unknown" {
		t.Errorf("address = %q, want unknown", usage.Address)
	}
}

func TestMeter_CustomRate(t *testing.T) {
	rate := big.NewInt(100000000000000000) // 0.1 BUNKER
	m := NewStorageMeter(MeteringConfig{
		RatePerGBMonth:   rate,
		RedundancyFactor: 2,
	})

	m.RecordUpload("w", 1024*1024*1024) // 1GB

	bill := m.CalculateMonthlyBill("w")
	// 0.1 × 1 × 2 = 0.2 BUNKER = 200000000000000000 wei
	expected := new(big.Int)
	expected.SetString("200000000000000000", 10)

	if bill.Cmp(expected) != 0 {
		t.Errorf("custom rate bill = %s, want %s", bill.String(), expected.String())
	}
}

func TestBillingRecord_FormatCost(t *testing.T) {
	cost := new(big.Int)
	cost.SetString("1500000000000000000", 10) // 1.5 BUNKER

	br := &BillingRecord{Cost: cost}
	formatted := br.FormatCost()

	if formatted != "1.500000 BUNKER" {
		t.Errorf("formatted = %q, want 1.500000 BUNKER", formatted)
	}
}

func TestBillingRecord_FormatCost_Zero(t *testing.T) {
	br := &BillingRecord{Cost: big.NewInt(0)}
	if br.FormatCost() != "0 BUNKER" {
		t.Errorf("zero cost format = %q", br.FormatCost())
	}

	br2 := &BillingRecord{}
	if br2.FormatCost() != "0 BUNKER" {
		t.Errorf("nil cost format = %q", br2.FormatCost())
	}
}

func TestMeter_GetUsage_ReturnsCopy(t *testing.T) {
	m := NewStorageMeter(DefaultMeteringConfig())
	m.RecordUpload("w", 100)

	usage := m.GetUsage("w")
	usage.TotalBytes = 999999 // mutate copy

	original := m.GetUsage("w")
	if original.TotalBytes != 100 {
		t.Error("mutation of copy should not affect original")
	}
}
