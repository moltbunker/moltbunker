package util

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRetry_SuccessOnFirstAttempt(t *testing.T) {
	ctx := context.Background()
	attempts := 0

	result := Retry(ctx, nil, func() error {
		attempts++
		return nil
	})

	if result.Attempts != 1 {
		t.Errorf("expected 1 attempt, got %d", result.Attempts)
	}
	if attempts != 1 {
		t.Errorf("expected function to be called once, got %d", attempts)
	}
	if result.LastError != nil {
		t.Errorf("expected no error, got %v", result.LastError)
	}
}

func TestRetry_SuccessAfterRetries(t *testing.T) {
	ctx := context.Background()
	attempts := 0
	failUntil := 3

	config := &RetryConfig{
		MaxRetries: 5,
		BaseDelay:  1 * time.Millisecond,
		MaxDelay:   10 * time.Millisecond,
		Multiplier: 2.0,
	}

	result := Retry(ctx, config, func() error {
		attempts++
		if attempts < failUntil {
			return errors.New("temporary error")
		}
		return nil
	})

	if result.Attempts != failUntil {
		t.Errorf("expected %d attempts, got %d", failUntil, result.Attempts)
	}
	if result.LastError != nil {
		t.Errorf("expected no error, got %v", result.LastError)
	}
}

func TestRetry_MaxRetriesExceeded(t *testing.T) {
	ctx := context.Background()
	attempts := 0
	testErr := errors.New("persistent error")

	config := &RetryConfig{
		MaxRetries: 3,
		BaseDelay:  1 * time.Millisecond,
	}

	result := Retry(ctx, config, func() error {
		attempts++
		return testErr
	})

	// MaxRetries=3 means 1 initial + 3 retries = 4 total attempts
	if result.Attempts != 4 {
		t.Errorf("expected 4 attempts, got %d", result.Attempts)
	}

	if !errors.Is(result.LastError, ErrMaxRetriesExceeded) {
		t.Error("expected ErrMaxRetriesExceeded in error chain")
	}

	if !errors.Is(result.LastError, testErr) {
		t.Error("expected original error in error chain")
	}
}

func TestRetry_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	attempts := 0

	config := &RetryConfig{
		MaxRetries: 100, // High number to ensure we don't hit max
		BaseDelay:  100 * time.Millisecond,
	}

	// Cancel context after a short delay
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	result := Retry(ctx, config, func() error {
		attempts++
		return errors.New("error")
	})

	if !errors.Is(result.LastError, ErrContextCanceled) {
		t.Error("expected ErrContextCanceled in error chain")
	}
}

func TestRetry_NonRetryableError(t *testing.T) {
	ctx := context.Background()
	attempts := 0
	testErr := MarkNonRetryable(errors.New("non-retryable"))

	config := &RetryConfig{
		MaxRetries: 5,
		BaseDelay:  1 * time.Millisecond,
		RetryIf:    DefaultRetryIf(),
	}

	_ = Retry(ctx, config, func() error {
		attempts++
		return testErr
	})

	if attempts != 1 {
		t.Errorf("expected 1 attempt for non-retryable error, got %d", attempts)
	}
}

func TestRetry_RetryIfFunction(t *testing.T) {
	ctx := context.Background()
	attempts := 0
	retryableErr := errors.New("retryable")
	nonRetryableErr := errors.New("non-retryable")

	config := &RetryConfig{
		MaxRetries: 5,
		BaseDelay:  1 * time.Millisecond,
		RetryIf: func(err error) bool {
			return err.Error() == "retryable"
		},
	}

	// First test: retryable error
	attempts = 0
	Retry(ctx, config, func() error {
		attempts++
		if attempts >= 3 {
			return nil
		}
		return retryableErr
	})

	if attempts != 3 {
		t.Errorf("expected 3 attempts for retryable error, got %d", attempts)
	}

	// Second test: non-retryable error
	attempts = 0
	Retry(ctx, config, func() error {
		attempts++
		return nonRetryableErr
	})

	if attempts != 1 {
		t.Errorf("expected 1 attempt for non-retryable error, got %d", attempts)
	}
}

