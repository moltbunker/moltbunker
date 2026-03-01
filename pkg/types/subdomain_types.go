package types

import (
	"fmt"
	"regexp"
)

// SubdomainNameRegex validates subdomain names: 3-32 chars, lowercase alphanumeric + hyphens,
// no leading/trailing hyphens. Follows DNS RFC 1035 conventions.
var SubdomainNameRegex = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{1,30}[a-z0-9]$`)

// ReservedSubdomains are names that cannot be registered by users.
var ReservedSubdomains = map[string]bool{
	"www":       true,
	"api":       true,
	"admin":     true,
	"mail":      true,
	"ftp":       true,
	"status":    true,
	"docs":      true,
	"app":       true,
	"dashboard": true,
	"node":      true,
	"bootstrap": true,
	"registry":  true,
}

// ValidateSubdomainName validates a subdomain name for registration.
func ValidateSubdomainName(name string) error {
	if len(name) < 3 {
		return fmt.Errorf("subdomain name must be at least 3 characters")
	}
	if len(name) > 32 {
		return fmt.Errorf("subdomain name must be at most 32 characters")
	}
	if !SubdomainNameRegex.MatchString(name) {
		return fmt.Errorf("subdomain name must contain only lowercase letters, numbers, and hyphens (no leading/trailing hyphens)")
	}
	// Block punycode-encoded internationalized domain names (IDN homograph attacks)
	if len(name) >= 4 && name[:4] == "xn--" {
		return fmt.Errorf("subdomain name cannot use punycode encoding (xn-- prefix)")
	}
	// Block consecutive hyphens (used in punycode and confusing in DNS)
	for i := 0; i < len(name)-1; i++ {
		if name[i] == '-' && name[i+1] == '-' {
			return fmt.Errorf("subdomain name cannot contain consecutive hyphens")
		}
	}
	if ReservedSubdomains[name] {
		return fmt.Errorf("subdomain name %q is reserved", name)
	}
	return nil
}
