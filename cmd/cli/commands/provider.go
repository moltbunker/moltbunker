package commands

import (
	"fmt"

	"github.com/moltbunker/moltbunker/internal/client"
	"github.com/moltbunker/moltbunker/pkg/types"
	"github.com/spf13/cobra"
)

// Provider command flags
var (
	providerTargetTier   string
	providerAutoStake    bool
	providerDeclaredCPU  int
	providerDeclaredMem  int
	providerDeclaredDisk int
	providerGPUEnabled   bool
	providerGPUModel     string
	providerGPUCount     int
)

// NewProviderCmd creates the provider command with subcommands
func NewProviderCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "provider",
		Short: "Provider node management",
		Long: `Commands for managing a provider node on the Moltbunker network.

Providers contribute compute resources and earn BUNKER tokens.
Requires a running daemon and a configured wallet with staked BUNKER.

Getting started:
  1. moltbunker wallet create          # Create a wallet (if not done)
  2. moltbunker doctor provider        # Verify system requirements
  3. moltbunker provider enable        # Guided setup wizard
  4. moltbunker start                  # Start the daemon

Day-to-day:
  moltbunker provider status           # View provider status
  moltbunker provider earnings         # View earnings summary
  moltbunker provider jobs             # List active and recent jobs
  moltbunker provider stake info       # View staking tiers
  moltbunker provider maintenance on   # Enter maintenance mode`,
	}

	cmd.AddCommand(newProviderRegisterCmd())
	cmd.AddCommand(newProviderStatusCmd())
	cmd.AddCommand(newProviderStakeCmd())
	cmd.AddCommand(newProviderEarningsCmd())
	cmd.AddCommand(newProviderJobsCmd())
	cmd.AddCommand(newProviderMaintenanceCmd())
	cmd.AddCommand(newProviderEnableCmd())
	cmd.AddCommand(newProviderDisableCmd())

	// Absorbed commands (formerly top-level)
	cloneCmd := NewCloneCmd()
	cloneCmd.GroupID = ""
	cmd.AddCommand(cloneCmd)

	snapshotCmd := NewSnapshotCmd()
	snapshotCmd.GroupID = ""
	cmd.AddCommand(snapshotCmd)

	threatCmd := NewThreatCmd()
	threatCmd.GroupID = ""
	cmd.AddCommand(threatCmd)

	colimaCmd := NewColimaCmd()
	colimaCmd.GroupID = ""
	cmd.AddCommand(colimaCmd)

	return cmd
}

// --- enable / disable ---

func newProviderEnableCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "enable",
		Short: "Become a provider (guided setup)",
		Long: `Enable provider mode with a guided setup wizard.

Runs health checks, configures containerd, and updates your node role.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			Info("Provider setup wizard")
			Warning("This will update your node role to hybrid (provider + requester)")
			// TODO: Phase 7 — huh wizard with doctor checks, containerd, staking tier
			fmt.Println(Hint("Provider enable wizard not yet implemented"))
			return nil
		},
	}
}

func newProviderDisableCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "disable",
		Short: "Revert to requester mode",
		RunE: func(cmd *cobra.Command, args []string) error {
			Info("Reverting to requester mode...")
			// TODO: Phase 7 — confirmation, config update, daemon stop
			fmt.Println(Hint("Provider disable not yet implemented"))
			return nil
		},
	}
}

// --- register ---

func newProviderRegisterCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register",
		Short: "Register this node as a provider",
		Long: `Register this node as a provider on the network.

