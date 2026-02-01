package tor

import (
	"context"
	"testing"
	"time"
)

func TestNewExitNodeSelector(t *testing.T) {
	selector := NewExitNodeSelector()

	if selector == nil {
		t.Fatal("NewExitNodeSelector returned nil")
	}

	if selector.client == nil {
		t.Error("HTTP client should be initialized")
	}

	if selector.cache == nil {
		t.Error("Cache should be initialized")
	}

	if selector.cacheTTL != 1*time.Hour {
		t.Errorf("Cache TTL mismatch: got %v, want 1h", selector.cacheTTL)
	}
}

func TestExitNodeSelector_SetExitNodeCountry(t *testing.T) {
	selector := NewExitNodeSelector()
	config := DefaultTorConfig("/tmp/tor")

	selector.SetExitNodeCountry(config, "us")

	if config.ExitNodeCountry != "US" {
		t.Errorf("ExitNodeCountry should be uppercase: got %s, want US", config.ExitNodeCountry)
	}

	if !config.StrictNodes {
		t.Error("StrictNodes should be true")
	}
}

func TestSortByBandwidth(t *testing.T) {
	nodes := []*ExitNodeInfo{
		{Fingerprint: "a", Bandwidth: 100},
		{Fingerprint: "b", Bandwidth: 300},
		{Fingerprint: "c", Bandwidth: 200},
	}

	sortByBandwidth(nodes)

	if nodes[0].Fingerprint != "b" {
		t.Error("First node should have highest bandwidth")
	}
	if nodes[1].Fingerprint != "c" {
		t.Error("Second node should have second highest bandwidth")
	}
	if nodes[2].Fingerprint != "a" {
		t.Error("Third node should have lowest bandwidth")
	}
}

// Integration test - requires network access
func TestExitNodeSelector_SelectExitNodeByCountry_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	selector := NewExitNodeSelector()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Try to find US exit nodes
	node, err := selector.SelectExitNodeByCountry(ctx, "us")
	if err != nil {
		t.Logf("Note: Could not fetch exit nodes (network may be unavailable): %v", err)
		return
	}

	if node == nil {
		t.Error("Expected non-nil node")
		return
	}

	if node.Country != "us" {
		t.Errorf("Expected country 'us', got '%s'", node.Country)
	}

	if node.Fingerprint == "" {
		t.Error("Fingerprint should not be empty")
	}
}

// Integration test - requires network access
func TestExitNodeSelector_ListAvailableCountries_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	selector := NewExitNodeSelector()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	countries, err := selector.ListAvailableCountries(ctx)
	if err != nil {
		t.Logf("Note: Could not fetch countries (network may be unavailable): %v", err)
		return
	}

	if len(countries) == 0 {
		// Empty response may be due to rate limiting or API issues
		// Don't fail the test, just log it
		t.Log("Note: No countries returned - API may be rate-limited or temporarily unavailable")
		return
	}

	t.Logf("Found %d countries with exit nodes", len(countries))
}
