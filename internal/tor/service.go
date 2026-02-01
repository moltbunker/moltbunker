package tor

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/cretz/bine/tor"
	"github.com/cretz/bine/torutil/ed25519"
)

// TorService manages Tor hidden service (.onion address)
type TorService struct {
	config       *TorConfig
	tor          *tor.Tor
	onion        *tor.OnionService
	onionAddr    string
	running      bool
	mu           sync.RWMutex
	onionServices map[int]*tor.OnionService
}

// NewTorService creates a new Tor service manager
func NewTorService(config *TorConfig) (*TorService, error) {
	if config == nil {
		return nil, fmt.Errorf("config required")
	}

	// Ensure data directory exists
	if err := os.MkdirAll(config.DataDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create Tor data directory: %w", err)
	}

	return &TorService{
		config:        config,
		onionServices: make(map[int]*tor.OnionService),
	}, nil
}

// Start starts the Tor service
func (ts *TorService) Start(ctx context.Context) error {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.running {
		return nil // Already running
	}

	// Start Tor instance
	startConf := &tor.StartConf{
		DataDir:         ts.config.DataDir,
		NoAutoSocksPort: !ts.config.EnableSocks,
	}

	torInstance, err := tor.Start(ctx, startConf)
	if err != nil {
		return fmt.Errorf("failed to start Tor: %w", err)
	}

	ts.tor = torInstance
	ts.running = true

	return nil
}

// Stop stops the Tor service
func (ts *TorService) Stop() error {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if !ts.running {
		return nil
	}

	// Close all onion services
	for _, service := range ts.onionServices {
		// Silently close onion services - errors are non-fatal during shutdown
		service.Close()
	}
	ts.onionServices = make(map[int]*tor.OnionService)

	// Close main onion service
	if ts.onion != nil {
		ts.onion.Close()
		ts.onion = nil
	}

	// Close Tor instance
	if ts.tor != nil {
		if err := ts.tor.Close(); err != nil {
			return fmt.Errorf("failed to close Tor: %w", err)
		}
		ts.tor = nil
	}

	ts.running = false
	ts.onionAddr = ""

	return nil
}

// IsRunning returns whether Tor is running
func (ts *TorService) IsRunning() bool {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return ts.running
}

// GetOnionAddress returns the main .onion address
func (ts *TorService) GetOnionAddress() string {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return ts.onionAddr
}

// CreateOnionService creates a Tor hidden service for a specific port
func (ts *TorService) CreateOnionService(ctx context.Context, port int) (string, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if !ts.running {
		return "", fmt.Errorf("Tor not running")
	}

	// Check if we already have a service for this port
	if service, exists := ts.onionServices[port]; exists {
		return service.ID + ".onion", nil
	}

	// Generate or load key
	keyPath := filepath.Join(ts.config.DataDir, fmt.Sprintf("onion_key_%d", port))
	privateKey, err := ts.loadOrGenerateKey(keyPath)
	if err != nil {
		return "", err
	}

	// Create onion service
	onionService, err := ts.tor.Listen(ctx, &tor.ListenConf{
		RemotePorts: []int{port},
		Key:         privateKey,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create onion service: %w", err)
	}

	ts.onionServices[port] = onionService
	addr := onionService.ID + ".onion"

	// Set as main address if this is the first one
	if ts.onionAddr == "" {
		ts.onionAddr = addr
		ts.onion = onionService
	}

	return addr, nil
}

// loadOrGenerateKey loads or generates an Ed25519 key for onion service
func (ts *TorService) loadOrGenerateKey(keyPath string) (ed25519.PrivateKey, error) {
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		// Generate new key
		keyPair, err := ed25519.GenerateKey(nil)
		if err != nil {
			return nil, fmt.Errorf("failed to generate onion key: %w", err)
		}

		privateKey, ok := keyPair.(ed25519.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("failed to convert key pair to PrivateKey")
		}

		// Save key
		if err := os.WriteFile(keyPath, []byte(privateKey), 0600); err != nil {
			return nil, fmt.Errorf("failed to save onion key: %w", err)
		}

		return privateKey, nil
	}

	// Load existing key
	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read onion key: %w", err)
	}

	return ed25519.PrivateKey(keyData), nil
}

// RotateCircuit requests a new Tor circuit
func (ts *TorService) RotateCircuit(ctx context.Context) error {
	ts.mu.RLock()
	torInstance := ts.tor
	running := ts.running
	ts.mu.RUnlock()

	if !running || torInstance == nil {
		return fmt.Errorf("Tor not running")
	}

	// Use control connection to signal NEWNYM (new identity)
	control := torInstance.Control
	if control == nil {
		return fmt.Errorf("Tor control connection not available")
	}

	// Send SIGNAL NEWNYM to get new circuits
	if err := control.Signal("NEWNYM"); err != nil {
		return fmt.Errorf("failed to rotate circuit: %w", err)
	}

	return nil
}

// GetCircuitInfo returns information about current circuits
func (ts *TorService) GetCircuitInfo(ctx context.Context) ([]CircuitInfo, error) {
	ts.mu.RLock()
	torInstance := ts.tor
	running := ts.running
	ts.mu.RUnlock()

	if !running || torInstance == nil {
		return nil, fmt.Errorf("Tor not running")
	}

	control := torInstance.Control
	if control == nil {
		return nil, fmt.Errorf("Tor control connection not available")
	}

	// Get circuit status
	circuits, err := control.GetInfo("circuit-status")
	if err != nil {
		return nil, fmt.Errorf("failed to get circuit status: %w", err)
	}

	// Parse circuit info (simplified)
	var result []CircuitInfo
	for _, c := range circuits {
		result = append(result, CircuitInfo{
			Raw: c.Val,
		})
	}

	return result, nil
}

// CircuitInfo contains information about a Tor circuit
type CircuitInfo struct {
	ID     string
	Status string
	Path   []string
	Raw    string
}

// OnionAddress returns the .onion address (legacy method)
func (ts *TorService) OnionAddress() string {
	return ts.GetOnionAddress()
}

// Tor returns the Tor instance
func (ts *TorService) Tor() *tor.Tor {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return ts.tor
}

// OnionService returns the main onion service
func (ts *TorService) OnionService() *tor.OnionService {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return ts.onion
}

// CloseOnionService closes a specific onion service
func (ts *TorService) CloseOnionService(port int) error {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	service, exists := ts.onionServices[port]
	if !exists {
		return nil
	}

	if err := service.Close(); err != nil {
		return err
	}

	delete(ts.onionServices, port)
	return nil
}

// ListOnionServices returns all active onion service addresses
func (ts *TorService) ListOnionServices() map[int]string {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	result := make(map[int]string)
	for port, service := range ts.onionServices {
		result[port] = service.ID + ".onion"
	}
	return result
}
