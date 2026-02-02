package commands

import (
	"fmt"

	"github.com/moltbunker/moltbunker/internal/client"
	"github.com/moltbunker/moltbunker/pkg/types"
	"github.com/spf13/cobra"
)

var (
	deployTorOnly      bool
	deployOnionService bool
	deployOnionPort    int
	deployCPU          int64
	deployMemory       int64
	deployDisk         int64
)

func NewDeployCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deploy [image]",
		Short: "Deploy a container to the network",
		Long: `Deploy a container image to the moltbunker network.

The container will be:
- Encrypted at rest (uninspectable by node operators)
- Replicated across 3 geographic regions
- Optionally accessible via Tor (.onion address)

Examples:
  moltbunker deploy nginx:latest
  moltbunker deploy myimage:tag --onion-service
  moltbunker deploy myimage:tag --tor-only --memory 1024`,
		Args: cobra.ExactArgs(1),
		RunE: runDeploy,
	}

	cmd.Flags().BoolVar(&deployTorOnly, "tor-only", false, "Deploy with Tor-only networking")
	cmd.Flags().BoolVar(&deployOnionService, "onion-service", false, "Request .onion address for container")
	cmd.Flags().IntVar(&deployOnionPort, "port", 80, "Port to expose via onion service (default 80)")
	cmd.Flags().Int64Var(&deployCPU, "cpu", 100000, "CPU quota in microseconds (default 100000)")
	cmd.Flags().Int64Var(&deployMemory, "memory", 536870912, "Memory limit in bytes (default 512MB)")
	cmd.Flags().Int64Var(&deployDisk, "disk", 10737418240, "Disk limit in bytes (default 10GB)")

	return cmd
}

func runDeploy(cmd *cobra.Command, args []string) error {
	image := args[0]

	daemonClient := client.NewDaemonClient(SocketPath)
	if err := daemonClient.Connect(); err != nil {
		return fmt.Errorf("daemon not running. Start with 'moltbunker start'")
	}
	defer daemonClient.Close()

	fmt.Printf("Deploying container: %s\n", image)
	if deployTorOnly {
		fmt.Println("  Mode: Tor-only networking")
	}
	if deployOnionService {
		fmt.Printf("  Requesting: .onion address (port %d)\n", deployOnionPort)
	}

	// Build deployment request
	req := &client.DeployRequest{
		Image: image,
		Resources: types.ResourceLimits{
			CPUQuota:    deployCPU,
			MemoryLimit: deployMemory,
			DiskLimit:   deployDisk,
		},
		TorOnly:      deployTorOnly,
		OnionService: deployOnionService,
		OnionPort:    deployOnionPort,
	}

	fmt.Println("\nSubmitting deployment request...")

	resp, err := daemonClient.Deploy(req)
	if err != nil {
		return fmt.Errorf("deployment failed: %w", err)
	}

	fmt.Println("\nDeployment Submitted")
	fmt.Println("====================")
	fmt.Printf("Container ID: %s\n", resp.ContainerID)
	fmt.Printf("Status:       %s\n", resp.Status)

	if resp.OnionAddress != "" {
		fmt.Printf("Onion:        %s\n", resp.OnionAddress)
	}

	fmt.Println("\nThe container is being deployed to the network.")
	fmt.Println("Use 'moltbunker logs <container-id>' to view logs.")

	return nil
}
