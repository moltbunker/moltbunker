package commands

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

func NewVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Long:  "Display the version of moltbunker CLI and build information.",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Moltbunker CLI")
			fmt.Println("==============")
			fmt.Printf("Version:    %s\n", GetVersion())
			fmt.Printf("Commit:     %s\n", GetCommit())
			fmt.Printf("Build Date: %s\n", BuildDate)
			fmt.Printf("Go Version: %s\n", GetGoVersion())
			fmt.Printf("OS/Arch:    %s/%s\n", runtime.GOOS, runtime.GOARCH)
		},
	}
}
