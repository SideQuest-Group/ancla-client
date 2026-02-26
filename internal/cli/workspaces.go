package cli

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(workspacesCmd)
	workspacesCmd.AddCommand(workspacesListCmd)
	workspacesCmd.AddCommand(workspacesGetCmd)
}

var workspacesCmd = &cobra.Command{
	Use:     "workspaces",
	Aliases: []string{"ws", "w"},
	Short:   "Manage workspaces",
	Long: `Manage workspaces on the Ancla platform.

Workspaces are the top-level grouping for projects and team members.
Use sub-commands to list your workspaces or inspect a specific one,
including its members, projects, and service counts.`,
	Example: "  ancla workspaces list\n  ancla workspaces get my-workspace",
	GroupID: "resources",
}

var workspacesListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List workspaces",
	Example: "  ancla workspaces list",
	RunE: func(cmd *cobra.Command, args []string) error {
		req, _ := http.NewRequest("GET", apiURL("/workspaces/"), nil)
		body, err := doRequest(req)
		if err != nil {
			return err
		}

		var workspaces []struct {
			ID           string `json:"id"`
			Name         string `json:"name"`
			Slug         string `json:"slug"`
			MemberCount  int    `json:"member_count"`
			ProjectCount int    `json:"project_count"`
		}
		if err := json.Unmarshal(body, &workspaces); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}

		if isJSON() {
			return printJSON(workspaces)
		}

		var rows [][]string
		for _, w := range workspaces {
			rows = append(rows, []string{w.Slug, w.Name, fmt.Sprintf("%d", w.MemberCount), fmt.Sprintf("%d", w.ProjectCount)})
		}
		table([]string{"SLUG", "NAME", "MEMBERS", "PROJECTS"}, rows)
		return nil
	},
}

var workspacesGetCmd = &cobra.Command{
	Use:               "get <slug>",
	Short:             "Get workspace details",
	Example:           "  ancla workspaces get my-workspace",
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeWorkspaces,
	RunE: func(cmd *cobra.Command, args []string) error {
		req, _ := http.NewRequest("GET", apiURL("/workspaces/"+args[0]), nil)
		body, err := doRequest(req)
		if err != nil {
			return err
		}

		var ws struct {
			Name         string `json:"name"`
			Slug         string `json:"slug"`
			ProjectCount int    `json:"project_count"`
			ServiceCount int    `json:"service_count"`
			Members      []struct {
				Username     string `json:"username"`
				Email        string `json:"email"`
				Admin        bool   `json:"admin"`
				ServiceCount int    `json:"service_count"`
			} `json:"members"`
		}
		if err := json.Unmarshal(body, &ws); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}

		if isJSON() {
			return printJSON(ws)
		}

		fmt.Printf("Workspace: %s (%s)\n", ws.Name, ws.Slug)
		fmt.Printf("Projects: %d  Services: %d\n\n", ws.ProjectCount, ws.ServiceCount)
		fmt.Println("Members:")
		var rows [][]string
		for _, m := range ws.Members {
			rows = append(rows, []string{m.Username, m.Email, fmt.Sprintf("%v", m.Admin), fmt.Sprintf("%d", m.ServiceCount)})
		}
		table([]string{"USERNAME", "EMAIL", "ADMIN", "SERVICES"}, rows)
		return nil
	},
}
