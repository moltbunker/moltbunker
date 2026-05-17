package crawl

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestServer() (*RESTHandler, *http.ServeMux) {
	scheduler := NewScheduler(DefaultSchedulerConfig())
	robots := NewRobotsChecker()
	handler := NewRESTHandler(scheduler, robots)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux,
		func(h http.HandlerFunc) http.HandlerFunc { return h },
		func(h http.HandlerFunc) http.HandlerFunc { return h },
	)
	return handler, mux
}

func TestHandler_CreateJob(t *testing.T) {
	_, mux := newTestServer()

	body := `{"urls":["https://example.com"],"max_pages":10}`
	req := httptest.NewRequest("POST", "/v1/crawl/jobs", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Moltbunker-Verified-Wallet", "0xtest")

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body: %s", w.Code, w.Body.String())
	}

	var job CrawlJob
	if err := json.NewDecoder(w.Body).Decode(&job); err != nil {
		t.Fatalf("decode job: %v", err)
	}
	if job.ID == "" {
		t.Error("job ID should not be empty")
	}
	if job.Owner != "0xtest" {
		t.Errorf("owner = %q, want 0xtest", job.Owner)
	}
}

func TestHandler_CreateJob_NoURLs(t *testing.T) {
	_, mux := newTestServer()

	body := `{"urls":[]}`
	req := httptest.NewRequest("POST", "/v1/crawl/jobs", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", w.Code)
	}
}

func TestHandler_CreateJob_InvalidBody(t *testing.T) {
	_, mux := newTestServer()

	req := httptest.NewRequest("POST", "/v1/crawl/jobs", bytes.NewBufferString("{bad"))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", w.Code)
	}
}

func TestHandler_ListJobs(t *testing.T) {
	h, mux := newTestServer()
	if _, err := h.scheduler.CreateJob(nil, "0xwallet1", CrawlConfig{URLs: []string{"https://a.com"}}); err != nil {
		t.Fatalf("CreateJob: %v", err)
	}

	req := httptest.NewRequest("GET", "/v1/crawl/jobs", nil)
	req.Header.Set("X-Moltbunker-Verified-Wallet", "0xwallet1")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	var jobs []CrawlJob
	if err := json.NewDecoder(w.Body).Decode(&jobs); err != nil {
		t.Fatalf("decode jobs: %v", err)
	}
	if len(jobs) != 1 {
		t.Errorf("jobs = %d, want 1", len(jobs))
	}
}

func TestHandler_ListJobs_NoWallet(t *testing.T) {
	_, mux := newTestServer()

	req := httptest.NewRequest("GET", "/v1/crawl/jobs", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", w.Code)
	}
}

func TestHandler_ListJobs_CrossTenantIsolation(t *testing.T) {
	h, mux := newTestServer()
	if _, err := h.scheduler.CreateJob(nil, "0xwallet1", CrawlConfig{URLs: []string{"https://a.com"}}); err != nil {
		t.Fatalf("CreateJob wallet1: %v", err)
	}
	if _, err := h.scheduler.CreateJob(nil, "0xwallet2", CrawlConfig{URLs: []string{"https://b.com"}}); err != nil {
		t.Fatalf("CreateJob wallet2: %v", err)
	}

	// wallet1 should only see their own job
	req := httptest.NewRequest("GET", "/v1/crawl/jobs", nil)
	req.Header.Set("X-Moltbunker-Verified-Wallet", "0xwallet1")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	var jobs []CrawlJob
	if err := json.NewDecoder(w.Body).Decode(&jobs); err != nil {
		t.Fatalf("decode wallet1 jobs: %v", err)
	}
	if len(jobs) != 1 {
		t.Errorf("wallet1 jobs = %d, want 1", len(jobs))
	}

	// wallet2 should only see their own job
	req = httptest.NewRequest("GET", "/v1/crawl/jobs", nil)
	req.Header.Set("X-Moltbunker-Verified-Wallet", "0xwallet2")
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if err := json.NewDecoder(w.Body).Decode(&jobs); err != nil {
		t.Fatalf("decode wallet2 jobs: %v", err)
	}
	if len(jobs) != 1 {
		t.Errorf("wallet2 jobs = %d, want 1", len(jobs))
	}
}

func TestHandler_GetJob(t *testing.T) {
	h, mux := newTestServer()
	job, _ := h.scheduler.CreateJob(nil, "0xowner", CrawlConfig{URLs: []string{"https://a.com"}})

	req := httptest.NewRequest("GET", "/v1/crawl/jobs/"+job.ID, nil)
	req.Header.Set("X-Moltbunker-Verified-Wallet", "0xowner")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	var got CrawlJob
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("decode job: %v", err)
	}
	if got.ID != job.ID {
		t.Errorf("id = %q, want %q", got.ID, job.ID)
	}
}

