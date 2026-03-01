package storage

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/moltbunker/moltbunker/internal/logging"
)

// RESTHandler serves the JSON REST API for object storage under /v1/storage/.
type RESTHandler struct {
	engine *StorageEngine
}

// NewRESTHandler creates a new REST handler for the storage service.
func NewRESTHandler(engine *StorageEngine) *RESTHandler {
	return &RESTHandler{engine: engine}
}

// RegisterRoutes registers storage REST routes on the given mux.
// The caller is responsible for wrapping with auth middleware.
func (h *RESTHandler) RegisterRoutes(mux *http.ServeMux, wrapRead, wrapWrite func(http.HandlerFunc) http.HandlerFunc) {
	mux.HandleFunc("/v1/storage/buckets", wrapWrite(h.handleBuckets))
	mux.HandleFunc("/v1/storage/buckets/", wrapWrite(h.handleBucketByName))
	mux.HandleFunc("/v1/storage/objects/", wrapWrite(h.handleObject))
	mux.HandleFunc("/v1/storage/usage", wrapRead(h.handleUsage))
}

// extractWallet extracts the verified wallet address from the request context.
// The identity is injected by the auth middleware after signature verification.
// Falls back to X-Moltbunker-Verified-Wallet header set by the middleware.
func extractWallet(r *http.Request) string {
	// Read from the internal verified header (set by auth middleware, not from client)
	if wallet := r.Header.Get("X-Moltbunker-Verified-Wallet"); wallet != "" {
		return strings.ToLower(wallet)
	}
	return ""
}

// --- Bucket endpoints ---

