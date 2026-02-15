package commands

import (
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/moltbunker/moltbunker/internal/identity"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// NewWalletCmd creates the wallet command group
func NewWalletCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wallet",
		Short: "Manage the node's Ethereum wallet",
		Long: `Manage the Ethereum wallet used for staking, payments, and identity.

A wallet is required for ALL roles:
  - Requesters: signs API requests for deploying containers and payments
  - Providers:  binds node identity to on-chain stake (EIP-191 announce)

The wallet is stored as an encrypted keystore file (geth V3 format).
These commands operate directly on keystore files — no daemon needed.

The wallet password is stored in your platform keyring:
  macOS:          Keychain
  Linux (desktop): GNOME Keyring / KDE Wallet
  Linux (server):  kernel keyring (volatile, lost on reboot)

Examples:
  moltbunker wallet create   # Generate a new wallet
  moltbunker wallet import   # Import from a private key
  moltbunker wallet show     # Show address and keystore path
  moltbunker wallet export   # Export private key (use with caution)`,
	}

	cmd.AddCommand(newWalletCreateCmd())
	cmd.AddCommand(newWalletImportCmd())
	cmd.AddCommand(newWalletShowCmd())
	cmd.AddCommand(newWalletExportCmd())
	cmd.AddCommand(newWalletForgetPasswordCmd())

	return cmd
}

func defaultKeystoreDir() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".moltbunker", "keystore")
}

// storePasswordInKeyring attempts to store the wallet password in the best available keyring.
// Tries: platform keyring (Keychain/Secret Service) → kernel keyring → prints fallback instructions.
func storePasswordInKeyring(password string) {
	// Try platform keyring first (macOS Keychain, Linux Secret Service)
	if backend, err := identity.StoreWalletPassword(password); err == nil {
		fmt.Printf("  Password saved to %s\n", backend)
		fmt.Println("  The wallet will be unlocked automatically for all CLI and daemon operations.")
		return
	}

	// Try Linux kernel keyring (headless servers)
	if err := identity.StoreKernelKeyring(password); err == nil {
		fmt.Println("  Password saved to kernel keyring (in-memory, lost on reboot)")
		fmt.Println("  The wallet will be unlocked automatically until next reboot.")
		return
	}

	// No keyring available — print manual instructions
	fmt.Println("  Could not store password in system keyring.")
	fmt.Println("  For automatic wallet unlock, set one of:")
	fmt.Println("    - MOLTBUNKER_WALLET_PASSWORD environment variable")
	fmt.Println("    - node.wallet_password_file in config.yaml")
}

func newWalletCreateCmd() *cobra.Command {
	var keystoreDir string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new wallet",
		Long:  "Create a new Ethereum wallet with a password-encrypted keystore file.",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check if wallet already exists
			wm, err := identity.LoadWalletManager(keystoreDir)
			if err != nil {
				return fmt.Errorf("failed to check keystore: %w", err)
			}
			if wm != nil {
				return fmt.Errorf("wallet already exists at %s (address: %s)", keystoreDir, wm.Address().Hex())
			}

			// Retry loop for password validation
			const maxAttempts = 3
			for attempt := 1; attempt <= maxAttempts; attempt++ {
				fmt.Fprint(os.Stderr, "Enter wallet password: ")
				password, err := readPasswordNoEcho()
				if err != nil {
					return fmt.Errorf("failed to read password: %w", err)
				}
				fmt.Fprintln(os.Stderr)

				if len(password) < 8 {
					Warning("Password must be at least 8 characters. Try again.")
					continue
				}

				fmt.Fprint(os.Stderr, "Confirm wallet password: ")
				confirm, err := readPasswordNoEcho()
				if err != nil {
					return fmt.Errorf("failed to read confirmation: %w", err)
				}
				fmt.Fprintln(os.Stderr)

				if password != confirm {
					Warning("Passwords do not match. Try again.")
					continue
				}

				// Create wallet
				wm, err = identity.CreateWalletManager(keystoreDir, password)
				if err != nil {
					return fmt.Errorf("failed to create wallet: %w", err)
				}

				fmt.Println()
				Success("Wallet created!")
				fmt.Println(StatusBox("Wallet", [][2]string{
					{"Address", wm.Address().Hex()},
					{"Keystore", keystoreDir},
				}))
				storePasswordInKeyring(password)
				fmt.Println()
				Warning("Back up your keystore directory and remember your password.")
				fmt.Println(Hint("If you lose either, your funds are unrecoverable."))
				return nil
			}

			return fmt.Errorf("too many failed attempts")
		},
	}

	cmd.Flags().StringVar(&keystoreDir, "keystore", defaultKeystoreDir(), "Path to keystore directory")

	return cmd
}

