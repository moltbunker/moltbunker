package molt

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/moltbunker/moltbunker/internal/logging"
	"github.com/moltbunker/moltbunker/internal/security"
)

const maxRequestBodySize = 10 * 1024 * 1024 // 10MB

// E2E encryption headers — requester sets X-Molt-Encrypted on request,
// provider mirrors it on response with encryption metadata.
const (
	HeaderMoltEncrypted          = "X-Molt-Encrypted"
	HeaderMoltEncryptionMetadata = "X-Molt-Encryption-Metadata"
)

// MoltHTTPHandler dispatches HTTP requests to a compiled Molt function.
// Implements http.Handler. Concurrency is managed by the runtime's semaphore.
type MoltHTTPHandler struct {
	runtime       *MoltRuntime
	compiled      *CompiledMolt
	deploymentID  string
	encryptionMgr *security.DeploymentEncryptionManager // optional — nil disables E2E encryption
}

// NewMoltHTTPHandler creates an HTTP handler that invokes the given compiled Molt.
func NewMoltHTTPHandler(runtime *MoltRuntime, compiled *CompiledMolt, deploymentID string) *MoltHTTPHandler {
	return &MoltHTTPHandler{
		runtime:      runtime,
		compiled:     compiled,
		deploymentID: deploymentID,
	}
}

// SetEncryptionManager enables E2E encryption for this handler's deployment.
// When set, requests with X-Molt-Encrypted: true are decrypted before invocation,
// and responses are encrypted before delivery.
func (h *MoltHTTPHandler) SetEncryptionManager(em *security.DeploymentEncryptionManager) {
	h.encryptionMgr = em
}

// ServeHTTP reads the incoming request, invokes the Molt, and writes the response.
func (h *MoltHTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Read body with size limit
	body, err := io.ReadAll(io.LimitReader(r.Body, maxRequestBodySize+1))
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusInternalServerError)
		return
	}
	if len(body) > maxRequestBodySize {
		http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
		return
	}

	// E2E decryption: if request is encrypted and we have keys, decrypt body
	requestEncrypted := r.Header.Get(HeaderMoltEncrypted) == "true"
	if requestEncrypted && h.encryptionMgr != nil && len(body) > 0 {
		decrypted, err := h.encryptionMgr.DecryptData(h.deploymentID, body)
		if err != nil {
			logging.Error("molt e2e decrypt failed", "deployment", h.deploymentID, "error", err)
			http.Error(w, "decryption failed", http.StatusBadRequest)
			return
		}
		body = decrypted
	}

	// Flatten headers (first value only), excluding encryption headers from guest
	headers := make(map[string]string, len(r.Header))
	for k, v := range r.Header {
		if len(v) > 0 && k != HeaderMoltEncrypted && k != HeaderMoltEncryptionMetadata {
			headers[k] = v[0]
		}
	}

	invocation := MoltInvocation{
		DeploymentID: h.deploymentID,
		Method:       r.Method,
		Path:         r.URL.Path,
		Headers:      headers,
		Body:         body,
	}

	result, err := h.runtime.Invoke(r.Context(), h.compiled, invocation)
	if err != nil {
		logging.Error("molt invocation failed", "deployment", h.deploymentID, "error", err)
		http.Error(w, "invocation failed", http.StatusBadGateway)
		return
	}

	if result.Error != "" {
		logging.Warn("molt returned error", "deployment", h.deploymentID, "error", result.Error, "status", result.StatusCode)
	}

	// Write response headers (from WASM/Deno guest)
	for k, v := range result.Headers {
		w.Header().Set(k, v)
	}

	// E2E encryption: if request was encrypted, encrypt response body
	responseBody := result.Body
	if requestEncrypted && h.encryptionMgr != nil && len(responseBody) > 0 {
		encrypted, err := h.encryptionMgr.EncryptData(h.deploymentID, responseBody)
		if err != nil {
			logging.Error("molt e2e encrypt failed", "deployment", h.deploymentID, "error", err)
			http.Error(w, "encryption failed", http.StatusInternalServerError)
			return
		}
		responseBody = encrypted

		// Attach encryption metadata so requester can decrypt
		metadata, err := h.encryptionMgr.GetEncryptionMetadata(h.deploymentID)
		if err == nil {
			if metaJSON, err := json.Marshal(metadata); err == nil {
				w.Header().Set(HeaderMoltEncrypted, "true")
				w.Header().Set(HeaderMoltEncryptionMetadata, string(metaJSON))
			}
		}
	}

	// Write status and body
	w.WriteHeader(result.StatusCode)
	if len(responseBody) > 0 {
		if _, err := w.Write(responseBody); err != nil {
			logging.Debug("molt http write error", "deployment", h.deploymentID, "error", err)
		}
	}
}
