package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewProxyCmd creates the proxy command group.
func NewProxyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "proxy",
		Short: "Manage the decentralized proxy service",
		Long: `Manage the decentralized SOCKS5/HTTP proxy service.

Route traffic through the Moltbunker P2P network with optional Tor exit.
Supports SOCKS5 and HTTP CONNECT protocols with bandwidth metering.

Examples:
  moltbunker proxy status              # Show proxy server status
  moltbunker proxy sessions            # List active proxy sessions
  moltbunker proxy usage               # Show bandwidth usage
  moltbunker proxy close <session-id>  # Close a proxy session`,
	}

	cmd.AddCommand(
		newProxyStatusCmd(),
		newProxySessionsCmd(),
		newProxyUsageCmd(),
		newProxyCloseCmd(),
	)

	return cmd
}

func newProxyStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show proxy server status",
		Args:  cobra.NoArgs,
		RunE:  runProxyStatus,
	}
}

func newProxySessionsCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "sessions",
		Aliases: []string{"ls", "list"},
		Short:   "List active proxy sessions",
		Args:    cobra.NoArgs,
		RunE:    runProxySessions,
	}
}

func newProxyUsageCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "usage",
		Short: "Show proxy bandwidth usage",
		Args:  cobra.NoArgs,
		RunE:  runProxyUsage,
	}
}

func newProxyCloseCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "close <session-id>",
		Short: "Close a proxy session",
		Args:  cobra.ExactArgs(1),
		RunE:  runProxyClose,
	}
}

// ── Handlers ─────────────────────────────────────────────────────────────────

func runProxyStatus(_ *cobra.Command, _ []string) error {
	c, err := GetClient()
	if err != nil {
		return err
	}
	defer c.Close()

	if err := c.RequireDaemon(); err != nil {
		return err
	}

	status, err := c.DaemonClient().ProxyStatus()
	if err != nil {
		return fmt.Errorf("failed to get proxy status: %w", err)
	}

	running := "stopped"
	if status.Running {
		running = "running"
	}

	fields := [][2]string{
		{"Status", running},
		{"SOCKS5", status.SOCKS5Addr},
		{"HTTP", status.HTTPAddr},
		{"Tor Exit", fmt.Sprintf("%v", status.UseTor)},
		{"Active Sessions", fmt.Sprintf("%d / %d", status.ActiveSessions, status.MaxSessions)},
	}

	fmt.Println(StatusBox("Proxy", fields))

	if !status.Running {
		fmt.Println(Hint("Proxy starts automatically when the daemon is configured with proxy.enabled=true"))
	}

	return nil
}

func runProxySessions(_ *cobra.Command, _ []string) error {
	c, err := GetClient()
	if err != nil {
		return err
	}
	defer c.Close()

	if err := c.RequireDaemon(); err != nil {
		return err
	}

	sessions, err := c.DaemonClient().ProxySessions()
	if err != nil {
		return fmt.Errorf("failed to list sessions: %w", err)
	}

	if len(sessions) == 0 {
		Info("No active proxy sessions")
		return nil
	}

	headers := []string{"ID", "Protocol", "Target", "Bytes In", "Bytes Out", "Started"}
	rows := make([][]string, 0, len(sessions))
	for _, s := range sessions {
		rows = append(rows, []string{
			FormatNodeID(s.ID),
			s.Protocol,
			s.Target,
			formatBytes(s.BytesIn),
			formatBytes(s.BytesOut),
			s.StartedAt.Format("15:04:05"),
		})
	}

	fmt.Println(RenderTable(headers, rows))

	return nil
}

func runProxyUsage(_ *cobra.Command, _ []string) error {
	c, err := GetClient()
	if err != nil {
		return err
	}
	defer c.Close()

	if err := c.RequireDaemon(); err != nil {
		return err
	}

	usage, err := c.DaemonClient().ProxyUsage()
	if err != nil {
		return fmt.Errorf("failed to get usage: %w", err)
	}

	fields := [][2]string{
		{"Wallet", FormatAddress(usage.Wallet)},
		{"Total In", formatBytes(usage.TotalIn)},
		{"Total Out", formatBytes(usage.TotalOut)},
		{"Sessions", fmt.Sprintf("%d", usage.SessionCount)},
	}

	fmt.Println(StatusBox("Proxy Usage", fields))

	return nil
}

func runProxyClose(_ *cobra.Command, args []string) error {
	sessionID := args[0]

	c, err := GetClient()
	if err != nil {
		return err
	}
	defer c.Close()

	if err := c.RequireDaemon(); err != nil {
		return err
	}

	if err := c.DaemonClient().ProxyCloseSession(sessionID); err != nil {
		return fmt.Errorf("close failed: %w", err)
	}

	Success(fmt.Sprintf("Closed session: %s", FormatNodeID(sessionID)))

	return nil
}

// formatBytes is already defined in snapshot.go — reused here.
