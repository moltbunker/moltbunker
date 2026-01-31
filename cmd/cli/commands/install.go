package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewInstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "install",
		Short: "Install daemon",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Installing moltbunker daemon...")
			// TODO: Implement installation
			return nil
		},
	}
}
