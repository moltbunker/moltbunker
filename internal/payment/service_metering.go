package payment

import (
	"math/big"
	"sync"
	"time"
)

// StorageMeteringHook is the metering contract the storage engine depends on.
// PaymentService satisfies it structurally via its RecordStorageUpload /
// RecordStorageDelete methods. Defining it here (and mirroring it in the
// storage package) lets storage call metering without importing payment,
// avoiding an import cycle.
type StorageMeteringHook interface {
	RecordStorageUpload(wallet string, bytes int64)
	RecordStorageDelete(wallet string, bytes int64)
}

// ProxyMeteringHook is the metering contract the proxy server depends on.
// PaymentService satisfies it structurally via RecordProxySession.
type ProxyMeteringHook interface {
	RecordProxySession(wallet string, bytesIn, bytesOut int64)
}

// CrawlMeteringHook is the metering contract the crawl scheduler depends on.
// PaymentService satisfies it structurally via RecordCrawlJob.
type CrawlMeteringHook interface {
	RecordCrawlJob(wallet string, pagesCrawled, resultBytes int64)
}

// AgentMeteringHook is the metering contract the agent REST handler depends on.
// PaymentService satisfies it structurally via RecordAgentInvocation.
type AgentMeteringHook interface {
	RecordAgentInvocation(wallet string, tokensUsed int64)
}

// ServiceType identifies which P0 service generated usage.
type ServiceType string

const (
	ServiceStorage ServiceType = "storage"
	ServiceProxy   ServiceType = "proxy"
	ServiceCrawl   ServiceType = "crawl"
	ServiceAgent   ServiceType = "agent"
)

// ServiceMeter tracks per-wallet, per-service resource usage and converts it
// to BUNKER costs. It acts as the billing bridge between P0 service operations
// and the payment system.
//
// Pricing models:
//   - Storage: per GB-month (sweep-based billing)
//   - Proxy: per GB transferred (deducted per session)
//   - Crawl: per page + per MB result (deducted per job)
//   - Agent: per invocation + per 1K tokens (deducted per invocation)
type ServiceMeter struct {
	mu      sync.RWMutex
	records map[string]*WalletServiceUsage // key: wallet address
	pricing ServicePricing
}

// ServicePricing holds the per-unit rates for each P0 service (in BUNKER wei).
type ServicePricing struct {
	StoragePerGBMonth  *big.Int // Storage cost per GB-month
	ProxyPerGB         *big.Int // Proxy bandwidth cost per GB
	CrawlPerPage       *big.Int // Crawl cost per page
	CrawlPerMBResult   *big.Int // Crawl result storage cost per MB
	AgentPerInvocation *big.Int // Agent base cost per invocation
	AgentPer1KTokens   *big.Int // Agent LLM token cost per 1K tokens
}

// DefaultServicePricing returns sensible pricing defaults.
// Based on 20,000 BUNKER = $1 USD.
func DefaultServicePricing() ServicePricing {
	return ServicePricing{
		StoragePerGBMonth:  mustParseWei("50000000000000000"),   // 0.05 BUNKER/GB-month
		ProxyPerGB:         mustParseWei("20000000000000000"),   // 0.02 BUNKER/GB
		CrawlPerPage:      mustParseWei("1000000000000000"),    // 0.001 BUNKER/page
		CrawlPerMBResult:  mustParseWei("5000000000000000"),    // 0.005 BUNKER/MB
		AgentPerInvocation: mustParseWei("10000000000000000"),   // 0.01 BUNKER/invocation
		AgentPer1KTokens:  mustParseWei("100000000000000000"),  // 0.1 BUNKER/1K tokens
	}
}

func mustParseWei(s string) *big.Int {
	v, _ := new(big.Int).SetString(s, 10)
	return v
}

// WalletServiceUsage tracks cumulative usage for a wallet across all services.
type WalletServiceUsage struct {
	Wallet string `json:"wallet"`

	// Storage
	StorageBytes int64 `json:"storage_bytes"`
	StorageOps   int64 `json:"storage_ops"` // PUT + GET + DELETE count

	// Proxy
	ProxyBytesIn  int64 `json:"proxy_bytes_in"`
	ProxyBytesOut int64 `json:"proxy_bytes_out"`
	ProxySessions int64 `json:"proxy_sessions"`

	// Crawl
	CrawlPages       int64 `json:"crawl_pages"`
	CrawlResultBytes int64 `json:"crawl_result_bytes"`
	CrawlJobs        int64 `json:"crawl_jobs"`

	// Agent
	AgentInvocations int64 `json:"agent_invocations"`
	AgentTokensUsed  int64 `json:"agent_tokens_used"`

	// Billing
	TotalCharged *big.Int  `json:"total_charged"` // Lifetime charges in wei
	LastActivity time.Time `json:"last_activity"`
}

// NewServiceMeter creates a new service meter with the given pricing.
func NewServiceMeter(pricing ServicePricing) *ServiceMeter {
	return &ServiceMeter{
		records: make(map[string]*WalletServiceUsage),
		pricing: pricing,
	}
}

// RecordStorageUpload records a storage upload operation for billing.
func (m *ServiceMeter) RecordStorageUpload(wallet string, sizeBytes int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	r := m.getOrCreate(wallet)
	r.StorageBytes += sizeBytes
	r.StorageOps++
	r.LastActivity = time.Now()
}

