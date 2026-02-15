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
			fmt.Println(StatusBox(Logo()+" CLI", [][2]string{
				{"Version", GetVersion()},
				{"Commit", GetCommit()},
				{"Build Date", BuildDate},
				{"Go", GetGoVersion()},
				{"OS/Arch", runtime.GOOS + "/" + runtime.GOARCH},
			}))
		},
	}
}
