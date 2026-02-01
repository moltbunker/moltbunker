package tor

import (
	"testing"
)

// Note: Circuit tests are largely integration tests that require a running Tor instance.
// These unit tests focus on configuration and setup logic.

func TestCircuitConfiguration(t *testing.T) {
	config := DefaultTorConfig("/tmp/tor")

	// Test default configuration
	if config.ControlPort == 0 {
		t.Error("ControlPort should have a default value")
	}

	if config.SOCKS5Port == 0 {
		t.Error("SOCKS5Port should have a default value")
	}
}

func TestCircuitExitNodeConfiguration(t *testing.T) {
	config := DefaultTorConfig("/tmp/tor")
	selector := NewExitNodeSelector()

	// Configure exit node country
	selector.SetExitNodeCountry(config, "de")

	if config.ExitNodeCountry != "DE" {
		t.Errorf("ExitNodeCountry should be DE, got %s", config.ExitNodeCountry)
	}

	if !config.StrictNodes {
		t.Error("StrictNodes should be enabled when exit country is set")
	}
}
