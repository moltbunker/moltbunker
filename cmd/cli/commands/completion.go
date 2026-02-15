package commands

import (
	"os"

	"github.com/spf13/cobra"
)

// NewCompletionCmd creates the completion command for shell auto-completion.
func NewCompletionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate shell completion scripts",
		Long: `Generate shell completion scripts for moltbunker.

To load completions:

Bash:
  $ source <(moltbunker completion bash)
  # To load completions for each session, execute once:
  # Linux:
  $ moltbunker completion bash > /etc/bash_completion.d/moltbunker
  # macOS:
  $ moltbunker completion bash > $(brew --prefix)/etc/bash_completion.d/moltbunker

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it. Execute the following once:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc
  # To load completions for each session, execute once:
  $ moltbunker completion zsh > "${fpath[1]}/_moltbunker"
  # You will need to start a new shell for this setup to take effect.

Fish:
  $ moltbunker completion fish | source
  # To load completions for each session, execute once:
  $ moltbunker completion fish > ~/.config/fish/completions/moltbunker.fish

PowerShell:
  PS> moltbunker completion powershell | Out-String | Invoke-Expression
  # To load completions for every new session, add the output to your profile.
`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return cmd.Root().GenBashCompletionV2(os.Stdout, true)
			case "zsh":
				return cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				return cmd.Root().GenFishCompletion(os.Stdout, true)
			case "powershell":
				return cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
			}
			return nil
		},
	}
	return cmd
}
