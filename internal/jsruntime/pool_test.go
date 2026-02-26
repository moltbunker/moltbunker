package jsruntime

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/moltbunker/moltbunker/internal/molt"
)

// requireDeno skips the test if Deno is not installed.
func requireDeno(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("deno"); err != nil {
		t.Skip("deno not installed, skipping JS runtime test")
	}
}

// writeTestScript writes a JS script to a temp file and returns its absolute path.
func writeTestScript(t *testing.T, code string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "handler.ts")
	if err := os.WriteFile(path, []byte(code), 0o644); err != nil {
		t.Fatalf("write test script: %v", err)
	}
	return path
}

// setupBindings writes bindings.ts to a temp dir and returns the path.
func setupBindings(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	path, err := WriteBindingsFile(dir)
	if err != nil {
		t.Fatalf("WriteBindingsFile: %v", err)
	}
	return path
}

func TestProtocol_WriteReadMessage(t *testing.T) {
	var buf bytes.Buffer

	msg := &Message{
		Type:   MsgTypeInvoke,
		Status: 200,
		Body:   "hello",
	}

	if err := WriteMessage(&buf, msg); err != nil {
		t.Fatalf("WriteMessage: %v", err)
	}

	reader := bufio.NewReader(&buf)
	got, err := ReadMessage(reader)
	if err != nil {
		t.Fatalf("ReadMessage: %v", err)
	}

	if got.Type != MsgTypeInvoke {
		t.Fatalf("Type = %q, want %q", got.Type, MsgTypeInvoke)
	}
	if got.Body != "hello" {
		t.Fatalf("Body = %q, want %q", got.Body, "hello")
	}
}

func TestProtocol_MessageTypes(t *testing.T) {
	types := []string{
		MsgTypeInvoke, MsgTypeHostCall, MsgTypeHostResult,
		MsgTypeResponse, MsgTypePing, MsgTypePong, MsgTypeShutdown,
	}

	for _, typ := range types {
		var buf bytes.Buffer
		msg := &Message{Type: typ}
		if err := WriteMessage(&buf, msg); err != nil {
			t.Fatalf("WriteMessage(%s): %v", typ, err)
		}

		reader := bufio.NewReader(&buf)
		got, err := ReadMessage(reader)
		if err != nil {
			t.Fatalf("ReadMessage(%s): %v", typ, err)
		}
		if got.Type != typ {
			t.Fatalf("Type = %q, want %q", got.Type, typ)
		}
	}
}

func TestProtocol_HostCallMessage(t *testing.T) {
	var buf bytes.Buffer

	args, _ := json.Marshal(map[string]string{"url": "https://example.com"})
	msg := &Message{
		Type: MsgTypeHostCall,
		ID:   42,
		Fn:   "http_request",
		Args: args,
	}

	if err := WriteMessage(&buf, msg); err != nil {
		t.Fatalf("WriteMessage: %v", err)
	}

	reader := bufio.NewReader(&buf)
	got, err := ReadMessage(reader)
	if err != nil {
		t.Fatalf("ReadMessage: %v", err)
	}

	if got.Type != MsgTypeHostCall {
		t.Fatalf("Type = %q", got.Type)
	}
	if got.ID != 42 {
		t.Fatalf("ID = %d, want 42", got.ID)
	}
	if got.Fn != "http_request" {
		t.Fatalf("Fn = %q", got.Fn)
	}
}

func TestDenoConfig_Defaults(t *testing.T) {
	cfg := DefaultDenoConfig()
	if cfg.PoolSize != 10 {
		t.Fatalf("PoolSize = %d, want 10", cfg.PoolSize)
	}
	if cfg.TimeoutMs != 30000 {
		t.Fatalf("TimeoutMs = %d, want 30000", cfg.TimeoutMs)
	}
	if cfg.MaxMemoryMB != 128 {
		t.Fatalf("MaxMemoryMB = %d, want 128", cfg.MaxMemoryMB)
	}
	if cfg.DenoPath != "deno" {
		t.Fatalf("DenoPath = %q, want %q", cfg.DenoPath, "deno")
	}
}

func TestJSInvocation_JSON(t *testing.T) {
	inv := JSInvocation{
		ScriptPath:   "/tmp/handler.ts",
		DeploymentID: "dep-123",
		Method:       "POST",
		URL:          "https://example.com/api",
		Headers:      map[string]string{"Content-Type": "application/json"},
		Body:         []byte(`{"key":"value"}`),
	}

	data, err := json.Marshal(inv)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var parsed JSInvocation
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if parsed.Method != "POST" || parsed.DeploymentID != "dep-123" {
		t.Fatalf("roundtrip failed: %+v", parsed)
	}
}

