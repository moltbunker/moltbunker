package agent

import (
	"context"
	"testing"
)

func TestDeploy_Success(t *testing.T) {
	rt := NewAgentRuntime(DefaultRuntimeConfig())
	ctx := context.Background()

	dep, err := rt.Deploy(ctx, AgentSpec{
		Name:      "test-agent",
		Framework: FrameworkLangGraph,
		Owner:     "0x1234",
	})
	if err != nil {
		t.Fatalf("Deploy: %v", err)
	}

	if dep.ID == "" {
		t.Error("agent ID should not be empty")
	}
	if dep.Status != AgentStatusPending {
		t.Errorf("status = %q, want pending", dep.Status)
	}
	if dep.Spec.Image == "" {
		t.Error("image should be set from framework defaults")
	}
	if dep.Spec.MemoryBucket == "" {
		t.Error("memory bucket should be auto-generated")
	}
}

func TestDeploy_CustomFramework(t *testing.T) {
	rt := NewAgentRuntime(DefaultRuntimeConfig())
	ctx := context.Background()

	dep, err := rt.Deploy(ctx, AgentSpec{
		Framework: FrameworkCustom,
		Image:     "my-agent:latest",
		Owner:     "0x1234",
	})
	if err != nil {
		t.Fatalf("Deploy custom: %v", err)
	}
	if dep.Spec.Image != "my-agent:latest" {
		t.Errorf("image = %q, want my-agent:latest", dep.Spec.Image)
	}
}

func TestDeploy_MissingOwner(t *testing.T) {
	rt := NewAgentRuntime(DefaultRuntimeConfig())
	ctx := context.Background()

	_, err := rt.Deploy(ctx, AgentSpec{Framework: FrameworkLangGraph})
	if err == nil {
		t.Error("expected error for missing owner")
	}
}

func TestDeploy_InvalidFramework(t *testing.T) {
	rt := NewAgentRuntime(DefaultRuntimeConfig())
	ctx := context.Background()

	_, err := rt.Deploy(ctx, AgentSpec{Framework: "unknown", Owner: "0x1234"})
	if err == nil {
		t.Error("expected error for invalid framework")
	}
}

func TestDeploy_WalletLimit(t *testing.T) {
	cfg := DefaultRuntimeConfig()
	cfg.MaxAgentsPerWallet = 2
	rt := NewAgentRuntime(cfg)
	ctx := context.Background()

	for i := 0; i < 2; i++ {
		_, err := rt.Deploy(ctx, AgentSpec{Framework: FrameworkCustom, Image: "img", Owner: "w1"})
		if err != nil {
			t.Fatalf("Deploy %d: %v", i, err)
		}
	}

	_, err := rt.Deploy(ctx, AgentSpec{Framework: FrameworkCustom, Image: "img", Owner: "w1"})
	if err == nil {
		t.Error("expected error when exceeding wallet limit")
	}
}

func TestStartAndStop(t *testing.T) {
	rt := NewAgentRuntime(DefaultRuntimeConfig())
	ctx := context.Background()

	dep, _ := rt.Deploy(ctx, AgentSpec{Framework: FrameworkCustom, Image: "img", Owner: "w1"})

	if err := rt.Start(ctx, dep.ID, "ctr-123"); err != nil {
		t.Fatalf("Start: %v", err)
	}

	agent, _ := rt.Get(dep.ID)
	if agent.Status != AgentStatusRunning {
		t.Errorf("status = %q, want running", agent.Status)
	}
	if agent.ContainerID != "ctr-123" {
		t.Errorf("container = %q, want ctr-123", agent.ContainerID)
	}

	if err := rt.Stop(ctx, dep.ID); err != nil {
		t.Fatalf("Stop: %v", err)
	}

	agent, _ = rt.Get(dep.ID)
	if agent.Status != AgentStatusStopped {
		t.Errorf("status = %q, want stopped", agent.Status)
	}
}

