package storage

import (
	"io"
	"time"
)

// BucketInfo contains metadata about a storage bucket.
type BucketInfo struct {
	Name      string    `json:"name"`
	Owner     string    `json:"owner"` // Wallet address
	CreatedAt time.Time `json:"created_at"`
	Region    string    `json:"region,omitempty"`
}

// ObjectInfo contains metadata about a stored object.
type ObjectInfo struct {
	Bucket      string    `json:"bucket"`
	Key         string    `json:"key"`
	Size        int64     `json:"size"`
	ContentType string    `json:"content_type,omitempty"`
	ETag        string    `json:"etag"`
	CID         string    `json:"cid,omitempty"` // IPFS CID of blob (empty for local-only)
	Owner       string    `json:"owner"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Encryption metadata (Phase 2)
	EncryptedDEK []byte `json:"encrypted_dek,omitempty"`
	DEKNonce     []byte `json:"dek_nonce,omitempty"`
}

// ListObjectsInput configures object listing.
type ListObjectsInput struct {
	Bucket            string
	Prefix            string
	Delimiter         string
	ContinuationToken string
	MaxKeys           int
}

// ListObjectsOutput is the result of listing objects.
type ListObjectsOutput struct {
	Objects               []ObjectInfo `json:"objects"`
	CommonPrefixes        []string     `json:"common_prefixes,omitempty"`
	IsTruncated           bool         `json:"is_truncated"`
	NextContinuationToken string       `json:"next_continuation_token,omitempty"`
	KeyCount              int          `json:"key_count"`
}

// PutObjectInput configures an object write.
type PutObjectInput struct {
	Bucket      string
	Key         string
	Body        io.Reader
	ContentType string
	Owner       string // Wallet address
	Size        int64  // -1 if unknown
}

// GetObjectOutput wraps a readable object stream.
type GetObjectOutput struct {
	Body        io.ReadCloser
	Info        ObjectInfo
	ContentType string
}

// UsageReport summarizes storage usage for a wallet.
type UsageReport struct {
	WalletAddress string `json:"wallet_address"`
	TotalBytes    int64  `json:"total_bytes"`
	ObjectCount   int64  `json:"object_count"`
	BucketCount   int    `json:"bucket_count"`
}
