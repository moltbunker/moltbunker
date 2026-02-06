package commands

import (
	"fmt"

	"github.com/moltbunker/moltbunker/internal/client"
	"github.com/spf13/cobra"
)

var (
	snapshotType       string
	snapshotTargetRegion string
	snapshotNewContainer bool
)

func NewSnapshotCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "snapshot",
		Short: "Manage container snapshots",
		Long: `Create, list, restore, and delete container snapshots.

Snapshots capture the state of a container for backup or migration purposes.`,
	}

	cmd.AddCommand(newSnapshotCreateCmd())
	cmd.AddCommand(newSnapshotListCmd())
	cmd.AddCommand(newSnapshotGetCmd())
	cmd.AddCommand(newSnapshotRestoreCmd())
	cmd.AddCommand(newSnapshotDeleteCmd())

	return cmd
}

func newSnapshotCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <container_id>",
		Short: "Create a snapshot of a container",
		Long: `Create a snapshot of the specified container.

Snapshot types:
  - full:        Complete state snapshot
  - incremental: Changes since last snapshot
  - checkpoint:  Quick checkpoint for recovery`,
		Args: cobra.ExactArgs(1),
		RunE: runSnapshotCreate,
	}

	cmd.Flags().StringVar(&snapshotType, "type", "full", "Snapshot type (full, incremental, checkpoint)")

	return cmd
}

func runSnapshotCreate(cmd *cobra.Command, args []string) error {
	containerID := args[0]

	daemonClient := client.NewDaemonClient(SocketPath)

	if err := daemonClient.Connect(); err != nil {
		return fmt.Errorf("daemon not running: %w", err)
	}
	defer daemonClient.Close()

	req := &client.SnapshotRequest{
		ContainerID: containerID,
		Type:        snapshotType,
	}

	fmt.Printf("Creating %s snapshot for container %s...\n", snapshotType, containerID)

	resp, err := daemonClient.SnapshotCreate(req)
	if err != nil {
		return fmt.Errorf("snapshot creation failed: %w", err)
	}

	fmt.Println("\nSnapshot Created")
	fmt.Println("================")
	fmt.Printf("Snapshot ID:  %s\n", resp.ID)
	fmt.Printf("Container ID: %s\n", resp.ContainerID)
	fmt.Printf("Type:         %s\n", resp.Type)
	fmt.Printf("Size:         %s\n", formatBytes(resp.Size))
	fmt.Printf("Checksum:     %s\n", resp.Checksum)
	fmt.Printf("Created:      %s\n", resp.CreatedAt.Format("2006-01-02 15:04:05"))

	return nil
}

func newSnapshotListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list [container_id]",
		Short: "List snapshots",
		Long:  "Display a list of snapshots, optionally filtered by container.",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runSnapshotList,
	}
}

func runSnapshotList(cmd *cobra.Command, args []string) error {
	daemonClient := client.NewDaemonClient(SocketPath)

	if err := daemonClient.Connect(); err != nil {
		return fmt.Errorf("daemon not running: %w", err)
	}
	defer daemonClient.Close()

	var containerID string
	if len(args) > 0 {
		containerID = args[0]
	}

	snapshots, err := daemonClient.SnapshotList(containerID)
	if err != nil {
		return fmt.Errorf("failed to list snapshots: %w", err)
	}

	if len(snapshots) == 0 {
		if containerID != "" {
			fmt.Printf("No snapshots found for container %s\n", containerID)
		} else {
			fmt.Println("No snapshots found")
		}
		return nil
	}

	fmt.Println("Snapshots")
	fmt.Println("=========")
	fmt.Printf("Total: %d\n\n", len(snapshots))

	fmt.Printf("%-20s %-20s %-12s %-10s %s\n", "SNAPSHOT ID", "CONTAINER", "TYPE", "SIZE", "CREATED")
	fmt.Println(repeatStr("-", 85))

	for _, snap := range snapshots {
		fmt.Printf("%-20s %-20s %-12s %-10s %s\n",
			truncateStr(snap.ID, 20),
			truncateStr(snap.ContainerID, 20),
			snap.Type,
			formatBytes(snap.Size),
			snap.CreatedAt.Format("2006-01-02 15:04"),
		)
	}

	return nil
}

func newSnapshotGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <snapshot_id>",
		Short: "Get snapshot details",
		Long:  "Display detailed information about a specific snapshot.",
		Args:  cobra.ExactArgs(1),
		RunE:  runSnapshotGet,
	}
}

func runSnapshotGet(cmd *cobra.Command, args []string) error {
	snapshotID := args[0]

	daemonClient := client.NewDaemonClient(SocketPath)

	if err := daemonClient.Connect(); err != nil {
		return fmt.Errorf("daemon not running: %w", err)
	}
	defer daemonClient.Close()

	snap, err := daemonClient.SnapshotGet(snapshotID)
	if err != nil {
		return fmt.Errorf("failed to get snapshot: %w", err)
	}

	fmt.Println("Snapshot Details")
	fmt.Println("================")
	fmt.Printf("Snapshot ID:  %s\n", snap.ID)
	fmt.Printf("Container ID: %s\n", snap.ContainerID)
	fmt.Printf("Type:         %s\n", snap.Type)
	fmt.Printf("Size:         %s\n", formatBytes(snap.Size))
	fmt.Printf("Checksum:     %s\n", snap.Checksum)
	fmt.Printf("Compressed:   %v\n", snap.Compressed)
	fmt.Printf("Created:      %s\n", snap.CreatedAt.Format("2006-01-02 15:04:05"))

	if snap.ParentID != "" {
		fmt.Printf("Parent:       %s\n", snap.ParentID)
	}

	if len(snap.Metadata) > 0 {
		fmt.Println("\nMetadata:")
		for k, v := range snap.Metadata {
			fmt.Printf("  %s: %s\n", k, v)
		}
	}

	return nil
}

func newSnapshotRestoreCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restore <snapshot_id>",
		Short: "Restore from a snapshot",
		Long: `Restore a container from a snapshot.

By default, restores to the original container. Use --new to create a new container.`,
		Args: cobra.ExactArgs(1),
		RunE: runSnapshotRestore,
	}

	cmd.Flags().StringVar(&snapshotTargetRegion, "region", "", "Target region for restoration")
	cmd.Flags().BoolVar(&snapshotNewContainer, "new", false, "Create a new container instead of replacing")

	return cmd
}

func runSnapshotRestore(cmd *cobra.Command, args []string) error {
	snapshotID := args[0]

	daemonClient := client.NewDaemonClient(SocketPath)

	if err := daemonClient.Connect(); err != nil {
		return fmt.Errorf("daemon not running: %w", err)
	}
	defer daemonClient.Close()

	fmt.Printf("Restoring from snapshot %s...\n", snapshotID)

	result, err := daemonClient.SnapshotRestore(snapshotID, snapshotTargetRegion, snapshotNewContainer)
	if err != nil {
		return fmt.Errorf("restore failed: %w", err)
	}

	fmt.Println("\nRestore Initiated")
	fmt.Println("=================")

	if containerID, ok := result["container_id"].(string); ok {
		fmt.Printf("Container ID: %s\n", containerID)
	}
	if status, ok := result["status"].(string); ok {
		fmt.Printf("Status:       %s\n", status)
	}
	if region, ok := result["region"].(string); ok {
		fmt.Printf("Region:       %s\n", region)
	}

	return nil
}

func newSnapshotDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <snapshot_id>",
		Short: "Delete a snapshot",
		Long:  "Permanently delete a snapshot.",
		Args:  cobra.ExactArgs(1),
		RunE:  runSnapshotDelete,
	}
}

func runSnapshotDelete(cmd *cobra.Command, args []string) error {
	snapshotID := args[0]

	daemonClient := client.NewDaemonClient(SocketPath)

	if err := daemonClient.Connect(); err != nil {
		return fmt.Errorf("daemon not running: %w", err)
	}
	defer daemonClient.Close()

	if err := daemonClient.SnapshotDelete(snapshotID); err != nil {
		return fmt.Errorf("delete failed: %w", err)
	}

	fmt.Printf("Snapshot '%s' deleted\n", snapshotID)

	return nil
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
