package p2p

import (
	"testing"
)

func TestGetRegionFromCountry(t *testing.T) {
	tests := []struct {
		country string
		want    string
	}{
		{"US", "Americas"},
		{"CA", "Americas"},
		{"GB", "Europe"},
		{"DE", "Europe"},
		{"CN", "Asia-Pacific"},
		{"JP", "Asia-Pacific"},
		{"ZA", "Africa"},
		{"SG", "Asia-Pacific"},
		{"AE", "Middle-East"},
		{"XX", "Unknown"}, // Unknown codes return "Unknown"
	}

	for _, tt := range tests {
		got := GetRegionFromCountry(tt.country)
		if got != tt.want {
			t.Errorf("GetRegionFromCountry(%s) = %s, want %s", tt.country, got, tt.want)
		}
	}
}

func TestIsDifferentRegion(t *testing.T) {
	loc1 := &GeoLocation{Country: "US"}
	loc2 := &GeoLocation{Country: "GB"}
	loc3 := &GeoLocation{Country: "CA"}

	if !IsDifferentRegion(loc1, loc2) {
		t.Error("US and GB should be in different regions")
	}

	if IsDifferentRegion(loc1, loc3) {
		t.Error("US and CA should be in the same region")
	}
}

func TestGeoLocator_GetLocationFromIP(t *testing.T) {
	// This test requires internet connection
	// Skip if running in CI or offline
	if testing.Short() {
		t.Skip("Skipping test that requires internet connection")
	}

	gl := NewGeoLocator()

	// Test with a known IP (Google DNS)
	location, err := gl.GetLocationFromIP("8.8.8.8")
	if err != nil {
		t.Skipf("Skipping test due to network error: %v", err)
	}

	if location == nil {
		t.Error("Location should not be nil")
	}

	if location.IP != "8.8.8.8" {
		t.Errorf("IP should be 8.8.8.8, got %s", location.IP)
	}
}
