package storage

import (
	"bytes"
	"context"
	"crypto/md5" // #nosec G501 -- non-security use: S3-compatible ETag (content identifier), not auth/integrity
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

// MeteringHook receives storage usage events for billing. It is defined locally
// (rather than imported from the payment package) so the storage package does
// not depend on payment, avoiding an import cycle. The payment.PaymentService
// satisfies this interface structurally via its RecordStorageUpload /
// RecordStorageDelete methods.
//
// The hook is optional: when nil, the storage engine behaves exactly as before
// (no metering), so existing callers and tests need no changes.
type MeteringHook interface {
	RecordStorageUpload(wallet string, bytes int64)
	RecordStorageDelete(wallet string, bytes int64)
}

// StorageEngine is the main orchestrator for object storage operations.
// Phase 1: local blob storage with bbolt metadata.
// Phase 2 adds encryption (X25519 envelope, opt-in via keyStore) and IPFS
// distribution.
type StorageEngine struct {
	dataDir   string
	metadata  *MetadataStore
	config    EngineConfig
	meter     MeteringHook  // optional usage metering hook (nil = no metering)
	keyStore  OwnerKeyStore // optional X25519 owner key store (nil = at-rest encryption disabled)
	encryptor *OwnerKeyEncryptor
}

// SetMeteringHook installs an optional metering hook. Pass nil to disable.
// Injected at startup so storage usage is recorded for billing.
func (e *StorageEngine) SetMeteringHook(h MeteringHook) {
	e.meter = h
}

// WithOwnerKeyStore enables at-rest object encryption. When set, PutObject
// encrypts the blob with a per-object DEK sealed to the owner's X25519 key and
// GetObject transparently decrypts. Pass nil (or never call) to keep plaintext
// blobs (the default, legacy behavior). Returns the engine for chaining.
func (e *StorageEngine) WithOwnerKeyStore(ks OwnerKeyStore) *StorageEngine {
	e.keyStore = ks
	if ks != nil && e.encryptor == nil {
		e.encryptor = NewOwnerKeyEncryptor()
	}
	return e
}

// EngineConfig configures the storage engine.
type EngineConfig struct {
	MaxBuckets    int   // Max buckets per wallet
	MaxObjectSize int64 // Max single object size in bytes

	// EncryptedMaxInMemoryBytes caps the size of an object that the at-rest
	// encryption path will accept. The current envelope scheme is NOT streaming:
	// PutObject reads the whole plaintext blob into memory to seal it, and
	// GetObject/decryptBlob reads the whole ciphertext into memory to open it.
	// With MaxObjectSize at 5GB, an unconstrained encrypted Put/Get would hold
	// multiple GB resident (plaintext + ciphertext copies) — an OOM/DoS vector
	// that the plaintext io.Copy path avoids. We therefore reject encrypted
	// objects larger than this ceiling with a clear error.
	//
	// Zero falls back to defaultEncryptedMaxInMemoryBytes (256MB). It is only
	// consulted when at-rest encryption is enabled (keyStore != nil); the
	// plaintext path streams and is unaffected. A full chunked/framed streaming
	// AEAD that removes this ceiling is tracked as a follow-up (see
	// daemon-todo.md).
	EncryptedMaxInMemoryBytes int64
}

// defaultEncryptedMaxInMemoryBytes is the default in-memory ceiling for the
// (non-streaming) encrypted object path. 256MB is large enough for typical
// config/state/object payloads while bounding per-request memory well below the
// 5GB plaintext MaxObjectSize.
const defaultEncryptedMaxInMemoryBytes = 256 * 1024 * 1024 // 256MB

// DefaultEngineConfig returns sensible defaults.
func DefaultEngineConfig() EngineConfig {
	return EngineConfig{
		MaxBuckets:                100,
		MaxObjectSize:             5 * 1024 * 1024 * 1024, // 5GB
		EncryptedMaxInMemoryBytes: defaultEncryptedMaxInMemoryBytes,
	}
}

// encryptedMaxInMemoryBytes returns the effective in-memory ceiling for the
// encrypted path, applying the default when the config value is unset (<= 0).
func (e *StorageEngine) encryptedMaxInMemoryBytes() int64 {
	if e.config.EncryptedMaxInMemoryBytes > 0 {
		return e.config.EncryptedMaxInMemoryBytes
	}
	return defaultEncryptedMaxInMemoryBytes
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

// HeadBucket checks if a bucket exists and the caller owns it.
func (e *StorageEngine) HeadBucket(ctx context.Context, name, owner string) (*BucketInfo, error) {
	bucket, err := e.metadata.GetBucket(ctx, name)
	if err != nil {
		return nil, err
	}
	if bucket == nil {
		return nil, nil
	}
	if bucket.Owner != owner {
		return nil, nil // hide existence from non-owners
	}
	return bucket, nil
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
	// #nosec G401 -- non-security use: S3-compatible ETag (content identifier), not auth/integrity
	hasher := md5.New()
	tee := io.TeeReader(input.Body, hasher)

	n, err := io.Copy(f, tee)
	if err != nil {
		_ = f.Close()
		_ = os.Remove(tempPath)
		return nil, fmt.Errorf("write blob: %w", err)
	}
	_ = f.Close()

	// Check size limit
	if e.config.MaxObjectSize > 0 && n > e.config.MaxObjectSize {
		_ = os.Remove(tempPath)
		return nil, fmt.Errorf("object too large: %d bytes exceeds max %d", n, e.config.MaxObjectSize)
	}

	etag := hex.EncodeToString(hasher.Sum(nil))

	// At-rest encryption (opt-in via keyStore): seal the blob to the owner's
	// X25519 key. ETag and Size are kept over the PLAINTEXT (S3 semantics); only
	// the on-disk bytes are ciphertext. We read the freshly-written temp file
	// back, encrypt, and overwrite it before the atomic rename.
	var sealedDEKBytes []byte
	if e.keyStore != nil && e.encryptor != nil {
		// In-memory ceiling: the envelope scheme buffers the whole blob (read
		// back + encrypt produces a second copy). Reject oversized objects rather
		// than risk OOM. n is the plaintext size already on disk in tempPath.
		if max := e.encryptedMaxInMemoryBytes(); n > max {
			_ = os.Remove(tempPath)
			return nil, fmt.Errorf("encrypted object too large: %d bytes exceeds in-memory encryption limit of %d bytes (at-rest encryption buffers the whole object; use a smaller object or disable storage encryption)", n, max)
		}
		ownerPub, keyErr := e.keyStore.PublicKeyForOwner(input.Owner)
		if keyErr != nil {
			_ = os.Remove(tempPath)
			return nil, fmt.Errorf("resolve owner key: %w", keyErr)
		}
		// #nosec G304 -- tempPath is a CreateTemp file in the blob dir we just wrote.
		plain, readErr := os.ReadFile(tempPath)
		if readErr != nil {
			_ = os.Remove(tempPath)
			return nil, fmt.Errorf("read blob for encryption: %w", readErr)
		}
		res, encErr := e.encryptor.Encrypt(plain, ownerPub)
		if encErr != nil {
			_ = os.Remove(tempPath)
			return nil, fmt.Errorf("encrypt blob: %w", encErr)
		}
		sealedDEKBytes, encErr = MarshalSealedDEK(res.SealedDEK)
		if encErr != nil {
			_ = os.Remove(tempPath)
			return nil, fmt.Errorf("marshal sealed DEK: %w", encErr)
		}
		// #nosec G703 G304 -- tempPath is the os.CreateTemp file we created in
		// filepath.Dir(blobPath); blobPath is validated by blobPath() (Clean +
		// prefix check rejects traversal). We are overwriting our own temp file
		// in place before the atomic rename, not following request input.
		if writeErr := os.WriteFile(tempPath, res.Ciphertext, 0600); writeErr != nil {
			_ = os.Remove(tempPath)
			return nil, fmt.Errorf("write encrypted blob: %w", writeErr)
		}
	}

	// Atomic rename
	if err := os.Rename(tempPath, blobPath); err != nil {
		_ = os.Remove(tempPath)
		return nil, fmt.Errorf("rename blob: %w", err)
	}

	now := time.Now().UTC()

	contentType := input.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	obj := &ObjectInfo{
		Bucket:       input.Bucket,
		Key:          input.Key,
		Size:         n,
		ContentType:  contentType,
		ETag:         etag,
		Owner:        input.Owner,
		CreatedAt:    now,
		UpdatedAt:    now,
		EncryptedDEK: sealedDEKBytes,
	}

	if err := e.metadata.PutObject(ctx, obj); err != nil {
		_ = os.Remove(blobPath)
		return nil, fmt.Errorf("persist object metadata: %w", err)
	}

	// Record usage for billing (optional, nil-safe).
	if e.meter != nil {
		e.meter.RecordStorageUpload(input.Owner, n)
	}

	logging.Debug("object stored",
		"bucket", input.Bucket,
		"key", input.Key,
		"size", n,
		logging.Component("storage"))

	return obj, nil
}

// GetObject retrieves an object from a bucket. The caller must own the bucket.
func (e *StorageEngine) GetObject(ctx context.Context, bucket, key, owner string) (*GetObjectOutput, error) {
	// Verify bucket ownership
	bucketInfo, err := e.metadata.GetBucket(ctx, bucket)
	if err != nil {
		return nil, fmt.Errorf("get bucket: %w", err)
	}
	if bucketInfo == nil {
		return nil, fmt.Errorf("bucket %q not found", bucket)
	}
	if bucketInfo.Owner != owner {
		return nil, fmt.Errorf("permission denied: bucket %q is owned by another wallet", bucket)
	}

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
	// #nosec G304 -- blobPath is validated by blobPath() (filepath.Clean + prefix check rejects traversal)
	f, err := os.Open(blobPath)
	if err != nil {
		return nil, fmt.Errorf("open blob: %w", err)
	}

	// At-rest decryption: if the object carries a sealed DEK, decrypt the blob
	// in memory and return the plaintext. Both the new X25519-envelope scheme and
	// the legacy AES-GCM-wrapped-DEK scheme are supported (back-compat read path).
	if len(obj.EncryptedDEK) > 0 {
		// In-memory ceiling: decryptBlob reads the entire ciphertext into memory.
		// obj.Size is the plaintext size (S3 semantics); the ciphertext is at most
		// that plus a small GCM/nonce overhead, so guarding on Size bounds resident
		// memory. Reject rather than risk OOM on a large encrypted object.
		if max := e.encryptedMaxInMemoryBytes(); obj.Size > max {
			_ = f.Close()
			return nil, fmt.Errorf("encrypted object too large to read: %d bytes exceeds in-memory decryption limit of %d bytes (at-rest decryption buffers the whole object)", obj.Size, max)
		}
		plain, decErr := e.decryptBlob(f, obj)
		_ = f.Close()
		if decErr != nil {
			return nil, decErr
		}
		return &GetObjectOutput{
			Body:        io.NopCloser(bytes.NewReader(plain)),
			Info:        *obj,
			ContentType: obj.ContentType,
		}, nil
	}

	return &GetObjectOutput{
		Body:        f,
		Info:        *obj,
		ContentType: obj.ContentType,
	}, nil
}

// decryptBlob reads the (encrypted) blob from r and returns the plaintext. It
// supports both the X25519-envelope scheme (new) and the legacy HKDF-derived
// symmetric scheme, distinguished by whether EncryptedDEK is a JSON envelope.
func (e *StorageEngine) decryptBlob(r io.Reader, obj *ObjectInfo) ([]byte, error) {
	ciphertext, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("read encrypted blob: %w", err)
	}

	// New scheme: JSON-encoded X25519 envelope sealed to the owner key.
	if env, ok := looksLikeX25519Envelope(obj.EncryptedDEK); ok {
		if e.keyStore == nil || e.encryptor == nil {
			return nil, fmt.Errorf("object is encrypted but no owner key store is configured")
		}
		ownerPriv, keyErr := e.keyStore.PrivateKeyForOwner(obj.Owner)
		if keyErr != nil {
			return nil, fmt.Errorf("resolve owner private key: %w", keyErr)
		}
		plain, decErr := e.encryptor.Decrypt(ciphertext, env, ownerPriv)
		if decErr != nil {
			return nil, fmt.Errorf("decrypt object: %w", decErr)
		}
		return plain, nil
	}

	// Legacy scheme: DEK wrapped with the HKDF-derived symmetric owner key.
	ownerKey, keyErr := legacyDeriveOwnerKey(obj.Owner)
	if keyErr != nil {
		return nil, fmt.Errorf("derive legacy owner key: %w", keyErr)
	}
	plain, decErr := NewObjectEncryptor().Decrypt(ciphertext, obj.EncryptedDEK, ownerKey)
	if decErr != nil {
		return nil, fmt.Errorf("decrypt legacy object: %w", decErr)
	}
	return plain, nil
}

// HeadObject returns object metadata without the body. The caller must own the bucket.
func (e *StorageEngine) HeadObject(ctx context.Context, bucket, key, owner string) (*ObjectInfo, error) {
	// Verify bucket ownership
	bucketInfo, err := e.metadata.GetBucket(ctx, bucket)
	if err != nil {
		return nil, fmt.Errorf("get bucket: %w", err)
	}
	if bucketInfo == nil {
		return nil, fmt.Errorf("bucket %q not found", bucket)
	}
	if bucketInfo.Owner != owner {
		return nil, fmt.Errorf("permission denied: bucket %q is owned by another wallet", bucket)
	}

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

	// Record usage for billing (optional, nil-safe).
	if e.meter != nil {
		e.meter.RecordStorageDelete(obj.Owner, obj.Size)
	}

	logging.Debug("object deleted",
		"bucket", bucket,
		"key", key,
		logging.Component("storage"))

	return nil
}

// ListObjects lists objects in a bucket. The caller must own the bucket.
func (e *StorageEngine) ListObjects(ctx context.Context, input *ListObjectsInput) (*ListObjectsOutput, error) {
	// Verify bucket exists and caller owns it
	bucket, err := e.metadata.GetBucket(ctx, input.Bucket)
	if err != nil {
		return nil, fmt.Errorf("get bucket: %w", err)
	}
	if bucket == nil {
		return nil, fmt.Errorf("bucket %q not found", input.Bucket)
	}
	if input.Owner == "" {
		return nil, fmt.Errorf("owner is required for listing objects")
	}
	if bucket.Owner != input.Owner {
		return nil, fmt.Errorf("permission denied: bucket %q is owned by another wallet", input.Bucket)
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
