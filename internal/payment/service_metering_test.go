package payment

import (
	"math/big"
	"sync"
	"testing"
)

func TestServiceMeter_RecordStorageUpload(t *testing.T) {
	m := NewServiceMeter(DefaultServicePricing())

	m.RecordStorageUpload("0xaabb", 1024*1024*100) // 100 MB
	m.RecordStorageUpload("0xaabb", 1024*1024*50)  // +50 MB

	usage := m.GetUsage("0xaabb")
	if usage == nil {
		t.Fatal("expected usage record")
	}
	if usage.StorageBytes != 150*1024*1024 {
		t.Fatalf("StorageBytes = %d, want %d", usage.StorageBytes, 150*1024*1024)
	}
	if usage.StorageOps != 2 {
		t.Fatalf("StorageOps = %d, want 2", usage.StorageOps)
	}
}

func TestServiceMeter_RecordStorageDelete(t *testing.T) {
	m := NewServiceMeter(DefaultServicePricing())

	m.RecordStorageUpload("0xaabb", 1024*1024*100)
	m.RecordStorageDelete("0xaabb", 1024*1024*30) // delete 30 MB

	usage := m.GetUsage("0xaabb")
	if usage.StorageBytes != 70*1024*1024 {
		t.Fatalf("StorageBytes = %d, want %d", usage.StorageBytes, 70*1024*1024)
	}
	if usage.StorageOps != 2 {
		t.Fatalf("StorageOps = %d, want 2", usage.StorageOps)
	}

	// Delete more than stored — should clamp to 0
	m.RecordStorageDelete("0xaabb", 1024*1024*200)
	usage = m.GetUsage("0xaabb")
	if usage.StorageBytes != 0 {
		t.Fatalf("StorageBytes = %d, want 0 (clamped)", usage.StorageBytes)
	}
}

func TestServiceMeter_RecordProxySession(t *testing.T) {
	m := NewServiceMeter(DefaultServicePricing())

	m.RecordProxySession("0xproxy", 1024*1024, 5*1024*1024) // 1MB in, 5MB out

	usage := m.GetUsage("0xproxy")
	if usage.ProxyBytesIn != 1024*1024 {
		t.Fatalf("ProxyBytesIn = %d, want %d", usage.ProxyBytesIn, 1024*1024)
	}
	if usage.ProxyBytesOut != 5*1024*1024 {
		t.Fatalf("ProxyBytesOut = %d, want %d", usage.ProxyBytesOut, 5*1024*1024)
	}
	if usage.ProxySessions != 1 {
		t.Fatalf("ProxySessions = %d, want 1", usage.ProxySessions)
	}
}

func TestServiceMeter_RecordCrawlJob(t *testing.T) {
	m := NewServiceMeter(DefaultServicePricing())

	m.RecordCrawlJob("0xcrawl", 100, 2*1024*1024) // 100 pages, 2MB result
	m.RecordCrawlJob("0xcrawl", 50, 512*1024)     // 50 pages, 512KB

	usage := m.GetUsage("0xcrawl")
	if usage.CrawlPages != 150 {
		t.Fatalf("CrawlPages = %d, want 150", usage.CrawlPages)
	}
	if usage.CrawlJobs != 2 {
		t.Fatalf("CrawlJobs = %d, want 2", usage.CrawlJobs)
	}
}

func TestServiceMeter_RecordAgentInvocation(t *testing.T) {
	m := NewServiceMeter(DefaultServicePricing())

	m.RecordAgentInvocation("0xagent", 1500) // 1500 tokens
	m.RecordAgentInvocation("0xagent", 3000) // 3000 tokens

	usage := m.GetUsage("0xagent")
	if usage.AgentInvocations != 2 {
		t.Fatalf("AgentInvocations = %d, want 2", usage.AgentInvocations)
	}
	if usage.AgentTokensUsed != 4500 {
		t.Fatalf("AgentTokensUsed = %d, want 4500", usage.AgentTokensUsed)
	}
}

