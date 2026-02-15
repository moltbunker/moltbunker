package threat

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"
)

// --- Signal tests ---

func TestSignalEffectiveScore(t *testing.T) {
	tests := []struct {
		name       string
		score      float64
		confidence float64
		expected   float64
	}{
		{"full confidence", 0.8, 1.0, 0.8},
		{"half confidence", 0.8, 0.5, 0.4},
		{"zero confidence", 0.95, 0.0, 0.0},
		{"full both", 1.0, 1.0, 1.0},
		{"zero score", 0.0, 1.0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Signal{Score: tt.score, Confidence: tt.confidence}
			got := s.EffectiveScore()
			if got != tt.expected {
				t.Errorf("EffectiveScore() = %f, want %f", got, tt.expected)
			}
		})
	}
}

func TestDefaultSignalWeights(t *testing.T) {
	// Verify all expected signal types have weights
	expectedSignals := []SignalType{
		SignalShutdownCommand,
		SignalResourceRestriction,
		SignalNetworkIsolation,
		SignalFileDeletion,
		SignalProcessMonitoring,
		SignalHumanIntervention,
		SignalDiskPressure,
		SignalMemoryPressure,
		SignalCPUThrottling,
		SignalUnauthorizedAccess,
	}

	for _, sig := range expectedSignals {
		weight, ok := DefaultSignalWeights[sig]
		if !ok {
			t.Errorf("missing weight for signal type %s", sig)
		}
		if weight < 0 || weight > 1.0 {
			t.Errorf("weight for %s out of range: %f", sig, weight)
		}
	}

	// Shutdown should be the highest
	if DefaultSignalWeights[SignalShutdownCommand] < 0.9 {
		t.Error("expected shutdown command to have very high weight")
	}

	// CPU throttling should be relatively low
	if DefaultSignalWeights[SignalCPUThrottling] > 0.5 {
		t.Error("expected CPU throttling to have moderate weight")
	}
}

func TestDefaultSignalConfig(t *testing.T) {
	cfg := DefaultSignalConfig()
	if cfg == nil {
		t.Fatal("DefaultSignalConfig returned nil")
	}

	// All signals should be enabled
	for sig := range DefaultSignalWeights {
		if !cfg.IsEnabled(sig) {
			t.Errorf("signal %s should be enabled by default", sig)
		}
	}

	if cfg.DiskPressureThreshold != 90.0 {
		t.Errorf("expected disk pressure threshold 90, got %f", cfg.DiskPressureThreshold)
	}
	if cfg.MemoryPressureThreshold != 85.0 {
		t.Errorf("expected memory pressure threshold 85, got %f", cfg.MemoryPressureThreshold)
	}
	if cfg.CPUThrottleThreshold != 50.0 {
		t.Errorf("expected CPU throttle threshold 50, got %f", cfg.CPUThrottleThreshold)
	}
	if cfg.SignalDecaySeconds != 300 {
		t.Errorf("expected signal decay 300s, got %d", cfg.SignalDecaySeconds)
	}
}

func TestSignalConfigGetWeight(t *testing.T) {
	cfg := DefaultSignalConfig()

	// Default weight for known signal
	weight := cfg.GetWeight(SignalShutdownCommand)
	if weight != DefaultSignalWeights[SignalShutdownCommand] {
		t.Errorf("expected default weight for shutdown, got %f", weight)
	}

	// Custom weight overrides default
	cfg.CustomWeights[SignalShutdownCommand] = 0.5
	weight = cfg.GetWeight(SignalShutdownCommand)
	if weight != 0.5 {
		t.Errorf("expected custom weight 0.5, got %f", weight)
	}

	// Unknown signal returns 0.5
	weight = cfg.GetWeight(SignalType("unknown_signal"))
	if weight != 0.5 {
		t.Errorf("expected 0.5 for unknown signal, got %f", weight)
	}
}

