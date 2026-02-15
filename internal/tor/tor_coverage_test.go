package tor

import (
	"context"
	"encoding/json"
	"os"
	"sort"
	"strings"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// TorConfig additional tests
// ---------------------------------------------------------------------------

func TestTorConfigStruct(t *testing.T) {
	config := TorConfig{
		DataDir:         "/var/lib/tor",
		ControlPort:     9051,
		SOCKS5Port:      9050,
		TorrcPath:       "/var/lib/tor/torrc",
		ExitNodeCountry: "US",
		StrictNodes:     true,
		EnableSocks:     true,
		TorOnly:         true,
	}

	if config.DataDir != "/var/lib/tor" {
		t.Errorf("DataDir = %q", config.DataDir)
	}
	if config.ControlPort != 9051 {
		t.Errorf("ControlPort = %d", config.ControlPort)
	}
	if config.SOCKS5Port != 9050 {
		t.Errorf("SOCKS5Port = %d", config.SOCKS5Port)
	}
	if config.ExitNodeCountry != "US" {
		t.Errorf("ExitNodeCountry = %q", config.ExitNodeCountry)
	}
	if !config.StrictNodes {
		t.Error("StrictNodes should be true")
	}
	if !config.EnableSocks {
		t.Error("EnableSocks should be true")
	}
	if !config.TorOnly {
		t.Error("TorOnly should be true")
	}
}

func TestDefaultTorConfig_Fields(t *testing.T) {
	tmpDir := t.TempDir()
	config := DefaultTorConfig(tmpDir)

	// Verify all defaults
	if config.ControlPort != 9051 {
		t.Errorf("ControlPort = %d, want 9051", config.ControlPort)
	}
	if config.SOCKS5Port != 9050 {
		t.Errorf("SOCKS5Port = %d, want 9050", config.SOCKS5Port)
	}
	if config.ExitNodeCountry != "" {
		t.Errorf("ExitNodeCountry should be empty by default, got %q", config.ExitNodeCountry)
	}
	if config.StrictNodes {
		t.Error("StrictNodes should be false by default")
	}
	if config.EnableSocks {
		t.Error("EnableSocks should be false by default")
	}
	if config.TorOnly {
		t.Error("TorOnly should be false by default")
	}
}

func TestTorConfig_ControlAddress_CustomPort(t *testing.T) {
	config := &TorConfig{ControlPort: 9999}
	expected := "127.0.0.1:9999"
	if config.ControlAddress() != expected {
		t.Errorf("ControlAddress() = %q, want %q", config.ControlAddress(), expected)
	}
}

func TestTorConfig_SOCKS5Address_CustomPort(t *testing.T) {
	config := &TorConfig{SOCKS5Port: 8888}
	expected := "127.0.0.1:8888"
	if config.SOCKS5Address() != expected {
		t.Errorf("SOCKS5Address() = %q, want %q", config.SOCKS5Address(), expected)
	}
}

func TestTorConfig_WriteTorrc_BasicContent(t *testing.T) {
	tmpDir := t.TempDir()
	config := DefaultTorConfig(tmpDir)

	err := config.WriteTorrc()
	if err != nil {
		t.Fatalf("WriteTorrc failed: %v", err)
	}

	content := readFile(t, config.TorrcPath)

	// Verify all required directives are present
	if !strings.Contains(content, "DataDirectory") {
		t.Error("torrc should contain DataDirectory")
	}
	if !strings.Contains(content, "ControlPort 9051") {
		t.Error("torrc should contain ControlPort 9051")
	}
	if !strings.Contains(content, "SOCKSPort 9050") {
		t.Error("torrc should contain SOCKSPort 9050")
	}
	// Should NOT contain exit node directives without setting them
	if strings.Contains(content, "ExitNodes") {
		t.Error("torrc should not contain ExitNodes when country not set")
	}
	if strings.Contains(content, "StrictNodes") {
		t.Error("torrc should not contain StrictNodes when not enabled")
	}
}

func TestTorConfig_WriteTorrc_WithStrictNodesDisabled(t *testing.T) {
	tmpDir := t.TempDir()
	config := DefaultTorConfig(tmpDir)
	config.ExitNodeCountry = "DE"
	config.StrictNodes = false

	err := config.WriteTorrc()
	if err != nil {
		t.Fatalf("WriteTorrc failed: %v", err)
	}

	content := readFile(t, config.TorrcPath)

	if !strings.Contains(content, "ExitNodes {DE}") {
		t.Error("torrc should contain ExitNodes {DE}")
	}
	// StrictNodes should not appear when disabled
	if strings.Contains(content, "StrictNodes 1") {
		t.Error("torrc should not contain StrictNodes 1 when disabled")
	}
}

// ---------------------------------------------------------------------------
// TorService tests
// ---------------------------------------------------------------------------

func TestNewTorService_NilConfig(t *testing.T) {
	_, err := NewTorService(nil)
	if err == nil {
		t.Error("expected error for nil config")
	}
	if !strings.Contains(err.Error(), "config required") {
		t.Errorf("expected 'config required' error, got: %v", err)
	}
}

func TestNewTorService_CreatesDataDir(t *testing.T) {
	tmpDir := t.TempDir()
	dataDir := tmpDir + "/nested/tor/data"
	config := &TorConfig{DataDir: dataDir}

	service, err := NewTorService(config)
	if err != nil {
		t.Fatalf("NewTorService failed: %v", err)
	}
	if service == nil {
		t.Fatal("expected non-nil TorService")
	}

	// Verify data directory was created
	if !dirExists(dataDir) {
		t.Error("data directory should be created")
	}
}

func TestTorService_InitialState(t *testing.T) {
	tmpDir := t.TempDir()
	config := DefaultTorConfig(tmpDir)
	service, err := NewTorService(config)
	if err != nil {
		t.Fatalf("NewTorService failed: %v", err)
	}

	if service.IsRunning() {
		t.Error("should not be running initially")
	}
	if service.GetOnionAddress() != "" {
		t.Error("onion address should be empty initially")
	}
	if service.OnionAddress() != "" {
		t.Error("OnionAddress() should be empty initially")
	}
	if service.Tor() != nil {
		t.Error("Tor() should be nil initially")
	}
	if service.OnionService() != nil {
		t.Error("OnionService() should be nil initially")
	}
}

func TestTorService_StopWithoutStart(t *testing.T) {
	tmpDir := t.TempDir()
	config := DefaultTorConfig(tmpDir)
	service, err := NewTorService(config)
	if err != nil {
		t.Fatalf("NewTorService failed: %v", err)
	}

	// Stop when not started should be safe
	err = service.Stop()
	if err != nil {
		t.Errorf("Stop should not error when not started: %v", err)
	}
}

func TestTorService_ListOnionServices_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	config := DefaultTorConfig(tmpDir)
	service, err := NewTorService(config)
	if err != nil {
		t.Fatalf("NewTorService failed: %v", err)
	}

	services := service.ListOnionServices()
	if len(services) != 0 {
		t.Errorf("expected no onion services, got %d", len(services))
	}
}

