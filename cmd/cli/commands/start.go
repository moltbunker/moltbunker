package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewStartCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Start daemon",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Starting moltbunker daemon...")
			// TODO: Implement daemon start
			return nil
		},
	}
}