func TestServiceMeter_CalculateProxyCost(t *testing.T) {
	m := NewServiceMeter(DefaultServicePricing())

	// 1 GB total bandwidth
	cost := m.CalculateProxyCost(512*1024*1024, 512*1024*1024)
	if cost.Sign() <= 0 {
		t.Fatalf("expected positive cost, got %s", cost.String())
	}

	// Default: 0.02 BUNKER/GB = 20000000000000000 wei/GB
	// 1 GB → cost should be exactly 0.02 BUNKER
	expected := DefaultServicePricing().ProxyPerGB
	if cost.Cmp(expected) != 0 {
		t.Fatalf("1GB cost = %s, want %s", cost.String(), expected.String())
	}

	// Zero bytes → zero cost
	zeroCost := m.CalculateProxyCost(0, 0)
	if zeroCost.Sign() != 0 {
		t.Fatalf("zero bytes cost = %s, want 0", zeroCost.String())
	}

	// Tiny amount → minimum 1 wei
	tinyCost := m.CalculateProxyCost(1, 0)
	if tinyCost.Cmp(big.NewInt(1)) < 0 {
		t.Fatalf("tiny cost = %s, want >= 1", tinyCost.String())
	}
}

func TestServiceMeter_CalculateCrawlCost(t *testing.T) {
	m := NewServiceMeter(DefaultServicePricing())
	pricing := DefaultServicePricing()

	// 10 pages, 5 MB result
	cost := m.CalculateCrawlCost(10, 5*1024*1024)

	// Expected: 10 * CrawlPerPage + 5 * CrawlPerMBResult
	expectedPageCost := new(big.Int).Mul(pricing.CrawlPerPage, big.NewInt(10))
	expectedResultCost := new(big.Int).Mul(pricing.CrawlPerMBResult, big.NewInt(5))
	expected := new(big.Int).Add(expectedPageCost, expectedResultCost)

	if cost.Cmp(expected) != 0 {
		t.Fatalf("crawl cost = %s, want %s", cost.String(), expected.String())
	}

	// Sub-MB result → rounds up to 1 MB
	smallCost := m.CalculateCrawlCost(1, 100)
	minPageCost := new(big.Int).Set(pricing.CrawlPerPage)
	minResultCost := new(big.Int).Set(pricing.CrawlPerMBResult)
	minExpected := new(big.Int).Add(minPageCost, minResultCost)
	if smallCost.Cmp(minExpected) != 0 {
		t.Fatalf("small crawl cost = %s, want %s", smallCost.String(), minExpected.String())
	}
}

func TestServiceMeter_CalculateAgentCost(t *testing.T) {
	m := NewServiceMeter(DefaultServicePricing())
	pricing := DefaultServicePricing()

	// 2500 tokens
	cost := m.CalculateAgentCost(2500)

	// Expected: AgentPerInvocation + ceil(2500/1000) * AgentPer1KTokens
	// = invocation + 3 * per1K
	expectedBase := new(big.Int).Set(pricing.AgentPerInvocation)
	expectedTokens := new(big.Int).Mul(pricing.AgentPer1KTokens, big.NewInt(3))
	expected := new(big.Int).Add(expectedBase, expectedTokens)

	if cost.Cmp(expected) != 0 {
		t.Fatalf("agent cost = %s, want %s", cost.String(), expected.String())
	}

	// Zero tokens → just base invocation cost
	zeroCost := m.CalculateAgentCost(0)
	if zeroCost.Cmp(pricing.AgentPerInvocation) != 0 {
		t.Fatalf("zero token cost = %s, want %s", zeroCost.String(), pricing.AgentPerInvocation.String())
	}
}

func TestServiceMeter_CalculateStorageBill(t *testing.T) {
	m := NewServiceMeter(DefaultServicePricing())

	// Upload 1 GB
	m.RecordStorageUpload("0xstore", 1024*1024*1024)

	bill := m.CalculateStorageBill("0xstore")
	expected := DefaultServicePricing().StoragePerGBMonth
	if bill.Cmp(expected) != 0 {
		t.Fatalf("1GB bill = %s, want %s", bill.String(), expected.String())
	}

	// No usage → 0
	zeroBill := m.CalculateStorageBill("0xunknown")
	if zeroBill.Sign() != 0 {
		t.Fatalf("unknown wallet bill = %s, want 0", zeroBill.String())
	}
}