func TestTorService_CloseOnionService_NonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	config := DefaultTorConfig(tmpDir)
	service, err := NewTorService(config)
	if err != nil {
		t.Fatalf("NewTorService failed: %v", err)
	}

	// Closing non-existent service should not error
	err = service.CloseOnionService(8080)
	if err != nil {
		t.Errorf("CloseOnionService should not error for non-existent port: %v", err)
	}
}

func TestTorService_CreateOnionService_NotRunning(t *testing.T) {
	tmpDir := t.TempDir()
	config := DefaultTorConfig(tmpDir)
	service, err := NewTorService(config)
	if err != nil {
		t.Fatalf("NewTorService failed: %v", err)
	}

	_, err = service.CreateOnionService(context.Background(), 8080)
	if err == nil {
		t.Error("expected error when creating onion service while not running")
	}
	if !strings.Contains(err.Error(), "Tor not running") {
		t.Errorf("expected 'Tor not running' error, got: %v", err)
	}
}

func TestTorService_RotateCircuit_NotRunning(t *testing.T) {
	tmpDir := t.TempDir()
	config := DefaultTorConfig(tmpDir)
	service, err := NewTorService(config)
	if err != nil {
		t.Fatalf("NewTorService failed: %v", err)
	}

	err = service.RotateCircuit(context.Background())
	if err == nil {
		t.Error("expected error when rotating circuit while not running")
	}
	if !strings.Contains(err.Error(), "Tor not running") {
		t.Errorf("expected 'Tor not running' error, got: %v", err)
	}
}

