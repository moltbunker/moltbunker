package storage

import (
	"bytes"
	"context"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/moltbunker/moltbunker/internal/logging"
)

// MultipartUpload tracks a multipart upload in progress.
type MultipartUpload struct {
	UploadID    string    `json:"upload_id"`
	Bucket      string    `json:"bucket"`
	Key         string    `json:"key"`
	Owner       string    `json:"owner"`
	ContentType string    `json:"content_type,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// PartInfo describes an uploaded part.
type PartInfo struct {
	PartNumber int    `json:"part_number"`
	ETag       string `json:"etag"`
	Size       int64  `json:"size"`
}

// CompletedPart identifies a part to include when completing a multipart upload.
type CompletedPart struct {
	PartNumber int    `json:"part_number"`
	ETag       string `json:"etag"`
}

// MultipartManager handles multipart upload lifecycle.
type MultipartManager struct {
	mu       sync.RWMutex
	uploads  map[string]*multipartState
	partsDir string
}

type multipartState struct {
	upload MultipartUpload
	parts  map[int]*PartInfo
}

// NewMultipartManager creates a new multipart upload manager.
func NewMultipartManager(partsDir string) *MultipartManager {
	if err := os.MkdirAll(partsDir, 0700); err != nil {
		logging.Warn("failed to create multipart parts directory",
			"path", partsDir,
			"err", err.Error(),
			logging.Component("storage"))
	}
	return &MultipartManager{
		uploads:  make(map[string]*multipartState),
		partsDir: partsDir,
	}
}

// InitUpload creates a new multipart upload.
func (m *MultipartManager) InitUpload(_ context.Context, bucket, key, owner, contentType string) (*MultipartUpload, error) {
	id, err := generateUploadID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate upload ID: %w", err)
	}

	upload := &MultipartUpload{
		UploadID:    id,
		Bucket:      bucket,
		Key:         key,
		Owner:       owner,
		ContentType: contentType,
		CreatedAt:   time.Now(),
	}

	// Create parts directory
	dir := filepath.Join(m.partsDir, id)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("create parts dir: %w", err)
	}

	m.mu.Lock()
	m.uploads[id] = &multipartState{
		upload: *upload,
		parts:  make(map[int]*PartInfo),
	}
	m.mu.Unlock()

	return upload, nil
}

// UploadPart stores a part for a multipart upload.
func (m *MultipartManager) UploadPart(_ context.Context, uploadID string, partNumber int, body io.Reader) (*PartInfo, error) {
	m.mu.RLock()
	state, ok := m.uploads[uploadID]
	m.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("upload %q not found", uploadID)
	}

	if partNumber < 1 || partNumber > 10000 {
		return nil, fmt.Errorf("part number must be between 1 and 10000")
	}

	// Write part to temp file
	partPath := filepath.Join(m.partsDir, uploadID, fmt.Sprintf("part-%05d", partNumber))
	f, err := os.Create(partPath)
	if err != nil {
		return nil, fmt.Errorf("create part file: %w", err)
	}

	h := md5.New()
	size, err := io.Copy(io.MultiWriter(f, h), body)
	f.Close()
	if err != nil {
		os.Remove(partPath)
		return nil, fmt.Errorf("write part: %w", err)
	}

	etag := hex.EncodeToString(h.Sum(nil))
	part := &PartInfo{
		PartNumber: partNumber,
		ETag:       etag,
		Size:       size,
	}

	m.mu.Lock()
	state.parts[partNumber] = part
	m.mu.Unlock()

	return part, nil
}

// CompleteUpload assembles parts into a single object body.
// Returns the assembled content as an io.Reader and the total size.
func (m *MultipartManager) CompleteUpload(_ context.Context, uploadID string, parts []CompletedPart) (io.Reader, int64, string, error) {
	m.mu.RLock()
	state, ok := m.uploads[uploadID]
	m.mu.RUnlock()
	if !ok {
		return nil, 0, "", fmt.Errorf("upload %q not found", uploadID)
	}

	// Sort parts by number
	sort.Slice(parts, func(i, j int) bool {
		return parts[i].PartNumber < parts[j].PartNumber
	})

	// Verify all parts exist and ETags match
	var totalSize int64
	for _, cp := range parts {
		stored, exists := state.parts[cp.PartNumber]
		if !exists {
			return nil, 0, "", fmt.Errorf("part %d not found", cp.PartNumber)
		}
		if stored.ETag != cp.ETag {
			return nil, 0, "", fmt.Errorf("part %d etag mismatch: expected %q, got %q", cp.PartNumber, stored.ETag, cp.ETag)
		}
		totalSize += stored.Size
	}

	// Assemble parts into a single buffer
	var buf bytes.Buffer
	buf.Grow(int(totalSize))

	for _, cp := range parts {
		partPath := filepath.Join(m.partsDir, uploadID, fmt.Sprintf("part-%05d", cp.PartNumber))
		data, err := os.ReadFile(partPath)
		if err != nil {
			return nil, 0, "", fmt.Errorf("read part %d: %w", cp.PartNumber, err)
		}
		buf.Write(data)
	}

	contentType := state.upload.ContentType

	// Clean up parts
	m.cleanup(uploadID)

	return &buf, totalSize, contentType, nil
}

// AbortUpload cancels a multipart upload and removes its parts.
func (m *MultipartManager) AbortUpload(_ context.Context, uploadID string) error {
	m.mu.RLock()
	_, ok := m.uploads[uploadID]
	m.mu.RUnlock()
	if !ok {
		return fmt.Errorf("upload %q not found", uploadID)
	}

	m.cleanup(uploadID)
	return nil
}

// ListUploads returns all active multipart uploads for a bucket.
func (m *MultipartManager) ListUploads(bucket string) []MultipartUpload {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []MultipartUpload
	for _, state := range m.uploads {
		if state.upload.Bucket == bucket {
			result = append(result, state.upload)
		}
	}
	return result
}

// ListParts returns the parts uploaded for a multipart upload.
func (m *MultipartManager) ListParts(uploadID string) ([]PartInfo, error) {
	m.mu.RLock()
	state, ok := m.uploads[uploadID]
	m.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("upload %q not found", uploadID)
	}

	parts := make([]PartInfo, 0, len(state.parts))
	for _, p := range state.parts {
		parts = append(parts, *p)
	}
	sort.Slice(parts, func(i, j int) bool {
		return parts[i].PartNumber < parts[j].PartNumber
	})
	return parts, nil
}

func (m *MultipartManager) cleanup(uploadID string) {
	m.mu.Lock()
	delete(m.uploads, uploadID)
	m.mu.Unlock()

	dir := filepath.Join(m.partsDir, uploadID)
	os.RemoveAll(dir)
}

func generateUploadID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate upload ID: %w", err)
	}
	return hex.EncodeToString(b), nil
}