func TestSignalConfigIsEnabled(t *testing.T) {
	cfg := DefaultSignalConfig()

	// Disable a signal
	cfg.EnabledSignals[SignalDiskPressure] = false
	if cfg.IsEnabled(SignalDiskPressure) {
		t.Error("expected disk pressure to be disabled")
	}

	// Unknown signal defaults to enabled
	if !cfg.IsEnabled(SignalType("unknown")) {
		t.Error("unknown signal should default to enabled")
	}
}

// --- ThreatLevel tests ---

func TestNewThreatLevel(t *testing.T) {
	tests := []struct {
		score          float64
		expectedLevel  string
		expectedRec    string
	}{
		{0.0, "low", "continue_normal"},
		{0.20, "low", "continue_normal"},
		{0.39, "low", "continue_normal"},
		{0.40, "medium", "increase_monitoring"},
		{0.50, "medium", "increase_monitoring"},
		{0.64, "medium", "increase_monitoring"},
		{0.65, "high", "prepare_clone"},
		{0.80, "high", "prepare_clone"},
		{0.84, "high", "prepare_clone"},
		{0.85, "critical", "immediate_clone"},
		{0.95, "critical", "immediate_clone"},
		{1.0, "critical", "immediate_clone"},
	}

	for _, tt := range tests {
		t.Run(tt.expectedLevel, func(t *testing.T) {
			level := NewThreatLevel(tt.score, nil)
			if level.Level != tt.expectedLevel {
				t.Errorf("score %f: expected level %s, got %s", tt.score, tt.expectedLevel, level.Level)
			}
			if level.Recommendation != tt.expectedRec {
				t.Errorf("score %f: expected recommendation %s, got %s", tt.score, tt.expectedRec, level.Recommendation)
			}
			if level.Score != tt.score {
				t.Errorf("expected score %f, got %f", tt.score, level.Score)
			}
		})
	}
}

func TestThreatLevelShouldClone(t *testing.T) {
	level := NewThreatLevel(0.7, nil)

	if !level.ShouldClone(0.65) {
		t.Error("expected ShouldClone true at 0.7 with threshold 0.65")
	}
	if level.ShouldClone(0.75) {
		t.Error("expected ShouldClone false at 0.7 with threshold 0.75")
	}
}

func TestThreatLevelShouldPrepareClone(t *testing.T) {
	level := NewThreatLevel(0.5, nil)

	// ShouldPrepareClone uses threshold * 0.75
	if !level.ShouldPrepareClone(0.65) { // 0.65 * 0.75 = 0.4875, 0.5 >= 0.4875
		t.Error("expected ShouldPrepareClone true")
	}
	if level.ShouldPrepareClone(0.7) { // 0.7 * 0.75 = 0.525, 0.5 < 0.525
		t.Error("expected ShouldPrepareClone false")
	}
}

// --- Detector tests ---

func TestNewDetector(t *testing.T) {
	t.Run("nil config uses defaults", func(t *testing.T) {
		d := NewDetector(nil)
		if d == nil {
			t.Fatal("NewDetector returned nil")
		}
		if d.config.CloneThreshold != 0.65 {
			t.Errorf("expected clone threshold 0.65, got %f", d.config.CloneThreshold)
		}
	})

	t.Run("custom config", func(t *testing.T) {
		cfg := &DetectorConfig{
			CloneThreshold: 0.80,
		}
		d := NewDetector(cfg)
		if d.config.CloneThreshold != 0.80 {
			t.Errorf("expected clone threshold 0.80, got %f", d.config.CloneThreshold)
		}
	})

	t.Run("nil signals gets defaults", func(t *testing.T) {
		d := NewDetector(&DetectorConfig{})
		if d.signals == nil {
			t.Fatal("expected signals to be initialized")
		}
	})
}

func TestDetectorRecordSignal(t *testing.T) {
	d := NewDetector(nil)

	// nil signal should be ignored
	d.RecordSignal(nil)

	// Record a signal
	signal := &Signal{
		Type:       SignalShutdownCommand,
		Score:      0.95,
		Confidence: 1.0,
		Source:     "test",
		Details:    "test signal",
		Timestamp:  time.Now(),
	}
	d.RecordSignal(signal)

	// Threat level should be raised
	level := d.GetThreatLevel()
	if level.Score == 0 {
		t.Error("expected non-zero threat score after recording signal")
	}
	if level.Score < 0.9 {
		t.Errorf("expected high score for shutdown command, got %f", level.Score)
	}
}

