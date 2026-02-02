package commands

import (
	"fmt"
	"os"
	"text/tabwriter"

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
		Short: "Provider node management commands",
		Long: `Commands for managing a provider node on the Moltbunker network.

Providers contribute compute resources to the network and earn BUNKER tokens
for running containers. Providers must stake BUNKER to join the network.

Examples:
  moltbunker provider register   # Register as a provider
  moltbunker provider status     # View provider status
  moltbunker provider stake      # Manage staking
  moltbunker provider earnings   # View earnings summary`,
	}

	cmd.AddCommand(newProviderRegisterCmd())
	cmd.AddCommand(newProviderStatusCmd())
	cmd.AddCommand(newProviderStakeCmd())
	cmd.AddCommand(newProviderEarningsCmd())
	cmd.AddCommand(newProviderJobsCmd())
	cmd.AddCommand(newProviderMaintenanceCmd())

	return cmd
}

// newProviderRegisterCmd creates the register subcommand
func newProviderRegisterCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register",
		Short: "Register this node as a provider",
		Long: `Register this node as a provider on the Moltbunker network.

This will:
1. Declare your node's compute resources (CPU, memory, disk)
2. Optionally stake BUNKER tokens to set your tier
3. Start accepting jobs from the network

You must have a configured wallet address to register.

Examples:
  moltbunker provider register --cpu 8 --memory 32 --disk 500
  moltbunker provider register --tier silver --auto-stake
  moltbunker provider register --gpu --gpu-model "RTX 4090" --gpu-count 2`,
		RunE: runProviderRegister,
	}

	cmd.Flags().IntVar(&providerDeclaredCPU, "cpu", 0, "Declared CPU cores (0 = auto-detect)")
	cmd.Flags().IntVar(&providerDeclaredMem, "memory", 0, "Declared memory in GB (0 = auto-detect)")
	cmd.Flags().IntVar(&providerDeclaredDisk, "disk", 0, "Declared storage in GB (0 = auto-detect)")
	cmd.Flags().StringVar(&providerTargetTier, "tier", "starter", "Target staking tier (starter, bronze, silver, gold, platinum)")
	cmd.Flags().BoolVar(&providerAutoStake, "auto-stake", false, "Automatically stake tokens to reach target tier")
	cmd.Flags().BoolVar(&providerGPUEnabled, "gpu", false, "Enable GPU provider mode")
	cmd.Flags().StringVar(&providerGPUModel, "gpu-model", "", "GPU model (e.g., 'RTX 4090', 'A100')")
	cmd.Flags().IntVar(&providerGPUCount, "gpu-count", 1, "Number of GPUs available")

	return cmd
}

func runProviderRegister(cmd *cobra.Command, args []string) error {
	daemonClient := client.NewDaemonClient(SocketPath)
	if err := daemonClient.Connect(); err != nil {
		return fmt.Errorf("daemon not running. Start with 'moltbunker start'")
	}
	defer daemonClient.Close()

	fmt.Println("Registering as Provider")
	fmt.Println("=======================")

	// Validate tier
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

	fmt.Printf("Target Tier:  %s\n", tier)

	if providerDeclaredCPU > 0 {
		fmt.Printf("Declared CPU: %d cores\n", providerDeclaredCPU)
	} else {
		fmt.Println("Declared CPU: (auto-detect)")
	}

	if providerDeclaredMem > 0 {
		fmt.Printf("Declared RAM: %d GB\n", providerDeclaredMem)
	} else {
		fmt.Println("Declared RAM: (auto-detect)")
	}

	if providerDeclaredDisk > 0 {
		fmt.Printf("Declared Disk: %d GB\n", providerDeclaredDisk)
	} else {
		fmt.Println("Declared Disk: (auto-detect)")
	}

	if providerGPUEnabled {
		fmt.Printf("GPU Enabled:  Yes\n")
		fmt.Printf("GPU Model:    %s\n", providerGPUModel)
		fmt.Printf("GPU Count:    %d\n", providerGPUCount)
	}

	// Build registration request
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

	resp, err := daemonClient.ProviderRegister(req)
	if err != nil {
		return fmt.Errorf("registration failed: %w", err)
	}

	fmt.Println()
	fmt.Println("Registration Successful")
	fmt.Println("=======================")
	fmt.Printf("Node ID:       %s\n", resp.NodeID)
	fmt.Printf("Current Tier:  %s\n", resp.CurrentTier)
	fmt.Printf("Staked:        %s BUNKER\n", resp.StakedAmount)

	if resp.RequiredStake != "" && resp.RequiredStake != "0" {
		fmt.Printf("Required:      %s BUNKER (to reach %s tier)\n", resp.RequiredStake, tier)
	}

	fmt.Println()
	fmt.Println("Your node is now accepting jobs from the network.")
	fmt.Println("Use 'moltbunker provider status' to monitor your node.")

	return nil
}

