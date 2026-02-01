package tor

import (
	"testing"
)

func TestTorService_NewTorService(t *testing.T) {
	config := DefaultTorConfig("/tmp/test-tor-service")
	service, err := NewTorService(config)
	if err != nil {
		t.Fatalf("NewTorService failed: %v", err)
	}

	if service == nil {
		t.Fatal("NewTorService returned nil")
	}

	if service.config.DataDir != config.DataDir {
		t.Errorf("dataDir mismatch: got %s, want %s", service.config.DataDir, config.DataDir)
	}
}

func TestTorService_OnionAddress_NotStarted(t *testing.T) {
	config := DefaultTorConfig("/tmp/test-tor")
	service, err := NewTorService(config)
	if err != nil {
		t.Fatalf("NewTorService failed: %v", err)
	}

	addr := service.OnionAddress()
	if addr != "" {
		t.Error("OnionAddress should be empty when service not started")
	}
}
