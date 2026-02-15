package api

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/moltbunker/moltbunker/internal/logging"
)

// Admin request/response types

type AdminMeResponse struct {
	IsAdmin bool   `json:"is_admin"`
	Wallet  string `json:"wallet"`
}

type AdminNodeResponse struct {
	NodeID      string    `json:"node_id"`
	Badges      []string  `json:"badges,omitempty"`
	Blocked     bool      `json:"blocked"`
	BlockReason string    `json:"block_reason,omitempty"`
	Notes       string    `json:"notes,omitempty"`
	UpdatedAt   time.Time `json:"updated_at"`
	UpdatedBy   string    `json:"updated_by,omitempty"`
}

type AdminNodeUpdateRequest struct {
	Badges      []string `json:"badges,omitempty"`
	Blocked     *bool    `json:"blocked,omitempty"`
	BlockReason string   `json:"block_reason,omitempty"`
	Notes       string   `json:"notes,omitempty"`
}

// requireJSON checks Content-Type header on write requests
func requireJSON(r *http.Request) bool {
	ct := r.Header.Get("Content-Type")
	return strings.HasPrefix(ct, "application/json")
}

// handleAdminMe handles GET /v1/admin/me
func (s *Server) handleAdminMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		s.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	wallet, ok := s.authenticateAdmin(r)
	if !ok {
		s.writeError(w, http.StatusForbidden, "not an admin")
		return
	}

	s.writeJSON(w, http.StatusOK, AdminMeResponse{
		IsAdmin: true,
		Wallet:  wallet,
	})
}

// handleAdminNodes handles GET /v1/admin/nodes
func (s *Server) handleAdminNodes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		s.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	if s.adminStore == nil {
		s.writeError(w, http.StatusServiceUnavailable, "admin store not initialized")
		return
	}

	all := s.adminStore.GetAll()
	result := make([]AdminNodeResponse, 0, len(all))
	for _, m := range all {
		result = append(result, AdminNodeResponse{
			NodeID:      m.NodeID,
			Badges:      m.Badges,
			Blocked:     m.Blocked,
			BlockReason: m.BlockReason,
			Notes:       m.Notes,
			UpdatedAt:   m.UpdatedAt,
			UpdatedBy:   m.UpdatedBy,
		})
	}

	s.writeJSON(w, http.StatusOK, result)
}

