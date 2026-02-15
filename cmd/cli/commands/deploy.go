package commands

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/moltbunker/moltbunker/internal/client"
	"github.com/spf13/cobra"
)

var (
	deployTorOnly      bool
	deployOnionService bool
	deployOnionPort    int
	deployCPU          int64
	deployMemory       int64
	deployDisk         int64
	deployMinTier      string
)

// Resource tier presets
type resourceTier struct {
	name   string
	cpu    int64
	memMB  int64
	diskMB int64
	bunker string // estimated cost per hour
}

var resourceTiers = []resourceTier{
	{"Minimal (1 vCPU, 512 MB, 5 GB)", 50000, 512, 5120, "~380 BUNKER/hr"},
	{"Standard (2 vCPU, 1 GB, 10 GB)", 100000, 1024, 10240, "~760 BUNKER/hr"},
	{"Performance (4 vCPU, 4 GB, 20 GB)", 200000, 4096, 20480, "~2,720 BUNKER/hr"},
	{"Custom", 0, 0, 0, "set your own"},
}

func NewDeployCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deploy [image]",
		Short: "Deploy a container to the network",
		Long: `Deploy a container image to the moltbunker network.

Without arguments, launches an interactive wizard with full control.
With an image argument, deploys non-interactively using flags.

Use Shift+Tab to go back to previous steps in the wizard.
Press Ctrl+C at any time to cancel without deploying.

The container will be encrypted at rest and replicated across 3 regions.

Examples:
  moltbunker deploy                           # Interactive wizard
  moltbunker deploy nginx:latest              # Direct deploy with defaults
  moltbunker deploy myimage:tag --tor-only    # Deploy with Tor-only networking
  moltbunker deploy myimage:tag --onion-service --port 8080
  moltbunker deploy myimage:tag --min-tier confidential`,
		Args: cobra.MaximumNArgs(1),
		RunE: runDeploy,
	}

	cmd.Flags().BoolVar(&deployTorOnly, "tor-only", false, "Deploy with Tor-only networking")
	cmd.Flags().BoolVar(&deployOnionService, "onion-service", false, "Request .onion address")
	cmd.Flags().IntVar(&deployOnionPort, "port", 80, "Port to expose via onion service")
	cmd.Flags().Int64Var(&deployCPU, "cpu", 100000, "CPU quota in microseconds")
	cmd.Flags().Int64Var(&deployMemory, "memory", 536870912, "Memory limit in bytes")
	cmd.Flags().Int64Var(&deployDisk, "disk", 10737418240, "Disk limit in bytes")
	cmd.Flags().StringVar(&deployMinTier, "min-tier", "", "Minimum provider tier (confidential, standard, dev)")

	return cmd
}

func runDeploy(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return runDeployWizard()
	}
	return runDeployDirect(args[0])
}

