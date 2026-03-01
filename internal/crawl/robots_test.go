package crawl

import (
	"testing"
)

func TestRobots_AllowAll(t *testing.T) {
	rc := NewRobotsChecker()
	rc.Parse("example.com", "User-agent: *\nAllow: /\n")

	if !rc.IsAllowed("example.com", "/anything") {
		t.Error("should allow everything")
	}
}

func TestRobots_DisallowAll(t *testing.T) {
	rc := NewRobotsChecker()
	rc.Parse("example.com", "User-agent: *\nDisallow: /\n")

	if rc.IsAllowed("example.com", "/anything") {
		t.Error("should disallow everything")
	}
}

func TestRobots_DisallowPath(t *testing.T) {
	rc := NewRobotsChecker()
	rc.Parse("example.com", "User-agent: *\nDisallow: /admin\nDisallow: /private/\n")

	if rc.IsAllowed("example.com", "/admin") {
		t.Error("should disallow /admin")
	}
	if rc.IsAllowed("example.com", "/admin/settings") {
		t.Error("should disallow /admin/settings")
	}
	if !rc.IsAllowed("example.com", "/public") {
		t.Error("should allow /public")
	}
	if rc.IsAllowed("example.com", "/private/data") {
		t.Error("should disallow /private/data")
	}
}

func TestRobots_AllowOverridesDisallow(t *testing.T) {
	rc := NewRobotsChecker()
	rc.Parse("example.com", "User-agent: *\nDisallow: /api\nAllow: /api/public\n")

	if !rc.IsAllowed("example.com", "/api/public") {
		t.Error("allow should override disallow for /api/public")
	}
	if rc.IsAllowed("example.com", "/api/private") {
		t.Error("should disallow /api/private")
	}
}

func TestRobots_MoltbunkerAgent(t *testing.T) {
	rc := NewRobotsChecker()
	rc.Parse("example.com", "User-agent: MoltbunkerCrawler\nDisallow: /secret\n")

	if rc.IsAllowed("example.com", "/secret") {
		t.Error("should disallow for MoltbunkerCrawler agent")
	}
}

func TestRobots_CrawlDelay(t *testing.T) {
	rc := NewRobotsChecker()
	rc.Parse("example.com", "User-agent: *\nCrawl-delay: 5\n")

	delay := rc.GetCrawlDelay("example.com")
	if delay != 5 {
		t.Errorf("delay = %d, want 5", delay)
	}
}

func TestRobots_CrawlDelay_NotSet(t *testing.T) {
	rc := NewRobotsChecker()
	rc.Parse("example.com", "User-agent: *\nDisallow: /x\n")

	delay := rc.GetCrawlDelay("example.com")
	if delay != 0 {
		t.Errorf("delay = %d, want 0", delay)
	}
}

func TestRobots_NoParsed(t *testing.T) {
	rc := NewRobotsChecker()

	if !rc.IsAllowed("unknown.com", "/anything") {
		t.Error("no robots.txt should allow by default")
	}

	delay := rc.GetCrawlDelay("unknown.com")
	if delay != 0 {
		t.Errorf("delay = %d, want 0 for unknown domain", delay)
	}
}

func TestRobots_CommentsAndBlanks(t *testing.T) {
	rc := NewRobotsChecker()
	rc.Parse("example.com", `
# This is a comment
User-agent: *

# Block admin
Disallow: /admin

# Allow public
Allow: /public
`)

	if rc.IsAllowed("example.com", "/admin") {
		t.Error("should disallow /admin")
	}
	if !rc.IsAllowed("example.com", "/public") {
		t.Error("should allow /public")
	}
}

func TestRobots_InvalidCrawlDelay(t *testing.T) {
	rc := NewRobotsChecker()
	rc.Parse("example.com", "User-agent: *\nCrawl-delay: abc\n")

	delay := rc.GetCrawlDelay("example.com")
	if delay != 0 {
		t.Errorf("delay = %d, want 0 for invalid value", delay)
	}
}

func TestRobots_EmptyDisallow(t *testing.T) {
	rc := NewRobotsChecker()
	rc.Parse("example.com", "User-agent: *\nDisallow:\n")

	if !rc.IsAllowed("example.com", "/anything") {
		t.Error("empty disallow should allow everything")
	}
}

func TestRobots_MultipleDomains(t *testing.T) {
	rc := NewRobotsChecker()
	rc.Parse("a.com", "User-agent: *\nDisallow: /x\n")
	rc.Parse("b.com", "User-agent: *\nDisallow: /y\n")

	if rc.IsAllowed("a.com", "/x") {
		t.Error("a.com should disallow /x")
	}
	if !rc.IsAllowed("a.com", "/y") {
		t.Error("a.com should allow /y")
	}
	if !rc.IsAllowed("b.com", "/x") {
		t.Error("b.com should allow /x")
	}
	if rc.IsAllowed("b.com", "/y") {
		t.Error("b.com should disallow /y")
	}
}

func TestRobots_OtherUserAgent(t *testing.T) {
	rc := NewRobotsChecker()
	rc.Parse("example.com", "User-agent: Googlebot\nDisallow: /secret\n")

	// Rules for Googlebot should not affect us
	if !rc.IsAllowed("example.com", "/secret") {
		t.Error("Googlebot rules should not apply to us")
	}
}