func TestServiceMeter_RecordCharge(t *testing.T) {
	m := NewServiceMeter(DefaultServicePricing())

	m.RecordCharge("0xwallet", big.NewInt(1000))
	m.RecordCharge("0xwallet", big.NewInt(2000))

	usage := m.GetUsage("0xwallet")
	if usage.TotalCharged.Cmp(big.NewInt(3000)) != 0 {
		t.Fatalf("TotalCharged = %s, want 3000", usage.TotalCharged.String())
	}
}

func TestServiceMeter_GetUsageReturnsSnapshot(t *testing.T) {
	m := NewServiceMeter(DefaultServicePricing())

	m.RecordStorageUpload("0xsnap", 1024)
	m.RecordCharge("0xsnap", big.NewInt(100))

	// Get snapshot
	usage := m.GetUsage("0xsnap")

	// Mutate the snapshot — should not affect meter
	usage.StorageBytes = 999999
	usage.TotalCharged.SetInt64(999999)

	original := m.GetUsage("0xsnap")
	if original.StorageBytes != 1024 {
		t.Fatalf("snapshot mutation leaked: StorageBytes = %d", original.StorageBytes)
	}
	if original.TotalCharged.Cmp(big.NewInt(100)) != 0 {
		t.Fatalf("snapshot mutation leaked: TotalCharged = %s", original.TotalCharged.String())
	}
}

func TestServiceMeter_ActiveWallets(t *testing.T) {
	m := NewServiceMeter(DefaultServicePricing())

	if m.ActiveWallets() != 0 {
		t.Fatalf("expected 0 active wallets")
	}

	m.RecordStorageUpload("0xa", 100)
	m.RecordProxySession("0xb", 100, 200)
	m.RecordCrawlJob("0xc", 1, 100)

	if m.ActiveWallets() != 3 {
		t.Fatalf("expected 3 active wallets, got %d", m.ActiveWallets())
	}

	// Same wallet again
	m.RecordAgentInvocation("0xa", 500)
	if m.ActiveWallets() != 3 {
		t.Fatalf("expected still 3, got %d", m.ActiveWallets())
	}
}

func TestServiceMeter_ConcurrentAccess(t *testing.T) {
	m := NewServiceMeter(DefaultServicePricing())

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(4)
		wallet := "0xconcurrent"
		go func() { defer wg.Done(); m.RecordStorageUpload(wallet, 1024) }()
		go func() { defer wg.Done(); m.RecordProxySession(wallet, 100, 200) }()
		go func() { defer wg.Done(); m.RecordCrawlJob(wallet, 1, 100) }()
		go func() { defer wg.Done(); m.RecordAgentInvocation(wallet, 50) }()
	}
	wg.Wait()

	usage := m.GetUsage("0xconcurrent")
	if usage.StorageBytes != 50*1024 {
		t.Fatalf("StorageBytes = %d, want %d", usage.StorageBytes, 50*1024)
	}
	if usage.ProxySessions != 50 {
		t.Fatalf("ProxySessions = %d, want 50", usage.ProxySessions)
	}
	if usage.CrawlJobs != 50 {
		t.Fatalf("CrawlJobs = %d, want 50", usage.CrawlJobs)
	}
	if usage.AgentInvocations != 50 {
		t.Fatalf("AgentInvocations = %d, want 50", usage.AgentInvocations)
	}
}

func TestServiceMeter_UnknownWallet(t *testing.T) {
	m := NewServiceMeter(DefaultServicePricing())

	usage := m.GetUsage("0xnonexistent")
	if usage != nil {
		t.Fatal("expected nil for unknown wallet")
	}
}

func TestServiceMeter_DefaultPricing(t *testing.T) {
	pricing := DefaultServicePricing()

	// Verify all fields are non-nil and positive
	fields := []*big.Int{
		pricing.StoragePerGBMonth,
		pricing.ProxyPerGB,
		pricing.CrawlPerPage,
		pricing.CrawlPerMBResult,
		pricing.AgentPerInvocation,
		pricing.AgentPer1KTokens,
	}
	names := []string{
		"StoragePerGBMonth", "ProxyPerGB", "CrawlPerPage",
		"CrawlPerMBResult", "AgentPerInvocation", "AgentPer1KTokens",
	}
	for i, f := range fields {
		if f == nil {
			t.Fatalf("%s is nil", names[i])
		}
		if f.Sign() <= 0 {
			t.Fatalf("%s = %s, want > 0", names[i], f.String())
		}
	}
}
