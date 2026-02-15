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

	// TODO: implement output retrieval from daemon
	fmt.Printf("Retrieving outputs for job: %s\n", args[0])
	Warning("Output retrieval not yet implemented")

	return nil
}