func TestBindingsEmbed(t *testing.T) {
	if len(bindingsTS) == 0 {
		t.Fatal("embedded bindings.ts is empty")
	}
	if !bytes.Contains(bindingsTS, []byte("globalThis")) {
		t.Fatal("bindings.ts should contain 'globalThis'")
	}
}

func TestWriteBindingsFile(t *testing.T) {
	dir := t.TempDir()
	path, err := WriteBindingsFile(dir)
	if err != nil {
		t.Fatalf("WriteBindingsFile: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if !bytes.Equal(data, bindingsTS) {
		t.Fatal("written file doesn't match embedded bindings")
	}
}

// --- Integration tests (require Deno) ---

func TestDenoWorker_SpawnAndClose(t *testing.T) {
	requireDeno(t)
	bindingsPath := setupBindings(t)

	cfg := DenoConfig{
		DenoPath:    "deno",
		MaxMemoryMB: 64,
	}

	w, err := NewDenoWorker(1, cfg, bindingsPath)
	if err != nil {
		t.Fatalf("NewDenoWorker: %v", err)
	}
	defer w.Close()

	if !w.Alive() {
		t.Fatal("worker should be alive after spawn")
	}

	if err := w.Close(); err != nil {
		// Deno may exit with non-zero when stdin closes, that's OK
		t.Logf("Close returned: %v (expected for stdin-driven exit)", err)
	}
}

func TestDenoPool_BasicExecution(t *testing.T) {
	requireDeno(t)
	bindingsPath := setupBindings(t)

	// Write a simple handler
	scriptPath := writeTestScript(t, `
export default function handler(request) {
  return new Response(JSON.stringify({ method: request.method, url: request.url }), {
    status: 200,
    headers: { "Content-Type": "application/json" },
  });
}
`)

	cfg := DenoConfig{
		Enabled:     true,
		DenoPath:    "deno",
		PoolSize:    2,
		TimeoutMs:   10000,
		MaxMemoryMB: 64,
	}

	svc := molt.NewHostServices(molt.HostCapabilities{})
	pool, err := NewDenoPool(cfg, bindingsPath, svc)
	if err != nil {
		t.Fatalf("NewDenoPool: %v", err)
	}
	defer pool.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := pool.Execute(ctx, JSInvocation{
		ScriptPath:   scriptPath,
		DeploymentID: "test-deploy",
		Method:       "GET",
		URL:          "http://localhost/test",
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if result.StatusCode != 200 {
		t.Fatalf("StatusCode = %d, want 200 (body: %s, error: %s)", result.StatusCode, string(result.Body), result.Error)
	}
	if result.Duration <= 0 {
		t.Fatal("Duration should be positive")
	}
}

func TestDenoPool_ErrorHandler(t *testing.T) {
	requireDeno(t)
	bindingsPath := setupBindings(t)

	scriptPath := writeTestScript(t, `
export default function handler(request) {
  throw new Error("intentional test error");
}
`)

	cfg := DenoConfig{
		Enabled:     true,
		DenoPath:    "deno",
		PoolSize:    1,
		TimeoutMs:   10000,
		MaxMemoryMB: 64,
	}

	svc := molt.NewHostServices(molt.HostCapabilities{})
	pool, err := NewDenoPool(cfg, bindingsPath, svc)
	if err != nil {
		t.Fatalf("NewDenoPool: %v", err)
	}
	defer pool.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := pool.Execute(ctx, JSInvocation{
		ScriptPath:   scriptPath,
		DeploymentID: "test-error",
		Method:       "GET",
		URL:          "http://localhost/error",
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if result.StatusCode != 500 {
		t.Fatalf("StatusCode = %d, want 500 (body: %s)", result.StatusCode, string(result.Body))
	}
	if result.Error == "" {
		t.Fatal("expected error message")
	}
}

func TestDenoPool_ClosedPool(t *testing.T) {
	requireDeno(t)
	bindingsPath := setupBindings(t)

	cfg := DenoConfig{
		Enabled:     true,
		DenoPath:    "deno",
		PoolSize:    1,
		TimeoutMs:   5000,
		MaxMemoryMB: 64,
	}

	svc := molt.NewHostServices(molt.HostCapabilities{})
	pool, err := NewDenoPool(cfg, bindingsPath, svc)
	if err != nil {
		t.Fatalf("NewDenoPool: %v", err)
	}

	pool.Close()

	_, err = pool.Execute(context.Background(), JSInvocation{
		ScriptPath: "/nonexistent.ts",
		Method:     "GET",
	})
	if err == nil {
		t.Fatal("expected error from closed pool")
	}
}
