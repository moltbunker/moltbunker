package upgrade

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		a    string
		b    string
		want int
	}{
		// Equal versions
		{"1.0.0", "1.0.0", 0},
		{"0.0.0", "0.0.0", 0},
		{"v1.0.0", "1.0.0", 0},

		// Major version differences
		{"1.0.0", "2.0.0", -1},
		{"2.0.0", "1.0.0", 1},
		{"10.0.0", "2.0.0", 1},

		// Minor version differences
		{"1.1.0", "1.2.0", -1},
		{"1.2.0", "1.1.0", 1},
		{"1.10.0", "1.2.0", 1},

		// Patch version differences
		{"1.0.1", "1.0.2", -1},
		{"1.0.2", "1.0.1", 1},
		{"1.0.10", "1.0.2", 1},

		// Mixed differences
		{"1.2.3", "1.2.4", -1},
		{"1.2.3", "1.3.0", -1},
		{"1.2.3", "2.0.0", -1},
		{"2.0.0", "1.9.9", 1},

		// With v prefix
		{"v1.2.3", "v1.2.4", -1},
		{"v2.0.0", "v1.0.0", 1},
		{"v1.0.0", "v1.0.0", 0},

		// Partial versions
		{"1", "1.0.0", 0},
		{"1.2", "1.2.0", 0},
		{"1", "2", -1},

		// Pre-release suffixes (stripped for comparison)
		{"1.0.0-beta", "1.0.0", 0},
		{"1.0.0-alpha", "1.0.0-beta", 0},
		{"1.0.0+build.1", "1.0.0", 0},

		// Empty or invalid
		{"", "", 0},
		{"invalid", "1.0.0", -1},
		{"1.0.0", "invalid", 1},
	}

	for _, tt := range tests {
		t.Run(tt.a+"_vs_"+tt.b, func(t *testing.T) {
			got := CompareVersions(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("CompareVersions(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestCheckForUpdate_WithMockServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify User-Agent header
		ua := r.Header.Get("User-Agent")
		if ua != "moltbunker/1.0.0" {
			t.Errorf("Expected User-Agent 'moltbunker/1.0.0', got %q", ua)
		}

		resp := versionResponse{
			Version:      "1.1.0",
			ReleaseNotes: "Bug fixes and improvements",
			DownloadURL:  "https://releases.moltbunker.com/v1.1.0",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	vc := NewVersionChecker("1.0.0")
	vc.SetCheckURL(server.URL)

	ctx := context.Background()
	info, err := vc.CheckForUpdate(ctx)
	if err != nil {
		t.Fatalf("CheckForUpdate failed: %v", err)
	}

	if info.Current != "1.0.0" {
		t.Errorf("Expected current version '1.0.0', got %q", info.Current)
	}
	if info.Latest != "1.1.0" {
		t.Errorf("Expected latest version '1.1.0', got %q", info.Latest)
	}
	if !info.UpdateAvailable {
		t.Error("Expected update to be available")
	}
	if info.ReleaseNotes != "Bug fixes and improvements" {
		t.Errorf("Unexpected release notes: %q", info.ReleaseNotes)
	}
	if info.DownloadURL != "https://releases.moltbunker.com/v1.1.0" {
		t.Errorf("Unexpected download URL: %q", info.DownloadURL)
	}
}

func TestCheckForUpdate_NoUpdateAvailable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := versionResponse{
			Version: "1.0.0",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	vc := NewVersionChecker("1.0.0")
	vc.SetCheckURL(server.URL)

	ctx := context.Background()
	info, err := vc.CheckForUpdate(ctx)
	if err != nil {
		t.Fatalf("CheckForUpdate failed: %v", err)
	}

	if info.UpdateAvailable {
		t.Error("Expected no update available when versions are equal")
	}
}

func TestCheckForUpdate_CurrentVersionNewer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := versionResponse{
			Version: "0.9.0",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	vc := NewVersionChecker("1.0.0")
	vc.SetCheckURL(server.URL)

	ctx := context.Background()
	info, err := vc.CheckForUpdate(ctx)
	if err != nil {
		t.Fatalf("CheckForUpdate failed: %v", err)
	}

	if info.UpdateAvailable {
		t.Error("Expected no update when current version is newer than latest")
	}
}

func TestCheckForUpdate_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	vc := NewVersionChecker("1.0.0")
	vc.SetCheckURL(server.URL)

	ctx := context.Background()
	_, err := vc.CheckForUpdate(ctx)
	if err == nil {
		t.Error("Expected error on server error response")
	}
}

func TestCheckForUpdate_ContextCancelled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second) // Slow server
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	vc := NewVersionChecker("1.0.0")
	vc.SetCheckURL(server.URL)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := vc.CheckForUpdate(ctx)
	if err == nil {
		t.Error("Expected error when context is cancelled")
	}
}

func TestCheckForUpdate_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("not json"))
	}))
	defer server.Close()

	vc := NewVersionChecker("1.0.0")
	vc.SetCheckURL(server.URL)

	ctx := context.Background()
	_, err := vc.CheckForUpdate(ctx)
	if err == nil {
		t.Error("Expected error on invalid JSON response")
	}
}

