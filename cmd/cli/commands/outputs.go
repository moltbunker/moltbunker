package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewOutputsCmd creates the outputs retrieval command.
func NewOutputsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "outputs [job-id]",
		Short: "Retrieve job outputs",
		Long:  "Retrieve and decrypt outputs from a completed job.",
		Args:  cobra.ExactArgs(1),
		RunE:  runOutputs,
	}
}

func runOutputs(cmd *cobra.Command, args []string) error {
	c, err := GetClient()
	if err != nil {
		return err
	}
	defer c.Close()

	if err := c.RequireDaemon(); err != nil {
		return err
	}

	dc := c.DaemonClient()
	containerID := args[0]

	// Get container detail which includes exposed ports and endpoints
	detail, err := dc.GetContainerDetail(containerID)
	if err != nil {
		return fmt.Errorf("failed to get container detail: %w", err)
	}

	fields := [][2]string{
		{"Container ID", containerID},
		{"Status", detail.Status},
		{"Image", detail.Image},
		{"Provider Node", FormatNodeID(detail.ProviderNodeID)},
	}
	if detail.ProviderAddress != "" {
		fields = append(fields, [2]string{"Provider Wallet", detail.ProviderAddress})
	}
	if detail.Owner != "" {
		fields = append(fields, [2]string{"Owner", detail.Owner})
	}

	fmt.Println(StatusBox("Container Outputs", fields))

	return nil
}
