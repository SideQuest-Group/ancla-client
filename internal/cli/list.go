package cli

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(listCmd)
}

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all your projects grouped by organization",
	Example: "  ancla list",
	GroupID: "resources",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Fetch all orgs
		orgReq, _ := http.NewRequest("GET", apiURL("/organizations/"), nil)
		orgBody, err := doRequest(orgReq)
		if err != nil {
			return err
		}

		var orgs []struct {
			Name string `json:"name"`
			Slug string `json:"slug"`
		}
		if err := json.Unmarshal(orgBody, &orgs); err != nil {
			return fmt.Errorf("parsing organizations: %w", err)
		}

		// Fetch all projects
		projReq, _ := http.NewRequest("GET", apiURL("/projects/"), nil)
		projBody, err := doRequest(projReq)
		if err != nil {
			return err
		}

		var projects []struct {
			Name             string `json:"name"`
			Slug             string `json:"slug"`
			OrganizationSlug string `json:"organization_slug"`
		}
		if err := json.Unmarshal(projBody, &projects); err != nil {
			return fmt.Errorf("parsing projects: %w", err)
		}

		if isJSON() {
			grouped := make(map[string][]string)
			for _, p := range projects {
				grouped[p.OrganizationSlug] = append(grouped[p.OrganizationSlug], p.Name)
			}
			return printJSON(grouped)
		}

		// Group projects by org slug
		byOrg := make(map[string][]string)
		for _, p := range projects {
			byOrg[p.OrganizationSlug] = append(byOrg[p.OrganizationSlug], p.Name)
		}

		for _, org := range orgs {
			bold.Println(org.Name)
			projs := byOrg[org.Slug]
			if len(projs) == 0 {
				fmt.Println(color.HiBlackString("  (no projects)"))
			}
			for _, name := range projs {
				fmt.Printf("  %s\n", name)
			}
			fmt.Println()
		}

		return nil
	},
}