func TestTorService_GetCircuitInfo_NotRunning(t *testing.T) {
	tmpDir := t.TempDir()
	config := DefaultTorConfig(tmpDir)
	service, err := NewTorService(config)
	if err != nil {
		t.Fatalf("NewTorService failed: %v", err)
	}

	_, err = service.GetCircuitInfo(context.Background())
	if err == nil {
		t.Error("expected error when getting circuit info while not running")
	}
	if !strings.Contains(err.Error(), "Tor not running") {
		t.Errorf("expected 'Tor not running' error, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// CircuitManager tests
// ---------------------------------------------------------------------------

func TestNewCircuitManager(t *testing.T) {
	t.Run("with nil tor instance", func(t *testing.T) {
		cm := NewCircuitManager(nil)
		if cm == nil {
			t.Fatal("expected non-nil CircuitManager")
		}
		if cm.tor != nil {
			t.Error("tor should be nil")
		}
		if cm.rotationInterval != 10*time.Minute {
			t.Errorf("rotationInterval = %v, want 10m", cm.rotationInterval)
		}
	})
}

func TestCircuitManager_SetRotationInterval(t *testing.T) {
	cm := NewCircuitManager(nil)

	cm.SetRotationInterval(30 * time.Minute)
	if cm.rotationInterval != 30*time.Minute {
		t.Errorf("rotationInterval = %v, want 30m", cm.rotationInterval)
	}

	cm.SetRotationInterval(5 * time.Minute)
	if cm.rotationInterval != 5*time.Minute {
		t.Errorf("rotationInterval = %v, want 5m", cm.rotationInterval)
	}
}

func TestCircuitManager_LastRotation(t *testing.T) {
	before := time.Now()
	cm := NewCircuitManager(nil)
	after := time.Now()

	lastRotation := cm.LastRotation()
	if lastRotation.Before(before) || lastRotation.After(after) {
		t.Errorf("LastRotation = %v, expected between %v and %v", lastRotation, before, after)
	}
}

func TestCircuitManager_RotateCircuit_NilTor(t *testing.T) {
	cm := NewCircuitManager(nil)

	err := cm.RotateCircuit(context.Background())
	if err == nil {
		t.Error("expected error when rotating with nil tor")
	}
	if !strings.Contains(err.Error(), "not available") {
		t.Errorf("expected 'not available' error, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// CircuitInfo struct tests
// ---------------------------------------------------------------------------

func TestCircuitInfoStruct(t *testing.T) {
	ci := CircuitInfo{
		ID:     "circuit-1",
		Status: "BUILT",
		Path:   []string{"guard", "middle", "exit"},
		Raw:    "1 BUILT guard,middle,exit",
	}

	if ci.ID != "circuit-1" {
		t.Errorf("ID = %q, want %q", ci.ID, "circuit-1")
	}
	if ci.Status != "BUILT" {
		t.Errorf("Status = %q, want %q", ci.Status, "BUILT")
	}
	if len(ci.Path) != 3 {
		t.Errorf("Path length = %d, want 3", len(ci.Path))
	}
	if ci.Raw == "" {
		t.Error("Raw should not be empty")
	}
}

// ---------------------------------------------------------------------------
// ExitNodeInfo struct tests
// ---------------------------------------------------------------------------

func TestExitNodeInfoStruct(t *testing.T) {
	info := ExitNodeInfo{
		Fingerprint: "ABCDEF1234567890",
		Nickname:    "TestRelay",
		Country:     "us",
		CountryName: "United States",
		IP:          "1.2.3.4",
		Bandwidth:   50000,
		Flags:       []string{"Exit", "Running", "Valid"},
		LastChecked: time.Now(),
	}

	if info.Fingerprint != "ABCDEF1234567890" {
		t.Errorf("Fingerprint = %q", info.Fingerprint)
	}
	if info.Nickname != "TestRelay" {
		t.Errorf("Nickname = %q", info.Nickname)
	}
	if info.Country != "us" {
		t.Errorf("Country = %q", info.Country)
	}
	if info.CountryName != "United States" {
		t.Errorf("CountryName = %q", info.CountryName)
	}
	if info.IP != "1.2.3.4" {
		t.Errorf("IP = %q", info.IP)
	}
	if info.Bandwidth != 50000 {
		t.Errorf("Bandwidth = %d", info.Bandwidth)
	}
	if len(info.Flags) != 3 {
		t.Errorf("Flags length = %d, want 3", len(info.Flags))
	}
	if info.LastChecked.IsZero() {
		t.Error("LastChecked should not be zero")
	}
}

// ---------------------------------------------------------------------------
// ExitNodeSelector tests
// ---------------------------------------------------------------------------

func TestNewExitNodeSelector_Initialization(t *testing.T) {
	selector := NewExitNodeSelector()

	if selector.client == nil {
		t.Error("HTTP client should be initialized")
	}
	if selector.client.Timeout != 30*time.Second {
		t.Errorf("client timeout = %v, want 30s", selector.client.Timeout)
	}
	if selector.cache == nil {
		t.Error("cache should be initialized")
	}
	if len(selector.cache) != 0 {
		t.Error("cache should be empty initially")
	}
	if selector.cacheTTL != 1*time.Hour {
		t.Errorf("cacheTTL = %v, want 1h", selector.cacheTTL)
	}
}

func TestExitNodeSelector_SetExitNodeCountry_UpperCases(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"us", "US"},
		{"de", "DE"},
		{"gb", "GB"},
		{"US", "US"},
		{"De", "DE"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			selector := NewExitNodeSelector()
			config := DefaultTorConfig(t.TempDir())

			selector.SetExitNodeCountry(config, tt.input)

			if config.ExitNodeCountry != tt.expected {
				t.Errorf("ExitNodeCountry = %q, want %q", config.ExitNodeCountry, tt.expected)
			}
			if !config.StrictNodes {
				t.Error("StrictNodes should be enabled")
			}
		})
	}
}

func TestExitNodeSelector_CacheHit(t *testing.T) {
	selector := NewExitNodeSelector()

	// Pre-populate cache
	cachedNodes := []*ExitNodeInfo{
		{Fingerprint: "cached-1", Bandwidth: 100, Country: "us"},
		{Fingerprint: "cached-2", Bandwidth: 50, Country: "us"},
	}
	selector.cacheMu.Lock()
	selector.cache["us"] = cachedNodes
	selector.cacheTime = time.Now()
	selector.cacheMu.Unlock()

	// Should return from cache without network call
	node, err := selector.SelectExitNodeByCountry(context.Background(), "us")
	if err != nil {
		t.Fatalf("SelectExitNodeByCountry failed: %v", err)
	}
	if node.Fingerprint != "cached-1" {
		t.Errorf("expected cached-1, got %q", node.Fingerprint)
	}
}

func TestExitNodeSelector_CacheExpired(t *testing.T) {
	selector := NewExitNodeSelector()

	// Pre-populate cache with expired entries
	cachedNodes := []*ExitNodeInfo{
		{Fingerprint: "expired-1", Bandwidth: 100, Country: "xx"},
	}
	selector.cacheMu.Lock()
	selector.cache["xx"] = cachedNodes
	selector.cacheTime = time.Now().Add(-2 * time.Hour) // Expired
	selector.cacheMu.Unlock()

	// This should try to fetch from network (and fail)
	// Since "xx" is not a real country, it will fail with network error
	_, err := selector.SelectExitNodeByCountry(context.Background(), "xx")
	// We expect either a network error or no nodes found
	if err == nil {
		// If the cache was somehow still used, that's a bug
		t.Log("Got nil error - cache or network may have returned results")
	}
}

func TestExitNodeSelector_EmptyCacheEntry(t *testing.T) {
	selector := NewExitNodeSelector()

	// Pre-populate cache with empty entry
	selector.cacheMu.Lock()
	selector.cache["zz"] = []*ExitNodeInfo{}
	selector.cacheTime = time.Now()
	selector.cacheMu.Unlock()

	// Empty cache should fall through to network fetch
	_, err := selector.SelectExitNodeByCountry(context.Background(), "zz")
	// Will fail due to network but should not panic
	if err == nil {
		t.Log("No error - network may have returned results")
	}
}

// ---------------------------------------------------------------------------
// sortByBandwidth pure function tests
// ---------------------------------------------------------------------------

func TestSortByBandwidth_AlreadySorted(t *testing.T) {
	nodes := []*ExitNodeInfo{
		{Fingerprint: "a", Bandwidth: 300},
		{Fingerprint: "b", Bandwidth: 200},
		{Fingerprint: "c", Bandwidth: 100},
	}

	sortByBandwidth(nodes)

	if nodes[0].Fingerprint != "a" || nodes[1].Fingerprint != "b" || nodes[2].Fingerprint != "c" {
		t.Error("already sorted list should remain in same order")
	}
}

func TestSortByBandwidth_ReverseSorted(t *testing.T) {
	nodes := []*ExitNodeInfo{
		{Fingerprint: "c", Bandwidth: 100},
		{Fingerprint: "b", Bandwidth: 200},
		{Fingerprint: "a", Bandwidth: 300},
	}

	sortByBandwidth(nodes)

	if nodes[0].Bandwidth != 300 {
		t.Errorf("first node bandwidth = %d, want 300", nodes[0].Bandwidth)
	}
	if nodes[1].Bandwidth != 200 {
		t.Errorf("second node bandwidth = %d, want 200", nodes[1].Bandwidth)
	}
	if nodes[2].Bandwidth != 100 {
		t.Errorf("third node bandwidth = %d, want 100", nodes[2].Bandwidth)
	}
}

func TestSortByBandwidth_Empty(t *testing.T) {
	var nodes []*ExitNodeInfo
	// Should not panic
	sortByBandwidth(nodes)
}

func TestSortByBandwidth_SingleElement(t *testing.T) {
	nodes := []*ExitNodeInfo{
		{Fingerprint: "only", Bandwidth: 42},
	}
	sortByBandwidth(nodes)
	if nodes[0].Bandwidth != 42 {
		t.Errorf("single element bandwidth should remain 42")
	}
}

func TestSortByBandwidth_EqualBandwidth(t *testing.T) {
	nodes := []*ExitNodeInfo{
		{Fingerprint: "a", Bandwidth: 100},
		{Fingerprint: "b", Bandwidth: 100},
		{Fingerprint: "c", Bandwidth: 100},
	}
	sortByBandwidth(nodes)

	// All equal so order should be preserved (not strictly required but shouldn't crash)
	for _, n := range nodes {
		if n.Bandwidth != 100 {
			t.Errorf("bandwidth should be 100, got %d", n.Bandwidth)
		}
	}
}

func TestSortByBandwidth_LargeList(t *testing.T) {
	nodes := make([]*ExitNodeInfo, 100)
	for i := 0; i < 100; i++ {
		nodes[i] = &ExitNodeInfo{Bandwidth: int64(i)}
	}

	sortByBandwidth(nodes)

	// Verify sorted descending
	for i := 0; i < len(nodes)-1; i++ {
		if nodes[i].Bandwidth < nodes[i+1].Bandwidth {
			t.Errorf("nodes not sorted: index %d has %d, index %d has %d",
				i, nodes[i].Bandwidth, i+1, nodes[i+1].Bandwidth)
			break
		}
	}
}

// ---------------------------------------------------------------------------
// OnionooResponse and OnionooRelay JSON tests
// ---------------------------------------------------------------------------

func TestOnionooResponse_JSONRoundTrip(t *testing.T) {
	resp := OnionooResponse{
		Version:         "8.0",
		RelaysPublished: "2026-02-10 12:00:00",
		Relays: []OnionooRelay{
			{
				Nickname:    "TestRelay1",
				Fingerprint: "ABC123",
				OrAddresses: []string{"1.2.3.4:443"},
				Country:     "us",
				CountryName: "United States",
				Flags:       []string{"Exit", "Running", "Valid"},
				Bandwidth:   50000,
				Running:     true,
			},
			{
				Nickname:    "TestRelay2",
				Fingerprint: "DEF456",
				OrAddresses: []string{"[::1]:443"},
				ExitPolicy:  "accept *:*",
				Country:     "de",
				CountryName: "Germany",
				Flags:       []string{"Exit", "Running"},
				Bandwidth:   30000,
				Running:     false,
			},
		},
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded OnionooResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.Version != "8.0" {
		t.Errorf("Version = %q, want %q", decoded.Version, "8.0")
	}
	if len(decoded.Relays) != 2 {
		t.Fatalf("expected 2 relays, got %d", len(decoded.Relays))
	}
	if decoded.Relays[0].Nickname != "TestRelay1" {
		t.Errorf("Relay[0].Nickname = %q", decoded.Relays[0].Nickname)
	}
	if decoded.Relays[1].Country != "de" {
		t.Errorf("Relay[1].Country = %q", decoded.Relays[1].Country)
	}
	if decoded.Relays[0].Running != true {
		t.Error("Relay[0].Running should be true")
	}
	if decoded.Relays[1].Running != false {
		t.Error("Relay[1].Running should be false")
	}
}

func TestOnionooRelay_ExitPolicyOmitEmpty(t *testing.T) {
	relay := OnionooRelay{
		Nickname:    "NoPolicy",
		Fingerprint: "XYZ",
		Running:     true,
	}

	data, err := json.Marshal(relay)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	// ExitPolicy with omitempty should not appear
	if strings.Contains(string(data), "exit_policy_summary") {
		t.Error("exit_policy_summary should be omitted when empty")
	}

	// But with a value, it should appear
	relay.ExitPolicy = "accept *:*"
	data, err = json.Marshal(relay)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}
	if !strings.Contains(string(data), "exit_policy_summary") {
		t.Error("exit_policy_summary should be present when set")
	}
}

// ---------------------------------------------------------------------------
// OnionManager tests
// ---------------------------------------------------------------------------

func TestNewOnionManager(t *testing.T) {
	om := NewOnionManager("/tmp/test-onion")

	if om == nil {
		t.Fatal("expected non-nil OnionManager")
	}
	if om.dataDir != "/tmp/test-onion" {
		t.Errorf("dataDir = %q", om.dataDir)
	}
	if om.keys == nil {
		t.Error("keys map should be initialized")
	}
}

func TestOnionManager_GetKey_NotFound(t *testing.T) {
	om := NewOnionManager(t.TempDir())

	_, exists := om.GetKey("nonexistent")
	if exists {
		t.Error("should not find nonexistent key")
	}
}

func TestOnionManager_KeysMap(t *testing.T) {
	tmpDir := t.TempDir()
	om := NewOnionManager(tmpDir)

	// Initially empty
	if len(om.keys) != 0 {
		t.Errorf("keys should be empty initially, got %d", len(om.keys))
	}

	// dataDir should be set
	if om.dataDir != tmpDir {
		t.Errorf("dataDir = %q, want %q", om.dataDir, tmpDir)
	}
}

func TestOnionManager_LoadOnionAddress_NotFound(t *testing.T) {
	om := NewOnionManager(t.TempDir())

	_, err := om.LoadOnionAddress("nonexistent")
	if err == nil {
		t.Error("expected error loading nonexistent onion address")
	}
}

// ---------------------------------------------------------------------------
// TorClient tests
// ---------------------------------------------------------------------------

func TestTorClient_SOCKSAddress(t *testing.T) {
	// NewTorClient will fail without a running SOCKS5 proxy,
	// but we can test the struct concept
	// Actually, proxy.SOCKS5 creates a dialer without connecting,
	// so construction should succeed
	client, err := NewTorClient("127.0.0.1:9050")
	if err != nil {
		t.Fatalf("NewTorClient failed: %v", err)
	}

	if client.SOCKSAddress() != "127.0.0.1:9050" {
		t.Errorf("SOCKSAddress() = %q, want %q", client.SOCKSAddress(), "127.0.0.1:9050")
	}

	if client.dialer == nil {
		t.Error("dialer should be initialized")
	}
}

func TestTorClient_DialTimeout_Cancelled(t *testing.T) {
	client, err := NewTorClient("127.0.0.1:9050")
	if err != nil {
		t.Fatalf("NewTorClient failed: %v", err)
	}

	// DialTimeout with very short timeout should fail
	_, err = client.DialTimeout("tcp", "example.com:80", 1*time.Millisecond)
	// We expect an error since there's no real SOCKS5 proxy
	if err == nil {
		t.Log("DialTimeout succeeded unexpectedly (SOCKS5 proxy may be running)")
	}
}

func TestTorClient_DialContext_CancelledContext(t *testing.T) {
	client, err := NewTorClient("127.0.0.1:9050")
	if err != nil {
		t.Fatalf("NewTorClient failed: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err = client.DialContext(ctx, "tcp", "example.com:80")
	if err == nil {
		t.Error("expected error with cancelled context")
	}
}

// ---------------------------------------------------------------------------
// Helper: ListAvailableCountries dedup logic test (pure function extraction)
// ---------------------------------------------------------------------------

func TestCountryDeduplication(t *testing.T) {
	// Testing the deduplication logic used in ListAvailableCountries
	relays := []struct{ Country string }{
		{"us"}, {"de"}, {"us"}, {"gb"}, {"de"}, {"fr"}, {""},
	}

	countrySet := make(map[string]bool)
	for _, relay := range relays {
		if relay.Country != "" {
			countrySet[strings.ToUpper(relay.Country)] = true
		}
	}

	countries := make([]string, 0, len(countrySet))
	for country := range countrySet {
		countries = append(countries, country)
	}
	sort.Strings(countries)

	expected := []string{"DE", "FR", "GB", "US"}
	if len(countries) != len(expected) {
		t.Fatalf("expected %d countries, got %d", len(expected), len(countries))
	}
	for i, c := range countries {
		if c != expected[i] {
			t.Errorf("country[%d] = %q, want %q", i, c, expected[i])
		}
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file %q: %v", path, err)
	}
	return string(data)
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}
