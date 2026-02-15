package commands

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/moltbunker/moltbunker/internal/identity"
	"github.com/spf13/cobra"
)

var initNonInteractive bool

func NewInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Interactive setup wizard",
		Long: `Set up moltbunker with a guided wizard.

Walks you through:
  1. Choose your role:
     - Requester: deploy containers, pay BUNKER (no daemon needed)
     - Provider:  host containers, earn BUNKER (daemon required)
     - Hybrid:    both requester and provider
  2. Create or import an Ethereum wallet (required for all roles)
  3. Configure API endpoint (requester) or P2P port (provider)
  4. Generate Ed25519 node identity keys

Use Shift+Tab or arrow keys to go back to previous steps.
Press Ctrl+C at any time to cancel without making changes.

Creates: ~/.moltbunker/config.yaml, ~/.moltbunker/keys/node.key

For non-interactive setup (CI/CD), use: moltbunker install`,
		RunE: runInit,
	}

	cmd.Flags().BoolVar(&initNonInteractive, "non-interactive", false, "Non-interactive mode (same as install)")

	return cmd
}

func runInit(cmd *cobra.Command, args []string) error {
	if initNonInteractive || !isTTY() {
		installCmd := NewInstallCmd()
		return installCmd.RunE(installCmd, args)
	}

	fmt.Println()
	fmt.Println(StatusBox(Logo()+" Setup", [][2]string{
		{"", "Welcome! Let's configure your node."},
		{"", "Use Shift+Tab to go back, Ctrl+C to cancel."},
	}))
	fmt.Println()

	// Check existing state
	homeDir, _ := os.UserHomeDir()
	dataDir := filepath.Join(homeDir, ".moltbunker")
	configPath := filepath.Join(dataDir, "config.yaml")
	keystoreDir := GetKeystoreDir()
	keyPath := filepath.Join(dataDir, "keys", "node.key")

	existingWallet, _ := identity.LoadWalletManager(keystoreDir)
	_, existingKeysErr := os.Stat(keyPath)
	hasExistingKeys := existingKeysErr == nil
	_, existingConfigErr := os.Stat(configPath)
	hasExistingConfig := existingConfigErr == nil

	// Form values
	var (
		role       string
		walletOp   string
		apiURL     string
		p2pPortStr string
		dataPath   string
		genKeys    bool
		overwrite  bool
		confirm    bool
	)

	// Defaults
	dataPath = dataDir
	genKeys = !hasExistingKeys
	if hasExistingKeys {
		genKeys = false
	}

	// Wallet display
	walletDesc := "A wallet is required for signing transactions and API requests"
	if existingWallet != nil {
		walletDesc = fmt.Sprintf("Existing wallet found: %s", FormatAddress(existingWallet.Address().Hex()))
	}

	// Single form, multiple groups — Shift+Tab navigates back between groups
	form := huh.NewForm(
		// Group 1: Role
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("What will you use moltbunker for?").
				Description("This determines which components are needed").
				Options(
					huh.NewOption("Requester — deploy containers, pay BUNKER (no daemon needed)", "requester"),
					huh.NewOption("Provider — host containers, earn BUNKER (daemon required)", "provider"),
					huh.NewOption("Both — full capabilities (daemon required)", "hybrid"),
				).
				Value(&role),
		),

		// Group 2: Wallet (hidden if already configured)
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Wallet setup").
				Description(walletDesc).
				Options(
					huh.NewOption("Create new wallet", "create"),
					huh.NewOption("Import existing private key", "import"),
					huh.NewOption("Skip (configure later with: moltbunker wallet create)", "skip"),
				).
				Value(&walletOp),
		).WithHideFunc(func() bool {
			return existingWallet != nil
		}),

		// Group 3: API endpoint (requester only)
		huh.NewGroup(
			huh.NewInput().
				Title("API endpoint").
				Description("URL of a moltbunker API server (leave empty for default)").
				Placeholder("https://api.moltbunker.com").
				Validate(func(s string) error {
					if s == "" {
						return nil
					}
					if !strings.HasPrefix(s, "http://") && !strings.HasPrefix(s, "https://") {
						return fmt.Errorf("must start with http:// or https://")
					}
					if _, err := url.ParseRequestURI(s); err != nil {
						return fmt.Errorf("invalid URL: %v", err)
					}
					return nil
				}).
				Value(&apiURL),
		).WithHideFunc(func() bool {
			return role != "requester"
		}),

		// Group 4: P2P settings (provider/hybrid only)
		huh.NewGroup(
			huh.NewInput().
				Title("P2P port").
				Description("Port for peer-to-peer connections (default: 9000)").
				Placeholder("9000").
				Validate(func(s string) error {
					if s == "" {
						return nil
					}
					var p int
					if _, err := fmt.Sscanf(s, "%d", &p); err != nil || p < 1024 || p > 65535 {
						return fmt.Errorf("port must be 1024-65535")
					}
					return nil
				}).
				Value(&p2pPortStr),
		).WithHideFunc(func() bool {
			return role == "requester"
		}),

		// Group 5: Advanced options
		huh.NewGroup(
			huh.NewInput().
				Title("Data directory").
				Description("Where to store config, keys, and container data").
				Placeholder(dataDir).
				Value(&dataPath),
			huh.NewConfirm().
				Title("Generate node identity keys?").
				Description("Ed25519 keypair for TLS certificates and P2P identity").
				Affirmative("Yes").
				Negative("No").
				Value(&genKeys),
		),

		// Group 6: Overwrite warning (only if config exists)
		huh.NewGroup(
			huh.NewConfirm().
				Title("Config file already exists. Overwrite?").
				Description(configPath).
				Affirmative("Overwrite").
				Negative("Keep existing").
				Value(&overwrite),
		).WithHideFunc(func() bool {
			return !hasExistingConfig
		}),

		// Group 7: Confirmation with summary
		huh.NewGroup(
			huh.NewConfirm().
				TitleFunc(func() string {
					return "Apply this configuration?"
				}, &role).
				DescriptionFunc(func() string {
					lines := []string{
						fmt.Sprintf("Role:   %s", role),
					}
					if existingWallet != nil {
						lines = append(lines, fmt.Sprintf("Wallet: %s (existing)", FormatAddress(existingWallet.Address().Hex())))
					} else {
						lines = append(lines, fmt.Sprintf("Wallet: %s", walletOp))
					}
					if role == "requester" && apiURL != "" {
						lines = append(lines, fmt.Sprintf("API:    %s", apiURL))
					}
					if role != "requester" {
						port := p2pPortStr
						if port == "" {
							port = "9000"
						}
						lines = append(lines, fmt.Sprintf("P2P:    port %s", port))
					}
					dp := dataPath
					if dp == "" {
						dp = dataDir
					}
					lines = append(lines, fmt.Sprintf("Data:   %s", dp))
					if genKeys {
						lines = append(lines, "Keys:   generate new")
					} else if hasExistingKeys {
						lines = append(lines, "Keys:   keep existing")
					} else {
						lines = append(lines, "Keys:   skip")
					}
					return strings.Join(lines, "\n")
				}, &confirm).
				Affirmative("Confirm").
				Negative("Cancel").
				Value(&confirm),
		),
	).WithTheme(huh.ThemeBase())

	if err := form.Run(); err != nil {
		return err
	}

	if !confirm {
		Info("Setup cancelled — no changes made")
		return nil
	}

	// --- Apply configuration ---

	// Resolve data directory
	if dataPath == "" {
		dataPath = dataDir
	}
	os.MkdirAll(dataPath, 0755)

	// Resolve port
	p2pPort := 9000
	if p2pPortStr != "" {
		fmt.Sscanf(p2pPortStr, "%d", &p2pPort)
	}

	// Step 1: Wallet operation (side-effectful, can't be in the form)
	if existingWallet != nil {
		Success(fmt.Sprintf("Wallet: %s (existing)", existingWallet.Address().Hex()))
	} else {
		switch walletOp {
		case "create":
			walletCmd := newWalletCreateCmd()
			if err := walletCmd.RunE(walletCmd, nil); err != nil {
				Warning(fmt.Sprintf("Wallet creation failed: %v", err))
				fmt.Println(Hint("Create later with: moltbunker wallet create"))
			}
		case "import":
			walletCmd := newWalletImportCmd()
			if err := walletCmd.RunE(walletCmd, nil); err != nil {
				Warning(fmt.Sprintf("Wallet import failed: %v", err))
				fmt.Println(Hint("Import later with: moltbunker wallet import"))
			}
		case "skip":
			Info("Skipping wallet setup")
			fmt.Println(Hint("Create later with: moltbunker wallet create"))
		}
	}

	// Step 2: Write config
	actualConfigPath := filepath.Join(dataPath, "config.yaml")
	if !hasExistingConfig || overwrite {
		configContent := fmt.Sprintf("# moltbunker configuration\n# Generated by: moltbunker init\n\nnode:\n  role: %s\n", role)
		if apiURL != "" {
			configContent += fmt.Sprintf("  api_endpoint: %s\n", apiURL)
		}
		if role != "requester" {
			configContent += fmt.Sprintf("\ndaemon:\n  port: %d\n", p2pPort)
		}
		if err := os.WriteFile(actualConfigPath, []byte(configContent), 0644); err != nil {
			return fmt.Errorf("failed to write config: %w", err)
		}
		Success(fmt.Sprintf("Config written to %s", actualConfigPath))
	} else {
		Info("Config file kept (not overwritten)")
	}

	// Step 3: Generate node keys
	actualKeyPath := filepath.Join(dataPath, "keys", "node.key")
	if genKeys {
		os.MkdirAll(filepath.Dir(actualKeyPath), 0700)
		// Remove empty files (NewKeyManager panics on them)
		if info, err := os.Stat(actualKeyPath); err == nil && info.Size() == 0 {
			os.Remove(actualKeyPath)
		}
		_, err := identity.NewKeyManager(actualKeyPath)
		if err != nil {
			Warning(fmt.Sprintf("Failed to generate keys: %v", err))
		} else {
			Success("Node keys generated")
		}
	}

	// Summary
	fmt.Println()
	summaryFields := [][2]string{
		{"Role", role},
		{"Config", actualConfigPath},
		{"Data", dataPath},
	}
	if apiURL != "" {
		summaryFields = append(summaryFields, [2]string{"API", apiURL})
	}
	if role != "requester" {
		summaryFields = append(summaryFields, [2]string{"P2P port", fmt.Sprintf("%d", p2pPort)})
	}

	fmt.Println(StatusBox("Setup Complete", summaryFields))

	// Next steps
	fmt.Println()
	switch role {
	case "requester":
		fmt.Println(Hint("Next: moltbunker doctor && moltbunker deploy <image>"))
	case "provider", "hybrid":
		fmt.Println(Hint("Next: moltbunker doctor provider && moltbunker start"))
	}

	return nil
}
