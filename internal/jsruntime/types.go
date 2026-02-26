package jsruntime

import "time"

// DenoConfig configures the Deno worker pool for JS/TS Molt execution.
type DenoConfig struct {
	// Enabled controls whether the Deno runtime is available.
	Enabled bool `yaml:"enabled" json:"enabled"`

	// DenoPath is the path to the Deno binary. Empty defaults to "deno" (found in PATH).
	DenoPath string `yaml:"deno_path" json:"deno_path"`

	// PoolSize is the number of warm Deno worker processes (default: 10).
	PoolSize int `yaml:"pool_size" json:"pool_size"`

	// TimeoutMs is the max execution time per invocation in milliseconds (default: 30000).
	TimeoutMs int `yaml:"timeout_ms" json:"timeout_ms"`

	// MaxMemoryMB is the V8 heap size limit in MB per worker (default: 128).
	MaxMemoryMB int `yaml:"max_memory_mb" json:"max_memory_mb"`
}

// DefaultDenoConfig returns sensible defaults.
func DefaultDenoConfig() DenoConfig {
	return DenoConfig{
		Enabled:     false,
		DenoPath:    "deno",
		PoolSize:    10,
		TimeoutMs:   30000,
		MaxMemoryMB: 128,
	}
}

// JSInvocation describes a single JS/TS function invocation.
type JSInvocation struct {
	ScriptPath   string            `json:"script_path"`
	DeploymentID string            `json:"deployment_id"`
	Method       string            `json:"method"`
	URL          string            `json:"url"`
	Headers      map[string]string `json:"headers,omitempty"`
	Body         []byte            `json:"body,omitempty"`
}

// JSResult is the outcome of a single JS/TS function invocation.
type JSResult struct {
	StatusCode int               `json:"status_code"`
	Headers    map[string]string `json:"headers,omitempty"`
	Body       []byte            `json:"body,omitempty"`
	Duration   time.Duration     `json:"duration"`
	Error      string            `json:"error,omitempty"`
}