func (h *RESTHandler) handleBuckets(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	wallet := extractWallet(r)

	switch r.Method {
	case http.MethodGet:
		buckets, err := h.engine.ListBuckets(ctx, wallet)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"buckets": buckets,
			"count":   len(buckets),
		})

	case http.MethodPost:
		var req struct {
			Name string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		bucket, err := h.engine.CreateBucket(ctx, req.Name, wallet)
		if err != nil {
			if strings.Contains(err.Error(), "already exists") {
				writeError(w, http.StatusConflict, err.Error())
				return
			}
			if strings.Contains(err.Error(), "invalid bucket name") || strings.Contains(err.Error(), "limit exceeded") {
				writeError(w, http.StatusBadRequest, err.Error())
				return
			}
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, bucket)

	default:
		w.Header().Set("Allow", "GET, POST")
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *RESTHandler) handleBucketByName(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	wallet := extractWallet(r)

	// Extract bucket name from path: /v1/storage/buckets/{name}
	name := strings.TrimPrefix(r.URL.Path, "/v1/storage/buckets/")
	if name == "" {
		writeError(w, http.StatusBadRequest, "bucket name is required")
		return
	}

	switch r.Method {
	case http.MethodGet, http.MethodHead:
		bucket, err := h.engine.HeadBucket(ctx, name)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if bucket == nil {
			writeError(w, http.StatusNotFound, fmt.Sprintf("bucket %q not found", name))
			return
		}
		if r.Method == http.MethodHead {
			w.WriteHeader(http.StatusOK)
			return
		}
		writeJSON(w, http.StatusOK, bucket)

	case http.MethodDelete:
		if err := h.engine.DeleteBucket(ctx, name, wallet); err != nil {
			if strings.Contains(err.Error(), "not found") {
				writeError(w, http.StatusNotFound, err.Error())
				return
			}
			if strings.Contains(err.Error(), "not empty") || strings.Contains(err.Error(), "permission denied") {
				writeError(w, http.StatusConflict, err.Error())
				return
			}
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		w.WriteHeader(http.StatusNoContent)

	default:
		w.Header().Set("Allow", "GET, HEAD, DELETE")
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// --- Object endpoints ---

func (h *RESTHandler) handleObject(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	wallet := extractWallet(r)

	// Parse path: /v1/storage/objects/{bucket}/{key...}
	path := strings.TrimPrefix(r.URL.Path, "/v1/storage/objects/")
	bucket, key, hasSep := strings.Cut(path, "/")
	if bucket == "" {
		writeError(w, http.StatusBadRequest, "bucket name is required")
		return
	}

	// If no key, this is a list operation
	if !hasSep || key == "" {
		if r.Method == http.MethodGet {
			h.handleListObjects(w, r, bucket)
			return
		}
		writeError(w, http.StatusBadRequest, "object key is required")
		return
	}

	switch r.Method {
	case http.MethodPut:
		contentType := r.Header.Get("Content-Type")
		if contentType == "" {
			contentType = "application/octet-stream"
		}

		obj, err := h.engine.PutObject(ctx, &PutObjectInput{
			Bucket:      bucket,
			Key:         key,
			Body:        r.Body,
			ContentType: contentType,
			Owner:       wallet,
			Size:        r.ContentLength,
		})
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				writeError(w, http.StatusNotFound, err.Error())
				return
			}
			if strings.Contains(err.Error(), "permission denied") {
				writeError(w, http.StatusForbidden, err.Error())
				return
			}
			if strings.Contains(err.Error(), "too large") || strings.Contains(err.Error(), "must not contain") {
				writeError(w, http.StatusBadRequest, err.Error())
				return
			}
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		w.Header().Set("ETag", `"`+obj.ETag+`"`)
		writeJSON(w, http.StatusOK, obj)

	case http.MethodGet:
		out, err := h.engine.GetObject(ctx, bucket, key)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				writeError(w, http.StatusNotFound, err.Error())
				return
			}
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		defer out.Body.Close()

		w.Header().Set("Content-Type", out.ContentType)
		w.Header().Set("ETag", `"`+out.Info.ETag+`"`)
		w.Header().Set("Content-Length", fmt.Sprintf("%d", out.Info.Size))
		w.WriteHeader(http.StatusOK)
		if _, err := io.Copy(w, out.Body); err != nil {
			logging.Debug("error streaming object",
				"bucket", bucket, "key", key,
				"error", err.Error(),
				logging.Component("storage"))
		}

	case http.MethodHead:
		obj, err := h.engine.HeadObject(ctx, bucket, key)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				writeError(w, http.StatusNotFound, err.Error())
				return
			}
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		w.Header().Set("Content-Type", obj.ContentType)
		w.Header().Set("ETag", `"`+obj.ETag+`"`)
		w.Header().Set("Content-Length", fmt.Sprintf("%d", obj.Size))
		w.WriteHeader(http.StatusOK)

	case http.MethodDelete:
		if err := h.engine.DeleteObject(ctx, bucket, key, wallet); err != nil {
			if strings.Contains(err.Error(), "not found") {
				writeError(w, http.StatusNotFound, err.Error())
				return
			}
			if strings.Contains(err.Error(), "permission denied") {
				writeError(w, http.StatusForbidden, err.Error())
				return
			}
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		w.WriteHeader(http.StatusNoContent)

	default:
		w.Header().Set("Allow", "GET, PUT, HEAD, DELETE")
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *RESTHandler) handleListObjects(w http.ResponseWriter, r *http.Request, bucket string) {
	ctx := r.Context()
	q := r.URL.Query()

	maxKeys := 1000
	if v := q.Get("max-keys"); v != "" {
		fmt.Sscanf(v, "%d", &maxKeys)
	}

	out, err := h.engine.ListObjects(ctx, &ListObjectsInput{
		Bucket:            bucket,
		Prefix:            q.Get("prefix"),
		Delimiter:         q.Get("delimiter"),
		ContinuationToken: q.Get("continuation-token"),
		MaxKeys:           maxKeys,
	})
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, out)
}

// --- Usage endpoint ---

func (h *RESTHandler) handleUsage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	wallet := extractWallet(r)
	report, err := h.engine.GetUsage(r.Context(), wallet)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, report)
}

// --- Helpers ---

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
