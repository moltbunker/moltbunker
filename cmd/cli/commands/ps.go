package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewPsCmd creates the container list command.
func NewPsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ps",
		Short: "List your containers",
		Long:  "List all containers deployed to the network. Works via daemon or HTTP API.",
		RunE:  runPs,
	}

	return cmd
}

func runPs(cmd *cobra.Command, args []string) error {
	c, err := GetClient()
	if err != nil {
		return err
	}
	defer c.Close()

	containers, err := c.List()
	if err != nil {
		return fmt.Errorf("failed to list containers: %w", err)
	}

	if len(containers) == 0 {
		Info("No containers running.")
		fmt.Println(Hint("Deploy one with: moltbunker deploy <image>"))
		return nil
	}

	headers := []string{"ID", "Image", "Status", "Regions", "Created"}
	var rows [][]string
	for _, ct := range containers {
		regions := "-"
		if len(ct.Regions) > 0 {
			regions = ct.Regions[0]
			if len(ct.Regions) > 1 {
				regions += fmt.Sprintf(" +%d", len(ct.Regions)-1)
			}
		}
		rows = append(rows, []string{
			FormatNodeID(ct.ID),
			ct.Image,
			ct.Status,
			regions,
			ct.CreatedAt.Format("Jan 02 15:04"),
		})
	}

	fmt.Println(RenderTable(headers, rows))
	fmt.Println(StyleMuted.Render(fmt.Sprintf("  %d container(s)", len(containers))))

	return nil
}
