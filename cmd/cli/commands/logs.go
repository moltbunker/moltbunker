package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewLogsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logs [container]",
		Short: "View container logs",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			containerID := args[0]
			fmt.Printf("Viewing logs for container: %s\n", containerID)
			// TODO: Implement log viewing
			return nil
		},
	}
}