// RecordStorageDelete records a storage delete operation.
func (m *ServiceMeter) RecordStorageDelete(wallet string, sizeBytes int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	r := m.getOrCreate(wallet)
	r.StorageBytes -= sizeBytes
	if r.StorageBytes < 0 {
		r.StorageBytes = 0
	}
	r.StorageOps++
	r.LastActivity = time.Now()
}

// RecordProxySession records proxy bandwidth usage for a completed session.
func (m *ServiceMeter) RecordProxySession(wallet string, bytesIn, bytesOut int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	r := m.getOrCreate(wallet)
	r.ProxyBytesIn += bytesIn
	r.ProxyBytesOut += bytesOut
	r.ProxySessions++
	r.LastActivity = time.Now()
}

// RecordCrawlJob records a completed crawl job for billing.
func (m *ServiceMeter) RecordCrawlJob(wallet string, pagesCrawled int64, resultBytes int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	r := m.getOrCreate(wallet)
	r.CrawlPages += pagesCrawled
	r.CrawlResultBytes += resultBytes
	r.CrawlJobs++
	r.LastActivity = time.Now()
}

// RecordAgentInvocation records an agent invocation for billing.
func (m *ServiceMeter) RecordAgentInvocation(wallet string, tokensUsed int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	r := m.getOrCreate(wallet)
	r.AgentInvocations++
	r.AgentTokensUsed += tokensUsed
	r.LastActivity = time.Now()
}

// CalculateProxyCost calculates the cost for a proxy session.
func (m *ServiceMeter) CalculateProxyCost(bytesIn, bytesOut int64) *big.Int {
	totalBytes := bytesIn + bytesOut
	if totalBytes <= 0 {
		return big.NewInt(0)
	}
	// cost = pricing.ProxyPerGB * totalBytes / bytesPerGB
	cost := new(big.Int).Mul(m.pricing.ProxyPerGB, big.NewInt(totalBytes))
	cost.Div(cost, big.NewInt(1024*1024*1024))
	if cost.Sign() <= 0 {
		return big.NewInt(1) // minimum 1 wei
	}
	return cost
}

// CalculateCrawlCost calculates the cost for a crawl job.
func (m *ServiceMeter) CalculateCrawlCost(pagesCrawled int64, resultBytes int64) *big.Int {
	// Page cost
	pageCost := new(big.Int).Mul(m.pricing.CrawlPerPage, big.NewInt(pagesCrawled))

	// Result storage cost (per MB)
	resultMB := (resultBytes + 1024*1024 - 1) / (1024 * 1024)
	if resultMB < 1 && resultBytes > 0 {
		resultMB = 1
	}
	resultCost := new(big.Int).Mul(m.pricing.CrawlPerMBResult, big.NewInt(resultMB))

	total := new(big.Int).Add(pageCost, resultCost)
	if total.Sign() <= 0 && pagesCrawled > 0 {
		return big.NewInt(1) // minimum 1 wei
	}
	return total
}

// CalculateAgentCost calculates the cost for an agent invocation.
func (m *ServiceMeter) CalculateAgentCost(tokensUsed int64) *big.Int {
	// Base invocation cost
	cost := new(big.Int).Set(m.pricing.AgentPerInvocation)

	// Token cost (per 1K tokens)
	if tokensUsed > 0 {
		thousands := (tokensUsed + 999) / 1000 // round up
		tokenCost := new(big.Int).Mul(m.pricing.AgentPer1KTokens, big.NewInt(thousands))
		cost.Add(cost, tokenCost)
	}

	return cost
}

// CalculateStorageBill calculates the monthly storage bill for a wallet.
func (m *ServiceMeter) CalculateStorageBill(wallet string) *big.Int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	r, ok := m.records[wallet]
	if !ok || r.StorageBytes <= 0 {
		return big.NewInt(0)
	}

	// cost = pricing.StoragePerGBMonth * bytes / bytesPerGB (ceiling div)
	cost := new(big.Int).Mul(m.pricing.StoragePerGBMonth, big.NewInt(r.StorageBytes))
	divisor := big.NewInt(1024 * 1024 * 1024)
	cost.Add(cost, new(big.Int).Sub(divisor, big.NewInt(1)))
	cost.Div(cost, divisor)
	return cost
}

// GetUsage returns usage data for a wallet, or nil if not tracked.
func (m *ServiceMeter) GetUsage(wallet string) *WalletServiceUsage {
	m.mu.RLock()
	defer m.mu.RUnlock()
	r, ok := m.records[wallet]
	if !ok {
		return nil
	}
	cp := *r
	if r.TotalCharged != nil {
		cp.TotalCharged = new(big.Int).Set(r.TotalCharged)
	}
	return &cp
}

// RecordCharge records that a charge was applied to a wallet.
func (m *ServiceMeter) RecordCharge(wallet string, amount *big.Int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	r := m.getOrCreate(wallet)
	if r.TotalCharged == nil {
		r.TotalCharged = new(big.Int)
	}
	r.TotalCharged.Add(r.TotalCharged, amount)
}

// ActiveWallets returns the number of wallets with recorded usage.
func (m *ServiceMeter) ActiveWallets() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.records)
}

func (m *ServiceMeter) getOrCreate(wallet string) *WalletServiceUsage {
	r, ok := m.records[wallet]
	if !ok {
		r = &WalletServiceUsage{
			Wallet:       wallet,
			TotalCharged: new(big.Int),
		}
		m.records[wallet] = r
	}
	return r
}