func newWalletImportCmd() *cobra.Command {
	var keystoreDir string

	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import a wallet from a private key",
		Long:  "Import an existing Ethereum private key into an encrypted keystore file.",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check if wallet already exists
			wm, err := identity.LoadWalletManager(keystoreDir)
			if err != nil {
				return fmt.Errorf("failed to check keystore: %w", err)
			}
			if wm != nil {
				return fmt.Errorf("wallet already exists at %s (address: %s)", keystoreDir, wm.Address().Hex())
			}

			// Prompt for private key with retry
			const maxAttempts = 3
			var privKeyHex string
			for attempt := 1; attempt <= maxAttempts; attempt++ {
				fmt.Fprint(os.Stderr, "Enter private key (hex, with or without 0x prefix): ")
				input, err := readPasswordNoEcho()
				if err != nil {
					return fmt.Errorf("failed to read private key: %w", err)
				}
				fmt.Fprintln(os.Stderr)

				input = strings.TrimPrefix(input, "0x")
				if len(input) != 64 {
					Warning(fmt.Sprintf("Private key must be 64 hex characters (32 bytes), got %d. Try again.", len(input)))
					continue
				}
				privKeyHex = input
				break
			}
			if privKeyHex == "" {
				return fmt.Errorf("too many failed attempts")
			}

			// Password with retry
			for attempt := 1; attempt <= maxAttempts; attempt++ {
				fmt.Fprint(os.Stderr, "Enter wallet password: ")
				password, err := readPasswordNoEcho()
				if err != nil {
					return fmt.Errorf("failed to read password: %w", err)
				}
				fmt.Fprintln(os.Stderr)

				if len(password) < 8 {
					Warning("Password must be at least 8 characters. Try again.")
					continue
				}

				fmt.Fprint(os.Stderr, "Confirm wallet password: ")
				confirm, err := readPasswordNoEcho()
				if err != nil {
					return fmt.Errorf("failed to read confirmation: %w", err)
				}
				fmt.Fprintln(os.Stderr)

				if password != confirm {
					Warning("Passwords do not match. Try again.")
					continue
				}

				// Import wallet
				wm, err = identity.ImportWalletManager(keystoreDir, privKeyHex, password)
				if err != nil {
					return fmt.Errorf("failed to import wallet: %w", err)
				}

				fmt.Println()
				Success("Wallet imported!")
				fmt.Println(StatusBox("Wallet", [][2]string{
					{"Address", wm.Address().Hex()},
					{"Keystore", keystoreDir},
				}))
				storePasswordInKeyring(password)
				return nil
			}

			return fmt.Errorf("too many failed attempts")
		},
	}

	cmd.Flags().StringVar(&keystoreDir, "keystore", defaultKeystoreDir(), "Path to keystore directory")

	return cmd
}

func newWalletShowCmd() *cobra.Command {
	var keystoreDir string

	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show wallet address and keystore path",
		Long:  "Display the wallet address and keystore directory. No password needed.",
		RunE: func(cmd *cobra.Command, args []string) error {
			wm, err := identity.LoadWalletManager(keystoreDir)
			if err != nil {
				return fmt.Errorf("failed to load wallet: %w", err)
			}
			if wm == nil {
				Info("No wallet found.")
				fmt.Println(Hint("Create one with: moltbunker wallet create"))
				return nil
			}

			pwStatus := "not stored (manual unlock required)"
			if pw, err := identity.RetrieveWalletPassword(); err == nil && pw != "" {
				pwStatus = "stored in platform keyring"
			} else if pw, err := identity.RetrieveKernelKeyring(); err == nil && pw != "" {
				pwStatus = "stored in kernel keyring"
			}

			fmt.Println(StatusBox("Wallet", [][2]string{
				{"Address", wm.Address().Hex()},
				{"Keystore", keystoreDir},
				{"Password", pwStatus},
			}))

			return nil
		},
	}

	cmd.Flags().StringVar(&keystoreDir, "keystore", defaultKeystoreDir(), "Path to keystore directory")

	return cmd
}

func newWalletExportCmd() *cobra.Command {
	var keystoreDir string

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export the wallet's private key",
		Long: `Export the wallet's private key in hex format.

WARNING: The private key controls all funds in this wallet.
Never share it, and clear your terminal history after use.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			wm, err := identity.LoadWalletManager(keystoreDir)
			if err != nil {
				return fmt.Errorf("failed to load wallet: %w", err)
			}
			if wm == nil {
				return fmt.Errorf("no wallet found at %s", keystoreDir)
			}

			fmt.Fprintf(os.Stderr, "WARNING: This will display your private key in plain text.\n")
			fmt.Fprintf(os.Stderr, "Anyone with this key can steal all funds in this wallet.\n\n")

			// Prompt for password
			fmt.Fprint(os.Stderr, "Enter wallet password: ")
			password, err := readPasswordNoEcho()
			if err != nil {
				return fmt.Errorf("failed to read password: %w", err)
			}
			fmt.Fprintln(os.Stderr)

			privKey, err := wm.ExportKey(password)
			if err != nil {
				return fmt.Errorf("failed to export key (wrong password?): %w", err)
			}

			privKeyBytes := crypto.FromECDSA(privKey)
			fmt.Println()
			fmt.Printf("Address:     %s\n", wm.Address().Hex())
			fmt.Printf("Private Key: %s\n", hex.EncodeToString(privKeyBytes))
			fmt.Println()
			fmt.Fprintln(os.Stderr, "Clear your terminal history: history -c && history -w")

			return nil
		},
	}

	cmd.Flags().StringVar(&keystoreDir, "keystore", defaultKeystoreDir(), "Path to keystore directory")

	return cmd
}

func newWalletForgetPasswordCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "forget-password",
		Short: "Remove the wallet password from the system keyring",
		Long: `Remove the stored wallet password from the platform keyring and kernel keyring.

After this, both CLI commands and the daemon will require manual
password configuration (env var or password file) to unlock the wallet.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			removed := false

			// Remove from platform keyring
			if err := identity.DeleteWalletPassword(); err == nil {
				fmt.Println("Removed password from platform keyring")
				removed = true
			}

			// Remove from kernel keyring
			if err := identity.DeleteKernelKeyring(); err == nil {
				fmt.Println("Removed password from kernel keyring")
				removed = true
			}

			if !removed {
				fmt.Println("No stored password found in any keyring.")
			}

			return nil
		},
	}

	return cmd
}

// readPasswordNoEcho reads a line from stdin with echo disabled.
func readPasswordNoEcho() (string, error) {
	password, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", err
	}
	return string(password), nil
}
