package runtime

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// HealthChecker unit tests
// ---------------------------------------------------------------------------

func TestNewHealthChecker(t *testing.T) {
	hc := NewHealthChecker()
	if hc == nil {
		t.Fatal("expected non-nil HealthChecker")
	}
	if hc.probes == nil {
		t.Error("probes map should be initialized")
	}
	if hc.httpClient == nil {
		t.Error("httpClient should be initialized")
	}
	if hc.running {
		t.Error("should not be running initially")
	}
	if hc.ProbeCount() != 0 {
		t.Errorf("expected 0 probes, got %d", hc.ProbeCount())
	}
}

func TestRegisterUnregisterProbe(t *testing.T) {
	hc := NewHealthChecker()

	t.Run("register adds probe", func(t *testing.T) {
		hc.RegisterProbe("container-1", HealthProbeConfig{
			Type:    ProbeTCP,
			TCPPort: 8080,
		})
		if hc.ProbeCount() != 1 {
			t.Errorf("expected 1 probe, got %d", hc.ProbeCount())
		}
	})

	t.Run("register replaces existing probe", func(t *testing.T) {
		hc.RegisterProbe("container-1", HealthProbeConfig{
			Type:    ProbeHTTP,
			HTTPPort: 9090,
		})
		if hc.ProbeCount() != 1 {
			t.Errorf("expected 1 probe after replace, got %d", hc.ProbeCount())
		}
		status := hc.GetProbeStatus("container-1")
		if status == nil {
			t.Fatal("expected non-nil status after re-register")
		}
		if status.State != ProbeStateUnknown {
			t.Errorf("expected unknown state after re-register, got %s", status.State)
		}
	})

	t.Run("unregister removes probe", func(t *testing.T) {
		hc.UnregisterProbe("container-1")
		if hc.ProbeCount() != 0 {
			t.Errorf("expected 0 probes after unregister, got %d", hc.ProbeCount())
		}
	})

	t.Run("unregister nonexistent is safe", func(t *testing.T) {
		hc.UnregisterProbe("nonexistent")
	})
}

func TestGetProbeStatus(t *testing.T) {
	hc := NewHealthChecker()

	t.Run("returns nil for unregistered container", func(t *testing.T) {
		status := hc.GetProbeStatus("nonexistent")
		if status != nil {
			t.Error("expected nil status for unregistered container")
		}
	})

	t.Run("returns status for registered container", func(t *testing.T) {
		hc.RegisterProbe("test-container", HealthProbeConfig{
			Type:    ProbeTCP,
			TCPPort: 8080,
		})
		status := hc.GetProbeStatus("test-container")
		if status == nil {
			t.Fatal("expected non-nil status")
		}
		if status.ContainerID != "test-container" {
			t.Errorf("expected container ID 'test-container', got %q", status.ContainerID)
		}
		if status.State != ProbeStateUnknown {
			t.Errorf("expected unknown state, got %s", status.State)
		}
	})
}

func TestCheckHealth_NoProbe(t *testing.T) {
	hc := NewHealthChecker()
	_, err := hc.CheckHealth(context.Background(), "nonexistent")
	if err == nil {
		t.Error("expected error for unregistered container")
	}
	if !strings.Contains(err.Error(), "no probe registered") {
		t.Errorf("unexpected error message: %s", err.Error())
	}
}

// ---------------------------------------------------------------------------
// HTTP probe tests
// ---------------------------------------------------------------------------

func TestHTTPProbe_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/healthz" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("ok"))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	// Parse the server address
	host, port := parseHostPort(t, server.Listener.Addr().String())

	hc := NewHealthChecker()
	hc.SetAddressFunc(func(containerID string) (string, error) {
		return host, nil
	})

	hc.RegisterProbe("http-ok", HealthProbeConfig{
		Type:             ProbeHTTP,
		HTTPPath:         "/healthz",
		HTTPPort:         port,
		HTTPExpectedCodes: []int{200},
		Timeout:          5 * time.Second,
		SuccessThreshold: 1,
		FailureThreshold: 3,
	})

	ok, err := hc.CheckHealth(context.Background(), "http-ok")
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if !ok {
		t.Error("expected probe to succeed")
	}

	status := hc.GetProbeStatus("http-ok")
	if status.State != ProbeStateHealthy {
		t.Errorf("expected healthy state, got %s", status.State)
	}
	if status.TotalChecks != 1 {
		t.Errorf("expected 1 total check, got %d", status.TotalChecks)
	}
	if status.TotalSuccesses != 1 {
		t.Errorf("expected 1 success, got %d", status.TotalSuccesses)
	}
}

func TestHTTPProbe_WrongStatusCode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	host, port := parseHostPort(t, server.Listener.Addr().String())

	hc := NewHealthChecker()
	hc.SetAddressFunc(func(containerID string) (string, error) {
		return host, nil
	})

	hc.RegisterProbe("http-503", HealthProbeConfig{
		Type:             ProbeHTTP,
		HTTPPath:         "/",
		HTTPPort:         port,
		HTTPExpectedCodes: []int{200},
		Timeout:          5 * time.Second,
		SuccessThreshold: 1,
		FailureThreshold: 1,
	})

	ok, err := hc.CheckHealth(context.Background(), "http-503")
	if ok {
		t.Error("expected probe to fail")
	}
	if err == nil {
		t.Error("expected error for wrong status code")
	}

	status := hc.GetProbeStatus("http-503")
	if status.State != ProbeStateUnhealthy {
		t.Errorf("expected unhealthy state with threshold 1, got %s", status.State)
	}
}

