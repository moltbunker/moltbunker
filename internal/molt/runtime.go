package molt

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"

	"github.com/moltbunker/moltbunker/internal/logging"
)

// MoltRuntime is the core WASM execution engine for Molt serverless functions.
// It manages compilation, caching, concurrency limiting, and invocation via
// stdin/stdout JSON.
type MoltRuntime struct {
	cfg     MoltConfig
	rt      wazero.Runtime
	cache   *ModuleCache
	metrics *MoltMetrics

	// Host services injected into WASM invocations (optional).
	// When set, host functions can access HTTP, storage, and crawl services.
	services *HostServices

	// Semaphore for concurrency limiting (buffered channel)
	sem chan struct{}

	// Monotonic counter for unique module instance names
	moduleCounter int64

	// Shutdown
	closed   bool
	closedMu sync.Mutex
}

// NewMoltRuntime creates a new WASM runtime with the given configuration.
// Zero-value fields in cfg are replaced with defaults.
func NewMoltRuntime(ctx context.Context, cfg MoltConfig) (*MoltRuntime, error) {
	defaults := DefaultMoltConfig()
	if cfg.MemoryLimitMB == 0 {
		cfg.MemoryLimitMB = defaults.MemoryLimitMB
	}
	if cfg.TimeoutMs == 0 {
		cfg.TimeoutMs = defaults.TimeoutMs
	}
	if cfg.MaxInstances == 0 {
		cfg.MaxInstances = defaults.MaxInstances
	}
	if cfg.MaxCacheEntries == 0 {
		cfg.MaxCacheEntries = defaults.MaxCacheEntries
	}
	if cfg.CacheDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("resolving home dir for molt cache: %w", err)
		}
		cfg.CacheDir = filepath.Join(home, ".moltbunker", "molt-cache")
	}

	// Create cache directory
	if err := os.MkdirAll(cfg.CacheDir, 0o700); err != nil {
		return nil, fmt.Errorf("creating molt cache dir %s: %w", cfg.CacheDir, err)
	}

	// Disk-backed compilation cache (survives restarts)
	compilationCache, err := wazero.NewCompilationCacheWithDir(cfg.CacheDir)
	if err != nil {
		return nil, fmt.Errorf("creating wazero compilation cache: %w", err)
	}

	// 1 WASM page = 64KB, so MB * 16 = pages
	memPages := cfg.MemoryLimitMB * 16

	rCfg := wazero.NewRuntimeConfig().
		WithCompilationCache(compilationCache).
		WithMemoryLimitPages(memPages).
		WithCloseOnContextDone(true)

	rt := wazero.NewRuntimeWithConfig(ctx, rCfg)

	// Initialize WASI (fd_read, fd_write, proc_exit, etc.)
	wasi_snapshot_preview1.MustInstantiate(ctx, rt)

	// Register host functions (host.log, stubs)
	if err := registerHostFunctions(ctx, rt); err != nil {
		_ = rt.Close(ctx)
		return nil, fmt.Errorf("registering molt host functions: %w", err)
	}

	logging.Info("molt runtime initialized",
		"memory_limit_mb", cfg.MemoryLimitMB,
		"timeout_ms", cfg.TimeoutMs,
		"max_instances", cfg.MaxInstances,
		"cache_dir", cfg.CacheDir,
	)

	return &MoltRuntime{
		cfg:     cfg,
		rt:      rt,
		cache:   NewModuleCache(cfg.MaxCacheEntries),
		metrics: NewMoltMetrics(),
		sem:     make(chan struct{}, cfg.MaxInstances),
	}, nil
}

// SetHostServices configures platform services available to WASM host functions.
// Must be called before Invoke. Safe to call with nil to disable host services.
func (m *MoltRuntime) SetHostServices(svc *HostServices) {
	m.services = svc
}

// Compile compiles WASM bytes into a reusable CompiledMolt.
// Results are cached by CID — repeated calls with the same CID return the cached module.
func (m *MoltRuntime) Compile(ctx context.Context, wasmBytes []byte, cid string) (*CompiledMolt, error) {
	if m.isClosed() {
		return nil, errors.New("molt runtime is closed")
	}

	// Check cache first
	if cached := m.cache.Get(cid); cached != nil {
		logging.Debug("molt cache hit", "cid", cid)
		return cached, nil
	}

	compiled, err := m.rt.CompileModule(ctx, wasmBytes)
	if err != nil {
		return nil, fmt.Errorf("compiling wasm module (cid=%s): %w", cid, err)
	}

	molt := &CompiledMolt{
		CID:        cid,
		Module:     compiled,
		CompiledAt: time.Now(),
		SizeBytes:  int64(len(wasmBytes)),
	}

	m.cache.Put(cid, molt)
	logging.Info("molt compiled and cached", "cid", cid, "size_bytes", molt.SizeBytes)

	return molt, nil
}

