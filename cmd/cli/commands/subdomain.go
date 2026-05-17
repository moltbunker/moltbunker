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
		newSubdomainRenewCmd(),
		newSubdomainReserveCmd(),
		newSubdomainClaimCmd(),
		newSubdomainCancelCmd(),
		newSubdomainMetadataCmd(),
		newSubdomainPrimaryCmd(),
		newSubdomainReclaimCmd(),
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
	_ = cmd.MarkFlagRequired("deployment")

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
	_ = cmd.MarkFlagRequired("to")

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
	_ = cmd.MarkFlagRequired("deployment")

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

func newSubdomainRenewCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "renew <name>",
		Short: "Extend subdomain expiration by 365 days",
		Long: `Renew a subdomain to extend its expiration by another year.
Anyone can renew any name (useful for keeping names alive).
Costs the registration fee for the name.`,
		Args: cobra.ExactArgs(1),
		RunE: runSubdomainRenew,
	}
}

func newSubdomainReserveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "reserve <name>",
		Short: "Reserve a subdomain name for 48 hours",
		Long: `Reserve a subdomain name without assigning a deployment yet.
The reservation lasts 48 hours. Use 'subdomain claim' to finalize
with a deployment ID, or 'subdomain cancel' to release.`,
		Args: cobra.ExactArgs(1),
		RunE: runSubdomainReserve,
	}
}

func newSubdomainClaimCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "claim <name>",
		Short: "Finalize a reserved subdomain",
		Long: `Claim a previously reserved subdomain by assigning a deployment ID.
Must be called within the 48-hour reservation window.`,
		Args: cobra.ExactArgs(1),
		RunE: runSubdomainClaim,
	}

	cmd.Flags().StringP("deployment", "d", "", "Deployment ID to point to (required)")
	_ = cmd.MarkFlagRequired("deployment")

	return cmd
}

func newSubdomainCancelCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "cancel <name>",
		Short: "Cancel a subdomain reservation",
		Long:  `Cancel a pending subdomain reservation, making the name available again.`,
		Args:  cobra.ExactArgs(1),
		RunE:  runSubdomainCancel,
	}
}

func newSubdomainMetadataCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "metadata <name>",
		Short: "Set subdomain description and avatar",
		Long: `Set metadata (description, avatar URL) for a subdomain you own.
Costs the change fee.`,
		Args: cobra.ExactArgs(1),
		RunE: runSubdomainMetadata,
	}

	cmd.Flags().String("description", "", "Description for the subdomain")
	cmd.Flags().String("avatar", "", "Avatar URL for the subdomain")

	return cmd
}

func newSubdomainPrimaryCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "primary <name>",
		Short: "Set as primary name for reverse resolution",
		Long:  `Set a subdomain as the primary name, enabling reverse resolution from deployment ID to name.`,
		Args:  cobra.ExactArgs(1),
		RunE:  runSubdomainPrimary,
	}
}

func newSubdomainReclaimCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "reclaim <name>",
		Short: "Reclaim a squatted subdomain",
		Long: `Reclaim a subdomain that has been squatted (registered but pointing to
a non-existent or inactive deployment). Anyone can call this.`,
		Args: cobra.ExactArgs(1),
		RunE: runSubdomainReclaim,
	}
}

func runSubdomainRenew(cmd *cobra.Command, args []string) error {
	name := strings.ToLower(args[0])

	c, err := GetClient()
	if err != nil {
		return err
	}
	defer c.Close()

	if err := c.RequireDaemon(); err != nil {
		return err
	}

	err = WithSpinner("Renewing subdomain", func() error {
		return c.DaemonClient().SubdomainRenew(name)
	})
	if err != nil {
		return fmt.Errorf("renewal failed: %w", err)
	}

	Success(fmt.Sprintf("Renewed subdomain: %s (+365 days)", name))
	return nil
}

func runSubdomainReserve(cmd *cobra.Command, args []string) error {
	name := strings.ToLower(args[0])

	c, err := GetClient()
	if err != nil {
		return err
	}
	defer c.Close()

	if err := c.RequireDaemon(); err != nil {
		return err
	}

	err = WithSpinner("Reserving subdomain", func() error {
		return c.DaemonClient().SubdomainReserve(name)
	})
	if err != nil {
		return fmt.Errorf("reservation failed: %w", err)
	}

	fields := [][2]string{
		{"Name", name},
		{"Status", "Reserved (48h)"},
	}
	fmt.Println(StatusBox("Subdomain Reserved", fields))
	fmt.Println(Hint("Finalize with: moltbunker subdomain claim " + name + " --deployment <id>"))
	return nil
}

func runSubdomainClaim(cmd *cobra.Command, args []string) error {
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

	err = WithSpinner("Claiming subdomain", func() error {
		return c.DaemonClient().SubdomainClaim(name, deploymentID)
	})
	if err != nil {
		return fmt.Errorf("claim failed: %w", err)
	}

	fields := [][2]string{
		{"Name", name},
		{"Deployment", FormatNodeID(deploymentID)},
		{"URL", fmt.Sprintf("https://%s.moltbunker.dev", name)},
	}
	fmt.Println(StatusBox("Subdomain Claimed", fields))
	return nil
}

func runSubdomainCancel(cmd *cobra.Command, args []string) error {
	name := strings.ToLower(args[0])

	c, err := GetClient()
	if err != nil {
		return err
	}
	defer c.Close()

	if err := c.RequireDaemon(); err != nil {
		return err
	}

	err = WithSpinner("Cancelling reservation", func() error {
		return c.DaemonClient().SubdomainCancel(name)
	})
	if err != nil {
		return fmt.Errorf("cancellation failed: %w", err)
	}

	Success(fmt.Sprintf("Cancelled reservation: %s", name))
	return nil
}

func runSubdomainMetadata(cmd *cobra.Command, args []string) error {
	name := strings.ToLower(args[0])
	description, _ := cmd.Flags().GetString("description")
	avatarURL, _ := cmd.Flags().GetString("avatar")

	c, err := GetClient()
	if err != nil {
		return err
	}
	defer c.Close()

	if err := c.RequireDaemon(); err != nil {
		return err
	}

	err = WithSpinner("Updating metadata", func() error {
		return c.DaemonClient().SubdomainSetMetadata(name, description, avatarURL)
	})
	if err != nil {
		return fmt.Errorf("metadata update failed: %w", err)
	}

	Success(fmt.Sprintf("Updated metadata for: %s", name))
	return nil
}

func runSubdomainPrimary(cmd *cobra.Command, args []string) error {
	name := strings.ToLower(args[0])

	c, err := GetClient()
	if err != nil {
		return err
	}
	defer c.Close()

	if err := c.RequireDaemon(); err != nil {
		return err
	}

	err = WithSpinner("Setting primary name", func() error {
		return c.DaemonClient().SubdomainSetPrimary(name)
	})
	if err != nil {
		return fmt.Errorf("set primary failed: %w", err)
	}

	Success(fmt.Sprintf("Set primary name: %s", name))
	return nil
}

func runSubdomainReclaim(cmd *cobra.Command, args []string) error {
	name := strings.ToLower(args[0])

	c, err := GetClient()
	if err != nil {
		return err
	}
	defer c.Close()

	if err := c.RequireDaemon(); err != nil {
		return err
	}

	err = WithSpinner("Reclaiming subdomain", func() error {
		return c.DaemonClient().SubdomainReclaim(name)
	})
	if err != nil {
		return fmt.Errorf("reclaim failed: %w", err)
	}

	Success(fmt.Sprintf("Reclaimed subdomain: %s", name))
	return nil
}
