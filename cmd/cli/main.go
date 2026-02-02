package main

import (
	"fmt"
	"os"

	"github.com/moltbunker/moltbunker/cmd/cli/commands"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "moltbunker",
	Short: "Moltbunker P2P Encrypted Container Runtime",
	Long:  "A permissionless, fully encrypted P2P network for containerized compute resources",
}

func init() {
	// Add global persistent flags
	rootCmd.PersistentFlags().StringVar(&commands.SocketPath, "socket", "", "Path to daemon socket (default: auto-detect)")
}

func main() {
	// Register commands
	rootCmd.AddCommand(commands.NewInstallCmd())
	rootCmd.AddCommand(commands.NewStartCmd())
	rootCmd.AddCommand(commands.NewStopCmd())
	rootCmd.AddCommand(commands.NewStatusCmd())
	rootCmd.AddCommand(commands.NewDeployCmd())
	rootCmd.AddCommand(commands.NewLogsCmd())
	rootCmd.AddCommand(commands.NewMonitorCmd())
	rootCmd.AddCommand(commands.NewConfigCmd())
	rootCmd.AddCommand(commands.NewInteractiveCmd())
	rootCmd.AddCommand(commands.NewTorCmd())
	rootCmd.AddCommand(commands.NewDoctorCmd())
	rootCmd.AddCommand(commands.NewProviderCmd())
	rootCmd.AddCommand(commands.NewRequesterCmd())
	rootCmd.AddCommand(commands.NewColimaCmd())
	rootCmd.AddCommand(commands.NewVersionCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
