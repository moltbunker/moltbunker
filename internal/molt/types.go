package molt

import (
	"time"

	"github.com/tetratelabs/wazero"
)

// RuntimeType identifies the execution engine for a Molt deployment.
type RuntimeType string

const (
	// RuntimeWASM uses wazero (in-process WASM execution).
	RuntimeWASM RuntimeType = "wasm"
	// RuntimeJS uses Deno worker pool (subprocess, stdio JSON-RPC).
	RuntimeJS RuntimeType = "js"
)

// MoltConfig configures the WASM runtime for Molt serverless functions.
type MoltConfig struct {
	// MemoryLimitMB is the max memory each WASM instance can use.
	// Each WASM page is 64KB, so this is converted to pages internally.
	MemoryLimitMB uint32 `yaml:"memory_limit_mb" json:"memory_limit_mb"`

	// TimeoutMs is the max execution time per invocation in milliseconds.
	TimeoutMs int `yaml:"timeout_ms" json:"timeout_ms"`

	// MaxInstances is the max concurrent WASM instances.
	MaxInstances int `yaml:"max_instances" json:"max_instances"`

	// CacheDir is the directory for the wazero compilation cache.
	CacheDir string `yaml:"cache_dir" json:"cache_dir"`

	// MaxCacheEntries is the max number of compiled modules kept in memory.
	MaxCacheEntries int `yaml:"max_cache_entries" json:"max_cache_entries"`

	// Host function capabilities (v2 — controls what services WASM modules can access)
	HTTPEnabled      bool     `yaml:"http_enabled" json:"http_enabled"`
	StorageEnabled   bool     `yaml:"storage_enabled" json:"storage_enabled"`
	CrawlEnabled     bool     `yaml:"crawl_enabled" json:"crawl_enabled"`
	HTTPAllowedHosts []string `yaml:"http_allowed_hosts,omitempty" json:"http_allowed_hosts,omitempty"`
	HTTPBlockedHosts []string `yaml:"http_blocked_hosts,omitempty" json:"http_blocked_hosts,omitempty"`
}

// DefaultMoltConfig returns a MoltConfig with sensible defaults.
func DefaultMoltConfig() MoltConfig {
	return MoltConfig{
		MemoryLimitMB:   256,
		TimeoutMs:       30000,
		MaxInstances:    100,
		CacheDir:        "", // resolved at runtime to ~/.moltbunker/molt-cache/
		MaxCacheEntries: 256,
	}
}

// CompiledMolt holds a compiled WASM module with metadata.
type CompiledMolt struct {
	CID        string
	Module     wazero.CompiledModule
	CompiledAt time.Time
	SizeBytes  int64
}

// MoltInvocation describes a single function invocation request.
type MoltInvocation struct {
	DeploymentID string            `json:"deployment_id"`
	Method       string            `json:"method"`
	Path         string            `json:"path"`
	Headers      map[string]string `json:"headers,omitempty"`
	Body         []byte            `json:"-"` // raw body, encoded to base64 in MoltHTTPRequest
}

// MoltResult is the outcome of a single function invocation.
type MoltResult struct {
	StatusCode      int               `json:"status_code"`
	Headers         map[string]string `json:"headers,omitempty"`
	Body            []byte            `json:"-"`
	Duration        time.Duration     `json:"duration"`
	MemoryUsedBytes uint32            `json:"memory_used_bytes"`
	Error           string            `json:"error,omitempty"`
}

// MoltHTTPRequest is the JSON envelope written to WASM stdin.
// Body is base64-encoded for binary safety.
type MoltHTTPRequest struct {
	Method  string            `json:"method"`
	Path    string            `json:"path"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    string            `json:"body"` // base64-encoded
}

// MoltHTTPResponse is the JSON envelope read from WASM stdout.
// Body is base64-encoded for binary safety.
type MoltHTTPResponse struct {
	StatusCode int               `json:"status_code"`
	Headers    map[string]string `json:"headers,omitempty"`
	Body       string            `json:"body"` // base64-encoded
}
