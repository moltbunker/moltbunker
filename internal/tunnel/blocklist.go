package tunnel

import (
	"encoding/json"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

// blocklist.go implements the operator takedown kill-switch (EDGE-02): the
// daemon side of the abuse-takedown flow. An operator can block a subdomain (or
// custom host) at ingress so the reverse tunnel server refuses to register it
// and refuses to open streams for it. This is the enforcement primitive that a
// higher-level abuse/legal workflow (LEGAL-01) drives.

// BlocklistChecker reports whether a host/subdomain is blocked at ingress.
// ReverseServer consults it on both registration and stream-open so a takedown
// takes effect immediately for new sessions and is enforced on every request
// for live sessions.
type BlocklistChecker interface {
	// IsBlocked reports whether the given subdomain/host is on the blocklist,
	// and the operator-supplied reason when it is.
	IsBlocked(subdomain string) (blocked bool, reason string)
}

// BlockEntry is a single blocklist record.
type BlockEntry struct {
	Subdomain string    `json:"subdomain"`
	Reason    string    `json:"reason,omitempty"`
	BlockedAt time.Time `json:"blocked_at"`
}

// Blocklist is a thread-safe in-memory set of blocked subdomains/hosts. It is
// the default BlocklistChecker. Persistence across restart is a documented
// follow-up; the abuse workflow re-applies blocks on startup.
type Blocklist struct {
	mu      sync.RWMutex
	entries map[string]BlockEntry // normalized subdomain -> entry
}

// NewBlocklist creates an empty blocklist.
func NewBlocklist() *Blocklist {
	return &Blocklist{entries: make(map[string]BlockEntry)}
}

// normalizeSub lowercases and trims a subdomain/host for stable matching.
func normalizeSub(s string) string {
	return strings.TrimSuffix(strings.ToLower(strings.TrimSpace(s)), ".")
}

// Block adds a subdomain/host to the blocklist with an optional reason.
func (b *Blocklist) Block(subdomain, reason string) {
	sub := normalizeSub(subdomain)
	if sub == "" {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.entries[sub] = BlockEntry{Subdomain: sub, Reason: reason, BlockedAt: time.Now()}
}

// Unblock removes a subdomain/host from the blocklist. Reports whether an entry
// was present.
func (b *Blocklist) Unblock(subdomain string) bool {
	sub := normalizeSub(subdomain)
	b.mu.Lock()
	defer b.mu.Unlock()
	_, ok := b.entries[sub]
	delete(b.entries, sub)
	return ok
}

// IsBlocked implements BlocklistChecker.
func (b *Blocklist) IsBlocked(subdomain string) (bool, string) {
	sub := normalizeSub(subdomain)
	b.mu.RLock()
	defer b.mu.RUnlock()
	e, ok := b.entries[sub]
	if !ok {
		return false, ""
	}
	return true, e.Reason
}

// List returns all blocklist entries sorted by subdomain (stable output).
func (b *Blocklist) List() []BlockEntry {
	b.mu.RLock()
	out := make([]BlockEntry, 0, len(b.entries))
	for _, e := range b.entries {
		out = append(out, e)
	}
	b.mu.RUnlock()
	sort.Slice(out, func(i, j int) bool { return out[i].Subdomain < out[j].Subdomain })
	return out
}

// Len returns the number of blocked entries.
func (b *Blocklist) Len() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.entries)
}

// Compile-time assertion.
var _ BlocklistChecker = (*Blocklist)(nil)

// --- Admin HTTP surface ---------------------------------------------------

const maxBlocklistBody = 4 << 10 // 4 KiB

// blockRequest is the body for POST (block) requests.
type blockRequest struct {
	Subdomain string `json:"subdomain"`
	Reason    string `json:"reason,omitempty"`
}

// BlocklistAdminHandler is a small admin API to manage the takedown blocklist.
// It carries no auth of its own; the daemon mounts it behind the existing
// API-key/admin middleware. Routes (under a mount prefix):
//
//	GET    .../blocklist            -> [{subdomain, reason, blocked_at}, ...]
//	POST   .../blocklist  {subdomain, reason}  -> 201
//	DELETE .../blocklist?subdomain=<host>      -> 204 / 404
type BlocklistAdminHandler struct {
	bl *Blocklist
}

// NewBlocklistAdminHandler builds the admin handler around a blocklist.
func NewBlocklistAdminHandler(bl *Blocklist) *BlocklistAdminHandler {
	return &BlocklistAdminHandler{bl: bl}
}

// ServeHTTP dispatches by method.
func (h *BlocklistAdminHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.writeJSON(w, http.StatusOK, h.bl.List())
	case http.MethodPost:
		var req blockRequest
		body, err := io.ReadAll(io.LimitReader(r.Body, maxBlocklistBody))
		if err != nil {
			h.writeErr(w, http.StatusBadRequest, "failed to read body")
			return
		}
		if err := json.Unmarshal(body, &req); err != nil {
			h.writeErr(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		if strings.TrimSpace(req.Subdomain) == "" {
			h.writeErr(w, http.StatusBadRequest, "subdomain is required")
			return
		}
		h.bl.Block(req.Subdomain, req.Reason)
		h.writeJSON(w, http.StatusCreated, BlockEntry{
			Subdomain: normalizeSub(req.Subdomain),
			Reason:    req.Reason,
			BlockedAt: time.Now(),
		})
	case http.MethodDelete:
		sub := r.URL.Query().Get("subdomain")
		if strings.TrimSpace(sub) == "" {
			h.writeErr(w, http.StatusBadRequest, "subdomain query param is required")
			return
		}
		if !h.bl.Unblock(sub) {
			h.writeErr(w, http.StatusNotFound, "subdomain not blocked")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		w.Header().Set("Allow", "GET, POST, DELETE")
		h.writeErr(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *BlocklistAdminHandler) writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func (h *BlocklistAdminHandler) writeErr(w http.ResponseWriter, status int, msg string) {
	h.writeJSON(w, status, map[string]string{"error": msg})
}