func runDeployWizard() error {
	if !isTTY() {
		return fmt.Errorf("interactive deploy requires a terminal. Use: moltbunker deploy <image>")
	}

	var (
		image        string
		tierIdx      int
		customCPU    string
		customMem    string
		customDisk   string
		networking   []string
		onionPort    string
		securityTier string
		region       string
		confirm      bool
	)

	// Build tier options
	tierOptions := make([]huh.Option[int], len(resourceTiers))
	for i, t := range resourceTiers {
		if t.cpu == 0 {
			tierOptions[i] = huh.NewOption(t.name, i)
		} else {
			tierOptions[i] = huh.NewOption(fmt.Sprintf("%s  %s", t.name, StyleMuted.Render(t.bunker)), i)
		}
	}

	isCustomTier := func() bool {
		return tierIdx == len(resourceTiers)-1
	}

	hasOnionService := func() bool {
		for _, n := range networking {
			if n == "onion-service" {
				return true
			}
		}
		return false
	}

	// Single form — Shift+Tab to go back between all groups
	form := huh.NewForm(
		// Group 1: Image
		huh.NewGroup(
			huh.NewInput().
				Title("Container image").
				Description("Docker image to deploy (e.g., nginx:latest, alpine:3.19)").
				Placeholder("image:tag").
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("image is required")
					}
					if !strings.Contains(s, ":") && !strings.Contains(s, "/") {
						return fmt.Errorf("include a tag (e.g., %s:latest)", s)
					}
					return nil
				}).
				Value(&image),
		),

		// Group 2: Resource tier
		huh.NewGroup(
			huh.NewSelect[int]().
				Title("Resource tier").
				Description("How much compute to allocate").
				Options(tierOptions...).
				Value(&tierIdx),
		),

		// Group 3: Custom resources (only shown for Custom tier)
		huh.NewGroup(
			huh.NewInput().
				Title("CPU (vCPUs)").
				Description("Number of virtual CPUs (e.g., 1, 2, 4, 8)").
				Placeholder("2").
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("CPU is required for custom tier")
					}
					var v int
					if _, err := fmt.Sscanf(s, "%d", &v); err != nil || v < 1 || v > 64 {
						return fmt.Errorf("must be 1-64")
					}
					return nil
				}).
				Value(&customCPU),
			huh.NewInput().
				Title("Memory (MB)").
				Description("Memory limit in megabytes (e.g., 512, 1024, 4096)").
				Placeholder("1024").
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("memory is required for custom tier")
					}
					var v int
					if _, err := fmt.Sscanf(s, "%d", &v); err != nil || v < 128 || v > 65536 {
						return fmt.Errorf("must be 128-65536 MB")
					}
					return nil
				}).
				Value(&customMem),
			huh.NewInput().
				Title("Disk (MB)").
				Description("Disk storage limit in megabytes (e.g., 5120, 10240)").
				Placeholder("10240").
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("disk is required for custom tier")
					}
					var v int
					if _, err := fmt.Sscanf(s, "%d", &v); err != nil || v < 512 || v > 1048576 {
						return fmt.Errorf("must be 512-1048576 MB")
					}
					return nil
				}).
				Value(&customDisk),
		).WithHideFunc(func() bool {
			return !isCustomTier()
		}),

		// Group 4: Networking
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Networking options").
				Description("Select any combination (or none)").
				Options(
					huh.NewOption("Tor-only networking (all traffic routed through Tor)", "tor-only"),
					huh.NewOption(".onion hidden service address", "onion-service"),
				).
				Value(&networking),
		),

		// Group 5: Onion port (only if onion-service selected)
		huh.NewGroup(
			huh.NewInput().
				Title("Onion service port").
				Description("Port to expose via the .onion address").
				Placeholder("80").
				Validate(func(s string) error {
					if s == "" {
						return nil // default to 80
					}
					var p int
					if _, err := fmt.Sscanf(s, "%d", &p); err != nil || p < 1 || p > 65535 {
						return fmt.Errorf("port must be 1-65535")
					}
					return nil
				}).
				Value(&onionPort),
		).WithHideFunc(func() bool {
			return !hasOnionService()
		}),

		// Group 6: Security tier
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Security tier").
				Description("Minimum provider isolation level").
				Options(
					huh.NewOption("Any provider (fastest availability)", ""),
					huh.NewOption("Standard (Linux, VM isolation)", "standard"),
					huh.NewOption("Confidential (SEV-SNP, hardware encryption)", "confidential"),
				).
				Value(&securityTier),
		),

		// Group 7: Region preference
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Region preference").
				Description("Preferred geographic region for the primary replica").
				Options(
					huh.NewOption("Auto (best availability)", "auto"),
					huh.NewOption("Europe", "eu"),
					huh.NewOption("Americas", "us"),
					huh.NewOption("Asia-Pacific", "ap"),
				).
				Value(&region),
		),

		// Group 8: Confirmation with full summary
		huh.NewGroup(
			huh.NewConfirm().
				Title("Deploy this container?").
				DescriptionFunc(func() string {
					lines := []string{
						fmt.Sprintf("Image:   %s", image),
					}
					if isCustomTier() {
						lines = append(lines, fmt.Sprintf("CPU:     %s vCPU", customCPU))
						lines = append(lines, fmt.Sprintf("Memory:  %s MB", customMem))
						lines = append(lines, fmt.Sprintf("Disk:    %s MB", customDisk))
					} else {
						t := resourceTiers[tierIdx]
						lines = append(lines, fmt.Sprintf("Tier:    %s", t.name))
						lines = append(lines, fmt.Sprintf("Cost:    %s", t.bunker))
					}
					if securityTier != "" {
						lines = append(lines, fmt.Sprintf("Tier:    %s", securityTier))
					}
					if region != "" && region != "auto" {
						lines = append(lines, fmt.Sprintf("Region:  %s", region))
					}
					var netOpts []string
					for _, n := range networking {
						switch n {
						case "tor-only":
							netOpts = append(netOpts, "Tor-only")
						case "onion-service":
							port := onionPort
							if port == "" {
								port = "80"
							}
							netOpts = append(netOpts, fmt.Sprintf(".onion:%s", port))
						}
					}
					if len(netOpts) > 0 {
						lines = append(lines, fmt.Sprintf("Network: %s", strings.Join(netOpts, ", ")))
					}
					lines = append(lines, "Replicas: 3 (encrypted, cross-region)")
					return strings.Join(lines, "\n")
				}, &confirm).
				Affirmative("Deploy").
				Negative("Cancel").
				Value(&confirm),
		),
	).WithTheme(huh.ThemeBase())

	if err := form.Run(); err != nil {
		return err
	}

	if !confirm {
		Info("Deployment cancelled")
		return nil
	}

	// Apply tier
	if isCustomTier() {
		var cpu, mem, disk int
		fmt.Sscanf(customCPU, "%d", &cpu)
		fmt.Sscanf(customMem, "%d", &mem)
		fmt.Sscanf(customDisk, "%d", &disk)
		deployCPU = int64(cpu) * 50000 // vCPU to microseconds
		deployMemory = int64(mem) * 1024 * 1024
		deployDisk = int64(disk) * 1024 * 1024
	} else {
		tier := resourceTiers[tierIdx]
		deployCPU = tier.cpu
		deployMemory = tier.memMB * 1024 * 1024
		deployDisk = tier.diskMB * 1024 * 1024
	}

	// Apply networking
	for _, n := range networking {
		switch n {
		case "tor-only":
			deployTorOnly = true
		case "onion-service":
			deployOnionService = true
		}
	}
	if onionPort != "" {
		fmt.Sscanf(onionPort, "%d", &deployOnionPort)
	}

	// Apply security tier from wizard
	if securityTier != "" {
		deployMinTier = securityTier
	}

	return runDeployDirect(image)
}

