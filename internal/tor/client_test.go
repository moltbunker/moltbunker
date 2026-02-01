package tor

import (
	"testing"
)

// Note: Client tests are largely integration tests that require a running Tor instance.
// These tests focus on configuration validation.

func TestTorClientConfiguration(t *testing.T) {
	config := DefaultTorConfig("/tmp/tor")

	// Validate SOCKS5 address format
	socks5Addr := config.SOCKS5Address()
	if socks5Addr == "" {
		t.Error("SOCKS5 address should not be empty")
	}

	// Validate control address format
	controlAddr := config.ControlAddress()
	if controlAddr == "" {
		t.Error("Control address should not be empty")
	}
}
