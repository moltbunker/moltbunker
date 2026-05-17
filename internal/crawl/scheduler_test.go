package crawl

import (
	"context"
	"testing"
)

func TestScheduler_CreateJob(t *testing.T) {
	s := NewScheduler(DefaultSchedulerConfig())

	job, err := s.CreateJob(context.Background(), "wallet-1", CrawlConfig{
		URLs: []string{"https://example.com"},
	})
	if err != nil {
		t.Fatalf("CreateJob: %v", err)
	}
	if job.ID == "" {
		t.Error("job ID should not be empty")
	}
	if job.Owner != "wallet-1" {
		t.Errorf("owner = %q, want wallet-1", job.Owner)
	}
	if job.Status != JobStatusPending {
		t.Errorf("status = %q, want pending", job.Status)
	}
	if job.Config.UserAgent != "MoltbunkerCrawler/1.0" {
		t.Errorf("user agent = %q, want default", job.Config.UserAgent)
	}
}

func TestScheduler_CreateJob_NoOwner(t *testing.T) {
	s := NewScheduler(DefaultSchedulerConfig())
	_, err := s.CreateJob(context.Background(), "", CrawlConfig{
		URLs: []string{"https://example.com"},
	})
	if err == nil {
		t.Error("expected error for empty owner")
	}
}

func TestScheduler_CreateJob_NoURLs(t *testing.T) {
	s := NewScheduler(DefaultSchedulerConfig())
	_, err := s.CreateJob(context.Background(), "w1", CrawlConfig{})
	if err == nil {
		t.Error("expected error for no URLs")
	}
}

func TestScheduler_CreateJob_InvalidURL(t *testing.T) {
	s := NewScheduler(DefaultSchedulerConfig())
	_, err := s.CreateJob(context.Background(), "w1", CrawlConfig{
		URLs: []string{"ftp://bad.com"},
	})
	if err == nil {
		t.Error("expected error for non-http scheme")
	}
}

func TestScheduler_CreateJob_WalletLimit(t *testing.T) {
	cfg := DefaultSchedulerConfig()
	cfg.MaxJobsPerWallet = 2
	s := NewScheduler(cfg)

	if _, err := s.CreateJob(context.Background(), "w1", CrawlConfig{URLs: []string{"https://a.com"}}); err != nil {
		t.Fatalf("CreateJob a: %v", err)
	}
	if _, err := s.CreateJob(context.Background(), "w1", CrawlConfig{URLs: []string{"https://b.com"}}); err != nil {
		t.Fatalf("CreateJob b: %v", err)
	}

	_, err := s.CreateJob(context.Background(), "w1", CrawlConfig{URLs: []string{"https://c.com"}})
	if err == nil {
		t.Error("expected job limit error")
	}
}

func TestScheduler_CreateJob_MaxPagesDefault(t *testing.T) {
	s := NewScheduler(DefaultSchedulerConfig())
	job, _ := s.CreateJob(context.Background(), "w1", CrawlConfig{
		URLs: []string{"https://example.com"},
	})
	if job.Config.MaxPages != 100 {
		t.Errorf("max pages = %d, want 100", job.Config.MaxPages)
	}
}

func TestScheduler_CreateJob_MaxPagesCapped(t *testing.T) {
	cfg := DefaultSchedulerConfig()
	cfg.MaxPagesPerJob = 500
	s := NewScheduler(cfg)

	job, _ := s.CreateJob(context.Background(), "w1", CrawlConfig{
		URLs:     []string{"https://example.com"},
		MaxPages: 9999,
	})
	if job.Config.MaxPages != 500 {
		t.Errorf("max pages = %d, want 500 (capped)", job.Config.MaxPages)
	}
}

func TestScheduler_StartJob(t *testing.T) {
	s := NewScheduler(DefaultSchedulerConfig())
	job, _ := s.CreateJob(context.Background(), "w1", CrawlConfig{
		URLs: []string{"https://example.com"},
	})

	if err := s.StartJob(job.ID); err != nil {
		t.Fatalf("StartJob: %v", err)
	}

	got, _ := s.GetJob(job.ID)
	if got.Status != JobStatusRunning {
		t.Errorf("status = %q, want running", got.Status)
	}
	if got.StartedAt.IsZero() {
		t.Error("started_at should be set")
	}
}

func TestScheduler_StartJob_NotFound(t *testing.T) {
	s := NewScheduler(DefaultSchedulerConfig())
	if err := s.StartJob("nonexistent"); err == nil {
		t.Error("expected error for missing job")
	}
}

func TestScheduler_StartJob_AlreadyRunning(t *testing.T) {
	s := NewScheduler(DefaultSchedulerConfig())
	job, _ := s.CreateJob(context.Background(), "w1", CrawlConfig{
		URLs: []string{"https://example.com"},
	})
	if err := s.StartJob(job.ID); err != nil {
		t.Fatalf("StartJob: %v", err)
	}

	if err := s.StartJob(job.ID); err == nil {
		t.Error("expected error for already-running job")
	}
}

