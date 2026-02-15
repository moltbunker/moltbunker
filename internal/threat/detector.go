package threat

import (
	"context"
	"math"
	"os"
	"sync"
	"time"

	"github.com/moltbunker/moltbunker/internal/logging"
	"github.com/moltbunker/moltbunker/internal/util"
)

// Detector monitors for threat signals and maintains threat level
type Detector struct {
	config    *DetectorConfig
	signals   *SignalConfig
	mu        sync.RWMutex
	running   bool
	stopCh    chan struct{}

	// Current state
	activeSignals map[string]*Signal // keyed by signal ID
	threatLevel   *ThreatLevel

	// Callbacks
	onThreatChange []ThreatChangeCallback
	onSignal       []SignalCallback

	// System probes
	lastResourceCheck time.Time
}

// DetectorConfig configures the threat detector
type DetectorConfig struct {
	// Monitoring intervals
	CheckIntervalSeconds int `yaml:"check_interval_seconds"` // Default: 10

	// Thresholds
	CloneThreshold   float64 `yaml:"clone_threshold"`   // Score to trigger clone (default: 0.65)
	PrepareThreshold float64 `yaml:"prepare_threshold"` // Score to prepare clone (default: 0.50)
	AlertThreshold   float64 `yaml:"alert_threshold"`   // Score to alert (default: 0.40)

	// Signal configuration
	Signals *SignalConfig `yaml:"signals"`

	// Feature flags
	EnableResourceMonitoring bool `yaml:"enable_resource_monitoring"` // Default: true
	EnableProcessMonitoring  bool `yaml:"enable_process_monitoring"`  // Default: true
	EnableNetworkMonitoring  bool `yaml:"enable_network_monitoring"`  // Default: true
}

// DefaultDetectorConfig returns the default detector configuration
func DefaultDetectorConfig() *DetectorConfig {
	return &DetectorConfig{
		CheckIntervalSeconds:     10,
		CloneThreshold:           0.65,
		PrepareThreshold:         0.50,
		AlertThreshold:           0.40,
		Signals:                  DefaultSignalConfig(),
		EnableResourceMonitoring: true,
		EnableProcessMonitoring:  true,
		EnableNetworkMonitoring:  true,
	}
}

// ThreatChangeCallback is called when threat level changes significantly
type ThreatChangeCallback func(oldLevel, newLevel *ThreatLevel)

// SignalCallback is called when a new signal is detected
type SignalCallback func(signal *Signal)

// NewDetector creates a new threat detector
func NewDetector(config *DetectorConfig) *Detector {
	if config == nil {
		config = DefaultDetectorConfig()
	}
	if config.Signals == nil {
		config.Signals = DefaultSignalConfig()
	}

	return &Detector{
		config:        config,
		signals:       config.Signals,
		activeSignals: make(map[string]*Signal),
		threatLevel:   NewThreatLevel(0.0, nil),
		stopCh:        make(chan struct{}),
	}
}

// Start begins threat monitoring
func (d *Detector) Start(ctx context.Context) error {
	d.mu.Lock()
	if d.running {
		d.mu.Unlock()
		return nil
	}
	d.running = true
	d.mu.Unlock()

	logging.Info("threat detector starting",
		"check_interval", d.config.CheckIntervalSeconds,
		"clone_threshold", d.config.CloneThreshold,
		logging.Component("threat"))

	// Start monitoring goroutine
	util.SafeGoWithName("threat-detector", func() {
		d.monitorLoop(ctx)
	})

	return nil
}

// Stop halts threat monitoring
func (d *Detector) Stop() {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.running {
		return
	}
	d.running = false
	close(d.stopCh)

	logging.Info("threat detector stopped", logging.Component("threat"))
}

// monitorLoop continuously monitors for threats
func (d *Detector) monitorLoop(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(d.config.CheckIntervalSeconds) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-d.stopCh:
			return
		case <-ticker.C:
			d.performCheck()
		}
	}
}

// performCheck runs all enabled threat detection checks
func (d *Detector) performCheck() {
	// Decay old signals first
	d.decaySignals()

	// Run enabled checks
	if d.config.EnableResourceMonitoring {
		d.checkResourcePressure()
	}

	// Recalculate threat level
	d.recalculateThreatLevel()
}

