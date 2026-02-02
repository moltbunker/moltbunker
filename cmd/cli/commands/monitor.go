package commands

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/moltbunker/moltbunker/internal/client"
	"github.com/spf13/cobra"
)

var (
	monitorRefresh int
)

func NewMonitorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "monitor",
		Short: "Real-time monitoring dashboard",
		Long: `Display a real-time monitoring dashboard.

Shows live updates on:
- Node status and peer count
- Container health
- Network activity
- Resource usage

Press Ctrl+C to exit.`,
		RunE: runMonitor,
	}

	cmd.Flags().IntVarP(&monitorRefresh, "refresh", "r", 2, "Refresh interval in seconds")

	return cmd
}

func runMonitor(cmd *cobra.Command, args []string) error {
	daemonClient := client.NewDaemonClient(SocketPath)
	if err := daemonClient.Connect(); err != nil {
		return fmt.Errorf("daemon not running. Start with 'moltbunker start'")
	}
	defer daemonClient.Close()

	// Handle Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.NewTicker(time.Duration(monitorRefresh) * time.Second)
	defer ticker.Stop()

	// Clear screen and show initial state
	clearScreen()
	if err := displayDashboard(daemonClient); err != nil {
		return err
	}

	for {
		select {
		case <-sigChan:
			fmt.Println("\nExiting monitor...")
			return nil
		case <-ticker.C:
			clearScreen()
			if err := displayDashboard(daemonClient); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		}
	}
}

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}

func displayDashboard(daemonClient *client.DaemonClient) error {
	status, err := daemonClient.Status()
	if err != nil {
		return fmt.Errorf("failed to get status: %w", err)
	}

	// Header
	fmt.Println("╔════════════════════════════════════════════════════════════════╗")
	fmt.Println("║              MOLTBUNKER MONITORING DASHBOARD                   ║")
	fmt.Println("╠════════════════════════════════════════════════════════════════╣")

	// Node Status
	statusIcon := "●"
	if status.Running {
		statusIcon = "\033[32m●\033[0m" // Green
	} else {
		statusIcon = "\033[31m●\033[0m" // Red
	}
	fmt.Printf("║ Node Status: %s %-48s ║\n", statusIcon, formatStatus(status.Running))
	fmt.Printf("║ Node ID:     %-51s ║\n", truncateID(status.NodeID, 48))
	fmt.Printf("║ Version:     %-51s ║\n", status.Version)
	fmt.Printf("║ Port:        %-51d ║\n", status.Port)

	fmt.Println("╠════════════════════════════════════════════════════════════════╣")

	// Network Status
	fmt.Println("║                       NETWORK STATUS                           ║")
	fmt.Println("╠════════════════════════════════════════════════════════════════╣")
	fmt.Printf("║ Connected Peers: %-47d ║\n", status.PeerCount)

	// Tor Status
	torStatus := "Disabled"
	if status.TorEnabled {
		torStatus = "Enabled"
	}
	fmt.Printf("║ Tor Service:     %-47s ║\n", torStatus)
	if status.TorAddress != "" {
		fmt.Printf("║ Onion Address:   %-47s ║\n", truncateID(status.TorAddress, 44))
	}

	fmt.Println("╠════════════════════════════════════════════════════════════════╣")

	// Peer List
	peers, err := daemonClient.GetPeers()
	if err == nil && len(peers) > 0 {
		fmt.Println("║                        CONNECTED PEERS                         ║")
		fmt.Println("╠════════════════════════════════════════════════════════════════╣")

		maxPeers := 5
		if len(peers) < maxPeers {
			maxPeers = len(peers)
		}

		for i := 0; i < maxPeers; i++ {
			peer := peers[i]
			region := peer.Region
			if region == "" {
				region = "???"
			}
			peerLine := fmt.Sprintf("%s... [%s]", truncateID(peer.ID, 12), region)
			fmt.Printf("║   %-60s ║\n", peerLine)
		}

		if len(peers) > maxPeers {
			fmt.Printf("║   ... and %d more peers %-38s ║\n", len(peers)-maxPeers, "")
		}
	} else {
		fmt.Println("║                        CONNECTED PEERS                         ║")
		fmt.Println("╠════════════════════════════════════════════════════════════════╣")
		fmt.Println("║   No peers connected                                           ║")
	}

	fmt.Println("╠════════════════════════════════════════════════════════════════╣")

	// Footer
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	fmt.Printf("║ Last Updated: %-50s ║\n", timestamp)
	fmt.Println("║ Press Ctrl+C to exit                                           ║")
	fmt.Println("╚════════════════════════════════════════════════════════════════╝")

	return nil
}

func formatStatus(running bool) string {
	if running {
		return "RUNNING"
	}
	return "STOPPED"
}

func truncateID(id string, maxLen int) string {
	if len(id) <= maxLen {
		return id
	}
	return id[:maxLen-3] + "..."
}
