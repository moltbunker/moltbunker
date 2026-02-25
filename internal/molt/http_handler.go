package molt

import (
	"io"
	"net/http"

	"github.com/moltbunker/moltbunker/internal/logging"
)

const maxRequestBodySize = 10 * 1024 * 1024 // 10MB

// MoltHTTPHandler dispatches HTTP requests to a compiled Molt function.
// Implements http.Handler. Concurrency is managed by the runtime's semaphore.
type MoltHTTPHandler struct {
	runtime      *MoltRuntime
	compiled     *CompiledMolt
	deploymentID string
}

// NewMoltHTTPHandler creates an HTTP handler that invokes the given compiled Molt.
func NewMoltHTTPHandler(runtime *MoltRuntime, compiled *CompiledMolt, deploymentID string) *MoltHTTPHandler {
	return &MoltHTTPHandler{
		runtime:      runtime,
		compiled:     compiled,
		deploymentID: deploymentID,
	}
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

	// Flatten headers (first value only)
	headers := make(map[string]string, len(r.Header))
	for k, v := range r.Header {
		if len(v) > 0 {
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

	// Write response headers
	for k, v := range result.Headers {
		w.Header().Set(k, v)
	}

	// Write status and body
	w.WriteHeader(result.StatusCode)
	if len(result.Body) > 0 {
		if _, err := w.Write(result.Body); err != nil {
			logging.Debug("molt http write error", "deployment", h.deploymentID, "error", err)
		}
	}
}