// newProviderStatusCmd creates the status subcommand
func newProviderStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show provider status and statistics",
		Long: `Display detailed status information for your provider node.

Shows:
- Current staking tier and amount
- Reputation score and tier
- Active jobs and capacity
- 30-day uptime and earnings
- Network connectivity

Examples:
  moltbunker provider status`,
		RunE: runProviderStatus,
	}

	return cmd
}

func runProviderStatus(cmd *cobra.Command, args []string) error {
	daemonClient := client.NewDaemonClient(SocketPath)
	if err := daemonClient.Connect(); err != nil {
		return fmt.Errorf("daemon not running. Start with 'moltbunker start'")
	}
	defer daemonClient.Close()

	resp, err := daemonClient.ProviderStatus()
	if err != nil {
		return fmt.Errorf("failed to get status: %w", err)
	}

	fmt.Println("Provider Status")
	fmt.Println("===============")
	fmt.Printf("Node ID:        %s\n", resp.NodeID)
	fmt.Printf("Wallet:         %s\n", resp.WalletAddress)
	fmt.Printf("Status:         %s\n", resp.Status)
	fmt.Println()

	fmt.Println("Staking")
	fmt.Println("-------")
	fmt.Printf("Current Tier:   %s\n", resp.Tier)
	fmt.Printf("Staked Amount:  %s BUNKER\n", resp.StakedAmount)
	fmt.Printf("Next Tier:      %s (requires %s BUNKER)\n", resp.NextTier, resp.NextTierRequired)
	fmt.Println()

	fmt.Println("Reputation")
	fmt.Println("----------")
	fmt.Printf("Score:          %d / 1000\n", resp.ReputationScore)
	fmt.Printf("Tier:           %s\n", resp.ReputationTier)
	fmt.Printf("30-Day Uptime:  %.2f%%\n", resp.Uptime30Days)
	fmt.Println()

	fmt.Println("Capacity")
	fmt.Println("--------")
	fmt.Printf("Active Jobs:    %d / %d\n", resp.ActiveJobs, resp.MaxConcurrentJobs)
	fmt.Printf("Jobs (7 days):  %d completed\n", resp.JobsCompleted7Days)
	fmt.Printf("Jobs (total):   %d completed\n", resp.TotalJobsCompleted)
	fmt.Println()

	fmt.Println("Resources")
	fmt.Println("---------")
	fmt.Printf("CPU:            %d cores (%.1f%% used)\n", resp.DeclaredCPU, resp.CPUUsagePercent)
	fmt.Printf("Memory:         %d GB (%.1f%% used)\n", resp.DeclaredMemory, resp.MemoryUsagePercent)
	fmt.Printf("Disk:           %d GB (%.1f%% used)\n", resp.DeclaredDisk, resp.DiskUsagePercent)
	if resp.GPUEnabled {
		fmt.Printf("GPU:            %dx %s\n", resp.GPUCount, resp.GPUModel)
	}

	return nil
}

