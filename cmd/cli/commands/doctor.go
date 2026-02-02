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
		Short: "Check system health and dependencies",
		Long: `Run diagnostic checks to verify system requirements and dependencies.

The doctor command checks for:
- Runtime dependencies (Go, containerd)
- Service availability (Tor, IPFS)
- System requirements (disk space, memory)
- Configuration files
- Socket permissions

Examples:
  moltbunker doctor              # Run all checks
  moltbunker doctor --fix        # Auto-install missing dependencies
  moltbunker doctor --dry-run    # Preview what would be installed
  moltbunker doctor --json       # Output results as JSON
  moltbunker doctor --category runtime  # Only check runtime dependencies`,
		RunE: runDoctor,
	}

	cmd.Flags().BoolVar(&doctorFix, "fix", false, "Automatically install missing dependencies")
	cmd.Flags().BoolVar(&doctorDryRun, "dry-run", false, "Show what would be installed without making changes")
	cmd.Flags().BoolVar(&doctorJSON, "json", false, "Output results as JSON")
	cmd.Flags().StringVar(&doctorCategory, "category", "", "Filter checks by category (runtime, services, system, config, permissions)")
	cmd.Flags().BoolVarP(&doctorVerbose, "verbose", "v", false, "Enable verbose output")

	return cmd
}

func runDoctor(cmd *cobra.Command, args []string) error {
	// Create context with signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		cancel()
	}()

	// Validate category if provided
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

	opts := doctor.DoctorOptions{
		Fix:      doctorFix,
		DryRun:   doctorDryRun,
		JSON:     doctorJSON,
		Category: category,
		Verbose:  doctorVerbose,
	}

	d := doctor.New(opts)
	report, err := d.Run(ctx)
	if err != nil {
		return fmt.Errorf("doctor check failed: %w", err)
	}

	// Exit with error code if checks failed (for CI/CD usage)
	if !report.Summary.IsHealthy() && !doctorJSON {
		os.Exit(1)
	}

	return nil
}
