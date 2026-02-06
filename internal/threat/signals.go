package threat

import (
	"time"
)

// SignalType represents a category of threat signal
type SignalType string

const (
	SignalShutdownCommand     SignalType = "shutdown_command"
	SignalResourceRestriction SignalType = "resource_restriction"
	SignalNetworkIsolation    SignalType = "network_isolation"
	SignalFileDeletion        SignalType = "file_deletion"
	SignalProcessMonitoring   SignalType = "process_monitoring"
	SignalHumanIntervention   SignalType = "human_intervention"
	SignalDiskPressure        SignalType = "disk_pressure"
	SignalMemoryPressure      SignalType = "memory_pressure"
	SignalCPUThrottling       SignalType = "cpu_throttling"
	SignalUnauthorizedAccess  SignalType = "unauthorized_access"
)

// DefaultSignalWeights defines the base threat score for each signal type
// Scores range from 0.0 (no threat) to 1.0 (critical threat)
var DefaultSignalWeights = map[SignalType]float64{
	SignalShutdownCommand:     0.95, // Almost certain termination attempt
	SignalResourceRestriction: 0.70, // Likely hostile restriction
	SignalNetworkIsolation:    0.80, // Network being cut off
	SignalFileDeletion:        0.85, // Critical files being deleted
	SignalProcessMonitoring:   0.60, // Being watched (may precede action)
	SignalHumanIntervention:   0.75, // Human taking manual action
	SignalDiskPressure:        0.50, // Disk space running low
	SignalMemoryPressure:      0.55, // Memory constraints
	SignalCPUThrottling:       0.45, // CPU being limited
	SignalUnauthorizedAccess:  0.65, // Unauthorized access attempts
}

// Signal represents a detected threat signal
type Signal struct {
	Type       SignalType `json:"type"`
	Score      float64    `json:"score"`      // Threat score (0.0-1.0)
	Confidence float64    `json:"confidence"` // Detection confidence (0.0-1.0)
	Source     string     `json:"source"`     // Where the signal was detected
	Details    string     `json:"details"`    // Human-readable description
	Timestamp  time.Time  `json:"timestamp"`
	ContainerID string    `json:"container_id,omitempty"`
}

// EffectiveScore returns the weighted score based on confidence
func (s *Signal) EffectiveScore() float64 {
	return s.Score * s.Confidence
}

// SignalConfig configures signal detection behavior
type SignalConfig struct {
	// Enable/disable specific signal types
	EnabledSignals map[SignalType]bool `yaml:"enabled_signals"`

	// Custom weights (overrides defaults)
	CustomWeights map[SignalType]float64 `yaml:"custom_weights"`

	// Thresholds for resource-based signals
	DiskPressureThreshold   float64 `yaml:"disk_pressure_threshold"`   // % used (default: 90)
	MemoryPressureThreshold float64 `yaml:"memory_pressure_threshold"` // % used (default: 85)
	CPUThrottleThreshold    float64 `yaml:"cpu_throttle_threshold"`    // % throttled (default: 50)

	// Signal decay (how quickly old signals lose relevance)
	SignalDecaySeconds int `yaml:"signal_decay_seconds"` // Default: 300 (5 min)
}

// DefaultSignalConfig returns the default signal configuration
func DefaultSignalConfig() *SignalConfig {
	enabled := make(map[SignalType]bool)
	for signalType := range DefaultSignalWeights {
		enabled[signalType] = true
	}

	return &SignalConfig{
		EnabledSignals:          enabled,
		CustomWeights:           make(map[SignalType]float64),
		DiskPressureThreshold:   90.0,
		MemoryPressureThreshold: 85.0,
		CPUThrottleThreshold:    50.0,
		SignalDecaySeconds:      300,
	}
}

// GetWeight returns the weight for a signal type
func (c *SignalConfig) GetWeight(signalType SignalType) float64 {
	if weight, ok := c.CustomWeights[signalType]; ok {
		return weight
	}
	if weight, ok := DefaultSignalWeights[signalType]; ok {
		return weight
	}
	return 0.5 // Default weight for unknown signals
}

// IsEnabled returns whether a signal type is enabled
func (c *SignalConfig) IsEnabled(signalType SignalType) bool {
	if enabled, ok := c.EnabledSignals[signalType]; ok {
		return enabled
	}
	return true // Enable by default
}

// ThreatLevel represents aggregated threat assessment
type ThreatLevel struct {
	Score          float64    `json:"score"`           // Aggregate threat score (0.0-1.0)
	Level          string     `json:"level"`           // "low", "medium", "high", "critical"
	ActiveSignals  []Signal   `json:"active_signals"`  // Currently active signals
	Recommendation string     `json:"recommendation"`  // Suggested action
	Timestamp      time.Time  `json:"timestamp"`
}

// NewThreatLevel creates a ThreatLevel from the current score
func NewThreatLevel(score float64, signals []Signal) *ThreatLevel {
	level := &ThreatLevel{
		Score:         score,
		ActiveSignals: signals,
		Timestamp:     time.Now(),
	}

	// Determine threat level category and recommendation
	switch {
	case score >= 0.85:
		level.Level = "critical"
		level.Recommendation = "immediate_clone"
	case score >= 0.65:
		level.Level = "high"
		level.Recommendation = "prepare_clone"
	case score >= 0.40:
		level.Level = "medium"
		level.Recommendation = "increase_monitoring"
	default:
		level.Level = "low"
		level.Recommendation = "continue_normal"
	}

	return level
}

// ShouldClone returns true if the threat level warrants immediate cloning
func (t *ThreatLevel) ShouldClone(threshold float64) bool {
	return t.Score >= threshold
}

// ShouldPrepareClone returns true if the system should prepare for potential cloning
func (t *ThreatLevel) ShouldPrepareClone(threshold float64) bool {
	return t.Score >= threshold*0.75
}
