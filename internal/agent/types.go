package agent

import (
	"time"
)

// Framework identifies the AI framework an agent uses.
type Framework string

const (
	FrameworkLangGraph Framework = "langgraph"
	FrameworkCrewAI    Framework = "crewai"
	FrameworkAutoGen   Framework = "autogen"
	FrameworkCustom    Framework = "custom"
)

// AgentStatus tracks the lifecycle state of a deployed agent.
type AgentStatus string

const (
	AgentStatusPending  AgentStatus = "pending"
	AgentStatusStarting AgentStatus = "starting"
	AgentStatusRunning  AgentStatus = "running"
	AgentStatusStopped  AgentStatus = "stopped"
	AgentStatusFailed   AgentStatus = "failed"
)

// AgentSpec describes how to deploy an AI agent.
type AgentSpec struct {
	Name          string            `json:"name"`
	Framework     Framework         `json:"framework"`
	Image         string            `json:"image,omitempty"`   // Container image CID or reference
	Config        map[string]string `json:"config,omitempty"`  // Framework-specific config
	Env           map[string]string `json:"env,omitempty"`     // Environment variables
	MCPTools      []MCPToolDef      `json:"mcp_tools,omitempty"`
	MemoryBucket  string            `json:"memory_bucket,omitempty"` // Object storage bucket for memory
	MaxTokens     int64             `json:"max_tokens,omitempty"`    // Budget cap
	TimeoutSec    int               `json:"timeout_sec,omitempty"`   // Max runtime in seconds (0 = unlimited)
	MemoryLimitMB int               `json:"memory_limit_mb,omitempty"`
	CPUCores      int               `json:"cpu_cores,omitempty"`
	Owner         string            `json:"owner"`
}

// MCPToolDef defines a tool available to the agent via MCP protocol.
type MCPToolDef struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Parameters  map[string]string `json:"parameters,omitempty"` // JSON Schema for parameters
}

// AgentDeployment tracks a deployed agent instance.
type AgentDeployment struct {
	ID          string      `json:"id"`
	Spec        AgentSpec   `json:"spec"`
	Status      AgentStatus `json:"status"`
	ContainerID string      `json:"container_id,omitempty"`
	NodeID      string      `json:"node_id,omitempty"`
	CreatedAt   time.Time   `json:"created_at"`
	StartedAt   time.Time   `json:"started_at,omitempty"`
	StoppedAt   time.Time   `json:"stopped_at,omitempty"`
	Error       string      `json:"error,omitempty"`

	// Metrics
	TokensUsed    int64   `json:"tokens_used"`
	InvocationCount int64 `json:"invocation_count"`
	TotalCostWei  string  `json:"total_cost_wei,omitempty"`
}

// AgentInvokeRequest is a message sent to an agent.
type AgentInvokeRequest struct {
	AgentID string `json:"agent_id"`
	Message string `json:"message"`
	Context map[string]string `json:"context,omitempty"`
}

// AgentInvokeResponse is the agent's reply.
type AgentInvokeResponse struct {
	AgentID    string `json:"agent_id"`
	Response   string `json:"response"`
	TokensUsed int64  `json:"tokens_used"`
	DurationMs int64  `json:"duration_ms"`
	Error      string `json:"error,omitempty"`
}

// MemoryEntry represents a key-value entry in an agent's persistent memory.
type MemoryEntry struct {
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	UpdatedAt time.Time `json:"updated_at"`
}

// BuiltinMCPTools returns the default MCP tools available to all agents.
func BuiltinMCPTools() []MCPToolDef {
	return []MCPToolDef{
		{
			Name:        "storage_read",
			Description: "Read a value from agent persistent memory (Object Storage)",
			Parameters:  map[string]string{"bucket": "string", "key": "string"},
		},
		{
			Name:        "storage_write",
			Description: "Write a value to agent persistent memory (Object Storage)",
			Parameters:  map[string]string{"bucket": "string", "key": "string", "data": "string"},
		},
		{
			Name:        "web_fetch",
			Description: "Fetch a URL via the decentralized proxy",
			Parameters:  map[string]string{"url": "string", "method": "string", "headers": "object", "body": "string"},
		},
		{
			Name:        "exec_command",
			Description: "Execute a command inside the agent's container",
			Parameters:  map[string]string{"cmd": "string"},
		},
	}
}

// FrameworkDefaults returns the default image and env vars for a framework.
func FrameworkDefaults(fw Framework) (image string, env map[string]string) {
	switch fw {
	case FrameworkLangGraph:
		return "python:3.12-slim", map[string]string{
			"FRAMEWORK": "langgraph",
		}
	case FrameworkCrewAI:
		return "python:3.12-slim", map[string]string{
			"FRAMEWORK": "crewai",
		}
	case FrameworkAutoGen:
		return "python:3.12-slim", map[string]string{
			"FRAMEWORK": "autogen",
		}
	case FrameworkCustom:
		return "", nil
	default:
		return "", nil
	}
}
