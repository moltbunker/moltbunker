package proxy

import (
	"strings"
	"sync"
)

// ACL manages domain allowlists and blocklists for the proxy.
type ACL struct {
	mu        sync.RWMutex
	allowlist map[string]bool // nil = allow all
	blocklist map[string]bool
}

// NewACL creates a new ACL with optional allow/block lists.
// If allowlist is empty, all domains are allowed by default.
func NewACL(allowlist, blocklist []string) *ACL {
	acl := &ACL{
		blocklist: make(map[string]bool),
	}

	if len(allowlist) > 0 {
		acl.allowlist = make(map[string]bool)
		for _, d := range allowlist {
			acl.allowlist[strings.ToLower(d)] = true
		}
	}

	for _, d := range blocklist {
		acl.blocklist[strings.ToLower(d)] = true
	}

	return acl
}

// IsAllowed checks if a domain is permitted by the ACL.
func (a *ACL) IsAllowed(domain string) bool {
	if a == nil {
		return true
	}

	a.mu.RLock()
	defer a.mu.RUnlock()

	domain = strings.ToLower(domain)

	// Check blocklist first
	if a.isBlocked(domain) {
		return false
	}

	// If allowlist is set, domain must be in it
	if a.allowlist != nil {
		return a.isInList(a.allowlist, domain)
	}

	return true
}

// isBlocked checks if the domain (or any parent) is in the blocklist.
func (a *ACL) isBlocked(domain string) bool {
	return a.isInList(a.blocklist, domain)
}

// isInList checks if the domain matches any entry in the list,
// including parent domain matching (e.g., "evil.com" blocks "sub.evil.com").
func (a *ACL) isInList(list map[string]bool, domain string) bool {
	// Exact match
	if list[domain] {
		return true
	}

	// Check parent domains
	parts := strings.Split(domain, ".")
	for i := 1; i < len(parts); i++ {
		parent := strings.Join(parts[i:], ".")
		if list[parent] {
			return true
		}
	}

	return false
}

// AddToBlocklist adds domains to the blocklist.
func (a *ACL) AddToBlocklist(domains ...string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, d := range domains {
		a.blocklist[strings.ToLower(d)] = true
	}
}

// RemoveFromBlocklist removes domains from the blocklist.
func (a *ACL) RemoveFromBlocklist(domains ...string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, d := range domains {
		delete(a.blocklist, strings.ToLower(d))
	}
}