func TestHTTPProbe_MultipleExpectedCodes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent) // 204
	}))
	defer server.Close()

	host, port := parseHostPort(t, server.Listener.Addr().String())

	hc := NewHealthChecker()
	hc.SetAddressFunc(func(containerID string) (string, error) {
		return host, nil
	})

	hc.RegisterProbe("http-204", HealthProbeConfig{
		Type:             ProbeHTTP,
		HTTPPath:         "/",
		HTTPPort:         port,
		HTTPExpectedCodes: []int{200, 204},
		Timeout:          5 * time.Second,
		SuccessThreshold: 1,
		FailureThreshold: 3,
	})

	ok, err := hc.CheckHealth(context.Background(), "http-204")
	if err != nil {
		t.Fatalf("expected success for 204 with expected codes [200, 204], got: %v", err)
	}
	if !ok {
		t.Error("expected probe to succeed")
	}
}

func TestHTTPProbe_Timeout(t *testing.T) {
	// Server that never responds
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
			return
		case <-time.After(30 * time.Second):
			return
		}
	}))
	defer server.Close()

	host, port := parseHostPort(t, server.Listener.Addr().String())

	hc := NewHealthChecker()
	hc.SetAddressFunc(func(containerID string) (string, error) {
		return host, nil
	})

	hc.RegisterProbe("http-timeout", HealthProbeConfig{
		Type:             ProbeHTTP,
		HTTPPath:         "/",
		HTTPPort:         port,
		Timeout:          100 * time.Millisecond, // Very short timeout
		SuccessThreshold: 1,
		FailureThreshold: 1,
	})

	ok, err := hc.CheckHealth(context.Background(), "http-timeout")
	if ok {
		t.Error("expected probe to fail on timeout")
	}
	if err == nil {
		t.Error("expected error on timeout")
	}
}

func TestHTTPProbe_ConnectionRefused(t *testing.T) {
	hc := NewHealthChecker()
	hc.SetAddressFunc(func(containerID string) (string, error) {
		return "127.0.0.1", nil
	})

	// Use a port that's very unlikely to be in use
	hc.RegisterProbe("http-refused", HealthProbeConfig{
		Type:             ProbeHTTP,
		HTTPPath:         "/",
		HTTPPort:         19999,
		Timeout:          1 * time.Second,
		SuccessThreshold: 1,
		FailureThreshold: 1,
	})

	ok, err := hc.CheckHealth(context.Background(), "http-refused")
	if ok {
		t.Error("expected probe to fail when connection refused")
	}
	if err == nil {
		t.Error("expected error when connection refused")
	}
}

// ---------------------------------------------------------------------------
// TCP probe tests
// ---------------------------------------------------------------------------

func TestTCPProbe_Success(t *testing.T) {
	// Start a TCP listener
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start TCP listener: %v", err)
	}
	defer listener.Close()

	// Accept connections in background
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			conn.Close()
		}
	}()

	host, port := parseHostPort(t, listener.Addr().String())

	hc := NewHealthChecker()
	hc.SetAddressFunc(func(containerID string) (string, error) {
		return host, nil
	})

	hc.RegisterProbe("tcp-ok", HealthProbeConfig{
		Type:             ProbeTCP,
		TCPPort:          port,
		Timeout:          5 * time.Second,
		SuccessThreshold: 1,
		FailureThreshold: 3,
	})

	ok, err := hc.CheckHealth(context.Background(), "tcp-ok")
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if !ok {
		t.Error("expected probe to succeed")
	}

	status := hc.GetProbeStatus("tcp-ok")
	if status.State != ProbeStateHealthy {
		t.Errorf("expected healthy state, got %s", status.State)
	}
}

func TestTCPProbe_ConnectionRefused(t *testing.T) {
	hc := NewHealthChecker()
	hc.SetAddressFunc(func(containerID string) (string, error) {
		return "127.0.0.1", nil
	})

	hc.RegisterProbe("tcp-refused", HealthProbeConfig{
		Type:             ProbeTCP,
		TCPPort:          19998, // Unlikely to be in use
		Timeout:          1 * time.Second,
		SuccessThreshold: 1,
		FailureThreshold: 1,
	})

	ok, err := hc.CheckHealth(context.Background(), "tcp-refused")
	if ok {
		t.Error("expected probe to fail when connection refused")
	}
	if err == nil {
		t.Error("expected error when connection refused")
	}

	status := hc.GetProbeStatus("tcp-refused")
	if status.State != ProbeStateUnhealthy {
		t.Errorf("expected unhealthy state, got %s", status.State)
	}
}

