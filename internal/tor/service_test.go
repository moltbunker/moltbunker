package tor

import (
	"testing"
)

func TestNewTorService(t *testing.T) {
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

	// Initially, tor and onion should be nil
	if service.tor != nil {
		t.Error("tor should be nil before Start")
	}

	if service.onion != nil {
		t.Error("onion should be nil before Start")
	}
}

func TestTorService_OnionAddress_BeforeStart(t *testing.T) {
	config := DefaultTorConfig("/tmp/test-tor")
	service, err := NewTorService(config)
	if err != nil {
		t.Fatalf("NewTorService failed: %v", err)
	}

	addr := service.OnionAddress()
	if addr != "" {
		t.Errorf("OnionAddress should be empty before start, got %s", addr)
	}
}

func TestTorService_Tor_BeforeStart(t *testing.T) {
	config := DefaultTorConfig("/tmp/test-tor")
	service, err := NewTorService(config)
	if err != nil {
		t.Fatalf("NewTorService failed: %v", err)
	}

	tor := service.Tor()
	if tor != nil {
		t.Error("Tor() should return nil before start")
	}
}

func TestTorService_OnionService_BeforeStart(t *testing.T) {
	config := DefaultTorConfig("/tmp/test-tor")
	service, err := NewTorService(config)
	if err != nil {
		t.Fatalf("NewTorService failed: %v", err)
	}

	onion := service.OnionService()
	if onion != nil {
		t.Error("OnionService() should return nil before start")
	}
}

func TestTorService_Stop_NotStarted(t *testing.T) {
	config := DefaultTorConfig("/tmp/test-tor")
	service, err := NewTorService(config)
	if err != nil {
		t.Fatalf("NewTorService failed: %v", err)
	}

	// Stop should not error when not started
	err = service.Stop()
	if err != nil {
		t.Errorf("Stop should not error when not started: %v", err)
	}
}

func TestTorService_IsRunning_BeforeStart(t *testing.T) {
	config := DefaultTorConfig("/tmp/test-tor")
	service, err := NewTorService(config)
	if err != nil {
		t.Fatalf("NewTorService failed: %v", err)
	}

	if service.IsRunning() {
		t.Error("IsRunning should return false before Start")
	}
}

func TestTorService_GetOnionAddress_BeforeStart(t *testing.T) {
	config := DefaultTorConfig("/tmp/test-tor")
	service, err := NewTorService(config)
	if err != nil {
		t.Fatalf("NewTorService failed: %v", err)
	}

	addr := service.GetOnionAddress()
	if addr != "" {
		t.Errorf("GetOnionAddress should be empty before start, got %s", addr)
	}
}
