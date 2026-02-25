package commands

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// NewSubdomainCmd creates the subdomain management command group.
func NewSubdomainCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "subdomain",
		Aliases: []string{"sub", "domain"},
		Short:   "Manage vanity subdomains",
		Long: `Register and manage vanity subdomains on moltbunker.dev.

Every deployment automatically gets a subdomain based on its ID prefix
(e.g., dep-a1b2c3d4 → a1b2c3d4.moltbunker.dev). Vanity subdomains let
you pick a custom name (e.g., myapp.moltbunker.dev) for 1,000,000 BUNKER.`,
	}

	cmd.AddCommand(
		newSubdomainRegisterCmd(),
		newSubdomainListCmd(),
		newSubdomainReleaseCmd(),
		newSubdomainResolveCmd(),
		newSubdomainTransferCmd(),
		newSubdomainUpdateCmd(),
	)

	return cmd
}

func newSubdomainRegisterCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register <name>",
		Short: "Register a vanity subdomain",
		Long: `Register a vanity subdomain pointing to a deployment.

The name must be 3-32 characters, lowercase alphanumeric with hyphens,
and cannot start or end with a hyphen. Registration costs 1,000,000 BUNKER
(80% burned, 20% treasury).

Examples:
  moltbunker subdomain register myapp --deployment dep-a1b2c3d4
  moltbunker subdomain register api-v2 --deployment dep-deadbeef`,
		Args: cobra.ExactArgs(1),
		RunE: runSubdomainRegister,
	}

	cmd.Flags().StringP("deployment", "d", "", "Deployment ID to point to (required)")
	cmd.MarkFlagRequired("deployment")

	return cmd
}

func newSubdomainListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List your registered subdomains",
		Args:  cobra.NoArgs,
		RunE:  runSubdomainList,
	}
}

func newSubdomainReleaseCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "release <name>",
		Short: "Release a subdomain you own",
		Long: `Release a vanity subdomain, making the name available again.

Note: The registration fee is NOT refunded.`,
		Args: cobra.ExactArgs(1),
		RunE: runSubdomainRelease,
	}
}

func newSubdomainResolveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "resolve <name>",
		Short: "Resolve a subdomain to its deployment",
		Args:  cobra.ExactArgs(1),
		RunE:  runSubdomainResolve,
	}
}

func newSubdomainTransferCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transfer <name>",
		Short: "Transfer subdomain ownership",
		Args:  cobra.ExactArgs(1),
		RunE:  runSubdomainTransfer,
	}

	cmd.Flags().String("to", "", "New owner wallet address (required)")
	cmd.MarkFlagRequired("to")

	return cmd
}

func newSubdomainUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <name>",
		Short: "Update subdomain's target deployment",
		Args:  cobra.ExactArgs(1),
		RunE:  runSubdomainUpdate,
	}

	cmd.Flags().StringP("deployment", "d", "", "New deployment ID (required)")
	cmd.MarkFlagRequired("deployment")

	return cmd
}

func runSubdomainRegister(cmd *cobra.Command, args []string) error {
	name := strings.ToLower(args[0])
	deploymentID, _ := cmd.Flags().GetString("deployment")

	c, err := GetClient()
	if err != nil {
		return err
	}
	defer c.Close()

	if err := c.RequireDaemon(); err != nil {
		return err
	}

	Info(fmt.Sprintf("Registering subdomain: %s", name))

	var url string
	err = WithSpinner("Registering on-chain", func() error {
		resp, e := c.DaemonClient().SubdomainRegister(name, deploymentID)
		if e != nil {
			return e
		}
		url = resp.URL
		return nil
	})
	if err != nil {
		return fmt.Errorf("registration failed: %w", err)
	}

	if url == "" {
		url = fmt.Sprintf("https://%s.moltbunker.dev", name)
	}

	fields := [][2]string{
		{"Name", name},
		{"Deployment", FormatNodeID(deploymentID)},
		{"URL", url},
	}

	fmt.Println(StatusBox("Subdomain Registered", fields))

	return nil
}

