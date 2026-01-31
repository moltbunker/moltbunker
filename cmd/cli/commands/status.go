package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show daemon status",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Daemon status:")
			// TODO: Implement status display
			return nil
		},
	}
}
