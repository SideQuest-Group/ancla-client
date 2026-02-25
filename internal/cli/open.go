package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	openCmd.Flags().Bool("dashboard", false, "Open the dashboard home, ignoring any linked org/project/app")
	rootCmd.AddCommand(openCmd)
}

var openCmd = &cobra.Command{
	Use:   "open",
	Short: "Open the Ancla dashboard in your browser",
	Long: `Open the Ancla dashboard in your default web browser.

When a link context is set (org, project, or app), the command opens the
most specific page available. Use --dashboard to ignore the link context
and open the dashboard home instead.`,
	Example: `  ancla open
  ancla open --dashboard`,
	GroupID: "workflow",
	RunE: func(cmd *cobra.Command, args []string) error {
		dashOnly, _ := cmd.Flags().GetBool("dashboard")

		url := serverURL() + "/dashboard"

		if !dashOnly {
			if cfg.Org != "" {
				url += "/orgs/" + cfg.Org
				if cfg.Project != "" {
					url += "/projects/" + cfg.Project
					if cfg.App != "" {
						url += "/apps/" + cfg.App
					}
				}
			}
		}

		fmt.Println("Opening", url)
		return openBrowser(url)
	},
}
