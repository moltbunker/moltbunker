package crawl

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/moltbunker/moltbunker/internal/logging"
)

// Scheduler manages crawl jobs and distributes work.
type Scheduler struct {
	mu   sync.RWMutex
	jobs map[string]*CrawlJob

	config    SchedulerConfig
	rateLimit *DomainRateLimiter
	dedup     *URLDedup
}

// SchedulerConfig configures the crawl scheduler.
type SchedulerConfig struct {
	MaxConcurrentJobs   int           // Max jobs running at once (default: 10)
	MaxJobsPerWallet    int           // Max jobs per wallet (default: 5)
	DefaultPageTimeout  time.Duration // Default per-page timeout
	MaxPagesPerJob      int           // Hard cap on pages per job
}

// DefaultSchedulerConfig returns sensible defaults.
func DefaultSchedulerConfig() SchedulerConfig {
	return SchedulerConfig{
		MaxConcurrentJobs:  10,
		MaxJobsPerWallet:   5,
		DefaultPageTimeout: 30 * time.Second,
		MaxPagesPerJob:     1000,
	}
}

// NewScheduler creates a new crawl scheduler.
func NewScheduler(cfg SchedulerConfig) *Scheduler {
	return &Scheduler{
		jobs:      make(map[string]*CrawlJob),
		config:    cfg,
		rateLimit: NewDomainRateLimiter(time.Second), // 1 req/sec per domain
		dedup:     NewURLDedup(),
	}
}

// CreateJob creates a new crawl job from the given config.
func (s *Scheduler) CreateJob(ctx context.Context, owner string, cfg CrawlConfig) (*CrawlJob, error) {
	if owner == "" {
		return nil, fmt.Errorf("owner is required")
	}
	if len(cfg.URLs) == 0 {
		return nil, fmt.Errorf("at least one URL is required")
	}

	// Validate URLs
	for _, u := range cfg.URLs {
		parsed, err := url.Parse(u)
		if err != nil {
			return nil, fmt.Errorf("invalid URL %q: %w", u, err)
		}
		if parsed.Scheme != "http" && parsed.Scheme != "https" {
			return nil, fmt.Errorf("URL %q must use http or https scheme", u)
		}
	}

	// Check wallet limit
	s.mu.RLock()
	activeCount := 0
	for _, j := range s.jobs {
		if j.Owner == owner && (j.Status == JobStatusPending || j.Status == JobStatusRunning) {
			activeCount++
		}
	}
	s.mu.RUnlock()

	if activeCount >= s.config.MaxJobsPerWallet {
		return nil, fmt.Errorf("job limit exceeded: max %d active jobs per wallet", s.config.MaxJobsPerWallet)
	}

	// Apply defaults
	if cfg.MaxPages <= 0 {
		cfg.MaxPages = 100
	}
	if cfg.MaxPages > s.config.MaxPagesPerJob {
		cfg.MaxPages = s.config.MaxPagesPerJob
	}
	if cfg.TimeoutSec <= 0 {
		cfg.TimeoutSec = int(s.config.DefaultPageTimeout.Seconds())
	}
	if cfg.UserAgent == "" {
		cfg.UserAgent = "MoltbunkerCrawler/1.0"
	}

	id, err := generateJobID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate job ID: %w", err)
	}
	job := &CrawlJob{
		ID:        id,
		Owner:     owner,
		Status:    JobStatusPending,
		Config:    cfg,
		CreatedAt: time.Now(),
	}

	s.mu.Lock()
	s.jobs[id] = job
	s.mu.Unlock()

	logging.Info("crawl job created",
		"job_id", id,
		"urls", len(cfg.URLs),
		"owner", owner,
		logging.Component("crawl"))

	return job, nil
}

// StartJob transitions a job from pending to running.
func (s *Scheduler) StartJob(jobID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	job, ok := s.jobs[jobID]
	if !ok {
		return fmt.Errorf("job %q not found", jobID)
	}
	if job.Status != JobStatusPending {
		return fmt.Errorf("job %q is %s, cannot start", jobID, job.Status)
	}

	job.Status = JobStatusRunning
	job.StartedAt = time.Now()
	return nil
}

// AddResult records a crawled page result.
func (s *Scheduler) AddResult(jobID string, result CrawlResult) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	job, ok := s.jobs[jobID]
	if !ok {
		return fmt.Errorf("job %q not found", jobID)
	}

	job.Results = append(job.Results, result)
	job.PagesCrawled++
	job.TotalBytes += result.ByteSize
	return nil
}

