package molt

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/moltbunker/moltbunker/internal/crawl"
)

func TestExecuteCrawl_CreateJob(t *testing.T) {
	ssrfBypass = true
	defer func() { ssrfBypass = false }()

	scheduler := crawl.NewScheduler(crawl.DefaultSchedulerConfig())
	svc := NewHostServices(HostCapabilities{CrawlEnabled: true})
	svc.Crawl = scheduler
	svc.Owner = "0xTestOwner"

	req := &CrawlRequest{
		URL:        "https://example.com",
		Selectors:  []string{"h1", ".content"},
		Screenshot: false,
		JavaScript: true,
	}

	job, err := ExecuteCrawl(context.Background(), svc, req)
	if err != nil {
		t.Fatalf("ExecuteCrawl: %v", err)
	}

	if job.ID == "" {
		t.Fatal("expected non-empty job ID")
	}
	if job.Owner != "0xTestOwner" {
		t.Fatalf("Owner = %q, want %q", job.Owner, "0xTestOwner")
	}
	if len(job.Config.URLs) != 1 || job.Config.URLs[0] != "https://example.com" {
		t.Fatalf("URLs = %v, want [https://example.com]", job.Config.URLs)
	}
	if job.Config.MaxPages != 1 {
		t.Fatalf("MaxPages = %d, want 1 (single page crawl)", job.Config.MaxPages)
	}
}

func TestExecuteCrawl_SSRFBlock(t *testing.T) {
	// SSRF bypass NOT enabled — private IPs should be blocked
	ssrfBypass = false

	scheduler := crawl.NewScheduler(crawl.DefaultSchedulerConfig())
	svc := NewHostServices(HostCapabilities{CrawlEnabled: true})
	svc.Crawl = scheduler
	svc.Owner = "0xTestOwner"

	req := &CrawlRequest{
		URL: "http://127.0.0.1/admin",
	}

	// SSRF check happens in hostCrawlPage before calling ExecuteCrawl,
	// so test the hostFromURL + validateHost chain directly
	host := hostFromURL(req.URL)
	err := validateHost(host)
	if err == nil {
		t.Fatal("expected SSRF block for 127.0.0.1")
	}
}

func TestExecuteCrawl_Disabled(t *testing.T) {
	svc := NewHostServices(HostCapabilities{CrawlEnabled: false})
	if svc.Config.CrawlEnabled {
		t.Fatal("CrawlEnabled should be false")
	}
}

func TestCrawlRequest_JSON(t *testing.T) {
	req := CrawlRequest{
		URL:        "https://example.com/page",
		Selectors:  []string{"h1", "p"},
		Screenshot: true,
		JavaScript: true,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var parsed CrawlRequest
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if parsed.URL != "https://example.com/page" {
		t.Fatalf("URL = %q", parsed.URL)
	}
	if len(parsed.Selectors) != 2 {
		t.Fatalf("Selectors len = %d, want 2", len(parsed.Selectors))
	}
	if !parsed.Screenshot || !parsed.JavaScript {
		t.Fatal("Screenshot/JavaScript should be true")
	}
}

func TestHostFromURL_EdgeCases(t *testing.T) {
	tests := []struct {
		url  string
		want string
	}{
		{"https://example.com", "example.com"},
		{"http://example.com:8080/path", "example.com"},
		{"https://sub.domain.com/path?q=1", "sub.domain.com"},
		{"ftp://evil.com", ""},
		{"not-a-url", ""},
	}

	for _, tt := range tests {
		got := hostFromURL(tt.url)
		if got != tt.want {
			t.Errorf("hostFromURL(%q) = %q, want %q", tt.url, got, tt.want)
		}
	}
}
