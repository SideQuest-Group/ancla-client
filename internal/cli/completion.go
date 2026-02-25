package cli

import (
	"os"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(completionCmd)
}

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion script",
	Long: `Generate a shell completion script for ancla.

To load completions:

  bash:
    source <(ancla completion bash)
    # Or permanently: ancla completion bash > /etc/bash_completion.d/ancla

  zsh:
    source <(ancla completion zsh)
    # Or permanently: ancla completion zsh > "${fpath[1]}/_ancla"

  fish:
    ancla completion fish | source
    # Or permanently: ancla completion fish > ~/.config/fish/completions/ancla.fish

  powershell:
    ancla completion powershell | Out-String | Invoke-Expression
`,
	Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "bash":
			return rootCmd.GenBashCompletion(os.Stdout)
		case "zsh":
			return rootCmd.GenZshCompletion(os.Stdout)
		case "fish":
			return rootCmd.GenFishCompletion(os.Stdout, true)
		case "powershell":
			return rootCmd.GenPowerShellCompletionWithDesc(os.Stdout)
		}
		return nil
	},
}
