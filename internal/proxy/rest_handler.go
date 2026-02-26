package proxy

import (
	"encoding/json"
	"net/http"
	"strings"
)

// RESTHandler serves the JSON REST API for proxy management under /v1/proxy/.
type RESTHandler struct {
	server *Server
}

// NewRESTHandler creates a new REST handler for the proxy service.
func NewRESTHandler(server *Server) *RESTHandler {
	return &RESTHandler{server: server}
}

// RegisterRoutes registers proxy REST routes on the given mux.
func (h *RESTHandler) RegisterRoutes(mux *http.ServeMux, wrapRead, wrapWrite func(http.HandlerFunc) http.HandlerFunc) {
	mux.HandleFunc("/v1/proxy/sessions", wrapRead(h.handleSessions))
	mux.HandleFunc("/v1/proxy/sessions/", wrapWrite(h.handleSessionByID))
	mux.HandleFunc("/v1/proxy/usage", wrapRead(h.handleUsage))
	mux.HandleFunc("/v1/proxy/status", wrapRead(h.handleStatus))
}

func (h *RESTHandler) handleSessions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		writeProxyError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	sessions := h.server.Tracker().List()
	writeProxyJSON(w, http.StatusOK, map[string]any{
		"sessions": sessions,
		"count":    len(sessions),
	})
}

func (h *RESTHandler) handleSessionByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/v1/proxy/sessions/")
	if id == "" {
		writeProxyError(w, http.StatusBadRequest, "session ID required")
		return
	}

	switch r.Method {
	case http.MethodGet:
		session, ok := h.server.Tracker().Get(id)
		if !ok {
			writeProxyError(w, http.StatusNotFound, "session not found")
			return
		}
		writeProxyJSON(w, http.StatusOK, session)

	case http.MethodDelete:
		h.server.Tracker().Remove(id)
		w.WriteHeader(http.StatusNoContent)

	default:
		w.Header().Set("Allow", "GET, DELETE")
		writeProxyError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *RESTHandler) handleUsage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		writeProxyError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	wallet := r.Header.Get("X-Wallet-Address")
	if wallet == "" {
		wallet = "api-key-user"
	}
	wallet = strings.ToLower(wallet)

	report := h.server.Usage(wallet)
	writeProxyJSON(w, http.StatusOK, report)
}

func (h *RESTHandler) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		writeProxyError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	writeProxyJSON(w, http.StatusOK, map[string]any{
		"running":        h.server.IsRunning(),
		"socks5_addr":    h.server.cfg.SOCKS5Addr,
		"http_addr":      h.server.cfg.HTTPAddr,
		"use_tor":        h.server.cfg.UseTor,
		"active_sessions": h.server.Tracker().Count(),
		"max_sessions":   h.server.cfg.MaxSessions,
	})
}

func writeProxyJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeProxyError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