func TestScheduler_AddResult(t *testing.T) {
	s := NewScheduler(DefaultSchedulerConfig())
	job, _ := s.CreateJob(context.Background(), "w1", CrawlConfig{
		URLs: []string{"https://example.com"},
	})

	err := s.AddResult(job.ID, CrawlResult{
		URL:        "https://example.com",
		StatusCode: 200,
		ByteSize:   1024,
	})
	if err != nil {
		t.Fatalf("AddResult: %v", err)
	}

	got, _ := s.GetJob(job.ID)
	if got.PagesCrawled != 1 {
		t.Errorf("pages = %d, want 1", got.PagesCrawled)
	}
	if got.TotalBytes != 1024 {
		t.Errorf("bytes = %d, want 1024", got.TotalBytes)
	}
}

func TestScheduler_CompleteJob(t *testing.T) {
	s := NewScheduler(DefaultSchedulerConfig())
	job, _ := s.CreateJob(context.Background(), "w1", CrawlConfig{
		URLs: []string{"https://example.com"},
	})
	if err := s.StartJob(job.ID); err != nil {
		t.Fatalf("StartJob: %v", err)
	}
	if err := s.CompleteJob(job.ID); err != nil {
		t.Fatalf("CompleteJob: %v", err)
	}

	got, _ := s.GetJob(job.ID)
	if got.Status != JobStatusCompleted {
		t.Errorf("status = %q, want completed", got.Status)
	}
}

func TestScheduler_FailJob(t *testing.T) {
	s := NewScheduler(DefaultSchedulerConfig())
	job, _ := s.CreateJob(context.Background(), "w1", CrawlConfig{
		URLs: []string{"https://example.com"},
	})
	if err := s.StartJob(job.ID); err != nil {
		t.Fatalf("StartJob: %v", err)
	}
	if err := s.FailJob(job.ID, "timeout"); err != nil {
		t.Fatalf("FailJob: %v", err)
	}

	got, _ := s.GetJob(job.ID)
	if got.Status != JobStatusFailed {
		t.Errorf("status = %q, want failed", got.Status)
	}
	if got.Error != "timeout" {
		t.Errorf("error = %q, want timeout", got.Error)
	}
}

func TestScheduler_CancelJob(t *testing.T) {
	s := NewScheduler(DefaultSchedulerConfig())
	job, _ := s.CreateJob(context.Background(), "w1", CrawlConfig{
		URLs: []string{"https://example.com"},
	})

	if err := s.CancelJob(job.ID); err != nil {
		t.Fatalf("CancelJob: %v", err)
	}

	got, _ := s.GetJob(job.ID)
	if got.Status != JobStatusCancelled {
		t.Errorf("status = %q, want cancelled", got.Status)
	}
}

func TestScheduler_CancelJob_AlreadyCompleted(t *testing.T) {
	s := NewScheduler(DefaultSchedulerConfig())
	job, _ := s.CreateJob(context.Background(), "w1", CrawlConfig{
		URLs: []string{"https://example.com"},
	})
	if err := s.StartJob(job.ID); err != nil {
		t.Fatalf("StartJob: %v", err)
	}
	if err := s.CompleteJob(job.ID); err != nil {
		t.Fatalf("CompleteJob: %v", err)
	}

	if err := s.CancelJob(job.ID); err == nil {
		t.Error("expected error for completed job")
	}
}

func TestScheduler_GetJob_DeepCopy(t *testing.T) {
	s := NewScheduler(DefaultSchedulerConfig())
	job, _ := s.CreateJob(context.Background(), "w1", CrawlConfig{
		URLs: []string{"https://example.com"},
	})
	if err := s.AddResult(job.ID, CrawlResult{URL: "https://example.com", StatusCode: 200}); err != nil {
		t.Fatalf("AddResult: %v", err)
	}

	got, _ := s.GetJob(job.ID)
	got.Results = append(got.Results, CrawlResult{URL: "mutated"})

	original, _ := s.GetJob(job.ID)
	if len(original.Results) != 1 {
		t.Errorf("mutation leaked: results = %d, want 1", len(original.Results))
	}
}

func TestScheduler_ListJobs(t *testing.T) {
	s := NewScheduler(DefaultSchedulerConfig())
	if _, err := s.CreateJob(context.Background(), "w1", CrawlConfig{URLs: []string{"https://a.com"}}); err != nil {
		t.Fatalf("CreateJob w1: %v", err)
	}
	if _, err := s.CreateJob(context.Background(), "w2", CrawlConfig{URLs: []string{"https://b.com"}}); err != nil {
		t.Fatalf("CreateJob w2: %v", err)
	}

	all := s.ListJobs("")
	if len(all) != 2 {
		t.Errorf("all jobs = %d, want 2", len(all))
	}

	w1 := s.ListJobs("w1")
	if len(w1) != 1 {
		t.Errorf("w1 jobs = %d, want 1", len(w1))
	}
}

