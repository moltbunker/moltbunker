package crawl

import (
	"bufio"
	"strings"
	"sync"
)

// RobotsChecker parses and evaluates robots.txt rules.
type RobotsChecker struct {
	mu    sync.RWMutex
	rules map[string]*RobotsRule // domain → parsed rules
}

// NewRobotsChecker creates a new robots.txt checker.
func NewRobotsChecker() *RobotsChecker {
	return &RobotsChecker{
		rules: make(map[string]*RobotsRule),
	}
}

// Parse parses a robots.txt body and stores rules for a domain.
func (rc *RobotsChecker) Parse(domain, body string) {
	rule := &RobotsRule{
		UserAgent: "*",
	}

	scanner := bufio.NewScanner(strings.NewReader(body))
	currentAgent := ""

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(strings.ToLower(parts[0]))
		value := strings.TrimSpace(parts[1])

		switch key {
		case "user-agent":
			currentAgent = value
		case "disallow":
			if currentAgent == "*" || strings.Contains(strings.ToLower(currentAgent), "moltbunker") {
				if value != "" {
					rule.Disallow = append(rule.Disallow, value)
				}
			}
		case "allow":
			if currentAgent == "*" || strings.Contains(strings.ToLower(currentAgent), "moltbunker") {
				if value != "" {
					rule.Allow = append(rule.Allow, value)
				}
			}
		case "crawl-delay":
			// Parse as integer seconds
			var delay int
			if _, err := parseIntSafe(value); err == nil {
				delay = mustParseInt(value)
				rule.CrawlDelay = delay
			}
		}
	}

	rc.mu.Lock()
	rc.rules[domain] = rule
	rc.mu.Unlock()
}

// IsAllowed checks if a path is allowed by the robots.txt rules for a domain.
func (rc *RobotsChecker) IsAllowed(domain, path string) bool {
	rc.mu.RLock()
	rule, ok := rc.rules[domain]
	rc.mu.RUnlock()

	if !ok {
		// No robots.txt loaded — allow by default
		return true
	}

	// Check allow rules first (more specific wins)
	for _, allow := range rule.Allow {
		if strings.HasPrefix(path, allow) {
			return true
		}
	}

	// Check disallow rules
	for _, disallow := range rule.Disallow {
		if disallow == "/" {
			return false // Disallow all
		}
		if strings.HasPrefix(path, disallow) {
			return false
		}
	}

	return true
}

// GetCrawlDelay returns the crawl-delay for a domain in seconds.
// Returns 0 if no delay is specified.
func (rc *RobotsChecker) GetCrawlDelay(domain string) int {
	rc.mu.RLock()
	defer rc.mu.RUnlock()

	rule, ok := rc.rules[domain]
	if !ok {
		return 0
	}
	return rule.CrawlDelay
}

// helpers to avoid strconv import for trivial int parsing

func parseIntSafe(s string) (int, error) {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, &intParseError{s}
		}
		n = n*10 + int(c-'0')
	}
	return n, nil
}

func mustParseInt(s string) int {
	n, _ := parseIntSafe(s)
	return n
}

type intParseError struct{ s string }

func (e *intParseError) Error() string { return "invalid integer: " + e.s }
