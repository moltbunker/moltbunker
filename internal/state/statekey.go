package state

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/moltbunker/moltbunker/internal/logging"
	"github.com/moltbunker/moltbunker/internal/security"
)

// stateKeyFile is the filename (under DataDir) for the persisted symmetric key
// used to encrypt bbolt state values at rest (R8).
const stateKeyFile = "state.key"

// stateKeySize is the AES-256 key length in bytes for at-rest state encryption.
const stateKeySize = 32

// Threat model (R8): at-rest encryption of state values with an on-disk key
// protects against a stolen disk, a leaked backup, or casual filesystem access
// where the attacker obtains moltbunker.db but NOT the daemon's runtime
// environment. It does NOT defend against a live host-root attacker who can read
// both moltbunker.db and state.key (sitting beside it with 0600 perms) — that
// threat requires hardware-backed key custody (SEV-SNP / TPM) and is out of
// scope for this layer.

// LoadOrCreateStateKey loads the persisted 32-byte state-encryption key from
// dataDir/state.key, generating and persisting a fresh random key if absent. The
// key file is written with 0600 permissions. The returned key is never logged.
// Mirrors the persistence pattern of internal/daemon/provider_key.go.
func LoadOrCreateStateKey(dataDir string) ([]byte, error) {
	if dataDir == "" {
		dataDir = os.TempDir()
	}
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
	}
	path := filepath.Join(dataDir, stateKeyFile)

	// Try to load an existing key.
	// #nosec G304 -- path is DataDir + constant filename, not user input.
	if data, err := os.ReadFile(path); err == nil {
		if len(data) != stateKeySize {
			return nil, fmt.Errorf("state key file %s has invalid size %d (expected %d)", path, len(data), stateKeySize)
		}
		key := make([]byte, stateKeySize)
		copy(key, data)
		logging.Info("loaded state-at-rest encryption key", logging.Component("state"))
		return key, nil
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("read state key file: %w", err)
	}

	// Generate a fresh key and persist it.
	key, err := security.GenerateKey(stateKeySize)
	if err != nil {
		return nil, fmt.Errorf("generate state key: %w", err)
	}
	if err := os.WriteFile(path, key, 0600); err != nil {
		return nil, fmt.Errorf("persist state key: %w", err)
	}
	logging.Info("generated new state-at-rest encryption key", logging.Component("state"))
	return key, nil
}
