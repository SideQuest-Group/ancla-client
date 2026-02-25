package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/SideQuest-Group/ancla-client/internal/config"
)

func init() {
	rootCmd.AddCommand(linkCmd)
	rootCmd.AddCommand(unlinkCmd)
}

var linkCmd = &cobra.Command{
	Use:   "link <org>[/<project>[/<app>]]",
	Short: "Associate this directory with an org, project, or app",
	Long: `Associate the current directory with an Ancla org, project, or application.

This creates a local .ancla/config.yaml that stores the link context so
subsequent commands (status, logs, run, deploy) can infer the target
without requiring explicit arguments.

Examples:
  ancla link my-org                     # link to org only
  ancla link my-org/my-project          # link to org and project
  ancla link my-org/my-project/my-app   # link to org, project, and app`,
	Example: "  ancla link my-org/my-project/my-app",
	GroupID: "auth",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		parts := strings.Split(args[0], "/")

		cfg.Org = parts[0]
		if len(parts) >= 2 {
			cfg.Project = parts[1]
		}
		if len(parts) >= 3 {
			cfg.App = parts[2]
		}

		if err := config.SaveLocal(cfg); err != nil {
			return fmt.Errorf("saving link: %w", err)
		}

		fmt.Printf("Linked to %s\n", args[0])
		return nil
	},
}

var unlinkCmd = &cobra.Command{
	Use:     "unlink",
	Short:   "Remove the directory link to an org/project/app",
	Long:    "Remove the local .ancla/config.yaml that associates this directory with an Ancla resource.",
	Example: "  ancla unlink",
	GroupID: "auth",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.RemoveLocal(); err != nil {
			return err
		}
		cfg.Org = ""
		cfg.Project = ""
		cfg.App = ""
		fmt.Println("Unlinked.")
		return nil
	},
}
