package storage

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/moltbunker/moltbunker/internal/logging"
	"github.com/moltbunker/moltbunker/internal/state"
)

// StorageEngine is the main orchestrator for object storage operations.
// Phase 1: local blob storage with bbolt metadata.
// Phase 2 adds encryption and IPFS distribution.
type StorageEngine struct {
	dataDir  string
	metadata *MetadataStore
	config   EngineConfig
}

// EngineConfig configures the storage engine.
type EngineConfig struct {
	MaxBuckets    int   // Max buckets per wallet
	MaxObjectSize int64 // Max single object size in bytes
}

// DefaultEngineConfig returns sensible defaults.
func DefaultEngineConfig() EngineConfig {
	return EngineConfig{
		MaxBuckets:    100,
		MaxObjectSize: 5 * 1024 * 1024 * 1024, // 5GB
	}
}

// NewStorageEngine creates a new storage engine.
// dataDir is the root directory for blob storage (e.g., ~/.moltbunker/storage).
func NewStorageEngine(dataDir string, store state.StateStore, cfg EngineConfig) (*StorageEngine, error) {
	blobDir := filepath.Join(dataDir, "blobs")
	if err := os.MkdirAll(blobDir, 0700); err != nil {
		return nil, fmt.Errorf("create blob directory: %w", err)
	}

	return &StorageEngine{
		dataDir:  dataDir,
		metadata: NewMetadataStore(store),
		config:   cfg,
	}, nil
}

// --- Bucket operations ---

// bucketNameRegex validates bucket names: 3-63 chars, lowercase letters, digits, hyphens.
var bucketNameRegex = regexp.MustCompile(`^[a-z0-9][a-z0-9\-]{1,61}[a-z0-9]$`)

