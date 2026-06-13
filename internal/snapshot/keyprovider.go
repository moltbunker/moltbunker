package snapshot

import (
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/99designs/keyring"

	"github.com/moltbunker/moltbunker/internal/logging"
	"github.com/moltbunker/moltbunker/internal/security"
)

// KeyProvider supplies the 32-byte AES-256 master key used to encrypt snapshot
// data at rest. It replaces the implicit raw-file-or-silent-ephemeral behavior
// of the old initEncryption() helper: every backend is explicit and named, and
// NONE of them silently generates an ephemeral key that is lost on restart. A
// provider that cannot produce a stable key returns a hard error so the daemon
// fails closed rather than writing unrecoverable snapshots.
//
// Backends:
//   - FileKeyProvider:    32 raw bytes at a 0600 file under DataDir (default).
//   - KeyringKeyProvider: OS keychain via github.com/99designs/keyring
//     (darwin Keychain, linux SecretService/kwallet). Headless servers without
//     a keyring daemon should use the "file" or "env" backend instead.
//   - EnvKeyProvider:     hex-encoded 32 bytes from MOLTBUNKER_SNAPSHOT_KEY,
//     for container/CI environments with no keyring and no writable DataDir.
type KeyProvider interface {
	// MasterKey returns the 32-byte master key, or an error. Implementations
	// must NEVER return a randomly generated key that is not persisted — a key
	// the caller cannot reproduce on the next start silently destroys every
	// snapshot. MasterKey is called on each encrypt/decrypt; implementations
	// that talk to an external store should cache after the first call.
	MasterKey() ([]byte, error)
}

const masterKeySize = 32

// snapshotKeyEnvVar is the environment variable read by EnvKeyProvider.
const snapshotKeyEnvVar = "MOLTBUNKER_SNAPSHOT_KEY"

// keyringService is the service/collection name used for the OS keyring entry.
const keyringService = "moltbunker-snapshot"

// keyringItemKey is the item key (label) for the snapshot master key entry.
const keyringItemKey = "snapshot-master-key"

// --- File backend ---

// FileKeyProvider reads (or, on first use, generates and persists) a 32-byte
// master key at a fixed path with 0600 permissions. This mirrors
// state.LoadOrCreateStateKey and daemon.LoadOrCreateProviderKey: the key
// survives restarts, so snapshots remain decryptable. It is the safe default.
type FileKeyProvider struct {
	path string
	mu   sync.Mutex
	key  []byte
}

// NewFileKeyProvider returns a file-backed provider for the given key path. The
// path is expected to be DataDir-scoped (never request input). The key is not
// read or generated until the first MasterKey call.
func NewFileKeyProvider(path string) (*FileKeyProvider, error) {
	if path == "" {
		return nil, fmt.Errorf("snapshot file key provider: empty key path")
	}
	return &FileKeyProvider{path: path}, nil
}

// MasterKey loads the key from disk, generating and persisting a fresh one if
// the file does not yet exist. A file with the wrong size is a hard error
// (never silently regenerated — that would orphan existing snapshots).
func (p *FileKeyProvider) MasterKey() ([]byte, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.key != nil {
		return cloneKey(p.key), nil
	}

	if err := os.MkdirAll(filepath.Dir(p.path), 0700); err != nil {
		return nil, fmt.Errorf("create snapshot key dir: %w", err)
	}

	// #nosec G304 -- path is DataDir-scoped config, not request input.
	if data, err := os.ReadFile(p.path); err == nil {
		if len(data) != masterKeySize {
			return nil, fmt.Errorf("snapshot key file %s has invalid size %d (expected %d)", p.path, len(data), masterKeySize)
		}
		p.key = cloneKey(data)
		logging.Debug("loaded snapshot master key from file", logging.Component("snapshot"))
		return cloneKey(p.key), nil
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("read snapshot key file: %w", err)
	}

	key, err := security.GenerateKey(masterKeySize)
	if err != nil {
		return nil, fmt.Errorf("generate snapshot key: %w", err)
	}
	if err := os.WriteFile(p.path, key, 0600); err != nil {
		return nil, fmt.Errorf("persist snapshot key: %w", err)
	}
	p.key = cloneKey(key)
	logging.Info("generated new snapshot master key", "path", p.path, logging.Component("snapshot"))
	return cloneKey(p.key), nil
}

