package types

import "time"

// ServiceCapabilities extends NodeCapabilities with P0 service availability.
// These are advertised during the announce handshake and stored on peer records.
type ServiceCapabilities struct {
	// Object Storage
	StorageAvailable  bool `json:"storage_available,omitempty"`
	StorageCapacityGB int  `json:"storage_capacity_gb,omitempty"`

	// Decentralized Proxy
	ProxyAvailable bool `json:"proxy_available,omitempty"`
	ProxyTor       bool `json:"proxy_tor,omitempty"`

	// AI Agent Runtime
	AgentAvailable  bool     `json:"agent_available,omitempty"`
	AgentFrameworks []string `json:"agent_frameworks,omitempty"`

	// Web Crawling
	CrawlAvailable bool `json:"crawl_available,omitempty"`
}

// --- Object Storage Types ---

// StorageBucket represents a named storage bucket owned by a wallet.
type StorageBucket struct {
	Name      string    `json:"name"`
	Owner     string    `json:"owner"`      // Wallet address
	CreatedAt time.Time `json:"created_at"`
	Region    string    `json:"region,omitempty"`
}

// StorageObject represents an object stored in a bucket.
type StorageObject struct {
	Bucket      string    `json:"bucket"`
	Key         string    `json:"key"`
	Size        int64     `json:"size"`
	ContentType string    `json:"content_type,omitempty"`
	ETag        string    `json:"etag"` // MD5 of plaintext content
	CID         string    `json:"cid"`  // IPFS CID of encrypted blob
	Owner       string    `json:"owner"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Encryption metadata
	EncryptedDEK []byte `json:"encrypted_dek,omitempty"` // DEK encrypted with owner's X25519 pubkey
	DEKNonce     []byte `json:"dek_nonce,omitempty"`     // Nonce used for DEK encryption
}

// StorageQuota tracks per-wallet storage usage.
type StorageQuota struct {
	WalletAddress string `json:"wallet_address"`
	UsedBytes     int64  `json:"used_bytes"`
	ObjectCount   int64  `json:"object_count"`
	BucketCount   int    `json:"bucket_count"`
}

// MultipartUpload tracks an in-progress multipart upload.
type MultipartUpload struct {
	UploadID  string    `json:"upload_id"`
	Bucket    string    `json:"bucket"`
	Key       string    `json:"key"`
	Owner     string    `json:"owner"`
	CreatedAt time.Time `json:"created_at"`
	Parts     []MultipartPart `json:"parts,omitempty"`
}

// MultipartPart represents a single uploaded part.
type MultipartPart struct {
	PartNumber int    `json:"part_number"`
	Size       int64  `json:"size"`
	ETag       string `json:"etag"`
	TempPath   string `json:"temp_path"` // Temporary local storage path
}

// --- Proxy Types ---

// ProxySession tracks an active proxy session.
type ProxySession struct {
	ID            string    `json:"id"`
	WalletAddress string    `json:"wallet_address"`
	Protocol      string    `json:"protocol"` // "socks5", "http", "https"
	Target        string    `json:"target,omitempty"`
	BytesIn       int64     `json:"bytes_in"`
	BytesOut      int64     `json:"bytes_out"`
	StartedAt     time.Time `json:"started_at"`
	ClosedAt      time.Time `json:"closed_at,omitempty"`
	UseTor        bool      `json:"use_tor,omitempty"`
}

// BandwidthReport summarizes bandwidth usage for a wallet.
type BandwidthReport struct {
	WalletAddress string        `json:"wallet_address"`
	TotalBytesIn  int64         `json:"total_bytes_in"`
	TotalBytesOut int64         `json:"total_bytes_out"`
	SessionCount  int           `json:"session_count"`
	Period        time.Duration `json:"period"`
}

// --- AI Agent Types ---

// AgentFramework identifies an AI agent framework.
type AgentFramework string

const (
	AgentFrameworkLangGraph AgentFramework = "langgraph"
	AgentFrameworkCrewAI    AgentFramework = "crewai"
	AgentFrameworkAutoGen   AgentFramework = "autogen"
	AgentFrameworkCustom    AgentFramework = "custom"
)

// AgentSpec defines a deployment specification for an AI agent.
type AgentSpec struct {
	Framework    AgentFramework    `json:"framework"`
	Image        string            `json:"image,omitempty"`        // Container image or CID
	Config       map[string]any    `json:"config,omitempty"`       // Framework-specific config
	EnvVars      map[string]string `json:"env_vars,omitempty"`     // Environment variables
	MCPTools     []string          `json:"mcp_tools,omitempty"`    // Enabled MCP tools
	MemoryBucket string            `json:"memory_bucket,omitempty"` // Object Storage bucket for memory
	SyncInterval time.Duration     `json:"sync_interval,omitempty"` // Memory checkpoint interval
	MaxTokens    int64             `json:"max_tokens,omitempty"`    // Budget cap
}

// AgentDeployment represents a deployed AI agent.
type AgentDeployment struct {
	ID           string         `json:"id"`
	Spec         AgentSpec      `json:"spec"`
	ContainerID  string         `json:"container_id"`
	Status       string         `json:"status"` // pending, running, suspended, stopped
	WalletAddress string        `json:"wallet_address"`
	TokensUsed   int64          `json:"tokens_used"`
	CreatedAt    time.Time      `json:"created_at"`
	LastActivity time.Time      `json:"last_activity"`
}

// MCPToolDef defines an MCP tool available to agents.
type MCPToolDef struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"input_schema"`
}

// --- Web Crawling Types ---

// CrawlJob represents a multi-page crawl job.
type CrawlJob struct {
	ID             string       `json:"id"`
	WalletAddress  string       `json:"wallet_address"`
	Status         string       `json:"status"` // pending, running, completed, failed, cancelled
	Targets        []CrawlTarget `json:"targets"`
	Config         CrawlJobConfig `json:"config"`
	CreatedAt      time.Time    `json:"created_at"`
	CompletedAt    time.Time    `json:"completed_at,omitempty"`
	PagesCompleted int          `json:"pages_completed"`
	PagesFailed    int          `json:"pages_failed"`
	TotalBytes     int64        `json:"total_bytes"`
}

// CrawlTarget is a single URL to crawl.
type CrawlTarget struct {
	URL       string   `json:"url"`
	Selectors []string `json:"selectors,omitempty"` // CSS/XPath selectors
}

// CrawlJobConfig configures crawl behavior.
type CrawlJobConfig struct {
	MaxDepth       int      `json:"max_depth,omitempty"`       // Max link-follow depth
	AllowedDomains []string `json:"allowed_domains,omitempty"` // Domain whitelist
	Screenshot     bool     `json:"screenshot,omitempty"`      // Take screenshots
	JavaScript     bool     `json:"javascript,omitempty"`      // Enable JS execution
	UseTor         bool     `json:"use_tor,omitempty"`         // Route through Tor
	MaxPages       int      `json:"max_pages,omitempty"`       // Max pages per job
}

// CrawlResult is the output of crawling a single page.
type CrawlResult struct {
	URL           string            `json:"url"`
	StatusCode    int               `json:"status_code"`
	ContentType   string            `json:"content_type,omitempty"`
	HTML          string            `json:"html,omitempty"`
	ExtractedText string            `json:"extracted_text,omitempty"`
	Selectors     map[string]string `json:"selectors,omitempty"` // selector → extracted text
	Links         []string          `json:"links,omitempty"`
	ScreenshotCID string            `json:"screenshot_cid,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
	CrawledAt     time.Time         `json:"crawled_at"`
	Duration      time.Duration     `json:"duration"`
	Error         string            `json:"error,omitempty"`
}