// CompleteJob marks a job as completed.
func (s *Scheduler) CompleteJob(jobID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	job, ok := s.jobs[jobID]
	if !ok {
		return fmt.Errorf("job %q not found", jobID)
	}

	job.Status = JobStatusCompleted
	job.CompletedAt = time.Now()
	return nil
}

// FailJob marks a job as failed.
func (s *Scheduler) FailJob(jobID, errMsg string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	job, ok := s.jobs[jobID]
	if !ok {
		return fmt.Errorf("job %q not found", jobID)
	}

	job.Status = JobStatusFailed
	job.Error = errMsg
	job.CompletedAt = time.Now()
	return nil
}

// CancelJob cancels a running or pending job.
func (s *Scheduler) CancelJob(jobID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	job, ok := s.jobs[jobID]
	if !ok {
		return fmt.Errorf("job %q not found", jobID)
	}

	if job.Status == JobStatusCompleted || job.Status == JobStatusFailed {
		return fmt.Errorf("job %q already finished", jobID)
	}

	job.Status = JobStatusCancelled
	job.CompletedAt = time.Now()
	return nil
}

// GetJob returns a copy of a job.
func (s *Scheduler) GetJob(jobID string) (*CrawlJob, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	job, ok := s.jobs[jobID]
	if !ok {
		return nil, false
	}

	cp := *job
	cp.Results = make([]CrawlResult, len(job.Results))
	copy(cp.Results, job.Results)
	return &cp, true
}

// ListJobs returns jobs for a wallet, or all if wallet is empty.
func (s *Scheduler) ListJobs(wallet string) []CrawlJob {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []CrawlJob
	for _, j := range s.jobs {
		if wallet == "" || j.Owner == wallet {
			cp := *j
			cp.Results = nil // Don't include full results in list
			result = append(result, cp)
		}
	}
	return result
}

// GetResults returns the results for a job.
func (s *Scheduler) GetResults(jobID string) ([]CrawlResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	job, ok := s.jobs[jobID]
	if !ok {
		return nil, fmt.Errorf("job %q not found", jobID)
	}

	results := make([]CrawlResult, len(job.Results))
	copy(results, job.Results)
	return results, nil
}

// ShouldCrawl checks if a URL should be crawled (dedup + domain check).
func (s *Scheduler) ShouldCrawl(jobID, targetURL string, depth int, cfg CrawlConfig) bool {
	if depth > cfg.MaxDepth {
		return false
	}

	if s.dedup.Seen(jobID, targetURL) {
		return false
	}

	if len(cfg.AllowedDomains) > 0 {
		parsed, err := url.Parse(targetURL)
		if err != nil {
			return false
		}
		host := strings.TrimPrefix(parsed.Hostname(), "www.")
		allowed := false
		for _, d := range cfg.AllowedDomains {
			if host == d || host == "www."+d {
				allowed = true
				break
			}
		}
		if !allowed {
			return false
		}
	}

	s.dedup.Mark(jobID, targetURL)
	return true
}

// CheckRateLimit returns true if we can crawl this domain now.
func (s *Scheduler) CheckRateLimit(domain string) bool {
	return s.rateLimit.Allow(domain)
}

// Stats returns scheduler statistics.
func (s *Scheduler) Stats() SchedulerStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := SchedulerStats{}
	for _, j := range s.jobs {
		stats.TotalJobs++
		switch j.Status {
		case JobStatusRunning:
			stats.RunningJobs++
		case JobStatusCompleted:
			stats.CompletedJobs++
		case JobStatusFailed:
			stats.FailedJobs++
		}
		stats.TotalPagesCrawled += j.PagesCrawled
		stats.TotalBytes += j.TotalBytes
	}
	return stats
}

// SchedulerStats provides aggregate crawl metrics.
type SchedulerStats struct {
	TotalJobs         int   `json:"total_jobs"`
	RunningJobs       int   `json:"running_jobs"`
	CompletedJobs     int   `json:"completed_jobs"`
	FailedJobs        int   `json:"failed_jobs"`
	TotalPagesCrawled int   `json:"total_pages_crawled"`
	TotalBytes        int64 `json:"total_bytes"`
}

func generateJobID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate job ID: %w", err)
	}
	return "crawl-" + hex.EncodeToString(b)[:12], nil
}
