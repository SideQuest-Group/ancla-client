package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	openCmd.Flags().Bool("dashboard", false, "Open the dashboard home, ignoring any linked context")
	rootCmd.AddCommand(openCmd)
}

var openCmd = &cobra.Command{
	Use:   "open",
	Short: "Open the Ancla dashboard in your browser",
	Long: `Open the Ancla dashboard in your default web browser.

When a link context is set (workspace, project, env, or service), the command
opens the most specific page available. Use --dashboard to ignore the link
context and open the dashboard home instead.`,
	Example: `  ancla open
  ancla open --dashboard`,
	GroupID: "workflow",
	RunE: func(cmd *cobra.Command, args []string) error {
		dashOnly, _ := cmd.Flags().GetBool("dashboard")

		url := serverURL() + "/dashboard"

		if !dashOnly {
			if cfg.Workspace != "" {
				url += "/workspaces/" + cfg.Workspace
				if cfg.Project != "" {
					url += "/projects/" + cfg.Project
					if cfg.Env != "" {
						url += "/envs/" + cfg.Env
						if cfg.Service != "" {
							url += "/services/" + cfg.Service
						}
					}
				}
			}
		}

		fmt.Println("Opening", url)
		return openBrowser(url)
	},
}
