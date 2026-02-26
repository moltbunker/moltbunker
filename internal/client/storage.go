package client

import (
	"encoding/json"
	"fmt"
	"time"
)

// StorageBucketInfo describes a storage bucket.
type StorageBucketInfo struct {
	Name      string    `json:"name"`
	Owner     string    `json:"owner"`
	CreatedAt time.Time `json:"created_at"`
}

// StorageObjectInfo describes a stored object.
type StorageObjectInfo struct {
	Key         string    `json:"key"`
	Bucket      string    `json:"bucket"`
	Size        int64     `json:"size"`
	ContentType string    `json:"content_type"`
	ETag        string    `json:"etag"`
	Owner       string    `json:"owner"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// StorageUsageInfo describes storage usage for a wallet.
type StorageUsageInfo struct {
	Wallet      string `json:"wallet"`
	BucketCount int    `json:"bucket_count"`
	ObjectCount int    `json:"object_count"`
	TotalBytes  int64  `json:"total_bytes"`
}

// StorageCreateBucket creates a new storage bucket.
func (c *DaemonClient) StorageCreateBucket(name string) (*StorageBucketInfo, error) {
	resp, err := c.call("storage_create_bucket", map[string]string{"name": name})
	if err != nil {
		return nil, err
	}
	var result StorageBucketInfo
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to parse create bucket response: %w", err)
	}
	return &result, nil
}

// StorageDeleteBucket deletes a storage bucket.
func (c *DaemonClient) StorageDeleteBucket(name string) error {
	_, err := c.call("storage_delete_bucket", map[string]string{"name": name})
	return err
}

// StorageListBuckets lists all storage buckets.
func (c *DaemonClient) StorageListBuckets() ([]StorageBucketInfo, error) {
	resp, err := c.call("storage_list_buckets", nil)
	if err != nil {
		return nil, err
	}
	var result []StorageBucketInfo
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to parse list buckets response: %w", err)
	}
	return result, nil
}

// StorageListObjects lists objects in a bucket.
func (c *DaemonClient) StorageListObjects(bucket, prefix string) ([]StorageObjectInfo, error) {
	resp, err := c.call("storage_list_objects", map[string]string{
		"bucket": bucket,
		"prefix": prefix,
	})
	if err != nil {
		return nil, err
	}
	var result []StorageObjectInfo
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to parse list objects response: %w", err)
	}
	return result, nil
}

// StorageDeleteObject deletes an object from a bucket.
func (c *DaemonClient) StorageDeleteObject(bucket, key string) error {
	_, err := c.call("storage_delete_object", map[string]string{
		"bucket": bucket,
		"key":    key,
	})
	return err
}

// StorageUsage returns storage usage for the current wallet.
func (c *DaemonClient) StorageUsage() (*StorageUsageInfo, error) {
	resp, err := c.call("storage_usage", nil)
	if err != nil {
		return nil, err
	}
	var result StorageUsageInfo
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to parse storage usage response: %w", err)
	}
	return &result, nil
}

// StoragePutObjectRequest is the request to upload an object.
type StoragePutObjectRequest struct {
	Bucket      string `json:"bucket"`
	Key         string `json:"key"`
	Data        []byte `json:"data"`
	ContentType string `json:"content_type"`
}

// StoragePutObject uploads an object to a bucket.
func (c *DaemonClient) StoragePutObject(req *StoragePutObjectRequest) (*StorageObjectInfo, error) {
	resp, err := c.call("storage_put_object", req)
	if err != nil {
		return nil, err
	}
	var result StorageObjectInfo
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to parse put object response: %w", err)
	}
	return &result, nil
}

// StorageGetObjectResponse is the response for getting an object.
type StorageGetObjectResponse struct {
	Data        []byte `json:"data"`
	ContentType string `json:"content_type"`
	Size        int64  `json:"size"`
}

// StorageGetObject downloads an object from a bucket.
func (c *DaemonClient) StorageGetObject(bucket, key string) (*StorageGetObjectResponse, error) {
	resp, err := c.call("storage_get_object", map[string]string{
		"bucket": bucket,
		"key":    key,
	})
	if err != nil {
		return nil, err
	}
	var result StorageGetObjectResponse
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to parse get object response: %w", err)
	}
	return &result, nil
}
