package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
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
	Example: "  ancla projects list\n  ancla projects get my-org/my-project",
	GroupID: "resources",
}

var projectsListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List projects",
	Example: "  ancla projects list",
	RunE: func(cmd *cobra.Command, args []string) error {
		req, _ := http.NewRequest("GET", apiURL("/projects/"), nil)
		body, err := doRequest(req)
		if err != nil {
			return err
		}

		var projects []struct {
			ID               string `json:"id"`
			Name             string `json:"name"`
			Slug             string `json:"slug"`
			OrganizationSlug string `json:"organization_slug"`
			ApplicationCount int    `json:"application_count"`
		}
		if err := json.Unmarshal(body, &projects); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}

		if isJSON() {
			return printJSON(projects)
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintf(w, "%s\t%s\t%s\n", colorHeader("ORG/PROJECT"), colorHeader("NAME"), colorHeader("APPS"))
		for _, p := range projects {
			fmt.Fprintf(w, "%s/%s\t%s\t%d\n", p.OrganizationSlug, p.Slug, p.Name, p.ApplicationCount)
		}
		return w.Flush()
	},
}

var projectsGetCmd = &cobra.Command{
	Use:     "get <org>/<project>",
	Short:   "Get project details",
	Example: "  ancla projects get my-org/my-project",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		req, _ := http.NewRequest("GET", apiURL("/projects/"+args[0]), nil)
		body, err := doRequest(req)
		if err != nil {
			return err
		}

		var project struct {
			Name             string `json:"name"`
			Slug             string `json:"slug"`
			OrganizationSlug string `json:"organization_slug"`
			OrganizationName string `json:"organization_name"`
			ApplicationCount int    `json:"application_count"`
			Created          string `json:"created"`
			Updated          string `json:"updated"`
		}
		if err := json.Unmarshal(body, &project); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}

		if isJSON() {
			return printJSON(project)
		}

		fmt.Printf("Project: %s (%s/%s)\n", project.Name, project.OrganizationSlug, project.Slug)
		fmt.Printf("Organization: %s\n", project.OrganizationName)
		fmt.Printf("Applications: %d\n", project.ApplicationCount)
		if project.Created != "" {
			fmt.Printf("Created: %s\n", project.Created)
		}
		return nil
	},
}
