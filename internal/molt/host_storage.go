package molt

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"

	"github.com/tetratelabs/wazero/api"

	"github.com/moltbunker/moltbunker/internal/logging"
	"github.com/moltbunker/moltbunker/internal/storage"
)

const storageMaxBodyBytes = 10 * 1024 * 1024 // 10MB per object via host function

// storageRequest is the JSON envelope for storage operations from WASM.
type storageRequest struct {
	Bucket      string `json:"bucket"`
	Key         string `json:"key"`
	Body        string `json:"body,omitempty"`         // base64-encoded (put only)
	ContentType string `json:"content_type,omitempty"` // put only
	Prefix      string `json:"prefix,omitempty"`       // list only
	MaxKeys     int    `json:"max_keys,omitempty"`     // list only
}

// hostStoragePut stores an object in a scoped storage bucket.
// Params: [req_ptr i32, req_len i32] → [handle i32]
func hostStoragePut(ctx context.Context, mod api.Module, stack []uint64) {
	reqPtr := api.DecodeU32(stack[0])
	reqLen := api.DecodeU32(stack[1])

	svc := servicesFromContext(ctx)
	if svc == nil {
		stack[0] = api.EncodeI32(-1)
		return
	}

	if !svc.Config.StorageEnabled || svc.Storage == nil {
		stack[0] = api.EncodeI32(svc.results.StoreError("storage: service disabled"))
		return
	}

	req, err := readStorageRequest(mod, reqPtr, reqLen)
	if err != nil {
		stack[0] = api.EncodeI32(svc.results.StoreError(fmt.Sprintf("storage_put: %v", err)))
		return
	}

	if err := enforceBucketScope(req.Bucket, svc.Config.StorageBucket); err != nil {
		stack[0] = api.EncodeI32(svc.results.StoreError(fmt.Sprintf("storage_put: %v", err)))
		return
	}

	// Decode body
	bodyBytes, err := base64.StdEncoding.DecodeString(req.Body)
	if err != nil {
		stack[0] = api.EncodeI32(svc.results.StoreError(fmt.Sprintf("storage_put: invalid base64 body: %v", err)))
		return
	}

	if len(bodyBytes) > storageMaxBodyBytes {
		stack[0] = api.EncodeI32(svc.results.StoreError("storage_put: body exceeds 10MB limit"))
		return
	}

	contentType := req.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	info, err := svc.Storage.PutObject(ctx, &storage.PutObjectInput{
		Bucket:      req.Bucket,
		Key:         req.Key,
		Body:        bytes.NewReader(bodyBytes),
		ContentType: contentType,
		Owner:       svc.Owner,
		Size:        int64(len(bodyBytes)),
	})
	if err != nil {
		stack[0] = api.EncodeI32(svc.results.StoreError(fmt.Sprintf("storage_put: %v", err)))
		return
	}

	infoJSON, err := json.Marshal(info)
	if err != nil {
		stack[0] = api.EncodeI32(svc.results.StoreError(fmt.Sprintf("storage_put: marshal info: %v", err)))
		return
	}

	logging.Debug("host.storage_put completed", "bucket", req.Bucket, "key", req.Key, "size", len(bodyBytes))
	stack[0] = api.EncodeI32(svc.results.Store(infoJSON))
}

// hostStorageGet retrieves an object from a scoped storage bucket.
// Params: [req_ptr i32, req_len i32] → [handle i32]
// Returns a handle with the raw object bytes.
func hostStorageGet(ctx context.Context, mod api.Module, stack []uint64) {
	reqPtr := api.DecodeU32(stack[0])
	reqLen := api.DecodeU32(stack[1])

	svc := servicesFromContext(ctx)
	if svc == nil {
		stack[0] = api.EncodeI32(-1)
		return
	}

	if !svc.Config.StorageEnabled || svc.Storage == nil {
		stack[0] = api.EncodeI32(svc.results.StoreError("storage: service disabled"))
		return
	}

	req, err := readStorageRequest(mod, reqPtr, reqLen)
	if err != nil {
		stack[0] = api.EncodeI32(svc.results.StoreError(fmt.Sprintf("storage_get: %v", err)))
		return
	}

	if err := enforceBucketScope(req.Bucket, svc.Config.StorageBucket); err != nil {
		stack[0] = api.EncodeI32(svc.results.StoreError(fmt.Sprintf("storage_get: %v", err)))
		return
	}

	output, err := svc.Storage.GetObject(ctx, req.Bucket, req.Key, svc.Owner)
	if err != nil {
		stack[0] = api.EncodeI32(svc.results.StoreError(fmt.Sprintf("storage_get: %v", err)))
		return
	}
	defer output.Body.Close()

	body, err := io.ReadAll(io.LimitReader(output.Body, storageMaxBodyBytes))
	if err != nil {
		stack[0] = api.EncodeI32(svc.results.StoreError(fmt.Sprintf("storage_get: read body: %v", err)))
		return
	}

	logging.Debug("host.storage_get completed", "bucket", req.Bucket, "key", req.Key, "size", len(body))
	stack[0] = api.EncodeI32(svc.results.Store(body))
}

