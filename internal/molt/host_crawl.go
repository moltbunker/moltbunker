package molt

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/tetratelabs/wazero/api"

	"github.com/moltbunker/moltbunker/internal/crawl"
	"github.com/moltbunker/moltbunker/internal/logging"
)

// CrawlRequest is the JSON envelope for crawl operations from WASM and Deno.
type CrawlRequest struct {
	URL        string   `json:"url"`
	Selectors  []string `json:"selectors,omitempty"`
	Screenshot bool     `json:"screenshot,omitempty"`
	JavaScript bool     `json:"js,omitempty"`
}

// hostCrawlPage crawls a single web page via the crawl scheduler.
// Params: [req_ptr i32, req_len i32] → [handle i32]
// Returns a handle with CrawlJob JSON (async — the job ID is returned immediately).
func hostCrawlPage(ctx context.Context, mod api.Module, stack []uint64) {
	reqPtr := api.DecodeU32(stack[0])
	reqLen := api.DecodeU32(stack[1])

	svc := servicesFromContext(ctx)
	if svc == nil {
		stack[0] = api.EncodeI32(-1)
		return
	}

	if !svc.Config.CrawlEnabled || svc.Crawl == nil {
		stack[0] = api.EncodeI32(svc.results.StoreError("crawl: service disabled"))
		return
	}

	mem := mod.Memory()
	if mem == nil {
		stack[0] = api.EncodeI32(svc.results.StoreError("crawl: no memory"))
		return
	}

	reqBytes, ok := mem.Read(reqPtr, reqLen)
	if !ok {
		stack[0] = api.EncodeI32(svc.results.StoreError("crawl: invalid memory read"))
		return
	}

	var req CrawlRequest
	if err := json.Unmarshal(reqBytes, &req); err != nil {
		stack[0] = api.EncodeI32(svc.results.StoreError(fmt.Sprintf("crawl: invalid JSON: %v", err)))
		return
	}

	if req.URL == "" {
		stack[0] = api.EncodeI32(svc.results.StoreError("crawl: url is required"))
		return
	}

	// Apply the same SSRF guard as HTTP requests
	if err := validateHost(hostFromURL(req.URL)); err != nil {
		stack[0] = api.EncodeI32(svc.results.StoreError(fmt.Sprintf("crawl: %v", err)))
		return
	}

	job, err := ExecuteCrawl(ctx, svc, &req)
	if err != nil {
		stack[0] = api.EncodeI32(svc.results.StoreError(fmt.Sprintf("crawl: %v", err)))
		return
	}

	jobJSON, err := json.Marshal(job)
	if err != nil {
		stack[0] = api.EncodeI32(svc.results.StoreError(fmt.Sprintf("crawl: marshal job: %v", err)))
		return
	}

	logging.Debug("host.crawl_page completed", "url", req.URL, "job_id", job.ID)
	stack[0] = api.EncodeI32(svc.results.Store(jobJSON))
}

// ExecuteCrawl creates a single-page crawl job. Exported for Deno host_call dispatch.
func ExecuteCrawl(ctx context.Context, svc *HostServices, req *CrawlRequest) (*crawl.CrawlJob, error) {
	cfg := crawl.CrawlConfig{
		URLs:       []string{req.URL},
		MaxDepth:   0, // single page
		MaxPages:   1,
		Selectors:  req.Selectors,
		Screenshot: req.Screenshot,
		JavaScript: req.JavaScript,
	}

	job, err := svc.Crawl.CreateJob(ctx, svc.Owner, cfg)
	if err != nil {
		return nil, fmt.Errorf("creating crawl job: %w", err)
	}

	return job, nil
}

// hostFromURL extracts the hostname from a URL string. Returns empty on parse failure.
func hostFromURL(rawURL string) string {
	// Quick parse — only need the host
	for _, prefix := range []string{"https://", "http://"} {
		if len(rawURL) > len(prefix) && rawURL[:len(prefix)] == prefix {
			rest := rawURL[len(prefix):]
			// Find end of host (: or / or end of string)
			for i, c := range rest {
				if c == '/' || c == ':' || c == '?' {
					return rest[:i]
				}
			}
			return rest
		}
	}
	return ""
}
