package tor

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// ExitNodeInfo represents information about a Tor exit node
type ExitNodeInfo struct {
	Fingerprint string
	Nickname    string
	Country     string
	CountryName string
	IP          string
	Bandwidth   int64
	Flags       []string
	LastChecked time.Time
}

// ExitNodeSelector selects Tor exit nodes by country/region
type ExitNodeSelector struct {
	client    *http.Client
	cache     map[string][]*ExitNodeInfo
	cacheMu   sync.RWMutex
	cacheTime time.Time
	cacheTTL  time.Duration
}

// OnionooResponse represents the response from onionoo.torproject.org
type OnionooResponse struct {
	Version          string         `json:"version"`
	RelaysPublished  string         `json:"relays_published"`
	Relays           []OnionooRelay `json:"relays"`
}

// OnionooRelay represents a relay from onionoo
type OnionooRelay struct {
	Nickname    string   `json:"nickname"`
	Fingerprint string   `json:"fingerprint"`
	OrAddresses []string `json:"or_addresses"`
	ExitPolicy  string   `json:"exit_policy_summary,omitempty"`
	Country     string   `json:"country"`
	CountryName string   `json:"country_name"`
	Flags       []string `json:"flags"`
	Bandwidth   int64    `json:"observed_bandwidth"`
	Running     bool     `json:"running"`
}

// NewExitNodeSelector creates a new exit node selector
func NewExitNodeSelector() *ExitNodeSelector {
	return &ExitNodeSelector{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		cache:    make(map[string][]*ExitNodeInfo),
		cacheTTL: 1 * time.Hour,
	}
}

// SelectExitNodeByCountry selects an exit node in a specific country
func (ens *ExitNodeSelector) SelectExitNodeByCountry(ctx context.Context, countryCode string) (*ExitNodeInfo, error) {
	countryCode = strings.ToLower(countryCode)

	// Check cache
	ens.cacheMu.RLock()
	if nodes, ok := ens.cache[countryCode]; ok && time.Since(ens.cacheTime) < ens.cacheTTL {
		if len(nodes) > 0 {
			ens.cacheMu.RUnlock()
			// Return first (highest bandwidth) node
			return nodes[0], nil
		}
	}
	ens.cacheMu.RUnlock()

	// Fetch from onionoo API
	nodes, err := ens.fetchExitNodes(ctx, countryCode)
	if err != nil {
		return nil, err
	}

	if len(nodes) == 0 {
		return nil, fmt.Errorf("no exit nodes found in country: %s", countryCode)
	}

	// Cache results
	ens.cacheMu.Lock()
	ens.cache[countryCode] = nodes
	ens.cacheTime = time.Now()
	ens.cacheMu.Unlock()

	return nodes[0], nil
}