func TestTCPProbe_Timeout(t *testing.T) {
	// Create a listener that accepts but never closes (simulates hanging)
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start TCP listener: %v", err)
	}
	// Close the listener immediately so the port is known but not accepting
	listener.Close()

	host, port := parseHostPort(t, listener.Addr().String())

	hc := NewHealthChecker()
	hc.SetAddressFunc(func(containerID string) (string, error) {
		return host, nil
	})

	hc.RegisterProbe("tcp-timeout", HealthProbeConfig{
		Type:             ProbeTCP,
		TCPPort:          port,
		Timeout:          100 * time.Millisecond,
		SuccessThreshold: 1,
		FailureThreshold: 1,
	})

	ok, err := hc.CheckHealth(context.Background(), "tcp-timeout")
	if ok {
		t.Error("expected probe to fail on closed port")
	}
	if err == nil {
		t.Error("expected error for closed port")
	}
}

// ---------------------------------------------------------------------------
// Exec probe tests
// ---------------------------------------------------------------------------

func TestExecProbe_NoExecFunc(t *testing.T) {
	hc := NewHealthChecker()
	// Do not set an exec function -- mock mode

	hc.RegisterProbe("exec-noop", HealthProbeConfig{
		Type:             ProbeExec,
		ExecCommand:      []string{"true"},
		Timeout:          5 * time.Second,
		SuccessThreshold: 1,
		FailureThreshold: 3,
	})

	ok, err := hc.CheckHealth(context.Background(), "exec-noop")
	if err != nil {
		t.Fatalf("expected no-op success, got error: %v", err)
	}
	if !ok {
		t.Error("expected probe to succeed as no-op")
	}
}

func TestExecProbe_Success(t *testing.T) {
	hc := NewHealthChecker()
	hc.SetExecFunc(func(ctx context.Context, containerID string, cmd []string) (int, error) {
		return 0, nil
	})

	hc.RegisterProbe("exec-ok", HealthProbeConfig{
		Type:             ProbeExec,
		ExecCommand:      []string{"/bin/check-health"},
		Timeout:          5 * time.Second,
		SuccessThreshold: 1,
		FailureThreshold: 3,
	})

	ok, err := hc.CheckHealth(context.Background(), "exec-ok")
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if !ok {
		t.Error("expected probe to succeed")
	}

	status := hc.GetProbeStatus("exec-ok")
	if status.State != ProbeStateHealthy {
		t.Errorf("expected healthy state, got %s", status.State)
	}
}

func TestExecProbe_NonZeroExit(t *testing.T) {
	hc := NewHealthChecker()
	hc.SetExecFunc(func(ctx context.Context, containerID string, cmd []string) (int, error) {
		return 1, nil
	})

	hc.RegisterProbe("exec-fail", HealthProbeConfig{
		Type:             ProbeExec,
		ExecCommand:      []string{"/bin/check-health"},
		Timeout:          5 * time.Second,
		SuccessThreshold: 1,
		FailureThreshold: 1,
	})

	ok, err := hc.CheckHealth(context.Background(), "exec-fail")
	if ok {
		t.Error("expected probe to fail on non-zero exit")
	}
	if err == nil {
		t.Error("expected error for non-zero exit code")
	}
	if !strings.Contains(err.Error(), "exit code 1") {
		t.Errorf("unexpected error message: %s", err.Error())
	}
}

func TestExecProbe_ExecError(t *testing.T) {
	hc := NewHealthChecker()
	hc.SetExecFunc(func(ctx context.Context, containerID string, cmd []string) (int, error) {
		return -1, fmt.Errorf("container not running")
	})

	hc.RegisterProbe("exec-err", HealthProbeConfig{
		Type:             ProbeExec,
		ExecCommand:      []string{"/bin/check-health"},
		Timeout:          5 * time.Second,
		SuccessThreshold: 1,
		FailureThreshold: 1,
	})

	ok, err := hc.CheckHealth(context.Background(), "exec-err")
	if ok {
		t.Error("expected probe to fail on exec error")
	}
	if err == nil {
		t.Error("expected error")
	}
}

func TestExecProbe_EmptyCommand(t *testing.T) {
	hc := NewHealthChecker()
	hc.SetExecFunc(func(ctx context.Context, containerID string, cmd []string) (int, error) {
		return 0, nil
	})

	hc.RegisterProbe("exec-empty", HealthProbeConfig{
		Type:             ProbeExec,
		ExecCommand:      []string{},
		Timeout:          5 * time.Second,
		SuccessThreshold: 1,
		FailureThreshold: 1,
	})

	ok, err := hc.CheckHealth(context.Background(), "exec-empty")
	if ok {
		t.Error("expected probe to fail with empty command")
	}
	if err == nil {
		t.Error("expected error for empty command")
	}
	if !strings.Contains(err.Error(), "no command configured") {
		t.Errorf("unexpected error message: %s", err.Error())
	}
}