// CreateBucket creates a new storage bucket.
func (e *StorageEngine) CreateBucket(ctx context.Context, name, owner string) (*BucketInfo, error) {
	if !bucketNameRegex.MatchString(name) {
		return nil, fmt.Errorf("invalid bucket name %q: must be 3-63 lowercase alphanumeric chars or hyphens", name)
	}

	if owner == "" {
		return nil, fmt.Errorf("owner is required")
	}

	// Check if bucket already exists
	existing, err := e.metadata.GetBucket(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("check existing bucket: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("bucket %q already exists", name)
	}

	// Check bucket limit per wallet
	buckets, err := e.metadata.ListBuckets(ctx, owner)
	if err != nil {
		return nil, fmt.Errorf("count buckets: %w", err)
	}
	if len(buckets) >= e.config.MaxBuckets {
		return nil, fmt.Errorf("bucket limit exceeded: max %d buckets per wallet", e.config.MaxBuckets)
	}

	bucket := &BucketInfo{
		Name:      name,
		Owner:     owner,
		CreatedAt: time.Now().UTC(),
	}

	if err := e.metadata.PutBucket(ctx, bucket); err != nil {
		return nil, fmt.Errorf("persist bucket: %w", err)
	}

	logging.Info("storage bucket created",
		"bucket", name,
		"owner", owner,
		logging.Component("storage"))

	return bucket, nil
}

// DeleteBucket deletes a bucket. The bucket must be empty.
func (e *StorageEngine) DeleteBucket(ctx context.Context, name, owner string) error {
	bucket, err := e.metadata.GetBucket(ctx, name)
	if err != nil {
		return fmt.Errorf("get bucket: %w", err)
	}
	if bucket == nil {
		return fmt.Errorf("bucket %q not found", name)
	}
	if bucket.Owner != owner {
		return fmt.Errorf("permission denied: bucket %q is owned by another wallet", name)
	}

	// Check bucket is empty
	count, err := e.metadata.CountObjectsInBucket(ctx, name)
	if err != nil {
		return fmt.Errorf("count objects: %w", err)
	}
	if count > 0 {
		return fmt.Errorf("bucket %q is not empty (%d objects)", name, count)
	}

	if err := e.metadata.DeleteBucket(ctx, name); err != nil {
		return fmt.Errorf("delete bucket: %w", err)
	}

	logging.Info("storage bucket deleted",
		"bucket", name,
		logging.Component("storage"))

	return nil
}

// ListBuckets lists all buckets for a wallet.
func (e *StorageEngine) ListBuckets(ctx context.Context, owner string) ([]BucketInfo, error) {
	return e.metadata.ListBuckets(ctx, owner)
}

// HeadBucket checks if a bucket exists and returns its info.
func (e *StorageEngine) HeadBucket(ctx context.Context, name string) (*BucketInfo, error) {
	return e.metadata.GetBucket(ctx, name)
}

// --- Object operations ---

// PutObject stores an object in a bucket.
// The body is read, stored locally as a blob, and metadata is persisted in bbolt.
func (e *StorageEngine) PutObject(ctx context.Context, input *PutObjectInput) (*ObjectInfo, error) {
	if input.Key == "" {
		return nil, fmt.Errorf("object key is required")
	}
	if strings.ContainsAny(input.Key, "\x00\n\r") || strings.HasPrefix(input.Key, "/") {
		return nil, fmt.Errorf("object key contains invalid characters")
	}
	if strings.Contains(input.Key, "..") {
		return nil, fmt.Errorf("object key must not contain '..'")
	}

	// Verify bucket exists and caller owns it
	bucket, err := e.metadata.GetBucket(ctx, input.Bucket)
	if err != nil {
		return nil, fmt.Errorf("check bucket: %w", err)
	}
	if bucket == nil {
		return nil, fmt.Errorf("bucket %q not found", input.Bucket)
	}
	if bucket.Owner != input.Owner {
		return nil, fmt.Errorf("permission denied: bucket %q is owned by another wallet", input.Bucket)
	}

	// Write blob to local file
	blobPath, err := e.blobPath(input.Bucket, input.Key)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(blobPath), 0700); err != nil {
		return nil, fmt.Errorf("create blob dir: %w", err)
	}

	f, err := os.CreateTemp(filepath.Dir(blobPath), ".tmp-*")
	if err != nil {
		return nil, fmt.Errorf("create temp file: %w", err)
	}
	tempPath := f.Name()

	// Hash while writing for ETag
	hasher := md5.New()
	tee := io.TeeReader(input.Body, hasher)

	n, err := io.Copy(f, tee)
	if err != nil {
		f.Close()
		os.Remove(tempPath)
		return nil, fmt.Errorf("write blob: %w", err)
	}
	f.Close()

	// Check size limit
	if e.config.MaxObjectSize > 0 && n > e.config.MaxObjectSize {
		os.Remove(tempPath)
		return nil, fmt.Errorf("object too large: %d bytes exceeds max %d", n, e.config.MaxObjectSize)
	}

	// Atomic rename
	if err := os.Rename(tempPath, blobPath); err != nil {
		os.Remove(tempPath)
		return nil, fmt.Errorf("rename blob: %w", err)
	}

	etag := hex.EncodeToString(hasher.Sum(nil))
	now := time.Now().UTC()

	contentType := input.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	obj := &ObjectInfo{
		Bucket:      input.Bucket,
		Key:         input.Key,
		Size:        n,
		ContentType: contentType,
		ETag:        etag,
		Owner:       input.Owner,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := e.metadata.PutObject(ctx, obj); err != nil {
		os.Remove(blobPath)
		return nil, fmt.Errorf("persist object metadata: %w", err)
	}

	logging.Debug("object stored",
		"bucket", input.Bucket,
		"key", input.Key,
		"size", n,
		logging.Component("storage"))

	return obj, nil
}

// GetObject retrieves an object from a bucket.
func (e *StorageEngine) GetObject(ctx context.Context, bucket, key string) (*GetObjectOutput, error) {
	obj, err := e.metadata.GetObject(ctx, bucket, key)
	if err != nil {
		return nil, fmt.Errorf("get metadata: %w", err)
	}
	if obj == nil {
		return nil, fmt.Errorf("object %q not found in bucket %q", key, bucket)
	}

	blobPath, err := e.blobPath(bucket, key)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(blobPath)
	if err != nil {
		return nil, fmt.Errorf("open blob: %w", err)
	}

	return &GetObjectOutput{
		Body:        f,
		Info:        *obj,
		ContentType: obj.ContentType,
	}, nil
}

// HeadObject returns object metadata without the body.
func (e *StorageEngine) HeadObject(ctx context.Context, bucket, key string) (*ObjectInfo, error) {
	obj, err := e.metadata.GetObject(ctx, bucket, key)
	if err != nil {
		return nil, fmt.Errorf("get metadata: %w", err)
	}
	if obj == nil {
		return nil, fmt.Errorf("object %q not found in bucket %q", key, bucket)
	}
	return obj, nil
}

// DeleteObject removes an object from a bucket.
func (e *StorageEngine) DeleteObject(ctx context.Context, bucket, key, owner string) error {
	// Verify ownership through bucket
	bucketInfo, err := e.metadata.GetBucket(ctx, bucket)
	if err != nil {
		return fmt.Errorf("get bucket: %w", err)
	}
	if bucketInfo == nil {
		return fmt.Errorf("bucket %q not found", bucket)
	}
	if bucketInfo.Owner != owner {
		return fmt.Errorf("permission denied")
	}

	// Check object exists
	obj, err := e.metadata.GetObject(ctx, bucket, key)
	if err != nil {
		return fmt.Errorf("get object: %w", err)
	}
	if obj == nil {
		return fmt.Errorf("object %q not found in bucket %q", key, bucket)
	}

	// Delete blob file
	blobPath, pathErr := e.blobPath(bucket, key)
	if pathErr != nil {
		return pathErr
	}
	if err := os.Remove(blobPath); err != nil && !os.IsNotExist(err) {
		logging.Warn("failed to delete blob file",
			"path", blobPath,
			"error", err.Error(),
			logging.Component("storage"))
	}

	// Delete metadata
	if err := e.metadata.DeleteObject(ctx, bucket, key); err != nil {
		return fmt.Errorf("delete metadata: %w", err)
	}

	logging.Debug("object deleted",
		"bucket", bucket,
		"key", key,
		logging.Component("storage"))

	return nil
}

// ListObjects lists objects in a bucket.
func (e *StorageEngine) ListObjects(ctx context.Context, input *ListObjectsInput) (*ListObjectsOutput, error) {
	// Verify bucket exists
	bucket, err := e.metadata.GetBucket(ctx, input.Bucket)
	if err != nil {
		return nil, fmt.Errorf("get bucket: %w", err)
	}
	if bucket == nil {
		return nil, fmt.Errorf("bucket %q not found", input.Bucket)
	}

	return e.metadata.ListObjects(ctx, input)
}

// GetUsage returns storage usage for a wallet.
func (e *StorageEngine) GetUsage(ctx context.Context, owner string) (*UsageReport, error) {
	return e.metadata.GetUsage(ctx, owner)
}

// --- Internal helpers ---

// blobPath returns the local filesystem path for an object's blob.
// Uses a flat directory structure under blobs/<bucket>/<key-path>.
// Returns an error if the resolved path escapes the data directory.
func (e *StorageEngine) blobPath(bucket, key string) (string, error) {
	p := filepath.Join(e.dataDir, "blobs", bucket, key)
	absPath := filepath.Clean(p)
	absRoot := filepath.Clean(filepath.Join(e.dataDir, "blobs")) + string(filepath.Separator)
	if !strings.HasPrefix(absPath+string(filepath.Separator), absRoot) {
		return "", fmt.Errorf("path traversal detected")
	}
	return absPath, nil
}
