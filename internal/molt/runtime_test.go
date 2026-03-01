package molt

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// loadTestWASM reads a .wasm file from testdata/
func loadTestWASM(t *testing.T, name string) []byte {
	t.Helper()
	path := filepath.Join("testdata", name)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to load test wasm %s: %v", name, err)
	}
	return data
}

// newTestRuntime creates a MoltRuntime with a temp cache dir for testing.
func newTestRuntime(t *testing.T) *MoltRuntime {
	t.Helper()
	ctx := context.Background()
	cfg := MoltConfig{
		MemoryLimitMB:   64,
		TimeoutMs:       5000,
		MaxInstances:    10,
		CacheDir:        t.TempDir(),
		MaxCacheEntries: 16,
	}
	rt, err := NewMoltRuntime(ctx, cfg)
	if err != nil {
		t.Fatalf("NewMoltRuntime: %v", err)
	}
	t.Cleanup(func() { rt.Close(ctx) })
	return rt
}

// --- MoltConfig Tests ---

func TestDefaultMoltConfig(t *testing.T) {
	cfg := DefaultMoltConfig()
	if cfg.MemoryLimitMB != 256 {
		t.Errorf("MemoryLimitMB = %d, want 256", cfg.MemoryLimitMB)
	}
	if cfg.TimeoutMs != 30000 {
		t.Errorf("TimeoutMs = %d, want 30000", cfg.TimeoutMs)
	}
	if cfg.MaxInstances != 100 {
		t.Errorf("MaxInstances = %d, want 100", cfg.MaxInstances)
	}
	if cfg.MaxCacheEntries != 256 {
		t.Errorf("MaxCacheEntries = %d, want 256", cfg.MaxCacheEntries)
	}
}

// --- ModuleCache Tests ---

func TestModuleCache_GetPut(t *testing.T) {
	cache := NewModuleCache(4)

	// Get from empty cache
	if got := cache.Get("nonexistent"); got != nil {
		t.Fatal("expected nil from empty cache")
	}

	molt := &CompiledMolt{CID: "cid1", CompiledAt: time.Now()}
	cache.Put("cid1", molt)

	if got := cache.Get("cid1"); got != molt {
		t.Fatal("expected to retrieve cached molt")
	}
	if cache.Size() != 1 {
		t.Fatalf("Size() = %d, want 1", cache.Size())
	}
}

func TestModuleCache_Evict(t *testing.T) {
	cache := NewModuleCache(4)
	molt := &CompiledMolt{CID: "cid1"}
	cache.Put("cid1", molt)

	cache.Evict("cid1")
	if got := cache.Get("cid1"); got != nil {
		t.Fatal("expected nil after eviction")
	}
	if cache.Size() != 0 {
		t.Fatalf("Size() = %d, want 0", cache.Size())
	}
}

func TestModuleCache_EvictsOldest(t *testing.T) {
	cache := NewModuleCache(2)

	cache.Put("old", &CompiledMolt{CID: "old"})
	time.Sleep(time.Millisecond) // ensure distinct timestamps
	cache.Put("new", &CompiledMolt{CID: "new"})

	// Third entry should evict "old"
	cache.Put("newest", &CompiledMolt{CID: "newest"})

	if cache.Size() != 2 {
		t.Fatalf("Size() = %d, want 2", cache.Size())
	}
	if got := cache.Get("old"); got != nil {
		t.Fatal("expected 'old' to be evicted")
	}
	if got := cache.Get("new"); got == nil {
		t.Fatal("expected 'new' to still be cached")
	}
	if got := cache.Get("newest"); got == nil {
		t.Fatal("expected 'newest' to still be cached")
	}
}

func TestModuleCache_PutUpdatesExisting(t *testing.T) {
	cache := NewModuleCache(2)

	m1 := &CompiledMolt{CID: "cid1", SizeBytes: 100}
	m2 := &CompiledMolt{CID: "cid1", SizeBytes: 200}

	cache.Put("cid1", m1)
	cache.Put("cid1", m2) // update, not a new entry

	if cache.Size() != 1 {
		t.Fatalf("Size() = %d, want 1 (update should not add)", cache.Size())
	}
	if got := cache.Get("cid1"); got.SizeBytes != 200 {
		t.Fatalf("SizeBytes = %d, want 200", got.SizeBytes)
	}
}

func TestModuleCache_Close(t *testing.T) {
	cache := NewModuleCache(4)
	cache.Put("a", &CompiledMolt{CID: "a"})
	cache.Put("b", &CompiledMolt{CID: "b"})
	cache.Close()

	if cache.Size() != 0 {
		t.Fatalf("Size() = %d, want 0 after Close", cache.Size())
	}
}

// --- Metrics Tests ---

