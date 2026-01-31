package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewInteractiveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "interactive",
		Short: "Interactive TUI mode",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Starting interactive TUI...")
			// TODO: Implement interactive TUI
			return nil
		},
	}
}
