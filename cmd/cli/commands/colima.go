package commands

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"
)

// NewColimaCmd creates the colima command for managing Colima container runtime
func NewColimaCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "colima",
		Short: "Manage Colima container runtime (macOS)",
		Long:  `Manage the Colima container runtime for running containers on macOS.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if runtime.GOOS != "darwin" {
				return fmt.Errorf("colima commands are only available on macOS")
			}
			return nil
		},
	}

	cmd.AddCommand(
		newColimaStartCmd(),
		newColimaStopCmd(),
		newColimaStatusCmd(),
	)

	return cmd
}

func newColimaStartCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Start Colima",
		Long:  `Start the Colima virtual machine with containerd runtime.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runColimaCommand("start")
		},
	}
}

func newColimaStopCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop Colima",
		Long:  `Stop the Colima virtual machine.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runColimaCommand("stop")
		},
	}
}

func newColimaStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Check Colima status",
		Long:  `Check the current status of the Colima virtual machine.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runColimaCommand("status")
		},
	}
}

func runColimaCommand(action string) error {
	path, err := exec.LookPath("colima")
	if err != nil {
		return fmt.Errorf("colima not found (install with: brew install colima)")
	}

	cmd := exec.Command(path, action)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