// newProviderStakeCmd creates the stake subcommand
func newProviderStakeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stake",
		Short: "Manage staking (stake, unstake, view)",
		Long: `Manage your BUNKER token staking.

Subcommands:
  stake add <amount>     - Stake additional BUNKER tokens
  stake withdraw         - Initiate unstaking (7-day cooldown)
  stake info             - View staking details and tiers

Examples:
  moltbunker provider stake add 1000
  moltbunker provider stake withdraw
  moltbunker provider stake info`,
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

			daemonClient := client.NewDaemonClient(SocketPath)
			if err := daemonClient.Connect(); err != nil {
				return fmt.Errorf("daemon not running")
			}
			defer daemonClient.Close()

			fmt.Printf("Staking %s BUNKER...\n", amount)

			resp, err := daemonClient.ProviderStakeAdd(amount)
			if err != nil {
				return fmt.Errorf("staking failed: %w", err)
			}

			fmt.Println()
			fmt.Println("Staking Successful")
			fmt.Println("==================")
			fmt.Printf("TX Hash:      %s\n", resp.TxHash)
			fmt.Printf("New Balance:  %s BUNKER\n", resp.NewStakedAmount)
			fmt.Printf("New Tier:     %s\n", resp.NewTier)

			return nil
		},
	}
}

func newProviderStakeWithdrawCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "withdraw",
		Short: "Initiate unstaking (7-day cooldown)",
		RunE: func(cmd *cobra.Command, args []string) error {
			daemonClient := client.NewDaemonClient(SocketPath)
			if err := daemonClient.Connect(); err != nil {
				return fmt.Errorf("daemon not running")
			}
			defer daemonClient.Close()

			fmt.Println("Initiating Unstake")
			fmt.Println("==================")
			fmt.Println()
			fmt.Println("WARNING: Unstaking has a 7-day cooldown period.")
			fmt.Println("During this time:")
			fmt.Println("  - You will stop receiving new jobs")
			fmt.Println("  - Active jobs must complete or transfer")
			fmt.Println("  - Your stake remains locked")
			fmt.Println()

			resp, err := daemonClient.ProviderStakeWithdraw()
			if err != nil {
				return fmt.Errorf("unstake initiation failed: %w", err)
			}

			fmt.Printf("Unstake initiated at:    %s\n", resp.InitiatedAt)
			fmt.Printf("Available for withdrawal: %s\n", resp.AvailableAt)
			fmt.Printf("Amount:                   %s BUNKER\n", resp.Amount)

			return nil
		},
	}
}

func newProviderStakeInfoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info",
		Short: "View staking tiers and requirements",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Staking Tiers")
			fmt.Println("=============")
			fmt.Println()

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "Tier\tMin Stake\tMax Jobs\tMultiplier\tPriority\tGovernance")
			fmt.Fprintln(w, "----\t---------\t--------\t----------\t--------\t----------")
			fmt.Fprintln(w, "Starter\t500 BUNKER\t3\t1.0x\tNo\tNo")
			fmt.Fprintln(w, "Bronze\t2,000 BUNKER\t10\t1.0x\tNo\tNo")
			fmt.Fprintln(w, "Silver\t10,000 BUNKER\t50\t1.05x\tYes\tNo")
			fmt.Fprintln(w, "Gold\t50,000 BUNKER\t200\t1.1x\tYes\tYes")
			fmt.Fprintln(w, "Platinum\t250,000 BUNKER\tUnlimited\t1.2x\tYes\tYes")
			w.Flush()

			fmt.Println()
			fmt.Println("Note: Higher tiers earn more per job and have access to premium features.")

			return nil
		},
	}
}

// newProviderEarningsCmd creates the earnings subcommand
func newProviderEarningsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "earnings",
		Short: "View earnings summary and history",
		Long: `Display earnings information for your provider node.

Shows:
- Total lifetime earnings
- 30-day earnings
- 7-day earnings
- Pending payouts
- Recent job payments

Examples:
  moltbunker provider earnings`,
		RunE: runProviderEarnings,
	}

	return cmd
}

func runProviderEarnings(cmd *cobra.Command, args []string) error {
	daemonClient := client.NewDaemonClient(SocketPath)
	if err := daemonClient.Connect(); err != nil {
		return fmt.Errorf("daemon not running")
	}
	defer daemonClient.Close()

	resp, err := daemonClient.ProviderEarnings()
	if err != nil {
		return fmt.Errorf("failed to get earnings: %w", err)
	}

	fmt.Println("Earnings Summary")
	fmt.Println("================")
	fmt.Printf("Total Earnings:     %s BUNKER\n", resp.TotalEarnings)
	fmt.Printf("30-Day Earnings:    %s BUNKER\n", resp.Earnings30Days)
	fmt.Printf("7-Day Earnings:     %s BUNKER\n", resp.Earnings7Days)
	fmt.Printf("Pending Payouts:    %s BUNKER\n", resp.PendingPayouts)
	fmt.Println()

	if len(resp.RecentPayments) > 0 {
		fmt.Println("Recent Payments")
		fmt.Println("---------------")
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "Date\tJob ID\tAmount\tTX Hash")
		for _, p := range resp.RecentPayments {
			fmt.Fprintf(w, "%s\t%s\t%s BUNKER\t%s\n", p.Date, p.JobID, p.Amount, p.TxHash[:16]+"...")
		}
		w.Flush()
	}

	return nil
}

