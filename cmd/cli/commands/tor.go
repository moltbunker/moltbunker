package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewTorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tor",
		Short: "Tor management commands",
		Long:  "Manage Tor integration for anonymous networking.",
	}

	cmd.AddCommand(NewTorStartCmd())
	cmd.AddCommand(NewTorStatusCmd())
	cmd.AddCommand(NewTorOnionCmd())
	cmd.AddCommand(NewTorRotateCmd())

	return cmd
}

func NewTorStartCmd() *cobra.Command {
	var exitCountry string

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start Tor service",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := GetClient()
			if err != nil {
				return err
			}
			defer c.Close()

			var result map[string]interface{}
			err = WithSpinner("Starting Tor service", func() error {
				var e error
				result, e = c.TorStart()
				return e
			})
			if err != nil {
				return fmt.Errorf("failed to start Tor: %w", err)
			}

			status, _ := result["status"].(string)
			switch status {
			case "started":
				Success("Tor service started")
			case "already_running":
				Info("Tor service is already running")
			}

			if addr, ok := result["address"].(string); ok && addr != "" {
				fmt.Println(KeyValue("Onion", addr))
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&exitCountry, "exit-country", "", "Preferred exit node country (e.g., US, DE)")
	return cmd
}

func NewTorStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show Tor status",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := GetClient()
			if err != nil {
				return err
			}
			defer c.Close()

			status, err := c.TorStatus()
			if err != nil {
				return fmt.Errorf("failed to get Tor status: %w", err)
			}

			fields := [][2]string{}
			if status.Running {
				fields = append(fields, [2]string{"Status", StatusBadge("running")})
				if status.OnionAddress != "" {
					fields = append(fields, [2]string{"Onion", status.OnionAddress})
				}
				if !status.StartedAt.IsZero() {
					fields = append(fields, [2]string{"Started", status.StartedAt.Format("2006-01-02 15:04:05")})
				}
				fields = append(fields, [2]string{"Circuits", fmt.Sprintf("%d", status.CircuitCount)})
			} else {
				fields = append(fields, [2]string{"Status", StatusBadge("stopped")})
				fields = append(fields, [2]string{"", Hint("Start with: moltbunker tor start")})
			}

			fmt.Println(StatusBox("Tor Service", fields))
			return nil
		},
	}
}

func NewTorOnionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "onion",
		Short: "Show .onion address",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := GetClient()
			if err != nil {
				return err
			}
			defer c.Close()

			status, err := c.TorStatus()
			if err != nil {
				return fmt.Errorf("failed to get Tor status: %w", err)
			}
			if !status.Running {
				return fmt.Errorf("Tor is not running. Start with: moltbunker tor start")
			}
			if status.OnionAddress == "" {
				return fmt.Errorf("no onion address available")
			}

			fmt.Println(StatusBox("Onion Address", [][2]string{
				{"Address", status.OnionAddress},
			}))
			return nil
		},
	}
}

func NewTorRotateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rotate",
		Short: "Rotate Tor circuit",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := GetClient()
			if err != nil {
				return err
			}
			defer c.Close()

			err = WithSpinner("Rotating Tor circuit", func() error {
				return c.TorRotate()
			})
			if err != nil {
				return fmt.Errorf("failed to rotate circuit: %w", err)
			}

			Success("Tor circuit rotated — new identity active")
			return nil
		},
	}
}
