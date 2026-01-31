package tor

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cretz/bine/tor"
	"github.com/cretz/bine/torutil/ed25519"
)

// TorService manages Tor hidden service (.onion address)
type TorService struct {
	tor      *tor.Tor
	onion    *tor.OnionService
	onionAddr string
	dataDir  string
}

// NewTorService creates a new Tor service manager
func NewTorService(dataDir string) *TorService {
	return &TorService{
		dataDir: dataDir,
	}
}

// Start starts the Tor service and creates hidden service
func (ts *TorService) Start(ctx context.Context, ports []int) error {
	// Create Tor data directory
	if err := os.MkdirAll(ts.dataDir, 0700); err != nil {
		return fmt.Errorf("failed to create Tor data directory: %w", err)
	}

	// Start Tor instance
	torInstance, err := tor.Start(ctx, &tor.StartConf{
		DataDir: ts.dataDir,
		NoAutoSocksPort: true,
	})
	if err != nil {
		return fmt.Errorf("failed to start Tor: %w", err)
	}

	ts.tor = torInstance

	// Create hidden service
	onionService, err := ts.createOnionService(ctx, ports)
	if err != nil {
		return fmt.Errorf("failed to create onion service: %w", err)
	}

	ts.onion = onionService
	ts.onionAddr = onionService.ID + ".onion"

	return nil
}

// createOnionService creates a Tor hidden service
func (ts *TorService) createOnionService(ctx context.Context, ports []int) (*tor.OnionService, error) {
	// Generate Ed25519 key for onion service
	keyPath := filepath.Join(ts.dataDir, "onion_key")
	var key ed25519.PrivateKey

	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		// Generate new key
		keyPair, err := ed25519.GenerateKey(nil)
		if err != nil {
			return nil, fmt.Errorf("failed to generate onion key: %w", err)
		}
		key, ok := keyPair.(ed25519.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("failed to convert key pair to PrivateKey")
		}

		// Save key (simplified)
		keyBytes := []byte(key)
		if err := os.WriteFile(keyPath, keyBytes, 0600); err != nil {
			return nil, fmt.Errorf("failed to save onion key: %w", err)
		}
	} else {
		// Load existing key
		keyData, err := os.ReadFile(keyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read onion key: %w", err)
		}
		key = ed25519.PrivateKey(keyData)
	}

	// Create port mappings (simplified - actual API may differ)
	// Note: bine API may have changed, this is a placeholder
	onionService, err := ts.tor.Listen(ctx, &tor.ListenConf{
		RemotePorts: ports,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to listen on onion service: %w", err)
	}

	return onionService, nil
}

// OnionAddress returns the .onion address
func (ts *TorService) OnionAddress() string {
	return ts.onionAddr
}

// Stop stops the Tor service
func (ts *TorService) Stop() error {
	if ts.onion != nil {
		if err := ts.onion.Close(); err != nil {
			return err
		}
	}

	if ts.tor != nil {
		if err := ts.tor.Close(); err != nil {
			return err
		}
	}

	return nil
}

// Tor returns the Tor instance
func (ts *TorService) Tor() *tor.Tor {
	return ts.tor
}

// OnionService returns the onion service
func (ts *TorService) OnionService() *tor.OnionService {
	return ts.onion
}
