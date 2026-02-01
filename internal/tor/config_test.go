package tor

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultTorConfig(t *testing.T) {
	dataDir := "/tmp/test-tor"
	config := DefaultTorConfig(dataDir)

	if config.DataDir != dataDir {
		t.Errorf("DataDir mismatch: got %s, want %s", config.DataDir, dataDir)
	}

	if config.ControlPort != 9051 {
		t.Errorf("ControlPort mismatch: got %d, want 9051", config.ControlPort)
	}

	if config.SOCKS5Port != 9050 {
		t.Errorf("SOCKS5Port mismatch: got %d, want 9050", config.SOCKS5Port)
	}

	expectedTorrcPath := filepath.Join(dataDir, "torrc")
	if config.TorrcPath != expectedTorrcPath {
		t.Errorf("TorrcPath mismatch: got %s, want %s", config.TorrcPath, expectedTorrcPath)
	}
}

func TestTorConfig_WriteTorrc(t *testing.T) {
	tempDir := t.TempDir()
	config := DefaultTorConfig(tempDir)

	if err := config.WriteTorrc(); err != nil {
		t.Fatalf("Failed to write torrc: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(config.TorrcPath); os.IsNotExist(err) {
		t.Error("torrc file was not created")
	}

	// Verify content
	content, err := os.ReadFile(config.TorrcPath)
	if err != nil {
		t.Fatalf("Failed to read torrc: %v", err)
	}

	contentStr := string(content)
	if len(contentStr) == 0 {
		t.Error("torrc file is empty")
	}
}

func TestTorConfig_WriteTorrc_WithExitCountry(t *testing.T) {
	tempDir := t.TempDir()
	config := DefaultTorConfig(tempDir)
	config.ExitNodeCountry = "US"
	config.StrictNodes = true

	if err := config.WriteTorrc(); err != nil {
		t.Fatalf("Failed to write torrc: %v", err)
	}

	content, err := os.ReadFile(config.TorrcPath)
	if err != nil {
		t.Fatalf("Failed to read torrc: %v", err)
	}

	contentStr := string(content)
	if !contains(contentStr, "ExitNodes {US}") {
		t.Error("torrc should contain ExitNodes directive")
	}
	if !contains(contentStr, "StrictNodes 1") {
		t.Error("torrc should contain StrictNodes directive")
	}
}

func TestTorConfig_ControlAddress(t *testing.T) {
	config := DefaultTorConfig("/tmp/tor")

	expected := "127.0.0.1:9051"
	if config.ControlAddress() != expected {
		t.Errorf("ControlAddress mismatch: got %s, want %s", config.ControlAddress(), expected)
	}
}

func TestTorConfig_SOCKS5Address(t *testing.T) {
	config := DefaultTorConfig("/tmp/tor")

	expected := "127.0.0.1:9050"
	if config.SOCKS5Address() != expected {
		t.Errorf("SOCKS5Address mismatch: got %s, want %s", config.SOCKS5Address(), expected)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
