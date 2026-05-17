package agent

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/moltbunker/moltbunker/internal/logging"
)

// AgentRuntime manages the lifecycle of AI agent deployments.
type AgentRuntime struct {
	mu     sync.RWMutex
	agents map[string]*AgentDeployment // id → deployment
	config RuntimeConfig
}

// RuntimeConfig configures the agent runtime.
type RuntimeConfig struct {
	MaxAgentsPerWallet int           // Max concurrent agents per wallet (default: 10)
	DefaultTimeout     time.Duration // Default agent timeout (default: 1 hour)
	MaxTokenBudget     int64         // Default max token budget (default: 1M)
	MemoryBucketPrefix string        // Prefix for agent memory buckets (default: "agents")
}

// DefaultRuntimeConfig returns sensible defaults.
func DefaultRuntimeConfig() RuntimeConfig {
	return RuntimeConfig{
		MaxAgentsPerWallet: 10,
		DefaultTimeout:     time.Hour,
		MaxTokenBudget:     1_000_000,
		MemoryBucketPrefix: "agents",
	}
}

// NewAgentRuntime creates a new agent runtime.
func NewAgentRuntime(cfg RuntimeConfig) *AgentRuntime {
	return &AgentRuntime{
		agents: make(map[string]*AgentDeployment),
		config: cfg,
	}
}

// Deploy creates a new agent deployment from the given spec.
func (r *AgentRuntime) Deploy(ctx context.Context, spec AgentSpec) (*AgentDeployment, error) {
	if spec.Owner == "" {
		return nil, fmt.Errorf("agent owner is required")
	}
	if spec.Framework == "" {
		return nil, fmt.Errorf("agent framework is required")
	}

	// Validate framework
	switch spec.Framework {
	case FrameworkLangGraph, FrameworkCrewAI, FrameworkAutoGen, FrameworkCustom:
	default:
		return nil, fmt.Errorf("unsupported framework: %q", spec.Framework)
	}

	// Check per-wallet limit
	r.mu.RLock()
	count := 0
	for _, a := range r.agents {
		if a.Spec.Owner == spec.Owner && a.Status != AgentStatusStopped && a.Status != AgentStatusFailed {
			count++
		}
	}
	r.mu.RUnlock()

	if count >= r.config.MaxAgentsPerWallet {
		return nil, fmt.Errorf("agent limit exceeded: max %d agents per wallet", r.config.MaxAgentsPerWallet)
	}

	// Apply framework defaults
	if spec.Image == "" {
		defaultImage, defaultEnv := FrameworkDefaults(spec.Framework)
		if defaultImage == "" && spec.Framework != FrameworkCustom {
			return nil, fmt.Errorf("no default image for framework %q, provide --image", spec.Framework)
		}
		spec.Image = defaultImage
		if spec.Env == nil {
			spec.Env = make(map[string]string)
		}
		for k, v := range defaultEnv {
			if _, exists := spec.Env[k]; !exists {
				spec.Env[k] = v
			}
		}
	}

	// Set defaults
	if spec.MaxTokens <= 0 {
		spec.MaxTokens = r.config.MaxTokenBudget
	}
	if spec.MemoryLimitMB <= 0 {
		spec.MemoryLimitMB = 512
	}
	if spec.CPUCores <= 0 {
		spec.CPUCores = 1
	}

	// Add builtin MCP tools
	spec.MCPTools = append(BuiltinMCPTools(), spec.MCPTools...)

	// Generate memory bucket if not set
	id, err := generateAgentID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate agent ID: %w", err)
	}
	if spec.MemoryBucket == "" {
		spec.MemoryBucket = fmt.Sprintf("%s/%s/memory", r.config.MemoryBucketPrefix, id)
	}

	deployment := &AgentDeployment{
		ID:        id,
		Spec:      spec,
		Status:    AgentStatusPending,
		CreatedAt: time.Now(),
	}

	r.mu.Lock()
	r.agents[id] = deployment
	r.mu.Unlock()

	logging.Info("agent deployment created",
		"agent_id", id,
		"framework", string(spec.Framework),
		"owner", spec.Owner,
		logging.Component("agent"))

	return deployment, nil
}

