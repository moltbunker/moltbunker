package payment

import (
	"math/big"
	"testing"
)

const (
	oneMB = int64(1024 * 1024)
	oneGB = int64(1024 * 1024 * 1024)
)

func newTestMeter() *ServiceMeter {
	return NewServiceMeter(DefaultServicePricing())
}

// TestServiceMeterStorageRoundTrip records a 1GB upload and asserts the monthly
// bill is non-zero and matches the configured per-GB-month rate.
func TestServiceMeterStorageRoundTrip(t *testing.T) {
	m := newTestMeter()
	const wallet = "0xabc"

	m.RecordStorageUpload(wallet, oneGB)

	bill := m.CalculateStorageBill(wallet)
	if bill.Sign() <= 0 {
		t.Fatalf("expected non-zero storage bill, got %s", bill)
	}
	// 1 GB exactly => one unit of StoragePerGBMonth.
	if bill.Cmp(DefaultServicePricing().StoragePerGBMonth) != 0 {
		t.Errorf("expected bill %s for 1GB, got %s",
			DefaultServicePricing().StoragePerGBMonth, bill)
	}

	// A delete returns usage (and therefore the bill) to zero.
	m.RecordStorageDelete(wallet, oneGB)
	if got := m.CalculateStorageBill(wallet); got.Sign() != 0 {
		t.Errorf("expected zero bill after delete, got %s", got)
	}
}

// TestServiceMeterProxyCost records a proxy session and asserts the cost scales
// with total bytes transferred.
func TestServiceMeterProxyCost(t *testing.T) {
	m := newTestMeter()
	const wallet = "0xdef"

	in := int64(100) * oneMB
	out := int64(200) * oneMB
	m.RecordProxySession(wallet, in, out)

	usage := m.GetUsage(wallet)
	if usage == nil {
		t.Fatal("expected usage recorded for proxy session")
	}
	if usage.ProxyBytesIn != in || usage.ProxyBytesOut != out || usage.ProxySessions != 1 {
		t.Errorf("usage mismatch: in=%d out=%d sessions=%d",
			usage.ProxyBytesIn, usage.ProxyBytesOut, usage.ProxySessions)
	}

	cost := m.CalculateProxyCost(in, out)
	if cost.Sign() <= 0 {
		t.Fatalf("expected non-zero proxy cost, got %s", cost)
	}
	// Cost must be proportional: double the bytes => (approximately) double the cost.
	doubleCost := m.CalculateProxyCost(in*2, out*2)
	want := new(big.Int).Mul(cost, big.NewInt(2))
	if doubleCost.Cmp(want) != 0 {
		t.Errorf("proxy cost not proportional: cost=%s doubleCost=%s want=%s",
			cost, doubleCost, want)
	}
}

// TestServiceMeterCrawlCost verifies the crawl cost accounts for both per-page
// and per-MB-of-result components.
func TestServiceMeterCrawlCost(t *testing.T) {
	m := newTestMeter()
	const wallet = "0x123"

	pages := int64(50)
	resultBytes := int64(5) * oneMB
	m.RecordCrawlJob(wallet, pages, resultBytes)

	usage := m.GetUsage(wallet)
	if usage == nil || usage.CrawlJobs != 1 || usage.CrawlPages != pages {
		t.Fatalf("crawl usage not recorded correctly: %+v", usage)
	}

	pricing := DefaultServicePricing()
	pageComponent := new(big.Int).Mul(pricing.CrawlPerPage, big.NewInt(pages))
	mbComponent := new(big.Int).Mul(pricing.CrawlPerMBResult, big.NewInt(5))
	wantTotal := new(big.Int).Add(pageComponent, mbComponent)

	cost := m.CalculateCrawlCost(pages, resultBytes)
	if cost.Cmp(wantTotal) != 0 {
		t.Errorf("crawl cost: want %s (pages %s + mb %s), got %s",
			wantTotal, pageComponent, mbComponent, cost)
	}
	// Both components must contribute (cost strictly greater than either alone).
	if cost.Cmp(pageComponent) <= 0 {
		t.Error("crawl cost must exceed the per-page component alone")
	}
}

// TestServiceMeterAgentInvocation verifies agent cost = base invocation + per-1K-token rate.
func TestServiceMeterAgentInvocation(t *testing.T) {
	m := newTestMeter()
	const wallet = "0x456"

	m.RecordAgentInvocation(wallet, 1000)

	usage := m.GetUsage(wallet)
	if usage == nil || usage.AgentInvocations != 1 || usage.AgentTokensUsed != 1000 {
		t.Fatalf("agent usage not recorded correctly: %+v", usage)
	}

	pricing := DefaultServicePricing()
	// 1000 tokens => exactly 1 * per-1K-token rate, plus the base invocation.
	want := new(big.Int).Add(pricing.AgentPerInvocation, pricing.AgentPer1KTokens)
	cost := m.CalculateAgentCost(1000)
	if cost.Cmp(want) != 0 {
		t.Errorf("agent cost: want base+1K rate %s, got %s", want, cost)
	}

	// Zero tokens => just the base invocation cost.
	base := m.CalculateAgentCost(0)
	if base.Cmp(pricing.AgentPerInvocation) != 0 {
		t.Errorf("agent base cost: want %s, got %s", pricing.AgentPerInvocation, base)
	}
}
