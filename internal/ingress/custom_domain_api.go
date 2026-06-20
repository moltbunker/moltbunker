package ingress

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/moltbunker/moltbunker/internal/logging"
)

// custom_domain_api.go exposes the two-step BYO custom-hostname verification
// workflow as HTTP handlers (EDGE-02). The handlers are intentionally separate
// from the storage/verification logic so they can be unit-tested without a live
// server. They carry NO authentication of their own — the daemon mounts them
// behind the existing API-key middleware, and the verifier's HMAC-bound token
// prevents a caller from claiming a host they do not control.
//
// Routes (mounted under a prefix by the daemon, e.g. /ingress/custom-domain/):
//
//	POST .../challenge  {host, deployment_id}
//	  -> {method, token, cname_target, txt_name, txt_value}   (nothing persisted)
//	POST .../verify     {host, deployment_id, owner_wallet?}
//	  -> 200 {host, deployment_id, expires_at} on proof success, else 4xx

const (
	// maxCustomDomainBody bounds the request body for the verification API.
	maxCustomDomainBody = 4 << 10 // 4 KiB
)

// challengeRequest is the body for POST .../challenge.
type challengeRequest struct {
	Host         string `json:"host"`
	DeploymentID string `json:"deployment_id"`
}

// challengeResponse describes what DNS record the customer must publish.
type challengeResponse struct {
	Method      VerifyMethod `json:"method"`
	Token       string       `json:"token"`
	CNAMETarget string       `json:"cname_target,omitempty"`
	TXTName     string       `json:"txt_name,omitempty"`
	TXTValue    string       `json:"txt_value,omitempty"`
}

// verifyRequest is the body for POST .../verify.
type verifyRequest struct {
	Host         string `json:"host"`
	DeploymentID string `json:"deployment_id"`
	OwnerWallet  string `json:"owner_wallet,omitempty"`
}

// verifyResponse is returned on a successful proof.
type verifyResponse struct {
	Host         string       `json:"host"`
	DeploymentID string       `json:"deployment_id"`
	Method       VerifyMethod `json:"method"`
	ExpiresAt    time.Time    `json:"expires_at"`
}

// CustomDomainHandler implements the challenge/verify workflow over HTTP.
type CustomDomainHandler struct {
	verifier  DomainVerifier
	store     *DomainOwnershipStore
	secret    []byte
	verifyDom string
	maxPerDep int // 0 = unlimited
}

// NewDomainVerifyHandler builds the custom-domain HTTP handler. verifyDom is the
// base domain custom hosts CNAME to (e.g. "moltbunker.dev"); maxPerDeployment
// caps how many hosts a single deployment may verify (0 = unlimited).
func NewDomainVerifyHandler(verifier DomainVerifier, store *DomainOwnershipStore, secret []byte, verifyDom string, maxPerDeployment int) *CustomDomainHandler {
	return &CustomDomainHandler{
		verifier:  verifier,
		store:     store,
		secret:    secret,
		verifyDom: verifyDom,
		maxPerDep: maxPerDeployment,
	}
}

// ServeHTTP routes the two sub-paths. It matches on the trailing path element so
// the daemon can mount it under any prefix.
func (h *CustomDomainHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case strings.HasSuffix(r.URL.Path, "/challenge"):
		h.handleChallenge(w, r)
	case strings.HasSuffix(r.URL.Path, "/verify"):
		h.handleVerify(w, r)
	default:
		http.NotFound(w, r)
	}
}

// handleChallenge returns the DNS record the customer must publish. It persists
// nothing — verification is confirmed in a later /verify call.
func (h *CustomDomainHandler) handleChallenge(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req challengeRequest
	if !decodeJSONBody(w, r, &req) {
		return
	}
	if req.Host == "" || req.DeploymentID == "" {
		writeJSONError(w, http.StatusBadRequest, "host and deployment_id are required")
		return
	}

	token := GenerateVerificationToken(h.secret, req.Host, req.DeploymentID)
	resp := challengeResponse{
		Method: h.verifier.Method(),
		Token:  token,
	}
	switch h.verifier.Method() {
	case MethodTXT:
		resp.TXTName, resp.TXTValue = TXTRecord(h.secret, req.Host, req.DeploymentID)
	default:
		resp.CNAMETarget = CNAMETarget(h.secret, h.verifyDom, req.Host, req.DeploymentID)
	}
	writeJSON(w, http.StatusOK, resp)
}

// handleVerify confirms the DNS proof and, on success, persists the mapping.
func (h *CustomDomainHandler) handleVerify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req verifyRequest
	if !decodeJSONBody(w, r, &req) {
		return
	}
	if req.Host == "" || req.DeploymentID == "" {
		writeJSONError(w, http.StatusBadRequest, "host and deployment_id are required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	if err := h.verifier.Verify(ctx, req.Host, req.DeploymentID); err != nil {
		logging.Info("custom domain verification failed",
			"host", req.Host,
			"deployment_id", req.DeploymentID,
			logging.Err(err),
			logging.Component("ingress"))
		writeJSONError(w, http.StatusUnprocessableEntity, "ownership verification failed: "+err.Error())
		return
	}

	// Persist the proven mapping, enforcing the per-deployment cap atomically
	// (check + insert under a single write lock) so concurrent verifies for the
	// same deployment cannot each slip past a separate count check. A re-verify
	// of an already-mapped host for this deployment is an in-place refresh and
	// is not counted against the cap.
	rec, ok := h.store.StoreIfUnderCap(req.Host, req.DeploymentID, req.OwnerWallet, h.verifier.Method(), h.maxPerDep)
	if !ok {
		writeJSONError(w, http.StatusConflict, "custom domain limit reached for deployment")
		return
	}
	logging.Info("custom domain verified",
		"host", req.Host,
		"deployment_id", req.DeploymentID,
		"method", string(h.verifier.Method()),
		logging.Component("ingress"))
	writeJSON(w, http.StatusOK, verifyResponse{
		Host:         rec.Host,
		DeploymentID: rec.DeploymentID,
		Method:       rec.Method,
		ExpiresAt:    rec.ExpiresAt,
	})
}

// decodeJSONBody reads and unmarshals a size-limited JSON body, writing a 400
// on failure. It reports whether decoding succeeded.
func decodeJSONBody(w http.ResponseWriter, r *http.Request, dst interface{}) bool {
	body, err := io.ReadAll(io.LimitReader(r.Body, maxCustomDomainBody))
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "failed to read body")
		return false
	}
	if err := json.Unmarshal(body, dst); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid JSON")
		return false
	}
	return true
}

// writeJSON writes a JSON response with the given status.
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeJSONError writes a {"error": msg} body with the given status.
func writeJSONError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