// Start transitions an agent from pending to running.
func (r *AgentRuntime) Start(ctx context.Context, agentID, containerID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	agent, ok := r.agents[agentID]
	if !ok {
		return fmt.Errorf("agent %q not found", agentID)
	}

	if agent.Status != AgentStatusPending && agent.Status != AgentStatusStarting {
		return fmt.Errorf("agent %q is %s, cannot start", agentID, agent.Status)
	}

	agent.Status = AgentStatusRunning
	agent.ContainerID = containerID
	agent.StartedAt = time.Now()

	logging.Info("agent started",
		"agent_id", agentID,
		"container_id", containerID,
		logging.Component("agent"))

	return nil
}

// Stop stops a running agent.
func (r *AgentRuntime) Stop(ctx context.Context, agentID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	agent, ok := r.agents[agentID]
	if !ok {
		return fmt.Errorf("agent %q not found", agentID)
	}

	if agent.Status == AgentStatusStopped {
		return nil
	}

	agent.Status = AgentStatusStopped
	agent.StoppedAt = time.Now()

	logging.Info("agent stopped",
		"agent_id", agentID,
		logging.Component("agent"))

	return nil
}

// RecordInvocation records token usage from an agent invocation.
func (r *AgentRuntime) RecordInvocation(agentID string, tokensUsed int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	agent, ok := r.agents[agentID]
	if !ok {
		return fmt.Errorf("agent %q not found", agentID)
	}

	agent.TokensUsed += tokensUsed
	agent.InvocationCount++

	// Check budget cap
	if agent.Spec.MaxTokens > 0 && agent.TokensUsed >= agent.Spec.MaxTokens {
		agent.Status = AgentStatusStopped
		agent.StoppedAt = time.Now()
		agent.Error = "token budget exceeded"

		logging.Warn("agent auto-stopped: token budget exceeded",
			"agent_id", agentID,
			"tokens_used", agent.TokensUsed,
			"max_tokens", agent.Spec.MaxTokens,
			logging.Component("agent"))
	}

	return nil
}

// Get returns a copy of an agent deployment.
func (r *AgentRuntime) Get(agentID string) (*AgentDeployment, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	agent, ok := r.agents[agentID]
	if !ok {
		return nil, false
	}

	cp := *agent
	return &cp, true
}

// List returns all agents for a wallet, or all agents if wallet is empty.
func (r *AgentRuntime) List(wallet string) []AgentDeployment {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []AgentDeployment
	for _, a := range r.agents {
		if wallet == "" || a.Spec.Owner == wallet {
			cp := *a
			result = append(result, cp)
		}
	}
	return result
}

// Delete removes an agent from tracking. It must be stopped first.
func (r *AgentRuntime) Delete(agentID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	agent, ok := r.agents[agentID]
	if !ok {
		return fmt.Errorf("agent %q not found", agentID)
	}

	if agent.Status == AgentStatusRunning || agent.Status == AgentStatusStarting {
		return fmt.Errorf("agent %q must be stopped before deletion", agentID)
	}

	delete(r.agents, agentID)
	return nil
}

// SetError marks an agent as failed with an error message.
func (r *AgentRuntime) SetError(agentID, errMsg string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if agent, ok := r.agents[agentID]; ok {
		agent.Status = AgentStatusFailed
		agent.Error = errMsg
		agent.StoppedAt = time.Now()
	}
}

// Stats returns aggregate runtime statistics.
func (r *AgentRuntime) Stats() RuntimeStats {
	r.mu.RLock()
	defer r.mu.RUnlock()

	stats := RuntimeStats{}
	for _, a := range r.agents {
		stats.TotalAgents++
		switch a.Status {
		case AgentStatusRunning:
			stats.RunningAgents++
		case AgentStatusStopped:
			stats.StoppedAgents++
		case AgentStatusFailed:
			stats.FailedAgents++
		}
		stats.TotalTokensUsed += a.TokensUsed
		stats.TotalInvocations += a.InvocationCount
	}
	return stats
}

// RuntimeStats provides aggregate agent metrics.
type RuntimeStats struct {
	TotalAgents      int   `json:"total_agents"`
	RunningAgents    int   `json:"running_agents"`
	StoppedAgents    int   `json:"stopped_agents"`
	FailedAgents     int   `json:"failed_agents"`
	TotalTokensUsed  int64 `json:"total_tokens_used"`
	TotalInvocations int64 `json:"total_invocations"`
}

func generateAgentID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate agent ID: %w", err)
	}
	return "agent-" + hex.EncodeToString(b)[:12], nil
}
