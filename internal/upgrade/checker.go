package upgrade

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/moltbunker/moltbunker/internal/logging"
	"github.com/moltbunker/moltbunker/internal/util"
)

const (
	// defaultCheckURL is the default endpoint for version checks
	defaultCheckURL = "https://api.moltbunker.com/v1/version"

	// defaultVersionCheckInterval is the default interval between update checks
	defaultVersionCheckInterval = 24 * time.Hour

	// httpTimeout is the timeout for HTTP requests to the version endpoint
	httpTimeout = 10 * time.Second
)

// VersionInfo holds information about current and latest versions.
type VersionInfo struct {
	Current         string `json:"current"`
	Latest          string `json:"latest"`
	UpdateAvailable bool   `json:"update_available"`
	ReleaseNotes    string `json:"release_notes,omitempty"`
	DownloadURL     string `json:"download_url,omitempty"`
}

// versionResponse is the expected JSON response from the version check endpoint.
type versionResponse struct {
	Version      string `json:"version"`
	ReleaseNotes string `json:"release_notes,omitempty"`
	DownloadURL  string `json:"download_url,omitempty"`
}

// VersionChecker periodically checks for new versions of the software.
type VersionChecker struct {
	currentVersion string
	checkURL       string
	checkInterval  time.Duration
	lastCheck      time.Time
	latestVersion  string
	releaseNotes   string
	downloadURL    string
	mu             sync.RWMutex
	cancel         context.CancelFunc
	stopped        chan struct{}
	running        bool
	httpClient     *http.Client
}

// NewVersionChecker creates a new VersionChecker with the given current version.
// The currentVersion should be a semver string (e.g., "1.2.3").
func NewVersionChecker(currentVersion string) *VersionChecker {
	return &VersionChecker{
		currentVersion: currentVersion,
		checkURL:       defaultCheckURL,
		checkInterval:  defaultVersionCheckInterval,
		httpClient: &http.Client{
			Timeout: httpTimeout,
		},
	}
}

// SetCheckURL overrides the default version check endpoint URL.
func (vc *VersionChecker) SetCheckURL(url string) {
	vc.mu.Lock()
	defer vc.mu.Unlock()
	vc.checkURL = url
}

// SetCheckInterval sets how frequently background update checks occur.
func (vc *VersionChecker) SetCheckInterval(d time.Duration) {
	vc.mu.Lock()
	defer vc.mu.Unlock()
	vc.checkInterval = d
}

// CheckForUpdate checks the remote endpoint for the latest version and returns
// a VersionInfo struct describing whether an update is available.
func (vc *VersionChecker) CheckForUpdate(ctx context.Context) (*VersionInfo, error) {
	vc.mu.RLock()
	checkURL := vc.checkURL
	vc.mu.RUnlock()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, checkURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create version check request: %w", err)
	}

	req.Header.Set("User-Agent", fmt.Sprintf("moltbunker/%s", vc.currentVersion))

	resp, err := vc.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to check for updates: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("version check returned status %d", resp.StatusCode)
	}

	var vr versionResponse
	if err := json.NewDecoder(resp.Body).Decode(&vr); err != nil {
		return nil, fmt.Errorf("failed to decode version response: %w", err)
	}

	// Update cached state
	vc.mu.Lock()
	vc.latestVersion = vr.Version
	vc.releaseNotes = vr.ReleaseNotes
	vc.downloadURL = vr.DownloadURL
	vc.lastCheck = time.Now()
	vc.mu.Unlock()

	updateAvailable := CompareVersions(vc.currentVersion, vr.Version) < 0

	info := &VersionInfo{
		Current:         vc.currentVersion,
		Latest:          vr.Version,
		UpdateAvailable: updateAvailable,
		ReleaseNotes:    vr.ReleaseNotes,
		DownloadURL:     vr.DownloadURL,
	}

	if updateAvailable {
		logging.Info("update available",
			"current_version", vc.currentVersion,
			"latest_version", vr.Version,
			logging.Component("upgrade"))
	}

	return info, nil
}

