package jsruntime

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/moltbunker/moltbunker/internal/logging"
	"github.com/moltbunker/moltbunker/internal/molt"
)

// DenoPool manages a pool of warm Deno worker processes for JS/TS Molt execution.
type DenoPool struct {
	cfg          DenoConfig
	bindingsPath string
	services     *molt.HostServices

	mu        sync.Mutex
	workers   []*DenoWorker
	available chan *DenoWorker
	nextID    int32

	closed   bool
	closedMu sync.Mutex
}

// NewDenoPool creates a pool of warm Deno workers.
// bindingsPath is the filesystem path to the embedded bindings.ts runtime script.
func NewDenoPool(cfg DenoConfig, bindingsPath string, services *molt.HostServices) (*DenoPool, error) {
	if cfg.PoolSize <= 0 {
		cfg.PoolSize = 10
	}
	if cfg.TimeoutMs <= 0 {
		cfg.TimeoutMs = 30000
	}
	if cfg.MaxMemoryMB <= 0 {
		cfg.MaxMemoryMB = 128
	}

	pool := &DenoPool{
		cfg:          cfg,
		bindingsPath: bindingsPath,
		services:     services,
		workers:      make([]*DenoWorker, 0, cfg.PoolSize),
		available:    make(chan *DenoWorker, cfg.PoolSize),
	}

	// Spawn initial workers
	for i := 0; i < cfg.PoolSize; i++ {
		w, err := pool.spawnWorker()
		if err != nil {
			pool.Close()
			return nil, fmt.Errorf("spawn initial worker %d: %w", i, err)
		}
		pool.workers = append(pool.workers, w)
		pool.available <- w
	}

	logging.Info("deno pool started", "pool_size", cfg.PoolSize, "timeout_ms", cfg.TimeoutMs, "max_memory_mb", cfg.MaxMemoryMB)
	return pool, nil
}

// Execute runs a JS/TS function in an available Deno worker.
func (p *DenoPool) Execute(ctx context.Context, invocation JSInvocation) (*JSResult, error) {
	if p.isClosed() {
		return nil, fmt.Errorf("deno pool is closed")
	}

	// Apply timeout
	timeout := time.Duration(p.cfg.TimeoutMs) * time.Millisecond
	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Acquire a worker
	var worker *DenoWorker
	select {
	case worker = <-p.available:
	case <-execCtx.Done():
		return &JSResult{
			StatusCode: 503,
			Duration:   0,
			Error:      "no available workers",
		}, nil
	}

	// Ensure worker is alive, restart if dead
	if !worker.Alive() {
		logging.Warn("deno worker dead, restarting", "worker_id", worker.id)
		worker.Close()
		var err error
		worker, err = p.spawnWorker()
		if err != nil {
			return nil, fmt.Errorf("restart worker: %w", err)
		}
	}

	// Execute
	result, err := worker.Invoke(execCtx, invocation, p.services)

	// Release worker back to pool (or restart if dead)
	if worker.Alive() {
		p.available <- worker
	} else {
		logging.Warn("deno worker died during invocation, spawning replacement", "worker_id", worker.id)
		worker.Close()
		newWorker, spawnErr := p.spawnWorker()
		if spawnErr != nil {
			logging.Error("failed to spawn replacement worker", "err", spawnErr)
		} else {
			p.available <- newWorker
		}
	}

	return result, err
}

// SetHostServices updates the host services used for Deno host_call dispatch.
func (p *DenoPool) SetHostServices(svc *molt.HostServices) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.services = svc
}

// Close shuts down all Deno workers in the pool.
func (p *DenoPool) Close() error {
	p.closedMu.Lock()
	defer p.closedMu.Unlock()

	if p.closed {
		return nil
	}
	p.closed = true

	// Drain available channel
	close(p.available)
	for w := range p.available {
		w.Close()
	}

	// Close any remaining workers
	p.mu.Lock()
	for _, w := range p.workers {
		w.Close()
	}
	p.workers = nil
	p.mu.Unlock()

	logging.Info("deno pool closed")
	return nil
}

func (p *DenoPool) spawnWorker() (*DenoWorker, error) {
	id := int(atomic.AddInt32(&p.nextID, 1))
	return NewDenoWorker(id, p.cfg, p.bindingsPath)
}

func (p *DenoPool) isClosed() bool {
	p.closedMu.Lock()
	defer p.closedMu.Unlock()
	return p.closed
}
