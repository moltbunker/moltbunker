package storage

import (
	"errors"
	"fmt"

	"github.com/moltbunker/moltbunker/internal/security"
)

// ErrOwnerKeyNotImplemented is returned by stubbed key stores that do not yet
// support per-wallet key resolution.
var ErrOwnerKeyNotImplemented = errors.New("storage: per-owner X25519 key resolution not implemented")

// OwnerKeyStore resolves the X25519 keypair used to wrap an object's per-object
// DEK. The public key is used to seal (encrypt) the DEK on PutObject; the
// private key is used to open (decrypt) it on GetObject. Both are 32 bytes.
//
// Two models are supported:
//   - ProviderKeyStore (single-node / self-recipient): every owner's objects are
//     sealed to the provider's own stable X25519 key. This matches the R5 image
//     encryption recipient model — a provider's data is readable only by that
//     same provider's daemon.
//   - WalletKeyStore (multi-tenant, future): maps a wallet address to a per-owner
//     keypair from a keyring entry. Stubbed with ErrOwnerKeyNotImplemented so the
//     interface is stable for the follow-up.
type OwnerKeyStore interface {
	// PublicKeyForOwner returns the 32-byte X25519 public key the object's DEK
	// is sealed to.
	PublicKeyForOwner(ownerAddr string) ([]byte, error)
	// PrivateKeyForOwner returns the 32-byte X25519 private key used to open the
	// sealed DEK. Daemon-internal use only; never serialized or logged.
	PrivateKeyForOwner(ownerAddr string) ([]byte, error)
}

// ProviderKeyResolver is the minimal interface ProviderKeyStore needs from the
// daemon's provider key manager. Defined here (rather than importing the daemon
// package) to avoid an import cycle: the concrete *daemon.ProviderKeyManager is
// wired in cmd/daemon/main.go and satisfies this structurally.
type ProviderKeyResolver interface {
	// PublicKey returns a copy of the 32-byte X25519 public key.
	PublicKey() []byte
	// PrivateKey returns a copy of the 32-byte X25519 private key.
	PrivateKey() []byte
}

// ProviderKeyStore implements OwnerKeyStore using the daemon's own stable
// X25519 keypair for every owner (self-recipient model).
type ProviderKeyStore struct {
	resolver ProviderKeyResolver
}

// NewProviderKeyStore builds a self-recipient key store backed by the provider's
// X25519 keypair.
func NewProviderKeyStore(resolver ProviderKeyResolver) (*ProviderKeyStore, error) {
	if resolver == nil {
		return nil, fmt.Errorf("storage: nil provider key resolver")
	}
	if len(resolver.PublicKey()) != security.X25519KeySize {
		return nil, fmt.Errorf("storage: provider public key has invalid size")
	}
	return &ProviderKeyStore{resolver: resolver}, nil
}

// PublicKeyForOwner returns the provider's own public key for any owner.
func (s *ProviderKeyStore) PublicKeyForOwner(_ string) ([]byte, error) {
	pub := s.resolver.PublicKey()
	if len(pub) != security.X25519KeySize {
		return nil, fmt.Errorf("storage: provider public key has invalid size %d", len(pub))
	}
	return pub, nil
}

// PrivateKeyForOwner returns the provider's own private key for any owner.
func (s *ProviderKeyStore) PrivateKeyForOwner(_ string) ([]byte, error) {
	priv := s.resolver.PrivateKey()
	if len(priv) != security.X25519KeySize {
		return nil, fmt.Errorf("storage: provider private key has invalid size %d", len(priv))
	}
	return priv, nil
}

// WalletKeyStore is a placeholder for per-wallet key resolution (multi-tenant).
// It returns ErrOwnerKeyNotImplemented until per-wallet key distribution is
// designed; it exists so the OwnerKeyStore interface is stable for callers.
type WalletKeyStore struct{}

// NewWalletKeyStore returns the stubbed per-wallet key store.
func NewWalletKeyStore() *WalletKeyStore { return &WalletKeyStore{} }

func (s *WalletKeyStore) PublicKeyForOwner(string) ([]byte, error) {
	return nil, ErrOwnerKeyNotImplemented
}

func (s *WalletKeyStore) PrivateKeyForOwner(string) ([]byte, error) {
	return nil, ErrOwnerKeyNotImplemented
}
