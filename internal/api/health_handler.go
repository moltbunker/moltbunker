package api

import (
	"fmt"
	"net/http"
	"time"
)

// startTime records when the server package was initialized for uptime calculation.
var startTime = time.Now()

// HealthResponse is the JSON response for the /health endpoint
type HealthResponse struct {
	Status    string `json:"status"`
	Uptime    string `json:"uptime"`
	PeerCount int64  `json:"peer_count"`
	Version   string `json:"version"`
	Reason    string `json:"reason,omitempty"`
}

// handleHealthCheck handles GET /health for load balancer health probes.
// No authentication is required.
func (s *Server) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if the server is running
	s.mu.RLock()
	running := s.running
	s.mu.RUnlock()

	if !running {
		s.writeJSON(w, http.StatusServiceUnavailable, HealthResponse{
			Status:  "unhealthy",
			Reason:  "server not running",
			Version: "1.0.0",
		})
		return
	}

	// Calculate uptime from metrics collector if available, otherwise use package-level start time
	var uptime string
	var peerCount int64

	if s.metricsCollector != nil {
		m := s.metricsCollector.GetMetrics()
		uptime = m.Uptime
		peerCount = m.PeerCount
	} else {
		uptime = fmt.Sprintf("%s", time.Since(startTime).Round(time.Second))
	}

	// Optionally check daemon bridge connectivity
	if s.daemonBridge != nil && !s.daemonBridge.IsConnected() {
		s.writeJSON(w, http.StatusServiceUnavailable, HealthResponse{
			Status:    "unhealthy",
			Uptime:    uptime,
			PeerCount: peerCount,
			Version:   "1.0.0",
			Reason:    "daemon bridge disconnected",
		})
		return
	}

	s.writeJSON(w, http.StatusOK, HealthResponse{
		Status:    "healthy",
		Uptime:    uptime,
		PeerCount: peerCount,
		Version:   "1.0.0",
	})
}
