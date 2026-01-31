package p2p

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// GeoLocation represents geographic location information
type GeoLocation struct {
	Country string
	Region  string
	City    string
	Lat     float64
	Lon     float64
	IP      string
}

// GeoLocator determines geographic location from IP address
type GeoLocator struct {
	client *http.Client
}

// NewGeoLocator creates a new geolocator
func NewGeoLocator() *GeoLocator {
	return &GeoLocator{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetLocationFromIP determines location from IP address
func (gl *GeoLocator) GetLocationFromIP(ip string) (*GeoLocation, error) {
	// Use ip-api.com for geolocation (free, no API key required)
	url := fmt.Sprintf("http://ip-api.com/json/%s", ip)

	resp, err := gl.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch geolocation: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result struct {
		Status      string  `json:"status"`
		Country     string  `json:"country"`
		RegionName  string  `json:"regionName"`
		City        string  `json:"city"`
		Lat         float64 `json:"lat"`
		Lon         float64 `json:"lon"`
		Query       string  `json:"query"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if result.Status != "success" {
		return nil, fmt.Errorf("geolocation lookup failed: %s", result.Status)
	}

	return &GeoLocation{
		Country: result.Country,
		Region:  result.RegionName,
		City:    result.City,
		Lat:     result.Lat,
		Lon:     result.Lon,
		IP:      result.Query,
	}, nil
}

// GetRegionFromCountry maps country to region
func GetRegionFromCountry(country string) string {
	// Map countries to regions
	regions := map[string]string{
		"US": "Americas", "CA": "Americas", "MX": "Americas", "BR": "Americas",
		"GB": "Europe", "DE": "Europe", "FR": "Europe", "IT": "Europe", "ES": "Europe",
		"CN": "Asia-Pacific", "JP": "Asia-Pacific", "KR": "Asia-Pacific", "IN": "Asia-Pacific", "AU": "Asia-Pacific",
		"ZA": "Africa", "NG": "Africa", "EG": "Africa", "KE": "Africa",
	}

	if region, ok := regions[country]; ok {
		return region
	}

	// Default regions based on country code patterns
	if len(country) == 2 {
		firstChar := country[0]
		switch {
		case firstChar >= 'A' && firstChar <= 'C':
			return "Americas"
		case firstChar >= 'D' && firstChar <= 'H':
			return "Europe"
		case firstChar >= 'I' && firstChar <= 'M':
			return "Asia-Pacific"
		case firstChar >= 'N' && firstChar <= 'S':
			return "Africa"
		default:
			return "Other"
		}
	}

	return "Unknown"
}

// IsDifferentRegion checks if two locations are in different regions
func IsDifferentRegion(loc1, loc2 *GeoLocation) bool {
	return GetRegionFromCountry(loc1.Country) != GetRegionFromCountry(loc2.Country)
}
