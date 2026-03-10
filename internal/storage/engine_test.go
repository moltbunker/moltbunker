package storage

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/moltbunker/moltbunker/internal/state"
)

func newTestEngine(t *testing.T) *StorageEngine {
	t.Helper()
	store := state.NewMemoryStore()
	dataDir := t.TempDir()
	engine, err := NewStorageEngine(dataDir, store, DefaultEngineConfig())
	if err != nil {
		t.Fatalf("NewStorageEngine: %v", err)
	}
	return engine
}

const testOwner = "0x1234567890abcdef1234567890abcdef12345678"

func TestCreateBucket(t *testing.T) {
	ctx := context.Background()
	e := newTestEngine(t)

	bucket, err := e.CreateBucket(ctx, "test-bucket", testOwner)
	if err != nil {
		t.Fatalf("CreateBucket: %v", err)
	}
	if bucket.Name != "test-bucket" {
		t.Errorf("name = %q, want %q", bucket.Name, "test-bucket")
	}
	if bucket.Owner != testOwner {
		t.Errorf("owner = %q, want %q", bucket.Owner, testOwner)
	}
}

func TestCreateBucket_InvalidName(t *testing.T) {
	ctx := context.Background()
	e := newTestEngine(t)

	tests := []struct {
		name string
	}{
		{"ab"},                 // too short
		{"AB"},                 // uppercase
		{"-bucket"},            // starts with hyphen
		{"bucket-"},            // ends with hyphen
		{"bucket name"},        // space
		{strings.Repeat("a", 64)}, // too long
	}

	for _, tt := range tests {
		_, err := e.CreateBucket(ctx, tt.name, testOwner)
		if err == nil {
			t.Errorf("CreateBucket(%q) should fail", tt.name)
		}
	}
}

func TestCreateBucket_Duplicate(t *testing.T) {
	ctx := context.Background()
	e := newTestEngine(t)

	if _, err := e.CreateBucket(ctx, "my-bucket", testOwner); err != nil {
		t.Fatalf("first create: %v", err)
	}

	_, err := e.CreateBucket(ctx, "my-bucket", testOwner)
	if err == nil {
		t.Fatal("duplicate create should fail")
	}
}

func TestDeleteBucket(t *testing.T) {
	ctx := context.Background()
	e := newTestEngine(t)

	if _, err := e.CreateBucket(ctx, "del-bucket", testOwner); err != nil {
		t.Fatalf("create: %v", err)
	}

	if err := e.DeleteBucket(ctx, "del-bucket", testOwner); err != nil {
		t.Fatalf("delete: %v", err)
	}

	// Should not exist anymore
	b, err := e.HeadBucket(ctx, "del-bucket", testOwner)
	if err != nil {
		t.Fatalf("head: %v", err)
	}
	if b != nil {
		t.Error("bucket still exists after delete")
	}
}

func TestDeleteBucket_NotEmpty(t *testing.T) {
	ctx := context.Background()
	e := newTestEngine(t)

	if _, err := e.CreateBucket(ctx, "notempty", testOwner); err != nil {
		t.Fatalf("create: %v", err)
	}

	// Put an object in it
	_, err := e.PutObject(ctx, &PutObjectInput{
		Bucket: "notempty",
		Key:    "file.txt",
		Body:   strings.NewReader("hello"),
		Owner:  testOwner,
	})
	if err != nil {
		t.Fatalf("put: %v", err)
	}

	err = e.DeleteBucket(ctx, "notempty", testOwner)
	if err == nil {
		t.Fatal("delete non-empty bucket should fail")
	}
}

func TestDeleteBucket_WrongOwner(t *testing.T) {
	ctx := context.Background()
	e := newTestEngine(t)

	if _, err := e.CreateBucket(ctx, "owned", testOwner); err != nil {
		t.Fatalf("create: %v", err)
	}

	err := e.DeleteBucket(ctx, "owned", "0xdifferent")
	if err == nil {
		t.Fatal("delete with wrong owner should fail")
	}
}

func TestListBuckets(t *testing.T) {
	ctx := context.Background()
	e := newTestEngine(t)

	for _, name := range []string{"alpha", "beta", "gamma"} {
		if _, err := e.CreateBucket(ctx, name, testOwner); err != nil {
			t.Fatalf("create %s: %v", name, err)
		}
	}

	buckets, err := e.ListBuckets(ctx, testOwner)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(buckets) != 3 {
		t.Fatalf("count = %d, want 3", len(buckets))
	}
	// Should be sorted
	if buckets[0].Name != "alpha" || buckets[1].Name != "beta" || buckets[2].Name != "gamma" {
		t.Errorf("unexpected order: %v", buckets)
	}
}

