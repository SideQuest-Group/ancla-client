package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/spf13/cobra"

	"github.com/SideQuest-Group/ancla-client/internal/config"
)

func init() {
	rootCmd.AddCommand(projectsCmd)
	projectsCmd.AddCommand(projectsListCmd)
	projectsCmd.AddCommand(projectsGetCmd)
}

var projectsCmd = &cobra.Command{
	Use:     "projects",
	Aliases: []string{"proj", "p"},
	Short:   "Manage projects",
	Long: `Manage projects within a workspace.

Projects group related services together under a workspace. Each project
can contain multiple environments and services that share the same
workspace-level permissions.
Use sub-commands to list all projects or inspect a specific one.`,
	Example: "  ancla projects list my-workspace\n  ancla projects get my-workspace/my-project",
	GroupID: "resources",
	RunE: func(cmd *cobra.Command, args []string) error {
		return projectsListCmd.RunE(cmd, args)
	},
}

var projectsListCmd = &cobra.Command{
	Use:     "list [workspace]",
	Short:   "List projects in a workspace",
	Example: "  ancla projects list my-workspace",
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var arg string
		if len(args) == 1 {
			arg = args[0]
		}
		ws, _, _, _, err := config.ResolveServicePath(arg, cfg)
		if err != nil {
			return err
		}
		if ws == "" {
			return fmt.Errorf("workspace is required\n\n  ancla projects <workspace>\n\n  Hierarchy: workspace → project → env → service\n  Hint: run `ancla link` to set a default workspace")
		}

		req, _ := http.NewRequest("GET", apiURL("/workspaces/"+ws+"/projects/"), nil)
		body, err := doRequest(req)
		if err != nil {
			return err
		}

		var projects []struct {
			ID            string `json:"id"`
			Name          string `json:"name"`
			Slug          string `json:"slug"`
			WorkspaceSlug string `json:"workspace_slug"`
			ServiceCount  int    `json:"service_count"`
		}
		if err := json.Unmarshal(body, &projects); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}

		if isJSON() {
			return printJSON(projects)
		}

		var rows [][]string
		for _, p := range projects {
			rows = append(rows, []string{p.WorkspaceSlug + "/" + p.Slug, p.Name, fmt.Sprintf("%d", p.ServiceCount)})
		}
		table([]string{"WS/PROJECT", "NAME", "SERVICES"}, rows)
		return nil
	},
}

var projectsGetCmd = &cobra.Command{
	Use:               "get <workspace>/<project>",
	Short:             "Get project details",
	Example:           "  ancla projects get my-workspace/my-project",
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeProjects,
	RunE: func(cmd *cobra.Command, args []string) error {
		parts := strings.SplitN(args[0], "/", 2)
		if len(parts) != 2 {
			return fmt.Errorf("argument must be in the form <workspace>/<project>")
		}
		ws, proj := parts[0], parts[1]

		req, _ := http.NewRequest("GET", apiURL("/workspaces/"+ws+"/projects/"+proj), nil)
		body, err := doRequest(req)
		if err != nil {
			return err
		}

		var project struct {
			Name          string `json:"name"`
			Slug          string `json:"slug"`
			WorkspaceSlug string `json:"workspace_slug"`
			WorkspaceName string `json:"workspace_name"`
			ServiceCount  int    `json:"service_count"`
			Created       string `json:"created"`
			Updated       string `json:"updated"`
		}
		if err := json.Unmarshal(body, &project); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}

		if isJSON() {
			return printJSON(project)
		}

		fmt.Printf("Project: %s (%s/%s)\n", project.Name, project.WorkspaceSlug, project.Slug)
		fmt.Printf("Workspace: %s\n", project.WorkspaceName)
		fmt.Printf("Services: %d\n", project.ServiceCount)
		if project.Created != "" {
			fmt.Printf("Created: %s\n", project.Created)
		}
		if project.Updated != "" {
			fmt.Printf("Updated: %s\n", project.Updated)
		}
		return nil
	},
}
