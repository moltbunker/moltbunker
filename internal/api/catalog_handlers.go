package api

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/moltbunker/moltbunker/internal/logging"
)

// handleCatalog handles GET /v1/catalog (public, no auth)
func (s *Server) handleCatalog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		s.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	if s.catalogStore == nil {
		s.writeError(w, http.StatusServiceUnavailable, "catalog not initialized")
		return
	}

	s.writeJSON(w, http.StatusOK, s.catalogStore.GetPublic())
}

// handleAdminCatalog handles GET/PUT /v1/admin/catalog
func (s *Server) handleAdminCatalog(w http.ResponseWriter, r *http.Request) {
	if s.catalogStore == nil {
		s.writeError(w, http.StatusServiceUnavailable, "catalog not initialized")
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.writeJSON(w, http.StatusOK, s.catalogStore.Get())

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

		body, err := io.ReadAll(io.LimitReader(r.Body, 1<<16))
		if err != nil {
			s.writeError(w, http.StatusBadRequest, "failed to read body")
			return
		}

		var catalog CatalogConfig
		if err := json.Unmarshal(body, &catalog); err != nil {
			s.writeError(w, http.StatusBadRequest, "invalid JSON")
			return
		}

		catalog.UpdatedBy = wallet
		if err := s.catalogStore.Replace(&catalog); err != nil {
			s.writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		logging.Info("catalog replaced",
			"admin_wallet", wallet,
			"presets", len(catalog.Presets),
			"categories", len(catalog.Categories),
			logging.Component("api"))

		s.writeJSON(w, http.StatusOK, s.catalogStore.Get())

	default:
		w.Header().Set("Allow", "GET, PUT")
		s.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// handleAdminCatalogPresets handles POST /v1/admin/catalog/presets
func (s *Server) handleAdminCatalogPresets(w http.ResponseWriter, r *http.Request) {
	if s.catalogStore == nil {
		s.writeError(w, http.StatusServiceUnavailable, "catalog not initialized")
		return
	}

	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		s.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	if !requireJSON(r) {
		s.writeError(w, http.StatusUnsupportedMediaType, "Content-Type must be application/json")
		return
	}

	wallet, ok := s.authenticateAdmin(r)
	if !ok {
		s.writeError(w, http.StatusForbidden, "not an admin")
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<16))
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "failed to read body")
		return
	}

	var preset CatalogPreset
	if err := json.Unmarshal(body, &preset); err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	if err := s.catalogStore.AddPreset(preset); err != nil {
		s.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	logging.Info("catalog preset added",
		"preset_id", preset.ID,
		"admin_wallet", wallet,
		logging.Component("api"))

	s.writeJSON(w, http.StatusCreated, preset)
}

// handleAdminCatalogPresetByID handles PUT/DELETE /v1/admin/catalog/presets/{id}
func (s *Server) handleAdminCatalogPresetByID(w http.ResponseWriter, r *http.Request) {
	if s.catalogStore == nil {
		s.writeError(w, http.StatusServiceUnavailable, "catalog not initialized")
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/v1/admin/catalog/presets/")
	if id == "" {
		s.writeError(w, http.StatusBadRequest, "preset ID required")
		return
	}

	switch r.Method {
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

		body, err := io.ReadAll(io.LimitReader(r.Body, 1<<16))
		if err != nil {
			s.writeError(w, http.StatusBadRequest, "failed to read body")
			return
		}

		var preset CatalogPreset
		if err := json.Unmarshal(body, &preset); err != nil {
			s.writeError(w, http.StatusBadRequest, "invalid JSON")
			return
		}

		if err := s.catalogStore.UpdatePreset(id, preset); err != nil {
			if strings.Contains(err.Error(), "not found") {
				s.writeError(w, http.StatusNotFound, err.Error())
			} else {
				s.writeError(w, http.StatusBadRequest, err.Error())
			}
			return
		}

		logging.Info("catalog preset updated",
			"preset_id", id,
			"admin_wallet", wallet,
			logging.Component("api"))

		s.writeJSON(w, http.StatusOK, preset)

	case http.MethodDelete:
		wallet, ok := s.authenticateAdmin(r)
		if !ok {
			s.writeError(w, http.StatusForbidden, "not an admin")
			return
		}

		if err := s.catalogStore.DeletePreset(id); err != nil {
			if strings.Contains(err.Error(), "not found") {
				s.writeError(w, http.StatusNotFound, err.Error())
			} else {
				s.writeError(w, http.StatusBadRequest, err.Error())
			}
			return
		}

		logging.Info("catalog preset deleted",
			"preset_id", id,
			"admin_wallet", wallet,
			logging.Component("api"))

		w.WriteHeader(http.StatusNoContent)

	default:
		w.Header().Set("Allow", "PUT, DELETE")
		s.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// handleAdminCatalogCategories handles POST /v1/admin/catalog/categories
func (s *Server) handleAdminCatalogCategories(w http.ResponseWriter, r *http.Request) {
	if s.catalogStore == nil {
		s.writeError(w, http.StatusServiceUnavailable, "catalog not initialized")
		return
	}

	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		s.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	if !requireJSON(r) {
		s.writeError(w, http.StatusUnsupportedMediaType, "Content-Type must be application/json")
		return
	}

	wallet, ok := s.authenticateAdmin(r)
	if !ok {
		s.writeError(w, http.StatusForbidden, "not an admin")
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<16))
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "failed to read body")
		return
	}

	var cat CatalogCategory
	if err := json.Unmarshal(body, &cat); err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	if err := s.catalogStore.AddCategory(cat); err != nil {
		s.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	logging.Info("catalog category added",
		"category_id", cat.ID,
		"admin_wallet", wallet,
		logging.Component("api"))

	s.writeJSON(w, http.StatusCreated, cat)
}

// handleAdminCatalogCategoryByID handles PUT/DELETE /v1/admin/catalog/categories/{id}
func (s *Server) handleAdminCatalogCategoryByID(w http.ResponseWriter, r *http.Request) {
	if s.catalogStore == nil {
		s.writeError(w, http.StatusServiceUnavailable, "catalog not initialized")
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/v1/admin/catalog/categories/")
	if id == "" {
		s.writeError(w, http.StatusBadRequest, "category ID required")
		return
	}

	switch r.Method {
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

		body, err := io.ReadAll(io.LimitReader(r.Body, 1<<16))
		if err != nil {
			s.writeError(w, http.StatusBadRequest, "failed to read body")
			return
		}

		var cat CatalogCategory
		if err := json.Unmarshal(body, &cat); err != nil {
			s.writeError(w, http.StatusBadRequest, "invalid JSON")
			return
		}

		if err := s.catalogStore.UpdateCategory(id, cat); err != nil {
			if strings.Contains(err.Error(), "not found") {
				s.writeError(w, http.StatusNotFound, err.Error())
			} else {
				s.writeError(w, http.StatusBadRequest, err.Error())
			}
			return
		}

		logging.Info("catalog category updated",
			"category_id", id,
			"admin_wallet", wallet,
			logging.Component("api"))

		s.writeJSON(w, http.StatusOK, cat)

	case http.MethodDelete:
		wallet, ok := s.authenticateAdmin(r)
		if !ok {
			s.writeError(w, http.StatusForbidden, "not an admin")
			return
		}

		if err := s.catalogStore.DeleteCategory(id); err != nil {
			if strings.Contains(err.Error(), "not found") {
				s.writeError(w, http.StatusNotFound, err.Error())
			} else {
				s.writeError(w, http.StatusBadRequest, err.Error())
			}
			return
		}

		logging.Info("catalog category deleted",
			"category_id", id,
			"admin_wallet", wallet,
			logging.Component("api"))

		w.WriteHeader(http.StatusNoContent)

	default:
		w.Header().Set("Allow", "PUT, DELETE")
		s.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// handleAdminCatalogTierByID handles PUT /v1/admin/catalog/tiers/{id}
func (s *Server) handleAdminCatalogTierByID(w http.ResponseWriter, r *http.Request) {
	if s.catalogStore == nil {
		s.writeError(w, http.StatusServiceUnavailable, "catalog not initialized")
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/v1/admin/catalog/tiers/")
	if id == "" {
		s.writeError(w, http.StatusBadRequest, "tier ID required")
		return
	}

	if r.Method != http.MethodPut {
		w.Header().Set("Allow", "PUT")
		s.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	if !requireJSON(r) {
		s.writeError(w, http.StatusUnsupportedMediaType, "Content-Type must be application/json")
		return
	}

	wallet, ok := s.authenticateAdmin(r)
	if !ok {
		s.writeError(w, http.StatusForbidden, "not an admin")
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<16))
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "failed to read body")
		return
	}

	var tier CatalogTier
	if err := json.Unmarshal(body, &tier); err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	if err := s.catalogStore.UpdateTier(id, tier); err != nil {
		if strings.Contains(err.Error(), "not found") {
			s.writeError(w, http.StatusNotFound, err.Error())
		} else {
			s.writeError(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	logging.Info("catalog tier updated",
		"tier_id", id,
		"admin_wallet", wallet,
		logging.Component("api"))

	s.writeJSON(w, http.StatusOK, tier)
}
