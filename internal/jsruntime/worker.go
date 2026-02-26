package jsruntime

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"time"

	"github.com/moltbunker/moltbunker/internal/logging"
	"github.com/moltbunker/moltbunker/internal/molt"
	"github.com/moltbunker/moltbunker/internal/storage"
)

// DenoWorker manages a single Deno subprocess communicating via stdio JSON-RPC.
type DenoWorker struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout *bufio.Reader
	stderr *bufio.Reader

	mu       sync.Mutex
	alive    bool
	id       int
	cfg      DenoConfig
	bindPath string // path to embedded bindings.ts
}

// NewDenoWorker spawns a Deno subprocess running the bindings runtime.
func NewDenoWorker(id int, cfg DenoConfig, bindingsPath string) (*DenoWorker, error) {
	denoPath := cfg.DenoPath
	if denoPath == "" {
		denoPath = "deno"
	}

	args := []string{
		"run",
		"--no-prompt",
		fmt.Sprintf("--v8-flags=--max-heap-size=%d", cfg.MaxMemoryMB),
		// Permissions: only read the bindings script and user scripts
		"--allow-read",
		// Network: controlled by host_call dispatch (no direct Deno network)
		// No --allow-net: all HTTP goes through Go-side host_calls
		bindingsPath,
	}

	cmd := exec.Command(denoPath, args...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("create stdin pipe: %w", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		stdin.Close()
		return nil, fmt.Errorf("create stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		stdin.Close()
		return nil, fmt.Errorf("create stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		stdin.Close()
		return nil, fmt.Errorf("start deno: %w", err)
	}

	w := &DenoWorker{
		cmd:      cmd,
		stdin:    stdin,
		stdout:   bufio.NewReaderSize(stdout, 64*1024),
		stderr:   bufio.NewReaderSize(stderr, 4*1024),
		alive:    true,
		id:       id,
		cfg:      cfg,
		bindPath: bindingsPath,
	}

	// Drain stderr in background
	go w.drainStderr()

	logging.Debug("deno worker started", "worker_id", id, "pid", cmd.Process.Pid)
	return w, nil
}

// Invoke executes a JS function invocation and handles host_call messages.
func (w *DenoWorker) Invoke(ctx context.Context, invocation JSInvocation, services *molt.HostServices) (*JSResult, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.alive {
		return nil, fmt.Errorf("worker %d is dead", w.id)
	}

	start := time.Now()

	// Encode body
	bodyB64 := ""
	if len(invocation.Body) > 0 {
		bodyB64 = base64.StdEncoding.EncodeToString(invocation.Body)
	}

	// Send invoke message
	invokeData, _ := json.Marshal(InvokeData{
		ScriptPath: invocation.ScriptPath,
		Request: InvokeRequest{
			Method:  invocation.Method,
			URL:     invocation.URL,
			Headers: invocation.Headers,
			Body:    bodyB64,
		},
	})

	err := WriteMessage(w.stdin, &Message{
		Type: MsgTypeInvoke,
		Data: invokeData,
	})
	if err != nil {
		w.alive = false
		return nil, fmt.Errorf("send invoke: %w", err)
	}

	// Read messages until we get a response
	for {
		select {
		case <-ctx.Done():
			return &JSResult{
				StatusCode: 504,
				Duration:   time.Since(start),
				Error:      "invocation timed out",
			}, nil
		default:
		}

		msg, err := ReadMessage(w.stdout)
		if err != nil {
			w.alive = false
			return nil, fmt.Errorf("read message: %w", err)
		}

		switch msg.Type {
		case MsgTypeHostCall:
			// Dispatch host_call to HostServices
			result, hostErr := w.dispatchHostCall(ctx, msg, services)

			resultMsg := &Message{
				Type: MsgTypeHostResult,
				ID:   msg.ID,
			}
			if hostErr != nil {
				resultMsg.Error = hostErr.Error()
			} else {
				resultData, _ := json.Marshal(result)
				resultMsg.Data = resultData
			}

			if err := WriteMessage(w.stdin, resultMsg); err != nil {
				w.alive = false
				return nil, fmt.Errorf("send host_result: %w", err)
			}

		case MsgTypeResponse:
			// Final response from the JS handler
			var body []byte
			if msg.Body != "" {
				body = []byte(msg.Body)
			}

			status := msg.Status
			if status == 0 {
				status = 200
			}

			result := &JSResult{
				StatusCode: status,
				Headers:    msg.Headers,
				Body:       body,
				Duration:   time.Since(start),
			}
			if msg.Error != "" {
				result.Error = msg.Error
				if status == 0 || status == 200 {
					result.StatusCode = 500
				}
			}
			return result, nil

		case MsgTypePong:
			// Ignore pong messages during invocation
			continue

		default:
			logging.Warn("deno worker: unexpected message type", "type", msg.Type, "worker_id", w.id)
		}
	}
}

// dispatchHostCall routes a host_call from Deno to the appropriate HostServices method.
func (w *DenoWorker) dispatchHostCall(ctx context.Context, msg *Message, services *molt.HostServices) (any, error) {
	if services == nil {
		return nil, fmt.Errorf("no host services available")
	}

	switch msg.Fn {
	case "http_request":
		var req molt.HostHTTPRequest
		if err := json.Unmarshal(msg.Args, &req); err != nil {
			return nil, fmt.Errorf("parse http_request args: %w", err)
		}
		return molt.ExecuteHTTPRequest(ctx, services, &req)

	case "storage_put", "storage_get", "storage_delete", "storage_list":
		return w.dispatchStorageCall(ctx, msg.Fn, msg.Args, services)

	case "crawl_page":
		var req molt.CrawlRequest
		if err := json.Unmarshal(msg.Args, &req); err != nil {
			return nil, fmt.Errorf("parse crawl_page args: %w", err)
		}
		return molt.ExecuteCrawl(ctx, services, &req)

	default:
		return nil, fmt.Errorf("unknown host function: %s", msg.Fn)
	}
}

// dispatchStorageCall handles storage_* host_calls.
func (w *DenoWorker) dispatchStorageCall(ctx context.Context, fn string, args json.RawMessage, services *molt.HostServices) (any, error) {
	if !services.Config.StorageEnabled || services.Storage == nil {
		return nil, fmt.Errorf("storage: service disabled")
	}

	var req storageCallArgs
	if err := json.Unmarshal(args, &req); err != nil {
		return nil, fmt.Errorf("parse %s args: %w", fn, err)
	}

	switch fn {
	case "storage_put":
		bodyBytes, err := base64.StdEncoding.DecodeString(req.Data)
		if err != nil {
			return nil, fmt.Errorf("decode put body: %w", err)
		}
		ct := req.ContentType
		if ct == "" {
			ct = "application/octet-stream"
		}
		return services.Storage.PutObject(ctx, &storage.PutObjectInput{
			Bucket:      req.Bucket,
			Key:         req.Key,
			Body:        bytes.NewReader(bodyBytes),
			ContentType: ct,
			Owner:       services.Owner,
			Size:        int64(len(bodyBytes)),
		})

	case "storage_get":
		output, err := services.Storage.GetObject(ctx, req.Bucket, req.Key)
		if err != nil {
			return nil, err
		}
		defer output.Body.Close()
		body, err := io.ReadAll(io.LimitReader(output.Body, 10*1024*1024))
		if err != nil {
			return nil, fmt.Errorf("read object body: %w", err)
		}
		return map[string]any{
			"data":         base64.StdEncoding.EncodeToString(body),
			"content_type": output.ContentType,
			"size":         output.Info.Size,
		}, nil

	case "storage_delete":
		return nil, services.Storage.DeleteObject(ctx, req.Bucket, req.Key, services.Owner)

	case "storage_list":
		maxKeys := 1000
		return services.Storage.ListObjects(ctx, &storage.ListObjectsInput{
			Bucket:  req.Bucket,
			Prefix:  req.Prefix,
			MaxKeys: maxKeys,
		})

	default:
		return nil, fmt.Errorf("unknown storage function: %s", fn)
	}
}

// Alive returns whether the worker process is still running.
func (w *DenoWorker) Alive() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.alive
}

// Close gracefully shuts down the Deno worker.
func (w *DenoWorker) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.alive {
		return nil
	}
	w.alive = false

	// Try graceful shutdown
	_ = WriteMessage(w.stdin, &Message{Type: MsgTypeShutdown})
	w.stdin.Close()

	// Wait with timeout
	done := make(chan error, 1)
	go func() { done <- w.cmd.Wait() }()

	select {
	case err := <-done:
		return err
	case <-time.After(5 * time.Second):
		_ = w.cmd.Process.Kill()
		return fmt.Errorf("worker %d: killed after 5s timeout", w.id)
	}
}

// drainStderr reads stderr and logs it. Prevents pipe buffer from filling.
func (w *DenoWorker) drainStderr() {
	for {
		line, err := w.stderr.ReadBytes('\n')
		if err != nil {
			return
		}
		if len(line) > 0 {
			logging.Debug("deno stderr", "worker_id", w.id, "output", string(line))
		}
	}
}

// storageCallArgs matches the JSON args from Deno storage host_calls.
type storageCallArgs struct {
	Bucket      string `json:"bucket"`
	Key         string `json:"key"`
	Data        string `json:"data,omitempty"`         // base64-encoded body (put)
	ContentType string `json:"content_type,omitempty"` // put only
	Prefix      string `json:"prefix,omitempty"`       // list only
}

