package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/moltbunker/moltbunker/internal/threat"
)

// ThreatDetector is set externally after daemon startup
var threatDetector *threat.Detector

// SetThreatDetector sets the global threat detector
func SetThreatDetector(d *threat.Detector) {
	threatDetector = d
}

// ThreatLevelResponse contains threat level information
type ThreatLevelResponse struct {
	Score          float64               `json:"score"`
	Level          string                `json:"level"`
	ActiveSignals  []ThreatSignalInfo    `json:"active_signals"`
	Recommendation string                `json:"recommendation"`
	Timestamp      time.Time             `json:"timestamp"`
}

// ThreatSignalInfo contains signal information for API response
type ThreatSignalInfo struct {
	Type       string    `json:"type"`
	Score      float64   `json:"score"`
	Confidence float64   `json:"confidence"`
	Source     string    `json:"source"`
	Details    string    `json:"details"`
	Timestamp  time.Time `json:"timestamp"`
}

// handleThreatLevel handles threat level queries
func (s *APIServer) handleThreatLevel(ctx context.Context, req *APIRequest) *APIResponse {
	if threatDetector == nil {
		return &APIResponse{
			Result: ThreatLevelResponse{
				Score:          0,
				Level:          "unknown",
				ActiveSignals:  []ThreatSignalInfo{},
				Recommendation: "threat_detector_not_initialized",
				Timestamp:      time.Now(),
			},
			ID: req.ID,
		}
	}

	level := threatDetector.GetThreatLevel()

	signals := make([]ThreatSignalInfo, 0, len(level.ActiveSignals))
	for _, sig := range level.ActiveSignals {
		signals = append(signals, ThreatSignalInfo{
			Type:       string(sig.Type),
			Score:      sig.Score,
			Confidence: sig.Confidence,
			Source:     sig.Source,
			Details:    sig.Details,
			Timestamp:  sig.Timestamp,
		})
	}

	response := ThreatLevelResponse{
		Score:          level.Score,
		Level:          level.Level,
		ActiveSignals:  signals,
		Recommendation: level.Recommendation,
		Timestamp:      level.Timestamp,
	}

	return &APIResponse{
		Result: response,
		ID:     req.ID,
	}
}

// handleThreatSignal handles reporting external threat signals
func (s *APIServer) handleThreatSignal(ctx context.Context, req *APIRequest) *APIResponse {
	if threatDetector == nil {
		return &APIResponse{
			Error: "threat detector not initialized",
			ID:    req.ID,
		}
	}

	var params struct {
		Type       string  `json:"type"`
		Confidence float64 `json:"confidence"`
		Source     string  `json:"source"`
		Details    string  `json:"details"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return &APIResponse{
			Error: fmt.Sprintf("invalid params: %v", err),
			ID:    req.ID,
		}
	}

	if params.Type == "" {
		return &APIResponse{
			Error: "signal type is required",
			ID:    req.ID,
		}
	}

	if params.Confidence <= 0 {
		params.Confidence = 1.0
	}
	if params.Source == "" {
		params.Source = "external"
	}

	// Record the signal
	threatDetector.RecordExternalSignal(
		threat.SignalType(params.Type),
		params.Confidence,
		params.Source,
		params.Details,
	)

	return &APIResponse{
		Result: map[string]interface{}{
			"status":      "recorded",
			"signal_type": params.Type,
			"timestamp":   time.Now(),
		},
		ID: req.ID,
	}
}

// handleThreatClear handles clearing threat signals
func (s *APIServer) handleThreatClear(ctx context.Context, req *APIRequest) *APIResponse {
	if threatDetector == nil {
		return &APIResponse{
			Error: "threat detector not initialized",
			ID:    req.ID,
		}
	}

	var params struct {
		Type        string `json:"type,omitempty"`
		ContainerID string `json:"container_id,omitempty"`
		All         bool   `json:"all,omitempty"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return &APIResponse{
			Error: fmt.Sprintf("invalid params: %v", err),
			ID:    req.ID,
		}
	}

	if params.All {
		threatDetector.ClearAllSignals()
		return &APIResponse{
			Result: map[string]interface{}{
				"status":  "cleared",
				"message": "all signals cleared",
			},
			ID: req.ID,
		}
	}

	if params.Type != "" {
		threatDetector.ClearSignal(threat.SignalType(params.Type), params.ContainerID)
		return &APIResponse{
			Result: map[string]interface{}{
				"status":      "cleared",
				"signal_type": params.Type,
			},
			ID: req.ID,
		}
	}

	return &APIResponse{
		Error: "specify type or all=true",
		ID:    req.ID,
	}
}