func TestMoltMetrics_RecordInvocation(t *testing.T) {
	m := NewMoltMetrics()

	m.RecordInvocation("dep1", 10*time.Millisecond, true, false)
	m.RecordInvocation("dep1", 20*time.Millisecond, false, false)
	m.RecordInvocation("dep1", 30*time.Millisecond, false, true)
	m.RecordInvocation("dep2", 5*time.Millisecond, true, false)

	global := m.GetGlobalStats()
	if global.TotalInvocations != 4 {
		t.Fatalf("total = %d, want 4", global.TotalInvocations)
	}
	if global.SuccessInvocations != 2 {
		t.Fatalf("success = %d, want 2", global.SuccessInvocations)
	}
	if global.ErrorInvocations != 1 {
		t.Fatalf("errors = %d, want 1", global.ErrorInvocations)
	}
	if global.TimeoutInvocations != 1 {
		t.Fatalf("timeouts = %d, want 1", global.TimeoutInvocations)
	}

	dep1 := m.GetStats("dep1")
	if dep1.TotalInvocations != 3 {
		t.Fatalf("dep1 total = %d, want 3", dep1.TotalInvocations)
	}

	dep2 := m.GetStats("dep2")
	if dep2.TotalInvocations != 1 {
		t.Fatalf("dep2 total = %d, want 1", dep2.TotalInvocations)
	}

	// Non-existent deployment
	dep3 := m.GetStats("dep3")
	if dep3.TotalInvocations != 0 {
		t.Fatalf("dep3 total = %d, want 0", dep3.TotalInvocations)
	}
}

func TestMoltMetrics_ActiveInvocations(t *testing.T) {
	m := NewMoltMetrics()

	m.IncrementActive()
	m.IncrementActive()
	m.IncrementActive()
	m.DecrementActive()

	global := m.GetGlobalStats()
	if global.ActiveInvocations != 2 {
		t.Fatalf("active = %d, want 2", global.ActiveInvocations)
	}
}

// --- Runtime Tests ---

func TestNewMoltRuntime(t *testing.T) {
	rt := newTestRuntime(t)
	if rt == nil {
		t.Fatal("runtime should not be nil")
	}
}

func TestNewMoltRuntime_DefaultsApplied(t *testing.T) {
	ctx := context.Background()
	cfg := MoltConfig{
		CacheDir: t.TempDir(),
	}
	rt, err := NewMoltRuntime(ctx, cfg)
	if err != nil {
		t.Fatalf("NewMoltRuntime: %v", err)
	}
	defer rt.Close(ctx)

	if rt.cfg.MemoryLimitMB != 256 {
		t.Errorf("MemoryLimitMB = %d, want 256", rt.cfg.MemoryLimitMB)
	}
	if rt.cfg.TimeoutMs != 30000 {
		t.Errorf("TimeoutMs = %d, want 30000", rt.cfg.TimeoutMs)
	}
	if rt.cfg.MaxInstances != 100 {
		t.Errorf("MaxInstances = %d, want 100", rt.cfg.MaxInstances)
	}
}

func TestCompile(t *testing.T) {
	rt := newTestRuntime(t)
	wasm := loadTestWASM(t, "noop.wasm")

	compiled, err := rt.Compile(context.Background(), wasm, "test-cid-noop")
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	if compiled.CID != "test-cid-noop" {
		t.Fatalf("CID = %q, want %q", compiled.CID, "test-cid-noop")
	}
	if compiled.Module == nil {
		t.Fatal("compiled.Module should not be nil")
	}
	if compiled.SizeBytes != int64(len(wasm)) {
		t.Fatalf("SizeBytes = %d, want %d", compiled.SizeBytes, len(wasm))
	}
}

func TestCompile_CachesResult(t *testing.T) {
	rt := newTestRuntime(t)
	wasm := loadTestWASM(t, "noop.wasm")

	c1, err := rt.Compile(context.Background(), wasm, "cached-cid")
	if err != nil {
		t.Fatalf("first Compile: %v", err)
	}

	c2, err := rt.Compile(context.Background(), wasm, "cached-cid")
	if err != nil {
		t.Fatalf("second Compile: %v", err)
	}

	if c1 != c2 {
		t.Fatal("expected same pointer from cache")
	}
	if rt.Cache().Size() != 1 {
		t.Fatalf("cache size = %d, want 1", rt.Cache().Size())
	}
}

func TestCompile_InvalidWASM(t *testing.T) {
	rt := newTestRuntime(t)

	_, err := rt.Compile(context.Background(), []byte("not wasm"), "bad-cid")
	if err == nil {
		t.Fatal("expected error for invalid WASM")
	}
}

