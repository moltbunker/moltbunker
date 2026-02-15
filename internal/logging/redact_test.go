package logging

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

// newTestRedactingLogger creates a RedactingHandler wrapping a JSON handler
// that writes to the given buffer.
func newTestRedactingLogger(buf *bytes.Buffer) *slog.Logger {
	inner := slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	return slog.New(NewRedactingHandler(inner))
}

func TestRedact_NormalValuesPassThrough(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestRedactingLogger(&buf)

	logger.Info("test message",
		"container_id", "mb-123456",
		"region", "americas",
		"port", 8080,
		"status", "running",
	)

	output := buf.String()

	// All normal values should appear unchanged
	for _, expected := range []string{"mb-123456", "americas", "8080", "running"} {
		if !strings.Contains(output, expected) {
			t.Errorf("expected output to contain %q, got: %s", expected, output)
		}
	}

	// No redaction markers
	if strings.Contains(output, "[REDACTED]") {
		t.Errorf("normal values should not be redacted, got: %s", output)
	}
}

func TestRedact_APIKeys(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		value    string
		contains string
		absent   string
	}{
		{
			name:     "live API key",
			key:      "api_key",
			value:    "mb_live_abc12345xyz67890longkeyherefortest",
			contains: "mb_live_abc12345...",
			absent:   "xyz67890longkeyherefortest",
		},
		{
			name:     "test API key",
			key:      "auth",
			value:    "mb_test_def99887xyz65432anotherlongkey",
			contains: "mb_test_def99887...",
			absent:   "xyz65432anotherlongkey",
		},
		{
			name:     "API key in sentence",
			key:      "message",
			value:    "used key mb_live_abc12345xyz67890longkeyherefortest for auth",
			contains: "mb_live_abc12345...",
			absent:   "xyz67890longkeyherefortest",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := newTestRedactingLogger(&buf)

			logger.Info("test", tt.key, tt.value)

			output := buf.String()
			if !strings.Contains(output, tt.contains) {
				t.Errorf("expected output to contain %q, got: %s", tt.contains, output)
			}
			if strings.Contains(output, tt.absent) {
				t.Errorf("expected output NOT to contain %q, got: %s", tt.absent, output)
			}
		})
	}
}

func TestRedact_EthereumPrivateKeys(t *testing.T) {
	ethKey := "0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"

	var buf bytes.Buffer
	logger := newTestRedactingLogger(&buf)

	logger.Info("wallet loaded", "key_data", ethKey)

	output := buf.String()

	// The full key should not appear
	if strings.Contains(output, ethKey) {
		t.Errorf("full Ethereum private key should be redacted, got: %s", output)
	}

	// Should contain the partial redaction (first 6 chars + ... + last 4 chars)
	if !strings.Contains(output, "0xac09") {
		t.Errorf("expected partial key prefix '0xac09' in output, got: %s", output)
	}
	if !strings.Contains(output, "ff80") {
		t.Errorf("expected partial key suffix 'ff80' in output, got: %s", output)
	}
}

func TestRedact_LongHexStrings(t *testing.T) {
	// 65+ character hex string that looks like a private key
	longHex := "aabbccdd" + strings.Repeat("ee", 30) + "ff1234"

	var buf bytes.Buffer
	logger := newTestRedactingLogger(&buf)

	logger.Info("key loaded", "data", longHex)

	output := buf.String()

	// Full hex string should not appear
	if strings.Contains(output, longHex) {
		t.Errorf("long hex string should be redacted, got: %s", output)
	}

	// Should contain [REDACTED] marker
	if !strings.Contains(output, "[REDACTED]") {
		t.Errorf("expected [REDACTED] marker in output, got: %s", output)
	}
}

func TestRedact_SensitiveFieldNames(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value string
	}{
		{"password field", "password", "my-secret-password-123"},
		{"db_password field", "db_password", "postgres123"},
		{"secret field", "secret", "very-secret-value"},
		{"client_secret field", "client_secret", "oauth-secret-xyz"},
		{"private_key field", "private_key", "-----BEGIN PRIVATE KEY-----"},
		{"ssh_private_key field", "ssh_private_key", "base64encodedkey=="},
		{"credential field", "credential", "some-credential"},
		{"api_credential field", "api_credential", "cred-abc123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := newTestRedactingLogger(&buf)

			logger.Info("test", tt.key, tt.value)

			output := buf.String()

			if strings.Contains(output, tt.value) {
				t.Errorf("sensitive value for key %q should be redacted, got: %s", tt.key, output)
			}
			if !strings.Contains(output, "[REDACTED]") {
				t.Errorf("expected [REDACTED] for key %q, got: %s", tt.key, output)
			}
		})
	}
}

