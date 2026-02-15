package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

// NewManCmd creates the man command for generating man pages.
func NewManCmd() *cobra.Command {
	var outputDir string

	cmd := &cobra.Command{
		Use:   "man",
		Short: "Generate man pages",
		Long:  "Generate man pages for moltbunker in the specified directory.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if outputDir == "" {
				outputDir = filepath.Join(".", "man")
			}
			if err := os.MkdirAll(outputDir, 0755); err != nil {
				return fmt.Errorf("failed to create output directory: %w", err)
			}

			header := &doc.GenManHeader{
				Title:   "MOLTBUNKER",
				Section: "1",
				Date:    &time.Time{},
				Source:  "Moltbunker " + GetVersion(),
				Manual:  "Moltbunker Manual",
			}

			now := time.Now()
			header.Date = &now

			if err := doc.GenManTree(cmd.Root(), header, outputDir); err != nil {
				return fmt.Errorf("failed to generate man pages: %w", err)
			}

			fmt.Printf("Man pages generated in %s\n", outputDir)
			return nil
		},
	}

	cmd.Flags().StringVarP(&outputDir, "dir", "d", "", "Output directory (default: ./man)")
	return cmd
}