func TestPutGetObject(t *testing.T) {
	ctx := context.Background()
	e := newTestEngine(t)

	if _, err := e.CreateBucket(ctx, "data", testOwner); err != nil {
		t.Fatalf("create bucket: %v", err)
	}

	content := "hello, moltbunker storage!"
	obj, err := e.PutObject(ctx, &PutObjectInput{
		Bucket:      "data",
		Key:         "greeting.txt",
		Body:        strings.NewReader(content),
		ContentType: "text/plain",
		Owner:       testOwner,
	})
	if err != nil {
		t.Fatalf("put: %v", err)
	}

	if obj.Size != int64(len(content)) {
		t.Errorf("size = %d, want %d", obj.Size, len(content))
	}
	if obj.ContentType != "text/plain" {
		t.Errorf("content type = %q, want %q", obj.ContentType, "text/plain")
	}
	if obj.ETag == "" {
		t.Error("etag should not be empty")
	}

	// Get
	out, err := e.GetObject(ctx, "data", "greeting.txt", testOwner)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer out.Body.Close()

	got, err := io.ReadAll(out.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if string(got) != content {
		t.Errorf("body = %q, want %q", string(got), content)
	}
	if out.Info.ETag != obj.ETag {
		t.Errorf("etag mismatch: get=%q, put=%q", out.Info.ETag, obj.ETag)
	}
}

func TestPutObject_LargeFile(t *testing.T) {
	ctx := context.Background()
	store := state.NewMemoryStore()
	dataDir := t.TempDir()

	// Engine with 100-byte max
	engine, err := NewStorageEngine(dataDir, store, EngineConfig{
		MaxBuckets:    100,
		MaxObjectSize: 100,
	})
	if err != nil {
		t.Fatalf("NewStorageEngine: %v", err)
	}

	if _, err := engine.CreateBucket(ctx, "small", testOwner); err != nil {
		t.Fatalf("create: %v", err)
	}

	_, err = engine.PutObject(ctx, &PutObjectInput{
		Bucket: "small",
		Key:    "big.dat",
		Body:   bytes.NewReader(make([]byte, 200)),
		Owner:  testOwner,
	})
	if err == nil {
		t.Fatal("should reject oversized object")
	}
}

func TestHeadObject(t *testing.T) {
	ctx := context.Background()
	e := newTestEngine(t)

	if _, err := e.CreateBucket(ctx, "meta", testOwner); err != nil {
		t.Fatalf("create bucket: %v", err)
	}

	if _, err := e.PutObject(ctx, &PutObjectInput{
		Bucket: "meta",
		Key:    "info.json",
		Body:   strings.NewReader(`{"key":"value"}`),
		Owner:  testOwner,
	}); err != nil {
		t.Fatalf("put: %v", err)
	}

	obj, err := e.HeadObject(ctx, "meta", "info.json", testOwner)
	if err != nil {
		t.Fatalf("head: %v", err)
	}
	if obj.Key != "info.json" {
		t.Errorf("key = %q, want %q", obj.Key, "info.json")
	}
	if obj.Size != 15 {
		t.Errorf("size = %d, want 15", obj.Size)
	}
}

func TestDeleteObject(t *testing.T) {
	ctx := context.Background()
	e := newTestEngine(t)

	if _, err := e.CreateBucket(ctx, "trash", testOwner); err != nil {
		t.Fatalf("create bucket: %v", err)
	}

	if _, err := e.PutObject(ctx, &PutObjectInput{
		Bucket: "trash",
		Key:    "junk.txt",
		Body:   strings.NewReader("delete me"),
		Owner:  testOwner,
	}); err != nil {
		t.Fatalf("put: %v", err)
	}

	if err := e.DeleteObject(ctx, "trash", "junk.txt", testOwner); err != nil {
		t.Fatalf("delete: %v", err)
	}

	_, err := e.HeadObject(ctx, "trash", "junk.txt", testOwner)
	if err == nil {
		t.Fatal("object should not exist after delete")
	}
}

func TestDeleteObject_WrongOwner(t *testing.T) {
	ctx := context.Background()
	e := newTestEngine(t)

	if _, err := e.CreateBucket(ctx, "private", testOwner); err != nil {
		t.Fatalf("create: %v", err)
	}

	if _, err := e.PutObject(ctx, &PutObjectInput{
		Bucket: "private",
		Key:    "secret.txt",
		Body:   strings.NewReader("secret"),
		Owner:  testOwner,
	}); err != nil {
		t.Fatalf("put: %v", err)
	}

	err := e.DeleteObject(ctx, "private", "secret.txt", "0xdifferent")
	if err == nil {
		t.Fatal("should fail with wrong owner")
	}
}

func TestListObjects(t *testing.T) {
	ctx := context.Background()
	e := newTestEngine(t)

	if _, err := e.CreateBucket(ctx, "files", testOwner); err != nil {
		t.Fatalf("create: %v", err)
	}

	keys := []string{"a.txt", "b.txt", "dir/c.txt", "dir/d.txt", "dir/sub/e.txt"}
	for _, k := range keys {
		if _, err := e.PutObject(ctx, &PutObjectInput{
			Bucket: "files",
			Key:    k,
			Body:   strings.NewReader("data"),
			Owner:  testOwner,
		}); err != nil {
			t.Fatalf("put %s: %v", k, err)
		}
	}

	// List all
	out, err := e.ListObjects(ctx, &ListObjectsInput{Bucket: "files", Owner: testOwner})
	if err != nil {
		t.Fatalf("list all: %v", err)
	}
	if out.KeyCount != 5 {
		t.Errorf("all count = %d, want 5", out.KeyCount)
	}

	// List with prefix
	out, err = e.ListObjects(ctx, &ListObjectsInput{
		Bucket: "files",
		Prefix: "dir/",
		Owner:  testOwner,
	})
	if err != nil {
		t.Fatalf("list prefix: %v", err)
	}
	if out.KeyCount != 3 {
		t.Errorf("prefix count = %d, want 3", out.KeyCount)
	}

	// List with prefix + delimiter (folder-like)
	out, err = e.ListObjects(ctx, &ListObjectsInput{
		Bucket:    "files",
		Prefix:    "dir/",
		Delimiter: "/",
		Owner:     testOwner,
	})
	if err != nil {
		t.Fatalf("list delimited: %v", err)
	}
	if out.KeyCount != 2 { // dir/c.txt, dir/d.txt
		t.Errorf("delimited count = %d, want 2", out.KeyCount)
	}
	if len(out.CommonPrefixes) != 1 { // dir/sub/
		t.Errorf("common prefixes = %v, want [dir/sub/]", out.CommonPrefixes)
	}
}

func TestListObjects_Pagination(t *testing.T) {
	ctx := context.Background()
	e := newTestEngine(t)

	if _, err := e.CreateBucket(ctx, "paged", testOwner); err != nil {
		t.Fatalf("create: %v", err)
	}

	for i := 0; i < 5; i++ {
		key := string(rune('a'+i)) + ".txt"
		if _, err := e.PutObject(ctx, &PutObjectInput{
			Bucket: "paged",
			Key:    key,
			Body:   strings.NewReader("data"),
			Owner:  testOwner,
		}); err != nil {
			t.Fatalf("put: %v", err)
		}
	}

	// First page (max 2)
	out, err := e.ListObjects(ctx, &ListObjectsInput{
		Bucket:  "paged",
		MaxKeys: 2,
		Owner:   testOwner,
	})
	if err != nil {
		t.Fatalf("page 1: %v", err)
	}
	if out.KeyCount != 2 {
		t.Errorf("page 1 count = %d, want 2", out.KeyCount)
	}
	if !out.IsTruncated {
		t.Error("page 1 should be truncated")
	}

	// Second page
	out, err = e.ListObjects(ctx, &ListObjectsInput{
		Bucket:            "paged",
		MaxKeys:           2,
		ContinuationToken: out.NextContinuationToken,
		Owner:             testOwner,
	})
	if err != nil {
		t.Fatalf("page 2: %v", err)
	}
	if out.KeyCount != 2 {
		t.Errorf("page 2 count = %d, want 2", out.KeyCount)
	}
}

func TestPutObject_Overwrite(t *testing.T) {
	ctx := context.Background()
	e := newTestEngine(t)

	if _, err := e.CreateBucket(ctx, "overwrite", testOwner); err != nil {
		t.Fatalf("create: %v", err)
	}

	// First version
	if _, err := e.PutObject(ctx, &PutObjectInput{
		Bucket: "overwrite",
		Key:    "file.txt",
		Body:   strings.NewReader("version 1"),
		Owner:  testOwner,
	}); err != nil {
		t.Fatalf("put v1: %v", err)
	}

	// Overwrite
	obj, err := e.PutObject(ctx, &PutObjectInput{
		Bucket: "overwrite",
		Key:    "file.txt",
		Body:   strings.NewReader("version 2"),
		Owner:  testOwner,
	})
	if err != nil {
		t.Fatalf("put v2: %v", err)
	}
	if obj.Size != int64(len("version 2")) {
		t.Errorf("size = %d, want %d", obj.Size, len("version 2"))
	}

	// Read back
	out, err := e.GetObject(ctx, "overwrite", "file.txt", testOwner)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer out.Body.Close()
	got, _ := io.ReadAll(out.Body)
	if string(got) != "version 2" {
		t.Errorf("body = %q, want %q", string(got), "version 2")
	}
}

func TestGetObject_NotFound(t *testing.T) {
	ctx := context.Background()
	e := newTestEngine(t)

	if _, err := e.CreateBucket(ctx, "empty", testOwner); err != nil {
		t.Fatalf("create: %v", err)
	}

	_, err := e.GetObject(ctx, "empty", "nope.txt", testOwner)
	if err == nil {
		t.Fatal("should fail for non-existent object")
	}
}

func TestGetUsage(t *testing.T) {
	ctx := context.Background()
	e := newTestEngine(t)

	if _, err := e.CreateBucket(ctx, "usage1", testOwner); err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := e.CreateBucket(ctx, "usage2", testOwner); err != nil {
		t.Fatalf("create: %v", err)
	}

	for _, pair := range []struct{ bucket, key, data string }{
		{"usage1", "a.txt", "hello"},
		{"usage1", "b.txt", "world"},
		{"usage2", "c.txt", "12345678"},
	} {
		if _, err := e.PutObject(ctx, &PutObjectInput{
			Bucket: pair.bucket,
			Key:    pair.key,
			Body:   strings.NewReader(pair.data),
			Owner:  testOwner,
		}); err != nil {
			t.Fatalf("put: %v", err)
		}
	}

	report, err := e.GetUsage(ctx, testOwner)
	if err != nil {
		t.Fatalf("usage: %v", err)
	}
	if report.BucketCount != 2 {
		t.Errorf("bucket count = %d, want 2", report.BucketCount)
	}
	if report.ObjectCount != 3 {
		t.Errorf("object count = %d, want 3", report.ObjectCount)
	}
	if report.TotalBytes != 18 { // 5 + 5 + 8
		t.Errorf("total bytes = %d, want 18", report.TotalBytes)
	}
}

func TestPutObject_InvalidKey(t *testing.T) {
	ctx := context.Background()
	e := newTestEngine(t)

	if _, err := e.CreateBucket(ctx, "keytest", testOwner); err != nil {
		t.Fatalf("create: %v", err)
	}

	tests := []struct {
		key     string
		wantErr bool
	}{
		{"", true},           // empty
		{"../escape", true},  // path traversal
		{"valid/key.txt", false},
		{"deep/nested/path/file.dat", false},
	}

	for _, tt := range tests {
		_, err := e.PutObject(ctx, &PutObjectInput{
			Bucket: "keytest",
			Key:    tt.key,
			Body:   strings.NewReader("x"),
			Owner:  testOwner,
		})
		if tt.wantErr && err == nil {
			t.Errorf("PutObject(%q) should fail", tt.key)
		}
		if !tt.wantErr && err != nil {
			t.Errorf("PutObject(%q) unexpected error: %v", tt.key, err)
		}
	}
}

func TestBucketLimit(t *testing.T) {
	ctx := context.Background()
	store := state.NewMemoryStore()
	dataDir := t.TempDir()

	engine, err := NewStorageEngine(dataDir, store, EngineConfig{
		MaxBuckets:    3,
		MaxObjectSize: 1024,
	})
	if err != nil {
		t.Fatalf("NewStorageEngine: %v", err)
	}

	for i := 0; i < 3; i++ {
		name := "bucket-" + string(rune('a'+i))
		if _, err := engine.CreateBucket(ctx, name, testOwner); err != nil {
			t.Fatalf("create %s: %v", name, err)
		}
	}

	_, err = engine.CreateBucket(ctx, "bucket-d", testOwner)
	if err == nil {
		t.Fatal("should fail when bucket limit exceeded")
	}
}