func TestHandler_GetJob_WrongOwner(t *testing.T) {
	h, mux := newTestServer()
	job, _ := h.scheduler.CreateJob(nil, "0xowner", CrawlConfig{URLs: []string{"https://a.com"}})

	req := httptest.NewRequest("GET", "/v1/crawl/jobs/"+job.ID, nil)
	req.Header.Set("X-Moltbunker-Verified-Wallet", "0xattacker")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404 for wrong owner", w.Code)
	}
}

func TestHandler_GetJob_NotFound(t *testing.T) {
	_, mux := newTestServer()

	req := httptest.NewRequest("GET", "/v1/crawl/jobs/nonexistent", nil)
	req.Header.Set("X-Moltbunker-Verified-Wallet", "0xsomeone")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", w.Code)
	}
}

func TestHandler_GetResults(t *testing.T) {
	h, mux := newTestServer()
	job, _ := h.scheduler.CreateJob(nil, "0xowner", CrawlConfig{URLs: []string{"https://a.com"}})
	if err := h.scheduler.AddResult(job.ID, CrawlResult{URL: "https://a.com", StatusCode: 200}); err != nil {
		t.Fatalf("AddResult: %v", err)
	}

	req := httptest.NewRequest("GET", "/v1/crawl/jobs/"+job.ID+"/results", nil)
	req.Header.Set("X-Moltbunker-Verified-Wallet", "0xowner")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	var results []CrawlResult
	if err := json.NewDecoder(w.Body).Decode(&results); err != nil {
		t.Fatalf("decode results: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("results = %d, want 1", len(results))
	}
}

func TestHandler_CancelJob(t *testing.T) {
	h, mux := newTestServer()
	job, _ := h.scheduler.CreateJob(nil, "0xowner", CrawlConfig{URLs: []string{"https://a.com"}})

	req := httptest.NewRequest("POST", "/v1/crawl/jobs/"+job.ID+"/cancel", nil)
	req.Header.Set("X-Moltbunker-Verified-Wallet", "0xowner")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	got, _ := h.scheduler.GetJob(job.ID)
	if got.Status != JobStatusCancelled {
		t.Errorf("status = %q, want cancelled", got.Status)
	}
}

func TestHandler_CancelJob_NotFound(t *testing.T) {
	_, mux := newTestServer()

	req := httptest.NewRequest("POST", "/v1/crawl/jobs/nonexistent/cancel", nil)
	req.Header.Set("X-Moltbunker-Verified-Wallet", "0xsomeone")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", w.Code)
	}
}

func TestHandler_CrawlPage(t *testing.T) {
	_, mux := newTestServer()

	body := `{"url":"https://example.com","selectors":["h1"]}`
	req := httptest.NewRequest("POST", "/v1/crawl/pages", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Moltbunker-Verified-Wallet", "0xtest")

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want 202; body: %s", w.Code, w.Body.String())
	}
}

func TestHandler_CrawlPage_NoURL(t *testing.T) {
	_, mux := newTestServer()

	body := `{}`
	req := httptest.NewRequest("POST", "/v1/crawl/pages", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", w.Code)
	}
}

func TestHandler_Stats(t *testing.T) {
	h, mux := newTestServer()
	job, _ := h.scheduler.CreateJob(nil, "w1", CrawlConfig{URLs: []string{"https://a.com"}})
	if err := h.scheduler.StartJob(job.ID); err != nil {
		t.Fatalf("StartJob: %v", err)
	}
	if err := h.scheduler.CompleteJob(job.ID); err != nil {
		t.Fatalf("CompleteJob: %v", err)
	}

	req := httptest.NewRequest("GET", "/v1/crawl/stats", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	var stats SchedulerStats
	if err := json.NewDecoder(w.Body).Decode(&stats); err != nil {
		t.Fatalf("decode stats: %v", err)
	}
	if stats.TotalJobs != 1 {
		t.Errorf("total = %d, want 1", stats.TotalJobs)
	}
}

func TestHandler_AnonymousOwner(t *testing.T) {
	_, mux := newTestServer()

	body := `{"urls":["https://example.com"]}`
	req := httptest.NewRequest("POST", "/v1/crawl/jobs", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201", w.Code)
	}

	var job CrawlJob
	json.NewDecoder(w.Body).Decode(&job)
	if job.Owner != "anonymous" {
		t.Errorf("owner = %q, want anonymous", job.Owner)
	}
}

func TestHandler_MethodNotAllowed(t *testing.T) {
	_, mux := newTestServer()

	req := httptest.NewRequest("DELETE", "/v1/crawl/jobs", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want 405", w.Code)
	}
}