func TestRecordInvocation(t *testing.T) {
	rt := NewAgentRuntime(DefaultRuntimeConfig())
	ctx := context.Background()

	dep, _ := rt.Deploy(ctx, AgentSpec{Framework: FrameworkCustom, Image: "img", Owner: "w1"})
	if err := rt.Start(ctx, dep.ID, "ctr-1"); err != nil {
		t.Fatalf("Start: %v", err)
	}

	if err := rt.RecordInvocation(dep.ID, 100); err != nil {
		t.Fatalf("RecordInvocation 100: %v", err)
	}
	if err := rt.RecordInvocation(dep.ID, 200); err != nil {
		t.Fatalf("RecordInvocation 200: %v", err)
	}

	agent, _ := rt.Get(dep.ID)
	if agent.TokensUsed != 300 {
		t.Errorf("tokens = %d, want 300", agent.TokensUsed)
	}
	if agent.InvocationCount != 2 {
		t.Errorf("invocations = %d, want 2", agent.InvocationCount)
	}
}

func TestRecordInvocation_BudgetExceeded(t *testing.T) {
	cfg := DefaultRuntimeConfig()
	cfg.MaxTokenBudget = 100
	rt := NewAgentRuntime(cfg)
	ctx := context.Background()

	dep, _ := rt.Deploy(ctx, AgentSpec{Framework: FrameworkCustom, Image: "img", Owner: "w1"})
	if err := rt.Start(ctx, dep.ID, "ctr-1"); err != nil {
		t.Fatalf("Start: %v", err)
	}

	if err := rt.RecordInvocation(dep.ID, 150); err != nil {
		t.Fatalf("RecordInvocation: %v", err)
	}

	agent, _ := rt.Get(dep.ID)
	if agent.Status != AgentStatusStopped {
		t.Errorf("status = %q, want stopped after budget exceeded", agent.Status)
	}
	if agent.Error == "" {
		t.Error("error should be set when budget exceeded")
	}
}

func TestList(t *testing.T) {
	rt := NewAgentRuntime(DefaultRuntimeConfig())
	ctx := context.Background()

	if _, err := rt.Deploy(ctx, AgentSpec{Framework: FrameworkCustom, Image: "img", Owner: "w1"}); err != nil {
		t.Fatalf("Deploy w1[0]: %v", err)
	}
	if _, err := rt.Deploy(ctx, AgentSpec{Framework: FrameworkCustom, Image: "img", Owner: "w1"}); err != nil {
		t.Fatalf("Deploy w1[1]: %v", err)
	}
	if _, err := rt.Deploy(ctx, AgentSpec{Framework: FrameworkCustom, Image: "img", Owner: "w2"}); err != nil {
		t.Fatalf("Deploy w2: %v", err)
	}

	all := rt.List("")
	if len(all) != 3 {
		t.Errorf("all agents = %d, want 3", len(all))
	}

	w1 := rt.List("w1")
	if len(w1) != 2 {
		t.Errorf("w1 agents = %d, want 2", len(w1))
	}
}

