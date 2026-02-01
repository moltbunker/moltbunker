package logging

import (
	"context"
	"io"
	"log/slog"
	"os"
	"sync"
)

var (
	defaultLogger *slog.Logger
	mu            sync.RWMutex
)

func init() {
	// Initialize with default JSON handler for production
	defaultLogger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}

// SetLogger sets the global logger
func SetLogger(logger *slog.Logger) {
	mu.Lock()
	defer mu.Unlock()
	defaultLogger = logger
}

// SetOutput sets the output destination for the logger
func SetOutput(w io.Writer) {
	mu.Lock()
	defer mu.Unlock()
	defaultLogger = slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}

// SetLevel sets the logging level
func SetLevel(level slog.Level) {
	mu.Lock()
	defer mu.Unlock()
	defaultLogger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))
}

// SetTextOutput sets up human-readable text output (for development)
func SetTextOutput(w io.Writer) {
	mu.Lock()
	defer mu.Unlock()
	defaultLogger = slog.New(slog.NewTextHandler(w, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
}

// Logger returns the default logger
func Logger() *slog.Logger {
	mu.RLock()
	defer mu.RUnlock()
	return defaultLogger
}

// With returns a logger with additional context
func With(args ...any) *slog.Logger {
	return Logger().With(args...)
}

// Debug logs at debug level
func Debug(msg string, args ...any) {
	Logger().Debug(msg, args...)
}

// Info logs at info level
func Info(msg string, args ...any) {
	Logger().Info(msg, args...)
}

// Warn logs at warn level
func Warn(msg string, args ...any) {
	Logger().Warn(msg, args...)
}

// Error logs at error level
func Error(msg string, args ...any) {
	Logger().Error(msg, args...)
}

// DebugContext logs at debug level with context
func DebugContext(ctx context.Context, msg string, args ...any) {
	Logger().DebugContext(ctx, msg, args...)
}

// InfoContext logs at info level with context
func InfoContext(ctx context.Context, msg string, args ...any) {
	Logger().InfoContext(ctx, msg, args...)
}

// WarnContext logs at warn level with context
func WarnContext(ctx context.Context, msg string, args ...any) {
	Logger().WarnContext(ctx, msg, args...)
}

// ErrorContext logs at error level with context
func ErrorContext(ctx context.Context, msg string, args ...any) {
	Logger().ErrorContext(ctx, msg, args...)
}

// Common field helpers
func ContainerID(id string) slog.Attr {
	return slog.String("container_id", id)
}

func NodeID(id string) slog.Attr {
	return slog.String("node_id", id)
}

func Region(region string) slog.Attr {
	return slog.String("region", region)
}

func Err(err error) slog.Attr {
	if err == nil {
		return slog.String("error", "")
	}
	return slog.String("error", err.Error())
}

func Component(name string) slog.Attr {
	return slog.String("component", name)
}
