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
	json.NewDecoder(w.Body).Decode(&job)
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
	h.scheduler.CreateJob(nil, "w1", CrawlConfig{URLs: []string{"https://a.com"}})

	req := httptest.NewRequest("GET", "/v1/crawl/jobs", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	var jobs []CrawlJob
	json.NewDecoder(w.Body).Decode(&jobs)
	if len(jobs) != 1 {
		t.Errorf("jobs = %d, want 1", len(jobs))
	}
}

func TestHandler_ListJobs_FilterByWallet(t *testing.T) {
	h, mux := newTestServer()
	h.scheduler.CreateJob(nil, "w1", CrawlConfig{URLs: []string{"https://a.com"}})
	h.scheduler.CreateJob(nil, "w2", CrawlConfig{URLs: []string{"https://b.com"}})

	req := httptest.NewRequest("GET", "/v1/crawl/jobs?wallet=w1", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	var jobs []CrawlJob
	json.NewDecoder(w.Body).Decode(&jobs)
	if len(jobs) != 1 {
		t.Errorf("jobs = %d, want 1 for w1", len(jobs))
	}
}

func TestHandler_GetJob(t *testing.T) {
	h, mux := newTestServer()
	job, _ := h.scheduler.CreateJob(nil, "w1", CrawlConfig{URLs: []string{"https://a.com"}})

	req := httptest.NewRequest("GET", "/v1/crawl/jobs/"+job.ID, nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	var got CrawlJob
	json.NewDecoder(w.Body).Decode(&got)
	if got.ID != job.ID {
		t.Errorf("id = %q, want %q", got.ID, job.ID)
	}
}

func TestHandler_GetJob_NotFound(t *testing.T) {
	_, mux := newTestServer()

	req := httptest.NewRequest("GET", "/v1/crawl/jobs/nonexistent", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", w.Code)
	}
}

func TestHandler_GetResults(t *testing.T) {
	h, mux := newTestServer()
	job, _ := h.scheduler.CreateJob(nil, "w1", CrawlConfig{URLs: []string{"https://a.com"}})
	h.scheduler.AddResult(job.ID, CrawlResult{URL: "https://a.com", StatusCode: 200})

	req := httptest.NewRequest("GET", "/v1/crawl/jobs/"+job.ID+"/results", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	var results []CrawlResult
	json.NewDecoder(w.Body).Decode(&results)
	if len(results) != 1 {
		t.Errorf("results = %d, want 1", len(results))
	}
}

func TestHandler_CancelJob(t *testing.T) {
	h, mux := newTestServer()
	job, _ := h.scheduler.CreateJob(nil, "w1", CrawlConfig{URLs: []string{"https://a.com"}})

	req := httptest.NewRequest("POST", "/v1/crawl/jobs/"+job.ID+"/cancel", nil)
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
	h.scheduler.StartJob(job.ID)
	h.scheduler.CompleteJob(job.ID)

	req := httptest.NewRequest("GET", "/v1/crawl/stats", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	var stats SchedulerStats
	json.NewDecoder(w.Body).Decode(&stats)
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