func TestInvoke_Noop(t *testing.T) {
	rt := newTestRuntime(t)
	wasm := loadTestWASM(t, "noop.wasm")

	compiled, err := rt.Compile(context.Background(), wasm, "noop")
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}

	result, err := rt.Invoke(context.Background(), compiled, MoltInvocation{
		DeploymentID: "test-deploy",
		Method:       "GET",
		Path:         "/health",
	})
	if err != nil {
		t.Fatalf("Invoke: %v", err)
	}

	// Noop produces no stdout → 200 with empty body
	if result.StatusCode != 200 {
		t.Fatalf("StatusCode = %d, want 200", result.StatusCode)
	}
	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	if result.Duration <= 0 {
		t.Fatal("duration should be positive")
	}
}

func TestInvoke_Echo(t *testing.T) {
	rt := newTestRuntime(t)
	wasm := loadTestWASM(t, "echo.wasm")

	compiled, err := rt.Compile(context.Background(), wasm, "echo")
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}

	// Echo pipes stdin→stdout unchanged. Invoke wraps the body into MoltHTTPRequest JSON
	// and then parses stdout as MoltHTTPResponse. Since both share the "body" JSON field,
	// the base64-encoded body roundtrips correctly. StatusCode defaults to 200 (no "status_code"
	// in MoltHTTPRequest, parsed as 0, defaulted to 200).
	result, err := rt.Invoke(context.Background(), compiled, MoltInvocation{
		DeploymentID: "echo-deploy",
		Method:       "POST",
		Path:         "/echo",
		Body:         []byte("hello molt"),
	})
	if err != nil {
		t.Fatalf("Invoke: %v", err)
	}

	if result.StatusCode != 200 {
		t.Fatalf("StatusCode = %d, want 200", result.StatusCode)
	}
	if string(result.Body) != "hello molt" {
		t.Fatalf("Body = %q, want %q", string(result.Body), "hello molt")
	}
	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	if result.MemoryUsedBytes == 0 {
		t.Log("note: echo module memory size reported as 0 (expected for minimal module)")
	}
}

func TestInvoke_Timeout(t *testing.T) {
	ctx := context.Background()
	cfg := MoltConfig{
		MemoryLimitMB:   16,
		TimeoutMs:       200, // 200ms timeout
		MaxInstances:    5,
		CacheDir:        t.TempDir(),
		MaxCacheEntries: 4,
	}
	rt, err := NewMoltRuntime(ctx, cfg)
	if err != nil {
		t.Fatalf("NewMoltRuntime: %v", err)
	}
	defer rt.Close(ctx)

	wasm := loadTestWASM(t, "spin.wasm")
	compiled, err := rt.Compile(ctx, wasm, "spin")
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}

	result, err := rt.Invoke(ctx, compiled, MoltInvocation{
		DeploymentID: "spin-deploy",
		Method:       "GET",
		Path:         "/spin",
	})
	if err != nil {
		t.Fatalf("Invoke returned error: %v", err)
	}

	if result.StatusCode != 504 {
		t.Fatalf("StatusCode = %d, want 504 (timeout)", result.StatusCode)
	}
	if result.Error == "" {
		t.Fatal("expected error message for timeout")
	}

	// Check metrics recorded the timeout
	stats := rt.Metrics().GetStats("spin-deploy")
	if stats.TimeoutInvocations != 1 {
		t.Fatalf("timeout count = %d, want 1", stats.TimeoutInvocations)
	}
}

func TestInvoke_ConcurrencyLimit(t *testing.T) {
	ctx := context.Background()
	cfg := MoltConfig{
		MemoryLimitMB:   16,
		TimeoutMs:       5000,
		MaxInstances:    2, // only 2 concurrent
		CacheDir:        t.TempDir(),
		MaxCacheEntries: 4,
	}
	rt, err := NewMoltRuntime(ctx, cfg)
	if err != nil {
		t.Fatalf("NewMoltRuntime: %v", err)
	}
	defer rt.Close(ctx)

	wasm := loadTestWASM(t, "noop.wasm")
	compiled, err := rt.Compile(ctx, wasm, "noop")
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}

	// Launch many concurrent invocations
	const n = 20
	var wg sync.WaitGroup
	errors := make([]error, n)

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_, errors[idx] = rt.Invoke(ctx, compiled, MoltInvocation{
				DeploymentID: fmt.Sprintf("conc-%d", idx),
				Method:       "GET",
				Path:         "/test",
			})
		}(i)
	}
	wg.Wait()

	// All should succeed (semaphore queues, doesn't reject)
	for i, err := range errors {
		if err != nil {
			t.Errorf("invocation %d failed: %v", i, err)
		}
	}

	// Verify total invocation count
	global := rt.Metrics().GetGlobalStats()
	if global.TotalInvocations != n {
		t.Fatalf("total invocations = %d, want %d", global.TotalInvocations, n)
	}
}

