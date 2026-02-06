package commands

import (
	"fmt"
	"strings"

	"github.com/moltbunker/moltbunker/internal/client"
	"github.com/spf13/cobra"
)

func NewThreatCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "threat",
		Short: "Show threat level and manage signals",
		Long: `Display the current threat level and manage threat signals.

The threat detection system monitors for signs of hostile activity
and calculates an aggregate threat score (0.0-1.0).

Use subcommands to view signals or clear them.`,
		RunE: runThreatLevel,
	}

	cmd.AddCommand(newThreatSignalsCmd())
	cmd.AddCommand(newThreatClearCmd())

	return cmd
}

func runThreatLevel(cmd *cobra.Command, args []string) error {
	daemonClient := client.NewDaemonClient(SocketPath)

	if err := daemonClient.Connect(); err != nil {
		return fmt.Errorf("daemon not running: %w", err)
	}
	defer daemonClient.Close()

	threat, err := daemonClient.ThreatLevel()
	if err != nil {
		return fmt.Errorf("failed to get threat level: %w", err)
	}

	fmt.Println("Threat Assessment")
	fmt.Println("=================")
	fmt.Printf("Score:          %.2f\n", threat.Score)
	fmt.Printf("Level:          %s\n", formatThreatLevel(threat.Level))
	fmt.Printf("Recommendation: %s\n", formatRecommendation(threat.Recommendation))
	fmt.Printf("Active Signals: %d\n", len(threat.ActiveSignals))
	fmt.Printf("Timestamp:      %s\n", threat.Timestamp.Format("2006-01-02 15:04:05"))

	if len(threat.ActiveSignals) > 0 {
		fmt.Println("\nActive Signals")
		fmt.Println("--------------")
		for _, sig := range threat.ActiveSignals {
			fmt.Printf("  [%s] %s (score: %.2f, confidence: %.2f)\n",
				sig.Type, sig.Source, sig.Score, sig.Confidence)
			if sig.Details != "" {
				fmt.Printf("          %s\n", sig.Details)
			}
		}
	}

	return nil
}

func newThreatSignalsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "signals",
		Short: "List active threat signals",
		Long:  "Display detailed information about all active threat signals.",
		RunE:  runThreatSignals,
	}
}

func runThreatSignals(cmd *cobra.Command, args []string) error {
	daemonClient := client.NewDaemonClient(SocketPath)

	if err := daemonClient.Connect(); err != nil {
		return fmt.Errorf("daemon not running: %w", err)
	}
	defer daemonClient.Close()

	threat, err := daemonClient.ThreatLevel()
	if err != nil {
		return fmt.Errorf("failed to get threat level: %w", err)
	}

	if len(threat.ActiveSignals) == 0 {
		fmt.Println("No active threat signals")
		return nil
	}

	fmt.Println("Active Threat Signals")
	fmt.Println("=====================")
	fmt.Printf("Total: %d signals\n\n", len(threat.ActiveSignals))

	for i, sig := range threat.ActiveSignals {
		fmt.Printf("Signal #%d\n", i+1)
		fmt.Printf("  Type:       %s\n", sig.Type)
		fmt.Printf("  Score:      %.2f\n", sig.Score)
		fmt.Printf("  Confidence: %.2f\n", sig.Confidence)
		fmt.Printf("  Source:     %s\n", sig.Source)
		if sig.Details != "" {
			fmt.Printf("  Details:    %s\n", sig.Details)
		}
		fmt.Printf("  Timestamp:  %s\n", sig.Timestamp.Format("2006-01-02 15:04:05"))
		fmt.Println()
	}

	return nil
}

var (
	clearSignalType  string
	clearContainerID string
	clearAll         bool
)

func newThreatClearCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clear",
		Short: "Clear threat signals",
		Long: `Clear active threat signals.

Examples:
  moltbunker threat clear --all                     # Clear all signals
  moltbunker threat clear --type network_isolation  # Clear specific signal type
  moltbunker threat clear --container abc123        # Clear signals for container`,
		RunE: runThreatClear,
	}

	cmd.Flags().StringVar(&clearSignalType, "type", "", "Signal type to clear")
	cmd.Flags().StringVar(&clearContainerID, "container", "", "Container ID to clear signals for")
	cmd.Flags().BoolVar(&clearAll, "all", false, "Clear all signals")

	return cmd
}

func runThreatClear(cmd *cobra.Command, args []string) error {
	if !clearAll && clearSignalType == "" && clearContainerID == "" {
		return fmt.Errorf("specify --all, --type, or --container")
	}

	daemonClient := client.NewDaemonClient(SocketPath)

	if err := daemonClient.Connect(); err != nil {
		return fmt.Errorf("daemon not running: %w", err)
	}
	defer daemonClient.Close()

	if err := daemonClient.ThreatClear(clearSignalType, clearContainerID, clearAll); err != nil {
		return fmt.Errorf("failed to clear signals: %w", err)
	}

	if clearAll {
		fmt.Println("All threat signals cleared")
	} else if clearSignalType != "" {
		fmt.Printf("Signals of type '%s' cleared\n", clearSignalType)
	} else {
		fmt.Printf("Signals for container '%s' cleared\n", clearContainerID)
	}

	return nil
}

func formatThreatLevel(level string) string {
	switch level {
	case "critical":
		return "CRITICAL"
	case "high":
		return "HIGH"
	case "medium":
		return "MEDIUM"
	case "low":
		return "LOW"
	default:
		return strings.ToUpper(level)
	}
}

func formatRecommendation(rec string) string {
	switch rec {
	case "immediate_clone":
		return "IMMEDIATE CLONE - High threat detected"
	case "prepare_clone":
		return "PREPARE CLONE - Elevated threat"
	case "increase_monitoring":
		return "INCREASE MONITORING - Moderate threat"
	case "continue_normal":
		return "CONTINUE NORMAL - Low threat"
	default:
		return rec
	}
}
