//go:build !linux

package identity

import "fmt"

// StoreKernelKeyring is not available on non-Linux platforms.
func StoreKernelKeyring(_ string) error {
	return fmt.Errorf("kernel keyring is only available on Linux")
}

// RetrieveKernelKeyring is not available on non-Linux platforms.
func RetrieveKernelKeyring() (string, error) {
	return "", fmt.Errorf("kernel keyring is only available on Linux")
}

// DeleteKernelKeyring is not available on non-Linux platforms.
func DeleteKernelKeyring() error {
	return fmt.Errorf("kernel keyring is only available on Linux")
}