func runDeployDirect(image string) error {
	c, err := GetClient()
	if err != nil {
		return err
	}
	defer c.Close()

	Info(fmt.Sprintf("Deploying: %s", image))

	req := &client.DeployRequest{
		Image: image,
		Resources: &client.ResourceLimits{
			CPUShares: deployCPU,
			MemoryMB:  deployMemory / (1024 * 1024),
			StorageMB: deployDisk / (1024 * 1024),
		},
		TorOnly:         deployTorOnly,
		OnionService:    deployOnionService,
		OnionPort:       deployOnionPort,
		MinProviderTier: deployMinTier,
	}

	var resp *client.DeployResponse
	err = WithSpinner("Submitting deployment", func() error {
		var e error
		resp, e = c.Deploy(req)
		return e
	})
	if err != nil {
		return fmt.Errorf("deployment failed: %w", err)
	}

	fields := [][2]string{
		{"Container", FormatNodeID(resp.ContainerID)},
		{"Status", resp.Status},
	}
	if resp.OnionAddress != "" {
		fields = append(fields, [2]string{"Onion", resp.OnionAddress})
	}
	if resp.ReplicaCount > 0 {
		fields = append(fields, [2]string{"Replicas", fmt.Sprintf("%d", resp.ReplicaCount)})
	}

	fmt.Println(StatusBox("Deployed", fields))
	fmt.Println(Hint(fmt.Sprintf("View logs: moltbunker logs %s", FormatNodeID(resp.ContainerID))))

	return nil
}