func TestDelete(t *testing.T) {
	rt := NewAgentRuntime(DefaultRuntimeConfig())
	ctx := context.Background()

	dep, _ := rt.Deploy(ctx, AgentSpec{Framework: FrameworkCustom, Image: "img", Owner: "w1"})
	if err := rt.Start(ctx, dep.ID, "ctr-1"); err != nil {
		t.Fatalf("Start: %v", err)
	}

	// Cannot delete running agent
	if err := rt.Delete(dep.ID); err == nil {
		t.Error("expected error deleting running agent")
	}

	if err := rt.Stop(ctx, dep.ID); err != nil {
		t.Fatalf("Stop: %v", err)
	}

	if err := rt.Delete(dep.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, ok := rt.Get(dep.ID)
	if ok {
		t.Error("agent should be gone after delete")
	}
}

func TestSetError(t *testing.T) {
	rt := NewAgentRuntime(DefaultRuntimeConfig())
	ctx := context.Background()

	dep, _ := rt.Deploy(ctx, AgentSpec{Framework: FrameworkCustom, Image: "img", Owner: "w1"})
	rt.SetError(dep.ID, "container crashed")

	agent, _ := rt.Get(dep.ID)
	if agent.Status != AgentStatusFailed {
		t.Errorf("status = %q, want failed", agent.Status)
	}
	if agent.Error != "container crashed" {
		t.Errorf("error = %q", agent.Error)
	}
}

func TestStats(t *testing.T) {
	rt := NewAgentRuntime(DefaultRuntimeConfig())
	ctx := context.Background()

	d1, _ := rt.Deploy(ctx, AgentSpec{Framework: FrameworkCustom, Image: "img", Owner: "w1"})
	d2, _ := rt.Deploy(ctx, AgentSpec{Framework: FrameworkCustom, Image: "img", Owner: "w1"})
	d3, _ := rt.Deploy(ctx, AgentSpec{Framework: FrameworkCustom, Image: "img", Owner: "w1"})

	if err := rt.Start(ctx, d1.ID, "c1"); err != nil {
		t.Fatalf("Start d1: %v", err)
	}
	if err := rt.Start(ctx, d2.ID, "c2"); err != nil {
		t.Fatalf("Start d2: %v", err)
	}
	if err := rt.Stop(ctx, d2.ID); err != nil {
		t.Fatalf("Stop d2: %v", err)
	}
	rt.SetError(d3.ID, "boom")

	if err := rt.RecordInvocation(d1.ID, 500); err != nil {
		t.Fatalf("RecordInvocation: %v", err)
	}

	stats := rt.Stats()
	if stats.TotalAgents != 3 {
		t.Errorf("total = %d, want 3", stats.TotalAgents)
	}
	if stats.RunningAgents != 1 {
		t.Errorf("running = %d, want 1", stats.RunningAgents)
	}
	if stats.StoppedAgents != 1 {
		t.Errorf("stopped = %d, want 1", stats.StoppedAgents)
	}
	if stats.FailedAgents != 1 {
		t.Errorf("failed = %d, want 1", stats.FailedAgents)
	}
	if stats.TotalTokensUsed != 500 {
		t.Errorf("tokens = %d, want 500", stats.TotalTokensUsed)
	}
}

func TestBuiltinMCPTools(t *testing.T) {
	tools := BuiltinMCPTools()
	if len(tools) != 4 {
		t.Fatalf("builtin tools = %d, want 4", len(tools))
	}

	names := make(map[string]bool)
	for _, tool := range tools {
		names[tool.Name] = true
	}

	expected := []string{"storage_read", "storage_write", "web_fetch", "exec_command"}
	for _, name := range expected {
		if !names[name] {
			t.Errorf("missing builtin tool: %s", name)
		}
	}
}

func TestFrameworkDefaults(t *testing.T) {
	image, env := FrameworkDefaults(FrameworkLangGraph)
	if image == "" {
		t.Error("LangGraph should have a default image")
	}
	if env["FRAMEWORK"] != "langgraph" {
		t.Errorf("FRAMEWORK env = %q, want langgraph", env["FRAMEWORK"])
	}

	image, _ = FrameworkDefaults(FrameworkCustom)
	if image != "" {
		t.Error("Custom framework should have no default image")
	}
}

func TestGetReturnsCopy(t *testing.T) {
	rt := NewAgentRuntime(DefaultRuntimeConfig())
	ctx := context.Background()

	dep, _ := rt.Deploy(ctx, AgentSpec{Framework: FrameworkCustom, Image: "img", Owner: "w1"})

	copy1, _ := rt.Get(dep.ID)
	copy1.Status = AgentStatusFailed // mutate copy

	copy2, _ := rt.Get(dep.ID)
	if copy2.Status == AgentStatusFailed {
		t.Error("mutation of copy should not affect original")
	}
}
