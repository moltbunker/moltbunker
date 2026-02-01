package commands

import (
	"fmt"

	"github.com/moltbunker/moltbunker/internal/client"
	"github.com/spf13/cobra"
)

func NewStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show daemon status",
		Long: `Display the current status of the moltbunker daemon.

Shows information about:
- Node ID and running state
- P2P port and peer count
- Tor service status
- Version information`,
		RunE: runStatus,
	}
}

func runStatus(cmd *cobra.Command, args []string) error {
	daemonClient := client.NewDaemonClient("")

	if err := daemonClient.Connect(); err != nil {
		fmt.Println("Daemon status: NOT RUNNING")
		fmt.Println("\nThe daemon is not running. Start it with:")
		fmt.Println("  moltbunker start")
		return nil
	}
	defer daemonClient.Close()

	status, err := daemonClient.Status()
	if err != nil {
		return fmt.Errorf("failed to get status: %w", err)
	}

	fmt.Println("Daemon Status")
	fmt.Println("=============")
	fmt.Printf("Status:      %s\n", formatRunningStatus(status.Running))
	fmt.Printf("Node ID:     %s\n", status.NodeID)
	fmt.Printf("Version:     %s\n", status.Version)
	fmt.Printf("Port:        %d\n", status.Port)
	fmt.Printf("Peers:       %d\n", status.PeerCount)

	fmt.Println("\nTor Status")
	fmt.Println("----------")
	if status.TorEnabled {
		fmt.Printf("Tor:         ENABLED\n")
		if status.TorAddress != "" {
			fmt.Printf("Onion:       %s\n", status.TorAddress)
		}
	} else {
		fmt.Printf("Tor:         DISABLED\n")
		fmt.Println("             Run 'moltbunker tor start' to enable")
	}

	// Get peer list
	peers, err := daemonClient.GetPeers()
	if err == nil && len(peers) > 0 {
		fmt.Println("\nConnected Peers")
		fmt.Println("---------------")
		for _, peer := range peers {
			region := peer.Region
			if region == "" {
				region = "unknown"
			}
			fmt.Printf("  %s (%s) - %s\n", peer.ID[:16]+"...", region, peer.Address)
		}
	}

	return nil
}

func formatRunningStatus(running bool) string {
	if running {
		return "RUNNING"
	}
	return "STOPPED"
}