func TestExecProbe_ReceivesCorrectArgs(t *testing.T) {
	var receivedContainerID string
	var receivedCmd []string

	hc := NewHealthChecker()
	hc.SetExecFunc(func(ctx context.Context, containerID string, cmd []string) (int, error) {
		receivedContainerID = containerID
		receivedCmd = cmd
		return 0, nil
	})

	expectedCmd := []string{"/bin/sh", "-c", "curl -f localhost:8080/health"}
	hc.RegisterProbe("exec-args", HealthProbeConfig{
		Type:             ProbeExec,
		ExecCommand:      expectedCmd,
		Timeout:          5 * time.Second,
		SuccessThreshold: 1,
		FailureThreshold: 3,
	})

	hc.CheckHealth(context.Background(), "exec-args")

	if receivedContainerID != "exec-args" {
		t.Errorf("expected container ID 'exec-args', got %q", receivedContainerID)
	}
	if len(receivedCmd) != len(expectedCmd) {
		t.Fatalf("expected %d cmd args, got %d", len(expectedCmd), len(receivedCmd))
	}
	for i, arg := range expectedCmd {
		if receivedCmd[i] != arg {
			t.Errorf("cmd arg[%d] = %q, want %q", i, receivedCmd[i], arg)
		}
	}
}

// ---------------------------------------------------------------------------
// Threshold transition tests
// ---------------------------------------------------------------------------

func TestThreshold_SuccessTransition(t *testing.T) {
	hc := NewHealthChecker()
	hc.SetExecFunc(func(ctx context.Context, containerID string, cmd []string) (int, error) {
		return 0, nil
	})

	hc.RegisterProbe("threshold-success", HealthProbeConfig{
		Type:             ProbeExec,
		ExecCommand:      []string{"true"},
		Timeout:          5 * time.Second,
		SuccessThreshold: 3, // Need 3 consecutive successes
		FailureThreshold: 1,
	})

	ctx := context.Background()

	// First check: success, but threshold not met
	hc.CheckHealth(ctx, "threshold-success")
	status := hc.GetProbeStatus("threshold-success")
	if status.State != ProbeStateUnknown {
		t.Errorf("after 1 success: expected unknown, got %s", status.State)
	}
	if status.ConsecutivePass != 1 {
		t.Errorf("expected 1 consecutive pass, got %d", status.ConsecutivePass)
	}

	// Second check
	hc.CheckHealth(ctx, "threshold-success")
	status = hc.GetProbeStatus("threshold-success")
	if status.State != ProbeStateUnknown {
		t.Errorf("after 2 successes: expected unknown, got %s", status.State)
	}

	// Third check: threshold met
	hc.CheckHealth(ctx, "threshold-success")
	status = hc.GetProbeStatus("threshold-success")
	if status.State != ProbeStateHealthy {
		t.Errorf("after 3 successes: expected healthy, got %s", status.State)
	}
	if status.ConsecutivePass != 3 {
		t.Errorf("expected 3 consecutive passes, got %d", status.ConsecutivePass)
	}
}

func TestThreshold_FailureTransition(t *testing.T) {
	hc := NewHealthChecker()
	hc.SetExecFunc(func(ctx context.Context, containerID string, cmd []string) (int, error) {
		return 1, nil // Always fail
	})

	hc.RegisterProbe("threshold-fail", HealthProbeConfig{
		Type:             ProbeExec,
		ExecCommand:      []string{"false"},
		Timeout:          5 * time.Second,
		SuccessThreshold: 1,
		FailureThreshold: 3, // Need 3 consecutive failures
	})

	ctx := context.Background()

	// First failure
	hc.CheckHealth(ctx, "threshold-fail")
	status := hc.GetProbeStatus("threshold-fail")
	if status.State != ProbeStateUnknown {
		t.Errorf("after 1 failure: expected unknown, got %s", status.State)
	}

	// Second failure
	hc.CheckHealth(ctx, "threshold-fail")
	status = hc.GetProbeStatus("threshold-fail")
	if status.State != ProbeStateUnknown {
		t.Errorf("after 2 failures: expected unknown, got %s", status.State)
	}

	// Third failure: threshold met
	hc.CheckHealth(ctx, "threshold-fail")
	status = hc.GetProbeStatus("threshold-fail")
	if status.State != ProbeStateUnhealthy {
		t.Errorf("after 3 failures: expected unhealthy, got %s", status.State)
	}
	if status.ConsecutiveFail != 3 {
		t.Errorf("expected 3 consecutive failures, got %d", status.ConsecutiveFail)
	}
}

