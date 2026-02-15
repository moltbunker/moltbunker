package logging

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
)

// sensitiveKeyPatterns lists substrings that indicate a log attribute key holds a secret value.
// Values logged under these keys will be fully redacted.
var sensitiveKeyPatterns = []string{
	"password",
	"secret",
	"private_key",
	"credential",
}

// apiKeyPattern matches Moltbunker API keys (mb_live_* or mb_test_*).
var apiKeyPattern = regexp.MustCompile(`\bmb_(live|test)_[A-Za-z0-9]+`)

// ethPrivateKeyPattern matches Ethereum-style private keys (0x followed by 64 hex chars).
var ethPrivateKeyPattern = regexp.MustCompile(`\b0x[0-9a-fA-F]{64}\b`)

// longHexPattern matches hex strings longer than 64 characters that look like private keys.
var longHexPattern = regexp.MustCompile(`\b[0-9a-fA-F]{65,}\b`)

// RedactingHandler wraps an slog.Handler and redacts sensitive values before they
// are passed to the inner handler.
type RedactingHandler struct {
	inner slog.Handler
}

// NewRedactingHandler creates a RedactingHandler that wraps the given inner handler.
func NewRedactingHandler(inner slog.Handler) *RedactingHandler {
	return &RedactingHandler{inner: inner}
}

// Enabled reports whether the inner handler handles records at the given level.
func (h *RedactingHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.inner.Enabled(ctx, level)
}

// Handle redacts sensitive attribute values and forwards the record to the inner handler.
func (h *RedactingHandler) Handle(ctx context.Context, r slog.Record) error {
	// Build a new set of attributes with redacted values
	var redacted []slog.Attr
	r.Attrs(func(a slog.Attr) bool {
		redacted = append(redacted, redactAttr(a))
		return true
	})

	// Create a clean record with the same metadata
	newRecord := slog.NewRecord(r.Time, r.Level, r.Message, r.PC)
	newRecord.AddAttrs(redacted...)

	return h.inner.Handle(ctx, newRecord)
}

// WithAttrs returns a new handler with the given attributes redacted.
func (h *RedactingHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	redacted := make([]slog.Attr, len(attrs))
	for i, a := range attrs {
		redacted[i] = redactAttr(a)
	}
	return &RedactingHandler{inner: h.inner.WithAttrs(redacted)}
}

// WithGroup returns a new handler with the given group name.
func (h *RedactingHandler) WithGroup(name string) slog.Handler {
	return &RedactingHandler{inner: h.inner.WithGroup(name)}
}

// redactAttr returns a copy of the attribute with its value redacted if necessary.
func redactAttr(a slog.Attr) slog.Attr {
	key := strings.ToLower(a.Key)

	// Check if the key name itself indicates sensitive data.
	// Special-case: skip keys that are clearly not secrets even though they
	// contain the word "token" (e.g. "token_amount", "token_balance").
	for _, pattern := range sensitiveKeyPatterns {
		if strings.Contains(key, pattern) {
			return slog.String(a.Key, "[REDACTED]")
		}
	}

	// Handle "token" key specifically: redact only if the value looks like a
	// secret (starts with mb_, wt_, 0x hex, or is a long hex string), not when
	// it could be a numeric token amount.
	if strings.Contains(key, "token") && !strings.Contains(key, "token_amount") &&
		!strings.Contains(key, "token_balance") && !strings.Contains(key, "token_count") &&
		!strings.Contains(key, "token_price") {
		val := a.Value.String()
		if strings.HasPrefix(val, "mb_") || strings.HasPrefix(val, "wt_") ||
			ethPrivateKeyPattern.MatchString(val) || longHexPattern.MatchString(val) {
			return slog.String(a.Key, "[REDACTED]")
		}
	}

	// For string values, scan and redact known secret patterns inline.
	if a.Value.Kind() == slog.KindString {
		val := a.Value.String()
		redacted := redactString(val)
		if redacted != val {
			return slog.String(a.Key, redacted)
		}
	}

	return a
}

// redactString scans a string value and replaces known secret patterns.
func redactString(val string) string {
	// Redact API keys: show prefix + first 8 chars, mask the rest
	val = apiKeyPattern.ReplaceAllStringFunc(val, func(match string) string {
		if len(match) <= 16 {
			// Very short key: show prefix only
			parts := strings.SplitN(match, "_", 3)
			if len(parts) >= 3 {
				prefix := parts[0] + "_" + parts[1] + "_"
				rest := parts[2]
				if len(rest) > 8 {
					return prefix + rest[:8] + "..."
				}
			}
			return match
		}
		// Show prefix + first 8 chars of the key portion
		parts := strings.SplitN(match, "_", 3)
		if len(parts) >= 3 {
			prefix := parts[0] + "_" + parts[1] + "_"
			rest := parts[2]
			if len(rest) > 8 {
				return prefix + rest[:8] + "..."
			}
			return match
		}
		return match
	})

	// Redact Ethereum private keys (0x + 64 hex chars)
	val = ethPrivateKeyPattern.ReplaceAllStringFunc(val, func(match string) string {
		return match[:6] + "..." + match[len(match)-4:]
	})

	// Redact long hex strings that look like private keys (> 64 chars)
	val = longHexPattern.ReplaceAllStringFunc(val, func(match string) string {
		return match[:8] + "...[REDACTED]"
	})

	return val
}

// EnableRedaction wraps the current global logger with a RedactingHandler.
// After calling this, all log output through the global logging functions
// will have sensitive values automatically redacted.
func EnableRedaction() {
	mu.Lock()
	defer mu.Unlock()

	handler := defaultLogger.Handler()

	// Avoid double-wrapping if already a RedactingHandler
	if _, ok := handler.(*RedactingHandler); ok {
		return
	}

	redacting := NewRedactingHandler(handler)
	defaultLogger = slog.New(redacting)
}

// NewRedactingLogger creates a new slog.Logger with redaction enabled.
func NewRedactingLogger(inner slog.Handler) *slog.Logger {
	return slog.New(NewRedactingHandler(inner))
}

// redactValue is a helper that redacts a value by its formatted string representation.
// It is used for non-string slog values.
func redactValue(v slog.Value) slog.Value {
	str := fmt.Sprintf("%v", v.Any())
	redacted := redactString(str)
	if redacted != str {
		return slog.StringValue(redacted)
	}
	return v
}