func TestIsUpdateAvailable_Caching(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := versionResponse{
			Version: "2.0.0",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	vc := NewVersionChecker("1.0.0")
	vc.SetCheckURL(server.URL)

	// Before any check, should return false
	if vc.IsUpdateAvailable() {
		t.Error("Expected false before any check is performed")
	}

	// Perform a check
	ctx := context.Background()
	_, err := vc.CheckForUpdate(ctx)
	if err != nil {
		t.Fatalf("CheckForUpdate failed: %v", err)
	}

	// After check, should return true (cached)
	if !vc.IsUpdateAvailable() {
		t.Error("Expected true after successful check showing newer version")
	}
}

func TestGetVersionInfo_NoCheck(t *testing.T) {
	vc := NewVersionChecker("1.0.0")

	info := vc.GetVersionInfo()
	if info != nil {
		t.Error("Expected nil VersionInfo before any check")
	}
}

func TestGetVersionInfo_AfterCheck(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := versionResponse{
			Version:      "1.5.0",
			ReleaseNotes: "New features",
			DownloadURL:  "https://example.com/download",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	vc := NewVersionChecker("1.0.0")
	vc.SetCheckURL(server.URL)

	ctx := context.Background()
	_, err := vc.CheckForUpdate(ctx)
	if err != nil {
		t.Fatalf("CheckForUpdate failed: %v", err)
	}

	info := vc.GetVersionInfo()
	if info == nil {
		t.Fatal("Expected non-nil VersionInfo after check")
	}

	if info.Current != "1.0.0" {
		t.Errorf("Expected current '1.0.0', got %q", info.Current)
	}
	if info.Latest != "1.5.0" {
		t.Errorf("Expected latest '1.5.0', got %q", info.Latest)
	}
	if !info.UpdateAvailable {
		t.Error("Expected update to be available")
	}
	if info.ReleaseNotes != "New features" {
		t.Errorf("Expected release notes 'New features', got %q", info.ReleaseNotes)
	}
	if info.DownloadURL != "https://example.com/download" {
		t.Errorf("Expected download URL 'https://example.com/download', got %q", info.DownloadURL)
	}
}

func TestStartStop(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		resp := versionResponse{
			Version: "1.0.0",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	vc := NewVersionChecker("1.0.0")
	vc.SetCheckURL(server.URL)
	vc.SetCheckInterval(50 * time.Millisecond)

	ctx := context.Background()
	vc.Start(ctx)

	// Wait for initial check + at least one periodic check
	time.Sleep(200 * time.Millisecond)

	vc.Stop()

	if requestCount < 2 {
		t.Errorf("Expected at least 2 requests (initial + periodic), got %d", requestCount)
	}

	// Verify Stop is idempotent
	vc.Stop()
}

func TestStartIdempotent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := versionResponse{Version: "1.0.0"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	vc := NewVersionChecker("1.0.0")
	vc.SetCheckURL(server.URL)
	vc.SetCheckInterval(50 * time.Millisecond)

	ctx := context.Background()
	vc.Start(ctx)
	vc.Start(ctx) // Should not start a second goroutine

	time.Sleep(100 * time.Millisecond)
	vc.Stop()
}

func TestParseVersion(t *testing.T) {
	tests := []struct {
		input string
		want  [3]int
	}{
		{"1.2.3", [3]int{1, 2, 3}},
		{"v1.2.3", [3]int{1, 2, 3}},
		{"0.0.0", [3]int{0, 0, 0}},
		{"10.20.30", [3]int{10, 20, 30}},
		{"1.2.3-beta.1", [3]int{1, 2, 3}},
		{"1.2.3+build.456", [3]int{1, 2, 3}},
		{"1.2", [3]int{1, 2, 0}},
		{"1", [3]int{1, 0, 0}},
		{"", [3]int{0, 0, 0}},
		{"invalid", [3]int{0, 0, 0}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := parseVersion(tt.input)
			if got != tt.want {
				t.Errorf("parseVersion(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
