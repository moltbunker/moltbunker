package util

import (
	"context"
	"errors"
	"math"
	"math/rand"
	"time"
)

// RetryConfig holds configuration for retry with exponential backoff
type RetryConfig struct {
	// MaxRetries is the maximum number of retry attempts (0 = no retries, -1 = unlimited)
	MaxRetries int
	// BaseDelay is the initial delay between retries
	BaseDelay time.Duration
	// MaxDelay is the maximum delay between retries
	MaxDelay time.Duration
	// Multiplier is the factor by which delay increases (default: 2.0)
	Multiplier float64
	// Jitter adds randomness to delays to prevent thundering herd (0.0 - 1.0)
	Jitter float64
	// RetryIf is an optional function to determine if an error is retryable
	RetryIf func(error) bool
}

// DefaultRetryConfig returns sensible defaults for retry configuration
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries: 3,
		BaseDelay:  100 * time.Millisecond,
		MaxDelay:   30 * time.Second,
		Multiplier: 2.0,
		Jitter:     0.1,
		RetryIf:    nil, // Retry all errors by default
	}
}

// RetryResult contains the result of a retry operation
type RetryResult struct {
	Attempts  int           // Number of attempts made
	LastError error         // Last error encountered
	Duration  time.Duration // Total duration of all attempts
}

// ErrMaxRetriesExceeded is returned when max retries is exceeded
var ErrMaxRetriesExceeded = errors.New("maximum retries exceeded")

// ErrContextCanceled is returned when context is canceled during retry
var ErrContextCanceled = errors.New("context canceled during retry")

// Retry executes a function with exponential backoff retry
func Retry(ctx context.Context, config *RetryConfig, fn func() error) *RetryResult {
	if config == nil {
		config = DefaultRetryConfig()
	}

	result := &RetryResult{}
	start := time.Now()

	for {
		result.Attempts++

		// Execute the function
		err := fn()
		if err == nil {
			result.LastError = nil // Clear any previous error on success
			result.Duration = time.Since(start)
			return result
		}

		result.LastError = err

		// Check if error is retryable
		if config.RetryIf != nil && !config.RetryIf(err) {
			result.Duration = time.Since(start)
			return result
		}

		// Check if we've exceeded max retries
		if config.MaxRetries >= 0 && result.Attempts > config.MaxRetries {
			result.LastError = errors.Join(ErrMaxRetriesExceeded, err)
			result.Duration = time.Since(start)
			return result
		}

		// Calculate delay with exponential backoff
		delay := calculateDelay(config, result.Attempts)

		// Wait with context cancellation support
		select {
		case <-ctx.Done():
			result.LastError = errors.Join(ErrContextCanceled, ctx.Err())
			result.Duration = time.Since(start)
			return result
		case <-time.After(delay):
			// Continue to next attempt
		}
	}
}

// RetryWithValue executes a function that returns a value with exponential backoff retry
func RetryWithValue[T any](ctx context.Context, config *RetryConfig, fn func() (T, error)) (T, *RetryResult) {
	if config == nil {
		config = DefaultRetryConfig()
	}

	var result T
	retryResult := &RetryResult{}
	start := time.Now()

	for {
		retryResult.Attempts++

		// Execute the function
		val, err := fn()
		if err == nil {
			retryResult.LastError = nil // Clear any previous error on success
			retryResult.Duration = time.Since(start)
			return val, retryResult
		}

		retryResult.LastError = err

		// Check if error is retryable
		if config.RetryIf != nil && !config.RetryIf(err) {
			retryResult.Duration = time.Since(start)
			return result, retryResult
		}

		// Check if we've exceeded max retries
		if config.MaxRetries >= 0 && retryResult.Attempts > config.MaxRetries {
			retryResult.LastError = errors.Join(ErrMaxRetriesExceeded, err)
			retryResult.Duration = time.Since(start)
			return result, retryResult
		}

		// Calculate delay with exponential backoff
		delay := calculateDelay(config, retryResult.Attempts)

		// Wait with context cancellation support
		select {
		case <-ctx.Done():
			retryResult.LastError = errors.Join(ErrContextCanceled, ctx.Err())
			retryResult.Duration = time.Since(start)
			return result, retryResult
		case <-time.After(delay):
			// Continue to next attempt
		}
	}
}

// calculateDelay calculates the delay for a given attempt number
func calculateDelay(config *RetryConfig, attempt int) time.Duration {
	// Calculate base delay with exponential backoff
	multiplier := config.Multiplier
	if multiplier <= 0 {
		multiplier = 2.0
	}

	// delay = baseDelay * multiplier^(attempt-1)
	delay := float64(config.BaseDelay) * math.Pow(multiplier, float64(attempt-1))

	// Apply jitter
	if config.Jitter > 0 {
		jitterRange := delay * config.Jitter
		delay = delay - jitterRange + (rand.Float64() * 2 * jitterRange)
	}

	// Clamp to max delay
	if config.MaxDelay > 0 && time.Duration(delay) > config.MaxDelay {
		delay = float64(config.MaxDelay)
	}

	return time.Duration(delay)
}

// RetryableError wraps an error and marks it as retryable
type RetryableError struct {
	Err error
}

func (e *RetryableError) Error() string {
	return e.Err.Error()
}

func (e *RetryableError) Unwrap() error {
	return e.Err
}

// IsRetryable checks if an error is marked as retryable
func IsRetryable(err error) bool {
	var retryable *RetryableError
	return errors.As(err, &retryable)
}

// MarkRetryable marks an error as retryable
func MarkRetryable(err error) error {
	if err == nil {
		return nil
	}
	return &RetryableError{Err: err}
}

// NonRetryableError wraps an error and marks it as non-retryable
type NonRetryableError struct {
	Err error
}

func (e *NonRetryableError) Error() string {
	return e.Err.Error()
}

func (e *NonRetryableError) Unwrap() error {
	return e.Err
}

// IsNonRetryable checks if an error is marked as non-retryable
func IsNonRetryable(err error) bool {
	var nonRetryable *NonRetryableError
	return errors.As(err, &nonRetryable)
}

// MarkNonRetryable marks an error as non-retryable
func MarkNonRetryable(err error) error {
	if err == nil {
		return nil
	}
	return &NonRetryableError{Err: err}
}

// DefaultRetryIf returns a function that retries all errors except non-retryable ones
func DefaultRetryIf() func(error) bool {
	return func(err error) bool {
		return !IsNonRetryable(err)
	}
}

// TemporaryError interface for errors that are temporary
type TemporaryError interface {
	Temporary() bool
}

// RetryIfTemporary returns a function that only retries temporary errors
func RetryIfTemporary() func(error) bool {
	return func(err error) bool {
		var temp TemporaryError
		if errors.As(err, &temp) {
			return temp.Temporary()
		}
		return false
	}
}
