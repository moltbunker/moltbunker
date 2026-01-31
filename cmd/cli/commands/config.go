package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewConfigCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "config",
		Short: "Configuration management",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Configuration management")
			// TODO: Implement config management
			return nil
		},
	}
}