// Invoke runs a compiled Molt with the given invocation parameters.
// The WASM module reads a MoltHTTPRequest from stdin and writes a MoltHTTPResponse to stdout.
func (m *MoltRuntime) Invoke(ctx context.Context, compiled *CompiledMolt, invocation MoltInvocation) (*MoltResult, error) {
	if m.isClosed() {
		return nil, errors.New("molt runtime is closed")
	}

	start := time.Now()
	m.metrics.IncrementActive()
	defer m.metrics.DecrementActive()

	// Acquire semaphore (blocks if at concurrency limit)
	select {
	case m.sem <- struct{}{}:
		defer func() { <-m.sem }()
	case <-ctx.Done():
		return nil, fmt.Errorf("acquiring molt semaphore: %w", ctx.Err())
	}

	// Apply timeout
	timeout := time.Duration(m.cfg.TimeoutMs) * time.Millisecond
	invokeCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Inject host services into context (if configured)
	if m.services != nil {
		invokeCtx = withHostServices(invokeCtx, m.services)
	}

	// Serialize request to stdin JSON
	httpReq := MoltHTTPRequest{
		Method:  invocation.Method,
		Path:    invocation.Path,
		Headers: invocation.Headers,
		Body:    base64.StdEncoding.EncodeToString(invocation.Body),
	}
	stdinBuf, err := json.Marshal(httpReq)
	if err != nil {
		return nil, fmt.Errorf("marshaling molt request: %w", err)
	}

	var stdout, stderr bytes.Buffer
	instanceID := atomic.AddInt64(&m.moduleCounter, 1)
	moduleName := fmt.Sprintf("molt-%d", instanceID)

	modCfg := wazero.NewModuleConfig().
		WithName(moduleName).
		WithStdin(bytes.NewReader(stdinBuf)).
		WithStdout(&stdout).
		WithStderr(&stderr).
		WithSysNanotime().
		WithSysWalltime()

	// Instantiate + run _start (this IS the invocation)
	mod, err := m.rt.InstantiateModule(invokeCtx, compiled.Module, modCfg)
	duration := time.Since(start)

	// Determine outcome
	isTimeout := invokeCtx.Err() == context.DeadlineExceeded

	if err != nil {
		m.metrics.RecordInvocation(invocation.DeploymentID, duration, false, isTimeout)

		if isTimeout {
			return &MoltResult{
				StatusCode: 504,
				Duration:   duration,
				Error:      "invocation timed out",
			}, nil
		}
		return &MoltResult{
			StatusCode: 500,
			Duration:   duration,
			Error:      err.Error(),
		}, nil
	}
	defer mod.Close(invokeCtx)

	// Read memory usage (guarded: modules without memory return a nil-wrapped interface)
	var memUsed uint32
	func() {
		defer func() { _ = recover() }()
		if mem := mod.Memory(); mem != nil {
			memUsed = mem.Size()
		}
	}()

	// Parse stdout JSON response
	result := &MoltResult{
		Duration:        duration,
		MemoryUsedBytes: memUsed,
	}

	outBytes := stdout.Bytes()
	if len(outBytes) == 0 {
		// Module produced no output — treat as 200 with empty body
		result.StatusCode = 200
		m.metrics.RecordInvocation(invocation.DeploymentID, duration, true, false)
		return result, nil
	}

	var httpResp MoltHTTPResponse
	if err := json.Unmarshal(outBytes, &httpResp); err != nil {
		// Module wrote non-JSON to stdout — return raw output as body
		result.StatusCode = 200
		result.Body = outBytes
		m.metrics.RecordInvocation(invocation.DeploymentID, duration, true, false)
		return result, nil
	}

	result.StatusCode = httpResp.StatusCode
	if result.StatusCode == 0 {
		result.StatusCode = 200
	}
	result.Headers = httpResp.Headers

	if httpResp.Body != "" {
		decoded, err := base64.StdEncoding.DecodeString(httpResp.Body)
		if err != nil {
			// Treat as plain text if base64 decode fails
			result.Body = []byte(httpResp.Body)
		} else {
			result.Body = decoded
		}
	}

	m.metrics.RecordInvocation(invocation.DeploymentID, duration, true, false)
	return result, nil
}

// Metrics returns the runtime's metrics collector.
func (m *MoltRuntime) Metrics() *MoltMetrics {
	return m.metrics
}

// Cache returns the runtime's module cache.
func (m *MoltRuntime) Cache() *ModuleCache {
	return m.cache
}

// Close shuts down the runtime, closing all cached modules and the wazero runtime.
// Safe to call multiple times.
func (m *MoltRuntime) Close(ctx context.Context) error {
	m.closedMu.Lock()
	defer m.closedMu.Unlock()

	if m.closed {
		return nil
	}
	m.closed = true

	m.cache.Close()

	if err := m.rt.Close(ctx); err != nil {
		return fmt.Errorf("closing wazero runtime: %w", err)
	}

	logging.Info("molt runtime closed")
	return nil
}

func (m *MoltRuntime) isClosed() bool {
	m.closedMu.Lock()
	defer m.closedMu.Unlock()
	return m.closed
}
