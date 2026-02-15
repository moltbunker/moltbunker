package main

import (
	"fmt"
	"os"

	"github.com/moltbunker/moltbunker/cmd/cli/commands"
	"github.com/spf13/cobra"
)

// Command group IDs
const (
	groupGettingStarted = "getting-started"
	groupContainers     = "containers"
	groupWallet         = "wallet"
	groupNetwork        = "network"
	groupProvider       = "provider"
	groupAdvanced       = "advanced"
)

var rootCmd = &cobra.Command{
	Use:   "moltbunker",
	Short: "Moltbunker — P2P Encrypted Container Runtime",
	Long:  "A permissionless, fully encrypted P2P network for containerized compute resources.",
}

func init() {
	// Persistent flags
	rootCmd.PersistentFlags().StringVar(&commands.SocketPath, "socket", "", "Path to daemon socket (default: auto-detect)")
	rootCmd.PersistentFlags().StringVar(&commands.APIEndpoint, "api", "", "HTTP API endpoint URL for daemonless mode")
	rootCmd.PersistentFlags().StringVar(&commands.OutputFormat, "output", "", "Output format: json, plain (default: auto)")

	// Command groups
	rootCmd.AddGroup(
		&cobra.Group{ID: groupGettingStarted, Title: "Getting Started:"},
		&cobra.Group{ID: groupContainers, Title: "Containers:"},
		&cobra.Group{ID: groupWallet, Title: "Wallet & Payments:"},
		&cobra.Group{ID: groupNetwork, Title: "Network:"},
		&cobra.Group{ID: groupProvider, Title: "Provider:"},
		&cobra.Group{ID: groupAdvanced, Title: "Advanced:"},
	)
}

func main() {
	// Getting Started
	addCmd(commands.NewInitCmd(), groupGettingStarted)
	addCmd(commands.NewDoctorCmd(), groupGettingStarted)
	addCmd(commands.NewStartCmd(), groupGettingStarted)
	addCmd(commands.NewStopCmd(), groupGettingStarted)

	// Containers
	addCmd(commands.NewDeployCmd(), groupContainers)
	addCmd(commands.NewPsCmd(), groupContainers)
	addCmd(commands.NewLogsCmd(), groupContainers)
	addCmd(commands.NewExecCmd(), groupContainers)

	// Wallet & Payments
	addCmd(commands.NewWalletCmd(), groupWallet)
	addCmd(commands.NewBalanceCmd(), groupWallet)

	// Network
	addCmd(commands.NewStatusCmd(), groupNetwork)
	addCmd(commands.NewPeersCmd(), groupNetwork)
	addCmd(commands.NewMonitorCmd(), groupNetwork)

	// Provider (shown to all, but some subcommands require daemon)
	addCmd(commands.NewProviderCmd(), groupProvider)

	// Advanced
	addCmd(commands.NewConfigCmd(), groupAdvanced)
	addCmd(commands.NewTorCmd(), groupAdvanced)
	addCmd(commands.NewVersionCmd(), groupAdvanced)
	addCmd(commands.NewOutputsCmd(), groupAdvanced)
	addCmd(commands.NewCompletionCmd(), groupAdvanced)
	addCmd(commands.NewManCmd(), groupAdvanced)

	// Hidden/deprecated — kept for backward compatibility
	hidden(commands.NewInstallCmd())   // replaced by 'init'
	hidden(commands.NewRequesterCmd()) // flattened into top-level

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func addCmd(cmd *cobra.Command, group string) {
	cmd.GroupID = group
	rootCmd.AddCommand(cmd)
}

func hidden(cmd *cobra.Command) {
	cmd.Hidden = true
	rootCmd.AddCommand(cmd)
}
