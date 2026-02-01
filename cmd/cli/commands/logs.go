package commands

import (
	"fmt"

	"github.com/moltbunker/moltbunker/internal/client"
	"github.com/spf13/cobra"
)

var (
	logsFollow bool
	logsTail   int
)

func NewLogsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logs [container-id]",
		Short: "View container logs",
		Long: `View logs from a deployed container.

Examples:
  moltbunker logs cnt-123456
  moltbunker logs cnt-123456 --follow
  moltbunker logs cnt-123456 --tail 100`,
		Args: cobra.ExactArgs(1),
		RunE: runLogs,
	}

	cmd.Flags().BoolVarP(&logsFollow, "follow", "f", false, "Follow log output")
	cmd.Flags().IntVarP(&logsTail, "tail", "n", 100, "Number of lines to show from the end")

	return cmd
}

func runLogs(cmd *cobra.Command, args []string) error {
	containerID := args[0]

	daemonClient := client.NewDaemonClient("")
	if err := daemonClient.Connect(); err != nil {
		return fmt.Errorf("daemon not running. Start with 'moltbunker start'")
	}
	defer daemonClient.Close()

	fmt.Printf("Fetching logs for container: %s\n", containerID)
	if logsFollow {
		fmt.Println("Following log output (Ctrl+C to stop)...")
	}
	fmt.Println("---")

	logs, err := daemonClient.GetLogs(containerID, logsFollow, logsTail)
	if err != nil {
		return fmt.Errorf("failed to get logs: %w", err)
	}

	fmt.Println(logs)

	return nil
}
