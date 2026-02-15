package api

import (
	"encoding/json"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/moltbunker/moltbunker/internal/logging"
	"github.com/moltbunker/moltbunker/internal/metrics"
)

// ─── Request / Response Types ────────────────────────────────────────────────

// GatewayKeyResponse is the API-facing representation of an API key (no secret).
type GatewayKeyResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	KeyPrefix   string    `json:"key_prefix"`
	Permissions []string  `json:"permissions"`
	Enabled     bool      `json:"enabled"`
	RateLimit   int       `json:"rate_limit,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   time.Time `json:"expires_at,omitempty"`
	LastUsedAt  time.Time `json:"last_used_at,omitempty"`
}

// GatewayKeyCreateRequest is the request body for creating a new API key.
type GatewayKeyCreateRequest struct {
	Name          string   `json:"name"`
	Permissions   []string `json:"permissions"`
	ExpiresInDays int      `json:"expires_in_days,omitempty"`
	RateLimit     int      `json:"rate_limit,omitempty"`
}

// GatewayKeyCreateResponse includes the plain-text key (shown only once).
type GatewayKeyCreateResponse struct {
	Key      string             `json:"key"`
	Metadata GatewayKeyResponse `json:"metadata"`
}

// GatewayKeyUpdateRequest allows partial updates to a key.
type GatewayKeyUpdateRequest struct {
	Name        *string  `json:"name,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
	RateLimit   *int     `json:"rate_limit,omitempty"`
	Enabled     *bool    `json:"enabled,omitempty"`
}

// GatewayMetricsResponse is the combined metrics for the API gateway.
type GatewayMetricsResponse struct {
	Uptime            string                        `json:"uptime"`
	UptimeSeconds     float64                       `json:"uptime_seconds"`
	TotalKeys         int                           `json:"total_keys"`
	ActiveKeys        int                           `json:"active_keys"`
	RevokedKeys       int                           `json:"revoked_keys"`
	RequestCounts     map[string]uint64             `json:"request_counts"`
	RequestLatencies  map[string]metrics.LatencyStats `json:"request_latencies"`
	ActiveConnections int64                         `json:"active_connections"`
	RateLimitConfig   GatewayRateLimitInfo          `json:"rate_limit_config"`
}

// GatewayRateLimitInfo exposes the global rate limit settings.
type GatewayRateLimitInfo struct {
	RequestsPerMinute int `json:"requests_per_minute"`
	BurstLimit        int `json:"burst_limit"`
}

// ─── Validation ──────────────────────────────────────────────────────────────

var gatewayKeyNamePattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]{0,99}$`)

var validPermissions = map[string]bool{
	"read":  true,
	"write": true,
	"admin": true,
}

const maxGatewayKeys = 100

func validateKeyName(name string) bool {
	return gatewayKeyNamePattern.MatchString(name)
}

func validatePermissions(perms []string) bool {
	if len(perms) == 0 {
		return false
	}
	for _, p := range perms {
		if !validPermissions[p] {
			return false
		}
	}
	return true
}

// ─── Handlers ────────────────────────────────────────────────────────────────

// handleGatewayKeys handles GET (list) and POST (create) on /v1/admin/gateway/keys
func (s *Server) handleGatewayKeys(w http.ResponseWriter, r *http.Request) {
	if s.apiKeyManager == nil {
		s.writeError(w, http.StatusServiceUnavailable, "API key manager not initialized")
		return
	}

	switch r.Method {
	case http.MethodGet:
		keys := s.apiKeyManager.ListKeys()
		resp := make([]GatewayKeyResponse, 0, len(keys))
		for _, k := range keys {
			resp = append(resp, toGatewayKeyResponse(k))
		}
		s.writeJSON(w, http.StatusOK, resp)

	case http.MethodPost:
		if !requireJSON(r) {
			s.writeError(w, http.StatusUnsupportedMediaType, "Content-Type must be application/json")
			return
		}

		wallet, ok := s.authenticateAdmin(r)
		if !ok {
			s.writeError(w, http.StatusForbidden, "not an admin")
			return
		}

		// Check key limit
		total, _, _ := s.apiKeyManager.CountByStatus()
		if total >= maxGatewayKeys {
			s.writeError(w, http.StatusConflict, "maximum number of API keys reached (100)")
			return
		}

		body, err := io.ReadAll(io.LimitReader(r.Body, 1<<16))
		if err != nil {
			s.writeError(w, http.StatusBadRequest, "failed to read body")
			return
		}

		var req GatewayKeyCreateRequest
		if err := json.Unmarshal(body, &req); err != nil {
			s.writeError(w, http.StatusBadRequest, "invalid JSON")
			return
		}

		// Validate name
		if !validateKeyName(req.Name) {
			s.writeError(w, http.StatusBadRequest, "invalid name: must be 1-100 alphanumeric characters, hyphens, or underscores")
			return
		}

		// Validate permissions
		if !validatePermissions(req.Permissions) {
			s.writeError(w, http.StatusBadRequest, "invalid permissions: allowed values are read, write, admin")
			return
		}

		// Validate rate limit
		if req.RateLimit < 0 || req.RateLimit > 10000 {
			s.writeError(w, http.StatusBadRequest, "rate_limit must be 0-10000 (0 = use global default)")
			return
		}

		// Validate expiry
		if req.ExpiresInDays < 0 || req.ExpiresInDays > 365 {
			s.writeError(w, http.StatusBadRequest, "expires_in_days must be 0-365 (0 = never)")
			return
		}

		key, plainKey, err := s.apiKeyManager.CreateKey(req.Name, req.Permissions, req.ExpiresInDays)
		if err != nil {
			s.writeError(w, http.StatusInternalServerError, "failed to create key")
			return
		}

		// Set custom rate limit if provided
		if req.RateLimit > 0 {
			rl := req.RateLimit
			s.apiKeyManager.UpdateKey(key.ID, nil, &rl, nil)
		}

		logging.Info("gateway key created via admin API",
			"key_id", key.ID,
			"admin_wallet", wallet,
			logging.Component("api"))

		s.writeJSON(w, http.StatusCreated, GatewayKeyCreateResponse{
			Key:      plainKey,
			Metadata: toGatewayKeyResponse(key),
		})

	default:
		w.Header().Set("Allow", "GET, POST")
		s.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// handleGatewayKeyByID handles single-key operations on /v1/admin/gateway/keys/{id}[/revoke]
func (s *Server) handleGatewayKeyByID(w http.ResponseWriter, r *http.Request) {
	if s.apiKeyManager == nil {
		s.writeError(w, http.StatusServiceUnavailable, "API key manager not initialized")
		return
	}

	// Parse: /v1/admin/gateway/keys/{id} or /v1/admin/gateway/keys/{id}/revoke
	remainder := strings.TrimPrefix(r.URL.Path, "/v1/admin/gateway/keys/")
	if remainder == "" {
		s.writeError(w, http.StatusBadRequest, "key ID required")
		return
	}

	var keyID string
	var isRevoke bool

	if strings.HasSuffix(remainder, "/revoke") {
		keyID = strings.TrimSuffix(remainder, "/revoke")
		isRevoke = true
	} else {
		keyID = remainder
	}

	// Handle revoke sub-route
	if isRevoke {
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", "POST")
			s.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		wallet, ok := s.authenticateAdmin(r)
		if !ok {
			s.writeError(w, http.StatusForbidden, "not an admin")
			return
		}

		if err := s.apiKeyManager.RevokeKey(keyID); err != nil {
			if strings.Contains(err.Error(), "not found") {
				s.writeError(w, http.StatusNotFound, "key not found")
			} else {
				s.writeError(w, http.StatusInternalServerError, "failed to revoke key")
			}
			return
		}

		logging.Info("gateway key revoked via admin API",
			"key_id", keyID,
			"admin_wallet", wallet,
			logging.Component("api"))

		key, _ := s.apiKeyManager.GetKeyByID(keyID)
		s.writeJSON(w, http.StatusOK, toGatewayKeyResponse(key))
		return
	}

	// Standard CRUD by ID
	switch r.Method {
	case http.MethodGet:
		key, exists := s.apiKeyManager.GetKeyByID(keyID)
		if !exists {
			s.writeError(w, http.StatusNotFound, "key not found")
			return
		}
		s.writeJSON(w, http.StatusOK, toGatewayKeyResponse(key))

	case http.MethodPut:
		if !requireJSON(r) {
			s.writeError(w, http.StatusUnsupportedMediaType, "Content-Type must be application/json")
			return
		}

		wallet, ok := s.authenticateAdmin(r)
		if !ok {
			s.writeError(w, http.StatusForbidden, "not an admin")
			return
		}

		// Verify key exists
		if _, exists := s.apiKeyManager.GetKeyByID(keyID); !exists {
			s.writeError(w, http.StatusNotFound, "key not found")
			return
		}

		body, err := io.ReadAll(io.LimitReader(r.Body, 1<<16))
		if err != nil {
			s.writeError(w, http.StatusBadRequest, "failed to read body")
			return
		}

		var req GatewayKeyUpdateRequest
		if err := json.Unmarshal(body, &req); err != nil {
			s.writeError(w, http.StatusBadRequest, "invalid JSON")
			return
		}

		// Validate name if provided
		if req.Name != nil && !validateKeyName(*req.Name) {
			s.writeError(w, http.StatusBadRequest, "invalid name: must be 1-100 alphanumeric characters, hyphens, or underscores")
			return
		}

		// Validate permissions if provided
		if req.Permissions != nil && !validatePermissions(req.Permissions) {
			s.writeError(w, http.StatusBadRequest, "invalid permissions: allowed values are read, write, admin")
			return
		}

		// Validate rate limit if provided
		if req.RateLimit != nil && (*req.RateLimit < 0 || *req.RateLimit > 10000) {
			s.writeError(w, http.StatusBadRequest, "rate_limit must be 0-10000")
			return
		}

		// Update mutable fields
		if err := s.apiKeyManager.UpdateKey(keyID, req.Name, req.RateLimit, req.Enabled); err != nil {
			s.writeError(w, http.StatusInternalServerError, "failed to update key")
			return
		}

		// Update permissions separately if provided
		if req.Permissions != nil {
			if err := s.apiKeyManager.UpdateKeyPermissions(keyID, req.Permissions); err != nil {
				s.writeError(w, http.StatusInternalServerError, "failed to update permissions")
				return
			}
		}

		logging.Info("gateway key updated via admin API",
			"key_id", keyID,
			"admin_wallet", wallet,
			logging.Component("api"))

		key, _ := s.apiKeyManager.GetKeyByID(keyID)
		s.writeJSON(w, http.StatusOK, toGatewayKeyResponse(key))

	case http.MethodDelete:
		wallet, ok := s.authenticateAdmin(r)
		if !ok {
			s.writeError(w, http.StatusForbidden, "not an admin")
			return
		}

		if err := s.apiKeyManager.DeleteKey(keyID); err != nil {
			if strings.Contains(err.Error(), "not found") {
				s.writeError(w, http.StatusNotFound, "key not found")
			} else {
				s.writeError(w, http.StatusInternalServerError, "failed to delete key")
			}
			return
		}

		logging.Info("gateway key deleted via admin API",
			"key_id", keyID,
			"admin_wallet", wallet,
			logging.Component("api"))

		w.WriteHeader(http.StatusNoContent)

	default:
		w.Header().Set("Allow", "GET, PUT, DELETE")
		s.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// handleGatewayMetrics handles GET /v1/admin/gateway/metrics
func (s *Server) handleGatewayMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		s.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var m *metrics.Metrics
	if s.metricsCollector != nil {
		m = s.metricsCollector.GetMetrics()
	}

	total, active, revoked := 0, 0, 0
	if s.apiKeyManager != nil {
		total, active, revoked = s.apiKeyManager.CountByStatus()
	}

	resp := GatewayMetricsResponse{
		TotalKeys:   total,
		ActiveKeys:  active,
		RevokedKeys: revoked,
		RateLimitConfig: GatewayRateLimitInfo{
			RequestsPerMinute: s.config.RateLimit,
			BurstLimit:        s.config.RateLimitBurst,
		},
	}

	if m != nil {
		resp.Uptime = m.Uptime
		resp.UptimeSeconds = m.UptimeSeconds
		resp.RequestCounts = m.RequestCounts
		resp.RequestLatencies = m.RequestLatencies
		resp.ActiveConnections = m.ActiveConnections
	}

	s.writeJSON(w, http.StatusOK, resp)
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func toGatewayKeyResponse(k *APIKey) GatewayKeyResponse {
	return GatewayKeyResponse{
		ID:          k.ID,
		Name:        k.Name,
		KeyPrefix:   k.KeyPrefix,
		Permissions: k.Permissions,
		Enabled:     k.Enabled,
		RateLimit:   k.RateLimit,
		CreatedAt:   k.CreatedAt,
		ExpiresAt:   k.ExpiresAt,
		LastUsedAt:  k.LastUsedAt,
	}
}
