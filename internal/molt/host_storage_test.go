package molt

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/moltbunker/moltbunker/internal/state"
	"github.com/moltbunker/moltbunker/internal/storage"
)

// newTestStorageServices creates a HostServices with a real StorageEngine backed by MemoryStore.
func newTestStorageServices(t *testing.T) *HostServices {
	t.Helper()
	dataDir := filepath.Join(t.TempDir(), "storage")
	store := state.NewMemoryStore()
	engine, err := storage.NewStorageEngine(dataDir, store, storage.DefaultEngineConfig())
	if err != nil {
		t.Fatalf("NewStorageEngine: %v", err)
	}

	// Create test bucket
	_, err = engine.CreateBucket(context.Background(), "test-bucket", "0xTestOwner")
	if err != nil {
		t.Fatalf("CreateBucket: %v", err)
	}

	svc := NewHostServices(HostCapabilities{
		StorageEnabled: true,
		StorageBucket:  "test-bucket",
	})
	svc.Storage = engine
	svc.Owner = "0xTestOwner"
	return svc
}

func TestStoragePutGetRoundtrip(t *testing.T) {
	svc := newTestStorageServices(t)
	ctx := context.Background()

	// Put an object
	input := &storage.PutObjectInput{
		Bucket:      "test-bucket",
		Key:         "hello.txt",
		Body:        bytes.NewReader([]byte("hello world")),
		ContentType: "text/plain",
		Owner:       svc.Owner,
		Size:        11,
	}
	info, err := svc.Storage.PutObject(ctx, input)
	if err != nil {
		t.Fatalf("PutObject: %v", err)
	}
	if info.Key != "hello.txt" {
		t.Fatalf("Key = %q, want %q", info.Key, "hello.txt")
	}

	// Get it back
	output, err := svc.Storage.GetObject(ctx, "test-bucket", "hello.txt", svc.Owner)
	if err != nil {
		t.Fatalf("GetObject: %v", err)
	}
	defer output.Body.Close()

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(output.Body); err != nil {
		t.Fatalf("ReadFrom: %v", err)
	}
	if buf.String() != "hello world" {
		t.Fatalf("body = %q, want %q", buf.String(), "hello world")
	}
}

func TestStorageRequest_BucketScopeEnforcement(t *testing.T) {
	// Bucket scope: only "test-bucket" is allowed
	err := enforceBucketScope("test-bucket", "test-bucket")
	if err != nil {
		t.Fatalf("unexpected error for matching bucket: %v", err)
	}

	err = enforceBucketScope("other-bucket", "test-bucket")
	if err == nil {
		t.Fatal("expected error for mismatched bucket")
	}

	err = enforceBucketScope("test-bucket", "")
	if err == nil {
		t.Fatal("expected error for empty scoped bucket")
	}
}

func TestStorageRequest_JSONParsing(t *testing.T) {
	req := storageRequest{
		Bucket:      "my-bucket",
		Key:         "data/file.json",
		Body:        base64.StdEncoding.EncodeToString([]byte(`{"data":true}`)),
		ContentType: "application/json",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var parsed storageRequest
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if parsed.Bucket != "my-bucket" || parsed.Key != "data/file.json" {
		t.Fatalf("roundtrip failed: %+v", parsed)
	}
}

func TestStorageListObjects(t *testing.T) {
	svc := newTestStorageServices(t)
	ctx := context.Background()

	// Put a few objects
	for _, key := range []string{"a.txt", "b.txt", "dir/c.txt"} {
		_, err := svc.Storage.PutObject(ctx, &storage.PutObjectInput{
			Bucket:      "test-bucket",
			Key:         key,
			Body:        bytes.NewReader([]byte("content")),
			ContentType: "text/plain",
			Owner:       svc.Owner,
			Size:        7,
		})
		if err != nil {
			t.Fatalf("PutObject(%s): %v", key, err)
		}
	}

	// List all
	output, err := svc.Storage.ListObjects(ctx, &storage.ListObjectsInput{
		Bucket:  "test-bucket",
		MaxKeys: 100,
		Owner:   svc.Owner,
	})
	if err != nil {
		t.Fatalf("ListObjects: %v", err)
	}

	if output.KeyCount < 3 {
		t.Fatalf("KeyCount = %d, want >= 3", output.KeyCount)
	}
}

func TestStorageDeleteObject(t *testing.T) {
	svc := newTestStorageServices(t)
	ctx := context.Background()

	// Put then delete
	_, err := svc.Storage.PutObject(ctx, &storage.PutObjectInput{
		Bucket:      "test-bucket",
		Key:         "delete-me.txt",
		Body:        bytes.NewReader([]byte("temp")),
		ContentType: "text/plain",
		Owner:       svc.Owner,
		Size:        4,
	})
	if err != nil {
		t.Fatalf("PutObject: %v", err)
	}

	err = svc.Storage.DeleteObject(ctx, "test-bucket", "delete-me.txt", svc.Owner)
	if err != nil {
		t.Fatalf("DeleteObject: %v", err)
	}

	// Get should fail
	_, err = svc.Storage.GetObject(ctx, "test-bucket", "delete-me.txt", svc.Owner)
	if err == nil {
		t.Fatal("expected error getting deleted object")
	}
}

func TestStoragePut_BodySizeLimit(t *testing.T) {
	// Verify the constant is set correctly
	if storageMaxBodyBytes != 10*1024*1024 {
		t.Fatalf("storageMaxBodyBytes = %d, want 10MB", storageMaxBodyBytes)
	}
}