func TestRetryWithValue_Success(t *testing.T) {
	ctx := context.Background()

	config := &RetryConfig{
		MaxRetries: 3,
		BaseDelay:  1 * time.Millisecond,
	}

	attempts := 0
	value, result := RetryWithValue(ctx, config, func() (string, error) {
		attempts++
		if attempts < 2 {
			return "", errors.New("not yet")
		}
		return "success", nil
	})

	if value != "success" {
		t.Errorf("expected 'success', got %q", value)
	}
	if result.Attempts != 2 {
		t.Errorf("expected 2 attempts, got %d", result.Attempts)
	}
}

func TestRetryWithValue_Failure(t *testing.T) {
	ctx := context.Background()

	config := &RetryConfig{
		MaxRetries: 2,
		BaseDelay:  1 * time.Millisecond,
	}

	value, result := RetryWithValue(ctx, config, func() (int, error) {
		return 0, errors.New("always fails")
	})

	if value != 0 {
		t.Errorf("expected zero value, got %d", value)
	}
	if !errors.Is(result.LastError, ErrMaxRetriesExceeded) {
		t.Error("expected ErrMaxRetriesExceeded")
	}
}

func TestCalculateDelay(t *testing.T) {
	config := &RetryConfig{
		BaseDelay:  100 * time.Millisecond,
		MaxDelay:   10 * time.Second,
		Multiplier: 2.0,
		Jitter:     0, // No jitter for predictable testing
	}

	tests := []struct {
		attempt  int
		expected time.Duration
	}{
		{1, 100 * time.Millisecond},  // 100 * 2^0
		{2, 200 * time.Millisecond},  // 100 * 2^1
		{3, 400 * time.Millisecond},  // 100 * 2^2
		{4, 800 * time.Millisecond},  // 100 * 2^3
		{5, 1600 * time.Millisecond}, // 100 * 2^4
	}

	for _, tt := range tests {
		delay := calculateDelay(config, tt.attempt)
		if delay != tt.expected {
			t.Errorf("attempt %d: expected %v, got %v", tt.attempt, tt.expected, delay)
		}
	}
}

func TestCalculateDelay_MaxDelay(t *testing.T) {
	config := &RetryConfig{
		BaseDelay:  100 * time.Millisecond,
		MaxDelay:   500 * time.Millisecond,
		Multiplier: 2.0,
		Jitter:     0,
	}

	// Attempt 4 would be 800ms, but should be capped at 500ms
	delay := calculateDelay(config, 4)
	if delay != 500*time.Millisecond {
		t.Errorf("expected max delay of 500ms, got %v", delay)
	}
}

func TestCalculateDelay_WithJitter(t *testing.T) {
	config := &RetryConfig{
		BaseDelay:  100 * time.Millisecond,
		MaxDelay:   10 * time.Second,
		Multiplier: 2.0,
		Jitter:     0.5, // 50% jitter
	}

	// With 50% jitter, delay should be between 50-150ms for first attempt
	for i := 0; i < 100; i++ {
		delay := calculateDelay(config, 1)
		if delay < 50*time.Millisecond || delay > 150*time.Millisecond {
			t.Errorf("delay %v is outside expected jitter range", delay)
		}
	}
}

func TestMarkRetryable(t *testing.T) {
	err := errors.New("test error")
	retryable := MarkRetryable(err)

	if !IsRetryable(retryable) {
		t.Error("expected error to be retryable")
	}

	if !errors.Is(retryable, err) {
		t.Error("expected original error in chain")
	}
}

func TestMarkNonRetryable(t *testing.T) {
	err := errors.New("test error")
	nonRetryable := MarkNonRetryable(err)

	if !IsNonRetryable(nonRetryable) {
		t.Error("expected error to be non-retryable")
	}

	if !errors.Is(nonRetryable, err) {
		t.Error("expected original error in chain")
	}
}

func TestDefaultRetryConfig(t *testing.T) {
	config := DefaultRetryConfig()

	if config.MaxRetries != 3 {
		t.Errorf("expected MaxRetries=3, got %d", config.MaxRetries)
	}
	if config.BaseDelay != 100*time.Millisecond {
		t.Errorf("expected BaseDelay=100ms, got %v", config.BaseDelay)
	}
	if config.MaxDelay != 30*time.Second {
		t.Errorf("expected MaxDelay=30s, got %v", config.MaxDelay)
	}
	if config.Multiplier != 2.0 {
		t.Errorf("expected Multiplier=2.0, got %f", config.Multiplier)
	}
}