func TestThreshold_HealthyToUnhealthy(t *testing.T) {
	callCount := 0
	hc := NewHealthChecker()
	hc.SetExecFunc(func(ctx context.Context, containerID string, cmd []string) (int, error) {
		callCount++
		if callCount <= 2 {
			return 0, nil // First 2 calls succeed
		}
		return 1, nil // Rest fail
	})

	hc.RegisterProbe("transition", HealthProbeConfig{
		Type:             ProbeExec,
		ExecCommand:      []string{"check"},
		Timeout:          5 * time.Second,
		SuccessThreshold: 2,
		FailureThreshold: 2,
	})

	ctx := context.Background()

	// Two successes -> healthy
	hc.CheckHealth(ctx, "transition")
	hc.CheckHealth(ctx, "transition")
	status := hc.GetProbeStatus("transition")
	if status.State != ProbeStateHealthy {
		t.Fatalf("expected healthy after 2 successes, got %s", status.State)
	}

	// One failure, still healthy
	hc.CheckHealth(ctx, "transition")
	status = hc.GetProbeStatus("transition")
	if status.State != ProbeStateHealthy {
		t.Errorf("expected healthy after 1 failure (threshold=2), got %s", status.State)
	}
	if status.ConsecutiveFail != 1 {
		t.Errorf("expected 1 consecutive fail, got %d", status.ConsecutiveFail)
	}

	// Second failure -> unhealthy
	hc.CheckHealth(ctx, "transition")
	status = hc.GetProbeStatus("transition")
	if status.State != ProbeStateUnhealthy {
		t.Errorf("expected unhealthy after 2 failures, got %s", status.State)
	}
}

func TestThreshold_UnhealthyToHealthy(t *testing.T) {
	callCount := 0
	hc := NewHealthChecker()
	hc.SetExecFunc(func(ctx context.Context, containerID string, cmd []string) (int, error) {
		callCount++
		if callCount <= 2 {
			return 1, nil // First 2 calls fail
		}
		return 0, nil // Rest succeed
	})

	hc.RegisterProbe("recovery", HealthProbeConfig{
		Type:             ProbeExec,
		ExecCommand:      []string{"check"},
		Timeout:          5 * time.Second,
		SuccessThreshold: 2,
		FailureThreshold: 2,
	})

	ctx := context.Background()

	// Two failures -> unhealthy
	hc.CheckHealth(ctx, "recovery")
	hc.CheckHealth(ctx, "recovery")
	status := hc.GetProbeStatus("recovery")
	if status.State != ProbeStateUnhealthy {
		t.Fatalf("expected unhealthy after 2 failures, got %s", status.State)
	}

	// One success, still unhealthy
	hc.CheckHealth(ctx, "recovery")
	status = hc.GetProbeStatus("recovery")
	if status.State != ProbeStateUnhealthy {
		t.Errorf("expected unhealthy after 1 success (threshold=2), got %s", status.State)
	}

	// Second success -> healthy
	hc.CheckHealth(ctx, "recovery")
	status = hc.GetProbeStatus("recovery")
	if status.State != ProbeStateHealthy {
		t.Errorf("expected healthy after 2 successes, got %s", status.State)
	}
}

func TestThreshold_FailureResetsSuccessCounter(t *testing.T) {
	callCount := 0
	hc := NewHealthChecker()
	hc.SetExecFunc(func(ctx context.Context, containerID string, cmd []string) (int, error) {
		callCount++
		if callCount == 2 {
			return 1, nil // Fail on second call
		}
		return 0, nil // Succeed on others
	})

	hc.RegisterProbe("reset-counter", HealthProbeConfig{
		Type:             ProbeExec,
		ExecCommand:      []string{"check"},
		Timeout:          5 * time.Second,
		SuccessThreshold: 3,
		FailureThreshold: 3,
	})

	ctx := context.Background()

	// Success
	hc.CheckHealth(ctx, "reset-counter")
	status := hc.GetProbeStatus("reset-counter")
	if status.ConsecutivePass != 1 {
		t.Errorf("expected 1 consecutive pass, got %d", status.ConsecutivePass)
	}

	// Failure resets success counter
	hc.CheckHealth(ctx, "reset-counter")
	status = hc.GetProbeStatus("reset-counter")
	if status.ConsecutivePass != 0 {
		t.Errorf("expected 0 consecutive passes after failure, got %d", status.ConsecutivePass)
	}
	if status.ConsecutiveFail != 1 {
		t.Errorf("expected 1 consecutive fail, got %d", status.ConsecutiveFail)
	}

	// Success again (counter starts fresh)
	hc.CheckHealth(ctx, "reset-counter")
	status = hc.GetProbeStatus("reset-counter")
	if status.ConsecutivePass != 1 {
		t.Errorf("expected 1 consecutive pass after recovery, got %d", status.ConsecutivePass)
	}
	if status.ConsecutiveFail != 0 {
		t.Errorf("expected 0 consecutive fails after success, got %d", status.ConsecutiveFail)
	}
}

// ---------------------------------------------------------------------------
// Total counters test
// ---------------------------------------------------------------------------

func TestTotalCounters(t *testing.T) {
	callCount := 0
	hc := NewHealthChecker()
	hc.SetExecFunc(func(ctx context.Context, containerID string, cmd []string) (int, error) {
		callCount++
		if callCount%2 == 0 {
			return 1, nil
		}
		return 0, nil
	})

	hc.RegisterProbe("counters", HealthProbeConfig{
		Type:             ProbeExec,
		ExecCommand:      []string{"check"},
		Timeout:          5 * time.Second,
		SuccessThreshold: 1,
		FailureThreshold: 1,
	})

	ctx := context.Background()

	for i := 0; i < 6; i++ {
		hc.CheckHealth(ctx, "counters")
	}

	status := hc.GetProbeStatus("counters")
	if status.TotalChecks != 6 {
		t.Errorf("expected 6 total checks, got %d", status.TotalChecks)
	}
	if status.TotalSuccesses != 3 {
		t.Errorf("expected 3 total successes, got %d", status.TotalSuccesses)
	}
	if status.TotalFailures != 3 {
		t.Errorf("expected 3 total failures, got %d", status.TotalFailures)
	}
}

