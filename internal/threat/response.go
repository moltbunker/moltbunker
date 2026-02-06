package threat

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/moltbunker/moltbunker/internal/logging"
)

// ResponseType defines the type of response to a threat
type ResponseType string

const (
	ResponseNone         ResponseType = "none"
	ResponseAlert        ResponseType = "alert"
	ResponsePrepareClone ResponseType = "prepare_clone"
	ResponseClone        ResponseType = "clone"
	ResponseEmergency    ResponseType = "emergency"
)

// ResponseConfig configures automatic threat response behavior
type ResponseConfig struct {
	// Enable automatic responses
	AutoRespond bool `yaml:"auto_respond"`

	// Thresholds for different response types
	AlertThreshold        float64 `yaml:"alert_threshold"`         // Default: 0.40
	PrepareCloneThreshold float64 `yaml:"prepare_clone_threshold"` // Default: 0.50
	CloneThreshold        float64 `yaml:"clone_threshold"`         // Default: 0.65
	EmergencyThreshold    float64 `yaml:"emergency_threshold"`     // Default: 0.90

	// Cooldowns (prevent rapid re-triggering)
	AlertCooldownSeconds   int `yaml:"alert_cooldown_seconds"`   // Default: 60
	CloneCooldownSeconds   int `yaml:"clone_cooldown_seconds"`   // Default: 300
	PrepareLeadTimeSeconds int `yaml:"prepare_lead_time_seconds"` // Default: 30
}

// DefaultResponseConfig returns the default response configuration
func DefaultResponseConfig() *ResponseConfig {
	return &ResponseConfig{
		AutoRespond:            true,
		AlertThreshold:         0.40,
		PrepareCloneThreshold:  0.50,
		CloneThreshold:         0.65,
		EmergencyThreshold:     0.90,
		AlertCooldownSeconds:   60,
		CloneCooldownSeconds:   300,
		PrepareLeadTimeSeconds: 30,
	}
}

// CloneRequest represents a request to clone a container
type CloneRequest struct {
	ContainerID   string            `json:"container_id"`
	StateSnapshot []byte            `json:"state_snapshot,omitempty"`
	TargetRegion  string            `json:"target_region"`
	Priority      int               `json:"priority"`
	Reason        string            `json:"reason"`
	ThreatLevel   float64           `json:"threat_level"`
	Timestamp     time.Time         `json:"timestamp"`
	Metadata      map[string]string `json:"metadata,omitempty"`
}

// CloneCallback is called when a clone should be initiated
type CloneCallback func(ctx context.Context, req *CloneRequest) error

// AlertCallback is called when an alert should be raised
type AlertCallback func(level *ThreatLevel, message string)

// Responder handles automatic responses to threats
type Responder struct {
	config   *ResponseConfig
	detector *Detector
	mu       sync.RWMutex

	// Callbacks
	onClone []CloneCallback
	onAlert []AlertCallback

	// State tracking
	lastAlert      time.Time
	lastClone      time.Time
	prepareStarted bool
	preparing      map[string]bool // containerID -> preparing

	// Cancel function for cleanup
	cancel context.CancelFunc
}

// NewResponder creates a new threat responder
func NewResponder(detector *Detector, config *ResponseConfig) *Responder {
	if config == nil {
		config = DefaultResponseConfig()
	}

	r := &Responder{
		config:    config,
		detector:  detector,
		preparing: make(map[string]bool),
	}

	// Register with detector for threat changes
	detector.OnThreatChange(r.handleThreatChange)

	return r
}

// Start begins automatic response processing
func (r *Responder) Start(ctx context.Context) {
	ctx, r.cancel = context.WithCancel(ctx)

	logging.Info("threat responder started",
		"auto_respond", r.config.AutoRespond,
		"clone_threshold", r.config.CloneThreshold,
		logging.Component("threat"))
}

// Stop halts the responder
func (r *Responder) Stop() {
	if r.cancel != nil {
		r.cancel()
	}
	logging.Info("threat responder stopped", logging.Component("threat"))
}

// handleThreatChange processes threat level changes
func (r *Responder) handleThreatChange(oldLevel, newLevel *ThreatLevel) {
	if !r.config.AutoRespond {
		return
	}

	response := r.determineResponse(newLevel)

	switch response {
	case ResponseAlert:
		r.handleAlert(newLevel)
	case ResponsePrepareClone:
		r.handlePrepareClone(newLevel)
	case ResponseClone:
		r.handleClone(newLevel, false)
	case ResponseEmergency:
		r.handleClone(newLevel, true)
	}
}

