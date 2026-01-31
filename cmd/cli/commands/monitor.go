package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewMonitorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "monitor",
		Short: "Real-time monitoring dashboard",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Starting monitoring dashboard...")
			// TODO: Implement monitoring dashboard
			return nil
		},
	}
}
