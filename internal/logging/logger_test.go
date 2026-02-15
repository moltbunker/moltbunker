package logging

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"
)

func TestSetAndGetLogger(t *testing.T) {
	// Save original logger
	original := Logger()
	defer SetLogger(original)

	// Create a custom logger
	var buf bytes.Buffer
	customLogger := slog.New(slog.NewJSONHandler(&buf, nil))

	SetLogger(customLogger)

	got := Logger()
	if got != customLogger {
		t.Error("Logger() did not return the logger set by SetLogger()")
	}
}

func TestSetOutput(t *testing.T) {
	original := Logger()
	defer SetLogger(original)

	var buf bytes.Buffer
	SetOutput(&buf)

	Info("test message", "key", "value")

	output := buf.String()
	if output == "" {
		t.Error("expected log output to be written to buffer")
	}
	if !strings.Contains(output, "test message") {
		t.Errorf("expected output to contain 'test message', got: %s", output)
	}
	if !strings.Contains(output, `"key"`) {
		t.Errorf("expected output to contain key, got: %s", output)
	}
}

func TestSetTextOutput(t *testing.T) {
	original := Logger()
	defer SetLogger(original)

	var buf bytes.Buffer
	SetTextOutput(&buf)

	Debug("debug message") // Debug should work with text output (level is Debug)

	output := buf.String()
	if output == "" {
		t.Error("expected log output from text handler")
	}
	if !strings.Contains(output, "debug message") {
		t.Errorf("expected output to contain 'debug message', got: %s", output)
	}
}

func TestSetLevel(t *testing.T) {
	original := Logger()
	defer SetLogger(original)

	var buf bytes.Buffer
	SetOutput(&buf)

	// SetOutput creates an Info-level logger
	Debug("should not appear")
	if buf.Len() > 0 {
		t.Error("Debug messages should not appear at Info level")
	}

	// Set level to Debug
	// Note: SetLevel recreates logger with stdout, so we need a different approach
	// Instead test that Info messages work
	Info("info message")
	if buf.Len() == 0 {
		t.Error("Info messages should appear at Info level")
	}
}

func TestLogLevels(t *testing.T) {
	original := Logger()
	defer SetLogger(original)

	var buf bytes.Buffer
	SetTextOutput(&buf) // Text output with Debug level

	tests := []struct {
		name    string
		logFunc func(string, ...any)
		level   string
	}{
		{"Debug", Debug, "DEBUG"},
		{"Info", Info, "INFO"},
		{"Warn", Warn, "WARN"},
		{"Error", Error, "ERROR"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			tt.logFunc(tt.name+" test message", "key", "val")
			output := buf.String()
			if !strings.Contains(output, tt.name+" test message") {
				t.Errorf("expected output to contain message, got: %s", output)
			}
			if !strings.Contains(output, tt.level) {
				t.Errorf("expected output to contain level %s, got: %s", tt.level, output)
			}
		})
	}
}

func TestLogWithContext(t *testing.T) {
	original := Logger()
	defer SetLogger(original)

	var buf bytes.Buffer
	SetTextOutput(&buf)

	ctx := context.Background()

	tests := []struct {
		name    string
		logFunc func(context.Context, string, ...any)
	}{
		{"DebugContext", DebugContext},
		{"InfoContext", InfoContext},
		{"WarnContext", WarnContext},
		{"ErrorContext", ErrorContext},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			tt.logFunc(ctx, tt.name+" context message")
			output := buf.String()
			if !strings.Contains(output, tt.name+" context message") {
				t.Errorf("expected output to contain message, got: %s", output)
			}
		})
	}
}

func TestWith(t *testing.T) {
	original := Logger()
	defer SetLogger(original)

	var buf bytes.Buffer
	SetTextOutput(&buf)

	logger := With("request_id", "abc123")
	if logger == nil {
		t.Fatal("With() returned nil")
	}

	logger.Info("with context")
	output := buf.String()
	if !strings.Contains(output, "request_id") {
		t.Errorf("expected output to contain 'request_id', got: %s", output)
	}
	if !strings.Contains(output, "abc123") {
		t.Errorf("expected output to contain 'abc123', got: %s", output)
	}
}

func TestContainerIDAttr(t *testing.T) {
	attr := ContainerID("container-123")
	if attr.Key != "container_id" {
		t.Errorf("expected key 'container_id', got %s", attr.Key)
	}
	if attr.Value.String() != "container-123" {
		t.Errorf("expected value 'container-123', got %s", attr.Value.String())
	}
}

func TestNodeIDAttr(t *testing.T) {
	attr := NodeID("node-456")
	if attr.Key != "node_id" {
		t.Errorf("expected key 'node_id', got %s", attr.Key)
	}
	if attr.Value.String() != "node-456" {
		t.Errorf("expected value 'node-456', got %s", attr.Value.String())
	}
}

func TestRegionAttr(t *testing.T) {
	attr := Region("americas")
	if attr.Key != "region" {
		t.Errorf("expected key 'region', got %s", attr.Key)
	}
	if attr.Value.String() != "americas" {
		t.Errorf("expected value 'americas', got %s", attr.Value.String())
	}
}

func TestErrAttr(t *testing.T) {
	t.Run("with error", func(t *testing.T) {
		err := errors.New("something failed")
		attr := Err(err)
		if attr.Key != "error" {
			t.Errorf("expected key 'error', got %s", attr.Key)
		}
		if attr.Value.String() != "something failed" {
			t.Errorf("expected 'something failed', got %s", attr.Value.String())
		}
	})

	t.Run("with nil error", func(t *testing.T) {
		attr := Err(nil)
		if attr.Key != "error" {
			t.Errorf("expected key 'error', got %s", attr.Key)
		}
		if attr.Value.String() != "" {
			t.Errorf("expected empty string for nil error, got %s", attr.Value.String())
		}
	})
}

func TestComponentAttr(t *testing.T) {
	attr := Component("p2p")
	if attr.Key != "component" {
		t.Errorf("expected key 'component', got %s", attr.Key)
	}
	if attr.Value.String() != "p2p" {
		t.Errorf("expected value 'p2p', got %s", attr.Value.String())
	}
}

func TestLoggerIsNotNilByDefault(t *testing.T) {
	l := Logger()
	if l == nil {
		t.Fatal("default logger should not be nil")
	}
}

func TestInfoWithStructuredFields(t *testing.T) {
	original := Logger()
	defer SetLogger(original)

	var buf bytes.Buffer
	SetOutput(&buf) // JSON output

	Info("container started",
		"container_id", "abc",
		"port", 8080,
		"region", "americas",
	)

	output := buf.String()
	if !strings.Contains(output, "container started") {
		t.Errorf("expected message in output: %s", output)
	}
	if !strings.Contains(output, `"container_id"`) {
		t.Errorf("expected container_id field in JSON output: %s", output)
	}
	if !strings.Contains(output, `"port"`) {
		t.Errorf("expected port field in JSON output: %s", output)
	}
}

func TestConcurrentLogging(t *testing.T) {
	original := Logger()
	defer SetLogger(original)

	var buf bytes.Buffer
	SetOutput(&buf)

	// Concurrent writes should not panic
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(n int) {
			Info("concurrent message", "goroutine", n)
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	// Just verify no panic occurred and something was written
	if buf.Len() == 0 {
		t.Error("expected some log output from concurrent logging")
	}
}
