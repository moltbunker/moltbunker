package crawl

import (
	"time"
)

// JobStatus tracks the lifecycle of a crawl job.
type JobStatus string

const (
	JobStatusPending    JobStatus = "pending"
	JobStatusRunning    JobStatus = "running"
	JobStatusCompleted  JobStatus = "completed"
	JobStatusFailed     JobStatus = "failed"
	JobStatusCancelled  JobStatus = "cancelled"
)

// CrawlJob describes a crawl task (single page or multi-page).
type CrawlJob struct {
	ID           string            `json:"id"`
	Owner        string            `json:"owner"`
	Status       JobStatus         `json:"status"`
	Config       CrawlConfig       `json:"config"`
	CreatedAt    time.Time         `json:"created_at"`
	StartedAt    time.Time         `json:"started_at,omitempty"`
	CompletedAt  time.Time         `json:"completed_at,omitempty"`
	Error        string            `json:"error,omitempty"`
	PagesCrawled int               `json:"pages_crawled"`
	TotalBytes   int64             `json:"total_bytes"`
	Results      []CrawlResult     `json:"results,omitempty"`
}

// CrawlConfig configures a crawl job.
type CrawlConfig struct {
	URLs            []string          `json:"urls"`
	MaxDepth        int               `json:"max_depth,omitempty"`        // 0 = single page, >0 = follow links
	MaxPages        int               `json:"max_pages,omitempty"`        // Max pages to crawl (default: 100)
	AllowedDomains  []string          `json:"allowed_domains,omitempty"`  // Only follow links to these domains
	Selectors       []string          `json:"selectors,omitempty"`        // CSS selectors to extract
	Screenshot      bool              `json:"screenshot,omitempty"`       // Capture screenshot
	JavaScript      bool              `json:"javascript,omitempty"`       // Enable JS execution
	UserAgent       string            `json:"user_agent,omitempty"`
	Headers         map[string]string `json:"headers,omitempty"`
	TimeoutSec      int               `json:"timeout_sec,omitempty"`      // Per-page timeout
	RespectRobots   bool              `json:"respect_robots,omitempty"`   // Obey robots.txt
	UseTor          bool              `json:"use_tor,omitempty"`          // Route through Tor
	StorageBucket   string            `json:"storage_bucket,omitempty"`   // Store results in Object Storage
}

// DefaultCrawlConfig returns sensible defaults.
func DefaultCrawlConfig() CrawlConfig {
	return CrawlConfig{
		MaxDepth:      0,
		MaxPages:      100,
		TimeoutSec:    30,
		RespectRobots: true,
		UserAgent:     "MoltbunkerCrawler/1.0",
	}
}

// CrawlResult holds the extracted content from a single page.
type CrawlResult struct {
	URL           string            `json:"url"`
	StatusCode    int               `json:"status_code"`
	ContentType   string            `json:"content_type,omitempty"`
	Title         string            `json:"title,omitempty"`
	HTML          string            `json:"html,omitempty"`
	Text          string            `json:"text,omitempty"`
	Links         []string          `json:"links,omitempty"`
	Selectors     map[string]string `json:"selectors,omitempty"`  // selector → extracted text
	ScreenshotCID string            `json:"screenshot_cid,omitempty"`
	CrawledAt     time.Time         `json:"crawled_at"`
	DurationMs    int64             `json:"duration_ms"`
	Error         string            `json:"error,omitempty"`
	ByteSize      int64             `json:"byte_size"`
}

// CrawlTarget identifies a single page to crawl within a job.
type CrawlTarget struct {
	URL   string `json:"url"`
	Depth int    `json:"depth"`
}

// RobotsRule represents a parsed robots.txt directive.
type RobotsRule struct {
	UserAgent  string   `json:"user_agent"`
	Disallow   []string `json:"disallow"`
	Allow      []string `json:"allow"`
	CrawlDelay int      `json:"crawl_delay"` // seconds
}