// hostStorageDelete deletes an object from a scoped storage bucket.
// Params: [req_ptr i32, req_len i32] → [result i32]
// Returns 0 on success, negative error handle on failure.
func hostStorageDelete(ctx context.Context, mod api.Module, stack []uint64) {
	reqPtr := api.DecodeU32(stack[0])
	reqLen := api.DecodeU32(stack[1])

	svc := servicesFromContext(ctx)
	if svc == nil {
		stack[0] = api.EncodeI32(-1)
		return
	}

	if !svc.Config.StorageEnabled || svc.Storage == nil {
		stack[0] = api.EncodeI32(svc.results.StoreError("storage: service disabled"))
		return
	}

	req, err := readStorageRequest(mod, reqPtr, reqLen)
	if err != nil {
		stack[0] = api.EncodeI32(svc.results.StoreError(fmt.Sprintf("storage_delete: %v", err)))
		return
	}

	if err := enforceBucketScope(req.Bucket, svc.Config.StorageBucket); err != nil {
		stack[0] = api.EncodeI32(svc.results.StoreError(fmt.Sprintf("storage_delete: %v", err)))
		return
	}

	if err := svc.Storage.DeleteObject(ctx, req.Bucket, req.Key, svc.Owner); err != nil {
		stack[0] = api.EncodeI32(svc.results.StoreError(fmt.Sprintf("storage_delete: %v", err)))
		return
	}

	logging.Debug("host.storage_delete completed", "bucket", req.Bucket, "key", req.Key)
	stack[0] = 0
}

// hostStorageList lists objects in a scoped storage bucket.
// Params: [req_ptr i32, req_len i32] → [handle i32]
func hostStorageList(ctx context.Context, mod api.Module, stack []uint64) {
	reqPtr := api.DecodeU32(stack[0])
	reqLen := api.DecodeU32(stack[1])

	svc := servicesFromContext(ctx)
	if svc == nil {
		stack[0] = api.EncodeI32(-1)
		return
	}

	if !svc.Config.StorageEnabled || svc.Storage == nil {
		stack[0] = api.EncodeI32(svc.results.StoreError("storage: service disabled"))
		return
	}

	req, err := readStorageRequest(mod, reqPtr, reqLen)
	if err != nil {
		stack[0] = api.EncodeI32(svc.results.StoreError(fmt.Sprintf("storage_list: %v", err)))
		return
	}

	if err := enforceBucketScope(req.Bucket, svc.Config.StorageBucket); err != nil {
		stack[0] = api.EncodeI32(svc.results.StoreError(fmt.Sprintf("storage_list: %v", err)))
		return
	}

	maxKeys := req.MaxKeys
	if maxKeys <= 0 || maxKeys > 1000 {
		maxKeys = 1000
	}

	output, err := svc.Storage.ListObjects(ctx, &storage.ListObjectsInput{
		Bucket:  req.Bucket,
		Prefix:  req.Prefix,
		MaxKeys: maxKeys,
		Owner:   svc.Owner,
	})
	if err != nil {
		stack[0] = api.EncodeI32(svc.results.StoreError(fmt.Sprintf("storage_list: %v", err)))
		return
	}

	outputJSON, err := json.Marshal(output)
	if err != nil {
		stack[0] = api.EncodeI32(svc.results.StoreError(fmt.Sprintf("storage_list: marshal: %v", err)))
		return
	}

	logging.Debug("host.storage_list completed", "bucket", req.Bucket, "prefix", req.Prefix, "count", output.KeyCount)
	stack[0] = api.EncodeI32(svc.results.Store(outputJSON))
}

// readStorageRequest reads and parses a storage request from WASM memory.
func readStorageRequest(mod api.Module, ptr, length uint32) (*storageRequest, error) {
	mem := mod.Memory()
	if mem == nil {
		return nil, fmt.Errorf("no memory")
	}

	buf, ok := mem.Read(ptr, length)
	if !ok {
		return nil, fmt.Errorf("invalid memory read at ptr=%d len=%d", ptr, length)
	}

	var req storageRequest
	if err := json.Unmarshal(buf, &req); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	if req.Bucket == "" {
		return nil, fmt.Errorf("bucket is required")
	}

	return &req, nil
}

// enforceBucketScope ensures the requested bucket matches the deployment's scoped bucket.
func enforceBucketScope(requested, scoped string) error {
	if scoped == "" {
		return fmt.Errorf("no storage bucket configured for this deployment")
	}
	if requested != scoped {
		return fmt.Errorf("bucket %q not allowed (scoped to %q)", requested, scoped)
	}
	return nil
}
