package agent

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// RESTHandler provides HTTP endpoints for agent management.
type RESTHandler struct {
	runtime *AgentRuntime
	memory  *MemoryStore
}

// NewRESTHandler creates a new agent REST handler.
func NewRESTHandler(runtime *AgentRuntime, memory *MemoryStore) *RESTHandler {
	return &RESTHandler{
		runtime: runtime,
		memory:  memory,
	}
}

// RegisterRoutes registers agent routes on a mux.
func (h *RESTHandler) RegisterRoutes(mux *http.ServeMux, wrapRead, wrapWrite func(http.HandlerFunc) http.HandlerFunc) {
	mux.HandleFunc("/v1/agents", wrapWrite(h.handleAgents))
	mux.HandleFunc("/v1/agents/", wrapWrite(h.handleAgent))
}

func (h *RESTHandler) handleAgents(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listAgents(w, r)
	case http.MethodPost:
		h.deployAgent(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *RESTHandler) handleAgent(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/v1/agents/")
	parts := strings.SplitN(path, "/", 2)
	agentID := parts[0]

	if len(parts) == 2 {
		switch parts[1] {
		case "invoke":
			h.invokeAgent(w, r, agentID)
		case "memory":
			h.handleMemory(w, r, agentID)
		case "stop":
			h.stopAgent(w, r, agentID)
		default:
			http.NotFound(w, r)
		}
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.getAgent(w, r, agentID)
	case http.MethodDelete:
		h.deleteAgent(w, r, agentID)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// extractAgentWallet returns the verified wallet from the auth middleware header.
func extractAgentWallet(r *http.Request) string {
	return r.Header.Get("X-Moltbunker-Verified-Wallet")
}

func (h *RESTHandler) listAgents(w http.ResponseWriter, r *http.Request) {
	wallet := extractAgentWallet(r)
	if wallet == "" {
		writeError(w, http.StatusForbidden, "no verified identity")
		return
	}
	agents := h.runtime.List(wallet)
	writeJSON(w, http.StatusOK, map[string]interface{}{"agents": agents})
}

func (h *RESTHandler) deployAgent(w http.ResponseWriter, r *http.Request) {
	wallet := extractAgentWallet(r)
	if wallet == "" {
		writeError(w, http.StatusForbidden, "no verified identity")
		return
	}

	var spec AgentSpec
	if err := json.NewDecoder(r.Body).Decode(&spec); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	// Set owner from verified identity, not from client input
	spec.Owner = wallet

	deployment, err := h.runtime.Deploy(r.Context(), spec)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, deployment)
}

func (h *RESTHandler) getAgent(w http.ResponseWriter, r *http.Request, agentID string) {
	agent, ok := h.runtime.Get(agentID)
	if !ok {
		writeError(w, http.StatusNotFound, fmt.Sprintf("agent %q not found", agentID))
		return
	}
	wallet := extractAgentWallet(r)
	if wallet != "" && agent.Spec.Owner != wallet {
		writeError(w, http.StatusNotFound, fmt.Sprintf("agent %q not found", agentID))
		return
	}
	writeJSON(w, http.StatusOK, agent)
}

func (h *RESTHandler) stopAgent(w http.ResponseWriter, r *http.Request, agentID string) {
	agent, ok := h.runtime.Get(agentID)
	if !ok {
		writeError(w, http.StatusNotFound, fmt.Sprintf("agent %q not found", agentID))
		return
	}
	wallet := extractAgentWallet(r)
	if wallet != "" && agent.Spec.Owner != wallet {
		writeError(w, http.StatusNotFound, fmt.Sprintf("agent %q not found", agentID))
		return
	}
	if err := h.runtime.Stop(r.Context(), agentID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *RESTHandler) deleteAgent(w http.ResponseWriter, r *http.Request, agentID string) {
	agent, ok := h.runtime.Get(agentID)
	if !ok {
		writeError(w, http.StatusNotFound, fmt.Sprintf("agent %q not found", agentID))
		return
	}
	wallet := extractAgentWallet(r)
	if wallet != "" && agent.Spec.Owner != wallet {
		writeError(w, http.StatusNotFound, fmt.Sprintf("agent %q not found", agentID))
		return
	}
	if err := h.runtime.Delete(agentID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *RESTHandler) invokeAgent(w http.ResponseWriter, r *http.Request, agentID string) {
	agent, ok := h.runtime.Get(agentID)
	if !ok {
		writeError(w, http.StatusNotFound, fmt.Sprintf("agent %q not found", agentID))
		return
	}
	wallet := extractAgentWallet(r)
	if wallet != "" && agent.Spec.Owner != wallet {
		writeError(w, http.StatusNotFound, fmt.Sprintf("agent %q not found", agentID))
		return
	}
	if agent.Status != AgentStatusRunning {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("agent is %s, not running", agent.Status))
		return
	}

	var req AgentInvokeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}
	req.AgentID = agentID

	// In production, this would forward to the agent's container.
	// For now, return a placeholder response.
	resp := AgentInvokeResponse{
		AgentID:    agentID,
		Response:   fmt.Sprintf("Agent %s received: %s", agentID, req.Message),
		TokensUsed: 0,
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *RESTHandler) handleMemory(w http.ResponseWriter, r *http.Request, agentID string) {
	// Verify ownership
	agent, ok := h.runtime.Get(agentID)
	if !ok {
		writeError(w, http.StatusNotFound, fmt.Sprintf("agent %q not found", agentID))
		return
	}
	wallet := extractAgentWallet(r)
	if wallet != "" && agent.Spec.Owner != wallet {
		writeError(w, http.StatusNotFound, fmt.Sprintf("agent %q not found", agentID))
		return
	}

	switch r.Method {
	case http.MethodGet:
		key := r.URL.Query().Get("key")
		if key != "" {
			entry, err := h.memory.Get(agentID, key)
			if err != nil {
				writeError(w, http.StatusNotFound, err.Error())
				return
			}
			writeJSON(w, http.StatusOK, entry)
		} else {
			entries := h.memory.List(agentID)
			writeJSON(w, http.StatusOK, map[string]interface{}{"entries": entries})
		}

	case http.MethodPost, http.MethodPut:
		var entry MemoryEntry
		if err := json.NewDecoder(r.Body).Decode(&entry); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
			return
		}
		if err := h.memory.Put(agentID, entry.Key, entry.Value); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		w.WriteHeader(http.StatusNoContent)

	case http.MethodDelete:
		key := r.URL.Query().Get("key")
		if key == "" {
			h.memory.Clear(agentID)
		} else {
			h.memory.Delete(agentID, key)
		}
		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
