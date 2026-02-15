package commands

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

func NewStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Node & network status",
		Long: `Display the current status of the moltbunker node.

Shows node ID, network info, Tor status, and connected peers.
Works via daemon socket or HTTP API.`,
		RunE: runStatus,
	}
}

func runStatus(cmd *cobra.Command, args []string) error {
	c, err := GetClient()
	if err != nil {
		fmt.Println(StatusBox("Daemon Status", [][2]string{
			{"Status", "NOT RUNNING"},
		}))
		fmt.Println(Hint("Start with: moltbunker start"))
		return nil
	}
	defer c.Close()

	status, err := c.Status()
	if err != nil {
		return fmt.Errorf("failed to get status: %w", err)
	}

	// Node status box
	statusStr := "STOPPED"
	if status.Running {
		statusStr = "RUNNING"
	}
	fmt.Println(StatusBox("Node", [][2]string{
		{"Status", StatusBadge(statusStr)},
		{"Node ID", FormatNodeID(status.NodeID)},
		{"Version", status.Version},
		{"Port", strconv.Itoa(status.Port)},
		{"Network", fmt.Sprintf("%d nodes", status.NetworkNodes)},
		{"Region", status.Region},
	}))

	// Tor box
	torStatus := "DISABLED"
	if status.TorEnabled {
		torStatus = "ENABLED"
	}
	torFields := [][2]string{
		{"Status", StatusBadge(torStatus)},
	}
	if status.TorAddress != "" {
		torFields = append(torFields, [2]string{"Onion", status.TorAddress})
	}
	if !status.TorEnabled {
		torFields = append(torFields, [2]string{"", Hint("Enable with: moltbunker tor start")})
	}
	fmt.Println(StatusBox("Tor", torFields))

	// Peer list
	peers, err := c.GetPeers()
	if err == nil && len(peers) > 0 {
		headers := []string{"Peer ID", "Region", "Address"}
		var rows [][]string
		for _, peer := range peers {
			region := peer.Region
			if region == "" {
				region = "unknown"
			}
			rows = append(rows, []string{
				FormatNodeID(peer.ID),
				region,
				peer.Address,
			})
		}
		fmt.Println(SectionHeader("Connected Peers"))
		fmt.Println(RenderTable(headers, rows))
	}

	return nil
}
