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
	Country     string
	CountryCode string
	Region      string
	City        string
	Lat         float64
	Lon         float64
	IP          string
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
		CountryCode string  `json:"countryCode"`
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
		Country:     result.Country,
		CountryCode: result.CountryCode,
		Region:      result.RegionName,
		City:        result.City,
		Lat:         result.Lat,
		Lon:         result.Lon,
		IP:          result.Query,
	}, nil
}

// GetRegionFromCountry maps ISO 3166-1 alpha-2 country code to region
func GetRegionFromCountry(country string) string {
	if region, ok := countryToRegion[country]; ok {
		return region
	}
	return "Unknown"
}

// countryToRegion maps ISO 3166-1 alpha-2 codes to geographic regions.
// Covers all countries with significant hosting infrastructure plus common codes.
var countryToRegion = map[string]string{
	// Americas
	"US": "Americas", "CA": "Americas", "MX": "Americas", "BR": "Americas",
	"AR": "Americas", "CL": "Americas", "CO": "Americas", "PE": "Americas",
	"VE": "Americas", "EC": "Americas", "UY": "Americas", "PY": "Americas",
	"BO": "Americas", "CR": "Americas", "PA": "Americas", "GT": "Americas",
	"CU": "Americas", "DO": "Americas", "HN": "Americas", "SV": "Americas",
	"NI": "Americas", "JM": "Americas", "TT": "Americas", "PR": "Americas",

	// Europe
	"GB": "Europe", "DE": "Europe", "FR": "Europe", "IT": "Europe",
	"ES": "Europe", "NL": "Europe", "BE": "Europe", "AT": "Europe",
	"CH": "Europe", "SE": "Europe", "NO": "Europe", "DK": "Europe",
	"FI": "Europe", "IE": "Europe", "PT": "Europe", "PL": "Europe",
	"CZ": "Europe", "RO": "Europe", "HU": "Europe", "GR": "Europe",
	"BG": "Europe", "HR": "Europe", "SK": "Europe", "SI": "Europe",
	"LT": "Europe", "LV": "Europe", "EE": "Europe", "LU": "Europe",
	"IS": "Europe", "UA": "Europe", "RS": "Europe", "BA": "Europe",
	"AL": "Europe", "MK": "Europe", "ME": "Europe", "MT": "Europe",
	"CY": "Europe", "MD": "Europe", "BY": "Europe", "RU": "Europe",
	"TR": "Europe", "GE": "Europe", "AM": "Europe", "AZ": "Europe",

	// Asia-Pacific
	"CN": "Asia-Pacific", "JP": "Asia-Pacific", "KR": "Asia-Pacific",
	"IN": "Asia-Pacific", "AU": "Asia-Pacific", "NZ": "Asia-Pacific",
	"SG": "Asia-Pacific", "HK": "Asia-Pacific", "TW": "Asia-Pacific",
	"TH": "Asia-Pacific", "VN": "Asia-Pacific", "MY": "Asia-Pacific",
	"ID": "Asia-Pacific", "PH": "Asia-Pacific", "BD": "Asia-Pacific",
	"PK": "Asia-Pacific", "LK": "Asia-Pacific", "MM": "Asia-Pacific",
	"KH": "Asia-Pacific", "LA": "Asia-Pacific", "NP": "Asia-Pacific",
	"MN": "Asia-Pacific", "KZ": "Asia-Pacific", "UZ": "Asia-Pacific",
	"FJ": "Asia-Pacific", "PG": "Asia-Pacific",

	// Middle East
	"AE": "Middle-East", "SA": "Middle-East", "IL": "Middle-East",
	"QA": "Middle-East", "KW": "Middle-East", "BH": "Middle-East",
	"OM": "Middle-East", "JO": "Middle-East", "LB": "Middle-East",
	"IQ": "Middle-East", "IR": "Middle-East", "YE": "Middle-East",
	"SY": "Middle-East",

	// Africa
	"ZA": "Africa", "NG": "Africa", "EG": "Africa", "KE": "Africa",
	"GH": "Africa", "TZ": "Africa", "ET": "Africa", "MA": "Africa",
	"TN": "Africa", "DZ": "Africa", "UG": "Africa", "RW": "Africa",
	"SN": "Africa", "CI": "Africa", "CM": "Africa", "MZ": "Africa",
	"AO": "Africa", "ZW": "Africa", "MU": "Africa", "LY": "Africa",
	"NA": "Africa", "BW": "Africa", "MG": "Africa",
}

// IsDifferentRegion checks if two locations are in different regions
func IsDifferentRegion(loc1, loc2 *GeoLocation) bool {
	return GetRegionFromCountry(loc1.Country) != GetRegionFromCountry(loc2.Country)
}