// ---------------------------------------------------------------------------
// Default values tests
// ---------------------------------------------------------------------------

func TestApplyDefaults(t *testing.T) {
	t.Run("HTTP defaults", func(t *testing.T) {
		config := applyDefaults(HealthProbeConfig{Type: ProbeHTTP})
		if config.HTTPPath != "/" {
			t.Errorf("expected default path '/', got %q", config.HTTPPath)
		}
		if config.HTTPPort != 80 {
			t.Errorf("expected default port 80, got %d", config.HTTPPort)
		}
		if len(config.HTTPExpectedCodes) != 1 || config.HTTPExpectedCodes[0] != 200 {
			t.Errorf("expected default expected codes [200], got %v", config.HTTPExpectedCodes)
		}
		if config.Interval != 10*time.Second {
			t.Errorf("expected default interval 10s, got %v", config.Interval)
		}
		if config.Timeout != 5*time.Second {
			t.Errorf("expected default timeout 5s, got %v", config.Timeout)
		}
		if config.SuccessThreshold != 1 {
			t.Errorf("expected default success threshold 1, got %d", config.SuccessThreshold)
		}
		if config.FailureThreshold != 3 {
			t.Errorf("expected default failure threshold 3, got %d", config.FailureThreshold)
		}
	})

	t.Run("TCP keeps custom values", func(t *testing.T) {
		config := applyDefaults(HealthProbeConfig{
			Type:             ProbeTCP,
			TCPPort:          9090,
			Interval:         30 * time.Second,
			Timeout:          10 * time.Second,
			SuccessThreshold: 5,
			FailureThreshold: 2,
		})
		if config.TCPPort != 9090 {
			t.Errorf("expected custom port 9090, got %d", config.TCPPort)
		}
		if config.Interval != 30*time.Second {
			t.Errorf("expected custom interval 30s, got %v", config.Interval)
		}
		if config.Timeout != 10*time.Second {
			t.Errorf("expected custom timeout 10s, got %v", config.Timeout)
		}
		if config.SuccessThreshold != 5 {
			t.Errorf("expected custom success threshold 5, got %d", config.SuccessThreshold)
		}
		if config.FailureThreshold != 2 {
			t.Errorf("expected custom failure threshold 2, got %d", config.FailureThreshold)
		}
	})
}

// ---------------------------------------------------------------------------
// Start/Stop lifecycle tests
// ---------------------------------------------------------------------------

func TestStartStop(t *testing.T) {
	hc := NewHealthChecker()

	if hc.IsRunning() {
		t.Error("should not be running before Start")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	started := make(chan struct{})
	go func() {
		close(started)
		hc.Start(ctx)
	}()

	<-started
	// Give it a moment to enter the loop
	time.Sleep(50 * time.Millisecond)

	if !hc.IsRunning() {
		t.Error("expected running after Start")
	}

	hc.Stop()

	if hc.IsRunning() {
		t.Error("expected not running after Stop")
	}
}

func TestStartStop_ContextCancel(t *testing.T) {
	hc := NewHealthChecker()

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		hc.Start(ctx)
		close(done)
	}()

	// Give it a moment to start
	time.Sleep(50 * time.Millisecond)

	cancel()

	select {
	case <-done:
		// Expected
	case <-time.After(3 * time.Second):
		t.Fatal("Start did not return after context cancellation")
	}

	if hc.IsRunning() {
		t.Error("expected not running after context cancellation")
	}
}

func TestDoubleStart(t *testing.T) {
	hc := NewHealthChecker()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	started := make(chan struct{})
	go func() {
		close(started)
		hc.Start(ctx)
	}()

	<-started
	time.Sleep(50 * time.Millisecond)

	// Second start should return immediately
	secondDone := make(chan struct{})
	go func() {
		hc.Start(ctx) // Should be a no-op
		close(secondDone)
	}()

	select {
	case <-secondDone:
		// Expected: second Start returned immediately
	case <-time.After(1 * time.Second):
		t.Error("second Start should have returned immediately")
	}

	hc.Stop()
}

func TestStartRunsProbes(t *testing.T) {
	var probeCount atomic.Int32

	hc := NewHealthChecker()
	hc.SetExecFunc(func(ctx context.Context, containerID string, cmd []string) (int, error) {
		probeCount.Add(1)
		return 0, nil
	})

	hc.RegisterProbe("auto-probe", HealthProbeConfig{
		Type:             ProbeExec,
		ExecCommand:      []string{"check"},
		Interval:         100 * time.Millisecond, // Fast interval for testing
		Timeout:          1 * time.Second,
		SuccessThreshold: 1,
		FailureThreshold: 3,
	})

	ctx, cancel := context.WithCancel(context.Background())

	go hc.Start(ctx)

	// Wait for a few probe cycles
	time.Sleep(500 * time.Millisecond)
	cancel()

	// Wait for start to finish
	time.Sleep(100 * time.Millisecond)

	count := probeCount.Load()
	if count < 1 {
		t.Errorf("expected at least 1 probe execution, got %d", count)
	}
}