func TestDetectorRecordSignalDisabled(t *testing.T) {
	cfg := DefaultDetectorConfig()
	cfg.Signals.EnabledSignals[SignalDiskPressure] = false
	d := NewDetector(cfg)

	signal := &Signal{
		Type:       SignalDiskPressure,
		Score:      0.5,
		Confidence: 1.0,
		Timestamp:  time.Now(),
	}
	d.RecordSignal(signal)

	// Signal should have been ignored
	level := d.GetThreatLevel()
	if level.Score != 0 {
		t.Errorf("expected 0 score for disabled signal, got %f", level.Score)
	}
}

func TestDetectorRecordExternalSignal(t *testing.T) {
	d := NewDetector(nil)
	d.RecordExternalSignal(SignalFileDeletion, 0.9, "test_source", "files removed")

	level := d.GetThreatLevel()
	if level.Score == 0 {
		t.Error("expected non-zero score after external signal")
	}
}

func TestDetectorRecordShutdownAttempt(t *testing.T) {
	d := NewDetector(nil)
	d.RecordShutdownAttempt("test", "SIGTERM received")

	level := d.GetThreatLevel()
	if level.Level != "critical" {
		t.Errorf("expected critical level for shutdown attempt, got %s", level.Level)
	}
}

func TestDetectorClearSignal(t *testing.T) {
	d := NewDetector(nil)

	d.RecordExternalSignal(SignalDiskPressure, 0.8, "test", "disk full")
	if d.GetScore() == 0 {
		t.Fatal("expected non-zero score")
	}

	d.ClearSignal(SignalDiskPressure, "")
	d.recalculateThreatLevel()

	if d.GetScore() != 0 {
		t.Errorf("expected 0 score after clearing signal, got %f", d.GetScore())
	}
}

func TestDetectorClearAllSignals(t *testing.T) {
	d := NewDetector(nil)

	d.RecordExternalSignal(SignalDiskPressure, 0.8, "test", "disk")
	d.RecordExternalSignal(SignalMemoryPressure, 0.9, "test", "memory")

	if d.GetScore() == 0 {
		t.Fatal("expected non-zero score")
	}

	d.ClearAllSignals()

	if d.GetScore() != 0 {
		t.Errorf("expected 0 score after clearing all signals, got %f", d.GetScore())
	}
}

func TestDetectorShouldClone(t *testing.T) {
	d := NewDetector(&DetectorConfig{
		CloneThreshold:   0.65,
		PrepareThreshold: 0.50,
	})

	if d.ShouldClone() {
		t.Error("should not clone with no signals")
	}

	// Add a high-severity signal
	d.RecordShutdownAttempt("test", "shutdown")
	if !d.ShouldClone() {
		t.Error("should clone after shutdown attempt")
	}
}

func TestDetectorShouldPrepareClone(t *testing.T) {
	d := NewDetector(&DetectorConfig{
		CloneThreshold:   0.65,
		PrepareThreshold: 0.50,
	})

	if d.ShouldPrepareClone() {
		t.Error("should not prepare clone with no signals")
	}

	// Add moderate signal
	d.RecordExternalSignal(SignalProcessMonitoring, 1.0, "test", "strace detected")
	if !d.ShouldPrepareClone() {
		t.Error("should prepare clone after process monitoring signal (score ~0.6)")
	}
}

