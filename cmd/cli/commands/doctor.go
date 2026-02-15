package commands

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/moltbunker/moltbunker/internal/doctor"
	"github.com/spf13/cobra"
)

var (
	doctorFix      bool
	doctorDryRun   bool
	doctorJSON     bool
	doctorCategory string
	doctorVerbose  bool
)

func NewDoctorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "System health check (role-aware)",
		Long: `Run diagnostic checks to verify system requirements.

Auto-detects your role from config and only checks relevant dependencies.

Checks for ALL roles:
  - Config file (~/.moltbunker/config.yaml)
  - Node keys (Ed25519 keypair for TLS and P2P identity)
  - Wallet (Ethereum keystore for signing transactions/API requests)
  - Disk space and memory

Additional checks for providers:
  - Colima (macOS) or native containerd (Linux)
  - IPFS for container image distribution
  - Socket permissions for daemon operation

Examples:
  moltbunker doctor              # Auto-detect role, run checks
  moltbunker doctor provider     # Check provider requirements
  moltbunker doctor requester    # Check requester requirements
  moltbunker doctor --fix        # Auto-install missing dependencies
  moltbunker doctor --category config  # Only check configuration`,
		RunE: func(cmd *cobra.Command, args []string) error {
			role := string(GetNodeRole())
			return runDoctorWithRole(role)
		},
	}

	cmd.Flags().BoolVar(&doctorFix, "fix", false, "Automatically install missing dependencies")
	cmd.Flags().BoolVar(&doctorDryRun, "dry-run", false, "Show what would be installed without making changes")
	cmd.Flags().BoolVar(&doctorJSON, "json", false, "Output results as JSON")
	cmd.Flags().StringVar(&doctorCategory, "category", "", "Filter checks by category (runtime, services, system, config, permissions)")
	cmd.Flags().BoolVarP(&doctorVerbose, "verbose", "v", false, "Enable verbose output")

	// Subcommands for explicit role checking
	cmd.AddCommand(&cobra.Command{
		Use:   "provider",
		Short: "Check provider requirements",
		Long: `Run all checks needed to operate as a provider node.

Checks: config, node keys, wallet, Go, Colima/containerd, IPFS,
disk space, memory, file descriptors, socket permissions, Tor.

Use this before 'moltbunker provider enable' to verify your system
has everything needed to host containers and earn BUNKER tokens.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDoctorWithRole("provider")
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "requester",
		Short: "Check requester requirements",
		Long: `Run checks needed to operate as a requester.

Checks: config, node keys, wallet, disk space, memory.
Skips provider-only checks (containerd, IPFS, socket permissions).

A valid wallet is required — the CLI uses it to sign API requests
for deploying containers and managing payments.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDoctorWithRole("requester")
		},
	})

	return cmd
}

func runDoctorWithRole(role string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		cancel()
	}()

	var category doctor.Category
	if doctorCategory != "" {
		switch doctorCategory {
		case "runtime":
			category = doctor.CategoryRuntime
		case "services":
			category = doctor.CategoryServices
		case "system":
			category = doctor.CategorySystem
		case "config":
			category = doctor.CategoryConfig
		case "permissions":
			category = doctor.CategoryPermissions
		default:
			return fmt.Errorf("invalid category: %s (valid: runtime, services, system, config, permissions)", doctorCategory)
		}
	}

	if !doctorJSON {
		label := fmt.Sprintf("role: %s", role)
		fmt.Println()
		fmt.Println(StatusBox(Logo()+" Doctor", [][2]string{
			{"Checking", label},
		}))
		fmt.Println()
	}

	opts := doctor.DoctorOptions{
		Fix:      doctorFix,
		DryRun:   doctorDryRun,
		JSON:     doctorJSON,
		Category: category,
		Role:     role,
		Verbose:  doctorVerbose,
	}

	d := doctor.New(opts)
	report, err := d.Run(ctx)
	if err != nil {
		return fmt.Errorf("doctor check failed: %w", err)
	}

	if !report.Summary.IsHealthy() && !doctorJSON {
		fmt.Println()
		// Show targeted hints based on what failed
		for _, check := range report.Checks {
			if check.Status == doctor.StatusError {
				switch check.Name {
				case "Wallet":
					fmt.Println(Hint("Create a wallet: moltbunker wallet create"))
				case "Node keys":
					fmt.Println(Hint("Generate keys: moltbunker init"))
				case "Config file":
					fmt.Println(Hint("Set up config: moltbunker init"))
				}
			}
		}
		if role == "requester" || role == "" {
			fmt.Println(Hint("To become a provider: moltbunker provider enable"))
		}
		os.Exit(1)
	}

	return nil
}