func TestStartRespectsInitialDelay(t *testing.T) {
	var probeCount atomic.Int32

	hc := NewHealthChecker()
	hc.SetExecFunc(func(ctx context.Context, containerID string, cmd []string) (int, error) {
		probeCount.Add(1)
		return 0, nil
	})

	hc.RegisterProbe("delayed-probe", HealthProbeConfig{
		Type:             ProbeExec,
		ExecCommand:      []string{"check"},
		InitialDelay:     5 * time.Second, // Long delay
		Interval:         100 * time.Millisecond,
		Timeout:          1 * time.Second,
		SuccessThreshold: 1,
		FailureThreshold: 3,
	})

	ctx, cancel := context.WithCancel(context.Background())

	go hc.Start(ctx)

	// Wait a bit but not enough for the initial delay
	time.Sleep(300 * time.Millisecond)
	cancel()

	// Wait for start to finish
	time.Sleep(100 * time.Millisecond)

	count := probeCount.Load()
	if count != 0 {
		t.Errorf("expected 0 probes during initial delay, got %d", count)
	}
}

// ---------------------------------------------------------------------------
// Address resolution tests
// ---------------------------------------------------------------------------

func TestAddressFunc(t *testing.T) {
	t.Run("default to localhost", func(t *testing.T) {
		hc := NewHealthChecker()
		host, err := hc.resolveAddress("any-container")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if host != "127.0.0.1" {
			t.Errorf("expected 127.0.0.1, got %s", host)
		}
	})

	t.Run("custom address func", func(t *testing.T) {
		hc := NewHealthChecker()
		hc.SetAddressFunc(func(containerID string) (string, error) {
			if containerID == "container-a" {
				return "10.0.0.1", nil
			}
			return "", fmt.Errorf("unknown container: %s", containerID)
		})

		host, err := hc.resolveAddress("container-a")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if host != "10.0.0.1" {
			t.Errorf("expected 10.0.0.1, got %s", host)
		}

		_, err = hc.resolveAddress("unknown")
		if err == nil {
			t.Error("expected error for unknown container")
		}
	})
}

func TestHTTPProbe_AddressResolutionFailure(t *testing.T) {
	hc := NewHealthChecker()
	hc.SetAddressFunc(func(containerID string) (string, error) {
		return "", fmt.Errorf("no address for container")
	})

	hc.RegisterProbe("no-addr", HealthProbeConfig{
		Type:             ProbeHTTP,
		HTTPPath:         "/health",
		HTTPPort:         8080,
		Timeout:          1 * time.Second,
		SuccessThreshold: 1,
		FailureThreshold: 1,
	})

	ok, err := hc.CheckHealth(context.Background(), "no-addr")
	if ok {
		t.Error("expected failure when address resolution fails")
	}
	if err == nil {
		t.Error("expected error when address resolution fails")
	}
	if !strings.Contains(err.Error(), "resolve container address") {
		t.Errorf("unexpected error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Concurrent access tests
// ---------------------------------------------------------------------------

func TestConcurrentCheckHealth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	host, port := parseHostPort(t, server.Listener.Addr().String())

	hc := NewHealthChecker()
	hc.SetAddressFunc(func(containerID string) (string, error) {
		return host, nil
	})

	// Register multiple probes
	for i := 0; i < 5; i++ {
		id := fmt.Sprintf("concurrent-%d", i)
		hc.RegisterProbe(id, HealthProbeConfig{
			Type:             ProbeHTTP,
			HTTPPath:         "/",
			HTTPPort:         port,
			Timeout:          5 * time.Second,
			SuccessThreshold: 1,
			FailureThreshold: 3,
		})
	}

	// Run health checks concurrently
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		id := fmt.Sprintf("concurrent-%d", i)
		go func(containerID string) {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				hc.CheckHealth(context.Background(), containerID)
			}
		}(id)
	}
	wg.Wait()

	// Verify all became healthy
	for i := 0; i < 5; i++ {
		id := fmt.Sprintf("concurrent-%d", i)
		status := hc.GetProbeStatus(id)
		if status == nil {
			t.Errorf("expected status for %s", id)
			continue
		}
		if status.State != ProbeStateHealthy {
			t.Errorf("container %s: expected healthy, got %s", id, status.State)
		}
		if status.TotalChecks != 10 {
			t.Errorf("container %s: expected 10 checks, got %d", id, status.TotalChecks)
		}
	}
}

func TestConcurrentRegisterUnregister(t *testing.T) {
	hc := NewHealthChecker()

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			id := fmt.Sprintf("container-%d", idx)
			hc.RegisterProbe(id, HealthProbeConfig{
				Type:    ProbeTCP,
				TCPPort: 8080 + idx,
			})
			hc.GetProbeStatus(id)
			hc.UnregisterProbe(id)
		}(i)
	}
	wg.Wait()
}

