package commands

import (
	"fmt"

	"github.com/moltbunker/moltbunker/internal/client"
	"github.com/spf13/cobra"
)

var (
	cloneTargetRegion string
	cloneReason       string
	cloneIncludeState bool
	clonePriority     int
	cloneActiveOnly   bool
	cloneLimit        int
)

func NewCloneCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clone",
		Short: "Manage container cloning",
		Long: `Trigger and manage container cloning operations.

The cloning system creates copies of containers on different nodes
for redundancy and threat mitigation.`,
	}

	cmd.AddCommand(newCloneCreateCmd())
	cmd.AddCommand(newCloneStatusCmd())
	cmd.AddCommand(newCloneListCmd())
	cmd.AddCommand(newCloneCancelCmd())

	return cmd
}

func newCloneCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create [container_id]",
		Short: "Create a clone of a container",
		Long: `Trigger a manual clone operation for a container.

The clone will be deployed to a different node for redundancy.
Use --region to specify a target region, or leave blank for automatic selection.`,
		Args: cobra.MaximumNArgs(1),
		RunE: runCloneCreate,
	}

	cmd.Flags().StringVar(&cloneTargetRegion, "region", "", "Target region for clone (americas, europe, asia_pacific)")
	cmd.Flags().StringVar(&cloneReason, "reason", "manual_clone", "Reason for cloning")
	cmd.Flags().BoolVar(&cloneIncludeState, "include-state", true, "Include container state in clone")
	cmd.Flags().IntVar(&clonePriority, "priority", 2, "Clone priority (1=low, 2=normal, 3=high)")

	return cmd
}

func runCloneCreate(cmd *cobra.Command, args []string) error {
	daemonClient := client.NewDaemonClient(SocketPath)

	if err := daemonClient.Connect(); err != nil {
		return fmt.Errorf("daemon not running: %w", err)
	}
	defer daemonClient.Close()

	var sourceID string
	if len(args) > 0 {
		sourceID = args[0]
	}

	req := &client.CloneRequest{
		SourceID:     sourceID,
		TargetRegion: cloneTargetRegion,
		Priority:     clonePriority,
		Reason:       cloneReason,
		IncludeState: cloneIncludeState,
	}

	fmt.Println("Initiating clone operation...")

	resp, err := daemonClient.Clone(req)
	if err != nil {
		return fmt.Errorf("clone failed: %w", err)
	}

	fmt.Println("\nClone Operation Initiated")
	fmt.Println("=========================")
	fmt.Printf("Clone ID:      %s\n", resp.CloneID)
	fmt.Printf("Source:        %s\n", resp.SourceID)
	fmt.Printf("Target Region: %s\n", resp.TargetRegion)
	fmt.Printf("Status:        %s\n", resp.Status)
	fmt.Printf("Created:       %s\n", resp.CreatedAt.Format("2006-01-02 15:04:05"))

	fmt.Println("\nUse 'moltbunker clone status", resp.CloneID+"' to check progress")

	return nil
}

func newCloneStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status <clone_id>",
		Short: "Check status of a clone operation",
		Long:  "Display the current status of a clone operation.",
		Args:  cobra.ExactArgs(1),
		RunE:  runCloneStatus,
	}
}

func runCloneStatus(cmd *cobra.Command, args []string) error {
	cloneID := args[0]

	daemonClient := client.NewDaemonClient(SocketPath)

	if err := daemonClient.Connect(); err != nil {
		return fmt.Errorf("daemon not running: %w", err)
	}
	defer daemonClient.Close()

	resp, err := daemonClient.CloneStatus(cloneID)
	if err != nil {
		return fmt.Errorf("failed to get clone status: %w", err)
	}

	fmt.Println("Clone Status")
	fmt.Println("============")
	fmt.Printf("Clone ID:      %s\n", resp.CloneID)
	fmt.Printf("Source ID:     %s\n", resp.SourceID)
	fmt.Printf("Target ID:     %s\n", resp.TargetID)
	fmt.Printf("Target Node:   %s\n", resp.TargetNodeID)
	fmt.Printf("Target Region: %s\n", resp.TargetRegion)
	fmt.Printf("Status:        %s\n", formatCloneStatus(resp.Status))
	fmt.Printf("Priority:      %d\n", resp.Priority)
	fmt.Printf("Reason:        %s\n", resp.Reason)
	fmt.Printf("Created:       %s\n", resp.CreatedAt.Format("2006-01-02 15:04:05"))

	if !resp.CompletedAt.IsZero() {
		fmt.Printf("Completed:     %s\n", resp.CompletedAt.Format("2006-01-02 15:04:05"))
	}

	if resp.Error != "" {
		fmt.Printf("Error:         %s\n", resp.Error)
	}

	return nil
}

func newCloneListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List clone operations",
		Long:  "Display a list of recent clone operations.",
		RunE:  runCloneList,
	}

	cmd.Flags().BoolVar(&cloneActiveOnly, "active", false, "Show only active clones")
	cmd.Flags().IntVar(&cloneLimit, "limit", 10, "Maximum number of clones to show")

	return cmd
}

func runCloneList(cmd *cobra.Command, args []string) error {
	daemonClient := client.NewDaemonClient(SocketPath)

	if err := daemonClient.Connect(); err != nil {
		return fmt.Errorf("daemon not running: %w", err)
	}
	defer daemonClient.Close()

	clones, err := daemonClient.CloneList(cloneActiveOnly, cloneLimit)
	if err != nil {
		return fmt.Errorf("failed to list clones: %w", err)
	}

	if len(clones) == 0 {
		if cloneActiveOnly {
			fmt.Println("No active clone operations")
		} else {
			fmt.Println("No clone operations found")
		}
		return nil
	}

	fmt.Println("Clone Operations")
	fmt.Println("================")
	fmt.Printf("Total: %d\n\n", len(clones))

	fmt.Printf("%-20s %-12s %-15s %-12s %s\n", "CLONE ID", "STATUS", "REGION", "PRIORITY", "CREATED")
	fmt.Println(repeatStr("-", 80))

	for _, clone := range clones {
		fmt.Printf("%-20s %-12s %-15s %-12d %s\n",
			truncateStr(clone.CloneID, 20),
			clone.Status,
			clone.TargetRegion,
			clone.Priority,
			clone.CreatedAt.Format("2006-01-02 15:04"),
		)
	}

	return nil
}

func newCloneCancelCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "cancel <clone_id>",
		Short: "Cancel a clone operation",
		Long:  "Cancel an in-progress clone operation.",
		Args:  cobra.ExactArgs(1),
		RunE:  runCloneCancel,
	}
}

func runCloneCancel(cmd *cobra.Command, args []string) error {
	cloneID := args[0]

	daemonClient := client.NewDaemonClient(SocketPath)

	if err := daemonClient.Connect(); err != nil {
		return fmt.Errorf("daemon not running: %w", err)
	}
	defer daemonClient.Close()

	if err := daemonClient.CloneCancel(cloneID); err != nil {
		return fmt.Errorf("failed to cancel clone: %w", err)
	}

	fmt.Printf("Clone operation '%s' cancelled\n", cloneID)

	return nil
}

func formatCloneStatus(status string) string {
	switch status {
	case "pending":
		return "PENDING"
	case "preparing":
		return "PREPARING"
	case "transferring":
		return "TRANSFERRING"
	case "deploying":
		return "DEPLOYING"
	case "verifying":
		return "VERIFYING"
	case "complete":
		return "COMPLETE"
	case "failed":
		return "FAILED"
	case "cancelled":
		return "CANCELLED"
	default:
		return status
	}
}

func truncateStr(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func repeatStr(s string, n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}