func TestScheduler_ListJobs_NoResults(t *testing.T) {
	s := NewScheduler(DefaultSchedulerConfig())
	job, _ := s.CreateJob(context.Background(), "w1", CrawlConfig{
		URLs: []string{"https://example.com"},
	})
	if err := s.AddResult(job.ID, CrawlResult{URL: "https://example.com"}); err != nil {
		t.Fatalf("AddResult: %v", err)
	}

	jobs := s.ListJobs("w1")
	if len(jobs) != 1 {
		t.Fatalf("jobs = %d, want 1", len(jobs))
	}
	if jobs[0].Results != nil {
		t.Error("list should not include results")
	}
}

func TestScheduler_GetResults(t *testing.T) {
	s := NewScheduler(DefaultSchedulerConfig())
	job, _ := s.CreateJob(context.Background(), "w1", CrawlConfig{
		URLs: []string{"https://example.com"},
	})
	if err := s.AddResult(job.ID, CrawlResult{URL: "https://example.com", StatusCode: 200}); err != nil {
		t.Fatalf("AddResult: %v", err)
	}

	results, err := s.GetResults(job.ID)
	if err != nil {
		t.Fatalf("GetResults: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("results = %d, want 1", len(results))
	}
}

func TestScheduler_ShouldCrawl(t *testing.T) {
	s := NewScheduler(DefaultSchedulerConfig())

	cfg := CrawlConfig{MaxDepth: 2}

	// First visit — should crawl
	if !s.ShouldCrawl("job1", "https://example.com", 0, cfg) {
		t.Error("first visit should be allowed")
	}

	// Duplicate — should not crawl
	if s.ShouldCrawl("job1", "https://example.com", 0, cfg) {
		t.Error("duplicate should be rejected")
	}

	// Different job — should crawl
	if !s.ShouldCrawl("job2", "https://example.com", 0, cfg) {
		t.Error("different job should be allowed")
	}
}

func TestScheduler_ShouldCrawl_DepthLimit(t *testing.T) {
	s := NewScheduler(DefaultSchedulerConfig())
	cfg := CrawlConfig{MaxDepth: 1}

	if s.ShouldCrawl("job1", "https://example.com/deep", 2, cfg) {
		t.Error("should reject depth > max")
	}
}

func TestScheduler_ShouldCrawl_AllowedDomains(t *testing.T) {
	s := NewScheduler(DefaultSchedulerConfig())
	cfg := CrawlConfig{
		MaxDepth:       5,
		AllowedDomains: []string{"example.com"},
	}

	if !s.ShouldCrawl("job1", "https://example.com/page", 0, cfg) {
		t.Error("allowed domain should pass")
	}

	if !s.ShouldCrawl("job1", "https://www.example.com/page", 0, cfg) {
		t.Error("www prefix should be allowed")
	}

	if s.ShouldCrawl("job1", "https://other.com/page", 0, cfg) {
		t.Error("non-allowed domain should be rejected")
	}
}

func TestScheduler_CheckRateLimit(t *testing.T) {
	s := NewScheduler(DefaultSchedulerConfig())

	if !s.CheckRateLimit("example.com") {
		t.Error("first request should be allowed")
	}

	// Immediate second request should be rejected (1s interval)
	if s.CheckRateLimit("example.com") {
		t.Error("immediate second request should be rate limited")
	}

	// Different domain should be allowed
	if !s.CheckRateLimit("other.com") {
		t.Error("different domain should be allowed")
	}
}

func TestScheduler_Stats(t *testing.T) {
	s := NewScheduler(DefaultSchedulerConfig())
	job1, _ := s.CreateJob(context.Background(), "w1", CrawlConfig{
		URLs: []string{"https://example.com"},
	})
	if err := s.StartJob(job1.ID); err != nil {
		t.Fatalf("StartJob job1: %v", err)
	}
	if err := s.AddResult(job1.ID, CrawlResult{URL: "https://example.com", ByteSize: 500}); err != nil {
		t.Fatalf("AddResult job1: %v", err)
	}
	if err := s.CompleteJob(job1.ID); err != nil {
		t.Fatalf("CompleteJob job1: %v", err)
	}

	job2, _ := s.CreateJob(context.Background(), "w1", CrawlConfig{
		URLs: []string{"https://other.com"},
	})
	if err := s.StartJob(job2.ID); err != nil {
		t.Fatalf("StartJob job2: %v", err)
	}
	if err := s.FailJob(job2.ID, "error"); err != nil {
		t.Fatalf("FailJob job2: %v", err)
	}

	stats := s.Stats()
	if stats.TotalJobs != 2 {
		t.Errorf("total = %d, want 2", stats.TotalJobs)
	}
	if stats.CompletedJobs != 1 {
		t.Errorf("completed = %d, want 1", stats.CompletedJobs)
	}
	if stats.FailedJobs != 1 {
		t.Errorf("failed = %d, want 1", stats.FailedJobs)
	}
	if stats.TotalPagesCrawled != 1 {
		t.Errorf("pages = %d, want 1", stats.TotalPagesCrawled)
	}
	if stats.TotalBytes != 500 {
		t.Errorf("bytes = %d, want 500", stats.TotalBytes)
	}
}