func runSubdomainList(cmd *cobra.Command, args []string) error {
	c, err := GetClient()
	if err != nil {
		return err
	}
	defer c.Close()

	if err := c.RequireDaemon(); err != nil {
		return err
	}

	subs, err := c.DaemonClient().SubdomainList()
	if err != nil {
		return fmt.Errorf("failed to list subdomains: %w", err)
	}

	if len(subs) == 0 {
		Info("No subdomains registered")
		fmt.Println(Hint("Register one: moltbunker subdomain register <name> --deployment <id>"))
		return nil
	}

	for _, sub := range subs {
		fields := [][2]string{
			{"Deployment", FormatNodeID(sub.DeploymentID)},
			{"URL", sub.URL},
			{"Owner", FormatNodeID(sub.Owner)},
		}
		fmt.Println(StatusBox(sub.Name, fields))
	}

	return nil
}

func runSubdomainRelease(cmd *cobra.Command, args []string) error {
	name := strings.ToLower(args[0])

	c, err := GetClient()
	if err != nil {
		return err
	}
	defer c.Close()

	if err := c.RequireDaemon(); err != nil {
		return err
	}

	err = WithSpinner("Releasing subdomain", func() error {
		return c.DaemonClient().SubdomainRelease(name)
	})
	if err != nil {
		return fmt.Errorf("release failed: %w", err)
	}

	Success(fmt.Sprintf("Released subdomain: %s", name))

	return nil
}

func runSubdomainResolve(cmd *cobra.Command, args []string) error {
	name := strings.ToLower(args[0])

	c, err := GetClient()
	if err != nil {
		return err
	}
	defer c.Close()

	if err := c.RequireDaemon(); err != nil {
		return err
	}

	info, err := c.DaemonClient().SubdomainResolve(name)
	if err != nil {
		return fmt.Errorf("resolve failed: %w", err)
	}

	fields := [][2]string{
		{"Name", info.Name},
		{"Deployment", FormatNodeID(info.DeploymentID)},
		{"Owner", FormatNodeID(info.Owner)},
		{"URL", info.URL},
		{"Registered", info.RegisteredAt.Format("2006-01-02 15:04:05")},
	}

	fmt.Println(StatusBox("Subdomain", fields))

	return nil
}

func runSubdomainTransfer(cmd *cobra.Command, args []string) error {
	name := strings.ToLower(args[0])
	newOwner, _ := cmd.Flags().GetString("to")

	c, err := GetClient()
	if err != nil {
		return err
	}
	defer c.Close()

	if err := c.RequireDaemon(); err != nil {
		return err
	}

	err = WithSpinner("Transferring ownership", func() error {
		return c.DaemonClient().SubdomainTransfer(name, newOwner)
	})
	if err != nil {
		return fmt.Errorf("transfer failed: %w", err)
	}

	Success(fmt.Sprintf("Transferred %s to %s", name, FormatNodeID(newOwner)))

	return nil
}

func runSubdomainUpdate(cmd *cobra.Command, args []string) error {
	name := strings.ToLower(args[0])
	deploymentID, _ := cmd.Flags().GetString("deployment")

	c, err := GetClient()
	if err != nil {
		return err
	}
	defer c.Close()

	if err := c.RequireDaemon(); err != nil {
		return err
	}

	err = WithSpinner("Updating subdomain", func() error {
		return c.DaemonClient().SubdomainUpdate(name, deploymentID)
	})
	if err != nil {
		return fmt.Errorf("update failed: %w", err)
	}

	fields := [][2]string{
		{"Name", name},
		{"Deployment", FormatNodeID(deploymentID)},
		{"URL", fmt.Sprintf("https://%s.moltbunker.dev", name)},
	}

	fmt.Println(StatusBox("Subdomain Updated", fields))

	return nil
}