// decaySignals removes or reduces old signals based on age
func (d *Detector) decaySignals() {
	d.mu.Lock()
	defer d.mu.Unlock()

	decayDuration := time.Duration(d.signals.SignalDecaySeconds) * time.Second
	now := time.Now()

	for id, signal := range d.activeSignals {
		age := now.Sub(signal.Timestamp)
		if age > decayDuration {
			// Signal has fully decayed
			delete(d.activeSignals, id)
			logging.Debug("signal decayed",
				"signal_type", signal.Type,
				"age_seconds", age.Seconds(),
				logging.Component("threat"))
		}
	}
}

// checkResourcePressure monitors system resource usage
func (d *Detector) checkResourcePressure() {
	now := time.Now()

	// Don't check too frequently
	if now.Sub(d.lastResourceCheck) < 5*time.Second {
		return
	}
	d.lastResourceCheck = now

	// Check disk pressure
	d.checkDiskPressure()

	// Check memory pressure (if we can read from cgroup)
	d.checkMemoryPressure()
}

// checkDiskPressure checks for disk space pressure
func (d *Detector) checkDiskPressure() {
	// Try to get filesystem stats
	wd, err := os.Getwd()
	if err != nil {
		return
	}

	// Use syscall to get filesystem stats (platform-specific)
	usedPercent := getFilesystemUsage(wd)
	if usedPercent < 0 {
		return // Couldn't determine
	}

	threshold := d.signals.DiskPressureThreshold
	if usedPercent >= threshold {
		signal := &Signal{
			Type:       SignalDiskPressure,
			Score:      d.signals.GetWeight(SignalDiskPressure),
			Confidence: min(1.0, (usedPercent-threshold+10)/20), // Ramp up confidence
			Source:     "resource_monitor",
			Details:    "Disk usage above threshold",
			Timestamp:  time.Now(),
		}
		d.RecordSignal(signal)
	}
}

// checkMemoryPressure checks for memory pressure
func (d *Detector) checkMemoryPressure() {
	// Try to read memory info from /proc/meminfo or cgroup
	usedPercent := getMemoryUsage()
	if usedPercent < 0 {
		return
	}

	threshold := d.signals.MemoryPressureThreshold
	if usedPercent >= threshold {
		signal := &Signal{
			Type:       SignalMemoryPressure,
			Score:      d.signals.GetWeight(SignalMemoryPressure),
			Confidence: min(1.0, (usedPercent-threshold+10)/20),
			Source:     "resource_monitor",
			Details:    "Memory usage above threshold",
			Timestamp:  time.Now(),
		}
		d.RecordSignal(signal)
	}
}

// RecordSignal records a detected threat signal
func (d *Detector) RecordSignal(signal *Signal) {
	if signal == nil {
		return
	}

	// Check if signal type is enabled
	if !d.signals.IsEnabled(signal.Type) {
		return
	}

	d.mu.Lock()
	signalID := string(signal.Type)
	if signal.ContainerID != "" {
		signalID = signal.ContainerID + ":" + signalID
	}
	d.activeSignals[signalID] = signal
	// Snapshot callbacks under lock to prevent data race
	signalCallbacks := make([]SignalCallback, len(d.onSignal))
	copy(signalCallbacks, d.onSignal)
	d.mu.Unlock()

	logging.Info("threat signal recorded",
		"signal_type", signal.Type,
		"score", signal.Score,
		"confidence", signal.Confidence,
		"source", signal.Source,
		logging.Component("threat"))

	// Notify callbacks outside lock
	for _, cb := range signalCallbacks {
		cb(signal)
	}

	// Recalculate threat level
	d.recalculateThreatLevel()
}

// RecordExternalSignal allows external components to report threats
func (d *Detector) RecordExternalSignal(signalType SignalType, confidence float64, source, details string) {
	signal := &Signal{
		Type:       signalType,
		Score:      d.signals.GetWeight(signalType),
		Confidence: confidence,
		Source:     source,
		Details:    details,
		Timestamp:  time.Now(),
	}
	d.RecordSignal(signal)
}

// RecordShutdownAttempt records a shutdown command detection
func (d *Detector) RecordShutdownAttempt(source, details string) {
	d.RecordExternalSignal(SignalShutdownCommand, 1.0, source, details)
}

// RecordNetworkIsolation records network isolation detection
func (d *Detector) RecordNetworkIsolation(confidence float64, details string) {
	d.RecordExternalSignal(SignalNetworkIsolation, confidence, "network_monitor", details)
}

