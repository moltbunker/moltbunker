package crawl

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/moltbunker/moltbunker/internal/logging"
)

// RESTHandler serves the JSON REST API for crawl management under /v1/crawl/.
type RESTHandler struct {
	scheduler *Scheduler
	robots    *RobotsChecker
}

// NewRESTHandler creates a new crawl REST handler.
func NewRESTHandler(scheduler *Scheduler, robots *RobotsChecker) *RESTHandler {
	return &RESTHandler{
		scheduler: scheduler,
		robots:    robots,
	}
}

// RegisterRoutes registers crawl REST routes on the given mux.
func (h *RESTHandler) RegisterRoutes(mux *http.ServeMux, wrapRead, wrapWrite func(http.HandlerFunc) http.HandlerFunc) {
	mux.HandleFunc("/v1/crawl/jobs", wrapWrite(h.handleJobs))
	mux.HandleFunc("/v1/crawl/jobs/", wrapWrite(h.handleJobByID))
	mux.HandleFunc("/v1/crawl/pages", wrapWrite(h.handleCrawlPage))
	mux.HandleFunc("/v1/crawl/stats", wrapRead(h.handleStats))
}

// createJobRequest is the JSON body for creating a crawl job.
type createJobRequest struct {
	URLs           []string          `json:"urls"`
	MaxDepth       int               `json:"max_depth,omitempty"`
	MaxPages       int               `json:"max_pages,omitempty"`
	AllowedDomains []string          `json:"allowed_domains,omitempty"`
	Selectors      []string          `json:"selectors,omitempty"`
	Screenshot     bool              `json:"screenshot,omitempty"`
	JavaScript     bool              `json:"javascript,omitempty"`
	UserAgent      string            `json:"user_agent,omitempty"`
	Headers        map[string]string `json:"headers,omitempty"`
	TimeoutSec     int               `json:"timeout_sec,omitempty"`
	RespectRobots  bool              `json:"respect_robots,omitempty"`
	UseTor         bool              `json:"use_tor,omitempty"`
	StorageBucket  string            `json:"storage_bucket,omitempty"`
}

// crawlPageRequest is the JSON body for a sync single-page crawl.
type crawlPageRequest struct {
	URL        string            `json:"url"`
	Selectors  []string          `json:"selectors,omitempty"`
	Screenshot bool              `json:"screenshot,omitempty"`
	JavaScript bool              `json:"javascript,omitempty"`
	UserAgent  string            `json:"user_agent,omitempty"`
	Headers    map[string]string `json:"headers,omitempty"`
	TimeoutSec int               `json:"timeout_sec,omitempty"`
}

func (h *RESTHandler) handleJobs(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.createJob(w, r)
	case http.MethodGet:
		h.listJobs(w, r)
	default:
		w.Header().Set("Allow", "GET, POST")
		writeCrawlError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *RESTHandler) handleJobByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/v1/crawl/jobs/")
	if path == "" {
		writeCrawlError(w, http.StatusBadRequest, "job ID required")
		return
	}

	// Check for sub-resources: /v1/crawl/jobs/{id}/results, /v1/crawl/jobs/{id}/cancel
	parts := strings.SplitN(path, "/", 2)
	jobID := parts[0]

	if len(parts) == 2 {
		switch parts[1] {
		case "results":
			h.getResults(w, r, jobID)
			return
		case "cancel":
			h.cancelJob(w, r, jobID)
			return
		default:
			writeCrawlError(w, http.StatusNotFound, "unknown sub-resource")
			return
		}
	}

	// /v1/crawl/jobs/{id}
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		writeCrawlError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	h.getJob(w, r, jobID)
}

