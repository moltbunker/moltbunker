package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewPeersCmd creates the peer list command.
func NewPeersCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "peers",
		Short: "Connected peers",
		Long:  "List connected P2P peers. Requires daemon.",
		RunE:  runPeers,
	}
}

func runPeers(cmd *cobra.Command, args []string) error {
	c, err := GetClient()
	if err != nil {
		return err
	}
	defer c.Close()

	peers, err := c.GetPeers()
	if err != nil {
		return fmt.Errorf("failed to get peers: %w", err)
	}

	if len(peers) == 0 {
		Info("No connected peers.")
		return nil
	}

	headers := []string{"Peer ID", "Region", "Address", "Last Seen"}
	var rows [][]string
	for _, p := range peers {
		region := p.Region
		if region == "" {
			region = "unknown"
		}
		rows = append(rows, []string{
			FormatNodeID(p.ID),
			region,
			p.Address,
			p.LastSeen.Format("15:04:05"),
		})
	}

	fmt.Println(RenderTable(headers, rows))
	fmt.Println(StyleMuted.Render(fmt.Sprintf("  %d peer(s)", len(peers))))

	return nil
}