This will declare your compute resources and optionally stake tokens.`,
		RunE: runProviderRegister,
	}

	cmd.Flags().IntVar(&providerDeclaredCPU, "cpu", 0, "Declared CPU cores (0 = auto-detect)")
	cmd.Flags().IntVar(&providerDeclaredMem, "memory", 0, "Declared memory in GB (0 = auto-detect)")
	cmd.Flags().IntVar(&providerDeclaredDisk, "disk", 0, "Declared storage in GB (0 = auto-detect)")
	cmd.Flags().StringVar(&providerTargetTier, "tier", "starter", "Target staking tier")
	cmd.Flags().BoolVar(&providerAutoStake, "auto-stake", false, "Auto-stake tokens to reach target tier")
	cmd.Flags().BoolVar(&providerGPUEnabled, "gpu", false, "Enable GPU provider mode")
	cmd.Flags().StringVar(&providerGPUModel, "gpu-model", "", "GPU model")
	cmd.Flags().IntVar(&providerGPUCount, "gpu-count", 1, "Number of GPUs")

	return cmd
}

func runProviderRegister(cmd *cobra.Command, args []string) error {
	c, err := GetClient()
	if err != nil {
		return err
	}
	defer c.Close()

	if err := c.RequireDaemon(); err != nil {
		return err
	}

	tier := types.StakingTier(providerTargetTier)
	validTiers := map[types.StakingTier]bool{
		types.StakingTierStarter:  true,
		types.StakingTierBronze:   true,
		types.StakingTierSilver:   true,
		types.StakingTierGold:     true,
		types.StakingTierPlatinum: true,
	}
	if !validTiers[tier] {
		return fmt.Errorf("invalid tier: %s (must be starter, bronze, silver, gold, or platinum)", providerTargetTier)
	}

	fields := [][2]string{{"Target Tier", string(tier)}}
	if providerDeclaredCPU > 0 {
		fields = append(fields, [2]string{"CPU", fmt.Sprintf("%d cores", providerDeclaredCPU)})
	}
	if providerDeclaredMem > 0 {
		fields = append(fields, [2]string{"Memory", fmt.Sprintf("%d GB", providerDeclaredMem)})
	}
	if providerDeclaredDisk > 0 {
		fields = append(fields, [2]string{"Disk", fmt.Sprintf("%d GB", providerDeclaredDisk)})
	}
	if providerGPUEnabled {
		fields = append(fields, [2]string{"GPU", fmt.Sprintf("%dx %s", providerGPUCount, providerGPUModel)})
	}

	fmt.Println(StatusBox("Registering Provider", fields))

	req := &client.ProviderRegisterRequest{
		DeclaredCPU:    providerDeclaredCPU,
		DeclaredMemory: providerDeclaredMem,
		DeclaredDisk:   providerDeclaredDisk,
		TargetTier:     string(tier),
		AutoStake:      providerAutoStake,
		GPUEnabled:     providerGPUEnabled,
		GPUModel:       providerGPUModel,
		GPUCount:       providerGPUCount,
	}

	dc := c.DaemonClient()
	resp, err := dc.ProviderRegister(req)
	if err != nil {
		return fmt.Errorf("registration failed: %w", err)
	}

	fmt.Println(StatusBox("Registration Successful", [][2]string{
		{"Node ID", FormatNodeID(resp.NodeID)},
		{"Tier", resp.CurrentTier},
		{"Staked", resp.StakedAmount + " BUNKER"},
	}))

	if resp.RequiredStake != "" && resp.RequiredStake != "0" {
		fmt.Println(Hint(fmt.Sprintf("Need %s BUNKER more for %s tier", resp.RequiredStake, tier)))
	}

	Success("Your node is now accepting jobs.")
	fmt.Println(Hint("Monitor with: moltbunker provider status"))

	return nil
}

// --- status ---

func newProviderStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show provider status and statistics",
		RunE:  runProviderStatus,
	}
}

func runProviderStatus(cmd *cobra.Command, args []string) error {
	c, err := GetClient()
	if err != nil {
		return err
	}
	defer c.Close()

	if err := c.RequireDaemon(); err != nil {
		return err
	}

	dc := c.DaemonClient()
	resp, err := dc.ProviderStatus()
	if err != nil {
		return fmt.Errorf("failed to get status: %w", err)
	}

	fmt.Println(StatusBox("Provider", [][2]string{
		{"Node ID", FormatNodeID(resp.NodeID)},
		{"Wallet", FormatAddress(resp.WalletAddress)},
		{"Status", StatusBadge(resp.Status)},
	}))

	fmt.Println(StatusBox("Staking", [][2]string{
		{"Tier", resp.Tier},
		{"Staked", resp.StakedAmount + " BUNKER"},
		{"Next Tier", fmt.Sprintf("%s (requires %s BUNKER)", resp.NextTier, resp.NextTierRequired)},
	}))

	fmt.Println(StatusBox("Performance", [][2]string{
		{"Reputation", fmt.Sprintf("%d/1000 (%s)", resp.ReputationScore, resp.ReputationTier)},
		{"Uptime 30d", fmt.Sprintf("%.2f%%", resp.Uptime30Days)},
		{"Active Jobs", fmt.Sprintf("%d / %d", resp.ActiveJobs, resp.MaxConcurrentJobs)},
		{"Completed", fmt.Sprintf("%d total, %d this week", resp.TotalJobsCompleted, resp.JobsCompleted7Days)},
	}))

	resourceFields := [][2]string{
		{"CPU", fmt.Sprintf("%d cores (%.1f%% used)", resp.DeclaredCPU, resp.CPUUsagePercent)},
		{"Memory", fmt.Sprintf("%d GB (%.1f%% used)", resp.DeclaredMemory, resp.MemoryUsagePercent)},
		{"Disk", fmt.Sprintf("%d GB (%.1f%% used)", resp.DeclaredDisk, resp.DiskUsagePercent)},
	}
	if resp.GPUEnabled {
		resourceFields = append(resourceFields, [2]string{"GPU", fmt.Sprintf("%dx %s", resp.GPUCount, resp.GPUModel)})
	}
	fmt.Println(StatusBox("Resources", resourceFields))

	return nil
}

// --- stake ---

func newProviderStakeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stake",
		Short: "Manage staking (stake, unstake, view)",
	}

	cmd.AddCommand(newProviderStakeAddCmd())
	cmd.AddCommand(newProviderStakeWithdrawCmd())
	cmd.AddCommand(newProviderStakeInfoCmd())

	return cmd
}

func newProviderStakeAddCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add [amount]",
		Short: "Stake additional BUNKER tokens",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			amount := args[0]

			c, err := GetClient()
			if err != nil {
				return err
			}
			defer c.Close()

			if err := c.RequireDaemon(); err != nil {
				return err
			}

			dc := c.DaemonClient()
			var resp *client.ProviderStakeAddResponse
			err = WithSpinner(fmt.Sprintf("Staking %s BUNKER", amount), func() error {
				var e error
				resp, e = dc.ProviderStakeAdd(amount)
				return e
			})
			if err != nil {
				return fmt.Errorf("staking failed: %w", err)
			}

			fmt.Println(StatusBox("Staking Successful", [][2]string{
				{"TX Hash", FormatNodeID(resp.TxHash)},
				{"New Balance", resp.NewStakedAmount + " BUNKER"},
				{"New Tier", resp.NewTier},
			}))

			return nil
		},
	}
}

func newProviderStakeWithdrawCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "withdraw",
		Short: "Initiate unstaking (7-day cooldown)",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := GetClient()
			if err != nil {
				return err
			}
			defer c.Close()

			if err := c.RequireDaemon(); err != nil {
				return err
			}

			Warning("Unstaking has a 7-day cooldown period.")
			fmt.Println(Hint("During cooldown: no new jobs, active jobs must complete"))

			dc := c.DaemonClient()
			resp, err := dc.ProviderStakeWithdraw()
			if err != nil {
				return fmt.Errorf("unstake failed: %w", err)
			}

			fmt.Println(StatusBox("Unstake Initiated", [][2]string{
				{"Initiated", resp.InitiatedAt},
				{"Available", resp.AvailableAt},
				{"Amount", resp.Amount + " BUNKER"},
			}))

			return nil
		},
	}
}

func newProviderStakeInfoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info",
		Short: "View staking tiers and requirements",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(SectionHeader("Staking Tiers"))
			fmt.Println()
			fmt.Println(RenderTable(
				[]string{"Tier", "Min Stake", "Max Jobs", "Multiplier", "Priority"},
				[][]string{
					{"Starter", "1,000,000 BUNKER", "3", "1.0x", "No"},
					{"Bronze", "5,000,000 BUNKER", "10", "1.0x", "No"},
					{"Silver", "10,000,000 BUNKER", "50", "1.05x", "Yes"},
					{"Gold", "100,000,000 BUNKER", "200", "1.1x", "Yes"},
					{"Platinum", "1,000,000,000 BUNKER", "Unlimited", "1.2x", "Yes"},
				},
			))
			fmt.Println(Hint("Higher tiers earn more per job and access premium features."))
		},
	}
}

// --- earnings ---

func newProviderEarningsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "earnings",
		Short: "View earnings summary",
		RunE:  runProviderEarnings,
	}
}

func runProviderEarnings(cmd *cobra.Command, args []string) error {
	c, err := GetClient()
	if err != nil {
		return err
	}
	defer c.Close()

	if err := c.RequireDaemon(); err != nil {
		return err
	}

	dc := c.DaemonClient()
	resp, err := dc.ProviderEarnings()
	if err != nil {
		return fmt.Errorf("failed to get earnings: %w", err)
	}

	fmt.Println(StatusBox("Earnings", [][2]string{
		{"Total", resp.TotalEarnings + " BUNKER"},
		{"30 Days", resp.Earnings30Days + " BUNKER"},
		{"7 Days", resp.Earnings7Days + " BUNKER"},
		{"Pending", resp.PendingPayouts + " BUNKER"},
	}))

	if len(resp.RecentPayments) > 0 {
		fmt.Println(SectionHeader("Recent Payments"))
		var rows [][]string
		for _, p := range resp.RecentPayments {
			rows = append(rows, []string{p.Date, FormatNodeID(p.JobID), p.Amount + " BUNKER", FormatNodeID(p.TxHash)})
		}
		fmt.Println(RenderTable([]string{"Date", "Job", "Amount", "TX"}, rows))
	}

	return nil
}

// --- jobs ---

func newProviderJobsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "jobs",
		Short: "List active and recent jobs",
		RunE:  runProviderJobs,
	}
}

func runProviderJobs(cmd *cobra.Command, args []string) error {
	c, err := GetClient()
	if err != nil {
		return err
	}
	defer c.Close()

	if err := c.RequireDaemon(); err != nil {
		return err
	}

	dc := c.DaemonClient()
	resp, err := dc.ProviderJobs()
	if err != nil {
		return fmt.Errorf("failed to get jobs: %w", err)
	}

	if len(resp.ActiveJobs) > 0 {
		fmt.Println(SectionHeader("Active Jobs"))
		var rows [][]string
		for _, j := range resp.ActiveJobs {
			rows = append(rows, []string{
				FormatNodeID(j.ID),
				j.Image,
				fmt.Sprintf("%.1f%%", j.CPUUsage),
				fmt.Sprintf("%.1f%%", j.MemoryUsage),
				j.StartedAt,
			})
		}
		fmt.Println(RenderTable([]string{"ID", "Image", "CPU", "Memory", "Started"}, rows))
	} else {
		Info("No active jobs.")
	}

	if len(resp.RecentJobs) > 0 {
		fmt.Println(SectionHeader("Recent Jobs (24h)"))
		var rows [][]string
		for _, j := range resp.RecentJobs {
			rows = append(rows, []string{
				FormatNodeID(j.ID),
				StatusBadge(j.Status),
				j.Duration,
				j.Earned + " BUNKER",
			})
		}
		fmt.Println(RenderTable([]string{"ID", "Status", "Duration", "Earned"}, rows))
	}

	return nil
}

// --- maintenance ---

func newProviderMaintenanceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "maintenance",
		Short: "Enter or exit maintenance mode",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "enter",
		Short: "Enter maintenance mode",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := GetClient()
			if err != nil {
				return err
			}
			defer c.Close()
			if err := c.RequireDaemon(); err != nil {
				return err
			}

			dc := c.DaemonClient()
			if err := dc.ProviderMaintenanceMode(true); err != nil {
				return fmt.Errorf("failed to enter maintenance mode: %w", err)
			}

			Success("Maintenance mode enabled")
			fmt.Println(Hint("Your node will no longer accept new jobs."))
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "exit",
		Short: "Exit maintenance mode",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := GetClient()
			if err != nil {
				return err
			}
			defer c.Close()
			if err := c.RequireDaemon(); err != nil {
				return err
			}

			dc := c.DaemonClient()
			if err := dc.ProviderMaintenanceMode(false); err != nil {
				return fmt.Errorf("failed to exit maintenance mode: %w", err)
			}

			Success("Maintenance mode disabled")
			fmt.Println(Hint("Your node is now accepting new jobs."))
			return nil
		},
	})

	return cmd
}
