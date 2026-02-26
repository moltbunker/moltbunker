package storage

import (
	"bytes"
	"context"
	"io"
	"testing"
)

func TestMultipart_InitUpload(t *testing.T) {
	mgr := NewMultipartManager(t.TempDir())
	ctx := context.Background()

	upload, err := mgr.InitUpload(ctx, "test-bucket", "big-file.zip", "wallet1", "application/zip")
	if err != nil {
		t.Fatalf("InitUpload: %v", err)
	}
	if upload.UploadID == "" {
		t.Error("upload ID should not be empty")
	}
	if upload.Bucket != "test-bucket" {
		t.Errorf("bucket = %q, want %q", upload.Bucket, "test-bucket")
	}
	if upload.Key != "big-file.zip" {
		t.Errorf("key = %q, want %q", upload.Key, "big-file.zip")
	}
}

func TestMultipart_UploadPartAndComplete(t *testing.T) {
	mgr := NewMultipartManager(t.TempDir())
	ctx := context.Background()

	upload, err := mgr.InitUpload(ctx, "test-bucket", "file.dat", "w1", "application/octet-stream")
	if err != nil {
		t.Fatalf("InitUpload: %v", err)
	}

	// Upload 3 parts
	part1Data := []byte("hello ")
	part2Data := []byte("multipart ")
	part3Data := []byte("world")

	p1, err := mgr.UploadPart(ctx, upload.UploadID, 1, bytes.NewReader(part1Data))
	if err != nil {
		t.Fatalf("UploadPart 1: %v", err)
	}
	p2, err := mgr.UploadPart(ctx, upload.UploadID, 2, bytes.NewReader(part2Data))
	if err != nil {
		t.Fatalf("UploadPart 2: %v", err)
	}
	p3, err := mgr.UploadPart(ctx, upload.UploadID, 3, bytes.NewReader(part3Data))
	if err != nil {
		t.Fatalf("UploadPart 3: %v", err)
	}

	// Complete upload
	reader, size, contentType, err := mgr.CompleteUpload(ctx, upload.UploadID, []CompletedPart{
		{PartNumber: 1, ETag: p1.ETag},
		{PartNumber: 2, ETag: p2.ETag},
		{PartNumber: 3, ETag: p3.ETag},
	})
	if err != nil {
		t.Fatalf("CompleteUpload: %v", err)
	}

	expected := "hello multipart world"
	if size != int64(len(expected)) {
		t.Errorf("size = %d, want %d", size, len(expected))
	}
	if contentType != "application/octet-stream" {
		t.Errorf("content-type = %q", contentType)
	}

	data, _ := io.ReadAll(reader)
	if string(data) != expected {
		t.Errorf("assembled = %q, want %q", string(data), expected)
	}
}

func TestMultipart_ListParts(t *testing.T) {
	mgr := NewMultipartManager(t.TempDir())
	ctx := context.Background()

	upload, _ := mgr.InitUpload(ctx, "bucket", "key", "w1", "")
	mgr.UploadPart(ctx, upload.UploadID, 3, bytes.NewReader([]byte("c")))
	mgr.UploadPart(ctx, upload.UploadID, 1, bytes.NewReader([]byte("a")))

	parts, err := mgr.ListParts(upload.UploadID)
	if err != nil {
		t.Fatalf("ListParts: %v", err)
	}
	if len(parts) != 2 {
		t.Fatalf("parts count = %d, want 2", len(parts))
	}
	// Should be sorted by part number
	if parts[0].PartNumber != 1 {
		t.Errorf("first part number = %d, want 1", parts[0].PartNumber)
	}
	if parts[1].PartNumber != 3 {
		t.Errorf("second part number = %d, want 3", parts[1].PartNumber)
	}
}

func TestMultipart_AbortUpload(t *testing.T) {
	mgr := NewMultipartManager(t.TempDir())
	ctx := context.Background()

	upload, _ := mgr.InitUpload(ctx, "bucket", "key", "w1", "")
	mgr.UploadPart(ctx, upload.UploadID, 1, bytes.NewReader([]byte("data")))

	if err := mgr.AbortUpload(ctx, upload.UploadID); err != nil {
		t.Fatalf("AbortUpload: %v", err)
	}

	// Should be gone
	_, err := mgr.ListParts(upload.UploadID)
	if err == nil {
		t.Error("listing parts after abort should fail")
	}
}

func TestMultipart_ListUploads(t *testing.T) {
	mgr := NewMultipartManager(t.TempDir())
	ctx := context.Background()

	mgr.InitUpload(ctx, "bucket-a", "key1", "w1", "")
	mgr.InitUpload(ctx, "bucket-a", "key2", "w1", "")
	mgr.InitUpload(ctx, "bucket-b", "key3", "w1", "")

	uploads := mgr.ListUploads("bucket-a")
	if len(uploads) != 2 {
		t.Errorf("bucket-a uploads = %d, want 2", len(uploads))
	}

	uploads = mgr.ListUploads("bucket-b")
	if len(uploads) != 1 {
		t.Errorf("bucket-b uploads = %d, want 1", len(uploads))
	}
}

func TestMultipart_ETagMismatch(t *testing.T) {
	mgr := NewMultipartManager(t.TempDir())
	ctx := context.Background()

	upload, _ := mgr.InitUpload(ctx, "bucket", "key", "w1", "")
	mgr.UploadPart(ctx, upload.UploadID, 1, bytes.NewReader([]byte("data")))

	_, _, _, err := mgr.CompleteUpload(ctx, upload.UploadID, []CompletedPart{
		{PartNumber: 1, ETag: "wrong-etag"},
	})
	if err == nil {
		t.Error("complete with wrong etag should fail")
	}
}

func TestMultipart_InvalidPartNumber(t *testing.T) {
	mgr := NewMultipartManager(t.TempDir())
	ctx := context.Background()

	upload, _ := mgr.InitUpload(ctx, "bucket", "key", "w1", "")

	_, err := mgr.UploadPart(ctx, upload.UploadID, 0, bytes.NewReader([]byte("data")))
	if err == nil {
		t.Error("part number 0 should fail")
	}

	_, err = mgr.UploadPart(ctx, upload.UploadID, 10001, bytes.NewReader([]byte("data")))
	if err == nil {
		t.Error("part number 10001 should fail")
	}
}

func TestMultipart_NotFound(t *testing.T) {
	mgr := NewMultipartManager(t.TempDir())
	ctx := context.Background()

	_, err := mgr.UploadPart(ctx, "nonexistent", 1, bytes.NewReader([]byte("data")))
	if err == nil {
		t.Error("uploading to nonexistent upload should fail")
	}

	err = mgr.AbortUpload(ctx, "nonexistent")
	if err == nil {
		t.Error("aborting nonexistent upload should fail")
	}
}
