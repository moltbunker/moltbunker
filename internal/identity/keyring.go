package identity

import (
	"fmt"
	"runtime"

	"github.com/99designs/keyring"
)

const (
	keyringServiceName = "moltbunker"
	walletPasswordKey  = "wallet-password"
)

// StoreWalletPassword stores the wallet password in the platform keyring.
// On macOS: Keychain. On Linux: Secret Service (GNOME Keyring / KDE Wallet).
// Returns the backend name on success (e.g. "macOS Keychain", "Secret Service").
func StoreWalletPassword(password string) (string, error) {
	ring, backend, err := openKeyring()
	if err != nil {
		return "", err
	}

	err = ring.Set(keyring.Item{
		Key:         walletPasswordKey,
		Data:        []byte(password),
		Label:       "Moltbunker Wallet Password",
		Description: "Password for the moltbunker node wallet keystore",
	})
	if err != nil {
		return "", fmt.Errorf("failed to store in %s: %w", backend, err)
	}

	return backend, nil
}

// RetrieveWalletPassword retrieves the wallet password from the platform keyring.
// Returns ("", nil) if the keyring is available but no password is stored.
// Returns ("", error) if the keyring is not available.
func RetrieveWalletPassword() (string, error) {
	ring, _, err := openKeyring()
	if err != nil {
		return "", err
	}

	item, err := ring.Get(walletPasswordKey)
	if err == keyring.ErrKeyNotFound {
		return "", nil
	}
	if err != nil {
		return "", err
	}

	return string(item.Data), nil
}

// DeleteWalletPassword removes the wallet password from the platform keyring.
func DeleteWalletPassword() error {
	ring, _, err := openKeyring()
	if err != nil {
		return err
	}
	err = ring.Remove(walletPasswordKey)
	if err == keyring.ErrKeyNotFound {
		return nil
	}
	return err
}

// openKeyring opens the platform-native keyring and returns the backend name.
func openKeyring() (keyring.Keyring, string, error) {
	backends := platformKeyringBackends()
	if len(backends) == 0 {
		return nil, "", fmt.Errorf("no keyring backend available on %s", runtime.GOOS)
	}

	ring, err := keyring.Open(keyring.Config{
		ServiceName:                    keyringServiceName,
		AllowedBackends:                backends,
		KeychainTrustApplication:       true,
		KeychainAccessibleWhenUnlocked: true,
		KeychainSynchronizable:         false,
	})
	if err != nil {
		return nil, "", fmt.Errorf("failed to open keyring: %w", err)
	}

	backendName := keyringBackendName()
	return ring, backendName, nil
}

// platformKeyringBackends returns the keyring backends for the current platform.
func platformKeyringBackends() []keyring.BackendType {
	switch runtime.GOOS {
	case "darwin":
		return []keyring.BackendType{keyring.KeychainBackend}
	case "linux":
		return []keyring.BackendType{
			keyring.SecretServiceBackend,
			keyring.KWalletBackend,
		}
	default:
		return nil
	}
}

// keyringBackendName returns a human-readable name for the platform keyring.
func keyringBackendName() string {
	switch runtime.GOOS {
	case "darwin":
		return "macOS Keychain"
	case "linux":
		return "Secret Service (GNOME Keyring / KDE Wallet)"
	default:
		return "system keyring"
	}
}