// --- Keyring backend ---

// keyringOpener is overridable in tests so the interface contract can be
// exercised against an in-memory keyring without an OS keychain prompt.
var keyringOpener = func(cfg keyring.Config) (keyring.Keyring, error) {
	return keyring.Open(cfg)
}

// KeyringKeyProvider stores the 32-byte master key in the OS keychain. On a
// fully headless Linux server with no SecretService/kwallet daemon, keyring.Open
// may fall back to an encrypted file backend that prompts for a password, which
// is unsuitable for an unattended daemon — use the "file" or "env" backend
// there instead. The key is fetched once and cached in memory.
type KeyringKeyProvider struct {
	service string
	itemKey string
	fileDir string // FileBackend fallback dir (DataDir-scoped)

	mu  sync.Mutex
	key []byte
}

// NewKeyringKeyProvider returns an OS-keyring-backed provider. fileDir is used
// only as the FileBackend fallback location (DataDir-scoped); it never holds a
// plaintext key (the FileBackend is itself encrypted by the keyring library).
func NewKeyringKeyProvider(service, itemKey, fileDir string) *KeyringKeyProvider {
	if service == "" {
		service = keyringService
	}
	if itemKey == "" {
		itemKey = keyringItemKey
	}
	return &KeyringKeyProvider{service: service, itemKey: itemKey, fileDir: fileDir}
}

func (p *KeyringKeyProvider) open() (keyring.Keyring, error) {
	return keyringOpener(keyring.Config{
		ServiceName:                    p.service,
		KeychainName:                   p.service,
		KeychainTrustApplication:       true,
		KeychainAccessibleWhenUnlocked: true,
		LibSecretCollectionName:        p.service,
		FileDir:                        p.fileDir,
		// FilePasswordFunc is only consulted for the encrypted FileBackend
		// fallback; a real headless deployment should use the "file"/"env"
		// snapshot backend rather than rely on this.
		FilePasswordFunc: func(string) (string, error) {
			return "", fmt.Errorf("keyring file backend requires interactive password; use the \"file\" or \"env\" snapshot key backend on headless hosts")
		},
	})
}

// MasterKey fetches the key from the OS keyring, generating and storing a fresh
// one on first use. A failure to open the keyring is a hard error (no silent
// ephemeral fallback).
func (p *KeyringKeyProvider) MasterKey() ([]byte, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.key != nil {
		return cloneKey(p.key), nil
	}

	ring, err := p.open()
	if err != nil {
		return nil, fmt.Errorf("open OS keyring (use \"file\" or \"env\" backend on headless hosts): %w", err)
	}

	item, err := ring.Get(p.itemKey)
	if err == nil {
		if len(item.Data) != masterKeySize {
			return nil, fmt.Errorf("keyring snapshot key has invalid size %d (expected %d)", len(item.Data), masterKeySize)
		}
		p.key = cloneKey(item.Data)
		logging.Debug("loaded snapshot master key from OS keyring", logging.Component("snapshot"))
		return cloneKey(p.key), nil
	}
	if err != keyring.ErrKeyNotFound {
		return nil, fmt.Errorf("read snapshot key from keyring: %w", err)
	}

	// Not present yet: generate and store.
	key, gerr := security.GenerateKey(masterKeySize)
	if gerr != nil {
		return nil, fmt.Errorf("generate snapshot key: %w", gerr)
	}
	if serr := ring.Set(keyring.Item{
		Key:         p.itemKey,
		Data:        key,
		Label:       p.itemKey,
		Description: "moltbunker snapshot master key (AES-256)",
	}); serr != nil {
		return nil, fmt.Errorf("store snapshot key in keyring: %w", serr)
	}
	p.key = cloneKey(key)
	logging.Info("generated new snapshot master key in OS keyring", logging.Component("snapshot"))
	return cloneKey(p.key), nil
}

