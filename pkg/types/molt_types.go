package types

import "time"

// RuntimeType distinguishes container workloads from Molt (WASM) workloads.
type RuntimeType string

const (
	RuntimeTypeContainer RuntimeType = "container"
	RuntimeTypeMolt      RuntimeType = "molt"
)

// MoltSpec describes a WASM module to deploy as a Molt serverless function.
// When set on a DeploymentRequest, the deployment is treated as a Molt
// instead of a container.
type MoltSpec struct {
	// ModuleCID is the IPFS CID of the compiled .wasm binary.
	ModuleCID string `json:"module_cid" yaml:"module_cid"`

	// EntryFunction is the WASM function to call on invocation.
	// Defaults to "_start" (WASI convention).
	EntryFunction string `json:"entry_function,omitempty" yaml:"entry_function,omitempty"`

	// MemoryLimitMB is the max WASM linear memory per instance.
	// Default: 256MB.
	MemoryLimitMB uint32 `json:"memory_limit_mb,omitempty" yaml:"memory_limit_mb,omitempty"`

	// TimeoutMs is the max execution time per invocation in milliseconds.
	// Default: 30000 (30s).
	TimeoutMs int `json:"timeout_ms,omitempty" yaml:"timeout_ms,omitempty"`

	// MaxInstances is the max concurrent instances on a single provider.
	// Default: 100.
	MaxInstances int `json:"max_instances,omitempty" yaml:"max_instances,omitempty"`

	// Environment variables passed to the WASM module.
	Environment map[string]string `json:"environment,omitempty" yaml:"environment,omitempty"`
}

// MoltDeploymentStatus represents the lifecycle state of a Molt deployment.
type MoltDeploymentStatus string

const (
	MoltStatusCompiled  MoltDeploymentStatus = "compiled"  // Module compiled, not yet serving
	MoltStatusRunning   MoltDeploymentStatus = "running"   // Serving invocations
	MoltStatusSuspended MoltDeploymentStatus = "suspended" // Suspended (e.g., credit depletion)
	MoltStatusStopped   MoltDeploymentStatus = "stopped"   // Stopped by requester
)

// MoltDeploymentMetrics holds aggregated metrics for a Molt deployment.
type MoltDeploymentMetrics struct {
	TotalInvocations   uint64        `json:"total_invocations"`
	SuccessInvocations uint64        `json:"success_invocations"`
	ErrorInvocations   uint64        `json:"error_invocations"`
	TimeoutInvocations uint64        `json:"timeout_invocations"`
	AvgLatency         time.Duration `json:"avg_latency"`
	CreditsRemaining   uint64        `json:"credits_remaining"` // BUNKER tokens remaining
	LastInvocation     time.Time     `json:"last_invocation,omitempty"`
}
