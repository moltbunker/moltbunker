package storage

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/moltbunker/moltbunker/internal/state"
)

func newTestServer(t *testing.T) (*RESTHandler, *http.ServeMux) {
	t.Helper()
	store := state.NewMemoryStore()
	dataDir := t.TempDir()
	engine, err := NewStorageEngine(dataDir, store, DefaultEngineConfig())
	if err != nil {
		t.Fatalf("NewStorageEngine: %v", err)
	}

	handler := NewRESTHandler(engine)
	mux := http.NewServeMux()

	// Simple pass-through wrappers (no auth in tests)
	pass := func(h http.HandlerFunc) http.HandlerFunc { return h }
	handler.RegisterRoutes(mux, pass, pass)

	return handler, mux
}

func doRequest(mux http.Handler, method, path string, body io.Reader, headers map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, body)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	if _, ok := headers["X-Moltbunker-Verified-Wallet"]; !ok {
		req.Header.Set("X-Moltbunker-Verified-Wallet", testOwner)
	}
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	return w
}

func TestREST_CreateListBuckets(t *testing.T) {
	_, mux := newTestServer(t)

	// Create bucket
	body := `{"name": "test-bucket"}`
	w := doRequest(mux, "POST", "/v1/storage/buckets", strings.NewReader(body), map[string]string{
		"Content-Type": "application/json",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("create: status %d, body: %s", w.Code, w.Body.String())
	}

	// List buckets
	w = doRequest(mux, "GET", "/v1/storage/buckets", nil, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("list: status %d", w.Code)
	}

	var listResp struct {
		Buckets []BucketInfo `json:"buckets"`
		Count   int          `json:"count"`
	}
	if err := json.NewDecoder(w.Body).Decode(&listResp); err != nil {
		t.Fatalf("decode list response: %v", err)
	}
	if listResp.Count != 1 {
		t.Errorf("count = %d, want 1", listResp.Count)
	}
	if listResp.Buckets[0].Name != "test-bucket" {
		t.Errorf("bucket name = %q", listResp.Buckets[0].Name)
	}
}

func TestREST_HeadDeleteBucket(t *testing.T) {
	_, mux := newTestServer(t)

	// Create
	w := doRequest(mux, "POST", "/v1/storage/buckets", strings.NewReader(`{"name":"head-bucket"}`), map[string]string{"Content-Type": "application/json"})
	if w.Code != http.StatusCreated {
		t.Fatalf("create: status %d, body: %s", w.Code, w.Body.String())
	}

	// Head
	w = doRequest(mux, "GET", "/v1/storage/buckets/head-bucket", nil, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("head: status %d, body: %s", w.Code, w.Body.String())
	}

	// Delete
	w = doRequest(mux, "DELETE", "/v1/storage/buckets/head-bucket", nil, nil)
	if w.Code != http.StatusNoContent {
		t.Fatalf("delete: status %d, body: %s", w.Code, w.Body.String())
	}

	// Verify gone
	w = doRequest(mux, "GET", "/v1/storage/buckets/head-bucket", nil, nil)
	if w.Code != http.StatusNotFound {
		t.Errorf("after delete: status %d, want 404", w.Code)
	}
}

func TestREST_DuplicateBucket(t *testing.T) {
	_, mux := newTestServer(t)

	body := `{"name":"dup-test"}`
	doRequest(mux, "POST", "/v1/storage/buckets", strings.NewReader(body), map[string]string{"Content-Type": "application/json"})

	w := doRequest(mux, "POST", "/v1/storage/buckets", strings.NewReader(body), map[string]string{"Content-Type": "application/json"})
	if w.Code != http.StatusConflict {
		t.Errorf("duplicate: status %d, want 409", w.Code)
	}
}

func TestREST_PutGetObject(t *testing.T) {
	_, mux := newTestServer(t)

	// Create bucket first
	doRequest(mux, "POST", "/v1/storage/buckets", strings.NewReader(`{"name":"data"}`), map[string]string{"Content-Type": "application/json"})

	// Put object
	content := "hello storage API"
	w := doRequest(mux, "PUT", "/v1/storage/objects/data/file.txt", strings.NewReader(content), map[string]string{
		"Content-Type": "text/plain",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("put: status %d, body: %s", w.Code, w.Body.String())
	}

	etag := w.Header().Get("ETag")
	if etag == "" {
		t.Error("put should return ETag header")
	}

	// Get object
	w = doRequest(mux, "GET", "/v1/storage/objects/data/file.txt", nil, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("get: status %d", w.Code)
	}
	if w.Body.String() != content {
		t.Errorf("body = %q, want %q", w.Body.String(), content)
	}
	if w.Header().Get("Content-Type") != "text/plain" {
		t.Errorf("content-type = %q", w.Header().Get("Content-Type"))
	}
}

func TestREST_HeadObject(t *testing.T) {
	_, mux := newTestServer(t)

	doRequest(mux, "POST", "/v1/storage/buckets", strings.NewReader(`{"name":"head-obj"}`), map[string]string{"Content-Type": "application/json"})
	doRequest(mux, "PUT", "/v1/storage/objects/head-obj/test.bin", bytes.NewReader(make([]byte, 42)), map[string]string{
		"Content-Type": "application/octet-stream",
	})

	w := doRequest(mux, "HEAD", "/v1/storage/objects/head-obj/test.bin", nil, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("head: status %d", w.Code)
	}
	if w.Header().Get("Content-Length") != "42" {
		t.Errorf("content-length = %q, want 42", w.Header().Get("Content-Length"))
	}
}

func TestREST_DeleteObject(t *testing.T) {
	_, mux := newTestServer(t)

	doRequest(mux, "POST", "/v1/storage/buckets", strings.NewReader(`{"name":"del-test"}`), map[string]string{"Content-Type": "application/json"})
	doRequest(mux, "PUT", "/v1/storage/objects/del-test/rm.txt", strings.NewReader("bye"), nil)

	w := doRequest(mux, "DELETE", "/v1/storage/objects/del-test/rm.txt", nil, nil)
	if w.Code != http.StatusNoContent {
		t.Fatalf("delete: status %d, body: %s", w.Code, w.Body.String())
	}

	// Verify gone
	w = doRequest(mux, "GET", "/v1/storage/objects/del-test/rm.txt", nil, nil)
	if w.Code != http.StatusNotFound {
		t.Errorf("after delete: status %d, want 404", w.Code)
	}
}

func TestREST_ListObjects(t *testing.T) {
	_, mux := newTestServer(t)

	doRequest(mux, "POST", "/v1/storage/buckets", strings.NewReader(`{"name":"list-obj"}`), map[string]string{"Content-Type": "application/json"})

	for _, k := range []string{"a.txt", "b.txt", "dir/c.txt"} {
		doRequest(mux, "PUT", "/v1/storage/objects/list-obj/"+k, strings.NewReader("data"), nil)
	}

	w := doRequest(mux, "GET", "/v1/storage/objects/list-obj/?prefix=&delimiter=", nil, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("list: status %d", w.Code)
	}

	var out ListObjectsOutput
	if err := json.NewDecoder(w.Body).Decode(&out); err != nil {
		t.Fatalf("decode list response: %v", err)
	}
	if out.KeyCount != 3 {
		t.Errorf("count = %d, want 3", out.KeyCount)
	}
}

func TestREST_ListObjectsDelimited(t *testing.T) {
	_, mux := newTestServer(t)

	doRequest(mux, "POST", "/v1/storage/buckets", strings.NewReader(`{"name":"folders"}`), map[string]string{"Content-Type": "application/json"})

	for _, k := range []string{"a.txt", "dir/b.txt", "dir/c.txt"} {
		doRequest(mux, "PUT", "/v1/storage/objects/folders/"+k, strings.NewReader("data"), nil)
	}

	w := doRequest(mux, "GET", "/v1/storage/objects/folders/?delimiter=/", nil, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("list: status %d", w.Code)
	}

	var out ListObjectsOutput
	if err := json.NewDecoder(w.Body).Decode(&out); err != nil {
		t.Fatalf("decode list response: %v", err)
	}
	if out.KeyCount != 1 { // only a.txt at root level
		t.Errorf("count = %d, want 1", out.KeyCount)
	}
	if len(out.CommonPrefixes) != 1 { // dir/
		t.Errorf("common prefixes = %v, want [dir/]", out.CommonPrefixes)
	}
}

func TestREST_Usage(t *testing.T) {
	_, mux := newTestServer(t)

	doRequest(mux, "POST", "/v1/storage/buckets", strings.NewReader(`{"name":"usage"}`), map[string]string{"Content-Type": "application/json"})
	doRequest(mux, "PUT", "/v1/storage/objects/usage/f1.txt", strings.NewReader("12345"), nil)

	w := doRequest(mux, "GET", "/v1/storage/usage", nil, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("usage: status %d", w.Code)
	}

	var report UsageReport
	if err := json.NewDecoder(w.Body).Decode(&report); err != nil {
		t.Fatalf("decode usage report: %v", err)
	}
	if report.ObjectCount != 1 {
		t.Errorf("object count = %d, want 1", report.ObjectCount)
	}
	if report.TotalBytes != 5 {
		t.Errorf("total bytes = %d, want 5", report.TotalBytes)
	}
}

func TestREST_NotFoundBucket(t *testing.T) {
	_, mux := newTestServer(t)

	w := doRequest(mux, "GET", "/v1/storage/buckets/nonexistent", nil, nil)
	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
}

func TestREST_NotFoundObject(t *testing.T) {
	_, mux := newTestServer(t)

	doRequest(mux, "POST", "/v1/storage/buckets", strings.NewReader(`{"name":"empty"}`), map[string]string{"Content-Type": "application/json"})

	w := doRequest(mux, "GET", "/v1/storage/objects/empty/nope.txt", nil, nil)
	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
}

func TestREST_NestedKey(t *testing.T) {
	_, mux := newTestServer(t)

	doRequest(mux, "POST", "/v1/storage/buckets", strings.NewReader(`{"name":"nested"}`), map[string]string{"Content-Type": "application/json"})

	// Deep nested key
	w := doRequest(mux, "PUT", "/v1/storage/objects/nested/a/b/c/d.txt", strings.NewReader("deep"), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("put nested: status %d, body: %s", w.Code, w.Body.String())
	}

	w = doRequest(mux, "GET", "/v1/storage/objects/nested/a/b/c/d.txt", nil, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("get nested: status %d", w.Code)
	}
	if w.Body.String() != "deep" {
		t.Errorf("body = %q, want %q", w.Body.String(), "deep")
	}
}