// --- Env backend ---

// EnvKeyProvider reads the master key from MOLTBUNKER_SNAPSHOT_KEY as a
// hex-encoded 32-byte value. It never touches disk. Intended for container/CI
// environments. A missing or malformed variable is a hard error.
type EnvKeyProvider struct {
	envVar string
}

// NewEnvKeyProvider returns an env-backed provider reading the default env var.
func NewEnvKeyProvider() *EnvKeyProvider {
	return &EnvKeyProvider{envVar: snapshotKeyEnvVar}
}

// MasterKey decodes the hex value from the environment. The decoded value must
// be exactly 32 bytes.
func (p *EnvKeyProvider) MasterKey() ([]byte, error) {
	raw := os.Getenv(p.envVar)
	if raw == "" {
		return nil, fmt.Errorf("snapshot env key provider: %s is not set", p.envVar)
	}
	key, err := hex.DecodeString(raw)
	if err != nil {
		return nil, fmt.Errorf("snapshot env key provider: %s is not valid hex: %w", p.envVar, err)
	}
	if len(key) != masterKeySize {
		return nil, fmt.Errorf("snapshot env key provider: %s decodes to %d bytes (expected %d)", p.envVar, len(key), masterKeySize)
	}
	return key, nil
}

// --- Factory ---

// NewKeyProviderFromConfig builds the KeyProvider selected by the snapshot
// config. Backend selection:
//   - "file" (default): FileKeyProvider at cfg.EncryptionKeyPath, or
//     dataDir/snapshots/.snapshot_key when the path is empty.
//   - "keyring":         KeyringKeyProvider (OS keychain).
//   - "env":             EnvKeyProvider (MOLTBUNKER_SNAPSHOT_KEY).
//
// For back-compat, an empty KeyProviderBackend defaults to "file" using the
// legacy EncryptionKeyPath / StoragePath-derived path so existing deployments
// keep decrypting their snapshots with the same key.
func NewKeyProviderFromConfig(cfg *SnapshotConfig, dataDir string) (KeyProvider, error) {
	if cfg == nil {
		return nil, fmt.Errorf("snapshot key provider: nil config")
	}

	backend := cfg.KeyProviderBackend
	if backend == "" {
		backend = "file"
	}

	switch backend {
	case "file":
		path := cfg.EncryptionKeyPath
		if path == "" {
			base := cfg.StoragePath
			if base == "" && dataDir != "" {
				base = filepath.Join(dataDir, "snapshots")
			}
			if base == "" {
				return nil, fmt.Errorf("snapshot file key provider: no key path and no storage path/data dir to derive one")
			}
			path = filepath.Join(base, ".snapshot_key")
		}
		return NewFileKeyProvider(path)
	case "keyring":
		service := cfg.KeyringServiceName
		if service == "" {
			service = keyringService
		}
		fileDir := cfg.StoragePath
		if fileDir == "" && dataDir != "" {
			fileDir = filepath.Join(dataDir, "snapshots")
		}
		return NewKeyringKeyProvider(service, keyringItemKey, fileDir), nil
	case "env":
		return NewEnvKeyProvider(), nil
	default:
		return nil, fmt.Errorf("snapshot key provider: unknown backend %q (want \"file\", \"keyring\", or \"env\")", backend)
	}
}

// cloneKey returns a defensive copy so the cached key slice is never aliased
// into callers (which could retain or mutate the live secret).
func cloneKey(k []byte) []byte {
	out := make([]byte, len(k))
	copy(out, k)
	return out
}