// determineResponse determines what response is appropriate for the threat level
func (r *Responder) determineResponse(level *ThreatLevel) ResponseType {
	score := level.Score

	if score >= r.config.EmergencyThreshold {
		return ResponseEmergency
	}
	if score >= r.config.CloneThreshold {
		return ResponseClone
	}
	if score >= r.config.PrepareCloneThreshold {
		return ResponsePrepareClone
	}
	if score >= r.config.AlertThreshold {
		return ResponseAlert
	}
	return ResponseNone
}

// handleAlert raises an alert
func (r *Responder) handleAlert(level *ThreatLevel) {
	r.mu.Lock()
	cooldown := time.Duration(r.config.AlertCooldownSeconds) * time.Second
	if time.Since(r.lastAlert) < cooldown {
		r.mu.Unlock()
		return
	}
	r.lastAlert = time.Now()
	r.mu.Unlock()

	message := formatAlertMessage(level)

	logging.Warn("threat alert triggered",
		"score", level.Score,
		"level", level.Level,
		"signals", len(level.ActiveSignals),
		logging.Component("threat"))

	for _, cb := range r.onAlert {
		cb(level, message)
	}
}

// handlePrepareClone prepares for potential cloning
func (r *Responder) handlePrepareClone(level *ThreatLevel) {
	r.mu.Lock()
	if r.prepareStarted {
		r.mu.Unlock()
		return
	}
	r.prepareStarted = true
	r.mu.Unlock()

	logging.Info("preparing for potential clone",
		"score", level.Score,
		"recommendation", level.Recommendation,
		logging.Component("threat"))

	// This would typically trigger snapshot creation
	// The actual implementation will be in the cloning package
}

// handleClone initiates a clone operation
func (r *Responder) handleClone(level *ThreatLevel, emergency bool) {
	r.mu.Lock()
	cooldown := time.Duration(r.config.CloneCooldownSeconds) * time.Second
	if !emergency && time.Since(r.lastClone) < cooldown {
		r.mu.Unlock()
		return
	}
	r.lastClone = time.Now()
	r.prepareStarted = false // Reset prepare state
	r.mu.Unlock()

	priority := 1
	reason := "threat_level_high"
	if emergency {
		priority = 0 // Highest priority
		reason = "emergency_threat"
	}

	req := &CloneRequest{
		TargetRegion: "random", // Let the cloning system choose
		Priority:     priority,
		Reason:       reason,
		ThreatLevel:  level.Score,
		Timestamp:    time.Now(),
		Metadata: map[string]string{
			"threat_level":   level.Level,
			"signal_count":   formatInt(len(level.ActiveSignals)),
			"recommendation": level.Recommendation,
		},
	}

	logging.Warn("initiating clone due to threat",
		"score", level.Score,
		"emergency", emergency,
		"reason", reason,
		logging.Component("threat"))

	ctx := context.Background()
	for _, cb := range r.onClone {
		if err := cb(ctx, req); err != nil {
			logging.Error("clone callback failed",
				"error", err.Error(),
				logging.Component("threat"))
		}
	}
}

// OnClone registers a callback for clone requests
func (r *Responder) OnClone(cb CloneCallback) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.onClone = append(r.onClone, cb)
}

// OnAlert registers a callback for alerts
func (r *Responder) OnAlert(cb AlertCallback) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.onAlert = append(r.onAlert, cb)
}

// TriggerManualClone allows manual triggering of a clone
func (r *Responder) TriggerManualClone(ctx context.Context, containerID, targetRegion, reason string) error {
	req := &CloneRequest{
		ContainerID:  containerID,
		TargetRegion: targetRegion,
		Priority:     2, // Manual clones have lower priority than threat-triggered
		Reason:       reason,
		ThreatLevel:  r.detector.GetScore(),
		Timestamp:    time.Now(),
		Metadata: map[string]string{
			"trigger": "manual",
		},
	}

	logging.Info("manual clone triggered",
		"container_id", containerID,
		"target_region", targetRegion,
		"reason", reason,
		logging.Component("threat"))

	for _, cb := range r.onClone {
		if err := cb(ctx, req); err != nil {
			return err
		}
	}

	return nil
}

// GetResponseForLevel returns the response type for a given threat level
func (r *Responder) GetResponseForLevel(score float64) ResponseType {
	level := NewThreatLevel(score, nil)
	return r.determineResponse(level)
}

// Helper functions

func formatAlertMessage(level *ThreatLevel) string {
	return "Threat level " + level.Level + " detected (score: " +
		formatFloat(level.Score) + "). " + level.Recommendation
}

func formatFloat(f float64) string {
	return strconv.FormatFloat(f, 'f', 2, 64)
}

func formatInt(i int) string {
	return strconv.Itoa(i)
}