func TestInvoke_AfterClose(t *testing.T) {
	ctx := context.Background()
	cfg := MoltConfig{
		CacheDir: t.TempDir(),
	}
	rt, err := NewMoltRuntime(ctx, cfg)
	if err != nil {
		t.Fatalf("NewMoltRuntime: %v", err)
	}

	wasm := loadTestWASM(t, "noop.wasm")
	compiled, err := rt.Compile(ctx, wasm, "noop")
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}

	rt.Close(ctx)

	_, err = rt.Invoke(ctx, compiled, MoltInvocation{
		DeploymentID: "closed",
		Method:       "GET",
		Path:         "/",
	})
	if err == nil {
		t.Fatal("expected error after Close")
	}
}

func TestClose_Idempotent(t *testing.T) {
	ctx := context.Background()
	cfg := MoltConfig{
		CacheDir: t.TempDir(),
	}
	rt, err := NewMoltRuntime(ctx, cfg)
	if err != nil {
		t.Fatalf("NewMoltRuntime: %v", err)
	}

	// Close multiple times — should not panic
	if err := rt.Close(ctx); err != nil {
		t.Fatalf("first Close: %v", err)
	}
	if err := rt.Close(ctx); err != nil {
		t.Fatalf("second Close: %v", err)
	}
}

// --- HTTP Handler Tests ---

func TestMoltHTTPHandler_Success(t *testing.T) {
	rt := newTestRuntime(t)
	wasm := loadTestWASM(t, "echo.wasm")

	compiled, err := rt.Compile(context.Background(), wasm, "echo-http")
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}

	// Use noop module — verifies the handler plumbing (compile, invoke, write response)
	noopWasm := loadTestWASM(t, "noop.wasm")
	noopCompiled, err := rt.Compile(context.Background(), noopWasm, "noop-http")
	if err != nil {
		t.Fatalf("Compile noop: %v", err)
	}
	_ = compiled // echo compiled above, used in other tests

	noopHandler := NewMoltHTTPHandler(rt, noopCompiled, "noop-http-deploy")

	req := httptest.NewRequest("GET", "/health", nil)
	rec := httptest.NewRecorder()
	noopHandler.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}

func TestMoltHTTPHandler_LargeBody(t *testing.T) {
	rt := newTestRuntime(t)
	wasm := loadTestWASM(t, "noop.wasm")
	compiled, err := rt.Compile(context.Background(), wasm, "noop-large")
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}

	handler := NewMoltHTTPHandler(rt, compiled, "large-deploy")

	// Body just over 10MB
	bigBody := strings.NewReader(strings.Repeat("x", maxRequestBodySize+1))
	req := httptest.NewRequest("POST", "/upload", bigBody)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want 413", rec.Code)
	}
}

func TestMoltHTTPHandler_POST(t *testing.T) {
	rt := newTestRuntime(t)
	wasm := loadTestWASM(t, "echo.wasm")
	compiled, err := rt.Compile(context.Background(), wasm, "echo-post")
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}

	handler := NewMoltHTTPHandler(rt, compiled, "echo-post-deploy")

	body := `{"key":"value"}`
	req := httptest.NewRequest("POST", "/api/data", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if string(rec.Body.Bytes()) != body {
		t.Fatalf("body = %q, want %q", rec.Body.String(), body)
	}
}

func TestMoltHTTPHandler_ConcurrentRequests(t *testing.T) {
	rt := newTestRuntime(t)
	wasm := loadTestWASM(t, "noop.wasm")
	compiled, err := rt.Compile(context.Background(), wasm, "conc-http")
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}

	handler := NewMoltHTTPHandler(rt, compiled, "conc-http-deploy")

	const n = 10
	var wg sync.WaitGroup
	statuses := make([]int, n)

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			req := httptest.NewRequest("GET", fmt.Sprintf("/test/%d", idx), nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			statuses[idx] = rec.Code
		}(i)
	}
	wg.Wait()

	for i, status := range statuses {
		if status != 200 {
			t.Errorf("request %d: status = %d, want 200", i, status)
		}
	}
}

func TestMoltHTTPHandler_DifferentMethods(t *testing.T) {
	rt := newTestRuntime(t)
	wasm := loadTestWASM(t, "noop.wasm")
	compiled, err := rt.Compile(context.Background(), wasm, "methods")
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}

	handler := NewMoltHTTPHandler(rt, compiled, "methods-deploy")

	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}
	for _, method := range methods {
		req := httptest.NewRequest(method, "/test", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != 200 {
			t.Errorf("%s: status = %d, want 200", method, rec.Code)
		}
	}
}

// --- Helpers ---

func mustJSON(v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return data
}