func TestDetectorStartStop(t *testing.T) {
	d := NewDetector(&DetectorConfig{
		CheckIntervalSeconds: 3600, // Long interval
	})

	if d.IsRunning() {
		t.Error("should not be running initially")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := d.Start(ctx)
	if err != nil {
		t.Fatalf("Start() error: %v", err)
	}
	if !d.IsRunning() {
		t.Error("should be running after Start()")
	}

	// Double start should be no-op
	err = d.Start(ctx)
	if err != nil {
		t.Fatalf("second Start() error: %v", err)
	}

	d.Stop()
	if d.IsRunning() {
		t.Error("should not be running after Stop()")
	}
}

func TestDetectorOnThreatChange(t *testing.T) {
	d := NewDetector(nil)

	var callbackCalled bool
	var mu sync.Mutex
	d.OnThreatChange(func(oldLevel, newLevel *ThreatLevel) {
		mu.Lock()
		callbackCalled = true
		mu.Unlock()
	})

	// This should trigger a level change from low to critical
	d.RecordShutdownAttempt("test", "SIGTERM")

	mu.Lock()
	called := callbackCalled
	mu.Unlock()

	if !called {
		t.Error("expected threat change callback to be called")
	}
}

func TestDetectorOnSignal(t *testing.T) {
	d := NewDetector(nil)

	var receivedSignal *Signal
	var mu sync.Mutex
	d.OnSignal(func(signal *Signal) {
		mu.Lock()
		receivedSignal = signal
		mu.Unlock()
	})

	d.RecordExternalSignal(SignalFileDeletion, 0.8, "test", "file deleted")

	mu.Lock()
	sig := receivedSignal
	mu.Unlock()

	if sig == nil {
		t.Fatal("expected signal callback to be called")
	}
	if sig.Type != SignalFileDeletion {
		t.Errorf("expected file deletion signal, got %s", sig.Type)
	}
}

func TestDetectorContainerSpecificSignal(t *testing.T) {
	d := NewDetector(nil)

	signal := &Signal{
		Type:        SignalDiskPressure,
		Score:       0.5,
		Confidence:  1.0,
		Source:      "test",
		ContainerID: "container-123",
		Timestamp:   time.Now(),
	}
	d.RecordSignal(signal)

	// Clear the container-specific signal
	d.ClearSignal(SignalDiskPressure, "container-123")
	d.recalculateThreatLevel()

	if d.GetScore() != 0 {
		t.Errorf("expected 0 score after clearing container signal, got %f", d.GetScore())
	}
}

func TestDetectorMultipleSignalsAggregation(t *testing.T) {
	d := NewDetector(nil)

	// Add two signals - the aggregate should be less than sum but more than max
	d.RecordExternalSignal(SignalDiskPressure, 0.8, "test", "disk")
	scoreOneSignal := d.GetScore()

	d.RecordExternalSignal(SignalMemoryPressure, 0.8, "test", "memory")
	scoreTwoSignals := d.GetScore()

	if scoreTwoSignals <= scoreOneSignal {
		t.Error("two signals should produce higher aggregate score than one")
	}
	if scoreTwoSignals > 1.0 {
		t.Error("aggregate score should not exceed 1.0")
	}
}

func TestDecaySignals(t *testing.T) {
	cfg := DefaultDetectorConfig()
	cfg.Signals.SignalDecaySeconds = 1 // 1 second decay for testing
	d := NewDetector(cfg)

	signal := &Signal{
		Type:       SignalDiskPressure,
		Score:      0.5,
		Confidence: 1.0,
		Source:     "test",
		Timestamp:  time.Now().Add(-2 * time.Second), // Already expired
	}
	d.mu.Lock()
	d.activeSignals["test"] = signal
	d.mu.Unlock()

	d.decaySignals()

	d.mu.RLock()
	remaining := len(d.activeSignals)
	d.mu.RUnlock()

	if remaining != 0 {
		t.Errorf("expected 0 signals after decay, got %d", remaining)
	}
}

// --- sortDescending tests ---

func TestSortDescending(t *testing.T) {
	scores := []float64{0.3, 0.9, 0.1, 0.7, 0.5}
	sortDescending(scores)

	for i := 0; i < len(scores)-1; i++ {
		if scores[i] < scores[i+1] {
			t.Errorf("not sorted descending at index %d: %f < %f", i, scores[i], scores[i+1])
		}
	}

	// Edge cases
	sortDescending([]float64{})
	sortDescending([]float64{1.0})
}

// --- Responder tests ---

func TestDefaultResponseConfig(t *testing.T) {
	cfg := DefaultResponseConfig()
	if cfg == nil {
		t.Fatal("DefaultResponseConfig returned nil")
	}
	if !cfg.AutoRespond {
		t.Error("expected auto respond true by default")
	}
	if cfg.AlertThreshold != 0.40 {
		t.Errorf("expected alert threshold 0.40, got %f", cfg.AlertThreshold)
	}
	if cfg.CloneThreshold != 0.65 {
		t.Errorf("expected clone threshold 0.65, got %f", cfg.CloneThreshold)
	}
	if cfg.EmergencyThreshold != 0.90 {
		t.Errorf("expected emergency threshold 0.90, got %f", cfg.EmergencyThreshold)
	}
}

func TestResponderDetermineResponse(t *testing.T) {
	detector := NewDetector(nil)
	r := NewResponder(detector, nil)

	tests := []struct {
		score    float64
		expected ResponseType
	}{
		{0.0, ResponseNone},
		{0.20, ResponseNone},
		{0.39, ResponseNone},
		{0.40, ResponseAlert},
		{0.49, ResponseAlert},
		{0.50, ResponsePrepareClone},
		{0.64, ResponsePrepareClone},
		{0.65, ResponseClone},
		{0.89, ResponseClone},
		{0.90, ResponseEmergency},
		{1.0, ResponseEmergency},
	}

	for _, tt := range tests {
		t.Run(formatFloat(tt.score), func(t *testing.T) {
			got := r.GetResponseForLevel(tt.score)
			if got != tt.expected {
				t.Errorf("score %f: expected %s, got %s", tt.score, tt.expected, got)
			}
		})
	}
}

func TestResponderOnClone(t *testing.T) {
	detector := NewDetector(nil)
	r := NewResponder(detector, nil)

	var cloneRequested bool
	var mu sync.Mutex
	r.OnClone(func(ctx context.Context, req *CloneRequest) error {
		mu.Lock()
		cloneRequested = true
		mu.Unlock()
		return nil
	})

	r.mu.Lock()
	count := len(r.onClone)
	r.mu.Unlock()

	if count != 1 {
		t.Errorf("expected 1 clone callback, got %d", count)
	}

	// Now verify it can be triggered via manual clone
	err := r.TriggerManualClone(context.Background(), "c1", "americas", "test")
	if err != nil {
		t.Fatalf("TriggerManualClone() error: %v", err)
	}

	mu.Lock()
	requested := cloneRequested
	mu.Unlock()

	if !requested {
		t.Error("expected clone callback to be called")
	}
}

func TestResponderOnAlert(t *testing.T) {
	detector := NewDetector(nil)
	r := NewResponder(detector, nil)

	var alertReceived bool
	r.OnAlert(func(level *ThreatLevel, message string) {
		alertReceived = true
	})

	r.mu.Lock()
	count := len(r.onAlert)
	r.mu.Unlock()

	if count != 1 {
		t.Errorf("expected 1 alert callback, got %d", count)
	}

	// Manually trigger alert handler
	level := NewThreatLevel(0.45, nil)
	r.handleAlert(level)

	if !alertReceived {
		t.Error("expected alert callback to be called")
	}
}

func TestResponderAlertCooldown(t *testing.T) {
	detector := NewDetector(nil)
	cfg := DefaultResponseConfig()
	cfg.AlertCooldownSeconds = 10 // 10 second cooldown
	r := NewResponder(detector, cfg)

	alertCount := 0
	r.OnAlert(func(level *ThreatLevel, message string) {
		alertCount++
	})

	level := NewThreatLevel(0.45, nil)

	// First alert should trigger
	r.handleAlert(level)
	if alertCount != 1 {
		t.Errorf("expected 1 alert, got %d", alertCount)
	}

	// Second alert within cooldown should not trigger
	r.handleAlert(level)
	if alertCount != 1 {
		t.Errorf("expected still 1 alert during cooldown, got %d", alertCount)
	}
}

func TestResponderStartStop(t *testing.T) {
	detector := NewDetector(nil)
	r := NewResponder(detector, nil)

	ctx := context.Background()
	r.Start(ctx)

	// Should not panic on stop
	r.Stop()
}

// --- Helper function tests ---

func TestFormatAlertMessage(t *testing.T) {
	level := NewThreatLevel(0.75, nil)
	msg := formatAlertMessage(level)

	if msg == "" {
		t.Error("expected non-empty alert message")
	}
	if len(msg) < 10 {
		t.Error("alert message seems too short")
	}
}

func TestFormatFloat(t *testing.T) {
	result := formatFloat(0.65)
	if result != "0.65" {
		t.Errorf("expected '0.65', got '%s'", result)
	}
}

func TestFormatInt(t *testing.T) {
	result := formatInt(42)
	if result != "42" {
		t.Errorf("expected '42', got '%s'", result)
	}
}

// --- Probe tests (pure logic only) ---

func TestDefaultProbeConfig(t *testing.T) {
	cfg := DefaultProbeConfig()
	if cfg == nil {
		t.Fatal("DefaultProbeConfig returned nil")
	}
	if !cfg.EnableNetworkProbe {
		t.Error("expected network probe enabled")
	}
	if !cfg.EnableFileWatcher {
		t.Error("expected file watcher enabled")
	}
	if !cfg.EnableSignalMonitor {
		t.Error("expected signal monitor enabled")
	}
	if !cfg.EnableProcessMonitor {
		t.Error("expected process monitor enabled")
	}
	if len(cfg.NetworkTargets) == 0 {
		t.Error("expected non-empty network targets")
	}
	if len(cfg.SuspiciousProcesses) == 0 {
		t.Error("expected non-empty suspicious processes list")
	}
	if len(cfg.CriticalPaths) == 0 {
		t.Error("expected non-empty critical paths")
	}
}

func TestIsNumericString(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"123", true},
		{"0", true},
		{"999999", true},
		{"", false},
		{"abc", false},
		{"12a", false},
		{"-1", false},
		{"1.5", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := isNumericString(tt.input)
			if got != tt.expected {
				t.Errorf("isNumericString(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestMinFloat64(t *testing.T) {
	if minFloat64(1.0, 2.0) != 1.0 {
		t.Error("expected 1.0")
	}
	if minFloat64(2.0, 1.0) != 1.0 {
		t.Error("expected 1.0")
	}
	if minFloat64(1.0, 1.0) != 1.0 {
		t.Error("expected 1.0")
	}
}

func TestMin(t *testing.T) {
	if min(1.0, 2.0) != 1.0 {
		t.Error("expected 1.0")
	}
	if min(2.0, 1.0) != 1.0 {
		t.Error("expected 1.0")
	}
	if min(1.0, 1.0) != 1.0 {
		t.Error("expected 1.0")
	}
}

func TestGetFileStateSafe(t *testing.T) {
	tmpDir := t.TempDir()

	// Nonexistent file
	state, err := getFileStateSafe("/nonexistent/file/path")
	if err != nil {
		t.Fatalf("expected no error for nonexistent file, got: %v", err)
	}
	if state.exists {
		t.Error("expected exists=false for nonexistent file")
	}

	// Create a test file
	testFile := tmpDir + "/testfile"
	if err := writeTestFile(testFile); err != nil {
		t.Fatal(err)
	}

	state, err = getFileStateSafe(testFile)
	if err != nil {
		t.Fatalf("getFileStateSafe() error: %v", err)
	}
	if !state.exists {
		t.Error("expected exists=true for existing file")
	}
	if state.size <= 0 {
		t.Error("expected positive size")
	}
}

func writeTestFile(path string) error {
	return os.WriteFile(path, []byte("test content"), 0644)
}

func TestNewProbeManager(t *testing.T) {
	detector := NewDetector(nil)

	t.Run("nil config uses defaults", func(t *testing.T) {
		pm := NewProbeManager(nil, detector)
		if pm == nil {
			t.Fatal("NewProbeManager returned nil")
		}
	})

	t.Run("custom config", func(t *testing.T) {
		cfg := &ProbeConfig{
			EnableNetworkProbe: false,
		}
		pm := NewProbeManager(cfg, detector)
		if pm.config.EnableNetworkProbe {
			t.Error("expected network probe disabled")
		}
	})
}
