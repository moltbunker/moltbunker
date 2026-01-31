package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewDeployCmd() *cobra.Command {
	var torOnly bool
	var onionService bool

	cmd := &cobra.Command{
		Use:   "deploy [image]",
		Short: "Deploy container",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			image := args[0]
			fmt.Printf("Deploying container: %s\n", image)
			if torOnly {
				fmt.Println("Tor-only mode enabled")
			}
			if onionService {
				fmt.Println("Onion service requested")
			}
			// TODO: Implement deployment
			return nil
		},
	}

	cmd.Flags().BoolVar(&torOnly, "tor-only", false, "Deploy with Tor-only mode")
	cmd.Flags().BoolVar(&onionService, "onion-service", false, "Request .onion address for container")

	return cmd
}