func TestRedact_TokenFieldWithSecretValue(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestRedactingLogger(&buf)

	// A "token" field with an API key value should be redacted
	logger.Info("auth", "token", "mb_live_abc12345xyz67890secretkey")

	output := buf.String()
	if strings.Contains(output, "xyz67890secretkey") {
		t.Errorf("token field with API key value should be redacted, got: %s", output)
	}
}

func TestRedact_TokenFieldWithNonSecretValue(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestRedactingLogger(&buf)

	// A "token_amount" field with a numeric value should NOT be redacted
	logger.Info("payment", "token_amount", "500")

	output := buf.String()
	if !strings.Contains(output, "500") {
		t.Errorf("token_amount field should not be redacted, got: %s", output)
	}
	if strings.Contains(output, "[REDACTED]") {
		t.Errorf("token_amount should not trigger redaction, got: %s", output)
	}
}

func TestRedact_EnableRedaction(t *testing.T) {
	// Save and restore original logger
	original := Logger()
	defer SetLogger(original)

	var buf bytes.Buffer
	SetOutput(&buf)

	EnableRedaction()

	Info("auth event", "password", "super-secret-123")

	output := buf.String()
	if strings.Contains(output, "super-secret-123") {
		t.Errorf("password should be redacted after EnableRedaction(), got: %s", output)
	}
	if !strings.Contains(output, "[REDACTED]") {
		t.Errorf("expected [REDACTED] marker after EnableRedaction(), got: %s", output)
	}
}

func TestRedact_EnableRedactionIdempotent(t *testing.T) {
	original := Logger()
	defer SetLogger(original)

	var buf bytes.Buffer
	SetOutput(&buf)

	// Call EnableRedaction twice - should not double-wrap
	EnableRedaction()
	EnableRedaction()

	Info("test", "password", "secret123")

	output := buf.String()
	if strings.Contains(output, "secret123") {
		t.Errorf("password should be redacted, got: %s", output)
	}
	// Count [REDACTED] occurrences - should appear exactly once per attribute
	count := strings.Count(output, "[REDACTED]")
	if count != 1 {
		t.Errorf("expected exactly 1 [REDACTED] occurrence, got %d in: %s", count, output)
	}
}

func TestRedact_NewRedactingLogger(t *testing.T) {
	var buf bytes.Buffer
	inner := slog.NewJSONHandler(&buf, nil)
	logger := NewRedactingLogger(inner)

	logger.Info("test", "secret", "my-secret-value")

	output := buf.String()
	if strings.Contains(output, "my-secret-value") {
		t.Errorf("secret should be redacted via NewRedactingLogger, got: %s", output)
	}
}

func TestRedact_WithAttrs(t *testing.T) {
	var buf bytes.Buffer
	inner := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	handler := NewRedactingHandler(inner)

	// WithAttrs should also redact sensitive attributes
	childHandler := handler.WithAttrs([]slog.Attr{
		slog.String("password", "persistent-secret"),
	})

	logger := slog.New(childHandler)
	logger.Info("test message")

	output := buf.String()
	if strings.Contains(output, "persistent-secret") {
		t.Errorf("password in WithAttrs should be redacted, got: %s", output)
	}
}

func TestRedact_WithGroup(t *testing.T) {
	var buf bytes.Buffer
	inner := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	handler := NewRedactingHandler(inner)

	childHandler := handler.WithGroup("auth")
	logger := slog.New(childHandler)
	logger.Info("test", "password", "group-secret")

	output := buf.String()
	if strings.Contains(output, "group-secret") {
		t.Errorf("password in group should be redacted, got: %s", output)
	}
}

func TestRedact_MixedSensitiveAndNormal(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestRedactingLogger(&buf)

	logger.Info("request processed",
		"method", "deploy",
		"password", "db-secret",
		"container_id", "mb-123",
		"api_key", "mb_live_abc12345xyz67890secretkey",
		"region", "americas",
	)

	output := buf.String()

	// Normal values should be present
	if !strings.Contains(output, "deploy") {
		t.Error("method value should be present")
	}
	if !strings.Contains(output, "mb-123") {
		t.Error("container_id value should be present")
	}
	if !strings.Contains(output, "americas") {
		t.Error("region value should be present")
	}

	// Sensitive values should be redacted
	if strings.Contains(output, "db-secret") {
		t.Error("password value should be redacted")
	}
	if strings.Contains(output, "xyz67890secretkey") {
		t.Error("full API key should be redacted")
	}
}

func TestRedact_ShortAPIKey(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestRedactingLogger(&buf)

	// A short API key (fewer than 8 chars after prefix) should still be handled
	logger.Info("test", "key", "mb_live_short")

	output := buf.String()
	// Short keys should pass through (they don't have enough chars to redact meaningfully)
	if !strings.Contains(output, "mb_live_short") {
		t.Errorf("short API key should pass through, got: %s", output)
	}
}
