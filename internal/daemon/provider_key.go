package daemon

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/moltbunker/moltbunker/internal/logging"
	"github.com/moltbunker/moltbunker/internal/security"
)

// providerKeyFile is the filename (under DataDir) for the persisted stable
// provider X25519 private key used to unwrap E2E-encrypted exec keys.
const providerKeyFile = "provider_x25519.key"

// ProviderKeyManager holds the daemon's STABLE X25519 keypair. The exec key is
// sealed by the CLI to the public key (ECIES) and unwrapped on the daemon with
// the private key before being mounted into the container. The keypair is
// generated on first start and persisted under DataDir with 0600 permissions so
// it survives daemon restarts.
type ProviderKeyManager struct {
	mu      sync.RWMutex
	pubKey  []byte // 32-byte X25519 public key
	privKey []byte // 32-byte X25519 private key (never logged, never leaves the daemon)
	path    string
}

// LoadOrCreateProviderKey loads the persisted X25519 keypair from
// dataDir/provider_x25519.key, generating and persisting a new one if absent.
// The private key file is written with 0600 permissions.
func LoadOrCreateProviderKey(dataDir string) (*ProviderKeyManager, error) {
	if dataDir == "" {
		dataDir = os.TempDir()
	}
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
	}
	path := filepath.Join(dataDir, providerKeyFile)

	pm := &ProviderKeyManager{path: path}

	// Try to load an existing key.
	// #nosec G304 -- path is composed from the daemon-controlled DataDir and a constant filename, not user input.
	if data, err := os.ReadFile(path); err == nil {
		if len(data) != security.X25519KeySize {
			return nil, fmt.Errorf("provider key file %s has invalid size %d (expected %d)", path, len(data), security.X25519KeySize)
		}
		priv := make([]byte, security.X25519KeySize)
		copy(priv, data)
		pub, err := security.X25519PublicFromPrivate(priv)
		if err != nil {
			return nil, fmt.Errorf("derive provider public key: %w", err)
		}
		pm.privKey = priv
		pm.pubKey = pub
		logging.Info("loaded provider X25519 keypair", logging.Component("exec"))
		return pm, nil
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("read provider key file: %w", err)
	}

	// Generate a fresh keypair and persist the private key.
	pub, priv, err := security.GenerateX25519KeyPair()
	if err != nil {
		return nil, fmt.Errorf("generate provider keypair: %w", err)
	}
	if err := os.WriteFile(path, priv, 0600); err != nil {
		return nil, fmt.Errorf("persist provider key: %w", err)
	}
	pm.privKey = priv
	pm.pubKey = pub
	logging.Info("generated new provider X25519 keypair", logging.Component("exec"))
	return pm, nil
}

// PublicKey returns a copy of the 32-byte X25519 public key. Safe to expose to
// requesters/CLI so they can seal the exec key to this provider.
func (pm *ProviderKeyManager) PublicKey() []byte {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	out := make([]byte, len(pm.pubKey))
	copy(out, pm.pubKey)
	return out
}

// privateKey returns the raw 32-byte private key for decryption. Unexported so
// the secret stays inside the daemon package.
func (pm *ProviderKeyManager) privateKey() []byte {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.privKey
}
