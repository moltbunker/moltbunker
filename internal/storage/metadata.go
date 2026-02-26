package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/moltbunker/moltbunker/internal/state"
)

// MetadataStore manages bucket and object metadata backed by bbolt.
// Object keys in bbolt use the format "bucket/key" for flat storage.
type MetadataStore struct {
	store state.StateStore
}

// NewMetadataStore creates a new metadata store backed by the given StateStore.
func NewMetadataStore(store state.StateStore) *MetadataStore {
	return &MetadataStore{store: store}
}

// --- Bucket operations ---

// PutBucket persists a bucket record.
func (m *MetadataStore) PutBucket(ctx context.Context, bucket *BucketInfo) error {
	data, err := json.Marshal(bucket)
	if err != nil {
		return fmt.Errorf("marshal bucket: %w", err)
	}
	return m.store.PutStorageBucket(ctx, bucket.Name, data)
}

// GetBucket retrieves a bucket by name. Returns nil if not found.
func (m *MetadataStore) GetBucket(ctx context.Context, name string) (*BucketInfo, error) {
	data, err := m.store.GetStorageBucket(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("get bucket: %w", err)
	}
	if data == nil {
		return nil, nil
	}
	var bucket BucketInfo
	if err := json.Unmarshal(data, &bucket); err != nil {
		return nil, fmt.Errorf("unmarshal bucket: %w", err)
	}
	return &bucket, nil
}

// DeleteBucket removes a bucket record.
func (m *MetadataStore) DeleteBucket(ctx context.Context, name string) error {
	return m.store.DeleteStorageBucket(ctx, name)
}

// ListBuckets returns all buckets, optionally filtered by owner.
func (m *MetadataStore) ListBuckets(ctx context.Context, owner string) ([]BucketInfo, error) {
	all, err := m.store.ListStorageBuckets(ctx)
	if err != nil {
		return nil, fmt.Errorf("list buckets: %w", err)
	}

	var buckets []BucketInfo
	for _, data := range all {
		var b BucketInfo
		if err := json.Unmarshal(data, &b); err != nil {
			continue // skip corrupt entries
		}
		if owner == "" || b.Owner == owner {
			buckets = append(buckets, b)
		}
	}

	sort.Slice(buckets, func(i, j int) bool {
		return buckets[i].Name < buckets[j].Name
	})
	return buckets, nil
}

// --- Object operations ---

// objectKey constructs the bbolt key for an object: "bucket/key".
func objectKey(bucket, key string) string {
	return bucket + "/" + key
}

// PutObject persists an object metadata record.
func (m *MetadataStore) PutObject(ctx context.Context, obj *ObjectInfo) error {
	data, err := json.Marshal(obj)
	if err != nil {
		return fmt.Errorf("marshal object: %w", err)
	}
	return m.store.PutStorageObject(ctx, objectKey(obj.Bucket, obj.Key), data)
}

// GetObject retrieves object metadata. Returns nil if not found.
func (m *MetadataStore) GetObject(ctx context.Context, bucket, key string) (*ObjectInfo, error) {
	data, err := m.store.GetStorageObject(ctx, objectKey(bucket, key))
	if err != nil {
		return nil, fmt.Errorf("get object: %w", err)
	}
	if data == nil {
		return nil, nil
	}
	var obj ObjectInfo
	if err := json.Unmarshal(data, &obj); err != nil {
		return nil, fmt.Errorf("unmarshal object: %w", err)
	}
	return &obj, nil
}

// DeleteObject removes an object metadata record.
func (m *MetadataStore) DeleteObject(ctx context.Context, bucket, key string) error {
	return m.store.DeleteStorageObject(ctx, objectKey(bucket, key))
}

// ListObjects lists objects in a bucket with prefix/delimiter filtering.
func (m *MetadataStore) ListObjects(ctx context.Context, input *ListObjectsInput) (*ListObjectsOutput, error) {
	all, err := m.store.ListStorageObjects(ctx)
	if err != nil {
		return nil, fmt.Errorf("list objects: %w", err)
	}

	bucketPrefix := input.Bucket + "/"
	searchPrefix := bucketPrefix + input.Prefix

	maxKeys := input.MaxKeys
	if maxKeys <= 0 {
		maxKeys = 1000
	}

	// Collect matching objects and common prefixes
	var objects []ObjectInfo
	prefixSet := make(map[string]bool)

	// Sort keys for consistent ordering and pagination
	sortedKeys := make([]string, 0, len(all))
	for k := range all {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)

	pastToken := input.ContinuationToken == ""

	for _, k := range sortedKeys {
		if !strings.HasPrefix(k, searchPrefix) {
			continue
		}

		// Pagination: skip until we pass the continuation token
		if !pastToken {
			if k > input.ContinuationToken {
				pastToken = true
			} else {
				continue
			}
		}

		// Extract the relative key (remove bucket prefix)
		relKey := strings.TrimPrefix(k, bucketPrefix)

		// Handle delimiter (e.g., "/" for folder-like listing)
		if input.Delimiter != "" {
			afterPrefix := strings.TrimPrefix(relKey, input.Prefix)
			idx := strings.Index(afterPrefix, input.Delimiter)
			if idx >= 0 {
				// This is a "folder" — add to common prefixes
				commonPrefix := input.Prefix + afterPrefix[:idx+len(input.Delimiter)]
				prefixSet[commonPrefix] = true
				continue
			}
		}

		var obj ObjectInfo
		if err := json.Unmarshal(all[k], &obj); err != nil {
			continue
		}
		objects = append(objects, obj)
	}

	// Apply max keys limit
	isTruncated := false
	nextToken := ""
	if len(objects) > maxKeys {
		isTruncated = true
		lastObj := objects[maxKeys-1]
		nextToken = objectKey(lastObj.Bucket, lastObj.Key)
		objects = objects[:maxKeys]
	}

	// Collect common prefixes
	var commonPrefixes []string
	for p := range prefixSet {
		commonPrefixes = append(commonPrefixes, p)
	}
	sort.Strings(commonPrefixes)

	return &ListObjectsOutput{
		Objects:               objects,
		CommonPrefixes:        commonPrefixes,
		IsTruncated:           isTruncated,
		NextContinuationToken: nextToken,
		KeyCount:              len(objects),
	}, nil
}

// CountObjectsInBucket counts objects in a bucket.
func (m *MetadataStore) CountObjectsInBucket(ctx context.Context, bucket string) (int, error) {
	all, err := m.store.ListStorageObjects(ctx)
	if err != nil {
		return 0, err
	}
	prefix := bucket + "/"
	count := 0
	for k := range all {
		if strings.HasPrefix(k, prefix) {
			count++
		}
	}
	return count, nil
}

// GetUsage computes storage usage for an owner across all buckets.
func (m *MetadataStore) GetUsage(ctx context.Context, owner string) (*UsageReport, error) {
	buckets, err := m.ListBuckets(ctx, owner)
	if err != nil {
		return nil, err
	}

	all, err := m.store.ListStorageObjects(ctx)
	if err != nil {
		return nil, err
	}

	report := &UsageReport{
		WalletAddress: owner,
		BucketCount:   len(buckets),
	}

	for _, data := range all {
		var obj ObjectInfo
		if err := json.Unmarshal(data, &obj); err != nil {
			continue
		}
		if obj.Owner == owner {
			report.TotalBytes += obj.Size
			report.ObjectCount++
		}
	}

	return report, nil
}