// handleAdminNodeByID handles GET/PUT/DELETE /v1/admin/nodes/{id}
func (s *Server) handleAdminNodeByID(w http.ResponseWriter, r *http.Request) {
	if s.adminStore == nil {
		s.writeError(w, http.StatusServiceUnavailable, "admin store not initialized")
		return
	}

	// Extract and validate node ID from path
	nodeID := strings.TrimPrefix(r.URL.Path, "/v1/admin/nodes/")
	if nodeID == "" {
		s.writeError(w, http.StatusBadRequest, "node ID required")
		return
	}
	if err := ValidateNodeID(nodeID); err != nil {
		s.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	switch r.Method {
	case http.MethodGet:
		meta := s.adminStore.Get(nodeID)
		if meta == nil {
			s.writeError(w, http.StatusNotFound, "no admin metadata for this node")
			return
		}
		s.writeJSON(w, http.StatusOK, AdminNodeResponse{
			NodeID:      meta.NodeID,
			Badges:      meta.Badges,
			Blocked:     meta.Blocked,
			BlockReason: meta.BlockReason,
			Notes:       meta.Notes,
			UpdatedAt:   meta.UpdatedAt,
			UpdatedBy:   meta.UpdatedBy,
		})

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

		body, err := io.ReadAll(io.LimitReader(r.Body, 1<<16)) // 64KB limit for node metadata
		if err != nil {
			s.writeError(w, http.StatusBadRequest, "failed to read body")
			return
		}

		var req AdminNodeUpdateRequest
		if err := json.Unmarshal(body, &req); err != nil {
			s.writeError(w, http.StatusBadRequest, "invalid JSON")
			return
		}

		// Validate badges
		if req.Badges != nil {
			if err := ValidateBadges(req.Badges); err != nil {
				s.writeError(w, http.StatusBadRequest, err.Error())
				return
			}
		}

		// Validate string lengths
		if len(req.BlockReason) > maxBlockReasonLen {
			s.writeError(w, http.StatusBadRequest, "block_reason too long (max 1024 chars)")
			return
		}
		if len(req.Notes) > maxNotesLen {
			s.writeError(w, http.StatusBadRequest, "notes too long (max 4096 chars)")
			return
		}

		// Get existing or create new
		meta := s.adminStore.Get(nodeID)
		if meta == nil {
			meta = &AdminNodeMetadata{NodeID: nodeID}
		}

		// Apply updates
		if req.Badges != nil {
			meta.Badges = req.Badges
		}
		if req.Blocked != nil {
			meta.Blocked = *req.Blocked
			// Clear block reason when unblocking
			if !*req.Blocked {
				meta.BlockReason = ""
			}
		}
		if req.BlockReason != "" {
			meta.BlockReason = req.BlockReason
		}
		if req.Notes != "" {
			meta.Notes = req.Notes
		}
		meta.UpdatedBy = wallet

		if err := s.adminStore.Set(meta); err != nil {
			logging.Error("failed to save admin metadata",
				"node_id", nodeID,
				"error", err.Error(),
				logging.Component("api"))
			s.writeError(w, http.StatusInternalServerError, "failed to save")
			return
		}

		logging.Info("admin metadata updated",
			"node_id", nodeID,
			"admin_wallet", wallet,
			"badges", meta.Badges,
			"blocked", meta.Blocked,
			logging.Component("api"))

		s.writeJSON(w, http.StatusOK, AdminNodeResponse{
			NodeID:      meta.NodeID,
			Badges:      meta.Badges,
			Blocked:     meta.Blocked,
			BlockReason: meta.BlockReason,
			Notes:       meta.Notes,
			UpdatedAt:   meta.UpdatedAt,
			UpdatedBy:   meta.UpdatedBy,
		})

	case http.MethodDelete:
		_, ok := s.authenticateAdmin(r)
		if !ok {
			s.writeError(w, http.StatusForbidden, "not an admin")
			return
		}

		logging.Info("admin metadata deleted",
			"node_id", nodeID,
			logging.Component("api"))

		if err := s.adminStore.Delete(nodeID); err != nil {
			s.writeError(w, http.StatusInternalServerError, "failed to delete")
			return
		}
		w.WriteHeader(http.StatusNoContent)

	default:
		w.Header().Set("Allow", "GET, PUT, DELETE")
		s.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// handleAdminPolicies handles GET/PUT /v1/admin/policies
func (s *Server) handleAdminPolicies(w http.ResponseWriter, r *http.Request) {
	if s.policyStore == nil {
		s.writeError(w, http.StatusServiceUnavailable, "policy store not initialized")
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.writeJSON(w, http.StatusOK, s.policyStore.Get())

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

		body, err := io.ReadAll(io.LimitReader(r.Body, 1<<16)) // 64KB limit
		if err != nil {
			s.writeError(w, http.StatusBadRequest, "failed to read body")
			return
		}

		var policies NetworkPolicies
		if err := json.Unmarshal(body, &policies); err != nil {
			s.writeError(w, http.StatusBadRequest, "invalid JSON")
			return
		}

		// Validate policy values
		if err := policies.Validate(); err != nil {
			s.writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		policies.UpdatedBy = wallet
		if err := s.policyStore.Update(&policies); err != nil {
			logging.Error("failed to save admin policies",
				"error", err.Error(),
				logging.Component("api"))
			s.writeError(w, http.StatusInternalServerError, "failed to save")
			return
		}

		logging.Info("admin policies updated",
			"admin_wallet", wallet,
			logging.Component("api"))

		s.writeJSON(w, http.StatusOK, s.policyStore.Get())

	default:
		w.Header().Set("Allow", "GET, PUT")
		s.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}
