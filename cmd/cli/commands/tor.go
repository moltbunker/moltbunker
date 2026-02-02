package commands

import (
	"fmt"

	"github.com/moltbunker/moltbunker/internal/client"
	"github.com/spf13/cobra"
)

func NewTorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tor",
		Short: "Tor management commands",
		Long: `Manage Tor integration for anonymous networking.

Subcommands:
  start   - Start the Tor service
  status  - Show Tor service status
  onion   - Display .onion address
  rotate  - Rotate Tor circuit for new identity`,
	}

	cmd.AddCommand(NewTorStartCmd())
	cmd.AddCommand(NewTorStatusCmd())
	cmd.AddCommand(NewTorOnionCmd())
	cmd.AddCommand(NewTorRotateCmd())

	return cmd
}

func NewTorStartCmd() *cobra.Command {
	var exitCountry string

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start Tor service",
		Long: `Start the Tor service for anonymous networking.

This will:
- Initialize Tor daemon
- Create onion service for inbound connections
- Configure SOCKS5 proxy for outbound traffic`,
		RunE: func(cmd *cobra.Command, args []string) error {
			daemonClient := client.NewDaemonClient(SocketPath)
			if err := daemonClient.Connect(); err != nil {
				return fmt.Errorf("daemon not running. Start with 'moltbunker start'")
			}
			defer daemonClient.Close()

			fmt.Println("Starting Tor service...")

			result, err := daemonClient.TorStart()
			if err != nil {
				return fmt.Errorf("failed to start Tor: %w", err)
			}

			status, ok := result["status"].(string)
			if ok {
				switch status {
				case "started":
					fmt.Println("Tor service started successfully")
				case "already_running":
					fmt.Println("Tor service is already running")
				}
			}

			if addr, ok := result["address"].(string); ok && addr != "" {
				fmt.Printf("Onion address: %s\n", addr)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&exitCountry, "exit-country", "", "Preferred exit node country code (e.g., US, DE)")

	return cmd
}

func NewTorStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show Tor status",
		Long:  "Display the current status of the Tor service.",
		RunE: func(cmd *cobra.Command, args []string) error {
			daemonClient := client.NewDaemonClient(SocketPath)
			if err := daemonClient.Connect(); err != nil {
				return fmt.Errorf("daemon not running. Start with 'moltbunker start'")
			}
			defer daemonClient.Close()

			status, err := daemonClient.TorStatus()
			if err != nil {
				return fmt.Errorf("failed to get Tor status: %w", err)
			}

			fmt.Println("Tor Service Status")
			fmt.Println("==================")

			if status.Running {
				fmt.Println("Status:   RUNNING")
				if status.OnionAddress != "" {
					fmt.Printf("Onion:    %s\n", status.OnionAddress)
				}
				if !status.StartedAt.IsZero() {
					fmt.Printf("Started:  %s\n", status.StartedAt.Format("2006-01-02 15:04:05"))
				}
				fmt.Printf("Circuits: %d\n", status.CircuitCount)
			} else {
				fmt.Println("Status:   NOT RUNNING")
				fmt.Println("\nStart Tor with: moltbunker tor start")
			}

			return nil
		},
	}
}

func NewTorOnionCmd() *cobra.Command {
	var showQR bool

	cmd := &cobra.Command{
		Use:   "onion",
		Short: "Show .onion address",
		Long:  "Display the node's .onion address for inbound Tor connections.",
		RunE: func(cmd *cobra.Command, args []string) error {
			daemonClient := client.NewDaemonClient(SocketPath)
			if err := daemonClient.Connect(); err != nil {
				return fmt.Errorf("daemon not running. Start with 'moltbunker start'")
			}
			defer daemonClient.Close()

			status, err := daemonClient.TorStatus()
			if err != nil {
				return fmt.Errorf("failed to get Tor status: %w", err)
			}

			if !status.Running {
				return fmt.Errorf("Tor is not running. Start with: moltbunker tor start")
			}

			if status.OnionAddress == "" {
				return fmt.Errorf("no onion address available")
			}

			fmt.Println("Onion Address")
			fmt.Println("=============")
			fmt.Println(status.OnionAddress)

			if showQR {
				fmt.Println("\nQR Code:")
				printSimpleQR(status.OnionAddress)
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&showQR, "qr", false, "Display QR code for onion address")

	return cmd
}

func NewTorRotateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rotate",
		Short: "Rotate Tor circuit",
		Long: `Request a new Tor circuit for a fresh identity.

This creates new paths through the Tor network, giving you:
- New exit IP address
- Fresh circuit for improved anonymity`,
		RunE: func(cmd *cobra.Command, args []string) error {
			daemonClient := client.NewDaemonClient(SocketPath)
			if err := daemonClient.Connect(); err != nil {
				return fmt.Errorf("daemon not running. Start with 'moltbunker start'")
			}
			defer daemonClient.Close()

			// Check if Tor is running
			status, err := daemonClient.TorStatus()
			if err != nil {
				return fmt.Errorf("failed to get Tor status: %w", err)
			}

			if !status.Running {
				return fmt.Errorf("Tor is not running. Start with: moltbunker tor start")
			}

			fmt.Println("Rotating Tor circuit...")

			if err := daemonClient.TorRotate(); err != nil {
				return fmt.Errorf("failed to rotate circuit: %w", err)
			}

			fmt.Println("Circuit rotated successfully")
			fmt.Println("You now have a new identity on the Tor network")

			return nil
		},
	}
}

// printSimpleQR prints a simple ASCII representation for the onion address
func printSimpleQR(address string) {
	// Simple box representation - actual QR would require a library
	fmt.Println("┌────────────────────────────────┐")
	fmt.Println("│ ▄▄▄▄▄ █▀▄▄▀█▄█▀█▄▀█ ▄▄▄▄▄ │")
	fmt.Println("│ █   █ █▀▄ ▀▀▄▀▄▄▀██ █   █ │")
	fmt.Println("│ █▄▄▄█ █▀▄▀█▀█ ▄▀▄▄█ █▄▄▄█ │")
	fmt.Println("│▄▄▄▄▄▄▄█▄█▄█ █▄█ █▄█▄▄▄▄▄▄▄│")
	fmt.Println("│ ▄▀▄▀█▄  ▄▀▀▄▄▀█▀▄█▄█▄▀▀▀▀ │")
	fmt.Println("│▄█▄█▄█▄▄▄ █▄▄▀▀ █▀▄ ▄▄▄ ▀█ │")
	fmt.Println("│ ▄▄▄▄▄ █▄▄█▀▄ ▄▀▄▀ █▄█ ▄▀▄│")
	fmt.Println("│ █   █ █ ▀█▄▄▀▄▄█▀▄▄▄  █▀▄│")
	fmt.Println("│ █▄▄▄█ █ ▄▀▀▄█▀▄▀▀█▄█▀██▄▀│")
	fmt.Println("│▄▄▄▄▄▄▄█▄█▄██▄█▄█▄████▄███│")
	fmt.Println("└────────────────────────────────┘")
	fmt.Println("(Note: Use a proper QR library for production)")
}
