//go:build e2e

package e2e

import (
	"os"
	"testing"
)

// TestMain sets up the E2E test environment
func TestMain(m *testing.M) {
	// Set up mock payments environment for all E2E tests
	os.Setenv("MOLTBUNKER_MOCK_PAYMENTS", "true")

	// Run tests
	code := m.Run()

	// Cleanup
	os.Unsetenv("MOLTBUNKER_MOCK_PAYMENTS")

	os.Exit(code)
}
