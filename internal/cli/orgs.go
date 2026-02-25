package cli

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(orgsCmd)
	orgsCmd.AddCommand(orgsListCmd)
	orgsCmd.AddCommand(orgsGetCmd)
}

var orgsCmd = &cobra.Command{
	Use:     "orgs",
	Aliases: []string{"org", "o"},
	Short:   "Manage organizations",
	Example: "  ancla orgs list\n  ancla orgs get my-org",
	GroupID: "resources",
}

var orgsListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List organizations",
	Example: "  ancla orgs list",
	RunE: func(cmd *cobra.Command, args []string) error {
		req, _ := http.NewRequest("GET", apiURL("/organizations/"), nil)
		body, err := doRequest(req)
		if err != nil {
			return err
		}

		var orgs []struct {
			ID           string `json:"id"`
			Name         string `json:"name"`
			Slug         string `json:"slug"`
			MemberCount  int    `json:"member_count"`
			ProjectCount int    `json:"project_count"`
		}
		if err := json.Unmarshal(body, &orgs); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}

		if isJSON() {
			return printJSON(orgs)
		}

		var rows [][]string
		for _, o := range orgs {
			rows = append(rows, []string{o.Slug, o.Name, fmt.Sprintf("%d", o.MemberCount), fmt.Sprintf("%d", o.ProjectCount)})
		}
		table([]string{"SLUG", "NAME", "MEMBERS", "PROJECTS"}, rows)
		return nil
	},
}

var orgsGetCmd = &cobra.Command{
	Use:     "get <slug>",
	Short:   "Get organization details",
	Example: "  ancla orgs get my-org",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		req, _ := http.NewRequest("GET", apiURL("/organizations/"+args[0]), nil)
		body, err := doRequest(req)
		if err != nil {
			return err
		}

		var org struct {
			Name             string `json:"name"`
			Slug             string `json:"slug"`
			ProjectCount     int    `json:"project_count"`
			ApplicationCount int    `json:"application_count"`
			Members          []struct {
				Username string `json:"username"`
				Email    string `json:"email"`
				Admin    bool   `json:"admin"`
			} `json:"members"`
		}
		if err := json.Unmarshal(body, &org); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}

		if isJSON() {
			return printJSON(org)
		}

		fmt.Printf("Organization: %s (%s)\n", org.Name, org.Slug)
		fmt.Printf("Projects: %d  Applications: %d\n\n", org.ProjectCount, org.ApplicationCount)
		fmt.Println("Members:")
		var rows [][]string
		for _, m := range org.Members {
			rows = append(rows, []string{m.Username, m.Email, fmt.Sprintf("%v", m.Admin)})
		}
		table([]string{"USERNAME", "EMAIL", "ADMIN"}, rows)
		return nil
	},
}
