//go:build linux

package identity

import (
	"fmt"
	"os/exec"
	"strings"
)

const kernelKeyringKeyName = "moltbunker-wallet"

// StoreKernelKeyring stores the wallet password in the Linux kernel keyring.
// The key lives in kernel memory only (user session keyring) and is lost on reboot.
// Requires the `keyctl` command (package: keyutils).
func StoreKernelKeyring(password string) error {
	cmd := exec.Command("keyctl", "padd", "user", kernelKeyringKeyName, "@u")
	cmd.Stdin = strings.NewReader(password)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("keyctl padd failed: %w (output: %s)", err, strings.TrimSpace(string(out)))
	}
	return nil
}

// RetrieveKernelKeyring retrieves the wallet password from the Linux kernel keyring.
// Returns ("", error) if the key is not found or keyctl is not available.
func RetrieveKernelKeyring() (string, error) {
	// Search for the key in the user session keyring
	out, err := exec.Command("keyctl", "search", "@u", "user", kernelKeyringKeyName).Output()
	if err != nil {
		return "", fmt.Errorf("keyctl search failed: %w", err)
	}
	keyID := strings.TrimSpace(string(out))

	// Read the key payload
	out, err = exec.Command("keyctl", "pipe", keyID).Output()
	if err != nil {
		return "", fmt.Errorf("keyctl pipe failed: %w", err)
	}
	return string(out), nil
}

// DeleteKernelKeyring removes the wallet password from the Linux kernel keyring.
func DeleteKernelKeyring() error {
	out, err := exec.Command("keyctl", "search", "@u", "user", kernelKeyringKeyName).Output()
	if err != nil {
		return nil // Key not found, nothing to delete
	}
	keyID := strings.TrimSpace(string(out))
	_, err = exec.Command("keyctl", "unlink", keyID, "@u").Output()
	return err
}
