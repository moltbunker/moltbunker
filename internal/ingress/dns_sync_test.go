package ingress

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestDNSSync(handler http.Handler) (*DNSSync, *httptest.Server) {
	ts := httptest.NewServer(handler)
	d := &DNSSync{
		apiToken:   "test-token",
		zoneID:     "zone123",
		ingressIP:  "1.2.3.4",
		domain:     "moltbunker.dev",
		httpClient: ts.Client(),
	}
	return d, ts
}

// cfMux builds a mock Cloudflare API. existingRecords maps FQDN → list of mock records.
func cfMux(existingRecords map[string][]cfDNSRecord, createdCount *int, deletedIDs *[]string) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// List records: GET /client/v4/zones/{zone}/dns_records
		if r.Method == http.MethodGet && strings.Contains(path, "/dns_records") && !strings.Contains(path[strings.LastIndex(path, "/dns_records")+len("/dns_records"):], "/") {
			name := r.URL.Query().Get("name")
			records := existingRecords[name]
			if records == nil {
				records = []cfDNSRecord{}
			}
			result, _ := json.Marshal(records)
			resp := cfAPIResponse{Success: true, Result: result}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
			return
		}

		// Create record: POST /client/v4/zones/{zone}/dns_records
		if r.Method == http.MethodPost && strings.Contains(path, "/dns_records") {
			if createdCount != nil {
				*createdCount++
			}
			result, _ := json.Marshal(cfDNSRecord{ID: "new-record-id"})
			resp := cfAPIResponse{Success: true, Result: result}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
			return
		}

		// Delete record: DELETE /client/v4/zones/{zone}/dns_records/{id}
		if r.Method == http.MethodDelete && strings.Contains(path, "/dns_records/") {
			parts := strings.Split(path, "/dns_records/")
			if len(parts) == 2 && deletedIDs != nil {
				*deletedIDs = append(*deletedIDs, parts[1])
			}
			result, _ := json.Marshal(map[string]string{"id": "deleted"})
			resp := cfAPIResponse{Success: true, Result: result}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
			return
		}

		http.NotFound(w, r)
	})
	return mux
}

func TestDNSSync_CreateRecord_New(t *testing.T) {
	var created int
	handler := cfMux(nil, &created, nil)
	d, ts := newTestDNSSync(handler)
	defer ts.Close()

	// Override base URL by replacing httpClient transport
	d.httpClient = &http.Client{
		Transport: &rewriteTransport{base: ts.URL, rt: http.DefaultTransport},
	}

	err := d.CreateRecord(context.Background(), "myapp")
	if err != nil {
		t.Fatalf("CreateRecord failed: %v", err)
	}
	if created != 1 {
		t.Errorf("expected 1 create call, got %d", created)
	}
}

func TestDNSSync_CreateRecord_Idempotent(t *testing.T) {
	var created int
	existing := map[string][]cfDNSRecord{
		"myapp.moltbunker.dev": {
			{ID: "rec1", Type: "A", Name: "myapp.moltbunker.dev", Content: "1.2.3.4", TTL: 300},
		},
	}
	handler := cfMux(existing, &created, nil)
	d, ts := newTestDNSSync(handler)
	defer ts.Close()
	d.httpClient = &http.Client{
		Transport: &rewriteTransport{base: ts.URL, rt: http.DefaultTransport},
	}

	err := d.CreateRecord(context.Background(), "myapp")
	if err != nil {
		t.Fatalf("CreateRecord failed: %v", err)
	}
	if created != 0 {
		t.Errorf("expected 0 create calls (idempotent), got %d", created)
	}
}

func TestDNSSync_DeleteRecord_Exists(t *testing.T) {
	var deleted []string
	existing := map[string][]cfDNSRecord{
		"myapp.moltbunker.dev": {
			{ID: "rec1", Type: "A", Name: "myapp.moltbunker.dev", Content: "1.2.3.4"},
		},
	}
	handler := cfMux(existing, nil, &deleted)
	d, ts := newTestDNSSync(handler)
	defer ts.Close()
	d.httpClient = &http.Client{
		Transport: &rewriteTransport{base: ts.URL, rt: http.DefaultTransport},
	}

	err := d.DeleteRecord(context.Background(), "myapp")
	if err != nil {
		t.Fatalf("DeleteRecord failed: %v", err)
	}
	if len(deleted) != 1 || deleted[0] != "rec1" {
		t.Errorf("expected [rec1] deleted, got %v", deleted)
	}
}

func TestDNSSync_DeleteRecord_NotExists(t *testing.T) {
	var deleted []string
	handler := cfMux(nil, nil, &deleted)
	d, ts := newTestDNSSync(handler)
	defer ts.Close()
	d.httpClient = &http.Client{
		Transport: &rewriteTransport{base: ts.URL, rt: http.DefaultTransport},
	}

	err := d.DeleteRecord(context.Background(), "noexist")
	if err != nil {
		t.Fatalf("DeleteRecord failed: %v", err)
	}
	if len(deleted) != 0 {
		t.Errorf("expected 0 delete calls, got %d", len(deleted))
	}
}

func TestDNSSync_APIError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := cfAPIResponse{
			Success: false,
			Errors:  []cfAPIError{{Code: 1003, Message: "Invalid zone"}},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	d, ts := newTestDNSSync(handler)
	defer ts.Close()
	d.httpClient = &http.Client{
		Transport: &rewriteTransport{base: ts.URL, rt: http.DefaultTransport},
	}

	err := d.CreateRecord(context.Background(), "myapp")
	if err == nil {
		t.Fatal("expected error from API error response")
	}
	if !strings.Contains(err.Error(), "cloudflare API error") {
		t.Errorf("unexpected error: %v", err)
	}
}

// rewriteTransport rewrites the URL to point to the test server.
type rewriteTransport struct {
	base string
	rt   http.RoundTripper
}

func (t *rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	req.URL.Scheme = "http"
	req.URL.Host = strings.TrimPrefix(t.base, "http://")
	return t.rt.RoundTrip(req)
}