// newProviderJobsCmd creates the jobs subcommand
func newProviderJobsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "jobs",
		Short: "List active and recent jobs",
		Long: `Display jobs running on your provider node.

Shows:
- Currently active jobs
- Recent completed jobs
- Job resource usage

Examples:
  moltbunker provider jobs`,
		RunE: runProviderJobs,
	}

	return cmd
}

func runProviderJobs(cmd *cobra.Command, args []string) error {
	daemonClient := client.NewDaemonClient(SocketPath)
	if err := daemonClient.Connect(); err != nil {
		return fmt.Errorf("daemon not running")
	}
	defer daemonClient.Close()

	resp, err := daemonClient.ProviderJobs()
	if err != nil {
		return fmt.Errorf("failed to get jobs: %w", err)
	}

	if len(resp.ActiveJobs) > 0 {
		fmt.Println("Active Jobs")
		fmt.Println("===========")
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tImage\tCPU\tMemory\tRunning Since")
		for _, j := range resp.ActiveJobs {
			fmt.Fprintf(w, "%s\t%s\t%.1f%%\t%.1f%%\t%s\n",
				j.ID[:12], j.Image, j.CPUUsage, j.MemoryUsage, j.StartedAt)
		}
		w.Flush()
		fmt.Println()
	} else {
		fmt.Println("No active jobs.")
		fmt.Println()
	}

	if len(resp.RecentJobs) > 0 {
		fmt.Println("Recent Jobs (Last 24h)")
		fmt.Println("======================")
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tStatus\tDuration\tEarned")
		for _, j := range resp.RecentJobs {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s BUNKER\n",
				j.ID[:12], j.Status, j.Duration, j.Earned)
		}
		w.Flush()
	}

	return nil
}

// newProviderMaintenanceCmd creates the maintenance subcommand
func newProviderMaintenanceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "maintenance",
		Short: "Enter or exit maintenance mode",
		Long: `Control maintenance mode for your provider node.

In maintenance mode:
- No new jobs are accepted
- Existing jobs continue running
- Your reputation is not affected by not accepting new jobs

Subcommands:
  maintenance enter    - Enter maintenance mode
  maintenance exit     - Exit maintenance mode
  maintenance status   - Check current mode

Examples:
  moltbunker provider maintenance enter
  moltbunker provider maintenance exit`,
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "enter",
		Short: "Enter maintenance mode",
		RunE: func(cmd *cobra.Command, args []string) error {
			daemonClient := client.NewDaemonClient(SocketPath)
			if err := daemonClient.Connect(); err != nil {
				return fmt.Errorf("daemon not running")
			}
			defer daemonClient.Close()

			if err := daemonClient.ProviderMaintenanceMode(true); err != nil {
				return fmt.Errorf("failed to enter maintenance mode: %w", err)
			}

			fmt.Println("Maintenance mode ENABLED")
			fmt.Println("Your node will no longer accept new jobs.")
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "exit",
		Short: "Exit maintenance mode",
		RunE: func(cmd *cobra.Command, args []string) error {
			daemonClient := client.NewDaemonClient(SocketPath)
			if err := daemonClient.Connect(); err != nil {
				return fmt.Errorf("daemon not running")
			}
			defer daemonClient.Close()

			if err := daemonClient.ProviderMaintenanceMode(false); err != nil {
				return fmt.Errorf("failed to exit maintenance mode: %w", err)
			}

			fmt.Println("Maintenance mode DISABLED")
			fmt.Println("Your node is now accepting new jobs.")
			return nil
		},
	})

	return cmd
}