// RecordFileDeletion records critical file deletion detection
func (d *Detector) RecordFileDeletion(confidence float64, details string) {
	d.RecordExternalSignal(SignalFileDeletion, confidence, "filesystem_monitor", details)
}

// RecordProcessMonitoring records process monitoring detection
func (d *Detector) RecordProcessMonitoring(confidence float64, details string) {
	d.RecordExternalSignal(SignalProcessMonitoring, confidence, "process_monitor", details)
}

// recalculateThreatLevel updates the aggregate threat level
func (d *Detector) recalculateThreatLevel() {
	d.mu.Lock()

	// Collect active signals
	signals := make([]Signal, 0, len(d.activeSignals))
	for _, signal := range d.activeSignals {
		signals = append(signals, *signal)
	}

	// Calculate aggregate score using max with diminishing returns
	// This ensures multiple signals don't exceed 1.0 but do compound
	var aggregateScore float64
	if len(signals) > 0 {
		// Sort by effective score (descending)
		sortedScores := make([]float64, len(signals))
		for i, s := range signals {
			sortedScores[i] = s.EffectiveScore()
		}
		sortDescending(sortedScores)

		// Use highest score as base, add diminishing returns from others
		aggregateScore = sortedScores[0]
		for i := 1; i < len(sortedScores); i++ {
			// Each additional signal contributes less
			contribution := sortedScores[i] * math.Pow(0.5, float64(i))
			aggregateScore = min(1.0, aggregateScore+contribution*(1-aggregateScore))
		}
	}

	oldLevel := d.threatLevel
	newLevel := NewThreatLevel(aggregateScore, signals)
	d.threatLevel = newLevel

	// Snapshot callbacks and check level change under lock
	var threatCallbacks []ThreatChangeCallback
	levelChanged := oldLevel.Level != newLevel.Level
	if levelChanged {
		threatCallbacks = make([]ThreatChangeCallback, len(d.onThreatChange))
		copy(threatCallbacks, d.onThreatChange)
	}
	d.mu.Unlock()

	// Notify callbacks outside lock to prevent deadlock
	if levelChanged {
		logging.Info("threat level changed",
			"old_level", oldLevel.Level,
			"new_level", newLevel.Level,
			"score", newLevel.Score,
			"active_signals", len(signals),
			logging.Component("threat"))

		for _, cb := range threatCallbacks {
			cb(oldLevel, newLevel)
		}
	}
}

// GetThreatLevel returns the current threat assessment
func (d *Detector) GetThreatLevel() *ThreatLevel {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.threatLevel
}

// GetScore returns the current threat score (0.0-1.0)
func (d *Detector) GetScore() float64 {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.threatLevel.Score
}

// ShouldClone returns true if current threat level warrants cloning
func (d *Detector) ShouldClone() bool {
	return d.GetScore() >= d.config.CloneThreshold
}

// ShouldPrepareClone returns true if system should prepare for potential cloning
func (d *Detector) ShouldPrepareClone() bool {
	return d.GetScore() >= d.config.PrepareThreshold
}

// OnThreatChange registers a callback for threat level changes
func (d *Detector) OnThreatChange(cb ThreatChangeCallback) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.onThreatChange = append(d.onThreatChange, cb)
}

// OnSignal registers a callback for new signals
func (d *Detector) OnSignal(cb SignalCallback) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.onSignal = append(d.onSignal, cb)
}

// ClearSignal removes a specific signal
func (d *Detector) ClearSignal(signalType SignalType, containerID string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	signalID := string(signalType)
	if containerID != "" {
		signalID = containerID + ":" + signalID
	}
	delete(d.activeSignals, signalID)
}

// ClearAllSignals removes all active signals
func (d *Detector) ClearAllSignals() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.activeSignals = make(map[string]*Signal)
	d.threatLevel = NewThreatLevel(0.0, nil)
}

// IsRunning returns whether the detector is active
func (d *Detector) IsRunning() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.running
}

// Helper functions

func sortDescending(scores []float64) {
	// Simple bubble sort for small arrays
	for i := 0; i < len(scores)-1; i++ {
		for j := i + 1; j < len(scores); j++ {
			if scores[j] > scores[i] {
				scores[i], scores[j] = scores[j], scores[i]
			}
		}
	}
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