// IsUpdateAvailable returns whether an update is available based on the last
// cached check. Returns false if no check has been performed yet.
func (vc *VersionChecker) IsUpdateAvailable() bool {
	vc.mu.RLock()
	defer vc.mu.RUnlock()

	if vc.latestVersion == "" {
		return false
	}

	return CompareVersions(vc.currentVersion, vc.latestVersion) < 0
}

// GetVersionInfo returns cached version information from the last check.
// Returns nil if no check has been performed yet.
func (vc *VersionChecker) GetVersionInfo() *VersionInfo {
	vc.mu.RLock()
	defer vc.mu.RUnlock()

	if vc.latestVersion == "" {
		return nil
	}

	return &VersionInfo{
		Current:         vc.currentVersion,
		Latest:          vc.latestVersion,
		UpdateAvailable: CompareVersions(vc.currentVersion, vc.latestVersion) < 0,
		ReleaseNotes:    vc.releaseNotes,
		DownloadURL:     vc.downloadURL,
	}
}

// Start begins periodic background version checks.
func (vc *VersionChecker) Start(ctx context.Context) {
	vc.mu.Lock()
	if vc.running {
		vc.mu.Unlock()
		return
	}
	vc.running = true
	vc.stopped = make(chan struct{})

	innerCtx, cancel := context.WithCancel(ctx)
	vc.cancel = cancel
	checkInterval := vc.checkInterval
	vc.mu.Unlock()

	util.SafeGoWithName("version-checker", func() {
		defer func() {
			vc.mu.Lock()
			vc.running = false
			vc.mu.Unlock()
			close(vc.stopped)
		}()

		// Perform an initial check
		if _, err := vc.CheckForUpdate(innerCtx); err != nil {
			logging.Debug("initial version check failed",
				logging.Err(err),
				logging.Component("upgrade"))
		}

		ticker := time.NewTicker(checkInterval)
		defer ticker.Stop()

		for {
			select {
			case <-innerCtx.Done():
				return
			case <-ticker.C:
				if _, err := vc.CheckForUpdate(innerCtx); err != nil {
					logging.Debug("periodic version check failed",
						logging.Err(err),
						logging.Component("upgrade"))
				}
			}
		}
	})
}

// Stop stops the background version checker and waits for the goroutine to exit.
func (vc *VersionChecker) Stop() {
	vc.mu.RLock()
	cancel := vc.cancel
	running := vc.running
	stopped := vc.stopped
	vc.mu.RUnlock()

	if !running || cancel == nil {
		return
	}

	cancel()

	if stopped != nil {
		<-stopped
	}
}

// CompareVersions compares two semver version strings.
// Returns -1 if a < b, 0 if a == b, 1 if a > b.
// Versions are expected in the form "major.minor.patch" with optional "v" prefix.
// Non-parseable versions are treated as "0.0.0".
func CompareVersions(a, b string) int {
	aParts := parseVersion(a)
	bParts := parseVersion(b)

	for i := 0; i < 3; i++ {
		if aParts[i] < bParts[i] {
			return -1
		}
		if aParts[i] > bParts[i] {
			return 1
		}
	}

	return 0
}

// parseVersion parses a version string like "v1.2.3" or "1.2.3" into [3]int.
// Missing or unparseable components default to 0.
func parseVersion(v string) [3]int {
	var result [3]int

	// Strip "v" prefix
	v = strings.TrimPrefix(v, "v")

	// Strip any pre-release or build metadata (e.g., "-beta.1", "+build.123")
	if idx := strings.IndexAny(v, "-+"); idx >= 0 {
		v = v[:idx]
	}

	parts := strings.Split(v, ".")
	for i := 0; i < 3 && i < len(parts); i++ {
		n, err := strconv.Atoi(parts[i])
		if err == nil && n >= 0 {
			result[i] = n
		}
	}

	return result
}
