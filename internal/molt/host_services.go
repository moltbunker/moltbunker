package molt

import (
	"context"
	"sync"

	"github.com/moltbunker/moltbunker/internal/crawl"
	"github.com/moltbunker/moltbunker/internal/proxy"
	"github.com/moltbunker/moltbunker/internal/storage"
)

// contextKey is an unexported type for context keys in this package.
type contextKey struct{}

// hostServicesKey is the context key for HostServices.
var hostServicesKey = contextKey{}

// HostServices provides access to Moltbunker platform services from WASM host functions.
// Each invocation gets its own HostServices instance with a scoped ResultStore.
// Nil fields mean the service is unavailable — host functions return error handles.
type HostServices struct {
	Storage *storage.StorageEngine
	Proxy   proxy.Dialer
	Crawl   *crawl.Scheduler
	Owner   string           // Wallet address — for storage ACL and crawl ownership
	Config  HostCapabilities // Feature flags and restrictions

	results *ResultStore // Per-invocation result handle store
}

// NewHostServices creates a HostServices with a fresh ResultStore.
func NewHostServices(cfg HostCapabilities) *HostServices {
	return &HostServices{
		Config:  cfg,
		results: NewResultStore(),
	}
}

// Results returns the per-invocation result store.
func (h *HostServices) Results() *ResultStore {
	return h.results
}

// Close frees all remaining result handles.
func (h *HostServices) Close() {
	h.results.Close()
}

// HostCapabilities controls which platform services are available to a Molt invocation.
type HostCapabilities struct {
	HTTPEnabled      bool     `yaml:"http_enabled" json:"http_enabled"`
	StorageEnabled   bool     `yaml:"storage_enabled" json:"storage_enabled"`
	CrawlEnabled     bool     `yaml:"crawl_enabled" json:"crawl_enabled"`
	HTTPAllowedHosts []string `yaml:"http_allowed_hosts,omitempty" json:"http_allowed_hosts,omitempty"`
	HTTPBlockedHosts []string `yaml:"http_blocked_hosts,omitempty" json:"http_blocked_hosts,omitempty"`
	StorageBucket    string   `yaml:"storage_bucket,omitempty" json:"storage_bucket,omitempty"` // Scoped bucket for this deployment
}

// withHostServices returns a context with HostServices attached.
func withHostServices(ctx context.Context, svc *HostServices) context.Context {
	return context.WithValue(ctx, hostServicesKey, svc)
}

// servicesFromContext retrieves HostServices from context. Returns nil if absent.
func servicesFromContext(ctx context.Context) *HostServices {
	svc, _ := ctx.Value(hostServicesKey).(*HostServices)
	return svc
}

// ResultStore holds opaque byte buffers keyed by int32 handles.
// Positive handles are data results; negative handles are error messages.
// Thread-safe via mutex. Handles are one-shot: freed on Read/ErrorMessage.
type ResultStore struct {
	mu      sync.Mutex
	data    map[int32][]byte // positive handles → data, negative handles → error strings
	counter int32            // monotonic, starts at 1
}

// NewResultStore creates an empty ResultStore.
func NewResultStore() *ResultStore {
	return &ResultStore{
		data: make(map[int32][]byte),
	}
}

// Store saves data and returns a positive handle.
func (r *ResultStore) Store(data []byte) int32 {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.counter++
	handle := r.counter
	r.data[handle] = data
	return handle
}

// StoreError saves an error message and returns a negative handle.
func (r *ResultStore) StoreError(msg string) int32 {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.counter++
	handle := -r.counter
	r.data[handle] = []byte(msg)
	return handle
}

// Size returns the byte length of a handle's data, or (0, false) if invalid.
func (r *ResultStore) Size(handle int32) (int, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	d, ok := r.data[handle]
	if !ok {
		return 0, false
	}
	return len(d), true
}

// Read returns the data for a positive handle and frees it (one-shot).
func (r *ResultStore) Read(handle int32) ([]byte, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	d, ok := r.data[handle]
	if !ok {
		return nil, false
	}
	delete(r.data, handle)
	return d, true
}

// ErrorMessage returns the error string for a negative handle and frees it.
func (r *ResultStore) ErrorMessage(handle int32) (string, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	d, ok := r.data[handle]
	if !ok {
		return "", false
	}
	delete(r.data, handle)
	return string(d), true
}

// Close frees all remaining handles.
func (r *ResultStore) Close() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data = make(map[int32][]byte)
	r.counter = 0
}

// Len returns the number of active handles (for testing/monitoring).
func (r *ResultStore) Len() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.data)
}