// ---------------------------------------------------------------------------
// Unknown probe type test
// ---------------------------------------------------------------------------

func TestUnknownProbeType(t *testing.T) {
	hc := NewHealthChecker()

	// Manually insert a probe with an unknown type
	hc.mu.Lock()
	hc.probes["unknown-type"] = &probeEntry{
		config: HealthProbeConfig{
			Type:             ProbeType("invalid"),
			Timeout:          5 * time.Second,
			SuccessThreshold: 1,
			FailureThreshold: 1,
		},
		status: ProbeStatus{
			ContainerID: "unknown-type",
			State:       ProbeStateUnknown,
		},
	}
	hc.mu.Unlock()

	_, err := hc.CheckHealth(context.Background(), "unknown-type")
	if err == nil {
		t.Error("expected error for unknown probe type")
	}
	if !strings.Contains(err.Error(), "unknown probe type") {
		t.Errorf("unexpected error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Reset test
// ---------------------------------------------------------------------------

func TestReset(t *testing.T) {
	hc := NewHealthChecker()

	ctx, cancel := context.WithCancel(context.Background())

	go hc.Start(ctx)
	time.Sleep(50 * time.Millisecond)

	// Stop via context
	cancel()
	time.Sleep(50 * time.Millisecond)

	// Reset should allow re-start
	hc.Reset()

	ctx2, cancel2 := context.WithCancel(context.Background())
	defer cancel2()

	go hc.Start(ctx2)
	time.Sleep(50 * time.Millisecond)

	if !hc.IsRunning() {
		t.Error("expected running after Reset + Start")
	}

	hc.Stop()
}

// ---------------------------------------------------------------------------
// LastError tracking test
// ---------------------------------------------------------------------------

func TestLastErrorTracking(t *testing.T) {
	hc := NewHealthChecker()
	hc.SetExecFunc(func(ctx context.Context, containerID string, cmd []string) (int, error) {
		return -1, fmt.Errorf("service crashed")
	})

	hc.RegisterProbe("error-track", HealthProbeConfig{
		Type:             ProbeExec,
		ExecCommand:      []string{"check"},
		Timeout:          5 * time.Second,
		SuccessThreshold: 1,
		FailureThreshold: 1,
	})

	hc.CheckHealth(context.Background(), "error-track")

	status := hc.GetProbeStatus("error-track")
	if status.LastError == "" {
		t.Error("expected last error to be set")
	}
	if !strings.Contains(status.LastError, "service crashed") {
		t.Errorf("unexpected last error: %s", status.LastError)
	}
}

func TestLastErrorClearedOnSuccess(t *testing.T) {
	callCount := 0
	hc := NewHealthChecker()
	hc.SetExecFunc(func(ctx context.Context, containerID string, cmd []string) (int, error) {
		callCount++
		if callCount == 1 {
			return -1, fmt.Errorf("temporary failure")
		}
		return 0, nil
	})

	hc.RegisterProbe("error-clear", HealthProbeConfig{
		Type:             ProbeExec,
		ExecCommand:      []string{"check"},
		Timeout:          5 * time.Second,
		SuccessThreshold: 1,
		FailureThreshold: 3,
	})

	ctx := context.Background()

	// First call fails
	hc.CheckHealth(ctx, "error-clear")
	status := hc.GetProbeStatus("error-clear")
	if status.LastError == "" {
		t.Error("expected last error after failure")
	}

	// Second call succeeds
	hc.CheckHealth(ctx, "error-clear")
	status = hc.GetProbeStatus("error-clear")
	if status.LastError != "" {
		t.Errorf("expected empty last error after success, got %q", status.LastError)
	}
}

// ---------------------------------------------------------------------------
// Context cancellation during probe
// ---------------------------------------------------------------------------

func TestCheckHealth_ContextCancelled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Block until context is cancelled
		<-r.Context().Done()
	}))
	defer server.Close()

	host, port := parseHostPort(t, server.Listener.Addr().String())

	hc := NewHealthChecker()
	hc.SetAddressFunc(func(containerID string) (string, error) {
		return host, nil
	})

	hc.RegisterProbe("ctx-cancel", HealthProbeConfig{
		Type:             ProbeHTTP,
		HTTPPath:         "/",
		HTTPPort:         port,
		Timeout:          10 * time.Second,
		SuccessThreshold: 1,
		FailureThreshold: 1,
	})

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		hc.CheckHealth(ctx, "ctx-cancel")
		close(done)
	}()

	// Cancel after a short delay
	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case <-done:
		// Expected
	case <-time.After(3 * time.Second):
		t.Fatal("CheckHealth did not return after context cancellation")
	}
}

// ---------------------------------------------------------------------------
// Helper functions
// ---------------------------------------------------------------------------

func parseHostPort(t *testing.T, addr string) (string, int) {
	t.Helper()
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		t.Fatalf("failed to parse address %q: %v", addr, err)
	}
	port := 0
	fmt.Sscanf(portStr, "%d", &port)
	if port == 0 {
		t.Fatalf("failed to parse port from %q", addr)
	}
	return host, port
}
