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
	Use:   "link <ws>[/<proj>[/<env>[/<svc>]]]",
	Short: "Associate this directory with a workspace, project, env, or service",
	Long: `Associate the current directory with an Ancla workspace, project,
environment, or service.

This creates a local .ancla/config.yaml that stores the link context so
subsequent commands (status, logs, run, deploy) can infer the target
without requiring explicit arguments.

Examples:
  ancla link my-ws                              # link to workspace only
  ancla link my-ws/my-proj                      # link to workspace and project
  ancla link my-ws/my-proj/staging              # link to workspace, project, and env
  ancla link my-ws/my-proj/staging/my-svc       # link to all four segments`,
	Example: "  ancla link my-ws/my-proj/staging/my-svc",
	GroupID: "auth",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		parts := strings.Split(args[0], "/")

		cfg.Workspace = parts[0]
		if len(parts) >= 2 {
			cfg.Project = parts[1]
		}
		if len(parts) >= 3 {
			cfg.Env = parts[2]
		}
		if len(parts) >= 4 {
			cfg.Service = parts[3]
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
	Short:   "Remove the directory link to a workspace/project/env/service",
	Long:    "Remove the local .ancla/config.yaml that associates this directory with an Ancla resource.",
	Example: "  ancla unlink",
	GroupID: "auth",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.RemoveLocal(); err != nil {
			return err
		}
		cfg.Workspace = ""
		cfg.Project = ""
		cfg.Env = ""
		cfg.Service = ""
		fmt.Println("Unlinked.")
		return nil
	},
}