// fetchExitNodes fetches exit nodes from onionoo.torproject.org
func (ens *ExitNodeSelector) fetchExitNodes(ctx context.Context, countryCode string) ([]*ExitNodeInfo, error) {
	// Build onionoo URL
	// flag:Exit - only exit relays
	// flag:Running - only running relays
	// flag:Valid - only valid relays
	url := fmt.Sprintf(
		"https://onionoo.torproject.org/details?flag=Exit&flag=Running&flag=Valid&country=%s&fields=nickname,fingerprint,or_addresses,country,country_name,flags,observed_bandwidth",
		countryCode,
	)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "moltbunker/1.0")

	resp, err := ens.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch exit nodes: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("onionoo returned status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var onionooResp OnionooResponse
	if err := json.Unmarshal(body, &onionooResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Convert to ExitNodeInfo and sort by bandwidth
	nodes := make([]*ExitNodeInfo, 0, len(onionooResp.Relays))
	for _, relay := range onionooResp.Relays {
		if !relay.Running {
			continue
		}

		// Extract IP from or_addresses
		ip := ""
		if len(relay.OrAddresses) > 0 {
			addr := relay.OrAddresses[0]
			// Format is IP:Port, extract IP
			if idx := strings.LastIndex(addr, ":"); idx > 0 {
				ip = addr[:idx]
			} else {
				ip = addr
			}
			// Handle IPv6 addresses in brackets
			ip = strings.Trim(ip, "[]")
		}

		nodes = append(nodes, &ExitNodeInfo{
			Fingerprint: relay.Fingerprint,
			Nickname:    relay.Nickname,
			Country:     relay.Country,
			CountryName: relay.CountryName,
			IP:          ip,
			Bandwidth:   relay.Bandwidth,
			Flags:       relay.Flags,
			LastChecked: time.Now(),
		})
	}

	// Sort by bandwidth (highest first)
	sortByBandwidth(nodes)

	return nodes, nil
}

// sortByBandwidth sorts nodes by bandwidth descending
func sortByBandwidth(nodes []*ExitNodeInfo) {
	for i := 0; i < len(nodes)-1; i++ {
		for j := i + 1; j < len(nodes); j++ {
			if nodes[j].Bandwidth > nodes[i].Bandwidth {
				nodes[i], nodes[j] = nodes[j], nodes[i]
			}
		}
	}
}

// GetExitNodeLocation gets the location of current exit node
func (ens *ExitNodeSelector) GetExitNodeLocation(ctx context.Context) (*ExitNodeInfo, error) {
	// Query current exit node location via external IP check service
	// This would typically involve making a request through Tor and checking the exit IP

	// Use check.torproject.org or similar service
	url := "https://check.torproject.org/api/ip"

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := ens.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to check exit node: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result struct {
		IP    string `json:"IP"`
		IsTor bool   `json:"IsTor"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &ExitNodeInfo{
		IP:          result.IP,
		LastChecked: time.Now(),
	}, nil
}

// SetExitNodeCountry sets the preferred exit node country in Tor config
func (ens *ExitNodeSelector) SetExitNodeCountry(config *TorConfig, countryCode string) {
	config.ExitNodeCountry = strings.ToUpper(countryCode)
	config.StrictNodes = true
}

// ListAvailableCountries returns a list of countries with exit nodes
func (ens *ExitNodeSelector) ListAvailableCountries(ctx context.Context) ([]string, error) {
	url := "https://onionoo.torproject.org/summary?flag=Exit&flag=Running"

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "moltbunker/1.0")

	resp, err := ens.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch countries: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var summaryResp struct {
		Relays []struct {
			Country string `json:"c"`
		} `json:"relays"`
	}
	if err := json.Unmarshal(body, &summaryResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Deduplicate countries
	countrySet := make(map[string]bool)
	for _, relay := range summaryResp.Relays {
		if relay.Country != "" {
			countrySet[strings.ToUpper(relay.Country)] = true
		}
	}

	countries := make([]string, 0, len(countrySet))
	for country := range countrySet {
		countries = append(countries, country)
	}

	return countries, nil
}

// GetExitNodesByBandwidth returns top N exit nodes by bandwidth
func (ens *ExitNodeSelector) GetExitNodesByBandwidth(ctx context.Context, limit int) ([]*ExitNodeInfo, error) {
	url := "https://onionoo.torproject.org/details?flag=Exit&flag=Running&flag=Valid&order=-observed_bandwidth&fields=nickname,fingerprint,or_addresses,country,country_name,flags,observed_bandwidth"

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "moltbunker/1.0")

	resp, err := ens.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch exit nodes: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var onionooResp OnionooResponse
	if err := json.Unmarshal(body, &onionooResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	nodes := make([]*ExitNodeInfo, 0, limit)
	for _, relay := range onionooResp.Relays {
		if len(nodes) >= limit {
			break
		}

		ip := ""
		if len(relay.OrAddresses) > 0 {
			addr := relay.OrAddresses[0]
			if idx := strings.LastIndex(addr, ":"); idx > 0 {
				ip = addr[:idx]
			}
			ip = strings.Trim(ip, "[]")
		}

		nodes = append(nodes, &ExitNodeInfo{
			Fingerprint: relay.Fingerprint,
			Nickname:    relay.Nickname,
			Country:     relay.Country,
			CountryName: relay.CountryName,
			IP:          ip,
			Bandwidth:   relay.Bandwidth,
			Flags:       relay.Flags,
			LastChecked: time.Now(),
		})
	}

	return nodes, nil
}