func (h *RESTHandler) createJob(w http.ResponseWriter, r *http.Request) {
	var req createJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeCrawlError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if len(req.URLs) == 0 {
		writeCrawlError(w, http.StatusBadRequest, "urls is required")
		return
	}

	owner := r.Header.Get("X-Moltbunker-Verified-Wallet")
	if owner == "" {
		owner = "anonymous"
	}

	cfg := CrawlConfig{
		URLs:           req.URLs,
		MaxDepth:       req.MaxDepth,
		MaxPages:       req.MaxPages,
		AllowedDomains: req.AllowedDomains,
		Selectors:      req.Selectors,
		Screenshot:     req.Screenshot,
		JavaScript:     req.JavaScript,
		UserAgent:      req.UserAgent,
		Headers:        req.Headers,
		TimeoutSec:     req.TimeoutSec,
		RespectRobots:  req.RespectRobots,
		UseTor:         req.UseTor,
		StorageBucket:  req.StorageBucket,
	}

	job, err := h.scheduler.CreateJob(r.Context(), owner, cfg)
	if err != nil {
		if strings.Contains(err.Error(), "job limit exceeded") {
			writeCrawlError(w, http.StatusTooManyRequests, err.Error())
			return
		}
		writeCrawlError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeCrawlJSON(w, http.StatusCreated, job)
}

func (h *RESTHandler) listJobs(w http.ResponseWriter, r *http.Request) {
	wallet := r.Header.Get("X-Moltbunker-Verified-Wallet")
	if wallet == "" {
		writeCrawlError(w, http.StatusForbidden, "no verified identity")
		return
	}
	jobs := h.scheduler.ListJobs(wallet)
	if jobs == nil {
		jobs = []CrawlJob{}
	}
	writeCrawlJSON(w, http.StatusOK, jobs)
}

func (h *RESTHandler) getJob(w http.ResponseWriter, r *http.Request, jobID string) {
	wallet := r.Header.Get("X-Moltbunker-Verified-Wallet")
	if wallet == "" {
		writeCrawlError(w, http.StatusForbidden, "no verified identity")
		return
	}
	job, ok := h.scheduler.GetJob(jobID)
	if !ok || job.Owner != wallet {
		writeCrawlError(w, http.StatusNotFound, "job not found")
		return
	}
	writeCrawlJSON(w, http.StatusOK, job)
}

func (h *RESTHandler) getResults(w http.ResponseWriter, r *http.Request, jobID string) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		writeCrawlError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// Verify ownership
	wallet := r.Header.Get("X-Moltbunker-Verified-Wallet")
	if wallet == "" {
		writeCrawlError(w, http.StatusForbidden, "no verified identity")
		return
	}
	job, ok := h.scheduler.GetJob(jobID)
	if !ok || job.Owner != wallet {
		writeCrawlError(w, http.StatusNotFound, "job not found")
		return
	}

	results, err := h.scheduler.GetResults(jobID)
	if err != nil {
		writeCrawlError(w, http.StatusNotFound, err.Error())
		return
	}
	writeCrawlJSON(w, http.StatusOK, results)
}

func (h *RESTHandler) cancelJob(w http.ResponseWriter, r *http.Request, jobID string) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		writeCrawlError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// Verify ownership
	wallet := r.Header.Get("X-Moltbunker-Verified-Wallet")
	if wallet == "" {
		writeCrawlError(w, http.StatusForbidden, "no verified identity")
		return
	}
	job, ok := h.scheduler.GetJob(jobID)
	if !ok || job.Owner != wallet {
		writeCrawlError(w, http.StatusNotFound, "job not found")
		return
	}

	if err := h.scheduler.CancelJob(jobID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeCrawlError(w, http.StatusNotFound, err.Error())
			return
		}
		writeCrawlError(w, http.StatusConflict, err.Error())
		return
	}
	writeCrawlJSON(w, http.StatusOK, map[string]string{"status": "cancelled"})
}

func (h *RESTHandler) handleCrawlPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		writeCrawlError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req crawlPageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeCrawlError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.URL == "" {
		writeCrawlError(w, http.StatusBadRequest, "url is required")
		return
	}

	owner := r.Header.Get("X-Moltbunker-Verified-Wallet")
	if owner == "" {
		owner = "anonymous"
	}

	cfg := CrawlConfig{
		URLs:       []string{req.URL},
		MaxDepth:   0,
		MaxPages:   1,
		Selectors:  req.Selectors,
		Screenshot: req.Screenshot,
		JavaScript: req.JavaScript,
		UserAgent:  req.UserAgent,
		Headers:    req.Headers,
		TimeoutSec: req.TimeoutSec,
	}

	job, err := h.scheduler.CreateJob(r.Context(), owner, cfg)
	if err != nil {
		writeCrawlError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeCrawlJSON(w, http.StatusAccepted, job)
}

func (h *RESTHandler) handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		writeCrawlError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	stats := h.scheduler.Stats()
	writeCrawlJSON(w, http.StatusOK, stats)
}

func writeCrawlJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		logging.Warn("failed to encode crawl JSON response",
			"err", err.Error(),
			logging.Component("crawl"))
	}
}

func writeCrawlError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(map[string]string{"error": message}); err != nil {
		logging.Warn("failed to encode crawl error response",
			"err", err.Error(),
			logging.Component("crawl"))
	}
}
