package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/moltbunker/moltbunker/internal/client"
	"github.com/spf13/cobra"
)

func NewInteractiveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "interactive",
		Short: "Interactive TUI mode",
		Long: `Start an interactive command-line interface for moltbunker.

This provides a REPL-style interface for interacting with the daemon.
Type 'help' for available commands, 'exit' to quit.`,
		RunE: runInteractive,
	}
}

func runInteractive(cmd *cobra.Command, args []string) error {
	daemonClient := client.NewDaemonClient(SocketPath)
	connected := false

	if err := daemonClient.Connect(); err == nil {
		connected = true
		defer daemonClient.Close()
	}

	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║           MOLTBUNKER INTERACTIVE MODE                    ║")
	fmt.Println("║                                                          ║")
	fmt.Println("║  Type 'help' for commands, 'exit' to quit                ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Println()

	if !connected {
		fmt.Println("Warning: Daemon not running. Some commands unavailable.")
		fmt.Println("Start with 'moltbunker start' in another terminal.")
		fmt.Println()
	}

	scanner := bufio.NewScanner(os.Stdin)
	for {
		if connected {
			fmt.Print("\033[32mmoltbunker>\033[0m ")
		} else {
			fmt.Print("\033[33mmoltbunker (offline)>\033[0m ")
		}

		if !scanner.Scan() {
			break
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		command := parts[0]
		cmdArgs := parts[1:]

		switch command {
		case "exit", "quit", "q":
			fmt.Println("Goodbye!")
			return nil

		case "help", "h", "?":
			printInteractiveHelp()

		case "status":
			if !connected {
				fmt.Println("Error: Daemon not connected")
				continue
			}
			status, err := daemonClient.Status()
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				continue
			}
			fmt.Printf("Node ID:  %s\n", status.NodeID)
			fmt.Printf("Running:  %v\n", status.Running)
			fmt.Printf("Port:     %d\n", status.Port)
			fmt.Printf("Peers:    %d\n", status.PeerCount)
			fmt.Printf("Tor:      %v\n", status.TorEnabled)

		case "peers":
			if !connected {
				fmt.Println("Error: Daemon not connected")
				continue
			}
			peers, err := daemonClient.GetPeers()
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				continue
			}
			if len(peers) == 0 {
				fmt.Println("No peers connected")
				continue
			}
			fmt.Printf("Connected peers: %d\n", len(peers))
			for _, peer := range peers {
				region := peer.Region
				if region == "" {
					region = "unknown"
				}
				fmt.Printf("  %s [%s] %s\n", peer.ID[:16]+"...", region, peer.Address)
			}

		case "deploy":
			if !connected {
				fmt.Println("Error: Daemon not connected")
				continue
			}
			if len(cmdArgs) < 1 {
				fmt.Println("Usage: deploy <image> [--onion]")
				continue
			}
			image := cmdArgs[0]
			onion := false
			for _, arg := range cmdArgs[1:] {
				if arg == "--onion" || arg == "-o" {
					onion = true
				}
			}

			fmt.Printf("Deploying %s...\n", image)
			resp, err := daemonClient.Deploy(&client.DeployRequest{
				Image:        image,
				OnionService: onion,
			})
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				continue
			}
			fmt.Printf("Container ID: %s\n", resp.ContainerID)
			fmt.Printf("Status: %s\n", resp.Status)
			if resp.OnionAddress != "" {
				fmt.Printf("Onion: %s\n", resp.OnionAddress)
			}

		case "logs":
			if !connected {
				fmt.Println("Error: Daemon not connected")
				continue
			}
			if len(cmdArgs) < 1 {
				fmt.Println("Usage: logs <container-id>")
				continue
			}
			containerID := cmdArgs[0]
			logs, err := daemonClient.GetLogs(containerID, false, 50)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				continue
			}
			fmt.Println(logs)

		case "tor":
			if !connected {
				fmt.Println("Error: Daemon not connected")
				continue
			}
			if len(cmdArgs) < 1 {
				fmt.Println("Usage: tor <start|status|rotate>")
				continue
			}
			switch cmdArgs[0] {
			case "start":
				result, err := daemonClient.TorStart()
				if err != nil {
					fmt.Printf("Error: %v\n", err)
					continue
				}
				fmt.Printf("Tor: %v\n", result)
			case "status":
				status, err := daemonClient.TorStatus()
				if err != nil {
					fmt.Printf("Error: %v\n", err)
					continue
				}
				fmt.Printf("Running: %v\n", status.Running)
				if status.OnionAddress != "" {
					fmt.Printf("Onion:   %s\n", status.OnionAddress)
				}
			case "rotate":
				if err := daemonClient.TorRotate(); err != nil {
					fmt.Printf("Error: %v\n", err)
					continue
				}
				fmt.Println("Circuit rotated")
			default:
				fmt.Println("Unknown tor subcommand")
			}

		case "connect":
			if connected {
				fmt.Println("Already connected")
				continue
			}
			if err := daemonClient.Connect(); err != nil {
				fmt.Printf("Failed to connect: %v\n", err)
				continue
			}
			connected = true
			fmt.Println("Connected to daemon")

		case "disconnect":
			if !connected {
				fmt.Println("Not connected")
				continue
			}
			daemonClient.Close()
			connected = false
			fmt.Println("Disconnected")

		case "clear", "cls":
			fmt.Print("\033[H\033[2J")

		default:
			fmt.Printf("Unknown command: %s\n", command)
			fmt.Println("Type 'help' for available commands")
		}

		fmt.Println()
	}

	return nil
}

func printInteractiveHelp() {
	fmt.Println(`
Available commands:
  status            Show daemon status
  peers             List connected peers
  deploy <image>    Deploy a container (use --onion for .onion address)
  logs <id>         View container logs
  tor start         Start Tor service
  tor status        Show Tor status
  tor rotate        Rotate Tor circuit
  connect           Connect to daemon
  disconnect        Disconnect from daemon
  clear             Clear screen
  help              Show this help
  exit              Exit interactive mode`)
}
