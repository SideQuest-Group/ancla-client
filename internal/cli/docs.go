package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

const docsBaseURL = "https://docs.ancla.dev"

func init() {
	rootCmd.AddCommand(docsCmd)
}

var docsCmd = &cobra.Command{
	Use:   "docs [topic]",
	Short: "Open Ancla documentation in your browser",
	Long: `Open the Ancla documentation site in your default browser.

Optionally provide a topic (e.g. "api", "cli") to jump directly to that
section. The topic is appended as a path segment to the docs base URL.`,
	Example: `  ancla docs
  ancla docs api
  ancla docs cli`,
	GroupID:           "workflow",
	Args:              cobra.MaximumNArgs(1),
	ValidArgsFunction: completeDocTopics,
	RunE: func(cmd *cobra.Command, args []string) error {
		url := docsBaseURL
		if len(args) == 1 {
			topic := strings.TrimLeft(args[0], "/")
			url = docsBaseURL + "/" + topic
		}

		if err := openBrowser(url); err != nil {
			fmt.Fprintf(cmd.OutOrStdout(), "Could not open browser: %v\nOpen manually: %s\n", err, url)
			return nil
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Opening %s ...\n", url)
		return nil
	},
}

// completeDocTopics provides shell completion for common doc topics.
func completeDocTopics(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	return []string{
		"api\tAPI reference",
		"cli\tCLI reference",
	}, cobra.ShellCompDirectiveNoFileComp
}
